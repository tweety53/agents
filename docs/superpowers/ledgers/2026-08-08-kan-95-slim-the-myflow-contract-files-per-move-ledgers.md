# Per-move ledgers — kan-95-slim-the-myflow-contract-files

The evidence `A move or eviction is recorded in a per-move ledger` requires: one row per
passage moved, evicted or extracted, across all ten tasks and all five panel fix rounds.

Preserved by hand. `scripts/preserve-session-records.sh` looks for the SDD ledger at
`.superpowers/sdd/tasks/progress.md`, derived from the plan's filename; this run wrote it to the
flat `.superpowers/sdd/progress.md`, so the script reported `skipped … (absent)` and would have
left every ledger below to be destroyed with the worktree at run 2 — KAN-77's exact shape.


---

## task-1-ledger.md

# Task 1 — per-move ledger

Every passage below was removed from `skills/myflow-contracts/pipeline.md`'s `## Pipeline flow`
section (the whole section, from the heading through the line before `## Command surface`) and
rewritten for a human reader into `README.md`'s new `## How the pipeline works` section, per
**Requirement: The pipeline diagram and its stage table live in `pipeline.md`, and nowhere else**
(`openspec/changes/kan-95-slim-the-myflow-contract-files/specs/myflow-contract-distribution/spec.md`).
The mirrored heading tree in `skills/myflow-contracts/pipeline-rationale.md` (`## Pipeline flow`
through `#### Self-review — /myflow-finish run 2`) was deleted alongside it, since a rationale
appendix's headings mirror its core and a heading whose core section no longer exists breaks that
mirror; the rationale prose itself is not separately ledgered because none of it survives as a
citable section (it argued for keeping the tables in the core, which this task reverses).

| Removed passage | Source heading | Destination | Pointer left |
|---|---|---|---|
| ` ```mermaid stateDiagram-v2 [*] --> STARTED: /myflow-start STARTED --> ` | Pipeline flow | — (rewritten for README.md) | none |
| **One row per command** — the five this | Level 1 — the stages of each command | — (rewritten for README.md) | none |
| Each expansion states the **structure** — the shape | Level 2 — the stages that hide substructure | — (rewritten for README.md) | none |
| superpowers:brainstorming runs its checklist in full and ends | Brainstorm — `/myflow-start` | — (rewritten for README.md) | none |
| superpowers:writing-plans enriches `tasks.md` from a checkbox scaffold into | Writing-plans — `/myflow-start` | — (rewritten for README.md) | none |
| One implementer dispatch per checkbox in `tasks.md`, or | SDD + TDD per task — `/myflow-do` | — (rewritten for README.md) | none |
| **Three required slots and four conditional ones.** Primary | The review panel — `/myflow-do` | — (rewritten for README.md) | none |
| `scripts/check-finish-preflight.sh` decides which run happens, from three | The preflight verdict — `/myflow-finish` | — (rewritten for README.md) | none |
| Runs **before** the landing question and before any | The unfinished-work gate — `/myflow-finish` run 1 | — (rewritten for README.md) | none |
| The operator is asked once, before any git | The landing routes — `/myflow-finish` run 1 | — (rewritten for README.md) | none |
| Every removal is *remove-or-move if present*, which is | Cleanup — `/myflow-finish` run 2 | — (rewritten for README.md) | none |
| Self-review runs only after `FINISHED` is written — | Self-review — `/myflow-finish` run 2 | — (rewritten for README.md) | none |

**Note on the last row.** Unlike the other eight `####` subsections, `Self-review` is not one of the
eight stages **The pipeline diagram and its stage table live in `pipeline.md`, and nowhere else**
names for Level 2 expansion — the delta spec's table of required expansions has exactly eight rows
and omits it. So this row's content is not reproduced in README's Level 2; README's Level 1 table
still names `self-review` as a stage of `/myflow-finish` run 2 (with no `▸`, since it carries no
expansion) and points to
**Requirement: Self-review runs only after FINISHED is written**
(`openspec/specs/myflow-self-review/spec.md`) — already the fuller, normative source this
subsection's own text deferred to — as the procedure's home. The `— (rewritten for README.md)`
destination is used for this row too, per this task's brief, since the passage still leaves the
loaded corpus for a document no command reads; the row above states precisely what "rewritten" means
for this one case, since it is thinner than for the other eleven rows.

---

## task-10-ledger.md

ledger: no passages removed

## Per-row budget re-anchor table

Every row in `scripts/check-contract-budget.sh`'s `budgets()` table was recomputed as
`floor(actual_bytes × 1.25)`, using the actual byte count of the file as it stands at the end of
this change (Task 10, the final task). This is the rule the guard's own header states: the budget
is the size the file actually had when the change that set its row landed, plus 25% — never a
target the file was made to hit.

| File | Actual bytes | Old budget | New budget |
|------|--------------:|-----------:|-----------:|
| `skills/myflow-contracts/SKILL.md` | 5,988 | 6,957 | 7,485 |
| `skills/myflow-contracts/build-green.md` | 3,646 | 4,557 | 4,557 |
| `skills/myflow-contracts/finish-contract.md` | 26,786 | 32,037 | 33,482 |
| `skills/myflow-contracts/handoff-blocks.md` | 14,063 | 16,765 | 17,578 |
| `skills/myflow-contracts/jira-followups.md` | 35,721 | 44,651 | 44,651 |
| `skills/myflow-contracts/jira-integration-rationale.md` | 2,472 | 3,090 | 3,090 |
| `skills/myflow-contracts/jira-integration.md` | 15,591 | 19,488 | 19,488 |
| `skills/myflow-contracts/pipeline-rationale.md` | 27,591 | 34,518 | 34,488 |
| `skills/myflow-contracts/pipeline.md` | 31,151 | 80,876 | 38,938 |
| `skills/myflow-contracts/plan-provenance.md` | 24,466 | 30,582 | 30,582 |
| `skills/myflow-contracts/project-configuration-rationale.md` | 13,225 | 16,531 | 16,531 |
| `skills/myflow-contracts/project-configuration.md` | 38,275 | 56,592 | 47,843 |
| `skills/myflow-contracts/state-file.md` | 15,362 | 19,938 | 19,202 |
| `skills/myflow-contracts/workspace-isolation-rationale.md` | 11,037 | 13,796 | 13,796 |
| `skills/myflow-contracts/workspace-isolation.md` | 25,317 | 38,848 | 31,646 |
| `skills/myflow-do/SKILL-rationale.md` | 7,732 | 8,370 | 9,665 |
| `skills/myflow-do/SKILL.md` | 35,825 | 47,356 | 44,781 |
| `skills/myflow-finish/SKILL-rationale.md` | 3,210 | 4,005 | 4,012 |
| `skills/myflow-finish/SKILL.md` | 25,640 | 32,005 | 32,050 |
| `skills/myflow-start/SKILL-rationale.md` | 3,660 | 4,588 | 4,575 |
| `skills/myflow-start/SKILL.md` | 25,956 | 32,498 | 32,445 |
| `skills/myflow-status/SKILL.md` | 14,120 | 21,861 | 17,650 |
| `skills/openspec-explore/SKILL.md` | 11,428 | 14,285 | 14,285 |

23 covered files, 23 rows — one-to-one, no file uncovered, no row naming a deleted file. The two
files deleted during this change, `skills/myflow-contracts/state-self-heal.md` and
`skills/myflow-info/SKILL.md`, already carried no row before this task's edit (an earlier task in
this change had already removed them); confirmed absent from both `scripts/check-contract-budget.sh`
and its fixtures in `scripts/test-check-contract-budget.sh`, so no removal was needed in this task.

Six rows landed already correctly re-anchored by earlier tasks in this change (old budget equals
new budget above): `build-green.md`, `jira-followups.md`, `jira-integration-rationale.md`,
`jira-integration.md`, `plan-provenance.md`, `project-configuration-rationale.md`,
`workspace-isolation-rationale.md`, `openspec-explore/SKILL.md`. The remaining rows were stale —
either left over from before this change's evictions shrank the file (`pipeline.md`,
`project-configuration.md`, `workspace-isolation.md`, `state-file.md`, `myflow-status/SKILL.md`),
or off by a small amount from an earlier task's arithmetic (`finish-contract.md`,
`handoff-blocks.md`, `pipeline-rationale.md`, `myflow-do/SKILL-rationale.md`,
`myflow-do/SKILL.md`, `myflow-finish/SKILL-rationale.md`, `myflow-finish/SKILL.md`,
`myflow-start/SKILL-rationale.md`, `myflow-start/SKILL.md`, `skills/myflow-contracts/SKILL.md`) —
all corrected to `floor(actual × 1.25)` in this task's edit.

---

## task-2-ledger.md

# Task 2 — per-move ledger

| Removed passage | Source heading | Destination | Pointer left |
|---|---|---|---|
| The block a state hands off is defined | `### The block each state renders` (under `## Handoff output`) | `skills/myflow-contracts/handoff-blocks.md` § The block each state renders | "The block a state hands off is defined in **The block each state renders** (`skills/myflow-contracts/handoff-blocks.md`), which is canonical for the three per-state templates, the run-only rule and the rendering-selection table. `/myflow-status` loads it; a producing command carries only the block it prints." |

---

## task-3-ledger.md

# Task 3 — per-move ledger

ledger: no passages removed

This task reshapes lines in place (folds related fields onto one line) and removes no prose. No
row is required by the per-move ledger shape in **A move or eviction is recorded in a per-move
ledger** — nothing here is a removed passage, a rule extraction, or an eviction.

## Value-preservation check (required deliverable)

### `STARTED` block — `skills/myflow-contracts/handoff-blocks.md` and `skills/myflow-start/SKILL.md`

Pre-fold values (4 separate lines → 1 folded `Recorded` line):
- Decisions recorded: count, or "none"
- Open questions: count, or "none"
- Planning effort: level in force, or "not recorded — planned at default" (start's own copy also
  enumerates "reused from the creating run")
- Models: implementation model or "not recorded", review panel model or "not recorded", panel
  fixes model or "not recorded"

Post-fold values (1 folded `Recorded` line):
- decisions count (canonical template: `<N>`, backed by the "reads `none` when zero" rule restated
  in the paragraph directly under the template; start's own copy enumerates `<N> decisions | none`
  inline)
