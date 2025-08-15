#!/usr/bin/env python3
"""Ephemeris accuracy harness.

Computes error metrics comparing a GTAB binary table against authoritative
astronomical values (Sun + Moon gravitational influence proxy) over a sampled
time window.

Outputs a JSON metrics document conforming (loosely) to metrics.schema.json.

Exit codes:
 0 success (metrics computed; passes flag may be true/false)
 2 usage / recoverable error (bad args, missing file)

NOTE: This implementation intentionally keeps dependencies minimal (skyfield,
numpy). Heavy analytics (e.g., pandas) avoided for CI runtime.
"""
from __future__ import annotations
import argparse
import json
import math
import struct
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import List, Dict, Any

try:
    from skyfield.api import load  # type: ignore
except Exception:  # pragma: no cover - runtime guard
    load = None  # type: ignore

MAGIC = b"GTAB1"
MU_SUN = 1.32712440018e11  # km^3/s^2
MU_MOON = 4.9048695e3      # km^3/s^2

@dataclass
class Thresholds:
    rel_error_median: float = 1.0e-3
    rel_error_p99: float = 5.0e-3
    peak_timing_drift_seconds_p99: float = 120.0
    peak_value_rel_drift_p99: float = 2.0e-3

    def to_dict(self) -> Dict[str, float]:
        return {
            'rel_error_median': self.rel_error_median,
            'rel_error_p99': self.rel_error_p99,
            'peak_timing_drift_seconds_p99': self.peak_timing_drift_seconds_p99,
            'peak_value_rel_drift_p99': self.peak_value_rel_drift_p99,
        }

def parse_args():
    p = argparse.ArgumentParser()
    p.add_argument('--table', required=True, help='Path to GTAB binary (v1)')
    p.add_argument('--dataset-id', required=True, help='Dataset identifier label')
    p.add_argument('--start', required=False, help='ISO start (default: NOW-7d)')
    p.add_argument('--end', required=False, help='ISO end (default: NOW)')
    p.add_argument('--cadence', type=int, default=60, help='Sampling cadence seconds (default 60)')
    p.add_argument('--thresholds', help='Optional JSON file overriding thresholds')
    p.add_argument('--out', required=True, help='Output metrics JSON path')
    p.add_argument('--kernel', help='Path to JPL DE kernel (e.g. de440s.bsp); if omitted attempt environment GTAB_KERNEL or fallback small kernel name')
    p.add_argument('--verbose', action='store_true')
    return p.parse_args()

def iso_parse(s: str) -> datetime:
    return datetime.fromisoformat(s.replace('Z', '+00:00')).astimezone(timezone.utc)

def load_thresholds(path: str | None) -> Thresholds:
    if not path:
        return Thresholds()
    data = json.loads(Path(path).read_text())
    t = Thresholds(
        rel_error_median=data.get('rel_error_median', 1.0e-3),
        rel_error_p99=data.get('rel_error_p99', 5.0e-3),
        peak_timing_drift_seconds_p99=data.get('peak_timing_drift_seconds_p99', 120.0),
        peak_value_rel_drift_p99=data.get('peak_value_rel_drift_p99', 2.0e-3),
    )
    return t

def read_gtab_header(path: Path):
    with path.open('rb') as f:
        hdr = f.read(5+2+8+8+4+4+16)
        if len(hdr) < 5 or hdr[:5] != MAGIC:
            raise ValueError('invalid magic')
        ver = struct.unpack('<H', hdr[5:7])[0]
        if ver != 1:
            raise ValueError(f'unsupported version {ver}')
        epoch = struct.unpack('<q', hdr[7:15])[0]
        dt_ns = struct.unpack('<q', hdr[15:23])[0]
        n = struct.unpack('<I', hdr[23:27])[0]
        fields_mask = struct.unpack('<I', hdr[27:31])[0]
    if dt_ns <= 0 or n == 0:
        raise ValueError('invalid header timing fields')
    return epoch, dt_ns, n, fields_mask

def gtab_lookup_tide_bps(path: Path, epoch: int, dt_ns: int, n: int, t: datetime) -> int:
    dt = (t - datetime.fromtimestamp(epoch, tz=timezone.utc)).total_seconds()
    pos = dt * 1e9 / dt_ns
    if pos < 0 or pos > (n - 1):
        raise IndexError('out of range')
    i = int(pos)
    frac = pos - i
    with path.open('rb') as f:
        header_size = 5+2+8+8+4+4+16
        rec_size = 2
        def read_u16(idx: int) -> int:
            off = header_size + idx * rec_size
            f.seek(off)
            b = f.read(2)
            return struct.unpack('<H', b)[0]
        if frac == 0:
            return read_u16(i)
        v0 = read_u16(i)
        v1 = read_u16(min(i+1, n-1))
    return int(v0 + (v1 - v0) * frac + 0.5)

