# Brainstorm and plan — the planner's sections

Sections **B**, **C** and **D** of the brainstorming phase, read by the planner subagent that
**Dispatch the planner** (`skills/flow/brainstorm.md`) sends out; every "you" below addresses that
planner. The parent never reads this file.

## B. Basic Workflow #1 — Brainstorming

**This section's stage marks are run by the parent, not the planner.** The `flow.brainstorm` begin
mark now lives in **Dispatch the planner** (`skills/flow/brainstorm.md`), and the `flow.brainstorm` end /
`flow.design-approval` begin/end marks under **Convergence** below stay exactly where they are,
run by the parent around the relayed HARD GATE approval. Everywhere else in this section — and in
**C** and **D** below — that addresses "you" means the dispatched planner: the seeded-note
discovery and its deletion, the checklist, and the convergence loop are the planner's own work,
relayed back through the returns **Dispatch the planner** (`skills/flow/brainstorm.md`) describes.

### Seed from a staged research note, if one exists

**Before starting the interactive checklist**, look in `<project>/docs/superpowers/research/` for a note using
this exact-filename rule, in order — **no fuzzy or substring matching beyond the wildcard fallback
named below**.

- **The change carries a linked Jira issue:** check `<jira-key-lowercased>.md` first — this is the
  exact shape `skills/flow-research/SKILL.md`'s "Staging a Note" produces, since its `<topic-slug>`
  "reuse[s] the change's eventual id (e.g. a Jira key like `kan-410`) when one is already known," and
  that skill writes or updates that single file rather than ever creating a second one for the same
  topic-slug — so the mechanized path alone never produces two files this exact check could confuse.
  If that exact file is absent, fall back to `<jira-key-lowercased>-*.md` — a wider glob kept for
  notes written before that convention existed (this change's own fixture,
  `<project>/docs/superpowers/research/kan-326-myflow-rework.md`, is one: hand-written pre-mechanization, so it
  carries a descriptive suffix the mechanized writer no longer adds). **Do not** also check
  `<name>.md` here: a Jira-linked change's resolved `<name>` is always `<key>-<slug>`
  (**Change naming**, `skills/flow-contracts/jira-integration.md`), which a research note's own
  topic-slug — the key alone, or a topic-derived slug chosen before any change existed — will not
  equal.
- **No linked issue:** check `<name>.md`, where `<name>` is the change's own resolved (slug-only)
  name. This is the "slug derived from the topic itself" case `skills/flow-research/SKILL.md`
  describes — an exact match only when the change's later slug happens to reuse the research
  session's own topic wording.

The exact-filename check is unambiguous by construction. The `-*.md` fallback is a genuine wildcard,
and nothing about it (hand-written notes predate the mechanized writer and are named freely) rules
out two matches — this change's own worktree could, in principle, hold both `kan-326-foo.md` and
`kan-326-bar.md`. **If the glob matches more than one file, do not silently pick one.** List every
match and ask the operator which to seed from (or neither):

> **Found more than one staged research note for this change: `<path-1>`, `<path-2>`, …. Which one
> should seed this round?**
> - **One of the listed paths** — seed from that note, per **If found** below
> - **None of them** — proceed with no seed

If the glob matches exactly one file, seed from it without asking. A topic captured under different
wording than the change later resolves to will not be found by this rule at all — that is an
accepted limit of a deterministic, exact-filename check, not a defect to patch with fuzzy
heuristics: guessing which note "probably" matches risks seeding from the wrong topic silently.

