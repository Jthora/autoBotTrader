# First Checkpoint Plan

## 0. Mission (Rehash)

Build an autonomous Paradex (Starknet L2) trading bot that:

- Aggregates unconventional signals (astrology + gravimetrics) today; cleanly extends to ML scores tomorrow.
- Keeps core decision & execution logic on-chain (transparent, event-sourced, upgrade-minimal).
- Minimizes trusted off-chain surface to only data acquisition & UX.
- Emphasizes safety (threshold + weights + cooldown) and observability (rich events) before performance.

Checkpoint Goal: Establish a reliable, minimally integrated slice: validated contract behavior (stateful tests), a real (or feature-flagged) tx path from Go /push -> Starknet, deployment script stub, and foundational Go tests—ready to layer UI & ML.

Success Criteria (Checkpoint):

1. Stateful contract test suite (>=6 core tests) all pass.
2. /push returns a non-mock tx hash when env credentials supplied (mock hash otherwise) within <500ms median (local).
3. Deployment script produces contract address & updates .env (dry-run acceptable if keys absent).
4. Redundant contract sources removed; task graph consistent & valid YAML.
5. Go unit tests cover normalization + handlers (push happy path + failure) with >80% statements in those packages (stretch).

Guiding Principles: Deterministic Core, Minimal State Surface, Event-First Auditability, Extensibility for ML, Safety Before Aggression, Progressive Decentralization.

This checkpoint captures the current implementation state and the focused path to reach a minimally integrated, test-validated system.

## 1. Snapshot Summary

- Contract logic present (scores, weights, threshold, cooldown, ml hook, execute_trade) in `contracts/src/lib.cairo`.
- Off-chain Go API live with endpoints: /health, /astrology, /gravimetrics, /predict, /push.
- Chain client stub integrated (env gated) but not performing real Starknet RPC.
- Tests: ONLY pure composite arithmetic tests (stateful tests deferred due to harness mismatch).
- Frontend scaffold present (Vite React) and decoupled from runtime; not yet wired to live API data.
- Deployment scripts absent; ML script absent.
- Task graph drift: statuses for `go-chain-client`, `contract-admin-access` not fully aligned; redundancy in contract source files (TradingBot.cairo, compute.cairo) pending consolidation.
- Ephemeris strategy decided for prototype: local-only runtime with precomputed data embedded in Go binary (fallback: pure-Go algorithms via Meeus/VSOP87 after validation).

## 2. Immediate Objectives (High Priority)

1. Reinstate essential stateful contract tests (deploy, access control, cooldown, execute_trade success & revert paths, weight recompute).
2. Consolidate contract sources (remove stale `TradingBot.cairo` and `compute.cairo` if superseded by `lib.cairo`).
3. Correct task graph statuses; add consolidation task; record decision for event-only trade logging (ensure ADR reference clear).
4. Implement minimal real Starknet transaction send in chain client (feature flag) OR stub interface that can be swapped without code churn.
5. Add Go unit tests: normalization, handlers (/predict, /push) with mock ChainClient, chain client validations.
6. Introduce EphemerisProvider in Go and wire `/gravimetrics` to it with a feature flag: `EPHEM_MODE=file|algo` (default file).
7. Generate and embed a 60–90 day ephemeris-derived scalar table (1–5 min cadence) into the Go binary (no network at runtime).

## 3. Secondary Objectives (Next Wave)

6. Add deployment script (starkli declare + deploy + output .env update).
7. Scaffold frontend (Vite React, basic dashboard hitting /predict, /push).
8. Introduce ML placeholder script in `ml/predict.py` producing random score 0-100.
9. Documentation sync: README updates, task graph indentation fixes, clarify testing strategy & roadmap.
10. Add minimal event query abstraction placeholder for future UI trade history.
11. Validate pure-Go ephemeris algorithm (Meeus/VSOP87) against Skyfield reference and switch to `algo` mode if acceptance criteria met.

## 4. Risks & Mitigations

