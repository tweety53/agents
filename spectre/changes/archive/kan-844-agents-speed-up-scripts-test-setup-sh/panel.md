# Review panel — kan-844-agents-speed-up-scripts-test-setup-sh

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | stats/internal/setuptest/containment_test.go:265 | the installed check-panel-reproducers.sh runs with the operator's full environment and no FLOW_GUARD_CACHE_DIR, so under cd stats && go test ./... it builds flow-guard into ${XDG_CACHE_HOME:-~/.cache}/flow-guard — outside the sandbox, sampled by neither fingerprint |   |
| F2 | primary | Minor | stats/internal/setuptest/fingerprint_test.go:131 | the leak group writes into only 4 locations where design.md says each sampled location; with .zcode and scripts/ dropped from the samples the leak group still passes |   |
| F3 | primary | Minor | stats/internal/setuptest/main_test.go:106 | cleanup is a defer in runMain, so a -timeout panic, a test panic or SIGINT leaves /tmp/flow-test-setup.* behind, where the package doc promises removal on success and on failure alike |   |
| F4 | principles | Important | skills/flow/implement.md:76 | implement.md:76 restates xhigh scope as dispatched for implementer groups alone, but implement.md:908-909 sends each big gated reviewer bundle on its group's model/effort pair, so an xhigh group also dispatches its reviewer on flow-xhigh |   |
| F5 | principles | Important | stats/internal/setuptest/containment_test.go:265 | the flow-guard cache write outside the sandbox contradicts the package doc's Nothing here writes outside the sandbox |   |
| F6 | principles | Minor | stats/internal/setuptest/fingerprint_test.go:158 | appendTo lives in fingerprint_test.go but serves delimiters_test.go; it belongs in helpers_test.go |   |
| F7 | principles | Minor | stats/internal/setuptest/main_test.go:149 | _ = os.RemoveAll(sandbox) swallows a cleanup failure where bash's rm -rf printed its error |   |
| F8 | exp-failure-modes | Important | stats/internal/setuptest/containment_test.go:265 | the installed check-panel-reproducers.sh inherits the test environment; stats-go sets no FLOW_GUARD_CACHE_DIR, so the shim builds flow-guard into the operator's real ~/.cache/flow-guard |   |
| F9 | exp-failure-modes | Minor | stats/internal/setuptest/main_test.go:106 | defer cleanup() never runs on a go test -timeout expiry, a test-goroutine panic or SIGINT: the sandbox leaks, closeOut is skipped, and orphaned setup.sh children keep writing |   |
| F10 | exp-failure-modes | Minor | stats/internal/setuptest/main_test.go:149 | the os.RemoveAll error is discarded, so a torn cleanup goes unreported |   |
| F11 | mutation | Important | setup.sh:804 | widening the append branch to begins==0 \|\| ends==0 survives: no delimiter case seeds an unbalanced file, so the package doc's defect 2 (1/0 → 2/1 → 3/2) can come back undetected |   |
| F12 | mutation | Minor | stats/internal/setuptest/helpers_test.go:59 | runSetup's HOME and project-dir sandbox refusals can be dropped and every test passes |   |
| F13 | mutation | Minor | stats/internal/setuptest/helpers_test.go:100 | seedGuard's three refusals can be dropped and every test passes |   |
| F14 | mutation | Minor | stats/internal/setuptest/main_test.go:148 | widening cleanup's prefix check to / is undetected |   |
| F15 | mutation | Minor | stats/internal/setuptest/main_test.go:118 | the close-out verdict is unpinned: a closeOut that never fails, or a runMain that ignores it, passes |   |
| F16 | mutation | Minor | stats/internal/setuptest/fingerprint_test.go:38 | dropping ~/.zcode from realHomeFingerprint survives |   |
| F17 | mutation | Minor | stats/internal/setuptest/fingerprint_test.go:95 | narrowing sourceTreeFingerprint to skills alone survives |   |
| F18 | mutation | Minor | stats/internal/setuptest/helpers_test.go:269 | countLinesMatching capped at 1 or matched as a substring survives |   |
| F19 | mutation | Minor | stats/internal/setuptest/delimiters_test.go:93 | backup_once's cp -p → cp survives: plain cp keeps 640 minus umask, so the 640-mode rationale comment is wrong |   |
| F20 | exp-failure-modes | Minor | stats/internal/setuptest/main_test.go:121 | the signal goroutine stops and waits for setup.sh runs only; test-side writes (seedFile, copyFile, log creation) continue and can race cleanup's RemoveAll, and a runSetup starting after stopRuns calls running.Add while running.Wait is in progress (WaitGroup misuse) |   |

findings-total: 20
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 fixed
finding-status: F15 fixed
finding-status: F16 fixed
finding-status: F17 fixed
finding-status: F18 fixed
finding-status: F19 fixed
finding-status: F20 deferred other — logged in KNOWN-BUGS.md

