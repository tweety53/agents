# Generic visual verification, with Gymie bound to Playwright

**Jira:** [KAN-171](https://tweety53.atlassian.net/browse/KAN-171)

## Why

KAN-16 shipped three defects invisible in a diff and obvious the moment the page was opened — a
400 on the run-detail dashboard, a white page where a dark palette was asked for, every stat panel
printing its label twice. All three survived a five-pass review panel, 296 Go tests and 110 SPA
tests, and took about two minutes to find in a browser.

`rules/design-mockups-are-specs.mdc` already says to run the app and screenshot the page. It was not
followed. This makes it a stage with an artifact instead of a rule to remember.

## What changes

- A new optional `## visual verification` key in `<project>/.flow/project.md`, canonical in
  `skills/flow-contracts/project-configuration.md`: declared UI path globs, a baseline command, a
  per-change capture command, and an optional regression checkout with an explicit push permission.
- A new `flow.visual-verify` stage in `skills/flow/verify-and-handoff.md`, between `flow.verify` and
  `flow.stage-diff`. It starts the stack if nothing answers, runs the baseline suite, captures the
  views this change touched, **reads every PNG**, writes `visual-verification.md`, commits to the
  regression checkout, and stops only the stack it started. It blocks the `IN_PROGRESS` handoff.
- A new guard, `check-visual-verification.sh`, with its mutation test.
- `gymie/.flow/project.md` gains the section, bound to `gymie-playwright` with an explicit,
  repository-scoped push permission.
- `stats/web` gains `@playwright/test` and a baseline covering the dashboard and run-detail views —
  the two KAN-16 broke — against the existing `make ui-test-up` fixture stack.

Design, decisions and rejected alternatives:
`docs/superpowers/specs/2026-08-27-kan-171-generic-visual-verification-step-design.md`.

## Added after the first human gate

The stage's ten steps were prose executed by an agent, with no test running them. The operator
required that everything visible be tested before handoff, and that e2e coverage be added:

- **Tasks 9–11** extract the stage's mechanizable logic into guards with mutation-tested harnesses —
  the trigger match, the screenshot resolution, and the push gate. What a script can decide should
  not be prose an agent re-derives each run.
- **Task 12** expands the Playwright baseline from three view-level specs to every control the SPA
  exposes, plus combinations.
- **Task 13** adds a rule the operator asked for directly: after a fix, reload the applications the
  handoff names, so the diff under review and the app being run are the same code. Scoped to a
  project's own applications — never the flow dev stack, which `CLAUDE.md` protects.

