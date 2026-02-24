package historical

import (
	"math"
	"sort"
	"time"

	"github.com/velamarket/refund-router/internal/model"
	"github.com/velamarket/refund-router/internal/router"
)

// Analyze performs historical cost analysis over a set of transactions.
// It computes what each refund would have cost naively vs. with smart routing.
func Analyze(txns []model.Transaction, r *router.Router, now time.Time) model.HistoricalAnalysis {
	result := model.HistoricalAnalysis{
		TotalTransactions: len(txns),
		MonthlySavings:    make(map[string]float64),
	}

	// Track per-corridor and per-processor costs
	type corridorKey struct {
		Country       model.Country
		PaymentMethod model.PaymentMethod
	}
	corridorCosts := make(map[corridorKey]struct {
		totalNaive float64
		totalSmart float64
		count      int
	})
	processorCosts := make(map[string]struct {
		totalNaive float64
		totalSmart float64
		count      int
	})

	for _, tx := range txns {
		route := r.SelectRoute(tx, now)

		naiveCost := route.NaiveCost
		smartCost := route.Selected.EstimatedCost
		savings := naiveCost - smartCost

		result.TotalActualCost += naiveCost
		result.TotalSmartCost += smartCost
		result.TotalSavings += savings

		// Monthly savings
		monthKey := tx.Timestamp.Format("2006-01")
		result.MonthlySavings[monthKey] += savings

		// Corridor tracking (country + payment method)
		ck := corridorKey{tx.Country, tx.PaymentMethod}
		entry := corridorCosts[ck]
		entry.totalNaive += naiveCost
		entry.totalSmart += smartCost
		entry.count++
		corridorCosts[ck] = entry

		// Processor tracking (original processor)
		pe := processorCosts[tx.ProcessorID]
		pe.totalNaive += naiveCost
		pe.count++
		processorCosts[tx.ProcessorID] = pe
	}

	// Round totals
	result.TotalActualCost = math.Round(result.TotalActualCost*100) / 100
	result.TotalSmartCost = math.Round(result.TotalSmartCost*100) / 100
	result.TotalSavings = math.Round(result.TotalSavings*100) / 100

	// Annual projection (extrapolate from data span)
	if len(txns) > 0 {
		// Find date range of transactions
		minTime := txns[0].Timestamp
		maxTime := txns[0].Timestamp
		for _, tx := range txns[1:] {
			if tx.Timestamp.Before(minTime) {
				minTime = tx.Timestamp
			}
			if tx.Timestamp.After(maxTime) {
				maxTime = tx.Timestamp
			}
		}
		spanDays := maxTime.Sub(minTime).Hours() / 24
		if spanDays > 0 {
			result.AnnualProjection = math.Round(result.TotalSavings / spanDays * 365 * 100) / 100
		}
	}

	// Build most expensive corridors (sorted by total naive cost desc)
	for ck, data := range corridorCosts {
		result.MostExpensiveCorridors = append(result.MostExpensiveCorridors, model.CostCorridor{
			Country:       ck.Country,
			PaymentMethod: ck.PaymentMethod,
			AvgCost:       math.Round(data.totalNaive/float64(data.count)*100) / 100,
			TotalCost:     math.Round(data.totalNaive*100) / 100,
			Count:         data.count,
		})
	}
	sort.Slice(result.MostExpensiveCorridors, func(i, j int) bool {
		return result.MostExpensiveCorridors[i].TotalCost > result.MostExpensiveCorridors[j].TotalCost
	})
	// Keep top 5
	if len(result.MostExpensiveCorridors) > 5 {
		result.MostExpensiveCorridors = result.MostExpensiveCorridors[:5]
	}

	// Build highest cost processors (sorted by total cost desc)
	for procID, data := range processorCosts {
		result.HighestCostProcessors = append(result.HighestCostProcessors, model.ProcessorCostRank{
			ProcessorID: procID,
			TotalCost:   math.Round(data.totalNaive*100) / 100,
			AvgCost:     math.Round(data.totalNaive/float64(data.count)*100) / 100,
			Count:       data.count,
		})
	}
	sort.Slice(result.HighestCostProcessors, func(i, j int) bool {
		return result.HighestCostProcessors[i].TotalCost > result.HighestCostProcessors[j].TotalCost
	})

	// Add complex refund rule notes
	result.ComplexRefundRules = []model.ComplexRuleNote{
		{
			Rule:        "OXXO_NO_SELF_REFUND",
			Description: "OXXO cash payments cannot be refunded as OXXO",
			Impact:      "Forces SPEI bank transfer, typically higher cost than same-method refunds",
		},
		{
			Rule:        "BOLETO_NO_SELF_REFUND",
			Description: "Boleto voucher payments cannot be refunded as Boleto",
			Impact:      "Requires PIX or bank transfer; PIX is much cheaper when within 90-day window",
		},
		{
			Rule:        "EFECTY_NO_SELF_REFUND",
			Description: "Efecty cash payments cannot be refunded as Efecty",
			Impact:      "Requires PSE or bank transfer; PSE is cheaper when within 60-day window",
		},
		{
			Rule:        "PIX_90_DAY_WINDOW",
			Description: "PIX-to-PIX refunds only available within 90 days of original transaction",
			Impact:      "After 90 days, must use bank transfer at ~3x the cost of PIX refund",
		},
		{
			Rule:        "PSE_60_DAY_WINDOW",
			Description: "PSE-to-PSE refunds only available within 60 days of original transaction",
			Impact:      "After 60 days, must use bank transfer at ~2x the cost of PSE refund",
		},
		{
			Rule:        "REVERSAL_24H_WINDOW",
			Description: "Free reversals (voids) only available for unsettled transactions within 24 hours",
			Impact:      "Catching transactions within this window saves 100% of refund fees",
		},
	}

	// Round monthly savings
	for k, v := range result.MonthlySavings {
		result.MonthlySavings[k] = math.Round(v*100) / 100
	}

	return result
}
