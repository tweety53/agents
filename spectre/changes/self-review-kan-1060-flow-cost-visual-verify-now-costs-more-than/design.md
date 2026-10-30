## Context

Found by the self-review pass of `/flow` change `kan-1060-flow-cost-visual-verify-now-costs-more-than`
(`.superpowers/sdd/self-review-pass.md` in this worktree); the fixes ride that change's own branch
`spectre/kan-1060-flow-cost-visual-verify-now-costs-more-than`, based at its archive commit. The brainstorm
design is `.superpowers/sdd/2026-10-30-self-review-kan-1060-flow-cost-visual-verify-now-costs-more-than-design.md`.

**Finding 1 — loop landings run no visual suite (flow-fix).** `pfVerify`
(`stats/internal/guard/pipelinefix.go`) runs the commands `pfChecks` derives from the touched
paths — `## lint` lines, `go test`, script harnesses, vitest — and nothing from
`## visual verification`. `pfLand` always runs `pfVerify` before pushing, so a red visual suite
would already refuse a landing once verify runs it. `run-verify-list.sh` already resolves the row
a `test` list ends with, in `rvlVisualVerify` (`stats/internal/guard/runverifylist.go`): the
declared `verify` cell, run only when the diff from the base branch touches a `ui paths` glob
(`visualTrigger`), from the `regression checkout`'s worktree when one is declared, with a `specs`
row's paths put in `<specs>`.

**Finding 2 — the on-top fix commit is never field-checked (flow-fix).**
`skills/flow/gated-review-fix.md` says the guard's re-run on the on-top route "covers nothing".
`check-task-commit-fields.sh` already judges an on-top fix commit together with every commit
carrying the task's `Task-Id:` trailer (KAN-976, case 159 of
`stats/internal/guard/check_task_commit_fields_test.go`), so the command exists and only the
contract omits it. Run on task 12's on-top fix commit of the kan-1060 run (key
`task-12-implementer-fix-1`), against the plan as it stood, it refuses:
<!-- measured: scripts/check-task-commit-fields.sh <scratch worktree at that fix commit> 12 <that fix commit> "" <scratch worktree> kan-1060-flow-cost-visual-verify-now-costs-more-than → exit 1, "task 12: file stats/internal/guard/visual_verify_pass_close_test.go is not declared in Files: and not covered by Allowed-collateral:" @ 2026-10-30, the guard of this worktree -->

**Finding 3 — zsh mangles `--app` (flow-fix).** `zsh -c 'W=/a/b; echo "$W:stats/web=http://x"'`
prints `/s/web=h/btp://x`; `"${W}:stats/web=http://x"` prints `/a/b:stats/web=http://x`.
<!-- measured: both zsh -c commands above @ 2026-10-30, zsh on darwin 25.5.0 -->
The mangled value has no `=` after its first `:`, so `prepare`'s `--app` parse
(`stats/internal/guard/visualverifypass.go`) exits 2 with `--app '…' is not
<worktree>:<app-root>=<url>`.

## Decisions

### Where the loop's visual suite is resolved

**ID:** loop-verify-reuses-visual-row
**Status:** active
**Chosen:** `pfVerify`, after its other commands, calls `rvlVisualVerify(env, wt, false)` and runs what it returns from the directory it returns, printing a `green:` or `red:` line like every other command: a `specs` row is run first and its paths fill `<specs>` (a non-zero exit is `red:`, no path mapped runs no visual command), and the no-merge-base answer (`cmd` empty) is `red:`. `land` needs no change: it already refuses on any red verify line.
**Considered:** matching `ui paths` against `pfTouched`'s list and running the `verify` cell from the worktree — simpler, but it ignores a `regression checkout` and a `specs` row, so a gymie own-branch self-review would run the wrong suite from the wrong directory; a separate `visual` subcommand the parent calls before `land` — a second step the parent can skip, which is the class of the defect.

### The change's own visual skip

**ID:** loop-verify-honours-decision-skip
**Status:** active
**Chosen:** `rvlVisualVerify`'s existing `decision.json` `visual.verify` `skipped` read applies unchanged; a fresh `self-review-<name>` loop worktree holds no `decision.json`, and an own-branch loop inherits the change's own judgement, as `run-verify-list.sh` does.
**Considered:** a flag making `pfVerify` ignore the skip — a second resolution path for one rare case, with no evidence it occurred.

