# Brainstorm and plan — sections B, C, D

Sections **B**, **C** and **D** of the brainstorming phase, run by `/flow`'s `planner` dispatch
(**Run brainstorming and planning through the planner role**, `skills/flow/brainstorm.md`) — or, for **C**, **D** and **Decide**, a `/flow-plan` capture
(**Capturing a new change**, `skills/flow-plan/SKILL.md`) — no one relays these sections, though
the session may dispatch the fresh subagents **Explore and design agents** below allows, and
neither does the self-review planner, which runs B, C and D as item 1 of **3. Fix every finding that is
not `big`** (`skills/flow-self-review/SKILL.md`) binds them, and the in-run fix planner, which
runs them as **Pipeline defects found mid-run** (`skills/flow-contracts/pipeline.md`) binds them. Every "you" below addresses that session, or that planner, directly.

In a `planner` dispatch, every **AskUserQuestion** these sections name is written as a question of
the round file instead, and the parent asks it, with the silence outcome the naming section
states.

**Load `skills/spectre-planning/SKILL.md`** — the generic prompts of sections B, C and D; this file carries the flow-only parts.

## B. Basic Workflow #1 — Brainstorming

### The checklist

**Load `skills/flow/withdrawal.md`** only when the linked issue's labels carry `flow-fix`,
`flow-cost` or `flow-speed` — its reachability check opens the checklist, before any design question.

- **A ticket that names a practice or a brief resolves it through the lessons home** (**Process
  lessons**, `skills/flow-contracts/lessons.md`) — never by searching repositories or archived
  narratives for it.

- Save the design to `<project>-worktrees/<name>/.superpowers/sdd/YYYY-MM-DD-<name>-design.md` — the
  worktree `flow.kickoff` created (**A. Resolve the change and write `STARTED`**, `skills/flow/brainstorm.md`). The
  path is gitignored: never stage or commit it, even where the brainstorming skill says to, and
  never write it to the main checkout.
- **A frame-specified design needs its handoff assets reachable from the tree.** When the
  design's specification is a drawn frame — a mockup the implementation must match (**Design mockups are a specification, not an inspiration**, `rules/design-mockups-are-specs.mdc`) — the design is not
  approvable, and no task whose specification is that frame is written, until the handoff assets
  are committed into the repository or their location is recorded in
  `<project>/.flow/project.md` (the `mockups` row of `## visual verification` is where a declared
  mockups directory lives — **Project configuration — visual verification**,
  `skills/flow-contracts/project-configuration-visual.md`). A commit of the assets carries
  `-m "Flow-Key: flow.brainstorm"` (**Commits outside every task**,
  `skills/flow-contracts/git-boundaries.md`).
- **A visible change is prototyped before its design is approved** (`plan-visual-prototype`,
  kan-1038). When the change can alter what a user sees — a page, a view, a chart, a dialog — and
  no mockup frame already specifies it, build a clickable prototype of the proposed result with
  sample data as a single HTML file, publish it with the Artifact tool, and iterate it on the
  operator's feedback each round until they approve; that approval is the convergence confirm's
  **approve the design and move on** (**Convergence**, below). The file draws the proposed page
  alone — no caption, note or screenshot beside it — since `flow.visual-verify` renders the whole
  file as the capture's frame; a note to the operator goes in the message (kan-1047: a caption
  over a JPEG of the live page never paired with a capture). Save the approved file under
  `<project>-worktrees/<name>/.superpowers/sdd/` and cite it, with its Artifact URL, from
  `design.md`'s `## Context` as the spec the implementation and `flow.visual-verify` compare
  against, and every capture test of a view it draws drives each interactive state it shows —
  hover, focus, pressed — so `flow.visual-verify` renders and compares that state too
  (kan-1038). `check-plan-shape.sh` F22 refuses a plan whose design cites the prototype and whose
  tasks touch a `ui paths` glob while no task's `**Files:**` names a file under `screenshots`
  (kan-1047). A change with nothing visible builds none. On a harness without the Artifact tool,
  write the file and give the operator its absolute path. Where the confirm is not asked — the
  `## decisions: recommended` mode, and the self-review and in-run fix planners, which take the
  recommendation (**Auto-resolution**, `skills/flow-contracts/operator-prompts-auto-resolution.md`)
  — the published prototype is the spec without an approval round. **The approved prototype is
  the source of truth** (kan-1038): any departure from it found later in the run — by a reviewer,
  verify or `flow.visual-verify` — is fixed to match it, auto-resolved and never asked; a UI
  change the operator wants after using it is a later change, not this one.

