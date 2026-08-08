# Slim the myflow contract files — design

**Change:** `kan-95-slim-the-myflow-contract-files`
**Jira:** KAN-95
**Date:** 2026-08-08

## Problem

Every `/myflow-*` command reads its contract files in full before it acts. A `/myflow-do` run loads
210,481 bytes of contract and skill prose before dispatching its first implementer, and that text is
resent on every tool call for the whole run.

<!-- verified: wc -c on each file listed below @ f763481 -->

| File | Bytes |
|------|-------|
| `skills/myflow-contracts/pipeline.md` | 64,701 |
| `skills/myflow-contracts/project-configuration.md` | 45,274 |
| `skills/myflow-do/SKILL.md` | 37,885 |
| `skills/myflow-contracts/workspace-isolation.md` | 31,079 |
| `skills/myflow-contracts/state-file.md` | 15,951 |
| `skills/myflow-contracts/jira-integration.md` | 15,591 |
| **total** | **210,481** |

KAN-95's issue text quotes 244 KB for the first five of these. That is the **budget** column of
`scripts/check-contract-budget.sh` — each budget is its file's size when its row landed plus 25% —
not the files' actual size. This design is built against measured bytes.

Two changes have already worked this problem. KAN-82 split the two largest contracts into a
normative core and a rationale appendix. KAN-87 moved the finish contract out of `pipeline.md` and
gave the three pipeline skills `SKILL-rationale.md` siblings. Both were pure file moves, and both
declared `project-configuration.md` and `workspace-isolation.md` out of scope.

## Scope

Five parts. KAN-95 as filed is part 1; parts 2 through 5 were added during this planning session and
are synced into the issue description.

| Part | What |
|------|------|
| 1 | Slim the four loaded contracts — levers 1, 2 and 4, under the bounded carve-out and the per-move ledger |
| 2 | Remove `/myflow-info`; its pipeline explanation is rewritten for humans into the root `README.md` |
| 3 | `/myflow-status`: drop the worktree column; read from the state file plus local git, never the network |
| 4 | Delete `state-self-heal.md` and the `myflow-state-integrity` capability; re-home the rules that cite them |
| 5 | Compress the handoff blocks corpus-wide; the block templates move to `handoff-blocks.md` |

The parts are sequentially coupled: part 2 is what makes part 1's lever 3 correct, and part 5 edits
a file part 1 creates. Implementing all five in one change was a deliberate choice, made against a
stated concern that a single review panel would read a diff spanning contracts, commands,
capabilities and output formats at once.

## Part 1 — slim the loaded contracts

### Lever 2 — split by consumer

One section of `pipeline.md` is read by exactly one command. It moves to its own file and leaves a
stub in the core.

| Section | Bytes | New file | Sole consumer |
|---------|-------|----------|---------------|
| `The block each state renders` and the three state block templates | ~14,000 | `skills/myflow-contracts/handoff-blocks.md` | `/myflow-status` |

<!-- verified: awk section-byte tally over skills/myflow-contracts/pipeline.md @ f763481 -->

The general rules under `Handoff output` — absolute paths, link-never-paste, the last-line
convention, the tab commands — stay in the core, because every command prints a handoff.

Two further candidates were considered and rejected:

- **`model-policy.md`** — its consumers are `/myflow-start` and `/myflow-do`, the two most expensive
  runs. Splitting it relieves only the cheap commands while adding a second read to the two that
  matter. It stays in the core and gets lever 1 instead.
- **`artifacts-registry.md`** — KAN-87 deliberately hoisted `Temporary artifacts registry` *into*
  the core because `project-configuration.md` and `workspace-isolation.md` cite it and `/myflow-do`
  loads both. Moving it out re-creates KAN-82's panel finding F1 by design.

`pipeline-stages.md` was designed and then dropped. It existed only because `/myflow-info` needed
the stage tables; part 2 removes that command, so the tables have no runtime consumer at all and go
to `README.md` instead. Nothing loads them.

### Lever 1 — evict inline rationale

Applied to `pipeline.md`, `skills/myflow-do/SKILL.md`, `project-configuration.md` and
`workspace-isolation.md`.

**What may be evicted.** A passage whose removal changes no instruction: measurement narrative,
design history, argument against a rejected alternative, and restated justification for a rule
stated elsewhere in the same file.

