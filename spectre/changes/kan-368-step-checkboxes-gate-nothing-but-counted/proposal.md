# kan-368-step-checkboxes-gate-nothing-but-counted

## Why

`skills/flow/implement.md` says a step's checkbox "tracks the step and gates nothing" — only a
task's own checkbox is ticked, after that task passes review. But
`scripts/check-unfinished-work.sh`'s `count_unticked()` matches `- [ ]` at any indent, so it counts
both column-0 task lines and the two-column-indented `- [ ] **Step N: …**` lines beneath them.
Since no implementer dispatch ever ticks a step box, this fires `OUTSTANDING` at `/flow`'s
integrate gate on every plan that has steps — which is every plan `writing-plans` produces. KAN-173
hit this: all 3 tasks ticked, all 13 steps (correctly) left unticked, gate reported 13 outstanding
items on a change that was actually done.

## What changes

`count_unticked()` in `scripts/check-unfinished-work.sh` matches only column-0 task lines
(`^- \[ \]`), matching the task-line grammar `skills/flow-contracts/build-green.md` already
defines (a task begins at column 0; a step is indented two columns beneath it and is never a task
of its own). The script's header comment documenting the anchoring choice is updated to say
column-0-only rather than "leading whitespace allowed, for a nested item." A regression test
asserts a plan with every task ticked but a step left unticked reports `CLEAR`.
