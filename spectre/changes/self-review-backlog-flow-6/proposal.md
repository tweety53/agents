# self-review-backlog-flow-6

## Why

The next five tickets in KAN's "Backlog Flow" status, oldest first, filed by earlier self-review
passes, are planned here as one self-review fix change. Four are already fixed in the tree
(`design.md`, **Context**); one names a gap the tree still has. The operator then asked for four
more live tickets: the next eleven in the status were checked, oldest first, and the first four
still live — KAN-733, KAN-740, KAN-802, KAN-803 — joined the change:

- **KAN-702** — kan-573's cross-repo visual-trigger match used `origin/main` as the frontend's
  merge base instead of the worktree's recorded one, pulling the files of three archived changes
  into the match set. The step's `<merge-base>` still says nothing about where a satellite
  worktree's base comes from.
- **KAN-733** — gymie's `AccessServiceImpl.setLevel` checks the accepted friendship, then writes
  the grant, in separate steps with no lock; an unfriend landing between them leaves a grant
  between non-friends.
- **KAN-740** — a live-check step the operator waives simply vanishes; kan-576's waived V34
  DOWN-block step, run anyway on a scratch database by a reviewer, caught a rollback that would
  have corrupted grants.
- **KAN-802** — the verifier leaves the clock real for any view the server dates from its own
  clock, though gymie's `pinClock` already pins the server's day with `X-Gymie-Today`.
- **KAN-803** — a re-baseline stages the PNGs `--update-snapshots` wrote without reading them;
  kan-741's reads found nine wrong-screen rewrites.

## What changes

- **Visual verification**'s cross-repo match (`skills/flow/verify-and-handoff.md`) binds
  `<merge-base>` to each worktree's own merge base from the run's working notes, never a
  hand-picked ref.
- **Live check** (`skills/flow/verify-and-handoff.md`) runs a waived step on its cheapest safe proxy.
- The verifier's clock rule (`skills/flow/visual-verify-verifier.md`) pins the server's end too where
  the checkout can.
- **The sweep** (`skills/flow-contracts/known-bugs.md`) and the fix loop's recapture
  (`skills/flow/verify-fix-loop.md`) read every rewritten baseline before staging it.
- gymie's `setLevel` holds the friendship row's lock through its check and write, and `unfriend`
  takes the same lock first.
