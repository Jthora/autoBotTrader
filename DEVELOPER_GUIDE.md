# Developer Guide

A quick, centralized reference for running, testing, and debugging all subsystems.

## Layout Overview

- `api/` Go module (see `go.work` at root so you can run from monorepo root)
- `contracts/` Cairo 1 project (Scarb)
- `frontend/` React + Vite
- `scripts/` Python utilities (ephemeris generation, progress tooling)
- `docs/` Design & specs

## First-Time Setup

1. Install Go 1.21+, Python 3.11+, Node 18+, Scarb (see `docs/SCARB_SETUP.md`).
2. (Optional) Create a virtualenv: `python -m venv .venv && source .venv/bin/activate` then `pip install -r requirements.txt`.
3. Install frontend deps: `cd frontend && npm install`.
4. (Optional) Cairo deps: `cd contracts && scarb build`.

## Fast Paths

Unified helpers:

- Root `Makefile` targets: `make help`.
- VS Code tasks (⇧⌘B / Run Task): API test/run, contracts test, frontend dev, progress update.
- `go.work` allows `go run ./api/cmd/server` from repo root.

## Running Components

### API Server

Mock mode (no external files):

```
go run ./api/cmd/server
```

File mode (ephemeris):

```
EPHEM_MODE=file EPHEM_TABLE_PATH=./ephem/gtab_1s.bin go run ./api/cmd/server
```

Optional env:

- `API_PORT` (default 8080)
- `EPHEM_DATASET_ID` (overrides meta JSON)
- `HYSTERESIS_BPS` (basis-point stickiness)

### Ephemeris Generation

```
python scripts/ephem/generate.py --out ephem
```

Produces: `gtab_1s.bin`, `gtab_60s.bin`, `gtab.meta.json`.

### Contracts

```
cd contracts
scarb test
```

### Frontend

```
cd frontend
npm run dev
```

## Testing Strategy Quick Reference

- Go unit tests: `make api-test`
- Cairo contract tests: `make contracts-test`
- Full suite: `make test`
- Race detector always on in Go test target.

Add new tests under `api/internal/...` mirroring package names. For endpoint tests use `httptest` + `NewRouter`.

## Debugging Tips

- Use `/health` for grav provider metadata (mode, dataset, stale).
- For deterministic grav tests, FileGravimetric exposes `FetchAt` (not interface) for direct time lookups.
- If server fails to start from root previously, ensure `go.work` is present (added) or run from `api/`.
- Add quick benchmarks: create `*_test.go` with `Benchmark...` functions.

## Common Issues

| Symptom                                          | Cause                                   | Fix                                                  |
| ------------------------------------------------ | --------------------------------------- | ---------------------------------------------------- |
| `cannot find main module` when running from root | No workspace file                       | Root has `go.work` now; or run inside `api/`         |
| 503 gravimetrics                                 | Ephemeris load failure / provider error | Check startup logs and EPHEM\_\* env vars            |
| Empty dataset id                                 | Missing or unreadable `gtab.meta.json`  | Supply `EPHEM_DATASET_ID` or ensure meta file exists |

## Next Enhancements (Nice-to-have)

- Add Go benchmark: ephem lookup & provider fetch.
- Wire ESLint/Prettier and a `lint` task.
- CI pipeline using `make test`.
- Add integration test harness combining file-mode server + HTTP requests.

## Current Status Snapshot (Aug 2025)

Implemented:

- GTAB binary format reader + file-backed gravimetric provider (hysteresis, stale detection, metadata exposure)
- Mock providers for astrology & gravimetrics
- Graceful shutdown (SIGINT/SIGTERM) in API
- Consistent JSON error envelopes + handler tests (success & error paths)
- Ephemeris generation script (Skyfield/JPL) with meta JSON output
- Developer ergonomics: go.work, Makefile, VS Code tasks, health metadata

Pending / Next Priority Work:

1. Integration test harness: start API in file mode with tiny GTAB, assert endpoint contract (JSON schema + meta) via HTTP client.
2. Performance microbenchmarks (LookupTideBPS, provider Fetch) to track regressions.
3. Accuracy harness nightly job comparing GTAB vs Skyfield reference (store metrics.json). (Scaffold directory added.)
4. Frontend wiring to display grav meta (mode/dataset/stale) and basic push simulation.
5. Cairo contract expansion (weight management, cooldown, role checks) + associated tests.
6. CI setup: lint + test matrix (Go race, Cairo, frontend build) + artifacts upload (coverage, gas snapshot, ephem metrics).
7. Add lint tooling (golangci-lint, ESLint config, Prettier) and integrate into Makefile.
8. Add minimal chain client mock/integration once Starknet RPC details finalized.
9. Integrate placeholder scan (`scripts/ci/check_placeholders.py`) into pre-commit.
10. Implement gas extraction parser (snforge output -> `docs/perf/gas_snapshot.json`).
11. Add accuracy harness script (`scripts/ephem/validation/compute_metrics.py`) + Makefile target `ephem-accuracy`.

Deferred (lower urgency):

- Desktop packaging (.app) + auto-browser open
- Load testing /predict
- Advanced ML pipeline integration hooks

Decision Log References:

- See `docs/progress/decisions.md` for rationale behind binary ephemeris, hysteresis approach, and mock-first strategy.

---

Happy hacking!
