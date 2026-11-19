-- 0073_prompt_tuning_roles.sql: dedicated roles for /flow-prompt-tuning's
-- replay and grader dispatches, which ran on self-review-pass's pair until
-- now. Each starts with self-review-pass's current base pairs, copied rather
-- than named, so no pair is written here; the operator edits them in flowd.
INSERT INTO role_pairs (role, tier, pair, marker, position)
SELECT r.role, rp.tier, rp.pair, rp.marker, rp.position
FROM role_pairs rp
CROSS JOIN (VALUES ('prompt-tuning-replay'), ('prompt-tuning-grader')) AS r(role)
WHERE rp.role = 'self-review-pass' AND rp.tier = 'base'
ON CONFLICT DO NOTHING;
