## Why

KAN-73's self-review measured ~1.98M subagent tokens across 12 dispatches and found one shape
recurring: a dispatch boundary discards work already done on the other side of it. Every review-panel
slot separately locates and reads the same proposal, design, plan, delta spec and principles file;
the fix round — the single largest dispatch at 321,403 tokens — re-reads what the slots had just
analysed, even though every finding already carries a file, a line and a runnable reproducer. The
same shape holds for the implementer dispatches, at 275,064 and 257,357 tokens.

Nothing measures this from inside the pipeline. Subagent cost is harvested and priced today, but it
collapses into one `sidechain` bucket per stage, so `do.review-panel` cannot distinguish a slot from
a fix round and `do.sdd-tdd` cannot distinguish an implementer from a per-task reviewer. The baseline
above came from a self-review written after the fact, which is not a thing a later change can be
measured against.

## What Changes

- **A gathered context bundle.** A new `scripts/gather-dispatch-context.sh` collects the change's
  proposal, design, plan, delta specs and engineering principles into one
  `.superpowers/sdd/dispatch-context.md`, following `gather-self-review-context.sh`'s contract and
  security posture exactly. It is rebuilt at the start of each dispatching stage and before each fix
  round, never once per run.
- **Every dispatch reads the bundle.** Implementer, panel-slot and fix-subagent prompts each gain one
  paragraph naming the bundle path. The bundle is **advisory**: a dispatch may open any file it names,
  and must still read the actual diff and the actual code. Sharing inputs is permitted; sharing
  conclusions is not.
- **The fix dossier carries located evidence.** The fix dispatch receives, per surviving finding, its
  `F<n>`, slot, severity, `file:line` taken verbatim from the findings table, theme, reproducer text
  and bounce history — stated as established fact rather than as a location to re-derive. No source
  excerpts are inlined.
- **Per-dispatch cost attribution.** `harvest` reads each subagent transcript's `agentId` and its
  sibling `.meta.json` descriptors, and `harvest.Delta` gains a `Dispatches` breakout feeding the
  metrics bag's `dispatches.<agentId>` key — built exactly as the existing `Models` breakout is.
  Additive only: `tokens.main`, `tokens.sidechain` and `models.*` keep their current meaning.
- **A surface for it.** `RunDetail`'s stage-run table gains an expandable row showing that run's own
  dispatches — description, agent type, model, tokens, cost — sorted by cost descending.

No **BREAKING** changes. Nothing here alters which slots run, the zero-open-findings handoff bar, the
reproducer verification rules, the commit-per-task model, or any existing aggregation.

## Capabilities

### New Capabilities

None. Every requirement below extends a capability that already exists.

### Modified Capabilities

- `myflow-dispatch-economy`: gains the gathered-bundle requirements — what the bundle contains, when
  it is rebuilt, that it is advisory rather than authoritative, and that its absence never stops a
  run.
- `myflow-review-panel-economics`: panel slots and the fix subagent read the bundle, and the fix
  dispatch carries each surviving finding's located evidence rather than a bare list.
- `myflow-run-telemetry`: subagent token usage is attributed per dispatch, keyed on `agentId`, with
  descriptors read from the transcript's sibling meta file.
- `myflow-stats-views`: one change's stage run opens onto its own per-dispatch cost rows.

## Impact

- **New:** `scripts/gather-dispatch-context.sh` and `scripts/test-gather-dispatch-context.sh`;
  `setup.sh` gains the guard's symlinks, enforced by `scripts/check-guard-symlinks.sh`.
- **Skills:** `skills/myflow-do/SKILL.md` sections 4 and 5 (bundle build points, three dispatch
  prompts, the fix dossier). `skills/myflow-contracts/pipeline.md` gains a **Temporary artifacts
  registry** row for `dispatch-context.md`.
- **Go:** `stats/internal/harvest/transcript.go` (decode `agentId`), `attribute.go` (the `Dispatches`
  breakout and the meta sidecar read), and their tests.
- **SPA:** `stats/web/src/views/RunDetail.tsx` and its test.
- **Lint:** no `scripts/check-contract-budget.sh` budget raise is expected — `pipeline.md` sits at
  46105 bytes against a 55728 budget and `skills/myflow-do/SKILL.md` at 57995 against 71317, so both
  additions fit their existing headroom.
- **Tests:** `.myflow/project.md`'s `## test` list gains the new harness.
