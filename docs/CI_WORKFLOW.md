# CI Workflow (Proposed)

This document specifies the continuous integration pipeline stages, commands, artifacts, and failure criteria.

## Goals

- Fast feedback (<8 min) on PRs.
- Deterministic, reproducible builds.
- Guardrails for performance, accuracy, and gas regressions.

## High-Level Stages

1. Checkout & Setup
2. Static Analysis / Lint
3. Unit & Integration Tests
4. Benchmarks (sample + compare)
5. Ephemeris Accuracy (scheduled/nightly)
6. Artifact Publication
7. Placeholder Scan (critical files free of tokens)

## Job Matrix Overview

| Job              | Trigger           | Key Tools                | Caches             | Fails On                              |
| ---------------- | ----------------- | ------------------------ | ------------------ | ------------------------------------- |
| lint-go          | PR, push          | go vet, golangci-lint    | Go module cache    | any lint issue                        |
| lint-frontend    | PR, push          | eslint, tsc --noEmit     | node_modules       | lint/type errors                      |
| lint-cairo       | PR, push          | scarb fmt --check        | scarb cache        | formatting mismatch                   |
| test-go          | PR, push          | go test -race -cover     | Go cache           | test failure, race, < target coverage |
| test-contracts   | PR, push          | scarb test               | scarb cache        | test failure                          |
| test-frontend    | PR, push          | npm test (future)        | node_modules       | test failure                          |
| bench-sample     | PR (label opt-in) | go test -bench (subset)  | Go cache           | >2x ns/op vs baseline                 |
| accuracy-nightly | schedule          | python validation script | pip + kernel cache | metrics.passes=false                  |
| gas-snapshot     | PR, push          | scarb test + parser      | scarb cache        | >5% gas delta w/o ADR                 |
| placeholders     | PR, push          | check_placeholders.py    | n/a                | placeholder tokens present            |

## Commands (Canonical)

- Go tests: `go test -race -cover ./api/...`
- Benchmarks: `go test -run NONE -bench BenchmarkGTAB_Lookup -benchmem ./api/internal/ephem`
- Contracts: `scarb test`
- Frontend build (sanity): `npm run build --prefix frontend`
- Accuracy (nightly): `python scripts/ephem/validation/compute_metrics.py --table ephem/gtab_1s.bin --dataset-id nightly --start <start> --end <end> --cadence 1 --out metrics.json`

## Coverage Targets

| Component | Threshold                                  |
| --------- | ------------------------------------------ |
| Go (api)  | 80% lines (excluding generated)            |
| Contracts | 90% critical logic (manual classification) |
| Frontend  | 70% key components                         |

## Performance Baseline Handling

- Baselines stored in `docs/perf/bench_baselines.json`.
- PR benchmark job runs selected microbenchmarks.
- If ns/op > baseline \* 2.0 (or absolute regression > +500ns, whichever larger) -> job fails.
- Updating baselines requires: status_log entry + rationale field appended to baseline item.

## Ephemeris Accuracy

- Nightly only (avoid long runtime on every PR).
- Kernel caching step: run `fetch_kernel.py` to ensure presence + integrity before metrics script.
- Fails if: rel_error_median > threshold OR rel_error_p99 > threshold OR peak_drift_seconds_p99 > threshold OR passes=false.
- Outputs: `scripts/ephem/validation/out/metrics.json` uploaded as artifact.
- To set initial kernel hash: download via fetch script (placeholder hash), run verify with `--update` to persist real sha256, commit hash update + status_log note.

## Gas Snapshots

- Extract gas per critical function (parser TBD) into `docs/perf/gas_snapshot.json`.
- Compare against previous main commit; if any exceeds +5% without ADR reference token in PR description -> fail.

## Artifact Summary

| Artifact         | Path               | Retention | Purpose                   |
| ---------------- | ------------------ | --------- | ------------------------- |
| Go coverage      | coverage.out       | 7 days    | trend & gating            |
| Bench results    | bench_results.json | 14 days   | performance trend         |
| Accuracy metrics | metrics.json       | 30 days   | ephemeris drift detection |
| Gas snapshot     | gas_snapshot.json  | 30 days   | cost regression tracking  |
| Frontend build   | frontend/dist      | 7 days    | quick deploy preview      |

## Failure Escalation

1. Automatic job failure annotates PR.
2. Label `needs-investigation` added by bot.
3. Maintainer or agent reviews root cause within 24h.
4. For acceptable intentional regressions: update baseline/threshold + add status_log entry and (if major) ADR.

## Open Questions

- Will accuracy job need caching of Skyfield kernels? (Kernel hash verification scaffold added; caching still TBD)
- Do we integrate smoke E2E before testnet deploy? (Deferred until basic chain client live.)

## Example GitHub Actions Skeleton (Excerpt)

```yaml
name: CI
on:
  pull_request:
  push:
    branches: [main]
  schedule:
    - cron: "0 3 * * *"

jobs:
  test-go:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "1.21" }
      - run: go test -race -cover ./api/...
  bench-sample:
    if: contains(github.event.pull_request.labels.*.name, 'run-bench')
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "1.21" }
      - run: go test -run NONE -bench BenchmarkGTAB_Lookup -benchmem ./api/internal/ephem | tee bench.txt
      - run: python scripts/ci/compare_bench.py --input bench.txt --baseline docs/perf/bench_baselines.json
  placeholders:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: python scripts/ci/check_placeholders.py
```

## Maintenance

- Review thresholds quarterly.
- Rotate artifacts (clean stale snapshots > retention window).
- Periodic audit of skipped/xfail tests (ensure not forgotten).
