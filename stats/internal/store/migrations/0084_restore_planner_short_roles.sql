-- 0084_restore_planner_short_roles.sql: /flow judges each new change easy or
-- hard again (flow-guard plan-triage), so the two roles 0083 dropped come
-- back, seeded as 0075 seeded them: plan-triage from decider's base pairs,
-- planner-short from planner's. Copied rather than named, so no pair is
-- written here; the operator edits them in flowd.
INSERT INTO role_pairs (role, tier, pair, marker, position)
SELECT r.role, rp.tier, rp.pair, rp.marker, rp.position
FROM role_pairs rp
JOIN (VALUES ('plan-triage', 'decider'), ('planner-short', 'planner')) AS r(role, source)
  ON rp.role = r.source
WHERE rp.tier = 'base'
ON CONFLICT DO NOTHING;