| Risk                     | Impact                         | Mitigation                                                          |
| ------------------------ | ------------------------------ | ------------------------------------------------------------------- |
| Missing stateful tests   | Undetected regressions         | Reintroduce minimal snforge suite ASAP                              |
| Chain client stub only   | Blocks real integration        | Implement minimal RPC signer or adapter soon                        |
| Redundant contract files | Confusion, accidental edits    | Consolidate and remove superseded sources                           |
| Task graph drift         | Planning automation errors     | Update statuses & add notes now                                     |
| Lack of frontend         | Hard to manually validate flow | Scaffold minimal UI early                                           |
| Time scales/ΔT handling  | Subtle inaccuracies            | Fix to UTC at runtime; ΔT handled in reference precompute           |
| Port conflicts/offline   | UX issues at demo time         | Auto-pick free port; offline ephemeris by default                   |
| Licensing of kernels     | Redistribution constraints     | Embed derived scalar table only; avoid bundling JPL kernels in repo |

## 5. Contract Test Reintroduction Plan

- Tooling: Evaluate current snforge / starknet testing harness for Cairo 2.12; adapt test harness calls.
- Minimal Test Set:
  1. `test_init_defaults` – asserts constructor-initialized values.
  2. `test_set_prediction_updates_composite` – sets scores, checks composite.
  3. `test_cooldown_enforced` – set cooldown, expect second immediate call revert.
  4. `test_access_control_not_admin` – weight update / threshold update revert.
  5. `test_execute_trade_threshold` – below threshold reverts, adjusting threshold allows trade.
  6. `test_set_ml_score_recomputes` – set ml weight >0, update ml score, composite changes.

## 6. Chain Client Incremental Implementation

- Phase 1 (current): Validation + latency stub (DONE).
- Phase 2: Add interface for RPC (config struct: RPC URL, Account Address, Private Key / Signer).
- Phase 3: Implement `PushPrediction` real call (construct calldata: astrology_score, gravity_score) -> send invoke -> parse tx hash.
- Phase 4: Add `GetComposite` as view call (simulate or RPC call to call entrypoint).
- Phase 5: Add optional retry / backoff & context cancellation tests.

## 7. Deployment Script Outline

- Inputs: compiled Sierra/Casm artifacts (scarb build output).
- Steps: declare -> deploy -> write address to `.env` -> append status_log entry.
- Future: integrate into CI pipeline for tagged releases.

## 8. Frontend Scaffold Goals

- Display: Astrology score, Gravimetric score, Composite (from /predict).
- Actions: Trigger /push (button) show tx hash or DRYRUN.
- Layout: Minimal responsive page; placeholder for trade events list.

## 9. ML Placeholder

- Script: `ml/predict.py` loads nothing, outputs JSON {"ml_score": <0-100>, "timestamp": ...}.
- Future: Replace with PPO training integration; feed into contract via chain client (set_ml_score) in a scheduled job.

## 10. Updated Task Grouping (Roadmap Alignment)

- Phase 0 Hardening: Test reinstatement + consolidation.
- Phase 1 Integration: Real chain client + deployment script + push flow.
- Phase 2 UX: Frontend scaffold + basic display.
- Phase 3 Intelligence: ML placeholder & hook usage.
- Phase 4 Optimization & Security.
- Phase 5 Testnet Deploy & Observability additions.

## 11. Acceptance Snapshot for Checkpoint Completion

Checkpoint considered achieved when:

- (A) Minimal stateful contract tests pass.
- (B) Chain client can send real tx (or feature-flagged execution path) returning non-mock hash in integration test.
- (C) Redundant contract sources pruned; task graph updated.
- (D) Go service has unit tests for normalization + handlers.
- (E) Deployment script present and dry-run outputs target address placeholder.
- (F) Ephemeris is fully local at runtime (no network); `/gravimetrics` powered by embedded table with deterministic outputs; P95 latency < 5ms locally.
- (G) Accuracy report artifact exists comparing Meeus/VSOP87 candidate (if implemented) vs Skyfield reference over ≥14 days (1s cadence for intraday); acceptance thresholds documented.

