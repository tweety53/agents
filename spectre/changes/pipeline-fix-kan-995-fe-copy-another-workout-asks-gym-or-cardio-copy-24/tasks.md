# pipeline-fix-kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy-24

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** a `/flow` run never ends on a "still open" or "your call" item the run could have done
itself — the land guard refuses a `noted` finding that defers to the operator, and the self-review
skill classifies such an observation as a finding, repairs the data a defect corrupted, and syncs a
`--base` missing a fix.

**Architecture:** one refusal added beside `LAND-UNTRACED-FINDING` in
`stats/internal/guard/landselfreviewreport.go`; prose in `skills/flow-self-review/SKILL.md` steps 2,
3 and 7, `skills/flow/cleanup.md`, and the two citing lines in `skills/flow/integrate.md` and
`skills/flow-fast/SKILL.md`.

**Spec:** `design.md` beside this file; the brainstorm record is
`.superpowers/sdd/2026-10-20-pipeline-fix-kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy-24-design.md`
in this worktree.

## Global constraints

- Pipeline-fix commits: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan and lists every sentence Tasks 1 and 2 reword or add. Never merge or push. No `Co-Authored-By` or other attribution trailer.
- Every task runs in this worktree, `/Users/tweety53/Projects/agents-worktrees/pipeline-fix-kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy-24` (branch `pipeline-fix-kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy-24`). Never touch a main checkout, and never create a worktree.
- Go: run `cd stats && gofmt -w .` before committing; `gofmt -l .` prints nothing and `go vet` stays clean. Open every Go test run with `make -s web-dist` (`stats/Makefile`). Never start, stop or restart the `flowd` on `127.0.0.1:4173` and never write to the live store.
- Prose: insert each block exactly as written — `scripts/check-verbatim-moves.sh` matches the sentences `verbatim-moves.txt` lists, and a reworded one fails it.
- No new dependency.

## Review Focus

- Task 1: only a `noted` line is refused — a `filed:`, `fixed:` or `declined` line saying "operator decision" still lands; the refusal exits 4 before anything is staged, like the untraced one.
- Task 2: the base sync pushes only on a clean merge and a green run, and never resolves a conflict.

