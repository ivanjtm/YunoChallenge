package router

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/ivanjtm/YunoChallenge/internal/model"
	"github.com/ivanjtm/YunoChallenge/internal/rules"
)

func boolPtr(b bool) *bool { return &b }

func testProcessorPayBR() model.Processor {
	return model.Processor{
		ID:                  "paybr",
		Name:                "PayBR",
		SupportedCountries:  []model.Country{model.CountryBR},
		SupportedCurrencies: []model.Currency{model.CurrencyBRL},
		RefundFees: []model.RefundMethodFee{
			{
				Method:         model.RefundReversal,
				PaymentMethods: []model.PaymentMethod{model.MethodCreditCard, model.MethodPIX},
				Currency:       model.CurrencyBRL,
				BaseFee:        0.0,
				PercentFee:     0.0,
				MinFee:         0.0,
				MaxFee:         0,
			},
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodPIX},
				Currency:       model.CurrencyBRL,
				BaseFee:        0.5,
				PercentFee:     0.005,
				MinFee:         0.75,
				MaxFee:         0,
			},
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodCreditCard},
				Currency:       model.CurrencyBRL,
				BaseFee:        1.5,
				PercentFee:     0.025,
				MinFee:         2.0,
				MaxFee:         150.0,
			},
			{
				Method:         model.RefundBankTransfer,
				PaymentMethods: []model.PaymentMethod{model.MethodPIX, model.MethodBoleto, model.MethodCreditCard},
				Currency:       model.CurrencyBRL,
				BaseFee:        1.0,
				PercentFee:     0.015,
				MinFee:         1.5,
				MaxFee:         100.0,
			},
		},
		DailyQuota: 1000,
		ProcessingDays: map[model.RefundMethod]int{
			model.RefundReversal:     0,
			model.RefundSameMethod:   1,
			model.RefundBankTransfer: 2,
		},
	}
}

func testProcessorMexPay() model.Processor {
	return model.Processor{
		ID:                  "mexpay",
		Name:                "MexPay",
		SupportedCountries:  []model.Country{model.CountryMX},
		SupportedCurrencies: []model.Currency{model.CurrencyMXN},
		RefundFees: []model.RefundMethodFee{
			{
				Method:         model.RefundReversal,
				PaymentMethods: []model.PaymentMethod{model.MethodCreditCard, model.MethodSPEI},
				Currency:       model.CurrencyMXN,
				BaseFee:        0.0,
				PercentFee:     0.0,
				MinFee:         0.0,
				MaxFee:         0,
			},
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodSPEI},
				Currency:       model.CurrencyMXN,
				BaseFee:        5.0,
				PercentFee:     0.008,
				MinFee:         8.0,
				MaxFee:         0,
			},
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodCreditCard},
				Currency:       model.CurrencyMXN,
				BaseFee:        15.0,
				PercentFee:     0.02,
				MinFee:         20.0,
				MaxFee:         2500.0,
			},
			{
				Method:         model.RefundBankTransfer,
				PaymentMethods: []model.PaymentMethod{model.MethodSPEI, model.MethodOXXO, model.MethodCreditCard},
				Currency:       model.CurrencyMXN,
				BaseFee:        10.0,
				PercentFee:     0.012,
				MinFee:         15.0,
				MaxFee:         1800.0,
			},
		},
		DailyQuota: 800,
		ProcessingDays: map[model.RefundMethod]int{
			model.RefundReversal:     0,
			model.RefundSameMethod:   1,
			model.RefundBankTransfer: 2,
		},
	}
}

func testProcessorColPay() model.Processor {
	return model.Processor{
		ID:                  "colpay",
		Name:                "ColPay",
		SupportedCountries:  []model.Country{model.CountryCO},
		SupportedCurrencies: []model.Currency{model.CurrencyCOP},
		RefundFees: []model.RefundMethodFee{
			{
				Method:         model.RefundReversal,
				PaymentMethods: []model.PaymentMethod{model.MethodCreditCard, model.MethodPSE},
				Currency:       model.CurrencyCOP,
				BaseFee:        0.0,
				PercentFee:     0.0,
				MinFee:         0.0,
				MaxFee:         0,
			},
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodPSE},
				Currency:       model.CurrencyCOP,
				BaseFee:        1500.0,
				PercentFee:     0.006,
				MinFee:         2000.0,
				MaxFee:         0,
			},
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodCreditCard},
				Currency:       model.CurrencyCOP,
				BaseFee:        3500.0,
				PercentFee:     0.022,
				MinFee:         5000.0,
				MaxFee:         350000.0,
			},
			{
				Method:         model.RefundBankTransfer,
				PaymentMethods: []model.PaymentMethod{model.MethodPSE, model.MethodEfecty, model.MethodCreditCard},
				Currency:       model.CurrencyCOP,
				BaseFee:        2500.0,
				PercentFee:     0.018,
				MinFee:         4000.0,
				MaxFee:         280000.0,
			},
		},
		DailyQuota: 600,
		ProcessingDays: map[model.RefundMethod]int{
			model.RefundReversal:     0,
			model.RefundSameMethod:   1,
			model.RefundBankTransfer: 3,
		},
	}
}

func testProcessorGlobalPay() model.Processor {
	return model.Processor{
		ID:                  "globalpay",
		Name:                "GlobalPay",
		SupportedCountries:  []model.Country{model.CountryBR, model.CountryMX, model.CountryCO},
		SupportedCurrencies: []model.Currency{model.CurrencyBRL, model.CurrencyMXN, model.CurrencyCOP},
		RefundFees: []model.RefundMethodFee{
			{
				Method:         model.RefundReversal,
				PaymentMethods: []model.PaymentMethod{model.MethodCreditCard},
				Currency:       model.CurrencyBRL,
				BaseFee:        0.0,
				PercentFee:     0.0,
				MinFee:         0.0,
				MaxFee:         0,
			},
			{
				Method:         model.RefundReversal,
				PaymentMethods: []model.PaymentMethod{model.MethodCreditCard},
				Currency:       model.CurrencyMXN,
				BaseFee:        0.0,
				PercentFee:     0.0,
				MinFee:         0.0,
				MaxFee:         0,
			},
			{
				Method:         model.RefundReversal,
				PaymentMethods: []model.PaymentMethod{model.MethodCreditCard},
				Currency:       model.CurrencyCOP,
				BaseFee:        0.0,
				PercentFee:     0.0,
				MinFee:         0.0,
				MaxFee:         0,
			},
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodCreditCard},
				Currency:       model.CurrencyBRL,
				BaseFee:        2.0,
				PercentFee:     0.02,
				MinFee:         3.0,
				MaxFee:         200.0,
			},
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodCreditCard},
				Currency:       model.CurrencyMXN,
				BaseFee:        20.0,
				PercentFee:     0.02,
				MinFee:         30.0,
				MaxFee:         3500.0,
			},
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodCreditCard},
				Currency:       model.CurrencyCOP,
				BaseFee:        5000.0,
				PercentFee:     0.02,
				MinFee:         7500.0,
				MaxFee:         500000.0,
			},
			{
				Method:         model.RefundBankTransfer,
				PaymentMethods: []model.PaymentMethod{model.MethodPIX, model.MethodBoleto, model.MethodCreditCard},
				Currency:       model.CurrencyBRL,
				BaseFee:        2.0,
				PercentFee:     0.02,
				MinFee:         3.0,
				MaxFee:         200.0,
			},
			{
				Method:         model.RefundBankTransfer,
				PaymentMethods: []model.PaymentMethod{model.MethodSPEI, model.MethodOXXO, model.MethodCreditCard},
				Currency:       model.CurrencyMXN,
				BaseFee:        20.0,
				PercentFee:     0.02,
				MinFee:         30.0,
				MaxFee:         3500.0,
			},
			{
				Method:         model.RefundBankTransfer,
				PaymentMethods: []model.PaymentMethod{model.MethodPSE, model.MethodEfecty, model.MethodCreditCard},
				Currency:       model.CurrencyCOP,
				BaseFee:        5000.0,
				PercentFee:     0.02,
				MinFee:         7500.0,
				MaxFee:         500000.0,
			},
		},
		DailyQuota: 2000,
		ProcessingDays: map[model.RefundMethod]int{
			model.RefundReversal:     0,
			model.RefundSameMethod:   2,
			model.RefundBankTransfer: 3,
		},
	}
}

