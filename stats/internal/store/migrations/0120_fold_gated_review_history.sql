-- 0120_fold_gated_review_history.sql: the gated per-task review retired
-- (0119) and its history folds into the panel's roles, so no reader keeps a
-- case for it: a gated-reviewer dispatch was a first-pass review, now
-- panel-reviewer, and a gated-re-reviewer one a re-review after a fix, now
-- panel-rerun. A gated review recorded no slot and filed its findings under
-- slot gated-reviewer (or gated-re-reviewer), which readers derived from its
-- task-<ids>-reviewer key; the slot-less rows take slot primary, the panel
-- slot a whole-branch first pass files under, and so do their findings.
-- Every other column -- the dispatch key, seq, round, cost, outcome -- is
-- kept. Idempotent: each update matches only the old names.
UPDATE dispatches SET slot = 'primary'
 WHERE role IN ('gated-reviewer', 'gated-re-reviewer') AND slot IS NULL;

UPDATE dispatches SET role = CASE role WHEN 'gated-reviewer' THEN 'panel-reviewer' ELSE 'panel-rerun' END
 WHERE role IN ('gated-reviewer', 'gated-re-reviewer');

UPDATE dispatches SET policy_role = CASE policy_role WHEN 'gated-reviewer' THEN 'panel-reviewer' ELSE 'panel-rerun' END
 WHERE policy_role IN ('gated-reviewer', 'gated-re-reviewer');

UPDATE findings SET slot = 'primary'
 WHERE slot IN ('gated-reviewer', 'gated-re-reviewer');
