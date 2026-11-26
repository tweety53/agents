# Brainstorm and plan — sections B, C, D

Sections **B**, **C** and **D** of the brainstorming phase, run by `/flow`'s `planner` dispatch
(**Run brainstorming and planning through the planner role**, `skills/flow/brainstorm.md`), in
`/flow`'s creating run and in a `/flow-plan` capture alike — no one relays these sections, though
the session may dispatch the fresh subagents **Explore and design agents** below allows. Every "you" below addresses that session, or that planner, directly.

In a `planner` dispatch, every **AskUserQuestion** these sections name is written as a question of
the round file instead, and the parent asks it, with the silence outcome the naming section
states.

## B. Basic Workflow #1 — Brainstorming

Investigate read-only, then present the design.

### The checklist

- **Every ambiguity is either asked or auto-taken — asked only in `/flow`'s planning up to its
  plan gate and in a `/flow-plan` session, and nowhere else.** An ambiguity is any requirement,
  scenario, name, format, state, edge case, ordering — the key or sort that pairs two collections by position included — error behaviour or scope boundary of the
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
  answer revising the design.

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
  change's design to the first slice, and each other slice is filed once planning ends as
  its own Story, its description opening `Split from <KEY>` and moved to **Slice queue** right after
  creation (**Transitions**, `skills/flow-contracts/jira-integration.md`).

- **Now or later is never asked.** Where the options are doing work the design needs in this
  change or deferring it to a later change or ticket, now is taken and recorded under
  `## Decisions`. The one exception is the plan check's confidence question (**Plan review**,
  below).

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
- **Shared constants are listed before any frame is approved.** For a change that carries UI, read
  the constants its frames share — bar heights, the title's y, rail colours, row heights, greys,
  label wording — from the app's own token file, and list them in `design.md` under
  `## Shared constants`, one per item: ``- `<name>`: <value> — <token and file it was read from>``.
  A frame that disagrees with the list is corrected in the frame, or the list is changed by an
  explicit decision recorded in `design.md` — never left for `flow.verify` to find. Each frame of the change states every listed
  constant — the value it draws, or `n/a` where it does not draw it — in
  a sidecar `<frame id>.constants` beside it, one `<name>: <value>` per line, or in
  `data-const-<name>="<value>"` attributes in an HTML frame. A sidecar's value is read off that
  frame — measured on its PNG with `flow-guard measure-visual-properties`, or read from the design
  source at the frame's tag — never copied from the list, which is what it is checked against.
  `flow-guard check-frame-constants <design.md> <frames dir or frame file>...`, given the change's
  frames, exits 0 before the design is approved; it exits 1 naming each frame that leaves a listed
  constant unstated and each constant two frames draw differently. A rendered frame's sidecar sits
  in its tag's directory, which is never committed: the check is this gate's, re-stated if the
  frames are rendered afresh. Whoever adds or changes a listed constant runs
  `flow-guard check-frame-constants` in the same step and fixes every frame it names — each
  `FRAME-CONSTANT-ADD:` line it prints is the sidecar line to add, its value read off the frame. Every colour a cited frame uses is also mapped to a theme token, or a
  new one the design adds, before the design is approved — never a hex at the use site.
- **A computed rule is run against every cited frame.** Where a screen's rows, titles or copy come
  from a computed rule, the planner runs it for each cited frame and compares; every mismatch is
  decided before the design is approved.
