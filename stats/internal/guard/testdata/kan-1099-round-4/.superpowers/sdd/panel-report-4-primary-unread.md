Critical: 0
Important: 1
Minor: 0
Unread: gymie-frontend

# Panel round 4 — primary (gymie-frontend layout/test files, second dispatch)

Coverage: read in full DoseLine.kt, AuthTitle.kt, DoseLineTest.kt, CardioStepDraftLengthTest.kt, WorkoutPrescriptionRangeTest.kt; scanned the whole gymie-frontend section's added lines for new suppressions, @Ignore, sleeps/delays and TODOs (none found). Not read: the remaining ~200 Compose layout and desktopTest files (low-effort dispatch, 224 files / 9.4k lines).

## Important

- gymie-frontend:shared/src/commonMain/kotlin/com/gymie/shared/ui/workout/DoseLine.kt:15 (task 3) — TX7 "wraps only after ' · '" is not met: the glued part replaces spaces with NBSP but leaves the en dash TX13 now prints in every range ("10–12"); UAX#14 gives a break opportunity after U+2013 (class BA), so a narrow row can wrap "10–|12". DoseLineTest only checks characters, not breaks. Fix: also insert U+2060 WORD JOINER after each en dash inside a glued part (and assert it in DoseLineTest). Reproducer: .superpowers/sdd/reproducers/4-primary-60.sh (JDK line BreakIterator finds a break at '3 sets × 10–|').

## Shared-state check
No pre-existing shared state read or written by the files reviewed (doseLine is a pure function).

## Assessment
Ready for the human gate? With fixes.
