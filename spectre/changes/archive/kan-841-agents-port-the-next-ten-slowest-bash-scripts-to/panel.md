# Review panel — kan-841-agents-port-the-next-ten-slowest-bash-scripts-to

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | important | scripts/check-finish-preflight.sh:76 | check-guard-symlinks rule 2 no longer requires the lib/ and check-worktree-location.sh siblings the shimmed check-finish-preflight needs at runtime |   |
| F2 | primary | minor | stats/internal/guard/mutate_and_verify_test.go:742 | parity tests git show d71a2327, so the package fails in a shallow clone |   |
| F3 | primary | minor | stats/internal/guard/resolveremotebase.go:34 | run from a deleted cwd the ports pin every git child to it and exit 2 where bash printed CLEAR/RUN1 |   |
| F4 | primary | minor | KNOWN-BUGS.md:47 | KNOWN-BUGS entries cite preparearchivebranch.go:79 and :212, stale after 70336a0b |   |
| F5 | primary | minor | scripts/mutate-and-verify.sh:30 | mutate-and-verify and prepare-archive-branch shim headers omit the flow-guard-build cannot-answer cause |   |
| F6 | principles | important | stats/internal/guard/postmutationcheck.go:94 | child-exit to bash status mapping duplicated three ways; gitExec returns -1 on a signal-killed git so mutate-and-verify exits 255, outside its contract |   |
| F7 | principles | important | scripts/check-finish-preflight.sh:76 | sibling dependency moved out of the only file check-guard-symlinks reads |   |
| F8 | principles | minor | stats/internal/guard/finishpreflight.go:164 | unset FLOW_GUARD_REPO_ROOT is not refused and resolves check-worktree-location.sh at the filesystem root |   |
| F9 | principles | minor | stats/internal/guard/basemoved.go:31 | shared helpers live in unrelated guards files under their prefixes |   |
| F10 | principles | minor | stats/internal/guard/mutate_and_verify_test.go:596 | ignored-signal tests sleep 500ms then look, so a slow handler passes with the defect present |   |
| F11 | principles | minor | stats/internal/guard/check_guard_symlinks_test.go:187 | tests depend on d71a2327 being in history |   |

findings-total: 11
finding-status: F1 fixed
finding-status: F2 deferred test-environment: history-dependent parity tests, documented as a handoff note
finding-status: F3 deferred edge case: deleted cwd, no caller runs a guard from one
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 deferred refactor: helper cohesion, out of this change's scope
finding-status: F10 deferred test strength: sleep-then-look ignored-signal tests
finding-status: F11 deferred test-environment: same history dependency as F2

reproducers-total: 11
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-primary-5.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F7 .superpowers/sdd/reproducers/0-principles-6.sh
finding-reproducer: F8 .superpowers/sdd/reproducers/0-principles-2.sh
finding-reproducer: F9 .superpowers/sdd/reproducers/0-principles-3.sh
finding-reproducer: F10 .superpowers/sdd/reproducers/0-principles-4.sh
finding-reproducer: F11 .superpowers/sdd/reproducers/0-principles-5.sh

## Pass log

### Round 0

- roster: full
- no addition this round — the resolved list ran alone.
- diff size: 24130 changed lines, over cap — proceeded unasked (single worktree)
- docs-only: exit 1 — first non-doc path scripts/break-and-prove.sh; roster dispatched primary+principles (opus/high)
- context bundle: rebuilt — no cached bundle
- standards passed: /Users/tweety53/Projects/agents/.worktrees/kan-841-agents-port-the-next-ten-slowest-bash-scripts-to/CLAUDE.md, /Users/tweety53/Projects/agents/.worktrees/kan-841-agents-port-the-next-ten-slowest-bash-scripts-to/AGENTS.md
- ceiling breach: primary+principles ran 31m (ceiling 15m); operator chose to keep its findings rather than close timed-out and re-dispatch

### Round 1

- fix round 1: panel-fix on F1 F4 F5 F6 F7 F8 (fixer opus/medium); F2 F3 F9 F10 F11 Minor, deferred at round close; pre-fix HEAD 785636e1; context bundle unchanged — reused
- F4 reproducer 0-primary-4.sh hardcoded the Go line numbers instead of reading the KNOWN-BUGS citations, so its pinned re-run answered 2 (identical verdict); re-authored as .superpowers/sdd/reproducers/1-parent-4.sh, prove-reproducer against 7a961366 PROOF HELD (pre 0-demonstrated, post not-demonstrated); F1 F5 F6 F7 F8 flipped on their pinned shas
- re-run: primary and principles each alone on rerun pair opus/low, targeted at their own fixed findings; reads fix-round-1.diff (785636e1..cfe152b9, 184 changed lines, under cap); docs-only exit 1 (scripts/check-base-moved.sh); no addition this round — the resolved list ran alone
- re-run clean: primary (F1 F4 F5 fixed) and principles (F6 F7 F8 fixed), no new finding
fix-mutation: stats/internal/guard/postmutationcheck.go — gitExec rrExitCode(ee.ProcessState) flipped back to ee.ExitCode() — TestGitExecSignalStatus
fix-mutation: stats/internal/guard/finishpreflight.go — unset FLOW_GUARD_WORKTREE_LOCATION refusal disabled — TestCheckFinishPreflight/port:_an_unset_FLOW_GUARD_WORKTREE_LOCATION_is_refused_by_name
fix-mutation: stats/internal/guard/finishpreflight.go — location read reverted to FLOW_GUARD_REPO_ROOT+/scripts/check-worktree-location.sh — TestCheckFinishPreflight (17 subtests)
fix-mutation: scripts/check-finish-preflight.sh — location export spelled without $SCRIPT_DIR — TestShimSiblingsDeclared/check-finish-preflight.sh/names_the_missing_check-worktree-location.sh
fix-mutation: scripts/mutate-and-verify.sh — lib loader reverted to the dirname/BASH_SOURCE spelling — TestShimSiblingsDeclared/mutate-and-verify.sh
fix-mutation: scripts/check-finish-preflight.sh — both sibling spellings reverted (guard defect state) — 1→0 check-guard-symlinks passes on the tree missing the siblings
fix-mutation: scripts/mutate-and-verify.sh — none — header comment only
fix-mutation: KNOWN-BUGS.md — none — doc citation only
fix-mutations-total: 8
