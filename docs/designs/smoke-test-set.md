# smoke-test-set
> Design record of change smoke-test-set, written at landing. Each decision's own **Status:** line is the status.

## Context

`flow-guard run-verify-list <worktree> test` runs the four suites of `.flow/project.md` `## test`
at the end of `flow.sdd-tdd` and again at `flow.verify`. The two Go suites dominate it: on the
afternoon of 2026-11-13 the four `guard-tests` runs took 7m10s to 7m56s and the four `stats-go`
runs 3m21s to 4m19s, whatever the change touched (`flow suite list -suite guard-tests`,
`-suite stats-go`; their medians over the last ten passing runs are 6m0s and 3m21s). The operator
asked for a fast set that covers the repository's core functionality and runs most of the time,
with the full suite run only on ask, and then fixed two points: the fast set is every light test
plus a core set that agents gather by reading the suites, and no full run starts unattended.

Nothing a user sees changes, so the design has no prototype.

Three findings shape it.

**The light rule cannot be read off a test's own `Elapsed`.** `go test -json` stops a test's clock
while it waits on `t.Parallel()` and leaves a parent's parallel subtests out of the parent's
figure. `stats/internal/guard` holds 937 `t.Parallel()` calls in 119 of its 141 test files, and 401
of its 560 top-level tests are parallel at top level. A test's **cost** is therefore its own
`Elapsed` plus the `Elapsed` of every descendant that paused. Measured on that package, one run
after the other on a loaded machine: skipping the 69 whose own `Elapsed` is 1 s or more ran
305 s and used 756 CPU-seconds; skipping the 185 whose cost is 1 s or more ran 80 s and used 95.
Those 185 carry 1385 of the package's 1478 cost-seconds.
<!-- measured: cd stats && go test -count=1 -json ./internal/guard/ (profile), then go test -count=1 -skip '<the 69 names>' and -skip '<the 185 names>' ./internal/guard/ @ 99ca4f38 on 2026-11-13; machine-local timings, not re-runnable to the same figures -->

**A suite its guard skipped is recorded as a run.** `run-verify-list` wraps the whole
`flow-guard unchanged … || { … }` line in `flow suite record`
(`stats/internal/guard/runverifylist.go`, the `list == "test"` branch that builds `argvs`), so a
skipped suite lands in the store as a passing run of about 0.2 s and enters its median. The
medians are the signal this design relies on to notice the default list slowing down, so the
defect is fixed here. In the store today `guard-tests` and `stats-go` each hold one such row from
2026-11-13, and `stats-spa`'s median of 8.1 s is taken over ten rows of which four are such skips,
of 141, 173, 214 and 230 ms; its real runs took 8 to 30 s.
<!-- measured: flow suite list -suite stats-spa, -suite guard-tests, -suite stats-go @ the dev store on 2026-11-13; a live store read, not re-runnable to the same rows -->

**The heavy tests of two packages read the repository outside `stats/`.** The fixture the tests of
`stats/internal/setuptest` run on holds a copy of the real `scripts/flow-guard`
(`stats/internal/setuptest/main_test.go`, `makeFixtureRepo`), and six heavy tests run the real
`setup.sh` over the real tree — `TestSetupAgents`, `TestKotlinRuleInstalledByNoMode`,
`TestGlobalInstallPopulatesSkillDirs`, `TestPreexistingSkillDirMovedOut`,
`TestRegisteredHookTimeoutIsChecked`, `TestRealKotlinStandardReachesProject` — which reads
`skills/`, `rules/`, `agents/`, `mods/`, `commands-claude/` and, in project modes, `templates/`.
In `stats/internal/repotest`, `TestFlowTaskListLanded` runs `mods/flow-task-list/hooks/landed.ts`
and `TestGoSourcesHashSeesGoEdits` reads `stats/Makefile` and `.flow/project.md`; every other
heavy test there runs a `hooks/*.py` file, `scripts/flow-guard`, or both. No name pairing sees any
of these inputs.

## Recommended approach

One Go-native switch, `-short`, on the lines that exist. A heavy test opens with one call that
skips it in a `-short` run; a core test is one left running, tagged with one comment. The diff
tells the run which heavy tests to keep. The full suite is the same list with every skip turned
off. Nothing is stored that the tree does not already state: no list of test names, no map from
files to tests, no flag in the state record.

### The default list

`.flow/project.md` `## test` keeps its four suites, their names, their `flow-guard unchanged`
guards, their `# inputs:` lines and their share of the cores. The two Go lines gain `-short` and a
keep list; the other two are today's, unchanged:

```bash unverified:run both Go lines once flow-guard smoke-keep and smoke.Heavy exist — neither does at 99ca4f38; the stats-spa and flow-task-list lines are copied from .flow/project.md as it stands
# suite: guard-tests
scripts/flow-guard unchanged . ':!docs/' ':!spectre/' ':!stats/web/' || { keep=$(scripts/flow-guard smoke-keep stats/go.mod=stats/ stats/go.sum=stats/ hooks/=stats/internal/repotest/ hooks/=stats/internal/setuptest/ scripts/=stats/internal/repotest/ scripts/=stats/internal/setuptest/ setup.sh=stats/internal/setuptest/) && cd stats && FLOW_SMOKE_KEEP="$keep" GOMAXPROCS=$(( ($(getconf _NPROCESSORS_ONLN) + 2) / 3 )) go test -race -short -count=1 ./internal/guard/...; }
# suite: stats-go
scripts/flow-guard unchanged . ':!docs/' ':!spectre/' ':!stats/web/' || { keep=$(scripts/flow-guard smoke-keep stats/go.mod=stats/ stats/go.sum=stats/ hooks/=stats/internal/repotest/ hooks/=stats/internal/setuptest/ scripts/=stats/internal/repotest/ scripts/=stats/internal/setuptest/ setup.sh=stats/internal/setuptest/) && cd stats && make -s web-dist && FLOW_SMOKE_KEEP="$keep" GOMAXPROCS=$(( ($(getconf _NPROCESSORS_ONLN) + 2) / 3 )) go test -race -short -count=1 $(go list ./... | grep -v '/internal/guard$'); }
# suite: stats-spa
# inputs: stats/web/ stats/Makefile
scripts/flow-guard unchanged stats/web/ stats/Makefile || { cd stats && make -s web-dist && cd web && npm test; }
# suite: flow-task-list
# inputs: mods/flow-task-list/
scripts/flow-guard unchanged mods/flow-task-list/ || claude plugin test mods/flow-task-list
```

