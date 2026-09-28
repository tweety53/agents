# Review panel — kan-850-agents-port-the-visual-verify-scripts-to-the-go

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | Important | stats/internal/guard/visualsection.go:151 | The fact "the awk and grep the scripts ran end a line at its first NUL" is applied in two callers' own code instead of in vvLines, the reader all three guards share; on one file whose heading line holds a NUL the guards answer verification 0 / trigger 2 / resolve 2, where the bash answered 0 / 0 / 0. |   |
| F2 | principles | Important | stats/internal/guard/composemockupframes.go:423 | Compose re-implements Python semantics the package already defines once (cmfIsSpace/cmfStrip = pyIsSpace/pyStrip/pySplit, cmfOSError = ppOSError, cmfRepr = ppRepr, cmfPILError's sniff = mvpImageError), and the copies have already drifted: for the same non-UTF-8 path compose prints 'x\udcff.png' and measure prints 'x<U+FFFD>.png'. |   |
| F3 | primary | Minor | stats/internal/guard/composemockupframes.go:132 | Every stdin capture is decoded and kept in captures for the whole run, where the Python loaded each one only to validate it; peak memory now grows with the length of the stdin list (20 captures of 2000x4000: 785 MB vs 209 MB). |   |
| F4 | primary | Minor | stats/internal/guard/testdata/compose-mockup-frames/map-not-utf8/map.mockups:1 | The non-UTF-8 fixture makes check-task-commit-fields.sh die with UnicodeDecodeError (exit 1) on task 7's own commit 001b5a6e, so the task can never be re-verified by its guard. |   |
| F5 | primary | Minor | spectre/changes/kan-850-agents-port-the-visual-verify-scripts-to-the-go/proposal.md:23 | "CLI contracts unchanged ... with three stated departures" no longer describes the change: design.md records six more accepted departures. |   |
| F6 | primary+principles | Minor | stats/internal/guard/resolvevisualscreenshots.go:59 | A read failure after the access check prints a new wording (reading <cfg> ... failed (<err>)) where the bash and both sibling ports print 'grep exited 2 while looking for the ... heading in <cfg>'; no design decision records the departure. |   |
| F7 | principles | Minor | stats/internal/guard/visualtrigger.go:37 | in, _ = io.ReadAll(env.Stdin) discards the read error, so a stdin the guard cannot read becomes "no changed paths" and the guard prints a clean VISUAL-TRIGGER-NO-MATCH verdict; the sibling compose refuses the same failure with exit 2. |   |

findings-total: 7
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed

reproducers-total: 7
finding-reproducer: F1 .superpowers/sdd/reproducers/0-principles-2.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: F6 none — reachable only by a read failing between access(2) and open(2), a race not constructible from the command line
finding-reproducer: F7 .superpowers/sdd/reproducers/0-principles-3.sh

## Pass log

### Round 0

- roster: compact — 58
- diff size: 19881 changed lines, over cap — proceeded automatically (mostly testdata goldens/PNGs; 48 non-testdata paths)
- docs-only: exit 1 (first non-doc path scripts/check-visual-trigger.sh) — roster dispatched: primary+principles, one dispatch, opus/high
- no addition this round — the resolved list ran alone
- standards passed: CLAUDE.md, AGENTS.md (absolute in the canonical worktree); citation check: scripts/check-references.sh captured to .superpowers/sdd/citation-check.md; no relocation-comparison.md
- bounced F3 (captures retained for the whole run): reproducer read not-demonstrated (593149952 B vs 3x209354752 B threshold, uniform fixtures compressed by macOS); F4 (non-UTF-8 fixture breaks check-task-commit-fields): demonstrates citation <0xff> did not resolve. Both re-authored by primary (panel-0-primary-bounce); exit contract now OK for 6 runnable findings.
- panel-fix-0 (opus/medium, sdd fixer pair): fixed F1-F7 in fe89bb88..db641d42, diff .superpowers/sdd/fix-round-0.diff (665 lines). F4, F5 flipped and touch named paths (.gitattributes the route primary named; proposal.md) -> fixed; F6 closes on path condition (resolvevisualscreenshots.go:59). F1, F2, F3, F7 re-runs refused: premise lines cite code the fix deleted (KAN-839 route) -> re-authoring by raising slots in the re-run. Parent added task 7/8 corrections for .gitattributes, planprovenance.go and the 67/79 subtest counts.
fix-mutation: stats/internal/guard/visualsection.go — vvLines' NUL cut replaced by a no-op — TestVisualSection/NUL_ends_the_heading_line, TestVisualSection/NUL_ends_a_section_line
fix-mutation: stats/internal/guard/planprovenance.go — ppRepr's \udcXX surrogate branch disabled — TestMeasureVisualProperties/a_non-UTF-8_path_reprs_as_Python's_surrogate_in_both_ports
fix-mutation: stats/internal/guard/pngrgb.go — shared pilImageError "cannot identify image file" text altered — TestMeasureVisualProperties/golden_exit2-not-an-image, TestComposeMockupFrames/case_8 and case_13
fix-mutation: stats/internal/guard/composemockupframes.go — stdin captures retained again (kept slice + KeepAlive) — TestComposeMockupFrames/stdin_captures_are_validated,_not_retained
fix-mutation: stats/internal/guard/visualtrigger.go — stdin read error swallowed again — TestCheckVisualTrigger/an_unreadable_stdin_is_a_refusal,_not_a_verdict
fix-mutation: .gitattributes — -diff entry removed — 1→0 check-task-commit-fields failures on task 7 (0-primary-3.sh)
fix-mutation: stats/internal/guard/resolvevisualscreenshots.go — none — reachable only through an access(2)/open(2) race
fix-mutation: stats/internal/guard/composemockupframes.go — none — per-line re-decode failure reachable only if a capture changes between validation and use
fix-mutations-total: 8

### Round 1

- primary (opus/low, targeted: F1 F3 F4 F5 F6) and principles (opus/low, targeted: F1 F2 F6 F7) re-ran alone on .superpowers/sdd/fix-round-0.diff (665 lines, under cap; branch not docs-only). Both clean: every finding fixed, no new finding. Reproducers for F1 F2 F3 F7 re-authored in place (premises on surviving lines), prove-reproducer PROOF HELD vs 0a15948d; parent pinned re-runs flipped (runner exit 1) -> fixed. Base CLEAR at the round boundary.
- check-panel-fix-single-dispatch exit 1: out-of-shape panel-fix key 'full-suite-fix-1' (task 10's full-suite fix dispatch, recorded -role panel-fix during flow.sdd-tdd) -- auto-resolved Continue, violation kept in this run's output
