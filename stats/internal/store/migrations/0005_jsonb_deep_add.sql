-- 0005_jsonb_deep_add.sql: an additive counterpart to jsonb_deep_merge
-- (0003_stage_runs.sql), used by CommitHarvestBatch (internal/store/harvest.go).
--
-- Why this exists: jsonb_deep_merge's own doc comment states its rule
-- plainly -- "a non-object value in b ... always replaces whatever was at
-- that key in a" -- which is correct for a stage-end mark's outcome keys
-- (through MergeMetrics) and for pricing's cost_usd, but wrong for the
-- harvester's token counts. The harvester (internal/harvest) reads a
-- transcript incrementally and, in an earlier version of this design,
-- resent the *cumulative* total it had computed so far on every cycle,
-- recomputed from a local offset-and-totals file kept outside Postgres.
-- That made the local file authoritative for a number Postgres was
-- supposed to own: if the file was ever lost or reset while the
-- transcripts behind it had already been pruned or rotated, the next
-- cycle's smaller recomputed total would *replace* -- silently lower --
-- the correct, larger figure already stored, with no error and no
-- warning (F1, task 9's post-commit review).
--
-- jsonb_deep_add fixes this at the root by making numeric leaves add
-- instead of replace, which is what lets the harvester send per-batch
-- *deltas* (never a cumulative total): CommitHarvestBatch adds a delta
-- onto whatever Postgres already holds, atomically alongside advancing
-- the transcript's consumed byte offset in the same row, so there is no
-- longer a local total (or a local offset) to protect at all -- Postgres
-- is the only place either number lives.
--
-- Structurally this is jsonb_deep_merge with one additional case: where
-- both sides hold a JSON *number* at the same key, the result is their
-- sum rather than b replacing a. Every other rule is unchanged from
-- jsonb_deep_merge -- recurse into shared object keys, take whichever
-- side holds a key the other lacks, and fall back to "b replaces a" for
-- any other non-object, non-numeric-pair case (two strings, a string
-- against a number, and so on). That fallback is what keeps this function
-- safe to use for the same patch shape MergeMetrics accepts: a "model" or
-- "effort" string leaf still behaves as a last-write-wins replace, exactly
-- as it did through jsonb_deep_merge, and a key entirely absent from one
-- side is still carried through untouched -- MergeMetrics' outcome keys,
-- pricing's cost_usd, and every other writer's fields survive a
-- CommitHarvestBatch call exactly as they survive a MergeMetrics call
-- (TestHarvestBatchPreservesStageEndOutcomeKeys, internal/store/harvest_test.go,
-- proves this directly for the stage-end-then-harvest ordering).
CREATE FUNCTION jsonb_deep_add(a JSONB, b JSONB) RETURNS JSONB AS $$
  SELECT CASE
    WHEN jsonb_typeof(a) = 'object' AND jsonb_typeof(b) = 'object' THEN
      COALESCE(
        (SELECT jsonb_object_agg(
           key,
           CASE
             WHEN a -> key IS NULL THEN b -> key
             WHEN b -> key IS NULL THEN a -> key
             ELSE jsonb_deep_add(a -> key, b -> key)
           END
         )
         FROM (
           SELECT key FROM jsonb_object_keys(a) AS key
           UNION
           SELECT key FROM jsonb_object_keys(b) AS key
         ) all_keys),
        '{}'::jsonb
      )
    WHEN jsonb_typeof(a) = 'number' AND jsonb_typeof(b) = 'number' THEN
      to_jsonb((a #>> '{}')::numeric + (b #>> '{}')::numeric)
    ELSE b
  END
$$ LANGUAGE sql IMMUTABLE;
