> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** Make a stage run bind to the session whose usage it should receive, so the statistics the
store already computes have something to compute over. Today every stage run has `session_id` NULL,
so every token and currency figure is empty.

**Architecture:** The calling skill writes a literal unique nonce into `stage begin`. That command
text lands in the calling session's own transcript. The daemon finds it there and binds the run's
`session_id` to that transcript's own `sessionId`, once, within a bounded window. Attribution after
binding is unchanged. Full reasoning in `design.md`.

## Global Constraints

- **The pipeline never blocks on this.** A mark that cannot reach the store, and a nonce that cannot
  be resolved, both leave the run recorded and the command exiting 0.
- **`internal/store` is the only package that builds SQL.**
- **No metric is ever recorded as zero to stand in for "not measured."** This change adds a third
  arm — *recorded, not measured* — and must not collapse it into either neighbour.
- **A session is never guessed.** Ambiguity records no session; it never picks one.
- **Binding is one-way.** A bound stage run is never re-bound, so each message is attributed at most
  once and the exactly-once harvest semantics are untouched.
- **Stage names come from `README.md`'s Level 1 table.** No skill invents one.
- **No task edits `openspec/` or `docs/superpowers/`** — `/myflow-finish` commits those.

## Baseline

**Measured on 2026-08-14 against commit `88db0a2`:** 296 top-level Go test functions
(`cd stats && go test ./... -count=1 -v | grep -c '^--- PASS'`) and 110 SPA tests
(`cd stats/web && npm test`).

<!-- measured: both commands run against 88db0a2 on 2026-08-14 -->

The live store held 1 change and 2 stage runs, both with `session_id` NULL and empty metrics, against
2,988 consumed transcript offsets.

<!-- measured: read from the running myflow-postgres on 2026-08-14 -->

| Measure | Now | After |
|---------|-----|-------|
| Stage runs with a session | 0 | every run marked on a harness with a transcript |
| `/myflow-fast` discrete stages | 2 | the union of the three commands it chains |
| `harness` on a recorded run | `unknown` | the harness the skill names |

---

### 1 The mark carries a nonce and a harness, and the store keeps them

**Build:** green

**Files:**
- Add: `stats/internal/store/migrations/0008_stage_run_session_nonce.sql`
- Modify: `stats/cmd/myflow/stage.go`
- Modify: `stats/cmd/myflow/stage_test.go`
- Modify: `stats/internal/api/stages.go`
- Modify: `stats/internal/api/stages_test.go`
- Modify: `stats/internal/store/stageruns.go`
- Modify: `stats/internal/store/stageruns_test.go`
- Modify: `stats/internal/client/client.go`
- Modify: `stats/internal/client/client_test.go`
- Modify: `stats/internal/reconcile/reconcile.go`
- Modify: `stats/internal/reconcile/stage_test.go`

The last two were not anticipated and are recorded after the fact. `reconcile.go` rebuilds a begin
mark from a journalled request when replaying the write-ahead fallback; without forwarding the new
nonce, **every replayed mark would be rejected by this task's own validation**, silently breaking the
mechanism that exists so the pipeline never blocks. That is a correctness gap this task opens, not
test maintenance.

**Interfaces:**
- Consumes: the existing `stage begin` path end to end.
- Produces: `stage_runs.session_nonce`, and `-nonce`/`-harness` on the CLI.

- [x] **Step 1: The column and the flags**

Migration `0008` adds a nullable `session_nonce TEXT` and a partial index on it restricted to rows
where it is non-null and `session_id` is null — the only rows the resolver ever scans for. It also
adds a nullable `stage_key TEXT` with an index, for task 3's vocabulary.

**`stage_key` is nullable and the two existing rows are left alone.** They predate the vocabulary and
carry only a prose stage; backfilling a key onto them would be inventing history.

`stage begin` gains `-nonce` (required) and keeps `-harness`, which is already declared but never
passed by any caller. **A missing `-nonce` is a caller mistake, not a stage outcome**: exit non-zero
naming the flag, exactly as an undocumented stage name already does. That is the one class of
nonzero exit the never-block rule permits.

