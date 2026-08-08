# Design — slim the myflow contract files

The approved design is `docs/superpowers/specs/2026-08-08-kan-95-slim-the-myflow-contract-files-design.md`.
This file carries the same design in OpenSpec form, plus the decision record.

## Context

A `/myflow-do` run loads 210,481 bytes of contract and skill prose before dispatching its first
implementer, and that text is resent on every tool call for the whole run.

<!-- verified: wc -c on skills/myflow-contracts/{pipeline,project-configuration,workspace-isolation,state-file,jira-integration}.md and skills/myflow-do/SKILL.md @ f763481 -->

KAN-95's issue text quotes ~244 KB. That is the **budget** column of
`scripts/check-contract-budget.sh` — size when the row landed, plus 25% — not measured size.

`pipeline.md`'s 64,701 bytes break down as: `Pipeline flow` (the state diagram and both stage-table
levels) 13,769; `Handoff output` and the three block templates 17,692; `Model policy` 8,548;
`Temporary artifacts registry` 5,571; everything else 19,121.

<!-- verified: awk section-byte tally over skills/myflow-contracts/pipeline.md @ f763481 -->

## Approach

Five parts, in dependency order.

1. **Slim the four loaded contracts.** Evict inline rationale from `pipeline.md`,
   `skills/myflow-do/SKILL.md`, `project-configuration.md` and `workspace-isolation.md`, creating
   rationale appendices for the latter two. Move the state block templates to `handoff-blocks.md`.
   Re-anchor the budget guard.
2. **Remove `/myflow-info`**, rewriting the diagram and stage tables into `README.md`.
3. **Slim `/myflow-status`** — six columns, state file plus local git, no network.
4. **Delete self-heal**, which part 3 leaves with no performer.
5. **Compress the handoff blocks** corpus-wide, both renderings in the same change.

Part 2 is what makes part 1's stage-table removal correct: with the command gone, the tables have no
runtime consumer, so they leave the corpus rather than moving to another loaded file. Part 5 edits a
file part 1 creates.

## Decisions

### The stage tables move to `README.md` rather than to a loaded contract file

**ID:** stage-tables-to-readme
**Status:** superseded by stage-tables-leave-the-corpus
**Chosen:** a new `pipeline-stages.md` loaded by `/myflow-info` alone — it removes 13,769 bytes from
every other command while honouring KAN-87's decision that `/myflow-info` keeps the stage table
**Considered:** deleting the tables outright, as KAN-95's lever 3 words it — rejected because
`/myflow-info`'s whole job is reading `pipeline.md` and explaining the pipeline, and KAN-87 decided
one change ago that it keeps them; leaving them in the core — rejected because 13,769 bytes then
stay on every command's load for one command's benefit

### The stage tables leave the loaded corpus entirely

**ID:** stage-tables-leave-the-corpus
**Status:** active
**Chosen:** `README.md`, rewritten for a human reader, with no `pipeline-stages.md` created — the
operator removed `/myflow-info` mid-planning, which removed the only runtime consumer the tables
had, so nothing needs to load them at all
**Considered:** keeping `pipeline-stages.md` as a loaded file with no consumer — rejected as a file
the budget guard would police for nobody; moving the tables verbatim rather than rewriting them —
rejected because the audience changed from a command to a human reader

### Lever 2 creates one file, not the three the ticket names

**ID:** lever-two-creates-one-file
**Status:** active
**Chosen:** `handoff-blocks.md` alone — its full set is needed only by `/myflow-status`, while every
producing command already carries the single block it prints
**Considered:** `model-policy.md` — rejected because its consumers are `/myflow-start` and
`/myflow-do`, the two most expensive runs, so splitting it relieves only the cheap commands while
adding a second read to the two that matter; `artifacts-registry.md` — rejected because KAN-87
deliberately hoisted `Temporary artifacts registry` into the core, `project-configuration.md` and
`workspace-isolation.md` citing it while `/myflow-do` loads both, so moving it out re-creates
KAN-82's panel finding F1 by design

### A mixed passage is handled by rule extraction, under a bounded carve-out