**If found**, parse it against **The Fixed Section Structure** (`skills/flow-research/SKILL.md`)
rather than reading it as loose prose — extract, by name: the `Source:` line, each `##`-level
topic/thread section, and the step-by-step breakdown section's `###` items. **A note missing any of
these three required elements is reported by name** (e.g. "the note has no step-by-step breakdown
section") and the note is still treated as a partial seed, never silently treated as empty and never
discarded outright for being incomplete. Present the parsed structure — not the raw file — to the
operator as the starting point for this round:

> "Found a staged research note for this topic at `<path>`. Here's what it already covers: <summary
> of its sections and step-by-step breakdown>. Starting the brainstorming round from this."

**Seeding never skips the interactive round.** Present the note, then run the full checklist below
exactly as if no note existed — the note answers what it answers, and the checklist still surfaces
what it does not, confirms what it does, and gives the operator a chance to correct it.

**Once the note's content is adopted into this change's own artifacts** (its design content folded
into `design.md`, its decisions and open questions carried in per **C** below), **delete the staging
note** rather than leaving it in place. Delete
it as part of **C**'s artifact-creation commit — it is a planning path, staged and committed the same
way the rest of `<project>/spectre/changes/` is: **C**'s "Delete the adopted staging note" step
removes it. If the note only partially seeded this round (a note covering one of several threads the
round expanded on), still delete it once adopted: its useful content now lives in `design.md`, which
is the canonical location from that point on.

**Carry the seeded note's path forward to C.** If a note was found and seeded above, remember its
path (e.g. `<project>/docs/superpowers/research/kan-326.md`) — **C** deletes exactly that file and only when
this note-found condition holds.

### The checklist

Invoke **superpowers:brainstorming** in full: checklist items 1–8, ending with the user approving
the design.

- Save the design to `<project>/docs/superpowers/specs/YYYY-MM-DD-<name>-design.md` and stage it
  when the brainstorming skill requires it.
- **HARD GATE:** do not run `spectre new` until the user approves the design. Approval is the
  merged confirm's first option under **Convergence** below; no separate approval question is
  asked.
- For multi-subsystem work, decompose before proposing.
- The design presentation does **not** end a section, or the whole design, with a "does this look
  right?" question — present the section(s) and proceed directly, section to section and then into
  artifact creation, unless the operator raises an objection during or after that presentation. This
  is a scoped override of `superpowers:brainstorming`'s hard design-approval gate, `/flow` only.

The approved design is the source for the change's `design.md`; adapt its format, never duplicate a
conflicting design.

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
names something opens another round. **This is the one prompt in this file where the safe default
and the recommended option differ, and deliberately so** — shape per Operator prompts
(`skills/flow-contracts/operator-prompts.md`): silence or a stalled prompt from an operator who is
present is not "approve the design and move on," and defaults to another round rather than to the
recommended choice. Print `⚠ another round — no explicit answer` when this default fires.

*Revise* is a round — it counts toward the third-round offer below exactly as *Another round*
does — and differs only in what the planner's next turn opens with: the changed design section(s),
re-presented before the next confirm, in place of new questions.

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
see **Stage exit — never the command's own judgment** (`skills/flow-contracts/pipeline.md`).

The explicit **approve the design and move on** answer is at once the convergence exit that closes
the checklist and the design approval the HARD GATE requires; the parent still marks
`flow.brainstorm` end, then `flow.design-approval` begin and end, around that one relayed answer:

```bash
flow stage end   -command '/flow' -stage flow.brainstorm -outcome completed <name>
flow stage begin -command '/flow' -stage flow.design-approval -harness <harness> -session-token mf-<literal-token> <name>
# … the operator's approve-and-move-on answer, relayed — this is the HARD GATE above …
flow stage end   -command '/flow' -stage flow.design-approval -outcome completed <name>
```

## C. Create the change and its artifacts

**This section's stage marks are run by the parent**, around the planner's `## Artifacts` return —
see **Dispatch the planner** (`skills/flow/brainstorm.md`). Everywhere below that addresses "you" means the planner:
`spectre new`, the three artifacts, and the staging-note deletion are its own work.

```bash
flow stage begin -command '/flow' -stage flow.create-artifacts -harness <harness> -session-token mf-<literal-token> <name>
# … the parent resumes the planner via SendMessage; everything from here is the planner's own turn …
spectre new "<name>"
```

`spectre new` scaffolds `<project>/spectre/changes/<name>/`, and refuses three ways: exit `2` and
*no tree found* when the project holds no `<project>/spectre/` tree at all; exit `2` and `invalid
change id` when `<name>` is not a single flat directory name; and exit `1` and `<path> already
exists` when the change is already there — **the ordinary case when resuming at `STARTED`** per
above. Revise the artifacts in place and never `--force` past it. That directory is the change
root — `<changeRoot>` below — by construction.

`spectre new` writes a stub `proposal.md` and an empty `tasks.md` and nothing else. Create and fill
these three artifacts:

- **proposal.md** — what and why, carrying `## Why` and `## What changes`
- **design.md** — how, from the approved design, including `## Decisions` and `## Open questions`
  (both below)
- **tasks.md** — a checkbox scaffold; **writing-plans enriches it next**

**A spec edit is planned here, never written here.** This phase runs before any branch or worktree
exists, so it has nowhere on the change branch to write to. A capability whose requirements the
change alters gets a task in `tasks.md` naming `<project>/spectre/specs/<capability>.md` in that
task's `**Files:**` field — the implementer writes and commits it on the change branch, in that
task's own commit.

**Delete the adopted staging note, if one was seeded.** If **B** found and seeded a staging note,
delete that same path (`<project>/docs/superpowers/research/<jira-key-lowercased>.md` or `<name>.md`, per which
branch of **B**'s discovery rule matched) now, alongside creating the three artifacts above, and stage
the deletion in the same commit. Skip this step outright when **B** found no note to seed from.

### Decisions

`## Decisions` in `design.md` is sourced from the brainstorming dialogue — the approach the user
chose, the alternatives on the table, and the tradeoff that ruled each one out. **A design that
forced no choices records none.**

```markdown
### <the decision>

**ID:** <kebab-case-slug>
**Status:** active
**Chosen:** <option> — <one-line rationale>
**Considered:** <other options, each with the tradeoff that ruled it out>
```

**ID** is assigned once, at creation, and is **immutable** — the match key a later round uses to
**supersede** a decision: set the old entry's `**Status:**` to `superseded by <new-id>` and append a
new entry with a fresh ID. **Never delete or rewrite a superseded entry.**

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

What this section holds is what the `STARTED` handoff counted under the old `/myflow-start` — under
`/flow`, `STARTED` is written before this section exists (per **A** above), so there is no
`STARTED` handoff to count it; the count instead appears in the `IN_PROGRESS` handoff **Verify and
hand off** (`skills/flow/verify-and-handoff.md`) prints once implementation completes.

```bash
flow stage end -command '/flow' -stage flow.create-artifacts -outcome completed <name>
```

## D. Basic Workflow #3 — Writing plans

**This section's stage marks are run by the parent**, around the planner's `## Plan` return — see
**Dispatch the planner** (`skills/flow/brainstorm.md`). Everywhere below that addresses "you" means the planner: the
writing-plans enrichment and the guards at the end of this section are its own work, and their
output is what the `## Plan` return carries.

```bash
flow stage begin -command '/flow' -stage flow.writing-plans -harness <harness> -session-token mf-<literal-token> <name>
```

Invoke **superpowers:writing-plans** to enrich `<changeRoot>/tasks.md` to plan quality: exact
paths, verification commands, bite-sized steps, no placeholders. Run its self-review (spec
coverage, placeholder scan, type consistency) before finishing.

**Tell it the task shape, because it is spectre's and not writing-plans' own.** A task is a
column-0 checkbox line, `- [ ] <n>. <title>`, whose `<n>` is a flat integer; that task's steps are
`  - [ ] **Step N: …**` lines indented two columns beneath it. The rule in full is the `Placement`
paragraph under **The build-green tag** (`skills/flow-contracts/build-green.md`).

**Load `skills/flow-contracts/plan-provenance.md`.** While enriching `tasks.md`, tag every fenced
block and every numeric claim per **Plan provenance**
(`skills/flow-contracts/plan-provenance.md`): code that cannot be verified is tagged `unverified:`
and **kept**.

**Load `skills/flow-contracts/build-green.md`.** While enriching `tasks.md`, also tag every task
with `**Build:**` per **The build-green tag**
(`skills/flow-contracts/build-green.md`), and with the mechanically-checkable field family
`flow-task-commit-fields` requires:

- `**Files:**` — the paths this task's commit will touch, with an optional
  `**Allowed-collateral:**` glob.
- `**Tests:**` — the names of the tests this task adds. A task adding none writes a field opening
  with the literal `none`, and `check-task-commit-fields.sh` then never reads that field's
  backticks as test names it must find in the commit's diff. **Bold `**none**` opens such a field;
  italic `_none_` does not** — the recognition ends on a word boundary, and `_` is a word
  character, so the trailing underscore swallows it.
- `**Regression:**` — per declared test, what fails if this task's commit is reverted.
- `**Baseline:**` — the expected test counts, as `before=<N> after=<N>`.
- `**Commit:**` — the commit subject line this task's implementer must use, scope naming the module
  the task's own `**Files:**` field carries, per **Commit scopes name the module**
  (`<agents repo>/rules/commit-scope-is-the-module.mdc`).

A task tagged `Build: red` additionally carries `**Squash-with:** Task <N>`, naming the green task
its commit folds into.

Add this header to `tasks.md`:

```markdown
> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
```

Add a second header line, in the same block, declaring whether this plan relocates existing
prose:

```markdown
> **Relocation:** yes — <one-line reason>
```

or `**Relocation:** no`. This line is required and explicit on every plan — never omitted, per
this repository's "missing rather than dropped" convention. `yes` scopes a mechanical passage
comparison (generated later in the pipeline, by a script this change adds elsewhere) to the union
of every task's own `**Files:**` field across the plan.

Before continuing, run `check-plan-shape.sh` — a shipped guard, run unconditionally — and the
project's configured plan-provenance guard and its configured build-green guard, if the project
declares them, and fix any hit.

```bash
flow stage end -command '/flow' -stage flow.writing-plans -outcome completed <name>
```

What happens once `## Plan` returns is the parent's job, not the planner's — see **Dispatch the
planner** (`skills/flow/brainstorm.md`). There is no human gate between brainstorming converging and implementation
starting.
