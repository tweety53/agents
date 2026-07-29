---
name: myflow-start
description: Propose an OpenSpec change with Superpowers brainstorming and writing-plans woven into OpenSpec artifacts, and publish a proposal artifact. Re-run to revise the proposal. Use for /myflow-start.
allowed-tools: Bash(openspec:*)
license: MIT
---

Propose an OpenSpec change with Superpowers Basic Workflow steps **#1** and **#3** fully
intertwined with OpenSpec artifact creation. **No code is written and no worktree is created.**

**Announce at start:** "Using myflow-start for change `<name>`."

**Load `skills/myflow-contracts/pipeline.md` first** — it is canonical for the states, the
command→state transition table, the wrong-state handoff, and the handoff output shape.

## State gate

Accepts **no change** (creates one) or **`STARTED`** (revises the existing proposal). Ends at
`STARTED`.

At `IN_PROGRESS` the proposal has already been implemented against — emit the wrong-state handoff
from **Wrong state for this command** and recommend `/myflow-do`, which documents a change of plan
in the proposal before implementing it. Only proceed on an explicit override.

## Superpowers Basic Workflow (this stage)

| Step | Skill | When |
|------|-------|------|
| **1** | **superpowers:brainstorming** | Before any OpenSpec artifacts — full checklist, design approval gate |
| **3** | **superpowers:writing-plans** | After OpenSpec draft artifacts — enrich `tasks.md` to plan quality |

Steps **2, 4–6** run in `myflow-do`. Do not run them here.

## A. Resolve the change

**Resolve the linked Jira issue first** — it decides the change name. Follow **Resolution** under
**Jira integration** (`skills/myflow-contracts/jira-integration.md`) exactly; it is canonical and
is not restated here. This is the only command that resolves a key.

Then the change name:

- **With a linked issue**, the name is `<lowercased-key>-<slug>`, per **Change naming** in
  **Jira integration**. Derive the slug from the issue summary when only a key was given.
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

While enriching `tasks.md`, tag every fenced block and every numeric claim per **Plan provenance**
(`skills/myflow-contracts/plan-provenance.md`): code that cannot be verified is tagged
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
  "prUrl": null,
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/myflow-start"
}
```

**Transition the issue to In Progress after the state write** — Jira must never be able to prevent
the state from being recorded. Sync added scope here too. Both mechanisms are defined once under
**Jira integration** (`skills/myflow-contracts/jira-integration.md`); follow them there. Description
sync applies **only** when this run added scope the issue does not already describe.

Stage the planning artifacts. The state file lives outside the repo — never `git add` it.

```
## Proposal ready — review required

**Change:** <name>
**Artifact:** <artifactUrl>
**Decisions recorded:** <N> | none
**Jira:** <KEY> → In Progress | <KEY> already In Progress (no transition) | none linked | ⚠ Jira: skipped — <reason>

Open in IntelliJ:
open -na "IntelliJ IDEA" --args "<absolute main checkout path>"

Read the artifact. Re-run this command to revise the plan.

Next:
/myflow-do <name>
```

The IntelliJ path is the **main checkout** — no worktree exists at this state. Resolve it via
`--git-common-dir`.

## Guardrails

- **Never skip** brainstorming (#1) or writing-plans (#3), or the design approval gate.
- **Never** leave `tasks.md` a thin scaffold.
- **Never** publish a plan carrying an untagged block or an unsourced number.
- **Never** finish without publishing the artifact and recording `artifactUrl`.
- **Never** mint a new artifact URL on a revision round.
- **Never** delete a superseded decision; mark it superseded.
- **Never** write code, create a worktree, or create a branch.
- **Never** let a Jira call block, delay, or alter the proposal — one skipped-with-reason line.
- **No flags.** The only argument is the optional change name; report anything else.
