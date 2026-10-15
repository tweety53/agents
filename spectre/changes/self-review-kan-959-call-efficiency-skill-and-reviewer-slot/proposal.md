# self-review-kan-959-call-efficiency-skill-and-reviewer-slot

## Why

The self-review pass over kan-959 found two defects, neither `big` — both `flow-fix`:

- `render-slot-prompt.sh` decides a slot's validity from a hardcoded Go map, `rspSlots`
  (`stats/internal/guard/renderslotprompt.go`), so a guard built before kan-959 — main's installed
  build — refuses the new `test-sanity` and `call-efficiency` slots although their templates are on
  disk. Every new slot needs a guard rebuild for no reason the template does not already carry.
- `check-verbatim-moves.sh`'s header says to copy each sentence into `verbatim-moves.txt` exactly as
  the FAIL line prints it after `::`, but a heading prints as `## …` and a `#`-leading line in that
  file is a comment, so the copied acknowledgement is silently ignored.

## What changes

- `render-slot-prompt.sh` accepts any slot id without a `/` whose `<skill-dir>/<id>-reviewer-prompt.md`
  exists; `mutation` and `exp-<name>` keep their own handling, and an id with no template is still
  refused.
- Both of `check-verbatim-moves.sh`'s FAIL lines print a sentence opening with `#` or `\` behind one
  `\`, so the text after `::` is the acknowledgement line as the file reads it; the header says so.
