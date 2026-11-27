-- 0113_guard_run_bootstrap.sql: the time scripts/flow-guard took before it
-- dispatched the guard -- hashing its sources, any `go build`, the exec --
-- from the $EPOCHREALTIME stamp it exports as FLOW_GUARD_START to the
-- binary's dispatch (cmd/flow-guard/telemetry.go's bootstrapMS). It lies
-- before the run's duration_ms, never inside it. NULL when the bootstrap
-- stamped nothing (a bash without $EPOCHREALTIME, a direct call of the
-- binary) and on every run recorded before this migration (no backfill).
ALTER TABLE guard_runs ADD COLUMN IF NOT EXISTS bootstrap_ms integer;