The SPA unit tests and the mod's tests run whole when their inputs changed, as today: they are
light already. The visual `verify` row is untouched; this repository declares no `specs` row.

### Heavy

A new package, `stats/internal/smoke`, holds one function. `smoke.Heavy(t)`, the first statement of
a top-level test, skips it when the run is `-short` — unless the run selects tests itself with
`-run` (`flag.Lookup("test.run")` is non-empty), `FLOW_VERIFY_FULL` is set, or `FLOW_SMOKE_KEEP`
names the test's file or a directory above it (the file read from `runtime.Caller(1)`). The four
cases were spiked in a scratch module during planning. `FLOW_SMOKE_KEEP` holds repository-relative
paths separated by spaces.

A test is heavy when its cost is 1 s or more in `go test -race -count=1 -json` of its package —
the default line's own flags less `-short` — on a machine doing nothing else. A new guard,
`flow-guard test-cost [-min <seconds>]`, reads that JSON on stdin and prints one line per
top-level test whose cost is at least `-min` (default 0), heaviest first: cost in seconds to two
decimals, the package as the JSON names it, and the test, separated by tabs. A test the run
skipped is not printed, and a stdin line that is not a JSON event is ignored. It exits 0 on a
printed table, empty included, and 2 on usage, and never judges a duration.

By the profile in hand the heavy tail is 185 guard tests and at most 175 in the other packages;
the second figure was taken under load, without `-race`, and an idle profile will mark fewer.
Every one gets the call except the core set below. The names marked are the ones the parent's idle
profile prints (decision `marks-come-from-a-parent-profile`), never this paragraph's counts.

### Core

The core set is recorded where the tests are. A core test carries
`// smoke: core — <the capability it proves>` as the last line of the comment directly above its
`func`, and never calls `smoke.Heavy`.
`git grep -n '^// smoke: core' -- stats ':!stats/internal/smoke'` lists the set with its reasons;
the package's own sources name the tag too, so they are left out of the listing. A static test in
`stats/internal/smoke` walks every `_test.go` under `stats/` with `go/parser` and fails when a function carries both the tag and the call, when `smoke.Heavy` is
called anywhere but as the first statement of a top-level test, or when `testing.Short()` is
called outside the package — so the tag cannot be contradicted, the keep list always sees the
test's own file, and the one skip switch stays one. A rename or a deletion carries the tag with
it; there is no list to go stale.

Heavy tests left running — the ones the tag changes the run for. Cost is in cost-seconds from the
planning profiles; the guard rows were profiled on a quiet machine, the others under load.

| Package | Test | Cost | What it proves |
|---|---|---|---|
| `internal/guard` | `TestKickoffWorktree` | 13.4 | kickoff: worktree, branch, setup, push, the `STARTED` record |
| `internal/guard` | `TestCloseTaskOrder` | 12.0 | a task's close; runs `check-plan-shape`, `check-task-commit-fields`, `check-task-commit-planning-paths`, `check-untasked-commits` and `check-review-gate` in process (`stats/internal/guard/closetask.go`) |
| `internal/guard` | `TestGatherDispatchContext` | 8.9 | the bundle every dispatch reads |
| `internal/guard` | `TestFixRunFloorVerdict` | 7.9 | the one gate that lets a run skip its tests |
| `internal/guard` | `TestSyncOntoBase` | 5.5 | the rebase onto a moved base |
| `internal/guard` | `TestIntegrateChangeHappyPath` | 5.2 | prepare, commit, land and cleanup over two repositories with only `flow` stubbed, so `check-finish-preflight`, `reshape-branch`, `commit-split` and `remove-change-worktrees` run for real |
| `internal/guard` | `TestPlanClass` | 3.8 | the class that sets the roster and the modes |
| `internal/guard` | `TestRunVerifyListVerifyFast` | 2.7 | the only test that `verify: fast` is enforced |
| `internal/guard` | `TestCheckPanelFindingsClosed` | 2.1 | a gate whose failure is silent |
| `internal/guard` | `TestFoldFixupFolds` | 1.8 | the fold's success path |
| `internal/guard` | `TestCheckReviewGate` | 1.7 | a gate whose failure is silent |
| `internal/guard` | `TestPanelRoundBeginsEnforcesTheJudgedRoster` | 1.6 | the only test that a judged roster is enforced |
| `cmd/flow` | `TestChangeSaveUpsertsBothFiles` | 3.0 | the CLI through a real `api.New` to a real store |
| `internal/harvest` | `TestMarkedStageBindsToRealSessionEndToEnd` | 2.0 | transcript to harvest to attribution to store |
| `internal/setuptest` | `TestGlobalInstallPopulatesSkillDirs` | 6.1 | the real `setup.sh global` over the real `skills/`, `mods/`, `rules/`, `agents/`, `commands-claude/` and `hooks/`: every skill and mod installs as a resolvable link, none dangles, the managed block renders |
| `internal/setuptest` | `TestCoreExcerptsLinksAndBaseline` | 2.2 | the lightest run of the real `setup.sh`, over a fixture tree |
| `internal/repotest` | `TestEnforceAgentBaselineHook` | 19.5 | the hook that gates every dispatch, through the real guard and CLI |
| `internal/repotest` | `TestFlowStopGuardHook` | 3.0 | the stop gate |
| `internal/repotest` | `TestFlowTaskListLanded` | 1.4 | the mod's landed check against real git — the only Go test that runs a file under `mods/` |
| `internal/repotest` | `TestGoSourcesHashSeesGoEdits` | 1.3 | the UI-test stack's fingerprint sees a Go edit; reads `stats/Makefile` and `.flow/project.md` |
| `cmd/uitest-seed` | `TestSeedPopulatesFixture` | 11.6 | the seeder a store change can break, otherwise seen only at the next UI change |
| `internal/sweep` | `TestConcurrentSweepersDoNotDoubleCloseOrCollide` | 1.7 | a data-race test |

The twelve guard rows sum to 66.6 cost-seconds — 6.2 sequential and 60.3 parallel, about 21 s of
wall at verify's four threads. A row the idle profile puts under 1 s keeps its tag and needs no
exception.

Light tests tagged so that no later marking pass removes them from the default run:

