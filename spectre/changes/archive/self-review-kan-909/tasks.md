# self-review-kan-909

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close KAN-909 — the setuptest SIGINT path waits for every sandbox writer, not only
installer runs, and no writer starts once it has begun.

**Architecture:** a `gate` (mutex-held closed flag plus a WaitGroup only `Add`ed while open)
replaces `running`; `newGroup` and `makeFixtureRepo` hold it; the signal handler closes it
between `stopRuns()` and `cleanup()`; a `testing/synctest` test pins it.

**Tech Stack:** Go 1.26 (`stats/internal/setuptest`), `testing/synctest`.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-21-self-review-kan-909-design.md` in this worktree.

## Global Constraints

- Self-review fix: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-909`, branch `self-review-kan-909`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- Run `cd stats && gofmt -w .` before committing; `cd stats && make -s web-dist && go vet ./... && gofmt -l .` stays clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.
- No new inline lint suppressions.

## Review Focus

- A group refused by a closed gate skips; it never fails or hangs the package.
- `leave` runs exactly once per successful `enter` — via `t.Cleanup` — including when the group's test fails with `t.Fatal`.
- No nested `enter` on one goroutine path is needed; `close` never deadlocks against a group that is itself blocked on nothing but its installer.
- An uninterrupted `scripts/test-setup.sh` run behaves exactly as before: same groups, same results, sandbox removed.
- `makeFixtureRepo` failing still returns exit 2 and leaves the gate's count balanced.

- [x] 1. KAN-909 — hold every sandbox writer behind one gate the SIGINT path closes
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Confirm on the untouched
    tree that `stats/internal/setuptest/main_test.go`'s handler waits only on `running`, that
    `running` is counted only in `runSetup` (`helpers_test.go`), and that every top-level test
    that writes the sandbox begins with `newGroup(t)` (`grep -n '^func Test' stats/internal/setuptest/*.go`
    against `grep -c 'newGroup(t)'`). Anything that disagrees with `design.md`'s Context is
    reported, not worked around.
  - [x] **Step 2: Write the failing test** `TestWritersGateHoldsCloseUntilEveryWriterLeaves` in
    `stats/internal/setuptest/selfcheck_test.go` (add `testing/synctest` to its imports):

```go verified:spiked as zz_spike_test.go in stats/internal/setuptest at b76c2cf1 (type renamed); go test -race -count=1 -run 'TestSpikeGate$' ./internal/setuptest/ passes, and fails with close's Wait removed
func TestWritersGateHoldsCloseUntilEveryWriterLeaves(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var w gate
		if !w.enter() {
			t.Fatal("an open gate refused a writer")
		}
		closed := false
		go func() { w.close(); closed = true }()
		synctest.Wait()
		if closed {
			t.Fatal("close returned while an admitted writer was still in")
		}
		if w.enter() {
			t.Fatal("a closing gate admitted a new writer")
		}
		w.leave()
		synctest.Wait()
		if !closed {
			t.Fatal("close did not return after the last writer left")
		}
	})
}
```

  - [x] **Step 3: Run it to see it fail.** Run:
    `cd stats && go test -count=1 -run 'TestWritersGateHoldsCloseUntilEveryWriterLeaves$' ./internal/setuptest/`
    Expected: build failure, `undefined: gate`.
  - [x] **Step 4: Add the gate** to `stats/internal/setuptest/main_test.go`, beside the package
    vars, per decision `gate-shape`: `type gate struct { mu sync.Mutex; closed bool; in sync.WaitGroup }`
    with `func (w *gate) enter() bool`, `func (w *gate) leave()` and `func (w *gate) close()`,
    each with a one-line doc comment. Replace the `running sync.WaitGroup` var with
    `writers gate // every sandbox writer in flight; the interrupt handler closes it before removing the sandbox`.
  - [x] **Step 5: Wire it.** In `main_test.go`: the handler body becomes
    `<-sig; stopRuns(); writers.close(); cleanup(); os.Exit(130)`, its comment saying it now
    waits for every sandbox writer, not only installers; `runMain` wraps
    `makeFixtureRepo(fixture, false)` in `if !writers.enter() { return 130 }` … `writers.leave()`,
    leaving on the error path too. In `stats/internal/setuptest/helpers_test.go`: `newGroup`,
    after `t.Parallel()` and before its `MkdirAll`, does
    `if !writers.enter() { t.Skip("interrupted") }` then `t.Cleanup(writers.leave)`; `runSetup`
    drops `running.Add(1)` and `defer running.Done()`, and its process-group comment keeps
    saying the installer is killed on an interrupt.
  - [x] **Step 6: Run the test and the package.** Run:
    `cd stats && go test -race -count=1 -run 'TestWritersGateHoldsCloseUntilEveryWriterLeaves$' ./internal/setuptest/`
    Expected: `ok`. Then prove the pin: delete `close`'s `w.in.Wait()` line, rerun — expected
    `FAIL … close returned while an admitted writer was still in` — and restore it. Then
    `scripts/test-setup.sh`; expected exit 0 with the close-out's three `✓` lines.
  - [x] **Step 7: Lint.** Run `cd stats && gofmt -w . && make -s web-dist && go vet ./... && gofmt -l .`;
    expected no output and exit 0.
  - [x] **Step 8: Commit** `stats/internal/setuptest/main_test.go`,
    `stats/internal/setuptest/helpers_test.go` and `stats/internal/setuptest/selfcheck_test.go`
    with the subject below.

**Build:** green
**Files:** `stats/internal/setuptest/main_test.go`, `stats/internal/setuptest/helpers_test.go`, `stats/internal/setuptest/selfcheck_test.go`
**Tests:** `TestWritersGateHoldsCloseUntilEveryWriterLeaves`
**Regression:** `TestWritersGateHoldsCloseUntilEveryWriterLeaves` — reverting removes the gate (the test no longer builds); a gate whose `close` skips the wait fails "close returned while an admitted writer was still in", and one that admits after close fails "a closing gate admitted a new writer".
**Baseline:** before=40 after=41
<!-- measured: cat stats/internal/setuptest/*_test.go | grep -cE '^func Test[A-Z]' @ b76c2cf1 -->
**Commit:** fix(setuptest): hold every sandbox writer behind the gate the interrupt handler closes
**After:** none

**Decision:** gate-at-the-group
**Decision:** gate-shape
**Decision:** gate-replaces-running
**Decision:** synctest-pin
