- [x] 5. Run: a second Ctrl-C kills a run that is judging its held cases

**Build:** green
**Files:** `cmd/flow-eval/main.go`, `cmd/flow-eval/main_test.go`, `README.md`
**Tests:** `TestSecondInterruptKillsARun`
**Regression:** `TestSecondInterruptKillsARun` fails when the first stop signal leaves the process catching later ones, so a second SIGTERM no longer terminates it.
**Baseline:** before=12 after=13
<!-- measured: cat cmd/flow-eval/*_test.go | grep -cE '^func Test\w*\(t \*testing\.T\)' @ branch spectre/kan-1075-impl-alignment-judge -->
**Commit:** `fix(run): let a second stop signal kill a run still judging its held cases`
**After:** Task 2

**Decision:** second-signal-kills

Go skills: `use-modern-go`, `go-coding-standards`, `go-test-quality`, `go-context`.

  - [x] **Step 1: Write `TestSecondInterruptKillsARun` (RED).** Re-exec the test binary as a child
    (env `FLOW_EVAL_SIGNAL_CHILD=1`; the child calls `stopContext()`, prints `ready`, waits on the
    context, prints `cancelled`, then blocks). The parent reads `ready`, sends SIGTERM, reads
    `cancelled`, sends SIGTERM again, and asserts the child exits by signal (`ProcessState` not
    success, `Signaled`) within 5 seconds. It fails to compile first (`stopContext` undefined).
  - [x] **Step 2: Add `stopContext`.** In `cmd/flow-eval/main.go` replace the run command's
    `signal.NotifyContext(context.Background(), stopSignals...)` (the one before `run.Run`) with
    `ctx, stop := stopContext()`, where `stopContext` wraps `signal.NotifyContext` and starts a
    goroutine that calls `stop()` once `ctx.Done()` fires, so the signals revert to their default
    action and a second one terminates the process; its comment says the run then loses its held
    lines, the trade-off of any kill. The other two `NotifyContext` calls are unchanged.
  - [x] **Step 3: Verify.**

```bash verified:the project's lint list (.flow/project.md) narrowed to this task's package
gofmt -w cmd/flow-eval
test -z "$(gofmt -l cmd/flow-eval)"
go vet ./cmd/flow-eval/
go test ./cmd/flow-eval/ -run 'TestSecondInterruptKillsARun|TestSIGTERMCancelsARun' -race -count=1
```
