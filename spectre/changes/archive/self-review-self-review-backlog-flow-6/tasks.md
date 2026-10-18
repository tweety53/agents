# self-review-self-review-backlog-flow-6

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the three non-`big` findings of `self-review-backlog-flow-6`'s self-review pass —
concurrent inline dispatch rows getting the wrong token share, a plan task creating a worktree,
and the verify runner every parent hand-builds.

**Architecture:** one guard in `DispatchWindowsForSession`'s cap, so a sibling inline begin under
`minInlineWindow` no longer cuts an open row; one rule in `skills/flow/brainstorm-planner.md`
section D; a new bash `scripts/run-verify-list.sh`, its harness, its `skills/flow/scripts/`
symlink and one citing sentence in `skills/flow/verify-and-handoff.md`.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-18-self-review-self-review-backlog-flow-6-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-6`, branch `spectre/self-review-backlog-flow-6`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- Each task's diff is the unified diff its fix step gives, applied with `git apply` from the worktree's root; all three were applied in order and checked in this worktree at `95deb422`. Save each by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork. Each prose sentence lands character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

## Review Focus

- Task 1: a later inline begin at or past `minInlineWindow` still caps an open row (`TestDispatchWindowsForSessionCapsAStagelessOpenRowAtTheNextInlineBegin` stays green); only a closer one is treated as a sibling.
- Task 3: every command runs with stdin `</dev/null` from the worktree, the list never stops at a failure, and the exit codes match the header, the undeclared list's exit 3 included.

- [x] 1. Leave a concurrent inline sibling's dispatch window open
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Save Step 2's diff, apply its
    test hunk alone (`git apply --include='stats/internal/store/records_test.go' <the saved diff>`)
    and run `cd stats && make -s web-dist && go test ./internal/store/ -run 'TestDispatchWindowsForSessionLeavesAConcurrentInlineSiblingOpen' -count=1`:
    it fails, the first window's `EndedAt` set to the sibling's begin 30 ms after its own.
    <!-- measured: the test hunk alone, go test -run TestDispatchWindowsForSession @ 95deb422 → FAIL, first window capped 30 ms in -->
  - [x] **Step 2: Apply the fix** — the rest (`git apply --exclude='stats/internal/store/records_test.go' <the saved diff>`):

````diff verified:applied with git apply to this worktree at 95deb422; go test ./internal/store/ ./internal/harvest/ -count=1 passes, the new test fails with records.go reverted, go vet and gofmt -l clean
diff --git a/stats/internal/store/records.go b/stats/internal/store/records.go
index b442b0a6..77223d9e 100644
--- a/stats/internal/store/records.go
+++ b/stats/internal/store/records.go
@@ -1251,7 +1251,12 @@ func (s *Store) MarkDispatchesUnattributed(ctx context.Context, token, reason st
 // stage-less row -- one matched by its own session_id, a loop row or the
 // inline row of a run that marks no stage -- is capped instead at the next
 // inline begin in the same session: a later run taking the main thread is
-// the point the never-ended one stopped holding it. With no such begin the
+// the point the never-ended one stopped holding it. A begin under
+// minInlineWindow after the row's own is no such successor but a sibling
+// begun beside it -- the per-worktree inline verifier rows one Bash call
+// begins together (skills/flow/verify-and-handoff.md, **Recording.**) --
+// so it caps nothing; capping there left the earlier row a 30 ms window
+// and gave the sibling every record of the run. With no such begin the
 // window stays open.
 //
 // Returning harvest.DispatchWindow directly, rather than a store-local
@@ -1265,7 +1270,8 @@ func (s *Store) DispatchWindowsForSession(ctx context.Context, sessionID string)
 		SELECT d.id, d.started_at, COALESCE(d.ended_at, sr.ended_at, (
 			SELECT MIN(n.started_at) FROM dispatches n
 			WHERE d.stage_run_id IS NULL AND n.session_id = d.session_id
-			  AND n.agent_id = 'inline' AND n.id <> d.id AND n.started_at > d.started_at
+			  AND n.agent_id = 'inline' AND n.id <> d.id
+			  AND n.started_at >= d.started_at + $2::float8 * interval '1 second'
 		)), d.agent_id
 		FROM dispatches d
 		LEFT JOIN stage_runs sr ON sr.id = d.stage_run_id
@@ -1274,7 +1280,7 @@ func (s *Store) DispatchWindowsForSession(ctx context.Context, sessionID string)
 			WHERE sr.session_id = $1 AND sr.session_token IS NOT NULL
 		)
 		ORDER BY d.started_at, d.id
-	`, sessionID)
+	`, sessionID, minInlineWindow.Seconds())
 	if err != nil {
 		return nil, fmt.Errorf("store: dispatch windows for session %s: %w", sessionID, err)
 	}
diff --git a/stats/internal/store/records_test.go b/stats/internal/store/records_test.go
index 62c37a46..15920e0d 100644
--- a/stats/internal/store/records_test.go
+++ b/stats/internal/store/records_test.go
@@ -4319,3 +4319,37 @@ func TestDispatchWindowsForSessionCapsAStagelessOpenRowAtTheNextInlineBegin(t *t
 		t.Fatalf("windows = %+v, want the first capped at %v and the latest open", windows, next.StartedAt)
 	}
 }
+
+// TestDispatchWindowsForSessionLeavesAConcurrentInlineSiblingOpen pins the
+// parallel per-worktree verifier rows: an inline begin under minInlineWindow
+// after another is a sibling running beside it, never its successor, so it
+// does not cap the earlier row's open window.
+func TestDispatchWindowsForSessionLeavesAConcurrentInlineSiblingOpen(t *testing.T) {
+	st, _ := newRecordStore(t)
+	ctx := context.Background()
+	projectKey := fmt.Sprintf("proj-inline-sibling-%d", time.Now().UnixNano())
+	seedChange(t, st, projectKey, "kan-1")
+
+	const sessionID = "session-inline-sibling"
+	first := baseDispatch("verifier", "opus")
+	first.SessionToken = "tok-a"
+	first.SessionID = sessionID
+	first.AgentID = "inline"
+	if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", first); err != nil {
+		t.Fatalf("RecordDispatch: %v", err)
+	}
+	sibling := first
+	sibling.SessionToken = "tok-b"
+	sibling.StartedAt = first.StartedAt.Add(30 * time.Millisecond)
+	if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", sibling); err != nil {
+		t.Fatalf("RecordDispatch: %v", err)
+	}
+
+	windows, err := st.DispatchWindowsForSession(ctx, sessionID)
+	if err != nil {
+		t.Fatalf("DispatchWindowsForSession: %v", err)
+	}
+	if len(windows) != 2 || windows[0].EndedAt != nil || windows[1].EndedAt != nil {
+		t.Fatalf("windows = %+v, want both siblings open", windows)
+	}
+}
````

  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./internal/store/ -run 'TestDispatchWindowsForSession' -count=1`
    passes; `cd stats && go vet ./internal/store && gofmt -l internal/store` prints nothing.
  - [x] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `stats/internal/store/records.go`, `stats/internal/store/records_test.go`