- [x] **Step 2: Reject a nonce that cannot identify anything**

A nonce containing `$(`, a backtick, or `$` followed by a name is **rejected** with a message saying
why: the transcript records the command before the shell expands it, so such a value is recorded
identically by every caller and identifies nothing. See `design.md`'s second decision.

Tests: a literal nonce is accepted and stored; each of the three substitution shapes is rejected with
its own case; a missing nonce exits non-zero; the store round-trips the nonce with the run.

**Tests:** `stage_test.go` covers the flags and the rejections; `stageruns_test.go` covers
persistence; `stages_test.go` covers the wire shape. Verification is `cd stats && go test ./...
-race -count=1`, `gofmt -l .`, `go vet ./...`.

**Regression:** Reverting this leaves the store with no correlator to bind on, and the whole change
inert.

**Baseline:** before=296 after=296+ Go top-level tests.
<!-- predicted: the new cases in the four test files this task names, none written yet -->

**Commit:** `feat(1): carry a session nonce and a harness on every stage mark`

---

### 2 The harvester binds a nonce to its session, once, within a bounded window

**Build:** green

**Files:**
- Modify: `stats/internal/harvest/watcher.go`
- Modify: `stats/internal/harvest/watcher_test.go`
- Modify: `stats/internal/harvest/transcript.go`
- Modify: `stats/internal/harvest/transcript_test.go`
- Modify: `stats/internal/store/stageruns.go`
- Modify: `stats/internal/store/stageruns_test.go`

**Interfaces:**
- Consumes: task 1's `session_nonce`.
- Produces: a bound `session_id`, after which existing attribution runs unchanged.

- [x] **Step 1: Find the nonce in the transcripts**

The transcript reader learns to report a recorded command's text (`tool_use.input.command`) alongside
what it already reads. Each harvest cycle, unresolved nonces are looked for across the transcripts
the cycle is already reading — **not** a separate full scan, since the watcher opens them anyway.

- [x] **Step 2: Bind, or refuse**

Exactly one transcript containing the nonce binds that run's `session_id` to that transcript's own
`sessionId`. **Zero matches leaves it unresolved. More than one records no session and reports the
ambiguity** — never picks. A run already carrying a `session_id` is never re-bound.

The test that matters: two transcripts, two runs, nonces crossed — each run binds to its own session
and neither receives the other's. Under a resolver that picked the newest transcript this test fails,
which is what makes it a real test of the mechanism rather than of the code as written.

- [x] **Step 3: Give up, bounded**

A nonce unresolved after a bounded number of cycles (or a wall-clock window — pick one and say which
in the code) stops being looked for, and its run stays recorded and unattributed. Tests: a nonce that
appears on a later cycle still binds; one that never appears stops being scanned and the run remains
unattributed; the give-up path logs once rather than every cycle.

**Tests:** `watcher_test.go` covers binding, the crossed-nonce case, the ambiguity refusal and the
bounded give-up; `transcript_test.go` covers reading the recorded command text. Verification as task
1.

**A batch that matched a pending nonce is withheld, not committed.** Attribution runs against windows
as they stand when the batch is read; the binding happens after the file loop. Committing such a
batch would advance the offset past records that produced no delta, and nothing re-reads a committed
offset — so the usage would be lost, not delayed. Worse, the `stage begin` command reaches the
transcript at *turn end*, in the same flush as that whole turn's messages, so the affected batch is
routinely the largest one a stage has. Withholding costs one re-read; the failure it prevents is
silent and permanent.

**Regression:** Reverting this records nonces nobody reads, leaving every run unattributed exactly as
today.

**Baseline:** before=296+ after=296++ Go top-level tests.
<!-- predicted: the new cases in the three test files this task names, none written yet -->

**Commit:** `feat(2): bind a stage run to the session whose transcript carries its nonce`

---

