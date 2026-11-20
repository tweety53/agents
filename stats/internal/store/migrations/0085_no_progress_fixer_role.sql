-- 0085_no_progress_fixer_role.sql: a loop that makes no progress escalates
-- onto its own role, no-progress-fixer, instead of a marked pair of the
-- loop's fixer role; marked pairs (a pair admitted on a whole prompt line)
-- are gone, the marker column with them. Seeded on the pair the marked
-- escalation rows named; the operator edits it in flowd.
DELETE FROM role_pairs WHERE marker <> '';
ALTER TABLE role_pairs DROP COLUMN marker;
INSERT INTO role_pairs (role, tier, pair, position)
VALUES ('no-progress-fixer', 'base', 'opus/flow-high', 0)
ON CONFLICT DO NOTHING;
