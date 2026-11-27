-- 0117_parent_splits.sql: stage_runs.parent_split, the parent session's
-- time split (harvest.TimeSplit: llm, tool per tool, human) inside a stage
-- run's [started_at, ended_at), so `flow stats bottlenecks`' parent rows sum
-- it in SQL.
--
-- The parent transcript is the session's own <session>.jsonl, found from the
-- run's session_id (harvest.SessionTranscriptPath), whichever agent marked
-- the run -- not from its session token, which on a /flow-interactive run
-- names the subagent that marked it, and on an mf-/fi- token names no
-- transcript at all. One rule for overlap: each stage run counts the
-- transcript of the one session id it carries, once; two parent sessions
-- overlapping in time are two stage runs' worth, and two runs of one session
-- that overlap each count the gaps they share, as overlapping stage runs'
-- durations already do on the stage rows.
--
-- flowd's harvest tick (cmd/flowd/timesplits.go) fills it with
-- time_split's semantics: SQL NULL is not computed yet; JSON null is
-- computed and no session transcript found, so the tick never looks again;
-- a change of either instant clears it through the trigger below.
ALTER TABLE stage_runs ADD COLUMN IF NOT EXISTS parent_split jsonb;

CREATE OR REPLACE FUNCTION reset_parent_split() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  IF NEW.started_at IS DISTINCT FROM OLD.started_at OR NEW.ended_at IS DISTINCT FROM OLD.ended_at THEN
    NEW.parent_split := NULL;
  END IF;
  RETURN NEW;
END $$;

CREATE TRIGGER stage_runs_reset_parent_split BEFORE UPDATE OF started_at, ended_at ON stage_runs
  FOR EACH ROW EXECUTE FUNCTION reset_parent_split();

CREATE INDEX stage_runs_parent_split_pending ON stage_runs (id)
  WHERE parent_split IS NULL AND session_id IS NOT NULL AND ended_at IS NOT NULL;
