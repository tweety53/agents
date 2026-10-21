# self-review-kan-918-flow-improvement-port-the-git-snapshot-and

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-918's six self-review findings, one task each — F1 (Task 1), F2 (Task 2), F3
(Task 3), F4 (Task 4), F5 (Task 5), F6 (Task 6), as `proposal.md` numbers them.

**Architecture:** three prose fixes to `skills/flow/`, one new finding in `check-plan-shape`, one
new flag on `flow record finding`, and the integrate-change harness run case by case in parallel.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-21-self-review-kan-918-flow-improvement-port-the-git-snapshot-and-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/kan-918-flow-improvement-port-the-git-snapshot-and`, branch `spectre/kan-918-flow-improvement-port-the-git-snapshot-and`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- Each task's diff is the unified diff its apply step gives, applied with `git apply` from the worktree's root, in task order; all six were applied in that order and checked in this worktree at `eff5ca1c`. Save a diff by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork.
- `spectre/changes/self-review-kan-918-flow-improvement-port-the-git-snapshot-and/verbatim-moves.txt` already lists every sentence Tasks 1, 2, 4 and 5 reword or add, so `scripts/check-verbatim-moves.sh` exits 0 after each of them.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

## Review Focus

- Task 3: F15 never fires on a plan outside git or with no resolvable base; a file the task declares is never read; a prefix hit needs the file's subject to be the name; another directory is never read.
- Task 5: a line's keys go through the single-finding validation unchanged; a `from` key is refused; the first failing line stops the run and names its line.
- Task 6: a case that fails, or halts, still fails the run; a case log with no tally counts as a failure; every case gets its own `$WORK` and EXIT trap.

- [ ] 1. Scope PLAN FIELDS to a fix folded into its task
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -c 'when it folds into' skills/flow/review-panel.md`
    prints `0`: the brief binds every fix route, the on-top `Flow-Key:` commit included.
    <!-- measured: grep -c 'when it folds into' skills/flow/review-panel.md @ eff5ca1c → 0 -->
  - [ ] **Step 2: Apply the diff** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at eff5ca1c; check-verbatim-moves.sh, check-markdown-integrity.py, check-references.sh and check-dispatch-paragraphs.sh exit 0
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index 9bbc64e3..41501f60 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -844,13 +844,16 @@ reduction read, or the small append, has its own carve-out (`skills/flow/review-
 
 **Every fix subagent's dispatch prompt also carries the PLAN FIELDS paragraph**:
 
-> **PLAN FIELDS:** your fix owns its plan record, as part of the fix itself: when it adds a test
+> **PLAN FIELDS:** your fix owns its plan record, as part of the fix itself, when it folds into a
+> task's commit with `fold-fixup.sh`: when it adds a test
 > case or changes what tests a task names, update that task's `**Tests:**` field; when it changes
 > what a task's `**Baseline:**` counts or `**Files:**` paths declare — a test case added, a file
 > created — update those fields too. All of it lands in the worktree's
 > `<changeRoot>/tasks.md` in this same pass — never left for a reviewer to
 > catch next round. Edit them; do not stage or commit them — the plan record is
-> a planning path and is committed later by the pipeline, never in a fixup.
+> a planning path and is committed later by the pipeline, never in a fixup. A fix that lands as
+> its own commit on top, carrying the round's `Flow-Key:` trailer, is outside every task: it
+> edits no task's fields.
 
 **Every fix subagent's dispatch prompt also carries the ROUND SCOPE paragraph**:
 
diff --git a/skills/flow/SKILL-rationale.md b/skills/flow/SKILL-rationale.md
index 754eeaee..67eb65eb 100644
--- a/skills/flow/SKILL-rationale.md
+++ b/skills/flow/SKILL-rationale.md
@@ -506,6 +506,10 @@ parent's inline fix drops all four, so a Minor now costs about what writing its
 > never left for a reviewer to
 > catch next round (gymie kan-454, gymie KAN-459).
 
+> Scoped to the `fold-fixup.sh` route (kan-918): an on-top commit carrying `Flow-Key:` is outside
+> every task, and `check-task-records`' rule 2 holds a task's `**Files:**` to that task's own
+> commit, so task 3's fields widened for an on-top fix were refused at verify and reverted.
+
 ### review-panel.md — the fix step
 
 > A round may not chunk freely: its chunk count is bounded by
diff --git a/skills/flow/review-panel-fix-round.md b/skills/flow/review-panel-fix-round.md
index d163b443..61b35457 100644
--- a/skills/flow/review-panel-fix-round.md
+++ b/skills/flow/review-panel-fix-round.md
@@ -373,7 +373,7 @@ beside the mutation lines: the fix's report names the probe assertion it landed
 actual rendered pixels or geometry, and a fix commit for this class of bug that landed no such
 probe is rejected at the fix step — the round does not close and the finding stays open, for the
 handback below — rather than the regression being discovered next round. **The same
-walk holds the fix subagent to its PLAN FIELDS obligation:** a hunk that adds a
+walk holds a folded fix to its PLAN FIELDS obligation:** a `fold-fixup.sh` hunk that adds a
 test case, adds a file, or changes what a task's `**Baseline:**` counts, whose task's `**Tests:**`,
 `**Baseline:**` or `**Files:**` field in `<changeRoot>/tasks.md` does not reflect it, does not close
 the round; it goes to the handback.
````

  - [ ] **Step 3: Verify** — `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.py`,
    `scripts/check-references.sh` and `scripts/check-dispatch-paragraphs.sh` each exit 0.

**Build:** green
**Files:** `skills/flow/review-panel.md`, `skills/flow/review-panel-fix-round.md`, `skills/flow/SKILL-rationale.md`
**Tests:** none — prose; the lint guards are the check
**Regression:** none — prose only
**Commit:** docs(flow): scope PLAN FIELDS to a fix folded into its task
**After:** none

**Decision:** plan-fields-fold-only

- [ ] 2. Start the live check once the lint and test groups exit
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -c 'After the lint and test list' skills/flow/verify-and-handoff.md`
    prints `1`, while the same file starts that list as two detached, polled groups.
    <!-- measured: grep -c 'After the lint and test list' skills/flow/verify-and-handoff.md @ eff5ca1c → 1 -->
  - [ ] **Step 2: Apply the diff** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at eff5ca1c after Task 1's diff; check-verbatim-moves.sh and check-markdown-integrity.py exit 0
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index 363e8c93..6bfa550a 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -142,8 +142,10 @@ flow record dispatch end -change <name> -key verify[-<worktree basename>] \
 
 **Runs only when `design.md` carries a `## Live check` section naming something to exercise**
 (**D. Basic Workflow #3 — Writing plans**, `skills/flow/brainstorm-planner.md`); a one-line
-no-runtime section, or none, skips it. After the lint and test list, with the parent's own Bash
-calls: when the stack's worktree-resolved URLs do not answer, or `check-dev-stack-fresh.sh
+no-runtime section, or none, skips it. Once both groups of the lint and test list have exited —
+never while either still runs: kan-918's live check, run beside the polled groups, wrote
+`live-verification.md` into the tree `run-guard-tests.sh` watches, failing its watched-tree check
+and costing a 272-second re-run — with the parent's own Bash calls: when the stack's worktree-resolved URLs do not answer, or `check-dev-stack-fresh.sh
 <worktree>` exits non-zero, start it as **One stack for the verification tasks**
 (`skills/flow/implement.md`) does; then exercise what the section names and write each
 before/after figure, beside the failure look the section states, to `live-verification.md` in
````

  - [ ] **Step 3: Verify** — `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.py`
    and `scripts/check-references.sh` each exit 0.

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** none — prose; the lint guards are the check
**Regression:** none — prose only
**Commit:** docs(flow): start the live check once the lint and test groups exit
**After:** none

**Decision:** live-check-after-groups

- [ ] 3. Flag a Tests name the base already declares outside the task's files
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Save Step 2's diff, apply its
    harness hunk alone (`git apply --include='scripts/test-check-plan-shape.sh' <the saved diff>`)
    and run `scripts/test-check-plan-shape.sh`: case 39's first and last assertions fail with `rc=0`,
    and its mutation assertion fails with `the mutation did not apply`.
    <!-- measured: the harness hunk alone, scripts/test-check-plan-shape.sh @ eff5ca1c → 3 failure(s), all case 39 -->
  - [ ] **Step 2: Apply the fix** — the rest (`git apply --exclude='scripts/test-check-plan-shape.sh' <the saved diff>`):

````diff verified:applied with git apply to this worktree at eff5ca1c; scripts/test-check-plan-shape.sh passes all cases and fails case 39 with check-plan-shape.py reverted; scripts/check-plan-shape.sh with no argument exits 0; kan-918's task 3 copied into spectre/changes/ exits 1 naming TestFoldFixup
diff --git a/scripts/check-plan-shape.py b/scripts/check-plan-shape.py
index 022f48c4..09aa8a6d 100755
--- a/scripts/check-plan-shape.py
+++ b/scripts/check-plan-shape.py
@@ -1,11 +1,11 @@
 #!/usr/bin/env python3
 """check-plan-shape.py — check that a `tasks.md` is SHAPED so the guard
 that will later judge its commits (`check-task-commit-fields.py`) can
-actually read it. Fourteen findings, F1-F14 (F1-F6's canonical definitions,
+actually read it. Fifteen findings, F1-F15 (F1-F6's canonical definitions,
 decisions and rationale: `spectre/changes/kan-121-run-the-guards-own-
 parsers-over-tasks-md-at-plan/design.md` — do not restate them here, a
 second copy is a Single Source of Truth violation, the same class of drift
-stats/internal/guard/planprovenance.go's ported docstring warns against; F7-F14's canonical
+stats/internal/guard/planprovenance.go's ported docstring warns against; F7-F15's canonical
 definitions are the rows in the Findings list below).
 
 This guard IMPORTS the real parsers rather than reimplementing their
@@ -34,7 +34,7 @@ Exit codes:
   0  clean — every task in the file is shaped so the real parsers read it
      without ambiguity (including a file with zero findings; a file with
      zero TASKS is F5, not clean — see below).
-  1  one or more of F1-F14 found. Printed one per line as
+  1  one or more of F1-F15 found. Printed one per line as
      `file:line: message`, naming the task id (F5 excepted: a file with no
      tasks has no task id to name).
   2  invocation error — wrong argument count, or the file cannot be read.
@@ -166,6 +166,25 @@ field-shaped line inside a worked example is not a declaration.
       paths, so the task lands as a planning commit and is ticked by hand
       (kan-941's Task 4, `visual-sweep.md`). Reported at the task line.
 
+  F15 A backticked `**Tests:**` name that a test-shaped file outside the
+      task's `**Files:**` (and its `**Allowed-collateral:**` globs), in a
+      directory one of those files lies in, already declares at the
+      change's base: a `func`/`fun`/`def` of exactly that name, or one
+      starting with it in the file whose subject the name is
+      (`TestFoldFixup` and `fold_fixup_test.go`). The task's commit cannot
+      touch that file, so check-task-commit-fields' diff search never finds
+      the name and refuses the commit (kan-918's Task 3 declared
+      `TestFoldFixup`, the prefix of five existing tests there). A new
+      `TestFoo` beside an existing `TestFooBar` passes: a file the task
+      declares is never read, and another file's `TestFooBar` is a prefix
+      hit only in `foo_test.go`. Another directory is never read — Go
+      allows one name per package, and `TestMain` is in many. The base is
+      the merge base of HEAD with the branch's recorded base
+      (`branch.<current>.flowBase`, its origin/ copy preferred) or else with
+      `refs/remotes/origin/HEAD`; a plan outside a git repository, or one
+      whose base does not resolve, is not checked for F15. Reported at the
+      task's `**Tests:**` line.
+
 A duplicate task id (two task lines sharing one id) is check-task-build-
 green.py's own violation, not this guard's: `collect_task_ids` de-
 duplicates by first occurrence, and `select_task`/`parse_task_fields`
@@ -182,6 +201,7 @@ import fnmatch
 import importlib.util
 import os
 import re
+import subprocess
 import sys
 from pathlib import Path
 from typing import Dict, List, NamedTuple, Optional, Set
@@ -554,6 +574,108 @@ def _check_task(
     return violations
 
 
+# DECL_NAME_RE — F15's test declaration: Go's `func`, Kotlin's `fun`,
+# Python's `def`, then the declared name. IDENT_RE — a `**Tests:**` name
+# shaped like one; any other name is never a declaration and is skipped.
+DECL_NAME_RE = re.compile(r"\b(?:func|fun|def)\s+([A-Za-z_]\w*)")
+IDENT_RE = re.compile(r"[A-Za-z_]\w*\Z")
+
+
+def _stem(path: str) -> str:
+    """A test-shaped file's subject: `fold_fixup_test.go` -> `fold_fixup`."""
+    name = os.path.basename(path)
+    for suffix in ("_test.go", ".spec.ts"):
+        name = name[: -len(suffix)] if name.endswith(suffix) else name
+    name = name.split(".test.")[0]
+    return name[len("test-"):] if name.startswith("test-") else name
+
+
+def _key(name: str) -> str:
+    """A name folded for F15's stem comparison: no leading `Test`/`test_`,
+    lower case, letters and digits only."""
+    return re.sub(r"[^a-z0-9]", "", re.sub(r"^test_?", "", name, flags=re.I).lower())
+
+
+def _git(root: str, *args: str) -> Optional[str]:
+    proc = subprocess.run(
+        ["git", "-C", root, "-c", "core.quotePath=false", *args],
+        capture_output=True,
+        text=True,
+    )
+    return proc.stdout if proc.returncode == 0 else None
+
+
+def _base_declarations(path: str) -> List[tuple]:
+    """F15's (file, name) test declarations in the test-shaped files of the
+    change's base tree; [] when the plan sits outside a git repository or
+    no base resolves."""
+    root = _git(os.path.dirname(os.path.abspath(path)), "rev-parse", "--show-toplevel")
+    if root is None:
+        return []
+    root = root.strip()
+    branch = (_git(root, "branch", "--show-current") or "").strip()
+    recorded = (_git(root, "config", "--get", f"branch.{branch}.flowBase") or "").strip() if branch else ""
+    candidates = [f"refs/remotes/origin/{recorded}", recorded] if recorded else ["refs/remotes/origin/HEAD"]
+    rev = None
+    for ref in candidates:
+        rev = _git(root, "merge-base", "HEAD", ref)
+        if rev is not None:
+            break
+    if rev is None:
+        return []
+    rev = rev.strip()
+    out = _git(
+        root, "grep", "-I", "-z", "-E", "(func|fun|def)[[:space:]]", rev, "--",
+        *(f":(glob)**/{pattern}" for pattern in TEST_SHAPED_PATTERNS),
+    ) or ""
+    decls = []
+    for record in out.splitlines():
+        where, _, text = record.partition("\0")
+        for name in DECL_NAME_RE.findall(text):
+            decls.append((where[len(rev) + 1:], name))
+    return decls
+
+
+def _check_f15(path: str, lines: List[str], ids: List[str]) -> List[str]:
+    """F15 — a `**Tests:**` name a base-tree test file outside the task's
+    own files already declares."""
+    decls = _base_declarations(path)
+    if not decls:
+        return []
+    violations: List[str] = []
+    for task_id in ids:
+        fields = parse_task_fields(lines, task_id)
+        if fields.unclosed_fence_line is not None:
+            continue
+        dirs = {os.path.dirname(f) for f in fields.files}
+        for spec in fields.tests:
+            if not IDENT_RE.match(spec.label):
+                continue
+            hits: Dict[str, List[str]] = {}
+            for file, name in decls:
+                same = name == spec.label or (
+                    name.startswith(spec.label) and _key(_stem(file)) == _key(spec.label)
+                )
+                owned = file in fields.files or any(
+                    fnmatch.fnmatch(file, g) for g in fields.allowed_collateral
+                )
+                if same and not owned and os.path.dirname(file) in dirs:  # F15
+                    hits.setdefault(file, []).append(f"`{name}`")
+            if not hits:
+                continue
+            found = select_task(lines, task_id)
+            field = _first_field_line(found.lines, found.body_start, "Tests")
+            where = "; ".join(f"{f} ({', '.join(n)})" for f, n in sorted(hits.items()))
+            violations.append(
+                f"{path}:{field.line}: task {task_id}'s **Tests:** names "
+                f"`{spec.label}`, which the base already declares outside its "
+                f"**Files:** — {where}; the task's commit cannot "
+                "add it, so check-task-commit-fields refuses the commit — "
+                "name the test this task adds, or list the file in **Files:**"
+            )
+    return violations
+
+
 def check_file(path: str) -> List[str]:
     """Read `path` and return its F1-F6 violations, formatted with `path`
     itself as the file portion of each `file:line: message` line."""
@@ -582,6 +704,7 @@ def check_file(path: str) -> List[str]:
     violations.extend(_check_after_references(path, lines))
     violations.extend(_check_file_ownership(path, lines))
     violations.extend(_check_baseline_convention(path, lines))
+    violations.extend(_check_f15(path, lines, ids))
     return violations
 
 
diff --git a/scripts/test-check-plan-shape.sh b/scripts/test-check-plan-shape.sh
index d7a53c44..4e263f63 100755
--- a/scripts/test-check-plan-shape.sh
+++ b/scripts/test-check-plan-shape.sh
@@ -8,7 +8,7 @@
 # REAL fixture files on disk — never a copy of its logic, never its Python
 # internals imported and asserted on directly.
 #
-# One case per finding (F1-F14), a clean-plan case, a no-arg aggregation
+# One case per finding (F1-F15), a clean-plan case, a no-arg aggregation
 # case, and three exit-2 cases (missing sibling module, unreadable file,
 # usage error). Additionally, for each of F1-F6 and F11, a MUTATION case: a
 # throwaway copy of check-plan-shape.py with that finding's own check
@@ -1076,6 +1076,52 @@ sed -i.bak 's#`docs/superpowers/notes.md`#`docs/visual-sweep.md`#' "$TASKS_MD"
 run_guard "$TASKS_MD"
 [ "$RC" -eq 0 ] && pass "case 38 (F14): one implementation path passes" || fail "case 38 (F14) with an implementation path: rc=$RC out=$OUT"
 
+# ===========================================================================
+# Case 39 (F15): a **Tests:** name a base test file outside the task's
+# **Files:** already declares cannot be added by the task's commit (kan-918's
+# Task 3, `TestFoldFixup`); a new `TestFoo` beside an existing `TestFooBar`
+# passes, in the task's own file and in another file of the package alike.
+# ===========================================================================
+new_fixture
+git init -q -b main "$FIXTURE"
+mkdir -p "$FIXTURE/pkg"
+printf 'package pkg\n\nfunc TestFoldFixupFolds(t *testing.T) {}\n' > "$FIXTURE/pkg/fold_fixup_test.go"
+printf 'package pkg\n\nfunc TestFooBar(t *testing.T) {}\n' > "$FIXTURE/pkg/foo_bar_test.go"
+git -C "$FIXTURE" add pkg
+git -C "$FIXTURE" -c user.name=t -c user.email=t@example.invalid -c core.hooksPath=/dev/null commit -qm base
+git -C "$FIXTURE" checkout -q -b work
+git -C "$FIXTURE" config branch.work.flowBase main
+{
+  printf -- '- [ ] 1. Port guard-autosquash\n\n'
+  printf '**Files:** `pkg/autosquash.go`, `pkg/autosquash_test.go`\n'
+  printf '**Tests:** `TestAutosquash`, `TestFoldFixup`\n'
+  printf '**Commit:** `feat(pkg): port guard-autosquash`\n'
+} > "$TASKS_MD"
+run_guard "$TASKS_MD"
+case "$RC:$OUT" in
+  1:*'names `TestFoldFixup`, which the base already declares outside its **Files:** — pkg/fold_fixup_test.go (`TestFoldFixupFolds`)'*) pass "case 39 (F15): an existing test family outside Files fails" ;;
+  *) fail "case 39 (F15): rc=$RC out=$OUT" ;;
+esac
+assert_mutation_removes_finding "F15" "$TASKS_MD" "which the base already declares" "case 39 (F15) mutation"
+sed -i.bak 's#`pkg/autosquash_test.go`#`pkg/autosquash_test.go`, `pkg/fold_fixup_test.go`#' "$TASKS_MD"
+run_guard "$TASKS_MD"
+[ "$RC" -eq 0 ] && pass "case 39 (F15): the family's file in Files passes" || fail "case 39 (F15) file in Files: rc=$RC out=$OUT"
+sed -i.bak 's#`TestFoldFixup`#`TestFoo`#' "$TASKS_MD"
+run_guard "$TASKS_MD"
+[ "$RC" -eq 0 ] && pass "case 39 (F15): a new TestFoo beside TestFooBar in another file passes" || fail "case 39 (F15) TestFoo: rc=$RC out=$OUT"
+sed -i.bak 's#`pkg/fold_fixup_test.go`#`pkg/foo_bar_test.go`#' "$TASKS_MD"
+run_guard "$TASKS_MD"
+[ "$RC" -eq 0 ] && pass "case 39 (F15): a new TestFoo beside TestFooBar in its own file passes" || fail "case 39 (F15) TestFoo own file: rc=$RC out=$OUT"
+printf 'package pkg\n\nfunc TestFoo(t *testing.T) {}\n' > "$FIXTURE/pkg/old_test.go"
+git -C "$FIXTURE" add pkg
+git -C "$FIXTURE" -c user.name=t -c user.email=t@example.invalid -c core.hooksPath=/dev/null commit -qm more
+git -C "$FIXTURE" branch -f main HEAD
+run_guard "$TASKS_MD"
+case "$RC:$OUT" in
+  1:*'names `TestFoo`, which the base already declares outside its **Files:** — pkg/old_test.go (`TestFoo`)'*) pass "case 39 (F15): an exact name outside Files fails" ;;
+  *) fail "case 39 (F15) exact name: rc=$RC out=$OUT" ;;
+esac
+
 # ===========================================================================
 # Unlanded archives: the scan cannot list the archived changes the base does
 # not carry -> exit 2, never a scan that silently skips them. A copy of the
