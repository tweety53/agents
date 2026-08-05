# Manual test guide — kan-13-myflow-planning-and-status-fixes

This repository declares no runnable application (`.myflow/project.md`'s `## apps` section). Run
the checks below from the repository root in this worktree:

```bash
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
scripts/check-task-build-green.sh
scripts/test-check-task-build-green.sh
openspec validate kan-13-myflow-planning-and-status-fixes --strict
```

## Checks

### `check-task-build-green.py`/`.sh` — the build-green guard

- [x] A `tasks.md` where every task is tagged `**Build:** green` passes (exit 0, no output).
- [x] A task with no `**Build:**` tag fails, naming the task's id and heading line.
- [x] A task tagged `**Build:** red` with no `merges with Task …` clause fails.
- [x] A task tagged `red — merges with Task <N>` where `<N>` does not exist in the plan fails,
      naming the missing partner.
- [x] A task tagged `red — merges with Task <N>` where `<N>` is itself tagged `red` fails, reported
      against the task carrying the tag, not the partner.
- [x] A task tagged `red — merges with Task <N>` where `<N>` is tagged `green` passes.
- [x] A malformed tag (e.g. `**Build:** yellow`) is treated as no tag, not a separate error class.
- [x] A `**Build:**`-shaped line inside a fenced code block is never mistaken for a real tag, and a
      `###`-shaped heading inside a fenced code block never opens a phantom task.
- [x] A duplicate task id (two `###` headings sharing one dotted id) fails, naming both the
      duplicate and the line of the first occurrence.
- [x] A heading with an id but no title text (`### 1.1` alone) is still recognized as a task.
- [x] A partner named twice in one clause (`merges with Task 2.1, 2.1`) produces exactly one
      violation line, not two.
- [x] Invoked with no arguments, scans every non-archived change's `tasks.md` under
      `openspec/changes/`, aggregates a non-zero result when any file violates, and excludes
      `openspec/changes/archive/`.
- [x] `CHECK_TASK_BUILD_GREEN_ROOT` overrides which tree the no-argument scan reads, for testing
      against a sandboxed fixture tree.
- [x] Run against this change's own `tasks.md`, exits 0 clean.

### `skills/myflow-contracts/build-green.md` — the contract

- [x] States the tag vocabulary (`green` / `red — merges with Task <N>`) and the placement rule
      (first matching line in a task's body) in prose matching the guard's actual regex behavior.
- [x] States the guard's scope and explicitly states it never verifies a `green` tag is true.
- [x] Is indexed in `skills/myflow-contracts/SKILL.md`'s table, in the same row shape as the other
      six contract files.

### `skills/myflow-start/SKILL.md` — writing-plans integration

- [x] Section D instructs tagging every task with `**Build:**`, citing `build-green.md` rather than
      restating the tag rules.
- [x] The pre-publish guard-run sentence names both the plan-provenance guard and the build-green
      guard.
- [x] The Guardrails section forbids publishing a plan with an untagged task or an unresolved `red`
      chain.
- [x] The "both name and description omitted" resolution step cites the shared **Change name
      resolution** section in `pipeline.md` rather than re-deriving `openspec list --json` filtering.

### `skills/myflow-contracts/pipeline.md` / `skills/myflow-status/SKILL.md` — enumeration fix

- [x] `## Change name resolution` states the union rule: `openspec list --json`'s non-archived
      names, plus the basenames of every file in the project's state directory, minus anything
      archived.
- [x] States that an unreadable state-directory file is reported and skipped, never silently
      dropped.
- [x] `/myflow-status`'s step 1 cites this shared section instead of running its own independent
      `openspec list --json`.

### `.myflow/project.md` — bookkeeping

- [x] `## lint` lists `scripts/check-task-build-green.sh` alongside the three existing guards.
- [x] `## test` lists `scripts/test-check-task-build-green.sh` alongside the nine existing test
      scripts.

### Repo-wide regression

- [x] Every guard in `## lint` (`check-vocabulary.sh`, `check-references.sh`,
      `check-plan-provenance.sh`, `check-task-build-green.sh`) exits 0.
- [x] Every test in `## test` (all ten `scripts/test-*.sh`) passes with 0 failures.
- [x] `openspec validate kan-13-myflow-planning-and-status-fixes --strict` passes.

## Known incomplete

None.
