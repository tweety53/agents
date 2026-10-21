# self-review-kan-920-flow-improvement-port-the-markdown-and-lint

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-920's six self-review findings, one task each — A1 (Task 1, dropped: review 1 disproved it), A2 (Task 2), A3
(Task 3), A4 (Task 4), S1 (Task 5), S2 (Task 6), as `proposal.md` names them.

**Architecture:** a new refusal in `flow record dispatch
begin`, a closed set for the decision's parent effort, an `output_unmeasured` count from the
harvester through pricing and both renderers, a concurrent `test` list, and the full suite moved
under pass 1 of the panel. Two migrations repair what A3 and A4 already stored; Task 2
repairs kan-920's own row directly.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-21-self-review-kan-920-flow-improvement-port-the-markdown-and-lint-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/kan-920-flow-improvement-port-the-markdown-and-lint`, branch `spectre/kan-920-flow-improvement-port-the-markdown-and-lint`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- Each task's diff is the unified diff its apply step gives, applied with `git apply` from the worktree's root, in task order; all six were applied in that order and checked in this worktree on a detached HEAD at `6a023063`. Save a diff by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork.
- `spectre/changes/self-review-kan-920-flow-improvement-port-the-markdown-and-lint/verbatim-moves.txt` already lists every sentence Tasks 2 and 6 add, so `scripts/check-verbatim-moves.sh` exits 0 after each of them.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container: the migrations of Tasks 3 and 4 apply when the operator next restarts `flowd`, and `design.md`'s **Live check** waits for that.
- The store write of Task 2 goes to the dev store only (`127.0.0.1:4173`, the `flow-postgres` container's `flow` database) — never anything else.

## Review Focus

- Task 2: only `-agent-id inline` without `-started-at` is judged; the change's own `tasks.md` is the one path allowed; a failing git call refuses nothing.
- Task 4: a message is unmeasured only when every line carries an explicit `"stop_reason": null`; an unmeasured message adds input and cache figures but no output or thinking; `cost` refuses any bucket with `output_unmeasured` above zero, and the combined main-plus-sidechain bucket carries it.
- Task 5: a failing command still fails the list; output stays in list order; skipped commands still print their skip line in place; the lock is outside the tree and released on exit.

- [ ] 1. DROPPED — not to be applied. Refuse a fixed round with no panel-fix row, and back-fill kan-920's
(decisions `unrecorded-fix-round`, `backfill-panel-fix-0`, both retired). **Disproved by review 1:** `skills/flow/review-panel.md` (lines 802-809 at this branch) has the parent fix a round that raised no Critical and no Important inline, with "no dispatch record"; kan-920's round 0 raised four Minors and nothing else, so the missing `panel-fix` row is the contract working, not a defect. The guard commit was dropped and the back-filled `panel-fix-0` row deleted from the dev store. No commit
carries it; nothing of it is to be applied.

**Files:** `stats/internal/guard/panelfixsingledispatch.go` (planned, never changed — task dropped)
**Tests:** none — task dropped
**After:** none

- [ ] 2. Refuse an inline dispatch begin over earlier work, and repair task 5's start
  - [ ] **Step 1: Reproduce before touching anything; do not guess.**
    `grep -c 'status", "--porcelain' stats/cmd/flow/record.go` prints `0`: no begin reads the
    worktree. kan-920's `task-5-implementer` row reads 22:48:47 to 22:48:54 +03:00 in
    `flow record dispatches -change kan-920-flow-improvement-port-the-markdown-and-lint -C .`.
    <!-- measured: flow record dispatches -change kan-920-flow-improvement-port-the-markdown-and-lint -C . on the dev store, read 2026-10-21 → started 22:48:47.687 +03:00, ended 22:48:54.752 +03:00 -->
  - [ ] **Step 2: Apply the diff** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at 6a023063 after Task 1's diff (since dropped); go test ./cmd/flow/ -count=1 passes, TestRecordDispatchBeginRefusesAnInlineRowOverEarlierWork fails with the refusal disabled, go vet and gofmt -l clean, check-verbatim-moves.sh exits 0
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index a5deed4f..67457881 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -199,6 +199,10 @@ these substitutions:
   no time; `flow record dispatch end` refuses a `completed` inline row closed under three seconds
   after its begin when it carries no `-tokens` report — which only a harness without a
   harvested transcript may pass; `-tokens` is refused where `CLAUDE_CODE_SESSION_ID` is set.
+  `flow record dispatch begin -agent-id inline` refuses while the worktree carries a tracked
+  modification other than the change's own `tasks.md`, naming each dirty path — work already
+  under way when its row begins (kan-920: task 5's row began after its edit and recorded 7
+  seconds).
 
 ## 1. Load context and validate the plan
 
diff --git a/stats/cmd/flow/record.go b/stats/cmd/flow/record.go
index 52286c8d..10baa1f8 100644
--- a/stats/cmd/flow/record.go
+++ b/stats/cmd/flow/record.go
@@ -1062,6 +1062,14 @@ func runRecordDispatchBegin(ctx context.Context, args []string, stdout, stderr i
 		fmt.Fprintf(stderr, "flow: %v\n", err)
 		return 2
 	}
+	// An inline row begins before its work's first edit (implement.md's
+	// Records); a back-filled row (-started-at) records work already done.
+	if *agentID == "inline" && *startedAt == "" {
+		if dirty := inlineDirtyPaths(f.dir, f.change); len(dirty) > 0 {
+			fmt.Fprintf(stderr, "flow: -agent-id inline begins before the work's first edit, and the worktree already carries tracked modifications: %s -- commit the earlier work, or begin the row before editing (skills/flow/implement.md)\n", strings.Join(dirty, ", "))
+			return 2
+		}
+	}
 
 	projectKey, _, err := fallback.ProjectKey(f.dir)
 	if err != nil {
@@ -3121,6 +3129,35 @@ func requireChangeCheckout(dir, change string, stderr io.Writer) bool {
 	return false
 }
 
+// inlineDirtyPaths is every tracked path of dir's worktree modified, staged
+// or not, save the change's own tasks.md -- close-task.sh's ticks stay
+// uncommitted between tasks. An inline row begun over any other records work
+// already under way (kan-920 self-review: task 5's row began after its
+// design.md edit and recorded 7 seconds). A git call that fails answers
+// nothing, and the begin is not refused on it.
+func inlineDirtyPaths(dir, change string) []string {
+	out, err := exec.Command("git", "-C", dir, "status", "--porcelain=v1", "-z", "--untracked-files=no").Output()
+	if err != nil {
+		return nil
+	}
+	own := "spectre/changes/" + change + "/tasks.md"
+	var dirty []string
+	entries := strings.Split(string(out), "\x00")
+	for i := 0; i < len(entries); i++ {
+		e := entries[i]
+		if len(e) < 4 {
+			continue
+		}
+		if e[0] == 'R' || e[0] == 'C' {
+			i++ // the rename's or copy's source path is the next entry
+		}
+		if p := e[3:]; p != own {
+			dirty = append(dirty, p)
+		}
+	}
+	return dirty
+}
+
 // mutationRoots is the -C repository's top level plus every key of the
 // worktrees map in the change's state record.
 func mutationRoots(ctx context.Context, f recordIdentityFlags, projectKey string) []string {
diff --git a/stats/cmd/flow/record_test.go b/stats/cmd/flow/record_test.go
index a42278c3..d1c2fd54 100644
--- a/stats/cmd/flow/record_test.go
+++ b/stats/cmd/flow/record_test.go
@@ -5679,3 +5679,63 @@ func TestRecordFindingFromFile(t *testing.T) {
 		t.Errorf("-note from: exit %d, stderr %q, posted %+v; want F6 recorded with note \"from\"", code, stderr.String(), posted)
 	}
 }
+
+// TestRecordDispatchBeginRefusesAnInlineRowOverEarlierWork pins kan-920's
+// self-review: task 5's inline row began after its design.md edit and
+// recorded 7 seconds. An inline begin refuses a worktree carrying tracked
+// modifications, naming them -- save the change's own tasks.md, whose
+// close-task ticks stay uncommitted -- while a dispatched row and a
+// back-filled one are not judged.
+func TestRecordDispatchBeginRefusesAnInlineRowOverEarlierWork(t *testing.T) {
+	repo := recordRepo(t, "kan-258")
+	isolatedStateRoot(t)
+	change := filepath.Join(repo, "spectre", "changes", "kan-258")
+	write := func(name, body string) {
+		t.Helper()
+		if err := os.WriteFile(filepath.Join(change, name), []byte(body), 0o644); err != nil {
+			t.Fatal(err)
+		}
+	}
+	write("tasks.md", "- [ ] 1. x\n")
+	write("design.md", "## Context\n")
+	for _, args := range [][]string{{"add", "-A"}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "plan"}} {
+		if out, err := exec.Command("git", append([]string{"-C", repo}, args...)...).CombinedOutput(); err != nil {
+			t.Fatalf("git %v: %v\n%s", args, err, out)
+		}
+	}
+	begin := func(extra ...string) (int, string, bool) {
+		t.Helper()
+		contacted := false
+		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
+			contacted = true
+			w.WriteHeader(http.StatusCreated)
+			_, _ = w.Write([]byte(`{"id":1,"seq":1}`))
+		}))
+		defer srv.Close()
+		argv := append([]string{"record", "dispatch", "begin", "-effort", "default", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
+			"-change", "kan-258", "-role", "implementer", "-model", "opus", "-key", "task-5-implementer",
+			"-session-token", "mf-record-inline-dirty"}, extra...)
+		var stdout, stderr bytes.Buffer
+		code := run(context.Background(), argv, strings.NewReader(""), &stdout, &stderr)
+		return code, stderr.String(), contacted
+	}
+
+	write("tasks.md", "- [x] 1. x\n")
+	if code, stderr, _ := begin("-agent-id", "inline"); code != 0 {
+		t.Fatalf("a ticked tasks.md alone: exit %d, want 0; stderr:\n%s", code, stderr)
+	}
+	write("design.md", "## Context\nedited\n")
+	code, stderr, contacted := begin("-agent-id", "inline")
+	if code != 2 || !strings.Contains(stderr, "spectre/changes/kan-258/design.md") || strings.Contains(stderr, "tasks.md") {
+		t.Fatalf("an inline begin over an edited design.md: exit %d, want 2 naming design.md alone; stderr:\n%s", code, stderr)
+	}
+	if contacted {
+		t.Error("a refused inline begin was sent to the daemon")
+	}
+	if code, stderr, _ := begin("-agent-id", "none"); code != 0 {
+		t.Errorf("a dispatched row over the same tree: exit %d, want 0; stderr:\n%s", code, stderr)
+	}
+	if code, stderr, _ := begin("-agent-id", "inline", "-started-at", "2026-10-21T19:48:16Z"); code != 0 {
+		t.Errorf("a back-filled inline row: exit %d, want 0; stderr:\n%s", code, stderr)
+	}
+}
````

  - [ ] **Step 3: Verify** — `cd stats && gofmt -l ./cmd/flow && go vet ./cmd/flow/ && go test ./cmd/flow/ -run 'TestRecordDispatchBegin' -count=1`
    passes with no `gofmt` output; `scripts/check-verbatim-moves.sh` and `scripts/check-references.sh` exit 0.
  - [ ] **Step 4: Repair kan-920's record** (decision `repair-task-5-start`):

