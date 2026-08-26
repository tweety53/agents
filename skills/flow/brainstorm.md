# Brainstorm and plan

Superpowers Basic Workflow steps **#1** (brainstorming) and **#3** (writing-plans), intertwined
with spectre artifact creation, run here — the same content `/myflow-start` carried, minus the
options-question round and the proposal publish, both removed per design.md (`ask-options-removed`,
`publish-proposal-removed`). Loaded by `skills/flow/SKILL.md` on a creating run (no state) or a
resuming run (`STARTED`).

## A. Resolve the change and write `STARTED`

**Resolve the linked Jira issue first** — it decides the change name. Follow **Resolution (how
`jiraIssue` is decided)** in `skills/flow-contracts/jira-integration.md` exactly. This is the
only phase that resolves a key.

Then the change name:

- **With a linked issue**, the name is `<lowercased-key>-<slug>`, per **Change naming**
  (`skills/flow-contracts/jira-integration.md`). Derive the slug from the issue summary when only
  a key was given.
- **Without one**, the name is the descriptive slug alone.
- If a name or description was given, use it (derive kebab-case from the description if only a
  description was given).
- **If both are omitted:** enumerate the candidate set exactly as **Change name resolution**
  (`skills/flow-contracts/pipeline.md`) defines it, restricted to changes with incomplete planning
  artifacts. Exactly one match → resume it, announcing which; multiple → **AskUserQuestion** listing
  each (name, state, last modified); zero → ask what to build.

**Transition the issue to In Progress now**, per **Transitions** in Jira integration
(`skills/flow-contracts/jira-integration.md`) — before brainstorming, so the board is correct
while planning runs. A failure is one skipped-with-reason line and planning continues; nothing about
this call may delay or alter the run.

**Mark `flow.kickoff` now that the name is fixed, and write `STARTED` immediately — before
brainstorming begins.** This is design.md's `started-redefined`: `STARTED` is a kickoff marker,
"the operator started this," not (as under the old `/myflow-start`) a record that a design was
approved and a proposal published. Write it here, at the top of this phase, rather than at the
bottom of it.

```bash
myflow stage begin -command '/flow' -stage flow.kickoff -harness <harness> -session-token mf-<literal-token> <name>
```

```json
{
  "state": "STARTED",
  "branch": null,
  "worktrees": {},
  "artifactUrl": null,
  "jiraIssue": "<resolved key, or null>",
  "planningEffort": null,
  "models": { "default": null },
  "prUrl": null,
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/flow"
}
```

`planningEffort` and `models.default` are written `null` and stay `null` for the life of the
change: `/flow` asks no planning-effort or model question on a creating run
(`ask-options-removed`), and models are resolved per run from the settings store
(`model-default-sonnet`, `settings-scope`), not recorded per change. `artifactUrl` stays `null` —
`/flow` publishes no proposal artifact (`publish-proposal-removed`).

```bash
myflow stage end -command '/flow' -stage flow.kickoff -outcome completed <name>
myflow stage begin -command '/flow' -stage flow.resolve-change -harness <harness> -session-token mf-<literal-token> <name>
myflow stage end   -command '/flow' -stage flow.resolve-change -outcome completed <name>
```

**No further command runs before this point on a creating run** — the state write above is the
first thing this invocation does once the name is fixed, ahead of even the design conversation. The
operator sees `STARTED` recorded the moment they invoke `/flow`, whether or not the run goes on to
finish brainstorming in the same sitting.

### Resuming at `STARTED`

A run finding `"state": "STARTED"` already recorded is resuming a creating run that stopped before
reaching `IN_PROGRESS` — an interrupted session, a context limit, an earlier stop. Skip **A** above
(the name and the `STARTED` write both already exist) and determine where the run actually left off
by reading, not by assuming:

- `spectre list --json`'s entry for this change's `done`/`total` — `total == 0` means no plan exists
  yet: resume at **B** below.
- The change root's own `tasks.md` — a scaffold with no enriched steps means writing-plans has not
  run: resume at **D** below; a plan meeting writing-plans quality (exact paths, verification
  commands, no placeholders) means planning is done: skip straight to
  `skills/flow/implement.md`.
- The state file's `worktrees` map — non-empty means a worktree already exists: resume implementation
  directly rather than re-running **A**–**D**.

This is a pragmatic re-entrancy rule, not an exhaustively-enumerated state machine — a run resuming
at `STARTED` reads what actually exists and continues from there, the same principle every other
`/flow` re-entry point already applies. State the resumption point plainly before continuing:
"resuming `<name>` at `<point>`."

## B. Basic Workflow #1 — Brainstorming

### Seed from a staged research note, if one exists

**Before starting the interactive checklist**, look in `<project>/docs/superpowers/research/` for a note using
this exact-filename rule, in order — **no fuzzy or substring matching beyond the wildcard fallback
named below**. This implements design.md's `flow-research-staging`, the *discovery* half of open
question `research-staging-mechanism` (the *write* half is `skills/flow-research/SKILL.md`'s own
job):

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
what it does not, confirms what it does, and gives the operator a chance to correct it. This is
`flow-research-staging`'s explicit choice: seed, never skip.

