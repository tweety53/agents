# Self-review context bundle for kan-594-flow-fix-the-project-key-coalesce-lives-as-two

found: 2 of 6 sources; skipped: 4 of 6 sources
skipped: spectre/changes/archive/kan-594-flow-fix-the-project-key-coalesce-lives-as-two/tasks.md (absent)
skipped: spectre/changes/archive/kan-594-flow-fix-the-project-key-coalesce-lives-as-two/design.md (absent)
skipped: spectre/changes/archive/kan-594-flow-fix-the-project-key-coalesce-lives-as-two/narrative.md (absent)
skipped: git log --stat (absent)

## .superpowers/sdd/ledgers/kan-594-flow-fix-the-project-key-coalesce-lives-as-two.md

# SDD ledger — kan-594-flow-fix-the-project-key-coalesce-lives-as-two

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-19T19:17:05Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-594-flow-fix-the-project-key-coalesce-lives-as-two-panel.md

# Review panel — kan-594-flow-fix-the-project-key-coalesce-lives-as-two

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|

findings-total: 0

reproducers-total: 0

## Pass log

### Round 0

- roster: compact — 63
- diff-size 32 lines — under cap, proceed
- docs-only: no — stats/internal/store/query.go; resolved roster runs
- no addition this round — the resolved list ran alone
- pass 1: primary+principles one dispatch — no findings; handshake line absent, single-model mapping satisfies it (zcode)
- build-green guard exit 1 (no **Build:** tag) — fixed as planning-path edit (**Build:** green), re-run exit 0

## Session narrative

This /flow-fast run implemented KAN-594 inline in the worktree: it held the stage-run effective project key expression `COALESCE(c.project_key, sr.project_key)` once as the unexported Go constant `projectKeyExpr` in `stats/internal/store/query.go` and repointed every live construction site at it — the stage-run allowlist's filter entry, `QueryStageRuns`' SELECT, and `ListRuns`' SELECT and project WHERE clause — updating the three doc comments that narrated the duplicated spelling (`StageRun.ProjectKey`, `TestQueryStageRunsReturnsChangeNameAndProject`, the allowlist entry). Folding `runs.go`'s two copies in was the run's own judgment beyond the issue's two named sites: with the constant in hand they were free to remove, and leaving them would have kept copies that could still drift. No new test was added — the behavior is pinned by the existing `TestQueryStageRunsReturnsChangeNameAndProject`, kept green, and no test can express "the expression is held once" short of the constant being the single holder, which is the fix itself. The round-0 panel (compact, primary+principles, one dispatch) raised no findings, and the full `## lint` list plus the scoped store suite (`go test ./internal/store/ -race -count=1`) passed in the worktree. Where the run struggled: the plan's first shape check failed on indented `- **Files:**` fields before the column-0 field grammar was met; the project's build-green guard demanded a `**Build:**` tag that flow-fast's reduced field list does not name (added as a planning-path edit, guard re-run to exit 0); and the panel reply omitted the `Model:` handshake line, which the zcode single-model mapping carve-out satisfies without a re-dispatch.
