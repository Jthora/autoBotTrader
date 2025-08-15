## 2025-08 Ephemeris and Packaging Decisions

- Ephemeris runtime is offline and local. Default provider reads an embedded precomputed series (60–90 days, 1–5min cadence). Alternate pure-Go provider (Meeus/VSOP87) is feature-flagged and will ship after accuracy validation.
- On-chain scope remains verification and governance: store weights/config + verifier pubkey; enforce cadence/bounds; accept signed inputs; emit events. No on-chain ephemeris computation.
- Packaging for demo: single Go binary serves UI and API; macOS `.app` wrapper launches the service and opens the browser automatically. No terminal required.
- Accuracy acceptance: median relative error ≤ 0.5%, P99 ≤ 1.5% vs Skyfield reference; peak timing drift ≤ 3 minutes over ≥14 days.

# Architecture & Strategy Decisions (ADR Log)

Format: ADR-<increment>: Title

## ADR-001: On-Chain Composite Computation

Context: Decide whether to push pre-computed composite or raw inputs.
Decision: Compute composite on-chain for transparency and verifiability.
Status: Accepted
Consequences: Slightly higher gas per update; removes trust in off-chain math.

## ADR-002: LegacyMap for Trades

Context: Storage of executed trades.
Decision: Use LegacyMap keyed by trade_id felt for sparsity & flexibility.
Status: Accepted
Consequences: Need external indexing for pagination.

## ADR-003: Single Admin Access Control (Phase 0-2)

Context: Simplicity vs. early decentralization.
Decision: Single admin variable with potential upgrade later.
Status: Accepted
Consequences: Centralized mutation risk early; documented as temporary.

## ADR-004: Go Preferred for Off-Chain Service

Context: Choice between Go & TypeScript.
Decision: Use Go for concurrency & alignment with future performance needs.
Status: Accepted
Consequences: Additional language in repo; TS fallback remains possible.

## ADR-005: Commit–Reveal Deferred

Context: Need to mitigate potential front‑running of prediction inputs.
Decision: Defer commit–reveal; rely on cooldown + batching first.
Status: Accepted (Deferred Implementation)
Consequences: MVP simpler; revisit if MEV becomes material.

## ADR-006: Event-First Trade Logging

Context: Persist trades vs events-only.
Decision: Use events only; omit on-chain trade map.
Status: Accepted
Consequences: Requires indexer; reduces storage/gas.