- `internal/guard`: `TestPlanShapeDesignSections`, `TestPlanDispatchBundles`,
  `TestPlanDispatchGroups`, `TestCheckTaskCommitFieldsFindsAClassDotMethodTest`,
  `TestCheckDispatchPairExitCodes`, `TestPlanAgentPairFollowsPolicy`,
  `TestRunVerifyListRecordsTheVerifyRow`, `TestRunVerifyListSkipsACommandGreenAtHead`,
  `TestCheckVerifyGreen`, `TestVisualVerifyPassPrepareBundle`, `TestVisualVerifyPassCloseDone`,
  `TestRemoveChangeWorktreesRefusesBeforeTheSave`, `TestPipelineFixStart`.
- `cmd/flow`: `TestStateSetSucceedsAgainstReachableStore`,
  `TestStateGetSucceedsAgainstReachableStore`,
  `TestStateSetFallbackWritesStateFileAndJournalEntry`, `TestStageBeginRecordsIdentityAndInstant`,
  `TestStageEndRecordsOutcomeAndMetrics`,
  `TestRecordDispatchBeginPrintsTheAllocatedSeqAndEndClosesIt`,
  `TestRecordWriteFallsBackToJournalAndExitsZero`, `TestRecordVerdictWritesAndFallsBackToJournal`,
  `TestRecordFindingTaskRoundTrips`, `TestDecisionRenderGolden`,
  `TestSuiteRecordRunsChildAndRecordsRow`, `TestSuiteListRendersRowsAndSummary`,
  `TestPolicyGetCachesAndFallsBack`, `TestTasksCountRecordsThePlanTaskCount`,
  `TestJournalFlushCommandReplaysOnDemand`.
- `internal/api`: `TestRecordDispatchRouteAllocatesSeqAndAnswers201`,
  `TestRecordFindingRouteAnswers201OnCreateAnd200OnUpdate`, `TestEveryViewAcceptsAPeriod`.
- `internal/store`: `TestMigrationsAreIdempotent`, `TestPutChangeRoundTripsEveryField`,
  `TestRunRecordOrdersDispatchesAndFindings`.
- `internal/harvest`: `TestWatcherHarvestsRolloutSource`,
  `TestRolloutSubagentBatchesCreditTheirOwnDispatchUnderConcurrency`. `internal/reconcile`:
  `TestReplayAppliesPendingRecordEntries`, `TestReconcilerRunSerializesConcurrentCalls`.
  `internal/fallback`: `TestAppendJournalEntryIsSafeUnderConcurrentWriters`. `cmd/flowd`:
  `TestDaemonWiresTheRealStore`.

Left out on purpose: `TestCheckTaskCommitFields` (66 cost-seconds), `TestCheckFinishPreflight`,
`TestRemoveChangeWorktrees`, `TestReshapeBranch` and `TestCommitSplit` run for real inside
`TestCloseTaskOrder` and `TestIntegrateChangeHappyPath`; `TestCheckStageMarkCalls`,
`TestCheckWorkspaceIsolation`, `TestCheckPlanProvenance` and `TestCheckSpecReach` test guards the
`## lint` list runs over the real tree at every verify; `TestWorkspace` costs 8 s in series and its
failure is loud the next time a run isolates its workspace; `TestEnforceLoopDispatchHook` and the
`TestProtectMainCheckout` family run whole whenever `hooks/` or `scripts/` changes; the other five
real-tree `setup.sh` tests assert narrower outcomes of the install
`TestGlobalInstallPopulatesSkillDirs` already runs.

**The rubric an agent gathers by**, now and at every later pass: keep a heavy test only when it is
the cheapest success-path proof — through the registered command, the CLI's `run` or a real script
— of something every run executes, of a gate whose failure is silent, or of an operation that
rewrites or destroys; when no lighter test covers that path and `## lint` does not already run
that guard over the real tree; and when no other core test already runs it for real. A heavy test
that is the only reader of a repository input nearly every change touches is kept under the same
rule, in place of a keep row on that input.

**How the set stays current.** Three triggers, none of them a duration check:

- a new test is unmarked and untagged, so it runs — the safe side;
- re-profiling is one command, the default Go line with `-json` piped into
  `flow-guard test-cost -min 1`: marked tests are skipped and so not printed, and what prints is
  the core set plus whatever grew slow — each new row meets the rubric and is tagged, or is marked;
- a full run that fails where the default list passed at the same HEAD prints the escape line of
  **The full suite**, and the fix tags that test, or a cheaper sibling of the same behaviour.

The drift signal is `flow suite list`: with the recording defect fixed, the `guard-tests` and
`stats-go` medians are default-run times and nothing else.

### What a touched file adds

`flow-guard smoke-keep [<path prefix>=<kept path>]...` reads the diff `flow-guard unchanged` reads
— HEAD and the working tree against HEAD's merge-base on `origin/main`, untracked files included —
and prints one line of repository-relative paths separated by spaces:

- a changed `_test.go` file: itself;
- a changed Go source file: every test file in its directory whose name pairs with it — both
  names lowercased and stripped of `_`, the test side also stripped of its `test` suffix and of a
  leading `check`, and one a prefix of the other (`runverifylist.go` and
  `run_verify_list_test.go`; `integratechange_land.go` and `integrate_change_test.go`). 135 of the
  154 guard source files pair this way, and 41 of them pair only with heavy tests — without the
  keep list a change to one of those would run none of its own tests;
- a changed path under an argument's prefix: that argument's kept path, a directory or one file.
  The arguments on the lines above say: a `stats/go.mod` or `stats/go.sum` change keeps everything
  under `stats/`; a `hooks/` or `scripts/` change keeps the whole of `internal/repotest` and
  `internal/setuptest`; a `setup.sh` change keeps `internal/setuptest`;
- a changed Go source file that pairs with nothing — the 19 shared helpers such as `exec.go` and
  `paths.go`: nothing is kept, and one stderr line names the file, so the suite's block in the
  `## Report` shows the gap.

It exits 0 with the line printed, possibly empty, and 2 on a usage or git failure, which fails the
suite through the line's `&&`. Kept tests run in the same `go test` invocation as the light ones —
one binary, one `TestMain`, each test once — and the command string stays constant, so the
green-commands file still means what it says. A touched guard source adds 0.7 s of wall at the
median and 6.5 s at the ninetieth percentile; `integratechange.go` adds 127 s.

The inputs of **Context**'s third finding that nearly every change touches — `skills/`, `rules/`,
`agents/`, `commands-claude/`, `mods/`, `.flow/project.md` — get no row: the three core rows that
read them run on every default list instead (decision `real-install-is-core-not-a-row`).

### The full suite

`flow-guard run-verify-list -full <worktree> test|visual [<recording arguments>]` runs the same
list with every skip off:

- it exports `FLOW_VERIFY_FULL=1` to every command: `flow-guard unchanged` answers "changed"
  without looking, so all four suites run, and `smoke.Heavy` skips nothing;
- it skips no command on the green sha or the green-commands file, and writes both as today — a
  full run covers everything the default list runs, so a later default run at that HEAD may skip
  on it, never the reverse;
- it sets a fix run's judged `verify: fast` aside for that run, with one line saying so, and
  leaves the verdict file alone;
- it runs the visual `verify` row whole — `<specs>` empty, the `specs` row not run — whenever the
  row is declared, whatever the diff touched and whatever `decision.json` says;
- its suite rows are recorded as `<name>-full` and `visual-full`, so no median mixes the two;
- with recording arguments it begins and ends the same `verifier` row, so a green full run stands
  as the verify gate (`flow-guard check-verify-green` judges a row by its outcome and HEAD);
- a command that fails while the green sha or the green-commands file says the default list
  passed it at this HEAD gets one more line in its block,
  `escaped the default list — it passed at <sha>` — the one thing a `-full` run reads those files
  for;
- it exits 2 with `lint`, or beside `-only` or `-detect-flaky`.

A project whose `## test` has no skip to turn off gets its list run with no green skip and its
visual suite whole — which is also the first command the existing sentence "the whole suite only
when the operator asks" has had.

The ask reaches a run in plain language. Given before a run's `test` list, the next `test` list
that run reaches runs `-full`; given when no list is ahead, the session runs it at once in the
change's worktrees. It is one narrative line and one `Decisions:` bullet, recorded nowhere else;
every later list runs the default one.

### Words

"The full suite" names the `-full` run and nothing else. The parent's run after the last
implementer group and every other run of the default list becomes "the test list" wherever the
skills, `.flow/project.md`, the guards' messages and `hooks/flow-stop-guard.py` call it the full
suite, the full `## test` list or the whole suite; each reworded run-loaded sentence is entered in
the change's `verbatim-moves.txt`. One name keeps the word: the fix dispatched for a failing
test-list run stays "a full-suite fix", after its key `full-suite-fix-<n>`, with one sentence in
`skills/flow/implement.md` saying the name is the key's and the run it repairs is the test list
(decision `full-suite-fix-keeps-its-name`). The twelve lines that decision counts are not
reworded, whether they name the fix, its failure or its rows. The fix-run judge's `verify: full`
stays too, with one sentence in `skills/flow/review-panel-late-fix.md` and the header of
`stats/internal/guard/fixrunfloor.go`: it is the whole test list, never `-full`. "Smoke" is this
repository's word for its default list, in `.flow/project.md`; the contract's word is "the test
list", since a project with nothing to skip has no smoke set.

### `/flow-fast` and `/flow-interactive`

Both keep what they do — the lint list, then tests scoped by hand to what the diff names. The
sentence about the full suite is `/flow-fast`'s, in its `## 5. Verify`, and gains the command;
`/flow-interactive`'s step 4 cites that section and holds no such sentence of its own, so it is not
edited. `.flow/project.md`'s `## test` prose, which both read, says what scoping means on these
lines: a command naming tests with `-run` runs them whatever else it carries, and one scoped to a
whole package drops `-short`.

### The recording defect

Before it runs a `test` command, `run-verify-list` matches a leading
`<path>flow-guard unchanged <pathspecs> ||`, runs that guard in process, and on its exit 0 prints
`- <command> — skipped, <the guard's line>` and runs nothing: no child, no suite row. The command
counts as skipped, so a list whose every command was skipped ends its `verifier` row `skipped`,
which `flow-guard check-verify-green` reads as green. A line it cannot match, and a guard that
exits anything else, run as today. A `-full` run skips the check.

### Expected run time

Verify gives each Go suite four threads on this ten-core machine; wall is the sequential part plus
the parallel part over four. For the whole guard package that model gives 90.5 + 1388.1 / 4 =
437 s, inside the 430 to 476 s its four runs of that afternoon recorded.

- `guard-tests`: light 33 + 60 / 4 = 48 s; core 6 + 60 / 4 = 21 s; about 5 s fixed — about 1m15s.
- `stats-go`: bounded by `internal/store`'s light tests, 48 s without `-race` and an estimated 50
  to 105 s with it.
- So about 1¼ to 1¾ min for a prose change, the same plus a few seconds for one guard, and the
  same for a store change.
<!-- predicted: the guard-tests and stats-go rows flow suite list records for the first default runs after this change lands -->

Rejected alternatives, one line each:

- dropping `-race` from the default list, a `-skip` regex or skip file, build tags
  (`smoke-is-short-on-the-existing-lines`);
- reading heaviness off a test's own `Elapsed`, or a 0.5 s or 2 s threshold
  (`heavy-is-cost-of-one-second`);
- marking subtests (`heavy-marks-top-level-tests-only`);
- a brief sentence in place of the `-run` exception (`selected-run-never-skips`);
- a core list in a file, or a `-run` command of its own (`core-is-tagged-in-source`);
- light and core only, one hop of references (`touched-files-keep-their-tests`);
- a call-graph or coverage selector, Go's test cache (`no-call-graph-selector`);
- a keep table inside the guard, or two more suite lines (`keep-rows-are-arguments-on-the-line`);
- a keep row for every input a heavy test reads, or none beyond the approved rows
  (`real-install-is-core-not-a-row`);
- the whole package for an unpaired source (`unpaired-source-is-named-not-widened`);
- a second declared list (`full-is-the-list-unskipped`);
- a stored full flag (`full-ask-is-one-shot`);
- a full run that leaves the visual suite targeted (`full-runs-the-visual-suite-whole`);
- renaming the default suites (`full-rows-carry-a-suffix`);
- the fast flows running the test list (`fast-flows-keep-hand-scoping`);
- keeping "the full suite" for the default run, renaming the fix key
  (`full-suite-names-one-thing`), or renaming the fix in prose alone
  (`full-suite-fix-keeps-its-name`);
- marking from the planning tables with no tool (`test-cost-is-a-guard`);
- an implementer running the profile (`marks-come-from-a-parent-profile`);
- a marker file or a `# changed:` directive for the recording fix
  (`skipped-suite-records-no-row`);
- a full-run failure with no escape line (`full-failure-names-its-escape`).

## Critical files

- `.flow/project.md` — the `## test` fence, and its prose: the canonical suite commands, what the
  default list skips, what scoping means for the fast flows, what the escape line asks of a fix.
