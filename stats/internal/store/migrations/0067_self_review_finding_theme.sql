-- 0067_self_review_finding_theme.sql: a self-review finding's theme, one of
-- records.SelfReviewThemes -- what a cost or speed observation is about, so
-- the same observation noted change after change can be counted. Optional: a
-- row recorded before it, or of another angle, keeps NULL.
ALTER TABLE self_review_findings ADD COLUMN IF NOT EXISTS theme TEXT;
CREATE INDEX IF NOT EXISTS self_review_findings_theme ON self_review_findings (theme, recorded_at) WHERE theme IS NOT NULL;
