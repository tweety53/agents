-- 0049_review_dynamics.sql: one severity scale and a record of prompt changes.
--
-- findings.severity is free text: reviewers have written Critical,
-- blocking, major, High, closed and two dozen other spellings. The Review
-- dynamics view groups by severity, so canonical_severity folds every
-- spelling onto critical / important / minor, case-insensitively, in SQL
-- beside the grouping that reads it (canonical-severity). Stored values
-- stay as the reviewer wrote them.
--
-- prompt_changes holds one row per change to a tunable prompt: a tuning
-- /flow-prompt-tuning applied (kind tuning) or a commit that edited the
-- prompt file (kind edit). sha is "<commit>:<target>", so one commit that
-- edits two prompts yields two rows (prompt-changes-table,
-- git-import-backfill). area names which chart family draws the row as a
-- marker (markers-by-area).
CREATE FUNCTION canonical_severity(raw text) RETURNS text
LANGUAGE sql IMMUTABLE STRICT AS $$
  SELECT CASE lower(btrim(raw))
    WHEN 'critical' THEN 'critical'
    WHEN 'blocking' THEN 'critical'
    WHEN 'severe' THEN 'critical'
    WHEN 'important' THEN 'important'
    WHEN 'major' THEN 'important'
    WHEN 'high' THEN 'important'
    WHEN 'medium' THEN 'important'
    WHEN 'moderate' THEN 'important'
    ELSE 'minor'
  END
$$;

CREATE TABLE prompt_changes (
    id      bigserial PRIMARY KEY,
    sha     text NOT NULL UNIQUE,
    at      timestamptz NOT NULL,
    kind    text NOT NULL CHECK (kind IN ('tuning', 'edit')),
    area    text NOT NULL CHECK (area IN ('review', 'impl', 'vv', 'fix')),
    slot    text,
    target  text NOT NULL,
    summary text NOT NULL,
    case_id text
);

CREATE INDEX prompt_changes_at_idx ON prompt_changes (at);
