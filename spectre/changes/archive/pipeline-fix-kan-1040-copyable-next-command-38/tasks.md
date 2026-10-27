# pipeline-fix-kan-1040-copyable-next-command-38

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** `integrate-change.sh land`'s final `NEXT:` line names the Finished handoff's template
and its never-omitted `Learning:` and `Prompt tuning:` lines, and the `flow-stop-guard.py` Stop
hook blocks a turn whose final handoff drops either line.

**Architecture:** one string in `stats/internal/guard/integratechange_cleanup.go`; one check in
`stats/internal/guard/integrate_change_test.go`'s happy path; a transcript check in
`hooks/flow-stop-guard.py` ahead of its store lookup, pinned in
`stats/internal/repotest/flow_stop_guard_hook_test.go`.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/`.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; it was applied to this worktree at 71c7e91f and checked there. A diff that does not
  apply is reported, never hand-merged by guesswork.
- No new inline suppressions. Never stop flowd or flow-postgres.

## Review Focus

- The line still starts `NEXT: print the Finished handoff`, so both existing `^NEXT:` greps hold
  (`next-line-names-template`).
- The handoff check runs before the store lookup, is skipped when `stop_hook_active` is true, and
  any fault still exits 0 with no output (`stop-hook-blocks-incomplete-handoff`).

- [x] 1. Name the Finished handoff's template in land's NEXT line
  - [x] **Step 1: Reproduce.** Save the diff below to `/tmp/pf38.diff`, apply its test hunk alone
    (`git apply --include='stats/internal/guard/integrate_change_test.go' /tmp/pf38.diff`), run
    `cd stats && make -s web-dist && go test ./internal/guard/ -run TestIntegrateChange`:
    `TestIntegrateChangeHappyPath/happy_path/the_NEXT_line_names_the_template_and_its_never-omitted_lines` fails.
  - [x] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='stats/internal/guard/integrate_change_test.go' /tmp/pf38.diff`):

````diff verified:applied with git apply to this worktree at 71c7e91f; go test ./internal/guard/ -run TestIntegrateChange passes with it and fails the new check alone with integratechange_cleanup.go untouched; go vet and gofmt -l clean
diff --git a/stats/internal/guard/integrate_change_test.go b/stats/internal/guard/integrate_change_test.go
index 37af78de..cf285b8f 100644
--- a/stats/internal/guard/integrate_change_test.go
+++ b/stats/internal/guard/integrate_change_test.go
@@ -298,7 +298,8 @@ check "the stage log runs preflight through refresh-main-checkout" \
   'head -1 "$WORK/stages.log" | grep -q "stage begin .*-stage flow.preflight" && tail -1 "$WORK/stages.log" | grep -q "stage end .*-stage flow.refresh-main-checkout -outcome completed"'
 check "run 2 keeps a's pre-reshape pin, locally and on origin" \
   'git -C "$WORK/a" rev-parse -q --verify refs/flow/pre-reshape/demo >/dev/null && git --git-dir="$WORK/a-origin.git" rev-parse -q --verify refs/flow/pre-reshape/demo >/dev/null'
-check "land ends on NEXT: Finished handoff" 'printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: print the Finished handoff"'`},
+check "land ends on NEXT: Finished handoff" 'printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: print the Finished handoff"'
+check "the NEXT line names the template and its never-omitted lines" 'printf "%s" "$OUT" | tail -1 | grep -qF "(skills/flow/cleanup.md) — its Learning: and Prompt tuning: lines included"'`},
 	})
 }
 
diff --git a/stats/internal/guard/integratechange_cleanup.go b/stats/internal/guard/integratechange_cleanup.go
index 95784f1a..aecd49ac 100644
--- a/stats/internal/guard/integratechange_cleanup.go
+++ b/stats/internal/guard/integratechange_cleanup.go
@@ -209,7 +209,7 @@ func (c *ic) cleanup() {
 		c.run("refresh-main-checkout", m, base)
 	}
 	c.markEnd("flow.refresh-main-checkout", "completed")
-	c.say("NEXT: print the Finished handoff")
+	c.say("NEXT: print the Finished handoff (skills/flow/cleanup.md) — its Learning: and Prompt tuning: lines included")
 }
 
 // bashIn is `(cd <dir> && bash -c <command>)`, both streams passed
````

  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run TestIntegrateChange`
    passes; `go vet ./...` and `gofmt -l .` are clean; `scripts/check-references.sh` exits 0.
  - [x] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/integratechange_cleanup.go`, `stats/internal/guard/integrate_change_test.go`
