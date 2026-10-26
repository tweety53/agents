- [x] 13. In-run fix 3 — visual-verify-2: AD1–AD7 control glyphs, widths, spacing and frame states
**Build:** green
**Files:** `components/WeekProgramWorkoutsField.tsx`, `components/TemplateBlocksField.tsx`, `components/Sidebar.tsx`, `components/Sidebar.test.tsx`, `lib/frameGeometry.ts`, `fidelity/admin-templates.spec.ts`, `fidelity/admin-templates.spec.ts-snapshots/ad1-review-sheet-darwin.png`, `fidelity/admin-templates.spec.ts-snapshots/ad2-reject-form-darwin.png`, `fidelity/admin-templates.spec.ts-snapshots/ad3-unpublish-sheet-darwin.png`, `fidelity/admin-templates.spec.ts-snapshots/ad4-warm-ups-darwin.png`, `fidelity/admin-templates.spec.ts-snapshots/ad5-edit-move-darwin.png`, `fidelity/admin-templates.spec.ts-snapshots/ad6-create-template-darwin.png`, `fidelity/admin-templates.spec.ts-snapshots/ad7-create-week-program-darwin.png`
**Allowed-collateral:** `app/(admin)/**`, `components/**`, `lib/**`
**Tests:** `sidebar highlights Templates on week program pages`, `week program workout options name the routine`
**Regression:** the sidebar test fails if `/week-program/*` stops highlighting Templates; the options test fails if a workout option drops its routine name; the capture fails on drift from the re-recorded baselines.
**Baseline:** before=412 after=414 <!-- measured: grep -rE "^\s*(it|test)\(" --include='*.test.ts' --include='*.test.tsx' app components lib middleware.test.ts | grep -c . @ spectre/kan-1017-templates-step-3-admin-templates-queue-create -->
**After:** Task 12
**Commit:** `fix(admin): match AD1–AD7 control glyphs, widths and frame states`

Cross-repo: admin paths in the admin worktree; `fidelity/` paths in the playwright worktree, a second commit with the same subject and `Task-Id: 13`. Source: `.superpowers/sdd/verify-report-visual-verify-2-rest-gymie.md` (backend worktree) and `visual-verification.md` "Step 10 — measured". Decisions taken (recommended): the sidebar link spacing is fixed in this change (a shared source); the capture seeds a similar public template for AD1 and hovers row 1 for AD4 so those frame states are compared; the full app suite stays member-only.

  - [x] **Step 1:** AD7: workout options read `<name> — <routine>` as the frame draws; select width 580px; enabled ↑/↓ #020817; ↑–↓ gap 35px; Remove's right edge per the frame; Save/Cancel weight per the frame. Sidebar: highlight Templates on `/week-program/**`; link pitch 41–42px per the frames.
  - [x] **Step 2:** AD6: set-row × 12×11 #676b74; block ↑/↓ #020817 at the frame's stroke; select caret 5×5. AD4: last row bottom hairline; Filter by name 200px, Any muscle 180px; Delete flush with the Actions header's right edge. AD1–AD3: sheet row pitch 38–39px; close × 12px; AD2 textarea line height 22px.
  - [x] **Step 3:** `fidelity/admin-templates.spec.ts`: seed a public template similar to the AD1 reviewed one so the `Similar public: …` line draws; hover row 1 before the AD4 capture.
  - [x] **Step 4: Verify.** Admin verify with the touched test files; the project's `capture` for `fidelity/admin-templates.spec.ts` (stack running, start nothing; export `BACKEND_URL=http://localhost:11890 FRONTEND_URL=http://localhost:6810`), PNGs deleted first, run twice with the second unchanged; measure each view against its raw frame; leftovers named with cause.
  - [x] **Step 5: Commit** in both repositories with the subject above.

