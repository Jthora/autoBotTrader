# Contributing

## Workflow
1. Pick a READY task from `docs/progress/task_graph.yaml`.
2. Create a branch: `feat/<task-id>-short-description`.
3. Implement; keep commits focused.
4. Run local checks.
5. Update task status & metrics via progress script.
6. Open PR using template.

## Commit Messages
Conventional style (examples):
- feat(contract): add cooldown logic
- fix(go): handle provider timeout
- docs: update roadmap
- chore(progress): auto-update metrics

## Code Style
- Python: Ruff (default rules) / 4-space indent.
- Go: gofmt / go vet.
- TS/JS: Prettier + ESLint (recommended config once frontend added).
- Cairo: `scarb fmt` when available.

## Testing
Add or update tests with behavior changes. Do not mark tasks COMPLETE without tests unless explicitly noted in acceptance criteria.

## Security
Never commit secrets. Use env vars and .env (ignored).