## 12. Next Action Queue (Concrete)

1. Prune redundant contract files (verify they are obsolete before deletion).
2. Patch task_graph.yaml statuses & add consolidation task.
3. Add Go tests for normalize (fast win).
4. Draft stateful test harness skeleton.
5. Implement chain client interface expansion scaffold for RPC.
6. Add `EphemerisProvider` interface and a file-backed implementation that loads an embedded CSV/JSON, with linear interpolation and simple hysteresis; wire into `/gravimetrics`.
7. Add a generator (Skyfield) under `scripts/ephem/` to produce binary files per `docs/EPHEMERIS_DATA_FORMAT.md` at 1s and 60s cadence (run offline; commit derived output). Document how to run and time span.
8. Add an accuracy harness in Go (or Python) that compares candidate vs reference and emits `metrics.json` and plots.

---

## 13. Ephemeris Strategy (Prototype)

- Requirement: No internet for ephemeris at runtime; speed and determinism favored.
- Approach: Precompute Sun/Moon vectors and a derived gravimetric stress/tide scalar at 1–5 minute cadence for the next 60–90 days using Skyfield + DE440s/DE421, offline. Commit only the derived table (CSV/JSON) and embed into the Go binary using `embed`.
- Alternative: Pure-Go computation using Meeus/VSOP87 (no data files, no cgo). Ship behind a feature flag (`EPHEM_MODE=algo`) after empirical validation. Not on-chain due to cost/complexity; on-chain holds only minimal governance (weights), inputs, and signature checks.

## 14. Packaging & Run (Demo UX)

- Single-binary service serves both API and UI: Build the React app and embed static assets in the Go binary. The server hosts `/` for UI and `/api/*` endpoints.
- macOS app: Provide a minimal `.app` wrapper that launches the server on localhost and opens the browser to `http://localhost:<port>`. No terminal required. Optional ad-hoc codesign for friendlier Gatekeeper behavior.
- Windows/Linux: Ship the standalone binary that opens the default browser on start. All ephemeris assets embedded; no setup.

## 15. Testing & Validation

- Accuracy: Compare tide-scalar from the runtime provider (file-backed or algo) against Skyfield reference over ≥14 days at 1–5 min cadence.
  - Metrics: relative error (median ≤ 0.5%, P99 ≤ 1.5%), peak/zero-crossing timing drift ≤ 3 minutes.
  - Artifacts: `metrics.json` and optional plots committed under `scripts/ephem/validation/`.
- Unit: JSON contracts for `/astrology`, `/gravimetrics`, `/predict`, `/push`; composite math (weights/edge cases); interpolation boundedness; hysteresis behavior.
- Integration: Startup smoke, offline mode (deny network except RPC if set), real tx path happy/failure.
- Performance: P95 endpoint latency < 5ms locally; startup < 300ms to first response.
- Packaging: `.app` opens browser, auto-port selection if 8080 busy; clean shutdown.

## 16. Edge Cases & Mitigations

- Time scales (UTC vs TT/TDB): Fix runtime to UTC; reference uses native scales internally and exports UTC stamps. Use simple ΔT model only if `algo` mode enabled.
- Data range: If current time beyond table, cap to last value and surface a stale flag; avoid extrapolation beyond N minutes.
- Port conflicts: Auto-pick a free port; log and display active port.
- Offline behavior: Ephemeris path never performs network I/O. `/push` remains optional; when RPC env absent, return dry-run hash.
- Licensing: Do not commit JPL kernels; commit derived series only. Pure-Go path avoids data licensing.

## 17. Decisions (Summary)

- Do not compute ephemeris on-chain. On-chain scope: verify signatures, enforce cadence/bounds, store weights and canonical inputs.
- Default provider is file-backed, embedded series (deterministic, ultra-fast). Pure-Go algorithmic provider gated by validation.
- Ship as a single, double-clickable app experience on macOS for the demo; cross-platform binaries available.

Append further updates as progress is made.
