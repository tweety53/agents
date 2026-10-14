# self-review-kan-980-no-flow-roles

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; never write `spectre/changes/`.
- Every prose block below is copied verbatim, line breaks aside: this change's `verbatim-moves.txt` acknowledges each new sentence by its exact text, and a reworded sentence fails `scripts/check-verbatim-moves.sh`.
- Go in `stats/` follows the Go skills `CLAUDE.md` names; run `cd stats && gofmt -w .` before `gofmt -l .`.
- No test contacts the dev workspace's `flowd`: a `flow record` test that could reach the store passes `-addr` explicitly.

- [x] 1. `flow record pass` refuses a note that opens with a marker label
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Add the test of step 2 first and run it: every subtest exits 0 with `⚠ flow: store unreachable — wrote local journal` on stderr — the defect: a marker line is accepted as a pass note.
  - [x] **Step 2: Write the failing test.** Append to `stats/cmd/flow/record_test.go`, after `TestRecordFindingRejectsMalformedRef`:

```go verified:applied with step 3's code at ff835bde — fails without the refusal (exit 0, journal written), passes with it; go vet and gofmt -l clean
// TestRecordPassRefusesMarkerNote pins the kan-980 self-review fix: a pass
// note opening with a marker label is a marker line recorded in the wrong
// place -- the renderer neutralises its colon, so the round loses the
// structural row it meant to carry -- and `flow record pass` refuses it
// before the store is contacted, naming the command that records it. The
// -addr is a closed port, so a regressed refusal falls through to the
// journal instead of reaching a live daemon.
func TestRecordPassRefusesMarkerNote(t *testing.T) {
	for _, note := range []string{
		"fix-mutation: stats/x.go — dropped the guard — TestX",
		"  finding-status: F9 fixed",
	} {
		t.Run(note, func(t *testing.T) {
			repo := gitRepo(t)
			isolatedStateRoot(t)

			var stdout, stderr bytes.Buffer
			code := run(context.Background(),
				[]string{"record", "pass", "-addr", "http://127.0.0.1:1", "-timeout", "500ms", "-C", repo,
					"-change", "kan-258", "-round", "4", "-note", note},
				strings.NewReader(""), &stdout, &stderr)
			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
			}
			if !strings.Contains(stderr.String(), "flow record mutation") || !strings.Contains(stderr.String(), "flow record finding") {
				t.Errorf("stderr = %q, want it to name flow record mutation and flow record finding", stderr.String())
			}
			if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
				t.Errorf("a refused marker note left a record journal behind")
			}
		})
	}
}
```

    Run `cd stats && go test ./cmd/flow/ -run 'TestRecordPassRefusesMarkerNote' -count=1` and confirm both subtests FAIL.
  - [x] **Step 3: Export the label list.** In `stats/internal/records/render.go` rename `markerLabels` to `MarkerLabels` (its doc comment's first word and its one use in `neutraliseMarkers`), and replace the doc comment's last line `// one.` with these two lines:

```go verified:applied at ff835bde — go test ./internal/records/ passes
// one. `flow record pass` reads the list too, to refuse a note that opens
// with one of them.
```

  - [x] **Step 4: Implement the refusal.** In `runRecordPass` (`stats/cmd/flow/record.go`), directly after the `requireRecordFlags(stderr, [2]string{"-note", *note})` block's closing `}` and before `projectKey, _, err := fallback.ProjectKey(f.dir)`, insert:

```go verified:applied with step 2's test at ff835bde — the test passes, every TestRecordPass* passes, go vet and gofmt -l clean
	// A note opening with a marker label is a marker line in the wrong
	// place: the renderer neutralises its colon, so the round would lose
	// the structural row the line meant to be.
	for _, label := range records.MarkerLabels {
		if strings.HasPrefix(strings.TrimSpace(*note), label+":") {
			fmt.Fprintf(stderr, "flow: record pass: -note opens with the marker label %s: — record a fix-mutation line with `flow record mutation` and a finding line with `flow record finding`, never as a pass note\n", label)
			return 2
		}
	}
```

  - [x] **Step 5: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./cmd/flow/ ./internal/records/ && go test ./cmd/flow/ -run 'TestRecordPass' -count=1 && go test ./internal/records/ -count=1` — all exit 0 with `gofmt -l` printing nothing. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/internal/records/render.go`
**Tests:** `TestRecordPassRefusesMarkerNote`
**Regression:** reverting the commit accepts a marker-label note: both subtests exit 0 and leave a journal behind.
**Baseline:** before=98 after=99
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ ff835bde -->
**Commit:** fix(flow): refuse a pass note that opens with a marker label
**After:** none

**Decision:** pass-note-refusal-in-cli

- [x] 2. Name the since-close sha after integrate's reshape
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -n "reshape\|archive commit" skills/flow/review-panel-late-fix.md` and confirm no line names the since-close sha once the branch is reshaped — the defect.
  - [x] **Step 2: Add the sentence.** In `skills/flow/review-panel-late-fix.md`, **The late-fix reduction**, directly after the line ending "peer worktree's since-close sha comes from the panel record's per-worktree sha list." and inside the same paragraph (no blank line either side), insert verbatim:

```markdown verified:check-verbatim-moves.sh, check-references.sh, check-installed-citations.sh, check-markdown-integrity.py and check-normative-inventory.sh exit 0 with this sentence and task 3's applied and this change's verbatim-moves.txt present, spiked at ff835bde
**After integrate's reshape** (**3. Commit the staged work**, `skills/flow/integrate.md`) the shas
the last clean close reviewed are gone from the branch, so the since-close sha is the canonical
worktree's archive commit and a peer worktree's reshaped branch head.
```

  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-installed-citations.sh`, `scripts/check-markdown-integrity.py` — all exit 0. Commit.

**Build:** green
**Files:** `skills/flow/review-panel-late-fix.md`
**Tests:** **none** — one prose sentence; the guards in step 3 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): name the since-close sha after integrate's reshape
**After:** none

**Decision:** since-close-after-reshape

- [x] 3. A pending citation send-back holds the stage close
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Read the paragraph opening "Exit 0 proceeds." in `skills/flow/review-panel-fix-round.md` and confirm nothing in it or after it says the round waits for the `-cite` dispatch — the defect.
  - [x] **Step 2: Add the sentence.** In `skills/flow/review-panel-fix-round.md`, directly after the line ending "finding still `UNCITED` goes to the handback below. Exit 2 stops the run." and inside the same paragraph (no blank line either side), insert verbatim:

```markdown verified:check-verbatim-moves.sh, check-references.sh, check-installed-citations.sh, check-markdown-integrity.py and check-normative-inventory.sh exit 0 with this sentence and task 2's applied and this change's verbatim-moves.txt present, spiked at ff835bde
**A pending send-back holds the stage:** the round neither closes nor hands to `flow.verify` until
the `-cite` dispatch has returned and `check-rerun-citations.sh` has been run on its report.
```

  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-installed-citations.sh`, `scripts/check-markdown-integrity.py` — all exit 0. Commit.

**Build:** green
**Files:** `skills/flow/review-panel-fix-round.md`
**Tests:** **none** — one prose sentence; the guards in step 3 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): a pending citation send-back holds the stage close
**After:** none

**Decision:** cite-sendback-holds-close
