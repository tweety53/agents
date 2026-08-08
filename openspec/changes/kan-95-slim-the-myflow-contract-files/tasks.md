# Slim the myflow contract files — implementation plan

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to
> implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Cut what every `/myflow-*` command loads per run, remove `/myflow-info`, slim
`/myflow-status`, and delete state self-heal.

**Architecture:** Prose moves out of the files a run loads — into rationale appendices a run never
loads, into `handoff-blocks.md` which only `/myflow-status` loads, or out of the corpus entirely into
`README.md` which nothing loads. Every removal is recorded in a per-move ledger the review panel
checks against the diff.

**Tech Stack:** Markdown contracts under `skills/`, Bash guard scripts under `scripts/`, OpenSpec
capability specs under `openspec/specs/`.

## Global Constraints

- **This repository has no auto-fix command.** Every guard reports `file:line` and is fixed by
  editing the offending line, never by weakening the guard or adding a suppression marker.
- **Lint commands** (all must exit 0 before a task is done): `scripts/check-vocabulary.sh`,
  `scripts/check-references.sh`, `scripts/check-plan-provenance.sh`,
  `scripts/check-task-build-green.sh`, `scripts/check-workspace-isolation.sh`,
  `scripts/check-contract-budget.sh`.
- **No commits.** `/myflow-do` stages; `/myflow-finish` commits. Steps below name `git add`, never
  `git commit`.
- **Every task emits a per-move ledger** for any prose it removes, in the four-column shape
  **A move or eviction is recorded in a per-move ledger**
  (`openspec/changes/kan-95-slim-the-myflow-contract-files/specs/myflow-contract-economy/spec.md`)
  defines. A task that removes no prose emits `ledger: no passages removed`.
- **Rule extraction is permitted only for a mixed passage**, under the three conditions in
  **The split is a verbatim partition, with citation repointing as its only edit** (same file). A
  wholly normative passage stays in the core whole; a wholly rationale passage moves whole.
- **`docs/` is untouched.** Twelve files there mention `/myflow-info`; they are records of past runs.

<!-- measured: wc -c on each file named below @ f763481 -->

Sizes at the start of this change, for the budget re-anchor in Task 10:

| File | Bytes |
|------|-------|
| `skills/myflow-contracts/pipeline.md` | 64,701 |
| `skills/myflow-contracts/project-configuration.md` | 45,274 |
| `skills/myflow-do/SKILL.md` | 37,885 |
| `skills/myflow-contracts/workspace-isolation.md` | 31,079 |
| `skills/myflow-contracts/state-self-heal.md` | 15,250 |
| `skills/myflow-info/SKILL.md` | 5,038 |

---

## File Structure

**Created**

- `skills/myflow-contracts/handoff-blocks.md` — the three per-state handoff block templates, the
  regeneration rules, and the `IN_PROGRESS` rendering-selection table. Loaded by `/myflow-status`.
- `skills/myflow-contracts/project-configuration-rationale.md` — rationale appendix. Loaded by
  nothing.
- `skills/myflow-contracts/workspace-isolation-rationale.md` — rationale appendix. Loaded by nothing.

**Deleted**

- `skills/myflow-info/SKILL.md`, `commands/myflow-info.md`, `commands-claude/myflow-info.md`
- `skills/myflow-contracts/state-self-heal.md`

**Modified** — `skills/myflow-contracts/{pipeline,pipeline-rationale,project-configuration,workspace-isolation,state-file,finish-contract,SKILL}.md`,
`skills/myflow-{do,start,finish,status}/SKILL.md`, `skills/myflow-{do,start,finish}/SKILL-rationale.md`,
`README.md`, `skills/README.md`, `CLAUDE.md`, `AGENTS.md`, `rules/myflow-manual-review.mdc`,
`scripts/check-contract-budget.sh`, `scripts/test-check-contract-budget.sh`.

---

### 1 Remove `/myflow-info` and move the pipeline explanation to `README.md`

**Build:** green

