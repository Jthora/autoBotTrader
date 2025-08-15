#!/usr/bin/env python3
"""Parse coverage reports and inject into metrics.json.
Supports:
  - Go: coverage.out (placeholder parsing)
  - Frontend: coverage/coverage-summary.json (Jest)
"""
from __future__ import annotations
from pathlib import Path
import json
from datetime import datetime, timezone

ROOT = Path(__file__).resolve().parents[2]
METRICS_PATH = ROOT / "docs/progress/metrics.json"

def update_go(metrics):
    cov_file = ROOT / 'coverage.out'
    if not cov_file.exists():
        return
    # Placeholder: assign dummy 0 (future integration)
    metrics.setdefault('go_service', {}).setdefault('coverage_pct', 0)

def update_frontend(metrics):
    summary = ROOT / 'coverage' / 'coverage-summary.json'
    if not summary.exists():
        return
    data = json.loads(summary.read_text())
    total = data.get('total', {}).get('lines', {})
    pct = total.get('pct')
    if pct is not None:
        metrics.setdefault('frontend', {})['coverage_pct'] = pct

def main():
    if METRICS_PATH.exists():
        metrics = json.loads(METRICS_PATH.read_text())
    else:
        metrics = {"version": 1}
    update_go(metrics)
    update_frontend(metrics)
    metrics['last_updated'] = datetime.now(timezone.utc).isoformat()
    METRICS_PATH.write_text(json.dumps(metrics, indent=2) + '\n')

if __name__ == '__main__':
    main()