**Tests:** `TestDispatchWindowsForSessionLeavesAConcurrentInlineSiblingOpen`
**Regression:** `TestDispatchWindowsForSessionLeavesAConcurrentInlineSiblingOpen` — reverting the commit caps the first of two inline rows begun 30 ms apart at the second's begin, and the test fails on its `EndedAt`.
**Baseline:** before=84 after=85
<!-- measured: grep -c '^func Test' stats/internal/store/records_test.go @ 95deb422 -->
**Commit:** fix(store): leave a concurrent inline sibling's dispatch window open
**After:** none

**Decision:** inline-sibling-min-window

- [x] 2. Tell the planner a task never creates a worktree
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n "git worktree add" skills/flow/brainstorm-planner.md; echo $?`: it prints only `1`.
    <!-- measured: the grep above in this worktree @ 95deb422 → exit 1, no line -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at 95deb422; check-verbatim-moves.sh exits 0 with this change's verbatim-moves.txt, and the runner's lint list over the tree fails only check-plan-shape.sh on the then-stub tasks.md
diff --git a/skills/flow/brainstorm-planner.md b/skills/flow/brainstorm-planner.md
index a4a98ce7..a71404ea 100644
--- a/skills/flow/brainstorm-planner.md
+++ b/skills/flow/brainstorm-planner.md
@@ -332,6 +332,11 @@ paragraph under **The build-green tag** (`skills/flow-contracts/build-green.md`)
 > code change, while an edit made before reproducing destroys the evidence the task set out to
 > gather.
 
+> **Never write a task that creates a worktree.** Every worktree a plan runs in exists before its
+> first task: **2. Isolate the workspace** (`skills/flow/implement.md`) creates each `## apps`
+> repository's on branch `spectre/<name>`, so a task names that worktree's path and no step
+> carries a `git worktree add` (self-review-backlog-flow-6).
+
 > **End a removal task on an absence sweep.** When a task's job is removing text — copy, a setting,
 > an identifier — its last verify step is one `git grep -n -F` over every string it removes, each
 > passed as its own `-e '<string>'`, scoped to the paths the task touches: the sweep prints
