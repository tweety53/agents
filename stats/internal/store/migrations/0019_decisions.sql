-- 0019_decisions.sql: one JSONB row per run recording the planner's
-- dynamic decision -- the execution mode, implementer model and review
-- panel roster it chose for that run, and the mechanical inputs it chose
-- them from.
--
-- One JSONB row rather than a column per cell (design.md's
-- decisions-jsonb-row): the decision's shape is still being tuned, so a
-- column layout would mean a migration every tuning round, and the
-- alternative of stashing it in dispatches.notes is invisible to any
-- aggregate. A GIN index keeps a query over the payload's own fields --
-- which toggle, which class, which model -- from costing a scan of the
-- whole table, the same tradeoff stage_runs.metrics and dispatches.metrics
-- already make.
--
-- (change_id, session_token) unique makes `flow record decision`
-- idempotent per run, exactly the property dispatches.key gives
-- RecordDispatch: a replayed write updates the one row a run already
-- recorded rather than inserting a second one.

CREATE TABLE decisions (
  id            BIGSERIAL PRIMARY KEY,
  change_id     BIGINT NOT NULL REFERENCES changes(id),
  session_token TEXT NOT NULL,
  recorded_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  decision      JSONB NOT NULL,
  CONSTRAINT decisions_session_key UNIQUE (change_id, session_token)
);
CREATE INDEX decisions_change_id ON decisions (change_id);
CREATE INDEX decisions_decision_gin ON decisions USING GIN (decision);
