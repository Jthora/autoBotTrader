#!/usr/bin/env python3
"""Compare Go benchmark output against baseline JSON.

Usage:
  python scripts/ci/compare_bench.py --input bench.txt --baseline docs/perf/bench_baselines.json [--out result.json]

Exit Codes:
  0 success (no violations)
  1 regression threshold exceeded
  2 usage / parsing error

Benchmark Lines Expected (go test -bench output subset):
BenchmarkGTAB_Lookup-8      451 ns/op    0 B/op   0 allocs/op

We parse name, ns/op, B/op, allocs/op.
"""
from __future__ import annotations
import argparse
import json
import re
import sys
import pathlib

def parse_args():
    p = argparse.ArgumentParser()
    p.add_argument('--input', required=True, help='Path to go benchmark output text file')
    p.add_argument('--baseline', required=True, help='Path to baseline JSON (bench_baselines.json)')
    p.add_argument('--out', help='Optional path to write structured comparison result JSON')
    p.add_argument('--factor', type=float, default=2.0, help='Hard fail multiplicative factor threshold (default 2.0)')
    p.add_argument('--abs', type=float, default=500.0, help='Hard fail absolute ns threshold (default 500)')
    p.add_argument('--warn', type=float, default=0.2, help='Soft warning factor (e.g. 0.2 = +20%)')
    return p.parse_args()

LINE_RE = re.compile(r'^(Benchmark\S+)\s+(\d+)\s+ns/op\s+(\d+)\s+B/op\s+(\d+)\s+allocs/op')

def parse_bench(text: str):
    results = {}
    for line in text.splitlines():
        m = LINE_RE.search(line.strip())
        if not m:
            continue
        name, ns, b, allocs = m.groups()
        results[name] = {
            'ns_per_op': int(ns),
            'bytes_per_op': int(b),
            'allocs_per_op': int(allocs),
        }
    return results

def load_baseline(path: str):
    data = json.loads(pathlib.Path(path).read_text())
    base = { b['name']: b for b in data.get('benchmarks', []) }
    return data, base

def compare(current, baseline, factor_thresh: float, abs_thresh: float, warn_factor: float):
    report = []
    violations = False
    warnings = False
    for name, base in baseline.items():
        cur = current.get(name)
        if not cur:
            report.append({'name': name, 'status': 'missing', 'message': 'Benchmark not present in current run'})
            warnings = True
            continue
        ns_cur = cur['ns_per_op']
        ns_base = base['ns_per_op']
        ratio = ns_cur / ns_base if ns_base else float('inf')
        delta = ns_cur - ns_base
        alloc_delta = cur['allocs_per_op'] - base.get('allocs_per_op', 0)
        status = 'ok'
        msg = ''
        if ratio > factor_thresh or delta > abs_thresh:
            status = 'fail'
            msg = f'Performance regression: {ns_cur} ns/op vs baseline {ns_base} (ratio {ratio:.2f}, delta {delta} > {abs_thresh}?)'
            violations = True
        elif ratio > (1 + warn_factor):
            status = 'warn'
            msg = f'Significant slowdown: {ns_cur} ns/op vs {ns_base} (ratio {ratio:.2f})'
            warnings = True
        if alloc_delta > 0:
            if status == 'ok':
                status = 'warn'
                warnings = True
            msg += f' alloc regression: +{alloc_delta} allocs/op'
        report.append({
            'name': name,
            'baseline_ns': ns_base,
            'current_ns': ns_cur,
            'ratio': ratio,
            'delta_ns': delta,
            'alloc_delta': alloc_delta,
            'status': status,
            'message': msg.strip()
        })
    return report, violations, warnings

def main():
    args = parse_args()
    try:
        bench_text = pathlib.Path(args.input).read_text()
    except Exception as e:
        print(f'ERR reading benchmark input: {e}', file=sys.stderr)
        return 2
    current = parse_bench(bench_text)
    if not current:
        print('ERR no benchmarks parsed', file=sys.stderr)
        return 2
    try:
        baseline_doc, baseline = load_baseline(args.baseline)
    except Exception as e:
        print(f'ERR loading baseline: {e}', file=sys.stderr)
        return 2
    report, violations, warnings = compare(current, baseline, args.factor, args.abs, args.warn)
    out_obj = {
        'violations': violations,
        'warnings': warnings,
        'factor_threshold': args.factor,
        'abs_threshold_ns': args.abs,
        'warn_factor': args.warn,
        'results': report
    }
    if args.out:
        pathlib.Path(args.out).write_text(json.dumps(out_obj, indent=2))
    print(json.dumps(out_obj, indent=2))
    return 1 if violations else 0

if __name__ == '__main__':
    sys.exit(main())
