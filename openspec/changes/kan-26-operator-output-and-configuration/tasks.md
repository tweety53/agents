# Operator-facing output and configuration — Implementation Plan

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** Make the pipeline's next step recoverable at any time, make what hides under each command
visible without copying a tuned value, rename planning effort, record three model choices per
change, give the progress view and the manual test guide the requirements neither has, and stop
follow-up issues accumulating one per change.

**Architecture:** This change is contract and skill text end to end — there is no new script and no
new guard. Tasks run in dependency order: the three canonical contracts first, then the five skills
that point at them, then the command trees and the docs, and the vocabulary literal **last**,
because adding it before the last `"effort":` is renamed makes `check-vocabulary.sh` fail on the
repository's own text.

**Tech Stack:** Markdown for every contract, skill, command and spec file; POSIX shell under `bash`
for the one guard edit. No new dependency of any kind.

## Global Constraints

- **No suppression markers, and no weakening of any guard's configuration.** This repository's lint
  policy is fix-first; a false hit is fixed by correcting the text or the pattern, never by silencing
  a line. The one exception already in the codebase — `# vocab-guard:allow` on the guard's own
  pattern lines — exists so the guard does not match its own literal list, and Task 11 follows that
  established idiom rather than inventing another.
- **No auto-fix command exists in this repository.** The Lint Fix Priority rule's auto-fix step is
  inapplicable here rather than skipped.
- **No per-task commits.** `pipeline.md`'s git boundaries give `/myflow-do` `git add` only, unless a
  `prUrl` is already recorded. This deliberately overrides the writing-plans template's per-task
  commit step — do not add commits.
- **All three lint guards must exit zero at the end of every task**:
  `scripts/check-vocabulary.sh`, `scripts/check-references.sh`, `scripts/check-plan-provenance.sh`.
- **All seven test harnesses must exit zero at the end of every task**, per `.myflow/project.md`
  `## test`. None of them asserts on the text this change edits, so all seven should stay green
  throughout; a failure means something unintended moved.
- **Every fenced block and numeric claim added to a planning artifact carries a provenance tag**, per
  `skills/myflow-contracts/plan-provenance.md`. The guard scans this change's `tasks.md`, `design.md`
  and `proposal.md`.
- **A cross-reference is written in one of the shapes `check-references.sh` recognises** —
  ``**Section** (`path`)``, ``**Section** in `path``, or ``see/per/under **Section** … `path`` — or
  it is not checked at all, which is worse than not writing it.
- **`myflow-effort` → `myflow-planning-effort` is a rename, not a rewrite.** Where a requirement's
  substance is unchanged, carry the existing wording across; only the concept name, the level names
  and the state-file key change.
- **Two `effort` matches in this repository are false positives and must be left alone.**
  `skills/myflow-contracts/jira-integration.md:193` reads "best-effort reconstruction" — ordinary
  English. `scripts/test-preserve-session-records.sh:8,369` contain the change *slug*
  `kan-19-finish-safety-records-and-effort`, which is an archived change's name and not a field.
  Renaming either is a defect.

## File structure

| File | Responsibility | Tasks |
|---|---|---|
| `skills/myflow-contracts/state-file.md` | `planningEffort`, `models`, the Planning effort table, the old-key read rule | 1 |
| `skills/myflow-contracts/state-self-heal.md` | The absent-key exception, widened to both new fields | 1 |
| `skills/myflow-contracts/pipeline.md` | The diagram, the two-level stage table, the handoff block template, the three model roles | 2 |
| `skills/myflow-contracts/jira-integration.md` | Follow-up naming, the join search, the append-only join write | 3 |
| `skills/myflow-start/SKILL.md` | Planning-effort question, three model prompts, tab lines, task registration | 4 |
| `skills/myflow-do/SKILL.md` | Task registration, models read from state, the guide's new register | 5 |
| `skills/myflow-finish/SKILL.md` | Task registration, tab lines, follow-up naming and joining | 6 |
| `skills/myflow-status/SKILL.md` | Handoff regeneration, `planningEffort` | 7 |
| `skills/myflow-info/SKILL.md` | Presenting the diagram it can now reach | 8 |
| `commands/`, `commands-claude/` | Whatever the sweep finds | 9 |
| `README.md`, `CLAUDE.md`, `AGENTS.md` | Diagram removed and linked; effort levels renamed | 10 |
| `scripts/check-vocabulary.sh` | The one retired literal | 11 |

**Verification note.** Nothing in this change is unit-testable: there is no new script, and the
seven existing harnesses assert on guard behaviour this change does not touch. Each task is verified
by the three guards, by the seven harnesses staying green, and by a targeted `grep` proving the
edit landed and its counterpart did not survive. Those greps are written out per task rather than
left as "check it worked".

---

### Task 1: The state file contract

**Files:**
- Modify: `skills/myflow-contracts/state-file.md`
- Modify: `skills/myflow-contracts/state-self-heal.md:30,67-68`

**Interfaces:**
- Consumes: nothing.
- Produces: the field names `planningEffort` and `models` (with sub-keys `implementation`,
  `reviewPanel`, `panelFix`), and the section name **Planning effort**. Tasks 2, 4, 5, 6, 7 and 10
  all reference these exact names; Task 2 and Task 4 cross-reference the section by title, so
  renaming it differently breaks `check-references.sh`.

- [x] **Step 1: Replace the JSON example's `effort` line**

`skills/myflow-contracts/state-file.md:31` currently reads `  "effort": null,`. Replace it, and add
the `models` object after it, so the example becomes:

```json verified:the surrounding object is state-file.md lines 22-36 read in full; only these keys change
{
  "state": "IN_PROGRESS",
  "branch": "openspec/<name>",
  "worktrees": {
    "/absolute/path/to/worktree": "<merge-base sha>"
  },
  "artifactUrl": null,
  "jiraIssue": null,
  "planningEffort": null,
  "models": {
    "implementation": null,
    "reviewPanel": null,
    "panelFix": null
  },
  "prUrl": null,
  "updatedAt": "2026-07-28T10:00:00Z",
  "updatedBy": "/myflow-do"
}
```

- [x] **Step 2: Rewrite the field's bullet**

Replace the `effort` bullet in the field list with two bullets — `planningEffort` and `models` —
each stating: who writes it (`/myflow-start`, on the creating run), that every other command carries
it forward verbatim, and that an **absent** key reads as "not recorded" while a present-and-null one
does not. Point `planningEffort` at **Planning effort** (`skills/myflow-contracts/state-file.md`)
below, and `models` at `openspec/specs/myflow-model-policy/spec.md` as canonical for the roles and
defaults. Do not restate the defaults here — a second copy is what this repository's reference guard
exists to prevent.

- [x] **Step 3: Add the old-key rule**

Immediately after those bullets, state that a file carrying the retired `effort` key is read as
recording the equivalent level (`medium` → `default`, `high` → `detailed`, `low` → `low`), is
rewritten under `planningEffort` on the next write it receives, and is **not** unparseable and **not**
announced as a correction. State that no migration pass is run. State the three rules the fallback
needs alongside it: every consumer performs the fallback (`(.planningEffort // .effort)`),
`planningEffort` wins when both keys are present, and a value outside the mapped three reads as
**not recorded** rather than making the file unparseable.

**That last rule is the one thing this step does not restore verbatim**, and the reason belongs in
the text: *unparseable* is a verdict only self-heal can act on, and no command routes a file to
self-heal on account of an unrecognised key — `/myflow-do` and `/myflow-finish` never invoke it, and
`/myflow-status` reads through a literal `jq` projection that ignores keys it does not name. Write
*not recorded*, which needs no detection to be true.

**This step was rewritten twice.** It briefly specified the opposite — the whole key unparseable,
with no compatibility read — after the review gate reversed the design decision; that reversal was
itself withdrawn when the measurement it rested on proved wrong. The history is under
`rename-reaches-capability` in `design.md`; the text above is what is in force.

- [x] **Step 4: Rename the `## Effort` section and its table**

Rename the section to `## Planning effort`. In its table, rename the levels: `medium` → `default`,
`high` → `detailed`, `low` unchanged. Update the surrounding prose so `default` is described as the
level offered as the recommendation. Update the pointer to the normative spec so it names
`openspec/specs/myflow-planning-effort/spec.md`, not the old path.

- [x] **Step 5: Widen the self-heal exception to both fields**

In `skills/myflow-contracts/state-self-heal.md`, line 30 lists the fields carried forward on a
rewrite — replace `effort` with `planningEffort` and add `models`. Lines 67-68 document the
single-field exception; widen it to cover both new fields, keeping the existing reasoning (a loud
correction for a value nobody had the opportunity to set). **Scope** the file's
"**There is no legacy-value migration**" claim to the retired `stage` field, and say what the two
cases differ in: `stage`'s twelve values were *collapsed* into three that do not correspond to them,
so any mapping would be inventing a target; the planning effort's three levels were *renamed*
one-to-one, so its mapping is a rename table over a vocabulary still in force. Leaving the claim
unqualified above the mapping table this change adds is a contract contradicting itself on the page.

