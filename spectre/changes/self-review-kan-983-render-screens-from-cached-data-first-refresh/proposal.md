# self-review-kan-983-render-screens-from-cached-data-first-refresh

## Why

kan-983's self-review pass found five pipeline and harness defects, none `big`, and planning found
a sixth: `origin/main`'s Go guard suite is red (`design.md`, **Context**).

- `run-verify-list.sh` runs every `## test` fence line literally, so gymie's and gymie-frontend's
  verify-bundle templates — `./gradlew verifyChange :<module>:test --tests '<selectors>'` — fail
  every run: the shell reads `<module>` as a redirect.
- `flow.visual-verify` dispatched its first verifier before the `regression checkout`'s worktree
  existed; the verifier blocked on the missing checkout.
- A verifier covering a sibling worktree through the canonical `ui paths` recorded only its own
  HEAD as `-diff-base`, so `check-visual-verify-dispatched.sh` found nothing for the sibling and the
  parent wrote a phantom `visual-verify-2-gymie-frontend` row.
- gymie-playwright's `group-session-journey.spec.ts` times out intermittently in `addToBuilder`,
  waiting for `add_participant_workout_field`.
- `visual-verify-2` re-ran the whole spec list at an unchanged HEAD only to clear one spec the
  previous run had already re-run clean.
- `TestNoPipeIntoEarlyExitGrep` fails on `scripts/test-check-task-records.sh`'s two
  `printf … | grep -q` pipes, landed by `646ac917`.

## What changes

- `run-verify-list.sh` refuses a list holding a `<word>` placeholder line, exit 2, before anything
  runs; gymie's and gymie-frontend's `.flow/project.md` state the verify-bundle template in prose
  outside the fence; `project-configuration.md`'s `## test` row says fence lines run literally.
- Visual verification's step 3 creates a missing `regression checkout` worktree through the
  cross-repo recipe before the pre-flight.
- A verifier's `-diff-base` names every worktree HEAD it covers, comma-joined; the guard judges
  each sha on its own.
- gymie-playwright's `addToBuilder` is fixed at the source of its race, opened on the failing trace.
- A re-dispatch at an unchanged HEAD names `rerun specs:`, and the verifier re-runs only those; a
  non-zero exit still blocks.
- `test-check-task-records.sh` reads `$OUT` through a here-string, and the Go guard suite is green.
