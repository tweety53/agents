# kan-442-baseline-verification-cycle-triples-gradle-runs — design

## Context

`scripts/check-task-commit-fields.py` (1346 lines) checks one task's declared fields against one
commit.
<!-- measured: wc -l scripts/check-task-commit-fields.py @ main 285cb7d --> `Files:`, `Tests:`, `Commit:` and the scope check read git objects
(`git diff --name-only`, `git diff`, `git log -1`). `Regression:` and `Baseline:` do not: lines
976–1239 hold
<!-- measured: grep -n 'def read_single_test_command\|def check_baseline' scripts/check-task-commit-fields.py @ main 285cb7d --> `read_single_test_command` (reads `.flow/project.md`'s `## test` fence, returns the
command only when it is exactly one line), `_run_test_command`, `_commit_reverted` (revert
--no-commit / yield / reset --hard), `_uncommitted_work_protected` (stash push
--include-untracked / yield / stash pop, with the refuse-to-reset-on-pop-conflict path KAN-162
added), `check_regression` and `check_baseline`. `check_task_commit` (lines 1241–1313) wraps the
two in the stash context
<!-- measured: grep -n 'def check_task_commit\|return violations, notices' scripts/check-task-commit-fields.py @ main 285cb7d --> and sorts their `CheckOutcome`s into `notices` (skipped) or
`violations`; `main` prints notices before violations.

The harness `scripts/test-check-task-commit-fields.sh` (3328 lines, 91 cases) exercises the
cycle
<!-- measured: wc -l scripts/test-check-task-commit-fields.sh; grep -c '^# Case' scripts/test-check-task-commit-fields.sh @ main 285cb7d --> in cases 11–14 (pass/skip for each field), 16 (KAN-162 stash-before-revert) and 83–86
(KAN-162 staged/untracked survival, pop-conflict refusal, in-flight revert failure), through
five fixtures: `write_project_md_test_section`, `write_test_runner`,
`write_side_effect_test_runner` and `write_double_side_effect_test_runner` (each a `run_suite.sh`
speaking the `COUNT:`/`RESULT` protocol) and `write_unsupported_test_runner`.

`skills/flow/implement.md` (29818 bytes; budget 32500 in `scripts/check-contract-budget.sh`)
runs the guard in its conductor step 2, overlapped with the next implementer, on the claim at
line 431 that the guard reads git objects and `tasks.md` only.
<!-- measured: wc -c skills/flow/implement.md; grep -n 'implement.md' scripts/check-contract-budget.sh; grep -n 'reads git objects' skills/flow/implement.md @ main 285cb7d -->

The frozen `openspec/specs/myflow-task-commit-fields/spec.md` carries **Requirement:
Regression and Baseline checks skip, rather than fail, when unsupported**. That tree is frozen
at the spectre cutover and governs nothing (`skills/flow-contracts/model-policy-rationale.md`);
`spectre/specs/` is empty. The live requirement for this guard is its own module docstring and
`implement.md`.

## Change

### 1. The guard

Delete, from `scripts/check-task-commit-fields.py`:

- `check_regression`, `check_baseline`, `_commit_reverted`, `_uncommitted_work_protected`,
  `_STASH_MESSAGE`, `read_single_test_command`, `_run_test_command`;
- `TARGETED_RESULT_RE`, `TOTAL_COUNT_RE`, `PROJECT_TEST_SECTION_RE`, `MARKDOWN_HEADING_RE`,
  `FENCE_LINE_RE` and the comment block explaining the test-runner contract;
- `CheckOutcome`;
- the `contextlib` and `shlex` imports (nothing else uses them; `subprocess` stays for
  `run_git`).

`check_task_commit` returns `List[str]` (violations); the `with _uncommitted_work_protected`
block and the `notices` list go. `main` prints violations and exits 1, or exits 0 silently.
The module docstring's first paragraph, its exit-code table and the `check_task_commit`
docstring stop describing the two checks; the citation of the frozen spec requirement is
reworded to state that the runtime check it described was removed by KAN-442 because it
mutated the worktree the conductor overlaps and verified nothing against any real runner. The
`TaskFields.baseline` field and the `Baseline`/`Regression` entries in the field regexes stay:
the plan grammar is unchanged.

