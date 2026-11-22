-- 0097_confidence_estimates.sql: one row per confidence estimate below 100% -- an agent's answer
-- to "how confident are you, 1-100%, that this plan is executed as written, and what would move
-- you to 100%?" (KAN-1130), kept in full to analyse later.
--
-- stage is where the estimate was made: plan (the plan check's cold reader; disposition says
-- whether the operator had the gaps closed -- raised -- or deferred them -- later) or implement
-- (an implementer at the start of its dispatch; task is its dispatch key's task ids, and it
-- carries no disposition because an implementer never stops for the number).
--
-- gaps is the gap list verbatim, one element per gap. tags is free-form: the filing step lists
-- the tags already in this table and reuses one that fits, or adds a new one -- no vocabulary
-- table and no CHECK, deliberately, until the recorded data shows which tags are worth keeping.
--
-- change is text, not a changes(id) FK, as in self_review_findings. No uniqueness constraint: a
-- re-dispatched implementer's second estimate is a second response.

CREATE TABLE confidence_estimates (
  id          BIGSERIAL PRIMARY KEY,
  project_key TEXT NOT NULL REFERENCES projects(project_key),
  change      TEXT NOT NULL,
  stage       TEXT NOT NULL CHECK (stage IN ('plan','implement')),
  task        TEXT,
  score       INT NOT NULL CHECK (score BETWEEN 1 AND 99),
  gaps        TEXT[] NOT NULL,
  tags        TEXT[] NOT NULL,
  disposition TEXT CHECK (disposition IN ('raised','later')),
  recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK ((stage = 'plan') = (disposition IS NOT NULL)),
  CHECK ((stage = 'implement') = (task IS NOT NULL))
);

CREATE INDEX confidence_estimates_tags ON confidence_estimates USING GIN (tags);
