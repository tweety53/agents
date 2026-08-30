## Context

This change touches one file, `skills/flow/review-panel.md`, inside its already-existing **Panel
re-runs** section — the fixup-commit-plus-autosquash mechanic and the reproducer-rerun/diff-check
verification that follows it are both already documented there; this change adds one rule that
connects them explicitly. It is one change because it is one paragraph in one place: nothing else
in the pipeline currently claims the connection this rule states.

## Decisions

### State the re-verify-by-content rule explicitly, next to the fixup/rebase instructions

**ID:** state-rule-not-guard
**Status:** active
**Chosen:** Add one bolded rule paragraph to `skills/flow/review-panel.md`'s Panel re-runs section,
immediately after the existing `git commit --fixup=<task-sha>` / `git rebase --autosquash`
instructions (after the current line 354) and before the Targeted/Full mode table — the rule names
the failure mode (a 3-way auto-merge resolving in favour of the pre-fix side, exiting 0 with no
conflict marker and no stray `fixup!` commit) and states that the reproducer rerun and diff-touch
check the section already runs after the fix subagent reports are what must catch it, never the
fixup commit's presence or the rebase's exit code.
**Considered:** A mechanical content-hash check (capture a hash per repaired location before the
rebase, compare after, fail the round on any difference) — ruled out for this change: the
reproducer rerun and diff-touch check the section already runs already read post-rebase file
content, so a hash check would duplicate an existing mechanism rather than close a real gap; the
operator declined it when asked, choosing the documented-rule-only scope.

## Open questions

None.
