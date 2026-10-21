# Review panel — kan-918-flow-improvement-port-the-git-snapshot-and

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | mutation | Important | scripts/test-git-config-pins.sh:196 | the new diff(1)-program skip regex lets one bare identifier argument come before the literal, so exec.Command(gitBin, "diff", "--name-only") — a real unpinned git diff read — passes the audit silently |   |
| F2 | mutation | Important | stats/internal/guard/planunchanged.go:117 | no test pins verify's git-status failure as exit 2 (cannot answer); mapping it to 1 (changed) survives TestCheckPlanUnchanged |   |
| F3 | mutation | Important | stats/internal/guard/planunchanged.go:76 | dropping --untracked-files=all survives; no case puts a new untracked file in a new subdirectory of the plan tree |   |
| F4 | mutation | Important | stats/internal/guard/planunchanged.go:101 | no test pins removing the partial snapshot when git status fails during snapshot; dropping the os.Remove survives |   |
| F5 | mutation | Important | stats/internal/guard/treemarkers.go:144 | downgrading grep -cE to the BRE grep -c survives; no marker in the test uses ERE-only syntax |   |
| F6 | mutation | Important | stats/internal/guard/treemarkers.go:74 | no case covers skipping a blank markers-file line; with the skip removed a blank line is refused and the test still passes |   |
| F7 | mutation | Important | stats/internal/guard/treemarkers.go:148 | grep exiting 2 is accepted as a count once the threshold moves to > 2; the grep failed — cannot answer branch is untested |   |
| F8 | mutation | Important | stats/internal/guard/guardautosquash.go:57 | the no resolvable HEAD cannot-answer branch (exit 2) is untested; removing it degrades the result to exit 1 violations |   |
| F9 | mutation | Important | stats/internal/guard/guardautosquash.go:35 | targets <wt> with zero shas is accepted as ok when the bound moves to >= 1; the usage refusal for an empty target list is unpinned |   |
| F10 | mutation | Important | stats/internal/guard/guardautosquash.go:127 | moving the sha-token minimum length from 7 to 6 survives; the 6/7 boundary is unpinned |   |
| F11 | mutation | Important | stats/internal/guard/paneldiffsize.go:37 | moving the default cap from 5000 to 4999 survives; no case runs without a cap argument at the 5000 boundary |   |
| F12 | mutation | Minor | stats/internal/guard/foldfixup.go:159 | fold-fixup ignoring the after verdict survives TestFoldFixup; mapping the guard's exit >=2 to 0 (foldfixup.go:61-62) also survives |   |
| F13 | mutation | Minor | stats/internal/guard/planunchanged.go:36 | planunchanged.go:36 and :47 — mode and name/snap usage checks: removing either survives |   |
| F14 | mutation | Minor | stats/internal/guard/planunchanged.go:74 | dropping LC_ALL=C survives. Likely not observable with ASCII fixtures |   |
| F15 | mutation | Minor | stats/internal/guard/treemarkers.go:43 | the non-directory worktree check survives when only the isDir half is removed |   |
| F16 | primary | Minor | stats/internal/guard/guardautosquash.go:83 | two exit codes now differ from the base bash (unreadable tasks.md exits 2; failing git diff exits 2) but neither is recorded as a design.md Decision or a tasks.md correction while proposal.md says exit codes unchanged |   |
| F17 | principles | Minor | stats/internal/guard/planunchanged.go:157 | diff and grep are found through Env's PATH but run with os.Environ(); the tree-markers diff also runs without the LC_ALL=C the bash set |   |
| F18 | test-sanity | Minor | stats/internal/guard/check_panel_diff_size_test.go:91 | the two prints the count Contains checks are implied by the exact-stdout checks on the same run, so they can never fail on their own |   |
| F19 | call-efficiency | Minor | stats/internal/guard/guardautosquash.go:93 | after starts 2 git processes for each hex token in tasks.md; a single git cat-file --batch-check would resolve them all in one process |   |
| F20 | mutation | Important | stats/internal/guard/guardautosquash.go:70 | dropping ^{commit} from the cat-file --batch-check input survives; no test passes an annotated-tag sha as a target or token, which the old cat-file -e sha^{commit} accepted |   |
| F21 | mutation | Minor | stats/internal/guard/treemarkers.go:184 | dropping LC_ALL=C on the diff run survives, because the fixtures are ASCII |   |

findings-total: 21
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
finding-status: F14 withdrawn git status --porcelain v1 output is locale-independent by contract, so no fixture can observe LC_ALL there
finding-status: F15 fixed
finding-status: F16 fixed
finding-status: F17 fixed
finding-status: F18 fixed
finding-status: F19 fixed
finding-status: F20 fixed
finding-status: F21 withdrawn diff normal-format output carries no locale text for these inputs, so no fixture can observe LC_ALL there

