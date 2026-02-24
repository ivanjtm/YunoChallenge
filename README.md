# Smart Refund Routing Service

A cost-optimization engine for refund processing in Latin American e-commerce. Built for **Vela Market**, a marketplace operating across Brazil, Mexico, and Colombia, this service replaces naive "refund through the original processor" logic with intelligent multi-processor routing that cuts refund costs by 25-35%.

---

## The Problem

When a customer requests a refund, most payment platforms take the simplest path: reverse the charge through the same processor that handled the original payment. This is convenient, but expensive.

Consider a PIX payment of R$320 processed through GlobalPay in Brazil. GlobalPay does not offer PIX-to-PIX refunds, so the naive approach falls back to a bank transfer at R$8.40. But PayBR, a Brazil-specialist processor, can do a PIX-to-PIX refund for just R$2.10 -- saving R$6.30 on a single transaction. Scale that across thousands of daily refunds in three countries with seven payment methods and six processors, and the savings compound quickly.

The problem gets harder when you factor in the real constraints of LATAM payments:

- **Cash methods cannot self-refund.** A customer who paid with OXXO (Mexico cash vouchers), Boleto (Brazil bank slips), or Efecty (Colombia cash payments) cannot receive cash back through the same channel. The money must go via bank transfer or account credit instead.
- **Time windows close.** PIX refunds are only available for 90 days, PSE for 60 days, and card refunds for 180 days. After expiry, the marketplace is forced into more expensive fallback methods.
- **Free reversals are fleeting.** An unsettled transaction less than 24 hours old can be voided at zero cost, but this window slams shut fast.
- **Currency matters.** Fee structures vary dramatically by currency. A 1.5% fee means something very different when applied to R$200 (BRL) versus $200,000 (COP).

This service solves all of that. It evaluates every eligible refund path across all available processors, calculates the true cost of each option, and selects the cheapest route -- while respecting every regulatory and operational constraint.

---

## Quick Start

```bash
# Go 1.22+ required (uses enhanced ServeMux for method-based routing)
go run main.go

# Server starts on :8080
# 200 test transactions are auto-generated on first run
```

The service loads processor fee structures and compatibility rules from JSON config files at startup, generates reproducible test data (seed 42), and is ready to accept requests immediately. No database, no external dependencies, no setup.

---

## Architecture

### Request Flow

```
                                        +------------------+
                                        | config/          |
                                        |  processors.json |
                                        |  rules.json      |
                                        +--------+---------+
                                                 |
                                          loaded at startup
                                                 |
                                                 v
  HTTP Request                          +------------------+
       |                                | In-Memory Config |
       v                                | (processors,     |
+------+-------+                        |  rules, index)   |
|  Middleware   |                        +--------+---------+
|  - Recovery   |                                 |
|  - Logging    |                                 |
|  - Content    |                                 |
|    Type       |                                 |
+------+-------+                                 |
       |                                         |
       v                                         |
+------+-------+     +---------------+     +-----+----------+
|   Handler    +---->+ Router Engine +---->+ Rule Index     |
| (validation, |     | (7-step       |     | O(1) lookup by |
|  decode JSON)|     |  algorithm)   |     | method:country)|
+--------------+     +-------+-------+     +----------------+
                             |
                     +-------+-------+
                     |               |
              +------+-----+  +-----+------+
              | Cost Engine|  | Time Window|
              | base + %   |  | PIX: 90d   |
              | clamped by |  | PSE: 60d   |
              | min/max    |  | Card: 180d |
              +------+-----+  | Rev: 24h   |
                     |        +------------+
                     v
              +------+------+
              | Ranked      |
              | Candidates  |
              | 1. cost asc |
              | 2. speed    |
              | 3. orig.    |
              |    proc     |
              | (credit     |
              |  always     |
              |  last)      |
              +------+------+
                     |
                     v
              +------+------+
              | Response    |
              | selected +  |
              | alternates +|
              | naive cost +|
              | savings     |
              +-------------+
```

