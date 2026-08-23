-- 0013_session_token_giveups.sql: a durable record of a session token the
-- watcher searched for and could not resolve, so a restart can retry it
-- instead of losing it the moment the in-memory pending set is gone.
--
-- The watcher (internal/harvest, Watcher.resolveSessionTokens) bounds how
-- long it will keep searching for the stage run or dispatch a session token
-- belongs to: past maxSessionTokenResolutionCycles, or on an ambiguous
-- match, it gives up on that token for the rest of the process's life. Until
-- this table, "gives up" meant only w.gaveUpTokens, an in-memory set --
-- restarting the daemon (the fix for kan-302's own trigger, most commonly)
-- silently discarded every token it held, so a session whose transcript
-- later did carry the missing mark was never searched again and never
-- resolved. This table is what a restart reads at start
-- (store.PersistedGiveUps) to re-seed that search.
--
-- Keyed on session_token, not on a stage_run id or a dispatch id: one
-- session token can be the reason several marks within the same run went
-- unresolved -- a stage-end mark and one or more dispatch-end marks all
-- name the same token, and it is the token, not any single mark, that the
-- watcher was searching for. Recording the give-up once per token, rather
-- than once per mark it would have resolved, is what lets
-- MarkDispatchesUnattributed (RecordSessionTokenGiveUp's dispatch-grain
-- counterpart) stamp every dispatch under that token from the one row.
--
-- THERE IS DELIBERATELY NO INDEX BEYOND THE PRIMARY KEY. The only two
-- queries against this table are RecordSessionTokenGiveUp's upsert, which
-- the primary key already serves, and PersistedGiveUps' unfiltered read of
-- every row at watcher start -- a table sized by how many tokens a watcher
-- has ever given up on, not by request volume. Add one only alongside a
-- query that would use it.
CREATE TABLE session_token_giveups (
  session_token TEXT PRIMARY KEY,
  reason        TEXT        NOT NULL,
  gave_up_at    TIMESTAMPTZ NOT NULL,
  retries       INT         NOT NULL DEFAULT 0
);