def percentile(sorted_vals: List[float], p: float) -> float:
    if not sorted_vals:
        return 0.0
    k = (len(sorted_vals)-1) * p
    f = math.floor(k)
    c = min(f+1, len(sorted_vals)-1)
    if f == c:
        return sorted_vals[f]
    return sorted_vals[f]*(c-k) + sorted_vals[c]*(k-f)

def compute_authoritative(times, ts, sun, earth, moon):
    sf_t = ts.from_datetimes(times)
    r_sun = (sun - earth).at(sf_t).position.km
    r_moon = (moon - earth).at(sf_t).position.km
    rs = (r_sun[0]**2 + r_sun[1]**2 + r_sun[2]**2) ** 0.5
    rm = (r_moon[0]**2 + r_moon[1]**2 + r_moon[2]**2) ** 0.5
    return MU_SUN/(rs**3) + MU_MOON/(rm**3)

def detect_peaks(vals: List[float]) -> List[int]:
    peaks = []
    for i in range(1, len(vals)-1):
        if (vals[i] > vals[i-1] and vals[i] > vals[i+1]) or (vals[i] < vals[i-1] and vals[i] < vals[i+1]):
            peaks.append(i)
    return peaks

def main():
    args = parse_args()
    tbl_path = Path(args.table)
    now = datetime.now(timezone.utc)
    start = iso_parse(args.start) if args.start else now - timedelta(days=7)
    end = iso_parse(args.end) if args.end else now
    if end <= start:
        raise SystemExit('end must be > start')
    cadence = max(1, args.cadence)
    thresholds = load_thresholds(args.thresholds)
    try:
        epoch_sec, dt_ns, n, _ = read_gtab_header(tbl_path)
    except Exception as e:
        metrics = {
            'dataset_id': args.dataset_id,
            'generated_at': now.isoformat(),
            'records_compared': 0,
            'passes': False,
            'reason': f'bad_header:{e}',
            'thresholds': thresholds.to_dict(),
        }
        Path(args.out).write_text(json.dumps(metrics, indent=2) + '\n')
        return 0
    tbl_start = datetime.fromtimestamp(epoch_sec, tz=timezone.utc)
    tbl_end = tbl_start + timedelta(seconds=(n-1) * dt_ns / 1e9)
    if start < tbl_start:
        start = tbl_start
    if end > tbl_end:
        end = tbl_end
    if end <= start:
        metrics = {
            'dataset_id': args.dataset_id,
            'generated_at': now.isoformat(),
            'records_compared': 0,
            'passes': False,
            'reason': 'coverage_gap',
            'thresholds': thresholds.to_dict(),
        }
        Path(args.out).write_text(json.dumps(metrics, indent=2) + '\n')
        return 0
    times: List[datetime] = []
    t = start
    while t <= end:
        times.append(t)
        t += timedelta(seconds=cadence)
    if len(times) < 10:
        metrics = {
            'dataset_id': args.dataset_id,
            'generated_at': now.isoformat(),
            'records_compared': len(times),
            'passes': False,
            'reason': 'insufficient_samples',
            'thresholds': thresholds.to_dict(),
        }
        Path(args.out).write_text(json.dumps(metrics, indent=2) + '\n')
        return 0
    if load is None:
        raise SystemExit('skyfield not installed')
    ker_path = args.kernel or Path.cwd() / 'de440s.bsp'
    if not Path(ker_path).exists():
        raise SystemExit(f'kernel not found: {ker_path}')
    ts = load.timescale()
    ker = load(str(ker_path))
    earth = ker['earth']
    sun = ker['sun']
    moon = ker['moon']
    authoritative = compute_authoritative(times, ts, sun, earth, moon)
    auth_list = authoritative.tolist()
    gtab_vals: List[int] = []
    for tsample in times:
        try:
            gtab_vals.append(gtab_lookup_tide_bps(tbl_path, epoch_sec, dt_ns, n, tsample))
        except Exception:
            gtab_vals.append(0)
    sorted_auth = sorted(auth_list)
    p5 = percentile(sorted_auth, 0.05)
    p95 = percentile(sorted_auth, 0.95)
    span = p95 - p5 if (p95 - p5) > 1e-12 else 1e-12
    auth_bps: List[int] = []
    for v in auth_list:
        bp = int(max(0.0, min(10000.0, (v - p5) / span * 10000.0)))
        auth_bps.append(bp)
    rel_errors: List[float] = []
    abs_errors: List[float] = []
    for i in range(1, len(times)-1):
        a = auth_bps[i]
        g = gtab_vals[i]
        abs_e = abs(g - a)
        rel_e = abs_e / max(1, a)
        abs_errors.append(abs_e)
        rel_errors.append(rel_e)
    def stats_summary(vals: List[float]) -> Dict[str, float]:
        if not vals:
            return {'mean': 0.0, 'median': 0.0, 'p95': 0.0, 'p99': 0.0, 'max': 0.0}
        s = sorted(vals)
        mean = sum(s)/len(s)
        def pct(p: float) -> float:
            return percentile(s, p)
        return {
            'mean': mean,
            'median': pct(0.5),
            'p95': pct(0.95),
            'p99': pct(0.99),
            'max': s[-1],
        }
    rel_stats = stats_summary(rel_errors)
    abs_stats = stats_summary(abs_errors)
    def detect_peaks(vals: List[int]) -> List[int]:
        peaks = []
        for i in range(1, len(vals)-1):
            if (vals[i] > vals[i-1] and vals[i] > vals[i+1]) or (vals[i] < vals[i-1] and vals[i] < vals[i+1]):
                peaks.append(i)
        return peaks
    peaks_auth = detect_peaks(auth_bps)
    peaks_gtab = detect_peaks(gtab_vals)
    peak_timing_drifts: List[float] = []
    peak_value_rel_drifts: List[float] = []
    for pa in peaks_auth:
        ta = times[pa]
        nearest_idx = None
        nearest_dt = None
        for pg in peaks_gtab:
            dt_seconds = abs((times[pg] - ta).total_seconds())
            if nearest_dt is None or dt_seconds < nearest_dt:
                nearest_dt = dt_seconds
                nearest_idx = pg
        if nearest_idx is not None and nearest_dt is not None and nearest_dt <= 2 * cadence:
            peak_timing_drifts.append(nearest_dt)
            av = auth_bps[pa]
            gv = gtab_vals[nearest_idx]
            peak_value_rel_drifts.append(abs(gv - av) / max(1, av))
    def stats_summary2(vals: List[float]) -> Dict[str, float]:
        if not vals:
            return {'mean': 0.0, 'median': 0.0, 'p95': 0.0, 'p99': 0.0, 'max': 0.0}
        s = sorted(vals)
        mean = sum(s)/len(s)
        return {
            'mean': mean,
            'median': percentile(s, 0.5),
            'p95': percentile(s, 0.95),
            'p99': percentile(s, 0.99),
            'max': s[-1],
        }
    timing_stats = stats_summary2(peak_timing_drifts)
    peak_val_stats = stats_summary2(peak_value_rel_drifts)
    passes = (
        rel_stats['median'] <= thresholds.rel_error_median and
        rel_stats['p99'] <= thresholds.rel_error_p99 and
        timing_stats['p99'] <= thresholds.peak_timing_drift_seconds_p99 and
        peak_val_stats['p99'] <= thresholds.peak_value_rel_drift_p99
    )
    metrics: Dict[str, Any] = {
        'dataset_id': args.dataset_id,
        'generated_at': now.isoformat(),
        'records_compared': len(times),
        'rel_error_mean': rel_stats['mean'],
        'rel_error_median': rel_stats['median'],
        'rel_error_p95': rel_stats['p95'],
        'rel_error_p99': rel_stats['p99'],
        'rel_error_max': rel_stats['max'],
        'abs_error_mean': abs_stats['mean'],
        'abs_error_p95': abs_stats['p95'],
        'abs_error_p99': abs_stats['p99'],
        'abs_error_max': abs_stats['max'],
        'peak_count': len(peaks_auth),
        'peak_timing_drift_seconds_mean': timing_stats['mean'],
        'peak_timing_drift_seconds_median': timing_stats['median'],
        'peak_timing_drift_seconds_p95': timing_stats['p95'],
        'peak_timing_drift_seconds_p99': timing_stats['p99'],
        'peak_timing_drift_seconds_max': timing_stats['max'],
        'peak_value_rel_drift_mean': peak_val_stats['mean'],
        'peak_value_rel_drift_p95': peak_val_stats['p95'],
        'peak_value_rel_drift_p99': peak_val_stats['p99'],
        'peak_value_rel_drift_max': peak_val_stats['max'],
        'passes': passes,
        'thresholds': thresholds.to_dict(),
    }
    Path(args.out).write_text(json.dumps(metrics, indent=2) + '\n')
    if args.verbose:
        print(json.dumps(metrics, indent=2))
    return 0

if __name__ == '__main__':  # pragma: no cover
    raise SystemExit(main())
