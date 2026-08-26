# Review panel — kan-337-fix-three-flow-finish-phase-friction-points

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Primary | Important | skills/myflow-contracts/finish-contract.md:384 | The new paragraph claims the journaled state write at step 8 is another instance of 'that same stack's absence showing up again' after check 5 stops the project's declared stack. This is a category error: the state write's journal/fallback path (per state-file.md) triggers on the myflow pipeline's own state store (myflowd daemon + Postgres at a fixed, project-independent path) being unreachable, which has no causal relationship to a given project's own docker/dev stack being stopped. Only the workspace-removal failure (step 5) and the cleanup-verification SKIPPED clause (step 7) are legitimately caused by the project's stack being down; the step-8 state write must be dropped from this list. |

findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 .superpowers/sdd/repro/F1.sh
