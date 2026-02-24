package rules

import (
	"testing"

	"github.com/ivanjtm/YunoChallenge/internal/model"
)

func boolPtr(b bool) *bool { return &b }

func pixBR() model.CompatibilityRule {
	return model.CompatibilityRule{
		OriginalMethod: model.MethodPIX,
		Country:        model.CountryBR,
		AllowedRefunds: []model.AllowedRefund{
			{Method: model.RefundReversal, MaxAgeDays: 0, RequireSettled: boolPtr(false)},
			{Method: model.RefundSameMethod, MaxAgeDays: 90},
			{Method: model.RefundBankTransfer, MaxAgeDays: 0},
		},
	}
}

func oxxoMX() model.CompatibilityRule {
	return model.CompatibilityRule{
		OriginalMethod: model.MethodOXXO,
		Country:        model.CountryMX,
		AllowedRefunds: []model.AllowedRefund{
			{Method: model.RefundBankTransfer, MaxAgeDays: 0},
			{Method: model.RefundAccountCredit, MaxAgeDays: 0},
		},
	}
}

func pseCO() model.CompatibilityRule {
	return model.CompatibilityRule{
		OriginalMethod: model.MethodPSE,
		Country:        model.CountryCO,
		AllowedRefunds: []model.AllowedRefund{
			{Method: model.RefundReversal, MaxAgeDays: 0, RequireSettled: boolPtr(false)},
			{Method: model.RefundSameMethod, MaxAgeDays: 60},
			{Method: model.RefundBankTransfer, MaxAgeDays: 0},
		},
	}
}

func creditCardBR() model.CompatibilityRule {
	return model.CompatibilityRule{
		OriginalMethod: model.MethodCreditCard,
		Country:        model.CountryBR,
		AllowedRefunds: []model.AllowedRefund{
			{Method: model.RefundReversal, MaxAgeDays: 0, RequireSettled: boolPtr(false)},
			{Method: model.RefundSameMethod, MaxAgeDays: 180},
			{Method: model.RefundBankTransfer, MaxAgeDays: 0},
		},
	}
}

func boletoBR() model.CompatibilityRule {
	return model.CompatibilityRule{
		OriginalMethod: model.MethodBoleto,
		Country:        model.CountryBR,
		AllowedRefunds: []model.AllowedRefund{
			{Method: model.RefundBankTransfer, MaxAgeDays: 0},
			{Method: model.RefundAccountCredit, MaxAgeDays: 0},
		},
	}
}

func speiMX() model.CompatibilityRule {
	return model.CompatibilityRule{
		OriginalMethod: model.MethodSPEI,
		Country:        model.CountryMX,
		AllowedRefunds: []model.AllowedRefund{
			{Method: model.RefundReversal, MaxAgeDays: 0, RequireSettled: boolPtr(false)},
			{Method: model.RefundSameMethod, MaxAgeDays: 0},
			{Method: model.RefundBankTransfer, MaxAgeDays: 0},
		},
	}
}

func efectyCO() model.CompatibilityRule {
	return model.CompatibilityRule{
		OriginalMethod: model.MethodEfecty,
		Country:        model.CountryCO,
		AllowedRefunds: []model.AllowedRefund{
			{Method: model.RefundBankTransfer, MaxAgeDays: 0},
			{Method: model.RefundAccountCredit, MaxAgeDays: 0},
		},
	}
}

func allRules() []model.CompatibilityRule {
	return []model.CompatibilityRule{
		pixBR(),
		oxxoMX(),
		pseCO(),
		creditCardBR(),
		boletoBR(),
		speiMX(),
		efectyCO(),
	}
}

func TestNewRuleIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		rules     []model.CompatibilityRule
		wantCount int
	}{
		{
			name:      "builds index from all rules",
			rules:     allRules(),
			wantCount: 7,
		},
		{
			name:      "empty slice produces empty index",
			rules:     []model.CompatibilityRule{},
			wantCount: 0,
		},
		{
			name:      "nil slice produces empty index",
			rules:     nil,
			wantCount: 0,
		},
		{
			name:      "single rule",
			rules:     []model.CompatibilityRule{pixBR()},
			wantCount: 1,
		},
		{
			name: "duplicate key last wins",
			rules: []model.CompatibilityRule{
				{
					OriginalMethod: model.MethodPIX,
					Country:        model.CountryBR,
					AllowedRefunds: []model.AllowedRefund{
						{Method: model.RefundSameMethod, MaxAgeDays: 30},
					},
				},
				{
					OriginalMethod: model.MethodPIX,
					Country:        model.CountryBR,
					AllowedRefunds: []model.AllowedRefund{
						{Method: model.RefundBankTransfer, MaxAgeDays: 0},
						{Method: model.RefundAccountCredit, MaxAgeDays: 0},
					},
				},
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			idx := NewRuleIndex(tt.rules)
			if idx == nil {
				t.Fatal("NewRuleIndex() returned nil")
			}
			if len(idx.index) != tt.wantCount {
				t.Errorf("index size = %d, want %d", len(idx.index), tt.wantCount)
			}
		})
	}
}

func TestNewRuleIndex_DuplicateKeyLastWins(t *testing.T) {
	t.Parallel()

	first := model.CompatibilityRule{
		OriginalMethod: model.MethodPIX,
		Country:        model.CountryBR,
		AllowedRefunds: []model.AllowedRefund{
			{Method: model.RefundSameMethod, MaxAgeDays: 30},
		},
	}
	second := model.CompatibilityRule{
		OriginalMethod: model.MethodPIX,
		Country:        model.CountryBR,
		AllowedRefunds: []model.AllowedRefund{
			{Method: model.RefundBankTransfer, MaxAgeDays: 0},
			{Method: model.RefundAccountCredit, MaxAgeDays: 0},
		},
	}

	idx := NewRuleIndex([]model.CompatibilityRule{first, second})
	got := idx.Lookup(model.MethodPIX, model.CountryBR)
	if got == nil {
		t.Fatal("Lookup() returned nil for duplicate key")
	}
	if len(got.AllowedRefunds) != 2 {
		t.Errorf("expected last-wins with 2 refunds, got %d", len(got.AllowedRefunds))
	}
	if got.AllowedRefunds[0].Method != model.RefundBankTransfer {
		t.Errorf("first refund method = %s, want %s", got.AllowedRefunds[0].Method, model.RefundBankTransfer)
	}
	if got.AllowedRefunds[1].Method != model.RefundAccountCredit {
		t.Errorf("second refund method = %s, want %s", got.AllowedRefunds[1].Method, model.RefundAccountCredit)
	}
}

