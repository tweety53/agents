# self-review-kan-1001-be-cardio-workouts-count-on-the-profile-s-12

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the six non-`big` findings of `kan-1001-be-cardio-workouts-count-on-the-profile-s-12`'s
self-review pass — the unclosable pass, the over-eager heredoc scan, the bundle's missing
`live-verification.md`, and gymie's seed-before-link wiring, hand-copied bootRun arguments and
orphaned workspaces.

**Architecture:** one early return in `pfEnd`; one predicate in `protect-main-checkout.py`; one
entry in `archivedFiles`; in gymie, two task dependencies, one shared argument builder with a
printing task, and two recipe sentences plus a one-off workspace cleanup. Each task applies its
committed patch from `patches/` beside this file.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-21-self-review-kan-1001-be-cardio-workouts-count-on-the-profile-s-12-design.md`
in this agents worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer, and no `Co-Authored-By` trailer. Never write `spectre/changes/` in this agents worktree: this change's `patches/`, `verbatim-moves.txt` and `memory-protect-main-checkout-hook-quirks.md` are already committed with the plan. Never merge or push.
- Each task names the repository it runs in, by its worktree; every `**Files:**` path is relative to that worktree's root. Tasks 1–3 run in this agents worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1001-be-cardio-workouts-count-on-the-profile-s-12`, branch `self-review-kan-1001-be-cardio-workouts-count-on-the-profile-s-12`. Tasks 4–6 run in the gymie worktree `/Users/tweety53/Projects/gymie-worktrees/kan-1001-be-cardio-workouts-count-on-the-profile-s-12`, branch `spectre/kan-1001-be-cardio-workouts-count-on-the-profile-s-12`, and ride that change's merge and push; never touch its `spectre/changes/` (archived). Never touch a main checkout, and never create a worktree.
- `P` below is this change root, `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1001-be-cardio-workouts-count-on-the-profile-s-12/spectre/changes/self-review-kan-1001-be-cardio-workouts-count-on-the-profile-s-12`. Each task's diff is `$P/patches/task-<n>.diff`, applied with `git apply` from its worktree's root. Tasks 1–3 are independent; Tasks 4, 5 and 6 apply in that order, each on the tree the one before commits. A patch that does not apply is reported, never hand-merged by guesswork.
- Run `cd stats && gofmt -w .` on Go changes before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.
- Every gymie Gradle call runs with a Bash `timeout: 600000`. In gymie run `./gradlew ktlintFormat` as its own invocation before `./gradlew verifyChange`. Never start, stop or rebuild the gymie dev stack.

## Review Focus

- Task 2: every existing deny of `scripts/test-protect-main-checkout.sh` still denies — a runner word, a substitution, a target run as a command word under any spelling — and a target named only as an argument is data.
- Task 5: the seed run's `--args` carries every value it did before (`--app.seed.enabled=true`, the invite, the fixed day, datasource, storage, Redis index, OAuth placeholders).
- Task 6: no workspace a live worktree derives is removed.

- [ ] 1. Close a self-review pass as nothing to close
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In this agents worktree, apply the test hunk alone — `git apply --include='stats/internal/guard/pipeline_fix_test.go' $P/patches/task-1.diff` — and run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestPipelineFixEndPassRecordsNothing' -count=1`: it fails, `rc 0:` with no `PIPELINE-FIX-END-NA` line.
    <!-- measured: the test hunk alone, go test -run TestPipelineFixEndPassRecordsNothing @ branch self-review-kan-1001-be-cardio-workouts-count-on-the-profile-s-12 (1c4c3e38) → FAIL, pipeline_fix_test.go:321: rc 0: -->
  - [ ] **Step 2: Apply the fix** — the rest: `git apply --exclude='stats/internal/guard/pipeline_fix_test.go' $P/patches/task-1.diff`.
  - [ ] **Step 3: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestPipelineFix|TestCheckLoopDispatch' -count=1` passes.
  - [ ] **Step 4: Commit** the two files below with the subject below.

