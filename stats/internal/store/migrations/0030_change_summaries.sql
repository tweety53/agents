-- 0030_change_summaries.sql: one row per change recording the run's change
-- summary -- the handoff report stating what changed and why, what was
-- verified and how, and what was deliberately left out. The self-review
-- context bundle serves it as its first source, so a deferred
-- /flow-self-review pass reads the change's own statement from the store
-- instead of finding no statement of the change anywhere in the bundle.
--
-- One row per change, last write wins (design: change-summary-last-write-wins):
-- a fix run's summary replaces the earlier one, because the change's current
-- verdict is the one the bundle serves. The UNIQUE (change_id) constraint,
-- named explicitly like its siblings, is what makes a replayed write reach
-- the row the run already recorded instead of inserting a second one --
-- the same idempotency decisions_session_key gives `flow record decision`,
-- and no session token is needed to provide it.
--
-- The summary is stored and served verbatim: it is the run's own Markdown,
-- never a render over rows, so there is no second schema for the bundle to
-- keep in step with the table.

CREATE TABLE change_summaries (
  id          BIGSERIAL PRIMARY KEY,
  change_id   BIGINT NOT NULL REFERENCES changes(id),
  recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  summary     TEXT NOT NULL,
  CONSTRAINT change_summaries_change_key UNIQUE (change_id)
);
CREATE INDEX change_summaries_change_id ON change_summaries (change_id);
