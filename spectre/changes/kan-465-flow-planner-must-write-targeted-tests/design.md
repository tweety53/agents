# kan-465-flow-planner-must-write-targeted-tests — design

## Context

The evidence that the planner, not the implementer, is now the source of whole-module test runs is
in `proposal.md` `## Why` and is not restated here.

What matters for the edit is the shape of the site. `skills/flow/brainstorm-planner.md` section
**D** ("Basic Workflow #3 — Writing plans") runs in four beats: the `superpowers:writing-plans`
invocation paragraph; the task-shape paragraph that tells the skill spectre's own checkbox-line
shape and cites the `Placement` paragraph of `skills/flow-contracts/build-green.md`; the
`skills/flow-contracts/plan-provenance.md` load; and the `skills/flow-contracts/build-green.md`
load with the `flow-task-commit-fields` list. Every one of those beats describes what a task
*declares* — its files, its tests, its baseline, its commit subject. None describes what a task's
verify step *runs*, which is why a planner writing `./gradlew :shared:desktopTest` into a step is
not contradicted by anything it has read.

`skills/flow/implement.md` §4 already holds the corresponding instruction for the implementer, as
the `**TARGETED TESTS:**` paragraph — including the selector vocabulary (`--tests '<class>'` for
Gradle, `-run '<name>'` for `go test`, `-t '<name>'` for vitest) and the statement that the full
`## test` list belongs to the last bundle and to `flow.verify`. The new paragraph mirrors that
vocabulary rather than inventing a second one, so the planner's instruction and the implementer's
read as one rule.

## Change

One blockquote inserted in `skills/flow/brainstorm-planner.md` section **D**, immediately after the
task-shape paragraph (the one ending at the `Placement` citation) and before the
`**Load skills/flow-contracts/plan-provenance.md.**` paragraph:

```markdown unverified:to be written on the change branch
> **Write each task's verify step as its own lint commands plus targeted tests.** A verify step
> names the lint commands the task's own `**Files:**` actually need — never the project's whole
> `## lint` list — and the build tool's own selector for each `**Tests:**` entry (`--tests
> '<class>'` for Gradle, `-run '<name>'` for `go test`, `-t '<name>'` for vitest), never the bare
> module or repository suite: that run belongs to the last bundle's FULL SUITE paragraph
> (`skills/flow/implement.md`) and to `flow.verify`. A task whose `**Tests:**` is `none` names lint
> alone and no test command.
```

Three clauses, in one paragraph: targeted selectors per `**Tests:**` entry; lint bounded to the
task's own `**Files:**`; lint alone for a `**Tests:** none` task.

The placement is load-bearing. Immediately after the task-shape paragraph, the planner has just
been told what a task and its steps look like and is told next what a step may contain — the two
sentences describe the same object. Placed instead inside the `flow-task-commit-fields` list, it
would read as a declared field like `**Tests:**` or `**Build:**`, which it is not: it constrains
prose in a step, not a mechanically-checked field, and `check-task-commit-fields.sh` does not and
will not read it.

## Verification

The change is prose in a file no test executes, so its verification is this repository's own guard
set, all of which run against a bare tree:

- `scripts/check-references.sh` — resolves the new `skills/flow/implement.md` citation.
- `scripts/check-vocabulary.sh` — the new text introduces no retired vocabulary.
- `scripts/check-markdown-integrity.py` — the inserted blockquote's structure.
- `scripts/check-contract-budget.sh` — the file is 19188 bytes against a declared budget of 23248
  <!-- measured: wc -c skills/flow/brainstorm-planner.md and the budgets() row in scripts/check-contract-budget.sh @ main 3865c4b -->
  so a ~600-byte blockquote passes with no budget edit.
- `scripts/check-normative-inventory.sh` — captured before the edit and again after, expected
  identical. The paragraph is worded "names / never" rather than `MUST` or `SHALL` precisely so
  this holds; a delta here means the wording drifted into a normative keyword and should be
  reworded rather than accepted.

**No new test, and none available.** There is no assertion that a future planner obeys the
paragraph — the same limit `scripts/check-dispatch-paragraphs.sh` states about the paragraphs it
checks. A presence-and-phrases row in that guard's own table was considered and rejected here: its
table is scoped to *dispatch* paragraphs at dispatch sites, this paragraph is ordinary instruction
prose in a skill file, and adding a row would be the guard `prompt-only-no-guard` already declined,
wearing a different guard's name. The task therefore declares `**Tests:** none` and rests on the
guard run above.

## Decisions

### Fix the planner's prompt rather than add a plan-shape guard

**ID:** prompt-only-no-guard
**Status:** active
**Chosen:** One paragraph in `skills/flow/brainstorm-planner.md` and no new rule in
`scripts/check-plan-shape.sh` — the post-KAN-441 measurement shows the implementer follows whatever
the plan says, so the plan text is the entire causal chain and correcting it is the entire fix.
**Considered:** A `check-plan-shape.sh` rule rejecting a bare module/repo suite in a step line —
offered during research and declined there. It would need a per-project vocabulary of "bare suite"
commands (`./gradlew test`, `go test ./...`, `npm test`, `pytest`) that the guard cannot derive
from `.flow/project.md`, and it enforces a habit no evidence yet shows surviving the prompt fix.
Deferred to the first plan found violating the paragraph.

### All three clauses in one paragraph, not a deferred follow-up

**ID:** three-clauses-one-paragraph
**Status:** active
**Chosen:** The targeted-selector clause, the `**Files:**`-bounded lint clause and the
`**Tests:** none` clause land together. All three describe the same object — the content of one
verify step — and a planner reading two of the three still has to guess at the third.
**Considered:** The named gap alone, with the lint bound and the `none` case left as the research
note's two "related gaps" — rejected by the operator: the `none` case is the one that produced the
2–4 whole-module runs in kan-454 and kan-455, the two changes whose implementers were otherwise
fully compliant, so deferring it leaves the measured residue in place.

### Placement after the task-shape paragraph

**ID:** placement-after-task-shape
**Status:** active
**Chosen:** Immediately after the task-shape paragraph, before the plan-provenance load — adjacent
to the only other paragraph in **D** describing what a step looks like.
**Considered:** Inside the `flow-task-commit-fields` bullet list — rejected: that list is the
mechanically-checked field family `check-task-commit-fields.sh` reads, and a paragraph about step
prose placed among them invites both a reader and a future guard author to treat it as a field.
Also considered appending it to the `superpowers:writing-plans` invocation paragraph — rejected as
burying a specific constraint inside the sentence that hands control to another skill.

### Mirror `implement.md`'s selector vocabulary rather than restate its rule

**ID:** mirror-implement-selectors
**Status:** active
**Chosen:** Reuse the three selectors `skills/flow/implement.md` §4 already names and cite that
file for the FULL SUITE half, rather than repeating why the suite belongs to the last bundle.
**Considered:** A self-contained paragraph restating the rationale — rejected under this
repository's cut-never-paraphrase convention: two copies of a normative sentence drift, and the
implementer's copy is the canonical one.

## Open questions

*(none — the `**Tests:** none` case, deferred in the issue text, was pulled into scope by the
operator and is decision `three-clauses-one-paragraph` above; the issue's other noted gap,
unbounded lint runs, is the paragraph's second clause.)*
