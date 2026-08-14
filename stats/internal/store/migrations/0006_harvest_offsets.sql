-- 0006_harvest_offsets.sql: where the harvester's transcript byte offsets
-- live, replacing the local harvest-offsets.json file task 9 originally
-- used.
--
-- That file was local, on-disk, process-local bookkeeping, separate from
-- the metrics its offsets governed -- and separate meant it could go out
-- of sync with them. Advancing the offset and adding a batch's token
-- deltas were two operations a crash or a store outage could interrupt
-- between, and neither ordering of those two operations was safe: offset
-- first risked silently under-counting a batch whose write then failed
-- (routinely, on the exact condition -- the store being briefly
-- unreachable -- this whole change exists to survive); metrics first
-- risked the reverse, an offset that never advanced being retried and
-- adding the same usage twice. Task 9's post-commit review named both
-- failure modes directly (findings F1 and the follow-up review that
-- reopened it): the fix is not a better-chosen ordering of two operations
-- but making them one.
--
-- harvest_offsets exists so CommitHarvestBatch (harvest.go) can advance a
-- transcript's consumed byte offset in the *same transaction* as the
-- additive metrics write that offset's bytes produced. One row per
-- transcript file, keyed by its absolute path -- which is what the
-- harvester already uses as this file's identity (internal/harvest never
-- assigns transcripts an id of their own).
CREATE TABLE harvest_offsets (
  transcript_path TEXT PRIMARY KEY,
  byte_offset     BIGINT NOT NULL,
  updated_at      TIMESTAMPTZ NOT NULL
);
