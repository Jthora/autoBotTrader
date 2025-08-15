# Gas Baselines & Governance (Scaffold)

Defines tracking and gating for contract function gas usage.

## Approach

1. Collect gas per critical function (update_prediction, execute_trade) via parsing standard `scarb test` output (heuristic) using `scripts/contracts/extract_gas.py`.
2. Store snapshot in `docs/perf/gas_snapshot.json` each CI run (see CI job `gas-snapshot`).
3. Compare against `gas_baselines.json` (run `make gas-compare`); fail if > threshold_percent increase (default 5%).

## Baseline File Structure

```
{
  "captured_at": "ISO",
  "threshold_percent": 5.0,
  "functions": [{"name": "execute_trade", "avg_gas": 12345, "notes": "PR #42 optimization" }]
}
```

## Update Protocol

1. Produce new snapshot on main after intentional change.
2. Verify regression is justified (ADR if >10%).
3. Update `gas_baselines.json` with new avg_gas and rationale in notes.
4. Append status_log entry.

## Open Items

### Gas Extraction Parser Specification (Current Heuristic)

Extraction currently parses lines of the form:

```
test <module>::<name> ... ok (gas usage est.: <number>)
```

Mapping: specific test name substrings -> logical function bucket.

Mapping Strategy:

- Maintain allowlist in `extract_gas.py`: test substring `set_prediction_updates_composite` -> `update_prediction`; `execute_trade_threshold` -> `execute_trade`.
- If multiple invocations per test, average all occurrences per function.
- Ignore setup/constructor gas.

Parser Steps:

1. Run contract test suite: `scarb test > test_output.txt`.
2. Execute: `python scripts/contracts/extract_gas.py --input test_output.txt --out docs/perf/gas_snapshot.json --fail-on-missing`.
3. Script scans for allowlisted test names, aggregates average gas.
4. Writes snapshot structure `{ generated_at, functions:[{name, avg_gas}] }`.
5. Comparison script `scripts/ci/compare_gas.py` enforces thresholds.

Failure Modes & Handling:

- Missing function: exit 1 with `MISSING_FUNCTION <name>`.
- All zeros: treat as placeholder (existing governance path) and exit 1 unless override env.
- Malformed output: exit 1 with `PARSE_ERROR` + line sample.

Future Enhancements:

- Percentile tracking (p50, p95) once multiple samples gathered (≥5 invocations).
- Separate gas buckets for execute_trade direction (long vs short) if future logic diverges materially.

### Relative Implementation Effort

- Parser & integration: Small (0.5–1 day).
- Baseline update workflow docs finalization: XS (≤1h).
- Percentile extension: S (add stats + schema revision).

## Open Items

- Implement instrumentation / extraction (replace placeholder zeros) — DONE (heuristic parser in place).
- Add multi-percentile tracking (p50, p95) if variance observed.
- Consider per-call-type baselines if execute_trade branches diverge.
