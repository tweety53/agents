# Final review panel — kan-23-myflow-self-review

**Diff:** `.superpowers/sdd/final-review.diff` (`git diff <merge-base>`, staged+unstaged)
**Mode:** pass 1 — full roster
**Roster:** Primary, Bugbot (generalPurpose fallback — `subagent_type: bugbot` unavailable in this
harness), Principles/Merged, Security (generalPurpose fallback — `subagent_type: security-review`
unavailable in this harness), Adversarial, Principles/Lens B, Principles/Lens C
**Model:** every slot dispatched on `models.reviewPanel` = Sonnet (recorded in state file), since
none of the seven ran under a real `subagent_type` agent definition in this harness

## Optional slot selection

Evaluated against `final-review.diff` (1571 lines) before dispatching:

- **Security** — included: the diff touches path/file handling (new script resolving
  ledger/panel/tasks.md paths from user-influenced input) and a config file (`.myflow/project.md`).
- **Adversarial** — included: diff exceeds ~300 changed lines (1571).
- **Lens B (simplicity & state)** — included: diff exceeds ~200 changed lines.
- **Lens C (robustness & ops)** — included: new script does error handling, config/env, external
  (git) integration.

All four conditional triggers were met mechanically (not borderline), so no operator ask was needed.

## History (narrative — every finding below is closed; see the live marker block at the end)

The passages below record what each pass and each fix round found and did, for anyone auditing how
this change's one recurring defect class converged. None of the `F<n>`/`N<n>` labels in this section
are live status markers — the single authoritative marker line sits at the very end of this file,
declaring zero remaining.

### Pass 1 — 8 findings

- **F1** (Security + Adversarial, independently) — Critical — `scripts/gather-self-review-context.sh`
  (`find_dated`, `cat` of `LEDGER_FILE`/`PANEL_FILE`/`TASKS_FILE`): symlink-following arbitrary file
  read — a symlink planted at a tracked, PR-editable path under `docs/superpowers/ledgers/`,
  `docs/superpowers/reviews/`, or `<archived-path>/tasks.md` was followed and its target's content
  read into the bundle, which is then committed and pushed. `preserve-session-records.sh`'s own
  root-boundary protection was not carried over.
- **F2** (Merged + Bugbot + Primary + Adversarial) — Minor — `design.md` said the skip line prints
  "on stderr" (spec.md and the implementation both correctly said stdout); the script's own NOTE
  comment misattributed the stale claim to "the delta spec" instead of the design doc.
- **F3** (Lens C + Adversarial) — Important — a nonexistent/invalid `<archived-change-path>`
  degraded identically to "the change legitimately has no records".
- **F4** (Lens C + Primary) — Important — the script's header and the delta spec both stated it
  "ALWAYS exits 0" unconditionally, but the script intentionally exits 2 on a bad invocation.
- **F5** (Lens B) — Minor — the `<state-dir>` parameter was required/validated but never read,
  undocumented as a deliberate decision.
- **F8** (Primary) — Minor — the `IMPL_SHA` filter excluded any commit whose subject merely
  contained "archive" anywhere, not just the dedicated archive-commit shape.
- **F9** (Adversarial) — Important — `IMPL_SHA`/`PLAN_SHA` git-log resolution was never actually
  asserted by the test suite, only the archive commit's content was checked.
- **F11** (Adversarial) — Minor — multi-fix-round history wasn't covered by the `IMPL_SHA`
  discovery logic or its tests; undocumented assumption.

**Fix round 1** (single combined subagent, model Sonnet, diff base
`f92fe89a6429693eb0e78ba9825c44ae3f5a7c90`, no commits): ported `resolve_file`/`within_root`/
`check_boundary` from `preserve-session-records.sh` (F1); corrected `design.md` to stdout and
reworded the script's stale NOTE (F2); added a distinct `note:` line for a nonexistent archived path
(F3); corrected the exit-2/exit-0 contract in both the header and the delta spec (F4); documented the
`<state-dir>` parameter as a named decision (F5); tightened the `IMPL_SHA` exclusion to the actual
archive-commit shape (F8); added `PLAN-COMMIT-BODY`/`IMPLEMENTATION-COMMIT-BODY` assertions (F9);
acknowledged the multi-fix-round assumption (F11). Re-verified: all three lint guards and the test
suite (46 assertions) passed clean.

