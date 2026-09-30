# kan-862-flow-fix-review-minors-inline-during-the-run

## Why

A review that raises nothing above Minor defers its Minors to `KNOWN-BUGS.md`, per-task gated
review and panel round alike. They pile up — 96 entries at `0b7e1c58`, 30 of them from kan-860
alone — and are then fixed by whole separate `/flow` runs: KAN-861, 17 tasks at opus/high over 67
of them. Fixing a Minor later costs a planning pass, implementers and a panel; fixing it where it
is raised costs a few edits.

## What changes

- A Minor is fixed at the moment its review raises it, by the parent, inline: one commit at the
  branch tip, no dispatch, reproducer, mutation proof or re-review. Per-task gated reviews and
  Minor-only panel rounds both take this path; a Minor beside a Critical or Important still rides
  that round's fix.
- Nothing is recorded `deferred` again; `withdrawn` is a Minor's only exit.
  `check-panel-findings-closed.sh` reads `deferred` as open and stops requiring a fixed Minor's
  slot to re-run.
- The deferral surface goes: the known-bugs contract's `## Deferred review findings`, the review
  panel's KNOWN-BUGS close step, the handoff's `Deferred:` line and `### Deferred minors` list
  (`flow record handoff-lines` included), and the dashboard's `Deferred` column and `Deferred
  minor` panel.
- Every Minor entry in `KNOWN-BUGS.md` is deleted, unfixed; the sweep's pre-existing-failure entry
  stays.
- The implementer gets a DRIFT CHECK paragraph aimed at the most common Minors — stale comments,
  headers, usage lines and citations, and untested exit paths — before they reach review.
