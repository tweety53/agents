# Panel report — round 4, test-sanity (gymie-frontend, unread tests)

Verdict: no findings.

Coverage: 131 test files changed (29 added, 102 modified) between 3f4ac836 and HEAD. Read every removed @Test (11) together with its production counterpart. Grepped all ~900 added and removed assertion lines for tautologies (assertTrue(true), self-equal, >= 0, @Ignore) and found none. Read ShippedScreenMeasuresUiTest in full, plus the PublicProfile SH6 block and the WorkoutEditorDateTime, EmptyBlocks and BlockNaming hunks. Sampled the mechanical re-pins by pattern instead of reading each one: colour role grey600 -> ink, dp/sp measures, and tag swaps after the preset pills were removed.

Removed tests. Each matches a deliberate removal in production code, not a weakened pin:
- InitialsTest three-word case: initialsOf now returns the first word's initial (Avatar.kt:34). The two-initial behaviour moved to twoInitialsOf.
- ShellBootChrome, ShellLoading and RouteRestore bar-total tests: shell_bar_weigh_in_total and shell_bar_calorie_total no longer exist in commonMain.
- WorkoutEditorDateTime theSheetsLoadMarksForTheShownSpan: commit 8d9ba372 removed the picker's mark fetch (SH10 draws no marks; marks = emptyMap()).
- Preset-pill tests (WorkoutEditorSheets, BlockNaming): add_text_preset_* is no longer drawn, and the helper now uses performTextInput.
- Clock hollow-ring test: replaced by PickerSheetsFidelityUiTest.

Exact-value pins (for example 31.5 ± 0.5 from helper to footer, and the 273dp sheet) state their frame arithmetic in KDoc (PublicShell.kt:173, ShippedScreenMeasuresUiTest.kt:113). The tests keep their expected values in their own constants instead of reusing production constants, so they fail if the layout moves.