- [ ] 1. Refuse a noted self-review finding that leaves its action to the operator
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run the **Live check** in `design.md` against the unmodified tree: it prints `LAND-NOTHING-TO-COMMIT` and exits 0 — kan-995's deferred line lands.
  - [ ] **Step 2: Write `TestLandSelfReviewReportRefusesDeferredFinding`** in `stats/internal/guard/land_self_review_report_test.go`, after `TestLandSelfReviewReportRefusesUntracedFinding` and shaped like it (same repo setup, same `Registry["land-self-review-report"]` call): the report `docs/self-review/kan-d-self-review.md` holds `## 1. Problems — \`flow-fix\`\n\n` then the line `- **[flow-fix]** \`feature/cardio\` lacks develop's devRestart isolation fix 98da4aad; a branch decision for the operator — noted`; assert rc 4, empty stdout, stderr exactly `LAND-DEFERRED-FINDING: docs/self-review/kan-d-self-review.md:3: <that line> — fix it in the run, never leave it to the operator; nothing added, committed, pulled or pushed\n`, and `HEAD` and the index unchanged. Then rewrite the report to three lines — the same finding ending `— fixed: abc1234`, `- **[flow-fix]** resolving it required an operator decision about another change's work — filed: KAN-56` and `- **[flow-fix]** it needed an operator decision to skip — declined` — and assert the land returns 0. Run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestLandSelfReviewReportRefusesDeferredFinding' -count=1`: it fails.
  - [ ] **Step 3: Implement** in `stats/internal/guard/landselfreviewreport.go` (decision `deferred-finding-guard`): in `lsrrUntraced`'s loop, after the untraced check, a second check on the same line — and extend the function's doc comment with one sentence naming the deferred refusal and this change:

  ```go unverified:compiles and gofmt-clean in lsrrUntraced's loop, l the lowered line it already computes
  		if strings.HasPrefix(line, "- **[") && strings.HasSuffix(strings.TrimSpace(l), "— noted") && lsrrDefers(l) {
  			fmt.Fprintf(stderr, "LAND-DEFERRED-FINDING: %s:%d: %s — fix it in the run, never leave it to the operator; nothing added, committed, pulled or pushed\n", addPath, i+1, strings.TrimSpace(line))
  			rc = 4
  		}
  ```

  and, beside the function:

  ```go unverified:compiles; strings is already imported
  // lsrrDefers reports whether a lowered finding line leaves its action to
  // the operator — a noted observation the run could have acted on
  // (pipeline-fix-kan-995-…-24).
  func lsrrDefers(l string) bool {
  	for _, p := range []string{"for the operator", "operator decision", "operator's call", "operator's choice", "your call", "if you want"} {
  		if strings.Contains(l, p) {
  			return true
  		}
  	}
  	return false
  }
  ```

  - [ ] **Step 4: Edit `scripts/land-self-review-report.sh`'s header**: the exit-code `4` entry becomes exactly:

  ```bash verified:authored in-tree for this change
  #   4  an open finding — one LAND-UNTRACED-FINDING line per finding line
  #      saying "not traced" or "untraced", and one LAND-DEFERRED-FINDING
  #      line per "— noted" finding line leaving its action to the operator
  #      ("for the operator", "operator decision", "operator's call",
  #      "operator's choice", "your call", "if you want"); checked right after
  #      the start-of-run branch assert, so nothing is added, committed,
  #      pulled or pushed
  ```

  - [ ] **Step 5: Edit the prose.** In `skills/flow-self-review/SKILL.md`, step 7, directly below the line opening ``` `land-self-review-report.sh` refuses a report whose finding line says ```, add this line:

  ```markdown verified:applied at plan time; check-verbatim-moves.sh, check-installed-citations.sh and check-markdown-integrity.py ran clean of every hit but the sentences verbatim-moves.txt lists
  It refuses a `noted` finding line that leaves its action to the operator — one saying `for the operator`, `operator decision`, `operator's call`, `operator's choice`, `your call` or `if you want` — the same way, `LAND-DEFERRED-FINDING`, exit 4, nothing committed: fix that finding per step 3 (`skills/flow-self-review/SKILL.md`), rewrite its line, and land again.
  ```

  In `skills/flow/integrate.md` and in `skills/flow-fast/SKILL.md`, the words ``A `LAND-UNTRACED-FINDING` refusal is not a stop:`` become ``A `LAND-UNTRACED-FINDING` or `LAND-DEFERRED-FINDING` refusal is not a stop:``, the rest of each line unchanged.
  - [ ] **Step 6: Verify.** `cd stats && gofmt -l . && make -s web-dist && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestLandSelfReviewReport' -count=1` passes; the **Live check** in `design.md` now exits 4 with exactly one `LAND-DEFERRED-FINDING` line, for line 15 of the kan-995 report; `scripts/check-references.sh`, `scripts/check-markdown-integrity.py` and `scripts/check-installed-citations.sh` exit 0.
  - [ ] **Step 7: Commit** the six files below with the subject below.

**Build:** green
**Files:** `stats/internal/guard/landselfreviewreport.go`, `stats/internal/guard/land_self_review_report_test.go`, `scripts/land-self-review-report.sh`, `skills/flow-self-review/SKILL.md`, `skills/flow/integrate.md`, `skills/flow-fast/SKILL.md`
**Tests:** `TestLandSelfReviewReportRefusesDeferredFinding`
**Regression:** `TestLandSelfReviewReportRefusesDeferredFinding` — reverting the commit lets kan-995's deferred line land, and its rc-4 assertion fails.
**Baseline:** before=2 after=3
<!-- measured: grep -c '^func Test' stats/internal/guard/land_self_review_report_test.go @ b2555b98 (2) -->
**Commit:** feat(guard): refuse a noted self-review finding left to the operator
**After:** none

**Decision:** deferred-finding-guard