**Files:**
- Delete: `skills/myflow-info/SKILL.md`, `commands/myflow-info.md`, `commands-claude/myflow-info.md`
- Modify: `README.md` (new section), `skills/myflow-contracts/pipeline.md` (delete
  `## Pipeline flow` entirely), `skills/myflow-contracts/pipeline-rationale.md` (delete its mirrored
  `## Pipeline flow` heading), `CLAUDE.md`, `AGENTS.md`, `skills/README.md`,
  `rules/myflow-manual-review.mdc`, `scripts/check-contract-budget.sh`,
  `scripts/test-check-contract-budget.sh`
- Repoint: every file citing `Pipeline flow` (14) or `Level 1 — the stages of each command` (7)

<!-- measured: grep -rl over the working tree excluding .git and archive @ f763481 -->

**Interfaces:**
- Produces: `README.md` carries the state diagram, the level-1 stage table and all eight level-2
  expansions. `pipeline.md` carries neither, and no stub for them — the section leaves the corpus
  rather than being extracted, so nothing points at it.

- [x] **Step 1: Capture the section before deleting it**

Run: `awk '/^## Pipeline flow$/,/^## Command surface$/' skills/myflow-contracts/pipeline.md > /tmp/kan95-pipeline-flow.md && wc -c /tmp/kan95-pipeline-flow.md`
Expected: about 13,769 bytes captured.

<!-- measured: awk section-byte tally over skills/myflow-contracts/pipeline.md @ f763481 -->

- [x] **Step 2: Write the README section, rewritten for a human reader**

Add to `README.md`, after its existing pipeline overview, a `## How the pipeline works` section
carrying: the mermaid state diagram verbatim (a diagram has no audience-specific phrasing to
rewrite); a level-1 table of one row per command with its stages in order and the human gate that
follows it; and one level-2 subsection per stage in the eight-row table under
**The pipeline diagram and its stage table live in `pipeline.md`, and nowhere else**
(`openspec/changes/kan-95-slim-the-myflow-contract-files/specs/myflow-contract-distribution/spec.md`).

Write each level-2 subsection for a reader who is not running a command: state the structure, keep
every citation to the file owning a tuned threshold, and do not restate a threshold.

```markdown verified:copied from skills/myflow-contracts/pipeline.md @ f763481
| Command | Stages, in order | Gate after it |
|---------|------------------|---------------|
| `/myflow-start` | resolve the change → ask the planning effort and the three model choices *(creating run only)* → brainstorm ▸ → design approval → create the OpenSpec artifacts → writing-plans ▸ → publish the proposal artifact → write `STARTED` | you read the proposal artifact |
```

- [x] **Step 3: Delete `## Pipeline flow` from `pipeline.md` and its mirrored heading from the appendix**

Delete the whole section, from `## Pipeline flow` to the line before `## Command surface`. In
`pipeline-rationale.md`, delete the mirrored `## Pipeline flow` heading — the appendix mirrors its
core's heading tree, and a heading whose core section no longer exists breaks that mirror.

- [x] **Step 4: Delete the command and its skill**

```bash verified:authored in-tree for this change
git -C "$WORKTREE" rm -r skills/myflow-info commands/myflow-info.md commands-claude/myflow-info.md
```

- [x] **Step 5: Strip every live reference to `/myflow-info`**

Run: `grep -rn 'myflow-info' . --exclude-dir=.git --exclude-dir=docs --exclude-dir=archive`
Expected after the edits: matches only in `openspec/changes/kan-95-*/` and
`openspec/specs/myflow-{command-surface,contract-distribution,progress-visibility,handoff-output}/spec.md`, whose
delta specs already carry the removal.

Edit `CLAUDE.md` and `AGENTS.md`: remove the `skills/myflow-info/` row from the skill index and the
`/myflow-info` row from the commands table; change "three pipeline commands and two read-only ones"
to "plus one read-only one" wherever it appears. Edit `skills/README.md` and
`rules/myflow-manual-review.mdc` the same way. In `skills/myflow-contracts/pipeline.md`, remove
`/myflow-info` from the `## Command surface` prose, the `## State transitions` table and the
`## Progress visibility` table.

- [x] **Step 6: Drop the budget row and its harness case**

In `scripts/check-contract-budget.sh`, delete the `skills/myflow-info/SKILL.md 6297` row. In
`scripts/test-check-contract-budget.sh`, remove or retarget any fixture case naming that path.

