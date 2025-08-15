# Progress Automation Scripts

Utilities aiding automated agents / CI to maintain project state.

## Scripts
| Script | Description |
|--------|-------------|
| update_progress.py | Validates task graph, recomputes metrics, lists ready tasks |

## Example
```
python scripts/progress/update_progress.py --summary --write
```

Add to CI / GitHub Actions to ensure freshness on each PR.
