import json
from pathlib import Path
import subprocess
import tempfile

SCRIPT = Path(__file__).resolve().parents[2] / 'scripts' / 'ephem' / 'validation' / 'verify_kernel.py'
HASHES = Path(__file__).resolve().parents[2] / 'scripts' / 'ephem' / 'validation' / 'kernel_hashes.json'

def test_placeholder_update_flow():
    with tempfile.TemporaryDirectory() as td:
        td = Path(td)
        # create fake kernel file
        kernel = td / 'fake_kernel.bsp'
        kernel.write_bytes(b'FAKEKERNELDATA')
        # copy hashes file and replace entry with placeholder
        hashes_dir = td / 'scripts' / 'ephem' / 'validation'
        hashes_dir.mkdir(parents=True, exist_ok=True)
        data = json.loads(HASHES.read_text())
        # ensure entry exists for fake kernel
        data['kernels'] = [{'name': kernel.name, 'sha256': 'REPLACE_WITH_ACTUAL_SHA256'}]
        hashes_path = hashes_dir / 'kernel_hashes.json'
        hashes_path.write_text(json.dumps(data, indent=2))
        # copy script
        (hashes_dir / 'verify_kernel.py').write_text(SCRIPT.read_text())
        r1 = subprocess.run(['python', str(hashes_dir / 'verify_kernel.py'), '--kernel', str(kernel), '--hashes', str(hashes_path)], capture_output=True, text=True)
        assert 'KERNEL_PLACEHOLDER' in r1.stdout
        r2 = subprocess.run(['python', str(hashes_dir / 'verify_kernel.py'), '--kernel', str(kernel), '--hashes', str(hashes_path), '--update'], capture_output=True, text=True)
        assert 'KERNEL_PLACEHOLDER_UPDATED' in r2.stdout
        # second update should be OK and show KERNEL_OK
        r3 = subprocess.run(['python', str(hashes_dir / 'verify_kernel.py'), '--kernel', str(kernel), '--hashes', str(hashes_path)], capture_output=True, text=True)
        assert 'KERNEL_OK' in r3.stdout


def test_kernel_corruption_detected():
    """After updating placeholder, a subsequent modification should trigger mismatch."""
    with tempfile.TemporaryDirectory() as td:
        td = Path(td)
        kernel = td / 'fake_kernel.bsp'
        kernel.write_bytes(b'ORIGINALDATA')
        hashes_dir = td / 'scripts' / 'ephem' / 'validation'
        hashes_dir.mkdir(parents=True, exist_ok=True)
        data = {'kernels': [{'name': kernel.name, 'sha256': 'REPLACE_WITH_ACTUAL_SHA256'}]}
        hashes_path = hashes_dir / 'kernel_hashes.json'
        hashes_path.write_text(json.dumps(data, indent=2))
        (hashes_dir / 'verify_kernel.py').write_text(SCRIPT.read_text())
        # update placeholder
        r_update = subprocess.run(['python', str(hashes_dir / 'verify_kernel.py'), '--kernel', str(kernel), '--hashes', str(hashes_path), '--update'], capture_output=True, text=True)
        assert r_update.returncode == 0 and 'KERNEL_PLACEHOLDER_UPDATED' in r_update.stdout
        # modify kernel (simulate corruption)
        kernel.write_bytes(b'CORRUPTEDDATA')
        r_corrupt = subprocess.run(['python', str(hashes_dir / 'verify_kernel.py'), '--kernel', str(kernel), '--hashes', str(hashes_path)], capture_output=True, text=True)
        assert r_corrupt.returncode == 1 and 'KERNEL_HASH_MISMATCH' in r_corrupt.stdout
