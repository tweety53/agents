Critical: 0
Important: 5
Minor: 0
Unread: gymie, gymie-playwright (mutation scoped to frontend desktop-unit logic; backend CalorieLedger and playwright left unmutated)

Throwaway copy: gymie-frontend ...-mutation-4, one Gradle run at a time, tree restored after each mutant. Baseline scoped run ~30 s. Task ids: no Task-Id mapping was read; all findings "none".

## Surviving mutants (each reproducer: `# mutation-reproducer`, exit 0 = survived)
1. Important - gymie-frontend:shared/src/commonMain/kotlin/com/gymie/shared/ui/workout/DoseLine.kt:12 - `<=` -> `<` (and 28 -> 29) survives: no test has a part of exactly 28 or 29 chars, so the TX7 glue boundary is unpinned. Reproducer: reproducers/4-mutation-1.sh. Task: none.
2. Important - gymie-frontend:shared/src/commonMain/kotlin/com/gymie/shared/sync/QueueingWorkoutApi.kt:164 - dropping `queue.entries.value.none { it.workoutId == id }` survives QueueingWorkoutApiTest + EditorSaveChecklist/ListModeFrames/ActiveWorkoutOffline UI tests: a finish may then overtake queued edits of the same workout and no test notices. Reproducer: 4-mutation-2.sh. Task: none.
3. Important - gymie-frontend:shared/src/commonMain/kotlin/com/gymie/shared/sync/QueueingWorkoutApi.kt:412 - `if (isActive) null else throw e` -> `null` survives: a cancelled caller's cancellation is swallowed and the finish queued. Only the timeout case is tested. Reproducer: 4-mutation-3.sh. Task: none.
4. Important - gymie-frontend:shared/src/commonMain/kotlin/com/gymie/shared/cardio/shorthand/ShorthandTokens.kt:15 - MAX_DECIMALS 3 -> 4 survives the shorthand and draft-length tests; the decimals bound is unpinned. Reproducer: 4-mutation-4.sh. Task: none.
5. Important - gymie-frontend:shared/src/commonMain/kotlin/com/gymie/shared/ui/profile/EditProfileState.kt:352 - dropping `ProfileHolder.invalidate()` on a landed PATCH survives EditProfileStateTest and the profile UI tests; the menu/Settings stale-name guarantee is untested. Reproducer: 4-mutation-5.sh. Task: none.

## Survivors without a reproducer (not blocking-grade, listed)
- QueueingWorkoutApi.kt:163 drop `!connectivity.offline.value`: survives; near-equivalent (an offline PUT fails and is queued).
- QueueingWorkoutApi.kt:174 drop storing the at-once response ETag: survives; the successful at-once path has no test at all (only 5xx and timeout are covered).
- EditProfileState.kt: `wanted == loadedExperience` -> false (always PATCH experience), and not updating `setupVersion`/`loadedExperience` after a landed experience PATCH: survive; a second save's stale version is untested.
- ShorthandTokens.kt:78 `1..MAX_SECONDS` -> `1 until MAX_SECONDS`: survives (86400 s boundary).

## Killed (tests are sound)
QueueingWorkoutApi 422 passthrough; DoseLine separator; EditProfile experience failure surfacing; WeighIn optimistic delete and refetch-in-finally; CalorieHistory remove, in-force label, held-days span, singular target; time lower bound; distance upper bound; draft length-edit clearing `checked`; Save-gated range line; blank-time hasLength; Undo hold.
