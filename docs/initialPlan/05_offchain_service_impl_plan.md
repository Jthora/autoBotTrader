# Off-Chain Go Service Implementation Plan

## Step Sequence

1. go mod init (module name configurable; placeholder: github.com/PROJECT/autoBotTrader/api)
2. Add dependencies: net/http, encoding/json, context, log/slog (or zerolog), retry (custom), starknet client lib (if available) or placeholder
3. Define domain models (responses + internal structs)
4. Implement provider interfaces (mockAstrologyProvider, mockGravimetricsProvider)
5. Normalization helpers with version constants `NormVersion = "v1"` + golden test vectors
6. Composite calculation replicating contract formula
7. HTTP handlers (pure, testable, no global state where possible)
8. On-chain client wrapper (interface for easy mocking in tests)
9. POST /push flow: fetch -> normalize -> integrity hash (sha256 raw JSON) -> call contract (batched if ml score present) -> return tx hash
10. Logging middleware + request ID
11. Graceful shutdown (context cancellation on SIGINT/SIGTERM)
12. Unit tests (providers, normalization, handlers with mocked providers, contract client)
13. Dockerfile (optional Phase 1) + multi-stage build

## Directory Layout

```
api/
  cmd/server/main.go
  internal/
    http/handlers.go
    http/router.go
    providers/astro.go
    providers/gravity.go
    normalize/normalize.go
    chain/client.go
    chain/mock_client.go
    logging/middleware.go
    config/config.go
  go.mod
  go.sum
```

## Config Loading

Priority order: env > defaults.
Variables:

- PORT (default 8080)
- STARKNET_RPC_URL
- CONTRACT_ADDRESS
- PRIVATE_KEY (hex) – warn if empty
- PUSH_HMAC_SECRET (optional)

## Pseudocode Snippet

```go
func (c *ChainClient) SetPrediction(ctx context.Context, a, g uint32) (string, error) {
    // build calldata
    // sign & send invoke
    // return tx hash
}
```

## Testing Strategy

- Table-driven tests for normalization
- Handler tests using httptest.Server
- Chain client mocked to return deterministic hash
- Race detector in CI (go test -race)

## Performance Considerations

- Reuse HTTP client with timeouts
- Limit provider call time (<=3s each) to keep latency predictable
- Pre-calculate golden vector checks at startup to detect accidental normalization drift

## Future Enhancements

- Real provider integrations
- Caching layer for last successful scores
- Circuit breaker per provider

---

## Aug 2025 Update — Ephemeris Implementation Steps

1. Define `EphemerisProvider` interface in Go and add a file-backed implementation that:

- Loads binary GTAB files (per `docs/EPHEMERIS_DATA_FORMAT.md`) at 1s and 60s cadences (60–90 days and 6–12 months, respectively).
- Interpolates current value; exposes raw basis and a normalized scalar 0–100.
- Applies optional hysteresis (basis points) to reduce flapping.

2. Add feature flag `EPHEM_MODE=file|algo` and a pure-Go algorithmic provider (Meeus/VSOP87) behind validation.
3. Generator script: `scripts/ephem/generate.py` (Skyfield + DE440s/DE421) runs offline and outputs binary files per `EPHEMERIS_DATA_FORMAT.md` at 1s and 60s cadence (commit derived data only).
4. Load via memory-mapped reader in Go (or `io.ReaderAt` fallback). For smaller bundles, `embed` remains viable; for 10–100MB, prefer external files.
5. Extend `/gravimetrics` to include `version`, `mode`, and `stale` fields.
6. Accuracy harness compares provider outputs vs Skyfield for ≥14 days; acceptance: median rel error ≤ 0.5%, P99 ≤ 1.5%, peak timing drift ≤ 3 minutes.
7. Performance targets: P95 < 1ms lookup (hot), < 5ms endpoint; startup < 300ms. Add provider microbenchmarks.

---

Next: See `06_frontend_spec.md`.
