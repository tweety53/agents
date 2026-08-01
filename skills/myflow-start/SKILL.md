---
name: myflow-start
description: Propose an OpenSpec change with Superpowers brainstorming and writing-plans woven into OpenSpec artifacts, and publish a proposal artifact. Re-run to revise the proposal. Use for /myflow-start.
allowed-tools: Bash(openspec:*)
license: MIT
---

Propose an OpenSpec change with Superpowers Basic Workflow steps **#1** and **#3** fully
intertwined with OpenSpec artifact creation. **No code is written and no worktree is created.**

**Announce at start:** "Using myflow-start for change `<name>`."

Immediately after that line, print these two commands for the operator to paste, per
**Handoff output** (`skills/myflow-contracts/pipeline.md`) — that section fixes the colour and
records why they are printed rather than invoked; do not restate its reasoning here:

```text
/rename <change-name>
/color cyan
```

**Load `skills/myflow-contracts/pipeline.md` first** — it is canonical for the states, the
command→state transition table, the wrong-state handoff, and the handoff output shape.

**Then register this run's steps** with the harness's task-list mechanism, before any work begins,
and keep each entry's status current as the run proceeds, per
**Progress visibility** (`skills/myflow-contracts/pipeline.md`) — that section names which steps
this command registers and is the one to read.

## State gate

Accepts **no change** (creates one) or **`STARTED`** (revises the existing proposal). Ends at
`STARTED`.

At `IN_PROGRESS` the proposal has already been implemented against — emit the wrong-state handoff
from **Wrong state for this command** (`skills/myflow-contracts/pipeline.md`) and recommend
`/myflow-do`, which documents a change of plan in the proposal before implementing it. Only proceed
on an explicit override.

## Superpowers Basic Workflow (this stage)

| Step | Skill | When |
|------|-------|------|
| **1** | **superpowers:brainstorming** | Before any OpenSpec artifacts — full checklist, design approval gate |
| **3** | **superpowers:writing-plans** | After OpenSpec draft artifacts — enrich `tasks.md` to plan quality |

Steps **2, 4–6** run in `myflow-do`. Do not run them here.

## A. Resolve the change

**Resolve the linked Jira issue first** — it decides the change name. Follow
**Resolution (how `jiraIssue` is decided)** in `skills/myflow-contracts/jira-integration.md`
exactly; it is canonical and is not restated here. This is the only command that resolves a key.

Then the change name:

- **With a linked issue**, the name is `<lowercased-key>-<slug>`, per
  **Change naming** (`skills/myflow-contracts/jira-integration.md`). Derive the slug from the issue
  summary when only a key was given.
- **Without one**, the name is the descriptive slug alone.
- If a name or description was given, use it (derive kebab-case from the description if only a
  description was given).
- **If both are omitted:** run `openspec list --json`, filter to non-archived changes with
  incomplete planning artifacts. Exactly one match → resume it, announcing which; multiple →
  **AskUserQuestion** listing each (name, state, last modified); zero → ask what to build.

**Revising an existing proposal** (the change is already at `STARTED`): skip brainstorming's
design gate unless the feedback reopens an architectural question, revise the artifacts in place,
and **republish the artifact to the same source path** (see **Publish the proposal artifact**),
which keeps its URL stable. Never mint a new URL.

**Transition the issue to In Progress now**, per
**Transitions** in Jira integration (`skills/myflow-contracts/jira-integration.md`) — before
brainstorming, so the board is correct while planning runs. A failure is one skipped-with-reason
line and planning continues; nothing about this call may delay or alter the proposal.

## Ask the planning effort and the models — creating runs only

**Ask once, on the run that creates the change**, and never again for it. "Creates" means the state
file does not exist — not a guess about the operator or the conversation. All four questions below
share that rule.

Use **AskUserQuestion**, the same mechanism `/myflow-finish` uses for its integration choice.
Neither the planning effort nor a model is **ever** an argument: the only argument this command
accepts is the optional change name, and anything else is still reported rather than interpreted.

> **How much planning effort should this change take?**
> - **`default`** *(recommended)* — the checklist followed with related questions grouped
> - **`detailed`** — each checklist item worked separately, every design section approved on its own
> - **`low`** — questions batched, the design presented once

Record the answer for the state write in section F. The levels and what they may change are defined
under **Planning effort** (`skills/myflow-contracts/state-file.md`) — that section carries the
operational table and is the one to read; do not restate it here.

Then ask **three more, one per model role**, each as its own question with its default marked as the
recommendation:

> **Which model should implement this change?** — the implementer subagents `/myflow-do` dispatches
> - **Opus** *(default, recommended)* — or the harness's strongest available model
> - any other model the harness offers

