-- 0074_fix_planner_minor_fix_roles.sql: two dedicated roles. fix-planner is
-- the fresh subagent that plans a /flow fix run (key fix-plan-<n>), which the
-- session planned inline until now; it starts with self-review-plan's current
-- base pairs. minor-fix is the fresh subagent that fixes a panel round's open
-- Minors (key panel-<round>-minor-fix), which the session fixed inline until
-- now; it starts with panel-fix's current base pairs. Both are copied rather
-- than named, so no pair is written here; the operator edits them in flowd.
INSERT INTO role_pairs (role, tier, pair, marker, position)
SELECT r.role, rp.tier, rp.pair, rp.marker, rp.position
FROM role_pairs rp
JOIN (VALUES ('fix-planner', 'self-review-plan'), ('minor-fix', 'panel-fix')) AS r(role, source)
  ON rp.role = r.source
WHERE rp.tier = 'base'
ON CONFLICT DO NOTHING;