**What may not.** Anything carrying an obligation — `never`, `always`, `must`, `SHALL` — or an
exit-code contract, a byte value, a path, or a command.

**Mixed passages** carry both, and are handled by the bounded carve-out below.

**Every eviction leaves a pointer**: one sentence naming the rule, then the appendix and the heading
holding the reasoning. A core section whose prose is gone and whose pointer is absent is precisely
the hollowed section `scripts/check-references.sh` cannot detect.

Two new appendices are created, following KAN-82's pattern:
`skills/myflow-contracts/project-configuration-rationale.md` and
`skills/myflow-contracts/workspace-isolation-rationale.md`. Neither file has a sibling today, and a
`/myflow-*` run loads neither.

### Lever 4 — re-anchor the budget guard

`scripts/check-contract-budget.sh` gains a row for each new file, drops the rows for each deleted
file, and has the row for every file this change touches re-anchored to its landed size plus 25%,
exactly as its own header specifies.

**No numeric target is set for any file.** The guard's header warns that a target-first budget
creates pressure to push normative text into an appendix to make a number, which is the one way this
change could do harm rather than merely disappoint. The sizes below are projections recorded so the
outcome can be compared against the expectation — not commitments.

| File | Now | Projected |
|------|-----|-----------|
| `skills/myflow-contracts/pipeline.md` | 64,701 | ~23,000 |
| `skills/myflow-do/SKILL.md` | 37,885 | ~26,000 |
| `skills/myflow-contracts/project-configuration.md` | 45,274 | ~28,000 |
| `skills/myflow-contracts/workspace-isolation.md` | 31,079 | ~20,000 |
| `skills/myflow-contracts/state-file.md`, `jira-integration.md` | 31,542 | untouched |
| **`/myflow-do` per run** | **210,481** | **~128,500** |

<!-- unverified: projections from per-section byte tallies and a judgment of which passages are evictable -->

### The bounded carve-out to the verbatim-partition rule

`openspec/specs/myflow-contract-economy/spec.md` carries
**Requirement: The split is a verbatim partition, with citation repointing as its only edit**:

> Every sentence of the original file SHALL land in exactly one of the two resulting files,
> byte-for-byte unchanged. No passage SHALL be summarised, reworded, merged, or authored afresh.

Two edits are permitted, both mechanical: repointing a citation, and deleting a stale position word.
The requirement closes by stating that the constraint is met by construction rather than by review.

**Lever 1's handling of a mixed passage authors a new sentence, which that requirement forbids.**
The conflict is resolved by widening the requirement in one narrow, stated way rather than by
abandoning it:

- The **original passage moves to the appendix byte-for-byte unchanged**. Every sentence of the
  original file still lands in exactly one resulting file, unmodified — the existing SHALL stays
  true as written.
- The **core gains a new sentence** stating the rule the passage carried.
- The new sentence is **quoted in the per-move ledger** alongside the appendix passage it replaces,
  so the panel checks the two against each other rather than taking the extraction on trust.

Replacing byte-identity with the ledger corpus-wide was rejected: it trades a mechanical property
for a reviewed one across every future split, including splits that need no such licence.

### Verification — the per-move ledger

Each task that evicts or moves prose writes a table into its own output, before the review panel
sees the diff:

| Removed passage (first 8 words) | Source heading | Destination | Pointer left |
|---|---|---|---|

`Destination` names an appendix or new-file heading; reads `— (rule extracted)` with the extracted
core sentence quoted; or reads `— (rewritten for README.md)` for the part 2 passages. The panel
checks the table against the diff in both directions: every deletion in the diff has a row, and
every row's destination exists.

Two alternatives were rejected. A **mechanical no-loss script** asserting that every removed line
survives somewhere is exact for pure moves and breaks on the rewording the carve-out requires; the
similarity threshold it would need cannot be tuned honestly. **Reading before-and-after without an
artifact** is KAN-87's method, and at this scope it leaves no record of which passages were actually
checked.

## Part 2 — remove `/myflow-info`

`skills/myflow-info/`, `commands/myflow-info.md` and `commands-claude/myflow-info.md` are deleted.

The state diagram, the Level 1 stage table and every Level 2 expansion — 13,769 bytes of
`pipeline.md`'s `Pipeline flow` section — are **rewritten for humans** into a new section of the
root `README.md`.

