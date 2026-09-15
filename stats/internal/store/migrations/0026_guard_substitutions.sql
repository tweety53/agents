-- 0026_guard_substitutions.sql: one row per guard a run's conductor could
-- not run and hand-substituted instead -- which guard failed, on which
-- project topology (the pipeline's single-repo/cross-repo shape vocabulary),
-- and what substitution was used (KAN-417). guard_verdicts records the
-- verdicts guards that RAN reached; this table records the evidence for the
-- guards that never got that far, which before this file lived only in the
-- handoff prose.
--
-- The table mirrors guard_verdicts' shape deliberately: change_id NOT NULL
-- (a substitution is always recorded mid-run, against the change in flight,
-- so project derives through changes exactly as a verdict's does), every
-- write inserts a new row (a guard hand-substituted five times in one run is
-- five rows of evidence, never one replayed write), and the payload columns
-- are plain text -- the substitution is the command or manual step the
-- conductor actually ran, carried verbatim rather than decomposed.
--
-- shape is CHECK-constrained to the two values the pipeline's own shape
-- vocabulary names for a change -- the eighth argument
-- gather-dispatch-context.sh validates, and Hazard.Applies composes with --
-- rather than a free-text column: "three-repo" is a kind of cross-repo, and
-- the substitution text is where the specifics belong.
--
-- This is a NEW file rather than an edit to any applied migration, for the
-- reason 0011_dispatch_agent_id.sql's header gives: migrations here are
-- tracked by filename with no checksum, so editing one already applied to
-- a real database would leave that database silently diverged from a
-- freshly migrated one.

CREATE TABLE guard_substitutions (
  id           BIGSERIAL PRIMARY KEY,
  change_id    BIGINT NOT NULL REFERENCES changes(id),
  guard        TEXT NOT NULL,
  shape        TEXT NOT NULL CHECK (shape IN ('single-repo', 'cross-repo')),
  substitution TEXT NOT NULL,
  recorded_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX guard_substitutions_change_id ON guard_substitutions (change_id);
