---
name: spectre-planning
description: The planning prompts of spectre's tool + planning prompts mode — brainstorming a design in plan mode, filling proposal.md, design.md and tasks.md, and writing a plan spectre's guards accept. The tool-only mode is the bare spectre CLI and loads none of this. Use when planning a spectre change.
---

# spectre planning

The generic planning prompts for a spectre change: brainstorm the design, create the change's three artifacts, and write the plan.

### Plan mode

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
  asked. On Claude Code, `/flow` and `/flow-plan` take the approval from `ExitPlanMode` instead —
  the merged confirm still closes the checklist (decision `exit-plan-mode-is-approval`, kan-1033).

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

A Done-when names only committed, tracked paths of the repository it belongs to — never a
gitignored or otherwise uncommitted side output, whose removal a later cleanup turns into a done
criterion nothing satisfies — and the archive commit refuses one that does, enforced by **Archive on
the change branch** (`skills/flow-contracts/finish-contract-run1.md`).

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

Within a single run, a stage that loops — most concretely `/flow`'s brainstorm
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
names something opens another round. **Here the safe default and the recommended option differ** — shape per Operator prompts
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

> **A task that adds a member to an existing class or interface** counts that type's members
> against the project's size-rule thresholds — detekt's `TooManyFunctions`, and the like — and,
> where the type has no headroom left, plans the split in that same task rather than a "no new
> type" constraint the lint then breaks (kan-999).

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

> **Write end-to-end spec and fidelity-capture tasks to run side by side.** Each such task's
> `**After:**` names only the feature tasks whose behaviour it exercises — never another
> verification task, since one spec file never needs another's commit — and each seeds its own
> accounts and data, so the tasks share one running stack (**One stack for the verification
> tasks**, `skills/flow/implement.md`) without touching each other's state. Global Constraints
> names the stack's start row and any build the tasks run against; no task's steps start, stop
> or rebuild the stack.

> **Write every click and capture in an end-to-end or fidelity spec on a structural locator.**
> Each such task's steps reach every control the spec clicks, types into or captures by its role
> and accessible name, its label, its text or its test id — never by a page coordinate, which
> names a position in one layout and lands on another control once any task shifts that layout.

> **Write a spec so it never waits out a production timing.** A scenario that depends on a
> production timeout, retry backoff or polling interval plans a test-only override that
> shortens it — the production value unchanged — in the task that adds the spec, and a suite
> whose specs are independent runs on the test runner's parallel workers.

> **Write a live check when the change touches a running service or persistent state — never a
> live-verification task.** When the change under plan touches a long-running service, a daemon,
> a store, a scheduler or anything else holding runtime state, `design.md` carries a
> `## Live check` section naming what to exercise against the real thing — the running system
> against its real state, not its fakes — the before/after figures to record, and what the
> change **not** working would look like, so a null result is recognisable as failure rather
> than read as success: a check that only asserts a command exited 0 cannot tell a repaired
> system from an untouched one. `flow.verify` runs it (**Live check**,
> `skills/flow/verify-and-handoff.md`); no task in `tasks.md` does. When the planner judges the
> change has no runtime to verify against, the section is one line saying so — the
> justification is required, the check is not. Every step of the check runs at `flow.verify`,
> before the change lands: a step that needs the change landed, a restarted shared service, or a
> write to the change's own integrate state is written as the stand-in that runs in its place — a
> scratch fixture, a workspace's own service, a scratch copy of the record — named in the section
> (kan-1019: a check deferred to integrate was never run). The no-runtime line opens with `None`.

> **Write a verification change so its found defects become their own tasks.** When a plan's task
> exists to exercise a surface and report what it finds — a final-verification pass over a feature
> group, an end-to-end sweep — its `**Allowed-collateral:**` names what the verification commit
> itself writes, the report or record files the pass produces, never the surface it inspects, and
> every real defect it finds becomes a new task appended to `tasks.md`, carrying the field family
> with the fix's own `**Files:**` and `**Allowed-collateral:**`. The
> appended task is what makes the fix declared — `check-task-commit-fields.sh` refuses a commit
> touching paths no task declared — so the change stays self-contained and every fix stays
> traceable to a declared task.