````

  - [x] **Step 3: Verify.** `grep -c "git worktree add" skills/flow/brainstorm-planner.md` prints `1`;
    `scripts/check-references.sh`, `scripts/check-guard-symlinks.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentence.
**Regression:** none — no test is declared.
**Commit:** docs(flow): tell the planner a task never creates a worktree
**After:** none

**Decision:** worktree-rule-prose

- [x] 3. Run the verify lists through `run-verify-list.sh`
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `ls scripts/run-verify-list.sh; grep -c "run-verify-list.sh" skills/flow/verify-and-handoff.md`:
    `ls` reports no such file and `grep` prints `0` — no runner exists.
    <!-- measured: both commands in this worktree @ 95deb422 → ls exit 1, grep 0 -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`) — the runner, its harness, the
    relative symlink `skills/flow/scripts/run-verify-list.sh` → `../../../scripts/run-verify-list.sh`
    and the citing sentence:

````diff verified:applied with git apply to this worktree at 95deb422; bash scripts/test-run-verify-list.sh prints all cases passed and fails the stdin case with </dev/null removed; check-guard-symlinks.sh, check-references.sh and check-verbatim-moves.sh exit 0
diff --git a/scripts/run-verify-list.sh b/scripts/run-verify-list.sh
new file mode 100755
index 00000000..024f2729
--- /dev/null
+++ b/scripts/run-verify-list.sh
@@ -0,0 +1,85 @@
+#!/usr/bin/env bash
+# run-verify-list.sh — run one of a project's verify lists the way **Inline
+# verify** (skills/flow/verify-and-handoff.md) prescribes, and print its
+# `## Report` lines, so no run hand-builds the runner (the
+# self-review-backlog-flow-6 self-review pass found the parent writing one
+# per run, its stdin detach patched in on a second call).
+#
+# Usage: run-verify-list.sh <worktree> lint|test
+#
+# THE LIST is every non-blank, non-`#` line inside a ```bash fence of
+# `project-get.sh <worktree> <list>`'s output, in the order printed; `lint`
+# appends `check-spec-reach.sh <worktree>`. Each command runs from
+# <worktree> in a fresh `bash -c`, stdin detached (</dev/null) so none waits
+# on a terminal, and the whole list always runs: a failure never stops it.
+# The caller exports the KEY=value lines prepare-workspace.sh printed for
+# <worktree> first; every command inherits them. This script retries
+# nothing — the one inline re-run of a failing command stays the caller's.
+#
+# OUTPUT, one block per command in the order run, stdout and stderr together:
+#   - `<command>` — exit <n>
+#     <each output line, indented two spaces; past 20 lines,
+#     `(truncated — last 20 of <N> lines)` and then the last 20>
+#
+# Exit codes:
+#   0  every command exited 0
+#   1  at least one command exited non-zero
+#   2  usage (not exactly `<worktree> lint|test`, or <worktree> is not a
+#      directory), or project-get.sh exit 2 — its stderr line says why
+#   3  the list is not declared (project-get.sh exit 1): the caller
+#      auto-detects, as **Inline verify** says
+set -uo pipefail
+
+SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
+
+if [ "$#" -ne 2 ] || [ ! -d "$1" ] || { [ "$2" != lint ] && [ "$2" != test ]; }; then
+  echo "usage: run-verify-list.sh <worktree> lint|test" >&2
+  exit 2
+fi
+WT="$1"
+LIST="$2"
+
+BODY="$("$SCRIPT_DIR/project-get.sh" "$WT" "$LIST")"
+case $? in
+  0) ;;
+  1) exit 3 ;;
+  *) exit 2 ;;
+esac
+
+CMDS=()
+while IFS= read -r line; do
+  CMDS+=("$line")
+done < <(printf '%s\n' "$BODY" | awk '
+  /^```bash[[:space:]]*$/ { f = 1; next }
+  /^```/                  { f = 0; next }
+  f && NF && $1 !~ /^#/
+')
+
+OUT="$(mktemp "${TMPDIR:-/tmp}/run-verify-list.XXXXXX")"
+trap 'rm -f "$OUT"' EXIT
+
+FAILED=0
+# run <command-as-shown> <command-to-run>
+run() {
+  local rc n
+  (cd "$WT" && bash -c "$2") </dev/null >"$OUT" 2>&1
+  rc=$?
+  [ "$rc" -eq 0 ] || FAILED=1
+  printf -- '- `%s` — exit %d\n' "$1" "$rc"
+  n="$(wc -l <"$OUT" | tr -d ' ')"
+  if [ "$n" -gt 20 ]; then
+    printf '  (truncated — last 20 of %d lines)\n' "$n"
+    tail -n 20 "$OUT" | sed 's/^/  /'
+  else
+    sed 's/^/  /' "$OUT"
+  fi
+}
+
+for i in "${!CMDS[@]}"; do
+  run "${CMDS[$i]}" "${CMDS[$i]}"
+done
+if [ "$LIST" = lint ]; then
+  run "check-spec-reach.sh $WT" "\"$SCRIPT_DIR/check-spec-reach.sh\" \"$WT\""
+fi
+
+exit "$FAILED"
diff --git a/scripts/test-run-verify-list.sh b/scripts/test-run-verify-list.sh
new file mode 100755
index 00000000..9edc1d25
--- /dev/null
+++ b/scripts/test-run-verify-list.sh
@@ -0,0 +1,84 @@
+#!/usr/bin/env bash
+# Assertion harness for run-verify-list.sh.
+#
+# Builds throwaway project roots under TMPDIR, each with a `.flow/project.md`
+# whose `## lint` / `## test` fences hold small shell commands, and runs the
+# REAL scripts/run-verify-list.sh against them — never a copy of its logic.
+# The fixture roots are no git repositories, so project-get.sh reads the file
+# on disk and check-spec-reach.sh answers `Spec reach: not configured`.
+set -euo pipefail
+
+SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
+BIN="$SCRIPT_DIR/run-verify-list.sh"
+FAILURES=0
+
+fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
+pass() { printf 'ok: %s\n' "$1"; }
+
+ROOT="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/run-verify-list-root.XXXXXX")" && pwd)"
+trap 'rm -rf "$ROOT"' EXIT
+mkdir -p "$ROOT/.flow"
+cat > "$ROOT/.flow/project.md" <<'CFG'
+## lint
+
+Prose outside the fence is never run.
+
+```bash
+# a comment line is never run
+printf 'one\n'
+printf 'to-stderr\n' >&2; exit 3
+cat
+seq 1 25
+pwd
+```
+
+## test
+
+```bash
+printf 'ok\n'
+```
+CFG
+
+# run_bin <args...> -> OUT, RC; stdin carries a line a command must never read.
+run_bin() {
+  set +e
+  OUT="$(printf 'LEAKED-STDIN\n' | "$BIN" "$@" 2>&1)"
+  RC=$?
+  set -e
+}
+
+expect_rc() { if [ "$RC" -eq "$2" ]; then pass "$1: exit $2"; else fail "$1: exit $RC, want $2"; fi; }
+expect_has() { case "$OUT" in *"$2"*) pass "$1" ;; *) fail "$1: missing [$2] in: $OUT" ;; esac; }
+expect_lacks() { case "$OUT" in *"$2"*) fail "$1: unexpected [$2] in: $OUT" ;; *) pass "$1" ;; esac; }
+
+run_bin "$ROOT" lint
+expect_rc "a failing command fails the list" 1
+expect_has "first command reported" $'- `printf \'one\\n\'` — exit 0\n  one'
+expect_has "stderr captured and exit kept" $'exit 3\n  to-stderr'
+expect_has "the list runs on past a failure" '- `pwd` — exit 0'
+expect_lacks "stdin is detached" 'LEAKED-STDIN'
+expect_lacks "comments and prose never run" 'comment line'
+expect_has "long output truncated to its last 20" $'(truncated — last 20 of 25 lines)\n  6\n'
+expect_lacks "truncated head dropped" $'\n  5\n'
+expect_has "commands run from the worktree" "  $ROOT"
+expect_has "lint appends check-spec-reach.sh" $'- `check-spec-reach.sh '"$ROOT"$'` — exit 0\n  Spec reach: not configured'
+
+run_bin "$ROOT" test
+expect_rc "a passing list" 0
+expect_has "test list reported" $'- `printf \'ok\\n\'` — exit 0\n  ok'
+expect_lacks "test never appends check-spec-reach.sh" 'check-spec-reach.sh'
+
+printf '## lint\n\n```bash\ntrue\n```\n' > "$ROOT/.flow/project.md"
+run_bin "$ROOT" test
+expect_rc "an undeclared list" 3
+
+run_bin "$ROOT" build
+expect_rc "an unknown list" 2
+run_bin "$ROOT/missing" lint
+expect_rc "a missing worktree" 2
+
+if [ "$FAILURES" -ne 0 ]; then
+  printf '%s case(s) failed\n' "$FAILURES" >&2
+  exit 1
+fi
+printf 'all cases passed\n'
diff --git a/skills/flow/scripts/run-verify-list.sh b/skills/flow/scripts/run-verify-list.sh
new file mode 120000
index 00000000..8837d041
--- /dev/null
+++ b/skills/flow/scripts/run-verify-list.sh
@@ -0,0 +1 @@
+../../../scripts/run-verify-list.sh
\ No newline at end of file
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index 55059c8a..24c00751 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -106,6 +106,8 @@ The parent writes this `## Report` itself and shows it as this stage's output.
 call; the lint and test run, this run's `flow record dispatch begin`, and the ledger render below
 are one more.
 
