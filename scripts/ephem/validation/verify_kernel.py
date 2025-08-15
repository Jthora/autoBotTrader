#!/usr/bin/env python3
"""Verify ephemeris kernel integrity against kernel_hashes.json.

Usage:
    python scripts/ephem/validation/verify_kernel.py --kernel path/to/de440s.bsp
    # (Optional) update placeholder entry to real hash (safe only if verified source)
    python scripts/ephem/validation/verify_kernel.py --kernel scripts/ephem/cache/de440s.bsp --update --hashes scripts/ephem/validation/kernel_hashes.json

Exits 0 if:
    - kernel hash matches expected OR
    - expected hash placeholder and --update not supplied (logs actual) OR
    - placeholder replaced successfully when using --update
Returns 1 on mismatch or missing kernel.
"""
from __future__ import annotations
import argparse
import hashlib
import json
from pathlib import Path

def sha256_file(path: Path) -> str:
    h = hashlib.sha256()
    with path.open('rb') as f:
        for chunk in iter(lambda: f.read(65536), b''):
            h.update(chunk)
    return h.hexdigest()

def main():
    p = argparse.ArgumentParser()
    p.add_argument('--kernel', required=True)
    p.add_argument('--hashes', default='scripts/ephem/validation/kernel_hashes.json')
    p.add_argument('--update', action='store_true', help='Replace placeholder hash in hashes file with computed value')
    args = p.parse_args()
    kernel_path = Path(args.kernel)
    if not kernel_path.exists():
        print(f'KERNEL_MISSING {kernel_path}')
        return 1
    computed = sha256_file(kernel_path)
    data = json.loads(Path(args.hashes).read_text())
    expected_entry = next((k for k in data.get('kernels', []) if k.get('name') == kernel_path.name), None)
    if not expected_entry:
        print(f'KERNEL_UNLISTED name={kernel_path.name} sha256={computed}')
        return 1
    expected = expected_entry.get('sha256')
    if expected == 'REPLACE_WITH_ACTUAL_SHA256':
        if args.update:
            expected_entry['sha256'] = computed
            Path(args.hashes).write_text(json.dumps(data, indent=2) + '\n')
            print(f'KERNEL_PLACEHOLDER_UPDATED sha256={computed}')
        else:
            print(f'KERNEL_PLACEHOLDER sha256={computed}')
        return 0
    if expected.lower() == computed.lower():
        print(f'KERNEL_OK sha256={computed}')
        return 0
    print(f'KERNEL_HASH_MISMATCH expected={expected} got={computed}')
    return 1

if __name__ == '__main__':
    raise SystemExit(main())
