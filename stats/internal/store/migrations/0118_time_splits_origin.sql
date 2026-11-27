-- 0118_time_splits_origin.sql: harvest.TranscriptGaps now ends a tool gap,
-- filed under its origin (peer, task-notification), at a user line the
-- harness wrote for a subagent's message or a background task's
-- notification, where it ended human time. Every split stored before is
-- cleared, so flowd's harvest tick recomputes it under the new rule; a JSON
-- null (no transcript) stays.
UPDATE stage_runs SET time_split = NULL WHERE jsonb_typeof(time_split) = 'object';
UPDATE stage_runs SET parent_split = NULL WHERE jsonb_typeof(parent_split) = 'object';
UPDATE dispatches SET time_split = NULL WHERE jsonb_typeof(time_split) = 'object';
