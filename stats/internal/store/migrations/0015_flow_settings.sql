-- 0015_flow_settings.sql: the stats app's global settings store --
-- /flow-settings' backing record for the harness-wide defaults every new
-- /flow run reads: which model implements, fixes and reviews
-- (default_model, collapsed to one field per the models-fields-collapse
-- decision) and which reviewer slots the panel dispatches by default
-- (reviewers, per review-panel-fixed-3).
--
-- Global, not per-change: the settings-store-stats-app decision makes the
-- stats app both validator and source of truth for these defaults,
-- replacing the per-change models.*/reviewPanelRoster fields the state
-- file contract used to carry. A single row -- id is a BOOLEAN CHECKed to
-- TRUE rather than a surrogate key, so a second row is a constraint
-- violation at INSERT time rather than a bug some later reader has to
-- notice by convention.
--
-- reviewers is JSONB holding a JSON array of reviewer-slot identifiers,
-- not a Postgres array or a join table: the Go layer (settings.go)
-- already validates every entry against the fixed vocabulary before a
-- write reaches this table, so there is no need for the database itself
-- to enforce membership -- the same division of labor changes.models and
-- changes.worktrees already use for validated-in-Go JSON content.
CREATE TABLE flow_settings (
  id             BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (id),
  default_model  TEXT NOT NULL,
  reviewers      JSONB NOT NULL,
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