> **Triage the sweep's findings before any fix runs** — a defect a verification change's sweep finds is triaged before anything is edited: its fix is its own appended task, never an edit inside the verification task itself — a pre-existing defect, or one on a surface another change owns, included (**The sweep**, `skills/flow-contracts/known-bugs.md`).

> **Open a task that verifies a reported behaviour with the reproduce-first step.** When a task's
> job is confirming or refuting a report — an operator's defect report, a review or self-review
> finding, a claimed behaviour — its first step reproduces the report against the untouched tree,
> before any edit. Write the step into the task's own step text — "reproduce before touching
> anything; do not guess" — never as intent the task leaves implicit: the implementer reads its
> task, and a report reproduced before anything is touched is what closes a non-defect with no
> code change, while an edit made before reproducing destroys the evidence the task set out to
> gather.

> **Never write a task that creates a worktree.** Every worktree a plan runs in exists before its
> first task: each `## apps` repository's, on branch `spectre/<name>`, is created at `flow.kickoff`
> (`skills/flow/brainstorm.md`) or in **2. Isolate the workspace** (`skills/flow/implement.md`), so a task names that worktree's path and no step
> carries a `git worktree add` (self-review-backlog-flow-6).

> **End a removal task on an absence sweep.** When a task's job is removing text — copy, a setting,
> an identifier — its last verify step is one `git grep -n -F` over every string it removes, each
> passed as its own `-e '<string>'`, scoped to the paths the task touches: the sweep prints
> nothing when the task is done, and any hit means it is not.
> Its paths never cover a later task's `**Files:**` while that file still holds a string the
> sweep removes — `check-plan-shape.sh` refuses that sweep (F18) — nor a comment or doc line a
> step of the plan tells the task to keep.
>
> A removed UI tag, test id or label that a declared `regression checkout` drives
> (**Project configuration — visual verification**,
> `skills/flow-contracts/project-configuration-visual.md`) is swept in that checkout too, with
> the same `-e` list, and the plan carries the task that updates every hit there: kan-998's
> sweep, scoped to its own paths, missed the labels gymie-playwright drove, and the regression
> suite failed 83 tests on them.

> **List the rationale appendix a changed rule's reasoning lives in.** A step that changes a rule
> whose reasoning a `*-rationale.md` appendix records lists that appendix in its task's
> `**Files:**` and updates it.

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
  `Case <N>` label — `<N>` a number with an optional letter suffix, `Case 13b` — is checked by
  that label alone and backticks beside it parse as nothing;
  otherwise every backticked token becomes a declared test name the commit's diff and the tree
  are searched for; any other prose in it is invisible to the check — write what covers the task
  outside the field, never inside it. A task adding none writes a field opening with the literal
  `none`, and `check-task-commit-fields.sh` then never reads that field's backticks as test names
  it must find in the commit's diff. **Bold `**none**` opens such a field; italic `_none_` does
  not** — the recognition ends on a word boundary, and `_` is a word character, so the trailing
  underscore swallows it.
