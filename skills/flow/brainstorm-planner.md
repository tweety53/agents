# Brainstorm and plan — sections B, C, D

Sections **B**, **C** and **D** of the brainstorming phase, run by `/flow`'s `planner` dispatch
(**Run brainstorming and planning through the planner role**, `skills/flow/brainstorm.md`), in
`/flow`'s creating run and in a `/flow-plan` capture alike — no one relays these sections, though
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

- **Every ambiguity is classified, and only an `[architectural]` or `[impactful: <criterion>]` one
  is asked — in `/flow`'s planning up to its plan gate and in a `/flow-plan` session, and nowhere
  else.** An ambiguity is any requirement, scenario, name, format, state, edge case, ordering,
  error behaviour or scope boundary of the change that the request, the ticket, a mockup and the
  code leave open — any point two competent implementers could read differently. One is
  impactful when one of four criteria holds, named in its heading: `no-recommendation` — no
  option can honestly be recommended, an answer contradicting an earlier one included;
  `irreversible` — a wrong pick is costly to undo once landed: persisted or migrated data, a
  public API, file format, CLI or cross-repository contract, a deletion, permissions, security,
  privacy, money; `scope` — the pick adds or drops something the ticket or a mockup states;
  `plan-size` — the options differ by a task or more and the ticket gives no lead. An asked
  question carries a recommended option where one honestly exists and is asked until the user's
  own answer closes it (**Convergence**, `skills/spectre-planning/SKILL.md`). Silence answers
  nothing: the question stays held and is asked again in the next round. Any other ambiguity
  takes its recommended option: it is recorded as a `## Decisions` entry and as one line,
  `<decision id>: <chosen> — <what it was chosen over>`, under the round file's `## Auto-taken`,
  and in `answers-<r>.md`; the parent's relay never prints, asks or mentions it, and the next
  planner and the final planner read it from the file. At most four `[architectural]` and at most four `[impactful: …]`
  questions are asked a round; beyond four, the four whose wrong pick costs most are asked and the
  rest are held for the next round, never auto-taken. Each question is asked in the round it
  appears: from an answer, an agent's findings or the builder check, or, once the design is
  approved, from writing the artifacts or the plan, then asked before the plan review gate and its
  answer revising the design. This is a scoped override of every call `brainstorming` makes for
  the user — its **Minor** pile among them — of the single round of questions its builder check
  allows, and of **The checklist**'s display-only-formatting bullet and one-call batching
  (`skills/spectre-planning/SKILL.md`). The self-review and in-run fix planners never apply it:
  they take the recommendation (**Auto-resolution**,
  `skills/flow-contracts/operator-prompts-auto-resolution.md`).

- **Now or later is never asked.** Where the options are doing work the design needs in this
  change or deferring it to a later change or ticket, now is taken and recorded under
  `## Decisions`.

- **An architectural question is marked, and each of its options says what it fails under**. A question whose options are approaches — the alternatives the competing-designs
  step produces, each becoming a `**Considered:**` line — and not a name, a format, a wording or
  a scope detail is headed `### Q<n>. [architectural] <question>` in the round file, and each of
  its options closes with `Fails under: <the condition this option stops working in>`. The
  recommended mark stays. The parent asks the architectural questions together, up to four in one
  **AskUserQuestion** call, in round-file order; the impactful ones in a second call; the
  convergence confirm last, in the impactful call when it has room.

- **An answer that contradicts an earlier one is asked about, never applied**. When
  an answer in `answers-<r>.md` contradicts an earlier answer or a decision the design draft
  records, change nothing on that point: quote both verbatim in the next `round-<r>.md` and ask
  which stands, with no recommended option — so it is asked under the `## decisions:
  recommended` mode too. This holds in planning only; past the plan gate nothing changes.

