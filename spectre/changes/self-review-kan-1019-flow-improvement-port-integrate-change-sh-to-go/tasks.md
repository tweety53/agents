# self-review-kan-1019-flow-improvement-port-integrate-change-sh-to-go

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** close kan-1019's six non-big self-review findings — A (Task 1), B (Task 2), D2(a)
(Task 3), C (Task 4), E (Task 5), D1 (Task 6), as `proposal.md` names them.

**Architecture:** `integrate-change prepare` gains a live-check gate; the relocation generator reads a
branch-deleted file as empty; three suite cases pin kan-1019's parity fixes; `check-plan-shape`
gains F17; each panel slot's `end` carries its report's mtime; two multi-fixture test cases split.

**Tech Stack:** Go (`stats/`), Bash, Markdown.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-22-self-review-kan-1019-flow-improvement-port-integrate-change-sh-to-go-design.md`
in this worktree.

## Global constraints

- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-1019-flow-improvement-port-integrate-change-sh-to-go` (`<wt>` below), on branch `spectre/kan-1019-flow-improvement-port-integrate-change-sh-to-go`; every `**Files:**` path is relative to it. Never touch the main checkout and never create a worktree.
- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`. Never merge or push.
- Go: `gofmt -w` every Go file you change before committing; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- Skill prose: write every sentence a task quotes exactly as quoted — `spectre/changes/self-review-kan-1019-flow-improvement-port-integrate-change-sh-to-go/verbatim-moves.txt` acknowledges those sentences verbatim, and `scripts/check-verbatim-moves.sh` fails any other wording.
- Never widen a timeout, retry or tolerance to make a test pass (`rules/fix-determinism-at-the-source.mdc`).
- Never stop or restart `flowd` on `127.0.0.1:4173` nor the `flow-postgres` container.

## Review Focus

- Task 1: a `None — …` live-check section wrapped over several lines needs no record; a `--accept-outstanding` re-run never passes a missing or refused record; an archived re-run reads the archived `design.md`.
- Task 2: a file missing on disk but still in the index still exits 2; a deleted file's passages land as `moved` when the text went to another scoped file.
- Task 3: each case fails with its fix reverted in place, and passes with it restored.
- Task 4: a Squash-with pair declaring one shared pair is not flagged; tasks under different `measured:` commands never chain.
- Task 5: the `-ended-at` value is an instant `flow record dispatch end` accepts; the re-timed rows keep their start instants.

- [ ] 1. Stop integrate prepare while a design's live check has no passing record
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Write Step 2's two cases first and run `cd stats && go test ./internal/guard -run 'TestIntegrateChangePrepareCommit/live_check' -count=1`: `live_check`'s first check fails (prepare exits 0 with no record).
  - [ ] **Step 2: Write the cases** at the end of `TestIntegrateChangePrepareCommit`'s list in `stats/internal/guard/integrate_change_prepare_test.go`, in the harness's own shape. `live_check`: `fixture`; write `$WORK/a-worktrees/demo/spectre/changes/demo/design.md` as `# demo\n\n## Live check\n\nRun prepare twice and compare the two runs.\n\n## Decisions\n`, then `g -C "$WORK/a-worktrees/demo" add -A && g -C "$WORK/a-worktrees/demo" commit -q --amend -m "chore(spectre): plan-gate demo"`; `sut prepare "$WORK/a" demo`; check `live_check: a named live check with no record stops prepare at STOP: live-check` (`[ "$RC" -eq 1 ]`, last line opening `STOP: live-check — `); write `live-verification.md` beside it as `Before figure: not measured\n`; `sut prepare "$WORK/a" demo --accept-outstanding`; check `live_check: a refused record stops prepare, --accept-outstanding or not` (same assertion); rewrite it as `Before: two lines. After: two lines.\n`; `sut prepare "$WORK/a" demo`; check `live_check: a passing record lets prepare through` (`[ "$RC" -eq 0 ]`). `live_check_none`: `fixture`; the same `design.md` write and amend with the section body `None — the change runs per invocation;\nno service holds its state.\n`; `sut prepare "$WORK/a" demo`; check `live_check_none: a wrapped no-runtime section needs no record` (`[ "$RC" -eq 0 ]`).
  - [ ] **Step 3: Implement** per decision `live-check-gate-in-prepare` in `stats/internal/guard/integratechange_prepare.go`: `func icLiveCheckNamed(design string) bool` — false when the file cannot be read or has no line equal to `## Live check` (trailing blanks trimmed); otherwise the section runs to the next line opening `# ` or `## `, and it names a check when its first non-blank line opens with none of `None`, `No `, `Not applicable`. `func (c *ic) liveCheckGate(dir string)` — when `icLiveCheckNamed(dir+"/design.md")`: a missing `dir+"/live-verification.md"` is `c.stop("live-check", dir+"/design.md names a live check and "+rec+" is missing — run Live check (skills/flow/verify-and-handoff.md), through the stand-in its section states wherever a step would write this change's integrate state, then re-run prepare; --accept-outstanding does not pass this")`; otherwise `c.run("check-live-verification", rec)` and a non-zero `c.rc` is `c.stop("live-check", fmt.Sprintf("check-live-verification exit %d on %s — measure what it names, rewrite the record, then re-run prepare; --accept-outstanding does not pass this", c.rc, rec))`. Call it in `prepare()` right after the `ARCHIVED:` lines and before `if archived && !newwork {`, with `live`, or `arch` when `archived`. In `stats/internal/guard/integratechange.go`'s doc comment, the prepare paragraph names the gate after "the archived decision," — a design naming a live check needs `live-verification.md` passing check-live-verification, else STOP: live-check, which --accept-outstanding never passes. In `scripts/integrate-change.sh`, add `check-live-verification` to the `for g in` list after `check-foreign-staged`, and `"$SCRIPT_DIR/check-live-verification.sh"` to the last `:` line.
  - [ ] **Step 4: Write the prose** per decision `live-check-stand-in`, each sentence exactly as quoted. In `skills/flow/verify-and-handoff.md` **Live check**: after "a one-line no-runtime section, or none, skips it." insert the first sentence below; after "never proceed past it (kan-998)." insert the second. In `skills/flow/brainstorm-planner.md`'s "Write a live check…" blockquote, after "the check is not." insert the third and fourth (keep the `> ` prefix on every wrapped line). In `skills/flow/integrate.md` section **1. Check for unfinished work**, insert the fifth as a bullet before `- **\`STOP: no-verdict\`**`. In `skills/flow-contracts/finish-contract-run1.md`, after the paragraph ending "the same as a missing `check-unfinished-work.sh` verdict line above." (the visual-verify one), insert the sixth as its own paragraph.

