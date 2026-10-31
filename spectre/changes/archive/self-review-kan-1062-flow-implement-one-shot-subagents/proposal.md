# self-review-kan-1062-flow-implement-one-shot-subagents

## Why

The self-review pass of `/flow` change `kan-1062-flow-implement-one-shot-subagents` found four
non-`big` pipeline defects:

1. The REPRODUCER paragraph of `skills/flow/review-panel.md` names a slot's reproducer `R-ID-N.sh`,
   and slots keep the `R-` as a literal prefix: `R-0-mutation-N.sh` in kan-1062, kan-1048 and
   kan-1047, each renamed by hand afterwards.
2. A whole-package `go test ./cmd/flow` under `run-reproducer.sh` always exits 3: `flow stage wrap`
   force-kills only its direct child, so the TERM-ignoring `sleep 30` of
   `TestRunStageWrapContextCancelForceKillsChildIgnoringTerm` survives it, and every reproducer
   following the contract's whole-package rule is refused as unverifiable.
3. The self-review bundle leaves out the gated per-task reviewers' findings: `RenderPanel` drops
   `GatedReviewerSlot` rows and no other bundle source carries them, so the pass cannot trace
   them (kan-1062's six gated findings, Task 6's `fix` verdict among them).
4. A live check on `live-store-copy.sh`'s copy cannot record against a scratch change until a
   `.superpowers/sdd/<change>/` directory and a STARTED state record exist, and the script does not
   say so (kan-862, kan-1062).

## What changes

- The REPRODUCER paragraph names `<round>-<id>-<n>.sh` with the worked example `0-mutation-1.sh`,
  and `render-slot-prompt.sh` states the round beside `<round>`.
- `flow stage wrap` runs its child in a process group of its own and, once cancelled, kills the
  whole group; a whole-package `go test ./cmd/flow` leaves no survivor.
- The self-review bundle carries a `gated per-task reviewer findings` source rendered from the
  store's gated rows.
- `live-store-copy.sh up <name>` registers a scratch change named `<name>` in the copy — the
  `.superpowers/sdd/<name>/` directory and a STARTED state record — and `down` removes the
  directory it created.
