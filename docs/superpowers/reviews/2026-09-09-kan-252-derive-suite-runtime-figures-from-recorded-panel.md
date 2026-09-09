# Review panel — kan-252-derive-suite-runtime-figures-from-recorded

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | mutation | minor | stats/cmd/flow/suite_test.go | the median summary value is unpinned: flipping the median to the upper middle survives the suite, so the summary can silently change meaning |
| F2 | mutation | minor | stats/internal/store/suiteruns_test.go | the store boundary validation (ErrInvalidSuiteRun) is uncovered: deleting the guard leaves every test green |
| F3 | mutation | minor | stats/internal/api/suites_test.go | the absent-limit default and the bad-limit refusal are uncovered: zeroing the default survives the suite |
| F4 | mutation | minor | stats/cmd/flow/suite_test.go | the child-could-not-start path (exit 127, nothing recorded) is untested: changing the exit code survives the suite |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-mutation-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-mutation-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-mutation-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-mutation-4.sh
