# Go Integration Guide: EphemerisProvider + mmap Reader

This doc explains how to integrate GTAB binary datasets into the Go API for O(1) second-level lookups and expose metadata in `/gravimetrics`.

See: `docs/EPHEMERIS_DATA_FORMAT.md` and `docs/ephem/README.md`.

Source of truth:

- Datasets are generated offline with Python Skyfield + JPL DE kernels into GTAB binaries. At runtime, Go only memory-maps GTAB files; no Python, no network, and no cross-language bridge is required.

## Provider Interface

```go
// Package ephem (api/internal/ephem)

type Sample struct {
    TideBPS      uint16   // 0..10000
    TideRaw      float32  // optional
    Version      string   // dataset_id
    Mode         string   // "file" | "algo"
    Stale        bool
}

type Provider interface {
    At(t time.Time) (Sample, error)
}
```

## File-backed Provider (mmap)

- Load header: validate magic ("GTAB1"), version=1, compute record size from fields_mask.
- Memory-map the file (read-only). Keep a small struct with: epoch_start, dt, n, recordSize, fieldsMask, data slice.
- At(t):
  - Compute i = floor((t.UnixNano() - epoch_start\*1e9)/dt_ns).
  - Clamp to [0, n-1]. If i == n-1 or out-of-range, set Stale=true and return last value.
  - Read record i and i+1, decode fields, linearly interpolate tide fields.
  - Return Sample with Version set from meta (dataset_id) and Mode="file".

## Multi-resolution Selection

- Load both `gtab_1s.bin` and `gtab_60s.bin` if present.
- Selection policy:
  - If request horizon ≤ 90 days, use 1s dataset.
  - Else use 60s and upsample.
- Provide a simple constructor that finds files via search order: `EPHEM_TABLE_PATH` -> app Resources -> ./ephem -> cwd.

## Hysteresis

- Optional smoothing to prevent flip-flop around thresholds:
  - Maintain last output and apply a deadband of `HYSTERESIS_BPS` (basis points) before switching.
  - Keep the logic deterministic and documented.

## Exposing API Fields

- `/gravimetrics` response should include:
  - `normalized_score` (tide_bps / 100)
  - `raw_scalar` (optional)
  - `version` (dataset_id)
  - `mode` ("file" or "algo")
  - `stale` (boolean)

## Algo Provider (Optional)

- Implement `Provider` using pure Go algorithms (Meeus/VSOP87) to compute Sun/Moon geocentric distances and raw scalar.
- Validate against GTAB over ≥14 days; gate via `EPHEM_MODE=algo`.

## Performance Targets

- Lookup: P95 < 1ms on localhost with warmed cache.
- Endpoint: P95 < 5ms including JSON serialization.
- Startup: < 300ms to ready (after mapping files).

## Testing

- Unit tests for header parsing, index math, bounds, linear interpolation.
- Fuzz tests on offsets and record sizes.
- Integration tests confirming `/gravimetrics` schema and metadata.

## Errors and Fallbacks

- If files missing: fallback to 60s or algo provider; mark `stale` as needed.
- If header invalid: return descriptive error; deny serving until corrected.

## Versioning

- Read `gtab.meta.json` and expose `dataset_id` in `/health` and `/gravimetrics`.
- Log both the dataset and the git commit of the Go binary.
