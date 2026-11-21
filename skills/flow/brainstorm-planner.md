# Brainstorm and plan — sections B, C, D

Sections **B**, **C** and **D** of the brainstorming phase, run by `/flow`'s `planner` dispatch
(**Run brainstorming and planning through the planner role**, `skills/flow/brainstorm.md`), in
`/flow`'s creating run and in a `/flow-plan` capture alike — no one relays these sections, though
the session may dispatch the fresh subagents **Explore and design agents** below allows, and
neither does the self-review planner, which runs B, C and D as item 1 of **3. Fix every finding that is
not `big`** (`skills/flow-self-review/SKILL.md`) binds them, and the self-review planner, which
runs them as **Pipeline defects found mid-run** (`skills/flow-contracts/pipeline.md`) binds them. Every "you" below addresses that session, or that planner, directly.

In a `planner` dispatch, every **AskUserQuestion** these sections name is written as a question of
the round file instead, and the parent asks it, with the silence outcome the naming section
states.

## B. Basic Workflow #1 — Brainstorming

Investigate read-only, then present the design.

### The checklist

- **Every ambiguity is either asked or auto-taken — asked only in `/flow`'s planning up to its
  plan gate and in a `/flow-plan` session, and nowhere else.** An ambiguity is any requirement,
  scenario, name, format, state, edge case, ordering, error behaviour or scope boundary of the
  change that the request, the ticket, a mockup and the code leave open — any point two competent
  implementers could read differently. It is asked, as `### Q<n>. <question>` under the round
  file's `## Questions`, when the operator would care about the pick: the options are different
  approaches; no option can honestly be recommended — an answer contradicting an earlier answer
  or a recorded decision included, both quoted verbatim and asked with no recommended option; a
  wrong pick is costly to undo once landed — persisted or migrated data, a public API, file
  format, CLI or cross-repository contract, a deletion, permissions, security, privacy, money;
  the pick adds or drops something the ticket or a mockup states; or the options differ by a task
  or more and the ticket gives no lead. An asked question carries a recommended option where one honestly exists and is asked until the user's
  own answer closes it (**Convergence**, below). Silence answers
  nothing: the question stays held and is asked again in the next round. Any other ambiguity
  takes its recommended option: it is recorded as a `## Decisions` entry and as one line,
  `<decision id>: <chosen> — <what it was chosen over>`, under the round file's `## Auto-taken`,
  and in the round file's `## Answers`; the parent's relay never prints, asks or mentions it, and the next
  planner and the final planner read it from the file. A round asks every question it holds — four
  is only the limit of one **AskUserQuestion** call, so more questions take more calls in the same
  round, never a later round. Each question is asked in the round it
  appears: from an answer or an agent's findings, or, once the design is
  approved, from writing the artifacts or the plan, then asked before the plan review gate and its
  answer revising the design. The self-review planner never asks:
  it takes the recommendation (**Auto-resolution**,
  `skills/flow-contracts/operator-prompts-auto-resolution.md`).

- **A reported defect is confirmed on the base before its fix is designed.** One already fixed
  there is recorded in the design as fixed, with the commit that fixed it, and planned no task.

- **A ticket that already carries the design answers the questions it settles** — a fully worked
  issue, a `/flow-plan` capture or a handoff note. Only what it leaves open is asked; the design
  approval is never taken from it. The note is input, never the plan: the plan is written through
  C and D like any other, with full repo-relative paths, and is never copied back into the note.

- **The design is approved before `spectre new` runs.** The approval is the operator's explicit
  **approve the design** answer (**Convergence**, below); nothing is created before it.

- **Every asked question explains itself.** Under its heading, before its options, a question
  carries `**Context:**` — two to five plain sentences an operator who has not read the code can
  decide from: what is being decided and where it shows up (the screen, the endpoint, the file,
  the behaviour), why it is open, and what each option concretely changes for the user or the
  code. No decision ids, task numbers or internal shorthand. Each option's own description says
  its concrete consequence in one line, never just its name. The parent prints every question's
  `**Context:**` as text right before the **AskUserQuestion** call that asks it.

