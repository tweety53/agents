# kan-923-flow-improvement-port-the-python-backed-guards — session narrative

## 2026-10-22 — creating run

- **Shim template defect, found by task 1.** The plan's verbatim template derived
  `FLOW_GUARD_REPO_ROOT` from the invoked path, so a shim run through `skills/flow/scripts/` (a
  symlink) scanned `skills/flow` and exited 0 having checked nothing. Task 1's implementer resolved
  the real path through `lib/resolve-file.sh`, as the old wrapper did; the fix was transcribed as a
  dated preamble Correction and every later port copied it.
- **Tracked `.py` skill symlinks.** Deleting each `.py` left a dangling
  `skills/*/scripts/<name>.py` symlink that fails `check-guard-symlinks.sh` rule 1 — not foreseen
  by the plan; tasks 1, 4, 5, 6 and 8 each widened `**Files:**` to remove theirs.
- **`**Files:** none` reads differently in Go and Python.** `tcfParseTask` reads it as no files,
  the Python as one path `none`; reusing the Go reading made F2 fire on 15 archived plans that pass
  at base. Kept the Python reading through `tcfPythonFiles`.
- **A Correction paragraph broke the fields guard.** A wrapped sentence put `**Files:**)` at the
  start of a line inside task 5's Correction, and `check-task-commit-fields` read it as the task's
  Files field, refusing the commit. Rewording the paragraph so no line opens with a field label
  fixed it — a Correction paragraph must never start a line with `**<Field>:**`.
- **design.md premise disproved.** `check-task-records.sh` took its root from `${1:-$PWD}`, not its
  own location; the shim drops the repo-root lines and design.md's Context was corrected.
- **Plan-tree guard false positives.** `check-plan-unchanged.sh verify` reported a changed plan
  twice; both times the only diff was the parent's own `flow tasks tick` made while a reviewer
  ran. Confirmed by diffing, not a reviewer mutation.
- **Panel Minor M1 overrode byte parity.** `generate-relocation-comparison`'s stderr label kept the
  deleted `.py` name for parity; no consumer reads it, so the inline Minor fix renamed it.
- **Verify flake.** `stats/web` `npm test` failed once on `src/views/views.test.tsx:383` and passed
  on its one inline re-run; this change touches nothing under `stats/web`.
- **Install path.** `~/.claude/skills/flow` points at the main checkout, which still runs the base
  Python until this lands, so the live check ran the worktree's own `skills/flow/scripts/` symlinks.
