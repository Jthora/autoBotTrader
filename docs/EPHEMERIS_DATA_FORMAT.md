# Ephemeris/Tide Data Format (Binary, Memory-Mapped)

Purpose: provide ultra-fast, offline lookups of geocentric gravity tide/stress metrics at second-level cadence for intraday trading, with optional interpolation to milliseconds, without running heavy ephemeris code at runtime.

## Design Goals

- Zero network at runtime; deterministic outputs.
- O(1) random access by timestamp with no string parsing.
- Small footprint suitable for embedding in the Go binary or shipping alongside.
- Multi-resolution: 1s for intraday; 60s/300s for longer horizons.

## File Layout (GTAB v1)

Header (little-endian):

- magic: 5 bytes = "GTAB1"
- version: uint16 = 1
- epoch_start_utc: int64 (unix seconds)
- dt_ns: int64 (nanoseconds per step) — e.g., 1_000_000_000 for 1s cadence
- n: uint32 (number of samples)
- fields_mask: uint32 bitmask
  - 0x01 tide_bps (uint16) — normalized tide scalar in basis points [0..10000]
  - 0x02 tide_raw_f32 (float32) — raw scalar in SI-like units (s^-2)
  - 0x04 moon_r_km_f32 (float32)
  - 0x08 sun_r_km_f32 (float32)
  - 0x10 moon_rinv3_f32 (float32) — μ_moon / r^3
  - 0x20 sun_rinv3_f32 (float32) — μ_sun / r^3
- reserved: 16 bytes (future use)

Data block:

- Fixed-size records, tightly packed, in this field order: tide_bps?, tide_raw_f32?, moon_r_km_f32?, sun_r_km_f32?, moon_rinv3_f32?, sun_rinv3_f32?
- Record size is determined by fields_mask. Offset(i) = header_size + i \* record_size.

Indexing:

- For a UTC timestamp t, compute i = floor((t - epoch_start_utc) \* 1e9 / dt_ns). If i < 0 or i >= n, the value is out of range.
- For sub-second, linearly interpolate using neighboring samples; for millisecond precision, Hermite interpolation is optional (see below).

## Multi-Resolution Pyramid

- Ship two files:
  - gtab_1s.bin: 1-second cadence, 60–90 days.
  - gtab_60s.bin: 60-second cadence, 6–12 months.
- At runtime, select the coarsest resolution that meets the requested horizon; upsample with interpolation as needed.

## Interpolation

- Default: linear interpolation in time for both tide_bps and tide_raw.
- Optional: cubic Hermite using precomputed slopes (store slope arrays as extra fields if needed). For the demo, linear suffices; acceptance criteria based on timing drift.

## Sizes (order of magnitude)

- 90 days @ 1s cadence: 7.78M samples. If storing only `tide_bps` (2 bytes): ~15.6 MB. With `tide_raw_f32` (+4 bytes): +31.1 MB. Adding two more float32 fields (+8 bytes) totals ~78 MB.
- 12 months @ 60s cadence: ~525k samples. `tide_bps` only: ~1.0 MB; with four float32 fields: ~9.5 MB.

## Embedding and Loading

- Embed via Go `//go:embed` or ship as external files. For large files, prefer external files and memory-map (mmap) for zero-copy reads.
- Loading contract:
  - Verify magic/version.
  - Compute record size from fields_mask; map file; store header metadata.
  - Provide `Lookup(tUTC) -> Sample` with O(1) index, bounds check, and interpolation.

### Loader Placement and Search Order

- macOS app: bundle files under `MyApp.app/Contents/Resources/ephem/`.
- Linux/Windows: place alongside the executable in `./ephem/`.
- Search order: `EPHEM_TABLE_PATH` override -> app Resources -> ./ephem -> current working dir.

## Provenance and Versioning

- Include dataset version and generator commit hash in the Go binary and API responses.
- Do not commit JPL kernels; commit only derived binary series.

## Why Not CSV/JSON?

- CSV/JSON incur string parsing overhead, larger size, and slow random access. The binary layout yields deterministic offsets, tiny parsing cost, and better cache locality.

## Optional: Derivative Storage for Hermite

- To enable high-fidelity ms-level interpolation, store per-sample d(tide_raw)/dt as float32. This doubles storage but remains compact; evaluate after baseline.