- **A change too big for one plan is raised in the rounds, never after the plan is written.**
  When the design would take more than about 15 tasks or 25 mockup frames, the round asks an
  asked question naming the slices it would split into; the answer narrows this
  change's design to the first slice, and the rest is filed as follow-up issues once planning
  ends (**Follow-up issues**, `skills/flow-contracts/jira-followups.md`).

- **Now or later is never asked.** Where the options are doing work the design needs in this
  change or deferring it to a later change or ticket, now is taken and recorded under
  `## Decisions`.

- **A round file opens with `## Summary`** — at most ten lines: the decisions the operator's answers settled
  and what is still open. The Summary names no auto-taken decision — an auto-taken pick lives only
  under `## Auto-taken` and in `## Answers`. The parent relays that section and the file's absolute path, and the
  design draft stays in the file. The design draft carries `### Settled` — every answer given so
  far, the operator's words verbatim beside the question and its round — so the latest round file
  alone is the design's whole state.

- Save the design to `<project>-worktrees/<name>/.superpowers/sdd/YYYY-MM-DD-<name>-design.md` — the
  worktree kickoff created (**A. Resolve the change and write `STARTED`**, `skills/flow/brainstorm.md`). The
  path is gitignored: never stage or commit it, and
  never write it to the main checkout.
- **A mockup is read from where the project declares mockups**, the `## visual verification`
  section (**Project configuration — visual verification**,
  `skills/flow-contracts/project-configuration-visual.md`), never copied into an app repository; a
  design specified by a mockup that is not there yet is not approvable until it is.
- **A visible change is prototyped before its design is approved** (`plan-visual-prototype`). When the change can alter what a user sees — a page, a view, a chart, a dialog — and
  no mockup frame already specifies it, build a clickable prototype of the proposed result with
  sample data as a single HTML file under `<project>-worktrees/<name>/.superpowers/sdd/`, and
  iterate it on the operator's feedback each round until they approve; that approval is the convergence confirm's
  **approve the design and move on** (**Convergence**, below). The prototype is
  built on the app's own shell and styles — the real header and nav, the project's theme,
  stylesheet tokens and existing component classes, the content width and chart size of the
  sibling views — never a standalone restyle. The file draws the proposed page
  alone — no caption, note or screenshot beside it — since `flow.verify` renders the whole
  file as the capture's frame; a note to the operator goes in the message. **The operator sees the prototype
  whenever one exists**: a planner only writes the file; the session that dispatched it publishes
  it with the Artifact tool after every return that wrote or changed it and prints its URL —
  `/flow`'s parent in the relay, the self-review and in-run fix loops' parent in its final
  report. Where the harness has no Artifact tool, that session prints the file's absolute path
  and says it could not be published. Cite the approved file, with its Artifact URL
  or, where it could not be published, that it was not, from `design.md`'s `## Context` as the
  spec the implementation and `flow.verify` compare
  against. A change with nothing visible builds none. Where the confirm is not asked — the
  `## decisions: recommended` mode, and the self-review planner, which take the
  recommendation (**Auto-resolution**, `skills/flow-contracts/operator-prompts-auto-resolution.md`)
  — the published prototype is the spec without an approval round. **The approved prototype is
  the source of truth**: any departure from it found later in the run — by a reviewer,
  verify or `flow.verify` — is fixed to match it, auto-resolved and never asked; a UI
  change the operator wants after using it is a later change, not this one.

### Explore and design agents

**Every Explore and Plan agent** runs on its role's pair (**Model and effort** below) — `explorer`
for an Explore agent, `plan-designer` for a Plan agent — passed as the Agent tool's `model` and
`effort`. Its prompt carries the two agent-baseline sentences and the report-tersely line
(`~/.claude/rules/dispatch-carries-the-baseline.md`) and a narrow, explicit scope: what to look at,
what is already known, what to return. Never a `fork`, and never an agent that edits files: each is dispatched on the read-only `Explore`
or `Plan` agent type, whose tool set has no Edit and no Write. The
agents investigate and propose; you decide. They record no ledger row. Read the code yourself
first; dispatch an agent only where that cannot settle a question. On an easy change (`planner-short`, **Judge easy or hard**,
`skills/flow/brainstorm.md`) only Explore agents are dispatched — never a Plan agent.

