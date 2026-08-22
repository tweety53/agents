# Store-native run record

## The problem

A derived record is a file, a file has a path, and no component agrees on the path. Preservation's
`skipped: <src> (absent)` is both "this change has no such record" and "the record is somewhere I did
not look" — one string for two facts, one of which is a silent data loss. That ambiguity is a
property of file-ness, not of any particular path, so naming a canonical path (KAN-77) removes an
instance and leaves the class.

## What is NOT the problem

**The authored plan.** `tasks.md` is written by `superpowers:writing-plans`, read as prose by every
implementer, parsed by three guards, reviewed in the PR diff and archived with the change. Moving it
into Postgres costs git-versioning, PR review and offline readability and buys nothing — it is
authored once and read linearly. The same holds for `proposal.md`, `design.md` and the delta specs.

**Task completion.** `pipeline.md` states `tasks.md` is the single source and that no second source
may exist, because `check-unfinished-work.sh` reads that file. Nothing here changes that.

The split this change takes: **authored artifacts stay in git; derived records become store-native;
the Markdown record becomes a rendering of the store**, kept so the archive stays readable with no
daemon running.

## Decisions

### The four records are two tables, not four

**ID:** `dispatches-is-one-table`
**Status:** active
**Chosen:** One `dispatches` table carries the SDD ledger, task → commit → model and per-dispatch
cost — three views of one row. `findings` is a second table because a finding has a different
lifecycle: raised by a dispatch, then updated by later rounds.
**Considered:** Four tables, one per subsumed ticket — three would share a primary key and be joined
on every read. One generic `run_records` table with a `kind` column and a JSONB payload — severity,
status, model and role are exactly the columns the ticket's own search cases filter on, and this
buries them in JSON. Modelling a dispatch as a nested `stage_runs` row — `(change_id, command,
stage, attempt)` uniqueness and the documented stage-key allowlist both fight it, and
`check-stage-mark-calls.sh` rejects any key absent from the Level 1 table.

### Scope is the whole umbrella, in one change

**ID:** `full-umbrella-scope`
**Status:** active
**Chosen:** All four record types land together — the operator's decision at the design gate.
**Considered:** Foundation plus the SDD ledger alone, with KAN-218/212/237 following as small
changes; foundation plus ledger and findings. Both were offered and declined.

### The Markdown record is generated output; the agent stops writing it

**ID:** `store-is-the-only-writer`
**Status:** active
**Chosen:** The skill instructs the agent to write records through the CLI only. Every Markdown
record becomes a render of the store.
**Considered:** Writing both, with the file as the render target — two sources that can disagree is
the problem this change exists to remove, and keeping the hand-write keeps it.

### `tasks.md`'s checkbox stays canonical for completion

**ID:** `completion-stays-in-the-file`
**Status:** active
**Chosen:** Store rows are a record, never a gate. `check-unfinished-work.sh` and `pipeline.md`'s
single-source rule are untouched.
**Considered:** Making the store canonical — it puts a hard gate behind a daemon that is explicitly
allowed to be unreachable, and the guard would then have to distinguish "no rows" from "cannot reach
the store" on the one path where guessing wrong is unrecoverable.

### The store starts empty

**ID:** `no-backfill`
**Status:** active
**Chosen:** No import of the 34 archived ledgers or the 42 preserved reviews. They stay as committed
Markdown.
<!-- measured: ls docs/superpowers/ledgers/ | wc -l; ls docs/superpowers/reviews/ | wc -l @ branch main -->
**Considered:** A best-effort `myflow record import` — the archived format was never stable (two
ledgers carry hand-typed filenames the existing search pattern cannot match), so the parser would be
written against drift. A required full backfill — same, plus it makes an unparseable historical file
block a change about new runs.

### The panel record renders at panel close

