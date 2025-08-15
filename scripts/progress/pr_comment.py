#!/usr/bin/env python3
"""Generate a PR comment summarizing ready tasks and overall progress.
Prints markdown to stdout. Intended for CI usage.
"""
from pathlib import Path
import json
import sys
import yaml  # type: ignore

ROOT = Path(__file__).resolve().parents[2]
TASK_GRAPH_PATH = ROOT / "docs/progress/task_graph.yaml"
METRICS_PATH = ROOT / "docs/progress/metrics.json"


def load():
    tasks = yaml.safe_load(TASK_GRAPH_PATH.read_text()).get("tasks", [])
    metrics = json.loads(METRICS_PATH.read_text()) if METRICS_PATH.exists() else {}
    return tasks, metrics


def main():
    tasks, metrics = load()
    complete = {t['id'] for t in tasks if t.get('status') == 'COMPLETE'}
    ready = [t for t in tasks if t.get('status') == 'PENDING' and all(d in complete for d in t.get('depends_on', []))]
    ready_sorted = sorted(ready, key=lambda x: (x.get('priority', 999), x.get('phase', 99), x.get('id')))
    print("### Project Progress Summary\n")
    print(f"Total Tasks: {len(tasks)}  |  Complete: {len(complete)}")
    print("\n**Ready Queue (next actionable):**")
    if not ready_sorted:
        print("- *(none)*")
    else:
        for t in ready_sorted[:10]:
            print(f"- `{t['id']}` – {t.get('title')} (phase {t.get('phase')}, priority {t.get('priority')})")
    if metrics:
        tasks_m = metrics.get('tasks', {})
        print("\n**Metrics:**")
        print(f"- Complete: {tasks_m.get('complete')} / {tasks_m.get('total')}")
        print(f"- In Progress: {tasks_m.get('in_progress')}")
        print(f"- Blocked: {tasks_m.get('blocked')}")
    print("\n_Generated automatically._")


if __name__ == '__main__':
    main()
