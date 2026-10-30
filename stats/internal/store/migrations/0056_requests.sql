-- 0056_requests.sql: one row per request a /flow-interactive session works --
-- the ask, its brainstorm reason, receipt, end, outcome and the landed range --
-- and the request each of the session's stage runs worked on (KAN-1061). A
-- request's state is derived on read (requestState, requests.go), never stored.
CREATE TABLE IF NOT EXISTS requests (
    id BIGSERIAL PRIMARY KEY,
    change_id BIGINT NOT NULL REFERENCES changes(id),
    n INT NOT NULL CHECK (n > 0),
    title TEXT NOT NULL DEFAULT '',
    brainstorm_reason TEXT NOT NULL DEFAULT '',
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at TIMESTAMPTZ,
    outcome TEXT CHECK (outcome IN ('landed', 'answered', 'dropped')),
    commit_sha TEXT,
    base_sha TEXT,
    UNIQUE (change_id, n),
    CHECK ((ended_at IS NULL) = (outcome IS NULL))
);

ALTER TABLE stage_runs ADD COLUMN IF NOT EXISTS request_n INT;
