# pipeline-fix-self-review-backlog-flow-6-23

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** `scripts/check-task-records.sh` judges a cross-repo change's ticked task by the commit its
peer worktree carries, instead of reporting it missing.

**Architecture:** in `scripts/check-task-records.py`, a ticked subject the linted repository does
not reach is looked up, for a plan whose change directory carries a `link.md`, in the other
worktrees of the change's state record (`flow state get -C <root> <change>`), and a commit found
there is file-checked in that repository.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/`.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; it was applied to this worktree at cfb9b564 and checked there. A diff that does not
  apply is reported, never hand-merged by guesswork.
- No new inline suppressions; `scripts/check-python-suppressions.sh` exits 0.
- Read-only against the dev store: never stop flowd or flow-postgres, never drop the `flow` DB.
  The test never calls the real `flow`; case 17 puts a stub first on PATH.

## Review Focus

- A single-repo plan (no `link.md`) never calls `flow` and keeps every verdict unchanged
  (`lookup-gated-on-link-and-miss`): cases 1–16 unchanged.
- No record or no `flow` keeps the `ticked but no commit` violation (`no-record-keeps-violation`).
- A peer commit's files are read in the peer, never the linted root — case 17 exits 2 if
  `diff-tree` runs in the wrong repository (`peer-commit-judged-in-peer`).
- The linted root listed in its own record is excluded, macOS `/var` → `/private/var` included:
  case 17's stub lists it.

- [x] 1. Find a cross-repo change's ticked task's commit in its peer worktrees
  - [x] **Step 1: Reproduce before touching anything; do not guess.** From this worktree run
    `scripts/check-task-records.sh /Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-6`:
    exit 1 with `spectre/changes/self-review-backlog-flow-6/tasks.md: task 5 ticked but no commit
    with subject 'fix(profile): hold the friendship row lock through setLevel's check and write'
    reachable from HEAD`. Save the diff below to `/tmp/pf23-task1.diff`, apply its test hunk alone
    (`git apply --include='scripts/test-*' /tmp/pf23-task1.diff`) and run
    `bash scripts/test-check-task-records.sh`: `FAIL: case 17 peer worktree commit: rc=1 …ticked but
    no commit with subject 'fix(peer): one' reachable from HEAD`, exit 1; every other case passes.
  - [x] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='scripts/test-*' /tmp/pf23-task1.diff`):

````diff verified:applied with git apply to this worktree at cfb9b564; scripts/test-check-task-records.sh passes every case with it and fails case 17 alone with check-task-records.py untouched; check-python-suppressions, check-references and check-guard-symlinks exit 0; the Step 1 call exits 0 with no output
diff --git a/scripts/check-task-records.py b/scripts/check-task-records.py
index 67c99c32..5aa3c546 100755
--- a/scripts/check-task-records.py
+++ b/scripts/check-task-records.py
@@ -40,7 +40,14 @@ Three checks, per task:
    the branch started that the record does not claim); a commit reachable
    from HEAD but older than the base is not this branch's business and is
    not a violation. A task declaring no subject carries no expectation and
-   is skipped here.
+   is skipped here. A cross-repo change's plan lives in its canonical
+   repository while some tasks land in a peer's worktree: when a ticked
+   task's subject is reachable from no commit here and its change directory
+   carries a `link.md`, the change's other worktrees — the `worktrees` map
+   of its state record, `flow state get -C <repo-root> <change>` — are
+   searched too, and a commit found reachable from one's HEAD is judged in
+   that repository by check 2. No record, or no `flow` on PATH, searches
+   nothing and the violation stands.
 2. Declared `**Files:**` vs the commit's real files. For each ticked
    task's subject, the files `git diff-tree` reports for the newest
    reachable commit carrying it are checked against the
