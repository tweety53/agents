-- 0040_dispatch_session_id.sql: the harness session a dispatch's begin ran in (KAN-967). A
-- dispatch row joins its session through stage_runs.session_token; a fix run that marks no stage
-- has no such row, so DispatchWindowsForSession also matches this column.
ALTER TABLE dispatches ADD COLUMN IF NOT EXISTS session_id text;