- open questions count (same treatment: `<N>` in the canonical template, `none`-when-zero rule
  restated in prose; start's own copy enumerates `<N> open questions | none` inline)
- planning effort: level in force, or "not recorded — planned at default" (start's own copy keeps
  its "reused from the creating run" enumeration)
- implementation model, or "not recorded"
- review panel model, or "not recorded"
- panel fixes model, or "not recorded"

Diff: identical sets. Nothing added, nothing removed. The two counts' "or none" wording moved from
the inline placeholder into the explanatory paragraph immediately below the template in
`handoff-blocks.md` (a layout change, called out explicitly in that paragraph); `skills/myflow-start/SKILL.md`'s
own copy keeps the "N | none" text inline, which stays inside the value space the canonical
placeholder describes. `Jira` and `Jira description (pre-edit)` were not folded (both run-only,
mixing them into the on-disk `Recorded` line is exactly what the new folded-line rule forbids) and
are unchanged.

### `IN_PROGRESS` (after `/myflow-do`) block — `skills/myflow-contracts/handoff-blocks.md` and `skills/myflow-do/SKILL.md`

Pre-fold values (2 separate lines → 1 folded `Staged` line):
- Progress: completed/total tasks
- Git: staged and uncommitted | committed and pushed to the PR branch | committed and pushed with
  no PR (canonical has all three; `/myflow-do`'s own copy enumerates only the first two, which it
  is the only command able to emit — the third is `/myflow-status`-only, per the existing pairing
  table)

Post-fold values (1 folded `Staged` line):
- completed/total tasks
- git state: staged and uncommitted | committed and pushed to the PR branch | committed and pushed
  with no PR (canonical); staged and uncommitted | committed as two commits and pushed to the PR
  branch (`/myflow-do`'s own copy, unchanged enumeration)

Diff: identical sets. `Panel` (run-only) and `Jira description (pre-edit)` (run-only) were not
folded into `Staged` (on-disk) — same folded-line rule. The Git-state/review-command pairing table
and its prose are unchanged in content; only "`Git` line" was reworded to "`Staged` line's git
state" / "git state" throughout, since the field is no longer a standalone line.

### Blocks not touched

`IN_PROGRESS` (after `/myflow-finish` run 1) and the run-2 terminal block, both carried in
`skills/myflow-finish/SKILL.md`, are unaffected: the brief's Step 1 and Step 2 fold only the
`STARTED` and post-`/myflow-do` templates, and neither of the finish blocks has two adjacent
same-kind fields the fold rule would apply to (`Change`/`Route`/`PR`/`Outstanding` alternate
on-disk and run-only kind field by field, so no two adjacent fields share a kind to fold). No edit
was made to `skills/myflow-finish/SKILL.md`'s block content.

---

## task-4-ledger.md

# Task 4 — per-move ledger

Every passage below was removed from `skills/myflow-status/SKILL.md`, per **The status table reads
the state file and local git, never the network**
(`openspec/changes/kan-95-slim-the-myflow-contract-files/specs/myflow-handoff-output/spec.md`). None
of it is rewritten elsewhere in the loaded corpus: the capability it described (a `gh pr list`
network probe, and the worktree/branch table column it partly fed) is removed by this change, not
relocated.

| Removed passage | Source heading | Destination | Pointer left |
|---|---|---|---|
| `5. PR — \`gh pr list --head <branch> --state all --json number,state,url\`, **only when \`gh\` is installed and the remote is a GitHub host**. Otherwise (Bitbucket, no \`gh\`, no network, branch never pushed) PR state is **unknown** — report it as unknown and treat it as inconclusive, never as "no PR exists".` | 2. Resolve each change's state | — (deleted, capability removed) | none |
| `Worktree / branch` table column header, its separator cell, and its two example cells (`` `/abs/path/openspec-kan-8-myflow-updates` @ `openspec/kan-8-myflow-updates` `` and `none`) | 3. Render the table | — (deleted, capability removed) | Replacement sentence: "The absolute worktree path is given in the detail view, taken from the `worktrees` keys." |
| `Worktree paths are **absolute**, taken from the \`worktrees\` keys. The **PR** column shows the number and state when known, \`—\` when \`prUrl\` is null, and \`?\` when PR state could not be determined.` | 3. Render the table | — (deleted, capability removed) | Replacement sentence: the new PR-column paragraph stating the number is parsed from `prUrl`, `—` when null, and that open/merged/closed is never reported because it needs a network call this command no longer makes |
| `#42 open` (PR cell value in the rendered-table example, replaced by `#42`) | 3. Render the table | — (deleted, capability removed) | none — the bare `#42` in the same example row |
| `PR number, state, and URL when one exists` | 4. Detail view | — (deleted, capability removed) | Replacement sentence: "PR number and URL when one exists — not whether it is open, merged or closed, which this report does not track; check the forge for that" |
| `Report \`gh\` being unavailable, or a non-GitHub forge, as "PR state unknown" rather than guessing — and never clear \`prUrl\` on that unknown.` (Guardrails) | Guardrails | — (deleted, capability removed) | Replacement sentence: "**Never** call `gh`, or any other network command, to determine PR state — the PR column reports only the number parsed from the recorded `prUrl`." |
| `Bash(gh:*)` from the `allowed-tools` frontmatter key | frontmatter | — (deleted, capability removed) | none — the tool grant is dropped outright, since no step in the file invokes `gh` any longer |
| `worktree, ` from the frontmatter `description`'s field list (fix round, review Finding 1) | frontmatter | — (deleted, capability removed) | none — the worktree path is still surfaced, but only in the detail view, which the no-argument-table-facing `description` no longer claims |

## The two paragraphs this task edits without removing (per the brief's, and the reviewer's, explicit instruction)

The self-heal bullet "Do apply the one permitted correction" (`### 2. Resolve each change's state`,
under "Self-heal is monotonic, and writes almost nothing but `state`") is **not** in the table above
because it was not deleted. Its evidentiary claim — "this command's own `gh pr list` probe (step 2,
item 5) conclusively answers..." — no longer holds once item 5 is gone, so the paragraph was reworded
in place to state that the permitted correction's precondition can no longer be met and it therefore
never fires from this command, while leaving the correction's definition itself untouched. The
paragraph's actual deletion is explicitly reserved to Task 5 (which removes state self-heal from this
skill entirely), per the brief: "Do not delete the correction — Task 5 owns that text." This is a
content edit, not a move or eviction, so it carries no row in the table above.

**Fix round (review Finding 2).** The Guardrails bullet "Never rewind a state, and never fabricate a
`prUrl`. Clearing a `prUrl`... is the one permitted correction..." (`## Guardrails`) carried the same
dead precondition as the bullet above but had not been caveated on the first pass. It is likewise
**not** in the table: nothing was deleted from it, a one-sentence caveat was appended pointing back at
the step-2 explanation above ("do not restate it here") rather than duplicating that explanation a
second time, per Single Source of Truth. Its removal, like the step-2 bullet's, is Task 5's.

---

## task-5-ledger.md

# Task 5 ledger — Delete state self-heal

Every passage that lived in `skills/myflow-contracts/state-self-heal.md` (15,250 bytes, 189 lines),
one row per removed unit, in the order it appeared in the file. `Destination` is
`state-file.md § <heading>` for a rule re-homed into the surviving contract, or
`— (deleted with the mechanism)` for one that goes away with self-heal itself, per the dispatch's
own convention for this task.

## Part A — passages removed from `state-self-heal.md`

