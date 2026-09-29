-- 0032_guard_runs.sql: one row per invocation of a flow-guard guard -- the
-- "does this guard ever catch anything" record the flow-health view reads.
--
-- guard_verdicts (0018) holds the verdict lines the two guards that record
-- one choose to keep, against a change. That cannot answer "which guards
-- fire at all": the other guards record nothing, so a guard that never
-- catches anything is indistinguishable from one that never ran. This
-- table is written by the flow-guard binary itself, for every guard it
-- dispatches, so the answer no longer depends on each guard remembering to
-- record.
--
-- project_key is stored directly, as incidents does, rather than derived
-- through a change: a guard runs in a worktree whose change the binary has
-- no way to name, and the worktree path is kept verbatim for a reader who
-- wants to trace one back. The FK is what makes an unknown project a
-- refused write -- a guard run inside a test fixture's throwaway repository
-- names a project no store holds, and is dropped rather than recorded.
--
-- outcome is the binary's classification of the exit code, closed at the
-- store: "clear" (exit 0), "cannot-answer" (the guard's own could-not-answer
-- code) or "fired" (anything else). exit_code is kept beside it because
-- the classification is per-guard knowledge the binary holds and a reader
-- auditing it needs the raw value.

CREATE TABLE guard_runs (
  id          BIGSERIAL PRIMARY KEY,
  project_key TEXT NOT NULL REFERENCES projects(project_key),
  guard       TEXT NOT NULL,
  worktree    TEXT NOT NULL,
  exit_code   INT  NOT NULL,
  outcome     TEXT NOT NULL CHECK (outcome IN ('clear', 'fired', 'cannot-answer')),
  duration_ms INT  NOT NULL CHECK (duration_ms >= 0),
  recorded_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX guard_runs_recorded_at ON guard_runs (recorded_at);
CREATE INDEX guard_runs_project_key ON guard_runs (project_key);
