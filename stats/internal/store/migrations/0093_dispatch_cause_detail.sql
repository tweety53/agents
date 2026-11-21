-- 0093_dispatch_cause_detail.sql: where a blocked dispatch's evidence lives.
--
-- A blocked verifier row carried only the closed `cause` word, so the red
-- verify's log path survived only in the run's prose. cause_detail is free
-- text beside it (the log path); it never joins GROUP BY cause, so the
-- closed cause vocabulary stays queryable. Nullable, and it follows the
-- outcome's last-write-wins semantics like cause.
ALTER TABLE dispatches ADD COLUMN IF NOT EXISTS cause_detail TEXT;
