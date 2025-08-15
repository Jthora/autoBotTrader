# Frontend (React + Vite + TypeScript) Specification

## Objectives
Provide real-time visibility into prediction scores and trades, allow configuration of weights (admin only), and enable manual trade execution/testing on testnet.

## Pages / Views
- Dashboard: composite score, individual scores, formula_version, normalization_version, last update time
- Trades: table (event-derived) with trade_id, direction, amount, executed_score, timestamp, improved_price_bps placeholder
- Admin Panel: weights editor, threshold editor, cooldown editor, ml weight/score viewer, role addresses display
- Disclaimer Panel: risk / experimental signal notice (link to README section)

## Core Components
| Component | Purpose |
|-----------|---------|
| App.tsx | Layout + routing |
| ScorePanel.tsx | Displays current scores & composite |
| TradesTable.tsx | Paginated trade events list |
| WeightsForm.tsx | Update weights & threshold (if admin) |
| ExecuteTradeForm.tsx | Input trade params + trigger execute_trade |
| MLPanel.tsx | Shows ml_score and weight |
| WalletConnect.tsx | Connects wallet (ArgentX/Braavos) |

## State Management
- useQuery (TanStack Query) for contract reads + caching
- Lightweight global store (Zustand) for wallet & config

## Contract Interaction Layer
`starknet.ts`:
- init provider
- getContract(address, abi)
- readComposite()
- readState()
- executeTrade(tradeId, amount, direction)
- updateWeights(a, g, ml)
- setPredictionInputs (for testing only; normally off-chain service)

## Data Refresh Strategy
- Poll composite + scores every 15s (adjustable)
- Listen to events via provider (if available); else fallback to polling

## Styling
- Minimal: TailwindCSS (fast prototyping) or simple CSS modules

## Security / UX
- Hide or disable admin & role-gated controls when wallet not authorized.
- Display role addresses + copy buttons.
- Show cooldown remaining timer if user attempts early update.
- Pending tx status with explorer link.

## Error Handling
- Toast notifications for failures
- Fallback UI: if contract unreachable, show offline banner

## Testing
- Component tests (React Testing Library) for ScorePanel + trade execution flow
- Mock starknet.js provider

## Future
- Historical charts (composite over time) via events indexing
- Dark mode toggle

---
Implementation plan: `07_frontend_impl_plan.md`.