**Load `skills/flow/seeded-note.md`** only when the ask arrives as a seeded research note — a
fully-worked issue description, a `/flow-plan` capture or a handoff package that already carries
the design, its decisions and the acceptance criteria.

**A decision that departs from a handoff ticket's Done-when amends that line before the plan-gate
commit.** When a `## Decisions` entry chooses otherwise than a handoff ticket's `## Done when`
states — a branch, a landing target, an acceptance criterion — rewrite that line to what the
decision chose, citing the decision's `**ID:**` and its date, in its own commit carrying
`-m "Flow-Key: flow.brainstorm"`, so integrate never stops on a criterion the plan already
overruled (kan-1003: decision `lands-on-develop` left "Branch `feature/one-block` in both repos, <!-- citations-guard:allow -->
never develop" standing, and integrate stopped and asked again).

A Done-when names only committed, tracked paths of the repository it belongs to — never a
gitignored or otherwise uncommitted side output, whose removal a later cleanup turns into a done
criterion nothing satisfies — and the archive commit refuses one that does, enforced by **Archive on
the change branch** (`skills/flow-contracts/finish-contract-run1.md`).

### Explore and design agents

**Every planning dispatch** takes its pair from `plan-agent-pair.sh <seed> <n>` — `<seed>` the
change name (in `/flow-plan` before capture, the session token; for the self-review and in-run fix
planners, their own `Key:`), `<n>` counting the session's planning dispatches from 1 — passing the
printed model as the Agent tool's `model` and the effort as its `effort`. Its prompt carries the
line `Role: explorer` for an Explore agent or `Role: plan-designer` for a Plan agent (**Allowed
pairs**, `skills/flow-contracts/model-policy.md`), the
two agent-baseline sentences and the report-tersely line (`~/.claude/rules/dispatch-carries-the-baseline.md`)
and a narrow, explicit scope: what to look at, what is already known, what to return. Never a
`fork`, and never an agent that edits files. The agents investigate and propose; you decide. Each
planning dispatch's row is ended when its agent hands back, and `flow stage end` of
`flow.brainstorm` refuses while one is open (kan-1046 self-review, 1e).

In `/flow` and `/flow-plan` the recommendation goes to the operator as part of the design, its
alternatives shown only when they help the operator decide. The self-review and in-run fix planners
take the recommendation without asking, per **Auto-resolution**
(`skills/flow-contracts/operator-prompts-auto-resolution.md`).

### Convergence

Under the `## decisions: recommended` mode (**Auto-resolution**,
`skills/flow-contracts/operator-prompts.md`) the confirm is not asked: the recommended
**approve the design and move on** is taken, and the silence default (**Convergence**, `skills/spectre-planning/SKILL.md`) never fires.

Under the mode the offer is not asked either: **Yes — run another round** is taken, and the
silence default (**Convergence**, `skills/spectre-planning/SKILL.md`) never fires (**Auto-resolution**,
`skills/flow-contracts/operator-prompts.md`).

The explicit **approve the design and move on** answer is at once the convergence exit that closes
the checklist and the design approval the HARD GATE requires; mark `flow.brainstorm` end, then
`flow.design-approval` begin and end, around that one answer:

```bash
flow stage end   -command '/flow' -stage flow.brainstorm -outcome completed <name>
flow stage begin -command '/flow' -stage flow.design-approval -harness <harness> -session-token mf-<literal-token> <name>
# … the operator's approve-and-move-on answer — this is the HARD GATE (The checklist, skills/spectre-planning/SKILL.md) …
flow stage end   -command '/flow' -stage flow.design-approval -outcome completed <name>
```

## C. Create the change and its artifacts

```bash
flow stage begin -command '/flow' -stage flow.create-artifacts -harness <harness> -session-token mf-<literal-token> <name>
spectre new "<name>"   # working directory: the worktree flow.kickoff created
```

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
> lint alone and no test command. A task whose `**Files:**` holds a Markdown file under `skills/` or
> `rules/` is the exception: its verify step runs the whole list, `run-verify-list.sh <worktree>
> lint`, because the prose guards that list carries are too many to hand-pick and the one a subset
> leaves out fails only at `flow.verify` (kan-1067 self-review, 1-F).

> **Plan the record a feature needs as its own task.** A field, a column, an endpoint, a stage or a
> recorded event the feature reads and the system does not keep yet is a task of the plan, never a
> question for the operator — **Add the missing record, never infer around it**
> (`skills/flow/engineering-principles.md`).

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
> (kan-1019: a check deferred to integrate was never run). When `<project>/.flow/project.md` names
> a tool for the check's stand-in — this repository's `<agents repo>/scripts/live-store-copy.sh` — the section
> invokes that tool by name and never re-specifies what it does (kan-1046 self-review, 4a). The
> no-runtime line opens with `None`.

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

**Committed content never cites a commit sha on the change's own branch.** Integrate's run 1
reshapes that branch (**Run 1 — the branch is not merged**, `skills/flow-contracts/finish-contract-run1.md`),
so a sha a task's committed file names stops resolving; cite a path on the default branch and the
ticket instead (kan-1034: task 2's vendored `UPSTREAM.md` named the task 1 commit, unreachable once
the branch was reshaped).

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
  outside the field, never inside it. A token written `Class.method` is found when both its
  segments are, so a test may be named as the pipeline's `fix-mutation:` lines name it (kan-1059
  self-review, 1f). A task adding none writes a field opening with the literal
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

