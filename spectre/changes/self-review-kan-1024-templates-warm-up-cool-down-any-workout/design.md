## Context

Self-review fix change for `kan-1024-templates-warm-up-cool-down-any-workout`, planned on
`self-review-kan-1024-templates-warm-up-cool-down-any-workout` at agents `c30fab36`
(`origin/main`). Ten non-`big` findings, as `proposal.md` names them, each carried by one task in
`tasks.md`: F-b (Task 1), F-i (Task 2), F-j (Task 3), F-k (Task 4), F-l (Task 5), F-m (Task 6),
F-n (Task 7), F-c (Task 8), F-d (Task 9), F-e (Task 10). Tasks 8 and 9 land in the gymie-playwright
worktree and Task 10 in the gymie worktree, each on `spectre/kan-1024-templates-warm-up-cool-down-any-workout`
(gymie-playwright at `0770c98`, gymie at `e923c6c8`). Each finding was traced before its fix was
designed, save F-c, whose task opens on the trace. Every question was auto-resolved to its
recommended option (**Auto-resolution**, `skills/flow-contracts/operator-prompts-auto-resolution.md`).
The brainstorm design, with the per-finding trace, is
`.superpowers/sdd/2026-10-28-self-review-kan-1024-templates-warm-up-cool-down-any-workout-design.md`
in this worktree.

## Decisions

### The merge order names every recorded worktree