````

  - [ ] **Step 3: Verify** — `scripts/test-check-plan-shape.sh` ends `all cases passed`;
    `scripts/check-plan-shape.sh` (no argument) and `scripts/check-python-suppressions.sh` exit 0.

**Build:** green
**Files:** `scripts/check-plan-shape.py`, `scripts/test-check-plan-shape.sh`
**Tests:** Case 39
**Regression:** Case 39 — reverting the commit lets a `**Tests:**` name a base test file outside the task's files already declares pass `check-plan-shape`, and the case's first and last assertions fail with `rc=0`.
**Baseline:** before=38 after=39
<!-- measured: grep -cE '^# Case [0-9]+' scripts/test-check-plan-shape.sh @ eff5ca1c -->
**Commit:** feat(scripts): flag a Tests name the base already declares outside the task's files
**After:** none

**Decision:** f15-in-plan-shape
**Decision:** f15-match-rule

- [ ] 4. Carry each slot's REPORT FILE path in its Agent call
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** The traced slot never read its
    brief: `python3 -c "import json,sys; print(sorted(b.get('name') for l in open(sys.argv[1]) for d in [json.loads(l)] if d.get('type')=='assistant' for b in (d['message'].get('content') or []) if isinstance(b,dict) and b.get('type')=='tool_use'))" ~/.claude/projects/-Users-tweety53-Projects-agents/5c844d2a-d709-4945-b838-b398826a1566/subagents/agent-a49d6b737c5810fb7.jsonl`
    prints `['Bash', 'SubagentHandback']` — no `Read`. `design.md` **Context** carries the full trace.
    <!-- measured: the python3 one-liner on the slot transcript; a machine-local file, read 2026-10-21 → ['Bash', 'SubagentHandback'] -->
  - [ ] **Step 2: Apply the diff** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at eff5ca1c after Tasks 1-3's diffs; check-verbatim-moves.sh and check-dispatch-paragraphs.sh exit 0
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index 41501f60..f43cc4d5 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -360,7 +360,7 @@ beside the rendered path where a dispatch bundles it — the render still carrie
 role's pass and the INDEPENDENT PASSES paragraph, and a dispatch carrying `mutation` alone renders
 the shared paragraphs and no PASS section, the brief typed beside that rendered path. The Agent call's prompt carries only the baseline
 pointer, MODEL HANDSHAKE, CONTEXT BUNDLE, the relocation-comparison pointer where one exists, the
