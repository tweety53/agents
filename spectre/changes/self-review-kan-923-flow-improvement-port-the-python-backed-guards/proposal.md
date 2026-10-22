# self-review-kan-923-flow-improvement-port-the-python-backed-guards

## Why

kan-923's self-review pass found five defects, none `big`:

- **F-A (flow-fix).** `scripts/check-plan-shape.sh`, `scripts/check-task-build-green.sh` and
  `scripts/plan-dispatch-bundles.sh` each hand-roll their repository root through
  `lib/resolve-file.sh` and a `SELF_REAL` walk, where `flow_guard_root` (`scripts/lib/flow-guard.sh`)
  already resolves the true root through any skill's `scripts/lib` symlink and is what
  `plan-class.sh`, `sync-onto-base.sh` and `check-loop-landing.sh` use.
- **F-B (flow-fix, flow-automation).** `check-plan-unchanged verify` compares the plan tree's
  `git status --porcelain` (`stats/internal/guard/planunchanged.go`), so the parent's own
  `flow tasks tick` while a reviewer is in flight reads as the reviewer changing the plan; it
  false-fired twice in kan-923's run and was cleared by hand-diffing.
- **F-C (flow-fix).** `tcfBodyFields` (`stats/internal/guard/planparse.go`) keeps a repeated field's
  last occurrence, so a later prose line opening `**Files:**` replaces the task's real `**Files:**`
  value (kan-923: `Files=[") and so on."]`).
- **F-D (flow-speed).** Five view tests find a `Panel` region, which renders while loading
  (`stats/web/src/components/Panel.tsx`), then read its loaded content synchronously; a slow
  fixture fails them (kan-923's failed verify log, `test-2.log`).
- **F-E (flow-cost).** `write-panel-diff` writes a deleted file's whole body into the panel diff;
  kan-923's `final-review.diff` was 17,735 lines, about 10.8k of them deleted-file bodies, read by
  every pass-1 slot.
  <!-- measured: wc -l .superpowers/sdd/final-review.diff by the self-review pass @ 5611091f, machine-local and gitignored, not re-runnable -->

## What changes

- The three shims derive `FLOW_GUARD_REPO_ROOT` with `flow_guard_root`, as the other shims do.
- `check-plan-unchanged` snapshots the plan tree's content, and verify passes a difference made
  only of `- [ ]` → `- [x]` checkbox flips.
- A repeated task field keeps its first occurrence.
- The five view-test sites await the loaded content before reading it.
- Panel diffs omit a deleted file's body (`--irreversible-delete`); reviewers read it with
  `git show <sha>:<path>`.