### 2. The harness

Delete cases 11, 12, 13, 14, 16, 83, 84, 85, 86 and the five fixtures
`write_project_md_test_section`, `write_test_runner`, `write_side_effect_test_runner`,
`write_double_side_effect_test_runner` and `write_unsupported_test_runner`. Add one case at the
end: a repository with a plan-declared task commit, then a staged edit to a tracked file, an
unstaged edit to another, and an untracked file; run the guard; assert exit 0, `git rev-parse
HEAD` unchanged, `git status --porcelain` byte-identical before and after, and `git stash list`
empty. Case numbering elsewhere is untouched — a gap in the numbers is cheaper than renumbering
the surviving cases and every cross-reference to them.

### 3. The conductor rule (KAN-438 c)

In `skills/flow/implement.md`, under **The next implementer overlaps the guard**, after step 2's
paragraph, one new paragraph:

> **A guard call that times out is inspected before it is retried.** Run
> `git status --porcelain=v2 --branch` and `git stash list` in that worktree first. A
> `# ... reverting`/rebase/merge state, a change the run did not make, or a stash entry the
> conductor did not push means the tree is not the one the run left — end the turn with
> `## Question` carrying both outputs verbatim; never re-run the guard on top of it.

Prose only — no `SHALL`/`MUST`, so `scripts/check-normative-inventory.sh` is unchanged; well
inside the byte budget (2682 bytes of headroom today).
<!-- measured: 32500 − 29818, the two figures under Context @ main 285cb7d -->

## Decisions

### Delete the runtime verification rather than relocate it

**ID:** delete-not-relocate
**Status:** active
**Chosen:** remove `check_regression`/`check_baseline` and the revert/stash machinery — the
fields stay as plan declarations only. Zero extra suites per task, and the guard becomes
genuinely read-only, which removes the incident class rather than fencing it.
**Considered:** running the cycle in a throwaway `git worktree add --detach <parent>` — no
mutation of the canonical tree, but still two full suites per task for a protocol no runner
speaks. Replacing `Baseline:` with `git grep -c "@Test" <sha>^` vs `<sha>` (KAN-438 a /
KAN-447 1) — no mutation and no suite, but annotation counts never equal a runner's executed
totals (parameterized tests, Kotlin/JS files, this repository's shell harnesses), so declared
`before=/after=` would fail real tasks, and the pattern needs per-project configuration.

### No lock file and no recovery script

**ID:** no-lock-no-recovery
**Status:** active
**Chosen:** leave KAN-438 (b) and KAN-447 (2) out. A read-only guard has nothing to serialize
and produces no state to recover.
**Considered:** a `.git/flow-guard.lock` — guards against re-entrancy the change removes.
A `flow-recover-guard.sh` — automates recovery from a state the guard can no longer create.

### Fold in KAN-438 (c) as one prose paragraph

**ID:** fold-in-438c
**Status:** active
**Chosen:** the post-timeout inspection rule lands in `skills/flow/implement.md`, the file that
already states the guard/implementer overlap contract; prose only, no guard.
**Considered:** leaving it to KAN-438 — rejected by the operator: the rule is a few lines in a
file this change already reasons about.

### No spec edit

**ID:** no-spec-edit
**Status:** active
**Chosen:** `openspec/specs/myflow-task-commit-fields/spec.md` stays frozen and untouched;
`spectre/specs/` stays empty; the guard's docstring records the removal.
**Considered:** editing the frozen spec — the tree governs nothing and has never been edited
since the cutover. Creating `spectre/specs/flow-task-commit-fields.md` — a new live spec for
one deletion, with no other capability in that tree to sit beside.

## Open questions

None recorded — every question raised in brainstorming was answered.
