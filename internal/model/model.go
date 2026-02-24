package model

import "time"

// Country codes
type Country string

const (
	CountryBR Country = "BR"
	CountryMX Country = "MX"
	CountryCO Country = "CO"
)

// Currencies
type Currency string

const (
	CurrencyBRL Currency = "BRL"
	CurrencyMXN Currency = "MXN"
	CurrencyCOP Currency = "COP"
)

// Payment methods (original)
type PaymentMethod string

const (
	MethodPIX        PaymentMethod = "PIX"
	MethodBoleto     PaymentMethod = "BOLETO"
	MethodCreditCard PaymentMethod = "CREDIT_CARD"
	MethodOXXO       PaymentMethod = "OXXO"
	MethodSPEI       PaymentMethod = "SPEI"
	MethodPSE        PaymentMethod = "PSE"
	MethodEfecty     PaymentMethod = "EFECTY"
)

// Refund methods (how money gets back to customer)
type RefundMethod string

const (
	RefundReversal      RefundMethod = "REVERSAL"       // void, free, <24h unsettled
	RefundSameMethod    RefundMethod = "SAME_METHOD"    // e.g. PIX->PIX, Card->Card
	RefundBankTransfer  RefundMethod = "BANK_TRANSFER"  // fallback, always available
	RefundAccountCredit RefundMethod = "ACCOUNT_CREDIT" // marketplace balance credit
)

