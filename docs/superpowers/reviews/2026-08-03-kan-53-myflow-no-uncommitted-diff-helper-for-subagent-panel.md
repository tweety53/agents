# Final whole-branch review panel — kan-53-myflow-no-uncommitted-diff-helper-for-subagent

## Roster

Required: Primary (Sonnet), Bugbot substitute (Sonnet — `bugbot` subagent_type unavailable in this
harness, substituted with a general-purpose defect-hunt reviewer using the same prompt intent),
Principles Merged (Sonnet).
Optional, triggered: Adversarial (>300 changed lines), Principles Lens B — simplicity & state
(>200 changed lines), Principles Lens C — robustness & ops (error-handling-heavy diff, judged
in-scope). Security not triggered (no auth/secrets/query-construction/CORS/dependency surface).

Diff reviewed: `.superpowers/sdd/final-review.diff` (merge base 01dd7b82a4bc6ca153490e0190e1a28c523d19d5).

## Findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Adversarial | Critical | skills/myflow-do/scripts/checkpoint | git stash create doesn't snapshot untracked files; two checkpoints around an untracked-only change collapse to the same value |
| F2 | Principles-Merged, Principles-LensB, Principles-LensC, Bugbot-substitute, Primary | Important | skills/myflow-do/scripts/uncommitted-review-package (and checkpoint after F1's fix) | bare `git add -A` stages openspec/, docs/manual-test/, docs/superpowers/ — forbidden by SKILL.md's NO-COMMITS rule |
| F3 | Principles-Merged, Adversarial | Important | scripts/test-uncommitted-review-package.sh | default-OUTFILE code path (the one every ordinary per-task call uses) had zero test coverage |
| F4 | Bugbot-substitute | Important | scripts/test-uncommitted-review-package.sh | SANDBOXES array registration lost via command-substitution subshell; every scratch repo leaked into $TMPDIR |
| F5 | Primary | Important | .myflow/project.md | new test script missing from the ## test list |

findings-total: 5
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed

## Deferred (from per-task review, triaged here, not load-bearing)

- Task 4: gap that led to F3 — closed by F3's fix (task 4.9 partially addressed it, F3's fix round completed coverage of the default-OUTFILE branch specifically).
- Task 4.9: round-one write assertion relies on `set -e` crash rather than an explicit `[ -e ... ]` check — cosmetic, confirmed non-load-bearing by two independent reviewers.
- Adversarial reviewer's Minor notes (submodule/symlink caveat in design.md, `find -quit` portability assumption, no tracked-file-deletion test case) — genuinely minor, no action taken.

## Fix wave

One consolidated fix dispatch covering F1–F5 (see `.superpowers/sdd/tasks/final-review-fix-report.md`),
followed by one scoped re-review confirming all 5 addressed with no new breakage, plus one
non-functional documentation-comment correction (test file's own header, made directly by the
controller — not a finding, purely descriptive).

## Result

All required and triggered optional slots ran; all 5 findings fixed and re-reviewed clean, matching
the marker block above (5 total, all `fixed`). Full repo test sweep (9/9) and all three lint guards
clean after the fix wave.
