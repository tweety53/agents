# Review panel — kan-279-give-the-parent-a-mutation-helper

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | code-review-low | minor | scripts/mutate-and-verify.sh:303 | unconditionally prints "ran clean — restoring touched files" before the EXIT trap attempts restore, so a failed restore (e.g. an added untracked file git checkout -- cannot remove) shows a contradictory "ran clean" immediately followed by "could not fully restore" |
| F2 | bugbot | minor | scripts/mutate-and-verify.sh:293 | the caught/suspicious-blast-radius boundary (new_count -le BOUND) has no fixture at new_count == BOUND exactly, so an off-by-one there is a surviving mutant |
| F3 | bugbot | major | scripts/mutate-and-verify.sh:134 | the could-not-fully-restore / exit-3 path has zero test coverage — no fixture leaves a touched file un-checkoutable to trigger the residual-dirty branch |

findings-total: 3
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed

reproducers-total: 3
finding-reproducer: F1 none — reviewer supplied a multi-step shell script, not a single runnable in-worktree command; reproduced by hand against a scratch git repo with a new-file patch
finding-reproducer: F2 none — demonstrated by flipping the -le to -lt at line 293 and rerunning scripts/test-mutate-and-verify.sh; not a single runnable command
finding-reproducer: F3 none — demonstrated by removing the rc=3 override in the EXIT traps residual-restore-failure branch and rerunning scripts/test-mutate-and-verify.sh; not a single runnable command
