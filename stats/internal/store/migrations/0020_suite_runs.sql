-- 0020_suite_runs.sql: one timed suite execution per row -- recorded by
-- `flow suite record` per run, per suite, on the machine that ran it, so
-- project configuration cites the store instead of pasting a measured
-- duration into prose (KAN-252: a pasted duration has no owner and nothing
-- checking it; the 118.63s figure this change's ticket quotes survived two
-- reworkings before anyone noticed).
--
-- suite_runs derives its project directly, the way incidents do: a suite
-- run can be recorded with no change in flight -- the guard suite and the
-- stats suites run outside any change too -- so change_id -> changes would
-- leave the ordinary case unrecordable and a nullable change_id unresolvable.
--
-- exit_code is the child's own exit status. A failed run's duration is
-- diagnostic (a timeout is visible as such), but it must never present
-- itself as a runtime figure; the reader's summary filters on this column
-- rather than the store guessing what the caller meant.
--
-- This is a NEW file rather than an edit to any applied migration, for the
-- reason 0018_guard_log.sql's header gives: migrations here are tracked by
-- filename with no checksum, so editing one already applied to a real
-- database would leave that database silently diverged from a freshly
-- migrated one.

CREATE TABLE suite_runs (
  id          BIGSERIAL PRIMARY KEY,
  project_key TEXT NOT NULL REFERENCES projects(project_key),
  suite       TEXT NOT NULL,
  host        TEXT NOT NULL,
  duration_ms BIGINT NOT NULL CHECK (duration_ms >= 0),
  exit_code   INT NOT NULL,
  ran_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX suite_runs_project_suite ON suite_runs (project_key, suite, ran_at);
