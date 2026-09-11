# Review panel — kan-493-flow-gather-self-review-context-sh-rejects-every

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | primary | Minor | scripts/gather-self-review-context.sh:241-242 | the --show-toplevel swap silently flips bare-repository roots from accepted (old check) to refused (exit 2); neither the proposal acceptance set nor any test declares a bare repo, and the flip is entailed by the chosen mechanism rather than decided |

findings-total: 1
finding-status: F1 deferred — Minor under /flow-fast (always deferred): bare-repository roots are outside the documented caller set (archive.md step 9 passes the landing worktree); the acceptance flip is acceptable and revisitable if a bare-repo caller appears

reproducers-total: 1
finding-reproducer: F1 git clone --bare of the repo, then gather-self-review-context.sh with the bare root as repo-root — exits 2 at HEAD, exited 0 at the merge base (command and transcript in .superpowers/sdd/panel-report-0-primary.md)
