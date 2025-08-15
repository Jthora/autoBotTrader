#!/usr/bin/env python3
"""Progress automation utility.

Functions:
  - Validate task graph structure & dependency integrity
  - Recompute metrics.json task counts
  - List READY tasks (all deps COMPLETE, status PENDING)
  - (Future) Gate transitions & enforce acceptance criteria stubs

Usage:
  python scripts/progress/update_progress.py --summary
  python scripts/progress/update_progress.py --write
  python scripts/progress/update_progress.py --json > progress.json

Exit Codes:
  0 OK
  2 Validation error
"""

from __future__ import annotations

import argparse
import json
import os
import sys
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, List, Set

try:
    import yaml  # type: ignore
except ImportError:  # pragma: no cover
    print("Missing dependency pyyaml. Install via: pip install pyyaml", file=sys.stderr)
    sys.exit(1)

_override = os.environ.get("PROGRESS_ROOT")
if _override:
    ROOT = Path(_override).resolve()
else:
    ROOT = Path(__file__).resolve().parents[2]
TASK_GRAPH_PATH = ROOT / "docs/progress/task_graph.yaml"
METRICS_PATH = ROOT / "docs/progress/metrics.json"
STATUS_LOG_PATH = ROOT / "docs/progress/status_log.md"


@dataclass
class Task:
    id: str
    title: str
    phase: int
    priority: int
    depends_on: List[str]
    status: str
    raw: dict

    def is_ready(self, completed: Set[str]) -> bool:
        return self.status == "PENDING" and all(d in completed for d in self.depends_on)


VALID_STATUSES = {"PENDING", "IN_PROGRESS", "BLOCKED", "IN_REVIEW", "COMPLETE", "DEFERRED"}


def load_tasks() -> List[Task]:
    data = yaml.safe_load(TASK_GRAPH_PATH.read_text())
    tasks_raw = data.get("tasks", [])
    tasks: List[Task] = []
    for t in tasks_raw:
        # Harden against stray string entries accidentally appended (e.g. timestamp notes outside notes array)
        if not isinstance(t, dict):
            continue
        tasks.append(
            Task(
                id=t["id"],
                title=t.get("title", ""),
                phase=int(t.get("phase", 0)),
                priority=int(t.get("priority", 1000)),
                depends_on=t.get("depends_on", []) or [],
                status=t.get("status", "PENDING"),
                raw=t,
            )
        )
    return tasks


def validate(tasks: List[Task]) -> List[str]:
    errors: List[str] = []
    seen: Dict[str, int] = {}
    for t in tasks:
        seen[t.id] = seen.get(t.id, 0) + 1
    for tid, count in seen.items():
        if count > 1:
            errors.append(f"Duplicate task id detected: {tid}")
    ids = set(seen.keys())
    for t in tasks:
        if t.status not in VALID_STATUSES:
            errors.append(f"Task {t.id}: invalid status {t.status}")
        for dep in t.depends_on:
            if dep not in ids:
                errors.append(f"Task {t.id}: missing dependency {dep}")
    graph = {t.id: set(t.depends_on) for t in tasks}
    visited: Set[str] = set()
    stack: Set[str] = set()

    def dfs(node: str):
        if node in stack:
            errors.append(f"Cycle detected involving {node}")
            return
        if node in visited:
            return
        visited.add(node)
        stack.add(node)
        for nxt in graph.get(node, set()):
            dfs(nxt)
        stack.remove(node)

    for node in graph:
        dfs(node)
    return errors


def recompute_metrics(tasks: List[Task], metrics: Dict) -> Dict:
    status_counts = {s: 0 for s in VALID_STATUSES}
    for t in tasks:
        if t.status in status_counts:
            status_counts[t.status] += 1
    metrics.setdefault("tasks", {})
    metrics["tasks"].update(
        {
            "total": len(tasks),
            "complete": status_counts["COMPLETE"],
            "in_progress": status_counts["IN_PROGRESS"],
            "blocked": status_counts["BLOCKED"],
        }
    )
    metrics["last_updated"] = datetime.now(timezone.utc).isoformat()
    return metrics


