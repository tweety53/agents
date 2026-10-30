# self-review-kan-1060-flow-cost-visual-verify-now-costs-more-than

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the three non-`big` findings of `kan-1060-flow-cost-visual-verify-now-costs-more-than`'s self-review pass.

**Architecture:** `pfVerify` reuses `run-verify-list.sh`'s visual-row resolution; the gated fix
contract names the field guard for the on-top fix commit; the `--app` usage line names the zsh cause.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-30-self-review-kan-1060-flow-cost-visual-verify-now-costs-more-than-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Work in the worktree `/Users/tweety53/Projects/agents-worktrees/kan-1060-flow-cost-visual-verify-now-costs-more-than`, branch `spectre/kan-1060-flow-cost-visual-verify-now-costs-more-than`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- The prose edit lands character for character through the script below, saved to a scratch file outside the worktree and run as `python3 -I <saved script> "$PWD" 2` from the worktree's root; it asserts its anchor matches exactly once. `scripts/check-verbatim-moves.sh` matches the result against `verbatim-moves.txt` — a hand edit that drifts by one character fails it. A run that asserts is reported, never hand-merged.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

```python verified:applied with task number 2 to this worktree at the plan's base; scripts/check-verbatim-moves.sh then printed exactly the one FAIL sentence verbatim-moves.txt lists, then reverted
import sys
root = sys.argv[1]
def sub(path, old, new):
    p = f"{root}/{path}"
    s = open(p).read()
    assert s.count(old) == 1, (path, old[:60], s.count(old))
    open(p, "w").write(s.replace(old, new))
which = sys.argv[2:]
if "2" in which:
    sub("skills/flow/gated-review-fix.md",
        "The\nparent re-runs the guard on every sha the fold's rebase rewrote — the on-top route rewrites none, so\nits re-run covers nothing — then re-dispatches the reviewer — one\n",
        "The\nparent re-runs the guard on every sha the fold's rebase rewrote, and on the on-top route on its fix\ncommit, `check-task-commit-fields.sh <canonical> <id> <fix-sha> \"\" <canonical> <name>`, which\njoins every commit carrying the task's `Task-Id:` trailer — a refusal handled as **The record\ncarries its own corrections** (`skills/flow/implement.md`) states — then re-dispatches the reviewer — one\n")
```

## Review Focus

- Task 1: the visual row runs only when the diff touches `ui paths`, from the directory `rvlVisualVerify` returns, with a `specs` row's paths in `<specs>`; a red visual line makes `land` refuse before any push.
- Task 2: the contract's command matches `check-task-commit-fields.sh`'s argument order, and case 159b fails on an undeclared path in the on-top commit alone.
- Task 3: the hint reaches the operator on every malformed `--app`, the zsh-mangled value included.

- [ ] 1. Run the visual suite in a loop's verify when the diff touches ui paths
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Add `TestPipelineFixVerifyRunsVisualOnUIPath` to `stats/internal/guard/pipeline_fix_test.go`. After `fx.pf(t, "start", "kan-1-demo", "6")` (rows from `lpRowsJSON("pipeline-fix-6-", lpBranch, "plan completed", "fix completed", "review-1 clean")`, as `TestPipelineFixLand` builds them), the loop worktree appends to `.flow/project.md` a `## visual verification` section whose settings table declares `` `ui paths` `` `` `web/src/**` `` and whose command table declares `` `verify` `` `` `test ! -e .visual-red` ``, and commits it. Cases, in order: (a) with only `docs/a.md` touched, `verify` exits 0 and prints no line naming `test ! -e .visual-red`; (b) after committing `web/src/a.ts`, `verify` exits 0 and prints `green: test ! -e .visual-red`; (c) with an untracked `.visual-red` in the worktree, `verify` exits 1 and prints `red: test ! -e .visual-red (exit 1)`; (d) `land` then exits 1 naming that red line, and origin's `main` log carries no commit of the branch. Run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestPipelineFixVerifyRunsVisualOnUIPath' -count=1`: it fails on case (b) before the fix.
    <!-- unverified: rvlVisualVerify resolves the base branch through resolve-base-branch in pfNew's fixture (an origin remote with origin/HEAD set) — if it cannot, its visualTrigger fallback runs the row on every case; the implementer confirms case (a) prints no visual line and, if not, makes the fixture resolvable rather than weakening case (a) -->
  - [ ] **Step 2: Fix `pfVerify`** (`stats/internal/guard/pipelinefix.go`): factor the `green:`/`red:` printing of the command loop into one closure `report(cmd, dir string)` running `kwRun(env, dir, "bash", "-c", cmd)`. After the loop, `if cmd, specs, dir, ok := rvlVisualVerify(env, wt, false); ok`: when `cmd` is empty print `red: visual verify — no merge base to scope the specs row to` and count it red; when `specs` is non-empty run it in `dir` — a non-zero exit is a `red:` line for `specs` (its last 40 lines), no path printed runs no visual command and prints `green: <specs> (mapped no spec)`, else `<specs>` in `cmd` is replaced by the space-joined paths; then `report(cmd, dir)`. Extend `pfChecks`' and `pfVerify`'s doc comments with one sentence each naming the visual row, and the `verify` paragraph of `scripts/pipeline-fix.sh`'s header with: "When the diff from the base branch touches a `ui paths` glob of `## visual verification`, the section's `verify` row runs last, resolved as run-verify-list.sh resolves it (from a declared `regression checkout`'s worktree, a `specs` row's paths in `<specs>`); a red one makes `land` refuse."
    <!-- measured: grep -n 'len(lines) > 40' stats/internal/guard/pipelinefix.go @ plan base — the existing red-line tail is 40 lines -->
  - [ ] **Step 3: Verify** — `gofmt -w` the Go files; `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestPipelineFix' -count=1` passes; `cd stats && make -s web-dist && go vet ./internal/guard/ && gofmt -l .` is clean; `scripts/check-guard-symlinks.sh` and `scripts/check-references.sh` exit 0.

