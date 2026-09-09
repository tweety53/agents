-- 0019_hazards.sql: proactive, per-project warnings a dispatch bundle
-- carries before they cost time -- the record KAN-452's motivating
-- self-review found missing, when hazards that predicted an incident lived
-- only in the operator's memory notes, so the first conductor dispatch of
-- that run did not carry them and the incident happened anyway.
--
-- applies is a closed vocabulary matching how the pipeline classifies a
-- change (worktree-resolution's one repository vs satellites): 'all'
-- injects into every bundle, 'cross-repo'/'single-repo' only into bundles
-- whose caller passed that shape. name is a stable identifier the CLI and
-- the retire path address rows by; UNIQUE makes a re-add a clean refusal,
-- not a duplicate row. Rows are retired (active=false), never deleted --
-- the same disposition incidents get -- so the record of what warnings a
-- project carried when an incident happened anyway survives.
--
-- This is a NEW file rather than an edit to any applied migration, for
-- the reason 0018_guard_log.sql's header gives: migrations here are
-- tracked by filename with no checksum, so editing one already applied to
-- a real database would leave that database silently diverged from a
-- freshly migrated one.

CREATE TABLE hazards (
  id          BIGSERIAL PRIMARY KEY,
  project_key TEXT NOT NULL REFERENCES projects(project_key),
  name        TEXT NOT NULL,
  body        TEXT NOT NULL,
  applies     TEXT NOT NULL CHECK (applies IN ('all','cross-repo','single-repo')),
  active      BOOLEAN NOT NULL DEFAULT true,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (project_key, name)
);

CREATE INDEX hazards_project_key ON hazards (project_key);
