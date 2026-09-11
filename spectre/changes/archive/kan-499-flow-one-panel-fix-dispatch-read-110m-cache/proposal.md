# kan-499-flow-one-panel-fix-dispatch-read-110m-cache

## Why

On KAN-459, panel-fix round 1 fixed 24 findings across two repos in **one** dispatch and read
~110M cache tokens — about a quarter of that run's ~435M total — because the fix subagent's
context ballooned re-reading the branch per finding. Rounds 3–5, which carried far fewer findings
each, held near ~30M cache reads. The fix-dispatch contract (`skills/flow/review-panel.md`) hands
**every** surviving open finding to one dispatch with no bound, so a finding-heavy round always
reproduces the blowup (KAN-499).

## What changes

Panel-fix dispatches are **chunked at most 10 findings per dispatch** (KAN-499's cap option,
chosen over splitting by repo — the cap bounds every round, single- or multi-repo, and matches the
measured ~30M-at-low-counts numbers):

- `skills/flow/review-panel.md`'s fix step: the one-dispatch-per-round contract becomes one
  dispatch **per chunk of at most 10 findings**, chunks dispatched sequentially, each awaited in
  the foreground; first chunk keeps key `panel-fix-<round>`, later chunks append `-<n>`
  (`panel-fix-<round>-2`, `-3`, …), one report file per dispatch; the close-guard handback wording
  follows the chunked reality.
- `scripts/check-panel-fix-single-dispatch.sh`: accepts the chunked key shape
  (`panel-fix-<round>[-<chunk>][-retry]`), requires numbered chunks contiguous from 2 with the
  bare key present when any chunk exists, keeps one-original-plus-at-most-one-retry per chunk key,
  and bounds a round's chunk count by `ceil(findings raised in earlier rounds / 10)` so the
  per-finding KAN-482 abuse stays caught (chunks are counted from the findings store — the guard
  reads no rendered document).
- `skills/flow/implement.md`'s dispatch-site closed list: the panel-fix row's key shape and
  "exactly one per round" wording updated to the chunked contract.
- `scripts/test-check-panel-fix-single-dispatch.sh`: cases for the chunked shape, contiguity,
  chunk-count bound, and per-chunk retry rules.

No stats-side change: stage keys, the store schema, and the `flow.*` vocabulary are untouched.