- [x] **Step 6: Verify**

```bash unverified:confirm the paths after the edits; the grep expressions themselves are exact
cd <worktree>
# The retired key is gone from both contract files, and the new names are present:
grep -n '"effort":' skills/myflow-contracts/state-file.md          # expect: no output
grep -n 'planningEffort\|"models"' skills/myflow-contracts/state-file.md   # expect: several hits
grep -n '## Planning effort' skills/myflow-contracts/state-file.md          # expect: one hit
grep -n 'planningEffort\|models' skills/myflow-contracts/state-self-heal.md # expect: hits on 30 and the exception
./scripts/check-vocabulary.sh && ./scripts/check-references.sh && ./scripts/check-plan-provenance.sh
```

Expected: every guard exits 0. `check-vocabulary.sh` still passes because the `"effort":` literal is
not added until Task 11.

---

### Task 2: The pipeline contract — diagram, stage table, handoff template, model roles

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md`

**Interfaces:**
- Consumes: the section name **Planning effort** and the field names from Task 1.
- Produces: three section names later tasks cite verbatim — **Pipeline flow**, carrying the diagram
  and the two-level stage table (Tasks 8 and 10); **Progress visibility**, carrying the task-registration
  rule and its per-command step mapping (Tasks 4, 5 and 6); and a per-state handoff block template
  plus the tab-command rule inside the existing **Handoff output** section (Tasks 4, 5, 6 and 7).
  Renaming any of the three breaks `check-references.sh` in the task that cites it.

- [x] **Step 1: Add the diagram and level-1 stage table**

Add a `## Pipeline flow` section carrying the mermaid state diagram moved out of `README.md`
(Task 10 removes it there), followed by the level-1 table — one row per command, its stages in
order, human gates marked, and a marker on each stage that carries a level-2 expansion:

```mermaid verified:copied verbatim from README.md lines 57-66, read in full
stateDiagram-v2
    [*] --> STARTED: /myflow-start
    STARTED --> STARTED: /myflow-start (revise the proposal)
    STARTED --> IN_PROGRESS: /myflow-do
    IN_PROGRESS --> IN_PROGRESS: /myflow-do (fix — never moves the state)
    IN_PROGRESS --> IN_PROGRESS: /myflow-finish (run 1 — integrate)
    IN_PROGRESS --> FINISHED: /myflow-finish (run 2 — after the merge)
    FINISHED --> [*]
```

- [x] **Step 2: Write the eight level-2 expansions**

One per stage named in `openspec/changes/kan-26-operator-output-and-configuration/specs/myflow-contract-distribution/spec.md`:
brainstorm, writing-plans, SDD + TDD per task, the review panel, the preflight verdict, the
unfinished-work gate, the landing routes, and run 2's cleanup.

**Each expansion states structure and cites tuned values.** For the review panel that means: three
required slots (Primary, Bugbot, Principles) and four conditional ones (Security, Adversarial,
Principles lens B, Principles lens C); every slot on Sonnet except the two dispatched by
`subagent_type`; no handoff while any finding is open at any severity; targeted re-runs by default
with automatic escalation to a full re-run; and a handback to the operator when a finding survives
its last round. It does **not** restate the changed-line counts, the per-slot trigger lists, or the
five escalation conditions — cite the section of `skills/myflow-do/SKILL.md` that owns them, in a
shape `check-references.sh` recognises.

- [x] **Step 3: Add the per-state handoff block template**

Inside the existing **Handoff output** section, add a template per state, stating that both the
command ending in that state and `/myflow-status <name>` render from it, and that no command stores
the emitted text. Contents per state are in
`.../specs/myflow-handoff-output/spec.md`. State that a value the state file does not carry is
reported as missing rather than dropped.

- [x] **Step 4: Add a Progress visibility section**

Add `## Progress visibility` to `pipeline.md`, stating the rule once so the three skills cite it
rather than carrying three copies: every pipeline command registers its steps with the harness's
task-list mechanism and keeps their statuses current; `/myflow-status` and `/myflow-info` register
nothing; the widget is a **view, never a record**, so `tasks.md` stays the single source of
completion state and no third checkbox marker is added to it; and a harness offering no such
mechanism prints the equivalent count line and per-task list instead. State the rule
harness-neutrally — myflow runs in Claude Code, Cursor and Codex, and a rule written against one
harness's tool is unimplementable in the other two.

Give the per-command step mapping here too — `/myflow-start` its brainstorming and artifact steps,
`/myflow-do` one entry per `tasks.md` item in plan order, `/myflow-finish` the steps of the run it is
performing — so Tasks 4, 5 and 6 each cite this section rather than restating the mapping.

- [x] **Step 5: Record why the tab commands are printed rather than invoked**

In the **Handoff output** section, beside the template added in Step 3, state that every pipeline
command prints `/rename <change-name>` and `/color cyan` at the **start** of its run for the operator
to paste, that read-only commands do not, and — the part that must not be lost — **why they cannot be
invoked**: the harness's `SlashCommand` tool exposes only commands of `type: "prompt"` while both of
these are `type: "local"`/`"local-jsx"`, and no writable `/dev/tty` is available to a command, so the
escape-sequence route is closed too.

Without this, the printing reads as an oversight and the next person to look will repeat the
investigation. The evidence is in **Implementation notes** of
`openspec/changes/kan-26-operator-output-and-configuration/design.md`; summarise its conclusion
here rather than copying the table.

- [x] **Step 6: Add the three model roles to Model policy**

In the existing `## Model policy` section, add the three roles, their defaults, and that they are
asked once on the creating run and recorded in the state file. State that the panel-fix default is
the strongest available model, **not** Sonnet, and why — the role names the agent that applies a
fix, which is an implementer, and fix rounds escalate breadth rather than model precisely because
implementers already sit at the ceiling. State that the recorded value is the operator override this
section already permits, that it does not replace the per-dispatch ledger line, and that slots
dispatched by `subagent_type` take no override and still record `unknown (agent-defined)`.

- [x] **Step 7: Verify**

```bash unverified:confirm the section titles chosen in steps 1, 3 and 4
cd <worktree>
grep -n '## Pipeline flow' skills/myflow-contracts/pipeline.md        # expect: one hit
grep -c 'stateDiagram-v2' skills/myflow-contracts/pipeline.md          # expect: 1
grep -n '## Progress visibility' skills/myflow-contracts/pipeline.md   # expect: one hit
grep -n 'SlashCommand\|/dev/tty' skills/myflow-contracts/pipeline.md   # expect: hits in Handoff output
grep -n 'panelFix\|reviewPanel\|implementation' skills/myflow-contracts/pipeline.md  # expect: hits in Model policy
./scripts/check-references.sh   # every citation added in step 2 must resolve
./scripts/check-vocabulary.sh && ./scripts/check-plan-provenance.sh
```

Expected: all exit 0. `check-references.sh` is the load-bearing one here — it is what proves the
level-2 citations point at sections that exist.

---

### Task 3: The Jira contract — follow-up naming and joining

**Files:**
- Modify: `skills/myflow-contracts/jira-integration.md`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: a section named **Follow-up issues** (Task 6 references it by title).

- [x] **Step 1: Add the naming rule**

Under the existing **Labels on issues the pipeline creates** section, add a `### Follow-up issues`
section: a follow-up is titled `<KEY> follow-up` from the change's linked issue, or
`myflow follow-up` when none is linked. State that the rule governs every site that files a
follow-up, and that today the only such site is finish run 1's unfinished-work gate.

- [x] **Step 2: Add the join search**

State the search: an `AI-generated`-labelled, follow-up-titled issue at a To Do status; the newest
match is joined; the source issue does not narrow the search. Enumerate the two qualifying statuses
by name — `To Do` and `TO DO URGENT` — and state that the set is **not** derived from Jira's
`statusCategory`, naming the reason: that category groups a custom `TO DO URGENT` with `In Progress`
under `indeterminate`, which is the inference **Unrecognised statuses**
(`skills/myflow-contracts/jira-integration.md`) forbids for transitions. State that an issue at any
other status is simply not a candidate and provokes no question.

- [x] **Step 3: State that this does not reopen the unrecognised-status rule**

One short paragraph: that rule governs transitions, where inferring a position freezes the board for
a whole change; a search filter performs no transition and can only miss a candidate, after which a
new follow-up is filed. Without this the two rules read as contradicting each other.

- [x] **Step 4: Add the join write**

The items append under a dated `## From <KEY>` heading, created if absent, with everything before it
byte-for-byte unchanged. Point at the existing **Description sync**
(`skills/myflow-contracts/jira-integration.md`) for the pre-write assertion rather than restating
it. Add: the title becomes `myflow follow-up` on the first join and is left alone if already
generic; the joined issue's labels are unioned with the incoming ones, or it becomes invisible to
anyone filtering by the second source; and a failed creation or join is one
`⚠ Jira: skipped — <reason>` line with the run continuing.