-reproducer rule, GATE FOCUS where it applies, and `read <rendered path> in full first — it is your brief`. The return message carries one findings summary per role under a heading naming the
+reproducer rule, GATE FOCUS where it applies, `read <rendered path> in full first — it is your brief`, and each role's REPORT FILE path — so a slot that answers without reading its brief still owes its file (kan-918: round 1's `primary+principles` re-run read only the diff and handed back in 14 seconds, writing neither report). The return message carries one findings summary per role under a heading naming the
 role; the parent records each finding under that role.
 
 Every bundle prompt also carries this paragraph verbatim:
````

  - [ ] **Step 3: Verify** — `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.py`
    and `scripts/check-dispatch-paragraphs.sh` each exit 0.

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** none — prose; the lint guards are the check
**Regression:** none — prose only
**Commit:** docs(flow): carry each slot's REPORT FILE path in its Agent call
**After:** Task 1

**Decision:** report-path-in-call

- [ ] 5. Record a round's findings from a JSON-lines file
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Save Step 2's diff, apply its
    test hunk alone (`git apply --include='stats/cmd/flow/record_test.go' <the saved diff>`) and run
    `cd stats && go test ./cmd/flow/ -run '^TestRecordFindingFromFile$' -count=1`: it fails — `-from`
    is an unknown flag.
    <!-- measured: the test hunk alone, go test -run '^TestRecordFindingFromFile$' @ eff5ca1c → FAIL -->
  - [ ] **Step 2: Apply the fix** — the rest (`git apply --exclude='stats/cmd/flow/record_test.go' <the saved diff>`):

