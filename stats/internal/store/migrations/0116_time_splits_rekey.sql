-- 0116_time_splits_rekey.sql: harvest.toolKey now files a Bash call under
-- the command it runs, past a leading cd and NAME=value assignments, where
-- it took the command's first word (Bash cd for nearly every call). Every
-- split stored before is cleared, so flowd's harvest tick recomputes it under
-- the new keys; a JSON null (no transcript) stays.
UPDATE stage_runs SET time_split = NULL WHERE jsonb_typeof(time_split) = 'object';
UPDATE dispatches SET time_split = NULL WHERE jsonb_typeof(time_split) = 'object';
