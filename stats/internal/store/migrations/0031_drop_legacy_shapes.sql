-- 0031_drop_legacy_shapes.sql: rewrite the row shapes retired writers left
-- behind into the shape written today, so no reader has to handle the old
-- ones (design: stats-legacy-migrate-then-remove).
--
-- - changes.updated_by: the synthetic-row sentinel loses its "myflow" prefix,
--   matching stages.SyntheticChangeUpdatedBy.
-- - decisions: a "panel.dispatches" element that is a bare slot-id array
--   becomes {"slots": <array>}, a top-level "groups" element that is a bare
--   bundle-id array becomes {"bundles": <array>}. Object elements and the
--   array order are kept as they are; a decision with no bare element is not
--   touched at all.
-- - pricing: the collapsed cache-write column, which nothing reads any more,
--   is dropped. No 1h rate is backfilled (design:
--   stats-legacy-migrate-no-pricing-backfill): 0007 set every pre-0007 row's
--   5m rate to its collapsed one, so "collapsed equals 5m" cannot tell a flat
--   row from one never given a 1h rate. The pricing seed alone writes
--   glm-5.3-flash's 1h rate; any other null-1h row keeps refusing a 1h or
--   unknown-split cache write rather than pricing it at the 5m rate.

UPDATE changes
SET updated_by = 'flow stage begin (synthetic)'
WHERE updated_by = 'myflow stage begin (synthetic)';

UPDATE decisions
SET decision = jsonb_set(decision, '{panel,dispatches}', (
    SELECT jsonb_agg(CASE WHEN jsonb_typeof(e.val) = 'array'
                          THEN jsonb_build_object('slots', e.val) ELSE e.val END
                     ORDER BY e.ord)
    FROM jsonb_array_elements(decision->'panel'->'dispatches') WITH ORDINALITY AS e(val, ord)))
WHERE jsonb_typeof(decision->'panel'->'dispatches') = 'array'
  AND EXISTS (SELECT 1 FROM jsonb_array_elements(decision->'panel'->'dispatches') AS e(val)
              WHERE jsonb_typeof(e.val) = 'array');

UPDATE decisions
SET decision = jsonb_set(decision, '{groups}', (
    SELECT jsonb_agg(CASE WHEN jsonb_typeof(e.val) = 'array'
                          THEN jsonb_build_object('bundles', e.val) ELSE e.val END
                     ORDER BY e.ord)
    FROM jsonb_array_elements(decision->'groups') WITH ORDINALITY AS e(val, ord)))
WHERE jsonb_typeof(decision->'groups') = 'array'
  AND EXISTS (SELECT 1 FROM jsonb_array_elements(decision->'groups') AS e(val)
              WHERE jsonb_typeof(e.val) = 'array');

ALTER TABLE pricing DROP COLUMN cache_write_per_mtok;
