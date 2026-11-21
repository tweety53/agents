# Visual verify

A green test suite does not notice a wrong fill colour or a missing border. When a change can alter
what a user sees, an agent runs the app and looks.

## When it runs

A project opts in by declaring a `## visual verification` section in its `.flow/project.md`: which
paths are UI, and the commands that run its screenshot suite. Then, per change:

- the plan review decides `required` or `skipped`
  ([Planning](planning.md#the-decision-how-it-is-built-and-reviewed));
- a guard matches the change's diff against the declared UI paths, and a diff that touches none
  skips the stage.

A project with no such section is `not configured`, and the stage never runs.

## What it does

1. **Pre-flight.** Before any agent is started, guards check that lint and tests are green on this
   exact commit and that the app's ports, URLs and browser tooling belong to this change's own
   workspace.
2. **Verify.** One verifier agent per worktree runs the project's checked-in screenshot suite, the
   regression gate.
3. **Capture.** It runs the change's own spec, which takes screenshots of the screens the change
   touched.
4. **Compare.** Where the project declares mockups, each capture is composed beside its mockup
   frame with a difference panel. The difference ratio is information, never a pass threshold: the
   verifier reads every capture.
5. **Report.** The verifier fixes nothing and edits no source. A defect it reports blocks the
   change and sends it back to be fixed, then verified again.

Animations and transitions the change adds, changes or removes are checked too. A flagged change
that draws nothing itself, such as a response a page consumes, is checked against the payload it
promised instead.

Canonical: **Visual verification** in
[`skills/flow/verify-and-handoff.md`](../skills/flow/verify-and-handoff.md) (the trigger),
[`skills/flow/visual-verify.md`](../skills/flow/visual-verify.md) (the stage),
[`skills/flow/visual-verify-verifier.md`](../skills/flow/visual-verify-verifier.md) (the verifier's
steps) and
[`skills/flow-contracts/project-configuration-visual.md`](../skills/flow-contracts/project-configuration-visual.md)
(what a project declares).
