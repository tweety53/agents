# Tasks — kan-327-review-per-round-delta-after-round-1

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

Baseline counts this plan measures against, on the metric
`cd stats && go test <pkg> -count=1 -v | grep -c '^=== RUN   Test'`:

- `./internal/store` reads 227.
  <!-- measured: cd stats && go test ./internal/store -count=1 -v | grep -c '^=== RUN   Test' @ branch fix/restart-rebuilds-myflowd-twice @ 2c4f312 -->
- `./internal/records` reads 38.
  <!-- measured: cd stats && go test ./internal/records -count=1 -v | grep -c '^=== RUN   Test' @ branch fix/restart-rebuilds-myflowd-twice @ 2c4f312 -->
- `./cmd/myflow` reads 104.
  <!-- measured: cd stats && go test ./cmd/myflow -count=1 -v | grep -c '^=== RUN   Test' @ branch fix/restart-rebuilds-myflowd-twice @ 2c4f312 -->

The metric counts subtests, not top-level functions. Where a task below adds a table-driven test,
its `**Baseline:**` moves by the number of cases that table declares, not by one — the numbers
stated per task are what to hold the run to; if a table's shape changes, correct the number rather
than reshaping the test to match it.

---

### 1 The store carries a dispatch's diff base

**Build:** green

**Files:**
- Add: `stats/internal/store/migrations/0014_dispatch_diff_base.sql`
- Modify: `stats/internal/records/types.go`
- Modify: `stats/internal/store/records.go`
- Modify: `stats/internal/store/records_test.go`

**Tests:** `TestRecordDispatchRoundTripsDiffBase`, `TestRecordDispatchOmitsAbsentDiffBase`

**Regression:** both fail if `diff_base` is dropped from `dispatchColumns` or from
`insertDispatch`'s column list. The list and the scan targets are counted against each other, so
either omission is a hard scan or placeholder-count error rather than a silent empty read — which is
the point of the single constant, and why the sha a panel slot read from cannot quietly become
indistinguishable from a dispatch that recorded none.
<!-- measured: task 1's implementer reproduced both omissions; the scan fails on 17 targets against 16 columns -->

**Baseline:** `./internal/store` before=227 after=229

**Interfaces:**
- Produces: `Dispatch.DiffBase`, persisted and read back.
- Consumed by: tasks 2 and 3.

**Commit:** `feat(store): record a dispatch's diff base`

- [x] **Step 1: Add the migration**

`stats/internal/store/migrations/0014_dispatch_diff_base.sql`, in the same single-column shape
`0011_dispatch_agent_id.sql` and `0012_dispatch_key.sql` already use — read one of them first and
match its comment convention rather than inventing a new one.

```sql unverified:the column type follows commit_sha in 0010_run_records.sql; confirm the migration runner needs no accompanying registration entry
ALTER TABLE dispatches ADD COLUMN diff_base TEXT;
```

No index. Nothing queries across this column: it is read back only as part of a row already
selected by change.

- [x] **Step 2: Add the field to the record type**

`stats/internal/records/types.go`, in `Dispatch`, beside `CommitSHA`:

```go unverified:field placement follows CommitSHA at types.go:79; confirm no golden JSON fixture pins the struct's field order
DiffBase string `json:"diffBase,omitempty"`
```

`omitempty` matters: an absent base means *not recorded*, and must not serialise as an empty string
that reads like a recorded one.

- [x] **Step 3: Carry it through the column list, the scan and the insert**

`stats/internal/store/records.go`. `dispatchColumns` is the single list every read selects
(`records.go:142`), and `qualifiedDispatchColumns` derives from it by splitting on commas, so
adding the name in one place reaches both. Then `scanDispatchRow` and `insertDispatch`
(`records.go:224`, the `RETURNING` at `records.go:235`, and the `SELECT` at `records.go:671`).

