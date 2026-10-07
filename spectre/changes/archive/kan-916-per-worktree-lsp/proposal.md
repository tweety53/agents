# kan-916-per-worktree-lsp

## Why

Code intelligence (gopls, kotlin-lsp) is useless on `/flow` worktrees — which is where all
implementation happens:

- Claude Code (2.1.292) drops every location result (references, definition, implementation,
  workspace symbol) whose path `git check-ignore`, run from the session cwd, reports ignored.
  `<repo>/.worktrees/` is ignored, so every worktree result vanishes. Measured: hover in
  `.worktrees/flow-verify-auto-fix/stats/health_test.go` answers; findReferences on the same symbol
  returns nothing, while the main-checkout copy returns 2 references.
- Claude Code starts one server per language, rooted at the launch cwd (`workspaceFolders: false`).
  kotlin-lsp imports only the Gradle project at that root — the main checkout on `develop`/`main`
  — so worktree Kotlin files do not resolve at all.

## What changes

- `/flow` worktrees live at `<parent-of-main-checkout>/<repo_name>-worktrees/<name>`, outside the
  repository; `check-worktree-location.sh` (ported to Go first) enforces exactly that layout.
  Outside the repository `git check-ignore` exits 128, so Claude Code filters nothing.
- Every existing `.worktrees/` worktree in agents, gymie, gymie-frontend, gymie-admin-frontend and
  gymie-playwright is moved to the new layout, its flow state record rewritten; a worktree a process
  holds is skipped and reported.
- A `worktree-lsp` wrapper (Go) replaces the official `gopls-lsp`/`kotlin-lsp` plugins: one real
  server per git worktree, started on first use, rooted there; a request waits until that server
  has finished indexing; workspace symbol search fans out to every running server; a worktree's
  servers stop at `/flow` cleanup and when its directory disappears.
- `/flow` and `/flow-fast` kickoff wait for the new worktree's servers to finish indexing and record
  the index time; cleanup stops them before its process check.
- Delivered as a skills-dir plugin (`mods/worktree-lsp/`), linked by `setup.sh global` like every
  other mod; `setup.sh` warns while the official LSP plugins are still enabled.

## Fix 1 — in-run, flow.verify (2026-10-07)

The live migration moved this change's own worktree, which holds the running guard's scripts, and
every later process check then failed: `migrate-worktrees` re-points its scripts directory at the
moved path after each move, so one run migrates every repository named.
