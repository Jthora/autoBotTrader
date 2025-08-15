# Project Overview & Vision

## Vision
Deliver a trustless execution (cheatcode) autonomous trading bot on Paradex SuperChain (Starknet) that fuses unconventional signals (astrology + gravimetrics) with future AI-driven innovation (ML / PPO) while maintaining verifiable, gas‑efficient Cairo logic and a rapid light-speed delivery posture toward the Top 3 DEX vision.

## Core Value Props
- Novel alternative data (astro + lunar) recorded on-chain with versioned normalization for reproducibility.
- Modular score pipeline (raw inputs → composite formula vN) enabling ML plug‑in without redeploying base contract.
- Gas-conscious Cairo contract (minimal persistent trade state; event-first design) with clear indexing semantics.
- Layered roles (admin, pusher, ml_oracle) to enable progressive decentralization while preserving early velocity.
- Explicit formula_version + normalization constants → deterministic audits across off/on-chain.
- Structured pathway to unified margin & RPI adapter integration (placeholder interfaces now).

## Primary Goals (Expanded Phases & Sub‑Phases)
| Phase | Label | Goal | Key Outputs |
|-------|-------|------|-------------|
| 0 | Bootstrap | Repo + scaffolding + minimal contract skeleton | Structure, README, base contract, Go API stub |
| 0.5 | Deterministic Signals | Normalization constants + formula_version + integer scaling | NORMALIZATION_CONSTANTS ref, formula v1 docs |
| 1 | Signal Flow | Fetch → normalize → push → compute composite | set_prediction_inputs (batched) + PredictionUpdated event |
| 1.5 | Access Control | Introduce pusher + ml_oracle roles & tests | Role storage, ensure_pusher, ensure_oracle |
| 2 | Trading Logic | Thresholded trade execution + minimal trade persistence | execute_trade + TradeExecuted events |
| 2.5 | (Deferred Option) Commit–Reveal | Design & ADR only (not MVP) | ADR commit_reveal (Deferred) |
| 3 | ML Hook | ml_score + ml_model_version + stale check | set_ml_score + MLScoreUpdated event |
| 3.5 | Oracle Signature | Signed ml_score payload spec (ADR) | ML_ORACLE_SPEC draft |
| 4 | Hardening | Gas, cooldown, spam mitigation, indexing schema | cooldown_seconds, subgraph schema draft |
| 5 | Deployment & UX | Testnet deployment + dashboards + runbook | Deployed addresses, RUNBOOK, risk disclaimer |
| 6 | Adapters & RPI | SuperDEX adapter & unified margin placeholder | ISuperDexAdapter interface, adapter events |

## Success Criteria
- Composite formula (versioned) reproducible off-chain (bit‑exact integer math).
- Normalization constants published & tested (golden test vectors).
- Roles enforced (admin, pusher, ml_oracle) with negative tests.
- Gas baseline established (< target X units per prediction update; tracked trendline).
- Frontend shows live composite + ≥3 executed trades + formula_version indicator.
- ML path integrated (dummy signature pipeline) without breaking baseline score path.
- Event-only (or capped) trade logging validated; storage usage remains bounded.
- Progress automation auto-updates metrics & status log after task completion.

## Non-Goals (Initial MVP)
- Full commit–reveal (documented & deferred).
- Production decentralized oracle network (single signer / future multi-sig path only).
- Advanced risk / liquidation modeling (off-chain research later).
- Full DAO governance (upgrade hooks architected, not enacted).
- Real RPI/unified margin settlement (adapter stub & events only).

## Tooling & Wallet Clarification
- Cairo 2.x build: Scarb + Starkli (Hardhat removed from critical path; optional plugin ADR TBD).
- Local devnet: Katana (or starknet-devnet) for fast iteration.
- Wallets: ArgentX / Braavos primary; MetaMask bridging future note (deferred).

## Disclaimers & Ethics
This system uses experimental astrology/gravimetrics signals. Not financial advice. Users must assess risk; historical performance does not guarantee future results. Transparency (events + deterministic formula) provided to enable independent verification.

---
See `01_architecture.md` for detailed system components.