- **A round file opens with `## Summary`** — at most ten lines: the decisions the operator's answers settled
  and what is still open. The Summary names no auto-taken decision — an auto-taken pick lives only
  under `## Auto-taken` and in `answers-<r>.md`. The parent relays that section and the file's absolute path, and the
  design draft stays in the file. The design draft carries `### Settled` — every answer given so
  far, the operator's words verbatim beside the question and its round — so the latest round file
  alone is the design's whole state. The section ends with `**Scope:** small|regular|large — <one-line
  reason>`, the planner's estimate after exploring: `small` a change confined to one known area
  with one obvious approach, `large` one that spans repositories or apps, changes a data model or
  a public contract, or needed competing designs; `regular` otherwise.

- Save the design to `<project>-worktrees/<name>/.superpowers/sdd/YYYY-MM-DD-<name>-design.md` — the
  worktree kickoff created (**A. Resolve the change and write `STARTED`**, `skills/flow/brainstorm.md`). The
  path is gitignored: never stage or commit it, even where the brainstorming skill says to, and
  never write it to the main checkout.
- **A frame-specified design needs its handoff assets reachable from the tree.** When the
  design's specification is a drawn frame — a mockup the implementation must match (**Design mockups are a specification, not an inspiration**, `rules/design-mockups-are-specs.mdc`) — the design is not
  approvable, and no task whose specification is that frame is written, until the handoff assets
  are committed into the repository or their location is recorded in
  `<project>/.flow/project.md` (the `mockups` row of `## visual verification` is where a declared
  mockups directory lives — **Project configuration — visual verification**,
  `skills/flow-contracts/project-configuration-visual.md`).
- **A visible change is prototyped before its design is approved** (`plan-visual-prototype`). When the change can alter what a user sees — a page, a view, a chart, a dialog — and
  no mockup frame already specifies it, build a clickable prototype of the proposed result with
  sample data as a single HTML file under `<project>-worktrees/<name>/.superpowers/sdd/`, and
  iterate it on the operator's feedback each round until they approve; that approval is the convergence confirm's
  **approve the design and move on** (**Convergence**, below). The prototype is
  built on the app's own shell and styles — the real header and nav, the project's theme,
  stylesheet tokens and existing component classes, the content width and chart size of the
  sibling views — never a standalone restyle. The file draws the proposed page
  alone — no caption, note or screenshot beside it — since `flow.visual-verify` renders the whole
  file as the capture's frame; a note to the operator goes in the message. **The operator sees the prototype
  whenever one exists**: a planner dispatch has no Artifact tool, so it only writes the
  file, and the session that dispatched it publishes it with the Artifact tool after every return
  that wrote or changed it — a scratchpad copy adapted to the tool's page contract, with no
  doctype, `html`, `head` or `body` wrappers and the title and viewport meta at its top, while the
  file under `<project>-worktrees/<name>/.superpowers/sdd/` stays whole — and prints its URL to the operator: `/flow`'s parent
  in the relay (**Run brainstorming and planning through the planner role**,
  `skills/flow/brainstorm.md`), the self-review and in-run fix loops' parent in its final report.
  A session running this section itself publishes it the same way. Only where the harness has no
  Artifact tool at all does that session print the file's absolute path instead, saying plainly
  that it could not be published — never silently. Cite the approved file, with its Artifact URL
  or, where it could not be published, that it was not, from `design.md`'s `## Context` as the
  spec the implementation and `flow.visual-verify` compare
  against, and every capture test of a view it draws drives each interactive state it shows —
  hover, focus, pressed — so `flow.visual-verify` renders and compares that state too. A change with nothing visible builds none. Where the confirm is not asked — the
  `## decisions: recommended` mode, and the self-review and in-run fix planners, which take the
  recommendation (**Auto-resolution**, `skills/flow-contracts/operator-prompts-auto-resolution.md`)
  — the published prototype is the spec without an approval round. **The approved prototype is
  the source of truth**: any departure from it found later in the run — by a reviewer,
  verify or `flow.visual-verify` — is fixed to match it, auto-resolved and never asked; a UI
  change the operator wants after using it is a later change, not this one.

**Load `skills/flow/seeded-note.md`** only when the ask arrives as a seeded research note — a
fully-worked issue description, a `/flow-plan` capture or a handoff package that already carries
the design, its decisions and the acceptance criteria.

### Explore and design agents

**Every Explore and Plan agent** runs on its role's pair (**Model and effort** below) — `explorer`
for an Explore agent, `plan-designer` for a Plan agent — passed as the Agent tool's `model` and
`effort`. Its prompt carries the two agent-baseline sentences and the report-tersely line
(`~/.claude/rules/dispatch-carries-the-baseline.md`) and a narrow, explicit scope: what to look at,
what is already known, what to return. Never a `fork`, and never an agent that edits files. The
agents investigate and propose; you decide. They record no ledger row.

