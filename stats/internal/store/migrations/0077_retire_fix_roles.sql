-- slim-flow-loop retired the fix planner, the fix-run judge and the late-fix reviewer:
-- a fix now runs through skills/flow/fix.md with one implementer and one primary reviewer,
-- so no dispatch runs on these roles and their allowed pairs go.
DELETE FROM role_pairs WHERE role IN ('fix-planner', 'fix-run-judge', 'late-fix');
