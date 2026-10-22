# self-review-kan-923-flow-improvement-port-the-python-backed-guards

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** close kan-923's five self-review findings — F-A (Task 1), F-B (Task 2), F-C (Task 3),
F-D (Task 4), F-E (Task 5), as `proposal.md` names them.

**Architecture:** three shims take the shared root helper; `check-plan-unchanged` compares content
and tolerates checkbox ticks; the shared field walk keeps a field's first occurrence; five view
tests await loaded content; panel diffs drop deleted bodies.

**Tech Stack:** Go (`stats/`), Bash, TypeScript + vitest (`stats/web/`), Markdown.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-22-self-review-kan-923-flow-improvement-port-the-python-backed-guards-design.md`
in this worktree.

## Global constraints

- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-923-flow-improvement-port-the-python-backed-guards` (`<wt>` below), on branch `spectre/kan-923-flow-improvement-port-the-python-backed-guards`; every `**Files:**` path is relative to it. Never touch the main checkout and never create a worktree.
- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`. Never merge or push.
- Go: `gofmt -w` every Go file you change before committing; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- Never widen a timeout, retry or tolerance to make a test pass (`rules/fix-determinism-at-the-source.mdc`).
- Never stop or restart `flowd` on `127.0.0.1:4173` nor the `flow-postgres` container.

## Review Focus

- Task 1: a shim reached through `skills/<skill>/scripts/` still scans the checkout's plans, never the skill directory's.
- Task 2: a tick plus a prose edit in one file still fails; a planted untracked file and a deleted plan file still fail; a tick on a step line passes.
- Task 3: a task whose real `**Files:**` precedes a prose line opening `**Files:**` keeps the real value in every guard that parses it.
- Task 4: each of the five sites fails under the delayed mock before the fix and passes after it.
- Task 5: a deleted file still appears in the panel diff and its `.touched` list; only its body is gone.

- [ ] 1. Derive the plan guards' root through flow_guard_root
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -c SELF_REAL scripts/check-plan-shape.sh scripts/check-task-build-green.sh scripts/plan-dispatch-bundles.sh` prints a non-zero count for each; `grep -n 'flow_guard_root()' scripts/lib/flow-guard.sh` names the helper.
    <!-- measured: grep -n 'flow_guard_root()' scripts/lib/flow-guard.sh @ 5611091f → line 80 -->
  - [ ] **Step 2: Implement** per decision `shim-root-flow-guard-root`. In each of the three shims, replace everything from the `# The root is this script's REAL parent's parent` comment through the `FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "$SELF_REAL")/.." && pwd)"` line with `FLOW_GUARD_REPO_ROOT="$(flow_guard_root)"` (the `export` line stays), and change the header's "the parent of this script's real directory" to name `flow_guard_root` (`scripts/lib/flow-guard.sh`) as what derives it.
  - [ ] **Step 3: Update the tests the shims' sandboxes rest on.** In `stats/internal/guard/check_plan_shape_test.go` (case 18, `psLink(t, psScripts+"/lib", sandbox+"/scripts/lib")`), `stats/internal/guard/check_task_build_green_test.go` (the "Beyond the harness" sandbox) and `stats/internal/guard/plan_dispatch_bundles_test.go` (case 11), replace the sandbox's `scripts/lib` symlink with `psCopy(t, psScripts+"/lib/flow-guard.sh", sandbox+"/scripts/lib/flow-guard.sh")` plus `psLink(t, filepath.Dir(psScripts)+"/stats", sandbox+"/stats")`; the `skills/flow/scripts/lib` symlink stays. In the "Unlanded archives" blocks of `check_plan_shape_test.go` and `check_task_build_green_test.go`, delete the `pin: a shim copy with no lib/resolve-file.sh exits 2 naming it` assertion and the `psCopy` of `lib/resolve-file.sh` after it.
  - [ ] **Step 4: Verify.** `grep -c SELF_REAL` over the three shims prints 0 for each; `cd stats && make -s web-dist && go test ./internal/guard -run 'TestCheckPlanShape$|TestCheckTaskBuildGreen$|TestPlanDispatchBundles$|TestCheckGuardSymlinks' -count=1` passes; from `/`, `bash <wt>/skills/flow/scripts/check-plan-shape.sh` and `bash <wt>/skills/flow/scripts/plan-dispatch-bundles.sh` print the same as `scripts/check-plan-shape.sh` and `scripts/plan-dispatch-bundles.sh` from `<wt>`; `scripts/check-guard-symlinks.sh` and `scripts/check-references.sh` exit 0. Commit.

