# Review panel — kan-451-add-a-per-project-incident-table-and-a-guard

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | primary | Major | scripts/check-unfinished-work.sh:351 | the write's stderr is fully discarded (>/dev/null 2>&1), contradicting design.md §6's stated stdout-only redirect that lets the store-unreachable warning pass through on stderr |
| F2 | principles | Major | scripts/check-unfinished-work.sh:351 | advisory write's failure is fully swallowed rather than merely non-blocking, per the diff's own header comment admitting both streams are discarded against design.md's explicit stdout-only intent |
| F3 | code-review-low | Minor | scripts/gather-dispatch-context.sh:95-98 | a caller-supplied path (e.g. PRINCIPLES_PATH) containing a literal ( silently loses its (absent) suffix in the skipped-line rendering, since the case match tests for any ( rather than the deliberate incidents-label shape |
| F4 | mutation | Minor | scripts/gather-dispatch-context.sh:453 | the outer 'command -v flow' gate is redundant with the inner error-handling chain that already falls back to the same skipped (flow unavailable) outcome, so the gate is dead weight no test asks for |
| F5 | primary | Minor | spectre/changes/kan-451-add-a-per-project-incident-table-and-a-guard/tasks.md | task 5's Tests/Baseline fields are stale after CASE 48 was folded into the same commit — still list Case 44-47/after=71 instead of 44-48/after=73 |
| F6 | code-review-low | Major | scripts/gather-dispatch-context.sh:504 | the narrowed suffix-matching pattern still misfires for a legitimately-absent PRINCIPLES_PATH ending in a parenthesized segment like notes (v2), silently dropping the (absent) suffix — the same shape-sniffing ambiguity as F3, narrowed but not eliminated |
| F7 | mutation | Major | scripts/gather-dispatch-context.sh:504 | dropping the required leading space in the suffix pattern (foo(bar), no space before the paren) also silently drops the (absent) suffix and is not caught by CASE 48, confirming the shape-sniffing heuristic itself is unreliable rather than merely under-tested |
| F8 | mutation | Minor | scripts/gather-dispatch-context.sh:503 | CASE 45/46 assert via unanchored substring match, so a regression that double-appends (absent) onto the already-self-describing incidents labels would go undetected |
| F9 | principles | Minor | scripts/gather-dispatch-context.sh | no in-script comment documents the format-at-push convention for SKIPPED_LABELS, so a future sixth push site could reintroduce the same defect class silently; the rationale currently lives only in the test file's CASE 49 comment |

findings-total: 9
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed

reproducers-total: 9
finding-reproducer: F1 none — unverifiable: check-panel-reproducers.sh refuses the grep-based command (shell metacharacters); defect confirmed by direct file inspection instead
finding-reproducer: F2 none — same defect as F1, reproduced there; principles review adds the observability-principle framing
finding-reproducer: F3 none — unverifiable: check-panel-reproducers.sh refuses the multi-line setup script (shell metacharacters, parens in the constructed path); defect confirmed by direct code inspection instead
finding-reproducer: F4 none — behaviorally redundant, no observable difference; bash scripts/test-gather-dispatch-context.sh passes unchanged with the outer command -v flow gate removed
finding-reproducer: F5 none — plan-doc field drift, not a runnable command; bash scripts/test-gather-dispatch-context.sh shows 73 ok lines vs the declared Case 44-47/after=71
finding-reproducer: F6 none — shell metacharacters (parens, space) in the constructed principals-path make a bare-path reproducer impossible; reproduced directly by the reviewer with a real invocation
finding-reproducer: F7 none — shell metacharacters (unescaped parens) in the constructed principals-path make a bare-path reproducer impossible; reproduced directly by the reviewer with a real invocation
finding-reproducer: F8 none — CASE 45/46's own unanchored grep -qF substring match is the gap; reproduced by the reviewer reverting render_body's loop to unconditional (absent) appending and confirming the harness still exits 0
finding-reproducer: F9 none — absence of a comment, not a runnable behavior
