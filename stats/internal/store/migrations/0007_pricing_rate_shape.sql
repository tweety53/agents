-- 0007_pricing_rate_shape.sql: the pricing table learns the rates that
-- actually exist (task 23).
--
-- Anthropic charges two different rates for a cache write, not one: a
-- 5-minute cache write costs 1.25x base input, a 1-hour write costs 2x.
-- The original pricing table (0003_stage_runs.sql) has a single
-- cache_write_per_mtok column, priced against the harvester's collapsed
-- cache_creation_input_tokens total -- which cannot say which of the two
-- rates applied to any given token, and defaulting to either one is a
-- 37.5%-to-60% pricing error in whichever direction is wrong (task 23's
-- own plan-provenance note measures this session's own transcripts as
-- ~100% 1-hour writes).
--
-- cache_write_5m_per_mtok and cache_write_1h_per_mtok are the two real
-- rates. The existing cache_write_per_mtok column is left in place --
-- dropping a column is a separate decision from a query ceasing to read
-- it -- and every row already in this table has its value backfilled into
-- cache_write_5m_per_mtok, so no already-seeded row silently loses its
-- rate the moment store.Store.Price stops reading the old column.
-- cache_write_1h_per_mtok is deliberately left NULL for those pre-existing
-- rows rather than defaulting to 0 or to the 5m rate: neither is a rate
-- anyone actually published for that row, and this table's absence-is-not-
-- a-value rule (the same rule the metrics bag follows) applies to a
-- pricing row's own columns exactly as it applies to a stage run's
-- metrics -- a 1-hour write priced against that row is unpriceable until a
-- real rate is published for it, not free.
--
-- fast_input_per_mtok and fast_output_per_mtok are the fast-mode rates
-- (Opus 5 and Opus 4.8 charge double for fast speed; Sonnet 5 and Haiku
-- 4.5 have no fast-mode rate at all, per the seed table in
-- pricing_seed.go). Nullable for the same reason: a model with no
-- published fast rate must price a fast-speed run as unavailable, never
-- at an invented rate.
ALTER TABLE pricing
  ADD COLUMN cache_write_5m_per_mtok NUMERIC,
  ADD COLUMN cache_write_1h_per_mtok NUMERIC,
  ADD COLUMN fast_input_per_mtok NUMERIC,
  ADD COLUMN fast_output_per_mtok NUMERIC;

UPDATE pricing SET cache_write_5m_per_mtok = cache_write_per_mtok
  WHERE cache_write_5m_per_mtok IS NULL;

ALTER TABLE pricing ALTER COLUMN cache_write_5m_per_mtok SET NOT NULL;
