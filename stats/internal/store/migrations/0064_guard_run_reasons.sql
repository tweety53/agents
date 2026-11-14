-- 0064_guard_run_reasons.sql: why a guard run refused or could not answer.
--
-- guard_runs (0032) keeps a run's exit code and the binary's outcome, and
-- one run of a guard such as check-plan-shape can break several rules, so
-- neither can say which rule or reason drives a guard's refusals. reasons
-- holds the distinct reason ids the run recorded through guard.Reason, in
-- first-seen order, so ranking them is a GROUP BY over this table.
--
-- The column is nullable. It is NULL on a clear run (the binary posts no
-- reasons for one), on a post from a binary older than this change (which
-- omits the field), and on every row recorded before this change, whose
-- reasons were never captured. The store refuses an empty-string id.

ALTER TABLE guard_runs ADD COLUMN reasons TEXT[];
