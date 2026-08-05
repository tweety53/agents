## Why

Two defects surfaced during the KAN-9 frontend run, both in myflow itself. `/myflow-start`
produced a plan whose task order left `:shared` uncompilable for seven consecutive tasks, requiring
a human-adjudicated mid-run repair. Separately, `/myflow-status` (and every other `/myflow-*`
command, via the same shared resolution logic) enumerates open changes from `openspec list --json`
alone, so a change whose planning artifacts exist only in a worktree — staged but never committed
to the main checkout — is silently invisible even while it sits at the human gate.

## What Changes

- Add a required per-task `**Build:**` tag (`green`, or `red — merges with Task <N>`) to every plan
  `/myflow-start` produces, enforced by a new guard script run before the proposal artifact
  publishes.
- Add `skills/myflow-contracts/build-green.md` as the canonical contract for the tag format and the
  guard's rule and scope, indexed alongside the other contract files.
- Fix the shared `## Change name resolution` section in `pipeline.md` (used by `/myflow-start`,
  `/myflow-do`, `/myflow-finish`, `/myflow-status`) to enumerate from the union of
  `openspec list --json` and the project's state directory, not `openspec list --json` alone.
- Update `skills/myflow-status/SKILL.md` step 1 to cite the shared resolution section instead of
  running `openspec list --json` on its own.
- Register the new guard and its test in this repo's `.myflow/project.md` (`## lint`, `## test`).

## Capabilities

### New Capabilities
- `myflow-build-green`: a plan produced by `/myflow-start` must declare, per task, whether the
  build (and that task's own verification command) is green after it, and a guard enforces the
  declaration is present and that no declared-red task is left unresolved.

### Modified Capabilities
- `myflow-command-surface`: change-name resolution (used by every `/myflow-*` command) must
  enumerate every change tracked in the project's state directory, not only those whose
  `openspec/changes/<name>/` directory happens to be present in the current checkout.

## Impact

- `skills/myflow-start/SKILL.md` — writing-plans stage (section D) gains the tagging instruction
  and the pre-publish guard run.
- `skills/myflow-contracts/build-green.md` — new file.
- `skills/myflow-contracts/SKILL.md` — index gains the new contract file's row.
- `skills/myflow-contracts/pipeline.md` — `## Change name resolution` section rewritten.
- `skills/myflow-status/SKILL.md` — step 1 updated to cite the shared section.
- `scripts/check-task-build-green.py`, `scripts/check-task-build-green.sh`,
  `scripts/test-check-task-build-green.sh` — new.
- `.myflow/project.md` — `## lint` and `## test` gain the new guard and its test.
- No other repo (`gymie`, `gymie-frontend`) is touched — those items of KAN-13 are separate
  changes.