```markdown verified:applied to a scratch clone at 1420a779; scripts/check-verbatim-moves.sh with this change's verbatim-moves.txt exits 0
A section is that no-runtime form when its first line opens with `None`, `No ` or `Not applicable`, however many lines its justification wraps to.

The check is never deferred past this stage: `integrate-change.sh prepare` stops at `STOP: live-check` while the section names something to exercise and `live-verification.md` is missing or fails `check-live-verification.sh` (kan-1019).

Every step of the check runs at `flow.verify`, before the change lands: a step that needs the change landed, a restarted shared service, or a write to the change's own integrate state is written as the stand-in that runs in its place — a scratch fixture, a workspace's own service, a scratch copy of the record — named in the section (kan-1019: a check deferred to integrate was never run).

The no-runtime line opens with `None`.

- **`STOP: live-check`** → the change's `design.md` names a live check and `live-verification.md` beside it is missing or fails `check-live-verification.sh`. Run **Live check** (`skills/flow/verify-and-handoff.md`) now, through the stand-in the section states wherever a step would write this change's integrate state, then re-run `prepare`. `--accept-outstanding` never passes it.

**Then the live-check record.** When the change's `design.md` carries a `## Live check` section naming something to exercise (**Live check**, `skills/flow/verify-and-handoff.md`), `live-verification.md` beside it exists and passes `check-live-verification.sh`; otherwise `prepare` stops at `STOP: live-check`, handled in `skills/flow/integrate.md`, and `--accept-outstanding` never passes it.
```

  - [ ] **Step 5: Verify.** `cd stats && gofmt -l .` prints nothing; `cd stats && make -s web-dist && go vet ./...` exits 0; `cd stats && go test ./internal/guard -run 'TestIntegrateChangePrepareCommit|TestIntegrateChangeHappyPath' -count=1` passes; `scripts/check-live-verification.sh spectre/changes/archive/kan-1019-flow-improvement-port-integrate-change-sh-to-go/live-verification.md` exits 0 (kan-1019's own record, committed with this plan, so a re-run of its `prepare` passes the gate); `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh`, `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-installed-citations.sh` and `scripts/check-ask-silence.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/integratechange_prepare.go`, `stats/internal/guard/integratechange.go`, `stats/internal/guard/integrate_change_prepare_test.go`, `scripts/integrate-change.sh`, `skills/flow/verify-and-handoff.md`, `skills/flow/brainstorm-planner.md`, `skills/flow/integrate.md`, `skills/flow-contracts/finish-contract-run1.md`
**Tests:** `live_check`, `live_check_none`
**Regression:** `live_check` fails — a design's live check is deferred past verify again and `prepare` lands the change with no record.
**Commit:** fix(flow-guard): stop integrate prepare while a design's live check has no passing record
**After:** none

**Decision:** live-check-gate-in-prepare
**Decision:** live-check-stand-in

- [ ] 2. Read a branch-deleted scoped file as empty in the relocation comparison
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `scripts/generate-relocation-comparison.sh <wt> <wt>/spectre/changes/archive/kan-1019-flow-improvement-port-integrate-change-sh-to-go a03f3c4b` exits 2, naming kan-1019's deleted `test-integrate-change.sh` harness as "scoped file missing from worktree".
  - [ ] **Step 2: Write the test first** in `stats/internal/guard/generate_relocation_comparison_test.go`, after case g: a `repo` whose base holds `old.md` (`## Old\n\nThis paragraph leaves a file the branch deletes outright.\n\n`) and `new.md` (`## New\n\nFiller.\n`); then `gitRun(t, dir, "rm", "-q", "old.md")` and `writeFile` `new.md` as `## New\n\nFiller.\n\nThis paragraph leaves a file the branch deletes outright.\n\n`; `gen` with `plan("yes", "old.md", "new.md")`; `ok("case g2: a file the branch deleted reads as empty text", …)` requiring a row naming `` `old.md` § Old `` and `` `new.md` § New `` with class `moved`. Relabel case g `case g: scoped file missing on disk but still tracked exits 2` (its byte-for-byte pin unchanged). Run `cd stats && go test ./internal/guard -run 'TestGenerateRelocationComparison$' -count=1` — case g2 fails with exit 2.
  - [ ] **Step 3: Implement** per decision `relocation-deleted-file-empty` in `stats/internal/guard/relocationcomparison.go`: when `grcReadWorktree` fails, `os.Lstat` of the path reports it absent, and `git -C <worktree> ls-files --error-unmatch -- <path>` (through `tcfExec`) exits 1, the after text is empty; every other failure keeps today's exit 2 and message. Update the header: the exit-2 list says a scoped file missing on disk while still in the worktree's index, and algorithm step 3 says a path absent on disk and from the index — deleted by the branch — is empty text, as an absent merge-base path is.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l .` prints nothing; `cd stats && go test ./internal/guard -run 'TestGenerateRelocationComparison$' -count=1` passes, case g still exit 2; Step 1's command exits 0 and `<wt>/.superpowers/sdd/relocation-comparison.md` carries rows whose source is the deleted `test-integrate-change.sh` harness. Commit.

**Build:** green
**Files:** `stats/internal/guard/relocationcomparison.go`, `stats/internal/guard/generate_relocation_comparison_test.go`
**Tests:** `case g2: a file the branch deleted reads as empty text`
**Regression:** `case g2: a file the branch deleted reads as empty text` fails — a relocation plan that deletes a file gets no comparison again.
**Commit:** fix(flow-guard): read a branch-deleted scoped file as empty in the relocation comparison
**After:** none

**Decision:** relocation-deleted-file-empty

- [ ] 3. Pin integrate-change's three parity fixes with suite cases
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n "CALLER_DIR\|padded\|indented" stats/internal/guard/integrate_change*_test.go` prints nothing; the shapes are `.superpowers/sdd/reproducers/0-primary-1.sh` (F1), `0-primary-2.sh` (F2) and `0-primary-3.sh` (F3).
  - [ ] **Step 2: Write the cases** per decision `parity-fix-suite-cases`. In `stats/internal/guard/integrate_change_land_test.go`'s list: `indented_fixed_row` — reproducer 2's case body with its label prefix `indented_fixed_row: ` and checks `indented_fixed_row: land exits 0` and `indented_fixed_row: an indented report row is not recorded as fixed`; `padded_orphan` — reproducer 3's case body with prefix `padded_orphan: ` and checks `padded_orphan: land exits 0` and `padded_orphan: a padded orphan id is refused and never removed`. In `stats/internal/guard/integrate_change_test.go`, a new `func TestIntegrateChangeShim(t *testing.T)` running `icRunCases` with one case, `shim_caller_cwd`: `fixture`; for each of `a` and `b`, push the worktree's `HEAD` to its origin's `main` and fetch in the main checkout; then, outside `sut`, run the real shim from inside the canonical worktree with a relative main checkout — `OUT="$(cd "$WORK/a-worktrees/demo" && PATH="$WORK/bin:$PATH" "$SUT" cleanup ../../a demo --harness test --session-token mf-test 2>"$WORK/stderr")"; RC=$?` — and check `shim_caller_cwd: cleanup run from inside a worktree is not HELD by its own cwd` (no `^HELD:` line) and `shim_caller_cwd: a relative main checkout resolves against the caller's directory` (`[ "$RC" -eq 0 ]`, `$WORK/a-worktrees/demo` gone, the record `FINISHED`). <!-- unverified: confirm the pushed fixture reads as merged to cleanup; reproducer 1 also pushes HEAD:spectre/demo -->
  - [ ] **Step 3: Prove each case.** `cd stats && go test ./internal/guard -run 'TestIntegrateChangeShim|TestIntegrateChangeLandCleanup/(indented_fixed_row|padded_orphan)' -count=1` passes. Then, one at a time, revert each fix in place, run the same command, see its case fail, and restore: delete the `cd / || …` line of `scripts/integrate-change.sh` (`shim_caller_cwd` fails, HELD); in `stats/internal/guard/integratechange_land.go`'s fixed-row loop match `strings.Trim(line, " \t")` instead of `line` (`indented_fixed_row` fails); in `stats/internal/guard/integratechange_cleanup.go`'s orphan loop add `o = strings.Trim(o, " \t")` as the loop's first statement (`padded_orphan` fails). `git diff -- scripts stats/internal/guard/integratechange_land.go stats/internal/guard/integratechange_cleanup.go` prints nothing afterwards.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l .` prints nothing; `cd stats && make -s web-dist && go vet ./...` exits 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/integrate_change_test.go`, `stats/internal/guard/integrate_change_land_test.go`
**Tests:** `shim_caller_cwd`, `indented_fixed_row`, `padded_orphan`
**Regression:** `shim_caller_cwd` fails — cleanup through the shim from inside a worktree is HELD by its own cwd again; `indented_fixed_row` fails — an indented report row is recorded as fixed; `padded_orphan` fails — a padded orphan id reaches the remove command.
**Commit:** test(flow-guard): pin integrate-change's three parity fixes with suite cases
**After:** none

**Decision:** parity-fix-suite-cases

- [ ] 4. Refuse a Baseline chain that falls back mid-plan
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `git show 8f1f3eb4:spectre/changes/kan-1019-flow-improvement-port-integrate-change-sh-to-go/tasks.md >"$TMPDIR/k1019-plan.md" && scripts/check-plan-shape.sh "$TMPDIR/k1019-plan.md"` exits 0, though its tasks 3 and 4 share one `measured:` command and both declare `before=1 after=2`.
  - [ ] **Step 2: Write the tests** in `stats/internal/guard/check_plan_shape_test.go` after case 40. Case 41 (F17): a plan carrying the `> **Baseline convention:**` line and two tasks whose Baselines carry `<!-- measured: grep -c x f @ abc -->`, task 1 `before=1 after=2`, task 2 `before=1 after=2` (`**After:** Task 1`) — exit 1, naming `task 2's **Baseline:** before=1 is below task 1's after=2`; `mutation("F17", …)`. Case 42 (F17 absent): the same with task 2 `before=2 after=3` passes; a pair where task 1 is `**Build:** red` with `**Squash-with:** Task 2` and both declare `before=1 after=2` passes; two tasks under different `measured:` commands, the second `before=1` after the first's `after=2`, pass.
  - [ ] **Step 3: Implement** per decision `plan-shape-baseline-chain` in `stats/internal/guard/planshape.go`: `func (p *psPlan) baselineChain() []string`, called in `check()` after `baselineConvention()`. Over `tcfIterTasks(p.lines)` in plan order, skipping a task whose fence never closes, `p.parse(id)`; for a task with a `baseline` and each command in its `baselineMeasured`, compare with the last task seen under that command: when `before` is below its `after`, and neither task names the other in `squashPartners`, report at the task's Baseline line (`psFieldLine`) `task <id>'s **Baseline:** before=<b> is below task <prev>'s after=<a> under the same measured: command — chain-position absolutes chain in plan order; restate before from task <prev>'s after`, gated `p.on("F17")`; then record this task as that command's last. Update the header: seventeen findings, `F1-F17` wherever `F1-F16` is written, and an F17 row stating the rule, the Squash-with exemption and why `<` rather than `!=` (a task between them may add tests to the same source set under another command — kan-916, kan-939). Write the sentence below into `skills/flow/brainstorm-planner.md` right after "A plan whose tasks declare no Baselines states no convention and writes no line.", exactly as quoted.

