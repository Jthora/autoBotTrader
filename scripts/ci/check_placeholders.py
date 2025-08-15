#!/usr/bin/env python3
"""Fail if placeholder tokens remain in critical files (governance gate).

Scans for tokens: PLACEHOLDER, REPLACE_WITH_ACTUAL, REPLACE_WITH_ACTUAL_SHA256
unless ALLOW_PLACEHOLDERS=1.
"""
from __future__ import annotations
import os
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
TOKENS = ["PLACEHOLDER", "REPLACE_WITH_ACTUAL", "REPLACE_WITH_ACTUAL_SHA256"]
TARGETS = [
    ROOT / 'docs' / 'perf' / 'gas_snapshot.json',
    ROOT / 'scripts' / 'ephem' / 'validation' / 'kernel_hashes.json',
]

def main() -> int:
    allow = os.getenv('ALLOW_PLACEHOLDERS', '').lower() in ('1','true','yes')
    if allow:
        print('Placeholder scan skipped (ALLOW_PLACEHOLDERS set)')
        return 0
    failures = []
    for path in TARGETS:
        if not path.exists():
            continue
        text = path.read_text(errors='ignore')
        for tok in TOKENS:
            if tok in text:
                failures.append((path, tok))
    if failures:
        for p, tok in failures:
            print(f'PLACEHOLDER_FOUND {p} token={tok}')
        return 1
    print('No placeholders detected in target files.')
    return 0

if __name__ == '__main__':
    raise SystemExit(main())
