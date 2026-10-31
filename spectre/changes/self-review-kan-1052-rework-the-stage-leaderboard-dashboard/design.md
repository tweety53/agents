## Context

Self-review fix change for `kan-1052-rework-the-stage-leaderboard-dashboard`, planned on that
change's own branch `spectre/kan-1052-rework-the-stage-leaderboard-dashboard` at agents `fab4520a`
(the fix-base). Seven non-`big` findings, as `proposal.md` names them, each carried by one task in
`tasks.md`: F1 (Task 1), F2 (Task 2), F5 (Task 3), F3 (Task 4), F4 (Task 5), F6 (Task 6), F16
(Task 7). None was disproved. Every question was auto-resolved to its recommended option
(**Auto-resolution**, `skills/flow-contracts/operator-prompts-auto-resolution.md`). The brainstorm
design, with each finding's trace and the history check, is
`.superpowers/sdd/2026-10-31-self-review-kan-1052-rework-the-stage-leaderboard-dashboard-design.md`
in this worktree.

Task 7 also lands in the spectre repository, which this change holds no worktree in: the parent
creates one off spectre's default branch before Task 7's dispatch, and it lands by spectre's own
`## default landing route` (`merge and push`, `/Users/tweety53/Projects/spectre/.flow/project.md`).
`flowd` on 4173, the `flow-postgres` container and the `flow` database are never stopped or dropped.

## Decisions

### No planning agents for this pass

**ID:** no-planning-agents
**Status:** active
**Chosen:** the planner traced each finding itself; no Explore or Plan agent was dispatched — every finding names its files and blast radius, and each trace took a handful of reads.
**Considered:** 1–3 Explore agents — they would re-read the same named files and return conclusions already in hand.

### ui paths covers the model-policy file flowd embeds

**ID:** ui-paths-embedded-policy
**Status:** active
**Chosen:** add `` `skills/flow-contracts/model-policy.md` `` to `.flow/project.md`'s `ui paths` cell, and a guard test that collects every `cp ../<src>` source of `stats/Makefile` and requires `check-visual-trigger` against the real `.flow/project.md` to match each, so a future embed copy fails the test until `ui paths` names it.
**Considered:** a `## lint` line piping the one path into `check-visual-trigger.sh` — catches only the file named today; `stats/internal/modelpolicy/embedded/**` — that copy is gitignored, so no diff ever names it.

### A visual list runs the verify row alone

**ID:** visual-list
**Status:** active
**Chosen:** `run-verify-list.sh <worktree> visual` runs only the `## visual verification` `verify` row, under exactly the conditions `rvlVisualVerify` already applies to a `test` list's tail, and prints one not-run line otherwise. **5. Verify** of `/flow-fast` (canonical) and step 4 of `/flow-interactive` run it; an intended look change is recaptured with the `capture` row and committed before landing.
**Considered:** prose telling both skills to run the `verify` row by hand — re-states the trigger, decision and checkout logic `rvlVisualVerify` owns, and was the gap the run fell through; `run-verify-list.sh <worktree> test` — runs the whole `## test` list, which both skills deliberately scope.

### The implementer key names every task of the group

**ID:** implementer-key-ids
**Status:** active
**Chosen:** `task-<ids>-implementer`, ids `+`-joined in plan order, in every implementer key of `skills/flow/implement.md` and the two code comments citing the shape; the per-task `-recommit-<k>` key stays `task-<n>`. No test: the hook (`hooks/enforce-agent-baseline.py`) already enforces the form, and `records.taskOf` and the store's inline check already parse it — the change aligns the documentation with enforced behaviour.
**Considered:** a guard pinning the prose to the hook — a prose-matching test with nothing to catch once the text is right.

### run-capped case (a) proves the kill by a marker, not a clock

**ID:** run-capped-completion-marker
**Status:** active
**Chosen:** case (a) runs the allocator followed by a write of a completion marker in the same `python3 -c`, and asserts status 137, the `CAPPED:` line and the marker's absence; the `date` timing and the `-le 6` bound go. The proof: on a scratch copy, a `run-capped.sh` that never kills fails case (a), and one whose first sample is delayed past 6 s fails the old assertion and passes the new.
**Considered:** widening the bound — forbidden (`rules/fix-determinism-at-the-source.mdc`); timing only the kill — still a clock under load.

### The Go suites run on a share of the cores

**ID:** go-suites-core-share
**Status:** active
**Chosen:** the `guard-tests` and `stats-go` lines of `## test` run under `GOMAXPROCS=$(( ($(getconf _NPROCESSORS_ONLN) + 2) / 3 ))`, which bounds the go command's build parallelism, its default `-p`, and each test binary's threads and `-parallel` — 4 + 4 beside vitest's 2 workers on this 10-core machine. The vitest half of the finding's remedy already exists (`maxWorkers: 2`, decision `threads-two-workers`, kan-1044). The task opens on measuring the timed-out SPA test alone and under the two `-race` suites' heaviest packages, and repeats the loaded measurement with the share; a guard test pins the prefix on both lines. Task steps never run a whole suite (`check-plan-shape.sh`'s plan-suite-steps), so the load is those packages, not the full list.
**Considered:** `go test -p` alone — bounds packages, not each `-race` binary's threads, and does not reach `test-go-guards.sh`'s single-package run; `nice` on the Go lines — reorders the starvation rather than bounding it; raising the vitest timeout — forbidden.

