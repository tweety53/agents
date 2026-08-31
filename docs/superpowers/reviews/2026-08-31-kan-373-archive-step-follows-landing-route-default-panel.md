# Review panel — kan-373-archive-step-follows-landing-route-default

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | principles | Critical | skills/flow/integrate.md:146-152 | the literal prompt block hardcodes Merge and push as (default, recommended) in the shared/generic skill file, contradicting the very next sentence that says the marker moves to whichever option resolves per-project — a reader of the template outside this repo would see this repo's own resolved answer baked in as if it were the universal default |
| F2 | principles | Minor | CLAUDE.md:125 | CLAUDE.md's digest line still states open PR as the unconditional default, now stale for this repo since its own .flow/project.md resolves to merge and push |
| F3 | principles | Minor | skills/flow-contracts/finish-contract-run2.md:217-236 | archive.md's step 10 restates the full two-row table and failure prose near-verbatim from finish-contract-run2.md, with no canonical-elsewhere disclaimer the way the neighboring step 9 already carries one |

findings-total: 3
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed

reproducers-total: 3
finding-reproducer: F1 none — documentation self-contradiction, verified by direct comparison of skills/flow/integrate.md lines 146-152
finding-reproducer: F2 none — read CLAUDE.md:125 next to .flow/project.md's new ## default landing route section
finding-reproducer: F3 none — side-by-side reading of the two hunks