- [x] **Step 5: Verify**

```bash unverified:confirm the section title chosen in step 1
cd <worktree>
grep -n 'Follow-up issues' skills/myflow-contracts/jira-integration.md   # expect: one hit
grep -n 'TO DO URGENT' skills/myflow-contracts/jira-integration.md       # expect: the existing rule plus the new search
grep -n 'statusCategory' skills/myflow-contracts/jira-integration.md     # expect: the existing rule plus the new caveat
./scripts/check-vocabulary.sh && ./scripts/check-references.sh && ./scripts/check-plan-provenance.sh
```

Expected: all exit 0.

---

### Task 4: `/myflow-start` — the questions, the tab lines, task registration

**Files:**
- Modify: `skills/myflow-start/SKILL.md` (the effort section; the state-write block at line 184; the
  announcement; section B)

**Interfaces:**
- Consumes: `planningEffort`, `models`, **Planning effort** (Task 1); the handoff template (Task 2).
- Produces: nothing later tasks depend on.

- [x] **Step 1: Rename the effort question**

Rewrite the "Ask the effort" section as "Ask the planning effort". The three offered levels become
`low` / `default` / `detailed`, with `default` marked as the recommendation. Keep every existing
rule intact: asked only when the state file is absent, never asked on a revision round, never a
command argument, and the levels' operational meanings living in **Planning effort**
(`skills/myflow-contracts/state-file.md`) rather than being restated.

- [x] **Step 2: Add the three model questions**

Immediately after the planning-effort question, add three separate questions — implementation,
review panel, panel fixes — each naming its default and marking it as the recommendation. State that
they are asked only on the creating run, and that a revision round states the recorded values
without asking. Defaults are the strongest available model, Sonnet, and the strongest available
model respectively.

- [x] **Step 3: Update the state-write block**

`skills/myflow-start/SKILL.md:184` currently reads `  "effort": "<low|medium|high, or null>",`.
Replace it and add `models`, so the block matches the object Task 1 wrote:

```json verified:the surrounding block is myflow-start/SKILL.md lines 176-190, read in full; only these keys change
{
  "state": "STARTED",
  "branch": null,
  "worktrees": {},
  "artifactUrl": "<published URL>",
  "jiraIssue": "<resolved key, or null>",
  "planningEffort": "<low|default|detailed, or null>",
  "models": {
    "implementation": "<model, or null>",
    "reviewPanel": "<model, or null>",
    "panelFix": "<model, or null>"
  },
  "prUrl": null,
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/myflow-start"
}
```

- [x] **Step 4: Update the handoff block's effort line**

The handoff's `**Effort:**` line becomes `**Planning effort:**`, with the fallback text naming
`default` rather than `medium`. Add a `**Models:**` line reporting the three recorded values, or
that they are not recorded. Both follow the template Task 2 added rather than defining a second
shape.

- [x] **Step 5: Add the tab lines and task registration**

After the `**Announce at start:**` line, print `/rename <change-name>` and `/color cyan` for the
operator to paste, per **Handoff output** (`skills/myflow-contracts/pipeline.md`). Add an
instruction to register this command's steps with the harness's task-list mechanism and keep their
statuses current, per **Progress visibility** (`skills/myflow-contracts/pipeline.md`). Cite both
sections rather than restating either rule — Task 2 put them in one place so three skills would not
carry three copies.

- [x] **Step 6: Update the guardrails**

The guardrail forbidding an effort level from skipping a gate keeps its meaning with the renamed
levels. Add one forbidding a model choice from being asked on a revision round, matching the
existing one for the effort level.

- [x] **Step 7: Verify**

```bash unverified:line 184 is the pre-edit location; confirm after the edit
cd <worktree>
grep -n '"effort":' skills/myflow-start/SKILL.md        # expect: no output
grep -n 'planningEffort' skills/myflow-start/SKILL.md   # expect: the state block and the handoff line
grep -n 'panelFix' skills/myflow-start/SKILL.md         # expect: the state block and the questions
grep -n '/color cyan' skills/myflow-start/SKILL.md      # expect: one hit near the announcement
./scripts/check-vocabulary.sh && ./scripts/check-references.sh && ./scripts/check-plan-provenance.sh
```

Expected: all exit 0.

---

### Task 5: `/myflow-do` — task registration, recorded models, the guide's register

**Files:**
- Modify: `skills/myflow-do/SKILL.md` (section 4's dispatch rule; section 5's slot table; section 6;
  the state-write paragraph at line 369; the announcement)

**Interfaces:**
- Consumes: `planningEffort`, `models` (Task 1); the handoff template and model roles (Task 2).
- Produces: nothing later tasks depend on.

- [x] **Step 1: Read the implementation model from the state file**

In section 4, the dispatch instruction currently names Opus unconditionally. Change it to: dispatch
on the model recorded in `models.implementation`, defaulting to Opus or the harness's strongest
available model when the field is absent or null. Keep every existing rule — the model is named
explicitly and never inherited, and the ledger records what each dispatch actually ran on.

- [x] **Step 2: Read the panel and fix models**

In section 5, state that directly-spawned slots take `models.reviewPanel`, defaulting to Sonnet, and
that the fix subagent takes `models.panelFix`, defaulting to the strongest available model. State
that slots 1 and 3 still receive **no** override whatever is recorded, and still record
`unknown (agent-defined)`. Do not alter the slot table's roster or its trigger table.

- [x] **Step 3: Rewrite section 6 for the new guide register**

Replace the guide's description with the register from
`.../specs/myflow-manual-test-guide/spec.md`: one tickable line per user-visible behaviour, grouped
by capability, scoped to the blast radius rather than to plan tasks; a short how-to-run preamble
with absolute paths; no transcripts, expected-output blocks or rationale. State explicitly that the
checkbox syntax and the `## Known incomplete` section are unchanged because
`scripts/check-unfinished-work.sh` reads both. Keep the existing fix-run refresh rule and the
no-skip-prompt rule verbatim.

Add the no-runnable-application clause: where `.myflow/project.md` declares no runnable app, each
check is stated as the command to run, one line each, tickable in the same way, and the guide is not
given an application shape the project does not have. This repository is that case, so a guide
written for it under the old wording would describe an app, a port and a URL that do not exist.

- [x] **Step 4: Update the state-write paragraph**

`skills/myflow-do/SKILL.md:369` reads "Carry `artifactUrl`, `jiraIssue`, `effort` and `prUrl`
forward verbatim." Replace `effort` with `planningEffort` and add `models`.

- [x] **Step 5: Add the tab lines and task registration**

After the announcement, print the two tab lines and register this command's steps, both per the
sections Task 2 added — **Handoff output** and **Progress visibility**
(`skills/myflow-contracts/pipeline.md`). What is specific to this command, and so stated here rather
than cited: an entry moves to in-progress when its implementer is dispatched and to completed when
that task passes **both** its spec and quality review — the same moment its `tasks.md` checkbox is
allowed to be ticked, so the widget and the file never disagree.

- [x] **Step 6: Verify**

```bash unverified:line 369 is the pre-edit location; confirm after the edit
cd <worktree>
grep -n '`effort`' skills/myflow-do/SKILL.md            # expect: no output
grep -n 'planningEffort\|models\.' skills/myflow-do/SKILL.md   # expect: dispatch rules and the state paragraph
grep -n 'Known incomplete' skills/myflow-do/SKILL.md    # expect: still present in section 6
./scripts/check-vocabulary.sh && ./scripts/check-references.sh && ./scripts/check-plan-provenance.sh
./scripts/test-check-unfinished-work.sh                  # the guide's machine-read shapes are unchanged
```

Expected: all exit 0, and the unfinished-work harness prints its pass line — that harness is the
check that section 6's rewrite did not disturb a shape the guard parses.

---

### Task 6: `/myflow-finish` — task registration, tab lines, follow-ups

**Files:**
- Modify: `skills/myflow-finish/SKILL.md` (the run-1 gate's filing option; the state-write paragraph
  at line 261; the announcement)

**Interfaces:**
- Consumes: **Follow-up issues** (Task 3); `planningEffort`, `models` (Task 1); the handoff template
  (Task 2).
- Produces: nothing later tasks depend on.

- [x] **Step 1: Point the filing option at the naming rule**

The unfinished-work gate's third course — "File a Jira task, then continue" — currently files an
issue with the labels rule applied. Point it at **Follow-up issues**
(`skills/myflow-contracts/jira-integration.md`) for the title and the join search, rather than
restating either. Keep the existing degradation rule: a failed filing is one skipped-with-reason
line and the run still takes the course the operator chose.

- [x] **Step 2: Update the state-write paragraph**

`skills/myflow-finish/SKILL.md:261` reads "Carry `artifactUrl`, `jiraIssue`, `effort` and `prUrl`
forward verbatim." Replace `effort` with `planningEffort` and add `models`. Both runs write state;
check whether run 2's write is described separately and update it too.

- [x] **Step 3: Add the tab lines and task registration**

