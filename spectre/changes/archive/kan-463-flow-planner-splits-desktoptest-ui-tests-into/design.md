# kan-463-flow-planner-splits-desktoptest-ui-tests-into — design

## Context

The evidence that the fused "feature plus `desktopTest` file" task shape is the tail cost is in
`proposal.md` `## Why` and is not restated here.

The site is `skills/flow/brainstorm-planner.md` section **D** ("Basic Workflow #3 — Writing
plans"). After the `superpowers:writing-plans` invocation and the task-shape paragraph, KAN-465
placed a blockquote telling the planner what a task's verify step runs. That blockquote is the only
paragraph in **D** that constrains what a task *contains* rather than what it *declares*; the new
paragraph constrains the same object — which files one task carries — and sits directly after it.

Everything downstream already does the right thing for two file-disjoint tasks and needs no edit:

- `scripts/plan-dispatch-bundles.py` joins two unchecked tasks only on a shared `**Files:**` path
  or a `**Squash-with:**` field, so a UI-test task naming only its UI-test file(s) is its own
  bundle and gets its own implementer dispatch.
- A task carrying no `**After:**` field runs after every earlier task (the serial default the
  same script resolves), so the follow-on runs after its feature task with no field at all.
- `skills/flow/implement.md` §4's FULL SUITE paragraph runs the resolved `## test` list once at the
  last bundle regardless of how the plan was split, and `flow.verify` runs it again.
- The follow-on task is `**Build:** green`: the feature has landed by the time it runs, and its UI
  tests go RED→GREEN inside the task exactly as any other task's tests do.

## Change

One blockquote inserted in `skills/flow/brainstorm-planner.md` section **D**, immediately after the
verify-step blockquote (the one beginning "**Write each task's verify step as its own lint commands
plus targeted tests.**") and before the `**Load skills/flow-contracts/plan-provenance.md.**`
paragraph, separated from each by one blank line:

```markdown verified:trial-inserted on the change branch at fb8f38a; check-references, check-vocabulary, check-markdown-integrity, check-contract-budget all exit 0, normative inventory unchanged; then reverted
> **Write a feature's UI tests as their own follow-on task.** When a feature's tests live in a
> UI-test source set of their own — Compose Multiplatform's `desktopTest`, and the like — the plan
> carries two tasks: the feature task, whose `**Files:**` names the source files and the unit-test
> (`commonTest`) files, then one follow-on task per feature task whose `**Files:**` names only that
> feature's UI-test file(s), all of them. The two are file-disjoint, so `plan-dispatch-bundles.sh`
> keeps them separate bundles and the UI-test iteration starts in a fresh implementer at a small
> context instead of the one that just wrote the feature. No `**After:**` field is needed — the
> serial default already runs the follow-on after its feature task — and the last bundle's FULL
> SUITE run still covers the pair.
```

Four clauses in one paragraph: the trigger (tests in a UI-test source set of their own, `desktopTest`
as the example); the split (feature task with source and `commonTest`, one follow-on per feature
task carrying all of its UI-test files); why it works (file-disjoint, separate bundles, fresh
context); and what does not change (`**After:**` unnecessary, FULL SUITE kept).

## Verification

The change is prose in a file no test executes, so its verification is this repository's own guard
set, run against the edited tree:

- `scripts/check-references.sh` — the paragraph's backticked `plan-dispatch-bundles.sh` resolves,
  as the existing citation of it in the same file already does.
- `scripts/check-vocabulary.sh` — no retired vocabulary.
- `scripts/check-markdown-integrity.py` — the blockquote's structure.
- `scripts/check-contract-budget.sh` — the file is 28392 bytes against a declared budget of 31664;
  the blockquote takes it to 29217.
  <!-- measured: wc -c skills/flow/brainstorm-planner.md before and after the trial insert, and the budgets() row in scripts/check-contract-budget.sh @ branch spectre/kan-463-flow-planner-splits-desktoptest-ui-tests-into -->
- `scripts/check-normative-inventory.sh` — captured before and after, identical. The paragraph is
  worded "carries / names / is needed" rather than `MUST` or `SHALL` so this holds.

All five were run on a trial insert of the exact text above and passed; the implementer's task
repeats them on the committed edit.

**No new test, and none available.** Nothing asserts that a future planner obeys the paragraph —
the same limit KAN-465's design records for its own paragraph. The task declares `**Tests:** none`
and rests on the guard run above.

## Decisions

### Planner prose only, no plan-shape guard

**ID:** prose-only-no-guard
**Status:** active
**Chosen:** One paragraph in `skills/flow/brainstorm-planner.md` and no new rule in
`scripts/check-plan-shape.sh` — the issue's fix statement is the planner's instruction, and the
post-KAN-441 evidence is that implementers follow whatever the plan says, so the plan text is the
causal chain.
**Considered:** A `check-plan-shape.sh` hit when one task's `**Files:**` mixes a `desktopTest/`
path with non-test paths — declined by the operator: it needs a per-project notion of which paths
are UI-test paths that the guard cannot derive from `.flow/project.md`, and enforces a habit no
evidence yet shows surviving the prompt fix.

### General wording, `desktopTest` as the named example

**ID:** general-rule-desktoptest-example
**Status:** active
**Chosen:** "Tests [that] live in a UI-test source set of their own — Compose Multiplatform's
`desktopTest`, and the like" — the rule covers `androidTest`, browser e2e suites and similar
separate source sets at the same paragraph length, without a second change per framework.
**Considered:** Literal `desktopTest` only — declined by the operator: precise for today's gymie
repos, silent for every other UI-test layout.

### One follow-on task per feature task, carrying all its UI-test files

**ID:** one-followon-per-feature
**Status:** active
**Chosen:** The follow-on's `**Files:**` names every UI-test file of its feature task. The lever is
a fresh context for the UI-test iteration, not maximal splitting; a second follow-on per file would
add a dispatch boundary and a commit for no further context saving.
**Considered:** One follow-on task per UI-test file — declined by the operator.

### Placement after the verify-step paragraph

**ID:** placement-after-verify-step
**Status:** active
**Chosen:** Immediately after KAN-465's verify-step blockquote, before the plan-provenance load —
adjacent to the only other paragraph in **D** that constrains a task's content rather than its
declared fields.
**Considered:** Inside the `flow-task-commit-fields` bullet list — rejected for the reason
KAN-465's `placement-after-task-shape` records: that list is the mechanically-checked field family
and a paragraph about task granularity placed among them reads as a field. Appending to the
`superpowers:writing-plans` invocation paragraph — rejected as burying a constraint inside the
sentence that hands control to another skill.

### No `**After:**` field on the follow-on

**ID:** no-after-field-on-followon
**Status:** active
**Chosen:** The paragraph says the field is not needed: a task without one runs after every earlier
task, which places the follow-on after its feature task by plan order alone.
**Considered:** Requiring `**After:** Task <feature>` on every follow-on — rejected: it restates
the serial default and invites a planner to annotate the feature task's siblings for parallelism
this change does not ask for.

## Open questions

_none_
