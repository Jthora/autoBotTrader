# Building & Running Components

## Prerequisites
- Go (>=1.21)
- Node.js (>=20)
- Scarb (Cairo package manager) & Starknet toolchain

## Contracts
```
cd contracts
scarb build
```

## Go Service
```
cd api
go build ./...
./cmd/server &
curl -s localhost:8080/health
```

## Frontend
```
cd frontend
npm install
npm run build
npm run preview
```

## Progress Automation
```
python3 scripts/progress/update_progress.py --summary --write --append-log
```

## Notes
- Install Go via homebrew: `brew install go`.
- Install Scarb: follow https://docs.swmansion.com/scarb/.
