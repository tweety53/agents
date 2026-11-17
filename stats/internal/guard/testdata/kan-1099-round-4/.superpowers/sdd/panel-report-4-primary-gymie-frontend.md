Critical: 0
Important: 1
Minor: 1
Unread: gymie-frontend

# Panel round 4 — primary — gymie-frontend section

Coverage: read in full the production-logic hunks — GymieApiClient, TrainingSetupApi, ShorthandTokens, BodyCaloriesModels, QueueingWorkoutApi, CalorieHistoryState, WeighInHistoryState, EditProfileState (whole `save()` read in the worktree), CardioStepDraft, NotificationState, GrantRevokedHandler, RollAnotherState, DoseLine, WorkoutPrescription. Did not read the Compose layout/styling hunks (~70 UI files) or the commonTest/desktopTest hunks; the section's ~20k diff lines exceeded this dispatch's budget, hence `Unread: gymie-frontend`.

### Shared-state check

- `EditProfileState.saving` / `loadedSlug` / `slug`: `save()` writers (EditProfileState.kt:309, :369-370) — unchanged; the new `saveExperience()` early return at :353 now skips the slug claim (see Important 1).
- `EditProfileState.experience`: written at load (setup launch, `if (experience == null)`) and by the row; read live in `saveExperience()` :387 after a suspension (Minor 1).
- `EditProfileState.setupVersion` / `loadedExperience`: written on setup read success and on patch success only — still correct.
- `CardioStepDraft.checked`: cleared by every length setter (length/time/distance/distanceUnit), set by Save — still correct.
- `Connectivity.owedWriteNotices` via `reportWriteFailure()` (GymieApiClient.kt HttpSend intercept): now also counts timeouts wrapped as cancellation; same behaviour as the pre-existing transport-failure path, which also queues — still correct.
- `QueueingWorkoutApi` cache `getWorkout` / `WORKOUT_VERSION` / replaced versions: the new sent-at-once finish writes all three, mirroring the replay's landing — still correct.
- `WeighInHistoryState.entries` / `CalorieHistoryState.entries`: optimistic removal, then refetch (in `finally` for weigh-ins) — still correct.

### Issues

#### Critical (Must Fix)
None.

#### Important (Should Fix)
1. **gymie-frontend:shared/src/commonMain/kotlin/com/gymie/shared/ui/profile/EditProfileState.kt:353** — Task 17.
   `if (!patchSucceeded || !saveExperience()) return false` returns before the slug claim (:355 onward). When a user changes both the link and the Experience and the training-setup PATCH is refused (or the setup read never answered), the profile PATCH lands, the toast says "Profile saved — experience not saved…", and the requested slug is silently never claimed: no `slugError`, no mention in the toast. The message tells the user only experience failed; their new link did not save either.
   Fix: let the slug claim run regardless of the experience outcome (claim first or independently), and return false only after both have run, so the partial-save toast is true.
   Reproducer: `.superpowers/sdd/reproducers/4-primary-10.sh`

#### Minor (Nice to Have)
1. **gymie-frontend:shared/src/commonMain/kotlin/com/gymie/shared/ui/profile/EditProfileState.kt:387** — Task 17.
   `save()` documents that "every value this save acts on is snapshotted here, before the first suspension" (`requestedSlug` at :323), but `saveExperience()` reads `experience` live after `patchProfile()` resumed, so an Experience tap during the round trip is what gets written. Snapshot `experience` next to `requestedSlug` and pass it in.
   Reproducer: `.superpowers/sdd/reproducers/4-primary-11.sh`

### Recommendations
- None beyond the fixes above.

### Assessment
**Ready for the human gate?** With fixes
**Reasoning:** The read logic hunks (online finish send, length bounds, optimistic history removal, undo hold) are sound; the Edit-profile save's new experience step drops a requested slug claim on a partial failure while reporting only the experience as unsaved.