**Build:** green
**Files:** `scripts/check-plan-shape.sh`, `scripts/check-task-build-green.sh`, `scripts/plan-dispatch-bundles.sh`, `stats/internal/guard/check_plan_shape_test.go`, `stats/internal/guard/check_task_build_green_test.go`, `stats/internal/guard/plan_dispatch_bundles_test.go`
**Tests:** none — repairs existing tests: `TestCheckPlanShape`, `TestCheckTaskBuildGreen`, `TestPlanDispatchBundles`
**Regression:** none — behaviour is unchanged; case 18 and case 11 fail if the root stops resolving to the checkout.
**Commit:** refactor(scripts): derive the plan guards' root through flow_guard_root
**After:** none

**Decision:** shim-root-flow-guard-root

- [ ] 2. Let check-plan-unchanged pass a task tick made during the flight
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Write the cases of Step 2 first and run `cd stats && go test ./internal/guard -run 'TestCheckPlanUnchanged$' -count=1`: case 17b fails (exit 1 where 0 is wanted).
  - [ ] **Step 2: Write the cases** in `stats/internal/guard/check_plan_unchanged_test.go`'s `cases`, after case 16. Case 17: an act writing `spectre/changes/demo/tasks.md` as `- [ ] 1. A\n  - [ ] **Step 1: do it**\n- [ ] 2. B\n` and committing it; `pucRun("snapshot", 0, "case 17a: snapshot of an unticked plan exits 0")`; an act rewriting it as `- [x] 1. A\n  - [x] **Step 1: do it**\n- [ ] 2. B\n` (uncommitted); `pucRun("verify", 0, "case 17b: verify after a task tick during the flight exits 0")`. Case 18: the same committed plan; snapshot; an act rewriting it as `- [x] 1. A edited\n  - [x] **Step 1: do it**\n- [ ] 2. B\n`; `pucRun("verify", 1, "case 18b: verify after a tick beside a prose edit exits 1")`; `pucNames("case 18c: names the edited line", "A edited")`. Update the file's header comment to name cases 17-18 as the tick tolerance.
  - [ ] **Step 3: Implement** per decision `plan-unchanged-content-ticks` in `stats/internal/guard/planunchanged.go`: `planStatus` becomes a content dump — `git -C <wt> ls-files -z --cached --others --exclude-standard -- spectre/changes/<name> spectre/changes/archive/<name>`, paths sorted and de-duplicated, each written as `== <path>\n` then its bytes (a final newline added when missing), or `== <path> (missing)\n` when absent on disk; a git failure is still `gitFailed()`. Verify: equal dumps print today's OK line; otherwise, when both split into the same number of lines and every differing pair satisfies `strings.HasPrefix(strings.TrimLeft(old, " "), "- [ ] ") && cur == strings.Replace(old, "- [ ] ", "- [x] ", 1)`, print `PLAN-UNCHANGED-OK: spectre/changes/<name> matches its pre-dispatch snapshot (checkbox ticks only)` and exit 0; anything else keeps today's exit-1 path and diff. Update the file's top comment, the snapshot success line's wording (`content pinned`), and `scripts/check-plan-unchanged.sh`'s header (the two mode lines and THE GUARD IS SCOPED paragraph) to say the snapshot is the plan tree's content and a checkbox tick is not a change.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCheckPlanUnchanged$' -count=1` passes, cases 1-16 unchanged; `scripts/check-references.sh` exits 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/planunchanged.go`, `stats/internal/guard/check_plan_unchanged_test.go`, `scripts/check-plan-unchanged.sh`
