# Final review panel — kan-13-myflow-planning-and-status-fixes

Diff: `.superpowers/sdd/final-review.diff` (9 files, 550 insertions / 8 deletions, vs merge-base `5d3245e`)

## Roster

Required: Primary (sonnet), Bugbot-fallback/general-purpose (sonnet — Cursor's native `bugbot`
subagent_type is unavailable in this harness), Principles-Merged (sonnet).
Conditional, selected: Security-fallback/general-purpose (sonnet — same fallback reason; triggered
by path/file handling in the new script), Adversarial (sonnet; triggered by >~300 changed lines,
558), Lens B — simplicity & state (sonnet; triggered by >~200 changed lines).
Conditional, not selected: Lens C — robustness & ops (borderline — error handling in the new script
is trivial exit-code handling, `.myflow/project.md` touch is two doc-list lines, not runtime
config; judged not to meet the bar).

## Pass 1 — findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles-Merged | Important | `skills/myflow-start/SKILL.md:63` | The "no name given" resolution step still runs standalone `openspec list --json` instead of citing the new shared **Change name resolution** section — the SSOT claim that section makes ("every command cites this section") isn't true yet for `/myflow-start` itself |
| F2 | Adversarial | Critical | `scripts/check-task-build-green.py` (TASK_HEADING_RE / BUILD_TAG_RE) | No fenced-code-block awareness: an illustrative `**Build:**`/`###` line inside a ```` ``` ```` block is parsed as real, causing both false negatives (real untagged task missed) and false positives (fake task invented, real task's body truncated). Confirmed reproducible |
| F3 | Bugbot-fallback | Important | `scripts/check-task-build-green.py:190` (`by_id`) | Duplicate task ids silently collapse — last one wins — so a red task self-referencing a duplicated id can resolve against an unrelated task's green tag instead of being flagged. Confirmed reproducible |
| F4 | Bugbot-fallback | Important | `scripts/check-task-build-green.py:112` (`TASK_HEADING_RE`) | A heading with an id but no trailing title text (e.g. `### 1.1` with nothing after, no trailing space) fails the regex's required `\s` and isn't recognized as a task at all — its tag line (if any) is silently orphaned. Confirmed reproducible |
| F5 | Bugbot-fallback + Adversarial | Important | `scripts/test-check-task-build-green.sh` | The wrapper's no-argument repo-wide scan path (aggregation across files, `archive/` exclusion, `[ -e ] || continue`) is never exercised by any fixture — only the single-file branch is tested |
| F6 | Lens B | Minor | `scripts/check-task-build-green.py:112,130` | The dotted-id pattern `\d+(?:\.\d+)*` is written twice (`TASK_HEADING_RE`, `PARTNER_ID_RE`) with nothing enforcing they stay in sync |
| F7 | Bugbot-fallback + Adversarial | Minor | `scripts/check-task-build-green.sh:60-62` (archive exclusion `case`) | Dead code: the one-level-deep glob (`*/tasks.md`) can never reach a two-levels-deep archived path in the first place, so the explicit exclusion never fires. Harmless but overstates what the code does |
| F8 | Bugbot-fallback | Minor | `scripts/check-task-build-green.sh` (no-arg loop) | Aggregate mode collapses a per-file exit 2 (invocation error) into the same `STATUS=1` as an ordinary content violation — not silent (stderr still names it) but not separately signaled at the exit-code level |
| F9 | Bugbot-fallback | Minor | `scripts/test-check-task-build-green.sh` (case 6) | Only asserts `RC -eq 0`, unlike case 1, which also asserts empty output — a broken implementation that printed spurious text while still exiting 0 would pass |
| F10 | Bugbot-fallback | Minor | `scripts/check-task-build-green.py` (partner loop) | A clause naming the same partner id twice (e.g. `Task 2.1, 2.1`) produces a duplicate violation line for the same pair — cosmetic only |

findings-total: 10
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed

## Clean slots (pass 1)

- Primary: Ready to merge — Yes (2 minor notes, non-blocking, not entered above: a latent level-4-heading body-boundary edge case not exercised by this plan, and the prose-only unreadable-state-file rule being unverifiable by any test)
- Security-fallback: Ready — Yes (2 minor observations, both explicitly "no action needed": symlink/no-size-cap on `open()`, symlinked change directories — negligible for a local CLI the operator runs against their own tree)

## Fix round 1 (model: sonnet)

One fix subagent dispatched with the union of all 10 findings (`.superpowers/sdd/checkpoint` base
`f840517`). Report: all 10 fixed, 6 new fixtures added to `test-check-task-build-green.sh` (cases
8-14), `CHECK_TASK_BUILD_GREEN_ROOT` env override added mirroring
`CHECK_PLAN_PROVENANCE_ROOT` for F5's fixture-tree testing.

**Scoped re-review performed by the controller directly** (not a dispatched subagent — every
finding had a concrete, independently-reproducible repro command already captured in the pass-1
table, so the controller re-ran each one against the fixed code rather than dispatching a fresh
reviewer to re-derive them):

- F1: `skills/myflow-start/SKILL.md:63` now cites **Change name resolution** — confirmed by reading
  the edited bullet.
- F2: fenced-example repro now correctly reports the real task's missing tag (`exit=1`) instead of
  being swallowed — confirmed.
- F3: duplicate-id repro now reports `task 1.1 is defined more than once` (`exit=1`) instead of
  silently resolving — confirmed.
- F4: titleless-heading repro (`### 1.1` with no tag) now correctly reports the missing tag
  (`exit=1`) — confirmed the heading is now parsed as a task at all.
- F5: `CHECK_TASK_BUILD_GREEN_ROOT` override confirmed present; new fixtures 13-14 pass.
- F6-F10: confirmed via the full `test-check-task-build-green.sh` run (28/28 assertions pass,
  including the new cases covering each).
- Full regression: all 10 of this repo's `scripts/test-*.sh` suites pass (0 failures); all 4 lint
  guards (`check-vocabulary.sh`, `check-references.sh`, `check-plan-provenance.sh`,
  `check-task-build-green.sh`) exit 0; `openspec validate kan-13-myflow-planning-and-status-fixes
  --strict` passes; re-dogfooded the guard against this change's own `tasks.md` (exit 0, clean).

No new breakage found in the fix diff. **Zero open findings, at any severity — panel clean.**
