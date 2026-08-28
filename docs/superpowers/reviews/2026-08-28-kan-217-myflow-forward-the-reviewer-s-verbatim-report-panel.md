# Review panel — kan-217-myflow-forward-the-reviewer-s-verbatim-report

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | Minor | skills/flow-contracts/artifacts-registry.md:35 | the added sentence restates a fact already established as canonical at skills/flow-contracts/finish-contract.md:576 ("<abs-worktree>/.superpowers/sdd/ records are all ignored and all irreplaceable"), applied here to one row when the identical .superpowers/sdd/ location already covers the pre-existing "Per-task and review diffs" row one line above without any such callout |
| F2 | Code review (low) | Major | scripts/check-dispatch-paragraphs.sh:160 | block-text extraction bleeds across adjacent blockquote paragraphs: the continuation walk only checks whether the next line starts with a quote marker and never stops at a second label line, so a required block immediately followed by any other blockquote content containing the missing phrase silently satisfies the phrase check by absorbing the neighbour's text |
| F3 | Code review (low) | Minor | scripts/test-check-dispatch-paragraphs.sh:25 | the header comment's mutation-testing claim is false as written: it states every case targeting a specific guard check is mutation-tested against a throwaway copy of the guard, and no such mechanism exists anywhere in the file; the claim is new in this diff and scripts/test-check-plan-shape.sh:20 now cites this file as already meeting that requirement, propagating the overclaim |
| F4 | Bugbot | Major | scripts/check-dispatch-paragraphs.sh:80 | the set-but-empty env-override refusal is a surviving mutant: deleting the whole branch, folding empty into use-the-default-root, still passes all 11 test cases and 22 assertions |
| F5 | Bugbot | Major | scripts/check-dispatch-paragraphs.sh:199 | the not-a-regular-file refusal is a surviving mutant with zero fixture coverage: the harness never places a directory, FIFO or device at a required site path, so deleting the check leaves all cases green |
| F6 | Bugbot | Major | scripts/check-dispatch-paragraphs.sh:202 | the not-readable refusal is likewise a surviving mutant: no fixture ever chmods a site file unreadable, so deleting the check leaves all cases green |
| F7 | Bugbot | Minor | scripts/check-dispatch-paragraphs.sh:105 | the review-panel.md REPRODUCE site's min-blocks value of 1 is unenforced by any test independent of the variant check: setting it to 0 still passes every case, because every fixture that zeroes that site's block count also fails the reviewer variant requirement, masking the min-blocks check entirely |
| F8 | Bugbot | Minor | scripts/check-dispatch-paragraphs.sh:105 | the implement.md site's min-blocks value of 2 has the same problem: lowering it to 1 still passes all cases, since the one-block case and the two-same-variant-blocks case both fail on the missing implementer variant, never on the block count itself |

findings-total: 8
finding-status: F1 fixed
finding-status: F2 withdrawn the guard's own WHAT A GREEN RUN DOES NOT PROVE section already discloses this class, F2 is a sharper instance of that disclosed gap, no Markdown-structural signal separates a long legitimate paragraph from one diluted with foreign prose, and the real corpus already spreads its required phrases across two sentences so every candidate tightening risks rejecting a legitimate rewording
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed

reproducers-total: 8
finding-reproducer: F1 grep -n gitignore skills/flow-contracts/artifacts-registry.md
finding-reproducer: F2 none — the demonstrating fixture requires a heredoc, whose shell metacharacters check-panel-reproducers.sh refuses; the dispatcher ran the slot-supplied command by hand and confirmed exit 0 on a deficient fixture
finding-reproducer: F3 grep -n mutation-tested scripts/test-check-dispatch-paragraphs.sh
finding-reproducer: F4 none — a surviving-mutant demonstration requires mutating the guard and re-running the harness, which is two steps and not a single runnable command; the dispatcher confirmed the mutation table the slot reported
finding-reproducer: F5 none — a surviving-mutant demonstration requires mutating the guard and re-running the harness, which is two steps and not a single runnable command; the dispatcher confirmed the mutation table the slot reported
finding-reproducer: F6 none — a surviving-mutant demonstration requires mutating the guard and re-running the harness, which is two steps and not a single runnable command; the dispatcher confirmed the mutation table the slot reported
finding-reproducer: F7 none — a surviving-mutant demonstration requires mutating the guard and re-running the harness, which is two steps and not a single runnable command; the dispatcher confirmed the mutation table the slot reported
finding-reproducer: F8 none — a surviving-mutant demonstration requires mutating the guard and re-running the harness, which is two steps and not a single runnable command; the dispatcher confirmed the mutation table the slot reported