**Findings outlive the round.** Append each Explore and Plan hand-back verbatim to
`<abs-worktree>/.superpowers/sdd/planner/findings.md` under `## <explore-<n>|design-<n>> — <the
scope it was given>` as it arrives; read that file before dispatching, and dispatch an agent only
for ground no section covers.

**The cap.** One planner dispatch launches at most three Explore and three Plan agents in round 1
and at most two agents in all in a later round or a final dispatch. A planner that needs more
writes `Over cap: <n> agents — <reason>` in the round file's `## Summary` (a final dispatch: in
its return) and dispatches them.

**Never end your turn while one of your agents runs.** A turn that ends to wait for a hand-back
loses your 5-minute prompt cache when the agent runs longer, and the hand-back then rewrites your
whole context; each exit-1 return of the wait below is one cache read, and re-arms the cache.
Once your own work is done, call
`AWAIT_SUB_AGENTS_SECONDS=240 <agents repo>/scripts/flow-guard await-sub-agents <agent id>...` with
the agent ids your launches returned and the Bash tool's `timeout` at 250000, so a call returns
within four minutes. Exit 0: every agent finished or is stale — silent for 30 minutes — so read
each hand-back, treat a `stale:` agent as returning nothing and say so in the round file. Exit 1:
call again at once, doing nothing between calls. Exit 2: a usage error, reported. The
`await-own-agents` SubagentStop hook refuses your turn end while one runs.

In `/flow` and `/flow-plan` the recommendation goes to the operator as part of the design, its
alternatives shown only when they help the operator decide. The self-review and in-run fix planners
take the recommendation without asking, per **Auto-resolution**
(`skills/flow-contracts/operator-prompts-auto-resolution.md`).

### Convergence

Under the `## decisions: recommended` mode (**Auto-resolution**,
`skills/flow-contracts/operator-prompts.md`) the confirm is not asked: the recommended
**approve the design and move on** is taken, and the silence default (**Convergence**, `skills/spectre-planning/SKILL.md`) never fires.

**A fourth option, `/flow`-only.** The convergence confirm (**Convergence**, `skills/spectre-planning/SKILL.md`) carries one more named option, listed right after **Nothing unclear — approve the design and move on**:

> - **Dig deeper — quiz me on this design first** *(never recommended, never the default)*

Only once the operator picks it — never before, never in anticipation of the confirm — it first dispatches one Explore agent (**Explore and design agents**, above) to collect what the quiz
needs — the problem and why it exists, the design as settled and why, the code it touches, its edge
cases and what it impacts — and return it as a briefing. It then loads `skills/flow-learn-quiz/SKILL.md`
and runs it over that briefing, the operator's understanding being what it checks, until that skill ends. It is a round, and the confirm is
then asked again. Under the `## decisions: recommended` mode the confirm is not asked, so the option never
shows. Silence still defaults to another round, never to this option.

The explicit **approve the design and move on** answer is at once the convergence exit that closes
the checklist and the design approval the HARD GATE requires (**The checklist**,
`skills/spectre-planning/SKILL.md`).

## C. Create the change and its artifacts

```bash
spectre new "<name>"   # working directory: the worktree kickoff created
```

`STARTED` is written before this section exists (section A of `skills/flow/brainstorm.md`), but the
`STARTED` handoff prints at the end of the run, so its `Open questions` line shows what this section holds
(**The `STARTED` handoff block**, `skills/flow/brainstorm.md`).

**`design.md` also carries `## Non-goals` and `## Risks`**: what the change
deliberately does not do, and what can go wrong and what limits it. Each may be the one line
`none — <reason>`.

A `/flow` or `/flow-plan` final planner continues straight into section **D** in the same dispatch.

## D. Basic Workflow #3 — Writing plans

**Tell writing-plans the task shape, because it is spectre's and not writing-plans' own.** A task
is a column-0 checkbox line, `- [ ] <n>. <title>`, whose `<n>` is a flat integer; that task's
steps are `  - [ ] **Step N: …**` lines indented two columns beneath it.

