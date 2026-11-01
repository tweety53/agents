-- 0059_harvest_backfills.sql: one row per one-shot harvest back-fill that
-- has run, so flowd runs each exactly once (decision
-- gap-backfill-once-at-startup). The row is inserted in the same
-- transaction as the back-fill's deltas (decision backfill-one-transaction);
-- `stage-gap-cost` is the spend the old attribution rule dropped between
-- two stage windows.
CREATE TABLE harvest_backfills (
    name   TEXT PRIMARY KEY,
    ran_at TIMESTAMPTZ NOT NULL
);