```go verified:stats/internal/store/records.go:142 @ 2c4f312
const dispatchColumns = `id, seq, stage_run_id, task_id, role, slot, model, commit_sha, outcome,
```

Scan the column as a nullable text and fold SQL `NULL` to the empty string, exactly as the other
optional text columns on this row already do — do not introduce a `*string` on `Dispatch`.

- [x] **Step 4: Test the round trip and the absence**

`stats/internal/store/records_test.go`, following the file's existing dispatch round-trip tests.
`TestRecordDispatchRoundTripsDiffBase` writes a dispatch carrying a base and reads back exactly the
sha it wrote. `TestRecordDispatchOmitsAbsentDiffBase` writes one carrying none and asserts the field
reads back empty and the row's other columns are unaffected.

---

### 2 The ledger names the base a dispatch read from

**Build:** green

**Files:**
- Modify: `stats/internal/records/render.go`
- Modify: `stats/internal/records/render_test.go`

**Tests:** `TestRenderLedgerNamesDiffBase`, `TestRenderLedgerOmitsAbsentDiffBase`,
`TestRenderPanelUnchangedByDiffBase`

**Regression:** `TestRenderLedgerNamesDiffBase` fails if the line is dropped — the audit trail stops
saying what any panel slot actually read, which is the whole reason the field is stored rather than
held in session. `TestRenderLedgerOmitsAbsentDiffBase` fails if the line is emitted unconditionally,
which would print a fabricated "no base" for every implementer dispatch. `TestRenderPanelUnchangedByDiffBase`
fails if the field reaches `RenderPanel`, which would put a new line into the record
`check-panel-reproducers.sh` and `check-unfinished-work.sh` parse.

**Baseline:** `./internal/records` before=38 after=41

**Interfaces:**
- Consumes: task 1's `Dispatch.DiffBase`.
- Produces: the `- Diff base:` line in the rendered SDD ledger.

**Commit:** `feat(records): render a dispatch's diff base in the ledger`

- [x] **Step 1: Emit the line in `RenderLedger`**

`stats/internal/records/render.go`, in the per-dispatch loop, immediately after the commit line:

```go verified:stats/internal/records/render.go:339 @ 2c4f312
fmt.Fprintf(&b, "- Commit: %s\n", orElse(d.CommitSHA, "no commit"))
```

The new line is **conditional**, in the shape the `Slot` and `Notes` lines already use — emitted
only when the field is non-empty, and passed through `neutraliseMarkers` for the same reason every
other rendered value is:

```go unverified:mirrors the Slot line at render.go:335-337; confirm neutraliseMarkers is the right wrapper for a value that is a sha rather than free text
if strings.TrimSpace(d.DiffBase) != "" {
    fmt.Fprintf(&b, "- Diff base: %s\n", neutraliseMarkers(d.DiffBase))
}
```

It is conditional rather than `orElse`-defaulted because every implementer dispatch and the primary
reviewer's own dispatch legitimately carry none, and a `- Diff base: no base` line on each of them
is noise that says nothing.

- [x] **Step 2: Leave `RenderPanel` alone, and prove it**

`RenderPanel` is not edited. `TestRenderPanelUnchangedByDiffBase` renders one `Run` twice — once with
every dispatch carrying a base, once with none — and asserts the two panel records are byte-for-byte
equal. This is the test that holds the marker contract the two panel guards parse.

---

### 3 `myflow record dispatch begin` accepts the base

**Build:** green

**Files:**
- Modify: `stats/cmd/myflow/record.go`
- Modify: `stats/cmd/myflow/record_test.go`

**Tests:** `TestRecordDispatchBeginAcceptsDiffBase`, `TestRecordDispatchBeginDiffBaseIsOptional`

**Regression:** `TestRecordDispatchBeginAcceptsDiffBase` fails if the flag is not wired to the
request body — the panel would call it and the store would record nothing, silently.
`TestRecordDispatchBeginDiffBaseIsOptional` fails if the flag becomes required, which would break
every implementer dispatch and every existing call site.

**Baseline:** `./cmd/myflow` before=104 after=106

**Interfaces:**
- Consumes: task 1's field.
- Produces: `-diff-base <sha>` on `dispatch begin`, for tasks 4 and 5 to call.

**Commit:** `feat(myflow): accept -diff-base on dispatch begin`

