# kan-1028-mutation-failure-modes-extra-dispatch

## Why

Failure-modes and mutation review only ran on a `full`-shape panel of a `regular`/`big` plan — a 10% compact roll — so most changes, and every `small` and `micro` one, shipped without either angle. The operator wants both on every run, as `test-sanity+call-efficiency` already is (kan-959).

## What changes

- Every review-panel pass 1, on every class (`micro` included), dispatches a second always-on bundle, `mutation+failure-modes`, beside `test-sanity+call-efficiency`: outside the ≤2 × ≤3 cap, never removed by the docs-only reduction, the late-fix reduction or the small append, and re-run after a fix like any role.
- Its pair is rolled on its own, like the first extra's: `sonnet` when `roll(<name>extramf) < 50`, else `opus`, always `low`. `plan-class.sh` prints it as a second `extra:` line; Decide records it as `extra_dispatch_mutation`; `flow decision render` prints it; `flow record decision` refuses an off-policy model in it; the reviewers view counts its slots.
- `failure-modes` and `mutation` leave the Decide tree's roster, so no role runs twice: the tree roster is `primary; principles`, plus an experimental slot on a `full`-shape `regular`/`big` panel, which now joins the first dispatch. A `default` panel's resolved roster drops any role an always-on extra dispatch carries.
- The learning note is written to `<project>-learning/<name>.md` — a folder beside the main checkout, like `<project>-worktrees`, outside every git repository — and its absolute path is printed as soon as the note is composed, at the start of implementation; the handoff names the path instead of printing the section. Operator's request mid-run (2026-10-23).
