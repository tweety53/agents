-- 0057_self_review_finding_severity.sql: a self-review finding's severity,
-- critical, important or minor -- optional, so a row recorded before it, or
-- by a caller that graded none, keeps NULL (KAN-1068).
ALTER TABLE self_review_findings ADD COLUMN IF NOT EXISTS severity TEXT
    CHECK (severity IN ('critical', 'important', 'minor'));
