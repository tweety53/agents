-- 0004_query_indexes.sql: btree indexes for the orderings query.go's
-- allowlisted fields actually produce.
--
-- stage_runs.metrics already has a GIN index (0003_stage_runs.sql) for the
-- containment queries the aggregation views run; this migration adds
-- nothing further for the metrics bag itself, since a JSONB path
-- expression such as (metrics #>> '{tokens,cache_read}') is not indexable
-- by a plain btree without a matching expression index per key, and no
-- single key is hot enough yet to justify one. This is a deliberately
-- deferred optimisation, not an oversight -- add a per-key expression
-- index once a real query pattern demands it.
--
-- stage_runs.change_id already has an implicit index? No: Postgres does
-- NOT automatically index the referencing side of a foreign key, only the
-- referenced side's primary key. QueryStageRuns joins stage_runs to
-- changes on every call (both to filter and search on the owning change's
-- fields), so this index is load-bearing for that join, not merely for
-- the ordering list below.

CREATE INDEX stage_runs_change_id ON stage_runs (change_id);
CREATE INDEX stage_runs_command    ON stage_runs (command);
CREATE INDEX stage_runs_stage      ON stage_runs (stage);
CREATE INDEX stage_runs_outcome    ON stage_runs (outcome);
CREATE INDEX stage_runs_ended_at   ON stage_runs (ended_at);
CREATE INDEX stage_runs_harness    ON stage_runs (harness);

CREATE INDEX changes_state           ON changes (state);
CREATE INDEX changes_updated_at      ON changes (updated_at);
CREATE INDEX changes_branch          ON changes (branch);
CREATE INDEX changes_jira_issue      ON changes (jira_issue);
CREATE INDEX changes_pr_url          ON changes (pr_url);
