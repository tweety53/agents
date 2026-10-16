# pipeline-fix-kan-954-user-row-loaded-twice-per-request-12

## Why

`go test -count=1 ./...` in `stats/` fails intermittently with `timeout: context deadline
exceeded` in store-backed tests — seen in the gymie `kan-954-user-row-loaded-twice-per-request`
run at `stats/internal/store/stageruns_test.go:635` (`create test database …`) and in
`TestProjectRootsListsRegisteredCheckouts` (`acquire migration lock`). Six per-package test helpers
— `internal/store`, `internal/reconcile`, `internal/sweep`, `internal/harvest`, `internal/api` and
`cmd/flow` — each create a fresh database from `template1` and run every embedded migration for
every test, under 3-second deadlines, against the one shared `flow-postgres` server. Under
`go test ./...` those packages run concurrently, the per-test migration run slows past the deadline,
and whichever test is unlucky fails. Two of two plan-time runs of the whole suite failed this way.

## What changes

- A new test-support package, `stats/internal/storetest`, migrates one template database per
  package process and hands every test a `CREATE DATABASE … TEMPLATE` copy of it; its `Main` drops
  the template and any test database whose cleanup drop failed, once the package's tests have run.
- The six helpers keep their names and call into `storetest`; each package gains a one-line
  `TestMain`. No deadline changes.
- Two tests: one proving migrated databases are template copies (fails when each is migrated on its
  own), one proving the leftover sweep drops a database a cleanup missed.
