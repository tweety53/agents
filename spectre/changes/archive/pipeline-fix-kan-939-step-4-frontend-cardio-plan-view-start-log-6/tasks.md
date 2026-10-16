# pipeline-fix-kan-939-step-4-frontend-cardio-plan-view-start-log-6

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** make the pipeline-fix and self-review fix loops impossible to skip — every loop dispatch
admitted off its pairs row and recorded in flowd, every loop landing refused until flowd holds its
plan, fix and clean review, and the loops' fixed steps run by one script.

**Architecture:** three `flow-guard` subcommands in Go (`check-loop-dispatch`,
`check-loop-landing`, `pipeline-fix`) behind thin shims; a new `PreToolUse` hook and the existing
`protect-main-checkout.py` call the two checks; the two skill files cite the script.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:`
  trailer; never write `spectre/changes/`.
- Each task's change is the unified diff its step 3 gives, applied from the repository root with
  `git apply` — every diff was applied in plan order to a scratch clone at `a1bca81b` and checked
  there. A diff that does not apply is reported, never hand-merged by guesswork.
- Every prose line a diff adds or rewords is acknowledged verbatim in this change's
  `verbatim-moves.txt`; any other wording fails `scripts/check-verbatim-moves.sh`.
- Go in `stats/` follows the Go skills `CLAUDE.md` names; run `cd stats && gofmt -w .` before
  `gofmt -l .`. `go vet ./...` needs the built SPA (`internal/web/dist`, `## worktree setup`).
- Two other agents are editing `scripts/integrate-change.sh`, `scripts/test-integrate-change.sh`,
  `skills/flow-contracts/finish-contract-run1.md`, `skills/flow/verify-and-handoff.md`,
  `skills/flow/unfinished-work-gate.md`, `skills/flow/integrate.md` and
  `skills/flow-contracts/known-bugs.md` in other worktrees: no task touches them.

## Review Focus

- A `Key:` line indented or followed by trailing spaces still triggers the hook; a prompt that
  merely mentions a loop key mid-sentence does not.
- A self-review name that itself ends in a role-like segment (`kan-1-plan`) still resolves its
  key to the longest name with a valid suffix.
- flowd down: a loop dispatch and a loop landing are denied, never admitted on a journalled write
  read back as "no rows".
- `land` on a base that moved runs `verify` again and pushes nothing when it is red.
- The pairs table and the guard cannot drift: the Go test parses the real `SKILL.md`.

- [x] 1. Admit a fix-loop dispatch only off its pairs row, and record it
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run `flow record dispatches -C /Users/tweety53/Projects/agents -change pipeline-fix-kan-939-step-4-frontend-cardio-plan-view-start-log-5` and confirm `[]` — the defect: a loop's dispatches are recorded nowhere, and nothing checks the `subagent_type` or model a loop dispatch carries against the pairs of `skills/flow-self-review/SKILL.md` step 3.
  - [x] **Step 2: Write the failing test.** Save step 3's diff to `/tmp/pf6-task1.diff`, apply its `stats/internal/guard/check_loop_dispatch_test.go` hunk alone (`git apply --include='stats/internal/guard/check_loop_dispatch_test.go' /tmp/pf6-task1.diff`) and run `cd stats && go test ./internal/guard/ -run 'TestLoopTableIsTheSkillsOwn|TestCheckLoopDispatch' -count=1`: it FAILS to build with `undefined: lpSkill`.
  - [x] **Step 3: Implement.** Apply the rest of the diff (`git apply --exclude='stats/internal/guard/check_loop_dispatch_test.go' /tmp/pf6-task1.diff`) — the guard, its shim, the hook and its harness, the registration snippets and their setuptest pins, the README rows, and the pairs table:

````diff verified:applied with git apply in plan order to a scratch clone at a1bca81b; gofmt -l, go vet, go test ./internal/guard/ ./internal/setuptest/ -race and scripts/run-guard-tests.sh pass with every task applied; TestCheckLoopDispatch fails with the plan prerequisite removed (fix_without_a_plan)
diff --git a/README.md b/README.md
index f9966ef3..21f356b2 100644
--- a/README.md
+++ b/README.md
@@ -30,7 +30,7 @@ any project.
 | `commands-claude/` | Thin slash-command wrappers for Claude Code and ZCode |
 | `agents/` | Subagent definitions (`flow-low`, `flow-medium`, `flow-high`) |
 | `mods/` | Claude Code plugins. `flow-task-list` draws a band above the prompt in the task-list look: first a `⎿ <model-effort> <marker> main: <label> — <state>` line for the main agent (`in progress` while its turn runs, `waiting` between turns while a subagent runs), labelled with the top-level unit it is on — the latest open unit of its own `rules/be-brief.mdc` status lines naming a task of the running `/flow` change's plan, else the stage of the run's latest `flow stage begin`, else the latest open unit of its own status lines; then, nested under it as `  ⎿ <model-effort> <marker> Task <x>/<n> (<desc>) — <state> · <elapsed>`, one line per running subagent (`<desc>` alone, unnumbered, unless its description's `Task <x>/<n>` prefix names tasks of the running `/flow` change's plan; a group reads `Task 1+2+3/<n>`), its state `in progress`, `review`, `fix` or `re-review` read from its description; then one `◻ Task <x>/<n> (<title>) — pending` line per task of that plan neither ticked nor landed and named by no running line nor the `main` line, read from `<project>-worktrees/<change>/spectre/changes/<change>/tasks.md` beside the main checkout; a failed subagent's `✘ … — blocked` line among the running ones until the next main turn starts; and, only when no task is pending, the last two done subagent lines of the marked `/flow` change — with none marked, of the current main turn — above the running ones, in the room they leave (`✔` done and struck through, a group as its highest task, `Task 3/<n>`). Status lines are never lines of their own. Five lines at most under `main`. It adds the tracker key, `/flow` phase and the plan's done-over-total task count (`19/25`; a task is done when ticked or its `Task-Id:` commit is on the change branch since its base, `branch.<current>.flowBase` else `origin/HEAD`), then ` │ ` and a tally of the subagent rows, to the hint line under it |
-| `hooks/` | `enforce-agent-baseline.py` (denies a dispatch missing the baseline pointer), `protect-main-checkout.py` (denies edits on a main checkout's default branch), `flow-active-change.py` (turns a plain problem report into a fix run) |
+| `hooks/` | `enforce-agent-baseline.py` (denies a dispatch missing the baseline pointer), `enforce-loop-dispatch.py` (admits a fix-loop dispatch only off its pairs row, and records it), `protect-main-checkout.py` (denies edits on a main checkout's default branch), `flow-active-change.py` (turns a plain problem report into a fix run) |
 | `scripts/` | The guards `/flow` runs, each with its `test-*.sh` harness |
 | `stats/` | `flowd` — the PostgreSQL-backed service holding pipeline state and per-stage telemetry, with a web UI. See `stats/README.md` |
 | `spectre/` | This repository's own specs and changes |
@@ -190,6 +190,9 @@ return a verdict. Each one has a `test-*.sh` harness in `scripts/`, and
 **Always on, in every session (hooks)**
 - `enforce-agent-baseline.py`: blocks any subagent dispatch whose prompt lacks the
   `agent-baseline.md` pointer, so the rules reach every subagent, however deep.
+- `enforce-loop-dispatch.py`: blocks a pipeline-fix or self-review loop dispatch off its pairs
+  row, a fix with no completed plan, or one while an earlier dispatch is open, and records the
+  admitted one in flowd (`scripts/check-loop-dispatch.sh`).
 - `protect-main-checkout.py`: blocks edits, staging, commits, stashes and resets on a main
   checkout's default branch. Work happens in a worktree.
 - `flow-active-change.py`: routes a plain problem report to the session's open change as a fix run.
diff --git a/hooks/enforce-loop-dispatch.py b/hooks/enforce-loop-dispatch.py
new file mode 100644
index 00000000..5bfe4ad2
--- /dev/null
+++ b/hooks/enforce-loop-dispatch.py
@@ -0,0 +1,82 @@
+#!/usr/bin/env python3
+"""PreToolUse hook: admit a pipeline-fix or self-review fix-loop dispatch only off its row.
+
+A loop dispatch is an Agent/Task call whose prompt carries a `Key: pipeline-fix-<k>-...` or
+`Key: self-review-...` line (skills/flow-contracts/pipeline.md **Pipeline defects found
+mid-run**, skills/flow-self-review/SKILL.md step 3). The loops were prose only, and a parent
+skipped their steps: fixers dispatched on flow-high without a planner, fixes landed with no
+review. This hook hands every such call to scripts/check-loop-dispatch.sh in the agents
+checkout it is installed from, which refuses a pair off the canonical table, a fix with no
+completed plan, a dispatch while an earlier one is still open -- and records the admitted
+dispatch in flowd, so check-loop-landing.sh can refuse a landing those records do not cover.
+
+The PreToolUse contract cannot rewrite tool input, so a refusal is a deny whose reason is the
+checker's own lines.
+
+Fails open for anything that is not a loop dispatch -- unparseable input, another tool, a
+prompt with no loop key. Fails CLOSED for a loop dispatch: a checker that cannot answer, or
+cannot run, denies it, because an unchecked loop dispatch is the skip this hook exists to stop.
+"""
+
+import json
+import os
+import re
+import subprocess
+import sys
+
+DISPATCH_TOOLS = {"Agent", "Task"}
+
+KEY_RE = re.compile(r"^[ \t]*Key:[ \t]*((?:pipeline-fix|self-review)-\S+)[ \t]*$", re.M)
+
+REPO = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))
+
+# The checker the hook runs; overridable for the hook's own harness only.
+CHECK = os.environ.get("LOOP_DISPATCH_CHECK") or os.path.join(REPO, "scripts", "check-loop-dispatch.sh")
+
+
+def deny(reason):
+    json.dump(
+        {
+            "hookSpecificOutput": {
+                "hookEventName": "PreToolUse",
+                "permissionDecision": "deny",
+                "permissionDecisionReason": reason,
+            }
+        },
+        sys.stdout,
+    )
+    return 0
+
+
+def main():
+    try:
+        payload = json.load(sys.stdin)
+    except (json.JSONDecodeError, ValueError):
+        return 0
+    if not isinstance(payload, dict) or payload.get("tool_name") not in DISPATCH_TOOLS:
+        return 0
+    tool_input = payload.get("tool_input")
+    if not isinstance(tool_input, dict) or not isinstance(tool_input.get("prompt"), str):
+        return 0
+    m = KEY_RE.search(tool_input["prompt"])
+    if not m:
+        return 0
+    key = m.group(1)
+    args = [CHECK, REPO, key, str(tool_input.get("subagent_type") or ""), str(tool_input.get("model") or "")]
+    try:
+        r = subprocess.run(args, capture_output=True, text=True, timeout=300)
+    except (OSError, subprocess.SubprocessError) as e:
+        return deny(f"Loop dispatch {key} blocked: check-loop-dispatch.sh could not run ({e}).")
+    if r.returncode == 0:
+        return 0
+    return deny(
+        f"Loop dispatch {key} blocked (check-loop-dispatch.sh exit {r.returncode}):\n"
+        f"{r.stdout}{r.stderr}"
+    )
+
+
+if __name__ == "__main__":
+    try:
+        sys.exit(main())
+    except Exception:  # only reached before a loop key is known: fail open
+        sys.exit(0)
diff --git a/scripts/check-loop-dispatch.sh b/scripts/check-loop-dispatch.sh
new file mode 100755
index 00000000..b87c8006
--- /dev/null
+++ b/scripts/check-loop-dispatch.sh
@@ -0,0 +1,38 @@
+#!/usr/bin/env bash
+# check-loop-dispatch.sh <agents-repo> <key> <subagent_type> <model> — admit
+# one dispatch of a pipeline-fix or self-review fix loop, or refuse it, and
+# record it (hooks/enforce-loop-dispatch.py runs it on every Agent/Task call
+# whose prompt carries a `Key: pipeline-fix-<k>-…` or `Key: self-review-…`
+# line).
+#
+# Refused (exit 1, one LOOP-DISPATCH-REFUSED line):
+#   KEY          no row of the pairs table in
+#                skills/flow-self-review/SKILL.md — the one place the pairs
+#                are read from — covers the key's suffix;
+#   NOT-STARTED  a pipeline-fix key with no pipeline-fix-<name>-<k> branch
+#                (scripts/pipeline-fix.sh start makes it);
+#   PAIR         subagent_type or model off the suffix's row;
+#   ESCALATION   the row's escalation type with no recorded escalation
+#                (scripts/pipeline-fix.sh escalate);
+#   KEY-REUSED   the key already has a row;
+#   OPEN         an earlier row of the loop is not closed
+#                (scripts/pipeline-fix.sh end);
+#   PLAN         a fix or review with no completed plan row.
+# Admitted (exit 0, LOOP-DISPATCH-OK): the row is opened with `flow record
+# dispatch begin` under the loop's change, its session token the loop's
+# branch name — except a `pass`, which is pair-checked and not recorded.
+# LOOP-DISPATCH-NA, exit 0: the key is no loop key.
+# Exit 2: cannot answer — the table unreadable, flowd unreadable, the begin
+# refused (a loop change no checkout carries: pipeline-fix.sh start), usage.
+#
+# Go, in flow-guard (stats/internal/guard/loop.go). flow-guard is built from
+# this checkout, never taken from PATH: scripts/lib/flow-guard.sh derives it,
+# and exits 2 with the cause when it cannot.
+set -euo pipefail
+# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
+SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
+[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
+  echo "check-loop-dispatch: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
+  exit 2
+}
+flow_guard_exec check-loop-dispatch 2 "check-loop-dispatch:" "$@"
diff --git a/scripts/test-enforce-loop-dispatch-hook.sh b/scripts/test-enforce-loop-dispatch-hook.sh
new file mode 100755
index 00000000..d0f59444
--- /dev/null
+++ b/scripts/test-enforce-loop-dispatch-hook.sh
@@ -0,0 +1,89 @@
+#!/usr/bin/env bash
+# Assertion harness for hooks/enforce-loop-dispatch.py. Points the hook at a stub
+# checker through LOOP_DISPATCH_CHECK — the guard itself is covered by
+# stats/internal/guard/check_loop_dispatch_test.go — and pipes PreToolUse payloads to
+# it, asserting the decision and what the checker was handed. Cases:
+#
+#   1. a dispatch with no loop key                    -> allow, checker never run
+#   2. a loop key the checker admits (exit 0)          -> allow, checker got
+#                                                         <repo> <key> <type> <model>
+#   3. a loop key the checker refuses (exit 1)         -> deny, its lines in the reason
+#   4. a loop key the checker cannot answer (exit 2)   -> deny
+#   5. a loop key whose checker cannot run             -> deny (fails closed)
+#   6. malformed JSON                                  -> allow, exit 0
+#   7. a Bash call carrying a loop key                 -> allow, checker never run
+set -euo pipefail
+
+SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
+REPO="$(cd "$SCRIPT_DIR/.." && pwd -P)"
+HOOK="$REPO/hooks/enforce-loop-dispatch.py"
+FAILURES=0
+
+fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
+pass() { printf 'ok: %s\n' "$1"; }
+
+SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/test-enforce-loop-dispatch-hook.XXXXXX")"
+trap 'rm -rf "$SANDBOX"' EXIT
+LOG="$SANDBOX/args"
+RC_FILE="$SANDBOX/rc"
+cat >"$SANDBOX/check" <<STUB
+#!/usr/bin/env bash
+printf '%s\n' "\$@" >"$LOG"
+echo "LOOP-DISPATCH-REFUSED: PAIR — stub"
+exit "\$(cat "$RC_FILE")"
+STUB
+chmod +x "$SANDBOX/check"
+
+payload() { # <tool> <prompt>
+  python3 -c 'import json,sys; print(json.dumps({"tool_name": sys.argv[1], "tool_input": {"prompt": sys.argv[2], "subagent_type": "flow-low", "model": "opus"}}))' "$1" "$2"
+}
+hook() { # <check> <stdin>
+  OUT="$(printf '%s' "$2" | LOOP_DISPATCH_CHECK="$1" python3 "$HOOK")" && RC=0 || RC=$?
+}
+expect() { # <case> <allow|deny> <needle>
+  if [ "$RC" -ne 0 ]; then fail "$1: exit $RC"; return; fi
+  case "$2" in
+    allow) [ -z "$OUT" ] && pass "$1" || fail "$1: expected allow, got $OUT" ;;
+    deny) if grep -q '"permissionDecision": "deny"' <<<"$OUT" && grep -qF -- "$3" <<<"$OUT"; then
+            pass "$1"
+          else
+            fail "$1: expected a deny carrying '$3', got $OUT"
+          fi ;;
+  esac
+}
+
+LOOP_PROMPT=$'Before anything else.\nKey: pipeline-fix-6-fix\nDo it.'
+
+rm -f "$LOG"; echo 0 >"$RC_FILE"
+hook "$SANDBOX/check" "$(payload Agent $'Key: task-3-implementer\nwork')"
+expect "1 no loop key" allow
+[ ! -e "$LOG" ] && pass "1 checker not run" || fail "1 checker ran"
+
+hook "$SANDBOX/check" "$(payload Agent "$LOOP_PROMPT")"
+expect "2 admitted" allow
+[ "$(tr '\n' ' ' <"$LOG")" = "$REPO pipeline-fix-6-fix flow-low opus " ] \
+  && pass "2 checker args" || fail "2 checker args: $(tr '\n' ' ' <"$LOG")"
+
+echo 1 >"$RC_FILE"
+hook "$SANDBOX/check" "$(payload Task "$LOOP_PROMPT")"
+expect "3 refused" deny "LOOP-DISPATCH-REFUSED: PAIR"
+
+echo 2 >"$RC_FILE"
+hook "$SANDBOX/check" "$(payload Agent "$LOOP_PROMPT")"
+expect "4 cannot answer" deny "exit 2"
+
+hook "$SANDBOX/missing" "$(payload Agent "$LOOP_PROMPT")"
+expect "5 checker cannot run" deny "could not run"
+
+hook "$SANDBOX/check" "not json"
+expect "6 malformed JSON" allow
+
+rm -f "$LOG"; echo 1 >"$RC_FILE"
+hook "$SANDBOX/check" "$(python3 -c 'import json; print(json.dumps({"tool_name": "Bash", "tool_input": {"command": "echo", "prompt": "Key: pipeline-fix-6-fix"}}))')"
+expect "7 not a dispatch" allow
+[ ! -e "$LOG" ] && pass "7 checker not run" || fail "7 checker ran"
+
+if [ "$FAILURES" -gt 0 ]; then
+  printf '%d case(s) failed\n' "$FAILURES" >&2
+  exit 1
+fi
diff --git a/setup.sh b/setup.sh
index 12e4eb52..196ae494 100755
--- a/setup.sh
+++ b/setup.sh
@@ -436,6 +436,15 @@ install_hooks_zcode() {
       ]
     }
   }