func TestLookup(t *testing.T) {
	t.Parallel()

	idx := NewRuleIndex(allRules())

	tests := []struct {
		name       string
		method     model.PaymentMethod
		country    model.Country
		wantNil    bool
		wantMethod model.PaymentMethod
	}{
		{
			name:       "PIX in BR",
			method:     model.MethodPIX,
			country:    model.CountryBR,
			wantNil:    false,
			wantMethod: model.MethodPIX,
		},
		{
			name:       "OXXO in MX",
			method:     model.MethodOXXO,
			country:    model.CountryMX,
			wantNil:    false,
			wantMethod: model.MethodOXXO,
		},
		{
			name:       "PSE in CO",
			method:     model.MethodPSE,
			country:    model.CountryCO,
			wantNil:    false,
			wantMethod: model.MethodPSE,
		},
		{
			name:       "CREDIT_CARD in BR",
			method:     model.MethodCreditCard,
			country:    model.CountryBR,
			wantNil:    false,
			wantMethod: model.MethodCreditCard,
		},
		{
			name:       "BOLETO in BR",
			method:     model.MethodBoleto,
			country:    model.CountryBR,
			wantNil:    false,
			wantMethod: model.MethodBoleto,
		},
		{
			name:       "SPEI in MX",
			method:     model.MethodSPEI,
			country:    model.CountryMX,
			wantNil:    false,
			wantMethod: model.MethodSPEI,
		},
		{
			name:       "EFECTY in CO",
			method:     model.MethodEfecty,
			country:    model.CountryCO,
			wantNil:    false,
			wantMethod: model.MethodEfecty,
		},
		{
			name:    "PIX in MX does not exist",
			method:  model.MethodPIX,
			country: model.CountryMX,
			wantNil: true,
		},
		{
			name:    "OXXO in BR does not exist",
			method:  model.MethodOXXO,
			country: model.CountryBR,
			wantNil: true,
		},
		{
			name:    "PSE in MX does not exist",
			method:  model.MethodPSE,
			country: model.CountryMX,
			wantNil: true,
		},
		{
			name:    "EFECTY in BR does not exist",
			method:  model.MethodEfecty,
			country: model.CountryBR,
			wantNil: true,
		},
		{
			name:    "unknown method returns nil",
			method:  "CRYPTO",
			country: model.CountryBR,
			wantNil: true,
		},
		{
			name:    "unknown country returns nil",
			method:  model.MethodPIX,
			country: "US",
			wantNil: true,
		},
		{
			name:    "empty method returns nil",
			method:  "",
			country: model.CountryBR,
			wantNil: true,
		},
		{
			name:    "empty country returns nil",
			method:  model.MethodPIX,
			country: "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := idx.Lookup(tt.method, tt.country)
			if tt.wantNil {
				if got != nil {
					t.Errorf("Lookup(%s, %s) = %+v, want nil", tt.method, tt.country, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("Lookup(%s, %s) = nil, want non-nil", tt.method, tt.country)
			}
			if got.OriginalMethod != tt.wantMethod {
				t.Errorf("OriginalMethod = %s, want %s", got.OriginalMethod, tt.wantMethod)
			}
			if got.Country != tt.country {
				t.Errorf("Country = %s, want %s", got.Country, tt.country)
			}
		})
	}
}

func TestLookup_ReturnsCorrectRefundCounts(t *testing.T) {
	t.Parallel()

	idx := NewRuleIndex(allRules())

	tests := []struct {
		name      string
		method    model.PaymentMethod
		country   model.Country
		wantCount int
	}{
		{
			name:      "PIX BR has 3 refund methods",
			method:    model.MethodPIX,
			country:   model.CountryBR,
			wantCount: 3,
		},
		{
			name:      "CREDIT_CARD BR has 3 refund methods",
			method:    model.MethodCreditCard,
			country:   model.CountryBR,
			wantCount: 3,
		},
		{
			name:      "BOLETO BR has 2 refund methods",
			method:    model.MethodBoleto,
			country:   model.CountryBR,
			wantCount: 2,
		},
		{
			name:      "OXXO MX has 2 refund methods",
			method:    model.MethodOXXO,
			country:   model.CountryMX,
			wantCount: 2,
		},
		{
			name:      "SPEI MX has 3 refund methods",
			method:    model.MethodSPEI,
			country:   model.CountryMX,
			wantCount: 3,
		},
		{
			name:      "PSE CO has 3 refund methods",
			method:    model.MethodPSE,
			country:   model.CountryCO,
			wantCount: 3,
		},
		{
			name:      "EFECTY CO has 2 refund methods",
			method:    model.MethodEfecty,
			country:   model.CountryCO,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := idx.Lookup(tt.method, tt.country)
			if got == nil {
				t.Fatalf("Lookup(%s, %s) = nil", tt.method, tt.country)
			}
			if len(got.AllowedRefunds) != tt.wantCount {
				t.Errorf("len(AllowedRefunds) = %d, want %d", len(got.AllowedRefunds), tt.wantCount)
			}
		})
	}
}

func TestLookup_EmptyIndex(t *testing.T) {
	t.Parallel()

	idx := NewRuleIndex(nil)
	got := idx.Lookup(model.MethodPIX, model.CountryBR)
	if got != nil {
		t.Errorf("Lookup on empty index = %+v, want nil", got)
	}
}

