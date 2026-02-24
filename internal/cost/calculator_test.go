package cost

import (
	"math"
	"testing"

	"github.com/ivanjtm/YunoChallenge/internal/model"
)

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
				Method:         model.RefundSameMethod,
				PaymentMethods: []model.PaymentMethod{model.MethodCreditCard},
				Currency:       model.CurrencyBRL,
				BaseFee:        2.0,
				PercentFee:     0.02,
				MinFee:         3.0,
				MaxFee:         200.0,
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
		},
		DailyQuota: 2000,
		ProcessingDays: map[model.RefundMethod]int{
			model.RefundReversal:     0,
			model.RefundSameMethod:   2,
			model.RefundBankTransfer: 3,
		},
	}
}

func TestCalculate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		amount float64
		fee    model.RefundMethodFee
		want   float64
	}{
		{
			name:   "normal fee calculation",
			amount: 200.0,
			fee: model.RefundMethodFee{
				Method:     model.RefundSameMethod,
				BaseFee:    1.5,
				PercentFee: 0.025,
				MinFee:     2.0,
				MaxFee:     150.0,
			},
			want: 6.5,
		},
		{
			name:   "min fee floor applied",
			amount: 10.0,
			fee: model.RefundMethodFee{
				Method:     model.RefundSameMethod,
				BaseFee:    0.5,
				PercentFee: 0.005,
				MinFee:     0.75,
				MaxFee:     0,
			},
			want: 0.75,
		},
		{
			name:   "max fee cap applied",
			amount: 50000.0,
			fee: model.RefundMethodFee{
				Method:     model.RefundSameMethod,
				BaseFee:    1.5,
				PercentFee: 0.025,
				MinFee:     2.0,
				MaxFee:     150.0,
			},
			want: 150.0,
		},
		{
			name:   "zero amount uses min fee",
			amount: 0.0,
			fee: model.RefundMethodFee{
				Method:     model.RefundSameMethod,
				BaseFee:    1.5,
				PercentFee: 0.025,
				MinFee:     2.0,
				MaxFee:     150.0,
			},
			want: 2.0,
		},
		{
			name:   "zero amount with zero base and min fee",
			amount: 0.0,
			fee: model.RefundMethodFee{
				Method:     model.RefundSameMethod,
				BaseFee:    0.0,
				PercentFee: 0.01,
				MinFee:     0.0,
				MaxFee:     100.0,
			},
			want: 0.0,
		},
		{
			name:   "reversal is always free",
			amount: 5000.0,
			fee: model.RefundMethodFee{
				Method:     model.RefundReversal,
				BaseFee:    10.0,
				PercentFee: 0.05,
				MinFee:     5.0,
				MaxFee:     500.0,
			},
			want: 0.0,
		},
		{
			name:   "reversal with zero amount is free",
			amount: 0.0,
			fee: model.RefundMethodFee{
				Method:     model.RefundReversal,
				BaseFee:    0.0,
				PercentFee: 0.0,
				MinFee:     0.0,
				MaxFee:     0,
			},
			want: 0.0,
		},
		{
			name:   "account credit is always free",
			amount: 3000.0,
			fee: model.RefundMethodFee{
				Method:     model.RefundAccountCredit,
				BaseFee:    5.0,
				PercentFee: 0.02,
				MinFee:     3.0,
				MaxFee:     200.0,
			},
			want: 0.0,
		},
		{
			name:   "max fee zero means no cap",
			amount: 100000.0,
			fee: model.RefundMethodFee{
				Method:     model.RefundSameMethod,
				BaseFee:    5.0,
				PercentFee: 0.008,
				MinFee:     8.0,
				MaxFee:     0,
			},
			want: 805.0,
		},
		{
			name:   "cost exactly at min fee boundary",
			amount: 50.0,
			fee: model.RefundMethodFee{
				Method:     model.RefundSameMethod,
				BaseFee:    0.5,
				PercentFee: 0.005,
				MinFee:     0.75,
				MaxFee:     0,
			},
			want: 0.75,
		},
		{
			name:   "cost exactly at max fee boundary",
			amount: 5920.0,
			fee: model.RefundMethodFee{
				Method:     model.RefundSameMethod,
				BaseFee:    1.5,
				PercentFee: 0.025,
				MinFee:     2.0,
				MaxFee:     150.0,
			},
			want: 149.5,
		},
		{
			name:   "result is rounded to two decimals",
			amount: 33.33,
			fee: model.RefundMethodFee{
				Method:     model.RefundBankTransfer,
				BaseFee:    1.0,
				PercentFee: 0.015,
				MinFee:     1.5,
				MaxFee:     100.0,
			},
			want: 1.5,
		},
		{
			name:   "large COP amount below cap",
			amount: 15000000.0,
			fee: model.RefundMethodFee{
				Method:     model.RefundSameMethod,
				BaseFee:    3500.0,
				PercentFee: 0.022,
				MinFee:     5000.0,
				MaxFee:     350000.0,
			},
			want: 333500.0,
		},
		{
			name:   "large COP amount hits cap",
			amount: 20000000.0,
			fee: model.RefundMethodFee{
				Method:     model.RefundSameMethod,
				BaseFee:    3500.0,
				PercentFee: 0.022,
				MinFee:     5000.0,
				MaxFee:     350000.0,
			},
			want: 350000.0,
		},
		{
			name:   "bank transfer mid range",
			amount: 1000.0,
			fee: model.RefundMethodFee{
				Method:     model.RefundBankTransfer,
				BaseFee:    10.0,
				PercentFee: 0.012,
				MinFee:     15.0,
				MaxFee:     1800.0,
			},
			want: 22.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Calculate(tt.amount, tt.fee)
			if got != tt.want {
				t.Errorf("Calculate(%v, %+v) = %v, want %v", tt.amount, tt.fee.Method, got, tt.want)
			}
		})
	}
}

