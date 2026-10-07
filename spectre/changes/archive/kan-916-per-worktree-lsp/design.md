## Context

- Two Claude Code behaviours, read from the installed 2.1.292 bundle, make LSP useless on worktrees:
  1. The LSP tool filters the locations of `findReferences`, `goToDefinition`,
     `goToImplementation` and `workspaceSymbol` through `git check-ignore <paths…>` run with
     `cwd` = the session cwd, and drops every path it prints — only when the command exits 0.
  2. Each plugin LSP server starts once per session, `rootUri`/`workspaceFolders` = the plugin's
     `workspaceFolder` or the session cwd, `capabilities.workspace.workspaceFolders: false`.
     `requestTimeout` (default 60 000 ms) and `startupTimeout` are per-server plugin keys.
     <!-- measured: `SNr=60000` read from the strings of ~/.local/share/claude/versions/2.1.292 — machine-local binary, not re-runnable at a ref -->
- `git check-ignore` on a path outside the repository exits 128 (measured), so a worktree outside
  the repository is never filtered.
- gopls already serves a worktree file from a session rooted elsewhere (hover answered); kotlin-lsp
  does not import a second Gradle project.
- Measured spikes, scratch plugin, 2026-10-07: a plugin's `lspServers` entry whose `command` is a
  wrapper script is started and used, with `requestTimeout` accepted, both from `--plugin-dir` and
  from a plugin folder linked into `~/.claude/skills/` (the mods mechanism); the official
  `gopls-lsp` was disabled through `enabledPlugins`. `claude -p --allowedTools LSP` runs the LSP tool
  headless — the live check's harness.
- `check-worktree-location.sh` is bash; project rule: a bash guard is ported to Go, byte-for-byte
  parity, before it is extended.
- flow state records key `worktrees` by absolute path, and `state set` replaces the map wholesale
  (`stats/internal/store/changes.go` upsert, `worktrees = EXCLUDED.worktrees`).
- One change: the layout move is what makes the wrapper's results survive, and the wrapper is what
  makes Kotlin resolve; neither delivers the operator's outcome alone.

## Decisions

### Worktree location — sibling `<repo>-worktrees/`, strict

**ID:** sibling-worktrees-layout
**Status:** active
**Chosen:** `<dirname main>/<basename main>-worktrees/<name>`; the guard accepts exactly that
directory and below, everything else is `STRAY` — the operator's choice; outside the repository
nothing is filtered, and nothing inside the repository needs ignoring.
**Considered:** keep `.worktrees/` and un-ignore it — worktrees show in `git status` and `git add -A`
commits them as embedded repositories; `.git/info/exclude` instead of `.gitignore` — `git
check-ignore` honours it, so results are still dropped.

### Port the location guard before changing it

**ID:** port-location-guard-first
**Status:** active
**Chosen:** port `check-worktree-location.sh` to `flow-guard` at byte parity, then change its rule —
the project's standing rule for extending a bash guard.
**Considered:** edit the bash in place — forbidden by `.flow/project.md` `## apps`.

### Existing worktrees — move them all now

**ID:** migrate-all-now
**Status:** superseded by migrate-before-preflight
**Chosen:** a `migrate-worktrees` guard moves every registered worktree under `<main>/.worktrees/`
with `git worktree move`, skipping one `check-worktree-processes.sh` reports `HELD`, and rewrites
each flow state record that names the old path; run over the five repositories during
implementation and again at cleanup run 2 — the operator's choice.
**Considered:** only agents and gymie — three repos keep failing the strict guard; leave them to
manual cleanup — the guard is red until someone finishes every old change.
**Superseded because:** run 2 is entered only on a `RUN2` preflight verdict, and the preflight
refuses a `.worktrees/` worktree as a stray, so the run 2 migration never ran for the worktree it
exists to move; and a per-repository run never rewrote a cross-repo change's record, which lives in
one project and names another repository's worktree (panel round 1).

### One wrapper per session, a child per worktree

**ID:** wrapper-per-session
**Status:** active
**Chosen:** `worktree-lsp -- <server> <args…>` is the plugin's server command; it is the only server
Claude Code sees and multiplexes children keyed by the file's `git rev-parse --show-toplevel`.
**Considered:** one shared daemon across sessions — two sessions' open-document state collides in
one server; sessions launched inside the worktree — the operator wants nothing about how they
work to change.

