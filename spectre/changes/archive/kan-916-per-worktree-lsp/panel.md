# Review panel — kan-916-per-worktree-lsp

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | skills/flow-contracts/finish-contract-run2.md:131 | run 2's migrate-worktrees.sh step can never be reached for the worktree it exists to move: check-finish-preflight.sh turns RUN2 into REFUSE on a .worktrees/ STRAY first |   |
| F2 | primary | Important | stats/internal/guard/migrateworktrees.go:63 | a cross-repo change's record keeps the other repo's old .worktrees/ key: the guard lists only -C <main> records and renames only from <main>'s own worktree list |   |
| F3 | primary | Minor | stats/internal/lspmux/mux.go:303 | a server that died is replaced without the documents the client already opened in it (their didOpen is not re-sent) |   |
| F4 | principles | Important | stats/internal/guard/worktreelocation.go:64 | the sibling <repo>-worktrees root rule is written out separately in worktreelocation.go, kickoffworktree.go and migrateworktrees.go; one siblingRoot(main) helper |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 none — needs a server to die mid-session with an unsaved buffer; no fixture exists
finding-reproducer: F4 .superpowers/sdd/reproducers/0-principles-1.sh

## Pass log

### Round 0

- roster: compact — 44
- no addition this round — the resolved list ran alone.
- diff size: 4538 lines, under cap — proceed
- docs-only: exit 1 (first non-doc path .gitignore) — roster primary+principles dispatched
- standards passed: CLAUDE.md, AGENTS.md
- citation check: scripts/check-references.sh captured to citation-check.md

### Round 1

- F1 reproducer: demonstrates citation finishpreflight.go:172 elides content ('...'), refused by the exit-contract audit; run-reproducer demonstrated (exit 1, sha 45325446…). Repaired by the fix subagent with both legs recorded, not bounced — the defect itself is confirmed.
- auto-decided F1: run the migration at the start of every bare /flow integrate run, before check-finish-preflight — the reviewer's recommended route
- auto-decided F2: migrate-worktrees takes every repository's main checkout at once and rewrites records across all of them
- panel-fix-1: F1,F2,F3,F4 to the fixer (opus/medium), reading final-review.diff
- re-run: primary+principles on fix-round-1.diff (rerun pair sonnet/low); both raised findings this round fixed
fix-mutation: stats/internal/guard/worktreelocation.go — siblingRoot suffix `-worktrees` → `-wt` — TestCheckWorktreeLocationSiblingLayout, TestKickoffWorktreeSiblingLayout, TestMigrateWorktrees failed (all three callers)
fix-mutation: stats/internal/guard/migrateworktrees.go — rename map built from the last repository only (`repos[len(repos)-1:]`) — TestMigrateWorktrees/a_cross-repo_record_is_rewritten_for_every_repository_named ("kan-1-free keeps an old key", b's key stale)
fix-mutation: stats/internal/guard/migrateworktrees.go — records read from the first repository only (`repos[:1]`) — TestMigrateWorktrees/a_cross-repo_record_is_rewritten_for_every_repository_named (record never rewritten)
fix-mutation: stats/internal/guard/migrateworktrees.go — `state set -C repos[0].main` instead of the record's own main — TestMigrateWorktrees/a_cross-repo_record_is_rewritten_for_every_repository_named
fix-mutation: skills/flow-contracts/finish-contract-run1.md + skills/flow/integrate.md — migration step absent before the preflight (pre-fix tree, ea98df8f) — 0-primary-1.sh exit 1→0 (preflight REFUSE stray → migrated, RUN2 reached in the end-to-end run)
fix-mutation: stats/internal/lspmux/mux.go — replay loop emptied (no didOpen to the new server) — TestRespawnReopensDocuments
fix-mutation: stats/internal/lspmux/mux.go — didChange not recorded (`docs[uri] = d` dropped) — TestRespawnReopensDocuments (replayed v1 "package x\n")
fix-mutation: stats/internal/lspmux/mux.go — didClose not recorded (`delete(docs, uri)` dropped) — TestRespawnReopensDocuments (b.go reopened)
fix-mutation: stats/internal/lspmux/mux.go — replay written before `initialized` — TestRespawnReopensDocuments (order initialize, didOpen, initialized)
fix-mutation: stats/internal/lspmux/mux.go — `track` before `route` (first server gets a didOpen twice) — TestRespawnReopensDocuments ("received 3 didOpens, want 2"; survived until ccc8e6a0 added the assertion)
fix-mutation: stats/internal/lspmux/ready.go — removed root's docs not deleted — TestRemovedRootStopsChild (the removed root's open documents are still remembered)
fix-mutations-total: 11

### Round 2

- late-fix reduction: in-run fix 1 (flow.verify) since 101e75c0; primary clean
- not dispatched — late-fix reduction: principles
