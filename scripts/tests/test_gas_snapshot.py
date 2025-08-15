import json
from pathlib import Path
import subprocess
import tempfile

SCRIPT = Path(__file__).resolve().parents[2] / 'scripts' / 'progress' / 'gas_snapshot.py'


def test_gas_snapshot_writes_file():
    with tempfile.TemporaryDirectory() as td:
        td = Path(td)
        docs_perf = td / 'docs' / 'perf'
        docs_perf.mkdir(parents=True)
        # copy script
        target_script = td / 'scripts' / 'progress'
        target_script.mkdir(parents=True, exist_ok=True)
        (target_script / 'gas_snapshot.py').write_text(SCRIPT.read_text())
        r = subprocess.run(['python', str(target_script / 'gas_snapshot.py')], cwd=td, capture_output=True, text=True)
        # Expect failure (exit 1) due to placeholder enforcement with zeros
        assert r.returncode == 1
        assert 'PLACEHOLDER_GAS_VALUES' in r.stdout
        snap = json.loads((docs_perf / 'gas_snapshot.json').read_text())
    assert 'generated_at' in snap
    assert len(snap['functions']) == 2
    assert all(isinstance(f['avg_gas'], int) for f in snap['functions'])