**Tests:** `TestIntegrateChangeHappyPath`
**Regression:** TestIntegrateChangeHappyPath — reverting the commit drops the template from the NEXT line: `the_NEXT_line_names_the_template_and_its_never-omitted_lines` fails.
**Baseline:** before=2 after=2
<!-- measured: grep -c '^func Test' stats/internal/guard/integrate_change_test.go @ 71c7e91f → 2, unchanged; the task adds a check inside an existing test -->
**Commit:** fix(guard): name the Finished handoff template in land's NEXT line
**After:** none

**Decision:** next-line-names-template

- [x] 2. Block a turn whose handoff drops its Learning or Prompt tuning line
  - [x] **Step 1: Reproduce.** Save the diff below to `/tmp/pf38-2.diff`, apply its test hunk alone
    (`git apply --include='stats/*' /tmp/pf38-2.diff`), run
    `cd stats && make -s web-dist && go test -count=1 ./internal/repotest/ -run TestFlowStopGuardHook`:
    subtests `p:_Finished_handoff_without_Prompt_tuning_blocks_naming_it` and
    `q:_flow-fast_report_without_Learning_blocks_naming_it` fail.
  - [x] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='stats/*' /tmp/pf38-2.diff`):

````diff verified:applied with git apply to this worktree at 0df205f5; go test -count=1 ./internal/repotest/ -run TestFlowStopGuardHook passes with it and fails subtests p and q with hooks/flow-stop-guard.py untouched; gofmt -l clean
diff --git a/hooks/flow-stop-guard.py b/hooks/flow-stop-guard.py
index 1ff84332..c426400d 100755
--- a/hooks/flow-stop-guard.py
+++ b/hooks/flow-stop-guard.py
@@ -16,6 +16,12 @@ It allows (exit 0, no output) at an AskUserQuestion, after an operator message s
 stage, when the session has no /flow run, and when stop_hook_active is true with nothing
 advanced since its last block (a fingerprint under ~/.claude/flow-stop-guard/).
 
+Before any of that, it reads the turn's last assistant text: a Finished handoff (a `## Finished`
+line, skills/flow/cleanup.md) or a /flow-fast landing report (recognised by carrying either line)
+that lacks `**Learning:**` or `**Prompt tuning:**` blocks with a reason naming the missing
+line(s) -- in any session, with no store lookup. With stop_hook_active true this check is skipped,
+so a turn that already got that block once is never trapped.
+
 Ticked/total counts are read from the change's tasks.md under the stage run's repoRoot:
 `flow tasks count` prints only the total and records a plan-growth observation as a side
 effect, so it is not used here.
@@ -42,6 +48,7 @@ _spec.loader.exec_module(_fac)
 TASK_RE = re.compile(r"^- \[([ xX])\] ")
 STOP_RE = re.compile(r"^\s*(stop|pause|wait)\b", re.IGNORECASE)
 TRANSCRIPT_TAIL_BYTES = 1 << 20
+HANDOFF_LINES = ("**Learning:**", "**Prompt tuning:**")
 
 
 def store_get(path: str, query: dict) -> dict:
@@ -90,13 +97,34 @@ def _texts(content) -> list:
     return []
 
 
-def transcript_allows(path) -> bool:
-    """True when the turn ends on an AskUserQuestion or the operator's last message starts with stop/pause/wait."""
+def transcript_entries(path) -> list:
     if not isinstance(path, str) or not os.path.isfile(path):
-        return False
+        return []
     with open(path, "rb") as fh:
         fh.seek(max(0, os.path.getsize(path) - TRANSCRIPT_TAIL_BYTES))
-        lines = fh.read().decode("utf-8", "replace").splitlines()
+        return fh.read().decode("utf-8", "replace").splitlines()
+
+
+def handoff_missing(lines: list) -> list:
+    """The never-omitted handoff lines the turn's last assistant text drops, if it is a handoff."""
+    last = ""
+    for line in lines:
+        try:
+            entry = json.loads(line)
+        except ValueError:
+            continue
+        if isinstance(entry, dict) and entry.get("type") == "assistant":
+            texts = _texts((entry.get("message") or {}).get("content"))
+            if any(t.strip() for t in texts):
+                last = "\n".join(texts)
+    present = [h for h in HANDOFF_LINES if h in last]
+    if not present and not re.search(r"^## Finished\s*$", last, re.MULTILINE):
+        return []
+    return [h for h in HANDOFF_LINES if h not in present]
+
+
+def transcript_allows(lines: list) -> bool:
+    """True when the turn ends on an AskUserQuestion or the operator's last message starts with stop/pause/wait."""
     last_user, last_tool = "", None
     for line in lines:
         try:
@@ -121,7 +149,21 @@ def main() -> int:
     payload = json.load(sys.stdin)
     if not isinstance(payload, dict):
         return 0
