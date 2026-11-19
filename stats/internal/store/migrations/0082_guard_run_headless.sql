-- 0082_guard_run_headless.sql
--
-- The headless model call a guard run made (pair-judge's judge,
-- plan-check's decider): the role it ran as, its pair, cost and duration.
-- NULL on every run that made none, and on every run recorded before this
-- migration (no backfill). The Role costs view counts each as a run of its
-- role.
ALTER TABLE guard_runs
    ADD COLUMN IF NOT EXISTS headless_role text,
    ADD COLUMN IF NOT EXISTS headless_model text,
    ADD COLUMN IF NOT EXISTS headless_effort text,
    ADD COLUMN IF NOT EXISTS headless_cost_usd double precision,
    ADD COLUMN IF NOT EXISTS headless_ms integer;
