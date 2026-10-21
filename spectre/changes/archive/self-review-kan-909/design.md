## Context

A standalone self-review fix loop for KAN-909, in worktree
`/Users/tweety53/Projects/agents-worktrees/self-review-kan-909` on branch `self-review-kan-909`,
based at `b76c2cf1` (origin/main). The finding was traced against that commit before its fix was
designed; the brainstorm design is `.superpowers/sdd/2026-10-21-self-review-kan-909-design.md` in
this worktree.

**The defect is live.** `stats/internal/setuptest/main_test.go:120-126` installs the handler
`<-sig; stopRuns(); running.Wait(); cleanup(); os.Exit(130)`, and `running` is counted only by
`runSetup` (`helpers_test.go:85-86`, `running.Add(1)` / `defer running.Done()`), with no check
that the harness is stopping. Two consequences:

1. **WaitGroup misuse.** A `runSetup` reaching `running.Add(1)` after `stopRuns()` while the
   handler is inside `running.Wait()` with the counter at zero is an Add from zero concurrent
   with Wait — the contract `sync.WaitGroup` documents as misuse.
2. **Unheld test-side writes.** `newGroup`'s `MkdirAll`, `newHome`, `seedFile`,
   `seedStaleSkillDir`, `copyFile`, `appendTo`, `leak` (`fingerprint_test.go`), raw
   `os.WriteFile`/`os.Symlink`/`os.Remove`/`os.RemoveAll` calls in `delimiters_test.go`,
   `harnesses_test.go`, `containment_test.go` and `selfcheck_test.go`, `runSetup`'s own log
   `os.Create`, and `makeFixtureRepo` on the main goroutine (called after the handler is
   installed) all write the sandbox with nothing holding `cleanup` off.

Every top-level test that writes the sandbox starts with `newGroup(t)`; the five that do not
(`TestMain`, `TestSetupRefusalKeepsTheInstallerInTheSandbox`,
`TestCleanupRemovesOnlyHarnessSandboxes`, `TestCloseOutFailsOnAnyMovedFingerprint`,
`TestLoopHookTimeoutsFailClosed`) write nothing under the sandbox. `newGroup` is therefore the one
entry point every test-side write and every installer run funnels through.

Blast radius, counted: three files —
`stats/internal/setuptest/main_test.go`,
`stats/internal/setuptest/helpers_test.go`, `stats/internal/setuptest/selfcheck_test.go`. Not
`big`.

## Live check

None: the change is test-harness code with no running service or persistent state; its
behaviour is pinned by the synctest test and exercised by `scripts/test-setup.sh`.

## Decisions

### Gate the sandbox at the group, not at each write

**ID:** gate-at-the-group
**Status:** active
**Chosen:** one `writers gate` entered once per assertion group in `newGroup` (left via
`t.Cleanup`) and once around `makeFixtureRepo` in `runMain` — the single entry point every
test-side write and installer run already funnels through; a group the closed gate refuses calls
`t.Skip("interrupted")`.
**Considered:** wrapping each write helper (`seedFile`, `copyFile`, `mkdirAll`, …) in
enter/leave — about twenty raw write sites across six files would each need it, and any write
added later without it races again; an `RWMutex` held for read by writers — a writer's nested
read lock deadlocks once the handler's write lock is pending, and a blocked `RLock` cannot tell a
group to skip.

### Gate shape: mutex-held closed flag plus a correctly used WaitGroup

**ID:** gate-shape
**Status:** active
**Chosen:** `type gate struct { mu sync.Mutex; closed bool; in sync.WaitGroup }` with
`enter() bool` (under `mu`: refuse when closed, else `in.Add(1)`), `leave()` (`in.Done()`) and
`close()` (set `closed` under `mu`, then `in.Wait()`). Every `Add` happens under the mutex while
the gate is open, so none can run concurrently with `close`'s `Wait`.
**Considered:** keeping the bare `running` WaitGroup with a `runCtx.Err()` check before `Add` —
the check and the Add are not atomic, so the race stays; a channel-based semaphore — more code
for the same guarantee.

### The gate replaces `running`

**ID:** gate-replaces-running
**Status:** active
**Chosen:** delete `running` and `runSetup`'s `running.Add/Done`; the group's hold covers its
installer runs. The handler is `<-sig; stopRuns(); writers.close(); cleanup(); os.Exit(130)` —
`stopRuns` first, so in-flight installers are killed and their groups finish; a group's next
`runSetup` fails to start on the cancelled `runCtx` (`cannot run setup.sh`) and ends the test.
**Considered:** keeping both counters — two waits for one fact.

### Pin the gate under testing/synctest

**ID:** synctest-pin
**Status:** active
**Chosen:** `TestWritersGateHoldsCloseUntilEveryWriterLeaves` in `selfcheck_test.go` (the file
pinning the harness's own safety logic) runs inside `synctest.Test` on a fresh `gate`: enter;
`close` in a goroutine; `synctest.Wait()`; assert `close` has not returned and a second `enter`
is refused; `leave`; `synctest.Wait()`; assert `close` returned. `sync.WaitGroup.Wait` is durably
blocking in a bubble when `Add` ran inside it (Go 1.25+, module is `go 1.26.5`), so the
assertions are exact, not timing-dependent.
<!-- measured: a scratch zz_spike_test.go in stats/internal/setuptest with this gate and test; go test -race -count=1 -run 'TestSpikeGate$' ./internal/setuptest/ → ok; with close's Wait removed → FAIL "close returned while a writer was in" @ b76c2cf1 -->
**Considered:** sending a real SIGINT to a child test binary — needs a timing window to land the
signal mid-write, the flake class the finding names; testing the handler goroutine itself — it
ends in `os.Exit`.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — concurrency code (a hard seam): a wrong ordering in `close` or a
missed `leave` deadlocks every interrupted run or the whole package.
**Considered:** `flow-low` — the diff is small and test-pinned, but the review must judge
lock/WaitGroup ordering the test pins only for the gate, not its wiring.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-kan-909-fix`, Task 1.
**Considered:** none — the plan has one task.

## Open questions
