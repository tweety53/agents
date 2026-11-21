-- A dispatch row's role is a role of the roles table (role_pairs): runner,
-- panel-reviewer, visual-verifier, panel-merger. Until now it held a coarse
-- name (verifier, reviewer, merger), and the roles table's own role sat in
-- policy_role where the prompt named one.
--
-- dispatch_class folds a roles-table role back to the coarse class the
-- queries that only need "a reviewer" or "an implementer" test; a role
-- absent from it is its own class. records.Class is its Go twin, and a test
-- holds the two equal.
CREATE FUNCTION dispatch_class(role text) RETURNS text
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
    WHEN 'planner-short' THEN 'planner' WHEN 'plan-designer' THEN 'planner' WHEN 'self-review-plan' THEN 'planner'
    WHEN 'pipeline-fix-plan' THEN 'planner'
    WHEN 'runner' THEN 'verifier' WHEN 'visual-verifier' THEN 'verifier'
    ELSE role
  END
$$;

-- The role each existing row is stored under from now on: the roles-table role
-- its prompt named, else the one its coarse role and key derive
-- (records.LaunchRoleID), else unchanged.
UPDATE dispatches SET role = policy_role
 WHERE policy_role IS NOT NULL AND policy_role <> '' AND role <> policy_role;

UPDATE dispatches SET role = CASE
    WHEN role = 'merger' THEN 'panel-merger'
    WHEN role = 'panel-fix' AND dispatch_key ~ '^panel-.*-minor-fix$' THEN 'minor-fix'
    WHEN role = 'reviewer' AND dispatch_key ~ '^task-[0-9]+(\+[0-9]+)*-reviewer.*-fix-' THEN 'gated-re-reviewer'
    WHEN role = 'reviewer' AND dispatch_key ~ '^task-[0-9]+(\+[0-9]+)*-reviewer' THEN 'gated-reviewer'
    WHEN role = 'reviewer' AND (dispatch_key ~ '^fix-[1-9][0-9]*-review$' OR dispatch_key ~ '^panel-[0-9]+-') THEN 'panel-reviewer'
    WHEN role = 'implementer' AND dispatch_key ~ '^task-[0-9]+(\+[0-9]+)*-implementer-(fix|recommit)-' THEN 'panel-fix'
    WHEN role = 'implementer' AND dispatch_key LIKE 'operator-%' THEN 'operator-asked'
    WHEN role = 'planner' AND dispatch_key LIKE 'plan-short-%' THEN 'planner-short'
    WHEN role = 'planner' AND dispatch_key LIKE 'design-%' THEN 'plan-designer'
    WHEN role = 'verifier' AND dispatch_key LIKE 'visual-verify%' THEN 'visual-verifier'
    WHEN role = 'verifier' AND (dispatch_key LIKE 'runner-%' OR dispatch_key = 'verify' OR dispatch_key LIKE 'verify-%') THEN 'runner'
    ELSE role
  END
 WHERE role IN ('merger', 'panel-fix', 'reviewer', 'implementer', 'planner', 'verifier');