def load_metrics() -> Dict:
    if METRICS_PATH.exists():
        return json.loads(METRICS_PATH.read_text())
    return {"version": 1}


def summarize(tasks: List[Task]) -> Dict:
    completed = {t.id for t in tasks if t.status == "COMPLETE"}
    ready = sorted(
        [t for t in tasks if t.is_ready(completed)],
        key=lambda x: (x.priority, x.phase, x.id),
    )
    blocked = [t for t in tasks if t.status == "BLOCKED"]
    return {
        "ready": [t.id for t in ready],
        "blocked": [t.id for t in blocked],
        "totals": {
            "pending": sum(1 for t in tasks if t.status == "PENDING"),
            "complete": sum(1 for t in tasks if t.status == "COMPLETE"),
        },
    }


def append_status_log(previous_metrics: Dict, new_metrics: Dict, summary: Dict):
    try:
        prev_complete = previous_metrics.get("tasks", {}).get("complete")
    except Exception:  # pragma: no cover
        prev_complete = None
    new_complete = new_metrics.get("tasks", {}).get("complete")
    diff_lines = []
    if prev_complete != new_complete:
        diff_lines.append(f"tasks.complete: {prev_complete} -> {new_complete}")
    # Always include timestamp & ready list
    ts = new_metrics.get("last_updated")
    entry = [f"## {ts}", "Tasks Updated:"]
    if diff_lines:
        for d in diff_lines:
            entry.append(f"- {d}")
    else:
        entry.append("- (no task count changes)")
    entry.append("\nSummary:")
    ready = summary.get("ready", [])
    entry.append(f"- Ready queue: {', '.join(ready) if ready else 'EMPTY'}")
    entry.append("\nMetrics Diff (partial):")
    if diff_lines:
        for d in diff_lines:
            entry.append(f"- {d}")
    else:
        entry.append("- none")
    entry.append("\nNotes:\n- Auto-generated by update_progress.py")
    with STATUS_LOG_PATH.open("a", encoding="utf-8") as fh:
        fh.write("\n".join(entry) + "\n\n---\n")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--write", action="store_true", help="Write updated metrics.json")
    parser.add_argument("--summary", action="store_true", help="Print human summary")
    parser.add_argument("--json", action="store_true", help="Print machine JSON summary to stdout")
    parser.add_argument("--append-log", action="store_true", help="Append status_log.md entry if metrics changed")
    args = parser.parse_args()

    tasks = load_tasks()
    errors = validate(tasks)
    if errors:
        for e in errors:
            print(f"ERROR: {e}", file=sys.stderr)
        sys.exit(2)

    previous_metrics = load_metrics()
    new_metrics = recompute_metrics(tasks, previous_metrics.copy())
    summary = summarize(tasks)

    metrics_changed = previous_metrics.get("tasks", {}).get("complete") != new_metrics.get("tasks", {}).get("complete")

    if args.write:
        METRICS_PATH.write_text(json.dumps(new_metrics, indent=2) + "\n")
        if args.append_log:
            append_status_log(previous_metrics, new_metrics, summary)

    if args.summary:
        print("Ready tasks (in priority order):")
        for tid in summary["ready"]:
            print(f"  - {tid}")
        if summary["blocked"]:
            print("Blocked tasks:")
            for tid in summary["blocked"]:
                print(f"  - {tid}")
        print("Totals:")
        for k, v in summary["totals"].items():
            print(f"  {k}: {v}")
        if args.append_log:
            print(f"Status log {'updated' if metrics_changed else 'appended (no count change)'}")

    if args.json:
        print(json.dumps({"summary": summary, "metrics": new_metrics}, indent=2))


if __name__ == "__main__":  # pragma: no cover
    main()
