-- 0091_general_role.sql: the general role -- any task no other role fits.
-- Operator-chosen default: sonnet/flow-low, the role's only base pair; the
-- operator adds pairs in the Experiments view.
INSERT INTO role_pairs (role, tier, pair, position)
VALUES ('general', 'base', 'sonnet/flow-low', 0)
ON CONFLICT DO NOTHING;
