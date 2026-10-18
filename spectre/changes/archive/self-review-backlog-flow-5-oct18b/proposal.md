# self-review-backlog-flow-5-oct18b

## Why

Five "Backlog Flow" tickets, triaged against `origin/main` `02d83a03` and gymie-frontend
`origin/develop` `755c74ea`, still hold:

- KAN-836 — a fidelity capture reached by a click asserts only that its room is on screen, so a
  missed click inside the room is saved as a baseline, and the pointer left on a control puts hover
  ink in the capture.
- KAN-834 — a capture under a scrim compares dimmed pixels, so a removed layer beneath it can stay
  under the per-pixel threshold and pass.
- KAN-833 — a geometry assertion on Compose's reported bounds passes on unfixed code, because the
  bounds are coerced to the parent; nothing in the implementer's required reading warns of it.
- KAN-832 — gymie-frontend has no `verifyChange` bundle and no `## test` / `## lint` in its
  `.flow/project.md`, so every frontend task's verify is serialized Gradle invocations.
- KAN-827 — flowd replays the CLI's journal only at startup and when its database ping goes from
  failing to succeeding; a mark journalled during a brief HTTP-level outage, the database up
  throughout, strands until flowd restarts, and store-reading guards see a stage that never ended.

## What changes

- `skills/flow/visual-verify-verifier.md` step 8 requires a click-reached capture to assert the
  state it clicked into and park the pointer, and a capture under a dimming layer to pin its subject
  structurally.
- `skills/flow/engineering-principles.md`'s Testing principles name a measured-bound assertion on a
  drawn property as a violation.
- gymie-frontend registers `verifyChange` and declares `## test` and `## lint`.
- `reconcile.Watch` replays the journal on every healthy ping;
  `skills/flow-contracts/state-file-internals.md` states the new replay trigger.
