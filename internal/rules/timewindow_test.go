package rules

import (
	"testing"
	"time"

	"github.com/ivanjtm/YunoChallenge/internal/model"
)

func TestIsReversalEligible(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		tx           model.Transaction
		now          time.Time
		wantEligible bool
	}{
		{
			name: "unsettled and fresh is eligible",
			tx: model.Transaction{
				ID:        "tx-1",
				Settled:   false,
				Timestamp: now.Add(-1 * time.Hour),
			},
			now:          now,
			wantEligible: true,
		},
		{
			name: "settled transaction is not eligible",
			tx: model.Transaction{
				ID:        "tx-2",
				Settled:   true,
				Timestamp: now.Add(-1 * time.Hour),
			},
			now:          now,
			wantEligible: false,
		},
		{
			name: "unsettled but older than 24h is not eligible",
			tx: model.Transaction{
				ID:        "tx-3",
				Settled:   false,
				Timestamp: now.Add(-25 * time.Hour),
			},
			now:          now,
			wantEligible: false,
		},
		{
			name: "settled and older than 24h is not eligible",
			tx: model.Transaction{
				ID:        "tx-4",
				Settled:   true,
				Timestamp: now.Add(-48 * time.Hour),
			},
			now:          now,
			wantEligible: false,
		},
		{
			name: "exactly at 24h boundary is not eligible",
			tx: model.Transaction{
				ID:        "tx-5",
				Settled:   false,
				Timestamp: now.Add(-24 * time.Hour),
			},
			now:          now,
			wantEligible: false,
		},
		{
			name: "23h59m old unsettled is eligible",
			tx: model.Transaction{
				ID:        "tx-6",
				Settled:   false,
				Timestamp: now.Add(-23*time.Hour - 59*time.Minute),
			},
			now:          now,
			wantEligible: true,
		},
		{
			name: "23h59m59s old unsettled is eligible",
			tx: model.Transaction{
				ID:        "tx-7",
				Settled:   false,
				Timestamp: now.Add(-23*time.Hour - 59*time.Minute - 59*time.Second),
			},
			now:          now,
			wantEligible: true,
		},
		{
			name: "unsettled created at exactly now is eligible",
			tx: model.Transaction{
				ID:        "tx-8",
				Settled:   false,
				Timestamp: now,
			},
			now:          now,
			wantEligible: true,
		},
		{
			name: "settled even if created just now is not eligible",
			tx: model.Transaction{
				ID:        "tx-9",
				Settled:   true,
				Timestamp: now,
			},
			now:          now,
			wantEligible: false,
		},
		{
			name: "24h and 1 second old is not eligible",
			tx: model.Transaction{
				ID:        "tx-10",
				Settled:   false,
				Timestamp: now.Add(-24*time.Hour - 1*time.Second),
			},
			now:          now,
			wantEligible: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			eligible, reason := IsReversalEligible(tt.tx, tt.now)
			if eligible != tt.wantEligible {
				t.Errorf("IsReversalEligible() eligible = %v, want %v; reason: %s", eligible, tt.wantEligible, reason)
			}
			if reason == "" {
				t.Error("IsReversalEligible() returned empty reason")
			}
		})
	}
}

