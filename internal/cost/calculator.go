package cost

import (
	"math"

	"github.com/velamarket/refund-router/internal/model"
)

// Calculate computes the refund fee for a given amount and fee structure.
// Formula: cost = base_fee + (amount * percent_fee), clamped to [min_fee, max_fee].
// MaxFee of 0 means no cap. Reversals and account credits are always free.
func Calculate(amount float64, fee model.RefundMethodFee) float64 {
	if fee.Method == model.RefundReversal {
		return 0
	}
	if fee.Method == model.RefundAccountCredit {
		return 0
	}

	cost := fee.BaseFee + (amount * fee.PercentFee)

	if cost < fee.MinFee {
		cost = fee.MinFee
	}
	if fee.MaxFee > 0 && cost > fee.MaxFee {
		cost = fee.MaxFee
	}

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
		for _, pm := range fee.PaymentMethods {
			if pm == originalMethod {
				return &proc.RefundFees[i]
			}
		}
	}
	return nil
}

// CalculateNaive computes the "naive" refund cost: what it would cost to refund
// through the original processor using the default method (SAME_METHOD, then
// BANK_TRANSFER, then a 3.5% worst-case estimate).
func CalculateNaive(tx model.Transaction, processors []model.Processor) float64 {
	var origProc *model.Processor
	for i, p := range processors {
		if p.ID == tx.ProcessorID {
			origProc = &processors[i]
			break
		}
	}
	if origProc == nil {
		return math.Round(tx.Amount*0.035*100) / 100
	}

	if fee := FindMatchingFee(*origProc, model.RefundSameMethod, tx.PaymentMethod, tx.Currency); fee != nil {
		return Calculate(tx.Amount, *fee)
	}

	if fee := FindMatchingFee(*origProc, model.RefundBankTransfer, tx.PaymentMethod, tx.Currency); fee != nil {
		return Calculate(tx.Amount, *fee)
	}

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