**Build:** green
**Files:** `stats/internal/guard/pipelinefix.go`, `stats/internal/guard/pipeline_fix_test.go`
**Tests:** `TestPipelineFixEndPassRecordsNothing`
**Regression:** `TestPipelineFixEndPassRecordsNothing` — reverting the commit sends the pass's end to `flow record dispatch end`, and the test fails on the missing `PIPELINE-FIX-END-NA` line.
**Baseline:** before=9 after=10
<!-- measured: grep -c '^func Test' stats/internal/guard/pipeline_fix_test.go @ branch self-review-kan-1001-be-cardio-workouts-count-on-the-profile-s-12, before and after patches/task-1.diff -->
**Commit:** fix(guard): close a self-review pass as nothing to close
**After:** none

**Decision:** pass-end-is-a-no-op

- [ ] 2. Keep a data heredoc's body as data when its target is only named
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In this agents worktree, apply the harness hunk alone — `git apply --include='scripts/test-protect-main-checkout.sh' $P/patches/task-2.diff` — and run `bash scripts/test-protect-main-checkout.sh`: cases 118 and 119 fail (`wanted allow, got deny`), every other case passes.
    <!-- measured: the harness hunk alone, bash scripts/test-protect-main-checkout.sh @ branch self-review-kan-1001-be-cardio-workouts-count-on-the-profile-s-12 (1c4c3e38) → 2 failure(s), cases 118 and 119 -->
  - [ ] **Step 2: Apply the fix** — the rest: `git apply --exclude='scripts/test-protect-main-checkout.sh' $P/patches/task-2.diff`.
  - [ ] **Step 3: Verify.** `bash scripts/test-protect-main-checkout.sh` ends `all protect-main-checkout cases pass`; `scripts/check-python-suppressions.sh` and `scripts/check-references.sh` exit 0.
  - [ ] **Step 4: Update the saved memory** (no commit — it lives outside the repository): copy `$P/memory-protect-main-checkout-hook-quirks.md` over `/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/memory/protect-main-checkout-hook-quirks.md`, then set its `modified:` line to the current UTC time. Its `MEMORY.md` index line stays.
  - [ ] **Step 5: Commit** the two files below with the subject below.

**Build:** green
**Files:** `hooks/protect-main-checkout.py`, `scripts/test-protect-main-checkout.sh`
**Tests:** Case 118, Case 119, Case 120, Case 121
**Regression:** Case 118 and Case 119 — reverting the hook denies a heredoc whose target is only named as an argument, or that sits beside another path-named command. Case 120 and Case 121 pin the denies the narrowing must keep.
**Baseline:** before=105 after=109
<!-- measured: grep -c '^expect ' scripts/test-protect-main-checkout.sh @ branch self-review-kan-1001-be-cardio-workouts-count-on-the-profile-s-12, before and after patches/task-2.diff -->
**Commit:** fix(hooks): keep a data heredoc's body as data when its target is only named
**After:** none

**Decision:** heredoc-target-run-not-named

- [ ] 3. Serve archived live-verification.md in the self-review bundle
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In this agents worktree, apply the test hunks alone — `git apply --include='stats/internal/selfreview/bundle_test.go' $P/patches/task-3.diff` — and run `cd stats && make -s web-dist && go test ./internal/selfreview/ -run 'TestBundleServesArchivedLiveVerification' -count=1`: it fails, the bundle missing the `live-verification.md` section.
    <!-- measured: with git.go reverted, go test -run TestBundleServesArchivedLiveVerification @ branch self-review-kan-1001-be-cardio-workouts-count-on-the-profile-s-12 (1c4c3e38) → FAIL -->
  - [ ] **Step 2: Apply the fix** — the rest: `git apply --exclude='stats/internal/selfreview/bundle_test.go' $P/patches/task-3.diff`.
  - [ ] **Step 3: Verify.** `cd stats && gofmt -l . && go vet ./internal/selfreview/ ./internal/api/ && go test ./internal/selfreview/ -run 'TestBundle' -count=1` passes; `scripts/check-verbatim-moves.sh` reports 0 violations (the four sentences are listed in `$P/verbatim-moves.txt`); `scripts/check-references.sh`, `scripts/check-installed-citations.sh` and `python3 scripts/check-markdown-integrity.py` exit 0.
  - [ ] **Step 4: Commit** the five files below with the subject below.

