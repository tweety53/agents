# Remove the manual test guide and its gate

## Why

`/myflow-do` writes `docs/manual-test/<name>.md` — a tickable behaviour checklist with a
`## Known incomplete` section — and `/myflow-finish`'s unfinished-work guard reads it back as two of
its four signals. The artifact costs a generation step on every run and a maintenance step on every
fix round, and it produces a second completion record beside `tasks.md` that a guard must then
police.

What the operator actually needs at `IN_PROGRESS` is the instructions for running whatever is in
scope, with absolute paths and this worktree's resolved URLs. That is the preamble of the guide, not
its checklist.

## What changes

Delete the guide, the directory holding it, and everything that reads either. In its place, the
`/myflow-do` handoff prints the run instructions directly — nothing written to disk, nothing
committed.

The human gate at `IN_PROGRESS` is unchanged in substance: the operator still reviews the staged
diff and still runs the apps. What goes away is the generated file, its checkboxes, and the two
guard signals derived from them.

Specifically:

- The `myflow-manual-test-guide` capability is removed in full.
- `myflow-handoff-output` gains one requirement governing the run instructions the `IN_PROGRESS`
  handoff must carry, absorbing the removed capability's no-runnable-application case.
- `/myflow-do` section 6 becomes **Resolve the run instructions**, keeping the absolute-path and
  per-worktree URL resolution rules it already carries and writing no file.
- The `IN_PROGRESS` handoff template in `handoff-blocks.md` replaces its `Test guide:` line with a
  `Run it:` section.
- The planning paths `/myflow-do` never stages drop from three to two: `openspec/` and
  `docs/superpowers/`.
- `check-unfinished-work.sh` drops from four signals to two — plan checkboxes and open panel
  findings.
- The second commit's subject becomes `chore(<name>): plan and session records`, and
  `gather-self-review-context.sh` matches both wordings so changes already in history keep resolving.
- `docs/manual-test/` and its fourteen committed guides are deleted.

Out of scope: the archived changes under `openspec/changes/archive/`, which record history and are
never rewritten, and `check-vocabulary.sh`'s banned-vocabulary allowlist, whose `manual-test`
entries ban retired command names rather than this artifact.

## Capabilities

- `myflow-manual-test-guide` — removed in full
- `myflow-handoff-output` — the run-instructions requirement, and the guide references in three
  existing requirements
- `myflow-state-machine` — what `/myflow-do` produces, and what `IN_PROGRESS` means
- `myflow-finish-cleanup` — the unfinished-work signal list
- `myflow-command-surface` — the staging scenario and the guide-generation scenario

## Impact

Fourteen contract, skill, command, rule and documentation files carry the guide's name and change.
Two shell scripts change behaviour (`check-unfinished-work.sh`, `gather-self-review-context.sh`) and
four test harnesses change with them. A change in flight when this lands keeps working: the guard's
surviving signals are unaffected, and the commit-subject grep matches both wordings.
