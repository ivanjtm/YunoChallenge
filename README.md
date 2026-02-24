# Smart Refund Routing Service

A backend service that implements intelligent refund routing for Vela Market, a LATAM e-commerce marketplace operating in Brazil, Mexico, and Colombia.

Instead of naively refunding through the original payment processor, this service evaluates all eligible refund routes and selects the cheapest option — saving an estimated 25-35% on refund processing fees.

## Quick Start

```bash
# Run the service (Go 1.22+ required)
go run main.go

# Server starts on :8080
# Test data (200 transactions) is auto-generated on first run
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/health` | Health check with config stats |
| `POST` | `/api/v1/refund` | Route a single refund optimally |
| `POST` | `/api/v1/refund/batch` | Batch analysis with savings report |
| `POST` | `/api/v1/simulation/quota` | Set processor availability overrides |
| `DELETE` | `/api/v1/simulation/quota` | Reset simulation to defaults |
| `POST` | `/api/v1/analysis/historical` | Historical cost analysis |

## Example Requests

### Single Refund Routing

```bash
curl -s -X POST http://localhost:8080/api/v1/refund \
  -H "Content-Type: application/json" \
  -d '{
    "transaction": {
      "id": "txn_test_001",
      "country": "BR",
      "currency": "BRL",
      "payment_method": "PIX",
      "processor_id": "paybr",
      "amount": 250.00,
      "timestamp": "2026-01-15T14:30:00Z",
      "settled": true,
      "customer_id": "cust_001"
    }
  }' | jq .
```

Response shows the selected route, alternatives, naive cost, and savings:

```json
{
  "transaction_id": "txn_test_001",
  "selected": {
    "processor_id": "valueproc",
    "processor_name": "ValueProc",
    "refund_method": "SAME_METHOD",
    "estimated_cost": 3.00,
    "processing_days": 3,
    "reasoning": "PIX-to-PIX via ValueProc: 1.00 base + 0.8% = 3.00 BRL, 3 days processing; ..."
  },
  "alternatives": [...],
  "naive_cost": 5.25,
  "savings": 2.25
}
```

### Batch Analysis (use pre-generated test data)

```bash
# Use the auto-generated test transactions
curl -s -X POST http://localhost:8080/api/v1/refund/batch \
  -H "Content-Type: application/json" \
  -d @data/transactions.json | jq '{
    total_transactions,
    total_naive_cost,
    total_smart_cost,
    total_savings,
    savings_percent,
    time_sensitive_count: (.time_sensitive | length),
    limited_options_count: (.limited_options | length)
  }'
```

Note: the batch endpoint expects `{"transactions": [...]}`. The generated file is a raw array, so wrap it:

```bash
curl -s -X POST http://localhost:8080/api/v1/refund/batch \
  -H "Content-Type: application/json" \
  -d "$(jq '{transactions: .}' data/transactions.json)" | jq '{
    total_transactions,
    total_naive_cost,
    total_smart_cost,
    total_savings,
    savings_percent
  }'
```

### Historical Analysis

```bash
curl -s -X POST http://localhost:8080/api/v1/analysis/historical \
  -H "Content-Type: application/json" \
  -d "$(jq '{transactions: .}' data/transactions.json)" | jq '{
    total_transactions,
    total_actual_cost,
    total_smart_cost,
    total_savings,
    annual_projection,
    top_corridors: .most_expensive_corridors[:3],
    complex_rules: [.complex_refund_rules[].rule]
  }'
```

### Processor Quota Simulation

```bash
# Mark PayBR as at capacity
curl -s -X POST http://localhost:8080/api/v1/simulation/quota \
  -H "Content-Type: application/json" \
  -d '{
    "processor_overrides": {
      "paybr": {"at_capacity": true},
      "quickrefund": {"available": false}
    }
  }' | jq .

# Reset simulation
curl -s -X DELETE http://localhost:8080/api/v1/simulation/quota | jq .
```

### Health Check

```bash
curl -s http://localhost:8080/api/v1/health | jq .
```

## Architecture

### Project Structure

