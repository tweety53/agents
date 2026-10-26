## Context

The self-review fix loop of kan-1017 (gymie), planned on branch
`self-review-kan-1017-templates-step-3-admin-templates-queue-create` of the agents repo, based at
`origin/main` `5e45b79a`. The pass is
`/Users/tweety53/Projects/gymie-worktrees/kan-1017-templates-step-3-admin-templates-queue-create/.superpowers/sdd/self-review-pass.md`;
the brainstorm trail is
`.superpowers/sdd/2026-10-26-self-review-kan-1017-templates-step-3-admin-templates-queue-create-design.md`
in this worktree. Each finding's code was located by read-only Explore dispatches against
`5e45b79a`, and the premises below were measured, not assumed.

Two findings reach outside the agents repo, placed per **Pipeline defects found mid-run**
(`skills/flow-contracts/pipeline.md`) on kan-1017's own change branches, where they ride that run's
route: SR-8's sidecar line in gymie-playwright
(`/Users/tweety53/Projects/gymie-playwright-worktrees/kan-1017-templates-step-3-admin-templates-queue-create`,
branch `spectre/kan-1017-templates-step-3-admin-templates-queue-create`) and SR-14's Redis fix in
gymie (`/Users/tweety53/Projects/gymie-worktrees/kan-1017-templates-step-3-admin-templates-queue-create`,
same branch name). SR-7's repair commits in that gymie worktree too.

**`flow record pass` is recorded many times inside one round.** Roster notes at round 0, diff-size
notes and bundle notes are all `flow record pass -round <n>` rows written before the round's
findings are recorded (`skills/flow/review-panel.md`, lines 141–350), so a refusal on `record pass`
would refuse those. The panel stage's own close is the one point the prose already requires the
guard before (**Before closing the stage**), and `stage end` already gates `flow.verify` in process
(`gateLiveVerification`, `stats/cmd/flow/stage.go`).

**The store refuses `-tokens` on a Claude Code session and an inline end under 3 s.**
`checkDispatchEndFields` (`stats/cmd/flow/record.go`) refuses `-tokens` where
`CLAUDE_CODE_SESSION_ID` is set and the store refuses it on a harvested session
(`ErrTokensOnHarvestedSession`); `minInlineWindow` is 3 s (`stats/internal/store/records.go`). A
skipped verify list closes in about one second, so neither `-tokens` nor a back-dated `-ended-at`
can close it. No outcome is `failed`: the closed set is `completed clean fix blocked aborted fallback
timed-out`, and `dispatches.outcome` is a plain `TEXT` column with no CHECK
(`stats/internal/store/migrations/0010_run_records.sql`).

**kan-1017's three verify rows, as stored.** `flow record dispatches -change kan-1017-…`:
<!-- measured: flow record dispatches -change kan-1017-templates-step-3-admin-templates-queue-create (gymie worktree) @ dev store, 2026-10-26 -->
row id 7820 (seq 47, `verify-test-fix-1`) began `2026-10-26T12:14:46.949903Z`, ended
`14:57:22Z`, outcome `completed`; id 7846 (seq 65, `verify-test-fix-2`) began
`13:53:37.04146Z`, ended `14:57:22Z`; id 7863 (seq 76, `verify-test-fix-3`) began
`15:00:01.436221Z`, ended `15:30:00Z` — a future instant when typed. All three carry the parent's
tokens for those inflated windows (cache read 743,948 / 12,902,437 / 9,424,497). The real ends
are about 12:16:04Z (a failed list), 13:53:38Z and 15:00:02Z (skipped lists), per the pass.
`flowd restamp-dispatch <id> <parent transcript>` recomputes an inline row's share from its window
but never moves the window; `flowd retime-dispatch <id> <started> <ended> <commit>` moves it
(`stats/cmd/flowd/main.go`); nothing rewrites an outcome.

