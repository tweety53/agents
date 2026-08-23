# Design — cut `pipeline.md`'s load cost

**Jira:** KAN-295

The approved design, adapted. `proposal.md` states the why and what; this file states the how.
## Measured section sizes

The five moved sections total **19,793 bytes**; the core retains **30,497**.

| Section | Bytes | Moves |
|---------|-------|-------|
| Model policy | 6661 | `model-policy.md` |
| Stage marks | 6630 | stays |
| Temporary artifacts registry | 6183 | `artifacts-registry.md` |
| Change name resolution | 4256 | stays |
| Git boundaries | 3946 | `git-boundaries.md` |
| Handoff output | 3875 | stays |
| Progress visibility | 2393 | stays |
| Guard presence check | 2161 | stays |
| State transitions | 1913 | stays |
| Artifact brevity | 1762 | stays |
| Guard resolution | 1527 | stays |
| Resolving a change's worktrees | 1517 | `worktree-resolution.md` |
| Rendering the session records | 1486 | `session-records.md` |
| States | 1450 | stays |
| Stage exit — never the command's own judgment | 817 | stays |
| IntelliJ commands | 651 | stays |
| Wrong state for this command | 615 | stays |
| Command surface | 561 | stays |
| Finish contract / State file / Project configuration / Jira integration pointers | 1165 | stay |

<!-- measured: awk section-size pass over skills/myflow-contracts/pipeline.md @ 15f11cc -->

## The six-file family

`pipeline.md` retains every section reachable from all five commands: States, Stage exit, Command
surface, State transitions, Wrong state for this command, Progress visibility, Stage marks, Handoff
output, Artifact brevity, IntelliJ commands, Guard resolution, Guard presence check, Change name
resolution, and the pointer sections it already carries for Finish contract, State file, Project
configuration and Jira integration.

Five sections move out whole:

| New file | Section moved | Loaded by |
|----------|---------------|-----------|
| `git-boundaries.md` | Git boundaries | `/myflow-do`, `/myflow-finish`, `/myflow-fast` |
| `model-policy.md` | Model policy | `/myflow-start`, `/myflow-do`, `/myflow-fast` |
| `artifacts-registry.md` | Temporary artifacts registry | `/myflow-do`, `/myflow-finish` run 2, `/myflow-fast` |
| `session-records.md` | Rendering the session records | `/myflow-do`, `/myflow-finish` run 1, `/myflow-fast` |
| `worktree-resolution.md` | Resolving a change's worktrees | `/myflow-do`, `/myflow-finish`, `/myflow-status`, `/myflow-fast` |

Each takes the shape `handoff-blocks.md` established: an `#` title, an intro naming which commands
load it, the sentence `This file is **canonical** for everything in it.`, then the original `##`
heading and its body verbatim. Canonical authority moves with the contract text, per
**Requirement: Canonical authority moves with the contract text**
(`openspec/specs/myflow-contract-distribution/spec.md`).

`pipeline.md` gains one short line naming its five siblings, so the family is discoverable from the
file every command already loads. It is a discovery line, not a stub carrying rules: it claims no
authority over text it no longer contains.

<!-- measured: pipeline.md 28924 bytes after the split and four rationale-sweep/repair passes, at
     HEAD. Pass two chased the 45000-byte target by splitting sentences at their punctuation joints
     — a paraphrase, caught by panel findings F1-F4. Pass three restored every split sentence whole
     (48291 bytes). Pass four (this one) found the restoration itself had cut five pointers outright
     and left four more paragraphs ending cold with no pointer at all (F11-F16), rebuilt the per-move
     ledger from the diff rather than memory (F17-F18), recomputed all twelve budget rows against
     current sizes (F19), and corrected three places where canonical authority for git boundaries had
     not moved with the text (F20-F22). Final family total: 49624 bytes — over the 45000-byte target,
     every command still a real reduction against the 50290-byte baseline, every claim in
     verification.md backed by a pasted command. See verification.md's "Repair round 2" and "Measured
     result" sections for the full accounting. -->

## Appendices

Each new file gets `<name>-rationale.md` beside it, per **Requirement: A contract file separates
its normative core from its rationale** (`openspec/specs/myflow-contract-economy/spec.md`), whose
scenario fixes the path as `skills/myflow-contracts/<name>-rationale.md`. Each mirrors its core's
heading tree, per **Requirement: An appendix mirrors its core's heading tree** (same spec),
following `handoff-blocks-rationale.md`: the core's `##`/`###` headings in order, with `### Why …`
subsections beneath them.

The matching sections move out of `pipeline-rationale.md`, which re-mirrors the reduced
`pipeline.md`. Without that move the two heading sequences diverge the moment `## Git boundaries`
leaves the core, which the mirroring requirement forbids.

