## Context

The fix plan for the self-review pass of `kan-1034-spectre-optional-bundled-planning-prompts-that`
(`.superpowers/sdd/self-review-pass.md`, angles 1, 2 and 4): every finding classified fixable and
not `big` is carried here by exactly one task. Excluded: the in-run pipeline fix `36eb356c`
(already landed), every `noted` item, and the out-of-scope spectre test comment (fixed by the run's
own fix round, its Task 3). The fixes ride the change's own branch
`spectre/kan-1034-spectre-optional-bundled-planning-prompts-that` in both repositories — the agents
worktree for Tasks 1–8, the spectre worktree for Task 9. The change itself is already archived on
that branch and is not touched.

Reproduced at plan time on `bef019f7`: `check-plan-shape.sh` exits 1 on the bold `**none**` repair
form and 0 on a Baseline whose `measured:` comment carries no ` @ <ref>`; `check-panel-findings-closed.sh`
prints `FINDINGS-CLOSED` on kan-1034 while F7's recorded reproducer
`.superpowers/sdd/reproducers/r1-F1-1.sh` is absent. The two placeholder reports the pass asked for
(`panel-report-0-mutation.md`, `panel-report-0-failure-modes.md`) already exist, written by the
parent before this plan.

## Decisions

### First review's effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — four of the nine tasks change guard logic other runs depend on (the panel close, plan shape, relocation comparison, Markdown integrity), so the first review reads behaviour, not wording.
**Considered:** flow-low — fits the prose and config tasks, but under-reads the guard changes.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** three opus fixer dispatches in series — group 1: Task 1; group 2: Tasks 2, 3, 4, 5; group 3: Tasks 6, 7, 8, 9 — the panel guard alone (the widest fixture rework), the other Go guards together, the prose and config last.
**Considered:** one fixer for all nine — a single context carries four unrelated Go guards plus prose and a second repository; nine fixers — dispatch overhead with nothing gained for one-sentence tasks.

### Report presence covers every completed reviewer dispatch

