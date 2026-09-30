# Design — basesync group (KAN-860) — WIP, no code landed

Stopped at the coordinator's request (usage limit) before any code or prompt edit landed. What is
here is the design worked out against the tree at 452e31d7, so a follow-up session can build it.

## Rows

MX6 (implement.md), RP34 (review-panel.md) and the finish "Syncing onto the base" row
(sync-onto-base.md). All three are **left, not started**: no Go code, no shim and no prompt edit
exist yet. No prompt bytes changed and there are no verbatim-moves lines.

## Planned shape (one shared Go core)

1. **Port `scripts/aside-planning-artifacts.sh` to Go first**
   (`stats/internal/guard/asideplanningartifacts.go`, shim via `flow_guard_exec`). Today it is
   bash only, and the convention is to port byte for byte before extending. Parity: run the retired
   bash at `452e31d7:scripts/aside-planning-artifacts.sh`, with its `lib/spec-root.sh`, on the
   same fixtures as the Go port. `scripts/test-aside-planning-artifacts.sh` keeps passing against
   the shim. Details to keep: git's stash chatter goes to stderr; the `${sha:0:12}` cut; the
   dir/git checks come before the unknown-action usage; the both-spec-roots warning prints `$WT` as
   given.
2. **Refactor `basemoved.go`** into `baseMoved(env, wt, ref, recorded, stderr) bmVerdict`
   (code, line, kind CLEAR/MOVED/REFUSE, ref, full sorted overlap). `checkBaseMoved` prints
   `v.line`, and its contract is unchanged. Callers get the full overlap, not the 10-path cut the
   line shows.
3. **`baserebase.go` core** `rebaseOntoTip(env, wt, ref, aside bool, …)`:
   - Resolve the tip sha, and rebase onto that sha, so a concurrent fetch cannot move it.
   - With `aside`: in-process aside → `git rebase <sha>` (all output to stderr).
   - Exit 0 → restore → `rebased onto <sha>`.
   - Non-zero with rebase-merge/rebase-apply present → `conflict` with the
     `diff --name-only --diff-filter=U` paths. The aside stays.
   - Non-zero with no rebase in progress → `refused`, the aside restored. Unlike the old prose,
     this is not reported as a conflict.
4. **`sync-onto-base.sh [--resume] <worktree> <base-ref> <merge-base>`**. It uses the aside.
   - It echoes each check-base-moved line.
   - CLEAR → `CLEAN: <wt> — nothing to rebase` (exit 0).
   - MOVED → the core, then a re-check against the new tip, looping while the re-check is MOVED
     (cap 3) → `REBASED: <wt> — onto <sha>` (exit 0). Before that line it prints
     `GUARD-TEST: <path> — <test>` / `NO-GUARD-TEST: <path>` for every overlap path: the
     `scripts/test-<stem>.sh` beside `scripts/<stem>.sh` in the agents repo. The shim exports the
     root as `cd -P "$(dirname -- "${BASH_SOURCE[0]}")/lib" && cd ../..`, the way flow-guard.sh
     finds stats/, which is rule-4 safe.
   - `CONFLICT: <wt> — onto <sha>; unmerged: <paths>` (exit 1): the worktree is left mid-rebase
     and the prose resolution rule applies.
   - Exit 2 on REFUSE, a cannot-answer, or a restore conflict (stop and ask).
   - `--resume <wt> <base-ref> <onto>` runs after the in-place resolution finishes. It refuses
     while a rebase is still in progress and requires `<onto>` to be an ancestor of HEAD. It then
     restores the aside and runs the same re-check loop, with no guard-test lines (the whole
     lint/test run applies instead).
5. **`sync-panel-base.sh [--rebase] <worktree> <working-notes-merge-base>`** (RP34).
   - It runs `resolve-base-branch` in-process (the fetch) → `BASE: <name>` → check-base-moved.
   - It rebases automatically only on MOVED with no overlap, or with `--rebase` (the operator's
     **Rebase**). It passes **no aside**, to match the panel prose, which prescribes a bare
     `git rebase`; adding the aside there would be a behaviour change and is left as a follow-up.
   - It re-checks with the same loop, stopping on a fresh overlap so the prompt is re-offered.
   - It prints `REBASED: <wt> — merge base <sha>`, the new working-notes merge base.
   - `CONFLICT` exit 1 means never resolve; hand off. Exit 2 means stop and ask.
6. **`refresh-plan-base.sh <worktree> <merge-base> <tasks.md> [<spec-path>…]`** (MX6).
   - `resolve-base-branch` fetches, then it counts `mb..origin/<base>`.
   - UNMOVED gets one line.
   - MOVED gets the `git diff --name-only mb origin/<base>` names, intersected in Go with the
     `**Files:**` backticked paths: an exact or leading-directory match, which avoids git's
     outside-repository pathspec error on multi-repo plans. Each named spec is printed at
     `origin/<base>` between delimiter lines, or reported absent there.
   - The collision judgment and "never rebases" stay prose.

The prompt edits follow each call line with its exit contract. Each needs a verbatim-moves block,
symlinks in `skills/flow/scripts/`, and Go table tests with git fixtures (set `user.name`/`email`
in the fixture repos, since the guards' git runs under the process environment).
