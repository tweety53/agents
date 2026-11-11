-- 0061_role_pairs.sql: every role's allowed pairs live here, not in
-- skills/flow-contracts/model-policy.md (kan-1103, decision pairs-in-db).
-- tier base: always admitted. tier optin: admitted only when the dispatch
-- prompt carries a whole line -- marker when set (No-progress escalation,
-- Operator-approved high), else `Operator pair: <role>` plus a
-- policy_grants row for the dispatch's change (decision optin-grant).
-- position keeps a tier's pairs in the order they were written: a role's
-- first base pair is its default (guard planagentpair).
CREATE TABLE IF NOT EXISTS role_pairs (
    role     text    NOT NULL CHECK (role ~ '^[a-z][a-z0-9-]*$'),
    tier     text    NOT NULL CHECK (tier IN ('base', 'optin')),
    pair     text    NOT NULL CHECK (pair ~ '^[a-z][a-z0-9.-]*/[A-Za-z][A-Za-z0-9._-]*(@[a-z]+)?$'),
    marker   text    NOT NULL DEFAULT '' CHECK (tier = 'optin' OR marker = ''),
    position integer NOT NULL CHECK (position >= 0),
    PRIMARY KEY (role, tier, pair)
);

-- An operator's say-so that <change>'s dispatches of <role> may run an
-- unmarked opt-in pair. Removed by `flow policy revoke`; it lapses when the
-- change reaches FINISHED (the guard reads only a change not FINISHED).
CREATE TABLE IF NOT EXISTS policy_grants (
    change     text        NOT NULL CHECK (change <> ''),
    role       text        NOT NULL CHECK (role <> ''),
    granted_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (change, role)
);

-- The seed: the Allowed pairs table as it stood when this migration landed.
INSERT INTO role_pairs (role, tier, pair, marker, position) VALUES
    ('explorer', 'base', 'opus/Explore@low', '', 0),
    ('explorer', 'base', 'opus/Explore@medium', '', 1),
    ('plan-designer', 'base', 'opus/Plan@low', '', 0),
    ('plan-designer', 'base', 'opus/Plan@medium', '', 1),
    ('plan-designer', 'base', 'opus/Plan@high', '', 2),
    ('plan-designer', 'base', 'opus/Plan@xhigh', '', 3),
    ('plan-designer', 'base', 'fable/Plan@xhigh', '', 4),
    ('plan-designer', 'base', 'fable/Plan@high', '', 5),
    ('plan-designer', 'base', 'fable/Plan@medium', '', 6),
    ('planner', 'base', 'opus/flow-planner-medium', '', 0),
    ('planner', 'base', 'opus/flow-planner-low', '', 1),
    ('planner', 'base', 'opus/flow-planner-high', '', 2),
    ('planner', 'base', 'opus/flow-planner-xhigh', '', 3),
    ('planner', 'base', 'fable/flow-planner-xhigh', '', 4),
    ('planner', 'base', 'fable/flow-planner-high', '', 5),
    ('planner', 'base', 'fable/flow-planner-medium', '', 6),
    ('implementer', 'base', 'opus/flow-low', '', 0),
    ('implementer', 'base', 'opus/flow-medium', '', 1),
    ('implementer', 'base', 'sonnet/flow-low', '', 2),
    ('implementer', 'base', 'sonnet/flow-medium', '', 3),
    ('panel-fix', 'base', 'opus/flow-low', '', 0),
    ('panel-fix', 'base', 'opus/flow-medium', '', 1),
    ('panel-fix', 'base', 'sonnet/flow-low', '', 2),
    ('panel-fix', 'base', 'sonnet/flow-medium', '', 3),
    ('operator-asked', 'base', 'opus/flow-medium', '', 0),
    ('operator-asked', 'base', 'opus/flow-high', '', 1),
    ('operator-asked', 'base', 'sonnet/flow-medium', '', 2),
    ('implementer', 'optin', 'opus/flow-high', 'No-progress escalation', 0),
    ('panel-fix', 'optin', 'opus/flow-high', 'No-progress escalation', 0),
    ('gated-reviewer', 'base', 'opus/flow-medium', '', 0),
    ('gated-re-reviewer', 'base', 'opus/flow-low', '', 0),
    ('panel-reviewer', 'base', 'opus/flow-low', '', 0),
    ('panel-reviewer', 'base', 'opus/flow-medium', '', 1),
    ('panel-reviewer', 'base', 'fable/flow-low', '', 2),
    ('panel-reviewer', 'base', 'fable/flow-medium', '', 3),
    ('panel-reviewer', 'optin', 'opus/flow-high', 'Operator-approved high', 0),
    ('panel-sanity-efficiency', 'base', 'opus/flow-low', '', 0),
    ('panel-sanity-efficiency', 'base', 'sonnet/flow-low', '', 1),
    ('panel-mutation-failure', 'base', 'opus/flow-low', '', 0),
    ('panel-mutation-failure', 'base', 'sonnet/flow-low', '', 1),
    ('panel-rerun', 'base', 'opus/flow-low', '', 0),
    ('panel-rerun', 'base', 'sonnet/flow-low', '', 1),
    ('late-fix', 'base', 'opus/flow-low', '', 0),
    ('panel-merger', 'base', 'sonnet/flow-low', '', 0),
    ('visual-verifier', 'base', 'opus/flow-low', '', 0),
    ('tooling-analyst', 'base', 'opus/flow-medium', '', 0),
    ('self-review-pass', 'base', 'opus/flow-medium', '', 0),
    ('self-review-plan', 'base', 'opus/flow-planner-medium', '', 0),
    ('self-review-plan', 'base', 'opus/flow-planner-low', '', 1),
    ('self-review-fix', 'base', 'opus/flow-medium', '', 0),
    ('self-review-fix', 'base', 'sonnet/flow-medium', '', 1),
    ('self-review-review-1', 'base', 'opus/flow-medium', '', 0),
    ('self-review-review-1', 'base', 'opus/flow-low', '', 1),
    ('self-review-review-r', 'base', 'sonnet/flow-low', '', 0),
    ('pipeline-fix-plan', 'base', 'opus/flow-planner-medium', '', 0),
    ('pipeline-fix-plan', 'base', 'opus/flow-planner-low', '', 1),
    ('pipeline-fix-fix', 'base', 'opus/flow-low', '', 0),
    ('pipeline-fix-fix', 'optin', 'opus/flow-high', 'No-progress escalation', 0),
    ('pipeline-fix-review-1', 'base', 'opus/flow-medium', '', 0),
    ('pipeline-fix-review-1', 'base', 'opus/flow-low', '', 1),
    ('pipeline-fix-review-r', 'base', 'sonnet/flow-low', '', 0)
ON CONFLICT DO NOTHING;