- [x] **Step 7: Emit the ledger and verify**

Emit one ledger row per passage removed from `pipeline.md`, each with destination
`— (rewritten for README.md)`.

Run:
```bash verified:commands taken from .myflow/project.md ## lint and ## test
scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-contract-budget.sh \
  && scripts/test-check-contract-budget.sh && scripts/test-setup.sh
```
Expected: all exit 0.

- [x] **Step 8: Stage**

```bash verified:authored in-tree for this change
git -C "$WORKTREE" add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'
```

---

### 2 Move the handoff block templates to `handoff-blocks.md`

**Build:** green

**Files:**
- Create: `skills/myflow-contracts/handoff-blocks.md`
- Modify: `skills/myflow-contracts/pipeline.md` (leave a stub under `## Handoff output`),
  `skills/myflow-status/SKILL.md` (load the new file), `skills/myflow-contracts/SKILL.md` (index
  row), `scripts/check-contract-budget.sh` (new row)
- Repoint: every file citing `The block each state renders` (19)

<!-- measured: grep -rl over the working tree excluding .git and archive @ f763481 -->

**Interfaces:**
- Consumes: nothing from Task 1.
- Produces: `skills/myflow-contracts/handoff-blocks.md`, carrying `### The block each state renders`
  as a level-3 heading under a `## Handoff blocks` level-2 heading. Task 3 compresses its templates.

- [x] **Step 1: Create the file and move the section whole**

Move `### The block each state renders` and everything under it — the three per-state templates, the
run-only rule, the missing-rather-than-dropped rule, the open-questions justification, the `Git`
line table, and the `IN_PROGRESS` rendering-selection table with its two merge-base paragraphs —
byte-for-byte into `skills/myflow-contracts/handoff-blocks.md`. Apply only the two permitted edits:
repointing a citation whose target is now in a different file, and deleting a position word the move
made false.

The file opens with the same header shape every contract uses:

```markdown verified:shape copied from skills/myflow-contracts/finish-contract.md @ f763481
# Handoff blocks

The per-state handoff block templates and the rules governing their regeneration.

**Loaded by `/myflow-status` and no other command.** A producing command carries the one block it
prints in its own skill and cites this file as the definition.

This file is **canonical** for everything in it.
```

- [x] **Step 2: Leave the stub in `pipeline.md`**

Under `## Handoff output`, which stays, add:

```markdown verified:authored in-tree for this change
### The block each state renders

The block a state hands off is defined in **The block each state renders**
(`skills/myflow-contracts/handoff-blocks.md`), which is canonical for the three per-state templates,
the run-only rule and the rendering-selection table. `/myflow-status` loads it; a producing command
carries only the block it prints.
```

- [x] **Step 3: Repoint the 19 citing files**

Run: `grep -rln 'The block each state renders' . --exclude-dir=.git --exclude-dir=docs --exclude-dir=archive`
Change each citation's backticked path from `skills/myflow-contracts/pipeline.md` to
`skills/myflow-contracts/handoff-blocks.md`, and delete any `above` or `below` the move made false.

- [x] **Step 4: Wire the loaders**

In `skills/myflow-status/SKILL.md`, add the load line for `handoff-blocks.md` at step 4's detail
view. In `skills/myflow-contracts/SKILL.md`, add the index row naming it and stating it is loaded by
`/myflow-status` and no other command.

- [x] **Step 5: Add the budget row**

Add `skills/myflow-contracts/handoff-blocks.md <size>` to `budgets()`, where `<size>` is the file's
actual byte count times 1.25, rounded down. Task 10 re-anchors it after Task 3 compresses the file.

- [x] **Step 6: Emit the ledger and verify**

Run:
```bash verified:commands taken from .myflow/project.md ## lint and ## test
scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-contract-budget.sh
```
Expected: all exit 0. `check-references.sh` is the one that catches a citation this task failed to
repoint.

- [x] **Step 7: Stage**

```bash verified:authored in-tree for this change
git -C "$WORKTREE" add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'
```

---

