# Brainstorm and plan — sections B, C, D

Sections **B**, **C** and **D** of the brainstorming phase, run by whichever session is executing
`/flow`'s brainstorming stage — or, for **C**, **D** and **Decide**, a `/flow-plan` capture
(**Capturing a new change**, `skills/flow-plan/SKILL.md`) — no dispatched subagent, no relay.
Every "you" below addresses that session directly.

## B. Basic Workflow #1 — Brainstorming

The `flow.brainstorm` begin mark lives in **Run brainstorming and planning directly**
(`skills/flow/brainstorm.md`), and the `flow.brainstorm` end / `flow.design-approval` begin/end
marks under **Convergence** below stay exactly where they are, run around the merged HARD GATE
approval.

### The checklist

**A run on a filed fix/cost finding verifies the defect still exists before planning.** When the
linked issue's labels carry `flow-fix` or `flow-cost` — the earlier `myflow-` spellings matched
too, per step 9 of **Run 2 — the branch is merged**
(`skills/flow-contracts/finish-contract-run2.md`) — the checklist opens, before any design
question, with a reachability check against the resolved base (the base the `flow.kickoff`
worktree was created from): state the finding's defect as a claim the tree can answer, then run
the cheapest thing that answers it — the guard the finding names, the contract section it says is
missing, the behaviour it reports. A finding the base already delivers — the guard passes, the
line is already there — ends the run: report the evidence, the command run or the line quoted,
and stop before convergence; nothing is planned, and the issue is the operator's to close. A
finding that still reproduces plans as normal, its evidence carried into proposal.md's `## Why`
when **C** writes it.

Invoke **superpowers:brainstorming** in full: checklist items 1–8, ending with the user approving
the design.

- Save the design to `<project>/.worktrees/<name>/.superpowers/sdd/YYYY-MM-DD-<name>-design.md` — the
  worktree `flow.kickoff` created (**A. Resolve the change and write `STARTED`**, `skills/flow/brainstorm.md`). The
  path is gitignored: never stage or commit it, even where the brainstorming skill says to, and
  never write it to the main checkout.
- **HARD GATE:** do not run `spectre new` until the user approves the design. Approval is the
  merged confirm's first option under **Convergence** below; no separate approval question is
  asked.
- **A frame-specified design needs its handoff assets reachable from the tree.** When the
  design's specification is a drawn frame — a mockup the implementation must match (**Design
  mockups are a specification**, `rules/design-mockups-are-specs.mdc`) — the design is not
  approvable, and no task whose specification is that frame is written, until the handoff assets
  are committed into the repository or their location is recorded in
  `<project>/.flow/project.md` (the `mockups` row of `## visual verification` is where a declared
  mockups directory lives — **Project configuration**,
  `skills/flow-contracts/project-configuration.md`). gymie kan-29's routes A–F were built with the
  handoffs outside the tree, which is what left every "different from the mockup" report
  unanswerable and let an unchecked caption survive review.
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

**The seeded-note path is legitimate, never a bypass to prevent.** When the ask arrives already
converged — a seeded research note whose own text carries the design, its decisions with their
`**ID:**` lines, and the acceptance criteria (a fully-worked issue description, or the equivalent
a `/flow-plan` capture or a handoff package brought) — the checklist questions the note already
answers are answered by the note and never re-asked: the note is the research the checklist would
gather. A question the note leaves open is still asked, batched as above, and the merged
convergence-and-approval confirm below runs as written — approval is never seeded. gymie kan-468's run
converged its plan this way; this paragraph preserves the path.

**The note is the research, never the plan's form.** The seeded plan is still written through C
and D like any other, and what the note abbreviates the plan spells out: every `**Files:**` field
carries full repo-relative paths — the note's shorthand (commonMain/… and its like) is expanded,
never copied, because `check-task-commit-fields.sh` matches a declared path against the commit's
diff literally — and `tasks.md`'s H1 stays the exact `# <change-id>` literal `spectre validate`
requires, never a title the note supplies. gymie kan-485's run corrected every `Files:` field by hand
before task 1's guard ran clean, and gymie kan-468's fixed the seeded H1 at load-context; this
requirement is both corrections, applied where the note is consumed.

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

```bash
flow stage begin -command '/flow' -stage flow.create-artifacts -harness <harness> -session-token mf-<literal-token> <name>
spectre new "<name>"   # working directory: the worktree flow.kickoff created
```