### Stale baselines already landed

**ID:** no-baseline-repair
**Status:** active
**Chosen:** nothing to repair: `dbb6b4f1` and the kan-1060 run's integrate (`Flow-Key: full-suite-fix-5`) already re-recorded the baselines `084f060f` left stale.
**Considered:** none.

### How the on-top fix commit is checked

**ID:** on-top-fix-runs-fields-guard
**Status:** active
**Chosen:** `skills/flow/gated-review-fix.md` has the parent run `check-task-commit-fields.sh <canonical> <id> <fix-sha> "" <canonical> <name>` on the on-top fix commit before the re-review is dispatched, a refusal handled as **The record carries its own corrections** (`skills/flow/implement.md`) states; case 159b of `check_task_commit_fields_test.go` pins that the guard refuses an on-top fix commit adding an undeclared path.
**Considered:** re-running `close-task.sh` without the `-end-*` flags, as the pass proposed — it also runs the review gate, ticks and pushes, steps that belong after the re-review comes back clean; a new guard refusing a reviewer re-dispatch until the check ran — a mechanism for a step one sentence states, with no second caller.

### The `--app` hint

**ID:** app-usage-names-zsh
**Status:** active
**Chosen:** the `--app` usage line always appends `— under zsh write "${W}:<app-root>=<url>"; "$W:s…" applies zsh's :s modifier`; the line fires only on a malformed value, so no condition on the value's shape is needed.
**Considered:** appending the hint only when the value looks mangled (no `=` after the first `:`, or a worktree that is not an existing directory) — a heuristic for a message that costs nothing to always show; rewriting the contract's `--app <worktree>:…` shape in `skills/flow/verify-and-handoff.md` — the placeholder is not a shell variable, and the error is where a mangled run lands.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 1 changes what `pipeline-fix.sh land` verifies before every loop landing push, across projects with a `regression checkout` and a `specs` row.
**Considered:** `flow-low` — Tasks 2 and 3 are a sentence and an error string, but a mis-resolved visual row in Task 1 would refuse or skip every UI loop landing.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** one fixer dispatch for Tasks 1, 2 and 3 — three file-disjoint tasks of a few dozen lines each, the largest being Task 1's.
**Considered:** one fixer per task — two extra dispatches for an error string and a sentence.

## Open questions

## Recommended approach

Fix each defect where every caller routes through it: `pfVerify`, which `land` already calls,
reuses `run-verify-list.sh`'s visual-row resolution (`loop-verify-reuses-visual-row`); the gated
fix contract names the field guard that already handles on-top commits
(`on-top-fix-runs-fields-guard`); the `--app` usage line names the zsh cause
(`app-usage-names-zsh`).

## Critical files

- `stats/internal/guard/pipelinefix.go` — `pfVerify`, `pfChecks`.
- `stats/internal/guard/pipeline_fix_test.go` — `pfNew` fixture, verify and land tests.
- `scripts/pipeline-fix.sh` — the `verify` subcommand's contract header.
- `skills/flow/gated-review-fix.md` — the on-top route's guard re-run.
- `stats/internal/guard/check_task_commit_fields_test.go` — case 159, the on-top fixture.
- `stats/internal/guard/visualverifypass.go`, `visual_verify_pass_prepare_test.go` — the `--app` parse and `TestVisualVerifyPassUsage`.

## Existing code to reuse

- `rvlVisualVerify` (`stats/internal/guard/runverifylist.go`) — the visual row, its directory and its `specs` row.
- `kwRun` and `pfVerify`'s own `green:`/`red:` printing — one closure both the lint/test commands and the visual commands use.
- `tcfPlans["159"]` and case 159's fixture — case 159b reuses both.

## Verification

- `cd stats && make -s web-dist && go test -count=1 ./internal/guard/ -run 'TestPipelineFix|TestCheckTaskCommitFields|TestVisualVerifyPass'` passes; `go vet ./...` and `gofmt -l .` are clean.
- `scripts/run-verify-list.sh <worktree> lint` is green, `check-verbatim-moves.sh` acknowledging this change's `verbatim-moves.txt`.

## Live check

None — the changes are a verify command's selection logic, a contract sentence and an error string, each exercised by its own test against fixtures; no running service or stored state changes.