```markdown verified:applied to a scratch clone at 1420a779; scripts/check-verbatim-moves.sh with this change's verbatim-moves.txt exits 0
Tasks whose Baselines carry the same `measured:` command form one chain in plan order: a task whose `before` is below the previous such task's `after` fails `check-plan-shape.sh` (F17), save the later task of a `**Squash-with:**` pair, which shares its partner's commit.
```

  - [ ] **Step 4: Verify.** `cd stats && gofmt -l .` prints nothing; `cd stats && go test ./internal/guard -run 'TestCheckPlanShape$' -count=1` passes; Step 1's command exits 1 naming task 4; `for f in $(git ls-files '*tasks.md'); do scripts/check-plan-shape.sh "$f"; done 2>&1 | grep -c 'is below task'` prints `0`; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/planshape.go`, `stats/internal/guard/check_plan_shape_test.go`, `skills/flow/brainstorm-planner.md`
**Tests:** Case 41, Case 42
**Regression:** Case 41 fails — a plan re-declaring a plan-time Baseline pair mid-chain passes check-plan-shape again and fails only at the task's commit.
**Commit:** feat(flow-guard): refuse a Baseline chain that falls back mid-plan
**After:** Task 1

**Decision:** plan-shape-baseline-chain

- [ ] 5. Stamp each panel slot's end at its own report time
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `flow record dispatches -change kan-1019-flow-improvement-port-integrate-change-sh-to-go | jq -c '.[] | select(.role=="reviewer") | {id,key,startedAt,endedAt}'` shows rows 7101 and 7102 ending at the same instant; `ls -lT <wt>/.superpowers/sdd/panel-report-0-*.md` shows the bundles' reports written at different times.
  - [ ] **Step 2: Write the prose** per decision `panel-end-report-mtime` in `skills/flow/review-panel.md`, exactly as quoted: in the fenced block under "**Every slot's dispatch is recorded**", the `end` command's last line becomes the two lines of the first block below; after "every `end` is recorded in one call once they all exist (the turn discipline of **4. Execute (SDD + TDD)**, `skills/flow/implement.md`)." insert the sentence of the second block.

