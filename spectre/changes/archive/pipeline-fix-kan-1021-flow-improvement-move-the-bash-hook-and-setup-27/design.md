## Context

In-run pipeline fix found by kan-1021's self-review pass. Evidence: the archived plan
`/Users/tweety53/Projects/agents-worktrees/kan-1021-flow-improvement-move-the-bash-hook-and-setup/spectre/changes/archive/kan-1021-flow-improvement-move-the-bash-hook-and-setup/tasks.md`
carries a `Correction (2026-10-23)` under Tasks 5 and 6: each sweep
`git grep -n -F -e '<harness>.sh' -- stats scripts hooks mods README.md CLAUDE.md AGENTS.md .flow`
still named `stats/internal/guard/pipelinefix.go` and `pipeline_fix_test.go`, which Task 10 owns.
At kan-1021's base `e77f951b` both files hold `test-git-config-pins.sh` and
`test-flow-addr-declaration.sh` (`git grep -n -F -e 'test-git-config-pins.sh' -e
'test-flow-addr-declaration.sh' e77f951b -- stats`). Tasks 1–4's misfire (panel F2) was a sweep
hitting doc comments a step told the task to keep.

The rule the plan broke is **End a removal task on an absence sweep**
(`skills/flow/brainstorm-planner.md`): a sweep is scoped to the paths the task touches. No guard
reads it. `check-plan-shape.sh` is the shim over `stats/internal/guard/planshape.go`, which already
resolves the change's base for F15 (`baseDeclarations`).

**Measured against history.** A path-cover-only rule (a sweep path equal to, or a directory above,
a later task's `**Files:**` path) refuses pairs in plans that ran clean:
119 pairs in kan-1021's plan and 110 in kan-887's
<!-- measured: python3 -I /private/tmp/claude-501/-Users-tweety53-Projects-agents/a7689118-38a4-48ca-9ece-631a1203fb42/scratchpad/measure.py over every spectre/changes/archive/*/tasks.md @ e77f951b, plus kan-1021's archived plan @ branch kan-1021-flow-improvement-move-the-bash-hook-and-setup; a session-local script, not re-runnable from the tree -->
— most of them sweeps for a string the later task's file never held. Adding the content check
(the covered file holds a swept string at the plan's base) leaves exactly Tasks 5 and 6 of
kan-1021, two files each, and nothing in the other archived plans
<!-- measured: python3 -I /private/tmp/claude-501/-Users-tweety53-Projects-agents/a7689118-38a4-48ca-9ece-631a1203fb42/scratchpad/measure2.py, base = parent of each plan's earliest task commit, over the same plans @ e77f951b; a session-local script -->
(6 of the 267 archived plans carry such a sweep at all).
<!-- measured: ls -d spectre/changes/archive/*/ | wc -l @ e77f951b → 267; the sweep count from measure2.py's per-plan sweep lines -->
The prototype guard (Task 1's diff) reproduces that: four lines on an unticked copy of kan-1021's
plan, none over `spectre/changes/archive/*/tasks.md`.
<!-- measured: scripts/check-plan-shape.sh .superpowers/sdd/kan-1021-unticked-tasks.md @ branch pipeline-fix-kan-1021-flow-improvement-move-the-bash-hook-and-setup-27 with Task 1's diff applied → exit 1, four lines; the per-archive loop → 0 'absence sweep searches' lines -->

Blast radius: 4 files plus this change's `verbatim-moves.txt` — not `big`.
<!-- measured: git apply --stat of Task 1's diff @ e77f951b → 4 files -->
No saved memory encodes the defect.

## Decisions

### Where the check lives

**ID:** sweep-scope-is-f18
**Status:** active
**Chosen:** finding F18 of `check-plan-shape` in `stats/internal/guard/planshape.go`, tested as case 43 of `TestCheckPlanShape` — the plan-time guard every planner already runs, reusing its parser and F15's base.
**Considered:** a separate plan sub-guard file like `plansuitesteps.go` — a new file, registry entry and test for one check that needs F15's base resolution anyway; a check in `check-task-commit-fields` at commit time — fires after the plan is already wrong, the misfire this fix exists to prevent.

### What the guard refuses

**ID:** sweep-content-at-base
**Status:** active
**Chosen:** a sweep is refused only when a covered later-task file holds one of its `-e` strings at the change's base (`git grep -l -F`), which is exactly what makes the sweep print at its own commit.
**Considered:** path cover alone — refuses hundreds of pairs in plans that ran clean (Context), so it would refuse legitimate history and train planners to ignore it.

### Ticked tasks are not read

**ID:** sweep-ticked-skipped
**Status:** active
**Chosen:** a task whose checkbox is ticked is skipped — its sweep already ran — so kan-1021's own archived plan, scanned as an unlanded archive while the base still holds the strings, is not refused once this lands.
**Considered:** reading the working tree instead of the base — still refuses a committed-but-unticked task mid-run and departs from F15's base; exempting a sweep followed by a `Correction:` line — a second grammar for the same fact.

### Which sweeps are recognised

**ID:** sweep-shape-prescribed
**Status:** active
**Chosen:** one line carrying a backticked `` `git grep … -- <paths>` `` followed by `prints nothing`, with each string as `-e '<string>'` — the shape the rule prescribes; anything else is not read (a `ponytail:` note in the code names the ceiling).
**Considered:** parsing wrapped steps and every quoting form — more parser for shapes no archived sweep uses.

### Base resolution shared with F15

**ID:** base-rev-shared
**Status:** active
**Chosen:** extract `baseDeclarations`' merge-base resolution into `(*psPlan).baseRev`, used by F15 and F18.
**Considered:** a second copy in F18 — two base resolutions that can drift.

### The rule's text

**ID:** sweep-scope-sentence
**Status:** active
**Chosen:** one sentence under **End a removal task on an absence sweep** naming the limit and F18, including the keep-a-doc-comment half (Tasks 1–4) as prose; the reason as one bullet in `skills/flow/SKILL-rationale.md`; the sentence listed in this change's `verbatim-moves.txt` for `check-verbatim-moves.sh`.
**Considered:** guarding the doc-comment half too — needs reading what a step's prose tells the task to keep, not mechanical; no sentence at all — leaves the guard's refusal unexplained to the planner reading the rule.

### Nothing to repair

**ID:** no-repair-needed
**Status:** active
**Chosen:** no repair step — the defect's only output, kan-1021's misfired sweeps, already carries its `Correction:` lines in that plan, and no store row or record was written wrong.
**Considered:** rewriting kan-1021's archived sweeps — rewrites a closed record its Corrections already explain.

### First review's effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — the guard runs in every plan's lint and shells out to git; a false positive blocks every later plan, which a medium review weighs better.
**Considered:** flow-low — the diff is small, but the risk is in the refusal's reach, not its size.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group — Task 1 alone, one fixer dispatch.
**Considered:** more groups — the plan has one task.

## Open questions

## Live check

None — the change is a plan-time guard with no running service or persistent state; Task 1's steps run it against kan-1021's plan and the archived plans.
