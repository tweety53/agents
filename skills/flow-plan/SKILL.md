---
name: flow-plan
description: Research mode - a thinking partner for exploring ideas, investigating problems, and clarifying requirements before or during a change. A captured session creates the change at STARTED and stops there. Use for /flow-plan.
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
- **Never advance past `STARTED`** — a captured session creates the change and stops there;
  implementing it is `/flow <name>`'s run
- **Commit only the planning artifacts, once, at capture** — `<project>/spectre/changes/<name>/`
  on `spectre/<name>` per **Capturing a new change** below; nothing else in this mode produces a
  commit
- **Don't run `spectre archive`** — that's bare `/flow`'s call, not this mode's

---

## The session does the thinking itself

`/flow-plan` runs entirely in the current session, on its own model, with any argument shape —
a Jira key, an existing change's name, a bare topic, or nothing at all. No subagent is dispatched:
the session itself reads the tree, runs the investigate-then-ask rounds and the convergence check
below, and creates the change or offers the `design.md` capture directly. Every "you" elsewhere
in this skill addresses the running session.

**Generate this session's token once, right here — `fp-<literal-token>`, a short unique literal
string, exactly as `/flow` generates `mf-<literal-token>` (**Stage keys**, `skills/flow/SKILL.md`)
— and mark the session as soon as the Jira key is known:**

```bash
flow stage begin -command '/flow-plan' -stage plan.session -harness <harness> -session-token fp-<literal-token> -jira-key <KEY>
```

`<KEY>` is the resolved Jira key, uppercase, exactly as **Resolution (how `jiraIssue` is decided)**
(`skills/flow-contracts/jira-integration.md`) resolves it. When the key is only created later, by
**Capturing a new change**, mark at that point instead. A session that ends with no key marks
nothing.

**This mark never blocks.** A refused or unreachable store prints one line and the session
continues; never branch on `flow stage begin`'s or `flow stage end`'s exit code. Close the session
with exactly one `stage end`, on whichever exit path this session takes — `started` when a change
was created, `captured` when the session wrote into an existing change's artifacts, `abandoned`
otherwise, including a session that just stops.

**`plan.session` is the one mark this mode makes.** The `flow.*` keys a capture executes
(**Capturing a new change**, below) are `/flow`'s rows in the Level 1 table and are not marked
here; the store attaches the open plan session to the change the moment the capture's own
`STARTED` write creates it. **No `flow record dispatch` call** either — this mode dispatches
nothing.

**Every read is of the main checkout, read-only.** This mode writes nothing until a capture:
**Capturing a new change** creates `<project>/.worktrees/<name>` through `/flow`'s kickoff and
writes there; a `design.md` addition is written in that change's own worktree —
`<project>/.worktrees/<name>`, found from `git worktree list`. The main checkout is never checked
out, staged, committed or written, whatever branch it sits on.

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

---

## Closing a topic

A topic is closed by an explicit convergence check, once per topic/thread: ask "anything else to
dig into on this?" before treating it as settled. A "no" (or equivalent) closes it; a "yes" starts
another round of investigation and questions. Ground each round in the code and the tree rather
than in inference, and confirm a pattern's use elsewhere before generalising from one site.

---

## Asking

Ask every pending question whose wording does not depend on another pending answer in one
**AskUserQuestion** call, up to four per call; a dependent question waits for the next turn. Don't
funnel the user through a fixed line of questioning — surface the interesting directions and let
them follow what resonates.

---

## Capturing the Outcome

There's no required ending, but there is a required *shape* once you do capture: every session that
produces a `design.md` — a new change's or an addition to an existing one's — produces it through
the fixed structure below, never as free-form prose (see **The Fixed Section Structure** and
**The Step-by-Step Breakdown**).

Two capture destinations, depending on what exists:

- **An existing change is in scope** (the user named one, the conversation is clearly about one
  already in `<project>/spectre/changes/`, or the key's candidate lookup in **Capturing a new
  change** finds one) — offer to write into that change's `design.md`, in the same fixed
  structure. Offer, don't auto-capture:
  - "That's a design decision — want it in design.md?"
  - "This is a new requirement — worth a note?"

  Once written, close the session's mark (the token generated at session start, above) with
  `-outcome captured`:

  ```bash
  flow stage end -command '/flow-plan' -stage plan.session -jira-key <KEY> -outcome captured
  ```
- **No change exists yet, or the topic doesn't belong to one** — offer to create the change
  (**Capturing a new change**, below). This is the default destination for a topic with no home
  yet.

A session that ends without either destination — no capture offered, or the offer declined — closes
the same mark with `-outcome abandoned`:

```bash
flow stage end -command '/flow-plan' -stage plan.session -jira-key <KEY> -outcome abandoned
```

### Capturing a new change

When there's no existing change to write into, offer to capture the session as a change at
`STARTED` rather than only letting it evaporate into the conversation — one **AskUserQuestion**:

> **Capture this as the change `<name>` at `STARTED`, planned and ready for `/flow <name>`?**
> - **Yes — create it**
> - **No — leave it in the conversation**

**Nothing below runs before that Yes.** The `STARTED` write, the worktree and the artifacts are
this answer's consequence; **No** is the abandoned path (**Capturing the Outcome**, above). The
plan review gate at step 6 then gates the commit and push alone — a change that reached it exists
already, at `STARTED`, and a **No** there revises the plan rather than removing the change.

