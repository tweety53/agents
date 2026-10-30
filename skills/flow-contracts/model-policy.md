# Model policy

Which model each role runs on, how an override applies, and per-harness enforcement.

**Loaded by `/flow`'s creating run, `/flow`'s implement phase and `/flow-fast`** — at
each implementer and panel dispatch.

This file is **canonical** for everything in it.

The reasoning behind this file lives in `skills/flow-contracts/model-policy-rationale.md`;
**a `/flow*` run never loads it.**

## Model policy

See **Model policy** (`skills/flow-contracts/model-policy-rationale.md`).

Planning — brainstorming, design and writing-plans — runs in the current session, on its own
model; see **Model resolution** (`skills/flow/SKILL.md`), canonical for what the session's model
prices. The parent session runs on the harness's own model for the whole of `/flow`, and **every
review-panel reviewer runs on its dispatch's decision pair** — regardless of the parent model; what
never varies is that the panel's model is *chosen*, not inherited from the parent session.

**Implementer subagents dispatched by `/flow`'s implement phase run on their group's decision
pair**, which **explicitly overrides** superpowers:subagent-driven-development's model
guidance. See **Model policy** (`skills/flow-contracts/model-policy-rationale.md`) for why that
guidance's cost savings do not apply here.

**Two further instructions in that same upstream skill are also overridden: dispatching the final
review on the most capable model, and escalating the model in fix rounds 4-5.** flow fixes every
panel slot at its dispatch's decision pair instead. See **Model policy**
(`skills/flow-contracts/model-policy-rationale.md`) for the reasoning.

**The planner makes the choice.** The implementer, the fixer and each panel dispatch run on the
model and effort the Decide step chose for them, per **Model and effort**
(`skills/flow/brainstorm-planner.md`), canonical for the choice's bounds. The fixer's pair is its
own, not the implementer's; a fix-round re-run runs on the decision's `panel.rerun_dispatch` pair.
Which model a dispatch with no recorded pair runs on — a micro panel, the no-decision dispatch,
the tooling analyst — is **Model and effort** (`skills/flow/brainstorm-planner.md`) as well.

**An explicit operator instruction overrides a decision's pair, in either direction** — raising
the implementer to Opus for a change that warrants it, or lowering it to Sonnet for genuinely
mechanical work. Record the instruction with the dispatch; an override nobody wrote down is indistinguishable
from a mistake.

**A subagent that repairs panel findings is implementer work, so the implementer rule above governs
it too.** See **Model policy** (`skills/flow-contracts/model-policy-rationale.md`) for why.

**A session instruction governs the run in which it is given** and is recorded with its dispatch
exactly as above.

**The ledger records what happened.** A recorded model choice does **not** replace the per-dispatch
ledger line, which remains the only evidence of the model a dispatch actually ran on. Every panel
slot is a prompt-driven role
(**The roster**, `skills/flow/review-panel.md`) and takes its dispatch's model the same way every
other slot does. The findings merger, which reviews nothing, and every gated per-task reviewer
bundle and its fix re-review run on their fixed **Allowed pairs** rows below.

**Every subagent dispatch records the model it used** in the SDD ledger, alongside the task it ran.
See **Model policy** (`skills/flow-contracts/model-policy-rationale.md`) for why, and for the history
behind this rule.

Where the dispatcher **cannot know** the model, the ledger records `unknown (agent-defined)` and
never a guess.

