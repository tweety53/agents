-- 0059_change_files.sql: the two files of a change that outlive its
-- worktree (KAN-1073) -- tasks.md, the board's plan of a finished change, and
-- self-review-context.md, a deferred /flow-self-review's whole input. The
-- change directory is never committed, so `flow change save` writes both here
-- once, in run 2's cleanup, before the worktrees are removed.
CREATE TABLE IF NOT EXISTS change_files (
    id BIGSERIAL PRIMARY KEY,
    change_id BIGINT NOT NULL REFERENCES changes(id),
    path TEXT NOT NULL CHECK (path <> ''),
    body TEXT NOT NULL,
    saved_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (change_id, path)
);
