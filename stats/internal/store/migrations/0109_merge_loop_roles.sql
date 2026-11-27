-- 0109_merge_loop_roles.sql: the self-review fix loop and the pipeline-fix
-- loop are one loop (skills/flow-contracts/pipeline.md **Pipeline defects
-- found mid-run** runs the self-review's), so their plan, fix and review
-- roles merge into one family: loop-plan, loop-fix, loop-review-1 and
-- loop-review-r, each starting with its self-review-* row's pairs, copied
-- rather than named. self-review-pass stays its own role: the pass reviews
-- the whole run and has no pipeline-fix counterpart. Historical dispatch rows keep their role; only
-- the role rows go.
INSERT INTO role_pairs (role, tier, pair, position)
SELECT 'loop-' || substr(rp.role, length('self-review-') + 1), rp.tier, rp.pair, rp.position
FROM role_pairs rp
WHERE rp.role IN ('self-review-plan', 'self-review-fix', 'self-review-review-1', 'self-review-review-r')
ON CONFLICT DO NOTHING;

DELETE FROM role_pairs
WHERE role IN ('self-review-plan', 'self-review-fix', 'self-review-review-1', 'self-review-review-r',
               'pipeline-fix-plan', 'pipeline-fix-fix', 'pipeline-fix-review-1', 'pipeline-fix-review-r');

-- records.Class folds the loop-* roles; dispatch_class, its SQL twin (0094),
-- folds them too, and keeps the retired names for their history rows.
CREATE OR REPLACE FUNCTION dispatch_class(role text) RETURNS text
LANGUAGE sql IMMUTABLE PARALLEL SAFE AS $$
  SELECT CASE role
    WHEN 'panel-reviewer' THEN 'reviewer' WHEN 'panel-rerun' THEN 'reviewer'
    WHEN 'gated-reviewer' THEN 'reviewer' WHEN 'gated-re-reviewer' THEN 'reviewer'
    WHEN 'self-review-pass' THEN 'reviewer' WHEN 'loop-review-1' THEN 'reviewer'
    WHEN 'loop-review-r' THEN 'reviewer'
    WHEN 'self-review-review-1' THEN 'reviewer'
    WHEN 'self-review-review-r' THEN 'reviewer' WHEN 'pipeline-fix-review-1' THEN 'reviewer'
    WHEN 'pipeline-fix-review-r' THEN 'reviewer'
    WHEN 'panel-merger' THEN 'merger'
    WHEN 'minor-fix' THEN 'panel-fix'
    WHEN 'operator-asked' THEN 'implementer' WHEN 'loop-fix' THEN 'implementer'
    WHEN 'self-review-fix' THEN 'implementer'
    WHEN 'pipeline-fix-fix' THEN 'implementer' WHEN 'no-progress-fixer' THEN 'implementer'
    WHEN 'planner-short' THEN 'planner' WHEN 'plan-designer' THEN 'planner' WHEN 'designer' THEN 'planner'
    WHEN 'loop-plan' THEN 'planner' WHEN 'self-review-plan' THEN 'planner'
    WHEN 'pipeline-fix-plan' THEN 'planner'
    WHEN 'runner' THEN 'verifier' WHEN 'visual-verifier' THEN 'verifier'
    ELSE role
  END
$$;
