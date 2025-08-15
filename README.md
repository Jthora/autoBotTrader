# autoBotTrader

Autonomous Starknet / Paradex trading bot leveraging high‑fidelity lunar gravimetric (tide force) signal ingestion today, with a roadmap for on‑chain integrated ML (PPO / policy gradient) to pursue passive yield via disciplined, thresholded execution and future market‑making strategies. Contracts (Cairo), off‑chain service (Go), frontend (React), and ephemeris tooling (Python) live in a single monorepo.

## Key Docs

- High-Level Overview: `docs/initialPlan/00_overview.md`
- Contract Spec: `docs/initialPlan/02_contract_spec.md`
- Off-Chain Service Spec: `docs/initialPlan/04_offchain_service_spec.md`
- Testing Strategy: `docs/initialPlan/09_testing_strategy.md`
- Developer Guide (run/test workflows): `DEVELOPER_GUIDE.md`
- Progress System & Metrics: `docs/progress/README.md`

## Quick Start

Prereqs: Go 1.21+, Python 3.11+, Node 18+, Scarb (Cairo 1). Optional: virtualenv.

Clone & install:

```bash
git clone https://github.com/Jthora/autoBotTrader.git
cd autoBotTrader
python -m venv .venv && source .venv/bin/activate && pip install -r requirements.txt  # optional but recommended
cd frontend && npm install && cd ..
```

Run API (mock mode):

```bash
go run ./api/cmd/server
```

Generate ephemeris & run in file mode:

```bash
python scripts/ephem/generate.py --out ephem
EPHEM_MODE=file EPHEM_TABLE_PATH=./ephem/gtab_1s.bin go run ./api/cmd/server
```

Contracts tests:

```bash
cd contracts && scarb test
```

Frontend dev server:

```bash
cd frontend && npm run dev
```

Unified make targets (`make help`) are also available; VS Code tasks provide one‑click runs.

## Governance & Quality Gates

- Placeholder enforcement: `scripts/ci/check_placeholders.py` fails CI if critical files contain placeholder tokens.
- Gas snapshot enforcement: `python scripts/progress/gas_snapshot.py` exits non-zero when avg_gas values are zero (unless `ALLOW_GAS_PLACEHOLDERS=1`).
- Kernel integrity: `scripts/ephem/validation/verify_kernel.py` manages kernel hash placeholders and detects corruption.
- Fuzzing: `go test ./api/internal/ephem -run Fuzz -fuzz FuzzLookupTideBPS -fuzztime=5s` (optional locally) exercises interpolation invariants.

See `docs/testing/TEST_PLAN_EXPANDED.md` for full test matrix status.

## Progress Automation

```bash
python scripts/progress/update_progress.py --summary --write
```

## Contributing

See `CONTRIBUTING.md` (task graph workflow, ADR templates under `docs/progress/templates/`).

## Security

See `SECURITY.md`. Experimental; not financial advice.

## License

Released under CC0 1.0 Universal (public domain dedication). See `LICENSE`.

To the extent possible under law, the authors have waived all copyright and related rights. Attribution is appreciated but not required.
