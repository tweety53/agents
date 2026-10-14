# self-review-kan-980-no-flow-roles

## Why

The self-review pass over kan-980 found three defects, none `big` — all `flow-fix`, the first also
`flow-stats-app`:

- `flow record pass` accepts any `-note`. In kan-980's panel round 4, F9's mutation proof was recorded as
  a pass note opening `fix-mutation:`, so the round carried no structural `fix-mutation` row and no
  `fix-mutations-total`, and the renderer's `neutraliseMarkers` rewrote the note's colon to U+A789.
- `skills/flow/review-panel-late-fix.md` defines the since-close sha only as the sha the last clean close
  reviewed; once integrate's reshape has rewritten the branch those shas are gone, and kan-980's fix run
  had to improvise one.
- `skills/flow/review-panel-fix-round.md` says nothing about the stage close waiting on a citation
  send-back; kan-980's panel close and verify began before its `-cite` dispatch had returned.

## What changes

- `flow record pass` refuses, exit 2 and before the store is contacted, a note opening with a marker
  label, naming `flow record mutation` and `flow record finding`.
- The late-fix reduction names the since-close sha after a reshape: the canonical worktree's archive
  commit, a peer worktree's reshaped branch head.
- The fix-round contract states that a pending `-cite` send-back holds the stage close.
