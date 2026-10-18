## Context

- kan-983's self-review pass returned five findings, none `big`, planned as one change per item 1
  of **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`). Brainstorm
  design: `.superpowers/sdd/2026-10-19-self-review-kan-983-render-screens-from-cached-data-first-refresh-design.md`
  in the agents worktree. Every question was auto-resolved to its recommended option
  (`skills/flow-contracts/operator-prompts-auto-resolution.md`). gymie targets PWA and web only.
- **Finding 1 — Tasks 1, 2 and 4.** `scripts/run-verify-list.sh` (landed by `b8a5c1d0`) runs
  every non-`#` line of the `## test` fence through `bash -c`. gymie's `.flow/project.md` line 458
  and gymie-frontend's line 35 hold the verify-bundle template
  `./gradlew verifyChange :<module>:test --tests '<selectors>'`; bash reads `<module>` as an input
  redirect, so gymie's list always exits 1, and gymie-frontend's runs `verifyChange` and then fails
  on the unmatched `'<selectors>'`. No guard reads that template line: the only other
  `verify bundle` hit in agents is a Go test fixture. Blast radius: five files.
- **Finding 2 — Task 5.** `skills/flow/visual-verify.md` dispatches the verifier with no step
  that creates the `regression checkout`'s worktree; **2. Isolate the workspace**
  (`skills/flow/implement.md`) says a later stage needing an `## apps` worktree runs the recipe of
  `skills/flow/cross-repo-worktrees.md`, and nothing in visual verification does. kan-983's first
  verifier blocked on the missing gymie-playwright worktree. Blast radius: one file.
- **Finding 3 — Task 6.** `check-visual-verify-dispatched.sh` counts a verifier row for a worktree
  only when the row's `diffBase` is a sha that worktree resolves
  (`stats/internal/guard/visualverifydispatched.go`). A sibling worktree matched through the
  canonical worktree's `gymie-frontend/**` glob was covered by the canonical verifier, whose
  `diffBase` names the canonical HEAD only, so the parent wrote a phantom
  `visual-verify-2-gymie-frontend` row carrying no tokens. `flow record dispatch begin -diff-base`
  stores any string (`stats/cmd/flow/record.go`), and no other reader of `diffBase` exists outside
  the guard and its tests. Blast radius: four files.
- **Finding 4 — Task 8.** gymie-playwright's `tests-full/group-session-journey.spec.ts` timed out
  in `addToBuilder` (`fixtures.ts`) waiting for `add_participant_workout_field`. `addToBuilder`
  taps `add_participant_row_<id>` as soon as `clickTag`'s settle check passes, while `filterTo`
  waits out the skeleton-row reflow first (KAN-761). Not reproduced at plan time — no stack was
  running and no failing trace survives — so the task opens on the reproduction and the trace.
  Blast radius: one file.
- **Finding 5 — Task 7.** **The verifier dispatch**'s re-dispatch rule (`skills/flow/visual-verify.md`)
  sends a prompt identical to the first, so the verifier's step 7
  (`skills/flow/visual-verify-verifier.md`) re-runs the whole mapped spec list even when HEAD has not
  moved; kan-983's `visual-verify-2` spent its run re-running every spec to clear one. Blast
  radius: two files.
- **Found while planning — Task 3.** `origin/main` @ `9df14302` fails
  `TestNoPipeIntoEarlyExitGrep` on `scripts/test-check-task-records.sh` lines 392 and 411, two
  `printf '%s' "$OUT" | grep -q` pipes `646ac917` added; the whole Go guard suite is red until they
  read a here-string. Blast radius: one file.

## Decisions

### Refuse a placeholder line, never run or skip it

**ID:** verify-list-placeholder-refusal
**Status:** active
**Chosen:** `run-verify-list.sh` scans the whole list before running anything and exits 2, its usage code, on any line holding a `<word>` token, naming the line on stderr — a template in the fence becomes a loud configuration error instead of a red test run.
**Considered:** skipping the line with a report row — hides a configured command nobody runs; stripping a trailing comment before the check — no fence line carries a placeholder only in its comment today, so it would be a speculative case.

### The verify-bundle template lives in prose

**ID:** verify-template-in-prose
**Status:** active
**Chosen:** gymie's and gymie-frontend's `## test` fences keep their full-suite line alone; the verify-bundle template moves to one paragraph below the fence, and project-configuration.md's `## test` row states that fence lines run literally.
**Considered:** a separate `## verify bundle` key — configurability no reader asked for; keeping the line commented out with `#` — the runner would skip it, but a `#` line reads as disabled, not as the template a task's verify step fills in.

### One task per repository for finding 1

