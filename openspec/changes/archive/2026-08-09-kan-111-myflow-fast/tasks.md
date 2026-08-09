# myflow-fast — implementation plan

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development
> (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `/myflow-fast` — a command that chains `/myflow-start`'s brainstorming stage and
`/myflow-do`'s implementation stage into one invocation with no human gate between them, and
chains `/myflow-finish`'s two runs into one invocation when the chosen route needs no external
merge — while writing and reading the exact same state file and the exact same three states every
other `/myflow-*` command uses.

**Architecture:** One new skill, `skills/myflow-fast/SKILL.md`, that does not duplicate any other
skill's stage content. Its brainstorming section cites `skills/myflow-start/SKILL.md` sections A–F
verbatim except for the artifact-publish step (E), which it skips; its implementation section
cites `skills/myflow-do/SKILL.md` sections 1–7 verbatim; its integrate section cites
`skills/myflow-finish/SKILL.md`'s run 1 and run 2 sections verbatim, adding only the
argument-presence disambiguation and the auto-continue rule for merge-and-push. One new command
file per tree (`commands/myflow-fast.md`, `commands-claude/myflow-fast.md`). Doc updates name the
new command everywhere the other three are named as a set. No state-file field, no state, and no
git boundary changes.

**Tech Stack:** Markdown skills and commands, Bash guard scripts, the `openspec` CLI. This
repository has no runnable application and no auto-fix command.

## Global Constraints

- **Cite, never duplicate.** `skills/myflow-fast/SKILL.md` states what is new (the chaining, the
  disambiguation, the skipped artifact, the recorded defaults) and points at the existing skill's
  section for everything else. A future change to `/myflow-do`'s panel behavior must not require a
  matching edit in `skills/myflow-fast/SKILL.md`.
- **No new state, no new state-file field, no new git boundary.** This change is a sequencing layer
  over the existing three-state machine. If a task in this plan appears to need a fourth state or a
  new field, stop and re-read `design.md`'s "Non-Goals" — that is a signal the task was
  misunderstood, not a signal to add one.
- **No suppression markers and no guard weakening.** Every lint hit is fixed by editing the
  offending line.
- `scripts/check-contract-budget.sh` is a ratchet keyed on the path relative to the repository
  root. `skills/myflow-fast/SKILL.md` has no row yet — Task 6 adds one, sized from the file's actual
  measured length, matching how every existing row was set (final size plus 25%).
- Every path any pipeline command prints stays absolute.
- The four pipeline commands are always named together as a set. A task that finds a place naming
  "three pipeline commands plus one read-only one" (or the equivalent) updates it to include
  `/myflow-fast`; a task that finds a place naming only `/myflow-start`, `/myflow-do` and
  `/myflow-finish` for a reason specific to those three (e.g. which command performs a git action)
  leaves it alone and does not add `/myflow-fast` where it does not belong.

## Baseline

The files this plan edits most, and their declared budgets, measured before any task runs:

| File | Size now | Budget in `budgets()` |
|------|----------|-----------------------|
| `skills/myflow-contracts/pipeline.md` | 33661 | 40161 |
| `skills/myflow-start/SKILL.md` | 27379 | 32465 |
| `skills/myflow-do/SKILL.md` | 41046 | 44781 |
| `skills/myflow-finish/SKILL.md` | 26795 | 32050 |

<!-- measured: for f in <the four paths>; do wc -c < "$f"; done, on the working tree at the start of this change -->

Every file has headroom for the small additions this plan makes to them (a command-surface row, a
model-policy cross-reference, a citation). `skills/myflow-fast/SKILL.md` is new and gets its
budget row in Task 6, from its own measured size — not estimated in advance.

---

### 1 The `myflow-fast` skill — brainstorming-into-implementation

**Build:** green

**Files:**
- Create: `skills/myflow-fast/SKILL.md`

**Interfaces:**
- Consumes: nothing from earlier tasks — this is the first task.
- Produces: the skill file every later task's command file and doc update points at.

- [x] **Step 1: Write the skill's frontmatter and opening**

Follow `skills/myflow-start/SKILL.md`'s opening shape: an announce line
(`"Using myflow-fast for change <name>."`), the `/rename` + `/color cyan` print block per
**Handoff output** (`skills/myflow-contracts/pipeline.md`), a line requiring
`skills/myflow-contracts/pipeline.md` to load first, and a task-list registration line per
**Progress visibility** (`skills/myflow-contracts/pipeline.md`).

- [x] **Step 2: State the state gate**

```markdown unverified:authored in-tree for this change; matches the shape skills/myflow-start/SKILL.md's own "State gate" section uses
## State gate

Accepts **no state** (creates a change) or **`IN_PROGRESS`**. On any other state, emit the
wrong-state handoff from **Wrong state for this command**
(`skills/myflow-contracts/pipeline.md`), naming the actual state, the states this command expects,
and the suggested command instead. Proceed only on an explicit override.
```

- [x] **Step 3: Write the "no state file" branch — brainstorming**

State that this branch runs `skills/myflow-start/SKILL.md` sections A through D **exactly as
written** — Jira resolution and naming, the planning-effort/models/roster questions, brainstorming
(#1) in full with its design-approval gate, OpenSpec artifact creation, and writing-plans (#3) — by
citation, not by re-deriving their content. Name each section by its letter and title so a reader
can find the source.

State the one difference from `/myflow-start`: skip section E (publish the proposal artifact)
entirely. `artifactUrl` is written as `null` in the state file (Step 6 below), and the handoff
(Step 8) does not offer an artifact link.

- [x] **Step 4: Write the continuation into implementation**

State that once section D (writing-plans) completes and the OpenSpec artifacts exist, this skill
continues — within the same invocation, with no further command from the operator — into
`skills/myflow-do/SKILL.md` sections 1 through 7 **exactly as written**: load context and validate
the plan, isolate the workspace, SDD + TDD per task, the review panel (dispatched at the roster
recorded in Step 2's question round), resolve the run instructions, and verify/stage/hand off.

State plainly that this is the one point in the pipeline where two skills' stage content runs
back-to-back inside a single command invocation with no operator action between them, and that this
is deliberate: there is no human gate between brainstorming converging and implementation starting
in the existing pipeline either, so nothing that used to pause now doesn't.

- [x] **Step 5: Write the `IN_PROGRESS` branch — argument disambiguation**

```markdown unverified:authored in-tree for this change; the two sub-branches below cite the existing skills rather than restating their content
## At IN_PROGRESS

**An argument present** — treat it as fix instructions. Run `skills/myflow-do/SKILL.md`'s
"Documenting a fix, before implementing it" (section 3) and the rest of sections 1–7 exactly as
`/myflow-do` runs them at `IN_PROGRESS`, using the argument text as the fix's guidance. The state
is written back unchanged, per **A fix never moves the state** (`skills/myflow-contracts/pipeline.md`).

**No argument (bare invocation)** — proceed to the integrate question. Run
`skills/myflow-finish/SKILL.md`'s "Deciding which run this is" through "1.3 Take the chosen route"
exactly as `/myflow-finish` run 1 runs them: the unfinished-work gate, the landing question (open
PR / merge and push / manual), the two commits, and the chosen route.
```

- [x] **Step 6: Write the merge-and-push auto-continuation**

```markdown unverified:authored in-tree for this change
### After merge-and-push specifically

Continue, within the same invocation and without a further command from the operator, into
`skills/myflow-finish/SKILL.md`'s run 2 procedure exactly as written: verify the merge, sync delta
specs, archive the change, commit and push the archive, remove the worktrees/branches, verify the
cleanup, and write `FINISHED`.

### After open PR or manual specifically

Stop after the route completes, printing the same handoff `skills/myflow-finish/SKILL.md`'s run 1
prints. Each of these two routes needs an action outside this command's control — an external
merge, or the operator's own manual steps — before archiving can happen, so nothing continues
automatically.
```

- [x] **Step 7: Write the recorded-defaults note**

State that the planning-effort/models/roster question round this skill runs (by citing
`skills/myflow-start/SKILL.md` section A) recommends `sonnet` for `models.implementation`,
`models.reviewPanel` and `models.panelFix`, and `light` for `reviewPanelRoster` — as the
**recommended** answer in each `AskUserQuestion` prompt, never a forced value. An operator who
answers differently has that answer recorded, exactly as under `/myflow-start`.

- [x] **Step 8: Write the state-write and handoff step**

State that the state file is written exactly per **State file**
(`skills/myflow-contracts/state-file.md`) — same shape, same fields, `updatedBy: "/myflow-fast"` —
with `artifactUrl: null` on the creating branch (Step 3) and every other field written by whichever
cited section actually ran (`/myflow-do`'s section 7 state write when ending at `IN_PROGRESS` from
brainstorming+implementation; `/myflow-finish`'s run 1 or run 2 state write when ending from the
`IN_PROGRESS` branch).

Add a handoff block for the `IN_PROGRESS`-with-no-artifact case specifically — the one shape none
of the three cited skills already prints, since none of them skips the artifact:

```markdown unverified:authored in-tree for this change; matches the shape of the other skills' own handoff blocks
## Implementation staged — review and test

**Change:** <name>
**Artifact:** none — myflow-fast does not publish one
**Recorded:** ... (same **Recorded:** line shape `/myflow-start`'s handoff uses)

<the rest of the block, identical to /myflow-do's own "Implementation staged" handoff — review the
staged diff and run the apps>

Next:
/myflow-fast <name>
```

Every other handoff shape (bare-at-`IN_PROGRESS` integrate handoffs, `FINISHED`) is exactly the
cited skill's own block — no new shape needed there.

- [x] **Step 9: Write the Guardrails section**

Carry forward, by citation, every guardrail `skills/myflow-start/SKILL.md`,
`skills/myflow-do/SKILL.md` and `skills/myflow-finish/SKILL.md` already state for the sections this
skill cites — do not re-list them. Add exactly the guardrails specific to this skill:

- Never auto-answer a brainstorming question. The design-approval gate is interactive, unchanged.
- Never continue from brainstorming into implementation before the design is approved and the
  OpenSpec artifacts exist.
- Never continue past open PR or manual without stopping — only merge-and-push auto-continues to
  run 2.
- Never treat a bare invocation at `IN_PROGRESS` as a fix, and never treat an invocation carrying
  an argument as ready-to-integrate.
- Never publish a proposal artifact.

- [x] **Step 10: Verify**

```bash verified:both commands are declared under `## lint` in this repository's .myflow/project.md
scripts/check-references.sh
scripts/check-vocabulary.sh
```

Expected: both exit 0. `check-references.sh` resolves every section citation this task added
(`skills/myflow-start/SKILL.md` sections A–D, `skills/myflow-do/SKILL.md` sections 1–7,
`skills/myflow-finish/SKILL.md`'s run 1/run 2 sections, and the contract files).

**Discovered during implementation:** `scripts/check-vocabulary.sh`'s retired-vocabulary blacklist
carries `myflow-fast([^-]|$)` — a leftover from the five-state pipeline's own retired `/myflow-fast`
command (folded into `/myflow-do`/`/myflow-finish`, distinct from this KAN-111 command of the same
name), added alongside `myflow-test([^-]|$)` and `myflow-review([^-]|$)` for the same collapse. It
matches every ordinary mention of the new command's name and fails Step 10's `check-vocabulary.sh`
run unconditionally. Confirmed with the operator on 2026-08-09: keep the name `myflow-fast`, scope
the guard instead — Task 1a below does that, the same shape as the `code-review` exception this
guard already carries.

---

### 1a Scope the retired-vocabulary guard past the name collision

**Build:** green

**Files:**
- Modify: `scripts/check-vocabulary.sh` — the `myflow-fast([^-]|$)` blacklist entry and its
  surrounding comment

**Interfaces:**
- Consumes: nothing from earlier tasks — this fixes a pre-existing guard, not something Task 1
  produced.
- Produces: a guard that admits the new command's name everywhere Task 1, 2 and 5 write it, while
  still catching the old retired command's own spellings.

- [x] **Step 1: Narrow the pattern**

The retired command was `/myflow-fast`, always as a bare command token — never followed by a
hyphen in this repository's history (`myflow-fast-path` is a *different*, separately-listed retired
name). The new command is always written either as the bare command `/myflow-fast`, or as a path
component followed by `/` or `.md` (`skills/myflow-fast/`, `commands/myflow-fast.md`), or inside
backticks. Replace the blanket `myflow-fast([^-]|$)` with a pattern that still catches the retired
command's actual attested shapes — a lone word boundary on both sides, matched only when not
immediately preceded by `/` (a path segment) or `` ` `` (an inline-code start) and not immediately
followed by `/` or `.` (a path continuation): `(?<![/\`])\bmyflow-fast\b(?![/.])` if the guard's
grep supports lookaround, or the nearest ERE-compatible equivalent this script's existing entries
already use (check how the `code-review` exception above expresses its own boundary — this file
uses plain ERE, not PCRE, so follow its actual established idiom rather than assuming lookaround is
available).

- [x] **Step 2: Document the exception in the guard's own comment block**

Immediately above the modified line, add a comment in the same style the `code-review` exception's
comment block already uses (see the `# vocab-guard:allow`-tagged block a few lines above,
documenting `requesting-code-review` and `code-review`): state that `myflow-fast` is now a real,
live command name (KAN-111), name the operator-approved date (2026-08-09) the same way the
`code-review` carve-out's own comment names its approval, and state precisely what the narrowed
pattern still catches (the old retired command's bare-token spelling) and what it now admits (the
new command's path and backtick spellings).

- [x] **Step 3: Verify**

```bash verified:declared under `## lint` in this repository's .myflow/project.md
scripts/check-vocabulary.sh
```

Expected: exit 0 against the tree as it stands after Task 1 (this file's own retired-command
mentions above are in `tasks.md`, under `openspec/`, which this guard does not scan — confirm that
by re-reading the guard's own file-selection logic before relying on it). Also run a synthetic
negative check: confirm the guard still flags a bare, non-path, non-backtick mention of the *old*
retired command shape, so the narrowing did not accidentally blind the guard to its original catch.
Use a scratch string via the guard's own matching logic (not a committed file) for this check — do
not write a retired-vocabulary violation into any tracked file to test it.

---

### 2 The command files

**Build:** green

**Files:**
- Create: `commands-claude/myflow-fast.md`
- Create: `commands/myflow-fast.md`

**Interfaces:**
- Consumes: the skill from Task 1.
- Produces: the two files `setup.sh` installs verbatim (it globs both trees; no registration list
  to update).

- [x] **Step 1: Write `commands-claude/myflow-fast.md`**

Follow `commands-claude/myflow-do.md`'s exact shape (frontmatter, then prose) as the template:

```markdown unverified:authored in-tree for this change; frontmatter and structure copied from commands-claude/myflow-do.md's own shape
---
model: sonnet
description: Fast — brainstorm, implement, and integrate in one command, pausing only at the human gates
---

Use the **myflow-fast** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

Follow that skill exactly. Accepts **no state** (creates a change) or **`IN_PROGRESS`**. On a
creating run it runs brainstorming (unchanged, fully interactive) and then, in the same
invocation, implementation and the review panel, ending at `IN_PROGRESS`. Re-invoked with an
argument at `IN_PROGRESS`, the argument is fix instructions. Re-invoked bare at `IN_PROGRESS`, it
asks how to land the branch; merge-and-push continues in the same invocation through archive to
`FINISHED`, while open PR and manual stop and hand off.

Publishes no proposal artifact — the operator is present for the brainstorming dialogue that
produces the design.

Also follow the myflow rule (`myflow-manual-review.mdc`) — installed globally, so let your harness
resolve it rather than assuming a project-local path. It is a stub: **load
`skills/myflow-contracts/pipeline.md` first**, which is canonical for the states, transitions, git
boundaries and the finish contract.

**Input:** the change name or a description/Jira key to seed a new change, from `$ARGUMENTS` or
the conversation — and nothing else. **This command takes no flags.** If omitted at `IN_PROGRESS`,
run `openspec list --json` and use the sole relevant open change, asking which when there are
several. Report any argument that is not a change name, description, or fix instruction rather
than ignoring it.

**When done:** at `IN_PROGRESS`, review the staged diff and run the apps, then re-run
`/myflow-fast <name>` (or `/myflow-fast <name> <fix>`) as needed. At `FINISHED`, nothing further.
```

- [x] **Step 2: Write `commands/myflow-fast.md`**

Follow `commands/myflow-do.md`'s exact shape — same content as Step 1's file, but with the
`name`/`id`/`category` frontmatter fields `commands/myflow-do.md` uses instead of `model:`, plus
that file's own "Model:" prose paragraph pattern (session on Sonnet by default; implementer
subagents dispatched by the brainstorming→implementation branch run on **Sonnet** here — this
change's recorded default, per `design.md`'s "Model defaults: Sonnet everywhere" — named explicitly
on each dispatch, never inherited, exactly as `commands/myflow-do.md` already states for its own
implementer dispatches, substituting Sonnet where that file says Opus).

- [x] **Step 3: Verify**

```bash verified:both commands are declared under `## lint` in this repository's .myflow/project.md
scripts/check-references.sh
scripts/check-vocabulary.sh
```

Expected: both exit 0.

---

### 3 `skills/myflow-contracts/pipeline.md` — command surface and state transitions

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md` — the "Command surface" paragraph, the "State
  transitions" table, and the "Change name resolution" section's description of the union (already
  generic — confirm it needs no edit rather than assuming)

**Interfaces:**
- Consumes: the skill from Task 1.
- Produces: the canonical command-surface fact every doc-update task (Task 5) cites back to.

- [x] **Step 1: Extend "Command surface"**

The paragraph currently reads "Three pipeline commands, plus one read-only one." Rewrite it to
name four pipeline commands plus one read-only one, with `/myflow-fast` described in one clause as
the composite command that chains the other three's stage content across gateless transitions —
citing `skills/myflow-fast/SKILL.md` rather than re-describing its behavior.

- [x] **Step 2: Extend "State transitions"**

Add a row to the table:

```markdown unverified:authored in-tree for this change; matches the existing table's column shape exactly
| `/myflow-fast` | *(none — creates the change)* · `IN_PROGRESS` | `IN_PROGRESS` from none; `IN_PROGRESS` or `FINISHED` from `IN_PROGRESS`, depending on the route chosen |
```

Confirm the sentence immediately below the table ("This table is authoritative...") still reads
correctly with the new row — it should, since it already speaks of "every command" rather than
enumerating the three by name.

- [x] **Step 3: Confirm "Change name resolution" needs no edit**

Read the section once. It already speaks of "every `/myflow-*` command" rather than naming the
three by identifier, so `/myflow-fast` is already covered. Make no edit here; this step exists so
the plan does not silently skip checking.

- [x] **Step 4: Verify**

```bash verified:both commands are declared under `## lint` in this repository's .myflow/project.md
scripts/check-references.sh
scripts/check-contract-budget.sh
```

Expected: both exit 0.

---

### 4 `skills/myflow-contracts/state-file.md` — confirm no edit needed

**Build:** green

**Files:**
- (none modified — this task verifies, and only edits if verification finds otherwise)

**Interfaces:**
- Consumes: nothing new.
- Produces: a recorded confirmation that this change introduces no state-file field.

- [x] **Step 1: Read the field list and confirm it is unchanged**

`/myflow-fast` reads and writes `state`, `branch`, `worktrees`, `artifactUrl`, `jiraIssue`,
`planningEffort`, `models`, `reviewPanelRoster`, `prUrl`, `updatedAt`, `updatedBy` — the exact set
`state-file.md` already documents, with no new key. `updatedBy` already accepts any command name as
a string; no enumeration of command names exists there to extend.

If this reading is wrong — if some step in Task 1 turns out to need a field this file does not
document — stop and revise Task 1 rather than adding an undocumented key here.

- [x] **Step 2: Verify**

```bash verified:both commands are declared under `## lint` in this repository's .myflow/project.md
scripts/check-references.sh
scripts/check-contract-budget.sh
```

Expected: both exit 0, confirming no accidental edit crept in.

---

### 5 Documentation digests — `CLAUDE.md`, `AGENTS.md`, `README.md`

**Build:** green

**Files:**
- Modify: `CLAUDE.md` — the command table (`### /myflow commands summary`) and the skill index
  (`### Skill index`)
- Modify: `AGENTS.md` — its own skill index and command table (same content, this file's own
  format)
- Modify: `README.md` — the directory-listing comment block, the "Skills (loaded on demand)"
  sentence, the pipeline mermaid diagram, the per-command stage table under "How the pipeline
  works", and the Cursor/Codex command tables further down

**Interfaces:**
- Consumes: the wording settled in Tasks 1–3.
- Produces: nothing later tasks depend on.

- [x] **Step 1: Confirm what each file actually says before editing**

```bash verified:authored in-tree for this change
grep -rn "myflow-start\`.*myflow-do\`.*myflow-finish\|three pipeline commands\|plus one read-only" CLAUDE.md AGENTS.md README.md
```

Each hit is a candidate — a place the three (or four, with status) pipeline commands are named as
a set. Edit only those; leave every place that names one command for a reason specific to it (e.g.
"only `/myflow-finish`... may create a commit") alone, per this plan's Global Constraints.

- [x] **Step 2: `CLAUDE.md`**

Add a `/myflow-fast` row to `### /myflow commands summary`'s table, describing it in the same style
the existing rows use — citing `skills/myflow-fast/SKILL.md` for its stage sequence rather than
repeating it, exactly as the `/myflow-start` row cites `skills/myflow-start/SKILL.md`. Add a
`skills/myflow-fast/` row to `### Skill index`, styled like the other four rows. Update the
`/myflow commands summary` digest's pipeline-state block only if the three-line
`STARTED`/`IN_PROGRESS`/`FINISHED` digest itself needs a fourth line — per this file's own stated
rule, that digest carries the **states and transitions**, not every command that can reach them, so
confirm before editing: `/myflow-fast` reaches states the digest already names, so the digest
itself likely needs no change. Record which conclusion was reached.

- [x] **Step 3: `AGENTS.md`**

Mirror Step 2's two additions (skill index row, command table row) in this file's own format.

- [x] **Step 4: `README.md` — the easy spots**

Add `myflow-fast/` to the directory-listing comment block (near the `myflow-start/`,
`myflow-do/`, `myflow-finish/`, `myflow-status/` lines), and add `/myflow-fast` to the "Skills
(loaded on demand)" sentence.

- [x] **Step 5: `README.md` — the mermaid diagram**

Add `/myflow-fast`'s transitions to the state diagram: a `[*] --> IN_PROGRESS: /myflow-fast` edge
would be wrong, since a creating run passes through `STARTED` internally before implementation
starts — represent it as the diagram already represents `/myflow-do`'s combined
start-through-implementation reach is *not* shown as a single edge for `/myflow-do` either (it
shows `[*] --> STARTED: /myflow-start` and `STARTED --> IN_PROGRESS: /myflow-do` as two edges).
Add `/myflow-fast` as its own edge set reflecting what it actually does in one invocation:
`[*] --> IN_PROGRESS: /myflow-fast (brainstorm + implement)`, plus
`IN_PROGRESS --> IN_PROGRESS: /myflow-fast (fix — argument present)` and
`IN_PROGRESS --> FINISHED: /myflow-fast (merge+push route)`. Read the diagram's existing edge-label
style before adding these and match it exactly.

- [x] **Step 6: `README.md` — the per-command stage table**

Add a `/myflow-fast` row to the table under "How the pipeline works", in this file's existing
column shape (stage sequence with `▸` markers where a sub-sequence is detailed further down, then
what the operator waits on). Cite `skills/myflow-fast/SKILL.md` for the sequence rather than
re-deriving it, matching how the `/myflow-finish` row already cites its own skill for run 1 and run
2 separately.

- [x] **Step 7: `README.md` — the Cursor/Codex sections**

The command tables further down (around the "Model:" paragraphs for Cursor and Codex) each list
the pipeline commands again in that harness's own frontmatter shape. Add a `/myflow-fast` row to
each, matching Task 2's two command files for the actual frontmatter/model content.

- [x] **Step 8: Verify**

```bash verified:all commands used here are declared under `## lint` in this repository's .myflow/project.md
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-contract-budget.sh
```

Expected: all three exit 0.

---

### 6 Budget row for the new skill file

**Build:** green

**Files:**
- Modify: `scripts/check-contract-budget.sh` — add one row to `budgets()`

**Interfaces:**
- Consumes: the finished `skills/myflow-fast/SKILL.md` from Task 1 (must run after Task 1's file
  reaches its final size, i.e. after any later task that might still touch it — confirm none does).
- Produces: a budget row future changes to this file will ratchet against.

- [x] **Step 1: Measure the finished file**

```bash verified:this is the same measurement method the Baseline table above and every prior change's baseline used
wc -c < skills/myflow-fast/SKILL.md
```

- [x] **Step 2: Add the row**

In `scripts/check-contract-budget.sh`'s `budgets()` heredoc, add a line for
`skills/myflow-fast/SKILL.md`, alphabetically ordered among the other `skills/myflow-*` rows,
valued at the Step 1 measurement plus 25% (rounded as the existing rows are — check one existing
row's exact rounding convention against its own file's current size before choosing a rounding
rule, rather than assuming).

- [x] **Step 3: Verify**

```bash verified:declared under `## lint` in this repository's .myflow/project.md
scripts/check-contract-budget.sh
```

Expected: exit 0, with the new row present in the guard's own listing of checked files.

---

### 7 Full guard and test sweep

**Build:** green

**Files:**
- Modify: `scripts/check-contract-budget.sh` — only if a row other than Task 6's own new one is
  exceeded

**Interfaces:**
- Consumes: every earlier task.
- Produces: a green tree.

- [x] **Step 1: Run every declared lint command**

```bash verified:this is the `## lint` list in this repository's .myflow/project.md, in order
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
scripts/check-task-build-green.sh
scripts/check-workspace-isolation.sh
scripts/check-contract-budget.sh
```

Expected: every one exits 0.

- [x] **Step 2: Raise a budget row only if one is exceeded**

If `check-contract-budget.sh` names a path other than `skills/myflow-fast/SKILL.md` (Task 6 already
sized that one correctly), raise **that row only**, to the file's new size plus a quarter, matching
how every existing row was set. Do not narrow the guard's scope and do not delete a row. The
Baseline table above shows headroom on every file this plan edits besides the new one, so a hit
here means a task grew a file more than planned — read the diff before raising the row.

- [x] **Step 3: Run every declared test harness**

```bash verified:this is the `## test` list in this repository's .myflow/project.md, in order
scripts/test-setup.sh
scripts/test-check-references.sh
scripts/test-check-plan-provenance.sh
scripts/test-check-finish-preflight.sh
scripts/test-preserve-session-records.sh
scripts/test-check-unfinished-work.sh
scripts/test-check-cleanup-complete.sh
scripts/test-gather-self-review-context.sh
scripts/test-uncommitted-review-package.sh
scripts/test-check-task-build-green.sh
scripts/test-check-workspace-isolation.sh
scripts/test-check-contract-budget.sh
scripts/test-check-vocabulary.sh
```

Expected: every one exits 0. `test-setup.sh` in particular exercises the installer, which is what
proves the two new command files and the new skill directory actually get installed — no test in
this list asserts anything about `/myflow-fast`'s own runtime behavior (there is no automated
Claude-Code-level harness in this repository for that), so a failure here means a task broke
something unrelated, not that `/myflow-fast` itself misbehaves.

- [x] **Step 4: Exercise the installer in a sandbox**

```bash verified:this is the `## run` recipe in this repository's .myflow/project.md
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
```

Expected: exit 0, and `myflow-fast.md` present under the sandbox's `.claude/commands/` and
`skills/myflow-fast/SKILL.md` present under the sandbox's `.claude/skills/`. Never run the
installer against the real home directory to check this.

- [x] **Step 5: Read the whole diff against the delta specs**

Open `openspec/changes/kan-111-myflow-fast/specs/myflow-fast-command/spec.md` and
`openspec/changes/kan-111-myflow-fast/specs/myflow-command-surface/spec.md` and confirm every
requirement has a corresponding edit. The requirements no script can check are the ones to read
most carefully: that brainstorming stays fully interactive with no auto-answering, that the
merge-and-push route actually continues into run 2 within the same invocation, and that open PR and
manual both stop rather than continuing.
