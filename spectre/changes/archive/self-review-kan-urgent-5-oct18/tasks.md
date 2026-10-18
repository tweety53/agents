# self-review-kan-urgent-5-oct18

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-18-self-review-kan-urgent-5-oct18-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; the commit body may name the task's Jira key. Never write `spectre/changes/`. Never merge or push.
- Each task names the repository it runs in, by its worktree; every `**Files:**` path is relative to that worktree's root. Task 1 runs in spectre's worktree `/Users/tweety53/Projects/spectre-worktrees/self-review-kan-urgent-5-oct18`, Task 4 in gymie's worktree `/Users/tweety53/Projects/gymie-worktrees/self-review-kan-urgent-5-oct18`, Tasks 2 and 3 in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-urgent-5-oct18`. Each worktree is on branch `self-review-kan-urgent-5-oct18`; never touch a main checkout.
- A task's diff is the unified diff its Step 2 gives, applied with `git apply` from its worktree's root, in task order — every diff was applied and checked there (`design.md`, **Context**). Save a diff by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean.
- KAN-965's merge-order half is disproved and carries no task; its link half is Task 1's (`design.md`, **Context**).

## Review focus

- A counterpart under `<repo>/.worktrees/<id>` must still be found first, and a peer with neither worktree must still read one-sided (Task 1).
- A launch whose prompt carries no fix-loop `Key:` line must pair with its begin exactly as before; a loop launch must never claim a begin, and must stamp only an empty `agent_id` (Task 2).
- A dispatch begun with no harness session in the environment must record no `session_id`, and `DispatchWindowsForSession` must still reach every row through its stage run's token (Task 2).
- The rule's `<!-- core -->` excerpt must be unchanged (Task 3).
- A backend `@Feature` code must still be checked in full, and a frontend code in a catalogued domain must still be an orphan when the catalog lacks it (Task 4).


