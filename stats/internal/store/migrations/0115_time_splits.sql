-- 0115_time_splits.sql: the transcript time split (harvest.TimeSplit: llm,
-- tool per tool, human) flowd's harvest tick (cmd/flowd/timesplits.go)
-- computes once per ended row and stores, so `flow stats trace`, the per-run
-- route and `flow stats bottlenecks` read it instead of re-reading
-- transcripts.
--
-- stage_runs.transcript_path links a stage run to the transcript of the
-- agent that marked it, derived from its cc-<session id>-<agent> session
-- token (harvest.TokenTranscriptPath): the session's own <session>.jsonl for
-- main, the agent's subagents/agent-<id>.jsonl otherwise. NULL on an mf-/fi-
-- token (no session id to derive it from), on a cc- token whose transcript
-- was never found, and on every run before this migration until the tick
-- links it.
--
-- time_split, on stage_runs and dispatches, is the split of that transcript
-- (a dispatch's: its agent's) inside the row's [started_at, ended_at). SQL
-- NULL is not computed yet; JSON null is computed and no transcript found,
-- so the tick never looks again. A change of either instant (a retime, a
-- restamp) clears it through the trigger below, and the next tick
-- recomputes it on the new window.
ALTER TABLE stage_runs ADD COLUMN IF NOT EXISTS transcript_path text;
ALTER TABLE stage_runs ADD COLUMN IF NOT EXISTS time_split jsonb;
ALTER TABLE dispatches ADD COLUMN IF NOT EXISTS time_split jsonb;

CREATE OR REPLACE FUNCTION reset_time_split() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  IF NEW.started_at IS DISTINCT FROM OLD.started_at OR NEW.ended_at IS DISTINCT FROM OLD.ended_at THEN
    NEW.time_split := NULL;
  END IF;
  RETURN NEW;
END $$;

CREATE TRIGGER stage_runs_reset_time_split BEFORE UPDATE OF started_at, ended_at ON stage_runs
  FOR EACH ROW EXECUTE FUNCTION reset_time_split();
CREATE TRIGGER dispatches_reset_time_split BEFORE UPDATE OF started_at, ended_at ON dispatches
  FOR EACH ROW EXECUTE FUNCTION reset_time_split();

CREATE INDEX stage_runs_time_split_pending ON stage_runs (id)
  WHERE time_split IS NULL AND session_token LIKE 'cc-%';
CREATE INDEX dispatches_time_split_pending ON dispatches (id)
  WHERE time_split IS NULL AND ended_at IS NOT NULL AND agent_id IS NOT NULL AND agent_id <> 'inline';
