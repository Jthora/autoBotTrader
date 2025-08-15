## Update: Off-chain Ephemeris and System Boundaries (Aug 2025)

- Ephemeris will NOT be computed on-chain. On-chain scope remains governance (weights/config), cadence/bounds validation, signature verification, and storage of canonical inputs + composite.
- Off-chain compute lives in the Go service. Runtime is fully local and offline for ephemeris: default provider reads a precomputed, embedded table (60–90 days, 1–5 min cadence). Alternate provider (feature-flagged) uses pure-Go Meeus/VSOP87 after validation.
- The Go service also serves the React UI (static assets embedded), exposing a simple localhost experience and minimizing moving parts.
- The frontend may visualize client-side ephemeris, but authoritative values come from the Go service to ensure determinism.
- Cryptographic signing: Go signs submitted inputs; Cairo stores verifier pubkey and checks signature + cadence before accepting updates.

# System Architecture

## High-Level Components

1. Cairo Smart Contract (core, deterministic trading gate)
2. Off-Chain Data Service (Go) – fetch + normalize + push
3. Optional TypeScript worker (fallback or client-side reprocessing)
4. React/TypeScript Frontend (monitor + control)
5. Future ML Pipeline (Python PPO) + oracle bridge
6. Deployment & Tooling (Scarb + Starkli + Katana devnet, CI)
7. Adapters Layer (ISuperDexAdapter placeholder)
8. Normalization & Versioning Artifacts (constants + formula docs)

## Data Flow Diagram (Conceptual)

```
External APIs (Astro, Tides) ---> Go Service ---> Starknet Call: set_prediction_inputs()
                                                 |-> Cairo Contract (stores latest, computes composite score)
Frontend <--- starknet.js (reads state + events)  |-> execute_trade() (keeper/user trigger)
Future Oracle/ML ---> set_ml_score() -------------^
```

## Cairo Contract Modules (Logical)

- Storage Layout (evolving spec — see 02_contract_spec)
  - admin: felt252
  - pusher_role: felt252 (authorized prediction submitter)
  - ml_oracle: felt252 (authorized ML score submitter)
  - execution_threshold: u32
  - cooldown_seconds: u32 (spam mitigation)
  - last_input_timestamp: u64
  - formula_version: u32 (starts at 1)
  - normalization_version: u32
  - weights: astrology_w, gravity_w, ml_w (u32, sum > 0, ml_w may be 0)
  - astrology_score, gravity_score, ml_score, ml_score_timestamp
  - composite_score
  - (Optional) trades map OR event-only (configurable compile-time feature)
  - reserved future slots for adapter integration
- Events
- PredictionUpdated(astrology, gravity, ml, composite, formula_version)
- TradeExecuted(trade_id, direction, amount, score, timestamp, improved_price_bps?)
- WeightsUpdated(astrology_w, gravity_w, ml_w)
- RolesUpdated(pusher_role, ml_oracle)
- CooldownUpdated(cooldown_seconds)
- MLScoreUpdated(ml_score, ml_score_timestamp, ml_model_version)
- Entry Points
  - set_prediction_inputs(a_score, g_score)
  - set_ml_score(ml_score)
  - update_weights(a_w, g_w, ml_w)
  - execute_trade(trade_id, amount, direction)
  - get_state() view
  - get_trade(trade_id) view

## Composite Score Formula (Versioned)

```
Inputs normalized off-chain deterministically (integer pipeline) to 0–100 (or 0–10_000 basis points internally) referencing a published constants table.
Formula v1: Composite = floor((a*aw + g*gw + ml*mw) / (aw + gw + mw)).
Formula changes increment formula_version; both on-chain and off-chain calculations must match golden vectors.
```

## Off-Chain Go Service

- HTTP Server Endpoints
  - GET /astrology → raw + normalized
  - GET /gravimetrics → raw + normalized
  - GET /predict → aggregated composite preview (mirrors formula) + (optional) contract push flag
  - POST /push → pushes current astro + gravity to contract
- Internals
  - Providers: interface { Fetch(ctx) (RawData, error) }
  - Parallel fetch with context.WithTimeout
  - Normalization strategies versioned (v1) stored in code constant
  - Starknet RPC client wrapper (invoke function call builder)

## Frontend Components

- hooks/useStarknetContract.ts
- components/ScorePanel.tsx
- components/TradesTable.tsx
- components/WeightsForm.tsx
- components/ExecuteButton.tsx
- state management (lightweight: React context or Zustand)

## ML/PPO (Future)

- Python script trains offline -> exports score
- Oracle path (not implemented): Off-chain signer posts ml_score periodically
- Contract neutrality: ml_weight=0 means ignored

## Security & Trust Boundary

- Roles: admin (governance), pusher_role (signal ingestion), ml_oracle (ML scores).
- Cooldown & optional future commit–reveal (deferred) to mitigate front‑running & spam.
- Replay prevention: trade_id uniqueness; optional hash preimage commit path (ADR deferred).
- Input validation strict: scores <= 100, weights sum > 0, oracle signatures (future) verified off-chain before submission or on-chain when feasible.

## Upgrade/Extensibility

- Plan for potential UDC (Universal Deployer) / separate proxy later
- Keep storage struct contiguous & documented for upgrade migration

## Indexing Strategy

- Event-first (no heavy trade storage) for scalability.
- Subgraph / indexer schema: prediction events keyed by block timestamp, formula_version, normalization_version.
- Gas snapshot + composite drift metrics derivable off-chain.

## Observability

- Structured JSON logs (off-chain) including correlation_id & formula_version.
- Metrics endpoints: /metrics (Prometheus) for provider latency, push success rate, cooldown rejections.
- Nightly stagnation + gas/coverage automation integrated with progress system.

---

See `02_contract_spec.md` for fine-grained contract design.
