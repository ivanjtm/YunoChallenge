package rules

import (
	"testing"
	"time"

	"github.com/ivanjtm/YunoChallenge/internal/model"
)

func TestFindEligiblePaths(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		tx          model.Transaction
		rules       []model.CompatibilityRule
		now         time.Time
		wantMethods []model.RefundMethod
	}{
		{
			name: "fresh unsettled PIX gets reversal and same method and bank transfer",
			tx: model.Transaction{
				ID:            "tx-pix-fresh",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-2 * time.Hour),
				Settled:       false,
			},
			rules:       allRules(),
			now:         now,
			wantMethods: []model.RefundMethod{model.RefundReversal, model.RefundSameMethod, model.RefundBankTransfer},
		},
		{
			name: "old PIX past 90 days gets only bank transfer",
			tx: model.Transaction{
				ID:            "tx-pix-old",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-91 * 24 * time.Hour),
				Settled:       true,
			},
			rules:       allRules(),
			now:         now,
			wantMethods: []model.RefundMethod{model.RefundBankTransfer},
		},
		{
			name: "OXXO has no same method and gets bank transfer and account credit",
			tx: model.Transaction{
				ID:            "tx-oxxo",
				Country:       model.CountryMX,
				PaymentMethod: model.MethodOXXO,
				Timestamp:     now.Add(-5 * 24 * time.Hour),
				Settled:       true,
			},
			rules:       allRules(),
			now:         now,
			wantMethods: []model.RefundMethod{model.RefundBankTransfer, model.RefundAccountCredit},
		},
		{
			name: "no rules found returns only account credit",
			tx: model.Transaction{
				ID:            "tx-unknown",
				Country:       "US",
				PaymentMethod: "CRYPTO",
				Timestamp:     now.Add(-1 * time.Hour),
				Settled:       false,
			},
			rules:       allRules(),
			now:         now,
			wantMethods: []model.RefundMethod{model.RefundAccountCredit},
		},
		{
			name: "empty rule index returns only account credit",
			tx: model.Transaction{
				ID:            "tx-empty-rules",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-1 * time.Hour),
				Settled:       false,
			},
			rules:       nil,
			now:         now,
			wantMethods: []model.RefundMethod{model.RefundAccountCredit},
		},
		{
			name: "settled transaction does not get reversal",
			tx: model.Transaction{
				ID:            "tx-settled",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-2 * time.Hour),
				Settled:       true,
			},
			rules:       allRules(),
			now:         now,
			wantMethods: []model.RefundMethod{model.RefundSameMethod, model.RefundBankTransfer},
		},
		{
			name: "unsettled PIX older than 24h does not get reversal",
			tx: model.Transaction{
				ID:            "tx-pix-25h",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-25 * time.Hour),
				Settled:       false,
			},
			rules:       allRules(),
			now:         now,
			wantMethods: []model.RefundMethod{model.RefundSameMethod, model.RefundBankTransfer},
		},
		{
			name: "PSE within 60 day window gets reversal and same method and bank transfer",
			tx: model.Transaction{
				ID:            "tx-pse-fresh",
				Country:       model.CountryCO,
				PaymentMethod: model.MethodPSE,
				Timestamp:     now.Add(-5 * time.Hour),
				Settled:       false,
			},
			rules:       allRules(),
			now:         now,
			wantMethods: []model.RefundMethod{model.RefundReversal, model.RefundSameMethod, model.RefundBankTransfer},
		},
		{
			name: "PSE past 60 day window loses same method",
			tx: model.Transaction{
				ID:            "tx-pse-old",
				Country:       model.CountryCO,
				PaymentMethod: model.MethodPSE,
				Timestamp:     now.Add(-61 * 24 * time.Hour),
				Settled:       true,
			},
			rules:       allRules(),
			now:         now,
			wantMethods: []model.RefundMethod{model.RefundBankTransfer},
		},
		{
			name: "credit card within 180 day window gets reversal and same method and bank transfer",
			tx: model.Transaction{
				ID:            "tx-card-fresh",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodCreditCard,
				Timestamp:     now.Add(-10 * time.Hour),
				Settled:       false,
			},
			rules:       allRules(),
			now:         now,
			wantMethods: []model.RefundMethod{model.RefundReversal, model.RefundSameMethod, model.RefundBankTransfer},
		},
		{
			name: "credit card past 180 day window gets only bank transfer",
			tx: model.Transaction{
				ID:            "tx-card-old",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodCreditCard,
				Timestamp:     now.Add(-181 * 24 * time.Hour),
				Settled:       true,
			},
			rules:       allRules(),
			now:         now,
			wantMethods: []model.RefundMethod{model.RefundBankTransfer},
		},
		{
			name: "all paths expired falls back to account credit",
			tx: model.Transaction{
				ID:            "tx-all-expired",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-91 * 24 * time.Hour),
				Settled:       true,
			},
			rules: []model.CompatibilityRule{
				{
					OriginalMethod: model.MethodPIX,
					Country:        model.CountryBR,
					AllowedRefunds: []model.AllowedRefund{
						{Method: model.RefundReversal, MaxAgeDays: 0, RequireSettled: boolPtr(false)},
						{Method: model.RefundSameMethod, MaxAgeDays: 90},
					},
				},
			},
			now:         now,
			wantMethods: []model.RefundMethod{model.RefundAccountCredit},
		},
		{
			name: "require settled true skips unsettled transaction",
			tx: model.Transaction{
				ID:            "tx-require-settled",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-25 * time.Hour),
				Settled:       false,
			},
			rules: []model.CompatibilityRule{
				{
					OriginalMethod: model.MethodPIX,
					Country:        model.CountryBR,
					AllowedRefunds: []model.AllowedRefund{
						{Method: model.RefundSameMethod, MaxAgeDays: 90, RequireSettled: boolPtr(true)},
					},
				},
			},
			now:         now,
			wantMethods: []model.RefundMethod{model.RefundAccountCredit},
		},
		{
			name: "require settled false skips settled transaction",
			tx: model.Transaction{
				ID:            "tx-require-unsettled",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-25 * time.Hour),
				Settled:       true,
			},
			rules: []model.CompatibilityRule{
				{
					OriginalMethod: model.MethodPIX,
					Country:        model.CountryBR,
					AllowedRefunds: []model.AllowedRefund{
						{Method: model.RefundSameMethod, MaxAgeDays: 90, RequireSettled: boolPtr(false)},
					},
				},
			},
			now:         now,
			wantMethods: []model.RefundMethod{model.RefundAccountCredit},
		},
		{
			name: "require settled nil allows both settled and unsettled",
			tx: model.Transaction{
				ID:            "tx-no-settle-req",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-25 * time.Hour),
				Settled:       true,
			},
			rules: []model.CompatibilityRule{
				{
					OriginalMethod: model.MethodPIX,
					Country:        model.CountryBR,
					AllowedRefunds: []model.AllowedRefund{
						{Method: model.RefundSameMethod, MaxAgeDays: 90},
					},
				},
			},
			now:         now,
			wantMethods: []model.RefundMethod{model.RefundSameMethod},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			idx := NewRuleIndex(tt.rules)
			paths := FindEligiblePaths(tt.tx, idx, tt.now)
			if len(paths) != len(tt.wantMethods) {
				methods := make([]model.RefundMethod, len(paths))
				for i, p := range paths {
					methods[i] = p.Method
				}
				t.Fatalf("FindEligiblePaths() returned %d paths %v, want %d paths %v", len(paths), methods, len(tt.wantMethods), tt.wantMethods)
			}
			for i, want := range tt.wantMethods {
				if paths[i].Method != want {
					t.Errorf("paths[%d].Method = %s, want %s", i, paths[i].Method, want)
				}
				if paths[i].Reason == "" {
					t.Errorf("paths[%d].Reason is empty", i)
				}
			}
		})
	}
}

