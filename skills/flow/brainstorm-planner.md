# Brainstorm and plan — sections B, C, D

Sections **B**, **C** and **D** of the brainstorming phase, run by whichever session is executing
`/flow`'s brainstorming stage — or, for **C**, **D** and **Decide**, a `/flow-plan` capture
(**Capturing a new change**, `skills/flow-plan/SKILL.md`) — no one relays these sections, though
the session may dispatch the fresh subagents **Explore and design agents** below allows, and
neither does the self-review planner, which runs B, C and D as item 1 of **3. Fix every finding that is
not `big`** (`skills/flow-self-review/SKILL.md`) binds them, and the in-run fix planner, which
runs them as **Pipeline defects found mid-run** (`skills/flow-contracts/pipeline.md`) binds them. Every "you" below addresses that session, or that planner, directly.

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

### Explore and design agents

**Every planning dispatch** takes its pair from `plan-agent-pair.sh <seed> <n>` — `<seed>` the
change name (in `/flow-plan` before capture, the session token; for the self-review and in-run fix
planners, their own `Key:`), `<n>` counting the session's planning dispatches from 1 — passing the
printed model as the Agent tool's `model` and the effort as its `effort`. Its prompt carries the
two agent-baseline sentences and the report-tersely line (`~/.claude/rules/dispatch-carries-the-baseline.md`)
and a narrow, explicit scope: what to look at, what is already known, what to return. Never a
`fork`, and never an agent that edits files. The agents investigate and propose; you decide.

In `/flow` and `/flow-plan` the recommendation goes to the operator as part of the design, its
alternatives shown only when they help the operator decide. The self-review and in-run fix planners
take the recommendation without asking, per **Auto-resolution**
(`skills/flow-contracts/operator-prompts-auto-resolution.md`).

### Convergence

Under the `## decisions: recommended` mode (**Auto-resolution**,
`skills/flow-contracts/operator-prompts.md`) the confirm is not asked: the recommended
**approve the design and move on** is taken, and the silence default above never fires.

Under the mode the offer is not asked either: **Yes — run another round** is taken, and the
silence default above never fires (**Auto-resolution**,
`skills/flow-contracts/operator-prompts.md`).

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
> round, applied at plan time (kan-1007: a plan-time capture was judged by eye).

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
- **the extra dispatches** — `test-sanity+call-efficiency` and `mutation+failure-modes`, each its own rolled pair, `sonnet` or `opus` at `low`: the only pass-1 panel dispatches that may run on `sonnet` (decision `extra-dispatch-sonnet-carve-out`, kan-959; kan-1028);
- **every gated per-task reviewer bundle** — `opus`, always, at the panel roll's effort, never `high`, since no decision pair records it; it is a first-pass review too (`skills/flow/implement.md`);
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
the pair carries a hard seam.** **Every pass-1 panel dispatch and every gated per-task reviewer
bundle takes `low` when `effort_roll < 60` and `medium` otherwise, whatever the band** — 60/40 per
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
`{verify, reason}`, `verify` one of `required`, `skipped` or `not configured`), `extra_dispatch` and `extra_dispatch_mutation` (each an object `{slots, model, effort, reason}`, recorded on every class), `parent` (the parent's own
model/effort, per **Harness mapping** (`skills/flow-contracts/model-policy.md`): `unknown
(agent-defined)` and `default` where the session does not state them), `overrides` (session-instruction
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
print the plan summary (**Plan review gate**, `skills/spectre-planning/SKILL.md`), then ask the gate question in one **AskUserQuestion** call. The
gate is mandatory: no answer, no commit and no implementation. It runs at the end of every
`/flow-plan` capture (**Capturing a new change**, `skills/flow-plan/SKILL.md`) and in `/flow`
after the `flow.decide` record sequence (**Run brainstorming and planning directly**,
`skills/flow/brainstorm.md`).

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
