- [x] 7. The handler-mapping coverage test and the rights matrix over the real stack
**Build:** green
**Files:** `src/app/src/test/kotlin/com/gymie/OwnerArgumentCoverageTest.kt`, `src/app/src/test/kotlin/com/gymie/FriendRightsMatrixIntegrationTest.kt`
**Tests:** `every friend-data handler takes Owner or is listed caller-only`, `a stranger and a friend one level too low get 403 on every Owner route`, `friend A's Edit does not reach friend B's workout on any by-id route`, `friend A's Edit does not reach friend B's cardio by id`
**Regression:** the coverage test fails with `Owner` removed from any one `WorkoutController` by-id handler and that handler not added to the caller-only list; the matrix fails with `OwnerWebConfig` not registering the resolver; the cross-friend tests fail with `findOwned` given the caller's id.
**Baseline:** before=347 after=351 <!-- measured: grep -rw '@Test' src/app/src/test | wc -l @ branch spectre/kan-1096-friends-view-rights -->
**After:** Task 3, 4, 5, 6
**Commit:** `test(app): every friend-data handler is rights-checked`

**Decision:** owner-argument-resolver
**Decision:** by-id-names-the-owner

  - [x] **Step 1:** `OwnerArgumentCoverageTest` (`@SpringBootTest` as `FriendGraphIntegrationTest` boots): walk `RequestMappingHandlerMapping.handlerMethods`; for every pattern under `/api/workouts`, `/api/cardio-workouts`, `/api/body`, `/api/week`, `/api/workouts/training-setup`, the handler takes an `Owner` parameter or is in a `CALLER_ONLY` map of handler → reason (lists, today, delete-all, templates, week programs, warm-up, preferences, progression offers, calorie calculator, exercise catalog, profile/avatar/slug/units, friends, access). Run it with `Owner` removed from one handler to see it fail, then restore.
  - [x] **Step 2:** `FriendRightsMatrixIntegrationTest`: register a viewer, friend A (Training EDIT, Body READ, setup READ), friend B (Training EDIT), stranger C; for each `Owner` route, C and A one level too low get 403 `grant_revoked`; with `ownerId=A` and B's workout id (and cardio id), every by-id route answers 404.
  - [x] **Step 3: Verify.** Backend verify with `:app:test --tests 'com.gymie.OwnerArgumentCoverageTest' --tests 'com.gymie.FriendRightsMatrixIntegrationTest'`.
  - [x] **Step 4: Commit** with the subject above.
