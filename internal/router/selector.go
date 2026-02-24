package router

import (
	"fmt"
	"sort"
	"time"

	"github.com/ivanjtm/YunoChallenge/internal/cost"
	"github.com/ivanjtm/YunoChallenge/internal/model"
	"github.com/ivanjtm/YunoChallenge/internal/rules"
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
	eligiblePaths := rules.FindEligiblePaths(tx, r.RuleIndex, now)

	var candidates []model.RefundCandidate

	for _, path := range eligiblePaths {
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

		for _, proc := range r.Processors {
			if !cost.SupportsCountryAndCurrency(proc, tx.Country, tx.Currency) {
				continue
			}

			fee := cost.FindMatchingFee(proc, path.Method, tx.PaymentMethod, tx.Currency)
			if fee == nil {
				continue
			}

			refundCost := cost.Calculate(tx.Amount, *fee)

			days := 0
			if d, ok := proc.ProcessingDays[path.Method]; ok {
				days = d
			}

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

	sort.Slice(candidates, func(i, j int) bool {
		// Account credit is always ranked last: it's free but worse for customer
		// experience since money stays locked in the marketplace balance.
		iIsCredit := candidates[i].RefundMethod == model.RefundAccountCredit
		jIsCredit := candidates[j].RefundMethod == model.RefundAccountCredit
		if iIsCredit != jIsCredit {
			return !iIsCredit
		}
		if candidates[i].EstimatedCost != candidates[j].EstimatedCost {
			return candidates[i].EstimatedCost < candidates[j].EstimatedCost
		}
		if candidates[i].ProcessingDays != candidates[j].ProcessingDays {
			return candidates[i].ProcessingDays < candidates[j].ProcessingDays
		}
		if candidates[i].ProcessorID == tx.ProcessorID && candidates[j].ProcessorID != tx.ProcessorID {
			return true
		}
		return false
	})

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

	naiveCost := cost.CalculateNaive(tx, r.Processors)

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
