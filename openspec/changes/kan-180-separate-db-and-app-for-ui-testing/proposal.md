## Why

UI testing of the myflow stats SPA has no stack of its own. The `myflow` CLI defaults to
`-addr http://127.0.0.1:4173`, and the SPA is served by the daemon on that same port, so every smoke
run, status check against a throwaway repository, and agent-driven UI exercise writes permanent rows
into the live database.

**Eight of the ten rows in `projects` were test artifacts** when the database was read on 2026-08-15,
before it was wiped:

<!-- measured: `select * from projects` and `select * from changes` against myflow-postgres on 2026-08-15, before the wipe -->

| Junk row | Recorded checkout path |
|---|---|
| `smoke`, `proj` | `/tmp/x` |
| `proj1` | `/tmp/smoke` |
| `reviewproj` | `/a/repo/one` |
| `review-repo-d9a8a557` | `/private/tmp/review-repo` |
| `verify-f1-1786665626` | `/tmp/verify-checkout` |
| `myflow-status-test-repo-1388118d` | `/private/tmp/myflow-status-test-repo` |
| `myflow-status-live-386682ad` | `/private/tmp/myflow-status-live` |

Two synthetic `changes` rows, `k174-live-test` and `k174-live-test2`, sat at `STARTED` beside them.
All of it shared a schema with the genuine `agents-a740d89c` and `gymie-7c1f238a` projects and the
real KAN-167, KAN-172 and KAN-174 changes, with nothing distinguishing one from the other. **The only
remedy available was a full wipe, which destroyed the real history along with the junk.**

**The test suite is not the cause and is not being changed.** Every store-backed test helper already
creates a per-test database it drops afterwards — `internal/store`, `internal/harvest`,
`internal/reconcile`, `internal/sweep`, `internal/api` and `cmd/myflow` all do this. `go test ./...`
never touches the shared database. The pollution came from manual and agent-driven UI testing, which
has no alternative target to name.

**Development shares the same stack too.** A `/myflow-do` apply worktree runs against the main
checkout's database and port, because this repository declares no `## workspace isolation` section.
Two changes in flight share one schema and collide on port 4173. That is a second, independent
failure with the same root shape, and both are in scope here.

## What Changes

**Per-worktree isolation, for development.**

- **`.myflow/project.md` gains a `## workspace isolation` section**, declaring `MYFLOWD_DSN` as the
  `database` row, `MYFLOWD_PORT` as a `port` row, and `MYFLOW_ADDR` as a `url` row built from that
  port. Every variable named is one the code already reads, so the section declares behaviour rather
  than requesting new plumbing, and each `Default` is today's literal value — the column where the
  backwards-compatibility promise is actually kept.
- **`scripts/workspace.sh`**, supplying the contract's `create`, `remove` and `survivors` commands.
  `survivors` filters with `awk` rather than `grep`, because `grep` exits 1 on no match and a
  non-zero exit is read as *the check could not run* — a project filtering with `grep` can never
  reach the one result that verifies its own cleanup. It also carries its own timeout inside the
  container, since the guard's sixty-second bound cannot terminate work behind `docker exec`.
- **`MYFLOW_STATE_DIR` and `MYFLOW_TRANSCRIPTS_DIR` are deliberately not isolated** — the contract's
  `Resource` vocabulary is closed and a directory path is none of the five words. Recorded in prose
  beside the tables, so the absence reads as a decision.
- **Two contract files stop citing this repository as the project with nothing to isolate.**
  `project-configuration.md` and `workspace-isolation.md` both use it as that worked example; `stats/`
  made the claim false.

**The UI-test stack, for ad-hoc testing from the main checkout.**

- **A second, complete myflow stack for UI testing**, sharing the `myflow-postgres` container and
  nothing else: database `myflow_uitest`, daemon port 4174, transcripts root and fallback state
  directory both pointed at temporary directories. Pointing the transcript root away from the real
  one is load-bearing rather than tidy — the daemon harvests every five seconds, so a test database
  left on the real root refills itself with genuine session data within seconds of starting.
- **`MYFLOW_ADDR`, an environment override for the CLI's default daemon address.** `cmd/myflow`
  passes a constant as the `-addr` default at three sites; one resolver reads the variable and the
  three take their default from it. An explicit `-addr` still wins. The existing flag already worked
  — it is precisely what nobody remembered to pass, so the fix is to remove the need to remember it.
- **A committed Go fixture seeder** so the test stack renders a populated UI rather than an empty
  one. Written against the store's own types, so a migration that changes a column breaks the build
  instead of drifting silently the way a SQL fixture would.
- **`make ui-test-up` / `make ui-test-down`**, which reset the test database on every bring-up. A
  persistent test database accumulates exactly the way the live one did.
- **A `_uitest` name guard** on the seeder and the teardown, checked before any statement runs.
  Without it, "do not point this at the live database" is advice — and advice is what already
  failed here.

**`MYFLOW_ADDR` serves both halves**, which is why they are one change rather than two: the same
variable points the CLI at the UI-test stack and at a worktree's own daemon.

## Capabilities

### Added Capabilities

- `myflow-ui-test-stack`: a UI-testing stack separate from the live one, how it is targeted, and the
  guard that keeps its destructive paths off the live database.

### Modified Capabilities

- `agents-repo-verification`: this repository declares workspace isolation, so the guard that
  validates the declaration now has one to check rather than passing silently.

## Impact

**Code** — `stats/cmd/myflow/state.go` and `stage.go` (the `-addr` default), a new
`stats/cmd/uitest-seed/`, `stats/Makefile`, a new `scripts/workspace.sh` and its test.

**Configuration** — `.myflow/project.md` gains `## workspace isolation` and a `## run` note.

**Docs** — `stats/README.md`; the stale worked example in
`skills/myflow-contracts/project-configuration.md` and
`skills/myflow-contracts/workspace-isolation.md`.

**Not changing** — the Go test suite's existing per-test-database isolation, the daemon's own config
resolution (`MYFLOWD_PORT`, `MYFLOWD_DSN`, `MYFLOW_TRANSCRIPTS_DIR` and `MYFLOW_STATE_DIR` are all
read today and are used as they stand), the compose stack, or the live daemon's behaviour on port
4173 from the main checkout.