````diff verified:applied with git apply to this worktree at eff5ca1c after Tasks 1-4's diffs; go test ./cmd/flow/ -count=1 passes, TestRecordFindingFromFile fails with record.go reverted, go vet and gofmt -l clean, check-verbatim-moves.sh exits 0
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index f43cc4d5..0f11cd73 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -733,3 +733,7 @@ discipline**'s never-`cat`-a-report rule (`skills/flow/implement.md`).
 records every finding itself. Every
-finding a round raised is recorded in one Bash call, one `flow record finding` per finding:
+finding a round raised is recorded in one Bash call, `flow record finding -change <name> -from
+<abs-worktree>/.superpowers/sdd/findings-<round>.jsonl`, over a file the parent writes first: one
+JSON object per finding, its keys the flags below without their dash, a flag the finding omits
+left out — never a shell loop of single `flow record finding` calls, whose quoting of an optional
+flag cost kan-918's run a retry.
 
diff --git a/stats/cmd/flow/record.go b/stats/cmd/flow/record.go
index 6888f74e..4ba91f24 100644
--- a/stats/cmd/flow/record.go
+++ b/stats/cmd/flow/record.go
@@ -290,6 +290,8 @@ const recordUsage = `usage: flow record dispatch begin [-addr url] [-timeout dur
                              [-dispatch-seq n] -note text
                              [-supersedes F<n>] [-regression-of F<n>]
                              [-pattern name] [-task id]
+       flow record finding  [-addr url] [-timeout dur] [-C dir]
+                             -change name -from file
        flow record status   [-addr url] [-timeout dur] [-C dir]
                              -change name -ref F<n> -status status
                              [-category cat]
@@ -1450,7 +1452,72 @@ func beginQueued(projectKey, change, key, sessionToken string) bool {
 // "recorded:" when the write inserted, "updated:" when it replaced. A
 // panel run can then tell a finding it has just raised from one a fix
 // round restated, which is a fact no caller can work out for itself.
+// splitFromFlag takes `-from <file>` (or `-from=<file>`) out of a `flow
+// record finding` command line, returning the file and the other arguments.
+func splitFromFlag(args []string) (path string, rest []string, ok bool) {
+	for i := 0; i < len(args); i++ {
+		a := strings.TrimPrefix(args[i], "-")
+		switch {
+		case (a == "from" || a == "-from") && i+1 < len(args):
+			path, ok = args[i+1], true
+			i++
+		case strings.HasPrefix(a, "from=") || strings.HasPrefix(a, "-from="):
+			path, ok = a[strings.Index(a, "=")+1:], true
+		default:
+			rest = append(rest, args[i])
+		}
+	}
+	return path, rest, ok
+}
+
+// runRecordFindingFrom implements `flow record finding -from <file>`: one
+// finding per non-blank line of <file>, each a JSON object whose keys are
+// the single-finding flags without their dash ({"ref":"F3","round":1,
+// "slot":"primary",...}), recorded in file order through runRecordFinding
+// with the command line's other flags (-change, -C, ...) ahead of the
+// line's own. JSON rather than a delimited row, so a note carrying a tab, a
+// quote or a `$` needs no shell quoting (kan-918's self-review: a
+// hand-built zsh loop over 21 findings word-split an optional -task). The
+// first line that fails stops the run, naming its line number; the lines
+// before it stay recorded and each line is a rewrite by ref, so a re-run
+// after the fix is safe.
+func runRecordFindingFrom(ctx context.Context, path string, rest []string, stdout, stderr io.Writer) int {
+	data, err := os.ReadFile(path)
+	if err != nil {
+		fmt.Fprintf(stderr, "flow: -from: %v\n", err)
+		return 2
+	}
+	for i, line := range strings.Split(string(data), "\n") {
+		if strings.TrimSpace(line) == "" {
+			continue
+		}
+		var row map[string]any
+		dec := json.NewDecoder(strings.NewReader(line))
+		dec.UseNumber()
+		if err := dec.Decode(&row); err != nil {
+			fmt.Fprintf(stderr, "flow: -from %s:%d: %v\n", path, i+1, err)
+			return 2
+		}
+		args := append([]string(nil), rest...)
+		for _, k := range slices.Sorted(maps.Keys(row)) {
+			if k == "from" {
+				fmt.Fprintf(stderr, "flow: -from %s:%d: a line cannot carry \"from\"\n", path, i+1)
+				return 2
+			}
+			args = append(args, "-"+k+"="+fmt.Sprint(row[k]))
+		}
+		if code := runRecordFinding(ctx, args, stdout, stderr); code != 0 {
+			fmt.Fprintf(stderr, "flow: -from %s:%d: not recorded; the lines after it were not tried\n", path, i+1)
+			return code
+		}
+	}
+	return 0
+}
+
 func runRecordFinding(ctx context.Context, args []string, stdout, stderr io.Writer) int {
+	if path, rest, ok := splitFromFlag(args); ok {
+		return runRecordFindingFrom(ctx, path, rest, stdout, stderr)
+	}
 	fset := flag.NewFlagSet("flow record finding", flag.ContinueOnError)
 	fset.SetOutput(stderr)
 	var f recordIdentityFlags
diff --git a/stats/cmd/flow/record_test.go b/stats/cmd/flow/record_test.go
index 9c7d41b5..707adc8d 100644
--- a/stats/cmd/flow/record_test.go
+++ b/stats/cmd/flow/record_test.go
@@ -5592,3 +5592,65 @@ func TestRecordMutationRefusesAPathNoWorktreeHolds(t *testing.T) {
 		}
 	})
 }
+
+// TestRecordFindingFromFile pins `flow record finding -from <file>`: every
+// non-blank JSON line is one finding, recorded in file order with the
+// command line's own flags, a note carrying a tab, quotes and a `$` arrives
+// intact, and a malformed line stops the run at its line number.
+func TestRecordFindingFromFile(t *testing.T) {
+	repo := recordRepo(t, "kan-918")
+	isolatedStateRoot(t)
+
+	var posted []records.Finding
+	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
+		body, _ := io.ReadAll(r.Body)
+		var f records.Finding
+		if err := json.Unmarshal(body, &f); err != nil {
+			t.Errorf("decode posted finding: %v\n%s", err, body)
+		}
+		posted = append(posted, f)
+		w.WriteHeader(http.StatusCreated)
+		_, _ = w.Write(body)
+	}))
+	defer srv.Close()
+
+	file := filepath.Join(t.TempDir(), "findings.jsonl")
+	lines := `{"ref":"F1","round":1,"slot":"primary","severity":"Important","location":"a.go:3","status":"open","reproducer":"scripts/x.sh","note":"tab\there \"q\" ${7:+-task $7}","dispatch-seq":4,"task":"3"}
+
+{"ref":"F2","slot":"principles","severity":"Minor","status":"open","reproducer":"none — prose","note":"second"}
+`
+	if err := os.WriteFile(file, []byte(lines), 0o644); err != nil {
+		t.Fatal(err)
+	}
+	var stdout, stderr bytes.Buffer
+	code := run(context.Background(),
+		[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "kan-918", "-from", file},
+		strings.NewReader(""), &stdout, &stderr)
+	if code != 0 {
+		t.Fatalf("record finding -from exit code = %d, want 0; stderr:\n%s", code, stderr.String())
+	}
+	if len(posted) != 2 || posted[0].Ref != "F1" || posted[1].Ref != "F2" {
+		t.Fatalf("posted = %+v, want F1 then F2", posted)
+	}
+	if got := posted[0]; got.Round != 1 || got.TaskID != "3" || got.DispatchSeq == nil || *got.DispatchSeq != 4 ||
+		got.Note != "tab\there \"q\" ${7:+-task $7}" {
+		t.Errorf("F1 = %+v, want round 1, task 3, dispatch-seq 4 and the note byte for byte", got)
+	}
+	if got := posted[1]; got.Round != 0 || got.TaskID != "" {
+		t.Errorf("F2 = %+v, want round 0 and no task", got)
+	}
+
+	if err := os.WriteFile(file, []byte(`{"ref":"F3","slot":"primary","severity":"Minor","status":"open","reproducer":"none — x","note":"ok"}`+"\n{not json\n"), 0o644); err != nil {
+		t.Fatal(err)
+	}
+	stderr.Reset()
+	code = run(context.Background(),
+		[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "kan-918", "-from=" + file},
+		strings.NewReader(""), &stdout, &stderr)
+	if code != 2 || !strings.Contains(stderr.String(), file+":2:") {
+		t.Errorf("malformed line 2: exit %d, stderr %q; want 2 naming %s:2", code, stderr.String(), file)
+	}
+	if len(posted) != 3 || posted[2].Ref != "F3" {
+		t.Errorf("posted = %+v, want line 1 (F3) recorded before line 2 stopped the run", posted)
+	}
+}
````

  - [ ] **Step 3: Verify** — `cd stats && gofmt -l . && make -s web-dist && go vet ./cmd/flow/` prints nothing
    and exits 0, then `cd stats && go test ./cmd/flow/ -run '^TestRecordFinding' -count=1` passes;
    `scripts/check-verbatim-moves.sh` and `scripts/check-markdown-integrity.py` exit 0.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `skills/flow/review-panel.md`