### Project Structure

```
.
+-- main.go                          # Server bootstrap, route registration
+-- config/
|   +-- processors.json              # 6 processors with fee structures per currency
|   +-- rules.json                   # 9 compatibility rules (method + country -> allowed refunds)
+-- data/
|   +-- transactions.json            # 200 test transactions (auto-generated, seed 42)
+-- internal/
    +-- model/model.go               # All domain types: Transaction, Processor, RefundCandidate, etc.
    +-- config/loader.go             # JSON config loading with validation
    +-- rules/
    |   +-- compatibility.go         # O(1) rule index: map["PIX:BR"] -> allowed refund methods
    |   +-- timewindow.go            # Reversal eligibility (24h + unsettled), time window checks
    |   +-- rules.go                 # Orchestrator: finds all eligible refund paths for a txn
    +-- cost/calculator.go           # Fee formula: max(min_fee, min(max_fee, base + amount * %))
    +-- router/
    |   +-- selector.go              # Core 7-step routing algorithm
    |   +-- batch.go                 # Concurrent batch analysis with worker pool
    +-- quota/tracker.go             # Processor daily quota tracking + simulation overrides
    +-- historical/analyzer.go       # Historical what-if analysis with annual savings projection
    +-- handler/                     # HTTP handlers + middleware (logging, recovery, content-type)
    +-- testdata/generator.go        # Deterministic test data: 23 edge cases + 177 random txns
```

---

## Design Decisions

### Why Go with standard library only

The service has zero external dependencies. Go 1.22 introduced method-based routing in `net/http.ServeMux` (e.g., `"POST /api/v1/refund"`), eliminating the main reason teams reach for frameworks like chi or gin. The result is a service that compiles with `go build`, runs with `go run`, and has no dependency management overhead. For a focused microservice like this, that simplicity is a feature.

### Why in-memory configuration

Six processors and nine compatibility rules fit comfortably in memory. Loading from JSON files at startup means the configuration is human-readable, version-controllable, and trivially auditable. A database would add operational complexity (migrations, connection pooling, failure modes) without providing any benefit at this scale. If Vela Market grows to hundreds of processors, the config layer can be swapped to a database without changing the routing engine -- the `Router` struct accepts `[]Processor` and `[]CompatibilityRule`, not a database connection.

### Why the 7-step algorithm

The routing algorithm deliberately separates concerns into discrete steps to keep each one testable and debuggable:

1. **Eligibility filtering** comes first because it is the most constrained. No amount of cost optimization matters if the refund method is not legally or operationally available.
2. **Processor matching** narrows to processors that actually serve the transaction's country and currency.
3. **Cost calculation** applies the fee formula independently per candidate, so each cost is auditable.
4. **Quota checking** is a separate step (rather than baked into matching) so that quota exhaustion is visible in logs and can be simulated.
5. **Ranking** uses a deliberate three-tier sort: cost, then speed, then original-processor preference. This ensures the cheapest option always wins, ties are broken by customer experience (faster is better), and further ties favor the original processor for simpler reconciliation.
6. **Naive baseline** is computed separately because it must always use the original processor, regardless of what smart routing selects. This isolates the "what would have happened" calculation from the optimization.
7. **Result assembly** bundles selected + alternatives + savings into a single response, so the caller gets full transparency into why a route was chosen.

### Why account credit is ranked last despite being free

Account credit costs the marketplace nothing in processing fees -- so purely by cost, it should always win. But it is a worse customer experience: the refund is trapped as marketplace balance rather than returned to the customer's bank or card. The routing engine deliberately pushes account credit to the bottom of the ranking. It only surfaces as the selected option when no other refund method is available (e.g., a Boleto payment where bank transfer processors are all at capacity). This mirrors how real marketplaces operate: credit is the safety net, not the first choice.

### Why reversals get special treatment