**A rewrite for humans is not a verbatim partition, and is outside that requirement rather than an
exception to it.** The verbatim-partition rule governs a contract's core/appendix split. This is a
removal from the loaded corpus plus fresh authorship in a document no command reads. The ledger rows
for these passages say so.

`/myflow-status` becomes the pipeline's only read-only command.

## Part 3 — slim `/myflow-status`

**The table drops the `Worktree / branch` column**, leaving six: Change, Jira, State, PR, Next,
Updated. The absolute worktree path stays in the detail view, where `myflow-handoff-output`'s
absolute-path requirement still binds it.

**The command reads the state file plus local git, and never the network.** The `gh pr list` probe
is removed. The `PR` column shows the number parsed from the recorded `prUrl`, or `—` when it is
`null`; it no longer reports whether the pull request is open, merged or closed.

Merge status is retained, because it is local git and it is what splits the `IN_PROGRESS` row's
next-command answer between integrating and archiving. Its three ordered steps, the resolve-before-
compare rule, and the multi-repo combination rule are unchanged.

## Part 4 — delete self-heal

`/myflow-status` is the only command that performs self-heal. Part 3 removes it from that command,
which leaves the mechanism with no performer, so the mechanism is deleted rather than left dormant.

- `skills/myflow-contracts/state-self-heal.md` is deleted, with its budget row and harness case.
- `openspec/specs/myflow-state-integrity/` is **removed as a capability**. Both of its requirements
  are self-heal requirements; nothing in it survives the deletion. Monotonic state writes are
  unaffected — they live in `state-file.md`.
- The `⚠` marker and the one permitted `prUrl` correction disappear from `/myflow-status`.
- Every citation is stripped: `rules/myflow-manual-review.mdc`, `skills/myflow-contracts/SKILL.md`,
  `skills/README.md`, `CLAUDE.md`, `AGENTS.md`, `pipeline.md`, `pipeline-rationale.md`,
  `finish-contract.md`, `state-file.md`.

**Rules that cite the deleted file are re-homed into `state-file.md` rather than left pointing at
nothing** — the closed-schema rule, the absent-key exception for `planningEffort` and `models`, and
the retired-key paragraph all depend on the notion of an unparseable file, which the deleted file
defined.

`Resolving a change's worktrees` **moves into `finish-contract.md`**. Its stated reason for living in
`pipeline.md` was that self-heal needed the scan and its performer did not load `finish-contract.md`.
That reason dies with self-heal — and the replacement reason this design first recorded, that
`/myflow-do` needs the scan too, turned out to be false: nothing but `finish-contract.md` cites the
section, and `/myflow-do` resolves its worktree through superpowers:using-git-worktrees and reads
paths from the state file. With one consumer left, the section goes to that consumer's contract, by
this change's own **A section reachable from only one command lives in its own file**. The false
reason is recorded here rather than quietly replaced.

**A later panel round changed this again, and the record follows rather than hides it.** The *general*
rule — that a step needing "the worktrees" resolves the set first and never loops the map directly,
and that an empty resolved set is never a vacuous pass — was hoisted into `pipeline.md`, which every
command loads, because `/myflow-do`'s workspace-isolation gate and `/myflow-status`'s merge-status
report have the same defect and can load neither `finish-contract.md` nor each other's files. So
`/myflow-do` and `/myflow-status` now cite that rule. What stayed finish-only is the concrete
scan-and-resolve *procedure* — the `git worktree list --porcelain` parse — which remains in
`finish-contract.md` as the fullest example of applying the rule.

**The cost of this part is stated plainly.** A state file that goes stale, is contradicted by
artifacts, or records a `prUrl` for a pull request that no longer exists is now never corrected by
anything, and never flagged. This was raised as a concern and the decision was to delete the
mechanism regardless.

## Part 5 — compress the handoff blocks

The block templates move to `handoff-blocks.md` (part 1, lever 2) and are compressed there. The
producing commands' copies are compressed identically in the same change, so the two renderings
never diverge.

**Related fields fold onto one line. No field is dropped**, so the missing-rather-than-dropped rule
and the run-only rule both survive unchanged.

```text
## Proposal ready — review required

**Change:** kan-95-slim-the-myflow-contract-files
**Artifact:** <artifactUrl, or "missing">
**Recorded:** <N> decisions · <N> open questions · effort <level> · models <impl>/<panel>/<fix>
**Jira:** (run-only) <issue key and the transition made, or "none linked", or a skipped line>

Open in IntelliJ:
open -na "IntelliJ IDEA" --args "<absolute main-checkout path>"

Next:
/myflow-do <name>
```