```bash verified:column and table names read from information_schema on the dev store; the instant from the creating run's transcript, design.md's Context
docker exec flow-postgres psql -U flow -d flow -v ON_ERROR_STOP=1 -c "UPDATE dispatches d SET started_at = '2026-10-21T19:48:16.246Z' FROM changes c WHERE c.id = d.change_id AND c.name = 'kan-920-flow-improvement-port-the-markdown-and-lint' AND d.session_token = 'mf-k920b' AND d.dispatch_key = 'task-5-implementer'"
```

    It prints `UPDATE 1`; anything else is reported, never retried with a widened `WHERE`. Nothing of it is committed.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `skills/flow/implement.md`
**Tests:** `TestRecordDispatchBeginRefusesAnInlineRowOverEarlierWork`
**Regression:** `TestRecordDispatchBeginRefusesAnInlineRowOverEarlierWork` — reverting the commit records an inline begin over an edited `design.md`, and the test fails on its exit-2 assertion.
**Baseline:** before=121 after=122
<!-- measured: grep -cE '^func Test' stats/cmd/flow/record_test.go @ branch spectre/kan-920-flow-improvement-port-the-markdown-and-lint -->
**Commit:** fix(flow): refuse an inline dispatch begin over earlier work
**After:** none

**Decision:** inline-begin-clean-tree
**Decision:** repair-task-5-start

- [ ] 3. Record an unstated parent effort as default
  - [ ] **Step 1: Reproduce before touching anything; do not guess.**
    `grep -n 'unknown. where the harness does not state one' skills/flow/brainstorm-planner.md` prints
    one line, and the micro case of `TestDecisionRenderGolden` renders a `parent.effort` of `unknown`
    without complaint.
    <!-- measured: grep -c 'unknown. where the harness does not state one' skills/flow/brainstorm-planner.md @ branch spectre/kan-920-flow-improvement-port-the-markdown-and-lint → 1 -->
  - [ ] **Step 2: Apply the diff** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at 6a023063 after Tasks 1-2's diffs; go test ./cmd/flow/ -run TestDecision passes, go test ./internal/store/ -run Migrat passes, the migration's UPDATE ran inside a rolled-back transaction on the dev store, check-verbatim-moves.sh exits 0
