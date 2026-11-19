# Model policy

Which model each role runs on, how an override applies, and per-harness enforcement.

**Loaded by `/flow`'s creating run, `/flow`'s implement phase and `/flow-fast`** — at
each dispatch.

This file is **canonical** for everything in it.

The reasoning behind this file lives in `skills/flow-contracts/model-policy-rationale.md`;
**a `/flow*` run never loads it.**

## Model policy

**Every dispatch runs on its role's pair**: the pair `flow-guard pair-judge` picks from the role's
allowed pairs at dispatch time, or, for a role the judge does not cover, the role's first pair
(**Model and effort**, `skills/flow/brainstorm-planner.md`). Nothing records a pair ahead of the
dispatch. The parent
session runs on the harness's own model for the whole of `/flow`. This **explicitly overrides**
superpowers:subagent-driven-development's model guidance, its final review on the most capable
model and its model escalation in fix rounds 4–5.

**An explicit operator instruction overrides a role's pair, in either direction**, for the run in
which it is given — raising the implementer to Opus for a change that warrants it, or lowering it
to Sonnet for genuinely mechanical work. Record it in the decision's `overrides`; an override
nobody wrote down is indistinguishable from a mistake.

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
  records that they did. Planning has no dispatch to name a model for — it runs on the session's
  own model.
- **ZCode**: one model, see **Harness mapping** below.
- **Muse**: the session's own model, see **Harness mapping** below.

## Allowed pairs

**flowd's store is canonical for which `{model, effort}` every flow dispatch may run on**: each role's pairs live there, read with `flow policy get` and edited by hand in the
Experiments view; the tables below name the roles only. A dispatch names its role with a
`Role: <role id>` line; a pair is `<model>` and `<subagent_type>` joined by `/`, an Explore or Plan
pair adding `@<effort>`. Any well-formed pair is accepted — which models a role may use is the
operator's call, made in the editor. A role's pairs come in three tiers:

- **base** — always admitted.
- **marked** — admitted only when the prompt carries the pair's marker as a whole line
  (`No-progress escalation`, `Operator-approved high`).
- **opt-in** — admitted only on the operator's explicit say-so for a change: the operator tells
  the session which roles may use them, the session records it with
  `flow policy grant -change <name> -roles <role>[,<role>…]`, and each such dispatch's prompt
  carries the whole line `Operator pair: <role>`. `flow policy revoke` removes a grant, and every
  grant lapses when its change reaches `FINISHED`.

`flow policy get` caches its last good answer under the flow state directory; with flowd down it
answers from that cache, and with neither a dispatch runs on `opus`.

Each `###` heading below is one pipeline area, and a role sits under exactly one, with its `Does`
cell, a plain-English line on what the role does. flowd serves both, with the role's pairs, on
`GET /api/v1/roles`, and the Experiments view groups its roles table by them.

### Planning

| Role | Does |
|---|---|
| `explorer` | Reads the codebase to answer a planner's question. |
| `plan-designer` | Proposes a design or plan for the planner to choose from. |
| `planner` | Writes the design and the task plan. |

### Implementation

| Role | Does |
|---|---|
| `implementer` | Writes the code for a group of tasks, tests first. |
| `panel-fix` | Fixes what the review panel found. |
| `operator-asked` | Runs a task the operator asked a subagent for, in their own words, inside a /flow-interactive request. |

### Review

| Role | Does |
|---|---|
| `design-reviewer` | Reviews the plan's three artifacts before implementation, from the documents alone. |
| `gated-reviewer` | Reviews risky tasks as they land, on big changes only. |
| `gated-re-reviewer` | Re-checks a gated review's findings after their fix. |
| `panel-reviewer` | Reviews the whole branch, first pass: the primary + principles bundle, primary alone on a small change. |
| `panel-sanity-efficiency` | Extra bundle, each role only when the change gives it something to read: checks each new test would fail if its behaviour broke, and spots wasted calls. |
| `panel-mutation-failure` | Extra bundle: breaks the code on purpose to prove the tests notice, on every change, and checks error, timeout and partial-write paths where the change does I/O, persistence or concurrency. |
| `panel-rerun` | Re-reviews the branch after a fix round. |
| `late-fix` | Reviews just a small fix made after the branch already passed review. |
| `fix-run-judge` | Decides, once per fix run, which reviewers and which verify a prose fix still needs. |
| `panel-merger` | Merges the reviewers' findings into one list without duplicates. |

### Visual verification

| Role | Does |
|---|---|
| `visual-verifier` | Screenshots the UI and compares it with the mockup. |
| `tooling-analyst` | Works out why visual verification missed a defect. |

### Self-review

| Role | Does |
|---|---|
| `self-review-pass` | Reviews a finished change and the run that built it. |
| `self-review-plan` | Plans the fixes for the self-review's findings. |
| `self-review-fix` | Makes one planned self-review fix. |
| `self-review-review-1` | First review of a self-review fix. |
| `self-review-review-r` | Re-review once that fix is corrected. |

### Pipeline fix

| Role | Does |
|---|---|
| `pipeline-fix-plan` | Plans the fix for a defect found in the flow tooling mid-run. |
| `pipeline-fix-fix` | Makes that tooling fix. |
| `pipeline-fix-review-1` | First review of the tooling fix. |
| `pipeline-fix-review-r` | Re-review once that fix is corrected. |

### Pair choice

| Role | Does |
|---|---|
| `pair-judge` | Picks each dispatch's model and effort from its role's allowed pairs. |

## Experiments

**An open experiment overrides roles' base pairs.** `flow experiment open -name <n>
-cell '<role>=<pair>[,<pair>…]'…` stores it in flowd; while it is open, `flow experiment overrides` prints its pairs, which replace
that role's base pairs for the pair judge and every other reader, and marked and opt-in pairs are never overridden. One
experiment may carry several roles; a role is in at most one open experiment. Role `main` names
the parent session's `<model>@<effort>`; only matching runs count in its after figures. flowd unreachable → no
experiment applies. **Make permanent** (`POST /api/v1/experiments/{id}/promote`, the Experiments
view's button) writes an open experiment's cells — `main` aside — into those roles' base pairs and
closes it `positive`; `/flow-experiment-close` calls it on a positive verdict.

## Harness mapping

**`<parent model>` and `<parent effort>` name the session's own pair, on every harness and in every
inline row.** Where the session does not state one, `<parent model>` is `unknown (agent-defined)` and
`<parent effort>` the literal `default`, the CLI accepting only `low|medium|high|xhigh|max|default`.

**On harness `zcode`, every model a dispatch would be given is `glm-5.3-flash` at effort `high`.**
Every role's pair and an operator override alike resolve and are recorded as they would be on Claude Code,
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
neither — no `model` parameter, no `subagent_type`. Every role's pair and an
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
