-- 0075_plan_triage_planner_short_roles.sql: two planning roles. plan-triage
-- is the headless call (flow-guard plan-triage) that decides whether a new
-- /flow change gets full or short planning; it starts with decider's current
-- base pairs. planner-short runs every planner dispatch of a change triaged
-- short (keys plan-short-<n>); it starts with planner's current base pairs.
-- Both are copied rather than named, so no pair is written here; the
-- operator edits them in flowd.
INSERT INTO role_pairs (role, tier, pair, marker, position)
SELECT r.role, rp.tier, rp.pair, rp.marker, rp.position
FROM role_pairs rp
JOIN (VALUES ('plan-triage', 'decider'), ('planner-short', 'planner')) AS r(role, source)
  ON rp.role = r.source
WHERE rp.tier = 'base'
ON CONFLICT DO NOTHING;
