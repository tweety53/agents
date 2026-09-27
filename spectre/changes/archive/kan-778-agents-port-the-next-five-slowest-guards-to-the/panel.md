# Review panel — kan-778-agents-port-the-next-five-slowest-guards-to-the

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | Important | spectre/changes/kan-778-agents-port-the-next-five-slowest-guards-to-the/tasks.md:106 | Plan commit 6e75d52a added stats/internal/guard/panelexitcontract.go to task 3's Files:, but task 3's commit 472c162f never touched it, so scripts/check-task-records.sh (a ## lint guard) exits 1 |   |
| F2 | primary+principles | Minor | stats/internal/guard/installedcitations.go:242 | Cleanup is defer os.RemoveAll(sandbox), which never runs when SIGINT kills the process — an interrupted run leaves a full sandboxed install in TMPDIR, where the Python's try/finally left nothing |   |
| F3 | principles | Minor | stats/internal/guard/panelreproducers.go:71 | Two DUPLICATED-on-purpose comments (panelreproducers.go:71, panelexitcontract.go:118) point at copies this diff removed |   |
| F4 | principles | Minor | stats/internal/guard/planprovenance.go:3299 | Uses Getenv, so an unset CHECK_PLAN_PROVENANCE_ROOT reports 'is set but empty' although this diff added Env.LookupEnv for exactly that distinction |   |
| F5 | principles | Minor | stats/internal/guard/planprovenance.go:42 | The doc comment at lines 40-43 still says Env.Getenv cannot tell unset from empty, so that case is the set-but-empty refusal (exit 2) — the F4 fix made that false |   |
| F6 | primary | Minor | stats/internal/guard/panelreproducers.go:140 | Four hand-written refusals are untested (null state record, a finding with status null, FLOW_GUARD_REPO_ROOT unset in two guards); disabling any one keeps the Go test green, and losing the null-state check fails open |   |
| F7 | primary | Minor | .flow/project.md:17 | The rewritten paragraph lists check-task-commit-fields.py as a running Python guard; it is a flow_guard_exec shim, the .py only a module check-plan-shape.py imports |   |
| F8 | primary | Minor | skills/flow/SKILL.md:156 | Names 2 of the 7 flow-guard shims in skills/flow/scripts/ as needing lib/flow-guard.sh, omitting check-panel-reproducers.sh, check-panel-reproducer-exit-contract.sh, check-cleanup-complete.sh, gather-dispatch-context.sh and run-reproducer.sh |   |
| F9 | principles | Minor | scripts/reproducer-metachars.sh:25 | This change removed the last bash consumer of reproducer-metachars.sh and lib/change-plan.sh but keeps both as a live second copy of what the Go port carries, with no decision recorded |   |
| F10 | principles | Minor | stats/internal/guard/guard.go:17 | Env.LookupEnv sits beside Getenv, which it could derive; an Env built with only Getenv, as existing tests build it, panics on a nil func in the three guards that call LookupEnv |   |
| F11 | principles | Minor | stats/internal/guard/check_panel_reproducers_test.go:160 | Case 20 is named cannot-answer, not violations-found but asserts exit 1 (violations-found) |   |
| F12 | principles | Minor | stats/internal/guard/installedcitations.go:347 | TMPDIR comes from the injected Env but setup.sh gets os.Environ(), bypassing the injected environment |   |
| F13 | primary | Minor | skills/flow-contracts/jira-followups.md:17 | Still frames integrate run 1 as one of several filing sites ('every site that files a follow-up', heading 'The filing site's outstanding items', 'that site's list') after the round-close filing was removed |   |

findings-total: 13
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 deferred the four refusals are correct at HEAD; only their tests are missing
finding-status: F7 deferred a documentation wording slip in .flow/project.md, no behaviour
finding-status: F8 deferred an incomplete list in SKILL.md prose, no behaviour
finding-status: F9 deferred retiring the bash copies is a separate cleanup outside this port's scope
finding-status: F10 deferred no caller builds a Getenv-only Env for these guards; a latent test-construction hazard
finding-status: F11 deferred a subtest name only; its assertion is correct
finding-status: F12 deferred setup.sh inherits the real process env the Python gave it too; the injection gap is test-only
finding-status: F13 deferred wording only — the contract still resolves to the one site

reproducers-total: 13
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-principles-4.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/1-principles-1.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/3-primary-1.sh
finding-reproducer: F7 .superpowers/sdd/reproducers/3-primary-2.sh
finding-reproducer: F8 .superpowers/sdd/reproducers/3-primary-3.sh
finding-reproducer: F9 .superpowers/sdd/reproducers/3-principles-1.sh
finding-reproducer: F10 .superpowers/sdd/reproducers/3-principles-2.sh
finding-reproducer: F11 .superpowers/sdd/reproducers/3-principles-3.sh
finding-reproducer: F12 .superpowers/sdd/reproducers/3-principles-4.sh
finding-reproducer: F13 .superpowers/sdd/reproducers/4-primary-1.sh

## Pass log

### Round 0

- roster: compact (rolled 40) — one dispatch primary+principles opus/high; docs-only: exit 1, first non-doc path .gitignore — resolved roster unchanged; no operator addition; relocation comparison: none generated; standards passed: CLAUDE.md, AGENTS.md
- diff size: 24519 changed lines, cap 5000 (single worktree) — over cap, proceeded unasked
- dedup: F1 and F2 raised by both primary and principles — one row each under primary+principles

### Round 1

- F1 fixed inline by the parent in planning commit 4f69e3c2: task 3 Files: back to 472c162f's files; panel-fix-0 dispatched on the fixer (opus/medium) for F2-F4 (Minors joining the round beside F1), read panel-report-0-*.md
- panel-fix-0 fixed F2 1f66e869, F4 4e20f904, F3 7d481191+bc80196e; parent inline c63b04c1 completed F3 (three more stale pointers of the same kind the fixer named); all 6 reproducers exit 0; fix diff fix-round-1.diff = 8c1c61f9..HEAD
- re-runs: primary (opus/low) F1 F2 fixed, nothing new; principles (opus/low) F1-F4 fixed, raised F5 Minor (stale doc comment left by the F4 fix)
fix-mutation: stats/internal/guard/installedcitations.go — signal.NotifyContext replaced by context.WithCancel (no SIGINT handling) — TestCheckInstalledCitationsSIGINT
fix-mutation: stats/internal/guard/installedcitations.go — deferred os.RemoveAll(sandbox) deleted — TestCheckInstalledCitationsSIGINT
fix-mutation: stats/internal/guard/installedcitations.go — SIGINT re-raise syscall.Kill deleted — TestCheckInstalledCitationsSIGINT
fix-mutation: stats/internal/guard/planprovenance.go — LookupEnv replaced by Getenv with set forced true — TestCheckPlanProvenance/root_override_unset_or_empty
fix-mutation: stats/internal/guard/panelexitcontract.go — none — comment-only fix, no executable behaviour changed
fix-mutation: stats/internal/guard/panelreproducers.go — none — comment-only fix, no executable behaviour changed
fix-mutation: stats/internal/guard/cleanupcomplete.go — none — comment-only fix (parent inline, c63b04c1), no executable behaviour changed
fix-mutation: spectre/changes/kan-778-agents-port-the-next-five-slowest-guards-to-the/tasks.md — task 3 Files: re-adds panelexitcontract.go — 1→0 declared-file mismatches reported by scripts/check-task-records.sh
fix-mutations-total: 8

### Round 2

- F5 fixed inline by the parent (comment-only), reproducer 1-principles-1.sh exits 0; an all-Minor round re-runs no slot — the full-policy final whole-branch pass follows
fix-mutation: stats/internal/guard/planprovenance.go — none — comment-only fix (parent inline), no executable behaviour changed
fix-mutations-total: 1

### Round 3

- final whole-branch pass (rerun policy full, pass 2 of the cap): primary+principles bundled on panel.dispatches opus/high; base CLEAR
- final whole-branch pass: primary 3 Minor (F6-F8), principles 4 Minor (F9-F12), no Critical/Important — every Minor deferred per the standing rule; no slot re-runs
- deferred-findings follow-up: operator chose Leave them unfiled — ⚠ Jira: skipped — deferred findings follow-up not filed — operator declined
- check-panel-fix-single-dispatch exit 1: nine sdd-phase per-task inline fixes were recorded -role panel-fix under task-<n>-implementer-fix-<k> keys (out of shape); the panel round itself carried one panel-fix-0 dispatch. Auto-resolved Continue — violation stays recorded

### Round 4

- late-fix reduction after panel close: operator instruction (drop the deferred-findings Jira prompt, record in KNOWN-BUGS.md) landed 8dbcbdd6, d51a398f, 50a36cca — docs-only delta b75d78cd..HEAD; primary alone on opus/low
- late-fix primary: F13 Minor only — deferred per the standing rule, recorded in KNOWN-BUGS.md after verify

### Round 5

- late-fix reduction: operator instruction (retire rerun policy full; re-runs read the diff plus neighbouring code at most) landed 80de983e + plan 99b1950d; primary alone on opus/low over a200bb43..HEAD
- late-fix primary: clean, no findings
