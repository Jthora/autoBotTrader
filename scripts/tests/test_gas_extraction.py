import subprocess
import json
import os
import sys
from pathlib import Path

SCRIPT = Path(__file__).resolve().parents[2] / 'scripts' / 'progress' / 'gas_snapshot.py'

FAKE_OUTPUT = """
set_prediction_inputs ... gas=1200
execute_trade ... gas=3400
set_prediction_inputs ... gas=1400
""".strip()

def test_gas_extraction_parser_average(tmp_path):
    # Write a helper script that simply prints FAKE_OUTPUT to stdout and exits 0
    helper = tmp_path / 'echo_gas.sh'
    helper.write_text(f"#!/usr/bin/env bash\ncat <<'EOF'\n{FAKE_OUTPUT}\nEOF\n")
    helper.chmod(0o755)
    env = os.environ.copy()
    env['GAS_TEST_CMD'] = str(helper)
    # Allow placeholders off so failure occurs if zeros remain
    if 'ALLOW_GAS_PLACEHOLDERS' in env:
        del env['ALLOW_GAS_PLACEHOLDERS']
    r = subprocess.run([sys.executable, str(SCRIPT)], capture_output=True, text=True, env=env)
    assert r.returncode == 0, r.stdout + r.stderr
    snap_path = Path(__file__).resolve().parents[2] / 'docs' / 'perf' / 'gas_snapshot.json'
    data = json.loads(snap_path.read_text())
    fn_map = {f['name']: f['avg_gas'] for f in data['functions']}
    # Average: (1200 + 1400)/2 = 1300
    assert fn_map['update_prediction'] == 1300
    assert fn_map['execute_trade'] == 3400

def test_gas_extraction_missing_function_fails(tmp_path):
    # Only one function present
    helper = tmp_path / 'echo_gas.sh'
    helper.write_text("#!/usr/bin/env bash\necho 'execute_trade gas=5000'\n")
    helper.chmod(0o755)
    env = os.environ.copy()
    env['GAS_TEST_CMD'] = str(helper)
    if 'ALLOW_GAS_PLACEHOLDERS' in env:
        del env['ALLOW_GAS_PLACEHOLDERS']
    r = subprocess.run([sys.executable, str(SCRIPT)], capture_output=True, text=True, env=env)
    assert r.returncode == 1
    assert 'PLACEHOLDER_GAS_VALUES' in r.stdout