**ID:** merge-order-from-state-record
**Status:** active
**Chosen:** after `flow state add-worktree`, `kickoffWorktree` reads `flow state get -C <from> <name>`, maps each recorded worktree to its `spectre/peers` name (the worktree's main checkout, `git rev-parse --git-common-dir`, against each peer path, as `kwPeerName` resolves the new one), skips `<from>`'s own repository, and calls `kwMergeOrder` once per name in `spectre/peers` order — the first call writes the file with `.` and that name, each later one appends a missing name. kan-1024's own `link.md` needs no repair: its archived copy already lists all four repositories.
**Considered:** listing every peer in `spectre/peers` — names repositories the change never touched; writing the names only into a new file — the same loop already appends to an existing one, so splitting the cases adds code and leaves a file created by an older kickoff short.

### A panel-fix begin needs its review round's findings and an unused key

**ID:** panel-fix-begin-needs-findings
**Status:** active
**Chosen:** in `stats/cmd/flow/record.go`, beside the panel-round-floor check and in its shape — `begin` only, never `backfill` (`-started-at` set), and a run-record read that fails skips the check — a key matching `panelFixKeyRE` with round `<r>` is refused, exit 2, when the run record holds no finding with round `<r>-1` and severity Critical or Important (case-insensitive), naming the round and `flow record finding -from`; and refused, exit 2, when a dispatch of the change already carries the key under another session token with an outcome other than `aborted`. The same token replays to the same row as before.
**Considered:** the store's `RecordDispatch` — a refusal there needs new error kinds mapped in `internal/api/server.go` and `internal/reconcile/reconcile.go`, where the CLI already reads the run record for the round floor and an outage there never becomes a refusal; refusing on round `<r>` — **Panel re-runs** (`skills/flow/review-panel-fix-round.md`) defines `panel-fix-<r>` as the fix of review round `<r>-1`. Measured against history: since 2026-10-12, 34 of 97 panel-fix begins would have been refused, 30 with no round `<r>-1` row at all and 4 over a Minor-only round the parent fixes inline — the defect this refuses, not legitimate runs.

### Plan-shape refuses a visual-verify fix task without the record

**ID:** plan-shape-visual-fix-record
**Status:** active
**Chosen:** a new `check-plan-shape` rule, F21: a task titled `In-run fix <k> — visual-verify…` whose `**Files:**` holds no path ending `visual-verification.md` is refused at its title line.
**Considered:** a prose reminder in `skills/flow/verify-fix-loop.md` — the rule is already stated there, and kan-1024's Task 14 missed it.

### The verdict skips rows superseded within the round

**ID:** verdict-skips-superseded
**Status:** active
**Chosen:** `RenderPanel`'s per-round count skips a finding another finding of the same round names in `Supersedes`; kan-1024's round 1 then reads `5 Important`. Nothing is repaired: the panel record is rendered from the rows on demand and no committed copy holds the line.
**Considered:** skipping a row superseded by any later round — the earlier round did raise it and did need the fix, so its verdict was right.

### An abandoned stage ends at the next stage's start

**ID:** abandoned-ends-at-next-stage
**Status:** active
**Chosen:** `SweepAbandoned` sets `ended_at` to the earliest `started_at` of another stage run of the same change after this one's start, else `now()`; a new `store.ClipAbandonedStageRuns` applies the same bound to rows already `abandoned`, run once through a new `flowd clip-abandoned`, which reports how many rows it changed.
**Considered:** the last activity of the session — the store keeps no per-session activity stamp; leaving the written rows — a forward-only fix leaves the record wrong (**3. Fix every finding that is not `big`**, `skills/flow-self-review/SKILL.md`).

### A dispatch end before its start is clamped or refused

**ID:** dispatch-end-before-start
**Status:** active
**Chosen:** `EndDispatch` reads the row's `started_at` for every outcome; an `-ended-at` before it but inside the same second (`started_at` truncated to the second) is stored as `started_at`, and an earlier one is refused with `ErrDispatchEndsBeforeStart` (exit 2 at the CLI, in the refusal set of `internal/api/server.go` and `internal/reconcile/reconcile.go`). The renderer prints `ends before its start` instead of a negative duration. Row 8281 is retimed to end at its start with the existing `flowd retime-dispatch`; the other 19 rows keep their stored instants, flagged by the renderer, since nothing records when they really ended.
**Considered:** refusing the same-second case too — the CLI's `-ended-at` carries whole seconds while `begin` stores microseconds, so a fast abort would be refused for a truncation; rewriting the 19 rows — any instant chosen would be invented.

### The slot prompt states the round, and the guard refuses a cross-round path

**ID:** reproducer-round-stated
**Status:** active
**Chosen:** `render-slot-prompt` replaces "`R` this round's number" in the rendered REPRODUCER paragraph with that phrase followed by `` — `<round>` here`` (the round it renders), exiting 2 when the phrase is missing, as WORKTREES' replacement does; `check-panel-reproducers.sh` refuses, for every status, a runnable reproducer path under `reproducers/` that findings of two different rounds record, naming both refs. The existing open-only round check stays. kan-1024's F8 and F11 are rewritten through `flow record finding` (same ref) to `none — overwritten by round 1's script of the same name`.
**Considered:** extending the round check to closed rows — it would refuse 67 rows in 10 changes whose scripts nobody overwrote, where a path two rounds share is the collision itself; refusing any shared path — **A finding recorded under a joined slot** (`skills/flow/review-panel.md`) re-records one defect per raising role with the same reproducer, 50 same-round pairs in history; recreating F8's and F11's scripts from their reports — the defects are fixed and their scripts would demonstrate nothing now. A later panel round on kan-1024 would refuse the two Important exemptions; its panel is closed.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Tasks 2 and 6 change what `flow record dispatch` admits on every run, Task 5 changes the daemon's sweep, and Tasks 2, 5, 6 and 7 write the live store.
**Considered:** `flow-low` — most diffs are small, but a refusal wrong by one condition stops every `/flow` panel-fix begin.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group per repository, in series: `-fix-g1` Tasks 1–7 in the agents worktree (Tasks 2, 4 and 6 share `stats/` files and run in order), `-fix-g2` Tasks 8–9 in the gymie-playwright worktree (both need the worktree's running stack), `-fix-g3` Task 10 in the gymie worktree.
**Considered:** one fixer for all ten — one context carrying Go store work, Playwright traces under load and Gradle; one fixer per task — ten dispatches, three of them sharing files.

## Open questions

## Live check

Each step runs before landing, with every task committed. Store reads use
`docker exec flow-postgres psql -U flow -d flow -tc "<query>"`; flow commands run from
`<agents-wt>/stats` as `go run ./cmd/flow` after `make -s web-dist`, against the live daemon.

1. **Panel-fix refusal.** `go run ./cmd/flow record dispatch begin -C /Users/tweety53/Projects/gymie-worktrees/kan-1024-templates-warm-up-cool-down-any-workout -change kan-1024-templates-warm-up-cool-down-any-workout -role panel-fix -model opus -effort low -agent-id none -key panel-fix-9 -session-token mf-k1024-20261028T000000Z` — after: exit 2 naming round 8 and no row written (`select count(*) from dispatches where dispatch_key='panel-fix-9'` → 0). Not working looks like exit 0 and a new row; if a row is written, delete it by id.
2. **Sweep repair.** The overlap count — abandoned rows of a change whose `ended_at` lies past another stage run's later start — before: 30; after `go run ./cmd/flowd clip-abandoned`: 0, and the command prints 30. Not working looks like 30 after.
3. **Dispatch row 8281.** `select ended_at - started_at from dispatches where id = 8281` — before: negative; after: `00:00:00`. `select count(*) from dispatches where ended_at < started_at` — before 20, after 19.
4. **F8 and F11.** `select ref, reproducer from findings f join changes c on c.id=f.change_id where c.name='kan-1024-templates-warm-up-cool-down-any-workout' and ref in ('F8','F11')` — after: both `none — overwritten by round 1's script of the same name`.
5. **Rendered round.** `scripts/render-slot-prompt.sh` for round `0` over a scratch canonical worktree prints a REPRODUCER paragraph carrying `` — `0` here``.
6. **Panel verdict.** `go run ./cmd/flow record render -kind panel` for kan-1024 into a scratch file — round 1 reads `fix (0 Critical, 5 Important)`; before: `7 Important`.

## Recommended approach

Fix each defect where it starts: the merge-order writer reads the state record; the panel-fix
begin reads the run record the CLI already fetches for the round floor
(`panel-fix-begin-needs-findings`); plan-shape gains one rule (`plan-shape-visual-fix-record`);
the renderer's count and duration print read what is stored (`verdict-skips-superseded`,
`dispatch-end-before-start`); the sweep bounds its end (`abandoned-ends-at-next-stage`); the slot
prompt states the round and the guard catches a cross-round path (`reproducer-round-stated`).
Store rows a defect already wrote wrong are repaired by `flowd` commands and `flow record finding`
in the same task. The gymie tasks fix the specs and the launch environment in place.

## Critical files

- `stats/internal/guard/kickoffworktree.go`, `kickoff_worktree_test.go` — merge-order writer and its test.
- `stats/cmd/flow/record.go`, `record_test.go` — dispatch begin checks.
- `stats/internal/guard/planshape.go`, `check_plan_shape_test.go`, `scripts/check-plan-shape.sh` — plan-shape rule and header.
- `stats/internal/records/render.go`, `render_test.go` — panel verdict and dispatch duration.
- `stats/internal/store/stageruns.go`, `stageruns_test.go`, `stats/cmd/flowd/main.go` — sweep and repair command.
- `stats/internal/store/records.go`, `records_test.go`, `stats/internal/api/server.go`, `stats/internal/reconcile/reconcile.go` — dispatch end refusal.
- `stats/internal/guard/renderslotprompt.go`, `render_slot_prompt_test.go`, `panelreproducers.go`, `check_panel_reproducers_test.go`, `scripts/check-panel-reproducers.sh` — round in the prompt, cross-round path.
- gymie-playwright `fidelity/templates.spec.ts`, `tests/offline-floor-actions.spec.ts`, `fidelity/admin-templates.spec.ts`.
- gymie `buildSrc/src/main/kotlin/com/gymie/gradle/AdminFrontendEnvOutcome.kt`, `buildSrc/src/test/kotlin/com/gymie/gradle/AdminFrontendEnvTest.kt`, `gradle/dev-lifecycle.gradle.kts`.

## Existing code to reuse

- `kwPeerName`, `kwMergeOrder` (`kickoffworktree.go`) — peer resolution and append.
- The panel-round-floor block and `panelFixKeyRE` (`stats/cmd/flow/record.go`) — run-record read, backfill exemption, key shape.
- `TestRecordDispatchRefusesPanelFixZero` and the panel-round-floor test (`record_test.go`) — fake-server fixtures.
- `RetimeDispatch` / `flowd retime-dispatch`, and `RetokenStageRun` / `flowd retoken-stage-run` as the shape for `clip-abandoned`.
- The WORKTREES replacement in `renderslotprompt.go`.
- `cprRoundsAndSlots` (`panelreproducers.go`).
- gymie-playwright's `verify` row (gymie `.flow/project.md`, `## visual verification`) for spec runs against the worktree stack; `psqlAll`/`setTemplateState` in `fidelity/admin-templates.spec.ts`.

## Verification

Beyond each task's verify step: `scripts/run-guard-tests.sh` and `cd stats && make -s web-dist &&
go test -race -count=1 $(go list ./... | grep -v '/internal/guard$')` pass in `<agents-wt>`; the
agents `## lint` list passes; `fidelity/templates.spec.ts`, `tests/offline-floor-actions.spec.ts`
and `fidelity/admin-templates.spec.ts` pass under the gymie-playwright `verify` row with
`--repeat-each 5 --workers 4`; gymie's `./gradlew verifyChange` and `./gradlew -p buildSrc test`
pass; the **Live check** above.
