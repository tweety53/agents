## Context

Self-review fix change for `kan-1059-repeat-row-keep-awake-calendar-swipe`, planned on
`self-review-kan-1059-repeat-row-keep-awake-calendar-swipe` at agents `a5e4f917` (origin/main). Seven
non-`big` findings, as `proposal.md` names them, each carried by one task in `tasks.md`: 1a (Task 1),
1b (Task 2), 1c (Task 3), 1f (Task 4), 5a (Task 5), 1d (Task 6), 1e (Task 7). Each was traced before
its fix was designed; every question was auto-resolved to its recommended option (**Auto-resolution**,
`skills/flow-contracts/operator-prompts-auto-resolution.md`). The evidence is the pass's
`/Users/tweety53/Projects/gymie-worktrees/kan-1059-repeat-row-keep-awake-calendar-swipe/.superpowers/sdd/self-review-pass.md`;
the traces are in `.superpowers/sdd/2026-10-29-self-review-kan-1059-repeat-row-keep-awake-calendar-swipe-design.md`
in this worktree.

Tasks 1–5 land in this worktree. Tasks 6 and 7 are gymie-playwright test-harness code and land in
`/Users/tweety53/Projects/gymie-playwright-worktrees/kan-1059-repeat-row-keep-awake-calendar-swipe`
on its branch `spectre/kan-1059-repeat-row-keep-awake-calendar-swipe`, riding that change's landing
(decision `playwright-fixes-ride-kan-1059`).

A fix repairs what its defect already wrote: Task 5 stamps the two commit-less rows of kan-1059 with
`flowd retime-dispatch` against the dev `flow` database — dev-workspace storage, not production.
`flowd` on 4173, the `flow-postgres` container and the `flow` database are never stopped, restarted
or dropped. The two kan-1059 panel-fix commits keyed a round low (da6bcf90, b1b5a22b) are left as
committed: Task 1's acceptance is what makes their recorded repair pass.

## Decisions

### The panel-close guard accepts a trailer one round low

**ID:** pfd-one-round-low
**Status:** active
**Chosen:** `pfdCheckCommits` (`stats/internal/guard/panelfixsingledispatch.go`) also accepts, on a
row `panel-fix-<r>[-<n>][-retry]`, a commit whose `Flow-Key:` trailers include
`panel-fix-<r-1>[-<n>]` — the row signal eleven of `check-unfinished-work` prescribes when it
back-fills a one-round-low commit under the real round's key (kan-1044 self-review, F1-c). Any
other round is still refused. The comment above the function names signal eleven as the reason.
<!-- measured: scripts/check-panel-fix-single-dispatch.sh /Users/tweety53/Projects/gymie-worktrees/kan-1059-repeat-row-keep-awake-calendar-swipe kan-1059-repeat-row-keep-awake-calendar-swipe mf-k1059-20261029T094524Z → exit 1, panel-fix-5 (da6bcf90) and panel-fix-7 (b1b5a22b) @ a5e4f917 -->
**Considered:** changing signal eleven to demand the trailer's own key — the kan-1044 and kan-1059
commits are already on their branches and cannot be re-keyed without rewriting landed history;
accepting any `panel-fix-*` trailer — would hide a commit credited to the wrong fix entirely.

### The re-review refuses a panel-fix trailer of another round

