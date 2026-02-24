package testdata

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/ivanjtm/YunoChallenge/internal/model"
)

func weightedPick[T any](rng *rand.Rand, items []T, weights []float64) T {
	total := 0.0
	for _, w := range weights {
		total += w
	}
	r := rng.Float64() * total
	cum := 0.0
	for i, w := range weights {
		cum += w
		if r < cum {
			return items[i]
		}
	}
	return items[len(items)-1]
}

type countryInfo struct {
	code     model.Country
	currency model.Currency
}

func GenerateTransactions(count int, now time.Time) []model.Transaction {
	txns := make([]model.Transaction, 0, count)

	edges := []model.Transaction{
		{
			ID: "txn_edge_001", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr",
			Amount: 250, Timestamp: now.Add(-30 * time.Minute), Settled: false,
		},
		{
			ID: "txn_edge_002", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodCreditCard, ProcessorID: "mexpay",
			Amount: 1500, Timestamp: now.Add(-45 * time.Minute), Settled: false,
		},
		{
			ID: "txn_edge_003", Country: model.CountryCO, Currency: model.CurrencyCOP,
			PaymentMethod: model.MethodPSE, ProcessorID: "colpay",
			Amount: 200000, Timestamp: now.Add(-20 * time.Minute), Settled: false,
		},
		{
			ID: "txn_edge_004", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodCreditCard, ProcessorID: "mexpay",
			Amount: 800, Timestamp: now.Add(-12 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_005", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr",
			Amount: 450, Timestamp: now.Add(-86 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_006", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "globalpay",
			Amount: 320, Timestamp: now.Add(-88 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_007", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "quickrefund",
			Amount: 180, Timestamp: now.Add(-89 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_008", Country: model.CountryCO, Currency: model.CurrencyCOP,
			PaymentMethod: model.MethodPSE, ProcessorID: "colpay",
			Amount: 350000, Timestamp: now.Add(-57 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_009", Country: model.CountryCO, Currency: model.CurrencyCOP,
			PaymentMethod: model.MethodPSE, ProcessorID: "globalpay",
			Amount: 180000, Timestamp: now.Add(-59 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_010", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodOXXO, ProcessorID: "mexpay",
			Amount: 2500, Timestamp: now.Add(-45 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_011", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodOXXO, ProcessorID: "globalpay",
			Amount: 800, Timestamp: now.Add(-30 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_012", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodOXXO, ProcessorID: "mexpay",
			Amount: 3200, Timestamp: now.Add(-100 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_013", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodBoleto, ProcessorID: "paybr",
			Amount: 600, Timestamp: now.Add(-60 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_014", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodBoleto, ProcessorID: "globalpay",
			Amount: 150, Timestamp: now.Add(-20 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_015", Country: model.CountryCO, Currency: model.CurrencyCOP,
			PaymentMethod: model.MethodEfecty, ProcessorID: "colpay",
			Amount: 450000, Timestamp: now.Add(-40 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_016", Country: model.CountryCO, Currency: model.CurrencyCOP,
			PaymentMethod: model.MethodEfecty, ProcessorID: "globalpay",
			Amount: 120000, Timestamp: now.Add(-15 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_017", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodCreditCard, ProcessorID: "quickrefund",
			Amount: 4800, Timestamp: now.Add(-10 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_018", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodCreditCard, ProcessorID: "mexpay",
			Amount: 14000, Timestamp: now.Add(-5 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_019", Country: model.CountryCO, Currency: model.CurrencyCOP,
			PaymentMethod: model.MethodCreditCard, ProcessorID: "colpay",
			Amount: 4800000, Timestamp: now.Add(-8 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_020", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodPIX, ProcessorID: "paybr",
			Amount: 15, Timestamp: now.Add(-3 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_021", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodSPEI, ProcessorID: "mexpay",
			Amount: 50, Timestamp: now.Add(-2 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_022", Country: model.CountryBR, Currency: model.CurrencyBRL,
			PaymentMethod: model.MethodCreditCard, ProcessorID: "paybr",
			Amount: 500, Timestamp: now.Add(-176 * 24 * time.Hour), Settled: true,
		},
		{
			ID: "txn_edge_023", Country: model.CountryMX, Currency: model.CurrencyMXN,
			PaymentMethod: model.MethodCreditCard, ProcessorID: "globalpay",
			Amount: 2000, Timestamp: now.Add(-179 * 24 * time.Hour), Settled: true,
		},
	}

	txns = append(txns, edges...)

	rng := rand.New(rand.NewSource(42))

	countries := []countryInfo{
		{model.CountryBR, model.CurrencyBRL},
		{model.CountryMX, model.CurrencyMXN},
		{model.CountryCO, model.CurrencyCOP},
	}
	countryWeights := []float64{0.45, 0.35, 0.20}

	paymentMethods := map[model.Country][]model.PaymentMethod{
		model.CountryBR: {model.MethodPIX, model.MethodCreditCard, model.MethodBoleto},
		model.CountryMX: {model.MethodCreditCard, model.MethodOXXO, model.MethodSPEI},
		model.CountryCO: {model.MethodCreditCard, model.MethodPSE, model.MethodEfecty},
	}
	paymentWeights := map[model.Country][]float64{
		model.CountryBR: {0.50, 0.35, 0.15},
		model.CountryMX: {0.40, 0.30, 0.30},
		model.CountryCO: {0.40, 0.35, 0.25},
	}

	processors := map[model.Country][]string{
		model.CountryBR: {"paybr", "globalpay", "quickrefund", "valueproc"},
		model.CountryMX: {"mexpay", "globalpay", "quickrefund", "valueproc"},
		model.CountryCO: {"colpay", "globalpay", "valueproc"},
	}
	processorWeights := map[model.Country][]float64{
		model.CountryBR: {0.50, 0.20, 0.15, 0.15},
		model.CountryMX: {0.45, 0.20, 0.20, 0.15},
		model.CountryCO: {0.50, 0.30, 0.20},
	}

	type amountParams struct {
		mu       float64
		sigma    float64
		min      float64
		max      float64
		decimals int
	}
	amountCfg := map[model.Currency]amountParams{
		model.CurrencyBRL: {mu: math.Log(150), sigma: 1.0, min: 15.0, max: 5000.0, decimals: 2},
		model.CurrencyMXN: {mu: math.Log(500), sigma: 1.0, min: 50.0, max: 15000.0, decimals: 2},
		model.CurrencyCOP: {mu: math.Log(150000), sigma: 1.0, min: 10000.0, max: 5000000.0, decimals: 0},
	}

	remaining := count - len(edges)
	if remaining < 0 {
		remaining = 0
	}

	for i := 0; i < remaining; i++ {
		ci := weightedPick(rng, countries, countryWeights)
		cc := ci.code
		cur := ci.currency

		pm := weightedPick(rng, paymentMethods[cc], paymentWeights[cc])
		proc := weightedPick(rng, processors[cc], processorWeights[cc])

		cfg := amountCfg[cur]
		raw := math.Exp(rng.NormFloat64()*cfg.sigma + cfg.mu)
		if raw < cfg.min {
			raw = cfg.min
		}
		if raw > cfg.max {
			raw = cfg.max
		}
		if cfg.decimals == 2 {
			raw = math.Round(raw*100) / 100
		} else {
			raw = math.Round(raw)
		}

		ts := now.Add(-time.Duration(rng.Intn(180*24)) * time.Hour)

		age := now.Sub(ts)
		var settled bool
		if age > 2*time.Hour {
			settled = rng.Float64() < 0.95
		} else {
			settled = rng.Float64() < 0.20
		}

		txns = append(txns, model.Transaction{
			ID:            fmt.Sprintf("txn_%06d", len(edges)+i+1),
			Country:       cc,
			Currency:      cur,
			PaymentMethod: pm,
			ProcessorID:   proc,
			Amount:        raw,
			Timestamp:     ts,
			Settled:       settled,
			CustomerID:    fmt.Sprintf("cust_%05d", rng.Intn(50000)),
		})
	}

	return txns
}

func GenerateAndSave(path string, count int, now time.Time) error {
	txns := GenerateTransactions(count, now)

	data, err := json.MarshalIndent(txns, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal transactions: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write file %s: %w", path, err)
	}

	return nil
}