**Findings outlive the round.** Append each Explore and Plan hand-back verbatim to
`<abs-worktree>/.superpowers/sdd/planner/findings.md` under `## <explore-<n>|design-<n>> — <the
scope it was given>` as it arrives; read that file before dispatching, and dispatch an agent only
for ground no section covers.

**The cap.** One planner dispatch launches at most five Explore and five Plan agents in round 1
and at most five agents in all in a later round or a final dispatch. A planner that needs more
writes `Over cap: <n> agents — <reason>` in the round file's `## Summary` (a final dispatch: in
its return) and dispatches them.

**Never end your turn while one of your agents runs.** A turn that ends to wait for a hand-back
loses your 5-minute prompt cache when the agent runs longer, and the hand-back then rewrites your
whole context. Once your own work is done, end your turn: the `await-own-agents` SubagentStop hook
waits for your agents itself, and returns to you when they finish, or every four minutes while one
still runs (stop again to keep waiting). Read each hand-back; say in the round file when an agent
returned nothing.

In `/flow` and `/flow-plan` the recommendation goes to the operator as part of the design, its
alternatives shown only when they help the operator decide. The self-review planner
take the recommendation without asking, per **Auto-resolution**
(`skills/flow-contracts/operator-prompts-auto-resolution.md`).

### Convergence

**A round ends on one question, asked last:** state what you believe settled — the answers and
the decisions the design now rests on — then ask **Anything unclear before I move on?** with two
options: **Approve the design and move on** *(recommended)* and **Another round — I have
something**. Only an explicit **approve** ends the rounds and approves the design; any other
answer, silence included, is another round, and the operator's words are that round's input. No
round count ends planning.

**The confirm asked alone shows the whole design** (decision `approval-by-question-preview`). A
round whose only question is the confirm is asked as one **AskUserQuestion** with exactly those
two options, the approve option's `preview` carrying the full design draft text. The session is
never switched into plan mode, and **Another round** only dispatches the next round — nothing is
stranded.

Under the `## decisions: recommended` mode (**Auto-resolution**,
`skills/flow-contracts/operator-prompts.md`) the question is not asked and **approve** is taken.

## C. Create the change and its artifacts

```bash
spectre new "<name>"   # working directory: the worktree kickoff created
```

It scaffolds `<changeRoot>` — `<project>-worktrees/<name>/spectre/changes/<name>/`, gitignored,
never staged — with stub `proposal.md`, `design.md` and `tasks.md`. Its exit 1 `already exists` is
the ordinary case on a resumed run: revise the artifacts in place, never `--force`. Keep every
stub heading and fill:

