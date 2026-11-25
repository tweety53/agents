-- 0102 added the designer role and records.Class folds it to planner; dispatch_class, its SQL twin
-- (0094), did not, so a designer dispatch counted as a class of its own in the queries.
CREATE OR REPLACE FUNCTION dispatch_class(role text) RETURNS text
LANGUAGE sql IMMUTABLE PARALLEL SAFE AS $$
  SELECT CASE role
    WHEN 'panel-reviewer' THEN 'reviewer' WHEN 'panel-rerun' THEN 'reviewer'
    WHEN 'gated-reviewer' THEN 'reviewer' WHEN 'gated-re-reviewer' THEN 'reviewer'
    WHEN 'self-review-pass' THEN 'reviewer' WHEN 'self-review-review-1' THEN 'reviewer'
    WHEN 'self-review-review-r' THEN 'reviewer' WHEN 'pipeline-fix-review-1' THEN 'reviewer'
    WHEN 'pipeline-fix-review-r' THEN 'reviewer'
    WHEN 'panel-merger' THEN 'merger'
    WHEN 'minor-fix' THEN 'panel-fix'
    WHEN 'operator-asked' THEN 'implementer' WHEN 'self-review-fix' THEN 'implementer'
    WHEN 'pipeline-fix-fix' THEN 'implementer' WHEN 'no-progress-fixer' THEN 'implementer'
    WHEN 'planner-short' THEN 'planner' WHEN 'plan-designer' THEN 'planner' WHEN 'designer' THEN 'planner' WHEN 'self-review-plan' THEN 'planner'
    WHEN 'pipeline-fix-plan' THEN 'planner'
    WHEN 'runner' THEN 'verifier' WHEN 'visual-verifier' THEN 'verifier'
    ELSE role
  END
$$;
