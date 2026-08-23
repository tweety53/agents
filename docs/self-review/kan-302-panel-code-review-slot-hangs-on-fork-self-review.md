# Self-review — kan-302-panel-code-review-slot-hangs-on-fork

**Change:** stop the panel's Code review slot hanging on a forked agent
**Issue:** KAN-302
**PR:** https://github.com/tweety53/agents/pull/36 (merged)
**Command:** `/myflow-fast` — creating run through to `FINISHED`, across two invocations

**Shape of the run:** 4 tasks, 12 dispatches (3 implementer, 9 reviewer), 1 panel round at the
`light` roster, 5 per-task findings all fixed, 0 panel findings. Documentation only: 114 lines of
contract text across `skills/myflow-do/SKILL.md` and `SKILL-rationale.md`.

## Problems and fixes — `myflow-fix`

- **[myflow-fix]** `check-task-commit-fields.sh` exits 2 in any repository holding more than one open change — it globs `openspec/changes/*/tasks.md` and refuses on more than one match, so with three changes in flight it never ran, and all four tasks were hand-checked instead — declined
- **[myflow-fix]** `check-unfinished-work.sh` requires a rendered panel record that no `/myflow-do` step renders — section 7 renders `-kind ledger` only, so the omission surfaced at finish run 1 rather than at panel close — declined
- **[myflow-fix]** `/myflow-fast`'s handoff prints `Next: /myflow-fast <name>`, which its own `IN_PROGRESS` branch reads as fix instructions rather than as the integrate route the surrounding prose promises — filed: KAN-311

## Cost — `myflow-cost`

- **[myflow-cost]** All 12 dispatch rows record `Tokens: not measured`, so this change's own cost angle had to be answered from dispatch counts rather than from its record (duplicate of KAN-212, open and urgent; this run's evidence added there as a comment) — declined
- **[myflow-cost]** Nine of twelve dispatches were reviewers and the panel found nothing — every finding came from the per-task layer, because a 114-line prose diff has no cross-commit seam for a whole-branch read to find — filed: KAN-312

## What went well — `myflow-improvement`

- **[myflow-improvement]** The per-task reviewers were given the plan and delta specs as first-class review targets, not just the diff, and three of the five findings were defects in the plan rather than the implementation — a diff-only review would have missed all three, because the implementation faithfully executed what the plan said — declined
- **[myflow-improvement]** Capturing the normative-sentence baseline before the first edit turned "did this change silently reword a requirement?" into a mechanical diff — 3 additions, 0 removals, verified independently by two reviewers — declined

## Automation candidates — `myflow-automation`

- **[myflow-automation]** The delta-spec sync at finish run 2 step 3 is done by hand — it is the one step that edits `openspec/specs/`, it is fully specified, and it was performed here with throwaway Python that nothing verified — filed: KAN-313

## Stats app and storage — `myflow-stats-app`

- **[myflow-stats-app]** The workspace port-block free check `workspace-isolation.md` requires is never performed — `prepare-workspace.sh` printed `MYFLOWD_PORT=5423` with no `lsof` probe, and the daemon already is the registry the probe-and-discard loop exists to substitute for — filed: KAN-314

## Notes on this run

Disclosed at the time; recorded here so the record does not depend on the transcript.

- **The landing route changed after the operator's first answer.** Merge-and-push was selected, which
  conflicts with `~/.claude/rules/no-direct-pushes-to-main.md`. The conflict was raised, and the
  operator chose instead to open a PR and merge it with `gh pr merge` — the same end state, with
  `main` updated only through a merged PR. Run 2 then followed in the same invocation.
- **Filed-issue labels follow the operator's recorded taxonomy, not the contract's inherit-everything
  rule.** `jira-integration.md` says a created issue carries every label on the change's linked issue
  plus `AI-generated`. KAN-302 itself carries `myflow-cost`, having been filed from KAN-295's cost
  angle, so inheriting all of its labels would put two angle labels on each child and make the
  taxonomy meaningless. Each filed issue carries `AI-generated`, `myflow`, and its own angle label.
- **The declared `## stop` command was not run before worktree removal.** It is `docker compose down`,
  which stops the *shared* `myflow-postgres` container. This worktree started no stack;
  `check-worktree-processes.sh` returned `CLEAR`, and the only running `myflowd` (PID 23987) had cwd
  `/Users/tweety53/Projects/agents/stats` — the main checkout. Running the shared teardown would have
  killed work this change never started. Check 6, the gate, passed on its own evidence.
