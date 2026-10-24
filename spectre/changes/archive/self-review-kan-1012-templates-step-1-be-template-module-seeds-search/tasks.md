# self-review-kan-1012-templates-step-1-be-template-module-seeds-search

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-1012's three self-review findings — F1 (Task 1), F2 (Task 2), F3 (Task 3, with
gymie's configuration in Task 4), as `proposal.md` names them.

**Architecture:** one predicate change in gymie's buildSrc port-block filter; a new flow-guard
`stage-task-files` behind a shim that implement.md's per-task commit recipe calls; a
project-declared exclusion in `write-panel-diff`, and gymie's declaration of it.

**Tech Stack:** Kotlin (gymie `buildSrc`), Gradle Kotlin DSL, Go (`stats/`), Bash, Markdown.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-24-self-review-kan-1012-templates-step-1-be-template-module-seeds-search-design.md`
in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer, and no `Co-Authored-By:` trailer. Never write `spectre/changes/` in either repository — this change's `patches/` and `verbatim-moves.txt` are already committed with the plan. Never merge or push.
- Tasks 2 and 3 run in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1012-templates-step-1-be-template-module-seeds-search` (`<agents-wt>` below), branch `self-review-kan-1012-templates-step-1-be-template-module-seeds-search`. Tasks 1 and 4 run in `/Users/tweety53/Projects/gymie-worktrees/kan-1012-templates-step-1-be-template-module-seeds-search` (`<gymie-wt>` below), branch `spectre/kan-1012-templates-step-1-be-template-module-seeds-search`, committing after its archive commit `612f5e9f` and rewriting none. Every `**Files:**` path is relative to its task's worktree root. Never touch a main checkout, and never create a worktree.
- A task's diff is `<agents-wt>/spectre/changes/self-review-kan-1012-templates-step-1-be-template-module-seeds-search/patches/task-<n>.diff`, applied with `git apply <that path>` from the task's worktree root. Task 2's diff creates `scripts/stage-task-files.sh` (mode 100755) and the symlink `skills/flow/scripts/stage-task-files.sh` → `../../../scripts/stage-task-files.sh`; confirm both with `ls -l` after applying. `git apply --check` passed for task-2 and task-3 on `<agents-wt>` at `ebfb5504`, and for task-1 and task-4 on `<gymie-wt>` at `612f5e9f`. A patch that does not apply is reported, never hand-merged by guesswork.
- Prose lands character for character as its patch carries it: `scripts/check-verbatim-moves.sh` matches every added and removed sentence against this change's `verbatim-moves.txt`. With Tasks 2 and 3 applied, `check-verbatim-moves.sh`, `check-guard-symlinks.sh`, `check-dispatch-paragraphs.sh`, `check-references.sh`, `check-markdown-integrity.sh`, `check-installed-citations.sh`, `check-normative-inventory.sh` and `check-python-suppressions.sh` exited 0, `cd stats && make -s web-dist && go vet ./...` exited 0, `cd stats && go test ./internal/guard -count=1` passed and `scripts/run-guard-tests.sh` passed.
- Go: run `gofmt -w` on every Go file you change before committing; `cd stats && gofmt -l .` prints nothing.
- Mutation proofs restore with `break-and-prove.sh` or `mutate-and-verify.sh`, or a saved diff re-applied with `git apply` — never `git checkout --` on a file carrying uncommitted edits.
- The gymie commits (Tasks 1 and 4) are commits outside every task of kan-1012's change (**Commits outside every task**, `skills/flow-contracts/git-boundaries.md`): add `-m "Flow-Key: self-review-kan-1012-templates-step-1-be-template-module-seeds-search-fix"`. Gymie's Gradle runs take a Bash `timeout: 600000`, one Gradle invocation at a time in `<gymie-wt>`. Never start, stop or restart `<gymie-wt>`'s dev stack; the live check is the parent's.

## Review Focus

- Task 1: a pinned service in `binding` whose recorded port differs from today's deterministic one (a flag toggled between runs) still loses its entry; a pinned service not running still loses its entry whatever its port.
- Task 2: a task deleting a file (`D ` in porcelain) and a task renaming one pass; a task path under `spectre/changes/` stays unstaged and is not reported; git's own error from `add` (an ignored or missing path) is printed and exits 2.
- Task 3: a worktree declaring no key gets byte-identical output to before (`TestWritePanelDiffParity`); a fence carrying several globs excludes each.

