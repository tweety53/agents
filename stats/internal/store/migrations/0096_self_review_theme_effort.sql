-- 0096_self_review_theme_effort.sql: a finding's theme and effort as indexed columns of their own.
-- Both are derived from text the pass already writes -- the note's `[theme: <tag>] ` prefix and the
-- detail's `Effort: <S|M|L>` line -- so a generated column keeps them in step with no write path
-- to change, and existing rows fill themselves.
ALTER TABLE self_review_findings
    ADD COLUMN IF NOT EXISTS theme TEXT GENERATED ALWAYS AS (substring(note from '^\[theme: ([a-z0-9-]+)\]')) STORED,
    ADD COLUMN IF NOT EXISTS effort TEXT GENERATED ALWAYS AS (substring(detail from '(?m)^Effort: ([A-Za-z]+)')) STORED;
CREATE INDEX IF NOT EXISTS self_review_findings_theme ON self_review_findings (project_key, theme);
