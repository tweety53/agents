## Context

- Five Backlog Flow findings planned as one change per item 1 of **3. Fix every finding that is
  not `big`** (`skills/flow-self-review/SKILL.md`). Brainstorm design:
  `.superpowers/sdd/2026-10-18-self-review-backlog-flow-4-design.md` in this worktree. Every
  question was auto-resolved to its recommended option (`skills/flow-contracts/operator-prompts-auto-resolution.md`).
- **KAN-691 — Task 1.** `git grep -n -i interceptor -- skills` finds nothing, and **Testing
  principles** (`skills/flow/engineering-principles.md`) names no error-path test against a
  hand-built payload. Implementers read that file as REQUIRED READING (`skills/flow/implement.md`)
  and the principles reviewer reads it on every diff (`skills/flow/review-panel.md`). Blast
  radius: one file.
- **KAN-698 — Task 2.** `git grep -n -i prerequisite -- skills` finds only build prerequisites;
  nothing asks a plan's prerequisite to be an ancestor of the worktree's own HEAD. **Plan
  provenance** (`skills/flow-contracts/plan-provenance.md`) owns the `measured:` tag and is loaded
  by section D of `skills/flow/brainstorm-planner.md`. Blast radius: one file.
- **KAN-652 — Task 3.** gymie-frontend `origin/develop` @ `5ce0d237`:
  `shared/src/desktopTest/kotlin/com/gymie/shared/ui/HomeCalendarMarksUiTest.kt:320` is
  `assertFalse(tag.contains("inverted"), …)` on the tag string the test itself passes in, and the
  file asserts no mark colour. The marks draw in `colors.accent` (own), `colors.accent2`
  (borrowed) and `colors.text` (session) (`HomeCalendarGrid.kt`, `CalendarDayMarksRow`), each a
  6dp mark on a 1dp ring. An exact-colour pixel check fails on the unchanged code — the ring
  anti-aliases, the strongest own-mark pixel reads `ff2799ba` against accent `ff0088b0`. The
  planned nearest-token read passes on the unchanged code, and fails with the borrowed mark
  drawn in accent (a swap) and with the own mark drawn in `neutral100` (an inversion to the
  disc's ink); the file's current test passes on that inversion. `./gradlew ktlintFormat` and
  `./gradlew verifyChange :shared:desktopTest --tests 'com.gymie.shared.ui.HomeCalendarMarksUiTest'`
  exit 0 with it applied; the helper sits at file level because one more class member trips
  detekt's `TooManyFunctions` (20). Blast radius: one file, in gymie-frontend.
  <!-- measured: ./gradlew :shared:desktopTest --tests 'com.gymie.shared.ui.HomeCalendarMarksUiTest.todayAndSelectedKeepEveryMarkBeneathTheDiscInItsOwnColour' with each mutation of HomeCalendarGrid.kt in a scratch worktree @ gymie-frontend 5ce0d237 → swap exit 1, inversion exit 1, inversion under the old test exit 0 -->
- **KAN-664 — already held, no task.** The verifier authors a missing `full-app-suite.spec.ts`
  (`skills/flow/visual-verify-verifier.md` step 8, since `f3df0a80`, 2026-09-18 — before
  kan-580's run). A procedure blocker is no longer a report the parent can accept:
  `check-verify-report.sh` (since `1f0b7657`, 2026-10-02) admits `blocked` only with a non-zero
  `exit <n>`, and an incomplete report's remainder is dispatched at once (**Every report is
  checked for completeness**, `skills/flow/visual-verify.md`); since `5d11de99` (2026-10-18) a
  sweep left undone is step 10 undone. A report carrying kan-580's two blockers —
  `8 blocked — no full-app-suite.spec.ts exists in gymie-playwright` and
  `10 blocked — matrix sweep not run` — gets `VERIFY-REPORT-INCOMPLETE` naming both.
  <!-- measured: scripts/check-verify-report.sh <that report> 0 @ 54033573 → exit 1 -->
- **KAN-701 — already fixed, no task.** The crop's top band is the project's declared geometry,
  gymie's `mockup frame` row `scale=2 status=26 border=1`, not a script constant. The nav-D frames
  kan-573 composed had no status band: gymie `aae79725` "re-export nav-D1..D5 with the status-bar
  band" landed on `develop` at 2026-09-19 23:57, an hour after kan-573's merge `4eb39cbb` (22:59)
  and outside it. On today's frames `compose-mockup-frames.sh` composes all five
  `nav-drawer-rooms.spec.ts.mockups` pairs from `gymie-playwright` @ `d416244`, exit 0, the bar
  and its title at row 0 in both panels; on the frames before `aae79725` it refuses every pair
  with a 52px height mismatch. The `back → D1` and `≡ → D2` chips are drawn in the mockups' bars,
  no bleed. `skills/flow/scripts/compose-mockup-frames.sh` is a symlink to
  `scripts/compose-mockup-frames.sh`.
  <!-- measured: compose-mockup-frames.sh fidelity/nav-drawer-rooms.spec.ts.mockups docs/design/mockups <out> 'scale=2 status=26 border=1' @ agents 54033573, gymie 8d2bb89d → exit 0, 5 composites; with the frames at gymie aae79725^ → exit 1, 5 size mismatches -->
- Tasks 1 and 2's diffs were applied in order to a scratch clone at `54033573` (this worktree's
  base): `check-references.sh`, `check-installed-citations.sh` and `check-markdown-integrity.py`
  exit 0, and `check-verbatim-moves.sh 54033573` fails on exactly the seven sentences this
  change's `verbatim-moves.txt` acknowledges.