### 3 Compress the handoff blocks

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/handoff-blocks.md`, `skills/myflow-start/SKILL.md`,
  `skills/myflow-do/SKILL.md`, `skills/myflow-finish/SKILL.md`

**Interfaces:**
- Consumes: `handoff-blocks.md` from Task 2.
- Produces: the folded template each producing skill mirrors. Task 4 renders the `/myflow-status`
  detail view from it.

- [x] **Step 1: Fold the `STARTED` template**

In `handoff-blocks.md`, replace the four separate `Decisions recorded`, `Open questions`,
`Planning effort` and `Models` lines with one folded line:

```text verified:authored in-tree for this change
## Proposal ready — review required

**Change:** <name>
**Artifact:** <artifactUrl, or "missing">
**Recorded:** <N> decisions · <N> open questions · effort <level, or "not recorded — planned at default"> · models <implementation>/<reviewPanel>/<panelFix>
**Jira:** (run-only) <issue key and the transition made, or "none linked", or a skipped-with-reason line>
**Jira description (pre-edit):** (run-only) <the text as it stood before the write, verbatim in a fenced block>

Open in IntelliJ:
open -na "IntelliJ IDEA" --args "<absolute main-checkout path>"

<what the operator does next>

Next:
/myflow-do <name>
```

The `Jira` and `Jira description (pre-edit)` lines stay unfolded: a folded line groups values of one
kind, and those two are run-only while the `Recorded` values are all on disk.

- [x] **Step 2: Fold the `/myflow-do` template**

Fold `Panel`, `Progress` and `Git` as far as their kinds allow. `Panel` is run-only and `Progress`
and `Git` are on-disk, so the fold is `**Staged:** <completed>/<total> tasks · <git state>` with
`**Panel:** (run-only) …` left on its own line.

- [x] **Step 3: Restate the label rule as a folded-line rule**

In `handoff-blocks.md`, change "the label set and the field set are identical" to require the same
folded lines, the same fields within each folded line, and the same order — and add the rule that a
folded line never mixes on-disk and run-only values, with the reason: `/myflow-status` omits one kind
and renders the other, so a mixed line could be neither omitted nor rendered without misreporting
something.

- [x] **Step 4: Mirror the fold into the three producing skills**

Each of `skills/myflow-start/SKILL.md`, `skills/myflow-do/SKILL.md` and
`skills/myflow-finish/SKILL.md` carries its own copy of the block it prints. Apply the identical
fold to each, keeping its enumeration of the literal alternatives that command writes.

**`skills/myflow-finish/SKILL.md` is expected to come out unchanged, and that is not a skipped
step.** Its run-1 block alternates kind field-by-field — `Change` on-disk, `Route` run-only, `PR`
on-disk, `Outstanding` run-only — so no two adjacent values share a kind and step 3's rule forbids
folding any pair. Confirm the field kinds rather than assuming, and leave the block alone if they
alternate. Run 2's terminal block is not a regenerated block at all and is out of scope.

- [x] **Step 5: Verify no field was dropped**

Run: for each of the three blocks, list the values in the pre-fold version and the post-fold version
and diff the two lists.
Expected: identical sets. A dropped value is a defect, not a saving — this task changes layout only.

Run: `scripts/check-references.sh && scripts/check-vocabulary.sh`
Expected: exit 0.

- [x] **Step 6: Stage**

```bash verified:authored in-tree for this change
git -C "$WORKTREE" add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'
```

---

### 4 Slim `/myflow-status` — columns and data sources

**Build:** green

**Files:**
- Modify: `skills/myflow-status/SKILL.md`

**Interfaces:**
- Consumes: the folded templates from Task 3, which the detail view renders.
- Produces: a `/myflow-status` that makes no network call. Task 5 removes its self-heal step.

- [x] **Step 1: Drop the worktree column**

In step 3's rendered table, remove the `Worktree / branch` column and its example values, leaving
`Change | Jira | State | PR | Updated` plus `Next`:

```text verified:authored in-tree for this change
## myflow status