func TestAllowedRefundMethods(t *testing.T) {
	t.Parallel()

	idx := NewRuleIndex(allRules())

	tests := []struct {
		name        string
		method      model.PaymentMethod
		country     model.Country
		wantNil     bool
		wantMethods []model.RefundMethod
	}{
		{
			name:    "PIX BR returns reversal, same_method, bank_transfer",
			method:  model.MethodPIX,
			country: model.CountryBR,
			wantMethods: []model.RefundMethod{
				model.RefundReversal,
				model.RefundSameMethod,
				model.RefundBankTransfer,
			},
		},
		{
			name:    "CREDIT_CARD BR returns reversal, same_method, bank_transfer",
			method:  model.MethodCreditCard,
			country: model.CountryBR,
			wantMethods: []model.RefundMethod{
				model.RefundReversal,
				model.RefundSameMethod,
				model.RefundBankTransfer,
			},
		},
		{
			name:    "BOLETO BR returns bank_transfer, account_credit",
			method:  model.MethodBoleto,
			country: model.CountryBR,
			wantMethods: []model.RefundMethod{
				model.RefundBankTransfer,
				model.RefundAccountCredit,
			},
		},
		{
			name:    "OXXO MX returns bank_transfer, account_credit",
			method:  model.MethodOXXO,
			country: model.CountryMX,
			wantMethods: []model.RefundMethod{
				model.RefundBankTransfer,
				model.RefundAccountCredit,
			},
		},
		{
			name:    "SPEI MX returns reversal, same_method, bank_transfer",
			method:  model.MethodSPEI,
			country: model.CountryMX,
			wantMethods: []model.RefundMethod{
				model.RefundReversal,
				model.RefundSameMethod,
				model.RefundBankTransfer,
			},
		},
		{
			name:    "PSE CO returns reversal, same_method, bank_transfer",
			method:  model.MethodPSE,
			country: model.CountryCO,
			wantMethods: []model.RefundMethod{
				model.RefundReversal,
				model.RefundSameMethod,
				model.RefundBankTransfer,
			},
		},
		{
			name:    "EFECTY CO returns bank_transfer, account_credit",
			method:  model.MethodEfecty,
			country: model.CountryCO,
			wantMethods: []model.RefundMethod{
				model.RefundBankTransfer,
				model.RefundAccountCredit,
			},
		},
		{
			name:    "unknown combo returns nil",
			method:  model.MethodPIX,
			country: model.CountryMX,
			wantNil: true,
		},
		{
			name:    "empty method returns nil",
			method:  "",
			country: model.CountryBR,
			wantNil: true,
		},
		{
			name:    "empty country returns nil",
			method:  model.MethodPIX,
			country: "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := idx.AllowedRefundMethods(tt.method, tt.country)
			if tt.wantNil {
				if got != nil {
					t.Errorf("AllowedRefundMethods(%s, %s) = %+v, want nil", tt.method, tt.country, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("AllowedRefundMethods(%s, %s) = nil, want non-nil", tt.method, tt.country)
			}
			if len(got) != len(tt.wantMethods) {
				t.Fatalf("len(AllowedRefundMethods) = %d, want %d", len(got), len(tt.wantMethods))
			}
			for i, want := range tt.wantMethods {
				if got[i].Method != want {
					t.Errorf("refund[%d].Method = %s, want %s", i, got[i].Method, want)
				}
			}
		})
	}
}

func TestAllowedRefundMethods_EmptyIndex(t *testing.T) {
	t.Parallel()

	idx := NewRuleIndex(nil)
	got := idx.AllowedRefundMethods(model.MethodPIX, model.CountryBR)
	if got != nil {
		t.Errorf("AllowedRefundMethods on empty index = %+v, want nil", got)
	}
}

func TestCashMethodsCannotSelfRefund(t *testing.T) {
	t.Parallel()

	idx := NewRuleIndex(allRules())

	tests := []struct {
		name    string
		method  model.PaymentMethod
		country model.Country
	}{
		{
			name:    "OXXO MX cannot self-refund",
			method:  model.MethodOXXO,
			country: model.CountryMX,
		},
		{
			name:    "BOLETO BR cannot self-refund",
			method:  model.MethodBoleto,
			country: model.CountryBR,
		},
		{
			name:    "EFECTY CO cannot self-refund",
			method:  model.MethodEfecty,
			country: model.CountryCO,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			refunds := idx.AllowedRefundMethods(tt.method, tt.country)
			if refunds == nil {
				t.Fatalf("AllowedRefundMethods(%s, %s) = nil", tt.method, tt.country)
			}
			for _, r := range refunds {
				if r.Method == model.RefundSameMethod {
					t.Errorf("cash method %s in %s should not allow SAME_METHOD refund", tt.method, tt.country)
				}
			}
		})
	}
}

