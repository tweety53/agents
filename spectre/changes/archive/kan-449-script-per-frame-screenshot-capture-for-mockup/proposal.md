# kan-449-script-per-frame-screenshot-capture-for-mockup

**Jira:** [KAN-449](https://tweety53.atlassian.net/browse/KAN-449)

## Why

`rules/design-mockups-are-specs.mdc` asks for the app to be run, screenshotted, and compared against
the mockup side by side. `flow.visual-verify` (KAN-171) mechanized everything up to the comparison:
it starts the stack, runs the per-change Playwright spec and resolves its PNGs. The comparison
itself is still a verifier holding a captured frame and a drawn frame in its head and writing prose
about the pair into `visual-verification.md`. KAN-423 did that for nineteen frames
(`gymie/docs/design/group-sessions/screens/G1.png … K3.png` against
`gymie-playwright/tests/group-session-mockup-*.spec.ts-snapshots/*.png`) — unreproducible, and
invisible to the review panel and the handoff, which read the prose but never see the two images
it describes. KAN-423's self-review filed this change.

## What changes

- A new guard-shaped script, `scripts/compose-mockup-frames.sh` (+ `.py`), that takes the resolved
  capture PNG paths on stdin, a `<spec>.mockups` sidecar map of `<screenshot name> <frame id>`
  pairs, a mockups root and an output directory, and writes one captured-left / mockup-right PNG per
  pair — with a mutation-tested harness, `scripts/test-compose-mockup-frames.sh`.
- `## visual verification` gains an optional `mockups` setting (the directory of `<frame id>.png`
  files), canonical in `skills/flow-contracts/project-configuration.md`;
  `check-visual-verification.sh`'s closed `Setting` vocabulary admits it, and its harness pins that.
- `flow.visual-verify` (`skills/flow/verify-and-handoff.md`) gains a step between "read every
  captured PNG" and "write `visual-verification.md`": compose, read every composite, and list each
  pair in the report; composites are committed under `<changeRoot>/visual-verification/`. The stage
  report gains `mockups:` and per-frame lines; blocking gains a compositor failure and a reported
  departure from the mockup.
- `gymie/.flow/project.md` declares `mockups: docs/design/group-sessions/screens`.

Design, decisions and the open question: `design.md` beside this file, mirrored at
`docs/superpowers/specs/2026-09-10-kan-449-script-per-frame-screenshot-capture-for-mockup-design.md`.