func testProcessorQuickRefund() model.Processor {
	return model.Processor{
		ID:                  "quickrefund",
		Name:                "QuickRefund",
		SupportedCountries:  []model.Country{model.CountryBR, model.CountryMX},
		SupportedCurrencies: []model.Currency{model.CurrencyBRL, model.CurrencyMXN},
		RefundFees: []model.RefundMethodFee{
			{
				Method:         model.RefundReversal,
				PaymentMethods: []model.PaymentMethod{model.MethodCreditCard},
				Currency:       model.CurrencyBRL,
				BaseFee:        0.0,
				PercentFee:     0.0,
				MinFee:         0.0,
				MaxFee:         0,
			},
			{
				Method:         model.RefundReversal,
				PaymentMethods: []model.PaymentMethod{model.MethodCreditCard},
				Currency:       model.CurrencyMXN,
				BaseFee:        0.0,
				PercentFee:     0.0,
				MinFee:         0.0,
				MaxFee:         0,
			},
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodPIX, model.MethodCreditCard},
				Currency:       model.CurrencyBRL,
				BaseFee:        3.0,
				PercentFee:     0.03,
				MinFee:         4.5,
				MaxFee:         0,
			},
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodSPEI, model.MethodCreditCard},
				Currency:       model.CurrencyMXN,
				BaseFee:        30.0,
				PercentFee:     0.03,
				MinFee:         45.0,
				MaxFee:         0,
			},
			{
				Method:         model.RefundBankTransfer,
				PaymentMethods: []model.PaymentMethod{model.MethodPIX, model.MethodBoleto, model.MethodCreditCard},
				Currency:       model.CurrencyBRL,
				BaseFee:        2.5,
				PercentFee:     0.025,
				MinFee:         4.0,
				MaxFee:         0,
			},
			{
				Method:         model.RefundBankTransfer,
				PaymentMethods: []model.PaymentMethod{model.MethodSPEI, model.MethodOXXO, model.MethodCreditCard},
				Currency:       model.CurrencyMXN,
				BaseFee:        25.0,
				PercentFee:     0.025,
				MinFee:         40.0,
				MaxFee:         0,
			},
		},
		DailyQuota: 300,
		ProcessingDays: map[model.RefundMethod]int{
			model.RefundReversal:     0,
			model.RefundSameMethod:   0,
			model.RefundBankTransfer: 1,
		},
	}
}

func testProcessorValueProc() model.Processor {
	return model.Processor{
		ID:                  "valueproc",
		Name:                "ValueProc",
		SupportedCountries:  []model.Country{model.CountryBR, model.CountryMX, model.CountryCO},
		SupportedCurrencies: []model.Currency{model.CurrencyBRL, model.CurrencyMXN, model.CurrencyCOP},
		RefundFees: []model.RefundMethodFee{
			{
				Method:         model.RefundReversal,
				PaymentMethods: []model.PaymentMethod{model.MethodCreditCard},
				Currency:       model.CurrencyBRL,
				BaseFee:        0.0,
				PercentFee:     0.0,
				MinFee:         0.0,
				MaxFee:         0,
			},
			{
				Method:         model.RefundReversal,
				PaymentMethods: []model.PaymentMethod{model.MethodCreditCard},
				Currency:       model.CurrencyMXN,
				BaseFee:        0.0,
				PercentFee:     0.0,
				MinFee:         0.0,
				MaxFee:         0,
			},
			{
				Method:         model.RefundReversal,
				PaymentMethods: []model.PaymentMethod{model.MethodCreditCard},
				Currency:       model.CurrencyCOP,
				BaseFee:        0.0,
				PercentFee:     0.0,
				MinFee:         0.0,
				MaxFee:         0,
			},
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodPIX, model.MethodCreditCard},
				Currency:       model.CurrencyBRL,
				BaseFee:        0.5,
				PercentFee:     0.008,
				MinFee:         1.0,
				MaxFee:         80.0,
			},
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodSPEI, model.MethodCreditCard},
				Currency:       model.CurrencyMXN,
				BaseFee:        5.0,
				PercentFee:     0.008,
				MinFee:         10.0,
				MaxFee:         1400.0,
			},
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodPSE, model.MethodCreditCard},
				Currency:       model.CurrencyCOP,
				BaseFee:        1500.0,
				PercentFee:     0.008,
				MinFee:         2500.0,
				MaxFee:         200000.0,
			},
			{
				Method:         model.RefundBankTransfer,
				PaymentMethods: []model.PaymentMethod{model.MethodPIX, model.MethodBoleto, model.MethodCreditCard},
				Currency:       model.CurrencyBRL,
				BaseFee:        0.75,
				PercentFee:     0.01,
				MinFee:         1.5,
				MaxFee:         100.0,
			},
			{
				Method:         model.RefundBankTransfer,
				PaymentMethods: []model.PaymentMethod{model.MethodSPEI, model.MethodOXXO, model.MethodCreditCard},
				Currency:       model.CurrencyMXN,
				BaseFee:        8.0,
				PercentFee:     0.01,
				MinFee:         12.0,
				MaxFee:         1800.0,
			},
			{
				Method:         model.RefundBankTransfer,
				PaymentMethods: []model.PaymentMethod{model.MethodPSE, model.MethodEfecty, model.MethodCreditCard},
				Currency:       model.CurrencyCOP,
				BaseFee:        2000.0,
				PercentFee:     0.01,
				MinFee:         3500.0,
				MaxFee:         250000.0,
			},
		},
		DailyQuota: 200,
		ProcessingDays: map[model.RefundMethod]int{
			model.RefundReversal:     0,
			model.RefundSameMethod:   3,
			model.RefundBankTransfer: 5,
		},
	}
}

func allProcessors() []model.Processor {
	return []model.Processor{
		testProcessorPayBR(),
		testProcessorMexPay(),
		testProcessorColPay(),
		testProcessorGlobalPay(),
		testProcessorQuickRefund(),
		testProcessorValueProc(),
	}
}