reproducers-total: 20
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-principles-2.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-principles-3.sh
finding-reproducer: F7 none — needs an undeletable file planted in a live sandbox
finding-reproducer: F8 .superpowers/sdd/reproducers/0-exp-failure-modes-1.sh
finding-reproducer: F9 .superpowers/sdd/reproducers/0-exp-failure-modes-2.sh
finding-reproducer: F10 none — needs an injected RemoveAll failure
finding-reproducer: F11 .superpowers/sdd/reproducers/0-mutation-1.sh
finding-reproducer: F12 .superpowers/sdd/reproducers/0-mutation-2.sh
finding-reproducer: F13 .superpowers/sdd/reproducers/0-mutation-3.sh
finding-reproducer: F14 .superpowers/sdd/reproducers/0-mutation-4.sh
finding-reproducer: F15 .superpowers/sdd/reproducers/0-mutation-5.sh
finding-reproducer: F16 .superpowers/sdd/reproducers/0-mutation-6.sh
finding-reproducer: F17 .superpowers/sdd/reproducers/0-mutation-7.sh
finding-reproducer: F18 .superpowers/sdd/reproducers/0-mutation-8.sh
finding-reproducer: F19 none — equivalent mutant for mode under a 022 umask
finding-reproducer: F20 none — the race exists only on the SIGINT path and is timing-dependent

## Pass log

### Round 0

- roster: full — primary+principles · exp-failure-modes+mutation (decided panel, grouping free)
- no addition this round — the resolved list ran alone
- diff size: 4712 lines, under cap (exit 0); docs-only: exit 1, first non-doc path scripts/test-setup-agents.sh; roster dispatched unchanged
- base moved (kan-761, overlap skills/flow/implement.md); operator chose Rebase; rebased clean onto 979ee6cb, recheck CLEAR
- standards passed: CLAUDE.md, AGENTS.md; principles: skills/flow/engineering-principles.md

### Round 1

- fix round 1: F1–F19 fixed inline (dispatch panel-fix-1; commits 1ac33883, a2e74e9d, 49f9d8e0). F1/F5/F8 fixed at the named site: the guard run's own FLOW_GUARD_CACHE_DIR, pinned by an assertion. F11 fixed in delimiters_test.go (lone begin/end cases), not setup.sh. F19 is a comment-only correction. F4 caps the gated reviewer at high per the operator's decision.
- round-1 re-runs (opus/low, one per role): primary F1–F3 fixed; principles F4–F7 fixed; exp-failure-modes F8–F10 fixed + new F20 (Minor, deferred to KNOWN-BUGS.md); mutation F11–F19 fixed, all eight reproducers caught, no new mutations tried (ceiling). Primary's claim that the fix report lacks F1/F3 rows is wrong — they sit in the combined 'F1, F5, F8' and 'F3, F9' rows.
fix-mutation: stats/internal/setuptest/main_test.go — signal.Notify(sig, os.Interrupt, syscall.SIGTERM) removed — SIGTERM mid-run leaves /tmp/flow-test-setup.* behind (fixed: 0 leftover, exit 130; mutant: 1 leftover, exit 143)
fix-mutation: stats/internal/setuptest/containment_test.go — the guard run's FLOW_GUARD_CACHE_DIR cmd.Env line removed — TestGuardsReachInstallAndRunFromIt: the installed guard built flow-guard inside the sandbox (1→0 builds)
fix-mutation: stats/internal/setuptest/fingerprint_test.go — one sampled location dropped from the leak table — TestContainmentChecksDetectLeaks — .superpowers/sdd/reproducers/0-primary-2.sh flips
fix-mutation: stats/internal/setuptest/helpers_test.go — setupRefusal returns nil — TestSetupRefusalKeepsTheInstallerInTheSandbox — 0-mutation-2.sh flips
fix-mutation: stats/internal/setuptest/helpers_test.go — seedRefusal returns nil — TestSeedRefusalRefusesOutsidePathsAndSymlinks — 0-mutation-3.sh flips
fix-mutation: stats/internal/setuptest/main_test.go — sandboxRemovable prefix widened to / — TestCleanupRemovesOnlyHarnessSandboxes — 0-mutation-4.sh flips
fix-mutation: stats/internal/setuptest/main_test.go — closeOut returns 0 always / exitCode ignores closeCode — TestCloseOutFailsOnAnyMovedFingerprint — 0-mutation-5.sh flips
fix-mutation: stats/internal/setuptest/fingerprint_test.go — ~/.zcode dropped from realHomeFingerprint — TestContainmentChecksDetectLeaks zcode rows — 0-mutation-6.sh flips
fix-mutation: stats/internal/setuptest/fingerprint_test.go — sourceTreeFingerprint narrowed to skills — TestContainmentChecksDetectLeaks source rows — 0-mutation-7.sh flips
fix-mutation: stats/internal/setuptest/helpers_test.go — countLinesMatching capped at 1 / substring match — TestCountLinesMatchingCountsWholeLines — 0-mutation-8.sh flips
fix-mutation: setup.sh — append branch widened to begins==0 || ends==0 — TestUnbalancedDelimitersAbortByteIdentical — 0-mutation-1.sh flips
fix-mutation: stats/internal/setuptest/fingerprint_test.go — appendTo moved back out of helpers_test.go — 0-principles-3.sh flips (placement finding; no behaviour)
fix-mutation: stats/internal/setuptest/main_test.go — none — cleanup's RemoveAll error print (F7/F10): needs an injected RemoveAll failure; exempt
fix-mutation: stats/internal/setuptest/delimiters_test.go — none — F19 is a comment-only correction; the mutant is equivalent under a 022 umask
fix-mutation: skills/flow/implement.md — none — F4 is prose (gated reviewer capped at high); 0-principles-1.sh flips on the text
fix-mutations-total: 15