### 3 Every stage gets a stable key and a readable name, and `/myflow-fast` gets real stages

**Build:** green

**Files:**
- Modify: `README.md`
- Modify: `stats/internal/stages/names.go`
- Modify: `stats/internal/stages/names_test.go`
- Modify: `stats/cmd/myflow/stage.go`
- Modify: `stats/cmd/myflow/stage_test.go`
- Modify: `stats/internal/api/stages_test.go`
- Modify: `stats/internal/reconcile/stage_test.go`

The last two were not anticipated and are recorded after the fact: both hardcoded the prose stage
name `SDD + TDD per task`, which stops validating once marks carry keys. Mechanical substitution, no
logic change.

**Interfaces:**
- Consumes: the README Level 1 table, which stays canonical.
- Produces: a `/myflow-fast` stage set composed from the commands it chains.

**The audit, already run** — splitting each row on its arrow separator:

| Command | Discrete stages |
|---------|-----------------|
| `/myflow-start` | 8 |
| `/myflow-do` | 11 |
| `/myflow-finish` | 13 |
| `/myflow-fast` | **2** |
| `/myflow-status` | marks nothing, correctly |

<!-- measured: awk over README.md's Level 1 table on 2026-08-14 -->

Only `/myflow-fast` is defective; the audit is recorded here so the conclusion is not re-derived.

- [x] **Step 1: The table gains keys**

`README.md`'s Level 1 entry stops being arrow-separated prose and becomes a real table: **key, name,
and the commands that run it**. Keys are short, lowercase and dotted, namespaced by the command that
*defines* the stage — `do.review-panel`, `start.brainstorm`, `finish.integrate` — and unique across
the whole pipeline.

**Name them so a stranger can read a dashboard.** `do.review-panel` beats `do.stage-5`; the key
appears in shell invocations, guard output and the group-by, so it is read far more often than it is
written. Keep each under about 30 characters.

`names.go` parses key, name and commands. Its existing README-parsing test governs, and gains a
**uniqueness assertion** — a duplicate key silently merges two stages in every statistic, so it must
fail loudly.

- [x] **Step 2: Compose rather than duplicate**

`/myflow-fast`'s allowed set is the **union** of `/myflow-start`'s, `/myflow-do`'s and
`/myflow-finish`'s, expressed in the table's Commands column rather than by restating names. A fast
run records `do.review-panel`, not `fast.review-panel`, so the two are one stage in the data.
Reasoning and the rejected alternatives are in `design.md`.

- [x] **Step 3: Marks pass the key**

`stage begin`/`stage end` take the key. The prose name stops being an identifier, which removes
backticks and apostrophes from every mark invocation — `-stage 'run the project'"'"'s lint and test
commands'` becomes `-stage do.lint-and-test`.

- [x] **Step 4: The test that would catch the next one**

A test asserting that **no command's documented stage set contains an entry longer than a plausible
stage name** — the defect here was a 400-character sentence sitting where a stage name belongs. Pick
the bound from the longest legitimate name plus headroom, and say in a comment what it is guarding
against, not merely what it does.

**Tests:** `names_test.go` gains the composition assertion and the length guard; its existing
README-parsing test is unchanged and still governs. Verification as task 1.

**Regression:** Reverting this returns `/myflow-fast` to recording one stage per run, making its
per-stage statistics meaningless even with attribution working.

**Commit:** `feat(3): identify each stage by a stable key and a readable name`

---

### 4 Every mark call passes a nonce and a harness

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md`
- Modify: `skills/myflow-start/SKILL.md`
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-finish/SKILL.md`
- Modify: `skills/myflow-fast/SKILL.md`
- Add: `scripts/check-stage-mark-calls.sh`
- Add: `scripts/test-check-stage-mark-calls.sh`
- Modify: `.myflow/project.md`

**Interfaces:**
- Consumes: tasks 1 and 3.
- Produces: mark calls that can actually be attributed.

- [x] **Step 1: The contract changes from *may* to *must***