**Build:** green
**Files:** `stats/internal/selfreview/git.go`, `stats/internal/selfreview/bundle_test.go`, `stats/internal/api/selfreview.go`, `skills/flow-contracts/finish-contract-run1.md`, `skills/flow-fast/SKILL.md`
**Tests:** `TestBundleServesArchivedLiveVerification`
**Regression:** `TestBundleServesArchivedLiveVerification` — reverting the commit drops `live-verification.md` from `archivedFiles`, and the test fails on the missing section.
**Baseline:** before=30 after=31
<!-- measured: grep -c '^func Test' stats/internal/selfreview/bundle_test.go @ branch self-review-kan-1001-be-cardio-workouts-count-on-the-profile-s-12, before and after patches/task-3.diff -->
**Commit:** fix(selfreview): serve archived live-verification.md in the bundle
**After:** none

**Decision:** live-verification-after-visual

- [ ] 4. Link .env before dbSeed and devWorkspaceCreate seed
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In the gymie worktree, run `./gradlew -q -m dbSeed` and `./gradlew -q -m devWorkspaceCreate`: each prints its own task line and no `:linkEnvFile SKIPPED` line.
    <!-- measured: ./gradlew -q -m dbSeed and -m devWorkspaceCreate @ branch spectre/kan-1001-be-cardio-workouts-count-on-the-profile-s-12 (94c2e80f) → ":dbSeed SKIPPED" / ":devWorkspaceCreate SKIPPED" alone -->
  - [ ] **Step 2: Apply** `git apply $P/patches/task-4.diff` in the gymie worktree (it creates `scripts/test-env-link-wiring.sh`, mode 755).
  - [ ] **Step 3: Verify.** `bash scripts/test-env-link-wiring.sh` ends `all env-link wiring cases pass`; `./gradlew ktlintFormat`, then `./gradlew verifyChange :app:test --tests 'com.gymie.CiWorkflowTest'` passes.
  - [ ] **Step 4: Commit** the four files below on branch `spectre/kan-1001-be-cardio-workouts-count-on-the-profile-s-12` with the subject below; it rides that change's merge and push.

**Build:** green
**Files:** `gradle/git-hooks.gradle.kts`, `scripts/test-env-link-wiring.sh`, `.github/workflows/ci.yml`, `.flow/project.md`
**Tests:** `test-env-link-wiring.sh`
**Regression:** `test-env-link-wiring.sh` — reverting the commit's wiring leaves `dbSeed` and `devWorkspaceCreate` without `linkEnvFile`, and the harness fails both.
**Baseline:** before=7 after=8
<!-- measured: git ls-files 'scripts/test-*.sh' | grep -c . @ branch spectre/kan-1001-be-cardio-workouts-count-on-the-profile-s-12, before and after patches/task-4.diff -->
**Commit:** fix(dev): link .env before dbSeed and devWorkspaceCreate seed
**After:** none

**Decision:** link-env-before-seed

