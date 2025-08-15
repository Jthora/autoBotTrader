#!/usr/bin/env python3
"""Compare current gas snapshot against baselines.

Exit codes:
 0 ok within thresholds
 1 regression detected
 2 structural / parse error
"""
from __future__ import annotations
import json
import sys
import pathlib


def load(path: str):
    try:
        return json.loads(pathlib.Path(path).read_text())
    except Exception as e:  # noqa: BLE001
        print(f"ERROR: failed to load {path}: {e}", file=sys.stderr)
        sys.exit(2)


def main():
    baseline = load('docs/perf/gas_baselines.json')
    snapshot = load('docs/perf/gas_snapshot.json')

    threshold = baseline.get('threshold_percent', 5.0)
    base_map = {f['name']: f['avg_gas'] for f in baseline['functions']}
    snap_map = {f['name']: f['avg_gas'] for f in snapshot['functions']}

    failed = []
    for name, base_val in base_map.items():
        cur = snap_map.get(name)
        if cur is None:
            failed.append((name, 'MISSING_SNAPSHOT', base_val, None))
            continue
        if base_val == 0:
            # treat first non-zero fill as establishing baseline
            continue
        if cur > base_val:
            pct = (cur - base_val) * 100.0 / max(1, base_val)
            if pct > threshold:
                failed.append((name, f"{pct:.2f}% > {threshold}%", base_val, cur))
    if failed:
        print('GAS REGRESSIONS:')
        for (name, reason, base, cur) in failed:
            print(f"  {name}: base={base} cur={cur} ({reason})")
        sys.exit(1)
    print('Gas comparison OK.')

if __name__ == '__main__':
    main()
