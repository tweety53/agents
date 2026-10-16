# kan-966-flow-mutation-proofs-bound-a-mutants-run-and — self-review

**Deferred:** reasoning pass run on opus from the in-run bundle
**Rating:** 3 — fine

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** The intentional usage-string change shipped with a stale pinned expectation, so the touched package's suite was red until the pin was updated in the panel fix commit — fixed: 6594368c
- **[flow-fix]** check-ask-silence failed identically on the untouched base commit; fixed in-run as a sweep commit and never offered — noted
- **[flow-fix]** The first clean-check in the timeout test false-failed on the run's own untracked fixture harness; resolved in-run by checking the touched file only — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Per-dispatch tokens record "not measured", so a run's cost is unmeasurable and no cost reduction is verifiable — filed: KAN-986
- **[flow-cost]** One stale pin consumed three reviewer slots plus an opus/low re-run — filed: KAN-985

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Compact roster at diff-size 259 under cap, the base move auto-rebased, per-task commit and push, and an opus/low re-run closed all six panel findings — keep small-class inline runs on exactly this shape — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** Automate panel-finding dedup by location and root cause before the fix dispatch — filed: KAN-985
- **[flow-automation]** Automate base-vs-branch lint attribution (re-run a failing check at merge-base and label it pre-existing) — declined

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** Move per-dispatch token capture into the Go store at dispatch close so cost is queryable, not re-derived per render — filed: KAN-986

## What can be sped up — `flow-speed`

- **[flow-speed]** The serial fix and re-review loop took about five minutes, inflated by the triplicated pin finding; dedup cuts one fix-verify cycle — filed: KAN-985
- **[flow-speed]** The fixture timeout raised from 1s to 3s trades about two seconds per run against a load-dependent baseline flake, accepted and fixed in the panel fix commit — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** impl Task 1 — bound every run in mutate-and-verify and count a mutated timeout as killed; operator chose to mine nothing — noted
  - task: `.superpowers/sdd/kan-966-flow-mutation-proofs-bound-a-mutants-run-and/tasks.md` Task 1, commit bb17ee45
  - code: `stats/internal/guard/mutateandverify.go:425` (`mvTimeout`/`mvRunBounded`)
  - why: self-contained verdict-semantics task with deterministic sleep-fixture tests and explicit rejected alternatives
- **[flow-eval]** impl Task 2 — run a project test command after `--` through mutate-and-verify; operator chose to mine nothing — noted
  - task: `.superpowers/sdd/kan-966-flow-mutation-proofs-bound-a-mutants-run-and/tasks.md` Task 2, commit 96b7169c
  - code: `stats/internal/guard/mutateandverify.go:502` (`mvRunCommand`)
  - why: argv-vs-shell-string design choice plus an exit-code verdict
- **[flow-eval]** review F2 — Important, `stats/internal/guard/mutateandverify.go:521`: a project command failing on both passes was reported caught when its exit code merely changed; fixed at `stats/internal/guard/mutateandverify.go:523` in 6594368c; operator chose to mine nothing — noted
  - bad code: `return 0, []string{fmt.Sprintf("exit-%d", code)}, true, false`
  - fixed code: `return 0, []string{"exit-nonzero"}, true, false`
  - why: a subtle set-difference verdict bug, invisible without tracing baseline/mutated name equality
- **[flow-eval]** review F1+F3+F4 — Important ×3 duplicate slots on the stale usage pin at `stats/internal/guard/mutate_and_verify_test.go:895`, fixed in 6594368c; operator chose to mine nothing — noted
  - why: tests whether a reviewer catches a self-inflicted suite-red from an intentional output change; also the duplicate-reporting specimen
- **[flow-eval]** review F5+F6 — Minor test gaps and the 1s fixture flake, both fixed in 6594368c (`stats/internal/guard/mutate_and_verify_test.go:577`); operator chose to mine nothing — noted
  - why: a Minor-grade calibration case
