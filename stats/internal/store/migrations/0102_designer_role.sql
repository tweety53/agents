-- 0102_designer_role.sql: the designer role -- draws, in a worktree of the
-- project's design repo, every element a planner's design.md lists under
-- `## Undrawn`. Default opus/flow-medium; the fable pairs are opt-in, as the
-- planner's are.
INSERT INTO role_pairs (role, tier, pair, position) VALUES
    ('designer', 'base', 'opus/flow-medium', 0),
    ('designer', 'base', 'opus/flow-low', 1),
    ('designer', 'base', 'opus/flow-high', 2),
    ('designer', 'optin', 'fable/flow-xhigh', 0),
    ('designer', 'optin', 'fable/flow-high', 1),
    ('designer', 'optin', 'fable/flow-medium', 2)
ON CONFLICT DO NOTHING;
