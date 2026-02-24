package router

import (
	"fmt"
	"testing"
	"time"

	"github.com/ivanjtm/YunoChallenge/internal/model"
)

func newTestRouter() *Router {
	return NewRouter(allProcessors(), allCompatRules())
}

func TestAnalyzeBatch_SingleTransaction(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID:            "tx-single",
			Country:       model.CountryBR,
			Currency:      model.CurrencyBRL,
			PaymentMethod: model.MethodPIX,
			ProcessorID:   "paybr",
			Amount:        200.0,
			Timestamp:     now.Add(-48 * time.Hour),
			Settled:       true,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	if result.TotalTransactions != 1 {
		t.Errorf("TotalTransactions = %d, want 1", result.TotalTransactions)
	}
	if len(result.Results) != 1 {
		t.Fatalf("len(Results) = %d, want 1", len(result.Results))
	}
	if result.Results[0].TransactionID != "tx-single" {
		t.Errorf("Results[0].TransactionID = %s, want tx-single", result.Results[0].TransactionID)
	}

	singleRoute := r.SelectRoute(txns[0], now)
	if result.Results[0].Selected.ProcessorID != singleRoute.Selected.ProcessorID {
		t.Errorf("batch route processor = %s, single route processor = %s",
			result.Results[0].Selected.ProcessorID, singleRoute.Selected.ProcessorID)
	}
	if !almostEqual(result.TotalNaiveCost, singleRoute.NaiveCost) {
		t.Errorf("TotalNaiveCost = %.2f, want %.2f", result.TotalNaiveCost, singleRoute.NaiveCost)
	}
	if !almostEqual(result.TotalSmartCost, singleRoute.Selected.EstimatedCost) {
		t.Errorf("TotalSmartCost = %.2f, want %.2f", result.TotalSmartCost, singleRoute.Selected.EstimatedCost)
	}
	if !almostEqual(result.TotalSavings, singleRoute.Savings) {
		t.Errorf("TotalSavings = %.2f, want %.2f", result.TotalSavings, singleRoute.Savings)
	}
}

