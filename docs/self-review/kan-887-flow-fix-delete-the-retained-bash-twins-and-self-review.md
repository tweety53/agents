# Self-review report for kan-887-flow-fix-delete-the-retained-bash-twins-and

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle
**Rating:** 4 — good

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `check-panel-findings-closed` counted a panel re-run only when closed `completed`, though `flow record dispatch end` accepts `clean` and `implement.md` directs it; now accepts `clean` — fixed: 166354fc
- **[flow-fix]** `Case <N>` labels rejected a letter suffix (`13b`) in both `CASE_LABEL_RE` and `tcfCaseLabelRE`, though 15 Go test files use them; both accept `[0-9]+[a-z]?` — fixed: e13bd943
- **[flow-fix]** `pipeline-fix.sh verify` never selected the Go-tree audit harnesses (`test-flow-addr-declaration.sh`, `test-git-config-pins.sh`) for a touched Go file, so main landed red twice; now selected — fixed: 99e908ba
- **[flow-fix]** The `pipeline.md` sentence edit needed a `verbatim-moves.txt` acknowledgement — the guard working as designed — noted
- **[flow-fix]** Inline panel-fix rows recorded 4 s and 9 s windows, suggesting a late `begin`; no mechanism can detect it, the true begins are unrecoverable, and `implement.md` already states the rule — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** The 5094-line panel diff is ~2904 lines of deleted-file bodies; dropping them would lose the test-sanity slot's audit evidence, which produced F1 — noted
- **[flow-cost]** Inline implementer rows carry 0.35–1.9M cache-read tokens each — the cost of the inline mode the decision chose — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Diffing each bash harness's `ok:` labels against the Go subtests, then reading uncovered cases, closed six coverage gaps before deletion — keep `harness-case-audit` as a plan decision on harness deletions — noted
- **[flow-improvement]** `change-plan.sh`'s normative 109-line header moved into `changeplan.go`'s header instead of being deleted with its file — noted
- **[flow-improvement]** The parent's full-suite run caught two red-on-base harnesses and fixed them in-run — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** Scripting the harness label diff: bash `ok:` labels and Go subtest names do not align and each gap needed reading; one pending consumer — noted

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

- **[flow-speed]** Verify re-ran the whole `## test` list ~14 minutes after a green full-suite run when the only delta was a test rename; `run-verify-list.sh` now records the green HEAD and skips commands whose `# inputs:` and `.flow/project.md` are untouched since — fixed: 30b34f1c
- **[flow-speed]** Tasks 7, 8 and 10 each ran `run-guard-tests.sh` for one absent line; 8 and 10 took 29 s and 23 s with the same step, so it is not task 7's 6m20s — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review case — F1, case 20's name contradicts its exit-1 assertion; operator chose not to mine — noted
  - Location: `stats/internal/guard/check_panel_reproducers_test.go:156`; bad 29fe7c9a, fixed 20c5fd5a
  - Files: `/Users/tweety53/Agents/flow/state/agents-a740d89c/flow-eval/kan-887-flow-fix-delete-the-retained-bash-twins-and/check_panel_reproducers_test.bad.go`, `…/check_panel_reproducers_test.fixed.go`
  - Why: name and assertion diverge after a mechanical rewrite; only checking the exit code catches it
- **[flow-eval]** impl case — Task 6, audit the aside harness before deleting it; operator chose not to mine — noted
  - Tests: `TestAsidePlanningArtifactsState`
  - Code: `/Users/tweety53/Agents/flow/state/agents-a740d89c/flow-eval/kan-887-flow-fix-delete-the-retained-bash-twins-and/aside_planning_artifacts_test.go:173`
  - Why: the audit finds a side-effect coverage gap the output-only parity test missed
