# kan-1034-spectre-optional-bundled-planning-prompts-that — session narrative

## 2026-10-26 — creating run

The implementation run started from a recorded plan and decision (regular, inline). The spectre repository had no worktree yet, so kickoff-worktree.sh --from created it and a `link.md` carrying only `## Merge order` (agents first, then spectre) was committed.

Task 1 moved the planning paragraphs by line range with a script, then repaired the residue by hand: a Load directive at the top of brainstorm-planner.md, a `### Convergence` heading for the recommended-mode lines left behind, and re-pointed citations of the headings that vanished (Decisions, Open questions, Stage exit). Guards beyond the plan's Files surfaced along the way: check-guard-symlinks wanted the skill to carry `check-plan-shape.sh` and `lib` symlinks, check-stage-mark-calls wanted the skill declared expected-zero, and TestLoadSets broke on the extended load set until its fixture held the new file. Each was transcribed as a Files widening plus a dated Correction. check-plan-shape refused `**Tests:** **none** — repairs existing tests:` because the bold form does not match the repair regex; the plain `none —` form passes. Task 2's Baseline comment used `→ 1 at e093bad` instead of `@ e093bad`, so close-task's re-measurement counted something else; the comment was corrected in place.

The relocation comparison could not be generated: generate-relocation-comparison.sh refuses a symlink among the scoped Files, so the panel audited the move from the diff and check-verbatim-moves.sh instead.

The panel's primary and principles roles found the vendored skill still cited flow files, flow guards and flow stages — the plan had classed those paragraphs as generic. The operator chose to move every flow-dependent paragraph back into brainstorm-planner.md (decision planning-skill-flow-free). The alternative — keeping the references and documenting them in spectre's README — was rejected because a spectre-only user would follow Load directives to files the plugin does not ship. The move-back also dropped the skill's guard symlinks (nothing cites them now), so check-guard-symlinks needed an expected-zero declaration instead. Mutation found the load-set entry unpinned (a 0-byte fixture file); the fixture now carries 8 bytes. The round-1 re-run left three Minors, fixed inline.

At verify, check-task-records failed task 2 under the exported workspace environment: its `flow state get` follows FLOW_ADDR to the workspace daemon, which holds no state record, so the cross-repo peer lookup found nothing. Re-run against the records address it passes; the substitution is recorded and in-run pipeline fix 31 repairs the guard.

## 2026-10-26 — integrate run

In-run pipeline fix: 36eb356ccb6d8002014382c42900e8380d6fa889 — check-task-records read a cross-repo peer's state from the workspace daemon under verify's exported FLOW_ADDR (blast radius 2 files)

The loop planned, fixed, reviewed clean and landed on main through pipeline-fix.sh, which moved main under this change. spectre's UPSTREAM.md named the task 1 commit sha, which integrate's reshape would orphan; it now names the skill by its path on agents main and the ticket that landed it.