| Change | Jira | State | PR | Next | Updated |
|--------|------|-------|----|------|---------|
| kan-8-myflow-updates | KAN-8 | IN_PROGRESS | #42 | review the diff + run the guide, then `/myflow-finish` | 2h ago (/myflow-do) |
| active-workout-session-editing | — | STARTED | — | read the artifact, then `/myflow-do` | 19h ago (/myflow-start) |
```

Add, below it: the absolute worktree path is given in the detail view, taken from the `worktrees`
keys.

- [x] **Step 2: Remove the `gh pr list` probe**

Delete item 5 of step 2 entirely. Replace the PR column's description: the number is parsed from the
recorded `prUrl`, `—` when it is `null`, and the column never reports whether the pull request is
open, merged or closed, because that answer needs a network call this command no longer makes.

- [x] **Step 3: Keep merge status, unchanged**

Item 4 of step 2 — the three ordered steps, the resolve-before-compare rule, the inconclusive cases
and the multi-repo combination rule — is untouched. It is local git, and it is what splits the
`IN_PROGRESS` row's Next value between integrating and archiving.

- [x] **Step 4: Verify**

Run: `grep -n 'gh pr list' skills/myflow-status/SKILL.md`
Expected: no matches.

Run: `scripts/check-references.sh && scripts/check-vocabulary.sh`
Expected: exit 0.

- [x] **Step 5: Stage**

```bash verified:authored in-tree for this change
git -C "$WORKTREE" add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'
```

---

### 5 Delete state self-heal

**Build:** green

**Files:**
- Delete: `skills/myflow-contracts/state-self-heal.md`
- Modify: `skills/myflow-status/SKILL.md` (remove the self-heal step and the `⚠` legend),
  `skills/myflow-contracts/state-file.md` (re-home the rules that cite the deleted file),
  `skills/myflow-contracts/pipeline.md` (delete `## State self-heal`, rewrite the
  `Resolving a change's worktrees` reason), `skills/myflow-contracts/pipeline-rationale.md`,
  `skills/myflow-contracts/finish-contract.md`, `skills/myflow-contracts/SKILL.md`,
  `rules/myflow-manual-review.mdc`, `CLAUDE.md`, `AGENTS.md`, `skills/README.md`,
  `scripts/check-contract-budget.sh`, `scripts/test-check-contract-budget.sh`

**Interfaces:**
- Consumes: Task 4's `/myflow-status`, which no longer gathers the artifact evidence self-heal needs.
- Produces: a corpus with no self-heal contract and no citation to one.

- [x] **Step 1: Re-home the rules that depend on the deleted file**

`state-file.md` cites `state-self-heal.md` for the closed-schema rule, for the `planningEffort` and
`models` absent-key exception, and for the retired-key paragraph — all of which rest on the notion of
an *unparseable* file that the deleted file defined. State that notion in `state-file.md` itself:

```markdown verified:authored in-tree for this change
**A state file is unparseable when it is not valid JSON, omits a documented field other than
`planningEffort` or `models`, or carries an undocumented one.** An unparseable file is reported and
skipped, never rebuilt from inference — no command infers a state file's contents.
```

Then delete every `**State self-heal** (…)` citation in that file, keeping the rule each one
introduced.

- [x] **Step 2: Remove self-heal from `/myflow-status`**

Delete from `skills/myflow-status/SKILL.md`: the "validate against artifacts" sentence and its
numbered items 1-3, the self-heal rewrite paragraph, the four monotonic-self-heal bullets, the `⚠`
legend line in the rendered table, and the `planningEffort`-is-never-a-`⚠` paragraphs. Replace with
the reporting rule: a state file that is missing or unparseable is named in this command's own output
and the change is omitted from the table, and this command writes nothing.

- [x] **Step 3: Delete the contract and strip its citations**

```bash verified:authored in-tree for this change
git -C "$WORKTREE" rm skills/myflow-contracts/state-self-heal.md
```

In `pipeline.md`, delete the `## State self-heal` section. In `Resolving a change's worktrees`,
rewrite the sentence explaining why the scan lives in `pipeline.md`: the reason was that self-heal
needed it and its performer did not load `finish-contract.md`; the new reason is that `/myflow-do`
and `/myflow-finish` both need it and only one of them loads `finish-contract.md`. Make the
corresponding deletion in `finish-contract.md`, `pipeline-rationale.md`,
`skills/myflow-contracts/SKILL.md`, `rules/myflow-manual-review.mdc`, `CLAUDE.md`, `AGENTS.md` and
`skills/README.md`.

