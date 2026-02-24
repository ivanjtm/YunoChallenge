package rules

import (
	"fmt"
	"time"

	"github.com/velamarket/refund-router/internal/model"
)

// EligiblePath represents a refund method that passed all eligibility checks.
type EligiblePath struct {
	Method model.RefundMethod
	Reason string
}

// FindEligiblePaths returns all refund methods that are available for this transaction right now.
func FindEligiblePaths(tx model.Transaction, ruleIndex *RuleIndex, now time.Time) []EligiblePath {
	allowed := ruleIndex.AllowedRefundMethods(tx.PaymentMethod, tx.Country)
	if len(allowed) == 0 {
		// No rules found -- only account credit is available
		return []EligiblePath{
			{Method: model.RefundAccountCredit, Reason: "No compatibility rules found; only account credit available"},
		}
	}

	var paths []EligiblePath
	for _, ar := range allowed {
		switch ar.Method {
		case model.RefundReversal:
			if ok, reason := IsReversalEligible(tx, now); ok {
				paths = append(paths, EligiblePath{Method: ar.Method, Reason: reason})
			}
		default:
			// Check settle requirement if specified
			if ar.RequireSettled != nil {
				if *ar.RequireSettled && !tx.Settled {
					continue // requires settled but isn't
				}
				if !*ar.RequireSettled && tx.Settled {
					continue // requires unsettled but is settled
				}
			}
			// Check time window
			if ok, reason := IsWithinTimeWindow(tx, ar, now); ok {
				paths = append(paths, EligiblePath{Method: ar.Method, Reason: reason})
			}
		}
	}

	// If no paths found at all, account credit is always available as last resort
	if len(paths) == 0 {
		paths = append(paths, EligiblePath{
			Method: model.RefundAccountCredit,
			Reason: "No eligible refund methods; falling back to account credit",
		})
	}

	return paths
}

// TimeSensitiveWindows returns windows that are approaching expiry (within threshold days).
// Used by batch analysis to flag urgent refunds.
func TimeSensitiveWindows(tx model.Transaction, ruleIndex *RuleIndex, now time.Time, thresholdDays int) []model.TimeSensitiveFlag {
	allowed := ruleIndex.AllowedRefundMethods(tx.PaymentMethod, tx.Country)
	var flags []model.TimeSensitiveFlag

	for _, ar := range allowed {
		if ar.Method == model.RefundReversal {
			// Check if reversal window is closing (within last 6 hours of 24h)
			hoursSince := now.Sub(tx.Timestamp).Hours()
			if !tx.Settled && hoursSince >= 18 && hoursSince < 24 {
				hoursLeft := 24 - hoursSince
				flags = append(flags, model.TimeSensitiveFlag{
					TransactionID: tx.ID,
					WindowType:    "REVERSAL_24H",
					ExpiresAt:     tx.Timestamp.Add(24 * time.Hour),
					DaysRemaining: 0,
					Message:       fmt.Sprintf("Free reversal window closes in %.1f hours", hoursLeft),
				})
			}
			continue
		}
		if ar.MaxAgeDays == 0 {
			continue // no time limit
		}
		remaining := DaysUntilExpiry(tx, ar, now)
		if remaining >= 0 && remaining <= thresholdDays {
			windowName := windowTypeName(tx.PaymentMethod, ar)
			flags = append(flags, model.TimeSensitiveFlag{
				TransactionID: tx.ID,
				WindowType:    windowName,
				ExpiresAt:     tx.Timestamp.AddDate(0, 0, ar.MaxAgeDays),
				DaysRemaining: remaining,
				Message:       fmt.Sprintf("%s refund window expires in %d days. After expiry, more expensive alternatives required.", windowName, remaining),
			})
		}
	}
	return flags
}

// windowTypeName generates a human-readable window type name.
func windowTypeName(method model.PaymentMethod, ar model.AllowedRefund) string {
	return fmt.Sprintf("%s_%s_%dD", method, ar.Method, ar.MaxAgeDays)
}
