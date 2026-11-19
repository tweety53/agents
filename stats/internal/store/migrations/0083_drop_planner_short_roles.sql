-- 0083_drop_planner_short_roles.sql: /flow plans with one planner role, so
-- planner-short (every planner dispatch of a change triaged short) and
-- plan-triage (the retired headless triage call) lose their allowed pairs.
-- Historical dispatch rows keep their keys; only the role rows go.
DELETE FROM role_pairs WHERE role IN ('planner-short', 'plan-triage');