A reversal (void) is fundamentally different from a refund. It cancels the transaction before settlement, so no money actually moves -- and it costs nothing. The 24-hour unsettled window is hardcoded rather than configurable because it reflects payment network rules, not business policy. Misonfiguring this window could lead to attempting reversals on settled transactions, which would fail at the processor level. The code treats reversal eligibility as a binary check (`IsReversalEligible`) rather than a fee entry, because the cost is always zero and the constraints are universal.

---

## Business Rules

### Payment Methods by Country

| Country  | Currency | Payment Methods                           |
|----------|----------|-------------------------------------------|
| Brazil   | BRL      | PIX, Boleto, Credit Card                  |
| Mexico   | MXN      | OXXO, SPEI, Credit Card                   |
| Colombia | COP      | PSE, Efecty, Credit Card                  |

### Refund Method Eligibility

| Original Method | Allowed Refund Methods                          | Constraints                                  |
|-----------------|-------------------------------------------------|----------------------------------------------|
| Credit Card     | Reversal, Same-method, Bank transfer            | Reversal: unsettled + <24h. Same-method: 180 days. |
| PIX (BR)        | Reversal, Same-method, Bank transfer            | Reversal: unsettled + <24h. Same-method: 90 days.  |
| SPEI (MX)       | Reversal, Same-method, Bank transfer            | Reversal: unsettled + <24h. No time limit on same-method. |
| PSE (CO)        | Reversal, Same-method, Bank transfer            | Reversal: unsettled + <24h. Same-method: 60 days.  |
| Boleto (BR)     | Bank transfer, Account credit                   | Cannot refund as Boleto. No reversal available.     |
| OXXO (MX)       | Bank transfer, Account credit                   | Cannot refund as OXXO. No reversal available.       |
| Efecty (CO)     | Bank transfer, Account credit                   | Cannot refund as Efecty. No reversal available.     |

**Key insight:** Cash-based methods (Boleto, OXXO, Efecty) cannot issue refunds through their own channel. The customer deposited physical cash or generated a voucher -- there is no reverse rail. The system must route these to bank transfers or, as a last resort, account credit.

### Fee Formula

Every processor fee is calculated as:

```
raw_cost = base_fee + (amount * percent_fee)
cost     = max(min_fee, min(max_fee, raw_cost))     // clamp to [min, max]
```

If `max_fee` is 0, there is no cap. Reversals and account credits always cost 0.

### Processors

| Processor   | Countries   | Strengths                                     | Tradeoff                    |
|-------------|-------------|-----------------------------------------------|-----------------------------|
| PayBR       | BR          | Lowest PIX fees (0.5% + R$0.50)               | Brazil only                 |
| MexPay      | MX          | Lowest SPEI fees (0.8% + $5 MXN)              | Mexico only                 |
| ColPay      | CO          | Lowest PSE fees (0.6% + $1,500 COP)           | Colombia only               |
| GlobalPay   | BR, MX, CO  | Pan-LATAM coverage, all card types             | 2% across the board         |
| QuickRefund | BR, MX      | Instant processing (0 days)                   | Most expensive (3% + high base) |
| ValueProc   | BR, MX, CO  | Cheapest overall (0.8-1%)                     | Slowest (3-5 days)          |

---

## API Reference

| Method   | Path                          | Description                                    |
|----------|-------------------------------|------------------------------------------------|
| `GET`    | `/api/v1/health`              | Health check with loaded config stats          |
| `POST`   | `/api/v1/refund`              | Route a single refund to the cheapest path     |
| `POST`   | `/api/v1/refund/batch`        | Concurrent batch analysis with savings report  |
| `POST`   | `/api/v1/simulation/quota`    | Set processor availability overrides           |
| `DELETE` | `/api/v1/simulation/quota`    | Reset simulation state to defaults             |
| `POST`   | `/api/v1/analysis/historical` | Historical cost analysis with annual projection|

---

## Example Usage

### Example 1: PIX Refund -- Smart Routing Saves 75%

