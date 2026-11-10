# F24 — Important — primary — round 1

**Location:** gymie-frontend:shared/src/commonMain/kotlin/com/gymie/shared/ui/workout/WorkoutEditorScreen.kt:1858

`onAsked(ids)` runs before the lookup suspends, so on the next recomposition `missing` turns false, re-keys LaunchedEffect(ids, missing) and cancels the request in flight; after a stale edit an exercise added elsewhere still shows Unknown exercise on any real network
