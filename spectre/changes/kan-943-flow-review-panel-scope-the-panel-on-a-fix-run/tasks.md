# kan-943-flow-review-panel-scope-the-panel-on-a-fix-run

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

- [x] 1. Guard: `check-late-fix-trigger` exit 3, the append-scope verdict
  - [x] **Step 1: In `stats/internal/guard/check_late_fix_trigger_test.go`, change the five condition-3/condition-4 cases (`condition 3 — over 40 lines`, `condition 3 — summed across worktrees`, `condition 3 — a binary numstat entry`, `condition 4 — a new task line`, `condition 4 — an untracked tasks.md`) to expect exit 3 with stdout exactly one line starting `append scope: ` (relabel them `append scope — over 40 lines`, `append scope — summed across worktrees`, `append scope — a binary numstat entry`, `append scope — a new task line`, `append scope — an untracked tasks.md`; the summed case's line reads `append scope: 41 changed lines since $SINCE`). Add `full path — a new task line with a moved base` (setup appends `- [ ] 2. a second task` to tasks.md, base verdict `MOVED:`) expecting exit 1 with two lines, the first `full path: condition 2 — `. Extend the result switch with a `tc.code == 3` arm (exactly one stdout line with the wanted prefix) and let the exit-1 arm accept a `lines int` field (default 1). Run `cd stats && go test ./internal/guard/ -run 'TestCheckLateFixTrigger' -count=1` and confirm the new expectations FAIL.**
<!-- measured: grep -n "lftMaxLines = 40" stats/internal/guard/latefixtrigger.go @ 3d108317 -->
  - [x] **Step 2: In `stats/internal/guard/latefixtrigger.go`, track the failed condition numbers; when every failed condition is 3 or 4, print `append scope: <lines> changed lines since <canonical since-close sha>` and return 3 (the binary case prints the counted lines, binary entries excluded); otherwise keep exit 1 with every failed line. Update the function comment's exit list.**
  - [x] **Step 3: In `scripts/check-late-fix-trigger.sh`, add to the header's stdout list `exit 3  append scope: <n> changed lines since <canonical since-close sha>` and to the exit codes `3  append scope — conditions 1, 2 and 5 hold and only 3 and/or 4 failed: pass 1 runs the decided roster on the since-close delta (skills/flow/review-panel-late-fix.md, The append scope)`.**
  - [x] **Step 4: Verify: `cd stats && go test ./internal/guard/ -run 'TestCheckLateFixTrigger' -count=1 -race && gofmt -l . && go vet ./internal/guard/` and `scripts/check-guard-symlinks.sh`.**
**Build:** green
**Files:** `stats/internal/guard/latefixtrigger.go`, `stats/internal/guard/check_late_fix_trigger_test.go`, `scripts/check-late-fix-trigger.sh`
**Tests:** `append scope — over 40 lines`, `append scope — summed across worktrees`, `append scope — a binary numstat entry`, `append scope — a new task line`, `append scope — an untracked tasks.md`, `full path — a new task line with a moved base`
<!-- measured: grep -n "lftMaxLines = 40" stats/internal/guard/latefixtrigger.go @ 3d108317 -->
**Regression:** reverting the commit returns exit 1 for a condition-3/4-only failure: every `append scope — …` case fails on the exit code; `full path — a new task line with a moved base` fails if the guard ever returned 3 while condition 2 also failed.
**Baseline:** before=1 after=1
**Commit:** feat(late-fix-trigger): add the append-scope verdict
**After:** none

**Decision:** append-scope-exit-3

**Decision:** append-scope-any-fix-run

- [x] 2. Contract: the append scope in the panel docs
  - [x] **Step 1: `skills/flow/review-panel-late-fix.md` — after the exit list (exit 0/1/2), add exit 3 → **The append scope** (read as the full path for the roster, never for the read scope), and add a section `## The append scope`: on exit 3, pass 1 runs the decided roster unchanged — the same `panel.dispatches`, pairs and ceilings as a full pass 1 (a `default` panel its settings-store grouping) — every dispatch reading `<abs-worktree>/.superpowers/sdd/late-fix.diff` written over the since-close range by the same `write-panel-diff.sh late-fix` call and rendered with `-diff late-fix`; every entry check still runs (base movement, diff-size cap — measured on the full branch as today —, docs-only guard, whose reduction to `primary` still applies, on the same delta); recorded with `flow record pass -round <round>`: `append scope: <n> changed lines since <sha>`; a Critical or Important it raises feeds the ordinary fix-round loop and voids nothing; a clean close moves the close sha to the round's HEAD as any close does.**
  - [x] **Step 2: `skills/flow/document-fix.md` — replace "the append never narrows the panel" and the sentence after it with: the append is panel-checked by the full decided roster — its read scope is the since-close delta under **The append scope** (`skills/flow/review-panel-late-fix.md`), and the roster-narrowing late-fix reduction stays closed to an append. Keep the MUTATION PROOF sentence unchanged.**
  - [x] **Step 3: `skills/flow/review-panel.md` — the `[DIFF_PATH]` row reads `late-fix.diff` "under the late-fix reduction or the append scope"; the "only automatic reductions" paragraph names the append scope as a read-scope narrowing that removes no slot. `skills/flow/review-panel-fix-round.md` pass-1 sentence adds "or the decided roster on the since-close delta under the append scope". `skills/flow/review-panel-experimental-slot.md` — the experimental slot rides an append-scope pass 1 as it rides a full one (it is in the decided roster).**
  - [x] **Step 4: Verify: `scripts/check-references.sh && scripts/check-markdown-integrity.py && scripts/check-plan-shape.sh spectre/changes/kan-943-flow-review-panel-scope-the-panel-on-a-fix-run/tasks.md`.**
**Build:** green
**Files:** `skills/flow/review-panel-late-fix.md`, `skills/flow/document-fix.md`, `skills/flow/review-panel.md`, `skills/flow/review-panel-fix-round.md`, `skills/flow/review-panel-experimental-slot.md`, `skills/flow/verify-fix-loop.md`
**Tests:** **none**
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(review-panel): scope a fix run's pass 1 to the since-close delta
**After:** Task 1

**Decision:** append-scope-since-close-full-roster

**Decision:** append-scope-reuse-late-fix-diff

**Decision:** append-scope-no-void

Correction (2026-10-08): the plan named five files; the drift check found `skills/flow/verify-fix-loop.md` step 3 stating the scope-growth condition "refuses every appended task", stale once that condition yields exit 3 — reworded to "denies every appended task the `primary`-alone reduction" in this task's commit, and `**Files:**` widened. `skills/flow/review-panel.md`'s late-fix load directive also names the append scope among what the loaded file carries.
