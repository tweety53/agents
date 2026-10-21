## Context

Found by kan-887's self-review pass (four non-`big` findings); planned on the change's own branch
`spectre/kan-887-flow-fix-delete-the-retained-bash-twins-and` in its worktree, fix-base `b9e7118b`.
Brainstorm design: `.superpowers/sdd/2026-10-21-self-review-kan-887-flow-fix-delete-the-retained-bash-twins-and-design.md`
in this worktree. Every question was auto-resolved to its recommended option. Each fix was applied
and checked in this worktree at `b9e7118b`; each task carries its diff.

**F-A — a re-run closed `clean` is invisible to `check-panel-findings-closed`.**
`stats/internal/guard/panelfindingsclosed.go`'s `covers` skips every dispatch whose outcome is not
exactly `completed`. `review-panel.md` closes a slot `completed`, but `flow record dispatch end`
accepts `clean` (`stats/cmd/flow/record.go`'s `recordOutcomes`), `skills/flow/implement.md` tells a
gated reviewer to close `clean`/`fix`, and the dev store holds 50 `panel-<round>-…` reviewer rows
closed `clean`.
<!-- measured: docker exec flow-postgres psql -U flow -d flow -Atc "select role, outcome, count(*) from dispatches where dispatch_key ~ '^panel-[0-9]+-' group by 1,2" → reviewer|clean|50 @ dev store 2026-10-21 (live read, not re-runnable at a ref) -->
A finding whose only later re-run of its slot closed `clean` is reported "fixed with no clean
re-run". Nothing to repair: the guard writes nothing, so a wrong verdict left no row behind.

**F-B — a lettered `Case <N>` label parses as nothing.** `CASE_LABEL_RE`
(`scripts/check-task-commit-fields.py`, imported by `check-plan-shape.py`) and `tcfCaseLabelRE`
(`stats/internal/guard/taskcommitfields.go`, also read by `planparse.go`'s `tcfParseTestSpecs`) are
`Case\s+(\d+)\b`: on `Case 13b` no digit run ends on a word boundary, so the field carries no label;
`check-plan-shape` then reports F6 "declares nothing". Fifteen Go test files name cases `"case 13b:"`.
<!-- measured: git grep -lE '"case [0-9]+[a-z]:' -- 'stats/**_test.go' | wc -l → 15 @ b9e7118b -->

**F-C — `pipeline-fix verify` skips the Go-tree audits.** `pfChecks`
(`stats/internal/guard/pipelinefix.go`) adds `go test` of each touched package and the harness named
after a touched script, so `scripts/test-flow-addr-declaration.sh` (audits `stats/cmd/flow`) and
`scripts/test-git-config-pins.sh` (audits `stats/internal/guard`) never run for a Go-only change.
While planning, `test-git-config-pins.sh` itself caught an unpinned `git status --porcelain` in this
plan's own first F-D draft — the class this fix keeps from landing.

**F-D — verify re-runs a list that already passed.** `skills/flow/implement.md`'s full-suite run
and `flow.verify`'s **Inline verify** (`skills/flow/verify-and-handoff.md`) each run the whole
`## test` list; nothing links the second to the first, so a delta of one test rename re-runs every
command. This project's Go tests and guard harnesses read files across the whole repository
(`stats/internal/guard`'s tests open `../../../skills/…`), so only its SPA line can be scoped.

Reproduced at `b9e7118b`: each task's new test fails with its fix reverted and passes with it;
`go test ./internal/guard/ -count=1`, `go vet ./...`, `gofmt -l .`, `scripts/test-run-verify-list.sh`,
`scripts/test-check-plan-shape.sh`, `scripts/test-git-config-pins.sh` and
`scripts/test-flow-addr-declaration.sh` pass with all four applied, and the `## lint` list's only red
is `check-verbatim-moves.sh`, whose five sentences `verbatim-moves.txt` beside this file acknowledges.
<!-- measured: go test ./internal/guard/ -run <each task's test> -count=1 with and without each fix, and the listed suites with all four diffs applied @ b9e7118b -->

## Live check

No runtime to verify against: the four fixes change guards and a script that run per invocation;
no daemon, store or scheduler holds their state, and the green sha lives in the worktree's git dir.

## Decisions

### Count a re-run closed `clean`

**ID:** clean-rerun-counts
**Status:** active
**Chosen:** a slot re-run qualifies when its outcome is `completed` or `clean` — the two outcomes a contract closes a finished, finding-free reviewer pass with.
**Considered:** also `fix` — a re-run that raised findings is not the clean re-run the guard's rule names; making every caller close panel slots `completed` and restamping the 50 rows — those rows record a true verdict, nothing is corrupt, and the guard would still break on the next legitimate `clean`.

### One optional letter after the case number

**ID:** lettered-case-label
**Status:** active
**Chosen:** `[0-9]+[a-z]?` in both regexes, the label kept as written (`Case 13b`) and its diff pattern `case\s*13b\b` — the shape the fifteen test files use.
**Considered:** `[a-z]*` — no test names a case with two letters; a second regex for the plan-shape side — the Python import exists so the grammar has one definition.

### The Go-tree audits run for any touched Go file

**ID:** go-audits-on-go
**Status:** active
**Chosen:** a fixed list, `pfGoAudits`, of the two harness paths, added when any `stats/**.go` path is touched and the harness exists in the worktree.
**Considered:** the whole `scripts/run-guard-tests.sh` — about a minute per verify for two harnesses; scanning harness bodies for `stats/` paths — a heuristic that picks up fixtures, for a list of two.

### The green sha lives in the worktree's git dir

**ID:** green-sha-git-dir
**Status:** active
**Chosen:** `run-verify-list.sh <wt> test`, exiting 0 on a tree clean before and after (`git status --porcelain --untracked-files=all` empty), writes HEAD to `git rev-parse --git-path flow-test-green`; a later run reads it only on a clean tree.
**Considered:** a `suite_runs` column in flowd — a schema change, and a store that is down would lose every skip; a file under `.superpowers/sdd/` — gitignored here but not in every consuming project, where it would dirty the tree; `git config` — shared by every worktree of the repository; a time window — not correct, a commit can land inside it.

### A command declares what it reads with `# inputs:`

**ID:** inputs-comment
**Status:** active
**Chosen:** a `# inputs: <pathspec>...` line inside the `## test` fence names what the next command reads; at a green sha other than HEAD a command is skipped only when `git diff --quiet <sha> HEAD -- <pathspecs>` exits 0, and a command declaring none always runs.
**Considered:** inferring inputs from the command — not derivable (this project's Go tests read outside `stats/`); skipping only when HEAD equals the sha — satisfies half of the finding, and drops the per-command half.

### The full-suite run goes through the same script

**ID:** full-suite-through-script
**Status:** active
**Chosen:** `skills/flow/implement.md`'s full-suite run invokes `run-verify-list.sh <canonical-worktree> test`, by hand only where it exits 3, so the green sha verify reads is the one that run recorded.
**Considered:** a separate `record` subcommand the parent calls after its own run — a second way to run the list, and a sha recorded for a run nobody checked.

### Only the agents SPA line declares inputs

**ID:** agents-inputs
**Status:** active
**Chosen:** `# inputs: stats/web/ stats/Makefile` above the `npm test` line of `.flow/project.md`'s `## test`; the guard harness and Go test lines stay undeclared.
**Considered:** `stats/` for the Go line — wrong, `stats/internal/guard`'s tests read `skills/`, `scripts/` and `setup.sh`, so a skill edit would skip them.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — F-D changes what `flow.verify` runs; a wrong skip lands an untested branch, and that needs a full first look.
**Considered:** `flow-low` — three one-line guard fixes, but the verify skip carries the risk.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-kan-887-flow-fix-delete-the-retained-bash-twins-and-fix`, Tasks 1, 2, 3 and 4 in order: each applies a diff already verified here, and the four are file-disjoint.
**Considered:** one fixer per task — four dispatches for four `git apply` runs; two groups split guards from the verify script — no shared state between them to justify the extra dispatch.

## Open questions
