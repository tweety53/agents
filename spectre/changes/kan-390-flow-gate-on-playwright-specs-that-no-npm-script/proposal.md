# kan-390-flow-gate-on-playwright-specs-that-no-npm-script

## Why

KAN-29's self-review found seven route A–F capture suites in `gymie-playwright` that never ran
once: `test:e2e` reaches `tests/`, `test:e2e:full` reaches `tests-full/`, and the root
`kan-*.spec.ts` files `flow.visual-verify`'s `capture` writes match neither. One (`kan-347`)
passed green against a 404 page for its entire life, found only when an ink change moved a
baseline in a suite nobody had run. Nothing in the pipeline asks whether a spec it just wrote is
reachable by anything the project runs.

## What changes

- A shipped guard, `scripts/check-spec-reach.sh <project-root>`, enumerates every `*.spec.ts`
  under the project's `regression checkout` and diffs it against what each `package.json` script
  invoking `playwright test` actually lists (`--list --reporter=json`); orphans fail it.
- `flow.verify` runs it after the `## test` commands, every change; `flow.visual-verify` runs it
  again right after `capture`. Non-zero blocks the handoff at both.
- A project with no `regression checkout` prints `Spec reach: not configured` and passes.
- Test harness, symlink, guard-presence list, `## lint` entry, budget rows.

Design, decisions and rejected alternatives: `design.md` beside this file.
