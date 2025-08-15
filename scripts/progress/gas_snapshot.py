#!/usr/bin/env python3
"""Generate gas snapshot (placeholder extraction with governance enforcement).

Writes docs/perf/gas_snapshot.json with average gas per function.
If any avg_gas values are zero (placeholders) and environment variable
ALLOW_GAS_PLACEHOLDERS is not set to a truthy value, exits non-zero to
enforce replacing placeholders in CI.
"""
from pathlib import Path
import json
from datetime import datetime, timezone
import os
import re
import subprocess

# Attempt JSON parsing if future snforge provides; fallback regex currently.

def run_contract_tests_capture() -> str:
    """Run scarb test (or snforge) capturing gas output. Returns raw stdout.

    Uses environment GAS_TEST_CMD override for testing (e.g., cat fixture file).
    """
    override = os.getenv('GAS_TEST_CMD')
    if override:
        # Simple shell execution (no arguments splitting for brevity)
        proc = subprocess.run(override, shell=True, capture_output=True, text=True)
    else:
        proc = subprocess.run(['scarb', 'test', '-q'], capture_output=True, text=True)
    if proc.returncode != 0:
        # Include snippet for debugging
        snippet = proc.stdout.splitlines()[-5:]
        raise RuntimeError(f'test run failed code={proc.returncode} tail="{' | '.join(snippet)}"')
    return proc.stdout

FUNC_ALLOWLIST = {
    'set_prediction_inputs': 'update_prediction',
    'execute_trade': 'execute_trade',
}

GAS_LINE_RE = re.compile(r"^(?P<fn>set_prediction_inputs|execute_trade)\b.*?gas(?:=|:)?(?P<gas>\d+)")

def extract_gas(raw: str) -> dict:
    buckets: dict[str, list[int]] = {v: [] for v in FUNC_ALLOWLIST.values()}
    for line in raw.splitlines():
        m = GAS_LINE_RE.search(line)
        if not m:
            continue
        fn = m.group('fn')
        gas = int(m.group('gas'))
        logical = FUNC_ALLOWLIST.get(fn)
        if logical:
            buckets[logical].append(gas)
    results = {}
    for logical, vals in buckets.items():
        if vals:
            results[logical] = sum(vals) // len(vals)
        else:
            results[logical] = 0
    return results

ROOT = Path(__file__).resolve().parents[2]
SNAP_PATH = ROOT / 'docs' / 'perf' / 'gas_snapshot.json'

FUNCTIONS = ['update_prediction', 'execute_trade']

def main():
    # Extraction phase
    try:
        raw = run_contract_tests_capture()
        gas_map = extract_gas(raw)
    except Exception as e:
        print(f'EXTRACTION_ERROR {e}')
        gas_map = {f:0 for f in FUNCTIONS}

    snapshot = {
        'generated_at': datetime.now(timezone.utc).isoformat(),
        'functions': [ { 'name': f, 'avg_gas': gas_map.get(f, 0) } for f in FUNCTIONS ]
    }
    SNAP_PATH.write_text(json.dumps(snapshot, indent=2) + '\n')
    print(f'wrote gas snapshot -> {SNAP_PATH}')
    allow = os.getenv('ALLOW_GAS_PLACEHOLDERS', '').lower() in ('1','true','yes')
    missing = [f for f,v in gas_map.items() if v == 0]
    if not allow and missing:
        print(f'PLACEHOLDER_GAS_VALUES: {", ".join(missing)}')
        return 1
    return 0

if __name__ == '__main__':
    raise SystemExit(main())
