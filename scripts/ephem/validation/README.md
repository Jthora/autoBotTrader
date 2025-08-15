# Ephemeris Accuracy Harness

This directory will host scripts and outputs used to validate GTAB ephemeris data against a high-precision Skyfield/JPL reference.

## Goals

- Detect drift or corruption in generated GTAB tables.
- Quantify interpolation error versus direct ephemeris computation.
- Provide reproducible metrics for CI gating (nightly or scheduled job).

## Metrics (metrics.json Schema v0)

```jsonc
{
  "dataset_id": "string", // GTAB dataset under test
  "generated_at": "RFC3339 timestamp", // time metrics file written
  "range": { "start": "RFC3339", "end": "RFC3339" },
  "cadence_seconds": 1, // sampling cadence
  "samples": 1209600, // total comparison points (example: 14 days @1s)
  "error_basis": "tide_bps", // variable compared
  "stats": {
    "rel_error_median": 0.0004, // median relative error
    "rel_error_p95": 0.007, // 95th percentile relative error
    "rel_error_p99": 0.011, // 99th percentile relative error
    "peak_drift_seconds_median": 45, // median absolute timing drift of local maxima (sec)
    "peak_drift_seconds_p99": 150 // p99 timing drift
  },
  "thresholds": {
    "rel_error_median_max": 0.005,
    "rel_error_p99_max": 0.015,
    "peak_drift_seconds_p99_max": 180
  },
  "passes": true,
  "notes": "optional free-form context"
}
```

## Workflow (Planned)

1. Generate GTAB via scripts/ephem/generate.py.
2. Run validation script (to be added) that:
   - Loads GTAB.
   - Recomputes reference tide metric via Skyfield for each sample (downsampling optional > N seconds).
   - Computes relative error array.
   - Detects peaks (simple local maxima) in both series and matches nearest times for drift stats.
3. Emit metrics.json following schema.
4. CI job reads metrics.json and fails if `passes=false` or any threshold fields exceeded.

## Future Enhancements

- Store historical metrics for trend analysis (rolling window).
- Plot generation (PNG/SVG) summarizing error distribution & drift.
- Adaptive sampling for long ranges.

---

Scaffold only; implementation pending.
