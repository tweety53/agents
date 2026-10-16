-- 0038_self_review_findings_once.sql: one row per self-review finding.
-- A finding is its (project, change, angle, disposition) plus its ref when
-- it carries one -- the landed sha of a fixed row, the issue key of a filed
-- one -- or its note when it does not. A rewritten sha is a new ref, so it
-- still adds a second row, as skills/flow-self-review/SKILL.md step 6 states.
-- Supersedes 0035's "no uniqueness constraint": a repeated write -- the
-- landing's own fixed: rows recorded again by hand, a write retried after a
-- client timeout that the store had already committed -- is never a new
-- finding. md5(note) keeps a long note inside the btree's row-size limit.
--
-- Existing repeats collapse onto their first row, which takes the first
-- non-null blast radius among them.

UPDATE self_review_findings a
SET blast_radius = (
  SELECT b.blast_radius FROM self_review_findings b
  WHERE b.project_key = a.project_key AND b.change = a.change AND b.angle = a.angle
    AND b.disposition = a.disposition
    AND COALESCE(b.ref, md5(b.note)) = COALESCE(a.ref, md5(a.note))
    AND b.blast_radius IS NOT NULL
  ORDER BY b.id LIMIT 1
)
WHERE a.blast_radius IS NULL;

DELETE FROM self_review_findings a
USING self_review_findings b
WHERE b.project_key = a.project_key AND b.change = a.change AND b.angle = a.angle
  AND b.disposition = a.disposition
  AND COALESCE(b.ref, md5(b.note)) = COALESCE(a.ref, md5(a.note))
  AND b.id < a.id;

CREATE UNIQUE INDEX self_review_findings_once
  ON self_review_findings (project_key, change, angle, disposition, (COALESCE(ref, md5(note))));
