## Context

Self-review fix change for `kan-1048-rework-live-board-and-run-view`, planned on the change's own
branch `spectre/kan-1048-rework-live-board-and-run-view` at agents `46092b9f` (the fix-base). Seven
non-`big` findings, as `proposal.md` names them, each carried by one task in `tasks.md`: 1-A
(Task 1), 1-B (Task 2), 1-D (Task 3), 1-E (Task 4), 1-F (Task 5), 1-J (Task 6), 8-A (Task 7). Each
was traced before its fix was designed; every question was auto-resolved to its recommended option
(**Auto-resolution**, `skills/flow-contracts/operator-prompts-auto-resolution.md`). The brainstorm
design, with the traces, is
`.superpowers/sdd/2026-10-29-self-review-kan-1048-rework-live-board-and-run-view-design.md` in this
worktree.

A fix repairs what its defect already wrote: Task 5 runs `flowd restamp-ends` against the dev `flow`
database — dev-workspace storage, not production. `flowd` on 4173, the `flow-postgres` container and
the `flow` database are never stopped, restarted or dropped. The kan-1048 mutation reproducers are
gitignored records of a closed panel and are not rewritten; the archived kan-1048 plan is history and
is not re-planned.

## Decisions

### The re-run number ends at a dash or the key's end

**ID:** verify-rerun-anchor
**Status:** active
**Chosen:** `vgFixSuffix` becomes `^verify-(fix-)?[0-9]+(-|$)`, replaced by `verify$2`, so the
group keeps its `-<basename>` and a digit-led repository name is never read as a re-run number.
<!-- measured: go run of both regexes over verify-3d-viewer, verify-2-3d-viewer, verify-fix-1-3d-viewer, verify-2, verify-fix-1, verify-2-fe, verify-fe @ 46092b9f -->
**Considered:** `\b` after the number — `3d` has no word boundary between `3` and `d`, but `[0-9]+\b`
is still one rule harder to read than the explicit alternation; parsing the key into parts — a
second parser for a one-line match.

### One test checks every guard's sibling execs against its shim

**ID:** shim-siblings-every-guard
**Status:** active
**Chosen:** `TestRunVerifyListShimNamesItsSiblings` becomes `TestGuardShimsNameTheirSiblings`,
in the same file renamed `stats/internal/repotest/guard_shim_siblings_test.go`. For every non-test
`stats/internal/guard/*.go`, it collects the sibling names the source execs — `scriptDir, "<x>.sh"`
(covers `filepath.Join` and `trsdSibling`), `scriptDir+"/<x>.sh"`, `sibling("<x>")` and
`c.run("<x>"` — maps the file to its guard id by the `Registry["<id>"] =` in the file, or in the file
named by dropping its `_<suffix>` (`integratechange_land.go` → `integratechange.go`), and requires
`scripts/<id>.sh` to spell `$SCRIPT_DIR/<x>.sh` for each. A source with sibling execs and no guard id
or no shim fails. The test fails when it collects no sibling at all.
**Considered:** every `"<x>.sh"` literal in a guard source — 7 false hits on prose and error
messages at `46092b9f`; a guard in `check-guard-symlinks` — a test is enough for a lexical
invariant of this repository's own sources.

### A mutation reproducer runs under errexit

**ID:** mutation-reproducer-errexit
**Status:** active
**Chosen:** `runReproducer` (`stats/internal/guard/runreproducer.go`), right after
`declaresMutation` returns true, refuses unrun a reproducer with no line matching
`^\s*set\s+(-[A-Za-z]*e[A-Za-z]*|-o\s+errexit)\b` — the refused class, exit 2 — naming the fix:
"add `set -e` so an edit that fails exits non-zero instead of reading as a surviving mutant". A
`ponytail:` comment names the ceiling: lexical, so a later `set +e` or an edit inside `if`/`||`
escapes it. `check-panel-reproducer-exit-contract` already reports a refused reproducer as a
violation, so the prose rule of `skills/flow/review-panel.md` (the mutation-testing brief) is now
enforced at both the panel gate and every fix-dispatch verdict. The brief and
`skills/flow/test-sanity-reviewer-prompt.md` state the `set -e` requirement; `scripts/run-reproducer.sh`'s
header names it under exit 2. Existing test fixtures that declare the marker gain a `set -e` line
after it.
<!-- measured: scripts/run-reproducer.sh "$PWD" .superpowers/sdd/reproducers/zz-spike.sh (marker, failing python3 assert, true) → exit 0 "defect demonstrated" @ 46092b9f -->
**Considered:** a runner-side landing proof (diffing the tree mid-run) — the reproducer restores on
`EXIT`, so nothing is left to read; matching a `|| exit` after each edit command — needs a parser for
every edit tool's shape; a check in `check-panel-reproducer-exit-contract` alone — the fix-dispatch
verdicts would still read a failed edit as survived.

