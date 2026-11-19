-- 0081_drop_experiments.sql: allowed-pairs experiments are retired. Role
-- pairs are edited directly in the Agent roles view; nothing overrides them.
DROP TABLE IF EXISTS experiment_roles;
DROP TABLE IF EXISTS experiments;
