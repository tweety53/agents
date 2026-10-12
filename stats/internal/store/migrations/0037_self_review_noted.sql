-- 0037_self_review_noted.sql: a fourth self-review disposition, noted -- a
-- finding that names no defect (what went well, an observation with nothing
-- to fix). It carries no ref, like declined, which now means only a real
-- defect deliberately neither fixed nor filed.
--
-- Rows written before this migration are not rewritten: a past declined
-- note cannot be told apart reliably from a past real decline.
ALTER TABLE self_review_findings DROP CONSTRAINT self_review_findings_disposition_check;
ALTER TABLE self_review_findings ADD CONSTRAINT self_review_findings_disposition_check
  CHECK (disposition IN ('fixed','filed','declined','noted'));
