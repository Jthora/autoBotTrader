#!/usr/bin/env python3
"""Fetch & cache JPL ephemeris kernel with integrity verification.

Usage:
  python scripts/ephem/validation/fetch_kernel.py \
      --name de440s.bsp \
      --cache-dir scripts/ephem/cache \
      --hashes scripts/ephem/validation/kernel_hashes.json

Behavior:
  - Reads expected hash + source URL from kernel_hashes.json (matching name).
  - Creates cache dir if missing.
  - If kernel already present and hash matches -> exit 0 (cached OK).
  - If present but hash mismatch -> report and exit 1.
  - If missing: downloads via urllib; on success verifies hash. If expected hash is placeholder,
    prints computed hash so maintainer can update kernel_hashes.json (no auto-edit to keep governance explicit).

No external deps beyond stdlib.
"""
from __future__ import annotations
import argparse
import hashlib
import json
import sys
import urllib.request
from pathlib import Path

def sha256_file(p: Path) -> str:
    h = hashlib.sha256()
    with p.open('rb') as f:
        for chunk in iter(lambda: f.read(65536), b''):
            h.update(chunk)
    return h.hexdigest()

def download(url: str, dest: Path):
    with urllib.request.urlopen(url) as resp:  # nosec B310 (trusted NASA host defined in hashes file)
        dest.write_bytes(resp.read())

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--name', default='de440s.bsp')
    ap.add_argument('--cache-dir', default='scripts/ephem/cache')
    ap.add_argument('--hashes', default='scripts/ephem/validation/kernel_hashes.json')
    ap.add_argument('--force', action='store_true', help='Re-download even if file exists')
    args = ap.parse_args()

    cache_dir = Path(args.cache_dir)
    cache_dir.mkdir(parents=True, exist_ok=True)
    hashes_path = Path(args.hashes)
    if not hashes_path.exists():
        print(f'ERROR missing hashes file: {hashes_path}', file=sys.stderr)
        return 1
    data = json.loads(hashes_path.read_text())
    entry = next((k for k in data.get('kernels', []) if k.get('name') == args.name), None)
    if not entry:
        print(f'ERROR kernel name not listed in hashes file: {args.name}', file=sys.stderr)
        return 1
    expected = entry.get('sha256', '')
    url = entry.get('source')
    target = cache_dir / args.name
    if target.exists() and not args.force:
        actual = sha256_file(target)
        if expected == 'REPLACE_WITH_ACTUAL_SHA256':
            print(f'KERNEL_CACHED_PLACEHOLDER actual_sha256={actual} path={target}')
            return 0
        if actual.lower() == expected.lower():
            print(f'KERNEL_CACHED_OK sha256={actual} path={target}')
            return 0
        print(f'KERNEL_HASH_MISMATCH cached_sha256={actual} expected={expected} path={target}', file=sys.stderr)
        return 1
    # Download path
    try:
        print(f'DOWNLOADING {url} -> {target}')
        download(url, target)
    except Exception as e:  # pragma: no cover (network failure)
        print(f'ERROR download failed: {e}', file=sys.stderr)
        return 1
    actual = sha256_file(target)
    if expected == 'REPLACE_WITH_ACTUAL_SHA256':
        print(f'KERNEL_DOWNLOADED_PLACEHOLDER actual_sha256={actual} path={target}')
        return 0
    if actual.lower() == expected.lower():
        print(f'KERNEL_DOWNLOADED_OK sha256={actual} path={target}')
        return 0
    print(f'KERNEL_HASH_MISMATCH_AFTER_DOWNLOAD expected={expected} got={actual}', file=sys.stderr)
    return 1

if __name__ == '__main__':  # pragma: no cover
    raise SystemExit(main())