diff --git a/skills/flow/brainstorm-planner.md b/skills/flow/brainstorm-planner.md
index 2d27bcca..a055a5c4 100644
--- a/skills/flow/brainstorm-planner.md
+++ b/skills/flow/brainstorm-planner.md
@@ -674,7 +674,8 @@ skipped for the cap is recorded as the string `"experimental": "skipped — bund
 sibling `groups_mechanical` (arrays of bundle ids), `groups_override` and
 `groups_reason` fields, or all four `null` when `execution` is inline), `visual` (an object
 `{verify, reason}`, `verify` one of `required`, `skipped` or `not configured`), `extra_dispatch` (an object `{slots, model, effort, reason}`, recorded on every class), `parent` (the parent's own
-model/effort, `unknown` where the harness does not state one), `overrides` (session-instruction
+model/effort, per **Harness mapping** (`skills/flow-contracts/model-policy.md`): `unknown
+(agent-defined)` and `default` where the session does not state them), `overrides` (session-instruction
 overrides to a *result*, each replacing the pair(s) it names for this run; empty unless one was
 given).
 
diff --git a/stats/cmd/flow/decision.go b/stats/cmd/flow/decision.go
index 904fe67f..f87827cc 100644
--- a/stats/cmd/flow/decision.go
+++ b/stats/cmd/flow/decision.go
@@ -7,6 +7,7 @@ import (
 	"fmt"
 	"io"
 	"os"
+	"slices"
 	"strconv"
 	"strings"
 )
@@ -127,7 +128,10 @@ type dDecision struct {
 	GroupsOverride   *string             `json:"groups_override"`
 	GroupsReason     string              `json:"groups_reason"`
 	Visual           *dVisual            `json:"visual"`
-	ExtraDispatch    *struct {
+	Parent           *struct {
+		Effort string `json:"effort"`
+	} `json:"parent"`
+	ExtraDispatch *struct {
 		Slots  []string `json:"slots"`
 		Model  string   `json:"model"`
 		Effort string   `json:"effort"`
@@ -253,6 +257,12 @@ func renderDecision(body []byte, _, reviewers string) (string, error) {
 			return "", fmt.Errorf("visual.reason is required")
 		}
 	}
+	// parent.effort is the -effort every inline dispatch row copies, so it
+	// takes that flag's closed set -- `default` where the session states none
+	// (kan-920 self-review: `unknown` was refused at task 1's begin).
+	if p := d.Parent; p != nil && !slices.Contains(recordEfforts, p.Effort) {
+		return "", fmt.Errorf("parent.effort %q is not one of: %s", p.Effort, strings.Join(recordEfforts, ", "))
+	}
 	var panelDefault string
 	var p dPanel
 	if json.Unmarshal(d.Panel, &panelDefault) != nil {
diff --git a/stats/cmd/flow/decision_test.go b/stats/cmd/flow/decision_test.go
index c87fc9de..9da57254 100644
--- a/stats/cmd/flow/decision_test.go
+++ b/stats/cmd/flow/decision_test.go
@@ -49,7 +49,7 @@ func TestDecisionRenderGolden(t *testing.T) {
 			"implementer":"skipped — inline","fixer":"skipped — inline","panel":"default",
 			"groups":null,"groups_mechanical":null,"groups_override":null,"groups_reason":null,
 			"visual":{"verify":"required","reason":"the Files list carries a page component"},
-			"parent":{"model":"claude-opus-5-5","effort":"unknown"},"overrides":[]}`,
+			"parent":{"model":"claude-opus-5-5","effort":"default"},"overrides":[]}`,
 			"## Decision\n\n" +
 				"- **Class:** micro · tasks 1 · files 1 · repos 1\n" +
 				"- **Execution:** inline\n" +
@@ -148,6 +148,7 @@ func TestDecisionRenderRefusals(t *testing.T) {
 		{"missing groups", strings.Replace(good, `,"groups":null`, "", 1), append([]string{"render", "-file", "<file>"}, flags...), "flow: decision render: missing required field \"groups\"\n"},
 		{"visual.verify outside the set", strings.Replace(good, `"groups":null`, `"groups":null,"visual":{"verify":"skip","reason":"x"}`, 1), append([]string{"render", "-file", "<file>"}, flags...), "flow: decision render: visual.verify \"skip\" is not one of required, skipped, not configured\n"},
 		{"visual with a blank reason", strings.Replace(good, `"groups":null`, `"groups":null,"visual":{"verify":"skipped","reason":" "}`, 1), append([]string{"render", "-file", "<file>"}, flags...), "flow: decision render: visual.reason is required\n"},
+		{"parent.effort outside the -effort set", strings.Replace(good, `"groups":null`, `"groups":null,"parent":{"model":"unknown (agent-defined)","effort":"unknown"}`, 1), append([]string{"render", "-file", "<file>"}, flags...), "flow: decision render: parent.effort \"unknown\" is not one of: low, medium, high, default\n"},
 		{"missing panel.rerun_dispatch", strings.Replace(good, `"panel":"default"`, `"panel":{"compact":true,"rerun":"delta","grouping":"static","dispatches":[]}`, 1), append([]string{"render", "-file", "<file>"}, flags...), "flow: decision render: missing required field \"panel.rerun_dispatch\"\n"},
 	}
 	for _, c := range cases {
diff --git a/stats/internal/store/migrations/0041_decision_parent_effort_default.sql b/stats/internal/store/migrations/0041_decision_parent_effort_default.sql
new file mode 100644
index 00000000..639168ce
--- /dev/null
+++ b/stats/internal/store/migrations/0041_decision_parent_effort_default.sql
@@ -0,0 +1,7 @@
+-- 0041_decision_parent_effort_default.sql: a decision's parent effort is the -effort every inline
+-- dispatch row copies, so it takes that flag's closed set (low, medium, high, default). Decide
+-- wrote `unknown` where the session stated none, which the flag refuses; `default` is the flag's
+-- word for that case (kan-920 self-review). `flow decision render` refuses it from now on.
+UPDATE decisions
+SET decision = jsonb_set(decision, '{parent,effort}', '"default"')
+WHERE decision->'parent'->>'effort' = 'unknown';
````

  - [ ] **Step 3: Verify** — `cd stats && gofmt -l ./cmd/flow && go vet ./cmd/flow/ && go test ./cmd/flow/ -run 'TestDecision' -count=1 && go test ./internal/store/ -run 'Migrat' -count=1`
    passes with no `gofmt` output; `scripts/check-verbatim-moves.sh` and `scripts/check-references.sh` exit 0.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`, `stats/cmd/flow/decision.go`, `stats/cmd/flow/decision_test.go`, `stats/internal/store/migrations/0041_decision_parent_effort_default.sql`
**Tests:** `parent.effort outside the -effort set`
**Regression:** `parent.effort outside the -effort set` — reverting the commit renders a decision whose parent effort is `unknown`, and the refusal case fails on its exit code.
**Baseline:** before=17 after=18
<!-- measured: grep -cE '^[[:blank:]]{2}[{]"' stats/cmd/flow/decision_test.go @ branch spectre/kan-920-flow-improvement-port-the-markdown-and-lint -->
**Commit:** fix(flow): record an unstated parent effort as default
**After:** none

**Decision:** parent-effort-default

- [ ] 4. Mark stream-start subagent output not measured
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** The panel slot's transcript
    `~/.claude/projects/-Users-tweety53-Projects-agents/31fac4f3-d562-4a7c-be01-42daf8793a8d/subagents/agent-ac204c32268548ea7.jsonl`
    carries `"stop_reason": null` on all 30 assistant lines and sums to the 184 output tokens its
    dispatch row stores.
    <!-- measured: a python3 count over that transcript, a machine-local file, read 2026-10-21 → 30 lines with stop_reason null, per-message maximum output summing to 184 -->
  - [ ] **Step 2: Apply the diff** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at 6a023063 after Tasks 1-3's diffs; go test ./internal/... ./cmd/... -count=1 passes, npm test and npx tsc -b pass, the fixture regenerated with UPDATE_METRICS_FIXTURE=1, both UPDATEs of the migration ran inside a rolled-back transaction on the dev store
diff --git a/stats/internal/harvest/attribute.go b/stats/internal/harvest/attribute.go
index 2becc024..1582d408 100644
--- a/stats/internal/harvest/attribute.go
+++ b/stats/internal/harvest/attribute.go
@@ -146,11 +146,22 @@ type Bucket struct {
 	CacheCreationUnknown int64 `json:"cache_creation_unknown,omitempty"`
 	CacheRead            int64 `json:"cache_read"`
 	Thinking             int64 `json:"thinking"`
+	// OutputUnmeasured counts the messages whose output no transcript line
+	// measured (Usage.OutputUnmeasured). Their output and thinking are left
+	// out of Output and Thinking, so a bucket carrying any reads its output
+	// as not measured, never as the first-chunk sum (kan-920 self-review);
+	// store.Store.Price refuses to price it.
+	OutputUnmeasured int64 `json:"output_unmeasured,omitempty"`
 }
 
 func (b *Bucket) add(u Usage) {
 	b.Input += u.InputTokens
-	b.Output += u.OutputTokens
+	if u.OutputUnmeasured {
+		b.OutputUnmeasured++
+	} else {
+		b.Output += u.OutputTokens
+		b.Thinking += u.ThinkingTokens
+	}
 	b.CacheCreation += u.CacheCreationInputTokens
 	if u.CacheSplitKnown {
 		b.CacheCreation5m += u.CacheCreation5mTokens
@@ -159,7 +170,6 @@ func (b *Bucket) add(u Usage) {
 		b.CacheCreationUnknown += u.CacheCreationInputTokens
 	}
 	b.CacheRead += u.CacheReadInputTokens
-	b.Thinking += u.ThinkingTokens
 }
 
 // TokenDelta is the usage found in *one batch* of newly read records,
@@ -926,6 +936,7 @@ func splitUsage(u Usage, weights []float64) []Usage {
 			CacheReadInputTokens:     parts[4][j],
 			OutputTokens:             parts[5][j],
 			ThinkingTokens:           parts[6][j],
+			OutputUnmeasured:         u.OutputUnmeasured,
 		}
 	}
 	return out
