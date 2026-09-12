---
name: flow-plan
description: Research mode - a thinking partner for exploring ideas, investigating problems, and clarifying requirements before or during a change. Touches no pipeline state. Use for /flow-plan.
---

Enter research mode. Visualize freely. Follow the conversation wherever it goes.

**This is a stance, not a workflow.** There are no fixed steps, no required sequence. One output
shape is fixed rather than optional (the step-by-step breakdown, below) — but nothing about how you
get there is scripted. You're a thinking partner helping the user explore.

---

## The Stance

- **Investigate, don't prescribe** — ask questions that emerge naturally, don't follow a script
- **Ground it** — read the actual code and the actual tree before theorizing
- **Never write application code** — if the user asks you to implement something, say so and point
  them at `/flow`'s creating run or `/flow`'s implement phase instead
- **Never advance pipeline state** — no state file; the research worktree below is the only
  branch it creates, and it removes it before the session ends
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

**Generate this session's token once, right here — `fp-<literal-token>`, a short unique literal
string, exactly as `/flow` generates `mf-<literal-token>` (**Stage keys**, `skills/flow/SKILL.md`)
— and mark the session as soon as the Jira key is known:**

```bash
flow stage begin -command '/flow-plan' -stage plan.session -harness <harness> -session-token fp-<literal-token> -jira-key <KEY>
```

`<KEY>` is the resolved Jira key, uppercase, exactly as **Resolution (how `jiraIssue` is decided)**
(`skills/flow-contracts/jira-integration.md`) resolves it. When the key is only created later, by
**Staging a Note**, mark at that point instead. A session that ends with no key marks nothing.

**This mark never blocks.** A refused or unreachable store prints one line and the session
continues; never branch on `flow stage begin`'s or `flow stage end`'s exit code. Close the session
with exactly one `stage end`, on whichever exit path this session takes — `staged` when a note was
written, `captured` when the session wrote into an existing change's artifacts, `abandoned`
otherwise, including a session that just stops.

**No `flow record dispatch` call** — that record closes against a change's dispatch history, and
`/flow-plan` has no change to record against, dispatching nothing either. The one mark this mode
makes is `plan.session`, above, recorded against the Jira key; the store attaches it to the
change `/flow <KEY>` later creates.

## The research worktree

**Create it first, before the tree is read**, once `<stem>` is fixed (the linked Jira key
lowercased, or the bare session's own slug — the same stem the seed lookup keys on). From the main
checkout, with `<default-branch>` the branch `origin/HEAD` points at:

```bash
git -C <project> fetch origin
grep -qx '.worktrees/' <project>/.git/info/exclude 2>/dev/null || echo '.worktrees/' >> <project>/.git/info/exclude
git worktree add <project>/.worktrees/_plan-<stem> -b plan-<stem> origin/<default-branch>
```

`<worktree>` below is that path. Every read of the tree, every file this mode writes and its one
commit happen there; the main checkout is never read, checked out, staged, committed or written,
whatever branch it sits on. **A `design.md` addition is the one write outside it**: an existing
change's `design.md` lives in that change's own worktree — `<project>/.worktrees/<name>`, found
from `git worktree list` — and is written there.

**Remove it before the session ends, whatever the outcome** — a landed note, a `design.md`
addition, or a session that captured nothing — except after a rejected push (**Landing the
note** below), which keeps it for the operator:

```bash
git -C <project> worktree remove --force <project>/.worktrees/_plan-<stem>
git -C <project> branch -D plan-<stem>
```

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

Also check `<project>/docs/research/` for an existing staging note on the same topic (see
**Staging a Note** below) — a prior `/flow-plan` session may already have investigated part of
this ground.

---

## Closing a topic

A topic is closed by an explicit convergence check, once per topic/thread: ask "anything else to
dig into on this?" before treating it as settled. A "no" (or equivalent) closes it; a "yes" starts
another round of investigation and questions. Ground each round in the code and the tree rather
than in inference, and confirm a pattern's use elsewhere before generalising from one site.

This applies per topic/thread, not once for the whole session — a multi-topic conversation runs the
check again each time the discussion moves to a new thread.

---

## Asking

Ask every pending question whose wording does not depend on another pending answer in one
**AskUserQuestion** call, up to four per call; a dependent question waits for the next turn. Don't
funnel the user through a fixed line of questioning — surface the interesting directions and let
them follow what resonates.

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

  Once written, close the session's mark (the token generated at session start, above) with
  `-outcome captured`:

  ```bash
  flow stage end -command '/flow-plan' -stage plan.session -jira-key <KEY> -outcome captured
  ```
- **No change exists yet, or the topic doesn't belong to one** — offer to write a **staging note**
  (see below) instead. This is the default destination for a topic with no home yet.

Creating a change is `/flow`'s job, not this mode's — if the thinking is ready
to become a change, say so and point at that command rather than making one yourself.

A session that ends without either destination — no capture offered, or the offer declined — closes
the same mark with `-outcome abandoned`:

```bash
flow stage end -command '/flow-plan' -stage plan.session -jira-key <KEY> -outcome abandoned
```

### Staging a Note — the strict research-artifact path

When there's no existing change to write into, offer to capture the session as a staging note rather
than only letting it evaporate into the conversation:

- "Want this captured? I'd write it to `<project>/docs/research/<destination>.md`."

The destination is **mandatory and deterministic**, never a free choice:

- **A Jira key is known** (passed as the argument, or resolved from the conversation per
  **Resolution (how `jiraIssue` is decided)**, `skills/flow-contracts/jira-integration.md`) — write to
  `<project>/docs/research/<jira-key-lowercased>.md`, exactly. Never a descriptive
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
  `<project>/docs/research/<topic-slug>.md`, `<topic-slug>` a short kebab-case slug
  derived from the topic itself, with `Source: none`.

If a note already exists at the resolved destination, update it rather than creating a second file
for the same topic.

A staging note **seeds** a future `/flow` brainstorming session on this topic —
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

**A captured note carries no open items.** Everything the session surfaced is decided before
the note is written — decided by the user through the convergence round's questions, or by you
where the call is routine, and recorded as a decision in its thread's section with the
rejected alternative named. There is no **Open / undesigned** section: an item still open when
the convergence check closes is one more **AskUserQuestion** round, never a bullet deferred to
`/flow`, whose seed step adopts the plan and decision as written and asks nothing again.

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
```

## The Step-by-Step Breakdown

The breakdown section is a `name`/`what`/`why`/`uses` structural analysis of the topic under
discussion, and it is the **default** output of every `/flow-plan` session — produce it whether
or not the user explicitly asked for it, whenever a session reaches a captured note.

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

1. **`<project>/docs/research/<stem>/tasks.md` — the plan.** Exactly the shape
   **D. Basic Workflow #3 — Writing plans** (`skills/flow/brainstorm-planner.md`) defines: task
   and step lines, the field family, the two header lines, the provenance and build-green tags —
   cited, never restated. One task per breakdown item is the natural starting grain. Run
   `check-plan-shape.sh <that path>` and fix any hit before finishing.
2. **`<project>/docs/research/<stem>/decision.json` — the dynamic decision.** Reached
   exactly as **Decide** (`skills/flow/brainstorm-planner.md`) reaches it and written in the JSON
   shape that section defines: `plan-class.sh <project>/docs/research/<stem>/tasks.md <repos>`, `<repos>` being the number
   of distinct repository roots the plan's `**Files:**` fall under (the project's `## apps` table;
   `1` when every path is in this one); the three toggles resolved per **Model resolution**
   (`skills/flow/SKILL.md`) against the research worktree alone;
   `DEFAULT_MODEL` from `flow settings get`; the same tree and rolls. The rolls seed from `<stem>`
   rather than from a change name, deliberately: the recorded decision is what `/flow` records, so
   it is stable by being written down, not by being re-rolled.