func allCompatRules() []model.CompatibilityRule {
	return []model.CompatibilityRule{
		{
			OriginalMethod: model.MethodCreditCard,
			Country:        model.CountryBR,
			AllowedRefunds: []model.AllowedRefund{
				{Method: model.RefundReversal, MaxAgeDays: 0, RequireSettled: boolPtr(false)},
				{Method: model.RefundSameMethod, MaxAgeDays: 180},
				{Method: model.RefundBankTransfer, MaxAgeDays: 0},
			},
		},
		{
			OriginalMethod: model.MethodPIX,
			Country:        model.CountryBR,
			AllowedRefunds: []model.AllowedRefund{
				{Method: model.RefundReversal, MaxAgeDays: 0, RequireSettled: boolPtr(false)},
				{Method: model.RefundSameMethod, MaxAgeDays: 90},
				{Method: model.RefundBankTransfer, MaxAgeDays: 0},
			},
		},
		{
			OriginalMethod: model.MethodBoleto,
			Country:        model.CountryBR,
			AllowedRefunds: []model.AllowedRefund{
				{Method: model.RefundBankTransfer, MaxAgeDays: 0},
				{Method: model.RefundAccountCredit, MaxAgeDays: 0},
			},
		},
		{
			OriginalMethod: model.MethodCreditCard,
			Country:        model.CountryMX,
			AllowedRefunds: []model.AllowedRefund{
				{Method: model.RefundReversal, MaxAgeDays: 0, RequireSettled: boolPtr(false)},
				{Method: model.RefundSameMethod, MaxAgeDays: 180},
				{Method: model.RefundBankTransfer, MaxAgeDays: 0},
			},
		},
		{
			OriginalMethod: model.MethodOXXO,
			Country:        model.CountryMX,
			AllowedRefunds: []model.AllowedRefund{
				{Method: model.RefundBankTransfer, MaxAgeDays: 0},
				{Method: model.RefundAccountCredit, MaxAgeDays: 0},
			},
		},
		{
			OriginalMethod: model.MethodSPEI,
			Country:        model.CountryMX,
			AllowedRefunds: []model.AllowedRefund{
				{Method: model.RefundReversal, MaxAgeDays: 0, RequireSettled: boolPtr(false)},
				{Method: model.RefundSameMethod, MaxAgeDays: 0},
				{Method: model.RefundBankTransfer, MaxAgeDays: 0},
			},
		},
		{
			OriginalMethod: model.MethodCreditCard,
			Country:        model.CountryCO,
			AllowedRefunds: []model.AllowedRefund{
				{Method: model.RefundReversal, MaxAgeDays: 0, RequireSettled: boolPtr(false)},
				{Method: model.RefundSameMethod, MaxAgeDays: 180},
				{Method: model.RefundBankTransfer, MaxAgeDays: 0},
			},
		},
		{
			OriginalMethod: model.MethodPSE,
			Country:        model.CountryCO,
			AllowedRefunds: []model.AllowedRefund{
				{Method: model.RefundReversal, MaxAgeDays: 0, RequireSettled: boolPtr(false)},
				{Method: model.RefundSameMethod, MaxAgeDays: 60},
				{Method: model.RefundBankTransfer, MaxAgeDays: 0},
			},
		},
		{
			OriginalMethod: model.MethodEfecty,
			Country:        model.CountryCO,
			AllowedRefunds: []model.AllowedRefund{
				{Method: model.RefundBankTransfer, MaxAgeDays: 0},
				{Method: model.RefundAccountCredit, MaxAgeDays: 0},
			},
		},
	}
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.005
}

func TestSelectRoute(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	procs := allProcessors()
	rules := allCompatRules()

	tests := []struct {
		name             string
		tx               model.Transaction
		wantProcessorID  string
		wantMethod       model.RefundMethod
		wantCost         float64
		wantMinAlts      int
		wantSavingsGt0   bool
		wantNaiveCostGt0 bool
	}{
		{
			name: "settled PIX in BR routes to cheapest same-method processor PayBR",
			tx: model.Transaction{
				ID:            "tx-pix-settled",
				Country:       model.CountryBR,
				Currency:      model.CurrencyBRL,
				PaymentMethod: model.MethodPIX,
				ProcessorID:   "globalpay",
				Amount:        200.0,
				Timestamp:     now.Add(-48 * time.Hour),
				Settled:       true,
			},
			wantProcessorID:  "paybr",
			wantMethod:       model.RefundSameMethod,
			wantCost:         1.5,
			wantMinAlts:      2,
			wantSavingsGt0:   true,
			wantNaiveCostGt0: true,
		},
		{
			name: "unsettled PIX in BR within 24h gets free reversal",
			tx: model.Transaction{
				ID:            "tx-pix-reversal",
				Country:       model.CountryBR,
				Currency:      model.CurrencyBRL,
				PaymentMethod: model.MethodPIX,
				ProcessorID:   "paybr",
				Amount:        500.0,
				Timestamp:     now.Add(-2 * time.Hour),
				Settled:       false,
			},
			wantProcessorID:  "paybr",
			wantMethod:       model.RefundReversal,
			wantCost:         0.0,
			wantMinAlts:      2,
			wantSavingsGt0:   true,
			wantNaiveCostGt0: true,
		},
		{
			name: "OXXO in MX goes to bank transfer since no same-method exists",
			tx: model.Transaction{
				ID:            "tx-oxxo",
				Country:       model.CountryMX,
				Currency:      model.CurrencyMXN,
				PaymentMethod: model.MethodOXXO,
				ProcessorID:   "mexpay",
				Amount:        1000.0,
				Timestamp:     now.Add(-5 * 24 * time.Hour),
				Settled:       true,
			},
			wantProcessorID:  "valueproc",
			wantMethod:       model.RefundBankTransfer,
			wantCost:         18.0,
			wantMinAlts:      3,
			wantSavingsGt0:   false,
			wantNaiveCostGt0: true,
		},
		{
			name: "unknown payment method and country falls back to account credit",
			tx: model.Transaction{
				ID:            "tx-unknown",
				Country:       "US",
				Currency:      "USD",
				PaymentMethod: "CRYPTO",
				ProcessorID:   "nonexistent",
				Amount:        100.0,
				Timestamp:     now.Add(-1 * time.Hour),
				Settled:       false,
			},
			wantProcessorID:  "internal",
			wantMethod:       model.RefundAccountCredit,
			wantCost:         0.0,
			wantMinAlts:      0,
			wantSavingsGt0:   true,
			wantNaiveCostGt0: true,
		},
		{
			name: "EFECTY in CO goes to bank transfer with account credit as alternative",
			tx: model.Transaction{
				ID:            "tx-efecty",
				Country:       model.CountryCO,
				Currency:      model.CurrencyCOP,
				PaymentMethod: model.MethodEfecty,
				ProcessorID:   "colpay",
				Amount:        50000.0,
				Timestamp:     now.Add(-10 * 24 * time.Hour),
				Settled:       true,
			},
			wantProcessorID:  "valueproc",
			wantMethod:       model.RefundBankTransfer,
			wantCost:         3500.0,
			wantMinAlts:      2,
			wantSavingsGt0:   true,
			wantNaiveCostGt0: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := NewRouter(procs, rules)
			result := r.SelectRoute(tt.tx, now)

			if result.TransactionID != tt.tx.ID {
				t.Errorf("TransactionID = %s, want %s", result.TransactionID, tt.tx.ID)
			}
			if result.Selected.ProcessorID != tt.wantProcessorID {
				t.Errorf("Selected.ProcessorID = %s, want %s", result.Selected.ProcessorID, tt.wantProcessorID)
			}
			if result.Selected.RefundMethod != tt.wantMethod {
				t.Errorf("Selected.RefundMethod = %s, want %s", result.Selected.RefundMethod, tt.wantMethod)
			}
			if !almostEqual(result.Selected.EstimatedCost, tt.wantCost) {
				t.Errorf("Selected.EstimatedCost = %.2f, want %.2f", result.Selected.EstimatedCost, tt.wantCost)
			}
			if len(result.Alternatives) < tt.wantMinAlts {
				t.Errorf("len(Alternatives) = %d, want >= %d", len(result.Alternatives), tt.wantMinAlts)
			}
			if tt.wantSavingsGt0 && result.Savings <= 0 {
				t.Errorf("Savings = %.2f, want > 0", result.Savings)
			}
			if tt.wantNaiveCostGt0 && result.NaiveCost <= 0 {
				t.Errorf("NaiveCost = %.2f, want > 0", result.NaiveCost)
			}
			if result.Selected.Reasoning == "" {
				t.Error("Selected.Reasoning is empty")
			}
		})
	}
}

