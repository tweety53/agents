-- 0054_experiments.sql: allowed-pairs experiments (KAN-1067). An experiment
-- replaces whole unmarked rows of skills/flow-contracts/model-policy.md's
-- Allowed pairs table -- cells maps role id to the full list of pairs -- from
-- started_at until it is closed with a verdict and a summary. A closed
-- experiment is never served as an override.
--
-- experiment_roles holds one row per role of an open experiment, so its
-- primary key is the one-open-experiment-per-role guarantee: two concurrent
-- opens naming one role cannot both commit. Closing deletes the rows.
CREATE TABLE IF NOT EXISTS experiments (
	id         BIGSERIAL PRIMARY KEY,
	name       TEXT NOT NULL,
	cells      JSONB NOT NULL,
	started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	closed_at  TIMESTAMPTZ,
	verdict    TEXT CHECK (verdict IN ('positive', 'negative', 'inconclusive')),
	summary    TEXT NOT NULL DEFAULT '',
	CHECK ((closed_at IS NULL) = (verdict IS NULL))
);

CREATE TABLE IF NOT EXISTS experiment_roles (
	role          TEXT PRIMARY KEY,
	experiment_id BIGINT NOT NULL REFERENCES experiments(id)
);