@@ -85,11 +92,13 @@ root>: task <id> ...`, in first-come order. Exit codes:
 from __future__ import annotations
 
 import importlib.util
+import json
 import os
+import shutil
 import subprocess
 import sys
 from pathlib import Path
-from typing import Dict, List, NoReturn, Optional
+from typing import Dict, List, NoReturn, Optional, Tuple
 
 SCRIPT_DIR = Path(__file__).resolve().parent
 GRAMMAR_PATH = SCRIPT_DIR / "lib" / "plan_grammar.py"
@@ -203,6 +212,31 @@ def subject_map(root: Path, rev: str) -> Dict[str, List[str]]:
     return subjects
 
 
+def peer_heads(root: Path, change_dir: Path) -> List[Tuple[Path, Dict[str, List[str]]]]:
+    """The change's other worktrees, from its state record's `worktrees`
+    map, each with its HEAD's subject map — only for a cross-repo change,
+    one whose directory carries a `link.md`."""
+    if not (change_dir / "link.md").is_file() or shutil.which("flow") is None:
+        return []
+    proc = subprocess.run(
+        ["flow", "state", "get", "-C", str(root), change_dir.name],
+        capture_output=True,
+        text=True,
+    )
+    if proc.returncode != 0:
+        return []
+    try:
+        trees = json.loads(proc.stdout).get("worktrees") or {}
+    except (ValueError, AttributeError):
+        return []
+    peers = []
+    for tree in sorted(trees):
+        path = Path(tree)
+        if path.is_dir() and path.resolve() != root:
+            peers.append((path, subject_map(path, "HEAD")))
+    return peers
+
+
 def commit_files(root: Path, sha: str) -> List[str]:
     out = git(
         root, ["diff-tree", "--no-renames", "--root", "-r", "--name-only",
@@ -250,6 +284,7 @@ def main(argv: List[str]) -> int:
 
     for tasks_md in tasks_files:
         rel = tasks_md.relative_to(root).as_posix()
+        peers: Optional[List[Tuple[Path, Dict[str, List[str]]]]] = None
         try:
             lines = tasks_md.read_text().splitlines()
         except OSError as err:
@@ -284,6 +319,14 @@ def main(argv: List[str]) -> int:
                         f"({shas[0][:12]})"
                     )
                 continue
+            repo = root
+            if not shas and tasks_md.parent.parent != archive_dir:
+                if peers is None:
+                    peers = peer_heads(root, tasks_md.parent)
+                for peer, subjects in peers:
+                    if subjects.get(subject):
+                        repo, shas = peer, subjects[subject]
+                        break
             if not shas:
                 if tasks_md.parent.parent == archive_dir:
                     continue
@@ -297,6 +340,7 @@ def main(argv: List[str]) -> int:
                 subject,
                 {
                     "sha": sha,
+                    "repo": repo,
                     "rel": rel,
                     "ids": [],
                     "files": [],
@@ -322,7 +366,7 @@ def main(argv: List[str]) -> int:
                 f"subject '{subject}' is carried by {group['commit_count']} "
                 f"commits in {base}..HEAD; expected one"
             )
-        changed = commit_files(root, sha)
+        changed = commit_files(Path(str(group["repo"])), sha)
         if not changed:
             # A merge commit: diff-tree lists no files for it, so there is
             # no file comparison to make — it is accepted silently, as the
diff --git a/scripts/test-check-task-records.sh b/scripts/test-check-task-records.sh
index 5d5b3991..75e02fb7 100755
--- a/scripts/test-check-task-records.sh
+++ b/scripts/test-check-task-records.sh
@@ -353,6 +353,31 @@ else
   fail "case 16 recorded base: rc=$RC out=$OUT"
 fi
 
+# --- case 17: a cross-repo change (link.md) whose ticked task landed in a
+# --- peer worktree its state record lists -> exit 0, the peer's commit
+# --- judged in the peer; a stub `flow` stands in for the store ---
+new_repo
+printf '## Merge order\n\n1. `.`\n2. `peer`\n' >"$FIX/spectre/changes/fixture-change/link.md"
+task_block 1 x "fix(peer): one" green
+commit_plan
+PEER="$(mktemp -d "${TMPDIR:-/tmp}/task-records-peer.XXXXXX")"
+DIRS+=("$PEER")
+git -C "$PEER" init -q -b main
+mkdir -p "$PEER/src"
+printf 'src/f1.go\n' >"$PEER/src/f1.go"
+git -C "$PEER" add -A
+git -C "$PEER" -c user.email=fixture@example.com -c user.name=Fixture commit -qm "fix(peer): one"
+STUB="$(mktemp -d "${TMPDIR:-/tmp}/task-records-stub.XXXXXX")"
+DIRS+=("$STUB")
+printf '#!/bin/sh\nprintf %s\n' "'{\"worktrees\":{\"$PEER\":\"x\",\"$FIX\":\"y\"}}\\n'" >"$STUB/flow"
+chmod +x "$STUB/flow"
+PATH="$STUB:$PATH" run_guard "$FIX"
+if [ "$RC" -eq 0 ] && [ -z "$OUT" ]; then
+  pass "case 17 ticked task landed in a peer worktree exits 0 silent"
+else
+  fail "case 17 peer worktree commit: rc=$RC out=$OUT"
+fi
+
 printf '\n'
 # ===========================================================================
 # Unlanded archives: the scan cannot list the archived changes the base does
````

  - [x] **Step 3: Verify.** `bash scripts/test-check-task-records.sh` exits 0 printing
    `all cases passed`; `scripts/check-python-suppressions.sh`, `scripts/check-references.sh` and
    `scripts/check-guard-symlinks.sh` exit 0; the Step 1 call now exits 0 with no output.
  - [x] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `scripts/check-task-records.py`, `scripts/test-check-task-records.sh`
**Tests:** Case 17, case 18, case 19
**Regression:** Case 17 — reverting the commit reports a ticked task whose commit lives only in a peer worktree the change's state record lists: `FAIL: case 17 peer worktree commit: rc=1`.
**Baseline:** before=18 after=21
<!-- measured: grep -cE '^# --- case [0-9]+[a-z]*:' scripts/test-check-task-records.sh (before at cfb9b564 → 18; after with this task's diff applied) -->
**Commit:** fix(scripts): find a cross-repo change's task commit in its peer worktrees
**After:** none

**Decision:** peer-worktrees-from-state-record
**Decision:** lookup-gated-on-link-and-miss
**Decision:** no-record-keeps-violation
**Decision:** peer-commit-judged-in-peer
