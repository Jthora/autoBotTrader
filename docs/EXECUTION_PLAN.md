# Execution Plan (7-day demo -> hardening)

Purpose: a concrete, timeboxed plan to deliver a double-click app demo with offline ephemeris and on-chain verification, then harden and extend.

## Milestones

- M0: Ephemeris format & docs solidified (Done)
- M1: Generator + datasets ready (GTAB 1s/60s)
- M2: Go EphemerisProvider (mmap) wired to /gravimetrics
- M3: Accuracy harness results accepted
- M4: Packaging: macOS .app, browser opens on start
- M5: On-chain push: signature + cadence enforced
- M6: Demo polish: UI metrics, versioning, dry-run mode

## Day-by-Day (target)

Day 1

- Implement `scripts/ephem/generate.py` (Skyfield + DE440s) to write GTAB v1 (1s/60s) + meta JSON.
- Generate initial datasets for the next 90 days (1s) and 12 months (60s).
- Commit GTAB outputs (no kernels) under `docs/_artifacts/ephem/` or repo root `ephem/`.

Day 2

- Implement Go `api/internal/ephem` with mmap reader and Provider interface.
- Add `EPHEM_MODE`, `EPHEM_TABLE_PATH`, `HYSTERESIS_BPS` configs; load 1s/60s files.
- Wire `/gravimetrics` to provider; include version/mode/stale/raw fields.

Day 3

- Microbenchmarks; ensure P95 < 1ms lookup.
- Update `/predict` to consume provider output; ensure composite unaffected.
- Add /health metadata with dataset_id and git commit.

Day 4

- Accuracy harness: compare 14 days @1s to Skyfield live; produce `metrics.json` and plots; document acceptance.
- Fix any deltas; lock normalization anchors if needed.

Day 5

- Cairo: store verifier pubkey; update function enforces cadence + signature; events emit inputs & composite.
- Go: sign payloads; implement `/push` happy path to devnet/testnet.

Day 6

- Packaging: embed UI; .app wrapper that launches server and opens browser.
- UX: show version, mode, stale; simple metrics panel; dry-run if RPC not set.

Day 7

- Hardening: unit/integration tests, docs, and demo rehearsal.
- Stretch: start `algo` provider (Meeus) behind flag and baseline vs GTAB.

## Post-demo hardening

- Add CI for generator validation and dataset signing.
- Commitment scheduling: publish Merkle root of GTAB window on-chain.
- Implement algo provider fully; add 1–2 planet terms as optional.
- macOS codesign/notarize; Windows installer; Linux AppImage.
- Observability: structured logs, basic metrics endpoint.

## Deliverables Checklist

- GTAB 1s/60s datasets + meta JSON (committed)
- Go ephem provider (mmap) + tests
- /gravimetrics with version/mode/stale/raw
- Accuracy report artifacts
- macOS `.app` bundle
- Cairo contract verifying signature + cadence + events

## Risks & Mitigations

- Kernel access friction: Document NAIF links; keep an internal cache path.
- Performance regressions: Bench in CI; mmap always preferred for large files.
- Dataset drift: Lock normalization anchors; version with dataset_id; disclose in UI.
- Time scale pitfalls: Use geometric positions at generation; UTC-only at runtime.