A relocation whose destination is consumed outside this repository — a skill another repository
vendors — checks, before the plan classes a paragraph as movable, that every file, guard and stage
the paragraph cites resolves in what that destination ships (kan-1034: flow-only paragraphs classed
generic moved into the vendored `skills/spectre-planning/SKILL.md`, and a fix round moved about 230
lines back).

Seed `> **Design shape:** plan-mode` into the same block of every plan: on a `tasks.md`
carrying it, `check-plan-shape.sh` refuses a sibling `design.md` missing any of the four plan sections
(**C. Create the change and its artifacts**, `skills/spectre-planning/SKILL.md`) (decision `design-shape-line`, kan-1033).

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

> **Write a capture for every frame on the change's frame list.** When the design is specified by
> mockup frames and the project declares `mockups`, every frame id `design.md` cites is named in
> some task's `**Files:**` or steps, as the capture spec screenshot of the view that frame draws
> plus its `<spec>.mockups` sidecar line (**Visual verification**,
> `skills/flow-contracts/project-configuration-visual.md`). Before finishing, list the frame ids
> and name the task covering each; a frame no task covers is a plan gap, closed by adding the
> capture to the task that builds that view or by a capture task of its own. **Cite each frame by
> its project-relative path, the `mockups` root joined to `<frame id>.png`, never as a bare
> `<frame id>.png`** — a capture screenshot's name takes that shape too, so only the path form is
> a citation `check-mockup-citations.sh` can resolve.

> **Write a capture task that baselines against a mockup to measure before it commits.** Such a
> task's steps, after its capture, run `compose-mockup-frames.sh` on its spec's `<spec>.mockups`
> sidecar and the band, seam and matrix reads of `measure-visual-properties.sh` on every pair, and
> fix until they come back clean or each leftover is named with its cause — the rule
> **The implementer measures before it hands back** in **The loop** (`skills/flow/verify-fix-loop.md`) states for a fix
> round, applied at plan time (kan-1007: a plan-time capture was judged by eye). The same
> compose-and-measure steps apply when the view's specification is an approved prototype, rendered
> as each capture's frame per step 10 of `skills/flow/visual-verify-verifier.md`, with no declared
> `mockups` root (kan-1046 self-review, 8c).

> **Write a frame a sibling ticket finishes as this change's parts only.** When sibling tickets
> share a mockup frame family and a frame on this change's list also draws parts a sibling ticket
> builds, `design.md` records a decision naming the parts of that frame this change builds, and
> the frame's `<spec>.mockups` sidecar line sits under a `#` comment naming the parts expected to
> differ, the sibling ticket that builds them and that decision's ID — the verifier then reads
> those parts as expected until the sibling lands, and judges every other part of the frame as
> usual (KAN-821).

When the project declares `mockups`, also run
`check-mockup-citations.sh <worktree> <mockups root> <changeRoot>/design.md <changeRoot>/tasks.md`
before continuing. Its exit 1 names a cited frame no file in the mockups root holds: a handoff asset
**The checklist** above still owes, never a frame to plan against — commit the asset or record its
location, or drop the citation and every task that rests on it. Its exit 2 is reported and stops
the plan.

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
steps 1–5 below record the micro row's values without choosing — execution `inline`, implementer
and fixer `skipped — inline`, panel the string `default`, groups `null` — with decision.json
carrying `class: micro`.

Decide, in this order — step 2 only when step 1 came out `sdd`:

1. **execution mode** — `inline` (`class` micro, small or regular) or `sdd` (`class` big).
2. **implementer and fixer model + effort** — only when step 1 came out `sdd`; two pairs, each
   chosen per **Model and effort** below. The implementer pair is every implementer group's
   default (step 5). The fixer pair is the panel-fix subagent's own and is chosen from what a fix
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
   fix-round re-run dispatch runs on — one bundled dispatch carrying the round's re-running roles,
   each targeted at the findings its role raised, the full-roster fallback separating them when
   the bundled findings need separation (**Panel re-runs**, `skills/flow/review-panel-fix-round.md`): its `model` is
   chosen per **Model and effort** below, and its `effort` is `low`, fixed, since a re-run reads a
   delta to confirm a fix and must be short and fast.
