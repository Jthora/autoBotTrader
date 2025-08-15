# Automated Progress Tracking System

This directory defines a machine-readable and human-readable framework so an AI agent (GitHub Copilot or future automation runner) can self-orient, select next actions, and update status without relying on ad-hoc chat prompts.

## Components
1. task_graph.yaml — Canonical source of truth for tasks, dependencies, statuses.
2. status_log.md — Append-only human-readable changelog snapshots.
3. metrics.json — Quantitative progress metrics (coverage, build status, counts).
4. decisions.md — Architecture/strategy decisions with ADR-style entries.
5. agent_guidelines.md — Operational rules for autonomous agents modifying the repo.
6. templates/ — Reusable snippets (ADR, task, status entry).
7. scripts/ — Automation scripts to validate & update tracking.
8. automation.md — Deep automation design & future enhancements.
9. automation scripts added: pr_comment.py, stagnation_check.py, parse_coverage.py, gas_snapshot.py

## Update Workflow (For Agents)
1. Read task_graph.yaml.
2. Identify tasks with status=PENDING and all deps COMPLETE.
3. Select highest priority (lowest `priority` number) or earliest phase.
4. Execute implementation steps.
5. Run validations (tests, lint).
6. Update task status -> IN_REVIEW or COMPLETE.
7. Append a snapshot block to status_log.md.
8. Update metrics.json (increment counters, set timestamps).

## Status Values
- PENDING: Not started.
- IN_PROGRESS: Being actively worked.
- BLOCKED: Awaiting dependency or external input (reason required).
- IN_REVIEW: Implementation done; awaiting validation or human review.
- COMPLETE: Finished & validated.
- DEFERRED: Explicitly postponed.

## Validation Hooks (Planned)
- Pre-commit script ensures no task moves to COMPLETE without notes.
- CI job parses task_graph.yaml and emits a summary comment on PRs.

## Extensibility
- Can integrate with a GitHub Action to call an AI agent periodically (cron) for unattended advancement.
- metrics.json may later ingest test coverage reports automatically.

See `agent_guidelines.md` for autonomous operation details.
