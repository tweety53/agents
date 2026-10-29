-- 0053_stage_run_auto_opened.sql: a stage run the store opened itself because a
-- dispatch began while its change had no open stage run (RecordDispatch). Such a
-- run closes when its last open dispatch ends, or when an explicit stage begin
-- supersedes it.
ALTER TABLE stage_runs ADD COLUMN auto_opened boolean NOT NULL DEFAULT false;