+SNIPPET
+  warn_unregistered_hook "$config" enforce-loop-dispatch enforce-loop-dispatch \
+    "  ⚠ The enforce-loop-dispatch hook is installed but NOT registered, so nothing yet admits" \
+    "    a pipeline-fix or self-review loop dispatch off its pairs row, or records it. Add" \
+    "    this entry to \"PreToolUse\" under \"hooks\".\"events\" in $config:" <<'SNIPPET'
+
+        { "matcher": "Agent|Task", "hooks": [
+          { "type": "command", "command": "python3 \"$HOME/.zcode/hooks/enforce-loop-dispatch.py\"" }
+        ] }
 SNIPPET
   warn_unregistered_hook "$config" flow-active-change flow-active-change \
     "  ⚠ The flow-active-change hook is installed but NOT registered, so a plain problem" \
@@ -543,6 +552,18 @@ install_hooks() {
         ]
       }
     ]
+SNIPPET
+  warn_unregistered_hook "$settings" enforce-loop-dispatch enforce-loop-dispatch \
+    "  ⚠ The enforce-loop-dispatch hook is installed but NOT registered, so nothing yet admits" \
+    "    a pipeline-fix or self-review loop dispatch off its pairs row, or records it." \
+    "    Add this entry to \"PreToolUse\" in $settings:" <<'SNIPPET'
+
+      {
+        "matcher": "Agent|Task",
+        "hooks": [
+          { "type": "command", "command": "python3 \"$HOME/.claude/hooks/enforce-loop-dispatch.py\"", "timeout": 300 }
+        ]
+      }
 SNIPPET
   warn_unregistered_hook "$settings" flow-active-change flow-active-change \
     "  ⚠ The flow-active-change hook is installed but NOT registered, so a plain problem" \
