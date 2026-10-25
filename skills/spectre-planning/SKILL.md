---
name: spectre-planning
description: The planning prompts of spectre's tool + planning prompts mode — brainstorming a design in plan mode, filling proposal.md, design.md and tasks.md, and writing a plan spectre's guards accept. The tool-only mode is the bare spectre CLI and loads none of this. Use when planning a spectre change.
---

# spectre planning

The generic planning prompts for a spectre change: brainstorm the design, create the change's three artifacts, and write the plan.

## Plan mode

**Planning runs in plan mode.** On Claude Code, call `EnterPlanMode` before exploring.
Investigation, questions and competing designs run inside plan mode. Once convergence closes,
write the design to the plan file the harness names and call `ExitPlanMode`: **Approve** is the
HARD GATE's design approval (**The checklist**, below); **reject with feedback** reopens
convergence with the feedback as the next round's input. On Cursor, switch to Plan mode before
exploring, or ask the user to switch when the agent cannot switch modes itself. A declined `EnterPlanMode` consent is one
line, and planning continues without plan mode, exactly as on another harness. Every other gate is
unchanged, the plan review gate included.

## B. Basic Workflow #1 — Brainstorming

### The checklist

Invoke **superpowers:brainstorming** in full: checklist items 1–8, ending with the user approving
the design.

- **Display-only formatting is decided, never asked.** A choice that changes only how a value is
  shown — numeric precision, an empty-cell placeholder, and the like — and that the request, ticket
  or mockup leaves open is chosen by you and recorded under `## Decisions`; no question is asked
  for it. A mockup or ticket that states it still governs.
- **A reported defect is checked against history before its fix is designed.** For each entry of a
  bug list, a findings file or a ticket naming symptoms, run `git log --since=<the report's date>`
  over the paths it names on the base, search `<project>/spectre/changes/archive/` for a recorded
  rejection of the remedy it proposes, and run its cheap reproducer. An entry already fixed is
  recorded in the design as fixed, with the commit that fixed it, and planned no task; a remedy an
  archived decision rejected is not designed again unless the design records what that rejection
  missed (kan-970: six entries were already fixed upstream and one remedy had been rejected twice).

- **HARD GATE:** do not run `spectre new` until the user approves the design. Approval is the
  merged confirm's first option under **Convergence** below; no separate approval question is
  asked. In plan mode the approval comes from `ExitPlanMode` instead (**Plan mode**, above) —
  the merged confirm still closes the checklist (decision `exit-plan-mode-is-approval`, kan-1033).

- The design presentation does **not** end a section, or the whole design, with a "does this look
  right?" question — present the section(s) and proceed directly, section to section and then into
  artifact creation, unless the operator raises an objection during or after that presentation. This
  is a scoped override of `superpowers:brainstorming`'s hard design-approval gate.
- Ask every pending question whose wording does not depend on another pending answer in one
  **AskUserQuestion** call, up to four questions per call; a question that only makes sense once
  another is answered waits for the next call; the convergence confirm and the third-round offer
  may be the last question in such a call. This is a scoped override of
  `superpowers:brainstorming`'s "Only one question per message".

The approved design is the source for the change's `design.md`; adapt its format, never duplicate a
conflicting design.

### Explore and design agents

**Explore before the first question.** Before the first design question — after the withdrawal
reachability check, when that runs — dispatch 1–3 Explore agents (`subagent_type: "Explore"`) in
one message, so they run concurrently, each on a different area: the code the change touches,
existing patterns and utilities to reuse, and the tests and guards around it. One when the change
is confined to one known area; more when its scope is uncertain or spans areas. Then read the
critical files the agents name yourself before asking anything. Their conclusions feed the
questions; they never replace asking.

**Competing designs for a change that is not trivial** — more than one valid approach, a
cross-cutting change, or an architectural choice: dispatch 1–3 Plan agents (`subagent_type:
"Plan"`) in one message, each briefed with the Explore conclusions and its own angle — simplest,
most robust, best fit with existing patterns. Compare their proposals and recommend one. Each
alternative you reject becomes the `**Considered:**` line of the matching `## Decisions` entry,
with the tradeoff that ruled it out. A change with one obvious approach skips this step and records
why under `## Recommended approach`.

### Stage exit — never the command's own judgment

Within a single run, a stage that loops — most concretely the brainstorm
stage, whose convergence test reopens after every planning-stage exchange that leaves a question the
command's inputs do not answer — never closes on the command's own judgment. It closes only on an
explicit operator answer: at a confirm, or by declining an offer, recording what is still open
rather than assuming it away. The one bounded exception is a session that cannot ask at all: it
records the confirm itself as an open question and ends the stage there, since no operator answer
could ever arrive through it. An operator who is present but silent is not that exception and still
gets another round. The same explicit answer may both close the checklist and grant the design
approval, as **Convergence** (`skills/spectre-planning/SKILL.md`) defines.

