# Rendering the session records — rationale

This file is the reasoning behind `skills/flow-contracts/session-records.md`.
**A `/myflow-*` run never loads it — appendices are for whoever edits a contract.**

## Preserving the session records

*The heading keeps its former name deliberately.* **Rendering the session records** (`skills/flow-contracts/session-records.md`)
cites this section inline by it, and `check-references.sh`
matches that citing bold token against this heading — renaming it here breaks the citation there.

**A non-zero exit is never silent and never a stop.** Those two rules pull in opposite directions
and both hold: a step able to abandon an integration whose work is already committed would be a
worse failure than the missing record, *and* a refused write that passed unmentioned would leave the
operator believing a record exists when none does. So it is reported and the run continues — and the
handoff says which records reached the repository and which did not. The remaining kinds are still
attempted after any one failure, so a single bad path costs one record, not every record.

**The ordering asymmetry this section used to record is gone, and is not to be re-derived.** It said
that `/myflow-finish` run 1 called the copy script *before* staging while `/myflow-do` called it
*after*, and that the asymmetry was what kept the copied records out of `/myflow-do`'s staged-only
path. Neither call exists: the records live in the store, the script is retired, and what keeps
`<project>/docs/superpowers/` out of a staged-only run is now a **condition rather than an
ordering** — `/myflow-do` renders on its `prUrl` commit path alone, and run 1 renders before staging
so the record lands in the same commit as the implementation it describes. Reading a rule back out
of where the two call sites now sit would be reading an accident as a design.
