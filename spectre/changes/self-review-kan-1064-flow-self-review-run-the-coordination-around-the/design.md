## Context

Found by the self-review pass of `/flow` change `kan-1064-flow-self-review-run-the-coordination-around-the`
(findings F1, F2 and F5 of its pass report, and F4, found and reproduced by the parent in the same
run); the fixes ride that change's own branch
`spectre/kan-1064-flow-self-review-run-the-coordination-around-the`, based at its archive commit.
The brainstorm design is
`.superpowers/sdd/2026-10-31-self-review-kan-1064-flow-self-review-run-the-coordination-around-the-design.md`.

**F1 — the pipeline-fix loop's escalation is forbidden instead of stated.** The no-escalation
sentence of **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`) names
"the `fix` row"; the next paragraph maps the suffix `fix` onto both the `self-review-*` and
`pipeline-fix-*` rows, and **Allowed pairs** (`skills/flow-contracts/model-policy.md`) still carries
`pipeline-fix-fix` `opus/flow-high` `No-progress escalation`. `skills/flow-contracts/pipeline.md`
makes this step canonical for the pipeline-fix loop's no-progress escalation.

**F2 — a scratch-applied block is tagged `verified:` without the lint list.** **Write each task's
verify step** (`skills/flow/brainstorm-planner.md`) runs `run-verify-list.sh <worktree> lint` for
skills Markdown at implementation; nothing makes the planner run it on the scratch copy it already
applied the block to.
<!-- measured: this plan's own prose script applied to a scratch clone at 6107fdfd, then scripts/run-verify-list.sh <clone> lint → check-verbatim-moves.sh exit 1, 7 violations; with this change's verbatim-moves.txt → only check-installed-rules.sh exit 1, which names the clone not being the installed checkout (the worktree itself: exit 0) @ 2026-10-31 -->

**F3 — a flowd restart fails a concurrent CLI call.** `Client` (`stats/internal/client/client.go`)
"never retries" and `do` maps every transport error to `ErrUnavailable`. `make restart`
(`stats/Makefile`) kills the daemon, then runs `docker compose up -d` and `go build` of the CLI, then
starts the new daemon, so the gap spans both steps; the CLI's default timeout is 2s
(`defaultTimeout`, `stats/cmd/flow/state.go`). A dead daemon must still fail fast: 59 `cmd/flow`
tests dial `deadPortAddr`, and the guard tests point `FLOW_RECORDS_ADDR` at `127.0.0.1:1` so a
records read "fails fast" (`stats/internal/guard/helpers_test.go`), through out-of-process `flow`
calls.
<!-- measured: grep -rn "deadPortAddr(t)" stats --include='*_test.go' | wc -l → 59 @ 6107fdfd -->

**F4 — `pipeline-fix.sh end` leaves the pass open.** Since e0947d88 `check-loop-dispatch.sh` records
the pass (`flow record dispatch begin`, change `<name>`, checkout from the caller's directory,
`stats/internal/guard/loop.go`), and refuses a later dispatch while a row is open; `pfEnd`
(`stats/internal/guard/pipelinefix.go`) still answers the pass with `PIPELINE-FIX-END-NA` and calls
nothing. Existing state needs no repair: the pass rows of every change archived since then are
closed.
<!-- measured: flow record dispatches -change <c> for kan-1045, kan-1060, kan-1061, kan-1062, kan-1067 → each self-review-<c>-pass row endedAt set, outcome completed @ 2026-10-31 live flowd (read-only) -->

## Decisions

### Where the no-progress escalation is stated

**ID:** escalation-scoped-to-pipeline-fix
**Status:** active
**Chosen:** the sentence names `self-review-fix`, and a second sentence states that the `pipeline-fix-fix` row keeps the no-progress escalation on its `opus/flow-high` row, admitted after `<agents repo>/scripts/pipeline-fix.sh escalate <key>`, citing **Fewest operator actions**; `skills/flow/implement.md` names "the pipeline-fix loop's `pipeline-fix-fix` escalation row" and `skills/flow/SKILL-rationale.md` "the in-run fix loop". The reworded sentences are listed in this change's `verbatim-moves.txt`.
**Considered:** stating the escalation in `skills/flow-contracts/pipeline.md` instead — it delegates the escalation to this step as canonical, so the statement belongs here; deleting the sentence — the self-review loop's own no-escalation rule (kan-1064's design) would go unstated.

### How the planner proves a scratch-applied block

**ID:** scratch-copy-runs-lint
**Status:** active
**Chosen:** a paragraph after the plan-provenance load in section D of `skills/flow/brainstorm-planner.md`: a block verified by applying it to a scratch copy runs `<agents repo>/scripts/run-verify-list.sh <scratch copy> lint` there first, fixes red lines in the block, and the tag names the run; otherwise the block is `unverified:`. Prose only — `run-verify-list.sh` already runs the project's list on any directory.
**Considered:** a guard parsing `verified:` tags for a lint mention — the tag's text is free-form, and a guard would only check wording, not that the run happened.

### When the client retries

**ID:** retry-refused-while-restarting
**Status:** active
**Chosen:** `make restart` writes `$TMPDIR/flowd-<port>.restarting` before it kills the daemon and removes it once the new daemon answers or the restart fails; the compose and CLI-install steps move ahead of the kill and the exit poll drops from 1s to 0.1s, so the gap is the daemon swap alone. `do` retries a request whose dial was refused (`syscall.ECONNREFUSED`) every 100ms while that marker exists for the request's port and is younger than 2 minutes, returning `ErrUnavailable` once the request's context is done; a request whose body cannot be replayed (`Body` set, `GetBody` nil) is not retried. With no fresh marker a refused dial fails at once, as today.
<!-- predicted: the 100ms, 0.1s and 2-minute figures are this decision's chosen values, not measurements -->
**Considered:** retrying every refused dial within the caller's context — each dead-daemon call (59 CLI tests, the guard tests' fail-fast records reads) would wait out its full timeout; an exported switch tests turn off — it reaches no out-of-process `flow` call; retrying past the caller's timeout while the marker stands — the caller's timeout is the bound the CLI's fallback contract rests on.

### Where the pass is closed

**ID:** end-closes-the-pass
**Status:** active
**Chosen:** `pfEnd` closes a `-pass` key with `flow record dispatch end` against the change and checkout `check-loop-dispatch.sh` began it under — one helper in `loop.go` both call: change `<name>` (the key's prefix without `self-review-`), checkout the caller's directory's main checkout when it resolves, `lp.main` otherwise; outcome `completed` or `aborted`. The `PIPELINE-FIX-END-NA` branch goes.
**Considered:** keeping the parent's hand-run `flow record dispatch end` — the contract names `pipeline-fix.sh end` for every dispatch, and the guard's own refusal names it.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 3 changes the transport path every `flow` CLI call routes through and the live daemon's restart recipe.
**Considered:** `flow-low` — Tasks 1, 2 and 4 are prose and a narrow guard branch, but Task 3's retry and marker lifecycle need the closer read.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** one fixer dispatch for Tasks 1–4 — four file-disjoint tasks, two of them prose a script applies.
**Considered:** one fixer per task — three extra dispatches for small, independent edits.

## Open questions

## Recommended approach

Fix each defect where every caller routes through it: the escalation sentence the pipeline-fix loop
reads as canonical (`escalation-scoped-to-pipeline-fix`), the planner's provenance rule
(`scratch-copy-runs-lint`), the client's one `do` and the one restart recipe
(`retry-refused-while-restarting`), and `pfEnd` beside the begin it mirrors (`end-closes-the-pass`).

## Critical files

- `skills/flow-self-review/SKILL.md`, `skills/flow/implement.md`, `skills/flow/SKILL-rationale.md`.
- `skills/flow/brainstorm-planner.md`.
- `stats/internal/client/client.go`, `stats/internal/client/client_test.go`, `stats/Makefile`.
- `stats/internal/guard/pipelinefix.go`, `stats/internal/guard/loop.go`, `stats/internal/guard/pipeline_fix_test.go`.

## Existing code to reuse

- `genuineDaemon` and `deadPortURL` (`stats/internal/client/client_test.go`).
- `pidfile.Path`'s naming, `$TMPDIR/flowd-<port>.pid` (`stats/internal/pidfile/pidfile.go`), for the marker's sibling name.
- `lpMain` and the pass begin's change derivation (`stats/internal/guard/loop.go`); `pfNew` and `fx.calls` (`stats/internal/guard/pipeline_fix_test.go`).

## Verification

- `cd stats && make -s web-dist && go test -count=1 ./internal/client/ -run 'TestGetChange'`, `./internal/guard/ -run 'TestPipelineFix|TestCheckLoopDispatch'` and `./internal/repotest/ -run 'TestMakeBuild'` pass; `go vet ./...` and `gofmt -l .` are clean.
- `scripts/run-verify-list.sh <worktree> lint` is green, `check-verbatim-moves.sh` acknowledging this change's `verbatim-moves.txt`.

## Live check

None before landing — the client's retry is exercised against real listeners by Task 3's tests, and the restart recipe is the protected `flowd`'s, which no step before landing may restart; the reinstall after landing (`make -C stats restart`) is its first run, where `/tmp/flow-live.log` shows the order and `ls $TMPDIR/flowd-4173.restarting` after it prints no file.