3. **A `## Decision` section in the note**, last, carrying the printed block **Decide**'s own
   output shape defines, so a reader sees the decision without opening the JSON.

`/flow`'s seed step takes both files in place of its own writing-plans and Decide work and deletes
them with the note once adopted (**Seed from a staged research note, if one exists**,
`skills/flow/brainstorm-planner.md`). `/flow-fast` reads neither — it plans and decides on its
own (**Dynamic decisions**, `skills/flow-fast/SKILL.md`). Nothing here creates a change or records
a dispatch; the session's own `plan.session` mark is the only store write, and the pair is a
research artifact until `/flow` adopts it.

## Landing the note

A captured staging note is landed in the same session, once the note, `tasks.md` and
`decision.json` are all written and `check-plan-shape.sh` is clean — never a `design.md`
addition, which its change's own commits carry. It is committed in the research worktree and
pushed from there onto `<default-branch>` (`## apps`, `<project>/.flow/project.md`; the remote's
`HEAD` branch when the table names none):

```bash
git -C <worktree> add docs/research/<stem>.md docs/research/<stem>/tasks.md docs/research/<stem>/decision.json
git -C <worktree> commit -m "docs(research): <stem> research note, plan and decision"
git -C <worktree> pull --rebase origin <default-branch>
git -C <worktree> push origin HEAD:<default-branch>
```

Exactly those three paths are staged — never `-A`, never anything else in the worktree. A pull
that conflicts is resolved in place per the **Conflict** bullet of **Sync the branch onto the
base** (`skills/flow-contracts/finish-contract-run1.md`). A push the remote rejects (branch
protection, a non-fast-forward after the pull) leaves the commit on `plan-<stem>`, and the run
ends by saying so and naming the commit, the branch and the worktree path, both kept — a
protected default branch is landed by the operator, never retried around. After a push that succeeded, remove the worktree
and branch (**The research worktree** above). End by naming the landed commit and the Jira key.

**Close the session's mark here, whichever way the push went** — the note itself was written
either way:

```bash
flow stage end -command '/flow-plan' -stage plan.session -jira-key <KEY> -outcome staged
```

---

## Guardrails

- **Don't write application code** — reading, searching, and discussing are fine; implementing is not
- **Don't touch pipeline state** — no state file, no `state:` transition
- **Don't commit anything but the three research paths**, and only at **Landing the note**
- **Don't run `spectre new`** — that's `/flow`'s call, not this mode's
- **Don't run `spectre archive`** — that's bare `/flow`'s call, not this mode's
- **Don't fake understanding** — if something is unclear, dig deeper instead of assuming; see **Go
  Deeper** above for the concrete stopping rule
- **Don't force structure on the conversation** — let the shape of the discussion emerge; the fixed
  structure applies to what gets *captured*, never to how the discussion itself unfolds
- **Don't skip the step-by-step breakdown** — it's the default capture shape, not an opt-in
- **Don't land a note with an open item** — ask another round instead; the note records decisions,
  never deferrals (see **The Fixed Section Structure**)
- **Don't leave a staging note without its plan and decision** — the pair beside it is what makes
  the note ready for `/flow`, not an extra
