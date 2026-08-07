# Cut per-command load further — design

**Change:** `kan-87-cut-per-command-load-further`
**Jira:** KAN-87
**Date:** 2026-08-07
**Follows:** KAN-82, which cut a `/myflow-start` run from 156,109 to 103,844 bytes

## Problem

KAN-82 split the two largest contract files into a normative core and a rationale appendix. Two
levers remain, both **pure file moves** using machinery that change built and its review panel
already tested across two full passes.

### `pipeline.md` still cannot be read in one page

Measured on this tree: 1,348 lines, 88,253 bytes, **30,351 tokens** — read in two pages because it
exceeds a 25,000-token cap. Every `/myflow-*` command loads it.

<!-- measured: Read tool on skills/myflow-contracts/pipeline.md @ 63fa6ee -->

By top-level section, with fenced blocks excluded from heading detection:

| Section | bytes | needed by |
|---------|-------|-----------|
| `Finish contract` | 27,712 | `/myflow-finish` |
| `Handoff output` | 17,580 | all commands |
| `Pipeline flow` | 13,476 | all commands |
| `Model policy` | 8,477 | all commands |
| `Temporary artifacts registry` | 5,302 | see below — **not** finish-only |

<!-- measured: python3 fence-aware section classifier over pipeline.md @ 63fa6ee -->

### The skill files were never measured at all

KAN-82 optimised the contracts and never touched the files that load them:

| File | bytes | loaded on |
|------|-------|-----------|
| `skills/myflow-do/SKILL.md` | 40,529 | every `/myflow-do` run |
| `skills/myflow-start/SKILL.md` | 28,944 | every `/myflow-start` run |
| `skills/myflow-finish/SKILL.md` | 28,086 | every `/myflow-finish` run |
| `skills/myflow-status/SKILL.md` | 17,482 | every `/myflow-status` run |
| `skills/myflow-info/SKILL.md` | 4,229 | every `/myflow-info` run |

<!-- measured: wc -c skills/*/SKILL.md @ 63fa6ee -->

`skills/myflow-do/SKILL.md` is larger than every contract file except `pipeline.md` and
`project-configuration.md`, and carries the same rule-plus-justification prose.

## The finish block is not purely finish-only, and that decides the design

A naive whole-block move would recreate KAN-82's panel finding F1 — a file a command loads citing a
file it does not. Measured by searching for citations of each section:

| Section | cited by | real dependency? |
|---------|----------|------------------|
| `Finish contract` | `myflow-do/SKILL.md` | **yes** — the `preserve-session-records.sh` outcome table is canonical for the `prUrl` commit path |
| `Finish contract` | `myflow-start/SKILL.md`, `myflow-status/SKILL.md` | no — rationale citations |
| `Temporary artifacts registry` | `project-configuration.md`, `workspace-isolation.md` | **yes** — `/myflow-do` loads both |
| `Worktree cleanup` | `myflow-finish/SKILL.md`, `project-configuration.md`, `state-self-heal.md` | finish-side only |

**Two passages are therefore hoisted into the core before anything moves:**

| Hoist | bytes |
|-------|-------|
| The `preserve-session-records.sh` outcome table and its two explanatory notes, promoted to its own core section | 1,951 |
| `Temporary artifacts registry`, left in the core entirely | 5,301 |

<!-- measured: python3 range extraction over pipeline.md @ 63fa6ee -->

## Architecture

```text verified:authored in-tree for this change
skills/myflow-contracts/
  pipeline.md              core, 62,493 bytes   every command
  pipeline-rationale.md    appendix             editors only
  finish-contract.md       whole, 25,760 bytes  /myflow-finish only

skills/myflow-do/     SKILL.md + SKILL-rationale.md
skills/myflow-start/  SKILL.md + SKILL-rationale.md
skills/myflow-finish/ SKILL.md + SKILL-rationale.md
```

`pipeline.md` lands at **62,493 bytes**, not the 55k the linked issue estimated. The difference is
exactly the two hoists, and the honest number is stated rather than the flattering one.

<!-- measured: 88253 minus (33012 block − 1951 − 5301) @ 63fa6ee -->

`finish-contract.md` carries its rules **and** its reasoning inline, because only one command loads
it — the rule already stated as **A section reachable from only one command lives in its own file**
in `openspec/specs/myflow-contract-economy/spec.md`.

`myflow-status` and `myflow-info` are not split: one is read-only, the other is 4,229 bytes.

## Rules inherited unchanged

Paragraph granularity, verbatim moves, the two permitted edits (repointing a citation; deleting a
position word the move made false), mirrored heading trees, and the line-multiset verification. All
are already requirements of `myflow-contract-economy`. This change **widens their scope** rather than
restating them: "a contract file under `skills/myflow-contracts/`" becomes any file the pipeline
loads per run.

## The guard

`scripts/check-contract-budget.sh` widens from `skills/myflow-contracts/*.md` to also cover
`skills/*/SKILL.md` and the `SKILL-rationale.md` siblings. Its harness gains cases for the new glob.

**Budgets are regenerated once, as the final action of the change.** KAN-82's finding F23 was that
the table drifted three times because it was recomputed while later fixes were still landing.

## Verification

| What must hold | How it is proved |
|----------------|------------------|
| No line lost | Line-multiset diff of each original against its resulting files, every removed line accounted for as a permitted edit |
| No citation left pointing at absent substance | **By reading, not by guard** — `check-references.sh` resolves a citation whenever a heading of that name exists, so it is green either way |
| The four `Finish contract` citations still reach what they promise | Each read individually; `myflow-do`'s is the one the hoist exists for |
| `/myflow-info` still explains the pipeline | The state diagram, the level-1 stage table and all level-2 expansions remain in the core |
| Heading trees mirror | Fence-aware heading diff per core/appendix pair |
| The cores stay small | `scripts/check-contract-budget.sh` at its widened scope |
| Nothing else regressed | Every command in `.myflow/project.md`'s `## lint` and `## test` |

