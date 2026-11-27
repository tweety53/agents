# Sync the branch onto the base

**Loaded by `skills/flow/integrate.md` only when** `flow-guard integrate-change` stops on `sync-conflict` or
`sync` — in its base-moved sync, or the merge-and-push route's re-sync after a rejected push; the
script runs the clean sync itself. The review panel's fix round, the panel's base sync
and the implement phase's pre-task-1 sync cite its **Conflict** bullet.

**Runs after the base-moved check and before the landing question, on every route — so what
lands is what was verified, and neither a merge nor a PR ever meets a conflict.** Once per
worktree in the resolved set whose verdict was `MOVED` — never one whose verdict was `CLEAR`, which
already sits on the tip:

```bash
flow-guard sync-onto-base <worktree> $BASE <recorded-merge-base>
```

The guard sets the run's own uncommitted planning artifacts aside around the rebase, rebases onto
the base's tip, and re-checks until the base holds still; its header
(`<agents repo>/scripts/flow-guard sync-onto-base`) is canonical for the exit contract. On every
stop-and-ask, an aside still set aside is named in the handoff with `git stash list` as the
recovery path.

`origin/$BASE` is current: `flow-guard resolve-base-branch` fetched when the caller resolved the base ref,
and the base-moved check performs no fetch of its own.

- **Clean** (exit 0, `REBASED: <worktree> — onto <sha>`): the merge base carried forward for the
  rest of **this run** becomes that `<sha>` — `<rebased-merge-base>`, a this-run-only value
  never written to the state file, which every later reader of this worktree's recorded merge
  base in this run means instead — most concretely the reshape in **Run 1 — the branch is not merged**
  (`skills/flow-contracts/finish-contract-run1.md`). Then
  proceed to the landing question. Between the two, the rebased tree runs the project's whole
  `## lint` list (`flow-guard run-verify-list <worktree> lint`, under the `KEY=value` lines
  `flow-guard prepare-workspace` prints for it): a textually clean rebase can still break the build, as one
  that dropped a Go import did, so a failure is `STOP: rebase-lint`, and so is a passing
  list that rewrote tracked files,
  leaving the worktree rebased.
- **Conflict** (exit 1, `CONFLICT: <worktree> — onto <sha>; unmerged: <paths>`): **resolve it in place, automatically.** For every path
  `git status` lists as unmerged, read both sides and write the file that keeps the upstream
  change *and* this change's intent, with no conflict marker left; `git add` it; then
  `git -C <worktree> rebase --continue`, repeating for every commit the rebase stops on until it
  finishes.
  Then run `flow-guard sync-onto-base --resume <worktree> $BASE <sha>`, `<sha>` from the `CONFLICT:` line;
  it restores the aside and re-checks, and its exits read as the first call's.
  Never `rebase --abort`, never `rebase --skip`, never `-X ours`/`-X theirs`, and never a
  resolution that drops one side wholesale — a conflict is two changes to the same lines, and both
  ship. **Load `skills/flow/sync-onto-base-visual.md` only when** the guard prints a `RECAPTURE:` line,
  or the conflict is in a UI file or a spec's screenshot baseline.

  A modify/delete conflict on a file this change deleted and carries a replacement for is resolved
  without asking: the upstream edit's intent is carried into the replacement, or shown to need
  nothing there, the file stays deleted, and the handoff names it.
  **Stop and ask** only where no honest resolution exists: any other modify/delete conflict, a binary
  file no `RECAPTURE:` line names, or an upstream commit that removed something this change
  depends on. In that case leave
  the worktree mid-rebase exactly as `git rebase` left it, report the file(s) and why, and hand off
  `git -C <worktree> rebase --continue` (after the operator resolves it) or
  `git -C <worktree> rebase --abort` as the next manual step; the run stops before the landing
  question, exactly as a `REFUSE` does.
- **Stop and ask** (exit 2): a `REFUSE`, a base still moving after three rebases, or anything the
  guard could not answer — report its output and stop before the landing question.
- **After a rebase that needed resolution**, run the project's whole `## lint` and `## test` lists
  (`<project>/.flow/project.md`) once, after the last resolution round, when the guard's re-check
  finds the base holding still — never once per conflict round — since a hand-merged hunk is code nobody verified — and stop on a
  failure before the landing question, leaving the worktree rebased. A clean rebase runs only what
  the calling stage says it runs. If this change's verification compares against a recorded
  baseline, recapture it now — a proof taken against the pre-rebase base is void.
- **The handoff names every file that conflicted and what its resolution kept.** A rebase that
  changed nothing is reported as such, not omitted.