- **Every visible element a task builds is drawn before the design is approved, and the
  `designer` draws what the mockups leave undrawn.** For a change that carries UI, list in
  `design.md` under `## Undrawn`, from the tasks as written, each element, state, icon, copy string,
  alignment and chrome variant (title placement, close control) that no cited frame draws, and each
  icon the project's icon set lacks — never inferred from a design template, a sibling screen or
  the nearest component. `flow-guard design-repo <abs-worktree>` — never
  `flow-guard project-get`, never the main checkout — says who draws them. Exit 0 prints
  `<clone> <remote>`: the planner leaves the items listed and returns, and **whenever a planner
  return leaves `## Undrawn` non-empty (the round file's design draft, or `design.md`), the session
  that dispatched it dispatches one `designer`** (**The designer**, below) for the whole list in
  the same turn, unasked; the next planner dispatch replaces each item with the frame id the
  designer returned. Exit 1, no design repo: the planner draws every item into the prototype HTML
  (`plan-visual-prototype`, below), cites it and empties the list. Exit 2 is a defect fixed in the
  run. **The answer to an undrawn item is always draw**: it is never put to the operator as a
  question, and neither ruling it out nor reusing something as is is ever offered. With a design
  repo the only operator step for the items is **The designer review**, below, which ends before
  the convergence confirm (**Convergence**, below) is asked. The design is not approvable while
  `## Undrawn` holds an item. An implementer meeting an element the plan missed reports it in its
  REPORT FILE and builds the nearest drawn pattern.
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
  `/flow`'s parent in the relay. Where the harness has no Artifact tool, that session prints the file's absolute path
  and says it could not be published. Cite the approved file, with its Artifact URL
  or, where it could not be published, that it was not, from `design.md`'s `## Context` as the
  spec the implementation and `flow.verify` compare
  against. A change with nothing visible builds none. Where the confirm is not asked — the
  `## decisions: recommended` mode, which takes the recommendation (**Auto-resolution**, `skills/flow-contracts/operator-prompts-auto-resolution.md`)
  — the published prototype is the spec without an approval round. **The approved prototype is
  the source of truth**: any departure from it found later in the run — by a reviewer,
  verify or `flow.verify` — is fixed to match it, auto-resolved and never asked, as **A departure
from the mockup** (`skills/flow-contracts/operator-prompts-auto-resolution.md`) resolves it; a UI
  change the operator wants after using it is a later change, not this one.

### Explore and design agents

**Every Explore and Plan agent** runs on its role's pair (**Model and effort** below) — `explorer`
for an Explore agent, `plan-designer` for a Plan agent — passed as the Agent tool's `model` and
`effort`. Its prompt carries the two agent-baseline sentences and the report-tersely line
(`~/.claude/rules/dispatch-carries-the-baseline.md`) and a narrow, explicit scope: what to look at,
what is already known, what to return, and the line `Role: explorer` or `Role: plan-designer`. Never a `fork`, and never an agent that edits files: each is dispatched on the read-only `Explore`
or `Plan` agent type, whose tool set has no Edit and no Write. The
agents investigate and propose; you decide. They record no ledger row. Read the code yourself
first; dispatch an agent only where that cannot settle a question.

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
alternatives shown only when they help the operator decide.

### The designer

**One `designer` dispatch draws every `## Undrawn` item** (**Every visible element a task builds is
drawn**, above), key `designer-<k>`, `<k>` counting from 1 per change, on the pair the pair judge picks for it
(**Model and effort** below, the prompt saved as `pair-judge/designer-<k>.md` first — a small
adjustment round is judged on its own prompt, never given the role's first pair by default), on a
`flow-*` agent type, since it writes. Its prompt carries the two
agent-baseline sentences, the report-tersely line, `Role: designer`, the `## Undrawn` list
verbatim, the `<clone> <remote>` `flow-guard design-repo` printed, the frames already cited, and on an adjustment the operator's words verbatim. The
designer:

1. works in a git worktree of that design-repo clone, `<design clone>-worktrees/<change
   name>` on branch `<change name>` — created from the clone's default branch when absent, reused
   as it stands when present;
2. reads that repository's own instructions first — its `CLAUDE.md`, `AGENTS.md`, README, style
   guide and tokens — and follows them exactly, inventing no style they do not state;
3. runs that repository's own stack where its instructions name one (a local server, a build, a
   render tool — gymie-design's is a local node server and Playwright), from its worktree, on a
   port of its own and never touching a server the operator runs; where they name no server, it
   serves the worktree with a plain static one. The server the review page (item 4) is served on
   stays running until the operator approves (**The designer review**, below);
4. draws each listed item as frames in that repository's own frame format, its copy in the voice
   those instructions set, and where they give no answer the simplest plausible copy, listed in
   its report — it never asks. Every frame it draws carries the description that repository's frames
   carry, in the same form (gymie-design's: id, title, a behaviour caption, a LINKED line) — never
   fewer than the frames beside it have. **Every drawn frame is part of the repository's current
   pages** — placed on the existing page and section its instructions say such a frame belongs to,
   never an orphan file — **and every new thing is also collected on one separate review page** for
   the change, in the repository's own page format, named after the change: each new frame with its
   description and a link to where it lives in the current pages, so the operator opens one page and
   searches for nothing. That review page is the only page of the review: every drawn or changed
   frame is on it, rendered in full, with its id, title, behaviour caption and LINKED line, and the
   operator needs no other URL, page or published copy. Where the stack renders, it renders every drawn frame,
   reads each image and fixes what the instructions' own checks would flag. Before it reports it
   opens the review page and checks that every `#<frame id>` anchor resolves and every frame
   renders;
5. commits on the branch, tags that commit `<change name>-designer-<k>` (`<k>` the dispatch's) and
   pushes the branch and the tag — `flow-guard render-frames` resolves a frame at a tag only
   (`skills/flow-contracts/project-configuration-visual.md`, **Frames at a design-repo tag**) — and
   reports the review page's URL, once with a `#<frame id>` anchor per drawn or changed frame, each
   with the frame's title — the only URL in the report: no link to another page, no PNG, no
   Artifact copy; a report without it is incomplete and the planner sends the designer back for
   it. It also reports each frame id and path drawn, on an adjustment the frame ids changed, the
   copy it wrote without a rule, the tag and the commit sha. The planner cites each frame as
   `<mockups root>/<tag>/<frame id>.png` with that tag; an adjustment's dispatch cites its own,
   newer tag.

The dispatching session appends that report to the latest round file's `## Answers`, where the next planner dispatch reads it. The designer is outside **The cap** above.

**The designer review** is an interactive pre-step of the approval, run by the session that
dispatched the planner after every return that drew or changed frames. That session prints the
review page's anchored URLs, every one in full, first — the operator reviews on that live page, and
is given no other URL, no PNG and no Artifact copy — then, after an adjustment, the frame ids that
changed, then the design-repo branch and commit.
It then asks through **AskUserQuestion**: **Continue with the planning** *(recommended)* or the
operator's own words on what to change. Words naming a change are one more `designer` dispatch on
the same worktree, carrying them verbatim; it applies them on the review page, keeps its server
running and reports which frame ids changed, then the review runs
again on its return; **Continue with the planning** ends the review, the session stops the server the
designer left running, **lands the design branch** and planning goes on. No convergence confirm is
asked while a review is open, and the confirm never carries the drawn frames.

**Landing the design branch is the approval's own step, unasked.** From the designer's worktree the
session pushes `HEAD:<default branch>` of the `design repo` clone — a fast-forward, the default
branch being what `git symbolic-ref --short refs/remotes/origin/HEAD` names with its `origin/` prefix
dropped — after rebasing the branch onto a fresher `origin/<default branch>` where the push is
refused as non-fast-forward, then runs `git -C <design clone> pull --ff-only`. The branch and its
`<change name>-designer-<k>` tags stay pushed, so `flow-guard render-frames` still resolves each
frame at its tag. A rebase conflict or any other refusal is a defect fixed in the run, never a stop.
The unapproved branch is never landed: a review that never reaches **Continue with the planning**
lands nothing. Where the review is not asked (below) the drawn frames stand and the same landing
runs once the plan is written.

Where the confirm is not asked (`## decisions: recommended`) the review is not asked either: the drawn frames stand.

### Convergence

**The confirm is never asked while `## Undrawn` holds an item or a designer review is open**: the
session that would ask it refuses, strikes it from the round's questions and goes to **The
designer**, above, instead.

**A round ends on one question, asked last:** state what you believe settled — the answers and
the decisions the design now rests on — then ask **Anything unclear before I move on?** with two
options: **Approve the design and move on** *(recommended)* and **Another round — I have
something**. Only an explicit **approve** ends the rounds and approves the design; any other
answer, silence included, is another round, and the operator's words are that round's input. No
round count ends planning.

**The confirm asked alone shows the whole design**. A
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
plans a task naming `<project>/README.md`. It is not written here.

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
- A test's expected value follows from the rule the plan's implement step states: the planner walks
  the fixture through that rule, and where the expected value and the rule disagree, it fixes
  whichever is wrong before returning the plan.
- A test that clicks, types or captures reaches each control by role, label, text or test id,
  never a coordinate, and never waits out a production timeout or backoff — the task adds a
  test-only override.
- A task that grows an existing type plans the split itself when the project's size lints would
  otherwise fail.
- A changed rule whose reasoning lives in a `*-rationale.md` appendix lists that appendix in its
  task's `**Files:**`.
- A view a mockup specifies gets a full-page capture for every frame `design.md` cites, its
  fixture seeding every element and state the frame shows. For each cited frame the capture task
  names how its fixture reaches that state and which task makes that possible; a state nothing can
  reach is decided at design time: add a seed route, redraw the frame, or accept the departure. A task that builds a drawn element on a
  native screen has a `**Tests:**` entry that compiles it and captures it against its frame; a
  capture no task creates leaves every implementer holding the same "unverified until run" gap.

Before returning, run `spectre validate "<name>"` and repair any exit-1 finding; its exit 2 is
reported and stops the plan.

### Plan review

**Every plan is checked** (**Plan check**, `skills/flow/brainstorm.md`); return `PLAN-WRITTEN`
alone. A fix run's re-plan is never reviewed.

**A dispatch whose prompt names a `plan-check.md` report triages it.** It reads the report,
`proposal.md`, `design.md` and `tasks.md`; it opens a file under
`<abs-worktree>/.superpowers/sdd/planner/` only for a finding whose fix turns on a round's answer.
Fix in the artifacts a finding with one obvious fix. A finding whose fix would change a
`## Decisions` entry returns `QUESTIONS`, the next `round-<r>.md` quoting the finding verbatim
beside the options. Write each finding still standing into `design.md`'s `## Risks` as
`Accepted: <the finding> — <why>` — a contest of the reviewer's reading names the path:line it
rests on — re-run `spectre validate`, and return `PLAN-WRITTEN`.

**A confidence below 100% is put to the operator, before anything is triaged.** When
`<abs-worktree>/.superpowers/sdd/confidence.json` exists and its `score` is below 100, the
triaging dispatch's first return is `QUESTIONS`, its round file holding one question:
`### Q<n>. The plan check is <score>% confident this plan is executed as written — raise it now?`
Its `**Context:**` lists every entry of the file's `gaps` verbatim, each followed by the
suggestion that closes it — the edit to the plan, the question the operator would be asked, or
what only running the work can show. It has exactly two options, **Raise it** — apply every
suggestion now — and **Later** *(recommended)* — file the estimate and go on. Under the
`## decisions: recommended` mode (**Auto-resolution**,
`skills/flow-contracts/operator-prompts-auto-resolution.md`) the question is not asked and
**Later** is taken. A gap that closes with `operator` is not asked in this round: it is asked
only once the answer is **Raise it**.

- **Raise it** — the fresh final planner that takes the answer triages the report as above, with
  every gap in scope: a `closes with: plan` gap is fixed in the artifacts, a `closes with:
  operator` gap is one more `QUESTIONS` round, a `closes with: running it` gap is written into
  `## Risks` as `Accepted: …`. It then files the estimate with `-disposition raised`.
- **Later** — it changes nothing for the gaps: each is written into `## Risks` as
  `Accepted: <the gap> — deferred at the confidence question`, the report's other
  findings are triaged as above, and the estimate is filed with `-disposition later`.

**Filing is list, then reuse or create.** Run `flow confidence tags`; for each kind of gap, reuse
a printed tag that fits it, and write a new lowercase kebab-case tag only when none does — a tag
names the kind of gap (`unchecked-premise`, `missing-fixture`), never the change. Then, once:

```bash
flow confidence record -change <name> -stage plan -score <score> -disposition <raised|later> \
  -gap '<one entry of gaps, verbatim>' [-gap …] -tag <tag> [-tag …]
```

`<score>` is `confidence.json`'s, the cold reader's number: it is never re-estimated after the
gaps are closed. Exit 2 is a refused estimate — correct the flag its stderr names and re-run; an
unreachable store warns and exits 0, and the planner goes on.

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
Besides the prompt, the judge reads only the file parts it names: `` `<abs-path>#<heading text>` ``
(in backticks; that markdown section) or `<abs-path>:<a>-<b>` / `:<a>` (those lines); a bare
`tasks.md` gives the sections of the tasks the prompt names, and any other bare path nothing.

A pair is a model and an agent type joined by `/`: `opus/flow-low` dispatches
`subagent_type: flow-low` on `model: opus`; `opus/flow-planner-medium` is `flow-planner-medium` on
`opus`; `opus/Explore@low` is `subagent_type: Explore` with `effort: low`. Every dispatch has a
role from **Allowed pairs** (`skills/flow-contracts/model-policy.md`). A plain-language session instruction overrides the pair it names
for this run only (**Model policy**, `skills/flow-contracts/model-policy.md`).
