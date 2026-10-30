# self-review-group-allowed-pairs-roles

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the four non-big self-review findings of `group-allowed-pairs-roles` as one
self-review fix loop on the change's own branch.

**Architecture:** one sentence in the mutation-testing brief, four in the visual-verify verifier,
one in the planner, a per-capture accounting check in `check-verify-report`, and a `capture` row
that chains the zip rebuild.

**Tech Stack:** Go (`stats/internal/guard`), Bash shim header, Markdown skills, the project
configuration.

**Spec:** `design.md` beside this file; the brainstorm trail is
`.superpowers/sdd/2026-10-30-self-review-group-allowed-pairs-roles-design.md` in this worktree.

## Global Constraints

- Work in `/Users/tweety53/Projects/agents-worktrees/group-allowed-pairs-roles`, branch `spectre/group-allowed-pairs-roles`; every `**Files:**` path is relative to its root. Never touch a main checkout, never create a worktree, never merge or push.
- One commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`.
- Every quoted sentence below is inserted word for word: `spectre/changes/self-review-group-allowed-pairs-roles/verbatim-moves.txt`, committed with this plan, lists each one exactly, and `check-verbatim-moves.sh` fails a sentence that differs. Line wrapping may differ; words and punctuation may not.
- Run `cd stats && gofmt -w .` on Go changes before committing; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` stay clean.
- Never start, stop or restart the dev `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container. Task 4 may start the disposable UI-test stack with `make -C stats ui-test-up`, as the `capture` row itself does.
- A task whose `**Files:**` holds a Markdown file under `skills/` runs `scripts/run-verify-list.sh /Users/tweety53/Projects/agents-worktrees/group-allowed-pairs-roles lint` in its verify step.

## Review Focus

- Task 3: a report with no `- spec:` line, a wire-contract report and a resolver exit other than 0 are never checked; a composed frame's view needs no capture-scope line; the existing `TestCheckVerifyReport` cases stay green unchanged.
- Task 4: arguments appended after `<spec>` reach Playwright, never the rebuild script; a failing Playwright run skips the rebuild and fails the row.

- [ ] 1. Scope the mutation slot's own mutant runs by its planned mutant count
  - [ ] **Step 1: Edit** (decision `sr1-slot-runs-by-mutant-count`) `skills/flow/review-panel.md`: in **The mutation-testing brief**, after the paragraph's last sentence, the one ending `(kan-1039: the slot discarded a whole cached sweep whose survivals were false).`, add this sentence on its own line in the same paragraph:

```markdown verified:authored for this change; the anchor sentence read from skills/flow/review-panel.md at branch spectre/group-allowed-pairs-roles
The slot's own mutant runs are scoped before the first mutation too: it measures one run of the mutated code's whole package, and when that run's seconds times the number of mutants it plans pass half the slot's ceiling, every mutant run filters by the name prefix the mutated code's test family shares (`-run '^TestCheckPanel'`), never a full test name (group-allowed-pairs-roles self-review: the slot ran the whole guard package per mutant and timed out at 15 minutes; the prefix-filtered retry finished in 88 s).
```

    Change no other text.
  - [ ] **Step 2: Verify.** `scripts/run-verify-list.sh /Users/tweety53/Projects/agents-worktrees/group-allowed-pairs-roles lint` exits 0.
  - [ ] **Step 3: Commit** with the subject below.

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** **none** — prose; the lint list's guards cover it.
**Regression:** none — a reverted sentence fails no test.
**Commit:** docs(flow): scope the mutation slot's own mutant runs by its planned mutant count
**After:** none

**Decision:** sr1-slot-runs-by-mutant-count

- [ ] 2. State the prototype region capture, its width and the frame none reads
  - [ ] **Step 1: Edit** (decision `sr2-region-capture-exception`) `skills/flow/visual-verify-verifier.md` step 8: after the sentence ending `frame and call the result a fidelity check.`, before `**\`capture\` creates this change's baseline**`, insert:

