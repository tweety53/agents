-- 0012_dispatch_key.sql: the dispatcher's own label for a dispatch, and the
-- uniqueness that makes recording one idempotent.
--
-- A dispatch is recorded in two calls -- `begin` when it starts and `end`
-- when it closes -- because the harvester commits its transcript offset
-- every few seconds and never re-reads past it. A row inserted only at
-- close is invisible to every harvest tick that elapsed while the dispatch
-- ran, so that usage is dropped or credited to an unrelated earlier
-- dispatch. Both halves need to name the same row, and the caller has no
-- seq to name it by: seq is allocated by the store, so a `begin` whose
-- response never came back left the caller with nothing.
--
-- dispatch_key is that name. It is written by the dispatcher as a literal,
-- unique within the run its session_token identifies, and the journalled
-- form of a record write carries the literal unchanged -- which is exactly
-- what makes a replay reproduce it. A `begin` replayed after a lost
-- response therefore collides here and updates the row the first attempt
-- inserted, instead of allocating a fresh seq and double-counting one
-- logical dispatch's cost.
--
-- The uniqueness is (change_id, session_token, dispatch_key) and NOT
-- (change_id, dispatch_key): the key is a label chosen inside one run, so
-- two runs against the same change may legitimately choose the same one --
-- `task-3-implementer` says nothing about which run dispatched it. The
-- session token is what separates them, and it is already the column the
-- attribution window resolves a dispatch's transcript through.
--
-- The column is NULLABLE and the constraint relies on SQL's NULLs-distinct
-- rule rather than forbidding an absent key. A row written without one
-- conflicts with nothing, including another row without one, so it inserts
-- exactly as it did before this migration and is simply never deduplicated
-- and never closed by `end`. That is what keeps the column addable to a
-- table that already holds rows; `myflow record dispatch begin` requires
-- the flag, so no path this repository ships produces such a row.
--
-- It is a named CONSTRAINT rather than a bare unique index for the reason
-- dispatches_seq_key and findings_ref_key are: RecordDispatch names it in
-- its ON CONFLICT clause, so the clause fails loudly if the constraint is
-- ever renamed instead of quietly matching some other index that happens
-- to cover the same columns.

ALTER TABLE dispatches ADD COLUMN dispatch_key TEXT;

ALTER TABLE dispatches
  ADD CONSTRAINT dispatches_key_key UNIQUE (change_id, session_token, dispatch_key);
