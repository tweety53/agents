Critical: 0
Important: 0
Minor: 0
Unread: gymie-frontend, gymie-playwright

### Summary
The gymie section's only read change widens the two existing ledger page native queries in `CalorieHistoryJpaDao` (`activity`, `goal_pct`, `protein_option`, `fat_pct` from `calorie_target`, NULL padding on the `not_counted_period` arm). There is no new query, no extra round-trip, and no per-row load: `targetProvenance(...)` is a pure in-memory mapping on the projection already fetched. `toDto()` is likewise pure. Lenses second caller, repeated within one request, per item of a list and unstable trigger have nothing to count here.

Coverage: read the full gymie `src/` diff; spectre spec and README hunks add no reads and were not read. The frontend and playwright sections belong to other dispatches.

### Issues
#### Critical (Must Fix)
None.
#### Important (Should Fix)
None.
#### Minor (Nice to Have)
None.

### Assessment
**Ready for the human gate?** Yes
**Reasoning:** The diff adds columns to a query that already runs, not calls.
