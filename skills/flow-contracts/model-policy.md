# Model policy

Which model each role runs on, how an override applies, and per-harness enforcement.

**Loaded by `/flow`'s creating run, and `/flow`'s implement phase** — at
each dispatch.

This file is **canonical** for everything in it.

The reasoning behind this file lives in `docs/rationale/flow-contracts/model-policy.md`.

## Model policy

**Every dispatch runs on its role's pair**: the pair `flow-guard pair-judge` picks from the role's
allowed pairs at dispatch time, or, for a role the judge does not cover, the role's first pair
(**Model and effort**, below). Nothing records a pair ahead of the
dispatch. Half of the implementer's `opus/flow-low` picks are moved to `sonnet/flow-low` by
the judge itself, so the two split that case evenly. The parent
session runs on the harness's own model for the whole of `/flow`. This **explicitly overrides**
superpowers:subagent-driven-development's model guidance, its final review on the most capable
model and its model escalation in fix rounds 4–5.

**An explicit operator instruction overrides a role's pair, in either direction**, for the run in
which it is given — raising the implementer to Opus for a change that warrants it, or lowering it
to Sonnet for genuinely mechanical work. The dispatch hook records the pair each dispatch
actually ran on.

**Every subagent dispatch records the model it used** in the SDD ledger, alongside the task it ran.

Where the dispatcher **cannot know** the model, the ledger records `unknown (agent-defined)` and
never a guess.