- `stats/internal/smoke/` — new: `Heavy`, and the static consistency test over every `_test.go`
  under `stats/`.
- `stats/internal/guard/testcost.go`, `stats/internal/guard/smokekeep.go` — new guards, each
  registered in `Registry` from its own `init` and tested beside it.
- `stats/internal/records/types.go` — `GuardHelpers` gains `test-cost` and `smoke-keep`: both
  print and refuse nothing, so a usage exit of either stays out of the guard-health view.
- `stats/internal/guard/unchanged.go` — `FLOW_VERIFY_FULL`, and the merge-base and working-tree
  diff `smoke-keep` shares; its test is `stats/internal/guard/unchanged_test.go`.
- `stats/internal/guard/runverifylist.go` — `-full`, the skipped-suite check, the escape line, the
  header contract; its test is `stats/internal/guard/run_verify_list_test.go`.
- The `_test.go` files under `stats/` holding a heavy or core test — one `smoke.Heavy(t)` call or
  one tag each; about 340 functions take the call by the planning profiles, fewer by the idle
  one, and 62 take the tag.
- `stats/internal/repotest/test_list_test.go` — pins the two Go lines of `## test`; moves with
  them.
- `stats/internal/repotest/full_suite_per_worktree_test.go` — pins the sentence
  `**The parent runs the full suite — never an implementer.**` of `skills/flow/implement.md`;
  moves with it.
- `stats/internal/guard/check_dispatch_paragraphs_test.go`,
  `stats/internal/guard/dispatchparagraphs.go` — read, not edited: the test's fixtures carry their
  own copy of the implementer brief's TARGETED TESTS paragraph, and "the full `## test` list" is
  in none of the three phrases the guard requires of that paragraph, so the reworded brief passes
  `flow-guard check-dispatch-paragraphs` as it stands.
- `skills/flow-contracts/project-configuration.md` — the `## test` row: the list is the default
  run, and `-full` is the full suite.
- `skills/flow-contracts/project-configuration-visual.md` — the `verify` and `specs` rows' "the
  operator asks for the whole suite" gains the command.
- `skills/flow/verify-and-handoff.md` — **Inline verify**: the two groups, and the `-full` ask.
- `skills/flow/implement.md` — the parent's run after the last group, its bold sentence, and the
  sentence that keeps "full-suite fix".
- `skills/flow/SKILL.md` — where an operator instruction for the run is read: the ask for the full
  suite.
- `skills/flow/review-panel.md`, `skills/flow/implementer-brief.md`,
  `skills/flow/brainstorm-planner.md` — sentences calling the default run the full suite or the
  full `## test` list.
- `skills/flow/SKILL-rationale.md` — the appendix section headed by the bold sentence of
  `skills/flow/implement.md`, which moves with it, and the one line of its history that calls a
  default run a full-suite run.
- `skills/flow/review-panel-late-fix.md`, `stats/internal/guard/fixrunfloor.go` — one sentence
  each on `verify: full`.
- `skills/flow-fast/SKILL.md` — the full-suite sentence of its `## 5. Verify` gains the command.
- `skills/flow-interactive/SKILL.md` — read, not edited: its step 4 cites that section.
- `stats/internal/guard/plansuitesteps.go` — its message and header name the full `## test` list.
- `hooks/flow-stop-guard.py` — its message names the full suite.
- `README.md` — the **Verify** line of `## Guardrails` gains `-full`, `smoke-keep` and
  `test-cost`.
- `stats/internal/guard/verifygreen.go`, `stats/internal/guard/gatherdispatch.go` — read, not
  edited: the first judges a `skipped` row green, the second carries `## test` verbatim to every
  implementer, which is why a copied `-short` must not skip a named test.

## Existing code to reuse

- `flow-guard unchanged`'s merge-base and working-tree diff (`stats/internal/guard/unchanged.go`)
  — shared with `smoke-keep`, never copied. `panelTouchedPaths`
  (`stats/internal/guard/paneltouchedpaths.go`) is not that diff: it takes a merge base as an
  argument, leaves untracked files out and applies the `## panel diff exclude` globs, so a path
  `unchanged` counts as changed could be missing from the keep list.
- `run-verify-list`'s narrowing flags `-only` and `-detect-flaky`
  (`stats/internal/guard/runverifylist.go`) — the argument shape `-full` follows.
- The green sha, the green-commands file, `# inputs:`, `# suite:` and the fix-run judge's verdict
  — unchanged.
- `stats/internal/storetest` — the precedent for a non-test package that takes `testing.TB`.
- `rvlParse` (`stats/internal/guard/runverifylist.go`) — reads the list `-full` runs; no parser
  change.
- The `Registry` and `runGuard` pattern of `stats/internal/guard` — how the two new guards are
  registered and table-tested.
- `FLOW_STATS_SKIP_SLOW_TESTS` (`stats/internal/api/stats_test.go`, read by
  `TestStatsViewTimesOutCleanly` alone) — the tree's one slow-test skip, and not the switch
  `smoke.Heavy` builds on: an environment variable no command sets, with no `-run` exception, no
  keep list and no full-run override. It stays as it is.

## Verification

Run by the parent at `flow.verify`, beyond each task's own verify step:

- **The default list is fast.** In the change's worktree, each Go line of `## test` with
  `FLOW_SMOKE_KEEP` emptied and `-json` piped into `flow-guard test-cost`: the wall time is what a
  later change touching none of these files gets, and the table shows the core rows ran and no
  marked test did. Observed against the 7m10s to 7m56s and 3m21s to 4m19s of **Context**.
- **A touched file keeps its tests.** With one heavy-only guard source touched in a scratch copy,
  `flow-guard smoke-keep` prints its paired test file and that file's heavy tests appear in the
  run's `-json` output as run, not skipped.
- **The full suite runs everything.** `flow-guard run-verify-list -full <worktree> test` at a HEAD
  the default list is green at: every suite runs, no block says skipped, and `flow suite list`
  shows rows named `guard-tests-full`, `stats-go-full`, `stats-spa-full` and
  `flow-task-list-full`.
- **A skipped suite records nothing** — the check of `## Live check`.
- **The set is consistent.** The static test of `stats/internal/smoke` passes, and
  `git grep -c '^// smoke: core' -- stats ':!stats/internal/smoke'` sums to the 62 tests of **Core**.
  <!-- predicted: git grep -c '^// smoke: core' -- stats ':!stats/internal/smoke', summed, once both marking tasks are closed -->
