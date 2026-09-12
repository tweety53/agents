# Model policy

Which model each role runs on, their defaults, how an override applies, and per-harness
enforcement.

**Loaded by `/flow`'s creating run, `/flow`'s implement phase and `/flow-fast`** — at the model questions and at
each implementer and panel dispatch.

This file is **canonical** for everything in it.

The reasoning behind this file lives in `skills/flow-contracts/model-policy-rationale.md`;
**a `/flow*` run never loads it.**

## Model policy

See **Model policy** (`skills/flow-contracts/model-policy-rationale.md`).

Planning — brainstorming, design and writing-plans — runs in the current session, on its own
model, with no dispatched planner subagent and no `PLANNING_MODEL` to resolve (kan-488); see
**Model resolution** (`skills/flow/SKILL.md`), canonical for what the session's model prices. The
parent session itself runs on **Sonnet** (or the harness's standard default) for the whole of
`/flow`, and **every review-panel reviewer runs on the panel's model — Sonnet by default** —
regardless of the parent model. Sonnet is the default rather than an absolute because a change may
record its own panel model, per the three roles below; what never varies is that the panel's model
is *chosen*, not inherited from the parent session.

**Implementer subagents dispatched by `/flow`'s implement phase run on Opus** (or the harness's strongest
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

**The model each role runs on is resolved per run from the settings store, never recorded per
change.** `DEFAULT_MODEL` and `REVIEWERS` resolve once near the top of every run, per **Model
resolution** (`skills/flow/SKILL.md`). The state file's `models.default` records only a model an
operator explicitly chose for the change; it is `null` on every change the creating run did not ask
about. See **State file** (`skills/flow-contracts/state-file.md`).

**A subagent that repairs panel findings is implementer work, so the implementer rule above governs
it too.** See **Model policy** (`skills/flow-contracts/model-policy-rationale.md`) for why.

**A session instruction governs the run in which it is given** and is recorded with its dispatch
exactly as above.

**These fields record intent; the ledger records what happened.** A recorded value does **not**
replace the per-dispatch ledger line, which remains the only evidence of the model a dispatch
actually ran on. Every panel slot, Bugbot and Security included, is a prompt-driven role dispatched
general-purpose (**The roster**, `skills/flow/review-panel.md`) and takes the recorded panel model
the same way every other slot does — there is no exception left that leaves a slot's ledger entry
unrecorded.

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
  own model (kan-488).
- **Cursor**: no per-command model frontmatter, so no model is enforceable from a command file —
  each `.cursor/commands/flow*.md` carries an explicit note; switch models manually in the
  composer/chat picker.
- **Codex**: no per-command/skill model override mechanism either — model is a session or profile
  level setting; switch manually before starting a new proposal.