**One pre-existing deviation is preserved, not repaired.** `pipeline.md`'s core heading is
`## Rendering the session records` while its appendix heading is `## Preserving the session
records`, and `pipeline.md` states in prose that the appendix "keeps its former name". Both travel
to `session-records.md` and `session-records-rationale.md` unchanged. Renaming either would be a
behavioural edit outside this change's scope, and the prose asserting the deviation moves with them.

## Rationale sweep

`pipeline.md` and all five extracted files are swept for inline argument prose, which moves to the
matching appendix leaving the normative sentence plus a pointer. The test is the one
`myflow-contract-economy` already states: a passage is core if removing it would change what an
agent does.

`/myflow-do` loads every file in the family, so it is the only consumer this sweep helps; the other
four commands are helped by the split alone. The sweep is nonetheless the ticket's item 1 and is
carried out across the family rather than only in the four sections the ticket named, three of which
are being extracted anyway.

## Load declarations

**No row is added to `rules/myflow-manual-review.mdc`.** Its contract table and its
`pipeline.md`-first instruction are untouched, so the always-on context cost is unchanged. The table
lists contracts a *step* reaches for; the five new files are the pipeline's own internals, split by
which command reads them.

Each command's `SKILL.md` instead names the extra contracts it loads, at the step that needs them:

| Skill | Gains a load line for |
|-------|-----------------------|
| `skills/myflow-start/SKILL.md` | `model-policy.md` |
| `skills/myflow-do/SKILL.md` | `git-boundaries.md`, `model-policy.md`, `artifacts-registry.md`, `session-records.md`, `worktree-resolution.md` |
| `skills/myflow-finish/SKILL.md` | `git-boundaries.md`, `artifacts-registry.md`, `session-records.md`, `worktree-resolution.md` |
| `skills/myflow-fast/SKILL.md` | all five |
| `skills/myflow-status/SKILL.md` | `worktree-resolution.md` |

## Citations

**57** section citations point into `pipeline.md` from `skills/`, `rules/`, `commands/`,
`commands-claude/`, `CLAUDE.md`, `AGENTS.md`, `README.md` and `openspec/specs/`. **33** of them name
one of the five moved sections and are repointed; the remaining 24 name a section that stays and are
untouched.

<!-- measured: grep -rno for the `**Section** (`skills/myflow-contracts/pipeline.md`)` shape @ 15f11cc -->

**Requirement: A passage another command depends on is hoisted before a single-command move**
(`openspec/specs/myflow-contract-economy/spec.md`) requires every citation into a moved section to
be resolved by reading, and forbids offering a green `check-references.sh` as evidence. Four
citations cross a load boundary — the citing file is loaded by a command that would not load the
destination. All four were read:

| Citation | Cites | Citing file loaded by | Destination loaded by | Verdict |
|----------|-------|-----------------------|-----------------------|---------|
| `skills/myflow-start/SKILL.md:537` | Git boundaries | `/myflow-start` | do, finish, fast | **Pointer, no hoist.** The rule `/myflow-start` obeys — "Never commit anything. Stage the planning artifacts and leave the commit to `/myflow-finish`" — is stated in full at the citing site; the citation points at the table for detail. |
| `skills/myflow-contracts/state-file.md:166` | Model policy | every command | start, do, fast | **Pointer, no hoist.** `/myflow-finish` and `/myflow-status` carry `models` forward verbatim, which `state-file.md` states itself; they need the roles' definitions not at all. |
| `skills/myflow-contracts/state-file.md:321` | Resolving a change's worktrees | every command | do, finish, status | **Pointer, no hoist.** `/myflow-start` is the only command loading `state-file.md` but not the destination, and it has no worktrees to resolve. |
| `skills/myflow-contracts/state-file.md:138` | States | every command | every command | Unaffected — States stays in the core. |

No passage is hoisted. Citations sitting in rationale appendices (`myflow-do/SKILL-rationale.md`,
`myflow-finish/SKILL-rationale.md`, `project-configuration-rationale.md`,
`workspace-isolation-rationale.md`) are repointed but cross no load boundary: no run loads an
appendix.

## Verification

Three layers, because no one of them is sufficient.

**1. The per-move ledger.** **Requirement: A move or eviction is recorded in a per-move ledger**
(`openspec/specs/myflow-contract-economy/spec.md`) mandates a four-column table — Removed passage
(first eight words, verbatim), Source heading, Destination, Pointer left — emitted in the task's own
output before its diff reaches the review panel, and checked against the diff in both directions.
That same requirement states a mechanical no-loss check **SHALL NOT** be substituted for it. It is
not substituted here; it is supplemented.

