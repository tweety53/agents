# kan-916-per-worktree-lsp — live verification (2026-10-07)

Run by `flow.verify`'s live check, `design.md` `## Live check`. Every headless session passed
`--plugin-dir <worktree>/mods/worktree-lsp` and disabled `gopls-lsp`/`kotlin-lsp@claude-plugins-official`
through `--settings`; nothing global changed.

## Migration

First run (`scripts/migrate-worktrees.sh` on agents, gymie, gymie-frontend, gymie-admin-frontend,
gymie-playwright, run from this change's worktree): every agents worktree `MIGRATED`, this one
included, then `FAILED: <wt> — check-worktree-processes.sh could not answer` for all fourteen gymie
worktrees, exit 1 — the guard lost its own scripts directory when it moved the worktree holding it.
Fixed in-run as task 10 (`fix(guard): re-point migrate-worktrees at its moved scripts directory`).

First run's agents moves:

```text
MIGRATED: /Users/tweety53/Projects/agents/.worktrees/flow-quiet-progress -> /Users/tweety53/Projects/agents-worktrees/flow-quiet-progress
MIGRATED: /Users/tweety53/Projects/agents/.worktrees/flow-quiet-progress-widen -> /Users/tweety53/Projects/agents-worktrees/flow-quiet-progress-widen
MIGRATED: /Users/tweety53/Projects/agents/.worktrees/flow-verify-auto-fix -> /Users/tweety53/Projects/agents-worktrees/flow-verify-auto-fix
MIGRATED: /Users/tweety53/Projects/agents/.worktrees/kan-678-flow-fix-baseline-fields-ran-on-three -> /Users/tweety53/Projects/agents-worktrees/kan-678-flow-fix-baseline-fields-ran-on-three
MIGRATED: /Users/tweety53/Projects/agents/.worktrees/kan-828-flow-fix-refreshing-main-checkouts-before-the -> /Users/tweety53/Projects/agents-worktrees/kan-828-flow-fix-refreshing-main-checkouts-before-the
MIGRATED: /Users/tweety53/Projects/agents/.worktrees/kan-861-fix-all-minor-findings-in-known-bugs-md -> /Users/tweety53/Projects/agents-worktrees/kan-861-fix-all-minor-findings-in-known-bugs-md
MIGRATED: /Users/tweety53/Projects/agents/.worktrees/kan-861-fix-all-minor-findings-in-known-bugs-md-wave-group-2 -> /Users/tweety53/Projects/agents-worktrees/kan-861-fix-all-minor-findings-in-known-bugs-md-wave-group-2
MIGRATED: /Users/tweety53/Projects/agents/.worktrees/kan-861-fix-all-minor-findings-in-known-bugs-md-wave-group-3 -> /Users/tweety53/Projects/agents-worktrees/kan-861-fix-all-minor-findings-in-known-bugs-md-wave-group-3
MIGRATED: /Users/tweety53/Projects/agents/.worktrees/kan-861-fix-all-minor-findings-in-known-bugs-md-wave-group-4 -> /Users/tweety53/Projects/agents-worktrees/kan-861-fix-all-minor-findings-in-known-bugs-md-wave-group-4
MIGRATED: /Users/tweety53/Projects/agents/.worktrees/kan-916-per-worktree-lsp -> /Users/tweety53/Projects/agents-worktrees/kan-916-per-worktree-lsp
```

Second run, after the fix, exit 0:

```text
MIGRATED: /Users/tweety53/Projects/gymie/.worktrees/ci-only-on-promotion -> /Users/tweety53/Projects/gymie-worktrees/ci-only-on-promotion
MIGRATED: /Users/tweety53/Projects/gymie/.worktrees/docs-navigation-handoff -> /Users/tweety53/Projects/gymie-worktrees/docs-navigation-handoff
MIGRATED: /Users/tweety53/Projects/gymie/.worktrees/kan-627-flow-fix-spectre-validate-is-red-on-develop-so -> /Users/tweety53/Projects/gymie-worktrees/kan-627-flow-fix-spectre-validate-is-red-on-develop-so
HELD: /Users/tweety53/Projects/gymie/.worktrees/kan-758-offline-l5-fe-group-sessions-offline-logging — <pids elided>
HELD: /Users/tweety53/Projects/gymie-frontend/.worktrees/kan-758-offline-l5-fe-group-sessions-offline-logging — <pids elided>
MIGRATED: /Users/tweety53/Projects/gymie-frontend/.worktrees/review-scratch-1316 -> /Users/tweety53/Projects/gymie-frontend-worktrees/review-scratch-1316
MIGRATED: /Users/tweety53/Projects/gymie-admin-frontend/.worktrees/kan-741-full-visual-verification-of-current -> /Users/tweety53/Projects/gymie-admin-frontend-worktrees/kan-741-full-visual-verification-of-current
MIGRATED: /Users/tweety53/Projects/gymie-playwright/.worktrees/kan-577-step-3-frontend-per-scope-rights-control-confirm -> /Users/tweety53/Projects/gymie-playwright-worktrees/kan-577-step-3-frontend-per-scope-rights-control-confirm
HELD: /Users/tweety53/Projects/gymie-playwright/.worktrees/kan-758-offline-l5-fe-group-sessions-offline-logging — <pids elided>
MIGRATED: /Users/tweety53/Projects/gymie-playwright/.worktrees/kan-758-offline-l5-fe-group-sessions-offline-logging-wave-group-10 -> /Users/tweety53/Projects/gymie-playwright-worktrees/kan-758-offline-l5-fe-group-sessions-offline-logging-wave-group-10
HELD: /Users/tweety53/Projects/gymie-playwright/.worktrees/kan-758-offline-l5-fe-group-sessions-offline-logging-wave-group-11 — <pids elided>
MIGRATED: /Users/tweety53/Projects/gymie-playwright/.worktrees/kan-758-offline-l5-fe-group-sessions-offline-logging-wave-group-8 -> /Users/tweety53/Projects/gymie-playwright-worktrees/kan-758-offline-l5-fe-group-sessions-offline-logging-wave-group-8
MIGRATED: /Users/tweety53/Projects/gymie-playwright/.worktrees/kan-758-offline-l5-fe-group-sessions-offline-logging-wave-group-9 -> /Users/tweety53/Projects/gymie-playwright-worktrees/kan-758-offline-l5-fe-group-sessions-offline-logging-wave-group-9
MIGRATED: /Users/tweety53/Projects/gymie-playwright/.worktrees/kan-761-speed-up-the-playwright-suite-drop-fixed-sleeps-wave-group-c -> /Users/tweety53/Projects/gymie-playwright-worktrees/kan-761-speed-up-the-playwright-suite-drop-fixed-sleeps-wave-group-c
MIGRATED: /Users/tweety53/Projects/gymie-playwright/.worktrees/known-bugs-desktop-flakes -> /Users/tweety53/Projects/gymie-playwright-worktrees/known-bugs-desktop-flakes
```

`check-worktree-location.sh <main>` afterwards: agents `LOCATION-OK`; gymie, gymie-frontend and
gymie-playwright `STRAY` only for the four `HELD` worktrees (a live kan-758 session holds them);
gymie-admin-frontend `STRAY` only for a prunable worktree under `/private/tmp`, already outside
`.worktrees/`. Matched.

## Go, agents (launched from the main checkout)

A = `agents-worktrees/kan-916-per-worktree-lsp/stats/internal/store/settings.go`, B = the same file in the main checkout.

| op | A | B |
|---|---|---|
| goToDefinition | 1 | 1 |
| findReferences | 7 | 7 |
| hover | 1 | 1 |
| documentSymbol | 9 | 9 |
| workspaceSymbol | 2 | 2 |
| goToImplementation | 1 | 1 |
| prepareCallHierarchy | 1 | 1 |
| incomingCalls | 3 | 3 |
| outgoingCalls | 1 | 1 |

Every result in its own tree; workspaceSymbol returns one hit from each tree (fan-out). One
transient: B's first documentSymbol answered `server is starting` from Claude Code; the retry
answered. Baseline before this change: findReferences 0 in a worktree, 2 in main. Matched.

## Kotlin, gymie (launched from the main checkout)

Against `gymie-worktrees/ci-only-on-promotion` (migrated, branch from 2026-09-21): hover 1,
documentSymbol 22, prepareCallHierarchy 1 rooted in the worktree, but goToDefinition 0,
findReferences 1, goToImplementation 0, incomingCalls 0 — the server is rooted there and cannot
index that branch's project. The official `kotlin-lsp` plugin, launched from inside the same
worktree, returns the same empty figures (findReferences 1, goToImplementation 0,
workspaceSymbol 0), so this is that branch's project, not the wrapper.

Against a fresh worktree of `develop` (`gymie-worktrees/kan-916-lsp-probe`, created for this check
and removed after it):

| op | A (worktree) | B (main) |
|---|---|---|
| goToDefinition | 1 | 1 |
| findReferences | 11 (5 files) | 11 (5 files) |
| hover | 1 | 1 |
| documentSymbol | 22 | 22 |
| workspaceSymbol | 8 | 8 |
| goToImplementation | 1 | 1 |
| prepareCallHierarchy | 1 | 1 |
| incomingCalls | 1 | 1 |
| outgoingCalls | 7 (3 files) | 7 (3 files) |

Every result in its own tree; workspaceSymbol carries both trees' four hits. Matched.

## Index ready times (`~/.cache/worktree-lsp/index.log`, ms)

```text
2026-10-07T16:25:57+03:00 gopls /Users/tweety53/Projects/agents-worktrees/kan-916-per-worktree-lsp ready 2254
2026-10-07T16:26:05+03:00 gopls /Users/tweety53/Projects/agents ready 2088
2026-10-07T16:27:42+03:00 kotlin-lsp /Users/tweety53/Projects/gymie ready 35349
2026-10-07T16:27:46+03:00 kotlin-lsp /Users/tweety53/Projects/gymie-worktrees/ci-only-on-promotion ready 40080
2026-10-07T16:31:24+03:00 kotlin-lsp /Users/tweety53/Projects/gymie-worktrees/kan-916-lsp-probe ready 59170
2026-10-07T16:31:57+03:00 kotlin-lsp /Users/tweety53/Projects/gymie-worktrees/kan-916-lsp-probe ready 10483
2026-10-07T16:32:05+03:00 kotlin-lsp /Users/tweety53/Projects/gymie ready 17954
```

Cold: gopls 2254 (worktree) / 2088 (main); kotlin-lsp 59170 (fresh worktree) / 35349 (main).

## Leftovers

`check-worktree-processes.sh` after each session: `CLEAR` for the agents worktree and
`ci-only-on-promotion`. For the fresh gymie worktree it printed `HELD` (the wrapper and its
kotlin server) immediately after `claude -p` returned, and `CLEAR` within 30 seconds — the
kotlin server's own shutdown time.

Live check: matched.