> **Which model should the review panel run on?** — every panel slot that takes a model override
> - **Sonnet** *(default, recommended)*
> - any other model the harness offers

> **Which model should apply panel fixes?** — the subagents that repair panel findings
> - **Opus** *(default, recommended)* — or the harness's strongest available model
> - any other model the harness offers

Record all three for the state write in section F, as `implementation`, `reviewPanel` and
`panelFix`. The roles, their defaults, why the panel-fix default is not Sonnet, and how a recorded
choice relates to a session instruction are defined under
**Model policy** (`skills/myflow-contracts/pipeline.md`) — that section is the one to read; do not
restate it here.

**Revising an existing proposal** (the change is already at `STARTED`): do not ask, for any of the
four. Read `planningEffort` and `models` from the state file, state which level and which models are
being reused, and proceed at them. A file that records no level is planned at `default`, and that is
said in the handoff too.

Read the level through the retired-key fallback, not from `planningEffort` alone — a file that
recorded a level under the old key would otherwise be reused at `default` rather than at the level
the operator chose, and this run then writes that mistake back. The fallback, the mapping, the
precedence when both keys are present, and what an unmapped value reads as are all
**Planning effort** (`skills/myflow-contracts/state-file.md`)'s, and are not restated here.

## B. Basic Workflow #1 — Brainstorming

Invoke **superpowers:brainstorming** in full: checklist items 1–8, ending with the user approving
the design.

- Save the design to `docs/superpowers/specs/YYYY-MM-DD-<name>-design.md` and commit it when the
  brainstorming skill requires it.
- **HARD GATE:** do not run `openspec new change` until the user approves the design.
- For multi-subsystem work, decompose before proposing.

The approved design is the source for OpenSpec `design.md`; adapt its format, never duplicate a
conflicting design.

## C. Create the change and its artifacts

```bash
openspec new change "<name>"
openspec status --change "<name>" --json
openspec instructions <artifact-id> --change "<name>" --json
```

Create every artifact `applyRequires` names:

- **proposal.md** — what and why
- **specs/** — delta specs, one file per capability named in the proposal
- **design.md** — how, from the approved design, including `## Decisions` (below)
- **tasks.md** — a checkbox scaffold; **writing-plans enriches it next**

Do not copy `<context>` / `<rules>` blocks from the CLI instructions into artifact files.

### Decisions

`## Decisions` in `design.md` is sourced from the **brainstorming dialogue** — the approach the
user chose, the alternatives on the table, and the tradeoff that ruled each one out. Whenever
brainstorming presented competing approaches and the user picked one, that is a decision.

**A design that forced no choices records none.** Leave the section empty rather than fabricating
entries.

```markdown
### <the decision>

**ID:** <kebab-case-slug>
**Status:** active
**Chosen:** <option> — <one-line rationale>
**Considered:** <other options, each with the tradeoff that ruled it out>
```

**ID** is assigned once, at creation, and is **immutable**. The heading prose may be reworded
across revision rounds; the ID never changes, because it is the match key a later round uses to
**supersede** a decision: set the old entry's `**Status:**` to `superseded by <new-id>` and append
a new entry with a fresh ID. **Never delete or rewrite a superseded entry** — the history is the
point. Matching on heading text alone breaks on any rewording.

## D. Basic Workflow #3 — Writing plans

Invoke **superpowers:writing-plans** to enrich `<changeRoot>/tasks.md` to plan quality: exact
paths, verification commands, bite-sized steps, no placeholders. Run its self-review (spec
coverage, placeholder scan, type consistency) before finishing.

While enriching `tasks.md`, tag every fenced block and every numeric claim per
**Plan provenance** (`skills/myflow-contracts/plan-provenance.md`): code that cannot be verified is tagged
`unverified:` and **kept** — a plan without the snippet is worse than a plan with a labelled guess.

Add this header to `tasks.md`:

```markdown
> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.
```

Before publishing the artifact, run the project's configured plan-provenance guard if it declares
one, and fix any hit.

## E. Publish the proposal artifact

Load the `artifact-design` skill, then build one self-contained page carrying the proposal's why
and what, the design including `## Decisions`, the delta specs, and the task list. Publish it with
the Artifact tool.

Write its source to the deterministic path
`/Users/tweety53/Agents/myflow/state/<project-key>/<name>-proposal-artifact.html`, resolving
`<project-key>` exactly as **State file** (`skills/myflow-contracts/state-file.md`) does. That
keeps it outside the repo, beside the state file — never `git add` or commit it. A revision round
republishes to this **same** path, which is what keeps the URL stable.

Record the returned URL as `artifactUrl`. This page is what the human reads at `STARTED`.

## F. Write state and hand off

Write the state file per **State file** (`skills/myflow-contracts/state-file.md`):

```json
{
  "state": "STARTED",
  "branch": null,
  "worktrees": {},
  "artifactUrl": "<published URL>",
  "jiraIssue": "<resolved key, or null>",
  "planningEffort": "<low|default|detailed, or null>",
  "models": {
    "implementation": "<model, or null>",
    "reviewPanel": "<model, or null>",
    "panelFix": "<model, or null>"
  },
  "prUrl": null,
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/myflow-start"
}
```

The In Progress transition already happened in section **A**. Sync added scope here, per
**Description sync** in Jira integration (`skills/myflow-contracts/jira-integration.md`) — only
when this run added scope the issue does not already describe.

Stage the planning artifacts. The state file lives outside the repo — never `git add` it.

The block below is **not** a second definition of the handoff. It is this command's rendering of the
`STARTED` template, which is defined once under
**The block each state renders** (`skills/myflow-contracts/pipeline.md`) and is canonical for the
labels, the field set and their order. What this block adds is the enumeration of the literal
alternatives `/myflow-start` writes for each placeholder that file describes. **Change the template
first and bring this block with it** — a field added here and not there is drift the moment
`/myflow-status <name>` regenerates the same state.

```
## Proposal ready — review required

**Change:** <name>
**Artifact:** <artifactUrl> | missing
**Decisions recorded:** <N> | none
**Jira:** <KEY> → In Progress | <KEY> already In Progress (no transition) | none linked | ⚠ Jira: skipped — <reason>
**Jira description (pre-edit):** <the text as it stood before the write, verbatim in a fenced block, inside <details> when long> | omitted — this run wrote no description
**Planning effort:** <level> | <level> (reused from the creating run) | not recorded — planned at default
**Models:** implementation <model | not recorded>, review panel <model | not recorded>, panel fixes <model | not recorded>

Open in IntelliJ:
open -na "IntelliJ IDEA" --args "<absolute main checkout path>"

Read the artifact. Re-run this command to revise the plan.

Next:
/myflow-do <name>
```

The IntelliJ path is the **main checkout** — no worktree exists at this state. Resolve it via
`--git-common-dir`.

**`missing` is a real alternative on the artifact line, not a defensive one.** This command's
guardrails forbid finishing without publishing, so its own runs print a URL; the alternative is
carried because `/myflow-status <name>` renders this same block from a state file whose
`artifactUrl` may be `null` — a file self-heal rebuilt from artifacts loses it, and self-heal names
it among the unrecovered fields when it does. Omitting the alternative here would narrow the
template and teach the next reader to drop the case the missing-rather-than-dropped rule requires.

The pre-edit description line is present only on a run that wrote the description, and reproduces
that text without summarising or reflowing it — the transcript is then the recovery path, since
there is no local backup. A run that wrote nothing omits the line rather than printing an empty
one. See **Description sync** (`skills/myflow-contracts/jira-integration.md`).

## Guardrails

- **Never skip** brainstorming (#1) or writing-plans (#3), or the design approval gate.
- **Never** leave `tasks.md` a thin scaffold.
- **Never** publish a plan carrying an untagged block or an unsourced number.
- **Never** finish without publishing the artifact and recording `artifactUrl`.
- **Never** mint a new artifact URL on a revision round.
- **Never** delete a superseded decision; mark it superseded.
- **Never** ask for a planning effort level on a revision round — read the recorded one and say so.
- **Never** ask for a model choice on a revision round — read the recorded ones and say so.
- **Never** let a planning effort level skip brainstorming, the design approval gate, writing-plans,
  or leave `tasks.md` a scaffold. It sizes the thinking inside the gates, never the gates.
- **Never** write code, create a worktree, or create a branch.
- **Never** let a Jira call block, delay, or alter the proposal — one skipped-with-reason line.
  **Exactly one carve-out is reachable from this command**, and it is
  **Unrecognised statuses** (`skills/myflow-contracts/jira-integration.md`): a single yes/no when
  the issue sits at a status outside the four ordered names, where anything but an explicit yes is
  that same skipped-with-reason line and the proposal is untouched either way. The contract bounds
  the set at **two**; the other is the join confirmation, which only ever occurs during
  `/myflow-finish` run 1 — `/myflow-start` files and joins nothing — so this is a count for this
  command, not a competing count for the pipeline.
- **Never** resolve an open question by assumption. Put it to the operator, at the point the
  answer is first needed, and do everything that does not depend on it in the meantime. A lower
  planning effort level may group questions into fewer rounds and batch related ones into one
  prompt; it may never turn a question into an assumption.
- **Never** ask for an approval in open prose. Offer named options, mark the recommended one, and
  say what each one will do.
- **No flags.** The only argument is the optional change name; report anything else.
