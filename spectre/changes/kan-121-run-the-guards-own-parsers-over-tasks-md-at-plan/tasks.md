# kan-121-run-the-guards-own-parsers-over-tasks-md-at-plan

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

`design.md` is canonical for every decision below, for the six findings F1–F6, and for the guard's
exit-code contract. Nothing here restates them.

**Baseline, measured before any edit:**

- `scripts/` carries 28 guards and 37 test harnesses.
<!-- measured: ls scripts/ | grep -c '^check-'; ls scripts/ | grep -c '^test-' @ a8a0ce9 (before this change) -->
- `spectre/changes/` holds three non-archived changes: this one, `kan-359-multi-glob-ui-paths-matches-nothing`
  and `panel-roster-follows-the-settings-list`. The latter two are `FINISHED` in the store but were
  never moved to `spectre/changes/archive/`, so the new guard's no-argument scan covers them.
<!-- measured: ls spectre/changes/; flow state list -C . @ a8a0ce9 (before this change) -->

**Every task that grows an owned `.md` file runs `scripts/check-contract-budget.sh` and reconciles
its `budgets()` row in the same commit** — raise a row only where the guard actually fails on that
file, and only to the size its own rule gives.

**One lint hit was repaired outside a task commit, and then dropped.** After task 2 landed, exactly
one task in the repository still tripped F6: kan-359's task 5, whose `**Tests:**` field named no test
at all. `spectre/changes/` is a **planning path** — per **Git boundaries**
(`skills/flow-contracts/git-boundaries.md`) an implementation commit is staged with it excluded and
the planning commit at integrate carries it — so task 3 repaired that line in the working tree and it
was **never** named in any task's `**Files:**` field.

**That repair is no longer part of this change.** kan-359 archived on `main` while this branch was
open, moving its plan under `spectre/changes/archive/`, which the guard's glob depth excludes. See
**Known consequence, and how it resolved** (`design.md`) for why the edit was dropped rather than
carried onto the archived file.

**Task 2 must land before tasks 3 and 4, and the order is load-bearing.** Both of those tasks
declare `**Tests:** none …` with script names in backticks. Under today's parser those backticks
are read as test names that must appear in the commit's diff, so either task would fail
`check-task-commit-fields.sh` if committed before task 2's `none` branch exists. This plan is its
own first regression test for that fix.

---

- [x] 1. The plan-shape guard and its harness

  Write `scripts/test-check-plan-shape.sh` first, then `scripts/check-plan-shape.py` and
  `scripts/check-plan-shape.sh` against it — RED before GREEN.

  `check-plan-shape.sh` is the thin wrapper, following `check-task-build-green.sh`'s split and
  calling convention exactly: no arguments scans every non-archived
  `<spec-root>/changes/*/tasks.md` and aggregates exit codes; one argument is an explicit
  `tasks.md` path. Resolve the spec root through `scripts/lib/spec-root.sh`, as both existing plan
  guards do. Name `$SCRIPT_DIR/check-task-commit-fields.py` and `$SCRIPT_DIR/lib/plan_grammar.py`
  in the wrapper and check each is present — a Python `import` is invisible to
  `check-guard-symlinks.sh` rule 2, which derives a guard's siblings by grepping for
  `$SCRIPT_DIR/<name>`, and a missing module must say which one rather than surface as a traceback.
  Probe `python3` with a trivial program, not `command -v` alone, for the reason
  `check-task-build-green.sh`'s own comment gives.

  `check-plan-shape.py` imports the real parsers rather than reimplementing them — the whole point
  of the change, per `design.md`'s `new-guard-importing-real-parsers`. Load
  `check-task-commit-fields.py` through `importlib.util.spec_from_file_location` against the
  script's own directory, and **register the module in `sys.modules` before executing it**:

  ```python verified:reproduced against Python 3.14.6 while writing this plan — without the sys.modules line the module's @dataclass raises AttributeError: 'NoneType' object has no attribute '__dict__'
  spec = importlib.util.spec_from_file_location("cfg", SCRIPT_DIR / "check-task-commit-fields.py")
  module = importlib.util.module_from_spec(spec)
  sys.modules["cfg"] = module
  spec.loader.exec_module(module)
  ```

  Implement F1–F6 exactly as `design.md`'s findings table defines them, each reported as
  `file:line` naming the task id, each exiting 1. Exclude fenced regions from every scan using
  `lib/plan_grammar.py`'s `FENCE_RE` — a field line inside a worked example is not a declaration.
  Exit `2` for python3 missing or non-functional, a required sibling absent, or an unreadable file;
  a `2` is never a silent skip.

  The harness builds every fixture under `TMPDIR` and **mutates the guard to prove each of F1–F6
  actually fails** — the requirement `scripts/test-check-reproduce-not-read.sh` and every other
  harness here already meets. Include a case per finding, a clean-plan case, and a `2` case.

**Files:** `scripts/check-plan-shape.py`, `scripts/check-plan-shape.sh`, `scripts/test-check-plan-shape.sh`
**Tests:** `scripts/test-check-plan-shape.sh`
**Regression:** reverting this commit removes the only check that a `tasks.md` is parseable by the
  guard that will later judge its commits, returning every plan-shape defect to surfacing after a
  task has been implemented, reviewed and committed.
**Baseline:** before=37 after=38 harnesses in `scripts/`
<!-- predicted: ls scripts/ | grep -c '^test-' after task 1 -->
**Commit:** `feat(scripts): guard tasks.md shape at plan time`
**Build:** green

