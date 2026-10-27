-- 0048_drop_stage_runs_stage_key.sql — drop the unused stage_runs.stage_key.
--
-- Migration 0008 added stage_runs.stage_key and its index for a vocabulary
-- task that never wrote or read it; the unused column misled kan-1037's
-- design, so it goes.
DROP INDEX IF EXISTS stage_runs_stage_key;
ALTER TABLE stage_runs DROP COLUMN IF EXISTS stage_key;