- **The light set passes with the heavy tests gone** — measured during planning for
  `stats/internal/guard` (the 80 s run of **Context** exited `ok`), unmeasured for the other
  packages until the marks exist.

## Non-goals

- A Playwright smoke subset: the visual suite takes 19 s and already runs only on a `ui paths`
  change.
- Narrower `# inputs:` for `stats-go`: tests in seven of its packages — `internal/repotest`,
  `internal/setuptest`, `internal/api`, `internal/fallback`, `internal/stages`,
  `internal/modelpolicy`, `internal/store` — read outside `stats/`.
- Making `internal/store`'s 364 light tests faster — they are the floor of the `stats-go` line —
  or `TestChangePlan`, or the guard package's parallelism.
- An unattended full run, and any stored "full" flag.
- A selector that computes which tests a change affects (decision `no-call-graph-selector`).

## Risks

- A regression only a heavy, unpaired, untagged test sees lands and waits for the operator's next
  full run: a shared helper's distant caller, a store change breaking a heavy `internal/api` test,
  a prose file only a heavy test reads, a `templates/` change only the two project-mode `setup.sh`
  tests read. The escape line is what turns each such case into a core entry.
- Marks drift: a test near 1 s flips between profiles, and a profile taken under load over-marks.
  Over-marking only moves a test from "always" to "when touched, and in the full suite".
- The default list slows as tests are added below the threshold; nothing fails, the medians show
  it.
- `origin/main` unfetched: `smoke-keep` exits 2 and both Go suites fail, loudly.
- `FLOW_VERIFY_FULL` and `FLOW_SMOKE_KEEP` reach every test of a run through the process
  environment, the tests of the code that reads them included: `TestUnchanged`
  (`stats/internal/guard/unchanged_test.go`) builds its `Env` on `os.Getenv`, so in a `-full` run
  it would be told "changed" where it asserts a skip. The tests of `unchanged`, of
  `run-verify-list`'s skipped-suite check and of `smoke.Heavy` pin both variables in the `Env` or
  the child environment they build — never by clearing them for the whole test binary, which
  would make `smoke.Heavy` skip in a full run — and the `-full` run of **Verification** shows one
  that does not.
- A change touching most test files, this one included, keeps nearly everything, so its own
  verify is close to a full run.
- The `-race` factor on the `stats-go` line is unmeasured; the first recorded rows are the real
  figure.
- The stderr line naming an unpaired source sits at the top of its suite's block, and a block
  longer than `run-verify-list`'s twenty-line tail drops it; the `guard-tests` block of a passing
  run is a few lines, and both Go lines print the same line.
- The static test parses every `_test.go` under `stats/` on each run; if that costs 1 s or more
  under `-race` it is itself a core row, never a marked one.

## Live check

The store is the runtime this change writes to: `run-verify-list` adds a suite row to flowd for
every suite it runs, and a row for a suite its guard skipped (decision
`skipped-suite-records-no-row`) is a defect only the store shows. Run at `flow.verify` from the
change's worktree `<worktree>`, on two stand-ins. A copy of the dev store stands in for the dev
store, because on unrepaired code the check writes the four false rows the medians must not get.
A throwaway worktree at `origin/main` stands in for a checkout whose branch changed nothing,
which the change's own worktree is not; it is a worktree and never a clone, since flowd refuses a
row for a project key it does not hold and a clone's key is its own.

1. `scripts/flow-guard live-store-copy up smoke-live`; export the `FLOW_ADDR` and
   `FLOW_RECORDS_ADDR` line it prints into every step below.
2. `scripts/flow-guard throwaway-worktree create <worktree> <project>-worktrees/smoke-test-set-live`,
   called `<copy>` below; `git -C <copy> checkout -q --detach origin/main`; write
   `{"visual":{"verify":"skipped"}}` to `<copy>/.superpowers/sdd/decision.json`. A detached
   `<copy>` has no state record to read a base from, and `run-verify-list` runs the visual
   `verify` row on a base that does not resolve: without that file the list restarts the UI-test
   stack on port 4174 from `<copy>`.
3. Control: `flow suite record -C <copy> -suite live-control -- true`, then
   `flow suite list -C <copy> -suite live-control` prints one `live-control` row. No row means
   the copy is not being written, and an unchanged count in step 6 would then prove nothing: the
   check has failed.
4. Before: for each of `guard-tests`, `stats-go`, `stats-spa` and `flow-task-list`, the count of
   rows `flow suite list -C <copy> -suite <name> -limit 100000` prints — the lines opening with
   the suite's name and a tab.
5. `scripts/flow-guard run-verify-list <copy> test`, run from `<worktree>`: the change's guard,
   reading the `## test` list of `origin/main`, where every suite's guard exits 0.
6. After: step 4's four counts again.
7. Teardown, whatever the result: `scripts/flow-guard throwaway-worktree remove <worktree> <copy>`,
   then `scripts/flow-guard live-store-copy down smoke-live`.

**Working:** step 5 exits 0 and prints four command lines, each ending
`— skipped, unchanged: nothing under <pathspecs> changed since <sha> — suite skipped`, with no
`recorded:` line under any; the four counts of step 6 equal those of step 4.

**Not working**, as the code this change starts from behaves: step 5 exits 0 just the same, and
each of its four command lines ends `— exit 0` with the guard's line and
`recorded: suite <name> 1.1s exit 0` under it; each count is one higher — 157 to 158 for
`guard-tests`, 134 to 135 for `stats-go`, 111 to 112 for `stats-spa`, 7 to 8 for
`flow-task-list`.
<!-- measured: steps 1 to 7 as written, run from the change's worktree at 99ca4f38 on 2026-11-13 against a copy of the dev store of that day; `origin/main` resolved to 99ca4f38 there, the copy answered on port 4175, the control row was listed, and the teardown left no worktree and no listener on 4175 -->

The rows a `-full` run writes are the third bullet of **Verification**: the change's own worktree
and the dev store, where those rows are true.

## Decisions

### The default list is `-short` on the lines that exist

**ID:** smoke-is-short-on-the-existing-lines
**Status:** active
**Chosen:** the two Go lines of `## test` gain `-short` and keep `-race`, their suite names, guards and `# inputs:` — one Go-native switch, one build flavour shared with the full run, and a data race in touched code is still caught by default.
**Considered:** dropping `-race` from the default list — races would surface only on a full run, the saving on the `stats-go` line is unmeasured, and a second build flavour is paid whenever a full run follows. A `-skip` regex or a skip file — a name list that goes stale on a rename. Build tags — the tests would have to be split across files.

