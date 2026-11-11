-- 0062_design_reviewer_pair.sql: KAN-1104 adds the design-reviewer role, the
-- plan reviewer dispatched before implementation; its one base pair, and so
-- its default, is opus/flow-low.
INSERT INTO role_pairs (role, tier, pair, marker, position) VALUES
    ('design-reviewer', 'base', 'opus/flow-low', '', 0)
ON CONFLICT DO NOTHING;