### Convergence

**Recording a question never satisfies the test when the command records it pre-emptively** — to
dodge asking a question the operator has not seen; a question recorded that way is still held. **A
question the operator has explicitly deferred is different**: once recorded under `## Open
questions` by the operator's own choice — an "I cannot answer this" response at any round, or
declining the round-3 offer — it stops counting as held. Record it immediately, right where it is
given, rather than at the round's end. The rest of that round's questions, if any, are still asked;
only the unanswerable one is deferred.

**When the test comes back empty, do not silently proceed.** State what you believe settled — the
answers and the decisions this stage now rests on — **and every question still recorded under `##
Open questions`**, including one deferred as far back as round one. Then ask, with named options:

> **That is everything I have settled. Anything still unclear before I move on?**
> - **Nothing unclear — approve the design and move on** *(recommended)*
> - **Another round — I have something** *(default — anything short of an explicit "approve and
>   move on" is treated as this)*
> - **Revise — I have a change to the design**

End the stage only on an explicit choice of **approve the design and move on**. An answer that
names something opens another round. **Here the safe default and the recommended option differ**: silence or a stalled prompt from an operator who is
present is not "approve the design and move on," and defaults to another round rather than to the
recommended choice. Print `⚠ another round — no explicit answer` when this default fires.

*Revise* is a round — it counts toward the third-round offer below exactly as *Another round*
does — and differs only in what happens next: the changed design section(s),
re-presented before the next confirm, in place of new questions.

**A batched confirm.** When the confirm rides along with a round's questions as the last block of
the turn, **approve the design and move on** folds that turn's answers into the design and ends the
stage; an answer to an accompanying question that names something new opens another round
regardless of the confirm's choice. The silence default above and its `⚠ another round — no
explicit answer` marker are unchanged.

When no answer is possible at all — no channel to ask through — record the confirm itself under `##
Open questions` and end the stage, printing `⚠ open question recorded — no answer was possible`.

**From the third round onward, do not open a round silently.** Show two lists — the full still-open
backlog and, separately, what round `<n>` itself would ask — and offer the round as a named choice:

> **Still open: `<full still-open backlog>`. Round `<n>` would ask about `<this round's slice>`.
> Open it?**
> - **Yes — run another round** *(default, recommended)*
> - **No — record everything still open and move on**

A decline records the **full still-open backlog** shown above. Silence, a stalled prompt, or any
answer that is not one of the two options above defaults to **Yes**. Print `⚠ another round — no
explicit answer` when this default fires.

Rounds one and two open without asking. **There is no hard cap.** No round count ends the stage —
see **Stage exit — never the command's own judgment** (`skills/spectre-planning/SKILL.md`).

## C. Create the change and its artifacts

`spectre new` scaffolds `<project>-worktrees/<name>/spectre/changes/<name>/`, and refuses three ways: exit `2` and
*no tree found* when the project holds no `<project>/spectre/` tree at all; exit `2` and `invalid
change id` when `<name>` is not a single flat directory name; and exit `1` and `<path> already
exists` when the change is already there — **the ordinary case when resuming at `STARTED`** per
above. Revise the artifacts in place and never `--force` past it. That directory is the change
root — `<changeRoot>` below — by construction.

`spectre new` writes three stubs and nothing else: `proposal.md` (`## Why`, `## What changes`),
`design.md` (`## Context`, `## Decisions`) and a title-only `tasks.md`. Fill these three artifacts,
keeping every stub heading:

- **proposal.md** — what and why, carrying `## Why` and `## What changes`
- **design.md** — how, from the approved design, carrying `## Context`, `## Decisions` and
  `## Open questions` (the last two below), and four plan sections, each kept even when short:
  - `## Recommended approach` — the one chosen approach in prose, with a one-line reason for each
    rejected alternative pointing at its decision ID;
  - `## Critical files` — every file the change edits or depends on, by repository-relative path,
    one line on its role each;
  - `## Existing code to reuse` — the functions, scripts, guards and patterns the change builds
    on, by path (and symbol); `none` only with a one-line reason;
  - `## Verification` — how to prove the change works end to end: the commands, the run or
    scenario to exercise, and what is observed, beyond each task's own verify step.
- **tasks.md** — a checkbox scaffold; **writing-plans enriches it next**

**A spec edit is planned here, never written here.** A capability whose requirements the
change alters gets a task in `tasks.md` naming `<project>/spectre/specs/<capability>.md` in that
task's `**Files:**` field — the implementer writes and commits it on the change branch, in that
task's own commit.

### Decisions

`## Decisions` in `design.md` is sourced from the brainstorming dialogue — the approach the user
chose, the alternatives on the table, and the tradeoff that ruled each one out. **A design that
forced no choices records none.**

**Size never exempts a judgment call.** A small decision —
where a file lives, what a change or file is named, what a task's scope takes in — is recorded
like any other, its `**Considered:**` naming the rejected alternative and the one-line reason it
lost; the small ones are the ones a later reader re-litigates first, because nothing else in the
record explains why the obvious alternative was not taken.

```markdown
### <the decision>

**ID:** <kebab-case-slug>
**Status:** active
**Chosen:** <option> — <one-line rationale>
**Considered:** <other options, each with the tradeoff that ruled it out>
```

**Every entry carries its `**ID:**` line — mandatory, never omitted.** A decision is later cited by
ID or it is re-argued from scratch, and which decisions a review round will want to cite is not
knowable at creation, so every decision the design records is written with one. **ID** is assigned
once, at creation, and is **immutable** — the match key a later round uses to **supersede** a
decision: set the old entry's `**Status:**` to `superseded by <new-id>` and append a new entry with a
fresh ID. The superseded entry keeps its reasoning untouched and gains one appended
`**Superseded because:**` line naming the trigger that displaced it — the already-shipped change
the plan collided with, the measurement that disproved a premise, the operator's answer to a pivot
ask — so the record states why the supersession happened, not only that it did. **Never delete or
rewrite a superseded entry** — the `**Superseded because:**` line is the one addition a
supersession makes to it.

### Open questions

`## Open questions` sits beside `## Decisions` and is shaped like it — the record the convergence
offer above names. **A stage that left nothing open records none.** Empty, not absent.

```markdown
### <the question>

**ID:** <kebab-case-slug>
**Status:** open
**Why it is open:** <deferred by the operator, blocked on something external, …>
**What it affects:** <what would change depending on the answer>
```

**ID** is assigned once, immutable, and unique across `## Decisions` and `## Open questions`
together. A round that answers the question sets that entry's `**Status:**` to `answered by
<decision-id>` and adds the answering entry under `## Decisions`. **Never delete or rewrite an entry
once recorded.**

## D. Basic Workflow #3 — Writing plans

Invoke **superpowers:writing-plans** to enrich `<changeRoot>/tasks.md` to plan quality: exact
paths, verification commands, bite-sized steps, no placeholders. Run its self-review (spec
coverage, placeholder scan, type consistency) before finishing.

> **A task that adds a member to an existing class or interface** counts that type's members
> against the project's size-rule thresholds — detekt's `TooManyFunctions`, and the like — and,
> where the type has no headroom left, plans the split in that same task rather than a "no new
> type" constraint the lint then breaks (kan-999).

> **Write every click and capture in an end-to-end or fidelity spec on a structural locator.**
> Each such task's steps reach every control the spec clicks, types into or captures by its role
> and accessible name, its label, its text or its test id — never by a page coordinate, which
> names a position in one layout and lands on another control once any task shifts that layout.

> **Write a spec so it never waits out a production timing.** A scenario that depends on a
> production timeout, retry backoff or polling interval plans a test-only override that
> shortens it — the production value unchanged — in the task that adds the spec, and a suite
> whose specs are independent runs on the test runner's parallel workers.

> **Open a task that verifies a reported behaviour with the reproduce-first step.** When a task's
> job is confirming or refuting a report — an operator's defect report, a review or self-review
> finding, a claimed behaviour — its first step reproduces the report against the untouched tree,
> before any edit. Write the step into the task's own step text — "reproduce before touching
> anything; do not guess" — never as intent the task leaves implicit: the implementer reads its
> task, and a report reproduced before anything is touched is what closes a non-defect with no
> code change, while an edit made before reproducing destroys the evidence the task set out to
> gather.

> **List the rationale appendix a changed rule's reasoning lives in.** A step that changes a rule
> whose reasoning a `*-rationale.md` appendix records lists that appendix in its task's
> `**Files:**` and updates it.

**A task cites the decision it implements, never restates it.** When a task exists to implement
an entry of design.md's `## Decisions`, it names that entry by its `**ID:**` in a
`**Decision:** <id>` field and copies nothing of the entry's text — the decision lives once,
under `## Decisions`, and the citation is the link. A task
implementing no recorded decision writes no such field. A block of `**Decision:**` lines sits
between blank lines — one before the first line of the block and one after the last, never
glued to the `**Commit:**` subject above it.

### Plan review gate

The summary is plain prose, not a listing of `tasks.md`: what will be implemented and how — the
behaviour being added or changed, the approach taken, the seams and decisions that matter (data
shapes, boundaries, ordering constraints, what was deliberately left out) — in enough detail that
the operator can judge the logic without opening the plan. No per-task rows, no file, test or
commit fields. Then the `## Decision` block, verbatim, under it.
