## Context

- Four non-`big` findings from the kan-951 self-review pass, planned as one change per item 1 of
  **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`). Each is carried by
  exactly one task; none is dropped. All were checked against the tree at `edda7db8` and, for the
  gymie finding, against the gymie worktree `kan-951-cors-preflight-max-age` at `f71533bf`:
  - A (Task 1): gymie `.flow/project.md`'s `## visual verification` `verify` and `capture` rows
    prefix `BACKEND_URL` alone. `gymie-playwright/playwright.config.ts` sets `baseURL` to
    `process.env.FRONTEND_URL ?? 'http://localhost:3000'`, and the section's own `url` row derives
    `FRONTEND_URL` from `FRONTEND_PORT`. `scripts/check-visual-verification.sh` validates the section
    with and without the fix.
  - B (Task 2): **The parent runs the full suite — never an implementer.**
    (`skills/flow/implement.md`) says the parent "commits on the route the branch's push state
    dictates" and records a `full-suite-fix-<n>` pair, but names neither the `Flow-Key:` trailer
    **Commits outside every task** (`skills/flow-contracts/git-boundaries.md`) requires, nor the
    row's `-commit`, nor the PROOF RUNS paragraph's reverted-fix proof. kan-951's ledger row
    `full-suite-fix-1` carries `697183272a`, the same sha as `task-1-implementer`.
  - C (Task 3): Visual verification's steps 1–2 (`skills/flow/verify-and-handoff.md`) feed
    `check-visual-trigger.sh` `<merge-base>..HEAD` on every run. On a fix run the late-fix reduction
    (`skills/flow/review-panel-late-fix.md`) has already resolved each worktree's since-close sha.
  - D (Task 4): traced. `UpsertFinding` (`stats/internal/store/records.go`) is
    `ON CONFLICT ON CONSTRAINT findings_ref_key DO UPDATE` — a ref is unique per change and a second
    write rewrites the row, by design, for replay and restatement. `flow record findings -change
    kan-951-cors-preflight-max-age` returns one row, `F1` round 3 ("stray trailing comma"), while
    the creating run's archived `panel.md` (gymie `777076ca`) shows `F1` round 0 ("commit c292e67d
    is outside tasks.md"): the third run numbered from `F1` and overwrote it. `review-panel.md`
    tells the parent "the store's own constraint enforces" ref uniqueness, which the upsert does
    not. The wrong-project hypothesis is disproved: the store holds no record of the change under
    the `gymie-playwright` or `gymie-frontend` project keys. Separately, dispatch
    `panel-1-primary+principles` ran with no pass, finding or mutation row for round 1, and
    `RenderPanel` (`stats/internal/records/render.go`) derives the pass log's rounds from pass and
    mutation rows alone, so the round is absent from the record.
