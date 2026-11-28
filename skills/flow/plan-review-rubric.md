# Plan review rubric

The lenses the plan reviewer (`skills/flow/plan-check-prompt.md`) applies to a change's
`proposal.md`, `design.md` and `tasks.md` before implementation. The reader is the dispatched
reviewer. Adapted from the review standards and reviewer personas of the Elephant-Goldfish Model,
<https://github.com/google/elephant-goldfish-model> (Apache-2.0) — adapted, not vendored: where
flow already owns a rule, a lens cites it and adds only the reviewer's question.

Apply the lenses in the order below. Structure and prose style are not lenses:
`spectre validate` and the prose guards own them.

## How to read

- **What you were given is all there is.** The three artifacts, this file and the repository. A
  thing the artifacts do not say is a finding against them, never a question to ask.
- **Every finding quotes its evidence**: the artifact text it is about, and the repository
  `path:line` that contradicts or duplicates it where the lens compares against code.
- **`No finding.` is a valid result for a lens.** A finding manufactured to fill a section is a
  defect of the review.
- **Critique first.** No praise, no summary of what the plan does well.
- **Severities.** `blocker` — the plan cannot be executed as written, or contradicts the code.
  `finding` — a defect with a fix. `advisory` — an idea, never owed a fix.

## Comprehension

A `##` section of `proposal.md` or `design.md` you cannot paraphrase in one sentence is a
`blocker`.

## Zoom-out

Of `proposal.md`, cost and timeline are not asked.

- Is the objective of `proposal.md` stated without jargon?
- Does `proposal.md` say how this is done today, and what limits that?
- Does `proposal.md` say what is new in this change?
- Does the change treat the root cause, or a symptom of it?
- Is each part needed now (`rules/build-the-simplest-thing.mdc`)? A part built for a need the
  artifacts do not state is a `finding`. A part that is needed but placed or built wrongly is
  **Design critic**'s.
- Is each part placed at the right layer — the guard, contract, prompt or store that already
  owns that concern?
- **Mandatory search.** For each new function, guard, file or contract section the plan adds,
  search the repository for an existing equivalent, and name the search you ran and what it
  found. An equivalent the plan does not reuse is a `finding`.

## Design critic

- A claim about why the design works — a premise, a benefit, a cost — with no evidence. A
  statement about existing code is **Codebase details**', not this lens's.
- A single point of failure the design does not name.
- A `## Decisions` entry whose `**Considered:**` names no real alternative, or no reason it lost.
- A risk the design runs that `## Risks` does not name.
- A `## Non-goals` line a task violates.
- Frames: two frames the plan cites disagree with each other (copy, label case, order) and the
  plan neither names the governing frame and why nor records the departure; picking one silently
  is a `finding`. Which frame wins where a spec is involved is decided by
  `rules/design-mockups-are-specs.mdc` and the departure bullet of
  `skills/flow-contracts/operator-prompts-auto-resolution.md`.
- A claim in `design.md`'s `## Verification` that no task runs is a `finding`. Every promised
  capture, recording or measurement names the command that produces it and where its output goes.
  A plan that starts or reuses a running stack names, before the run, the freshness check of the
  project's `.flow/project.md`.
- Which check proves success — does `design.md`'s `## Verification` name it, and does a task
  run it?

## Implementation readiness

Asked of each task in `tasks.md`: from this text alone, do I know exactly what to write, where
and how? Any choice left to the implementer is a `finding` naming the choice — the ambiguity
rule of **The checklist** (`skills/flow/brainstorm-planner.md`), tested from a fresh reader's
seat. A task that pairs, merges, restores or diffs two collections, or whose result depends on
an iteration order, names the key or the sort that decides the pairing; a positional match with
none stated is a `finding` closed by a line in `design.md`. A task that names a step by number or
index, a count, or a cross-reference matches the artifact it cites (the numbered list, the frame
list, the file it points at); a mismatch is a `finding` quoting both sides. A gap that closes only
by running it is a probe step in the task it concerns (**Plan triage**,
`skills/flow/plan-triage.md`); a task that should carry one and does not is a `finding`.

## Codebase details