**Each task carries these fields**, each on its own column-0 line:

- `**Files:**` — the repo-relative paths this task's commit will touch, each backticked.
- `**Tests:**` — the tests this task adds, each backticked, or `none`.
- `**After:**` — `Task <id>[, <id>…]` (one `Task` prefix for the whole list) or `none`: the tasks
  this one's own files depend on. Every task writes it, so the implementer grouping
  (`flow-guard plan-dispatch-bundles`) never has to guess the order.
- `**Size:**` — optional, `small`, `medium` or `big`; a task without one is `medium`. Write `big`
  on a task expected to fill most of an implementer's context alone. `flow-guard
  plan-dispatch-groups` weighs a `big` task 3 and every other task 1 and caps every group at 3.
- `**Commit:**` — the commit subject line, scope naming the module the task's `**Files:**` carry
  (**Commit scopes name the module**, `<agents repo>/rules/commit-scope-is-the-module.mdc`).

Every `## Decisions` entry whose `**Chosen:**` states behaviour is pinned by a test some task
names in `**Tests:**`.

> **Write each task's verify step as its own lint commands plus targeted tests.** A verify step
> names the lint commands the task's own `**Files:**` actually need — never the project's whole
> `## lint` list — and the build tool's own selector for each `**Tests:**` entry (`--tests
> '<class>'` for Gradle, `-run '<name>'` for `go test`, `-t '<name>'` for vitest), never the bare
> module or repository suite: that run belongs to the parent's test-list run after the last
> group (`skills/flow/implement.md`) and to `flow.verify`. A task whose `**Tests:**` is `none` names
> lint alone and no test command.

> **Run only what is installed.** A step's command runs a tool already on the machine or declared
> in `<project>/.flow/project.md` — never npx without `--no-install` or `--no`: with its yes flag,
> or bare in an agent's shell, where no TTY makes npm assume it, npx fetches and runs a package
> the machine lacks: the external code auto mode denies. A package in npx's cache runs as
> `npx --no-install <bin>`; one nowhere on the machine is a dependency a task adds
> (`rules/dependency-versions.mdc`).

> **Plan the record a feature needs as its own task.** A field, a column, an endpoint, a stage or a
> recorded event the feature reads and the system does not keep yet is a task of the plan, never a
> question for the operator — **Add the missing record, never infer around it**
> (`skills/flow/engineering-principles.md`).

> **Write a feature's UI tests as their own follow-on task.** When a feature's tests live in a
> UI-test source set of their own — Compose Multiplatform's `desktopTest`, and the like — the plan
> carries two tasks: the feature task, whose `**Files:**` names the source files and the unit-test
> (`commonTest`) files, then one follow-on task per feature task whose `**Files:**` names only that
> feature's UI-test file(s), all of them, its `**After:**` naming the feature task — so the UI-test
> iteration starts in a fresh implementer at a small context.

> **Write end-to-end spec and fidelity-capture tasks to run side by side.** Each such task's
> `**After:**` names only the feature tasks whose behaviour it exercises — never another
> verification task — and each seeds its own accounts and data, so the tasks share one running
> stack (**One stack for the verification tasks**, `skills/flow/implement.md`) without touching
> each other's state. Global Constraints names the stack's start row and any build the tasks run
> against; no task's steps start, stop or rebuild the stack.

> **Write a live check when the change touches a running service or persistent state — never a
> live-verification task.** When the change under plan touches a long-running service, a daemon,
> a store, a scheduler or anything else holding runtime state, `design.md` carries a
> `## Live check` section naming what to exercise against the real thing — the running system
> against its real state, not its fakes — the before/after figures to record, and what the
> change **not** working would look like, so a null result is recognisable as failure rather
> than read as success. `flow.verify` runs it (**Live check**,
> `skills/flow/verify-and-handoff.md`); no task in `tasks.md` does. When the change has no
> runtime to verify against, the section is one line opening with `None` and saying why. Every
> step of the check runs at `flow.verify`, before the change lands: a step that needs the change
> landed, a restarted shared service, or a write to the change's own integrate state is written
> as the stand-in that runs in its place — a scratch fixture, a workspace's own service, a
> scratch copy of the record — named in the section. When `<project>/.flow/project.md` names a
> tool for the check's stand-in — this repository's `<project>/scripts/flow-guard
> live-store-copy`, run from the change's own worktree — the section invokes that tool by name.

