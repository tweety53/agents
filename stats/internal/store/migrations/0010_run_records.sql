-- 0010_run_records.sql: a change's derived run record -- every subagent
-- dispatch and every review-panel finding -- as store rows rather than as
-- files in a worktree.
--
-- These records were Markdown until now, and a file has a path no two
-- components agreed on: "this change has no such record" and "the record is
-- somewhere I did not look" were the same string, one of them a silent data
-- loss. A row attached to changes(id) has no path to disagree about, and it
-- survives the worktree that produced it being removed.
--
-- dispatches is ONE table, not three. The SDD ledger's row, the
-- task -> commit -> model record and the per-dispatch cost bag are three
-- views of the same dispatch, so splitting them would give three tables
-- sharing a primary key and joined on every read. They are columns rather
-- than one JSONB payload because model, role, task_id and started_at are
-- exactly what a cross-change query filters on, and a payload buries them.
--
-- stage_run_id is nullable: a dispatch made outside a marked stage still
-- records, and "no stage run" is an absence, never an unknown one -- the
-- same rule stage_runs.repo_root already follows.
--
-- (change_id, seq) is unique and the constraint is NAMED, exactly as
-- stage_runs_attempt_key is and for the same reason: RecordDispatch
-- allocates the next seq inside a single INSERT ... SELECT and uses this
-- constraint as its race detector, so the name is part of the Go code's
-- contract with this schema rather than an identifier Postgres happened to
-- pick. Two concurrent writers that compute the same next seq collide here
-- and the loser retries, instead of one silently overwriting the other.
--
-- metrics carries the harvester's bag in the same shape stage_runs.metrics
-- does, and is merged into by jsonb_deep_add (0005_jsonb_deep_add.sql) for
-- the same reason that function exists at all: the harvester writes token
-- figures under a nested "tokens" object once per batch, as a delta rather
-- than a total, so a shallow `||` would discard whatever an earlier batch
-- recorded beside them and a plain recursive merge would replace the
-- figures themselves with only the last batch's. NOT NULL DEFAULT '{}' is
-- what makes the merge total -- there is no null bag to special-case on
-- the way in.
--
-- (change_id, ref) is unique on findings -- per change, NOT per round. A
-- fix round updates a finding's status in place, so the record of a
-- change's findings can never become cumulative, which is what makes a
-- count guard over the rendered record satisfiable at all rather than
-- merely correct today.
--
-- findings.dispatch_id records the slot that raised a finding where that is
-- known; it is nullable because the finding wire shape carries no dispatch
-- reference, so today's write path leaves it unset rather than inventing
-- one.
--
-- The indexes are the columns a cross-change question filters on -- which
-- model ran which task, which slot raises which severity -- plus a GIN
-- index on metrics matching stage_runs_metrics_gin, so a query over the
-- cost bag costs nothing proportional to the table's full size.

CREATE TABLE dispatches (
  id            BIGSERIAL PRIMARY KEY,
  change_id     BIGINT NOT NULL REFERENCES changes(id),
  stage_run_id  BIGINT REFERENCES stage_runs(id),
  seq           INT  NOT NULL,
  task_id       TEXT,
  role          TEXT NOT NULL,
  slot          TEXT,
  model         TEXT NOT NULL,
  commit_sha    TEXT,
  outcome       TEXT,
  session_token TEXT,
  started_at    TIMESTAMPTZ NOT NULL,
  ended_at      TIMESTAMPTZ,
  metrics       JSONB NOT NULL DEFAULT '{}'::jsonb,
  notes         TEXT,
  CONSTRAINT dispatches_seq_key UNIQUE (change_id, seq)
);

CREATE INDEX dispatches_change_id  ON dispatches (change_id);
CREATE INDEX dispatches_model      ON dispatches (model);
CREATE INDEX dispatches_role       ON dispatches (role);
CREATE INDEX dispatches_task_id    ON dispatches (task_id);
CREATE INDEX dispatches_started_at ON dispatches (started_at);
CREATE INDEX dispatches_metrics_gin ON dispatches USING GIN (metrics);

CREATE TABLE findings (
  id          BIGSERIAL PRIMARY KEY,
  change_id   BIGINT NOT NULL REFERENCES changes(id),
  dispatch_id BIGINT REFERENCES dispatches(id),
  ref         TEXT NOT NULL,
  round       INT  NOT NULL,
  slot        TEXT NOT NULL,
  severity    TEXT NOT NULL,
  location    TEXT,
  note        TEXT NOT NULL,
  status      TEXT NOT NULL,
  reproducer  TEXT,
  CONSTRAINT findings_ref_key UNIQUE (change_id, ref)
);

CREATE INDEX findings_change_id ON findings (change_id);
CREATE INDEX findings_severity  ON findings (severity);
CREATE INDEX findings_status    ON findings (status);
CREATE INDEX findings_slot      ON findings (slot);