**2. Sentence-set diff.** A throwaway script captures every sentence of `pipeline.md` and
`pipeline-rationale.md` before the first edit, normalises whitespace, and after the last edit
asserts each appears somewhere in the twelve-file family. Any residue is justified in the task's
notes rather than accepted. It is not committed as a guard: it answers a question about this change
only, and a future relocation would have no baseline for it to diff against.

**3. `scripts/check-normative-inventory.sh`** before the first edit and after the last, diffed, as
the ticket requires. It matches `SHALL`/`SHALL NOT`/`MUST`/`MUST NOT` as whole words only — most of
this corpus asserts in bold prose instead — which is precisely why layers 1 and 2 exist.

### Guards and budgets

- `scripts/check-contract-budget.sh` — ten new `budgets()` rows, one per new core and appendix, each
  at the file's landing size plus 25%. **`pipeline.md`'s existing row must be lowered** from 62862
  to its new size plus 25%; left as it is, the ratchet silently permits regrowth to today's size.
  `pipeline-rationale.md`'s row is lowered the same way.
- `scripts/check-references.sh` and `scripts/check-installed-citations.sh` must both exit clean.
  Neither is offered as evidence that a citation still reaches its substance — the table above is.
- `setup.sh` needs no edit. `install_skills` symlinks whole skill directories into
  `~/.claude/skills/`, `~/.cursor/skills/` and `~/.codex/skills/`, so new files under
  `skills/myflow-contracts/` install on the next run with no installer change.

## Decisions

### How far the consumer split goes

**ID:** split-depth
**Status:** active
**Chosen:** Complete consumer split — five extractions — because Rendering the session records and
Resolving a change's worktrees are as unreachable from `/myflow-start` and `/myflow-status` as the
ticket's own three, and leaving them behind would make the core "everything one command happens to
need" rather than "everything every command needs".
**Considered:** The ticket's three extractions exactly — smaller diff, ~20 citations repointed, but
leaves `pipeline.md` at ~33 KB and the core still carrying two single-consumer sections. An
aggressive core also evicting Change name resolution and Stage marks — reaches ~20 KB, but roughly
doubles the citations repointed and evicts Stage marks, which every producing command marks against
on every run.

### Where each extracted file's rationale lives

**ID:** appendix-shape
**Status:** active
**Chosen:** A per-file `<name>-rationale.md` beside each new core, with the matching section moved
out of `pipeline-rationale.md`. This is not a preference: `myflow-contract-economy` states it as two
**SHALL** requirements — the appendix path is derived from the core path, and an appendix mirrors
its core's heading tree.
**Considered:** One shared appendix, every new core pointing back into `pipeline-rationale.md` —
initially chosen for its lower file count, then withdrawn once both requirements were read; it
contradicts each. Amending the two requirements by delta spec was available and rejected, because
the ticket puts changing what a contract says out of scope and a behavioural change riding along
would be invisible in a diff this size. Keeping rationale inline in all five new files — permitted
where reasoning is inseparable, but here it is separable, and it would forfeit the sweep's saving
for `/myflow-do`, the one command loading all five.

### How the no-cut claim is proven

**ID:** no-cut-proof
**Status:** active
**Chosen:** All three layers above. The ledger is the artifact the review panel checks; the
sentence-set diff and the normative inventory are additional nets.
**Considered:** The ledger plus the normative inventory alone, as the ticket specifies — leaves the
bold-assertion class, most of this corpus, checked by reading only. Committing the sentence-set
check as a permanent guard with its own harness, lint row and budgets rows — rejected as roughly
three tasks beyond the ticket for a check no future change could run without a baseline it never
captured.

### Where the load instruction is declared

**ID:** load-declaration-site
**Status:** active
**Chosen:** Each command's own `SKILL.md`, at the step needing the contract. Costs nothing in a
session that runs no `/myflow-*` command, and keeps the always-on rule's table at the meaning it
states — contracts a step reaches for.
**Considered:** Adding five rows to `rules/myflow-manual-review.mdc`'s table — one canonical index,
but roughly +700 bytes in every session forever, against a saving paid only per run, and it would
raise that rule's own budgets row. Pointer stubs inside `pipeline.md`, one per extracted file —
zero always-on cost and no `SKILL.md` edits, but it puts ~1.25 KB back into the file this change
exists to shrink.

### The session-records heading deviation

**ID:** session-records-heading
**Status:** active
**Chosen:** Carry the existing core/appendix heading mismatch — `## Rendering the session records`
against `## Preserving the session records` — into the new pair unchanged, along with the prose
asserting it.
**Considered:** Renaming the appendix heading to match, which would satisfy a strict reading of the
mirroring requirement — rejected because `pipeline.md` deliberately states the appendix keeps its
former name, so the rename is a behavioural edit this change puts out of scope.

## Open questions

*(none — every question raised in this session was answered before the design was written)*