4. **extra dispatches** — on every class, `micro` included: record `plan-class.sh`'s two `extra:`
   lines, each `{slots, model, effort, reason}` — the first as `extra_dispatch`, `slots`
   `["test-sanity", "call-efficiency"]`, the second as `extra_dispatch_mutation`, `slots`
   `["mutation", "failure-modes"]` — `model` and `effort` each line's pair, `reason` naming its
   roll. Neither is chosen: the roll decides the model and the effort is `low`, always.
5. **implementer groups** — on every run whose step 1 came out `sdd`: run `plan-dispatch-bundles.sh <changeRoot>/tasks.md`, then
   `plan-dispatch-groups.sh <changeRoot>/tasks.md` for the mechanical default — a deterministic
   grouping biased toward fewer, larger groups (no roll, no static table, no per-group ceiling; at
   most two implementer dispatches in flight per wave), recorded as `groups_mechanical`. `groups`
   is `groups_mechanical` verbatim unless the planner **splits** a mechanical group into more
   groups — a chain judged too long for one implementer's context, or a real parallel-wave benefit
   the mechanical grouping's fold pass declined — with a one-line `groups_override` reason;
   **never merge across the mechanical result**, since merging two groups it kept apart would
   collapse a real parallel wave. **A mechanical group holding two or more
   verification-only bundles** — end-to-end specs or fidelity captures, none named in another's
   after-set — **is split, one group per such bundle**, `groups_override` naming the parallel
   wave, since the chain merge joins each to the group holding the feature task it waits on and
   runs them in series. `groups_override` is `null` when the mechanical grouping is
   taken verbatim; `groups_reason` defaults to the literal `mechanical`, or names the split's own
   reason when `groups_override` is set. **Each group carries its own `model` and `effort`**:
   step 2's implementer value by default, or a different pair the planner picks for that group
   alone per **Model and effort** (a group of mechanical, well-specified tasks at a lower effort; a
   group carrying the change's hardest seam one step up, never past `medium`), the reason in that group's own `reason`.
   `groups`, `groups_mechanical` and `groups_override` are all `null` when step 1 is
   inline.
6. **visual verification** — on every class, `micro` included: it judges what the change can
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
itself, the extra dispatches, step 6's visual judgement, the verify stage, or the self-review.

#### Model and effort

**This section is canonical for how a dispatch's model is chosen.** Every pair a Decide step
assigns — the implementer, the fixer, each panel dispatch, the rerun pair, each implementer group
— carries a planner-chosen `model`, one of `opus` or `sonnet`, within these bounds:

- **implementer, fixer and each implementer group** — `opus` by default; `sonnet` only for simple,
  predictable work near-certain to succeed on the first try, the pair's `reason` saying so; at
  effort `medium` or `low`, never `high`;
- **a pair carrying a hard seam** — concurrency, platform interop, a data-model change,
  performance-sensitive code — `opus`, always;
- **every other pass-1 panel dispatch** — `opus`, always;
- **the extra dispatches** — `test-sanity+call-efficiency` and `mutation+failure-modes`, each its own rolled pair off the `panel-extra` row of **Allowed pairs** (`skills/flow-contracts/model-policy.md`): the only pass-1 panel dispatches that may run on `sonnet` (decision `extra-dispatch-sonnet-carve-out`, kan-959; kan-1028);
- **every gated per-task reviewer bundle** — its **Allowed pairs** row (`skills/flow-contracts/model-policy.md`), whatever the roll, since no decision pair records it, and a fix re-review bundle its own;
- **the rerun pair** — `opus` or `sonnet`; nothing requires it to differ from a pass-1 dispatch.

A micro decision records no pair, and a dispatch with no recorded pair — a micro panel,
**review-panel.md**'s no-decision dispatch, the tooling analyst — runs on the literal `opus`. `flow record
decision` refuses a decision whose pairs name any other model. Pairs may repeat — two dispatches,
or a pass-1 dispatch and the rerun pair, on the same model and effort is not a defect. The rerun
pair's effort is fixed at `low` (step 3). **A verification-only implementer group is `opus` at `low`, fixed, whatever the
roll** — a group whose every task only adds end-to-end specs, fidelity captures or baselines
against design frames of an already-built feature, and changes no
production code. **No implementer, fixer or implementer group runs at `high`**, on either side of
the roll: a hard seam takes `medium`. The one `high` fix is the no-progress escalation of **Fewest
operator actions** (`skills/flow-contracts/pipeline.md`), a dispatch no decision records. `flow
record decision` refuses a decision whose implementer, fixer or group pair names `high`. **When
`effort_roll < 90`, the implementer, the fixer and every implementer group take `low` — `medium` where
the pair carries a hard seam.** **Every pass-1 panel dispatch takes `low` when `effort_roll < 60` and `medium` otherwise, whatever the band** — 60/40 per
change. **`high` reaches a pass-1 panel dispatch — a `panel.dispatches` entry, never a gated bundle — one way only:** the operator's explicit **Yes** at the plan
gate (**Plan review gate** below), and that dispatch's prompt opens with the line `Operator-approved high`.
When `effort_roll ≥ 90`, the implementer, fixer and group efforts are the planner's own choice,
one of `low`/`medium` — decided from what that dispatch will actually
do: the complexity of its tasks, the time and space
complexity of the code it writes or reviews, and the scalability the change has to hold up under.
A mechanical, well-specified dispatch sits at the cheap end; a dispatch carrying a concurrency
seam, a data-model change or a performance-sensitive path sits at the expensive end; nothing in
between is a default. Each pair carries a one-line `reason` beside it in the JSON; the `## Decision`
block prints it only for a group whose pair departs from the implementer's. **On harnesses `zcode` and `muse` the chosen pair is recorded as chosen and
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
`{verify, reason}`, `verify` one of `required`, `skipped` or `not configured`), `extra_dispatch` and `extra_dispatch_mutation` (each an object `{slots, model, effort, reason}`, recorded on every class), `parent` (`{model, effort}`, the parent's own
model/effort, per **Harness mapping** (`skills/flow-contracts/model-policy.md`): `unknown
(agent-defined)` and `default` where the session does not state them), `overrides` (session-instruction
overrides to a *result*, each replacing the pair(s) it names for this run; empty unless one was
given), `session_cap` (the run-wide model and effort ceiling a session instruction or a saved memory states — `{model, effort, instruction}`, or `null` when none is given; `flow record dispatch begin` refuses a dispatch above it, an unstated `default` effort included).

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
print the plan summary (**Plan review gate**, `skills/spectre-planning/SKILL.md`), then ask the gate question in one **AskUserQuestion** call. The
gate is mandatory: no answer, no commit and no implementation. It runs at the end of every
`/flow-plan` capture (**Capturing a new change**, `skills/flow-plan/SKILL.md`) and in `/flow`
after the `flow.decide` record sequence (**Run brainstorming and planning through the planner role**,
`skills/flow/brainstorm.md`).

