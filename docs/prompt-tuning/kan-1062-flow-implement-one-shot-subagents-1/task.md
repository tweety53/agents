- [x] 3. Number a findings file's ref-less lines from the change's next free ref
  - [x] **Step 1: Write the test.** `stats/cmd/flow/record_test.go`:
    `TestRecordFindingFromAssignsRefs`, on `TestRecordFindingFromFile`'s fixture with a fake daemon
    whose run record for the change holds findings `F1`–`F3`. (a) A file of two lines carrying
    `task`, `severity`, `location` and `note` and no `ref`, loaded with
    `-slot gated-reviewer -round 0 -status open -reproducer none -dispatch-seq 4 -from <file>`,
    posts `F4` then `F5`, each with slot `gated-reviewer`, round 0, dispatch-seq 4 and its own line's
    fields, and the file now carries `"ref":"F4"` and `"ref":"F5"` on those lines. (b) Loading the
    rewritten file again posts `F4` and `F5` again and no `F6`. (c) A file whose first line names
    `"ref":"F9"` and whose second names none posts the second as `F10`. (d) With the run-record read
    answering 500, a ref-less file exits 1, posts nothing, and its bytes are unchanged.
    `TestRecordFindingFromFile` still passes: a file whose every line names its ref reads no run
    record. Run `cd stats && go test ./cmd/flow -run 'TestRecordFindingFrom' -count=1` — fails.
  - [x] **Step 2: Implement.** `stats/cmd/flow/record.go`, `runRecordFindingFrom`: when any line
    carries no `ref`, read the change's run record once, as `runRecordNextRef` does through
    `readNextRecord`, with the command line's `-addr`, `-timeout`, `-C` and `-change`; number every
    ref-less line in file order from one above the higher of the store's highest `F<n>` and every
    `F<n>` the file names; rewrite those lines of the file with their refs before the first is
    recorded, leaving every other line's bytes as they were; then record as today. A failed read
    exits 1 naming the file, before anything is posted or written. Its comment and the `-from` usage
    line state the numbering (decision `from-assigns-refs`).
  - [x] **Step 3: Prove.** `scripts/break-and-prove.sh` dropping the write-back — case (b) of
    `TestRecordFindingFromAssignsRefs` fails on an `F6`; restored, it passes.
  - [x] **Step 4: Verify.** `cd stats && go test ./cmd/flow -run 'TestRecordFinding|TestRunRecordNextRef' -count=1 | tail -20`
    passes; `cd stats && gofmt -l cmd/flow` prints nothing; `cd stats && make -s web-dist && go vet ./cmd/flow`
    exits 0. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordFindingFromAssignsRefs`
**Regression:** `TestRecordFindingFromAssignsRefs` fails — a gated reviewer's findings file is refused for its missing refs, or a re-run records every finding a second time under new refs.
**Commit:** feat(stats): number a findings file's ref-less lines from the change's next free ref
**After:** Task 2

**Decision:** from-assigns-refs

