-- 0026_finding_patterns.sql: a persisted registry of finding patterns, so
-- a panel can query a defect pattern's recurrence instead of depending on
-- a reviewer happening to recall it (KAN-416). One row per labeled finding
-- occurrence; a pattern's recurrence is a count of its rows, which is why
-- the pattern name lives as an indexed column rather than buried in a
-- payload.
--
-- (change_id, finding_ref) is UNIQUE and NAMED: a finding labels one
-- pattern, and the write path replays -- a fix round restating the finding
-- re-sends its pattern -- so the named constraint is the upsert's race
-- detector exactly as findings_ref_key is for the findings table, and the
-- row it guards doubles as the replay's convergence point.

CREATE TABLE finding_patterns (
  id          BIGSERIAL PRIMARY KEY,
  pattern     TEXT NOT NULL,
  change_id   BIGINT NOT NULL REFERENCES changes(id),
  finding_ref TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT finding_patterns_occurrence_key UNIQUE (change_id, finding_ref)
);

CREATE INDEX finding_patterns_pattern ON finding_patterns (pattern);