```markdown verified:authored for this change; the anchor read from skills/flow/visual-verify-verifier.md step 8 at branch spectre/group-allowed-pairs-roles
The one exception is a region an approved prototype draws alone (step 10): its element capture is the frame-bound capture, taken beside the view's own full-page capture, never in place of it (group-allowed-pairs-roles self-review).
```

  - [ ] **Step 2: Edit** (decision `sr2-prototype-region-width`) the same file, step 10's **An approved prototype is the frame where no mockup is** paragraph: after the sentence ending `` `<changeRoot>/visual-verification/prototype/<screenshot stem>.png`. ``, before `Then write`, insert:

```markdown verified:authored for this change; the anchor read from skills/flow/visual-verify-verifier.md step 10 at branch spectre/group-allowed-pairs-roles
For a region capture (step 8's exception) the script sizes the prototype's region to the capture's width, injecting CSS that overrides any standalone width cap such as a `max-width`, and `visual-verification.md` records the injected rule under the frame's entry (group-allowed-pairs-roles self-review: a 1100px prototype panel beside a 1232px capture left no band, seam or matrix reading possible).
```

    Then, in the same paragraph, after the sentence ending `` of `<abs-worktree>/<mockups>`. ``, before `Every frame-bound reading below`, insert:

```markdown verified:authored for this change; the anchor read from skills/flow/visual-verify-verifier.md step 10 and the frame none contract from scripts/compose-mockup-frames.sh's header at branch spectre/group-allowed-pairs-roles
Under that `frame none` line the compose step pairs each capture with its frame at native size, prints each pair's `diff=` ratio as n/a and writes no `<stem>.frame.png` — by design, never a defect — so the band, seam and matrix reads run `measure-visual-properties.sh` on the prototype PNG and the capture directly.
```

  - [ ] **Step 3: Edit** (decision `sr2-region-capture-exception`) `skills/flow/brainstorm-planner.md`: in the paragraph opening **Write a capture task that baselines against a mockup to measure before it commits.**, after `byte-identical pair (kan-1061 self-review).`, append inside the same `>` quote:

```markdown verified:authored for this change; the anchor read from skills/flow/brainstorm-planner.md at branch spectre/group-allowed-pairs-roles
> The one exception is a region an approved prototype draws
> alone, captured as an element beside the view's own full-page capture, per step 8 of
> `skills/flow/visual-verify-verifier.md`.
```

    Rewrap the paragraph's lines near 100 columns; change no other text in either file.
  - [ ] **Step 4: Verify.** `scripts/run-verify-list.sh /Users/tweety53/Projects/agents-worktrees/group-allowed-pairs-roles lint` exits 0.
  - [ ] **Step 5: Commit** with the subject below.

**Build:** green
**Files:** `skills/flow/visual-verify-verifier.md`, `skills/flow/brainstorm-planner.md`
**Tests:** **none** — prose; the lint list's guards cover it.
**Regression:** none — a reverted sentence fails no test.
**Commit:** docs(flow): state the prototype region capture, its width and the frame none reads
**After:** none

**Decision:** sr2-region-capture-exception, sr2-prototype-region-width