**ID:** bounded-carve-out-for-mixed-passages
**Status:** active
**Chosen:** the original passage moves to the appendix byte-for-byte, the core gains a new sentence
stating the rule, and the ledger quotes that sentence — the existing SHALL stays true as written,
because no original passage is ever reworded, and what is added is additive and individually
auditable
**Considered:** staying strictly verbatim and moving only wholly-rationale passages — rejected
because most prose in these four files is mixed, which is exactly why it reads as argument; widening
the requirement to permit rewriting, with the ledger replacing byte-identity corpus-wide — rejected
because it trades a mechanical property for a reviewed one across every future split, including
splits that need no such licence

### Verification is a per-move ledger

**ID:** per-move-ledger
**Status:** active
**Chosen:** each eviction task emits removed-passage → destination-or-pointer, checked by the panel
against the diff in both directions — it makes the loss question answerable from an artifact rather
than from a reviewer's attention, and it survives the rewording the carve-out requires
**Considered:** a mechanical no-loss script asserting every removed line survives somewhere —
rejected because it is exact for pure moves and breaks on rule extraction, and its similarity
threshold cannot be tuned honestly; reading before-and-after with no artifact, which is KAN-87's
method — rejected because at four files and ~180 KB it leaves no record of which passages were
actually checked

### KAN-84 is not fixed here

**ID:** kan-84-stays-open
**Status:** active
**Chosen:** verify by reading plus the ledger, and state the gap in the proposal — it keeps this
change a move rather than a guard redesign, and the ledger covers the same failure from a different
direction
**Considered:** teaching `check-references.sh` to detect a hollowed section first — rejected because
the guard would have to be designed before the split could start; splitting first and filing a
follow-up to re-verify — rejected because the ledger already produces the record such a follow-up
would need

### No numeric target is set

**ID:** no-numeric-target
**Status:** active
**Chosen:** move what moves and re-anchor each budget to what lands, plus 25% — the budget guard's
own header warns that a target-first budget creates pressure to push normative text into an appendix
to make a number
**Considered:** the ticket's ~25,000 B target for `pipeline.md` — rejected as derived from the
80,876 budget rather than the 64,701 actual, and as the pressure the guard warns against; stating
the goal as a per-run load ceiling — rejected for the same reason one level up

### `/myflow-status` keeps local git and loses the network

**ID:** status-keeps-local-git
**Status:** active
**Chosen:** state file plus local git, dropping `gh pr list` alone — merge status is what splits the
`IN_PROGRESS` next-command answer between integrating and archiving, and it is local
**Considered:** rendering the table from the state file alone and moving every probe to the detail
view — rejected because the Next column would stop distinguishing the two `/myflow-finish` runs;
state file only everywhere, including the detail view — rejected as removing more than was asked

### Self-heal is deleted rather than re-homed

**ID:** self-heal-deleted
**Status:** active
**Chosen:** delete `state-self-heal.md` and the `myflow-state-integrity` capability — `/myflow-status`
was its only performer, and a contract whose mechanism nothing performs is not kept against a future
that may never arrive
**Considered:** moving it to `/myflow-finish`, where a contradicted state file does irreversible
damage through a wrong preflight verdict — rejected by the operator after the cost was stated;
moving it to `/myflow-do` — rejected as putting 15,250 bytes on the command this change exists to
slim

### All five parts ship as one change

**ID:** five-parts-one-change
**Status:** active
**Chosen:** one proposal, one worktree, one review panel, one merge
**Considered:** splitting parts 3-5 into their own ticket, leaving KAN-95 as filed plus the
`/myflow-info` removal that enables its lever 3 — rejected by the operator after the concern was
stated that a single panel would read a diff spanning contracts, commands, capabilities and output
formats at once; a three-way split — rejected for the same reason

## Open questions

*(none — every question raised during planning was answered)*

## Risks

**The guard cannot see the failure part 1 can cause.** `scripts/check-references.sh` resolves a
citation whenever a heading of that name exists, so a stub carrying a heading and a pointer resolves
identically to a stub carrying a heading and nothing. That is KAN-84, and it stays open. The ledger
and the mandatory pointer are the mitigation.

**A stale state file is now never corrected, and never flagged.** Part 4, accepted knowingly.

**`/myflow-status` can no longer say whether a pull request is open.** Part 3, accepted knowingly.

**Scope.** Five parts in one change, the largest this repository has run.
