# kan-378-flow-deletion-only-cut-of-stale-restated-and

Linked issue: KAN-378 — "flow: deletion-only cut of stale, restated and rationale text in run-loaded files".

## Why

kan-372 made the router's contract loads lazy but measured phase files only. Followed to the leaf —
router, phase files, every contract a phase file loads at its point of use, the superpowers skills
each phase invokes, and the always-on files — a creating run's parent still loads roughly 342k bytes
of instruction text before it reads a line of project code, and every subagent inherits roughly 32k
bytes before its role prompt. Part of that text is stale (a disclosure that `names.go` carries no
`flow.*` key, written before it did), part is restated (the state-transition table appears four
times in one creating run), part is reasoning for whoever edits a file rather than instruction for
a run (rejected-alternative rationale sitting in run-loaded files despite the `-rationale.md`
convention), and part is read twice (`brainstorm.md` in full by the parent and again by the planner;
`engineering-principles.md` in the implementer's bundle and again as required reading).

## What changes

Deletion-only, no behaviour change. Every runtime-governing statement — states, transitions, stage
keys, guard names, field names, exit codes, ordering constraints, operator-prompt wording — survives;
only stale text, restated copies and rationale move or go. Rejected-alternative reasoning is moved,
never deleted.

- `skills/flow-contracts/pipeline.md` becomes the sole canonical transition table: its `/flow` row
  expands into one row per accepted state, and `skills/flow/SKILL.md`'s **State transitions**
  section and project `CLAUDE.md`'s `/flow commands summary` are deleted. The global always-on stub
  keeps its three-line trigger.
- `skills/flow/SKILL.md`'s stale **Stage keys** disclosure and its `model-policy.md` caveat are
  deleted; the key table stays.
- Rejected-alternative reasoning in `skills/flow/SKILL.md`, `brainstorm.md` and `integrate.md`
  moves verbatim to a new `skills/flow/SKILL-rationale.md`, which no run loads.
- `skills/flow/brainstorm.md` splits: the parent keeps section A and **Dispatch the planner**; the
  planner's sections B–D move to `skills/flow/brainstorm-planner.md`, each file read once.
- `skills/flow/implement.md` drops the `superpowers:subagent-driven-development` invocation it
  already overrides, and points the implementer's required reading at the bundle's own
  engineering-principles section instead of a second file read.
- `rules/flow-manual-review.mdc` gains a `core` marker so the managed block every session and
  subagent inherits carries the trigger and the load-first instruction only.
- `skills/flow/principles-reviewer-prompt.md` loses its roster-history paragraph and its
  `project-configuration.md` cite; the dispatcher already resolves `[STANDARDS_PATHS]` before
  dispatch.
- Two rows join `scripts/check-contract-budget.sh` for the two new files; no existing row moves.
- Before/after byte and token figures per load set are recorded in this change's `design.md`,
  measured to the leaf, so KAN-379 (mechanics-to-code) starts from this change's after-figures.

`rules/agent-baseline.md`'s rule table is kept, per the issue.
