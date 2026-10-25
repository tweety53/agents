# kan-1032-prompt-tuning-angle

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

- [ ] 1. Go: the `prompttuning` package and `flow prompt-tuning render`
  - [ ] **Step 1: Add `stats/internal/prompttuning/prompttuning.go` — package doc naming design.md §2 as the case shape; `type Source struct{Project, Change, Jira, Repo, FixSHA, FindingRef, Slot, Severity string; Round int}` and `type Case struct{ID, Type, ReviewerKind string; Source Source; Target, TargetSHA, DefectClass, WhyPicked, Before, After string}` with snake_case JSON tags exactly as the §2 table names them (`reviewer_kind,omitempty`); `func Load(dir string) (Case, Files, error)` reading `case.json`, `finding.md`, `bad.diff`, `good.diff`, `task.md` (an absent file is an error naming it); `var RequiredFiles = []string{"case.json","finding.md","bad.diff","good.diff","task.md","index.html"}`.**
  - [ ] **Step 2: Add `func Render(c Case, f Files) ([]byte, error)` using one `html/template` page: `<title>` = the case id, inline `<style>` only (no `<script>`, no external URL), sections in design.md §2's order — type (and reviewer kind), defect class, finding with `bad.diff` in `<pre><code>`, fix with `good.diff` in `<pre><code>`, why picked, prompt update (target path, before and after in two side-by-side `<pre>` blocks that wrap under 700px). Output ends with a single newline and is byte-stable for equal input.**
  - [ ] **Step 3: Add `stats/cmd/flow/prompttuning.go` — `flow prompt-tuning render <case-dir>`: `Load`, `Render`, write `<case-dir>/index.html`; exit 0 silent, 1 on a load/render error with `flow: prompt-tuning render: <error>`, 2 on usage. Register `case "prompt-tuning"` in `stats/cmd/flow/main.go`'s dispatch and add its usage line to the commands list.**
  - [ ] **Step 4: Tests — `stats/internal/prompttuning/prompttuning_test.go`: `TestRenderEscapesAndOrders` (a `before` holding `<script>` renders escaped; the section headings appear in §2's order; no `<script` tag and no `http` URL in the output), `TestRenderIsDeterministic` (two renders byte-equal), `TestLoadNamesMissingFile` (a folder without `good.diff` errors naming it); `stats/cmd/flow/prompttuning_test.go`: `TestPromptTuningRenderWritesIndex` (a fixture folder gets an `index.html` equal to `Render`'s output; a missing argument exits 2). Verify: `cd stats && gofmt -l . && go vet ./internal/prompttuning/... ./cmd/flow/... && go test ./internal/prompttuning/... ./cmd/flow/... -run 'TestRender|TestLoadNamesMissingFile|TestPromptTuningRenderWritesIndex' -race -count=1`.**
**Build:** green
**Files:** `stats/internal/prompttuning/prompttuning.go`, `stats/internal/prompttuning/prompttuning_test.go`, `stats/cmd/flow/prompttuning.go`, `stats/cmd/flow/prompttuning_test.go`, `stats/cmd/flow/main.go`
**Tests:** `TestRenderEscapesAndOrders`, `TestRenderIsDeterministic`, `TestLoadNamesMissingFile`, `TestPromptTuningRenderWritesIndex`
**Regression:** reverting the commit removes the package and the subcommand: the four declared tests no longer compile.
**Baseline:** before=0 after=4 <!-- measured: grep -c '^func Test' over the two new test files → absent at 43b6caf7 -->
**Commit:** feat(stats): render a prompt-tuning case page
**After:** none

**Decision:** rendered-html