### The mutation brief carries the node_modules rule

**ID:** mutation-brief-node-modules
**Status:** active
**Chosen:** one sentence at the end of **The mutation-testing brief** (`skills/flow/review-panel.md`), which `render-slot-prompt.sh` renders into every mutation pass, and a render test that requires it. **The throwaway worktree**'s parent-facing sentence stays.
**Considered:** the MUTATION ENTRY CONTEXT paragraph — typed by the parent per dispatch, so nothing mechanical carries it.

### The discriminating-fixture rule lives in the generic planning prompts

**ID:** discriminating-fixture-rule
**Status:** active
**Chosen:** one `>` rule in section D of `skills/spectre-planning/SKILL.md`, citing nothing flow-only, re-copied byte for byte into `/Users/tweety53/Projects/spectre/.claude/skills/spectre-planning/SKILL.md` with `UPSTREAM.md`'s Landed line naming the agents ticket rather than a sha on this branch. No test: no mechanical check can judge whether a fixture discriminates.
**Considered:** `skills/flow/brainstorm-planner.md` — the rule binds any planner, not only flow's.

### The parent creates spectre's worktree

**ID:** spectre-worktree-by-parent
**Status:** active
**Chosen:** before Task 7's dispatch the parent creates `/Users/tweety53/Projects/spectre-worktrees/self-review-kan-1052-rework-the-stage-leaderboard-dashboard`, branch `self-review-kan-1052-rework-the-stage-leaderboard-dashboard`, off spectre's `origin/main`; after the agents branch lands, spectre's branch lands by `merge and push` and its worktree is removed.
**Considered:** a task step creating it — no task creates a worktree (`skills/flow/brainstorm-planner.md`).

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 2 adds a list mode to `run-verify-list`, which every flow run calls, and Task 3 changes the `## test` list every verify runs.
**Considered:** `flow-low` — most diffs are small, but a wrong visual-list condition or core share silently skips or slows every run's verify.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** three groups in series: `-fix-g1` Tasks 1–3 on `opus` — Tasks 1 and 3 share `.flow/project.md`, Tasks 2 and 3 share `run_verify_list_test.go`, and Task 3 runs the long measurements; `-fix-g2` Tasks 4–6 on `opus` — prose plus a test harness and a render test; `-fix-g3` Task 7 on `sonnet` — one literal rule the plan spells out and a byte-for-byte copy, nothing it can break.
**Considered:** one fixer for all seven — Task 3's two full concurrent suite runs would sit in the same context as every other fix; one per task — seven dispatches, three sharing files.

## Open questions

## Live check

None — no running service or persistent state changes: the tasks edit guards, a test harness,
`## test` command lines and prose. Task 3's before/after measurements are its own steps, and the
`## Verification` runs below exercise the rest.

## Recommended approach

Fix each defect where it starts: `ui paths` names the embedded file, pinned by a test that derives
the embed sources from the Makefile; `run-verify-list` gains a `visual` list reusing
`rvlVisualVerify`, which the two lightweight skills call; the key shape is corrected where it is
documented; case (a) asserts the kill by a marker; the Go suites get a `GOMAXPROCS` share; the
mutation brief and the planning prompts each gain one rule. Rejected alternatives are under each
decision ID above.

## Critical files

- `.flow/project.md` — `ui paths` (Task 1) and the `## test` list (Task 3).
- `stats/internal/guard/check_visual_trigger_test.go` — the embed-sources test (Task 1).
- `stats/internal/guard/runverifylist.go`, `run_verify_list_test.go`, `scripts/run-verify-list.sh` — the `visual` list (Task 2) and the core-share test (Task 3).
- `skills/flow-fast/SKILL.md`, `skills/flow-interactive/SKILL.md` — run the `visual` list (Task 2).
- `skills/flow/implement.md`, `stats/internal/records/render.go`, `stats/internal/records/render_test.go`, `stats/internal/store/records.go` — the key shape (Task 4).
- `scripts/test-run-capped.sh` — case (a) (Task 5).
- `skills/flow/review-panel.md`, `stats/internal/guard/render_slot_prompt_test.go` — the mutation brief (Task 6).
- `skills/spectre-planning/SKILL.md`; spectre's `.claude/skills/spectre-planning/SKILL.md` and `UPSTREAM.md` (Task 7).

## Existing code to reuse

- `rvlVisualVerify` and the visual tail of `runVerifyList` (`stats/internal/guard/runverifylist.go`).
- `Registry["check-visual-trigger"]` and the `vtRun` harness (`check_visual_trigger_test.go`).
- `TestRunVerifyListRunsTheVisualVerify`'s git fixture (`run_verify_list_test.go`).
- `TestRenderSlotPromptRendersMutationPass`'s render fixture (`rspSkillDir`, `rspNewFx`).
- `scripts/break-and-prove.sh` for the mutation proofs.

## Verification

Beyond each task's verify step: `scripts/run-verify-list.sh <worktree> lint` and
`scripts/run-verify-list.sh <worktree> test` exit 0 in the agents worktree, the latter now
including Task 3's bounded Go suites and, since `.flow/project.md` changed, the visual verify;
`scripts/run-verify-list.sh <worktree> visual` on this branch runs the `verify` row
(its diff touches `stats/web/src/`); in the spectre worktree, `cmp` of the two
`spectre-planning/SKILL.md` copies is silent and `go test ./...` passes.
