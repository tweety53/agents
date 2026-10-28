## Context

`verifygreen.go` groups this session's verifier rows by key (fix suffix stripped) and judges the latest row of each group: `completed` passes, empty is `running`, anything else fails. `run-verify-list.sh` records `skipped` only for a whole-list cache skip whose last green run on a clean tree covered HEAD.

## Decisions

- **Accept `skipped` in the guard, not change `run-verify-list`'s outcome.** Recording `completed` for a cache skip would erase the distinction the ledger keeps between a run and a skip; the guard is the one caller that mis-reads it. Simplest change: one `case "completed", "skipped":`.
- **`skills/flow/visual-verify.md` is not edited.** It defers to the guard's header as canonical for the verdict; restating it there would duplicate the header.
- **Success line wording:** `VERIFY-GREEN: <change> — <n> inline verify row(s) green` replaces `completed`, since the count now includes skipped rows. The header's verdict-line list changes to match, and the existing tests' needles that match `row(s) completed` change with it.
- **No saved memory** encodes this defect; none to update.

## Open questions

None.
