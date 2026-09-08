# kan-440-review-panel-roster-drift-mutation-slot-and

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.

> **Relocation:** no

- [x] 1. Align DefaultReviewers with the ratified default roster
**Build:** green
**Files:** `stats/internal/store/settings.go`, `stats/internal/store/settings_test.go`, `skills/flow/SKILL.md`
**Tests:** `TestDefaultReviewersRoster`
**Regression:** reverting the commit puts the 3-slot list back and `TestDefaultReviewersRoster` fails
**Baseline:** before=394 after=395
<!-- measured: cd stats && go test ./internal/store/ ./internal/api/ -count=1 -v | grep -c '^=== RUN' @ branch spectre/kan-440-review-panel-roster-drift-mutation-slot-and -->
<!-- predicted: same command after task 1 (one test added, none removed) -->
**Commit:** fix(stats): align DefaultReviewers with the ratified default roster

  - [x] **Step 1:** RED — add `TestDefaultReviewersRoster` to `stats/internal/store/settings_test.go`, asserting the exact list `primary, principles, code-review-low, mutation` in order; run it targeted and record the failure.
  - [x] **Step 2:** GREEN — change `DefaultReviewers` (`stats/internal/store/settings.go:58`) to the 4-slot list and update its doc comment (the no-row fallback semantics at `GetSettings` are unchanged); re-run targeted → pass.
  - [x] **Step 3:** update `skills/flow/SKILL.md`'s Model-resolution "Unreachable" row to enumerate the 4-slot fallback.
  - [x] **Step 4:** verify (block below), then commit with the declared subject and a `Task-Id: 1` trailer.

```bash verified:resolved from .flow/project.md's ## lint and ## test lists; gofmt/vet/test run clean at this plan's baseline
cd stats && gofmt -l internal/store/
cd stats && go vet ./internal/store/ ./internal/api/
cd stats && go test ./internal/store/ -run TestDefaultReviewersRoster -count=1
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-markdown-integrity.py
scripts/check-model-resolution-shell.sh
```

- [x] 2. Dedup spec_root_leaf's dual-tree warning behind a TMPDIR marker
**Build:** green
**Files:** `scripts/lib/spec-root.sh`, `skills/flow/scripts/lib/spec-root.sh`, `scripts/test-lib-spec-root.sh`
**Tests:** `scripts/test-lib-spec-root.sh`
**Regression:** reverting the commit restores warn-per-call; the harness's dedup cases fail
**Baseline:** before=0 after=6
<!-- predicted: bash scripts/test-lib-spec-root.sh after task 2 reports 6 cases, 0 failures — the harness is new, so before is its absence -->
**Commit:** fix(scripts): warn once per tmp lifetime on a dual spec tree

  - [x] **Step 1:** RED — add `scripts/test-lib-spec-root.sh` (auto-discovered by `scripts/run-guard-tests.sh`) with exactly six cases: spectre-only selects `spectre` silently; openspec-only selects `openspec` silently; neither selects `spectre` silently; both warns once and selects `spectre`; a second call on the same dir is silent; a distinct dir still warns. Each case gets a private `TMPDIR` (the shared-tempdir failure mode `.flow/project.md`'s `## test` records).
  - [x] **Step 2:** GREEN — implement the marker in `scripts/lib/spec-root.sh` (warn only when `${TMPDIR:-/tmp}/spec-root-dual-tree.<dir with / folded to ->` is absent, then write it best-effort; selection logic byte-unchanged), copy byte-identical to `skills/flow/scripts/lib/spec-root.sh`; harness passes.
  - [x] **Step 3:** verify (block below), then commit with the declared subject and a `Task-Id: 2` trailer.

```bash verified:command shapes from scripts/run-guard-tests.sh's discovery rule and the existing scripts/test-lib-*.sh harness convention
bash -n scripts/lib/spec-root.sh
bash scripts/test-lib-spec-root.sh
bash scripts/test-lib-change-plan.sh
cmp scripts/lib/spec-root.sh skills/flow/scripts/lib/spec-root.sh
```
