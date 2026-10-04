# Sync the branch onto the base

**Loaded by `skills/flow/integrate.md`'s step 2 only when** `check-base-moved.sh` reported `MOVED`
for a worktree, or by run 1's merge-and-push route when its push was rejected and the sync is re-run. `/flow-fast` and the review panel's fix round cite its **Conflict** bullet.

**Runs after the base-moved check and before the landing question, on every route — so what
lands is what was verified, and neither a merge nor a PR ever meets a conflict.** Once per
worktree in the resolved set whose verdict was `MOVED` — never one whose verdict was `CLEAR`, which
already sits on the tip:

```bash
sync-onto-base.sh <worktree> $BASE <recorded-merge-base>
```

The guard sets the run's own uncommitted planning artifacts aside around the rebase, rebases onto
the base's tip, and re-checks until the base holds still; its header
(`<agents repo>/scripts/sync-onto-base.sh`) is canonical for the exit contract. On every
stop-and-ask, an aside still set aside is named in the handoff with `git stash list` as the
recovery path.

`origin/$BASE` is current: `resolve-base-branch.sh` fetched when the caller resolved the base ref,
and `check-base-moved.sh` performs no fetch of its own.

- **Clean** (exit 0, `REBASED: <worktree> — onto <sha>`): the merge base carried forward for the
  rest of **this run** becomes that `<sha>` — `<rebased-merge-base>`, a this-run-only value
  never written to the state file, which every later reader of this worktree's recorded merge
  base in this run means instead — most concretely the reshape in **Run 1 — the branch is not merged**
  (`skills/flow-contracts/finish-contract-run1.md`). Run **Scoped re-verification** below, then
  proceed to the landing question.
- **Conflict** (exit 1, `CONFLICT: <worktree> — onto <sha>; unmerged: <paths>`): **resolve it in place, automatically.** For every path
  `git status` lists as unmerged, read both sides and write the file that keeps the upstream
  change *and* this change's intent, with no conflict marker left; `git add` it; then
  `git -C <worktree> rebase --continue`, repeating for every commit the rebase stops on until it
  finishes.
  Then run `sync-onto-base.sh --resume <worktree> $BASE <sha>`, `<sha>` from the `CONFLICT:` line;
  it restores the aside and re-checks, and its exits read as the first call's.
  Never `rebase --abort`, never `rebase --skip`, never `-X ours`/`-X theirs`, and never a
  resolution that drops one side wholesale — a conflict is two changes to the same lines, and both
  ship. **Stop and ask** only where no honest resolution exists: a modify/delete conflict, a binary
  file, or an upstream commit that removed something this change depends on. In that case leave
  the worktree mid-rebase exactly as `git rebase` left it, report the file(s) and why, and hand off
  `git -C <worktree> rebase --continue` (after the operator resolves it) or
  `git -C <worktree> rebase --abort` as the next manual step; the run stops before the landing
  question, exactly as a `REFUSE` does.
- **Stop and ask** (exit 2): a `REFUSE`, a base still moving after three rebases, or anything the
  guard could not answer — report its output and stop before the landing question.
- **After a rebase that needed resolution**, run the project's whole `## lint` and `## test` lists
  (`<project>/.flow/project.md`) — a hand-merged hunk is code nobody verified — and stop on a
  failure before the landing question, leaving the worktree rebased. A clean rebase runs only what
  the calling stage says it runs. If this change's verification compares against a recorded
  baseline, recapture it now — a proof taken against the pre-rebase base is void.
- **The handoff names every file that conflicted and what its resolution kept.** A rebase that
  changed nothing is reported as such, not omitted.

**Scoped re-verification**: run the test every `GUARD-TEST: <path> — <test>` line names.
A `NO-GUARD-TEST: <path>` is
stated in the handoff as having no verification to run, not silently skipped. A non-zero exit from
any discovered test blocks this stage exactly like any other verify-stage failure: report it, leave
the change `IN_PROGRESS`, stop before the landing question, and close the mark `stopped`. **Never**
re-run the project's whole `## lint`/`## test` list here. A clean rebase whose overlap set clears
this stage proceeds to the landing question and closes the mark `completed`.
