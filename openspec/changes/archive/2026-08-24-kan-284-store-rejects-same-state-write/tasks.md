# Tasks — kan-284-store-rejects-same-state-write

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

Baseline counts this plan measures against, on the metric
`go test <pkg> -count=1 -v | grep -c '^=== RUN   Test'`: `stats/cmd/myflow` reads 93 and
`stats/internal/store` reads 221.
<!-- measured: cd stats && go test ./cmd/myflow/ -count=1 -v | grep -c '^=== RUN   Test'  and  go test ./internal/store/ -count=1 -v | grep -c '^=== RUN   Test' @ branch main -->

**That metric counts subtests, not only top-level test functions.** Every `**Baseline:**` field below
is stated on it, so a task adding a table-driven test moves the number by more than the count of
`func Test` it adds. Do not collapse a table-driven test into an inline loop to make a number match —
correct the number.
<!-- measured: the discrepancy was found implementing task 2, whose three added tests moved the metric by 7 because one is table-driven with 4 cases @ branch openspec/kan-284-store-rejects-same-state-write -->

### 1 The CLI stamps `updatedAt` on every `state set`

- [x] Add a `stampUpdatedAt(body []byte) ([]byte, error)` helper beside `withMainCheckoutPath` in
  `stats/cmd/myflow/state.go` that decodes the body as a JSON object, sets `updatedAt` to
  `time.Now().UTC().Format(time.RFC3339Nano)`, and re-encodes it. Call it in `runStateSet`
  immediately after the `isJSONObject` check and before `fallback.ProjectKey`, assigning back into
  `body` so the store request, the on-disk fallback file and the journal entry all carry the stamped
  value.

The helper mirrors `withMainCheckoutPath`'s shape — same decode/assign/re-encode, same nil-map
guard — because the two do the same kind of job and a reader should not have to learn a second
pattern:

```go unverified:confirm encoding/json decodes a JSON null body into a nil map here exactly as withMainCheckoutPath's comment already records
func stampUpdatedAt(body []byte) ([]byte, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	if raw == nil {
		raw = map[string]json.RawMessage{}
	}
	stamped, err := json.Marshal(time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	raw["updatedAt"] = stamped
	return json.Marshal(raw)
}
```

**Files:** `stats/cmd/myflow/state.go`, `stats/cmd/myflow/state_test.go`
**Tests:** `TestStateSetStampsUpdatedAtOverBodyValue`,
`TestStateSetStampsSameInstantIntoRequestAndJournal`,
`TestStateSetStampIsFinerThanSecondPrecision`
**Regression:** `TestStateSetStampsUpdatedAtOverBodyValue` fails if the body's own `updatedAt`
reaches the store, which is the defect: a second-truncated value loses to the sub-second instant a
stage mark's synthetic bootstrap already wrote. `TestStateSetStampsSameInstantIntoRequestAndJournal`
fails if the stamp is applied to the request but not to the journalled body, which would make a
replayed entry order by the wrong instant. `TestStateSetStampIsFinerThanSecondPrecision` fails if
the format is narrowed back to `RFC3339`, reintroducing the truncation this task removes.
**Baseline:** before=93 after=96
<!-- measured: cd stats && go test ./cmd/myflow/ -count=1 -v | grep -c '^=== RUN   Test' @ branch openspec/kan-284-store-rejects-same-state-write -->
**Commit:** `fix(myflow): stamp updatedAt in the CLI, at full precision`
**Build:** green

### 2 The CLI refuses a `worktrees` value that is not a sha or `null`

- [x] Add a `validateWorktreeMergeBases(body []byte) error` helper in `stats/cmd/myflow/state.go`
  that decodes `worktrees` as `map[string]*string` and returns an error naming the first offending
  worktree path and its rejected value unless every value is `nil` or matches `^[0-9a-f]{40}$`.
  Call it in `runStateSet` after the stamp from task 1 and before `fallback.ProjectKey`; on error
  print one stderr line and `return 2`, taking neither the store path nor the fallback path.

Placing the check before `fallback.ProjectKey` is what guarantees the refusal writes nothing: the
fallback state file and journal paths are both derived from the project key, so a return that
happens first cannot have touched either.

```go unverified:confirm a package-scope regexp is the pattern the surrounding file already uses rather than a local MustCompile
var mergeBaseRE = regexp.MustCompile(`^[0-9a-f]{40}$`)
```

**Files:** `stats/cmd/myflow/state.go`, `stats/cmd/myflow/state_test.go`
**Tests:** `TestStateSetAcceptsShaAndNullMergeBases`,
`TestStateSetRefusesMalformedMergeBase`, `TestStateSetRefusalWritesNoFallback`
**Regression:** `TestStateSetRefusesMalformedMergeBase` fails if a worktree path, a short sha, an
uppercase sha or an empty string is accepted in the merge-base position — the value KAN-265 recorded
and could not correct. `TestStateSetRefusalWritesNoFallback` fails if the refusal takes the
never-block fallback path, which would journal the bad value and hide it until the finish gate.
`TestStateSetAcceptsShaAndNullMergeBases` fails if the check rejects `null`, which the state file
contract defines as *no merge base recorded* and which must stay legal.
**Baseline:** before=96 after=104
<!-- measured: cd stats && go test ./cmd/myflow/ -count=1 -v | grep -c '^=== RUN   Test' @ branch openspec/kan-284-store-rejects-same-state-write -->
**Commit:** `fix(myflow): refuse a worktrees value that is not a sha or null`
**Build:** green

