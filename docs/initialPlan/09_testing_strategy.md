## Aug 2025 Update — Ephemeris Accuracy & Offline Testing

See also: `docs/testing/TEST_PLAN_EXPANDED.md` for granular per-component subtest specifications and implementation order.

- Accuracy Harness: Compare the runtime ephemeris provider (file/algo) against a Skyfield reference over ≥14 days at 1s cadence (downsampled checks for longer ranges).
  - Metrics: relative error (median ≤ 0.5%, P99 ≤ 1.5%), peak/zero-crossing timing drift ≤ 3 minutes.
  - Artifacts: metrics.json and optional plots committed under `scripts/ephem/validation/`.
- Offline Mode: Block outbound network and verify /health, /predict, /gravimetrics continue to work; only /push requires RPC when enabled.
- Performance: P95 < 5ms for provider-backed endpoints; startup < 300ms to first response.
- Packaging: macOS `.app` launches and opens the browser; auto-port selection if default busy; clean shutdown verified.
- Binary Format: Validate header/magic, field sizes, and record count; fuzz offsets; test both `embed` and `mmap` loaders.
- Multi-Resolution: Verify 1s/60s coherence; upsampling interpolants remain within error thresholds.

# Testing & Quality Strategy

## Layers

1. Cairo Contract Unit Tests
2. Integration (Off-chain service -> local devnet contract)
3. Frontend Component & Integration Tests
4. End-to-End (Simulated flow: fetch -> push -> execute -> UI reflect)
5. Static & Linting

## Contract Tests (Scarb + Starkli / snforge)

- Focus: state transitions, role enforcement, cooldown, gas invariants.
- Suite Groups:
  - Initialization
  - Input Updates
  - Weight & Threshold Management
  - Trade Execution (event path)
  - Cooldown & Spam Mitigation
  - Roles & Access (admin/pusher/ml_oracle)
  - Versioning (formula_version / normalization_version)
  - ML Score Handling

### Concrete Cases

- init sets defaults; owner/admin set; initial weights sum and threshold sane
- set_prediction updates raw inputs; composite recomputed; event emitted
- set_weights reweighs composite; rejects invalid totals (sum=0) or out-of-bounds
- cooldown: immediate second call reverts; after cooldown passes
- access control: non-admin reverts on updates; ML-only setter gated
- threshold gating: execute_trade below threshold reverts; above threshold passes
- ml_score: when ml_weight > 0, ml score changes composite

### Edge Cases

- weights at extremes (100/0 splits); inputs at 0 and 100 bounds
- time-dependent fields (cadence checks) when added; signature/bounds verification
- gas snapshot deltas tracked across PRs (<= 5% without ADR)

## Off-Chain Go Tests

- Providers (mock/timeouts)
- Normalization logic
- Chain client (mocked transport)
- HTTP handlers (httptest)
- Race detector

### Concrete Cases

- normalize: astrology/grav bounds, midpoints, clamping
- handlers: /astrology, /gravimetrics, /predict, /push happy paths; Content-Type assertions
- handlers: provider/chain errors → 503 or dry-run; JSON error bodies stable
- chain client: respects context; returns errors; fallback to dry-run in handler

### Edge Cases

- EPHEM_MODE=file with missing/invalid table → mock fallback logged
- concurrent requests under -race; no data races
- graceful shutdown: signal triggers shutdown; provider Close() called

## Frontend Tests

- Render ScorePanel with mocked provider
- Execute trade button triggers contract call (mock)
- Weights form disabled for non-admin

### Concrete Cases

- displays scores, composite, and grav meta (mode/dataset_id/stale)
- network failure shows friendly state; retry path
- push button shows DRYRUN when RPC env absent

## End-to-End (Phase 2)

- Spin up local Starknet devnet
- Deploy contract
- Run Go service pointing to devnet
- Script: call /push then execute_trade via contract
- Assert events accessible & UI displays results (headless Playwright / Cypress)

## Tooling & Commands (Proposed)

```
# Contracts
yarn test:contracts OR scarb test

# Go
go test ./... -race -cover

# Frontend
npm run test

# E2E (future)
npm run e2e
```

## CI Pipeline Stages

1. Lint (Go vet, golangci-lint; ESLint; Cairo fmt check)
2. Unit Tests (contracts, Go, frontend)
3. Build Artifacts (Go binary, frontend dist)
4. (Optional) E2E on ephemeral devnet

## Metrics

- Contract: ≥90% critical logic + gas snapshot tracked (baseline + delta <= +5% allowed without ADR)
- Go: ≥80% logic; golden vector normalization tests pass
- Frontend: key components >70%; accessibility checks (a11y smoke)

## Non-Functional Tests (Later)

