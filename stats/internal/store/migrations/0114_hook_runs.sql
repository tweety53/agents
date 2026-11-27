-- 0114_hook_runs.sql: one row per `flow-guard hook <name>` invocation -- the
-- hook, its wall time in microseconds, its end. The hook never calls flowd
-- for it: it appends a line to the state root's hook-runs.jsonl
-- (records.HookRunsPath) and flowd ingests that file on its harvest tick
-- (cmd/flowd/hookruns.go). No project key and no worktree: a hook runs in
-- whichever session fires it, so `flow stats bottlenecks` claims a hook run
-- for the changes whose wall clock it fell inside. Like guard_runs, no row
-- is ever pruned.
CREATE TABLE hook_runs (
  id          BIGSERIAL PRIMARY KEY,
  hook        TEXT NOT NULL CHECK (hook <> ''),
  duration_us BIGINT NOT NULL CHECK (duration_us >= 0),
  ran_at      TIMESTAMPTZ NOT NULL
);

CREATE INDEX hook_runs_ran_at ON hook_runs (ran_at);
