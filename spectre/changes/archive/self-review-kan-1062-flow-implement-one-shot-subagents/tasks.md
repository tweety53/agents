# self-review-kan-1062-flow-implement-one-shot-subagents

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the four non-`big` findings (1-B, 1-C, 1-D, 1-E) of `kan-1062-flow-implement-one-shot-subagents`'s self-review pass.

**Architecture:** the REPRODUCER paragraph names its file in the contract's own placeholder shape;
`flow stage wrap` kills its child's process group; the self-review bundle gains a gated-findings
source; `live-store-copy.sh up` registers its scratch change.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-31-self-review-kan-1062-flow-implement-one-shot-subagents-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Work in the worktree `/Users/tweety53/Projects/agents-worktrees/kan-1062-flow-implement-one-shot-subagents`, branch `spectre/kan-1062-flow-implement-one-shot-subagents`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- The prose edit lands character for character through the script below, saved to a scratch file outside the worktree and run as `python3 -I <saved script> "$PWD" 1` from the worktree's root; it asserts its anchor matches exactly once. `scripts/check-verbatim-moves.sh` matches the result against `verbatim-moves.txt` — a hand edit that drifts by one character fails it. A run that asserts is reported, never hand-merged.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

```python verified:applied with task number 1 to this worktree at the plan's base; scripts/check-verbatim-moves.sh then printed only "ok   acknowledged" lines for the two sentences verbatim-moves.txt lists, then reverted
import sys
root = sys.argv[1]
def sub(path, old, new):
    p = f"{root}/{path}"
    s = open(p).read()
    assert s.count(old) == 1, (path, old[:60], s.count(old))
    open(p, "w").write(s.replace(old, new))
which = sys.argv[2:]
if "1" in which:
    sub("skills/flow/review-panel.md",
        "named `R-ID-N.sh` — `R` this round's number, `ID` your own resolved reviewer id, `N` your own 1-based finding index — and recorded",
        "named `<round>-<id>-<n>.sh` — `<round>` this round's number, `<id>` your own resolved reviewer id, `<n>` your own 1-based finding index, so round 0's first mutation finding is `0-mutation-1.sh` — and recorded")
```

## Review Focus

- Task 1: the rendered slot prompt names `<round>-<id>-<n>.sh` with the round stated, and carries no `R-ID-N`.
- Task 2: the group kill fires only on a cancelled context, reaches the TERM-ignoring grandchild, and leaves a normally exiting child's background work alone; the whole package leaves no survivor under `run-reproducer.sh`.
- Task 3: the gated rows reach the bundle and stay out of the panel record; a run without them reports the source absent.
- Task 4: `up` refuses a pre-existing scratch directory before any side effect, and `down` removes only the directory `up` created.

