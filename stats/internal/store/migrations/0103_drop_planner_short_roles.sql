-- 0103_drop_planner_short_roles.sql: /flow plans every change on the one
-- planner role again (the pair judge picks a weaker pair for a small task), so
-- the two roles 0084 restored lose their allowed pairs, as in 0083.
-- Historical dispatch rows keep their keys; only the role rows go.
DELETE FROM role_pairs WHERE role IN ('planner-short', 'plan-triage');
