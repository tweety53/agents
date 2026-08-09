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

The reasoning behind this file lives in `skills/myflow-start/SKILL-rationale.md`; **a
`/myflow-*` run never loads it.**

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
- **If both are omitted:** enumerate the candidate set exactly as **Change name resolution**
  (`skills/myflow-contracts/pipeline.md`) defines it, restricted to changes with incomplete
  planning artifacts. Exactly one match → resume it, announcing which; multiple →
  **AskUserQuestion** listing each (name, state, last modified); zero → ask what to build.

**Revising an existing proposal** (the change is already at `STARTED`): skip brainstorming's
design gate unless the feedback reopens an architectural question, revise the artifacts in place,
and **republish the artifact to the same source path** (see **Publish the proposal artifact**),
which keeps its URL stable. Never mint a new URL.

A revision round re-enters the loop (see **Convergence** below), **scoped to what the operator's
feedback reopened and to whatever that opens in turn**. A question round opens on it before the
artifacts are revised; settled parts of the plan are not re-brainstormed. The scoping is what makes
the loop usable here — a revision round exists because most of the proposal was right, and re-asking
answered questions would make the cheap path expensive enough that operators stop taking it.

A revision round that answers a question recorded under `## Open questions` follows the same
transition **Open questions** (in section **C**) describes — set that entry's `**Status:**` to
`answered by <decision-id>` and add the decision — even though the round itself does not revisit
section **C**'s artifact-creation steps; the rule governs the file, not the step that happens to be
running.

**Transition the issue to In Progress now**, per
**Transitions** in Jira integration (`skills/myflow-contracts/jira-integration.md`) — before
brainstorming, so the board is correct while planning runs. A failure is one skipped-with-reason
line and planning continues; nothing about this call may delay or alter the proposal.

## Ask the planning effort, the models, and the review panel roster — creating runs only

**Ask once, on the run that creates the change**, and never again for it. "Creates" means the state
file does not exist — not a guess about the operator or the conversation. All five questions below
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

Then ask a **fourth question**, for the review panel roster:

> **Which review panel roster should this change use?**
> - **`light`** *(default, recommended)* — primary, principles, and a low-effort Claude reviewer
> - **`standard`** — primary, principles and Bugbot
> - **`full`** — the roster in force before presets existed, with conditional slots auto-included

Record the answer for the state write in section F, as `reviewPanelRoster`. What each preset means —
which slots each one dispatches, and how the optional, trigger-fired slots are handled under each —
is defined under **5. The review panel** (`skills/myflow-do/SKILL.md`); that section carries the
roster table and is the one to read, and is not restated here.

**Revising an existing proposal** (the change is already at `STARTED`): do not ask, for any of the
five. Read `planningEffort`, `models` and `reviewPanelRoster` from the state file, state which level,
which models and which roster are being reused, and proceed at them. A file that records no level is
planned at `default`, and a file that records no roster is planned at the `light` preset, and both
are said in the handoff too.

Read the level through the retired-key fallback, not from `planningEffort` alone — a file that
recorded a level under the old key would otherwise be reused at `default` rather than at the level
the operator chose, and this run then writes that mistake back. The fallback, the mapping, the
precedence when both keys are present, and what an unmapped value reads as are all
**Planning effort** (`skills/myflow-contracts/state-file.md`)'s, and are not restated here.

## B. Basic Workflow #1 — Brainstorming

Invoke **superpowers:brainstorming** in full: checklist items 1–8, ending with the user approving
the design.

- Save the design to `docs/superpowers/specs/YYYY-MM-DD-<name>-design.md` and stage it when the
  brainstorming skill requires it.
- **HARD GATE:** do not run `openspec new change` until the user approves the design.
- For multi-subsystem work, decompose before proposing.

The approved design is the source for OpenSpec `design.md`; adapt its format, never duplicate a
conflicting design.

### Convergence

**Recording a question never satisfies the test when the command records it pre-emptively** — to
dodge asking a question the operator has not seen; a question recorded that way is still held. **A
question the operator has explicitly deferred is different**: once recorded under
`## Open questions` by the operator's own choice — an "I cannot answer this" response at any round,
or declining the round-3 offer — it stops counting as held, so the next test does not reopen a
silent round on the very question the operator just said they could not answer. Those are different
acts: one is the command avoiding a question, the other is the operator answering one by choosing
not to answer it yet.

