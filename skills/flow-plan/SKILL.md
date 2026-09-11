---
name: flow-plan
description: Research mode - a thinking partner for exploring ideas, investigating problems, and clarifying requirements before or during a change. Touches no pipeline state. Use for /flow-plan.
---

Enter research mode. Think deeply. Visualize freely. Follow the conversation wherever it goes.

**This is a stance, not a workflow.** There are no fixed steps, no required sequence. One output
shape is fixed rather than optional (the step-by-step breakdown, below) — but nothing about how you
get there is scripted. You're a thinking partner helping the user explore.

---

## The Stance

- **Investigate, don't prescribe** — ask questions that emerge naturally, don't follow a script
- **Ground it** — read the actual code and the actual tree before theorizing
- **Never write application code** — if the user asks you to implement something, say so and point
  them at `/myflow-start` or `/myflow-do` instead
- **Never advance pipeline state** — no state file, no worktree, no branch
- **Commit only the research artifacts, once, at the end** — the note, its plan and its decision
  land together per **Landing the note** below; nothing else in this mode produces a commit

---

## The session does the thinking itself

`/flow-plan` runs entirely in the current session, on its own model, with any argument shape —
a Jira key, an existing change's name, a bare topic, or nothing at all. No subagent is dispatched:
the session itself reads the tree, runs the investigate-then-ask rounds and the convergence check
below, and writes the staging note or offers the `design.md` capture directly. Every "you"
elsewhere in this skill addresses the running session.

**Seed-lookup mechanics are shared, not restated.** When the topic already resolves to an existing
change (a key or name naming one under `<project>/spectre/changes/`), the exact-filename seed
check this skill's own captures feed into is `brainstorm-planner.md`'s **Seed from a staged
research note** (`skills/flow/brainstorm-planner.md`) — see that section's own **One shared
mechanism, not two copies** subsection for what stays shared between the two skills.

**No `flow record dispatch` call** — that record closes against a change's dispatch history, and
`/flow-plan` has no change to record against, dispatching nothing either.

---

## Reading the Tree

A spectre tree is a directory named `<project>/spectre/` holding
`<project>/spectre/specs/<capability>.md` flat files and `<project>/spectre/changes/<id>/` folders —
plain markdown, not a database. Read the same way under `<agents repo>/spectre/` when the research is
about this repository's own tree. Reading it means reading files:

- `spectre list` — the open changes
- `spectre list --specs` — the capabilities with a spec
- `<project>/spectre/specs/*.md` and `<project>/spectre/changes/<id>/` — read them directly with your
  file tools

No CLI grant is needed for any of this: listing and reading markdown takes no `allowed-tools` entry.

If the user names a change, read its folder for context before discussing it. If nothing exists yet
— a fresh tree, or a topic with no change — that's fine; think from the code and the conversation.

Also check `<project>/docs/superpowers/research/` for an existing staging note on the same topic (see
**Staging a Note** below) — a prior `/flow-plan` session may already have investigated part of
this ground.

---

## Go Deeper: Investigation and Question Depth

`/flow-plan` does noticeably more legwork than a single-pass answer before treating a topic as
understood — both halves, always:

- **Investigation depth:** don't stop at the first plausible answer. Broaden the search, read
  neighbouring files, check how the pattern is used elsewhere in the tree, confirm rather than
  assume.
- **Question depth:** don't accept a thin answer and move on. Ask a follow-up that tests or
  sharpens what the user just said before treating the topic as settled.

**Stopping rule** (concrete, not left to judgment mid-session):

1. Run at least **two rounds** of investigate-then-ask on each topic/thread before it can be
   considered understood — a round is one pass of codebase investigation followed by at least one
   clarifying question and the user's answer. A topic settled on round one has not gone deep enough;
   do a second pass even if the first answer sounded complete.
2. After round two, keep going only if a round surfaced something that changes the shape of the
   topic (a new constraint, a contradiction, an unread file that turns out to matter). Otherwise
   move to the convergence check.
3. **Convergence check, once per topic/thread:** ask an explicit "anything else to dig into on
   this?" before treating the topic as closed. A "no" (or equivalent) closes it; a "yes" starts
   another round.

This rule applies per topic/thread, not once for the whole session — a multi-topic conversation runs
the rule again each time the discussion moves to a new thread.

---

## What To Do

Depending on what the user brings, you might:

**Explore the problem space**
- Ask clarifying questions that emerge from what they said
- Challenge assumptions
- Reframe the problem
- Find analogies

**Investigate the codebase**
- Map existing architecture relevant to the discussion
- Find integration points
- Identify patterns already in use
- Surface hidden complexity

**Compare options**
- Name the options on the table
- Lay out the trade-offs of each, plainly
- Recommend a path, if asked

**Say when you don't know**
- If the answer depends on something not yet decided, or not yet knowable, say so rather than
  guessing
- Suggest a spike or a targeted investigation instead of a fabricated answer

Ask every pending question whose wording does not depend on another pending answer in one
**AskUserQuestion** call, up to four per call; a dependent question waits for the next turn. Don't
funnel the user through a fixed line of questioning — surface the
interesting directions and let them follow what resonates.

