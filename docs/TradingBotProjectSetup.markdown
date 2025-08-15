# GitHub Copilot Guide for Developing a Decentralized Trading Bot dApp on Paradex Chain

This document outlines the essential information GitHub Copilot needs to assist in developing a **decentralized application (dApp)** on the **Paradex Chain**, a Starknet-based Layer 2 blockchain (referred to as the "SuperChain"). The dApp is an autonomous trading bot that predicts market movements using **astrology** (planetary positions from ephemeris data) and **gravimetrics** (lunar tidal forces), with hooks for future **ML/PPO (Proximal Policy Optimization)** integration from another project. The bot trades perpetual futures/options on the **Paradex SuperDEX**, leveraging the chain’s high throughput (>1,000 TPS), low fees (<$0.01), and liquidity ($4.38B volume, 75 markets). The project uses **Cairo** for on-chain smart contracts, **Golang** or **TypeScript** for off-chain APIs, and **React/TypeScript** for the frontend, ensuring a fully decentralized backend with a user-friendly interface.

## Project Overview
- **Objective**: Build a dApp that autonomously trades on SuperDEX using astrology (e.g., Mercury retrograde) and gravimetrics (e.g., lunar gravity) to generate market predictions, with modular support for future ML/PPO integration.
- **Components**:
  - **On-Chain**: Cairo smart contract to execute trades based on prediction signals, integrated with SuperDEX’s unified margin system.
  - **Off-Chain**: Golang/TypeScript API to fetch/process astrology and gravimetric data (e.g., via Astrodienst, NOAA).
  - **Frontend**: React/TypeScript UI for trade monitoring and bot configuration.
  - **ML/PPO Hooks**: Modular design to incorporate ML-generated prediction scores via oracles.
- **Decentralization**: Core trading logic runs on CairoVM, ensuring trustless execution with ZK-Rollup settlement on Ethereum.
- **Alignment**: Leverages SuperChain’s scalability, SuperDEX’s liquidity, and Paradex’s 2025 AI-driven vault plans.

## Development Environment
- **Tools**:
  - **VS Code with GitHub Copilot**: For code generation, suggestions, and documentation.
  - **Starknet Tools**: Hardhat (v2.22.0+), Starkli (v0.4.0) for Cairo contract development/deployment.
  - **Node.js/NPM**: v20+ for React/TypeScript frontend.
  - **Golang**: v1.21+ for off-chain APIs.
  - **MetaMask**: For wallet interaction with SuperChain.
  - **Starknet Sepolia**: Testnet for development (use Paradex’s RPC endpoints, e.g., Juno, Pathfinder).
- **Setup Commands**:
  - Initialize Hardhat: `npm install --save-dev hardhat starkli`.
  - Initialize Golang: `go mod init trading-bot`.
  - Initialize Vite/React/TypeScript: `npm create vite@latest trading-bot -- --template react-ts`.
  - Configure Starknet in `hardhat.config.js` with Sepolia RPC.

**Copilot Prompt**:
```plaintext
Generate a hardhat.config.js for deploying Cairo smart contracts on Paradex Chain (Starknet-based). Include Starknet Sepolia testnet configuration and a Golang module setup for fetching external data.
```

## Project Structure
- **Directory Layout**:
  ```
  trading-bot/
  ├── contracts/                # Cairo smart contracts
  │   └── TradingBot.cairo      # Trading logic for predictions and trades
  ├── api/                      # Golang/TypeScript APIs
  │   ├── main.go              # Golang REST API for data fetching
  │   └── data.ts              # TypeScript fallback for data processing
  ├── frontend/                 # Vite React/TypeScript frontend
  │   ├── src/
  │   │   ├── App.tsx          # Main UI component
  │   │   └── starknet.ts      # Starknet.js for contract interaction
  ├── tests/                    # Hardhat test scripts
  │   └── testBot.js           # Contract unit tests
  ├── scripts/                  # Deployment scripts
  │   └── deploy.ts            # Starkli deployment script
  ├── ml/                       # Future ML/PPO integration
  │   └── predict.py           # Placeholder for PPO model
  └── README.md                 # Project documentation
  ```

**Copilot Prompt**:
```plaintext
Create a project structure for a decentralized trading bot dApp on Paradex Chain using Cairo for smart contracts, Golang for off-chain APIs, and React/TypeScript for the frontend. Include files for fetching astrology and gravimetric data, with placeholders for ML/PPO integration.
```

## Cairo Smart Contract
- **Purpose**: Implement trading logic to:
  - Store astrology/gravimetric prediction scores.
  - Execute trades (e.g., buy/sell BTC perps) based on scores (>50 triggers buy).
  - Emit events for trade outcomes (e.g., PNL).
- **Requirements**:
  - Use Cairo 2.11.x for Starknet compatibility.
  - Integrate with SuperDEX’s unified margin system (mock interface for now).
  - Optimize for gas efficiency using ZK-Rollups.
- **Key Functions**:
  - `set_prediction_score(astrology_score, gravity_score)`: Calculate weighted prediction score.
  - `execute_trade(trade_id, amount)`: Execute trade if score exceeds threshold.
  - `TradeExecuted` event: Log trade details.