**ID:** report-presence-every-dispatch
**Status:** active
**Chosen:** for every dispatch row with role `reviewer`, outcome `completed` or `clean` and a `panel-<round>-` key, each `+`-component of its slot needs a non-empty `panel-report-<round>-<component>.md`; the `no verbatim report captured` placeholder is accepted, as `review-panel.md`'s **Recording findings** prescribes it. A file the re-run coverage line already names is not named again.
**Considered:** re-runs only (today's `covers()`) — the round-0 slot that raised F7 left no report and no guard noticed; rejecting the placeholder here — it is the parent's sanctioned record of a lost report.

### Reproducer existence for fixed findings, resolved across the change's worktrees

**ID:** fixed-reproducer-resolves
**Status:** active
**Chosen:** `check-panel-findings-closed` checks every `fixed` finding whose reproducer is runnable (not the `none` forms, not empty): its path token — the first space- or tab-separated field of the first line, as `panelexitcontract.go` derives it — resolves in the worktree, else in one of the state record's worktrees (`flow state get`, `pcWorktreeKeys`, `resolvedRecordedTrees`), read only when the worktree lacks the path.
**Considered:** extending `check-panel-reproducer-exit-contract` — it runs per fix round, not at the panel's close, and skips non-open findings by design; checking every status — an open finding is that guard's, a withdrawn one claims nothing; the worktree alone — a cross-repo peer's reproducer lives in its own worktree (KAN-658) and would fail falsely.

### Repairing kan-1034's record

**ID:** kan-1034-repair
**Status:** active
**Chosen:** the parent recreates F7's reproducer at `.superpowers/sdd/reproducers/r1-F1-1.sh` from F7's demonstration (self-review-pass.md angle 7), as the new recording rule prescribes; the file is gitignored, so it is a parent action, not a task.
**Considered:** re-recording F7's reproducer as `none — <reason>` — `check-panel-reproducers` refuses the exemption on an Important finding; leaving it — the new guard would then refuse this change's own panel record.

### Symlinked scoped paths are skipped, not refused

**ID:** relocation-skip-symlink
**Status:** active
**Chosen:** a scoped path that is a symlink in the worktree or a mode-120000 entry at `<merge-base>` contributes no passages and prints `generate-relocation-comparison: skipped symlink <path>` on stderr; the comparison continues to exit 0. Neither side ever resolves a link, which is all the refusal protected.
**Considered:** following links — one side would resolve what the other does not; keeping the refusal — `check-guard-symlinks` rule 2 requires such symlinks in a legitimate plan's `**Files:**`, so the panel lost the comparison.

### An unreadable measured: comment is a plan-shape violation

**ID:** measured-comment-f20
**Status:** active
**Chosen:** plan-shape rule F20 refuses any `<!-- measured: … -->` comment in a task's joined Baseline value that `tcfMeasuredRE` does not match, naming the `@ <ref>` form.
**Considered:** relaxing `tcfMeasuredRE` to read a comment with no ` @ ` — the `@ <ref>` is the provenance annotation plan-provenance requires, so a comment without it is the defect, not the parser.

### Bold none opens the repair form

**ID:** bold-none-repair
**Status:** active
**Chosen:** `tcfRepairsOpenRE` accepts `[*_]*` after `none\b`, matching the `[*_]*` it already accepts before `none`.
**Considered:** rewording `brainstorm-planner.md` to forbid bold — the planner prompt already tells planners bold opens the field, and `tcfNoneOpenRE` agrees.

### Markdown layout signals live in check-markdown-integrity's scope

**ID:** markdown-layout-signals
**Status:** active
**Chosen:** signals 5 (missing final newline) and 6 (a blank line directly after a blank line, outside a fence block) in `check-markdown-integrity`, scanning its existing scope (`skills/**/*.md`, `rules/*.mdc`); the three sites they flag are fixed in the same task.
**Considered:** widening the scope to `commands-claude/` as the pass measured — no hit there today and nobody asked for the wider scope.

### spectre's gofmt line fails on a hit

**ID:** spectre-gofmt-fails
**Status:** active
**Chosen:** `! gofmt -l . | grep .` — prints each unformatted file and exits 1 when there is one; `run-verify-list.sh` runs every line through `bash -c`.
**Considered:** `gofmt -l .` verbatim from spectre's `AGENTS.md` — it exits 0 whatever it lists, so the line could never fail.

### Planning rules live in the flow-only planner file

**ID:** rules-in-brainstorm-planner
**Status:** active
**Chosen:** the relocation-consumer and branch-sha rules go into `skills/flow/brainstorm-planner.md`, beside the `**Relocation:**` header rule and the plan-time premise rules; both rest on flow machinery (the relocation comparison, integrate's reshape).
**Considered:** `skills/spectre-planning/SKILL.md` — generic and vendored by spectre, so every edit there also re-vendors spectre's copy, and both rules cite flow-only stages.

## Open questions

## Recommended approach

Fix each defect at the guard or prompt that owns it, one task per finding (Task 6 carries the two
`brainstorm-planner.md` rules): extend the existing Go guards in place — `panelfindingsclosed.go`,
`planparse.go`, `planshape.go`, `relocationcomparison.go`, `markdownintegrity.go` — each with a
test that fails without the fix, update each shim header whose contract changes, and add the prose
rules with their sentences acknowledged in `verbatim-moves.txt`. Rejected alternatives are recorded
per decision: `fixed-reproducer-resolves`, `relocation-skip-symlink`, `measured-comment-f20`,
`markdown-layout-signals`, `rules-in-brainstorm-planner`.

## Critical files

- `stats/internal/guard/panelfindingsclosed.go` — the panel close guard (Task 1).
- `stats/internal/guard/check_panel_findings_closed_test.go` — its tests (Task 1).
- `scripts/check-panel-findings-closed.sh` — the shim whose header is the exit-code contract (Task 1).
- `skills/flow/review-panel.md` — recording rule and the close guard's exit-1 classes (Task 1).
- `stats/internal/guard/planparse.go` — `tcfRepairsOpenRE` (Task 2), `tcfMeasuredRE` (read by Task 3).
- `stats/internal/guard/planshape.go` — rule list and per-task checks, new F20 (Task 3).
- `stats/internal/guard/check_plan_shape_test.go` — Cases 44 and 45 (Tasks 2, 3).
- `stats/internal/guard/relocationcomparison.go` — symlink handling and header (Task 4).
- `stats/internal/guard/generate_relocation_comparison_test.go` — symlink cases (Task 4).
- `stats/internal/guard/markdownintegrity.go` — the two new signals (Task 5).
- `stats/internal/guard/check_markdown_integrity_test.go` — their tests (Task 5).
- `scripts/check-markdown-integrity.sh` — shim header naming the signals (Task 5).
- `skills/flow-contracts/workspace-isolation.md`, `skills/flow/SKILL-rationale.md` — the three flagged sites (Task 5).
- `skills/flow/brainstorm-planner.md` — two planning rules (Task 6).
- `skills/flow/brainstorm.md` — planning-dispatch recording (Task 7).
- `CLAUDE.md`, `AGENTS.md` — the `skills/spectre-planning/` index row (Task 8).
- spectre repository `.flow/project.md` — `## lint`, `## test` (Task 9).
- `spectre/changes/self-review-kan-1034-spectre-optional-bundled-planning-prompts-that/verbatim-moves.txt` — the six acknowledged new sentences, committed with this plan.

## Existing code to reuse

- `pcFlow`, `pcWorktreeKeys`, `resolvedRecordedTrees`, `pcExists` (`stats/internal/guard/panelexitcontract.go`) — the state read and cross-repo tree resolution for Task 1.
- `cfcReportWritten`, `cfcPanelRoundRE`, `cfcBinDispatches` (`panelfindingsclosed.go`, its test) — report-file reading and the stub `flow` for Task 1.
- `tcfMeasuredRE`, `tcfBodyFields`, `psFieldLine`, `p.on`/`psMutant` (`planparse.go`, `planshape.go`, `check_plan_shape_test.go`) — Task 3's rule and its mutation proof.
- `grcGitShow`'s mode-120000 detection (`relocationcomparison.go`) — Task 4.
- `mdiParseBlocks`'s fence blocks and `mdiFinding` (`markdownintegrity.go`) — Task 5.

## Verification

- `cd stats && make -s web-dist && go vet ./... && gofmt -l .` clean, and `go test ./internal/guard/ -count=1` green after the last task.
- `scripts/check-plan-shape.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-references.sh`, `scripts/check-verbatim-moves.sh` exit 0 on the branch.
- `scripts/check-panel-findings-closed.sh <agents worktree> kan-1034-spectre-optional-bundled-planning-prompts-that`: exit 1 naming `r1-F1-1.sh` until the parent recreates F7's reproducer, `FINDINGS-CLOSED` after.
- `scripts/project-get.sh <spectre worktree> lint` and `… test` exit 0, and `scripts/run-verify-list.sh <spectre worktree> lint` and `… test` exit 0.

## Live check

None — no task changes a running service or stored state; the panel guard only reads the store, and its run against kan-1034's real records is the third `## Verification` item.
