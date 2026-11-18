-- 0071_verify_runs_flaky.sql: a verify run's flaky retry count -- tests that
-- passed only on their retry -- as run-verify-list's Playwright summary
-- counts them, so the dashboard's flaky-retries count is recorded rather
-- than empty (decision verify-run-flaky-column). NULL on a row recorded
-- without one, every row before this migration among them.
ALTER TABLE verify_runs ADD COLUMN IF NOT EXISTS flaky INT CHECK (flaky >= 0);
