# self-review-kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy

## Why

The self-review pass of `kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy` found eight
fixable defects in the pipeline, the stats harvester and gymie's dev tooling:

- `verify-fix-loop.md` step 3 ran a fix round's re-review as `primary` alone, under no named key,
  without the always-on extra dispatch, and kept the reduction after a Critical or Important voided
  it — four mis-keyed rounds and a late full-roster round (F1).
- A three-repository change landed with no `link.md`, and no guard noticed; the self-review bundle
  lost the peer repositories' commits and the merge order fell back silently (F2).
- gymie's `verify` and `capture` rows read `GATEWAY_PORT`/`FRONTEND_PORT`, which no agent shell
  carries, so from a worktree they target the main checkout's ports (F3).
- The verifier's motion step allowed sampled `page.screenshot` where a screencast always exists, and
  a one-frame defect slipped past at ~40 ms sampling (F4).
  <!-- measured: kan-995 visual-verification.md "Motion (round 2)" against round 3's screencast, as recorded by the self-review pass @ commit 3cf7aaaa -->
- A full-suite flake, `fidelity/kan-940-cardio-rows.spec.ts:134` (cardio-R1f), was left untraced
  (F5).
- Playwright wipes `test-results/` at every run start, so a one-off failure's trace is gone by the
  time anyone looks (F6).
- Superseded motion strips (`-r1`…`-r4`) were committed into the product repository: 28.6 MB of a
  43 MB archive (F7).
  <!-- measured: du -sh over the kan-995 archive and its motion directories, as recorded by the self-review pass @ commit 3cf7aaaa -->
- flowd's harvester stamped the flake agent's id onto a panel row and the primary reviewer's onto the
  extra-dispatch row: a hook-denied Agent call's transcript line fails to decode, so its denial is
  never seen (F9).

## What changes

- `verify-fix-loop.md` step 3 names the keys, the always-on extra dispatch and the voiding.
- `check-unfinished-work` reports a multi-repository change whose `link.md` `## Merge order` does
  not name every repository its ranges span.
- gymie's `verify` and `capture` rows derive the worktree's ports from the backend worktree's
  `.dev-stack/ports/`, and `verify` and `capture` write each run's artifacts to their own
  `test-results/<UTC stamp>` directory.
- The verifier records motion by CDP screencast only, and a re-record replaces its strip.
- cardio-R1f is reproduced under load with and without `dadc7396` and fixed where the trace points,
  or closed with the measured evidence that `dadc7396` already covers it.
- The harvester decodes a user line whose `toolUseResult` is a string, so hook denials retire their
  begins.
