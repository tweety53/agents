# Self-review report for kan-1060-flow-cost-visual-verify-now-costs-more-than

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** In-run pipeline fix: the loop hook's fixed 10 s `record_incident` timeout killed the incident stub under `-race`; hook tests now lift its bounds — fixed: 3deff9fa
- **[flow-fix]** In-run pipeline fix: hook tests inherited `FLOW_RECORDS_ADDR` and read the live flowd's experiment overrides; `runHook` pins both store addresses dead — fixed: 81ca7e1b
- **[flow-fix]** In-run pipeline fix: main's visual suite was red after the state-board perf commits (Project chips, open poll, `ui-test-up` prerequisite, stale baselines) — fixed: dbb6b4f1
- **[flow-fix]** Loop landings ran no visual suite when their diff touched `ui paths`, so 084f060f landed a run-view change with stale baselines; `pipeline-fix verify` now runs the `## visual verification` verify row on a UI-path diff — fixed: 723719ee
- **[flow-fix]** A gated reviewer's on-top fix commit was never field-checked, so an undeclared test file surfaced only at inline verify (two blocked verify rows, ~21 min); the contract now runs `check-task-commit-fields.sh` on it before the re-review — fixed: 95023d6f
- **[flow-fix]** zsh's `:s` modifier mangled `--app $W:<root>=<url>` and the usage error named no cause; it now names the `"${W}:…"` spelling — fixed: 5b3df651
- **[flow-fix]** The protect-main-checkout hook refused writes in the live-check scratch clone sitting on the change branch; correct under its `origin/HEAD` rule, and a local branch cleared it in one step — noted
- **[flow-fix]** The design's live-check token lacked a UTC stamp; `prepare` warns on the refused stage begin by design, the same pattern as `integrate-change` — noted
- **[flow-fix]** Verify row `verify-3` recorded no HEAD; it was written by the installed pre-change script, a transitional gap this change's own runner closes — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Planning at xhigh read ~81M cache tokens against ~30M for every implementer; plan-4 alone was 51M and 50 min — noted
- **[flow-cost]** Inline verify ran 8 times, driven by the three fixed findings above and four base moves — noted
- **[flow-cost]** The 33.4% Main-share figure covers the 30 days before the change, so it proves the columns work, not the saving — noted
- **[flow-cost]** Two implementer dispatches aborted after 45 s; negligible — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** The scripted prepare → one verifier → close pass ran on the real gate three times, each verifier pass taking 1–3 min — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** Nothing beyond the zsh `--app` hint, fixed under flow-fix — noted

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** The stage file stays a file, by design decision; nothing else to move — noted

## What can be sped up — `flow-speed`

- **[flow-speed]** Each of the four base moves forced a fresh verify, plus an unscoped closing visual run after a clean scoped one (decision `full-run-stays`) — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review case — F1 (Important, failure-modes): `die` exited 2 without ending the stage `markBegin` opened, not chosen for mining — noted
  - finding: stats/internal/guard/visualverifypass.go:80, "`die` exits 2 without ending a stage `markBegin` opened"
  - fixed code: commit 52b1be88 (`die` ends the open stage `stopped` and begun rows `aborted`)
  - why: a classic leaked-resource-on-error-path defect a first-round failure-modes slot should catch
- **[flow-eval]** impl case — Task 13's main-session cost rule, not chosen for mining — noted
  - task: Task 13 of the archived tasks.md (Main session = run cost minus summed dispatch cost, floored at 0, NULL when a dispatch lacks cost)
  - code: /Users/tweety53/Projects/agents/stats/internal/store/aggregate.go `GREATEST(cost - dispatch_sum, 0)`
  - why: a small SQL rule with a NULL-propagation edge the store tests pin

## Instruction gaps that cost review-panel or visual-verify fix rounds — `flow-prompt-tuning`

- **[flow-prompt-tuning]** None picked: F1–F5 came in round 0 from the slots built to catch them, and visual verify needed no fix rounds — noted
