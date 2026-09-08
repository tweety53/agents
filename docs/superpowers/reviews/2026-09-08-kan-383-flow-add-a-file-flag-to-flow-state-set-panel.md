# Review panel — kan-383-flow-add-a-file-flag-to-flow-state-set

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | mutation | minor | stats/cmd/flow/state_test.go:TestStateSetFileFlagMissingFileIsUsageError | surviving mutant: with the read-error return dropped, a missing -file falls through to the non-object branch whose message also names the path, so no test distinguishes a read error from a format error |

findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 .superpowers/sdd/reproducers/0-mutation-1.sh
