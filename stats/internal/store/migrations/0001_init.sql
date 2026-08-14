-- 0001_init.sql: the projects and changes tables.
--
-- The two table definitions are copied verbatim from design.md's "Data
-- model" section. state_rank is an addition this task needs to enforce the
-- monotonic-state rule inside the changes table's own upsert, in SQL, per
-- the plan's requirement that a violation is refused by the store itself.

CREATE TABLE projects (
  project_key        TEXT PRIMARY KEY,
  main_checkout_path TEXT NOT NULL,
  first_seen         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE changes (
  id                  BIGSERIAL PRIMARY KEY,
  project_key         TEXT NOT NULL REFERENCES projects(project_key),
  name                TEXT NOT NULL,
  state               TEXT NOT NULL,
  branch              TEXT,
  worktrees           JSONB NOT NULL DEFAULT '{}'::jsonb,
  artifact_url        TEXT,
  jira_issue          TEXT,
  planning_effort     TEXT,
  models              JSONB,
  review_panel_roster TEXT,
  pr_url              TEXT,
  updated_at          TIMESTAMPTZ NOT NULL,
  updated_by          TEXT NOT NULL,
  UNIQUE (project_key, name)
);

-- state_rank orders the three pipeline states so a plain integer comparison
-- enforces the monotonic-state rule inside a single upsert statement. An
-- unrecognised state ranks below every known state so it can never appear
-- to move the record forward.
CREATE FUNCTION state_rank(state TEXT) RETURNS INT AS $$
  SELECT CASE state
    WHEN 'STARTED' THEN 0
    WHEN 'IN_PROGRESS' THEN 1
    WHEN 'FINISHED' THEN 2
    ELSE -1
  END
$$ LANGUAGE sql IMMUTABLE;
