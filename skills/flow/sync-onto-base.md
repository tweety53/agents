# Sync the branch onto the base

**Loaded by `skills/flow/integrate.md`'s step 2 only when** `check-base-moved.sh` reported `MOVED`
for a worktree. `/flow-fast` and the review panel's fix round cite its **Conflict** bullet.

**Runs after the base-moved check and before the landing question, on every route — so what
lands is what was verified, and neither a merge nor a PR ever meets a conflict.** Once per
worktree in the resolved set whose verdict was `MOVED` — never one whose verdict was `CLEAR`, which
already sits on the tip:

```bash
git -C <worktree> rebase origin/$BASE
```

The rebase never meets the run's own uncommitted planning artifacts:
`aside-planning-artifacts.sh <aside|restore> <worktree>` around each `MOVED` worktree's rebase —
set aside before it, restored once that worktree's rebase has finished or aborted; never
mid-way, restore refuses while the rebase is still unresolved, and on a stop-and-ask exit the
aside stays set aside, named in the handoff with `git stash list` as the recovery path. The
helper sets the planning paths aside — the spec tree's changes directory (the leaf
`<agents repo>/scripts/lib/spec-root.sh` resolves) and `<project>/docs/superpowers/` — and
nothing else: implementation WIP stays exactly where the unfinished-work gate owns it. A
clean rebase runs **Scoped re-verification** below, then proceeds to the landing question.

`origin/$BASE` is current: `resolve-base-branch.sh` fetched when the caller resolved the base ref,
and `check-base-moved.sh` performs no fetch of its own.

- **Clean** (exit 0): the merge base carried forward for the rest of **this run** becomes
  `origin/$BASE`'s resolved tip at rebase time — `<rebased-merge-base>`, a this-run-only value
  never written to the state file, which every later reader of this worktree's recorded merge
  base in this run means instead — most concretely the reshape in **Run 1 — the branch is not merged**
  (`skills/flow-contracts/finish-contract-run1.md`). Re-run
  `check-base-moved.sh` once against it; anything but `CLEAR` is a base that moved during the
  rebase, and the rebase runs once more.
- **Conflict** (non-zero exit): **resolve it in place, automatically.** For every path
  `git status` lists as unmerged, read both sides and write the file that keeps the upstream
  change *and* this change's intent, with no conflict marker left; `git add` it; then
  `git -C <worktree> rebase --continue`, repeating for every commit the rebase stops on until it
  finishes. Never `rebase --abort`, never `rebase --skip`, never `-X ours`/`-X theirs`, and never a
  resolution that drops one side wholesale — a conflict is two changes to the same lines, and both
  ship. **Stop and ask** only where no honest resolution exists: a modify/delete conflict, a binary
  file, or an upstream commit that removed something this change depends on. In that case leave
  the worktree mid-rebase exactly as `git rebase` left it, report the file(s) and why, and hand off
  `git -C <worktree> rebase --continue` (after the operator resolves it) or
  `git -C <worktree> rebase --abort` as the next manual step; the run stops before the landing
  question, exactly as a `REFUSE` does.
- **After a rebase that needed resolution**, run the project's whole `## lint` and `## test` lists
  (`<project>/.flow/project.md`) — a hand-merged hunk is code nobody verified — and stop on a
  failure before the landing question, leaving the worktree rebased. A clean rebase runs only what
  the calling stage says it runs. If this change's verification compares against a recorded
  baseline, recapture it now — a proof taken against the pre-rebase base is void.
- **The handoff names every file that conflicted and what its resolution kept.** A rebase that
  changed nothing is reported as such, not omitted.

**Scoped re-verification**: for each path
`check-base-moved.sh` reported under `overlaps:`, look for a discoverable guard test —
`<agents repo>/scripts/test-<basename-without-ext>.sh` beside `<agents repo>/scripts/<name>.sh`,
the same naming `<agents repo>/scripts/run-guard-tests.sh` already discovers by glob — and run it
if found. A path with none is
stated in the handoff as having no verification to run, not silently skipped. A non-zero exit from
any discovered test blocks this stage exactly like any other verify-stage failure: report it, leave
the change `IN_PROGRESS`, stop before the landing question, and close the mark `stopped`. **Never**
re-run the project's whole `## lint`/`## test` list here. A clean rebase whose overlap set clears
this stage proceeds to the landing question and closes the mark `completed`.