- **`proposal.md`** — `## Why` and `## What changes`.
- **`design.md`** — `## Context`; `## Decisions` and `## Open questions` (below);
  `## Recommended approach` (the chosen approach, one line per rejected alternative pointing at
  its decision ID); `## Critical files` (every file the change edits or depends on, one line on
  its role each); `## Existing code to reuse` (by path and symbol, or `none — <reason>`);
  `## Verification` (how the change is proved end to end, beyond each task's own verify step);
  `## Non-goals` and `## Risks` (each may be `none — <reason>`); and `## Live check` when the
  change touches a running service or stored state — what to run against the real thing before
  landing and what "not working" looks like, else one line opening with `None`.
- **`tasks.md`** — written in **D**.

A change that alters user-visible behaviour, a command, a flag, a file format or the pipeline
plans a task naming `<project>/README.md`; one that alters a capability's requirements plans a
task naming `<project>/spectre/specs/<capability>.md`. Neither is written here.

### Decisions

`## Decisions` records every choice the design made — small ones too: where a file lives, a
name, a task's scope — each with the alternatives and why they lost. A design that forced no
choice records none.

```markdown
### <the decision>

**ID:** <kebab-case-slug>
**Chosen:** <option> — <one-line rationale>
**Considered:** <other options, each with the tradeoff that ruled it out>
```

**ID** is mandatory, assigned once and never changed. A later round that overturns a decision appends
a new entry naming the old ID in its `**Chosen:**` line; a recorded entry is never deleted or
rewritten, so the later entry overrides it.

### Open questions

`## Open questions` is present and empty: planning leaves nothing open. An entry an earlier
design recorded there (`**ID:**`, `**Status:** open`, `**Why it is open:**`, `**What it
affects:**`) is asked before any plan is written; its answer sets `**Status:** answered by
<decision-id>`. IDs are unique across both sections.

A final planner continues straight into section **D** in the same dispatch.

## D. Basic Workflow #3 — Writing plans

Invoke the **writing-plans** skill to write `<changeRoot>/tasks.md` to plan quality — exact paths,
verification commands, bite-sized steps, no placeholders — and run its self-review before
finishing. **Tell writing-plans the task shape, because it is spectre's and not writing-plans' own.** A task
is a column-0 checkbox line, `- [ ] <n>. <title>`, whose `<n>` is a flat integer; that task's
steps are `  - [ ] **Step N: …**` lines indented two columns beneath it.

**Each task carries these fields**, each on its own column-0 line:

- `**Files:**` — the repo-relative paths this task's commit will touch, each backticked.
- `**Tests:**` — the tests this task adds, each backticked, or `none`.
- `**After:**` — `Task <id>[, <id>…]` or `none`: the tasks this one builds on.
- `**Size:**` — optional, `small`, `medium` or `big`; `big` for a task that alone fills most of an
  implementer's context.
- `**Commit:**` — the commit subject line, scope naming the module the task's `**Files:**` carry
  (**Commit scopes name the module**, `<agents repo>/rules/commit-scope-is-the-module.mdc`).

**Keep the plan plain:**

- A behaviour the design decides is pinned by a test some task names.
- A task's verify step runs its own lint and its own tests by name — never the whole suite — and
  installs nothing.
- A premise a task depends on — how a tool, a library or the harness behaves — is checked before
  the plan relies on it, and the task says what the check showed.
- A multi-repository plan sorts its tasks under `## <repository>` headings, named by the basename
  of each `## apps` repository root.
- A task cites the decision it implements as `**Decision:** <id>`, never restating it.
- A task that verifies a reported behaviour opens with "reproduce on the untouched tree before any
  edit".
- A test the plan spells out is built so the nearest wrong version of the rule fails it — its
  fixture and step order tell the rule apart from that version.
- A test that clicks, types or captures reaches each control by role, label, text or test id,
  never a coordinate, and never waits out a production timeout or backoff — the task adds a
  test-only override.
- A task that grows an existing type plans the split itself when the project's size lints would
  otherwise fail.
- A changed rule whose reasoning lives in a `*-rationale.md` appendix lists that appendix in its
  task's `**Files:**`.
- A view a mockup specifies gets a full-page capture for every frame `design.md` cites, its
  fixture seeding every element and state the frame shows.

Before returning, run `spectre validate "<name>"` and repair any exit-1 finding; its exit 2 is
reported and stops the plan.

### Plan review

**Every plan is checked** (**Plan check**, `skills/flow/brainstorm.md`); return `PLAN-WRITTEN`
alone. The self-review planner and a fix run's re-plan are never reviewed.

**A dispatch whose prompt names a `plan-check.md` report triages it.** It reads the report,
`proposal.md`, `design.md` and `tasks.md`; it opens a file under
`<abs-worktree>/.superpowers/sdd/planner/` only for a finding whose fix turns on a round's answer.
Fix in the artifacts a finding with one obvious fix. A finding whose fix would change a
`## Decisions` entry returns `QUESTIONS`, the next `round-<r>.md` quoting the finding verbatim
beside the options. Write each finding still standing into `design.md`'s `## Risks` as
`Accepted: <the finding> — <why>` — a contest of the reviewer's reading names the path:line it
rests on — re-run `spectre validate`, and return `PLAN-WRITTEN`.

### Plan check gate

**One headless pass reviews the plan and decides how it is built and reviewed**, once the final
planner returned `PLAN-WRITTEN` — the parent never decides itself. One call:

```bash
flow-guard plan-check <worktree> <changeRoot> <abs-worktree>/.superpowers/sdd/decision.json
```

Its header is canonical for the call, its checks, its fallback and its exits;
`skills/flow/plan-check-prompt.md` is canonical for the lenses, the rules and the shape. It
writes `<abs-worktree>/.superpowers/sdd/plan-check.md`, whose last line is `PLAN-CHECK-CLEAN`,
`PLAN-CHECK-FINDINGS <n>` or `PLAN-CHECK-NOT-RUN`, and the decision. Exit 2 stops the run with its
stderr. Gate **No** re-runs it as `plan-check … -decide-only`, which rewrites
`decision.json` and leaves `plan-check.md` as it is. **That re-run is the only one: it happens
once whenever `tasks.md` changed after the check — a triage fix or a gate **No** — and the decision
is recorded again.**

The parent renders the file and prints the output verbatim; the `STARTED`
handoff prints it again (**The `STARTED` handoff block**, `skills/flow/brainstorm.md`):

```bash
flow decision render -file <abs-worktree>/.superpowers/sdd/decision.json -session-model '<the model named in this session's own system prompt>'
```

It exits 0 on a printed block and 2 on a usage error or an unreadable file.

**The plan is never acted on unread.** Once the plan-check step has printed its `## Decision` block,
print the plan summary — plain prose on what will be built and how: the behaviour added or changed, the approach, the seams and decisions that matter and what was left out; no per-task rows — and, whenever `design.md` cites a prototype, the line `**Prototype:**` of **The `STARTED` handoff block** (`skills/flow/brainstorm.md`) under it. The
summary ends on two lines of their own: `**Design:** <absolute path of design.md>` and
`**Plan check:** <absolute path of the report> — clean | <n> fixed, <n> accepted | not run`.

Then the question — `/flow-plan`'s wording is **Save the plan?**, `/flow`'s is **Proceed to
implementation?** — with exactly two options: **Yes** *(recommended)* and **No (plan needs
updates)**.

**`/flow` never asks its gate question: it takes **Yes** at once.** The summary and the `## Decision` block still print, and the run ends with the
`STARTED` handoff. An operator who wants the plan changed says so before running the handoff's
`/flow <name>` — that message re-plans as **No** below. `/flow-plan`'s **Save the plan?** is
still asked.

On **No**, take the operator's changes (a follow-up **AskUserQuestion** round when the option
carried none), write them verbatim to the latest round file's `## Answers`, and dispatch a fresh final planner
on them as a `QUESTIONS` return is taken (**Run brainstorming and planning through the planner
role**, `skills/flow/brainstorm.md`) — it revises `tasks.md` and re-runs `spectre validate
"<name>"`. Then print the summary and ask again. Loop until **Yes**. **Yes** is the only exit. Under the `## decisions: recommended` mode the
gate takes **Yes** without asking — the summary and the `## Decision` block still print, so the
record of what was approved stays complete (**Auto-resolution**,
`skills/flow-contracts/operator-prompts.md`).

