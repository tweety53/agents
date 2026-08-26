# Tasks

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

`findings` rows already exist in the store (KAN-258). `check-unfinished-work.sh` signal two and all
of `check-panel-reproducers.sh` still re-derive status/reproducer facts by grep/awk over
`docs/superpowers/reviews/*-panel.md`, the Markdown *rendering* of those rows. Tasks 1-2 add
write-time validation in Go; 3-4 add a new read verb returning the rows as JSON; 5-6 rewrite
`check-panel-reproducers.sh`; 7-8 rewrite `check-unfinished-work.sh`'s findings signal; 9 deletes
the marker-parsing library neither guard needs any more and repoints its citations. All commands
below run from the repository root unless a `cd` is shown.

- [x] 1. Write-time validation tests for `record finding` / `record status`

  Add test functions to `stats/cmd/myflow/record_test.go`, modelled directly on the existing
  `TestRecordRejectsUnknownRoleWithoutContactingStore` in the same file: a `contacted` flag on the
  fake server's handler, asserted `false`; exit code 2; stderr naming the accepted vocabulary; no
  journal entry written. Cover, for both `record finding` and `record status` where applicable: an
  unknown status (not `open`/`fixed`/`withdrawn ...`) rejected before the store is contacted; a bare
  `withdrawn` or `withdrawn   ` (whitespace only, no reason) rejected the same way; an empty
  `-reproducer` on `record finding` rejected; a bare `none` or `none  ` reproducer (no reason)
  rejected; and, as the control case, a well-formed call (`-status open`/`-reproducer
  "scripts/x.sh"`, or `-status "withdrawn duplicate of F1"`/`-reproducer "none — duplicate"`) that
  still reaches the fake server. Run `cd stats && go test ./cmd/myflow -run TestRecordFinding
  -count=1` and confirm it fails — every new case reaches the fake server today because nothing
  rejects it yet.

**Build:** green
**Files:** `stats/cmd/myflow/record_test.go`
**Tests:** `TestRecordFindingRejectsUnknownStatus`, `TestRecordFindingRejectsWithdrawnWithNoReason`, `TestRecordFindingRejectsEmptyReproducer`, `TestRecordFindingRejectsBareNoneReproducer`, `TestRecordFindingAcceptsWellFormedStatusAndReproducer`, `TestRecordStatusRejectsUnknownStatus`, `TestRecordStatusRejectsWithdrawnWithNoReason`
**Regression:** reverting this task removes the only assertions that a malformed status or reproducer is rejected before it reaches the store; a bad write would once again reach the fake server unchallenged, which every test above directly checks via its `contacted` flag.
**Baseline:** before=73 after=80 <!-- measured: cd stats && go test ./cmd/myflow -list '.*' | grep -c '^Test' @ branch main -->
**Commit:** `test(myflow): assert record finding/status reject malformed status and reproducer`

- [x] 2. `validateFindingStatus` / `validateFindingReproducer` and their call sites

  In `stats/cmd/myflow/record.go`, add `validateFindingStatus(status string) error` — legal iff
  `status == "open"`, `status == "fixed"`, or `status` starts with `withdrawn` followed by at least
  one non-space character elsewhere in the string; anything else is an error naming the accepted
  vocabulary — and `validateFindingReproducer(reproducer string) error` — empty is an error; a value
  whose first word is exactly `none` is legal only when followed by `" — "` and at least one
  non-space character, a bare `none` is an error naming the required `none — <reason>` form, and any
  other non-empty value (a command) passes, since the command's own shape (metacharacters, absolute
  paths, `..` segments) stays a guard-side check per this change's `reproducer-safety-in-shell`
  decision. Call both from `runRecordFinding`, and `validateFindingStatus` from `runRecordStatus`,
  rejecting with exit 2 before `callRecord` is invoked — the same ordering `-role`'s
  `slices.Contains` check already uses in `runRecordDispatchBegin`. Update the `-status` and
  `-reproducer` flags' own help text to name the accepted vocabulary, the way `-role`'s help text
  already does. Run `cd stats && gofmt -w . && go test ./cmd/myflow -count=1 && go vet ./...` and
  confirm it passes.