### Round-3 re-run — 2 new findings (escalated: fix round 1 altered a delta spec)

All 7 slots re-dispatched; F1–F11 all independently re-confirmed fixed with fresh manual
reproductions.

- **F12** (Security-fallback) — Critical — the archived-change directory itself could be a symlink
  (git tracks symlinks as committed blobs); `ARCHIVED_REAL` resolved through it and became the trust
  boundary `tasks.md` was checked against, so `tasks.md` content was read with no refusal. Reproduced
  with both relative and absolute paths.
- **N1** (Primary) — Minor — a second design doc still said "always exits 0" unconditionally; the
  round-1 fix updated the script header and the delta spec but missed it.

**Fix round 2**: added `ARCHIVED_PATH_SYMLINK`/`ARCHIVED_PATH_MISSING`/`ARCHIVED_PATH_INVALID`
gating before `REPO_ROOT`/`ARCHIVED_REAL` derivation (F12); corrected the second design doc (N1). New
test case and delta-spec scenario for F12. Re-verified clean.

### Round-4 re-run — 4 new findings (escalated again)

F12/N1 both independently reconfirmed fixed with fresh reproductions.

- **F13** (Adversarial) — Critical — `[ -L "$ARCHIVED_PATH" ]` with a trailing slash resolves
  through the symlink before testing (POSIX behavior), and the documented caller invokes with exactly
  that trailing-slash shape — fully reopened F12.
- **F14** (Lens C) — Important — the "narrow same-run window" justification for checking only the
  leaf was unsound: `openspec/changes/archive/` (or its own ancestors) is itself a tracked path a
  merged PR could pre-plant as a symlink, and `git mv` in finish step 2 follows it transparently.
- **M1** (Primary) — Minor — a stale "Always exits 0" claim survived in `tasks.md`.
- **M2** (Primary) — Minor — the commit+push snippet's `&&`/`||` grouping let `push` fire even when
  nothing was committed.

**Fix round 3**: fixed F13/F14/M1/M2 directly, with new tests for F13/F14. Re-verified clean.

### Round-5 re-run — 8 new findings (escalated again)

F13/F14/M1/M2 all independently reconfirmed fixed.

- **F20** (Security-fallback) — Critical — a `..` component in `$ARCHIVED_PATH` desynced the
  lexical `dirname`-based 3-hop ancestor walk from the real (kernel/git-resolved) ancestors —
  twice-reproduced, attacker-controlled content read straight through.
- **F17** (Merged) — Important — a script comment cited a worktree-scoped file
  (`.superpowers/sdd/final-review-panel.md`) removed at finish run 2, dangling for the script's
  entire permanent life.
- **F15** (Lens B) — Minor — path-validation logic (4 booleans + ancestor walk) should consolidate
  into one function, matching the `resolve_file`/`within_root`/`check_boundary` pattern already used
  elsewhere in the script.
- **F16** (Bugbot-fallback) — Minor — a verify-block comment's assertion count was stale.
- **F18** (Merged) — Minor — near-duplicate note-printing across branches (defensible WET call,
  flagged for awareness).
- **F19** (Merged) — Minor — a helper's placement broke the file's helpers-first convention.
- **F21** (Lens C) — Minor — the ancestor-symlink test only covered hop 1 of 3.
- **N2** (Adversarial) — Minor — the trailing-slash-leaf-symlink behavior had normative prose but no
  scenario block, unlike its siblings.

**Fix round 4**: consolidated all path-validation logic into one `validate_archived_path()` function,
combining the fixes for F15 and F20; fixed F17/F16/F18/F19/F21/N2 directly. Re-verified clean.

### Round-6 re-run — 1 new finding (escalated again)

Six reviewers found nothing further.

