# Review panel — kan-298-preflight-accepts-bare-local-branch-as-base-ref

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Bugbot | Important | scripts/check-finish-preflight.sh:54 | If the `>&2` redirect is dropped (or ever regresses), the usage text lands on stdout instead of stderr — violating the script's own stated contract — and no test in either harness catches this: the missing-argument test case checks only the exit code, never which stream the text arrived on, and the new KAN-298 assertion greps the source file, not runtime output. |
| F2 | Bugbot | Minor | scripts/check-finish-preflight.sh:60 | Removing the explicit `exit 2` still yields exit 2 because empty $WORKTREE fails the very next `[ ! -d "$WORKTREE" ]` check. Pre-existing behavior, not introduced by this diff — no fix required, noted only because it is a surviving mutant against the "missing arguments -> exit 2" test. |
| F3 | Principles | Minor | scripts/test-check-base-moved.sh:504 | WET: the two `EXPECTED_USAGE` literals are identical apart from the guard name, mirroring the guards own near-duplication (already an accepted design tradeoff, "both-guards-share-the-defect"). Not worth extracting for two call sites under KISS — noted only as a maintenance cost: a future wording edit now touches two test literals plus two source-grep phrases plus two guards (four sites, all mechanically necessary, none silently divergent). |
| F4 | Primary | Important | spectre/changes/kan-298-preflight-accepts-bare-local-branch-as-base-ref/tasks.md:9 | Task 1 and task 2 **Files:** fields omit `scripts/lib/base-ref-usage.sh`, which task 1 commit actually created and task 2 harness sources. The plan is out of step with the branch and needs correcting before handoff. |
| F5 | Primary | Important | scripts/lib/base-ref-usage.sh:36 | The worktree carried two uncommitted stray mutations during this review: check-base-moved.sh had an extra space before (origin/main), and lib/base-ref-usage.sh had its last line replaced with STALE MUTATED LIB. Consistent with a mutate-verify-restore cycle whose restore step never completed. |

findings-total: 5
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 withdrawn investigated under fix round 3 — no residue exists in any commit, in all history via pickaxe, in the index, in the working tree or in untracked files; the branch is byte-exact and the suites are 20/20 green, so no defect remains for a further round to act on

reproducers-total: 5
finding-reproducer: F1 sed -i.bak "s/cat >&2 <<'EOF'/cat <<'EOF'/" scripts/check-finish-preflight.sh; ./scripts/test-check-finish-preflight.sh; echo rc=$?; mv scripts/check-finish-preflight.sh.bak scripts/check-finish-preflight.sh
finding-reproducer: F2 none — pre-existing redundancy independently confirmed via manual mutation of check-finish-preflight.sh:60-61, restored, not a regression from this diff
finding-reproducer: F3 none — a style observation, not a runnable check
finding-reproducer: F4 none — static comparison of tasks.md Files fields against the commit stat, independently confirmed by check-task-commit-fields.py exiting 1 on task 1
finding-reproducer: F5 none — a transient worktree state observed during review, not a property of any commit
