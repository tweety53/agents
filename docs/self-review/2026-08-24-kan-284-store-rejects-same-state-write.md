# Self-review — kan-284-store-rejects-same-state-write

One `/myflow-fast` change, four tasks, run end to end across three invocations: brainstorm through
implementation, integrate as a pull request, then merge and archive. Panel roster `light`, three
required slots, zero findings on the first pass. Every per-task review found exactly one real defect.

Every finding below was explained in full before the filing prompt fired, and the operator declined
filing for all of them.

## Problems encountered, and what pipeline change would avoid them — `myflow-fix`

- **[myflow-fix]** The `IN_PROGRESS` handoff prints `Next: /myflow-fast <name>` while its own prose says "bare to move on to integrating it", but at `IN_PROGRESS` an argument means fix instructions and only a bare invocation integrates — so pasting the line the handoff supplies asks for the opposite of what the handoff says it does; it forced a blocking question on both later invocations of this run — declined
- **[myflow-fix]** `check-task-commit-fields.sh`'s wrapper requires exactly one non-archived `tasks.md` under `openspec/changes/`, and three `FINISHED`-but-unarchived changes (`kan-212`, `kan-242`, `kan-295`) make that condition unsatisfiable for every change in this repository, so all eight invocations had to call `check-task-commit-fields.py` directly; accepting an explicit `<changeRoot>` the caller already knows would fix it — declined
- **[myflow-fix]** `handoff-blocks.md`'s `Staged` enumeration asserts `/myflow-do` "only ever emits the first two", but under commit-per-task with no `prUrl` it leaves work committed and unpushed — a fourth state absent from both the enumeration and the review-command table, where choosing `--cached` prints nothing, which that same contract calls a wrong answer rather than a sparse one — declined
- **[myflow-fix]** `/myflow-start`'s git boundary stages planning artifacts in the main checkout, the implementation commits those same files from the worktree, and run 2's `prepare-archive-branch.sh` then refuses the main checkout as dirty over them; clearing them required verifying all six were byte-identical to `origin/main` by hand, and the obvious shortcut of deleting them unchecked would destroy work had they diverged — declined

## Token/time cost, and what would reduce it without quality loss — `myflow-cost`

- **[myflow-cost]** The one-implementer-per-worktree rule serialised four tasks whose declared `**Files:**` sets do not intersect (`cmd/myflow`, `internal/store` plus `internal/api`, `skills/`), costing roughly 25 minutes of implementer wall-clock where about 10 would have done; `**Files:**` is already the declared data needed to prove disjointness before dispatch — declined
- **[myflow-cost]** Four per-task reviewers and three panel slots each independently ran `go build`, `go vet`, `gofmt -l` and the affected test packages against a tree that changed little between them — roughly seven redundant full verifications — declined

## What went well, and how to reproduce it — `myflow-improvement`

- **[myflow-improvement]** Per-task review found one real defect in every one of the four tasks — a wall-clock-dependent test, a determinism claim no test pinned, a 500-instead-of-400 status mapping, and a wrong causal mechanism in contract prose — after which the panel found nothing; the reproducible part is that each reviewer was handed the specific claims its implementer had made, to verify rather than trust, instead of a bare "review this diff" — declined
- **[myflow-improvement]** Task 3's reviewer verified the journal-retirement fix by reverting the added `case`, rerunning the test, watching it fail, and restoring the file — turning an agent's claim into evidence, and confirming that this design's own assumption about `IsDefinitiveChangeOutcome` had been wrong — declined

## What could be automated or moved to a script — `myflow-automation`

- **[myflow-automation]** `check-normative-inventory.sh` reports a set rather than a verdict, so guarding a prose change means capturing a baseline before the first edit and diffing after the last — a manual ritual that yields no signal at all when a run forgets it; a wrapper that snapshots to a well-known path and diffs on demand would make it mechanical — declined
- **[myflow-automation]** `stats/internal/web/dist` is gitignored Vite output that `//go:embed all:dist` requires, so the declared `go vet ./...` lint command cannot pass in a fresh worktree until someone runs `npm run build`; building it during worktree isolation would remove the trap — declined

## What could move to the Go app or its persistent storage — `myflow-stats-app`

- **[myflow-stats-app]** `final-review-panel.md` is worktree-lifetime and is the only record of the panel's diff-size measurement, the cap in force, and which optional slots fired and were declined, while the store's panel record carries findings alone — so this run's decision to decline three fired triggers was destroyed with the worktree; the pass log belongs in the store as rows, like findings and dispatches already are — declined
- **[myflow-stats-app]** The stage-mark synthetic-change bootstrap still writes `STARTED` rows at nanosecond precision for names no operator chose, leaving three such rows in the store — one under the bogus project key `.claude-485b1ff8`, created by a mark fired outside any project; this change fixed the consequence, not the bootstrap itself — declined
