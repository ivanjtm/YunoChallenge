package router

import (
	"fmt"
	"sort"
	"time"

	"github.com/velamarket/refund-router/internal/cost"
	"github.com/velamarket/refund-router/internal/model"
	"github.com/velamarket/refund-router/internal/rules"
)

// Router is the refund routing engine.
type Router struct {
	Processors []model.Processor
	RuleIndex  *rules.RuleIndex
}

// NewRouter creates a new Router from config.
func NewRouter(processors []model.Processor, compatRules []model.CompatibilityRule) *Router {
	return &Router{
		Processors: processors,
		RuleIndex:  rules.NewRuleIndex(compatRules),
	}
}

// SelectRoute finds the optimal refund route for a transaction.
func (r *Router) SelectRoute(tx model.Transaction, now time.Time) model.RefundRouteResult {
	// Step 1: Find eligible refund methods
	eligiblePaths := rules.FindEligiblePaths(tx, r.RuleIndex, now)

	// Step 2+3: For each eligible method, find processors and calculate costs
	var candidates []model.RefundCandidate

	for _, path := range eligiblePaths {
		// Account credit is special — no processor involved, zero cost
		if path.Method == model.RefundAccountCredit {
			candidates = append(candidates, model.RefundCandidate{
				ProcessorID:    "internal",
				ProcessorName:  "Account Credit",
				RefundMethod:   model.RefundAccountCredit,
				EstimatedCost:  0,
				ProcessingDays: 0,
				Reasoning:      path.Reason + "; funds credited to customer marketplace balance",
			})
			continue
		}

		// For each processor, check if it can handle this refund method
		for _, proc := range r.Processors {
			// Must support the transaction's country and currency
			if !cost.SupportsCountryAndCurrency(proc, tx.Country, tx.Currency) {
				continue
			}

			// Find matching fee entry
			fee := cost.FindMatchingFee(proc, path.Method, tx.PaymentMethod, tx.Currency)
			if fee == nil {
				continue
			}

			// Calculate cost
			refundCost := cost.Calculate(tx.Amount, *fee)

			// Get processing time
			days := 0
			if d, ok := proc.ProcessingDays[path.Method]; ok {
				days = d
			}

			// Build reasoning string
			reasoning := buildReasoning(tx, proc, path, *fee, refundCost, days)

			candidates = append(candidates, model.RefundCandidate{
				ProcessorID:    proc.ID,
				ProcessorName:  proc.Name,
				RefundMethod:   path.Method,
				EstimatedCost:  refundCost,
				ProcessingDays: days,
				Reasoning:      reasoning,
			})
		}
	}

	// Step 5: Rank candidates
	sort.Slice(candidates, func(i, j int) bool {
		// Primary: cost ascending
		if candidates[i].EstimatedCost != candidates[j].EstimatedCost {
			return candidates[i].EstimatedCost < candidates[j].EstimatedCost
		}
		// Secondary: processing time ascending
		if candidates[i].ProcessingDays != candidates[j].ProcessingDays {
			return candidates[i].ProcessingDays < candidates[j].ProcessingDays
		}
		// Tertiary: prefer original processor for simpler reconciliation
		if candidates[i].ProcessorID == tx.ProcessorID && candidates[j].ProcessorID != tx.ProcessorID {
			return true
		}
		return false
	})

	// Handle no candidates (shouldn't happen due to account credit fallback)
	if len(candidates) == 0 {
		candidates = []model.RefundCandidate{{
			ProcessorID:    "internal",
			ProcessorName:  "Account Credit",
			RefundMethod:   model.RefundAccountCredit,
			EstimatedCost:  0,
			ProcessingDays: 0,
			Reasoning:      "No eligible refund methods found; defaulting to account credit",
		}}
	}

	// Step 6: Calculate naive cost
	naiveCost := cost.CalculateNaive(tx, r.Processors)

	// Step 7: Assemble result
	selected := candidates[0]
	var alternatives []model.RefundCandidate
	if len(candidates) > 1 {
		alternatives = candidates[1:]
	}

	return model.RefundRouteResult{
		TransactionID: tx.ID,
		Selected:      selected,
		Alternatives:  alternatives,
		NaiveCost:     naiveCost,
		Savings:       naiveCost - selected.EstimatedCost,
	}
}

// buildReasoning creates a human-readable explanation for a routing choice.
func buildReasoning(tx model.Transaction, proc model.Processor, path rules.EligiblePath, fee model.RefundMethodFee, refundCost float64, days int) string {
	methodDesc := string(path.Method)
	switch path.Method {
	case model.RefundReversal:
		return fmt.Sprintf("Free reversal via %s; %s", proc.Name, path.Reason)
	case model.RefundSameMethod:
		methodDesc = fmt.Sprintf("%s-to-%s", tx.PaymentMethod, tx.PaymentMethod)
	case model.RefundBankTransfer:
		methodDesc = "bank transfer"
	}

	costDesc := ""
	if fee.BaseFee > 0 && fee.PercentFee > 0 {
		costDesc = fmt.Sprintf("%.2f base + %.1f%% = %.2f %s", fee.BaseFee, fee.PercentFee*100, refundCost, tx.Currency)
	} else if fee.PercentFee > 0 {
		costDesc = fmt.Sprintf("%.1f%% = %.2f %s", fee.PercentFee*100, refundCost, tx.Currency)
	} else {
		costDesc = fmt.Sprintf("%.2f %s", refundCost, tx.Currency)
	}

	timeDesc := ""
	if days == 0 {
		timeDesc = "instant"
	} else if days == 1 {
		timeDesc = "1 day"
	} else {
		timeDesc = fmt.Sprintf("%d days", days)
	}

	return fmt.Sprintf("%s via %s: %s, %s processing; %s", methodDesc, proc.Name, costDesc, timeDesc, path.Reason)
}