### Plan-shape refuses a planning path in a task's fields

**ID:** plan-shape-planning-entry
**Status:** active
**Chosen:** a new row F23 of `check-plan-shape` (`stats/internal/guard/planshape.go`): a task whose
`**Files:**` or `**Allowed-collateral:**` holds an entry starting with one of `psPlanning`'s prefixes
(`spectre/changes/`, `openspec/changes/`, `docs/superpowers/`) and not ending `visual-verification.md`
is refused at that field's line, naming the entry and saying the path rides a planning or
`Flow-Key: flow.visual-verify` commit, never a task commit — the exemption
`check-task-commit-planning-paths` makes (`taskcommitplanningpaths.go`). F14 stays (it names the
all-planning case). `skills/flow/verify-fix-loop.md` step 1 adds: the round's evidence directory,
`<changeRoot>/visual-verification/`, is never in the task's fields; the stage's step-12 commit
(`skills/flow/visual-verify.md`) carries it.
**Considered:** exempting `visual-verification/**` in the commit guard — the evidence directory
rides the stage's own commit already, so a task commit has no reason to carry it; a check at commit
time only — the plan is where the wrong declaration is cheapest to fix.

### flowd restamps recent stamped ends at startup

**ID:** startup-restamp-window
**Status:** active
**Chosen:** `run` in `stats/cmd/flowd/main.go`, once `agentEnd` is built, calls a new
`restampRecentEnds(ctx, st, now, agentEnd, logger)` that runs `st.RestampStampedEnds` with
`since = now - 7 days` (`startupRestampWindow`) and logs `flowd restamped stamped dispatch ends`
with the count and `since`; an error is logged as a warning and never stops startup. Seven days
covers any daemon left running on an older build between `make restart`s; rows with no transcript
stay candidates until they age out of the window. Repair: `flowd restamp-ends 2026-10-28T00:00:00Z`
once, from the task.
<!-- measured: docker exec flow-postgres psql -U flow -d flow -tAc "select count(*) from dispatches where ended_at >= '2026-10-28T00:00:00Z' and agent_id is not null and agent_id <> 'inline' and ended_at <> date_trunc('milliseconds', ended_at)" → 19 @ 2026-10-29 -->
**Considered:** restamping the whole table at startup — every startup re-reads transcripts for
every row the restamp could not move, unbounded; a periodic sweep — a timer for what only a daemon
older than its build produces.

### The activity hook ignores subagent tool events

**ID:** activity-ignores-subagents
**Status:** active
**Chosen:** `activity()` in `hooks/flow-session-activity.py` returns `None` for a `PreToolUse` or
`PostToolUse` payload carrying a non-empty `agent_id` — the field Claude Code adds to a subagent's
tool events, absent on the main thread's (https://code.claude.com/docs/en/hooks: "the input carries
the `agent_id` and `agent_type` common input fields that identify the subagent"). `Notification`,
`UserPromptSubmit` and `Stop` are unchanged: a subagent's permission prompt still blocks the
operator. A parallel main-thread tool's `PostToolUse` while a question is open is not traced to a
defect here and stays as is.
**Considered:** `agent_type` — present on the main thread under `--agent`; dropping every event with
`agent_id` — would hide a subagent's permission prompt from the board.

### A one-frame departure is confirmed before it is reported