**ID:** finding-1-one-task-per-repo
**Status:** active
**Chosen:** finding 1 is carried by Task 4, the agents fix; Tasks 1 and 2 are its gymie and gymie-frontend halves, one commit each in their own worktree, since one commit cannot span three repositories. Tasks 1 and 2 land before Task 4.
**Considered:** one task committing in three repositories — `check-task-commit-fields.sh` finds a task's commit in one worktree, so two of the three commits would go unchecked.

### The outside worktrees exist before the first task

**ID:** outside-worktrees-at-plan-time
**Status:** active
**Chosen:** the planner created gymie's, gymie-frontend's and gymie-playwright's worktrees at plan time, each `git worktree add -b self-review-kan-983-render-screens-from-cached-data-first-refresh` off its default branch, then ran that repository's `## worktree setup`; tasks name the paths.
**Considered:** a task step creating the worktree — **D. Basic Workflow #3 — Writing plans** (`skills/flow/brainstorm-planner.md`) forbids it; `kickoff-worktree.sh --from` — it needs a `STARTED` state record this self-review pass does not have.

### The regression checkout's worktree is created before the pre-flight

**ID:** regression-worktree-before-preflight
**Status:** active
**Chosen:** step 3 of **Steps 3–13** opens with the parent running the `## apps` entry's recipe for a missing `regression checkout` worktree, before the pre-flight checks.
**Considered:** a new guard refusing the dispatch — names the gap but still costs a turn; leaving it to **2. Isolate the workspace** — the stage that already missed it.

### One verifier names every HEAD it covers

**ID:** comma-joined-diff-base
**Status:** active
**Chosen:** a verifier covering a sibling worktree through the canonical `ui paths` records each covered HEAD in its `-diff-base`, comma-joined, and the guard splits on `,` and judges each sha alone — a sha resolves in its own repository only, so each worktree's freshness check reads its own.
**Considered:** a row per sibling — the phantom row the finding names, a dispatch that never ran; the guard counting any canonical row for a sibling — drops kan-939's freshness check for every sibling worktree.

### A re-dispatch at an unchanged HEAD re-runs the failed specs alone

**ID:** rerun-failed-specs-at-unchanged-head
**Status:** active
**Chosen:** when every worktree of the resolved set is at the HEAD the previous dispatch verified and that report's `verify` exit was non-zero, the parent names `rerun specs:` with the failing spec paths from that report; the verifier runs those alone, and their exit is `verify`'s — a non-zero one still blocks.
**Considered:** Playwright's `--last-failed` — it reads a results file every capture run rewrites, so the set it re-runs is not the one the report named; keeping the full re-run — the cost the finding measured.

### Fix the group-session flake from its trace

**ID:** addtobuilder-trace-first
**Status:** active
**Chosen:** Task 8 reproduces on repeated runs against a live stack, reads the failing trace, and fixes the race it shows in `addToBuilder`; the leading hypothesis is a row tap during the skeleton-to-list reflow, fixed as `filterTo` fixes its own.
**Considered:** a longer `waitForTag` timeout or a retry around the tap — a widened tolerance (`rules/fix-determinism-at-the-source.mdc`).

### Repair the red guard suite with here-strings

**ID:** grep-pipe-here-string
**Status:** active
**Chosen:** the two `printf '%s' "$OUT" | grep -q` pipes in `scripts/test-check-task-records.sh` become `grep -q … <<<"$OUT"`, the form `TestNoPipeIntoEarlyExitGrep` prescribes.
**Considered:** exempting the file in the Go test — weakens the guard.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — the change moves a Go guard integrate gates on, a runner every inline verify uses, and a harness fix whose cause the plan could not reproduce.
**Considered:** flow-low — Tasks 1–3, 5 and 7 are prose or one-line edits, but Tasks 4, 6 and 8 change behaviour a low-effort pass could under-read.

### Three fixer groups, one per stretch of worktrees

**ID:** fixer-groups
**Status:** active
**Chosen:** three groups, in series: g1 — Tasks 1 and 2, the gymie and gymie-frontend configuration moves (`self-review-kan-983-render-screens-from-cached-data-first-refresh-fix-g1`); g2 — Tasks 3–7 in order, the agents worktree (`-fix-g2`); g3 — Task 8, the gymie-playwright harness fix against a live stack (`-fix-g3`).
**Considered:** one group — one fixer would carry Task 8's long stack-and-trace session on top of five agents tasks; one group per task — eight dispatches for work three finish.

## Open questions

## Live check

Not applicable: the change is guards, prose and a test harness — the Go guard is built from the
checkout on every call, never served by `flowd`, and Task 8's live stack is its own reproduction,
not a running system this change alters.
