# self-review-kan-988-ledger-test-pool-starvation

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-16-self-review-kan-988-ledger-test-pool-starvation-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; never write `spectre/changes/`. Never merge or push.
- Each task's change is the unified diff its step gives, applied with `git apply` from the repository root — every diff was applied and checked there (`design.md`, **Context**). A diff that does not apply is reported, never hand-merged by guesswork.
- Go in `stats/` follows the Go skills `CLAUDE.md` names; run `cd stats && gofmt -w .` before `gofmt -l .`. Each task's verify step vets and tests only the package it touches.

## Review focus

- A reproducer with an unreadable body must still reach the copy step's own refusal, not a new one (`TestProveReproducerBashParity`'s unreadable case, run in Task 1's step 4).
- A message straddling two reads must count its input once and its output to the final figure — never the first line's count plus the last line's in full (Task 3's straddle test).
- A batch with no id'd usage record must keep the previous carry, as `lastUsageMessageID`'s caller did.

- [x] 1. prove-reproducer refuses a reproducer naming the live worktree by absolute path
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -n 'cannot reach past the scratch' scripts/prove-reproducer.sh` and confirm one hit, and `grep -n 'ReadFile' stats/internal/guard/provereproducer.go` and confirm none — the defect: the header claims the scratch contains a hardcoding script, and nothing reads the body.
  - [x] **Step 2: Write the failing test.** Apply the test hunk alone (`git apply --include='stats/internal/guard/prove_reproducer_test.go' <the diff in step 3>`) and run `cd stats && go test ./internal/guard/ -run TestProveReproducerRefusesLiveWorktreePath -count=1`: it FAILS with `want exit 2 … got 1`.
  - [x] **Step 3: Implement the fix.** Apply the rest of this diff (`git apply --exclude='stats/internal/guard/prove_reproducer_test.go' <the diff>`):

````diff verified:git apply --check at 4f64d5f8; gofmt -l, go vet ./internal/guard/ and go test ./internal/guard/ -count=1 pass with it applied; TestProveReproducerRefusesLiveWorktreePath fails with provereproducer.go reverted (got 1)
diff --git a/scripts/prove-reproducer.sh b/scripts/prove-reproducer.sh
index 35db99ac..30827cf8 100755
--- a/scripts/prove-reproducer.sh
+++ b/scripts/prove-reproducer.sh
@@ -16,10 +16,12 @@
 # pre-fix commit in a throwaway directory under TMPDIR, the reproducer is
 # copied to the SAME worktree-relative path inside it, and the leg runs with
 # its working directory set to that scratch — so the cwd contract holds and
-# no absolute path is ever needed. F17's ROOT class is then structural: a
-# script that hardcodes its own tree cannot reach past the scratch, because
-# the scratch is the tree the leg runs in. The scratch is removed on every
-# path out of this script, failed proof or not.
+# no absolute path is ever needed. The scratch does not stop F17's ROOT
+# class: a script that hardcodes the live tree's absolute path still reads
+# the already-fixed tree from inside the scratch, so a reproducer whose body
+# names <worktree>'s absolute path is refused (exit 2) before any leg runs.
+# The scratch is removed on every path out of this script, failed proof or
+# not.
 #
 # EACH LEG IS A run-reproducer CALL, NOT A RE-IMPLEMENTATION -- in-process,
 # the Go run-reproducer (stats/internal/guard/runreproducer.go) called with
@@ -40,8 +42,9 @@
 #   1  the proof did not hold — a leg read the wrong demonstration; the
 #      verdict names which leg and what it read.
 #   2  cannot answer — usage, an unresolvable <pre-fix-ref>, an unusable
-#      <reproducer-path>, a git plumbing failure, or a leg run-reproducer
-#      refused (2), could not verdict (3) or could not answer (4). Never a
+#      <reproducer-path> or one whose body names <worktree>'s absolute
+#      path, a git plumbing failure, or a leg run-reproducer refused (2),
+#      could not verdict (3) or could not answer (4). Never a
 #      verdict on the defect.
 # flow-guard is built from this checkout, never taken from PATH:
 # scripts/lib/flow-guard.sh derives it, and exits 2 (this script's
diff --git a/stats/internal/guard/prove_reproducer_test.go b/stats/internal/guard/prove_reproducer_test.go
index f8c22395..f6225d35 100644
--- a/stats/internal/guard/prove_reproducer_test.go
+++ b/stats/internal/guard/prove_reproducer_test.go
@@ -332,6 +332,19 @@ func prBash(t *testing.T, env Env, args ...string) (int, string, string) {
 // alike, which is all that differs between two runs' messages.
 var prTempName = regexp.MustCompile(`prove-reproducer\.[A-Za-z0-9]+`)
 
+// TestProveReproducerRefusesLiveWorktreePath pins that a reproducer whose
+// body names the live worktree by absolute path is refused before any leg
+// runs: from inside the pre-fix scratch it would read the already-fixed
+// tree, and a gymie /flow run's pre-fix leg did exactly that (kan-988).
+func TestProveReproducerRefusesLiveWorktreePath(t *testing.T) {
+	t.Parallel()
+	repo, pre := prRepo(t, "repo")
+	prReproducer(t, repo, prRel, "exec grep -q old-behaviour "+repo+"/check.txt")
+	r := prRun(t, repo, Env{}, repo, pre, prRel)
+	prWant(t, r, 2, "prove-reproducer: the reproducer names the live worktree's absolute path "+repo+" -- the pre-fix leg would read the already-fixed tree; name paths relative to the working directory\n")
+	prUnchanged(t, r)
+}
+
 // TestProveReproducerBashParity pins the refusals after the scratch base
 // exists against the bash at d71a2327, stdout, stderr and exit: the
 // worktree add, mkdir -p and cp -p failures with the tools' own
diff --git a/stats/internal/guard/provereproducer.go b/stats/internal/guard/provereproducer.go
index e3ef4d09..27eb03b1 100644
--- a/stats/internal/guard/provereproducer.go
+++ b/stats/internal/guard/provereproducer.go
@@ -7,6 +7,7 @@ import (
 	"os"
 	"os/exec"
 	"os/signal"
+	"path/filepath"
 	"strings"
 	"syscall"
 	"time"
@@ -83,6 +84,21 @@ func proveReproducer(args []string, env Env, stdout, stderr io.Writer) int {
 	if syscall.Access(abs(wt+"/"+rel), rrExecOK) != nil {
 		return die("reproducer is not executable: %s/%s", wt, rel)
 	}
+	// A body naming the live worktree by its absolute path reads the
+	// already-fixed tree from inside the pre-fix scratch (F17's ROOT class),
+	// so the pre-fix leg would never run at the pre-fix commit. An
+	// unreadable body is left to the copy step's own refusal below.
+	if body, err := os.ReadFile(abs(wt + "/" + rel)); err == nil {
+		live := []string{filepath.Clean(abs(wt))}
+		if real, err := filepath.EvalSymlinks(live[0]); err == nil && real != live[0] {
+			live = append(live, real)
+		}
+		for _, p := range live {
+			if bytes.Contains(body, []byte(p)) {
+				return die("the reproducer names the live worktree's absolute path %s -- the pre-fix leg would read the already-fixed tree; name paths relative to the working directory", p)
+			}
+		}
+	}
 
 	// The bash's "${TMPDIR:-/tmp}/prove-reproducer.XXXXXX", slash for slash.
 	tmp := env.Getenv("TMPDIR")
````

  - [x] **Step 4: Verify.** `cd stats && gofmt -l ./internal/guard` prints nothing; `go vet ./internal/guard/` and `go test ./internal/guard/ -run 'TestProveReproducer' -count=1` pass; from the repository root `scripts/check-guard-symlinks.sh` exits 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/provereproducer.go`, `stats/internal/guard/prove_reproducer_test.go`, `scripts/prove-reproducer.sh`
**Tests:** `TestProveReproducerRefusesLiveWorktreePath`
**Regression:** reverting the commit lets a reproducer naming the live worktree's absolute path run both legs against the fixed tree, and the proof exits 1 instead of refusing.
**Baseline:** before=5 after=6
<!-- measured: grep -c '^func Test' stats/internal/guard/prove_reproducer_test.go @ 4f64d5f8 (before; after with this task's diff applied) -->
**Commit:** fix(guard): prove-reproducer refuses a reproducer naming the live worktree by absolute path
**After:** none

**Decision:** live-path-refusal


- [x] 2. The -effort refusal names the default rule
  - [x] **Step 1: Reproduce before touching anything; do not guess.** From `stats/`, run `go run ./cmd/flow record dispatch begin -agent-id none -C .. -change self-review-kan-988-ledger-test-pool-starvation -role implementer -model opus -effort unknown -key k1 -session-token mf-x` and confirm it is refused with exactly `flow: -effort "unknown" is not one of: low, medium, high, default` — no rule for an unstated effort.
  - [x] **Step 2: Write the failing test.** Apply the test hunk alone (`git apply --include='stats/cmd/flow/record_test.go' <the diff in step 3>`) and run `cd stats && go test ./cmd/flow/ -run TestRunRecordDispatchBeginUnknownEffortNamesDefault -count=1`: it FAILS on the stderr assertion.
  - [x] **Step 3: Implement the fix.** Apply the rest of this diff (`git apply --exclude='stats/cmd/flow/record_test.go' <the diff>`):

````diff verified:git apply --check at 4f64d5f8; gofmt -l, go vet ./cmd/flow/ and go test ./cmd/flow/ -count=1 pass with it applied; TestRunRecordDispatchBeginUnknownEffortNamesDefault fails with record.go reverted
diff --git a/stats/cmd/flow/record.go b/stats/cmd/flow/record.go
index e3af75a3..581803ac 100644
--- a/stats/cmd/flow/record.go
+++ b/stats/cmd/flow/record.go
@@ -1012,7 +1012,7 @@ func runRecordDispatchBegin(ctx context.Context, args []string, stdout, stderr i
 		return 2
 	}
 	if !slices.Contains(recordEfforts, *effort) {
-		fmt.Fprintf(stderr, "flow: -effort %q is not one of: %s\n", *effort, strings.Join(recordEfforts, ", "))
+		fmt.Fprintf(stderr, "flow: -effort %q is not one of: %s -- an effort the session does not state is recorded as default, never guessed\n", *effort, strings.Join(recordEfforts, ", "))
 		return 2
 	}
 	if err := validateSessionToken(*sessionToken); err != nil {
diff --git a/stats/cmd/flow/record_test.go b/stats/cmd/flow/record_test.go
index 49b09e56..7ef41f57 100644
--- a/stats/cmd/flow/record_test.go
+++ b/stats/cmd/flow/record_test.go
@@ -3295,6 +3295,29 @@ func TestRunRecordDispatchBeginRejectsUnknownEffort(t *testing.T) {
 	}
 }
 
+// TestRunRecordDispatchBeginUnknownEffortNamesDefault pins that the
+// refusal of an effort the session cannot state names the `default` rule
+// (model-policy.md, **Harness mapping**): kan-988's inline rows recorded
+// an invented `high` after `-effort unknown` was refused without it.
+func TestRunRecordDispatchBeginUnknownEffortNamesDefault(t *testing.T) {
+	repo := recordRepo(t, "kan-258")
+	isolatedStateRoot(t)
+
+	var stdout, stderr bytes.Buffer
+	code := run(context.Background(),
+		[]string{"record", "dispatch", "begin", "-agent-id", "none", "-addr", "http://127.0.0.1:1", "-timeout", "500ms", "-C", repo,
+			"-change", "kan-258", "-role", "implementer", "-model", "opus", "-effort", "unknown",
+			"-key", "k1", "-session-token", "mf-record-unknown-effort"},
+		strings.NewReader(""), &stdout, &stderr)
+
+	if code != 2 {
+		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
+	}
+	if want := "an effort the session does not state is recorded as default, never guessed"; !strings.Contains(stderr.String(), want) {
+		t.Errorf("stderr does not name the default rule %q:\n%s", want, stderr.String())
+	}
+}
+
 // TestRecordDispatchBeginDefaultsEffortToDefault pins that an omitted
 // -effort is sent as the literal "default" rather than left absent from
 // the wire body -- the store column is NOT NULL DEFAULT 'default' because
````

  - [x] **Step 4: Verify.** `cd stats && gofmt -l ./cmd/flow` prints nothing; `go vet ./cmd/flow/` and `go test ./cmd/flow/ -run 'TestRunRecordDispatchBegin.*Effort' -count=1` pass. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRunRecordDispatchBeginUnknownEffortNamesDefault`
**Regression:** reverting the commit drops the default rule from the refusal and the test's stderr assertion fails.
**Baseline:** before=107 after=108
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ 4f64d5f8 (before; after with this task's diff applied) -->
**Commit:** fix(flow-cli): the -effort refusal names the default rule for an unstated effort
**After:** none

**Decision:** effort-refusal-names-default


- [x] 3. The harvester counts each message's final output
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -n 'seen\[id\] = true' stats/internal/harvest/transcript.go` and confirm one hit — the first line per id is kept whatever its output count — and `grep -n 'lastMessageID map\[string\]string' stats/internal/harvest/watcher.go` and confirm one hit — the carry holds no count.
  - [x] **Step 2: Write the failing tests.** Apply the two test hunks alone (`git apply --include='stats/internal/harvest/*_test.go' <the diff in step 3>`) and run `cd stats && go test ./internal/harvest/ -run 'TestParseAssistantRecordsTakesLargestOutputPerMessage|TestMessageStraddlingTwoReadsCountsFinalOutput' -count=1`: both FAIL, output `4` where `135` is wanted.
  - [x] **Step 3: Implement the fix.** Apply the rest of this diff (`git apply --exclude='stats/internal/harvest/*_test.go' <the diff>`):

````diff verified:git apply --check at 4f64d5f8; gofmt -l, go vet ./... and go test ./... -count=1 pass with it applied; both declared tests fail with transcript.go and watcher.go reverted, and the straddle test fails with watcher.go alone reverted
diff --git a/stats/internal/harvest/transcript.go b/stats/internal/harvest/transcript.go
index 6c543a37..bdc053a7 100644
--- a/stats/internal/harvest/transcript.go
+++ b/stats/internal/harvest/transcript.go
@@ -278,8 +278,10 @@ func SplitCompleteLines(raw []byte) (complete, tail []byte) {
 // (as SplitCompleteLines defines "complete") into a Record, in file
 // order -- one Record per message.id, not per line: Claude Code writes
 // one JSONL line per content block of a single API response, every line
-// repeating that response's message.id and usage verbatim, so only the
-// first line for a given id is kept. Every other line -- a different
+// repeating that response's message.id and usage, so one Record is kept
+// per id at its first line's position -- carrying the largest output and
+// thinking counts any of its lines holds, since an early line can carry a
+// streaming-time output count the message's last line supersedes. Every other line -- a different
 // "type", or a line that fails to decode as JSON at all -- is skipped
 // rather than treated as fatal: unlike a journal entry (this package's
 // own format, task 6), a transcript line is written by the harness
@@ -292,7 +294,7 @@ func SplitCompleteLines(raw []byte) (complete, tail []byte) {
 // window happened to start at the Unix epoch.
 func ParseAssistantRecords(complete []byte) []Record {
 	var out []Record
-	seen := map[string]bool{}
+	seen := map[string]int{} // message.id -> its Record's index in out
 	start := 0
 	for start < len(complete) {
 		idx := bytes.IndexByte(complete[start:], '\n')
@@ -319,9 +321,12 @@ func ParseAssistantRecords(complete []byte) []Record {
 		}
 
 		// One API response is written as one line per content block, each
-		// line repeating the same message.id and the same usage (measured:
-		// 56 assistant lines for 15 ids on one real transcript). The first
-		// line per id is the whole message's usage; the rest are repeats.
+		// line repeating the same message.id and usage (measured: 56
+		// assistant lines for 15 ids on one real transcript). Input and
+		// cache counts repeat verbatim; output does not always -- an early
+		// line can carry the streaming-time count (measured: 260 of 10336
+		// ids across two days of transcripts, e.g. 4 then 135, the last
+		// line always the largest), so the kept Record takes the maximum.
 		// A line with no id is an older or foreign shape and is kept as is.
 		// Dedupe here is per call; a message whose lines straddle two reads
 		// (its lines land seconds apart, so most multi-line messages do --
@@ -330,18 +335,20 @@ func ParseAssistantRecords(complete []byte) []Record {
 		// ponytail: that carry is in memory, so a daemon restart can still
 		// count one message twice per transcript; persist it beside
 		// harvest_offsets if that ever shows in the data.
-		if id := raw.Message.ID; id != "" {
-			if seen[id] {
-				continue
-			}
-			seen[id] = true
-		}
-
 		u := raw.Message.Usage
 		var thinking int64
 		if u.OutputTokensDetails != nil {
 			thinking = u.OutputTokensDetails.ThinkingTokens
 		}
+		if id := raw.Message.ID; id != "" {
+			if i, ok := seen[id]; ok {
+				kept := &out[i].Usage
+				kept.OutputTokens = max(kept.OutputTokens, u.OutputTokens)
+				kept.ThinkingTokens = max(kept.ThinkingTokens, thinking)
+				continue
+			}
+			seen[id] = len(out)
+		}
 		var (
 			cache5m, cache1h int64
 			splitKnown       bool
diff --git a/stats/internal/harvest/transcript_test.go b/stats/internal/harvest/transcript_test.go
index 02257b91..9da6c23a 100644
--- a/stats/internal/harvest/transcript_test.go
+++ b/stats/internal/harvest/transcript_test.go
@@ -2,6 +2,7 @@ package harvest_test
 
 import (
 	"errors"
+	"fmt"
 	"os"
 	"path/filepath"
 	"testing"
@@ -345,6 +346,24 @@ func TestParseAssistantRecordsDeduplicatesByMessageID(t *testing.T) {
 	}
 }
 
+// TestParseAssistantRecordsTakesLargestOutputPerMessage pins that a
+// message whose early line carries a streaming-time output count keeps
+// the larger count a later line of the same id carries -- kan-988's
+// ledger rows stored 28 and 178 output tokens for reviewers that wrote
+// whole reports.
+func TestParseAssistantRecordsTakesLargestOutputPerMessage(t *testing.T) {
+	line := func(output, thinking int) string {
+		return fmt.Sprintf(`{"type":"assistant","timestamp":"2026-01-01T00:00:01Z","sessionId":"s","message":{"id":"msg_a","model":"m","usage":{"input_tokens":5,"output_tokens":%d,"output_tokens_details":{"thinking_tokens":%d}}}}`+"\n", output, thinking)
+	}
+	records := harvest.ParseAssistantRecords([]byte(line(4, 2) + line(4, 2) + line(135, 60)))
+	if len(records) != 1 {
+		t.Fatalf("got %d records, want 1 (three lines share msg_a)", len(records))
+	}
+	if u := records[0].Usage; u.OutputTokens != 135 || u.ThinkingTokens != 60 || u.InputTokens != 5 {
+		t.Errorf("usage = %+v, want output 135, thinking 60, input 5 -- the last line's final counts, input once", u)
+	}
+}
+
 // TestParseAssistantRecordsAgentIDAbsent is the negative companion: a
 // parent-session line carries no "agentId" at all, and must decode to
 // AgentID == "" -- never a placeholder, per this change's own
diff --git a/stats/internal/harvest/watcher.go b/stats/internal/harvest/watcher.go
index c8d41dad..9cab636c 100644
--- a/stats/internal/harvest/watcher.go
+++ b/stats/internal/harvest/watcher.go
@@ -420,13 +420,15 @@ type Watcher struct {
 	// for the same reasons tokenCycles is not.
 	seenOffsets map[string]int64
 
-	// lastMessageID is, per path, the message.id of the last usage record
-	// a committed batch carried. One API response is written as several
-	// lines seconds apart, so a read often ends mid-message; the next
-	// read's leading lines repeat that message's usage verbatim and are
-	// dropped here, or every straddling message counts twice. Not
+	// lastMessageID is, per path, the last usage record a committed batch
+	// carried -- its message.id and the output and thinking counts already
+	// counted for it. One API response is written as several lines seconds
+	// apart, so a read often ends mid-message; the next read's leading
+	// lines repeat that message's usage and are dropped here, or every
+	// straddling message counts twice -- save the output a later line adds
+	// over the count already counted, which is kept as its own record. Not
 	// persisted, for the same reasons seenOffsets is not.
-	lastMessageID map[string]string
+	lastMessageID map[string]carriedMessage
 
 	// pendingDispatchMeta remembers, per subagent transcript path whose
 	// most recently committed batch carried dispatch tokens but no
@@ -530,7 +532,7 @@ func NewWatcher(sources []Source, sink HarvestSink, attributor *Attributor, deps
 		gaveUpTokens:        make(map[string]bool),
 		retriedTokens:       make(map[string]bool),
 		seenOffsets:         make(map[string]int64),
-		lastMessageID:       make(map[string]string),
+		lastMessageID:       make(map[string]carriedMessage),
 		pendingDispatchMeta: make(map[string]map[int64]string),
 		dispatchMetaCycles:  make(map[string]int),
 		gaveUpDispatchMeta:  make(map[string]bool),
@@ -638,6 +640,7 @@ func (w *Watcher) RunOnce(ctx context.Context) (int, error) {
 			}
 			records, commands, launches, denials, newOffset :=
 				batch.Records, batch.Commands, batch.Launches, batch.Denials, batch.NewOffset
+			read := records
 			records = dropCarriedMessage(records, w.lastMessageID[path])
 
 			matchedHere := w.matchSessionTokens(pendingSessionTokens, commands, matchedSessions)
@@ -705,9 +708,7 @@ func (w *Watcher) RunOnce(ctx context.Context) (int, error) {
 			}
 			touchedFiles++
 			w.seenOffsets[path] = newOffset
-			if id := lastUsageMessageID(records); id != "" {
-				w.lastMessageID[path] = id
-			}
+			w.lastMessageID[path] = lastCarriedMessage(read, w.lastMessageID[path])
 
 			// The second, dispatch-grain attribution pass over the very same
 			// records, in the same batch -- deliberately here, after
@@ -1808,15 +1809,31 @@ func (w *Watcher) creditStampedAgentFile(ctx context.Context, parentPath, agentI
 	w.attributeAgentFile(ctx, ParseAssistantRecords(raw[:offset]), agentPath, agentID)
 }
 
+// carriedMessage is the message a path's previous committed read ended
+// on, with the output and thinking counts already counted for it.
+type carriedMessage struct {
+	id               string
+	output, thinking int64
+}
+
 // dropCarriedMessage removes the usage records repeating carried, the
-// message the previous committed read of the same path ended on.
-func dropCarriedMessage(records []Record, carried string) []Record {
-	if carried == "" {
+// message the previous committed read of the same path ended on -- save
+// the output a repeat counts over carried's, an early line's
+// streaming-time count superseded, which stays as an output-only record.
+func dropCarriedMessage(records []Record, carried carriedMessage) []Record {
+	if carried.id == "" {
 		return records
 	}
 	out := records[:0:0]
 	for _, r := range records {
-		if r.Signal == nil && r.MessageID == carried {
+		if r.Signal == nil && r.MessageID == carried.id {
+			if r.Usage.OutputTokens > carried.output {
+				r.Usage = Usage{
+					OutputTokens:   r.Usage.OutputTokens - carried.output,
+					ThinkingTokens: max(0, r.Usage.ThinkingTokens-carried.thinking),
+				}
+				out = append(out, r)
+			}
 			continue
 		}
 		out = append(out, r)
@@ -1824,15 +1841,22 @@ func dropCarriedMessage(records []Record, carried string) []Record {
 	return out
 }
 
-// lastUsageMessageID returns the message.id of records' last usage record
-// that carries one, or "" when none does.
-func lastUsageMessageID(records []Record) string {
+// lastCarriedMessage returns records' last usage record that carries a
+// message.id as the next read's carry, its counts raised to prev's when it
+// is prev's message, or prev when no record carries an id.
+func lastCarriedMessage(records []Record, prev carriedMessage) carriedMessage {
 	for i := len(records) - 1; i >= 0; i-- {
-		if records[i].Signal == nil && records[i].MessageID != "" {
-			return records[i].MessageID
+		r := records[i]
+		if r.Signal != nil || r.MessageID == "" {
+			continue
+		}
+		c := carriedMessage{r.MessageID, r.Usage.OutputTokens, r.Usage.ThinkingTokens}
+		if c.id == prev.id {
+			c.output, c.thinking = max(c.output, prev.output), max(c.thinking, prev.thinking)
 		}
+		return c
 	}
-	return ""
+	return prev
 }
 
 // survivors returns the begins no pass consumed -- neither paired with a
diff --git a/stats/internal/harvest/watcher_test.go b/stats/internal/harvest/watcher_test.go
index d419665d..61dd6f71 100644
--- a/stats/internal/harvest/watcher_test.go
+++ b/stats/internal/harvest/watcher_test.go
@@ -3776,6 +3776,45 @@ func TestMessageStraddlingTwoReadsCountsOnce(t *testing.T) {
 	}
 }
 
+// TestMessageStraddlingTwoReadsCountsFinalOutput: a read that ends on a
+// message's early line counts its streaming-time output; the next read's
+// later line of that message carries the final count, and only what it
+// adds over the count already counted is counted -- never dropped as a
+// repeat (kan-988), never counted twice.
+func TestMessageStraddlingTwoReadsCountsFinalOutput(t *testing.T) {
+	line := func(id, ts string, output int) string {
+		return fmt.Sprintf(`{"type":"assistant","timestamp":"%s","sessionId":"%s","message":{"id":"%s","model":"m","usage":{"input_tokens":100,"output_tokens":%d}}}`+"\n", ts, mainSessionID, id, output)
+	}
+	dir := t.TempDir()
+	path := filepath.Join(dir, "session.jsonl")
+	if err := os.WriteFile(path, []byte(line("msg_x", "2026-01-01T00:00:00Z", 4)), 0o644); err != nil {
+		t.Fatal(err)
+	}
+	windows := &fakeWindowSource{bySession: openWindowForMainSession(1)}
+	sink := newFakeHarvestSink()
+	w := harvest.NewWatcher([]harvest.Source{harvest.NewClaudeSource(dir)}, sink, harvest.NewAttributor(windows), harvest.NoDeps{}, nil)
+	runCycles(t, w, 1)
+
+	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
+	if err != nil {
+		t.Fatal(err)
+	}
+	if _, err := f.WriteString(line("msg_x", "2026-01-01T00:00:05Z", 135)); err != nil {
+		t.Fatal(err)
+	}
+	if err := f.Close(); err != nil {
+		t.Fatal(err)
+	}
+	runCycles(t, w, 1)
+
+	if got := sink.totals[1].Main.Output; got != 135 {
+		t.Fatalf("tokens.main.output = %v, want 135: msg_x's later line carries its final output count", got)
+	}
+	if got := sink.totals[1].Main.Input; got != 100 {
+		t.Fatalf("tokens.main.input = %v, want 100: msg_x's input is counted once", got)
+	}
+}
+
 // lateStampDeps is a store whose dispatch row for one agent gains its
 // agent id only when the watcher stamps it, the way a row begun after its
 // agent had already run (kan-933's task-1-reviewer-fix-1 and
````

  - [x] **Step 4: Verify.** `cd stats && gofmt -l ./internal/harvest` prints nothing; `go vet ./internal/harvest/` and `go test ./internal/harvest/ -run 'TestParseAssistantRecords|TestMessageStraddlingTwoReads' -count=1` pass. Commit.

**Build:** green
**Files:** `stats/internal/harvest/transcript.go`, `stats/internal/harvest/watcher.go`, `stats/internal/harvest/transcript_test.go`, `stats/internal/harvest/watcher_test.go`
**Tests:** `TestParseAssistantRecordsTakesLargestOutputPerMessage`, `TestMessageStraddlingTwoReadsCountsFinalOutput`
**Regression:** reverting the commit keeps each message's first-line output — 4 where its last line says 135 — within one read and across two.
**Baseline:** before=151 after=153
<!-- measured: cat stats/internal/harvest/*_test.go | grep -c '^func Test' @ 4f64d5f8 (before; after with this task's diff applied) -->
**Commit:** fix(harvest): count each message's final output across its lines and reads
**After:** none

**Decision:** harvest-final-output

