# Review panel — kan-372-flow-the-router-eager-loads-11-contracts-before

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Code review (low) | Important | CLAUDE.md:86 | CLAUDE.md, AGENTS.md, and README.md still cite skills/flow-contracts/finish-contract.md, which Task 2 deleted and split into finish-contract-run1.md and finish-contract-run2.md; check-references.sh does not catch this because it only validates section anchors inside files that exist, not that a cited file itself exists |
| F2 | Code review (low) | Minor | scripts/gather-dispatch-context.sh:379-388 | the refused: <label> (resolves outside the change directory) message is printed for every outside-refusal regardless of source, but the new project-commands source is checked against WORKTREE_REAL not CHANGE_ROOT_REAL, so the message would name the wrong root if that source is ever refused |
| F3 | Principles | Critical | skills/flow-contracts/artifacts-registry.md:56-57 | cites the Run 1 section against finish-contract-run2.md, but that heading now lives only in finish-contract-run1.md — Task 2 misdirected this one citation |
| F4 | Bugbot | Important | scripts/gather-dispatch-context.sh:337 | surviving mutant: mutating the section-boundary exit condition from /^## / to /^### / makes a ## lint section swallow the following ## test section whole, and the test suite still passes — no case asserts the boundary stops at the next top-level heading |
| F5 | Bugbot | Important | scripts/gather-dispatch-context.sh:349 | surviving mutant: weakening the emptiness check from a whitespace-stripped test to a bare -n test makes a whitespace-only section body wrongly emitted as populated, and the test suite still passes — no case covers a whitespace-only section body |
| F6 | Bugbot | Important | scripts/gather-dispatch-context.sh:336 | surviving mutant: weakening the awk key match from == to a substring/regex ~ match makes a ## linter heading wrongly captured as the ## lint section, and the test suite still passes — no case covers a heading that is a superstring of a real key |

findings-total: 6
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed

reproducers-total: 6
finding-reproducer: F1 none — grep for the deleted filename across the repo: grep -rn 'finish-contract\.md' CLAUDE.md AGENTS.md README.md
finding-reproducer: F2 none — code-review finding: read the refusal-message branch and compare which root each REFUSED_REASONS[i]=="outside" source was actually checked against
finding-reproducer: F3 grep -A1 'Run 1 — the branch is not merged' skills/flow-contracts/artifacts-registry.md; grep -c 'Run 1 — the branch is not merged' skills/flow-contracts/finish-contract-run2.md
finding-reproducer: F4 sed -i.bak 's|/\^## /|/^### /|' scripts/gather-dispatch-context.sh; scripts/test-gather-dispatch-context.sh; mv scripts/gather-dispatch-context.sh.bak scripts/gather-dispatch-context.sh
finding-reproducer: F5 none — manual mutation test: weaken the emptiness check to a bare -n test and confirm scripts/test-gather-dispatch-context.sh still passes on a whitespace-only section body
finding-reproducer: F6 none — manual mutation test: weaken the key match from exact-line equality to substring/regex match and confirm scripts/test-gather-dispatch-context.sh still passes with a heading like ## linter present
