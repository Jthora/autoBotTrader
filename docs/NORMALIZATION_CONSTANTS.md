Note (Aug 2025):

- Normalization ranges remain invariant across ephemeris providers. If the basis distribution shifts after switching providers or updating datasets, bump the calc/version and record the change in progress/status with a brief rationale.
- Include the ephemeris dataset version/hash in API responses to tie outputs to a specific normalization context.

# Normalization Constants (Version 1)

Normalization version: `1`

Deterministic constants & integer scaling rules for transforming raw signal inputs into bounded scores (0–100).

## Strategy

1. Use basis points (0–10_000) internally.
2. Clamp to declared ranges.
3. Floor divide by 100 for final score.

## Astrology (Placeholder)

| Metric           | Min | Max | Notes                       |
| ---------------- | --- | --- | --------------------------- |
| volatility_index | 0   | 720 | Combined angular dispersion |

Formula:

```
clamped = clamp(volatility_index, 0, 720)
bp = (clamped * 10_000) / 720
score = bp / 100
```

## Gravimetrics (Placeholder)

| Metric           | Min | Max | Notes         |
| ---------------- | --- | --- | ------------- |
| lunar_tide_force | 80  | 130 | Example range |

Formula:

```
clamped = clamp(lunar_tide_force, 80, 130)
bp = ((clamped - 80) * 10_000) / 50
score = bp / 100
```

## Golden Vectors

| Astro Input | Astro Score | Tide Input | Tide Score |
| ----------- | ----------- | ---------- | ---------- |
| 0           | 0           | 80         | 0          |
| 360         | 50          | 105        | 50         |
| 720         | 100         | 130        | 100        |

## Versioning

Changes require:

1. New section with differences.
2. Increment `normalization_version` on-chain & off-chain.
3. Update golden vectors & tests.