After the announcement, print the two tab lines and register this command's steps, both per
**Handoff output** and **Progress visibility** (`skills/myflow-contracts/pipeline.md`). What is
specific to this command: run 1 and run 2 have different step lists, so the entries registered
follow whichever verdict `check-finish-preflight.sh` returned — a run that registers run 1's steps
and then archives would show the operator a list that never matched the work.

- [x] **Step 4: Verify**

```bash unverified:line 261 is the pre-edit location; confirm after the edit
cd <worktree>
grep -n '`effort`' skills/myflow-finish/SKILL.md        # expect: no output
grep -n 'planningEffort' skills/myflow-finish/SKILL.md  # expect: the state paragraph(s)
grep -n 'Follow-up issues' skills/myflow-finish/SKILL.md # expect: the citation added in step 1
./scripts/check-vocabulary.sh && ./scripts/check-references.sh && ./scripts/check-plan-provenance.sh
```

Expected: all exit 0. `check-references.sh` proves the **Follow-up issues** citation resolves against
the section Task 3 created.

---

### Task 7: `/myflow-status` — regenerating the handoff

**Files:**
- Modify: `skills/myflow-status/SKILL.md` (the `jq` read at line 43; the effort-surfacing paragraph
  at lines 96-104; section 4)

**Interfaces:**
- Consumes: the per-state handoff template (Task 2); `planningEffort`, `models` (Task 1).
- Produces: nothing later tasks depend on.

- [x] **Step 1: Update the state read**

Line 43's `jq` expression selects `.effort` among the fields read. Replace it with the
`planningEffort` read — falling back to the retired key, per the compatibility read restored in
round 3 — and add `.models`:

```bash verified:copied from skills/myflow-status/SKILL.md:43 in the worktree after the round-3 fix; matches byte for byte
jq -r '.state, .branch, .prUrl, .artifactUrl, .jiraIssue, (.planningEffort // .effort), (.models // {} | tojson), .updatedAt, .updatedBy, (.worktrees // {} | keys[])' \
  "$STATE_FILE" 2>/dev/null
```

- [x] **Step 2: Rewrite the effort-surfacing paragraph**

Lines 96-104 describe surfacing `effort`, with the fallback text naming the default level. Rename
the field, name `default` as the fallback level, and keep every existing rule: an absent value is
legal and never a `⚠`, and this report never writes the field. Add the same treatment for `models`.

- [x] **Step 3: Add handoff regeneration to the detail view**

In section 4, add: when a change name is given, regenerate and print the full handoff block for the
change's current state, rendered from the template in **Handoff output**
(`skills/myflow-contracts/pipeline.md`). State that nothing is read back from a stored copy, that a
value the state file does not carry is reported as missing, that `FINISHED` changes remain omitted,
and that the no-argument table form is unchanged.

- [x] **Step 4: Keep the read-only guardrail honest**

Add a guardrail: the regenerated block is printed, never acted on — no command it names is executed,
nothing is staged, no state is written. Without this, a block ending in a bare command reads as an
instruction to a future reader of the skill.

- [x] **Step 5: Verify**

```bash verified:each grep run in the worktree after the round-3 fix; the expectations below are its measured output
cd <worktree>
grep -n '\.effort' skills/myflow-status/SKILL.md         # expect: 2 hits — the jq fallback (line 43) and the paragraph explaining it (line 48); a bare `.effort` read is what must not appear
grep -n 'planningEffort' skills/myflow-status/SKILL.md   # expect: 6 hits — the jq line, the fallback paragraph (48, 52) and the surfacing paragraph (141, 162, 164)
grep -n 'Handoff output' skills/myflow-status/SKILL.md   # expect: 1 hit — the citation added in step 3
./scripts/check-vocabulary.sh && ./scripts/check-references.sh && ./scripts/check-plan-provenance.sh
```

Expected: all exit 0.

---

### Task 8: `/myflow-info` — presenting the diagram

**Files:**
- Modify: `skills/myflow-info/SKILL.md`

**Interfaces:**
- Consumes: **Pipeline flow** (Task 2).
- Produces: nothing.

- [x] **Step 1: Add the diagram to what it may present**

The skill currently leads a general "how does this work" answer with a three-line text shape. Add
that it may present the diagram and the stage table it read from **Pipeline flow**
(`skills/myflow-contracts/pipeline.md`), and that a question about what a specific command does is
answered from that command's level-2 expansion.

- [x] **Step 2: Keep the never-from-memory rule pointed at the new section**

The existing guardrail forbids describing a state, command or flag not in `pipeline.md`. Extend it
to the diagram and the stage table: present the one read during this invocation, never a remembered
one. This is the whole reason the diagram moved into `pipeline.md`, so the guardrail is what makes
the move pay off.

- [x] **Step 3: Verify**

```bash unverified:confirm the section title matches what Task 2 created
cd <worktree>
grep -n 'Pipeline flow' skills/myflow-info/SKILL.md   # expect: at least one hit
./scripts/check-references.sh                          # proves the citation resolves
./scripts/check-vocabulary.sh && ./scripts/check-plan-provenance.sh
```

Expected: all exit 0.

---

### Task 9: The command trees

**Files:**
- Modify: `commands/myflow-*.md` and `commands-claude/myflow-*.md`, as the sweep finds

**Interfaces:**
- Consumes: every name established in Tasks 1-8.
- Produces: nothing.

- [x] **Step 1: Sweep both trees**

```bash verified:run at merge-base on 2026-07-31; returned no matching files
grep -rni 'effort' commands commands-claude
```

Expected at merge base: **no output**. No command file mentions the effort level today.
<!-- measured: the grep above, run in the main checkout on 2026-07-31 @ branch main -->

If the sweep is still empty, this task's only work is Step 2. If it is not — because an earlier task
in this plan added a mention — update those files so both trees agree with the skills they delegate
to, per `openspec/specs/myflow-command-surface/spec.md`'s requirement that a command and its skill
state the same thing.

- [x] **Step 2: Check the state sets are still accurate**

No command's accepted states change in this plan. Confirm rather than assume:

```bash unverified:confirm the phrasing used in the command files
cd <worktree>
grep -rn 'STARTED\|IN_PROGRESS\|FINISHED' commands commands-claude | head -30
```

Expected: each command's stated states match the transition table in
**State transitions** (`skills/myflow-contracts/pipeline.md`), unchanged by this plan. Change
nothing if they do.

- [x] **Step 3: Verify**

```bash verified:the three commands are .myflow/project.md's ## lint block, read in full
cd <worktree>
./scripts/check-vocabulary.sh && ./scripts/check-references.sh && ./scripts/check-plan-provenance.sh
```

Expected: all exit 0.

---

### Task 10: The README and the two agent instruction files

**Files:**
- Modify: `README.md:55-84`
- Modify: `CLAUDE.md:106`
- Modify: `AGENTS.md:152`

**Interfaces:**
- Consumes: **Pipeline flow** (Task 2); the renamed levels (Task 1).
- Produces: nothing.

- [x] **Step 1: Remove the README's diagram and text shape, and link instead**

`README.md` lines 57-66 hold the mermaid diagram and lines 68-72 the three-line text shape. Remove
both, and replace them with a sentence pointing at **Pipeline flow**
(`skills/myflow-contracts/pipeline.md`) for the diagram and the per-command stage table. Keep the
prose at lines 74-82 — it describes behaviour, not the diagram. Line 84's existing "See …" sentence
already points at `pipeline.md`; fold the new pointer into it rather than adding a second.

- [x] **Step 2: Rename the effort levels in both instruction files**

`CLAUDE.md:106` and `AGENTS.md:152` carry the same `/myflow-start` table row naming the effort
question and citing **Effort** in `state-file.md`. In both: call it the planning effort, cite
**Planning effort** (`skills/myflow-contracts/state-file.md`), and add that the creating run also
asks for and records the three model choices. The two rows must stay identical to each other — they
are the same table maintained in two files.

- [x] **Step 3: Verify**

```bash unverified:line numbers are pre-edit; confirm after the edits
cd <worktree>
grep -c 'stateDiagram-v2' README.md          # expect: 0
grep -n 'Pipeline flow' README.md            # expect: at least one hit
diff <(grep -A0 '/myflow-start <name>' CLAUDE.md) <(grep -A0 '/myflow-start <name>' AGENTS.md)
./scripts/check-vocabulary.sh && ./scripts/check-references.sh && ./scripts/check-plan-provenance.sh
```

Expected: the `diff` prints nothing — the two rows are identical — and every guard exits 0.

---

### Task 11: The vocabulary guard

**Files:**
- Modify: `scripts/check-vocabulary.sh` (`check_retired_stage_vocabulary`)

**Interfaces:**
- Consumes: every rename from Tasks 1-10 being complete. **This task runs last, and the ordering is
  load-bearing:** the guard scans `skills rules commands commands-claude scripts README.md AGENTS.md
  CLAUDE.md`, so adding the literal while any `"effort":` survives in those targets fails the
  repository's own lint.
