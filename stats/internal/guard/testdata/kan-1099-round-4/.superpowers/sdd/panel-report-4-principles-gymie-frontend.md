Critical: 0
Important: 0
Minor: 0
Unread: gymie-frontend

# Panel round 4 — principles — gymie-frontend section

Coverage: same as the primary pass — production-logic hunks read in full; Compose layout/styling hunks and test hunks not read (section ~20k diff lines), hence `Unread: gymie-frontend`.

### Derived-state check

- `CardioStepDraft.checked` — summarises "Save pressed since the last length edit"; an event fact, not computable from other state. Writers: the four length setters (clear), Save (set). Derives correctly; `lengthError` itself is derived (getter), not stored.
- `EditProfileState.loadedExperience` / `setupVersion` — the server's last answer, mirroring the existing `loadedSlug` pattern; written only from server answers. Not a summary of other state.
- `Notification.undoHoldMillis` / `lift` — per-notification inputs, not summaries.
- No count/total/"any failed" flag introduced in the read hunks.

### Summary
Touches the shared Compose frontend: text rules (dose line, en-dash ranges), body histories (optimistic delete, page size), cardio step bounds (SE11), Edit profile Experience (SH29), online finish send (SH13). Principles applied: Single Source of Truth, DRY, KISS. Project standards resolved (CLAUDE.md, CONTRIBUTING.md); the Kotlin backend layering standard does not apply to the frontend. DRY is served well: the step sheet's bounds and range copy are read off the shorthand's (`timeInBounds`, `distanceInBounds`, `LENGTH_*_REFUSAL`) rather than restated; the online finish reuses the delete path's sent-at-once shape. Searched `shared/src/commonMain` for an existing equivalent of `doseLine` and `sentOrDropped`-style helpers: `sendFinish` differs (422 passthrough) from `sentOrDropped` (null-status only), a defensible separate helper. No new suppressions or lint-config changes in the read hunks.

### Issues
#### Critical (Must Fix)
None.
#### Important (Should Fix)
None.
#### Minor (Nice to Have)
None.

### Assessment
**Principles-compliant?** Yes
**Reasoning:** The read logic derives what it can, reuses existing bounds and copy, and adds no suppressions; the one snapshot-convention slip is reported under the primary pass as a correctness Minor.
