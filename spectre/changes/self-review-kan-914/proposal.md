# self-review-kan-914

## Why

KAN-914 (flow-fix): `scripts/workspace.sh create <id>` brings up a fresh `flow_<id>` database, and
a daemon started on it migrates the schema but holds no `projects` row. Every record table but
`changes` and plan stage runs references `projects(project_key)` — `guard_runs`, `hazards`,
`incidents`, `suite_runs`, `spec_lastruns`, `self_review_findings` and the non-plan `stage_runs`
rows — and only `PutChange` and `insertPlanStageRun` bootstrap the row (`ensureProject`,
`stats/internal/store/changes.go`). So the first such write against a workspace daemon is refused:
on a fresh migrated store at b76c2cf1, `RecordGuardRun(ctx, "ws-0000beef", …)` returns
`store: project not found: ws-0000beef`. kan-875's live verification had to seed the row by hand
with `flow state set` before its self-review finding writes went through.

## What changes

- A workspace daemon — one whose `FLOWD_DSN` names a `flow_<id>` database that is not a
  `_uitest` one — seeds its working directory's project row at startup, after its migrations, so
  record writes against it succeed from the first one. A failed seed is logged and the daemon
  starts as before.
- `store.Store` gains `EnsureProject`, the existing `ensureProject` bootstrap in its own
  transaction.
- The dev daemon on the `flow` database and the UI-test daemon on `flow_<id>_uitest` behave
  exactly as before.
