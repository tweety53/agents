## Context

Self-review fixes for `kan-1037-stats-weekly-costs-dashboard`, on the change's own branch
`spectre/kan-1037-stats-weekly-costs-dashboard` in
`/Users/tweety53/Projects/agents-worktrees/kan-1037-stats-weekly-costs-dashboard`
(`fixes-ride-the-change`). The change is archived; this plan never touches `spectre/changes/archive/`.

**F1.** `stats/internal/guard/runverifylist.go` records `end -outcome skipped` when every command of a
test group was cache-skipped (its comment: a lint list always ran `check-spec-reach.sh`, so it never
ends skipped). Signal seven in `stats/internal/guard/unfinishedwork.go` counts only
`d.Outcome == "completed"`. `skills/flow/verify-and-handoff.md` **Recording** still says a group that
exited non-zero records no `end` and the parent ends the row by hand.

**F2.** `RecordDispatch` keeps the first `startedAt` of a (change, token, key) row, and signal seven
compares the newest code commit's author time with the verify row's `startedAt`, so a re-run under
the same key can never cover a later commit.

**F3.** `stats/internal/guard/prove_reproducer_test.go`, in `TestProveReproducerSignals`: for a
non-ignored signal the test waits for the run to die, then `os.OpenFile(release, os.O_WRONLY, 0)` to
unblock the orphaned leg's `read -r _ < release`. A FIFO's blocking write-open waits for a reader;
when the leg is already gone — the SIGINT case signals the whole process group (`target = -target`),
killing the leg's shell before it opens the FIFO, or the leg never reached its `read` — no reader
ever arrives and the open blocks until the test binary's timeout. Not reproduced in 45 runs: the
window is the leg's death before its open.

**F4.** The mutation slot's throwaway copy (`throwaway-worktree.sh create`) holds no
`node_modules`; the slot symlinked the shared one in, and `make web-dist` → `npm ci` removed the
shared directory's entries through the symlink.

**F5.** `ccWaveGroupCopy` (`stats/internal/guard/cleanupcomplete.go`) matches a detached worktree
whose last path element is `<name>-wave-group-<digits>`; the panel's copy is
`<worktree>-mutation-<round>` (`skills/flow/review-panel-optional-slots.md`, **The throwaway
worktree**), also detached, and is not matched.

**F6.** No `signal.Ignore(syscall.SIGPIPE)` anywhere in flow-guard; Go's default kills the process
on a write to a broken stdout/stderr pipe, so the deferred `end` never runs.

**F7.** `stats/cmd/flow/stage.go` errors "expected exactly one argument, the change name, or
-jira-key with no argument"; `flow record` subcommands take `-change`.

**F8.** `stats/internal/store/migrations/0008_stage_run_session_token.sql` adds
`stage_runs.stage_key` and index `stage_runs_stage_key`; `git grep stage_key -- stats` finds no
other Go or SQL use. The newest migration is `0047_drop_loop_dispatch_twins.sql`.

**Repairs of what the defects wrote.** F1/F2 left verify rows `skipped` or stale — the fix makes
signal seven read the existing `skipped` rows correctly, so no row needs rewriting. F4 emptied a
`node_modules`, restored by the next `npm ci`. F5's copy was removed in kan-1037's run. F6's unended
row is the verify-key re-run's concern (F2). None writes a store row that needs a repair command.

## Decisions

### A skipped verify row covers the commits before it

**ID:** f1-skipped-is-verified
**Status:** active
**Chosen:** signal seven counts a verifier row with outcome `completed` or `skipped`. A `skipped`
row only ever comes from a test group whose every command was unchanged since a green clean run, and
the lint group never ends skipped. Correct the **Recording** paragraph of `verify-and-handoff.md` to
say the script ends the row itself: `completed`, `skipped` when every test command was cache-skipped,
and `blocked -cause test-failure` when a group exits non-zero.
**Considered:** making `run-verify-list.sh` end skipped groups `completed` — loses the store's
distinction between a ran and a skipped list.

### A verify re-run after a code commit takes a fresh key

**ID:** f2-fresh-verify-key
**Status:** active
**Chosen:** prose plus the guard's reason line: `verify-and-handoff.md` states a verify re-run after a
new code commit uses a key no earlier verify row of the change used (`verify-<k>`, `<k>` counting
from 2), and signal seven's reason ends "— a re-run after a new commit takes a fresh `verify-<k>`
key". **Considered:** making `RecordDispatch` refresh `startedAt` on re-begin — changes the
idempotency every other dispatch row relies on.

### Release the FIFO without blocking