Then the question — `/flow-plan`'s wording is **Push artifacts?**, `/flow`'s is **Proceed to
implementation?** — with exactly two options: **Yes** *(recommended)* and **No (plan needs
updates)**.

**`/flow` never asks its gate question: it takes **Yes** at once**, at the operator's instruction
(2026-10-26). The summary and the `## Decision` block still print, the high-raise question below is
not asked and takes **No**, and the run continues to the plan-gate commit and the `STARTED`
handoff. An operator who wants the plan changed says so before running the handoff's `/flow
<name>` — that message re-plans as **No** below. `/flow-plan`'s **Push artifacts?** is still asked.

**A `panel.dispatches` entry the planner judges worth `high`** — a hard seam the roll's effort may not carry —
adds one question per such dispatch to the same AskUserQuestion call: **Raise `<dispatch>` to
`high` — `<seam>`?**, **No** *(default, recommended)* / **Yes**. Only an explicit **Yes**
re-records that pair at `high` through the `flow.decide` record sequence before implementation;
anything else leaves the roll's effort. Under the `## decisions: recommended` mode the question is
not asked and **No** is taken.

On **No**,
take the operator's changes (a follow-up **AskUserQuestion** round when the option carried none),
revise `tasks.md`, re-run `spectre validate "<name>"`, `check-plan-shape.sh` and the project's configured guards, re-run the
Decide step from `plan-class.sh` on — in `/flow`, the `flow.decide` record sequence again, a second
decision row — then print the summary and ask again. Loop until **Yes**. **Yes** is the only exit.
Under the `## decisions: recommended` mode the gate takes **Yes** without asking — the summary
and the `## Decision` block still print, so the record of what was approved stays complete
(**Auto-resolution**, `skills/flow-contracts/operator-prompts.md`).

`/flow-plan` commits and ends with its own report (**Capturing a new change**, `skills/flow-plan/SKILL.md`).
