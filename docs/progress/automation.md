# Deep Automation Design

## Goals
- Self-maintaining progress status with minimal manual intervention.
- Provide machine-readable hooks for AI agents & CI pipelines.
- Preserve auditability (append-only logs, explicit diffs).

## Layers Implemented
1. Data Layer: `task_graph.yaml`, `metrics.json`
2. Logic Layer: `scripts/progress/update_progress.py`
3. CI Layer: `.github/workflows/progress.yml` auto-updates metrics on relevant changes
4. IDE Layer: VS Code task `Progress: Validate & Update`

## Recommended Extensions
- YAML language support
- GitLens for diff review

## Extended Ideas (Future)
| Idea | Description | Benefit |
|------|-------------|---------|
| PR Comment Bot | Summarize ready tasks & status deltas | Faster reviews |
| Auto Status Entry | Append status_log.md entry from script | Historical snapshots |
| Coverage Integration | Parse test coverage & inject into metrics.json | Quantitative quality tracking |
| Gas Report | Add gas usage per commit to metrics.json | Track optimization trends |
| AI Plan Validator | LLM checks that implementation matches spec docs | Early drift detection |

## Safety Mechanisms
- CI only auto-commits on non-main branches.
- Main branch remains curated (manual merge or separate workflow).
- Validation rejects cyclic dependencies.

## Agent Loop (Augmented)
1. Run update_progress.py to get READY tasks.
2. Pick first READY.
3. Implement code, run tests.
4. Mark task IN_REVIEW -> after CI green -> COMPLETE.
5. Create status_log entry (future automation step).

## Integration Points
- Pre-commit hook (optional) could invoke validation.
- GitHub Action for nightly run to detect stagnation (no status change > X days).

---
This document evolves as deeper automation layers are added.