- Produces: nothing.

- [x] **Step 1: Confirm the rename is complete first**

```bash verified:both hits located in the main checkout on 2026-07-31, before any edit
grep -rn '"effort":' skills rules commands commands-claude scripts README.md AGENTS.md CLAUDE.md
```

At merge base this returns exactly two hits — `skills/myflow-start/SKILL.md:184` and
`skills/myflow-contracts/state-file.md:31`. After Tasks 1 and 4 it must return **nothing**.
<!-- measured: the grep above, run in the main checkout on 2026-07-31 @ branch main -->

If it returns anything, stop and fix that file before touching the guard. Adding the literal first
would make the guard report the repository as dirty and there would be no honest way to silence it.

- [x] **Step 2: Add the literal to the retired pattern**

Append one alternation to the pattern built in `check_retired_stage_vocabulary`, following the
existing idiom exactly — each pattern line carries a trailing `# vocab-guard:allow` so the guard
does not match its own list:

```bash verified:idiom copied from scripts/check-vocabulary.sh lines 220-231, read in full
  pattern+='|"effort":'                                                  # vocab-guard:allow
```

Add a comment above the block recording **why only this literal**: `medium` and `high` are ordinary
English throughout this repository, `Medium` is also a Jira priority name, and the guard's own header
states it proves a fixed list of literals is absent rather than that a rename is complete. Matching
them would produce hits that are not drift, and the only way to silence those is a
`vocab-guard:allow` marker on a line that is telling the truth — which the header already says
teaches the guard to lie. This is the same reasoning the existing `checkpoint` comment records, so
follow its wording rather than inventing a new argument.

- [x] **Step 3: Prove the guard catches a reintroduction**

**Mutate a sandbox copy, never the worktree.** The guard resolves its own repo root from the script's
location, so a copied tree is a complete, self-scoping target and the real tree is never touched.

**The control set covers every spelling the pattern claims to catch**, one sandbox per spelling.
Task 14 Step 10 widened the entry after pass 5 measured two of these passing clean, so the set below
is the one that proves the widened pattern fires rather than the one that proved its predecessor did.

```bash verified:run in this worktree during fix round 5; each of the three sandboxes exited 1 and the real tree exited 0
cd <worktree>
# The guard's DEFAULT_TARGETS, plus scripts/ so the copied guard resolves its root to the sandbox.
# Positive controls: the guard must FAIL on each spelling of a reintroduced key, in the COPY.
for payload in '{ "effort": null }' '{ "effort" : null }' "{ 'effort': null }"; do
  SB="$(mktemp -d)"
  cp -R skills rules commands commands-claude scripts README.md AGENTS.md CLAUDE.md "$SB"/
  printf '%s\n' "$payload" >> "$SB"/skills/myflow-contracts/state-file.md
  "$SB"/scripts/check-vocabulary.sh >/dev/null 2>&1; echo "$payload -> expected non-zero, got $?"
  rm -rf "$SB"
done
# Negative control: the real tree, unmodified throughout, still passes.
./scripts/check-vocabulary.sh; echo "expected 0, got $?"
```

Expected: non-zero for each sandbox, with the offending `file:line` reported under its `$SB` when the
run is not silenced, then 0 for the real tree. A guard entry never demonstrated to fire is an entry
nobody knows works — and this repository has already paid for that lesson once, in
`check-plan-provenance`'s history.

**Never revert the mutation with `git checkout`, `git restore`, `git stash` or `git reset`.** This
change is implemented without per-task commits, so the whole of it is uncommitted in the worktree:
any of those discards real work and the guard then exits 0 over the loss, reporting a clean run that
is clean only because the work is gone. Copy, mutate the copy, delete the copy.

- [x] **Step 4: Full verification**

```bash verified:the ten commands are .myflow/project.md's ## lint and ## test blocks, read in full
cd <worktree>
./scripts/check-vocabulary.sh
./scripts/check-references.sh
./scripts/check-plan-provenance.sh
./scripts/test-setup.sh
./scripts/test-check-references.sh
./scripts/test-check-plan-provenance.sh
./scripts/test-check-finish-preflight.sh
./scripts/test-preserve-session-records.sh
./scripts/test-check-unfinished-work.sh
./scripts/test-check-cleanup-complete.sh
```

Expected: all three guards and all seven harnesses exit 0. This is the whole-change gate — every
earlier task ran the three guards, and this is the first point at which the seven harnesses are all
required to pass together with the guard's new entry in place.

- [x] **Step 5: Sandboxed installer pass**

```bash verified:copied from .myflow/project.md ## run, read in full
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
```

Expected: exits 0. The renamed contract sections and the new `## Pipeline flow` section are installed
as part of `skills/myflow-contracts/`, so a broken symlink or a missing file surfaces here rather
than at the next command invocation.

---

### Task 12: Reconcile the handoff templates with what the commands print

**Added after the plan was approved**, on the operator's decision, because the review panel found a
requirement this plan introduces and then violates. **This task runs before Task 7**, which builds
`/myflow-status`'s regeneration on the templates it corrects. It is numbered 12 rather than
inserted so no earlier task's number moves.

**Why.** `myflow-handoff-output`'s first requirement says the block's shape is defined in exactly
one place and that *both* the command ending in a state and `/myflow-status <name>` render from that
template. Task 2 added the templates without reconciling them against the blocks the three commands
already print, and those blocks disagree with the templates in label style and in field set. Left
alone, `/myflow-status` and `/myflow-start` would print visibly different blocks for the same state
— the drift this capability exists to prevent.

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md` (`### The block each state renders`)
- Modify: `specs/myflow-handoff-output/spec.md` (Step 5 — added mid-task, after this task's own review
  failed it for leaving the spec contradicting the contract)

**Interfaces:**
- Consumes: the templates and section titles from Task 2; `planningEffort` and `models` from Task 1.
- Produces: templates Task 7 renders from. Task 7 must not need to edit them.

**Divergences to reconcile** — established by reading all four blocks at the merge base plus Tasks
1–5, not assumed:

| Renderer | Carries beyond the template | Label style |
|---|---|---|
| `STARTED` template | — | `Proposal artifact:` |
| `/myflow-start` | `Change`, `Decisions recorded`, `Jira description (pre-edit)`, `Models` | `**Artifact:**` |
| `IN_PROGRESS` template | — | `Worktree:` |
| `/myflow-do` | `Change`, `Panel`, `Progress`, `Git`, `Jira description (pre-edit)` | **mixed** — bold for the summary fields, plain and column-padded for `Worktree:` / `Test guide:` |
| `/myflow-finish` run 1 | `Change`, `Route`, `PR`, `Outstanding`; shares only `Next:` with the template | `**Change:**` |
| `/myflow-finish` run 2 | `## Finished`, and `## Cleanup incomplete` — a third `IN_PROGRESS` rendering | `**Change:**` |

<!-- measured: the four blocks read in full in the apply worktree on 2026-07-31, after Task 5; the
     /myflow-do label-style, run-1 and run-2 rows corrected on the same day after Task 12's
     implementer checked them against the files and found the original three rows wrong -->

**The corrected rows are the load-bearing ones.** Reading `/myflow-do`'s style as uniformly bold and
applying "adopt the bold style" across the board creates a fresh mismatch instead of closing one;
and run 2 was missing entirely, though the contract denied its `## Finished` block existed at all.
Implement per the files, not per this table.

- [x] **Step 1: Make the template carry what the commands print**

Extend the `STARTED` and `IN_PROGRESS` templates to carry the fields the commands already emit, and
adopt the `**Label:**` style all three skills use, so one style is described and rendered. The three
skills then already conform and need no edit — which is why this task's Files list holds one file.

- [x] **Step 2: Name the run-only fields as an explicit exception**

`Jira description (pre-edit)` cannot be regenerated: no local copy of the pre-edit text survives the
run that wrote it, so `/myflow-status` can never reproduce it. State it in the template as a
**run-only** field — emitted by the command that ends in the state, absent from a regenerated block
— rather than leaving it a silent mismatch. This is the same honesty the section's existing
"reported as missing, not dropped" rule already requires of a value the state file lacks.

- [x] **Step 3: Give `/myflow-finish` run 1 its own template**

Run 1 ends at `IN_PROGRESS` but hands off a branch waiting on a merge, not a diff waiting on review:
a worktree path, a test-guide path and a staged-diff command are all wrong for it. State that
`IN_PROGRESS` has two renderings — after `/myflow-do`, and after `/myflow-finish` run 1 — and give
the second its own template carrying `Route`, `PR` and `Outstanding`. State which one
`/myflow-status` regenerates and how it decides: a recorded `prUrl` means run 1 has happened.

Without this the requirement is unsatisfiable rather than merely unsatisfied — no single
`IN_PROGRESS` block is correct for both commands.

- [x] **Step 4: Qualify the level-2 panel expansion's Sonnet claim**

