-- 0069_dispatch_stage_run.sql: link a dispatch to the stage run holding its
-- begin instant (decision stage-link-backfill-migration, KAN-1132).
--
-- The rule: the candidates are the change's stage runs whose window
-- [started_at, ended_at) holds the begin instant, an open run's window
-- running on without end. Exactly one candidate under the dispatch's own
-- session token wins; else exactly one candidate in total wins; else the
-- link stays NULL -- a tie is never guessed onto a run. The begin instant,
-- never "the run open now", decides, so a journal replay links where its
-- begin happened. recordDispatchTx calls the same function at record time.
CREATE OR REPLACE FUNCTION dispatch_stage_run(p_change BIGINT, p_token TEXT, p_at TIMESTAMPTZ)
RETURNS BIGINT LANGUAGE sql STABLE AS $$
  WITH candidates AS (
    SELECT id, session_token FROM stage_runs
    WHERE change_id = p_change AND started_at <= p_at AND (ended_at IS NULL OR p_at < ended_at)
  ), same_token AS (
    SELECT id FROM candidates WHERE session_token = p_token
  )
  SELECT CASE
    WHEN (SELECT COUNT(*) FROM same_token) = 1 THEN (SELECT id FROM same_token)
    WHEN (SELECT COUNT(*) FROM candidates) = 1 THEN (SELECT id FROM candidates)
  END
$$;

-- The back-fill: idempotent, it touches only NULL links and only those the
-- rule resolves, so a re-run changes no row.
UPDATE dispatches d SET stage_run_id = dispatch_stage_run(d.change_id, d.session_token, d.started_at)
WHERE d.stage_run_id IS NULL AND dispatch_stage_run(d.change_id, d.session_token, d.started_at) IS NOT NULL;