// Transaction represents an original payment
type Transaction struct {
	ID            string        `json:"id"`
	Country       Country       `json:"country"`
	Currency      Currency      `json:"currency"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	ProcessorID   string        `json:"processor_id"`
	Amount        float64       `json:"amount"`
	Timestamp     time.Time     `json:"timestamp"`
	Settled       bool          `json:"settled"`
	CustomerID    string        `json:"customer_id"`
}

// Processor config
type Processor struct {
	ID                  string               `json:"id"`
	Name                string               `json:"name"`
	SupportedCountries  []Country            `json:"supported_countries"`
	SupportedCurrencies []Currency           `json:"supported_currencies"`
	RefundFees          []RefundMethodFee    `json:"refund_fees"`
	DailyQuota          int                  `json:"daily_quota"`
	ProcessingDays      map[RefundMethod]int `json:"processing_days"`
}

type RefundMethodFee struct {
	Method         RefundMethod    `json:"method"`
	PaymentMethods []PaymentMethod `json:"payment_methods"`
	Currency       Currency        `json:"currency"`
	BaseFee        float64         `json:"base_fee"`
	PercentFee     float64         `json:"percent_fee"`
	MinFee         float64         `json:"min_fee"`
	MaxFee         float64         `json:"max_fee"` // 0 = no cap
}

// Compatibility rules (loaded from rules.json)
type CompatibilityRule struct {
	OriginalMethod PaymentMethod   `json:"original_method"`
	Country        Country         `json:"country"`
	AllowedRefunds []AllowedRefund `json:"allowed_refunds"`
}

type AllowedRefund struct {
	Method         RefundMethod `json:"method"`
	MaxAgeDays     int          `json:"max_age_days"`    // 0 = no limit
	RequireSettled *bool        `json:"require_settled"` // nil = don't care
}

// --- Routing results ---

type RefundCandidate struct {
	ProcessorID    string       `json:"processor_id"`
	ProcessorName  string       `json:"processor_name"`
	RefundMethod   RefundMethod `json:"refund_method"`
	EstimatedCost  float64      `json:"estimated_cost"`
	ProcessingDays int          `json:"processing_days"`
	Reasoning      string       `json:"reasoning"`
}

type RefundRouteResult struct {
	TransactionID string            `json:"transaction_id"`
	Selected      RefundCandidate   `json:"selected"`
	Alternatives  []RefundCandidate `json:"alternatives"`
	NaiveCost     float64           `json:"naive_cost"`
	Savings       float64           `json:"savings"`
}

// --- Batch types ---

type BatchRefundRequest struct {
	Transactions []Transaction `json:"transactions"`
}

type BatchRefundResult struct {
	TotalTransactions int                         `json:"total_transactions"`
	TotalNaiveCost    float64                     `json:"total_naive_cost"`
	TotalSmartCost    float64                     `json:"total_smart_cost"`
	TotalSavings      float64                     `json:"total_savings"`
	SavingsPercent    float64                     `json:"savings_percent"`
	Results           []RefundRouteResult         `json:"results"`
	ByProcessor       map[string]ProcessorSummary `json:"by_processor"`
	ByPaymentMethod   map[string]MethodSummary    `json:"by_payment_method"`
	TimeSensitive     []TimeSensitiveFlag         `json:"time_sensitive"`
	LimitedOptions    []LimitedOptionFlag         `json:"limited_options"`
}

type ProcessorSummary struct {
	ProcessorID      string  `json:"processor_id"`
	NaiveCost        float64 `json:"naive_cost"`
	SmartCost        float64 `json:"smart_cost"`
	Savings          float64 `json:"savings"`
	TransactionCount int     `json:"transaction_count"`
}

type MethodSummary struct {
	Method           string  `json:"method"`
	NaiveCost        float64 `json:"naive_cost"`
	SmartCost        float64 `json:"smart_cost"`
	Savings          float64 `json:"savings"`
	TransactionCount int     `json:"transaction_count"`
}

type TimeSensitiveFlag struct {
	TransactionID string    `json:"transaction_id"`
	WindowType    string    `json:"window_type"`
	ExpiresAt     time.Time `json:"expires_at"`
	DaysRemaining int       `json:"days_remaining"`
	Message       string    `json:"message"`
}

type LimitedOptionFlag struct {
	TransactionID    string `json:"transaction_id"`
	OriginalMethod   string `json:"original_method"`
	AvailableOptions int    `json:"available_options"`
	Message          string `json:"message"`
}

// --- Quota types (stretch) ---

type QuotaStatus struct {
	ProcessorID       string `json:"processor_id"`
	DailyQuota        int    `json:"daily_quota"`
	UsedToday         int    `json:"used_today"`
	Remaining         int    `json:"remaining"`
	IsAvailable       bool   `json:"is_available"`
	UnavailableReason string `json:"unavailable_reason,omitempty"`
}

type SimulationRequest struct {
	ProcessorOverrides map[string]ProcessorOverride `json:"processor_overrides"`
}

type ProcessorOverride struct {
	Available  *bool `json:"available,omitempty"`
	AtCapacity *bool `json:"at_capacity,omitempty"`
	QuotaUsed  *int  `json:"quota_used,omitempty"`
}

// --- Historical types (stretch) ---

type HistoricalAnalysis struct {
	TotalTransactions      int                `json:"total_transactions"`
	TotalActualCost        float64            `json:"total_actual_cost"`
	TotalSmartCost         float64            `json:"total_smart_cost"`
	TotalSavings           float64            `json:"total_savings"`
	AnnualProjection       float64            `json:"annual_projection"`
	MostExpensiveCorridors []CostCorridor     `json:"most_expensive_corridors"`
	HighestCostProcessors  []ProcessorCostRank `json:"highest_cost_processors"`
	ComplexRefundRules     []ComplexRuleNote  `json:"complex_refund_rules"`
	MonthlySavings         map[string]float64 `json:"monthly_savings"`
}

type CostCorridor struct {
	Country       Country       `json:"country"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	AvgCost       float64       `json:"avg_cost"`
	TotalCost     float64       `json:"total_cost"`
	Count         int           `json:"count"`
}

type ProcessorCostRank struct {
	ProcessorID string  `json:"processor_id"`
	TotalCost   float64 `json:"total_cost"`
	AvgCost     float64 `json:"avg_cost"`
	Count       int     `json:"count"`
}

type ComplexRuleNote struct {
	Rule        string `json:"rule"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

// --- API request wrappers ---

type SingleRefundRequest struct {
	Transaction Transaction `json:"transaction"`
}

type HistoricalRequest struct {
	Transactions []Transaction `json:"transactions"`
}
