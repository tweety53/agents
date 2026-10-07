# kan-877-flowd-burns-80-cpu-while-idle-ish

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** stop flowd's harvest watcher from re-reading every transcript on every cycle.
`design.md` is canonical for the three decisions and the measurements; `proposal.md` for why. Nothing here
restates them.

**Global constraints.**

- Every edit is in `stats/internal/harvest/watcher.go` and its tests; no new dependency, no store or schema
  change, no change to `cmd/flowd`.
- What binds, what gives up, and what is attributed must not change: every existing test in
  `stats/internal/harvest/` stays green with its expectations untouched.
- Tests are black-box (`package harvest_test`); an unexported helper a test needs is exposed through
  `stats/internal/harvest/export_test.go`, the shape that file already uses.
- Verify per task: `cd stats && gofmt -l internal/harvest && go vet ./internal/harvest` print nothing, then
  the task's own `-run` selector with `-race -count=1`.

**Baseline, measured before any edit:** `stats/internal/harvest/watcher_test.go` declares 47 tests.
<!-- measured: grep -c '^func Test' stats/internal/harvest/watcher_test.go @ merge-base 8178b32b (before this change) -->

**Review focus.**

- A retried token whose mark sits in bytes already behind the committed offset still binds on the first
  cycle (`TestPersistedGiveUpBindsFromAFullyConsumedTranscript`, `TestScanRetriedTokensReadsRolloutCommands`).
- A file vanishing between discovery and the retried scan does not keep the scan alive forever.
- A command carrying two `-session-token` flags, or a quoted token, matches exactly as before.
- A transcript that is truncated or replaced by a shorter one is read again, never skipped.
- A late dispatch-meta sidecar is still backfilled on an otherwise unchanged file
  (`TestBackfillsDispatchMetaWhenSidecarArrivesLate`).

---

- [x] 1. Scan already-read bytes for retried tokens once per process

  - [x] **Step 1: Write the failing tests** in `stats/internal/harvest/watcher_test.go`, beside
    `TestScanRetriedTokensReadsRolloutCommands`. Both wrap a Claude source's `ReadAllCmds` with a counter
    and seed one persisted give-up whose token appears in no transcript, so the token stays pending for
    every cycle driven.

```go unverified:compile against watcher_test.go's countingSessionTokenBinder, fakeWindowSource and newFakeHarvestSink
func TestRetriedTokenScanReadsEachTranscriptOnce(t *testing.T) {
	dir := t.TempDir()
	const token = "mf-scan-once"
	binder := &countingSessionTokenBinder{sessionToken: token, stageRunID: 811}
	binder.seededGiveUps = []harvest.GiveUp{{Token: token, Reason: "session-never-bound", Retries: 1}}
	path := filepath.Join(dir, "unrelated.jsonl")
	line := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-other","message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"echo nothing to see"}}]}}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	src := harvest.NewClaudeSource(dir)
	inner := src.ReadAllCmds
	reads := 0
	src.ReadAllCmds = func(p string) ([]harvest.CommandRecord, error) {
		reads++
		return inner(p)
	}
	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
	w := harvest.NewWatcher([]harvest.Source{src}, newFakeHarvestSink(), harvest.NewAttributor(windows), sessionBinderDeps{binder: binder}, nil)
	for i := range 5 {
		if _, err := w.RunOnce(context.Background()); err != nil {
			t.Fatalf("RunOnce (cycle %d): %v", i, err)
		}
	}
	if reads != 1 {
		t.Fatalf("ReadAllCmds called %d times over 5 cycles, want 1: bytes behind a committed offset never change, so one whole-file pass per process is enough", reads)
	}
}
```

    `TestRetriedTokenScanRerunsAfterReadError` has the same setup and two subtests over the counter's
    return value on its **first** call only (every later call delegates to `inner`), each driving 3 cycles:
    - `transient error reruns` — first call returns `errors.New("transient")` → want `reads == 2`.
    - `vanished file counts as scanned` — first call returns
      `fmt.Errorf("harvest: open %s: %w", p, fs.ErrNotExist)` → want `reads == 1`.

  - [x] **Step 2: Run them and see both fail** —
    `cd stats && go test ./internal/harvest -run 'TestRetriedTokenScanReadsEachTranscriptOnce|TestRetriedTokenScanRerunsAfterReadError' -count=1`;
    expect `reads` to equal the cycle count today.

  - [x] **Step 3: Implement** in `stats/internal/harvest/watcher.go`: a `retriedScanDone bool` field on
    `Watcher`; `scanRetriedTokens` returns at once when it is set, and sets it after a pass in which every
    `ReadAllCmds` error satisfied `errors.Is(err, fs.ErrNotExist)` (the warning is still logged for every
    error). Leave the early returns on an empty `retriedTokens` or empty `toScan` as they are — they do
    not set the flag. Rewrite the method's doc comment where it says the scan runs every cycle, citing
    `retried-scan-once-per-process`.
    measured: `ReadAllCommands` (transcript.go:691) and `ReadRolloutAllCommands` (rollout.go:288) both
    wrap the `os.Open` error with `%w`, so `errors.Is(err, fs.ErrNotExist)` sees a vanished file; neither
    file changed. <!-- measured: sed -n on both files @ f9d09199 -->

  - [x] **Step 4: Verify** — `cd stats && gofmt -l internal/harvest && go vet ./internal/harvest` print
    nothing; `cd stats && go test ./internal/harvest -run 'TestRetriedTokenScan|TestPersistedGiveUp|TestRetryStillBounded|TestScanRetriedTokensReadsRolloutCommands' -race -count=1` passes.

  - [x] **Step 5: Commit** with the subject below.