**This record outlives the change.** See **Run 1 — the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`) for the render duty itself.

**A persisting record must not fill in `unknown (agent-defined)` on the way into the repository** —
neither the write into the store nor the render out of it invents a model slug.

- **Claude Code**: a command's `model:` frontmatter (`commands-claude/*.md`) applies only to the
  turn that command starts, never to any turn after it — so it enforces no **session** model for a
  multi-turn run like `/flow`, and no **subagent's** model either. Every dispatched role's model is
  set at dispatch time instead, which every harness supports equally: the implementer and panel
  dispatches each name their model explicitly, and the ledger line for that dispatch is what
  records that they did.

## Model and effort

**A dispatch's pair is picked by the pair judge**, at dispatch time: write the dispatch's prompt
to `<abs-worktree>/.superpowers/sdd/pair-judge/<key>.md`, `<key>` the dispatch's key or, for an
Explore or Plan agent, `explore-<n>` / `design-<n>`, then

```bash
flow-guard pair-judge <role> <abs-worktree>/.superpowers/sdd/pair-judge/<key>.md
```

It prints one line, `<model> <effort> <judge|fallback|single> <reason>`, and saves it beside the
prompt as `<key>.pick`, which the dispatch hook records with the row. The judge picks only among
the role's allowed pairs in flowd; no pair is ever written into a prompt or a skill. A judge
that cannot answer never blocks: it prints the role's first candidate with source `fallback`. A
role the judge does not cover (it exits 2) runs on its first `base` pair in `flow policy get`.
A role with exactly one `base` pair needs no judge call: the hook asks for no `.pick`, records
`-pair-pick single: the only candidate` on the row, and the dispatch runs on that pair. Besides the prompt, the judge reads only the file parts it names: `` `<abs-path>#<heading text>` ``
(in backticks; that markdown section) or `<abs-path>:<a>-<b>` / `:<a>` (those lines); a bare
`tasks.md` gives the sections of the tasks the prompt names, and any other bare path nothing.

A pair is a model and an agent type joined by `/`: `opus/flow-low` dispatches
`subagent_type: flow-low` on `model: opus`; `opus/flow-planner-medium` is `flow-planner-medium` on
`opus`; `opus/Explore@low` is `subagent_type: Explore` with `effort: low`. Every dispatch has a
role from **Allowed pairs** (below). A plain-language session instruction overrides the pair it names
for this run only (**Model policy**, above).

## Allowed pairs

**flowd's store is canonical for which `{model, effort}` every flow dispatch may run on**: each role's pairs live there, read with `flow policy get` and edited by hand in the
Agent roles view; the tables below name the roles only. A dispatch names its role with a
`Role: <role id>` line; a pair is `<model>` and `<subagent_type>` joined by `/`, an Explore or Plan
pair adding `@<effort>`. Any well-formed pair is accepted — which models a role may use is the
operator's call, made in the editor. A role's pairs come in two tiers:

- **base** — always admitted.
- **opt-in** — admitted only on the operator's explicit say-so for a change: the operator tells
  the session which roles may use them, the session records it with
  `flow policy grant -change <name> -roles <role>[,<role>…]`, and each such dispatch's prompt
  carries the whole line `Operator pair: <role>`. `flow policy revoke` removes a grant, and every
  grant lapses when its change reaches `FINISHED`.

`flow policy get` caches its last good answer under the flow state directory; with flowd down it
answers from that cache, and with neither a dispatch runs on the session's own pair (`<parent model>` / `<parent effort>`, **Harness mapping** below).

Each `###` heading below is the pipeline stage a role runs in (`flow stage keys`; self-review's,
the fix loop's and pair choice's dispatches have their own), and a role sits under exactly one, with
its `Does` cell, a plain-English line on what the role does. flowd serves both, with the role's
pairs, on `GET /api/v1/roles`, and the Agent roles view groups its roles table by them.

### Planning

| Role | Does |
|---|---|
| `explorer` | Reads the codebase to answer a planner's question. |
| `plan-designer` | Proposes a design or plan for the planner to choose from. |
| `designer` | Draws, in a worktree of the project's design repo and in its own style, every element the planner's design lists as undrawn. |
| `planner` | Writes the design and the task plan. |
| `decider` | Reviews the approved plan and decides the implementer groups, the review panel's bundles and visual verification (headless, `plan-check`). |

### Implementation

| Role | Does |
|---|---|
| `implementer` | Writes the code for a group of tasks, tests first. |
| `gated-reviewer` | Reviews each implementer group's tasks as they land. |
| `gated-re-reviewer` | Re-checks a gated review's findings after their fix. |
| `no-progress-fixer` | Takes over a fix loop that made no progress — any loop's fixer, with every earlier round's report. |
| `operator-asked` | Runs a task the operator asked a subagent for, in their own words, inside a /flow-interactive request. |

### Review

| Role | Does |
|---|---|
| `panel-reviewer` | Reviews the whole branch, first pass. |
| `panel-rerun` | Re-reviews the branch after a fix round. |
| `panel-fix` | Fixes what the review panel found. |
| `minor-fix` | Fixes a review round's Minors when the round raised nothing worse. |

### Verification

| Role | Does |
|---|---|
| `visual-verifier` | Screenshots the UI and compares it with the mockup. |

### Self-review

| Role | Does |
|---|---|
| `self-review-pass` | Reviews a finished change and the run that built it. |

### Fix loop

| Role | Does |
|---|---|
| `loop-plan` | Plans the fixes for the self-review's findings, or for a defect found in the flow tooling mid-run. |
| `loop-fix` | Makes one planned fix. |
| `loop-review-1` | First review of that fix. |
| `loop-review-r` | Re-review once that fix is corrected. |

### General

| Role | Does |
|---|---|
| `general` | Triages a failing test `flow-guard run-verify-list` marks `undecided`, and any task no other role fits; the operator adds pairs in the Agent roles view. |

### Pair choice

| Role | Does |
|---|---|
| `pair-judge` | Picks each dispatch's model and effort from its role's allowed pairs. |

### Choosing a role

**Never invent a role name. Every `flow-*` dispatch's prompt carries a `Role: <id>` line naming a role above, on a pair that role admits** — `hooks/enforce-agent-baseline.py` denies one that does not. A task no role fits takes `general`.

| Situation | Role |
|---|---|
| A conflict or red check after a rebase or base sync, fixed outside any loop | `minor-fix` |
| A fix loop that made no progress | `no-progress-fixer` |
| Screenshot recapture, browser driving or a capture run | `visual-verifier` |
| Triaging a failing test `flow-guard run-verify-list` marks `undecided` | `general` |
| Anything else | `general` |

## Harness mapping

**`<parent model>` and `<parent effort>` name the session's own pair, on every harness and in every
inline row.** Where the session does not state one, `<parent model>` is `unknown (agent-defined)`.
No dispatch runs at a model's built-in effort: its effort is its `flow-<effort>` /
`flow-planner-<effort>` type's, or the Agent call's `effort` on any other type, and the dispatch
hook (`hooks/enforce-agent-baseline.py`) refuses a flow-session Agent call carrying neither.

