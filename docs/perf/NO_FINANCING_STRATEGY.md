# Zero-Cost Contract Testing & Gas Estimation Strategy

You mentioned having no financing to pay on-chain gas fees. This repo can still reach high confidence in correctness and approximate gas costs **without spending real funds** by layering fully local approaches.

## Goals

1. Validate stateful logic (roles, cooldown, thresholds, weight + ML score updates, trade execution) off-chain.
2. Derive repeatable gas baselines locally.
3. Keep CI green and deterministic (no flaky network dependencies).
4. Make later mainnet/testnet deployment a “formality” rather than a discovery phase.

## Testing / Validation Pyramid (No Real Gas Required)

| Layer                           | Cost         | Purpose                                                        | Tooling                             |
| ------------------------------- | ------------ | -------------------------------------------------------------- | ----------------------------------- |
| Pure function tests             | 0            | Math/composite correctness                                     | `cairo-test` (already present)      |
| Stateful simulation (in‑memory) | 0            | Storage transitions, access control                            | Custom harness module (proposed)    |
| Local devnet execution          | 0 (faux ETH) | Real contract ABI, syscall semantics, gas estimation           | `starknet-devnet` (Docker)          |
| Property / fuzz tests           | 0            | Edge cases, invariants (cooldown monotonicity, role isolation) | Off-chain model + random generation |
| (Deferred) Public testnet       | Real funds   | Final pre-prod sanity                                          | Minimal selective scenarios         |

## Recommended Immediate Steps

1. Add a pure in-memory state struct mirror of on-chain storage for rapid unit tests (no deployment).
2. Introduce a local devnet script (added: `scripts/contracts/devnet_local.sh`) to spin up `starknet-devnet` with free accounts.
3. Write a thin deploy + call Python helper (future) to:
   - Compile contract (Scarb)
   - Declare & deploy on devnet
   - Invoke critical external functions, capturing gas usage from traces
   - Emit JSON summary (to update `docs/perf/gas_snapshot.json`).
4. Incrementally replace placeholder stateful tests with simulated harness tests; only once stable, add a devnet integration test subset (small N) gated behind a make target so normal `scarb test` stays fast.

## Gas Estimation Locally

Devnet enforces Cairo execution logic and produces fee estimates. Strategy:

1. Start devnet with fixed seed & zero-fee mode OR mintable accounts.
2. Use `starknet` CLI or Python client to send `estimate_fee` / `invoke` (dry-run first).
3. Record `gas_consumed` (or equivalent execution resources) into a fresh JSON (not committed) then normalize & update `docs/perf/gas_snapshot.json` via a script that enforces % drift thresholds.

### Deterministic Devnet Parameters

- Seed (e.g. `--seed 20240815`)
- Disable auto gas price variance; if supported: `--gas-price 1` (small constant).
- Fixed initial balance per predeployed account (accept defaults).

## Proposed File Additions (This PR)

Added `scripts/contracts/devnet_local.sh` (see below) and this strategy doc. Next scripts (not yet added) could include:

- `scripts/contracts/compile_and_deploy.py`
- `scripts/contracts/estimate_gas.py`

## Integration with Progress Metrics

Once gas collection script exists:

1. Run locally, produce `gas_measurements.new.json`.
2. Compare to committed snapshot using existing comparison pattern (similar to Python perf scripts already in `scripts/ci`).
3. If outside tolerance, fail CI & prompt update.

## Minimal In-Memory Harness Outline (Future Work)

```cairo
// contracts/tests/state_sim.cairo (planned)
struct SimState { /* mirrors storage fields */ }
impl SimLogic of SimTrait { /* pure functions mirroring externals */ }
// Tests mutate SimState then assert invariants.
```

Advantages: zero syscall complexity, instant feedback. Later cross-check a subset against devnet deployed contract.

## Fuzz / Property Test Ideas

- Cooldown: invoking action before `last_action + cooldown` always reverts / is blocked.
- Role isolation: only admin can update weights; others fail.
- Threshold monotonicity: lowering threshold cannot invalidate previously valid trades (depending on semantics).

Can be implemented off-chain in Python replicating logic or in Cairo if/when property test utilities mature.

## When Real Funds Are Actually Needed

Only for: public testnet deployment rehearsal, verifying compatibility with ecosystem tooling (explorers, wallets). Defer until core invariants and gas envelopes are stable.

## Quick Start (Local Only)

```bash
# 1. Start devnet (free funds)
./scripts/contracts/devnet_local.sh

# 2. (Future) Compile & deploy
# scarb build
# python scripts/contracts/compile_and_deploy.py --network localhost

# 3. (Future) Estimate gas
# python scripts/contracts/estimate_gas.py --out docs/perf/gas_snapshot.json
```

## Summary

You can achieve high confidence with zero real financing by layering pure logic tests, in-memory simulations, and a deterministic devnet for gas & syscall coverage. This document anchors that plan.