### What makes a test heavy

**ID:** heavy-is-cost-of-one-second
**Status:** active
**Chosen:** cost of 1 s or more — own `Elapsed` plus that of every descendant that paused for `t.Parallel()`, under the default line's flags less `-short`, on an idle machine — which marks 185 guard tests and leaves a guard light set of about 48 s.
**Considered:** the test's own `Elapsed` — the measured run disproved it: 305 s against 80 s. A 0.5 s threshold — 258 guard tests marked for a light set of about 21 s, 73 more marks to keep current. A 2 s threshold — 123 marked and a light set of about 85 s.

### Where a heavy mark may sit

**ID:** heavy-marks-top-level-tests-only
**Status:** active
**Chosen:** `smoke.Heavy` is the first statement of a top-level test and nowhere else — one rule the static test states in one line, and a skipped test shows as one skipped test in `go test -json`.
**Considered:** also allowing it as the first statement of a `t.Run` body, for a test one subtest row dominates — it keeps that test's light rows in the default run, but the static test must then walk closures, and a test such as `TestChangePlan` is out of this change's scope either way; marked whole, it still runs when its file is touched and in every full run.

### A run that names its tests never skips them

**ID:** selected-run-never-skips
**Status:** active
**Chosen:** `smoke.Heavy` skips nothing when the run carries `-run` — an implementer or a mutation reviewer who copies `-short` from `## test` into `go test -run '<name>'` still runs the test they named.
**Considered:** a sentence in the implementer brief alone — a copied `-short` would still print `ok` for a test that never ran.

### Where the core set is recorded

**ID:** core-is-tagged-in-source
**Status:** active
**Chosen:** a `// smoke: core — <capability>` comment above the test and no skip call, held consistent by one static test — a rename or deletion carries the tag with it.
**Considered:** a file, `.flow/smoke-core.txt`, checked by a new lint guard — a second copy of names the tree already holds, and one more guard to keep it honest. A `-run '^(…)$'` command of its own in `## test` — a renamed test drops out of the regex with `ok` still printed, and the guard package is built and started twice.

### Where the core tag sits in a comment

**ID:** core-tag-ends-the-doc-comment
**Status:** active
**Chosen:** the tag is the last line of the comment directly above the `func` — true for a test with no comment and for one with a doc comment alike, and one `go/parser` read of the function's `Doc` finds it.
**Considered:** the first line of the comment — Go's convention opens a doc comment with the function's name.

### What the default run adds for touched files

**ID:** touched-files-keep-their-tests
**Status:** active
**Chosen:** `flow-guard smoke-keep` derives, from the diff on every run, the test files paired by name with each touched Go file, and the declared paths for non-Go inputs — nothing stored, and the 41 guard sources whose own tests are all heavy still run them when touched.
**Considered:** light and core only — nothing would run those 41 sources' tests until a full run. Adding the tests of files that reference a touched source, one hop — 38 s at the ninetieth percentile, and blind to a guard reached through `Registry[name]` or a shell line.

### How the keep list is passed

**ID:** keep-list-is-space-separated
**Status:** active
**Chosen:** `smoke-keep` prints its paths on one line separated by spaces and `FLOW_SMOKE_KEEP` carries that line as printed — what `keep=$(…)` yields with no further quoting.
**Considered:** one path per line — the same content, and a multi-line environment value on a `## test` line for nothing.

### Where the rows for non-Go inputs live

**ID:** keep-rows-are-arguments-on-the-line
**Status:** active
**Chosen:** `<path prefix>=<kept path>` arguments to `smoke-keep` on the two Go lines of `## test` — the rows are read where the suites are declared.
**Considered:** a table inside the guard's source — invisible from `.flow/project.md`. Two more `unchanged`-guarded suite lines for `internal/repotest` and `internal/setuptest` — no code, but those packages' light tests run twice, a third and fourth Go command share ten cores, and both packages run a second time in every full run.

### Inputs nearly every change touches are covered by core tests, not rows

**ID:** real-install-is-core-not-a-row
**Status:** active
**Chosen:** one more row, `scripts/` keeping `internal/setuptest`, and three more core tags — `TestGlobalInstallPopulatesSkillDirs`, `TestFlowTaskListLanded` and `TestGoSourcesHashSeesGoEdits` — so the real install and the two small readers of `mods/`, `stats/Makefile` and `.flow/project.md` run on every default list for about 2 s of wall.
**Considered:** a row for every input a heavy test of the two packages reads — a change under `skills/` or `rules/`, the commonest here, would keep all 26 heavy `internal/setuptest` tests, 186 cost-seconds, on a list meant to take a minute and a half. The rows as first approved — a `scripts/flow-guard` change would run none of the 26 setup tests that copy it, and a prose change that breaks `setup.sh global` would pass every default run.

### A source file that pairs with no test

**ID:** unpaired-source-is-named-not-widened
**Status:** active
**Chosen:** nothing is kept for it and one stderr line names it — the run stays light and core, and the report shows the gap.
**Considered:** running its whole package unskipped — the full guard package on the one guard commit in twenty that touches a shared helper.

### How the full suite is declared

**ID:** full-is-the-list-unskipped
**Status:** active
**Chosen:** `-full` exports `FLOW_VERIFY_FULL=1` and runs the same `## test` list — one list, so the full suite can never be smaller than the default one.
**Considered:** a second section, `## test full`, holding the commands without their guards — a suite added to one list and forgotten in the other, held only by a pin test that the two stay in step.

### How long an ask for the full suite lasts

**ID:** full-ask-is-one-shot
**Status:** active
**Chosen:** the ask covers the next `test` list the run reaches, or runs at once when none is ahead, and is stored nowhere — one narrative line and one `Decisions:` bullet.
**Considered:** a field in the state record, as `reviewBeforeLanding` is, so every later verify of the change is full until it lands — a store field, an API field and a contract change.

### What a full run does with the visual suite

**ID:** full-runs-the-visual-suite-whole
**Status:** active
**Chosen:** it runs the `verify` row whole whenever the project declares the row — "full" means every skip off, and a Go API change can move a view with no `ui paths` file touched.
**Considered:** whole only when the diff touches `ui paths` — the case above stays unseen. Leaving the row targeted and asking for the visual suite separately — a full run that leaves a skip on.

### How a full run's suite rows are named

