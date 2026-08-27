# The panel roster follows the settings list — design

**Change:** `panel-roster-follows-the-settings-list`
**Jira:** none — run unlinked, by the operator's choice
**Date:** 2026-08-27

## Why

The settings store holds a `reviewers` list. `/flow-settings` writes it. **No run reads it.**

```bash verified:grepped skills/ at 1717d19 — SKILL.md:107-108 are the only settings reads, and neither takes .reviewers
SETTINGS_JSON="$(flow settings get)"
DEFAULT_MODEL="$(printf '%s' "$SETTINGS_JSON" | jq -r '.defaultModel')"
```

The roster instead lives as prose in `skills/flow/review-panel.md` — three required slots, plus
Bugbot and Security dispatched only on an explicit per-run instruction. Two consequences, both bad:
changing the roster means editing a contract file, and a `/flow-settings` write that appears to
change the roster does nothing at all.

## Resolution

**The resolved list is the roster.** `skills/flow/SKILL.md` reads `.reviewers` in the same step that
already reads `.defaultModel`, and the panel dispatches exactly what it resolves.

| Store state | Resolved roster |
|-------------|-----------------|
| Reachable, list non-empty | exactly the list |
| Reachable, list empty | `primary` alone |
| Unreachable | `primary`, `principles`, `code-review-low` — reported as a **fallback**, never as a resolved value |

**The empty-list row is a defensive floor, not a reachable configuration.**
`stats/cmd/flow/settings.go:123` refuses an empty `-reviewers` as a caller mistake, so
`/flow-settings` cannot write one. The row exists because the resolver must define what it does with
a value the store could hold, not because an operator can produce it. **This is stated so nobody
reads the table as an offer to turn the panel off.**

**A per-run operator instruction still adds a slot for that run**, from the round it was made and
never retroactively to a closed pass, and is never written back to the store. That mechanism is
unchanged.

**Each slot keeps its own spawn shape.** Bugbot and Security are dispatched by `subagent_type`,
carry their own agent definitions, take **no** model override, and record `unknown (agent-defined)`
as their dispatch row's `-model`. The list decides *which* slots run; it does not flatten *how* they
are spawned.

## Decisions

### The roster is resolved from the settings store, not written in prose

**ID:** `roster-from-settings`
**Status:** active
**Chosen:** `skills/flow/SKILL.md` resolves `.reviewers` alongside `.defaultModel`; the panel
dispatches the resolved list. `/flow-settings` becomes the one place a roster changes, and no
contract file is edited per roster update.
**Considered:** keeping the roster in prose and having `/flow-settings` refuse to write `reviewers`
at all — honest about today's behaviour and rejected because it removes a capability rather than
connecting one, leaving every roster change a contract edit; a second config file naming the roster
— rejected as a third source of truth beside the store and the prose.

### `review-panel-fixed-3` is superseded

**ID:** `supersedes-review-panel-fixed-3`
**Status:** active
**Chosen:** the fixed three-required-plus-two-on-demand roster is replaced by the resolved list.
**Considered:** keeping the fixed three as an unremovable floor with the list adding only Bugbot and
Security — rejected by the operator, who chose the list as the whole truth.
**Supersedes:** `review-panel-fixed-3`, recorded in the archived `kan-326-myflow-rework-redesign`
change's `design.md`. That entry is **not** rewritten — it belongs to a closed change, and this
entry names it instead, which is the same relationship a superseding decision has to its predecessor
within one file.

### An empty list floors at `primary`, rather than dispatching nothing

**ID:** `empty-list-floors-at-primary`
**Status:** active
**Chosen:** a reachable store holding an empty list resolves to `primary` alone.
**Considered:** dispatching zero reviewers and saying so loudly in the handoff — rejected by the
operator; dispatching zero silently — rejected outright, since a misconfigured store would then ship
unreviewed work with nothing drawing attention to it; refusing to write an empty list — already the
CLI's behaviour, and therefore not a resolver rule.

### An unreachable store falls back to the store's own three defaults

**ID:** `unreachable-falls-back-to-defaults`
**Status:** active
**Chosen:** fall back to `primary`, `principles`, `code-review-low` — `DefaultReviewers` in
`stats/internal/store/settings.go` — and name it as a fallback in the handoff.
**Considered:** blocking the run — rejected, the pipeline's stated rule is that an unreachable
settings store never blocks implementation; treating unreachable as empty and flooring at `primary`
— rejected, an outage would silently reduce review rather than preserving the normal roster.

### An unspawnable id is substituted, not skipped, and must mutation-test

**ID:** `unspawnable-id-substitutes`
**Status:** active
**Chosen:** a resolved id whose agent type the running harness does not offer is dispatched as a
general-purpose subagent carrying that slot's brief, and that substitute **must perform mutation
testing** — changing the code to prove each finding is real, and reverting afterwards. The
substitution is recorded as a substitution: the dispatch row names the slot it stood in for, and
`-model` records the model actually given rather than `unknown (agent-defined)`.
**Considered:** reporting the id and running the rest — rejected by the operator, because it lets
the harness silently shrink the panel; blocking the run — rejected, a harness without one agent type
installed could then never run a panel at all.
**The objection, recorded because it was raised and overruled:** a substituted slot recorded as the
real one corrupts the record of what actually reviewed a branch. The operator's mutation-testing
requirement answers the *rigour* half of that — a general-purpose agent that must prove each finding
is not obviously weaker than the real slot — and the recording rule above answers the *honesty*
half. Neither is optional; the substitution is only sound with both.
**Reachable today, not hypothetical:** this session's Claude Code offers no `bugbot` agent type
while the store's list carries `bugbot`.

## Open questions

None.
