Critical: 0
Important: 0
Minor: 1
Unread: none

# Panel round 4 — primary — gymie-playwright section

Coverage: read in full the text hunks of the gymie-playwright section (spec/fixture/frames/scripts/tests changes, the new kan-1099-*.spec.ts and .mockups files at hunk level); binary PNG/zip baselines not inspected pixel-by-pixel. Other worktree sections not read (separate dispatches). Ran `node --test scripts/specs-for-diff.test.mjs` (34 pass) and confirmed the SMOKE line numbers (tests/auth.spec.ts:42, :131) match the `@smoke` tests; confirmed every `v2-plan-2026-11-09b/*` frame id cited by the kan-1099 specs and .mockups files is keyed in fidelity/frames.ts FRAME_HEIGHT (comm: no gaps).

### Shared-state check

- `page.clock` (paused/running) — writers in the diff: fidelity/fixtures.ts `typeOver` (pauseAt, resumes in `finally`); fidelity/add-participant.spec.ts caretOff followed by `page.clock.resume()` after each screenshot; tests/body-weigh-in-history-mockup.spec.ts existing `page.clock.resume()` before Cancel. Verdict: each pause is paired with a resume on its own path; still correct.
- `FRAME_HEIGHT` map (fidelity/frames.ts) — additions only, 816 per the 1688-row frames; no existing key overwritten. Still correct.
- `SMOKE` list (scripts/specs-for-diff.test.mjs) — updated to the shifted lines; verified against tests/auth.spec.ts. Still correct.

### Issues

#### Critical (Must Fix)
None.

#### Important (Should Fix)
None.

#### Minor (Nice to Have)
- **gymie-playwright:tests/auth.spec.ts:32** — `emailLine` now masks the email field's whole tag box (border, fill, label row inside the tag) instead of the per-run text line. The two smoke captures (`register-from-invite`, `sign-in-invalid-credentials`) no longer pin the restyled field's chrome, so a regression in the field border/fill would pass. Fix: inset the tag box to the text line (e.g. shrink to the box's inner text row) so only the per-run address is masked. Task: none (spans the re-record commit). Reproducer: `.superpowers/sdd/reproducers/4-primary-20.sh`.

### Recommendations
- The `pace()` 120 ms sleep in kan-1099-body.spec.ts is rate-limit pacing on API seeding, not a flake tolerance; acceptable.

### Assessment
**Ready for the human gate?** Yes
**Reasoning:** The playwright section is baselines plus new fidelity specs keyed correctly; one minor over-broad mask.
