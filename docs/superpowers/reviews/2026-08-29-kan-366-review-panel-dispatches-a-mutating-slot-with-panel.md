# Review panel — kan-366-review-panel-dispatches-a-mutating-slot-with

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Code review (low) | Major | skills/flow/review-panel.md:71 | git diff --binary emits nothing when there are no tracked changes, so the piped git apply exits 128 before the untracked-file copy loop runs; the prose claims the sequence behaves exactly as documented unconditionally |
| F2 | Code review (low) | Major | skills/flow/review-panel.md:72-74 | the untracked-file copy loop breaks on any filename git status --porcelain quotes (spaces, non-ASCII); awk captures the quote characters into $f, so cp -a fails and the file is silently dropped from the throwaway copy |
| F3 | Primary | Major | skills/flow/review-panel.md:197 | git diff --binary \| git apply fails with exit 128 whenever the apply worktree has no uncommitted changes at dispatch time, which is the common case since task/fix-round work is committed; same root cause as F1 |
| F4 | Principles | Critical | skills/flow/review-panel.md:71 | git diff --binary (unstaged only) silently drops staged changes when transplanting, breaking the design's own guarantee that Bugbot sees exactly the same working tree the other slots see |
| F5 | Bugbot | Major | skills/flow/review-panel.md:203-213 | git diff --binary \| git apply fails outright (exit 1, nothing applied) when the transplanted diff contains a rename, and the documented sequence has no exit-code check to catch it |
| F6 | Bugbot | Minor | skills/flow/review-panel.md:203-213 | confirms F2 independently: the untracked-file copy loop breaks on quoted filenames (spaces/special chars) from git status --porcelain |
| F7 | Bugbot | Minor | scripts/check-contract-budget.sh:194 | the artifacts-registry.md budget row (8447) is a surviving mutant — the guard also passes at the old value (7620) since the file is only 6757 bytes; not a functional defect |

findings-total: 7
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 withdrawn — raising a budget row ahead of strict need is explicitly harmless per check-contract-budget.sh's own stated ratchet policy (a budget is 'the size at landing plus 25%', a floor for future growth, not a tight bound); no functional behavior regresses and the row is not wrong, only generous

reproducers-total: 7
finding-reproducer: F1 none — requires a scratch worktree fixture with only untracked files, not runnable inline
finding-reproducer: F2 none — requires a scratch worktree fixture with a quoted filename, not runnable inline
finding-reproducer: F3 none — requires a scratch git repo fixture, not runnable inline
finding-reproducer: F4 none — requires a scratch git repo fixture with a staged change, not runnable inline
finding-reproducer: F5 none — requires a scratch git repo fixture with a rename, not runnable inline
finding-reproducer: F6 none — requires a scratch git repo fixture with a quoted filename, not runnable inline
finding-reproducer: F7 none — mutation testing performed manually against the guard, not a standalone runnable command
