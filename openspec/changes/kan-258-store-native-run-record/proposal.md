## Why

The pipeline's **derived** records — the SDD ledger, the review panel's findings, each dispatch's
model and cost — are files in a gitignored worktree. A file has a path, and no component agrees on
it. When they disagree, `preserve-session-records.sh` prints `skipped: <src> (absent)` and exits 0,
which the outcome table in **Preserving the session records** (`skills/myflow-contracts/pipeline.md`)
defines as a legitimate outcome, because a change may genuinely have no such record. So **"never
written" and "written somewhere else" are indistinguishable to every caller**, and the record dies
with the worktree at run 2.

KAN-77 fixed one instance of that by naming one canonical path and adding a rescue list. The class
survives, because it exists only while the record is a file.

Evidence here: **34 preserved ledgers against 42 preserved panel records** — eight runs whose ledger
is simply gone.
<!-- measured: ls docs/superpowers/ledgers/ | wc -l; ls docs/superpowers/reviews/ | wc -l @ branch main -->
`kan-13-myflow-planning-and-status-fixes` lost its ledger outright. KAN-73's dispatch history was one
`git worktree remove --force` from destruction, caught only because a human read a skip line
carefully — **192 lines**.
<!-- measured: KAN-258's own description, "Why now" section @ the issue as fetched 2026-08-22 -->

`stats/` already owns the state store and the stage-mark telemetry, with a daemon, PostgreSQL, a
journal-on-failure path and a reconciler. Extending it to own the run record is a schema and an
ingest path, not a new application. In a store there is no path to disagree about, and "no rows for
this run" is a different value from "rows exist and nobody looked in the right place".

## What Changes

- **Two new tables** — `dispatches` and `findings` (migration `0010_run_records.sql`). `dispatches`
  is one row per subagent dispatch and carries the SDD ledger, task → commit → model, and
  per-dispatch cost together; they are three views of one row, not three records.
- **A new CLI verb** — `myflow record dispatch | finding | status | render`, traversing the same four
  layers `myflow stage` already does (`cmd/myflow` → `internal/client` → `internal/api` →
  `internal/store`). A store failure journals the intent, warns once and exits 0, exactly as a stage
  mark does.
- **The Markdown record becomes generated output.** The agent stops hand-writing the SDD ledger and
  the panel record; `myflow record render` writes them into `<repo>/docs/superpowers/` directly. The
  panel record renders at panel close, so `check-unfinished-work.sh` reads the same marker block it
  reads today and needs no change.
- **Per-dispatch cost is derived, never self-reported.** The dispatch records task, role, model and
  slot — recorded intent. `internal/harvest` attributes token usage against that row in a second,
  independent pass over the same transcript records; `stage_runs.metrics` is unaffected.
- **The `/myflow-do` handoff names journalled record writes.** A never-block guarantee that is also
  silent is the failure this change exists to remove.
- **BREAKING — `preserve-session-records.sh` is retired**, together with
  `scripts/test-preserve-session-records.sh`. `pipeline.md`'s "Preserving the session records"
  outcome table is replaced by a render-outcome table; `finish-contract.md`'s run 1 preservation duty
  becomes a render duty; both call sites, the guard-presence lists in three skills, `setup.sh`'s
  symlink set and `check-guard-symlinks.sh` move with it. The `finish.preserve-sessions` **stage key
  is unchanged** — only what it invokes changes.

**Not in scope, deliberately:** `tasks.md`, `proposal.md`, `design.md` and the delta specs stay
Markdown — they are authored artifacts, git-versioned and PR-reviewed. Task **completion** stays in
`tasks.md`'s checkbox, so `check-unfinished-work.sh` and `pipeline.md`'s single-source rule are
untouched. The SPA is untouched. The store starts empty; the 34 archived ledgers and 42 preserved
reviews stay as committed Markdown, unparsed.

## Capabilities

### New Capabilities

- `myflow-run-record`: the store-native record of a change's dispatches and review findings — its
  schema, its write path and never-block guarantee, its rendering into the repository, and the
  retirement of file-based preservation.

