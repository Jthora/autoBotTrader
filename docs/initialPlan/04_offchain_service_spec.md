# Off-Chain Go Service Specification

## Purpose

Fetch external signals (astrology, gravimetrics), normalize them into bounded scores (0–100), compute preview composite, and push on-chain inputs.

## Endpoints

| Method | Path          | Description                                   |
| ------ | ------------- | --------------------------------------------- |
| GET    | /health       | Liveness check                                |
| GET    | /astrology    | Returns raw + normalized astrology data       |
| GET    | /gravimetrics | Returns raw + normalized gravity data         |
| GET    | /predict      | Aggregates both, returns composite preview    |
| POST   | /push         | Forces fetch + on-chain set_prediction_inputs |
| POST   | /ml           | (Future) Push ml_score                        |

## Data Models (JSON)

```
AstrologyResponse {
  provider: string,
  raw: { planets: PlanetPosition[] },
  normalized_score: int (0-100),
  calc_version: string
}

GravimetricsResponse {
  provider: string,
  raw: { lunar_tide_force: float, phase: string },
  normalized_score: int,
  calc_version: string
}

PredictResponse {
  astrology: AstrologyResponse,
  gravimetrics: GravimetricsResponse,
  composite_preview: int,
  weights: { astrology: int, gravity: int, ml: int },
  version: string
}
```

## Providers Abstraction

```
type Provider interface {
    Name() string
    Fetch(ctx context.Context) (RawData, error)
}
```

Each provider handles its own API (mock now). Add retry with backoff.

## Deterministic Normalization Strategy v1

- All arithmetic integer-based (avoid float drift).
- Publish constants (min/max ranges) in NORMALIZATION_CONSTANTS.md with version (normalization_version = 1).
- Astrology: compute volatility_index -> scale to 0–10_000 (basis points) → compress to 0–100 by /100.
- Gravimetrics: linear map from (minForce, maxForce) range to 0–10_000; clamp → /100.
- Provide golden vectors (input → normalized) for regression tests.

## Concurrency Pattern

```
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
wg + channels collect results; failing provider sets degraded flag.
```

## On-Chain Interaction

- Env: STARKNET_RPC_URL, CONTRACT_ADDRESS, PUSHER_PRIVATE_KEY, ML_ORACLE_PRIVATE_KEY (optional).
- Build calldata referencing formula_version & normalization_version (readable from contract or cached).
- Retries with exponential backoff (max 3) + jitter; abort after successive provider failures.

## Error Handling

- Distinguish provider failure vs on-chain failure
- /predict returns partial with warnings array

## Logging

- Structured JSON: level, ts, msg, component, correlation_id

## Security

- Pusher key signing; server restricts /push to internal or authenticated calls.
- Optional HMAC on POST /push (header X-Signature).
- Rate limit by IP + cooldown mirrored with contract state (do not attempt if cooldown active).
- ML score signature validation placeholder (future ECDSA/Stark signature).

## Future

- Rate limiting via token bucket
- Metrics: Prometheus /metrics endpoint

---

## Appendix: Aug 2025 Update — Ephemeris and Runtime Model

- Runtime ephemeris is fully local and offline. Default provider uses a binary, memory-mapped series per `docs/EPHEMERIS_DATA_FORMAT.md` (coverage: 60–90 days @ 1s cadence; plus 6–12 months @ 60s).
- Feature flag `EPHEM_MODE` toggles providers: `file` (default) or `algo` (pure-Go Meeus/VSOP87 after validation).
- The service signs submitted inputs; the contract stores a verifier public key and enforces signature validity and cadence/bounds before state updates.
- The Go service also serves the React UI (embedded static assets) to enable a double-click, no-terminal experience.

### Endpoint Additions

- GET /gravimetrics response now includes:
  - `version`: ephemeris dataset version and git commit hash
  - `mode`: "file" or "algo"
  - `stale`: boolean; true if current time is outside embedded table range

### Configuration Keys

- `EPHEM_MODE`: one of `file` (default) or `algo`.
- `HYSTERESIS_BPS`: integer basis points to stabilize decisions around thresholds.
- `EPHEM_TABLE_PATH` (dev-only): override path(s) for the binary files (e.g., `gtab_1s.bin`, `gtab_60s.bin`).

### Lookup Semantics

- O(1) index by timestamp into fixed-size records; linear interpolation between neighboring samples.
- If requested time is outside the dataset range: return last value with `stale=true` or fail closed per config.

- `API_PORT`: server port; auto-fallback to a free port if busy.
- `CONTRACT_ADDRESS`, `STARKNET_RPC`, `STARKNET_KEY`: enable /push on Starknet.

---

Implementation plan in `05_offchain_service_impl_plan.md`.