- [ ] 3. Refuse a verify report that leaves a capture unnamed or unswept
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `scripts/check-verify-report.sh /Users/tweety53/Projects/agents-worktrees/group-allowed-pairs-roles/.superpowers/sdd/verify-report-visual-verify-rest.md 0` and record its verdict line: it prints `VERIFY-REPORT-COMPLETE` today although the report names none of `experiments-open`, `-closed`, `-hover` or `-focus-darwin.png`.
  - [ ] **Step 2: Write the failing test** `TestCheckVerifyReportAccountsForEveryCapture` in `stats/internal/guard/check_verify_report_test.go`. Each case builds a temp root holding `.flow/project.md` with `rvsSection("tests/visual")` (`stats/internal/guard/resolve_visual_screenshots_test.go`), empty PNGs `tests/visual/a.spec.ts-snapshots/a-open-darwin.png` and `tests/visual/a.spec.ts-snapshots/a-roles-darwin.png`, and the report at `<root>/.superpowers/sdd/verify-report-visual-verify.md` whose `## Report` carries the all-done steps line, `- spec: <root>/tests/visual/a.spec.ts`, and:
    - both PNGs on `- a-open: <abs path> — …` and `- a-roles: <abs path> — …` lines, a `- a-roles: <root>/vv/a-roles.png diff=n/a — …` frame line with its `- a-roles matrix:` line, and `- per view, no frame composed: a-open sweeps: text done | …` → exit 0, `VERIFY-REPORT-COMPLETE`;
    - the same with the `a-open` view line removed → exit 1, the verdict naming `a-open-darwin.png`;
    - the same with the `per view, no frame composed: a-open sweeps:` line removed → exit 1, the verdict naming `a-open`;
    - the frame line and its matrix line removed while the `a-open` per-view line stays → exit 1, the verdict naming `a-roles`;
    - the first case with the `- spec:` line removed → exit 0 (not checked);
    - the first case with `screenshots` pointing at a directory that does not exist → exit 0 (resolver exit 1, not checked).

    Run it: `cd stats && go test ./internal/guard/ -run 'TestCheckVerifyReportAccountsForEveryCapture' -count=1` fails.
  - [ ] **Step 3: Implement** (decision `sr3-every-capture-accounted`) in `checkVerifyReport` (`stats/internal/guard/verifyreport.go`): read the first `- spec: <path>` line of the `## Report` section; after the line loop, when it is set and no wire contract names a view, call `resolveVisualScreenshots([]string{<worktree three directories above the report>, filepath.Base(<spec>)}, env, &buf, io.Discard)`; on exit 0, for each printed path append to `undone` `<png basename> is on no "- <view>: <path>" line — step 9 states every capture` when no `## Report` line matches `^\s*- (\S+): <quoted path>(\s|$)`, else, when that line's view is not among the `vrFrameLine` frame ids and no line contains `per view, no frame composed: <view> sweeps:`, `<view> has no frame and no "per view, no frame composed: <view> sweeps:" line`. Use `regexp.QuoteMeta` for the path.
  - [ ] **Step 4: Edit** `scripts/check-verify-report.sh`: append to the header comment, after the paragraph ending `(kan-1016: one matrix line covered nine frames).`, a paragraph stating: given a `## Report` line `- spec: <path>`, the PNGs `resolve-visual-screenshots.sh <worktree> <spec basename>` prints, the worktree three directories above <report file>, each need a `- <view>: <absolute PNG path>` line, and each such view that no `- <view>: <composite> diff=` line names needs a `per view, no frame composed: <view> sweeps:` line; each gap is undone; no `- spec:` line, a wire contract naming an endpoint, or a resolver exit other than 0 skips the check (group-allowed-pairs-roles self-review: two round-1 reports named only the prototype-framed captures).
  - [ ] **Step 5: Edit** `skills/flow/visual-verify-verifier.md` step 9: after the sentence ending `and state, per view, what was seen.`, before `An unreadable PNG is reported and blocks`, insert:

```markdown verified:authored for this change; the anchor read from skills/flow/visual-verify-verifier.md step 9 and the report template's view and per-view lines at branch spectre/group-allowed-pairs-roles
Each printed PNG gets its own `- <view>: <absolute PNG path>` line, and each view no frame is composed for its `per view, no frame composed: <view> sweeps:` line — `check-verify-report.sh` refuses a report missing either (group-allowed-pairs-roles self-review: two round-1 verifiers named only the prototype-framed captures, and a clipped chart label on an unnamed full-page capture surfaced a round later).
```

  - [ ] **Step 6: Verify.** `cd stats && go test ./internal/guard/ -run 'TestCheckVerifyReport' -count=1` green; Step 1's command now exits 1 naming `experiments-open-darwin.png`; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean; `scripts/run-verify-list.sh /Users/tweety53/Projects/agents-worktrees/group-allowed-pairs-roles lint` exits 0.
  - [ ] **Step 7: Commit** with the subject below.

