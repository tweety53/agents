# pipeline-fix-kan-939-step-4-frontend-cardio-plan-view-start-log-7

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** `hooks/protect-main-checkout.py` denies a landing push it cannot read — one whose
repository, remote or refspec goes through an unexpanded shell expansion and whose destination is
the default branch or cannot be told — instead of letting it land unchecked.

**Architecture:** one new function, `unreadable_landing`, in the hook's push scan ahead of the two
existing landing checks; harness cases in `scripts/test-protect-main-checkout.sh`.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:`
  trailer; never write `spectre/changes/`.
- The task's change is the unified diff its step 3 gives, applied from the repository root with
  `git apply`; it was applied to this worktree at 881de558 and checked there. A diff that does not
  apply is reported, never hand-merged by guesswork.
- No new inline suppressions (`scripts/check-python-suppressions.sh`).

## Review Focus

- A push with a literal non-default destination and a variable repository
  (`git -C $W push origin spectre/<name>`) still passes — branch backups depend on it.
- `cd $W && git push origin HEAD` is denied: `target` None must never reach `git(None, …)`, which
  would raise outside the `git` helper's caught exceptions and fail the whole hook open.
- `$(…)` tokenizes into separate `$`, `(` … `)` words and a backtick refspec splits across words
  (one of them option-looking); both still deny.

- [x] 1. Deny a default-branch push named through an unexpanded shell expansion
  - [x] **Step 1: Reproduce before touching anything; do not guess.** From the repository root
    run `printf '{"tool_name":"Bash","tool_input":{"command":"git -C $W push origin $B:main"},"cwd":"/tmp"}' | python3 hooks/protect-main-checkout.py; echo "rc=$?"` —
    it prints nothing and `rc=0`: the push is allowed, the defect.
  - [x] **Step 2: Write the failing test.** Save step 3's diff to `/tmp/pf7-task1.diff`, apply its
    harness hunks alone (`git apply --include='scripts/test-protect-main-checkout.sh' /tmp/pf7-task1.diff`)
    and run `bash scripts/test-protect-main-checkout.sh`: it FAILS — every `unreadable landing`
    case, `cd to an unexpanded path, then push HEAD` and `push of an unexpanded branch` print
    `wanted deny, got allow`.
  - [x] **Step 3: Implement.** Apply the rest (`git apply --exclude='scripts/test-protect-main-checkout.sh' /tmp/pf7-task1.diff`):

````diff verified:applied with git apply to this worktree at 881de558; scripts/test-protect-main-checkout.sh passes with it and fails the six deny cases with the hook hunk reverted
diff --git a/hooks/protect-main-checkout.py b/hooks/protect-main-checkout.py
index 21cc93d3..7fb42dce 100755
--- a/hooks/protect-main-checkout.py
+++ b/hooks/protect-main-checkout.py
@@ -35,6 +35,16 @@ denied too unless scripts/check-loop-landing.sh, in the agents checkout this hoo
 from, admits it — flowd holding the loop's completed plan, fix and clean last review. That one
 check fails closed: a checker that cannot answer or cannot run denies the push.
 
+Both landing checks read the push from the command text, so a push the hook cannot read is denied
+rather than let through: one whose repository (`-C <path>`, a `cd`), remote or refspec goes through
+a shell expansion the hook cannot resolve — `$W`, `"${B}:main"`, `$(git branch --show-current)`, a
+backtick — and whose destination is the default branch or cannot be told (0b055e56 landed a
+`pipeline-fix-*` branch with no flowd records through `git -C $W push origin $B:main`). The reason
+asks for literal names. A push whose destination is a literal other branch (`git -C $W push origin
+spectre/<name>`) passes; an assignment the command itself makes (`B=x; git push origin $B:main`) is
+expanded first, so only names from outside the command count. Scripts are untouched — the hook only
+sees the typed Bash command, never what a script it calls runs.
+
 Allowed everywhere: reads, `git pull --ff-only`, `git fetch`, `git checkout -b`, `git switch -c`,
 `git worktree add`, `git push`, and the landing scripts (`land-self-review-report.sh`,
 `refresh-main-checkout.sh`) — they are not `git` verbs in the command text, and each owns its own
@@ -404,6 +414,37 @@ def push_refspecs(args, target):
     return out
 
 
+def unreadable_landing(args, target):
+    """The deny reason when the push may land on the default branch and the hook cannot read it:
+    its repository is unknown (target None, a `-C $W` or `cd $W`), or any of its words holds a
+    shell expansion left unexpanded, and some refspec's destination is the default branch, an
+    expansion, or unknown. Else None."""
+    if target is not None and not any("$" in a or "`" in a for a in args):
+        return None
+    default = default_branch(target) if target else None
+    defaults = {default} if default else FALLBACK_DEFAULT_BRANCHES
+    unknown = any(a.startswith("-") and ("$" in a or "`" in a) for a in args)
+    pos = [a for a in args if not a.startswith("-")]
+    for a in pos[1:] or ["HEAD"]:
+        src, colon, dst = a.lstrip("+").partition(":")
+        dst = (dst if colon else src).removeprefix("refs/heads/")
+        if dst == "HEAD":
+            dst = git(target, "branch", "--show-current") if target else None
+        if not dst or "$" in dst or "`" in dst or dst in defaults:
+            unknown = True
+    if not unknown:
+        return None
+    return (
+        "Blocked: this push may land on the default branch, and part of it — the repository "
+        "(`-C`, `cd`), the remote or a refspec — goes through a shell expansion (`$VAR`, "
+        "`${VAR}`, `$(...)`, a backtick) the hook cannot resolve, so its landing checks (an open "
+        "spectre change, a fix loop's records) cannot see what it lands. Write the push with "
+        "literal names:\n\n"
+        "  git -C /abs/path/to/worktree push origin <branch>:<default-branch>\n\n"
+        "A fix loop lands with scripts/pipeline-fix.sh land <worktree>."
+    )
+
+
 def loop_landing(args, target):
     """The deny reason when the push lands, on the default branch, a commit that
     contains the tip of a loop branch (or a change branch) not yet on
@@ -505,10 +546,13 @@ def bash_hits(command, cwd, landings=None):
                         if t in ("&&", "||", ";", "|"):
                             break
                         args.append(t)
-                    hit = unarchived_landing(args, target)
+                    unreadable = unreadable_landing(args, target)
+                    if unreadable:
+                        landings.append(unreadable)
+                    hit = None if unreadable else unarchived_landing(args, target)
                     if hit:
                         landings.append(landing_reason(*hit))
-                    loop = loop_landing(args, target)
+                    loop = None if unreadable else loop_landing(args, target)
                     if loop:
                         landings.append(loop)
                 i = j + 1
diff --git a/scripts/test-protect-main-checkout.sh b/scripts/test-protect-main-checkout.sh
index 77a3d0d2..777344c4 100755
--- a/scripts/test-protect-main-checkout.sh
+++ b/scripts/test-protect-main-checkout.sh
@@ -92,6 +92,13 @@
 #       the checker refuses (self-review rows left open)              -> deny
 #   89. two such pushes in one command, each check inside a per-push budget
 #       but together over the command's one budget                    -> deny
+#   90-93. a landing push the hook cannot read, every existing check
+#       admitting it: `git -C $W push origin $B:main`, `"${B}:main"`,
+#       `$(git branch --show-current):main`, a backtick refspec       -> deny each
+#   94. `cd $W && git push origin HEAD` (repository and destination unknown) -> deny
+#   95. `git -C $W push -u origin $B` (destination unknown)          -> deny
+#   96. `git -C $W push origin spectre/x` (a literal non-default destination) -> allow
+#   97. `B=change; git -C <wt> push origin $B` (assigned in the command) -> allow
 set -euo pipefail
 SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
 HOOK="$SCRIPT_DIR/../hooks/protect-main-checkout.py"
@@ -445,6 +452,24 @@ printf '#!/usr/bin/env bash\nsleep 1.5\n' >"$ROOT/slowcheck"
 chmod +x "$ROOT/slowcheck"
 expect "89 two landing pushes share one checker budget" deny \
   "$(LOOP_LANDING_CHECK="$ROOT/slowcheck" LOOP_LANDING_TIMEOUT=2 run "$ROOT" Bash "{\"command\":$(q "git -C $LWT push origin spectre/chg:main; git -C $LWT push origin spectre/chg:main")}")"
+# 90-97: a push named through an unexpanded shell expansion cannot be checked;
+# the stub admits, so only the unreadable-landing check can deny.
+echo 0 >"$LOOP_RC"
+n=90
+for c in "git -C \$W push origin \$B:main" "git -C $WT push origin \"\${B}:main\"" \
+  "git -C $WT push origin \$(git branch --show-current):main" \
+  "git -C $WT push origin \`git branch --show-current\`:main"; do
+  expect "$n unreadable landing: $c" deny "$(run "$ROOT" Bash "{\"command\":$(q "$c")}")"
+  n=$((n + 1))
+done
+expect "94 cd to an unexpanded path, then push HEAD" deny \
+  "$(run "$ROOT" Bash "{\"command\":$(q 'cd $W && git push origin HEAD')}")"
+expect "95 push of an unexpanded branch" deny \
+  "$(run "$ROOT" Bash "{\"command\":$(q 'git -C $W push -u origin $B')}")"
+expect "96 unexpanded repository, literal non-default destination" allow \
+  "$(run "$ROOT" Bash "{\"command\":$(q 'git -C $W push origin spectre/x')}")"
+expect "97 a branch assigned in the command itself" allow \
+  "$(run "$ROOT" Bash "{\"command\":$(q "B=change; git -C $WT push origin \$B")}")"
 unset LOOP_LANDING_CHECK
 
 # 3 last: moving the main checkout off main lifts the protection
````

  - [x] **Step 4: Verify.** From the repository root: `bash scripts/test-protect-main-checkout.sh`
    (ends `all protect-main-checkout cases pass`), `python3 -m py_compile hooks/protect-main-checkout.py`,
    `scripts/check-python-suppressions.sh` and `scripts/check-references.sh` — all exit 0.
  - [x] **Step 5: Update the saved memories** (outside the repository, never committed), per
    decision `callers-and-memories`: in
    `/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/memory/protect-main-checkout-hook-quirks.md`
    extend the `$VAR` bullet with: a `git push` whose repository, remote or refspec is an
    unexpanded `$VAR`/`$(…)`/backtick and whose destination is the default branch or unknown is
    denied — write pushes with literal names; in
    `/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/memory/archive-push-base-checked-out-in-main.md`
    add to **How to apply** that the push after `update-ref` names the landing worktree's `-C`
    path and the branch literally, since the hook denies a default-branch push through a variable.
  - [x] **Step 6: Commit** both files with the subject below.

**Build:** green
**Files:** `hooks/protect-main-checkout.py`, `scripts/test-protect-main-checkout.sh`
**Tests:** `unreadable landing`, `cd to an unexpanded path, then push HEAD`, `push of an unexpanded branch`, `unexpanded repository, literal non-default destination`, `a branch assigned in the command itself`
**Regression:** reverting the commit lets `git -C $W push origin $B:main`, `"${B}:main"`, `$(git branch --show-current):main`, a backtick refspec, `cd $W && git push origin HEAD` and `git -C $W push -u origin $B` through unchecked — each `unreadable landing` case and the two other deny cases print `wanted deny, got allow`.
**Baseline:** before=82 after=87
<!-- measured: grep -cE '^[[:space:]]*expect ' scripts/test-protect-main-checkout.sh (before at 881de558; after with this task's diff applied) -->
**Commit:** fix(hooks): deny a default-branch push named through an unexpanded shell expansion
**After:** none

**Decision:** check-in-push-scan
**Decision:** deny-unreadable-push
**Decision:** unknown-repository-counts
**Decision:** literal-non-default-passes
**Decision:** callers-and-memories