The fold has a property the design targets rather than discovers: it groups the **on-disk** values
onto one `**Recorded:**` line and leaves the **run-only** fields — `Jira` and
`Jira description (pre-edit)` — as their own lines. `/myflow-status` renders the folded line in full
and omits the run-only lines, exactly as it does today. `myflow-handoff-output`'s identical-label-set
rule becomes an identical-folded-line-set rule.

**`Jira description (pre-edit)` is never folded and never dropped.** `editJiraIssue` replaces the
whole description field and there is no local backup, so the handoff transcript is the recovery
path. It prints only on a run that wrote a description, verbatim in a fenced block.

## Capabilities

**Modified**

- **`myflow-contract-economy`** — the verbatim-partition requirement gains the carve-out; a new
  requirement carries the per-move ledger. `A section reachable from only one command lives in its
  own file` and `A byte budget ratchets every file the pipeline loads per run` are applied rather
  than changed.
- **`myflow-contract-distribution`** — **Requirement: The pipeline diagram and its stage table live
  in `pipeline.md`, and nowhere else** is contradicted directly: they move to `README.md`. Its
  index-and-directory agreement rule gains `handoff-blocks.md` and two appendices and loses
  `state-self-heal.md`.
- **`myflow-handoff-output`** — **Requirement: The handoff block is defined once and rendered by two
  commands** stays true; the file holding the definition changes, and the identical-label-set rule
  becomes an identical-folded-line-set rule.
- **`myflow-command-surface`** — **Requirement: The command surface is three pipeline commands plus
  two read-only ones** becomes three plus one.
- **`myflow-progress-visibility`** — its read-only-commands-register-nothing rule names
  `/myflow-info`.
- **`myflow-state-machine`** — the transition table's `/myflow-info` row.

**Removed**

- **`myflow-state-integrity`** — both requirements are self-heal requirements, and part 4 deletes
  the mechanism.

## Repointing load

<!-- verified: grep -rl over the working tree excluding .git and archive @ f763481 -->

| Moving section | Files citing it |
|---|---|
| `The block each state renders` | 19 |
| `Pipeline flow` | 14 |
| `Level 1 — the stages of each command` | 7 |

`Model policy` (29 citations), `Handoff output` (12) and `Temporary artifacts registry` (18) stay in
the core and need no repointing, which is a large part of why they stay.

Files edited beyond the four contracts: `skills/myflow-status/SKILL.md`,
`skills/myflow-contracts/SKILL.md`, `scripts/check-contract-budget.sh` and its harness, `CLAUDE.md`,
`AGENTS.md`, `rules/myflow-manual-review.mdc`, `README.md`, `skills/README.md`, and every file citing
a section that moved. Twelve files under `docs/` mention `/myflow-info`; they are records of past
runs and are left untouched.

## Not a behaviour change, except where it is

Parts 1 and 5 change no rule. The carve-out permits a rule to be restated more briefly in the core
while its original wording survives byte-for-byte in an appendix; the fold changes a block's layout
and drops no field.

Parts 2, 3 and 4 **are** behaviour changes, and are named as such rather than folded into the
slimming claim: a command is removed, a report loses a column and a network probe, and a correctness
mechanism is deleted.

## Risks

1. **The guard cannot see the failure part 1 can cause.** `scripts/check-references.sh` resolves a
   citation whenever a heading of that name exists, so a stub carrying a heading and a pointer
   resolves identically to a stub carrying a heading and nothing. This is KAN-84, which stays open.
   The ledger and the mandatory pointer are the mitigation; the proposal states the gap rather than
   implying the guard covers it.
2. **A stale state file is now never corrected** — part 4, accepted knowingly.
3. **`/myflow-status` can no longer say whether a pull request is open** — part 3, accepted
   knowingly.
4. **Scope.** Five parts, one worktree, one review panel, one merge — the largest single change this
   repository has run.

## Out of scope

- `state-file.md` and `jira-integration.md` beyond the re-homing part 4 forces on the first.
- KAN-84 itself. This change verifies by reading plus the ledger, and says so.
- The harness system prompt, which is an operator action against an installed environment.