**Copilot Prompt**:
```plaintext
Write a Cairo smart contract for a trading bot on Paradex Chain. Include functions to receive astrology and gravimetric data, calculate a market prediction score, and execute trades on SuperDEX perpetual futures. Optimize for gas efficiency and include event emissions.
```

## Off-Chain API (Golang/TypeScript)
- **Purpose**: Fetch and process:
  - **Astrology Data**: Planetary positions (e.g., via Astrodienst API).
  - **Gravimetric Data**: Lunar tidal forces (e.g., via NOAA API).
  - Send processed scores to the Cairo contract via Starknet.js or direct calls.
- **Requirements**:
  - Golang preferred (per Paradex’s stack), TypeScript as fallback.
  - REST API with endpoints: `/astrology`, `/gravimetrics`, `/predict`.
  - Concurrent processing (goroutines in Golang, async/await in TypeScript).
- **Integration**: Use Starknet.js (v7.0.1) to call Cairo contract’s `set_prediction_score`.

**Copilot Prompt**:
```plaintext
Generate a Golang REST API to fetch astrology data (planetary positions from an ephemeris API) and gravimetric data (lunar tidal forces from NOAA). Include endpoints to process data into a prediction score and send it to a Cairo smart contract on Paradex Chain. Provide TypeScript alternative.
```

## React/TypeScript Frontend
- **Purpose**: Provide a UI to:
  - Display trade history and PNL.
  - Configure bot settings (e.g., astrology/gravimetric weights).
  - Interact with Cairo contract (e.g., trigger trades).
- **Requirements**:
  - Use Vite/React/TypeScript for fast development.
  - Integrate Starknet.js for wallet (MetaMask) and contract interaction.
  - Deployable to Vercel (Web2 hosting, but logic remains on-chain).
- **Components**:
  - `App.tsx`: Main UI with trade dashboard.
  - `starknet.ts`: Handles contract calls and wallet connection.

**Copilot Prompt**:
```plaintext
Create a Vite React/TypeScript frontend for a trading bot dApp on Paradex Chain. Include components for displaying trade history, configuring astrology/gravimetric weights, and interacting with a Cairo smart contract using Starknet.js.
```

## Testing and Deployment
- **Testing**:
  - Unit tests for Cairo contract using Hardhat (test prediction logic, trade execution).
  - API tests using Go’s `testing` package or Jest for TypeScript.
  - Frontend tests with React Testing Library.
- **Deployment**:
  - Deploy Cairo contract to Starknet Sepolia: `starkli deploy`.
  - Deploy frontend to Vercel: `vercel deploy`.
  - Connect APIs to contract via Starknet.js RPC calls.
- **Copilot Prompt**:
```plaintext
Generate Hardhat test scripts for a Cairo trading bot contract on Paradex Chain. Include unit tests for prediction score calculation and trade execution. Provide a Starkli deployment script for Starknet Sepolia.
```

## ML/PPO Integration Hooks
- **Purpose**: Prepare for future integration of an ML/PPO model (from another project) to enhance market predictions.
- **Approach**:
  - Off-chain: Python/TensorFlow script to train PPO model on market, astrology, and gravimetric data.
  - On-chain: Cairo function to accept ML scores via an oracle (e.g., Chainlink).
- **Requirements**:
  - Modular contract design with a placeholder function (e.g., `set_ml_score`).
  - API endpoint to feed ML scores to the contract.
- **Copilot Prompt**:
```plaintext
Generate a Python script for a PPO model to predict market movements using astrology and gravimetric data. Include a placeholder Cairo function to receive ML scores via an oracle for a trading bot on Paradex Chain.
```

## Optimization and Documentation
- **Optimization**:
  - Cairo: Minimize gas costs with efficient storage (e.g., LegacyMap) and ZK-Rollup-friendly logic.
  - Golang: Use goroutines for concurrent API calls.
  - TypeScript: Optimize async operations for low latency.
- **Documentation**:
  - Inline comments for all code.
  - README.md with setup, usage, and ML integration instructions.
- **Copilot Prompt**:
```plaintext
Add inline comments to a Cairo trading bot contract and generate a README.md for a dApp project, including setup, usage, and ML integration plans.
```

## Key Considerations for Copilot
- **Cairo Syntax**: Use Cairo 2.11.x, Starknet-compatible, with `#[starknet::contract]` annotations.
- **Golang Concurrency**: Leverage goroutines and channels for API performance.
- **Starknet.js Integration**: Use v7.0.1 for contract/wallet interactions (RPC 0.8).
- **Data Sources**: Mock Astrodienst/NOAA APIs for now, with placeholders for real endpoints.
- **ML Modularity**: Ensure Cairo contract and API are extensible for PPO scores.
- **Error Handling**: Include robust checks (e.g., `assert_le`, `assert_nn` in Cairo; HTTP error codes in Golang).

## Expected Output
- **Cairo Contract**: Trading bot with prediction and trade execution logic, gas-optimized.
- **Golang/TypeScript API**: REST endpoints for astrology/gravimetric data and contract interaction.
- **React Frontend**: User-friendly UI for bot monitoring, deployable to Vercel.
- **Tests**: Comprehensive unit tests for contract, API, and frontend.
- **ML Hooks**: Placeholder functions for PPO integration.
- **Documentation**: Clear README and inline comments for maintainability.