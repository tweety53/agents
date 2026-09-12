-- 0023_plan_sessions.sql: a /flow-plan session is a stage run recorded
-- before its change exists. change_id becomes nullable; project_key and
-- jira_key identify the row until PutChange backfills change_id from the
-- change carrying that jira_issue. Both columns are kept after backfill
-- as the record that the row arrived first.
ALTER TABLE stage_runs ALTER COLUMN change_id DROP NOT NULL;
ALTER TABLE stage_runs ADD COLUMN project_key TEXT REFERENCES projects(project_key);
ALTER TABLE stage_runs ADD COLUMN jira_key TEXT;
ALTER TABLE stage_runs ADD CONSTRAINT stage_runs_owner_check
  CHECK (change_id IS NOT NULL OR (project_key IS NOT NULL AND jira_key IS NOT NULL));
CREATE INDEX stage_runs_unattached_plan ON stage_runs (project_key, jira_key) WHERE change_id IS NULL;
