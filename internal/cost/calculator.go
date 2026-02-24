package cost

import (
	"math"

	"github.com/velamarket/refund-router/internal/model"
)

// Calculate computes the refund fee for a given amount and fee structure.
// Formula: cost = base_fee + (amount * percent_fee)
// Then clamp: max(min_fee, min(max_fee, cost))
// If max_fee is 0, there's no cap.
func Calculate(amount float64, fee model.RefundMethodFee) float64 {
	if fee.Method == model.RefundReversal {
		return 0 // reversals are always free
	}
	if fee.Method == model.RefundAccountCredit {
		return 0 // account credits have no processor fee
	}

	cost := fee.BaseFee + (amount * fee.PercentFee)

	if cost < fee.MinFee {
		cost = fee.MinFee
	}
	if fee.MaxFee > 0 && cost > fee.MaxFee {
		cost = fee.MaxFee
	}

	// Round to 2 decimal places (or 0 for COP)
	cost = math.Round(cost*100) / 100

	return cost
}

// FindMatchingFee finds the fee entry in a processor that matches the given
// refund method, original payment method, and currency.
// Returns nil if no matching fee is found.
func FindMatchingFee(proc model.Processor, refundMethod model.RefundMethod, originalMethod model.PaymentMethod, currency model.Currency) *model.RefundMethodFee {
	for i, fee := range proc.RefundFees {
		if fee.Method != refundMethod {
			continue
		}
		if fee.Currency != currency && fee.Currency != "" {
			continue
		}
		// Check if this fee covers the original payment method
		for _, pm := range fee.PaymentMethods {
			if pm == originalMethod {
				return &proc.RefundFees[i]
			}
		}
	}
	return nil
}

// CalculateNaive computes the "naive" refund cost -- what it would cost to
// refund through the original processor using the default method.
// Strategy: try SAME_METHOD first, then BANK_TRANSFER, then return a high estimate.
func CalculateNaive(tx model.Transaction, processors []model.Processor) float64 {
	// Find the original processor
	var origProc *model.Processor
	for i, p := range processors {
		if p.ID == tx.ProcessorID {
			origProc = &processors[i]
			break
		}
	}
	if origProc == nil {
		// Unknown processor -- return a high default cost (3.5% as worst case)
		return math.Round(tx.Amount*0.035*100) / 100
	}

	// Try SAME_METHOD through original processor
	if fee := FindMatchingFee(*origProc, model.RefundSameMethod, tx.PaymentMethod, tx.Currency); fee != nil {
		return Calculate(tx.Amount, *fee)
	}

	// Try BANK_TRANSFER through original processor
	if fee := FindMatchingFee(*origProc, model.RefundBankTransfer, tx.PaymentMethod, tx.Currency); fee != nil {
		return Calculate(tx.Amount, *fee)
	}

	// Original processor can't handle this refund at all -- use worst case
	return math.Round(tx.Amount*0.035*100) / 100
}

// SupportsCountryAndCurrency checks if a processor supports the given country and currency.
func SupportsCountryAndCurrency(proc model.Processor, country model.Country, currency model.Currency) bool {
	countryOK := false
	for _, c := range proc.SupportedCountries {
		if c == country {
			countryOK = true
			break
		}
	}
	if !countryOK {
		return false
	}
	for _, c := range proc.SupportedCurrencies {
		if c == currency {
			return true
		}
	}
	return false
}
