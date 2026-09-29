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
panel slot at its dispatch's decision pair instead and escalates breadth (the conditional Failure-modes and
Mutation slots) rather than the model. See **Model policy**
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
other slot does.

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

## Harness mapping

**On harness `zcode`, every model a dispatch would be given is `glm-5.3-flash` at effort `high`.**
`VERIFY_MODEL`, the literal `opus` of a dispatch with no recorded pair, a decision's implementer,
fixer, group and panel pairs, and an operator override alike resolve and are recorded as they would be on Claude Code,
and are replaced at the dispatch: the Agent tool's `model` parameter is `glm-5.3-flash`, the
`subagent_type` is `flow-high` (or the site's own non-flow type, unchanged), and the dispatch's
ledger line records `-model glm-5.3-flash -effort high` — the model the dispatch actually ran
on, never the pre-mapping value. A reply's `Model:` line is not compared on this harness — the recorded mapping satisfies the handshake, per **The handshake** (`skills/flow/implement.md`, **The parent orchestrates directly**). The harness is
the same value the run's `-harness` marks carry (**Stage marks**,
`skills/flow-contracts/pipeline.md`). No other harness maps anything.

**On a single-model harness the recorded mapping satisfies the handshake.** Where the harness maps
every dispatch to one recorded model that no re-dispatch can change — harness `zcode`, the one
mapping (**Harness mapping**, `skills/flow-contracts/model-policy.md`) — a first reply whose
`Model:` line is missing or names anything else is not a mismatch: no `-outcome fallback`, no
`<key>-retry`, no second-mismatch question. The MODEL HANDSHAKE paragraph stays in every
dispatch prompt, and the comparison governs in full on every harness no mapping covers.
