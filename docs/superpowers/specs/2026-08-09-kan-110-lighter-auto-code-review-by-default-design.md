# Lighter auto code review by default — design

**Change:** `kan-110-lighter-auto-code-review-by-default`
**Jira:** KAN-110 — "Lighter auto code review by default"
**Date:** 2026-08-09
**Planning effort:** default

## Problem

Every `/myflow-do` run dispatches the same required review panel: Primary, Bugbot and Principles,
plus whichever conditional slots the trigger table auto-includes. That roster is fixed. A
documentation-only change and a migration touching auth pay the same three required readings, and
Bugbot — the most expensive of the three — is mandatory on both.

The roster is a per-change judgment the operator is never asked to make. It should be one, with a
cheaper reading as the default.

## Approach

Record a **review panel roster preset** per change, chosen once on the run that creates it, and let
`/myflow-do` select both the panel's required slots and the shape of its per-task review from that
preset.

Three presets:

| Preset | Required slots | Conditional slots |
|--------|----------------|-------------------|
| `light` *(default)* | Primary · Principles · Code review (low) | one multi-select prompt |
| `standard` | Primary · Principles · Bugbot | one multi-select prompt |
| `full` | Primary · Bugbot · Principles | auto-include on trigger |

`full` is today's behaviour exactly. The two lighter presets change which third reading runs and
downgrade the conditional trigger table from auto-include to a prompt.

## The recorded field

`reviewPanelRoster` is a top-level state-file string — `"light"`, `"standard"` or `"full"` — beside
`planningEffort`.

- **Absent reads as `light`.** It joins `planningEffort` and `models` in the closed-schema rule's
  absent-key exception, so no state file written before this change becomes unparseable, and no
  migration pass runs. A change created before the field existed gets the new default, the same way
  an absent model reads as its default.
- **Written only by `/myflow-start`, on the run that creates the change.** Every other command
  carries it forward verbatim, exactly as it carries `jiraIssue`, `planningEffort` and `models`.
- **A revision round does not ask.** It states the recorded value and proceeds at it.
- `/myflow-status` surfaces it the way it surfaces the planning effort level: the recorded value
  verbatim, and the default named as the default when nothing is recorded.
There is no retired-key fallback to write, because the field is new, and nothing is added to
`scripts/check-vocabulary.sh`: that guard's list is *retired* vocabulary, and this field retires
nothing. The guard does constrain the wording, though — `full-panel` is a retired token in it, so
the scanned trees write "the `full` preset" and never that hyphenated form.

### Where each fact is canonical

The split mirrors how the model roles are already handled, and neither file restates the other:

- **`skills/myflow-contracts/state-file.md`** owns the field: its shape, its three values, the
  absent-key rule, and the carry-forward duty.
- **`skills/myflow-do/SKILL.md` section 5** owns what each preset *means* — the roster table and
  the ask mode — because that file is already canonical for the panel.
- **`skills/myflow-contracts/pipeline.md`'s Model policy is untouched.** A roster is not a model.
  The panel's model remains `models.reviewPanel` and is chosen independently of the roster.

## The `light` preset's third slot

Under `light`, slot 1 is **Code review (low)**: a `general-purpose` subagent dispatched on
`models.reviewPanel`, instructed to invoke the harness's `code-review` skill at effort `low`
against `.superpowers/sdd/final-review.diff` in the worktree, and to return its findings as panel
rows in its report back.

- **Its model is known**, because the dispatcher names it. The ledger records that model and never
  `unknown (agent-defined)` — that value is reserved for slots dispatched by `subagent_type`, whose
  agent definitions the dispatcher does not read.
- **Its findings are ordinary findings.** They take `F<n>` rows in the findings table and marker
  lines in the marker block. Nothing about the panel record's format changes.
- **The subagent returns findings as text.** The `code-review` skill reports through a host tool
  whose output the parent does not see, so the prompt requires the findings in the report itself.

### Fallback when the harness has no `code-review` skill

`code-review` is harness-provided. Cursor and Codex may not offer it. Where it is unavailable, the
slot becomes a `general-purpose` reviewer on the panel model with a high-confidence-defects-only
brief, and **the panel record names the substitution**.

The slot is never dropped. A `light` panel is always three required readings; degrading to two
would make an unavailable skill a silent way to weaken review.

## The ask mode

Under `light` and `standard`, the conditional trigger table is evaluated against
`final-review.diff` exactly as it is today. What changes is what happens next: every slot whose
trigger fired goes into **one** multi-select prompt naming each slot and the trigger that fired it,
with include-all the recommended answer.

Under `full`, triggered slots auto-include, which is today's behaviour. The trigger table's
existing borderline *ask* rows keep their current behaviour under `full`.

Which slots were included, which were excluded, and why, is recorded as it is today.

## Per-task review

Section 4's per-task review also reads the preset:

- Under `light` and `standard`, subagent-driven-development's spec-compliance and code-quality
  reviewers collapse to **one** combined reviewer per task, on the panel model.
- Under `full`, both run, which is today's behaviour.

The ledger line records which shape ran, so the choice is auditable after the fact.

## Bugbot's mutation-testing brief

Bugbot's dispatch prompt gains a mutation-testing section: for each behaviour the diff changes,
mutate it — flip a condition, drop a guard, move a boundary, remove a branch — and ask whether an
existing test fails. A mutation no test catches is a **surviving mutant**, and it is reported as an
ordinary finding: an `F<n>` row and an `open` marker line, blocking the handoff until a test is
added or the operator withdraws it with a reason.

This is reasoned mutation testing, not a tool. No mutation-testing framework is added, and none is
run. Bugbot is dispatched by `subagent_type` and carries its own agent definition, so the brief in
its dispatch prompt is the only lever available — which is why this rides on the prompt rather than
on the agent.