diff --git a/stats/internal/harvest/attribute_test.go b/stats/internal/harvest/attribute_test.go
index 46c51441..5310d341 100644
--- a/stats/internal/harvest/attribute_test.go
+++ b/stats/internal/harvest/attribute_test.go
@@ -170,6 +170,32 @@ func TestMessageOutsideItsWindowIntervalIsNotAttributed(t *testing.T) {
 
 func ptrTime(t time.Time) *time.Time { return &t }
 
+// TestAttributeCountsUnmeasuredOutputApart pins kan-920's self-review: a
+// message whose output no transcript line measured adds its input to the
+// dispatch's bucket and one to output_unmeasured, never its first-chunk
+// output or thinking.
+func TestAttributeCountsUnmeasuredOutputApart(t *testing.T) {
+	at := mustParse(t, "2026-01-01T00:00:01Z")
+	records := []harvest.Record{
+		{Timestamp: at, SessionID: mainSessionID, IsSidechain: true, MessageID: "a",
+			Usage: harvest.Usage{InputTokens: 5, OutputTokens: 28, ThinkingTokens: 9, OutputUnmeasured: true}},
+		{Timestamp: at, SessionID: mainSessionID, IsSidechain: true, MessageID: "b",
+			Usage: harvest.Usage{InputTokens: 7, OutputTokens: 400, ThinkingTokens: 50}},
+	}
+	windows := &fakeDispatchWindowSource{bySession: map[string][]harvest.DispatchWindow{
+		mainSessionID: {{DispatchID: 5, StartedAt: mustParse(t, "2026-01-01T00:00:00Z"),
+			EndedAt: ptrTime(mustParse(t, "2026-01-01T00:00:02Z"))}},
+	}}
+	deltas, err := harvest.NewDispatchAttributor(windows).Attribute(context.Background(), records)
+	if err != nil {
+		t.Fatalf("Attribute: %v", err)
+	}
+	got := deltas[5].Tokens.Sidechain
+	if got.Input != 12 || got.Output != 400 || got.Thinking != 50 || got.OutputUnmeasured != 1 {
+		t.Errorf("sidechain = %+v, want input 12, output 400, thinking 50, output_unmeasured 1", got)
+	}
+}
+
 // TestSidechainAccumulatesSeparately attributes main-thread and sidechain
 // fixture records that share one session id into one window, and checks
 // that main and sidechain totals land in separate buckets rather than
@@ -212,8 +238,10 @@ func TestSidechainAccumulatesSeparately(t *testing.T) {
 	if d.Total.Sidechain.CacheRead != 77676 {
 		t.Errorf("tokens.sidechain.cache_read = %v, want 77676", d.Total.Sidechain.CacheRead)
 	}
-	if d.Total.Sidechain.Output != 125 {
-		t.Errorf("tokens.sidechain.output = %v, want 125", d.Total.Sidechain.Output)
+	// The fixture's four sidechain lines carry "stop_reason": null, as a
+	// subagent transcript writes them: their output is not measured.
+	if d.Total.Sidechain.Output != 0 || d.Total.Sidechain.OutputUnmeasured != 4 {
+		t.Errorf("tokens.sidechain output = %v, output_unmeasured = %v, want 0 and 4", d.Total.Sidechain.Output, d.Total.Sidechain.OutputUnmeasured)
 	}
 }
 
@@ -1082,8 +1110,8 @@ func TestDispatchWindowAttributesSidechainUsage(t *testing.T) {
 	if d.Tokens.Sidechain.CacheRead != 77676 {
 		t.Errorf("tokens.sidechain.cache_read = %v, want 77676", d.Tokens.Sidechain.CacheRead)
 	}
-	if d.Tokens.Sidechain.Output != 125 {
-		t.Errorf("tokens.sidechain.output = %v, want 125", d.Tokens.Sidechain.Output)
+	if d.Tokens.Sidechain.Output != 0 || d.Tokens.Sidechain.OutputUnmeasured != 4 {
+		t.Errorf("tokens.sidechain output = %v, output_unmeasured = %v, want 0 and 4 -- the fixture's stream-start lines", d.Tokens.Sidechain.Output, d.Tokens.Sidechain.OutputUnmeasured)
 	}
 }
 
@@ -1298,7 +1326,7 @@ func TestDispatchAttributionLeavesStageAttributionUnchanged(t *testing.T) {
 			CacheRead: 109042,
 		},
 		Sidechain: harvest.Bucket{
-			Input: 8, Output: 125,
+			Input: 8, OutputUnmeasured: 4,
 			CacheCreation: 35348, CacheCreation5m: 35348,
 			CacheRead: 77676,
 		},
@@ -1365,11 +1393,11 @@ func TestDispatchWindowIntervalIsHalfOpen(t *testing.T) {
 		// Only the fixture's first record (00:05:00.500) precedes the
 		// seam; the record *on* the seam and the two after it belong to
 		// the window that is starting.
-		if got := deltas[1].Tokens.Sidechain.Output; got != 30 {
-			t.Errorf("order %v: dispatch 1 output = %v, want 30 (its one record before the seam)", order, got)
+		if got := deltas[1].Tokens.Sidechain.OutputUnmeasured; got != 1 {
+			t.Errorf("order %v: dispatch 1 output_unmeasured = %v, want 1 (its one record before the seam)", order, got)
 		}
-		if got := deltas[2].Tokens.Sidechain.Output; got != 95 {
-			t.Errorf("order %v: dispatch 2 output = %v, want 95 (the record on the seam plus the two after it)", order, got)
+		if got := deltas[2].Tokens.Sidechain.OutputUnmeasured; got != 3 {
+			t.Errorf("order %v: dispatch 2 output_unmeasured = %v, want 3 (the record on the seam plus the two after it)", order, got)
 		}
 		if got := deltas[1].Tokens.Sidechain.CacheCreation; got != 20000 {
 			t.Errorf("order %v: dispatch 1 cache_creation = %v, want 20000", order, got)
diff --git a/stats/internal/harvest/transcript.go b/stats/internal/harvest/transcript.go
index 273ccce0..4503c612 100644
--- a/stats/internal/harvest/transcript.go
+++ b/stats/internal/harvest/transcript.go
@@ -88,6 +88,15 @@ type Usage struct {
 	CacheReadInputTokens     int64
 	OutputTokens             int64
 	ThinkingTokens           int64
+	// OutputUnmeasured is set when every line of the message carries an
+	// explicit "stop_reason": null. A subagent transcript writes each
+	// message's usage at stream start that way and never its final count,
+	// so OutputTokens and ThinkingTokens are a first-chunk figure, never the
+	// message's (kan-920 self-review: a 5035-character tool input recorded
+	// 28 output tokens). A line with no stop_reason key at all is an older
+	// shape and stays measured. Bucket.add counts such a message instead of
+	// adding its output.
+	OutputUnmeasured bool
 	// Speed is the message's own "speed" field verbatim -- "standard",
 	// "fast", or "" when the transcript line carried none (an older
 	// format, or a harness that never sets it). Fast mode doubles Opus's
@@ -188,10 +197,11 @@ type rawCompactMeta struct {
 }
 
 type rawMessage struct {
-	ID      string          `json:"id"`
-	Model   string          `json:"model"`
-	Usage   *rawUsage       `json:"usage"`
-	Content json.RawMessage `json:"content"`
+	ID         string          `json:"id"`
+	Model      string          `json:"model"`
+	Usage      *rawUsage       `json:"usage"`
+	Content    json.RawMessage `json:"content"`
+	StopReason json.RawMessage `json:"stop_reason"`
 }
 
 // contentBlocks decodes a message's "content" field into its content
@@ -356,11 +366,13 @@ func ParseAssistantRecords(complete []byte) []Record {
 		if u.OutputTokensDetails != nil {
 			thinking = u.OutputTokensDetails.ThinkingTokens
 		}
+		streaming := string(raw.Message.StopReason) == "null"
 		if id := raw.Message.ID; id != "" {
 			if i, ok := seen[id]; ok {
 				kept := &out[i].Usage
 				kept.OutputTokens = max(kept.OutputTokens, u.OutputTokens)
 				kept.ThinkingTokens = max(kept.ThinkingTokens, thinking)
+				kept.OutputUnmeasured = kept.OutputUnmeasured && streaming
 				continue
 			}
 			seen[id] = len(out)
@@ -391,6 +403,7 @@ func ParseAssistantRecords(complete []byte) []Record {
 				CacheReadInputTokens:     u.CacheReadInputTokens,
 				OutputTokens:             u.OutputTokens,
 				ThinkingTokens:           thinking,
+				OutputUnmeasured:         streaming,
 				Speed:                    u.Speed,
 			},
 		})
diff --git a/stats/internal/harvest/transcript_test.go b/stats/internal/harvest/transcript_test.go
index 9da6c23a..d6aadea4 100644
--- a/stats/internal/harvest/transcript_test.go
+++ b/stats/internal/harvest/transcript_test.go
@@ -364,6 +364,29 @@ func TestParseAssistantRecordsTakesLargestOutputPerMessage(t *testing.T) {
 	}
 }
 
+// TestParseAssistantRecordsMarksStreamingOutputUnmeasured pins kan-920's
+// self-review: a subagent transcript writes every line of a message with an
+// explicit "stop_reason": null and its stream-start usage, so its output is
+// not measured. One line carrying a stop_reason measures the message; a line
+// with no stop_reason key is an older shape and stays measured.
+func TestParseAssistantRecordsMarksStreamingOutputUnmeasured(t *testing.T) {
+	line := func(id, stop string) string {
+		return fmt.Sprintf(`{"type":"assistant","timestamp":"2026-01-01T00:00:01Z","sessionId":"s","message":{"id":%q,"model":"m",%s"usage":{"input_tokens":5,"output_tokens":28}}}`+"\n", id, stop)
+	}
+	records := harvest.ParseAssistantRecords([]byte(
+		line("msg_stream", `"stop_reason":null,`) + line("msg_stream", `"stop_reason":null,`) +
+			line("msg_final", `"stop_reason":null,`) + line("msg_final", `"stop_reason":"tool_use",`) +
+			line("msg_old", "")))
+	if len(records) != 3 {
+		t.Fatalf("got %d records, want 3", len(records))
+	}
+	for i, want := range []bool{true, false, false} {
+		if got := records[i].Usage.OutputUnmeasured; got != want {
+			t.Errorf("%s: OutputUnmeasured = %v, want %v", records[i].MessageID, got, want)
+		}
+	}
+}
+
 // TestParseAssistantRecordsAgentIDAbsent is the negative companion: a
 // parent-session line carries no "agentId" at all, and must decode to
 // AgentID == "" -- never a placeholder, per this change's own
diff --git a/stats/internal/harvest/watcher.go b/stats/internal/harvest/watcher.go
index 9bbb4f09..a8900513 100644
--- a/stats/internal/harvest/watcher.go
+++ b/stats/internal/harvest/watcher.go
@@ -2029,8 +2029,9 @@ func dropCarriedMessage(records []Record, carried carriedMessage) []Record {
 		if r.Signal == nil && r.MessageID == carried.id {
 			if r.Usage.OutputTokens > carried.output {
 				r.Usage = Usage{
-					OutputTokens:   r.Usage.OutputTokens - carried.output,
-					ThinkingTokens: max(0, r.Usage.ThinkingTokens-carried.thinking),
+					OutputTokens:     r.Usage.OutputTokens - carried.output,
+					ThinkingTokens:   max(0, r.Usage.ThinkingTokens-carried.thinking),
+					OutputUnmeasured: r.Usage.OutputUnmeasured,
 				}
 				out = append(out, r)
 			}
diff --git a/stats/internal/harvest/wireshape_test.go b/stats/internal/harvest/wireshape_test.go
index 331e01ef..ae3f676a 100644
--- a/stats/internal/harvest/wireshape_test.go
+++ b/stats/internal/harvest/wireshape_test.go
@@ -31,12 +31,12 @@ var spaMetricsFixturePatch = harvest.MetricsPatch{
 		Main: harvest.Bucket{
 			Input: 600000, Output: 300000, CacheCreation: 50000,
 			CacheCreation5m: 30000, CacheCreation1h: 15000, CacheCreationUnknown: 5000,
-			CacheRead: 950000, Thinking: 10000,
+			CacheRead: 950000, Thinking: 10000, OutputUnmeasured: 2,
 		},
 		Sidechain: harvest.Bucket{
 			Input: 400000, Output: 100000, CacheCreation: 20000,
 			CacheCreation5m: 12000, CacheCreation1h: 6000, CacheCreationUnknown: 2000,
-			CacheRead: 300000, Thinking: 5000,
+			CacheRead: 300000, Thinking: 5000, OutputUnmeasured: 3,
 		},
 	},
 }
diff --git a/stats/internal/records/render.go b/stats/internal/records/render.go
index 4559e103..f0b7d9cf 100644
--- a/stats/internal/records/render.go
+++ b/stats/internal/records/render.go
@@ -395,10 +395,11 @@ type tokenTotals struct {
 }
 
 type tokenBucket struct {
-	Input         int64 `json:"input"`
-	Output        int64 `json:"output"`
-	CacheCreation int64 `json:"cache_creation"`
-	CacheRead     int64 `json:"cache_read"`
+	Input            int64 `json:"input"`
+	Output           int64 `json:"output"`
+	CacheCreation    int64 `json:"cache_creation"`
+	CacheRead        int64 `json:"cache_read"`
+	OutputUnmeasured int64 `json:"output_unmeasured"`
 }
 
 // tokenLine renders a dispatch's token figures, or the words that say why
@@ -467,9 +468,15 @@ func tokenLine(raw json.RawMessage) string {
 	if m.Tokens != nil {
 		t := *m.Tokens
 		sum := func(a, b int64) int64 { return a + b }
-		line := fmt.Sprintf("input %d, output %d, cache read %d, cache creation %d",
+		// A bucket carrying output_unmeasured left out output no transcript
+		// line measured (internal/harvest's Bucket), so its sum is no figure.
+		output := strconv.FormatInt(sum(t.Main.Output, t.Sidechain.Output), 10)
+		if sum(t.Main.OutputUnmeasured, t.Sidechain.OutputUnmeasured) > 0 {
+			output = reasonNotMeasured
+		}
+		line := fmt.Sprintf("input %d, output %s, cache read %d, cache creation %d",
 			sum(t.Main.Input, t.Sidechain.Input),
-			sum(t.Main.Output, t.Sidechain.Output),
+			output,
 			sum(t.Main.CacheRead, t.Sidechain.CacheRead),
 			sum(t.Main.CacheCreation, t.Sidechain.CacheCreation))
 		if m.Apportioned != nil && m.Apportioned.Records > 0 {
diff --git a/stats/internal/records/render_test.go b/stats/internal/records/render_test.go
index 1e2b1a9f..5e02341f 100644
--- a/stats/internal/records/render_test.go
+++ b/stats/internal/records/render_test.go
@@ -359,6 +359,24 @@ func TestTokenLineRendersApportionedQualifier(t *testing.T) {
 	}
 }
 
+// TestTokenLineRendersUnmeasuredOutput pins kan-920's self-review: a bucket
+// carrying output_unmeasured left out output no transcript line measured, so
+// its output renders as not measured, never as the undercount.
+func TestTokenLineRendersUnmeasuredOutput(t *testing.T) {
+	out := records.RenderLedger(records.Run{
+		Change: "demo",
+		Dispatches: []records.Dispatch{{
+			Seq: 1, Role: "reviewer", Model: "opus",
+			StartedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
+			Metrics:   json.RawMessage(`{"tokens":{"main":{"input":0,"output":0,"cache_read":0,"cache_creation":0},"sidechain":{"input":32,"output":184,"cache_read":811464,"cache_creation":75632,"output_unmeasured":1}}}`),
+		}},
+	})
+	want := "- Tokens: input 32, output not measured, cache read 811464, cache creation 75632\n"
+	if !strings.Contains(out, want) {
+		t.Errorf("unmeasured output must render as not measured %q:\n%s", strings.TrimSuffix(want, "\n"), out)
+	}
+}
+
 // TestTokenLineWithoutTokensIgnoresApportioned pins the pointer rule on
 // the read side: an `apportioned` record without a `tokens` object is
 // contradictory input the render resolves the same way it resolves a
diff --git a/stats/internal/store/migrations/0042_dispatch_output_unmeasured.sql b/stats/internal/store/migrations/0042_dispatch_output_unmeasured.sql
new file mode 100644
index 00000000..ffb82505
--- /dev/null
+++ b/stats/internal/store/migrations/0042_dispatch_output_unmeasured.sql
@@ -0,0 +1,18 @@
+-- 0042_dispatch_output_unmeasured.sql: a subagent transcript writes each message's usage at stream
+-- start ("stop_reason": null) and never its final count, so every sidechain output figure the
+-- harvester stored before kan-920's self-review is a first-chunk sum, never the real output. The
+-- true counts are in no file; the harvester now counts such messages under output_unmeasured
+-- instead, and this marks every dispatch row and stage run already stored with sidechain output the
+-- same way, so the ledger and the SPA render that output as not measured rather than as the
+-- undercount. A cost_usd already written stays.
+UPDATE dispatches
+SET metrics = jsonb_set(metrics, '{tokens,sidechain,output_unmeasured}', '1')
+WHERE jsonb_typeof(metrics->'tokens'->'sidechain') = 'object'
+  AND COALESCE((metrics->'tokens'->'sidechain'->>'output')::numeric, 0) > 0
+  AND NOT (metrics->'tokens'->'sidechain' ? 'output_unmeasured');
+
+UPDATE stage_runs
+SET metrics = jsonb_set(metrics, '{tokens,sidechain,output_unmeasured}', '1')
+WHERE jsonb_typeof(metrics->'tokens'->'sidechain') = 'object'
+  AND COALESCE((metrics->'tokens'->'sidechain'->>'output')::numeric, 0) > 0
+  AND NOT (metrics->'tokens'->'sidechain' ? 'output_unmeasured');
diff --git a/stats/internal/store/pricing.go b/stats/internal/store/pricing.go
index e0c7343d..9a2af758 100644
--- a/stats/internal/store/pricing.go
+++ b/stats/internal/store/pricing.go
@@ -98,6 +98,9 @@ type chargeableTokens struct {
 	CacheCreation1h      *float64 `json:"cache_creation_1h"`
 	CacheCreationUnknown *float64 `json:"cache_creation_unknown"`
 	CacheRead            *float64 `json:"cache_read"`
+	// OutputUnmeasured is internal/harvest's Bucket.OutputUnmeasured: how
+	// many messages' output the bucket's Output leaves out, never a charge.
+	OutputUnmeasured *float64 `json:"output_unmeasured"`
 }
 
 // hasCharge reports whether t carries at least one of the chargeable
@@ -141,6 +144,13 @@ func (t chargeableTokens) hasCharge() bool {
 //     model with no fast rate is unpriceable, not silently priced at the
 //     standard rate.
 func (t chargeableTokens) cost(rate PricingRate, fast bool) (cost float64, ok bool) {
+	// A bucket whose Output leaves out messages no transcript line measured
+	// prices nothing: a cost from it reads as complete while under-pricing
+	// output (kan-920 self-review) -- the partial-sum failure the
+	// CacheCreationUnknown rule below refuses too.
+	if t.OutputUnmeasured != nil && *t.OutputUnmeasured > 0 {
+		return 0, false
+	}
 	if t.CacheCreationUnknown != nil {
 		if rate.CacheWrite1hPerMTok == nil || *rate.CacheWrite1hPerMTok != rate.CacheWrite5mPerMTok {
 			return 0, false
@@ -220,6 +230,7 @@ func combineMainAndSidechain(main, sidechain *chargeableTokens) chargeableTokens
 		CacheCreation1h:      sumOptional(m.CacheCreation1h, sc.CacheCreation1h),
 		CacheCreationUnknown: sumOptional(m.CacheCreationUnknown, sc.CacheCreationUnknown),
 		CacheRead:            sumOptional(m.CacheRead, sc.CacheRead),
+		OutputUnmeasured:     sumOptional(m.OutputUnmeasured, sc.OutputUnmeasured),
 	}
 }
 
diff --git a/stats/internal/store/pricing_test.go b/stats/internal/store/pricing_test.go
index 53f7985d..9d61e8a8 100644
--- a/stats/internal/store/pricing_test.go
+++ b/stats/internal/store/pricing_test.go
@@ -64,3 +64,19 @@ func TestCostPricesUnknownSplitAtFlatRate(t *testing.T) {
 }
 
 func ptr(v float64) *float64 { return &v }
+
+// TestCostRefusesUnmeasuredOutput pins kan-920's self-review: a bucket whose
+// output_unmeasured counts messages its Output leaves out prices nothing --
+// a cost from it would read as complete while under-pricing output.
+func TestCostRefusesUnmeasuredOutput(t *testing.T) {
+	rate := PricingRate{Model: "claude-test", InputPerMTok: 1, OutputPerMTok: 2, CacheWrite5mPerMTok: 6.25, CacheReadPerMTok: 0.1}
+	if _, ok := (chargeableTokens{Input: ptr(1000), Output: ptr(400), OutputUnmeasured: ptr(1)}).cost(rate, false); ok {
+		t.Error("ok = true, want false: output_unmeasured leaves output out of the bucket")
+	}
+	if _, ok := (chargeableTokens{Input: ptr(1000), Output: ptr(400)}).cost(rate, false); !ok {
+		t.Error("ok = false, want true: a fully measured bucket prices")
+	}
+	if got := combineMainAndSidechain(&chargeableTokens{Input: ptr(1)}, &chargeableTokens{OutputUnmeasured: ptr(2)}).OutputUnmeasured; got == nil || *got != 2 {
+		t.Errorf("combined output_unmeasured = %v, want 2", got)
+	}
+}
diff --git a/stats/web/src/metrics.test.ts b/stats/web/src/metrics.test.ts
index 204a93e6..6696af6f 100644
--- a/stats/web/src/metrics.test.ts
+++ b/stats/web/src/metrics.test.ts
@@ -144,11 +144,18 @@ describe("token readers", () => {
 
   it("sum a present numeric field across both buckets, for each per-field key", () => {
     expect(readInputTokens(bag)).toBe(main.input + sidechain.input);
-    expect(readOutputTokens(bag)).toBe(main.output + sidechain.output);
+    expect(readOutputTokens({ tokens: { main: { ...main, output_unmeasured: 0 }, sidechain: { ...sidechain, output_unmeasured: 0 } } })).toBe(
+      main.output + sidechain.output,
+    );
     expect(readCacheReadTokens(bag)).toBe(main.cache_read + sidechain.cache_read);
     expect(readCacheCreationTokens(bag)).toBe(main.cache_creation + sidechain.cache_creation);
   });
 
+  it("read output as absent when a bucket carries output_unmeasured", () => {
+    expect(readOutputTokens(bag)).toBeNull();
+    expect(readOutputTokens({ tokens: { main: { output: 9 }, sidechain: { output: 184, output_unmeasured: 1 } } })).toBeNull();
+  });
+
   it("sum each bucket's own fields into that bucket's grand total", () => {
     expect(readMainTokens(bag)).toBe(main.input + main.output + main.cache_creation + main.cache_read + main.thinking);
     expect(readSidechainTokens(bag)).toBe(
diff --git a/stats/web/src/metrics.ts b/stats/web/src/metrics.ts
index 02eb203b..eb4a3445 100644
--- a/stats/web/src/metrics.ts
+++ b/stats/web/src/metrics.ts
@@ -172,8 +172,14 @@ export function readInputTokens(bag: unknown): number | null {
   return readSummedBucketField(bag, "input");
 }
 
-/** `output`, summed across `tokens.main` and `tokens.sidechain`. */
+/**
+ * `output`, summed across `tokens.main` and `tokens.sidechain` -- or null
+ * when either bucket carries `output_unmeasured`: `internal/harvest.Bucket`
+ * then leaves out the output no transcript line measured, so the sum is no
+ * figure (kan-920 self-review).
+ */
 export function readOutputTokens(bag: unknown): number | null {
+  if ((readSummedBucketField(bag, "output_unmeasured") ?? 0) > 0) return null;
   return readSummedBucketField(bag, "output");
 }
 
diff --git a/stats/web/src/testdata/metricsPatchFixture.json b/stats/web/src/testdata/metricsPatchFixture.json
index b9701572..52e84799 100644
--- a/stats/web/src/testdata/metricsPatchFixture.json
+++ b/stats/web/src/testdata/metricsPatchFixture.json
@@ -8,7 +8,8 @@
       "cache_creation_1h": 15000,
       "cache_creation_unknown": 5000,
       "cache_read": 950000,
-      "thinking": 10000
+      "thinking": 10000,
+      "output_unmeasured": 2
     },
     "sidechain": {
       "input": 400000,
@@ -18,7 +19,8 @@
       "cache_creation_1h": 6000,
       "cache_creation_unknown": 2000,
       "cache_read": 300000,
-      "thinking": 5000
+      "thinking": 5000,
+      "output_unmeasured": 3
     }
   }
 }
````

  - [ ] **Step 3: Verify** — `cd stats && gofmt -l ./internal && go vet ./internal/... && go test ./internal/harvest/ ./internal/records/ -count=1 && go test ./internal/store/ -run 'TestCost|Migrat' -count=1`
    passes with no `gofmt` output, and `cd stats/web && npx tsc -b && npx vitest run src/metrics.test.ts -t 'read output as absent when a bucket carries output_unmeasured'` passes.

**Build:** green
**Files:** `stats/internal/harvest/transcript.go`, `stats/internal/harvest/attribute.go`, `stats/internal/harvest/watcher.go`, `stats/internal/harvest/transcript_test.go`, `stats/internal/harvest/attribute_test.go`, `stats/internal/harvest/wireshape_test.go`, `stats/internal/store/pricing.go`, `stats/internal/store/pricing_test.go`, `stats/internal/store/migrations/0042_dispatch_output_unmeasured.sql`, `stats/internal/records/render.go`, `stats/internal/records/render_test.go`, `stats/web/src/metrics.ts`, `stats/web/src/metrics.test.ts`, `stats/web/src/testdata/metricsPatchFixture.json`
**Tests:** `TestParseAssistantRecordsMarksStreamingOutputUnmeasured`, `TestAttributeCountsUnmeasuredOutputApart`, `TestCostRefusesUnmeasuredOutput`, `TestTokenLineRendersUnmeasuredOutput`, `read output as absent when a bucket carries output_unmeasured`
**Regression:** `TestParseAssistantRecordsMarksStreamingOutputUnmeasured`, `TestAttributeCountsUnmeasuredOutputApart`, `TestCostRefusesUnmeasuredOutput` and `TestTokenLineRendersUnmeasuredOutput` — reverting the commit removes the `OutputUnmeasured` fields they read, and the package does not compile; `read output as absent when a bucket carries output_unmeasured` — `readOutputTokens` returns the undercount instead of null.
**Baseline:** before=105 after=109
<!-- measured: cat stats/internal/harvest/transcript_test.go stats/internal/harvest/attribute_test.go stats/internal/store/pricing_test.go stats/internal/records/render_test.go | grep -cE '^func Test' @ branch spectre/kan-920-flow-improvement-port-the-markdown-and-lint -->
**Commit:** fix(harvest): mark stream-start subagent output not measured
**After:** none

**Decision:** output-unmeasured
**Decision:** output-unmeasured-repair

- [ ] 5. Run the test list's commands concurrently
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `scripts/run-verify-list.sh`'s
    loop calls `run` for each command in turn and returns only when it exits; kan-920's run spent
    6:25 and 6:58 wall in its `## test` list.
    <!-- measured: the creating run's transcript timeline, a machine-local file, read 2026-10-21 → 19:49:02-19:55:27 and 19:59:24-20:06:22 -->
  - [ ] **Step 2: Apply the diff** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at 6a023063 after Tasks 1-4's diffs; scripts/test-run-verify-list.sh passes every case and fails both new ones on the serial script; two concurrent make -s web-dist with dist absent ran npm ci once and both exited 0
diff --git a/scripts/run-verify-list.sh b/scripts/run-verify-list.sh
index 3b7c0670..fb24dde3 100755
--- a/scripts/run-verify-list.sh
+++ b/scripts/run-verify-list.sh
@@ -17,6 +17,11 @@
 # so the list is refused before anything runs (kan-983 self-review). Each command runs from
 # <worktree> in a fresh `bash -c`, stdin detached (</dev/null) so none waits
 # on a terminal, and the whole list always runs: a failure never stops it.
+# A `test` list's commands run concurrently, each into its own output file,
+# reported in list order once every one has exited (kan-920 self-review: the
+# guard harnesses and `go test -race` ran one after the other, 6½ minutes; run
+# together they took 307 s, each slowed by the other's CPU). A `lint` list
+# runs in order, one command at a time.
 # The caller exports the KEY=value lines prepare-workspace.sh printed for
 # <worktree> first; every command inherits them. This script retries
 # nothing — the one inline re-run of a failing command stays the caller's.
@@ -36,7 +41,7 @@
 # ponytail: the sha ignores the exported KEY=value lines; a run under other
 # isolation values reuses the green, so key the file by them if they ever vary.
 #
-# OUTPUT, one block per command in the order run, stdout and stderr together:
+# OUTPUT, one block per command in list order, stdout and stderr together:
 #   - `<command>` — exit <n>
 #     <each output line, indented two spaces; past 20 lines,
 #     `(truncated — last 20 of <N> lines)` and then the last 20>
@@ -113,32 +118,48 @@ unchanged() {
   git -C "$WT" diff --quiet "$SINCE" "$HEAD_SHA" -- "${specs[@]}" .flow/project.md 2>/dev/null
 }
 
-OUT="$(mktemp "${TMPDIR:-/tmp}/run-verify-list.XXXXXX")"
-trap 'rm -f "$OUT"' EXIT
+OUTDIR="$(mktemp -d "${TMPDIR:-/tmp}/run-verify-list.XXXXXX")"
+trap 'rm -rf "$OUTDIR"' EXIT
 
 FAILED=0
-# run <command-as-shown> <command-to-run>
-run() {
-  local rc n
-  (cd "$WT" && bash -c "$2") </dev/null >"$OUT" 2>&1
-  rc=$?
-  [ "$rc" -eq 0 ] || FAILED=1
-  printf -- '- `%s` — exit %d\n' "$1" "$rc"
-  n="$(awk 'END{print NR}' "$OUT")"
+# report <command-as-shown> <exit> <output-file>
+report() {
+  local n
+  [ "$2" -eq 0 ] || FAILED=1
+  printf -- '- `%s` — exit %d\n' "$1" "$2"
+  n="$(awk 'END{print NR}' "$3")"
   if [ "$n" -gt 20 ]; then
     printf '  (truncated — last 20 of %d lines)\n' "$n"
-    tail -n 20 "$OUT" | awk '{print "  " $0}'
+    tail -n 20 "$3" | awk '{print "  " $0}'
   else
-    awk '{print "  " $0}' "$OUT"
+    awk '{print "  " $0}' "$3"
   fi
 }
+# run <command-as-shown> <command-to-run>, in the foreground
+run() {
+  (cd "$WT" && bash -c "$2") </dev/null >"$OUTDIR/out" 2>&1
+  report "$1" "$?" "$OUTDIR/out"
+}
 
+# PIDS[i] is the background pid of a `test` command, or `skip`.
+PIDS=()
 for i in "${!CMDS[@]}"; do
   if unchanged "${INPUTS[$i]}"; then
+    PIDS[i]=skip
+  elif [ "$LIST" = test ]; then
+    (cd "$WT" && bash -c "${CMDS[$i]}") </dev/null >"$OUTDIR/$i" 2>&1 &
+    PIDS[i]=$!
+  else
+    run "${CMDS[$i]}" "${CMDS[$i]}"
+  fi
+done
+for i in "${!PIDS[@]}"; do
+  if [ "${PIDS[$i]}" = skip ]; then
     printf -- '- `%s` — skipped, no input changed since %s passed the list\n' "${CMDS[$i]}" "$SINCE"
     continue
   fi
-  run "${CMDS[$i]}" "${CMDS[$i]}"
+  wait "${PIDS[$i]}"
+  report "${CMDS[$i]}" "$?" "$OUTDIR/$i"
 done
 if [ -n "$HEAD_SHA" ] && [ "$FAILED" -eq 0 ] &&
   STATUS="$(git -C "$WT" status --porcelain --untracked-files=all 2>/dev/null)" && [ -z "$STATUS" ]; then
diff --git a/scripts/test-run-verify-list.sh b/scripts/test-run-verify-list.sh
index 0cb19be7..596ab4f0 100755
--- a/scripts/test-run-verify-list.sh
+++ b/scripts/test-run-verify-list.sh
@@ -71,6 +71,14 @@ expect_has "test list reported" $'- `printf \'ok\\n\'` — exit 0\n  ok'
 expect_has "output without a final newline ends its line" $'  nonl\n- `echo next` — exit 0'
 expect_lacks "test never appends check-spec-reach.sh" 'check-spec-reach.sh'
 
+# The first command waits for a file only the second creates: run one after
+# the other it gives up after five seconds and fails the list.
+printf '## test\n\n```bash\nfor i in $(seq 50); do [ -f made-by-two ] && echo saw-two && exit 0; sleep 0.1; done; exit 1\ntouch made-by-two\n```\n' > "$ROOT/.flow/project.md"
+run_bin "$ROOT" test
+rm -f "$ROOT/made-by-two"
+expect_rc "test commands run concurrently" 0
+expect_has "concurrent output reported in list order" $'  saw-two\n- `touch made-by-two` — exit 0'
+
 printf '## lint\n\n```bash\ntrue\n```\n' > "$ROOT/.flow/project.md"
 run_bin "$ROOT" test
 expect_rc "an undeclared list" 3
diff --git a/stats/Makefile b/stats/Makefile
index 2c62ff43..20537de1 100644
--- a/stats/Makefile
+++ b/stats/Makefile
@@ -44,8 +44,17 @@ web-build:
 # (KAN-987).
 web-dist: internal/web/dist/index.html
 
+# The `## test` list's commands run concurrently (scripts/run-verify-list.sh),
+# and each opens with `make -s web-dist`, so the build takes a lock in the
+# git dir -- outside the tree, where the suites watch for changes -- and a
+# second caller waits for the first build instead of running `npm ci` over it.
+# ponytail: a build killed with SIGKILL leaves the lock and the next caller
+# waits on it forever; remove the directory it names if that ever happens.
 internal/web/dist/index.html:
-	$(MAKE) web-build
+	@lock="$$(git rev-parse --path-format=absolute --git-path web-build.lock 2>/dev/null)" || lock="$${TMPDIR:-/tmp}/flow-web-build.lock"; \
+	until mkdir "$$lock" 2>/dev/null; do sleep 1; done; \
+	trap 'rmdir "$$lock"' EXIT; \
+	[ -f $@ ] || $(MAKE) web-build
 
 # Both binaries: the Vite build, then `go build`, in that order -- never
 # the other way around -- so the embedded assets `go build` bakes in are
````

  - [ ] **Step 3: Verify** — `bash -n scripts/run-verify-list.sh` and `scripts/test-run-verify-list.sh` end `all cases passed`;
    `scripts/run-verify-list.sh . test` exits 0 with its three blocks in list order.

**Build:** green
**Files:** `scripts/run-verify-list.sh`, `scripts/test-run-verify-list.sh`, `stats/Makefile`
**Tests:** `test commands run concurrently`, `concurrent output reported in list order`
**Regression:** `test commands run concurrently` and `concurrent output reported in list order` — reverting the commit runs the first command alone, which gives up waiting for the file the second creates, and both cases fail with exit 1.
**Baseline:** before=32 after=34
<!-- measured: grep -cE '^expect_' scripts/test-run-verify-list.sh @ branch spectre/kan-920-flow-improvement-port-the-markdown-and-lint -->
**Commit:** perf(scripts): run the test list's commands concurrently
**After:** none

**Decision:** concurrent-test-list

- [ ] 6. Run the full suite while pass 1 reviews
  - [ ] **Step 1: Reproduce before touching anything; do not guess.**
    `grep -c 'The pass-1 panel reads while the suite runs' skills/flow/implement.md` prints `0`: the
    suite runs before `flow.sdd-tdd` closes and the panel is entered.
    <!-- measured: grep -c 'The pass-1 panel reads while the suite runs' skills/flow/implement.md @ branch spectre/kan-920-flow-improvement-port-the-markdown-and-lint → 0 -->
  - [ ] **Step 2: Apply the diff** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at 6a023063 after Tasks 1-5's diffs; check-verbatim-moves.sh, check-references.sh, check-markdown-integrity.py and check-dispatch-paragraphs.sh exit 0
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index 67457881..8f3f038d 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -673,6 +673,13 @@ committed PNGs are deleted and `capture` runs with `<spec>` substituted for its
 as PNGs that do not yet exist, and they are committed by the regression checkout's rules, step 12
 of `skills/flow/visual-verify.md`.
 
+**The pass-1 panel reads while the suite runs.** On a run whose `skills/flow/review-panel.md`
+opens a pass 1, the full-suite run above waits for that pass: the parent closes `flow.sdd-tdd`,
+enters the panel, and runs the list there, between recording the pass-1 begins and the round's
+wait, as that file's **Bundled dispatch** states. The slots are read-only, so their verdict and the
+suite's arrive together rather than one after the other (kan-920: the panel waited six and a half
+minutes for a suite it did not read).
+
 **The next implementer overlaps the guard.** The unit is the group — the decision's `groups`
 entry, one or more bundles `plan-dispatch-bundles.sh` emits. At each boundary, in this order:
 
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index 49bf108e..0764daf0 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -424,6 +424,13 @@ launched pass's report file; every `end` is recorded in one call once they all e
 (SDD + TDD)**, `skills/flow/implement.md`). A dispatch whose report never appears within its
 ceiling takes the breach path under **No forking, and a wall-clock ceiling on every slot** below.
 
+**On pass 1 the full suite runs between the begins and the wait.** The parent runs it there, in
+the foreground, while the read-only slots review (**The pass-1 panel reads while the suite
+runs.**, `skills/flow/implement.md`). A full-suite fix commit lands after the diff the slots
+read, so once the round's reports exist, every pass-1 role re-runs in the next round over the
+delta that commit adds, under **Panel re-runs** (`skills/flow/review-panel-fix-round.md`) — a
+role that raised nothing included; a suite that passes adds no commit and re-dispatches nothing.
+
 The same plan-tree discipline rides every round, re-run rounds included: before the launches the
 parent makes the reviewer-dispatch planning commit (**Planning commits**,
 `skills/flow-contracts/git-boundaries.md`) and runs
````

  - [ ] **Step 3: Verify** — `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py` and `scripts/check-dispatch-paragraphs.sh` exit 0.

**Build:** green
**Files:** `skills/flow/implement.md`, `skills/flow/review-panel.md`
**Tests:** none — prose; the lint guards are the check
**Regression:** none — prose only
**Commit:** perf(flow): run the full suite while pass 1 reviews
**After:** Task 2

**Decision:** panel-during-suite