- [ ] 2. Guard: `check-prompt-tuning-cases`
  - [ ] **Step 1: Add `stats/internal/guard/prompttuningcases.go` — `Registry["check-prompt-tuning-cases"]`; header comment stating findings, arguments and exit codes (0 clean, 1 one `<path>: <reason>` line per violation, 2 cannot answer). Root: the optional `dir` argument, else `<FLOW_GUARD_REPO_ROOT>/docs/prompt-tuning`; an absent root is clean. For every subdirectory: each of `prompttuning.RequiredFiles` present; `case.json` parses and every field of design.md §2 is non-empty; `type` ∈ {`implementer`,`reviewer`}; `reviewer_kind` ∈ {`miss`,`false-positive`} exactly when `type` is `reviewer`; `target` begins `skills/` and `before` occurs verbatim in `git -C <repo root> show <target_sha>:<target>`; `index.html` byte-equal to `prompttuning.Render`.**
  - [ ] **Step 2: Genericity, in the same file: build the forbidden set — `source.change`, `source.jira`, every path and basename on a `+++ b/`/`--- a/` header line of `bad.diff` and `good.diff` — and refuse an `after` containing any member (case-insensitive) or matching `(?i)\bkan-\d+\b`, each hit one line naming the term.**
  - [ ] **Step 3: Ledger: when `<root>/LEDGER.md` exists, every line beginning `- ` matches design.md §5's entry shape (`- <YYYY-MM-DD> · \`<id>\` · implementer|reviewer · \`<target>\` · <class> — <summary> — applied <7-40 hex>` or `— rejected: <reason>`); a non-matching line is a violation naming its line number.**
  - [ ] **Step 4: Add the shim `scripts/check-prompt-tuning-cases.sh` modelled on `scripts/check-loop-landing.sh` (sources `lib/flow-guard.sh`, exports `FLOW_GUARD_REPO_ROOT`, `flow_guard_exec check-prompt-tuning-cases 2 "check-prompt-tuning-cases:" "$@"`), executable; add `scripts/check-prompt-tuning-cases.sh` to `.flow/project.md`'s `## lint` block after `scripts/check-self-review-report.sh`.**
  - [ ] **Step 5: Tests — `stats/internal/guard/check_prompt_tuning_cases_test.go`, one harness `TestCheckPromptTuningCases` with subtests over `t.TempDir()` fixtures in a scratch git repo: a complete generic case passes; a missing `good.diff` is caught; `reviewer` without `reviewer_kind` is caught; a `before` absent at `target_sha` is caught; an `after` naming the change, the Jira key, `KAN-12` or a diff basename is caught (four subtests); a hand-edited `index.html` is caught; a malformed ledger line is caught with its line number; an absent root exits 0. Verify: `cd stats && gofmt -l . && go vet ./internal/guard/... && go test ./internal/guard/... -run 'TestCheckPromptTuningCases' -race -count=1 && cd .. && scripts/check-prompt-tuning-cases.sh && scripts/check-guard-symlinks.sh`.**
**Build:** green
**Files:** `stats/internal/guard/prompttuningcases.go`, `stats/internal/guard/check_prompt_tuning_cases_test.go`, `scripts/check-prompt-tuning-cases.sh`, `.flow/project.md`
**Tests:** `TestCheckPromptTuningCases`
**Regression:** reverting the commit removes the guard: `TestCheckPromptTuningCases` no longer compiles, and a case naming its own Jira key in `after` passes unnoticed.
**Baseline:** before=0 after=1 <!-- measured: grep -c '^func Test' stats/internal/guard/check_prompt_tuning_cases_test.go → absent at 43b6caf7 -->
**Commit:** feat(guard): check prompt-tuning cases are complete and generic
**After:** Task 1

**Decision:** genericity-guard-in-go