**Files:** `stats/internal/harvest/watcher.go`, `stats/internal/harvest/watcher_test.go`
**Tests:** `TestRetriedTokenScanReadsEachTranscriptOnce`, `TestRetriedTokenScanRerunsAfterReadError`
**Regression:** reverting returns the scan to every cycle: `TestRetriedTokenScanReadsEachTranscriptOnce`
  counts 5 reads, and the transient subtest of `TestRetriedTokenScanRerunsAfterReadError` counts 3.
**Baseline:** before=47 after=49
<!-- measured: grep -c '^func Test' stats/internal/harvest/watcher_test.go @ merge-base 8178b32b (before this change) -->
**Commit:** `perf(harvest): scan retried give-up tokens once per process`
**After:** none
**Build:** green

**Decision:** retried-scan-once-per-process

Correction (2026-10-07): the plan declared `transcript.go` and `rollout.go` in case either read function
failed to wrap its `os.Open` error with `%w`; both already do, so neither file changed and both left
`**Files:**`. The tests also share a `newRetriedScanCounter` helper rather than repeating the setup.

Correction (2026-10-07): review panel round 0 (F1) found the pass marked done when its files were read
rather than when its matches were acted on — a failed `BindSession` or a failed ambiguity give-up record
then retried the token from an empty match set. Fix commit `fix(harvest): rescan retried tokens after a
failed bind or refusal` (2c38a406, on the branch tip, not folded) adds `rescanIfRetried` and
`TestRetriedTokenScanRerunsAfterFailedResolve`, taking the file's test count to 53.

- [x] 2. Match commands against a token set

  - [x] **Step 1: Expose the matcher** in `stats/internal/harvest/export_test.go`:

```go unverified:compile; matchSessionTokens is a method on *Watcher
// MatchSessionTokensForTest exposes matchSessionTokens to harvest_test's
// black-box tests -- its many-tokens contract is pinned more directly by
// calling it straight than by driving a binder per token through RunOnce.
var MatchSessionTokensForTest = (*Watcher).matchSessionTokens
```

  - [x] **Step 2: Write the test** `TestManyPendingTokensEachBindToTheirOwnMark` in
    `stats/internal/harvest/watcher_test.go`: 200 pending tokens `mf-t0`…`mf-t199`; commands
    `flow stage begin -session-token mf-t7` (session `s-a`), `flow stage end --session-token='mf-t42'`
    (session `s-b`), `flow stage mark -session-token=mf-t42 x` (session `s-c`),
    `grep mf-t9 log.txt` (session `s-d`, a mention), `echo -session-token mf-t11` (session `s-e`, no mark
    verb). Call `harvest.MatchSessionTokensForTest(harvest.NewWatcher(nil, newFakeHarvestSink(), harvest.NewAttributor(&fakeWindowSource{}), harvest.NoDeps{}, nil), pending, commands, matched)`
    and assert: it returns true; `matched` holds exactly `mf-t7 → {s-a}` and `mf-t42 → {s-b, s-c}`.

  - [x] **Step 3: Run it** — `cd stats && go test ./internal/harvest -run 'TestManyPendingTokensEachBindToTheirOwnMark' -count=1`.
    It passes on today's O(commands × tokens) loop too: it pins the behaviour the rewrite must keep.

  - [x] **Step 4: Implement** in `stats/internal/harvest/watcher.go`: `matchSessionTokens` builds
    `want := map[string]bool` from `pending`'s values once; per command it returns early on
    `!stageMarkInvocationPattern.MatchString`, then walks `strings.Fields` once, collecting every value
    the three `isSessionMarkCommand` cases accept (after `trimTokenQuotes`), and records each one in
    `want`. Replace `isSessionMarkCommand` with that extractor (`sessionTokensInMark(command string) []string`)
    — it has no other caller — and carry its doc comment's mention-versus-invocation reasoning over,
    citing `match-by-token-set`.

  - [x] **Step 5: Verify** — gofmt/vet clean; `cd stats && go test ./internal/harvest -run 'TestManyPendingTokensEachBindToTheirOwnMark|TestCommandMerelyMentioningTokenDoesNotBind|TestMentionAfterOwnMarkIsConsumedDoesNotMisattribute|TestMarkRecognizedWhereverItSitsInTheCommand|TestCommandsThatOnlyMentionTokenNeverBind|TestEchoedMarkExampleIsAnAcceptedResidual|TestCrossedSessionTokensBindEachRunToItsOwnSession' -race -count=1` passes.

  - [x] **Step 6: Commit** with the subject below.

