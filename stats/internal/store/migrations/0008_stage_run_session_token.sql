-- 0008_stage_run_session_token.sql: the correlator that lets a later
-- harvest cycle bind a stage run to the session that marked it (KAN-172,
-- task 1; reworked from one correlator per mark to one per session in
-- task 4b -- amended in place rather than superseded by a second
-- migration, since this one is unreleased and in-branch, and a column
-- nothing has ever written has no history to preserve).
--
-- Every stage run in the live store carries session_id = NULL, because
-- Claude Code exposes no session identifier a mark could read at `stage
-- begin` time (design.md, "bind after the fact, by a correlator the
-- caller writes"). session_token is that correlator: a literal, unique
-- token a command generates once, at the start of its run, and passes
-- unchanged on every mark that run makes -- never one nonce per mark,
-- which is what this column was called and shaped before task 4b (a
-- deliberately-repeated value is not a nonce, and design.md's own
-- "one token per session, not one per mark" is why calling it one would
-- mislead every future reader). It lands in the calling session's own
-- transcript on every mark that carries it. A later harvest cycle
-- (task 2) looks for the token there and binds session_id once it finds
-- exactly one match -- and, once bound, every other stage run already
-- carrying that same token binds too (task 4b), not just the run whose
-- mark first revealed it.
--
-- Nullable, like session_id itself: a stage run this build's daemon
-- records always carries one (the CLI's own -session-token flag is
-- required from task 1 on), but the two rows already in the live store
-- predate this column and are left alone -- no backfill, per design.md's
-- "What is deliberately not changed".
--
-- The partial index restricts itself to exactly the rows a resolver ever
-- scans: session_token IS NOT NULL (there is a token to look for) AND
-- session_id IS NULL (binding has not already happened -- binding is
-- one-way, design.md's "unbinding never happens"). A row that has already
-- bound -- including one bound at insert time because its token had
-- already resolved from an earlier mark in the same run (task 4b) -- or
-- one that was never marked with a token at all, never needs to be found
-- by this index again.
ALTER TABLE stage_runs ADD COLUMN session_token TEXT;

CREATE INDEX stage_runs_unresolved_session_token ON stage_runs (session_token)
  WHERE session_token IS NOT NULL AND session_id IS NULL;

-- stage_key is task 3's vocabulary column -- a stable, declared identifier
-- for a stage, alongside the prose name stage_runs.stage already carries
-- (design.md, "a declared key, and a name that may change"). It is added
-- here, nullable, and left entirely unused by this task: nothing in task 1
-- writes or reads it. Nullable for the same no-backfill reason as
-- session_token -- the two pre-existing rows carry only a prose stage and
-- predate the key vocabulary; inventing one for them would be inventing
-- history.
ALTER TABLE stage_runs ADD COLUMN stage_key TEXT;

CREATE INDEX stage_runs_stage_key ON stage_runs (stage_key);