- [ ] 3. Guard: admit a cases-only self-review branch in `check-loop-landing`
  - [ ] **Step 1: In `stats/internal/guard/looplanding.go`, before the flowd row checks: for a `self-review-*` branch, list `git diff --name-only origin/<default-branch>...<branch>`; when non-empty and every path begins `docs/prompt-tuning/`, print `LOOP-LANDING-OK (prompt-tuning cases only)` and return 0. Extend the header comment of the function and of `scripts/check-loop-landing.sh` with this admission and its reason (design.md §4).**
  - [ ] **Step 2: Tests in `stats/internal/guard/check_loop_landing_test.go`, beside the existing harness: `TestLoopLandingAdmitsPromptTuningOnly` — a `self-review-x` branch whose only commit adds `docs/prompt-tuning/x-1/case.json` exits 0 with no flowd rows; the same branch with one more commit touching `skills/flow/implement.md` is still refused for its missing rows. Verify: `cd stats && gofmt -l . && go vet ./internal/guard/... && go test ./internal/guard/... -run 'TestLoopLandingAdmitsPromptTuningOnly' -race -count=1`.**
**Build:** green
**Files:** `stats/internal/guard/looplanding.go`, `stats/internal/guard/check_loop_landing_test.go`, `scripts/check-loop-landing.sh`
**Tests:** `TestLoopLandingAdmitsPromptTuningOnly`
**Regression:** reverting the commit drops the admission: `TestLoopLandingAdmitsPromptTuningOnly` fails with `LOOP-LANDING-REFUSED` on the cases-only branch.
**Baseline:** before=2 after=3 <!-- measured: grep -c '^func Test' stats/internal/guard/check_loop_landing_test.go @ 43b6caf7 → 2 -->
**Commit:** feat(guard): land a self-review branch carrying only prompt-tuning cases
**After:** none

**Decision:** cases-ride-fix-branch

- [ ] 4. Self-review: the eighth angle
  - [ ] **Step 1: `skills/flow-self-review/SKILL.md` step 2 — add row 8 to the angle table exactly as design.md §1 words it; add the angle-8 paragraph after angle 7's: what it looks at (Critical/Important, a useful Minor), the two types and the reviewer kinds, the four-part picking bar, "zero cases is a normal result" and "the pass never lowers the bar". State that every case is `noted`, never fixed, offered or asked about, and give the report line shape of design.md §1. Change every "seven angles"/"seven-angle"/"all seven" in the file to eight.**
  - [ ] **Step 2: Same file — a new subsection after step 2, `### 2a. Write the prompt-tuning cases`: this session writes each case per design.md §2 (folder, files, extraction with `git -C <repo> show`, `flow prompt-tuning render`, `scripts/check-prompt-tuning-cases.sh`, a red line rewritten or the case dropped) and commits `docs(prompt-tuning): <name> cases` on the branch design.md §4 names, creating `self-review-<name>` with `pipeline-fix.sh start <name>` when step 3 has no worktree, and landing a cases-only branch by step 3 item 5's route. A pass with no case writes nothing.**
  - [ ] **Step 3: `skills/flow-self-review/SKILL-rationale.md` — a section `### Angle 8 — prompt tuning` carrying design.md's Context bullet on prompt-audit and the rejected alternatives of decisions `picking-bar`, `never-ask-in-run` and `cases-ride-fix-branch`.**
  - [ ] **Step 4: "seven" → "eight" angles in `skills/flow-contracts/finish-contract-run1.md` (the **Run the self-review pass** paragraph), `skills/flow-fast/SKILL.md` (its self-review paragraph) and `commands-claude/flow-self-review.md`.**
  - [ ] **Step 5: `stats/internal/guard/check_self_review_report_test.go` — update every fixture the canonical table now rejects for a missing `flow-prompt-tuning` section by adding that section, and add the subtest `case: report missing the angle-8 section is caught` beside case 2. Verify: `cd stats && gofmt -l . && go test ./internal/guard/... -run 'TestCheckSelfReviewReport' -race -count=1 && cd .. && scripts/check-self-review-report.sh && scripts/check-references.sh && scripts/check-markdown-integrity.sh && scripts/check-normative-inventory.sh && scripts/check-verbatim-moves.sh && scripts/check-ask-silence.sh && scripts/check-dispatch-paragraphs.sh`.**