```bash verified:applied to a scratch clone at 1420a779; scripts/check-verbatim-moves.sh with this change's verbatim-moves.txt exits 0
  -session-token mf-<literal-token> -outcome completed \
  -ended-at "$(date -u -r <report-file> +%Y-%m-%dT%H:%M:%SZ)"
```

```markdown verified:applied to a scratch clone at 1420a779; scripts/check-verbatim-moves.sh with this change's verbatim-moves.txt exits 0
Each `end` passes its own slot's report-file modification time as `-ended-at`, so a slot that finished early is never stamped with the round's last finish (kan-1019: two slots of one round both read 15m31s).
```

  - [ ] **Step 3: Repair kan-1019's two rows** in the dev store (`flowd`'s default DSN, the dev workspace's `flow` database — never production), keeping each row's start instant from Step 1's output: `cd stats && go run ./cmd/flowd retime-dispatch 7101 <7101's startedAt> "$(date -u -r <wt>/.superpowers/sdd/panel-report-0-principles.md +%Y-%m-%dT%H:%M:%SZ)" ""` and `cd stats && go run ./cmd/flowd retime-dispatch 7102 <7102's startedAt> "$(date -u -r <wt>/.superpowers/sdd/panel-report-0-test-sanity.md +%Y-%m-%dT%H:%M:%SZ)" ""`. Re-run Step 1's `flow record dispatches` read: the two rows now end at their report times and keep their starts.
  - [ ] **Step 4: Verify.** `date -u -r <wt>/.superpowers/sdd/panel-report-0-primary.md +%Y-%m-%dT%H:%M:%SZ` prints one UTC instant `flow record dispatch end -ended-at` accepts; `scripts/check-verbatim-moves.sh`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-stage-mark-calls.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** none