**Tests:** `TestRecordFindingFromFile`
**Regression:** `TestRecordFindingFromFile` — reverting the commit makes `-from` an unknown flag and the test fails at its first `run`.
**Baseline:** before=120 after=121
<!-- measured: grep -cE '^func Test' stats/cmd/flow/record_test.go @ eff5ca1c -->
**Commit:** feat(stats): record a round's findings from a JSON-lines file
**After:** Task 1, 4

**Decision:** record-finding-from-jsonl

- [ ] 6. Run the integrate-change harness cases in parallel
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `time bash scripts/test-integrate-change.sh`
    ends `195 passed, 0 failed` after several minutes of wall.
    <!-- measured: time bash scripts/test-integrate-change.sh @ eff5ca1c → 195 passed, 6:30 and 5:39 wall on two runs -->
  - [ ] **Step 2: Apply the diff** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at eff5ca1c; bash -n passes; the harness passes every check in a fraction of its serial wall; a copy with every halt-guarded prepare forced to halt exits 1, its failures counted
diff --git a/scripts/test-integrate-change.sh b/scripts/test-integrate-change.sh
index b0c7e7da..1e7cfa25 100755
--- a/scripts/test-integrate-change.sh
+++ b/scripts/test-integrate-change.sh
@@ -249,6 +249,54 @@ to_land() {
   g -C "$WORK/a-worktrees/demo" commit -qm "docs(self-review): demo self-review report"
 }
 
+# peer_fix_case <prefix>: the report names a fix committed in b, b's base
+# moves, and land runs — checked under <prefix>.
+peer_fix_case() {
+  to_land "$1" no-report
+  printf 'fix\n' >"$WORK/b-worktrees/demo/fix-b.txt"
+  g -C "$WORK/b-worktrees/demo" add fix-b.txt && g -C "$WORK/b-worktrees/demo" commit -qm "fix(b): the peer fix"
+  old="$(git -C "$WORK/b-worktrees/demo" rev-parse --short HEAD)"
+  mkdir -p "$WORK/a-worktrees/demo/docs/self-review"
+  printf '# demo self-review\n\n- **[flow-fix]** the peer fix — fixed: %s\n' "$old" >"$WORK/a-worktrees/demo/docs/self-review/demo-self-review.md"
+  g -C "$WORK/a-worktrees/demo" add docs/self-review/demo-self-review.md
+  g -C "$WORK/a-worktrees/demo" commit -qm "docs(self-review): demo self-review report"
+  printf 'moved\n' >"$WORK/b/moved.txt"
+  g -C "$WORK/b" add moved.txt && g -C "$WORK/b" commit -qm "base: moved" && g -C "$WORK/b" push -q origin main
+  sut land "$WORK/a" demo
+  new="$(git --git-dir="$WORK/b-origin.git" log --format=%h --grep="^fix(b): the peer fix$" -1 main)"
+  check "${1}land exits 0" '[ "$RC" -eq 0 ]'
+  check "${1}the re-sync rewrote b's fix sha" '[ -n "$new" ] && [ "$new" != "$old" ]'
+  check "${1}the report on a's origin main names the rewritten sha" \
+    'git --git-dir="$WORK/a-origin.git" show main:docs/self-review/demo-self-review.md | grep -q "fixed: $new$"'
+  check "${1}a fixed row is recorded for b's fix" \
+    'grep -q -- "-disposition fixed -ref $new " "$WORK/fixed.log"'
+}
+
+# ui_a: a declares `ui paths` its task commit touches, and carries the
+# verifier's record a verifier row counts only with.
+ui_a() {
+  printf '\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `work-a.txt` |\n' >>"$WORK/a-worktrees/demo/.flow/project.md"
+  printf '# Visual verification — demo\n' >"$WORK/a-worktrees/demo/spectre/changes/demo/visual-verification.md"
+  g -C "$WORK/a-worktrees/demo" add spectre/changes/demo/visual-verification.md
+  g -C "$WORK/a-worktrees/demo" commit -qam "chore(spectre): plan-gate demo" --amend
+}
+# flag <worktree> <sha> [<change>]: an operator's false-positive flag on
+# <worktree>'s MISSING verdict row naming HEAD <sha>.
+flag() {
+  printf '[{"change":"%s","guard":"check-visual-verify-dispatched","worktree":"%s","verdict":"VISUAL-VERIFY-MISSING: no verifier — HEAD %s","falsePositive":true,"falsePositiveReason":"hand-verified","flaggedAt":"2026-01-01T00:00:00Z"}]' \
+    "${3:-demo}" "$(cd "$1" && pwd -P)" "$2" >"$WORK/verdicts.json"
+}
+
+# Every case below runs in its own background subshell — its own fixture
+# ($WORK), counters and log — and the run waits for them all, so the suite
+# takes as long as its slowest case rather than their sum: run serially,
+# this harness alone set the guard suite's wall (kan-918's self-review).
+LOGS="$(mktemp -d)"
+PIDS=()
+# end_case: last in a case's log, its own tally; its status, the case's.
+end_case() { printf '%d passed, %d failed\n' "$PASS" "$FAIL"; [ "$FAIL" -eq 0 ]; }
+
+( trap cleanup_work EXIT
 # --- narrative_format ---
 fixture
 sut prepare "$WORK/a" demo
@@ -274,7 +322,10 @@ narrative_passes "a prose mention of the token" 'The `In-run pipeline fix:` line
 narrative_passes "a well-formed line with trailing whitespace" '- In-run pipeline fix: 1a2b3c4 — the defect (blast radius 2 files)   '
 narrative_passes "a deferred line" '- In-run pipeline fix: deferred — the defect (blast radius 1 file)'
 narrative_passes "a backticked sha" '- In-run pipeline fix: `1a2b3c4` — the defect (blast radius 3 files).' 
+end_case ) >"$LOGS/01.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- narrative_runs ---
 fixture
 echo 1 >"$WORK/fix-sessions"
@@ -296,7 +347,10 @@ sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b
 check "narrative_runs: a failed stage-sessions read names the failure in the stop" \
   '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: narrative — check-narrative-runs: flow stage sessions"'
 rm -f "$WORK/fail-sessions"
+end_case ) >"$LOGS/02.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- happy_path ---
 fixture
 sut prepare "$WORK/a" demo
@@ -335,7 +389,10 @@ check "the stage log runs preflight through refresh-main-checkout" \
 check "run 2 keeps a's pre-reshape pin, locally and on origin" \
   'git -C "$WORK/a" rev-parse -q --verify refs/flow/pre-reshape/demo >/dev/null && git --git-dir="$WORK/a-origin.git" rev-parse -q --verify refs/flow/pre-reshape/demo >/dev/null'
 check "land ends on NEXT: Finished handoff" 'printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: print the Finished handoff"'
+end_case ) >"$LOGS/03.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- foreign_staged ---
 fixture
 printf 'stray\n' >"$WORK/b/stray.txt"
@@ -345,7 +402,10 @@ check "staged work in a main checkout stops prepare with exit 1" '[ "$RC" -eq 1
 check "the stop is STOP: foreign-staged" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: foreign-staged"'
 check "the guard's own STAGED-FOREIGN line is relayed" 'printf "%s" "$OUT" | grep -q "^STAGED-FOREIGN: $WORK/b"'
 check "no preflight ran" '! grep -q flow.preflight "$WORK/stages.log" 2>/dev/null'
+end_case ) >"$LOGS/04.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- no_report ---
 fixture
 to_land "no_report: " no-report
@@ -359,7 +419,10 @@ check "no_report: nothing was landed" \
   '[ "$(git --git-dir="$WORK/b-origin.git" log --format=%s main -1)" = base ] && [ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -1)" = base ]'
 check "no_report: a's remote branch still sits at commit's HEAD" \
   '[ "$(git --git-dir="$WORK/a-origin.git" rev-parse spectre/demo)" = "$pushed" ]'
+end_case ) >"$LOGS/05.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- loop_landing ---
 fixture
 to_land "loop_landing: "
@@ -369,7 +432,10 @@ check "loop_landing: an open own-branch self-review loop stops land at STOP: loo
   '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: loop-landing"'
 check "loop_landing: nothing landed on either main" \
   '[ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -1)" = base ] && [ "$(git --git-dir="$WORK/b-origin.git" log --format=%s main -1)" = base ]'
