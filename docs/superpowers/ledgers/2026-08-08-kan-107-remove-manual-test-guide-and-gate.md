# SDD ledger — plan: openspec/changes/kan-107-remove-manual-test-guide-and-gate/tasks.md

Worktree: /Users/tweety53/Projects/agents/.worktrees/kan-107-remove-manual-test-guide-and-gate
Merge base: 58bff0d
Models: implementation Sonnet, reviewPanel Sonnet, panelFix Sonnet (recorded in state file)

Task 1: fix round 1/5 (3 addressed, 0 open — stale Signal N refs, 8e guide prose, 8f dead fixture)
Task 1: complete (uncommitted, review clean, implementer model: Sonnet, reviewer model: Sonnet, re-reviewer model: Sonnet)
Task 2: complete (uncommitted, review clean, implementer model: Sonnet, reviewer model: Sonnet)
Task 3: fix round 1/5 (1 addressed, 0 open — stale "three planning paths" at pipeline-rationale.md:30)
Task 3: complete (uncommitted, review clean, implementer model: Sonnet, reviewer model: Sonnet, re-reviewer model: Sonnet)
  Scope ruling: six guide-generation sites (myflow-do/SKILL.md:327,536; SKILL-rationale.md:79; commands/myflow-do.md:14; commands-claude/myflow-do.md:10; test-check-references.sh:111) left to Tasks 4/6/7 — upheld by review.
  Deferred note: checkpoint + uncommitted-review-package pathspec edits are behaviour-invisible to test-uncommitted-review-package.sh, since Step 1 stripped the docs/manual-test fixtures that would have exercised them.
Task 4: complete (uncommitted, review clean, implementer model: Sonnet, reviewer model: Sonnet)
  Controller ruling: skills/myflow-do/SKILL.md:3 (frontmatter description) and :8 (intro sentence) still name the guide. Assigned to Task 6, whose sweep covers contract and skill prose. Task 7 covers commands/ and commands-claude/ only.
  Deferred minor: SKILL-rationale.md could state explicitly that run instructions come from `## apps` rather than per plan task; not a defect as shipped.
Task 5: complete (uncommitted, review clean, implementer model: Sonnet, reviewer model: Sonnet)
  Controller ruling: skills/myflow-status/SKILL.md:112 (example table row) and :158 (next-command mapping, spec prose) still say "run the guide". Assigned to Task 6 with the rest of the skill prose sweep. Reviewer marked :158 Important.
Task 6: fix round 1/5 (2 addressed, 0 open — SKILL.md:8 "stage both", stale numeric counts)
Task 6: complete (uncommitted, review clean, implementer model: Sonnet, reviewer model: Sonnet, re-reviewer model: Sonnet)
  Incidental correction: check-plan-provenance.py's "110 files" figure was already stale before this change; corrected to 216 with the method stated. A third unreproducible figure at plan-provenance.md:293-297 was removed rather than restated.
Task 7: complete (uncommitted, review clean, implementer model: Sonnet, reviewer model: Sonnet)
Task 8: complete (uncommitted, no files changed — verification pass, all 12 harnesses + 6 lint guards + openspec validate --strict green on first run, implementer model: Sonnet)
  No task-diff review dispatched: the task changed no files, so its evidence is the report, not a diff.
