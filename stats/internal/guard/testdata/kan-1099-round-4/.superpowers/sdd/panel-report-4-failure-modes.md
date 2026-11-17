Critical: 0
Important: 0
Minor: 2
Unread: gymie, gymie-playwright (frontend only: QueueingWorkoutApi, EditProfileState, WeighIn/CalorieHistoryState read at hunks)

### Failure-recovery check
- `EditProfileState.loadError`: set in loadProfile's failure branch; cleared at the success branch of loadProfile (`loadError = null`, same function). That success re-reads the profile fields (fill) - confirmed. It does not touch `setupVersion`/`loadedExperience`/`experience`, which a failed setup read leaves unknown: re-confirmed only by a later `load()` setup success (`setupVersion = doc.version`, load()). Partial recovery is bounded: a failed setup read leaves `setupVersion` null and saveExperience refuses with a visible message, so no wrong write results.
- `ProfileHolder` invalidation, `notifications.failure` toasts: transient, no clearing state.
- WeighIn `delete` optimistic removal: failure path refetch restores the row (`finally { refetch() }`); CalorieHistory `remove` is followed by the delete's refetch - confirmed by killed mutants.

### Summary
Lenses bearing: error return (finish-at-once, experience PATCH), timeout (setup read, at-once finish), concurrent re-entry (finish vs queue). Verdict: no Critical/Important; two Minor.

### Issues
#### Minor
1. gymie-frontend:shared/src/commonMain/kotlin/com/gymie/shared/ui/profile/EditProfileState.kt:~205 (load, timeout lens) - `coroutineScope { launch { setup read }; loadProfile() }` does not return until the setup read ends, so a hanging read keeps `load()` suspended (the form is unblocked, `hasLoaded` set in loadProfile); a back-online refetch or `scope.launch { load() }` stays pending behind it. The ktor client timeout bounds it in practice. Fix: launch the read in a scope outside load's. Task: none. Reproducer: none - needs a Gradle test file; traced in code.
2. gymie-frontend:shared/src/commonMain/kotlin/com/gymie/shared/ui/profile/EditProfileState.kt:~395 (saveExperience, error return) - a failed experience PATCH (e.g. a version conflict) leaves `setupVersion` stale; a retried Save re-PATCHes the profile and then fails the same way until the form is reopened. The failure response is not used to refresh the version. Task: none. Reproducer: none - needs a Gradle test file; traced in code.
### Assessment
**Ready for the human gate?** With fixes (both Minor; optional)