func TestSelectRoute_ReversalCostIsZero(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	tx := model.Transaction{
		ID:            "tx-reversal-free",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodPIX,
		ProcessorID:   "paybr",
		Amount:        10000.0,
		Timestamp:     now.Add(-1 * time.Hour),
		Settled:       false,
	}

	result := r.SelectRoute(tx, now)

	if result.Selected.RefundMethod != model.RefundReversal {
		t.Fatalf("Selected.RefundMethod = %s, want REVERSAL", result.Selected.RefundMethod)
	}
	if result.Selected.EstimatedCost != 0 {
		t.Errorf("Selected.EstimatedCost = %.2f, want 0", result.Selected.EstimatedCost)
	}
	if result.Selected.ProcessingDays != 0 {
		t.Errorf("Selected.ProcessingDays = %d, want 0", result.Selected.ProcessingDays)
	}
}

func TestSelectRoute_AccountCreditRankedLast(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	tx := model.Transaction{
		ID:            "tx-oxxo-credit-last",
		Country:       model.CountryMX,
		Currency:      model.CurrencyMXN,
		PaymentMethod: model.MethodOXXO,
		ProcessorID:   "mexpay",
		Amount:        500.0,
		Timestamp:     now.Add(-3 * 24 * time.Hour),
		Settled:       true,
	}

	result := r.SelectRoute(tx, now)

	if result.Selected.RefundMethod == model.RefundAccountCredit {
		t.Error("Selected route should not be ACCOUNT_CREDIT when bank transfer is available")
	}

	if len(result.Alternatives) == 0 {
		t.Fatal("expected at least one alternative")
	}

	lastAlt := result.Alternatives[len(result.Alternatives)-1]
	if lastAlt.RefundMethod != model.RefundAccountCredit {
		t.Errorf("last alternative should be ACCOUNT_CREDIT, got %s", lastAlt.RefundMethod)
	}
	if lastAlt.EstimatedCost != 0 {
		t.Errorf("ACCOUNT_CREDIT cost = %.2f, want 0", lastAlt.EstimatedCost)
	}
}

func TestSelectRoute_SavingsCalculation(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	tx := model.Transaction{
		ID:            "tx-savings",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodPIX,
		ProcessorID:   "paybr",
		Amount:        500.0,
		Timestamp:     now.Add(-1 * time.Hour),
		Settled:       false,
	}

	result := r.SelectRoute(tx, now)

	expectedSavings := result.NaiveCost - result.Selected.EstimatedCost
	if !almostEqual(result.Savings, expectedSavings) {
		t.Errorf("Savings = %.2f, want NaiveCost(%.2f) - SelectedCost(%.2f) = %.2f",
			result.Savings, result.NaiveCost, result.Selected.EstimatedCost, expectedSavings)
	}

	if result.Selected.RefundMethod == model.RefundReversal && result.Selected.EstimatedCost != 0 {
		t.Error("reversal should have zero cost")
	}
	if result.Selected.RefundMethod == model.RefundReversal && result.NaiveCost > 0 {
		if result.Savings != result.NaiveCost {
			t.Errorf("when selected is free reversal, Savings(%.2f) should equal NaiveCost(%.2f)",
				result.Savings, result.NaiveCost)
		}
	}
}

func TestSelectRoute_AlternativesSortedByCost(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	tx := model.Transaction{
		ID:            "tx-alts-sorted",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodPIX,
		ProcessorID:   "quickrefund",
		Amount:        200.0,
		Timestamp:     now.Add(-48 * time.Hour),
		Settled:       true,
	}

	result := r.SelectRoute(tx, now)

	all := append([]model.RefundCandidate{result.Selected}, result.Alternatives...)

	for i := 1; i < len(all); i++ {
		prev := all[i-1]
		curr := all[i]

		prevIsCredit := prev.RefundMethod == model.RefundAccountCredit
		currIsCredit := curr.RefundMethod == model.RefundAccountCredit

		if prevIsCredit && !currIsCredit {
			t.Errorf("ACCOUNT_CREDIT at position %d ranked before non-credit at position %d", i-1, i)
		}
		if !prevIsCredit && !currIsCredit {
			if prev.EstimatedCost > curr.EstimatedCost {
				t.Errorf("candidate[%d] cost=%.2f > candidate[%d] cost=%.2f",
					i-1, prev.EstimatedCost, i, curr.EstimatedCost)
			}
		}
	}
}

func TestSelectRoute_OriginalProcessorTiebreaker(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	procA := model.Processor{
		ID:                  "proc_a",
		Name:                "ProcA",
		SupportedCountries:  []model.Country{model.CountryBR},
		SupportedCurrencies: []model.Currency{model.CurrencyBRL},
		RefundFees: []model.RefundMethodFee{
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodPIX},
				Currency:       model.CurrencyBRL,
				BaseFee:        1.0,
				PercentFee:     0.01,
				MinFee:         0.0,
				MaxFee:         0,
			},
		},
		DailyQuota: 1000,
		ProcessingDays: map[model.RefundMethod]int{
			model.RefundSameMethod: 2,
		},
	}

	procB := model.Processor{
		ID:                  "proc_b",
		Name:                "ProcB",
		SupportedCountries:  []model.Country{model.CountryBR},
		SupportedCurrencies: []model.Currency{model.CurrencyBRL},
		RefundFees: []model.RefundMethodFee{
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodPIX},
				Currency:       model.CurrencyBRL,
				BaseFee:        1.0,
				PercentFee:     0.01,
				MinFee:         0.0,
				MaxFee:         0,
			},
		},
		DailyQuota: 1000,
		ProcessingDays: map[model.RefundMethod]int{
			model.RefundSameMethod: 2,
		},
	}

	compatRules := []model.CompatibilityRule{
		{
			OriginalMethod: model.MethodPIX,
			Country:        model.CountryBR,
			AllowedRefunds: []model.AllowedRefund{
				{Method: model.RefundSameMethod, MaxAgeDays: 90},
			},
		},
	}

	tx := model.Transaction{
		ID:            "tx-tiebreaker",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodPIX,
		ProcessorID:   "proc_b",
		Amount:        100.0,
		Timestamp:     now.Add(-48 * time.Hour),
		Settled:       true,
	}

	r := NewRouter([]model.Processor{procA, procB}, compatRules)
	result := r.SelectRoute(tx, now)

	if result.Selected.ProcessorID != "proc_b" {
		t.Errorf("Selected.ProcessorID = %s, want proc_b (original processor tiebreaker)", result.Selected.ProcessorID)
	}
}

func TestSelectRoute_OriginalProcessorTiebreaker_NoEffectWhenCostDiffers(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	cheapProc := model.Processor{
		ID:                  "cheap",
		Name:                "CheapProc",
		SupportedCountries:  []model.Country{model.CountryBR},
		SupportedCurrencies: []model.Currency{model.CurrencyBRL},
		RefundFees: []model.RefundMethodFee{
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodPIX},
				Currency:       model.CurrencyBRL,
				BaseFee:        0.5,
				PercentFee:     0.005,
				MinFee:         0.0,
				MaxFee:         0,
			},
		},
		DailyQuota: 1000,
		ProcessingDays: map[model.RefundMethod]int{
			model.RefundSameMethod: 1,
		},
	}

	expensiveProc := model.Processor{
		ID:                  "expensive",
		Name:                "ExpensiveProc",
		SupportedCountries:  []model.Country{model.CountryBR},
		SupportedCurrencies: []model.Currency{model.CurrencyBRL},
		RefundFees: []model.RefundMethodFee{
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodPIX},
				Currency:       model.CurrencyBRL,
				BaseFee:        5.0,
				PercentFee:     0.05,
				MinFee:         0.0,
				MaxFee:         0,
			},
		},
		DailyQuota: 1000,
		ProcessingDays: map[model.RefundMethod]int{
			model.RefundSameMethod: 1,
		},
	}

	compatRules := []model.CompatibilityRule{
		{
			OriginalMethod: model.MethodPIX,
			Country:        model.CountryBR,
			AllowedRefunds: []model.AllowedRefund{
				{Method: model.RefundSameMethod, MaxAgeDays: 90},
			},
		},
	}

	tx := model.Transaction{
		ID:            "tx-cost-wins",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodPIX,
		ProcessorID:   "expensive",
		Amount:        200.0,
		Timestamp:     now.Add(-48 * time.Hour),
		Settled:       true,
	}

	r := NewRouter([]model.Processor{cheapProc, expensiveProc}, compatRules)
	result := r.SelectRoute(tx, now)

	if result.Selected.ProcessorID != "cheap" {
		t.Errorf("Selected.ProcessorID = %s, want cheap (cost should win over tiebreaker)", result.Selected.ProcessorID)
	}
}

