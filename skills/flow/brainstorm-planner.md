# Brainstorm and plan — sections B, C, D

Sections **B**, **C** and **D** of the brainstorming phase, run by whichever session is executing
`/flow`'s brainstorming stage (or, for the seed-lookup and checklist mechanics **B** defines,
`/flow-plan` — see **One shared mechanism, not two copies** under **B** below) — no dispatched
subagent, no relay. Every "you" below addresses that session directly.

## B. Basic Workflow #1 — Brainstorming

**This section's stage marks are run inline, by this same session.** The `flow.brainstorm` begin
mark lives in **Run brainstorming and planning directly** (`skills/flow/brainstorm.md`), and the
`flow.brainstorm` end / `flow.design-approval` begin/end marks under **Convergence** below stay
exactly where they are, run around the merged HARD GATE approval. Everywhere else in this
section — and in **C** and **D** below — "you" means this same running session: the seeded-note
discovery and its deletion, the checklist, and the convergence loop are all done directly, with no
return or relay step in between.

### Seed from a staged research note, if one exists

**Before starting the interactive checklist**, check for a note using this exact-filename rule —
one `test -f`, never a glob, never inference:

- **The change carries a linked Jira issue:** check `<project>/docs/superpowers/research/<jira-key-lowercased>.md`.
  This is the exact, mandatory shape `skills/flow-plan/SKILL.md`'s "Staging a Note" writes — see
  **The strict research-artifact path** there for why it is the only destination a keyed session
  ever writes.
- **No linked issue:** check `<project>/docs/superpowers/research/<name>.md`, where `<name>` is the
  change's own resolved (slug-only) name. This is the "slug derived from the topic itself" case
  `skills/flow-plan/SKILL.md` describes — an exact match only when the change's later slug happens
  to reuse the research session's own topic wording.

The exact-filename check is unambiguous by construction: there is exactly one path to test per
case, never a wildcard and never more than one candidate to disambiguate between. A topic captured
under different wording than the change later resolves to will not be found by this rule at all —
that is an accepted limit of a deterministic, exact-filename check, not a defect to patch with
fuzzy heuristics: guessing which note "probably" matches risks seeding from the wrong topic
silently.

**If found**, parse it against **The Fixed Section Structure** (`skills/flow-plan/SKILL.md`)
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

### One shared mechanism, not two copies

`/flow-plan`'s own investigate-then-ask session (`skills/flow-plan/SKILL.md`) reads its
seed-lookup rule from this section rather than restating it in its own words — the same
exact-filename check above, keyed the same way off a linked Jira issue or a bare session's own
slug. `/flow-plan` never runs `spectre new` and creates no worktree, so it applies this section's
lookup and checklist mechanics without **C**'s artifact-creation or **D**'s writing-plans steps,
which are `/flow`'s alone. Keeping the lookup rule in one place is what keeps `/flow`'s inline
brainstorm checklist and `/flow-plan`'s inline research checklist from drifting apart as either one
changes.

### The checklist

Invoke **superpowers:brainstorming** in full: checklist items 1–8, ending with the user approving
the design.

- Save the design to `<project>/docs/superpowers/specs/YYYY-MM-DD-<name>-design.md`, in the main
  checkout — no worktree exists yet at this point (**Worktree creation moves to the end of
  planning**, `design.md`) — and stage it there when the brainstorming skill requires it; never
  commit it on the main checkout's own branch.
- **HARD GATE:** do not run `spectre new` until the user approves the design. Approval is the
  merged confirm's first option under **Convergence** below; no separate approval question is
  asked.
- For multi-subsystem work, decompose before proposing.
- The design presentation does **not** end a section, or the whole design, with a "does this look
  right?" question — present the section(s) and proceed directly, section to section and then into
  artifact creation, unless the operator raises an objection during or after that presentation. This
  is a scoped override of `superpowers:brainstorming`'s hard design-approval gate, `/flow` only.
- Ask every pending question whose wording does not depend on another pending answer in one
  **AskUserQuestion** call, up to four questions per call; a question that only makes sense once
  another is answered waits for the next call; the convergence confirm and the third-round offer
  may be the last question in such a call. This is a scoped override of
  `superpowers:brainstorming`'s "Only one question per message", `/flow` only.

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
see **Stage exit — never the command's own judgment** (`skills/flow-contracts/pipeline.md`).

