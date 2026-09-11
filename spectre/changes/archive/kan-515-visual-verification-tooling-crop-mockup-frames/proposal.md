# kan-515-visual-verification-tooling-crop-mockup-frames

KAN-515 · the agents-repo half of KAN-514.

## Why

`scripts/compose-mockup-frames.py` pastes a captured screenshot beside a mockup frame at native
size, with no cropping. The group-sessions mockups are 2x exports of a 390px HTML phone frame:
a 26px status line the app never renders, a 1px border inside the 390, a content-hugging height,
and a dark caption band carrying the design doc's annotation text. So every composite so far
compared a 1200x1164 desktop capture against a 780x2028 phone frame with its chrome — the
"never apples-to-apples" root cause KAN-514 records.

Two panels read by eye also missed padding, shadow and colour departures across four fix rounds
of kan-30: a verifier comparing two images has no mechanical signal for where they differ.

## What changes

- `compose-mockup-frames.py` learns an optional geometry argument, crops each frame to its
  content area (caption band bottom detected per frame), writes a three-panel composite
  (capture | cropped frame | binary difference) and prints the differing-pixel ratio per pair.
- A capture whose size differs from the cropped frame becomes an exit-1 finding, so a wrong
  viewport blocks instead of composing as a layout departure.
- `.flow/project.md`'s `## visual verification` gains one optional `mockup frame` row;
  `check-visual-verification.sh`'s closed vocabulary gains the name.
- `flow.visual-verify` step 9 passes the resolved row and reads the third panel.
- Both guard harnesses gain cases.

No row declared leaves today's behaviour byte-for-byte: no crop, two panels, no size check.

## Added scope — the four `main` blockers

`flow.verify` found four failing commands, every one reproducing byte-identically at the merge
base `b9e4020` and none of them this change's: four missing `skills/flow-fast/scripts/` symlinks,
`skills/flow-fast/SKILL.md` over its budget, `TestStageKeysMatchFlowFastSkillTable` on that file's
absent `## Stage keys` heading, and `check-vocabulary.sh` on a `MERGE_BASE` shell local in
`check-visual-verify-dispatched.sh` colliding with the retired state field of that name. The
operator chose to clear them here rather than block the run or land them separately; task 5
carries them.
