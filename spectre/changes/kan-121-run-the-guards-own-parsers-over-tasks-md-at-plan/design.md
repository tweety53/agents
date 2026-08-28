# kan-121-run-the-guards-own-parsers-over-tasks-md-at-plan

## Context

`proposal.md` is canonical for why this change exists and what it changes. This file is canonical
for every decision below and for the guard's own contract.

## What is actually live today

Verified by running `check-task-commit-fields.py`'s own `parse_task_fields` over a fixture, not
read off the ticket. Two of the ticket's five findings have changed shape since it was filed.

| # | Ticket defect | Status |
|---|---------------|--------|
| 1 | duplicate `**Files:**` fields | **Live.** `**Files:** \`a.md\`` followed by `**Files:** \`b.md\`` parses to `['b.md']`; `a.md` is silently discarded. `parse_task_fields`'s `fields[current_name] = current_parts` overwrites. |
| 2 | a `**Files:**` field smaller than its own prose | **Live.** A task with no `**Files:**` line at all parses to `[]` with no complaint at plan time. |
| 3 | a task heading the parser cannot find | **Obsolete as written.** The grammar moved from `### <dotted-id>` headings to `- [ ] <n>.` checkbox lines (`lib/plan_grammar.py`'s `select_task`). The class survives in a new form: `FIELD_RE` anchors at column 0, so a `**Files:**` line indented under its own task line is invisible to the guard — and plan bodies are indented. |
| 4 | backticks in a `Tests:` field misread as declared test names | **Live, and shipped.** `_parse_test_specs` still falls back to `BACKTICK_RE.findall(value)`. `Tests: none added — verified via the \`scripts/check-references.sh\` guard` yields `scripts/check-references.sh` as a test name that must appear in the commit's diff. `spectre/changes/archive/kan-171-generic-visual-verification-step/tasks.md` task 2 carries exactly this shape today. |
| 5 | a plan with no checkbox lines | **Live.** Now identical to KAN-114: zero checkbox lines yields zero tasks, and both plan guards report clean. A vacuous pass, not a detection. |

## Decisions

### Scope of the change

**ID:** scope-all-five-plus-parser-fix
**Status:** active
**Chosen:** All five defects, plus the real parser fix the ticket asks for under defect 4 — a
recognised literal so a no-test task can say `none` without its backticks becoming test names.
**Considered:**
- *Checks only, no parser change* — ruled out: without an opt-out, a no-test task still has no way
  to say so, and plans keep being reworded to satisfy a guard. The ticket names this explicitly:
  "The plan was reworded to satisfy the guard — prose changed, nothing real did."
- *Guard plus vacuity only* — ruled out for the same reason; leaves defect 4 as a warning about a
  problem the tooling created.
- *Minimum: run the existing parsers and report what they already report* — ruled out: the existing
  parsers report nothing at all for defects 1, 2 and 4, which is the whole failure. Running them
  unchanged surfaces only defect 3.

### Where the plan-time check lives

**ID:** new-guard-importing-real-parsers
**Status:** active
**Chosen:** A new `scripts/check-plan-shape.sh` + `scripts/check-plan-shape.py` that **imports**
`check-task-commit-fields.py` and `lib/plan_grammar.py` directly, so parser identity is structural
rather than asserted.
**Considered:**
- *Fold into `check-task-build-green`* — ruled out: conflates "is every task tagged" with "is every
  task parseable", and widens the exit-code contract of a guard callers already depend on.
- *A `--plan-only` mode on `check-task-commit-fields`* — ruled out: that guard's whole contract is
  one task against one commit; a mode with no commit argument means a second calling convention and
  a second exit-code meaning inside a 1231-line module.

The import is the point. The ticket's own "Shape of the fix" asks for a self-check that runs "the
**actual downstream parsers**, rather than prose in `/myflow-start` restating rules that live in
`build-green.md` and in the guards' source" — a reimplementation would be that prose in another
language.

### How a no-test task declares no tests

**ID:** tests-none-literal
**Status:** active
**Chosen:** A `Tests:` field whose value opens with the literal `none` declares zero tests; the
rest of the field is free prose and its backticks are never read as test names.
**Considered:**
- *Restrict the backtick fallback to test-looking tokens* — ruled out: `scripts/check-references.sh`
  and `scripts/test-check-visual-verification.sh` are indistinguishable by shape, so the exact case
  that broke KAN-106 and KAN-109 would still break. A heuristic here fails silently; the literal
  fails loudly.
- *Both the literal and the heuristic* — ruled out: two mechanisms to hold in step, and the
  heuristic's failure mode stays silent where the literal's is loud.

The opt-out cannot swallow a genuine declaration, because F4 below fails a `none` field on a task
whose `Files:` names a test-shaped path.

### Severity model

**ID:** hard-fail-certain-only
**Status:** active
**Chosen:** Exit 1 for every mechanically-decidable defect. No heuristic for defect 2's "a `Files:`
field smaller than its own prose".
**Considered:**
- *Warn on the fuzzy one* — ruled out: it would be this repository's first guard printing a finding
  without failing, and F2 below already catches the empty case that caused the real failures.
- *Hard-fail the heuristic too* — ruled out: a false positive blocks a correct plan, and this
  repository forbids suppression markers by policy, so there is no escape hatch.

### Who closes the vacuous pass

**ID:** vacuity-in-new-guard-only
**Status:** active
**Chosen:** `check-plan-shape` fails a `tasks.md` it can parse no tasks from.
`check-task-build-green` keeps its current exit contract unchanged.
**Considered:** *Close it in `check-task-build-green` as well* — ruled out for this change: it
alters a shipped guard's exit behaviour, so its harness and every caller assuming "clean" would need
revisiting, and the vacuity is already closed at the plan gate where the ticket asks for it.

### Which guard family it joins

**ID:** shipped-not-project-configured
**Status:** active
**Chosen:** Shipped — symlinked into `skills/flow/scripts/` beside `check-task-commit-fields.py`,
and named in this repository's own `## lint` list.
**Considered:** *Project-configured, like `check-plan-provenance` and `check-task-build-green`* —
ruled out: those two are resolved through a project's own `.flow/project.md` and run only where a
project declares them, but the guard this change exists to protect (`check-task-commit-fields.sh`)
is itself shipped and runs in every project `/flow` touches. A plan-shape guard that only exists
where a project opts in would leave that guard unprotected everywhere else.

`skills/flow/scripts/` is a directory of committed symlinks into `../../../scripts/`, so shipping
costs two symlinks and no `setup.sh` change. As a shipped guard it is cited by basename, per
**Guard resolution** (`skills/flow-contracts/pipeline.md`), and is therefore **not** added to
`check-guard-symlinks.sh`'s rule-3 `EXEMPT` table.

## The guard

### Calling convention

Mirrors `check-task-build-green.sh` exactly:

- no arguments — scan every non-archived `<spec-root>/changes/*/tasks.md`, aggregating exit codes;
- one argument — treat it as an explicit `tasks.md` path and scan only that file.

Exit `0` clean, `1` violations found, `2` it cannot answer (python3 missing or non-functional, a
required sibling module absent, an unreadable file). A `2` is never a silent skip.

The wrapper names `$SCRIPT_DIR/check-task-commit-fields.py` and `$SCRIPT_DIR/lib/plan_grammar.py`
explicitly and checks each is present, for the two reasons `check-task-commit-fields.sh`'s own
header gives: a Python `import` is invisible to `check-guard-symlinks.sh` rule 2, which derives a
guard's required siblings by grepping for `$SCRIPT_DIR/<name>`; and a missing module should say
which one rather than surface as a traceback.

### Importing the real parsers

`check-task-commit-fields.py`'s name carries a hyphen, so it is not importable by name. The guard
loads it through `importlib.util.spec_from_file_location` against its own `$SCRIPT_DIR`, and
registers the module in `sys.modules` before executing it — without that registration, the
module's `@dataclass` decorator raises `AttributeError: 'NoneType' object has no attribute
'__dict__'` on Python 3.14, because `dataclasses._is_type` looks the defining module up by name.

From it: `parse_task_fields`, `collect_task_ids`, `FIELD_RE`, `BACKTICK_RE`, `CASE_LABEL_RE`.
From `lib/plan_grammar.py`: `iter_tasks`, `select_task`, `unclosed_fence`, `FENCE_RE`.

### Findings

Every finding exits 1 and is reported as `file:line`, naming the task id.

| Key | Finding | Ticket defect |
|-----|---------|---------------|
| **F1** | A second `**Files:**`, `**Tests:**`, `**Commit:**`, `**Baseline:**` or `**Allowed-collateral:**` line in one task body — naming which occurrence currently wins, since the parser keeps the last | 1 |
| **F2** | No `**Files:**` line in a task body, or one whose backtick token list is empty | 2 |
| **F3** | A field-shaped line indented past column 0 inside a task body, invisible to `FIELD_RE`; and a task body whose fence never closes, surfaced at plan time rather than at commit time | 3 |
| **F4** | A `Tests:` field opening with `none` on a task whose `Files:` names a test-shaped path (`test-*`, `*_test.go`, `*.test.*`, `*.spec.ts`) — a contradiction between the opt-out and the declared files | 4 |
| **F5** | A `tasks.md` the grammar finds zero tasks in | 5 / KAN-114 |
| **F6** | A `Tests:` field that is neither `none`-opening, nor carries a `Case N` label, nor carries a backtick token — it declares nothing, so `check_tests` verifies nothing | *beyond the ticket* |

F6 is an addition beyond KAN-121's five, of the same class as F5: a field that passes because it
asserts nothing. It was presented as strikeable and kept.

Fences are excluded from every scan, per `lib/plan_grammar.py`'s `FENCE_RE` — a `**Files:**` line
inside a worked example is not a declaration, and reading one as such is the defect fix round 10
(F21) closed in `check-task-commit-fields.py`.

## The parser fix

`_parse_test_specs` gains one branch ahead of the `Case N` scan: a value whose first word is the
literal `none` — case-insensitive, leading Markdown emphasis and whitespace stripped — returns an
empty list. Everything after it is prose, and `BACKTICK_RE` is never applied.

`Tests: none added — verified via the \`scripts/check-references.sh\` guard` then declares nothing,
instead of requiring `scripts/check-references.sh` to appear in the commit's diff.

This is documented in `check-task-commit-fields.py`'s own module docstring, next to the existing
explanation of why case labels take priority over backtick tokens, and in
`skills/flow/brainstorm.md`'s `**Tests:**` field line — the field family's only live normative
statement. `spectre/specs/` is empty; the original capability spec is frozen in the `openspec/`
archive and is never written to again.

## Where it runs

| Site | What changes |
|------|--------------|
| `.flow/project.md` `## lint` | one entry, `scripts/check-plan-shape.sh` |
| `skills/flow/brainstorm.md` **D** | the writing-plans step's guard sentence gains the plan-shape guard, cited by basename |
| `skills/flow/implement.md` | the pre-dispatch step runs it before any task is dispatched |
| `skills/flow/SKILL.md` | the guard-presence list gains `check-plan-shape.sh` |
| `skills/flow/scripts/` | two committed symlinks, `check-plan-shape.sh` and `check-plan-shape.py` |

## Verification

`scripts/test-check-plan-shape.sh`, added to `.flow/project.md`'s `## test` list. Fixtures are
built under `TMPDIR`, and the harness **mutates the guard to prove each of F1–F6 actually fails** —
the requirement KAN-197 imposed on every harness in this repository.

Every owned `.md` this change grows reconciles its `check-contract-budget.sh` `budgets()` row in
the same commit, raising a row only where the guard actually fails on that file.

## Known consequence, and how it resolved

When this change was planned, `spectre/changes/` held two changes the store reported `FINISHED` but
which had never been moved to `archive/` — `kan-359-multi-glob-ui-paths-matches-nothing` and
`panel-roster-follows-the-settings-list` — so the new lint step's no-argument scan covered both.
Exactly one task tripped F6: kan-359's task 5, whose `**Tests:**` field named no test at all. It was
repaired in the working tree during task 3.

**Both changes archived on `main` while this branch was open** (PRs #64 and #62), which moved their
plans under `spectre/changes/archive/` — a path the guard's glob depth excludes. The F6 hit is
therefore gone at its source, and the repair was **dropped rather than carried onto the archived
file**: an archived plan is a frozen historical record, and rewriting one is what the archive exists
to prevent. Note that a plain rebase does *not* do this by itself — git followed the rename and
reapplied the edit to the archived path silently, which had to be undone by hand.

That archiving also cleared a second consequence, unrelated to lint: with more than one root change
under `spectre/changes/`, `check-task-commit-fields.sh` refuses with exit 2 (`cannot resolve which
change`), so the shipped guard was out of service for the whole of this change's implementation and
its per-task field checks had to be run by invoking the Python guard directly.

## Open questions

None.
