# Contract Implementation Plan

## Steps
1. Scaffold Scarb project (Starkli deploy) with TradingBot.cairo (no Hardhat in core path)
2. Define storage struct per spec (order critical)
3. Implement events
4. Implement helpers: ensure_admin(), compute_composite()
5. Entry point: init(admin)
6. Implement set_prediction_inputs() with cooldown + recompute guard
7. Implement set_ml_score() (timestamp + version) [oracle role]
8. Implement update_weights(), update_threshold(), update_cooldown(), set_roles(), change_admin()
9. Implement execute_trade() (event-only OR map toggle)
10. Views: get_composite(), get_trade(), get_state()
11. Add unit tests (prediction calc, weight update, threshold fail, trade success)
12. Gas review to reduce redundant writes
13. Static analysis (cairo-format, cairo lint tools) & finalize

## Helper Pseudocode
```
fn compute_composite(a, g, ml, aw, gw, mw) -> u32 {
    let total = aw + gw + mw;
    assert(total > 0, 'ZERO_WEIGHTS');
    return (a*aw + g*gw + ml*mw) / total;
}
```

## Anti-Replay (trade_id)
- Use LegacyMap::get(trade_id) -> Option
- If exists -> revert 'TRADE_EXISTS'

## Timestamp Source
- Starknet block timestamp

## Minimal Error Strings (felt hashing optional later)
- 'NOT_ADMIN'
- 'INVALID_SCORE'
- 'ZERO_WEIGHTS'
- 'LOW_SCORE'
- 'TRADE_EXISTS'

## Test Matrix (Expanded)
| Test | Purpose |
|------|---------|
| init sets admin | constructor correctness |
| set_prediction_inputs valid | score recompute |
| set_prediction_inputs invalid (>100) | revert |
| update_weights non-admin | revert |
| update_weights recompute composite | ensure recompute done |
| execute_trade below threshold | revert |
| execute_trade success | event emitted (and stored map if enabled) |
| cooldown enforced | second immediate set_prediction_inputs reverts |
| pusher role required | unauthorized caller reverts |
| ml_oracle role required | unauthorized ml score push reverts |
| ml stale rejection (future) | stale timestamp reverts (if threshold set) |
| batch path (if implemented) | single gas-efficient update |
| trade_id reuse | revert |
| ml_score path (weight=0) | no effect |
| ml_score path (weight>0) | recompute |

## Future Optimizations
- Batch update inputs (include ml) to reduce recompute frequency
- Use event-only trade log (omit trades map) if storage becomes heavy

---
Next: See `04_offchain_service_spec.md` for Go API details.
