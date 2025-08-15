# Benchmark Baselines & Governance

Defines current performance baselines, comparison method, and update protocol.

## Philosophy

Microbenchmarks are guardrails, not vanity metrics. We accept intentional slowdowns only when they create net user value (correctness, security, clarity) justified in writing.

## Current Baselines (Captured YYYY-MM-DD)

See `bench_baselines.json` (machine: GitHub Actions ubuntu-latest, go 1.21.x).

| Benchmark                     | ns/op | B/op | allocs/op | Notes                              |
| ----------------------------- | ----- | ---- | --------- | ---------------------------------- |
| BenchmarkGTAB_Lookup          | 451   | ~0   | 0         | Hot cache, 1s cadence table        |
| BenchmarkFileGravimetricFetch | 853   | ~0   | 0         | Includes provider hysteresis check |

## Storage File

`docs/perf/bench_baselines.json` (example):

```json
{
  "captured_at": "2024-01-01T00:00:00Z",
  "go_version": "1.21.x",
  "benchmarks": [
    {
      "name": "BenchmarkGTAB_Lookup",
      "ns_per_op": 451,
      "bytes_per_op": 0,
      "allocs_per_op": 0,
      "rationale": "Initial baseline after optimization pass"
    },
    {
      "name": "BenchmarkFileGravimetricFetch",
      "ns_per_op": 853,
      "bytes_per_op": 0,
      "allocs_per_op": 0,
      "rationale": "Includes meta/stale evaluation"
    }
  ]
}
```

## Comparison Rules

- Regression threshold: fail if ns/op > baseline \* 2.0 OR (ns/op - baseline) > 500ns (whichever larger window) AND allocs increase > 0 unless justified.
- Soft warning if ns/op increase between +20% and +100%; post comment, add `perf-watch` label.

## Update Protocol

1. Run full benchmarks locally (3 repetitions) with CPU governor (if possible) or in a temp GitHub Actions run.
2. Compute median ns/op of each benchmark.
3. If change required, edit `bench_baselines.json`: update values and append/modify `rationale` (must mention PR #).
4. Add entry to `docs/progress/status_log.md` summarizing change.
5. If slowdown >10%: include mini ADR or link to existing ADR.

## Adding a Benchmark

- Ensure stable (RSD < 5%) across 5 runs before inclusion.
- Favor microbenchmarks that isolate a single hot path.
- Name must start with `Benchmark` and reflect component (e.g., `BenchmarkGTAB_Lookup`).
- Add to baseline file with rationale.

## Removing a Benchmark

- Only if obsolete or superseded; document reasoning in status_log + PR description.
- Mark as deprecated one release before removal when possible.

## Tooling

`compare_bench.py` (planned) will:

- Parse `go test -bench` output.
- Normalize names.
- Load baseline JSON.
- Emit JSON result + nonzero exit on violation.

## Risk Mitigations

- Noise: limit PR runs to small subset; run full suite nightly if expanded.
- Skew: avoid running on heterogeneous runners for gating (stick to ubuntu-latest).
- Drift: quarterly review ensures baselines still representative.

## Open Items

- (DONE) Implement `scripts/ci/compare_bench.py`.
- Consider p95 HTTP latency tracking if/when higher-level endpoint benchmarks added.
