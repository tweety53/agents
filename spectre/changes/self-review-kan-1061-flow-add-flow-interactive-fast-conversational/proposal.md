# self-review-kan-1061-flow-add-flow-interactive-fast-conversational

## Why

kan-1061's self-review pass raised four fixable pipeline findings, none `big`, and two
prompt-tuning cases:

- **Panel rows record the wrong slot.** `flow record dispatch begin` checks only that a panel
  reviewer's `-slot` is non-empty, so rounds 3–5 recorded `test-sanity` or `mutation` under
  `panel-<r>-test-sanity+call-efficiency` and `panel-<r>-mutation+failure-modes` keys, and the
  store credits each bundle to its first role only.
- **Capture shape is a verifier-only rule.** Full-page or viewport captures, each showing a state no
  sibling capture shows, are stated only in the verifier's steps 8 and 9; the planner's
  capture-task paragraph says nothing of either, so kan-1061 planned element screenshots and then
  two byte-identical captures.
- **The parent rebuilds launch lines by hand.** The dispatch hooks refuse a launch whose
  `description` does not open with its row or whose prompt lacks its `Role:` line; `begin` knows
  both and prints neither.
- **Hook denials leave no record.** Neither dispatch hook records a denial, so a pass cannot count
  or attribute their cost.
- **Two prompt-tuning cases** (angle 8) are owed as committed case folders.

## What changes

- `begin` refuses a panel reviewer row whose `-slot` is not its key's slot part, and a new
  `flowd reslot-dispatch` repairs kan-1061's five mis-stamped rows.
- `brainstorm-planner.md`'s capture-task paragraph carries the capture-shape rule.
- A live `begin` prints the row's `description:` opening and, where the row determines one, its
  `Role:` line; `implement.md` tells the parent to paste them.
- Both dispatch hooks write one `flow record incident` row per denial, failing open.
- `docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-{1,2}/` hold the two
  rendered cases.
