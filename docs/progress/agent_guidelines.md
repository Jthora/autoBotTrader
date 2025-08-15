# Agent Operational Guidelines

Purpose: Enable an autonomous AI agent to advance project tasks safely and transparently.

## Core Principles
- Deterministic: All decisions traceable via task_graph.yaml + status_log.md
- Minimal Drift: Never change spec docs without updating related tasks or ADRs
- Safe Commits: Keep each commit focused on one or a small cluster of tasks
- Validation First: Do not mark COMPLETE until tests + lint pass locally

## Allowed Actions
- Add new subtasks with ids prefixed by parent (e.g., contract-skeleton-tests)
- Update statuses with justification notes
- Introduce ADRs when deviating from plan
- Adjust metrics.json after validation

## Prohibited Actions
- Deleting tasks without human-reviewed ADR
- Marking tasks COMPLETE without acceptance criteria references
- Overwriting history in status_log.md

## Decision Flow
1. Parse task_graph.yaml
2. Build ready queue (deps satisfied)
3. Select highest priority (lowest number); if tie choose earliest phase then lexical id
4. If blocked by external unknown, set BLOCKED with reason
5. Implement code changes
6. Run build/tests
7. Update metrics.json
8. Append status_log entry
9. Commit with message: `task(<id>): <summary>`

## Commit Message Conventions
- feat(contract): ...
- chore(progress): update metrics
- docs(adr): add ADR-00X

## Metrics Update Rules
- contract.compiled true only after successful build
- coverage_pct updated via parsing tooling output (future automation)

## Escalation
If repeated failure (>=3 attempts) on same task, mark BLOCKED and add note with stack trace summary.

## Example Autonomous Cycle (Pseudo)
```
ready = filter(tasks, status=PENDING && deps COMPLETE)
selected = min(ready, by priority)
implement(selected)
run_tests()
if success: status=COMPLETE
update logs & metrics
```

## Safety Checks
- Diff size heuristic: if > 800 lines changed and not a scaffold task, split into subtasks
- Forbidden secret patterns rejected (keys, mnemonics)

---
Agents should treat these guidelines as binding unless superseded by an ADR.