func TestIsWithinTimeWindow(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		tx           model.Transaction
		allowed      model.AllowedRefund
		now          time.Time
		wantEligible bool
	}{
		{
			name: "within window",
			tx: model.Transaction{
				ID:        "tx-1",
				Timestamp: now.Add(-5 * 24 * time.Hour),
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundSameMethod,
				MaxAgeDays: 30,
			},
			now:          now,
			wantEligible: true,
		},
		{
			name: "past window",
			tx: model.Transaction{
				ID:        "tx-2",
				Timestamp: now.Add(-31 * 24 * time.Hour),
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundSameMethod,
				MaxAgeDays: 30,
			},
			now:          now,
			wantEligible: false,
		},
		{
			name: "no time limit always eligible",
			tx: model.Transaction{
				ID:        "tx-3",
				Timestamp: now.Add(-365 * 24 * time.Hour),
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundBankTransfer,
				MaxAgeDays: 0,
			},
			now:          now,
			wantEligible: true,
		},
		{
			name: "exactly at boundary day is eligible",
			tx: model.Transaction{
				ID:        "tx-4",
				Timestamp: now.Add(-30 * 24 * time.Hour),
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundSameMethod,
				MaxAgeDays: 30,
			},
			now:          now,
			wantEligible: true,
		},
		{
			name: "one day past boundary is not eligible",
			tx: model.Transaction{
				ID:        "tx-5",
				Timestamp: now.Add(-31 * 24 * time.Hour),
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundSameMethod,
				MaxAgeDays: 30,
			},
			now:          now,
			wantEligible: false,
		},
		{
			name: "transaction created today is eligible",
			tx: model.Transaction{
				ID:        "tx-6",
				Timestamp: now,
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundSameMethod,
				MaxAgeDays: 7,
			},
			now:          now,
			wantEligible: true,
		},
		{
			name: "1 day window with 23h old transaction is eligible",
			tx: model.Transaction{
				ID:        "tx-7",
				Timestamp: now.Add(-23 * time.Hour),
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundSameMethod,
				MaxAgeDays: 1,
			},
			now:          now,
			wantEligible: true,
		},
		{
			name: "1 day window with 25h old transaction is not eligible",
			tx: model.Transaction{
				ID:        "tx-8",
				Timestamp: now.Add(-25 * time.Hour),
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundSameMethod,
				MaxAgeDays: 1,
			},
			now:          now,
			wantEligible: true,
		},
		{
			name: "1 day window with 49h old transaction is not eligible",
			tx: model.Transaction{
				ID:        "tx-9",
				Timestamp: now.Add(-49 * time.Hour),
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundSameMethod,
				MaxAgeDays: 1,
			},
			now:          now,
			wantEligible: false,
		},
		{
			name: "no time limit with very old transaction",
			tx: model.Transaction{
				ID:        "tx-10",
				Timestamp: now.Add(-1000 * 24 * time.Hour),
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundAccountCredit,
				MaxAgeDays: 0,
			},
			now:          now,
			wantEligible: true,
		},
		{
			name: "reversal method within short window",
			tx: model.Transaction{
				ID:        "tx-11",
				Timestamp: now.Add(-12 * time.Hour),
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundReversal,
				MaxAgeDays: 1,
			},
			now:          now,
			wantEligible: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			eligible, reason := IsWithinTimeWindow(tt.tx, tt.allowed, tt.now)
			if eligible != tt.wantEligible {
				t.Errorf("IsWithinTimeWindow() eligible = %v, want %v; reason: %s", eligible, tt.wantEligible, reason)
			}
			if reason == "" {
				t.Error("IsWithinTimeWindow() returned empty reason")
			}
		})
	}
}

func TestDaysUntilExpiry(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		tx      model.Transaction
		allowed model.AllowedRefund
		now     time.Time
		want    int
	}{
		{
			name: "normal remaining days",
			tx: model.Transaction{
				ID:        "tx-1",
				Timestamp: now.Add(-10 * 24 * time.Hour),
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundSameMethod,
				MaxAgeDays: 30,
			},
			now:  now,
			want: 20,
		},
		{
			name: "no limit returns negative one",
			tx: model.Transaction{
				ID:        "tx-2",
				Timestamp: now.Add(-100 * 24 * time.Hour),
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundBankTransfer,
				MaxAgeDays: 0,
			},
			now:  now,
			want: -1,
		},
		{
			name: "expired returns negative value",
			tx: model.Transaction{
				ID:        "tx-3",
				Timestamp: now.Add(-35 * 24 * time.Hour),
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundSameMethod,
				MaxAgeDays: 30,
			},
			now:  now,
			want: -5,
		},
		{
			name: "expires today returns zero",
			tx: model.Transaction{
				ID:        "tx-4",
				Timestamp: now.Add(-30 * 24 * time.Hour),
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundSameMethod,
				MaxAgeDays: 30,
			},
			now:  now,
			want: 0,
		},
		{
			name: "created just now full window remaining",
			tx: model.Transaction{
				ID:        "tx-5",
				Timestamp: now,
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundSameMethod,
				MaxAgeDays: 7,
			},
			now:  now,
			want: 7,
		},
		{
			name: "1 day remaining",
			tx: model.Transaction{
				ID:        "tx-6",
				Timestamp: now.Add(-29 * 24 * time.Hour),
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundSameMethod,
				MaxAgeDays: 30,
			},
			now:  now,
			want: 1,
		},
		{
			name: "deeply expired returns large negative",
			tx: model.Transaction{
				ID:        "tx-7",
				Timestamp: now.Add(-365 * 24 * time.Hour),
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundSameMethod,
				MaxAgeDays: 30,
			},
			now:  now,
			want: -335,
		},
		{
			name: "partial day uses floor for days since",
			tx: model.Transaction{
				ID:        "tx-8",
				Timestamp: now.Add(-10*24*time.Hour - 23*time.Hour),
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundSameMethod,
				MaxAgeDays: 30,
			},
			now:  now,
			want: 20,
		},
		{
			name: "no limit with fresh transaction still returns negative one",
			tx: model.Transaction{
				ID:        "tx-9",
				Timestamp: now,
			},
			allowed: model.AllowedRefund{
				Method:     model.RefundAccountCredit,
				MaxAgeDays: 0,
			},
			now:  now,
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := DaysUntilExpiry(tt.tx, tt.allowed, tt.now)
			if got != tt.want {
				t.Errorf("DaysUntilExpiry() = %d, want %d", got, tt.want)
			}
		})
	}
}
