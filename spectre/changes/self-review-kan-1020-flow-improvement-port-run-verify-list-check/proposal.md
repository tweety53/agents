# self-review-kan-1020-flow-improvement-port-run-verify-list-check

## Why

kan-1020's self-review pass found three defects, none `big`:

- **1 (flow-fix).** `skills/flow/implement.md`'s rule "A deviation from the plan records as a dated
  `Correction:` paragraph" never says a Correction is plan prose `check-plan-provenance.sh` reads.
  In kan-1020 the bare numbers of Task 3's Correction were refused only at verify, costing a lint
  re-run.
- **2 (flow-fix).** `skills/flow/SKILL.md` says to load `skills/flow-contracts/pipeline.md` first,
  but `pipeline.md` and `finish-contract-run1.md` are larger than the shell tool's output cap, so a
  `cat` load arrives truncated — as it did in kan-1020's run.
- **3 (flow-fix, record repair).** Task 2 of kan-1020's archived `tasks.md` changed its test course
  — the Go test moves the PNG's mtime back an hour, because the harness's re-touch to the current
  time sat inside zip's DOS-time granularity — without the `Correction (2026-10-22):` paragraph
  `implement.md` requires; Task 3's analogous change has one.

## What changes

- `implement.md` tells the parent to run `check-plan-provenance.sh` when transcribing a Correction
  and to give each numeric claim its `measured:` comment or word it without a count.
- `skills/flow/SKILL.md` tells a run to load every contract file with the file-read tool, in offset
  chunks where one read cannot return it whole, never with `cat`.
- kan-1020's archived Task 2 carries its Correction paragraph, worded without counts.
