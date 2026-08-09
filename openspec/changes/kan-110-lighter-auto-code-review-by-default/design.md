## Context

`/myflow-do` dispatches a fixed required review panel — Primary, Bugbot and Principles — plus
whichever conditional slots the trigger table auto-includes, and runs a spec-compliance and a
code-quality reviewer on every task. Nothing about that roster varies with the change. A
documentation-only diff and a migration touching auth pay the same three required readings, and
Bugbot, the most expensive of the three, is mandatory on both.

Three things already vary per change and are recorded in its state file: the planning effort level
and the three model roles. They are asked once, on the run that creates the change, and carried
forward verbatim by every later command. The roster is the obvious fourth, and it is the one that
governs most of the cost.

The approved design is
`docs/superpowers/specs/2026-08-09-kan-110-lighter-auto-code-review-by-default-design.md`.

## Goals / Non-Goals

**Goals:**

- Make the panel's roster a recorded, per-change choice, with a cheaper default.
- Keep a three-slot required panel under every preset, so no preset is a way to run less than three
  independent readings.
- Reduce the per-task reviewer count under the lighter presets, which is where most dispatches
  happen.
- Give the operator sight of a fired conditional trigger even when the slot is not auto-included.
- Add a mutation-testing reading to the slot that already hunts defects.

**Non-Goals:**

- Changing the handoff bar, the escalation ladder, the fix-round rules, or the panel record's
  format. All are untouched.
- Changing how the panel's *model* is chosen. `models.reviewPanel` is unaffected; a roster is not a
  model.
- Adopting a mutation-testing framework. The mutation reading is reasoned, not tooled.
- Migrating existing state files. An absent key reads as the default.
- Changing the conditional trigger table's own triggers. Only what happens after a trigger fires
  changes, and only under the two lighter presets.

## Decisions

### Where the preset is chosen

**ID:** `preset-chosen-at-start`
**Status:** active
**Chosen:** Asked once by `/myflow-start` on the creating run and recorded in the state file — it
matches how the planning effort and the three model roles are already chosen, so one run of
questions covers everything about how the change will be built.
**Considered:** Asked at `/myflow-do` before the panel dispatches — nothing recorded, and a fix
round could silently run a different roster than the first pass, which makes the panel record's
meaning depend on an unrecorded choice. A project-level default in `.myflow/project.md` — one
roster per repository, which cannot distinguish a documentation change from a migration in the same
repository.

### Conditional slots under the lighter presets

**ID:** `conditionals-ask-not-auto`
**Status:** active
**Chosen:** Both lighter presets downgrade the trigger table from auto-include to one multi-select
prompt. The trigger still fires and the operator still sees which one, with include-all
recommended; only the automatic inclusion goes.
**Considered:** Leaving the trigger table unchanged — the lighter presets would still auto-include
up to four extra slots, so the cheap preset is not reliably cheap. Cutting the conditional slots
entirely under `light` — a migration or an auth diff would get no extra lens at all, with nothing
telling the operator a trigger had fired.

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
subagent on the panel model. Its low and medium levels are defined as fewer, high-confidence
findings, which is exactly the reading a light panel wants.
**Considered:** The `cavecrew-reviewer` agent — cheapest, but its output is deliberately
compressed, and a committed panel record should read as ordinary prose. The Superpowers
code-reviewer with a narrowed brief — that is the same reviewer slot 0 already runs, so the panel
would carry two readings from one reviewer rather than two independent ones.

### Behaviour when `code-review` is unavailable

**ID:** `code-review-fallback`
**Status:** active
**Chosen:** Substitute a `general-purpose` reviewer with a high-confidence-defects-only brief, and
name the substitution in the panel record.
**Considered:** Dropping the slot — an unavailable harness skill would silently become a two-slot
panel. Falling back to Bugbot — that is the `standard` preset, so a missing skill would silently
convert the operator's choice into a different, more expensive one.

### An unrecorded preset

**ID:** `absent-reads-as-light`
**Status:** active
**Chosen:** Absent reads as `light`, matching the absent-key exception `planningEffort` and `models`
already carry.
**Considered:** Absent reads as `full` — pre-existing changes would keep the review they were
planned under, but every change created before the field existed would then be the expensive case
forever, and the default would be a value nothing ever writes.

### Field name and preset names

**ID:** `field-and-preset-names`
**Status:** active
**Chosen:** `reviewPanelRoster`, with values `light`, `standard` and `full`. Roster and model stay
separate words, so `reviewPanelRoster` cannot be misread as `models.reviewPanel`.
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

## Where each fact is canonical

The split mirrors how the model roles are handled, and no file restates another:

| Fact | Canonical file |
|------|----------------|
| The field's shape, values, absent-key rule, carry-forward duty | `skills/myflow-contracts/state-file.md` |
| What each preset means — roster table, ask mode, per-task shape | `skills/myflow-do/SKILL.md` section 5 |
| The creating-run question and its recommendation | `skills/myflow-start/SKILL.md` |
| Model roles and their defaults | `skills/myflow-contracts/pipeline.md` — **untouched** |

`pipeline.md`'s Model policy section is deliberately not edited. A roster is not a model, and the
panel's model stays `models.reviewPanel`, chosen independently of the roster.

## Risks / Trade-offs

- **A lighter default finds fewer defects.** → The required count stays at three under every
  preset, the conditional triggers still fire and are surfaced rather than silently dropped, and the
  zero-open-findings bar is unchanged. `standard` and `full` are one recorded value away.
- **The `code-review` skill is harness-provided and may vanish or be renamed.** → The stated
  fallback substitutes a `general-purpose` reviewer and names the substitution in the record, so the
  panel never silently shrinks.
- **The subagent's findings arrive as prose, not as structured tool output.** → The dispatch prompt
  requires the findings in the report itself, because the skill's own reporting tool writes to a
  host surface the parent does not read.
- **A multi-select prompt is one more interruption per run under the lighter presets.** → It fires
  only when a trigger actually fired, which the trigger table already keeps rare, and include-all is
  the recommended answer.
- **Reasoned mutation testing has no ground truth and can produce false survivors.** → A survivor is
  an ordinary finding, so the existing handback and withdrawal-with-reason path resolves a disputed
  one exactly as it resolves any other disputed finding.
- **Several contract and skill files grow, and `check-contract-budget.sh` is a ratchet.** → Budget
  rows are raised deliberately as part of this change, which is the response that guard's own
  documentation prescribes for a genuine addition.

## Migration Plan

No migration runs. `reviewPanelRoster` is a new key, absent from every existing state file, and an
absent key reads as `light`. Every change already in flight therefore picks up the new default on
its next `/myflow-do`, and any operator who wants the previous behaviour records `full`.

Rollback is the inverse: removing the field's consumers restores the fixed roster, and a state file
carrying an unread `reviewPanelRoster` key would then be unparseable under the closed-schema rule —
so a rollback removes the key from the field list in the same edit.

## Open questions

*(none — every question raised during planning was answered)*
