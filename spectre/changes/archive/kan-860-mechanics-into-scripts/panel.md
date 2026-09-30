# Review panel — kan-860-mechanics-into-scripts

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | skills/flow/review-panel-fix-round.md:340 | The fix-round close still invokes check-task-commit-fields.sh but 144de4e9 deleted skills/flow/scripts/check-task-commit-fields.sh, so the call no longer resolves under Guard resolution; check-guard-symlinks misses it because the invocation wraps. |   |
| F2 | primary | Important | skills/flow-fast/SKILL.md:87 | flow-fast runs implement.md section 4 except that check-task-commit-fields.sh is not run and no task is ticked, but section 4 close is now one close-task.sh call that exits 2 in flow-fast layout, so the carve-out cannot be followed. |   |
| F3 | primary | Important | stats/internal/guard/removechangeworktrees.go:244 | Check 4 applies the .png/.jpg rule before the regeneratable-directory rule, so ignored images under node_modules/, dist/, build/ or .superpowers/sdd/ are UNCLASSIFIED and force the disclosure stop. |   |
| F4 | primary | Important | stats/internal/guard/removechangeworktrees.go:261 | A ## stop body without a code fence is silently skipped and the worktree force-removed, while project-configuration.md:25 still defines ## stop as The command with no fence requirement. |   |
| F5 | primary | Minor | .superpowers/sdd/relocation-comparison.md | relocation-comparison.md is absent although tasks.md declares Relocation: yes. |   |
| F6 | principles | Important | skills/flow/review-panel-fix-round.md:340 | Invoking check-task-commit-fields.sh from skill text after deleting its skills/flow/scripts symlink breaks check-guard-symlinks rule 2 and the tasks.md:15 convention; the guard passes only through its wrapped-line blind spot. |   |
| F7 | principles | Important | skills/flow-contracts/finish-contract-run2.md:341 | Single Source of Truth: the ## stop key shape is now defined twice and differently — run 2 and the guard require a fence, the canonical table in project-configuration.md:25 does not. |   |
| F8 | principles | Minor | stats/internal/guard/baserebase.go:69 | DRY: the rebase-in-progress check is written three times (baserebase.go:69, foldfixup.go:210, syncontobase.go:70) though ffRebasing exists. |   |
| F9 | principles | Minor | stats/internal/guard/kickoffworktree.go:40 | Single Source of Truth: two new readers of a fenced command block in .flow/project.md (kickoffworktree.go:40, removechangeworktrees.go:261) disagree on block selection and fence detection. |   |
| F10 | principles | Minor | scripts/plan-class.sh:99 | DRY: the agents-repo-root derivation is pasted into plan-class.sh:99 and sync-onto-base.sh:67 though lib/flow-guard.sh:79 already derives it. |   |

findings-total: 10
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 withdrawn no tree change can produce it — an existing rule makes symlinked Files uncomparable; the generator now fails loudly and round 0 compared the relocation by hand
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed

reproducers-total: 10
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: F5 none — a missing process artifact; the relocation itself was compared verbatim and is intact
finding-reproducer: F6 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F7 .superpowers/sdd/reproducers/0-principles-2.sh
finding-reproducer: F8 none — the three copies agree today
finding-reproducer: F9 none — needs a project config shape no project here has
finding-reproducer: F10 none — the copies agree today

## Pass log

### Round 0

- base moved: 6 commits on origin/main, overlap; a rebase would conflict in plan-class.sh, planclass.go, plan_class_test.go, brainstorm-planner.md — operator said continue; Continue — review as is, conflict left to integrate
- roster: compact — true; diff 13358 lines over cap, proceeded unasked; docs-only exit 1 (scripts/aside-planning-artifacts.sh); dispatched primary+principles on opus/high; standards CLAUDE.md, AGENTS.md; no operator addition this round — the resolved list ran alone

### Round 1

- fix round 1: one panel-fix chunk (10 findings, 6 Important + 4 Minor) on the decision fixer pair opus/high; reads final-review.diff findings; no bounces
- auto-decided F5: withdrawn — the comparison cannot be generated for symlinked Files under an existing rule; the generator now fails loudly
- round boundary: base still MOVED with overlap — Continue, review as is (same decision as entry); re-runs: primary and principles each alone on rerun pair sonnet/low reading fix-round-1.diff plus their finding sites
fix-mutation: stats/internal/guard/guardsymlinks.go — the unclosed-span classification reverted — TestCheckGuardSymlinks/3m
fix-mutation: scripts/check-guard-symlinks.sh — rule 2 wrap fix reverted, run with the symlink removed — 0→1 rule 2 violations
fix-mutation: skills/flow/scripts/check-task-commit-fields.sh — none — a symlink with no logic; reproducers 1→0
fix-mutation: skills/flow-fast/SKILL.md — none — prose only; reproducer 0-primary-2 1→0
fix-mutation: stats/internal/guard/removechangeworktrees.go — image-first branch re-inserted — TestRemoveChangeWorktrees/clean_removal
fix-mutation: stats/internal/guard/removechangeworktrees.go — SKIPPED print removed — TestRemoveChangeWorktreesStopCommand/no_fence:_skipped
fix-mutation: stats/internal/guard/removechangeworktrees.go — kwFenced replaced by the old first-block reader — TestRemoveChangeWorktreesStopCommand/fenced:_every_fence
fix-mutation: skills/flow-contracts/project-configuration.md — none — prose only; reproducers 1→0
fix-mutation: stats/internal/guard/baserebase.go — inProgress found-marker return flipped to false — TestRebaseOntoTip, TestFoldFixupEmptyDrop/tip, TestSyncOntoBaseResume
fix-mutation: scripts/lib/flow-guard.sh — cd -P dropped from flow_guard_root — TestFlowGuardRootThroughSkillSymlink
fix-mutation: scripts/generate-relocation-comparison.py — Files read back through parse_tasks task.files — test-generate-relocation-comparison.sh case ii
fix-mutations-total: 11
