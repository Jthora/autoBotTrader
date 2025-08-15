# Ephemeris/Tide Generator Guide (GTAB 1s/60s)

This guide describes how to generate the binary, memory-mapped geocentric gravity tide/stress datasets (GTAB v1) used by the Go service for ultra-fast lookups.

See also: `docs/EPHEMERIS_DATA_FORMAT.md` (file layout and loader placement).

## Goals

- Fully offline at runtime (no network).
- O(1) random access by timestamp; second-level cadence with millisecond interpolation.
- Deterministic and versioned datasets for reproducibility.

## Outputs

- `gtab_1s.bin` — 1-second cadence, 60–90 days coverage (Sun+Moon; optionally distances/raw terms).
- `gtab_60s.bin` — 60-second cadence, 6–12 months coverage.
- `gtab.meta.json` — human-readable metadata (dataset_id, generator commit, kernel info, ranges, fields_mask). The loader can surface the `dataset_id` in API responses.

## Dependencies

- Python 3.10+
- Packages: `skyfield`, `jplephem`, `numpy`, (optional) `click` for CLI

## Recommended Source of Truth

- Ephemeris: JPL Development Ephemerides — prefer DE440s (1849–2150) for current work; DE421 (1900–2050) acceptable for smaller files.
- Library: Python Skyfield (with jplephem) for generation. It reads JPL SPK kernels and returns geometric state vectors easily.
- Download: Fetch the SPK kernel (e.g., `de440s.bsp` or `de421.bsp`) from NAIF’s generic kernels (planets) directory and store it locally; pass path via `--kernel`.
  - Example filenames: `de440s.bsp`, `de421.bsp`.
  - Do not commit kernels to the repo; only commit derived GTAB files and metadata.

Kernel acquisition (links):

- Planets (SPK): https://naif.jpl.nasa.gov/pub/naif/generic_kernels/spk/planets/
- Moons (if ever needed): https://naif.jpl.nasa.gov/pub/naif/generic_kernels/spk/satellites/

Why this choice (and not others):

- Skyfield + JPL DE provides authoritative positions with simple APIs and permissive redistribution for derived data. Swiss Ephemeris has licensing constraints; CSPICE adds cgo/toolchain complexity for our Go runtime. We keep Python confined to the offline generator.

## Ephemeris and Geometry

- Use JPL DE kernels (e.g., DE440s or DE421). Download locally before running.
- Use geometric positions (no light-time/apparent corrections) in ICRF.
- Compute geocentric vectors at time t:
  - r_e = barycentric Earth geocenter position (km)
  - r_s = barycentric Sun position (km); r_geo_sun = r_s − r_e
  - r_m = barycentric Moon position (km); r_geo_moon = r_m − r_e
- Distances: |r*geo*\*| in km.
- Tidal scalar choices (pick one; default first):
  - raw_scalar = μ_sun/|r_geo_sun|^3 + μ_moon/|r_geo_moon|^3 (km^-3 s^-2 units)
  - optional: Frobenius norm of tidal tensor if desired (heavier; omit by default)
- Normalized scalar: map raw_scalar to 0–100 via robust percentiles over the generated window (e.g., p5→0, p95→100; clamp).

Cross-language interface?

- No runtime cross-language calls. The Python generator runs offline to produce GTAB binary files. The Go service memory-maps those files at runtime; no Python or network required in production.

## Fields and Size Trade-offs

- Minimal fast path: `tide_bps` (uint16) only — smallest files.
- Recommended: `tide_bps` + `tide_raw_f32` — keeps debug/analysis possible.
- Optional extras: `moon_r_km_f32`, `sun_r_km_f32`, `moon_rinv3_f32`, `sun_rinv3_f32` — helpful for diagnostics; increase size.
- Set `fields_mask` accordingly (see EPHEMERIS_DATA_FORMAT.md).

## CLI Sketch

Example flags:

- `--kernel /path/de440s.bsp`
- `--start 2025-08-01T00:00:00Z`
- `--end 2025-11-01T00:00:00Z`
- `--cadences 1s,60s`
- `--out-dir ./ephem`
- `--fields tide_bps,tide_raw` (or include distances)
- `--norm p5,p95` (percentile anchors for normalization)

Behavior:

- Generate a monotonic UTC timeline at each cadence.
- For each t, compute geometric r_geo vectors and distances; compute raw_scalar and tide_bps.
- Write GTAB v1 files (little-endian) with correct header and fixed-size records.
- Write `gtab.meta.json` with keys: dataset_id (e.g., sha256 of config), generator_commit, kernel_name, start/end, cadences, fields_mask, norm_anchors, created_at.

## Versioning & Provenance

- Compute a deterministic `dataset_id` as a short hash of: kernel name+version, time range, cadence(s), fields_mask, normalization anchors, and generator code version.
- The Go API should surface this `dataset_id` in `/health` and `/gravimetrics` responses.

## Validation (Quick)

- Sample-check a handful of timestamps against Skyfield live computations (geometric) to ensure parity.
- Run the accuracy harness (see testing strategy) for ≥14 days at 1s cadence; confirm thresholds.

## Ops Notes

- Do not commit JPL kernels; commit only GTAB files and meta JSON.
- Use `scripts/ephem/validation/fetch_kernel.py` to populate local cache (`scripts/ephem/cache/`).
- Verify integrity: `verify_kernel.py --kernel scripts/ephem/cache/de440s.bsp` (add `--update` once to replace placeholder hash after first trusted download).
- Regenerate on a schedule (e.g., monthly) to extend coverage; always bump `dataset_id`.
- Keep files under `MyApp.app/Contents/Resources/ephem/` (macOS) or `./ephem/` beside the binary.
