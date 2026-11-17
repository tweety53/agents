Critical: 0
Important: 0
Minor: 0
Unread: gymie-frontend, gymie-playwright

# Panel round 4 — principles — gymie section

Coverage: whole gymie section of final-review.diff read; standards: CLAUDE.md managed Kotlin backend standard (layout, core purity), CONTRIBUTING suppression policy.

### Derived-state check
- `LedgerEntry.provenance` / `LedgerRow.provenance`: derived on read from the four calculator columns (CalorieHistoryRepositoryAdapter.kt:50 via `targetProvenance`), never stored. Derives from every fact.
- `fromCalculator` (pre-existing, SQL `activity IS NOT NULL`) overlaps `provenance != null`; both are computed per query, not stored, so no SSOT break. Kept for wire compatibility — defensible.

### Summary
Diff adds calculator provenance to the calorie ledger rows. DRY improved: the provenance-from-columns rule and the DTO mapping each extracted to one function (`targetProvenance`, `TargetProvenance.toDto()`), each now with two callers. Searched `src/app/tracking` for other `TargetProvenance(` / `ProvenanceDto(` builders — none remain. Layer purity: core CalorieLedger imports only `core.domain.calorietarget.TargetProvenance`; no Spring/JPA in core. No new suppressions, no lint config touched. Complies.

### Issues
#### Critical (Must Fix)
None.
#### Important (Should Fix)
None.
#### Minor (Nice to Have)
None.

### Assessment
**Principles-compliant?** Yes
**Reasoning:** Derived, not stored; duplication removed; layering respected.