- Every task's diff was applied in plan order to the worktree at `edda7db8`, checked, and reverted:
  `go vet`, `gofmt -l` and `go test ./internal/guard/ ./cmd/flow/ ./internal/records/` pass with all
  applied; both new Go tests fail with Task 4's production files reverted; `check-verbatim-moves.sh`
  (with this change's `verbatim-moves.txt`), `check-references.sh`, `check-markdown-integrity.py`,
  `check-dispatch-paragraphs.sh`, `check-installed-citations.sh`, `check-normative-inventory.sh`,
  `check-guard-symlinks.sh`, `check-stage-mark-calls.sh` and `check-ask-silence.sh` exit 0.
  Task 1's diff passes `git apply --check` in the gymie worktree.

## Decisions

### gymie's verify and capture rows set FRONTEND_URL from FRONTEND_PORT

**ID:** frontend-url-in-verify-rows
**Status:** active
**Chosen:** prefix both rows with `FRONTEND_URL=http://localhost:${FRONTEND_PORT:-3000}`, the same
`${VAR:-default}` shape the rows already use for `BACKEND_URL`, so the main checkout's run is unchanged.
**Considered:** changing `playwright.config.ts`'s fallback to read `FRONTEND_PORT` — a second
repository edited for what one config row states; making the verifier export every `KEY=value` line
into its `verify` shell — a contract change across every project for a row this project can carry.

### The full-suite fix names its trailer, its sha and its proof

**ID:** full-suite-fix-obligations
**Status:** active
**Chosen:** four sentences in the full-suite fix paragraph: the commit carries
`Flow-Key: full-suite-fix-<n>`; `check-untasked-commits.sh` runs before the push, while a missing
trailer can still be amended on the unpushed tip; the pair's `end` records the sha `rev-parse HEAD`
prints after that commit; a fix for a failure that does not recur every run is held to PROOF RUNS —
shown returning with the fix reverted, and not committed when it never returns. No new guard: the
trailer's guard already exists and is now run where its exit 1 is still curable.
**Considered:** running the guard at the panel's entry (`write-panel-diff.sh`) — by then the commit
is pushed and cannot be amended without a rewrite **Panel re-runs** forbids on pushed history; a new
guard comparing a `full-suite-fix-<n>` row's `-commit` against its `Flow-Key:` commit — no second
caller, and the sentence naming the source of the sha closes the observed defect.

### A fix run's visual trigger reads the since-close diff

**ID:** fix-run-visual-trigger-since-close
**Status:** active
**Chosen:** on a fix run, steps 1–2 and the `--repo` call feed the trigger
`<since-close-sha>..HEAD` per worktree, the sha the late-fix reduction already read; a worktree whose
since-close sha is `-` keeps the merge-base diff, and every worktree keeps it when the change's
newest verifier dispatch row (the newest row `flow record dispatches -change <name>` returns with `role` `verifier` and a `key` starting with `visual-verify` — the rows `scripts/check-visual-verify-dispatched.sh` counts) does not have `outcome` `completed`. The since-close range is safe only when every
earlier run's visual verify finished: a run whose panel closed clean but whose verifier never
completed leaves a UI change before the since-close sha, which the narrowed range would miss.
**Considered:** the fix run's entry HEAD — never persisted, and unsafe when the previous run's visual
verify never finished; a new persisted "last verified sha" — a stored value where one is already
derived. The in-run full-suite `verify` in `implement.md` is left on the merge-base diff: the finding
names the stage's call site, and blast radius stays inside it.

### A run opens its finding refs where the store's last one ends

**ID:** next-ref-from-store
**Status:** active
**Chosen:** `flow record next-ref -change <name>` prints `F<n>`, one past the highest ref any finding
row of the change carries (numeric, so `F10` beats `F9`), `F1` for none, nothing and exit 1 on a
failed read — `next-round`'s twin, sharing its record read. `review-panel.md` opens a run's refs
there and corrects the sentence claiming the store's constraint enforces uniqueness.
**Considered:** the store refusing a ref collision — the upsert is how a journal replay and a fix
round's restatement land, so a refusal breaks both; re-keying findings on (change, round, ref) —
`flow record status -ref F<n>` becomes ambiguous and every guard reading refs changes.

### A panel round with no pass row still renders

**ID:** render-dispatched-round
**Status:** active
**Chosen:** once a change has any pass-log row, `passLogRounds` also takes the round of every
`reviewer` dispatch keyed `panel-<round>-…`; such a round with no pass row renders
`- no pass row recorded — dispatched: <keys>`. A change with no pass-log row at all renders as before.
**Considered:** an integrate guard refusing a round with no pass row — a second gate for a fact the
record can state itself; rendering dispatch-derived rounds for every change — old records with no
pass log would gain a section, breaking the byte-identical rendering `TestRenderPanelWithoutPassLogIsUnchanged` pins.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `medium` — Task 4 adds a CLI subcommand and changes the panel renderer's round set, and
Task 2 changes when a guard runs against pushed history; a low-effort read could pass over either.
**Considered:** `low` — three of the four tasks are a config row or a few prose lines, but Task 4
carries the review risk.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** one group, g1: Tasks 1–4 in plan order, one fixer — the four tasks share no file, the
three agents-repo diffs are small, and Task 1 is a two-line config edit in the gymie worktree.
**Considered:** two groups splitting Task 4's Go work from the rest — one more dispatch for a plan
whose largest task is one spiked diff.

## Open questions

## Live check

None — Task 4's subcommand only reads the store and its CLI test drives it against an HTTP double of
the daemon; exercising it against the running flowd is a read the plan already made by hand
(`flow record findings`), and the renderer change is pure. Task 1's rows run only inside a
visual-verify stage, Tasks 2 and 3 change prose.