**Build:** green
**Files:** `skills/flow-self-review/SKILL.md`, `skills/flow-self-review/SKILL-rationale.md`, `skills/flow-contracts/finish-contract-run1.md`, `skills/flow-fast/SKILL.md`, `commands-claude/flow-self-review.md`, `stats/internal/guard/check_self_review_report_test.go`
**Allowed-collateral:** `skills/flow-contracts/pipeline.md`, `skills/flow-contracts/jira-integration-finish.md`, `docs/self-review/`
**Tests:** `TestCheckSelfReviewReport`
**Regression:** reverting the commit removes row 8: the angle-8 subtest fails because a report without the section is accepted.
**Baseline:** before=1 after=1 <!-- measured: grep -c '^func Test' stats/internal/guard/check_self_review_report_test.go @ 43b6caf7 → 1 -->
**Commit:** feat(flow-self-review): add the prompt-tuning angle
**After:** Task 2, 3

**Decision:** picking-bar

**Decision:** never-ask-in-run

**Decision:** minor-when-useful

**Decision:** reviewer-both-kinds

**Decision:** cases-in-agents-repo

- [ ] 5. `/flow-prompt-tuning`: the command, the ledger and the index rows
  - [ ] **Step 1: Add `skills/flow-prompt-tuning/SKILL.md` (frontmatter `name`, `description`, `allowed-tools` for `git`, `flow`, `scripts/check-prompt-tuning-cases.sh`, `project-get.sh`; `license: MIT`) carrying design.md §5 in full: no argument, no state or stage marks, the worktree `prompt-tuning-<UTC date>` off `origin/<default-branch>`, steps 1–7 (walk, group, re-anchor, replay with its arms, grader and pass rule, apply, ledger entry shape and the one `docs(prompt-tuning): ledger <date>` commit, land by `project-get.sh <agents repo> 'default landing route'` without asking), every dispatch on `opus`, every dispatch prompt carrying the two agent-baseline sentences, and the final one-line-per-group report.**
  - [ ] **Step 2: Add `commands-claude/flow-prompt-tuning.md` modelled on `commands-claude/flow-self-review.md`, pointing at the skill; add `docs/prompt-tuning/LEDGER.md` holding a `# Prompt-tuning ledger` heading, one sentence on what an entry records (citing `skills/flow-prompt-tuning/SKILL.md`), and no entries.**
  - [ ] **Step 3: Add the `/flow-prompt-tuning` row to the skill index of `CLAUDE.md`, `AGENTS.md`, `templates/CLAUDE.md`, `templates/AGENTS.md`, to `skills/README.md`'s command table and tree, and to `README.md`'s command table — "Applies the generic implementer/reviewer prompt updates the self-review's prompt-tuning angle saved, after an A/B replay proves each; records each case in docs/prompt-tuning/LEDGER.md and removes its folder. Standalone, not a pipeline stage".**
  - [ ] **Step 4: Verify: `scripts/check-references.sh && scripts/check-markdown-integrity.sh && scripts/check-installed-rules.sh && scripts/check-installed-citations.sh && scripts/check-normative-inventory.sh && scripts/check-ask-silence.sh && scripts/check-dispatch-paragraphs.sh && scripts/check-prompt-tuning-cases.sh && scripts/check-plan-shape.sh spectre/changes/kan-1032-prompt-tuning-angle/tasks.md`.**
**Build:** green
**Files:** `skills/flow-prompt-tuning/SKILL.md`, `commands-claude/flow-prompt-tuning.md`, `docs/prompt-tuning/LEDGER.md`, `CLAUDE.md`, `AGENTS.md`, `templates/CLAUDE.md`, `templates/AGENTS.md`, `skills/README.md`, `README.md`
**Allowed-collateral:** `setup.sh`, `skills/flow-contracts/pipeline.md`
**Tests:** **none**
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** feat(flow-prompt-tuning): validate and apply saved prompt updates
**After:** Task 4

**Decision:** replay-ab-validation

**Decision:** apply-and-land

**Decision:** ledger-file

**Decision:** three-runs-per-arm

**Decision:** no-review-panel-for-tuning