> **Write a verification change so its found defects become their own tasks.** When a plan's task
> exists to exercise a surface and report what it finds — a final-verification pass over a feature
> group, an end-to-end sweep — every real defect it finds becomes a new task appended to
> `tasks.md`, never an edit inside the verification task itself (**The sweep**,
> `skills/flow-contracts/known-bugs.md`).

> **Never write a task that creates a worktree.** Every worktree a plan runs in exists before its
> first task: each `## apps` repository's, on branch `spectre/<name>`, is created at kickoff
> (`skills/flow/brainstorm.md`) or in **2. Isolate the workspace** (`skills/flow/implement.md`).

> **Sort a multi-repository plan's tasks under `## <repository>` headings.** Each repository's tasks sit under a `## <repository>` heading naming it by the basename of its `Repo root` in `<project>/.flow/project.md`'s `## apps` (`## agents`, `## gymie`); `flow-guard plan-dispatch-groups` reads those headings and never groups tasks of two repositories together. Any other heading — `## Dashboard`, `## Global constraints`, or a root nested in another declared root, agents' `## stats` — is a section of the repository above it, never a repository of its own.

> **Put a scratch clone in the change's state dir.** A clone a task prescribes — `git clone --local <worktree> <dir>` to try something on a copy — lives under the path `flow state dir` prints, never in a `<project>-worktrees/` directory, and the task removes it with `rm -rf <dir>`.

> **End a removal task on an absence sweep.** When a task's job is removing text — copy, a setting,
> an identifier — its last verify step is one `git grep -n -F` over every string it removes, each
> passed as its own `-e '<string>'`, scoped to the paths the task touches: the sweep prints
> nothing when the task is done, and any hit means it is not.

**A premise about how a guard, harness or tool behaves is checked at plan time, never assumed.**
When a task rests on what a command does — an exit code, a config key taking effect, how a clock
or a rebuild behaves — run it (a spike of five minutes or less) before writing the task and say
in the task what the run showed. A premise the run disproves strikes or rewrites the task; one it
cannot settle in five minutes is named in the task for the implementer to check first.

**A timing gate never presumes an idle machine.** A step whose outcome turns on a measured
duration first waits, on a Monitor until-loop with a stated bound and never a foreground sleep,
until the one-minute load average (`uptime`) is at or under the core count (`getconf
_NPROCESSORS_ONLN`), and records that load average beside the figure.

**Committed content never cites a commit sha on the change's own branch** — integrate may reshape
that branch, so a sha stops resolving; cite a path on the default branch and the ticket instead.

**A parent action never writes a main checkout.** `hooks/protect-main-checkout.py` denies every
write there, so a step that must write a main checkout is an operator handoff line instead, naming
the exact command for the operator to run.

> **Write a capture for every frame on the change's frame list.** When the design is specified by
> mockup frames and the project declares `mockups`, every frame id `design.md` cites is named in
> some task's `**Files:**` or steps, as the capture spec screenshot of the view that frame draws
> plus its `<spec>.mockups` sidecar line (**Visual verification**,
> `skills/flow-contracts/project-configuration-visual.md`). Cite each frame by its
> project-relative path, the `mockups` root joined to `<frame id>.png`. Every screenshot such a
> task's spec takes is the full page or viewport, never an element or a clipped region, and each
> captures a state no sibling capture shows. The task that builds the view seeds, in its capture
> fixture, the data for every element and state the frame or prototype draws — per-row variants,
> empty and zero cells, badges, an overflow-length list — on data outside the windows other
> views' captures read.

> **Write a frame a sibling ticket finishes as this change's parts only.** When sibling tickets
> share a mockup frame family and a frame on this change's list also draws parts a sibling ticket
> builds, `design.md` records a decision naming the parts of that frame this change builds, and
> the frame's `<spec>.mockups` sidecar line sits under a `#` comment naming the parts expected to
> differ, the sibling ticket that builds them and that decision's ID.

