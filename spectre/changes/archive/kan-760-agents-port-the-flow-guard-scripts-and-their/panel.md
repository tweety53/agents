# Review panel — kan-760-agents-port-the-flow-guard-scripts-and-their

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | scripts/check-task-commit-fields.sh:100 | The shims exec whichever flow-guard is first on PATH. Nothing ties that binary to the checked-out source it claims to be. |   |
| F2 | primary | Important | scripts/check-cleanup-complete.sh:211 | Landing this change makes main's five guards depend on a binary no /flow step installs, so they stop answering on any machine that has not run make install-guard. |   |
| F3 | primary | Important | spectre/changes/kan-760-agents-port-the-flow-guard-scripts-and-their/tasks.md:503 | Task 11's Files: declares scripts/prove-reproducer.sh, but task 11's commit 50a52695 never touches it. |   |
| F4 | primary | Important | spectre/changes/kan-760-agents-port-the-flow-guard-scripts-and-their/tasks.md:478 | Two ## lint guards are red on HEAD over this change's own plan. |   |
| F5 | primary | Minor | stats/README.md:50 | Three sites still say a missing binary is exit 2 for every guard; run-reproducer's shim exits 4. |   |
| F6 | primary | Minor | scripts/check-task-commit-fields.sh:84 | The shim's header still promises every exit-2 refusal prints the COULD NOT JUDGE opening; the shim's own missing-binary exit 2 does not. |   |
| F7 | primary | Minor | stats/internal/guard/runreproducer_other.go:16 | A failed process-table read still returns an empty table silently. |   |
| F8 | principles | Important | scripts/run-reproducer.sh:114 | Each ported guard's behaviour is now defined twice: by the checked-out Go source and by the binary last copied to ~/.local/bin/flow-guard, with nothing keeping them equal. |   |
| F9 | principles | Important | scripts/check-task-commit-fields.py:1524 | The Python module keeps its whole verdict half although nothing executes it any more. |   |
| F10 | principles | Important | spectre/changes/kan-760-agents-port-the-flow-guard-scripts-and-their/tasks.md:688 | Two listed ## lint guards are red on HEAD over this change's own plan. |   |
| F11 | principles | Minor | stats/internal/guard/cleanupcomplete.go:377 | ccPlainName and changePlanNameOK encode one change-name allowlist twice inside a single Go package. |   |
| F12 | principles | Minor | scripts/lib/within-root.sh:3 | Both libraries now say NO BASH CALLER REMAINS and are sourced by nothing. |   |
| F13 | principles | Minor | stats/internal/guard/helpers_test.go:38 | TestMain writes outside t.TempDir() and rewrites the process-wide PATH; the plan's isolation rule was relaxed without the plan saying so. |   |
| F14 | principles | Minor | stats/internal/guard/check_task_commit_fields_test.go:2727 | The Go-vs-Python parity tests t.Skip when python3 is absent, turning the drift check off with a passing package. |   |
| F15 | principles | Minor | stats/internal/guard/runreproducer_darwin.go:31 | A failed process-table read returns an empty table; a live survivor goes unnamed. |   |
| F16 | principles | Minor | stats/cmd/flow-guard/main.go:9 | The comment calls exit 2 every guard's cannot-answer code; run-reproducer's is 4. |   |
| F17 | primary | Important | scripts/lib/flow-guard.sh:96 | The cache key ignores GOOS/GOARCH/GOFLAGS, so one shim call with a cross-compile env caches a foreign binary and every later call of all five guards exits 126 (Exec format error), not its cannot-answer code, until the cache is deleted. |   |
| F18 | primary | Important | scripts/test-git-config-pins.sh:67 | The KAN-596 git-pin audit scans scripts/ only; the ported pins in taskcommitfields.go:1305/1315 (--no-renames) and runreproducer.go:868 (--untracked-files=normal) can be removed with the audit and go test still green. |   |
| F19 | primary | Important | spectre/changes/kan-760-agents-port-the-flow-guard-scripts-and-their/proposal.md:35 | proposal.md still says make install-guard, make restart and setup.sh global build flow-guard (all removed by guard-binary-built-from-checkout), and its ~60s panel-reproducers bullet contradicts the recorded After measurement. |   |
| F20 | primary | Minor | scripts/gather-dispatch-context.sh:201 | ${BASH_SOURCE[0]%/*} in all five shims fails when a shim is invoked by bare filename (cannot load lib/flow-guard.sh); the bash guard used dirname and worked. |   |
| F21 | primary | Minor | scripts/lib/flow-guard.sh:106 | An unchecked mv -f under set -e exits 1 on a failed rename, a verdict code for three guards rather than their cannot-answer code. |   |
| F22 | principles | Important | scripts/lib/flow-guard.sh:49 | The cached artifact is keyed on less than it is built from (no GOOS/GOARCH/GOFLAGS), so the first caller's environment decides what every later caller runs — Single Source of Truth, Idempotency, Robustness. |   |
| F23 | principles | Minor | stats/internal/guard/gatherdispatch_test.go:29 | TestGatherDispatchContext omits t.Parallel(), against the plan header's test-isolation rule. |   |
| F24 | principles | Minor | stats/internal/guard/taskcommitfields.go:1537 | os.MkdirTemp("", …) reads the process TMPDIR rather than Env, so cases 112–122 create git worktrees outside t.TempDir(). |   |
| F25 | principles | Minor | stats/internal/guard/cleanupcomplete.go:382 | The package-wide name allowlist carries the cleanup-complete prefix ccPlainName while gather, panel and changeplan call it — Least Astonishment. |   |

findings-total: 25
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
finding-status: F20 fixed
finding-status: F21 fixed
finding-status: F22 fixed
finding-status: F23 fixed
finding-status: F24 fixed
finding-status: F25 fixed

reproducers-total: 25
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-primary-5.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-primary-6.sh
finding-reproducer: F7 none — the failing branch is the non-darwin build or a sysctl error, not drivable on this darwin host without editing the code
finding-reproducer: F8 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F9 .superpowers/sdd/reproducers/0-principles-2.sh
finding-reproducer: F10 .superpowers/sdd/reproducers/0-principles-3.sh
finding-reproducer: F11 .superpowers/sdd/reproducers/0-principles-4.sh
finding-reproducer: F12 .superpowers/sdd/reproducers/0-principles-5.sh
finding-reproducer: F13 .superpowers/sdd/reproducers/0-principles-6.sh
finding-reproducer: F14 .superpowers/sdd/reproducers/0-principles-7.sh
finding-reproducer: F15 none — the failing branch is a sysctl error or the non-darwin ps exec, not drivable on this darwin host without editing the code
finding-reproducer: F16 .superpowers/sdd/reproducers/0-principles-9.sh
finding-reproducer: F17 .superpowers/sdd/reproducers/2-primary-1.sh
finding-reproducer: F18 .superpowers/sdd/reproducers/2-primary-2.sh
finding-reproducer: F19 .superpowers/sdd/reproducers/2-primary-3.sh
finding-reproducer: F20 .superpowers/sdd/reproducers/2-primary-4.sh
finding-reproducer: F21 .superpowers/sdd/reproducers/2-primary-5.sh
finding-reproducer: F22 .superpowers/sdd/reproducers/2-principles-1.sh
finding-reproducer: F23 .superpowers/sdd/reproducers/2-principles-2.sh
finding-reproducer: F24 .superpowers/sdd/reproducers/2-principles-3.sh
finding-reproducer: F25 .superpowers/sdd/reproducers/2-principles-4.sh

## Pass log

### Round 0

- base MOVED (41 commits, overlaps check-task-commit-fields.py, test-check-task-commit-fields.sh, finish-contract-run1.md): Continue — review as is, per the operator's recorded decision port-base-moves-into-go (sync at integrate) and instruction to finish
- roster: compact — 8; diff size 25421 lines over the cap, proceeded automatically; docs-only exit 1 (first non-doc path scripts/check-cleanup-complete.sh), roster primary+principles; no operator-added slot; standards passed: CLAUDE.md (AGENTS.md is its Codex rendering)

### Round 1

- panel-fix-1 (opus/medium, F1 F2 F5 F6 F7 F8 F9 F11 F15 F16) and panel-fix-1-2 (opus/medium, F12 F13 F14) on fix-round-1.diff; F3 F4 F10 fixed by the parent in tasks.md; F12 and F13 reproducers re-authored (the originals were unsatisfiable by a correct fix) and proved both legs with prove-reproducer.sh against c9109bee; every other reproducer flipped pinned to its dispatch sha; F7 F15 closed on the path condition
- round boundary base check: MOVED with the same overlaps — Continue, as at entry (sync at integrate)
- targeted re-runs primary (F1-F7) and principles (F8-F16) on opus/low: all fixed, no new findings — clean
fix-mutation: scripts/run-reproducer.sh — code 4 to 2 — TestRunReproducer/case_15
fix-mutation: stats/internal/guard/check_task_commit_fields_test.go — both t.Fatal lines flipped back to t.Skip (2 substitutions confirmed by grep -c) — TestTaskFieldParseMatchesPython and TestTcfFnmatchMatchesPython went FAIL→SKIP with PATH empty; 0-principles-7.sh went 0→1; restored, diff clean
fix-mutation: stats/internal/guard/helpers_test.go — none — the fix deletes a test-harness env mutation; no production code or assertion depends on it; the observable is the grep (os.Setenv("PATH" count 1 before → 0 after) and unchanged timing
fix-mutation: scripts/lib/within-root.sh, scripts/lib/lexical-normalize.sh — none — deleting uncalled code changes no executable behaviour; the observable is the absence of any reference (grep over the tree, excluding spectre/ and .superpowers/, finds only the "deleted" history notes)
fix-mutation: scripts/lib/flow-guard.sh — `exec "$bin"` → `exec flow-guard` (the F1/F8 fix reverted; the stand-in binary answers, as the reproducer's pre-fix exit 1 did) — test-lib-flow-guard.sh case 1 (STALE-BINARY-ANSWERED), cases 2, 3
fix-mutation: scripts/lib/flow-guard.sh — cache key printed as a constant — test-lib-flow-guard.sh cases 4a, 4b, 5
fix-mutation: scripts/lib/flow-guard.sh — `_test.go` no longer excluded from the key — test-lib-flow-guard.sh case 5
fix-mutation: scripts/lib/flow-guard.sh — no-go branch exits 1 instead of the cannot-answer code — test-lib-flow-guard.sh case 6
fix-mutation: scripts/lib/flow-guard.sh — failed-build branch exits 1 — test-lib-flow-guard.sh case 7
fix-mutation: scripts/lib/flow-guard.sh — empty-cache-location check disabled — test-lib-flow-guard.sh case 8
fix-mutation: scripts/lib/flow-guard.sh — default cache without the /flow-guard suffix — test-lib-flow-guard.sh case 9
fix-mutation: scripts/lib/flow-guard.sh — mid-build edit branch exits 0 — test-lib-flow-guard.sh case 11
fix-mutation: scripts/lib/flow-guard.sh — post-build re-hash disabled — test-lib-flow-guard.sh case 11
fix-mutation: scripts/lib/flow-guard.sh — no-stats branch exits 1 — test-lib-flow-guard.sh case 10
fix-mutation: scripts/check-task-commit-fields.sh — opening without "COULD NOT JUDGE — not a commit verdict:" (F6) — TestCheckTaskCommitFields/case_56
fix-mutation: scripts/run-guard-tests.sh — FLOW_GUARD_CACHE_DIR export removed — test-run-guard-tests.sh case 10
fix-mutation: scripts/run-guard-tests.sh — companion marker reverted to `exec flow-guard <name>` — test-run-guard-tests.sh case 8
fix-mutation: stats/cmd/flow-guard/main.go — cannotAnswer("run-reproducer") 4 → 2 — TestCannotAnswerIsTheGuardsOwnCode
fix-mutation: stats/internal/guard/runreproducer.go — `if res.tableErr != nil` → `if false` — TestRunReproducer/a_failed_process-table_read_is_never_a_verdict_(exit_4) (exit 1, want 4)
fix-mutation: stats/internal/guard/runreproducer.go — read() stops recording the table error — TestRunReproducer/a_failed_process-table_read_is_never_a_verdict_(exit_4)
fix-mutation: stats/internal/guard/cleanupcomplete.go — ccPlainName's `name == "" ||` removed — TestPlainNameRefusesEmpty (survived the existing suites; the test was added for it)
fix-mutation: scripts/check-task-commit-fields.py — none — F9 deletes code no importer or test reaches; the live importers' harnesses and both Python parity tests stay green
fix-mutation: stats/Makefile, setup.sh — none — deletion of install paths; test-setup.sh and test-make-build.sh stay green
fix-mutation: stats/internal/guard/changeplan.go, gatherdispatch.go, panelexitcontract.go — none — call sites renamed to the one allowlist function; its empty-name branch is the only behaviour delta, mutated above
fix-mutation: stats/internal/guard/runreproducer_other.go, runreproducer_darwin.go — none — the error returns are driven through the Env.ProcTable seam; the real sysctl and ps failures cannot be induced on this darwin host without editing the code (the reviewers' own exemption)
fix-mutations-total: 25

### Round 2

- final-pass entry: base MOVED (41 commits), overlaps check-task-commit-fields.py, test-check-task-commit-fields.sh, finish-contract-run1.md — unchanged set, Continue (auto, recommended; resolved at integrate rebase)
- diff size 27271 over cap (exit 1) — proceeded automatically; docs-only exit 1 (first non-doc path scripts/check-cleanup-complete.sh) — roster primary+principles unchanged, one bundled dispatch opus/high per panel.dispatches
- final whole-branch pass (opus/high) raised F17-F25 (4 Important, 5 Minor; F22 duplicates F17 across passes); panel-fix-2 (opus/medium) fixed F17 F18 F20-F25 in 6 commits, F19 fixed by the parent in proposal.md; all 9 reproducers flipped pinned to their shas; residual accepted: a GOOS persisted via go env -w is not cleared by env -u (host-level misconfiguration breaking all go use)
fix-mutation: scripts/lib/flow-guard.sh — removed `env -u GOOS -u GOARCH -u GOFLAGS` from the go build — scripts/test-lib-flow-guard.sh case 12 (first cross-env call rc=126)
fix-mutation: scripts/lib/flow-guard.sh — reverted the checked `if ! mv …` block to a bare `mv -f "$tmp" "$bin"` — scripts/test-lib-flow-guard.sh case 13 (rc=1)
fix-mutation: scripts/lib/flow-guard.sh — dropped `|| :` from the error branch's `rm -f "$tmp"` — scripts/test-lib-flow-guard.sh case 13 (rc=1)
fix-mutation: scripts/lib/flow-guard.sh — reverted the lib's `dirname` stats lookup to `${BASH_SOURCE[0]%/*}` — none: survives; equivalent mutant — every shim sources the lib as `<dir>/lib/flow-guard.sh`, so the lib's BASH_SOURCE[0] always contains a slash and `%/*` equals `dirname`
fix-mutation: scripts/check-cleanup-complete.sh — reverted the lib source line to `${BASH_SOURCE[0]%/*}` — scripts/test-lib-flow-guard.sh case 14 (check-cleanup-complete)
fix-mutation: scripts/check-panel-reproducer-exit-contract.sh — reverted the lib source line — scripts/test-lib-flow-guard.sh case 14 (check-panel-reproducer-exit-contract)
fix-mutation: scripts/check-task-commit-fields.sh — reverted the lib source line — scripts/test-lib-flow-guard.sh case 14 (check-task-commit-fields)
fix-mutation: scripts/gather-dispatch-context.sh — reverted the lib source line — scripts/test-lib-flow-guard.sh case 14 (gather-dispatch-context)
fix-mutation: scripts/run-reproducer.sh — reverted the lib source line — scripts/test-lib-flow-guard.sh case 14 (run-reproducer)
fix-mutation: scripts/test-git-config-pins.sh — removed the `.go` → scan_go dispatch — case 6 (rc=0, no hits); guard observable on the defect state (every Go pin stripped, reproducer 2-primary-2): audit rc=0 before the fix, rc=1 after
fix-mutation: scripts/test-git-config-pins.sh — dropped the `_test.go` exclusion from targets — case 6 and case 7 (test-file git calls flagged)
fix-mutation: stats/internal/guard/taskcommitfields.go — `os.MkdirTemp(tmpdir, …)` back to `os.MkdirTemp("", …)` — TestCheckTaskCommitFields/case_124a
fix-mutation: stats/internal/guard/check_task_commit_fields_test.go — tcfRepo.getenv's TMPDIR override disabled (key renamed) — reproducer 2-principles-3 (DEFECT: worktree directly under TMPDIR); no Go test asserts where the harness's own cases write, the reproducer is the check of that test-isolation rule
fix-mutation: stats/internal/guard/gatherdispatch_test.go — removed the added `t.Parallel()` — reproducer 2-principles-2; not observable through any test outcome (scheduling only)
fix-mutation: stats/internal/guard/cleanupcomplete.go — none — F25 is a pure rename; the compiler is the check
fix-mutations-total: 15

### Round 3

- round-3 entry: base check — same overlap set, Continue (auto, recommended)
- targeted re-runs primary (F17-F21) and principles (F22-F25) on opus/low: all fixed, no new findings — clean; go env -w GOOS residual judged negligible by primary (loud exit 126, never a wrong verdict)

### Round 4

- final-pass repeat (full policy, one unasked): third whole-roster read, so demoted to DEFAULT_MODEL opus at low effort in place of opus/high per panel.dispatches; base MOVED same overlap set, Continue (auto); diff 28625 lines over cap, proceeded automatically; roster primary+principles
- final-pass repeat (opus/low, demoted): primary and principles both no findings — clean; shim-vs-bash parity 16/16 byte-identical
- auto-resolved: check-panel-fix-single-dispatch exit 1 (out-of-shape keys task-13-implementer-fix-1b, task-13-implementer-fix-2 — per-task review fixes recorded under the panel-fix role) → Continue (recommended); violation stays recorded
- integrate: rebased onto origin/main ad72b904 (66 commits); conflicts: scripts/test-check-task-commit-fields.sh modify/delete — kept the deletion (upstream's only change, KAN-676 cases 140-146, is ported to Go by task 12); scripts/check-task-commit-fields.py content — kept the branch's trimmed module (upstream's check_evidence_tags sits in the deleted verdict half, ported to taskcommitfields.go); finish-contract-run1.md merged cleanly; full lint+test re-run 33/33 exit 0, 84/84 harnesses