-    if transcript_allows(payload.get("transcript_path")):
+    lines = transcript_entries(payload.get("transcript_path"))
+    missing = [] if payload.get("stop_hook_active") else handoff_missing(lines)
+    if missing:
+        json.dump(
+            {
+                "decision": "block",
+                "reason": (
+                    f"flow-stop-guard: the handoff lacks its {' and '.join(missing)} line(s), which are "
+                    f"never omitted (skills/flow/cleanup.md). Reprint the handoff with them."
+                ),
+            },
+            sys.stdout,
+        )
+        return 0
+    if transcript_allows(lines):
         return 0
     session_id = _fac.resolve_session_id(payload)
     if not session_id:
diff --git a/stats/internal/repotest/flow_stop_guard_hook_test.go b/stats/internal/repotest/flow_stop_guard_hook_test.go
index 382fbf99..2cd2553c 100644
--- a/stats/internal/repotest/flow_stop_guard_hook_test.go
+++ b/stats/internal/repotest/flow_stop_guard_hook_test.go
@@ -28,6 +28,7 @@ type stopFixture struct {
 	lastUser string // last operator message
 	lastTool string // last assistant tool use
 	metaUser string // isMeta user entry after the operator message (hook feedback, skill text)
+	lastText string // last assistant text
 }
 
 func runStopGuard(t *testing.T, f stopFixture, home string, active bool) string {
@@ -87,6 +88,9 @@ func runStopGuard(t *testing.T, f stopFixture, home string, active bool) string
 	if f.lastTool != "" {
 		transcript += fmt.Sprintf(`{"type":"assistant","message":{"content":[{"type":"tool_use","name":%q,"input":{}}]}}`+"\n", f.lastTool)
 	}
+	if f.lastText != "" {
+		transcript += fmt.Sprintf(`{"type":"assistant","message":{"content":[{"type":"text","text":%q}]}}`+"\n", f.lastText)
+	}
 	tp := filepath.Join(t.TempDir(), "sess-1.jsonl")
 	if err := os.WriteFile(tp, []byte(transcript), 0o644); err != nil {
 		t.Fatal(err)
@@ -178,6 +182,22 @@ func TestFlowStopGuardHook(t *testing.T) {
 				allow(t, runStopGuard(t, f, t.TempDir(), false))
 			}
 		}},
+		{"p: Finished handoff without Prompt tuning blocks naming it", func(t *testing.T) {
+			f := stopFixture{state: "FINISHED", lastText: "## Finished\n\n**Change:** x\n**Learning:** /a/n.md\n"}
+			block(t, runStopGuard(t, f, t.TempDir(), false), "**Prompt tuning:**")
+		}},
+		{"q: flow-fast report without Learning blocks naming it", func(t *testing.T) {
+			f := stopFixture{state: "FINISHED", lastText: "Landed.\n**Prompt tuning:** NONE\n"}
+			block(t, runStopGuard(t, f, t.TempDir(), false), "**Learning:**")
+		}},
+		{"r: complete Finished handoff allows", func(t *testing.T) {
+			f := stopFixture{state: "FINISHED", lastText: "## Finished\n**Learning:** NONE — off\n**Prompt tuning:** NONE\n"}
+			allow(t, runStopGuard(t, f, t.TempDir(), false))
+		}},
+		{"s: incomplete handoff with stop_hook_active allows", func(t *testing.T) {
+			f := stopFixture{state: "FINISHED", lastText: "## Finished\n**Learning:** NONE — off\n"}
+			allow(t, runStopGuard(t, f, t.TempDir(), true))
+		}},
 		{"k: store 500 allows", func(t *testing.T) {
 			f := started(7, 7, true)
 			f.status = 500
````

  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go test -count=1 ./internal/repotest/ -run TestFlowStopGuardHook`
    passes; `go vet ./...` and `gofmt -l .` are clean; `scripts/check-references.sh` exits 0.
  - [x] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `hooks/flow-stop-guard.py`, `stats/internal/repotest/flow_stop_guard_hook_test.go`
**Tests:** `TestFlowStopGuardHook`
**Regression:** TestFlowStopGuardHook — reverting the hook change lets an incomplete handoff end the turn: subtests `p` and `q` fail.
**Baseline:** before=1 after=1
<!-- measured: grep -c '^func Test' stats/internal/repotest/flow_stop_guard_hook_test.go @ 0df205f5 → 1, unchanged; the task adds subtests inside an existing test -->
**Commit:** fix(hooks): block a turn whose handoff drops its Learning or Prompt tuning line
**After:** none

**Decision:** stop-hook-blocks-incomplete-handoff