**Build:** green
**Files:** `stats/internal/guard/verifyreport.go`, `stats/internal/guard/check_verify_report_test.go`, `scripts/check-verify-report.sh`, `skills/flow/visual-verify-verifier.md`
**Allowed-collateral:** `stats/**/*_test.go`
**Tests:** `TestCheckVerifyReportAccountsForEveryCapture`
**Regression:** `TestCheckVerifyReportAccountsForEveryCapture` — reverting admits a report naming no line for `a-open-darwin.png`.
**Baseline:** before=7 after=8
<!-- measured: grep -c '^func Test' stats/internal/guard/check_verify_report_test.go @ 0197cb92 -->
**Commit:** fix(guard): refuse a verify report that leaves a capture unnamed or unswept
**After:** Task 2

**Decision:** sr3-every-capture-accounted

- [ ] 4. Chain the full-app-suite zip rebuild into the capture row
  - [ ] **Step 1: Edit** (decision `sr4-capture-chains-zip`) `.flow/project.md`'s `## visual verification` command table: replace the `capture` row with

```markdown verified:spiked 2026-10-30 in stats/web of this worktree — a zsh and a bash function with this shape passed appended args (--update-snapshots=all, -g 'a b') to its first command as separate words, a failing first command failed the row, and the rebuild with these two arguments exited 0 leaving git status clean
| `capture` | `cd stats/web && { ../../scripts/check-dev-stack-fresh.sh ../.. \|\| make -C .. ui-test-up; } && capture() { npx playwright test "$@" && ../../scripts/rebuild-full-app-suite-zip.sh "$(cd ../.. && pwd)" "$PWD/tests/visual"; } && capture <spec> --update-snapshots` |
```

    and after the paragraph opening **`capture` and `verify` check the stack first.**, add the paragraph:

```markdown verified:authored for this change
**`capture` rebuilds `full-app-suite.zip` after Playwright.** The row runs Playwright inside a shell
function, so the arguments a caller appends — `--update-snapshots=all`, `-g '<titles>'` — still
reach Playwright, and on its exit 0 runs `rebuild-full-app-suite-zip.sh` with the worktree and the
`screenshots` root resolved absolute, so no recapture leaves the zip trailing its PNGs
(group-allowed-pairs-roles self-review).
```

  - [ ] **Step 2: Edit** `skills/flow/visual-verify-verifier.md` step 8: after the sentence ending ``(`skills/flow/verify-fix-loop.md`, `skills/flow/sync-onto-base.md`).``, insert:

```markdown verified:authored for this change; the anchor read from skills/flow/visual-verify-verifier.md step 8 and the relative-root refusal from scripts/rebuild-full-app-suite-zip.sh's header at branch spectre/group-allowed-pairs-roles
Both arguments are absolute — the script refuses a relative root — and a project whose `capture` row chains this rebuild after Playwright has run it with `capture`, a second run rewriting nothing (group-allowed-pairs-roles self-review).
```

  - [ ] **Step 3: Verify.** From the worktree root run the new `capture` row with `<spec>` = `tests/visual/experiments.spec.ts` and `--update-snapshots=all` appended last: exit 0, its output ending on the rebuild, and `git status --short stats/web/tests/visual` printing nothing. `scripts/check-visual-verification.sh .` exits 0; `scripts/run-verify-list.sh /Users/tweety53/Projects/agents-worktrees/group-allowed-pairs-roles lint` exits 0.
  - [ ] **Step 4: Commit** with the subject below.

**Build:** green
**Files:** `.flow/project.md`, `skills/flow/visual-verify-verifier.md`
**Tests:** **none** — a configuration row and prose; Step 3 runs the row live.
**Regression:** none — reverting restores the separate rebuild step and fails no test.
**Commit:** fix(flow-config): chain the full-app-suite zip rebuild into capture
**After:** Task 3

**Decision:** sr4-capture-chains-zip