**Build:** green
**Files:** `stats/cmd/myflow/record.go`
**Allowed-collateral:** `stats/cmd/myflow/record_test.go` — three pre-existing tests
(`TestRecordWritePrintsOneLineAndExitsZero`'s `finding created`/`finding updated` subtests,
`TestRecordWriteFallsBackToJournalAndExitsZero`'s `finding` case) call `record finding` with no
`-reproducer` flag, which `validateFindingReproducer`'s "empty is an error" rule now rejects. Add a
`-reproducer` value to each of those three call sites — this is a direct, in-scope consequence of
this task's own rule, not scope creep: every finding the pipeline has ever written down was already
expected to carry a reproducer (the retired shell guard's `MISSING`/`EXTRA` check compared the
finding-status and finding-reproducer identifier sets as a bijection), so these three tests were
exercising a state a real panel record could never legally reach; they were just never updated to
say so explicitly.
**Tests:** none named — task 1 carries the new ones; the three collateral edits above keep existing tests passing rather than adding new ones
**Regression:** reverting this task makes every case from task 1 fail again — the store would once more accept a malformed status or reproducer.
**Baseline:** before=80 after=80; all pass.
**Commit:** `feat(myflow): validate finding status and reproducer before writing`

- [x] 3. `record findings` read verb — tests

  Add test functions to `stats/cmd/myflow/record_test.go` covering the not-yet-existing `myflow
  record findings -change <name> [-C dir] [-addr] [-timeout]` verb: an existing change's findings
  (fake server answering `GetRunRecord` with two findings, the same shape
  `TestRecordRenderWritesBothFilesAndPrintsRendered` already stubs) print as a JSON array on stdout,
  decodable into objects carrying `ref`, `status`, and `reproducer`; a change with no rows (fake
  server answering `client.ErrNotFound`, mirroring `TestRecordRenderWithNoLedgerRowsPrintsMissingAndWritesNothing`)
  prints exactly `[]` and exits 0; an unreachable store (`-addr` pointed at a closed port) exits
  non-zero with no JSON array on stdout — a read has nothing to journal, so it must never report
  success for a question it could not answer; and omitting `-change` exits 2 with the shared
  `requireRecordFlags` message. Run `cd stats && go test ./cmd/myflow -run TestRecordFindings
  -count=1` and confirm it fails with `unknown command` or equivalent, since `findings` is not yet a
  case in `runRecord`'s switch.

**Build:** green
**Files:** `stats/cmd/myflow/record_test.go`
**Tests:** `TestRecordFindingsPrintsJSONArray`, `TestRecordFindingsWithNoRowsPrintsEmptyArray`, `TestRecordFindingsStoreUnreachableExitsNonZero`, `TestRecordFindingsRequiresChange`
**Regression:** reverting this task removes every assertion that the new verb exists, prints valid JSON, and refuses to fabricate a result when the store cannot be reached.
**Baseline:** before=80 after=84 <!-- measured: cd stats && go test ./cmd/myflow -list '.*' | grep -c '^Test' @ branch main -->
**Commit:** `test(myflow): assert the record findings verb's JSON output and failure modes`

- [x] 4. `runRecordFindings` and the usage block

  Add `runRecordFindings` to `stats/cmd/myflow/record.go`: register `recordIdentityFlags`, resolve
  `projectKey` exactly as `runRecordRender` does, call `cl.GetRunRecord`, and on success
  `json.Marshal` the `Run.Findings` slice alone (never the whole `Run` — a guard needs findings
  only) to stdout. `errors.Is(err, client.ErrNotFound)` prints `[]` and exits 0, matching how
  `runRecordRender` treats "no rows for this run." Any other error prints one line to stderr and
  returns a non-zero exit — no journal write, per this change's `new-findings-verb` decision, since
  a failed read has nothing to replay. Wire `"findings"` into `runRecord`'s switch and into
  `main.go`'s usage constant. Run `cd stats && gofmt -w . && go test ./cmd/myflow -count=1 && go vet
  ./...` and confirm it passes.

**Build:** green
**Files:** `stats/cmd/myflow/record.go`, `stats/cmd/myflow/main.go`
**Tests:** none named — task 3 carries them
**Regression:** reverting this task makes every case from task 3 fail again — `record findings` would not exist.
**Baseline:** before=84 after=84; all pass.
**Commit:** `feat(myflow): add the record findings read verb`

- [x] 5. `check-panel-reproducers.sh` reads the store — test fixtures

  In `scripts/test-check-panel-reproducers.sh`, replace every case's Markdown-fixture setup with a
  stub `myflow` script placed ahead of the real one on `PATH` inside the test's own sandbox,
  printing a canned JSON array for `record findings -change <name>` (or a non-zero exit for the
  store-unreachable case) — each case's *assertion* (verdict line, exit code, named violation) is
  unchanged; only how the finding data reaches the guard changes. Add a case for the
  store-unreachable path: stub `myflow` exits non-zero for `record findings`, and
  `check-panel-reproducers.sh` must exit 2 ("cannot determine anything"), never 1 or 0. Drop the
  cases that exist only to test Markdown-parsing failure modes with no JSON equivalent: the
  checksum-mismatch, marker-span, table/marker-agreement, and NUL-byte cases — a `jq` decode failure
  has one shape, not the six the hand-rolled parser had to distinguish. Run
  `scripts/test-check-panel-reproducers.sh` and confirm it fails, since the guard still reads
  `docs/superpowers/reviews/` and the new fixtures produce no file there for it to find.

**Build:** green
**Files:** `scripts/test-check-panel-reproducers.sh`
**Tests:** none named — this task rewrites existing shell test cases rather than adding Go/Python test functions; verified by the file's own pass/fail summary line
**Regression:** reverting this task leaves the test file asserting behavior the still-Markdown-based guard already had — it passes vacuously against the old guard and fails loudly against the new one, which is what this task's own final run confirms in the other direction.
**Baseline:** before=67 after=<case count after the rewrite and the drops> — unverified:the exact post-rewrite case count depends on how many of the retired failure-mode cases are dropped, decided during the rewrite itself
**Commit:** `test(scripts): assert check-panel-reproducers.sh reads findings from the store`

- [x] 6. `check-panel-reproducers.sh` — implementation

  In `scripts/check-panel-reproducers.sh`, delete the `source .../lib/panel-record.sh` line, the
  `panel_record_path` call, the NUL-byte scan, and every `count_matching`/`grep_lines_of`/`ids_of`
  call. Replace panel-record resolution with `FINDINGS_JSON="$(myflow record findings -change
  "$NAME" -C "$WORKTREE")"`; a non-zero exit from that call becomes this guard's own exit 2, worded
  the way the retired "no readable panel record" message was. Re-derive `STATUS_IDS`/`REPRO_IDS`
  with `jq -r '.[].ref'` (already unique — `findings_ref_key` is a store constraint, so the
  duplicate-identifier checks this guard used to run are unreachable and are deleted, not kept as
  dead code); check for a finding with an empty/null reproducer with `jq` as a refusal (exit 2)
  rather than an assumed impossibility, even though task 2's write-time validation should make it
  unreachable; keep the command-safety loop (metacharacters, leading `-` on the path token, absolute
  path, `..` segment — the current file's lines 249-298) unchanged in substance, reading `cmd_text`
  from the decoded JSON field instead of a `grep`-matched line. Retire the
  `reproducers-total`/span/table-agreement machinery entirely — the JSON array's length is the
  count, so the closing line becomes `REPRODUCERS-OK (<n> finding(s) declared)` from `jq 'length'`
  on the findings that carry a reproducer. Update the file's header comment: the "Reads ONLY the two
  marker blocks" paragraph and the exit-2 description move from "unreadable file" to
  "store unreachable"; the exit-code numbers (0/1/2) stay the same. Run
  `scripts/test-check-panel-reproducers.sh && scripts/check-vocabulary.sh &&
  scripts/check-references.sh` and confirm all three pass.

**Build:** green
**Files:** `scripts/check-panel-reproducers.sh`
**Allowed-collateral:** `scripts/test-check-panel-reproducers.sh` — a stub-generator bug in
`make_worktree_json` (task 5's own code): the JSON payload was interpolated into a single-quoted
`printf` argument inside a heredoc, so a reproducer value under test carrying a literal single quote
(case 17's apostrophe metachar) produced syntactically invalid bash regardless of the guard's
implementation. Fixed by writing the JSON to its own file and having the stub `cat` it. No case's
assertion changes — this is a defect in the test's own fixture plumbing, not a behavior or spec
change, uncovered only once the guard actually started calling the stub.
**Tests:** none named — task 5 carries them; the collateral fix above keeps case 17 correct rather than adding a new one
**Regression:** reverting this task makes every case from task 5 fail again — the guard would once more look for a Markdown file the new fixtures never write.
**Baseline:** before=<count from task 5> after=<same count>; all pass.
**Commit:** `feat(scripts): read panel reproducers from the store, not rendered markdown`

- [x] 7. `check-unfinished-work.sh` signal two — test fixtures

  In `scripts/test-check-unfinished-work.sh`, replace every signal-two case's Markdown-fixture setup
  with a stub `myflow` printing the equivalent JSON array, exactly as task 5 does for the
  reproducers guard — signal one (the `tasks.md` checkbox count) and its fixtures are untouched.
  Each case's asserted verdict line (`CLEAR: ...` / `OUTSTANDING: ...`) is unchanged. Add a case for
  the store-unreachable path: stub `myflow` exits non-zero, and `check-unfinished-work.sh` must exit
  2, never 0 (a signal it cannot evaluate must never read as `CLEAR`) and never 1. Drop the
  checksum/marker-span/table-agreement/duplicate-identifier cases — same reasoning as task 5. Run
  `scripts/test-check-unfinished-work.sh` and confirm it fails, since signal two still reads
  `docs/superpowers/reviews/`.

**Build:** green
**Files:** `scripts/test-check-unfinished-work.sh`
**Tests:** none named — this task rewrites existing shell test cases; verified by the file's own pass/fail summary line
**Regression:** reverting this task leaves the test file asserting behavior only the new guard has — it fails against the still-Markdown-based guard, confirming the tests changed for a reason.
**Baseline:** before=136 after=<case count after the rewrite and the drops> — unverified:the exact post-rewrite case count depends on how many of the retired failure-mode cases are dropped, decided during the rewrite itself
**Commit:** `test(scripts): assert check-unfinished-work.sh's findings signal reads the store`

- [x] 8. `check-unfinished-work.sh` signal two — implementation

  In `scripts/check-unfinished-work.sh`, delete `source .../lib/panel-record.sh`, `panel_record_path`,
  and every `count_matching`/`ids_of`/`line_numbers` call signal two uses — signal one's
  `count_unticked` call is untouched, since it never sourced the library. A non-zero exit from
  `myflow record findings -change "$NAME" -C "$WORKTREE"` becomes signal two's own `unreadable`-style
  refusal (exit 2). Re-derive the open-finding count with `jq '[.[] | select((.status != "fixed")
  and (.status | startswith("withdrawn") | not))] | length'` over the JSON array. Retire the
  `M_UNRECOGNISED` and `M_NOREASON` branches entirely — task 2's write-time validation guarantees
  every stored status is `open`, `fixed`, or `withdrawn <reason>`, so neither malformed shape can
  reach this guard any more. Retire the checksum/span/table-agreement/duplicate-identifier code the
  same way task 6 retires it for reproducers. Update the header comment: the block cataloguing the
  six historical table-parser defects (current lines 278-346) becomes a short note that the guard
  now reads store rows through `myflow record findings`, so none of that document-grammar
  re-implementation exists to have those defects — keep the "Signal two" framing and the "MISSING
  FILE COUNTS AS OUTSTANDING" framing above it, both of which still hold against a JSON answer
  rather than a file's presence. Run `scripts/test-check-unfinished-work.sh &&
  scripts/check-vocabulary.sh && scripts/check-references.sh` and confirm all three pass.

**Build:** green
**Files:** `scripts/check-unfinished-work.sh`
**Allowed-collateral:** `scripts/test-check-unfinished-work.sh` — case 11 ("a missing library is a
refusal") copies the guard alone (without `lib/panel-record.sh`) and asserts it exits 2 naming the
missing file. Once this task removes signal two's `source` of that library, the guard no longer
depends on any sibling file at all — the scenario the case exercises can no longer occur, and the
case's own header comment already says so ("task 8 is what removes it"). Delete the whole case
(comment header and all three assertions) rather than repointing it: there is no remaining
sourced-dependency scenario for this guard to test in its place, unlike
`check-panel-reproducers.sh`, which still depends on `reproducer-metachars.sh`.
**Tests:** none named — task 7 carries them; the case-11 deletion above removes a now-unreachable scenario rather than adding one
**Regression:** reverting this task makes every case from task 7 fail again.
**Baseline:** before=<count from task 7> after=<count from task 7, minus case 11>; all pass.
**Commit:** `feat(scripts): read unfinished-findings signal from the store, not rendered markdown`

- [x] 9. Retire `scripts/lib/panel-record.sh` and repoint its dependents

  Delete `scripts/lib/panel-record.sh` — after tasks 6 and 8 land, nothing sources it; confirm with
  `grep -rl "panel-record.sh\|panel_record_path\|PANEL_RECORD_DIR" scripts/*.sh` naming only the
  file itself before deletion. `scripts/check-guard-symlinks.sh` and `scripts/test-lib-coverage.sh`
  each credit `panel-record.sh`'s header for the `-a`/`rc > 1`/`--` grep disciplines it originated —
  reword each citation to state the discipline inline (a sentence, not a file reference) rather than
  point at a deleted file. `scripts/test-setup.sh`'s "check-unfinished-work.sh runs through the
  installed path and finds its lib/" group currently exercises `lib/panel-record.sh` specifically as
  "the guard with the hardest dependency" — repoint it at `check-panel-reproducers.sh`'s still-live
  dependency on `reproducer-metachars.sh` (guard path, expected usage-on-no-args behavior, and the
  "cannot determine the banned-character set" string in place of "cannot determine the marker helper
  definitions"), since the installed-symlink-chain assertion the group makes is still worth making
  against some guard with a sourced dependency. `skills/myflow-contracts/pipeline-rationale.md`'s
  **Guard presence check** section names `` `check-unfinished-work.sh` needing
  `<agents repo>/scripts/lib/panel-record.sh` `` as one of three examples of a guard resolving a
  sibling from its own `$SCRIPT_DIR` — replace that example with `` `check-panel-reproducers.sh`
  needing `<agents repo>/scripts/reproducer-metachars.sh` ``, leaving the other two examples
  (`check-task-commit-fields.sh`/`check-task-commit-fields.py`,
  `prepare-workspace.sh`/`check-workspace-isolation.sh`) unchanged. Run
  `scripts/check-references.sh`, `scripts/check-guard-symlinks.sh`, `scripts/test-lib-coverage.sh`,
  and `scripts/test-setup.sh`, and confirm all four exit 0/PASS with no reference to
  `panel-record.sh` remaining anywhere in the repository outside this change's own
  `spectre/changes/` directory and `git log`.

**Build:** green
**Files:** `scripts/lib/panel-record.sh` (deleted), `scripts/check-guard-symlinks.sh`, `scripts/test-lib-coverage.sh`, `scripts/test-setup.sh`, `skills/myflow-contracts/pipeline-rationale.md`
**Tests:** none named — `scripts/check-guard-symlinks.sh`'s and `scripts/test-lib-coverage.sh`'s own existing cases continue to pass against the reworded citations (prose, not a checked fact, so no new case is added); `scripts/test-setup.sh`'s repointed "hardest dependency" group is the one behavioral change, verified by that file's own pass/fail summary
**Regression:** reverting this task leaves `scripts/lib/panel-record.sh` in the tree with nothing sourcing it — the next unrelated run of `check-references.sh` or `check-guard-symlinks.sh` flags it as a dead file.
**Baseline:** before=337 after=337 (`test-setup.sh` assertions); before=25 after=25 (`test-lib-coverage.sh` cases); before=82 after=82 (`test-check-guard-symlinks.sh` cases); all pass — this task reworks existing assertions, it adds none. <!-- measured: scripts/test-setup.sh | tail -1 @ branch main -->
**Commit:** `chore(scripts): retire panel-record.sh and repoint its citations`