`spectre new` scaffolds `<project>/.worktrees/<name>/spectre/changes/<name>/`, and refuses three ways: exit `2` and
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

**Every entry carries its `**ID:**` line — mandatory, never omitted.** A decision is later cited by
ID or it is re-argued from scratch, and which decisions a review round will want to cite is not
knowable at creation, so every decision the design records is written with one. **ID** is assigned
once, at creation, and is **immutable** — the match key a later round uses to **supersede** a
decision: set the old entry's `**Status:**` to `superseded by <new-id>` and append a new entry with a
fresh ID. The same key is what let gymie kan-459's review panel name the specific decision a finding found
stale instead of re-arguing the design. **Never delete or rewrite a superseded entry.**

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

`STARTED` is written before this section exists (per **A** above), so no `STARTED` handoff counts
what this section holds; the count instead appears in the `IN_PROGRESS` handoff **Verify and
hand off** (`skills/flow/verify-and-handoff.md`) prints once implementation completes.

```bash
flow stage end -command '/flow' -stage flow.create-artifacts -outcome completed <name>
```

## D. Basic Workflow #3 — Writing plans

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
> context instead of the one that just wrote the feature. The follow-on declares
> `**After:** Task <N>` naming its feature task, and the feature task declares its own
> predecessors or `none` — the declarations alone run the follow-on after the feature task — and
> the last bundle's FULL SUITE run still covers the pair.

> **Write a live-verification task when the change touches a running service or persistent
> state.** When the change under plan touches a long-running service, a daemon, a store, a
> scheduler or anything else holding runtime state, the plan carries a final task that exercises
> the real thing — the running system against its real state, not its fakes — and records the
> before/after figures it observed. That task also states what the change **not** working would
> look like, so a null result is recognisable as failure rather than read as success: a task that
> only asserts the command exited 0 cannot tell a repaired system from an untouched one. When the
> planner judges the change has no runtime to verify against, it writes one line in the plan
> saying so — the justification is required, the task is not. This is deliberately not an
> integration-test stage on every change: a step that usually resolves to "nothing to do" trains
> everyone to skip it.

> **Write a verification change so its found defects become their own tasks.** When a plan's task
> exists to exercise a surface and report what it finds — a final-verification pass over a feature
> group, an end-to-end sweep — its `**Allowed-collateral:**` names what the verification commit
> itself writes, the report or record files the pass produces, never the surface it inspects, and
> every real defect it finds becomes a new task appended to `tasks.md`, carrying the field family
> with the fix's own `**Files:**` and `**Allowed-collateral:**` (the gymie KAN-29/gymie KAN-30 precedent). The
> appended task is what makes the fix declared — `check-task-commit-fields.sh` refuses a commit
> touching paths no task declared — so the change stays self-contained and every fix stays
> traceable to a declared task.

**Load `skills/flow-contracts/plan-provenance.md`.** While enriching `tasks.md`, tag every fenced
block, every numeric claim and every assumption the plan cannot verify at plan time per **Plan
provenance** (`skills/flow-contracts/plan-provenance.md`): code that cannot be verified is tagged
`unverified:` and **kept**, and an unverifiable assumption carries `unverified:<what-to-check>` in
the task that depends on it.

**Load `skills/flow-contracts/build-green.md`.** While enriching `tasks.md`, also tag every task
with `**Build:**` per **The build-green tag**
(`skills/flow-contracts/build-green.md`), and with the mechanically-checkable field family
`flow-task-commit-fields` requires:

- `**Files:**` — the paths this task's commit will touch, with an optional
  `**Allowed-collateral:**` glob.
- `**Tests:**` — the names of the tests this task adds. **The field is parsed, not read:** a
  `Case <N>` label is checked by that label alone and backticks beside it parse as nothing;
  otherwise every backticked token becomes a declared test name the commit's diff and the tree
  are searched for; any other prose in it is invisible to the check — write what covers the task
  outside the field, never inside it. A task adding none writes a field opening with the literal
  `none`, and `check-task-commit-fields.sh` then never reads that field's backticks as test names
  it must find in the commit's diff. **Bold `**none**` opens such a field; italic `_none_` does
  not** — the recognition ends on a word boundary, and `_` is a word character, so the trailing
  underscore swallows it.
