-- 0027_spec_lastruns.sql: one last-run timestamp per spec file -- the
-- inventory KAN-418 proposes, so a suite unrun for N changes surfaces as a
-- stat rather than by accident (seven visual suites had never run, and
-- that only surfaced because a baseline moved in task 42).
--
-- One row per (project, spec) holding the LATEST run only, not one row per
-- run the way suite_runs does: the ask is a last-run inventory, and the
-- changes-since figure the stat is built from needs the newest timestamp
-- alone. A run recorded out of order (a retried capture, a backfilled
-- timestamp) must never drag the inventory backwards, which the upsert's
-- GREATEST enforces in SQL rather than trusting every caller to.
--
-- spec_lastruns derives its project directly, the way suite_runs does: a
-- spec can be recorded with no change in flight -- a capture runs outside
-- any change too -- so change_id -> changes would leave the ordinary case
-- unrecordable and a nullable change_id unresolvable.
--
-- This is a NEW file rather than an edit to any applied migration, for the
-- reason 0020_suite_runs.sql's header gives: migrations here are tracked by
-- filename with no checksum, so editing one already applied to a real
-- database would leave that database silently diverged from a freshly
-- migrated one.

CREATE TABLE spec_lastruns (
  project_key TEXT NOT NULL REFERENCES projects(project_key),
  spec        TEXT NOT NULL,
  last_ran_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (project_key, spec)
);
