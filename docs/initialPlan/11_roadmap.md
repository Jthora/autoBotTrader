## Aug 2025 Update — Oracle Boundary and Future Hardening

- Short term: keep ephemeris offline and local to Go; precomputed series by default for speed/determinism.
- Medium term: validate and switch to pure-Go algorithmic provider if accuracy thresholds hold.
- On-chain roadmap: add commitment scheduling (pre-published Merkle roots of series) and signature rotation; explore proof systems for selected computations later.
- UX roadmap: move from browser tab to native window via Tauri/Electron if needed; keep single-binary path as baseline.

# Roadmap

## Phase 0 – Bootstrap (Week 1)

- Repo structure
- Planning docs (DONE)
- Contract skeleton + build pipeline
- Go service skeleton
- Frontend scaffold

## Phase 0.5 – Deterministic Signals (Week 1)

- Publish NORMALIZATION_CONSTANTS.md
- Implement integer normalization helpers (off-chain) + golden vectors
- Add formula_version & normalization_version placeholders

## Phase 1 – Signal Flow (Weeks 2–3)

- Implement set_prediction_inputs
- Normalize & push via Go service
- Frontend displays live scores
- Basic unit tests

## Phase 1.5 – Access Control (Week 3)

- Add pusher_role + ml_oracle fields
- Role enforcement tests

## Phase 2 – Trading Logic (Weeks 4–5)

- execute_trade implementation
- Trades events & display table
- Threshold/weights admin UI
- E2E test against local devnet

## Phase 2.5 – (Deferred) Commit–Reveal ADR

- Draft ADR (defer implementation unless front‑running observed)

## Phase 3 – ML Hook (Week 6)

- set_ml_score functional
- Python PPO placeholder
- Composite formula w/ ml included
- Tests for ml weight scenarios

## Phase 3.5 – Oracle Signature Spec (Week 6–7)

- ML_ORACLE_SPEC draft

## Phase 4 – Hardening (Weeks 7–8)

- Gas optimization review
- Access control refinement (pusher role)
- Add logging + metrics to Go service
- Frontend improvements (loading states, error toasts)

## Phase 5 – Deployment & Ops (Week 9)

- Testnet deploy (contract + service + frontend)
- Runbook finalization
- Smoke + regression test matrix

## Phase 6 – Adapters & Extended Features (Post MVP)

- Unified margin / RPI adapter stub
- Advanced charts, analytics
- Oracle decentralization path
- Charting & analytics (historical composite)
- Advanced strategy parameters (dynamic thresholds)
- Governance / DAO path design

## Stretch / R&D

- ZK-proof of ML inference correctness
- Multi-signal modular plugin registry
- Automated keeper for execute_trade

## Milestones & KPIs

| Milestone          | KPI                                       |
| ------------------ | ----------------------------------------- |
| MVP Deployed       | Live composite (v1) & >=3 trades executed |
| ML Hook Live       | ml_score influences composite (ml_w>0)    |
| Hardening Complete | Audit + gas regression guard in CI        |
| Adoption Start     | First external wallet interaction         |

---

Refer back to `00_overview.md` for vision alignment.

---

## Aug 2025 Documentation Retrofit Notes

The following clarifications were added after initial implementation waves to guide the next autonomous steps:

Immediate Focus Order (next tasks):

1. Contract unit tests (`contract-tests` task) fleshing out: prediction inputs, weight updates, cooldown, ml score path.
2. Gas extraction implementation to replace placeholder zeros (`gas-extraction-impl`).
3. Chain client real Starknet tx path (`go-chain-client` completion) then `/push` integration.
4. Ephemeris accuracy harness (`accuracy-harness`) — implement `compute_metrics.py` and nightly job wiring.
5. Placeholder scan script (added) now part of CI gating; extend target list when new critical artifacts emerge.
6. Frontend scaffold & score display once push path can supply live composites.

Governance Guarantees Now Enforced:

- Kernel hash placeholder must be updated before corruption detection passes (see verify script tests).
- Gas snapshot zeros fail unless explicitly overridden.
- Placeholder tokens in critical files fail CI via `check_placeholders.py`.

Pending Documentation Additions:

- Detailed gas extraction parser spec (snforge output -> JSON schema) to be added to `docs/perf/GAS_BASELINES.md` once implemented.
- Accuracy harness metrics schema documentation (`metrics_schema.md`) once script lands.

This retrofit section should be pruned once all above tasks reach COMPLETE and their specs moved into canonical docs.
