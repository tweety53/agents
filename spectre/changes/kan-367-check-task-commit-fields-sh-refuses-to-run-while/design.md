## Context

`check-task-commit-fields.sh` gates every task commit in `/flow`'s implementation phase (step 4 of
`implement.md`): a nonzero exit sends the task back to the same implementer, never a review finding.
It resolves *which* `tasks.md` a task's fields live in by globbing
`<worktree>/spectre/changes/*/tasks.md`, applying a highest-numbered-`-fix-N`-sibling rule to
collapse a fix round's own sub-change into its parent, and refusing (exit 2) when more than one
unrelated root remains. That refusal is correct when the guard genuinely cannot tell which change a
task belongs to — but every caller inside `/flow` already knows: `<name>` is resolved once per run
and threaded through every stage, including the one that invokes this guard.

## Decisions

### Add an optional change-name argument instead of changing the call convention

**ID:** change-name-arg-optional
**Status:** active
**Chosen:** append `<change-name>` as a new, optional 6th positional argument — resolve directly
against it when given, and fall back to today's glob-and-refuse behavior unchanged when it is
omitted.
**Considered:**
- **Make it required.** Rejected: the hand-run fallback (`implement.md`'s "when the script cannot be
  located" path, and anyone invoking the guard by hand) has no `<name>` to hand it, and forcing one
  would turn every such call into a usage error instead of the conservative glob it runs today.
- **Replace the glob entirely, dropping the ambiguity-refusal path.** Rejected: it is still the
  correct behavior for two genuinely unrelated live changes when no name is supplied, and the
  existing test suite (cases 65–67) pins that behavior — removing it would be validating a
  regression it exists to catch.

### Scope the fix-sibling resolution to the named root only

**ID:** scope-fix-siblings-to-named-root
**Status:** active
**Chosen:** when `<change-name>` is given, look only at `$CHANGES_DIR/<change-name>` and its own
`<change-name>-fix-N` siblings — never at any other directory under `spectre/changes/`.
**Considered:**
- **Still enumerate every directory, using `<change-name>` only as a tie-breaker.** Rejected: this
  keeps the O(all live changes) scan and the exact failure mode this change removes — a directory
  with a malformed or unreadable `tasks.md` elsewhere under `spectre/changes/` would still be able to
  affect a call that names its own change explicitly.

### Root-cause the leftover-archive report before designing a fix for it

**ID:** investigate-leftover-archive-first
**Status:** active
**Chosen:** check whether `kan-297` and `kan-363`'s archive PRs existed and were mergeable before
assuming a pipeline defect.
**Considered:**
- **Design a code fix for `check-cleanup-complete.sh` or the archive flow directly**, per the
  ticket's framing. Rejected once the investigation showed both archive commits existed, were
  pushed, and had open PRs (#77, #79) that simply were never merged — the same was true of a third,
  unrelated change (`kan-173`, PR #81), found while checking. There is no pipeline defect to fix;
  merging the three PRs resolved the reported state directly.

## Open questions

None — the round converged with no unanswered questions.
