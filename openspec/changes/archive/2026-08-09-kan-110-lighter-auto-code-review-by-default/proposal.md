## Why

Every `/myflow-do` run dispatches the same required review panel — Primary, Bugbot and Principles —
plus whichever conditional slots the trigger table auto-includes, and runs a spec-compliance and a
code-quality reviewer on every single task. That roster is fixed, so a documentation-only change
pays what a migration touching auth pays. The roster is a per-change judgment the operator is never
given the chance to make, and the default is the expensive end of it.

## What Changes

- A change records a **review panel roster preset** — `light`, `standard` or `full` — chosen once,
  on the `/myflow-start` run that creates it, alongside the planning effort and the three model
  roles. `light` is the default and the recommended answer.
- `light` runs Primary, Principles and a new **Code review (low)** slot: a `general-purpose`
  subagent on the panel model that invokes the harness's `code-review` skill at effort `low`.
  `standard` runs Primary, Principles and Bugbot. `full` is today's behaviour unchanged — Primary,
  Bugbot and Principles.
- Under `light` and `standard`, the conditional trigger table stops auto-including. Every slot
  whose trigger fired goes into **one** multi-select prompt naming the slot and its trigger, with
  include-all recommended. `full` auto-includes as today.
- Under `light` and `standard`, the per-task spec-compliance and code-quality reviewers collapse to
  **one** combined reviewer per task. `full` keeps both.
- **Bugbot's dispatch brief gains reasoned mutation testing**: for each behaviour the diff changes,
  mutate it and ask whether an existing test fails. A surviving mutant is an ordinary blocking
  finding. No mutation-testing framework is added.
- The state file gains `reviewPanelRoster`. An absent key reads as `light`, joining the absent-key
  exception `planningEffort` and `models` already carry, so no existing state file becomes
  unparseable and no migration pass runs.
- **Not breaking.** `full` reproduces today's panel exactly, and every existing change can reach it
  by recording the preset.

## Capabilities

### New Capabilities

- `myflow-review-panel-roster`: the recorded roster preset — where it is chosen, what each preset's
  roster is, the ask mode for conditional slots, the per-task collapse, the `code-review` slot and
  its fallback, and the rule that no preset moves the handoff bar.

### Modified Capabilities

- `myflow-state-machine`: the state file's field list gains `reviewPanelRoster`, and the absent-key
  exception extends to it.
- `myflow-review-panel-economics`: Bugbot's dispatch brief gains the mutation-testing reading, and
  a surviving mutant is stated to be an ordinary finding under the existing zero-open-findings bar.

## Impact

- `skills/myflow-start/SKILL.md` — a fifth creating-run question; the handoff line reports the
  recorded roster.
- `skills/myflow-do/SKILL.md` — section 4's per-task review shape; section 5's roster table, the
  `code-review` slot, its fallback, the ask mode, and Bugbot's mutation brief; the state write's
  carry-forward list.
- `skills/myflow-contracts/state-file.md` — the `reviewPanelRoster` field and its absent-key rule.
- `skills/myflow-status/SKILL.md` — the `jq` projection and the surfaced line.
- `skills/myflow-finish/SKILL.md` — the carry-forward list.
- `commands/` and `commands-claude/` — the `/myflow-start` and `/myflow-do` command files.
- `CLAUDE.md`, `AGENTS.md`, `README.md` — the command-surface descriptions.
- `scripts/check-contract-budget.sh` — budget rows for every file that grows past its ratchet.

- `scripts/check-vocabulary.sh` — one carve-out, added during implementation. See below.

## Added during implementation

**2026-08-09 — `scripts/check-vocabulary.sh` gains one carve-out.** This proposal originally stated
the guard was not modified, because its list is retired vocabulary and this change retires nothing.
That held until the new slot needed a name. The guard's retired-state list carries
`(^|[^-])\bcode-review\b` — the myflow command retired in the twelve-stage collapse — and that is
byte-for-byte the name of the harness skill the `light` preset's third slot invokes. The guard
already carries a structural carve-out for `requesting-code-review` for exactly this reason; the
harness skill joins it, documented in the guard's own comment block. No other token is added or
removed.

The alternatives were rejected: rewording the skill's name passes the guard while naming a skill
the harness does not have, and a per-line suppression marker is forbidden by this repository's lint
policy. The operator chose the carve-out explicitly.

The guard still constrains the new wording in every other respect — `full-panel` remains a retired
token, so the scanned trees say "the `full` preset" and never that hyphenated form.
