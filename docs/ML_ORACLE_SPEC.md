## Aug 2025 Note — Data Path Integrity

- Ephemeris-derived gravity signal is computed locally (no network at runtime). The off-chain service signs the payloads sent on-chain.
- ML scores, when present, follow the same path and are subject to the same cadence/bounds checks on-chain.
- Avoid mixing client-side computed values into authoritative paths; the UI remains a visualization layer only.

# ML Oracle Specification (Draft)

## Payload Structure

```
domain = "ML_SCORE_V1"
fields = model_version (u32) | ml_score (u32 0-100) | score_timestamp (u64) | expiry (u64)
message = keccak256(domain || fields)
```

## Signature Phases

- Phase 1: Single signer (ml_oracle role)
- Phase 2: Multi-sig / committee
- Phase 3: (R&D) ZK proof of inference

## Validation (Future Implementation)

1. score_timestamp <= block.timestamp
2. block.timestamp <= expiry
3. signature over message matches authorized key

## Replay Protection

Expiry window + differing timestamps per update.

## Staleness

Optional future `STALE_WINDOW` to reject old scores.
