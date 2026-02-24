package cost

import (
	"math"

	"github.com/ivanjtm/YunoChallenge/internal/model"
)

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
