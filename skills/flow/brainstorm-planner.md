# Brainstorm and plan — sections B, C, D

Sections **B**, **C** and **D** of the brainstorming phase, run by whichever session is executing
`/flow`'s brainstorming stage — or, for **C**, **D** and **Decide**, a `/flow-plan` capture
(**Capturing a new change**, `skills/flow-plan/SKILL.md`) — no dispatched subagent, no relay.
Every "you" below addresses that session directly.

## B. Basic Workflow #1 — Brainstorming

### The checklist

**Load `skills/flow/withdrawal.md`** only when the linked issue's labels carry `flow-fix`,
`flow-cost` or `flow-speed` — its reachability check opens the checklist, before any design question.

Invoke **superpowers:brainstorming** in full: checklist items 1–8, ending with the user approving
the design.

- **A ticket that names a practice or a brief resolves it through the lessons home** (**Process
  lessons**, `skills/flow-contracts/lessons.md`) — never by searching repositories or archived
  narratives for it.

- Save the design to `<project>/.worktrees/<name>/.superpowers/sdd/YYYY-MM-DD-<name>-design.md` — the
  worktree `flow.kickoff` created (**A. Resolve the change and write `STARTED`**, `skills/flow/brainstorm.md`). The
  path is gitignored: never stage or commit it, even where the brainstorming skill says to, and
  never write it to the main checkout.
- **HARD GATE:** do not run `spectre new` until the user approves the design. Approval is the
  merged confirm's first option under **Convergence** below; no separate approval question is
  asked.
- **A frame-specified design needs its handoff assets reachable from the tree.** When the
  design's specification is a drawn frame — a mockup the implementation must match (**Design mockups are a specification, not an inspiration**, `rules/design-mockups-are-specs.mdc`) — the design is not
  approvable, and no task whose specification is that frame is written, until the handoff assets
  are committed into the repository or their location is recorded in
  `<project>/.flow/project.md` (the `mockups` row of `## visual verification` is where a declared
  mockups directory lives — **Project configuration — visual verification**,
  `skills/flow-contracts/project-configuration-visual.md`).
- The design presentation does **not** end a section, or the whole design, with a "does this look
  right?" question — present the section(s) and proceed directly, section to section and then into
  artifact creation, unless the operator raises an objection during or after that presentation. This
  is a scoped override of `superpowers:brainstorming`'s hard design-approval gate, `/flow` only.
- Ask every pending question whose wording does not depend on another pending answer in one
  **AskUserQuestion** call, up to four questions per call; a question that only makes sense once
  another is answered waits for the next call; the convergence confirm and the third-round offer
  may be the last question in such a call. This is a scoped override of
  `superpowers:brainstorming`'s "Only one question per message", `/flow` only.

**Load `skills/flow/seeded-note.md`** only when the ask arrives as a seeded research note — a
fully-worked issue description, a `/flow-plan` capture or a handoff package that already carries
the design, its decisions and the acceptance criteria.

The approved design is the source for the change's `design.md`; adapt its format, never duplicate a
conflicting design.

A Done-when names only committed, tracked paths of the repository it belongs to — never a
gitignored or otherwise uncommitted side output, whose removal a later cleanup turns into a done
criterion nothing satisfies — and the archive commit refuses one that does, enforced by **Archive on
the change branch** (`skills/flow-contracts/finish-contract-run1.md`).

### Stage exit — never the command's own judgment

Within a single run, a stage that loops — most concretely `/flow`'s brainstorm
stage, whose convergence test reopens after every planning-stage exchange that leaves a question the
command's inputs do not answer — never closes on the command's own judgment. It closes only on an
explicit operator answer: at a confirm, or by declining an offer, recording what is still open
rather than assuming it away. The one bounded exception is a session that cannot ask at all: it
records the confirm itself as an open question and ends the stage there, since no operator answer
could ever arrive through it. An operator who is present but silent is not that exception and still
gets another round. The same explicit answer may both close the checklist and grant the design
approval, as **Convergence** (`skills/flow/brainstorm-planner.md`) defines.

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
names something opens another round. **Here the safe default and the recommended option differ** — shape per Operator prompts
(`skills/flow-contracts/operator-prompts.md`): silence or a stalled prompt from an operator who is
present is not "approve the design and move on," and defaults to another round rather than to the
recommended choice. Print `⚠ another round — no explicit answer` when this default fires.