func TestAllowedRefundMethods_MaxAgeDays(t *testing.T) {
	t.Parallel()

	idx := NewRuleIndex(allRules())

	tests := []struct {
		name          string
		method        model.PaymentMethod
		country       model.Country
		refundMethod  model.RefundMethod
		wantMaxAge    int
	}{
		{
			name:         "PIX BR SAME_METHOD has 90 day window",
			method:       model.MethodPIX,
			country:      model.CountryBR,
			refundMethod: model.RefundSameMethod,
			wantMaxAge:   90,
		},
		{
			name:         "CREDIT_CARD BR SAME_METHOD has 180 day window",
			method:       model.MethodCreditCard,
			country:      model.CountryBR,
			refundMethod: model.RefundSameMethod,
			wantMaxAge:   180,
		},
		{
			name:         "PSE CO SAME_METHOD has 60 day window",
			method:       model.MethodPSE,
			country:      model.CountryCO,
			refundMethod: model.RefundSameMethod,
			wantMaxAge:   60,
		},
		{
			name:         "SPEI MX SAME_METHOD has no time limit",
			method:       model.MethodSPEI,
			country:      model.CountryMX,
			refundMethod: model.RefundSameMethod,
			wantMaxAge:   0,
		},
		{
			name:         "PIX BR BANK_TRANSFER has no time limit",
			method:       model.MethodPIX,
			country:      model.CountryBR,
			refundMethod: model.RefundBankTransfer,
			wantMaxAge:   0,
		},
		{
			name:         "OXXO MX ACCOUNT_CREDIT has no time limit",
			method:       model.MethodOXXO,
			country:      model.CountryMX,
			refundMethod: model.RefundAccountCredit,
			wantMaxAge:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			refunds := idx.AllowedRefundMethods(tt.method, tt.country)
			if refunds == nil {
				t.Fatalf("AllowedRefundMethods(%s, %s) = nil", tt.method, tt.country)
			}
			var found bool
			for _, r := range refunds {
				if r.Method == tt.refundMethod {
					found = true
					if r.MaxAgeDays != tt.wantMaxAge {
						t.Errorf("MaxAgeDays for %s = %d, want %d", tt.refundMethod, r.MaxAgeDays, tt.wantMaxAge)
					}
					break
				}
			}
			if !found {
				t.Errorf("refund method %s not found in AllowedRefunds for %s+%s", tt.refundMethod, tt.method, tt.country)
			}
		})
	}
}

func TestLookup_ReturnedRuleIsIndependent(t *testing.T) {
	t.Parallel()

	idx := NewRuleIndex(allRules())

	got1 := idx.Lookup(model.MethodPIX, model.CountryBR)
	got2 := idx.Lookup(model.MethodPIX, model.CountryBR)

	if got1 == got2 {
		t.Error("Lookup should return distinct pointers on each call")
	}
	if got1.OriginalMethod != got2.OriginalMethod {
		t.Error("both lookups should return equivalent data")
	}
}