reproducers-total: 21
finding-reproducer: F1 .superpowers/sdd/reproducers/0-mutation-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-mutation-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-mutation-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-mutation-4.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-mutation-5.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-mutation-6.sh
finding-reproducer: F7 .superpowers/sdd/reproducers/0-mutation-7.sh
finding-reproducer: F8 .superpowers/sdd/reproducers/0-mutation-8.sh
finding-reproducer: F9 .superpowers/sdd/reproducers/0-mutation-9.sh
finding-reproducer: F10 .superpowers/sdd/reproducers/0-mutation-10.sh
finding-reproducer: F11 .superpowers/sdd/reproducers/0-mutation-11.sh
finding-reproducer: F12 .superpowers/sdd/reproducers/0-mutation-12.sh
finding-reproducer: F13 none — usage-check removal, covered by the fix's added cases
finding-reproducer: F14 none — not observable with ASCII fixtures
finding-reproducer: F15 none — covered by the fix's added case
finding-reproducer: F16 none — the code is right; only the change's records are wrong
finding-reproducer: F17 none — not observable
finding-reproducer: F18 none — logical implication, not a runtime defect
finding-reproducer: F19 none — cost depends on tasks.md size at runtime
finding-reproducer: F20 .superpowers/sdd/reproducers/1-mutation-1.sh
finding-reproducer: F21 .superpowers/sdd/reproducers/1-mutation-2.sh

## Pass log

### Round 0

- diff size: 4029 lines, cap 5000, under cap; docs-only: exit 1 (scripts/check-panel-diff-size.sh); roster: full — primary+principles+failure-modes, mutation, extra test-sanity+call-efficiency; no addition this round — the resolved list ran alone; standards: CLAUDE.md, AGENTS.md

### Round 1

- plan-tree verify exit 1 after panel-fix-1: the one change is the fixer's PLAN FIELDS edit of task 6 Tests:, sanctioned by its brief; no damage
- panel-fix-1 (F1-F10) and panel-fix-1-2 (F11-F19) ran on opus/low; F14 auto-decided withdrawn — locale-independent output; plan-tree verify diffs are the fixers' sanctioned PLAN FIELDS/Decision edits
fix-mutation: scripts/test-git-config-pins.sh — skip regex reverted to `(?:exec\.Command|lookPath)\((?:\s*\w+\s*,)?\s*$` (break-and-prove, byte-exact restore) — case 6 failed
fix-mutation: stats/internal/guard/planunchanged.go — verify `return gitFailed()`→`return 1` (0-mutation-2.sh) — TestCheckPlanUnchanged/case 16b
fix-mutation: stats/internal/guard/planunchanged.go — `--untracked-files=all` dropped (0-mutation-3.sh) — TestCheckPlanUnchanged/case 15b
fix-mutation: stats/internal/guard/planunchanged.go — snapshot-failure `os.Remove` dropped (0-mutation-4.sh) — TestCheckPlanUnchanged/case 11b
fix-mutation: stats/internal/guard/treemarkers.go — `-cE`→`-c` (0-mutation-5.sh) — TestCheckTreeMarkers/case 12b,12c
fix-mutation: stats/internal/guard/treemarkers.go — blank-line skip disabled (0-mutation-6.sh) — TestCheckTreeMarkers/case 12a
fix-mutation: stats/internal/guard/treemarkers.go — `ExitCode() > 1`→`> 2` (0-mutation-7.sh) — TestCheckTreeMarkers/case 13a
fix-mutation: stats/internal/guard/guardautosquash.go — HEAD check disabled (0-mutation-8.sh) — TestGuardAutosquash/no resolvable HEAD
fix-mutation: stats/internal/guard/guardautosquash.go — targets `>= 2`→`>= 1` (0-mutation-9.sh) — TestGuardAutosquash/targets with no sha
fix-mutation: stats/internal/guard/guardautosquash.go — token floor 7→6 (0-mutation-10.sh) — TestGuardAutosquash/six-letter hex word
fix-mutation: stats/internal/guard/paneldiffsize.go — default cap 5000 -> 4999 (0-mutation-11.sh) — TestCheckPanelDiffSize case 2b
fix-mutation: stats/internal/guard/foldfixup.go — `after` verdict ignored (0-mutation-12.sh) — TestFoldFixupAfterVerdict/after_exits_1
fix-mutation: stats/internal/guard/foldfixup.go — line 62 `return 2` -> `return 0` (exit >=2 mapped to 0) — TestFoldFixupAfterVerdict/after_exits_3
fix-mutation: stats/internal/guard/planunchanged.go — mode check -> `if false` — TestCheckPlanUnchanged case 9b
fix-mutation: stats/internal/guard/planunchanged.go — name/snap check -> `if false` — TestCheckPlanUnchanged cases 9e/9f
fix-mutation: stats/internal/guard/treemarkers.go — `!isDir(...)` half removed — TestCheckTreeMarkers case 7k
fix-mutation: stats/internal/guard/guardautosquash.go — batch-check verdict `strings.Contains(lines[i], " commit ")` -> `true` — TestGuardAutosquash
fix-mutation: stats/internal/guard/treemarkers.go — none — LC_ALL=C on diff(1) is not observable: diff's normal-format output carries no translated text for these inputs; deleting the line leaves TestCheckTreeMarkers green (measured). It is kept for bash parity, as F17 asks.
fix-mutation: stats/internal/guard/planunchanged.go — none — F14: git status --porcelain is locale-independent
fix-mutations-total: 19

### Round 2

- panel-2-mutation (sonnet/low, rerun pair) re-ran on fix-round-2.diff for F20 — clean, F20 cited fixed; no other slot raised a finding the round touched; no addition this round — the resolved list ran alone
fix-mutation: stats/internal/guard/guardautosquash.go — dropped ^{commit} from the batch-check input — TestGuardAutosquash/targets_accepts_an_annotated_tag's_object_sha
fix-mutations-total: 1
