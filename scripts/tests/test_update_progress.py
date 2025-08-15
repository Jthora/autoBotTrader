import json
import os
from pathlib import Path
import subprocess
import tempfile
import shutil

ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / 'scripts' / 'progress' / 'update_progress.py'


def write_task_graph(tmp: Path, tasks_yaml: str):
    p = tmp / 'docs' / 'progress'
    p.mkdir(parents=True, exist_ok=True)
    (p / 'task_graph.yaml').write_text(tasks_yaml)
    (p / 'metrics.json').write_text('{}')


def run(args, cwd):
    env = dict(**os.environ)
    env['PROGRESS_ROOT'] = str(cwd)
    return subprocess.run(['python', str(SCRIPT), *args], cwd=cwd, capture_output=True, text=True, env=env)


def test_duplicate_id_flagged():
    td = Path(tempfile.mkdtemp(prefix='progress_test_'))
    try:
        write_task_graph(td, 'tasks:\n  - id: a\n    phase: 0\n    priority: 1\n    depends_on: []\n    status: PENDING\n  - id: a\n    phase: 0\n    priority: 2\n    depends_on: []\n    status: PENDING\n')
        r = run(['--summary'], cwd=td)
        assert r.returncode == 2, r.stderr
        assert 'Duplicate task id detected: a' in r.stderr
    finally:
        shutil.rmtree(td, ignore_errors=True)


def test_ready_computation_simple():
    td = Path(tempfile.mkdtemp(prefix='progress_test_'))
    try:
        write_task_graph(td, 'tasks:\n  - id: a\n    phase: 0\n    priority: 1\n    depends_on: []\n    status: COMPLETE\n  - id: b\n    phase: 0\n    priority: 2\n    depends_on: [a]\n    status: PENDING\n')
        r = run(['--json'], cwd=td)
        assert r.returncode == 0
        data = json.loads(r.stdout)
        assert data['summary']['ready'] == ['b']
    finally:
        shutil.rmtree(td, ignore_errors=True)

def test_cycle_and_invalid_status_detection():
    td = Path(tempfile.mkdtemp(prefix='progress_test_'))
    try:
        # a depends on b, b depends on a (cycle). Also invalid status FOOBAR.
        yaml = 'tasks:\n  - id: a\n    phase: 0\n    priority: 1\n    depends_on: [b]\n    status: PENDING\n  - id: b\n    phase: 0\n    priority: 2\n    depends_on: [a]\n    status: FOOBAR\n'
        write_task_graph(td, yaml)
        r = run(['--summary'], cwd=td)
        assert r.returncode == 2
        stderr = r.stderr
        assert 'Cycle detected' in stderr
        assert 'invalid status' in stderr
    finally:
        shutil.rmtree(td, ignore_errors=True)

