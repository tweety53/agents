# pipeline-fix-kan-1062-flow-implement-one-shot-subagents-50

## Why

`check-panel-fix-single-dispatch.sh` refuses a chunked fix round's later-chunk row
`panel-fix-<r>-<n>` whose recorded commit carries `Flow-Key: panel-fix-<r>`, the round's key —
the trailer `skills/flow/review-panel-fix-round.md` (**The fix dispatch**) prescribes ("A fix
commit's `Flow-Key:` trailer names that same fix round") and `skills/flow-contracts/git-boundaries.md`
(**Commits outside every task**) keys a commit to. Run kan-1062 (session mf-k1062-20261031T063115Z)
hit it on row `panel-fix-1-2`, commit `7d3de8536059b219be89a65729b0b33b2a9cd3ba`.

## What changes

- `pfdCheckCommits` (`stats/internal/guard/panelfixsingledispatch.go`) also accepts a chunk row's
  round key, `panel-fix-<r>` (the `-<n>` chunk suffix stripped); its comment and the
  `scripts/check-panel-fix-single-dispatch.sh` header say so.
- One test, `TestCheckPanelFixSingleDispatchAcceptsAChunkRowKeyedToItsRound`, fails without the fix.