`pipeline.md`'s **Stage marks** section: a mark carries a nonce and a harness. State plainly that the
nonce is a literal the command writes, and why a shell substitution cannot work — a reader who does
not know that will reintroduce the defect.

- [x] **Step 2: Every call site**

All four marking skills pass `-nonce <literal>` and `-harness <harness>`. `/myflow-status` marks
nothing and is untouched.

- [x] **Step 3: A guard, because prose does not enforce itself**

`check-stage-mark-calls.sh` fails when a `myflow stage begin` in any skill omits `-nonce` or
`-harness`, or when its nonce contains a `$(`, a backtick or a `$VAR`. Added to `## lint` in
`.myflow/project.md`.

**This guard is the deliverable of the task**, not the edits: four skills currently pass neither
flag, and nothing noticed for the life of KAN-16.

**Tests:** `test-check-stage-mark-calls.sh` covers a compliant call, each omission, and each
substitution shape. Verification is that harness plus the repository's existing guards.

**Regression:** Reverting this leaves the flags unpassed and every run unattributed, with the code
from tasks 1–3 fully working and unused.

**Commit:** `feat(4): pass a nonce and a harness from every stage mark`

---

### 4b Rework the correlator from per-mark to per-session

**Build:** green

**Added mid-implementation at the operator's instruction** — *"just pass it to all commands to
identify the session"* — after tasks 1–4 had been built on a per-mark nonce. Recorded as its own task
rather than folded into 1–4, so the change of shape is visible in the history instead of rewritten
out of it.