Under the `## decisions: recommended` mode (**Auto-resolution**,
`skills/flow-contracts/operator-prompts.md`) the confirm is not asked: the recommended
**approve the design and move on** is taken, and the silence default above never fires.

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

Under the mode the offer is not asked either: **Yes — run another round** is taken, and the
silence default above never fires (**Auto-resolution**,
`skills/flow-contracts/operator-prompts.md`).

Rounds one and two open without asking. **There is no hard cap.** No round count ends the stage —
see **Stage exit — never the command's own judgment** (`skills/flow/brainstorm-planner.md`).

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

`STARTED` is written before this section exists (section A of `skills/flow/brainstorm.md`), but the
`STARTED` handoff prints at the end of the run, so its `Open questions` line shows what this section holds
(**The `STARTED` handoff block**, `skills/flow/brainstorm.md`); the `IN_PROGRESS` handoff carries no count.

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
> module or repository suite: that run belongs to the parent's full-suite run after the last
> group (`skills/flow/implement.md`) and to `flow.verify`. A task whose `**Tests:**` is `none` names
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
> the parent's full-suite run after the last group still covers the pair.

> **Write a capture for every frame on the change's frame list.** When the design is specified by
> mockup frames and the project declares `mockups`, every frame id `design.md` cites is named in
> some task's `**Files:**` or steps, as the capture spec screenshot of the view that frame draws
> plus its `<spec>.mockups` sidecar line (**Visual verification**,
> `skills/flow-contracts/project-configuration-visual.md`). Before finishing, list the frame ids
> and name the task covering each; a frame no task covers is a plan gap, closed by adding the
> capture to the task that builds that view or by a capture task of its own.

> **Write a live-verification task when the change touches a running service or persistent
> state.** When the change under plan touches a long-running service, a daemon, a store, a
> scheduler or anything else holding runtime state, the plan carries a final task that exercises
> the real thing — the running system against its real state, not its fakes — and records the
> before/after figures it observed. That task also states what the change **not** working would
> look like, so a null result is recognisable as failure rather than read as success: a task that
> only asserts the command exited 0 cannot tell a repaired system from an untouched one. When the
> planner judges the change has no runtime to verify against, it writes one line in the plan
> saying so — the justification is required, the task is not.

> **Write a verification change so its found defects become their own tasks.** When a plan's task
> exists to exercise a surface and report what it finds — a final-verification pass over a feature
> group, an end-to-end sweep — its `**Allowed-collateral:**` names what the verification commit
> itself writes, the report or record files the pass produces, never the surface it inspects, and
> every real defect it finds becomes a new task appended to `tasks.md`, carrying the field family
> with the fix's own `**Files:**` and `**Allowed-collateral:**`. The
> appended task is what makes the fix declared — `check-task-commit-fields.sh` refuses a commit
> touching paths no task declared — so the change stays self-contained and every fix stays
> traceable to a declared task.

> **Triage the sweep's findings before any fix runs** — a defect a verification change's sweep finds is triaged before anything is edited: on the surface this change verifies, the fix is its own appended task, never an edit inside the verification task itself; anywhere else — pre-existing, or a surface another change owns — this change never fixes it, and the finding takes the sweep's recorded course, recorded and never repaired, per **The sweep** (`skills/flow-contracts/known-bugs.md`).