- [x] **Step 4: Drop the budget row and its harness case**

Delete `skills/myflow-contracts/state-self-heal.md 19062` from `budgets()`, and any fixture case in
`scripts/test-check-contract-budget.sh` naming that path.

- [x] **Step 5: Verify**

Run: `grep -rn 'state-self-heal\|State self-heal' . --exclude-dir=.git --exclude-dir=docs --exclude-dir=archive`
Expected: matches only under `openspec/changes/kan-95-*/` and
`openspec/specs/myflow-state-integrity/spec.md`, whose removal delta this change carries.

Run:
```bash verified:commands taken from .myflow/project.md ## lint and ## test
scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-contract-budget.sh \
  && scripts/test-check-contract-budget.sh
```
Expected: all exit 0.

- [x] **Step 6: Stage**

```bash verified:authored in-tree for this change
git -C "$WORKTREE" add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'
```

---

### 6 Evict rationale from `pipeline.md`

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md`, `skills/myflow-contracts/pipeline-rationale.md`

**Interfaces:**
- Consumes: the `pipeline.md` left by Tasks 1, 2 and 5 — no `Pipeline flow`, no block templates, no
  `State self-heal`.
- Produces: a `pipeline.md` carrying rules and pointers only.

- [x] **Step 1: Classify every remaining section**

Sections to work, with the classification test applied per paragraph, bullet or table: `States`,
`Command surface`, `State transitions`, `Wrong state for this command`, `Git boundaries`,
`Preserving the session records`, `Progress visibility`, `Handoff output`, `IntelliJ commands`,
`Resolving a change's worktrees`, `Temporary artifacts registry`, `Model policy`,
`Change name resolution`.

A passage is **core** if removing it would change what an agent does, and **rationale** if removing
it changes no agent behaviour. A prose paragraph stating a rule is core.

- [x] **Step 2: Move the wholly-rationale passages**

Move each into `pipeline-rationale.md` under the mirrored heading, byte-for-byte, with only the two
permitted edits. Named candidates, each carrying a rule stated elsewhere in the same file: the
`&&`-chain-versus-`set -e` argument in `Git boundaries`; the "do not harmonise the two orderings"
paragraph in `Preserving the session records`; the no-third-checkbox-marker paragraph in
`Progress visibility`; the argument for `substr` over `$2` in `Resolving a change's worktrees`; the
override of subagent-driven-development in `Model policy`.

- [x] **Step 3: Extract the mixed passages**

For each mixed passage: move the original to the appendix byte-for-byte; leave in the core the rule
**plus whatever operative detail the rule cannot be followed without**; and cite the appendix heading.
The bound is on kind, not on sentence count — a rule an agent cannot act on without its mechanism
keeps that mechanism in the core. What the core may not gain is a second telling of *why*. Example
shape, for a rule that needs no mechanism beside it:

```markdown verified:authored in-tree for this change
**The `worktree ` path is taken with `substr`, never `$2`** — see **Resolving a change's worktrees**
(`skills/myflow-contracts/pipeline-rationale.md`) for why a field reference truncates a path
containing a space.
```

- [x] **Step 4: Leave a pointer at every eviction**

A core section whose prose is gone and whose pointer is absent is the hollowed section
`scripts/check-references.sh` cannot detect. Every evicted passage leaves one sentence naming the
rule and citing the appendix heading.

- [x] **Step 5: Emit the ledger and verify**

Emit the four-column ledger. Every row's destination heading must exist in `pipeline-rationale.md`.

Run: `scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-contract-budget.sh`
Expected: exit 0.

Run: `wc -c skills/myflow-contracts/pipeline.md`
Expected: roughly 23,000 bytes.
<!-- predicted: wc -c after this task; the projection is from per-section byte tallies and a judgment of which passages are evictable, not a measurement -->

- [x] **Step 6: Stage**

```bash verified:authored in-tree for this change
git -C "$WORKTREE" add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'
```

---

