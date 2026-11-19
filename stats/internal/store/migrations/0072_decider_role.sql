-- 0072_decider_role.sql: the decider role, the fresh subagent planning's
-- Decide step dispatches to write decision.json -- the implementer groups,
-- the review panel's bundles and visual verification; its one base pair, and
-- so its default, is opus/flow-low.
INSERT INTO role_pairs (role, tier, pair, marker, position) VALUES
    ('decider', 'base', 'opus/flow-low', '', 0)
ON CONFLICT DO NOTHING;