- [x] 1. Find a peer's counterpart in the sibling worktree layout
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/spectre-worktrees/self-review-kan-urgent-5-oct18`, apply Step 2's diff to the two test files alone (`git apply --include='internal/check/link_test.go' --include='internal/cmd/link_test.go' <the saved diff>`) and run `go test ./internal/check/ ./internal/cmd/ -run 'SiblingWorktree' -count=1`: `TestLinkFindingsCounterpartOnlyInPeerSiblingWorktree` fails `one-sided link`, and `TestLinkCounterpartOnlyInPeerSiblingWorktree` fails `canonical change "kan-9-cross" does not exist in peer "can"` — the defect. Keep the applied hunks; Step 2 applies the rest with `git apply --exclude='internal/check/link_test.go' --exclude='internal/cmd/link_test.go' <the saved diff>`.
  - [x] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at spectre c3b62de; with it applied go test ./... passes, and with internal/tree/peer.go reverted both SiblingWorktree tests fail
diff --git a/internal/check/link_test.go b/internal/check/link_test.go
index 28ae7ea..caf36aa 100644
--- a/internal/check/link_test.go
+++ b/internal/check/link_test.go
@@ -305,6 +305,30 @@ func TestLinkFindingsCounterpartOnlyInPeerWorktree(t *testing.T) {
 	}
 }
 
+// TestLinkFindingsCounterpartOnlyInPeerSiblingWorktree pins the same
+// fallback for the sibling layout flow creates worktrees in,
+// <repo>-worktrees/<id>/spectre (KAN-975): without it validate reads a
+// link whose counterpart lives there as one-sided.
+func TestLinkFindingsCounterpartOnlyInPeerSiblingWorktree(t *testing.T) {
+	parent := t.TempDir()
+	satBase := testtree.Build(t, parent, "sat", nil, "can ../can\n")
+	canBase := testtree.Build(t, parent, "can", nil, "sat ../sat\n") // can's own changes/ stays empty
+	testtree.Change(t, satBase, "x", false, map[string]string{"link.md": satLinkMD("b", "1")})
+	testtree.Change(t, filepath.Join(canBase+"-worktrees", "x"), "x", false, map[string]string{
+		"link.md":     canLinkMD("b", goodMergeOrder),
+		"proposal.md": goodProposal,
+		"tasks.md":    oneTaskTasksFile,
+	})
+
+	sat, err := tree.Open(satBase)
+	if err != nil {
+		t.Fatal(err)
+	}
+	if got := findings(t, sat); len(got) != 0 {
+		t.Errorf("want a clean round trip against a peer sibling worktree, got:\n%s", findingMsgs(got))
+	}
+}
+
 // TestStructuralLinkPlusDesignKeepsDesignChecks pins that "nothing else"
 // in the satellite definition (design.md's pointer-tree-not-full-tree)
 // includes design.md: a change directory carrying link.md and a design.md
diff --git a/internal/cmd/link.go b/internal/cmd/link.go
index dde197c..97b2ac8 100644
--- a/internal/cmd/link.go
+++ b/internal/cmd/link.go
@@ -78,7 +78,7 @@ func Link(args []string, stdout, stderr io.Writer) int {
 		return Usage
 	}
 	// The counterpart may exist only in the peer repository's own worktree
-	// (<repo>/.worktrees/<id>/spectre): before either side of a cross-repo
+	// (tree.CounterpartWorktree's layouts): before either side of a cross-repo
 	// link has landed on a primary checkout, that worktree is the only
 	// place the canonical change exists — the same fallback
 	// check.resolveCounterpart applies on the validate side. When it
diff --git a/internal/cmd/link_test.go b/internal/cmd/link_test.go
index 379fd8d..8bd8640 100644
--- a/internal/cmd/link_test.go
+++ b/internal/cmd/link_test.go
@@ -145,6 +145,35 @@ func TestLinkCounterpartOnlyInPeerWorktree(t *testing.T) {
 	}
 }
 
+// TestLinkCounterpartOnlyInPeerSiblingWorktree pins the same fallback for
+// the sibling layout flow creates worktrees in,
+// <repo>-worktrees/<id>/spectre (KAN-975).
+func TestLinkCounterpartOnlyInPeerSiblingWorktree(t *testing.T) {
+	if _, err := exec.LookPath("git"); err != nil {
+		t.Skip("git not on PATH")
+	}
+	satBase, canBase := linkFixture(t, "kan-9-cross", "feature-cross", nil)
+	wt := filepath.Join(canBase+"-worktrees", "kan-9-cross")
+	wtChanges := filepath.Join(wt, "spectre", "changes", "kan-9-cross")
+	// A real git worktree carrying the change, then the change removed
+	// from the primary checkout, so the worktree is the only place it lives.
+	if out, err := exec.Command("git", "-C", canBase, "worktree", "add", "-q", "-b", "kan-9-cross", wt).CombinedOutput(); err != nil {
+		t.Fatalf("git worktree add: %v (%s)", err, out)
+	}
+	if err := os.RemoveAll(filepath.Join(canBase, "spectre", "changes", "kan-9-cross")); err != nil {
+		t.Fatal(err)
+	}
+	commitAll(t, canBase)
+
+	var out, errBuf bytes.Buffer
+	if code := Link([]string{"--root", satBase, "can:kan-9-cross"}, &out, &errBuf); code != OK {
+		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
+	}
+	if _, err := os.ReadFile(filepath.Join(wtChanges, "link.md")); err != nil {
+		t.Fatalf("canonical link.md not written into the peer sibling worktree: %v", err)
+	}
+}
+
 func TestLink(t *testing.T) {
 	satBase, canBase := linkFixture(t, "kan-9-cross", "feature-cross", nil)
 
diff --git a/internal/tree/peer.go b/internal/tree/peer.go
index 7a47065..827b0b7 100644
--- a/internal/tree/peer.go
+++ b/internal/tree/peer.go
@@ -102,20 +102,25 @@ func NamesFor(declared map[string]string, root string) map[string]bool {
 }
 
 // CounterpartWorktree returns p's repository's own worktree tree named
-// after id — <repo>/.worktrees/<id>/spectre, repoParent's convention —
-// reporting whether it exists and opens. Before either side of a
+// after id — <repo>/.worktrees/<id>/spectre, or the sibling layout
+// <repo>-worktrees/<id>/spectre that flow creates (KAN-975), in that
+// order — reporting whether one exists and opens. Before either side of a
 // cross-repo link has landed on a primary checkout, that worktree is the
 // only place the counterpart change exists at all.
 func CounterpartWorktree(p *Tree, id string) (*Tree, bool) {
-	wtSpectre := filepath.Join(filepath.Dir(p.Root), ".worktrees", id, "spectre")
-	if fi, err := os.Stat(wtSpectre); err != nil || !fi.IsDir() {
-		return nil, false
-	}
-	wt, err := Open(wtSpectre)
-	if err != nil {
-		return nil, false
+	repo := filepath.Dir(p.Root)
+	for _, wtSpectre := range []string{
+		filepath.Join(repo, ".worktrees", id, "spectre"),
+		filepath.Join(repo+"-worktrees", id, "spectre"),
+	} {
+		if fi, err := os.Stat(wtSpectre); err != nil || !fi.IsDir() {
+			continue
+		}
+		if wt, err := Open(wtSpectre); err == nil {
+			return wt, true
+		}
 	}
-	return wt, true
+	return nil, false
 }
 
 func sameDir(a, b string) bool {
````

  - [x] **Step 3: Verify.** In `/Users/tweety53/Projects/spectre-worktrees/self-review-kan-urgent-5-oct18`: `gofmt -l .` prints nothing, `go vet ./...` exits 0, and `go test ./internal/check/ ./internal/cmd/ -run 'CounterpartOnlyInPeer' -count=1` passes. Commit.

**Build:** green
**Files:** `internal/tree/peer.go`, `internal/cmd/link.go`, `internal/check/link_test.go`, `internal/cmd/link_test.go`
**Tests:** `TestLinkFindingsCounterpartOnlyInPeerSiblingWorktree`, `TestLinkCounterpartOnlyInPeerSiblingWorktree`
**Regression:** reverting `CounterpartWorktree`'s sibling candidate leaves a `<repo>-worktrees/<id>` counterpart unfound, so validate reports `one-sided link` and link refuses `does not exist in peer`, failing both tests.
**Baseline:** before=15 after=17
<!-- measured: cat internal/check/link_test.go internal/cmd/link_test.go | grep -c '^func Test' @ spectre c3b62de (before; after with this task's diff applied) -->
**Commit:** fix(tree): find a peer's counterpart in the sibling worktree layout
**After:** none

**Decision:** counterpart-sibling-layout

- [x] 2. Measure fix-loop dispatches and inline rows of a run that marks no stage
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/agents-worktrees/self-review-kan-urgent-5-oct18`, apply Step 2's diff to the three test files alone (`git apply --include='stats/internal/harvest/watcher_test.go' --include='stats/internal/store/records_test.go' --include='stats/cmd/flow/record_test.go' <the saved diff>`) and run `cd stats && go test ./internal/harvest/ ./cmd/flow/ -run 'TestWatcherStampsLoopLaunchByItsKey|TestRecordDispatchBeginSendsTheSessionID' -count=1`: both fail — no loop stamp, no `sessionId` sent — and `go vet ./internal/store/` fails `mine.SessionID undefined`, the store half the tree lacks. Keep the applied hunks; Step 2 applies the rest with `git apply --exclude='stats/internal/harvest/watcher_test.go' --exclude='stats/internal/store/records_test.go' --exclude='stats/cmd/flow/record_test.go' <the saved diff>`. The defect itself is the dev store's (`design.md`, **Context**): rows under fix-loop keys carry no `agentId` and empty `metrics`.
  - [x] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at agents b61a1e6a; with it applied go test ./internal/harvest/ ./internal/store/ ./internal/records/ ./internal/api/ ./cmd/flow/ passes, and each new test fails with its own fix line reverted (the stampLoopLaunches call, the session_id match, the begin's SessionID)
diff --git a/stats/cmd/flow/record.go b/stats/cmd/flow/record.go
index e1f15d8d..6cfcb30e 100644
--- a/stats/cmd/flow/record.go
+++ b/stats/cmd/flow/record.go
@@ -1060,6 +1060,10 @@ func runRecordDispatchBegin(ctx context.Context, args []string, stdout, stderr i
 	// original attempt was made rather than dating the dispatch to the
 	// replay.
 	attemptAt := time.Now()
+	sessionID := ""
+	if v := sessionIDFromEnv(); v != nil {
+		sessionID = *v
+	}
 	in := records.Dispatch{
 		AgentID:      *agentID,
 		Key:          *key,
@@ -1070,6 +1074,7 @@ func runRecordDispatchBegin(ctx context.Context, args []string, stdout, stderr i
 		Effort:       *effort,
 		DiffBase:     *diffBase,
 		SessionToken: *sessionToken,
+		SessionID:    sessionID,
 		StartedAt:    started,
 	}
 	out, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.Dispatch, error) {
diff --git a/stats/cmd/flow/record_test.go b/stats/cmd/flow/record_test.go
index ffbba3fe..8987c4f6 100644
--- a/stats/cmd/flow/record_test.go
+++ b/stats/cmd/flow/record_test.go
@@ -800,6 +800,19 @@ func TestRecordDispatchBeginAcceptsDiffBase(t *testing.T) {
 	}
 }
 
+// TestRecordDispatchBeginSendsTheSessionID pins KAN-967's CLI half: a
+// begin carries the harness session it ran in, so the store can bind a row
+// no stage run's token reaches.
+func TestRecordDispatchBeginSendsTheSessionID(t *testing.T) {
+	t.Setenv("CLAUDE_CODE_SESSION_ID", "session-kan967")
+
+	sent := dispatchBeginBody(t)
+
+	if sent["sessionId"] != "session-kan967" {
+		t.Errorf("sessionId = %v, want session-kan967", sent["sessionId"])
+	}
+}
+
 // TestRecordDispatchBeginDiffBaseIsOptional pins the flag's absence, which
 // is an ordinary state rather than a degraded one: every implementer
 // dispatch, and every panel slot reading the whole diff, legitimately
diff --git a/stats/internal/harvest/deps.go b/stats/internal/harvest/deps.go
index a6bba2a9..7b8feb48 100644
--- a/stats/internal/harvest/deps.go
+++ b/stats/internal/harvest/deps.go
@@ -19,6 +19,11 @@ import (
 // store.Store.StampDispatchAgent exists.
 type DispatchAgentStamper interface {
 	StampDispatchAgent(ctx context.Context, sessionToken, key, agentID string) (stamped bool, err error)
+	// StampLoopDispatchAgent fills an empty agent_id on the latest row
+	// recorded under a fix-loop key alone (AgentLaunch.LoopKey): a loop
+	// row's begin is never a transcript command, so no session token
+	// reaches the watcher with it.
+	StampLoopDispatchAgent(ctx context.Context, key, agentID string) (stamped bool, err error)
 }
 
 // Deps is everything a Watcher needs beyond its root, its sink and its
@@ -114,3 +119,8 @@ func (NoDeps) DispatchWindowsForAgent(ctx context.Context, agentID string) ([]Di
 func (NoDeps) StampDispatchAgent(ctx context.Context, sessionToken, key, agentID string) (stamped bool, err error) {
 	return false, nil
 }
+
+// StampLoopDispatchAgent stamps nothing and reports false.
+func (NoDeps) StampLoopDispatchAgent(ctx context.Context, key, agentID string) (stamped bool, err error) {
+	return false, nil
+}
diff --git a/stats/internal/harvest/transcript.go b/stats/internal/harvest/transcript.go
index bdc053a7..77833ead 100644
--- a/stats/internal/harvest/transcript.go
+++ b/stats/internal/harvest/transcript.go
@@ -27,6 +27,7 @@ import (
 	"io"
 	"os"
 	"path/filepath"
+	"regexp"
 	"slices"
 	"sort"
 	"strings"
@@ -158,6 +159,7 @@ type rawLine struct {
 type rawToolUseResult struct {
 	AgentID string `json:"agentId"`
 	Status  string `json:"status"`
+	Prompt  string `json:"prompt"`
 }
 
 // rawCompactMeta is a compact_boundary system line's "compactMetadata"
@@ -429,6 +431,26 @@ type AgentLaunch struct {
 	AgentID   string
 	Timestamp time.Time
 	Line      int
+	// LoopKey is the fix-loop dispatch key the launched prompt's own
+	// `Key:` line names (loopKeyPattern), empty for any other launch. A
+	// loop dispatch's begin is recorded by the PreToolUse hook
+	// (hooks/enforce-loop-dispatch.py), never by a Bash command this
+	// transcript carries, so no begin event exists to pair with: the key
+	// names the row instead (KAN-967).
+	LoopKey string
+}
+
+// loopKeyPattern reads a fix-loop dispatch key off a launched prompt --
+// the `Key: pipeline-fix-<k>-...` or `Key: self-review-...` line
+// hooks/enforce-loop-dispatch.py admits and records the row under.
+var loopKeyPattern = regexp.MustCompile(`(?m)^[ \t]*Key:[ \t]*((?:pipeline-fix|self-review)-\S+)[ \t]*$`)
+
+// loopKey returns the fix-loop key prompt names, or "".
+func loopKey(prompt string) string {
+	if m := loopKeyPattern.FindStringSubmatch(prompt); m != nil {
+		return m[1]
+	}
+	return ""
 }
 
 // launchStatus is the toolUseResult status that says a dispatch LAUNCHED
@@ -531,6 +553,7 @@ func ParseAgentLaunches(complete []byte) []AgentLaunch {
 					AgentID:   raw.ToolUseResult.AgentID,
 					Timestamp: ts,
 					Line:      lineIndex,
+					LoopKey:   loopKey(raw.ToolUseResult.Prompt),
 				})
 			}
 		}
diff --git a/stats/internal/harvest/watcher.go b/stats/internal/harvest/watcher.go
index 9cab636c..88b6618d 100644
--- a/stats/internal/harvest/watcher.go
+++ b/stats/internal/harvest/watcher.go
@@ -1657,6 +1657,7 @@ func (w *Watcher) stampDispatchAgents(ctx context.Context, path string, commands
 			}
 		}
 	}
+	launches = w.stampLoopLaunches(ctx, path, launches)
 	begins = append(w.pendingBegins[path], begins...)
 	launches = append(w.pendingLaunches[path], launches...)
 	denials = append(w.pendingDenials[path], denials...)
@@ -1775,6 +1776,31 @@ func (w *Watcher) stampDispatchAgents(ctx context.Context, path string, commands
 	w.pendingDenials[path] = appendPending(w.pendingDenials[path], denials)
 }
 
+// stampLoopLaunches stamps each fix-loop launch (AgentLaunch.LoopKey) onto
+// the row its key names and returns the other launches for the begin
+// pairing. A loop launch never pairs with a begin: its begin was recorded
+// by the PreToolUse hook, so any begin in reach belongs to another
+// dispatch (KAN-967). The failure posture is stampDispatchAgents' own.
+func (w *Watcher) stampLoopLaunches(ctx context.Context, path string, launches []AgentLaunch) []AgentLaunch {
+	rest := launches[:0:0]
+	for _, launch := range launches {
+		if launch.LoopKey == "" {
+			rest = append(rest, launch)
+			continue
+		}
+		stamped, err := w.deps.StampLoopDispatchAgent(ctx, launch.LoopKey, launch.AgentID)
+		if err != nil {
+			w.warn("harvest: stamp loop dispatch agent failed, the row keeps no id from this launch",
+				"path", path, "key", launch.LoopKey, "agent_id", launch.AgentID, "error", err)
+			continue
+		}
+		if stamped {
+			w.creditStampedAgentFile(ctx, path, launch.AgentID)
+		}
+	}
+	return rest
+}
+
 // creditStampedAgentFile credits a just-stamped dispatch row with the part
 // of its agent's transcript already harvested. Every batch of that file
 // committed before the stamp met no window for the agent's id -- none
diff --git a/stats/internal/harvest/watcher_test.go b/stats/internal/harvest/watcher_test.go
index 47919890..1f4d75c7 100644
--- a/stats/internal/harvest/watcher_test.go
+++ b/stats/internal/harvest/watcher_test.go
@@ -10,6 +10,7 @@ import (
 	"log/slog"
 	"os"
 	"path/filepath"
+	"reflect"
 	"strings"
 	"testing"
 	"time"
@@ -3248,7 +3249,13 @@ func TestGrownTranscriptIsReadAfterSkip(t *testing.T) {
 type stampRecordingDeps struct {
 	harvest.NoDeps
 
-	stamps []stampCall
+	stamps     []stampCall
+	loopStamps []stampCall
+}
+
+func (d *stampRecordingDeps) StampLoopDispatchAgent(ctx context.Context, key, agentID string) (bool, error) {
+	d.loopStamps = append(d.loopStamps, stampCall{"", key, agentID})
+	return true, nil
 }
 
 type stampCall struct {
@@ -3305,6 +3312,34 @@ func TestWatcherStampsLaunchAgainstPrecedingBegin(t *testing.T) {
 	}
 }
 
+// TestWatcherStampsLoopLaunchByItsKey pins KAN-967: a fix-loop launch,
+// whose begin the PreToolUse hook recorded and no transcript command
+// carries, stamps the row its prompt's Key line names -- and never pairs
+// with a begin that happens to sit in reach, which belongs to another
+// dispatch.
+func TestWatcherStampsLoopLaunchByItsKey(t *testing.T) {
+	dir := t.TempDir()
+	path := filepath.Join(dir, "session.jsonl")
+	loopLaunch := `{"type":"user","timestamp":"2026-01-01T00:00:01Z","sessionId":"s","toolUseResult":{"isAsync":true,"status":"async_launched","agentId":"aloop1","prompt":"Before anything else, read the baseline.\n\nKey: self-review-kan-1-plan\n\nPlan it."}}`
+	if err := os.WriteFile(path, stampFixtureLines(stampBeginLine, loopLaunch), 0o644); err != nil {
+		t.Fatalf("write fixture: %v", err)
+	}
+
+	deps := &stampRecordingDeps{}
+	w := harvest.NewWatcher([]harvest.Source{harvest.NewClaudeSource(dir)}, newFakeHarvestSink(), harvest.NewAttributor(&fakeWindowSource{}), deps, nil)
+	if _, err := w.RunOnce(context.Background()); err != nil {
+		t.Fatalf("RunOnce: %v", err)
+	}
+
+	want := []stampCall{{"", "self-review-kan-1-plan", "aloop1"}}
+	if !reflect.DeepEqual(deps.loopStamps, want) {
+		t.Errorf("loop stamps = %+v, want %+v", deps.loopStamps, want)
+	}
+	if len(deps.stamps) != 0 {
+		t.Errorf("the loop launch paired with a begin: %+v", deps.stamps)
+	}
+}
+
 // TestWatcherCarriesBeginAcrossBatchBoundary pins the cross-batch case:
 // the begin commits in one batch and the launch lands in bytes the next
 // RunOnce reads, so the pairing must survive between cycles -- one begin
diff --git a/stats/internal/records/types.go b/stats/internal/records/types.go
index 983baabe..08f1364e 100644
--- a/stats/internal/records/types.go
+++ b/stats/internal/records/types.go
@@ -86,6 +86,12 @@ import (
 // handshaken back from the dispatched role -- a model cannot report its
 // own effort -- so it carries only what the dispatcher itself set:
 // `low`, `medium`, `high`, or `default` where none was set.
+//
+// SessionID is the harness session the begin ran in, from
+// CLAUDE_CODE_SESSION_ID or MUSE_SESSION_ID. It binds the row to its
+// session where no stage run carries SessionToken -- a fix run that marks
+// no stage -- so an inline row still takes the parent's main-thread usage
+// (KAN-967). It is written at begin and never read back.
 type Dispatch struct {
 	ID           int64           `json:"id"`
 	AgentID      string          `json:"agentId,omitempty"`
@@ -102,6 +108,7 @@ type Dispatch struct {
 	Outcome      string          `json:"outcome,omitempty"`
 	Cause        string          `json:"cause,omitempty"`
 	SessionToken string          `json:"sessionToken,omitempty"`
+	SessionID    string          `json:"sessionId,omitempty"`
 	StartedAt    time.Time       `json:"startedAt"`
 	EndedAt      *time.Time      `json:"endedAt,omitempty"`
 	Metrics      json.RawMessage `json:"metrics,omitempty"`
diff --git a/stats/internal/store/migrations/0040_dispatch_session_id.sql b/stats/internal/store/migrations/0040_dispatch_session_id.sql
new file mode 100644
index 00000000..de552bbf
--- /dev/null
+++ b/stats/internal/store/migrations/0040_dispatch_session_id.sql
@@ -0,0 +1,4 @@
+-- 0040_dispatch_session_id.sql: the harness session a dispatch's begin ran in (KAN-967). A
+-- dispatch row joins its session through stage_runs.session_token; a fix run that marks no stage
+-- has no such row, so DispatchWindowsForSession also matches this column.
+ALTER TABLE dispatches ADD COLUMN IF NOT EXISTS session_id text;
diff --git a/stats/internal/store/records.go b/stats/internal/store/records.go
index 66e4a766..d0b2c5c5 100644
--- a/stats/internal/store/records.go
+++ b/stats/internal/store/records.go
@@ -423,12 +423,12 @@ func (s *Store) insertDispatch(ctx context.Context, projectKey, change string, i
 		INSERT INTO dispatches (
 			change_id, stage_run_id, seq, task_id, role, slot, model, commit_sha, outcome,
 			session_token, started_at, ended_at, metrics, notes, agent_id, dispatch_key,
-			diff_base, effort
+			diff_base, effort, session_id
 		)
 		SELECT
 			c.id, $3,
 			COALESCE((SELECT MAX(d.seq) FROM dispatches d WHERE d.change_id = c.id), 0) + 1,
-			$4, $5, canonical_slot($6), $7, $8, $9, $10, $11, $12, $13::jsonb, $14, $15, $16, $17, $18
+			$4, $5, canonical_slot($6), $7, $8, $9, $10, $11, $12, $13::jsonb, $14, $15, $16, $17, $18, $19
 		FROM changes c
 		WHERE c.project_key = $1 AND c.name = $2
 		ON CONFLICT ON CONSTRAINT `+dispatchesKeyConstraint+` DO UPDATE SET
@@ -442,7 +442,7 @@ func (s *Store) insertDispatch(ctx context.Context, projectKey, change string, i
 		projectKey, change, in.StageRunID, nullIfEmpty(in.TaskID), in.Role, nullIfEmpty(in.Slot),
 		in.Model, nullIfEmpty(in.CommitSHA), nullIfEmpty(in.Outcome), nullIfEmpty(in.SessionToken),
 		in.StartedAt, in.EndedAt, metrics, nullIfEmpty(in.Notes), nullIfEmpty(in.AgentID),
-		nullIfEmpty(in.Key), nullIfEmpty(in.DiffBase), effort,
+		nullIfEmpty(in.Key), nullIfEmpty(in.DiffBase), effort, nullIfEmpty(in.SessionID),
 	))
 }
 
@@ -1257,7 +1257,7 @@ func (s *Store) DispatchWindowsForSession(ctx context.Context, sessionID string)
 		SELECT d.id, d.started_at, COALESCE(d.ended_at, sr.ended_at), d.agent_id
 		FROM dispatches d
 		LEFT JOIN stage_runs sr ON sr.id = d.stage_run_id
-		WHERE d.session_token IN (
+		WHERE d.session_id = $1 OR d.session_token IN (
 			SELECT sr.session_token FROM stage_runs sr
 			WHERE sr.session_id = $1 AND sr.session_token IS NOT NULL
 		)
@@ -1373,6 +1373,29 @@ func (s *Store) StampDispatchAgent(ctx context.Context, sessionToken, key, agent
 	return tag.RowsAffected() > 0, nil
 }
 
+// StampLoopDispatchAgent is StampDispatchAgent for a fix-loop row, named
+// by its key alone (harvest.AgentLaunch.LoopKey): the loop's begin is
+// recorded by the PreToolUse hook, so the transcript carries no session
+// token for it. It fills only an empty agent_id, on the latest row
+// recorded under key -- check-loop-dispatch refuses a key that already
+// ran, so an older row sharing it is a different loop's.
+func (s *Store) StampLoopDispatchAgent(ctx context.Context, key, agentID string) (bool, error) {
+	tag, err := s.pool.Exec(ctx, `
+		UPDATE dispatches
+		SET agent_id = $2
+		WHERE id = (
+			SELECT d.id FROM dispatches d
+			WHERE d.dispatch_key = $1
+			ORDER BY d.started_at DESC, d.id DESC
+			LIMIT 1
+		) AND COALESCE(agent_id, '') = ''
+	`, key, agentID)
+	if err != nil {
+		return false, fmt.Errorf("store: stamp loop dispatch agent for key %q: %w", key, err)
+	}
+	return tag.RowsAffected() > 0, nil
+}
+
 // RunRecord returns one change's whole derived record: its dispatches in
 // seq order and its findings in the order the numbers their refs spell,
 // which is the order every rendering of the record reads them in.
diff --git a/stats/internal/store/records_test.go b/stats/internal/store/records_test.go
index cdf5c475..4b601b4c 100644
--- a/stats/internal/store/records_test.go
+++ b/stats/internal/store/records_test.go
@@ -1071,6 +1071,38 @@ func TestDispatchWindowsForSessionResolvesTheBoundToken(t *testing.T) {
 	}
 }
 
+// TestDispatchWindowsForSessionMatchesTheRowsOwnSession pins KAN-967: a
+// dispatch whose begin recorded its session id has a window in that
+// session though no stage run carries its token -- a fix run that marks no
+// stage still has its inline rows' windows.
+func TestDispatchWindowsForSessionMatchesTheRowsOwnSession(t *testing.T) {
+	st, _ := newRecordStore(t)
+	ctx := context.Background()
+	projectKey := fmt.Sprintf("proj-dispatch-own-session-%d", time.Now().UnixNano())
+	seedChange(t, st, projectKey, "kan-1")
+
+	const sessionID = "session-dispatch-own"
+	mine := baseDispatch("implementer", "opus")
+	mine.SessionToken = "mf-kan967-no-stage"
+	mine.SessionID = sessionID
+	mine.AgentID = "inline"
+	recorded, err := st.RecordDispatch(ctx, projectKey, "kan-1", mine)
+	if err != nil {
+		t.Fatalf("RecordDispatch: %v", err)
+	}
+
+	windows, err := st.DispatchWindowsForSession(ctx, sessionID)
+	if err != nil {
+		t.Fatalf("DispatchWindowsForSession: %v", err)
+	}
+	if len(windows) != 1 || windows[0].DispatchID != recorded.ID || windows[0].AgentID != "inline" {
+		t.Fatalf("windows = %+v, want the one row begun in this session", windows)
+	}
+	if other, err := st.DispatchWindowsForSession(ctx, "session-someone-else"); err != nil || len(other) != 0 {
+		t.Errorf("another session's windows = %+v (%v), want none", other, err)
+	}
+}
+
 // TestDispatchWindowsForSessionCapsAnOpenDispatchAtItsStageRunEnd pins the
 // cap: a dispatch row whose end never came is closed, for attribution, at
 // its owning stage run's end, so it cannot take every later record of the
@@ -3201,6 +3233,50 @@ func TestStampDispatchAgentFillsEmptyID(t *testing.T) {
 	}
 }
 
+// TestStampLoopDispatchAgentFillsLatestRowByKey pins KAN-967's store
+// half: a fix-loop row is named by its key alone, and only the latest
+// row recorded under it is filled -- an older row sharing the key is
+// another loop's and keeps its empty id.
+func TestStampLoopDispatchAgentFillsLatestRowByKey(t *testing.T) {
+	st, pool := newRecordStore(t)
+	ctx := context.Background()
+
+	projectKey := fmt.Sprintf("proj-stamp-loop-%d", time.Now().UnixNano())
+	seedChange(t, st, projectKey, "kan-1")
+	const key = "self-review-kan-1-plan"
+	seedStampedDispatch(t, st, ctx, projectKey, "self-review-kan-1-old", key)
+	seedStampedDispatch(t, st, ctx, projectKey, "self-review-kan-1", key)
+
+	stamped, err := st.StampLoopDispatchAgent(ctx, key, "aloop1")
+	if err != nil {
+		t.Fatalf("StampLoopDispatchAgent: %v", err)
+	}
+	if !stamped {
+		t.Fatalf("StampLoopDispatchAgent reported false, want the latest row filled")
+	}
+	for token, want := range map[string]string{"self-review-kan-1": "aloop1", "self-review-kan-1-old": ""} {
+		var agentID *string
+		if err := pool.QueryRow(ctx,
+			"SELECT agent_id FROM dispatches d WHERE d.session_token = $1 AND d.dispatch_key = $2",
+			token, key,
+		).Scan(&agentID); err != nil {
+			t.Fatalf("read agent_id back: %v", err)
+		}
+		got := ""
+		if agentID != nil {
+			got = *agentID
+		}
+		if got != want {
+			t.Errorf("%s: agent_id = %q, want %q", token, got, want)
+		}
+	}
+
+	stamped, err = st.StampLoopDispatchAgent(ctx, key, "aloop2")
+	if err != nil || stamped {
+		t.Errorf("a second stamp = (%v, %v), want (false, nil): the filled row keeps its id", stamped, err)
+	}
+}
+
 // TestStampDispatchAgentNeverOverwrites pins the recorded-intent rule: a
 // begin that carried an explicit -agent-id (or an end that supplied one)
 // stated what the harness reported, by hand, and the automatic stamp must
````

  - [x] **Step 3: Verify.** `cd stats && gofmt -w . && make -s web-dist && go vet ./... && gofmt -l .` exits clean, and `go test ./internal/harvest/ ./internal/store/ ./cmd/flow/ -run 'TestWatcherStampsLoopLaunchByItsKey|TestStampLoopDispatchAgentFillsLatestRowByKey|TestDispatchWindowsForSessionMatchesTheRowsOwnSession|TestRecordDispatchBeginSendsTheSessionID' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/internal/harvest/transcript.go`, `stats/internal/harvest/deps.go`, `stats/internal/harvest/watcher.go`, `stats/internal/harvest/watcher_test.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `stats/internal/store/migrations/0040_dispatch_session_id.sql`, `stats/internal/records/types.go`, `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestWatcherStampsLoopLaunchByItsKey`, `TestStampLoopDispatchAgentFillsLatestRowByKey`, `TestDispatchWindowsForSessionMatchesTheRowsOwnSession`, `TestRecordDispatchBeginSendsTheSessionID`
**Regression:** reverting the watcher's `stampLoopLaunches` call leaves a loop launch unstamped; reverting the `session_id` match leaves a stage-less inline row without a window; reverting the begin's `SessionID` sends no session — each fails its own test.
**Baseline:** before=255 after=259
<!-- measured: cat stats/internal/harvest/watcher_test.go stats/internal/store/records_test.go stats/cmd/flow/record_test.go | grep -c '^func Test' @ agents b61a1e6a (before; after with this task's diff applied) -->
**Commit:** fix(stats): measure fix-loop dispatches and inline rows of a run that marks no stage
**After:** none

**Decision:** loop-launch-stamp-by-key
**Decision:** dispatch-session-id

- [x] 3. Throttle a rare flake until it fails, and review a flake fix for its measured proof
  - [x] **Step 1: Confirm the gap before touching anything; do not guess.** In `/Users/tweety53/Projects/agents-worktrees/self-review-kan-urgent-5-oct18`, `grep -n -i 'throttl' rules/fix-determinism-at-the-source.mdc skills/flow/primary-reviewer-prompt.md` prints nothing, and `grep -n -i 'flak' skills/flow/primary-reviewer-prompt.md` prints nothing.
  - [x] **Step 2: Apply the edit.** Apply this diff:

````diff verified:git apply at agents b61a1e6a; with it applied scripts/check-references.sh, scripts/check-markdown-integrity.py and scripts/check-dispatch-paragraphs.sh exit 0
diff --git a/rules/fix-determinism-at-the-source.mdc b/rules/fix-determinism-at-the-source.mdc
index 8c53a58e..5e37c22d 100644
--- a/rules/fix-determinism-at-the-source.mdc
+++ b/rules/fix-determinism-at-the-source.mdc
@@ -28,6 +28,12 @@ Run the test until you have seen it fail, and name what actually varied — time
 leftover state, a live network, machine load. A flake you have not reproduced is a guess, and a
 guess routes straight to the tolerance.
 
+When it will not fail at its natural rate, make it fail: throttle what the nondeterminism runs on —
+the CPU (a parallel load, a CPU cap), the network (a throttled connection), or a delay injected at
+the suspected seam — until it does. Then measure the rate under that condition, failures over runs.
+That rate, not one observed failure, sizes the proof runs, and the same condition is the one the
+fix and the mechanism-removed run below are proven under.
+
 ## Fix at the source
 
 - **Pin identity.** When the assertion matches on something that varies run to run — a generated
diff --git a/skills/flow/primary-reviewer-prompt.md b/skills/flow/primary-reviewer-prompt.md
index 8087aec0..32a33a4f 100644
--- a/skills/flow/primary-reviewer-prompt.md
+++ b/skills/flow/primary-reviewer-prompt.md
@@ -49,6 +49,12 @@
     generic review may not: tests exercise real behavior rather than mocks, and a changed
     schema or public interface carries its migration and backward-compatibility story.
 
+    **A flake fix carries its measured proof.** When the diff fixes a failure that does not
+    recur on every run, its commits or the change's records state the failure rate measured
+    before the fix, the run count, and a run failing with the fix's mechanism removed
+    (`~/.claude/rules/fix-determinism-at-the-source.md`). Report a fix that lacks any of the
+    three, and one that widens a tolerance, timeout, retry count or sleep in their place.
+
     **Shared state the diff reuses — a required step, done before anything else:**
     1. List every pre-existing field, flag or variable that the diff's new code reads or
        writes.
````

  - [x] **Step 3: Verify.** `scripts/check-references.sh`, `scripts/check-markdown-integrity.py` and `scripts/check-dispatch-paragraphs.sh` exit 0, and `git diff HEAD -- rules/fix-determinism-at-the-source.mdc` touches no line between `<!-- core -->` and `<!-- /core -->`. Commit.

**Build:** green
**Files:** `rules/fix-determinism-at-the-source.mdc`, `skills/flow/primary-reviewer-prompt.md`
**Tests:** **none** — prose only; the doc guards in Step 3 cover it.
**Regression:** reverting the commit drops the throttle-and-measure step and the primary reviewer's flake-proof check.
**Commit:** docs(rules): throttle a rare flake until it fails and review a flake fix for its measured proof
**After:** none

**Decision:** flake-proof-throttle-and-review

- [x] 4. Check only the catalog's domains among frontend @Feature codes, and back-fill auth-ui
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/gymie-worktrees/self-review-kan-urgent-5-oct18`, `./gradlew verifyFeatureCatalog` fails `Orphan @Feature annotations not in catalog:` listing codes of uncatalogued domains (`workout-ui.*`, `ui-design-system.*`, …) — the defect.
  - [x] **Step 2: Apply the scan fix.** Apply this diff:

````diff verified:git apply at gymie 6a8eddc7; with it applied ./gradlew verifyFeatureCatalog lists only auth-ui orphans, and with the back-fill of Step 3 stubbed in it prints Feature catalog OK
diff --git a/gradle/feature-catalog.gradle.kts b/gradle/feature-catalog.gradle.kts
index e4e320ae..f3aa9d0c 100644
--- a/gradle/feature-catalog.gradle.kts
+++ b/gradle/feature-catalog.gradle.kts
@@ -88,32 +88,42 @@ tasks.register("verifyFeatureCatalog") {
                 root.parent?.parent?.parent?.resolve("gymie-frontend"),
             ).filterNotNull().firstOrNull { Files.exists(it) }
 
-        val scanRoots =
-            buildList {
-                add(root.resolve("src/app/auth/src/test"))
-                add(root.resolve("src/app/src/test"))
-                add(root.resolve("src/gateway/src/test"))
-                if (frontendRoot != null) {
-                    add(frontendRoot.resolve("shared/src/commonTest"))
-                    add(frontendRoot.resolve("shared/src/desktopTest"))
-                }
-            }
+        val backendScanRoots =
+            listOf(
+                root.resolve("src/app/auth/src/test"),
+                root.resolve("src/app/src/test"),
+                root.resolve("src/gateway/src/test"),
+            )
+        val frontendScanRoots =
+            listOfNotNull(
+                frontendRoot?.resolve("shared/src/commonTest"),
+                frontendRoot?.resolve("shared/src/desktopTest"),
+            )
 
-        val annotatedFeatures = mutableSetOf<String>()
         val featureAnnotationPattern = Regex("""@Feature\s*\(\s*"([^"]+)"\s*\)""")
 
-        scanRoots.forEach { scanRoot ->
-            if (!Files.exists(scanRoot)) return@forEach
-            Files.walk(scanRoot)
-                .filter { it.toString().endsWith(".kt") }
-                .forEach { file ->
-                    file.toFile().readText().let { content ->
-                        featureAnnotationPattern.findAll(content).forEach { match ->
-                            annotatedFeatures.add(match.groupValues[1])
-                        }
+        fun annotationsUnder(scanRoots: List<Path>): Set<String> =
+            scanRoots
+                .filter { Files.exists(it) }
+                .flatMap { scanRoot ->
+                    Files.walk(scanRoot).use { paths ->
+                        paths
+                            .filter { it.toString().endsWith(".kt") }
+                            .toList()
+                            .flatMap { file ->
+                                featureAnnotationPattern.findAll(file.toFile().readText()).map { it.groupValues[1] }.toList()
+                            }
                     }
-                }
-        }
+                }.toSet()
+
+        // The frontend names most of its codes after the OpenSpec capability they trace to
+        // (`workout-ui.*`, ...), which this catalog does not index (KAN-962). Of the frontend's
+        // annotations, only those in a domain the catalog holds are checked; a backend annotation
+        // is always checked.
+        val catalogDomains = catalogEntries.map { it.code.substringBefore('.') }.toSet()
+        val annotatedFeatures =
+            annotationsUnder(backendScanRoots) +
+                annotationsUnder(frontendScanRoots).filter { it.substringBefore('.') in catalogDomains }
 
         val orphanAnnotations = annotatedFeatures.filter { code -> catalogEntries.none { it.code == code } }
         if (orphanAnnotations.isNotEmpty()) {
````

    Run `./gradlew verifyFeatureCatalog` again: the orphans are now exactly these `auth-ui` codes — `auth-ui.banner.resend-action`, `auth-ui.error.field-level-error-stays-on-its-field`, `auth-ui.error.page-level-error-renders-on-a-flag-above-the-form`, `auth-ui.gate.failed-read-falls-through`, `auth-ui.home.logout-with-deep-link-slug`, `auth-ui.keyboard-avoidance.auth-content-stays-reachable`, `auth-ui.keyboard-avoidance.auth-screen-applies-inset`, `auth-ui.links.one-treatment`, `auth-ui.login.invite-required-copy`, `auth-ui.login.no-signup-control`, `auth-ui.oauth.account-conflict-message`, `auth-ui.oauth.email-not-verified-message`, `auth-ui.oauth.invite-required-message`, `auth-ui.oauth.missing-email-message`, `auth-ui.register.dead-invite`, `auth-ui.register.email-conflict-visible`, `auth-ui.register.gender`, `auth-ui.register.invite-preview`, `auth-ui.register.invite-preview-not-found`, `auth-ui.register.invite-token`, `auth-ui.register.invited-email-locked`, `auth-ui.register.opens-from-invite`, `auth-ui.register.preview-transport-failure`, `auth-ui.register.submit-gated-on-preview`, `auth-ui.session.boot-refresh-stores-the-rotated-tokens`.
  - [x] **Step 3: Back-fill and prune the catalog.** In `docs/features/catalog.yaml`, under `# --- auth-ui ---`, add one entry per code listed in Step 2, in the shape of the existing `auth-ui` entries: `code`, `title`, `spec: auth-ui`, `scenario`, `status: implemented`. Take `scenario` from the matching `#### Scenario:` heading in `openspec/specs/auth-ui/spec.md`; where no scenario matches, read the annotated test in the frontend (`git -C /Users/tweety53/Projects/gymie-frontend grep -n '<code>' -- shared/src`) and write one sentence naming the behaviour it asserts, used as both `title` and `scenario`. Delete the entries `auth-ui.navigation.open-register` and `baseline.schema.auth-tables-only`, whose annotations are gone (`design.md`, **Context**). Mirror both in the tables listing codes: `docs/features/README.md` and `docs/features/auth-ui/README.md` gain a row per added code (the latter's Primary tests column naming the test class that carries the annotation) and lose the removed codes' rows, and `docs/features/baseline/README.md` loses `baseline.schema.auth-tables-only`'s row. In `docs/features/README.md`'s **Rules**, replace `Every `@Feature` annotation must appear in `catalog.yaml` (no orphans).` with `Every backend `@Feature` annotation, and every frontend one in a domain the catalog holds, must appear in `catalog.yaml` (no orphans); the frontend's codes in other domains trace to their OpenSpec capability instead.`
  - [x] **Step 4: Verify.** `./gradlew verifyFeatureCatalog` prints `Feature catalog OK`, then `./gradlew ktlintFormat` and `./gradlew verifyChange` exit 0. Commit.

**Build:** green
**Files:** `gradle/feature-catalog.gradle.kts`, `docs/features/catalog.yaml`, `docs/features/README.md`, `docs/features/auth-ui/README.md`, `docs/features/baseline/README.md`
**Tests:** **none** — `verifyFeatureCatalog` is itself the guard: red before this task, green after.
**Regression:** reverting the gradle hunk fails `verifyFeatureCatalog` on every uncatalogued frontend domain again; reverting the catalog back-fill fails it on the auth-ui orphans.
**Commit:** fix(build): check only the catalog's domains among frontend @Feature codes
**After:** none

**Decision:** feature-catalog-frontend-scope