### 7 Evict rationale from `skills/myflow-do/SKILL.md`

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md`, `skills/myflow-do/SKILL-rationale.md`

**Interfaces:**
- Consumes: the folded `/myflow-do` handoff block from Task 3, which sits in this file.
- Produces: nothing later tasks read except its byte count, for Task 10.

- [x] **Step 1: Work the three largest sections**

`## 5. The review panel` is 11,697 bytes, `## 7. Verify, stage, and hand off` 9,700 and
`## 6. Write the manual test guide` 5,033.

<!-- measured: awk section-byte tally over skills/myflow-do/SKILL.md @ f763481 -->

`SKILL-rationale.md` exists and holds 3,327 bytes — KAN-87 created it and moved almost nothing into
it, so the mirrored heading tree is present and mostly empty.

<!-- measured: wc -c skills/myflow-do/SKILL-rationale.md @ f763481 -->

- [x] **Step 2: Move and extract, same rules as Task 6**

Wholly-rationale passages move whole; mixed passages are extracted with the original moved
byte-for-byte and one new core sentence left behind. The slot roster table, the trigger lists, the
zero-open-findings rule, the marker-line shape and every command an agent runs are core and stay.

- [x] **Step 3: Emit the ledger and verify**

Run: `scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-contract-budget.sh`
Expected: exit 0.

Run: `wc -c skills/myflow-do/SKILL.md`
Expected: roughly 26,000 bytes.
<!-- predicted: wc -c after this task; projection, not a measurement -->

- [x] **Step 4: Stage**

```bash verified:authored in-tree for this change
git -C "$WORKTREE" add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'
```

---

### 8 Split `project-configuration.md` into a core and a new appendix

**Build:** green

**Files:**
- Create: `skills/myflow-contracts/project-configuration-rationale.md`
- Modify: `skills/myflow-contracts/project-configuration.md`,
  `skills/myflow-contracts/SKILL.md` (appendix row), `scripts/check-contract-budget.sh` (new row)

**Interfaces:**
- Produces: an appendix whose heading tree mirrors the core's, including an empty heading for any
  section that turns out wholly normative.

- [x] **Step 1: Create the appendix with the mirrored heading tree**

The core's headings are `# Project configuration`, `## Where the agents repository is` and
`## workspace isolation`. The appendix carries the same three, in the same order.

<!-- measured: grep -n '^#\+ ' skills/myflow-contracts/project-configuration.md @ f763481 -->

- [x] **Step 2: Work `## Where the agents repository is`, which is 23,911 bytes**

<!-- measured: awk section-byte tally over skills/myflow-contracts/project-configuration.md @ f763481 -->

The rule is five lines: `AGENTS_DATA` wins when set; otherwise resolve the directory holding the
`SKILL.md` you are reading, following it if it is a symlink, and take two levels above it; confirm
the derived root before using it and treat a miss as an absence. Everything establishing *why* that
is right — the measured global-install layout, the project-local no-op case, the copied-directory
miss case — is rationale.

- [x] **Step 3: Move and extract, same rules as Task 6**

- [x] **Step 4: Add the budget row and index the appendix**

Add `skills/myflow-contracts/project-configuration-rationale.md <size>` to `budgets()`, and an
appendix row to `skills/myflow-contracts/SKILL.md`'s **Rationale appendices** table.

- [x] **Step 5: Emit the ledger and verify**

Run:
```bash verified:commands taken from .myflow/project.md ## lint
scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-contract-budget.sh \
  && scripts/check-workspace-isolation.sh
```
Expected: exit 0. `check-workspace-isolation.sh` reads this file's `## workspace isolation` section
and is the one that catches a rule evicted out of it by mistake.

Run: `wc -c skills/myflow-contracts/project-configuration.md`
Expected: roughly 28,000 bytes.
<!-- predicted: wc -c after this task; projection, not a measurement -->

- [x] **Step 6: Stage**

```bash verified:authored in-tree for this change
git -C "$WORKTREE" add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'
```

---

### 9 Split `workspace-isolation.md` into a core and a new appendix

**Build:** green

**Files:**
- Create: `skills/myflow-contracts/workspace-isolation-rationale.md`
- Modify: `skills/myflow-contracts/workspace-isolation.md`,
  `skills/myflow-contracts/SKILL.md` (appendix row), `scripts/check-contract-budget.sh` (new row)

