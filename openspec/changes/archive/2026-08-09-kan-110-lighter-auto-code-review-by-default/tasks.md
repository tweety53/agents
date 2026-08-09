# Lighter auto code review by default — implementation plan

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development
> (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Record a review panel roster preset per change, default it to a lighter roster, and let
`/myflow-do` select both the panel's required slots and the per-task review's shape from it.

**Architecture:** One new state-file field, `reviewPanelRoster`, written by `/myflow-start` on the
creating run and carried forward by every later command. `skills/myflow-do/SKILL.md` gains a roster
table keyed by that field, a new `Code review (low)` slot, an ask mode for conditional slots, and a
mutation-testing brief on Bugbot's dispatch prompt. Every other file either writes the field, reads
it, or describes the command surface that asks for it. No script logic changes; the guards are the
verification.

**Tech Stack:** Markdown contracts and skills, Bash guard scripts, the `openspec` CLI. This
repository has no runnable application and no auto-fix command.

## Global Constraints

- **The canonical split.** `skills/myflow-contracts/state-file.md` owns the field's shape, values,
  absent-key rule and carry-forward duty. `skills/myflow-do/SKILL.md` section 5 owns what each
  preset means. `skills/myflow-start/SKILL.md` owns the creating-run question. No file restates
  another's fact; cite it instead.
- **`skills/myflow-contracts/pipeline.md` is not edited by this change.** A roster is not a model,
  and Model policy stays exactly as it is.
- **Never write `full-panel`.** That hyphenated token is retired vocabulary in
  `scripts/check-vocabulary.sh`, and `skills/`, `rules/`, `commands/`, `commands-claude/`,
  `scripts/`, `README.md`, `AGENTS.md` and `CLAUDE.md` are all scanned. Write "the `full` preset".
- **`scripts/check-vocabulary.sh` is modified in exactly one respect, and no other.** This
  constraint originally read "is not modified", on the belief that the guard's list was retired
  vocabulary this change had no reason to touch. Implementation disproved it: the guard's
  retired-state list carries `(^|[^-])\bcode-review\b` — the myflow command of that name, retired
  in the twelve-stage collapse — and that is byte-for-byte the name of the harness skill the
  `light` preset's new slot must invoke. The guard already carries a structural carve-out for
  `requesting-code-review` for the same reason. Extend that carve-out to the harness skill and
  document why in the guard's own comment block, exactly as the existing exception is documented.
  Nothing else in the guard changes, and no other token is added or removed.

  **Do not answer this collision by rewording the skill's name.** Writing it with a space passes
  the guard and names a skill the harness does not have, which is a broken instruction wearing a
  green check. The operator approved the carve-out explicitly on 2026-08-09, in preference to both
  the reworded prose and a per-line suppression marker.
- No suppression markers and no guard weakening. Every lint hit is fixed by editing the offending
  line.
- `scripts/check-contract-budget.sh` is a ratchet keyed on the path relative to the repository
  root. A file that grows past its row has its row raised deliberately, in the task that grew it.
- The three preset names are exactly `light`, `standard` and `full`, and the field is exactly
  `reviewPanelRoster`. Do not introduce a synonym in any file.
- Every path any pipeline command prints stays absolute.

## Baseline

The five files this plan edits most, and their declared budgets, measured before any task runs:

| File | Size now | Budget in `budgets()` |
|------|----------|-----------------------|
| `skills/myflow-contracts/state-file.md` | 15595 | 19202 |
| `skills/myflow-do/SKILL.md` | 35460 | 44781 |
| `skills/myflow-start/SKILL.md` | 25972 | 32465 |
| `skills/myflow-status/SKILL.md` | 15252 | 17650 |
| `skills/myflow-finish/SKILL.md` | 26774 | 32050 |

<!-- measured: for f in <the five paths>; do wc -c < "$f"; done, and sed -n '68,94p' scripts/check-contract-budget.sh @ merge-base f59b354 (the sizes BEFORE this change) -->

Every file has headroom, so no budget row is expected to need raising. Task 10 checks rather than
assumes.

---

### 1 The `reviewPanelRoster` field in the state-file contract

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/state-file.md` — the JSON example, the field bullet list, the
  absent-key exception paragraph, and the carry-forward paragraph

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the field every later task reads or writes — key `reviewPanelRoster`, top-level, string
  valued `light` | `standard` | `full`, absent reads as `light`, written only by `/myflow-start` on
  the creating run, carried forward verbatim by every other command.

- [x] **Step 1: Add the key to the JSON example**

In the JSON block near the top of the file, add the key immediately after the `models` object, so
the example reads in the order the field bullets are written:

```json unverified:confirm the surrounding block's exact indentation and trailing-comma placement before editing
  "models": {
    "implementation": null,
    "reviewPanel": null,
    "panelFix": null
  },
  "reviewPanelRoster": null,
```

- [x] **Step 2: Add the field bullet**

After the `models` bullet and before the `prUrl` bullet, add a bullet stating: the field carries the
review panel roster preset chosen for the change; its value is one of `light`, `standard` or `full`;
it is written only by `/myflow-start` on the run that creates the change, and every other command
carries it forward verbatim; its live consumer is `/myflow-do`, which selects the panel's required
slots and the per-task review's shape from it; and
`skills/myflow-do/SKILL.md` is canonical for what each preset means, which this file does not
restate. State that it is top-level rather than nested under `models` because a roster is not a
model.

- [x] **Step 3: Extend the absent-key exception**

The paragraph beginning "A state file that omits `planningEffort` or `models` entirely is valid"
currently names two keys. Rewrite it to name three, and add one sentence that is true only of the
new one: for `reviewPanelRoster`, *not recorded* resolves to the default preset rather than leaving
the panel unconfigured, so a command never has to ask which roster to use at panel time.

Leave the retired-`effort` paragraphs untouched. There is no retired key for this field.

- [x] **Step 4: Extend the carry-forward paragraph**

In the paragraph beginning "Because writes render the whole object", add `reviewPanelRoster` to the
list of fields every command must read and re-emit. Do **not** give it the special-case treatment
`planningEffort` has — there is no mapping to perform, so "re-emit as read" is the whole rule.

- [x] **Step 5: Verify**

```bash verified:both commands are declared under `## lint` in this repository's .myflow/project.md
scripts/check-references.sh
scripts/check-contract-budget.sh
```

Expected: both exit 0. `check-references.sh` resolves the new pointer to
`skills/myflow-do/SKILL.md`; `check-contract-budget.sh` confirms the file is still inside its row.

---

### 2 The roster table and the `Code review (low)` slot

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md` — section 5, "The review panel", the roster table and the
  slot table above it
- Modify: `skills/myflow-do/SKILL-rationale.md` — one subsection recording why the light preset's
  third slot is the harness's `code-review` skill and why an unavailable skill substitutes rather
  than drops

**Interfaces:**
- Consumes: the `reviewPanelRoster` field from Task 1.
- Produces: the roster table every later task refers to, and the slot name `Code review (low)`
  exactly as written here — Task 3, Task 5 and Task 9 all reproduce it.

- [x] **Step 1: Add the roster table**

At the top of section 5, before the existing slot table, state that `/myflow-do` reads
`reviewPanelRoster` from the state file, defaulting to `light` when the field is absent or null,
and add:

```markdown verified:authored in-tree for this change
| Preset | Required slots |
|--------|----------------|
| `light` *(default)* | Primary · Principles · Code review (low) |
| `standard` | Primary · Principles · Bugbot |
| `full` | Primary · Bugbot · Principles |
```

State that every preset dispatches exactly three required slots and no preset reduces that number,
and that `full` reproduces the roster in force before this change.

- [x] **Step 2: Add the slot to the existing slot table**

The existing table's rows are numbered and carry a "Required?" and a "Model" column. Add a row for
the new slot and change the "Required?" cells of the Bugbot and Code review rows so each names the
presets it belongs to rather than the bare word "always". The Primary and Principles rows stay
required under every preset.

The new row's model column is `models.reviewPanel`, and its spawn column is
`general-purpose` + the `code-review` skill at effort `low`. It is **not** a `subagent_type` slot,
so the paragraph below the table stating which slots take no model override is unchanged and must
not gain this one.

- [x] **Step 3: State how the slot is dispatched**

Below the table, add a short subsection for the slot: a `general-purpose` subagent on
`models.reviewPanel`, told to invoke the harness's `code-review` skill at effort `low` against
`.superpowers/sdd/final-review.diff` in the worktree, and to return its findings **in its report
back** — because the skill reports through a host surface the parent does not read. State that its
findings take ordinary `F<n>` rows and marker lines, and that the ledger records its real model and
never `unknown (agent-defined)`, since the dispatcher names it.

- [x] **Step 4: State the fallback**

In the same subsection: where the harness offers no `code-review` skill, the slot becomes a
`general-purpose` reviewer on the panel model briefed for high-confidence defects only, and the
panel record names the substitution. The slot is never dropped and the panel never falls back to
two required slots.

- [x] **Step 5: Record the reasoning in the rationale file**

Add a subsection to `skills/myflow-do/SKILL-rationale.md` under the same heading text the skill
section uses, covering: why the `code-review` skill rather than a narrowed Superpowers reviewer
(the latter is the same reviewer slot 0 runs, so the panel would carry two readings from one
reviewer); and why an unavailable skill substitutes rather than drops (dropping would make a
missing harness feature a silent way to weaken review, and falling back to Bugbot would silently
convert the operator's choice into a different preset).

- [x] **Step 6: State that no preset moves the handoff bar**

Immediately after the roster table, state that a preset selects how much reading the panel does and
nothing else: handoff still requires zero open findings at any severity under every preset, a minor
finding still blocks exactly as a critical one does, and the escalation ladder, fix-round rules,
panel record format, marker-line rules and operator handback are all unchanged.

This is the one requirement in this change that no script can check, so it is stated where the
roster is chosen rather than left implied by the sections that already carry the bar.

- [x] **Step 7: Verify**

```bash verified:both commands are declared under `## lint` in this repository's .myflow/project.md
scripts/check-vocabulary.sh
scripts/check-references.sh
```

Expected: both exit 0. `check-vocabulary.sh` in particular confirms no retired token entered the
scanned trees — the hyphenated form of "the `full` preset" is one of them.

---

### 3 The ask mode for conditional slots

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md` — the "Optional slot selection" subsection of section 5

**Interfaces:**
- Consumes: the roster table from Task 2.
- Produces: the ask-mode rule Task 9's command-surface descriptions summarise.

- [x] **Step 1: Split the subsection by preset**

Keep the trigger table exactly as it is — its triggers do not change under any preset. Below it,
state that what happens when a trigger fires depends on the preset: under `full` a fired trigger
auto-includes its slot, which is today's behaviour, and the table's existing borderline *ask* rows
keep their current behaviour there.

- [x] **Step 2: State the prompt for the lighter presets**

Under `light` and `standard`, every slot whose trigger fired goes into **one** multi-select prompt.
Write the prompt in the same shape the file's other operator prompts use — named options, the
recommended one marked:

```markdown verified:authored in-tree for this change; matches the named-options shape this file already uses for its other prompts
> **These triggers fired on this diff. Which slots should the panel include?**
> - **Security** — <the trigger that fired>
> - **Adversarial** — <the trigger that fired>
> - **Lens B — simplicity & state** — <the trigger that fired>
> - **Lens C — robustness & ops** — <the trigger that fired>
>
> Including all of them is the recommended answer.
```

State that only slots whose triggers actually fired appear, and that a prompt is not shown at all
when nothing fired.

- [x] **Step 3: Keep the recording rule, and extend it**

The subsection already requires recording which optional slots were included and which were
excluded and why. Add that a slot the operator declined is recorded as **declined**, distinctly from
a slot whose trigger never fired — the two are different facts about the same diff, and a reader of
the panel record must be able to tell them apart.

- [x] **Step 4: Verify**

```bash verified:both commands are declared under `## lint` in this repository's .myflow/project.md
scripts/check-vocabulary.sh
scripts/check-contract-budget.sh
```

Expected: both exit 0.

---

### 4 Bugbot's mutation-testing brief

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md` — the Bugbot row's dispatch description in section 5, plus a
  short subsection stating the brief
- Modify: `skills/myflow-do/SKILL-rationale.md` — why the brief rides on the prompt and why a
  survivor is an ordinary finding

**Interfaces:**
- Consumes: the roster table from Task 2, which is what decides whether Bugbot is dispatched at all.
- Produces: nothing later tasks depend on.

- [x] **Step 1: State the brief**

Add a subsection stating that wherever the panel dispatches Bugbot, its dispatch prompt carries a
mutation-testing brief: for each behaviour the diff changes, mutate it — flip a condition, drop a
guard, move a boundary, remove a branch — and establish whether an existing test fails. A mutation
no test catches is a **surviving mutant**.

State plainly that this is reasoned mutation testing: no framework is added, adopted or executed,
and no mutation score is computed.

- [x] **Step 2: State where a survivor lands**

A surviving mutant is an ordinary finding — an `F<n>` row and a marker line — and blocks the handoff
under the existing zero-open-findings bar until a test is added or the operator withdraws it with a
reason. It is not an advisory note outside the findings table.

- [x] **Step 3: State the scope**

The brief applies wherever Bugbot is dispatched, and nowhere else. No other slot acquires it, and
this brief adds no slot to any preset. Point at the roster table for which presets dispatch Bugbot
rather than repeating the answer here.

- [x] **Step 4: Keep the no-override rule intact**

Bugbot is still dispatched by `subagent_type` with no model override, and its ledger entry still
reads `unknown (agent-defined)`. Adding a brief to a prompt changes neither. Confirm by re-reading
the paragraph below the slot table that no edit in this task contradicts it.

- [x] **Step 5: Record the reasoning**

In `skills/myflow-do/SKILL-rationale.md`: the brief rides on the dispatch prompt because Bugbot
carries its own agent definition, which the dispatcher does not edit — the prompt is the only lever.
A survivor is an ordinary finding because an advisory class would be a second kind of finding that
nothing enforces, sitting beside a bar that enforces every other one.

- [x] **Step 6: Verify**

```bash verified:both commands are declared under `## lint` in this repository's .myflow/project.md
scripts/check-references.sh
scripts/check-contract-budget.sh
```

Expected: both exit 0.

---

### 5 The per-task review shape and the state write

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md` — section 4, "Execute (SDD + TDD)", the per-task review
  paragraph and its ledger line; and the state-write step's carry-forward list

**Interfaces:**
- Consumes: the roster table from Task 2 and the field from Task 1.
- Produces: the ledger line shape Task 9 does not need but a reviewer will check.

- [x] **Step 1: State the collapsed per-task review**

In section 4, after the per-task review paragraph, state that the shape depends on
`reviewPanelRoster`: under `light` and `standard` a **single** combined reviewer per task covers
spec compliance and code quality together, dispatched on `models.reviewPanel`; under `full` the
spec-compliance and code-quality reviewers both run, which is today's behaviour.

Cite the roster table in section 5 for the preset definitions rather than repeating them.

- [x] **Step 2: Extend the ledger line**

The existing ledger line records the task's completion and the model. Add that it also records which
per-task review shape ran — combined or the pair — so the choice is auditable after the fact, in the
same sentence that already explains why the model is recorded.

- [x] **Step 3: Extend the carry-forward list**

In the state-write step near the end of the file, add `reviewPanelRoster` to the list of fields
carried forward verbatim, beside `artifactUrl`, `jiraIssue`, `planningEffort`, `models` and `prUrl`.
It needs no special-case sentence: unlike `planningEffort`, there is no mapping to perform.

- [x] **Step 4: Verify**

```bash verified:both commands are declared under `## lint` in this repository's .myflow/project.md
scripts/check-references.sh
scripts/check-contract-budget.sh
```

Expected: both exit 0.

---

### 6 The creating-run question in `/myflow-start`

**Build:** green

**Files:**
- Modify: `skills/myflow-start/SKILL.md` — the "Ask the planning effort and the models" section, its
  heading, the revision-round rule, the state-write JSON, the handoff block, and the Guardrails list

**Interfaces:**
- Consumes: the field from Task 1.
- Produces: the recorded value every other command reads.

- [x] **Step 1: Rename the section and add the question**

The section is currently titled for the effort and the models. Retitle it so the roster is included,
and add a fourth question after the three model questions:

```markdown verified:authored in-tree for this change; matches the named-options shape the three model questions in this file already use
> **Which review panel roster should this change use?**
> - **`light`** *(default, recommended)* — primary, principles, and a low-effort Claude reviewer
> - **`standard`** — primary, principles and Bugbot
> - **`full`** — the roster in force before presets existed, with conditional slots auto-included
```

State that the answer is recorded as `reviewPanelRoster`, and cite `skills/myflow-do/SKILL.md` for
what each preset means rather than restating the roster here.

- [x] **Step 2: Extend the ask-once rule**

The section already states that the questions are asked only on the run that **creates** the change
— the state file does not exist — and never again. Extend every place that says "all four" or names
the count to cover the new question, and extend the revision-round paragraph so it states the
recorded roster alongside the recorded level and models rather than asking.

- [x] **Step 3: Add the field to the state-write JSON**

In section F's JSON block, add the key after `models`, matching Task 1's ordering:

```json unverified:confirm the surrounding block's exact key order and trailing commas before editing
  "reviewPanelRoster": "<light|standard|full, or null>",
```

- [x] **Step 4: Extend the handoff block**

The handoff's `**Recorded:**` line already names the decision count, the open-question count, the
effort level and the three models. Append the roster in the same style, with the same
"reused from the creating run" and "not recorded" variants the effort level already carries — a
not-recorded roster reports that it is planned at the default preset.

- [x] **Step 5: Extend the Guardrails**

Add a bullet beside the two that forbid asking for an effort level or a model on a revision round:
never ask for a roster on a revision round — read the recorded one and say so. Add a second bullet
stating that no roster may skip brainstorming, the design approval gate, writing-plans, or leave
`tasks.md` a scaffold, and that a roster never moves the handoff bar in `/myflow-do`.

- [x] **Step 6: Verify**

```bash verified:all three commands are declared under `## lint` in this repository's .myflow/project.md
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-contract-budget.sh
```

Expected: all three exit 0.

---

### 7 Reporting the roster in `/myflow-status`

**Build:** green

**Files:**
- Modify: `skills/myflow-status/SKILL.md` — the `jq` projection and the paragraph that describes
  what is surfaced

**Interfaces:**
- Consumes: the field from Task 1.
- Produces: nothing later tasks depend on.

- [x] **Step 1: Extend the projection**

The projection is a single literal `jq` expression listing every field the report reads, in order.
Add the new field after the `models` element:

```bash unverified:confirm the surrounding expression's exact element order before editing — the reader consumes these positionally
jq -r '.state, .branch, .prUrl, .artifactUrl, .jiraIssue, (.planningEffort // .effort), (.models // {} | tojson), (.reviewPanelRoster // null), .updatedAt, .updatedBy, (.worktrees // {} | keys[])' \
```

`// null` rather than `// "light"`: the projection reports what the file records, and resolving the
default is the reading command's job, not the projection's. A reader that positionally consumes this
output must be updated in the same edit — check the lines that consume it before changing the order.

- [x] **Step 2: Surface it**

Beside the paragraph describing how the planning effort is surfaced, state that the roster is
surfaced the same way: the recorded value verbatim when there is one, and the default named as the
default when nothing is recorded. A not-recorded roster is never a warning marker — it is the
default, which is a normal state.

- [x] **Step 3: Verify**

```bash verified:both commands are declared under `## lint` in this repository's .myflow/project.md
scripts/check-references.sh
scripts/check-contract-budget.sh
```

Expected: both exit 0.

---

### 8 Carrying the roster forward in `/myflow-finish`

**Build:** green

**Files:**
- Modify: `skills/myflow-finish/SKILL.md` — the run 2 state-write step's carry-forward list

**Interfaces:**
- Consumes: the field from Task 1.
- Produces: nothing later tasks depend on.

- [x] **Step 1: Extend the carry-forward list**

In the state-write step, add `reviewPanelRoster` to the list beside `artifactUrl`, `jiraIssue`,
`planningEffort`, `models` and `prUrl`. Leave the `planningEffort` mapping sentence untouched and do
not extend it to the new field — there is nothing to map.

- [x] **Step 2: Verify**

```bash verified:both commands are declared under `## lint` in this repository's .myflow/project.md
scripts/check-references.sh
scripts/check-contract-budget.sh
```

Expected: both exit 0.

---

### 9 The command trees and the documentation digests

**Build:** green

**Files:**
- Modify: `commands/myflow-start.md` and `commands-claude/myflow-start.md` — the description of what
  the creating run asks
- Modify: `commands/myflow-do.md` and `commands-claude/myflow-do.md` — the description of the panel
- Modify: `CLAUDE.md` — the `/myflow-start` and `/myflow-do` rows of the command table
- Modify: `AGENTS.md` — the same rows, in whatever shape that file carries them
- Modify: `README.md` — the pipeline description's account of the panel

**Interfaces:**
- Consumes: the wording settled in Tasks 2, 3 and 6.
- Produces: nothing later tasks depend on.

- [x] **Step 1: Confirm what each file actually says before editing**

```bash verified:authored in-tree for this change
grep -rn "planning effort\|review panel\|Bugbot" commands commands-claude CLAUDE.md AGENTS.md README.md
```

Each hit is a candidate. Edit only the ones that describe **what the creating run asks** or **which
slots the panel runs** — those are the two facts this change alters. Leave every other hit alone.

- [x] **Step 2: Update the `/myflow-start` descriptions**

Wherever a file says the creating run asks the planning effort and the three models, add the roster
as a fourth thing asked, with `light` named as the default. Follow each file's existing convention
for whether it names the defaults inline or cites the canonical section — do not convert one style
to the other.

- [x] **Step 3: Update the `/myflow-do` descriptions**

Wherever a file lists the panel's required slots as primary, Bugbot and Principles, restate it as
the roster the change records, with the three presets named and `light` the default. Keep every
description's existing statement that the panel hands off only at zero open findings at any
severity — this change does not touch that, and a description that drops it would read as if it had.

Write "the `full` preset". Never the hyphenated form.

- [x] **Step 4: Verify**

```bash verified:all three commands are declared under `## lint` in this repository's .myflow/project.md
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-contract-budget.sh
```

Expected: all three exit 0. `check-vocabulary.sh` scans exactly the trees this task edits, so a
retired token introduced here is caught here.

---

### 10 Full guard and test sweep

**Build:** green

**Files:**
- Modify: `scripts/check-contract-budget.sh` — only if a row is actually exceeded

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

`check-contract-budget.sh` names the path and both numbers when a file outgrows its row. If it
fires, raise **that row only**, to the file's new size plus a quarter, matching how every existing
row was set. Do not narrow the guard's scope and do not delete a row.

The baseline table above shows headroom on every file this plan edits, so a hit here means a task
grew a file far more than planned — read the diff before raising the row.

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
```

Expected: every one exits 0. No test in this list asserts anything about the panel's roster, so a
failure here means a task broke something unrelated.

- [x] **Step 4: Exercise the installer in a sandbox**

```bash verified:this is the `## run` recipe in this repository's .myflow/project.md
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
```

Expected: exit 0, and every file this plan modified present under the sandbox's `.claude/skills/`.
Never run the installer against the real home directory to check this.

- [x] **Step 5: Read the whole diff against the delta specs**

Open each of the three delta specs under
`openspec/changes/kan-110-lighter-auto-code-review-by-default/specs/` and confirm every requirement
has a corresponding edit. The requirements no script can check are the ones to read most carefully:
that no preset moves the handoff bar, that every preset dispatches three required slots, and that
`full` reproduces the previous behaviour exactly.