The explicit **approve the design and move on** answer is at once the convergence exit that closes
the checklist and the design approval the HARD GATE requires; mark `flow.brainstorm` end, then
`flow.design-approval` begin and end, around that one answer:

```bash
flow stage end   -command '/flow' -stage flow.brainstorm -outcome completed <name>
flow stage begin -command '/flow' -stage flow.design-approval -harness <harness> -session-token mf-<literal-token> <name>
# … the operator's approve-and-move-on answer — this is the HARD GATE above …
flow stage end   -command '/flow' -stage flow.design-approval -outcome completed <name>
```

## C. Create the change and its artifacts

**This section's stage marks are run inline, by this same session** — see **Run brainstorming
and planning directly** (`skills/flow/brainstorm.md`). Everywhere below, "you" means this same
session: `spectre new`, the three artifacts, and the staging-note deletion are its own work.

```bash
flow stage begin -command '/flow' -stage flow.create-artifacts -harness <harness> -session-token mf-<literal-token> <name>
# … no worktree exists yet — this and the rest of C, plus all of D, run directly against the
# main checkout's own spectre tree (skills/flow/brainstorm.md's "Worktree creation moves to the
# end of planning") …
spectre new "<name>"
```

`spectre new` scaffolds `<project>/spectre/changes/<name>/` **in the main checkout**, and refuses three ways: exit `2` and
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

**A spec edit is planned here, never written here.** A capability whose requirements the
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

**This section's stage marks are run inline, by this same session** — see **Run brainstorming
and planning directly** (`skills/flow/brainstorm.md`). Everywhere below, "you" means this same
session: the writing-plans enrichment and the guards at the end of this section are its own work.

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

> **Write each task's verify step as its own lint commands plus targeted tests.** A verify step
> names the lint commands the task's own `**Files:**` actually need — never the project's whole
> `## lint` list — and the build tool's own selector for each `**Tests:**` entry (`--tests
> '<class>'` for Gradle, `-run '<name>'` for `go test`, `-t '<name>'` for vitest), never the bare
> module or repository suite: that run belongs to the last bundle's FULL SUITE paragraph
> (`skills/flow/implement.md`) and to `flow.verify`. A task whose `**Tests:**` is `none` names
> lint alone and no test command.

> **Write a feature's UI tests as their own follow-on task.** When a feature's tests live in a
> UI-test source set of their own — Compose Multiplatform's `desktopTest`, and the like — the plan
> carries two tasks: the feature task, whose `**Files:**` names the source files and the unit-test
> (`commonTest`) files, then one follow-on task per feature task whose `**Files:**` names only that
> feature's UI-test file(s), all of them. The two are file-disjoint, so `plan-dispatch-bundles.sh`
> keeps them separate bundles and the UI-test iteration starts in a fresh implementer at a small
> context instead of the one that just wrote the feature. No `**After:**` field is needed — the
> serial default already runs the follow-on after its feature task — and the last bundle's FULL
> SUITE run still covers the pair.

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

An optional `**After:**` field — `Task <ids>` or `none` — declares the task's predecessors, and
its absence means the task runs after every earlier task. Write it on file-disjoint tasks with no
caller/helper relationship, and write it consistently across a `**Squash-with:**` pair (union
semantics merge the pair into one bundle). A task left unannotated stays fully serial, so opting in
is per task.

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

### Decide

Run `plan-class.sh <changeRoot>/tasks.md <repos>` — `<repos>` is the size of the resolved worktree
set. Its three lines carry `class_mechanical` and the `compact`/`experimental` rolls. **You may
raise `class_mechanical` one step, never lower it** — `small`→`regular` or `regular`→`big` — with a
one-line reason recorded as `override`; the raised value is `class`. Leave `override` `null` and
`class = class_mechanical` otherwise. `red` and `unverified` are recorded from the same output and
move no class.

Read `EXECUTION_MODE_TOGGLE`, `IMPLEMENTER_MODEL_TOGGLE` and `REVIEW_PANEL_TOGGLE` from this
run's own earlier resolution (**Model resolution**, `skills/flow/SKILL.md`) — already in scope,
since nothing dispatched this section. Decide, in this order, each step only when its own toggle is
`dynamic` and (for steps 2-3) step 1 held; otherwise the step takes the stated default and is
recorded as such:

1. **execution mode** — `inline` (`class` small or regular) or `sdd` (`class` big); default `sdd`.
2. **implementer/fixer model + effort** — only when step 1 came out `sdd`; from **the tree** below,
   keyed on `class`. Recorded `skipped — inline` when step 1 is inline, `default` when the toggle
   is off.
3. **review panel** — roster, compact/experimental, rerun policy, and its **grouping** —
   `bundle_roll < 30` the class's static row, else free within ≤2 dispatches × ≤3 roles with a
   one-line `grouping_reason`; `primary` + `simple-reviewer` + `principles`, bundled as one
   dispatch, is the universal floor for every roster this tree assigns — compact or full, every
   class — its `-slot` `primary+simple-reviewer+principles`, deterministic, no roll. It already
   uses the bundle cap's full three-role capacity, so it never has spare room for anything else. A
   compact roster's three roles are the floor bundle itself, nothing more to group. A full roster's
   remaining roles form the second dispatch entirely — there is no overflow case left, since the
   floor already holds every reading/judgment role; a rolled experimental slot joins the second
   dispatch only when one exists and has room, else is `skipped — bundle cap` — from **the tree**
   below, keyed on `class` and the rolls. Default: today's settings-store roster on
   `DEFAULT_MODEL`, delta rerun, grouped by the static table deterministically (no roll), recorded
   `default`.
4. **implementer groups** — on every run whose step 1 came out `sdd` (`## execution mode` toggle or
   not): run `plan-dispatch-bundles.sh <changeRoot>/tasks.md`, group its bundles freely (no roll, no
   static table, no per-group ceiling; at most two implementer dispatches in flight per wave), and
   record `groups` (arrays of bundle ids) and a one-line `groups_reason`. `groups: null` when step 1
   is inline.

