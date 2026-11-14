-- 0065_pair_judge.sql — the pair-judge role and the pair pick a dispatch was begun with (KAN-1119).
--
-- pair-judge is the role flow-guard pair-judge runs its headless `claude -p`
-- call on, to pick a dispatch's {model, effort} from its role's allowed
-- pairs; its one base pair, and so its default, is sonnet/flow-low. It is a
-- model call, never a dispatch, so no dispatch row ever carries the role.
--
-- dispatches.pair_pick is the judge's `<source>: <reason>` a dispatch was
-- begun with (`flow record dispatch begin -pair-pick`). Nullable: every
-- existing row and every dispatch no judge picked carries none.
INSERT INTO role_pairs (role, tier, pair, marker, position) VALUES
    ('pair-judge', 'base', 'sonnet/flow-low', '', 0)
ON CONFLICT DO NOTHING;

ALTER TABLE dispatches ADD COLUMN pair_pick text;
