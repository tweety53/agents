-- 0088_runner_role.sql: the runner role (kan-1118) -- a small-context agent
-- that runs the test list and any other long command and returns a verdict,
-- so the parent's cache never lapses under a long run. Operator-chosen
-- default: sonnet/flow-low, the role's only base pair.
INSERT INTO role_pairs (role, tier, pair, position)
VALUES ('runner', 'base', 'sonnet/flow-low', 0)
ON CONFLICT DO NOTHING;