**The tree**, one row per `class`, `effort` one of `low`/`medium`/`high`:

| class | execution | implementer/fixer | full roster (slot: model/effort) | compact roster | rerun | static grouping (full roster) |
|---|---|---|---|---|---|---|
| small | inline | — | primary: sonnet/medium; simple-reviewer: haiku/medium; principles: haiku/medium | primary: sonnet/medium; simple-reviewer: haiku/medium; principles: haiku/medium | delta | `primary+simple-reviewer+principles` |
| regular | inline | — | primary: sonnet/high; simple-reviewer: haiku/medium; principles: sonnet/medium; mutation: sonnet/medium | primary: sonnet/high; simple-reviewer: haiku/medium; principles: sonnet/medium | delta | `primary+simple-reviewer+principles` · `mutation` |
| big | sdd | opus/high | primary: opus/high; simple-reviewer: sonnet/high; principles: opus/medium; mutation: sonnet/high; bugbot; security | primary: opus/high; simple-reviewer: sonnet/high; principles: opus/medium | full | `primary+simple-reviewer+principles` · `mutation+bugbot+security` |

`bugbot` and `security` are prompt-driven roles like every other slot: on big they take the class's
`simple-reviewer` model/effort, recorded like any other slot's. Compact when `compact_roll < 90`
(small) or `< 60` (regular, big);
experimental when `experimental_roll < 30` (every class, at most one slot), appended to whichever
roster on sonnet/medium (small, regular) or sonnet/high (big).

**Which prompt, when experimental rolled true:** `ls <agents repo>/skills/flow/experimental/*.md`,
sorted, is the ordered set of candidates; the picked file is the one at index `experimental_roll mod
count` into that sorted list. `<name>` is its basename with `.md` dropped, and its own `prompt` /
`description` fields for the roster entry are that file's path (`skills/flow/experimental/<name>.md`)
and its line-1 `description:` value. An absent directory or one holding no `*.md` file (`count = 0`)
records `experimental: none available` and adds nothing — never a division by zero. `delta`/`full`
are **Panel re-runs**' own rerun policies (`skills/flow/review-panel.md`); the docs-only reduction
there still applies and still only removes.

