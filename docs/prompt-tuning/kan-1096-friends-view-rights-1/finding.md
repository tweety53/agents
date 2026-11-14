# F27 — test-sanity, round 0, Important

Location: src/app/src/test/kotlin/com/gymie/FriendRightsMatrixIntegrationTest.kt:179 (gymie)

> The on-any-by-id-route test claims more than it checks: child path variables are random UUIDs so those routes 404 whether or not the ownership check works, with no positive control under ownerId=b.

Same class, same round (F26, test-sanity, Important, src/app/workout/src/test/kotlin/com/gymie/workout/core/services/week/WeekServiceImplTest.kt:92):

> an owner whose grant was revoked produces no group can't fail: no input supplies ownerAId; it appears only in the assertion that she is absent.

Fixed by panel-fix-1, a9c56081 (gymie): real child ids, plus a control that the same routes do not 404 under the row's own owner.