`/flow-plan` ends with its own report (**Capturing a new change**, `skills/flow-plan/SKILL.md`).

### Model and effort

**A dispatch's pair is picked by the pair judge**, at dispatch time: write the dispatch's prompt
to `<abs-worktree>/.superpowers/sdd/pair-judge/<key>.md`, `<key>` the dispatch's key or, for an
Explore or Plan agent, `explore-<n>` / `design-<n>`, then

```bash
flow-guard pair-judge <role> <abs-worktree>/.superpowers/sdd/pair-judge/<key>.md
```

It prints one line, `<model> <effort> <judge|fallback|single> <reason>`, and saves it beside the
prompt as `<key>.pick`, which the dispatch hook records with the row. The judge picks only among
the role's allowed pairs in flowd; no pair is ever written into a prompt or a skill. A judge
that cannot answer never blocks: it prints the role's first candidate with source `fallback`. A
role the judge does not cover (it exits 2) runs on its first `base` pair in `flow policy get`.

A pair is a model and an agent type joined by `/`: `opus/flow-low` dispatches
`subagent_type: flow-low` on `model: opus`; `opus/flow-planner-medium` is `flow-planner-medium` on
`opus`; `opus/Explore@low` is `subagent_type: Explore` with `effort: low`. Every dispatch has a
role from **Allowed pairs** (`skills/flow-contracts/model-policy.md`). A plain-language session instruction overrides the pair it names
for this run only (**Model policy**, `skills/flow-contracts/model-policy.md`).
