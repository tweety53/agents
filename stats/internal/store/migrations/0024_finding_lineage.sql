-- 0024_finding_lineage.sql: a finding's lineage -- which earlier finding
-- it supersedes, and which earlier finding's fix caused it -- as two
-- columns on the findings row rather than as prose in its note.
--
-- Until now a chain like F40 -> F50 -> F56 (the same defect re-raised
-- across a fix round) was visible only to a reader who compared notes
-- closely, and "findings caused by fixes" -- the churn a fix round
-- introduces -- was hindsight, not a metric. Two nullable columns make
-- both a fact the store holds instead of one a reader reconstructs.
--
-- Each column carries a composite foreign key onto (change_id, ref),
-- which findings_ref_key already makes unique per change: a link names a
-- finding of the SAME change or names nothing. A link to a ref the
-- change does not hold is refused, never stored dangling -- a chain whose
-- hops could point at nothing would have to be re-verified by every
-- reader, which is the note-reading this change exists to end. The store
-- maps the violation to a typed error naming the offending ref.
--
-- A CHECK on each column refuses a finding linking to itself: such a row
-- is a typo, never a chain of one.
--
-- Both columns are nullable and default NULL: the overwhelmingly common
-- finding raises no lineage, and rows written before this migration read
-- back as unlinked without a backfill. UpsertFinding replaces both on
-- conflict, exactly as it replaces note and severity -- a restatement
-- that carries no lineage clears the columns rather than preserving a
-- link its caller did not restate.

ALTER TABLE findings
  ADD COLUMN supersedes TEXT,
  ADD COLUMN regression_of TEXT,
  ADD CONSTRAINT findings_supersedes_fk
    FOREIGN KEY (change_id, supersedes) REFERENCES findings(change_id, ref),
  ADD CONSTRAINT findings_regression_of_fk
    FOREIGN KEY (change_id, regression_of) REFERENCES findings(change_id, ref),
  ADD CONSTRAINT findings_supersedes_not_self
    CHECK (supersedes IS NULL OR supersedes <> ref),
  ADD CONSTRAINT findings_regression_of_not_self
    CHECK (regression_of IS NULL OR regression_of <> ref);
