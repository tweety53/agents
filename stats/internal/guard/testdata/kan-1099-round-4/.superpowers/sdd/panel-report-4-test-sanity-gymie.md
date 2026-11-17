Critical: 0
Important: 0
Minor: 0
Unread: gymie-frontend, gymie-playwright

### Summary
The gymie section adds three tests and changes none otherwise:
- `CalorieLedgerTest` — "a target's provenance rides its ledger entry…": asserts `CalorieLedger.assemble` copies `LedgerRow.provenance` onto the `LedgerEntry`. `LedgerEntry.provenance` defaults to null, so dropping `provenance = row.provenance` in `CalorieLedger.kt` makes the assertion fail. Can fail; honest name.
- `CalorieRepositoryAdapterTest` — "carries a calculator-made target's provenance on the ledger page, and none on a period": runs the real native UNION query against Postgres; fails if the SELECT drops the four columns or the adapter stops calling `targetProvenance`. The hand-entered target and the period rows pin the null branch. Can fail; honest name.
- `BodyCaloriesControllerTest` — four jsonPath asserts on `entries[1].provenance.*`; fail if `BodyCaloriesWebConverters` stops mapping `provenance` (DTO default null). Can fail.
No duplicates found among the existing ledger/adapter/controller tests. Judged from code; no scratch-copy mutation was run because each test's failure path is a single dropped assignment that the default-null makes observable.

Coverage: read in full the gymie `src/` diff hunks; spectre spec and README hunks hold no tests and were not read. The frontend and playwright sections belong to other dispatches.

### Issues
#### Critical (Must Fix)
None.
#### Important (Should Fix)
None.
#### Minor (Nice to Have)
None.

### Assessment
**Ready for the human gate?** Yes
**Reasoning:** Every added test fails when the behaviour it names breaks.
