## Context

Self-review fix change for `kan-1051-rework-the-trend-dashboard`, planned on that change's own branch
`spectre/kan-1051-rework-the-trend-dashboard` at `1f07f9ce` (the pass's `<fix-base>`), in
`/Users/tweety53/Projects/agents-worktrees/kan-1051-rework-the-trend-dashboard`. Six non-`big` findings,
as `proposal.md` names them: F7 (Task 1), F4 (Task 2), F5 (Task 3), F6 (Task 4), F18 (Task 5), F17
(Task 6). Each was traced on that HEAD before its fix was designed; every question was auto-resolved to
its recommended option (**Auto-resolution**, `skills/flow-contracts/operator-prompts-auto-resolution.md`).
The traces, the F7 session binding and the `go test -overlay` spike are in
`.superpowers/sdd/2026-11-01-self-review-kan-1051-rework-the-trend-dashboard-design.md` in this worktree.

`flowd` on 4173, the `flow-postgres` container and the `flow` database are never stopped, restarted or
dropped. Task 1's repair writes one dispatch row through the existing `flowd restamp-dispatch`
subcommand, which opens the store directly and leaves the daemon running.

## Decisions

### A loop body ends at any list terminator before `done`

**ID:** loop-body-terminators
**Status:** active
**Chosen:** `scanLoopBody` (`stats/internal/harvest/watcher.go`) ends a body at `;`, `&` or `\n`
followed (after spaces) by the keyword `done`, returning the body up to that terminator. verify-5's
call ended `… ) & done`, so the loop was never expanded and the settle pass made no candidate for row
9234. The finding's own premise is disproved: the six records in the window belong to session
`4cfa74e7-cc40-4657-8c45-4fe2a33a3c0e` (kan-1073), not to `82eaa458-f303-4fbb-bbf5-ed58a6d64147`, the
session the run's token binds, which holds none there — so the correct value is zero tokens, not 3,322.
**Considered:** crediting the window from every session's transcript — would charge another run's
spend to this one; `&` alone — a multi-line `for` body ending on a newline before `done` misses the same
way, and newline is the third bash list terminator.

### Row 9234 is repaired with `flowd restamp-dispatch`

**ID:** repair-row-9234
**Status:** active
**Chosen:** after the fix commits, Task 1 runs a `flowd` built from this worktree:
`restamp-dispatch 9234 ~/.claude/projects/-Users-tweety53-Projects-agents/82eaa458-f303-4fbb-bbf5-ed58a6d64147.jsonl`,
which marks the agentless row inline and writes `ClosedInlinePatch` over the parent transcript — the
zero patch, since that session has no record in the window.
**Considered:** a flowd restart so the settle pass reruns — harvest offsets are committed, so the old
command is never read again; `flow record dispatch end -tokens` — refused on a harvested session
(`ErrTokensOnHarvestedSession`).

### The since-close sha after a restore is the restoring planning commit

**ID:** since-close-after-restore
**Status:** active
**Chosen:** in `skills/flow/review-panel-late-fix.md`, after integrate's reshape the canonical worktree's
since-close sha is the drop commit — or, when a fix run's planning commit restored the change directory
after that drop (**A change already dropped is never asked**, `skills/flow/document-fix.md`), that
planning commit. Rounds 4 and 5 of kan-1051 used exactly those (`edc923af`, `97e45598`).
**Considered:** teaching `check-late-fix-trigger.sh` to discount the restored directory — the sha is
the caller's input and the contract names it, so the contract sentence is the defect.

### Reproducers never write the shared tree

**ID:** reproducer-writes-nothing-in-tree
**Status:** active
**Chosen:** `skills/flow/review-panel.md`'s reproducer rule and its slot-facing REPRODUCER paragraph
state that a reproducer writes nothing under `$PWD`: a Go test it adds is passed to
`go test -overlay <json>` from a file under `mktemp -d`, and any other file it writes lives in its own
`mktemp -d` directory. The `reproducer` row of `dpEntries`
(`stats/internal/guard/dispatchparagraphs.go`) gains the phrase `go test -overlay`, so
`check-dispatch-paragraphs.sh` fails while the paragraph lacks it.
<!-- measured: go -C stats test -overlay <json> -run TestZZOverlaySpike -count=1 -v ./internal/harvest, overlay adding stats/internal/harvest/zz_overlay_spike_test.go from mktemp -d — PASS, git status --porcelain stats/ empty @ 1f07f9ce -->
**Considered:** a scratch copy of the worktree, as the test-sanity brief does — a reproducer must run
in `$PWD`, the tree `run-reproducer.sh` passes, so a copy would test another tree; a `run-reproducer`
refusal of any new untracked file — too late, a concurrent suite already saw the file.

### The pass hands back its files' contents

**ID:** pass-hands-back-contents
**Status:** active
**Chosen:** step 2 of `skills/flow-self-review/SKILL.md`: the pass returns `findings.md`,
`report-draft.md` and every case file in full in its hand-back, and this session writes them to
`<sr-dir>` before step 2a. The "paths, never contents" rule names this one exception.
**Considered:** granting the pass a write permission on `<sr-dir>` — a harness permission outside the
skill, and the harness refused on all three passes of this change.

### On merge and push, land is the verify

**ID:** land-is-the-verify
**Status:** active
**Chosen:** on a `self-review-<name>` branch, and in a pipeline-fix loop, whose route is merge and push,
the separate `pipeline-fix.sh verify` is not run: `pipeline-fix.sh land <wt> > <sr-dir>/verify-<v>.log 2>&1`
is that verify, its `red:` lines feeding the `-fix-v<v>` loop. The separate verify stays on the change's
own branch and on every other route. `pfLand` already runs `pfVerify` after its rebase, and `pfVerify`
writes no store record any landing check reads.
**Considered:** a `--no-verify` flag on land — the land's verify is the one on the rebased HEAD, the
one that must stay; skipping land's verify when HEAD equals the last verified HEAD — state nobody keeps.

### `live-store-copy.sh restart`

**ID:** live-store-copy-restart
**Status:** active
**Chosen:** a `restart <name>` verb: the pidfile-checked stop `down` performs, then the launch `up`
performs against the recorded port, binary and transcripts directory (`up` now records the latter),
with no copy and no build. After the stop it waits for flowd to release `/tmp/flowd-<port>.pid`, and
removes the file when it remains while its pid is dead. It exits 1 when flowd does not answer within
the existing 60 s poll, 2 with no prior `up`. Stop and launch are two shell functions shared by the verbs.
**Considered:** `down` then `up` — drops and recopies the database and refuses the existing scratch
change; a rebuild in `restart` — the finding asks for none, and `up` already builds.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 1 changes the harvester's command parser on every transcript and
writes a store row, Task 6 changes a dev tool's process handling, and Tasks 2–5 change contracts every
run reads.
**Considered:** `flow-low` — four of six tasks are prose, but the parser change and the process handling
need a real read.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, `-fix`, Tasks 1–6 in series in this worktree, on `opus`; Task 1 first.
**Considered:** one fixer per task — six small diffs, each paying a fresh context; sonnet for the prose
tasks — Tasks 3 and 5 touch a guard table and the landing loop, so opus throughout.

## Open questions

## Recommended approach

Fix the parser and repair the row (`loop-body-terminators`, `repair-row-9234`), then the four contract
edits (`since-close-after-restore`, `reproducer-writes-nothing-in-tree`, `pass-hands-back-contents`,
`land-is-the-verify`), then the dev tool verb (`live-store-copy-restart`). Each rejected alternative is
the `**Considered:**` line of the named decision.

## Critical files

- `stats/internal/harvest/watcher.go`, `stats/internal/harvest/watcher_test.go` — `scanLoopBody` (F7).
- `stats/cmd/flowd/main.go` — `restampDispatch`, used unchanged for the repair.
- `skills/flow/review-panel-late-fix.md` — since-close sha (F4).
- `skills/flow/document-fix.md` — the restore it cites, read only.
- `skills/flow/review-panel.md`, `stats/internal/guard/dispatchparagraphs.go` — reproducer rule (F5).
- `skills/flow-self-review/SKILL.md` — step 2 (F6) and step 4 (F18).
- `skills/flow-contracts/pipeline.md` — **Pipeline defects found mid-run** item 4 (F18).
- `stats/internal/guard/pipelinefix.go` — `pfLand`/`pfVerify`, read only.
- `scripts/live-store-copy.sh`, `scripts/test-live-store-copy.sh` — F17.

## Existing code to reuse

- `judgeInvocations`, `runVerifyListBeginFromInvocation`, `TestWatcherStampsKan925LoopBegins`'s
  `bash(...)` helper shape (`stats/internal/harvest/watcher_test.go`).
- `flowd restamp-dispatch` — the repair.
- `dpEntries` — the paragraph table `check-dispatch-paragraphs` reads.
- `live-store-copy.sh`'s `port_ok`, pidfile check and launch loop — moved into functions, not rewritten.
- `scripts/break-and-prove.sh` — mutation proofs.

## Verification

- `cd stats && go test ./internal/harvest -run 'TestJudgeInvocationsEndsLoopBodyAtAnyTerminator|TestWatcherStampsKan925LoopBegins' -count=1` passes.
- `flow record dispatches -change kan-1051-rework-the-trend-dashboard` shows row 9234 with `tokens.main` all 0.
- `scripts/check-dispatch-paragraphs.sh` exits 0, and 1 with the new sentence removed.
- `scripts/test-live-store-copy.sh` prints `all cases passed`.
- `scripts/run-verify-list.sh <worktree> lint` exits 0 after the prose tasks.

## Live check

From this worktree, after every task is committed: `scripts/live-store-copy.sh up sr1051 4176`, then
`scripts/live-store-copy.sh restart sr1051`, timed. Working: exit 0 within the 60 s poll, the address
printed, `/tmp/flowd-4176.pid` naming a new pid whose `curl -sf http://127.0.0.1:4176/api/v1/changes`
answers, and `flow_sr1051_live` still holding the copied `changes` rows. Not working: exit 1 on the poll,
or the old pid still in the pidfile. Then `scripts/live-store-copy.sh down sr1051`. Before figure: the
pass recorded a hand relaunch that exited on the old pidfile and an 8.5 min wait. Row 9234's repair is
read on the dev store: before `{}`, after `tokens.main` all 0.