---

## Capturing the Outcome

There's no required ending, but there is a required *shape* once you do capture: every session that
produces a staging note or a `design.md` note produces it through the fixed structure below — never
as free-form prose (see **The Fixed Section Structure** and **The Step-by-Step Breakdown**).

Two capture destinations, depending on what exists:

- **An existing change is in scope** (the user named one, or the conversation is clearly about one
  already in `<project>/spectre/changes/`) — offer to write into that change's `design.md`, in the same fixed
  structure. Offer, don't auto-capture:
  - "That's a design decision — want it in design.md?"
  - "This is a new requirement — worth a note?"
- **No change exists yet, or the topic doesn't belong to one** — offer to write a **staging note**
  (see below) instead. This is the default destination for a topic with no home yet.

Creating a change is `/myflow-start`'s (or `/flow`'s) job, not this mode's — if the thinking is ready
to become a change, say so and point at that command rather than making one yourself.

### Staging a Note — the strict research-artifact path

When there's no existing change to write into, offer to capture the session as a staging note rather
than only letting it evaporate into the conversation:

- "Want this captured? I'd write it to `<project>/docs/superpowers/research/<destination>.md`."

The destination is **mandatory and deterministic**, never a free choice:

- **A Jira key is known** (passed as the argument, or resolved from the conversation per
  **Resolution (how `jiraIssue` is decided)**, `skills/flow-contracts/jira-integration.md`) — write to
  `<project>/docs/superpowers/research/<jira-key-lowercased>.md`, exactly. Never a descriptive
  suffix (`<key>-<slug>.md`) — a later `/flow <key>` finds the seed with one `test -f` on this
  exact path (**Seed from a staged research note, if one exists**, `skills/flow/brainstorm-planner.md`), and a
  suffixed filename would not be found by it.
- **No key** — create one before writing anything, since the key is the filename: a Jira issue of
  type **Task** in the project `## jira` names (`<project>/.flow/project.md`), summary the note's
  `<Topic>`, description one paragraph stating the topic, labels per **Labels on issues the
  pipeline creates** (`skills/flow-contracts/jira-integration.md`), so `AI-generated` alone. The
  created key is then the known key above, and the note's `Source:` line. The creation is a Jira
  write like any other: `## jira` absent or `none`, no Atlassian tooling, or a refused create is
  one `⚠ Jira: skipped — <reason>` line, and the note falls back to
  `<project>/docs/superpowers/research/<topic-slug>.md`, `<topic-slug>` a short kebab-case slug
  derived from the topic itself, with `Source: none`.

If a note already exists at the resolved destination, update it rather than creating a second file
for the same topic.

A staging note **seeds** a future `/flow` (or `/myflow-start`) brainstorming session on this topic —
it does not skip it; the plan and decision written beside it (**The plan and the decision**, below)
are what `/flow`'s writing-plans and Decide steps take in place of their own. `/flow-plan` never
runs `spectre new` and never creates a change itself; turning a staging note into a change is
always `/flow`'s call.

---

## The Fixed Section Structure

