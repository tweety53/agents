# self-review-kan-919-flow-improvement-port-the-dev-stack-and-project

## Why

kan-919's self-review pass found five pipeline defects, none `big`:

- **F-A (flow-fix).** `stats/internal/guard/check_dev_stack_fresh_test.go`'s case 3 races: the
  httptest handler (line 71) reads the captured `served` string while the test goroutine rewrites
  it (line 75). The request travels through a `curl` subprocess, so nothing orders the two
  accesses for the race detector. `go test ./internal/guard -race -count=1` failed one run in three
  on the whole package, and `go test ./internal/guard -race -run 'TestCheckDevStackFresh$' -count=1`
  fails every time.
- **F-B (flow-fix).** `scripts/run-verify-list.sh` writes each command's output to a `mktemp`
  directory removed on exit and reports only its last 20 lines, so a `-race` failure's stacks — at <!-- measured: sed -n '131,133p' scripts/run-verify-list.sh @ 9468cc05 → tail -n 20 -->
  the top of the output — are gone once the run ends.
- **F-C (flow-fix).** **Inline verify — a failing command** (`skills/flow/verify-and-handoff.md`)
  gives every failing command one inline re-run as the environmental-flake case, so a data race
  that passes its re-run is waived — the tolerance-widening `rules/fix-determinism-at-the-source.mdc`
  forbids.
- **F-D (flow-automation).** `_extract_backtick_tokens` (`scripts/check-task-commit-fields.py`)
  returns a value with no backticks as one token, so `check-plan-shape.sh` passes an unbackticked
  multi-path `**Files:**` line — kan-919's plan (75ded465) with its backticks stripped exits 0 — and
  the commit guard later reads the whole list as one path.
- **F-E (flow-automation).** `declaresMutation` (`stats/internal/guard/runreproducer.go`) matches
  only the exact line `# mutation-reproducer`; a near-miss such as `# mutation-reproducer: …` runs
  silently under the generic convention, its exit status read inverted from what its author meant.

## What changes

- Case 3 of `TestCheckDevStackFresh` passes under `-race`.
- A failing verify-list command's full output is kept under the worktree's git dir and its
  `## Report` block names the path.
- A failing command whose output carries a Go race banner takes **The loop** whatever its re-run
  exits, and `run-verify-list.sh` marks its block.
- `check-plan-shape.sh` refuses (F16) a `**Files:**` or `**Allowed-collateral:**` value with no
  backticked path that carries a comma or whitespace.
- `run-reproducer.sh` refuses (exit 2, never run) a reproducer one of whose first 10 lines opens <!-- measured: grep -n 'headLines(path, 10)' stats/internal/guard/runreproducer.go @ 9468cc05 → line 502 -->
  with the mutation marker but is not exactly it.
