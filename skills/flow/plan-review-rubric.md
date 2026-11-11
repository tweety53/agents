# Plan review rubric

The lenses the plan reviewer (`skills/flow/plan-reviewer-prompt.md`) applies to a change's
`proposal.md`, `design.md` and `tasks.md` before implementation. The reader is the dispatched
reviewer. Adapted from the review standards and reviewer personas of the Elephant-Goldfish Model,
<https://github.com/google/elephant-goldfish-model> (Apache-2.0) — adapted, not vendored: where
flow already owns a rule, a lens cites it and adds only the reviewer's question.

Apply the lenses in the order below. Structure and prose style are not lenses:
`flow-guard check-plan-shape` and the prose guards own them.

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

Paraphrase each `##` section of `proposal.md` and `design.md` in one sentence, and keep every
paraphrase in the report so the planner can spot a wrong one. A section you cannot paraphrase is
a `blocker`.

## Zoom-out

- Does the change treat the root cause, or a symptom of it?
- Is each part needed now (`rules/build-the-simplest-thing.mdc`)? A part built for a need the
  artifacts do not state is a `finding`.
- Is each part placed at the right layer — the guard, contract, prompt or store that already
  owns that concern?
- **Mandatory search.** For each new function, guard, file or contract section the plan adds,
  search the repository for an existing equivalent, and name the search you ran and what it
  found. An equivalent the plan does not reuse is a `finding`.

## Design critic

- A claim with no evidence in the artifacts or the code.
- A single point of failure the design does not name.
- A `## Decisions` entry whose `**Considered:**` names no real alternative, or no reason it lost.
- A risk the design runs that `## Risks` does not name.
- A `## Non-goals` line a task violates.

## Proposal exam

Four questions, asked of `proposal.md`. Cost and timeline are not asked.

- Is the objective stated without jargon?
- Does it say how this is done today, and what limits that?
- Does it say what is new in this change?
- Which check proves success — does `design.md`'s `## Verification` name it, and does a task
  run it?

## Implementation readiness

A `tasks.md` that is still the title-only scaffold — the review before writing-plans — skips this
lens.

Asked of each task in `tasks.md`: from this text alone, do I know exactly what to write, where
and how? Any choice left to the implementer is a `finding` naming the choice — the ambiguity
rule of **The checklist** (`skills/flow/brainstorm-planner.md`), tested from a fresh reader's
seat. File boundaries and build state are guard-enforced and not re-checked.

## Codebase details

Check every statement the artifacts make about existing code, a guard's behaviour or a file's
content against the repository, and check that every path and symbol they name exists. A
statement the code refutes is a `blocker`: quote the statement and the code line.

## Inventive catalyst

Name at most three trade-offs the design accepted and, for each, a pattern that would remove the
trade-off rather than balance it. Every entry is `advisory`. With none, write the literal line
`No false compromise detected.`