**ID:** rereview-refuses-offround-trailer
**Status:** active
**Chosen:** `runRecordDispatchBegin` (`stats/cmd/flow/record.go`), for a live (no `-started-at`)
reviewer key `panel-<r>-*` with `r ≥ 1`, reads the run record and takes the earliest `StartedAt`
among round `r-1`'s reviewer rows (`records.PanelDispatchRound`); in every root `mutationRoots`
names, it lists `git log --since=<that instant> --format=%H %at %(trailers:key=Flow-Key,valueonly,separator=%x20) HEAD`
and refuses, exit 2, when a commit whose author time is at or after that instant carries a
`Flow-Key: panel-fix-<k>[-<n>][-retry]` with `k ≠ r`, naming the commit, its trailer and the fix:
amend the trailer to `panel-fix-<r>` before the re-review. Author time, not committer time, since a
fixup fold rewrites committer dates of every later commit. A record read that fails, or a git read
that fails, refuses nothing. This catches the slip while the commit is still the newest and
amendable — kan-1059's panel-fix-5 and panel-fix-7 commits carried `panel-fix-4` and `panel-fix-6`
when panel rounds 5 and 7 began.
**Considered:** refusing `flow record dispatch end`/`backfill` on a trailer mismatch — the slip in
kan-1059 had no row at all until after the closing verify, and a backfill refusal would also refuse
signal eleven's own repair; a commit-msg hook — the parent commits with `-m "Flow-Key: …"` from any
worktree of any project, and no hook is installed into project repositories.

### The re-review refuses while mutation rows are missing

