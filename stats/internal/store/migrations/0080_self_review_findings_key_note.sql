-- 0080_self_review_findings_key_note.sql: a self-review finding is its
-- (project, change, angle, disposition), its ref and its note. 0038 keyed on
-- the ref alone when one was present, so findings filed together under one
-- issue -- the same ref, different notes -- collapsed onto one row. An exact
-- repeat still writes no second row. Every existing row is distinct under the
-- new key, since it was distinct under the narrower old one.

DROP INDEX self_review_findings_once;

CREATE UNIQUE INDEX self_review_findings_once
  ON self_review_findings (project_key, change, angle, disposition, (COALESCE(ref, '')), (md5(note)));
