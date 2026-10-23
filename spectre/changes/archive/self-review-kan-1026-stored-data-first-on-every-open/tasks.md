# self-review-kan-1026-stored-data-first-on-every-open

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** fix the seven non-`big` findings of `kan-1026-stored-data-first-on-every-open`'s self-review pass and repair the ledger
rows the stamping defect already wrote wrong.

**Architecture:** Task 1 enforces keyed launch descriptions in `hooks/enforce-agent-baseline.py`,
anchors a lone begin at its row in `stats/internal/harvest/watcher.go` (new store read), and
restamps the kan-1026 rows; Task 2 adds the `panel-round-begins` flow-guard command; Task 3 closes
a stamped row at its agent transcript's last write (flowd's end handler); Task 4 is visual-verify
prose; Task 5 narrows `hooks/protect-main-checkout.py`'s heredoc scan; Task 6 moves the `-tokens`
refusal into the store; Task 7 elides raw measurement sections from the self-review bundle.

**Tech Stack:** Go (`stats/`), Postgres via `pgx`, Python hooks, Markdown skills.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-23-self-review-kan-1026-stored-data-first-on-every-open-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer.
  Never write `spectre/changes/` — this change's `verbatim-moves.txt` already lists every prose
  sentence the diffs add or reword, exactly as they spell them. Never merge or push.
- Every task runs in this worktree, `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1026-stored-data-first-on-every-open`, on branch `self-review-kan-1026-stored-data-first-on-every-open`; every `**Files:**` path is relative to
  its root. Never touch a main checkout, and never create a worktree.
