package router

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/ivanjtm/YunoChallenge/internal/model"
	"github.com/ivanjtm/YunoChallenge/internal/rules"
)

// indexedRoute pairs a routing result with its original index to preserve order
// after parallel fan-out.
type indexedRoute struct {
	index int
	tx    model.Transaction
	route model.RefundRouteResult
}

// AnalyzeBatch processes multiple refund requests concurrently and produces an
// optimization report. SelectRoute calls are fanned out across NumCPU workers;
// accumulation happens single-threaded to avoid lock contention on maps.
func (r *Router) AnalyzeBatch(txns []model.Transaction, now time.Time) model.BatchRefundResult {
	n := len(txns)
	result := model.BatchRefundResult{
		TotalTransactions: n,
		Results:           make([]model.RefundRouteResult, n),
		ByProcessor:       make(map[string]model.ProcessorSummary),
		ByPaymentMethod:   make(map[string]model.MethodSummary),
		TimeSensitive:     make([]model.TimeSensitiveFlag, 0),
		LimitedOptions:    make([]model.LimitedOptionFlag, 0),
	}

	// Fan out routing across workers
	workers := runtime.NumCPU()
	if workers > n {
		workers = n
	}
	if workers < 1 {
		workers = 1
	}

	jobs := make(chan indexedRoute, n)
	results := make(chan indexedRoute, n)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				j.route = r.SelectRoute(j.tx, now)
				results <- j
			}
		}()
	}

	for i, tx := range txns {
		jobs <- indexedRoute{index: i, tx: tx}
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	// Accumulate single-threaded — no locks needed on maps
	for ir := range results {
		result.Results[ir.index] = ir.route
		route := ir.route
		tx := ir.tx

		result.TotalNaiveCost += route.NaiveCost
		result.TotalSmartCost += route.Selected.EstimatedCost
		result.TotalSavings += route.Savings

		ps := result.ByProcessor[tx.ProcessorID]
		ps.ProcessorID = tx.ProcessorID
		ps.NaiveCost += route.NaiveCost
		ps.SmartCost += route.Selected.EstimatedCost
		ps.Savings += route.Savings
		ps.TransactionCount++
		result.ByProcessor[tx.ProcessorID] = ps

		methodKey := string(tx.PaymentMethod)
		ms := result.ByPaymentMethod[methodKey]
		ms.Method = methodKey
		ms.NaiveCost += route.NaiveCost
		ms.SmartCost += route.Selected.EstimatedCost
		ms.Savings += route.Savings
		ms.TransactionCount++
		result.ByPaymentMethod[methodKey] = ms

		tsFlags := rules.TimeSensitiveWindows(tx, r.RuleIndex, now, 15)
		result.TimeSensitive = append(result.TimeSensitive, tsFlags...)

		switch tx.PaymentMethod {
		case model.MethodOXXO, model.MethodBoleto, model.MethodEfecty:
			totalOptions := 1 + len(route.Alternatives)
			result.LimitedOptions = append(result.LimitedOptions, model.LimitedOptionFlag{
				TransactionID:    tx.ID,
				OriginalMethod:   string(tx.PaymentMethod),
				AvailableOptions: totalOptions,
				Message:          fmt.Sprintf("%s cannot be refunded via %s; requires alternative method. %d routing option(s) available.", tx.PaymentMethod, tx.PaymentMethod, totalOptions),
			})
		}
	}

	if result.TotalNaiveCost > 0 {
		result.SavingsPercent = (result.TotalSavings / result.TotalNaiveCost) * 100
	}

	result.TotalNaiveCost = roundTo2(result.TotalNaiveCost)
	result.TotalSmartCost = roundTo2(result.TotalSmartCost)
	result.TotalSavings = roundTo2(result.TotalSavings)
	result.SavingsPercent = roundTo2(result.SavingsPercent)

	return result
}

func roundTo2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