- [ ] 5. Print the before backend's bootRun arguments from the seed run's builder
  - [ ] **Step 1: Apply** `git apply $P/patches/task-5.diff` in the gymie worktree, on Task 4's commit.
  - [ ] **Step 2: Verify.** `./gradlew ktlintFormat`, then `./gradlew verifyChange` passes, and `buildSrc`'s suite (it runs as `:buildSrc:jar`'s finalizer on any invocation) reports `BootRunArgsTest` passed in `buildSrc/build/test-results/test/TEST-com.gymie.gradle.BootRunArgsTest.xml`; `./gradlew -q devBootRunArgs -PbootPort=18080` prints one `--args=` line naming this worktree's database `gymie_kan_1001_be_5a0d`, and `./gradlew -q devBootRunArgs` fails with `-PbootPort=<free port> is required`.
    <!-- measured: ./gradlew -q devBootRunArgs -PbootPort=18080 with patches 4 and 5 applied @ branch spectre/kan-1001-be-cardio-workouts-count-on-the-profile-s-12 (94c2e80f) → one --args= line, database gymie_kan_1001_be_5a0d, redis database 1 -->
  - [ ] **Step 3: Commit** the four files below on branch `spectre/kan-1001-be-cardio-workouts-count-on-the-profile-s-12` with the subject below.

**Build:** green
**Files:** `buildSrc/src/main/kotlin/com/gymie/gradle/BootRunArgs.kt`, `buildSrc/src/test/kotlin/com/gymie/gradle/BootRunArgsTest.kt`, `gradle/dev-lifecycle.gradle.kts`, `.flow/project.md`
**Tests:** `BootRunArgsTest`
**Regression:** `BootRunArgsTest` — reverting the commit removes the shared builder, and the test no longer compiles.
**Baseline:** before=250 after=251
<!-- measured: grep -rhow '@Test' buildSrc/src/test/kotlin | grep -c . @ branch spectre/kan-1001-be-cardio-workouts-count-on-the-profile-s-12, before and after patches/task-5.diff -->
**Commit:** feat(dev): print the before backend's bootRun arguments from the seed run's builder
**After:** Task 4

**Decision:** boot-run-args-builder

- [ ] 6. Remove the before worktree's workspace, and the orphans left so far
  - [ ] **Step 1: Apply** `git apply $P/patches/task-6.diff` in the gymie worktree, on Task 5's commit.
  - [ ] **Step 2: Verify.** `git diff --stat` names `.flow/project.md` alone, and `git diff | grep -c '^+```'` prints `0`.
  - [ ] **Step 3: Commit** `.flow/project.md` on branch `spectre/kan-1001-be-cardio-workouts-count-on-the-profile-s-12` with the subject below.
  - [ ] **Step 4: Repair (no commit) — reproduce before touching anything; do not guess.** In the gymie worktree run `./gradlew devWorkspaceInventory` and record its workspace lines. Then, for each worktree `git -C /Users/tweety53/Projects/gymie worktree list` and `git -C /Users/tweety53/Projects/gymie-frontend worktree list` print (main checkouts excluded), run `./gradlew devWorkspace -PworkspaceChange=<worktree directory name>` and record the id on its first `workspace <id>:` line.
    <!-- measured: devWorkspaceInventory → fix-backend-f56d, kan-1000-fe-a183, kan-1001-be-5a0d, kan-1001-be-9051, kan-692-flow-23fd, kan-825-flow-9836, self-review-4fe5; live worktree ids → kan-1001-be-5a0d, ci-only-on-2d89, docs-ce28, kan-627-flow-9a73, review-a3ed, self-review-b008 @ 2026-10-21 -->
  - [ ] **Step 5: Remove** each of `kan-1001-be-9051`, `kan-1000-fe-a183`, `kan-692-flow-23fd`, `kan-825-flow-9836`, `fix-backend-f56d`, `self-review-4fe5` that the inventory still lists and no live worktree derived: `./gradlew devWorkspaceRemove -PworkspaceId=<id>`, then `./gradlew -q devWorkspaceSurvivors -PworkspaceId=<id>` prints nothing. An id a live worktree derives is left in place and reported. Re-run `./gradlew devWorkspaceInventory` and report what it lists.

**Build:** green
**Files:** `.flow/project.md`
**Tests:** **none** — project configuration prose; the repair is a one-off command.
**Regression:** none — prose.
**Commit:** docs(dev): remove the before worktree's workspace before the worktree
**After:** Task 4, 5

**Decision:** before-workspace-teardown
