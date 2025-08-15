## Aug 2025 Update — Demo Packaging and Local Run

- Build frontend and embed assets into the Go binary; serve UI and API from one process bound to localhost.
- Provide a macOS `.app` wrapper to launch the binary and open the browser automatically.
- Default ephemeris mode is `file` (embedded series); no network required at runtime.
- Optional on-chain push requires env vars: `CONTRACT_ADDRESS`, `STARKNET_RPC`, `STARKNET_KEY`.
- Include version metadata (git commit, ephemeris dataset version) in `/health` for traceability.

# Deployment & Operations Runbook

## Environments

| Env     | Purpose        | Network          | Notes       |
| ------- | -------------- | ---------------- | ----------- |
| local   | Dev iteration  | Local devnet     | Fast cycles |
| testnet | Shared QA      | Starknet Sepolia | Public RPC  |
| prod    | Future mainnet | Paradex mainnet  | TBD         |

## Prerequisites

- Starkli installed (v0.4.0+)
- Scarb for building Cairo contracts
- Node.js v20+, Go 1.21+
- Wallet private key (test) exported as env

## Environment Variables

```
STARKNET_RPC_URL=https://sepolia.starknet.io
CONTRACT_ADDRESS=0x...
PRIVATE_KEY=0x...               # Admin key (secure)
PUSHER_PRIVATE_KEY=0x...        # Signal ingestion key
ML_ORACLE_PRIVATE_KEY=0x...     # ML oracle key (optional until used)
API_PORT=8080
PUSH_HMAC_SECRET=optional
VITE_CONTRACT_ADDRESS=0x...
VITE_RPC_URL=https://sepolia.starknet.io
```

## Contract Deployment (Testnet Example)

1. Build: `scarb build`
2. Declare class: `starkli declare target/.../trading_bot_contract.json`
3. Deploy: `starkli deploy <CLASS_HASH> <admin_address>`
4. Record address in docs + `.env`

## Off-Chain Service Deployment

- Containerize (Dockerfile multi-stage)
- Run: `docker run -e STARKNET_RPC_URL -e CONTRACT_ADDRESS -e PRIVATE_KEY image:tag`
- Health check: GET /health (expect 200 JSON)

## Frontend Deployment

- Build: `npm run build`
- Deploy dist/ to Vercel (configure env vars)
- Post-deploy smoke test: load dashboard, verify composite fetch

## Updating Weights / Threshold / Cooldown / Roles

1. Connect admin wallet
2. Call update_weights or update_threshold via frontend Admin Panel
3. Verify WeightsUpdated / ThresholdUpdated events

## Rolling Out ML Score (Future)

- Set ml_w > 0 via admin
- Start oracle service posting ml_score
- Monitor composite change frequency

## Incident Response

| Issue                      | Detection               | Mitigation                                               |
| -------------------------- | ----------------------- | -------------------------------------------------------- |
| Off-chain providers down   | /predict warnings       | Use last cached values; alert if >1h stale               |
| Contract stuck (no pushes) | Composite timestamp old | Investigate off-chain logs; manual set_prediction_inputs |
| High gas / fail            | Revert messages         | Optimize recompute or batch updates                      |
| Wrong weights set          | Event audit             | Reset via admin transaction                              |

## Observability (Future)

- Add Prometheus metrics for request latency & push success rate
- Log shipping to ELK or Loki

## Backup & Recovery

## Risk Disclaimer (Communicate to Users)

Experimental strategy using non-traditional signals. No guarantee of profit. Users should size positions responsibly. Transparent events enable community verification.

- Contract state immutable (on-chain)
- Off-chain service stateless; redeploy from image
- Keep infrastructure IaC (Terraform) (future)

---

Roadmap in `11_roadmap.md`.