Check every statement the artifacts make about existing code, a guard's behaviour or a file's
content against the repository, and check that every path and symbol they name exists. A
statement the code refutes is a `blocker`: quote the statement and the code line. A task's
`**Tests:**` file sits beside an existing test of the same kind — a full-HTTP test where the
repository's full-HTTP tests live, a unit test in the module under test — named in the plan; a test
path in a module that holds no test of that kind, and so cannot host it, is a `finding`.
A task that adds native SQL, a migration or a query has a `**Tests:**` entry that executes it
against real or seeded data; with none, nothing in the plan can verify it, and that is a `finding`.
A task that changes a symbol's signature or behaviour has a `**Files:**` entry for the places that
use it, or says why none need editing; a definition listed alone leaves its call sites to be found
while editing.
Every scenario `design.md` states has a `**Tests:**` entry that names it, or the plan says it is
checked by hand or on screen; a scenario no entry names is a `finding`, since the implementer can
only carry it as an open gap.

A UI task has a `**Tests:**` entry that captures what it builds against its frame, and builds
nothing `design.md`'s `## Undrawn` still lists; a UI element, state, icon or copy string the tasks
need that no cited frame draws and `## Undrawn` omits is a `finding`. A task whose acceptance
compares against a capture, ordered before the task that creates that capture spec, is a `finding`.

A seam a task introduces — an interface, provider, callback, composition local or hook with a
default or no-op — has a task that supplies the real implementation, or `design.md` states the
feature is inert until a named later change; a seam nothing supplies is a `finding`, since the
feature ships doing nothing.

## Inventive catalyst

Name at most three trade-offs the design accepted and, for each, a pattern that would remove the
trade-off rather than balance it. Every entry is `advisory`. With none, write the literal line
`No false compromise detected.`

## Execution confidence

Asked last, of the plan as a whole, from the seat of the implementers who are handed its tasks
with nothing but the three artifacts and the repository:

- How confident are you, 1–100%, that those implementers build this plan as written, and that the
  check `design.md`'s `## Verification` names then passes, without a pivot, a `BLOCKED` or a
  question back to the operator?
- What would move that to 100%? One gap per thing, each naming the action that closes it.

What this lens adds beyond the lenses above it: a premise about a tool, a library or the harness
that the plan relies on and no task checks; something a task needs that is not there (a fixture,
a mockup frame, a credential, a running service); a task too large for one implementer's context
that is not marked `**Size:** big`; an ordering the `**After:**` fields do not express; a
verification step that could pass with the behaviour broken. A gap an earlier lens already raised
is cited by its `F<k>`, never raised a second time.

The section's shape is fixed, because `flow-guard plan-check` reads it:

```text
### Execution confidence

CONFIDENCE: 78%
Lowered by: F3, F7
**F9** [finding] tasks.md:41 — closes with: plan — Task 4 relies on a flag no task checks — stats/cmd/flow/record.go:112 parses it as a positive integer
**F10** [finding] design.md:63 — closes with: operator — the export's column order is not stated and the ticket shows two — proposal.md:14 against the ticket's sample
**F11** [advisory] tasks.md:58 — closes with: running it — whether the deadline holds on a cold cache is not knowable from the text — no measurement in design.md — probe: run the deadline test on a cold cache → under the deadline; on failure: add a warm-up task
```

- `CONFIDENCE: <n>%` — one line, `<n>` an integer 1–100.
- `Lowered by: F<k>[, F<k>…]` — the earlier lenses' findings the number already accounts for;
  omitted when none.
- One finding line per new gap, in the report's own shape and numbering, its "what is wrong"
  opening `closes with: plan|operator|running it —`: what closes the gap is an edit to the plan,
  an answer only the operator has, or nothing short of running the work.
- A gap that closes with `plan` or `operator` is a `finding`; one that closes only with
  `running it` is `advisory`. The lens raises no `blocker`.
- The Confidence page shows these lines without the report around them, so each gap line reads on its
  own: the task or file it concerns and what is unknown, in words, never a bare `F<k>` or "see
  above". `flow-guard plan-check` spells each `Lowered by:` finding out from the report's own line.
- A gap that closes only by running it names, in its own line, the cheapest probe that settles it,
  the expected result and the fallback on failure, in the shape `probe: <command or step> →
  <expected>; on failure: <fallback>`. A running-it gap with no probe or no fallback is raised as a
  `finding` closing with `plan` ("name how it will be found out"), never as `advisory`. An
  `advisory` gap that carries a probe and a fallback is listed but does not lower the CONFIDENCE
  number: the number counts only gaps an edit to the plan or an answer from the operator can close,
  plus unprobed running-it gaps.
- `CONFIDENCE: 100%` carries no gap line and no `Lowered by:`; below 100%, at least one of the
  two is present.