func TestFindEligiblePaths_ReasonsAreNonEmpty(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	idx := NewRuleIndex(allRules())

	tx := model.Transaction{
		ID:            "tx-reasons",
		Country:       model.CountryBR,
		PaymentMethod: model.MethodPIX,
		Timestamp:     now.Add(-2 * time.Hour),
		Settled:       false,
	}

	paths := FindEligiblePaths(tx, idx, now)
	if len(paths) == 0 {
		t.Fatal("FindEligiblePaths() returned no paths")
	}
	for i, p := range paths {
		if p.Reason == "" {
			t.Errorf("paths[%d].Reason is empty for method %s", i, p.Method)
		}
	}
}

func TestTimeSensitiveWindows(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		tx            model.Transaction
		rules         []model.CompatibilityRule
		now           time.Time
		thresholdDays int
		wantCount     int
		wantTypes     []string
	}{
		{
			name: "reversal at 20h is flagged",
			tx: model.Transaction{
				ID:            "tx-rev-20h",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-20 * time.Hour),
				Settled:       false,
			},
			rules:         allRules(),
			now:           now,
			thresholdDays: 7,
			wantCount:     1,
			wantTypes:     []string{"REVERSAL_24H"},
		},
		{
			name: "reversal at 18h is flagged",
			tx: model.Transaction{
				ID:            "tx-rev-18h",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-18 * time.Hour),
				Settled:       false,
			},
			rules:         allRules(),
			now:           now,
			thresholdDays: 7,
			wantCount:     1,
			wantTypes:     []string{"REVERSAL_24H"},
		},
		{
			name: "reversal at 10h is not flagged",
			tx: model.Transaction{
				ID:            "tx-rev-10h",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-10 * time.Hour),
				Settled:       false,
			},
			rules:         allRules(),
			now:           now,
			thresholdDays: 7,
			wantCount:     0,
			wantTypes:     nil,
		},
		{
			name: "reversal at 24h is not flagged",
			tx: model.Transaction{
				ID:            "tx-rev-24h",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-24 * time.Hour),
				Settled:       false,
			},
			rules:         allRules(),
			now:           now,
			thresholdDays: 7,
			wantCount:     0,
			wantTypes:     nil,
		},
		{
			name: "settled transaction reversal not flagged even at 20h",
			tx: model.Transaction{
				ID:            "tx-rev-settled",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-20 * time.Hour),
				Settled:       true,
			},
			rules:         allRules(),
			now:           now,
			thresholdDays: 7,
			wantCount:     0,
			wantTypes:     nil,
		},
		{
			name: "PIX at 85 days is flagged with threshold 7",
			tx: model.Transaction{
				ID:            "tx-pix-85d",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-85 * 24 * time.Hour),
				Settled:       true,
			},
			rules:         allRules(),
			now:           now,
			thresholdDays: 7,
			wantCount:     1,
			wantTypes:     []string{"PIX_SAME_METHOD_90D"},
		},
		{
			name: "PSE at 55 days is flagged with threshold 7",
			tx: model.Transaction{
				ID:            "tx-pse-55d",
				Country:       model.CountryCO,
				PaymentMethod: model.MethodPSE,
				Timestamp:     now.Add(-55 * 24 * time.Hour),
				Settled:       true,
			},
			rules:         allRules(),
			now:           now,
			thresholdDays: 7,
			wantCount:     1,
			wantTypes:     []string{"PSE_SAME_METHOD_60D"},
		},
		{
			name: "credit card at 175 days is flagged with threshold 7",
			tx: model.Transaction{
				ID:            "tx-card-175d",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodCreditCard,
				Timestamp:     now.Add(-175 * 24 * time.Hour),
				Settled:       true,
			},
			rules:         allRules(),
			now:           now,
			thresholdDays: 7,
			wantCount:     1,
			wantTypes:     []string{"CREDIT_CARD_SAME_METHOD_180D"},
		},
		{
			name: "nothing expiring soon returns empty",
			tx: model.Transaction{
				ID:            "tx-fresh",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-5 * 24 * time.Hour),
				Settled:       true,
			},
			rules:         allRules(),
			now:           now,
			thresholdDays: 7,
			wantCount:     0,
			wantTypes:     nil,
		},
		{
			name: "bank transfer with no time limit is never flagged",
			tx: model.Transaction{
				ID:            "tx-bt-old",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-1000 * 24 * time.Hour),
				Settled:       true,
			},
			rules:         allRules(),
			now:           now,
			thresholdDays: 7,
			wantCount:     0,
			wantTypes:     nil,
		},
		{
			name: "OXXO with no time-limited methods is never flagged",
			tx: model.Transaction{
				ID:            "tx-oxxo-flag",
				Country:       model.CountryMX,
				PaymentMethod: model.MethodOXXO,
				Timestamp:     now.Add(-30 * 24 * time.Hour),
				Settled:       true,
			},
			rules:         allRules(),
			now:           now,
			thresholdDays: 7,
			wantCount:     0,
			wantTypes:     nil,
		},
		{
			name: "reversal at 20h and PIX same method at 85 days both flagged",
			tx: model.Transaction{
				ID:            "tx-multi-flag",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-20 * time.Hour),
				Settled:       false,
			},
			rules: []model.CompatibilityRule{
				{
					OriginalMethod: model.MethodPIX,
					Country:        model.CountryBR,
					AllowedRefunds: []model.AllowedRefund{
						{Method: model.RefundReversal, MaxAgeDays: 0, RequireSettled: boolPtr(false)},
						{Method: model.RefundSameMethod, MaxAgeDays: 1},
					},
				},
			},
			now:           now,
			thresholdDays: 7,
			wantCount:     2,
			wantTypes:     []string{"REVERSAL_24H", "PIX_SAME_METHOD_1D"},
		},
		{
			name: "expired method is not flagged",
			tx: model.Transaction{
				ID:            "tx-expired-flag",
				Country:       model.CountryBR,
				PaymentMethod: model.MethodPIX,
				Timestamp:     now.Add(-100 * 24 * time.Hour),
				Settled:       true,
			},
			rules:         allRules(),
			now:           now,
			thresholdDays: 7,
			wantCount:     0,
			wantTypes:     nil,
		},
		{
			name: "no rules returns empty flags",
			tx: model.Transaction{
				ID:            "tx-no-rules-flag",
				Country:       "US",
				PaymentMethod: "CRYPTO",
				Timestamp:     now.Add(-1 * time.Hour),
				Settled:       false,
			},
			rules:         allRules(),
			now:           now,
			thresholdDays: 7,
			wantCount:     0,
			wantTypes:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			idx := NewRuleIndex(tt.rules)
			flags := TimeSensitiveWindows(tt.tx, idx, tt.now, tt.thresholdDays)
			if len(flags) != tt.wantCount {
				types := make([]string, len(flags))
				for i, f := range flags {
					types[i] = f.WindowType
				}
				t.Fatalf("TimeSensitiveWindows() returned %d flags %v, want %d", len(flags), types, tt.wantCount)
			}
			for i, wantType := range tt.wantTypes {
				if flags[i].WindowType != wantType {
					t.Errorf("flags[%d].WindowType = %s, want %s", i, flags[i].WindowType, wantType)
				}
				if flags[i].TransactionID != tt.tx.ID {
					t.Errorf("flags[%d].TransactionID = %s, want %s", i, flags[i].TransactionID, tt.tx.ID)
				}
				if flags[i].Message == "" {
					t.Errorf("flags[%d].Message is empty", i)
				}
				if flags[i].ExpiresAt.IsZero() {
					t.Errorf("flags[%d].ExpiresAt is zero", i)
				}
			}
		})
	}
}