**Tests:** `case 17b: verify after a task tick during the flight exits 0`, `case 18b: verify after a tick beside a prose edit exits 1`
**Regression:** `case 17b: verify after a task tick during the flight exits 0` fails — the parent's tick during a reviewer's flight reads as the reviewer changing the plan again.
**Commit:** fix(guard): let check-plan-unchanged pass a task tick made during the flight
**After:** none

**Decision:** plan-unchanged-content-ticks

- [ ] 3. Keep a repeated task field's first occurrence
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Write Step 2's test first and run `cd stats && go test ./internal/guard -run 'TestTcfBodyFieldsFirstOccurrenceWins$' -count=1`: it fails, the task's files reading `[") and so on."]`.
  - [ ] **Step 2: Write the test** `TestTcfBodyFieldsFirstOccurrenceWins` at the end of `stats/internal/guard/check_task_commit_fields_test.go`: `tcfParseTask` over the lines of `- [ ] 1. One\n\n**Files:** \`a.go\`\n**Tests:** none\n**Commit:** fix: one\n\nCorrection: the shim moved\n**Files:**) and so on.\n` for task `1` yields `files == []string{"a.go"}`.
  - [ ] **Step 3: Implement** per decision `task-field-first-wins`: in `stats/internal/guard/planparse.go`'s `tcfBodyFields`, `commitField` records `fields[current]` only when `current` is not already a key; its comment says a repeated field keeps its first occurrence. In `stats/internal/guard/planshape.go`, F1's message ends "the parser keeps the first occurrence and ignores this one".
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestTcfBodyFieldsFirstOccurrenceWins$|TestCheckTaskCommitFields$|TestCheckPlanShape$|TestPlanDispatchBundles$' -count=1` passes; `for f in $(git ls-files '*tasks.md'); do scripts/check-plan-shape.sh "$f"; done > "$TMPDIR/after.txt" 2>&1` gives output identical to the same loop run at the task's parent commit (`git stash` is not needed — run it before Step 3 into `$TMPDIR/before.txt`). Commit.
    <!-- measured: for f in $(git ls-files '*tasks.md'); do scripts/check-plan-shape.sh "$f"; done 2>&1 | grep -c 'declares a second' @ 5611091f → 0 -->

**Build:** green
**Files:** `stats/internal/guard/planparse.go`, `stats/internal/guard/planshape.go`, `stats/internal/guard/check_task_commit_fields_test.go`
**Tests:** `TestTcfBodyFieldsFirstOccurrenceWins`
**Regression:** `TestTcfBodyFieldsFirstOccurrenceWins` fails — a later prose line opening `**Files:**` replaces the task's declared files again.
**Commit:** fix(guard): keep a repeated task field's first occurrence
**After:** none

**Decision:** task-field-first-wins

