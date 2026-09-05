# kan-433-flow-verify-never-runs-lint-test-inline-build

**Jira:** KAN-433 — "flow: verify never runs lint/test inline, build SPA dist once per worktree"
**Source:** `docs/superpowers/research/flow-speedup.md` — cost lever 2 (widened in section 8) and
wall-clock lever 3.

## Why

Two observed failures of the same class — the conductor doing verification work itself that the
pipeline says a recorded subagent does:

- **kan-404.** The verifier reported a lint failure and the conductor fixed it inline after the
  review panel had closed: about 35 turns, about $1.8, 8 minutes, no dispatch record, no slot
<!-- measured: docs/superpowers/research/flow-speedup.md section 6 round three, from kan-404's conductor transcript usage rows @ 2026-09-04 -->
  re-run. `skills/flow/verify-and-handoff.md` binds only the verifier ("fixes nothing, edits no
  source"); nothing binds the conductor, so the stale-result rule **Panel re-runs**
  (`skills/flow/review-panel.md`) exists to enforce had no sentence to fire on.
- **kan-389.** The conductor ran the whole `## test` list itself between two verifier dispatches
  — about 4 minutes on a 200 k context — because the fresh worktree's first `go test` failed on
<!-- measured: docs/superpowers/research/flow-speedup.md section 8, from kan-389's flow.verify stage run and conductor transcript @ 2026-09-04 -->
  the missing `stats/internal/web/dist/`: that directory is gitignored and `//go:embed all:dist`
  refuses to compile without it. Not a defect in the branch; a fresh-worktree fact nothing built
  for.

A third cost sits beside them: `implement.md` §2 invokes `superpowers:using-git-worktrees`, whose
Step 2 installs dependencies and Step 3 runs the whole test suite as a baseline — 2–3 minutes that
<!-- measured: docs/superpowers/research/flow-speedup.md section 5 lever 3, from the isolate-workspace stage runs in the dev stats store @ 2026-09-04 -->
`flow.verify` repeats at the end of the same run.

## What changes

1. **The conductor never runs the `## lint`/`## test` lists itself and never edits source after
   the panel closes.** Its verify-stage work is `prepare-workspace.sh`, the verifier dispatch(es)
   and the ledger render. A verifier reporting a non-zero exit is re-dispatched once, under
   `-key verify-2`, carrying the first report verbatim; a second non-zero exit ends the
   conductor's turn with `## Question` naming the failing command and its output. **Panel
   re-runs**' stale-result rule is widened to cover any source change after a slot's last read,
   from any stage.
2. **A new optional `## worktree setup` key in `<project>/.flow/project.md`**, run by
   `implement.md` §2 from the worktree root immediately after `git worktree add`, first run only,
   resolved through `project-get.sh <worktree> "worktree setup"`. This repository declares
   `cd stats && make web-build` there, so a fresh worktree carries `dist` before any `go test`.
3. **`superpowers:using-git-worktrees` is no longer invoked.** `implement.md` §2 states the
   worktree creation commands itself; the baseline test suite and dependency install are gone.
4. **`flow record dispatch`'s usage text lists `verifier`**, which `recordRoles` already accepts;
   a test pins the usage text to the slice.

Touches: `skills/flow-contracts/project-configuration.md`, `.flow/project.md`,
`skills/flow/implement.md`, `skills/flow/verify-and-handoff.md`, `skills/flow/review-panel.md`,
`skills/README.md`, `skills/flow-contracts/finish-contract-run1.md`,
`stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`.