## Decisions

### Error-path reachability is a testing principle

**ID:** error-path-real-trigger
**Status:** active
**Chosen:** a paragraph under **Testing principles** (`skills/flow/engineering-principles.md`): an error-path fix's test traces the real response through every interceptor to the handler and drives the real trigger, never a hand-built payload the backend never sends.
**Considered:** a planner rule in section D of `skills/flow/brainstorm-planner.md` — panel and fix-run fixers never pass through the planner, and kan-527's fix was a panel fix; a guard — whether a payload is one the backend sends is a judgment no static check can make.

### A prerequisite is an ancestor of the worktree's HEAD

**ID:** prerequisite-worktree-ancestor
**Status:** active
**Chosen:** a section in **Plan provenance** (`skills/flow-contracts/plan-provenance.md`): the plan measures a prerequisite with `git -C <worktree> merge-base --is-ancestor <sha> HEAD`, the dependent task's first step re-runs it, and an exit 1 is a stale base reported, never built past — rebasing is the operator's call.
**Considered:** a `**Requires:**` task field checked by a new Go guard — a field and a guard for a one-line git check the plan already writes as the task's first step; asserting against `origin/<base>` after a fetch — the defect was exactly a remote that carried what the worktree did not.

### The calendar marks test reads its ink by nearest token

**ID:** marks-colour-nearest-token
**Status:** active
**Chosen:** `assertMarkBeneathDisc` takes the mark's expected colour; the captured mark's pixel furthest from the page background must lie nearest that colour among accent, accent-2, text and `neutral100`, the disc's ink — replacing the tag-literal check.
**Considered:** exact pixel equality — measured failing on the unchanged code, no anti-aliased 1dp ring pixel reads a token exactly; a per-channel tolerance — a threshold to tune where ranking against the four candidate inks needs none; making the fixture's marks FINISHED so they draw solid — changes the scenario the test pins (planned marks under the disc).

### KAN-652 lands in gymie-frontend off develop

**ID:** kan-652-gymie-frontend-worktree
**Status:** active
**Chosen:** per **Pipeline defects found mid-run** (`skills/flow-contracts/pipeline.md`) for a repository the pass holds no worktree in: a worktree `/Users/tweety53/Projects/gymie-frontend-worktrees/self-review-backlog-flow-4`, branch `self-review-backlog-flow-4` off `origin/develop`, created by Task 3's first step; landed by gymie-frontend's `## default landing route`, `merge and push` onto `develop`, by the session at step 5 — never by the fixer.
**Considered:** riding the agents branch — the test lives in gymie-frontend and only lands there; leaving it for a gymie change — the pass fixes every non-`big` finding.

### KAN-664 is already held

**ID:** procedure-blocker-already-held
**Status:** active
**Chosen:** no task — step 8 already authors a missing suite, and `check-verify-report.sh` refuses a procedure blocker that cites no failing exit, so its remainder is dispatched in the same run.
**Considered:** a sentence calling a procedure blocker a work order — restates what the guard already enforces mechanically.

### KAN-701 is already fixed

**ID:** compose-crop-already-fixed
**Status:** active
**Chosen:** no task — the clip came from nav-D frames exported without the declared status band, re-exported in gymie `aae79725`; the script composes today's frames with titles intact.
**Considered:** changing the crop — it is the project's declared geometry, right for every frame now in the tree; detecting the status band from the image — a heuristic for frames that no longer exist, and a frame whose band is missing is already refused by the size check.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — two prose paragraphs and one test-only change in another repository, its colour read mutation-proven in a scratch worktree; no production code moves.
**Considered:** `flow-medium` — the gymie-frontend task is a pixel assertion, but it was run against both mutations and the unchanged code, which a first look can re-run.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** two groups in series — `self-review-backlog-flow-4-fix-g1`: Tasks 1 and 2 (agents prose, one `git apply` each); `-fix-g2`: Task 3 (gymie-frontend worktree, a Gradle build and test run).
**Considered:** one group of three — one fixer carries a Gradle run's output beside two prose edits for no gain; three groups — a dispatch per one-paragraph edit.

## Open questions