- [x] **Step 1: Add the flag**

`stats/cmd/myflow/record.go`, on the `dispatch begin` flag set, beside `-agent-id` — which is the
closest existing precedent: an optional string, absent meaning *not reported*. Carry it into the
`records.Dispatch` the command sends.

- [x] **Step 2: Update the usage block**

The usage text `myflow record` prints is the CLI's own documentation. Add `[-diff-base sha]` to the
`dispatch begin` line, in the bracketed-optional style the line already uses for `[-task id]`,
`[-slot name]` and `[-agent-id id]`.

- [x] **Step 3: Test both shapes**

`stats/cmd/myflow/record_test.go`, following the file's existing `dispatch begin` flag tests: one
asserting the sha reaches the request body, one asserting the command succeeds with the flag absent
and sends no base.

---

### 4 A Full-mode re-run gives each slot only what it has not read

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md`

**Commit:** `feat(myflow-do): give each panel slot its own delta in a full re-run`

**Interfaces:**
- Consumes: task 3's `-diff-base` flag.
- Produces: the rule the panel actually runs under.

- [x] **Step 1: Rewrite the re-run table's Full row**

`skills/myflow-do/SKILL.md`, the **Panel re-runs** table. The Targeted row is unchanged. The Full
row reads today:

```markdown verified:skills/myflow-do/SKILL.md:808 @ 2c4f312
| **Full** (escalation) | Every **required** slot; a **conditional** slot (Security, Adversarial, Lens B, Lens C) only when its own row in the optional-slot trigger table still fires against `fix-round-N.diff` | rewritten `final-review.diff` |
```

Its *Diff they get* cell becomes per-slot: slot 0 the rewritten `final-review.diff`, every other
diff-reading slot its own delta. Who re-runs is unchanged in this cell — the empty-delta exclusion
is task 5's, and stating it here as well would put one rule in two places.

- [x] **Step 2: State the per-slot delta, under the table**

A new paragraph directly under the table, stating, in this order:

- the delta is `git diff <the HEAD sha that slot last reviewed> HEAD`, written to
  `<abs-worktree>/.superpowers/sdd/slot-delta-<round>-<slot>.diff`;
- it is anchored at that slot's own last read, never at the current round, because Targeted mode
  re-runs only slot 0 and the slots that raised findings, so a slot reaching a Full pass may have
  missed rounds in between;
- the range is tree-to-tree and needs no ancestry, so a base recorded before a
  `git rebase --autosquash` stays valid after that rebase rewrote the task commit;
- a slot for which no last-reviewed sha is held reads the whole `final-review.diff` — an unrecorded
  base is not a small delta;
- Bugbot and Security read no diff file and are unaffected;
- coverage holds because pass 1 always runs the full roster against the full diff, and there is no
  exemption for the pass that closes the panel.

Do not restate the zero-open-findings bar here: *No preset moves the handoff bar*, earlier in the
same section, already carries it.

- [x] **Step 3: Require each prompt to say what it is holding**

One sentence in the same paragraph: every slot's dispatch prompt names the path it was given and,
for a delta, the sha that delta starts from. This is what makes a reviewer's "I did not see X"
checkable after the fact rather than a matter of recollection.

- [x] **Step 4: Carry the base into the dispatch record**

The `myflow record dispatch begin` block in section 5 gains `-diff-base <sha>` on a slot dispatched
against a delta, with one sentence saying that a slot reading the whole diff passes none, and one
saying that scheduling reads the dispatcher's own in-session value rather than the store — which is
what keeps this on the never-block guarantee the rest of the block already relies on.

---

### 5 An empty delta is a third recorded silence

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md`

**Commit:** `feat(myflow-do): skip a panel slot whose delta is empty`

**Interfaces:**
- Consumes: task 4's per-slot delta.
- Produces: the third not-re-run disposition.

- [x] **Step 1: State the empty-delta skip**

In section 5, after task 4's paragraph: a required slot reading a delta whose delta is empty is not
dispatched, and the panel record says `not re-run — nothing new since its last read`. Slot 0 has no
delta and is never scoped out by this.

