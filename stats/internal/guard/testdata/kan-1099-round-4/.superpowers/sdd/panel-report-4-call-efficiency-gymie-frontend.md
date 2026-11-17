Critical: 0
Important: 0
Minor: 0
Unread: gymie, gymie-frontend, gymie-playwright, gymie-admin-frontend

### Summary

Scope: the `gymie-frontend` section. Reads and effects the diff adds or changes, found by diffing `shared/src/commonMain` for API, sync, state and `LaunchedEffect`/`scope.launch` changes:
- `EditProfileState.load` now also calls `TrainingSetupApi.getTrainingSetup` (SH29 Experience row), running alongside the profile load. Its only other caller is `ui/settings/TrainingSetupState.kt:60`, which belongs to a different screen, so there is no second caller on the same user action. `saveExperience` PATCHes only when the value changed.
- `QueueingWorkoutApi.updateWorkout`: an online FINISH with nothing queued is now sent at once (`sendFinish`) instead of being enqueued. On success it returns without enqueuing; on any failure other than a 422 it falls through to the queue. Either way the server receives one request. No duplicate.
- `GymieApiClient`: the timeout classification only changes which failures report a write failure. It adds no call.
- `WeighInHistoryState.delete`: same call count as before (delete, then refetch); it now hides the row optimistically and refetches in `finally`.
- `CalorieHistoryScreen.deleteWithUndo`: `deleteEntry` + `refetch` replaces the old confirm path, which made the same calls. Undo makes one save + one refetch. `RefetchOnBodySync` fires only on sync `bodySettled` events, as it did before the diff.
- `ENTRIES_PAGE_SIZE` / `HISTORY_PAGE_SIZE` drop from 60 to 50. That shifts payload size per page, not the number of calls per screen.
- The new `LaunchedEffect`s (DOB `bringIntoView`, focus requests, `weightField`) are local UI effects with no boundary call.

### Issues

#### Critical (Must Fix)
None.

#### Important (Should Fix)
None.

#### Minor (Nice to Have)
None.

### Assessment
**Ready for the human gate?** Yes
**Reasoning:** Every read or effect the diff adds has one caller per user action, and none runs per item or on an unstable key.