### Modified Capabilities

- `myflow-finish-cleanup`: **Session records are preserved in the repository** — preservation by
  file copy becomes rendering from the store, with a new outcome vocabulary.
- `myflow-model-policy`: **Every dispatch records the model it used** — the dispatch's model is
  recorded in the store rather than in a preserved file, and the audit trail no longer depends on a
  file having survived.
- `myflow-review-panel-economics`: **Panel findings are recorded in a table**, **Every finding's
  state is recorded on an anchored marker line**, and **The panel record declares how many findings
  it carries** — the findings become store rows and the record becomes a render of them. The marker
  block's format and the guard that reads it are unchanged.
- `myflow-handoff-output`: the `IN_PROGRESS` handoff carries a line naming how many record writes
  for this change are still journalled.
- `agents-repo-verification`: **This repository declares its own myflow project configuration** — the
  `## test` list this requirement names drops `scripts/test-preserve-session-records.sh`, retired
  with the script it tested. Every other harness and clause is unchanged.
- `myflow-self-review`: **Context is gathered deterministically, not by the model re-reading files**
  — the ledger's date-prefixed location and the `skipped:` outcome vocabulary are stated as
  `gather-self-review-context.sh`'s own, rather than as a match to the retired copying script's.

## Impact

**Go / SQL** — migrations `0010_run_records.sql` and `0011_dispatch_agent_id.sql`; new
`stats/internal/store/records.go`; new `stats/internal/records/` (the wire types and the Markdown
renderer); `stats/internal/api/records.go` and four routes on `internal/api/server.go`;
`stats/internal/client/client.go`; `stats/cmd/myflow/record.go` and `main.go`'s usage block;
`stats/internal/harvest/attribute.go` and `watcher.go` (a second attribution pass);
`stats/internal/reconcile/reconcile.go` (`replayRecordFile`) and `stats/cmd/myflow/journal.go`, its
third `reconcile.New` call site; `stats/cmd/myflowd/main.go` (wiring); and
`stats/internal/web/embed_test.go`, whose fake store had to gain the new interface's methods.

**Contracts and skills** — `skills/myflow-contracts/pipeline.md`,
`skills/myflow-contracts/finish-contract.md`, `skills/myflow-contracts/handoff-blocks.md`,
`skills/myflow-do/SKILL.md`, `skills/myflow-finish/SKILL.md`, `skills/myflow-fast/SKILL.md`, the
three rationale files beside them (`pipeline-rationale.md`, `myflow-do/SKILL-rationale.md`,
`myflow-finish/SKILL-rationale.md`), and `.myflow/project.md`.

**Guards** — `scripts/preserve-session-records.sh` and `scripts/test-preserve-session-records.sh`
deleted, together with the three skill-local symlinks to the former under
`skills/myflow-{do,fast,finish}/scripts/`. Modified: `check-unfinished-work.sh` and
`check-panel-reproducers.sh` (both now read the panel record where it is rendered),
`check-stage-mark-calls.sh` (it now covers `myflow record dispatch`'s session token),
`check-contract-budget.sh`'s budget table, and the nine scripts whose comments cited the retired
script as canonical — `check-cleanup-complete.sh`, `check-workspace-isolation.sh`,
`gather-self-review-context.sh`, `lib/panel-record.sh`, `lib/resolve-file.sh`,
`plan-dispatch-bundles.sh`, `resolve-base-branch.sh` among them — plus their harnesses.

**`setup.sh` is NOT modified.** It never named the script; it symlinks whole skill directories, so
retiring a guard is a matter of deleting the committed symlinks, not of editing the installer.

**Unaffected, and asserted so by test** — `openspec/changes/**/tasks.md` and every parser that reads
it (`check-task-build-green.py`, `check-task-commit-fields.py`, `lib/plan_grammar.py`), and
everything under `stats/web/`. `check-unfinished-work.sh` is **not** in this list: the marker
**contract** it parses is unchanged, but its **path** moved with the record, so the guard itself was
edited.