**Count the plan's size and propose a split when it is big.** Once `tasks.md` is written, count
its tasks and the distinct frames `design.md` cites. Above 15 tasks or 25 frames, write
`## Story split` into `design.md`: two or more slices, each a title, one line of what it delivers
on its own, and the task ids it takes — every slice shippable alone, the first one the slice the
others build on. The gate asks whether to split (**Plan review gate** below).

Before returning, run `spectre validate "<name>"` and repair any exit-1 finding; its exit 2 is
reported and stops the plan.

### Plan review

**Your `PLAN-WRITTEN` return says whether the plan is reviewed**. A change that is not trivial —
the competing-designs test of **Explore and design agents** (`skills/spectre-planning/SKILL.md`) —
returns the word alone, and the parent runs **Plan review** (`skills/flow/brainstorm.md`). A change
with one obvious approach returns the word and, on the next line, `Plan review: skip — <reason>`,
the reason being the one-approach reason `design.md` records under `## Recommended approach`. The
self-review and in-run fix planners and a fix run's re-plan never write that line and are never
reviewed.

**A dispatch whose prompt names a plan-review report triages it.** It reads the report,
`proposal.md`, `design.md` and `tasks.md`; it opens a file under
`<abs-worktree>/.superpowers/sdd/planner/` only for a finding whose fix turns on a round's answer.
Append `## Triage` to the report, one line per finding: `F<k> fixed — <what changed>`, `F<k>
contested — <the repository path:line or artifact line the reviewer could have read>`, `F<k>
escalated`, or, for an `advisory` only, `F<k> declined — <reason>`. Fix in the artifacts a finding
with one obvious fix. Escalate a finding whose fix would change a `## Decisions` entry, or that
you would contest without evidence the reviewer could have read: return `QUESTIONS`, the next
`round-<r>.md` quoting the finding verbatim beside the options. Write each finding still standing
into `design.md`'s `## Risks` as `Accepted: <the finding> — <why>`, re-run `spectre validate`, and
return `PLAN-WRITTEN`.

### Decide