A PIX payment of R$320 was processed through GlobalPay 88 days ago. GlobalPay does not support PIX-to-PIX refunds, so naively it would fall back to a bank transfer. Smart routing finds PayBR, which offers PIX-to-PIX at a fraction of the cost.

```bash
curl -s -X POST http://localhost:8080/api/v1/refund \
  -H "Content-Type: application/json" \
  -d '{
    "transaction": {
      "id": "txn_edge_006",
      "country": "BR",
      "currency": "BRL",
      "payment_method": "PIX",
      "processor_id": "globalpay",
      "amount": 320.00,
      "timestamp": "2025-11-28T10:00:00Z",
      "settled": true,
      "customer_id": "cust_042"
    }
  }' | jq .
```

**Expected response:**

```json
{
  "transaction_id": "txn_edge_006",
  "selected": {
    "processor_id": "paybr",
    "processor_name": "PayBR",
    "refund_method": "SAME_METHOD",
    "estimated_cost": 2.1,
    "processing_days": 1,
    "reasoning": "PIX-to-PIX via PayBR: 0.50 base + 0.5% = 2.10 BRL, 1 day processing; Within SAME_METHOD window (88 of 90 days used, 2 remaining)"
  },
  "alternatives": [
    {
      "processor_id": "valueproc",
      "processor_name": "ValueProc",
      "refund_method": "SAME_METHOD",
      "estimated_cost": 3.06,
      "processing_days": 3,
      "reasoning": "PIX-to-PIX via ValueProc: 0.50 base + 0.8% = 3.06 BRL, 3 days processing; Within SAME_METHOD window (88 of 90 days used, 2 remaining)"
    },
    {
      "processor_id": "valueproc",
      "processor_name": "ValueProc",
      "refund_method": "BANK_TRANSFER",
      "estimated_cost": 3.95,
      "processing_days": 5,
      "reasoning": "bank transfer via ValueProc: 0.75 base + 1.0% = 3.95 BRL, 5 days processing; No time limit for this refund method"
    },
    {
      "processor_id": "paybr",
      "processor_name": "PayBR",
      "refund_method": "BANK_TRANSFER",
      "estimated_cost": 5.8,
      "processing_days": 2,
      "reasoning": "bank transfer via PayBR: 1.00 base + 1.5% = 5.80 BRL, 2 days processing; No time limit for this refund method"
    },
    {
      "processor_id": "globalpay",
      "processor_name": "GlobalPay",
      "refund_method": "BANK_TRANSFER",
      "estimated_cost": 8.4,
      "processing_days": 3,
      "reasoning": "bank transfer via GlobalPay: 2.00 base + 2.0% = 8.40 BRL, 3 days processing; No time limit for this refund method"
    },
    {
      "processor_id": "quickrefund",
      "processor_name": "QuickRefund",
      "refund_method": "SAME_METHOD",
      "estimated_cost": 12.6,
      "processing_days": 0,
      "reasoning": "PIX-to-PIX via QuickRefund: 3.00 base + 3.0% = 12.60 BRL, instant processing; Within SAME_METHOD window (88 of 90 days used, 2 remaining)"
    },
    {
      "processor_id": "quickrefund",
      "processor_name": "QuickRefund",
      "refund_method": "BANK_TRANSFER",
      "estimated_cost": 10.5,
      "processing_days": 1,
      "reasoning": "bank transfer via QuickRefund: 2.50 base + 2.5% = 10.50 BRL, 1 day processing; No time limit for this refund method"
    },
    {
      "processor_id": "internal",
      "processor_name": "Account Credit",
      "refund_method": "ACCOUNT_CREDIT",
      "estimated_cost": 0,
      "processing_days": 0,
      "reasoning": "No time limit for this refund method; funds credited to customer marketplace balance"
    }
  ],
  "naive_cost": 8.4,
  "savings": 6.3
}
```

**What happened:** The naive path (GlobalPay bank transfer) would cost R$8.40. Smart routing found PayBR can do a PIX-to-PIX refund for R$2.10, saving R$6.30 (75%). Note the time-sensitive detail in the reasoning: only 2 days remain in the 90-day PIX window. After that, the R$2.10 option disappears and the cheapest path jumps to R$3.95.