**ID:** full-rows-carry-a-suffix
**Status:** active
**Chosen:** `<name>-full` and `visual-full` — no median mixes a full run with a default one, and the four default names stay.
**Considered:** renaming the default suites `*-smoke` — it breaks the two pins that hold the names, `stats/internal/repotest/test_list_test.go` and `stats/internal/guard/run_verify_list_test.go`, and restarts four recorded medians, for no gain once the full rows are named apart.

### What `/flow-fast` and `/flow-interactive` run

**ID:** fast-flows-keep-hand-scoping
**Status:** active
**Chosen:** they keep scoping tests by hand to what the diff names — that is already their fast path, and in another project `## test` may be a slow list with nothing to skip.
**Considered:** both running the test list — about a minute and a half on every interactive request. `/flow-fast` alone — two rules for two flows that share one verify section.

### What "the full suite" names

**ID:** full-suite-names-one-thing
**Status:** active
**Chosen:** the `-full` run and nothing else; the default run is "the test list", and the dispatch key `full-suite-fix-<n>` stays as an identifier.
**Considered:** leaving "the parent runs the full suite" in `skills/flow/implement.md` — an operator's "run the full suite" would then resolve to the default list. Renaming the key — a hook, the store and 61 lines in 18 files for a word.
<!-- measured: git grep -c -E 'full-suite-fix' -- skills rules hooks README.md .flow stats @ 99ca4f38 -->

### What the fix for a failing test-list run is called

**ID:** full-suite-fix-keeps-its-name
**Status:** active
**Chosen:** "a full-suite fix" stays wherever prose names that dispatch — the name its key carries — with one sentence in `skills/flow/implement.md` saying the run it repairs is the test list.
**Considered:** renaming it "a test-list fix" on the 12 lines in 6 skill files that use it — prose would then call a dispatch by a name its key does not carry.
<!-- measured: git grep -c -i -E 'full-suite (fix|failure|row|fixer)' -- skills rules hooks README.md .flow @ 99ca4f38 -->

### How heaviness is measured after this change

**ID:** test-cost-is-a-guard
**Status:** active
**Chosen:** `flow-guard test-cost`, the Go form of the reader that made the planning tables — re-profiling is one command anyone can run.
**Considered:** marking once from the planning tables and adding no tool — the next re-profile would start from a reader that existed only in a gitignored directory.

### What `test-cost` prints

**ID:** test-cost-prints-tabs
**Status:** active
**Chosen:** one line per test — cost in seconds to two decimals, the package as `go test -json` names it, the test — separated by tabs, heaviest first, so `cut` and `sort` read it with no parsing.
**Considered:** aligned columns for reading — a second consumer, the marking tasks, would have to split on runs of spaces. The package trimmed to its path inside the module — a module-path lookup for a shorter column.

### Who profiles, and when

**ID:** marks-come-from-a-parent-profile
**Status:** active
**Chosen:** the parent runs the two profiles at flow.sdd-tdd, once the task adding `flow-guard test-cost` is closed and before the first marking task is dispatched, and the marking tasks mark the rows it wrote — the marks then come from one idle run under the default line's own flags.
- [x] Parent: at flow.sdd-tdd, once the task adding `flow-guard test-cost` is closed and before the first marking task is dispatched, with no other test, build or dispatch of this run in flight: from `<worktree>/stats`, run each Go line's `go test -race -count=1` command of today's `## test` with `-json` added and its `GOMAXPROCS` kept, pipe each into `flow-guard test-cost -min 1`, and write the two tables to `<worktree>/.superpowers/sdd/smoke-profile/guard-idle.tsv` and `<worktree>/.superpowers/sdd/smoke-profile/stats-go-idle.tsv`.
**Considered:** an implementer profiling inside a marking task — `flow-guard plan-suite-steps` refuses a task step that runs a package's suite, and the implementer brief forbids it. Marking from the planning tables — the `stats-go` one was taken under load and without `-race`.

### Where the idle profiles are written

**ID:** idle-profiles-get-their-own-files
**Status:** active
**Chosen:** `guard-idle.tsv` and `stats-go-idle.tsv` under the worktree's gitignored `.superpowers/sdd/smoke-profile/` — beside the planning tables, which stay as the figures this design cites.
**Considered:** overwriting `guard-cost.tsv` and `statsgo-cost.tsv` — the planning figures would be gone, and their six-column shape is not what `test-cost` prints.

### A suite its guard skipped

**ID:** skipped-suite-records-no-row
**Status:** active
**Chosen:** `run-verify-list` matches a leading `flow-guard unchanged <pathspecs> ||`, runs the guard in process, and on its exit 0 prints the skip and runs nothing — one touch point, and the command counts as skipped for the row's outcome.
**Considered:** a marker file `unchanged` writes and `flow suite record` reads — three touch points against one. A `# changed:` directive replacing the guard prefix — a contract change for a recording fix.

### A full-run failure the default list let through

**ID:** full-failure-names-its-escape
**Status:** active
**Chosen:** the failing command's block says `escaped the default list — it passed at <sha>` when the green files show the default list passed it at this HEAD — the one read a `-full` run makes of them — so each escape becomes a core entry.
**Considered:** no line — the core set's only trigger would be a re-profile, which sees slowness and never a missed regression.

### A selector that computes affected tests

**ID:** no-call-graph-selector
**Status:** active
**Chosen:** none is built; the tree already names a guard's tests after the guard, so the diff alone finds them, and the payoff of anything finer is capped at the minute or two the default list costs.
**Considered:** a static call graph — 50 guard test files reach a guard through `Registry["<name>"]`, a map of 119 same-signature functions, so any sound resolution reaches every guard from every test; others run the shim, the built binary, `python3 hooks/<x>.py` or `setup.sh`, and tests in eight packages — `internal/guard` and the seven of **Non-goals** — read `skills/`, `rules/`, `hooks/`, `.flow/project.md` or `scripts/` at run time. A per-test coverage map — stored state regenerated by an instrumented run over more than two thousand tests, blind to both points above, whose staleness is a test that silently does not run. Go's test cache — it tracks only files inside the package's module (`go help test`), and these tests read outside it, which is why every `## test` line carries `-count=1`.
<!-- measured: grep -l 'Registry\["' stats/internal/guard/*_test.go | wc -l (50) and git grep -h -E 'Registry\[[^]]+\] *= ' -- 'stats/**/*.go' ':!stats/**/*_test.go' | wc -l (119) @ 99ca4f38 -->

## Open questions