- [ ] 2. Self-review fixes an actionable observation, repairs corrupted data, and syncs a base missing a fix
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `git grep -n -F -e 'An observation that names' -e 'A fix repairs what its defect' -e 'The base sync' -- skills/flow-self-review/SKILL.md` prints nothing, and `git grep -n -F -e 'left as the operator' -- skills/flow/cleanup.md` prints nothing.
  - [ ] **Step 2: Step 2 of `skills/flow-self-review/SKILL.md`** (decision `observation-with-action-is-a-finding`): the text `is ever offered for filing**.` ending the sentence that opens "A finding that names no defect" is followed, on the same paragraph, by:

  ```markdown verified:applied at plan time; check-verbatim-moves.sh, check-installed-citations.sh and check-markdown-integrity.py ran clean of every hit but the sentences verbatim-moves.txt lists
  **An observation that names
  an action the run can take is a fixable finding, never `noted`** — never an operator decision, a
  "your call" or a "still open" line — unless the action is `big`, reaches production
  (`rules/never-touch-production.mdc`) or deletes what cannot be restored (kan-995: "feature/cardio
  lacks develop's fix 98da4aad; a branch decision for the operator" was recorded `noted` and relayed
  as an open decision, and the operator then had the run merge the branch).
  ```

  - [ ] **Step 3: Step 3, the planner paragraph** (decision `fix-repairs-corrupted-data`): directly after `the operator answered "fix them now").` and before `A finding a re-plan`, insert the block below on lines of its own, so `A finding a re-plan` starts the line after it:

  ```markdown verified:applied at plan time; check-verbatim-moves.sh, check-installed-citations.sh and check-markdown-integrity.py ran clean of every hit but the sentences verbatim-moves.txt lists
  **A fix repairs what its defect already wrote wrong** — a store row, a record or a file the defect
  left corrupted is repaired in the same plan, a small command added where none exists, never left
  as a forward-only fix with the repair an operator's option (kan-995: decision `k995-rows-left`
  left three mis-stamped ledger rows wrong, and `flowd restamp-dispatch`, `b2555b98`, repaired them
  once the operator asked).
  ```

  - [ ] **Step 4: Step 3, the base sync** (decision `base-sync-by-session`): in the sentence opening `**This session plans, changes and reviews nothing itself**`, the words `save item 4's ticks and archive commit:` become `save item 4's ticks and archive commit and **The base sync** below:`; and after item 5's last line (`merges.`), before `### 4. Explain, then ask`, insert this paragraph with a blank line on each side:

  ```markdown verified:applied at plan time; check-verbatim-moves.sh, check-installed-citations.sh and check-markdown-integrity.py ran clean of every hit but the sentences verbatim-moves.txt lists
  **The base sync.** A finding that a `--base` change's `<base>` (`skills/flow/SKILL.md`) lacks a
  commit the project's `<default-branch>` carries is no task of the loop: in each repository of the
  change whose `<base>` lacks it, this session adds a detached worktree of `origin/<base>`, named
  `base-sync-<name>`, in that repository's worktrees directory, runs
  `git merge --no-edit origin/<default-branch>` in it, then the `## lint` and `## test` commands
  `<project>/.flow/project.md` declares for that repository; on a clean merge and a green run it
  pushes the merge to `<base>` on `origin` without asking, removes the worktree, and records the
  finding `fixed` with the merge commit's sha. A merge that conflicts, or a red run, pushes nothing,
  removes the worktree, and is offered in step 4 as a `big` finding is.
  ```

  - [ ] **Step 5: `skills/flow/cleanup.md`**: in the paragraph opening `**The Finished handoff, and the run's final report beside it, carry no open item**`, the words `"still open", unverified or untraced.` become `"still open", unverified, untraced or left as the operator's call.`
  - [ ] **Step 6: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-normative-inventory.sh` and `scripts/check-ask-silence.sh` exit 0; Step 1's two greps each print their new lines.
  - [ ] **Step 7: Commit** the two files below with the subject below.

**Build:** green
**Files:** `skills/flow-self-review/SKILL.md`, `skills/flow/cleanup.md`
**Tests:** **none** — a contract wording change; the verbatim, reference, citation and markdown guards cover it, and Task 1's guard refuses the report line it rules out.
**Regression:** none — no test is added; reverting restores the wording Step 1's greps find absent.
**Commit:** fix(flow-self-review): act on actionable observations, repair corrupted data
**After:** Task 1

**Decision:** observation-with-action-is-a-finding
**Decision:** fix-repairs-corrupted-data
**Decision:** base-sync-by-session