- `**Regression:**` — per declared test, what fails if this task's commit is reverted.
- `**Baseline:**` — the expected test counts, as `before=<N> after=<N>`, counted statically from
  test declarations (`@Test` and its language equivalents), never from a test run, and each
  task's delta its own: `after` is `before` plus the tests the task's own `**Tests:**` field
  names, so two tasks dispatched in parallel count independently — neither's baseline depends on
  the other's edits. The counts are exact integers: `after=~N` is unparseable by the guard and
  silently disables the re-measurement. A Baseline that records a plan-provenance `<!-- measured:
  <command> @ <ref>` `-->` comment is re-measured once the task's commit exists:
  `check-task-commit-fields.sh` re-runs that command at the commit's parent and at the commit, and
  a count differing from the declared one fails the task — record a command whose stdout is one
  integer (the `| grep -c` pipelines in kan-271 and kan-298's plans are the shape), and read
  "never from a test run" as constraining how the declared numbers are derived at plan time, not
  the guard's re-measurement.
- `**Commit:**` — the commit subject line this task's implementer must use, scope naming the module
  the task's own `**Files:**` field carries, per **Commit scopes name the module**
  (`<agents repo>/rules/commit-scope-is-the-module.mdc`).

A task tagged `Build: red` additionally carries `**Squash-with:** Task <N>`, naming the green task
its commit folds into.

An `**After:**` field — `Task <ids>` or `none` — declares the task's predecessors, and **every
task writes it**: a plan states its ordering in declarations, so `plan-dispatch-bundles.sh`
resolves every bundle's after-set from what the tasks say and never needs a human judgment call
about ordering. The dispatcher still reads an absent field as after every earlier task, so an
older plan parses unchanged — the planner never relies on that inference. Name exactly the
tasks this one's own files depend on, `none` when there are none, and write it consistently
across a `**Squash-with:**` pair (union semantics merge the pair into one bundle).

**A task cites the decision it implements, never restates it.** When a task exists to implement
an entry of design.md's `## Decisions`, it names that entry by its `**ID:**` in a
`**Decision:** <id>` field and copies nothing of the entry's text — the decision lives once,
under `## Decisions`, and the citation is the link. Restated decision prose drifts from its
entry the first time either is edited, which is why gymie kan-468's seeded plan cited instead. A task
implementing no recorded decision writes no such field.

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

With the plan validated, record its size as the change's first task-count observation — the
planned figure every later fix round's growth is measured against (KAN-415):

```bash
flow tasks count -C <worktree> <name>
```

### Decide

Run `plan-class.sh <changeRoot>/tasks.md <repos> <abs-worktree> <merge-base>` — `<repos>` is the
size of the resolved worktree set, `<abs-worktree>` the worktree `flow.kickoff` created, and
`<merge-base>` this run's working-notes merge base; the two trailing arguments are what let the
script classify `micro` (**Micro** below), and the two-argument form stays valid and never
classifies `micro`. Its three lines carry `class_mechanical` and the `compact`/`experimental` rolls.
**You may raise `class_mechanical` one step, never lower it** — `micro`→`small`, `small`→`regular`
or `regular`→`big` — with a one-line reason recorded as `override`; the raised value is `class`.
Leave `override` `null` and `class = class_mechanical` otherwise. `red` and `unverified` are
recorded from the same output and move no class.

**Micro** — when the script prints `class: micro`, decide collapses to a recorded default decision:
steps 1–4 below record the micro row's values without choosing — execution `inline`, implementer
and fixer `skipped — inline`, panel the string `default`, groups `null` — with decision.json
carrying `class: micro`.

Read `EXECUTION_MODE_TOGGLE`, `IMPLEMENTER_MODEL_TOGGLE` and `REVIEW_PANEL_TOGGLE` from this
run's own earlier resolution (**Model resolution**, `skills/flow/SKILL.md`) — already in scope,
since nothing dispatched this section. Decide, in this order, each step only when its own toggle is
`dynamic` and (for steps 2-3) step 1 held; otherwise the step takes the stated default and is
recorded as such:

1. **execution mode** — `inline` (`class` micro, small or regular) or `sdd` (`class` big); default `sdd`.
2. **implementer and fixer model + effort** — only when step 1 came out `sdd`; two pairs, each
   chosen per **Model and effort** below. The implementer pair is every implementer group's
   default (step 4). The fixer pair is the panel-fix subagent's own and is chosen from what a fix
   round does — repair named findings against their reproducers, usually a narrower job than the
   implementation — so it may differ from the implementer pair in model, effort or both, its
   `reason` saying why. Both recorded `skipped — inline` when step 1 is inline, `default` when
   the toggle is off.
3. **review panel** — roster, compact/experimental, rerun policy, the **rerun pair**, and its
   **grouping** —
   `bundle_roll < 30` the class's static row, else free within ≤2 dispatches × ≤3 roles with a
   one-line `grouping_reason`. **Model and effort are a property of each dispatch, never of a
   slot**: the roles of one bundle run in one subagent and cannot differ in model or effort. A
   static and a free grouping alike assign each dispatch its own pair per **Model and effort**
   below; `primary` + `principles`, bundled as one
   dispatch, is the universal floor for every roster this tree assigns — compact or full, every
   class — its `-slot` `primary+principles`, deterministic, no roll. Its spare seat under the
   bundle cap stays empty: nothing else ever joins the floor bundle. A
   compact roster's two roles are the floor bundle itself, nothing more to group. A full roster's
   remaining roles form the second dispatch entirely — there is no overflow case left, since the
   floor already holds every reading/judgment role; a rolled experimental slot joins the second
   dispatch only when one exists and has room, else is `skipped — bundle cap` — from **the tree**
   below, keyed on `class` and the rolls. **The rerun pair** (`panel.rerun_dispatch`) is the one pair every
   fix-round re-run dispatch runs on — one dispatch per re-running role, each targeted at the
   findings that role raised (**Panel re-runs**, `skills/flow/review-panel.md`): its `model` is any `ValidModels`
   member **no `panel.dispatches` entry uses** — a re-review by the model that raised the finding
   is not a second pair of eyes — and its `effort` is `low`, fixed, since a re-run reads a delta
   to confirm a fix and must be short and fast; its `reason` names the model choice only. Default: today's settings-store roster on
   `DEFAULT_MODEL` and `default` effort for every dispatch, delta rerun, grouped by the static
   table deterministically (no roll), recorded `default`.
4. **implementer groups** — on every run whose step 1 came out `sdd` (`## execution mode` toggle or
   not): run `plan-dispatch-bundles.sh <changeRoot>/tasks.md`, then
   `plan-dispatch-groups.sh <changeRoot>/tasks.md` for the mechanical default — a deterministic
   grouping biased toward fewer, larger groups (no roll, no static table, no per-group ceiling; at
   most two implementer dispatches in flight per wave), recorded as `groups_mechanical`. `groups`
   is `groups_mechanical` verbatim unless the planner **splits** a mechanical group into more
   groups — a chain judged too long for one implementer's context, or a real parallel-wave benefit
   the mechanical grouping's fold pass declined — with a one-line `groups_override` reason;
   **never merge across the mechanical result**, since merging two groups it kept apart would
   collapse a real parallel wave. `groups_override` is `null` when the mechanical grouping is
   taken verbatim; `groups_reason` defaults to the literal `mechanical`, or names the split's own
   reason when `groups_override` is set. **Each group carries its own `model` and `effort`**:
   step 2's implementer value by default, or — only when `IMPLEMENTER_MODEL_TOGGLE` is
   `dynamic` — a different pair the planner picks for that group alone per **Model and effort**
   (a group of mechanical, well-specified tasks on a cheaper model; a group carrying the change's
   hardest seam one step up), the reason in that group's own `reason`. On `default` every group is `DEFAULT_MODEL` /
   `default`. `groups`, `groups_mechanical` and `groups_override` are all `null` when step 1 is
   inline.

**The tree**, one row per `class`:

| class | execution | implementer/fixer | full roster | compact roster | rerun | static grouping (full roster) |
|---|---|---|---|---|---|---|
| micro | inline | — | none — defaults only | none — defaults only | — | — |
| small | inline | — | primary; principles | primary; principles | delta | `primary+principles` |
| regular | inline | — | primary; principles; mutation | primary; principles | delta | `primary+principles` · `mutation` |
| big | sdd | chosen | primary; principles; mutation; bugbot; security | primary; principles | full | `primary+principles` · `mutation+bugbot+security` |

**The micro row** records defaults, never choices: what it skips is every roster, model/effort and
grouping choice, and every roll the script printed; what it never skips is the decision record
itself, the verify stage, or the self-review.

#### Model and effort

The tree fixes no model and no effort: every pair a `dynamic` step
assigns — the implementer, the fixer, each panel dispatch, the rerun pair's model, each implementer group — is the planner's
own choice, `model` any member of the store's `ValidModels` set (`haiku`, `sonnet`, `opus`,
`fable`; `flow settings models` prints it) and `effort` one of `low`/`medium`/`high`, decided
from what that dispatch will actually do: the complexity of its tasks, the time and space
complexity of the code it writes or reviews, and the scalability the change has to hold up under.
A mechanical, well-specified dispatch sits at the cheap end; a dispatch carrying a concurrency
seam, a data-model change or a performance-sensitive path sits at the expensive end; nothing in
between is a default. Each pair carries a one-line `reason` beside it in the JSON and in the
`## Decision` block's rule cell. **On harness `zcode` the chosen pair is recorded as chosen and
replaced at dispatch** — **Harness mapping** (`skills/flow-contracts/model-policy.md`).