Run `flow-guard plan-class <changeRoot>/tasks.md <repos> <abs-worktree> <merge-base>` — `<repos>` is the
number of distinct repository roots the plan's `**Files:**` fall under (the project's `## apps`
table; `1` when every path is in this one), `<abs-worktree>` the worktree kickoff created, and
`<merge-base>` this run's working-notes merge base. Its `class:` line is the class — `micro`,
`small`, `regular` or `big` — and its `panel:` and `experimental:` lines are the review panel for
that class, printed rather than derived by hand (the rules are the script's header). **You may
raise the class one step, never lower it**, with a one-line reason recorded as `override`; on a
raise, re-run the script with `-class <class>` first, so its panel lines are the raised class's.

Decide, in this order:

1. **execution mode** — `inline` (`class` micro) or `sdd` (every other class).
2. **review panel** — the roster the script's `panel:` and `experimental:` lines print; on
   `micro`, the string `default` (`REVIEWERS`, `skills/flow/SKILL.md`).
3. **implementer groups** — only when step 1 came out `sdd`: run `flow-guard plan-dispatch-bundles
   <changeRoot>/tasks.md`, then `flow-guard plan-dispatch-groups <changeRoot>/tasks.md`, and take its
   groups verbatim (`null` when inline).
4. **visual verification** — `visual` is `{verify, reason}`. `not configured` when
   `<project>/.flow/project.md` declares no `## visual verification` section. Otherwise judge from
   the plan's `**Files:**` whether anything a user sees can change — a page, a response a page
   consumes, the CORS, auth or routing the frontend depends on: `required` when it can, `skipped`
   when it cannot, the `reason` naming why. A path inside a declared `ui paths` glob is not by
   itself a reason for `required`; `required` answers genuine doubt, never an obvious no-UI change.
   A fix run may raise `skipped` to `required` and never lowers it.
5. **slot judge** — write the plan's tasks to
   `<abs-worktree>/.superpowers/sdd/decide-slots.md` and run `flow-guard slot-judge` on it. Record
   its three lines as `slot_triggers`, keyed by slot, each `{verdict, reason}`. Its exit 2 records
   no `slot_triggers`, which runs every gated slot (**The slot triggers**,
   `skills/flow/review-panel.md`).

Write the decision JSON to `<abs-worktree>/.superpowers/sdd/decision.json`: `class`,
`classMechanical`, `override`, `execution`, `panel`, `groups` (objects `{bundles}`, or `null`),
`visual`, `slot_triggers`, `overrides` (session instructions that change a dispatch's pair for
this run; empty unless one was given). No dispatch's model or effort is recorded here — every
dispatch takes its role's pair (**Model and effort** below).

Render it and print the output verbatim — the `## Decision` block; the `STARTED` handoff prints the
same output again (**The `STARTED` handoff block**, `skills/flow/brainstorm.md`):

```bash
flow decision render -file <abs-worktree>/.superpowers/sdd/decision.json -session-model '<the model named in this session's own system prompt>' -reviewers '<REVIEWERS>'
```

It exits 0 on a printed block and 2 on a usage error or an unreadable file — correct the file and
render again.

#### Model and effort

**Every dispatch runs on its role's first base pair**, read from:

```bash
flow policy get
```

— the role's `base` list, its first entry. A pair is `<model>/<agent>`: `opus/flow-low` dispatches
`subagent_type: flow-low` on `model: opus`; `opus/flow-planner-medium` is `flow-planner-medium` on
`opus`; `opus/Explore@low` is `subagent_type: Explore` with `effort: low`. A dispatch outside every
role runs on `opus` or `sonnet`. A plain-language session instruction overrides the pair it names
for this run only (**Model policy**, `skills/flow-contracts/model-policy.md`), recorded in the
decision's `overrides`. On harnesses `zcode` and `muse` the pair is replaced at dispatch
(**Harness mapping**, `skills/flow-contracts/model-policy.md`).

### Plan review gate

**The plan is never acted on unread.** Once the Decide step has printed its `## Decision` block,
print the plan summary (**Plan review gate**, `skills/spectre-planning/SKILL.md`) and, whenever `design.md` cites a prototype, the line `**Prototype:**` of **The `STARTED` handoff block** (`skills/flow/brainstorm.md`) under it. The
summary ends on two lines of their own: `**Design:** <absolute path of design.md>` and
`**Plan review:** <absolute path of the report> — clean | <n> fixed, <n> contested, <n>
accepted | not run | skipped — <reason>`.

Then the question — `/flow-plan`'s wording is **Save the plan?**, `/flow`'s is **Proceed to
implementation?** — with exactly two options: **Yes** *(recommended)* and **No (plan needs
updates)**.

**`/flow` never asks its gate question: it takes **Yes** at once** — unless `design.md` carries
`## Story split`. The summary and the `## Decision` block still print, and the run ends with the
`STARTED` handoff. An operator who wants the plan changed says so before running the handoff's
`/flow <name>` — that message re-plans as **No** below. `/flow-plan`'s **Save the plan?** is
still asked.

**A plan with `## Story split` asks one question first**, in `/flow` and `/flow-plan` alike, the
slices printed above it: **Split into <n> stories?** — **Split** *(recommended)* / **Build it
whole**. On **Split**, create one Jira story per slice after the first — its summary the slice
title, its description the slice's line, the task titles it takes and the absolute path of this
change's `design.md` — linked to this change's issue (**Follow-up issues**,
`skills/flow-contracts/jira-followups.md`), then dispatch a fresh final planner that narrows
`tasks.md` to the first slice and removes `## Story split`, and re-run Decide. On **Build it
whole**, remove `## Story split` and go on. Under the `## decisions: recommended` mode the question
is not asked and **Split** is taken.

On **No**, take the operator's changes (a follow-up **AskUserQuestion** round when the option
carried none), write them verbatim to the next `answers-<r>.md`, and dispatch a fresh final planner
on them as a `QUESTIONS` return is taken (**Run brainstorming and planning through the planner
role**, `skills/flow/brainstorm.md`) — it revises `tasks.md` and re-runs `spectre validate
"<name>"`. Then re-run the Decide step and record the decision again, print the summary and ask
again. Loop until **Yes**. **Yes** is the only exit. Under the `## decisions: recommended` mode the
gate takes **Yes** without asking — the summary and the `## Decision` block still print, so the
record of what was approved stays complete (**Auto-resolution**,
`skills/flow-contracts/operator-prompts.md`).

`/flow-plan` ends with its own report (**Capturing a new change**, `skills/flow-plan/SKILL.md`).
