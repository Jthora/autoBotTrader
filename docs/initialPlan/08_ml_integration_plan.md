# ML / PPO Integration Plan

## Goal
Prepare seamless future inclusion of ML-derived predictive scores (e.g., PPO reinforcement learning model) without refactoring core contract logic.

## Current Placeholder
- Contract fields: ml_score (u32), ml_score_timestamp (u64), ml_model_version (u32), ml_w weight.
- Function: set_ml_score(ml_score, ml_model_version, score_timestamp) [ml_oracle role].
- Composite formula already accommodates ml_score & weight (ml_w may be 0 initially).

## Off-Chain ML Pipeline (Future)
1. Data Assembly: Market OHLCV + astrology + gravimetrics
2. Feature Engineering: Normalize temporal features, cyclical encoding for planetary angles
3. PPO Training:
   - Environment sim: reward = realized PnL adjusted by volatility penalty
   - Model outputs probability/score (0–1) -> scale to 0–100
4. Score Publishing:
   - Signed payload {score, timestamp, model_version}
   - Oracle relays to contract Via set_ml_score

## Python Skeleton (predict.py)
```
class PPOModel:
    def __init__(self, config): ...
    def train(self, data): ...
    def predict(self, features) -> float: ... # 0..1

if __name__ == '__main__':
    # load data (placeholder)
    # train or load weights
    # output scaled score
```

## Oracle Strategy Options
| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| Centralized Signer | Single key pushes ml_score | Simple | Trust assumption |
| Multi-sig | m-of-n signers push | Reduced single-point trust | Coordination overhead |
| Chainlink External Adapter | Custom adapter feeds score | Decentralized path | Setup complexity |
| Validity Proof (ZK ML) | Prove inference correctness | High trustlessness | Heavy R&D |

## Versioning
- ml_model_version stored & emitted; increments when model architecture / training regime changes.
- formula_version distinct (affects composite math) vs ml_model_version (affects input semantics).

## Testing Plan
- Unit test: set_ml_score with ml_w=0 (no composite change)
- Unit test: set_ml_score with ml_w>0 (changes composite)

## Risk Mitigation
- Cap ml_score <= 100 enforced on-chain.
- Stale score timeout threshold (configurable future) prevents outdated ML influence.
- Signature scheme (future) ensures authenticity (message = hash(model_version|score|timestamp|expiry)).

---
Next: `09_testing_strategy.md`.
