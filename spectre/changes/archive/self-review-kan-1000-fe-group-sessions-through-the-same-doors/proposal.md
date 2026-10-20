# self-review-kan-1000-fe-group-sessions-through-the-same-doors

## Why

The kan-1000 self-review pass found five pipeline defects, none `big`, and the rows and records
they already wrote wrong (`design.md`, **Context**):

- flowd's harvester pairs agent launches with dispatch begins by time alone, so a round whose
  begins went out in another order than its launches put every agent on another row, and a round
  whose begins carried an unresolved `-key panel-$R-…` left its rows with no agent — 22 kan-1000
  rows mis-measured or unmeasured.
- A dispatch row back-filled as `inline` after its window closed is never measured: the parent's
  usage inside the window was committed before the row existed, and the watcher never re-reads it
  — four kan-1000 rows.
- A panel slot's Important that the parent never recorded as a finding row passes
  `check-panel-findings-closed.sh`, which reads store rows only — round 11's Important.
- Rebuilding `full-app-suite.zip` over unchanged PNGs makes a commit with no change behind it: the
  zip stores each file's mtime.
- `flow record dispatch begin` accepts a visual-verify verifier row with no `-diff-base`, which
  only integrate's `check-visual-verify-dispatched.sh` notices.

## What changes

- The harvester stamps a launch onto the row its Agent-tool `description`'s first word names in
  the launching session, before any time pairing; the 22 kan-1000 rows are restamped.
- A back-filled `inline` row is measured from the parent transcript when the watcher sees its
  backfill, and `flowd restamp-dispatch` repairs one from a parent transcript; the four kan-1000
  rows are repaired.
- Every panel slot's report opens with `Critical: <n>` and `Important: <n>`, and
  `check-panel-findings-closed.sh` exits 1 when a report counts more than the store holds for its
  round and slot; kan-1000's round-11 Important is recorded as F10.
- The verifier pins each PNG's mtime before zipping the full-app suite.
- `flow record dispatch begin` refuses a `visual-verify*` verifier row with no `-diff-base`.