- Each task's diff applies with `git apply` from the worktree root onto the previous task's
  commit; all seven were applied in order and checked at `c1282bc4`. A diff that does not apply is
  reported, never hand-merged by guesswork. Each prose sentence lands character for character:
  `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly.
- Run `cd stats && gofmt -w .` before committing any Go change; `gofmt -l .` and `make -s web-dist
  && go vet ./...` stay clean.
- Task 1's repair step writes the live dev flow store (`flow-postgres`, `localhost:5433`) — local,
  not production, and allowed for this change. It runs once, after Task 1's commit; a command that
  fails is reported with its output, never retried blind. Never start, stop or restart the
  protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.
- Every Go test command below runs with the Bash tool's longest foreground timeout; `internal/store`'s tests need
  the `flow-postgres` stack up, and a store test that reports `SKIP` is a failed verify, not a pass.
- The installed `protect-main-checkout.py` is HEAD's until this branch lands: a heredoc command it
  denies over its body (the Task 5 defect) is rewritten as a file written with the Write tool, never
  pushed past.

## Review Focus

- Task 1: a lone begin keeps the command's instant until its row exists, and begins sharing one
  Bash call never read the store — the shared instant is what pairs a round in order.
- Task 1: the description gate only fires on a transcript carrying a flow slash command, and never
  on a loop dispatch carrying its `Key:` line.
- Task 5: a process substitution, a backtick, a nested `$(`, a shell runner, an env-assignment
  prefix on a non-interpreter, or the target run as a command still keeps a body scanned.
- Task 6: the two kan-525 fixtures move to `muse` and nothing else in them changes.

- [x] 1. Stamp every flow launch by its row's key, anchor a lone begin at its row, and repair kan-1026's rows
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -o -- '-key panel-\$R[^ ]*' /Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/50d27d94-f31d-4308-90ff-eef33cdbaa77.jsonl | sort | uniq -c`: it prints the substituted keys
    (`panel-$R-$s`, `panel-$R-primary`, `panel-$R-primary+test-sanity`). Run
    `flow record dispatches -C /Users/tweety53/Projects/gymie -change kan-1026-stored-data-first-on-every-open`: seqs 37, 38, 52, 57, 58, 61, 62, 64, 68, 69, 73 and
    74 carry no `agentId`; seq 53 carries `a18dd75d85c0d7828` and seq 76 `abe0a67c405742bad`.
    <!-- measured: the grep above @ machine-local transcript, 2026-10-23 → 16 substituted-key begins -->
  - [x] **Step 2: Apply the fix.** Save the diff below by copying the lines between its fence markers into a scratch file outside the worktree, then `git apply <that file>` from the worktree root.

````diff verified:applied with git apply in order with the other tasks' diffs and prose replacements onto c1282bc4 in this worktree; the result equals the spiked tree, go vet ./... and gofmt -l clean, go test of internal/harvest, internal/store, internal/api, internal/guard, internal/repotest, internal/selfreview, internal/reconcile, cmd/flow and cmd/flowd passing, and check-verbatim-moves.sh, check-references.sh, check-guard-symlinks.sh, check-markdown-integrity.sh and check-installed-citations.sh exiting 0
diff --git a/hooks/enforce-agent-baseline.py b/hooks/enforce-agent-baseline.py
index 46b24862..35bb2e9b 100755
--- a/hooks/enforce-agent-baseline.py
+++ b/hooks/enforce-agent-baseline.py
@@ -22,6 +22,13 @@ It denies every fork in a flow session -- one whose transcript records a /flow,
 fresh agent whose prompt carries the paths it needs (KAN-1000: a parent forked itself into a
 self-review loop and dragged ~300k tokens of its own context into it).
 
+It denies, in a flow session, a dispatch whose `description` does not open with its row: `Task
+<x>/<n>` or `Tasks <a+b>/<n>` for plan tasks, else the row's literal `-key` (**Every dispatch's
+Agent-tool `description`**, skills/flow/implement.md) -- the word the harvester stamps the launch
+by. A loop dispatch, whose prompt carries a `Key: self-review-`/`Key: pipeline-fix-` line, is
+stamped by that line and exempt (kan-1026: 29 of 41 launches opened "Round-11 primary review" and
+the like, and 14 ledger rows were left empty or stamped with another dispatch's agent).
+
 Fails open: any unexpected input, parse error, or internal fault exits 0 and lets the
 dispatch through. A guardrail on a convention must never be the reason work stops.
 """
@@ -95,6 +102,41 @@ FLOW_HIGH_MARKERS = (
 )
 
 
+# A description naming its row: plan tasks by id over the plan's count, any other row by a key-shaped
+# first word -- lowercase, no spaces (`panel-4-primary`, `visual-verify-fix-11-gymie`, `planner`).
+ROW_DESCRIPTION = re.compile(r"Tasks? \d+(?:\+\d+)*/\d+(?:\s|$)|[a-z][a-z0-9+._-]*(?:\s|$)")
+LOOP_KEY_LINE = re.compile(r"^Key: (?:self-review-|pipeline-fix-)", re.M)
+
+DESCRIPTION_REASON = (
+    "Dispatch blocked: in a flow run, a dispatch's `description` opens with the row it records "
+    "-- `Task <x>/<n> (...)` for plan tasks, else the row's literal `-key` and a few words, "
+    "`panel-4-primary correctness review` (skills/flow/implement.md, kan-1026). The daemon stamps "
+    "the launch onto the row that word names; a description like \"Round-11 primary review\" "
+    "leaves the row to time pairing, which mis-stamps.\n\n"
+    "Re-send this call with `description` opening with its row's key, nothing else changed."
+)
+
+
+def description_reason(payload: dict, tool_input: dict, prompt: str) -> "str | None":
+    if LOOP_KEY_LINE.search(prompt):
+        return None
+    description = tool_input.get("description")
+    if isinstance(description, str) and ROW_DESCRIPTION.match(description):
+        return None
+    return DESCRIPTION_REASON if in_flow_session(payload) else None
+
+
+def in_flow_session(payload: dict) -> bool:
+    path = payload.get("transcript_path")
+    if not isinstance(path, str):
+        return False
+    try:
+        with open(path, encoding="utf-8", errors="replace") as f:
+            return any(is_flow_command(line) for line in f)
+    except OSError:
+        return False
+
+
 def flow_high_reason(prompt: str, tool_input: dict) -> "str | None":
     if tool_input.get("subagent_type") != "flow-high" or "/.zcode/" in __file__:
         return None
@@ -133,21 +175,15 @@ def fork_reason(payload: dict, tool_input: dict) -> "str | None":
     prompt = tool_input.get("prompt")
     if isinstance(prompt, str) and "skills/flow" in prompt:
         return FORK_REASON
-    path = payload.get("transcript_path")
-    if not isinstance(path, str):
-        return None
-    try:
-        with open(path, encoding="utf-8", errors="replace") as f:
-            if any(is_flow_command(line) for line in f):
-                return FORK_REASON
-    except OSError:
-        return None
-    return None
+    return FORK_REASON if in_flow_session(payload) else None
 
 
 def is_flow_command(line: str) -> bool:
     # Only a user turn with string content is a typed slash command; a tool result quoting the
-    # tag carries list content and does not count.
+    # tag carries list content and does not count. The substring test spares a json parse of
+    # every line of a long transcript that holds no flow command at all.
+    if "<command-name>/flow" not in line:
+        return False
     try:
         entry = json.loads(line)
     except ValueError:
@@ -203,7 +239,11 @@ def main() -> int:
         deny(DENY_REASON)
         return 0
 
-    reason = pinned_role_reason(prompt, tool_input) or flow_high_reason(prompt, tool_input)
+    reason = (
+        pinned_role_reason(prompt, tool_input)
+        or flow_high_reason(prompt, tool_input)
+        or description_reason(payload, tool_input, prompt)
+    )
     if reason:
         deny(reason)
     return 0
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index 4a5d0d42..7acb7c8d 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -88,7 +88,10 @@ shape.** A dispatch carrying plan tasks — an implementer, a gated reviewer bun
 by plan id over the plan's task count, then a few words in parentheses: `Task 3/22 (port guard)`
 for one task, `Tasks 3+4+7/22 (review)` for a group, its ids in the order its key lists them. Every other row is its key and a few
 words, with no parentheses: `panel-1-primary correctness review`. The `flow-task-list` mod
-(`mods/flow-task-list/`) reads that prefix to number its row.
+(`mods/flow-task-list/`) reads that prefix to number its row. The daemon stamps a launch onto the
+row its description's first word names, so `hooks/enforce-agent-baseline.py` denies, in a flow run,
+a dispatch whose description opens any other way (kan-1026: launches described `Round-11 primary
+review` left their rows to time pairing, which stamped two with another dispatch's agent).
 
 **Every dispatch is one-shot — a finished child is never resumed.** The parent never sends a
 `SendMessage` to a child that has written its report. Whatever a
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index c68cc81b..7783e18c 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -466,7 +466,7 @@ instead — the one pair the store accepts on a `zcode` session token — and on
 the session's own (**Harness mapping**, `skills/flow-contracts/model-policy.md`; KAN-898).
 
 **`-agent-id` is never typed on a slot's row, never invented** — the daemon captures the launch identifier,
-stamping each launch onto the row its Agent-tool `description`'s first word names in the launching session, and pairing a launch whose description names no row with the begin nearest it in time;
+stamping each launch onto the row its Agent-tool `description`'s first word names in the launching session — each slot's description opens with its row's `-key` (**Dispatch sites — the parent's closed list**, `skills/flow/implement.md`) — and pairing a launch whose description names no row with the begin nearest it in time;
 record the round's `begin`s in the one Bash call after its launch message, as above, and never reuse a `-key`.
 
 **Record a slot's dispatch before recording that slot's findings**, and carry the seq the command
diff --git a/stats/internal/harvest/deps.go b/stats/internal/harvest/deps.go
index 50f91a48..ff4bad50 100644
--- a/stats/internal/harvest/deps.go
+++ b/stats/internal/harvest/deps.go
@@ -28,6 +28,11 @@ type DispatchAgentStamper interface {
 	// recorded under key in session sessionID -- the row a launch's
 	// description names (AgentLaunch.Key).
 	StampSessionDispatchAgent(ctx context.Context, sessionID, key, agentID string) (stamped bool, err error)
+	// DispatchStartedAt reads the started_at of the latest row recorded
+	// under (sessionToken, key) -- the instant its begin actually ran,
+	// which Watcher.anchorBegins pairs a lone begin by; found is false
+	// while no such row exists yet.
+	DispatchStartedAt(ctx context.Context, sessionToken, key string) (startedAt time.Time, found bool, err error)
 }
 
 // Deps is everything a Watcher needs beyond its root, its sink and its
@@ -143,6 +148,11 @@ func (NoDeps) StampSessionDispatchAgent(ctx context.Context, sessionID, key, age
 	return false, nil
 }
 
+// DispatchStartedAt never finds a row.
+func (NoDeps) DispatchStartedAt(ctx context.Context, sessionToken, key string) (startedAt time.Time, found bool, err error) {
+	return time.Time{}, false, nil
+}
+
 // InlineDispatchID never finds a row.
 func (NoDeps) InlineDispatchID(ctx context.Context, sessionToken, key string) (id int64, found bool, err error) {
 	return 0, false, nil
diff --git a/stats/internal/harvest/watcher.go b/stats/internal/harvest/watcher.go
index 8b5387b6..e6a18089 100644
--- a/stats/internal/harvest/watcher.go
+++ b/stats/internal/harvest/watcher.go
@@ -1155,6 +1155,11 @@ type dispatchBeginEvent struct {
 	SessionToken string
 	Key          string
 	StartedAt    time.Time
+	// Alone marks the only begin its Bash command recorded, and Anchored
+	// a begin whose StartedAt is already its row's own started_at
+	// (Watcher.anchorBegins).
+	Alone    bool
+	Anchored bool
 }
 
 // maxStampTimeTolerance is how far a begin's own -started-at may sit from
@@ -1775,17 +1780,20 @@ func (w *Watcher) stampDispatchAgents(ctx context.Context, path string, commands
 
 	var begins []dispatchBeginEvent
 	for _, c := range commands {
-		for _, ev := range dispatchBeginsFromCommand(c.Command) {
+		evs := dispatchBeginsFromCommand(c.Command)
+		for _, ev := range evs {
 			// The command line's own instant is the begin's time anchor; a
 			// command whose line carries none can never pair.
 			if !c.Timestamp.IsZero() {
 				ev.StartedAt = c.Timestamp
+				ev.Alone = len(evs) == 1
 				begins = append(begins, ev)
 			}
 		}
 	}
 	launches = w.stampLoopLaunches(ctx, path, launches)
 	begins = append(w.pendingBegins[path], begins...)
+	w.anchorBegins(ctx, path, begins)
 	launches = append(w.pendingLaunches[path], launches...)
 	denials = append(w.pendingDenials[path], denials...)
 	launches, begins = w.stampKeyedLaunches(ctx, path, launches, begins)
@@ -1915,6 +1923,34 @@ func (w *Watcher) stampDispatchAgents(ctx context.Context, path string, commands
 	w.pendingDenials[path] = appendPending(w.pendingDenials[path], denials)
 }
 
+// anchorBegins moves each lone begin's time anchor from its Bash command's
+// transcript instant to its row's own started_at. The command line is
+// written when the call starts, and a begin at the end of a long call ran
+// minutes later: kan-1026's row 76, begun 2m56s into its call, sat 42 s
+// from another dispatch's launch made before the call and 3m14s from its
+// own, and took the other one. Begins sharing one command keep its
+// instant: that shared instant is what pairs a panel round's begins with
+// its launches in order. A begin whose row is not written yet keeps the
+// command's instant -- its stamp then finds no row, and the pair is
+// retried -- and is looked up again next batch.
+func (w *Watcher) anchorBegins(ctx context.Context, path string, begins []dispatchBeginEvent) {
+	for i := range begins {
+		b := &begins[i]
+		if !b.Alone || b.Anchored {
+			continue
+		}
+		at, found, err := w.deps.DispatchStartedAt(ctx, b.SessionToken, b.Key)
+		if err != nil {
+			w.warn("harvest: read dispatch started_at failed, the begin keeps its command's instant",
+				"path", path, "key", b.Key, "error", err)
+			continue
+		}
+		if found {
+			b.StartedAt, b.Anchored = at, true
+		}
+	}
+}
+
 // stampKeyedLaunches stamps each launch whose description's first word
 // (AgentLaunch.Key) names an empty row in its session, and returns the
 // launches it did not stamp and the begins no keyed stamp used. A begin
diff --git a/stats/internal/harvest/watcher_test.go b/stats/internal/harvest/watcher_test.go
index 9781bdd8..7e3abb92 100644
--- a/stats/internal/harvest/watcher_test.go
+++ b/stats/internal/harvest/watcher_test.go
@@ -4237,6 +4237,55 @@ func TestLateStampLeavesAResumedAgentsFileAlone(t *testing.T) {
 // launch, the panel's real launches, and the flake's own begin. The
 // denied lines must decode, so both panel begins retire and only the
 // flake begin stamps -- its own launch.
+// anchoredStampDeps is stampRecordingDeps whose store answers each key's
+// row started_at from startedAt.
+type anchoredStampDeps struct {
+	stampRecordingDeps
+	startedAt map[string]time.Time
+}
+
+func (d *anchoredStampDeps) DispatchStartedAt(ctx context.Context, sessionToken, key string) (time.Time, bool, error) {
+	at, ok := d.startedAt[key]
+	return at, ok, nil
+}
+
+// TestWatcherAnchorsALoneBeginAtItsRowsStartedAt is kan-1026's row 76: the
+// begin closed a three-minute Bash call, so the call's transcript instant
+// sat 42 s from another dispatch's launch made before the call and 3m14s
+// from its own. Anchored at the row's started_at, the begin takes its own.
+func TestWatcherAnchorsALoneBeginAtItsRowsStartedAt(t *testing.T) {
+	bash := func(ts, command string) string {
+		c, err := json.Marshal(command)
+		if err != nil {
+			t.Fatalf("marshal command: %v", err)
+		}
+		return `{"type":"assistant","timestamp":"2026-10-23T` + ts + `","sessionId":"s","message":{"model":"m","usage":{"input_tokens":1},"content":[{"type":"tool_use","name":"Bash","input":{"command":` + string(c) + `}}]}}`
+	}
+	launch := func(ts, agentID string) string {
+		return `{"type":"user","timestamp":"2026-10-23T` + ts + `","sessionId":"s","toolUseResult":{"isAsync":true,"status":"async_launched","agentId":"` + agentID + `"}}`
+	}
+	dir := t.TempDir()
+	path := filepath.Join(dir, "session.jsonl")
+	if err := os.WriteFile(path, stampFixtureLines(
+		launch("09:36:33.893Z", "abe0a67c405742bad"),
+		bash("09:37:15.000Z", "npx playwright test >/dev/null; flow record dispatch begin -change kan-1026 -role verifier -key visual-verify-fix-11-gymie -session-token mf-k1026r1"),
+		launch("09:40:29.885Z", "a549fee896ec5faec"),
+	), 0o644); err != nil {
+		t.Fatalf("write fixture: %v", err)
+	}
+	deps := &anchoredStampDeps{startedAt: map[string]time.Time{
+		"visual-verify-fix-11-gymie": mustParse(t, "2026-10-23T09:40:11Z"),
+	}}
+	w := harvest.NewWatcher([]harvest.Source{harvest.NewClaudeSource(dir)}, newFakeHarvestSink(), harvest.NewAttributor(&fakeWindowSource{}), deps, nil)
+	if _, err := w.RunOnce(context.Background()); err != nil {
+		t.Fatalf("RunOnce: %v", err)
+	}
+	want := []stampCall{{"mf-k1026r1", "visual-verify-fix-11-gymie", "a549fee896ec5faec"}}
+	if !reflect.DeepEqual(deps.stamps, want) {
+		t.Errorf("stamps = %+v, want %+v", deps.stamps, want)
+	}
+}
+
 func TestWatcherRetiresBeginsOfHookDeniedLaunches(t *testing.T) {
 	bash := func(ts, command string) string {
 		c, err := json.Marshal(command)
diff --git a/stats/internal/repotest/enforce_agent_baseline_hook_test.go b/stats/internal/repotest/enforce_agent_baseline_hook_test.go
index 658e4e0a..1a8f6867 100644
--- a/stats/internal/repotest/enforce_agent_baseline_hook_test.go
+++ b/stats/internal/repotest/enforce_agent_baseline_hook_test.go
@@ -68,9 +68,13 @@ func TestEnforceAgentBaselineHook(t *testing.T) {
 	statusT := transcript("status.jsonl", `{"type":"user","message":{"content":"<command-name>/flow-status</command-name>"}}`)
 	toolT := transcript("tool.jsonl", `{"type":"user","message":{"content":[{"type":"tool_result","content":"<command-name>/flow</command-name>"}]}}`)
 
-	// payload builds an Agent payload; model "" omits it, transcript "" omits transcript_path.
-	payload := func(prompt, typ, model, transcript string) string {
+	// described builds an Agent payload; description "" and model "" omit those fields,
+	// transcript "" omits transcript_path.
+	described := func(description, prompt, typ, model, transcript string) string {
 		ti := map[string]string{"prompt": prompt, "subagent_type": typ}
+		if description != "" {
+			ti["description"] = description
+		}
 		if model != "" {
 			ti["model"] = model
 		}
@@ -84,6 +88,9 @@ func TestEnforceAgentBaselineHook(t *testing.T) {
 		}
 		return string(b)
 	}
+	payload := func(prompt, typ, model, transcript string) string {
+		return described("", prompt, typ, model, transcript)
+	}
 
 	const hook = "enforce-agent-baseline.py"
 	cases := []struct {
@@ -112,6 +119,12 @@ func TestEnforceAgentBaselineHook(t *testing.T) {
 		{"11 malformed JSON", hook, "{not json", nil},
 		{"12 a planner citing the verifier file, no marker line", hook, payload("Key: self-review-x-plan\n"+baseline+" Finding 2 cites /x/skills/flow/visual-verify-verifier.md:120.", "flow-medium", "opus", ""), nil},
 		{"12b an implementer citing both files, no marker line", hook, payload(baseline+" Task 4 re-reads .superpowers/sdd/tooling-analysis-1.md and step 8 of skills/flow/visual-verify-verifier.md.", "flow-low", "sonnet", ""), nil},
+		{"13 a /flow dispatch described in prose", hook, described("Round-11 primary review", baseline+" review the diff", "flow-low", "opus", flowT), []string{"opens with the row it records"}},
+		{"13b a /flow dispatch with no description", hook, payload(baseline+" review the diff", "flow-low", "opus", flowT), []string{"opens with the row it records"}},
+		{"13c a /flow dispatch opening with its key", hook, described("panel-11-primary correctness review", baseline+" review the diff", "flow-low", "opus", flowT), nil},
+		{"13d a /flow task dispatch", hook, described("Tasks 3+4+7/22 (review)", baseline+" review tasks 3, 4 and 7", "flow-low", "opus", flowT), nil},
+		{"13e a /flow loop dispatch carrying its Key line", hook, described("Self-review plan", "Key: self-review-x-plan\n"+baseline+" plan", "flow-medium", "opus", flowT), nil},
+		{"13f a prose description outside a flow session", hook, described("Round-11 primary review", baseline+" review the diff", "flow-low", "opus", statusT), nil},
 	}
 	for _, c := range cases {
 		t.Run(c.name, func(t *testing.T) {
diff --git a/stats/internal/store/records.go b/stats/internal/store/records.go
index dac7ce3b..522f2b87 100644
--- a/stats/internal/store/records.go
+++ b/stats/internal/store/records.go
@@ -1538,6 +1538,26 @@ func (s *Store) StampDispatchAgent(ctx context.Context, sessionToken, key, agent
 	return tag.RowsAffected() > 0, nil
 }
 
+// DispatchStartedAt reads the started_at of the latest row recorded under
+// (sessionToken, key) -- harvest.DispatchAgentStamper's anchor for a lone
+// begin. found is false, with a nil error, while no such row exists.
+func (s *Store) DispatchStartedAt(ctx context.Context, sessionToken, key string) (time.Time, bool, error) {
+	var at time.Time
+	err := s.pool.QueryRow(ctx, `
+		SELECT started_at FROM dispatches
+		WHERE session_token = $1 AND dispatch_key = $2
+		ORDER BY started_at DESC, id DESC
+		LIMIT 1
+	`, sessionToken, key).Scan(&at)
+	if errors.Is(err, pgx.ErrNoRows) {
+		return time.Time{}, false, nil
+	}
+	if err != nil {
+		return time.Time{}, false, fmt.Errorf("store: dispatch started_at for key %q: %w", key, err)
+	}
+	return at, true, nil
+}
+
 // StampLoopDispatchAgent is StampDispatchAgent for a fix-loop row, named
 // by its key alone (harvest.AgentLaunch.LoopKey): the loop's begin is
 // recorded by the PreToolUse hook, so the transcript carries no session
diff --git a/stats/internal/store/records_test.go b/stats/internal/store/records_test.go
index 0b35d5c5..1b3cdee4 100644
--- a/stats/internal/store/records_test.go
+++ b/stats/internal/store/records_test.go
@@ -3328,6 +3328,28 @@ func TestStampLoopDispatchAgentFillsLatestRowByKey(t *testing.T) {
 // keyed stamp: a launch whose description names its row stamps the latest
 // empty row under that key in the launching session only -- never another
 // session's row, never a row already carrying an agent.
+func TestDispatchStartedAtReadsTheLatestRowsBegin(t *testing.T) {
+	st, _ := newRecordStore(t)
+	ctx := context.Background()
+
+	projectKey := fmt.Sprintf("proj-started-at-%d", time.Now().UnixNano())
+	seedChange(t, st, projectKey, "kan-1")
+	in := baseDispatch("verifier", "opus")
+	in.SessionToken, in.Key = "mf-k1", "visual-verify-fix-11-gymie"
+	in.StartedAt = time.Date(2026, 10, 23, 9, 40, 11, 0, time.UTC)
+	if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", in); err != nil {
+		t.Fatalf("RecordDispatch: %v", err)
+	}
+
+	at, found, err := st.DispatchStartedAt(ctx, "mf-k1", "visual-verify-fix-11-gymie")
+	if err != nil || !found || !at.Equal(in.StartedAt) {
+		t.Errorf("DispatchStartedAt = (%v, %v, %v), want (%v, true, nil)", at, found, err, in.StartedAt)
+	}
+	if _, found, err := st.DispatchStartedAt(ctx, "mf-k1", "panel-1-primary"); err != nil || found {
+		t.Errorf("an unknown key = (found %v, %v), want (false, nil)", found, err)
+	}
+}
+
 func TestStampSessionDispatchAgentFillsTheSessionsLatestEmptyRow(t *testing.T) {
 	st, pool := newRecordStore(t)
 	ctx := context.Background()
````

  - [x] **Step 3: Prove the anchor.** Comment out the `w.anchorBegins(ctx, path, begins)` line in
    `stats/internal/harvest/watcher.go`, run
    `cd stats && go test ./internal/harvest/ -run 'TestWatcherAnchorsALoneBeginAtItsRowsStartedAt' -count=1`:
    it fails with `stamps = [{... agentID:abe0a67c405742bad}]`. Restore the line.
    <!-- measured: the same mutation at plan time @ branch self-review-kan-1026-stored-data-first-on-every-open → FAIL stamping abe0a67c405742bad -->
  - [x] **Step 4: Prove the description gate.** With `hooks/enforce-agent-baseline.py` at HEAD's
    content (`git show HEAD:hooks/enforce-agent-baseline.py` written over it), cases
    `13 a /flow dispatch described in prose` and `13b a /flow dispatch with no description` of
    `TestEnforceAgentBaselineHook` fail; restore the patched hook.
    <!-- measured: cd stats && go test ./internal/repotest/ -run TestEnforceAgentBaselineHook with the hook at c1282bc4 @ branch self-review-kan-1026-stored-data-first-on-every-open → 13 and 13b FAIL -->
  - [x] **Step 5: Verify.** `cd stats && gofmt -w . && gofmt -l . && make -s web-dist && go vet ./...`
    is clean; `cd stats && go test ./internal/harvest/ ./internal/store/ ./internal/repotest/ -run
    'TestWatcherAnchorsALoneBeginAtItsRowsStartedAt|TestWatcherStamps|TestWatcherPairs|TestWatcherRetires|TestDispatchStartedAtReadsTheLatestRowsBegin|TestEnforceAgentBaselineHook' -count=1`
    passes; `cd stats && go build ./cmd/flowd` succeeds (`*store.Store` still satisfies
    `harvest.Deps`); from the worktree root `scripts/check-verbatim-moves.sh`,
    `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.
  - [x] **Step 6: Repair the store.** Build the binary once and restamp each row onto its true
    agent, each pair `<dispatch id>:<agent id>`:

```bash verified:every agent file below exists under the parent transcript's subagents/ directory, and each agent's launch description in the parent transcript names that row's round and slot (dispatch ids read from flow record dispatches, 2026-10-23)
cd /Users/tweety53/Projects/agents-worktrees/self-review-kan-1026-stored-data-first-on-every-open/stats && go build -o /Users/tweety53/Projects/agents-worktrees/self-review-kan-1026-stored-data-first-on-every-open/.superpowers/sdd/flowd-restamp ./cmd/flowd
A=/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/50d27d94-f31d-4308-90ff-eef33cdbaa77/subagents
for pair in 7210:a37fffabbbbae9907 7211:a2764487139412782 7232:a18dd75d85c0d7828 \
  7233:ae17fcffb06f55709 7237:af3a362078da4cf37 7238:a2f8a2b11314cc1b8 \
  7241:a110acc9991987109 7242:aeabb9d3b463e9eb6 7244:ab7134fb446cfa8f0 \
  7251:a2c90a50df4f939be 7252:ae06cbf9d823c3e20 7263:abe0a67c405742bad \
  7264:a0c3edb1c3b45da93 7266:a549fee896ec5faec; do
  /Users/tweety53/Projects/agents-worktrees/self-review-kan-1026-stored-data-first-on-every-open/.superpowers/sdd/flowd-restamp restamp-dispatch "${pair%%:*}" "$A/agent-${pair#*:}.jsonl" || echo "FAILED ${pair%%:*}"
done
```

    Then `flow record dispatches -C /Users/tweety53/Projects/gymie -change kan-1026-stored-data-first-on-every-open` shows each of those 14 ids carrying its pair's
    agent and a `tokens` bag, and `flow record cost-status -C /Users/tweety53/Projects/gymie -change kan-1026-stored-data-first-on-every-open` no longer prints 15
    unattributed. Report any `FAILED` line with its output, and both commands' output.
    <!-- measured: flow record cost-status -change kan-1026-stored-data-first-on-every-open @ live store, 2026-10-23 → 15 unattributed — not measured: 15 -->
    <!-- predicted: 3 unattributed after this step, the three inline verifier rows -->

**Build:** green
**Files:** `hooks/enforce-agent-baseline.py`, `skills/flow/implement.md`, `skills/flow/review-panel.md`, `stats/internal/harvest/deps.go`, `stats/internal/harvest/watcher.go`, `stats/internal/harvest/watcher_test.go`, `stats/internal/repotest/enforce_agent_baseline_hook_test.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`
**Tests:** `TestWatcherAnchorsALoneBeginAtItsRowsStartedAt`, `TestDispatchStartedAtReadsTheLatestRowsBegin`, `13 a /flow dispatch described in prose`, `13b a /flow dispatch with no description`, `13c a /flow dispatch opening with its key`, `13d a /flow task dispatch`, `13e a /flow loop dispatch carrying its Key line`, `13f a prose description outside a flow session`
**Regression:** reverting `anchorBegins` stamps row 76's orphan launch, failing `TestWatcherAnchorsALoneBeginAtItsRowsStartedAt`; reverting the store method fails `TestDispatchStartedAtReadsTheLatestRowsBegin` to compile; reverting the hook's `description_reason` lets cases 13 and 13b through, failing them; 13c–13f fail if the gate over-reaches onto a keyed, task, loop or non-flow dispatch.
**Baseline:** before=187 after=195
<!-- measured: cat stats/internal/harvest/watcher_test.go stats/internal/store/records_test.go stats/internal/repotest/enforce_agent_baseline_hook_test.go | grep -cE '^func Test|^		\{"[0-9]+[a-z]? ' @ c1282bc4 and branch self-review-kan-1026-stored-data-first-on-every-open -->
**Commit:** fix(harvest): stamp every flow launch by its row's key and anchor a lone begin at its row
**After:** none

**Decision:** enforce-launch-description-key
**Decision:** anchor-lone-begin-at-row
**Decision:** pairing-order-unchanged
**Decision:** gymie-ledger-refresh-by-parent

- [x] 2. Record a panel round's begins under literal keys
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `ls scripts/panel-round-begins.sh`
    fails (no such file), and Task 1 Step 1's grep shows the parent's hand-built `-key panel-$R-$s`
    loop.
  - [x] **Step 2: Apply the fix.** Save the diff below by copying the lines between its fence markers into a scratch file outside the worktree, then `git apply <that file>` from the worktree root.

````diff verified:applied with git apply in order with the other tasks' diffs and prose replacements onto c1282bc4 in this worktree; the result equals the spiked tree, go vet ./... and gofmt -l clean, go test of internal/harvest, internal/store, internal/api, internal/guard, internal/repotest, internal/selfreview, internal/reconcile, cmd/flow and cmd/flowd passing, and check-verbatim-moves.sh, check-references.sh, check-guard-symlinks.sh, check-markdown-integrity.sh and check-installed-citations.sh exiting 0
diff --git a/scripts/panel-round-begins.sh b/scripts/panel-round-begins.sh
new file mode 100755
index 00000000..785c018c
--- /dev/null
+++ b/scripts/panel-round-begins.sh
@@ -0,0 +1,23 @@
+#!/usr/bin/env bash
+# panel-round-begins.sh <worktree> <change> <round> <session-token> <slot>:<model>:<effort>[:<diff-base>]...
+# — records every slot's begin of one review-panel round, each under the
+# literal key panel-<round>-<slot>, in the one Bash call after the round's
+# launch message (skills/flow/review-panel.md, "Every slot's dispatch is
+# recorded"). <round> is the number itself, never a substitution.
+#
+# stdout: one `<key>: <flow's answer>` line per slot, in slot order.
+#
+# Exit codes: 0 every begin recorded; 1 a begin failed (its output relayed,
+# the later slots still recorded); 2 usage, or no flow binary on PATH.
+#
+# Go, in flow-guard (stats/internal/guard/panelroundbegins.go). flow-guard
+# is built from this checkout, never taken from PATH: scripts/lib/flow-guard.sh
+# derives it, and exits 2 with the cause when it cannot.
+set -euo pipefail
+# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
+SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
+[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
+  echo "panel-round-begins: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
+  exit 2
+}
+flow_guard_exec panel-round-begins 2 "panel-round-begins:" "$@"
diff --git a/stats/internal/guard/panel_round_begins_test.go b/stats/internal/guard/panel_round_begins_test.go
new file mode 100644
index 00000000..3960ae74
--- /dev/null
+++ b/stats/internal/guard/panel_round_begins_test.go
@@ -0,0 +1,59 @@
+package guard
+
+import (
+	"bytes"
+	"os"
+	"path/filepath"
+	"strings"
+	"testing"
+)
+
+// TestPanelRoundBeginsRecordsLiteralKeys pins kan-1026's self-review: each
+// slot's begin carries panel-<round>-<slot> as a literal key, in slot
+// order, with its own model, effort and diff base.
+func TestPanelRoundBeginsRecordsLiteralKeys(t *testing.T) {
+	dir := t.TempDir()
+	log := filepath.Join(dir, "calls")
+	stub := "#!/bin/sh\necho \"$*\" >> " + log + "\ncase \"$*\" in *-key\\ panel-4-bad*) echo refused; exit 1;; esac\necho \"recorded: dispatch 7\"\n"
+	if err := os.WriteFile(filepath.Join(dir, "flow"), []byte(stub), 0o755); err != nil {
+		t.Fatal(err)
+	}
+	env := Env{Getenv: func(k string) string {
+		if k == "PATH" {
+			return dir
+		}
+		return ""
+	}}
+
+	var out, errb bytes.Buffer
+	code := panelRoundBegins([]string{"/wt", "kan-1", "4", "mf-k1", "primary:opus:low:a515c200", "test-sanity+call-efficiency:sonnet:low"}, env, &out, &errb)
+	if code != 0 {
+		t.Fatalf("exit %d, stderr %s", code, errb.String())
+	}
+	calls, _ := os.ReadFile(log)
+	want := "record dispatch begin -C /wt -change kan-1 -role reviewer -slot primary -model opus -effort low -diff-base a515c200 -key panel-4-primary -session-token mf-k1\n" +
+		"record dispatch begin -C /wt -change kan-1 -role reviewer -slot test-sanity+call-efficiency -model sonnet -effort low -key panel-4-test-sanity+call-efficiency -session-token mf-k1\n"
+	if string(calls) != want {
+		t.Errorf("flow calls:\n%s\nwant:\n%s", calls, want)
+	}
+	if !strings.Contains(out.String(), "panel-4-primary: recorded: dispatch 7\n") {
+		t.Errorf("stdout = %q, want each key beside flow's answer", out.String())
+	}
+
+	out.Reset()
+	if code := panelRoundBegins([]string{"/wt", "kan-1", "4", "mf-k1", "bad:opus:low", "primary:opus:low"}, env, &out, &errb); code != 1 {
+		t.Errorf("a refused begin: exit %d, want 1", code)
+	}
+	if !strings.Contains(out.String(), "panel-4-primary: recorded") {
+		t.Errorf("a refused begin stopped the round's later begins: %q", out.String())
+	}
+	for _, args := range [][]string{
+		{"/wt", "kan-1", "$R", "mf-k1", "primary:opus:low"},
+		{"/wt", "kan-1", "4", "mf-k1", "primary:opus"},
+		{"/wt", "kan-1", "4", "mf-k1"},
+	} {
+		if code := panelRoundBegins(args, env, &out, &errb); code != 2 {
+			t.Errorf("%q: exit %d, want 2", args, code)
+		}
+	}
+}
diff --git a/stats/internal/guard/panelroundbegins.go b/stats/internal/guard/panelroundbegins.go
new file mode 100644
index 00000000..f9147dbd
--- /dev/null
+++ b/stats/internal/guard/panelroundbegins.go
@@ -0,0 +1,67 @@
+package guard
+
+import (
+	"fmt"
+	"io"
+	"regexp"
+	"strings"
+)
+
+// panel-round-begins (scripts/panel-round-begins.sh) records every slot's
+// begin of one review-panel round, each under the literal key
+// panel-<round>-<slot> -- the round's begins block the parent hand-wrote
+// before, where kan-1026's `-key panel-$R-$s` from `R=$(flow record
+// next-round ...)` left the daemon nothing to read the key from
+// (skills/flow/review-panel.md, "Every slot's dispatch is recorded").
+// Exit 0 every begin recorded; 1 a begin failed (its output relayed, the
+// rest still recorded); 2 usage, or no flow binary on PATH.
+func init() {
+	Registry["panel-round-begins"] = panelRoundBegins
+}
+
+const prbUsage = "usage: panel-round-begins.sh <worktree> <change> <round> <session-token> <slot>:<model>:<effort>[:<diff-base>]...\n"
+
+var prbRound = regexp.MustCompile(`^[0-9]+$`)
+
+func panelRoundBegins(args []string, env Env, stdout, stderr io.Writer) int {
+	if len(args) < 5 || !prbRound.MatchString(args[2]) {
+		fmt.Fprint(stderr, prbUsage)
+		return 2
+	}
+	wt, change, round, token := args[0], args[1], args[2], args[3]
+	type slot struct{ name, model, effort, diffBase string }
+	var slots []slot
+	for _, spec := range args[4:] {
+		f := strings.Split(spec, ":")
+		if len(f) < 3 || len(f) > 4 || f[0] == "" || f[1] == "" || f[2] == "" {
+			fmt.Fprintf(stderr, "panel-round-begins: slot %q is not <slot>:<model>:<effort>[:<diff-base>]\n%s", spec, prbUsage)
+			return 2
+		}
+		s := slot{name: f[0], model: f[1], effort: f[2]}
+		if len(f) == 4 {
+			s.diffBase = f[3]
+		}
+		slots = append(slots, s)
+	}
+	flow, ok := lookPath(env, "flow")
+	if !ok {
+		fmt.Fprintln(stderr, "panel-round-begins: no flow binary on PATH")
+		return 2
+	}
+	code := 0
+	for _, s := range slots {
+		key := "panel-" + round + "-" + s.name
+		argv := []string{"record", "dispatch", "begin", "-C", wt, "-change", change, "-role", "reviewer",
+			"-slot", s.name, "-model", s.model, "-effort", s.effort}
+		if s.diffBase != "" {
+			argv = append(argv, "-diff-base", s.diffBase)
+		}
+		argv = append(argv, "-key", key, "-session-token", token)
+		rc, out := kwRun(env, "", flow, argv...)
+		fmt.Fprintf(stdout, "%s: %s\n", key, strings.TrimSpace(out))
+		if rc != 0 {
+			code = 1
+		}
+	}
+	return code
+}
diff --git a/skills/flow/scripts/panel-round-begins.sh b/skills/flow/scripts/panel-round-begins.sh
new file mode 120000
index 00000000..7bb36009
--- /dev/null
+++ b/skills/flow/scripts/panel-round-begins.sh
@@ -0,0 +1 @@
+../../../scripts/panel-round-begins.sh
\ No newline at end of file
````

  - [x] **Step 3: Edit the skill.** In `skills/flow/review-panel.md`, replace

```markdown verified:applied at plan time on c1282bc4 in order with the other tasks; check-verbatim-moves.sh exits 0 against this change's verbatim-moves.txt, check-references.sh and check-markdown-integrity.sh exit 0
Every dispatch of a round launches in one message; every `begin` is recorded in the next Bash
call, one call for all; the round's wait
```

    with

```markdown verified:applied at plan time on c1282bc4 in order with the other tasks; check-verbatim-moves.sh exits 0 against this change's verbatim-moves.txt, check-references.sh and check-markdown-integrity.sh exit 0
Every dispatch of a round launches in one message; every `begin` is recorded in the next Bash
call, one call for all, as `panel-round-begins.sh <worktree> <name> <round> mf-<literal-token>
<slot>:<model>:<effort>[:<diff-base>]...` — the begin above once per slot, under the literal key
`panel-<round>-<slot>`, `<round>` written as the number itself, never a shell substitution; the round's wait
```

  - [x] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/` is clean;
    `cd stats && go test ./internal/guard/ -run 'TestPanelRoundBeginsRecordsLiteralKeys' -count=1`
    passes; `scripts/panel-round-begins.sh` with no arguments exits 2 printing its usage; from the
    worktree root `scripts/check-guard-symlinks.sh`, `scripts/check-verbatim-moves.sh`,
    `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `scripts/panel-round-begins.sh`, `skills/flow/scripts/panel-round-begins.sh`, `skills/flow/review-panel.md`, `stats/internal/guard/panelroundbegins.go`, `stats/internal/guard/panel_round_begins_test.go`, `stats/internal/harvest/watcher.go`, `stats/internal/harvest/watcher_test.go`
**Tests:** `TestPanelRoundBeginsRecordsLiteralKeys`
**Regression:** reverting `panelroundbegins.go` fails `TestPanelRoundBeginsRecordsLiteralKeys` to compile; a key built from anything but the validated round fails its recorded-call comparison.
**Baseline:** before=276 after=277
<!-- measured: cat stats/internal/guard/*_test.go | grep -c '^func Test' @ c1282bc4 and branch self-review-kan-1026-stored-data-first-on-every-open -->
**Commit:** feat(guard): record a panel round's begins under literal keys
**After:** Task 1

**Decision:** panel-round-begins-script

- [x] 3. Close a stamped row at its agent transcript's last write
  - [x] **Step 1: Reproduce before touching anything; do not guess.**
    `grep -n 'date -u -r <report-file>' skills/flow/review-panel.md` prints the `-ended-at` line of
    the end template, and `grep -rn 'AgentTranscriptEnd' stats` prints nothing.
  - [x] **Step 2: Apply the fix.** Save the diff below by copying the lines between its fence markers into a scratch file outside the worktree, then `git apply <that file>` from the worktree root.

````diff verified:applied with git apply in order with the other tasks' diffs and prose replacements onto c1282bc4 in this worktree; the result equals the spiked tree, go vet ./... and gofmt -l clean, go test of internal/harvest, internal/store, internal/api, internal/guard, internal/repotest, internal/selfreview, internal/reconcile, cmd/flow and cmd/flowd passing, and check-verbatim-moves.sh, check-references.sh, check-guard-symlinks.sh, check-markdown-integrity.sh and check-installed-citations.sh exiting 0
diff --git a/stats/cmd/flowd/main.go b/stats/cmd/flowd/main.go
index 44345172..2b4e6a9a 100644
--- a/stats/cmd/flowd/main.go
+++ b/stats/cmd/flowd/main.go
@@ -210,8 +210,19 @@ func run(logger *slog.Logger) error {
 		return err
 	}
 
+	transcriptsRoot, err := harvest.DefaultTranscriptsRoot()
+	if err != nil {
+		return err
+	}
 	srv, err := api.New(cfg, st, st, st, st, st, logger, api.WithSPA(spaHandler), api.WithGuardRuns(st),
-		api.WithCommitBase(commitBaseFromRemote(originRemote())), api.WithBuild(buildIdentity()))
+		api.WithCommitBase(commitBaseFromRemote(originRemote())), api.WithBuild(buildIdentity()),
+		api.WithAgentEnds(func(agentID string) (time.Time, bool) {
+			at, found, err := harvest.AgentTranscriptEnd(transcriptsRoot, agentID)
+			if err != nil {
+				logger.Warn("flowd: read agent transcript end failed, the row ends at the daemon's clock", "agent_id", agentID, "error", err)
+			}
+			return at, found
+		}))
 	if err != nil {
 		// api.New calls config.Config.Validate itself, which keeps
 		// internal/api correct for any other caller, but the loopback
@@ -254,10 +265,6 @@ func run(logger *slog.Logger) error {
 	// see watcher.go's own doc comment on HarvestSink for why a purely
 	// local, separately-committed offset was unsafe (task 9's post-commit
 	// review, findings F1 and its follow-up).
-	transcriptsRoot, err := harvest.DefaultTranscriptsRoot()
-	if err != nil {
-		return err
-	}
 	rolloutRoot, err := harvest.DefaultZcodeRolloutRoot()
 	if err != nil {
 		return err
diff --git a/stats/internal/api/records.go b/stats/internal/api/records.go
index 7ba2f378..d46ac1ee 100644
--- a/stats/internal/api/records.go
+++ b/stats/internal/api/records.go
@@ -491,6 +491,9 @@ func ApplySubstitutionRecord(ctx context.Context, rw RecordStore, projectKey, ch
 type recordHandler struct {
 	store  RecordStore
 	logger *slog.Logger
+	// agentEnd answers the instant an agent's own transcript last wrote
+	// (WithAgentEnds); nil leaves every end at the daemon's clock.
+	agentEnd func(agentID string) (time.Time, bool)
 }
 
 // findingStatusRequest is the wire shape
@@ -568,7 +571,12 @@ func (h *recordHandler) endDispatch(w http.ResponseWriter, r *http.Request) {
 		return
 	}
 
+	stamped := in.EndedAt.IsZero()
 	out, err := ApplyDispatchEnd(r.Context(), h.store, project, change, in)
+	if at, ok := h.agentEndOf(out, err, stamped); ok {
+		in.EndedAt = at
+		out, err = ApplyDispatchEnd(r.Context(), h.store, project, change, in)
+	}
 	if err != nil {
 		if errors.Is(err, ErrInvalidRecord) {
 			writeError(w, http.StatusBadRequest, err.Error())
@@ -581,6 +589,25 @@ func (h *recordHandler) endDispatch(w http.ResponseWriter, r *http.Request) {
 	writeJSON(w, http.StatusOK, out)
 }
 
+// agentEndOf answers the end instant a just-closed row takes from its own
+// agent's transcript: the instant that transcript last wrote, when the end
+// carried no -ended-at, the row carries a launched agent's id, and that
+// instant falls inside the row's window. A panel round closes its slots in
+// one call once every report exists, so the daemon's clock would give each
+// the round's last finish (kan-1019); the agent's own last record is its
+// finish (kan-1026's self-review: the parent derived it from report-file
+// mtimes in Bash).
+func (h *recordHandler) agentEndOf(out records.Dispatch, err error, stamped bool) (time.Time, bool) {
+	if err != nil || !stamped || h.agentEnd == nil || out.AgentID == "" || out.AgentID == "inline" || out.EndedAt == nil {
+		return time.Time{}, false
+	}
+	at, ok := h.agentEnd(out.AgentID)
+	if !ok || !at.Before(*out.EndedAt) || at.Before(out.StartedAt) {
+		return time.Time{}, false
+	}
+	return at, true
+}
+
 // recordFinding serves POST /api/v1/records/{project}/{change}/findings.
 // It answers 201 when the write inserted a row and 200 when it replaced
 // one, so a caller can tell the round that first raised a finding from a
diff --git a/stats/internal/api/records_test.go b/stats/internal/api/records_test.go
index aadf6e06..b90e431d 100644
--- a/stats/internal/api/records_test.go
+++ b/stats/internal/api/records_test.go
@@ -16,6 +16,7 @@ import (
 	"time"
 
 	"github.com/tweety53/agents/stats/internal/api"
+	"github.com/tweety53/agents/stats/internal/config"
 	"github.com/tweety53/agents/stats/internal/records"
 	"github.com/tweety53/agents/stats/internal/store"
 )
@@ -911,6 +912,52 @@ func TestEndDispatchRouteStampsAZeroEndedAt(t *testing.T) {
 	}
 }
 
+// TestEndDispatchRouteTakesAStampedRowsEndFromItsAgent pins kan-1026's
+// self-review: an end naming no instant on a row carrying a launched
+// agent's id closes at that agent's own last transcript write, so a panel
+// round closed in one call keeps each slot's own finish; a supplied
+// instant is kept, and the agent is never asked.
+func TestEndDispatchRouteTakesAStampedRowsEndFromItsAgent(t *testing.T) {
+	fs := newFakeStore()
+	fs.changes[changeKey("proj", "kan-1")] = store.Change{ProjectKey: "proj", Name: "kan-1", State: store.StateInProgress}
+	finished := time.Date(2026, 8, 22, 9, 5, 0, 0, time.UTC)
+	asked := 0
+	srv, err := api.New(config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}, fs, fs, fs, fs, fs, nil,
+		api.WithAgentEnds(func(agentID string) (time.Time, bool) {
+			asked++
+			return finished, agentID == "a-slot"
+		}))
+	if err != nil {
+		t.Fatalf("api.New: %v", err)
+	}
+	ts := httptest.NewServer(srv.Handler())
+	t.Cleanup(ts.Close)
+
+	beginBody := dispatchBody("reviewer", "opus")
+	beginBody["agentId"] = "a-slot"
+	if resp, body := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches", beginBody); resp.StatusCode != http.StatusCreated {
+		t.Fatalf("POST dispatches = %d (%s), want 201", resp.StatusCode, body)
+	}
+	endBody := dispatchEndBody("task-3-reviewer")
+	delete(endBody, "endedAt")
+	if resp, body := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches/end", endBody); resp.StatusCode != http.StatusOK {
+		t.Fatalf("POST dispatches/end = %d (%s), want 200", resp.StatusCode, body)
+	}
+	if got := fs.dispatches[0].dispatch.EndedAt; got == nil || !got.Equal(finished) {
+		t.Errorf("endedAt = %v, want the agent's last write %s", got, finished)
+	}
+
+	supplied := time.Date(2026, 8, 22, 9, 7, 0, 0, time.UTC)
+	endBody["endedAt"] = supplied.Format(time.RFC3339)
+	asked = 0
+	if resp, body := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches/end", endBody); resp.StatusCode != http.StatusOK {
+		t.Fatalf("POST dispatches/end with endedAt = %d (%s), want 200", resp.StatusCode, body)
+	}
+	if got := fs.dispatches[0].dispatch.EndedAt; got == nil || !got.Equal(supplied) || asked != 0 {
+		t.Errorf("endedAt = %v after %d agent reads, want the supplied %s and none", got, asked, supplied)
+	}
+}
+
 // TestRecordFindingRouteAnswers201OnCreateAnd200OnUpdate pins the one
 // thing an upsert's status code has to say: whether the write inserted a
 // row or replaced one. A fix round restating F1 must be distinguishable
diff --git a/stats/internal/api/server.go b/stats/internal/api/server.go
index fceb92d8..f71ae8ae 100644
--- a/stats/internal/api/server.go
+++ b/stats/internal/api/server.go
@@ -177,6 +177,7 @@ type serverOptions struct {
 	guardRuns  GuardRunStore
 	commitBase string
 	build      string
+	agentEnd   func(agentID string) (time.Time, bool)
 }
 
 // WithSPA mounts h at "/" as the daemon's user interface, task 12's
@@ -212,6 +213,17 @@ func WithBuild(build string) Option {
 	}
 }
 
+// WithAgentEnds closes a dispatch row whose end names no instant at the
+// instant its own agent's transcript last wrote, where end answers one
+// (recordHandler.agentEndOf) -- harvest.AgentTranscriptEnd over the
+// daemon's transcripts root in production. Passing none leaves every such
+// end at the daemon's clock.
+func WithAgentEnds(end func(agentID string) (time.Time, bool)) Option {
+	return func(o *serverOptions) {
+		o.agentEnd = end
+	}
+}
+
 // apiPathPrefix is the prefix every route this daemon's API owns falls
 // under. It is registered on the mux in its own right (see New, below) so
 // that a request under it which matches none of the specific routes above
@@ -317,7 +329,7 @@ func New(cfg config.Config, cs ChangeStore, ss StageStore, sts StatsStore, rs Re
 	h := &changeHandler{store: cs, logger: logger}
 	sh := &stageHandler{store: ss, logger: logger}
 	sth := &statsHandler{store: sts, logger: logger, commitBase: so.commitBase}
-	rh := &recordHandler{store: rs, logger: logger}
+	rh := &recordHandler{store: rs, logger: logger, agentEnd: so.agentEnd}
 	srb := &selfreviewHandler{store: rs, git: selfreview.ExecRunner{}, build: so.build, logger: logger}
 	lh := &lessonsHandler{store: rs, logger: logger}
 	hz := &hazardHandler{store: rs, logger: logger}
diff --git a/stats/internal/harvest/transcript.go b/stats/internal/harvest/transcript.go
index 4503c612..b5caa38a 100644
--- a/stats/internal/harvest/transcript.go
+++ b/stats/internal/harvest/transcript.go
@@ -800,6 +800,31 @@ type DispatchMeta struct {
 // against a real transcript tree during planning, not assumed.
 const subagentsDirName = "subagents"
 
+// AgentTranscriptEnd answers the latest record instant of agentID's own
+// transcript under root -- Claude Code's <project>/<session>/subagents/
+// agent-<id>.jsonl -- the instant the agent last wrote, which is when it
+// finished once its report exists. found is false where no such file, or
+// no timestamped line in it, exists.
+func AgentTranscriptEnd(root, agentID string) (end time.Time, found bool, err error) {
+	paths, err := filepath.Glob(filepath.Join(root, "*", "*", subagentsDirName, "agent-"+agentID+".jsonl"))
+	if err != nil || len(paths) == 0 {
+		return time.Time{}, false, err
+	}
+	raw, err := os.ReadFile(paths[0])
+	if err != nil {
+		return time.Time{}, false, err
+	}
+	for line := range bytes.Lines(raw) {
+		var rec struct {
+			Timestamp time.Time `json:"timestamp"`
+		}
+		if json.Unmarshal(line, &rec) == nil && rec.Timestamp.After(end) {
+			end, found = rec.Timestamp, true
+		}
+	}
+	return end, found, nil
+}
+
 // ReadDispatchMeta reads the sidecar meta file alongside transcriptPath
 // and returns that dispatch's descriptors. It reports false, never an
 // error, in every case a dispatch's tokens must still be attributable
diff --git a/stats/internal/harvest/transcript_test.go b/stats/internal/harvest/transcript_test.go
index d6aadea4..189b14ea 100644
--- a/stats/internal/harvest/transcript_test.go
+++ b/stats/internal/harvest/transcript_test.go
@@ -673,3 +673,28 @@ func TestParseDispatchEventsCarriesOpenCallsAcrossBatches(t *testing.T) {
 		t.Errorf("stillOpen = %v, want empty (the carried call resolved)", stillOpen)
 	}
 }
+
+// TestAgentTranscriptEndReadsTheAgentsLastWrite pins the end instant
+// flowd closes a stamped row at (api.WithAgentEnds): the latest timestamp
+// in <root>/<project>/<session>/subagents/agent-<id>.jsonl.
+func TestAgentTranscriptEndReadsTheAgentsLastWrite(t *testing.T) {
+	root := t.TempDir()
+	dir := filepath.Join(root, "-proj", "sess", "subagents")
+	if err := os.MkdirAll(dir, 0o755); err != nil {
+		t.Fatal(err)
+	}
+	body := `{"type":"user","timestamp":"2026-10-23T08:16:07.000Z"}` + "\n" +
+		`{"type":"assistant","timestamp":"2026-10-23T08:21:44.500Z"}` + "\n" +
+		`not json` + "\n"
+	if err := os.WriteFile(filepath.Join(dir, "agent-af3a362078da4cf37.jsonl"), []byte(body), 0o644); err != nil {
+		t.Fatal(err)
+	}
+
+	end, found, err := harvest.AgentTranscriptEnd(root, "af3a362078da4cf37")
+	if err != nil || !found || !end.Equal(time.Date(2026, 10, 23, 8, 21, 44, 500_000_000, time.UTC)) {
+		t.Errorf("AgentTranscriptEnd = (%v, %v, %v), want (08:21:44.5Z, true, nil)", end, found, err)
+	}
+	if _, found, err := harvest.AgentTranscriptEnd(root, "a-none"); err != nil || found {
+		t.Errorf("an absent agent = (found %v, %v), want (false, nil)", found, err)
+	}
+}
````

  - [x] **Step 3: Edit the skill.** In `skills/flow/review-panel.md`'s dispatch template, replace

```markdown verified:applied at plan time on c1282bc4 in order with the other tasks; check-verbatim-moves.sh exits 0 against this change's verbatim-moves.txt, check-references.sh and check-markdown-integrity.sh exit 0
  -session-token mf-<literal-token> -outcome completed \
  -ended-at "$(date -u -r <report-file> +%Y-%m-%dT%H:%M:%SZ)"
```

    with

```markdown verified:applied at plan time on c1282bc4 in order with the other tasks; check-verbatim-moves.sh exits 0 against this change's verbatim-moves.txt, check-references.sh and check-markdown-integrity.sh exit 0
  -session-token mf-<literal-token> -outcome completed
```

    and replace the line

```markdown verified:applied at plan time on c1282bc4 in order with the other tasks; check-verbatim-moves.sh exits 0 against this change's verbatim-moves.txt, check-references.sh and check-markdown-integrity.sh exit 0
Each `end` passes its own slot's report-file modification time as `-ended-at`, so a slot that finished early is never stamped with the round's last finish (kan-1019: two slots of one round both read 15m31s). A bundled dispatch's `<report-file>` is the newest of its slots' reports.
```

    with

```markdown verified:applied at plan time on c1282bc4 in order with the other tasks; check-verbatim-moves.sh exits 0 against this change's verbatim-moves.txt, check-references.sh and check-markdown-integrity.sh exit 0
Each `end` passes no `-ended-at`: the daemon closes a stamped row at its own agent transcript's last write, so a slot that finished early is never stamped with the round's last finish (kan-1019: two slots of one round both read 15m31s).
```

  - [x] **Step 4: Prove the handler.** Change `h.agentEndOf(out, err, stamped); ok {` to
    `h.agentEndOf(out, err, stamped && false); ok {` in `stats/internal/api/records.go`, run
    `cd stats && go test ./internal/api/ -run 'TestEndDispatchRouteTakesAStampedRowsEndFromItsAgent' -count=1`:
    it fails with `endedAt = <now>, want the agent's last write`. Restore the line.
    <!-- measured: the same mutation at plan time @ branch self-review-kan-1026-stored-data-first-on-every-open → FAIL -->
  - [x] **Step 5: Verify.** `cd stats && gofmt -w . && gofmt -l . && make -s web-dist && go vet ./...`
    is clean; `cd stats && go test ./internal/api/ ./internal/harvest/ ./cmd/flowd/ -run
    'TestEndDispatchRoute|TestAgentTranscriptEndReadsTheAgentsLastWrite|TestDaemonWiresTheRealStore' -count=1`
    passes; from the worktree root `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`
    and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/review-panel.md`, `stats/cmd/flowd/main.go`, `stats/internal/api/records.go`, `stats/internal/api/records_test.go`, `stats/internal/api/server.go`, `stats/internal/harvest/transcript.go`, `stats/internal/harvest/transcript_test.go`
**Tests:** `TestEndDispatchRouteTakesAStampedRowsEndFromItsAgent`, `TestAgentTranscriptEndReadsTheAgentsLastWrite`
**Regression:** reverting `agentEndOf` closes the row at the daemon's clock, failing `TestEndDispatchRouteTakesAStampedRowsEndFromItsAgent`; reverting `AgentTranscriptEnd` fails `TestAgentTranscriptEndReadsTheAgentsLastWrite` to compile.
**Baseline:** before=78 after=80
<!-- measured: cat stats/internal/api/records_test.go stats/internal/harvest/transcript_test.go | grep -c '^func Test' @ c1282bc4 and branch self-review-kan-1026-stored-data-first-on-every-open -->
**Commit:** fix(api): close a stamped row at its agent transcript's last write
**After:** Task 1, 2

**Decision:** end-from-agent-transcript

- [x] 4. Leave a worktree's visual-verify commit to Branch backup
  - [x] **Step 1: Reproduce before touching anything; do not guess.**
    `grep -n 'Never push' skills/flow/visual-verify.md` prints the step's line, and the hand-push
    paragraph after it names the change's own `spectre/<name>` among the branches it prints a push
    for.
  - [x] **Step 2: Apply the fix.** In `skills/flow/visual-verify.md`, replace

```markdown verified:applied at plan time on c1282bc4 in order with the other tasks; check-verbatim-moves.sh exits 0 against this change's verbatim-moves.txt, check-references.sh and check-markdown-integrity.sh exit 0
    the main checkout directly. **Never push**: a file inside a repository cannot authorise a push to another repository, so no
    guard here grants one, worktree or not. `regression repo` records which repository the
    checkout is expected to be, and `check-visual-verification.sh` reports a mismatch against
    its real `origin`, but that is an identity assertion, not an authorisation. When a commit landed
    in a `regression checkout`, print the push command for the operator to run by hand, naming
    whichever branch actually received the commit — the change's own `spectre/<name>` when a
    worktree resolved, the main checkout's current branch otherwise:
```

    with

```markdown verified:applied at plan time on c1282bc4 in order with the other tasks; check-verbatim-moves.sh exits 0 against this change's verbatim-moves.txt, check-references.sh and check-markdown-integrity.sh exit 0
    the main checkout directly. **This step pushes nothing**: a file inside a repository cannot authorise a push to another repository, so no
    guard here grants one. `regression repo` records which repository the
    checkout is expected to be, and `check-visual-verification.sh` reports a mismatch against
    its real `origin`, but that is an identity assertion, not an authorisation. A commit in a
    resolved worktree is on the change's own `spectre/<name>`, and **Branch backup**
    (`skills/flow-contracts/git-boundaries.md`) pushes it with every other commit the run makes
    on that branch — no hand push is printed for it (kan-1026: the operator was told to push a
    worktree integrate had already pushed). When the commit landed in the main checkout, print
    the push command for the operator to run by hand, naming that checkout's current branch:
```

    and in `skills/flow-contracts/project-configuration-visual.md`, replace

```markdown verified:applied at plan time on c1282bc4 in order with the other tasks; check-verbatim-moves.sh exits 0 against this change's verbatim-moves.txt, check-references.sh and check-markdown-integrity.sh exit 0
`regression checkout` when one is declared and stops there; the handoff prints the push command for
the operator to run by hand.
```

    with

```markdown verified:applied at plan time on c1282bc4 in order with the other tasks; check-verbatim-moves.sh exits 0 against this change's verbatim-moves.txt, check-references.sh and check-markdown-integrity.sh exit 0
`regression checkout` when one is declared and stops there: a commit in the change's worktree of
that checkout rides **Branch backup**, and only a commit in its main checkout leaves the handoff
printing the push command for the operator to run by hand.
```

  - [x] **Step 3: Verify.** From the worktree root `scripts/check-verbatim-moves.sh`,
    `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh`,
    `scripts/check-installed-citations.sh` and `scripts/check-visual-verification.sh .` exit 0.
    Commit.

**Build:** green
**Files:** `skills/flow/visual-verify.md`, `skills/flow-contracts/project-configuration-visual.md`, `skills/flow-contracts/git-boundaries.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): leave a worktree's visual-verify commit to Branch backup
**After:** none

**Decision:** worktree-commit-rides-branch-backup

- [x] 5. Judge a non-shell heredoc body by its git invocations only
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Write a scratch script with
    the Write tool that feeds `hooks/protect-main-checkout.py` a `Bash` payload, `cwd` `/Users/tweety53/Projects/gymie` (a
    main checkout on `develop`), for each of two commands — `python3 -I - x.png <<'EOF'` /
    `if v<230 or d>0: print(y)` / `EOF`, and `D=/tmp/d; cat > $D/r.md <<'EOF'` /
    ``crop from `x-<n>.png` here`` / `EOF` / `echo $N` — and run it: the hook prints
    `"permissionDecision": "deny"` for both.
    <!-- measured: both payloads through hooks/protect-main-checkout.py @ c1282bc4 → deny, with the fix → allow; the eleven denied commands of the kan-1026 transcript → deny at c1282bc4, allow with the fix -->
  - [x] **Step 2: Apply the fix.** Save the diff below by copying the lines between its fence markers into a scratch file outside the worktree, then `git apply <that file>` from the worktree root.

````diff verified:applied with git apply in order with the other tasks' diffs and prose replacements onto c1282bc4 in this worktree; the result equals the spiked tree, go vet ./... and gofmt -l clean, go test of internal/harvest, internal/store, internal/api, internal/guard, internal/repotest, internal/selfreview, internal/reconcile, cmd/flow and cmd/flowd passing, and check-verbatim-moves.sh, check-references.sh, check-guard-symlinks.sh, check-markdown-integrity.sh and check-installed-citations.sh exiting 0
diff --git a/hooks/protect-main-checkout.py b/hooks/protect-main-checkout.py
index e5bc9ab7..1d0dbf8e 100755
--- a/hooks/protect-main-checkout.py
+++ b/hooks/protect-main-checkout.py
@@ -16,6 +16,9 @@ landing as staged changes. This hook denies the tools that create such residue:
     `cat > ABS <<EOF`, `cat >> ABS <<EOF` or `tee [-a] ABS <<EOF` to a literal absolute path,
     its body and closing line, nothing after — where only ABS is judged — and except the body of
     any `cat`/`tee` heredoc inside a longer command, which is that writer's data, never commands.
+    A body no shell runs — a `python`/`python3`/`node` program on stdin, or a file the rest of
+    the command runs only through one of those or `npx playwright test` — is that language's
+    text: only the git invocations its lines name are judged (kan-1026: `d>0` read as a write).
 
 "Main checkout" is the worktree whose git dir IS the common dir — a linked worktree's `--git-dir`
 is `<common>/worktrees/<name>`, so `<repo>-worktrees/<change>` beside the main checkout passes,
@@ -305,10 +308,18 @@ def names_target(word, target):
 # Commands known not to execute a file they are given: the heredoc target named only as an
 # argument of one of these is data. Anything else naming it may run it, so the body stays scanned.
 READER_WORDS = {
-    "cat", "tee", "head", "tail", "wc", "grep", "diff", "cmp", "ls", "test", "[", "stat",
+    "cat", "tee", "head", "tail", "wc", "grep", "diff", "cmp", "ls", "test", "[", "stat", "cd",
+    "echo", "printf",
 }
+# Shell keywords that open a command without being one: `do x`'s command is `x`.
+LEADING_KEYWORDS = {"do", "then", "else", "elif", "!"}
 GIT_READ_SUBCOMMANDS = {"diff", "add", "show", "log", "status"}
 CHECK_SCRIPT_RE = re.compile(r"check-[\w.-]*\.(?:sh|py)")
+# Pipeline scripts that only read a file argument: render-slot-prompt.sh names its -fix-report
+# paths in the prompt it prints (kan-1026: a panel round's report write was denied over its body).
+PIPELINE_READERS = {"render-slot-prompt.sh"}
+# Interpreters whose program text is not shell: a body they run is scanned for git verbs only.
+PROGRAM_RUNNERS = {"python", "python3", "node"}
 
 
 def reads_only(words, target):
@@ -319,6 +330,8 @@ def reads_only(words, target):
     name = words[0].rsplit("/", 1)[-1]
     if CHECK_SCRIPT_RE.fullmatch(name) and name != target.rsplit("/", 1)[-1]:
         return True
+    if name in PIPELINE_READERS:
+        return True
     if name in READER_WORDS:
         return True
     if name == "git":
@@ -328,6 +341,34 @@ def reads_only(words, target):
     return False
 
 
+def command_segments(text):
+    """(words, assigned) per command of text — quotes stripped, a leading `do`/`then`/`!`
+    dropped, `assigned` when NAME=value words led it — or None when text holds a substitution
+    this cannot read: a process substitution, a backtick, a nested `$(`. A plain `$(...)` is a
+    command of its own. A literal NAME=value expands later words, so `$D/r.md` reads as the file
+    D names; a name no literal assigns stays an expansion `names_target` reads as hiding any."""
+    spans = re.findall(r"\$\(([^()]*)\)", text)
+    flat = re.sub(r"\$\([^()]*\)", " ", text)
+    if SUBST_RE.search(flat):
+        return None
+    env, out = {}, []
+    for seg in COMMAND_CUT_RE.split(flat + "\n" + "\n".join(spans)):
+        words = [w.strip("'\"") for w in seg.split()]
+        while words and words[0] in LEADING_KEYWORDS:
+            words.pop(0)
+        assigned = False
+        while words and ASSIGNMENT_RE.match(words[0]):
+            name, _, value = words.pop(0).partition("=")
+            value = expand_vars(value, env)
+            if value and "$" not in value:
+                env[name] = value
+            else:
+                env.pop(name, None)
+            assigned = True
+        out.append(([expand_vars(w, env) for w in words], assigned))
+    return out
+
+
 def may_run_code(text, target):
     """True when text — the command outside one heredoc body — could run that body or the file
     target it writes. Fails closed: a process or command substitution (`>(sh)`, `bash <(`,
@@ -338,16 +379,12 @@ def may_run_code(text, target):
     `./gradlew`) leaves the body as data; `php $T`, `setsid $T`, `find . -exec`, `ssh h $T`
     keep it scanned, as does any command behind a leading `NAME=value` or run under
     `env`/`export`. Over-matching only keeps the body scanned."""
-    if SUBST_RE.search(text):
+    segments = command_segments(text)
+    if segments is None:
         return True
-    for seg in COMMAND_CUT_RE.split(text):
-        words = [w.strip("'\"") for w in seg.split()]
+    for words, assigned in segments:
         if any(w.rsplit("/", 1)[-1] in RUNNER_WORDS for w in words):
             return True
-        assigned = False
-        while words and ASSIGNMENT_RE.match(words[0]):
-            words.pop(0)
-            assigned = True
         # An env assignment can make any command run a file (GIT_EXTERNAL_DIFF, PAGER,
         # LD_PRELOAD, …): a command carrying one, or run under `env`/`export`, is never a reader.
         if words and (assigned or words[0].rsplit("/", 1)[-1] in {"env", "export"}):
@@ -357,6 +394,45 @@ def may_run_code(text, target):
     return False
 
 
+def runs_as_program(text, target):
+    """True when text — the command outside one heredoc body — runs target only as a non-shell
+    program: no substitution, and every command naming target is a reader (`reads_only`), a
+    PROGRAM_RUNNERS interpreter given it (`python3 x.py`, `node x.js`) or `npx playwright test`,
+    with no other runner word anywhere. Such a body is that language's text, never shell."""
+    segments = command_segments(text)
+    if segments is None:
+        return False
+    for words, assigned in segments:
+        if not words:
+            continue
+        if words[0].rsplit("/", 1)[-1] in PROGRAM_RUNNERS or words[:3] == ["npx", "playwright", "test"]:
+            if any(w.rsplit("/", 1)[-1] in RUNNER_WORDS for w in words[1:]):
+                return False
+            continue
+        if any(w.rsplit("/", 1)[-1] in RUNNER_WORDS for w in words):
+            return False
+        if assigned or words[0].rsplit("/", 1)[-1] in {"env", "export"}:
+            return False
+        if any(names_target(w, target) for w in words) and not reads_only(words, target):
+            return False
+    return True
+
+
+PROGRAM_GIT_RE = re.compile(r"(?:^|\s)(git\s.*)$")
+
+
+def program_git_lines(body):
+    """The git invocations a non-shell program body may make, one scan line each: quotes and
+    brackets become spaces, so `subprocess.run(["git", "-C", d, "commit"])` reads `git -C d
+    commit`, and every other word — a `d>0` comparison, an `rm` in a string — is dropped."""
+    out = []
+    for line in body.split("\n"):
+        m = PROGRAM_GIT_RE.search(re.sub(r"[\"'()\[\],<>|;&{}]", " ", line))
+        if m:
+            out.append(m.group(1))
+    return out
+
+
 # A `$(...)` without a nested paren, or a backtick pair: the spans of an unquoted heredoc body
 # that bash runs.
 SUBST_SPAN_RE = re.compile(r"\$\([^()]*\)|`[^`]*`")
@@ -400,6 +476,20 @@ def strip_data_heredocs(command):
         words = seg.split()
         while words and ASSIGNMENT_RE.match(words[0]):
             words.pop(0)
+        delim = m.group(1) or m.group(2) or m.group(3)
+        tabs = line[m.start() : m.start() + 3] == "<<-"
+        end = next(
+            (j for j in range(k, len(lines)) if (lines[j].lstrip("\t") if tabs else lines[j]).rstrip("\r") == delim),
+            None,
+        )
+        if words and words[0].rsplit("/", 1)[-1] in PROGRAM_RUNNERS and end is not None:
+            # A python/node program on stdin: its text is not shell (kan-1026: `d>0`).
+            body = "\n".join(lines[k:end])
+            if m.group(3) and any(c in body for c in ("$(", "`")):
+                continue
+            out.extend(program_git_lines(body))
+            k = end + 1
+            continue
         if not words or words[0] not in ("cat", "tee"):
             continue
         if words[0] == "cat":
@@ -407,17 +497,15 @@ def strip_data_heredocs(command):
             target = t.group(1) if t else None
         else:
             target = next((w for w in words[1:] if not w.startswith(("-", ">"))), None)
-        delim = m.group(1) or m.group(2) or m.group(3)
-        tabs = line[m.start() : m.start() + 3] == "<<-"
-        end = next(
-            (j for j in range(k, len(lines)) if (lines[j].lstrip("\t") if tabs else lines[j]).rstrip("\r") == delim),
-            None,
-        )
         if not target or end is None:
             continue
         if target.startswith("/dev/") or "(" in target:
             continue
-        if may_run_code("\n".join(lines[:k] + lines[end + 1 :]), target):
+        rest = "\n".join(lines[:k] + lines[end + 1 :])
+        if may_run_code(rest, target):
+            if runs_as_program(rest, target) and not (m.group(3) and any(c in "\n".join(lines[k:end]) for c in ("$(", "`"))):
+                out.extend(program_git_lines("\n".join(lines[k:end])))
+                k = end + 1
             continue
         # An escaped backtick is a literal in an unquoted body; dropped, it cannot pair with a real one.
         body = "\n".join(lines[k:end]).replace("\\`", "")
diff --git a/stats/internal/repotest/protect_main_checkout_test.go b/stats/internal/repotest/protect_main_checkout_test.go
index 1c561700..1122503c 100644
--- a/stats/internal/repotest/protect_main_checkout_test.go
+++ b/stats/internal/repotest/protect_main_checkout_test.go
@@ -456,6 +456,17 @@ func TestProtectMainCheckout(t *testing.T) {
 	f.bash("142 an unbalanced quote, then '>' glued to a main-checkout path", "deny", "{ROOT}", `echo it's >{MAIN}/f.txt`)
 	f.bash("149 '>&' into the main checkout", "deny", "{ROOT}", `echo x >& {MAIN}/f.txt`)
 
+	// 160-167: a body no shell runs is not shell text (kan-1026's self-review: eleven worktree-only
+	// commands denied); its git invocations are still judged.
+	f.bash("160 a python heredoc body comparing d>0, cwd = main", "allow", "{MAIN}", "python3 -I - x.png <<'EOF'\nif v<230 or d>0: print(y)\nEOF")
+	f.bash("161 a python heredoc body running git commit in main", "deny", "{ROOT}", "python3 - <<'EOF'\nsubprocess.run([\"git\", \"-C\", \"{MAIN}\", \"commit\", \"-m\", \"x\"])\nEOF")
+	f.bash("162 a report heredoc then echo of an unassigned name, cwd = main", "allow", "{MAIN}", "D={SIB}; cat > $D/r.md <<'EOF'\ncrop from `x-<n>.png` here\nEOF\necho $N")
+	f.bash("163 a report heredoc then render-slot-prompt.sh -fix-report it, cwd = main", "allow", "{MAIN}", "D={SIB}; SK=$(readlink -f ~/.claude/skills/flow); cat > $D/r.md <<'EOF'\nfrom `x-<n>.png`\nEOF\nfor s in a b; do render-slot-prompt.sh $SK 4 -fix-report $D/r.md; done")
+	f.bash("164 a python file written then run by python3, cwd = main", "allow", "{MAIN}", "O={SIB}; cat > $O/t.py <<'EOF'\nif a > 0: open(p, 'w')\nEOF\npython3 $O/t.py {SIB}")
+	f.bash("165 a spec written then run by npx playwright test, cwd = main", "allow", "{MAIN}", "cat > {SIB}/a.spec.ts <<'EOF'\nconst k = (x) => x > 0\nEOF\ncd {SIB} && U=http://x npx playwright test a.spec.ts")
+	f.bash("166 a python file written then run by bash keeps its body scanned", "deny", "{ROOT}", "cat > {SIB}/t.py <<'EOF'\nrm {MAIN}/f.txt\nEOF\nbash {SIB}/t.py")
+	f.bash("167 a report heredoc then a substitution running it keeps its body scanned", "deny", "{ROOT}", "cat > {SIB}/r.md <<'EOF'\nrm {MAIN}/f.txt\nEOF\nX=$({SIB}/r.md)")
+
 	// 151-155: a restore that would discard unstaged edits in a linked worktree (kan-1006's self-review)
 	f.write(wt+"/f.txt", "flipped\n")
 	f.bash("151 checkout -- of a path with unstaged edits", "deny", "{ROOT}", `git -C {WT} checkout -- f.txt`)
````

  - [x] **Step 3: Verify.** `python3 -m py_compile hooks/protect-main-checkout.py` succeeds;
    `cd stats && go test ./internal/repotest/ -run 'TestProtectMainCheckout' -count=1` passes; with
    the hook at HEAD's content cases 160–165 fail, and 166 and 167 pass either way (restore the
    patched hook); Step 1's script now prints no deny; from the worktree root
    `scripts/check-python-suppressions.sh` exits 0. Commit.
    <!-- measured: TestProtectMainCheckout with the hook at c1282bc4 @ branch self-review-kan-1026-stored-data-first-on-every-open → 160-165 FAIL, 166-167 PASS -->

**Build:** green
**Files:** `hooks/protect-main-checkout.py`, `stats/internal/repotest/protect_main_checkout_test.go`
**Tests:** `160 a python heredoc body comparing d>0, cwd = main`, `161 a python heredoc body running git commit in main`, `162 a report heredoc then echo of an unassigned name, cwd = main`, `163 a report heredoc then render-slot-prompt.sh -fix-report it, cwd = main`, `164 a python file written then run by python3, cwd = main`, `165 a spec written then run by npx playwright test, cwd = main`, `166 a python file written then run by bash keeps its body scanned`, `167 a report heredoc then a substitution running it keeps its body scanned`
**Regression:** reverting the program-body route denies 160, 164 and 165 again and lets 161's python `git commit` through; reverting the segment reading denies 162 and 163; 166 and 167 fail if the narrowing lets a shell-run or substitution-run body go unscanned.
**Baseline:** before=127 after=135
<!-- measured: grep -c '^	f.bash("' stats/internal/repotest/protect_main_checkout_test.go @ c1282bc4 and branch self-review-kan-1026-stored-data-first-on-every-open -->
**Commit:** fix(hooks): judge a non-shell heredoc body by its git invocations only
**After:** none

**Decision:** program-body-git-only

- [x] 6. Refuse a token report on a harvested session
  - [x] **Step 1: Reproduce before touching anything; do not guess.**
    `grep -n 'CLAUDE_CODE_SESSION_ID' stats/cmd/flow/record.go stats/internal/store/records.go`
    prints only `record.go`'s environment check: nothing in the store refuses `-tokens`.
  - [x] **Step 2: Apply the fix.** Save the diff below by copying the lines between its fence markers into a scratch file outside the worktree, then `git apply <that file>` from the worktree root.

````diff verified:applied with git apply in order with the other tasks' diffs and prose replacements onto c1282bc4 in this worktree; the result equals the spiked tree, go vet ./... and gofmt -l clean, go test of internal/harvest, internal/store, internal/api, internal/guard, internal/repotest, internal/selfreview, internal/reconcile, cmd/flow and cmd/flowd passing, and check-verbatim-moves.sh, check-references.sh, check-guard-symlinks.sh, check-markdown-integrity.sh and check-installed-citations.sh exiting 0
diff --git a/stats/cmd/flow/record.go b/stats/cmd/flow/record.go
index d7b83534..29865f6a 100644
--- a/stats/cmd/flow/record.go
+++ b/stats/cmd/flow/record.go
@@ -562,7 +562,11 @@ otherwise read "not measured" in the ledger and in every cost summary
 forever. A report replaces whatever the harvester had credited by the
 time the end lands; main-thread records inside the window that are
 harvested after it still add on top. Where CLAUDE_CODE_SESSION_ID is set
-the harvester reads that transcript, so -tokens is refused there. The value
+the harvester reads that transcript, so -tokens is refused there -- and the
+daemon refuses it for any session token begun on a harness whose
+transcript it harvests (claude-code, zcode), whatever the caller's
+environment holds; a begin and commit made in one call is recorded with
+"flow record dispatch backfill -agent-id inline -commit <sha>". The value
 is comma-separated key=value pairs; the keys are input, output, cache-read
 and cache-creation, and every value is a non-negative integer. Any
 malformed value is a caller mistake refused before the store is ever
diff --git a/stats/internal/api/server.go b/stats/internal/api/server.go
index f71ae8ae..4b22feb6 100644
--- a/stats/internal/api/server.go
+++ b/stats/internal/api/server.go
@@ -604,7 +604,8 @@ func mapStoreError(logger *slog.Logger, action string, err error) (status int, m
 		return http.StatusNotFound, err.Error()
 	case errors.Is(err, store.ErrAgentIDInvalid):
 		return http.StatusBadRequest, err.Error()
-	case errors.Is(err, store.ErrInlineWindowTooShort), errors.Is(err, store.ErrInlineCommitMissing), errors.Is(err, store.ErrInlineBegunAfterCommit):
+	case errors.Is(err, store.ErrInlineWindowTooShort), errors.Is(err, store.ErrInlineCommitMissing), errors.Is(err, store.ErrInlineBegunAfterCommit),
+		errors.Is(err, store.ErrTokensOnHarvestedSession):
 		return http.StatusConflict, err.Error()
 	case errors.Is(err, store.ErrDispatchPairInvalid):
 		return http.StatusBadRequest, err.Error()
diff --git a/stats/internal/reconcile/reconcile.go b/stats/internal/reconcile/reconcile.go
index 682f6cc0..98181f1e 100644
--- a/stats/internal/reconcile/reconcile.go
+++ b/stats/internal/reconcile/reconcile.go
@@ -845,13 +845,16 @@ func isDefinitiveRecordOutcome(err error) bool {
 		// above get -- leaving it queued would block every valid entry
 		// behind it forever without ever making progress.
 		return true
-	case errors.Is(err, store.ErrInlineWindowTooShort), errors.Is(err, store.ErrInlineCommitMissing), errors.Is(err, store.ErrInlineBegunAfterCommit):
+	case errors.Is(err, store.ErrInlineWindowTooShort), errors.Is(err, store.ErrInlineCommitMissing), errors.Is(err, store.ErrInlineBegunAfterCommit),
+		errors.Is(err, store.ErrTokensOnHarvestedSession):
 		// An inline end closed too soon (or with no commit, or with a
 		// commit made before its begin -- kan-1003) after its begin is refused on the
 		// begin's recorded start and the end's own ended_at, neither of
 		// which a later pass changes, so every replay is refused
-		// identically. Retiring it keeps it from blocking every valid
-		// entry behind it forever.
+		// identically. A token report on a harvested session is refused on
+		// the session's recorded harness, which no replay changes either.
+		// Retiring it keeps it from blocking every valid entry behind it
+		// forever.
 		return true
 	default:
 		return false
diff --git a/stats/internal/store/records.go b/stats/internal/store/records.go
index 522f2b87..e344d7e9 100644
--- a/stats/internal/store/records.go
+++ b/stats/internal/store/records.go
@@ -106,6 +106,19 @@ var ErrAgentIDInvalid = errors.New("store: invalid agent id")
 // cost unmeasured.
 var ErrInlineWindowTooShort = errors.New("store: inline dispatch window too short")
 
+// ErrTokensOnHarvestedSession is returned by EndDispatch when an end carries
+// a caller token report (-tokens) for a session token begun on a harness
+// whose transcript the harvester reads (harvestedHarnesses): the harvester
+// credits that row, and the report would replace what it measured.
+// Refused on the store's own record, never the caller's environment --
+// kan-1026's full-suite-fix-1 row went through with CLAUDE_CODE_SESSION_ID
+// unset for one call.
+var ErrTokensOnHarvestedSession = errors.New("store: token report refused on a harvested session")
+
+// harvestedHarnesses are the stage-run harnesses whose transcripts the
+// daemon's watcher reads (cmd/flowd: the Claude and ZCode rollout sources).
+var harvestedHarnesses = []string{"claude-code", "zcode"}
+
 // ErrInlineCommitMissing is returned by EndDispatch when an inline row on a
 // task's implementer key (`task-<n>-implementer…`) closes completed with no
 // commit sha: kan-1003's tasks 10–22 recorded none, so no row could be tied
@@ -552,6 +565,17 @@ func (s *Store) EndDispatch(ctx context.Context, projectKey, change string, in r
 	if err := validateAgentID(in.AgentID); err != nil {
 		return records.Dispatch{}, fmt.Errorf("store: end dispatch %q for %s/%s: %w", in.Key, projectKey, change, err)
 	}
+	if in.Tokens != nil {
+		var harvested bool
+		if err := s.pool.QueryRow(ctx, `
+			SELECT EXISTS (SELECT 1 FROM stage_runs WHERE session_token = $1 AND harness = ANY($2))`,
+			in.SessionToken, harvestedHarnesses).Scan(&harvested); err != nil {
+			return records.Dispatch{}, fmt.Errorf("store: end dispatch %q for %s/%s: read session harness: %w", in.Key, projectKey, change, err)
+		}
+		if harvested {
+			return records.Dispatch{}, fmt.Errorf("%w: %s/%s key %q -- session token %q runs on a harness whose transcript the harvester reads; drop -tokens, and record a begin and commit made in one call with flow record dispatch backfill -agent-id inline -commit <sha>", ErrTokensOnHarvestedSession, projectKey, change, in.Key, in.SessionToken)
+		}
+	}
 	if in.Outcome == "completed" {
 		var agentID *string
 		var startedAt time.Time
diff --git a/stats/internal/store/records_test.go b/stats/internal/store/records_test.go
index 1b3cdee4..41db3a0f 100644
--- a/stats/internal/store/records_test.go
+++ b/stats/internal/store/records_test.go
@@ -3807,6 +3807,39 @@ func TestListFindingPatternOccurrencesRefusesSeparatorsOnly(t *testing.T) {
 // stores it under the bag's "tokens.main" and stamps the row "reported" --
 // the two facts the ledger's caller-reported qualifier and CostStatusOf's
 // counted-as-measured rule both read.
+func TestEndDispatchRefusesTokensOnAHarvestedSession(t *testing.T) {
+	st, _ := newRecordStore(t)
+	ctx := context.Background()
+
+	projectKey := fmt.Sprintf("proj-end-tokens-harvested-%d", time.Now().UnixNano())
+	seedChange(t, st, projectKey, "kan-1")
+	const token, key = "mf-k1026r1", "full-suite-fix-1"
+	begin := baseBeginInput(projectKey, "kan-1", "/flow", "implementation")
+	begin.SessionToken = ptr(token)
+	if _, err := st.BeginStage(ctx, begin); err != nil {
+		t.Fatalf("BeginStage: %v", err)
+	}
+	dispatch := baseDispatch("implementer", "opus")
+	dispatch.SessionToken, dispatch.Key, dispatch.AgentID = token, key, "inline"
+	if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", dispatch); err != nil {
+		t.Fatalf("RecordDispatch: %v", err)
+	}
+
+	_, err := st.EndDispatch(ctx, projectKey, "kan-1", records.DispatchEnd{
+		SessionToken: token,
+		Key:          key,
+		Outcome:      "completed",
+		EndedAt:      dispatch.StartedAt,
+		Tokens:       &records.TokenReport{Input: 1, Output: 1},
+	})
+	if !errors.Is(err, store.ErrTokensOnHarvestedSession) {
+		t.Fatalf("EndDispatch with -tokens on a claude-code session = %v, want ErrTokensOnHarvestedSession", err)
+	}
+	if !strings.Contains(err.Error(), "flow record dispatch backfill -agent-id inline -commit") {
+		t.Errorf("refusal names no sanctioned route: %v", err)
+	}
+}
+
 func TestEndDispatchWritesReportedTokensIntoBag(t *testing.T) {
 	st, _ := newRecordStore(t)
 	ctx := context.Background()
@@ -3819,6 +3852,9 @@ func TestEndDispatchWritesReportedTokensIntoBag(t *testing.T) {
 	)
 	begin := baseBeginInput(projectKey, "kan-1", "/flow-fast", "inline implementation")
 	begin.SessionToken = ptr(token)
+	// A harness whose transcript the harvester never reads: -tokens is its
+	// only measurement (ErrTokensOnHarvestedSession).
+	begin.Harness = "muse"
 	if _, err := st.BeginStage(ctx, begin); err != nil {
 		t.Fatalf("BeginStage: %v", err)
 	}
@@ -3895,6 +3931,9 @@ func TestEndDispatchTokenReplayIsIdempotent(t *testing.T) {
 	)
 	begin := baseBeginInput(projectKey, "kan-1", "/flow-fast", "inline implementation")
 	begin.SessionToken = ptr(token)
+	// A harness whose transcript the harvester never reads: -tokens is its
+	// only measurement (ErrTokensOnHarvestedSession).
+	begin.Harness = "muse"
 	if _, err := st.BeginStage(ctx, begin); err != nil {
 		t.Fatalf("BeginStage: %v", err)
 	}
````

  - [x] **Step 3: Verify.** `cd stats && gofmt -w . && gofmt -l . && make -s web-dist && go vet ./...`
    is clean; `cd stats && go test ./internal/store/ ./internal/api/ ./internal/reconcile/ ./cmd/flow/ -run
    'TestEndDispatch|Token|Inline|Reconcile' -count=1` passes, `TestEndDispatchRefusesTokensOnAHarvestedSession`
    among them and not skipped. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/internal/api/server.go`, `stats/internal/reconcile/reconcile.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`
**Tests:** `TestEndDispatchRefusesTokensOnAHarvestedSession`
**Regression:** reverting the store refusal lets the claude-code session's token report through, failing `TestEndDispatchRefusesTokensOnAHarvestedSession`.
**Baseline:** before=97 after=98
<!-- measured: grep -c '^func Test' stats/internal/store/records_test.go @ branch self-review-kan-1026-stored-data-first-on-every-open (after Task 1, and after this task) -->
**Commit:** fix(store): refuse a token report on a harvested session
**After:** Task 1, 3

**Decision:** tokens-refused-on-harvested-session

- [x] 7. Elide the verification record's bands, seams and ink sections from the bundle
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In the kan-1026 bundle the
    pass read, the `## Bands and seams (raw summaries)` and `## Ink` sections of
    `visual-verification.md` sit between its matrix sections and `live-verification.md`, and
    `grep -n 'Bands and seams\|elideMeasurement' stats/internal/selfreview/git.go` prints nothing.
  - [x] **Step 2: Apply the fix.** Save the diff below by copying the lines between its fence markers into a scratch file outside the worktree, then `git apply <that file>` from the worktree root.

````diff verified:applied with git apply in order with the other tasks' diffs and prose replacements onto c1282bc4 in this worktree; the result equals the spiked tree, go vet ./... and gofmt -l clean, go test of internal/harvest, internal/store, internal/api, internal/guard, internal/repotest, internal/selfreview, internal/reconcile, cmd/flow and cmd/flowd passing, and check-verbatim-moves.sh, check-references.sh, check-guard-symlinks.sh, check-markdown-integrity.sh and check-installed-citations.sh exiting 0
diff --git a/skills/flow-contracts/finish-contract-run1.md b/skills/flow-contracts/finish-contract-run1.md
index aeb39ff6..70b0f8ad 100644
--- a/skills/flow-contracts/finish-contract-run1.md
+++ b/skills/flow-contracts/finish-contract-run1.md
@@ -349,7 +349,9 @@ archive commit of the change, oldest first, followed by each peer the archived `
 `## Merge order` names, its `spectre/<name>` commits against its base under a `### <peer>`
 heading — and prints it as one Markdown document. Its `visual-verification.md` keeps every heading and
 drops the table rows under each heading naming a matrix — any level, `matrix` in its text, any
-case — and its `git log --stat` lines leave out every `*.png` path. A source that is absent is reported
+case — replaces each section whose heading names bands, seams or ink, through the next heading of
+its level or higher, with one line naming the lines it held, its heading and the record's path,
+and its `git log --stat` lines leave out every `*.png` path. A source that is absent is reported
 `skipped: <source> (absent)` inside the bundle, never fatal.
 The session appends `## Session narrative` (one paragraph it writes for run 1 itself; the
 archived `narrative.md` is already a bundle section, and a change predating the narrative rule
diff --git a/stats/internal/selfreview/bundle_test.go b/stats/internal/selfreview/bundle_test.go
index e70af2bf..3b366dbe 100644
--- a/stats/internal/selfreview/bundle_test.go
+++ b/stats/internal/selfreview/bundle_test.go
@@ -267,6 +267,37 @@ func TestBundleElidesMatrixRowsUnderAnyMatrixHeading(t *testing.T) {
 	}
 }
 
+func TestBundleElidesBandsSeamsAndInkSections(t *testing.T) {
+	repo := gitRepo(t)
+	commitAll(t, repo, "a.go", "package a\n", "feat(demo): a")
+	writeArchiveBranch(t, repo, "demo", map[string]string{
+		"tasks.md": "# demo tasks\n",
+		"visual-verification.md": "# Visual verification — demo\n\nHEADER-LINE\n\n## Verify\n\nVERIFY-LINE\n\n" +
+			"## Views\n\n- q2: departure — VERDICT-LINE\n\n" +
+			"## Bands and seams (raw summaries)\n\n### q2-answer\n\nBAND-LINE\n\n" +
+			"## Ink (runs through each band)\n\nINK-LINE\n\n## Motions\n\nMOTION-LINE\n",
+	})
+
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, ExecRunner{})
+	if err != nil {
+		t.Fatalf("Bundle: %v", err)
+	}
+
+	label := "spectre/changes/archive/demo/visual-verification.md"
+	for _, want := range []string{"HEADER-LINE", "VERIFY-LINE", "VERDICT-LINE", "MOTION-LINE",
+		"_(6 lines elided: Bands and seams (raw summaries) — in " + label + ")_",
+		"_(4 lines elided: Ink (runs through each band) — in " + label + ")_"} {
+		if !strings.Contains(bundle, want) {
+			t.Errorf("bundle missing %q:\n%s", want, bundle)
+		}
+	}
+	for _, gone := range []string{"BAND-LINE", "INK-LINE", "### q2-answer"} {
+		if strings.Contains(bundle, gone) {
+			t.Errorf("bundle still carries measurement line %q:\n%s", gone, bundle)
+		}
+	}
+}
+
 func TestBundleElidesMatrixRowsAndPNGStatLines(t *testing.T) {
 	repo := gitRepo(t)
 	if err := os.WriteFile(filepath.Join(repo, "shot.png"), []byte("\x89PNG"), 0o644); err != nil {
diff --git a/stats/internal/selfreview/git.go b/stats/internal/selfreview/git.go
index 83d5a8df..de16a757 100644
--- a/stats/internal/selfreview/git.go
+++ b/stats/internal/selfreview/git.go
@@ -489,6 +489,51 @@ func elideMatrixRows(md string) string {
 	return b.String()
 }
 
+// measurementHeading names a verification record's raw measurement
+// section: the bands and seams summaries and the ink runs the verifier
+// writes beside its verdicts (kan-1026's self-review: about 200 of the
+// bundle's lines, read by no angle).
+var measurementHeading = regexp.MustCompile(`(?i)\b(bands|seams|ink)\b`)
+
+// elideMeasurementSections drops every section whose heading names a raw
+// measurement (measurementHeading), from that heading through the line
+// before the next heading of its level or higher, leaving one `_(<n> lines
+// elided: <heading> — in <label>)_` line per section. The header, the
+// verify output and every verdict section stay whole, and so does a matrix
+// section, whose rows elideMatrixRows already dropped. A line inside a ```
+// fence is never a heading.
+func elideMeasurementSections(md, label string) string {
+	var b strings.Builder
+	inFence, level, n, heading := false, 0, 0, ""
+	flush := func() {
+		if level > 0 {
+			fmt.Fprintf(&b, "_(%d lines elided: %s — in %s)_\n\n", n, heading, label)
+		}
+		level, n = 0, 0
+	}
+	for l := range strings.Lines(md) {
+		if strings.HasPrefix(strings.TrimSpace(l), "```") {
+			inFence = !inFence
+		}
+		if !inFence && strings.HasPrefix(l, "#") {
+			depth := len(l) - len(strings.TrimLeft(l, "#"))
+			if level > 0 && depth <= level {
+				flush()
+			}
+			if level == 0 && measurementHeading.MatchString(l) {
+				level, heading = depth, strings.TrimSpace(strings.TrimLeft(l, "# "))
+			}
+		}
+		if level > 0 {
+			n++
+			continue
+		}
+		b.WriteString(l)
+	}
+	flush()
+	return b.String()
+}
+
 func trimmedOutput(b []byte, err error) string {
 	if err != nil {
 		return ""
@@ -620,7 +665,7 @@ func Bundle(change string, run records.Run, summary string, summaryFound bool, p
 			content, err := show(g, repo, rev, label)
 			text := string(content)
 			if file == "visual-verification.md" {
-				text = elideMatrixRows(text)
+				text = elideMeasurementSections(elideMatrixRows(text), label)
 			}
 			add(label, text, err == nil)
 		}
````

  - [x] **Step 3: Verify.** `cd stats && gofmt -l . && go vet ./internal/selfreview/` is clean;
    `cd stats && go test ./internal/selfreview/ -run 'TestBundleElides' -count=1` passes; with the
    `elideMeasurementSections(` call reverted to `text = elideMatrixRows(text)` the new test fails
    (restore it); from the worktree root `scripts/check-verbatim-moves.sh` and
    `scripts/check-references.sh` exit 0. Commit.
    <!-- measured: the same revert at plan time @ branch self-review-kan-1026-stored-data-first-on-every-open → TestBundleElidesBandsSeamsAndInkSections FAIL -->

**Build:** green
**Files:** `skills/flow-contracts/finish-contract-run1.md`, `stats/internal/selfreview/bundle_test.go`, `stats/internal/selfreview/git.go`
**Tests:** `TestBundleElidesBandsSeamsAndInkSections`
**Regression:** reverting `elideMeasurementSections` keeps the bands and ink lines in the bundle, failing `TestBundleElidesBandsSeamsAndInkSections`.
**Baseline:** before=31 after=32
<!-- measured: grep -c '^func Test' stats/internal/selfreview/bundle_test.go @ c1282bc4 and branch self-review-kan-1026-stored-data-first-on-every-open -->
**Commit:** fix(selfreview): elide the verification record's bands, seams and ink sections
**After:** none

**Decision:** bundle-elides-raw-measurements
