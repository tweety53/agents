**Severity:** Important
**Slot:** failure-modes (round 0)
**Location:** src/app/template/src/main/kotlin/com/gymie/template/infrastructure/persistence/TemplateRepositoryAdapter.kt:120

Read-then-write with an unconditional update: admin approve racing the owner's withdraw both pass the WAITING check; last writer wins, so a withdrawn template can end PUBLIC.
