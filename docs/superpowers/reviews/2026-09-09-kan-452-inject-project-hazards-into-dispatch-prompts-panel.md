# Review panel — kan-452-inject-project-hazards-into-dispatch-prompts

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | primary | minor | spectre/changes/kan-452-inject-project-hazards-into-dispatch-prompts/tasks.md:task-2-step-1 | plan text calls the handler-test pattern a real store over httptest; the repo's actual pattern is the in-memory fakeStore, which is what the tests correctly use |
| F2 | primary | minor | stats/internal/api/hazards.go:35 | ApplyHazardRecord refuses with "name and body are all required" — "are all required" fits three-or-more items and reads wrong for two |

findings-total: 2
finding-status: F1 fixed
finding-status: F2 fixed

reproducers-total: 2
finding-reproducer: F1 none — plan-text mismatch verified by reading tasks.md task 2 Step 1 beside stats/internal/api/records_test.go's fakeStore scaffold
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