> **Open a task that verifies a reported behaviour with the reproduce-first step.** When a task's
> job is confirming or refuting a report — an operator's defect report, a review or self-review
> finding, a claimed behaviour — its first step reproduces the report against the untouched tree,
> before any edit. Write the step into the task's own step text — "reproduce before touching
> anything; do not guess" — never as intent the task leaves implicit: the implementer reads its
> task, and a report reproduced before anything is touched is what closes a non-defect with no
> code change, while an edit made before reproducing destroys the evidence the task set out to
> gather.

**Load `skills/flow-contracts/plan-provenance.md`.** While enriching `tasks.md`, tag every fenced
block, every numeric claim and every assumption the plan cannot verify at plan time per **Plan
provenance** (`skills/flow-contracts/plan-provenance.md`).

**A task premise about a guard's behaviour is run at plan time, never assumed.** When a task
written here rests on what a guard or check currently does — that a ratchet trips or does not,
an exit code, the figures a guard prints — planning runs that guard before the task is written
and cites its output in the task as a `measured:` comment per **Plan provenance**
(`skills/flow-contracts/plan-provenance.md`), naming the command and the ref and quoting the
exit code or figures the run produced. A premise the run disproves strikes the task during
planning, before the panel reads the plan.

**A task premise about a harness or tool is spiked live at plan time, never assumed.** When a
task written here rests on how a harness or tool behaves — that a config key takes effect where
the task puts it, how a clock or timer behaves, what a dev-stack rebuild picks up — planning runs
a live spike of five minutes or less for each such assumption before the task is written, and
records the result in the task as a `measured:` comment per **Plan provenance**
(`skills/flow-contracts/plan-provenance.md`), naming the command and the ref. A premise the spike
disproves is rewritten or struck during planning; one the spike cannot settle in five minutes
stays tagged `unverified:<what-to-check>` for the implementer.

**Load `skills/flow-contracts/build-green.md`.** While enriching `tasks.md`, also tag every task
with `**Build:**` per **The build-green tag**
(`skills/flow-contracts/build-green.md`), and with the mechanically-checkable field family
`flow-task-commit-fields` requires:

- `**Files:**` — the paths this task's commit will touch, with an optional
  `**Allowed-collateral:**` glob. Every path is written literal and
  repo-relative: a shorthand the plan's preamble defines — gs, commonMain
  and their like — is expanded by the writer, never left in the field for
  the guard to expand.
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
  names. The counts are exact integers: `after=~N` is unparseable by the guard and
  silently disables the re-measurement. A Baseline that records a plan-provenance `<!-- measured:
  <command> @ <ref>` `-->` comment is re-measured once the task's commit exists:
  `check-task-commit-fields.sh` re-runs that command at the commit's parent and at the commit, and
  a count differing from the declared one fails the task — record a command whose stdout is one
  integer (a `| grep -c` pipeline is the shape), and read
  "never from a test run" as constraining how the declared numbers are derived at plan time, not
  the guard's re-measurement. The recorded command names no ref of its own — no tree, sha, branch
  or revision argument anywhere inside it: the re-measurement checks each of its two points out
  itself and runs the command in that working tree. The
  comment's `@ <ref>` is the provenance annotation, never part of the command.
- `**Commit:**` — the commit subject line this task's implementer must use, scope naming the module
  the task's own `**Files:**` field carries, per **Commit scopes name the module**
  (`<agents repo>/rules/commit-scope-is-the-module.mdc`).

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
under `## Decisions`, and the citation is the link. A task
implementing no recorded decision writes no such field. A block of `**Decision:**` lines sits
between blank lines — one before the first line of the block and one after the last, never
glued to the `**Commit:**` subject above it.

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

