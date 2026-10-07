# kan-916-per-worktree-lsp — session narrative

## 2026-10-07 — creating run

**Shape.** The run resumed at `STARTED` (plan in `3d56584c`) and executed five implementer groups
in waves, each in a throwaway `<wt>-wave-group-<g>` worktree whose commits were cherry-picked onto
the change branch. Group 4 (task 8) was decided sonnet but dispatched on opus: it adds a guard
symlink and a `check_guard_symlinks_test.go` row, i.e. a Go test whose result was in doubt.

**Design moves during the run.** `plugin-as-mod-go-run` was superseded by
`plugin-as-mod-build-exec`: under `go run` the plugin's pid is `go`, not the wrapper, so a
non-graceful stop never reached the wrapper and orphaned its servers; the shim now builds into
`${XDG_CACHE_HOME:-$HOME/.cache}/worktree-lsp/<cksum>/` and `exec`s the binary. `migrate-all-now`
was superseded by `migrate-before-preflight` in panel round 1: a cross-repo change's record names
another repository's worktree, which only a run over every repository can rewrite, and run 1's
strict preflight must already see migrated worktrees — so integrate migrates every main checkout
of the change before `check-finish-preflight.sh`, and run 2 no longer migrates.

**The board mod reads `$.session.root()`.** A session launched inside a worktree resolves its root
to that worktree, so `<root>-worktrees/<change>` names nothing there; that launch shape is
unsupported, as the main-checkout hook already treats it.

**lspmux deadlock.** The first mux wrote to a child's stdin under the mux lock; a server that stops
reading its stdin (gopls mid-index) then stalled every root. Fixed by one writer goroutine per child
draining an outbox (`cc75add5`), proven with the fake server's `flood=<method>` option. Named and
left: a live server that never reads its stdin grows that outbox without bound — every bound found
drops messages the client is owed.

**Time sinks.**
- `git checkout -- <file>` after a mutation twice reverted uncommitted fixes of my own
  (`removechangeworktrees.go`'s restructure, `ready.go`'s document forget) — both redone. The fix
  commit must land before the first flip; the MUTATION PROOF paragraph already says so.
- One "surviving" mutant was a compile failure (an unused variable); redone as `stopped || true`.
- A zsh loop variable named `path` clobbered `PATH` and broke a whole Bash call.
- An already-pushed commit was amended by mistake; repaired with `git reset --soft` and a new commit
  on top (`f99e96a1`), never a force-push.
- A subagent claimed the excludes staging pattern stages nothing for new files; measured in a
  scratch repository it staged fine, so no pipeline change was made.

**Panel.** Round 0 raised F1–F4 (F1, F2, F4 Important; F3 Minor); the fix round routes for F1 and
F2 were auto-decided on their recommended options. The sonnet re-review read the round clean.

**Live check — in-run fix 1.** The migration, run from this change's own worktree, moved that
worktree and then could not run `check-worktree-processes.sh` for any later worktree: every gymie
worktree came back `FAILED … could not answer`, exit 1. Fixed as task 10 (re-point the scripts
directory after each move), reviewed clean by the gated task reviewer and the late-fix `primary`
read, and the migration re-run exited 0.

**Live check — Kotlin.** The first gymie worktree tried, `ci-only-on-promotion` (a 2026-09-21
branch), answered hover and documentSymbol but nothing cross-file. The official `kotlin-lsp` plugin
launched inside the same worktree returned the same empty figures, so the wrapper was not the
cause; a fresh `develop` worktree (created for the probe, removed after) answered all nine
operations identically to the main checkout. Also seen: one Go `server is starting` error from
Claude Code on a first call (retry answered), and the kotlin server taking up to 30 seconds to exit
after `claude -p` returned, during which `check-worktree-processes.sh` prints `HELD`.

**Suite records.** The verify run exported the worktree's isolated `FLOW_ADDR`
(`127.0.0.1:4373`), where no daemon runs, so `flow suite record` dropped its rows with a warning;
exit codes were unaffected.

## 2026-10-07 — integrate run

**Preflight.** `check-foreign-staged.sh` was clean, but `check-main-checkout-drift.sh` reported
`DRIFT-DIRTY` on the main checkout: four uncommitted files lowering the implementer in-flight cap
per wave from three to two (`scripts/plan-dispatch-groups.py`, its test, `brainstorm-planner.md`,
`sdd-dispatch.md`) — unrelated to this change. The operator chose to verify and commit it. It was
moved into a throwaway worktree off `origin/main`, its test and `check-references.sh` passed, and it
landed on `main` as `670bd7ad`. Fast-forwarding the main checkout afterwards was refused by git,
because its identical uncommitted copies would be overwritten; the main checkout was left for the
operator to reset. Two auto-mode permission denials (the pull and the base-resolving fetch) paused
the run until the operator granted them.

**Gates.** `check-finish-preflight.sh` returned `RUN1`; `check-unfinished-work.sh` returned `CLEAR`;
`check-visual-verify-dispatched.sh` returned `VISUAL-VERIFY-OK` (no UI paths).

**Rebase.** `check-base-moved.sh` returned `MOVED` — one commit (`670bd7ad`) overlapping
`skills/flow/brainstorm-planner.md`. `sync-onto-base.sh` rebased all 40 commits cleanly onto it
with no conflict; the overlap has no guard test (`NO-GUARD-TEST`). The landing route came from the
project's configured default, merge and push.
