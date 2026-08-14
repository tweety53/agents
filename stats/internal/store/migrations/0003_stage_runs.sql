-- 0003_stage_runs.sql: stage runs and the pricing table that costs them.
--
-- stage_runs.repo_root is nullable and, when set, is validated against
-- change_repos' composite primary key (change_id, repo_root) through a
-- foreign key. Postgres' default FK match semantics (MATCH SIMPLE) skip the
-- check entirely when any referencing column is NULL, so a NULL repo_root
-- -- meaning "the whole unit of work", never "unknown repository" -- never
-- needs a matching change_repos row. This was verified directly against
-- postgres:18-alpine by this task's own migration test
-- (TestStageRunsNullRepoRootSkipsFKCheck /
-- TestStageRunsRepoRootMustMatchChangeRepos in stageruns_test.go): a NULL
-- repo_root inserts with no change_repos rows at all, and a non-NULL
-- repo_root absent from change_repos is rejected.
--
-- (change_id, command, stage, attempt) is unique, named explicitly rather
-- than left to Postgres' default identifier so the constraint name BeginStage
-- checks for is not tied to Postgres' naming convention. BeginStage
-- allocates the next attempt number for its triple inside a single
-- INSERT ... SELECT, with this constraint as the race detector: two
-- concurrent inserts that compute the same next attempt collide on it, and
-- the loser retries rather than silently overwriting the winner's row.

CREATE TABLE stage_runs (
  id          BIGSERIAL PRIMARY KEY,
  change_id   BIGINT NOT NULL REFERENCES changes(id),
  repo_root   TEXT,
  harness     TEXT NOT NULL,
  session_id  TEXT,
  command     TEXT NOT NULL,
  stage       TEXT NOT NULL,
  attempt     INT  NOT NULL,
  started_at  TIMESTAMPTZ NOT NULL,
  ended_at    TIMESTAMPTZ,
  outcome     TEXT,
  metrics     JSONB NOT NULL DEFAULT '{}'::jsonb,
  CONSTRAINT stage_runs_attempt_key UNIQUE (change_id, command, stage, attempt),
  CONSTRAINT stage_runs_repo_root_fk FOREIGN KEY (change_id, repo_root)
    REFERENCES change_repos (change_id, repo_root)
);

CREATE INDEX stage_runs_metrics_gin ON stage_runs USING GIN (metrics);
CREATE INDEX stage_runs_started_at  ON stage_runs (started_at);

CREATE TABLE pricing (
  model                TEXT NOT NULL,
  effective_from       TIMESTAMPTZ NOT NULL,
  input_per_mtok       NUMERIC NOT NULL,
  output_per_mtok      NUMERIC NOT NULL,
  cache_write_per_mtok NUMERIC NOT NULL,
  cache_read_per_mtok  NUMERIC NOT NULL,
  PRIMARY KEY (model, effective_from)
);

-- jsonb_deep_merge(a, b) merges b into a recursively: where both a and b
-- hold a JSON object at the same key, their keys are merged recursively
-- rather than one object replacing the other; any key present in only one
-- side survives untouched; and a non-object value in b (a number, string,
-- array, bool or null) always replaces whatever was at that key in a,
-- because there is nothing under a non-object value to preserve.
--
-- This exists because MergeMetrics' bag is nested -- every token figure
-- lives under one "tokens" object -- and a plain top-level `a || b` JSONB
-- concatenation replaces "tokens" wholesale on any write that touches it,
-- discarding sibling keys a previous, unrelated write already recorded
-- there (for example the harvester recording tokens.input/output, then a
-- later write recording tokens.main/sidechain, would otherwise erase
-- input/output). A writer must never need to know what other keys already
-- exist under a shared parent in order to avoid destroying them -- the
-- reason this has to be a merge and not a caller-side read-modify-write:
-- the latter reintroduces exactly the lost-update race this package has
-- already had to design around twice (see changes.go's project bootstrap
-- and stageruns.go's attempt allocation).
--
-- A JSON null in b is a non-object value like any other, so it replaces
-- the key rather than deleting it -- there is no delete operation here at
-- all, only merge. This is deliberately not RFC 7396 JSON Merge Patch
-- semantics, where null in the patch means "delete this key"; a reader
-- who knows that convention should not assume it applies here.
--
-- b must not be SQL NULL: jsonb_deep_merge(a, NULL) returns NULL (the
-- CASE's ELSE branch), which the caller's UPDATE would then try to store
-- into stage_runs.metrics, a NOT NULL column. MergeMetrics (stageruns.go)
-- guards against this in Go, ahead of ever reaching this function, with a
-- typed ErrNilMetricsPatch rather than letting it surface as this
-- constraint violation.
CREATE FUNCTION jsonb_deep_merge(a JSONB, b JSONB) RETURNS JSONB AS $$
  SELECT CASE
    WHEN jsonb_typeof(a) = 'object' AND jsonb_typeof(b) = 'object' THEN
      COALESCE(
        (SELECT jsonb_object_agg(
           key,
           CASE
             WHEN a -> key IS NULL THEN b -> key
             WHEN b -> key IS NULL THEN a -> key
             ELSE jsonb_deep_merge(a -> key, b -> key)
           END
         )
         FROM (
           SELECT key FROM jsonb_object_keys(a) AS key
           UNION
           SELECT key FROM jsonb_object_keys(b) AS key
         ) all_keys),
        '{}'::jsonb
      )
    ELSE b
  END
$$ LANGUAGE sql IMMUTABLE;