**Once the note's content is adopted into this change's own artifacts** (its design content folded
into `design.md`, its decisions and open questions carried in per **C** below), **delete the staging
note** rather than leaving it in place. This resolves the remaining half of `research-staging-mechanism`
left open by design.md: a staging note that outlives its adoption is a second, driftable copy of
what the change's own `design.md` now states canonically, and `<project>/docs/superpowers/research/` is meant
to hold notes still waiting for a home, not a permanent archive of every note that found one. Delete
it as part of **C**'s artifact-creation commit — it is a planning path, staged and committed the same
way the rest of `<project>/spectre/changes/` is: **C**'s "Delete the adopted staging note" step
removes it. If the note only partially seeded this round (a note covering one of several threads the
round expanded on), still delete it once adopted: its useful content now lives in `design.md`, which
is the canonical location from that point on.

**Carry the seeded note's path forward to C.** If a note was found and seeded above, remember its
path (e.g. `<project>/docs/superpowers/research/kan-326.md`) — **C** deletes exactly that file and only when
this note-found condition holds.

### The checklist

```bash
myflow stage begin -command '/flow' -stage flow.brainstorm -harness <harness> -session-token mf-<literal-token> <name>
```

Invoke **superpowers:brainstorming** in full: checklist items 1–8, ending with the user approving
the design.

- Save the design to `<project>/docs/superpowers/specs/YYYY-MM-DD-<name>-design.md` and stage it
  when the brainstorming skill requires it.
- **HARD GATE:** do not run `spectre new` until the user approves the design.
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
> - **Nothing unclear — move on** *(recommended)*
> - **Another round — I have something** *(default — anything short of an explicit "move on" is
>   treated as this)*

End the stage only on an explicit choice of **move on**. An answer that names something opens
another round. **This is the one prompt in this file where the safe default and the recommended
option differ, and deliberately so** — shape per Operator prompts
(`skills/flow-contracts/operator-prompts.md`): silence or a stalled prompt from an operator who is
present is not "move on," and defaults to another round rather than to the recommended choice. Print
`⚠ another round — no explicit answer` when this default fires.

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

The convergence loop's own exit — an explicit **move on** at the confirm above — is what closes the
checklist itself; the **design approval** the HARD GATE requires is a separate, later act by the
same operator and gets its own mark:

```bash
myflow stage end   -command '/flow' -stage flow.brainstorm -outcome completed <name>
myflow stage begin -command '/flow' -stage flow.design-approval -harness <harness> -session-token mf-<literal-token> <name>
# … the operator approves the design — this is the HARD GATE above …
myflow stage end   -command '/flow' -stage flow.design-approval -outcome completed <name>
```

## C. Create the change and its artifacts

```bash
myflow stage begin -command '/flow' -stage flow.create-artifacts -harness <harness> -session-token mf-<literal-token> <name>
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
myflow stage end -command '/flow' -stage flow.create-artifacts -outcome completed <name>
```

## D. Basic Workflow #3 — Writing plans

```bash
myflow stage begin -command '/flow' -stage flow.writing-plans -harness <harness> -session-token mf-<literal-token> <name>
```

Invoke **superpowers:writing-plans** to enrich `<changeRoot>/tasks.md` to plan quality: exact
paths, verification commands, bite-sized steps, no placeholders. Run its self-review (spec
coverage, placeholder scan, type consistency) before finishing.

**Tell it the task shape, because it is spectre's and not writing-plans' own.** A task is a
column-0 checkbox line, `- [ ] <n>. <title>`, whose `<n>` is a flat integer; that task's steps are
`  - [ ] **Step N: …**` lines indented two columns beneath it. The rule in full is the `Placement`
paragraph under **The build-green tag** (`skills/flow-contracts/build-green.md`).

While enriching `tasks.md`, tag every fenced block and every numeric claim per **Plan provenance**
(`skills/flow-contracts/plan-provenance.md`): code that cannot be verified is tagged `unverified:`
and **kept**.

While enriching `tasks.md`, also tag every task with `**Build:**` per **The build-green tag**
(`skills/flow-contracts/build-green.md`), and with the mechanically-checkable field family
`myflow-task-commit-fields` requires:

- `**Files:**` — the paths this task's commit will touch, with an optional
  `**Allowed-collateral:**` glob.
- `**Tests:**` — the names of the tests this task adds.
- `**Regression:**` — per declared test, what fails if this task's commit is reverted.
- `**Baseline:**` — the expected test counts, as `before=<N> after=<N>`.
- `**Commit:**` — the commit subject line this task's implementer must use, scope naming the module
  the task's own `**Files:**` field carries, per **Commit scopes name the module**
  (`<agents repo>/rules/commit-scope-is-the-module.mdc`).

A task tagged `Build: red` additionally carries `**Squash-with:** Task <N>`, naming the green task
its commit folds into.

Add this header to `tasks.md`:

```markdown
> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
```

Before continuing, run the project's configured plan-provenance guard and its configured
build-green guard, if the project declares them, and fix any hit.

```bash
myflow stage end -command '/flow' -stage flow.writing-plans -outcome completed <name>
```

Once this phase completes and the change's artifacts exist, continue — within the same invocation
and without a further command from the operator — into `skills/flow/implement.md`, exactly as
`/flow` runs implementation from `STARTED`. There is no human gate between brainstorming converging
and implementation starting.
