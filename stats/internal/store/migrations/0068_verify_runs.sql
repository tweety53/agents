-- 0068_verify_runs.sql: one row per verify run (run-verify-list or
-- recapture) of a change -- its wall-clock seconds and test counts, so a
-- run's cost is visible.
CREATE TABLE IF NOT EXISTS verify_runs (
    id          BIGSERIAL PRIMARY KEY,
    change_id   BIGINT NOT NULL REFERENCES changes(id),
    kind        TEXT NOT NULL CHECK (kind IN ('run-verify-list', 'recapture')),
    label       TEXT NOT NULL,
    seconds     INT NOT NULL,
    tests       INT NOT NULL,
    failed      INT NOT NULL,
    exit_code   INT NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS verify_runs_change_id ON verify_runs (change_id);
