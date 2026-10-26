# kan-1025-flow-improvement-require-a-suite-regression-test

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** a fixed Critical/Important panel finding cannot close without a tracked suite test (or a stated reason none can exist).

**Spec:** `design.md` beside this file.

## Global constraints

- Work in `/Users/tweety53/Projects/agents-worktrees/kan-1025-flow-improvement-require-a-suite-regression-test`, branch `spectre/kan-1025-flow-improvement-require-a-suite-regression-test`; paths are relative to its root. Never touch the main checkout; never create a worktree.
- Run `gofmt -w` on every Go file changed; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.
- Go test commands run as `cd stats && make -s web-dist && go test ./internal/guard -run '<name>' -count=1`, Bash timeout 600000.

- [x] 1. Refuse a panel close whose fixed Critical/Important finding names no tracked regression test
  - [x] **Step 1: Write the failing test** `TestCheckPanelFindingsClosedFixedNeedsARegressionTest` in `stats/internal/guard/check_panel_findings_closed_test.go`, following `TestCheckPanelFindingsClosedFixedReproducerExists`'s fixtures (a git-initialised scratch worktree, faked findings/dispatches with every other check satisfied). Subtests: fixed Important with no `regression-test: F1` line in any `.superpowers/sdd/panel-fix-report-*.md` → exit 1, stderr names `F1`; line naming an untracked path → 1; line naming a path under `.superpowers/` → 1; line naming a committed `foo_test.go` → 0; `regression-test: F1 — none — prose-only fix` → 0; fixed Minor with no line → 0. Run it; the exit-1 cases fail.
  - [x] **Step 2: Implement.** In `checkPanelFindingsClosed` (`stats/internal/guard/panelfindingsclosed.go`), after the fixed-reproducer block: read every `panel-fix-report-*.md` under `<worktree>/.superpowers/sdd/` (absent dir = no lines), collect `regression-test: F<n> — <rest>` lines (em-dash separators). For each finding with `status == "fixed"` and severity `Critical` or `Important` (case-insensitive): no line → missing; a line whose second field is `none` passes; otherwise the path field must not start with `.superpowers/` and must be tracked — `git -C <tree> ls-files --error-unmatch -- <path>` exits 0 in the worktree or in one of the recorded worktrees (reuse the lazy state-record read the fixed-reproducer block uses, hoisting it into a shared closure when both blocks need it). Print `<cfcPrefix>fixed Critical/Important finding(s) with no tracked regression test: F1 (<reason>) …` and set `violated`. Add a comment block in the file's style citing kan-1019.
  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./internal/guard -run 'TestCheckPanelFindingsClosed' -count=1` passes; `go vet ./internal/guard && gofmt -l internal/guard` prints nothing. Mutation: comment out the new block's `violated = true`, re-run — `TestCheckPanelFindingsClosedFixedNeedsARegressionTest` fails; restore.
  - [x] **Step 4: Commit** the two files.

**Build:** green
**Files:** `stats/internal/guard/panelfindingsclosed.go`, `stats/internal/guard/check_panel_findings_closed_test.go`
**Tests:** `TestCheckPanelFindingsClosedFixedNeedsARegressionTest`
**Regression:** `TestCheckPanelFindingsClosedFixedNeedsARegressionTest` — reverting lets a fixed Important with no regression-test line close with exit 0.
**Baseline:** before=5 after=6
<!-- measured: grep -cE '^func Test[A-Z]' stats/internal/guard/check_panel_findings_closed_test.go @ 97a55ef9 -->
**Commit:** fix(guard): refuse a panel close whose fixed finding names no tracked regression test
**After:** none

**Decision:** regression-test-report-line
**Decision:** guard-checks-tracked-not-red
**Decision:** none-form-allowed

- [x] 2. State the regression-test rule in the panel contracts
  - [x] **Step 1: Edit `skills/flow/review-panel.md`.** In the REPORT FILE paragraph (the `> **REPORT FILE:**` block before "The fix subagent's own dispatch is recorded too"), after the sentence ending "the failure that run shows with the fix reverted", add: "Each fixed Critical or Important finding also carries one line, `regression-test: F<n> — <repo-relative path> — <test name>`, naming the git-tracked suite test that fails at the pre-fix code — a reproducer promoted into the suite, never one left under `.superpowers/` — or `regression-test: F<n> — none — <reason>` where no suite test can exist." In the close-guard paragraph beginning "Exit 0 proceeds to the stage close below", extend the "It fires too when…" sentence with: "or when a `fixed` Critical or Important finding has no `regression-test:` line in any `panel-fix-report-*.md`, or names a path that is untracked or under `.superpowers/`".
  - [x] **Step 2: Edit `skills/flow/review-panel-fix-round.md`.** After the paragraph beginning "**A fix round that fixed a Critical or Important finding and changed no executable behaviour", add: "**Every fixed Critical or Important finding leaves a tracked suite test behind** — the `regression-test:` line of the REPORT FILE paragraph (`skills/flow/review-panel.md`), so a regression of a reviewed defect is still caught once the worktree and its reproducers are gone (kan-1019). Where the parent fixes inline, it writes those lines to `panel-fix-report-<round>.md` itself. `check-panel-findings-closed` refuses the close without them; a fixed Minor needs none."
  - [x] **Step 3: Edit `scripts/check-panel-findings-closed.sh`'s header comment** — add the new exit-1 violation to its list, in the list's existing style.
  - [x] **Step 4: Verify.** `scripts/check-references.sh` and every other `scripts/check-*.sh` guard `.flow/project.md`'s `## lint` names exit clean.
  - [x] **Step 5: Commit** the three files.

**Build:** green
**Files:** `skills/flow/review-panel.md`, `skills/flow/review-panel-fix-round.md`, `scripts/check-panel-findings-closed.sh`
**Tests:** **none** — prose contract and a script header comment.
**Regression:** none — no test; Task 1's test pins the behaviour.
**Commit:** docs(flow): require a tracked regression test for every fixed Critical or Important finding
**After:** Task 1

**Decision:** inline-parent-writes-lines
