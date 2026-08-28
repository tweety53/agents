# Review panel — kan-88-finish-never-reconciles-the-base-branch

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Bugbot | Minor | scripts/check-finish-preflight.sh:121 | Collapsing the ANCESTOR_RC -eq 1 / -ne 0 two-step into "any nonzero means RUN1" survives the whole suite — no regression test protects the explicit exit-2 branch for a merge-base failure that is not "not an ancestor". |
| F2 | Bugbot | Minor | scripts/check-base-moved.sh:85 | Removing the COUNT assignment failure guard survives the full suite untouched — nothing exercises a failing git rev-list in this guard, so nothing proves a git failure can never be read as a verdict. |
| F3 | Bugbot | Minor | scripts/check-base-moved.sh:137 | The exactly-10-overlap cap boundary is untested: changing TOTAL -gt 10 to -ge 10 survives the suite because the only cap case exercises 11 overlaps. |
| F4 | Principles | Important | scripts/test-check-finish-preflight.sh:246 | fix round 1's new case 8c copies the entire 14-line failing-git shim block verbatim from pre-existing case 8b, and a third copy of the same technique now sits in the sibling harness — one piece of knowledge written out three times, with nothing keeping the copies in step. |
| F5 | Primary | Minor | spectre/changes/kan-88-finish-never-reconciles-the-base-branch/tasks.md:100 | Task 1's Baseline line still reads after=29 assertions; fix round 1 added case 8c, so the harness now makes 32. |
| F6 | Primary | Minor | spectre/changes/kan-88-finish-never-reconciles-the-base-branch/tasks.md:166 | Task 2's Baseline line still reads 12 cases in the new harness; fix round 1 added cases 6b and 9b, so it now carries 14. |
| F7 | Bugbot | Important | scripts/check-finish-preflight.sh:77 | Dropping the failure guard on the HEAD_SHA capture is caught by no test — the suite never exercises a worktree whose HEAD does not resolve. |
| F8 | Bugbot | Important | scripts/check-base-moved.sh:95 | Dropping the failure guard on the MOVED_RAW capture is caught by no test, and the sibling captures COMMITTED_RAW, STAGED_RAW and UNSTAGED_RAW share the same shape and the same gap. |
| F9 | Bugbot | Minor | scripts/lib/resolve-remote-base.sh:35 | Removing --end-of-options from the rev-parse call is caught by neither suite: no case passes a base ref beginning with a dash, which is the exact case the header says the flag exists for. |
| F10 | Principles | Minor | scripts/lib/test-git-shim.sh:1 | Every other file in scripts/lib/ is sourced by at least one production check-*.sh guard; test-git-shim.sh is the only entry with no production consumer, so it breaks what the directory otherwise signals to a reader. |
| F11 | Primary | Minor | spectre/changes/kan-88-finish-never-reconciles-the-base-branch/tasks.md:100 | Task 1's Baseline line reads after=32 assertions; fix round 2 added cases 8d, 13 and 14, so the harness now makes 38. |
| F12 | Primary | Minor | spectre/changes/kan-88-finish-never-reconciles-the-base-branch/tasks.md:166 | Task 2's Baseline line reads 14 cases; fix round 2 added 9c, 9d, 9e and 9f, so it now carries 18. |
| F13 | Primary | Minor | spectre/changes/kan-88-finish-never-reconciles-the-base-branch/tasks.md:249 | Task 3's Baseline byte counts are the pre-change sizes, not the sizes the task produced; all three files are still well under budget, so this is an inaccurate record rather than a budget breach. |
| F14 | Bugbot | Important | scripts/lib/test-git-shim.sh:33 | Broadening the shim's exact-equality argument match to a substring match fails no test in either suite, so the helper's own documented requirement that the match argument be specific enough to fail only the call under test is unenforced. |
| F15 | Code review (low) | Important | scripts/lib/test-git-shim.sh:24 | The shim-based capture-guard cases were reported as intermittently failing because the generated git shim sometimes does not intercept its target call. |
| F16 | Primary | Minor | spectre/changes/kan-88-finish-never-reconciles-the-base-branch/tasks.md:96 | Task 1's Files field omits scripts/lib/test-git-shim.sh, which that task's commit now adds. |
| F17 | Bugbot | Minor | scripts/lib/test-git-shim.sh:95 | Making assert_shim_fired unconditionally pass is caught by nothing, since no test tests the assertion helpers themselves — the slot raising it says plainly this is inherent to test-helper code and not a shippable defect. |
| F18 | Bugbot | Minor | scripts/lib/test-git-shim.sh:45 | Sourcing the library twice while a shim sits on PATH recaptures the shim as the real git; unreachable today because both harnesses source it exactly once before touching PATH, so it is latent fragility rather than a live bug. |
| F19 | Bugbot | Minor | scripts/lib/test-git-shim.sh:45 | F18's own fix is unpinned: reverting the idempotent real-git capture to an unconditional assignment passes all three suites, so nothing keeps the repair in place. |

findings-total: 19
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

reproducers-total: 19
finding-reproducer: F1 none — a surviving mutant, provable only by a mutate/run/revert sequence rather than one runnable command
finding-reproducer: F2 none — a surviving mutant, provable only by a mutate/run/revert sequence rather than one runnable command
finding-reproducer: F3 none — a surviving mutant at a boundary the suite never reaches, provable only by a mutate/run/revert sequence
finding-reproducer: F4 grep -c SHIM_DIR scripts/test-check-finish-preflight.sh
finding-reproducer: F5 grep -c pass scripts/test-check-finish-preflight.sh
finding-reproducer: F6 grep -c ok scripts/test-check-base-moved.sh
finding-reproducer: F7 none — a surviving mutant, provable only by a mutate/run/revert sequence rather than one runnable command
finding-reproducer: F8 none — a surviving mutant, provable only by a mutate/run/revert sequence rather than one runnable command
finding-reproducer: F9 none — a surviving mutant, provable only by a mutate/run/revert sequence rather than one runnable command
finding-reproducer: F10 none — a directory-convention judgment, not a condition one runnable command decides
finding-reproducer: F11 grep -c pass scripts/test-check-finish-preflight.sh
finding-reproducer: F12 grep -c ok scripts/test-check-base-moved.sh
finding-reproducer: F13 wc -c skills/flow/integrate.md
finding-reproducer: F14 none — a surviving mutant, provable only by a mutate/run/revert sequence rather than one runnable command
finding-reproducer: F15 none — the reported intermittency did not reproduce; see the dispatcher note on this finding
finding-reproducer: F16 git show d3d762d --stat
finding-reproducer: F17 none — a meta-level gap: the assertion helper has no test of its own, and any test of it would have the same property one level up
finding-reproducer: F18 none — needs a double-source of the library with a shim already on PATH, which no worktree-relative single command expresses
finding-reproducer: F19 none — a surviving mutant, provable only by a mutate/run/revert sequence rather than one runnable command