**The key comes first, because it names the change.** A Jira key is known when it was passed as
the argument or resolved from the conversation per **Resolution (how `jiraIssue` is decided)**
(`skills/flow-contracts/jira-integration.md`). With no key, create one before anything else: a
Jira issue of type **Task** in the project `## jira` names (`<project>/.flow/project.md`), summary
the session's `<Topic>`, description one paragraph stating the topic, labels per **Labels on
issues the pipeline creates** (`skills/flow-contracts/jira-integration.md`), so `AI-generated`
alone. The created key is then the known key, and this session's `plan.session` mark fires now
when it did not at start. The creation is a Jira write like any other: `## jira` absent or
`none`, no Atlassian tooling, or a refused create is one `⚠ Jira: skipped — <reason>` line, and
the change is unlinked — named by its descriptive slug alone, as `/flow` names an unlinked change.

**Then run `/flow`'s own planning sections, in this order, each as written there** — this mode
restates none of them, and every `mf-<literal-token>` they show is this session's
`fp-<literal-token>`:

1. **A. Resolve the change and write `STARTED`** (`skills/flow/brainstorm.md`) — the key is
   already resolved above, so its candidate lookup, the name, the In Progress transition and the
   `STARTED` write are what runs. A lookup that finds an existing change for this key is the
   existing-change destination above, never a second change.
2. The kickoff steps 1–5 in the same section — `<project>/.worktrees/<name>` on `spectre/<name>`,
   pushed. `flow.kickoff` itself is not marked (**The session does the thinking itself**, above).
3. **C. Create the change and its artifacts** (`skills/flow/brainstorm-planner.md`) — `spectre
   new` in the worktree, then the three artifacts. `design.md`'s body is **The Fixed Section
   Structure** below — the thread sections and the step-by-step breakdown — followed by C's
   `## Decisions`, filled from the answers the convergence rounds settled, and its `## Open
   questions`, present and empty. `proposal.md`'s `## Why` and `## What changes` come from the
   same session. `tasks.md` starts as one task per breakdown item.
4. **D. Basic Workflow #3 — Writing plans** (`skills/flow/brainstorm-planner.md`) — the
   enrichment, its guards and `flow tasks count`, as written.
5. **Decide** (`skills/flow/brainstorm-planner.md`) — `plan-class.sh` on `<changeRoot>/tasks.md`
   with `<repos>` the number of distinct repository roots the plan's `**Files:**` fall under
   (the project's `## apps` table; `1` when every path is in this one); the three toggles per
   **Model resolution** (`skills/flow/SKILL.md`) against the main checkout alone;
   `DEFAULT_MODEL` from `flow settings get`; the JSON at
   `<abs-worktree>/.superpowers/sdd/decision.json`; the `## Decision` block printed. Record it
   against the change:

   ```bash
   flow record decision -change <name> -session-token fp-<literal-token> -file <abs-worktree>/.superpowers/sdd/decision.json
   ```

6. **Plan review gate** (`skills/flow/brainstorm-planner.md`) — the prose summary, the
   `## Decision` block, then **Push artifacts?** with **Yes** / **No (plan needs updates)**.
   Mandatory: **No** revises and re-decides per that section, recording each re-decision through
   the call above, and the commit below happens only on **Yes**.

Every guard those sections invoke resolves per **Guard resolution**
(`skills/flow-contracts/pipeline.md`) against this skill's own scripts directory, which carries
each one those sections run: `check-planning-commit-location.sh <worktree> <name>` (the commit
below), `check-worktree-location.sh <project>` and
`project-get.sh <project> <key>` (kickoff), `check-plan-shape.sh <tasks.md>` (D and the gate),
`plan-class.sh <tasks.md> <repos>`, `plan-dispatch-bundles.sh <tasks.md>` and
`plan-dispatch-groups.sh <tasks.md>` (Decide).

**On Yes, commit and push the planning artifacts** from the change worktree:

```bash
check-planning-commit-location.sh <worktree> <name> \
  && git -C <worktree> add spectre/changes/<name> \
  && git -C <worktree> commit -m "chore(spectre): plan" \
  && git -C <worktree> push origin spectre/<name>
```

The guard's non-zero exit stops the capture before anything is staged — the commit lands only in
the change worktree on `spectre/<name>` (**Planning commits**,
`skills/flow-contracts/git-boundaries.md`, canonical for its verdicts and its by-hand fallback).
Exactly that directory is staged — never `-A`, never anything else in the worktree. A push the
remote rejects leaves the commit on `spectre/<name>` in the worktree; say so and name the commit,
the branch and the worktree path. The worktree is kept either way — it is the change's, resumed by
`/flow <name>` exactly as any `STARTED` change's is (**Resuming at `STARTED`**,
`skills/flow/brainstorm.md`).

End by naming the change, the Jira key, the commit and the worktree's absolute path, and close
the session's mark:

```bash
flow stage end -command '/flow-plan' -stage plan.session -jira-key <KEY> -outcome started
```

The last line is the next command, bare:

```text
Next:
/flow <name>
```

---

## The Fixed Section Structure

A captured `design.md` — a new change's body or an addition to an existing one's — is not
free-form prose. It follows this section structure every time, so a later reader (human or
`/flow`'s implement phase) can parse it mechanically instead of re-reading loose prose. At
minimum:

1. **One section per topic/thread discussed**, each a `##`-level heading with a short descriptive
   title, holding whatever prose, lists, or code excerpts the discussion produced for that thread.
   Order sections in the order the threads came up.
2. **A step-by-step breakdown section**, always present (see below).

**A captured design carries no open items.** Everything the session surfaced is decided before
the artifacts are written — decided by the user through the convergence round's questions, or by
you where the call is routine, and recorded as a decision in its thread's section with the
rejected alternative named, and under `## Decisions` with its `**ID:**`. There is no **Open /
undesigned** section, and `## Open questions` is present and empty: an item still open when the
convergence check closes is one more **AskUserQuestion** round, never a bullet deferred to
`/flow <name>`, which resumes the change as planned and asks nothing again.

### Template

```markdown
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
or not the user explicitly asked for it, whenever a session reaches a captured design.

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
