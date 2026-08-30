# kan-368-step-checkboxes-gate-nothing-but-counted — design

## Decisions

### Anchor `count_unticked` to column-0 task lines, not the step-ticking side

**ID:** anchor-not-tick-steps
**Status:** active
**Chosen:** narrow `count_unticked`'s regex from `^[[:space:]]*- \[ \]` to `^- \[ \]` — one line,
matches the grammar `build-green.md` already states, changes no other contract.
**Considered:** having the implementer dispatch tick step boxes as it completes them instead — the
KAN-368 ticket's second candidate fix. Rejected: `implement.md` is explicit and long-standing that
"a step's checkbox tracks the step and gates nothing," and making the guard the source of truth
(narrowing what it counts) is a smaller, more localized change than making every implementer
dispatch tick a box the contract already says is inert.

### Keep the anchor at column 0, not fence-tracking

**ID:** column-0-not-fence-tracking
**Status:** active
**Chosen:** column-0 anchoring, unchanged from before this task except widening from "leading
whitespace allowed" to "column 0 only."
**Considered:** tracking fenced code blocks so a quoted `- [ ]` example inside a fence never counts
at all, regardless of its indent. Already rejected in the script's own header comment before this
task, for the same reason restated here: real complexity, and what it would still miss (a fenced
example whose line begins with `- [ ]` at column 0) already errs toward `OUTSTANDING`, which is the
safe direction. Not reopened by this change.

## Open questions

None.