```
.
├── main.go                          # Server entry point
├── config/
│   ├── processors.json              # 6 processor fee structures
│   └── rules.json                   # 9 compatibility rules
├── data/
│   └── transactions.json            # 200 test transactions (auto-generated)
└── internal/
    ├── model/model.go               # All domain structs and enums
    ├── config/loader.go             # JSON config loading with validation
    ├── rules/
    │   ├── compatibility.go         # O(1) method compatibility lookup
    │   ├── timewindow.go            # Time-based eligibility checks
    │   └── rules.go                 # Orchestrator: eligible refund paths
    ├── cost/calculator.go           # Fee calculation engine
    ├── router/
    │   ├── selector.go              # Core routing algorithm
    │   └── batch.go                 # Batch analysis with savings report
    ├── quota/tracker.go             # Processor quota simulation
    ├── historical/analyzer.go       # Historical cost analysis
    ├── handler/                     # HTTP handlers + middleware
    └── testdata/generator.go        # Reproducible test data generator
```

### Routing Algorithm

The core routing engine follows a 7-step process:

1. **Find eligible refund methods** — Look up compatibility rules for the transaction's `(payment_method, country)` pair. Check time windows (PIX 90d, PSE 60d, Card 180d). Check reversal eligibility (unsettled + <24h).

2. **Match processors** — For each eligible method, find processors supporting the transaction's country, currency, and payment method.

3. **Calculate cost** — Apply fee formula: `cost = max(min_fee, min(max_fee, base_fee + amount × percent_fee))`. Reversals are always free.

4. **Check quotas** — (Stretch) Filter out processors at capacity or unavailable.

5. **Rank candidates** — Sort by cost (asc), then processing time (asc), then prefer original processor.

6. **Calculate naive baseline** — What it would cost through the original processor.

7. **Return result** — Selected route + alternatives + savings breakdown + reasoning.

### Domain Rules

**Payment Methods by Country:**
- Brazil: PIX, Boleto, Credit Card (BRL)
- Mexico: OXXO, SPEI, Credit Card (MXN)
- Colombia: PSE, Efecty, Credit Card (COP)

**Key Constraints:**
- Cash methods (OXXO, Boleto, Efecty) **cannot** be refunded the same way — must use bank transfer or account credit
- PIX-to-PIX refunds only within **90 days**
- PSE-to-PSE refunds only within **60 days**
- Card refunds within **180 days**
- Free reversals (voids) only for **unsettled transactions < 24 hours old**
- Bank transfer is always available as a (more expensive) fallback

### Processors

| Processor | Countries | Specialty |
|-----------|-----------|-----------|
| PayBR | BR | Low PIX fees (0.5%) |
| MexPay | MX | Low SPEI fees (0.8%) |
| ColPay | CO | Low PSE fees (0.6%) |
| GlobalPay | BR, MX, CO | Moderate fees, broad coverage |
| QuickRefund | BR, MX | Fast (instant) but expensive (3%) |
| ValueProc | BR, MX, CO | Cheapest (0.8%) but slowest (3-5 days) |

## Test Data

200 transactions are auto-generated with a fixed seed (42) for reproducibility. The dataset includes:

- **Country distribution**: BR 45%, MX 35%, CO 20%
- **Payment methods**: Weighted realistically per country
- **23 edge cases**: Reversal candidates, time-sensitive windows, cash methods, very large/small amounts
- **6-month span**: Transactions from August 2025 to February 2026

## Design Decisions

- **Go standard library only** — Zero external dependencies. Uses Go 1.22+ enhanced `ServeMux` for method-based routing.
- **In-memory config** — JSON files loaded at startup. No database needed for this scale.
- **Currency-aware fees** — Multi-country processors have separate fee entries per currency to handle magnitude differences (BRL vs COP).
- **Reversal logic hardcoded** — The 24h + unsettled rule is universal, not configurable, to prevent misconfiguration.
- **Account credit as last resort** — Always available at zero cost but ranked below bank transfer for better customer experience.
- **Fixed seed test data** — Every developer gets identical data for reproducible results.

## Running Tests

```bash
go test ./...
```

## Configuration

Modify `config/processors.json` to add processors or change fees. Modify `config/rules.json` to update compatibility rules. Restart the server to pick up changes.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server listen port |
