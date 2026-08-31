# Review panel — kan-288-skip-dispatch-context-bundle-rebuild-unchanged

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | principles | Important | scripts/gather-dispatch-context.sh:116-148 | sha256_hex() is copied verbatim from scripts/check-cleanup-complete.sh instead of extracted to scripts/lib/, which is this codebase's own convention once a second caller exists |
| F2 | primary | Minor | scripts/gather-dispatch-context.sh:150,194 | capturing BODY via command substitution strips the bundle's trailing newline, an undocumented byte-level change from the pre-change stdout-redirected output |
| F3 | bugbot | Important | scripts/gather-dispatch-context.sh:495 | no test covers stale .hash cleanup on the no-hash-tool path; removing rm -f $HASH_PATH there is a surviving mutant that lets a later run wrongly trust a stale hash and reuse content that no longer matches the inputs |
| F4 | bugbot | Minor | scripts/gather-dispatch-context.sh:472 | no test covers the OUTPUT_PATH-missing half of the no-cached-bundle OR-guard; dropping it is a surviving mutant, though the shipped guard itself is already correct |
| F5 | bugbot | Important | scripts/lib/sha256-hex.sh:40 | no test isolates an openssl-only environment (excluding shasum and sha256sum); weakening the 64-hex-char shape filter to a bare first-field grab is a surviving mutant that would silently corrupt the hash on such a machine |
| F6 | bugbot | Minor | scripts/gather-dispatch-context.sh:476 | the F2 trailing-newline fix has no permanent regression test in the shipped suite; dropping the fix's own printf '%s\n' back to printf '%s' is a surviving mutant, only caught by this round's throwaway .superpowers/sdd/repro-f2-trailing-newline.sh reproducer, not by the shipped test file |

findings-total: 6
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed

reproducers-total: 6
finding-reproducer: F1 none — structural DRY finding (duplicated sha256_hex helper), no pass/fail check demonstrates it; verified by inspection against scripts/check-cleanup-complete.sh:516-532
finding-reproducer: F2 .superpowers/sdd/repro-f2-trailing-newline.sh
finding-reproducer: F3 none — requires editing the script to remove rm -f $HASH_PATH, then a 3-call sequence; see panel report
finding-reproducer: F4 none — requires editing the script to drop the [ ! -f "$OUTPUT_PATH" ] disjunct, then delete the bundle file while keeping .hash; see panel report
finding-reproducer: F5 none — no fixture isolates an openssl-only PATH; shipped shape-check is already correct but has zero coverage of that branch
finding-reproducer: F6 none — every content assertion in scripts/test-gather-dispatch-context.sh reads the file via $(cat ...), which strips trailing newlines and cannot distinguish the fixed shape from the F2 defect
