# Review panel — kan-445-manual-incident-recovery-pattern-was-correct

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | principles | Important | scripts/recover-guard-incident.sh:57-63 | repo-dir is never resolved to the toplevel, so path... resolves relative to the given directory instead of the repo root — a false no-files diagnosis in exactly the pressured-incident context this tool exists for (also raised Medium by code-review-low) |
| F2 | primary | Low | docs/superpowers/specs/2026-09-07-kan-445-manual-incident-recovery-pattern-was-correct-design.md:38-40 | the design precondition heading contradicts its own item 1 and the implementation — fix the wording, not the code (also raised by principles and code-review-low) |
| F3 | principles | Low | scripts/recover-guard-incident.sh:108 | substring match marks absent files as overwrites — compare newline-anchored instead (also raised by code-review-low) |
| F4 | principles | Low | scripts/recover-guard-incident.sh:109 | the dry-run plan appends the overwrite annotation to the mkdir command text itself, so the printed plan is not runnable as printed |
| F5 | principles | Low | scripts/test-recover-guard-incident.sh:68-254 | the overwrite-annotation behavior has zero coverage; a mutation deleting it keeps the suite green (mutation slot M8 concurs — overwrite naming droppable) |
| F6 | code-review-low | Low | scripts/recover-guard-incident.sh:43 | die() expands the exit-code argument into the stderr message, so the code lands in the sentence a mid-incident operator reads (primary noted the same, informational) |
| F7 | principles | Low | scripts/recover-guard-incident.sh:53-54 | only --help/--apply in position 1 are parsed; an unknown flag after repo-dir falls through to a precondition failure instead of the usage-error exit |
| F8 | mutation | Medium | scripts/recover-guard-incident.sh:106-108 | M9 survives: the cat-file -e HEAD half of the tracked-target guard is droppable — no case exercises a target tracked in HEAD only; dropping it lets the tool plan a restore over committed content, the exact refuse-tracked-targets violation the design rules out |
| F9 | mutation | Low | scripts/recover-guard-incident.sh:85 | M17 survives: no case passes explicit path arguments, so a tool that ignores user paths entirely is green — the path... half of the CLI contract is untested |
| F10 | mutation | Low | scripts/recover-guard-incident.sh:118 | M14 survives: the empty-restore-set refusal is untested — a silent partial recovery is possible |
| F11 | mutation | Low | scripts/recover-guard-incident.sh:121 | M12 survives: no case bounds the reflog block length promised as the last 15 entries |
| F12 | primary | Low | docs/superpowers/specs/2026-09-07-kan-445-manual-incident-recovery-pattern-was-correct-design.md:45-47 | reported file by file overstates the tracked-target refusal — the tool reports only the first offending file; reword to match the implemented refusal (mirrored in spectre design.md:33-35) |
| F13 | primary | Informational | spectre/changes/kan-445-manual-incident-recovery-pattern-was-correct/tasks.md:76-78 | tasks.md dry-run case description misstates its own fixture — the implemented harness asserts both files correctly, the plan sentence is loose |
| F14 | primary | Low | scripts/recover-guard-incident.sh:30 | header comment at line 30 says a tracked restore target refuses the whole run file by file, but the code and the round's own corrected design docs say it refuses on the first offending file only (raised independently by primary as F-1 and principles as finding 1) |
| F15 | primary | Low | spectre/changes/kan-445-manual-incident-recovery-pattern-was-correct/tasks.md:142,172,187,314 | tasks.md correctly says thirteen cases in several places this round edited but left six in four other places (lines ~142, ~172, ~187, ~314), so the file now disagrees with itself about its own harness case count (raised independently by primary as F-2 and principles as finding 2) |
| F16 | code-review-low | Medium | scripts/test-recover-guard-incident.sh:289-315 | case 9 (HEAD-tracked-target-refusal) means to pin the cat-file -e HEAD:$f half of the tracked-target guard, but stash_planning_files's git stash -q -u resets the target back into the index before the tool runs, so the refusal actually fires via the ls-files (index) half — deleting the cat-file -e HEAD:$f disjunct leaves case 9 green (code-review-low's D1) |
| F17 | primary | Medium | spectre/changes/kan-445-manual-incident-recovery-pattern-was-correct/tasks.md:172,187,190 | the F15 fix changed six to thirteen in most places but line 190's measured-provenance comment still says six FAIL lines, contradicting line 187's own thirteen right above it; worse, neither 13 nor 6 is the actual FAIL-line count (measured 38), since 13 is the case count not the FAIL: line count |