- [ ] 4. Await loaded panel content in the view tests
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `stats/web/src/views/views.test.tsx`'s `vi.mock("../api", …)` factory, temporarily return `fetchStatsView: (...a: unknown[]) => new Promise((r) => setTimeout(r, 500)).then(() => fetchStatsViewMock(...a))` in place of `fetchStatsView: fetchStatsViewMock`, and the same in `stats/web/src/views/RunDetail.test.tsx`'s factory; `cd stats/web && npx vitest run src/views/views.test.tsx src/views/RunDetail.test.tsx -t 'decisions shows a class|decisions groups runs by implementer model|flow health|still reads the header totals'` fails at the five sites below (and only there). Keep the delay in place until Step 3.
  - [ ] **Step 2: Fix** per decision `await-loaded-content`: `views.test.tsx` "decisions shows a class × execution summary…" — `expect(await within(summaryRegion).findByText("regular / inline"))`; "decisions groups runs by implementer model" — `const rows = (await within(region).findAllByRole("row")).slice(1)`; "flow health shows eight totals tiles" — after the region `findByRole`, `await tile("Changes started").findByText("4")`; "flow health reads an unfinished period's means as unavailable" — `expect(await within(lead).findByTestId("unavailable"))`; `RunDetail.test.tsx` "Total cost" — `expect(await within(totalCostPanel).findByText("$99.99"))`. Later sync reads stay sync.
  - [ ] **Step 3: Verify.** With the delay still in place, the vitest run of Step 1 passes; revert both mock factories to `fetchStatsView: fetchStatsViewMock` (`git diff` shows no factory change); run it again — passes. `cd stats && make -s web-dist && cd web && npx tsc -b` exits 0. Commit.

**Build:** green
**Files:** `stats/web/src/views/views.test.tsx`, `stats/web/src/views/RunDetail.test.tsx`
**Tests:** none — repairs existing tests: the five view tests named in Step 2, the last `still reads the header totals from the server's aggregate after the panel recomposition (task 20)`.
**Regression:** none — Step 1's delayed mock fails the five sites again on revert.
**Commit:** test(web): await loaded panel content before reading it
**After:** none

**Decision:** await-loaded-content

- [ ] 5. Omit deleted files' bodies from panel diffs
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'irreversible-delete' stats/internal/guard/writepaneldiff.go` prints nothing; the fixture of `TestWritePanelDiffParity` deletes `gone.txt` (body `gone`) between `held` and the tip.
  - [ ] **Step 2: Write the test first** in `stats/internal/guard/write_panel_diff_test.go`: `TestWritePanelDiffOmitsDeletedBodies` builds one `wpdNewWT` worktree, runs `write-panel-diff final <wt> <wt> <mb>`, and asserts `final-review.diff` contains `diff --git a/gone.txt b/gone.txt` and `deleted file mode` but no line `-gone`, and `final-review.diff.touched` still lists `D\tgone.txt`. Run `cd stats && go test ./internal/guard -run 'TestWritePanelDiffOmitsDeletedBodies$' -count=1` — it fails.
  - [ ] **Step 3: Implement** per decision `panel-diff-irreversible-delete`: in `stats/internal/guard/writepaneldiff.go` the diff entry's flags become `[]string{"--irreversible-delete"}` (the `--name-status` entry unchanged); in the parity test's `wpdRecipe` call sites, the diff's flags argument becomes `--irreversible-delete` where it was `""`. In `scripts/write-panel-diff.sh`'s header, say each diff section runs with `--irreversible-delete`, so a deleted file shows as its header alone, its old body read with `git show <sha>:<path>`. In `skills/flow/review-panel.md`'s ENTRY CONTEXT block, append the sentence `> A deleted file shows in the diff as its header alone; read its old body with \`git show <sha>:<path>\`, the sha the section's diff range starts from — for \`slot-delta\` with a held sha, the held sha; otherwise the sha its section's \`# worktree:\` header names.` inside the block, before "Name the coverage trade".
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestWritePanelDiff|TestRenderSlotPrompt|TestCheckDispatchParagraphs' -count=1` passes; `scripts/check-dispatch-paragraphs.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/writepaneldiff.go`, `stats/internal/guard/write_panel_diff_test.go`, `scripts/write-panel-diff.sh`, `skills/flow/review-panel.md`
**Tests:** `TestWritePanelDiffOmitsDeletedBodies`
**Regression:** `TestWritePanelDiffOmitsDeletedBodies` fails — a deleted file's whole body is written into every panel diff again.
**Commit:** perf(guard): omit deleted files' bodies from panel diffs
**After:** none

**Decision:** panel-diff-irreversible-delete
