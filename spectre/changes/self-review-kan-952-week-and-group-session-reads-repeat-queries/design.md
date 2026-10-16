## Context

- Three non-`big` findings from the kan-952 self-review pass, planned as one change per item 1 of
  **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`). Each is carried by
  exactly one task; none is dropped. All were checked against this worktree at `6d14366a` and the
  gymie worktree `kan-952-week-and-group-session-reads-repeat-queries` at `e289b126`:
  - A (Task 1): **Write `IN_PROGRESS`** (`skills/flow/verify-and-handoff.md`) runs
    `check-unfinished-work.sh` only on a fix run (added `b3ad06ea`). The decision behind that scope,
    `fix-run-self-check` in the archived `self-review-kan-937-step-2-frontend-logic-cardio-api-models-step`
    `design.md`, rejected every run "— a creating run's panel already closes its own findings".
    `c89c8e7f` then gave the guard signal five, a panel round dispatched with no pass row, which
    `check-panel-findings-closed.sh` (the panel's own close) does not cover. kan-952's store holds
    reviewer row `panel-1-primary+principles` and no round-1 pass row; its archived `narrative.md`
    records the integrate run stopping on exactly that. The finding's three archived records are
    history and are not edited; the supersession is recorded below.
  - B (Task 2): kan-952's store row `panel-fix-0` (agent `inline`) carries
    `commitSha 2f0fe509…`, subject `chore(spectre): plan kan-952-… — before review`. The fix landed
    as `50af0ddf fix(groupsession): review findings …`, which carries `Flow-Key: panel-fix-0`.
    `skills/flow/review-panel.md` says "`-commit` is the task commit the fixup was folded into", but
    **Panel re-runs** (`skills/flow/review-panel-fix-round.md`) makes a standalone commit at the tip
    carrying the round's `Flow-Key:` trailer the normal route for a branch the remote holds — so the
    finding's "standalone commit instead of fixups" half matches that route and is not a violation,
    and the guard must accept the pushed route's commit or it would refuse every pushed fix round.
    The wrong `-commit` is the defect. `stats/internal/guard/panelfixsingledispatch.go` reads only
    each row's key, role and token. A second `flow record dispatch end` on a closed row rewrites its
    `commit_sha` (`EndDispatch`, `stats/internal/store/records.go`), so re-ending is the remedy.
  - C (Task 3): gymie's archived kan-952 `narrative.md`: "`docker logs --since` sliced the Postgres
    log wrongly; psql marker statements around each request gave stable slices"; its
    `live-verification.md` and the Postgres log show the method by hand — `ALTER DATABASE … SET
    log_statement`, marker selects, `pg_terminate_backend` on reset. gymie-postgres logs to stderr
    (`log_destination stderr`, `logging_collector off`) with prefix `%m [%p] `; local `docker exec
    psql -U gymie` needs no password.
- Spiked at plan time: Tasks 1 and 2 applied together to this worktree — `go vet`, `gofmt -l`, and
  `go test ./internal/guard/ -count=1` pass; `TestCheckPanelFixSingleDispatchCommitIsATaskCommit`
  fails with `panelfixsingledispatch.go` reverted (`expected exit 1, got 0`);
  `check-verbatim-moves.sh` (with this change's `verbatim-moves.txt`), `check-references.sh`,
  `check-dispatch-paragraphs.sh`, `check-installed-citations.sh`, `check-guard-symlinks.sh`,
  `check-stage-mark-calls.sh`, `check-ask-silence.sh` and `check-markdown-integrity.py` exit 0, and
  `check-normative-inventory.sh` prints the same inventory before and after. Task 3 applied to the
  gymie worktree: its harness passes and fails with the statement filter removed, and the script
  run live against the running kan-952 stack (`gymie_kan_952_week_fd3b`, backend 8361) printed
  `http=200 statements=0 begin=0` for `/actuator/health` — whose only measured line was the pool's
  empty validation statement — and left the database's setting reset. Every edit was reverted.

## Decisions

### Every run writing IN_PROGRESS checks itself for unfinished work

**ID:** unfinished-work-every-in-progress
**Status:** active
**Chosen:** the **Write `IN_PROGRESS`** paragraph's subject widens from "A fix run" to every run
writing `IN_PROGRESS`, a creating run as well as a fix run; arguments, the `OUTSTANDING:` loop and
the exit-2 stop are unchanged, and the paragraph names the pass-row signal beside the appended-task
one.
**Considered:** keeping fix runs only — kan-952's creating run handed off over a missing pass row;
extending `check-panel-findings-closed.sh` with the pass-row signal — a second copy of a predicate
`check-unfinished-work.sh` already holds, and it would still miss signals three and four on a
creating run.

### The fix run checks itself at the IN_PROGRESS handoff

**ID:** fix-run-self-check
**Status:** superseded by unfinished-work-every-in-progress
**Chosen:** recorded in the archived `self-review-kan-937-step-2-frontend-logic-cardio-api-models-step`
`design.md`.
**Considered:** recorded there.
**Superseded because:** its rejection of every run rested on a creating run's panel closing its own
findings; `c89c8e7f` added the pass-row signal, which no panel close guard covers, and kan-952's
creating run handed off over a round-1 panel dispatch with no pass row.

### A panel-fix row's commit is a task commit or the round's own fix commit

**ID:** panel-fix-commit-is-landing
**Status:** active
**Chosen:** the guard judges each of this token's panel-fix rows that recorded a commit: the commit,
found in the worktree or a peer the spec root's `peers` file names, must carry a task's
`**Commit:**` subject from the change's `tasks.md` (the fold route) or a `Flow-Key:` trailer naming
the row's key, `-retry` stripped (the pushed route); otherwise exit 1 naming the row. A non-hex
recorded value is a violation and never reaches git; a `tasks.md` that cannot be read while such a
row exists is exit 2; a row with no commit is not judged. `review-panel.md`'s `-commit` sentence
names both routes, and a commit violation is not the Continue prompt: the parent re-ends the row
with the commit the fix landed in and re-runs the guard.
**Considered:** the task subject alone, as the finding proposed — refuses every fix round on a
pushed branch, whose route is a standalone `Flow-Key:` commit (50af0ddf's own shape); folding the
standalone commit as the remedy — rewrites pushed history the route forbids; the worktree alone —
misses a fix committed in a satellite; a new guard — a second read of the same rows at the same
close.

### gymie counts per-request SQL with a marker-sliced script

**ID:** request-query-counter
**Status:** active
**Chosen:** `scripts/count-request-queries.sh <database> [<curl-option>...] -- <url>...` in gymie:
`ALTER DATABASE … SET log_statement = 'all'` on a `gymie_<id>` database only, the database's
sessions terminated so the pool reconnects with it, one unmeasured call per url, then the measured
call between two psql marker selects; `docker logs` is polled until the end marker appears and
sliced between the markers, counting `statement:`/`execute …:` lines from the database's own
pids, skipping the empty validation statement and the driver's `SET application_name`; reset and
terminate again on exit. Its `count_slice` is sourced by `scripts/test-count-request-queries.sh`,
the `smoke-check.sh` pattern. One paragraph in gymie `.flow/project.md` `## run` points at it.
**Considered:** `ALTER SYSTEM` plus reload — cluster-wide, logging every workspace on the shared
Postgres; a Gradle task in `gradle/dev-lifecycle.gradle.kts` — a 3000-line file for a 100-line
shell tool with no Gradle state to share; slicing by `docker logs --since` alone — kan-952's
failure; no `.flow/project.md` pointer — the next live check would not find it.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `medium` — Task 2 adds an exit-1 class to a guard every panel close runs, and Task 3
alters a database setting on the shared Postgres; a low-effort read could pass over either.
**Considered:** `low` — Task 1 is one paragraph, but Tasks 2 and 3 carry the review risk.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** one group, g1: Tasks 1–3 in plan order, one fixer — the tasks share no file, each is a
spiked diff, and Task 3 commits in the gymie worktree rather than this one.
**Considered:** two groups splitting Task 3 off by repository — one more dispatch for a diff applied
with `git apply` in a named directory.

## Open questions

## Live check

None for this change's run: Tasks 1 and 2 change prose and a guard covered by Go tests against git
fixtures, and Task 3's script was itself run live at plan time against the kan-952 worktree stack
(**Context**); it runs again whenever a gymie live check uses it.