**ID:** `render-panel-at-close`
**Status:** active
**Chosen:** Rendering fires when the panel closes, so `check-unfinished-work.sh` reads the same
marker block it reads today and needs no change.
**Considered:** Rendering only at finish run 1 — the guard runs during `/myflow-do`, long before
run 1, so it would have nothing to read. This choice also satisfies KAN-205 ("validate the panel
record's marker block at panel close, not at finish run 1") as a consequence, and makes KAN-65's
cumulative-record defect unreachable, since each render replaces the file from the current rows.

### `preserve-session-records.sh` is retired outright

**ID:** `retire-preservation-script`
**Status:** active
**Chosen:** The skills call `myflow record render` directly; the script and its test harness are
deleted, and `pipeline.md`'s outcome table is replaced.
**Considered:** Keeping it as a thin wrapper that invokes the render and reports the same four words
— it preserves an indirection that no longer indirects anything. Rendering into the worktree and
letting the script keep copying — that keeps the file-path indirection this change is about. Both
were offered and declined.

### The `finish.preserve-sessions` stage key is unchanged

**ID:** `keep-the-stage-key`
**Status:** active
**Chosen:** The stage keeps its key and its place in run 1; only what it invokes changes.
**Considered:** Renaming it to `finish.render-records` — it would invalidate every recorded stage run
carrying the old key and force a matching edit to `README.md`'s Level 1 table, which
`stats/internal/stages/names_test.go` parses. Cost with no benefit: the stage does the same job at
the same point.

### Cost comes from the harvester, intent from the dispatch

**ID:** `harvester-owns-cost`
**Status:** active
**Chosen:** The parent writes task, role, model and slot as the dispatch *opens* (see
`dispatch-begin-and-end`, which supersedes this decision's original "at dispatch close" wording on
that one point and nothing else); `internal/harvest` attributes token usage against that row
afterwards, in a second, independent pass over the same
transcript records.
**Considered:** The dispatching agent reporting tokens explicitly — Cursor and Codex write no
transcript, so the figure would be absent or guessed on two of three harnesses. Deriving the model
from the transcript too — `pipeline.md`'s Model policy requires recorded intent and forbids a
plausible-looking guess.

### A dispatch is recorded by a begin/end pair, not by one call at close

**ID:** `dispatch-begin-and-end`
**Status:** active
**Supersedes:** the "One call per dispatch, at close" ruling in **2. The write path** below, which is
left in place, struck through, rather than deleted — a decision that was wrong is part of this
change's record.
**Chosen:** `myflow record dispatch begin` writes the row as the dispatch starts, carrying the task,
role, slot, model, agent id, session token, dispatch key and start instant, and prints the seq the
store allocated. `myflow record dispatch end` writes the commit, the outcome and the end instant.
Both halves journal-and-exit-0 on a store failure, and `internal/reconcile` replays both.

**Why the original reasoning was wrong.** It said the parent already knows the task, the commit, the
model and the outcome at close, "so a begin/end pair would double the call count and buy nothing".
Every clause of that is true and the conclusion does not follow, because the reasoning is about what
the *dispatcher* knows and the row's other reader is the *harvester*. `internal/harvest`'s watcher
commits the transcript offset it has consumed on every cycle and never re-reads behind it, so each
transcript record is offered to attribution exactly once — at the tick that consumed it. A dispatch
row that appears only at the close did not exist for any tick that ran while the subagent worked, so
every one of those records found no window: dropped by `DispatchAttributor.Attribute`'s
`if !ok { continue }`, or credited to whichever earlier dispatch still had an open window. A
subagent dispatch runs for minutes; a harvest tick is seconds. **The offset commit is what makes the
begin mandatory** — it is not an optimisation, it is the only moment the row can exist in time to be
found.

The `end` half is mandatory for the mirror-image reason. `ended_at` NULL is an OPEN window, and an
open window contains every later instant forever; before this decision no production path set the
column at all, so every dispatch this table ever recorded went on claiming the usage of everything
that came after it.

**The pair needs a name the caller can reproduce, so a dispatch carries a key.** `end` cannot name
its row by `seq`: the store allocates the seq, and a `begin` that fell back to the journal returned
none. The key is written by the dispatcher as a literal, unique within the run its session token
names, and `(change, session_token, dispatch_key)` is UNIQUE — which makes `begin` **idempotent**,
the third property this decision buys. A record write that could not reach the store is journalled
and replayed, and a lost response is indistinguishable from a store never reached; without a key to
collide on the replay allocated a fresh seq and inserted a second row for one logical dispatch,
double-counting its cost. `UpsertFinding` and `SetFindingStatus` were already idempotent; this was
the write whose duplicate costs money. The journal stores the request as built, so a replay
reproduces the key byte for byte — nothing in it is derived from the clock or from the store's own
allocation.

**One consequence is deliberate and narrow:** a 404 from `dispatch end` is journalled rather than
refused, where every other record 404 is definitive. "No dispatch under this key" has an ordinary
transient cause — the begin is still queued in the same journal ahead of it — and refusing it would
lose the end and leave the window open, reintroducing the defect by a narrower route.

**Considered:** Keeping one call at close and having the harvester re-read behind its committed
offset — rejected: the offset is what makes a delta safe to compute twice, and re-reading behind it
would double-count every stage-grain figure to fix a dispatch-grain one. Recording the begin only
and deriving the end from the next dispatch's begin — rejected: it is wrong for the last dispatch of
a run, wrong across a gap in which nothing was dispatched, and it would make one dispatch's window
depend on another dispatch existing.

### Concurrent panel slots are separated by agent id, not by window

**ID:** `agent-id-separates-concurrent-slots`
**Status:** active
**Chosen:** `dispatches` carries a nullable `agent_id`, recorded by the dispatcher when its harness
exposes one, and attribution prefers a window whose agent id matches the record's own. With no id on
either side it falls back to the window rule unchanged.
**Considered:** Recording the limitation and shipping — rejected: a review panel dispatches its slots
concurrently against one parent session, so their windows overlap and every record in the overlap is
credited to whichever slot started last. The change's total stays correct, but the per-slot split is
exactly what KAN-212 and KAN-218 are about, and this repository already ships a panel-economics view
that would read it. Serialising the panel so windows never overlap — rejected: it costs roughly three
times the panel's wall-clock to buy attribution precision, the opposite of what the roster presets
exist to control.

**The fallback is not a legacy path.** Cursor and Codex expose no agent id at all, so the window rule
remains the normal path on two of the three supported harnesses and must keep working. An absent id
is `""` and means "not reported"; it never matches another absent id.

**The tie-break is made stable at the same time.** At an exact `StartedAt` tie, "whichever started
last" is not a rule — the comparison is false in both directions — and the windows query had no
secondary sort key, so which dispatch won was not reproducible across harvest cycles. `ORDER BY
d.started_at, d.id` fixes it, and the comment now says what happens at a tie rather than asserting a
temporal order that does not exist there.

### The SPA is untouched

**ID:** `no-spa-in-this-change`
**Status:** active
**Chosen:** API and CLI only. `GET /api/v1/records/{project}/{change}` ships; nothing under
`stats/web/` changes.
**Considered:** Extending `RunDetail.tsx` with dispatch and finding panels; adding a cross-change
findings view. Both were offered and declined.

## Open questions

*(none — every question raised at the design gate was answered)*

## Design

### 1. Schema — `0010_run_records.sql`

`dispatches`, one row per subagent dispatch:

| Column | Type | Notes |
|--------|------|-------|
| `id` | `BIGSERIAL PRIMARY KEY` | |
| `change_id` | `BIGINT NOT NULL REFERENCES changes(id)` | |
| `stage_run_id` | `BIGINT REFERENCES stage_runs(id)` | nullable — a dispatch outside a marked stage still records |
| `seq` | `INT NOT NULL` | append order within the change |
| `task_id` | `TEXT` | dotted id from the `tasks.md` heading; NULL where the role runs against no task |
| `role` | `TEXT NOT NULL` | `implementer` · `reviewer` · `panel-fix` · `red-partner` |
| `slot` | `TEXT` | panel slot name; NULL for implementers |
| `model` | `TEXT NOT NULL` | recorded intent, or the literal `unknown (agent-defined)` |
| `commit_sha` | `TEXT` | |
| `outcome` | `TEXT` | |
| `session_token` | `TEXT` | the run's own token, for harvest binding |
| `started_at` | `TIMESTAMPTZ NOT NULL` | |
| `ended_at` | `TIMESTAMPTZ` | |
| `metrics` | `JSONB NOT NULL DEFAULT '{}'` | the harvester's bag, same shape as `stage_runs.metrics` |
| `notes` | `TEXT` | |

`CONSTRAINT dispatches_seq_key UNIQUE (change_id, seq)`, named explicitly rather than left to
Postgres' default identifier, exactly as `stage_runs_attempt_key` is — the constraint name is what
`RecordDispatch` checks for when it retries a lost race. Indexes on `(change_id)`, `(model)`,
`(role)`, `(task_id)`, `(started_at)`, plus a GIN index on `metrics`.

`findings`, one row per review-panel finding:

| Column | Type | Notes |
|--------|------|-------|
| `id` | `BIGSERIAL PRIMARY KEY` | |
| `change_id` | `BIGINT NOT NULL REFERENCES changes(id)` | |
| `dispatch_id` | `BIGINT REFERENCES dispatches(id)` | the slot that raised it |
| `ref` | `TEXT NOT NULL` | `F1`, `F2`, … |
| `round` | `INT NOT NULL` | `0` = the initial panel; `1..n` = fix rounds |
| `slot` | `TEXT NOT NULL` | |
| `severity` | `TEXT NOT NULL` | |
| `location` | `TEXT` | |
| `note` | `TEXT NOT NULL` | |
| `status` | `TEXT NOT NULL` | |
| `reproducer` | `TEXT` | the command, or `none — <reason>` |

`findings` are read back in **natural** ref order — the numeric part of `ref` compared as a number,
with a ref carrying no digits sorted last by `ref` — not lexically. A plain `ORDER BY ref` returns
`F1, F10, F2`, so a ten-finding panel would render out of order; this was found by task 1+2's
implementer and is corrected in task 4 rather than left to the renderer.

`dispatch_id` is written from the finding's own `DispatchSeq`, resolved against `(change_id, seq)` in
the same statement. A finding no single dispatch raised leaves it NULL, which is legitimate; a
column that could never be written would not be.

`CONSTRAINT findings_ref_key UNIQUE (change_id, ref)` — `ref` is unique **per change, not per
round**: a fix round updates a row's `status` in place. Nothing accumulates, which is what makes
KAN-65's defect (a cumulative record can never satisfy its own count guard) unreachable rather than
merely fixed. Indexes on `(change_id)`, `(severity)`, `(status)`, `(slot)`.

### 2. The write path

```text unverified:the flag set is fixed by this design; the command does not exist yet
myflow record dispatch begin -change <name> [-task <id>] -role <role> [-slot <s>] -model <m> \
                        [-agent-id <id>] -key <key> -session-token <tok> \
                        -started-at <ts> [-C <dir>]
myflow record dispatch end   -change <name> -key <key> -session-token <tok> \
                        [-commit <sha>] [-outcome <o>] -ended-at <ts> [-C <dir>]
myflow record finding    -change <name> -ref F<n> [-round <r>] -slot <s> -severity <sev> \
                        [-location <l>] [-dispatch-seq <n>] -status <st> [-reproducer <cmd>] -note <text>
myflow record status     -change <name> -ref F<n> -status <st>
myflow record render     -change <name> -kind ledger|panel|all -repo <abs-repo-root>
```

~~**One call per dispatch, at close.** The parent already waits for a dispatch's commit sha before
dispatching the next implementer into the same worktree, so at close it knows the task, the role, the
model, the commit and the outcome. A begin/end pair would double the call count and buy nothing.~~

**Superseded by `dispatch-begin-and-end` above** (this change's own final review panel, findings F6,
F7 and F8). The struck paragraph is kept rather than deleted: it reasoned entirely about what the
dispatcher knows and never about the harvester, whose committed offset means a row written only at
close is invisible to every harvest tick the dispatch ran through. The write path is a begin/end
pair, keyed by a caller-supplied dispatch key that also makes the begin idempotent under replay.

There are also two `POST` routes for a dispatch rather than one, and a dispatch end's 404 is the one
record refusal the CLI journals — both stated in the superseding decision.

Routes, all scoped by project and change so no body carries an identifier the path could:

```text unverified:the routes are fixed by this design; they are not registered yet
POST   /api/v1/records/{project}/{change}/dispatches
POST   /api/v1/records/{project}/{change}/findings
PATCH  /api/v1/records/{project}/{change}/findings/{ref}
GET    /api/v1/records/{project}/{change}
```

The wire types live in one package, `stats/internal/records`, imported by `internal/api`,
`internal/client` and the renderer alike — four consumers of one shape, so a shared type is the
simple choice here rather than a speculative abstraction. That package carries the Markdown renderer
too: rendering is a pure function of those types, and splitting it into its own package would buy an
import edge and nothing else.

**Store unreachable → journal, warn once, exit 0.** `fallback.AppendJournalEntry` is reused
unchanged; the journal file is `<state-dir>/<project-key>/<name>.journal.record`, following
`stage.go`'s existing `.journal.stage` suffix. `internal/reconcile` gains `replayRecordFile` beside
`replayFile` and `replayStageFile`. A caller **never** branches on `myflow record`'s exit code as a
signal about the record: a write that could not reach the store still exits 0.

**Caller mistakes still exit non-zero** — an unknown role, a missing required flag, or a
`-session-token` carrying a shell substitution. That is the same split `stage begin` draws, and
`check-stage-mark-calls.sh`'s substitution check extends to `myflow record` call sites in skill
source.

### 3. Cost attribution

`internal/harvest` already attributes usage to a `Window` — `(session_id, [started_at, ended_at))`,
half-open — and already splits main-thread from sidechain and buckets per model. A dispatch is a
sidechain inside a stage's window, so attribution adds a **second, independent pass** against
dispatch windows.

`stage_runs.metrics` is written exactly as today and its values do not change; `dispatches.metrics`
is new. The two are different grains over the same usage, not a double count.

`WindowSource` gains a sibling `DispatchWindowSource` rather than a kind discriminator on `Window`:
the consumer-side interface stays one method, and `internal/harvest` keeps importing nothing from
`internal/store`.

### 4. Rendering

| Kind | Destination |
|------|-------------|
| `ledger` | `<repo>/docs/superpowers/ledgers/<YYYY-MM-DD>-<change>.md` |
| `panel` | `<repo>/docs/superpowers/reviews/<YYYY-MM-DD>-<change>-panel.md` |

The date is fixed at the **first** render for a change: an existing file for that change is reused,
so a fix round overwrites in place rather than leaving one dated duplicate per round. That rule is
carried over verbatim from `preserve-session-records.sh` and is the only thing about it that
survives.

**The guards read the panel record at its rendered destination**, not in the worktree's
`.superpowers/sdd/`. Once the findings are store rows, nothing writes a findings table or a marker
block into `final-review-panel.md` — that file survives as the pass log alone (mode, slots, diff path,
`fix-mutation:` lines, bounces), and it carries no `findings-total:` line. A guard left reading it
would report every change OUTSTANDING at finish run 1 and fail every fix round. Both guards therefore
resolve `<worktree>/docs/superpowers/reviews/*-<change>-panel.md`, anchored.

**Writing both paths was rejected.** It would put the same record in two places that can disagree,
which is the problem this change exists to remove, and `docs/superpowers/reviews/` is where finish
run 1 commits it from in any case. The marker **contract** is unchanged by any of this; only the
**location** moves.

Firing points: the **panel record at panel close**, before `check-unfinished-work.sh` runs; the
**ledger at finish run 1**, where preservation used to happen; **either on demand**, via the command.

**`-kind panel` always writes; `MISSING:` never applies to it.** A panel that raised nothing produces
no finding rows, and reporting `MISSING:` for it would leave no record — which
`check-unfinished-work.sh` reads as OUTSTANDING for a change that is genuinely clean.
**The panel record declares how many findings it carries**
(`openspec/specs/myflow-review-panel-economics/spec.md`) already settles this: a panel that raised no
finding says so with `findings-total: 0`, which is a declaration and clears, where silence is not. The
command is invoked at panel close, so the invocation is itself the evidence a panel ran; no sentinel
row is stored for it. `MISSING:` keeps its meaning for `-kind ledger`, where a change with no dispatch
rows genuinely has no ledger.

Outcome words, one line per kind:

| Word | Means | Exit |
|------|-------|------|
| `rendered: <dest>` | rows existed and the file was written | 0 |
| `MISSING: <kind> — no rows for <change>` | the store holds nothing of this kind | 0 |
| `journalled: <kind>` | the store was unreachable; nothing was written | 0 |
| a message on stderr | the destination was refused or could not be written | non-zero |

Non-zero keeps the single meaning it has today — a write attempted and refused or failed — so a
caller branching on exit status never reads an empty record as a failure.

**The path protections move with the duty, into Go.** The change name is validated against the same
allowlist (one leading alphanumeric, then letters, digits, `.`, `_`, `-`), and the destination is
required to be contained within the repository root — enforced by the renderer rather than by an
accident of shell string concatenation.

**An unmeasured dispatch is distinguished by key presence, never by byte length.** `insertDispatch`
defaults an empty metrics bag to the JSON object `{}` — two bytes, never zero-length and never SQL
NULL — so a reader testing `len(raw) == 0` sees a measured zero where there was no measurement at
all. That is the state of every dispatch between its record and the harvester's next pass, and the
**permanent** state on Cursor and Codex, which write no transcript. The renderer therefore reads
whether the `tokens` key is present; a present `tokens` object renders whatever it carries, explicit
zeros included.

`omitempty` does not fire on `Metrics` for the same reason, so `"metrics":{}` is on the wire for every
unattributed dispatch. That is harmless to the renderer, which reads key presence — but a future
consumer testing that field by byte length inherits the same trap, which is why it is written down
here rather than left to be rediscovered.

### 5. What the retirement touches

- **`pipeline.md`** — the "Preserving the session records" section and its outcome table are replaced
  by the render-outcome table above; the `Panel record` and `SDD ledger` rows in the **Temporary
  artifacts registry** change their "Removed by" column, since neither is preserved out of the
  worktree any more; **Model policy**'s sentence naming the preserved ledger as the audit trail is
  repointed at the store.
- **`finish-contract.md`** — run 1's preservation duty becomes a render duty.
- **`handoff-blocks.md`** — the `IN_PROGRESS` block gains the `Records:` line, so `/myflow-status`
  renders it as well as the command that produced it. `myflow-fast/SKILL.md`'s own
  `IN_PROGRESS`-with-no-artifact block — the one shape no other skill prints — gains it alongside.
- **`myflow-do/SKILL.md`** — sections 4, 5 and 7: ledger and panel-record *writing* becomes
  `myflow record …` calls; the `prUrl` push path's preservation call becomes a render call; section
  7's `test -f` ledger assert becomes a store query. The paragraph stating that
  `superpowers:subagent-driven-development`'s workspace script writes the ledger to
  `.superpowers/sdd/tasks/progress.md` is removed — the parent records each dispatch instead.
- **`myflow-finish/SKILL.md`** — run 1's `finish.preserve-sessions` stage keeps its key; only what it
  invokes changes.
- **Guard-presence lists** in `myflow-do`, `myflow-finish` and `myflow-fast`; `setup.sh`'s symlink
  set; `check-guard-symlinks.sh`; `check-contract-budget.sh`'s budget rows for every contract file
  edited above.

### 6. Testing

- **Migration** — `0010` applies twice idempotently against `postgres:18-alpine`, beside the existing
  migration tests.
- **Store** — `RecordDispatch` allocates `seq` under concurrency without overwriting, the same
  concurrent-insert shape `BeginStage` is already tested with; `UpsertFinding` updates status in
  place; `MergeDispatchMetrics` deep-merges rather than replacing.
- **Fallback** — a record write with the daemon down journals to `<name>.journal.record` and exits 0;
  `replayRecordFile` replays it.
- **Harvest** — a sidechain record inside a dispatch window lands in `dispatches.metrics`, and
  `stage_runs.metrics` for the same session is byte-identical to what the existing pass produces.
- **Render** — a change with rows renders both kinds; a change with none prints `MISSING:` and exits
  0; a change name carrying `/` or a glob metacharacter is refused; a destination outside the repo
  root is refused; a second render for the same change reuses the first render's date.
- **Guard** — `check-unfinished-work.sh` passes unchanged against a rendered panel record, proving
  the marker block survived the move.
- **Removal** — `check-guard-symlinks.sh` and `check-references.sh` pass with the retired script
  gone, and no skill still names it.