**Files:**
- Modify: `stats/internal/store/migrations/0008_stage_run_session_nonce.sql`
- Modify: `stats/internal/store/stageruns.go`
- Modify: `stats/internal/store/stageruns_test.go`
- Modify: `stats/cmd/myflow/stage.go`
- Modify: `stats/cmd/myflow/stage_test.go`
- Modify: `stats/internal/api/stages.go`
- Modify: `stats/internal/api/stages_test.go`
- Modify: `stats/internal/client/client.go`
- Modify: `stats/internal/client/client_test.go`
- Modify: `stats/internal/reconcile/reconcile.go`
- Modify: `stats/internal/harvest/watcher.go`
- Modify: `stats/internal/harvest/watcher_test.go`
- Modify: `skills/myflow-contracts/pipeline.md`
- Modify: `skills/myflow-start/SKILL.md`
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-finish/SKILL.md`
- Modify: `skills/myflow-fast/SKILL.md`
- Modify: `scripts/check-stage-mark-calls.sh`
- Modify: `scripts/test-check-stage-mark-calls.sh`
- Modify: `stats/internal/reconcile/stage_test.go`

The last was not anticipated and is recorded after the fact: it carries the request field the rename
touches, so it stops compiling otherwise. Same shape as task 1's own unanticipated reconcile edit.

**Interfaces:**
- Consumes: tasks 1-4 as built.
- Produces: one correlator per run, reused across every mark that run makes.

- [x] **Step 1: Rename, because a reused value is not a nonce**

`-nonce` becomes `-session-token`; `session_nonce` becomes `session_token`. Migration `0008` is
amended rather than superseded — it is unreleased and in-branch, so a second migration renaming a
column nothing has ever written would be ceremony. **Calling a deliberately-repeated value a nonce
would mislead every future reader about whether repetition is a bug.**

- [x] **Step 2: One token per run, reused**

Each command generates one token at the start of its run and passes it on every mark. The store binds
**every** stage run carrying a token when that token resolves — not just the one whose mark revealed
it — and a mark arriving with an already-bound token is bound at write time, with no resolution work
at all.

- [x] **Step 3: Withhold only while unbound**

Task 2's withholding is now needed only for a batch carrying an **unbound** token. Once bound,
nothing is withheld. This is the point of the change: per mark it cost one withheld re-read per
stage, roughly 34 a run; per session it costs one, ever.

The test that pins it: a run whose token is already bound makes a further mark, and its batch is
committed in the same cycle rather than withheld. Under the per-mark shape this test fails.

- [x] **Step 4: The guard follows**

`check-stage-mark-calls.sh` checks `-session-token` rather than `-nonce`, and keeps both existing
rules — no shell substitution in the token, no hardcoded harness literal.

**Tests:** the existing nonce tests are rewritten against the new shape rather than duplicated; new
cases cover a second mark binding immediately, and a batch not being withheld once the token is
bound. Verification is `cd stats && gofmt -l . && go vet ./... && go test ./... -race -count=1`, plus
the guard and its harness.

**Regression:** Reverting this returns to a correlator per mark — correct, but paying a withheld
re-read for every stage to re-establish a fact the run's first mark already settled.

**Baseline:** before=319 after=319+ Go top-level tests.
<!-- predicted: the rewritten and added cases in the files this task names, none written yet -->

**Commit:** `refactor(4b): identify the session once per run, not once per mark`

---

### 5 A recorded but unmeasured run is visible as such

**Build:** green

**Files:**
- Modify: `stats/internal/store/aggregate.go`
- Modify: `stats/internal/store/aggregate_test.go`
- Modify: `stats/internal/api/stats.go`
- Modify: `stats/internal/api/stats_test.go`
- Modify: `stats/web/src/api.ts`
- Modify: `stats/web/src/components/Panel.tsx`
- Modify: `stats/web/src/components/Panel.test.tsx`
- Modify: `stats/web/src/views/views.test.tsx`
- Modify: `stats/internal/client/client_test.go`
- Modify: `stats/internal/web/embed_test.go`
- Modify: `stats/web/src/views/RunDetail.test.tsx`

The last three were not anticipated and are recorded after the fact: each implements or constructs the
`StatsStore` interface or the `StatsResponse` type, so growing either stops them compiling.
Mechanical, no logic change.

**Interfaces:**
- Consumes: the existing `recorded` flag in the statistics envelope.
- Produces: a third state the interface can render.

- [x] **Step 1: The server can say it**

The statistics envelope distinguishes *no runs in this period* from *runs recorded, none measured*.
The existing `recorded` flag already carries the first; add the second rather than overloading it.

- [x] **Step 2: The interface renders it**

A period whose runs are all unattributed reads as **"runs were recorded, but none was measured"** —
not "no data was recorded", which is what it says today and which is why this defect stayed invisible
while the dashboards merely looked empty.

The test asserts all three states separately: no runs; runs but no metrics; a measured zero. **They
must not collapse into two** — that collapse is the defect.

**Tests:** `aggregate_test.go` and `stats_test.go` for the server; `Panel.test.tsx` and
`views.test.tsx` for the interface. Verification is the Go and SPA suites plus `npx tsc -b`.

**Regression:** Reverting this makes a recording misconfiguration indistinguishable from a quiet
period — the condition that let this change's defect go unnoticed through a five-pass review panel.

**Baseline:** before=110 after=110+ SPA tests; Go as task 2 leaves it.
<!-- predicted: the new cases in the four test files this task names, none written yet -->

**Commit:** `feat(5): distinguish a recorded but unmeasured run from an absent one`

---

### 6 Prove it end to end, against a real session

**Build:** green

**Files:**
- Add: `stats/internal/harvest/endtoend_test.go`
- Modify: `stats/README.md`

**Interfaces:**
- Consumes: everything above.
- Produces: the first test in this repository that would have caught the defect.

- [x] **Step 1: The test that was missing**

A test that marks a stage with a nonce, writes a synthetic transcript containing that nonce **in the
recorded-command position** along with usage entries, runs a harvest cycle, and asserts the run ends
up bound and carrying tokens and a cost.

**Write it first and watch it fail against `main`.** It fails there because nothing binds — which is
the whole defect, and the reason KAN-168 exists.

- [x] **Step 2: Say how to check it by hand**

`stats/README.md` gains a short section: mark a stage, wait a cycle, query the run, expect a session
and metrics. This defect was found by a human looking at an empty dashboard and asking why; the
README should make that check a minute's work rather than an investigation.

**Tests:** `endtoend_test.go` is the test. Verification is the full Go suite, plus one live check
against a running daemon.

**Regression:** Reverting this removes the only test that exercises mark → transcript → binding →
attribution as one path.

**Commit:** `test(6): prove a marked stage is attributed end to end`

---

### 7 The daemon never wires the binder, so the whole change is inert

**Build:** green

**Found by running it**, after the review panel's Primary slot returned `findings-total: 0`. Not by
any test: 329 Go tests, 112 SPA tests and nine guards were all green while the live daemon bound
nothing.

**Files:**
- Modify: `stats/cmd/myflowd/main.go`
- Add: `stats/cmd/myflowd/wiring_test.go`
- Modify: `stats/internal/harvest/watcher.go`
- Modify: `stats/internal/harvest/transcript.go`
- Modify: `stats/internal/harvest/transcript_test.go`

**Interfaces:**
- Consumes: tasks 1–6, all of which work and none of which is reachable.
- Produces: a daemon that actually binds, and a test that fails if it stops.

**The defect.** `main.go:172` builds the watcher as
`harvest.NewWatcher(transcriptsRoot, st, attributor, logger, harvest.WithPricer(st))` — with no
`WithSessionTokenBinder(st)`. The binder is a functional option, so it stays nil, so
`pendingSessionTokens` returns nothing, so no token is matched and nothing is ever bound. `main.go`
appears in **no** task's `**Files:**` declaration and was never touched.

<!-- verified: reproduced against the running daemon on 2026-08-14 — a real mark recorded session_token mf-k172-live-7f3a91c, the token appears in the session transcript in tool_use.input position, the transcript was harvested to EOF, and the stage run stayed unbound -->

**This is KAN-16's `Price` defect, one change later.** A component correct in isolation, tested
thoroughly, wired to nothing. The panel could not see it because a reviewer reads a **diff**, and an
absent call site is not in one. `docs/self-review/kan-16-myflow-stats-app-self-review.md` names this
exact class and [KAN-169](https://tweety53.atlassian.net/browse/KAN-169) was filed for it this
morning; it recurred before that follow-up was picked up.

- [x] **Step 1: Wire it**

Pass `harvest.WithSessionTokenBinder(st)`. `*store.Store` already implements
`UnresolvedSessionTokens` and `BindSession`, so this is one argument.

- [x] **Step 2: The test that makes the wiring assertable**

**This is the deliverable, not step 1.** A test that fails when the daemon does not wire the binder.
Extract the watcher's construction into a small function `main.go` calls and the test can call —
then assert the constructed watcher has a binder. Do not assert by reading `main.go`'s source text;
assert the constructed value, so a refactor that keeps the text and drops the effect still fails.

Do the same for the pricer while you are there: it is wired today, and nothing would notice if it
stopped being.

- [x] **Step 3: Finish the rename the panel found half-done**

`watcher.go:229`'s parameter is still `nb` (nonce-binder) and its doc comment reads "session
sessionTokens" — a mechanical substitution that produced a doubled phrase.
`transcript.go` and `transcript_test.go` still say `nonce` throughout, including `-nonce mf-abc123`
fixture strings naming a flag that no longer exists. Rename them; a maintainer reading
`transcript.go` currently learns the wrong flag name.

**Tests:** `wiring_test.go` asserts the daemon's watcher carries both a session-token binder and a
pricer. Verification is `cd stats && gofmt -l . && go vet ./... && go test ./... -race -count=1`,
plus a **live check**: mark a stage with a literal token, wait two harvest cycles, and confirm the
run is bound and carries metrics.

**Regression:** Reverting this returns the daemon to binding nothing, with every test still green —
which is the exact condition this task was found in.

**Baseline:** before=329 after=329+ Go top-level tests.
<!-- predicted: the new wiring cases, none written yet -->

**Commit:** `fix(7): wire the session-token binder into the daemon`