+`run-verify-list.sh <worktree> lint` and `run-verify-list.sh <worktree> test` are the two groups: each runs its list as above, stdin detached, and prints that list's `## Report` lines, its exit codes canonical in its header — never a runner built by hand.
+
 **Inline verify — a failing command.** A non-zero exit from any command in the list earns **one**
 inline re-run of that command — the environmental-flake case. A second non-zero exit from the same
 command is the branch's own defect and takes **The loop** (`skills/flow/verify-fix-loop.md`) once
````

  - [x] **Step 3: Verify.** `bash scripts/test-run-verify-list.sh` prints `all cases passed`;
    `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh`,
    `scripts/check-installed-citations.sh`, `python3 scripts/check-markdown-integrity.py`
    and `scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 4: Commit** the four paths with the subject below.

**Build:** green
**Files:** `scripts/run-verify-list.sh`, `scripts/test-run-verify-list.sh`, `skills/flow/scripts/run-verify-list.sh`, `skills/flow/verify-and-handoff.md`
**Tests:** `test-run-verify-list.sh`
**Regression:** `test-run-verify-list.sh` — reverting the commit removes the runner, and the harness's every case fails on a missing `run-verify-list.sh`.
**Baseline:** before=43 after=44
<!-- measured: ls scripts/test-*.sh | grep -c . @ 95deb422 -->
**Commit:** feat(scripts): run a project's verify list through run-verify-list.sh
**After:** none

**Decision:** runner-is-bash
**Decision:** runner-contract
**Decision:** inline-verify-keeps-prose
