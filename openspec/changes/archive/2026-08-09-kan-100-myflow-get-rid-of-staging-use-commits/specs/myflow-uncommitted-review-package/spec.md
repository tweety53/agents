## REMOVED Requirements

### Requirement: A checkpoint isolates uncommitted work without touching any ref

**Reason**: `/myflow-do` now commits each task before review instead of leaving it uncommitted, per
the new requirement **Each task commits after finishing, before review** in `myflow-task-commits`.
A checkpoint snapshot is no longer needed once work is genuinely committed.

**Migration**: `skills/myflow-do/scripts/checkpoint` is deleted. Review isolation now comes from git
history directly — each task's own commit range.

### Requirement: A per-task or per-fix-round review package diffs BASE against the live working tree

**Reason**: Review now reads `git diff <task-base>..<task-sha>` directly from real commits, per the
new requirement **Review reads a real commit diff, not a checkpoint snapshot** in
`myflow-task-commits`.

**Migration**: `skills/myflow-do/scripts/uncommitted-review-package` is deleted. A reviewer is given
a commit range instead of a generated package file.

### Requirement: Invalid arguments fail the same way review-package's do

**Reason**: The script this requirement governs is deleted.

**Migration**: None needed — no replacement script exists to carry this argument-validation
behavior forward.
