## Aug 2025 Update — Packaging and Run Experience

- The React app will be built to static assets and embedded in the Go binary via `embed`. The Go server will serve `/` for the UI and `/api/*` endpoints.
- For the demo, provide a macOS `.app` wrapper that launches the server on localhost and opens the default browser to the active port. No terminal required.
- Keep the UI read-only for ephemeris visualization; authoritative values must come from the Go API to preserve determinism.

# Frontend Implementation Plan

## Steps

1. `npm create vite@latest frontend -- --template react-ts`
2. Install deps: starknet@7.x, zustand, @tanstack/react-query, (optional) tailwindcss
3. Setup project structure:

```
frontend/src/
  components/
  hooks/
  lib/starknet.ts
  pages/
  state/
  App.tsx
  main.tsx
```

4. Configure QueryClientProvider + Zustand store
5. Implement starknet.ts wrapper with provider + contract interactions
6. Create ABI placeholder (import once generated)
7. Implement ScorePanel (reads composite + scores + formula_version + normalization_version)
8. Implement TradesTable (mock data first, then event-derived)
9. Implement WalletConnect (connect/disconnect + address display)
10. Implement ExecuteTradeForm (calls execute_trade)
11. Implement WeightsForm (admin restricted) + CooldownForm + RolesDisplay
12. Global error boundary & toast system
13. Add env handling: VITE_CONTRACT_ADDRESS, VITE_RPC_URL
14. Basic tests for ScorePanel + ExecuteTradeForm
15. Build script & preview for verification

## Contract ABI Handling

- Place ABI JSON into `frontend/src/abi/TradingBot.json`
- Update when contract changes; consider generating types with `typechain-target-starknet` (future)

## Example Hook

```ts
export function useComposite() {
  return useQuery(["composite"], async () => contract.readComposite());
}
```

## Trade ID Strategy

- Default: hash(block_timestamp | wallet | nonce) to avoid collisions vs simple timestamp

## Testing Commands

```
npm run test
```

## Performance

- Keep render minimal; memo heavy tables

## Future Enhancements

- Charts via Recharts or ECharts
- Persisted queries & offline cache

---

Next: `08_ml_integration_plan.md`.