- [x] 1. Name the REPRODUCER paragraph's file in the contract's placeholder shape
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `stats/internal/guard/render_slot_prompt_test.go`, change `TestRenderSlotPromptStatesTheRound` to require ``"`<round>` this round's number — `3` here"`` and ``"`<round>-<id>-<n>.sh`"`` in the rendered prompt and to fail when it contains `R-ID-N`; change `TestRenderSlotPromptCarriesReproducer`'s expected block to replace ``"`<round>` this round's number"`` with ``"`<round>` this round's number — `0` here"``. Run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestRenderSlotPrompt' -count=1`: both tests fail before the fix.
  - [x] **Step 2: Prose** — run the Global constraints script with task number `1`.
  - [x] **Step 3: Fix the renderer** — in `stats/internal/guard/renderslotprompt.go`, the REPRODUCER phrase constant becomes ``"`<round>` this round's number"``; the append (`" — `"+round+"` here"`) and the refusal message stay as they are.
  - [x] **Step 4: Verify** — `gofmt -w` the Go files; the test command above passes; `cd stats && make -s web-dist && go vet ./internal/guard/ && gofmt -l .` is clean; `scripts/run-verify-list.sh /Users/tweety53/Projects/agents-worktrees/kan-1062-flow-implement-one-shot-subagents lint` is green.

**Build:** green
**Files:** `skills/flow/review-panel.md`, `stats/internal/guard/renderslotprompt.go`, `stats/internal/guard/render_slot_prompt_test.go`
**Tests:** `TestRenderSlotPromptStatesTheRound`
<!-- predicted: the edit lies inside TestRenderSlotPromptStatesTheRound, so git diff's default hunk header names the function and the guard finds the name in the diff -->
**Regression:** `TestRenderSlotPromptStatesTheRound` — reverting the commit restores `R-ID-N.sh` and the `R` phrase, so the round is no longer stated after `<round>` and the test fails.
**Commit:** fix(flow): name the REPRODUCER paragraph's file <round>-<id>-<n>.sh
**After:** none

**Decision:** reproducer-name-placeholders

- [x] 2. Kill stage wrap's whole process group on cancel
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `stats/cmd/flow/stage_test.go`, change `TestRunStageWrapContextCancelForceKillsChildIgnoringTerm`'s child to `"sh", "-c", "trap \"\" TERM; sleep 30 & echo $! >\"$0\"; wait", <pidfile under t.TempDir()>`; after `run` returns, read the pid, register `t.Cleanup` sending it SIGKILL, and poll `syscall.Kill(pid, 0)` every 50ms for up to 2s: the test fails when the grandchild is still alive. Update its doc comment to name the grandchild. Run `cd stats && make -s web-dist && go test ./cmd/flow/ -run 'TestRunStageWrapContextCancel' -count=1`: it fails before the fix.
  - [x] **Step 2: Fix `runStageWrap`** (`stats/cmd/flow/stage.go`) per decision `stage-wrap-kills-group`: `cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}`; `cmd.Cancel` returns `syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)`; after `cmd.Run()`, when `ctx.Err() != nil && cmd.Process != nil`, `_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)`. Extend the existing comment above `exec.CommandContext` with why the group, and why only on cancel.
  - [x] **Step 3: Verify** — `gofmt -w` the Go files; `cd stats && make -s web-dist && go test ./cmd/flow/ -run 'TestRunStageWrap' -count=1` passes; `cd stats && make -s web-dist && go vet ./cmd/flow/ && gofmt -l .` is clean; `scripts/run-reproducer.sh /Users/tweety53/Projects/agents-worktrees/kan-1062-flow-implement-one-shot-subagents .superpowers/sdd/reproducers/sr-1c-whole-package.sh` exits 1 (ran, exited 0) and names no surviving pid — exit 3 means a survivor remains, traced and fixed in this task.
    <!-- measured: the same run-reproducer.sh command → exit 3, "surviving process pid(s): 78536" @ 2026-10-31, this worktree at the plan base -->

**Build:** green
**Files:** `stats/cmd/flow/stage.go`, `stats/cmd/flow/stage_test.go`
**Tests:** `TestRunStageWrapContextCancelForceKillsChildIgnoringTerm`
<!-- predicted: the edit lies inside TestRunStageWrapContextCancelForceKillsChildIgnoringTerm, so git diff's default hunk header names the function and the guard finds the name in the diff -->
**Regression:** `TestRunStageWrapContextCancelForceKillsChildIgnoringTerm` — reverting the commit leaves the TERM-ignoring `sleep` alive past the 2s poll.
**Commit:** fix(stats): kill stage wrap's whole process group on cancel
**After:** none

**Decision:** stage-wrap-kills-group

- [x] 3. Carry the gated per-task reviewers' findings in the self-review bundle
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Add `TestBundleServesGatedReviewerFindings` to `stats/internal/selfreview/bundle_test.go`: a `records.Run` whose `Findings` are `{Ref: "F1", Slot: records.GatedReviewerSlot, TaskID: "6", Severity: "Important", Location: "skills/flow/implement.md:12", Note: "no Model: line", Status: "fixed"}` and `{Ref: "F2", Slot: "primary", Severity: "Minor", Note: "panel row", Status: "fixed"}`; the bundle contains `## gated per-task reviewer findings` and `| F1 | 6 | Important |`, the text after that heading carries no `F2`, and the panel section carries no `F1`. In `TestBundleAssemblySkipsAbsentSources` also require `skipped: gated per-task reviewer findings (absent)`. Every `of 9 sources` / `9 of 9 sources` expectation in the file becomes `of 10` / `10 of 10`, and `found: N` counts stay as they are. Run `cd stats && make -s web-dist && go test ./internal/selfreview/ -run 'TestBundle' -count=1`: it fails before the fix.
    <!-- measured: grep -c "of 9 sources" stats/internal/selfreview/bundle_test.go → 6 @ plan base -->
  - [x] **Step 2: Fix** — add `func RenderGatedFindings(r Run) (string, bool)` to `stats/internal/records/render.go` per decision `gated-findings-own-source`: the header row `| ID | Task | Severity | Location | Note | Status |`, its separator, one row per `GatedReviewerSlot` finding through `tableCell`, false with no such row. In `Bundle` (`stats/internal/selfreview/git.go`), right after the panel source, `gated, ok := records.RenderGatedFindings(run)` and `add("gated per-task reviewer findings", gated, ok)`; extend `Bundle`'s doc comment with one sentence naming the source.
  - [x] **Step 3: Verify** — `gofmt -w` the Go files; `cd stats && make -s web-dist && go test ./internal/selfreview/ -run 'TestBundle' -count=1` passes; `cd stats && make -s web-dist && go vet ./internal/selfreview/ ./internal/records/ && gofmt -l .` is clean.

