-- 0066_suite_run_tests.sql — each suite run's per-test costs (KAN-1124).
--
-- Decision per-test-costs-in-flowd: `flow suite record` folds a Go line's
-- `go test -json` stream into one cost per top-level test and sends them
-- with the suite run; `flow suite list -tests` reads them back. Decision
-- store-every-test: one row per non-skipped top-level test of a run, with
-- no cost floor and no retention; the rows cascade-delete with their run.
CREATE TABLE suite_run_tests (
  suite_run_id BIGINT NOT NULL REFERENCES suite_runs(id) ON DELETE CASCADE,
  package      TEXT NOT NULL,
  test         TEXT NOT NULL,
  cost_ms      BIGINT NOT NULL CHECK (cost_ms >= 0),
  PRIMARY KEY (suite_run_id, package, test)
);
