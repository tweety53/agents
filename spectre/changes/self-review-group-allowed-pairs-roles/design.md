## Context

The self-review fix loop of `group-allowed-pairs-roles`, planned on the change's own branch
`spectre/group-allowed-pairs-roles` in
`/Users/tweety53/Projects/agents-worktrees/group-allowed-pairs-roles`; its fix commits ride that
run's own route. The pass is `.superpowers/sdd/self-review-pass.md` in that worktree; the
brainstorm trail is `.superpowers/sdd/2026-10-30-self-review-group-allowed-pairs-roles-design.md`.
Every question was auto-resolved to its recommended option.

**F1 — what the brief scopes today.** **The mutation-testing brief**
(`skills/flow/review-panel.md`) tells the slot to mutate each changed behaviour and run the tests,
with no scope for those runs. Its whole-package / name-prefix sentence ("When one run of that whole
package takes more than a third of the slot's ceiling …") is about a surviving mutant's reproducer
command, one run. The slot itself runs one test run per mutant, so its cost is runs times mutants.

**F2 — the prototype frame for a region.** Step 10 of `skills/flow/visual-verify-verifier.md`
renders the prototype "at the capture test's own viewport, device scale factor and colour scheme".
The change's prototype `.panel` carries a standalone `max-width`; the capture's region is wider, so
the first verifier's pair was unreadable and `visual-verify-rest` re-rendered with
`.panel{max-width:none}` injected. `scripts/compose-mockup-frames.sh`'s header states that `frame
none` composes at native size, uncropped, with an n/a ratio and no `.frame.png`; step 10 does not,
and the first verifier reported it as a defect. Step 8 ("Every screenshot this spec takes is the
full page or viewport, never a clipped region") and the capture-task paragraph of
`skills/flow/brainstorm-planner.md` contradict a prototype that draws one region of a view.
**Nothing to repair:** the change's roles captures (`experiments-roles*-darwin.png`) sit beside the
full-page `experiments-open-darwin.png` of the same view, which the exception admits.

**F3 — the trace.** The round-1 reports (`.superpowers/sdd/verify-report-visual-verify.md`,
`verify-report-visual-verify-rest.md`) name only the three roles captures. The change moved
`experiments-open`, `-closed`, `-hover` and `-focus-darwin.png` too
(`git diff --stat f1163fad 0197cb92 -- stats/web/tests`), and no round-1 line names them, nor any
`per view, no frame composed: <view> sweeps:` line. Step 9 already says "Read every printed path …
and state, per view, what was seen", the report template carries the capture-scope sweeps line, and
**Blocking** (`skills/flow/visual-verify.md`) blocks on a captured view no frame was composed for
whose report carries none — but `stats/internal/guard/verifyreport.go` checks neither, so both
reports passed. Round 2 named `experiments-open` and saw the clipped labels; rounds 2 and 3 still
left PNGs unnamed (roles-hover and roles-focus in round 2; closed, hover and focus in round 3). The
round-1 dispatch prompts are not in the store — `flow record dispatches -change
group-allowed-pairs-roles` holds keys, pairs and metrics only — so whether the brief narrowed the
verifier cannot be read; the reports alone show the unenforced requirement. Cause: a prose-only
requirement with no guard, not reviewer variance.

**F4 — the zip rebuild.** `.flow/project.md`'s `capture` row runs Playwright alone; step 8 of the
verifier, `skills/flow/verify-fix-loop.md` and `skills/flow/sync-onto-base.md` each tell the caller
to run `rebuild-full-app-suite-zip.sh <worktree> <screenshots root>` afterwards, and the script
refuses a relative root. The verifier appends `--update-snapshots=all`, and `-g '<titles>'` for the
full app suite, last on the `capture` command line, so a rebuild chained with a bare `&&` would
receive those arguments instead of Playwright.

## Decisions

### Scope the slot's mutant runs by planned runs, not one run

**ID:** sr1-slot-runs-by-mutant-count
**Status:** active
**Chosen:** one added sentence in the brief: before the first mutation the slot measures one run of the mutated code's whole package, and when that run's seconds times the mutants it plans pass half the slot's ceiling, every mutant run filters by the test-family name prefix, never a full test name — the measured run and the planned count are both known before the first mutation.
**Considered:** rewording the reproducer sentence to cover both — the reproducer is one run, its third-of-ceiling threshold is right for it, and a reword changes a rule nobody found wrong; a fixed per-run timeout — a timeout is a tolerance, and it would still let several slow mutants exhaust the ceiling.

### Admit a region capture beside the full-page capture

**ID:** sr2-region-capture-exception
**Status:** active
**Chosen:** one exception sentence after each "never a clipped region" sentence (step 8 of the verifier, the capture-task paragraph of the planner): a region an approved prototype draws alone is captured as an element, beside the view's own full-page capture, never in place of it.
**Considered:** render the prototype into the full view instead — the prototype draws one panel, not the page, so no full-page frame exists to compare against; a clip check in `verifyreport.go` — the clip rule is now conditional on the frame source, which no report line states mechanically, and no second incident asks for it.

### Size the prototype's region and state the frame none reads

**ID:** sr2-prototype-region-width
**Status:** active
**Chosen:** two step-10 sentences: the render script sizes the prototype's region to the capture's width by injected CSS overriding a standalone width cap, recorded in `visual-verification.md`; under `frame none` the compose step prints an n/a ratio and writes no `<stem>.frame.png` by design, and the band, seam and matrix reads run `measure-visual-properties.sh` on the prototype PNG and the capture directly.
**Considered:** resize the capture to the prototype's width — the capture is the app as built, and only the frame is the render's to shape; leaving (b) to the compose script's header — the verifier reads step 10, not the script header, and called the n/a a defect.

### Enforce per-capture accounting in check-verify-report

**ID:** sr3-every-capture-accounted
**Status:** active
**Chosen:** `check-verify-report.sh`, given a report with a `- spec: <path>` line, runs `resolve-visual-screenshots` in-process for the worktree three directories above the report and that spec's basename; on its exit 0, every printed PNG needs a `## Report` line `- <view>: <that absolute path>`, and every such view that no `- <view>: <composite> diff=` line names needs a `per view, no frame composed: <view> sweeps:` line; each gap is undone. A report with no `- spec:` line, a wire-contract report, or a resolver exit other than 0 is not checked (step 9 blocks on the resolver's own exit).
**Considered:** a sentence in step 9 alone — the rule is already prose in three places and two verifiers broke it; a `--captures <file>` argument fed by the parent — one more plumbing step to forget, where the guard can resolve the same list itself; require only the changed captures — needs the merge base the guard is not given, and **Blocking** already says every captured view.

### Chain the zip rebuild into this repository's capture row

**ID:** sr4-capture-chains-zip
**Status:** active
**Chosen:** the `capture` row defines a shell function `capture() { npx playwright test "$@" && ../../scripts/rebuild-full-app-suite-zip.sh "$(cd ../.. && pwd)" "$PWD/tests/visual"; }` and calls `capture <spec> --update-snapshots`, so appended arguments still reach Playwright and the rebuild gets absolute arguments; one step-8 sentence says a chained `capture` has run the rebuild and both arguments are absolute. The `verify-fix-loop.md` and `sync-onto-base.md` sentences stay: they are true for every project — other projects' rows do not chain the rebuild — and a second rebuild rewrites nothing.
**Considered:** a bare `&& rebuild …` after Playwright — the appended `--update-snapshots=all` and `-g` would land on the rebuild script, which refuses them; a Playwright `globalTeardown` — it would run on every `verify` and developer run too; a new `scripts/capture.sh` wrapper — a new file and its test for what one row does; rewording the three skill sentences to drop the step — wrong for every project whose row does not chain it.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 3 changes a guard every visual-verify report passes through, and Task 4 changes the `capture` row every capture in this repository runs.
**Considered:** `flow-low` — the findings are not all mechanical.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, key `self-review-group-allowed-pairs-roles-fix`, Tasks 1–4 in plan order — Tasks 2, 3 and 4 share `skills/flow/visual-verify-verifier.md` and run in order, and the four are small.
**Considered:** a separate docs group for Tasks 1 and 2 — the fix row's one pair is `opus/flow-low` either way, so a second dispatch buys nothing.

## Open questions

## Recommended approach

Each finding is closed where its run reads it: one brief sentence scopes the slot's mutant runs
(sr1-slot-runs-by-mutant-count); step 8, step 10 and the planner state the region-capture exception,
the region width and the `frame none` reads (sr2-region-capture-exception,
sr2-prototype-region-width); `check-verify-report.sh` refuses a report that leaves a capture
unnamed or unswept, and step 9 names that refusal (sr3-every-capture-accounted); the `capture` row
chains the zip rebuild through a function so the appended flags still reach Playwright
(sr4-capture-chains-zip). Rewording the reproducer rule, a clip guard, a parent-fed capture list, a
teardown hook and a wrapper script were each rejected in their decisions.

## Critical files

- `skills/flow/review-panel.md` — **The mutation-testing brief**: Task 1's sentence.
- `skills/flow/visual-verify-verifier.md` — steps 8, 9 and 10: Tasks 2, 3 and 4's sentences.
- `skills/flow/brainstorm-planner.md` — the capture-task paragraph: Task 2's exception sentence.
- `stats/internal/guard/verifyreport.go` — `checkVerifyReport`: Task 3's accounting check.
- `stats/internal/guard/check_verify_report_test.go` — Task 3's test.
- `scripts/check-verify-report.sh` — the guard's header contract: Task 3's paragraph.
- `.flow/project.md` — the `capture` row and its explanatory paragraph: Task 4.
- `spectre/changes/self-review-group-allowed-pairs-roles/verbatim-moves.txt` — the seven new
  run-loaded sentences, committed with this plan.

## Existing code to reuse

- `resolveVisualScreenshots` (`stats/internal/guard/resolvevisualscreenshots.go`) — the resolver
  Task 3 calls in-process; `rvsSection` (`stats/internal/guard/resolve_visual_screenshots_test.go`)
  builds the test's `## visual verification` section.
- `vrFrameLine` (`stats/internal/guard/verifyreport.go`) — the composed-frame ids Task 3 reads; the
  three-up worktree derivation already used for `--design`.
- `runGuard` (guard package tests) — the test harness.
- `scripts/rebuild-full-app-suite-zip.sh` — chained by Task 4, unchanged.

## Verification

- `cd stats && go test ./internal/guard/ -run 'TestCheckVerifyReport' -count=1` green, the new test
  failing with `verifyreport.go` reverted.
- `scripts/check-verify-report.sh /Users/tweety53/Projects/agents-worktrees/group-allowed-pairs-roles/.superpowers/sdd/verify-report-visual-verify-rest.md 0`
  exits 1 naming `experiments-open-darwin.png` — the round-1 report the guard now refuses.
- `scripts/run-verify-list.sh /Users/tweety53/Projects/agents-worktrees/group-allowed-pairs-roles lint`
  exits 0, `check-verbatim-moves.sh` among it.
- The new `capture` row run on `tests/visual/experiments.spec.ts` with `--update-snapshots=all`
  appended exits 0 and leaves `git status --short stats/web/tests/visual` empty.

## Live check

None — no long-running service or store changes; the `capture` row runs against the disposable
UI-test stack in Task 4's own verify step, never the dev `flowd` on 4173.