+end_case ) >"$LOGS/06.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- untasked ---
 fixture
 printf 'stray\n' >"$WORK/b-worktrees/demo/stray.txt"
@@ -379,7 +445,10 @@ check "untasked: prepare stops at STOP: unfinished-work" 'printf "%s" "$OUT" | t
 check "untasked: the stray commit is listed" 'printf "%s" "$OUT" | grep -q "^  [0-9a-f]\{12\} test(b): a stray edit$"'
 sut prepare "$WORK/a" demo --accept-outstanding
 check "untasked: --accept-outstanding passes the gate" '[ "$RC" -eq 0 ]'
+end_case ) >"$LOGS/07.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- archived_rerun ---
 fixture
 to_land "archived_rerun: "
@@ -388,7 +457,10 @@ g -C "$WORK/a-worktrees/demo" add more.txt && g -C "$WORK/a-worktrees/demo" comm
 sut prepare "$WORK/a" demo
 check "archived_rerun: prepare passes the gate" '[ "$RC" -eq 0 ]'
 check "archived_rerun: b's reshaped commit is not listed" '! printf "%s" "$OUT" | grep -q "feat(b): do the work"'
+end_case ) >"$LOGS/08.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- no_verdict ---
 fixture
 : >"$WORK/fail-findings"
@@ -396,7 +468,10 @@ sut prepare "$WORK/a" demo --accept-outstanding
 check "no_verdict: prepare stops with exit 1" '[ "$RC" -eq 1 ]'
 check "no_verdict: the stop is STOP: no-verdict, not passed by --accept-outstanding" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: no-verdict — check-unfinished-work"'
 check "no_verdict: the gate mark closes stopped" 'tail -1 "$WORK/stages.log" | grep -q "stage end .*-stage flow.unfinished-work-gate -outcome stopped"'
+end_case ) >"$LOGS/09.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- sync_conflict ---
 fixture
 printf 'base side\n' >"$WORK/b/work-b.txt"
@@ -405,7 +480,10 @@ sut prepare "$WORK/a" demo
 check "sync_conflict: prepare stops with exit 1" '[ "$RC" -eq 1 ]'
 check "sync_conflict: the stop line names the --resume call and what follows it" \
   'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: sync-conflict — .*sync-onto-base.sh. --resume, the whole ## lint and ## test lists, then re-run prepare$"'
+end_case ) >"$LOGS/10.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- guard_test ---
 # A scoped re-verification is <agents repo>/scripts/test-<stem>.sh, the repo
 # derived from the guards' own path: run a copy of scripts/ whose stats/ is
@@ -425,7 +503,10 @@ SUT="$SUT_SAVED"
 check "guard_test: prepare stops with exit 1" '[ "$RC" -eq 1 ]'
 check "guard_test: the stop is STOP: guard-test" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: guard-test"'
 check "guard_test: the failing test's own output is relayed" 'printf "%s" "$OUT" | grep -qx "gt-test-own-output"'
+end_case ) >"$LOGS/11.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- rebase_lint ---
 # b's base gains a .flow/project.md whose ## lint list fails and nothing this
 # change touches: the rebase is clean, and the rebased tree's lint list runs.
@@ -438,7 +519,10 @@ check "rebase_lint: prepare stops with exit 1" '[ "$RC" -eq 1 ]'
 check "rebase_lint: the rebase was clean" 'printf "%s" "$OUT" | grep -q "^REBASED: $WORK/b-worktrees/demo"'
 check "rebase_lint: the stop is STOP: rebase-lint" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: rebase-lint — the ## lint list failed on $WORK/b-worktrees/demo"'
 check "rebase_lint: the list's own output is relayed" 'printf "%s" "$OUT" | grep -q "lint-own-output"'
+end_case ) >"$LOGS/12.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- rebase_lint_env: the list runs under prepare-workspace.sh's exports ---
 fixture
 mkdir -p "$WORK/b/.flow"
@@ -446,7 +530,10 @@ printf '# project\n\n## lint\n\n```bash\necho "lint-env:${DEMO_ADDR:-unset}"; ex
 g -C "$WORK/b" add .flow/project.md && g -C "$WORK/b" commit -qm "base: lint list" && g -C "$WORK/b" push -q origin main
 sut prepare "$WORK/a" demo
 check "rebase_lint_env: the list saw the workspace's exported variable" 'printf "%s" "$OUT" | grep -q "lint-env:http://127.0.0.1:2"'
+end_case ) >"$LOGS/13.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- rebase_lint_rewrite: a passing list that rewrote tracked files stops ---
 fixture
 mkdir -p "$WORK/b/.flow"
@@ -455,7 +542,10 @@ g -C "$WORK/b" add .flow/project.md && g -C "$WORK/b" commit -qm "base: lint lis
 sut prepare "$WORK/a" demo
 check "rebase_lint_rewrite: prepare stops with exit 1" '[ "$RC" -eq 1 ]'
 check "rebase_lint_rewrite: the stop names the rewrite" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: rebase-lint — the ## lint list rewrote tracked files on $WORK/b-worktrees/demo"'
+end_case ) >"$LOGS/14.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- rebase_lint_land: land's re-sync after a rejected push runs the list ---
 fixture
 sut prepare "$WORK/a" demo
@@ -473,7 +563,10 @@ sut land "$WORK/a" demo
 check "rebase_lint_land: land stops at rebase-lint naming what was already pushed" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: rebase-lint — already pushed: $WORK/b-worktrees/demo; the ## lint list failed on $WORK/a-worktrees/demo"'
 check "rebase_lint_land: the list's own output is relayed" 'printf "%s" "$OUT" | grep -q "land-lint-output"'
 check "rebase_lint_land: a was not pushed to its base" '[ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -1)" = "base: lint list" ]'
+end_case ) >"$LOGS/15.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- gh_open_merged ---
 fixture
 printf '#!/usr/bin/env bash\n[ "$1 $2" = "pr view" ] && { echo OPEN; exit 0; }\nexit 1\n' >"$WORK/bin/gh"
@@ -483,7 +576,10 @@ to_land "gh_open_merged: "
 sut land "$WORK/a" demo
 check "gh_open_merged: land runs to the Finished handoff" '[ "$RC" -eq 0 ] && printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: print the Finished handoff"'
 check "gh_open_merged: the record is FINISHED" '[ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'
+end_case ) >"$LOGS/16.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- disclose ---
 fixture
 for r in a b; do
@@ -501,7 +597,10 @@ check "disclose: the re-run stops at STOP: disclose for b" \
   '[ "$RC" -eq 1 ] && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: disclose — relay $WORK/b.s "'
 sut cleanup "$WORK/a" demo --proceed "$WORK/a" --proceed "$WORK/b"
 check "disclose: --proceed for both finishes" '[ "$RC" -eq 0 ] && [ ! -d "$WORK/b-worktrees/demo" ] && [ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'
+end_case ) >"$LOGS/17.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- rerun_leftover ---
 fixture
 mkdir -p "$WORK/a/spectre/changes/demo"
@@ -516,7 +615,10 @@ check "rerun_leftover: the re-run reaches b and deletes its remote branch" \
   '! git --git-dir="$WORK/b-origin.git" rev-parse -q --verify refs/heads/spectre/demo >/dev/null'
 check "rerun_leftover: b's cleanup check ran" 'printf "%s" "$OUT" | grep -q "^COMPLETE: .*$WORK/b"'
 check "rerun_leftover: the re-run writes FINISHED" '[ "$RC" -eq 0 ] && [ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'
+end_case ) >"$LOGS/18.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- no_state_dir ---
 fixture
 to_land "no_state_dir: "
@@ -525,7 +627,10 @@ sut land "$WORK/a" demo
 check "no_state_dir: run 2 reaches COMPLETE, not STOP: leftover" \
   '[ "$RC" -eq 0 ] && printf "%s" "$OUT" | grep -q "^COMPLETE: " && ! printf "%s" "$OUT" | grep -q "^STOP: leftover"'
 check "no_state_dir: the record is FINISHED" '[ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'
+end_case ) >"$LOGS/19.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- pr_url ---
 fixture
 printf '#!/usr/bin/env bash\necho "gh: no forge" >&2\nexit 1\n' >"$WORK/bin/gh"
@@ -540,7 +645,10 @@ check "pr_url: b's answer does not answer for a" \
 sut land "$WORK/a" demo --route "pull request" --pr-url "$WORK/b-worktrees/demo=https://forge.invalid/b/1" --pr-url "$WORK/a-worktrees/demo=https://forge.invalid/a/1"
 check "pr_url: with both answered, IN_PROGRESS records the canonical PR" \
   '[ "$RC" -eq 0 ] && [ "$(jq -r "[.state, .prUrl] | join(\" \")" "$WORK/state.json")" = "IN_PROGRESS https://forge.invalid/a/1" ]'