- Load test /predict endpoint
- Gas snapshot comparisons per commit (fail CI if regression > threshold)
- Stagnation detection (already via workflow) ensures progress cadence

---

Deployment procedures in `10_deployment_runbook.md`.

---

## Ephemeris Subsystem — Detailed Tests

### Binary GTAB (Go loader)

- header validation: magic, version, dt_ns>0, n>0, fields_mask!=0
- size validation: exact expected size; truncated error
- lookup: exact index, interpolation mid-point, out-of-range returns false
- boundary: last sample has no i+1; no interpolation; out-of-range after end false
- fields_mask combos: tide_bps + tide_raw alignment; offset math correct

### File-backed provider (Go)

- hysteresis: first sample sentinel; small deltas stick, large deltas pass
- stale window: before/after coverage → stale=true; clamp to edges
- mapping: bps 0→80.0, 10000→130.0; in-range stays within [80,130]
- context: cancellation returns ctx.Err() quickly
- dataset_id: auto-read from sibling gtab.meta.json when not provided

### Generator (Python)

- CLI: writes gtab_1s.bin, gtab_60s.bin, gtab.meta.json; fields_mask honored
- determinism: same config → same dataset_id
- short windows: n=1,n=2 files still valid
- missing kernel: clear error message and exit code
- E2E: tiny range generate + Go loader reads coverage and first/last values

### Accuracy Harness

- 14 days @1s comparison vs Skyfield (nightly/offline job)
- metrics: median/P99 relative error; peak/zero-cross drift ≤ 3 min
- artifacts: scripts/ephem/validation/metrics.json + optional plots

### Performance

- microbench: IndexFor + LookupTideBPS P95 < 1 ms
- endpoint latency: /gravimetrics P95 < 5 ms locally under light load

## Packaging & Run

- .app wrapper starts server, opens browser; fallback to free port if default busy
- offline ephemeris: dataset discovered via EPHEM_TABLE_PATH or ./ephem; no network I/O
- shutdown: SIGINT/SIGTERM triggers graceful server shutdown and provider Close()

## Artifacts & Reporting

- Upload coverage reports (Go), gas snapshots, accuracy metrics.json, and perf summaries to CI
- Store small GTAB fixtures under tests (avoid large binaries); generate larger windows on-demand

### Current Performance Baselines (Aug 2025, Apple M2, Go 1.21)

| Benchmark            | ns/op | allocs/op | Notes                                       |
| -------------------- | ----- | --------- | ------------------------------------------- |
| GTAB_Lookup          | ~451  | 0         | 1h table, 1s cadence                        |
| FileGravimetricFetch | ~853  | 0         | 24h table, includes interpolation + mapping |

Regression threshold: flag if >2x these numbers without justification (CI gating planned).

## Integration Progress

- [x] GTAB unit tests (header validation, truncation, interpolation)
- [x] File provider unit tests (hysteresis, stale, dataset id)
- [x] HTTP handler tests (success + error paths)
- [x] File-mode gravimetrics integration test
- [x] File-mode predict integration test
- [ ] Accuracy harness nightly job
- [ ] Benchmark CI gating (fail if >2x baseline)

## Upcoming Additions (Planned Documentation First)

### Accuracy Harness Implementation Plan

1. Implement `scripts/ephem/validation/compute_metrics.py`:

- Args: `--table <path>` `--dataset-id <id>` `--start <RFC3339>` `--end <RFC3339>` `--cadence <s>` `--out metrics.json`
- Loads GTAB (Go code path reference replicated in Python if lightweight, else shell out to Go helper binary later).
- Computes reference via Skyfield; downsample to 5s if duration > 2 days, else 1s.
- Collects errors, peak alignment (simple local maxima detection > N window), computes stats vs thresholds.
- Writes metrics.json conforming to schema; sets `passes`.

2. Add Makefile target `ephem-accuracy` invoking script and writing outputs under `scripts/ephem/validation/out/`.
3. CI job (nightly) runs generation + accuracy, uploads metrics artifact; fails on `passes=false`.
4. Extend schema v1 if additional fields needed (e.g., interpolation_bias, sample_gap_count).

### CI Benchmark Gating Plan

1. Add `bench.sh` script running `go test -run NONE -bench BenchmarkGTAB_Lookup -benchmem` and `BenchmarkFileGravimetricFetch` with JSON output (use `-bench` & parse via awk/go tool).
2. Parse ns/op; compare to baseline file `bench_baselines.json` (committed) with allowed factor (default 2.0) or absolute slack.
3. Fail workflow if exceed; instruct developer to update baseline only with justification in status log (link to commit).
4. Optional: store historical benchmark results in GitHub Actions artifacts for trend.
