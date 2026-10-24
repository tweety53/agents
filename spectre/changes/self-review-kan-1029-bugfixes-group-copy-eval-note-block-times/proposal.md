# self-review-kan-1029-bugfixes-group-copy-eval-note-block-times

## Why

The self-review pass of `kan-1029-bugfixes-group-copy-eval-note-block-times` found nine fixable
defects:

- F1: `flow.verify` closed `completed` with no `check-live-verification.sh` chained before it
  (stage run 17075) — the gate lives only in prose.
- F2: gymie-frontend's `## test` (`./gradlew test allTests`) runs `:shared:linkDebugTestIosSimulatorArm64`,
  which needs Xcode, on a project that ships PWA and web only.
- F3: `render-slot-prompt.sh` never renders the `mutation` pass; since kan-1028 every pass 1 sends a
  `mutation+failure-modes` bundle, and kan-1029's round 2 ran `failure-modes` alone.
- F4: `hooks/enforce-agent-baseline.py` accepts any lowercase first word as a row key, so launches
  described `verify-fix-1 …` and `verify live check …` went out with no dispatches row.
- F5: a `flow record dispatch begin` inside a `for` loop keyed `-key ${k#*:}` never got its agent
  id — rows 7367 and 7368 stayed unattributed.
- F6: `-session-token mf-k1029-PLACEHOLDER` was accepted (stage runs 17078 and 17079), though the
  token must end in the run's UTC start time.
- F7: gymie's `devPhone` reverses 3000/8080/9000 to themselves, so a workspace whose frontend and
  gateway moved is unreachable from the phone.
- F8: nothing warns that `./gradlew --stop` kills the dev stack's frontend; gymie carries no hazard.
- F9: `rebuild-full-app-suite-zip.sh` zipped into the backend main checkout given as
  `<screenshots root>`, leaving `/Users/tweety53/Projects/gymie/full-app-suite.zip`.

## What changes

- `flow stage end` refuses a completed `/flow` `flow.verify` end until the live-verification check
  passes on the change's directory.
- gymie-frontend's suite drops the iOS simulator test run and keeps the iOS test compile.
- The mutation pass is rendered from the mutation-testing brief, naming its throwaway copies.
- The hook accepts a description's first word only in a Key shape the closed list names.
- The harvester reads `${var#…}`/`${var%…}` strips of a loop variable; rows 7367/7368 are restamped.
- `flow stage begin` and `wrap` refuse a token without the UTC suffix; rows 17078/17079 are
  retokened by a new `flowd retoken-stage-run`.
- `devPhone` reverses the workspace's resolved frontend and gateway ports and 9000.
- gymie gains a hazard row naming `./gradlew --stop`.
- The zip rebuild refuses a root no resolved PNG lies under.
