# Tasks — make per-dispatch cost attribution true

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** stop attributing a concurrent group's whole spend to one dispatch, attribute by the exact
identifier rather than by a hand-typed interval, recover sessions the bounded search abandoned, and
say plainly where cost could not be attributed.

**Architecture:** Go in `stats/` — `internal/harvest` (attribution and the watcher),
`internal/store` (one migration, the give-up table, the unattributed stamp), `internal/records`
(the ledger's three states), `internal/api` and `cmd/myflow` (`-agent-id` on `end`, and a
`cost-status` read) — plus `skills/myflow-do/SKILL.md`, which captures the identifier and prints the
handoff line. No SPA source: this change adds no view.

**Spec:** `design.md` beside this file, and the two delta specs under `specs/`.

## Global constraints

- **Every task ends with `.myflow/project.md`'s `## lint` exiting clean**, including
  `cd stats && go vet ./... && gofmt -l .`. A hit is fixed by editing the offending line, never by a
  suppression marker or a weakened guard.
- **`internal/harvest` imports nothing from `internal/store`.** `TestHarvestNeedsNoDatabase` rests on
  it, and every new store capability this plan adds reaches the harvester through a consumer-side
  interface declared in `internal/harvest`, exactly as `DispatchWindowSource` already is.
- **The stage grain is untouched.** `Attributor`, `tokens.main`, `tokens.sidechain` and
  `models.<model>` keep their present meaning and their present values. Only the dispatch grain
  changes. `TestAttributeSplitsByAgentID` and `TestAttributeNoAgentIDCreatesNoDispatch` must still
  pass unchanged.
- **An unmeasured value is never written as zero.** The delta specs restate the rule; every task
  below that touches a metrics bag is bound by it.
- **The store's migrations are append-only.** Add `0013_…`; never edit an applied migration.

**Baseline, before any edit.** 57 dispatch rows across 5 changes, 20 carrying a `tokens` key.
<!-- measured: SELECT count(*), count(*) FILTER (WHERE metrics ? 'tokens') FROM dispatches @ 2026-08-23, branch main -->
Task 12 re-reads this figure after a restart and records what changed.

---

### 1 Attribution tests — identity beats interval, and ambiguity attributes to nothing

**Build:** red

**Squash-with:** Task 2

**Files:**
- Modify: `stats/internal/harvest/attribute_test.go`

**Tests:** `TestDispatchIdentityBeatsInterval`, `TestDispatchDuplicateAgentIDAttributesToNeither`,
`TestDispatchAmbiguousOverlapAttributesToNone`, `TestDispatchNonOverlappingIntervalStillAttributes`