**ID:** rereview-needs-mutation-rows
**Status:** active
**Chosen:** the same live `panel-<r>-*` begin (`r ≥ 1`) refuses, exit 2, while round `r-1` holds a
finding of severity Critical or Important with status `fixed` and the run holds no mutation row at
round `r` — the fix round; `uwUnprovedFixRounds`'s `m.Round-1` slack (`stats/internal/guard/unfinishedwork.go`)
is not copied, since at begin time a row at round `r-1` is fix round `r-1`'s proof of round `r-2` and
would admit the re-review vacuously (self-review review 1, I1) — evaluated on the run record already read — naming the round and the remedy: one
`flow record mutation -round <r>` per fix-mutation line, or one `-mutated none -test <reason>` row.
`status fixed` is recorded before the re-review (`skills/flow/review-panel-fix-round.md`), so the
predicate is live at begin time. One sentence beside the existing ordering rule in
`skills/flow/review-panel-fix-round.md` states it.
**Considered:** a check in `panel-round-begins.sh` — a single-slot begin outside that script (the
late-fix re-review's `primary`) would escape it; leaving it to `-late` — the order is then policed
only after the reviewers already read the tree.

### A non-recurring verify failure is fixed, never closed

**ID:** gate-keeps-loop-rule
**Status:** active
**Chosen:** `skills/flow/unfinished-work-gate.md`'s "item no change to the tree can finish" keeps
only a verdict the guard reaches over a report line naming a passing spec; a verify failure that does
not recur is sent to **The loop**'s "fixed anyway" path (`skills/flow/verify-fix-loop.md`), the
stricter rule. No guard encodes the removed run count.
**Considered:** keeping the five-run route and carving the loop's rule — the loop's rule is canonical
for every fixer of an observed failure and the gate's route is how kan-1059 landed with wr-E3b's
candidate unfixed.

### The paused-clock wait is bounded by ticked time

**ID:** paused-clock-tick-bound
**Status:** active
**Chosen:** `onPausedClock` (`fidelity/kan-932-forms.spec.ts`) drops the wall-clock `deadline` and
throws once `ticked` reaches `TAG_TIMEOUT_MS` of fake-clock time with `until` still not live,
the message naming ticked clock time. Each iteration still yields its `waitForTimeout` of real
time, so the network-bound tag keeps a floor of real time on an idle machine and more on a loaded
one, where the wall bound gave it fewer ticks. Proven by a forced
reproduction — CDP `Emulation.setCPUThrottlingRate` on the page, uncommitted — failing before the
fix and passing after.
<!-- predicted: 150 iterations × 50 ms = 7.5 s minimum real time, from TAG_TIMEOUT_MS 15_000 (fixtures.ts:685) and FRAME_TICK_MS 100 (kan-932-forms.spec.ts:137) at gymie-playwright 4ed55ad; confirmed by Task 6's step 4 run -->
**Considered:** raising `TAG_TIMEOUT_MS` — widening a tolerance (`rules/fix-determinism-at-the-source.mdc`);
waiting on the catalog response with `page.waitForResponse` — which request gates the tag is
unverified, and the rows' rise is clock-driven anyway.

### The API default follows GATEWAY_PORT

**ID:** api-default-gateway-port
**Status:** active
**Chosen:** `export const API = process.env.BACKEND_URL ?? \`http://localhost:${process.env.GATEWAY_PORT ?? '8080'}\``
in `api.ts`; a `BACKEND_URL` naming another port still throws in `fixtures.ts`, as intended. A node
test beside `/Users/tweety53/Projects/gymie-playwright-worktrees/kan-1059-repeat-row-keep-awake-calendar-swipe/scripts/fresh-identity.test.mjs` pins it.
**Considered:** dropping `fixtures.ts`'s throw — it guards the real mistake of pointing
`BACKEND_URL` at the app server.

### Class.method test tokens match by segment

**ID:** tests-class-dot-method
**Status:** active
**Chosen:** a declared `**Tests:**` token that is not found literally and matches
`^[A-Z][A-Za-z0-9_]*\.[a-z][A-Za-z0-9_]*$` is found when both its segments are — in the diff
(`tcfCheckTests`) and in the tree (`tcfMissingTreeNames`), `stats/internal/guard/`;
`tcfTestsParseRule` and the `**Tests:**` bullet of `skills/flow/brainstorm-planner.md` say so.
**Considered:** the method segment alone — a method name shared by two classes would pass for the
wrong one; the class matched by file name — a second lookup for what two substring checks answer.

### close-task refuses an appended task whose rows carry no commit

**ID:** closetask-appended-needs-commit
**Status:** active
**Chosen:** `ctUndispatched` (`stats/internal/guard/closetask.go`) also reports an appended task
every naming row of which carries no `commitSha`; close-task refuses it, exit 1, nothing ticked or
pushed, naming each row's id and the repair, `flowd retime-dispatch <id> <started-at> <ended-at> <the task's sha>`,
or `-commit` on the back-fill. Refuse, not stamp: close-task writes the store only through
`flow record dispatch end`, which never edits a closed row. The repair of kan-1059 runs
`flowd retime-dispatch` on rows 8623 and 8632, instants unchanged, commits
`f595824df4329065b6fbc58d9c2239048e4b45ec` and `c3f7de89d1c2f416dcc39d24248afd3e96ffa614`.
No subagent of the run was live inside either row's window, so the parent ran both inline:
`flowd restamp-dispatch <id> <parent transcript>` marks each row inline and attributes its tokens,
the parent transcript being
`/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/19e8be9a-7414-433b-8e3e-a443871be918.jsonl`.
<!-- measured: flow record dispatches -change kan-1059-repeat-row-keep-awake-calendar-swipe → seq 22 id 8623, seq 29 id 8632, commitSha null, agentId null @ 2026-10-29 dev store -->
<!-- measured: first and last timestamps of every subagents/agent-*.jsonl of session 19e8be9a; none overlaps 12:01:55Z–12:18:42Z or 13:05:00Z–13:06:55Z @ 2026-10-29 machine-local transcripts, not re-runnable once pruned -->
**Considered:** stamping the commit from close-task — a direct store write from a guard;
`flowd restamp-dispatch` alone — it marks the row inline and attributes tokens but never sets a
commit.

### Gymie-playwright fixes ride kan-1059's branch

**ID:** playwright-fixes-ride-kan-1059
**Status:** active
**Chosen:** Tasks 6 and 7 commit on `spectre/kan-1059-repeat-row-keep-awake-calendar-swipe` in the
gymie-playwright worktree, as the dispatch directs; 1d is recorded in scope as test-harness code.
**Considered:** a separate gymie-playwright branch — a second landing for two test-harness lines.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Tasks 1 and 2 add refusals to `flow record dispatch begin` that every
panel round passes through, and Task 5 one to every close of an appended task; a condition wrong by
one stops every run.
**Considered:** `flow-low` — the diffs are small, but the refusals sit on every run's path.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** two groups in series: `-fix-g1` Tasks 1–5, the agents worktree (Go guards, the `flow`
CLI, prose, the live repair); `-fix-g2` Tasks 6–7, the gymie-playwright worktree.
**Considered:** one fixer for all seven — one context across two repositories and a running dev
stack; one per task — Tasks 1 and 2 share `record.go` and must run in one context anyway.

## Open questions

## Live check

Each step runs before landing, with every task committed.

1. **Panel close.** `scripts/check-panel-fix-single-dispatch.sh /Users/tweety53/Projects/gymie-worktrees/kan-1059-repeat-row-keep-awake-calendar-swipe kan-1059-repeat-row-keep-awake-calendar-swipe mf-k1059-20261029T094524Z`
   from this worktree — before exit 1 naming panel-fix-5 and panel-fix-7 (measured at plan time),
   after exit 0. It reads the dev store read-only. Not working looks like the same two lines.
2. **Repaired rows.** `flow record dispatches -change kan-1059-repeat-row-keep-awake-calendar-swipe -C /Users/tweety53/Projects/gymie-worktrees/kan-1059-repeat-row-keep-awake-calendar-swipe`
   — rows 8623 and 8632 carry `commitSha` `f595824d…` and `c3f7de89…`; before both null. Not working
   looks like null.
3. **Begin refusals.** Not run against real data: kan-1059's records already carry round 7, so the
   panel-round floor refuses any re-review begin of an earlier round before Tasks 1 and 2's checks
   are reached; the Go tests over a fake store are the evidence.

## Recommended approach

Fix each defect where it starts: the panel-close guard's trailer match and, upstream, the re-review's
begin, where both the off-round trailer and the missing mutation rows are still repairable; the gate
prose; the plan guard's test-name match; close-task's appended-task check plus a one-off
`retime-dispatch`; the spec's wait bound; the harness's default origin. Rejected alternatives are
under each decision.

## Critical files

- `stats/internal/guard/panelfixsingledispatch.go`, `stats/internal/guard/check_panel_fix_single_dispatch_test.go` — 1a acceptance.
- `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go` — 1a and 1b begin refusals.
- `skills/flow/review-panel-fix-round.md` — the two begin rules.
- `skills/flow/unfinished-work-gate.md` — 1c.
- `stats/internal/guard/taskcommitfields.go`, `stats/internal/guard/check_task_commit_fields_test.go`, `skills/flow/brainstorm-planner.md` — 1f.
- `stats/internal/guard/closetask.go`, `stats/internal/guard/close_task_test.go`, `scripts/close-task.sh` — 5a.
- gymie-playwright `fidelity/kan-932-forms.spec.ts` — 1d.
- gymie-playwright `api.ts`, `scripts/api-origin.test.mjs` — 1e.

## Existing code to reuse

- `tcfRunGit`, `pfdShaRE` in `panelfixsingledispatch.go`.
- `records.PanelDispatchRound`, `mutationRoots`, `callRecord`, `panelFixKeyedARoundLow`'s shape in `record.go`; its test fakes in `record_test.go` (`TestRecordDispatchRefusesAPanelFixKeyedARoundLow`).
- `uwUnprovedFixRounds` (predicate mirrored), `uwPanelFixKey`.
- `tcfFoldedPattern`, `tcfFoldWS`, `tcfFoldedTreeText`.
- `ctUndispatched`, `uwRowTasks`, `pcFlow`; `TestCloseTaskRefusesAnUndispatchedAppendedTask`'s fixture.
- `flowd retime-dispatch` (`stats/cmd/flowd/main.go`, `store.Store.RetimeDispatch`).
- `/Users/tweety53/Projects/gymie-playwright-worktrees/kan-1059-repeat-row-keep-awake-calendar-swipe/scripts/fresh-identity.test.mjs`'s import of `../api.ts`; `npm run test:scripts`.

## Verification

Beyond each task's verify step: `scripts/run-guard-tests.sh` and `cd stats && make -s web-dist &&
go test -race -count=1 $(go list ./... | grep -v '/internal/guard$')` pass; the agents `## lint` list
passes; in gymie-playwright, `npm run test:scripts` passes and wr-E3b passes under the CPU throttle;
the **Live check** above.