**This record outlives the change.** See **Model policy**
(`skills/flow-contracts/model-policy-rationale.md`) for why, and
**Run 1 — the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`) for the render
duty itself.

**A persisting record must not fill in `unknown (agent-defined)` on the way into the repository** —
neither the write into the store nor the render out of it invents a model slug. See **Model policy**
(`skills/flow-contracts/model-policy-rationale.md`) for why.

- **Claude Code**: a command's `model:` frontmatter (`commands-claude/*.md`) applies only to the
  turn that command starts, never to any turn after it — so it enforces no **session** model for a
  multi-turn run like `/flow`, and no **subagent's** model either. Every dispatched role's model is
  set at dispatch time instead, which every harness supports equally: the implementer and panel
  dispatches each name their model explicitly, and the ledger line for that dispatch is what
  records that they did. Planning has no dispatch to name a model for — it runs on the session's
  own model.
- **ZCode**: one model, see **Harness mapping** below.
- **Muse**: the session's own model, see **Harness mapping** below.

## Allowed pairs

This table is canonical for which `{model, effort}` every flow dispatch may run on: a dispatch
names its role with a `Role: <role id>` line, and a row with a marker admits its pairs only when
the prompt carries that marker as a whole line; a pair is `<model>` and `<subagent_type>` joined
by `/`, an Explore or Plan pair adding `@<effort>`.

Each `###` heading below is one pipeline area, and a role sits under exactly one; a role's
unmarked row carries its `Does` cell, a plain-English line on what the role does, and a marked
row's `Does` is `—`. flowd serves both on `GET /api/v1/roles`, and the Experiments view groups its
roles table by them.

### Planning

| Role | Allowed pairs | Requires marker | Does |
|---|---|---|---|
| `explorer` | `opus/Explore@low`, `opus/Explore@medium` | — | Reads the codebase to answer a planner's question. |
| `plan-designer` | `opus/Plan@low`, `opus/Plan@medium` | — | Proposes a design or plan for the planner to choose from. |
| `planner` | `opus/flow-planner-medium`, `opus/flow-planner-low` | — | Writes the design and the task plan. |

### Implementation

| Role | Allowed pairs | Requires marker | Does |
|---|---|---|---|
| `implementer` | `opus/flow-low`, `opus/flow-medium`, `sonnet/flow-low`, `sonnet/flow-medium` | — | Writes the code for a group of tasks, tests first. |
| `panel-fix` | `opus/flow-low`, `opus/flow-medium`, `sonnet/flow-low`, `sonnet/flow-medium` | — | Fixes what the review panel found. |
| `implementer` | `opus/flow-high` | `No-progress escalation` | — |
| `panel-fix` | `opus/flow-high` | `No-progress escalation` | — |

### Review

| Role | Allowed pairs | Requires marker | Does |
|---|---|---|---|
| `gated-reviewer` | `opus/flow-medium` | — | Reviews risky tasks as they land, on big changes only. |
| `gated-re-reviewer` | `opus/flow-low` | — | Re-checks a gated review's findings after their fix. |
| `panel-reviewer` | `opus/flow-low`, `opus/flow-medium` | — | Reviews the whole branch, first pass: the primary + principles bundle. |
| `panel-reviewer` | `opus/flow-high` | `Operator-approved high` | — |
| `panel-sanity-efficiency` | `opus/flow-low`, `sonnet/flow-low` | — | Always-on extra bundle: checks each new test would fail if its behaviour broke, and spots wasted calls. |
| `panel-mutation-failure` | `opus/flow-low`, `sonnet/flow-low` | — | Always-on extra bundle: breaks the code on purpose to prove the tests notice, and checks error, timeout and partial-write paths. |
| `panel-rerun` | `opus/flow-low`, `sonnet/flow-low` | — | Re-reviews the branch after a fix round. |
| `late-fix` | `opus/flow-low` | — | Reviews just a small fix made after the branch already passed review. |
| `panel-merger` | `sonnet/flow-low` | — | Merges the reviewers' findings into one list without duplicates. |

### Visual verification

| Role | Allowed pairs | Requires marker | Does |
|---|---|---|---|
| `visual-verifier` | `opus/flow-low` | — | Screenshots the UI and compares it with the mockup. |
| `tooling-analyst` | `opus/flow-medium` | — | Works out why visual verification missed a defect. |

### Self-review

| Role | Allowed pairs | Requires marker | Does |
|---|---|---|---|
| `self-review-pass` | `opus/flow-medium` | — | Reviews a finished change and the run that built it. |
| `self-review-plan` | `opus/flow-planner-medium`, `opus/flow-planner-low` | — | Plans the fixes for the self-review's findings. |
| `self-review-fix` | `opus/flow-low` | — | Makes one planned self-review fix. |
| `self-review-fix` | `opus/flow-high` | `No-progress escalation` | — |
| `self-review-review-1` | `opus/flow-medium`, `opus/flow-low` | — | First review of a self-review fix. |
| `self-review-review-r` | `sonnet/flow-low` | — | Re-review once that fix is corrected. |

### Pipeline fix

| Role | Allowed pairs | Requires marker | Does |
|---|---|---|---|
| `pipeline-fix-plan` | `opus/flow-planner-medium`, `opus/flow-planner-low` | — | Plans the fix for a defect found in the flow tooling mid-run. |
| `pipeline-fix-fix` | `opus/flow-low` | — | Makes that tooling fix. |
| `pipeline-fix-fix` | `opus/flow-high` | `No-progress escalation` | — |
| `pipeline-fix-review-1` | `opus/flow-medium`, `opus/flow-low` | — | First review of the tooling fix. |
| `pipeline-fix-review-r` | `sonnet/flow-low` | — | Re-review once that fix is corrected. |

## Experiments

**An open experiment overrides whole cells of the table above.** `flow experiment open -name <n>
-cell '<role>=<pair>[,<pair>…]'…` stores it in flowd; while it is open, every reader of the table
(check-dispatch-pair, check-loop-dispatch, plan-agent-pair, plan-class, pipeline-fix) admits the
experiment's pairs in place of that role's unmarked row, and a marked row is never overridden. One
experiment may carry several roles; a role is in at most one open experiment. Role `main` names
the parent session's `<model>@<effort>`: `/flow` warns on a mismatch (**Model resolution**,
`skills/flow/SKILL.md`) and only matching runs count in its after figures. flowd unreachable → the
table alone. The table stays the permanent source: `/flow-experiment-close` closes an experiment
and, on a positive verdict, starts the change that writes its winning cells here (KAN-1067).

## Harness mapping

**`<parent model>` and `<parent effort>` name the session's own pair, on every harness and in every
inline row.** Where the session does not state one, `<parent model>` is `unknown (agent-defined)` and
`<parent effort>` the literal `default`, the CLI accepting only `low|medium|high|xhigh|max|default`.

**On harness `zcode`, every model a dispatch would be given is `glm-5.3-flash` at effort `high`.**
`VERIFY_MODEL`, the literal `opus` of a dispatch with no recorded pair, a decision's implementer,
fixer, group and panel pairs, and an operator override alike resolve and are recorded as they would be on Claude Code,
and are replaced at the dispatch: the Agent tool's `model` parameter is `glm-5.3-flash`, the
`subagent_type` is `general-purpose` (or the site's own non-flow type, unchanged) — ZCode's Agent tool offers no
`flow-<effort>` type, `<agents repo>/setup.sh` installing those definitions into `~/.claude/agents/` alone, so the NO
DELEGATION paragraph every such prompt carries is what keeps the dispatch from forking — and the dispatch's
ledger line records `-model glm-5.3-flash -effort high` — the model the dispatch actually ran
on, never the pre-mapping value. A reply's `Model:` line is not compared on this harness — the recorded mapping satisfies the handshake, per the paragraph below. The harness is
the same value the run's `-harness` marks carry (**Stage marks**,
`skills/flow-contracts/pipeline.md`). No other harness maps anything to a fixed pair.

**On harness `muse`, every dispatch inherits the session's own model and effort.** The harness
offers no per-dispatch model parameter and no `flow-<effort>` definitions, so the dispatch names
neither — no `model` parameter, no `subagent_type`. `VERIFY_MODEL`, the literal `opus` of a
dispatch with no recorded pair, a decision's implementer, fixer, group and panel pairs, and an
operator override alike resolve and are recorded as they would be on Claude Code, and are replaced
at the dispatch: the dispatch runs the session's pair, and its ledger line records `-model <parent
model> -effort <parent effort>` — the model the dispatch actually ran on, never the pre-mapping
value. A reply's `Model:` line
is not compared on this harness either — the recorded mapping satisfies the handshake, per the
paragraph below.

**On a single-model harness the recorded mapping satisfies the handshake.** Where the harness maps
every dispatch to one recorded model that no re-dispatch can change — harness `zcode`, the fixed
pair above, and harness `muse`, the session's own pair — a first reply whose
`Model:` line is missing or names anything else is not a mismatch: no `-outcome fallback`, no
`<key>-retry`, no second-mismatch question. The MODEL HANDSHAKE paragraph stays in every
dispatch prompt, and the comparison governs in full on every harness no mapping covers.