func TestSelectRoute_ProcessingDaysTiebreaker(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	fastProc := model.Processor{
		ID:                  "fast",
		Name:                "FastProc",
		SupportedCountries:  []model.Country{model.CountryBR},
		SupportedCurrencies: []model.Currency{model.CurrencyBRL},
		RefundFees: []model.RefundMethodFee{
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodPIX},
				Currency:       model.CurrencyBRL,
				BaseFee:        1.0,
				PercentFee:     0.01,
				MinFee:         0.0,
				MaxFee:         0,
			},
		},
		DailyQuota: 1000,
		ProcessingDays: map[model.RefundMethod]int{
			model.RefundSameMethod: 1,
		},
	}

	slowProc := model.Processor{
		ID:                  "slow",
		Name:                "SlowProc",
		SupportedCountries:  []model.Country{model.CountryBR},
		SupportedCurrencies: []model.Currency{model.CurrencyBRL},
		RefundFees: []model.RefundMethodFee{
			{
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodPIX},
				Currency:       model.CurrencyBRL,
				BaseFee:        1.0,
				PercentFee:     0.01,
				MinFee:         0.0,
				MaxFee:         0,
			},
		},
		DailyQuota: 1000,
		ProcessingDays: map[model.RefundMethod]int{
			model.RefundSameMethod: 5,
		},
	}

	compatRules := []model.CompatibilityRule{
		{
			OriginalMethod: model.MethodPIX,
			Country:        model.CountryBR,
			AllowedRefunds: []model.AllowedRefund{
				{Method: model.RefundSameMethod, MaxAgeDays: 90},
			},
		},
	}

	tx := model.Transaction{
		ID:            "tx-speed",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodPIX,
		ProcessorID:   "slow",
		Amount:        200.0,
		Timestamp:     now.Add(-48 * time.Hour),
		Settled:       true,
	}

	r := NewRouter([]model.Processor{slowProc, fastProc}, compatRules)
	result := r.SelectRoute(tx, now)

	if result.Selected.ProcessorID != "fast" {
		t.Errorf("Selected.ProcessorID = %s, want fast (same cost, faster processing)", result.Selected.ProcessorID)
	}
	if result.Selected.ProcessingDays != 1 {
		t.Errorf("Selected.ProcessingDays = %d, want 1", result.Selected.ProcessingDays)
	}
}

func TestSelectRoute_NoCandidatesFallsBackToAccountCredit(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	compatRules := []model.CompatibilityRule{
		{
			OriginalMethod: model.MethodPIX,
			Country:        model.CountryBR,
			AllowedRefunds: []model.AllowedRefund{
				{Method: model.RefundSameMethod, MaxAgeDays: 5},
			},
		},
	}

	tx := model.Transaction{
		ID:            "tx-expired-all",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodPIX,
		ProcessorID:   "paybr",
		Amount:        200.0,
		Timestamp:     now.Add(-10 * 24 * time.Hour),
		Settled:       true,
	}

	r := NewRouter(allProcessors(), compatRules)
	result := r.SelectRoute(tx, now)

	if result.Selected.ProcessorID != "internal" {
		t.Errorf("Selected.ProcessorID = %s, want internal", result.Selected.ProcessorID)
	}
	if result.Selected.RefundMethod != model.RefundAccountCredit {
		t.Errorf("Selected.RefundMethod = %s, want ACCOUNT_CREDIT", result.Selected.RefundMethod)
	}
	if result.Selected.EstimatedCost != 0 {
		t.Errorf("Selected.EstimatedCost = %.2f, want 0", result.Selected.EstimatedCost)
	}
	if len(result.Alternatives) != 0 {
		t.Errorf("len(Alternatives) = %d, want 0", len(result.Alternatives))
	}
	if result.Selected.Reasoning == "" {
		t.Error("Selected.Reasoning is empty")
	}
}

func TestSelectRoute_MultipleAlternativesSorted(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	tx := model.Transaction{
		ID:            "tx-multi-alts",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodPIX,
		ProcessorID:   "quickrefund",
		Amount:        1000.0,
		Timestamp:     now.Add(-48 * time.Hour),
		Settled:       true,
	}

	result := r.SelectRoute(tx, now)

	if len(result.Alternatives) < 3 {
		t.Fatalf("expected at least 3 alternatives, got %d", len(result.Alternatives))
	}

	all := append([]model.RefundCandidate{result.Selected}, result.Alternatives...)
	var nonCreditCandidates []model.RefundCandidate
	for _, c := range all {
		if c.RefundMethod != model.RefundAccountCredit {
			nonCreditCandidates = append(nonCreditCandidates, c)
		}
	}
	for i := 1; i < len(nonCreditCandidates); i++ {
		if nonCreditCandidates[i].EstimatedCost < nonCreditCandidates[i-1].EstimatedCost {
			t.Errorf("non-credit candidate[%d] cost=%.2f < candidate[%d] cost=%.2f; not sorted ascending",
				i, nonCreditCandidates[i].EstimatedCost, i-1, nonCreditCandidates[i-1].EstimatedCost)
		}
	}
}

func TestSelectRoute_PIX_BR_CostValues(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	tx := model.Transaction{
		ID:            "tx-pix-costs",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodPIX,
		ProcessorID:   "paybr",
		Amount:        200.0,
		Timestamp:     now.Add(-48 * time.Hour),
		Settled:       true,
	}

	result := r.SelectRoute(tx, now)

	paybr_same := 0.5 + 200.0*0.005
	if paybr_same < 0.75 {
		paybr_same = 0.75
	}
	paybr_same = math.Round(paybr_same*100) / 100

	if result.Selected.ProcessorID != "paybr" || result.Selected.RefundMethod != model.RefundSameMethod {
		t.Fatalf("expected PayBR SAME_METHOD as selected, got %s %s", result.Selected.ProcessorID, result.Selected.RefundMethod)
	}
	if !almostEqual(result.Selected.EstimatedCost, paybr_same) {
		t.Errorf("Selected cost = %.2f, want %.2f (PayBR PIX same-method)", result.Selected.EstimatedCost, paybr_same)
	}

	naiveCost := result.NaiveCost
	if !almostEqual(naiveCost, paybr_same) {
		t.Errorf("NaiveCost = %.2f, want %.2f (naive through paybr same-method for PIX)", naiveCost, paybr_same)
	}
}

