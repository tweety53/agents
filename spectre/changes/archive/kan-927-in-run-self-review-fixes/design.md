## Context

- `finish-contract-run1.md` **Save the self-review context bundle**: "Self-review is always deferred: no reasoning pass and no prompt run here." Integrate commits `docs/self-review/<name>-context.md`; `/flow-fast` 5. Verify does the same.
- `/flow-self-review` (KAN-875) fixes every non-`big` finding, but only when someone runs it by hand on the default branch. About 16 bundles are pending, and none has been processed since auto-fix shipped.
- Every fix, review and re-review it dispatches runs on `opus` at `flow-high`. So do the in-run pipeline fixes (`pipeline.md` **Pipeline defects found mid-run**, KAN-829).
- In-run pipeline fixes are recorded only as a narrative line. `self_review_findings` holds 92 rows, 1 of them `fixed`.
- The store serves `GET /api/v1/self-review/{project}/{change}/findings` (one change). `stats/web` has no view of it.
- The operator's intent (KAN-927): every non-`big` finding, from any angle, is fixed during the run, fast and cheap. Every fix is stored, and a UI view shows them.

## Decisions

### The pass runs in the integrate run

**ID:** in-run-pass-at-integrate
**Status:** active
**Chosen:** integrate run 1 runs the six-angle pass inline after the archive commit and before the route, on the integrate session's model, over the same bundle `flow self-review bundle` assembles (held in memory, not committed). `/flow-fast` runs the same pass at **5. Verify**. The report `docs/self-review/<name>-self-review.md` is committed on the change branch and lands with the change. `/flow-self-review <name>` stays as the fallback, for a bundle an older run saved.
**Considered:** at the end of implementation, before the `IN_PROGRESS` handoff — the archive, narrative and full ledger don't exist yet, and the pass would run again on every fix run; keep the pass standalone and only auto-invoke it — it is still a separate session on the default branch, so the fixes still don't land with the change.

### Where self-review fixes land

**ID:** fixes-ride-the-change
**Status:** active
**Chosen:** a change whose canonical repository is `<agents repo>` takes the fix commits on `spectre/<name>`, after the archive commit, so they ride the run's own route. Any other project's change puts them on `self-review-<name>` in an `<agents repo>` worktree, landed by `<agents repo>`'s `## default landing route` as **Pipeline defects found mid-run** step 4 does.
**Considered:** always a separate `<agents repo>` branch — a second merge to `main` for every agents change, which the operator rejected ("the fixes land with the change"); always the change branch — a gymie branch cannot carry `<agents repo>` files.

### Cheap dispatch pairs

**ID:** cheap-dispatch-pairs
**Status:** active
**Chosen:**
- One fixer for the whole batch of non-`big` findings: `flow-medium`, `opus` by default, `sonnet` when every fix is a literal edit with nothing it can break (the reason recorded with the dispatch).
- The first review: `flow-medium` `opus`.
- Every re-review after a fix: `flow-medium` `sonnet`.
- The same pairs apply to in-run pipeline fixes (**Pipeline defects found mid-run**).
- The no-progress escalation in **Fewest operator actions** stays `flow-high`.
**Considered:** `sonnet` for the first review too — the operator's message allows it, but the global model rule keeps every first-pass review on `opus`; `flow-high` everywhere (today) — the cost the operator objected to.

### Targeted verification

**ID:** targeted-verify
**Status:** active
**Chosen:** after a clean review, the session runs only the `## lint` lines the touched files need, plus the tests covering them (`go test` on the touched packages with `-run` where a test is named, the `scripts/test-*.sh` harness of a touched script, `npx vitest run <file>` for a touched SPA file). No full suite. A red line loops back through the fixer, as today.
**Considered:** the full `## lint` and `## test` list (today) — about 4 minutes per pass, the cost the operator objected to.
<!-- measured: kan-916 self-review verify, 2026-10-07 — run-guard-tests.sh "57s wall", store package alone 94.9s, full lint list before them; machine-local timing, not re-runnable at a ref -->

### Every fix is a store row

**ID:** every-fix-a-row
**Status:** active
**Chosen:**
- An in-run pipeline fix, landed or deferred, calls `flow self-review finding -change <name> -angle flow-fix -disposition fixed -ref <sha> -blast-radius <N>` when landed. A deferred one gets no row until the pass decides it.
- The pass records every finding as today.
- The narrative line stays as a copy.
- No new disposition and no schema change.
**Considered:** a `deferred` disposition — the pass already resolves a deferred fix to `fixed`/`filed`/`declined` in the same run.

### The filing ask and the rating stay

**ID:** filing-ask-stays
**Status:** active
**Chosen:** the pass asks the one combined filing-and-rating question at integrate (shape per today's step 4). Silence files nothing. A bare `/flow` invocation means an operator is present.
**Considered:** filing `big` findings without asking — a Jira create is outward-facing, and the operator wants findings explained first.

### Self-review fixes view

**ID:** self-review-view
**Status:** active
**Chosen:**
- A new stats view `self-review` ("Self-review fixes"), served as `GET /api/v1/stats/self-review` with the existing period/project filters. One row per finding: recorded at, change, angle, finding, outcome, ref.
- `fixed` renders the ref as a commit link built from the daemon checkout's `origin` remote (GitHub https form, `commitUrl` in the DTO), or as plain text when the remote is not GitHub.
- The DataTable's column filter on outcome serves as the outcome filter.
- A second table holds per-change fixed/filed/declined counts, derived client-side from the same rows.
**Considered:** a section inside RunDetail — per change only, with no cross-change view; a Jira link for `filed` — the daemon knows no Jira base URL.

## Live check

- flowd on `127.0.0.1:4173` (the dev workspace's, never restarted by the run; the operator reloads it):
  - `curl -s 'http://127.0.0.1:4173/api/v1/stats/self-review?from=2026-10-01T00:00:00Z&to=2026-10-09T00:00:00Z'` returns the 96 existing rows (92 plus kan-916's 4).
  - kan-916's two `fixed` rows carry a `commitUrl` under `github.com/tweety53/agents/commit/`.
- The SPA at that address shows the "Self-review fixes" tab with those rows, and the per-change table shows kan-916 with 2/2/0.
- Failure looks like: 404 on the route, an empty table while the DB holds rows, or `fixed` rows with no link.

## Open questions