A compact roster is the floor bundle alone, on the floor bundle's model/effort. `bugbot` and
`security` are prompt-driven roles like every other slot, dispatched in whichever bundle carries
them on that bundle's model/effort. Compact when `compact_roll < 90`
(small) or `< 60` (regular, big);
experimental when `experimental_roll < 30` (every class, at most one slot), appended to whichever
roster and run on the model/effort of the dispatch it joins.

**Which prompt, when experimental rolled true:** `ls <agents repo>/skills/flow/experimental/*.md`,
sorted, is the ordered set of candidates; the picked file is the one at index `experimental_roll mod
count` into that sorted list. `<name>` is its basename with `.md` dropped, and its own `prompt` /
`description` fields for the roster entry are that file's path (`skills/flow/experimental/<name>.md`)
and its line-1 `description:` value. An absent directory or one holding no `*.md` file (`count = 0`)
records `experimental: none available` and adds nothing — never a division by zero. `delta`/`full`
are **Panel re-runs**' own rerun policies (`skills/flow/review-panel.md`); the docs-only reduction
there still applies and still only removes.

Write the decision JSON to `<abs-worktree>/.superpowers/sdd/decision.json` — on a first creating
run, a resumed `STARTED` run and a fix run alike, since the worktree exists from `flow.kickoff`
(**A. Resolve the change and write `STARTED`**, `skills/flow/brainstorm.md`). The JSON carries: `toggles`, `class`, `classMechanical`,
`override`, `inputs` (the four `plan-class.sh` booleans plus `tasks`/`files`/`repos`), `rolls`
(`compact`, `experimental`, `bundle`), `execution`, `implementer` and `fixer` (each an object
`{model, effort, reason}` or one of the two recorded strings above), `panel` (an object — `compact`, `rerun`, `roster:
[{slot, experimental, prompt?, description?}, …]`, `grouping` (`static`/`free`),
`dispatches` (one to two objects `{slots, model, effort, reason}`, `slots` one to three slot ids in roster
order — the one place a pass-1 reviewer's model and effort are recorded), `rerun_dispatch` (an object
`{model, effort, reason}`, the rerun pair — the one place a fix-round re-run's model and effort are
recorded),
`grouping_reason` (`null` on a static grouping) — or the string `default`; an experimental slot
skipped for the cap is recorded as the string `"experimental": "skipped — bundle cap"` beside
`roster`), `groups` (objects `{bundles, model, effort, reason}`, `bundles` an array of bundle ids, plus
sibling `groups_mechanical` (arrays of bundle ids), `groups_override` and
`groups_reason` fields, or all four `null` when `execution` is inline), `parent` (the parent's own
model/effort, `unknown` where the harness does not state one), `overrides` (session-instruction
overrides to a *result*, never a toggle; empty unless one was given). Print this exact shape as the
run's own output once the Decide step completes, filling every cell from what was just decided:

```markdown
## Decision

| Input              | Rule             | Value |
|--------------------|------------------|-------|
| class              | mechanical <class_mechanical> | <class> (override: <reason or "none">) |
| inputs             | plan-class.sh    | tasks <N> · files <N> · repos <N> · migration <yes\|no> · spec <yes\|no> · red <yes\|no> · unverified <yes\|no> |
| roll: compact      | <N> <"<" or "≥"> <threshold> | <compact\|full> |
| roll: experimental | <N> <"<" or "≥"> 30 | <slot name\|"no slot"\|"none available"> |
| roll: bundle       | <N> <"<" or "≥"> 30 | <static\|free> grouping |

| Setting            | Toggle           | Result |
|--------------------|------------------|--------|
| execution mode     | <default\|dynamic> | <inline\|sdd> |
| implementer model  | <default\|dynamic> — <reason> | <"skipped — inline"\|"default"\|model/effort> |
| ↳ fixer            | <model> / <effort> — <reason> | <"skipped — inline"\|"default"\|model/effort> |
| review panel       | <default\|dynamic> | <"default"\|<compact\|full> · <delta\|full> rerun> |
| ↳ dispatch <n>     | <model> / <effort> — <reason> | <roles `+`-joined in roster order> |
| ↳ rerun            | <model> / low — <reason> | every fix-round re-run, one role per dispatch |
| ↳ grouping         | free             | <grouping_reason> |
| implementer groups | —                | <"skipped — inline"\|"mechanical"\|<groups_reason>> |
| ↳ group <bundle ids> | <model> / <effort> — <reason> | <bundle ids> (mechanical: <groups_mechanical>; override: <groups_override>) |
```