Found by Task 5's reviewer. `#### The review panel — /myflow-do` still states "every slot runs on
Sonnet except the two dispatched by `subagent_type`" as an absolute, while the same file's
`## Model policy` now carries the recorded-override mechanism. Qualify it as the default it is, and
cite Model policy rather than restating the override. Both statements are in `pipeline.md`, so one
file disagreeing with itself is this task's to close.

- [x] **Step 5: Bring the delta spec's `IN_PROGRESS` row with the split**

Found by Task 12's own review, which failed the task on it. `specs/myflow-handoff-output/spec.md`
states one `IN_PROGRESS` contents row and applies it to "the regenerated block", so the two-rendering
split of Step 3 contradicts the very requirement this task exists to satisfy. Split the row on
whether `prUrl` is recorded, and scope or duplicate the scenarios beneath it so none asserts of both
renderings what is true of one. Point at `pipeline.md` for the discriminator's limitation rather than
explaining it twice.

The row must describe what a **regenerated** block carries, which is thinner than what run 1 prints:
`Route` and `Outstanding` are run-only, and are omitted rather than reported missing.

- [x] **Step 6: Verify**

```bash unverified:confirm the heading names after the edits
cd <worktree>
grep -n 'run-only' skills/myflow-contracts/pipeline.md            # expect: the Step 2 exception
grep -c '^\*\*`' skills/myflow-contracts/pipeline.md              # expect: one per template
./scripts/check-vocabulary.sh && ./scripts/check-references.sh && ./scripts/check-plan-provenance.sh
```

Then, by hand, diff each template against the block its command prints and confirm every field in
the block appears in the template or is named run-only. That comparison is the actual gate; the
greps only prove the edit landed.

---

### Task 13: Fix round 4 — the pass-4 panel findings

The pass-4 panel raised F40–F48 and the run stopped rather than opening this round, recording its
reasoning for the operator to overrule. The operator overruled it by re-invoking `/myflow-do`. Every
finding below was re-verified against the working tree before this task was written; none was taken
on the reviewer's word alone.

**Files:**
- Modify: `skills/myflow-contracts/jira-integration.md` (F40, F43)
- Modify: `skills/myflow-contracts/state-self-heal.md` (F41, F44, F45, F48)
- Modify: `skills/myflow-contracts/state-file.md` (F42)
- Modify: `openspec/changes/kan-26-operator-output-and-configuration/design.md` (F41)
- Modify: `openspec/changes/kan-26-operator-output-and-configuration/proposal.md` (F46)
- Modify: `openspec/changes/kan-26-operator-output-and-configuration/tasks.md` (F47)

**Interfaces:**
- Consumes: the contracts as they stand after fix round 3.
- Produces: nothing later tasks depend on — this is the final round.

- [x] **Step 1: Scope the partial-join retry claim to the merge boundary (F40, Critical)**

`jira-integration.md:404` claims without qualification that "a partial join is re-attempted, not
abandoned, and keeps its `⚠` until it is complete". The retry path lives in run 1, which is reached
only while the branch is unmerged; once preflight returns `RUN2` the path is unreachable and no
durable record names which follow-up was partially joined. Qualify the claim to the window in which
it is true, and say what happens past it. This is a **carryover**: it was folded into round 3's brief
as prose under F37 instead of being recorded as its own finding, and `fix-round-3.diff` left
`:384-385` and `:404-412` byte-identical.

- [x] **Step 2: Give the append guard a provenance check (F43, Important)**

`jira-integration.md:422`'s "already present" set is built from the target issue's live description,
which any project member can edit. A forged `## From <KEY>` section makes a run skip appending real
outstanding work and report clean success, and the confirmation gate shows title and status only —
never the description — so it cannot catch it. Close it within the existing guard rather than adding
a trust model the proposal does not describe.

- [x] **Step 3: Condition the `planningEffort` exemption on the case it argues (F41, Critical)**

`state-self-heal.md:112` makes the exemption from being named among unrecovered fields
**unconditional**, but its justification argues only the absent-or-unmapped case. A file holding a
real operator-set level that is unparseable for an *unrelated* reason is rebuilt, the level nulled,
and the announcement never names it — contradicting the same file's own "name every field that could
not be recovered". Corroborated: `design.md:139-142` omits `planningEffort` from its loss list, which
names `branch`, `artifactUrl`, `jiraIssue`, `models`, `prUrl` and the merge base. Fix both.

- [x] **Step 4: Resolve the self-heal contract's self-contradiction (F44, Important)**

`state-self-heal.md:25` says self-heal infers `state` **and**, on a rebuild, the keys of `worktrees`;
line 30 still says its "only owned field is `state`" and that `worktrees` is "re-emitted exactly as
read". The canonical contract disagrees with itself five lines apart.

- [x] **Step 5: Stop restating the level mapping (F45, Important)**

`state-self-heal.md:87` restates the `medium`/`high`/`low` pairs to make its `stage`-vs-effort
argument, then twenty lines later claims the mapping is "stated once" elsewhere — false of this very
file. The argument needs only "renamed 1:1", not the literal pairs.

- [x] **Step 6: Bring the worked example with the clause it demonstrates (F48, Minor)**

`state-self-heal.md:43`'s correction-announcement example was not updated to show the `worktrees`
clause the paragraph beneath it now mandates — the one template an implementer copies literally.

- [x] **Step 7: State the unmapped-value justification once (F42, Important)**

The same three empirical claims justifying "unmapped reads as *not recorded*, not *unparseable*" are
independently authored in `state-file.md:107-121` and `state-self-heal.md:112-122`. State it once and
cite it from the other — the discipline round 3 applied to F38.

- [x] **Step 8: Correct the disproved Impact bullet (F46, Important)**

`proposal.md:99` still reads "none in flight records a non-null effort, so the rename affects no
change currently open" — the exact claim pass 3 disproved and `design.md:131-137` corrected: five
state files carry the retired key, four hold a non-null value, and one of those is this change's own
open file. Planning artifacts archive verbatim, so this ships as durable wrong text if left.

- [x] **Step 9: Resync Task 7's jq block and its companion grep (F47, Important)**

`tasks.md:569` shows the `jq` expression without the `// .effort` fallback that round 3 restored,
under a provenance tag certifying "the jq expression itself is exact"; the real line 43 of
`skills/myflow-status/SKILL.md` reads `(.planningEffort // .effort)`. Its companion grep at `:597`
expects no output but now returns two hits. Correct the block, the tag if it no longer holds, and the
grep's expectation.

- [x] **Step 10: Verify**

```bash unverified:heading and line numbers shift with the edits above; confirm after they land
cd /Users/tweety53/Projects/agents-worktrees/openspec-kan-26-operator-output-and-configuration
./scripts/check-vocabulary.sh && ./scripts/check-references.sh && ./scripts/check-plan-provenance.sh
```

Expected: all exit 0. Then, by hand, confirm each of the nine findings above is closed at its own
location — the guards prove the tree is well-formed, not that a finding was answered.

---

### Task 14: Fix round 5 — the pass-5 panel findings

Pass 5 ran the full 7-slot roster against the round-4 tree. All nine round-4 findings verified closed
by three independent slots. Pass 5 raised nine new findings of its own: none Critical against the
diff, four Important, five Minor. Every one below was re-verified against the working tree before
this task was written.

Pass 5's adversarial slot reported after this task was dispatched, adding F58, F59 and F60 — two
Important and one Minor — as Steps 9 to 11. The verification step moved to the end behind them,
which is where every other task in this plan keeps it.

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md` (F49, F56, F58)
- Modify: `CLAUDE.md`, `AGENTS.md`, `README.md`, `skills/README.md` (F49) — `skills/README.md` also
  carries F57
- Modify: `skills/myflow-contracts/jira-integration.md` (F50, F51, F52, F54, F60)
- Modify: `skills/myflow-contracts/state-self-heal.md` (F55)
- Modify: `scripts/check-vocabulary.sh` (F59), and Task 11 Step 3's control set with it
- Modify: `specs/myflow-handoff-output/spec.md` (F56, F58) and `specs/myflow-jira-projection/spec.md`
  (F50, F51, F52, F54, F60)
- **Unchanged:** `skills/myflow-status/SKILL.md`. F56's claim there is that `prUrl` is consulted only
  where the merge probe was inconclusive, and the file explicitly defers to the selection table for
  the rule; correcting that table makes the claim true where it stands, and restating it here would
  be the duplication the file forbids.

**Interfaces:**
- Consumes: the contracts as they stand after fix round 4.
- Produces: nothing later tasks depend on.

- [x] **Step 1: Narrow or satisfy the stage-table uniqueness claim (F49, Important)**

`pipeline.md:43` asserts the diagram and stage table "are the only copy in this repository" and that
"no skill carries a second copy". `CLAUDE.md:107` and `AGENTS.md:155` each carry an ordered
arrow-form stage sequence per command, and they have **drifted**: `state gate` and `document the fix`
appear 0 times in either file while `pipeline.md`'s `/myflow-do` row lists both. Both files already
apply the citation discipline twice ("deliberately not repeated here") in the same table rows.

- [x] **Step 2: Scope the follow-up search to the project (F50, Important)**

`jira-integration.md:259` describes the join search as scanning "the project", and every disclosure
of the attack surface says "any project member". No JQL project constraint exists anywhere —
`grep -rn 'searchJiraIssuesUsingJql\|project = '` over `skills/ commands/ scripts/ rules/` returns
only two prose mentions. So the population able to plant a join candidate is anyone who can create an
issue in any project the connection can query. `design.md:298-313` built its risk-acceptance on the
single-project premise. Add the constraint, derived from `## jira` or the linked issue's key prefix.