State plainly that it closes, softens and expires no finding, and that a slot with an open finding
still blocks the handoff — the same sentence the conditional-slot rule already carries, because this
is the point at which a reader will otherwise read the skip as a waiver.

- [x] **Step 2: Correct the conditional-slot paragraph's boundary sentence**

The paragraph under the re-run table that today explains why trigger-scoping reaches conditional
slots alone rests on "a required slot has no bounded region — it reads the whole diff". That premise
is no longer true of a required delta-slot. Rewrite the sentence to say that **trigger-based**
scoping still reaches conditional slots alone, and that a required delta-slot is scoped, if at all,
only by its own delta being empty — two mechanisms, two recorded reasons.

Both the `not re-run — subject unchanged` wording and the new one must survive verbatim: the record
distinguishes three silences, and a reader diffing this file must be able to see that neither
replaced the other.

---

### 6 Record the reasoning, and verify

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL-rationale.md`
- Modify: `scripts/check-contract-budget.sh`

**Commit:** `docs(myflow-do): record why the panel delta is anchored per slot`

`check-contract-budget.sh` is in this task's files because the four entries below plus the step 1
correction do not fit `SKILL-rationale.md`'s declared budget, and trimming to fit would cut recorded
reasons — which the writing standard forbids. Raise that one row to the file's new size plus 25%,
the ratchet formula `.myflow/project.md` states, and change nothing else in the guard. Narrowing the
guard's scope or deleting a row is not the response, and neither is a suppression.
<!-- measured: SKILL-rationale.md had 1,279 bytes of headroom (19,623 of 20,902) against a need of about 2,400 -->

**Interfaces:**
- Consumes: tasks 4 and 5.

- [x] **Step 1: Correct the section's stale premise**

`skills/myflow-do/SKILL-rationale.md:165-166` opens `### Panel re-runs` with "Why the not-re-run
scoping reaches conditional slots alone: … while a required slot has no bounded region to check
staleness against." Task 5 falsified the second half: a required delta-slot's bounded region is its
own delta. Rewrite that entry to say why **trigger-based** scoping reaches conditional slots alone,
and name the empty-delta skip as the separate mechanism that scopes a required slot. This entry sits
in task 6's file rather than task 5's, which is why task 5 could not reach it — task 5 rewrote
`SKILL.md`'s pointer sentence, and this is the pointed-at text.
<!-- measured: the tasks 4+5 implementer located the sentence at SKILL-rationale.md:166, outside task 5's declared Files -->

- [x] **Step 2: Add the reasoning to the rationale's `### Panel re-runs` section**

Add, beside the existing entries and without otherwise touching them:

- why the delta is anchored at the slot's own last read rather than at the current round — the
  Targeted-mode skip, and the coverage-waiver rule it would breach. **Carry KAN-315's numbers from
  `proposal.md`**, which records them with their provenance, rather than restating them from
  memory;
- why the range is tree-to-tree rather than `A..B` — `--autosquash` rewrites the task commit, and an
  ancestry-dependent range would silently stop being computable;
- why Primary is the exception — plan alignment and the change's history are not carried by a delta;
- why the base is stored on the dispatch row but never scheduled off — the never-block guarantee
  only stays safe while nothing depends on the write having landed.

- [x] **Step 3: Run the guards and the suites**

```bash verified:the lists these commands come from are .myflow/project.md's `## lint` and `## test` sections
scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-normative-inventory.sh \
  && scripts/check-markdown-integrity.py && scripts/check-plan-provenance.sh \
  && scripts/check-task-build-green.sh
```

```bash verified:.myflow/project.md's `## test` section names this as its own invocation, split out because the full list exceeds the tool timeout
cd stats && gofmt -l . && go vet ./... && go test ./... -race -count=1
```

`gofmt -l .` printing any path is a failure, not a report: run `gofmt -w .` and re-run. The guard
list is split from the Go suite deliberately — `.myflow/project.md`'s `## test` section measures the
full set at roughly 144s against a 120s tool timeout, so one invocation of everything times out
without anything having failed.