It therefore applies to the `standard` and `full` presets, the two that dispatch Bugbot. The
`light` default gets no mutation reading, which is consistent with it being the cheap option.

## What does not change

**The handoff bar does not move.** `/myflow-do` still hands off only at zero open findings at any
severity, under every preset. Escalation, the fix-round ladder, the panel record's format, the
marker-line rules and the operator handback are all unchanged.

This is stated in the delta spec rather than left implied, for the reason the planning-effort
capability already states about its own levels: a preset that could switch a gate off would be a
way to skip review, not a way to size it.

## Decisions

### Where the preset is chosen

**ID:** `preset-chosen-at-start`
**Status:** active
**Chosen:** Asked once by `/myflow-start` on the creating run and recorded in the state file — it
matches how the planning effort and the three model roles are already chosen, so one run of
questions covers everything about how the change will be built.
**Considered:** Asked at `/myflow-do` before the panel dispatches — no recording, and a fix round
could silently run a different roster than the first pass, which makes the panel record's meaning
depend on an unrecorded choice. A project-level default in `.myflow/project.md` — one roster per
repository, which cannot distinguish a documentation change from a migration in the same repo.

### Conditional slots under the lighter presets

**ID:** `conditionals-ask-not-auto`
**Status:** active
**Chosen:** Both lighter presets downgrade the trigger table from auto-include to one prompt. The
trigger still fires and the operator still sees it; only the automatic inclusion goes.
**Considered:** Leaving the trigger table unchanged — the lighter presets would then still
auto-include up to four extra slots, so the cheap preset is not reliably cheap. Cutting the
conditional slots entirely under `light` — a migration or an auth diff would get no extra lens at
all, with nothing telling the operator a trigger had fired.

### What `full` means

**ID:** `full-is-today`
**Status:** active
**Chosen:** `full` is today's behaviour exactly — Primary, Bugbot and Principles required,
conditional slots auto-included on trigger. It is the no-change option, which makes the whole
feature opt-out rather than a rewrite of the panel.
**Considered:** Every slot always, ignoring triggers — a heavier panel than has ever run, and the
trigger table becomes dead text under the one preset that would use it most. Every slot plus the
low-effort reviewer — a seventh reading with no stated reason to exist.

### Scope of the preset

**ID:** `preset-covers-per-task-review`
**Status:** active
**Chosen:** The preset governs the per-task review as well as the final panel. Per-task review is
where most reviewer dispatches happen — one pair per task — so a preset that skipped it would leave
the bulk of the cost untouched.
**Considered:** Final panel only — simpler to specify, but the lighter presets would barely be
lighter on a change with many tasks.

### The low-effort reviewer

**ID:** `low-reviewer-is-code-review-skill`
**Status:** active
**Chosen:** The harness's `code-review` skill at effort `low`, invoked by a `general-purpose`
subagent on the panel model. Its low/medium levels are defined as fewer, high-confidence findings,
which is exactly the reading a light panel wants.
**Considered:** The `cavecrew-reviewer` agent — cheapest, but its output is deliberately
compressed, and a committed panel record should read as ordinary prose. The Superpowers
code-reviewer with a narrowed brief — that is the same reviewer slot 0 already runs, so the panel
would carry two readings from one reviewer rather than two independent ones.

### Behaviour when `code-review` is unavailable

**ID:** `code-review-fallback`
**Status:** active
**Chosen:** Substitute a `general-purpose` reviewer with a high-confidence-defects-only brief and
name the substitution in the panel record.
**Considered:** Dropping the slot — an unavailable harness skill would silently become a two-slot
panel. Falling back to Bugbot — that is the `standard` preset, so a missing skill would silently
convert the operator's choice into a different, more expensive one.

### An unrecorded preset

**ID:** `absent-reads-as-light`
**Status:** active
**Chosen:** Absent reads as `light`, matching the absent-key exception `planningEffort` and
`models` already carry.
**Considered:** Absent reads as `full` — pre-existing changes would keep the review they were
planned under, but every change created before the field existed would then be the expensive case
forever, and the default would be a value nothing ever writes.

### Field name and preset names

**ID:** `field-and-preset-names`
**Status:** active
**Chosen:** `reviewPanelRoster`, with values `light`, `standard` and `full`. Roster and model stay
separate words, so `reviewPanelRoster` cannot be confused with `models.reviewPanel`.
**Considered:** `panelPreset` — shorter, but reads as a bundle of model and roster.
`models.reviewPanelRoster` — nesting a roster under `models` blurs what that object holds.
`low`/`default`/`full` for the values — mirrors the planning-effort vocabulary, but would name a
preset `default` that is not the default.

### Mutation testing rides on Bugbot's brief

**ID:** `mutation-testing-in-bugbot-brief`
**Status:** active
**Chosen:** Reasoned mutation testing, added to Bugbot's dispatch brief, with a surviving mutant
raised as an ordinary blocking finding. No tooling; Bugbot's own agent definition is not editable
from here, so the prompt is the only lever.
**Considered:** Running a real mutation-testing framework — this repository is Bash, Python and
Markdown with no such framework declared, so it would mean adopting a dependency for one panel
slot. Carrying the brief on the light preset's low-effort reviewer as well — that would put the
most expensive reading into the cheapest preset. A separate conditional Mutation slot — a fourth
required-ish reading, when the slot that already hunts defects is the natural home. Reporting a
survivor as an advisory note — it would create a second class of finding that nothing enforces,
beside a bar that enforces every other one.

## Open questions

*(none — every question raised during planning was answered)*