### Static initialize, lazy children

**ID:** static-initialize
**Status:** active
**Chosen:** the wrapper answers `initialize` itself with a fixed capability set (full text sync,
hover, definition, references, implementation, document symbol, workspace symbol, call
hierarchy) and starts a child only when a document of its root is first opened or queried.
**Considered:** an eager child at the session root to borrow its capabilities — starts a JVM in
every session, Kotlin or not.

### Readiness — progress tokens settle

**ID:** readiness-progress-settle
**Status:** active
**Chosen:** a root is ready once its child answered `initialize`, every `$/progress` token it began
has ended, and no new token began for a settle window; requests for the root wait until then. The
ready time is appended to `~/.cache/worktree-lsp/index.log`. The plugin raises `requestTimeout` so
the first request can outlast a Kotlin index.
**Considered:** answering immediately — the first queries on a fresh worktree come back empty,
which is the failure the operator asked to wait out.

### Workspace symbol — fan out and merge

**ID:** workspace-symbol-fanout
**Status:** active
**Chosen:** `workspace/symbol` goes to every running child; results concatenated — the operator's
choice; each result's path names its checkout.
**Considered:** the session-root server only — agents in a worktree get main-checkout paths they
must not edit; the most recently used checkout — depends on request order.

### Memory — stop at cleanup and on removal

**ID:** stop-at-cleanup-and-removal
**Status:** active
**Chosen:** `remove-change-worktrees`' Go guard calls `lspmux.StopUnder(<worktree>)` immediately
before its check 6 — killing every process whose cwd is at or under the worktree and whose parent
is a `worktree-lsp` process — so `/flow` and `/flow-fast` cleanup both stop them with no new
command; independently the wrapper stats each root periodically and stops a child whose root is
gone — the operator's choice. `scripts/lib/flow-guard.sh`'s cache key gains
`internal/lspmux/*.go`, so the guard binary rebuilds when that package changes.
**Considered:** a `worktree-lsp stop` subcommand called from the cleanup docs — a second command
and a doc step a run can skip, for what one call inside the guard already covers; cleanup only — a
hand-removed worktree's server lives until the session ends; an idle timeout — a re-index costs
minutes on Kotlin.

### Delivery — a skills-dir plugin, built from the checkout

**ID:** plugin-as-mod-go-run
**Status:** superseded by plugin-as-mod-build-exec
**Chosen:** `mods/worktree-lsp/` carries `.claude-plugin/plugin.json` (`lspServers` `gopls` and
`kotlin-lsp`) and `bin/worktree-lsp`, a shim that `go run`s `stats/cmd/worktree-lsp` from the
checkout it resolves to; `setup.sh global`'s existing `install_mods` links it. The Go build cache
makes every start after the first a link step, and a merged change is live at the next session
with no install step.
**Considered:** a marketplace plugin — copied into the plugin cache, so a merged change needs an
update step; `go build` into `~/.local/bin` from `setup.sh` — a stale binary until someone re-runs
setup, the skew `scripts/lib/flow-guard.sh`'s header records; a `flow-guard` registry entry — a
guard `Func` has no stdin, and flow-guard's cache key does not hash a new package.
**Superseded because:** `go run` does not forward SIGTERM to the program it runs (measured at
review: the wrapper and its gopls outlived a SIGTERM to the `go` pid), so a kill reached the
`go` process and never the wrapper.

### Delivery — a skills-dir plugin that builds and execs the wrapper

**ID:** plugin-as-mod-build-exec
**Status:** active
**Chosen:** as `plugin-as-mod-go-run`, except the shim `go build`s `stats/cmd/worktree-lsp` into
`${XDG_CACHE_HOME:-$HOME/.cache}/worktree-lsp/<checksum of the checkout's stats path>/`, under a
per-process name renamed into place, and `exec`s it — a signal to the plugin's pid reaches the
wrapper. The build cache still makes every start after the first a link step.
**Considered:** keep `go run` and rely on Claude Code's shutdown/exit — every non-graceful stop
leaves the wrapper and its servers running until stdin closes.