func TestAnalyzeBatch_MultipleTransactions(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID: "tx-1", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 200.0,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
		{
			ID: "tx-2", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodCreditCard, ProcessorID: "mexpay", Amount: 1000.0,
			Timestamp: now.Add(-3 * time.Hour), Settled: false,
		},
		{
			ID: "tx-3", Country: model.CountryCO, Currency: model.CurrencyCOP,
			PaymentMethod: model.MethodPSE, ProcessorID: "colpay", Amount: 50000.0,
			Timestamp: now.Add(-10 * 24 * time.Hour), Settled: true,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	if result.TotalTransactions != 3 {
		t.Errorf("TotalTransactions = %d, want 3", result.TotalTransactions)
	}
	if len(result.Results) != 3 {
		t.Fatalf("len(Results) = %d, want 3", len(result.Results))
	}

	seenIDs := make(map[string]bool)
	for _, rr := range result.Results {
		seenIDs[rr.TransactionID] = true
	}
	for _, tx := range txns {
		if !seenIDs[tx.ID] {
			t.Errorf("transaction %s not found in results", tx.ID)
		}
	}
}

func TestAnalyzeBatch_EmptyBatch(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	result := r.AnalyzeBatch(nil, now)

	if result.TotalTransactions != 0 {
		t.Errorf("TotalTransactions = %d, want 0", result.TotalTransactions)
	}
	if result.Results == nil {
		t.Error("Results is nil, want non-nil empty slice")
	}
	if len(result.Results) != 0 {
		t.Errorf("len(Results) = %d, want 0", len(result.Results))
	}
	if result.TotalNaiveCost != 0 {
		t.Errorf("TotalNaiveCost = %.2f, want 0", result.TotalNaiveCost)
	}
	if result.TotalSmartCost != 0 {
		t.Errorf("TotalSmartCost = %.2f, want 0", result.TotalSmartCost)
	}
	if result.TotalSavings != 0 {
		t.Errorf("TotalSavings = %.2f, want 0", result.TotalSavings)
	}
	if result.SavingsPercent != 0 {
		t.Errorf("SavingsPercent = %.2f, want 0", result.SavingsPercent)
	}
	if result.ByProcessor == nil {
		t.Error("ByProcessor is nil, want non-nil map")
	}
	if result.ByPaymentMethod == nil {
		t.Error("ByPaymentMethod is nil, want non-nil map")
	}
	if result.TimeSensitive == nil {
		t.Error("TimeSensitive is nil, want non-nil slice")
	}
	if result.LimitedOptions == nil {
		t.Error("LimitedOptions is nil, want non-nil slice")
	}
}

func TestAnalyzeBatch_EmptySlice(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	result := r.AnalyzeBatch([]model.Transaction{}, now)

	if result.TotalTransactions != 0 {
		t.Errorf("TotalTransactions = %d, want 0", result.TotalTransactions)
	}
	if len(result.Results) != 0 {
		t.Errorf("len(Results) = %d, want 0", len(result.Results))
	}
	if result.ByProcessor == nil {
		t.Error("ByProcessor is nil, want non-nil map")
	}
	if result.ByPaymentMethod == nil {
		t.Error("ByPaymentMethod is nil, want non-nil map")
	}
	if result.TimeSensitive == nil {
		t.Error("TimeSensitive is nil, want non-nil slice")
	}
	if result.LimitedOptions == nil {
		t.Error("LimitedOptions is nil, want non-nil slice")
	}
}

func TestAnalyzeBatch_SavingsAccumulation(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID: "tx-s1", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 200.0,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
		{
			ID: "tx-s2", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodSPEI, ProcessorID: "mexpay", Amount: 500.0,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
		{
			ID: "tx-s3", Country: model.CountryCO, Currency: model.CurrencyCOP,
			PaymentMethod: model.MethodCreditCard, ProcessorID: "colpay", Amount: 100000.0,
			Timestamp: now.Add(-3 * time.Hour), Settled: false,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	var expectedNaive, expectedSmart, expectedSavings float64
	for _, tx := range txns {
		route := r.SelectRoute(tx, now)
		expectedNaive += route.NaiveCost
		expectedSmart += route.Selected.EstimatedCost
		expectedSavings += route.Savings
	}
	expectedNaive = roundTo2(expectedNaive)
	expectedSmart = roundTo2(expectedSmart)
	expectedSavings = roundTo2(expectedSavings)

	if !almostEqual(result.TotalNaiveCost, expectedNaive) {
		t.Errorf("TotalNaiveCost = %.2f, want %.2f", result.TotalNaiveCost, expectedNaive)
	}
	if !almostEqual(result.TotalSmartCost, expectedSmart) {
		t.Errorf("TotalSmartCost = %.2f, want %.2f", result.TotalSmartCost, expectedSmart)
	}
	if !almostEqual(result.TotalSavings, expectedSavings) {
		t.Errorf("TotalSavings = %.2f, want %.2f", result.TotalSavings, expectedSavings)
	}

	if result.TotalNaiveCost > 0 {
		expectedPct := roundTo2((result.TotalSavings / result.TotalNaiveCost) * 100)
		if !almostEqual(result.SavingsPercent, expectedPct) {
			t.Errorf("SavingsPercent = %.2f, want %.2f", result.SavingsPercent, expectedPct)
		}
	}
}

func TestAnalyzeBatch_SavingsByProcessor(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID: "tx-bp1", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 200.0,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
		{
			ID: "tx-bp2", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodCreditCard, ProcessorID: "paybr", Amount: 500.0,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
		{
			ID: "tx-bp3", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodCreditCard, ProcessorID: "mexpay", Amount: 1000.0,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	if len(result.ByProcessor) == 0 {
		t.Fatal("ByProcessor is empty")
	}

	paybr, ok := result.ByProcessor["paybr"]
	if !ok {
		t.Fatal("ByProcessor missing paybr entry")
	}
	if paybr.TransactionCount != 2 {
		t.Errorf("paybr TransactionCount = %d, want 2", paybr.TransactionCount)
	}
	if paybr.ProcessorID != "paybr" {
		t.Errorf("paybr ProcessorID = %s, want paybr", paybr.ProcessorID)
	}

	mexpay, ok := result.ByProcessor["mexpay"]
	if !ok {
		t.Fatal("ByProcessor missing mexpay entry")
	}
	if mexpay.TransactionCount != 1 {
		t.Errorf("mexpay TransactionCount = %d, want 1", mexpay.TransactionCount)
	}

	var totalProcSavings float64
	for _, ps := range result.ByProcessor {
		totalProcSavings += ps.Savings
	}
	if !almostEqual(totalProcSavings, result.TotalSavings) {
		t.Errorf("sum of processor savings = %.2f, TotalSavings = %.2f", totalProcSavings, result.TotalSavings)
	}
}

func TestAnalyzeBatch_SavingsByMethod(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID: "tx-bm1", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 300.0,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
		{
			ID: "tx-bm2", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 700.0,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
		{
			ID: "tx-bm3", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodOXXO, ProcessorID: "mexpay", Amount: 500.0,
			Timestamp: now.Add(-5 * 24 * time.Hour), Settled: true,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	if len(result.ByPaymentMethod) == 0 {
		t.Fatal("ByPaymentMethod is empty")
	}

	pix, ok := result.ByPaymentMethod["PIX"]
	if !ok {
		t.Fatal("ByPaymentMethod missing PIX entry")
	}
	if pix.TransactionCount != 2 {
		t.Errorf("PIX TransactionCount = %d, want 2", pix.TransactionCount)
	}
	if pix.Method != "PIX" {
		t.Errorf("PIX Method = %s, want PIX", pix.Method)
	}

	oxxo, ok := result.ByPaymentMethod["OXXO"]
	if !ok {
		t.Fatal("ByPaymentMethod missing OXXO entry")
	}
	if oxxo.TransactionCount != 1 {
		t.Errorf("OXXO TransactionCount = %d, want 1", oxxo.TransactionCount)
	}

	var totalMethodSavings float64
	for _, ms := range result.ByPaymentMethod {
		totalMethodSavings += ms.Savings
	}
	if !almostEqual(totalMethodSavings, result.TotalSavings) {
		t.Errorf("sum of method savings = %.2f, TotalSavings = %.2f", totalMethodSavings, result.TotalSavings)
	}
}

func TestAnalyzeBatch_TimeSensitive_PIXNearExpiry(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID: "tx-pix-expiring", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 200.0,
			Timestamp: now.Add(-85 * 24 * time.Hour), Settled: true,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	if len(result.TimeSensitive) == 0 {
		t.Fatal("expected TimeSensitive flags for PIX at 85 days (within 15-day threshold of 90-day window)")
	}

	found := false
	for _, flag := range result.TimeSensitive {
		if flag.TransactionID == "tx-pix-expiring" {
			found = true
			if flag.DaysRemaining > 15 {
				t.Errorf("DaysRemaining = %d, want <= 15", flag.DaysRemaining)
			}
		}
	}
	if !found {
		t.Error("tx-pix-expiring not found in TimeSensitive flags")
	}
}

func TestAnalyzeBatch_TimeSensitive_ReversalNearExpiry(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID: "tx-reversal-expiring", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 200.0,
			Timestamp: now.Add(-20 * time.Hour), Settled: false,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	if len(result.TimeSensitive) == 0 {
		t.Fatal("expected TimeSensitive flags for reversal candidate at 20 hours (>= 18h and < 24h)")
	}

	found := false
	for _, flag := range result.TimeSensitive {
		if flag.TransactionID == "tx-reversal-expiring" && flag.WindowType == "REVERSAL_24H" {
			found = true
		}
	}
	if !found {
		t.Error("REVERSAL_24H flag not found for tx-reversal-expiring")
	}
}

func TestAnalyzeBatch_TimeSensitive_NoFlagWhenFarFromExpiry(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID: "tx-pix-fresh", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 200.0,
			Timestamp: now.Add(-10 * 24 * time.Hour), Settled: true,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	for _, flag := range result.TimeSensitive {
		if flag.TransactionID == "tx-pix-fresh" {
			t.Errorf("unexpected TimeSensitive flag for tx-pix-fresh at 10 days (80 remaining, threshold 15): %s", flag.WindowType)
		}
	}
}

func TestAnalyzeBatch_LimitedOptions(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	tests := []struct {
		name   string
		tx     model.Transaction
		method string
	}{
		{
			name: "OXXO flagged as limited",
			tx: model.Transaction{
				ID: "tx-oxxo-ltd", Country: model.CountryMX, Currency: model.CurrencyMXN,
				PaymentMethod: model.MethodOXXO, ProcessorID: "mexpay", Amount: 500.0,
				Timestamp: now.Add(-5 * 24 * time.Hour), Settled: true,
			},
			method: "OXXO",
		},
		{
			name: "BOLETO flagged as limited",
			tx: model.Transaction{
				ID: "tx-boleto-ltd", Country: model.CountryBR, Currency: model.CurrencyBRL,
				PaymentMethod: model.MethodBoleto, ProcessorID: "paybr", Amount: 300.0,
				Timestamp: now.Add(-5 * 24 * time.Hour), Settled: true,
			},
			method: "BOLETO",
		},
		{
			name: "EFECTY flagged as limited",
			tx: model.Transaction{
				ID: "tx-efecty-ltd", Country: model.CountryCO, Currency: model.CurrencyCOP,
				PaymentMethod: model.MethodEfecty, ProcessorID: "colpay", Amount: 50000.0,
				Timestamp: now.Add(-5 * 24 * time.Hour), Settled: true,
			},
			method: "EFECTY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := r.AnalyzeBatch([]model.Transaction{tt.tx}, now)

			if len(result.LimitedOptions) == 0 {
				t.Fatalf("expected LimitedOptions flag for %s", tt.method)
			}

			found := false
			for _, flag := range result.LimitedOptions {
				if flag.TransactionID == tt.tx.ID {
					found = true
					if flag.OriginalMethod != tt.method {
						t.Errorf("OriginalMethod = %s, want %s", flag.OriginalMethod, tt.method)
					}
					if flag.AvailableOptions < 1 {
						t.Errorf("AvailableOptions = %d, want >= 1", flag.AvailableOptions)
					}
					if flag.Message == "" {
						t.Error("Message is empty")
					}
				}
			}
			if !found {
				t.Errorf("LimitedOptions flag not found for %s", tt.tx.ID)
			}
		})
	}
}

func TestAnalyzeBatch_NoLimitedOptionsForRegularMethods(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID: "tx-pix-nolimit", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 200.0,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
		{
			ID: "tx-cc-nolimit", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodCreditCard, ProcessorID: "mexpay", Amount: 1000.0,
			Timestamp: now.Add(-3 * time.Hour), Settled: false,
		},
		{
			ID: "tx-spei-nolimit", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodSPEI, ProcessorID: "mexpay", Amount: 800.0,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	if len(result.LimitedOptions) != 0 {
		t.Errorf("expected 0 LimitedOptions for PIX/CC/SPEI, got %d", len(result.LimitedOptions))
	}
}

func TestAnalyzeBatch_ResultOrderMatchesInput(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := make([]model.Transaction, 20)
	for i := range txns {
		txns[i] = model.Transaction{
			ID: fmt.Sprintf("tx-order-%d", i), Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: float64(100 + i*50),
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		}
	}

	result := r.AnalyzeBatch(txns, now)

	if len(result.Results) != len(txns) {
		t.Fatalf("len(Results) = %d, want %d", len(result.Results), len(txns))
	}

	for i, rr := range result.Results {
		if rr.TransactionID != txns[i].ID {
			t.Errorf("Results[%d].TransactionID = %s, want %s", i, rr.TransactionID, txns[i].ID)
		}
	}
}

func TestAnalyzeBatch_Concurrency_ResultCountMatchesInput(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	n := 150
	txns := make([]model.Transaction, n)
	methods := []struct {
		m model.PaymentMethod
		c model.Country
		u model.Currency
		p string
	}{
		{model.MethodPIX, model.CountryBR, model.CurrencyBRL, "paybr"},
		{model.MethodCreditCard, model.CountryMX, model.CurrencyMXN, "mexpay"},
		{model.MethodPSE, model.CountryCO, model.CurrencyCOP, "colpay"},
		{model.MethodOXXO, model.CountryMX, model.CurrencyMXN, "mexpay"},
		{model.MethodBoleto, model.CountryBR, model.CurrencyBRL, "paybr"},
		{model.MethodEfecty, model.CountryCO, model.CurrencyCOP, "colpay"},
		{model.MethodSPEI, model.CountryMX, model.CurrencyMXN, "mexpay"},
		{model.MethodCreditCard, model.CountryBR, model.CurrencyBRL, "paybr"},
	}

	for i := 0; i < n; i++ {
		m := methods[i%len(methods)]
		txns[i] = model.Transaction{
			ID: fmt.Sprintf("tx-conc-%d", i), Country: m.c, Currency: m.u,
			PaymentMethod: m.m, ProcessorID: m.p, Amount: float64(100 + i*10),
			Timestamp: now.Add(-time.Duration(i+1) * 24 * time.Hour), Settled: i%2 == 0,
		}
	}

	result := r.AnalyzeBatch(txns, now)

	if result.TotalTransactions != n {
		t.Errorf("TotalTransactions = %d, want %d", result.TotalTransactions, n)
	}
	if len(result.Results) != n {
		t.Fatalf("len(Results) = %d, want %d", len(result.Results), n)
	}

	for i, rr := range result.Results {
		if rr.TransactionID != txns[i].ID {
			t.Errorf("Results[%d].TransactionID = %s, want %s", i, rr.TransactionID, txns[i].ID)
		}
		if rr.Selected.ProcessorID == "" {
			t.Errorf("Results[%d].Selected.ProcessorID is empty", i)
		}
		if rr.Selected.RefundMethod == "" {
			t.Errorf("Results[%d].Selected.RefundMethod is empty", i)
		}
	}
}

func TestAnalyzeBatch_Concurrency_Deterministic(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	n := 100
	txns := make([]model.Transaction, n)
	for i := 0; i < n; i++ {
		txns[i] = model.Transaction{
			ID: fmt.Sprintf("tx-det-%d", i), Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: float64(100 + i*25),
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		}
	}

	result1 := r.AnalyzeBatch(txns, now)
	result2 := r.AnalyzeBatch(txns, now)

	if !almostEqual(result1.TotalNaiveCost, result2.TotalNaiveCost) {
		t.Errorf("run1 TotalNaiveCost = %.2f, run2 = %.2f", result1.TotalNaiveCost, result2.TotalNaiveCost)
	}
	if !almostEqual(result1.TotalSmartCost, result2.TotalSmartCost) {
		t.Errorf("run1 TotalSmartCost = %.2f, run2 = %.2f", result1.TotalSmartCost, result2.TotalSmartCost)
	}
	if !almostEqual(result1.TotalSavings, result2.TotalSavings) {
		t.Errorf("run1 TotalSavings = %.2f, run2 = %.2f", result1.TotalSavings, result2.TotalSavings)
	}
	if !almostEqual(result1.SavingsPercent, result2.SavingsPercent) {
		t.Errorf("run1 SavingsPercent = %.2f, run2 = %.2f", result1.SavingsPercent, result2.SavingsPercent)
	}

	for i := 0; i < n; i++ {
		if result1.Results[i].TransactionID != result2.Results[i].TransactionID {
			t.Errorf("Results[%d] TransactionID mismatch: %s vs %s",
				i, result1.Results[i].TransactionID, result2.Results[i].TransactionID)
		}
		if result1.Results[i].Selected.ProcessorID != result2.Results[i].Selected.ProcessorID {
			t.Errorf("Results[%d] Selected.ProcessorID mismatch: %s vs %s",
				i, result1.Results[i].Selected.ProcessorID, result2.Results[i].Selected.ProcessorID)
		}
		if !almostEqual(result1.Results[i].Selected.EstimatedCost, result2.Results[i].Selected.EstimatedCost) {
			t.Errorf("Results[%d] Selected.EstimatedCost mismatch: %.2f vs %.2f",
				i, result1.Results[i].Selected.EstimatedCost, result2.Results[i].Selected.EstimatedCost)
		}
	}
}

func TestAnalyzeBatch_Concurrency_MixedPaymentMethods(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	n := 120
	txns := make([]model.Transaction, n)
	scenarios := []struct {
		method    model.PaymentMethod
		country   model.Country
		currency  model.Currency
		processor string
		settled   bool
		ageHours  int
	}{
		{model.MethodPIX, model.CountryBR, model.CurrencyBRL, "paybr", false, 2},
		{model.MethodPIX, model.CountryBR, model.CurrencyBRL, "paybr", true, 48},
		{model.MethodCreditCard, model.CountryBR, model.CurrencyBRL, "paybr", false, 5},
		{model.MethodCreditCard, model.CountryMX, model.CurrencyMXN, "mexpay", true, 72},
		{model.MethodOXXO, model.CountryMX, model.CurrencyMXN, "mexpay", true, 120},
		{model.MethodBoleto, model.CountryBR, model.CurrencyBRL, "paybr", true, 120},
		{model.MethodPSE, model.CountryCO, model.CurrencyCOP, "colpay", true, 240},
		{model.MethodEfecty, model.CountryCO, model.CurrencyCOP, "colpay", true, 240},
		{model.MethodSPEI, model.CountryMX, model.CurrencyMXN, "mexpay", false, 3},
		{model.MethodCreditCard, model.CountryCO, model.CurrencyCOP, "colpay", false, 10},
	}

	for i := 0; i < n; i++ {
		s := scenarios[i%len(scenarios)]
		txns[i] = model.Transaction{
			ID: fmt.Sprintf("tx-mix-%d", i), Country: s.country, Currency: s.currency,
			PaymentMethod: s.method, ProcessorID: s.processor, Amount: float64(200 + i*5),
			Timestamp: now.Add(-time.Duration(s.ageHours) * time.Hour), Settled: s.settled,
		}
	}

	result := r.AnalyzeBatch(txns, now)

	if len(result.Results) != n {
		t.Fatalf("len(Results) = %d, want %d", len(result.Results), n)
	}

	for i, rr := range result.Results {
		if rr.TransactionID != txns[i].ID {
			t.Errorf("Results[%d].TransactionID = %s, want %s", i, rr.TransactionID, txns[i].ID)
		}
	}

	if result.TotalNaiveCost < 0 {
		t.Errorf("TotalNaiveCost = %.2f, want >= 0", result.TotalNaiveCost)
	}
	if result.TotalSmartCost < 0 {
		t.Errorf("TotalSmartCost = %.2f, want >= 0", result.TotalSmartCost)
	}
	if result.TotalSavings < 0 {
		t.Errorf("TotalSavings = %.2f, want >= 0", result.TotalSavings)
	}
}

func TestAnalyzeBatch_SavingsEqualsSumOfIndividual(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID: "tx-sum1", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 150.0,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
		{
			ID: "tx-sum2", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodCreditCard, ProcessorID: "mexpay", Amount: 2000.0,
			Timestamp: now.Add(-72 * time.Hour), Settled: true,
		},
		{
			ID: "tx-sum3", Country: model.CountryCO, Currency: model.CurrencyCOP,
			PaymentMethod: model.MethodEfecty, ProcessorID: "colpay", Amount: 75000.0,
			Timestamp: now.Add(-10 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "tx-sum4", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodCreditCard, ProcessorID: "paybr", Amount: 800.0,
			Timestamp: now.Add(-2 * time.Hour), Settled: false,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	var sumNaive, sumSmart, sumSavings float64
	for _, rr := range result.Results {
		sumNaive += rr.NaiveCost
		sumSmart += rr.Selected.EstimatedCost
		sumSavings += rr.Savings
	}
	sumNaive = roundTo2(sumNaive)
	sumSmart = roundTo2(sumSmart)
	sumSavings = roundTo2(sumSavings)

	if !almostEqual(result.TotalNaiveCost, sumNaive) {
		t.Errorf("TotalNaiveCost = %.2f, sum of individual NaiveCost = %.2f", result.TotalNaiveCost, sumNaive)
	}
	if !almostEqual(result.TotalSmartCost, sumSmart) {
		t.Errorf("TotalSmartCost = %.2f, sum of individual SmartCost = %.2f", result.TotalSmartCost, sumSmart)
	}
	if !almostEqual(result.TotalSavings, sumSavings) {
		t.Errorf("TotalSavings = %.2f, sum of individual Savings = %.2f", result.TotalSavings, sumSavings)
	}
}

func TestAnalyzeBatch_ByProcessorCostConsistency(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID: "tx-pc1", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 200.0,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
		{
			ID: "tx-pc2", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodCreditCard, ProcessorID: "mexpay", Amount: 1000.0,
			Timestamp: now.Add(-72 * time.Hour), Settled: true,
		},
		{
			ID: "tx-pc3", Country: model.CountryCO, Currency: model.CurrencyCOP,
			PaymentMethod: model.MethodPSE, ProcessorID: "colpay", Amount: 50000.0,
			Timestamp: now.Add(-10 * 24 * time.Hour), Settled: true,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	for procID, ps := range result.ByProcessor {
		expectedSavings := ps.NaiveCost - ps.SmartCost
		if !almostEqual(ps.Savings, expectedSavings) {
			t.Errorf("ByProcessor[%s] Savings = %.2f, want NaiveCost(%.2f) - SmartCost(%.2f) = %.2f",
				procID, ps.Savings, ps.NaiveCost, ps.SmartCost, expectedSavings)
		}
	}
}

func TestAnalyzeBatch_ByMethodCostConsistency(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID: "tx-mc1", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 200.0,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
		{
			ID: "tx-mc2", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodCreditCard, ProcessorID: "mexpay", Amount: 1000.0,
			Timestamp: now.Add(-72 * time.Hour), Settled: true,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	for method, ms := range result.ByPaymentMethod {
		expectedSavings := ms.NaiveCost - ms.SmartCost
		if !almostEqual(ms.Savings, expectedSavings) {
			t.Errorf("ByPaymentMethod[%s] Savings = %.2f, want NaiveCost(%.2f) - SmartCost(%.2f) = %.2f",
				method, ms.Savings, ms.NaiveCost, ms.SmartCost, expectedSavings)
		}
	}
}

func TestAnalyzeBatch_SavingsPercentCalculation(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID: "tx-pct1", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodCreditCard, ProcessorID: "paybr", Amount: 1000.0,
			Timestamp: now.Add(-2 * time.Hour), Settled: false,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	if result.TotalNaiveCost > 0 {
		expectedPct := roundTo2((result.TotalSavings / result.TotalNaiveCost) * 100)
		if !almostEqual(result.SavingsPercent, expectedPct) {
			t.Errorf("SavingsPercent = %.2f, want %.2f", result.SavingsPercent, expectedPct)
		}
	}
}

func TestAnalyzeBatch_SavingsPercentZeroWhenNoNaiveCost(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	r := NewRouter(nil, nil)

	result := r.AnalyzeBatch(nil, now)

	if result.SavingsPercent != 0 {
		t.Errorf("SavingsPercent = %.2f, want 0 when no transactions", result.SavingsPercent)
	}
}

func TestAnalyzeBatch_LimitedOptionsMessage(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID: "tx-oxxo-msg", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodOXXO, ProcessorID: "mexpay", Amount: 500.0,
			Timestamp: now.Add(-5 * 24 * time.Hour), Settled: true,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	if len(result.LimitedOptions) == 0 {
		t.Fatal("expected LimitedOptions flag for OXXO")
	}

	flag := result.LimitedOptions[0]
	if flag.TransactionID != "tx-oxxo-msg" {
		t.Errorf("TransactionID = %s, want tx-oxxo-msg", flag.TransactionID)
	}
	if flag.OriginalMethod != "OXXO" {
		t.Errorf("OriginalMethod = %s, want OXXO", flag.OriginalMethod)
	}

	route := r.SelectRoute(txns[0], now)
	expectedOptions := 1 + len(route.Alternatives)
	if flag.AvailableOptions != expectedOptions {
		t.Errorf("AvailableOptions = %d, want %d", flag.AvailableOptions, expectedOptions)
	}
}

func TestAnalyzeBatch_MultipleTimeSensitiveFlags(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID: "tx-ts1", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 200.0,
			Timestamp: now.Add(-85 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "tx-ts2", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 300.0,
			Timestamp: now.Add(-20 * time.Hour), Settled: false,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	ts1Found := false
	ts2Found := false
	for _, flag := range result.TimeSensitive {
		if flag.TransactionID == "tx-ts1" {
			ts1Found = true
		}
		if flag.TransactionID == "tx-ts2" && flag.WindowType == "REVERSAL_24H" {
			ts2Found = true
		}
	}

	if !ts1Found {
		t.Error("tx-ts1 (PIX near 90-day expiry) not found in TimeSensitive")
	}
	if !ts2Found {
		t.Error("tx-ts2 (reversal near 24h expiry) not found in TimeSensitive")
	}
}

func TestAnalyzeBatch_AllLimitedMethodsInOneBatch(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID: "tx-lo-oxxo", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodOXXO, ProcessorID: "mexpay", Amount: 500.0,
			Timestamp: now.Add(-5 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "tx-lo-boleto", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodBoleto, ProcessorID: "paybr", Amount: 300.0,
			Timestamp: now.Add(-5 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "tx-lo-efecty", Country: model.CountryCO, Currency: model.CurrencyCOP,
			PaymentMethod: model.MethodEfecty, ProcessorID: "colpay", Amount: 50000.0,
			Timestamp: now.Add(-5 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "tx-lo-pix", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 200.0,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	if len(result.LimitedOptions) != 3 {
		t.Errorf("expected 3 LimitedOptions (OXXO, BOLETO, EFECTY), got %d", len(result.LimitedOptions))
	}

	flaggedIDs := make(map[string]bool)
	for _, flag := range result.LimitedOptions {
		flaggedIDs[flag.TransactionID] = true
	}

	for _, wantID := range []string{"tx-lo-oxxo", "tx-lo-boleto", "tx-lo-efecty"} {
		if !flaggedIDs[wantID] {
			t.Errorf("expected %s in LimitedOptions", wantID)
		}
	}
	if flaggedIDs["tx-lo-pix"] {
		t.Error("PIX should not be in LimitedOptions")
	}
}

func TestAnalyzeBatch_TransactionCountInSummaries(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID: "tx-cnt1", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 100.0,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
		{
			ID: "tx-cnt2", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 200.0,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
		{
			ID: "tx-cnt3", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 300.0,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	var totalByProc, totalByMethod int
	for _, ps := range result.ByProcessor {
		totalByProc += ps.TransactionCount
	}
	for _, ms := range result.ByPaymentMethod {
		totalByMethod += ms.TransactionCount
	}

	if totalByProc != 3 {
		t.Errorf("sum of ByProcessor transaction counts = %d, want 3", totalByProc)
	}
	if totalByMethod != 3 {
		t.Errorf("sum of ByPaymentMethod transaction counts = %d, want 3", totalByMethod)
	}
}

func TestAnalyzeBatch_RoundedValues(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID: "tx-round1", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 333.33,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
		{
			ID: "tx-round2", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodSPEI, ProcessorID: "mexpay", Amount: 777.77,
			Timestamp: now.Add(-48 * time.Hour), Settled: true,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	checkRounded := func(name string, value float64) {
		rounded := roundTo2(value)
		if !almostEqual(value, rounded) {
			t.Errorf("%s = %.10f, not properly rounded to 2 decimal places (expected %.2f)", name, value, rounded)
		}
	}

	checkRounded("TotalNaiveCost", result.TotalNaiveCost)
	checkRounded("TotalSmartCost", result.TotalSmartCost)
	checkRounded("TotalSavings", result.TotalSavings)
	checkRounded("SavingsPercent", result.SavingsPercent)
}

func TestAnalyzeBatch_BatchMatchesSingleRoutes(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	r := newTestRouter()

	txns := []model.Transaction{
		{
			ID: "tx-match1", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr", Amount: 500.0,
			Timestamp: now.Add(-2 * time.Hour), Settled: false,
		},
		{
			ID: "tx-match2", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodOXXO, ProcessorID: "mexpay", Amount: 1000.0,
			Timestamp: now.Add(-5 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "tx-match3", Country: model.CountryCO, Currency: model.CurrencyCOP,
			PaymentMethod: model.MethodCreditCard, ProcessorID: "colpay", Amount: 200000.0,
			Timestamp: now.Add(-100 * 24 * time.Hour), Settled: true,
		},
	}

	result := r.AnalyzeBatch(txns, now)

	for i, tx := range txns {
		singleRoute := r.SelectRoute(tx, now)
		batchRoute := result.Results[i]

		if batchRoute.Selected.ProcessorID != singleRoute.Selected.ProcessorID {
			t.Errorf("tx %s: batch ProcessorID = %s, single = %s",
				tx.ID, batchRoute.Selected.ProcessorID, singleRoute.Selected.ProcessorID)
		}
		if batchRoute.Selected.RefundMethod != singleRoute.Selected.RefundMethod {
			t.Errorf("tx %s: batch RefundMethod = %s, single = %s",
				tx.ID, batchRoute.Selected.RefundMethod, singleRoute.Selected.RefundMethod)
		}
		if !almostEqual(batchRoute.Selected.EstimatedCost, singleRoute.Selected.EstimatedCost) {
			t.Errorf("tx %s: batch EstimatedCost = %.2f, single = %.2f",
				tx.ID, batchRoute.Selected.EstimatedCost, singleRoute.Selected.EstimatedCost)
		}
		if !almostEqual(batchRoute.NaiveCost, singleRoute.NaiveCost) {
			t.Errorf("tx %s: batch NaiveCost = %.2f, single = %.2f",
				tx.ID, batchRoute.NaiveCost, singleRoute.NaiveCost)
		}
		if !almostEqual(batchRoute.Savings, singleRoute.Savings) {
			t.Errorf("tx %s: batch Savings = %.2f, single = %.2f",
				tx.ID, batchRoute.Savings, singleRoute.Savings)
		}
		if len(batchRoute.Alternatives) != len(singleRoute.Alternatives) {
			t.Errorf("tx %s: batch alternatives = %d, single = %d",
				tx.ID, len(batchRoute.Alternatives), len(singleRoute.Alternatives))
		}
	}
}

func TestRoundTo2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{name: "exact", in: 1.50, want: 1.50},
		{name: "round up", in: 1.555, want: 1.56},
		{name: "round down", in: 1.554, want: 1.55},
		{name: "zero", in: 0.0, want: 0.0},
		{name: "large", in: 123456.789, want: 123456.79},
		{name: "negative", in: -3.456, want: -3.45},
		{name: "many decimals", in: 0.1234567, want: 0.12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := roundTo2(tt.in)
			if !almostEqual(got, tt.want) {
				t.Errorf("roundTo2(%.10f) = %.10f, want %.2f", tt.in, got, tt.want)
			}
		})
	}
}
