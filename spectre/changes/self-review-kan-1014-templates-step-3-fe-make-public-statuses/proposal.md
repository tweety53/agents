# self-review-kan-1014-templates-step-3-fe-make-public-statuses

## Why

kan-1014's self-review pass found three non-`big` defects: `run-verify-list.sh` cannot record a
verify row for a multi-repo change's non-canonical worktree, because it passes that worktree as
`-C` and `requireChangeCheckout` refuses a repository holding no `spectre/changes/<name>`;
`flow record finding` accepts an absolute reproducer path that `check-panel-findings-closed` can
never resolve; and nothing tells a gymie planner that detekt's `FunctionNaming` applies to
`desktopTest`, so backticked UI-test names fail lint.

## What changes

- `run-verify-list.sh` takes an optional 8th argument, `<canonical-worktree>`, used as the
  recording `-C` target (default `<worktree>`); `skills/flow/verify-and-handoff.md` documents it
  and the manual `dispatch end` carries `-C`.
- `flow record finding` refuses a reproducer whose first token is an absolute path (exit 2).
- gymie's `.flow/project.md` `## lint` states that `desktopTest` test functions are camelCase.