**ID:** f3-nonblocking-release
**Status:** active
**Chosen:** a test helper `prReleaseFIFO(path string)` opening with `os.O_WRONLY|syscall.O_NONBLOCK`
(ENXIO when no reader — ignored), used at both release sites; a test
`TestPRReleaseFIFOWithNoReaderReturns` creates a FIFO with no reader and asserts the helper returns
within 5 s (it hung forever with the blocking open; proven by swapping the flag back in the
implementer's run). **Considered:** `O_RDWR` — never blocks but on a FIFO is unspecified by POSIX.

### No shared node_modules in a throwaway copy

**ID:** f4-no-symlinked-node-modules
**Status:** active
**Chosen:** one sentence in **The throwaway worktree** forbidding a symlink to the shared
`node_modules` (the copy's own `npm ci` installs it), and in `stats/Makefile` `web-test` and
`web-build` open with `test ! -L web/node_modules || { echo "...refusing npm ci over a symlinked web/node_modules" >&2; exit 1; }`.
Test: a case in `stats/internal/repotest/make_build_test.go`'s file, new test
`TestMakeWebBuildRefusesSymlinkedNodeModules`, running `make -n`-free: a temp copy of the Makefile and
a `web/node_modules` symlink, `make web-build` exits non-zero with the message and the symlink target
untouched.

### Match the mutation throwaway copy

**ID:** f5-mutation-copy
**Status:** active
**Chosen:** `ccWaveGroupCopy` accepts `<name>-wave-group-<digits>` or `<name>-mutation-<digits>`
(loop over the two suffixes), doc comment updated; a case in `TestCheckCleanupComplete` with a
detached `<name>-mutation-1` worktree reporting it.

### Ignore SIGPIPE in runVerifyList

**ID:** f6-ignore-sigpipe
**Status:** active
**Chosen:** `signal.Ignore(syscall.SIGPIPE)` as the first statement of `runVerifyList`; writes then
return EPIPE and the list runs to its `end`. Test `TestRunVerifyListSurvivesAClosedStdout` runs the
built flow-guard `run-verify-list` subprocess with stdout a pipe whose read end is closed, and
asserts the row's `end` was recorded (the fake store's log) and the process did not die of SIGPIPE.

### -change on flow stage

**ID:** f7-stage-change-flag
**Status:** active
**Chosen:** `-change <name>` registered on begin/end/wrap/mark/sessions' flag sets; when set, it
fills the change and no positional is allowed (both given → usage error); usage text updated. Test
`TestStageBeginAcceptsChangeFlag`.

### Drop stage_key

**ID:** f8-drop-stage-key
**Status:** active
**Chosen:** `stats/internal/store/migrations/0048_drop_stage_runs_stage_key.sql`:
`DROP INDEX IF EXISTS stage_runs_stage_key; ALTER TABLE stage_runs DROP COLUMN IF EXISTS stage_key;`
with a header comment citing kan-1037. Test `TestStageRunsCarriesNoStageKey` in
`stats/internal/store/drop_stage_key_test.go`, asserting `information_schema.columns` holds no
`stage_runs.stage_key` after migration, using the package's existing test-database helper.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — eight tasks across guard, CLI, store and Makefile code, several touching
signal and FIFO semantics.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** three fixer dispatches: g1 Tasks 1, 2, 6 (verify row: `unfinishedwork.go`,
`runverifylist.go`, `verify-and-handoff.md`); g2 Tasks 3, 5 (guard tests and cleanup guard);
g3 Tasks 4, 7, 8 (Makefile, CLI, migration).

## Open questions

None.

## Recommended approach

The decisions above, one task each, in `tasks.md`.

## Critical files

`stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/runverifylist.go`,
`stats/internal/guard/prove_reproducer_test.go`, `stats/internal/guard/cleanupcomplete.go`,
`stats/cmd/flow/stage.go`, `stats/Makefile`, `stats/internal/store/migrations/`,
`skills/flow/verify-and-handoff.md`, `skills/flow/review-panel-optional-slots.md`.

## Existing code to reuse

`TestCheckUnfinishedWork`'s `verified(...)` fixture, `TestCheckCleanupComplete`'s wave-group case,
`run_verify_list_test.go`'s fake store, `stage_test.go`'s fake server, the store package's
test-database helper.

## Verification

Each task's targeted tests; `cd stats && make -s web-dist && go vet ./... && gofmt -l .`;
`scripts/check-references.sh`.

## Live check

None — no running service or persistent state is exercised beyond the per-task tests; the migration
is exercised by the store test's own migrated database.
