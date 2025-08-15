#!/usr/bin/env python3
"""Compare gas snapshot against baseline.

Baseline file: docs/perf/gas_baselines.json
Snapshot file: docs/perf/gas_snapshot.json (produced in CI)

Exit codes: 0 ok / warn, 1 violation, 2 parsing error.
"""
from __future__ import annotations
import json
from pathlib import Path

def load_json(path: Path):
    return json.loads(path.read_text()) if path.exists() else None

def main():
    baseline_path = Path('docs/perf/gas_baselines.json')
    snap_path = Path('docs/perf/gas_snapshot.json')
    base = load_json(baseline_path)
    snap = load_json(snap_path)
    if not base or not snap:
        print('MISSING_BASE_OR_SNAPSHOT')
        return 2
    violations = []
    warn_only = []
    thresh_pct = base.get('threshold_percent', 5.0)
    for item in base.get('functions', []):
        name = item['name']
        base_gas = item['avg_gas']
        snap_entry = next((f for f in snap.get('functions', []) if f['name'] == name), None)
        if not snap_entry:
            warn_only.append(f'missing function in snapshot: {name}')
            continue
        cur = snap_entry['avg_gas']
        if base_gas == 0:
            continue
        delta_pct = (cur - base_gas) / base_gas * 100.0
        if delta_pct > thresh_pct:
            violations.append(f'{name} gas regression {delta_pct:.2f}% > {thresh_pct:.2f}% (base={base_gas} cur={cur})')
    result = {
        'violations': violations,
        'warnings': warn_only,
        'threshold_percent': thresh_pct,
    }
    print(json.dumps(result, indent=2))
    return 1 if violations else 0

if __name__ == '__main__':
    raise SystemExit(main())
