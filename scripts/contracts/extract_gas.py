#!/usr/bin/env python3
"""Extract pseudo 'gas' metrics from cairo-test output.

Current limitation: cairo-test output in this project emits lines like:
  test <path>::<module>::<name> ... ok (gas usage est.: 7770)

We map selected test names to logical function buckets (e.g., set_prediction_inputs -> update_prediction).
This is a heuristic until a richer trace or devnet fee estimation path is added.

Usage:
  python scripts/contracts/extract_gas.py --input <captured_test_output.txt> --out docs/perf/gas_snapshot.json

To capture test output:
  scarb test > test_output.txt 2>&1

Exit codes:
  0 success
  1 parse error / missing data
"""
from __future__ import annotations
import argparse
import json
import re
import sys
import statistics
import datetime
import pathlib

# Allowlist mapping: test substring -> logical function name
FUNCTION_MAP = {
    'set_prediction_updates_composite': 'update_prediction',
    'execute_trade_threshold': 'execute_trade',
}

LINE_RE = re.compile(r"gas usage est\.: (\d+)")
TEST_NAME_RE = re.compile(r"test ([^ ]+) \.\.\. ok \(gas usage est\.: (\d+)\)")


def parse(path: pathlib.Path):
    buckets = {fn: [] for fn in set(FUNCTION_MAP.values())}
    missing_keys = set(FUNCTION_MAP.values())
    with path.open() as f:
        for line in f:
            m = TEST_NAME_RE.search(line)
            if not m:
                continue
            test_full = m.group(1)
            gas = int(m.group(2))
            # find mapped logical function(s)
            for needle, logical in FUNCTION_MAP.items():
                if needle in test_full:
                    buckets[logical].append(gas)
                    if logical in missing_keys:
                        missing_keys.remove(logical)
    return buckets, missing_keys


def summarize(buckets):
    functions = []
    for name, values in buckets.items():
        if not values:
            avg = 0
        else:
            avg = int(statistics.mean(values))
        functions.append({"name": name, "avg_gas": avg})
    return functions


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--input', required=True)
    ap.add_argument('--out', required=True)
    ap.add_argument('--fail-on-missing', action='store_true', default=False, help='Fail if any logical function missing')
    args = ap.parse_args()

    in_path = pathlib.Path(args.input)
    if not in_path.exists():
        print(f"ERROR: input file not found: {in_path}", file=sys.stderr)
        sys.exit(1)

    buckets, missing = parse(in_path)
    if args.fail_on_missing and missing:
        print(f"ERROR: missing function coverage for: {', '.join(sorted(missing))}", file=sys.stderr)
        sys.exit(1)

    snapshot = {
        "generated_at": datetime.datetime.utcnow().isoformat() + 'Z',
        "functions": summarize(buckets)
    }

    out_path = pathlib.Path(args.out)
    out_path.write_text(json.dumps(snapshot, indent=2) + '\n')
    print(f"Wrote snapshot -> {out_path}")

if __name__ == '__main__':
    main()