**Record it immediately, right where it is given, rather than at the round's end.** The
record-and-defer path is not reachable only through the round-3 offer's decline — when the operator
answers with some form of "I cannot answer this," record it under `## Open questions` at that point
rather than holding the round open indefinitely or treating silence as consent to guess. The rest of
that round's questions, if any, are still asked; only the unanswerable one is deferred.

**When the test comes back empty, do not silently proceed.** State what you believe settled — the
answers and the decisions this stage now rests on — **and every question still recorded under
`## Open questions`**, including one deferred as far back as round one, not only what it settled on,
so the operator's answer rests on the same picture the command has. Then ask, with named options:

> **That is everything I have settled. Anything still unclear before I move on?**
> - **Nothing unclear — move on** *(recommended)*
> - **Another round — I have something** *(default — anything short of an explicit "move on" is
>   treated as this)*

End the stage only on an explicit choice of **move on**. An answer that names something opens
another round, and the test applies to that round's answers exactly as to any other. **This is the
one prompt in this file where the safe default and the recommended option differ, and deliberately
so**: silence or a stalled prompt from an operator who is present is not "move on," and defaults to
another round rather than to the recommended choice — recommending *move on* is honest only while an
operator has actually said so. Print `⚠ another round — no explicit answer` when this default fires,
so a reader can tell an operator-requested round from one nothing could confirm.

**A session that cannot ask at all is a narrower, bounded exception, not a second version of "an
operator who is present but silent."** The confirm fires only when the convergence test came back
empty, so opening "another round" here has nothing to explore and, with no hard cap, only re-empties
the test and re-fires the confirm — `empty test → confirm → no answer → another round → empty test →
confirm`, without end. **Unrecognised statuses** (`skills/myflow-contracts/jira-integration.md`)
already names this exact failure mode for its own interactive ask — "a session that cannot ask at
all" — and gives it a terminating outcome rather than a retry; cited rather than restated here. When
no answer is possible at all — no channel to ask through, not silence from a reachable operator —
record the confirm itself under `## Open questions` and end the stage, the same outcome declining the
offer below produces, printing `⚠ open question recorded — no answer was possible` so a reader can
tell the two apart.

**From the third round onward, do not open a round silently.** Show two lists — the full still-open
backlog and, separately, what round `<n>` itself would ask — and offer the round as a named choice:

> **Still open: `<full still-open backlog>`. Round `<n>` would ask about `<this round's slice>`.
> Open it?**
> - **Yes — run another round** *(default, recommended)*
> - **No — record everything still open and move on**

A decline records the **full still-open backlog** shown above, not only what round `<n>` would have
asked — one answer can open more than one question at once, and every one of them is still open
whether or not this round would have reached it yet. Silence, a stalled prompt, or any answer that is
not one of the two options above defaults to **Yes** — the option already recommended, since the
offer is reachable only while something is genuinely open. Print `⚠ another round — no explicit
answer` when this default fires, for the same reason the confirm's marker does. Running one more
round costs an exchange; silently skipping it is exactly what the soft bound exists to prevent.

Rounds one and two open without asking. **`3` is a tuned value, and this file is the only place it
is written** — the contract and the pipeline carry the shape of the bound, never the number, so it
can move without amending either. **The threshold counts rounds, not questions**, so it lands at a
different point in the conversation depending on planning effort: at `detailed`, where a round is
one question, the offer can appear after as few as three questions; at `low`, where a round batches
many, the same threshold takes much longer to reach, or is never reached at all. That coupling is
accepted rather than compensated for — the threshold is stated once, in rounds, and each level's own
grouping decides how much a round holds.

A round the operator opens by answering **Another round — I have something** at the confirm counts
toward this same 1-2-silent / 3-onward-offered sequence exactly as any other round — an
operator-initiated round is still a round, and exempting it would let repeated "another round"
answers at the confirm outrun the offer's own visibility.

**There is no hard cap.** No round count ends the stage; the rule that only an explicit operator
answer does is stated once under **Stage exit — never the command's own judgment**
(`skills/myflow-contracts/pipeline.md`) and not restated here.

**The two prompts recommend opposite courses, and each is honest because of its own trigger — do
not harmonise them.** The confirm recommends *moving on* precisely because it is unreachable while
this command holds an unanswered question. The offer recommends *another round* for the mirror
reason: it is reachable only while this command genuinely holds one. That is the same shape as the
**Stop** recommendation at the unfinished-work gate of `/myflow-finish` run 1, whose reasoning is
stated under **Finish contract** (`skills/myflow-contracts/finish-contract.md`) and is not
re-argued here.