**Files:** `stats/internal/harvest/watcher.go`, `stats/internal/harvest/watcher_test.go`, `stats/internal/harvest/export_test.go`, `stats/cmd/flow/stage.go`
**Tests:** `TestManyPendingTokensEachBindToTheirOwnMark`
**Regression:** the test pins behaviour, not cost; reverting the implementation keeps it green, while a
  rewrite that drops the quoted, `=`-joined or double-dash forms, or binds a bare mention, fails it.
**Baseline:** before=49 after=50
<!-- predicted: grep -c '^func Test' stats/internal/harvest/watcher_test.go after task 1 -->
**Commit:** `perf(harvest): match session tokens through a set`
**After:** Task 1
**Build:** green

**Decision:** match-by-token-set

Correction (2026-10-07): `stats/cmd/flow/stage.go` cites the harvester's matcher by name in a comment, so
the rename to `sessionTokensInMark` touched it too and it joined `**Files:**`.

- [x] 3. Skip a transcript whose size equals its last-seen committed offset

  - [x] **Step 1: Write the failing tests** in `stats/internal/harvest/watcher_test.go`, using a sink that
    embeds `*fakeHarvestSink` and counts `GetHarvestOffset` calls per path:
    - `TestUnchangedTranscriptSkipsOffsetLookup` — one transcript (the assistant line from task 1's test);
      cycle 1 → 1 lookup; cycles 2 and 3 → still 1 lookup in total.
    - `TestGrownTranscriptIsReadAfterSkip` — after two cycles, append a second assistant line; cycle 3 → a
      second lookup, and `sink.offsets[path]` equals the file's new size. Then truncate the file to its
      first line; cycle 4 → a third lookup (truncation is read, never skipped).

  - [x] **Step 2: Run them and see the first fail** —
    `cd stats && go test ./internal/harvest -run 'TestUnchangedTranscriptSkipsOffsetLookup|TestGrownTranscriptIsReadAfterSkip' -count=1`.

  - [x] **Step 3: Implement** in `stats/internal/harvest/watcher.go`: a `seenOffsets map[string]int64`
    field, made in `NewWatcher`. In `RunOnce`'s per-file loop, before `GetHarvestOffset`: `os.Stat(path)`;
    when it succeeds and `seenOffsets[path] == info.Size()` with the key present, call
    `w.maybeBackfillDispatchMeta(ctx, path)` and `continue`. Record `seenOffsets[path] = offset` on the
    `newOffset == offset` branch, and `seenOffsets[path] = newOffset` after `CommitHarvestBatch` reports
    applied. Record nothing on a withheld batch, a failed lookup, read, attribute or commit, or a lost
    race. A failed `Stat` takes the ordinary path. Comment the skip citing `skip-unchanged-by-size`.

  - [x] **Step 4: Verify** — gofmt/vet clean; `cd stats && go test ./internal/harvest -race -count=1` passes
    (the whole package: every existing harvest test is a regression check for this skip).

  - [x] **Step 5: Commit** with the subject below.

**Files:** `stats/internal/harvest/watcher.go`, `stats/internal/harvest/watcher_test.go`
**Tests:** `TestUnchangedTranscriptSkipsOffsetLookup`, `TestGrownTranscriptIsReadAfterSkip`
**Regression:** reverting makes every cycle query every file again: `TestUnchangedTranscriptSkipsOffsetLookup`
  counts 3 lookups.
**Baseline:** before=50 after=52
<!-- predicted: grep -c '^func Test' stats/internal/harvest/watcher_test.go after task 2 -->
**Commit:** `perf(harvest): skip transcripts unchanged since their last committed offset`
**After:** Task 1, 2
**Build:** green

**Decision:** skip-unchanged-by-size
