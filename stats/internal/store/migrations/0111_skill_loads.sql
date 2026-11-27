-- 0111_skill_loads.sql: one row per instruction file a session loaded -- a
-- Read of a skills/, rules/ or docs/ file of the agents repository, or a
-- Skill call of its SKILL.md -- written by the flow-skill-load hook
-- (stats/internal/guard/hookskillload.go). `flow-guard prune-candidates`
-- counts them per path over its window to list the skill files no session
-- loads, which guard_runs cannot show: it records guards, never the files a
-- session reads.
--
-- path is relative to the agents checkout the file lies in, so a load from a
-- worktree and one from the main checkout count as the same file. No project
-- key: these files belong to the agents repository whichever project's
-- session loads them. session_id is the harness session's, '' when the
-- payload carries none. Like guard_runs, no row is ever pruned.
CREATE TABLE skill_loads (
  id         BIGSERIAL PRIMARY KEY,
  path       TEXT NOT NULL CHECK (path <> ''),
  session_id TEXT NOT NULL DEFAULT '',
  loaded_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX skill_loads_loaded_at ON skill_loads (loaded_at);
