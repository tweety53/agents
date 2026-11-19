-- 0079_retire_tooling_analyst.sql — slim-flow-loop moved the visual tooling
-- analysis out of the loop into the operator-run flow-visual-tooling-analysis
-- skill, so the `tooling-analyst` role has no dispatch and no allowed pairs.
DELETE FROM role_pairs WHERE role = 'tooling-analyst';
