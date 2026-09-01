# Design — kan-374-flow-run-planning-stages-on-fable-opus-fallback

## Context

Bounded change; no separate spec file. `proposal.md` carries why; this file carries how. Every
planning stage of a creating `/flow` run executes in the parent session today, on whatever model
that session runs, and the only model-setting mechanism reachable from inside a run is the model a
subagent is dispatched with.

## The planner protocol

The parent `/flow` session keeps section A of `skills/flow/brainstorm.md` (name, Jira transition,
`STARTED`), the guard-presence check, every stage mark, the dispatch record and the operator
conversation. One planner subagent, dispatched with the Agent tool's `model` set to
`PLANNING_MODEL`, owns sections B, C and D and is resumed with `SendMessage` between rounds so its
context carries from the first question to the finished plan.

- **Handshake.** The planner's first reply opens with `Model: <name from its own system prompt>`.
  The parent compares it with `PLANNING_MODEL`; on a mismatch it ends the dispatch record with
  `-outcome fallback` and re-dispatches once with `model: opus`; a second mismatch proceeds on
  whatever answered, reports it, and records that model. Never blocks.
- **Question relay.** A subagent cannot reach the operator, so the planner ends a turn with exactly
  one `## Question` block (the question plus named options when it has them). The parent asks it
  verbatim through the harness's question tool and resumes the planner with the answer. The
  convergence confirm, the third-round offer and the design approval are the same relay.
- **Three returns.** After brainstorming converges the planner returns `## Design` (the full design
  text) and the parent runs the convergence confirm and the design-approval gate. After section C it
  returns `## Artifacts` with the three paths. After section D it returns `## Plan` with the guard
  results. The parent marks `flow.brainstorm`, `flow.design-approval`, `flow.create-artifacts` and
  `flow.writing-plans` around those returns.
- **Record.** `flow record dispatch begin -role planner -model <PLANNING_MODEL> -key planner` goes out
  before the dispatch; `end` follows the `## Plan` return with `-agent-id`. Per-model token metrics
  are already harvested per stage run, so nothing else is stored.
- **Fix runs.** `flow.document-fix` dispatches the same planner (`-key planner-fix-<n>`) with the fix
  instructions; the "where should the fix go" prompt stays a relay.
- **Research.** `/flow-research` dispatches a research subagent on `PLANNING_MODEL` with the
  research skill as its instructions; the subagent writes the staging note itself. No dispatch
  record: research has no change to record against.

## Resolution

```bash unverified:confirm jq's handling of a missing key yields the empty string, not "null"
SETTINGS_JSON="$(flow settings get)"
PLANNING_MODEL="$(printf '%s' "$SETTINGS_JSON" | jq -r '.planningModel // empty')"
# <project>/.flow/project.md's `## planning model` body, when present and a ValidModels member, wins
[ -z "$PLANNING_MODEL" ] && PLANNING_MODEL=fable
```

## Decisions

### Planning moves into a subagent, not a session-model switch

**ID:** planner-subagent
**Status:** active
**Chosen:** one planner subagent on `PLANNING_MODEL`, parent relays — the only mechanism that sets
a model from inside a run, in every harness.
**Considered:** `model: fable` frontmatter on `/flow` — covers one turn only, and the docs confirm
no agent-side switch exists. Printing `/model fable` for the operator — unenforced and unrecorded.
Delegating writing-plans alone — leaves brainstorming, the stage that benefits most, uncovered.

### The planning model is a settings-store field with a project override

**ID:** planning-model-setting
**Status:** active
**Chosen:** `planningModel` in the settings store (empty resolves to `fable`), overridable per
project by `## planning model` in `<project>/.flow/project.md` and per run by a session instruction.
**Considered:** a hardcoded `fable` in the skill — the operator asked for a setting a project can
override at any time.

### The project key is a single-line literal

**ID:** project-override-key
**Status:** active
**Chosen:** `## planning model` holds one `ValidModels` member and nothing else, matched after
trimming, like `## default landing route`; an invalid body is reported by name and dropped.
**Considered:** a per-project settings row in the store — the store is harness-wide by the
`settings-scope` decision and `project.md` already carries per-project keys.

### Fallback to opus is verified by the planner's own identity line

**ID:** opus-fallback-verified
**Status:** active
**Chosen:** the planner reports its model first; a mismatch re-dispatches once on `opus`.
**Considered:** trusting the dispatch parameter — the harness silently substitutes the inherited
model when the requested one is blocked, so the record would lie.

### Dispatch records gain a `planner` role

**ID:** planner-role
**Status:** active
**Chosen:** add `planner` to the CLI's closed role list so the dispatch is recorded like an
implementer's.
**Considered:** recording under `implementer` — a false role in the audit trail. Not recording —
model-policy's own rule: an unrecorded model is indistinguishable from a mistake.

### `flow.resolve-change` folds into `flow.kickoff`

**ID:** resolve-change-collapse
**Status:** active
**Chosen:** remove the key everywhere it is written; keep historical rows.
**Considered:** keeping the row as a truthful zero — it carries nothing `flow.kickoff` does not.

### `flow.document-fix` runs on the planner

**ID:** document-fix-on-planner
**Status:** active
**Chosen:** a fix run's proposal and plan rewrite is a planner dispatch — operator decision.
**Considered:** keeping it in the parent as a short edit — rejected by the operator.

### `/flow-research` runs on the planning model

**ID:** research-on-planner
**Status:** active
**Chosen:** a research subagent on `PLANNING_MODEL`, same relay shape, one setting for both.
**Considered:** frontmatter on the research command — the same one-turn limit.

## Open questions

None recorded.