diff --git a/skills/flow-self-review/SKILL.md b/skills/flow-self-review/SKILL.md
index a1c518c9..373dcacf 100644
--- a/skills/flow-self-review/SKILL.md
+++ b/skills/flow-self-review/SKILL.md
@@ -131,11 +131,20 @@ handshake** there applying to each reply. Every dispatch is one-shot: no `SendMe
 it returns. The no-progress escalation of **Fewest operator actions**
 (`skills/flow-contracts/pipeline.md`) still runs on `flow-high`.
 
-**The pairs are this file's own**, and **Pipeline defects found mid-run** cites them: **the planner
-pair** is `subagent_type: flow-medium` on `opus`; **the fixer pair** is `flow-low` on `opus`;
-**the review pair** is `opus` for the first review (`<r>` = 1), on `flow-medium` or `flow-low` as
-the planner recorded — `flow-medium` when it recorded none — and `flow-low` on `sonnet` for
-every re-review after a fix.
+**The pairs are this file's own**, and **Pipeline defects found mid-run** cites them — one row per
+key suffix, the table `check-loop-dispatch.sh` reads and refuses a loop dispatch off:
+
+| Key suffix | `subagent_type` | Model | On a recorded escalation |
+|---|---|---|---|
+| `pass` | `flow-medium` | `opus` | — |
+| `plan`, `plan-<n>` | `flow-medium` | `opus` | `flow-high` |
+| `fix`, `fix-<r>`, `fix-g<g>`, `fix-v<v>` | `flow-low` | `opus` | `flow-high` |
+| `review-1` | `flow-medium`, `flow-low` | `opus` | — |
+| `review-<r>` | `flow-low` | `sonnet` | — |
+
+**The planner pair** is the `plan` row, **the fixer pair** the `fix` row and **the review pair**
+the two `review` rows: the first review on `flow-medium` or `flow-low` as the planner recorded —
+`flow-medium` when it recorded none — and every re-review after a fix on `flow-low` and `sonnet`.
 
 `<agents-base>` is the one **Pipeline defects found mid-run** binds. `<fix-branch>` and
 `<fix-base>` are where the fixes go (`design.md`, `fixes-ride-the-change`): inside a run whose
diff --git a/stats/internal/guard/check_loop_dispatch_test.go b/stats/internal/guard/check_loop_dispatch_test.go
new file mode 100644
index 00000000..812576bd
--- /dev/null
+++ b/stats/internal/guard/check_loop_dispatch_test.go
@@ -0,0 +1,157 @@
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
+// lpFx is one loop fixture: an agents repo carrying the real pairs table, a
+// `flow` stub on PATH that answers `record dispatches` from rows and logs
+// every call, and the guard's Env.
+type lpFx struct {
+	repo, log, rows string
+	env             Env
+}
+
+// lpNew builds the fixture with the given branches; rows is the JSON
+// `flow record dispatches` prints ("" makes it fail).
+func lpNew(t *testing.T, rows string, branches ...string) lpFx {
+	t.Helper()
+	tmp, err := filepath.EvalSymlinks(t.TempDir())
+	if err != nil {
+		t.Fatal(err)
+	}
+	fx := lpFx{repo: tmp + "/agents", log: tmp + "/flow.log", rows: tmp + "/rows.json"}
+	skill, err := os.ReadFile("../../../" + lpSkill)
+	if err != nil {
+		t.Fatal(err)
+	}
+	writeFile(t, fx.repo+"/"+lpSkill, string(skill))
+	gitRun(t, "", "init", "-q", "-b", "main", fx.repo)
+	gitRun(t, fx.repo, "add", "-A")
+	gitRun(t, fx.repo, "commit", "-q", "-m", "init")
+	for _, b := range branches {
+		gitRun(t, fx.repo, "branch", b)
+	}
+	if rows != "" {
+		writeFile(t, fx.rows, rows)
+	}
+	bin := tmp + "/bin"
+	writeExec(t, bin+"/flow", `#!/usr/bin/env bash
+printf '%s\n' "$*" >>`+fx.log+`
+case "$1 $2" in
+  "record dispatches") [ -f `+fx.rows+` ] && exec cat `+fx.rows+`; exit 1 ;;
+  "record dispatch") echo "recorded: dispatch 1" ;;
+esac
+`)
+	fx.env = Env{Getenv: pathEnv(bin), Dir: tmp}
+	return fx
+}
+
+func (fx lpFx) run(t *testing.T, args ...string) (int, string) {
+	t.Helper()
+	var o, e bytes.Buffer
+	rc := checkLoopDispatch(append([]string{fx.repo}, args...), fx.env, &o, &e)
+	return rc, o.String() + e.String()
+}
+
+func (fx lpFx) calls(t *testing.T) string {
+	t.Helper()
+	b, _ := os.ReadFile(fx.log)
+	return string(b)
+}
+
+const (
+	lpBranch   = "pipeline-fix-kan-1-demo-6"
+	lpPlanDone = `[{"seq":1,"key":"pipeline-fix-6-plan","outcome":"completed","sessionToken":"pipeline-fix-kan-1-demo-6","endedAt":"2026-10-16T10:00:00Z"}]`
+	lpPlanOpen = `[{"seq":1,"key":"pipeline-fix-6-plan","sessionToken":"pipeline-fix-kan-1-demo-6"}]`
+)
+
+func TestLoopTableIsTheSkillsOwn(t *testing.T) {
+	t.Parallel()
+	rows, err := lpTable("../../..")
+	if err != nil {
+		t.Fatal(err)
+	}
+	want := map[string]string{"pass": "flow-medium opus", "plan-2": "flow-medium opus", "fix-g1": "flow-low opus",
+		"fix-v3": "flow-low opus", "review-1": "flow-medium flow-low opus", "review-2": "flow-low sonnet"}
+	for suffix, pair := range want {
+		got := ""
+		for _, r := range rows {
+			if r.suffix.MatchString(suffix) {
+				got = strings.Join(append(append([]string{}, r.types...), r.model), " ")
+				break
+			}
+		}
+		if got != pair {
+			t.Errorf("suffix %s: pair %q, want %q", suffix, got, pair)
+		}
+	}
+}
+
+func TestCheckLoopDispatch(t *testing.T) {
+	t.Parallel()
+	for _, tc := range []struct {
+		name, rows, key, typ, model string
+		branches, escalated         []string
+		rc                          int
+		out, begin                  string // begin "" asserts no begin call
+	}{
+		{name: "plan admitted and recorded", rows: `[]`, key: "pipeline-fix-6-plan", typ: "flow-medium", model: "opus", branches: []string{lpBranch},
+			out: "LOOP-DISPATCH-OK", begin: "-change " + lpBranch + " -role planner -model opus -effort medium -key pipeline-fix-6-plan -session-token " + lpBranch},
+		{name: "fix without a plan", rows: `[]`, key: "pipeline-fix-6-fix", typ: "flow-low", model: "opus", branches: []string{lpBranch}, rc: 1, out: "PLAN —"},
+		{name: "fix after an open plan", rows: lpPlanOpen, key: "pipeline-fix-6-fix", typ: "flow-low", model: "opus", branches: []string{lpBranch}, rc: 1, out: "OPEN — pipeline-fix-6-plan"},
+		{name: "fix after a completed plan", rows: lpPlanDone, key: "pipeline-fix-6-fix", typ: "flow-low", model: "opus", branches: []string{lpBranch},
+			out: "LOOP-DISPATCH-OK", begin: "-role implementer -model opus -effort low -key pipeline-fix-6-fix"},
+		{name: "fix on flow-high unescalated", rows: lpPlanDone, key: "pipeline-fix-6-fix", typ: "flow-high", model: "opus", branches: []string{lpBranch}, rc: 1, out: "ESCALATION —"},
+		{name: "fix on flow-high escalated", rows: lpPlanDone, key: "pipeline-fix-6-fix", typ: "flow-high", model: "opus", branches: []string{lpBranch},
+			escalated: []string{"pipeline-fix-6-fix"}, out: "LOOP-DISPATCH-OK", begin: "-effort high"},
+		{name: "fix on sonnet", rows: lpPlanDone, key: "pipeline-fix-6-fix", typ: "flow-low", model: "sonnet", branches: []string{lpBranch}, rc: 1, out: "PAIR —"},
+		{name: "re-review on opus", rows: lpPlanDone, key: "pipeline-fix-6-review-2", typ: "flow-low", model: "opus", branches: []string{lpBranch}, rc: 1, out: "PAIR —"},
+		{name: "first review on flow-low", rows: lpPlanDone, key: "pipeline-fix-6-review-1", typ: "flow-low", model: "opus", branches: []string{lpBranch},
+			out: "LOOP-DISPATCH-OK", begin: "-role reviewer"},
+		{name: "a reused key", rows: lpPlanDone, key: "pipeline-fix-6-plan", typ: "flow-medium", model: "opus", branches: []string{lpBranch}, rc: 1, out: "KEY-REUSED —"},
+		{name: "no loop branch", rows: `[]`, key: "pipeline-fix-6-plan", typ: "flow-medium", model: "opus", rc: 1, out: "NOT-STARTED —"},
+		{name: "an unknown suffix", rows: `[]`, key: "pipeline-fix-6-implementer", typ: "flow-low", model: "opus", branches: []string{lpBranch}, rc: 1, out: "KEY —"},
+		{name: "self-review on its own branch files under the change", rows: `[]`, key: "self-review-kan-1-demo-plan", typ: "flow-medium", model: "opus",
+			out: "LOOP-DISPATCH-OK", begin: "-change kan-1-demo -role planner"},
+		{name: "self-review loop files under its branch", rows: `[]`, key: "self-review-kan-1-demo-plan", typ: "flow-medium", model: "opus", branches: []string{"self-review-kan-1-demo"},
+			out: "LOOP-DISPATCH-OK", begin: "-change self-review-kan-1-demo -role planner -model opus -effort medium -key self-review-kan-1-demo-plan -session-token self-review-kan-1-demo"},
+		{name: "the pass is checked, never recorded", rows: `[]`, key: "self-review-kan-1-demo-pass", typ: "flow-medium", model: "opus", out: "LOOP-DISPATCH-OK"},
+		{name: "the pass off its pair", rows: `[]`, key: "self-review-kan-1-demo-pass", typ: "flow-low", model: "opus", rc: 1, out: "PAIR —"},
+		{name: "flowd unreadable", rows: "", key: "pipeline-fix-6-fix", typ: "flow-low", model: "opus", branches: []string{lpBranch}, rc: 2, out: "cannot answer"},
+	} {
+		t.Run(tc.name, func(t *testing.T) {
+			t.Parallel()
+			fx := lpNew(t, tc.rows, tc.branches...)
+			for _, k := range tc.escalated {
+				gitRun(t, fx.repo, "config", "--add", "flow.escalation", k)
+			}
+			rc, out := fx.run(t, tc.key, tc.typ, tc.model)
+			if rc != tc.rc || !strings.Contains(out, tc.out) {
+				t.Fatalf("rc %d, want %d; output %q, want it to carry %q", rc, tc.rc, out, tc.out)
+			}
+			calls := fx.calls(t)
+			if tc.begin == "" && strings.Contains(calls, "dispatch begin") {
+				t.Fatalf("a begin was recorded:\n%s", calls)
+			}
+			if tc.begin != "" && !strings.Contains(calls, tc.begin) {
+				t.Fatalf("flow calls %q carry no %q", calls, tc.begin)
+			}
+		})
+	}
+}
+
+func TestCheckLoopDispatchUnreadableTable(t *testing.T) {
+	t.Parallel()
+	fx := lpNew(t, `[]`, lpBranch)
+	if err := os.Remove(fx.repo + "/" + lpSkill); err != nil {
+		t.Fatal(err)
+	}
+	if rc, out := fx.run(t, "pipeline-fix-6-plan", "flow-medium", "opus"); rc != 2 {
+		t.Fatalf("rc %d, want 2: %s", rc, out)
+	}
+}
diff --git a/stats/internal/guard/loop.go b/stats/internal/guard/loop.go
new file mode 100644
index 00000000..99b89676
--- /dev/null
+++ b/stats/internal/guard/loop.go
@@ -0,0 +1,319 @@
+package guard
+
+import (
+	"encoding/json"
+	"fmt"
+	"io"
+	"os"
+	"path/filepath"
+	"regexp"
+	"slices"
+	"strings"
+)
+
+// The pipeline-fix and self-review fix loops (skills/flow-contracts/pipeline.md
+// **Pipeline defects found mid-run**, skills/flow-self-review/SKILL.md step 3)
+// as mechanics rather than prose: check-loop-dispatch.sh admits each loop
+// dispatch and records it in flowd, check-loop-landing.sh refuses a landing
+// those records do not cover, and pipeline-fix.sh runs the loop's fixed
+// steps. A loop is its own change in <agents repo>'s flowd project, named
+// after its branch -- pipeline-fix-<name>-<k> or self-review-<name> --
+// except a self-review on the change's own branch, which has no branch of its
+// own and files its rows under <name> itself. Either way the rows carry the
+// loop's branch name as their session token, so one loop's rows never mix
+// with the run's own.
+func init() {
+	Registry["check-loop-dispatch"] = checkLoopDispatch
+}
+
+// lpRow is one row of the pairs table in skills/flow-self-review/SKILL.md:
+// the key suffixes it covers, the subagent types and model it runs on, and
+// the subagent types a recorded no-progress escalation also admits.
+type lpRow struct {
+	suffix *regexp.Regexp
+	types  []string
+	model  string
+	escal  []string
+}
+
+// lpSkill is the one canonical place the pairs are read from.
+const lpSkill = "skills/flow-self-review/SKILL.md"
+
+var (
+	lpTick        = regexp.MustCompile("`([^`]+)`")
+	lpPlaceholder = regexp.MustCompile(`<[a-z]+>`)
+	lpPipelineKey = regexp.MustCompile(`^pipeline-fix-([0-9]+)-(.+)$`)
+	lpDigits      = regexp.MustCompile(`^[0-9]+$`)
+)
+
+// lpTable parses the table headed `| Key suffix |` out of <repo>/lpSkill.
+// Each cell's backticked tokens are its values; `<x>` in a suffix stands for
+// a number. A file with no such table, or a row missing a cell, is an error:
+// a loop dispatch is never admitted off a table that cannot be read.
+func lpTable(repo string) ([]lpRow, error) {
+	b, err := os.ReadFile(filepath.Join(repo, lpSkill))
+	if err != nil {
+		return nil, err
+	}
+	var rows []lpRow
+	in := false
+	for _, line := range strings.Split(string(b), "\n") {
+		t := strings.TrimSpace(line)
+		if strings.HasPrefix(t, "| Key suffix |") {
+			in = true
+			continue
+		}
+		if !in {
+			continue
+		}
+		if !strings.HasPrefix(t, "|") {
+			break
+		}
+		if strings.HasPrefix(t, "|---") {
+			continue
+		}
+		cells := strings.Split(strings.Trim(t, "|"), "|")
+		if len(cells) != 4 {
+			return nil, fmt.Errorf("%s: pairs row %q has %d cells, want 4", lpSkill, t, len(cells))
+		}
+		ticks := func(c string) []string {
+			var out []string
+			for _, m := range lpTick.FindAllStringSubmatch(c, -1) {
+				out = append(out, m[1])
+			}
+			return out
+		}
+		var alts []string
+		for _, s := range ticks(cells[0]) {
+			alts = append(alts, lpPlaceholder.ReplaceAllString(regexp.QuoteMeta(s), "[0-9]+"))
+		}
+		model := ticks(cells[2])
+		if len(alts) == 0 || len(ticks(cells[1])) == 0 || len(model) != 1 {
+			return nil, fmt.Errorf("%s: pairs row %q names no suffix, type or single model", lpSkill, t)
+		}
+		rows = append(rows, lpRow{
+			suffix: regexp.MustCompile("^(?:" + strings.Join(alts, "|") + ")$"),
+			types:  ticks(cells[1]),
+			model:  model[0],
+			escal:  ticks(cells[3]),
+		})
+	}
+	if len(rows) == 0 {
+		return nil, fmt.Errorf("%s carries no `| Key suffix |` pairs table", lpSkill)
+	}
+	return rows, nil
+}
+
+// lpLoop names one loop's records: the change they are filed under, the
+// session token (the loop's branch name) and the key prefix of its rows.
+type lpLoop struct {
+	change, token, prefix string
+}
+
+// lpMain is <repo>'s main checkout, the -C every flow call runs with.
+func lpMain(env Env, git, repo string) (string, bool) {
+	rc, out := kwRun(env, "", git, "-C", repo, "rev-parse", "--path-format=absolute", "--git-common-dir")
+	if rc != 0 {
+		return "", false
+	}
+	return filepath.Dir(strings.TrimSpace(out)), true
+}
+
+// lpBranchExists is whether refs/heads/<b> exists in repo.
+func lpBranchExists(env Env, git, repo, b string) bool {
+	rc, _ := kwRun(env, "", git, "-C", repo, "rev-parse", "-q", "--verify", "refs/heads/"+b)
+	return rc == 0
+}
+
+// lpFromKey splits key into its loop and its suffix, and finds the suffix's
+// pairs row. ok false with a message is a refusal (exit 1); ok false with
+// msg "" means the key is no loop key at all.
+func lpFromKey(env Env, git, repo, key string, rows []lpRow) (lpLoop, string, lpRow, string, bool) {
+	rowOf := func(s string) (lpRow, bool) {
+		for _, r := range rows {
+			if r.suffix.MatchString(s) {
+				return r, true
+			}
+		}
+		return lpRow{}, false
+	}
+	if m := lpPipelineKey.FindStringSubmatch(key); m != nil {
+		k, suffix := m[1], m[2]
+		row, ok := rowOf(suffix)
+		if !ok {
+			return lpLoop{}, "", lpRow{}, fmt.Sprintf("KEY — %s: no row of the pairs table (%s) covers suffix %q", key, lpSkill, suffix), false
+		}
+		_, out := kwRun(env, "", git, "-C", repo, "for-each-ref", "--format=%(refname:short)", "refs/heads/")
+		var hits []string
+		for _, b := range strings.Fields(out) {
+			if strings.HasPrefix(b, "pipeline-fix-") && strings.HasSuffix(b, "-"+k) && plainChangeName(b) {
+				hits = append(hits, b)
+			}
+		}
+		if len(hits) != 1 {
+			return lpLoop{}, "", lpRow{}, fmt.Sprintf("NOT-STARTED — %s: %d branches pipeline-fix-<name>-%s; start the loop with scripts/pipeline-fix.sh start <name> %s", key, len(hits), k, k), false
+		}
+		return lpLoop{change: hits[0], token: hits[0], prefix: "pipeline-fix-" + k + "-"}, suffix, row, "", true
+	}
+	if !strings.HasPrefix(key, "self-review-") {
+		return lpLoop{}, "", lpRow{}, "", false
+	}
+	rest := strings.TrimPrefix(key, "self-review-")
+	// The longest name wins: a role suffix is the short tail.
+	for i := len(rest) - 1; i > 0; i-- {
+		if rest[i] != '-' {
+			continue
+		}
+		name, suffix := rest[:i], rest[i+1:]
+		row, ok := rowOf(suffix)
+		if !ok || !plainChangeName(name) {
+			continue
+		}
+		loop := lpLoop{token: "self-review-" + name, prefix: "self-review-" + name + "-", change: name}
+		if lpBranchExists(env, git, repo, loop.token) {
+			loop.change = loop.token
+		}
+		return loop, suffix, row, "", true
+	}
+	return lpLoop{}, "", lpRow{}, fmt.Sprintf("KEY — %s: no row of the pairs table (%s) covers its suffix", key, lpSkill), false
+}
+
+// lpDispatch is one `flow record dispatches` row, the fields the loop reads.
+type lpDispatch struct {
+	Seq          int     `json:"seq"`
+	Key          string  `json:"key"`
+	Outcome      string  `json:"outcome"`
+	SessionToken string  `json:"sessionToken"`
+	EndedAt      *string `json:"endedAt"`
+}
+
+// lpRows is the loop's own rows, in seq order. ok false: flowd could not be
+// read, which is never "no rows".
+func lpRows(env Env, main string, lp lpLoop) ([]lpDispatch, bool) {
+	out, _, rc := pcFlow(env, false, "record", "dispatches", "-change", lp.change, "-C", main)
+	if rc != 0 {
+		return nil, false
+	}
+	var all []lpDispatch
+	if json.Unmarshal(out, &all) != nil {
+		return nil, false
+	}
+	var mine []lpDispatch
+	for _, d := range all {
+		if d.SessionToken == lp.token && strings.HasPrefix(d.Key, lp.prefix) {
+			mine = append(mine, d)
+		}
+	}
+	slices.SortFunc(mine, func(a, b lpDispatch) int { return a.Seq - b.Seq })
+	return mine, true
+}
+
+// lpFamily is a suffix's place in the loop: plan, fix (the first fixers,
+// `fix` and `fix-g<g>`), fix-r (a review's fix), fix-v (a red line's fix),
+// review or pass.
+func lpFamily(suffix string) string {
+	switch {
+	case suffix == "pass":
+		return "pass"
+	case suffix == "plan" || strings.HasPrefix(suffix, "plan-"):
+		return "plan"
+	case suffix == "fix" || strings.HasPrefix(suffix, "fix-g"):
+		return "fix"
+	case strings.HasPrefix(suffix, "fix-v"):
+		return "fix-v"
+	case strings.HasPrefix(suffix, "fix-"):
+		return "fix-r"
+	case strings.HasPrefix(suffix, "review-"):
+		return "review"
+	}
+	return ""
+}
+
+// lpEscalated is whether `scripts/pipeline-fix.sh escalate` recorded key.
+func lpEscalated(env Env, git, repo, key string) bool {
+	_, out := kwRun(env, "", git, "-C", repo, "config", "--get-all", "flow.escalation")
+	return slices.Contains(strings.Fields(out), key)
+}
+
+func checkLoopDispatch(args []string, env Env, stdout, stderr io.Writer) int {
+	const self = "check-loop-dispatch"
+	if len(args) != 4 || args[0] == "" || args[1] == "" {
+		fmt.Fprint(stderr, "usage: check-loop-dispatch.sh <agents-repo> <key> <subagent_type> <model>\n")
+		return 2
+	}
+	repo, key, typ, model := smcAbs(env, args[0]), args[1], args[2], args[3]
+	git, ok := panelResolveGit(env, self, stderr)
+	if !ok {
+		return 2
+	}
+	rows, err := lpTable(repo)
+	if err != nil {
+		fmt.Fprintf(stderr, "%s: %v\n", self, err)
+		return 2
+	}
+	main, ok := lpMain(env, git, repo)
+	if !ok {
+		fmt.Fprintf(stderr, "%s: %s is not a git checkout\n", self, repo)
+		return 2
+	}
+	refuse := func(msg string) int {
+		fmt.Fprintf(stdout, "LOOP-DISPATCH-REFUSED: %s\n", msg)
+		return 1
+	}
+	lp, suffix, row, msg, ok := lpFromKey(env, git, main, key, rows)
+	if !ok {
+		if msg == "" {
+			fmt.Fprintf(stdout, "LOOP-DISPATCH-NA %s\n", key)
+			return 0
+		}
+		return refuse(msg)
+	}
+	if model != row.model || !slices.Contains(row.types, typ) {
+		if model != row.model || !slices.Contains(row.escal, typ) {
+			return refuse(fmt.Sprintf("PAIR — %s runs on subagent_type %s, model %s (the pairs table, %s); this dispatch carries subagent_type %q, model %q",
+				key, strings.Join(row.types, " or "), row.model, lpSkill, typ, model))
+		}
+		if !lpEscalated(env, git, main, key) {
+			return refuse(fmt.Sprintf("ESCALATION — %s on %s needs its no-progress escalation recorded first: scripts/pipeline-fix.sh escalate %s <reason>", key, typ, key))
+		}
+	}
+	family := lpFamily(suffix)
+	if family == "pass" {
+		fmt.Fprintf(stdout, "LOOP-DISPATCH-OK %s\n", key)
+		return 0
+	}
+	mine, ok := lpRows(env, main, lp)
+	if !ok {
+		fmt.Fprintf(stderr, "%s: flow record dispatches -change %s failed -- cannot answer (is flowd up?)\n", self, lp.change)
+		return 2
+	}
+	planDone := false
+	for _, d := range mine {
+		if d.Key == key {
+			return refuse(fmt.Sprintf("KEY-REUSED — %s already ran; the next dispatch takes the next key", key))
+		}
+		if d.EndedAt == nil {
+			return refuse(fmt.Sprintf("OPEN — %s has not been closed: scripts/pipeline-fix.sh end %s <outcome>", d.Key, d.Key))
+		}
+		if lpFamily(strings.TrimPrefix(d.Key, lp.prefix)) == "plan" {
+			planDone = d.Outcome == "completed"
+		}
+	}
+	if family != "plan" && !planDone {
+		return refuse(fmt.Sprintf("PLAN — %s needs a completed %splan dispatch first", key, lp.prefix))
+	}
+	role := map[string]string{"plan": "planner", "fix": "implementer", "fix-r": "implementer", "fix-v": "implementer", "review": "reviewer"}[family]
+	effort := strings.TrimPrefix(typ, "flow-")
+	if !slices.Contains([]string{"low", "medium", "high"}, effort) {
+		effort = "default"
+	}
+	out, _, rc := pcFlow(env, true, "record", "dispatch", "begin", "-C", main, "-change", lp.change,
+		"-role", role, "-model", model, "-effort", effort, "-key", key, "-session-token", lp.token)
+	if rc != 0 {
+		fmt.Fprintf(stderr, "%s: flow record dispatch begin failed (exit %d) -- cannot answer:\n%s", self, rc, out)
+		fmt.Fprintf(stderr, "%s: a loop's rows need its change: scripts/pipeline-fix.sh start creates it\n", self)
+		return 2
+	}
+	fmt.Fprintf(stdout, "LOOP-DISPATCH-OK %s\n", key)
+	return 0
+}
diff --git a/stats/internal/setuptest/containment_test.go b/stats/internal/setuptest/containment_test.go
index 014e6a87..2b0da38a 100644
--- a/stats/internal/setuptest/containment_test.go
+++ b/stats/internal/setuptest/containment_test.go
@@ -85,6 +85,10 @@ func TestCoreExcerptsLinksAndBaseline(t *testing.T) {
 	g.assertContains("the protect-main-checkout registration snippet is printed", g.log, ".claude/hooks/protect-main-checkout.py")
 	g.assertContains("the claude agent-baseline registration snippet is printed", g.log, ".claude/hooks/enforce-agent-baseline.py")
 	g.assertContains("the claude flow-active-change registration snippet is printed", g.log, ".claude/hooks/flow-active-change.py")
+	g.assertExists("the enforce-loop-dispatch hook is installed", filepath.Join(hooks, "enforce-loop-dispatch.py"))
+	g.assertContains("an unregistered enforce-loop-dispatch hook is reported", g.log,
+		"enforce-loop-dispatch hook is installed but NOT registered")
+	g.assertContains("the claude enforce-loop-dispatch registration snippet is printed", g.log, ".claude/hooks/enforce-loop-dispatch.py")
 
 	// An unbalanced marker renders the wrong amount of text into every block, so it must abort
 	// the run rather than ship a half-rule.
diff --git a/stats/internal/setuptest/harnesses_test.go b/stats/internal/setuptest/harnesses_test.go
index 1a39bfd5..7d460cce 100644
--- a/stats/internal/setuptest/harnesses_test.go
+++ b/stats/internal/setuptest/harnesses_test.go
@@ -139,6 +139,8 @@ func TestZcodeGlobalSelfContained(t *testing.T) {
 	g.assertContains("the claude block keeps its own pointers — no rewrite leak", filepath.Join(home, ".claude/CLAUDE.md"), "Full rule: `~/.claude/rules/cored.md`")
 	g.assertContains("the zcode hook registration snippet is printed", g.log, ".zcode/hooks/enforce-agent-baseline.py")
 	g.assertContains("the zcode flow-active-change registration snippet is printed", g.log, ".zcode/hooks/flow-active-change.py")
+	g.assertSymlink("the zcode enforce-loop-dispatch hook is linked", filepath.Join(z, "hooks/enforce-loop-dispatch.py"))
+	g.assertContains("the zcode enforce-loop-dispatch registration snippet is printed", g.log, ".zcode/hooks/enforce-loop-dispatch.py")
 
 	// --- idempotence: a second global must refresh the zcode layer, never duplicate it.
 	home = g.newHome()
diff --git a/stats/internal/setuptest/main_test.go b/stats/internal/setuptest/main_test.go
index 89d2472f..30fb31b2 100644
--- a/stats/internal/setuptest/main_test.go
+++ b/stats/internal/setuptest/main_test.go
@@ -295,6 +295,9 @@ func makeFixtureRepo(d string, badRule bool) error {
 		// the "Core excerpts" group assert that warning, so the fixture has to install the hook
 		// the warning names for their precondition to be real.
 		{"hooks/protect-main-checkout.py", "#!/usr/bin/env python3\n\"\"\"fixture hook\"\"\"\n", 0o644},
+		// The loop-dispatch hook. Named as the real one is, because install_hooks and
+		// install_hooks_zcode report on its registration by looking for that name.
+		{"hooks/enforce-loop-dispatch.py", "#!/usr/bin/env python3\n\"\"\"fixture hook\"\"\"\n", 0o644},
 	}
 	if badRule {
 		// An always-on rule whose BODY carries a bare begin delimiter. Inlining it would put a
````

  - [x] **Step 4: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/guard/ ./internal/setuptest/ && go test ./internal/guard/ -run 'TestLoopTableIsTheSkillsOwn|TestCheckLoopDispatch' -count=1 && go test ./internal/setuptest/ -run 'TestCoreExcerptsLinksAndBaseline|TestZcodeGlobalSelfContained' -count=1`, then from the repository root `scripts/test-enforce-loop-dispatch-hook.sh`, `bash -n setup.sh scripts/check-loop-dispatch.sh`, `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-guard-symlinks.sh`, `scripts/check-installed-citations.sh` — all exit 0 with `gofmt -l` printing nothing. Commit.

**Build:** green
**Files:** `README.md`, `hooks/enforce-loop-dispatch.py`, `scripts/check-loop-dispatch.sh`, `scripts/test-enforce-loop-dispatch-hook.sh`, `setup.sh`, `skills/flow-self-review/SKILL.md`, `stats/internal/guard/check_loop_dispatch_test.go`, `stats/internal/guard/loop.go`, `stats/internal/setuptest/containment_test.go`, `stats/internal/setuptest/harnesses_test.go`, `stats/internal/setuptest/main_test.go`
**Tests:** `TestLoopTableIsTheSkillsOwn`, `TestCheckLoopDispatch`, `TestCheckLoopDispatchUnreadableTable`
**Regression:** reverting the commit admits every loop dispatch unchecked and unrecorded: a fixer on `flow-high` with no escalation, a fix before any plan and a re-review on `opus` all pass.
**Baseline:** before=178 after=181
<!-- measured: cat stats/internal/guard/*_test.go | grep -c '^func Test' (before at a1bca81b; after with this task's diff applied in the scratch clone) -->
**Commit:** feat(hooks): admit a fix-loop dispatch only off its pairs row and record it
**After:** none

**Decision:** loop-is-its-own-change
**Decision:** pairs-table-in-skill
**Decision:** replan-key
**Decision:** loop-guards-in-go
**Decision:** separate-loop-hook
**Decision:** loop-checks-fail-closed

- [x] 2. Refuse a fix-loop landing its flowd records do not cover
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Read `unarchived_landing` in `hooks/protect-main-checkout.py`: a landing push is refused only for an open `spectre/changes/<change>/`, so an archived `pipeline-fix-*` branch with no review at all lands — the defect: b8e0e063, bb273948 and 3fd761da reached agents `main` with no review dispatch.
  - [x] **Step 2: Write the failing test.** Save step 3's diff to `/tmp/pf6-task2.diff`, apply its `stats/internal/guard/check_loop_landing_test.go` and `scripts/test-protect-main-checkout.sh` hunks alone (`git apply --include='stats/internal/guard/check_loop_landing_test.go' --include='scripts/test-protect-main-checkout.sh' /tmp/pf6-task2.diff`) and run `cd stats && go test ./internal/guard/ -run TestCheckLoopLanding -count=1` — it FAILS to build with `undefined: checkLoopLanding` — and `scripts/test-protect-main-checkout.sh`, which FAILS cases 77, 79 and 80 (`wanted deny, got allow`).
  - [x] **Step 3: Implement.** Apply the rest of the diff (`git apply --exclude='stats/internal/guard/check_loop_landing_test.go' --exclude='scripts/test-protect-main-checkout.sh' /tmp/pf6-task2.diff`) — the landing guard, its shim, the hook's loop-landing check and the README line:

````diff verified:applied with git apply in plan order to a scratch clone at a1bca81b; gofmt -l, go vet, go test ./internal/guard/ ./internal/setuptest/ -race and scripts/run-guard-tests.sh pass with every task applied; TestCheckLoopLanding fails with the clean-review check removed, and cases 77, 79 and 80 fail with the hook at a1bca81b
diff --git a/README.md b/README.md
index 21f356b2..8df3e1b9 100644
--- a/README.md
+++ b/README.md
@@ -194,7 +194,9 @@ return a verdict. Each one has a `test-*.sh` harness in `scripts/`, and
   row, a fix with no completed plan, or one while an earlier dispatch is open, and records the
   admitted one in flowd (`scripts/check-loop-dispatch.sh`).
 - `protect-main-checkout.py`: blocks edits, staging, commits, stashes and resets on a main
-  checkout's default branch. Work happens in a worktree.
+  checkout's default branch. Work happens in a worktree. It also blocks a landing push of a
+  `pipeline-fix-*` or `self-review-*` branch that flowd's records do not cover
+  (`scripts/check-loop-landing.sh`).
 - `flow-active-change.py`: routes a plain problem report to the session's open change as a fix run.
 
 **Plan**
diff --git a/hooks/protect-main-checkout.py b/hooks/protect-main-checkout.py
index 05af7cb9..af4597fa 100755
--- a/hooks/protect-main-checkout.py
+++ b/hooks/protect-main-checkout.py
@@ -30,7 +30,10 @@ where `<src>`'s tree still holds `spectre/changes/<branch>/` — `<branch>` bein
 current branch for `HEAD`, with a `/flow` branch's `spectre/` prefix dropped. A change lands
 archived (`spectre archive <change>`, then commit), never as an open folder on the default branch
 that no later run owns. `/flow`'s run 1 archives before its push, so only a run that skipped it or
-an ad-hoc landing meets this.
+an ad-hoc landing meets this. A landing push of a `pipeline-fix-*` or `self-review-*` branch is
+denied too unless scripts/check-loop-landing.sh, in the agents checkout this hook is installed
+from, admits it — flowd holding the loop's completed plan, fix and clean last review. That one
+check fails closed: a checker that cannot answer or cannot run denies the push.
 
 Allowed everywhere: reads, `git pull --ff-only`, `git fetch`, `git checkout -b`, `git switch -c`,
 `git worktree add`, `git push`, and the landing scripts (`land-self-review-report.sh`,
@@ -70,6 +73,13 @@ DENIED_GIT_VERBS = {
 
 FALLBACK_DEFAULT_BRANCHES = {"main", "master", "develop"}
 
+LOOP_BRANCH_RE = re.compile(r"^(pipeline-fix-.+-[0-9]+|self-review-.+)$")
+
+# The loop-landing checker; overridable for the hook's own harness only.
+LOOP_CHECK = os.environ.get("LOOP_LANDING_CHECK") or os.path.join(
+    os.path.dirname(os.path.dirname(os.path.realpath(__file__))), "scripts", "check-loop-landing.sh"
+)
+
 _git_cache = {}
 
 
@@ -369,6 +379,36 @@ def unarchived_landing(args, target):
     return None
 
 
+def loop_landing(args, target):
+    """The deny reason when the push args land a loop branch `<src>:<default>` that
+    check-loop-landing.sh refuses, else None."""
+    if target is None:
+        return None
+    default = default_branch(target)
+    for a in args:
+        if ":" not in a or a.startswith("-"):
+            continue
+        src, _, dst = a.lstrip("+").partition(":")
+        dst = dst.removeprefix("refs/heads/")
+        if not src or dst not in ({default} if default else FALLBACK_DEFAULT_BRANCHES):
+            continue
+        branch = git(target, "branch", "--show-current") if src == "HEAD" else src
+        branch = (branch or "").removeprefix("refs/heads/")
+        if not LOOP_BRANCH_RE.match(branch):
+            continue
+        try:
+            r = subprocess.run([LOOP_CHECK, target, branch], capture_output=True, text=True, timeout=300)
+        except (OSError, subprocess.SubprocessError) as e:
+            return f"Blocked: landing {branch} needs check-loop-landing.sh, which could not run ({e})."
+        if r.returncode != 0:
+            return (
+                f"Blocked: {branch} is a fix loop, and its records do not cover this landing "
+                f"(check-loop-landing.sh exit {r.returncode}):\n{r.stdout}{r.stderr}\n"
+                "Land it with scripts/pipeline-fix.sh land <worktree> once they do."
+            )
+    return None
+
+
 def bash_hits(command, cwd, landings=None):
     """Paths a Bash command writes into or git-mutates, resolved against cwd and `cd`.
 
@@ -420,7 +460,10 @@ def bash_hits(command, cwd, landings=None):
                         args.append(t)
                     hit = unarchived_landing(args, target)
                     if hit:
-                        landings.append(hit)
+                        landings.append(landing_reason(*hit))
+                    loop = loop_landing(args, target)
+                    if loop:
+                        landings.append(loop)
                 i = j + 1
                 continue
             if tok == "sed" and any(t.startswith("-i") for t in tokens[i + 1 : i + 4]):
@@ -521,7 +564,7 @@ def main():
                         "hookSpecificOutput": {
                             "hookEventName": "PreToolUse",
                             "permissionDecision": "deny",
-                            "permissionDecisionReason": landing_reason(*landings[0]),
+                            "permissionDecisionReason": landings[0],
                         }
                     },
                     sys.stdout,
diff --git a/scripts/check-loop-landing.sh b/scripts/check-loop-landing.sh
new file mode 100755
index 00000000..fd3fd0d1
--- /dev/null
+++ b/scripts/check-loop-landing.sh
@@ -0,0 +1,26 @@
+#!/usr/bin/env bash
+# check-loop-landing.sh <repo> <branch> — may a pipeline-fix-* or
+# self-review-* branch land on the default branch?
+#
+# Only once flowd holds, under the loop's change (its branch name), a
+# completed plan dispatch, a completed first fix (`fix` or `fix-g<g>`), at
+# least one review, every row closed, the highest-numbered review closed
+# `clean` — no Critical, no Important — and no `fix-v<v>` after it.
+# hooks/protect-main-checkout.py runs it on a landing push typed by hand;
+# pipeline-fix.sh land runs the same check before its own push.
+#
+# Exit codes: 0 LOOP-LANDING-OK, or LOOP-LANDING-NA for any other branch;
+# 1 one LOOP-LANDING-REFUSED line per reason; 2 cannot answer (flowd
+# unreadable, not a git checkout, usage).
+#
+# Go, in flow-guard (stats/internal/guard/looplanding.go). flow-guard is
+# built from this checkout, never taken from PATH: scripts/lib/flow-guard.sh
+# derives it, and exits 2 with the cause when it cannot.
+set -euo pipefail
+# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
+SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
+[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
+  echo "check-loop-landing: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
+  exit 2
+}
+flow_guard_exec check-loop-landing 2 "check-loop-landing:" "$@"
diff --git a/scripts/test-protect-main-checkout.sh b/scripts/test-protect-main-checkout.sh
index 62f334e1..b119f0d5 100755
--- a/scripts/test-protect-main-checkout.sh
+++ b/scripts/test-protect-main-checkout.sh
@@ -78,6 +78,12 @@
 #   75. `bash <(` then `cat > /dev/stdout <<'EOF'`, body `rm <main>/f.txt` -> deny (the enclosing command runs it)
 #   76. `cat > <sibling>/x.sh <<'EOF'`, body `rm <main>/f.txt`, then
 #       `sh "<sibling>"/x.sh`                                           -> deny (written, then run under another spelling)
+#   77. Bash `git -C <loop wt> push origin pipeline-fix-x-6:main`, the
+#       loop checker refusing                                         -> deny, its lines in the reason
+#   78. the same push, the loop checker admitting                     -> allow, checker handed <wt> <branch>
+#   79. Bash `git -C <self-review wt> push origin HEAD:main`, refusing -> deny
+#   80. the loop push with no runnable loop checker                   -> deny (fails closed)
+#   81. Bash `git -C <wt> push origin change:main`, archived           -> allow, loop checker never run
 set -euo pipefail
 SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
 HOOK="$SCRIPT_DIR/../hooks/protect-main-checkout.py"
@@ -367,6 +373,44 @@ rm $MAIN/f.txt
 EOF
 sh \"$SIB\"/x.sh")}")"
 
+# 77-81: a fix loop lands only once check-loop-landing.sh admits it; the
+# checker is a stub here (the guard is covered by check_loop_landing_test.go).
+LOOP_LOG="$ROOT/loop-args"
+LOOP_RC="$ROOT/loop-rc"
+cat >"$ROOT/loopcheck" <<STUB
+#!/usr/bin/env bash
+printf '%s\n' "\$@" >"$LOOP_LOG"
+echo "LOOP-LANDING-REFUSED: REVIEW — stub"
+exit "\$(cat "$LOOP_RC")"
+STUB
+chmod +x "$ROOT/loopcheck"
+export LOOP_LANDING_CHECK="$ROOT/loopcheck"
+git_q -C "$MAIN" worktree add "$ROOT/main-worktrees/pipeline-fix-x-6" -b pipeline-fix-x-6 origin/main
+git_q -C "$MAIN" worktree add "$ROOT/main-worktrees/self-review-x" -b self-review-x origin/main
+LWT="$ROOT/main-worktrees/pipeline-fix-x-6"
+echo 1 >"$LOOP_RC"
+out="$(printf '{"tool_name":"Bash","tool_input":{"command":%s},"cwd":"%s"}' "$(q "git -C $LWT push origin pipeline-fix-x-6:main")" "$ROOT" | python3 "$HOOK")"
+if grep -q '"permissionDecision": *"deny"' <<<"$out" && grep -q "LOOP-LANDING-REFUSED: REVIEW" <<<"$out"; then
+  pass "77 loop push the checker refuses"
+else
+  fail "77 loop push the checker refuses: $out"
+fi
+echo 0 >"$LOOP_RC"
+expect "78 loop push the checker admits" allow \
+  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $LWT push origin pipeline-fix-x-6:main")}")"
+[ "$(tr '\n' ' ' <"$LOOP_LOG")" = "$LWT pipeline-fix-x-6 " ] \
+  && pass "78 checker args" || fail "78 checker args: $(tr '\n' ' ' <"$LOOP_LOG")"
+echo 1 >"$LOOP_RC"
+expect "79 self-review HEAD:main push the checker refuses" deny \
+  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $ROOT/main-worktrees/self-review-x push origin HEAD:main")}")"
+expect "80 loop push with no runnable checker" deny \
+  "$(LOOP_LANDING_CHECK="$ROOT/missing" run "$ROOT" Bash "{\"command\":$(q "git -C $LWT push origin pipeline-fix-x-6:main")}")"
+rm -f "$LOOP_LOG"
+expect "81 a non-loop landing" allow \
+  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $WT push origin change:main")}")"
+[ ! -e "$LOOP_LOG" ] && pass "81 loop checker not run" || fail "81 loop checker ran"
+unset LOOP_LANDING_CHECK
+
 # 3 last: moving the main checkout off main lifts the protection
 git_q -C "$MAIN" checkout -q -b feature
 expect "3 Write in main checkout on a feature branch" allow \
diff --git a/stats/internal/guard/check_loop_landing_test.go b/stats/internal/guard/check_loop_landing_test.go
new file mode 100644
index 00000000..b368c5d3
--- /dev/null
+++ b/stats/internal/guard/check_loop_landing_test.go
@@ -0,0 +1,80 @@
+package guard
+
+import (
+	"bytes"
+	"fmt"
+	"strings"
+	"testing"
+)
+
+// lpRowsJSON is `flow record dispatches` output for the pipeline-fix fixture
+// loop: each entry "<suffix> <outcome>", an outcome "open" leaving the row
+// unclosed; a leading "other:" files the row under another session token.
+func lpRowsJSON(prefix, token string, rows ...string) string {
+	var out []string
+	for i, r := range rows {
+		tok := token
+		if rest, ok := strings.CutPrefix(r, "other:"); ok {
+			r, tok = rest, "mf-run"
+		}
+		suffix, outcome, _ := strings.Cut(r, " ")
+		ended := `,"endedAt":"2026-10-16T10:00:00Z"`
+		if outcome == "open" {
+			outcome, ended = "", ""
+		}
+		out = append(out, fmt.Sprintf(`{"seq":%d,"key":"%s%s","outcome":"%s","sessionToken":"%s"%s}`, i+1, prefix, suffix, outcome, tok, ended))
+	}
+	return "[" + strings.Join(out, ",") + "]"
+}
+
+func TestCheckLoopLanding(t *testing.T) {
+	t.Parallel()
+	pf := func(rows ...string) string { return lpRowsJSON("pipeline-fix-6-", lpBranch, rows...) }
+	for _, tc := range []struct {
+		name, branch, rows string
+		rc                 int
+		out                []string
+	}{
+		{name: "another branch", branch: "spectre/kan-1-demo", rows: `[]`, out: []string{"LOOP-LANDING-NA"}},
+		{name: "a reviewed loop", branch: lpBranch, rows: pf("plan completed", "fix completed", "review-1 fix", "fix-1 completed", "review-2 clean"),
+			out: []string{"LOOP-LANDING-OK"}},
+		{name: "a Minor-only review closed by its fix", branch: lpBranch, rows: pf("plan completed", "fix-g1 completed", "fix-g2 completed", "review-1 clean", "fix-1 completed"),
+			out: []string{"LOOP-LANDING-OK"}},
+		{name: "fixes landed with no review", branch: lpBranch, rows: pf("plan completed", "fix completed"), rc: 1,
+			out: []string{"REVIEW — no pipeline-fix-6-review-<r>"}},
+		{name: "a fixer with no planner", branch: lpBranch, rows: pf("fix completed", "review-1 clean"), rc: 1,
+			out: []string{"PLAN — no completed pipeline-fix-6-plan"}},
+		{name: "a planner with no fixer", branch: lpBranch, rows: pf("plan completed", "review-1 clean"), rc: 1,
+			out: []string{"FIX — no completed pipeline-fix-6-fix"}},
+		{name: "the last review raised an Important", branch: lpBranch, rows: pf("plan completed", "fix completed", "review-1 clean", "fix-1 completed", "review-2 fix"), rc: 1,
+			out: []string{`REVIEW — pipeline-fix-6-review-2 ended "fix"`}},
+		{name: "a red-line fix after the last review", branch: lpBranch, rows: pf("plan completed", "fix completed", "review-1 clean", "fix-v1 completed"), rc: 1,
+			out: []string{"REVIEW — a pipeline-fix-6-fix-v<v> dispatch came after pipeline-fix-6-review-1"}},
+		{name: "a row left open", branch: lpBranch, rows: pf("plan completed", "fix completed", "review-1 open"), rc: 1,
+			out: []string{"OPEN — pipeline-fix-6-review-1", `ended ""`}},
+		{name: "the run's own rows do not count", branch: lpBranch, rows: pf("other:plan completed", "other:fix completed", "other:review-1 clean"), rc: 1,
+			out: []string{"PLAN —", "FIX —", "REVIEW —"}},
+		{name: "a self-review loop", branch: "self-review-kan-1-demo",
+			rows: lpRowsJSON("self-review-kan-1-demo-", "self-review-kan-1-demo", "plan completed", "fix completed", "review-1 clean"), out: []string{"LOOP-LANDING-OK"}},
+		{name: "flowd unreadable", branch: lpBranch, rows: "", rc: 2, out: []string{"cannot answer"}},
+	} {
+		t.Run(tc.name, func(t *testing.T) {
+			t.Parallel()
+			fx := lpNew(t, tc.rows)
+			var o, e bytes.Buffer
+			rc := checkLoopLanding([]string{fx.repo, tc.branch}, fx.env, &o, &e)
+			out := o.String() + e.String()
+			if rc != tc.rc {
+				t.Fatalf("rc %d, want %d: %s", rc, tc.rc, out)
+			}
+			for _, w := range tc.out {
+				if !strings.Contains(out, w) {
+					t.Fatalf("output %q carries no %q", out, w)
+				}
+			}
+			if calls := fx.calls(t); tc.rows != `[]` && tc.branch != "spectre/kan-1-demo" && !strings.Contains(calls, "-change "+tc.branch) {
+				t.Fatalf("flow calls %q never read the loop's change", calls)
+			}
+		})
+	}
+}
diff --git a/stats/internal/guard/looplanding.go b/stats/internal/guard/looplanding.go
new file mode 100644
index 00000000..e882841f
--- /dev/null
+++ b/stats/internal/guard/looplanding.go
@@ -0,0 +1,118 @@
+package guard
+
+import (
+	"fmt"
+	"io"
+	"regexp"
+	"strconv"
+	"strings"
+)
+
+// check-loop-landing.sh <repo> <branch>: may a pipeline-fix-* or
+// self-review-* branch land on the default branch? Only once flowd holds,
+// for its loop, a completed plan, a completed first fix and a review, every
+// row closed, the highest review `clean` -- no Critical or Important -- and
+// no red-line fix (`fix-v<v>`) after it. protect-main-checkout.py runs it on
+// a landing push typed by hand; pipeline-fix.sh land runs it before its own.
+func init() {
+	Registry["check-loop-landing"] = checkLoopLanding
+}
+
+var (
+	lpPipelineBranch = regexp.MustCompile(`^pipeline-fix-(.+)-([0-9]+)$`)
+	lpReviewSuffix   = regexp.MustCompile(`^review-([0-9]+)$`)
+)
+
+// lpFromBranch is the loop a branch is, or ok false for any other branch.
+func lpFromBranch(branch string) (lpLoop, bool) {
+	if m := lpPipelineBranch.FindStringSubmatch(branch); m != nil && plainChangeName(branch) {
+		return lpLoop{change: branch, token: branch, prefix: "pipeline-fix-" + m[2] + "-"}, true
+	}
+	if name, ok := strings.CutPrefix(branch, "self-review-"); ok && plainChangeName(name) {
+		return lpLoop{change: branch, token: branch, prefix: branch + "-"}, true
+	}
+	return lpLoop{}, false
+}
+
+// lpLandingRefusals is every reason the loop's rows refuse a landing, none
+// when it may land.
+func lpLandingRefusals(lp lpLoop, rows []lpDispatch) []string {
+	var plan, fix bool
+	var open []string
+	lastReview, lastReviewSeq, lastFixVSeq := -1, -1, -1
+	lastReviewOutcome, lastReviewKey := "", ""
+	for _, d := range rows {
+		suffix := strings.TrimPrefix(d.Key, lp.prefix)
+		if d.EndedAt == nil {
+			open = append(open, d.Key)
+		}
+		switch lpFamily(suffix) {
+		case "plan":
+			plan = plan || d.Outcome == "completed"
+		case "fix":
+			fix = fix || d.Outcome == "completed"
+		case "fix-v":
+			lastFixVSeq = d.Seq
+		case "review":
+			n, _ := strconv.Atoi(lpReviewSuffix.FindStringSubmatch(suffix)[1])
+			if n > lastReview {
+				lastReview, lastReviewSeq, lastReviewOutcome, lastReviewKey = n, d.Seq, d.Outcome, d.Key
+			}
+		}
+	}
+	var out []string
+	if !plan {
+		out = append(out, fmt.Sprintf("PLAN — no completed %splan dispatch", lp.prefix))
+	}
+	if !fix {
+		out = append(out, fmt.Sprintf("FIX — no completed %sfix dispatch", lp.prefix))
+	}
+	for _, k := range open {
+		out = append(out, fmt.Sprintf("OPEN — %s has not been closed: scripts/pipeline-fix.sh end %s <outcome>", k, k))
+	}
+	switch {
+	case lastReview < 0:
+		out = append(out, fmt.Sprintf("REVIEW — no %sreview-<r> dispatch", lp.prefix))
+	case lastReviewOutcome != "clean":
+		out = append(out, fmt.Sprintf("REVIEW — %s ended %q, not clean: fix its Critical and Important findings, then the next review", lastReviewKey, lastReviewOutcome))
+	case lastFixVSeq > lastReviewSeq:
+		out = append(out, fmt.Sprintf("REVIEW — a %sfix-v<v> dispatch came after %s: run the next review", lp.prefix, lastReviewKey))
+	}
+	return out
+}
+
+func checkLoopLanding(args []string, env Env, stdout, stderr io.Writer) int {
+	const self = "check-loop-landing"
+	if len(args) != 2 || args[0] == "" || args[1] == "" {
+		fmt.Fprint(stderr, "usage: check-loop-landing.sh <repo> <branch>\n")
+		return 2
+	}
+	lp, ok := lpFromBranch(args[1])
+	if !ok {
+		fmt.Fprintf(stdout, "LOOP-LANDING-NA %s\n", args[1])
+		return 0
+	}
+	git, ok := panelResolveGit(env, self, stderr)
+	if !ok {
+		return 2
+	}
+	main, ok := lpMain(env, git, smcAbs(env, args[0]))
+	if !ok {
+		fmt.Fprintf(stderr, "%s: %s is not a git checkout\n", self, args[0])
+		return 2
+	}
+	rows, ok := lpRows(env, main, lp)
+	if !ok {
+		fmt.Fprintf(stderr, "%s: flow record dispatches -change %s failed -- cannot answer (is flowd up?)\n", self, lp.change)
+		return 2
+	}
+	refusals := lpLandingRefusals(lp, rows)
+	for _, r := range refusals {
+		fmt.Fprintf(stdout, "LOOP-LANDING-REFUSED: %s\n", r)
+	}
+	if len(refusals) > 0 {
+		return 1
+	}
+	fmt.Fprintf(stdout, "LOOP-LANDING-OK %s\n", args[1])
+	return 0
+}
````

  - [x] **Step 4: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckLoopLanding|TestCheckLoopDispatch' -count=1`, then from the repository root `scripts/test-protect-main-checkout.sh`, `bash -n scripts/check-loop-landing.sh`, `scripts/check-references.sh`, `scripts/check-python-suppressions.sh`, `scripts/check-guard-symlinks.sh`, `scripts/check-markdown-integrity.py` — all exit 0 with `gofmt -l` printing nothing. Commit.

**Build:** green
**Files:** `README.md`, `hooks/protect-main-checkout.py`, `scripts/check-loop-landing.sh`, `scripts/test-protect-main-checkout.sh`, `stats/internal/guard/check_loop_landing_test.go`, `stats/internal/guard/looplanding.go`, `stats/internal/setuptest/hook_timeouts_test.go`
**Tests:** `TestCheckLoopLanding`
**Regression:** reverting the commit lets a `pipeline-fix-*` or `self-review-*` branch land on `main` with no plan, no fix or no clean review recorded.
**Baseline:** before=181 after=182
<!-- measured: cat stats/internal/guard/*_test.go | grep -c '^func Test' (before and after with Tasks 1 and 2 applied in the scratch clone at a1bca81b) -->
**Commit:** feat(hooks): refuse a fix-loop landing its flowd records do not cover
**After:** Task 1

**Decision:** landing-in-protect-main-checkout
**Decision:** loop-checks-fail-closed
**Decision:** forward-only

- [x] 3. Run the fix loop's fixed mechanics from `pipeline-fix.sh`
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Read item 1 and item 4 of **Pipeline defects found mid-run** (`skills/flow-contracts/pipeline.md`) and items 1, 4 and 5 of step 3 of `skills/flow-self-review/SKILL.md`: the worktree add, worktree setup, lint-and-test selection, rebase, push, pull and removal are restated as commands for the parent to type, and no script runs them — the defect: each is a step the parent can skip.
  - [x] **Step 2: Write the failing test.** Save step 3's diff to `/tmp/pf6-task3.diff`, apply its `stats/internal/guard/pipeline_fix_test.go` hunk alone (`git apply --include='stats/internal/guard/pipeline_fix_test.go' /tmp/pf6-task3.diff`) and run `cd stats && go test ./internal/guard/ -run TestPipelineFix -count=1`: it FAILS to build with `undefined: pipelineFix`.
  - [x] **Step 3: Implement.** Apply the rest of the diff (`git apply --exclude='stats/internal/guard/pipeline_fix_test.go' /tmp/pf6-task3.diff`) — the script, kickoff's worktree-setup step shared with it, and the two skill files citing it:

````diff verified:applied with git apply in plan order to a scratch clone at a1bca81b; gofmt -l, go vet, go test ./internal/guard/ ./internal/setuptest/ -race and scripts/run-guard-tests.sh pass with every task applied; TestPipelineFixLand fails with the post-rebase verify removed (a_moved_base_is_rebased_and_verified_first)
diff --git a/scripts/pipeline-fix.sh b/scripts/pipeline-fix.sh
new file mode 100755
index 00000000..637aab43
--- /dev/null
+++ b/scripts/pipeline-fix.sh
@@ -0,0 +1,63 @@
+#!/usr/bin/env bash
+# pipeline-fix.sh — the fixed mechanics of a pipeline-fix or self-review fix
+# loop (skills/flow-contracts/pipeline.md **Pipeline defects found mid-run**,
+# skills/flow-self-review/SKILL.md step 3), run by the parent, so no step is
+# retyped or skipped. The loop is its own change in this repository's flowd
+# project, named after its branch — pipeline-fix-<name>-<k> or
+# self-review-<name> — and check-loop-dispatch.sh files every loop dispatch
+# under it (a self-review on the change's own branch files under <name>).
+#
+# Usage:
+#   pipeline-fix.sh start <name> [<k>]
+#       git fetch origin; `git worktree add -b <branch> <repo>-worktrees/<branch>
+#       origin/<default>`, <branch> pipeline-fix-<name>-<k> with <k>, else
+#       self-review-<name>; creates <worktree>/.superpowers/sdd/<branch>/ and the
+#       loop's STARTED state record (flow state set); runs ## worktree setup.
+#       stdout: `worktree: <path>`.
+#   pipeline-fix.sh end <key> <outcome>
+#       closes the dispatch <key> (flow record dispatch end). A review takes
+#       `clean` — it raised no Critical and no Important — or `fix`; every
+#       other dispatch `completed` or `aborted`.
+#   pipeline-fix.sh escalate <key>
+#       records <key>'s no-progress escalation (git config flow.escalation),
+#       which check-loop-dispatch.sh requires before admitting <key> on its
+#       pairs row's escalation type; refused unless the pairs table admits an
+#       escalation for <key> and an earlier round of its kind completed.
+#   pipeline-fix.sh verify <worktree>
+#       runs the ## lint lines and tests the branch's touched paths need,
+#       against origin/<default>: every guard line, the `cd stats` lines for
+#       touched Go, the `cd stats/web` lines for a touched SPA file; go test
+#       of each touched stats package, scripts/test-<n>.sh or
+#       scripts/test-<n>-hook.sh of a touched script or hook, the setuptest
+#       package for setup.sh, vitest of a touched SPA test or the test beside
+#       a touched source. One `green:` or `red:` line per command, a red one
+#       followed by its last 40 output lines.
+#   pipeline-fix.sh land <worktree>
+#       the merge-and-push route: refuses any other `## default landing
+#       route`; git fetch origin; a branch whose base moved is rebased onto
+#       origin/<default> and verified again before anything is pushed (a
+#       stopped rebase is aborted); check-loop-landing's refusals; `git push
+#       origin <branch>:<default>`; the loop's FINISHED state record; `git
+#       pull --ff-only` in the main checkout; removes the worktree and the
+#       branch. stdout: `landed: <sha>`.
+#   pipeline-fix.sh drop <worktree>
+#       a deferred loop: the FINISHED state record with withdrawn: true, then
+#       removes the worktree (--force) and the branch, landing nothing.
+#
+# Exit codes: 0 done; 1 refused, red, or a step failed (its output relayed);
+# 2 usage or it cannot answer (no flowd answer, no git, no origin/HEAD).
+#
+# Go, in flow-guard (stats/internal/guard/pipelinefix.go). flow-guard is
+# built from this checkout, never taken from PATH: scripts/lib/flow-guard.sh
+# derives it, and exits 2 with the cause when it cannot.
+set -euo pipefail
+# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
+SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
+[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
+  echo "pipeline-fix: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
+  exit 2
+}
+FLOW_GUARD_REPO_ROOT="$(flow_guard_root)"
+FLOW_GUARD_PROJECT_GET="$SCRIPT_DIR/project-get.sh"
+export FLOW_GUARD_REPO_ROOT FLOW_GUARD_PROJECT_GET
+flow_guard_exec pipeline-fix 2 "pipeline-fix:" "$@"
diff --git a/skills/flow-contracts/pipeline.md b/skills/flow-contracts/pipeline.md
index 58da0b21..253e6e7e 100644
--- a/skills/flow-contracts/pipeline.md
+++ b/skills/flow-contracts/pipeline.md
@@ -190,11 +190,9 @@ later run of the change reuses a name; its name
 the change, its branch, its worktree and `<fix-branch>` — and `origin/<agents-base>` for
 `<fix-base>`; the parent is that file's "this session".
 
-1. **Plan, then fix** — the parent creates the fix's own worktree,
-   `git -C <agents repo> worktree add -b pipeline-fix-<name>-<k> <agents repo>-worktrees/pipeline-fix-<name>-<k> origin/<agents-base>`
-   — never the main checkout, never the change's own worktree — and runs `<agents repo>`'s
-   `## worktree setup` in it (`project-get.sh <agents repo> 'worktree setup'`). It dispatches
-   `pipeline-fix-<k>-plan`, its prompt carrying the defect, its evidence and the counted blast
+1. **Plan, then fix** — the parent creates the fix's own worktree with
+   `<agents repo>/scripts/pipeline-fix.sh start <name> <k>` — never the main checkout, never the
+   change's own worktree. It dispatches `pipeline-fix-<k>-plan`, its prompt carrying the defect, its evidence and the counted blast
    radius, then `pipeline-fix-<k>-fix`, `-fix-g<g>` when the planner recorded more than one group,
    each as item 1 there states. The fixer also adds one test or guard that fails without the fix,
    and updates a saved memory that encoded the defect.
@@ -202,17 +200,14 @@ the change, its branch, its worktree and `<fix-branch>` — and `origin/<agents-
    `git diff origin/<agents-base>...pipeline-fix-<name>-<k>`, as item 2 there states.
 3. **Fix the findings** — a fresh `pipeline-fix-<k>-fix-<r>`, as item 3 there states, under
    **Fewest operator actions** above. A defect whose commit the review drops is deferred, as a
-   `big` one is: the parent removes the worktree and the branch, landing nothing.
+   `big` one is: the parent runs `<agents repo>/scripts/pipeline-fix.sh drop <worktree>`, landing
+   nothing.
 4. **Land** — once the parent has ticked and archived the plan in one
    `chore(spectre): archive pipeline-fix-<name>-<k>` commit and run the lint lines and tests, as
    item 4 there states, a red line fixed by `pipeline-fix-<k>-fix-v<v>`, it lands by
    `<agents repo>`'s `## default landing route`
-   (`project-get.sh <agents repo> 'default landing route'`), without asking. Merge and push is:
-   `git -C <agents repo> fetch origin`, rebase `pipeline-fix-<name>-<k>` onto
-   `origin/<agents-base>`, `git push origin pipeline-fix-<name>-<k>:<agents-base>`,
-   `git -C <agents repo> pull --ff-only`, then remove the worktree and the branch. A rebase onto
-   a moved `origin/<agents-base>` runs the lint lines and tests again before the push, as item 5
-   there states.
+   (`project-get.sh <agents repo> 'default landing route'`), without asking. Merge and push is
+   `<agents repo>/scripts/pipeline-fix.sh land <worktree>`.
 
 The loop runs in the background while the run's own work continues; the run still never ends a
 turn with one of its children in flight.
diff --git a/skills/flow-self-review/SKILL.md b/skills/flow-self-review/SKILL.md
index 373dcacf..de5c7f22 100644
--- a/skills/flow-self-review/SKILL.md
+++ b/skills/flow-self-review/SKILL.md
@@ -37,8 +37,8 @@ checkout on any other branch: print the branch and stop.
 
 ### 2. Read the bundle and run the seven angles
 
-Dispatch the pass, in-run and standalone alike: key `self-review-<name>-pass`,
-`subagent_type: flow-medium` on `opus`, its prompt carrying the bundle in full (`<bundle>`
+Dispatch the pass, in-run and standalone alike: key `self-review-<name>-pass`, on the `pass` row
+of the pairs table (step 3), its prompt carrying the bundle in full (`<bundle>`
 standalone, the bundle held in memory in-run), this step's rules and step 3's classification, the
 dispatch paragraphs step 3 names and READ-ONLY REVIEW. It returns every finding with its angle, its
 evidence and its classification — `noted` or fixable, in or out of scope, and for a fixable one its
@@ -118,8 +118,9 @@ shows, with evidence, that the defect it names does not exist. A planner report
 finding on any other ground is sent back to a fresh planner dispatch carrying the objection as a
 constraint, never offered for filing in step 4 (kan-982: the `Task-Id:` guard and the scope check
 were dropped as imprecise and offered for filing; the operator answered "fix them now"). A finding hard
-to solve escalates the planner rather than leaving the loop: the re-plan may run on `flow-high`
-(effort `high`) in place of the planner pair's `flow-medium`. Where `high` still cannot produce a
+to solve escalates the planner rather than leaving the loop: the re-plan, key
+`self-review-<name>-plan-<n>`, may run on `flow-high` (effort `high`) in place of the planner
+pair's `flow-medium`. Where `high` still cannot produce a
 design that holds, the next step is effort `xhigh` — and an `xhigh` dispatch is never made
 without the operator's approval, asked first with the finding, the objection and what `high`
 returned; a refusal leaves the finding offered for filing in step 4. **This session plans,
@@ -146,6 +147,19 @@ key suffix, the table `check-loop-dispatch.sh` reads and refuses a loop dispatch
 the two `review` rows: the first review on `flow-medium` or `flow-low` as the planner recorded —
 `flow-medium` when it recorded none — and every re-review after a fix on `flow-low` and `sonnet`.
 
+**Every loop dispatch is admitted and recorded mechanically.**
+`<agents repo>/hooks/enforce-loop-dispatch.py` hands each Agent call whose prompt carries a
+`Key: pipeline-fix-<k>-…` or `Key: self-review-<name>-…` line to
+`<agents repo>/scripts/check-loop-dispatch.sh`, which refuses it off its row of the table above,
+while an earlier dispatch of the loop is still open, or — any dispatch but the plan and the pass —
+before a completed plan, and records the admitted one in flowd. The parent closes each dispatch
+once it returns with `<agents repo>/scripts/pipeline-fix.sh end <key> <outcome>`: `completed`, or
+`aborted` for one that did not finish, and for a review `clean` when it raised no Critical and no
+Important, `fix` otherwise. A dispatch on its row's escalation column is admitted only after
+`<agents repo>/scripts/pipeline-fix.sh escalate <key>`, and every loop branch's landing push only
+once `<agents repo>/scripts/check-loop-landing.sh` finds its plan, its fix and a last review
+closed `clean`.
+
 `<agents-base>` is the one **Pipeline defects found mid-run** binds. `<fix-branch>` and
 `<fix-base>` are where the fixes go (`design.md`, `fixes-ride-the-change`): inside a run whose
 change's canonical repository is `<agents repo>`, the change's own branch in its own
@@ -153,11 +167,9 @@ worktree — `spectre/<name>` on `/flow`, `<name>` on `/flow-fast` — based at
 holds when the pass starts; standalone, or inside
 any other project's run, `self-review-<name>` based at `origin/<agents-base>`.
 
-1. **Plan, then fix** — on `self-review-<name>`, this session creates the worktree,
-   `git -C <agents repo> worktree add -b self-review-<name> <agents repo>-worktrees/self-review-<name> origin/<agents-base>`
-   — never the main checkout — runs `<agents repo>`'s `## worktree setup` in it
-   (`project-get.sh <agents repo> 'worktree setup'`); on the change's own branch its worktree is
-   already set up. It then dispatches key `self-review-<name>-plan` on the planner pair, its
+1. **Plan, then fix** — on `self-review-<name>`, this session creates the worktree with
+   `<agents repo>/scripts/pipeline-fix.sh start <name>` — never the main checkout; on the change's
+   own branch its worktree is already set up. It then dispatches key `self-review-<name>-plan` on the planner pair, its
    prompt carrying every non-`big` finding with its evidence and counted blast radius. The planner
    plans them all in full as one spectre change, `self-review-<name>`, in that worktree, per
    `skills/flow/brainstorm-planner.md`: brainstorming (section B), the artifacts (section C —
@@ -211,11 +223,9 @@ any other project's run, `self-review-<name>` based at `origin/<agents-base>`.
    unticked, its finding found `big` or its commit dropped — and commits both as
    `chore(spectre): archive self-review-<name>`, on either branch, before any landing push —
    `<agents repo>`'s push hook refuses an unarchived change on its default branch. That commit
-   alone carries the ticks, and is made once. This session then runs, in the fix worktree, only the
-   `<agents repo>/.flow/project.md` `## lint` lines the touched files need, plus the tests
-   covering them: `go test` of the touched packages, the `scripts/test-*.sh` harness of a touched
-   script, `npx vitest run <file>` of a touched SPA file — each output through `tail`, never the
-   full `## lint` or `## test` list. A red line is fixed by a fresh dispatch of step 3, key
+   alone carries the ticks, and is made once. This session then runs
+   `<agents repo>/scripts/pipeline-fix.sh verify <fix worktree>`. A red line is fixed by a fresh
+   dispatch of step 3, key
    `self-review-<name>-fix-v<v>`, its prompt carrying the failing command and its output
    verbatim; then step 2, then this step's lint lines and tests again, `<v>` plus one. The
    autosquash keeps the archive commit last; a task commit dropped after it has its task unticked
@@ -242,8 +252,7 @@ any other project's run, `self-review-<name>` based at `origin/<agents-base>`.
    land by `<agents repo>`'s `## default landing route`
    (`project-get.sh <agents repo> 'default landing route'`), without asking — merge and push as
    step 4 of **Pipeline defects found mid-run** states it, with `self-review-<name>` as the
-   branch. A rebase onto a moved `origin/<agents-base>` runs item 4's lint lines and tests again
-   before the push — never its ticks or archive, which the rebase carries. The pass is not done until every fixed finding's commit is on `origin/<agents-base>`. Each fixed
+   branch. The pass is not done until every fixed finding's commit is on `origin/<agents-base>`. Each fixed
    finding's sha is read off `<agents-base>` after the landing, never from
    the branch before its rebase; on a route that opens a pull request, once that pull request
    merges.
diff --git a/stats/internal/guard/kickoffworktree.go b/stats/internal/guard/kickoffworktree.go
index 39630fc6..9b096621 100644
--- a/stats/internal/guard/kickoffworktree.go
+++ b/stats/internal/guard/kickoffworktree.go
@@ -154,8 +154,23 @@ func kickoffWorktree(args []string, env Env, stdout, stderr io.Writer) int {
 	}
 
 	// 3. The project's worktree setup, fenced lines only, in order.
-	// stdout alone is the section body; a warning project-get prints on
-	// stderr is relayed, never parsed as a command.
+	if rc := kwSetup(env, projectGet, wt, self, stderr); rc != 0 {
+		return rc
+	}
+
+	// 4. The branch exists on the remote from its first minute.
+	if rc, out := kwRun(env, "", git, "-C", wt, "push", "-u", "origin", branch); rc != 0 {
+		return relay(1, out)
+	}
+	fmt.Fprintf(stdout, "worktree: %s\nmerge-base: %s\n", wt, mergeBase)
+	return 0
+}
+
+// kwSetup runs wt's `## worktree setup` fenced lines, in order, from wt:
+// 0 done or none declared, 1 a command failed (named, its output relayed),
+// 2 project-get cannot answer. stdout alone is the section body; a warning
+// project-get prints on stderr is relayed, never parsed as a command.
+func kwSetup(env Env, projectGet, wt, self string, stderr io.Writer) int {
 	get := exec.Command(projectGet, wt, "worktree setup")
 	get.Dir = env.Dir
 	var body, getErr bytes.Buffer
@@ -166,19 +181,15 @@ func kickoffWorktree(args []string, env Env, stdout, stderr io.Writer) int {
 		for _, c := range kwFenced(body.String()) {
 			if rc, cout := kwRun(env, wt, "bash", "-c", c); rc != 0 {
 				fmt.Fprintf(stderr, "%s: worktree setup command failed (exit %d): %s\n", self, rc, c)
-				return relay(1, cout)
+				fmt.Fprint(stderr, cout)
+				return 1
 			}
 		}
 	case 1:
 		fmt.Fprintf(stderr, "%s: no ## worktree setup declared — skipped\n", self)
 	default:
-		return relay(2, getErr.String())
-	}
-
-	// 4. The branch exists on the remote from its first minute.
-	if rc, out := kwRun(env, "", git, "-C", wt, "push", "-u", "origin", branch); rc != 0 {
-		return relay(1, out)
+		fmt.Fprint(stderr, getErr.String())
+		return 2
 	}
-	fmt.Fprintf(stdout, "worktree: %s\nmerge-base: %s\n", wt, mergeBase)
 	return 0
 }
diff --git a/stats/internal/guard/pipeline_fix_test.go b/stats/internal/guard/pipeline_fix_test.go
new file mode 100644
index 00000000..55b642f5
--- /dev/null
+++ b/stats/internal/guard/pipeline_fix_test.go
@@ -0,0 +1,239 @@
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
+// pfFx is one pipeline-fix fixture: a bare origin, the agents main checkout
+// cloned from it (the pairs table, a .flow/project.md, a script and its
+// harness), a second clone that moves origin, and a `flow` stub logging every
+// call, its stdin included, and answering `record dispatches` from rows.
+type pfFx struct {
+	lpFx
+	origin, other string
+}
+
+const pfProject = "# fixture\n\n## worktree setup\n\n```bash\ntouch .setup-ran\n```\n\n" +
+	"## lint\n\n```bash\ntrue\ncd stats && false\n```\n\n## default landing route\n\nmerge and push\n"
+
+func pfNew(t *testing.T, rows string) pfFx {
+	t.Helper()
+	tmp, err := filepath.EvalSymlinks(t.TempDir())
+	if err != nil {
+		t.Fatal(err)
+	}
+	fx := pfFx{lpFx: lpFx{repo: tmp + "/agents", log: tmp + "/flow.log", rows: tmp + "/rows.json"},
+		origin: tmp + "/origin.git", other: tmp + "/other"}
+	skill, err := os.ReadFile("../../../" + lpSkill)
+	if err != nil {
+		t.Fatal(err)
+	}
+	gitRun(t, "", "init", "-q", "--bare", "-b", "main", fx.origin)
+	gitRun(t, "", "clone", "-q", fx.origin, fx.repo)
+	writeFile(t, fx.repo+"/"+lpSkill, string(skill))
+	writeFile(t, fx.repo+"/.flow/project.md", pfProject)
+	writeFile(t, fx.repo+"/.gitignore", ".setup-ran\n.superpowers/\n")
+	writeExec(t, fx.repo+"/scripts/foo.sh", "#!/usr/bin/env bash\necho foo\n")
+	writeExec(t, fx.repo+"/scripts/test-foo.sh", "#!/usr/bin/env bash\nscripts/foo.sh | grep -q foo\n")
+	gitRun(t, fx.repo, "add", "-A")
+	gitRun(t, fx.repo, "commit", "-q", "-m", "init")
+	gitRun(t, fx.repo, "push", "-q", "origin", "HEAD:main")
+	gitRun(t, fx.repo, "remote", "set-head", "origin", "main")
+	gitRun(t, "", "clone", "-q", fx.origin, fx.other)
+	if rows != "" {
+		writeFile(t, fx.rows, rows)
+	}
+	bin := tmp + "/bin"
+	writeExec(t, bin+"/flow", `#!/usr/bin/env bash
+printf '%s\n' "$*" >>`+fx.log+`
+case "$1 $2" in
+  "record dispatches") [ -f `+fx.rows+` ] && exec cat `+fx.rows+`; exit 1 ;;
+  "state set") cat >>`+fx.log+`; echo >>`+fx.log+` ;;
+esac
+`)
+	scripts := tcfScriptsDir(t)
+	fx.env = Env{Dir: tmp, Getenv: func(k string) string {
+		switch k {
+		case "FLOW_GUARD_REPO_ROOT":
+			return fx.repo
+		case "FLOW_GUARD_PROJECT_GET":
+			return scripts + "/project-get.sh"
+		}
+		return pathEnv(bin)(k)
+	}}
+	return fx
+}
+
+func (fx pfFx) pf(t *testing.T, args ...string) (int, string) {
+	t.Helper()
+	var o, e bytes.Buffer
+	rc := pipelineFix(args, fx.env, &o, &e)
+	return rc, o.String() + e.String()
+}
+
+func (fx pfFx) wt(branch string) string {
+	return filepath.Join(filepath.Dir(fx.repo), "agents-worktrees", branch)
+}
+
+// pfIdentity gives the guard's own git calls (a rebase commits) an identity:
+// they run under the process environment, so these tests do not run in parallel.
+func pfIdentity(t *testing.T) {
+	for _, k := range []string{"GIT_AUTHOR_NAME", "GIT_COMMITTER_NAME"} {
+		t.Setenv(k, "test")
+	}
+	for _, k := range []string{"GIT_AUTHOR_EMAIL", "GIT_COMMITTER_EMAIL"} {
+		t.Setenv(k, "test@example.com")
+	}
+}
+
+func TestPipelineFixStart(t *testing.T) {
+	pfIdentity(t)
+	for _, tc := range []struct {
+		args   []string
+		branch string
+	}{
+		{[]string{"start", "kan-1-demo", "6"}, lpBranch},
+		{[]string{"start", "kan-1-demo"}, "self-review-kan-1-demo"},
+	} {
+		fx := pfNew(t, "")
+		rc, out := fx.pf(t, tc.args...)
+		wt := fx.wt(tc.branch)
+		if rc != 0 || !strings.Contains(out, "worktree: "+wt) {
+			t.Fatalf("%v: rc %d: %s", tc.args, rc, out)
+		}
+		if b, _ := os.ReadFile(wt + "/.git"); len(b) == 0 {
+			t.Fatalf("%v: no worktree at %s", tc.args, wt)
+		}
+		for _, p := range []string{wt + "/.superpowers/sdd/" + tc.branch, wt + "/.setup-ran"} {
+			if _, err := os.Stat(p); err != nil {
+				t.Fatalf("%v: %v", tc.args, err)
+			}
+		}
+		calls := fx.calls(t)
+		if !strings.Contains(calls, "state set -C "+wt+" "+tc.branch) || !strings.Contains(calls, `"state":"STARTED"`) {
+			t.Fatalf("%v: no STARTED record for the loop: %s", tc.args, calls)
+		}
+	}
+}
+
+func TestPipelineFixEndAndEscalate(t *testing.T) {
+	pfIdentity(t)
+	fx := pfNew(t, lpRowsJSON("pipeline-fix-6-", lpBranch, "plan completed"))
+	gitRun(t, fx.repo, "branch", lpBranch)
+	if rc, out := fx.pf(t, "end", "pipeline-fix-6-review-1", "completed"); rc != 2 {
+		t.Fatalf("a review closed completed: rc %d: %s", rc, out)
+	}
+	if rc, out := fx.pf(t, "end", "pipeline-fix-6-review-1", "clean"); rc != 0 ||
+		!strings.Contains(fx.calls(t), "record dispatch end -C "+fx.repo+" -change "+lpBranch+" -key pipeline-fix-6-review-1 -session-token "+lpBranch+" -outcome clean") {
+		t.Fatalf("rc %d: %s\n%s", rc, out, fx.calls(t))
+	}
+	if rc, out := fx.pf(t, "escalate", "pipeline-fix-6-fix-1"); rc != 1 || !strings.Contains(out, "no completed earlier round") {
+		t.Fatalf("a first-round escalation: rc %d: %s", rc, out)
+	}
+	if rc, out := fx.pf(t, "escalate", "pipeline-fix-6-review-2"); rc != 1 || !strings.Contains(out, "admits no escalation") {
+		t.Fatalf("a review escalation: rc %d: %s", rc, out)
+	}
+	if rc, out := fx.pf(t, "escalate", "pipeline-fix-6-plan-2"); rc != 0 {
+		t.Fatalf("a re-plan escalation: rc %d: %s", rc, out)
+	}
+	if !lpEscalated(fx.env, fixtureGit, fx.repo, "pipeline-fix-6-plan-2") {
+		t.Fatal("the escalation was not recorded where check-loop-dispatch reads it")
+	}
+}
+
+func TestPipelineFixVerify(t *testing.T) {
+	pfIdentity(t)
+	fx := pfNew(t, "")
+	if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 0 {
+		t.Fatal(out)
+	}
+	wt := fx.wt(lpBranch)
+	writeExec(t, wt+"/scripts/foo.sh", "#!/usr/bin/env bash\necho foo bar\n")
+	gitRun(t, wt, "commit", "-q", "-am", "fix: foo")
+	rc, out := fx.pf(t, "verify", wt)
+	if rc != 0 || !strings.Contains(out, "green: true") || !strings.Contains(out, "green: scripts/test-foo.sh") || strings.Contains(out, "cd stats") {
+		t.Fatalf("rc %d: %s", rc, out)
+	}
+	writeExec(t, wt+"/scripts/foo.sh", "#!/usr/bin/env bash\necho bar\n")
+	gitRun(t, wt, "commit", "-q", "-am", "fix: break foo")
+	if rc, out := fx.pf(t, "verify", wt); rc != 1 || !strings.Contains(out, "red: scripts/test-foo.sh (exit 1)") {
+		t.Fatalf("a red harness: rc %d: %s", rc, out)
+	}
+}
+
+func TestPipelineFixLand(t *testing.T) {
+	pfIdentity(t)
+	reviewed := lpRowsJSON("pipeline-fix-6-", lpBranch, "plan completed", "fix completed", "review-1 clean")
+	for _, tc := range []struct {
+		name, rows string
+		moved      bool
+		rc         int
+	}{
+		{name: "a reviewed loop lands", rows: reviewed},
+		{name: "a moved base is rebased and verified first", rows: reviewed, moved: true},
+		{name: "an unreviewed loop is refused", rows: lpRowsJSON("pipeline-fix-6-", lpBranch, "plan completed", "fix completed"), rc: 1},
+	} {
+		t.Run(tc.name, func(t *testing.T) {
+			fx := pfNew(t, tc.rows)
+			if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 0 {
+				t.Fatal(out)
+			}
+			wt := fx.wt(lpBranch)
+			writeFile(t, wt+"/docs/fix.md", "fix\n")
+			gitRun(t, wt, "add", "docs/fix.md")
+			gitRun(t, wt, "commit", "-q", "-m", "docs: fix")
+			if tc.moved {
+				writeFile(t, fx.other+"/docs/other.md", "other\n")
+				gitRun(t, fx.other, "add", "docs/other.md")
+				gitRun(t, fx.other, "commit", "-q", "-m", "docs: other")
+				gitRun(t, fx.other, "push", "-q", "origin", "HEAD:main")
+			}
+			rc, out := fx.pf(t, "land", wt)
+			if rc != tc.rc {
+				t.Fatalf("rc %d, want %d: %s", rc, tc.rc, out)
+			}
+			_, log := kwRun(fx.env, "", fixtureGit, "--git-dir", fx.origin, "log", "--format=%s", "main")
+			if tc.rc != 0 {
+				if strings.Contains(log, "docs: fix") || !strings.Contains(out, "LOOP-LANDING-REFUSED: REVIEW") {
+					t.Fatalf("refused landing pushed or named no review: %s\n%s", out, log)
+				}
+				return
+			}
+			if !strings.Contains(log, "docs: fix") || !strings.Contains(out, "landed: ") {
+				t.Fatalf("not landed: %s\n%s", out, log)
+			}
+			if tc.moved && (!strings.Contains(log, "docs: other") || !strings.Contains(out, "green: true")) {
+				t.Fatalf("the moved base was not rebased onto and verified: %s\n%s", out, log)
+			}
+			if _, err := os.Stat(wt); err == nil || lpBranchExists(fx.env, fixtureGit, fx.repo, lpBranch) {
+				t.Fatalf("worktree or branch left behind: %s", out)
+			}
+			if !strings.Contains(fx.calls(t), `"state":"FINISHED"`) {
+				t.Fatalf("no FINISHED record: %s", fx.calls(t))
+			}
+		})
+	}
+}
+
+func TestPipelineFixDrop(t *testing.T) {
+	pfIdentity(t)
+	fx := pfNew(t, "")
+	if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 0 {
+		t.Fatal(out)
+	}
+	wt := fx.wt(lpBranch)
+	writeFile(t, wt+"/scratch.txt", "left\n")
+	if rc, out := fx.pf(t, "drop", wt); rc != 0 || !strings.Contains(out, "dropped: "+lpBranch) {
+		t.Fatalf("rc %d: %s", rc, out)
+	}
+	if _, err := os.Stat(wt); err == nil || lpBranchExists(fx.env, fixtureGit, fx.repo, lpBranch) {
+		t.Fatal("worktree or branch left behind")
+	}
+	if !strings.Contains(fx.calls(t), `"withdrawn":true`) {
+		t.Fatalf("no withdrawn record: %s", fx.calls(t))
+	}
+}
diff --git a/stats/internal/guard/pipelinefix.go b/stats/internal/guard/pipelinefix.go
new file mode 100644
index 00000000..91a5a721
--- /dev/null
+++ b/stats/internal/guard/pipelinefix.go
@@ -0,0 +1,451 @@
+package guard
+
+import (
+	"bytes"
+	"encoding/json"
+	"fmt"
+	"io"
+	"os"
+	"os/exec"
+	"path/filepath"
+	"slices"
+	"strings"
+)
+
+// pipeline-fix.sh: the fixed mechanics of a pipeline-fix or self-review fix
+// loop (skills/flow-contracts/pipeline.md **Pipeline defects found
+// mid-run**), so no step is retyped or skipped. The shim's header is the
+// contract.
+func init() {
+	Registry["pipeline-fix"] = pipelineFix
+}
+
+const pfUsage = `usage: pipeline-fix.sh start <name> [<k>]
+       pipeline-fix.sh end <key> <outcome>
+       pipeline-fix.sh escalate <key>
+       pipeline-fix.sh verify <worktree>
+       pipeline-fix.sh land <worktree>
+       pipeline-fix.sh drop <worktree>
+`
+
+// pfEnv is what every subcommand resolves first.
+type pfEnv struct {
+	git, main, projectGet string
+}
+
+func pipelineFix(args []string, env Env, stdout, stderr io.Writer) int {
+	const self = "pipeline-fix"
+	if len(args) == 0 {
+		fmt.Fprint(stderr, pfUsage)
+		return 2
+	}
+	repo, projectGet := env.Getenv("FLOW_GUARD_REPO_ROOT"), env.Getenv("FLOW_GUARD_PROJECT_GET")
+	if repo == "" || projectGet == "" {
+		fmt.Fprintf(stderr, "%s: FLOW_GUARD_REPO_ROOT or FLOW_GUARD_PROJECT_GET is unset — run scripts/pipeline-fix.sh, which sets them\n", self)
+		return 2
+	}
+	git, ok := panelResolveGit(env, self, stderr)
+	if !ok {
+		return 2
+	}
+	main, ok := lpMain(env, git, repo)
+	if !ok {
+		fmt.Fprintf(stderr, "%s: %s is not a git checkout\n", self, repo)
+		return 2
+	}
+	pe := pfEnv{git: git, main: main, projectGet: projectGet}
+	sub, rest := args[0], args[1:]
+	switch {
+	case sub == "start" && (len(rest) == 1 || len(rest) == 2):
+		return pfStart(env, pe, rest, stdout, stderr)
+	case sub == "end" && len(rest) == 2:
+		return pfEnd(env, pe, rest[0], rest[1], stdout, stderr)
+	case sub == "escalate" && len(rest) == 1:
+		return pfEscalate(env, pe, rest[0], stdout, stderr)
+	case sub == "verify" && len(rest) == 1:
+		return pfVerify(env, pe, smcAbs(env, rest[0]), stdout, stderr)
+	case sub == "land" && len(rest) == 1:
+		return pfLand(env, pe, smcAbs(env, rest[0]), stdout, stderr)
+	case sub == "drop" && len(rest) == 1:
+		return pfDrop(env, pe, smcAbs(env, rest[0]), stdout, stderr)
+	}
+	fmt.Fprint(stderr, pfUsage)
+	return 2
+}
+
+// pfState writes the loop change's state record (skills/flow-contracts/state-file.md).
+func pfState(env Env, dir, change string, rec map[string]any) (int, string) {
+	flow, ok := lookPath(env, "flow")
+	if !ok {
+		return 2, "no flow binary on PATH\n"
+	}
+	body, _ := json.Marshal(rec)
+	cmd := exec.Command(flow, "state", "set", "-C", dir, change)
+	cmd.Dir, cmd.Stdin = env.Dir, bytes.NewReader(body)
+	var out bytes.Buffer
+	cmd.Stdout, cmd.Stderr = &out, &out
+	return exitCode(cmd.Run()), out.String()
+}
+
+func pfRecord(state, branch string, worktrees map[string]string) map[string]any {
+	return map[string]any{"state": state, "branch": branch, "worktrees": worktrees,
+		"artifactUrl": nil, "jiraIssue": nil, "prUrl": nil, "updatedBy": "/flow"}
+}
+
+// pfBase is origin/HEAD's short name, e.g. origin/main.
+func pfBase(env Env, pe pfEnv) (string, bool) {
+	rc, out := kwRun(env, "", pe.git, "-C", pe.main, "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
+	return strings.TrimSpace(out), rc == 0 && strings.HasPrefix(out, "origin/")
+}
+
+func pfStart(env Env, pe pfEnv, rest []string, stdout, stderr io.Writer) int {
+	const self = "pipeline-fix start"
+	name := rest[0]
+	if !plainChangeName(name) || (len(rest) == 2 && !lpDigits.MatchString(rest[1])) {
+		fmt.Fprintf(stderr, "%s: <name> must be a plain change name and <k> a number\n", self)
+		return 2
+	}
+	branch := "self-review-" + name
+	if len(rest) == 2 {
+		branch = "pipeline-fix-" + name + "-" + rest[1]
+	}
+	if rc, out := kwRun(env, "", pe.git, "-C", pe.main, "fetch", "origin"); rc != 0 {
+		fmt.Fprint(stderr, out)
+		return 1
+	}
+	base, ok := pfBase(env, pe)
+	if !ok {
+		fmt.Fprintf(stderr, "%s: refs/remotes/origin/HEAD is unset (git remote set-head origin --auto)\n", self)
+		return 2
+	}
+	physical, err := filepath.EvalSymlinks(pe.main)
+	if err != nil {
+		fmt.Fprintf(stderr, "%s: %v\n", self, err)
+		return 2
+	}
+	wt := filepath.Join(siblingRoot(physical), branch)
+	if rc, out := kwRun(env, "", pe.git, "-C", pe.main, "worktree", "add", "-b", branch, wt, base); rc != 0 {
+		fmt.Fprint(stderr, out)
+		return 1
+	}
+	// The loop's change must have a checkout before its first row: flow
+	// record refuses a change no worktree carries, and spectre/changes/<loop>/
+	// only exists once the planner commits.
+	if err := os.MkdirAll(filepath.Join(wt, ".superpowers", "sdd", branch), 0o755); err != nil {
+		fmt.Fprintf(stderr, "%s: %v\n", self, err)
+		return 1
+	}
+	rc, out := kwRun(env, "", pe.git, "-C", wt, "rev-parse", "HEAD")
+	if rc != 0 {
+		fmt.Fprint(stderr, out)
+		return 1
+	}
+	if rc, out := pfState(env, wt, branch, pfRecord("STARTED", branch, map[string]string{wt: strings.TrimSpace(out)})); rc != 0 {
+		fmt.Fprint(stderr, out)
+		return 1
+	}
+	if rc := kwSetup(env, pe.projectGet, wt, self, stderr); rc != 0 {
+		return rc
+	}
+	fmt.Fprintf(stdout, "worktree: %s\n", wt)
+	return 0
+}
+
+// pfLoop resolves key's loop for end and escalate.
+func pfLoop(env Env, pe pfEnv, key string, stderr io.Writer) (lpLoop, string, lpRow, int) {
+	rows, err := lpTable(pe.main)
+	if err != nil {
+		fmt.Fprintf(stderr, "pipeline-fix: %v\n", err)
+		return lpLoop{}, "", lpRow{}, 2
+	}
+	lp, suffix, row, msg, ok := lpFromKey(env, pe.git, pe.main, key, rows)
+	if !ok {
+		if msg == "" {
+			msg = key + " is no loop key"
+		}
+		fmt.Fprintf(stderr, "pipeline-fix: %s\n", msg)
+		return lpLoop{}, "", lpRow{}, 1
+	}
+	return lp, suffix, row, 0
+}
+
+func pfEnd(env Env, pe pfEnv, key, outcome string, stdout, stderr io.Writer) int {
+	lp, suffix, _, rc := pfLoop(env, pe, key, stderr)
+	if rc != 0 {
+		return rc
+	}
+	want := []string{"completed", "aborted"}
+	if lpFamily(suffix) == "review" {
+		// clean: the review raised no Critical and no Important.
+		want = []string{"clean", "fix"}
+	}
+	if !slices.Contains(want, outcome) {
+		fmt.Fprintf(stderr, "pipeline-fix end: %s takes outcome %s\n", key, strings.Join(want, " or "))
+		return 2
+	}
+	out, _, frc := pcFlow(env, true, "record", "dispatch", "end", "-C", pe.main, "-change", lp.change,
+		"-key", key, "-session-token", lp.token, "-outcome", outcome)
+	fmt.Fprint(stdout, string(out))
+	if frc != 0 {
+		return 1
+	}
+	return 0
+}
+
+func pfEscalate(env Env, pe pfEnv, key string, stdout, stderr io.Writer) int {
+	lp, suffix, row, rc := pfLoop(env, pe, key, stderr)
+	if rc != 0 {
+		return rc
+	}
+	if len(row.escal) == 0 {
+		fmt.Fprintf(stderr, "pipeline-fix escalate: the pairs table admits no escalation for %s\n", key)
+		return 1
+	}
+	rows, ok := lpRows(env, pe.main, lp)
+	if !ok {
+		fmt.Fprintf(stderr, "pipeline-fix escalate: flow record dispatches -change %s failed -- cannot answer\n", lp.change)
+		return 2
+	}
+	// No progress needs an earlier round of the same kind to have made none.
+	isFix := strings.HasPrefix(lpFamily(suffix), "fix")
+	earlier := false
+	for _, d := range rows {
+		f := lpFamily(strings.TrimPrefix(d.Key, lp.prefix))
+		if d.Outcome == "completed" && (f == lpFamily(suffix) || (isFix && strings.HasPrefix(f, "fix"))) {
+			earlier = true
+		}
+	}
+	if !earlier {
+		fmt.Fprintf(stderr, "pipeline-fix escalate: %s has no completed earlier round to escalate from\n", key)
+		return 1
+	}
+	if rc, out := kwRun(env, "", pe.git, "-C", pe.main, "config", "--add", "flow.escalation", key); rc != 0 {
+		fmt.Fprint(stderr, out)
+		return 1
+	}
+	fmt.Fprintf(stdout, "escalated: %s\n", key)
+	return 0
+}
+
+// pfTouched is every path wt's branch changed against base.
+func pfTouched(env Env, pe pfEnv, wt, base string) ([]string, bool) {
+	rc, out := kwRun(env, "", pe.git, "-C", wt, "diff", "--no-renames", "--name-only", base+"...HEAD")
+	return strings.Fields(out), rc == 0
+}
+
+// pfChecks is the lint lines and tests the touched paths need: every guard
+// line of `## lint`, its `cd stats` lines only for touched Go and its
+// `cd stats/web` lines only for a touched SPA file; `go test` of each
+// touched package, the scripts/test-*.sh harness of a touched script or
+// hook, and vitest of a touched SPA test file or the test beside a touched
+// source file.
+func pfChecks(env Env, pe pfEnv, wt string, touched []string) ([]string, int) {
+	get := exec.Command(pe.projectGet, wt, "lint")
+	var body bytes.Buffer
+	get.Stdout = &body
+	if rc := exitCode(get.Run()); rc != 0 {
+		return nil, 2
+	}
+	has := func(pred func(string) bool) bool { return slices.ContainsFunc(touched, pred) }
+	goTouched := has(func(p string) bool { return strings.HasPrefix(p, "stats/") && strings.HasSuffix(p, ".go") })
+	webTouched := has(func(p string) bool { return strings.HasPrefix(p, "stats/web/") })
+	var cmds []string
+	for _, l := range kwFenced(body.String()) {
+		switch {
+		case strings.HasPrefix(l, "cd stats/web"):
+			if webTouched {
+				cmds = append(cmds, l)
+			}
+		case strings.HasPrefix(l, "cd stats"):
+			if goTouched {
+				cmds = append(cmds, l)
+			}
+		default:
+			cmds = append(cmds, l)
+		}
+	}
+	exists := func(rel string) bool { return isFile(filepath.Join(wt, rel)) }
+	add := func(c string) {
+		if !slices.Contains(cmds, c) {
+			cmds = append(cmds, c)
+		}
+	}
+	for _, p := range touched {
+		base := filepath.Base(p)
+		stem := strings.TrimSuffix(base, filepath.Ext(base))
+		switch {
+		case strings.HasPrefix(p, "stats/web/"):
+			rel := strings.TrimPrefix(p, "stats/web/")
+			ext := filepath.Ext(p)
+			sibling := strings.TrimSuffix(rel, ext) + ".test" + ext
+			switch {
+			case strings.Contains(base, ".test.") && exists(p):
+				add("cd stats/web && npx vitest run " + rel)
+			case (ext == ".ts" || ext == ".tsx") && exists("stats/web/"+sibling):
+				add("cd stats/web && npx vitest run " + sibling)
+			}
+		case strings.HasPrefix(p, "stats/") && strings.HasSuffix(p, ".go"):
+			if isDir(filepath.Join(wt, filepath.Dir(p))) {
+				add("cd stats && go test ./" + strings.TrimPrefix(filepath.Dir(p), "stats/") + "/ -count=1")
+			}
+		case strings.HasPrefix(p, "scripts/test-") && strings.HasSuffix(p, ".sh") && exists(p):
+			add(p)
+		case strings.HasPrefix(p, "scripts/") || strings.HasPrefix(p, "hooks/"):
+			for _, h := range []string{"scripts/test-" + stem + ".sh", "scripts/test-" + stem + "-hook.sh"} {
+				if exists(h) {
+					add(h)
+				}
+			}
+		case p == "setup.sh":
+			add("cd stats && go test ./internal/setuptest/ -count=1")
+		}
+	}
+	return cmds, 0
+}
+
+func pfVerify(env Env, pe pfEnv, wt string, stdout, stderr io.Writer) int {
+	base, ok := pfBase(env, pe)
+	if !ok {
+		fmt.Fprint(stderr, "pipeline-fix verify: refs/remotes/origin/HEAD is unset\n")
+		return 2
+	}
+	touched, ok := pfTouched(env, pe, wt, base)
+	if !ok {
+		fmt.Fprintf(stderr, "pipeline-fix verify: cannot diff %s against %s\n", wt, base)
+		return 2
+	}
+	cmds, rc := pfChecks(env, pe, wt, touched)
+	if rc != 0 {
+		fmt.Fprintf(stderr, "pipeline-fix verify: project-get cannot read %s's ## lint\n", wt)
+		return 2
+	}
+	red := 0
+	for _, c := range cmds {
+		crc, out := kwRun(env, wt, "bash", "-c", c)
+		if crc == 0 {
+			fmt.Fprintf(stdout, "green: %s\n", c)
+			continue
+		}
+		red++
+		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
+		if len(lines) > 40 {
+			lines = lines[len(lines)-40:]
+		}
+		fmt.Fprintf(stdout, "red: %s (exit %d)\n%s\n", c, crc, strings.Join(lines, "\n"))
+	}
+	if red > 0 {
+		return 1
+	}
+	return 0
+}
+
+// pfLoopWorktree is wt's loop branch, or ok false (named on stderr).
+func pfLoopWorktree(env Env, pe pfEnv, wt, self string, stderr io.Writer) (string, lpLoop, bool) {
+	rc, out := kwRun(env, "", pe.git, "-C", wt, "branch", "--show-current")
+	branch := strings.TrimSpace(out)
+	lp, ok := lpFromBranch(branch)
+	if rc != 0 || !ok {
+		fmt.Fprintf(stderr, "%s: %s is not on a pipeline-fix-* or self-review-* branch\n", self, wt)
+		return "", lpLoop{}, false
+	}
+	return branch, lp, true
+}
+
+// pfRemove removes wt and its branch from the main checkout.
+func pfRemove(env Env, pe pfEnv, wt, branch string, force bool, stderr io.Writer) int {
+	rm := []string{"-C", pe.main, "worktree", "remove", wt}
+	if force {
+		rm = append(rm, "--force")
+	}
+	for _, a := range [][]string{rm, {"-C", pe.main, "branch", "-D", branch}} {
+		if rc, out := kwRun(env, "", pe.git, a...); rc != 0 {
+			fmt.Fprint(stderr, out)
+			return 1
+		}
+	}
+	return 0
+}
+
+func pfLand(env Env, pe pfEnv, wt string, stdout, stderr io.Writer) int {
+	const self = "pipeline-fix land"
+	branch, lp, ok := pfLoopWorktree(env, pe, wt, self, stderr)
+	if !ok {
+		return 2
+	}
+	get := exec.Command(pe.projectGet, wt, "default landing route")
+	var route bytes.Buffer
+	get.Stdout = &route
+	_ = get.Run()
+	if first, _, _ := strings.Cut(strings.TrimSpace(route.String()), "\n"); strings.TrimSpace(first) != "merge and push" {
+		fmt.Fprintf(stderr, "%s: the default landing route is %q, not merge and push — land it by that route\n", self, first)
+		return 1
+	}
+	base, ok := pfBase(env, pe)
+	if !ok {
+		fmt.Fprintf(stderr, "%s: refs/remotes/origin/HEAD is unset\n", self)
+		return 2
+	}
+	if rc, out := kwRun(env, "", pe.git, "-C", wt, "fetch", "origin"); rc != 0 {
+		fmt.Fprint(stderr, out)
+		return 1
+	}
+	if rc, _ := kwRun(env, "", pe.git, "-C", wt, "merge-base", "--is-ancestor", base, "HEAD"); rc != 0 {
+		if rc, out := kwRun(env, "", pe.git, "-C", wt, "rebase", base); rc != 0 {
+			kwRun(env, "", pe.git, "-C", wt, "rebase", "--abort")
+			fmt.Fprintf(stderr, "%s: rebase onto %s stopped — resolve it in %s, then land again\n%s", self, base, wt, out)
+			return 1
+		}
+		// The base moved: the lint lines and tests run again before the push.
+		if rc := pfVerify(env, pe, wt, stdout, stderr); rc != 0 {
+			fmt.Fprintf(stderr, "%s: red after the rebase onto %s — nothing pushed\n", self, base)
+			return 1
+		}
+	}
+	rows, ok := lpRows(env, pe.main, lp)
+	if !ok {
+		fmt.Fprintf(stderr, "%s: flow record dispatches -change %s failed -- cannot answer\n", self, lp.change)
+		return 2
+	}
+	if refusals := lpLandingRefusals(lp, rows); len(refusals) > 0 {
+		for _, r := range refusals {
+			fmt.Fprintf(stdout, "LOOP-LANDING-REFUSED: %s\n", r)
+		}
+		return 1
+	}
+	target := strings.TrimPrefix(base, "origin/")
+	if rc, out := kwRun(env, "", pe.git, "-C", wt, "push", "origin", branch+":"+target); rc != 0 {
+		fmt.Fprint(stderr, out)
+		return 1
+	}
+	_, sha := kwRun(env, "", pe.git, "-C", wt, "rev-parse", "HEAD")
+	if rc, out := pfState(env, pe.main, lp.change, pfRecord("FINISHED", branch, map[string]string{})); rc != 0 {
+		fmt.Fprintf(stderr, "%s: pushed, but the loop's state was not written:\n%s", self, out)
+	}
+	if rc, out := kwRun(env, "", pe.git, "-C", pe.main, "pull", "--ff-only"); rc != 0 {
+		fmt.Fprintf(stderr, "%s: pushed, but the main checkout did not fast-forward:\n%s", self, out)
+	}
+	if rc := pfRemove(env, pe, wt, branch, false, stderr); rc != 0 {
+		return rc
+	}
+	fmt.Fprintf(stdout, "landed: %s\n", strings.TrimSpace(sha))
+	return 0
+}
+
+func pfDrop(env Env, pe pfEnv, wt string, stdout, stderr io.Writer) int {
+	const self = "pipeline-fix drop"
+	branch, lp, ok := pfLoopWorktree(env, pe, wt, self, stderr)
+	if !ok {
+		return 2
+	}
+	rec := pfRecord("FINISHED", branch, map[string]string{})
+	rec["withdrawn"] = true
+	if rc, out := pfState(env, pe.main, lp.change, rec); rc != 0 {
+		fmt.Fprint(stderr, out)
+		return 1
+	}
+	if rc := pfRemove(env, pe, wt, branch, true, stderr); rc != 0 {
+		return rc
+	}
+	fmt.Fprintf(stdout, "dropped: %s\n", branch)
+	return 0
+}
````

  - [x] **Step 4: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestPipelineFix|TestKickoffWorktree|TestCheckLoop' -count=1`, then from the repository root `bash -n scripts/pipeline-fix.sh`, `scripts/test-git-config-pins.sh`, `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-installed-citations.sh`, `scripts/check-guard-symlinks.sh`, `scripts/check-normative-inventory.sh >/dev/null` — all exit 0 with `gofmt -l` printing nothing. Commit.

**Build:** green
**Files:** `scripts/pipeline-fix.sh`, `skills/flow-contracts/pipeline.md`, `skills/flow-self-review/SKILL.md`, `stats/internal/guard/kickoffworktree.go`, `stats/internal/guard/pipeline_fix_test.go`, `stats/internal/guard/pipelinefix.go`, `scripts/integrate-change.sh`, `scripts/test-integrate-change.sh`, `skills/flow/integrate.md`, `skills/flow/scripts/check-loop-landing.sh`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`
**Tests:** `TestPipelineFixStart`, `TestPipelineFixEndAndEscalate`, `TestPipelineFixVerify`, `TestPipelineFixLand`, `TestPipelineFixDrop`
**Regression:** reverting the commit leaves the loops' worktree, verification and landing steps as prose the parent types by hand, with no `end` or `escalate` to satisfy the dispatch check.
**Baseline:** before=182 after=187
<!-- measured: cat stats/internal/guard/*_test.go | grep -c '^func Test' (before and after with Tasks 1–3 applied in the scratch clone at a1bca81b) -->
**Commit:** feat(scripts): run the fix loop's fixed mechanics from pipeline-fix.sh
**After:** Task 1, 2

**Decision:** close-with-outcome
**Decision:** escalation-marker
**Decision:** start-optional-k
**Decision:** verify-lines-by-path
**Decision:** drop-subcommand

