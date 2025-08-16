#!/usr/bin/env python3
"""True fee estimation against a running local Starknet devnet.

Prerequisites:
  1. Run devnet locally (e.g. ./scripts/contracts/devnet_local.sh)
  2. Install dependencies: pip install -r requirements.txt (starknet-py, requests)
  3. Ensure `scarb build` has produced artifacts under contracts/target/dev

Process:
  - Fetch first predeployed account from devnet
  - Compile (already via Scarb) & declare the contract if not yet declared in this session
  - Deploy the contract (constructor takes admin + pusher addresses; we use the same account for both by default)
  - Estimate fees for: set_prediction_inputs, execute_trade (after setting threshold high enough if needed)
  - Output JSON snapshot: docs/perf/gas_snapshot_devnet.json with overall_fee (wei) and gas_unit_price if available

Notes:
  - This performs real declare + deploy transactions on devnet (free). Re-running will redeclare unless you pass --reuse-address with an existing contract address (to only estimate invokes).
  - If declaration already performed, you can skip by providing --class-hash.
  - Auto-fee estimation is used; underlying gas vs fee separation depends on devnet configuration. We record overall_fee.
"""
from __future__ import annotations
import argparse
import asyncio
import json
import sys
from pathlib import Path
from datetime import datetime, timezone

try:
    import requests  # type: ignore
    from starknet_py.net.full_node_client import FullNodeClient
    from starknet_py.net.account.account import Account
    from starknet_py.net.models import StarknetChainId
    from starknet_py.net.signer.key_pair import KeyPair
    from starknet_py.contract import Contract
except ImportError as e:  # pragma: no cover
    print(f"IMPORT_ERROR: {e}. Install dependencies via requirements.txt (starknet-py, requests).", file=sys.stderr)
    sys.exit(1)

DEVNET_URL_DEFAULT = "http://localhost:5050"
RPC_SUFFIX = "/rpc"

def find_artifacts() -> tuple[Path, Path]:
    base = Path('contracts/target/dev')
    candidates = sorted(base.rglob('*.contract_class.json'), key=lambda p: p.stat().st_mtime, reverse=True)
    if not candidates:
        raise FileNotFoundError("No .contract_class.json artifacts found. Run 'scarb build' in contracts directory.")
    sierra = candidates[0]
    stem = sierra.name.replace('.contract_class.json', '')
    casm_patterns = [f"**/{stem}.compiled_contract_class.json", f"**/{stem}.casm.json"]
    casm_file = None
    for pat in casm_patterns:
        matches = list(base.rglob(pat))
        if matches:
            casm_file = matches[0]
            break
    if not casm_file:
        raise FileNotFoundError("CASM artifact not found for sierra artifact: " + sierra.name)
    return sierra, casm_file


def load_predeployed(devnet_url: str) -> dict:
    import requests
    resp = requests.get(f"{devnet_url}/predeployed_accounts", timeout=10)
    resp.raise_for_status()
    accounts = resp.json()
    if not accounts:
        raise RuntimeError("No predeployed accounts returned by devnet")
    return accounts[0]


async def ensure_contract(account: Account, sierra_path: Path, casm_path: Path, class_hash: str | None, reuse_address: str | None, admin_addr: int, pusher_addr: int):
    if reuse_address:
        return await Contract.from_address(reuse_address, account)
    sierra_dict = json.loads(sierra_path.read_text())
    casm_dict = json.loads(casm_path.read_text())
    if not class_hash:
        decl = await Contract.declare_v2(
            account=account,
            compiled_contract=sierra_dict,
            compiled_contract_casm=casm_dict,
            auto_estimate=True,
        )
        await decl.wait_for_acceptance()
        class_hash = hex(decl.class_hash)
    deploy_result = await Contract.deploy_v2(
        account=account,
        class_hash=int(class_hash, 16),
        constructor_args=[admin_addr, pusher_addr],
        auto_estimate=True,
    )
    deployed = await deploy_result.wait_for_acceptance()
    return deployed


async def estimate_calls(contract: Contract, account: Account):
    call1 = await contract.functions["set_prediction_inputs"].invoke(60, 40, auto_estimate=True)
    fee1 = call1.invoke_transaction.fee_estimation.overall_fee
    await contract.functions["set_prediction_inputs"].invoke(90, 90, auto_estimate=True)
    call2 = await contract.functions["execute_trade"].invoke(1, 100, 1, 10, auto_estimate=True)
    fee2 = call2.invoke_transaction.fee_estimation.overall_fee
    return {
        "update_prediction": int(fee1),
        "execute_trade": int(fee2),
    }


async def run(args):
    devnet_url = args.devnet
    rpc_url = devnet_url + RPC_SUFFIX
    pre = load_predeployed(devnet_url)
    priv_key_hex = pre['private_key']
    address_hex = pre['address']
    key_pair = KeyPair.from_private_key(int(priv_key_hex, 16))
    client = FullNodeClient(node_url=rpc_url)
    account = Account(address=int(address_hex, 16), client=client, key_pair=key_pair, chain=StarknetChainId.TESTNET, supported_tx_version=1)

    sierra, casm = find_artifacts()
    contract = await ensure_contract(account, sierra, casm, args.class_hash, args.reuse_address, int(address_hex,16), int(address_hex,16))

    fees = await estimate_calls(contract, account)
    snapshot = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "network": devnet_url,
        "overall_fees": fees,
        "unit": "wei"
    }
    out_path = Path('docs/perf/gas_snapshot_devnet.json')
    out_path.write_text(json.dumps(snapshot, indent=2) + '\n')
    print(f"Wrote devnet fee snapshot -> {out_path}")


def parse_args():
    ap = argparse.ArgumentParser()
    ap.add_argument('--devnet', default=DEVNET_URL_DEFAULT, help='Base devnet URL (no /rpc)')
    ap.add_argument('--class-hash', help='Reuse existing declared class hash (hex) to skip declaring')
    ap.add_argument('--reuse-address', help='Skip deploy/declare and reuse existing contract address (hex)')
    return ap.parse_args()


def main():
    args = parse_args()
    try:
        asyncio.run(run(args))
    except Exception as e:  # pragma: no cover
        print(f"ERROR: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == '__main__':
    main()
