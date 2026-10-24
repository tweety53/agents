## Context

Self-review fix change for `kan-1012-templates-step-1-be-template-module-seeds-search`, planned on
`self-review-kan-1012-templates-step-1-be-template-module-seeds-search` at agents `ebfb5504`
(`origin/main`). Three non-`big` findings, F1–F3 as `proposal.md` names them: F1 is carried by
Task 1, F2 by Task 2, F3 by Task 3 with its gymie configuration in Task 4. Tasks 1 and 4 land on
gymie's own change branch `spectre/kan-1012-templates-step-1-be-template-module-seeds-search`
(backend at `612f5e9f`) and ride that change's merge and push to `develop`, per **Pipeline defects
found mid-run** (`skills/flow-contracts/pipeline.md`). Each finding was traced before its fix was
designed and `origin/main` holds no fix for any of them. F2, which the pass could not reproduce,
was reproduced here: git 2.50.1 (Apple Git-155) given the task's paths first and an
`:(exclude)` after them stages only the tracked paths on kan-1012's layout and exits 0, while the
same add with the excludes first stages them all — the brainstorm design records the trace. Every
question was auto-resolved to its recommended option (**Auto-resolution**,
`skills/flow-contracts/operator-prompts-auto-resolution.md`). The brainstorm design is
`.superpowers/sdd/2026-10-24-self-review-kan-1012-templates-step-1-be-template-module-seeds-search-design.md`
in this worktree. Every fix was built and tested and saved as `patches/task-<n>.diff`; every
reworded or added run-loaded sentence is listed in `verbatim-moves.txt`.

## Decisions

### A pinned entry already on today's port survives the binding

**ID:** pinned-entry-kept-on-deterministic-port
**Status:** active
**Chosen:** `withoutStaleOAuthPinnedEntries(deterministic, binding, running)` drops a pinned entry when its service is not running, or is in `binding` on a port other than its `deterministic` one; the one caller in `gradle/dev-lifecycle.gradle.kts` passes the `block` it already computed. The KDoc's "In [binding]" bullet states the new condition and the kan-1012 refusal.
**Considered:** never dropping an in-binding entry — a run that toggled `-PisolateOAuthPorts` would again trust a port recorded under the other flag, the defect the drop exists for; teaching `resolveWholeBlock` that a running unplaced service on its own deterministic port is ours — a second place deciding what the record means, where the filter already owns it.

### The archived live-verification line stays as written

**ID:** k1012-live-verification-line-left
**Status:** active
**Chosen:** leave `spectre/changes/archive/kan-1012-…/live-verification.md` line 3 ("held by another workspace's stack") unedited, and record here that the holder was this worktree's own gateway, which Task 1's defect stopped `devStart` restarting.
**Considered:** correcting the line in Task 1's gymie commit — a commit on the change branch touching `spectre/changes/` is what `check-task-commit-planning-paths.sh` refuses at kan-1012's own boundary, and `stage-task-files.sh` excludes the path by design; not cheap.

### A guard stages the task's files and checks the result

**ID:** stage-task-files-guard
**Status:** active
**Chosen:** a flow-guard `stage-task-files` with shim `scripts/stage-task-files.sh` and symlink `skills/flow/scripts/stage-task-files.sh`: the clearing reset, `git add --` with the three excludes ahead of the paths, then `git status --porcelain --untracked-files=all -- <files>` — any line outside the planning prefixes whose worktree column is not blank is printed and exits 1. implement.md's COMMIT-PER-TASK fence becomes `stage-task-files.sh <abs-worktree> <files> && git -C … commit … -- <files>`. Two Go tests: kan-1012's layout fails with the excludes after the paths; a stub git whose `add` stages nothing is refused with every path named.
**Considered:** the post-add check typed into the prose fence — a chained `status` filter an implementer retypes every task, untestable; reordering the excludes alone — cures the layout measured but leaves the next silent miss undetected; dropping the excludes — the reset already clears planning paths, but a task path under `spectre/changes/` would then be staged.

### write-panel-diff reads a project's exclusion globs

**ID:** panel-diff-project-excludes
**Status:** active
**Chosen:** (b) — `write-panel-diff` reads the fenced lines of `## panel diff exclude` in each worktree's own `.flow/project.md` and appends `-- . ':(glob,exclude)<glob>'…` to both its diff and its `--name-status` call; absent, the calls are unchanged. `project-configuration.md` gains the row and the script header the pathspec. Measured on kan-1012's backend against `1787e99e`: `final-review.diff` 6,340,872 → 234,273 bytes, `handoff_` lines in `.touched` 65 → 0.
**Considered:** (a) `.gitattributes` `-diff` on `handoff_*` — hides the bundles' diffs from humans and from GitHub pull requests too; a hard-coded `handoff_*` exclusion in the guard — one project's directory name in the pipeline, which `project-configuration.md` forbids; reading the canonical worktree's `project.md` for every worktree — globs are repository-relative, so each repository declares its own.

### Gymie's exclusion is its own task

**ID:** gymie-exclude-own-task
**Status:** active
**Chosen:** Task 4 adds `## panel diff exclude` with `handoff_*/**` to gymie's `.flow/project.md` on the gymie change branch, `**After:** Task 3`; F3 is carried by Task 3, Task 4 is its configuration in the other repository.
**Considered:** folding the key into Task 3 — one commit cannot span two repositories; into Task 1's gymie commit — mixes F1 and F3 in one commit the review maps to one finding.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 2 changes the command every `/flow` implementer runs for every task commit, and a refusal wrong by one column stops every run.
**Considered:** `flow-low` — each diff is small, but Task 1's predicate and Task 2's porcelain parsing are exactly the one-character defects a light review passes.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-kan-1012-templates-step-1-be-template-module-seeds-search-fix`, Tasks 1 to 4 in order: four small patches, already built and applied in that order.
**Considered:** one fixer per repository — two dispatches for two one-file gymie diffs.

## Open questions

## Live check

Each step runs before landing, from the named worktree with every task committed:

1. **devStart re-run.** In the gymie worktree, `./gradlew devStart` with its
   stack already up — before: fails naming "gateway is running and this workspace's
   .dev-stack/port-block record does not place it"; after: restarts the stack and prints its URLs.
   Not working looks like that refusal again.
2. **Staging guard.** In a scratch repo built as `TestStageTaskFilesStagesNewFilesUnderNewDirectories`
   builds it, `scripts/stage-task-files.sh <repo> src/app/template/src/main/k/A.kt
   src/app/template/src/test/x/T.kt` — after: exit 0 and both paths in `git diff --cached
   --name-only`.
3. **Panel diff.** From this worktree, `scripts/write-panel-diff.sh final <scratch-dir>
   <gymie-worktree> 1787e99e` — after: `grep -c handoff_ <scratch-dir>/.superpowers/sdd/final-review.diff.touched`
   prints `0`; before: `65`.