findings-total: 17
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

reproducers-total: 17
finding-reproducer: F1 none — multi-command mktemp sandbox repro quoted verbatim in .superpowers/sdd/panel-report-0-code-review-low.md section 1 (run tool at a subdirectory: rc=1 false refusal; --apply aborts then fails the restore, leaving a zero-byte plan.md)
finding-reproducer: F2 none — documentation wording; the precondition heading promises exit 1 while precondition 1 (non-repo) exits 2, as the tool and harness case 1 assert
finding-reproducer: F3 none — multi-command sandbox repro quoted verbatim in .superpowers/sdd/panel-report-0-code-review-low.md section 4 (with only kan-plan.md.bak present, the dry-run marks kan-plan.md, which does not exist, as an overwrite)
finding-reproducer: F4 none — scenario in .superpowers/sdd/panel-report-0-principles.md Minor 4 (dry-run, paste the annotated mkdir line, shell syntax error at the parenthetical)
finding-reproducer: F5 none — mutation repro in .superpowers/sdd/panel-report-0-principles.md Minor 5 (deleting every OVERWRITES/MARK line from a temp copy keeps the suite 33/33 green)
finding-reproducer: F6 none — single call on a nonexistent directory shows the stray exit code in stderr (report section 2)
finding-reproducer: F7 none — bash scripts/recover-guard-incident.sh <repo> --bogus exits 1 where the header and design promise 2 (report Minor 2)
finding-reproducer: F8 none — full mktemp repro script quoted verbatim in .superpowers/sdd/panel-report-0-mutation.md F1 (target committed to HEAD then git rm --cached: original refuses, HEAD-half-dropped mutant plans the clobber)
finding-reproducer: F9 none — scenario in .superpowers/sdd/panel-report-0-mutation.md F3 (stash ^3 holds notes/incident.md; tool run with notes: mutant discards the user path and dies naming defaults)
finding-reproducer: F10 none — scenario in .superpowers/sdd/panel-report-0-mutation.md F4 (revert in progress, ^3 populated, path the ^3 does not hold: mutant exits 0 and with --apply aborts the revert having restored nothing)
finding-reproducer: F11 none — scenario in .superpowers/sdd/panel-report-0-mutation.md F5 (reflog -n 15 changed to -n 2 survives; the diagnosis can hide the alternating reset pattern)
finding-reproducer: F12 none — documentation wording; two staged tracked targets name only the first on stderr (primary report section 2)
finding-reproducer: F13 none — plan-prose wording; the dry-run bullet says every restore lands under docs/superpowers/ while the fixture second path is spectre/changes/kan-423/tasks.md
finding-reproducer: F14 none — the header comment's stale wording (still says the tracked-target refusal is reported file by file) is a doc-only defect; a grep-based reproducer needs a pipe/quotes the reproducer shape bans
finding-reproducer: F15 none — stale prose count in a planning doc; a grep-based reproducer needs a pipe the reproducer shape bans
finding-reproducer: F16 none — demonstrating the surviving mutant requires a sed-built mutant file and an env override, both needing shell metacharacters the reproducer shape bans
finding-reproducer: F17 none — demonstrating the wrong FAIL-line count requires a shell pipe/redirect the reproducer shape bans