+end_case ) >"$LOGS/20.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- nested_link ---
 # commit-split adds `<spec>/changes/*/link.md` only when a change directory
 # holds a link.md, so b's base carries one; the branch adds a nested one.
@@ -560,7 +668,10 @@ sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b
 check "nested_link: commit exits 0" '[ "$RC" -eq 0 ]'
 check "nested_link: b's implementation commit carries the nested link.md" \
   '[ "$(git -C "$WORK/b-worktrees/demo" log --format=%s --diff-filter=A -1 -- spectre/changes/archive/y/link.md)" = "feat(b): nest a link" ]'
+end_case ) >"$LOGS/21.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- mixed_roots ---
 fixture
 mkdir -p "$WORK/b-worktrees/demo/openspec/changes/demo"
@@ -569,7 +680,10 @@ g -C "$WORK/b-worktrees/demo" add -A && g -C "$WORK/b-worktrees/demo" commit -qm
 sut prepare "$WORK/a" demo
 check "mixed_roots: prepare exits 0" '[ "$RC" -eq 0 ]'
 check "mixed_roots: b's diffstat lists no openspec/changes/ path" '! printf "%s" "$OUT" | grep -q "openspec/changes/"'
+end_case ) >"$LOGS/22.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- planning_only ---
 fixture
 g -C "$WORK/b-worktrees/demo" reset -q --hard main
@@ -585,7 +699,10 @@ sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work"
 check "planning_only: commit with no subject for b exits 0" '[ "$RC" -eq 0 ]'
 check "planning_only: b's branch is the base plus its planning and archive commits" \
   '[ "$(git -C "$WORK/b-worktrees/demo" log --format=%s main..HEAD | tr "\n" "|")" = "chore(spectre): archive demo|chore(spectre): plan demo|" ]'
+end_case ) >"$LOGS/23.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- satellite_archive ---
 fixture
 mkdir -p "$WORK/b-worktrees/demo/spectre/changes/demo"
@@ -596,7 +713,10 @@ check "satellite_archive: b's change directory is archived" \
   '[ ! -d "$WORK/b-worktrees/demo/spectre/changes/demo" ] && [ -f "$WORK/b-worktrees/demo/spectre/changes/archive/demo/link.md" ]'
 check "satellite_archive: b's branch ends on its own archive commit" \
   '[ "$(git -C "$WORK/b-worktrees/demo" log -1 --format=%s)" = "chore(spectre): archive demo" ]'
+end_case ) >"$LOGS/24.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- link_only_canon ---
 fixture
 g -C "$WORK/a-worktrees/demo" reset -q --hard main
@@ -614,48 +734,37 @@ sut commit "$WORK/a" demo "$WORK/b-worktrees/demo=feat(b): do the work"
 check "link_only_canon: commit with no subject for a exits 0" '[ "$RC" -eq 0 ]'
 check "link_only_canon: a's link.md lands in a commit subjected chore(spectre): merge order" \
   '[ "$(git -C "$WORK/a-worktrees/demo" log --format=%s --diff-filter=A -1 -- spectre/changes/archive/demo/link.md spectre/changes/demo/link.md)" = "chore(spectre): merge order" ]'
+end_case ) >"$LOGS/25.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- commit_pushes ---
 fixture
 for r in a b; do g -C "$WORK/$r-worktrees/demo" push -q -u origin spectre/demo; done
 to_land "commit_pushes: " no-report
 check "commit_pushes: each remote branch is the reshaped HEAD" \
   '[ "$(git --git-dir="$WORK/a-origin.git" rev-parse spectre/demo)" = "$(git -C "$WORK/a-worktrees/demo" rev-parse HEAD)" ] && [ "$(git --git-dir="$WORK/b-origin.git" rev-parse spectre/demo)" = "$(git -C "$WORK/b-worktrees/demo" rev-parse HEAD)" ]'
+end_case ) >"$LOGS/26.log" 2>&1 &
+PIDS+=($!)
 
-# peer_fix_case <prefix>: the report names a fix committed in b, b's base
-# moves, and land runs — checked under <prefix>.
-peer_fix_case() {
-  to_land "$1" no-report
-  printf 'fix\n' >"$WORK/b-worktrees/demo/fix-b.txt"
-  g -C "$WORK/b-worktrees/demo" add fix-b.txt && g -C "$WORK/b-worktrees/demo" commit -qm "fix(b): the peer fix"
-  old="$(git -C "$WORK/b-worktrees/demo" rev-parse --short HEAD)"
-  mkdir -p "$WORK/a-worktrees/demo/docs/self-review"
-  printf '# demo self-review\n\n- **[flow-fix]** the peer fix — fixed: %s\n' "$old" >"$WORK/a-worktrees/demo/docs/self-review/demo-self-review.md"
-  g -C "$WORK/a-worktrees/demo" add docs/self-review/demo-self-review.md
-  g -C "$WORK/a-worktrees/demo" commit -qm "docs(self-review): demo self-review report"
-  printf 'moved\n' >"$WORK/b/moved.txt"
-  g -C "$WORK/b" add moved.txt && g -C "$WORK/b" commit -qm "base: moved" && g -C "$WORK/b" push -q origin main
-  sut land "$WORK/a" demo
-  new="$(git --git-dir="$WORK/b-origin.git" log --format=%h --grep="^fix(b): the peer fix$" -1 main)"
-  check "${1}land exits 0" '[ "$RC" -eq 0 ]'
-  check "${1}the re-sync rewrote b's fix sha" '[ -n "$new" ] && [ "$new" != "$old" ]'
-  check "${1}the report on a's origin main names the rewritten sha" \
-    'git --git-dir="$WORK/a-origin.git" show main:docs/self-review/demo-self-review.md | grep -q "fixed: $new$"'
-  check "${1}a fixed row is recorded for b's fix" \
-    'grep -q -- "-disposition fixed -ref $new " "$WORK/fixed.log"'
-}
-
+( trap cleanup_work EXIT
 # --- peer_fix ---
 fixture
 peer_fix_case "peer_fix: "
+end_case ) >"$LOGS/27.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- peer_fix_canon_first ---
 fixture
 printf '# demo\n\n## Parts\n\n`b:demo`\n\n## Branch\n\nmain\n\n## Merge order\n\n1. `.`\n2. `b`\n' \
   >"$WORK/a-worktrees/demo/spectre/changes/demo/link.md"
 g -C "$WORK/a-worktrees/demo" commit -qam "chore(spectre): plan-gate demo" --amend
 peer_fix_case "peer_fix_canon_first: "
+end_case ) >"$LOGS/28.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- pre_reshape_pin ---
 fixture
 pa="$(git -C "$WORK/a-worktrees/demo" rev-parse HEAD)"; pb="$(git -C "$WORK/b-worktrees/demo" rev-parse HEAD)"
@@ -683,7 +792,10 @@ printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/cha
 sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
 check "pre_reshape_pin: a differing pin on origin stops commit, untouched" \
   'printf "%s" "$OUT" | grep -q "^STOP: pre-reshape" && [ "$(git --git-dir="$WORK/b-origin.git" rev-parse refs/flow/pre-reshape/demo)" = "$stray" ]'
+end_case ) >"$LOGS/29.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- rerun_pin ---
 rerun_pin_case() {
   fixture
@@ -702,22 +814,11 @@ rerun_pin_case() {
 }
 rerun_pin_case "rerun_pin: "
 rerun_pin_case "rerun_pin reclone: " reclone
+end_case ) >"$LOGS/30.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- visual_stop ---
-# ui_a: a declares `ui paths` its task commit touches, and carries the
-# verifier's record a verifier row counts only with.
-ui_a() {
-  printf '\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `work-a.txt` |\n' >>"$WORK/a-worktrees/demo/.flow/project.md"
-  printf '# Visual verification — demo\n' >"$WORK/a-worktrees/demo/spectre/changes/demo/visual-verification.md"
-  g -C "$WORK/a-worktrees/demo" add spectre/changes/demo/visual-verification.md
-  g -C "$WORK/a-worktrees/demo" commit -qam "chore(spectre): plan-gate demo" --amend
-}
-# flag <worktree> <sha> [<change>]: an operator's false-positive flag on
-# <worktree>'s MISSING verdict row naming HEAD <sha>.
-flag() {
-  printf '[{"change":"%s","guard":"check-visual-verify-dispatched","worktree":"%s","verdict":"VISUAL-VERIFY-MISSING: no verifier — HEAD %s","falsePositive":true,"falsePositiveReason":"hand-verified","flaggedAt":"2026-01-01T00:00:00Z"}]' \
-    "${3:-demo}" "$(cd "$1" && pwd -P)" "$2" >"$WORK/verdicts.json"
-}
 fixture
 ui_a
 A="$WORK/a-worktrees/demo"
