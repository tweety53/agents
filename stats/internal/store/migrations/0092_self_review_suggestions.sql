-- 0092_self_review_suggestions.sql: a self-review finding carries its full
-- detail (a markdown body: what, why, how, definition of done, evidence), and
-- a fifth disposition, suggested -- an improvement the pass proposes and the
-- operator approves or declines on the Self-review dashboard, never a Jira
-- task. review_status is set exactly on suggested rows: to-review at first,
-- then approved or declined.
ALTER TABLE self_review_findings ADD COLUMN IF NOT EXISTS detail TEXT;
ALTER TABLE self_review_findings ADD COLUMN IF NOT EXISTS review_status TEXT
    CHECK (review_status IN ('to-review', 'approved', 'declined'));
ALTER TABLE self_review_findings DROP CONSTRAINT self_review_findings_disposition_check;
ALTER TABLE self_review_findings ADD CONSTRAINT self_review_findings_disposition_check
  CHECK (disposition IN ('fixed','filed','declined','noted','suggested'));
ALTER TABLE self_review_findings ADD CONSTRAINT self_review_findings_status_matches_disposition
  CHECK ((disposition = 'suggested') = (review_status IS NOT NULL));