- **F22** (Security-fallback) — Important — the bounded-3-hop lexical approach was fragile by
  construction to any path shape deviating from the documented exact invocation (e.g. one extra
  nesting level defeats the hop count); not exploitable via the actual documented caller today, but
  the root cause behind four rounds of findings was still open. Recommended: resolve
  `$ARCHIVED_PATH` semantically once, derive the trusted repo root independently, and require an
  exact match.

**Fix round 5**: replaced the bounded-hop heuristic entirely with a general
lexically-normalize-vs-semantically-resolve-then-compare mechanism in `validate_archived_path()`,
subsuming F12/F13/F14/F20/F22 in one check. Re-verified clean.

### Round-7 re-run — 4 new findings (escalated again)

Six reviewers found nothing beyond documentation drift; two found real (non-Critical) issues.

- **F23** (Lens C) — Important — bare `git rev-parse --show-toplevel` returns the *worktree* root
  when invoked from inside a worktree, not the main repo root — this repo's own `state-file.md`
  documents this exact hazard and prescribes `--git-common-dir`.
- **F26** (Bugbot-fallback) — Important — an unquoted `for part in $abs_input` loop let glob
  metacharacters in `$ARCHIVED_PATH` trigger pathname expansion against cwd, falsifying the "pure
  string manipulation" invariant. Failed safe in every case tried, but a real defect.
- **F24** (Lens B) — Minor — the trailing-slash strip mutated the global for cosmetic reasons inside
  a security-named function.
- **F25** (Primary) — Minor — a leaf symlink to a file (not a directory) got a misleading reason
  label.

**Fix round 6**: derived `TRUSTED_REPO_ROOT` via `--git-common-dir` (F23); replaced the glob-prone
loop with `read -ra` (F26); moved the cosmetic strip out of the security function (F24); distinguished
the file-symlink case (F25). Re-verified clean.

### Round-8 full re-run — 4 new findings (escalated: "3+ fix rounds" trigger)

Six reviewers found nothing beyond documentation drift; three independently found the same
delta-spec inconsistency.

- **F27** (Merged, Primary, Adversarial — independent) — Important — the delta spec's illustrative
  example for deriving the trusted repo root still cited `git rev-parse --show-toplevel`, the exact
  command F23 proved wrong.
- **F28** (Merged, Primary, Bugbot-fallback, Adversarial — independent) — Minor — the assertion-count
  comment was stale again (claimed 80, actual 90) — same recurring class as F16.
- **F29** (Adversarial) — Minor — a draft snippet in `tasks.md` still showed the pre-M2 buggy
  grouping M2 fixed in the real `SKILL.md`.
- **F30** (Adversarial) — Minor — several verify-block tags still said "run once the edit is made"
  for edits long since made and verified.

**Fix round 7** (documentation-only, no code change): corrected the delta spec's example to
`--git-common-dir` (F27); corrected the assertion count (F28); corrected the draft snippet (F29);
converted the stale tags to `verified:` with real evidence (F30, 4 of 5 instances found in the
sweep).

### Round-9 targeted re-run — 1 leftover instance

Targeted 3-slot re-run (Primary, Adversarial, Principles/Merged — the three who raised F27–F30; not
a full 7-slot round, since this fix touched no code file and the security surface had just been
exhaustively re-verified in round 8 with zero code touched that round). Two slots found nothing;
Merged found one leftover instance of the F30 pattern (one stale tag the sweep missed) — fixed
directly, no further round needed (same trivial class, zero risk, already covered by the same three
reviewers' sign-off on every other instance of this exact pattern).

## Panel converged

9 rounds total. 1 Critical class (symlink/path-traversal trust-boundary bypass) required 5 rounds to
converge from bounded per-case patches to a single general mechanism (`validate_archived_path()`);
the general mechanism then needed 2 more rounds to fix its own trust-anchor derivation (F23) and a
quoting bug (F26), plus one round of pure documentation drift (F27–F30) and one trivial leftover.
All other findings were Important/Minor code-quality, test-coverage, and documentation-consistency
issues, each fixed same-round. Final state: zero open findings at any severity, full lint + test
suite green, no commits made throughout.

findings-total: 0