- [x] 2. A `Tests:` field that declares none

  Add one branch to `_parse_test_specs` in `scripts/check-task-commit-fields.py`, ahead of the
  `Case N` scan: a value whose first word is the literal `none` — case-insensitive, leading
  Markdown emphasis and whitespace stripped — returns an empty list, and `BACKTICK_RE` is never
  applied to it.

  This is the real fix `design.md`'s `tests-none-literal` records. Today a `Tests:` field reading "none added — verified via the guard" with a script name in
  backticks yields that script name as a test that must appear in the commit's diff, which is what
  pushed plan authors to vaguer wording in the one field that should be precise.

  Document the branch in the module docstring's `Tests:` section, beside the existing explanation
  of why case labels take priority over backtick tokens. Add harness cases to
  `scripts/test-check-task-commit-fields.sh`: a `none`-opening field with backticks declares
  nothing; a `none`-opening field with a `Case N` label still declares nothing; a field merely
  containing the word "none" mid-sentence is unaffected.

**Files:** `scripts/check-task-commit-fields.py`, `scripts/test-check-task-commit-fields.sh`, `scripts/check-plan-shape.py`
**Tests:** `scripts/test-check-task-commit-fields.sh`
**Regression:** reverting this commit restores the backtick fallback over a no-test field, so a
  task that legitimately adds no test is failed for not adding a test whose name the guard invented
  out of the field's own prose.
**Baseline:** before=38 after=38 harnesses in `scripts/`
<!-- predicted: this task adds cases to an existing harness and no new harness file -->
**Commit:** `fix(scripts): let a Tests field declare none`
**Build:** green

- [x] 3. Ship the guard and declare it

  Add the two committed symlinks in `skills/flow/scripts/`, pointing at `../../../scripts/`, the
  same shape every other entry there has. As a shipped guard it is cited by basename per **Guard
  resolution** (`skills/flow-contracts/pipeline.md`), so it gets **no** entry in
  `check-guard-symlinks.sh`'s rule-3 `EXEMPT` table — that table is for project-configured guards
  only.

  Add `scripts/check-plan-shape.sh` to `.flow/project.md`'s `## lint` list and
  `scripts/test-check-plan-shape.sh` to its `## test` list. Add the guard's `budgets()` row in
  `scripts/check-contract-budget.sh` if the file grows past its row, and reconcile any `.md` row
  this task's edits trip.

  **`scripts/check-plan-shape.sh` is in this task's file set because shipping the guard broke it.**
Its no-argument `REPO_ROOT` derivation used `$SCRIPT_DIR/..`, correct only from `scripts/` and
silently wrong once the guard is also reachable through `skills/flow/scripts/`, where it resolves
to `skills/flow` instead of the repository root. `scripts/check-guard-symlinks.sh` rule 4 catches
it the moment the symlinks land. The fix follows `check-workspace-isolation.sh`, the existing
shipped guard with the same problem: source `scripts/lib/resolve-file.sh` and derive the root from
the script's real, symlink-resolved location. `scripts/check-contract-budget.sh` left this field
when no `budgets()` row turned out to need raising.

**Repair the one outstanding F6 hit in the working tree here**, per the note above the task list:
  reword kan-359's task 5 so its `**Tests:**` field opens with `none` or names its harness in
  backticks. Do **not** stage or commit it — the planning commit at integrate carries it. *(Done at
  the time; the repair was later dropped when kan-359 archived — see the note above the task list.)*

  Run `scripts/check-guard-symlinks.sh` and `scripts/check-plan-shape.sh` and leave both clean.

**Files:** `skills/flow/scripts/check-plan-shape.sh`, `skills/flow/scripts/check-plan-shape.py`, `.flow/project.md`, `scripts/check-plan-shape.sh`
**Tests:** none — this task adds symlinks and configuration entries; what verifies it is
  `scripts/check-guard-symlinks.sh` and the guard's own presence in the lint run
**Regression:** reverting this commit leaves the guard written but installed nowhere and declared
  nowhere, so no `/flow` run in any project ever invokes it.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `chore(scripts): ship and declare the plan-shape guard`
**Build:** green

- [x] 4. Run it before dispatch

  Three skill edits, each citing the guard by basename.

  `skills/flow/brainstorm.md` — in **D. Basic Workflow #3 — Writing plans**, the sentence "run the
  project's configured plan-provenance guard and its configured build-green guard, if the project
  declares them" gains the plan-shape guard, worded as what it is: a shipped guard, run
  unconditionally, not one resolved through the project's configuration. In the same file's
  `**Tests:**` field-family line, state that a task adding no test writes a field opening with the
  literal `none`, and that its backticks are then never read as test names — that line is the field
  family's only live normative statement, `spectre/specs/` being empty and the original capability
  spec frozen in the `openspec/` archive.

  `skills/flow/implement.md` — in **1. Load context and validate the plan**, run the guard against
  the change's `tasks.md` immediately after `spectre validate`, before any task is dispatched. A
  nonzero exit is a plan defect repaired here, before any code is touched, exactly as an exit-1
  from `spectre validate` already is.

  `skills/flow/SKILL.md` — add `check-plan-shape.sh` to the guard-presence list, in the union
  carried by `skills/flow/scripts/`.

  Reconcile each edited file's `check-contract-budget.sh` row in this commit.

**Files:** `skills/flow/brainstorm.md`, `skills/flow/implement.md`, `skills/flow/SKILL.md`
**Tests:** none — this task edits skill prose; the guards that verify prose in this repository are
  `scripts/check-references.sh`, `scripts/check-vocabulary.sh`, `scripts/check-contract-budget.sh`,
  `scripts/check-markdown-integrity.py` and `scripts/check-installed-citations.sh`
**Regression:** reverting this commit leaves a shipped, installed guard that no `/flow` stage ever
  runs, so every plan-shape defect still reaches dispatch unchecked — the exact failure this change
  exists to end.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `docs(flow): run the plan-shape guard before dispatch`
**Build:** green