### Official plugins — disabled by the operator's install, warned by setup

**ID:** official-plugins-warned
**Status:** active
**Chosen:** `setup.sh global` prints the two `claude plugin disable …` commands while
`gopls-lsp@claude-plugins-official` or `kotlin-lsp@claude-plugins-official` is enabled in
`~/.claude/settings.json`; it never edits that file — `install_hooks`' standing rule. Verification
disables them per headless session only (`## Live check`); the operator's machine switches over
after the merge.
**Considered:** setup edits `settings.json` — breaks the rule that the file is the user's own.

### Kickoff warm-up through the LSP tool

**ID:** kickoff-warmup-via-lsp-tool
**Status:** active
**Chosen:** after `kickoff-worktree.sh`, the run calls the LSP tool `documentSymbol` on one tracked
file per wrapped language present in the new worktree (`git ls-files '*.go'`, `'*.kt'`); the
wrapper holds the answer until ready, so the pipeline waits; the run copies the `index.log` ready
line into its narrative. A worktree with no such file skips it.
**Considered:** a standalone `worktree-lsp warm` process — it indexes in a server the session never
talks to, so the session's own server still starts cold.

### Existing worktrees — moved before the finish preflight, every repository at once

**ID:** migrate-before-preflight
**Status:** active
**Chosen:** as `migrate-all-now`, except the guard takes every repository's main checkout in one
call — moving each one's `.worktrees/` worktrees, building one rename map from all of them and
rewriting the records of every listed project — and bare `/flow`'s integrate run calls it once,
naming every distinct main checkout of the change, after surfacing foreign staged work and before
`check-finish-preflight.sh`, so run 1 and run 2 both see migrated worktrees. Run 2 no longer
migrates.
**Considered:** keep the run 2 step and let the preflight tolerate `.worktrees/` — the strict guard
would stop being strict exactly where it is read; the guard reading records across every project
— `flow state list` is per project, and a guard told about one repository cannot know another
repository's sibling root.

## Live check

Against the operator's real machine:

- **Migrate first:** `scripts/migrate-worktrees.sh` on agents, gymie, gymie-frontend,
  gymie-admin-frontend and gymie-playwright (this change's own worktree moves with them; the run
  continues in the moved path); every `MIGRATED`/`HELD` line recorded; then
  `check-worktree-location.sh <main>` per repository — `LOCATION-OK` except worktrees reported
  `HELD` or already outside `.worktrees/`.
- **Plugin under test:** every headless session passes `--plugin-dir <worktree>/mods/worktree-lsp`
  and `--settings '{"enabledPlugins":{"gopls-lsp@claude-plugins-official":false,"kotlin-lsp@claude-plugins-official":false}}'`
  — nothing global changes before the merge. After the merge, the operator's install is
  `setup.sh global` (links the mod) plus the two `claude plugin disable` commands setup prints.

- **Go, agents:** headless `claude -p --allowedTools LSP`, launched from `/Users/tweety53/Projects/agents`
  with this worktree's `mods/worktree-lsp` as the plugin and the official plugins disabled, runs
  all nine operations (goToDefinition, findReferences, hover, documentSymbol, workspaceSymbol,
  goToImplementation, prepareCallHierarchy, incomingCalls, outgoingCalls) on a Go file in a
  migrated `agents-worktrees/` worktree and in the main checkout.
- **Kotlin, gymie:** the same nine from `/Users/tweety53/Projects/gymie` on a Kotlin file in a
  `gymie-worktrees/` worktree and in the main checkout.
- **Record:** each operation's result count per file; `index.log` ready time per root (Go and
  Kotlin, cold).
- **No leftovers:** after each headless session exits, `check-worktree-processes.sh <worktree>`
  prints `CLEAR` for both worktrees. `StopUnder` itself is proven against real child processes by
  task 5's test.
- **Failure looks like:** findReferences/goToDefinition in the worktree returning nothing while the
  main checkout returns results (filter still applies); Kotlin hover/definition empty in the
  worktree (child not rooted there); `HELD` after a session exits. Before this change the same Go
  findReferences returns 0 in a worktree and 2 in main — the baseline figure.

## Open questions