or `**Relocation:** no`. This line is required and explicit on every plan — never omitted. `yes` scopes a mechanical passage
comparison (generated later in the pipeline, by `generate-relocation-comparison.sh`,
run before the review panel's `final-review.diff`) to the union
of every task's own `**Files:**` field across the plan.

Before continuing, run `check-plan-shape.sh` — a shipped guard, run unconditionally — and the
project's configured plan-provenance guard and its configured build-green guard, if the project
declares them, and fix any hit. **A hit is exit 1, and only exit 1.** The refusal exits are
never repaired by editing the plan: check-plan-provenance's exit 2 (environment) and
exit 3 (containment — a security finding first, per the guard's own contract) and the
not-a-verdict exit 2 the shape and build-green guards share (a missing or unreadable plan file)
are reported and stop the plan. Provenance's exit 4 (content-classification) is the one refusal
whose remedy is editing the plan — fix the line it names and run the guard again, its own
prescribed course, never accepting the plan while it stands.

With the plan validated, record its size as the change's first task-count observation — the
planned figure every later fix round's growth is measured against:

```bash
flow tasks count -C <worktree> <name>
```

### Decide

Run `plan-class.sh <changeRoot>/tasks.md <repos> <abs-worktree> <merge-base>` — `<repos>` is the
number of distinct repository roots the plan's `**Files:**` fall under (the project's `## apps`
table; `1` when every path is in this one), `<abs-worktree>` the worktree `flow.kickoff` created, and
`<merge-base>` this run's working-notes merge base; the two trailing arguments are what let the
script classify `micro` (**Micro** below), and the two-argument form stays valid and never
classifies `micro`. Its `inputs:`, `class:` and `rolls:` lines carry `class_mechanical` and the
four rolls; its `tree:`, `panel:`, `grouping:` and `experimental:` lines are **the tree** for that
class — execution, implementer, roster, rerun policy, grouping, dispatches and the experimental
slot, printed rather than derived by hand (the rules are the script's header).
**You may raise `class_mechanical` one step, never lower it** — `micro`→`small`, `small`→`regular`
or `regular`→`big` — with a one-line reason recorded as `override`; the raised value is `class`.
On a raise, re-run the script with `-class <class>` first, so its tree lines are the raised
class's; it exits 2 on any other step.
Leave `override` `null` and `class = class_mechanical` otherwise. `red` and `unverified` are
recorded from the same output and move no class.

**Micro** — when the script prints `class: micro`, decide collapses to a recorded default decision:
steps 1–4 below record the micro row's values without choosing — execution `inline`, implementer
and fixer `skipped — inline`, panel the string `default`, groups `null` — with decision.json
carrying `class: micro`.

Decide, in this order — step 2 only when step 1 came out `sdd`:

1. **execution mode** — `inline` (`class` micro, small or regular) or `sdd` (`class` big).
2. **implementer and fixer model + effort** — only when step 1 came out `sdd`; two pairs, each
   chosen per **Model and effort** below. The implementer pair is every implementer group's
   default (step 4). The fixer pair is the panel-fix subagent's own and is chosen from what a fix
   round does — repair named findings against their reproducers, usually a narrower job than the
   implementation — so it may differ from the implementer pair, its `reason` saying why. Both
   recorded `skipped — inline` when step 1 is inline, where the parent applies panel fixes on its
   own model (`skills/flow/implement.md`).
3. **review panel** — roster, compact/experimental, rerun policy, the **rerun pair**, and its
   **grouping** — as the script's `panel:`, `grouping:` and `experimental:` lines print them, a
   free grouping with a one-line `grouping_reason`. **Model and effort are a property of each dispatch, never of a
   slot**: the roles of one bundle run in one subagent and cannot differ in model or effort. A
   static and a free grouping alike assign each dispatch its own pair per **Model and effort**
   below. **The rerun pair** (`panel.rerun_dispatch`) is the one pair every
   fix-round re-run dispatch runs on — one dispatch per re-running role, each targeted at the
   findings that role raised (**Panel re-runs**, `skills/flow/review-panel-fix-round.md`): its `model` is
   chosen per **Model and effort** below, and its `effort` is `low`, fixed, since a re-run reads a
   delta to confirm a fix and must be short and fast.
4. **implementer groups** — on every run whose step 1 came out `sdd`: run `plan-dispatch-bundles.sh <changeRoot>/tasks.md`, then
   `plan-dispatch-groups.sh <changeRoot>/tasks.md` for the mechanical default — a deterministic
   grouping biased toward fewer, larger groups (no roll, no static table, no per-group ceiling; at
   most three implementer dispatches in flight per wave), recorded as `groups_mechanical`. `groups`
   is `groups_mechanical` verbatim unless the planner **splits** a mechanical group into more
   groups — a chain judged too long for one implementer's context, or a real parallel-wave benefit
   the mechanical grouping's fold pass declined — with a one-line `groups_override` reason;
   **never merge across the mechanical result**, since merging two groups it kept apart would
   collapse a real parallel wave. `groups_override` is `null` when the mechanical grouping is
   taken verbatim; `groups_reason` defaults to the literal `mechanical`, or names the split's own
   reason when `groups_override` is set. **Each group carries its own `model` and `effort`**:
   step 2's implementer value by default, or a different pair the planner picks for that group
   alone per **Model and effort** (a group of mechanical, well-specified tasks at a lower effort; a
   group carrying the change's hardest seam one step up), the reason in that group's own `reason`.
   `groups`, `groups_mechanical` and `groups_override` are all `null` when step 1 is
   inline.
5. **visual verification** — on every class, `micro` included: it judges what the change can
   alter, not what its run costs. `visual` is `{verify, reason}`. `not configured` when
   `<project>/.flow/project.md` declares no `## visual verification` section, recorded without
   choosing. Otherwise judge from the plan's `**Files:**` whether anything a user sees can change —
   a page, a response a page consumes, the CORS, auth or routing the frontend depends on:
   `required` when it can, `skipped` when it cannot, the `reason` naming why. A path inside a
   declared `ui paths` glob is not by itself a reason for `required`, since a glob is coarse;
   `required` answers genuine doubt, never an obvious no-UI change. Worked example, gymie kan-863:
   `**Files:**` a gateway `MetricsScrapeTokenFilter` matching only `/actuator/prometheus|metrics`,
   management config, the Caddyfile, deploy workflows and tests — all inside the declared
   `src/gateway/src/main/**` glob, yet nothing a user sees can change → `skipped`, reason
   "actuator-only filter and deploy config; no page, no response a page consumes, no CORS/route
   the frontend uses". Its contrast: a gateway CORS rule or route the frontend calls →
   `required`. A fix run, which passes no plan gate, may raise `skipped` to `required` and never
   lowers `required` to `skipped` — a decision row with no `visual` counts as `required`. What
   `skipped` does is **Visual verification** (`skills/flow/verify-and-handoff.md`).

**The micro row** records defaults, never choices: what it skips is every roster, model/effort and
grouping choice, and every roll the script printed; what it never skips is the decision record
itself, step 5's visual judgement, the verify stage, or the self-review.

#### Model and effort

**This section is canonical for how a dispatch's model is chosen.** Every pair a Decide step
assigns — the implementer, the fixer, each panel dispatch, the rerun pair, each implementer group
— carries a planner-chosen `model`, one of `opus` or `sonnet`, within these bounds:

- **implementer, fixer and each implementer group** — `opus` by default; `sonnet` only for simple,
  predictable work near-certain to succeed on the first try, the pair's `reason` saying so;
- **a pair carrying a hard seam** — concurrency, platform interop, a data-model change,
  performance-sensitive code — `opus`, always;
- **every pass-1 panel dispatch** — `opus`, always;
- **every gated per-task reviewer bundle** — `opus`, always, at its group's effort; it is a first-pass review too (`skills/flow/implement.md`);
- **the rerun pair** — `opus` or `sonnet`; nothing requires it to differ from a pass-1 dispatch.

A micro decision records no pair, and a dispatch with no recorded pair — a micro panel,
**review-panel.md**'s no-decision dispatch, the tooling analyst — runs on the literal `opus`. `flow record
decision` refuses a decision whose pairs name any other model. Pairs may repeat — two dispatches,
or a pass-1 dispatch and the rerun pair, on the same model and effort is not a defect. The rerun
pair's effort is fixed at `low` (step 3). **A verification-only implementer group is `opus` at `low`, fixed, whatever the
roll** — a group whose every task only adds end-to-end specs, fidelity captures or baselines
against design frames, or a live-verification record of an already-built feature, and changes no
production code. **When `effort_roll < 80`, every other `effort` is
`medium`**, save a pair carrying a hard seam, which may take `high`, its `reason` naming the seam.
When `effort_roll ≥ 80`, every other `effort` is the
planner's own choice, one of `low`/`medium`/`high`, decided from what that dispatch will actually
do: the complexity of its tasks, the time and space
complexity of the code it writes or reviews, and the scalability the change has to hold up under.
A mechanical, well-specified dispatch sits at the cheap end; a dispatch carrying a concurrency
seam, a data-model change or a performance-sensitive path sits at the expensive end; nothing in
between is a default. Each pair carries a one-line `reason` beside it in the JSON; the `## Decision`
block prints it only for a group whose pair departs from the implementer's. **On harness `zcode` the chosen pair is recorded as chosen and
replaced at dispatch** — **Harness mapping** (`skills/flow-contracts/model-policy.md`).

An experimental slot runs on the model/effort of the dispatch it joins; its roster entry's
`prompt` and `description` are the path and description the `experimental:` line prints. `delta`
is **Panel re-runs**' own rerun policy (`skills/flow/review-panel-fix-round.md`); the docs-only reduction
there still applies and still only removes.

Write the decision JSON to `<abs-worktree>/.superpowers/sdd/decision.json` — on a first creating
run, a resumed `STARTED` run and a fix run alike, since the worktree exists from `flow.kickoff`
(**A. Resolve the change and write `STARTED`**, `skills/flow/brainstorm.md`). The JSON carries: `class`, `classMechanical`,
`override`, `inputs` (the four `plan-class.sh` booleans plus `tasks`/`files`/`repos`), `rolls`
(`compact`, `experimental`, `bundle`, `effort`), `execution`, `implementer` and `fixer` (each an object
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
`groups_reason` fields, or all four `null` when `execution` is inline), `visual` (an object
`{verify, reason}`, `verify` one of `required`, `skipped` or `not configured`), `parent` (the parent's own
model/effort, `unknown` where the harness does not state one), `overrides` (session-instruction
overrides to a *result*, each replacing the pair(s) it names for this run; empty unless one was
given).

Once the Decide step completes, render the run's own output from that file and print its output
verbatim — the `## Decision` block, one short line per choice; the `STARTED` handoff prints the
same output again (**The `STARTED` handoff block**, `skills/flow/brainstorm.md`):

```bash
flow decision render -file <abs-worktree>/.superpowers/sdd/decision.json -session-model '<the model named in this session's own system prompt>' -reviewers '<REVIEWERS>'
```

It exits 0 on a printed block and 2 on a usage error, an unreadable file or a `decision.json`
missing a required field or carrying a malformed `visual` — correct the file and render again.

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
implementation?** — with exactly two options: **Yes** *(recommended)* and **No (plan needs
updates)**. On **No**,
take the operator's changes (a follow-up **AskUserQuestion** round when the option carried none),
revise `tasks.md`, re-run `check-plan-shape.sh` and the project's configured guards, re-run the
Decide step from `plan-class.sh` on — in `/flow`, the `flow.decide` record sequence again, a second
decision row — then print the summary and ask again. Loop until **Yes**. **Yes** is the only exit.
Under the `## decisions: recommended` mode the gate takes **Yes** without asking — the summary
and the `## Decision` block still print, so the record of what was approved stays complete
(**Auto-resolution**, `skills/flow-contracts/operator-prompts.md`).

`/flow-plan` commits and ends with its own report (**Capturing a new change**, `skills/flow-plan/SKILL.md`).
