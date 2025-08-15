#!/usr/bin/env python3
"""Detect stagnation (no metrics change in N days)."""
from pathlib import Path
import json
from datetime import datetime, timezone

ROOT = Path(__file__).resolve().parents[2]
METRICS_PATH = ROOT / "docs/progress/metrics.json"
THRESHOLD_DAYS = 3

def main():
    if not METRICS_PATH.exists():
        print("No metrics file; cannot evaluate stagnation.")
        return
    data = json.loads(METRICS_PATH.read_text())
    ts = data.get("last_updated")
    if not ts:
        print("Missing last_updated in metrics.json")
        return
    last = datetime.fromisoformat(ts.replace('Z','+00:00'))
    delta_days = (datetime.now(timezone.utc) - last).days
    if delta_days >= THRESHOLD_DAYS:
        print(f"STAGNATION: last update {delta_days} days ago")
        exit(1)
    print(f"Active: last update {delta_days} days ago (< {THRESHOLD_DAYS})")

if __name__ == '__main__':
    main()
