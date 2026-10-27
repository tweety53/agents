# kan-1039-end-to-end-autonomy-rule

## Why

Runs still stop on questions the operator always answers with the recommended option — kan-1037
cost about 380k uncached tokens of such picks, and on kan-1038 the operator answered "recommended —
don't ask such things again". The operator's priority (KAN-1039): apps and tooling that work end to
end, however many fix rounds, tooling updates or tooling fixes it takes. The only interactive parts
they want are planning, and manual testing of the implementation with the fixes it finds — the
latter only when they explicitly asked to stop after implementation. No global rule says this, and
the `/flow` contracts still stop past planning in several places.

## What changes

- A new always-on global rule, `rules/work-end-to-end.mdc`, states the priority for every session
  and, through `rules/agent-baseline.md`, every subagent.
- `/flow` auto-resolution covers everything past planning — integrate, cleanup, the self-review
  pass and the run-opening prompts, not only implementation and fix runs; **What still stops**
  gains the operator's manual-testing stop and loses the integrate/cleanup carve-out.
- A guard's exit 2, a failing tool or a fixable environment fault is cured in the run, never a
  stop: one precedence paragraph in **Fewest operator actions** governs every phase file.
- Prompts whose recommended course was a stop now recommend going on: the review panel's moved-base,
  bundle-failure and wall-clock prompts; a self-review product defect continues straight into its
  fix run; `/flow-fast` without a `## handoff` key lands; the self-review asks no rating-only
  question and no `xhigh` approval.
- Stale "three courses" pointers that steered the unfinished-work gate into a prompt it never
  offers are corrected, the runtime message included.
- `flow-xhigh` and `flow-planner-xhigh` agent definitions exist for later use; the agent-baseline
  hook refuses any dispatch on them until a change wires them in.
