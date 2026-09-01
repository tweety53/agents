# Model policy

Which model each role runs on, their defaults, how an override applies, and per-harness
enforcement.

**Loaded by `/myflow-start`, `/myflow-do` and `/myflow-fast`** — at the model questions and at
each implementer and panel dispatch.

This file is **canonical** for everything in it.

The reasoning behind this file lives in `skills/flow-contracts/model-policy-rationale.md`;
**a `/flow*` run never loads it.**

## Model policy

See **Model policy** (`skills/flow-contracts/model-policy-rationale.md`).

**There is no requirements layer above this one; change this section.** See **Model policy**
(`skills/flow-contracts/model-policy-rationale.md`) for where the rule was first written and why
that layer no longer governs.

Planning — brainstorming, design and writing-plans — runs in a planner subagent dispatched on
`PLANNING_MODEL` (`fable` by default, `opus` on a verified fallback), never on the parent session;
see **Model resolution** (`skills/flow/SKILL.md`), canonical for the resolution and the dispatch.
The parent session itself runs on **Sonnet** (or the harness's standard default) for the whole of
`/flow`, and **every review-panel reviewer runs on the panel's model — Sonnet by default** —
regardless of the parent model. Sonnet is the default rather than an absolute because a change may
record its own panel model, per the three roles below; what never varies is that the panel's model
is *chosen*, not inherited from the parent session.

**Implementer subagents dispatched by `/myflow-do` run on Opus** (or the harness's strongest
available model), which **explicitly overrides** superpowers:subagent-driven-development's model
guidance. See **Model policy** (`skills/flow-contracts/model-policy-rationale.md`) for why that
guidance's cost savings do not apply here.

**Two further instructions in that same upstream skill are also overridden: dispatching the final
review on the most capable model, and escalating the model in fix rounds 4-5.** flow fixes every
panel slot at the panel's model instead and escalates breadth (the conditional Security, Adversarial
and extra-principle slots) rather than the model. See **Model policy**
(`skills/flow-contracts/model-policy-rationale.md`) for the reasoning.

**An explicit operator instruction overrides either default, in either direction** — raising the
panel to Opus for a change that warrants it, or lowering the implementer for genuinely mechanical
work. Record the instruction with the dispatch; an override nobody wrote down is indistinguishable
from a mistake.

**Three model roles are chosen once per change and recorded in its state file.** The run that
**creates** a change — `/myflow-start` finding no state file, exactly as the planning-effort
question determines it — asks three separate questions, one per role, each naming its default and
marking it as the recommendation. A revision round states the recorded values and does not ask
again, and every other command carries them forward verbatim, as it does the linked Jira issue.

| Role | Key under `models` | Default |
|------|--------------------|---------|
| The implementer subagents `/myflow-do` dispatches | `implementation` | Opus, or the harness's strongest available model |
| Every review-panel slot that takes a model override | `reviewPanel` | Sonnet |
| The subagents that repair panel findings | `panelFix` | Opus, or the harness's strongest available model |

See **State file** (`skills/flow-contracts/state-file.md`).

**The panel-fix default is the strongest available model, and deliberately not Sonnet** — the role
applies fixes, which is implementer work, so the implementer rule above governs it too. See **Model
policy** (`skills/flow-contracts/model-policy-rationale.md`) for why.

**A recorded choice is the operator override this section already permits, made durable.** It
applies to every run of the change without being restated, which is the point of recording it. A
session instruction is the narrower and later of the two: it governs the run in which it is given
and is recorded with its dispatch exactly as above.

**These fields record intent; the ledger records what happened.** A recorded value does **not**
replace the per-dispatch ledger line, which remains the only evidence of the model a dispatch
actually ran on. Slots dispatched by `subagent_type` take no override from this mechanism either —
no recorded panel model is passed to them, none is written for them in the ledger, and their entries
still read `unknown (agent-defined)`.

**Every subagent dispatch records the model it used** in the SDD ledger, alongside the task it ran.
See **Model policy** (`skills/flow-contracts/model-policy-rationale.md`) for why, and for the history
behind this rule.

Where the dispatcher **cannot know** the model, the ledger records `unknown (agent-defined)` and
never a guess. Slots dispatched by `subagent_type` (Bugbot, Security Review) resolve their model
from their own agent definition, which the dispatcher does not read; writing a plausible slug for
them puts an unmeasured value into the audit trail.

**This record outlives the change.** See **Model policy**
(`skills/flow-contracts/model-policy-rationale.md`) for why, and
**Run 1 — the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`) for the render
duty itself.

**A persisting record must not fill in `unknown (agent-defined)` on the way into the repository** —
neither the write into the store nor the render out of it invents a model slug. See **Model policy**
(`skills/flow-contracts/model-policy-rationale.md`) for why.

- **Claude Code**: a command's `model:` frontmatter (`commands-claude/*.md`) applies only to the
  turn that command starts, never to any turn after it — so it enforces no **session** model for a
  multi-turn run like `/flow`, and no **subagent's** model either. Every role's model is set at
  dispatch time instead, which every harness supports equally: the implementer, panel and planner
  dispatches each name their model explicitly, and the ledger line for that dispatch is what
  records that they did.
- **Cursor**: not enforceable yet (no per-command model frontmatter support as of this writing) —
  each `.cursor/commands/flow*.md` file carries an explicit note; switch models manually in the
  composer/chat picker.
- **Codex**: no per-command/skill model override mechanism either — model is a session or profile
  level setting; switch manually before starting a new proposal.