**Interfaces:**
- Produces: an appendix mirroring the core's five sections.

- [x] **Step 1: Create the appendix with the mirrored heading tree**

The core's sections are `## The workspace id` (8,290 B), `## What the id derives` (5,500),
`## The cache index` (5,362), `## The empty id` (5,419) and `## Creation and cleanup` (4,909).

<!-- measured: awk section-byte tally over skills/myflow-contracts/workspace-isolation.md @ f763481 -->

- [x] **Step 2: Move and extract, same rules as Task 6**

Core here is dense: every derivation formula, the probe procedure for the cache index, the empty-id
rule and the survivor-report contract are core. The rationale is the argument for probing rather than
deriving, the worked examples, and the asymmetry justification in `Creation and cleanup`.

- [x] **Step 3: Add the budget row and index the appendix**

- [x] **Step 4: Emit the ledger and verify**

Run:
```bash verified:commands taken from .myflow/project.md ## lint and ## test
scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-contract-budget.sh \
  && scripts/check-workspace-isolation.sh && scripts/test-check-workspace-isolation.sh
```
Expected: exit 0.

Run: `wc -c skills/myflow-contracts/workspace-isolation.md`
Expected: roughly 20,000 bytes.
<!-- predicted: wc -c after this task; projection, not a measurement -->

- [x] **Step 5: Stage**

```bash verified:authored in-tree for this change
git -C "$WORKTREE" add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'
```

---

### 10 Re-anchor every budget row and record the result

**Build:** green

**Files:**
- Modify: `scripts/check-contract-budget.sh`, `scripts/test-check-contract-budget.sh`

**Interfaces:**
- Consumes: the final byte count of every file Tasks 1-9 touched.
- Produces: the measured per-run load figure the manual test guide reports.

- [x] **Step 1: Measure every covered file**

```bash verified:authored in-tree for this change
wc -c skills/myflow-contracts/*.md skills/*/SKILL.md skills/*/SKILL-rationale.md
```

- [x] **Step 2: Set each row to size × 1.25, rounded down**

This is the rule the guard's own header states — the size the file actually had when the change that
set its row landed, plus 25%. **Do not choose a number the file has not reached**: a target-first
budget creates pressure to push normative text into an appendix to make it.

- [x] **Step 3: Verify the guard covers every file and no row is stale**

Run: `scripts/check-contract-budget.sh && scripts/test-check-contract-budget.sh`
Expected: exit 0, and no row naming a file that no longer exists.

- [x] **Step 4: Measure the per-run load**

```bash verified:authored in-tree for this change
wc -c skills/myflow-do/SKILL.md skills/myflow-contracts/{pipeline,project-configuration,workspace-isolation,state-file,jira-integration}.md | tail -1
```

Starting figure was 210,481 bytes; the projection is roughly 128,500.
<!-- measured: wc -c on those six files @ f763481 for the starting figure -->
<!-- predicted: the same command after Task 9; the target figure is a projection, not a commitment -->

- [x] **Step 5: Run every lint and test command**

```bash verified:commands taken from .myflow/project.md ## lint and ## test
scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-plan-provenance.sh \
  && scripts/check-task-build-green.sh && scripts/check-workspace-isolation.sh \
  && scripts/check-contract-budget.sh
```

```bash verified:commands taken from .myflow/project.md ## test
scripts/test-setup.sh && scripts/test-check-references.sh && scripts/test-check-plan-provenance.sh \
  && scripts/test-check-finish-preflight.sh && scripts/test-preserve-session-records.sh \
  && scripts/test-check-unfinished-work.sh && scripts/test-check-cleanup-complete.sh \
  && scripts/test-gather-self-review-context.sh && scripts/test-uncommitted-review-package.sh \
  && scripts/test-check-task-build-green.sh && scripts/test-check-workspace-isolation.sh \
  && scripts/test-check-contract-budget.sh
```
Expected: all exit 0.

- [x] **Step 6: Stage**

```bash verified:authored in-tree for this change
git -C "$WORKTREE" add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'
```