## C. Create the change and its artifacts

```bash
openspec new change "<name>"
openspec status --change "<name>" --json
openspec instructions <artifact-id> --change "<name>" --json
```

Create every artifact `applyRequires` names:

- **proposal.md** — what and why
- **specs/** — delta specs, one file per capability named in the proposal
- **design.md** — how, from the approved design, including `## Decisions` and `## Open questions`
  (both below)
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

### Open questions

`## Open questions` sits beside `## Decisions` and is shaped like it. It is the record the
convergence offer in section **B** names when the operator answers *No — record what is open and
move on*: one entry for each question this stage ends still holding. A decision records a choice
that was made; an entry here records one that was knowingly not.

**A stage that left nothing open records none.** Leave the section empty rather than inventing
entries — the same rule `## Decisions` carries above. Empty, not absent: a section that is missing
reads as a stage that never applied the test.

```markdown
### <the question>

**ID:** <kebab-case-slug>
**Status:** open
**Why it is open:** <deferred by the operator, blocked on something external, …>
**What it affects:** <what would change depending on the answer>
```

**ID** is assigned once, at creation, and is **immutable**, for the reason a decision's ID is: it
is the match key a later round uses to find the entry again, and matching on heading text alone
breaks on any rewording. **An ID SHALL be unique across `## Decisions` and `## Open questions`
together, not merely within its own section** — the two are one namespace, because
`answered by <decision-id>` reaches from one into the other. On a collision — a new entry about to
be assigned an ID that already exists in either section — assign a fresh, distinguishing ID instead
of reusing it; the entries do not merge. A round that answers the question sets that entry's
`**Status:**` to `answered by <decision-id>` and adds the answering entry under `## Decisions`.
**Never delete or rewrite an entry once recorded** — what was left open, and which decision closed
it, is the point.

What this section holds is what the `STARTED` handoff counts. That line is defined once under
**The block each state renders** (`skills/myflow-contracts/handoff-blocks.md`), which also states why it
is regenerable rather than run-only.

## D. Basic Workflow #3 — Writing plans

Invoke **superpowers:writing-plans** to enrich `<changeRoot>/tasks.md` to plan quality: exact
paths, verification commands, bite-sized steps, no placeholders. Run its self-review (spec
coverage, placeholder scan, type consistency) before finishing.

While enriching `tasks.md`, tag every fenced block and every numeric claim per
**Plan provenance** (`skills/myflow-contracts/plan-provenance.md`): code that cannot be verified is tagged
`unverified:` and **kept** — a plan without the snippet is worse than a plan with a labelled guess.

While enriching `tasks.md`, also tag every task with `**Build:**` per **The build-green tag**
(`skills/myflow-contracts/build-green.md`).

While enriching `tasks.md`, also tag every task with the mechanically-checkable field family per
`myflow-task-commit-fields`'s requirement **Every task declares mechanically-checkable fields** —
checked by a runtime guard during `/myflow-do`:

- `**Files:**` — the paths this task's commit will touch, with an optional
  `**Allowed-collateral:**` glob for paths a legitimate sweep may also touch without being a
  declared file.
- `**Tests:**` — the names of the tests this task adds.
- `**Regression:**` — per declared test, what fails if this task's commit is reverted.
- `**Baseline:**` — the expected test counts, as `before=<N> after=<N>`.
- `**Commit:**` — the commit subject line this task's implementer must use.

A task tagged `Build: red` additionally carries `**Squash-with:** Task <N>`, naming the green task
its commit folds into. `tasks.md` in this change's own `openspec/changes/` directory demonstrates
the real syntax for all of these fields.

Add this header to `tasks.md`:

```markdown
> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.
```

Before publishing the artifact, run the project's configured plan-provenance guard and its
configured build-green guard, if the project declares them, and fix any hit.

## E. Publish the proposal artifact

Load the `artifact-design` skill, then build one self-contained page carrying the proposal's why
and what, the design including `## Decisions` and `## Open questions`, the delta specs, and the
task list. Publish it with the Artifact tool. The open questions are carried so that stopping with
something unanswered is visible at the gate the operator actually reads, rather than held only in
the session transcript.

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
  "reviewPanelRoster": "<light|standard|full, or null>",
  "prUrl": null,
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/myflow-start"
}
```

The In Progress transition already happened in section **A**. Sync added scope here, per
**Description sync** in Jira integration (`skills/myflow-contracts/jira-integration.md`) — only
when this run added scope the issue does not already describe.

Stage the planning artifacts. The state file lives outside the repo — never `git add` it.

```
## Proposal ready — review required

**Change:** <name>
**Artifact:** <artifactUrl> | missing
**Recorded:** <N> decisions | none · <N> open questions | none · effort <level> | <level> (reused from the creating run) | not recorded — planned at default · models implementation <model | not recorded>, review panel <model | not recorded>, panel fixes <model | not recorded> · roster <preset> | <preset> (reused from the creating run) | not recorded — planned at the `light` preset
**Jira:** <KEY> → In Progress | <KEY> already In Progress (no transition) | none linked | ⚠ Jira: skipped — <reason>
**Jira description (pre-edit):** <the text as it stood before the write, verbatim in a fenced block, inside <details> when long> | omitted — this run wrote no description

Open in IntelliJ:
open -na "IntelliJ IDEA" --args "<absolute main checkout path>"

Read the artifact. Re-run this command to revise the plan.

Next:
/myflow-do <name>
```

The IntelliJ path is the **main checkout** — no worktree exists at this state. Resolve it via
`--git-common-dir`.

The pre-edit description line is present only on a run that wrote the description, and reproduces
that text without summarising or reflowing it — the transcript is then the recovery path, since
there is no local backup. A run that wrote nothing omits the line rather than printing an empty
one. See **Description sync** (`skills/myflow-contracts/jira-integration.md`).

## Guardrails

- **Never skip** brainstorming (#1) or writing-plans (#3), or the design approval gate.
- **Never** leave `tasks.md` a thin scaffold.
- **Never** publish a plan carrying an untagged block or an unsourced number.
- **Never** publish a plan carrying a task with no `**Build:**` tag, or a `red` tag with no
  resolvable green partner.
- **Never** finish without publishing the artifact and recording `artifactUrl`.
- **Never** mint a new artifact URL on a revision round.
- **Never** delete a superseded decision; mark it superseded.
- **Never** ask for a planning effort level on a revision round — read the recorded one and say so.
- **Never** ask for a model choice on a revision round — read the recorded ones and say so.
- **Never** ask for a review panel roster on a revision round — read the recorded one and say so.
- **Never** let a planning effort level skip brainstorming, the design approval gate, writing-plans,
  or leave `tasks.md` a scaffold. It sizes the thinking inside the gates, never the gates.
- **Never** let a review panel roster skip brainstorming, the design approval gate, writing-plans, or
  leave `tasks.md` a scaffold, and never let a roster move the handoff bar in `/myflow-do` — every
  preset still hands off only at zero open findings at any severity.
- **Never** write code, create a worktree, or create a branch.
- **Never** commit anything. Stage the planning artifacts and leave the commit to `/myflow-finish`,
  per **Git boundaries** (`skills/myflow-contracts/pipeline.md`).
- **Never** let a Jira call block, delay, or alter the proposal — one skipped-with-reason line.
  **Exactly one carve-out is reachable from this command**, and it is
  **Unrecognised statuses** (`skills/myflow-contracts/jira-integration.md`): a single yes/no when
  the issue sits at a status matching no name mapped onto the four ordered positions, where anything
  but an explicit yes is that same skipped-with-reason line and the proposal is untouched either
  way. The contract bounds the set at **two**; the other is the join confirmation, which only ever
  occurs during `/myflow-finish` run 1 — `/myflow-start` files and joins nothing — so this is a
  count for this command, not a competing count for the pipeline.
- **Never** resolve an open question by assumption. Put it to the operator, at the point the
  answer is first needed, and do everything that does not depend on it in the meantime. A lower
  planning effort level may group questions into fewer rounds and batch related ones into one
  prompt; it may never turn a question into an assumption.
- **Never** end the stage on this command's own judgment, and never leave brainstorming holding a
  question the operator was never asked — the rule is stated once under
  **Stage exit — never the command's own judgment**
  (`skills/myflow-contracts/pipeline.md`); its tuned exits are **Convergence** in section **B**, and
  neither is restated here.
- **Never** let a planning effort level end the loop early. A level may group more questions into
  one round; it may never decide that no further round opens.
- **Never** ask for an approval in open prose. Offer named options, mark the recommended one, and
  say what each one will do.
- **No flags.** The only argument is the optional change name; report anything else.
