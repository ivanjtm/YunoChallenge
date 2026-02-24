package rules

import (
	"fmt"
	"math"
	"time"

	"github.com/velamarket/refund-router/internal/model"
)

// IsReversalEligible checks if a transaction can be reversed (voided).
// Reversals require: transaction < 24 hours old AND not yet settled.
func IsReversalEligible(tx model.Transaction, now time.Time) (eligible bool, reason string) {
	hoursSince := now.Sub(tx.Timestamp).Hours()
	if tx.Settled {
		return false, "Transaction already settled; reversal not available"
	}
	if hoursSince >= 24 {
		return false, fmt.Sprintf("Transaction is %.0f hours old; reversal requires < 24 hours", hoursSince)
	}
	return true, fmt.Sprintf("Transaction is %.1f hours old and unsettled; free reversal available", hoursSince)
}

// IsWithinTimeWindow checks if a refund method's time window is still open.
// max_age_days=0 means no limit.
// For REVERSAL, use IsReversalEligible instead.
func IsWithinTimeWindow(tx model.Transaction, allowed model.AllowedRefund, now time.Time) (eligible bool, reason string) {
	if allowed.MaxAgeDays == 0 {
		return true, "No time limit for this refund method"
	}
	daysSince := int(math.Floor(now.Sub(tx.Timestamp).Hours() / 24))
	if daysSince > allowed.MaxAgeDays {
		return false, fmt.Sprintf("Transaction is %d days old; %s window is %d days", daysSince, allowed.Method, allowed.MaxAgeDays)
	}
	remaining := allowed.MaxAgeDays - daysSince
	return true, fmt.Sprintf("Within %s window (%d of %d days used, %d remaining)", allowed.Method, daysSince, allowed.MaxAgeDays, remaining)
}

// DaysUntilExpiry returns how many days remain before a time window closes.
// Returns -1 if the window has no limit, or if already expired returns negative days.
func DaysUntilExpiry(tx model.Transaction, allowed model.AllowedRefund, now time.Time) int {
	if allowed.MaxAgeDays == 0 {
		return -1
	}
	daysSince := int(math.Floor(now.Sub(tx.Timestamp).Hours() / 24))
	return allowed.MaxAgeDays - daysSince
}