func TestSelectRoute_OXXO_MX_BankTransferCandidates(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	tx := model.Transaction{
		ID:            "tx-oxxo-bt",
		Country:       model.CountryMX,
		Currency:      model.CurrencyMXN,
		PaymentMethod: model.MethodOXXO,
		ProcessorID:   "mexpay",
		Amount:        1000.0,
		Timestamp:     now.Add(-5 * 24 * time.Hour),
		Settled:       true,
	}

	result := r.SelectRoute(tx, now)

	all := append([]model.RefundCandidate{result.Selected}, result.Alternatives...)

	btProcessors := make(map[string]bool)
	for _, c := range all {
		if c.RefundMethod == model.RefundBankTransfer {
			btProcessors[c.ProcessorID] = true
		}
	}

	expectedBT := []string{"mexpay", "globalpay", "quickrefund", "valueproc"}
	for _, pid := range expectedBT {
		if !btProcessors[pid] {
			t.Errorf("expected bank transfer candidate from %s", pid)
		}
	}

	var hasAccountCredit bool
	for _, c := range all {
		if c.RefundMethod == model.RefundAccountCredit {
			hasAccountCredit = true
			break
		}
	}
	if !hasAccountCredit {
		t.Error("expected ACCOUNT_CREDIT in candidates for OXXO")
	}
}

func TestSelectRoute_EmptyProcessorList(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	tx := model.Transaction{
		ID:            "tx-no-procs",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodPIX,
		ProcessorID:   "paybr",
		Amount:        200.0,
		Timestamp:     now.Add(-48 * time.Hour),
		Settled:       true,
	}

	r := NewRouter(nil, allCompatRules())
	result := r.SelectRoute(tx, now)

	if result.Selected.ProcessorID != "internal" {
		t.Errorf("Selected.ProcessorID = %s, want internal (no processors available)", result.Selected.ProcessorID)
	}
	if result.Selected.RefundMethod != model.RefundAccountCredit {
		t.Errorf("Selected.RefundMethod = %s, want ACCOUNT_CREDIT", result.Selected.RefundMethod)
	}
}

func TestSelectRoute_TransactionIDPassedThrough(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	tests := []struct {
		name string
		txID string
	}{
		{name: "standard id", txID: "tx-12345"},
		{name: "uuid style", txID: "550e8400-e29b-41d4-a716-446655440000"},
		{name: "empty id", txID: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tx := model.Transaction{
				ID:            tt.txID,
				Country:       model.CountryBR,
				Currency:      model.CurrencyBRL,
				PaymentMethod: model.MethodPIX,
				ProcessorID:   "paybr",
				Amount:        100.0,
				Timestamp:     now.Add(-48 * time.Hour),
				Settled:       true,
			}
			result := r.SelectRoute(tx, now)
			if result.TransactionID != tt.txID {
				t.Errorf("TransactionID = %s, want %s", result.TransactionID, tt.txID)
			}
		})
	}
}

func TestSelectRoute_AccountCreditAlwaysLastInOXXOCandidates(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	amounts := []float64{100, 500, 1000, 5000, 50000}

	for _, amt := range amounts {
		tx := model.Transaction{
			ID:            "tx-oxxo-credit-pos",
			Country:       model.CountryMX,
			Currency:      model.CurrencyMXN,
			PaymentMethod: model.MethodOXXO,
			ProcessorID:   "mexpay",
			Amount:        amt,
			Timestamp:     now.Add(-5 * 24 * time.Hour),
			Settled:       true,
		}

		result := r.SelectRoute(tx, now)
		all := append([]model.RefundCandidate{result.Selected}, result.Alternatives...)

		creditIdx := -1
		for i, c := range all {
			if c.RefundMethod == model.RefundAccountCredit {
				creditIdx = i
				break
			}
		}

		if creditIdx == -1 {
			t.Errorf("amount=%.0f: ACCOUNT_CREDIT not found in candidates", amt)
			continue
		}

		for i := creditIdx + 1; i < len(all); i++ {
			if all[i].RefundMethod != model.RefundAccountCredit {
				t.Errorf("amount=%.0f: non-ACCOUNT_CREDIT candidate found after ACCOUNT_CREDIT at position %d", amt, i)
			}
		}
	}
}

func TestBuildReasoning(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	tests := []struct {
		name            string
		tx              model.Transaction
		wantSubstrings  []string
	}{
		{
			name: "reversal reasoning mentions free and processor name",
			tx: model.Transaction{
				ID:            "tx-reason-rev",
				Country:       model.CountryBR,
				Currency:      model.CurrencyBRL,
				PaymentMethod: model.MethodPIX,
				ProcessorID:   "paybr",
				Amount:        200.0,
				Timestamp:     now.Add(-2 * time.Hour),
				Settled:       false,
			},
			wantSubstrings: []string{"Free reversal", "PayBR"},
		},
		{
			name: "same method reasoning mentions method and cost",
			tx: model.Transaction{
				ID:            "tx-reason-same",
				Country:       model.CountryBR,
				Currency:      model.CurrencyBRL,
				PaymentMethod: model.MethodPIX,
				ProcessorID:   "paybr",
				Amount:        200.0,
				Timestamp:     now.Add(-48 * time.Hour),
				Settled:       true,
			},
			wantSubstrings: []string{"PIX-to-PIX", "BRL"},
		},
		{
			name: "bank transfer reasoning mentions bank transfer",
			tx: model.Transaction{
				ID:            "tx-reason-bt",
				Country:       model.CountryMX,
				Currency:      model.CurrencyMXN,
				PaymentMethod: model.MethodOXXO,
				ProcessorID:   "mexpay",
				Amount:        1000.0,
				Timestamp:     now.Add(-5 * 24 * time.Hour),
				Settled:       true,
			},
			wantSubstrings: []string{"bank transfer"},
		},
		{
			name: "account credit reasoning mentions account credit",
			tx: model.Transaction{
				ID:            "tx-reason-credit",
				Country:       "US",
				Currency:      "USD",
				PaymentMethod: "CRYPTO",
				ProcessorID:   "nonexistent",
				Amount:        100.0,
				Timestamp:     now.Add(-1 * time.Hour),
				Settled:       false,
			},
			wantSubstrings: []string{"account credit"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := r.SelectRoute(tt.tx, now)

			reasoning := result.Selected.Reasoning
			if reasoning == "" {
				t.Fatal("Selected.Reasoning is empty")
			}

			for _, substr := range tt.wantSubstrings {
				if !strings.Contains(strings.ToLower(reasoning), strings.ToLower(substr)) {
					t.Errorf("reasoning %q does not contain %q", reasoning, substr)
				}
			}
		})
	}
}

