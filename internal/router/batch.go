package router

import (
	"fmt"
	"time"

	"github.com/velamarket/refund-router/internal/model"
	"github.com/velamarket/refund-router/internal/rules"
)

// AnalyzeBatch processes multiple refund requests and produces an optimization report.
func (r *Router) AnalyzeBatch(txns []model.Transaction, now time.Time) model.BatchRefundResult {
	result := model.BatchRefundResult{
		TotalTransactions: len(txns),
		ByProcessor:       make(map[string]model.ProcessorSummary),
		ByPaymentMethod:   make(map[string]model.MethodSummary),
	}

	for _, tx := range txns {
		// Route each transaction
		route := r.SelectRoute(tx, now)
		result.Results = append(result.Results, route)

		// Accumulate totals
		result.TotalNaiveCost += route.NaiveCost
		result.TotalSmartCost += route.Selected.EstimatedCost
		result.TotalSavings += route.Savings

		// Group by original processor
		ps := result.ByProcessor[tx.ProcessorID]
		ps.ProcessorID = tx.ProcessorID
		ps.NaiveCost += route.NaiveCost
		ps.SmartCost += route.Selected.EstimatedCost
		ps.Savings += route.Savings
		ps.TransactionCount++
		result.ByProcessor[tx.ProcessorID] = ps

		// Group by original payment method
		methodKey := string(tx.PaymentMethod)
		ms := result.ByPaymentMethod[methodKey]
		ms.Method = methodKey
		ms.NaiveCost += route.NaiveCost
		ms.SmartCost += route.Selected.EstimatedCost
		ms.Savings += route.Savings
		ms.TransactionCount++
		result.ByPaymentMethod[methodKey] = ms

		// Check for time-sensitive windows (threshold: 15 days)
		tsFlags := rules.TimeSensitiveWindows(tx, r.RuleIndex, now, 15)
		result.TimeSensitive = append(result.TimeSensitive, tsFlags...)

		// Check for limited options (fewer than 3 candidates total)
		if len(route.Alternatives) < 2 { // selected + fewer than 2 alternatives = limited
			totalOptions := 1 + len(route.Alternatives)
			// Only flag methods that inherently have limited options
			switch tx.PaymentMethod {
			case model.MethodOXXO, model.MethodBoleto, model.MethodEfecty:
				result.LimitedOptions = append(result.LimitedOptions, model.LimitedOptionFlag{
					TransactionID:    tx.ID,
					OriginalMethod:   string(tx.PaymentMethod),
					AvailableOptions: totalOptions,
					Message:          fmt.Sprintf("%s cannot be refunded via %s. Only %d routing option(s) available.", tx.PaymentMethod, tx.PaymentMethod, totalOptions),
				})
			}
		}
	}

	// Calculate savings percentage
	if result.TotalNaiveCost > 0 {
		result.SavingsPercent = (result.TotalSavings / result.TotalNaiveCost) * 100
	}

	// Round totals
	result.TotalNaiveCost = roundTo2(result.TotalNaiveCost)
	result.TotalSmartCost = roundTo2(result.TotalSmartCost)
	result.TotalSavings = roundTo2(result.TotalSavings)
	result.SavingsPercent = roundTo2(result.SavingsPercent)

	return result
}

func roundTo2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