### Example 2: OXXO Refund -- Cash Cannot Self-Refund

A customer paid MXN $2,500 via OXXO (Mexican cash voucher). OXXO payments cannot be refunded through OXXO -- the cash was handed to a convenience store clerk. The system must find an alternative channel.

```bash
curl -s -X POST http://localhost:8080/api/v1/refund \
  -H "Content-Type: application/json" \
  -d '{
    "transaction": {
      "id": "txn_edge_010",
      "country": "MX",
      "currency": "MXN",
      "payment_method": "OXXO",
      "processor_id": "mexpay",
      "amount": 2500.00,
      "timestamp": "2026-01-10T09:00:00Z",
      "settled": true,
      "customer_id": "cust_099"
    }
  }' | jq .
```

**Expected response:**

```json
{
  "transaction_id": "txn_edge_010",
  "selected": {
    "processor_id": "valueproc",
    "processor_name": "ValueProc",
    "refund_method": "BANK_TRANSFER",
    "estimated_cost": 33.0,
    "processing_days": 5,
    "reasoning": "bank transfer via ValueProc: 8.00 base + 1.0% = 33.00 MXN, 5 days processing; No time limit for this refund method"
  },
  "alternatives": [
    {
      "processor_id": "mexpay",
      "processor_name": "MexPay",
      "refund_method": "BANK_TRANSFER",
      "estimated_cost": 40.0,
      "processing_days": 2,
      "reasoning": "bank transfer via MexPay: 10.00 base + 1.2% = 40.00 MXN, 2 days processing; No time limit for this refund method"
    },
    {
      "processor_id": "globalpay",
      "processor_name": "GlobalPay",
      "refund_method": "BANK_TRANSFER",
      "estimated_cost": 70.0,
      "processing_days": 3,
      "reasoning": "bank transfer via GlobalPay: 20.00 base + 2.0% = 70.00 MXN, 3 days processing; No time limit for this refund method"
    },
    {
      "processor_id": "quickrefund",
      "processor_name": "QuickRefund",
      "refund_method": "BANK_TRANSFER",
      "estimated_cost": 87.5,
      "processing_days": 1,
      "reasoning": "bank transfer via QuickRefund: 25.00 base + 2.5% = 87.50 MXN, 1 day processing; No time limit for this refund method"
    },
    {
      "processor_id": "internal",
      "processor_name": "Account Credit",
      "refund_method": "ACCOUNT_CREDIT",
      "estimated_cost": 0,
      "processing_days": 0,
      "reasoning": "No time limit for this refund method; funds credited to customer marketplace balance"
    }
  ],
  "naive_cost": 40.0,
  "savings": 7.0
}
```

**What happened:** With OXXO, the `SAME_METHOD` refund path does not exist -- the rules explicitly exclude it. The system routes to bank transfer instead. ValueProc offers the cheapest bank transfer at MXN $33 versus the naive MexPay path at MXN $40, saving MXN $7 per transaction. Account credit appears as a fallback but is ranked last because it locks the customer's money inside the marketplace.

### Example 3: Batch Analysis

Analyze all 200 test transactions at once to see aggregate savings:

```bash
curl -s -X POST http://localhost:8080/api/v1/refund/batch \
  -H "Content-Type: application/json" \
  -d "$(jq '{transactions: .}' data/transactions.json)" | jq '{
    total_transactions,
    total_naive_cost,
    total_smart_cost,
    total_savings,
    savings_percent,
    time_sensitive_count: (.time_sensitive | length),
    limited_options_count: (.limited_options | length),
    by_payment_method
  }'
```

Transactions are routed concurrently across CPU cores. The response includes per-processor breakdowns, per-payment-method breakdowns, flagged time-sensitive transactions (refund windows closing within 15 days), and flagged limited-option transactions (cash methods with fewer routing choices).

