# kan-1067-flow-allowed-pairs-experiments-stored-in-flowd — session narrative

## 2026-10-30 — creating run

- Task 4's planned `//go:embed` in `internal/modelpolicy` would have broken plain `go build ./cmd/flow-guard` (the embedded file is gitignored and only make writes it); the embed moved to `internal/modelpolicy/embedded/` and flowd passes the table in through `api.WithExperiments`, guarded by `TestFlowCLIDoesNotImportEmbeddedPolicy`.
- A throwaway mutation copy cannot build `internal/api` (the make-copied embed is absent), so the mutation slots copied it in by hand from round 1 on; `TestRolesRoute` mutations were not run in round 0.
- The full suite was red on `origin/main` too (stale baseline, controls, combinations and run-detail visual specs against the reworked UI); fixed in-run and re-baselined rather than deferred.
- First verify failed three lint guards (plan records naming files that shipped in later review-fix commits, two tasks' `Tests: none` beside test files, two new AskUserQuestion sites with no silence outcome) and then `check-verbatim-moves` on the reworded sentences; fixed by In-run fix 1 and an acknowledgement-file entry.
- Verify keys suffixed `-fix-1b`/`-fix-2b` are not folded into their `-fix-<k>` group by `check-verify-green.sh`, so the green re-run was not seen until re-keyed `verify-fix-2`; key verify re-runs strictly `verify-fix-<k>`.
- The live check's step 4 (non-zero before figures right after opening) contradicts the decided equal-length windows: a just-opened experiment has an empty before window. Measured on the scratch copy with `started_at` moved 14 days back instead.
- Visual verification took four fix rounds on the Experiments view: list order and a wrapping delta (plus empty charts from a seed lacking token/wall-clock metrics), then an unchanged metric drawn as better/worse, then a mean-after line with no before drawn as worse — each found only by the next scoped run; the last scoped run was briefed to sweep every before/after combination on both experiments and came back clean.
- A dispatched implementer committed the visual evidence directory inside its task commit; `close-task` swept it, and the commit was split so the evidence rides a `Flow-Key: flow.visual-verify` commit.

## 2026-10-30 — fix run

- The store counts this session's `flow.document-fix` marks — the four in-run verify and visual-verify fix rounds (Tasks 14–17) — as a fix run; their account is the last three bullets of the creating run above.

## 2026-10-30 — integrate run

- `origin/main` had moved 3 commits; the sync's rebase conflicted in `stats/web/src/styles.css`, where the base's `.lb-clip` rule and this change's Experiments rules were added at the same spot — both kept.
- The base's one-line activity text shortened the state board by ~120px, failing the dashboard, combination-chips-plus-tile and full-state-board captures; re-captured, read, zip rebuilt, and a full visual run on the synced HEAD confirmed only the base moved them.
