Critical: 0
Important: 0
Minor: 0
Unread: gymie-frontend, gymie-playwright

# Panel round 4 — primary — gymie section

Coverage: read the whole gymie `# worktree:` section of final-review.diff (README, 8 spectre specs, 7 main + 3 test Kotlin files) and CalorieLedger.kt in full; other sections belong to their own dispatches.

### Shared-state check
- `LedgerEntry` (new field `provenance`): writers CalorieLedger.kt:79 (TARGET — now sets `provenance = row.provenance`) and CalorieLedger.kt:98 (NOT_COUNTED — default null). Both correct: a period never has provenance.
- `LedgerRow.provenance`: sole writer CalorieHistoryRepositoryAdapter.kt:50, via `targetProvenance(...)`; both DAO queries select the four columns for TARGET and NULLs for NOT_COUNTED (column counts match, 14 each). Correct.
- `ProvenanceDto` construction: `TargetProvenance.toDto()` replaces the inline build in CalorieCalculatorWebConverters; same field order and types. Correct.
- `CalorieTargetEntity.toDomain` provenance: extracted to `targetProvenance`, behaviour identical (same all-four-non-null rule). Correct.

### Issues
#### Critical (Must Fix)
None.
#### Important (Should Fix)
None.
#### Minor (Nice to Have)
None.

### Recommendations
- None. The ledger now exposes calculator provenance so a deleted target's Undo can re-create it whole; covered at domain, adapter (real Postgres) and controller levels. README carries the user-facing text rules / SE11 note; it does not document ledger response fields, so no README update owed for the additive `provenance` field.

### Assessment
**Ready for the human gate?** Yes
**Reasoning:** Additive, nullable response field with a behaviour-preserving extraction; tests exercise the real query.