- [x] 1. Keep a running pinned service's entry when it sits on today's port
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `<gymie-wt>`, `grep -n 'name in placed && !(name in running && name !in binding)' buildSrc/src/main/kotlin/com/gymie/gradle/FrontendPortContract.kt` prints the predicate, which drops a running gateway's entry for being in the binding whatever its port.
    <!-- measured: that grep @ 612f5e9f → FrontendPortContract.kt:131 -->
  - [x] **Step 2: Apply `patches/task-1.diff`** from `<gymie-wt>`: the `deterministic` parameter, the new predicate and the KDoc bullet (`buildSrc/src/main/kotlin/com/gymie/gradle/FrontendPortContract.kt`); `block` passed at the one call site (`gradle/dev-lifecycle.gradle.kts`); `TODAY` passed by every existing call and the new test `re-running a whole-stack devStart over its own partly up stack is not refused` (`buildSrc/src/test/kotlin/com/gymie/gradle/FrontendPortContractTest.kt`).
    <!-- measured: git apply --check patches/task-1.diff @ 612f5e9f → exit 0 -->
  - [x] **Step 3: Prove the test pins the fix.** With a saved diff re-applied afterwards, replace the predicate `name in placed && (name !in running || (name in binding && placed[name] != deterministic[name]))` with `name in placed && !(name in running && name !in binding) && deterministic.isNotEmpty()` — `./gradlew :buildSrc:test --tests 'com.gymie.gradle.FrontendPortContractTest'` fails `re-running a whole-stack devStart over its own partly up stack is not refused`; restore.
    <!-- measured: that flip @ 612f5e9f with the patch → the new test FAILED; restored → 24 tests, 0 failures -->
  - [x] **Step 4: Verify.** `./gradlew ktlintFormat`, then `./gradlew verifyChange :buildSrc:test --tests 'com.gymie.gradle.FrontendPortContractTest'` exits 0 (the buildSrc build runs its own ktlint and detekt first). Commit with the Flow-Key line.
    <!-- measured: ./gradlew :buildSrc:test --tests 'com.gymie.gradle.FrontendPortContractTest' @ 612f5e9f with the patch → BUILD SUCCESSFUL, 24 tests -->

**Build:** green
**Files:** `buildSrc/src/main/kotlin/com/gymie/gradle/FrontendPortContract.kt`, `buildSrc/src/test/kotlin/com/gymie/gradle/FrontendPortContractTest.kt`, `gradle/dev-lifecycle.gradle.kts`
**Tests:** `com.gymie.gradle.FrontendPortContractTest`
**Regression:** `re-running a whole-stack devStart over its own partly up stack is not refused` fails — the running gateway's entry is dropped again and `resolveBlock` refuses the re-run.
**Baseline:** before=23 after=24 <!-- measured: grep -cw '@Test' buildSrc/src/test/kotlin/com/gymie/gradle/FrontendPortContractTest.kt @ 612f5e9f → 23, and 24 with the patch -->
**Commit:** fix(gradle): keep a running pinned service's entry when it sits on today's port
**After:** none

**Decision:** pinned-entry-kept-on-deterministic-port

**Decision:** k1012-live-verification-line-left

- [x] 2. Stage a task's files through a guard that refuses a path left unstaged
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In a scratch repo (`git init -q -b work r && cd r`), commit `src/app/template/src/test/x/T.kt` and `spectre/changes/c/tasks.md`, then create `src/app/template/src/main/k/A.kt`: `git add -- src/app/template/src/main/k/A.kt ':(exclude)spectre/changes/'; echo $?; git diff --cached --name-only` prints `0` and nothing staged.
    <!-- measured: that sequence @ git 2.50.1 (Apple Git-155) → exit 0, nothing staged; with the exclude first → A.kt staged -->
  - [x] **Step 2: Apply `patches/task-2.diff`** from `<agents-wt>`: `stageTaskFiles` (`stats/internal/guard/stagetaskfiles.go`), `TestStageTaskFilesStagesNewFilesUnderNewDirectories` and `TestStageTaskFilesRefusesAPathLeftUnstaged` (`stats/internal/guard/stage_task_files_test.go`), the shim `scripts/stage-task-files.sh`, its symlink `skills/flow/scripts/stage-task-files.sh`, and the COMMIT-PER-TASK fence and its following sentences (`skills/flow/implement.md`).
    <!-- measured: git apply --check patches/task-2.diff @ ebfb5504 → exit 0 -->
  - [x] **Step 3: Prove the tests pin the fix.** With a saved diff re-applied afterwards: in `stagetaskfiles.go` replace `append(add, files...)}` with `append(append([]string{"add", "--"}, files...), add[2:]...)}` (the excludes after the paths) — `cd stats && go test ./internal/guard -run 'TestStageTaskFiles' -count=1` fails `TestStageTaskFilesStagesNewFilesUnderNewDirectories` with exit 1; restore. Then replace `if len(left) > 0 {` with `if false && len(left) > 0 {` — the same command fails `TestStageTaskFilesRefusesAPathLeftUnstaged`; restore.
    <!-- measured: both flips @ ebfb5504 with the patch → the named test FAILs each time; restored → ok -->
  - [x] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestStageTaskFiles|TestShim|TestCheckGuardSymlinks' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-guard-symlinks.sh`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-installed-citations.sh` exit 0. Commit.
    <!-- measured: those commands @ ebfb5504 with the patch → ok, exit 0 each -->