func TestTimeSensitiveWindows_ReversalExpiresAt(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	txTime := now.Add(-20 * time.Hour)
	tx := model.Transaction{
		ID:            "tx-expiry-check",
		Country:       model.CountryBR,
		PaymentMethod: model.MethodPIX,
		Timestamp:     txTime,
		Settled:       false,
	}
	idx := NewRuleIndex(allRules())
	flags := TimeSensitiveWindows(tx, idx, now, 7)

	var found bool
	for _, f := range flags {
		if f.WindowType == "REVERSAL_24H" {
			found = true
			wantExpiry := txTime.Add(24 * time.Hour)
			if !f.ExpiresAt.Equal(wantExpiry) {
				t.Errorf("ExpiresAt = %v, want %v", f.ExpiresAt, wantExpiry)
			}
			if f.DaysRemaining != 0 {
				t.Errorf("DaysRemaining = %d, want 0", f.DaysRemaining)
			}
		}
	}
	if !found {
		t.Error("REVERSAL_24H flag not found")
	}
}

func TestTimeSensitiveWindows_MethodExpiresAt(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	txTime := now.Add(-85 * 24 * time.Hour)
	tx := model.Transaction{
		ID:            "tx-method-expiry",
		Country:       model.CountryBR,
		PaymentMethod: model.MethodPIX,
		Timestamp:     txTime,
		Settled:       true,
	}
	idx := NewRuleIndex(allRules())
	flags := TimeSensitiveWindows(tx, idx, now, 7)

	if len(flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(flags))
	}
	wantExpiry := txTime.AddDate(0, 0, 90)
	if !flags[0].ExpiresAt.Equal(wantExpiry) {
		t.Errorf("ExpiresAt = %v, want %v", flags[0].ExpiresAt, wantExpiry)
	}
	if flags[0].DaysRemaining != 5 {
		t.Errorf("DaysRemaining = %d, want 5", flags[0].DaysRemaining)
	}
}

func TestWindowTypeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method model.PaymentMethod
		ar     model.AllowedRefund
		want   string
	}{
		{
			name:   "PIX same method 90 days",
			method: model.MethodPIX,
			ar:     model.AllowedRefund{Method: model.RefundSameMethod, MaxAgeDays: 90},
			want:   "PIX_SAME_METHOD_90D",
		},
		{
			name:   "PSE same method 60 days",
			method: model.MethodPSE,
			ar:     model.AllowedRefund{Method: model.RefundSameMethod, MaxAgeDays: 60},
			want:   "PSE_SAME_METHOD_60D",
		},
		{
			name:   "credit card same method 180 days",
			method: model.MethodCreditCard,
			ar:     model.AllowedRefund{Method: model.RefundSameMethod, MaxAgeDays: 180},
			want:   "CREDIT_CARD_SAME_METHOD_180D",
		},
		{
			name:   "bank transfer zero days",
			method: model.MethodPIX,
			ar:     model.AllowedRefund{Method: model.RefundBankTransfer, MaxAgeDays: 0},
			want:   "PIX_BANK_TRANSFER_0D",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := windowTypeName(tt.method, tt.ar)
			if got != tt.want {
				t.Errorf("windowTypeName() = %s, want %s", got, tt.want)
			}
		})
	}
}