**Build:** green
**Files:** `stats/internal/guard/pipelinefix.go`, `stats/internal/guard/pipeline_fix_test.go`, `scripts/pipeline-fix.sh`
**Tests:** `TestPipelineFixVerifyRunsVisualOnUIPath`
**Regression:** `TestPipelineFixVerifyRunsVisualOnUIPath` — reverting the commit lets `verify` print no visual line for a `ui paths` diff and `land` push a branch whose visual suite is red.
**Baseline:** before=13 after=14
<!-- measured: grep -cE '^func Test' stats/internal/guard/pipeline_fix_test.go @ plan base (13) -->
**Commit:** fix(stats): run the visual suite in a loop's verify when the diff touches ui paths
**After:** none

**Decision:** loop-verify-reuses-visual-row
**Decision:** loop-verify-honours-decision-skip
**Decision:** no-baseline-repair

- [ ] 2. Field-check the gated reviewer's on-top fix commit
  - [ ] **Step 1: Pin the guard's premise.** Add case `159b` to `TestCheckTaskCommitFields` in `stats/internal/guard/check_task_commit_fields_test.go`, beside case 159 and on `tcfPlans["159"]`: the same history, but the on-top fix commit (`Task-Id: 1`) also adds `gamma2.txt`; `r.run("1", fix)` exits 1 and names `file gamma2.txt is not declared in Files:`. It passes on the untouched guard — it pins the behaviour the contract now relies on (`design.md`, Context, Finding 2's measurement).
  - [ ] **Step 2: Prose** — run the Global constraints script with task number `2`.
  - [ ] **Step 3: Verify** — `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckTaskCommitFields' -count=1` passes; `gofmt -l stats/internal/guard` prints nothing; `scripts/run-verify-list.sh /Users/tweety53/Projects/agents-worktrees/kan-1060-flow-cost-visual-verify-now-costs-more-than lint` is green.

**Build:** green
**Files:** `skills/flow/gated-review-fix.md`, `stats/internal/guard/check_task_commit_fields_test.go`
**Tests:** `Case 159b`
**Regression:** `Case 159b` — a guard change that judged an on-top fix commit without its paths against `**Files:**` fails it; the contract sentence itself is held by `check-verbatim-moves.sh` against `verbatim-moves.txt`.
**Baseline:** before=144 after=145
<!-- measured: grep -cE '^\s+\{"case [0-9]+[a-z]?",' stats/internal/guard/check_task_commit_fields_test.go @ plan base (144) -->
**Commit:** fix(flow): field-check the gated reviewer's on-top fix commit
**After:** none

**Decision:** on-top-fix-runs-fields-guard

- [ ] 3. Name the zsh cause in visual-verify-pass's --app usage error
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `TestVisualVerifyPassUsage` (`stats/internal/guard/visual_verify_pass_prepare_test.go`), give the case table a `hint string` field, add the case `{"a zsh-mangled app", …, "--app", "/s/web=h/btp://x"}` — the value `zsh -c 'W=/a/b; echo "$W:stats/web=http://x"'` prints — and give it and the existing "an app not <wt>:<root>=<url>" case `hint: "${W}"`; when `hint` is set, stderr must contain it. Run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestVisualVerifyPassUsage' -count=1`: both app cases fail before the fix.
  - [ ] **Step 2: Fix** — in `stats/internal/guard/visualverifypass.go`, the `--app` usage line becomes `"--app '" + a + "' is not <worktree>:<app-root>=<url> — under zsh write \"${W}:<app-root>=<url>\"; \"$W:s…\" applies zsh's :s modifier"`.
  - [ ] **Step 3: Verify** — `gofmt -w` the Go files; the test command above passes; `cd stats && make -s web-dist && go vet ./internal/guard/ && gofmt -l .` is clean.

**Build:** green
**Files:** `stats/internal/guard/visualverifypass.go`, `stats/internal/guard/visual_verify_pass_prepare_test.go`
**Tests:** `TestVisualVerifyPassUsage`
<!-- predicted: the edit lies inside TestVisualVerifyPassUsage, so git diff's default hunk header names the function and the guard finds the name in the diff -->
**Regression:** `TestVisualVerifyPassUsage` — reverting the commit drops `${W}` from the `--app` usage line, failing both app cases.
**Commit:** fix(stats): name the zsh cause in visual-verify-pass's --app usage error
**After:** none

**Decision:** app-usage-names-zsh