- `**Regression:**` — per declared test, what fails if this task's commit is reverted.
- `**Baseline:**` — the expected test counts, as `before=<N> after=<N>`, on ONE convention the
  plan states in its own header (`> **Baseline convention:**`, seeded below), so later tasks
  cannot drift between counting methods: **chain-position absolutes against each commit's
  parent** — `before` is the count at this task's own commit's parent and `after` at the commit,
  every task measured at its own position in the chain, never a plan-time pair that a fixup
  landing earlier in the chain stales; when a fixup shifts the counts, the guard's re-measurement
  failure names the numbers it measured and the declaration is restated from them, never
  re-derived by hand; **split by source set** — each Baseline counts one source set's own counter
  (one module's `@Test` annotations, one suite's cases), never one tree-wide number that a change
  anywhere shifts; **a word-boundary-safe counter** — whole tokens anchored, so `@Test` never
  substring-matches `@Testcontainers` (KAN-626); and **no pinned ref** — the recorded command
  carries no tree, sha, branch or revision argument anywhere inside it (KAN-661). The counts are
  counted statically from test declarations (`@Test` and its language equivalents), never from a
  test run, and each task's delta its own: `after` is `before` plus the tests the task's own
  `**Tests:**` field names. The counts are exact integers: `after=~N` is unparseable by the guard
  and silently disables the re-measurement. A Baseline that records a plan-provenance `<!--
  measured: <command> @ <ref>` `-->` comment is re-measured once the task's commit exists:
  `check-task-commit-fields.sh` re-runs that command at the commit's parent and at the commit, and
  a count differing from the declared one fails the task — record a command whose stdout is one
  integer (a `| grep -c` pipeline is the shape), and read
  "never from a test run" as constraining how the declared numbers are derived at plan time, not
  the guard's re-measurement. The re-measurement checks each of its two points out itself and
  runs the command in that working tree. The
  comment's `@ <ref>` is the provenance annotation, never part of the command.
- `**Commit:**` — the commit subject line this task's implementer must use, scope naming the module
  the task's own `**Files:**` field carries, per **Commit scopes name the module**
  (`<agents repo>/rules/commit-scope-is-the-module.mdc`).

An `**After:**` field — `Task <id>[, <id>…]`, one `Task` prefix for the whole list (`Task 5, 6`, never `Task 5, Task 6`), or `none` — declares the task's predecessors, and **every
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

Seed `> **Design shape:** plan-mode` into the same block of every plan: on a `tasks.md`
carrying it, `check-plan-shape.sh` refuses a sibling `design.md` missing any of section **C**'s
four plan sections (decision `design-shape-line`, kan-1033).

Seed a third header line, in the same block, into every plan whose tasks declare `**Baseline:**`
fields — the exact line below, copied verbatim, so every plan states its baselines' convention the
same way and no later task can drift between conventions (KAN-678):

```markdown
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.
```

A plan that declares `**Baseline:**` fields and carries no such line fails `check-plan-shape.sh`
(F12). A plan whose tasks declare no Baselines states no convention and writes no line.
Tasks whose Baselines carry the same `measured:` command form one chain in plan order: a task whose `before` is below the previous such task's `after` fails `check-plan-shape.sh` (F17), save the later task of a `**Squash-with:**` pair, which shares its partner's commit.

Before continuing, run `spectre validate "<name>"` — implementation's entry check (section **1**,
`skills/flow/implement.md`), run here so a capture it would refuse never reaches the gate — and
repair any exit-1 finding; its exit 2 is reported and stops the plan.

Before continuing, run `check-plan-shape.sh` — a shipped guard, run unconditionally — and the
project's configured plan-provenance guard and its configured build-green guard, if the project
declares them, and fix any hit. **A hit is exit 1, and only exit 1.** The refusal exits are
never repaired by editing the plan: check-plan-provenance's exit 2 (environment) and
exit 3 (containment — a security finding first, per the guard's own contract) and the
not-a-verdict exit 2 the shape and build-green guards share (a missing or unreadable plan file)
are reported and stop the plan. Provenance's exit 4 (content-classification) is the one refusal
whose remedy is editing the plan — fix the line it names and run the guard again, its own
prescribed course, never accepting the plan while it stands.

### Plan review gate

The summary is plain prose, not a listing of `tasks.md`: what will be implemented and how — the
behaviour being added or changed, the approach taken, the seams and decisions that matter (data
shapes, boundaries, ordering constraints, what was deliberately left out) — in enough detail that
the operator can judge the logic without opening the plan. No per-task rows, no file, test or
commit fields. Then the `## Decision` block, verbatim, under it.

Then the question — `/flow-plan`'s wording is **Push artifacts?**, `/flow`'s is **Proceed to
implementation?** — with exactly two options: **Yes** *(recommended)* and **No (plan needs
updates)**.