### 3 The store refuses a malformed merge base on write

- [x] Add `ErrInvalidMergeBase` to `stats/internal/store/changes.go` beside `ErrInvalidState`, with
  a doc comment saying it is a caller-content fault distinguishable from both an invalid state and a
  monotonic refusal. In `PutChange`, before the transaction opens, return it for any `c.Repos` entry
  whose `MergeBase` is non-nil and does not match `^[0-9a-f]{40}$`, naming the repo root and the
  rejected value.

This is unreachable for a write made through the CLI, which refused in task 2. It exists for
`internal/reconcile` replaying a hand-edited fallback file — the case
`skills/myflow-contracts/state-file.md` already names. `IsDefinitiveChangeOutcome` already treats a
content-invalid outcome as definitive, so such an entry retires from the journal rather than
replaying forever; confirm that during implementation and add nothing if it already holds.

`IsDefinitiveChangeOutcome` turned out **not** to hold already, so `stats/internal/api/changes.go`
gains one `case` for the new sentinel and `stats/internal/reconcile/reconcile_test.go` gains the test
that pins it — without them a journal entry carrying a bad merge base never retires and blocks every
valid entry behind it. `stats/internal/store/repos_test.go` carried placeholder merge bases
(`abc123`, `def456`, `aaa111`) that the new check refuses; they are widened to 40 hex characters
keeping their mnemonic prefixes, since those tests assert repo-set round-tripping rather than
merge-base shape.

`mapStoreError` gains the new sentinel at 400 alongside its three caller-content siblings, and
`TestErrorStatusMapping` gains the one table row that pins it — leaving it unmapped returned 500 for
a client mistake and wrote a `logger.Error` line per occurrence.

**Files:** `stats/internal/store/changes.go`, `stats/internal/store/changes_test.go`,
`stats/internal/api/changes.go`, `stats/internal/reconcile/reconcile_test.go`,
`stats/internal/store/repos_test.go`, `stats/internal/api/server.go`,
`stats/internal/api/changes_test.go`
**Tests:** `TestPutChangeRefusesMalformedMergeBase`,
`TestPutChangeMergeBaseErrorIsDistinctFromInvalidState`
**Regression:** `TestPutChangeRefusesMalformedMergeBase` fails if a non-sha merge base reaches
`change_repos`, which is the path that bypasses the CLI entirely.
`TestPutChangeMergeBaseErrorIsDistinctFromInvalidState` fails if the two errors collapse into one,
which would stop a caller telling a malformed value apart from a malformed state. The existing
`TestPutChangeAppliesInOrderWriteAtSameState` and the identical-retry refusal test MUST stay green
**unedited** — they are the evidence that the monotonic rule this change deliberately does not touch
still holds.
**Baseline:** before=221 after=227
<!-- measured: cd stats && go test ./internal/store/ -count=1 -v | grep -c '^=== RUN   Test' @ branch openspec/kan-284-store-rejects-same-state-write -->
**Commit:** `fix(store): refuse a merge base that is not a 40-hex sha`
**Build:** green

### 4 The contract records `updatedAt` as CLI-owned

- [x] In `skills/myflow-contracts/state-file.md`, rewrite the `updatedAt` bullet to say the CLI
  stamps it on every `state set` at full precision, that the stamped value is what the store row,
  the fallback file and the journal entry all carry, and that a value a caller supplies is ignored
  rather than refused. Remove the `date -u +%Y-%m-%dT%H:%M:%SZ` instruction and the "never invent or
  placeholder it" warning — a field a skill cannot write cannot be fabricated.
- [x] Extend the `worktrees` bullet with the value constraint from task 2 and name where it is
  refused. Leave the existing paragraph about a `null` value untouched.
- [x] Add one sentence to **Writes are monotonic in both dimensions** recording that the instant it
  orders by now comes from a single writer. Leave its normative rule verbatim.

`state-file.md` is the only file under `skills/` that mentions reading the clock for this field, so
no other skill file changes:

```bash verified:run at branch main, this returns exactly one hit — state-file.md:215
grep -rn "date -u" skills/
```

**Files:** `skills/myflow-contracts/state-file.md`
**Tests:** *(none — prose; covered by the repository's own guards)*
**Regression:** `scripts/check-normative-inventory.sh` captured before the first edit and diffed
after the last must show no removed or reworded normative sentence other than the `updatedAt`
instruction this task deliberately retires; `scripts/check-contract-budget.sh` must still pass, with
`state-file.md`'s budget raised deliberately if the additions trip it.
**Baseline:** before=104 after=104
<!-- predicted: cd stats && go test ./cmd/myflow/ -count=1 -v | grep -c '^=== RUN   Test' after task 2 -->
**Commit:** `docs(myflow-contracts): make updatedAt CLI-owned and constrain merge-base values`
**Build:** green