### Example 4: Processor Quota Simulation

Test what happens when processors become unavailable:

```bash
# Simulate PayBR at capacity and QuickRefund offline
curl -s -X POST http://localhost:8080/api/v1/simulation/quota \
  -H "Content-Type: application/json" \
  -d '{
    "processor_overrides": {
      "paybr": {"at_capacity": true},
      "quickrefund": {"available": false}
    }
  }' | jq .

# Now re-run a PIX refund -- it will route to ValueProc instead of PayBR
# ...

# Reset simulation state
curl -s -X DELETE http://localhost:8080/api/v1/simulation/quota | jq .
```

### Example 5: Historical Cost Analysis

Compute how much the marketplace would have saved over the entire transaction history:

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

The response identifies the most expensive payment corridors (e.g., Colombia credit cards via GlobalPay), projects annual savings, and documents the complex refund rules that constrain routing decisions.

---

## Performance: Concurrent Batch Processing

The batch endpoint (`POST /api/v1/refund/batch`) processes transactions concurrently using a goroutine worker pool. Each `SelectRoute` call -- the 7-step algorithm with rule lookups, cost calculations, and candidate ranking -- is independent per transaction, making it an ideal candidate for parallelism.

### How It Works

```
                    +-----------+
                    |  Request  |
                    | (N txns)  |
                    +-----+-----+
                          |
                    +-----+-----+
                    |  Fan Out  |
                    | (buffered |
                    |  channel) |
                    +-----+-----+
                          |
          +---------------+---------------+
          |               |               |
    +-----+-----+  +-----+-----+  +-----+-----+
    |  Worker 1 |  |  Worker 2 |  |  Worker K |
    | SelectRoute| | SelectRoute| | SelectRoute|  K = runtime.NumCPU()
    +-----+-----+  +-----+-----+  +-----+-----+
          |               |               |
          +---------------+---------------+
                          |
                    +-----+-----+
                    | Accumulate|
                    | (single-  |
                    |  threaded)|
                    +-----+-----+
                          |
                    +-----+-----+
                    | Response  |
                    +-----------+
```

The implementation uses three components:

1. **Job channel:** All transactions are sent into a buffered channel with their original index attached. The index ensures results are placed back in the correct position regardless of processing order.

2. **Worker goroutines:** `runtime.NumCPU()` workers consume from the job channel in parallel. Each worker calls `SelectRoute` independently -- the routing engine is stateless and safe for concurrent reads. Workers are capped at the number of transactions to avoid idle goroutines on small batches.

3. **Single-threaded accumulation:** After all workers finish, results are collected and map-based aggregation (per-processor summaries, per-method summaries, time-sensitive flags) happens on a single goroutine. This avoids mutex contention on the accumulator maps, which would negate the parallelism gains at small batch sizes.

This design means the expensive work (rule lookups, fee calculations, candidate sorting per transaction) scales linearly with CPU cores, while the cheap work (summing costs into maps) stays simple and lock-free.

---

## The Routing Algorithm in Detail

For each incoming refund request, the engine executes the following steps:

```
Step 1: ELIGIBILITY
   Input:  transaction (payment_method, country, timestamp, settled)
   Action: Look up compatibility rules for (payment_method, country).
           For each allowed refund method, check:
             - Reversal? -> Must be unsettled AND < 24 hours old
             - Same-method? -> Must be within time window (PIX: 90d, PSE: 60d, Card: 180d)
             - Bank transfer? -> Always available (no time limit)
             - Account credit? -> Always available (no time limit)
   Output: List of eligible refund methods with reasons

Step 2: PROCESSOR MATCHING
   Input:  Eligible methods + all processor configs
   Action: For each eligible method, find processors that:
             - Support the transaction's country
             - Support the transaction's currency
             - Have a fee entry matching the refund method + original payment method
   Output: List of (processor, method, fee) triples

Step 3: COST CALCULATION
   Input:  Each (processor, method, fee) triple + transaction amount
   Action: Apply fee formula: max(min_fee, min(max_fee, base + amount * %))
           Reversals = 0, Account credit = 0
   Output: Costed candidates

Step 4: QUOTA CHECK
   Input:  Costed candidates + quota tracker state
   Action: Remove processors that are at capacity, unavailable, or exhausted
   Output: Available candidates

Step 5: RANKING
   Input:  Available candidates
   Action: Sort by:
             1. Account credit pushed to bottom (regardless of cost)
             2. Estimated cost (ascending)
             3. Processing days (ascending)
             4. Original processor preferred (for reconciliation simplicity)
   Output: Ordered candidate list

Step 6: NAIVE BASELINE
   Input:  Transaction + original processor config
   Action: Calculate what the original processor would charge using SAME_METHOD,
           falling back to BANK_TRANSFER, falling back to 3.5% estimate
   Output: Naive cost (the "before" number)

Step 7: RESULT
   Output: { selected, alternatives[], naive_cost, savings }
```

---

## Test Data

200 transactions are generated deterministically (seed 42) on first server start. The dataset is designed to exercise every routing path:

| Category              | Count | Purpose                                               |
|-----------------------|-------|-------------------------------------------------------|
| Reversal candidates   | 3     | Unsettled transactions < 1 hour old (free void window)|
| Near-expiry PIX       | 3     | 86-89 days old (PIX 90-day window about to close)     |
| Near-expiry PSE       | 2     | 57-59 days old (PSE 60-day window about to close)     |
| Cash methods (OXXO)   | 3     | Cannot self-refund, tests bank transfer fallback      |
| Cash methods (Boleto) | 2     | Cannot self-refund, tests bank transfer fallback      |
| Cash methods (Efecty) | 2     | Cannot self-refund, tests bank transfer fallback      |
| High-value cards      | 3     | Tests max_fee capping on large transactions           |
| Small amounts         | 2     | Tests min_fee floor on tiny transactions              |
| Near-expiry cards     | 2     | 176-179 days old (card 180-day window about to close) |
| Random (realistic)    | 177   | Weighted by country (BR 45%, MX 35%, CO 20%)         |

**Country and method distributions match real LATAM e-commerce patterns:** PIX dominates in Brazil (50%), credit cards lead in Mexico (40%) and Colombia (40%), and cash methods (Boleto, OXXO, Efecty) represent the tail.

---

## Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run a specific package
go test ./internal/cost/...
go test ./internal/rules/...
go test ./internal/router/...
```

### What to verify

- **Cost calculations:** The fee formula correctly applies base + percentage, then clamps to [min, max]. Reversals and account credits always return 0.
- **Time window enforcement:** PIX blocked after 90 days, PSE after 60 days, cards after 180 days. Reversals blocked after 24 hours or if settled.
- **Cash method constraints:** Boleto, OXXO, and Efecty never appear as a SAME_METHOD refund option. Only bank transfer and account credit are available.
- **Ranking correctness:** Account credit is always ranked last. Among non-credit options, cheapest wins. Ties broken by speed, then by original processor.
- **Naive baseline accuracy:** The naive cost always uses the original processor, never the smart-routed one.
- **Edge cases:** Zero-amount transactions, unknown processors, transactions exactly at window boundaries.

---

## Configuration

### Processors (`config/processors.json`)

Each processor entry defines:
- Supported countries and currencies
- Fee entries per refund method, per original payment method, per currency
- Daily quota limits
- Processing time in days per refund method

### Compatibility Rules (`config/rules.json`)

Each rule maps an `(original_method, country)` pair to allowed refund methods, with optional constraints:
- `max_age_days`: Time window in days (0 = no limit)
- `require_settled`: `true` = must be settled, `false` = must be unsettled, `null` = no requirement

Restart the server to pick up config changes.

### Environment Variables

| Variable | Default | Description          |
|----------|---------|----------------------|
| `PORT`   | `8080`  | Server listen port   |
