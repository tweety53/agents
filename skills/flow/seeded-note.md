# The seeded-note path

Loaded by `skills/flow/brainstorm-planner.md`'s **The checklist** when the ask arrives as a seeded
research note.

**The seeded-note path is legitimate, never a bypass to prevent.** When the ask arrives already
converged — a seeded research note whose own text carries the design, its decisions with their
`**ID:**` lines, and the acceptance criteria (a fully-worked issue description, or the equivalent
a `/flow-plan` capture or a handoff package brought) — the checklist questions the note already
answers are answered by the note and never re-asked: the note is the research the checklist would
gather. A question the note leaves open is still asked, batched per **The checklist**, and the merged
convergence-and-approval confirm of **Convergence** runs as written (`skills/flow/brainstorm-planner.md`) — approval is never seeded.

**The note is the research, never the plan's form.** The seeded plan is still written through C
and D like any other, and what the note abbreviates the plan spells out: every `**Files:**` field
carries full repo-relative paths — the note's shorthand (commonMain/… and its like) is expanded,
never copied, because `check-task-commit-fields.sh` matches a declared path against the commit's
diff literally — and `tasks.md`'s H1 stays the exact `# <change-id>` literal `spectre validate`
requires, never a title the note supplies.

**The note is an immutable input, never a living copy of the plan.** The plan is canonical from
the moment it is written, and mid-run corrections — file corrections, baseline re-measures — land
in `tasks.md` alone; the note is never updated to match. A note carrying a plan copy of its own —
the retired `<project>/docs/research/<stem>/` capture layout wrote one — holds a snapshot that
goes stale by design: no stage propagates corrections to it, and no reader should expect it to
track the plan.

**The note's verification tags are evidence-checked at seeding.** A `verified:`/`measured:` tag
the note carries is copied only when its evidence is in hand; a tag naming nothing is rewritten to
the honest `unverified:`/`predicted:` tag, or dropped, never copied — per **Plan provenance**'s
evidence rule (`skills/flow-contracts/plan-provenance.md`); seeding is how
an unverifiable verification tag enters a plan believing itself checked. The task-close guard
refuses a close over the shape.
