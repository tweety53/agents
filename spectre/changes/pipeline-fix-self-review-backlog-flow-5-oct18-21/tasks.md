# pipeline-fix-self-review-backlog-flow-5-oct18-21

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** a `/flow` change named `self-review-…` can dispatch again; a loop dispatch that dropped
its `Key:` line is still denied.

**Architecture:** one `git for-each-ref` check in `hooks/enforce-loop-dispatch.py`'s keyless-name
branch — a token with a `spectre/<token>` branch in the payload `cwd`'s repository is a `/flow`
change's own and is skipped.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/`.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; it was applied to this worktree at 52438f0d and checked there. A diff that does not
  apply is reported, never hand-merged by guesswork.
- No new inline suppressions; `scripts/check-python-suppressions.sh` exits 0.
- Read-only against the dev store: never stop flowd or flow-postgres, never drop the `flow` DB.

## Review Focus

- A loop dispatch that dropped its Key line is still denied: case 8 (no `cwd`), case 13's deny
  twin (a `cwd` without the `spectre/` branch).
- A prompt naming both the running `/flow` change and a real loop worktree is denied
  (`every-token-checked`).
- A failing or timed-out git call denies (`no-cwd-keeps-deny`).

- [ ] 1. Tell a `/flow` change named `self-review-…` from a loop by its `spectre/<name>` branch
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** From this worktree pipe the
    real slot prompt as an Agent payload into the untouched hook:
    `python3 -c 'import json,sys; print(json.dumps({"tool_name":"Agent","cwd":"/Users/tweety53/Projects/agents","tool_input":{"prompt":open(sys.argv[1]).read()}}))' '/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-5-oct18/.superpowers/sdd/slot-prompt-0-primary+principles.md' | python3 hooks/enforce-loop-dispatch.py`
    prints a deny naming `self-review-backlog-flow-5-oct18`. Save the diff below to
    `/tmp/pf21-task1.diff`, apply its test hunk alone
    (`git apply --include='scripts/*' /tmp/pf21-task1.diff`) and run
    `bash scripts/test-enforce-loop-dispatch-hook.sh`: `FAIL: 13 self-review-named /flow change`,
    exit 1; every other case passes.
  - [ ] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='scripts/*' /tmp/pf21-task1.diff`):

````diff verified:applied with git apply to this worktree at 52438f0d; scripts/test-enforce-loop-dispatch-hook.sh passes every case with it and fails case 13's allow alone with the hook untouched; check-python-suppressions, check-references and check-guard-symlinks exit 0; the real slot prompt is allowed and the same prompt naming self-review-no-such-loop is still denied
diff --git a/hooks/enforce-loop-dispatch.py b/hooks/enforce-loop-dispatch.py
index ae8c4024..e470e4a9 100644
--- a/hooks/enforce-loop-dispatch.py
+++ b/hooks/enforce-loop-dispatch.py
@@ -19,9 +19,12 @@ Fails open for anything that is not a loop dispatch -- unparseable input, anothe
 prompt with no loop key and no loop worktree or branch named. Fails CLOSED for a loop dispatch: a checker that cannot answer, or
 cannot run, denies it, because an unchecked loop dispatch is the skip this hook exists to stop.
 A prompt that names a loop's worktree or branch (`pipeline-fix-<name>-<k>`, `self-review-<name>`)
-with no Key line is denied too: dropping the line must not drop the check. A loop dispatch whose
-prompt lacks the agent-baseline pointer is denied before the checker records it, since
-enforce-agent-baseline.py would deny it after, leaving its row open.
+with no Key line is denied too: dropping the line must not drop the check -- unless that name is a
+`/flow` change's own, a `spectre/<name>` branch, local or on origin, in the payload `cwd`'s
+repository: a `/flow` change may itself be named `self-review-...`, and a loop's branch is its bare
+name, never `spectre/<name>`. A loop dispatch whose prompt lacks the agent-baseline pointer is
+denied before the checker records it, since enforce-agent-baseline.py would deny it after, leaving
+its row open.
 
 The checker's wait stays well under the 300s timeout setup.sh registers, so a slow checker is
 denied here rather than the harness timing the hook out, which fails open.
@@ -67,6 +70,19 @@ def deny(reason):
     return 0
 
 
+def flow_change(name, cwd):
+    """Is name a /flow change's own -- a spectre/<name> branch, local or on origin, in cwd's repository?"""
+    try:
+        r = subprocess.run(
+            ["git", "-C", cwd, "for-each-ref", "--count=1", "--format=x",
+             f"refs/heads/spectre/{name}", f"refs/remotes/origin/spectre/{name}"],
+            capture_output=True, text=True, timeout=10,
+        )
+    except (OSError, subprocess.SubprocessError):
+        return False
+    return r.returncode == 0 and r.stdout.strip() == "x"
+
+
 def main():
     try:
         payload = json.load(sys.stdin)
@@ -78,10 +94,13 @@ def main():
     if not isinstance(tool_input, dict) or not isinstance(tool_input.get("prompt"), str):
         return 0
     prompt = tool_input["prompt"]
+    cwd = payload.get("cwd")
+    cwd = cwd if isinstance(cwd, str) and os.path.isdir(cwd) else None
     m = KEY_RE.search(prompt)
     if not m:
-        named = LOOP_NAME_RE.search(prompt)
-        if named:
+        for named in LOOP_NAME_RE.finditer(prompt):
+            if cwd and flow_change(named.group(0), cwd):
+                continue
             return deny(
                 f"Loop dispatch blocked: the prompt names {named.group(0)} but carries no "
                 "`Key: pipeline-fix-<k>-<suffix>` or `Key: self-review-<name>-<suffix>` line."
@@ -96,8 +115,6 @@ def main():
     if BASELINE not in prompt:
         return deny(f"Loop dispatch {key} blocked: the prompt carries no {BASELINE} pointer.")
     args = [CHECK, REPO, key, str(tool_input.get("subagent_type") or ""), str(tool_input.get("model") or "")]
-    cwd = payload.get("cwd")
-    cwd = cwd if isinstance(cwd, str) and os.path.isdir(cwd) else None
     try:
         r = subprocess.run(args, capture_output=True, text=True, timeout=CHECK_TIMEOUT, cwd=cwd)
     except (OSError, subprocess.SubprocessError) as e:
diff --git a/scripts/test-enforce-loop-dispatch-hook.sh b/scripts/test-enforce-loop-dispatch-hook.sh
index 17d422d4..cdbb959a 100755
--- a/scripts/test-enforce-loop-dispatch-hook.sh
+++ b/scripts/test-enforce-loop-dispatch-hook.sh
@@ -17,6 +17,9 @@
 #  10. a Key line with trailing text                   -> deny, checker never run
 #  11. a pipeline-fix key with a non-numeric <k>       -> deny, checker never run
 #  12. a payload carrying a cwd                       -> the checker runs in it
+#  13. a prompt naming a /flow change called self-review-<x>, no Key line, its
+#      spectre/<name> branch in the payload's cwd    -> allow, checker never run;
+#      the same prompt in a cwd without that branch  -> deny
 set -euo pipefail
 
 SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
@@ -112,6 +115,21 @@ expect "12 admitted with a cwd" allow
 [ "$(cat "$SANDBOX/cwd")" = "$(cd "$SANDBOX/project" && pwd -P)" ] \
   && pass "12 checker ran in the payload's cwd" || fail "12 checker ran in $(cat "$SANDBOX/cwd")"
 
+# Case 13: a /flow change named self-review-<x> is told from a loop by its spectre/<name> branch.
+cwd_payload() { # <cwd> <prompt>
+  python3 -c 'import json,sys; print(json.dumps({"tool_name": "Agent", "cwd": sys.argv[1], "tool_input": {"prompt": sys.argv[2], "subagent_type": "flow-low", "model": "opus"}}))' "$1" "$2"
+}
+FLOW_PROMPT=$'Review /x/agents-worktrees/self-review-backlog-demo-5 now.'
+git init -q -b main "$SANDBOX/flowrepo"
+git -C "$SANDBOX/flowrepo" -c user.name=t -c user.email=t@t commit -q --allow-empty -m init
+rm -f "$LOG"
+hook "$SANDBOX/check" "$(cwd_payload "$SANDBOX/flowrepo" "$FLOW_PROMPT")"
+expect "13 self-review-named change, no spectre branch" deny "carries no"
+git -C "$SANDBOX/flowrepo" branch spectre/self-review-backlog-demo-5
+hook "$SANDBOX/check" "$(cwd_payload "$SANDBOX/flowrepo" "$FLOW_PROMPT")"
+expect "13 self-review-named /flow change" allow
+[ ! -e "$LOG" ] && pass "13 checker not run" || fail "13 checker ran"
+
 if [ "$FAILURES" -gt 0 ]; then
   printf '%d case(s) failed\n' "$FAILURES" >&2
   exit 1
````

  - [ ] **Step 3: Verify.** `bash scripts/test-enforce-loop-dispatch-hook.sh` exits 0 with every
    case `ok`; `scripts/check-python-suppressions.sh`, `scripts/check-references.sh` and
    `scripts/check-guard-symlinks.sh` exit 0; the Step 1 call now prints nothing (allow), and the
    same call with `.replace("self-review-backlog-flow-5-oct18","self-review-no-such-loop")`
    applied to the prompt still prints a deny naming `self-review-no-such-loop`.
  - [ ] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `hooks/enforce-loop-dispatch.py`, `scripts/test-enforce-loop-dispatch-hook.sh`
**Tests:** Case 13
**Regression:** Case 13 — reverting the commit denies a prompt naming a `/flow` change called `self-review-backlog-demo-5` whose `spectre/` branch exists in the payload's `cwd`: `FAIL: 13 self-review-named /flow change`.
**Baseline:** before=12 after=14
<!-- measured: grep -cE '^expect ' scripts/test-enforce-loop-dispatch-hook.sh (before at 52438f0d; after with this task's diff applied) -->
**Commit:** fix(hooks): tell a /flow change named self-review-… from a loop by its spectre branch
**After:** none

**Decision:** spectre-branch-discriminator
**Decision:** no-cwd-keeps-deny
**Decision:** every-token-checked