**Regression:** none — prose; Step 3's read-back shows the repaired rows.
**Commit:** fix(flow): stamp each panel slot's end at its own report time
**After:** none

**Decision:** panel-end-report-mtime

- [ ] 6. Run integrate-change's multi-fixture cases in parallel
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `cd stats && go test ./internal/guard -run 'TestIntegrateChangePrepareCommit' -count=1 -v 2>&1 | grep -E -- '^    --- PASS: TestIntegrateChangePrepareCommit/(narrative_format|rerun_pin) '` shows the two cases' times; save `cd stats && go test ./internal/guard -run 'TestIntegrateChangePrepareCommit' -count=1 -v 2>&1 | grep -E -- '--- (PASS|FAIL)' | sed 's#.*/##; s# (.*##' | sort` as `$TMPDIR/labels-before.txt`.
  - [ ] **Step 2: Split** per decision `split-multi-fixture-cases` in `stats/internal/guard/integrate_change_prepare_test.go`: a Go string constant `icNarrativePasses` holding the `narrative_passes` function and one `icRerunPinCase` holding `rerun_pin_case`, each prepended to the scripts that call it; `narrative_format` keeps its stop case alone; `narrative_format_in_format`, `narrative_format_prose_mention`, `narrative_format_trailing_space`, `narrative_format_deferred` and `narrative_format_backticked_sha` each make one `narrative_passes` call with today's label and line; `rerun_pin` keeps `rerun_pin_case "rerun_pin: "` and `rerun_pin_reclone` makes the `reclone` call. Every check label is unchanged.
  - [ ] **Step 3: Verify.** `cd stats && gofmt -l .` prints nothing; Step 1's label list, regenerated after the split, equals `$TMPDIR/labels-before.txt` once the new case names are removed from it — no check label lost or renamed; `narrative_format` and `rerun_pin` no longer lead Step 1's timing read. Commit.
    <!-- predicted: Step 3's timing read shows the longest remaining case well below narrative_format's pre-split time -->

**Build:** green
**Files:** `stats/internal/guard/integrate_change_prepare_test.go`
**Tests:** none — repairs existing tests: `narrative_format`, `rerun_pin`
**Regression:** none — the same checks run; only their grouping into parallel cases changes.
**Commit:** perf(flow-guard): run integrate-change's multi-fixture cases in parallel
**After:** Task 1

**Decision:** split-multi-fixture-cases