func TestBuildReasoning_AllCandidatesHaveNonEmptyReasoning(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	txs := []model.Transaction{
		{
			ID: "tx-r1", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 200.0,
			Timestamp: now.Add(-2 * time.Hour), Settled: false,
		},
		{
			ID: "tx-r2", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodOXXO, ProcessorID: "mexpay", Amount: 500.0,
			Timestamp: now.Add(-5 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "tx-r3", Country: model.CountryCO, Currency: model.CurrencyCOP,
			PaymentMethod: model.MethodPSE, ProcessorID: "colpay", Amount: 100000.0,
			Timestamp: now.Add(-3 * time.Hour), Settled: false,
		},
	}

	for _, tx := range txs {
		result := r.SelectRoute(tx, now)
		if result.Selected.Reasoning == "" {
			t.Errorf("tx=%s: Selected.Reasoning is empty", tx.ID)
		}
		for i, alt := range result.Alternatives {
			if alt.Reasoning == "" {
				t.Errorf("tx=%s: Alternatives[%d].Reasoning is empty (processor=%s, method=%s)",
					tx.ID, i, alt.ProcessorID, alt.RefundMethod)
			}
		}
	}
}

func TestBuildReasoning_ReversalFormat(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	proc := testProcessorPayBR()
	fee := model.RefundMethodFee{
		Method:     model.RefundReversal,
		BaseFee:    0,
		PercentFee: 0,
	}
	path := rules.EligiblePath{
		Method: model.RefundReversal,
		Reason: "Transaction is 2.0 hours old and unsettled; free reversal available",
	}
	tx := model.Transaction{
		ID:            "tx-rev-fmt",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodPIX,
		ProcessorID:   "paybr",
		Amount:        500.0,
		Timestamp:     now.Add(-2 * time.Hour),
	}

	reasoning := buildReasoning(tx, proc, path, fee, 0, 0)

	if !strings.Contains(reasoning, "Free reversal") {
		t.Errorf("reversal reasoning should contain 'Free reversal', got %q", reasoning)
	}
	if !strings.Contains(reasoning, "PayBR") {
		t.Errorf("reversal reasoning should contain processor name 'PayBR', got %q", reasoning)
	}
}

func TestBuildReasoning_SameMethodFormat(t *testing.T) {
	t.Parallel()

	proc := testProcessorPayBR()
	fee := model.RefundMethodFee{
		Method:     model.RefundSameMethod,
		BaseFee:    0.5,
		PercentFee: 0.005,
		MinFee:     0.75,
		MaxFee:     0,
	}
	path := rules.EligiblePath{
		Method: model.RefundSameMethod,
		Reason: "Within SAME_METHOD window (2 of 90 days used, 88 remaining)",
	}
	tx := model.Transaction{
		ID:            "tx-same-fmt",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodPIX,
		ProcessorID:   "paybr",
		Amount:        200.0,
	}

	reasoning := buildReasoning(tx, proc, path, fee, 1.50, 1)

	if !strings.Contains(reasoning, "PIX-to-PIX") {
		t.Errorf("same-method reasoning should contain 'PIX-to-PIX', got %q", reasoning)
	}
	if !strings.Contains(reasoning, "PayBR") {
		t.Errorf("same-method reasoning should contain processor name, got %q", reasoning)
	}
	if !strings.Contains(reasoning, "BRL") {
		t.Errorf("same-method reasoning should contain currency, got %q", reasoning)
	}
	if !strings.Contains(reasoning, "1 day") {
		t.Errorf("same-method reasoning should contain '1 day', got %q", reasoning)
	}
}

func TestBuildReasoning_BankTransferFormat(t *testing.T) {
	t.Parallel()

	proc := testProcessorPayBR()
	fee := model.RefundMethodFee{
		Method:     model.RefundBankTransfer,
		BaseFee:    1.0,
		PercentFee: 0.015,
		MinFee:     1.5,
		MaxFee:     100.0,
	}
	path := rules.EligiblePath{
		Method: model.RefundBankTransfer,
		Reason: "No time limit for this refund method",
	}
	tx := model.Transaction{
		ID:            "tx-bt-fmt",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodPIX,
		ProcessorID:   "paybr",
		Amount:        1000.0,
	}

	reasoning := buildReasoning(tx, proc, path, fee, 16.0, 2)

	if !strings.Contains(reasoning, "bank transfer") {
		t.Errorf("bank transfer reasoning should contain 'bank transfer', got %q", reasoning)
	}
	if !strings.Contains(reasoning, "2 days") {
		t.Errorf("bank transfer reasoning should contain '2 days', got %q", reasoning)
	}
}

func TestBuildReasoning_InstantProcessing(t *testing.T) {
	t.Parallel()

	proc := testProcessorQuickRefund()
	fee := model.RefundMethodFee{
		Method:     model.RefundSameMethod,
		BaseFee:    3.0,
		PercentFee: 0.03,
		MinFee:     4.5,
	}
	path := rules.EligiblePath{
		Method: model.RefundSameMethod,
		Reason: "Within SAME_METHOD window",
	}
	tx := model.Transaction{
		ID:            "tx-instant",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodPIX,
		ProcessorID:   "quickrefund",
		Amount:        100.0,
	}

	reasoning := buildReasoning(tx, proc, path, fee, 6.0, 0)

	if !strings.Contains(reasoning, "instant") {
		t.Errorf("zero-day processing should say 'instant', got %q", reasoning)
	}
}

func TestBuildReasoning_CostDescriptions(t *testing.T) {
	t.Parallel()

	proc := testProcessorPayBR()
	tx := model.Transaction{
		ID: "tx-cost-desc", Country: model.CountryBR, Currency: model.CurrencyBRL,
		PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 200.0,
	}
	path := rules.EligiblePath{Method: model.RefundSameMethod, Reason: "test"}

	tests := []struct {
		name           string
		fee            model.RefundMethodFee
		cost           float64
		wantSubstrings []string
	}{
		{
			name: "base plus percent",
			fee: model.RefundMethodFee{
				Method: model.RefundSameMethod, BaseFee: 0.5, PercentFee: 0.005,
			},
			cost:           1.50,
			wantSubstrings: []string{"0.50 base", "0.5%", "1.50"},
		},
		{
			name: "percent only",
			fee: model.RefundMethodFee{
				Method: model.RefundSameMethod, BaseFee: 0, PercentFee: 0.015,
			},
			cost:           3.0,
			wantSubstrings: []string{"1.5%", "3.00"},
		},
		{
			name: "flat fee only",
			fee: model.RefundMethodFee{
				Method: model.RefundSameMethod, BaseFee: 5.0, PercentFee: 0,
			},
			cost:           5.0,
			wantSubstrings: []string{"5.00 BRL"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reasoning := buildReasoning(tx, proc, path, tt.fee, tt.cost, 1)
			for _, substr := range tt.wantSubstrings {
				if !strings.Contains(reasoning, substr) {
					t.Errorf("reasoning %q does not contain %q", reasoning, substr)
				}
			}
		})
	}
}

func TestSelectRoute_NaiveCostUsesOriginalProcessor(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	tx := model.Transaction{
		ID:            "tx-naive-check",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodPIX,
		ProcessorID:   "quickrefund",
		Amount:        200.0,
		Timestamp:     now.Add(-48 * time.Hour),
		Settled:       true,
	}

	result := r.SelectRoute(tx, now)

	quickRefundSameMethod := 3.0 + 200.0*0.03
	if quickRefundSameMethod < 4.5 {
		quickRefundSameMethod = 4.5
	}
	quickRefundSameMethod = math.Round(quickRefundSameMethod*100) / 100

	if !almostEqual(result.NaiveCost, quickRefundSameMethod) {
		t.Errorf("NaiveCost = %.2f, want %.2f (naive cost through quickrefund same-method)",
			result.NaiveCost, quickRefundSameMethod)
	}
}

func TestSelectRoute_ProcessorNamePopulated(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	tx := model.Transaction{
		ID:            "tx-proc-name",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodPIX,
		ProcessorID:   "paybr",
		Amount:        200.0,
		Timestamp:     now.Add(-48 * time.Hour),
		Settled:       true,
	}

	result := r.SelectRoute(tx, now)

	if result.Selected.ProcessorName == "" {
		t.Error("Selected.ProcessorName is empty")
	}

	for i, alt := range result.Alternatives {
		if alt.ProcessorName == "" {
			t.Errorf("Alternatives[%d].ProcessorName is empty", i)
		}
	}
}

func TestSelectRoute_CreditCardBR_UnsettledFreshGetsReversal(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	tx := model.Transaction{
		ID:            "tx-cc-br-rev",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodCreditCard,
		ProcessorID:   "paybr",
		Amount:        1000.0,
		Timestamp:     now.Add(-3 * time.Hour),
		Settled:       false,
	}

	result := r.SelectRoute(tx, now)

	if result.Selected.RefundMethod != model.RefundReversal {
		t.Errorf("Selected.RefundMethod = %s, want REVERSAL for fresh unsettled CC", result.Selected.RefundMethod)
	}
	if result.Selected.EstimatedCost != 0 {
		t.Errorf("Reversal cost = %.2f, want 0", result.Selected.EstimatedCost)
	}
}

func TestSelectRoute_BoletoBR_NoReversalOrSameMethod(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	tx := model.Transaction{
		ID:            "tx-boleto",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodBoleto,
		ProcessorID:   "paybr",
		Amount:        300.0,
		Timestamp:     now.Add(-5 * 24 * time.Hour),
		Settled:       true,
	}

	result := r.SelectRoute(tx, now)

	all := append([]model.RefundCandidate{result.Selected}, result.Alternatives...)
	for _, c := range all {
		if c.RefundMethod == model.RefundReversal {
			t.Error("BOLETO should not have REVERSAL as candidate")
		}
		if c.RefundMethod == model.RefundSameMethod {
			t.Error("BOLETO should not have SAME_METHOD as candidate")
		}
	}

	if result.Selected.RefundMethod != model.RefundBankTransfer {
		t.Errorf("Selected.RefundMethod = %s, want BANK_TRANSFER for BOLETO", result.Selected.RefundMethod)
	}
}

func TestSelectRoute_LargeAmount_MaxFeeCap(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	tx := model.Transaction{
		ID:            "tx-large-cc",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodCreditCard,
		ProcessorID:   "paybr",
		Amount:        50000.0,
		Timestamp:     now.Add(-48 * time.Hour),
		Settled:       true,
	}

	result := r.SelectRoute(tx, now)

	all := append([]model.RefundCandidate{result.Selected}, result.Alternatives...)

	paybr_sm_cap := 150.0
	paybr_bt_cap := 100.0
	globalpay_cap := 200.0

	for _, c := range all {
		if c.RefundMethod == model.RefundAccountCredit {
			continue
		}
		switch c.ProcessorID {
		case "paybr":
			if c.RefundMethod == model.RefundSameMethod && c.EstimatedCost > paybr_sm_cap {
				t.Errorf("paybr SAME_METHOD cost %.2f exceeds cap %.2f", c.EstimatedCost, paybr_sm_cap)
			}
			if c.RefundMethod == model.RefundBankTransfer && c.EstimatedCost > paybr_bt_cap {
				t.Errorf("paybr BANK_TRANSFER cost %.2f exceeds cap %.2f", c.EstimatedCost, paybr_bt_cap)
			}
		case "globalpay":
			if c.EstimatedCost > globalpay_cap {
				t.Errorf("globalpay %s cost %.2f exceeds cap %.2f", c.RefundMethod, c.EstimatedCost, globalpay_cap)
			}
		case "valueproc":
			if c.RefundMethod == model.RefundSameMethod && c.EstimatedCost > 80.0 {
				t.Errorf("valueproc SAME_METHOD cost %.2f exceeds cap 80.0", c.EstimatedCost)
			}
			if c.RefundMethod == model.RefundBankTransfer && c.EstimatedCost > 100.0 {
				t.Errorf("valueproc BANK_TRANSFER cost %.2f exceeds cap 100.0", c.EstimatedCost)
			}
		}
	}
}

func TestSelectRoute_EligiblePath_AccountCreditExplicitInRules(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	tx := model.Transaction{
		ID:            "tx-oxxo-ac",
		Country:       model.CountryMX,
		Currency:      model.CurrencyMXN,
		PaymentMethod: model.MethodOXXO,
		ProcessorID:   "mexpay",
		Amount:        500.0,
		Timestamp:     now.Add(-5 * 24 * time.Hour),
		Settled:       true,
	}

	result := r.SelectRoute(tx, now)

	all := append([]model.RefundCandidate{result.Selected}, result.Alternatives...)
	var creditFound bool
	for _, c := range all {
		if c.RefundMethod == model.RefundAccountCredit {
			creditFound = true
			if c.ProcessorID != "internal" {
				t.Errorf("ACCOUNT_CREDIT ProcessorID = %s, want internal", c.ProcessorID)
			}
			if c.ProcessorName != "Account Credit" {
				t.Errorf("ACCOUNT_CREDIT ProcessorName = %s, want 'Account Credit'", c.ProcessorName)
			}
			if c.EstimatedCost != 0 {
				t.Errorf("ACCOUNT_CREDIT cost = %.2f, want 0", c.EstimatedCost)
			}
			if c.ProcessingDays != 0 {
				t.Errorf("ACCOUNT_CREDIT ProcessingDays = %d, want 0", c.ProcessingDays)
			}
		}
	}
	if !creditFound {
		t.Error("ACCOUNT_CREDIT not found in candidates for OXXO")
	}
}

func TestSelectRoute_PSE_CO_SettledGetsMultipleCandidates(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	tx := model.Transaction{
		ID:            "tx-pse-co",
		Country:       model.CountryCO,
		Currency:      model.CurrencyCOP,
		PaymentMethod: model.MethodPSE,
		ProcessorID:   "colpay",
		Amount:        100000.0,
		Timestamp:     now.Add(-10 * 24 * time.Hour),
		Settled:       true,
	}

	result := r.SelectRoute(tx, now)

	all := append([]model.RefundCandidate{result.Selected}, result.Alternatives...)
	methodSet := make(map[model.RefundMethod]bool)
	for _, c := range all {
		methodSet[c.RefundMethod] = true
	}

	if !methodSet[model.RefundSameMethod] {
		t.Error("expected SAME_METHOD in candidates for PSE CO within 60-day window")
	}
	if !methodSet[model.RefundBankTransfer] {
		t.Error("expected BANK_TRANSFER in candidates for PSE CO")
	}
}

func TestSelectRoute_ResultStructure(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := NewRouter(allProcessors(), allCompatRules())

	tx := model.Transaction{
		ID:            "tx-structure",
		Country:       model.CountryBR,
		Currency:      model.CurrencyBRL,
		PaymentMethod: model.MethodPIX,
		ProcessorID:   "paybr",
		Amount:        500.0,
		Timestamp:     now.Add(-2 * time.Hour),
		Settled:       false,
	}

	result := r.SelectRoute(tx, now)

	if result.Selected.ProcessorID == "" {
		t.Error("Selected.ProcessorID is empty")
	}
	if result.Selected.ProcessorName == "" {
		t.Error("Selected.ProcessorName is empty")
	}
	if result.Selected.RefundMethod == "" {
		t.Error("Selected.RefundMethod is empty")
	}
	if result.Selected.Reasoning == "" {
		t.Error("Selected.Reasoning is empty")
	}
	if result.Selected.EstimatedCost < 0 {
		t.Errorf("Selected.EstimatedCost = %.2f, want >= 0", result.Selected.EstimatedCost)
	}
	if result.Selected.ProcessingDays < 0 {
		t.Errorf("Selected.ProcessingDays = %d, want >= 0", result.Selected.ProcessingDays)
	}
}

func TestNewRouter(t *testing.T) {
	t.Parallel()

	procs := allProcessors()
	rules := allCompatRules()

	r := NewRouter(procs, rules)

	if r == nil {
		t.Fatal("NewRouter() returned nil")
	}
	if len(r.Processors) != len(procs) {
		t.Errorf("len(Processors) = %d, want %d", len(r.Processors), len(procs))
	}
	if r.RuleIndex == nil {
		t.Error("RuleIndex is nil")
	}
}

func TestNewRouter_EmptyInputs(t *testing.T) {
	t.Parallel()

	r := NewRouter(nil, nil)
	if r == nil {
		t.Fatal("NewRouter(nil, nil) returned nil")
	}
	if r.RuleIndex == nil {
		t.Error("RuleIndex is nil even with nil rules")
	}
}
