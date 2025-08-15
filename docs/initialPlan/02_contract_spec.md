# Cairo Contract Specification

Version: Draft v0.1
Target Cairo: 2.11.x

## Objectives

Provide deterministic gating of trades based on a composite prediction score derived from multiple signal sources while minimizing storage writes and emitting granular events for off-chain indexing.

## Storage Layout (Proposed Order v1 — versioned)

1. admin: felt252
2. pusher_role: felt252
3. ml_oracle: felt252
4. execution_threshold: u32 (default 50)
5. cooldown_seconds: u32 (default 0)
6. last_input_timestamp: u64
7. formula_version: u32 (starts 1)
8. normalization_version: u32 (starts 1)
9. astrology_w: u32 (default 50)
10. gravity_w: u32 (default 50)
11. ml_w: u32 (default 0)
12. astrology_score: u32
13. gravity_score: u32
14. ml_score: u32
15. ml_score_timestamp: u64
16. ml_model_version: u32
17. composite_score: u32
18. trades_len: u64 (optional if map retained)
19. trades: LegacyMap<felt252, Trade> (MAY be disabled in favor of event-only)
20. reserved_0, reserved_1 (future adapter usage)

Trade struct:

```
struct Trade {
    amount: u128,
    direction: u8,        // 0 = short, 1 = long
    executed_score: u32,
    timestamp: u64,
}
```

## Events

```
PredictionUpdated(astrology: u32, gravity: u32, ml: u32, composite: u32, formula_version: u32, normalization_version: u32)
TradeExecuted(trade_id: felt252, direction: u8, amount: u128, score: u32, timestamp: u64, improved_price_bps: u32)
WeightsUpdated(astrology_w: u32, gravity_w: u32, ml_w: u32)
ThresholdUpdated(threshold: u32)
CooldownUpdated(cooldown_seconds: u32)
RolesUpdated(pusher_role: felt252, ml_oracle: felt252)
MLScoreUpdated(ml_score: u32, ml_score_timestamp: u64, ml_model_version: u32)
AdminChanged(new_admin: felt252)
```

## Entry Points

- init(admin: felt252, pusher_role: felt252)
- set_prediction_inputs(astrology_score: u32, gravity_score: u32) [pusher]
- set_prediction_inputs_batch(astrology: u32, gravity: u32, ml: u32?, ml_ts: u64?) [pusher/oracle combined; gas optimization] (optional)
- set_ml_score(ml_score: u32, ml_model_version: u32, score_timestamp: u64) [ml_oracle]
- update_weights(a_w: u32, g_w: u32, ml_w: u32) [admin]
- update_threshold(threshold: u32) [admin]
- update_cooldown(cooldown_seconds: u32) [admin]
- set_roles(pusher_role: felt252, ml_oracle: felt252) [admin]
- execute_trade(trade_id: felt252, amount: u128, direction: u8)
- get_composite() -> u32 (view)
- get_trade(trade_id: felt252) -> Trade (view) (if map retained)
- get_state() -> (threshold, weights tuple, composite, inputs, formula_version)

## Composite Score Computation

```
require all scores <= 100
require weights valid: (a_w + g_w + ml_w) > 0
composite = floor((a*aw + g*gw + ml*mw) / (aw + gw + mw))
```

Logic executed inside set_prediction_inputs (and set_ml_score if ml changes and others exist).

## Gas Optimization Techniques

- Batch update path to avoid double storage writes for ml + base scores.
- Skip recomputation when weighted sum unchanged.
- Event-only trade logging option (reduce persistent storage writes).
- Cooldown prevents spam invoking set_prediction_inputs.

## Validation & Errors (Pseudo)

- assert(a_score <= 100 && g_score <= 100 && ml_score <= 100)
- assert(weights sum > 0)
- assert(trade not already exists) when executing (trade_id uniqueness)
- assert(composite >= execution_threshold) in execute_trade

## Access Control

```
fn ensure_admin(caller)     { assert(caller == admin, 'NOT_ADMIN'); }
fn ensure_pusher(caller)    { assert(caller == pusher_role, 'NOT_PUSHER'); }
fn ensure_ml_oracle(caller) { assert(caller == ml_oracle, 'NOT_ORACLE'); }
```

RolesUpdated event emitted when changed.

## Upgrade Path (Future)

- Introduce proxy with same storage layout prefix
- Version variable (u32) to guard migration logic

## Open Questions

- Commit–reveal viability vs simple cooldown (ADR will finalize: likely Deferred for MVP).
- Persistent trades map necessity vs pure events (ADR: trade_logging_strategy).
- On-chain signature verification for ml_score (cost vs trust tradeoff; likely off-chain pre-verified first).
- Improved price metric accuracy once adapter exists.

---

## Aug 2025 Update — On-chain Scope and Verification

- The contract does not compute ephemerides. It remains a neutral registry and verifier:
  - Stores weights/config and a verifier public key.
  - Accepts signed inputs (astrology_score, gravity_score, optional ml_score) with timestamp.
  - Enforces cadence and bounds; rejects invalid signatures or out-of-window submissions.
  - Emits events with all inputs and the computed composite for auditability.
- Optional roadmap: accept commitments (e.g., Merkle root of a precomputed series) for stronger guarantees without heavy on-chain math.

### Event Extensions

- `PredictionPushed(address submitter, u32 astro, u32 grav, u32 ml, u64 ts, u32 composite)` — emitted on each accepted update.

### Security Notes

- Public key can be rotated by admin; store version to correlate signatures.
- Rate-limiting via cooldown/cadence prevents spam; keep parameters configurable.

---

Implementation plan in `03_contract_implementation_plan.md`.
