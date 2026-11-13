-- 0063_fix_run_judge_pair.sql: the fix-run-judge role, the one-shot judge a
-- fix run dispatches to decide its review-panel roster and verify mode above
-- flow-guard fix-run-floor's floor; its one base pair, and so its default, is
-- opus/flow-low.
INSERT INTO role_pairs (role, tier, pair, marker, position) VALUES
    ('fix-run-judge', 'base', 'opus/flow-low', '', 0)
ON CONFLICT DO NOTHING;