func TestFindMatchingFee(t *testing.T) {
	t.Parallel()

	paybr := testProcessorPayBR()
	mexpay := testProcessorMexPay()
	globalpay := testProcessorGlobalPay()

	tests := []struct {
		name           string
		proc           model.Processor
		refundMethod   model.RefundMethod
		originalMethod model.PaymentMethod
		currency       model.Currency
		wantNil        bool
		wantMethod     model.RefundMethod
	}{
		{
			name:           "paybr reversal credit card BRL",
			proc:           paybr,
			refundMethod:   model.RefundReversal,
			originalMethod: model.MethodCreditCard,
			currency:       model.CurrencyBRL,
			wantNil:        false,
			wantMethod:     model.RefundReversal,
		},
		{
			name:           "paybr same method PIX BRL",
			proc:           paybr,
			refundMethod:   model.RefundSameMethod,
			originalMethod: model.MethodPIX,
			currency:       model.CurrencyBRL,
			wantNil:        false,
			wantMethod:     model.RefundSameMethod,
		},
		{
			name:           "paybr bank transfer boleto BRL",
			proc:           paybr,
			refundMethod:   model.RefundBankTransfer,
			originalMethod: model.MethodBoleto,
			currency:       model.CurrencyBRL,
			wantNil:        false,
			wantMethod:     model.RefundBankTransfer,
		},
		{
			name:           "paybr no match for OXXO",
			proc:           paybr,
			refundMethod:   model.RefundSameMethod,
			originalMethod: model.MethodOXXO,
			currency:       model.CurrencyBRL,
			wantNil:        true,
		},
		{
			name:           "paybr no match for wrong currency",
			proc:           paybr,
			refundMethod:   model.RefundSameMethod,
			originalMethod: model.MethodPIX,
			currency:       model.CurrencyMXN,
			wantNil:        true,
		},
		{
			name:           "mexpay same method SPEI MXN",
			proc:           mexpay,
			refundMethod:   model.RefundSameMethod,
			originalMethod: model.MethodSPEI,
			currency:       model.CurrencyMXN,
			wantNil:        false,
			wantMethod:     model.RefundSameMethod,
		},
		{
			name:           "mexpay bank transfer OXXO MXN",
			proc:           mexpay,
			refundMethod:   model.RefundBankTransfer,
			originalMethod: model.MethodOXXO,
			currency:       model.CurrencyMXN,
			wantNil:        false,
			wantMethod:     model.RefundBankTransfer,
		},
		{
			name:           "mexpay no match for non-existent refund method",
			proc:           mexpay,
			refundMethod:   model.RefundAccountCredit,
			originalMethod: model.MethodCreditCard,
			currency:       model.CurrencyMXN,
			wantNil:        true,
		},
		{
			name:           "globalpay reversal credit card BRL",
			proc:           globalpay,
			refundMethod:   model.RefundReversal,
			originalMethod: model.MethodCreditCard,
			currency:       model.CurrencyBRL,
			wantNil:        false,
			wantMethod:     model.RefundReversal,
		},
		{
			name:           "globalpay bank transfer PIX BRL",
			proc:           globalpay,
			refundMethod:   model.RefundBankTransfer,
			originalMethod: model.MethodPIX,
			currency:       model.CurrencyBRL,
			wantNil:        false,
			wantMethod:     model.RefundBankTransfer,
		},
		{
			name:           "globalpay no reversal for PIX",
			proc:           globalpay,
			refundMethod:   model.RefundReversal,
			originalMethod: model.MethodPIX,
			currency:       model.CurrencyBRL,
			wantNil:        true,
		},
		{
			name:           "empty processor returns nil",
			proc:           model.Processor{},
			refundMethod:   model.RefundSameMethod,
			originalMethod: model.MethodCreditCard,
			currency:       model.CurrencyBRL,
			wantNil:        true,
		},
		{
			name: "fee with empty currency matches any currency",
			proc: model.Processor{
				ID: "wildcard",
				RefundFees: []model.RefundMethodFee{
					{
						Method:         model.RefundSameMethod,
						PaymentMethods: []model.PaymentMethod{model.MethodCreditCard},
						Currency:       "",
						BaseFee:        1.0,
						PercentFee:     0.01,
						MinFee:         0.5,
						MaxFee:         50.0,
					},
				},
			},
			refundMethod:   model.RefundSameMethod,
			originalMethod: model.MethodCreditCard,
			currency:       model.CurrencyCOP,
			wantNil:        false,
			wantMethod:     model.RefundSameMethod,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FindMatchingFee(tt.proc, tt.refundMethod, tt.originalMethod, tt.currency)
			if tt.wantNil {
				if got != nil {
					t.Errorf("FindMatchingFee() = %+v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("FindMatchingFee() = nil, want non-nil")
			}
			if got.Method != tt.wantMethod {
				t.Errorf("FindMatchingFee().Method = %v, want %v", got.Method, tt.wantMethod)
			}
		})
	}
}

func TestFindMatchingFee_ReturnsPointerToOriginalSlice(t *testing.T) {
	t.Parallel()

	proc := testProcessorPayBR()
	fee := FindMatchingFee(proc, model.RefundSameMethod, model.MethodPIX, model.CurrencyBRL)
	if fee == nil {
		t.Fatal("FindMatchingFee() = nil, want non-nil")
	}
	if fee != &proc.RefundFees[1] {
		t.Error("FindMatchingFee() did not return pointer to original slice element")
	}
}

func TestCalculateNaive(t *testing.T) {
	t.Parallel()

	paybr := testProcessorPayBR()
	mexpay := testProcessorMexPay()
	processors := []model.Processor{paybr, mexpay}

	tests := []struct {
		name string
		tx   model.Transaction
		want float64
	}{
		{
			name: "uses same method fee for credit card via paybr",
			tx: model.Transaction{
				ID:            "tx-1",
				ProcessorID:   "paybr",
				PaymentMethod: model.MethodCreditCard,
				Currency:      model.CurrencyBRL,
				Amount:        500.0,
			},
			want: 14.0,
		},
		{
			name: "uses same method fee for PIX via paybr",
			tx: model.Transaction{
				ID:            "tx-2",
				ProcessorID:   "paybr",
				PaymentMethod: model.MethodPIX,
				Currency:      model.CurrencyBRL,
				Amount:        200.0,
			},
			want: 1.5,
		},
		{
			name: "falls back to bank transfer when same method not available",
			tx: model.Transaction{
				ID:            "tx-3",
				ProcessorID:   "paybr",
				PaymentMethod: model.MethodBoleto,
				Currency:      model.CurrencyBRL,
				Amount:        1000.0,
			},
			want: 16.0,
		},
		{
			name: "unknown processor falls back to 3.5 percent",
			tx: model.Transaction{
				ID:            "tx-4",
				ProcessorID:   "unknown",
				PaymentMethod: model.MethodCreditCard,
				Currency:      model.CurrencyBRL,
				Amount:        1000.0,
			},
			want: 35.0,
		},
		{
			name: "mexpay credit card same method",
			tx: model.Transaction{
				ID:            "tx-5",
				ProcessorID:   "mexpay",
				PaymentMethod: model.MethodCreditCard,
				Currency:      model.CurrencyMXN,
				Amount:        2000.0,
			},
			want: 55.0,
		},
		{
			name: "mexpay OXXO falls back to bank transfer",
			tx: model.Transaction{
				ID:            "tx-6",
				ProcessorID:   "mexpay",
				PaymentMethod: model.MethodOXXO,
				Currency:      model.CurrencyMXN,
				Amount:        500.0,
			},
			want: 16.0,
		},
		{
			name: "unknown processor with zero amount",
			tx: model.Transaction{
				ID:            "tx-7",
				ProcessorID:   "nonexistent",
				PaymentMethod: model.MethodPIX,
				Currency:      model.CurrencyBRL,
				Amount:        0.0,
			},
			want: 0.0,
		},
		{
			name: "paybr credit card hits max fee cap",
			tx: model.Transaction{
				ID:            "tx-8",
				ProcessorID:   "paybr",
				PaymentMethod: model.MethodCreditCard,
				Currency:      model.CurrencyBRL,
				Amount:        50000.0,
			},
			want: 150.0,
		},
		{
			name: "no matching payment method or bank transfer falls back to 3.5 percent",
			tx: model.Transaction{
				ID:            "tx-9",
				ProcessorID:   "mexpay",
				PaymentMethod: model.MethodPSE,
				Currency:      model.CurrencyMXN,
				Amount:        1000.0,
			},
			want: 35.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CalculateNaive(tt.tx, processors)
			if got != tt.want {
				t.Errorf("CalculateNaive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateNaive_EmptyProcessorList(t *testing.T) {
	t.Parallel()

	tx := model.Transaction{
		ID:            "tx-empty",
		ProcessorID:   "paybr",
		PaymentMethod: model.MethodCreditCard,
		Currency:      model.CurrencyBRL,
		Amount:        1000.0,
	}
	got := CalculateNaive(tx, nil)
	want := math.Round(1000.0*0.035*100) / 100
	if got != want {
		t.Errorf("CalculateNaive() with nil processors = %v, want %v", got, want)
	}
}

func TestSupportsCountryAndCurrency(t *testing.T) {
	t.Parallel()

	paybr := testProcessorPayBR()
	globalpay := testProcessorGlobalPay()

	tests := []struct {
		name     string
		proc     model.Processor
		country  model.Country
		currency model.Currency
		want     bool
	}{
		{
			name:     "paybr supports BR and BRL",
			proc:     paybr,
			country:  model.CountryBR,
			currency: model.CurrencyBRL,
			want:     true,
		},
		{
			name:     "paybr does not support MX",
			proc:     paybr,
			country:  model.CountryMX,
			currency: model.CurrencyMXN,
			want:     false,
		},
		{
			name:     "paybr country match but currency mismatch",
			proc:     paybr,
			country:  model.CountryBR,
			currency: model.CurrencyMXN,
			want:     false,
		},
		{
			name:     "paybr currency match but country mismatch",
			proc:     paybr,
			country:  model.CountryMX,
			currency: model.CurrencyBRL,
			want:     false,
		},
		{
			name:     "globalpay supports BR and BRL",
			proc:     globalpay,
			country:  model.CountryBR,
			currency: model.CurrencyBRL,
			want:     true,
		},
		{
			name:     "globalpay supports MX and MXN",
			proc:     globalpay,
			country:  model.CountryMX,
			currency: model.CurrencyMXN,
			want:     true,
		},
		{
			name:     "globalpay supports CO and COP",
			proc:     globalpay,
			country:  model.CountryCO,
			currency: model.CurrencyCOP,
			want:     true,
		},
		{
			name:     "globalpay BR country with COP currency",
			proc:     globalpay,
			country:  model.CountryBR,
			currency: model.CurrencyCOP,
			want:     true,
		},
		{
			name:     "empty processor supports nothing",
			proc:     model.Processor{},
			country:  model.CountryBR,
			currency: model.CurrencyBRL,
			want:     false,
		},
		{
			name: "processor with countries but no currencies",
			proc: model.Processor{
				SupportedCountries: []model.Country{model.CountryBR},
			},
			country:  model.CountryBR,
			currency: model.CurrencyBRL,
			want:     false,
		},
		{
			name: "processor with currencies but no countries",
			proc: model.Processor{
				SupportedCurrencies: []model.Currency{model.CurrencyBRL},
			},
			country:  model.CountryBR,
			currency: model.CurrencyBRL,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SupportsCountryAndCurrency(tt.proc, tt.country, tt.currency)
			if got != tt.want {
				t.Errorf("SupportsCountryAndCurrency(%v, %v, %v) = %v, want %v",
					tt.proc.ID, tt.country, tt.currency, got, tt.want)
			}
		})
	}
}