**Ticked tasks with unticked steps are common in landed history, and the guard never reads it.**
`check-task-records` reads live plans and archived plans the base does not carry yet
(`stats/internal/guard/taskrecords.go`). Landed archives with ticked tasks over `[ ]` steps: agents
10 or more (75 in `self-review-kan-1015-…`), gymie 10 or more (27 in kan-1017, 29 in kan-1015).
<!-- measured: bash nullglob loop over spectre/changes/archive/*/tasks.md running awk that counts '  - [ ] ' lines under unfenced '- [x] N.' task lines @ agents 5e45b79a and gymie kan-1017 worktree b18bf74d -->
`flow tasks tick` refuses an already-ticked task (`ErrTaskAlreadyTicked`), so it cannot repair
kan-1017's plan.

**The dispatch row records no worktree** (`records.Dispatch`, `stats/internal/records/types.go`);
a visual verifier row is role `verifier` with a key starting `visual-verify`. **The store cannot see
a task close**: `close-task.sh` ticks `tasks.md`, and the `In-run fix <k> — ` title
(`uwInRunFix`, `stats/internal/guard/unfinishedwork.go`) is the only record of an in-run fix;
`dispatch begin` already reads the plan through `planPath`.

**The handoff source is a bundled page.** `handoff_templates/design/Gymie Templates Flow.html`
holds a `__bundler/manifest` JSON of base64 (gzip when `compressed`) assets and a
`__bundler/template` JSON string holding the page's HTML with its CSS inline. The one-liner in Task
13 decoded it to 21 files whose `index.html` carries 8 `border-radius` declarations.
<!-- measured: python3 -I one-liner of Task 13 on the kan-1017 gymie worktree's handoff page, then ls | wc -l and grep -c border-radius index.html @ gymie kan-1017 worktree b18bf74d -->

**SR-14's cause is unknown.** The shared per-JVM Testcontainers Redis
(`src/app/src/test/kotlin/com/gymie/AppIntegrationTest.kt`) has no `waitingFor`, registers
`spring.data.redis.password` as `""`, and `AuthIntegrationTestSupport.cleanup()` flushes over the
Spring `RedisTemplate` (Lettuce); none is confirmed.

## Decisions

### Gate the panel stage's close, not each pass row

**ID:** sr1-gate-stage-end
**Status:** active
**Chosen:** `flow stage end` for `-command '/flow' -stage flow.review-panel -outcome completed` runs `guard.Registry["check-panel-findings-closed"]` in process on `[<-C dir>, <change>]` and refuses the end (exit 1) on its exit 1 or 2, printing the guard's stderr — `gateLiveVerification`'s shape. The whole guard, not only the report-count half: the prose already requires it before this close.
**Considered:** refusing `flow record pass` when a round's reports out-count the store (the pass's first design) — pass rows are recorded many times inside a round, before its findings are, so it would refuse legitimate roster and diff-size notes. A hook on the Agent call — runs before the slots report.

### Name the open refs in the late-fix reason

**ID:** sr2-append-stderr
**Status:** active
**Chosen:** condition 1's reason appends the findings-closed guard's stderr, trimmed, after `(check-panel-findings-closed.sh exited 1)` as `: <lines joined by "; ">`.
**Considered:** printing the stderr separately — the reason line is what the parent reads.

### A skipped verify list ends `skipped`; a failed one ends `blocked`

**ID:** sr4-skipped-outcome
**Status:** active
**Chosen:** `run-verify-list` ends the row it began on every return path after `begin`: `-outcome completed` when every list ran green, `-outcome skipped` when every entry was skipped on unchanged inputs, `-outcome blocked -cause test-failure` when any failed. `skipped` joins `recordOutcomes`; the store's inline floor (`minInlineWindow`) admits an inline end whose outcome is `skipped` — a list that ran nothing has no work to measure. No `-tokens` is passed.
**Considered:** `-tokens input=0,output=0` (the pass's first design) — refused on every Claude Code session. Sleeping to 3 s before `end` — pads a measurement. Recording no row for a skipped list — loses the record that verify ran. Exempting every non-`completed` inline end — admits the hand-closed-together case the floor exists for.

### Repair kan-1017's three verify rows

**ID:** sr4-repair-rows
**Status:** active
**Chosen:** a new `flowd reoutcome-dispatch <dispatch-id> <outcome> [<cause>]` (store `SetDispatchOutcome`, outcome from the same closed set, cause only with `blocked`), then per row: `flowd retime-dispatch` to its real window, `flowd reoutcome-dispatch` to its real outcome, `flowd restamp-dispatch <id> <parent transcript>` to recompute its token share — 7820 `12:16:04Z` `blocked test-failure`, 7846 `13:53:38Z` `skipped`, 7863 `15:00:02Z` `skipped`. A parent step after Task 3's commit, run from this worktree's build.
**Considered:** `flowd restamp-dispatch` alone (the dispatch prompt's wording) — it never moves a window or an outcome, so the rows would keep wrong ends. Leaving 47's outcome `completed` — the list failed.

### Refuse a verifier beside a mutator change-wide

**ID:** sr5-change-wide
**Status:** active
**Chosen:** `flow record dispatch begin` reads the run record (`GetRunRecord`, the panel-round-floor pattern; a read failure skips the check) and refuses (exit 2) a `visual-verify*` key while a row of role `implementer`, `panel-fix` or `red-partner` has no `EndedAt`, and such a role while an open `visual-verify*` row exists, naming the open row's key. Change-wide: a parallel mutator in another repository's worktree waits for the verifier too.
**Considered:** a `-worktree` flag and column — a migration and every begin call site for a precision the one-mutator rule does not need in practice. Matching repository key suffixes — a naming convention, not a record.

### Fold the porcelain check into the tree-markers guard

**ID:** sr6-porcelain-in-tree-markers
**Status:** active
**Chosen:** `check-tree-markers.sh snapshot` also writes `git -C <worktree> status --porcelain` (paths under `.superpowers/` dropped) into the snapshot file after a `#porcelain` line; `verify` recomputes it and exits 1 naming each path present now and absent from the snapshot. A snapshot without the `#porcelain` section verifies as before.
**Considered:** a check on `flow record dispatch end` for role `reviewer` (the pass's first design) — the end has no pre-dispatch snapshot; the tree-markers guard already brackets every round and every per-task review per worktree.

### Refuse ticked tasks with unticked steps where the guard already reads

**ID:** sr7-guard-and-repair
**Status:** active
**Chosen:** `check-task-records` adds a violation for a `- [x] <n>.` task whose body holds an unfenced `  - [ ] ` line, on the plans it already reads (live and unlanded archives). kan-1017's 27 step lines are ticked by a line-bounded `sed` on lines 231–322 of its archived `tasks.md` — a parent step committed on the gymie change branch.
**Considered:** reading landed archives too — fails every lint run on history no run can change. A `flow tasks tick` repair mode — a second command for one repair.

### A sidecar line overrides the frame geometry; the project row stays

**ID:** sr8-sidecar-frame
**Status:** active
**Chosen:** the first non-comment line of a `<spec>.mockups` sidecar may be `frame <geometry>` (parsed by `cmfParseGeometry`, malformed exit 2) or `frame none` (native size, uncropped), overriding the fourth argument for that sidecar; a `frame` line anywhere else is a map-line finding (exit 1). kan-1017's `fidelity/admin-templates.spec.ts.mockups` gains `frame none`; gymie's `.flow/project.md` `mockup frame` row stays `scale=2 status=26 border=1`, right for the phone frames.
**Considered:** changing gymie's project row — breaks every phone frame. A per-frame column — the admin sidecar is all desktop frames.

### One matrix call, regions clipped to the image

**ID:** sr11-matrix-and-clip
**Status:** active
**Chosen:** `--props matrix` (opt-in, page-level) runs `measureElements` on both images, pairs the elements and measures each pair's own region with its kind's props (text `ink`; control/row/card `box,radius,border,fill,shadow,content,gap`; rule `runs`), one row per pair. `mvpOpen` clamps a region that starts inside the image to the image's edge and marks it clipped; one starting outside still exits 2.
**Considered:** clipping only for `matrix` and `bands` — the band at y=795 h=7 on an 800 px image is a plain region read.

### Band colour is the darkest ink plus the top-two share

**ID:** sr13-darkest-ink
**Status:** active
**Chosen:** `mvpCounter.darkest()` — lowest luminance among colours holding at least 2% of the band's pixels — is the band's `colour`; a new `share` field carries `(n1+n2)/total` of the top two colours. `edge` is unchanged.
**Considered:** reporting both top colours — still a flip in the first slot.

### Name the sources in the prompt; refuse `no mockup source` when one exists

**ID:** sr12-design-flag
**Status:** active
**Chosen:** `check-verify-report.sh` takes an optional `--design <design.md>`: it collects the backticked `.html`, `.htm` and `.css` paths `design.md` cites that exist relative to the worktree (three directories above the report) and, when any exists and the report or `visual-verification.md` beside `design.md` contains `no mockup source`, reports step 10 undone (exit 1). `visual-verify.md` names `mockup sources:` in the prompt and documents the bundled-page decode one-liner.
**Considered:** a `scripts/` decode helper (the pass's first design) — one consumer; the documented one-liner is the same code with no harness to keep. Taking `<changeRoot>` positionally — breaks the existing three-positional contract.

### kickoff `--from` writes the merge-order record

**ID:** sr15-kickoff-writes-merge-order
**Status:** active
**Chosen:** after a successful `--from` cut, when `<from-worktree>/spectre/changes/<name>/` exists, `kickoff-worktree.sh` writes `link.md` there with `## Merge order` `1. \`.\`` then the new repository's name, or appends the name as the next numbered entry when the file has the section and lacks it; the name is the new repository's entry in `<from-worktree>/spectre/peers`, a missing entry exit 2 before any write. It never commits; it prints the path written.
**Considered:** leaving the record to the agent — the hand-written record failed `spectre validate` twice in kan-1017.

### A verb that widens one task's Files

**ID:** sr16-files-verb
**Status:** active
**Chosen:** `flow tasks files [-C <dir>] <change> <id> add <path>…` appends each path not already backticked in task `<id>`'s column-0 `**Files:**` field as `, \`<path>\`` on the field's last line (`**Files:** none…` is replaced by the paths); every other byte unchanged. The task-body scan of `tickTask` moves to one helper both use.
**Considered:** a guard catching a sed that hit other tasks — refuses after the damage.

### Read ticked in-run fixes from the plan at begin

**ID:** sr17-tasks-md-read
**Status:** active
**Chosen:** for a `visual-verify*` key, `dispatch begin` reads the plan; `k` is the highest `<k>` of a ticked `In-run fix <k> — ` task with no completed row whose key carries `-fix-<k>-full`. With such a `k`, it refuses (exit 2) a key without `-fix-<k>`, and a non-`-full` key without `-round-base` when a completed `visual-verify*` row's key ends with this key's text after `-fix-<k>` (the repository suffix; empty matches every row).
**Considered:** a store-only signal (an implementer `-fix-` row ending after the newest verifier) — misses inline fixes. Requiring `-round-base` on every re-run — refuses the unscoped run of a worktree with no earlier verifier, which the loop allows.

### Bisect the Redis flake before any fix

**ID:** sr14-bisect-first
**Status:** active
**Chosen:** Task 15 follows `docs/flake-bisection.md`: the failing class alone, then `--tests` pairs in the failing run's class order, repeated runs; then a fix at the source and a proof by removing it. When bounded runs do not reproduce, the task commits nothing and reports the runs as evidence; no timeout, retry or sleep is widened.
**Considered:** adding `waitingFor` or dropping the empty password up front — a guess with no failing run behind it.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Tasks 3, 4 and 5 change what the store admits and what every dispatch begin refuses, and Task 1 gates every panel close.
**Considered:** `flow-low` — most tasks are small guard additions, but a wrong begin refusal blocks every run.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** three groups, in series. `fix-g1`: Tasks 1–7 (the `flow` CLI, store and task-record guards; Tasks 3–6 share `stats/cmd/flow/record.go` or `skills/flow/implement.md` and run in order). `fix-g2`: Tasks 8–14 (tree-markers and the visual and kickoff tooling, plus the gymie-playwright sidecar commit). `fix-g3`: Task 15 (gymie, Gradle, its own worktree).
**Considered:** one fixer — fifteen tasks across three repositories in one context. One per task — fifteen dispatches.

## Open questions

## Recommended approach

Each finding gets the smallest mechanical refusal or tool at the point the run already passes
through: an in-process gate on the panel stage's close (`sr1-gate-stage-end`, not the pass row), a
`skipped` outcome and a reoutcome repair command (`sr4-skipped-outcome`, `sr4-repair-rows`, not
`-tokens`), change-wide begin refusals (`sr5-change-wide`, `sr17-tasks-md-read`, not a worktree
column), a porcelain section in the existing tree-markers snapshot (`sr6-porcelain-in-tree-markers`),
one more violation in `check-task-records` (`sr7-guard-and-repair`), a sidecar `frame` line
(`sr8-sidecar-frame`), a one-call matrix and edge clipping (`sr11-matrix-and-clip`), a darkest-ink band
colour (`sr13-darkest-ink`), a `--design` refusal plus a documented decode (`sr12-design-flag`, not a
script), kickoff writing the merge order (`sr15-kickoff-writes-merge-order`), a `flow tasks files`
verb (`sr16-files-verb`), and a bisection-first flake fix (`sr14-bisect-first`). Every past record the
defects wrote wrong is repaired: three dispatch rows and kan-1017's archived steps.

## Critical files

- `stats/cmd/flow/stage.go`, `stats/cmd/flow/stage_test.go` — panel close gate (Task 1).
- `stats/internal/guard/latefixtrigger.go`, `stats/internal/guard/check_late_fix_trigger_test.go` — condition-1 reason (Task 2).
- `stats/internal/guard/runverifylist.go`, `stats/internal/guard/run_verify_list_test.go`, `stats/cmd/flow/record.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `stats/cmd/flowd/main.go`, `scripts/run-verify-list.sh` — verify rows and the repair command (Task 3).
- `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go` — begin refusals (Tasks 4, 5).
- `stats/cmd/flow/tasks.go`, `stats/cmd/flow/tick.go`, `stats/cmd/flow/tasks_test.go` — `flow tasks files` (Task 6).
- `stats/internal/guard/taskrecords.go`, `stats/internal/guard/check_task_records_test.go` — ticked-step check (Task 7).
- `stats/internal/guard/treemarkers.go`, `stats/internal/guard/check_tree_markers_test.go`, `scripts/check-tree-markers.sh` — porcelain (Task 8).
- `stats/internal/guard/visualpreflight.go`, `stats/internal/guard/check_visual_preflight_test.go` — siblings (Task 9).
- `stats/internal/guard/composemockupframes.go`, `stats/internal/guard/compose_mockup_frames_test.go`, `scripts/compose-mockup-frames.sh` — sidecar frame (Task 10).
- `stats/internal/guard/measurevisualproperties.go`, `stats/internal/guard/measure_visual_properties_test.go`, `scripts/measure-visual-properties.sh` — matrix, clip, band colour (Tasks 11, 12).
- `stats/internal/guard/verifyreport.go`, `stats/internal/guard/check_verify_report_test.go`, `scripts/check-verify-report.sh` — `--design` (Task 13).
- `stats/internal/guard/kickoffworktree.go`, `stats/internal/guard/kickoff_worktree_test.go`, `scripts/kickoff-worktree.sh` — merge order (Task 14).
- `skills/flow/review-panel.md`, `skills/flow/implement.md`, `skills/flow/verify-fix-loop.md`, `skills/flow/visual-verify.md`, `skills/flow/visual-verify-verifier.md`, `skills/flow-contracts/project-configuration-visual.md`, `skills/flow/cross-repo-worktrees.md` — one added sentence each per task, listed in `verbatim-moves.txt`.
- gymie-playwright `fidelity/admin-templates.spec.ts.mockups` (Task 10); gymie `src/app/src/test/kotlin/com/gymie/AppIntegrationTest.kt` and its support files (Task 15); gymie `spectre/changes/archive/kan-1017-templates-step-3-admin-templates-queue-create/tasks.md` (Task 7's repair).

## Existing code to reuse

- `gateLiveVerification` (`stats/cmd/flow/stage.go`) and `TestStageEndFlowVerifyRunsLiveCheck` (`stats/cmd/flow/stage_test.go`) — the in-process stage gate and its test.
- `checkPanelFindingsClosed` via `guard.Registry["check-panel-findings-closed"]` (`stats/internal/guard/panelfindingsclosed.go`).
- `lftNewFx`, `lftEnv` and the `open` fixture (`stats/internal/guard/check_late_fix_trigger_test.go`).
- The `record` closure (`stats/internal/guard/runverifylist.go`), `TestRunVerifyListRecordsTheVerifyRow`'s argv-logging fake `flow` (`run_verify_list_test.go`).
- `RetimeDispatch`, `RekeyDispatch` (`stats/internal/store/records.go`) and `retimeDispatch` (`stats/cmd/flowd/main.go`) — the shape of `SetDispatchOutcome` and `reoutcome-dispatch`.
- `runRecordDispatchBegin`'s panel-round-floor read of `GetRunRecord`, `recordRepo`, `genuineDaemon`, `TestRecordDispatchBeginRefusesAVisualVerifyVerifierWithoutDiffBase` (`stats/cmd/flow/record_test.go`, `state_test.go`); `planPath`, `taskLinePattern`, `fencePattern` (`stats/cmd/flow/tasks.go`, `tick.go`).
- `tickTask`'s task-body scan (`stats/cmd/flow/tick.go`).
- `tcfSelectTask`, `tcfUnfenced`, `tcfTaskLineRE` (`stats/internal/guard/planparse.go`); `trRepo`, `fx.task` (`check_task_records_test.go`).
- `checkTreeMarkers` snapshot/verify (`stats/internal/guard/treemarkers.go`).
- `tcfAppRoots`, `projectSection`, `vpRepoRoot`, `vpTree`, `vpRun` (`stats/internal/guard/`).
- `cmfParseGeometry`, `cmfRun`, `cmfSize` (`stats/internal/guard/composemockupframes.go`, its test).
- `measureElements`, `measureBands`, `mvpOpen`, `mvpCounter`, `mvpPNG`, `mvpRun`, `mvpJSON` (`stats/internal/guard/measurevisualproperties.go`, its test).
- `checkVerifyReport` (`stats/internal/guard/verifyreport.go`), `TestCheckVerifyReportFingerprintReason`'s temp-file setup.
- `kickoffWorktree`'s `--from` path and its subtests (`stats/internal/guard/kickoffworktree.go`, `kickoff_worktree_test.go`).

## Verification

- `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean; `cd stats && go test ./cmd/flow/ ./cmd/flowd/ ./internal/guard/ ./internal/store/ -count=1`.
- `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` exit 0.
- `scripts/check-task-records.sh /Users/tweety53/Projects/gymie-worktrees/kan-1017-templates-step-3-admin-templates-queue-create` exits 0 after Task 7's repair and 1 at the gymie branch's commit before it.
- `flow record dispatches -change kan-1017-templates-step-3-admin-templates-queue-create` shows rows 7820, 7846, 7863 with their real ends and outcomes after Task 3's repair.
- `compose-mockup-frames.sh` over kan-1017's admin sidecar composes every AD frame at native size (exit 0).
- Task 15's reproducing `--tests` pair passes repeatedly with the fix and fails with it reverted.

## Live check

Two runtime surfaces, each run at `flow.verify` on this worktree's build, never restarting the live
`flowd`:

- **The repaired ledger rows (Task 3).** Before: `flow record dispatches -change kan-1017-templates-step-3-admin-templates-queue-create` shows 7820, 7846, 7863 ending `14:57:22Z`, `14:57:22Z`, `15:30:00Z`, all `completed`, with millions of cache-read tokens. After: ends `12:16:04Z`, `13:53:38Z`, `15:00:02Z`; outcomes `blocked` (cause `test-failure`), `skipped`, `skipped`; tokens recomputed by `restamp-dispatch` from the real windows. Not working looks like any of the old ends or outcomes still present.
- **The verify-list end (Task 3) against a scratch store.** `go test ./internal/store/ -run 'SkippedInline' -count=1` creates and drops its own database in `flow-postgres`; a scratch `run-verify-list` with unchanged inputs ends its row `skipped` against that test database. Not working looks like `inline dispatch window too short` or an open row.