**ID:** one-frame-departure-confirmed
**Status:** active
**Chosen:** a paragraph in the motion step of `skills/flow/visual-verify-verifier.md`, after the
frame rules: a departure lasting one frame or less is reported only once confirmed — by a DOM read at
the next animation frame after the trigger (`page.evaluate` resolving inside a
`requestAnimationFrame` callback, reading the departing element's presence or computed style) or by
a second recording showing the same frame; an unconfirmed one is not a departure. Cites kan-1048
self-review 8-A.
**Considered:** dropping single-frame departures outright — a real one-frame flash is a departure;
a rule in `visual-verify.md` — the verifier, not the stage, judges the strip.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 3 changes what the reproducer runner refuses for every panel and
fix dispatch, Task 4 what every plan's shape guard refuses, and Task 5 writes the live store and the
daemon's startup.
**Considered:** `flow-low` — the diffs are small, but a refusal wrong by one condition stops every
panel or plan.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** two groups in series: `-fix-g1` Tasks 1–4, the guard and repotest work in
`stats/internal/guard` and `stats/internal/repotest`; `-fix-g2` Tasks 5–7, the `flowd` startup with
its live repair, the hook and the verifier prose.
**Considered:** one fixer for all seven — one context carrying guard code, a live store write and
prose; one per task — seven dispatches for file-disjoint tasks of a few lines each.

## Open questions

## Live check

Each step runs before landing, with every task committed, in the worktree.

1. **Stamped ends.** `docker exec flow-postgres psql -U flow -d flow -tAc "select count(*) from
   dispatches where ended_at >= '2026-10-28T00:00:00Z' and agent_id is not null and agent_id <>
   'inline' and ended_at <> date_trunc('milliseconds', ended_at)"` — before 19 (measured at plan
   time); after, 19 less the count Task 5's `restamp-ends` printed. kan-1048's three `panel-0-*` rows
   no longer share the 22:17:12Z end. Not working looks like 19 and the shared end.
2. **Startup restamp.** `scripts/live-store-copy.sh up k1048sr` builds `flowd` from this checkout and
   starts it on a copy of the dev store; its log under `/tmp/flow-live-k1048sr/` carries one
   `flowd restamped stamped dispatch ends` line with `since` seven days before the start. Not working
   looks like no such line.
3. **Hook.** With step 2's copy still up and its printed `FLOW_ADDR` exported, pipe to
   `python3 hooks/flow-session-activity.py`, in order: an `AskUserQuestion` `PreToolUse` for session
   `s-live-k1048sr`, then a `PostToolUse` of tool `Bash` carrying `"agent_id":"a1"`. `docker exec
   flow-postgres psql -U flow -d flow_k1048sr_live -tAc "select state from session_activity where
   session_id = 's-live-k1048sr'"` reads `question`; a third payload, the same `PostToolUse` without
   `agent_id`, turns it `working`. Not working looks like `working` after the second payload. Then
   `scripts/live-store-copy.sh down k1048sr`.

## Recommended approach

Fix each defect where it starts: the verify re-run regex, one repotest sweep over every guard's
siblings, the reproducer runner (the one place every mutation verdict passes), the plan-shape guard
and the loop prose, the daemon's startup plus a one-off `restamp-ends`, the hook's event filter, and
the verifier's motion step. Rejected alternatives are under each decision.

## Critical files

- `stats/internal/guard/verifygreen.go`, `check_verify_green_test.go`.
- `stats/internal/repotest/run_verify_list_siblings_test.go` → `guard_shim_siblings_test.go`.
- `stats/internal/guard/runreproducer.go`, `runreproducer_test.go`, `scripts/run-reproducer.sh`, `skills/flow/review-panel.md`, `skills/flow/test-sanity-reviewer-prompt.md`.
- `stats/internal/guard/planshape.go`, `check_plan_shape_test.go`, `skills/flow/verify-fix-loop.md`.
- `stats/cmd/flowd/main.go`, `wiring_test.go`.
- `hooks/flow-session-activity.py`, `stats/internal/repotest/flow_session_activity_hook_test.go`.
- `skills/flow/visual-verify-verifier.md`.

## Existing code to reuse

- `declaresMutation`, `refuse` in `runReproducer`.
- `psPlanning`, `tcfBodyFields`, `p.at`, `p.on` in `planshape.go`; `psFixture`, `runGuard` in its test.
- `store.Store.RestampStampedEnds`, `restampEnds`, `harvest.AgentTranscriptEnd`; `storetest.Main`, `storetest.NewStore`.
- `repoRoot` in `stats/internal/repotest`.
- `scripts/live-store-copy.sh`, `scripts/break-and-prove.sh`.

## Verification

Beyond each task's verify step: `scripts/run-guard-tests.sh` and `cd stats && make -s web-dist &&
go test -race -count=1 $(go list ./... | grep -v '/internal/guard$')` pass; the agents `## lint` list
passes; the **Live check** above.
