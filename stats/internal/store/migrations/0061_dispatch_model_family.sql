-- 0061_dispatch_model_family.sql: every Claude model spelling recorded under
-- its family (kan-1104 self-review, F17). `flow record dispatch begin` folded
-- only claude-opus-* and claude-sonnet-* onto their alias, so sonnet-5-5,
-- claude-fable-* and claude-haiku-* rows -- and every long-form row recorded
-- before that fold -- split one family into several values of
-- dispatches.model, which reviewerScopeSQL filters by exact match.
--
-- The full ID moves into metrics.modelVersion first, where that key is
-- absent: modelVersionSQL (stats/internal/store/reviewdynamics.go) reads a
-- dispatch's version from metrics->>'modelVersion', else from a model that is
-- a full ID, so folding the model alone would turn each such row's version
-- into "(version unrecorded)". A prefix-less spelling (sonnet-5-5) was never
-- read as a version and moves nothing. Both statements are no-ops on a second
-- run.
UPDATE dispatches
SET metrics = jsonb_set(COALESCE(metrics, '{}'::jsonb), '{modelVersion}', to_jsonb(model))
WHERE model ~ '^claude-(opus|sonnet|fable|haiku)(-[0-9]+)*$'
  AND NOT (COALESCE(metrics, '{}'::jsonb) ? 'modelVersion');
UPDATE dispatches SET model = model_family(model)
WHERE model ~ '^(claude-)?(opus|sonnet|fable|haiku)(-[0-9]+)*$'
  AND model <> model_family(model);
