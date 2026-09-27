# Design

## Context

KAN-820 (flow-cost): the operator's model policy lives in memory and narratives, not in
anything the dispatcher reads; each mismatch costs a wasted dispatch plus a prose correction.
One change because the policy's home, its resolution order and its surfacing are one mechanism
split across five contract/skill files plus this repo's own project config — editing any
subset would leave the dispatcher reading two disagreeing sources.

Constraints: the settings store stays harness-wide and schema-frozen; `VERIFY_MODEL` stays a
fixed literal; `SELF_REVIEW_MODEL` keeps its own key and precedence; the zcode harness mapping
replaces every dispatch's model at dispatch and is not changed; a run-scoped operator
instruction still overrides in either direction and stays recorded; projects declaring nothing
must resolve exactly as today.

## Approach

One new literal-body key, `## model`, in `<project>/.flow/project.md`, and a new first step in
`DEFAULT_MODEL`'s resolution. Everything downstream of the resolution — effort selection, the
review panel's shape, the override rule, the zcode mapping — is unchanged. The resolution is
surfaced (preamble line, run summary) and recorded (`decision.json` `resolved` object) so the
pre-dispatch value and its provenance are machine-readable, per the finding's ask.

The key is validated exactly like `## self review model`, the precedent: one member of the
settings store's `ValidModels`, body matched byte-for-byte with leading/trailing whitespace
trimmed and nothing else normalized; a body matching no member is reported by name and dropped,
resolving as if the key were absent. `flow-settings` keeps governing the store only — the
project key is edited in the project's own committed config, which is what makes it
machine-readable to the dispatcher and reviewable in pull requests.

Because the key lives in the repository, it stays readable when the settings store is
unreachable: a project's declared policy survives a store outage, and the fallback chain
(project → store → literal `opus`) degrades in the same order of specificity.

## Decisions

### The per-project model policy lives in a project.md key

**ID:** project-key-home
**Status:** active
**Chosen:** `## model` in `<project>/.flow/project.md`, one `ValidModels` member, validated like `## self review model` — in-repo, dispatcher-readable, no store migration
**Considered:** settings-store per-project rows — a store schema change plus an authoring surface for exactly what a committed project file already carries; per-change state field `models.default` — policy becomes per-change, every change re-decides it, and `/flow` deliberately never asks

### The project key is the first source; store default and the literal follow

**ID:** resolution-precedence
**Status:** active
**Chosen:** project `## model` → store `defaultModel` → literal `opus`, resolved once per run — the most specific policy wins, and the key still resolves during a store outage
**Considered:** store-first — a harness-wide default would shadow every project's policy, making the key dead on any machine with a store row

### The key governs every role that reads DEFAULT_MODEL today

**ID:** governed-roles
**Status:** active
**Chosen:** implementer, fixer, every panel dispatch, the rerun pair — the resolution's source changes, the consumer set does not; `VERIFY_MODEL` stays the fixed literal, `SELF_REVIEW_MODEL` keeps its own key and precedence, effort stays the Decide step's
**Considered:** implementer+fixer only — matches the incident narrowly but splits one resolution into two sources; per-role key shape — a parsed multi-row section and more resolution rules for the same fix

### A policy violation is surfaced and recorded, never gated

**ID:** enforcement-strength
**Status:** active
**Chosen:** resolution names model + source once; the Decide preamble `models:` line and the run summary carry both; `decision.json` gains `resolved: {"model": …, "source": …}` — visible before dispatches go out, nothing blocks
**Considered:** hard pre-dispatch gate — on zcode it could only ever key on the pre-mapping value, blocking runs over a value the harness replaces anyway; lint-time guard — sees only this repository, not the consumer projects the policy is for

### The key governs the pre-mapping value on harnesses that map

**ID:** mapping-interplay
**Status:** active
**Chosen:** the zcode mapping still replaces every dispatch's model at dispatch; resolution and the decision/summary records name the resolved policy model, and the dispatch ledger keeps recording the mapped model actually run — no contract change to the mapping
**Considered:** key inert on mapping harnesses — honest about power but the policy never bites on this machine; policy declares the mapped model — makes the key per-harness in practice, defeating "records the operator's policy"

## Open questions

## Consequences

- Projects declaring nothing resolve exactly as today — store default, then the literal; no
  behavior change anywhere the key is absent.
- The mismatch cost the finding names (wasted dispatch, operator correction) becomes visible at
  resolution time: the run states its model and where it came from before anything dispatches,
  and an operator instruction that overrides the key is recorded beside a named source instead
  of riding in prose alone.
- This repository's own `.flow/project.md` gains `## model opus`, which equals the store
  default here — the path is exercised by every subsequent run of this repo without changing
  which model resolves.