- [x] **Step 3: Neutralise fence delimiters in the displayed title (F51, Important)**

The confirmation gate renders the candidate's title inside a ` ```text ` fence, then prints the
`<n>` of `<m>` already-recorded count **after** the closing fence. The sanitisation folds control
characters, whitespace and terminal escapes but says nothing about backticks, which are ordinary
printable ASCII. A title that is a bare backtick run closes the fence early; the template's own
closing fence then opens a dangling block that swallows the count — **the exact signal round 4 added
to make forged evidence visible before the write** — along with the Yes/No options.

- [x] **Step 4: Cover Unicode format characters in the fold (F52, Minor)**

"Every control character" plausibly reads as category `Cc` only, leaving `Cf` — bidi overrides,
zero-width joiners — uncovered, in a paragraph that otherwise takes care over terminal escapes.

- [x] **Step 5: Give the join outcome table a row for partial-append-then-failure (F54, Minor)**

`jira-integration.md`'s outcome table has `joined — new items only` (implying the retitle and label
union succeeded) and `partially joined` (silent on whether the append was full or partial). A retry
that appends only the newly-missing items and then fails the retitle has no literal row, so an
implementer must improvise. The safe direction is preserved either way; this is a reporting gap.

- [x] **Step 6: Name the multi-repo limit of worktree recovery (F55, Minor)**

`state-self-heal.md`'s new rebuild recovers `worktrees` keys by scanning "each affected repository",
but for a multi-repo change the only on-disk record of which repositories participate is the
`worktrees` key set — the very thing that is unparseable. A rebuild therefore recovers the current
repository's worktree and silently drops siblings. Pre-existing in net effect, but now load-bearing
for a feature this change adds, so name it rather than leaving it implicit.

- [x] **Step 7: Correct the `prUrl` scope claim (F56, Important)**

`pipeline.md:565` says `prUrl` is the tiebreaker "only where the stronger signal is genuinely
absent"; `:616` says the one-way gap "now applies only to the inconclusive rows above". The table at
`:568-575` splits `not merged (proven)` by `prUrl` too, rendering `none` as `/myflow-do`. That row is
reachable: the *handle it manually* route completes run 1, commits, pushes and leaves `prUrl` null,
so an unmerged branch renders "Implementation staged — review and test" for work already committed
and already past the human gate. `myflow-status/SKILL.md:216` repeats the claim. Either correct both
scope claims, or route that row by the `HEAD`-differs-from-merge-base signal, which is already
available and independent of `prUrl`.

- [x] **Step 8: Rewrap the pasted line (F57, Minor)**

`skills/README.md:30` was pasted without rewrapping to the ~90-character width the rest of the file
uses. Cosmetic only.

- [x] **Step 9: Mark the `STARTED` Jira line run-only (F58, Important)**

`pipeline.md`'s `STARTED` template carries `**Jira:** <issue key and the transition made, …>` with no
`(run-only)` marker, while the line under it has one. But the field needs the transition *this run
made*, and the state file holds the bare `jiraIssue` key with no transition history;
`myflow-status/SKILL.md` forbids calling Jira at all, and even a live read cannot separate a fresh
transition from an already-correct status without the pre-run status, which nothing records. Mark it,
add the short why-paragraph its siblings have, and bring the delta spec's contents table and its
run-only enumeration into agreement.

- [x] **Step 10: Widen the retired `effort` literal (F59, Important)**

The entry matches one exact spelling. Measured with Task 11 Step 3's sandbox recipe: a space before
the colon and a single-quoted key both pass clean. Tolerate optional whitespace before the colon and
either quote style — a widening, which the lint policy permits where narrowing is forbidden — keep
the `# vocab-guard:allow` idiom on the pattern line, and extend Task 11 Step 3's control set so every
spelling the entry claims is one somebody proved fires.

- [x] **Step 11: Do not retitle a change's own follow-up (F60, Minor)**

The retitle fires whenever the title is not yet `myflow follow-up`, including when the matched
candidate is this change's own `<KEY> follow-up`. Harmless but pointless — it renames a correctly
named issue. One clause on the retitle rule and on the idempotency guard that mirrors it.

- [x] **Step 12: Verify**

```bash verified:run in this worktree after every edit of this task landed; all three guards and all seven harnesses exited 0
cd /Users/tweety53/Projects/agents-worktrees/openspec-kan-26-operator-output-and-configuration
./scripts/check-vocabulary.sh && ./scripts/check-references.sh && ./scripts/check-plan-provenance.sh
for t in scripts/test-*.sh; do "$t" >/dev/null 2>&1 || echo "FAILED: $t"; done; echo harnesses-done
```

Expected: all three guards exit 0 and the loop prints `harnesses-done` with no `FAILED:` line. Then
confirm each finding is closed at its own location.

---

### Task 15: Fix round 6 — the pass-6 panel findings

**Operator override, recorded.** The escalation ladder allows five fix rounds and this is the sixth.
At the handback the operator was offered a bounded course (fix the four Important, withdraw the six
Minor), a stop, and a fix-without-review; they chose **fix all ten and re-run the full panel**. The
ladder's limit is therefore overridden deliberately, not exceeded by omission, and pass 7 runs the
complete roster afterwards.

Pass 6 raised no Critical against the diff. All twelve round-5 findings were verified closed, several
independently. Every finding below was re-verified against the working tree before this task was
written.

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md` (F61)
- Modify: `commands/myflow-finish.md`, `commands-claude/myflow-finish.md` (F62)
- Modify: `specs/myflow-handoff-output/spec.md` (F63)
- Modify: `skills/myflow-contracts/jira-integration.md` (F64, F65)
- Modify: `openspec/changes/kan-26-operator-output-and-configuration/design.md` (F66)
- Modify: `skills/myflow-contracts/state-self-heal.md` (F67)
- Modify: `scripts/check-vocabulary.sh` (F68)
- Modify: `skills/myflow-info/SKILL.md` (F69)
- Modify: `CLAUDE.md`, `AGENTS.md`, `skills/README.md` (F70)

**Interfaces:**
- Consumes: the contracts as they stand after fix round 5.
- Produces: nothing later tasks depend on.

- [x] **Step 1: Stop restating cited reasoning (F61, Minor)**

`pipeline.md:709-714` cites the progress-visibility passage as answering "the identical question",
then restates the general claim anyway — `:366-367` and `:711-712` say the same thing in different
words. `:601-604` argues against exactly this. Keep the citation, drop the restatement, go straight to
the two Claude-Code-specific facts that are actually new there.

- [x] **Step 2: Bring both command files in step with their skill (F62, Important)**

`commands/myflow-finish.md:14` and `commands-claude/myflow-finish.md:10` both read "offers exactly
three courses: continue, stop, or file a Jira task". The skill and `pipeline.md:816` now read "File or
join a Jira follow-up, then continue", joining being the usual outcome since F18. `README.md:383-385`
states the rule: a command contradicting its skill is a defect, not a shorthand. Task 9's sweep was
`grep -rni 'effort'`, so no task ever looked at this wording — check both trees for any other Jira
phrasing that drifted the same way.

- [x] **Step 3: Make the contents table's completeness claim true (F63, Minor)**

`specs/myflow-handoff-output/spec.md:29` claims to list "the template's fields minus its run-only
ones" but omits `**Change:**`, which is neither run-only nor missing-capable.

- [x] **Step 4: Constrain and specify the JQL project clause (F64, Important)**

The clause is built from `## jira`, which `project-configuration.md:28` constrains in no way, and the
contract never says how the clause is assembled or quoted. The fallback path *is* pinned to
`[A-Z]{2,10}`. A value carrying JQL syntax widens the very scoping F50 established. Require the same
key shape and reject anything else, **and** state how the clause is assembled — the exacting treatment
the title's untrusted characters already get.

- [x] **Step 5: Cover line and paragraph separators in the fold (F65, Minor)**

The `Cc`/`Cf` fold does not reach `Zl`/`Zp` (U+2028, U+2029), so "step 1 guarantees the title is
exactly one line" is not justified for a renderer honouring them as hard breaks. Extend the fold or
state why they are out of scope.

- [x] **Step 6: Record the project-clause decision in design.md (F66, Minor)**

No decision entry names the mandatory JQL project clause; the existing entries still read as though a
project-wide scan was implemented behaviour rather than a gap just closed. This file archives verbatim
and has drifted this way twice before (F34, F46).

