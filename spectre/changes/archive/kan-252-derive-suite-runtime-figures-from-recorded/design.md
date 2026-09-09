# design — suite runtimes recorded in the store, cited by project.md

## Context

KAN-252: a measured suite runtime pasted into `.flow/project.md`'s `## test`
prose goes stale invisibly — the 118.63s figure the issue quotes survived two
reworkings; kan-362's "52 to 56s" replacement is pasted prose with the same
defect. The stats app already records durations per stage run, so it can hold
measured suite runtimes: recorded per run, per suite, on the machine that ran
them. `spectre/specs/` is empty — no capability spec edit. The change follows
one existing end-to-end precedent at every layer: `hazards` (store
`hazards.go` → routes in `api/records.go`/`server.go` → CLI `cmd/flow/hazard.go`).

## Shape

- **Migration `0019_suite_runs.sql`** — project-scoped table, not change-scoped:
  suites run outside any change too (the reason `incidents` in 0018 carries
  `project_key` directly). Columns: `id`, `project_key` (FK projects),
  `suite TEXT`, `host TEXT`, `duration_ms BIGINT CHECK (>= 0)`, `exit_code INT`,
  `ran_at TIMESTAMPTZ NOT NULL DEFAULT now()`; index on
  `(project_key, suite, ran_at)`. `exit_code` is recorded so a failed run's
  duration never presents itself as a runtime figure. New migration file — never
  an edit to an applied one (0018's header states why).
- **Store** — `InsertSuiteRun`, `ListSuiteRuns(project, suite optional, limit)`
  in `internal/store/suiteruns.go` (+ test against the per-test-DSN harness the
  sibling store tests use; skips when the compose stack is unreachable).
- **API** — `POST /api/v1/suites/{project}/runs`,
  `GET /api/v1/suites/{project}/runs?suite=&limit=` registered in
  `api/server.go`, handlers in `api/suites.go` (+ test), shape after the
  verdicts/incidents routes.
- **Client** — `RecordSuiteRun`, `ListSuiteRuns` in `internal/client/suites.go`
  (+ test) with the standard fallback contract: an unreachable store never
  alters the wrapped run's outcome.
- **CLI `cmd/flow/suite.go`** (dispatch `case "suite"` in `main.go`):
  - `flow suite record -suite <name> -- <command…>` — child stdio passes
    through untouched; wall-timed; records duration + hostname + exit code;
    exits with the child's own exit code. Store unreachable → one `⚠` stderr
    line, exit code unchanged. Child cannot start → 127, nothing recorded.
  - `flow suite list [-suite <name>] [-limit N] [-json]` — recent rows (suite,
    ran_at, duration, host, exit) plus one summary line per (suite, host):
    **median duration of the last 10 passing runs** — the figure project.md
    cites. Median: one cold-cache run must not move the number.
- **`.flow/project.md` `## test` rewrite** — the "Measured runtime … 52 to 56s"
  paragraph and its two `<!-- measured: … -->` comments are replaced by the
  citation: runtimes live in the store, query with `flow suite list`; the
  harness's default tool-timeout advice stays, anchored to recorded figures. Suite
  names are free-form; the canonical three named there: `guard-tests`
  (`scripts/run-guard-tests.sh`), `stats-go` (`go test ./... -race -count=1`),
  `stats-spa` (`stats/web` `npm test`). The TMPDIR incident note stays
  (operational history). The 0.84 s `check-installed-citations` note stays as
  prose — a single guard's cost, not a suite runtime; widening scope to it is
  not this change.
- **Seeding** — after implementation, each of the three suites is recorded once
  through the wrapper on this machine so `flow suite list` answers from real
  data (each fits inside the harness's default tool timeout individually). Records land in the
  dev store (project `agents-a740d89c`), which is where machine-local figures
  belong.

The issue's interim step (hand-correct the current number) is dropped as moot:
the mechanism and the prose replacement land in the same change, so there is no
window where the stale number has to be held honest by hand.

## Decisions

### Cite the app; no drift guard

**ID:** cite-app-only
**Status:** active
**Chosen:** `## test` prose names `flow suite list` and carries no number — nothing stale can exist.
**Considered:** a guard reading the store and failing on drift past a documented threshold — rejected: it makes every lint run depend on the daemon being up, and cite-only already removes the stale figure; a guard can be added later if recorded figures go ignored.

### Record via a CLI wrapper command

**ID:** cli-wrapper-recorder
**Status:** active
**Chosen:** `flow suite record -suite <name> -- <command…>` times and records any suite; no harness edits.
**Considered:** harnesses self-report at exit — rejected: touches three surfaces (guard runner, Go, npm) to do what one wrapper does; `flow verify` records as a side effect — rejected: figures would exist only for runs made during changes, and flow would have to identify which `## test` command is which suite.

### Suite runs are project-scoped

**ID:** project-scoped-suite-runs
**Status:** active
**Chosen:** `suite_runs` carries `project_key` directly, no `change_id`.
**Considered:** change-scoped like `stage_runs`/`guard_verdicts` — rejected: suites run outside changes too; a nullable change_id would leave project derivation broken exactly as 0018's header documents for incidents.

### Host is a recorded column, and the summary is per host

**ID:** per-host-figures
**Status:** active
**Chosen:** `host` recorded per run; `flow suite list`'s summary line is per (suite, host).
**Considered:** one global figure — rejected: machine-dependence is a feature here (KAN-252); a figure measured on the developer's laptop is not a fact about CI, and mixing hosts would average across machines.

### Exit code recorded; summary counts passing runs only

**ID:** passing-runs-only-summary
**Status:** active
**Chosen:** `exit_code INT` column; the median-of-last-10 summary ignores failed runs.
**Considered:** recording durations only for passing runs — rejected: a failed run's duration is diagnostic (a timeout is visible as such), it just must not present itself as a runtime figure.

## Open questions

<!-- none — the convergence confirm closed with "approve and move on" and no
     recorded deferrals. A stage that left nothing open records none; empty,
     not absent. -->