@@ -746,7 +847,10 @@ printf '[{"key":"task-1-implementer","outcome":"completed"},{"key":"verify","rol
   "$(git -C "$A" rev-parse HEAD)" >"$WORK/dispatches.json"
 sut prepare "$WORK/a" demo
 check "visual_stop: a verifier of the HEAD passes the gate" '[ "$RC" -eq 0 ]'
+end_case ) >"$LOGS/31.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- visual_archived: an archived re-run with no new work runs the gate ---
 fixture
 ui_a
@@ -760,7 +864,10 @@ check "visual_archived: commit archives, then stops before a's unverified branch
 sut prepare "$WORK/a" demo
 check "visual_archived: the re-run is archived without new work" 'printf "%s\n" "$OUT" | grep -q "^ARCHIVED: demo — without new work"'
 check "visual_archived: a flag on the pre-archive HEAD does not pass the gate" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — "'
+end_case ) >"$LOGS/32.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- visual_rebased: a base move in prepare re-runs the gate ---
 fixture
 ui_a
@@ -771,7 +878,10 @@ g -C "$WORK/a" add moved.txt && g -C "$WORK/a" commit -qm "base: moved" && g -C
 sut prepare "$WORK/a" demo
 check "visual_rebased: a was rebased" 'printf "%s\n" "$OUT" | grep -q "^REBASED: "'
 check "visual_rebased: the rebased HEAD no verifier saw stops prepare" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — rebased onto origin/main after the gate: .*a-worktrees/demo: "'
+end_case ) >"$LOGS/33.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- visual_resync: land's re-sync re-runs the gate before that push ---
 fixture
 ui_a
@@ -798,7 +908,10 @@ check "visual_resync: a was not pushed to its base" '[ "$(git --git-dir="$WORK/a
 sut land "$WORK/a" demo
 check "visual_resync: re-running land after the stop still stops at visual-verify" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — "'
 check "visual_resync: the re-run did not push a's unverified HEAD" '[ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -1)" = "base: moved" ]'
+end_case ) >"$LOGS/34.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- visual_resync_reread: land's re-sync rewrites the canonical report first ---
 # a is canonical and its own fix sha is rebased by the re-sync, so
 # reread_fix_shas commits the rewrite into a; a flag on the pre-rewrite
@@ -834,7 +947,10 @@ unset GIT_COMMITTER_DATE
 check "visual_resync_reread: the re-sync rewrote a's own fix sha" 'printf "%s\n" "$OUT" | grep -q "^FIXED-SHA: $old -> "'
 check "visual_resync_reread: a flag on the pre-rewrite HEAD does not pass the rewrite commit" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — "'
 check "visual_resync_reread: a was not pushed to its base" '[ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -1)" = "base: moved" ]'
+end_case ) >"$LOGS/35.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- visual_reread: reread_fix_shas's canonical re-push runs the gate ---
 # The report is a's declared UI path, so the rewrite commit alone — all a's
 # HEAD adds past its landed push — is a UI change no verifier saw.
@@ -869,7 +985,10 @@ check "visual_reread: the rewrite of b's fix sha stops the canonical re-push" \
   'printf "%s\n" "$OUT" | grep -q "^FIXED-SHA: $old -> " && printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — the rewritten report was not pushed"'
 check "visual_reread: neither a's branch nor its base carries the rewritten report" \
   'git --git-dir="$WORK/a-origin.git" show spectre/demo:docs/self-review/demo-self-review.md | grep -q "fixed: $old$" && git --git-dir="$WORK/a-origin.git" show main:docs/self-review/demo-self-review.md | grep -q "fixed: $old$"'
+end_case ) >"$LOGS/36.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- visual_uncommitted: an uncommitted UI edit folded in by commit stops every push ---
 for route in "merge and push" "pull request" manual; do
   fixture
@@ -893,7 +1012,10 @@ for route in "merge and push" "pull request" manual; do
   check "visual_uncommitted ($route): land stops at visual-verify" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: visual-verify — "'
   check "visual_uncommitted ($route): land pushed neither a's branch nor its base" '! git --git-dir="$WORK/a-origin.git" rev-parse -q --verify spectre/demo >/dev/null && [ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -1)" = base ]'
 done
+end_case ) >"$LOGS/37.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- base_mismatch ---
 fixture
 for r in a b; do g -C "$WORK/$r" push -q origin main:refs/heads/feature/x; done
@@ -911,7 +1033,10 @@ g -C "$WORK/a" push -q origin main:refs/heads/feature/x
 g -C "$WORK/a-worktrees/demo" config branch.spectre/demo.flowBase feature/x
 sut prepare "$WORK/a" demo
 check "base_mismatch: feature/x on a's origin only passes the gate" '! printf "%s" "$OUT" | grep -q "STOP: base-mismatch" && grep -q flow.preflight "$WORK/stages.log"'
+end_case ) >"$LOGS/38.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- refresh_on_recorded_base ---
 fixture
 g -C "$WORK/a" push -q origin main:refs/heads/feature/x
@@ -934,7 +1059,10 @@ check "refresh_on_recorded_base: a's main checkout on its recorded base is refre
 sut cleanup "$WORK/a" demo
 check "refresh_on_recorded_base: a cleanup re-run after worktree removal still refreshes onto feature/x" \
   'printf "%s\n" "$OUT" | grep -qE "^REFRESH-(DONE|CURRENT): $WORK/a( |\$)" && ! printf "%s\n" "$OUT" | grep -q "^REFRESH-REFUSED: $WORK/a "'
+end_case ) >"$LOGS/39.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- orphan_sweep ---
 fixture
 printf '\n## workspace isolation\n\n| Command | Runs |\n|---------|------|\n| `remove` | `echo <id> >>%s/removed.log; [ <id> != orph-bad ]` |\n| `orphans` | `printf "orph-1\\norph-bad\\n"` |\n' "$WORK" >>"$WORK/a/.flow/project.md"
@@ -946,7 +1074,10 @@ check "orphan_sweep: each orphan id goes through the remove command" '[ "$(cat "
 check "orphan_sweep: a removed orphan is reported" 'printf "%s\n" "$OUT" | grep -qx "WORKSPACE-ORPHAN-REMOVED: orph-1"'
 check "orphan_sweep: a failed orphan removal is reported and cleanup still finishes" \
   'printf "%s\n" "$OUT" | grep -qx "WORKSPACE-ORPHAN-REMOVE-FAILED: orph-bad — exit 1" && [ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'
+end_case ) >"$LOGS/40.log" 2>&1 &
+PIDS+=($!)
 
+( trap cleanup_work EXIT
 # --- outstanding ---
 fixture
 sut prepare "$WORK/a" demo
@@ -958,6 +1089,20 @@ sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b
 check "outstanding: commit exits 0" '[ "$RC" -eq 0 ]'
 check "outstanding: the planning commit keeps its own subject and carries the text in its body" \
   '[ "$(g -C "$WORK/a-worktrees/demo" log --format=%s --grep="^Outstanding: " -1)" = "chore(spectre): plan demo" ] && [ "$(g -C "$WORK/a-worktrees/demo" log --format=%b --grep="^Outstanding: " -1)" = "Outstanding: Task 3 is unticked" ]'
+end_case ) >"$LOGS/41.log" 2>&1 &
+PIDS+=($!)
 
+for pid in "${PIDS[@]}"; do wait "$pid"; done
+for log in "$LOGS"/*.log; do
+  tally="$(tail -1 "$log")"
+  case "$tally" in
+    *" passed, "*" failed")
+      sed '$d' "$log"
+      f="${tally#* passed, }"
+      PASS=$((PASS + ${tally%% *})); FAIL=$((FAIL + ${f%% *})) ;;
+    *) cat "$log"; fail "$log ended without its tally" ;;
+  esac
+done
+rm -rf "$LOGS"
 printf '\n%d passed, %d failed\n' "$PASS" "$FAIL"
 [ "$FAIL" -eq 0 ]
````

  - [ ] **Step 3: Verify** — `bash -n scripts/test-integrate-change.sh` exits 0; `time bash scripts/test-integrate-change.sh`
    exits 0, ends `195 passed, 0 failed`, and its wall is a fraction of Step 1's.
    <!-- measured: time bash scripts/test-integrate-change.sh with the diff applied @ branch spectre/kan-918-flow-improvement-port-the-git-snapshot-and → 195 passed, 52 s -->

**Build:** green
**Files:** `scripts/test-integrate-change.sh`
**Tests:** none — repairs existing tests: `scripts/test-integrate-change.sh`, restructured to run each case in its own background subshell
**Regression:** the harness's own checks — a case that fails or halts still fails the run, so reverting changes only the wall.
**Commit:** perf(scripts): run the integrate-change harness cases in parallel
**After:** none

**Decision:** integrate-harness-parallel