| Removed passage | Source heading | Destination | Pointer left |
|---|---|---|---|
| "Artifacts are the source of truth; the state" (intro, contradiction table, apply-worktree-first rule) | State self-heal (document intro) | — (deleted with the mechanism) | "A state file that is missing or unparseable is named in this command's own output, and the change is omitted from the table." (`skills/myflow-status/SKILL.md` § 2. Resolve each change's state) |
| "The one permitted correction. Monotonicity forbids writing" | State self-heal | — (deleted with the mechanism) | none — the correction itself is deleted outright (per dispatcher's explicit instruction), not replaced |
| "A check that cannot be performed is not" | State self-heal | — (deleted with the mechanism) | none |
| "When the file is missing, unparseable, or contradicted" (rewrite + `⚠` announcement) | State self-heal | — (deleted with the mechanism) | none — superseded by the pre-existing "reported and skipped" rule under **Change name resolution** (`skills/myflow-contracts/pipeline.md`), per `myflow-state-integrity`'s Migration note; not authored by this task |
| "Self-heal obeys the monotonicity rule stated under" | State self-heal | — (deleted with the mechanism) | none |
| "A rewrite carries forward every field it did" (general self-heal write duty) | State self-heal | — (deleted with the mechanism) | "Because writes render the whole object, every command must first **read the existing file and carry forward** every field it does not itself own..." (`state-file.md`, pre-existing, verified independent — untouched by this task) |
| "For `planningEffort`, \"exactly as read\" is not a" | State self-heal | — (deleted with the mechanism) | "Carrying the planning effort forward is what performs the rewrite..." (`state-file.md`, pre-existing, verified independent — untouched by this task) |
| "When the prior file is missing or unparseable" (unrecovered-field announcement + template) | State self-heal | — (deleted with the mechanism) | none |
| "That example is exhaustive on purpose, and it" | State self-heal | — (deleted with the mechanism) | none |
| "`worktrees` is the one field a rebuild recovers" | State self-heal | — (deleted with the mechanism) | none |
| "The scan reaches one repository, and for a" (multi-repo rebuild limit) | State self-heal | — (deleted with the mechanism) | none |
| "That is a limit, not a regression, and" | State self-heal | — (deleted with the mechanism) | none |
| "What the scan cannot recover is the merge" (null merge base) | State self-heal | — (deleted with the mechanism) | "A `null` value is legal and means *no merge base recorded* for that path — it can occur in a hand-edited or out-of-band-modified file." (`state-file.md` § `worktrees` field bullet, edited in place — self-heal citation removed, rule kept) |
| "JSON that parses but is missing one or" (parseable-but-incomplete = unparseable) | State self-heal | `state-file.md` § unparseable definition (new paragraph after the JSON schema block) | "JSON that parses but is missing one or more of the fields this contract requires is unparseable in full on that account alone, not partially recovered." |
| "A file that read successfully but was merely" (contradiction carry-forward) | State self-heal | — (deleted with the mechanism) | none |
| "There is no legacy-value migration for the retired" (`stage` field) | State self-heal | — (deleted with the mechanism) | none — the general closed-schema rule already makes any undocumented key (including `stage`) unparseable; the `stage`-specific callout added no rule a surviving file depends on (checked: no citation to it anywhere in `skills/`, `rules/`, `commands*/`; only the pre-existing, out-of-scope baseline `openspec/specs/myflow-state-machine/spec.md` mentions `stage` by name) |
| "That claim is scoped to `stage`, and the" (stage vs. effort contrast) | State self-heal | — (deleted with the mechanism) | none |
| "The schema is closed in both directions: just" (the closed-schema rule itself) | State self-heal | `state-file.md` § unparseable definition (new paragraph after the JSON schema block) | "A state file is unparseable when it is not valid JSON, omits a documented field other than `planningEffort` or `models`, or carries an undocumented one." |
| "Two documented exceptions exist to the missing-field half" (`planningEffort`/`models` absent-key exception) | State self-heal | `state-file.md` § `models` field bullet (pre-existing paragraph, self-heal citation repointed) | "A state file that omits `planningEffort` or `models` entirely is valid, and each absent key is read as *not recorded*. This is a deliberate exception to the closed-schema rule stated above..." |
| "One documented exception exists to the undocumented-key half" (retired `effort` key) | State self-heal | — (deleted with the mechanism; duplicate) | "A file carrying the retired `effort` key is read as recording the equivalent level..." (`state-file.md`, pre-existing, verified independent — untouched by this task) |
| "The exception is unconditional on the value, and" | State self-heal | — (deleted with the mechanism; duplicate) | "A value outside those three reads as *not recorded*, and never makes the file unparseable." (`state-file.md`, pre-existing, verified independent — untouched by this task) |
| "Exemption from *being named among the unrecovered fields*" | State self-heal | — (deleted with the mechanism) | none |
| "A recorded level lost to a rebuild is" | State self-heal | — (deleted with the mechanism) | none |
| "A state file whose `prUrl` is `null` while" (## The stale-`prUrl` gap — recorded, not closed, whole section) | The stale-`prUrl` gap — recorded, not closed | — (deleted with the mechanism) | none — the general cost ("never corrected by anything, and never flagged") is stated in `openspec/changes/kan-95-slim-the-myflow-contract-files/specs/myflow-state-integrity/spec.md`'s Migration note, an OpenSpec authority this task does not edit |

24 rows, matching the file's 24 distinct argumentative units end to end — nothing in the source file
is unaccounted for.

## Part B — mechanical citation repairs in surviving files (no rule content lost)

These are deletions the diff also contains, outside `state-self-heal.md` itself: every one removes a
*citation* to the deleted file or its concepts (a table row, an index entry, a budget row, a word in
a sentence), never a rule. Listed for the "every deletion has a row" check, kept terse because none
of them carries an argument to lose.

| File | What was removed | Why it is safe |
|---|---|---|
| `skills/myflow-status/SKILL.md` | intro's "one exception is state self-heal" clause; "Follow all three contracts" → "both"; items 1-3 of step 2 + their intro sentence; the self-heal rewrite paragraph + four monotonic bullets (replaced with the reporting rule); the `⚠` legend line; the `planningEffort`/`models` "never a `⚠`" clauses (surfacing sentences kept); `⚠` mention in the retired-value paragraph; "item 4" renumbered to unnamed "Merge status" + cross-references retargeted; the `⚠`-correction detail-view bullet; three Guardrails bullets collapsed to two plain read-only bullets | Command no longer performs self-heal at all (Task 4 already removed its evidence gathering); nothing here stated a rule another file depends on — verified via full end-to-end re-read of the file after editing |
| `skills/myflow-contracts/pipeline.md` | `## State self-heal` section (heading + two sentences) | Purely an index/pointer section for a file that no longer exists |
| `skills/myflow-contracts/pipeline.md` | "Resolving a change's worktrees" reason sentence rewritten | Not a deletion of a rule — the section and its mechanism are unchanged; only the *reason it lives here* changed, per this task's explicit instruction |
| `skills/myflow-contracts/finish-contract.md` | Same reason sentence, same rewrite | Same as above |
| `skills/myflow-contracts/pipeline-rationale.md` | Empty `## State self-heal` heading stub; "and self-heal may clear one that was real" clause | Stub carried no rationale body; the clause was a parenthetical aside, not a rule — the sentence's own point (the `prUrl` test is one-way) survives unchanged |
| `skills/myflow-contracts/SKILL.md` | `state-self-heal.md` word from frontmatter description; its index row | File deleted; index must not name a file that is not present (this directory's own requirement) |
| `rules/myflow-manual-review.mdc` | `State self-heal` row from the contract table | File deleted, so the table entry pointed nowhere |
| `CLAUDE.md`, `AGENTS.md` | "self-heal, " word from the `skills/myflow-contracts/` skill-index description | Same |
| `skills/README.md` | `` `state-self-heal.md`, `` from the file list | Same |
| `scripts/check-contract-budget.sh` | `skills/myflow-contracts/state-self-heal.md 19062` budget row | File deleted; a budget row for a nonexistent file would itself violate the guard's own invariant |
| `scripts/test-check-contract-budget.sh` | nothing — grepped for a fixture case naming the path; none existed | n/a |
| `commands/myflow-status.md`, `commands-claude/myflow-status.md` | "Its one write is state self-heal: correcting a stale cache..." clause | Command is now entirely read-only; found via a repo-wide case-insensitive sweep after the brief's named-files list, not itself named in the brief |
| `skills/myflow-start/SKILL-rationale.md` | "a file self-heal rebuilt from artifacts loses it, and self-heal names it among the unrecovered fields when it does" clause | The underlying fact (`artifactUrl` may be `null`) holds independent of self-heal — the field is simply nullable; found via the same sweep |

## Part C — fix round: `Resolving a change's worktrees` moved to `finish-contract.md`

Review found my Part-B reason rewrite ("`/myflow-do` and `/myflow-finish` both need it") false:
`/myflow-do` never cites the section and never uses the `git worktree list --porcelain` scan — it
creates its worktree via `superpowers:using-git-worktrees` and reads paths from the state file's
`worktrees` map, as does `/myflow-status`. With self-heal deleted, `/myflow-finish` is the section's
**only** consumer, and it already loads `finish-contract.md`. Per **A section reachable from only one
command lives in its own file** (`openspec/specs/myflow-contract-economy/spec.md`), the fix is a
verbatim move — rules and reasoning together, no core/appendix split — not another reworded reason.

| Removed passage | Source heading | Destination | Pointer left |
|---|---|---|---|
| "The scan that finds the worktrees carrying a change's branch." (description sentence only — kept, moved verbatim) | Resolving a change's worktrees (`pipeline.md`) | `finish-contract.md` § Resolving a change's worktrees | "The scan that finds the worktrees carrying a change's branch." (verbatim, same file, no reason sentence follows) |
| "Both `/myflow-do` and `/myflow-finish` need it, and" (the false placement-reason sentence, my own Part-B rewrite) | Resolving a change's worktrees (`pipeline.md`) | — (deleted with the mechanism's fix; false, and per the coordinator's explicit instruction no replacement reason is invented — the section needs no placement justification once it sits in the contract of its one consumer) | none |
| "The set of worktrees is the **keys of" (worktrees-map / absent-map scan bash snippet) | Resolving a change's worktrees (`pipeline.md`) | `finish-contract.md` § Resolving a change's worktrees | verbatim, unchanged (checked against `scripts/check-cleanup-complete.sh`'s own `awk`/`substr($0, 10)` usage — still agrees) |
| "**Never guess a path.** Worktree layout differs" | Resolving a change's worktrees (`pipeline.md`) | `finish-contract.md` § Resolving a change's worktrees | verbatim, unchanged |
| "**The path is taken with `substr`, never `$2`.**" | Resolving a change's worktrees (`pipeline.md`) | `finish-contract.md` § Resolving a change's worktrees | verbatim, unchanged |
| "Which worktrees those are, and the `git worktree" (Worktree cleanup's citation sentence, my own Part-B rewrite) | Worktree cleanup (`finish-contract.md`) | `finish-contract.md` § Worktree cleanup (edited in place — cross-file citation with a reason clause repointed to a same-file reference, reason clause dropped since the two sections now share a file) | "...are **Resolving a change's worktrees** above." |
| index row prose "and resolving a change's worktrees" | `skills/myflow-contracts/SKILL.md` § Index (pipeline.md row) | `skills/myflow-contracts/SKILL.md` § Index (finish-contract.md row, worded "resolving a change's worktrees, and worktree cleanup") | moved to the row naming the file that now holds the section |

Two permitted edits only, per **The split is a verbatim partition, with citation repointing as its
only edit**: (1) the `Worktree cleanup` citation's bold token + path repointed from
`**Resolving a change's worktrees** (`skills/myflow-contracts/pipeline.md`)` to
`**Resolving a change's worktrees** above` — a same-file reference needs no backticked path; (2) no
stale position word existed in the moved passage itself to delete. The false reason sentence was not
"repointed" — it was deleted outright, per the coordinator's explicit instruction not to invent a
replacement, which the verbatim-partition rule does not forbid: that sentence was never part of the
section's rules or its reasoning about worktree resolution — it was reasoning about the section's
*file placement*, which stops being needed once the section sits in the file of its only reader.

Re-verified after the move: `scripts/check-references.sh`, `scripts/check-vocabulary.sh`,
`scripts/check-contract-budget.sh`, `scripts/test-check-contract-budget.sh`, `scripts/test-setup.sh`
all exit 0; `grep -rn "Resolving a change's worktrees" . --exclude-dir=.git --exclude-dir=docs
--exclude-dir=archive --exclude-dir=.superpowers` shows only `finish-contract.md` (heading + one
same-file reference) outside `openspec/`; `scripts/check-cleanup-complete.sh`'s own
`worktree list --porcelain` / `substr($0, 10)` usage still agrees with the moved snippet (unchanged
by the move — verified by reading both side by side).

## Note on the `stage` field (row 16 above)

I checked whether any surviving file depends on self-heal's specific claim that a legacy `stage`
field has no migration table (unlike `effort`). Grepped `skills/`, `rules/`, `commands*/`,
`CLAUDE.md`, `AGENTS.md` and `openspec/specs/` for `` `stage` `` / `"stage"` / `stage field`. The
only hit outside `state-self-heal.md` is `openspec/specs/myflow-state-machine/spec.md:210`, a
pre-existing baseline spec this change's own delta specs do not touch and which this task is
forbidden from editing (`openspec/` is off-limits). Nothing in the corpus this task owns cites the
`stage` paragraph, and the general closed-schema rule already implies the same outcome for `stage`
as for any other undocumented key, so no re-home was needed.

---

## task-6-ledger.md

# Task 6 — per-move ledger

Every passage evicted from `skills/myflow-contracts/pipeline.md`'s remaining sections (States,
Command surface, State transitions, Wrong state for this command, Git boundaries, Preserving the
session records, Progress visibility, Handoff output, IntelliJ commands, Temporary artifacts
registry, State file, Project configuration, Jira integration, Model policy, Change name
resolution). `Resolving a change's worktrees` no longer exists in `pipeline.md` — Task 5 moved it
to `finish-contract.md` — so it is out of scope here, confirmed by re-reading Task 5's own ledger
and report before starting.

Sections classified **wholly core, unchanged** (no row below, because no prose was removed): States,
Command surface, State transitions, Wrong state for this command, IntelliJ commands, Finish contract,
State file, Project configuration, Jira integration. Each was read paragraph-by-paragraph against the
core/rationale test; none carries a paragraph that is pure argument, history, or alternatives
considered separable from its rule without gutting the rule. Their mirrored appendix headings in
`pipeline-rationale.md` are present with no new body from this task (several already carried
unrelated pre-existing content from earlier tasks, left untouched).

`Destination` is `pipeline-rationale.md § <heading>` for a wholly-rationale move, or `— (rule
extracted)` with the new core sentence quoted for a mixed passage. `Pointer left` names the sentence
now standing in the core, or `none`.

| Removed passage | Source heading | Destination | Pointer left |
|---|---|---|---|
| "Both commits are guarded, and an empty one is skipped rather than failed. `git commit` exits non-zero..." | Git boundaries | — (rule extracted) | "**Both commits are guarded, and an empty one is skipped rather than failed.** Each commit is preceded by a staged-changes test, and the whole sequence is one `&&` chain, run as a single command. See **Git boundaries** (`skills/myflow-contracts/pipeline-rationale.md`) for the ordinary cases this guards against and why it is a chain rather than `set -e`." |
| "A skipped commit is reported, and a FAILED commit stops the sequence. Those are different outcomes..." | Git boundaries | — (rule extracted) | "**A skipped commit is reported, and a FAILED commit — one a hook rejects — is a git failure: report git's own output and stop.** See **Git boundaries** (`skills/myflow-contracts/pipeline-rationale.md`) for what an unguarded sequence would do instead." |
| "A non-zero exit is never silent and never a stop. Those two rules pull in opposite directions..." | Preserving the session records | — (rule extracted) | "**A non-zero exit is never silent and never a stop; the remaining sources are still attempted after any one failure, and the handoff names which records were preserved and which were not.** See **Preserving the session records** (`skills/myflow-contracts/pipeline-rationale.md`) for why." |
| "That ordering is deliberate, and deliberately differs from `/myflow-do`'s. Here the preservation call..." | Preserving the session records | — (rule extracted) | "**Do not harmonise the two orderings for symmetry** — here the preservation call runs before staging; in `/myflow-do` it runs after. See **Preserving the session records** (`skills/myflow-contracts/pipeline-rationale.md`) for why the asymmetry is what keeps preserved records out of the staged-only path." |
| "No third checkbox marker is added to `tasks.md` to carry an in-progress state. A marker written..." | Progress visibility | — (rule extracted) | "**No third checkbox marker is added to `tasks.md`** to carry an in-progress state; the in-progress count comes from the harness's task list alone, which no run persists. See **Progress visibility** (`skills/myflow-contracts/pipeline-rationale.md`) for why a marker would be unsafe." |
| "Stated against the mechanism, never against one harness's tool. myflow runs in Claude Code, Cursor..." | Progress visibility | — (rule extracted) | "**Stated against the mechanism, never against one harness's tool.** Where a harness offers no task-list mechanism, the command prints the equivalent block instead — a count line naming how many steps are done, in progress and open, followed by one line per step marked done or not done — and no harness has to gain a task tool to satisfy the rule. See **Progress visibility** (`skills/myflow-contracts/pipeline-rationale.md`) for why the rule is stated against the mechanism." |
| "The next command is the last line — bare, copy-pasteable, with no prose after it. An agent..." | Handoff output | — (rule extracted) | "- **The next command is the last line** — bare, copy-pasteable, with no prose after it. See **Handoff output** (`skills/myflow-contracts/pipeline-rationale.md`) for why." |
| "`/myflow-do` never stages `openspec/`, `docs/manual-test/` or `docs/superpowers/` before finish...." | Handoff output | — (rule extracted) | "- **`/myflow-do` never stages `openspec/`, `docs/manual-test/` or `docs/superpowers/` before finish**, and the list is fixed here rather than configured per project. `/myflow-finish` run 1 stages them and commits them separately from the implementation, so nothing is lost. See **Handoff output** (`skills/myflow-contracts/pipeline-rationale.md`) for why leaving them unstaged — rather than filtering a display — is what keeps them out of every view of the staging area:" |
| "They sit at the start of the run rather than in the block above because labelling a tab..." | The tab commands, printed at the start of a run | — (rule extracted) | "**They sit at the start of the run, not in the handoff block, and the colour is one fixed value — `cyan` — for every command and change, signifying only that a pipeline command owns the tab.** `/myflow-status` prints neither line: a read-only report does not own the tab. See **The tab commands, printed at the start of a run** (`skills/myflow-contracts/pipeline-rationale.md`) for why cyan was chosen and why the lines are printed at the start rather than the end." |
| "This section is canonical for the model roles, their defaults and how an override applies. One..." | Model policy | — (rule extracted) | "**This section is canonical for the model roles, their defaults and how an override applies** — the one file every `/myflow-*` command loads for them. See **Model policy** (`skills/myflow-contracts/pipeline-rationale.md`) for why this is the one place the rule lives." |
| "Which file to change first. The normative requirements behind this section belong to the OpenSpec..." | Model policy | — (rule extracted) | "**Change the capability first and bring this section with it: a section that contradicts the OpenSpec requirement is this file's defect, not the spec's.** See **Model policy** (`skills/myflow-contracts/pipeline-rationale.md`) for which file governs which layer and why runtime reads this section rather than the live spec." |
| "Implementer subagents dispatched by `/myflow-do` run on Opus (or the harness's strongest available..." | Model policy | — (rule extracted) | "**Implementer subagents dispatched by `/myflow-do` run on Opus** (or the harness's strongest available model), which **explicitly overrides** superpowers:subagent-driven-development's model guidance. See **Model policy** (`skills/myflow-contracts/pipeline-rationale.md`) for why that guidance's cost savings do not apply here." |
| "Two further instructions in that same upstream skill are also overridden, and are named here..." | Model policy | — (rule extracted) | "**Two further instructions in that same upstream skill are also overridden: dispatching the final review on the most capable model, and escalating the model in fix rounds 4-5.** myflow fixes every panel slot at the panel's model instead and escalates breadth (the conditional Security, Adversarial and extra-principle-lens slots) rather than the model. See **Model policy** (`skills/myflow-contracts/pipeline-rationale.md`) for the reasoning." |
| "The panel-fix default is the strongest available model, and deliberately not Sonnet. The role..." | Model policy | — (rule extracted) | "**The panel-fix default is the strongest available model, and deliberately not Sonnet** — the role applies fixes, which is implementer work, so the implementer rule above governs it too. See **Model policy** (`skills/myflow-contracts/pipeline-rationale.md`) for why." |
| "Every subagent dispatch records the model it used in the SDD ledger, alongside the task it ran...." | Model policy | — (rule extracted) | "**Every subagent dispatch records the model it used** in the SDD ledger, alongside the task it ran. See **Model policy** (`skills/myflow-contracts/pipeline-rationale.md`) for why, and for the history behind this rule." |
| "This record outlives the change. The ledger is authored under `.superpowers/`, which is gitignored..." | Model policy | — (rule extracted) | "**This record outlives the change: the ledger is preserved under `docs/superpowers/ledgers/` at run 1, before the worktree carrying `.superpowers/sdd/` is removed.** See **Model policy** (`skills/myflow-contracts/pipeline-rationale.md`) for why, and **Run 1 — the branch is not merged** (`skills/myflow-contracts/finish-contract.md`) for the preservation duty itself." |
| "Durability is a stronger reason to leave an unobserved entry unobserved, not a weaker one. A..." | Model policy | — (rule extracted); permitted edit 2 applied (stale `above` deleted — see Notes) | "**A persisting record must not fill in `unknown (agent-defined)` on the way into the repository** — no preservation step invents a model slug. See **Model policy** (`skills/myflow-contracts/pipeline-rationale.md`) for why." |
| "This table is the one place a cleanup rule is stated. Everything else that mentions a removal..." | Temporary artifacts registry | — (rule extracted) | "**This table is the one place a cleanup rule is stated.** Everything else that mentions a removal points here rather than restating it. **Worktree cleanup** (`skills/myflow-contracts/finish-contract.md`) is the *procedure* for the rows removed there, not a second statement of the rule. See **Temporary artifacts registry** (`skills/myflow-contracts/pipeline-rationale.md`) for why a stale second copy would be dangerous." |
| "An artifact no row accounts for is a defect in the registry, corrected by adding the row...." | Temporary artifacts registry | — (rule extracted) | "**An artifact no row accounts for is a defect in the registry**, corrected by adding the row — never left unaccounted for on the grounds that something probably removes it. See **Temporary artifacts registry** (`skills/myflow-contracts/pipeline-rationale.md`) for the incident that established this." |
| "This is the one row whose removal is verified by asking rather than by looking, and the..." | Temporary artifacts registry | — (rule extracted) | "**This is the one row whose removal is verified by asking rather than by looking: a survivor is established from the project's own survivor report, never inferred from the removal's exit code** — stated once under **Creation and cleanup** (`skills/myflow-contracts/workspace-isolation.md`), with the report's output and exit-code contract under **Project configuration** (`skills/myflow-contracts/project-configuration.md`). A report that could not reach its service is skipped rather than failed. See **Temporary artifacts registry** (`skills/myflow-contracts/pipeline-rationale.md`) for why asking, not looking, is required here." |
| "Nothing removes the claimed cache index, and nothing in this pipeline can — which is why..." | Temporary artifacts registry | — (rule extracted) | "**Nothing removes the claimed cache index, and nothing in this pipeline can.** It is not written into the state file, and the project's `remove` command does not touch it either — stated as a property of the `cache index` resource word under **Project configuration** (`skills/myflow-contracts/project-configuration.md`). See **Temporary artifacts registry** (`skills/myflow-contracts/pipeline-rationale.md`) for why: guessing an index to sweep risks flushing another workspace's." |
| "This is why the second source matters: `openspec list --json` only sees change directories present..." | Change name resolution | `pipeline-rationale.md § Change name resolution (all `/myflow-*` commands)` | "See **Change name resolution (all `/myflow-*` commands)** (`skills/myflow-contracts/pipeline-rationale.md`) for why the second source (the state directory) is needed alongside `openspec list --json`." |

22 rows. Every destination heading listed above exists in `pipeline-rationale.md` (verified by
`scripts/check-references.sh`, which resolves every bold-token citation added in this task's diff —
all pass).

## Notes

**The Change name resolution row is not a "rule extracted" row like the others.** Unlike the other
21, the two flanking rules in that paragraph ("drop any name whose directory reached archive" and
"the state directory is per-project, so a change's state file is reachable regardless of worktree")
were **already** stated in the paragraph and are left untouched, byte-for-byte, in the core. Only the
middle sentence — pure justification, no rule of its own — moved out. The new core text is a bare
pointer, not a rule-stating sentence, so its ledger entry is filed as a plain wholly-rationale move
with a pointer rather than "(rule extracted)".

**One stale position word was found and deleted, not substituted**, per the sole other permitted
edit: the Model policy row for "Durability is a stronger reason..." originally closed with `stays
exactly as written above and no step fills it in`. `above` referred to two mentions of `unknown
(agent-defined)` earlier in the ORIGINAL `pipeline.md` Model policy section — both of which stayed
in the core, not moved — so in the appendix `above` pointed at nothing. Deleted the word only;
nothing was substituted for it. Also deleted a now-false `below` in the "Which file to change
first" passage (Model policy): it originally said the cited requirement "anchors the defaults
below," where "the defaults" (the Opus/Sonnet table) stayed in the core. Both fixes are the
`above`/`below`-deletion edit **The split is a verbatim partition, with citation repointing as its
only edit** permits.

**A pre-existing "one paragraph above" reference was checked and left alone.** In the moved
Temporary artifacts registry passage "Nothing removes the claimed cache index...", the sentence "The
rule one paragraph above is that an artifact no row accounts for is a defect in the registry" was
already loose/rhetorical in the *original* `pipeline.md` — three other paragraphs actually separated
the two in the source file, not one. The move (which drops two of those three intervening
core-and-untouched paragraphs from the appendix) does not make the reference any less accurate than
it already was; if anything the two passages are now closer together. Left unchanged — not a case
the verbatim-partition rule's `above`/`below` clause covers, since the phrase was never a literal,
move-broken position claim to begin with.

**The two "Wrong state for this command" headings pre-date this change and are not a duplicate
section.** `pipeline.md` has one real `## Wrong state for this command` heading (with its own rule
prose) followed by a fenced `text`-less code block that *itself* contains the literal line `##
Wrong state for this command` — that line is the literal handoff-block template the section
documents, not a second section. `grep -n "^#"` matches both because grep is not fence-aware; a
markdown renderer or `scripts/check-references.sh`'s own `strip_fenced_lines` helper is. No merge
performed — there is nothing to merge, and deleting either would either delete the real section or
corrupt the example template it exists to show. Reported per the brief's instruction rather than
silently resolved.

---

## task-7-ledger.md

# Task 7 — per-move ledger

Every passage evicted from `skills/myflow-do/SKILL.md` into its sibling
`skills/myflow-do/SKILL-rationale.md`. All ten rows below are rule extractions: the mixed passage's
argument/history half was removed from the core, the original full passage (rule sentence(s) plus
argument, byte-for-byte) now lives in the rationale file under the mirrored heading, and the core
keeps its rule sentence(s) — several of them unchanged verbatim — plus one new pointer sentence
naming the appendix heading. No wholly-rationale (argument-only) paragraph was found standing free
of a rule in this file; every candidate was a rule with an argument welded onto it in the same
sentence or paragraph, consistent with the brief's warning that this file is unusually rule-dense.

Sections read paragraph-by-paragraph and found **wholly core, unchanged** (no row below): the
frontmatter/intro block, `## State gate`, `## Superpowers Basic Workflow`, `## 1. Load context and
validate the plan`, `## 3. Documenting a fix, before implementing it`, the four required implementer
dispatch blocks and the per-task-review paragraph in `## 4. Execute (SDD + TDD)`, the slot roster
table, the `[STANDARDS_PATHS]` security paragraph, the withdrawn-marker-reason paragraph, and the
optional-slot trigger table and panel-re-run tables/paragraphs in `## 5. The review panel`, the
behaviour-checklist rules, the absolute-path/URL-resolution bullet, the apps-in-scope and
no-runnable-application bullets, and the machine-read-shapes closing paragraph in `## 6. Write the
manual test guide`, and the guard-invocation table, the script-not-found paragraph, the `When.` and
cache-index and dropped-row bullets, the staging commands and their exclusion-pathspec note, the
one-commit-exception procedure, the state-file-write paragraph, and the handoff block and its two
trailing notes in `## 7. Verify, stage, and hand off`, plus all of `## Guardrails`. Each of these was
checked against the core/rationale test and carries no paragraph that is pure argument, history, or
alternatives separable from its rule without gutting the rule — several are the exact content the
brief named as must-stay (slot roster, trigger lists, marker-line rules, dispatch blocks, staging
pathspec, exit-code table, handback prompt). Their mirrored appendix headings are present with no
new body from this task where nothing was extracted from them.

`Destination` is `— (rule extracted)` for every row (no plain wholly-rationale move occurred).
`Pointer left` quotes the sentence(s) now standing in the core in place of the removed argument.

| Removed passage | Source heading | Destination | Pointer left |
|---|---|---|---|
| "the derivation is deterministic, so a later run reproduces the same id rather than reading one back, which is why nothing about it is written to the state file. Two later steps consume that one value — section 6 writes the guide's URLs from it, and section 7 resolves the project's declared isolation rows against it — so an id derived twice in one run is two chances to disagree." | 2. Isolate the workspace (first run only) | — (rule extracted) | "**Then compute this worktree's workspace id from the change name.** The derivation is stated once under **The workspace id** (`skills/myflow-contracts/workspace-isolation.md`), which is canonical for it — do not restate it here, and do not re-derive it by hand. Compute it once per run, on a fix run exactly as on the first. See **2. Isolate the workspace (first run only)** (`skills/myflow-do/SKILL-rationale.md`) for why." |
| "the panel's cost must not depend on which model the operator happens to be running, and a recorded value is a deliberate decision for one change rather than an inheritance path." | 5. The review panel | — (rule extracted) | "**Every slot the panel spawns directly runs on the model the state file records under `models.reviewPanel`, defaulting to Sonnet** when that field is absent or null. There is no parent-model inheritance and no economy tier. See **5. The review panel** (`skills/myflow-do/SKILL-rationale.md`) for why." |
| "The subagent's working directory is the project worktree, which has no `skills/` tree, so a repo-relative path opens nothing and the reviewer runs with no principle list." | 5. The review panel | — (rule extracted) | "**Resolve `[PRINCIPLES_PATH]` before dispatching any principles slot.** It is the **absolute** path of `engineering-principles.md` in the directory you are reading this file from — under a global install, `~/.claude/skills/myflow-do/engineering-principles.md`. Confirm the file exists before spawning; if it does not, stop and report rather than dispatching a blind reviewer. See **5. The review panel** (`skills/myflow-do/SKILL-rationale.md`) for why it must be absolute." |
| "So an unescaped `\|` inside a cell is just text, a reordered header changes nothing, and a row that lost a boundary pipe still counts. That is the point of the split — the previous shape asked a hand-rolled table parser to recover one fact from a grammar defined in prose, and it failed **open** six distinct ways across three review passes before it was replaced." | 5. The review panel | — (rule extracted) | "`scripts/check-unfinished-work.sh` reads **only the marker block**. It never parses the table: not its header, not its column order, not its cell boundaries, not where it starts or stops. See **5. The review panel** (`skills/myflow-do/SKILL-rationale.md`) for why." |
| "so two distinct findings labelled `F1` in both the table and the marker block cannot cancel out — that shape hid an open Critical, with the word `open` never appearing in a marker at all." | 5. The review panel | — (rule extracted) | "- Each `F<n>` names **one** finding. A reused identifier is reported on each side separately. See **5. The review panel** (`skills/myflow-do/SKILL-rationale.md`) for why." |
| "This is what stops a marker quoted elsewhere — inside a fenced example, say — standing in for a marker that was never written, which is the one route that still under-counted when the redesign was attacked." | 5. The review panel | — (rule extracted) | "- The marker lines sit on **consecutive lines**, one unbroken block. See **5. The review panel** (`skills/myflow-do/SKILL-rationale.md`) for why." |
| "A status cell beside the marker is a second surface that can silently disagree with the line that governs: the machine's direction is protected — a marker reading `open` blocks whatever a cell says — but nothing protects a reader who sees `fixed` in the table and believes it. State the fact once." | 5. The review panel | — (rule extracted) | "**The table carries no status column, on purpose.** A finding's state is written once, on its marker line. To read a finding's state, look up its `F<n>` in the marker block. See **5. The review panel** (`skills/myflow-do/SKILL-rationale.md`) for why." |
| "This is why reviewing and testing are one gate: both surfaces are produced together and can never drift apart." | 6. Write the manual test guide | — (rule extracted) | "In the same run, write or refresh `docs/manual-test/<name>.md`. See **6. Write the manual test guide** (`skills/myflow-do/SKILL-rationale.md`) for why." |
| "Creating a workspace database and bucket for a run that only ever linted would leave behind resources nobody asked for and only `/myflow-finish` run 2 removes." | 7. Verify, stage, and hand off | — (rule extracted) | "**This step does not call the project's `create` command, and that is a decision rather than a gap.** `create` is called by whatever starts the project's applications, per **Project configuration** (`skills/myflow-contracts/project-configuration.md`), and this command starts none of them — it exports, lints, tests, and hands off. The applications are started at the review gate by the operator, through the project's own `## run` commands, which is where the creation and its one-time notice belong. See **7. Verify, stage, and hand off** (`skills/myflow-do/SKILL-rationale.md`) for why." |
| "A value this run resolved and did not export is a value nothing reads, so the workspace is isolated on paper while the applications and their checks still reach the project's shared resource — the same silent wrong answer as having derived nothing." | 7. Verify, stage, and hand off | — (rule extracted) | "- **Every declared row, not a subset.** See **7. Verify, stage, and hand off** (`skills/myflow-do/SKILL-rationale.md`) for why." |

**Round 2 (post-review, four rows below).** The panel's review found one Critical — a line-wrapping
defect in row 2's appendix copy, not a content defect — fixed by regenerating the passage from the
pre-task image (see the report's "Round 2" section for detail; this is not a ledger row because no
prose moved or changed, only its byte-for-byte wrapping was corrected). It also found three Important
findings: argument left in the core in three more places that Round 1 missed. A fourth eviction (the
"worktree ports" sentence) was named by the coordinator as one the reviewer flagged without pressing.
All four rows below are new rule extractions, same rules as Round 1.

| "A reviewer too many costs tokens; one too few costs a defect." | 5. The review panel § Optional slot selection | — (rule extracted) | "**Borderline → ask**, with **include** as the default. See **Optional slot selection** (`skills/myflow-do/SKILL-rationale.md`) for why." |
| "The two reasons are one case, deliberately: both end with nothing having run the guard, and a session that recognised only..." | 7. Verify, stage, and hand off | — (rule extracted) | "handoff that the validation was performed manually **and why the script was not run**. See **7. Verify, stage, and hand off** (`skills/myflow-do/SKILL-rationale.md`) for why the two reasons are treated as one case. It is never skipped for want of the script, and 'the declaration was validated' is never reported for a run in which nothing validated it." |
| "The exclusion is what keeps them out of the diff, rather than a filter applied when the diff is displayed..." | 7. Verify, stage, and hand off | — (rule extracted) | "> **Those three paths are never staged.** See **7. Verify, stage, and hand off** (`skills/myflow-do/SKILL-rationale.md`) for why." |
| "A worktree's applications bind their own ports, so the documented URL an operator opens out of habit..." | 6. Write the manual test guide | — (rule extracted) | "declared base. See **6. Write the manual test guide** (`skills/myflow-do/SKILL-rationale.md`) for why. **Resolve each URL from this worktree's workspace id, the way section 2 computed it** — ..." |

## Notes for the review panel

- No passage in this file was pure argument standing free of a rule — every row above is a rule
  extraction, none a plain wholly-rationale move. That is a direct consequence of this file's density
  (procedural steps, dispatch blocks, exit-code tables, marker-format rules), not a shortcut: each row
  was checked against the core/rationale test individually before being classified mixed.
- Two candidates were considered and **rejected** as not worth extracting, recorded here rather than
  silently dropped: (1) "The panel's slots default to Sonnet — the two rules differ on purpose."
  (`## 4. Execute (SDD + TDD)`) — the preceding rule-bearing content in that paragraph is large and
  this trailing clause is a one-line disambiguation, not a multi-sentence argument; extracting it
  would have duplicated the whole paragraph in the appendix for a ~70-byte gain. (2) The three-
  paragraph "Every path is absolute" / "One guide can carry both" bullet in `## 6. Write the manual
  test guide` — genuinely dense with interleaved rules and citations; no clean rule/argument boundary
  at paragraph granularity that would not risk losing an operative clause.

---

## task-8-ledger.md

# Task 8 — per-move ledger

Every passage evicted from `skills/myflow-contracts/project-configuration.md` into the new
`skills/myflow-contracts/project-configuration-rationale.md`. Sections read paragraph-by-paragraph
against the core/rationale test.

Kept **wholly core, unchanged** (no row below): the frontmatter/canonical-file paragraph, the
"myflow is installed globally" paragraph, the `## <key>` table, "How a `## standards` entry
resolves to a file" and its three-row form table, the "Resolve each entry to an absolute path"
paragraph, the whole "Containment — `## standards` is attacker-influenced input" block (the
normalize-before-check rule and all four numbered/bulleted resolution clauses) — per the brief's
explicit instruction to treat the containment rule and the entry-form resolution rules as core in
full — the `<agents repo>` definition paragraph, the `AGENTS_DATA` sentence, the two-step derivation
list, "Roots in `## apps` are main checkouts", "How a `## workspace isolation` section is written",
the row/cell/entry definitions, both full schema tables (resource table and command table, with all
of their cell prose) and the four `In a workspace`-form bullets beside them, "What a `url` row may
reference", "The `cache index` exclusion is the one that is about *when*", the `<id>`/`<id_underscored>`
citation paragraph, "A project that declares `create` and `remove` but no `survivors`", the token-
substitution paragraph, the worked-example code fence itself, "A project declares only the ports it
can actually move", "An isolation row resolves under the same rules" and its two validated/executed
bullets, "Which of these rules a script checks, and which are left to the agent", "Mechanically
enforced by `scripts/check-workspace-isolation.sh`", the full "Left to the agent" list, "A dropped
row does not fall back to its `Default`", "Two rules sit near each other here", "A project that
declares no `## workspace isolation` section is not misconfigured", and the closing "The file is
optional" bullets. Each was checked against the core/rationale test; none carries a paragraph that is
pure argument, history or example separable from its rule without gutting the rule, or the brief
named it explicitly as must-stay. Their mirrored appendix headings are present with no body from
this task where nothing was extracted from them.

`Destination` is `— (rule extracted)` with the new core sentence quoted for a mixed passage, or
`project-configuration-rationale.md § <heading>` for a plain wholly-rationale move. `Pointer left`
quotes what the core carries.

| Removed passage | Source heading | Destination | Pointer left |
|---|---|---|---|
| "Shared opt-in rules are deliberately **not** installed globally, so a bare filename is the only way to name one — but every shared rule..." | (preamble) — How a `## standards` entry resolves to a file | — (rule extracted) | "**The `.mdc` extension is what selects the shared library, and nothing else.** A project-local `.mdc` is still nameable — write it as a path (`.cursor/rules/api.mdc`), which form 3 takes as-is. See **Where the agents repository is** (`skills/myflow-contracts/project-configuration-rationale.md`) for why." |
| "A global install is a real `~/.claude/skills/` (likewise `~/.cursor/skills/` and `~/.codex/skills/`) holding **one symlink per skill**..." | Where the agents repository is | — (rule extracted) | "**The link to resolve is the per-skill one, never the `skills/` directory above it.** See **Where the agents repository is** (`skills/myflow-contracts/project-configuration-rationale.md`) for the measured global-install layout this guards against." |
| "The same two steps are correct for a project-local install, where nothing is a symlink at all: the skill directory is already physical..." | Where the agents repository is | `project-configuration-rationale.md § Where the agents repository is` | "The same two steps are correct for a project-local install, where nothing is a symlink at all. See **Where the agents repository is** (`skills/myflow-contracts/project-configuration-rationale.md`) for why." |
| "Check that the path you are about to use exists under it. A skill directory that was **copied** into a project rather than linked resolves two levels up..." | Where the agents repository is | — (rule extracted) | "**Confirm the derived root before using it, and treat a miss as an absence rather than a nearer guess.** Check that the path you are about to use exists under it. See **Where the agents repository is** (`skills/myflow-contracts/project-configuration-rationale.md`) for the copied-directory case this guards against." |
| "*Not read* is a statement about the resolver, not a restriction on the author: a project may write whatever a human reader needs..." | Where the agents repository is | — (rule extracted) | "**Prose beside the tables is permitted and is never read** by the resolver. See **Where the agents repository is** (`skills/myflow-contracts/project-configuration-rationale.md`) for what that prose is for." |
| "**The duplication that follows is an accepted cost, not an oversight.** A project whose public base URL and endpoint share a host must write that host..." | Where the agents repository is | `project-configuration-rationale.md § Where the agents repository is` | "A project whose rows share a host writes that host in both, since neither `url` row may reference the other. See **Where the agents repository is** (`skills/myflow-contracts/project-configuration-rationale.md`) for why that duplication is an accepted cost." |
| "a project writes one when a variable its applications read is genuinely the same in both checkouts and it wants the exported set to be the complete set..." | Where the agents repository is | — (rule extracted) | "**A `url` row carrying no token at all is a legitimate declaration, not a mistake.** Its workspace value is its `Default`, unchanged, and nothing removes or creates anything for it. See **Where the agents repository is** (`skills/myflow-contracts/project-configuration-rationale.md`) for why a project would write one." |
| "A base URL falling back to its `Default` there points at the project's shared bucket, which is the same silent wrong answer as falling back..." | Where the agents repository is | — (rule extracted) | "**A dropped `url` row is not exempt from the refusal below.** The refusal is keyed on the row having been dropped, never on whether the row names something removable: a dropped `url` row refuses in an apply worktree exactly as a dropped `database` row does, and so does a `url` row dropped because its `<value:…>` reference named no row or named another `url` row. See **Where the agents repository is** (`skills/myflow-contracts/project-configuration-rationale.md`) for the argument, also stated under **The empty id** (`skills/myflow-contracts/workspace-isolation.md`)." |
| "It has to be stated because a command in this table is routinely repo-relative: `./gradlew workspaceRemove` and `./scripts/workspace remove`..." | Where the agents repository is | — (rule extracted) | "**Each command runs with a repository root as its working directory, and which root is fixed here rather than left to the caller.** See **Where the agents repository is** (`skills/myflow-contracts/project-configuration-rationale.md`) for why it must be stated." |
| "`scripts/check-cleanup-complete.sh` launches it as `( cd \"$REPO\" && exec bash -o pipefail -c \"$cmd\" )` against the repository it was handed..." | Where the agents repository is | — (rule extracted) | "- **`survivors` runs from the main checkout**, and that is not a convention invented here. See **Where the agents repository is** (`skills/myflow-contracts/project-configuration-rationale.md`) for how `scripts/check-cleanup-complete.sh` invokes it." |
| "so the apply worktree is by then not a directory anything could run from — and two commands out of one table, called by one run of one command, must not disagree..." | Where the agents repository is | — (rule extracted) | "- **`remove` runs from the main checkout** too. Run 2 calls it after the worktree half of its cleanup step, per **Run 2 — the branch is merged** (`skills/myflow-contracts/finish-contract.md`). See **Where the agents repository is** (`skills/myflow-contracts/project-configuration-rationale.md`) for why." |
| "the same first-start-in-a-worktree fact the row in **Temporary artifacts registry** (`pipeline.md`) records. The asymmetry is the rule working..." | Where the agents repository is | — (rule extracted) | "- **`create` runs from the apply worktree**, because of who calls it: whatever starts the project's applications does, and that is the worktree whose applications need the resources. See **Where the agents repository is** (`skills/myflow-contracts/project-configuration-rationale.md`) for why the asymmetry is the rule working rather than an exception to it." |
| "A command that cannot answer in a minute needs a different design — a cached answer, a narrower query — rather than a longer bound. The outcome deliberately differs..." | Where the agents repository is | `project-configuration-rationale.md § Where the agents repository is` | "See **Where the agents repository is** (`skills/myflow-contracts/project-configuration-rationale.md`) for why the bound is stated here and how its outcome deliberately differs from the stop command's." |
| "The termination is `scripts/check-cleanup-complete.sh`'s, and it signals the process group it created — which reaches a pipeline, a subshell..." | Where the agents repository is | — (rule extracted) | "**What the bound terminates is a process group on the machine running the guard, and a command that puts its real work outside that group outlives the bound** — most of all **a command that reaches its service through a container runtime**, `docker exec …` most of all. See **Where the agents repository is** (`skills/myflow-contracts/project-configuration-rationale.md`) for the mechanism and a measured example." |
| "and it is stated here because the command is the only place it can be fixed. A guard that stays project-agnostic and single-file cannot hold..." | Where the agents repository is | — (rule extracted) | "- **That is a limitation of the bound, not a defect the guard can close.** **So a `survivors` command that crosses a container boundary carries its own timeout inside the container** — `docker exec`'s own, the client tool's connect and statement timeouts, or a wrapper that ends the query rather than the proxy. See **Where the agents repository is** (`skills/myflow-contracts/project-configuration-rationale.md`) for why the guard cannot close this itself." |
| "The three `url` rows are the shape a project reaches for as soon as one of its variables carries a URL rather than a port..." | Where the agents repository is | `project-configuration-rationale.md § Where the agents repository is` | "See **Where the agents repository is** (`skills/myflow-contracts/project-configuration-rationale.md`) for a walkthrough of the three `url` rows in the worked example above." |
| "A realistic command row often has no path to check at all: `docker exec <container> dropdb <id_underscored>` names none at all..." | Where the agents repository is | — (rule extracted) | "**No path is isolated inside a command, and that narrowing is deliberate.** Containment binds what the pipeline **reads**, validation binds what it **substitutes**, and a command is **run** — not contained. If a key in this section ever named a file the pipeline reads, that path would be contained exactly as a `## standards` path is — the rule follows the read, not the section. See **Where the agents repository is** (`skills/myflow-contracts/project-configuration-rationale.md`) for why." |
| "That is what makes the enforcement reach every project myflow is installed into, rather than only the repository the guard ships in..." | Where the agents repository is | — (rule extracted) | "**Where that enforcement actually happens, stated exactly, because \"a guard exists\" is not \"a guard ran\".** The guard runs at the point this section is *read*: `/myflow-do` runs it against each apply worktree before it resolves or exports a single row, per section 7 of `skills/myflow-do/SKILL.md`, and a non-zero exit stops that run. A project that declares no section passes silently. `/myflow-finish` deliberately does not repeat the validation. See **Where the agents repository is** (`skills/myflow-contracts/project-configuration-rationale.md`) for why." |

18 rows. Every destination heading listed above exists in `project-configuration-rationale.md`
(verified by `scripts/check-references.sh`, which resolves every bold-token citation added in this
task's diff — passes).

## Byte-verification

Every one of the 18 removed passages was extracted from the pre-task image with `sed -n '<range>p'
/tmp/t8-before.md` (image taken via `git show
47f508711197037c917a88f60f5bfa22944bafab:skills/myflow-contracts/project-configuration.md`, 45,284
bytes, matching the brief's stated starting size) and diffed byte-for-byte against the text placed
in the appendix, with the two `above`/`below` deletions applied to separate copies before that diff.
All 18 diffs are empty (script output: `E1..E18: MATCH before-image lines <range>`, `ALL_MATCH=1`).
No passage was retyped from memory anywhere in this task; every moved block was extracted
programmatically with `sed` and never hand-copied.

## Notes for the review panel

**Plan-provenance discrepancy — Step 1's "measured" heading count does not hold.** The task-8 brief
states the core's headings are `# Project configuration`, `## Where the agents repository is` and
`## workspace isolation`, tagged `<!-- measured: grep -n '^#\+ '
skills/myflow-contracts/project-configuration.md @ f763481 -->`. Re-running that exact grep at that
exact commit does return three lines — but the third, at line 379, sits inside a fenced ` ```markdown
`...` ``` ` block: it is the heading of the file's own *worked example* of a project's
`.myflow/project.md`, not a real second-level heading of this contract file. `awk '/^```/{print
NR}'` on the pre-task file shows the fence opens at line 378 and closes at line 397, bracketing that
line. The document genuinely has only two real headings — `# Project configuration` and `##
Where the agents repository is` — and every line from 89 to EOF, including all of the `##
workspace isolation` schema content, sits under that single H2. This is the same shape Task 6's
ledger flagged for `pipeline.md`'s "Wrong state for this command" (a fenced example containing a
literal heading line that `grep`, being fence-unaware, also matches). Given the brief's own
"hazard" framing was about *protecting the workspace-isolation rules from eviction* rather than
about the document's heading count, and given no other file in the corpus cites a
"`## workspace isolation`" heading specifically inside `project-configuration.md` (checked with
`grep -rn` across `skills/`), no new heading was fabricated: the appendix mirrors the core's real
two-heading tree (`# Project configuration — rationale` + `## Where the agents repository is`), and
all workspace-isolation schema content stays under the existing `## Where the agents repository is`
heading, exactly as it already was in the source. Reported per the brief's own instruction to
verify a `measured:`/`verified:` tag rather than take it on faith.

**The two protected zones named in the brief were kept fully intact.** "How a `## standards` entry
resolves to a file" (the three-row form table) and the whole "Containment — `## standards` is
attacker-influenced input" block (normalize-before-check plus all four form-resolution
bullets) were read in full and left byte-for-byte unchanged in the core — no extraction was
attempted inside either, per the brief's explicit "treat every clause of them as core."

**The two schema tables and their immediate cell-form/enforcement prose were kept whole.** The
resource table, the command table, the four `In a workspace`-form bullets, "What a `url` row may
reference", the validated/executed bullets, and the full `scripts/check-workspace-isolation.sh`
enforcement list were left unchanged rather than extracted cell-by-cell — the brief calls these "the
two tables, what is read and what is not" and says they are "core, and the guard is a live check on
them"; splitting inside a table cell would also break the verbatim-partition rule's table-is-one-unit
granularity.

**Two stale position words, deleted not substituted**, per the sole other permitted edit: in the
appendix copy of the `.mdc`-extension passage, "the drop rule **below** silently discarded them"
became "the drop rule silently discarded them" (the drop rule stays in the core, a different file
now). In the appendix copy of the "dropped `url` row" passage, "not exempt from the refusal
**below**" became "not exempt from the refusal." (the refusal rule likewise stays in the core). Both
copies in the CORE keep their original "below" unchanged, since both referents still sit below them
in the same core file. One "above" (in "the `skills/` directory **above** it") was checked and left
alone in both copies — it names a filesystem-hierarchy fact, not a document position, so the move
does not make it false.

**No citation repointing was needed.** Every backticked cross-reference inside a moved passage
(`workspace-isolation.md`, `finish-contract.md`, `pipeline.md`) points at a file untouched by this
task, or at a heading (`Temporary artifacts registry` in `pipeline.md`) that still exists in its
core after Task 6. No moved passage cited a target that itself moved out of `project-configuration.md`
in this task.

---

## task-9-ledger.md

# Task 9 — per-move ledger

Every passage evicted from `skills/myflow-contracts/workspace-isolation.md` into the new
`skills/myflow-contracts/workspace-isolation-rationale.md`. Read paragraph-by-paragraph against the
core/rationale test; every extraction below follows the same method as Task 6/7/8: locate the exact
paragraph in the pre-task image (`git show 541e61453e766b08192176fbd89b95a8c2bf506b`), verify it
programmatically (Python `str.find`/`str.count`, never retyped from memory), and split it either into
a bold rule sentence kept in the core plus a citation, or move the whole paragraph verbatim when it
carries no rule of its own.

Kept **wholly core, unchanged** (no row below): the three-values bullet list and its cache-index-
exclusion paragraph, "The shared services are not changed" paragraph, the full four-step derivation
formula, both `verified:` bash code blocks under **The workspace id**, the `<id_underscored>`
construction paragraph and its code block, the "block is checked free" paragraph and its `lsof` code
block, "The cache index is claimed by probing" paragraph, "The claim is taken atomically" paragraph
(the probe/atomic-claim procedure), "A run that cannot get an index of its own" paragraph, "Every
value derived from a workspace id resolves" paragraph (the empty-id rule itself), "The main checkout
is the empty-id case" paragraph, "A malformed row does not fall back" paragraph, both "In the main
checkout"/"In an apply worktree" refusal bullets, "The refusal is keyed on the row having been
dropped" paragraph, and "Removal is not verification" (the full survivor-report contract). Each was
checked against the core/rationale test and carries no paragraph that is pure argument, history, or
alternatives separable from its rule without gutting the rule — several are the exact content the
brief named as must-stay (derivation formulas, the `<id>`/`<id_underscored>` construction, the probe
procedure, the empty-id rule, the survivor-report contract). Their mirrored appendix headings are
present with no body from this task where nothing was extracted from them.

**One deliberate deviation from the brief's general classification hint, reported rather than
silently applied.** The brief names "the argument for probing rather than deriving the cache index"
as rationale. The paragraph carrying that argument — "**The reason is the size of the space.** A
cache offers **sixteen** indices..." — was kept **wholly core, unchanged** instead, because the very
next paragraph refers to it by position: "**That argument holds only while a probe can see the
previous claim...**" opens by pointing straight at it. Per **The split is a verbatim partition, with
citation repointing as its only edit**'s own rule — "a passage that refers to its neighbour by
position stays with that neighbour" — moving the size-of-the-space paragraph away would leave that
opening clause referring to nothing in the same file, and the only sanctioned edits to a moved or
retained passage are repointing a citation and deleting a stale `above`/`below`, neither of which
covers rewording "That argument". Keeping both paragraphs together respects the reference without
inventing a third kind of edit. Every fact in the moved paragraph (sixteen indices, the six-percent
collision rate) is also independently restated in a definitely-core paragraph later in the same
section ("Fifteen claimable indices is a real ceiling..."), so no fact is lost from the appendix by
this choice — it is simply that the size-of-the-space argument itself stays beside the rule it feeds.

`Destination` is `workspace-isolation-rationale.md § <heading>` for every row — no rule extraction in
the sense of "core states a NEW short sentence with the argument stripped out" happened here that
needed a second, distinct destination marker: every mixed passage below kept its own original bold
rule sentence(s) verbatim in the core (never rewritten) and moved the argument tail verbatim to the
appendix, with a fresh citation sentence added in the gap. `Pointer left` quotes the **entire**
resulting core paragraph (kept original wording plus the new citation sentence), so it can be checked
against the diff directly.

32 rows. Every destination heading listed below exists in `workspace-isolation-rationale.md`
(verified by `scripts/check-references.sh`, which resolves every bold-token citation added in this
task's diff — passes). Every `moved_text` chunk quoted implicitly by these rows (the appendix content,
not reproduced here to keep the table a manageable size) was verified programmatically as an exact,
unique substring of the pre-task image, and separately verified to appear verbatim in the assembled
appendix — see the report's Verification section for the method.

| Removed passage | Source heading | Destination | Pointer left |
|---|---|---|---|
| "This file is canonical for workspace isolation.** Skills,..." | (preamble) | workspace-isolation-rationale.md § The workspace id | "This file is canonical for workspace isolation.** Skills, guards, and the projects that declare isolation reference it by name; none of them restate the derivation, the port rule, or the empty-id promise. See **The workspace id** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why a second copy of a derivation is worse than a second copy of a procedure. If a rule below and a skill, a guard, or a project's own configuration ever disagree, this file wins." |
| "It exists for one failure. Two apply worktrees..." | (preamble) | workspace-isolation-rationale.md § The workspace id | "See **The workspace id** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for the failure this contract exists to prevent." |
| "**What is isolated is the logical resource, never..." | (preamble) | workspace-isolation-rationale.md § The workspace id | "**What is isolated is the logical resource, never the service that holds it.** A workspace gets its own database inside the shared database server, its own index inside the shared cache, and its own bucket inside the shared object store. See **The workspace id** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why duplicating whole services per workspace was not the design taken instead." |
| "**Every step is defined over the change name's..." | The workspace id | workspace-isolation-rationale.md § The workspace id | "**Every step is defined over the change name's UTF-8 bytes, and nothing in it is defined over characters.** See **The workspace id** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why a byte, rather than a character, is the unit two independent implementations can agree on. The derivation, stated precisely enough that two independent implementations agree:" |
| "**`LC_ALL=C` on both `tr` calls is the whole..." | The workspace id | workspace-isolation-rationale.md § The workspace id | "**`LC_ALL=C` on both `tr` calls is the whole of step 2's locale independence, and it is in the block rather than left to prose.** See **The workspace id** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for the measured locale divergence this guards against." |
| "**The normalisation runs first, before any length is..." | The workspace id | workspace-isolation-rationale.md § The workspace id | "**The normalisation runs first, before any length is counted, and that order is load-bearing.** See **The workspace id** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why the order matters and what it produces for a name carrying a dot." |
| "**A name already within `[a-z0-9-]` passes through step..." | The workspace id | workspace-isolation-rationale.md § The workspace id | "See **The workspace id** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why a name already within `[a-z0-9-]` passes through step 2 unchanged, and what that means for testing a second implementation." |
| "**`printf '%s'` rather than `echo` is the load-bearing..." | The workspace id | workspace-isolation-rationale.md § The workspace id | "**`printf '%s'` rather than `echo` is the load-bearing detail.** See **The workspace id** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why, and for which tools are acceptable." |
| "**The prefix is for humans; the digest is..." | The workspace id | workspace-isolation-rationale.md § The workspace id | "**The prefix is for humans; the digest is for correctness.** See **The workspace id** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for the collision example this separates." |
| "**Upper case is folded by the ASCII rule..." | The workspace id | workspace-isolation-rationale.md § The workspace id | "**Upper case is folded by the ASCII rule only**, never by Unicode case folding. `A`–`Z` lower;   every other byte, `İ` and `Ä` included, is a separator. See **The workspace id**   (`skills/myflow-contracts/workspace-isolation-rationale.md`) for the three-implementation   disagreement this rule settles." |
| "**Only trailing `-` are trimmed, and all of..." | The workspace id | workspace-isolation-rationale.md § The workspace id | "**Only trailing `-` are trimmed, and all of them are.** Leading and interior runs stay, so an id   may begin with `-` — a name that normalises to nothing but separators leaves the prefix empty and   the id is then `-<digest>`. See **The workspace id**   (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why every value derived from such   an id is still legal." |
| "The digest is still taken over the original,..." | The workspace id | workspace-isolation-rationale.md § The workspace id | "See **The workspace id** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why the digest, unlike the prefix, is locale-independent already and cannot be merged by normalising." |
| "**What determinism does not promise.** Four hex characters..." | The workspace id | workspace-isolation-rationale.md § The workspace id | "See **The workspace id** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for what determinism does not promise about a digest collision, and the remedy if one ever bites." |
| "**A derived database name must be safe unquoted.**..." | What the id derives | workspace-isolation-rationale.md § What the id derives | "**A derived database name must be safe unquoted.** It is spelled with `_` separators rather than the `-` the id itself uses, so it is a legal SQL identifier with no quoting at all. See **What the id derives** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why this is required of the one call site added next year, not only the ones that exist today." |
| "**A bucket name is not a SQL identifier,..." | What the id derives | workspace-isolation-rationale.md § What the id derives | "**A bucket name is not a SQL identifier, so it takes the id verbatim.** See **What the id derives** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why no second spelling is used here. The bucket carries a project prefix for the same reason the database name does — an operator listing buckets should be able to tell which project a workspace's bucket belongs to — and then the id exactly as this file derives it." |
| "That yields a multiple of 10 between 10..." | What the id derives | workspace-isolation-rationale.md § What the id derives | "That yields a multiple of 10 between 10 and 4000, so a change's ports are stable across sessions and can be bookmarked with no registry and no coordination — the same property the id has, for the same reason. See **What the id derives** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why one offset covers the whole block rather than one per port." |
| "**Any bound port discards the entire block.** If..." | What the id derives | workspace-isolation-rationale.md § What the id derives | "**Any bound port discards the entire block.** If a single port in the block is already held — by another workspace, or by an unrelated process that knows nothing about this pipeline — the **whole** block is abandoned in favour of free-port discovery. It is never repaired port by port. See **What the id derives** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why a patched block is worse than either alternative. A workspace's ports are therefore always either wholly its deterministic block or wholly discovered." |
| "**The ports actually bound are written into the..." | What the id derives | workspace-isolation-rationale.md § What the id derives | "**The ports actually bound are written into the change's manual test guide.** The guide names the URLs of the worktree that resolved them, never the project's declared defaults. See **What the id derives** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for the cost this mitigates and the failure it prevents." |
| "**What the free check does not promise.** It..." | What the id derives | workspace-isolation-rationale.md § What the id derives | "**What the free check does not promise.** What the contract requires is that the failure stay loud: a bind that fails is surfaced and the block rediscovered, never silently reassigned to whatever happened to be free at the time. See **What the id derives** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why the check only narrows the race rather than closing it." |
| "**That argument holds only while a probe can..." | The cache index | workspace-isolation-rationale.md § The cache index | "**That argument holds only while a probe can see the previous claim, so the claim is written where the next probe looks — inside the cache itself.** See **The cache index** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why a private claim would invert the comparison the section is built on. **A probing scheme whose claim is private is the derived scheme with its collision promoted to the default outcome.** A claim is therefore recorded in the shared cache, under an entry naming the workspace that holds it, so it is visible to the next claimant at the moment it is made." |
| "**Deriving the index and refusing to start on..." | The cache index | workspace-isolation-rationale.md § The cache index | "See **The cache index** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for the deriving-and-refusing alternative that was considered and rejected for this design." |
| "**There is no expiry, deliberately.** A workspace's stack..." | The cache index | workspace-isolation-rationale.md § The cache index | "**There is no expiry, deliberately.** See **The cache index** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why a lease-based lifetime would expire under a live workspace. Identity replaces it: a claim naming its holder can be released, reported and listed by name. The claim also lives exactly as long as what it protects, since it is kept in the index it reserves — a cache that restarts without persistence loses the claim and the entries together, which is correct, because there is then nothing left to keep apart." |
| "**The accepted cost is that the index is..." | The cache index | workspace-isolation-rationale.md § The cache index | "**The accepted cost is that the index is not stable across restarts.** Every other value a workspace derives is the same one a week later; this one depends on what else happened to be running when the run first claimed. See **The cache index** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why that is an accepted cost. A workspace that lands on a different index loses a login, not data — so nothing that must survive a restart may be kept there." |
| "**This pipeline releases nothing at finish, and the..." | The cache index | workspace-isolation-rationale.md § The cache index | "**This pipeline releases nothing at finish, and the claimed index has a registry row saying so.** See **The cache index** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why run 2 cannot know which index to sweep. What that leaves behind and why it is acceptable are stated once under **Temporary artifacts registry** (`skills/myflow-contracts/pipeline.md`), which is canonical for every artifact's lifetime. **A project whose claim is visible can do better than the pipeline can**, and the ceiling above is the reason to: releasing the claim in its own `remove` command, reporting a claim that outlived cleanup through `survivors`, and listing every claim on the machine without needing an id — which is what an abandoned change does not leave behind. None of that is required here, and none of it changes the registry row." |
| "**Isolation in an apply worktree is automatic, never..." | The empty id | workspace-isolation-rationale.md § The empty id | "**Isolation in an apply worktree is automatic, never opt-in.** A worktree that declares isolation receives its own database, cache index, bucket and port block without anybody asking for them, and the main checkout receives the defaults. See **The empty id** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for the opt-in alternative this rejects and why. An operator who genuinely wants the shared default still has one: the project's declared defaults are reachable through the same variables the empty-id case resolves to, so choosing them is a deliberate act rather than an omission." |
| "**Refusing rather than reporting-and-continuing is the same argument..." | The empty id | workspace-isolation-rationale.md § The empty id | "See **The empty id** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why refusing, rather than reporting and continuing, follows the same argument that makes isolation automatic." |
| "**Why the empty id is the default rather..." | The empty id | workspace-isolation-rationale.md § The empty id | "See **The empty id** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why the empty id is the default rather than a special case, and the uniform alternative it rejects. The backwards-compatibility promise is therefore load-bearing, and a change to a derived value that does not preserve it is a defect in that change, not a limitation of this contract." |
| "**The resources are created on demand, and the..." | Creation and cleanup | workspace-isolation-rationale.md § Creation and cleanup | "**The resources are created on demand, and the first creation is reported.** A run that finds no database or bucket for its workspace creates them and says so, once. See **Creation and cleanup** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why the first creation is reported. Later runs find the resources already there and print nothing, so the notice carries real information — it means *this is new*, not merely *this is here*." |
| "**A new database starts empty and is brought..." | Creation and cleanup | workspace-isolation-rationale.md § Creation and cleanup | "**A new database starts empty and is brought up to date by the project's normal migration and seeding path**, which is the same path a fresh machine and a clean CI job take. It is never created by copying an existing database. See **Creation and cleanup** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why copying was rejected." |
| "**At finish, the resources are removed.** Each is..." | Creation and cleanup | workspace-isolation-rationale.md § Creation and cleanup | "**At finish, the resources are removed.** Each is a temporary artifact and therefore has a row naming what creates it, where it lives and what removes it, in **Temporary artifacts registry** (`skills/myflow-contracts/pipeline.md`). What performs the removal is the command the project declares for it, given the workspace id so the teardown targets this change's resources rather than the project's defaults. **Which command that is belongs to the project, not to this contract** — a project names its create and remove commands in its own configuration, per **Project configuration** (`skills/myflow-contracts/project-configuration.md`), and whether it reuses a command it already had or adds one is its decision to record there. See **Creation and cleanup** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why no mechanism is named here." |
| "**A service that is not running is reported..." | Creation and cleanup | workspace-isolation-rationale.md § Creation and cleanup | "**A service that is not running is reported and skipped, rather than failed.** If the database server is down when run 2 reaches cleanup, there is nothing to remove at that moment: the skip is reported by name and the run continues. This is deliberately unlike a reported survivor, which blocks the terminal state under **Project configuration** (`skills/myflow-contracts/project-configuration.md`), and the asymmetry is stated here rather than left for a reader to find and mistake for an oversight. See **Creation and cleanup** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for the cost this asymmetry trades against, and why it matches the project-supplied stop check. A change whose project declares no isolation passes the same way: a step whose artifact is already absent is a success, which is the re-entrancy rule run 2 follows everywhere else." |
| "**That skip is signalled by one non-zero exit..." | Creation and cleanup | workspace-isolation-rationale.md § Creation and cleanup | "**That skip is signalled by one non-zero exit rather than two, deliberately.** See **Creation and cleanup** (`skills/myflow-contracts/workspace-isolation-rationale.md`) for why one exit code, and what cost that accepts." |

---

## panel-fix-ledger.md

# Panel fix ledger — moves and evictions, rounds 1–3

Per **A move or eviction is recorded in a per-move ledger**
(`openspec/changes/kan-95-slim-the-myflow-contract-files/specs/myflow-contract-economy/spec.md`).
Rounds 1 and 2 performed several rule extractions and one whole-section move with no ledger
recorded at the time; this file reconstructs them from `.superpowers/sdd/panel-fix-round-1.md`,
`.superpowers/sdd/panel-fix-round-2.md`, `.superpowers/sdd/fix-round-2.diff` and the current file
contents, and adds round 3's own moves.

A row whose `Destination` reads `— (rule extracted)` means the destination gained a freshly
authored short rule statement rather than a byte-for-byte copy of the removed passage; the
`Removed passage` column then quotes what the core file gained, not the source's original wording,
since the two differ by design (a copy would be the SSOT violation this ledger exists to prevent).

## Round 1

| Removed passage | Source heading | Destination | Pointer left |
|---|---|---|---|
| *(none evicted at this round — see the round 3 row below)*. Round 1 added a freshly authored rule to the destination, stating the same rule `README.md`'s brainstorm-expansion prose already carried, but did not remove that prose — the gap round 3's F22 later closed. | `README.md` — "How the pipeline works" § Brainstorm expansion | `skills/myflow-contracts/pipeline.md` — new **Stage exit — never the command's own judgment** section: "a stage that loops … never closes on the command's own judgment. It closes only on an explicit operator answer … The one bounded exception is a session that cannot ask at all …" | Two citation sites in `skills/myflow-start/SKILL.md` (former lines ~235, ~448) repointed from `` (`README.md`) `` to **Stage exit — never the command's own judgment** (`skills/myflow-contracts/pipeline.md`) |
| "…it is never a licence to skip the step, and never a reason to search the filesystem for a checkout that might be one." | `skills/myflow-contracts/project-configuration-rationale.md` § Confirm the derived root before using it | `skills/myflow-contracts/project-configuration.md` — appended to the core sentence ending "…treat a miss as an absence rather than a nearer guess." | **Confirm the derived root before using it** (`skills/myflow-contracts/project-configuration-rationale.md`), already present at that paragraph |

**Note on row 2:** the rationale's own copy of that clause was left in place, unchanged (per
`panel-fix-round-1.md`'s F6 section) — this row is therefore a duplication, not a clean
extraction-with-eviction, the same shape F22 found and round 3 fixed for the row above. It was not
raised as a panel finding this round and is left as-is here; flagged for a future pass.

`skills/myflow-contracts/project-configuration-rationale.md`'s single `## Where the agents
repository is` heading was also split into 13 `###` sub-headings this round (F2/F8), with the 18
citation sites in `project-configuration.md` repointed at the specific sub-heading each already
cited. No prose crossed a file boundary in that change — heading restructuring and repointing only
— so it is not a ledger row.

## Round 2

All four rows below are the "four mixed passages" `panel-fix-round-2.md`'s F9/F4 section describes,
extracted from `skills/myflow-contracts/handoff-blocks.md` (then a single file) into the new
`skills/myflow-contracts/handoff-blocks-rationale.md`, plus the one whole-section move.

| Removed passage | Source heading | Destination | Pointer left |
|---|---|---|---|
| The regeneration-vs-storage argument (why `/myflow-status` never reads back a stored copy) | `handoff-blocks.md` § The block each state renders (pre-split) | `handoff-blocks-rationale.md` § The block each state renders (pre-split; now **Why regeneration beats storage**, per round 3's F19) | — (rule extracted): core now states "No command stores the emitted text — `/myflow-status <name>` always regenerates it … never from a stored copy." and cites the appendix for why |
| The open-questions-is-not-run-only argument | same | same (now **Why the open-questions count is on-disk, not run-only**) | — (rule extracted): core states the operative "renders `none` when zero" rule directly and cites the appendix for why it is not run-only |
| The pre-check-must-run-before-the-ancestor-test mechanism/argument | same | same (now **Why the pre-check must run before the ancestor test**) | Core keeps the short factual anchor already covered by the rendering-selection table, cites the appendix for the mechanism |
| The recorded-merge-base three-conditions mechanism/argument | same | same (round 2's version; round 3's F15 later moved the *definitions* themselves back into core — see below) | Core keeps the short factual anchor, cites the appendix for the mechanism |
| The whole `### The block each state renders` orphan (~86 lines: why `Jira` is run-only, why `IN_PROGRESS` has two renderings, why `Route`/`Outstanding`/`Panel` are run-only, and the rest of the merge-status/`prUrl` reasoning) | `skills/myflow-contracts/pipeline-rationale.md` § The block each state renders (orphan, no inbound citation) | `skills/myflow-contracts/handoff-blocks-rationale.md` § The block each state renders (pre-split; now several of round 3's F19 sub-headings) | No stub left in `pipeline-rationale.md` — deleted outright, since `pipeline.md`'s own stub for that heading already cited `handoff-blocks.md`, never `pipeline-rationale.md` |

`skills/myflow-contracts/pipeline-rationale.md:83` also had "five-command" corrected to
"four-command" this round (F10) — a wording fix, not a move, so not a ledger row. Five short
lead-in sentences were added to `skills/myflow-do/SKILL-rationale.md` (F11) — pure insertion, not
a move, so not a ledger row either.

## Round 3 (this round)

| Removed passage | Source heading | Destination | Pointer left |
|---|---|---|---|
| "The stage ends only on an explicit operator answer — at the confirm, or by declining the offer, recording what is still open rather than assuming it away — never on the command's own judgment, with one bounded exception: a session that cannot ask at all records the confirm itself as an open question and ends the stage, since no operator answer could ever arrive through it. An operator who is present but silent is not that exception and still gets another round." | `README.md` § How the pipeline works § Brainstorm expansion (`#### Brainstorm — /myflow-start`) | — (rule extracted; already resident in `skills/myflow-contracts/pipeline.md`'s **Stage exit — never the command's own judgment** section since round 1 — this row closes the gap that round 1 left open, F22) | `README.md` now reads: "The stage closes only the way any pipeline stage does — **Stage exit — never the command's own judgment** (`skills/myflow-contracts/pipeline.md`) — and is not restated here." |
| The three recorded-merge-base conditions (Recorded-and-resolving / Absent / Recorded-but-unresolvable) and the imperative "resolve, then compare, never compare alone" | `skills/myflow-contracts/handoff-blocks-rationale.md` § The recorded merge base has three conditions (round 2's version, under the pre-split heading) | `skills/myflow-contracts/handoff-blocks.md` § (the paragraph beginning "The recorded merge base has three conditions, not two, and two of them are `inconclusive`.") | `handoff-blocks.md` cites **Why "recorded but unresolvable" is the dangerous condition** (`skills/myflow-contracts/handoff-blocks-rationale.md`) for why comparing it as a bare string is the mistake this guards against; the appendix paragraph was trimmed to keep only that "why" (F15) |

**Also this round (not ledger rows — no prose crossed a file boundary):**

- F19 subdivided `handoff-blocks-rationale.md`'s single `### The block each state renders` heading
  into 14 `###` sub-headings and repointed all seven citation sites in `handoff-blocks.md` at the
  specific sub-heading each one actually needs — the same shape as round 1's F2/F8, and excluded
  from this ledger for the same reason.
- F12 repointed eight bare `` (`README.md`) `` citations in `CLAUDE.md`/`AGENTS.md` (both installed
  into every target project) to installed, loaded files (`skills/myflow-contracts/pipeline.md` and
  each command's own `SKILL.md`), and F14 repointed one more in
  `skills/myflow-start/SKILL-rationale.md`. Wording/target changes only — no prose relocated.
- F13 made the preflight verdict and the unfinished-work gate (both in
  `skills/myflow-contracts/finish-contract.md` and `skills/myflow-finish/SKILL.md`) resolve the
  worktree set through **Resolving a change's worktrees** instead of reading the raw state-file map,
  and added the resolved-set-still-empty stop condition to that section. New normative text, not a
  relocation of existing prose.
- F21 rewrote two script comments (`scripts/check-cleanup-complete.sh`,
  `scripts/test-check-cleanup-complete.sh`) from "absent" to "absent or empty" to match the contract
  they describe. Wording only.

## Round 4 (this round)

| Removed passage | Source heading | Destination | Pointer left |
|---|---|---|---|
| The self-review procedure (skippable per run default-yes, gathering input via `scripts/gather-self-review-context.sh`, one combined reasoning pass across all four angles plus the 1-5 rating, the per-finding Jira filing ask, and the report path `docs/self-review/<name>-self-review.md`) | `skills/myflow-contracts/pipeline.md` § `#### Self-review — /myflow-finish run 2` (deleted by task 1 of this branch; recovered from base commit `f7634817738fcf451a673f81170f328d04c15fe9`) | `skills/myflow-contracts/finish-contract.md` § **Run 2 — the branch is merged**, step 8 (an installed file `/myflow-finish` actually loads — `openspec/` is never installed by `setup.sh`) | `skills/myflow-finish/SKILL.md` step 8 repointed from "`openspec/specs/myflow-self-review/spec.md` as canonical for the procedure" to **Run 2 — the branch is merged** (`skills/myflow-contracts/finish-contract.md`) for the procedure, keeping the openspec `Requirement:` heading only as the requirement to change first — a role, not a runtime source. `README.md`'s own self-review line (found on this round's sweep, not in the original panel brief) was repointed the same way. |
| The "a resolved worktree set that comes back empty is never a vacuous pass" rule and the list of commands it binds | `skills/myflow-contracts/finish-contract.md` § Resolving a change's worktrees (added round 3, F13) | `skills/myflow-contracts/pipeline.md` — new **Resolving a change's worktrees** section, loaded by every command | `finish-contract.md`'s own section now cites `pipeline.md`'s section for the rule instead of restating it, and keeps only what is specific to `/myflow-finish`: the git-scan fallback and the `substr`-not-`$2` path-parsing note. `skills/myflow-do/SKILL.md`'s workspace-isolation gate (line ~387) and `skills/myflow-status/SKILL.md`'s merge-status report (line ~57, plus a second occurrence at line ~83 found on this round's sweep) now resolve through the same `pipeline.md` section instead of a raw `worktrees` map read — neither of those two commands loads `finish-contract.md`, so before this round the rule they needed did not exist anywhere they could cite it. |

**Also this round (not ledger rows — wording/citation fixes, no prose relocated across a file
boundary beyond the two rows above):**

- A2 reworded five "once per recorded worktree" / "every recorded worktree" descriptions that
  contradicted the scan-fallback and empty-set-stop behaviour now in `finish-contract.md`/
  `pipeline.md`: `README.md` (three sites), `commands/myflow-finish.md`, and
  `commands-claude/myflow-finish.md` — each now describes "the resolved set" and cites **Resolving a
  change's worktrees** (`skills/myflow-contracts/pipeline.md`) instead of restating the mechanism.
  `CLAUDE.md`/`AGENTS.md`'s own `/myflow-finish` row carried the same defect and was fixed in the
  same edit as F24 below, since both problems sat on the same line.
- F24 gave six of the eight "spelled out in its own `SKILL.md`" citations in `CLAUDE.md`/`AGENTS.md`
  (the three per-command table rows in each file) a named, guard-checked section: `/myflow-start` →
  **A. Resolve the change** (`skills/myflow-start/SKILL.md`); `/myflow-do` → **1. Load context and
  validate the plan** (`skills/myflow-do/SKILL.md`); `/myflow-finish` → **Run 1 — integrate**
  (`skills/myflow-finish/SKILL.md`). The remaining two (`CLAUDE.md:86`, `AGENTS.md:132`) cite the
  bare, non-path token `` `SKILL.md` `` as a description of the general per-command pattern rather
  than one specific file, so no single path resolves for them and `check-references.sh` was never
  able to check them before or after this round; left as-is.
- S3 repointed `skills/myflow-start/SKILL-rationale.md:23-25` from **Convergence**
  (`skills/myflow-start/SKILL.md`) — circular, since that rationale file is the appendix *to* that
  section — to **Stage exit — never the command's own judgment**
  (`skills/myflow-contracts/pipeline.md`), which is where the convergence test and the
  one-test-not-a-rule-per-gate reasoning actually live.
- A1 dropped the stale count in `skills/myflow-finish/SKILL.md:48` ("the four preflight checks" →
  "the preflight checks"), since round 3 added a fifth outcome bullet (the resolved-set-empty case)
  to the list it was counting.

## Round 5 (this round)

No rows — every fix this round is new normative text or a wording/citation correction, not a
relocation of prose across a file boundary.

- Finding 1 (Critical) gave `skills/myflow-do/SKILL.md` section 2 a `/myflow-do`-specific worktree
  resolution (the worktree this run created or resumed, recorded in the run's own working notes
  before section 7's guard runs) — newly authored text, not a copy of anything that existed
  elsewhere. Round 4's F13/S2 rule in `pipeline.md`'s **Resolving a change's worktrees** explicitly
  left "how a command resolves the set beyond reading the map" to each command; `/myflow-do` had
  never stated its own answer, so its section 7 guard (repointed at that rule by round 4) had no
  non-empty set to iterate on any first run — the state file's `worktrees` map is `{}` until
  `/myflow-do`'s own section 7 write, which runs *after* the guard. Section 7's guard text was
  reworded to cite section 2's resolution rather than the bare pipeline rule.
- Finding 2 (Bugbot, red) reworded one sentence each in `CLAUDE.md:128` and `AGENTS.md:174` to name
  both `myflow-finish/SKILL.md` run headings (**Run 1 — integrate**, **Run 2 — archive and clean
  up**) instead of citing only the first for both runs' stages.
- Finding 3 (Bugbot, yellow) replaced one stale "above" in `finish-contract.md:312` with the
  citation it stood in for, **Resolving a change's worktrees** (`pipeline.md`).
- Finding 4 (Minor) qualified "authoritative list of affected worktrees" to "authoritative
  **recorded** list" in `state-file.md:147` and `myflow-do/SKILL.md` section 2, and added a sentence
  to `state-file.md:147` distinguishing the recorded map from the resolved iteration set.
