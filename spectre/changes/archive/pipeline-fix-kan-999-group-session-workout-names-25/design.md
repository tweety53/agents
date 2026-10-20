## Context

Defect found mid-run in gymie kan-999's `/flow` run 2 (cleanup): `scripts/remove-change-worktrees.sh`
check 4 → exit 3, relayed by `scripts/integrate-change.sh:873` as `STOP: disclose`, on
`UNCLASSIFIED: <worktree> — .env` and `REGENERATABLE: … — 6236`. The parent then asked the operator
whether to delete the worktree. Check 4's contract (`skills/flow-contracts/finish-contract-run2.md`)
asks only about an irreplaceable, unpreserved entry; a worktree's `.env` is neither — it is the
per-worktree copy of `/Users/tweety53/Projects/gymie/.env`, the same keys with the worktree's
isolated ports and `ALLOWED_ORIGINS`, which the next worktree's setup writes again.

Cause: `rcwIdentical` (`stats/internal/guard/removechangeworktrees.go`, since `e6931c42`, kan-924)
preserves only an entry byte-identical to the main checkout's. A per-worktree config copy differs by
construction, so it always fell through to unpreserved.

Reproduced at `94818c76`: the guard test fixture with the worktree's `.env` `PORT=3100` and the main
checkout's `PORT=3000` exits 3 with `UNCLASSIFIED: … — .env`; the existing subtest
"an entry that differs from the main checkout's stops" encodes exactly that behaviour.
<!-- measured: cd stats && go test ./internal/guard -run TestRemoveChangeWorktrees -count=1 @ 94818c76, with Task 1's test hunk alone → FAIL: TestRemoveChangeWorktreesDisclose/an_entry_that_differs_from_the_main_checkout's_proceeds -->

History: `e6931c42` introduced the byte-identical rule for kan-924's identical `.env`; no archived
decision rejects a counterpart rule — kan-924 recorded none for the differing case.

Blast radius: 6 files outside archives cite the names this fix changes (`rcwIdentical`,
<!-- measured: git grep -l -e rcwIdentical -e 'preserved: identical' -- . ':!spectre/changes/archive' | wc -l @ 94818c76 → 6 -->
`preserved: identical`): the Go guard and its test, the shell shim's header, `finish-contract-run2.md`,
and `spectre/changes/faster-integrate/{proposal.md,verbatim-moves.txt}` — the last two another
change's own record, left as written. The fix touches four of them plus this change's
`verbatim-moves.txt`. Not `big`. No saved memory encodes the old behaviour
(`grep -rl -e UNCLASSIFIED -e disclose -e remove-change-worktrees ~/.claude/projects/*/memory/` finds none).

## Decisions

### What makes an unclassified entry preserved

**ID:** counterpart-in-main-checkout
**Status:** active
**Chosen:** a regular file in the worktree whose same path in the main checkout is a distinct regular file — identical or not — is preserved. It covers every per-worktree copy of a main-checkout local config (`.env`, a local override) with one existence check and no name list, and a file with no counterpart there still stops.
**Considered:** a name allowlist (`.env*`) — `finish-hand-fallbacks.md` trusts no allowlist of names beyond `.DS_Store`, and a `.env.local` or `local.properties` would stop the next run; comparing the key sets of env-style files — parses a format per file type for the same verdict the existence check reaches; keeping byte-identity and letting the parent auto-proceed — moves a mechanical verdict into prose the next run can misread, which is how this ask happened.

### The preserved line names whether the copy differs

**ID:** differs-label-kept-visible
**Status:** active
**Chosen:** the line ends `— preserved: identical in <main-checkout>` when the bytes match and `— preserved: differs from <main-checkout>` when they do not — the operator still sees which worktree copy carried its own content.
**Considered:** one label `— preserved: in <main-checkout>` for both — hides that the worktree's own values (its ports) go with it, and renames a label `faster-integrate`'s record cites.

### Symlinked counterpart stays unpreserved

**ID:** samefile-not-a-counterpart
**Status:** active
**Chosen:** a main-checkout path that reaches the worktree's own file (`os.SameFile`, e.g. a symlinked directory into the worktree) is no counterpart, as before — removal would destroy the only copy.
**Considered:** none — dropping it would preserve a file the removal destroys.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — one predicate in a Go guard widened, one subtest flipped, one contract sentence; the no-counterpart and symlink subtests already pin the cases that must still stop.
**Considered:** `flow-medium` — the guard gates destruction of ignored files, and the change preserves only entries at the worktree root with a main-checkout counterpart, even when they differ; nested entries, entries with no counterpart and symlink-reached entries still stop, each with its own test.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — the plan holds one task.
**Considered:** none — a single task has no split.

## Open questions

## Live check

None needed against a running service: the guard holds no runtime state. From this worktree, with a
throwaway repo whose worktree `.env` differs from the main checkout's, the rebuilt guard exits 0 and
prints `— .env — preserved: differs from …`; the Go subtest is that check. Exit 3 with a bare
`UNCLASSIFIED: … — .env` after the fix means the counterpart rule never ran.
