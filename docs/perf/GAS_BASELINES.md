# Gas Baselines & Governance (Scaffold)

Defines tracking and gating for contract function gas usage.

## Approach

1. Collect gas per critical function (update_prediction, execute_trade) via test harness instrumentation (TBD).
2. Store snapshot in `docs/perf/gas_snapshot.json` each CI run.
3. Compare against `gas_baselines.json`; fail if > threshold_percent increase (default 5%).

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

### Gas Extraction Parser Specification (Planned)

Extraction Source Options (choose available):

1. snforge JSON output (preferred if stable): invoke `scarb test --json` or dedicated gas report command; parse emitted gas per test & map to function calls by pattern `fn <entry_point>`.
2. Text output fallback: regex patterns capturing `gas_usage=` or similar tokens from verbose test runs.

Mapping Strategy:

- Maintain allowlist: `set_prediction_inputs` -> `update_prediction` logical bucket; `execute_trade` direct.
- If multiple invocations per test, average all occurrences per function.
- Ignore setup/constructor gas.

Parser Steps:

1. Run contract test suite with gas reporting flag (TBD flag; prototype with verbose output).
2. Collect raw output to temp file.
3. Attempt JSON decode; on failure fallback to regex line scan.
4. Accumulate gas values per allowlisted function name.
5. Write snapshot structure `{ generated_at, functions:[{name, avg_gas}] }`.
6. Validate non-zero; emit FUNCTION_MISSING errors for any allowlisted names absent.

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

- Implement instrumentation / extraction (replace placeholder zeros) — IN PROGRESS (spec above).
- Add multi-percentile tracking (p50, p95) if variance observed.
- Consider per-call-type baselines if execute_trade branches diverge.