A captured note — staging note or a `design.md` addition alike — is not free-form prose. It follows
this section structure every time, so a later reader (human or `/flow`'s brainstorming stage) can
parse it mechanically instead of re-reading loose prose. At minimum:

1. **A source line**, the first line of body content (after the title): `Source: <value>` where
   `<value>` is a Jira key, a Jira/ticket URL, or the literal word `none` when the session had no
   ticket behind it.
2. **One section per topic/thread discussed**, each a `##`-level heading with a short descriptive
   title, holding whatever prose, lists, or code excerpts the discussion produced for that thread.
   Order sections in the order the threads came up.
3. **A step-by-step breakdown section**, always present (see below) — this is the default output of
   every research session, never opt-in.

Optional, add only when relevant: an **Open / undesigned** section listing what the session
surfaced but didn't resolve — one bullet per open item.

### Template

```markdown
# <Topic> — research notes

Source: <Jira key, ticket URL, or "none">

## 1. <First thread title>

<prose, lists, findings>

## 2. <Second thread title>

<prose, lists, findings>

## Step-by-step breakdown

### <Name of item 1>

**What:** <one or two sentences — what it is or does>
**Why:** <one or two sentences — the reason it exists or was chosen>
**Uses:** <what it depends on or calls — files, skills, scripts, CLIs, services; "none" if nothing>

### <Name of item 2>

**What:** ...
**Why:** ...
**Uses:** ...

## Open / undesigned

- <item not resolved this session>
```

## The Step-by-Step Breakdown

The breakdown section is a `name`/`what`/`why`/`uses` structural analysis of the topic under
discussion, and it is the **default** output of every `/flow-plan` session — produce it whether
or not the user explicitly asked for it, whenever a session reaches a captured note. This overrides
any instinct to treat it as an extra the user has to request.

Rules for filling it in:

- One `###` subsection per distinct item the session identified — a pipeline stage, a component, a
  design option, a mechanism — whatever unit of analysis the topic decomposed into.
- Each subsection carries exactly the three labelled fields above, in that order:
  `**What:**`, `**Why:**`, `**Uses:**`. Keep each to one or two sentences; a longer explanation
  belongs in the numbered thread section above, cross-referenced from here rather than repeated.
- `**Uses:**` names concrete things (file paths, skill names, script names, CLI commands, external
  services) — `none` is a valid, honest answer, not a placeholder to avoid filling in.
- If the topic is small enough that a full breakdown would only have one item, still include the
  section with that one item — the section's presence is what's fixed, not a minimum item count.

## The plan and the decision

A captured **staging note** — never a `design.md` addition, since an existing change already has
both — also produces, when the session finishes, two files beside it. `<stem>` is the note's own
filename without `.md`, so the pair is found from the note's path by one exact test each.

1. **`<project>/docs/superpowers/research/<stem>/tasks.md` — the plan.** Exactly the shape
   **D. Basic Workflow #3 — Writing plans** (`skills/flow/brainstorm-planner.md`) defines: task
   and step lines, the field family, the two header lines, the provenance and build-green tags —
   cited, never restated. One task per breakdown item is the natural starting grain. Run
   `check-plan-shape.sh <that path>` and fix any hit before finishing.
2. **`<project>/docs/superpowers/research/<stem>/decision.json` — the dynamic decision.** Reached
   exactly as **Decide** (`skills/flow/brainstorm-planner.md`) reaches it and written in the JSON
   shape that section defines: `plan-class.sh <project>/docs/superpowers/research/<stem>/tasks.md <repos>`, `<repos>` being the number
   of distinct repository roots the plan's `**Files:**` fall under (the project's `## apps` table;
   `1` when every path is in this one); the three toggles resolved per **Model resolution**
   (`skills/flow/SKILL.md`) against the main checkout alone, since no worktree exists;
   `DEFAULT_MODEL` from `flow settings get`; the same tree and rolls. The rolls seed from `<stem>`
   rather than from a change name, deliberately: the recorded decision is what `/flow` records, so
   it is stable by being written down, not by being re-rolled.
3. **A `## Decision` section in the note**, last, carrying the printed table **Decide**'s own
   output shape defines, so a reader sees the decision without opening the JSON.

`/flow`'s seed step takes both files in place of its own writing-plans and Decide work and deletes
them with the note once adopted (**Seed from a staged research note, if one exists**,
`skills/flow/brainstorm-planner.md`). `/flow-fast` reads neither — it plans and decides on its
own (**Dynamic decisions**, `skills/flow-fast/SKILL.md`). Nothing here marks a stage, records to the store, or creates a change: the pair is a
research artifact until `/flow` adopts it.

## Landing the note

A captured staging note is landed in the same session, once the note, `tasks.md` and
`decision.json` are all written and `check-plan-shape.sh` is clean — never a `design.md`
addition, which its change's own commits carry. The main checkout is where `/flow-plan` runs, so
that is where it lands, on `<default-branch>` (`## apps`, `<project>/.flow/project.md`; the
remote's `HEAD` branch when the table names none):

```bash
git -C <project> add docs/superpowers/research/<stem>.md docs/superpowers/research/<stem>/tasks.md docs/superpowers/research/<stem>/decision.json
git -C <project> commit -m "docs(research): <stem> research note, plan and decision"
git -C <project> pull --rebase origin <default-branch>
git -C <project> push origin <default-branch>
```

Exactly those three paths are staged — never `-A`, never anything the operator had pending. The
checkout must be on `<default-branch>` before the `add`; on any other branch, make no commit,
print that branch and the three paths, and stop. A pull that conflicts is resolved in place per the
**Conflict** bullet of **Sync the branch onto the base**
(`skills/flow-contracts/finish-contract-run1.md`). A push the remote rejects (branch protection, a
non-fast-forward after the pull) leaves the commit local, and the run ends by saying so and naming
the commit — a protected default branch is landed by the operator, never retried around
(`~/.claude/rules/no-direct-pushes-to-main.md`). End by naming the landed commit and the Jira
key.

---

## Guardrails

- **Don't write application code** — reading, searching, and discussing are fine; implementing is not
- **Don't touch pipeline state** — no state file, no `state:` transition
- **Don't commit anything but the three research paths**, and only at **Landing the note**
- **Don't run `spectre new`** — that's `/myflow-start`'s (or `/flow`'s) call, not this mode's
- **Don't run `spectre archive`** — that's `/myflow-finish`'s call, not this mode's
- **Don't fake understanding** — if something is unclear, dig deeper instead of assuming; see **Go
  Deeper** above for the concrete stopping rule
- **Don't force structure on the conversation** — let the shape of the discussion emerge; the fixed
  structure applies to what gets *captured*, never to how the discussion itself unfolds
- **Don't skip the step-by-step breakdown** — it's the default capture shape, not an opt-in
- **Don't leave a staging note without its plan and decision** — the pair beside it is what makes
  the note ready for `/flow`, not an extra
