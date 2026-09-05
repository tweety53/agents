# Review panel — kan-433-flow-verify-never-runs-lint-test-inline-build

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | principles | Minor | skills/flow/implement.md:166-171 | "run every printed line" for `## worktree setup` does not hedge for `project-get.sh` printing the whole section body — fence markers and this repository's own trailing rationale prose included, not just the command's own text — which the key's own contract at project-configuration.md:29 promises is "one command per line inside the fence" |
| F2 | mutation | Minor | stats/cmd/flow/record_test.go:1962-1967 | TestRecordUsageNamesEveryRole only pins usage-text-contains-every-role, never the reverse — dropping `verifier` from `recordRoles` while `recordUsage` still advertises it leaves the CLI rejecting a role the documented usage claims is valid, with zero test signal |

findings-total: 2
finding-status: F1 fixed
finding-status: F2 fixed

reproducers-total: 2
finding-reproducer: F1 none — the defect is that project-get.sh exits 0 while printing more than the worktree setup contract promises (fence markers plus trailing prose), not that any command fails; no non-zero-exit reproducer demonstrates an over-print
finding-reproducer: F2 none — a surviving mutant is a case where every harness exits 0 both before and after the mutation, so no command can exit non-zero to demonstrate it; mutate-and-verify.sh also always exits 0 on a clean run regardless of per-harness verdict