- [x] **Step 7: State the retired-key rewrite where self-heal writes (F67, Important)**

`state-self-heal.md:31` lists `planningEffort` among fields "re-emitted exactly as read", without the
non-byte-copy caveat `state-file.md:196-202` makes prominent and that `/myflow-do` and `/myflow-finish`
each received. An `effort`-keyed file routed through self-heal for an **unrelated** contradiction could
be re-emitted unmigrated, or have `planningEffort` written `null` — the erasure the pass-3 handback
restored the compatibility apparatus to prevent. One sentence, citing rather than restating.

- [x] **Step 8: Correct the vocabulary guard's claim or its anchor (F68, Minor)**

`check-vocabulary.sh:265` matches quoted-word-plus-colon with no structural JSON anchor. Measured:
`Two things determine 'effort': the level and the model.` trips it, exit 1. Dormant today. By the
guard's own recorded reasoning for excluding `medium` and `high`, a pattern that can hit truthful
prose invites a `vocab-guard:allow` marker on a line telling the truth. Correct the comment's claim or
tighten the anchor — **never** add a suppression marker.

- [x] **Step 9: Resolve myflow-info's self-contradiction (F69, Important)**

`skills/myflow-info/SKILL.md:47` hardcodes the three-line shape while `:80-82` forbids presenting a
remembered diagram. The frozen copy has already drifted: it reads `terminal (it integrates on its
first run)` where `pipeline.md:20` reads `terminal (second run — see the finish contract)`. A live
self-contradiction in the one command whose mandate is never to answer from memory.

- [x] **Step 10: Settle the three-line digest's treatment (F70, Minor)**

`CLAUDE.md:66`, `AGENTS.md:112` and `skills/README.md:7` keep the digest that `README.md`'s equivalent
was replaced with a citation for. Either treat all four alike or record why these three differ. Note
this digest is the always-on rule's trigger block, which may be the reason — if so, say so.

- [x] **Step 11: Verify**

```bash unverified:line numbers shift with the edits above; confirm after they land
cd /Users/tweety53/Projects/agents-worktrees/openspec-kan-26-operator-output-and-configuration
./scripts/check-vocabulary.sh && ./scripts/check-references.sh && ./scripts/check-plan-provenance.sh
for t in scripts/test-*.sh; do "$t" >/dev/null 2>&1 || echo "FAILED: $t"; done; echo harnesses-done
```

Expected: all exit 0. Then confirm each of the ten findings is closed at its own location.

---

### Task 16: Fix round 7 — the pass-7 panel findings

**Second operator override, recorded.** The ladder allows five fix rounds; this is the seventh. At the
pass-7 handback the operator chose **fix all seven, closing the guard findings by narrowing the
guard's claim rather than widening its pattern**, followed by a targeted re-run rather than a full
pass.

Pass 7 raised zero Critical, three of seven slots returned clean, and the primary answered *ready for
the human gate*. The count is falling: 12 → 10 → 7.

**Files:** `CLAUDE.md`, `AGENTS.md`, `rules/myflow-manual-review.mdc` or `skills/myflow-contracts/pipeline.md` (F71); `specs/myflow-handoff-output/spec.md` (F72); `skills/myflow-contracts/jira-integration.md` (F74, F75); `scripts/check-vocabulary.sh` + `docs/manual-test/<name>.md` (F76, F77). F73 is already corrected.

- [x] **Step 1: Settle the trigger digest against its canonical source (F71, Important)**

The three trigger copies are byte-identical to each other and assert they are "kept identical", but
read `terminal (second run — it integrates first)` where `pipeline.md:20` reads `terminal (second run
— see the finish contract)`. F69 made `/myflow-info` read that block **live**, so an operator now sees
both renderings in one session — from the session-start file and from the command whose mandate is
never to answer from memory. Either align the wording, or state that the trigger copy may paraphrase
and will not be kept byte-identical. The "kept identical" claim must not survive unqualified.

- [x] **Step 2: Rewrap the long line (F72, Minor)**

`specs/myflow-handoff-output/spec.md:45` is 125 characters against the file's ~90–100 wrap.

- [x] **Step 3: Pin the key check to a whole-value match, and state tokenization (F74, Important)**

The "no escaping needed" argument rests entirely on `[A-Z]{2,10}` being a **whole-value** match, which
the contract never says. The same file uses *search* language for a near-identical pattern at `:28`,
demonstrably knows how to say otherwise at `:316` ("an exact match on two shapes, never a substring
search"), and its own counter-example `KAN" OR project != "KAN` begins with `KAN` — so it would
survive a search reading. Add the anchoring qualifier in the style already used at `:316`, and state
how the `## jira` body is split into candidate keys before the shape check runs, the way `## standards`
gets an explicit normalization procedure.

- [x] **Step 4: Replace the false ordering illustration (F75, Minor)**

`jira-integration.md:390`'s worked example does not hold: step 2 collapses whitespace runs **to a
single space, never zero**, so two delimiter pairs separated by a space stay separated and cannot join
into a run of four. **The adopted order is sound** — fold and collapse never touch delimiters and never
delete a separator entirely, so they can only preserve or split adjacency, and truncation is
suffix-only. Replace the illustration with one that demonstrates the real danger, or state the
invariant abstractly. Do not change the order.

- [x] **Step 5: Narrow the guard's claim rather than widen its pattern (F76, F77, Important)**

Two measured defects: the comma anchor reopens the false-positive class (`Three settings exist,
"effort": low, medium, high are the names.` trips it), and both alternatives require the key to be
**quoted**, so every unquoted spelling evades — `effort: low`, `- effort: low`, `` `effort: null` ``,
and the exact sentence this change deletes. The unquoted gap is **not** a round-6 regression: the
round-5 pattern required quotes too, so that coverage never existed.

**The operator's instruction is to fix the claim, not chase the pattern.** This entry has produced a
finding in three consecutive passes (F59 → F68 → F76/F77), each fix breeding the next, because the
pattern is being asked to prove something no literal list can. The guard's own header already states
the truth — it proves a fixed list of literals is absent, not that a rename is complete. Bring the
entry's comment, and the residual bullet in `docs/manual-test/<name>.md`, into line with what the
guard actually does. Removing the comma anchor's false positive is in scope; a new claim of
completeness is not.

**Measured, round 7.** The control set below was run against the round-6 guard (taken from the git
index) and against the round-7 guard, one payload per run, appended to a **copy** of the scan set:

```bash verified:run in this worktree; 24 payloads x 2 guard versions, results in the table below
SB="$(mktemp -d)"; cp -R skills rules commands commands-claude scripts \
  README.md AGENTS.md CLAUDE.md "$SB"/
printf '%s\n' '<payload>' >"$SB"/skills/probe-payload.md
"$SB"/scripts/check-vocabulary.sh >/dev/null 2>&1; echo "rc=$?"   # 1 = caught, 0 = clean
rm -rf "$SB"
```

| Payload | round 6 | round 7 |
|---------|---------|---------|
| `{ "effort": null }`, `{ "effort" : null }`, `{ 'effort': null }` | caught | caught |
| `"effort": "low"` at line start; `"effort": "low"` mid-sentence | caught | caught |
| `{ "a": 1, "effort": null }`, `{ "a": 1, "effort": <level> }`, `{ …, "effort": "low" }` | caught | caught |
| `` The `"effort": null` entry is retired. `` | caught | caught |
| `Three settings exist, "effort": low, medium, high are the names.` (and the `'…'` spelling) | **caught (F76)** | clean |
| `As for scope, "effort": that is what this section is about.` | **caught (F76)** | clean |
| `{ "a": 1, "effort": low }`, `settings, "effort": medium` — comma, unquoted value | caught | **clean** |
| `effort: low`, `- effort: low`, `` `effort: null` ``, `` `effort` ``, `effort: "low"` | clean (F77) | clean (F77) |
| `Two things determine 'effort': the level and the model.` (F68) | clean | clean |
| best-effort reconstruction; the `kan-19-…-and-effort` slug; "planning effort" prose | clean | clean |

Row 7 is the whole cost of dropping the comma: a quoted key after a comma whose value is *also*
unquoted, which is not JSON and is the false-positive shape itself. Row 8 is F77 — unchanged, and
now stated in the comment rather than implied.

- [x] **Step 6: Verify**

```bash verified:run in this worktree after every edit of this task landed; all three guards, all seven harnesses, openspec validate --strict and the sandboxed installer exited 0
cd /Users/tweety53/Projects/agents-worktrees/openspec-kan-26-operator-output-and-configuration
./scripts/check-vocabulary.sh && ./scripts/check-references.sh && ./scripts/check-plan-provenance.sh
for t in scripts/test-*.sh; do "$t" >/dev/null 2>&1 || echo "FAILED: $t"; done; echo harnesses-done
```

Expected: all exit 0. Re-run the guard's full control set — every spelling it still claims to catch
must be caught, and no claim may outrun the measurement.
