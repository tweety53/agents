-- 0018_guard_log.sql: every verdict a guard reached, and a per-project
-- incident row recording a guard's symptom, the recovery taken and the
-- minutes lost -- the two records that, before this file, lived only in
-- memory and chat (KAN-451).
--
-- guard_verdicts derives its project through changes, the same way
-- dispatches and findings do: a verdict always belongs to a change in
-- flight, so change_id -> changes.project_key is the one path to "which
-- project" and a second, independently-writable project_key column would
-- just be a place for the two to disagree. incidents stores project_key
-- directly instead, because its change_id is nullable -- an incident can be
-- recorded against a project with no change currently in flight, or after
-- the change that produced it has archived, and a NULL change_id would
-- leave nothing to derive a project from.
--
-- verdict is the guard's whole verdict line, stored verbatim
-- ("OUTSTANDING: <worktree> -- <breakdown>" or "CLEAR: ..."), rather than a
-- second schema decomposing it into a breakdown of its own: the guard
-- already computed the line once, and a parallel structured form would be
-- a second representation of the same fact that could drift from the text
-- an operator actually saw at the gate.
--
-- false_positive / false_positive_reason / flagged_at are a flag on the
-- verdict row rather than a separate table: a flag names the verdict it
-- corrects, so keeping it on that same row is one write and one row to
-- read back, instead of a join every reader would otherwise have to do to
-- learn whether a given verdict was later overridden.
--
-- This is a NEW file rather than an edit to any applied migration, for the
-- reason 0011_dispatch_agent_id.sql's header gives: migrations here are
-- tracked by filename with no checksum, so editing one already applied to
-- a real database would leave that database silently diverged from a
-- freshly migrated one.

CREATE TABLE guard_verdicts (
  id                    BIGSERIAL PRIMARY KEY,
  change_id             BIGINT NOT NULL REFERENCES changes(id),
  guard                 TEXT NOT NULL,
  worktree              TEXT NOT NULL,
  verdict               TEXT NOT NULL,
  recorded_at           TIMESTAMPTZ NOT NULL,
  false_positive        BOOLEAN NOT NULL DEFAULT false,
  false_positive_reason TEXT,
  flagged_at            TIMESTAMPTZ
);

CREATE INDEX guard_verdicts_change_id ON guard_verdicts (change_id);

CREATE TABLE incidents (
  id           BIGSERIAL PRIMARY KEY,
  project_key  TEXT NOT NULL REFERENCES projects(project_key),
  change_id    BIGINT REFERENCES changes(id),
  guard        TEXT NOT NULL,
  symptom      TEXT NOT NULL,
  recovery     TEXT NOT NULL,
  minutes_lost INT NOT NULL CHECK (minutes_lost >= 0),
  occurred_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX incidents_project_key ON incidents (project_key);