**Build:** green
**Files:** `stats/internal/guard/stagetaskfiles.go`, `stats/internal/guard/stage_task_files_test.go`, `scripts/stage-task-files.sh`, `skills/flow/scripts/stage-task-files.sh`, `skills/flow/implement.md`
**Tests:** `TestStageTaskFilesStagesNewFilesUnderNewDirectories`, `TestStageTaskFilesRefusesAPathLeftUnstaged`
**Regression:** `TestStageTaskFilesStagesNewFilesUnderNewDirectories` fails — the add drops new files on kan-1012's layout again; `TestStageTaskFilesRefusesAPathLeftUnstaged` fails — a silent miss reaches the commit.
**Baseline:** before=0 after=2 <!-- measured: grep -c -E '^func Test' stats/internal/guard/stage_task_files_test.go @ ebfb5504 → absent (0), and 2 with the patch -->
**Commit:** fix(guard): stage a task's files and refuse one git left unstaged
**After:** none

**Decision:** stage-task-files-guard

- [x] 3. Leave a project's declared globs out of the panel diff
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `grep -c 'exclude' stats/internal/guard/writepaneldiff.go` prints `0`.
    <!-- measured: that grep @ ebfb5504 → 0 -->
  - [x] **Step 2: Apply `patches/task-3.diff`** from `<agents-wt>`: `wpdExcludes` and its use in both diff calls (`stats/internal/guard/writepaneldiff.go`), `TestWritePanelDiffHonoursProjectExcludes` (`stats/internal/guard/write_panel_diff_test.go`), the exclusion paragraph (`scripts/write-panel-diff.sh`), the `## panel diff exclude` row (`skills/flow-contracts/project-configuration.md`).
    <!-- measured: git apply --check patches/task-3.diff @ ebfb5504 → exit 0 -->
  - [x] **Step 3: Prove the test pins the fix.** With a saved diff re-applied afterwards, replace `excludes := wpdExcludes(smcAbs(env, wt))` with `excludes := []string(nil)` — `cd stats && go test ./internal/guard -run 'TestWritePanelDiffHonoursProjectExcludes' -count=1` fails with `final-review.diff carries the excluded handoff_x/`; restore.
    <!-- measured: that flip @ ebfb5504 with the patch → FAIL on both files; restored → ok -->
  - [x] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestWritePanelDiff' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-installed-citations.sh` exit 0. Commit.
    <!-- measured: those commands @ ebfb5504 with tasks 2 and 3 → ok, exit 0 each -->

**Build:** green
**Files:** `stats/internal/guard/writepaneldiff.go`, `stats/internal/guard/write_panel_diff_test.go`, `scripts/write-panel-diff.sh`, `skills/flow-contracts/project-configuration.md`, `stats/internal/guard/paneldiffsize.go`, `stats/internal/guard/paneltouchedpaths.go`, `stats/internal/guard/planclass.go`
**Tests:** `TestWritePanelDiffHonoursProjectExcludes`
**Regression:** `TestWritePanelDiffHonoursProjectExcludes` fails — a declared glob's paths reach the panel diff and its `.touched` list again.
**Baseline:** before=4 after=5 <!-- measured: grep -c -E '^func Test' stats/internal/guard/write_panel_diff_test.go @ ebfb5504 → 4, and 5 with the patch -->
**Commit:** feat(guard): leave a project's declared globs out of the panel diff
**After:** none

**Decision:** panel-diff-project-excludes

- [x] 4. Declare gymie's handoff bundles excluded from the panel diff
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `<gymie-wt>`, `grep -c '^## panel diff exclude' .flow/project.md` prints `0`, and `git ls-files 'handoff_*' | wc -l` prints `102`.
    <!-- measured: both @ 612f5e9f → 0 and 102 -->
  - [x] **Step 2: Apply `patches/task-4.diff`** from `<gymie-wt>`: the `## panel diff exclude` section, fence `handoff_*/**`, before `## jira` (`.flow/project.md`).
    <!-- measured: git apply --check patches/task-4.diff @ 612f5e9f → exit 0 -->
  - [x] **Step 3: Verify.** After Task 3's commit, from `<agents-wt>`: `scripts/project-get.sh <gymie-wt> 'panel diff exclude'` exits 0 printing the fence; `mkdir -p "$TMPDIR/k1012-canon" && scripts/write-panel-diff.sh final "$TMPDIR/k1012-canon" <gymie-wt> 1787e99e` exits 0 and `grep -c handoff_ "$TMPDIR/k1012-canon/.superpowers/sdd/final-review.diff.touched"` prints `0`. Commit with the Flow-Key line.
    <!-- measured: write-panel-diff with tasks 3 and 4 applied, canonical dir in the scratchpad → 234273-byte diff, 0 handoff_ lines; without task 4 → 6340872 bytes, 65 lines -->

**Build:** green
**Files:** `.flow/project.md`
**Tests:** none — a configuration value; Task 3's test covers the mechanism and Step 3 reads this value through it
**Regression:** none — no test is declared.
**Commit:** chore: exclude the handoff bundles from the panel diff
**After:** Task 3

**Decision:** gymie-exclude-own-task
