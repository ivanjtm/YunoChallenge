package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ivanjtm/YunoChallenge/internal/model"
)

type AppConfig struct {
	Processors   []model.Processor
	Rules        []model.CompatibilityRule
	Transactions []model.Transaction
}

func Load(processorsPath, rulesPath string) (*AppConfig, error) {
	processors, err := loadProcessors(processorsPath)
	if err != nil {
		return nil, fmt.Errorf("loading processors: %w", err)
	}

	rules, err := loadRules(rulesPath)
	if err != nil {
		return nil, fmt.Errorf("loading rules: %w", err)
	}

	cfg := &AppConfig{
		Processors: processors,
		Rules:      rules,
	}

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	return cfg, nil
}

func LoadWithTransactions(processorsPath, rulesPath, transactionsPath string) (*AppConfig, error) {
	cfg, err := Load(processorsPath, rulesPath)
	if err != nil {
		return nil, err
	}

	transactions, err := loadTransactions(transactionsPath)
	if err != nil {
		return nil, fmt.Errorf("loading transactions: %w", err)
	}

	cfg.Transactions = transactions
	return cfg, nil
}

func (c *AppConfig) ProcessorByID(id string) (model.Processor, bool) {
	for _, p := range c.Processors {
		if p.ID == id {
			return p, true
		}
	}
	return model.Processor{}, false
}

func loadProcessors(path string) ([]model.Processor, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var processors []model.Processor
	if err := json.NewDecoder(f).Decode(&processors); err != nil {
		return nil, err
	}
	return processors, nil
}

func loadRules(path string) ([]model.CompatibilityRule, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var rules []model.CompatibilityRule
	if err := json.NewDecoder(f).Decode(&rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func loadTransactions(path string) ([]model.Transaction, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var transactions []model.Transaction
	if err := json.NewDecoder(f).Decode(&transactions); err != nil {
		return nil, err
	}
	return transactions, nil
}

func validate(cfg *AppConfig) error {
	for i, p := range cfg.Processors {
		if p.ID == "" {
			return fmt.Errorf("processor at index %d has empty ID", i)
		}
		if p.Name == "" {
			return fmt.Errorf("processor %q has empty Name", p.ID)
		}
		if len(p.SupportedCountries) == 0 {
			return fmt.Errorf("processor %q has no supported countries", p.ID)
		}
		if len(p.RefundFees) == 0 {
			return fmt.Errorf("processor %q has no refund fees", p.ID)
		}

		feeCurrencies := make(map[model.Currency]bool)
		for _, fee := range p.RefundFees {
			feeCurrencies[fee.Currency] = true
		}

		countryCurrency := map[model.Country]model.Currency{
			model.CountryBR: model.CurrencyBRL,
			model.CountryMX: model.CurrencyMXN,
			model.CountryCO: model.CurrencyCOP,
		}
		for _, country := range p.SupportedCountries {
			if cur, ok := countryCurrency[country]; ok {
				if !feeCurrencies[cur] {
					fmt.Printf("[WARNING] processor %q supports country %s but has no fees for currency %s\n", p.ID, country, cur)
				}
			}
		}
	}

	for i, r := range cfg.Rules {
		if r.OriginalMethod == "" {
			return fmt.Errorf("rule at index %d has empty original_method", i)
		}
	}

	return nil
}