**Regression:** Reverting task 2 leaves `TestDispatchIdentityBeatsInterval` failing (a record whose
id matches a dispatch outside that dispatch's window attributes to nothing) and
`TestDispatchAmbiguousOverlapAttributesToNone` failing (the latest-started candidate is credited).

**Baseline:** `go test ./internal/harvest` before=all pass after=4 new failures.

**Interfaces:**
- Consumes: `harvest.DispatchWindow`, `harvest.NewDispatchAttributor`, the existing
  `fakeDispatchWindowSource`.
- Produces: the executable statement of `bestDispatchWindow`'s new contract. Task 2 satisfies it.

Follow the idioms already in this file — `newDispatchRecord`, `fakeDispatchWindowSource`, assertions
on the returned `map[int64]TokenDelta`.

- [x] **Step 1: A record whose agent id matches a dispatch outside that dispatch's window attributes to it**

One window `10:00 → 10:05` recording `agent-one`; one sidechain record carrying `agent-one` at
`10:07`. Assert the delta lands on that dispatch. This is the `kan-286` case: `-ended-at` is typed by
the dispatching agent, so the record legitimately falls outside a window that is approximate by
construction.

- [x] **Step 2: Two dispatches recording the same agent id attribute to neither**

Two windows both recording `agent-one`, one record carrying `agent-one` inside both. Assert the
returned map is empty. A tie broken by interval or by row order would reintroduce the same silent
misattribution one layer down.

- [x] **Step 3: Overlapping windows with no id on either side attribute to no dispatch**

Three windows with byte-identical intervals and no agent id, one record with no agent id inside them.
Assert the returned map is empty — **not** that the latest-started wins. This is the `kan-295` case
in miniature.

- [x] **Step 4: A single containing window with no id still attributes**

One window, one record, neither carrying an id. Assert the delta lands. The interval rule is the
fallback, not a legacy path: it is the only rule available on Cursor and Codex, and it is exactly
correct wherever windows do not overlap.

- [x] **Step 5: Run and confirm they fail**

```bash verified:the package and its test binary exist today; `go test` is .myflow/project.md's ## test command for stats
cd stats && go test ./internal/harvest -run 'TestDispatch(Identity|Duplicate|Ambiguous|NonOverlapping)' -count=1
```

Expected: FAIL on steps 1, 2 and 3; step 4 passes already and is the guard against over-correcting.

---

### 2 Rewrite `bestDispatchWindow`

**Build:** green

**Files:**
- Modify: `stats/internal/harvest/attribute.go`

**Tests:** the four from task 1, plus the existing `TestAttributeSplitsByAgentID`,
`TestAttributeNoAgentIDCreatesNoDispatch`, `TestConcurrentSlotsAreSeparatedByAgentID`.

**Regression:** Reverting this task fails all three of task 1's new negative cases.

**Baseline:** `go test ./internal/harvest` before=4 failures after=all pass.
<!-- predicted: the count is the number of tests the paired red task adds; confirmed by task 12 -->

**Commit:** `fix(harvest): attribute a dispatch by its agent id, never by a typed interval`

**Interfaces:**
- Consumes: `[]DispatchWindow`, the record's `AgentID` and `Timestamp`.
- Produces: `(DispatchWindow, bool)` plus, for an ambiguous outcome, the candidate count task 4
  stamps.

The function becomes two passes, in this order:

1. **By identifier.** Collect every window whose non-empty `AgentID` equals the record's non-empty
   `AgentID`, ignoring the interval entirely. Exactly one → that window. More than one → ambiguous,
   attribute to none. An empty id on either side matches nothing, including another empty one.
2. **By interval**, reached only when pass 1 found no candidate. Collect every window that
   `contains(ts)`. Exactly one → that window. More than one → ambiguous, attribute to none.

- [x] **Step 1: Replace the single-pass loop with the two passes above**

Delete the latest-started tie-break and the `DispatchID` tie-break with it: both exist only to choose
among candidates that this change refuses to choose among. Keep `contains` and `intervalContains`
untouched — `Window.contains` shares them.

- [x] **Step 2: Return the ambiguity, do not swallow it**

Widen the return so a caller can tell "no candidate" from "too many candidates, N of them". Task 4
records the difference; a boolean alone cannot.

- [x] **Step 3: Rewrite the doc comment against what the code now does**

The present comment states the agent id is preferred *among windows whose interval contains the
record*, and gives the tie-break rules this task deletes. Replace it, and record why identity now
outranks the interval: the interval is written by the dispatching agent at mark time and the audit
found values rounded to the minute. Do not leave a stale sentence describing the old rule.

- [x] **Step 4: Run the package**

```bash verified:same command as task 1 step 5, unrestricted
cd stats && go test ./internal/harvest -count=1
```

Expected: PASS, including every pre-existing test.

---

### 3 Store tests — the give-up table and the unattributed stamp

**Build:** red

**Squash-with:** Task 4

**Files:**
- Modify: `stats/internal/store/records_test.go`
- Create: `stats/internal/store/giveups_test.go`

**Tests:** `TestRecordSessionTokenGiveUp`, `TestPersistedGiveUpsAreListed`,
`TestGiveUpIsIdempotent`, `TestMarkDispatchesUnattributed`

**Regression:** Reverting task 4 fails every one of them — the table and both methods are absent.

**Baseline:** `go test ./internal/store` before=all pass after=4 new failures.

**Interfaces:**
- Consumes: the store's existing test harness (`testsupport_test.go`).
- Produces: the executable statement of `RecordSessionTokenGiveUp`, `PersistedGiveUps` and
  `MarkDispatchesUnattributed`. Task 4 satisfies them.

- [x] **Step 1: A give-up is recorded and read back**

Record a give-up for one token with a reason; assert `PersistedGiveUps` returns it.

- [x] **Step 2: Recording the same token twice does not duplicate it**

The watcher may give up on the same token again after a restart-and-retry. Assert the second write
updates rather than inserts, and that the retry count rises.

- [x] **Step 3: `MarkDispatchesUnattributed` stamps every dispatch carrying a token**

Insert two dispatches under one session token and one under another; stamp the first token; assert
exactly two bags carry the unattributed key and the third is untouched.

- [x] **Step 4: The stamp never overwrites a real token figure**

Insert a dispatch, merge real token metrics into it, then stamp it. Assert `tokens` survives. A
dispatch that was measured and *then* had its session marked given up is a contradiction the store
must not resolve by destroying the measurement.

- [x] **Step 5: Run and confirm they fail**

```bash unverified:giveups_test.go does not exist until step 1 of this task creates it
cd stats && go test ./internal/store -run 'GiveUp|Unattributed' -count=1
```

Expected: FAIL — neither the table nor the methods exist.

---

### 4 Migration `0013`, the give-up table and the unattributed stamp

**Build:** green

**Files:**
- Create: `stats/internal/store/migrations/0013_session_token_giveups.sql`
- Create: `stats/internal/store/giveups.go`
- Modify: `stats/internal/store/records.go` — `MarkDispatchesUnattributed`

**Tests:** the four from task 3.

**Regression:** Reverting fails all four; `go test ./internal/store` cannot even reach the new
methods.

**Baseline:** `go test ./internal/store` before=4 failures after=all pass.
<!-- predicted: the count is the number of tests the paired red task adds; confirmed by task 12 -->

**Commit:** `feat(store): persist an abandoned session token and stamp its uncosted dispatches`

**Interfaces:**
- Consumes: the pool, `jsonb_deep_add` (migration `0005`).
- Produces: `RecordSessionTokenGiveUp`, `PersistedGiveUps`, `MarkDispatchesUnattributed`, consumed by
  task 6's watcher through an interface declared in `internal/harvest`.

- [x] **Step 1: Write the migration**

```sql unverified:the migration file does not exist until this step writes it
CREATE TABLE session_token_giveups (
  session_token TEXT PRIMARY KEY,
  reason        TEXT        NOT NULL,
  gave_up_at    TIMESTAMPTZ NOT NULL,
  retries       INT         NOT NULL DEFAULT 0
);
```

Carry a header comment in the idiom of `0011` and `0012`: what the table is for, why the token rather
than the stage run is the key (one token covers every mark a run makes), and why no index beyond the
primary key is added.

- [x] **Step 2: `RecordSessionTokenGiveUp` upserts**

`ON CONFLICT (session_token) DO UPDATE` setting the reason and instant and incrementing `retries`.
Step 2 of task 3 states the contract.

- [x] **Step 3: `PersistedGiveUps` lists them**

Return token, reason and retry count. This is what task 6 reads at start.

- [x] **Step 4: `MarkDispatchesUnattributed` merges the reason into the bag**

`UPDATE dispatches SET metrics = jsonb_deep_add(metrics, $2::jsonb) WHERE session_token = $1`, with a
payload shaped `{"unattributed": {"reason": "...", "candidates": N}}` — `candidates` omitted where the
reason is not an ambiguity. `jsonb_deep_add` merges recursively, so an existing `tokens` key survives,
which is exactly what task 3 step 4 asserts.

- [x] **Step 5: Run the package**

```bash verified:same command as task 3 step 5, unrestricted
cd stats && go test ./internal/store -count=1
```

Expected: PASS.

---

### 5 Watcher tests — persist the give-up, retry it, stamp the dispatches

**Build:** red

**Squash-with:** Task 6

**Files:**
- Modify: `stats/internal/harvest/watcher_test.go`

**Tests:** `TestGiveUpIsPersisted`, `TestPersistedGiveUpIsRetriedOnStart`,
`TestRetryStillBounded`, `TestAmbiguousDispatchIsStamped`

**Regression:** Reverting task 6 leaves the give-up in memory only — `TestGiveUpIsPersisted` fails,
and `TestPersistedGiveUpIsRetriedOnStart` fails because nothing re-enters the pending set.

**Baseline:** `go test ./internal/harvest` before=all pass after=4 new failures.

**Interfaces:**
- Consumes: the existing fake `SessionTokenBinder`, extended with the give-up methods.
- Produces: the executable statement of task 6's behaviour.

- [x] **Step 1: Exhausting the bounded window persists the give-up**

Drive `RunOnce` past `maxSessionTokenResolutionCycles` with a token that matches nothing; assert the
fake recorded a give-up naming the token and the reason.

- [x] **Step 2: Ambiguity persists a give-up too, with its own reason**

Two sessions matching one token. Assert the give-up's reason distinguishes it from the exhausted
case — they are different failures and the ledger tells them apart.

- [x] **Step 3: A persisted give-up re-enters the pending set at start**

Build a fresh `Watcher` over a binder already holding a persisted give-up whose transcript now
carries the mark; drive one cycle; assert it binds. This is `kan-302`'s recovery.

- [x] **Step 4: The retry is bounded exactly as the first attempt is**

Drive the retry past the bound with a token that still matches nothing; assert it gives up again and
`retries` rises rather than the loop continuing.

- [x] **Step 5: An ambiguous dispatch attribution is stamped**

Feed records that task 2's second pass refuses to attribute; assert `MarkDispatchesUnattributed` was
called with the ambiguity reason and the candidate count.

- [x] **Step 6: Run and confirm they fail**

```bash verified:the file and package exist today
cd stats && go test ./internal/harvest -run 'GiveUp|Retry|Stamped' -count=1
```

Expected: FAIL.

---

### 6 Persist and retry the give-up; stamp what could not be attributed

**Build:** green

**Files:**
- Modify: `stats/internal/harvest/watcher.go`
- Modify: `stats/internal/harvest/attribute.go` — surface the ambiguous candidates' dispatch ids
- Modify: `stats/internal/store/giveups.go` — return the harvest-declared give-up type
- Modify: `stats/internal/store/records.go` — stamp by dispatch id
- Modify: `stats/internal/store/records_test.go` — the by-id stamp's own test
- Modify: `stats/cmd/myflowd/main.go` — wire the store into the widened interface

**Tests:** the four from task 5, plus `TestMarkDispatchesUnattributedByID`.

**Three corrections to this task's original scope, found by task 5 and recorded here rather than
discovered again.** The original file list named `watcher.go` and `main.go` alone, and each of these
reaches past it:

1. **The give-up type must live in `internal/harvest`.** Task 4 shipped
   `store.PersistedGiveUps() ([]store.GiveUp, error)`. For `*store.Store` to satisfy the widened,
   consumer-side `SessionTokenBinder` **with no adapter** — the property `DispatchWindowSource`
   already has, and the one this task's own text claims — the returned type must be declared in
   `internal/harvest` and the store's signature must return that, exactly as
   `DispatchWindowsForSession` already returns `harvest.DispatchWindow`. The dependency runs
   store → harvest, which is what keeps `TestHarvestNeedsNoDatabase` true.
2. **The ambiguity stamp is by dispatch id, not by session token.**
   `MarkDispatchesUnattributed(token, …)` stamps **every** dispatch under a session token, which is
   right for `session never bound` and wrong for an ambiguity: it would stamp dispatches that
   attributed correctly, contradicting step 4's own "Do not stamp a dispatch that attributed". The
   ambiguous candidates are specific rows, and `DispatchWindow.DispatchID` already names them. Add
   `MarkDispatchesUnattributedByID(ctx, ids []int64, reason string, candidates int) error` beside the
   existing method rather than replacing it — the two reasons have genuinely different scopes.
3. **`DispatchAttributor.Attribute` must surface the candidates.** It returns
   `map[int64]TokenDelta` today and discards `bestDispatchWindow`'s third return value. It has to
   report which dispatch ids were indistinguishable, or the watcher has nothing to stamp.
   `DispatchWindow` carries no session token deliberately, so recovering one is not the route.

**Regression:** Reverting loses `kan-302`'s recoverability and leaves every ambiguous dispatch
indistinguishable from an unmeasured one.

**Baseline:** `go test ./internal/harvest ./cmd/myflowd` before=4 failures after=all pass.
<!-- predicted: the count is the number of tests the paired red task adds; confirmed by task 12 -->

**Commit:** `fix(harvest): persist an abandoned session token and retry it on restart`

**Interfaces:**
- Consumes: `SessionTokenBinder`, widened with `RecordSessionTokenGiveUp`, `PersistedGiveUps` and
  `MarkDispatchesUnattributed` — declared here, in `internal/harvest`, and satisfied by `*store.Store`
  with no adapter, exactly as `DispatchWindowSource` already is.
- Produces: durable give-ups and stamped dispatches.

- [x] **Step 1: Widen the binder interface**

Keep it consumer-side. `cmd/myflowd` asserts at compile time that `*store.Store` satisfies it, in the
idiom already used there for `DispatchWindowSource`.

- [x] **Step 2: Both give-up branches persist**

`resolveSessionTokens`' `case 0` and its `default` (ambiguity) branch each call `RecordSessionTokenGiveUp` with
its own distinct reason before setting `w.gaveUpTokens`. Keep the in-memory set: it is what stops the
same process re-searching within one run. The store is what survives the process.

- [x] **Step 3: Seed the pending set from `PersistedGiveUps` at start**

On the first cycle, read persisted give-ups and clear them from `w.gaveUpTokens` so they are searched
again. Bounded exactly as before — `w.tokenCycles` starts fresh, so a token that still matches
nothing gives up again after the same 60 cycles and increments `retries`.

- [x] **Step 4: Stamp dispatches the second pass refused to attribute**

Where `bestDispatchWindow` returns an ambiguity, call `MarkDispatchesUnattributed` for the candidates
with the reason and the count. Do not stamp a dispatch that attributed.

- [x] **Step 5: Run both packages**

```bash verified:both packages exist; `go test ./...` is .myflow/project.md's ## test command for stats
cd stats && go test ./internal/harvest ./cmd/myflowd -count=1
```

Expected: PASS.

---

### 6.1 Stamp a given-up session's own dispatches

**Build:** green

**Files:**
- Modify: `stats/internal/harvest/watcher.go`
- Modify: `stats/internal/harvest/watcher_test.go`

**Tests:** `TestGiveUpStampsItsOwnDispatches`

**Regression:** Reverting leaves the `cost unattributed — session never bound` state with **no
producer**: task 8 renders it, and nothing ever writes it, so `kan-302`'s dispatches keep rendering
`not measured` — the exact ambiguity this change exists to remove.

**Baseline:** `go test ./internal/harvest` before=1 new failure after=all pass.
<!-- predicted: one test, added by this task's own RED step; confirmed by task 12 -->

**Commit:** `fix(harvest): stamp the dispatches of a session that never bound`

**Interfaces:**
- Consumes: `MarkDispatchesUnattributed(ctx, token, reason string, candidates int) error` — the
  **token** form task 4 shipped, added to the widened `SessionTokenBinder`.
- Produces: the only writer of the `session never bound` reason task 8 renders.

**Why this task exists, found by task 6 rather than planned.** Task 6 wired two things: the session
give-up into `session_token_giveups`, and the dispatch-grain ambiguity into
`MarkDispatchesUnattributedByID`. Neither stamps the dispatches of a session that gave up. Task 6's
implementer correctly declined to add it — no test in its scope called for it, and inventing
untested behaviour is worse than reporting the gap — so it is a task of its own here.

The token form and the id form are both needed and are not alternatives: a session that never bound
has **every** one of its dispatches uncosted, which is what the token form expresses; an ambiguity
concerns specific rows, which is what the id form expresses.

- [x] **Step 1: Write the failing test**

`TestGiveUpStampsItsOwnDispatches`: drive `resolveSessionTokens` to each give-up branch in turn and
assert the fake binder recorded a `MarkDispatchesUnattributed` call carrying that token and the same
distinct reason the give-up itself recorded. Assert it for **both** branches — the exhausted window
and the ambiguous match — since both leave every dispatch of that session uncosted.

- [x] **Step 2: Confirm it fails**

```bash verified:the package and its test file exist on this branch as of 062dc11
cd stats && go test ./internal/harvest -run TestGiveUpStampsItsOwnDispatches -count=1
```

Expected: FAIL — nothing calls the token form.

- [x] **Step 3: Add the token form to the widened interface and call it**

Add `MarkDispatchesUnattributed` to `SessionTokenBinder` beside `MarkDispatchesUnattributedByID`, and
call it from both give-up branches of `resolveSessionTokens`, immediately after
`RecordSessionTokenGiveUp`, with that branch's own reason. `*store.Store` already implements it, so
the compile-time assertion in `cmd/myflowd/main.go` needs no change.

- [x] **Step 4: A stamp failure never blocks the give-up**

A store error here is reported and the give-up still stands — the never-block guarantee the whole
record path already carries. Assert it: a binder whose stamp returns an error must not stop the
token being recorded as given up.

- [x] **Step 5: Run the package**

```bash verified:same command as step 2, unrestricted
cd stats && go test ./internal/harvest -count=1
```

Expected: PASS.

---

### 6.2 Bind a given-up token by scanning for its marks, not by waiting for new bytes

**Build:** green

**Files:**
- Modify: `stats/internal/harvest/watcher.go`
- Modify: `stats/internal/harvest/watcher_test.go`

**Tests:** `TestPersistedGiveUpBindsFromAFullyConsumedTranscript`

**Regression:** Reverting restores the state task 12's live restart measured: the retry re-runs, finds
nothing, and gives up again — permanently, for every already-consumed transcript, which is every
transcript belonging to a run that has finished.

**Baseline:** `go test ./internal/harvest` before=1 new failure after=all pass.
<!-- predicted: one test, added by this task's own RED step; re-measured live by task 12 -->

**Commit:** `fix(harvest): bind a retried token by scanning, not by waiting for new bytes`

**Interfaces:**
- Consumes: the transcript reader, and `PersistedGiveUps` from task 6.
- Produces: a binding path that does not depend on the usage-attribution offset.

**Why this task exists, found by task 12's live restart rather than by any unit test.** Task 6's
retry re-adds a persisted give-up to the pending set, and task 5's tests pass because their fake
binder supplies matches directly. Against the real store it recovers nothing:
`matchSessionTokens` scans only the Bash commands **newly read** from a transcript in that cycle, and
`harvest_offsets` for a finished run's transcript already equals the file's full length — measured at
3,147,539 bytes for `kan-302`'s own transcript, exactly its size. The retry therefore searches a
stream with no new bytes, exhausts the same 60-cycle window and gives up again.

**The defect is that binding is gated on the usage-attribution offset, and the two are different
concerns.** The offset exists so usage is counted exactly once; binding a correlator to a session is
a search for a string that is already on disk and does not move. A retry must be able to look at
bytes the attribution pass has read past, **without** re-attributing their usage.

- [x] **Step 1: Write the failing test**

`TestPersistedGiveUpBindsFromAFullyConsumedTranscript`: a transcript whose offset is already at EOF
and whose earlier bytes contain a mark carrying a persisted give-up's token. Drive the watcher and
assert the token binds. Assert **also** that no usage is attributed a second time — the whole reason
the offset exists.

- [x] **Step 2: Confirm it fails**

```bash verified:the package and its test file exist on this branch as of 519e1df
cd stats && go test ./internal/harvest -run TestPersistedGiveUpBindsFromAFullyConsumedTranscript -count=1
```

Expected: FAIL — nothing looks behind the offset.

- [x] **Step 3: Add an offset-independent scan for pending tokens**

When, and only when, there are pending tokens the newly-read batches did not match, scan the
transcript files for mark invocations carrying those tokens, reading independently of
`harvest_offsets` and **without** producing usage records. Reuse `isSessionMarkCommand` so the
recogniser stays one function — the mention-versus-invocation distinction and the ambiguity rule must
hold identically on this path, or the scan reintroduces the bug KAN-172's finding F4 closed.

- [x] **Step 4: Keep it bounded**

The scan is work proportional to the transcripts on disk, so it must not run every cycle. Gate it on
having pending tokens to resolve, and keep the existing per-token cycle bound so a token whose marks
genuinely are not on disk still gives up rather than rescanning forever.

- [x] **Step 5: Ambiguity still refuses**

A token matching marks in two transcripts refuses to bind on this path exactly as on the other.
Assert it.

- [x] **Step 6: Run the package**

```bash verified:same command as step 2, unrestricted
cd stats && go test ./internal/harvest -count=1
```

Expected: PASS.

---

### 7 Ledger render tests — three states, told apart

**Build:** red

**Squash-with:** Task 8

**Files:**
- Modify: `stats/internal/records/render_test.go`

**Tests:** `TestLedgerSaysSessionNeverBound`, `TestLedgerSaysAmbiguousWithCount`,
`TestLedgerStillSaysNotMeasured`, `TestLedgerPrefersTokensOverUnattributed`,
`TestLedgerSaysSessionTokenMatchedManySessions`

**A fourth reason exists that this task's original four cases did not cover, found by task 7's own
implementer and added here.** The producers write **three** reasons, not two —
`reasonSessionNeverBound` (`session never bound`), `reasonSessionAmbiguous`
(`matched more than one session`) and `reasonDispatchAmbiguous` (`matched more than one dispatch`),
all in `stats/internal/harvest/watcher.go`. The delta spec originally enumerated wordings for the
first and third only. The second is a real state a producer writes, and rendering it under the
concurrent-dispatch wording would print a **session** count behind the words "concurrent
dispatches" — a misreported measurement, not an abbreviation. The delta spec now enumerates four
states and requires the two ambiguities to render differently; task 8 renders all four.

**Regression:** Reverting task 8 renders `not measured` for all three, which is the defect: a loss
becomes indistinguishable from an ordinary unmeasurable run.

**Baseline:** `go test ./internal/records` before=all pass after=3 new failures (the fourth passes
already).

**Interfaces:**
- Consumes: `records.Run`, `records.Dispatch`, `RenderLedger`.
- Produces: the executable statement of the three wordings.

- [x] **Step 1: A bag carrying `unattributed.reason = "session-never-bound"` renders that**

Assert the exact line `- Tokens: cost unattributed — session never bound`.

- [x] **Step 2: A bag carrying an ambiguity reason renders the count**

Assert `- Tokens: cost unattributed — indistinguishable from 3 concurrent dispatches`.

- [x] **Step 3: An empty bag still renders `not measured`**

The existing behaviour, restated as a test so task 8 cannot widen the new wording over it. `{}` is the
permanent shape on Cursor and Codex.

- [x] **Step 4: A bag carrying both `tokens` and `unattributed` renders the tokens**

A real measurement outranks a stale stamp. Assert the figures render and the unattributed wording does
not appear.

- [x] **Step 5: Run and confirm they fail**

```bash verified:the package and its test file exist today
cd stats && go test ./internal/records -run TestLedger -count=1
```

Expected: FAIL on steps 1 and 2; steps 3 and 4 pass already.

---

### 8 Render the three states

**Build:** green

**Files:**
- Modify: `stats/internal/records/render.go`

**Tests:** the four from task 7.

**Regression:** Reverting collapses the three states back to one string.

**Baseline:** `go test ./internal/records` before=3 failures after=all pass.
<!-- predicted: the count is the number of tests the paired red task adds; confirmed by task 12 -->

**Commit:** `feat(records): tell an unattributed dispatch from an unmeasured one in the ledger`

**Interfaces:**
- Consumes: the dispatch's `Metrics` bag.
- Produces: `tokenLine`'s widened output.

- [x] **Step 1: Extend `dispatchMetrics` with the unattributed key**

Add `Unattributed *unattributed` beside `Tokens *tokenTotals`, pointer for the same reason `Tokens` is
one: absence must be distinguishable from a zero value, and `{}` is the ordinary shape.

- [x] **Step 2: `tokenLine` checks `Tokens` first, then `Unattributed`, then falls back**

Order matters and task 7 step 4 pins it. Keep the existing "an explicit zero is a real figure and is
not hidden" behaviour untouched.

- [x] **Step 2b: Render all four states, and never one ambiguity under the other's wording**

Map each producer reason to its own wording: `session never bound` →
`cost unattributed — session never bound`; `matched more than one session` →
`cost unattributed — the session token matched N sessions`; `matched more than one dispatch` →
`cost unattributed — indistinguishable from N concurrent dispatches`. The two ambiguities carry
counts of different things — sessions and dispatches — so a shared phrase would label one as the
other.

- [x] **Step 3: Render an unrecognised reason verbatim rather than as `not measured`**

A reason this code does not know is still a statement that attribution failed. Falling back to
`not measured` for it would reintroduce the ambiguity this task removes. Pass it through
`neutraliseMarkers`, as every other externally-sourced string in this file is.

- [x] **Step 4: Run the package**

```bash verified:same command as task 7 step 5, unrestricted
cd stats && go test ./internal/records -count=1
```

Expected: PASS.

---

### 9 CLI and API tests — `-agent-id` on `end`, and `record cost-status`

**Build:** red

**Squash-with:** Task 10

**Files:**
- Modify: `stats/cmd/myflow/record_test.go`
- Modify: `stats/internal/api/records_test.go`

**Tests:** `TestDispatchEndAcceptsAgentID`, `TestDispatchEndAgentIDReachesTheStore`,
`TestCostStatusPrintsOneLine`, `TestCostStatusNeverBlocks`

**Regression:** Reverting task 10 rejects `-agent-id` on `end` as an unknown flag, and
`cost-status` is an unknown verb.

**Baseline:** `go test ./cmd/myflow ./internal/api` before=all pass after=4 new failures.

**Interfaces:**
- Consumes: the existing CLI test harness and the API's `fakeStore`.
- Produces: the executable statement of both surfaces.

- [x] **Step 1: `end` accepts `-agent-id` and forwards it**

- [x] **Step 2: `end` without `-agent-id` is unchanged**

An identifier is optional at both ends. Assert the request carries no id rather than an empty one.

- [x] **Step 3: `cost-status` prints one line and exits 0**

Shape it on `journal-count`: one line on stdout, `unknown` where no answer could be produced.

- [x] **Step 4: `cost-status` exits 0 when the store is unreachable**

It exists so a handoff can state the figure honestly, so it must never become the reason the handoff
does not print. Same contract `journal-count` carries.

- [x] **Step 5: Run and confirm they fail**

```bash verified:both packages and both test files exist today
cd stats && go test ./cmd/myflow ./internal/api -run 'AgentID|CostStatus' -count=1
```

Expected: FAIL.

---

### 10 `-agent-id` on `end`, and `myflow record cost-status`

**Build:** green

**Files:**
- Modify: `stats/cmd/myflow/record.go`
- Modify: `stats/internal/records/types.go` — `DispatchEnd.AgentID`, the wire field the CLI forwards
- Modify: `stats/internal/api/records.go`
- Modify: `stats/internal/api/server.go`
- Modify: `stats/internal/store/records.go`
- Modify: `stats/internal/client` — the two calls

**One correction to this task's file list, found by task 9.** `records.DispatchEnd` is where the
`AgentID` field has to land, and `internal/records/types.go` was absent from the list above. Task 9
also updated `internal/api/records_test.go`'s `fakeStore.EndDispatch` to thread the field, so that
this task can reach green by touching production files alone.

**Tests:** the four from task 9.

**Regression:** Reverting leaves a dispatch whose identifier is knowable only at close permanently
without one — the case `/myflow-do` hits on every synchronous dispatch.

**Baseline:** `go test ./cmd/myflow ./internal/api ./internal/store` before=4 failures after=all
pass.
<!-- predicted: the count is the number of tests the paired red task adds; confirmed by task 12 -->

**Commit:** `feat(myflow): record a dispatch's agent id at close, and report unattributed cost`

**Interfaces:**
- Consumes: the four-layer path `myflow record dispatch` already traverses.
- Produces: `-agent-id` on `end`, and the `cost-status` read task 11's handoff line prints.

- [x] **Step 1: Thread `-agent-id` through `end`, at every layer**

`ApplyDispatchEnd` sets it only when non-empty — an `end` that omits it must never clear an id
`begin` recorded.

- [x] **Step 2: Add `cost-status`**

One line naming how many of a change's dispatches are in an unattributed state and why, or that none
are. It reads the store; on any failure it prints `unknown` and exits 0.

- [x] **Step 3: Extend `recordUsage`**

Both the `end` signature and the new verb. The usage block is the CLI's own documentation and a stale
one is a defect.

- [x] **Step 4: Run the three packages**

```bash verified:same packages as task 9 step 5, unrestricted
cd stats && go test ./cmd/myflow ./internal/api ./internal/store -count=1
```

Expected: PASS.

---

### 11 `/myflow-do` captures the identifier and reports unattributed cost

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-do/SKILL-rationale.md`

**Tests:** none — this file is prose. `scripts/check-references.sh`,
`scripts/check-vocabulary.sh` and `scripts/check-contract-budget.sh` are what check it.

**Regression:** Reverting leaves `agent_id` NULL on every dispatch this repository records, which is
the state the audit found and the whole cause of the panel's wrong split.

**Baseline:** `scripts/check-contract-budget.sh` before=clean after=clean. `skills/myflow-do/SKILL.md`
is 87,388 bytes against a 103,578-byte budget before this change.
<!-- measured: git show 4a305d0:skills/myflow-do/SKILL.md | wc -c; scripts/check-contract-budget.sh:192 @ 4a305d0 -->

**This figure was wrong when the plan was written, and is corrected here rather than quietly
replaced.** It first read 85,556 bytes, which is `kan-302`'s own *pre-edit* measurement taken at
`441ed2b` and mislabelled `@ 4a305d0`. That change then grew the file by 3,890 bytes, so 87,388 is
what `main` actually carries. Task 11's implementer measured the file on disk, found the mismatch and
reported it instead of adopting the stale number — which is what a provenance tag is for.

**Commit:** `fix(myflow-do): capture the dispatched agent's identifier and report uncosted dispatches`

**Interfaces:**
- Consumes: task 10's `-agent-id` on `end` and `cost-status`.
- Produces: dispatch rows that actually carry an identifier.

- [x] **Step 1: Say where the identifier comes from**

The present text treats it as something the harness may or may not expose. On Claude Code an async
agent launch returns it in the parent's own tool result, at launch. State that, so the dispatcher
knows to read it rather than to hope for it. Keep **Never invent one** exactly as written — it is the
rule that stops a placeholder pairing two unrelated slots.

- [x] **Step 2: Record it at `begin` where known, at `end` otherwise**

Add `-agent-id` to the `end` invocation at all four dispatch sites, and say the opening call is never
delayed to obtain one: the row must exist before the first harvest tick the dispatch runs through.

- [x] **Step 3: Add the handoff line**

One line in **7. Verify, stage, and hand off**, printed from `myflow record cost-status`, in the
idiom the `Records:` line already uses. It renders what the command printed, `unknown` included, and
never derives the figure by hand.

- [x] **Step 4: Record the reasoning in `SKILL-rationale.md`**

Why identity outranks the interval, and the audit figure behind it — 20 of 57 dispatches costed, the
panel's slots collapsing onto one row. A rule whose reason is not written down is the one a later
change reverts.

- [x] **Step 5: Run the guards**

```bash verified:every script named is listed in .myflow/project.md's ## lint section
scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-contract-budget.sh
```

Expected: exit 0 from each.

---

### 12 Full verification, and re-measure the baseline

**Build:** green

**Files:**
- Create: `openspec/changes/kan-212-persist-per-dispatch-cost-tokens-model-and-role/verification.md`

**Tests:** the whole suite — every test named above, plus everything already green.

**Regression:** Nothing to revert; this task is the evidence the other eleven hold together.

**Baseline:** `cd stats && go test ./...` before=all pass after=all pass.

**Commit:** `test(stats): verify per-dispatch attribution end to end`

**Interfaces:**
- Consumes: every preceding task.
- Produces: the recorded before/after figures the proposal's claim rests on.

- [x] **Step 1: Run every lint command**

```bash verified:each command is listed in .myflow/project.md's ## lint section
cd stats && gofmt -w . && go vet ./... && gofmt -l .
```

Expected: `gofmt -l .` prints nothing and `go vet` exits 0.

- [x] **Step 2: Run the whole Go suite**

```bash verified:`go test ./...` is .myflow/project.md's ## test command for stats
cd stats && go test ./... -count=1
```

- [x] **Step 3: Run every guard in `## lint`**

Each `scripts/check-*.sh` the project declares, and its `scripts/test-check-*.sh` partner.

- [x] **Step 4: Restart the daemon and re-read the baseline**

The give-up retry runs at start, so this is where `kan-302`'s 15 stage runs and 12 dispatches either
recover or do not.

```bash unverified:the exact restart invocation depends on how myflowd is running on the machine at the time
cd stats && docker compose exec -T myflow-postgres psql -U myflow -d myflow -c \
  "SELECT count(*) AS dispatches, count(*) FILTER (WHERE metrics ? 'tokens') AS costed FROM dispatches;"
```

Baseline was 57 dispatches, 20 costed. Record the figure actually observed, whatever it is — a
number that did not move is a finding, not a failure to report.

- [x] **Step 5: Write `verification.md`**

Every command run, its output, and the before/after dispatch figures. Where a figure did not move,
say so and say what that implies about `kan-302-give-up-trigger`, the open question in `design.md`.
