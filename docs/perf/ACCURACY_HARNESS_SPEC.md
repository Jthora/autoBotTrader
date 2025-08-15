# Ephemeris Accuracy Harness Specification

Formal definition of the nightly validation job measuring drift between on-disk GTAB tables and authoritative astronomical ephemerides (Skyfield/JPL).

## Objectives

- Detect silent corruption or stale GTAB generation.
- Quantify interpolation error characteristics (central vs edge intervals).
- Track peak timing and magnitude drift for tidal force cycles.

## Scope

Validates gravimetric proxy values derived from GTAB against recomputed values from authoritative sources over a rolling window (e.g., last 7 days + next 1 day predictive horizon if available).

## Data Sources

- GTAB binary file (v1) used in production (`$GTAB_PATH`).
- Astronomical kernels: DE421+ (cached) via Skyfield.
- Location assumption: Earth-Moon-Sun system (extendable with additional bodies later).

## Metrics (All per-run, JSON schema already in `scripts/ephem/validation/metrics.schema.json`)

- records_compared
- rel_error_mean, rel_error_median, rel_error_p95, rel_error_p99, rel_error_max
- abs_error_mean, abs_error_p95, abs_error_p99, abs_error_max
- peak_count
- peak_timing_drift_seconds_mean, median, p95, p99, max
- peak_value_rel_drift_mean, p95, p99, max
- passes (boolean)
- thresholds object (echo of config)

## Threshold Defaults

| Metric                        | Threshold              |
| ----------------------------- | ---------------------- |
| rel_error_median              | <= 1.0e-3              |
| rel_error_p99                 | <= 5.0e-3              |
| peak_timing_drift_seconds_p99 | <= 120                 |
| peak_value_rel_drift_p99      | <= 2.0e-3              |
| abs_error_max                 | informational (logged) |

`passes=true` iff all hard thresholds satisfied.

## Sampling Strategy

- Cadence: 60s (configurable) or native table cadence if finer (<60s).
- Range: From (NOW - 7d) to (NOW) ensuring GTAB coverage.
- If coverage gap: mark `passes=false` with reason `coverage_gap`.
- Edge Handling: Exclude first & last record of table from error stats (avoid partial interpolation bias); still counted in coverage assessment.

## Computation Steps

1. Load GTAB (validate header).
2. Build authoritative timescale with Skyfield for sampling timestamps.
3. For each timestamp:
   - Get GTAB value (interpolate if needed).
   - Compute authoritative value (physics function TBD; currently placeholder using gravitational parameter + distances).
   - Record absolute & relative error.
4. Peak Detection:
   - Identify local maxima/minima in authoritative series using 3-point window.
   - For each peak, locate nearest peak in GTAB series within ±2 \* cadence seconds.
   - Compute timing drift and relative magnitude drift.
5. Aggregate metrics; apply thresholds.
6. Emit JSON.

## CLI Interface (Proposed `compute_metrics.py`)

```
usage: compute_metrics.py --table PATH --dataset-id ID --start ISO8601 --end ISO8601 --cadence SECONDS --out PATH [--thresholds PATH] [--verbose]
```

## Error Modes

| Condition                       | Action                                                            |
| ------------------------------- | ----------------------------------------------------------------- |
| Header invalid                  | exit 2, emit minimal metrics with passes=false, reason=bad_header |
| Coverage gap                    | passes=false, reason=coverage_gap                                 |
| Authoritative failure (network) | retry 2x, else passes=false, reason=auth_fetch_failed             |
| <10 records after sampling      | passes=false, reason=insufficient_samples                         |

## Extensibility

- Additional bodies: parameterize gravitational influences.
- Multiple datasets: allow list of GTAB files, aggregate comparative metrics per dataset.
- Trend analysis: future job can diff last N nightly metrics to detect gradual drift.

## Output Example

```json
{
  "dataset_id": "nightly-2024-01-01",
  "generated_at": "2024-01-01T03:05:00Z",
  "records_compared": 10080,
  "rel_error_median": 5.2e-4,
  "rel_error_p99": 3.1e-3,
  "peak_count": 28,
  "peak_timing_drift_seconds_p99": 74,
  "peak_value_rel_drift_p99": 1.4e-3,
  "passes": true,
  "thresholds": { ... }
}
```

## CI Integration

- Scheduled nightly workflow runs harness.
- If fails: create issue (or reopen existing) labeled `accuracy-regression` with diff summary (prior vs current p95/p99 metrics).
- Upload metrics artifact and (optional) CSV of raw sample pairs for deep dive.

## Security & Integrity

- Pin versions of ephemeris libraries (requirements.txt) to avoid silent numerical drift.
- Hash GTAB input file; include sha256 in metrics JSON for traceability.

## Open Items

- Define authoritative physics function precisely (document formula reference & constants).
- Decide on storing historical metrics (S3, repo, or issue comments).
- Implement CSV optional export flag.
