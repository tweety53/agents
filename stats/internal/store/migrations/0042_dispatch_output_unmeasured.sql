-- 0042_dispatch_output_unmeasured.sql: a subagent transcript writes each message's usage at stream
-- start ("stop_reason": null) and never its final count, so every sidechain output figure the
-- harvester stored before kan-920's self-review is a first-chunk sum, never the real output. The
-- true counts are in no file; the harvester now counts such messages under output_unmeasured
-- instead, and this marks every dispatch row and stage run already stored with sidechain output the
-- same way, so the ledger and the SPA render that output as not measured rather than as the
-- undercount. A cost_usd already written stays.
UPDATE dispatches
SET metrics = jsonb_set(metrics, '{tokens,sidechain,output_unmeasured}', '1')
WHERE jsonb_typeof(metrics->'tokens'->'sidechain') = 'object'
  AND COALESCE((metrics->'tokens'->'sidechain'->>'output')::numeric, 0) > 0
  AND NOT (metrics->'tokens'->'sidechain' ? 'output_unmeasured');

UPDATE stage_runs
SET metrics = jsonb_set(metrics, '{tokens,sidechain,output_unmeasured}', '1')
WHERE jsonb_typeof(metrics->'tokens'->'sidechain') = 'object'
  AND COALESCE((metrics->'tokens'->'sidechain'->>'output')::numeric, 0) > 0
  AND NOT (metrics->'tokens'->'sidechain' ? 'output_unmeasured');
