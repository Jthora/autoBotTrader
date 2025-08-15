## Aug 2025 Update — Ephemeris Providers

- Providers:
  - `file`: reads embedded precomputed series (default; offline, deterministic).
  - `algo`: computes locally using pure-Go Meeus/VSOP87 (feature-flagged after validation).
- Selection via `EPHEM_MODE=file|algo`. Both providers expose the same interface and return a normalized 0–100 scalar plus raw basis metadata.

# Adapters Overview (Draft)

Adapters abstract external execution (e.g., SuperDEX unified margin, RPI) away from the core contract.

## Concept

```
trait IExecutionAdapter {
    fn execute(order: ExecutionOrder) -> (filled_amount: u128, improved_price_bps: u32);
}
```

## Rationale

- Keep core minimal & auditable.
- Allow iterative adapter deployment / replacement.

## Security

- Malicious adapter risk -> off-chain reconciliation of fills.

## Status

Documentation only; implementation post Phase 6.