**Build:** green
**Files:** `stats/internal/records/render.go`, `stats/internal/selfreview/git.go`, `stats/internal/selfreview/bundle_test.go`
**Tests:** `TestBundleServesGatedReviewerFindings`
**Regression:** `TestBundleServesGatedReviewerFindings` — reverting the commit leaves the bundle with no gated section, so the gated row is invisible to the pass.
**Baseline:** before=34 after=35
<!-- measured: grep -cE '^func Test' stats/internal/selfreview/bundle_test.go @ plan base (34) -->
**Commit:** fix(stats): carry the gated per-task reviewers' findings in the self-review bundle
**After:** none

**Decision:** gated-findings-own-source

- [x] 4. Register a scratch change on the live-store copy
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `scripts/test-live-store-copy.sh`, add a `flow` stub on `PATH` that appends `flow $*`, `FLOW_ADDR=$FLOW_ADDR` and its stdin to `$LOG`, and add `.superpowers/sdd/$name` under the repository root to the `EXIT` trap's removal. Cases: (a) after `up "$name" "$port"`, the log carries `flow state set -C <repo root> $name`, `FLOW_ADDR=http://127.0.0.1:$port` and `"state":"STARTED"`, `.superpowers/sdd/$name` is a directory, and the output carries `scratch change: $name`; (b) after `down`, that directory is gone; (c) with `.superpowers/sdd/$name` created beforehand, `up` exits 2 and the log stays empty, then the directory is removed. Run `scripts/test-live-store-copy.sh`: it fails before the fix.
  - [x] **Step 2: Fix `scripts/live-store-copy.sh`** per decision `live-copy-registers-scratch-change`: with `sdd="$repo/.superpowers/sdd/$name"`, `up` refuses with exit 2 when `$sdd` exists, beside its other refusals; after `mkdir -p "$dir/…"` it creates `$sdd` and writes its path to `$dir/scratch`; once flowd answers it pipes `{"state":"STARTED","branch":"<git -C $repo branch --show-current>","worktrees":{"<repo>":"<git -C $repo rev-parse HEAD>"},"artifactUrl":null,"jiraIssue":null,"prUrl":null,"updatedBy":"/flow"}` into `FLOW_ADDR=http://127.0.0.1:$port flow state set -C "$repo" "$name"` (a failure: `live-store-copy: state record for $name failed` and exit 1), then prints `scratch change: $name` after the two existing lines. `down`, before `rm -rf "$dir"`, removes the path `$dir/scratch` names when that file exists. The header's `up`/`down` paragraphs gain one sentence each. In `.flow/project.md`'s live-check paragraph, after "printing the `FLOW_ADDR`/`FLOW_RECORDS_ADDR` exports that point a session at it", add "and registers a scratch change named `<name>` in the copy — a STARTED state record and `.superpowers/sdd/<name>/` in that checkout — so `flow record -change <name>` writes there".
    <!-- unverified: flow state set against the copy accepts pfRecord's JSON shape — pfStart sends the same body to the dev store; the harness stubs flow, so the implementer confirms the shape against stats/cmd/flow/state.go's decoder -->
  - [x] **Step 3: Verify** — `scripts/test-live-store-copy.sh` prints `all cases passed`; `shellcheck scripts/live-store-copy.sh scripts/test-live-store-copy.sh` is clean when installed; `scripts/check-references.sh` exits 0.

**Build:** green
**Files:** `scripts/live-store-copy.sh`, `scripts/test-live-store-copy.sh`, `.flow/project.md`
**Tests:** `test-live-store-copy.sh`
**Regression:** `test-live-store-copy.sh` — reverting the commit fails its case (a): no `flow state set` call and no scratch directory.
**Commit:** fix(scripts): register a scratch change on the live-store copy
**After:** none

**Decision:** live-copy-registers-scratch-change
