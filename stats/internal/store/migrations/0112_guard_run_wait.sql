-- 0112_guard_run_wait.sql: the part of a guard run's duration_ms it spent
-- waiting rather than working -- a heavy slot's acquisition (flowd's
-- /api/v1/slots), await-rows blocked between reads -- as flow-guard's
-- telemetry posts it (guard.Reasons.Wait). The bottleneck ranking reads a
-- guard's work as duration_ms - wait_ms and ranks the wait apart. 0 on every
-- run recorded before this migration and on a post from an older binary (no
-- backfill): an unmeasured wait reads as none.
ALTER TABLE guard_runs ADD COLUMN IF NOT EXISTS wait_ms integer NOT NULL DEFAULT 0;
