#!/usr/bin/env python3
"""
GTAB generator (v1) using Skyfield + JPL DE kernels.
Writes: gtab_1s.bin, gtab_60s.bin, gtab.meta.json

Note: Requires local kernel file (e.g., de440s.bsp). Do not commit kernels.
"""
from __future__ import annotations
import argparse
import hashlib
import json
import struct
from dataclasses import dataclass
from datetime import datetime, timezone, timedelta
from pathlib import Path
from typing import Tuple, List

try:
    from skyfield.api import load  # type: ignore
except Exception:
    load = None  # type: ignore

MAGIC = b"GTAB1"
VERSION = 1
FIELD_TIDE_BPS = 0x01
FIELD_TIDE_RAW = 0x02

MU_SUN = 1.32712440018e11  # km^3/s^2
MU_MOON = 4.9048695e3      # km^3/s^2

@dataclass
class Config:
    kernel: str
    start: datetime
    end: datetime
    out_dir: Path
    fields_mask: int = FIELD_TIDE_BPS | FIELD_TIDE_RAW


def percentiles(values: List[float], p: float) -> float:
    if not values:
        return 0.0
    s = sorted(values)
    k = (len(s)-1) * p
    f = int(k)
    c = min(f+1, len(s)-1)
    if f == c:
        return s[f]
    d0 = s[f] * (c - k)
    d1 = s[c] * (k - f)
    return d0 + d1


def make_dataset_id(cfg: Config) -> str:
    h = hashlib.sha256()
    h.update(Path(cfg.kernel).name.encode())
    h.update(str(cfg.start).encode())
    h.update(str(cfg.end).encode())
    h.update(str(cfg.fields_mask).encode())
    h.update(b"gtab_v1")
    return h.hexdigest()[:12]


def write_gtab(path: Path, epoch: int, dt_ns: int, samples: List[Tuple[int, float]], fields_mask: int):
    # Header
    hdr = bytearray()
    hdr.extend(MAGIC)
    hdr.extend(struct.pack('<H', VERSION))
    hdr.extend(struct.pack('<q', epoch))
    hdr.extend(struct.pack('<q', dt_ns))
    hdr.extend(struct.pack('<I', len(samples)))
    hdr.extend(struct.pack('<I', fields_mask))
    hdr.extend(b"\x00" * 16)  # reserved
    # Record packing in fixed order
    recs = bytearray()
    for bps, raw in samples:
        if fields_mask & FIELD_TIDE_BPS:
            recs.extend(struct.pack('<H', bps))
        if fields_mask & FIELD_TIDE_RAW:
            recs.extend(struct.pack('<f', raw))
    path.write_bytes(hdr + recs)


def generate(cfg: Config):
    if load is None:
        raise SystemExit("Missing skyfield. pip install skyfield jplephem numpy")
    ts = load.timescale()
    ker = load(cfg.kernel)
    earth = ker['earth']
    sun = ker['sun']
    moon = ker['moon']

    def compute_series(dt_seconds: int):
        times: List[datetime] = []
        raw_vals: List[float] = []
        t = cfg.start
        while t <= cfg.end:
            times.append(t)
            t += timedelta(seconds=dt_seconds)
        if not times:
            return [], 0, 0
        # skyfield Time object
        sf_t = ts.from_datetimes(times)
        # Geocentric vectors
        r_sun = (sun - earth).at(sf_t).position.km  # shape (3, N)
        r_moon = (moon - earth).at(sf_t).position.km
        # Distances
        import numpy as np  # type: ignore
        rs = np.linalg.norm(r_sun, axis=0)
        rm = np.linalg.norm(r_moon, axis=0)
        # raw scalar mu/r^3
        raw = MU_SUN/(rs**3) + MU_MOON/(rm**3)
        raw_vals = raw.tolist()
        # Normalize via percentiles p5->0, p95->100
        p5 = percentiles(raw_vals, 0.05)
        p95 = percentiles(raw_vals, 0.95)
        span = max(p95 - p5, 1e-12)
        samples: List[Tuple[int, float]] = []
        for v in raw_vals:
            bp = int(max(0.0, min(10000.0, (v - p5) / span * 10000.0)))
            samples.append((bp, float(v)))
        epoch = int(cfg.start.replace(tzinfo=timezone.utc).timestamp())
        return samples, epoch, int(dt_seconds * 1_000_000_000)

    # 1s cadence
    s1, epoch, dt1 = compute_series(1)
    # 60s cadence
    s60, _, dt60 = compute_series(60)

    cfg.out_dir.mkdir(parents=True, exist_ok=True)
    (cfg.out_dir / 'gtab_1s.bin').unlink(missing_ok=True)
    (cfg.out_dir / 'gtab_60s.bin').unlink(missing_ok=True)

    write_gtab(cfg.out_dir / 'gtab_1s.bin', epoch, dt1, s1, cfg.fields_mask)
    write_gtab(cfg.out_dir / 'gtab_60s.bin', epoch, dt60, s60, cfg.fields_mask)

    meta = {
        "dataset_id": make_dataset_id(cfg),
        "kernel": Path(cfg.kernel).name,
        "start": cfg.start.astimezone(timezone.utc).isoformat(),
        "end": cfg.end.astimezone(timezone.utc).isoformat(),
        "fields_mask": cfg.fields_mask,
        "version": VERSION,
        "created_at": datetime.now(timezone.utc).isoformat(),
    }
    (cfg.out_dir / 'gtab.meta.json').write_text(json.dumps(meta, indent=2) + "\n")


def parse_args():
    p = argparse.ArgumentParser()
    p.add_argument('--kernel', required=True, help='Path to local JPL DE kernel (e.g., de440s.bsp)')
    p.add_argument('--start', required=True, help='ISO start, e.g., 2025-08-01T00:00:00Z')
    p.add_argument('--end', required=True, help='ISO end, e.g., 2025-11-01T00:00:00Z')
    p.add_argument('--out-dir', default='ephem', help='Output directory')
    return p.parse_args()


def main():
    args = parse_args()
    start = datetime.fromisoformat(args.start.replace('Z', '+00:00'))
    end = datetime.fromisoformat(args.end.replace('Z', '+00:00'))
    cfg = Config(kernel=args.kernel, start=start, end=end, out_dir=Path(args.out_dir))
    generate(cfg)
    print(f"Wrote GTAB to {cfg.out_dir}")


if __name__ == '__main__':
    main()