One fact per row, every reason in the middle column, nothing printed outside the two tables. The
first table is the input side — `class`, the four `plan-class.sh` booleans with `tasks`/`files`/
`repos`, and the three rolls, each roll's rule cell the roll against its threshold and its value
cell the interpretation. On a micro run the three roll rows' value cells read `not consulted —
micro`, and the decision table records the micro row's values, so no `↳` row appears anywhere
in it. The second is the decision side. `↳` rows are sub-rows of the setting
above them: a `↳ fixer` row under the implementer row, its cells the `fixer` value in the
implementer row's own shape; one `↳ dispatch <n>` row per object in `panel.dispatches`, in order, its rule cell that
dispatch's model and effort with its `reason` and its value cell the roles `+`-joined in roster order; a `↳ rerun`
row after the last dispatch row, its rule cell the rerun pair with its `reason`; a `↳ grouping`
row only on a free grouping (omitted on a static one); an experimental slot skipped for the cap
adds `· experimental: skipped — bundle cap` to the review-panel value cell. The implementer-groups
row is `skipped — inline` on an inline run, else `groups_reason`, followed by one `↳ group` row per
object in `groups`, its rule cell the group's model and effort with its `reason` and its value cell the bundle ids
(`plan-dispatch-bundles.sh`'s ids), the `(mechanical: …; override: …)` suffix only on a split
(`groups_override` non-`null`) — mirroring the `class` row's own `override` shape. When `panel` is
the string `default` the review-panel value cell is `default` and no `↳` row follows it; when
`implementer` is a string the `↳ fixer` row carries the same string and no pair.

Prepend these three lines directly above the `## Decision` block — the one place these choices
appear in a run, never printed twice:

```text
planning:  inline, this session (<DEFAULT_MODEL>)
toggles:   execution mode <default|dynamic> · implementer model <default|dynamic> · review panel <default|dynamic>
models:    default <DEFAULT_MODEL> · reviewers <REVIEWERS>
```

```bash
flow stage end -command '/flow' -stage flow.writing-plans -outcome completed <name>
```

### Plan review gate

**The plan is never acted on unread.** Once the Decide step has printed its `## Decision` block,
print the plan summary below, then ask the gate question in one **AskUserQuestion** call. The
gate is mandatory: no answer, no commit and no implementation. It runs at the end of every
`/flow-plan` capture (**Capturing a new change**, `skills/flow-plan/SKILL.md`) and in `/flow`
after the `flow.decide` record sequence (**Run brainstorming and planning directly**,
`skills/flow/brainstorm.md`).

The summary is plain prose, not a listing of `tasks.md`: what will be implemented and how — the
behaviour being added or changed, the approach taken, the seams and decisions that matter (data
shapes, boundaries, ordering constraints, what was deliberately left out) — in enough detail that
the operator can judge the logic without opening the plan. No per-task rows, no file, test or
commit fields. Then the `## Decision` block, verbatim, under it.

Then the question — `/flow-plan`'s wording is **Push artifacts?**, `/flow`'s is **Proceed to
implementation?** — with exactly two options: **Yes** and **No (plan needs updates)**. On **No**,
take the operator's changes (a follow-up **AskUserQuestion** round when the option carried none),
revise `tasks.md`, re-run `check-plan-shape.sh` and the project's configured guards, re-run the
Decide step from `plan-class.sh` on — in `/flow`, the `flow.decide` record sequence again, a second
decision row — then print the summary and ask again. Loop until **Yes**. **Yes** is the only exit.

What happens once this section's plan enrichment completes is stated in **Run brainstorming and
planning directly** (`skills/flow/brainstorm.md`) — continuing directly into
`skills/flow/implement.md`. There is no human gate between brainstorming converging and
implementation starting.