Write the decision JSON: on a first creating run, to
`<project>/spectre/changes/<name>/.superpowers-sdd-decision.json` in the main checkout, since no
worktree exists yet (**Worktree creation moves to the end of planning**, `design.md`) —
`skills/flow/brainstorm.md` moves it into `<abs-worktree>/.superpowers/sdd/decision.json` once the
worktree is created. On a resumed `STARTED` run or a fix run, the worktree already exists (**Resume
and fix runs** below), so write directly to its usual `<abs-worktree>/.superpowers/sdd/decision.json`
path. Either way the JSON carries: `toggles`, `class`, `classMechanical`,
`override`, `inputs` (the four `plan-class.sh` booleans plus `tasks`/`files`/`repos`), `rolls`
(`compact`, `experimental`, `bundle`), `execution`, `implementer` (an object or one of the two
recorded strings above), `panel` (an object — `compact`, `rerun`, `roster:
[{slot, model, effort, experimental, prompt?, description?}, …]`, `grouping` (`static`/`free`),
`dispatches` (one to two arrays of one to three slot ids each, roster order within each array),
`grouping_reason` (`null` on a static grouping) — or the string `default`; an experimental slot
skipped for the cap is recorded as the string `"experimental": "skipped — bundle cap"` beside
`roster`), `groups` (arrays of bundle ids, plus a sibling `groups_reason`, or `null` when
`execution` is inline), `parent` (the parent's own model/effort, `unknown` where the harness does
not state one), `overrides` (session-instruction overrides to a *result*, never a toggle; empty
unless one was given). Print this exact shape as the run's own output once the Decide step
completes, filling every cell from what was just decided:

```markdown
## Decision

| Setting | Toggle | Result |
|---|---|---|
| execution mode | <default\|dynamic> | <inline\|sdd> |
| implementer model | <default\|dynamic> | <"skipped — inline"\|"default"\|model/effort> |
| review panel | <default\|dynamic> | <"default"\|roster summary; rerun policy>; dispatches: <group> · <group> |
| implementer groups | — | <"skipped — inline"\|"[1,2] · [3] · [4,5]" — <reason>> |

class: <class> (mechanical: <class_mechanical>; override: <reason or "none">)
inputs: tasks=<N> files=<N> repos=<N> migration=<yes|no> spec=<yes|no> red=<yes|no> unverified=<yes|no>
rolls: compact <N> (<interpretation>) · experimental <N> (<interpretation>) · bundle <N> (<interpretation>)
```

The review-panel cell's `dispatches:` suffix is each group's roles `+`-joined in roster order; on a
free grouping, the line right after it is `grouping: free — <grouping_reason>` (omitted on a static
grouping). The implementer-groups row is `skipped — inline` on an inline run, else the groups as
bundle ids (`plan-dispatch-bundles.sh`'s ids) followed by `— <groups_reason>`.

**On a no-seed run** (`skills/flow/SKILL.md`'s startup-visibility print skipped the seeded-path <!-- refs-guard:allow -->
block, since no research seed was found at kickoff), prepend these three lines directly above the
`## Decision` table — the one place these choices appear for a no-seed run, never printed twice:

```text
planning:  inline, this session (<DEFAULT_MODEL>)
toggles:   execution mode <default|dynamic> · implementer model <default|dynamic> · review panel <default|dynamic>
models:    default <DEFAULT_MODEL> · reviewers <REVIEWERS>
```

**On a seeded run**, these three lines already printed at kickoff (**Model resolution**,
`skills/flow/SKILL.md`) — do not print them again here; only the `## Decision` table itself prints,
completing that earlier block's `decision:` line.

```bash
flow stage end -command '/flow' -stage flow.writing-plans -outcome completed <name>
```

What happens once this section's plan enrichment completes is stated in **Run brainstorming and
planning directly** (`skills/flow/brainstorm.md`) — continuing directly into
`skills/flow/implement.md`. There is no human gate between brainstorming converging and
implementation starting.