**The second row is the one that matters.** KAN-82's panel proved that `check-references.sh` cannot
detect a citation whose section was hollowed out, because the mirroring rule guarantees a heading
exists in both files. Three of that change's findings lived in that gap. The same gap is open here
and is closed by reading, which is why the plan names the four citations individually.

## Decisions

### The finish block's shared passages are hoisted into the core before the move

**ID:** hoist-before-cut
**Status:** active
**Chosen:** Promote the `preserve-session-records.sh` outcome table to its own core section and
leave `Temporary artifacts registry` in the core, then move what genuinely remains finish-only.
**Considered:** Moving the whole block and repointing the citations — rejected because
`workspace-isolation.md` and `project-configuration.md` are loaded by `/myflow-do`, so a citation
from them into `finish-contract.md` names a file that command never loads. That is KAN-82's finding
F1 reintroduced deliberately rather than by accident. Moving only `Finish contract` and leaving the
registry — rejected as leaving the `myflow-do` outcome-table dependency unresolved, which is the
harder half of the problem.

**The cost is 7,252 bytes of the 33,012**, so the core lands at 62,493 rather than nearer 55,000.
Paid deliberately: a smaller core with a dangling reference is worse than a larger one without.

### The rationale for a skill lives beside its `SKILL.md`

**ID:** skill-rationale-sibling
**Status:** active
**Chosen:** `skills/<skill>/SKILL-rationale.md`, mirroring the contracts pattern.
**Considered:** Putting skill rationale into `skills/myflow-contracts/` — rejected because a skill's
reasoning is not a contract, and it would make that directory a catch-all the budget guard then has
to police. A single shared rationale file for all skills — rejected for the reason KAN-82 rejected
one shared contract appendix: an editor of one skill would load the reasoning of all of them.

The harness loads `SKILL.md` and nothing else, and `skills/myflow-do/` already holds five `.md`
files, so an additional sibling changes nothing about skill resolution.

### Only the three pipeline skills are split

**ID:** three-pipeline-skills-only
**Status:** active
**Chosen:** `myflow-do`, `myflow-start`, `myflow-finish` — 97,559 bytes, the files loaded on every
pipeline run.
**Considered:** Adding `myflow-status` (17,482) — rejected because it is read-only and cheap to run,
so the saving accrues to the command that matters least. All five — rejected because
`myflow-info` is 4,229 bytes and its appendix would hold almost nothing.

### The economy capability widens rather than gaining a twin

**ID:** widen-economy-capability
**Status:** active
**Chosen:** MODIFIED deltas generalising `myflow-contract-economy`'s requirements from "a contract
file under `skills/myflow-contracts/`" to any file the pipeline loads per run.
**Considered:** ADDED requirements covering skill files separately — rejected as two near-identical
rule sets to keep in step, which is the duplication this corpus repeatedly punishes. A new
capability for skill-file economy — rejected because today the rules are identical, so it would be
two capabilities stating one thing.

### `/myflow-info` keeps the stage table and loses the contract detail

**ID:** myflow-info-core-only
**Status:** active
**Chosen:** The state diagram, the level-1 stage table and every level-2 expansion stay in the core,
so `/myflow-info` still describes both finish runs and every stage. It gains one line naming where
the finish contract's detail now lives.
**Considered:** Having `/myflow-info` load `finish-contract.md` too — rejected because it would make
the one read-only command pay the full load, undercutting the split for the command that benefits
least from it.

## Open questions

*(none — every question raised across three rounds was answered before the design gate)*
