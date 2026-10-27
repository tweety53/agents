# self-review-kan-1039-end-to-end-autonomy-rule

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the nine non-`big` findings of the self-review pass of
`kan-1039-end-to-end-autonomy-rule` — 1.1 through 1.6, 2.1, 5.1 and 6.1.

**Architecture:** prose in `skills/flow/` for the begin order, the pass-1 launch order and the
mutation brief; a quote-aware substitution test in the main-checkout hook; four Go fixes in
`stats/internal/guard/` (pipeline-fix verify's script scanners, the derived in-run fix line, the
panel-close fix-mutation check, the verify list's suite recording); the back-fill floor in
`stats/cmd/flow/record.go` and the repair of the row it mis-stamped.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-27-self-review-kan-1039-end-to-end-autonomy-rule-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Every task works in this worktree `/Users/tweety53/Projects/agents-worktrees/kan-1039-end-to-end-autonomy-rule`, branch `spectre/kan-1039-end-to-end-autonomy-rule`; `**Files:**` paths are relative to its root. Never touch a main checkout, and never create a worktree.
- A task whose step gives a unified diff applies it with `git apply` from the worktree's root, in task order; each was applied in that order and checked at plan time. A diff that does not apply is reported, never hand-merged by guesswork. Each prose sentence lands character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container. Task 7's repair writes the dev store's `flow` database through `go run ./cmd/flowd <repair>` with `FLOWD_DSN='postgres://flow:flow@localhost:5433/flow?sslmode=disable'` set on the command, and touches only the row it names.
- Every new test fails with its task's fix reverted: run it once against the reverted source before committing and state the failure in the report.

## Review Focus

- Task 2: a backtick or `$(` inside double quotes, or unquoted, still keeps a heredoc body scanned; an unclosed quote reads exactly as before the change.
- Task 3: a worktree of another project, without `stats/internal/guard` or `stats/internal/repotest`, gets neither scanner line.
- Task 5: a row with an empty or multi-line note, or no blast radius, still stops `narrative`; an appended line matches `icFixFormat`, so a re-run of `commit` reads it as named.
- Task 6: a failed `flow record mutations` read is exit 2, never FINDINGS-CLOSED; a Minor-only or withdrawn finding needs no row.
- Task 8: a `## test` command without a `# suite:` line, a skipped command, and a run with no `flow` on PATH run exactly as before.

- [x] 1. State the begin order the dispatch hook enforces
  - [x] **Step 1: Apply the edits** (decision `begins-before-launch`):

````diff verified:applied with git apply in task order to this worktree's branch spectre/kan-1039-end-to-end-autonomy-rule before any fix commit; check-verbatim-moves, check-references, check-markdown-integrity green
diff --git a/scripts/panel-round-begins.sh b/scripts/panel-round-begins.sh
index 785c018c..9d2b0a39 100755
--- a/scripts/panel-round-begins.sh
+++ b/scripts/panel-round-begins.sh
@@ -3,5 +3,6 @@
 # — records every slot's begin of one review-panel round, each under the
-# literal key panel-<round>-<slot>, in the one Bash call after the round's
-# launch message (skills/flow/review-panel.md, "Every slot's dispatch is
-# recorded"). <round> is the number itself, never a substitution.
+# literal key panel-<round>-<slot>, in the one Bash call just before the
+# round's launch message — hooks/enforce-agent-baseline.py denies a launch
+# whose row is not open yet (skills/flow/review-panel.md, "Every slot's
+# dispatch is recorded"). <round> is the number itself, never a substitution.
 #
diff --git a/skills/flow/SKILL-rationale.md b/skills/flow/SKILL-rationale.md
index 017c51a2..2db84208 100644
--- a/skills/flow/SKILL-rationale.md
+++ b/skills/flow/SKILL-rationale.md
@@ -774,3 +774,3 @@ The invariant is the working tree, not the build tool, and it holds however file
 
-`begin` is recorded immediately before the dispatch — or, when several launches go out together in one message, in the next Bash call, one call for all (step 3 of **The next implementer overlaps the guard**, below) — carrying no `-agent-id`: the daemon captures the agent's identifier — Claude Code writes it into the parent transcript's own launch tool result, the harvester stamps that result onto the row its `description`'s first word names, else pairs it with the begin nearest it in time, and the row's empty `agent_id` is filled from it, which is why `begin` never waits for the launch's id.
+`begin` is recorded immediately before the dispatch — when several launches go out together in one message, in one Bash call for all just before it (step 3 of **The next implementer overlaps the guard**, below) — carrying no `-agent-id`: the daemon captures the agent's identifier — Claude Code writes it into the parent transcript's own launch tool result, the harvester stamps that result onto the row its `description`'s first word names, else pairs it with the begin nearest it in time, and the row's empty `agent_id` is filled from it, which is why `begin` never waits for the launch's id.
 
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index 583964a6..1ecd87aa 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -439,4 +439,4 @@ flow record dispatch end -change <name> -key task-<n>-implementer \
 
-Both calls are required. `begin` is recorded immediately before the dispatch — or, when several
-launches go out together in one message, in the next Bash call, one call for all (step 3 of **The
+Both calls are required. `begin` is recorded immediately before the dispatch — when several
+launches go out together in one message, in one Bash call for all just before it (step 3 of **The
 next implementer overlaps the guard**, below) — carrying no
@@ -809,4 +809,5 @@ entry, one or more bundles `plan-dispatch-bundles.sh` emits. At each boundary, i
    guard on top of it.
-3. **One message launches group N+2's implementer and, on `big` when any gate fired in group N+1,
-   that group's one reviewer bundle (below). The next Bash call records every launch's `begin`**.
+3. **One Bash call records the `begin` of group N+2's implementer and, on `big` when any gate
+   fired in group N+1, of that group's one reviewer bundle (below); the next message launches
+   them all**.
 
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index cb0059a8..8eedd155 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -442,4 +442,4 @@ flow record dispatch end -change <name> -key panel-<round>-<slot|slot+slot+slot>
 
-Every dispatch of a round launches in one message; every `begin` is recorded in the next Bash
-call, one call for all, as `panel-round-begins.sh <worktree> <name> <round> mf-<literal-token>
+Every `begin` of a round is recorded in one Bash call immediately before the one message that
+launches every dispatch of the round, as `panel-round-begins.sh <worktree> <name> <round> mf-<literal-token>
 <slot>:<model>:<effort>[:<diff-base>]...` — the begin above once per slot, under the literal key
@@ -449,2 +449,3 @@ launched pass's report file; every `end` is recorded in one call once they all e
 (SDD + TDD)**, `skills/flow/implement.md`).
+`hooks/enforce-agent-baseline.py` denies a launch whose `panel-<round>-<slot>` row is not open yet, so a round whose begins follow its launch message is refused whole and sent again (kan-1039: all three `panel-0-*` launches were denied).
 On claude-code each `end` passes no `-ended-at`: the daemon closes a stamped row at its own agent transcript's last write, so a slot that finished early is never stamped with the round's last finish (kan-1019: two slots of one round both read 15m31s). On any other harness the daemon reads no agent transcript, so each `end` passes its own slot's report-file modification time as `-ended-at`; a bundled dispatch's `<report-file>` is the newest of its slots' reports.
````
  - [x] **Step 2: Verify.** `scripts/check-verbatim-moves.sh && scripts/check-references.sh && scripts/check-markdown-integrity.sh` — green.

**Build:** green
**Files:** `skills/flow/review-panel.md`, `skills/flow/implement.md`, `skills/flow/SKILL-rationale.md`, `scripts/panel-round-begins.sh`
**Tests:** **none** — prose and a script header; `check-verbatim-moves.sh` covers the acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): record a round's begins before its launch message
**After:** none

- [x] 2. Read a single-quoted backtick as a literal in the main-checkout hook
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run `cd stats && go test ./internal/repotest/ -run '^TestProtectMainCheckout' -count=1` — green on the untouched tree.
  - [x] **Step 2: Apply the test and the fix** (decision `single-quotes-literal`). Apply the diff below. Save its `hooks/` part with `git diff -- hooks/` to a scratch file, `git apply -R` that file, run `cd stats && go test ./internal/repotest/ -run '^TestProtectMainCheckoutReadsASingleQuotedBacktickAsLiteral$' -count=1` and see it fail with `wanted allow, got deny`, then `git apply` the file again.

````diff verified:applied with git apply to this worktree's branch before any fix commit; go test ./internal/repotest/ -count=1 green with it, and the new test failed with "wanted allow, got deny" twice with the hooks/ part reverted
diff --git a/hooks/protect-main-checkout.py b/hooks/protect-main-checkout.py
index 74086d4e..00143c91 100755
--- a/hooks/protect-main-checkout.py
+++ b/hooks/protect-main-checkout.py
@@ -368,6 +368,40 @@ def reads_only(words, target):
     return False
 
 
+def outside_single_quotes(text):
+    """text with every single-quoted span — `'…'`, and `$'…'` with its backslash escapes —
+    blanked, since bash reads a backtick or `$(` inside one as a plain character; a `'` inside
+    `"…"` or behind a backslash is literal. None when a quote never closes, so the caller reads
+    the text as before."""
+    out, i, quote = [], 0, None
+    while i < len(text):
+        c = text[i]
+        if quote in ("'", "$'"):
+            if quote == "$'" and c == "\\":
+                i += 2
+                continue
+            if c == "'":
+                quote = None
+            i += 1
+            continue
+        if c == "\\":
+            out.append(text[i:i + 2])
+            i += 2
+            continue
+        if quote == '"':
+            quote = None if c == '"' else quote
+        elif c == '"':
+            quote = '"'
+        elif c == "'":
+            quote = "$'" if text[i - 1:i] == "$" else "'"
+            out.append(" ")
+            i += 1
+            continue
+        out.append(c)
+        i += 1
+    return None if quote else "".join(out)
+
+
 def command_segments(text):
     """(words, assigned) per command of text — quotes stripped, a leading `do`/`then`/`!`
     dropped, `assigned` when NAME=value words led it — or None when text holds a substitution
@@ -376,7 +410,8 @@ def command_segments(text):
     D names; a name no literal assigns stays an expansion `names_target` reads as hiding any."""
     spans = re.findall(r"\$\(([^()]*)\)", text)
     flat = re.sub(r"\$\([^()]*\)", " ", text)
-    if SUBST_RE.search(flat):
+    bare = outside_single_quotes(flat)
+    if SUBST_RE.search(flat if bare is None else bare):
         return None
     env, out = {}, []
     for seg in COMMAND_CUT_RE.split(flat + "\n" + "\n".join(spans)):
diff --git a/stats/internal/repotest/protect_main_checkout_test.go b/stats/internal/repotest/protect_main_checkout_test.go
index 474fb185..6c104e36 100644
--- a/stats/internal/repotest/protect_main_checkout_test.go
+++ b/stats/internal/repotest/protect_main_checkout_test.go
@@ -652,3 +652,33 @@ func TestProtectMainCheckoutReadsADataHeredocBodyAsData(t *testing.T) {
 	f.bash("continued flow opener", "deny", "{ROOT}", "flow x <<'EOF' \\\n  | sh\nrm {MAIN}/f.txt\nEOF")
 	f.bash("unbalanced ) in body", "deny", "{ROOT}", "flow x <<EOF\n$(case a in a) rm {MAIN}/f.txt;; esac)\nEOF")
 }
+
+// TestProtectMainCheckoutReadsASingleQuotedBacktickAsLiteral pins the
+// kan-1039 self-review: a backtick inside single quotes after a data
+// heredoc is a plain character, so the heredoc body stays data; an
+// unquoted or double-quoted backtick still keeps the body scanned.
+func TestProtectMainCheckoutReadsASingleQuotedBacktickAsLiteral(t *testing.T) {
+	t.Parallel()
+	root, err := filepath.EvalSymlinks(t.TempDir())
+	if err != nil {
+		t.Fatal(err)
+	}
+	f := &pmFixture{t: t, home: t.TempDir(), vars: map[string]string{"ROOT": root}}
+	main := root + "/main"
+	f.vars["MAIN"], f.vars["SIB"] = main, root+"/sib"
+	f.git("init", "-q", "--bare", root+"/origin.git")
+	f.git("clone", "-q", root+"/origin.git", main)
+	f.write(main+"/f.txt", "a\n")
+	f.git("-C", main, "add", "f.txt")
+	f.git("-C", main, "commit", "-q", "-m", "init")
+	f.git("-C", main, "branch", "-M", "main")
+	f.git("-C", main, "push", "-q", "-u", "origin", "main")
+	f.git("-C", main, "remote", "set-head", "origin", "main")
+
+	f.bash("single-quoted backtick after a data heredoc", "allow", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\necho '`'")
+	f.bash("ANSI-quoted backtick after a data heredoc", "allow", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\necho $'\\'`'")
+	f.bash("unquoted backtick keeps the body scanned", "deny", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\necho `true`")
+	f.bash("double-quoted backtick keeps the body scanned", "deny", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\necho \"`true`\"")
+	f.bash("an apostrophe in double quotes opens no span", "deny", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\necho \"it's `true`\"")
+	f.bash("an unclosed quote reads as before", "deny", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\necho '`")
+}
````
  - [x] **Step 3: Verify.** `cd stats && go test ./internal/repotest/ -run '^TestProtectMainCheckout' -count=1 && go vet ./internal/repotest/ && gofmt -l internal/repotest`, then `scripts/check-python-suppressions.sh` — green, `gofmt -l` silent.

**Build:** green
**Files:** `hooks/protect-main-checkout.py`, `stats/internal/repotest/protect_main_checkout_test.go`
**Tests:** `TestProtectMainCheckoutReadsASingleQuotedBacktickAsLiteral`
**Regression:** `TestProtectMainCheckoutReadsASingleQuotedBacktickAsLiteral` — reverting the commit denies the two single-quoted rows.
**Baseline:** before=21 after=22
<!-- measured: cat stats/internal/repotest/*_test.go | grep -c -E '^func Test' @ branch spectre/kan-1039-end-to-end-autonomy-rule -->
**Commit:** fix(hooks): read a backtick inside single quotes as a plain character
**After:** none

- [x] 3. Run the script scanners when pipeline-fix verify sees a touched script
  - [x] **Step 1: Write the failing test.** Add `TestPipelineFixChecksScriptScanners` to `stats/internal/guard/pipeline_fix_test.go`, on `TestPipelineFixChecksGoAudits`' fixture (`pfProject`, `tcfScriptsDir(t) + "/project-get.sh"`), with `stats/internal/guard/doc.go` and `stats/internal/repotest/doc.go` written into the worktree: for each of `scripts/pipeline-fix.sh`, `scripts/lib/flow-guard.sh` and `setup.sh`, `pfChecks` returns rc 0 and its commands contain both `cd stats && go test ./internal/guard/ -run '^Test(NoPipeIntoEarlyExitGrep|ShimMissingLibCannotAnswer)$' -count=1` and `cd stats && go test ./internal/repotest/ -run '^TestGitConfigPins$' -count=1`; for `skills/flow/implement.md` they contain neither; in a second worktree holding neither directory, `scripts/pipeline-fix.sh` selects neither. Run `cd stats && go test ./internal/guard/ -run '^TestPipelineFixChecksScriptScanners$' -count=1`: it fails.
  - [x] **Step 2: Fix.** In `stats/internal/guard/pipelinefix.go`, beside `pfGoAudits`, add the two commands as constants `pfScriptScans` and `pfScriptPins`, commented as the tests that read every shell script of the tree and are named after no script (kan-1039 self-review: a loop landed a script that turned `TestNoPipeIntoEarlyExitGrep` red on main). In `pfChecks`' loop over `touched`, before the `switch`, when `p == "setup.sh"` or `p` has the prefix `scripts/` and the suffix `.sh`, `add(pfScriptScans)` when `isDir(filepath.Join(wt, "stats/internal/guard"))` and `add(pfScriptPins)` when `repotest` (decision `pf-script-scanners`). Extend `pfChecks`' doc comment by the same clause. In `scripts/pipeline-fix.sh`'s header, the `verify` paragraph's list gains `the guard and repotest tests that scan every shell script, for a touched setup.sh or scripts/**/*.sh,` after `the setuptest package for setup.sh,`.
  - [x] **Step 3: Verify.** `cd stats && gofmt -l internal/guard && go vet ./internal/guard/ && go test ./internal/guard/ -run '^TestPipelineFix' -count=1`, then `scripts/check-references.sh` — green.

**Build:** green
**Files:** `stats/internal/guard/pipelinefix.go`, `stats/internal/guard/pipeline_fix_test.go`, `scripts/pipeline-fix.sh`
**Tests:** `TestPipelineFixChecksScriptScanners`
**Regression:** `TestPipelineFixChecksScriptScanners` — reverting the commit leaves a touched script's checks without either scanner line.
**Baseline:** before=344 after=345
<!-- measured: cat stats/internal/guard/*_test.go | grep -c -E '^func Test' @ branch spectre/kan-1039-end-to-end-autonomy-rule -->
**Commit:** fix(guard): run the script scanners when pipeline-fix verify sees a touched script
**After:** none

- [x] 4. Launch the pass-1 mutation bundle once the suite has exited
  - [x] **Step 1: Apply the edits** (decision `mutation-after-suite`):

````diff verified:applied with git apply in task order to this worktree's branch spectre/kan-1039-end-to-end-autonomy-rule before any fix commit; check-verbatim-moves, check-references, check-markdown-integrity green
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index 1ecd87aa..ff9d4c70 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -746,9 +746,10 @@ of `skills/flow/visual-verify.md`.
 **The pass-1 panel reads while the suite runs.** On a run whose `skills/flow/review-panel.md`
 opens a pass 1, the full-suite run above waits for that pass: the parent closes `flow.sdd-tdd`,
 enters the panel, and runs the list there, between recording the pass-1 begins and the round's
-wait, as that file's **Bundled dispatch** states. The slots are read-only, so their verdict and the
+wait, as that file's **Bundled dispatch** states. The reading slots are read-only, so their verdict and the
 suite's arrive together rather than one after the other (kan-920: the panel waited six and a half
-minutes for a suite it did not read). A full-suite fix is applied only after the round's wait ends, never
+minutes for a suite it did not read).
+The mutation slot is not read-only: it runs the tests once per mutant on the same CPUs, so the `mutation+failure-modes` dispatch is begun and launched only once the suite has exited (kan-1039: the guard package hit `go test`'s 10-minute timeout under the mutation slot's runs, and passed alone in 332 s). A full-suite fix is applied only after the round's wait ends, never
 while the slots are reading the tree.
 
 **The next implementer overlaps the guard.** The unit is the group — the decision's `groups`
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index 8eedd155..de7dacb9 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -454,7 +454,9 @@ ceiling takes the breach path under **No forking, and a wall-clock ceiling on ev
 
 **On pass 1 the full suite runs between the begins and the wait.** The parent runs it there, in
 the foreground, while the read-only slots review (**The pass-1 panel reads while the suite
-runs.**, `skills/flow/implement.md`). A full-suite fix is applied only once the round's wait
+runs.**, `skills/flow/implement.md`).
+The `mutation+failure-modes` dispatch stays out of pass 1's launch message: once the suite has exited, a second `panel-round-begins.sh` call records that dispatch's begin alone and the next message launches it, before the round's wait.
+A full-suite fix is applied only once the round's wait
 ends, never under the reading slots, so once the round's reports exist, every pass-1 role re-runs in the next round over the
 delta that commit adds, under **Panel re-runs** (`skills/flow/review-panel-fix-round.md`) — a
 role that raised nothing included; a suite that passes adds no commit and re-dispatches nothing.
````
  - [x] **Step 2: Verify.** `scripts/check-verbatim-moves.sh && scripts/check-references.sh && scripts/check-markdown-integrity.sh` — green.

**Build:** green
**Files:** `skills/flow/review-panel.md`, `skills/flow/implement.md`, `scripts/panel-round-begins.sh`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): launch the pass-1 mutation bundle once the suite has exited
**After:** Task 1

- [x] 5. Derive a missing in-run fix line from its store row at commit
  - [x] **Step 1: Write the failing test.** Add `TestIntegrateChangeDerivesInRunFixLine` to `stats/internal/guard/integrate_change_prepare_test.go`, beside `TestIntegrateChangeRequiresInRunFixLine` and on its `icRunCases` shape. Case `fix_line_derived` writes `[{"angle":"flow-fix","disposition":"fixed","ref":"1a2b3c4d5e6f","note":"the renderer could not render an exp- slot","blastRadius":2}]` to `$WORK/sr-findings.json`, runs `sut commit` as `fix_line_missing` does, and checks exit 0, that `$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md` holds the line `- In-run pipeline fix: 1a2b3c4 — the renderer could not render an exp- slot (blast radius 2 files)`, and that `$OUT` carries `narrative: appended`. Case `fix_line_multiline_note` gives the row a note holding a newline and checks exit 1 with a `STOP: narrative` line naming `1a2b3c4`. `TestIntegrateChangeRequiresInRunFixLine`, whose rows carry no note, stays unchanged and still stops. Run `cd stats && go test ./internal/guard/ -run '^TestIntegrateChangeDerivesInRunFixLine$' -count=1`: `fix_line_derived` fails.
    <!-- predicted: the fixture's own literals; TestIntegrateChangeDerivesInRunFixLine's fix_line_derived check confirms them -->
  - [x] **Step 2: Fix.** In `stats/internal/guard/integratechange_commit.go`, compute `nf` (the live narrative, else the archived one) before the `requireFixLines` call and pass it in. `requireFixLines(named []string, nf string)` also decodes `Note string` and `BlastRadius *int` (JSON key `blastRadius`). For each missing row whose note is non-empty and holds no newline and whose `BlastRadius` is set, while `nf` is a file, it appends `- In-run pipeline fix: <ref, first 7> — <note> (blast radius <N> files)` on a line of its own (adding a newline first when the file does not end in one) and calls `c.say("narrative: appended " + line)`; every other missing row stops `narrative` as today (decision `derive-fix-line`). Rewrite the function's doc comment, and the matching clause of the header comment in `stats/internal/guard/integratechange.go` (the clause on a `flow-fix` row whose sha no in-format line names), to say the line is appended from the row and the stop is for a row lacking its note or blast radius. Then apply the contract's diff:

````diff verified:applied with git apply in task order to this worktree's branch spectre/kan-1039-end-to-end-autonomy-rule before any fix commit; check-verbatim-moves, check-references, check-markdown-integrity green
diff --git a/skills/flow/integrate.md b/skills/flow/integrate.md
index be280592..31b306aa 100644
--- a/skills/flow/integrate.md
+++ b/skills/flow/integrate.md
@@ -151,7 +151,7 @@ archives the change on its branch, commits the archive and pushes each reshaped
 
 **`STOP: narrative`** → an `In-run pipeline fix:` line in the narrative is off the format **Pipeline defects found mid-run** fixes for it (`skills/flow-contracts/pipeline.md`): rewrite that line to that format and re-run `commit`.
 
-**`STOP: narrative`** also stops on a `flow-fix` row `flow self-review findings -change <name>` lists as `fixed` whose sha no `In-run pipeline fix:` line names: append that line, in that format, and re-run `commit`.
+A `flow-fix` row `flow self-review findings -change <name>` lists as `fixed` whose sha no `In-run pipeline fix:` line names has that line appended to `narrative.md` by `commit` itself, from the row's ref, note and blast radius, and printed as `narrative: appended <the line>` (kan-1039: a fix the planning session landed had no narrative to be recorded in). **`STOP: narrative`** stops on such a row only when it carries no note or no blast radius: append that line by hand, in that format, and re-run `commit`.
 
 **`STOP: narrative`** also stops on a fix or integrate run whose section `narrative.md` lacks (`check-narrative-runs.sh`, the check `flow.write-in-progress` runs): append the missing section and re-run `commit`.
 
````
  - [x] **Step 3: Verify.** `cd stats && gofmt -l internal/guard && go vet ./internal/guard/ && go test ./internal/guard/ -run '^TestIntegrateChange' -count=1`, then `scripts/check-verbatim-moves.sh && scripts/check-references.sh && scripts/check-markdown-integrity.sh` — green.

**Build:** green
**Files:** `stats/internal/guard/integratechange_commit.go`, `stats/internal/guard/integratechange.go`, `stats/internal/guard/integrate_change_prepare_test.go`, `skills/flow/integrate.md`
**Tests:** `TestIntegrateChangeDerivesInRunFixLine`
**Regression:** `TestIntegrateChangeDerivesInRunFixLine` — reverting the commit stops `commit` on the row that carries its note and blast radius.
**Baseline:** before=345 after=346
<!-- measured: cat stats/internal/guard/*_test.go | grep -c -E '^func Test' @ branch spectre/kan-1039-end-to-end-autonomy-rule (plus each earlier guard task's one test) -->
**Commit:** fix(guard): derive a missing in-run fix line from its store row at commit
**After:** Task 3

- [x] 6. Refuse the panel's close while a fixed round has no fix-mutation row
  - [x] **Step 1: Write the failing test.** In `stats/internal/guard/check_panel_findings_closed_test.go`, give `cfcBinDispatches`' fake `flow` a `*"record mutations"*)` case answering `$dir/mutations.json` when that file exists and otherwise `$dir/findings.json` (a finding row carries `round`, so a test writing no mutations file proves every round, as today). Add `TestCheckPanelFindingsClosedNeedsAFixMutationRow`: findings `cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "fixed", Round: 1, Slot: "primary"})` with `mutations.json` holding `[]` — stderr names `no fix-mutation row` and round 1, rc 1; the same with `[{"round":2,"path":"a","mutated":"none","test":"doc-only"}]` — stderr lacks that line; a `Minor` fixed finding with `[]` — stderr lacks it; a fake whose `record mutations` exits 1 — rc 2. Run `cd stats && go test ./internal/guard/ -run '^TestCheckPanelFindingsClosedNeedsAFixMutationRow$' -count=1`: it fails.
  - [x] **Step 2: Fix.** In `stats/internal/guard/unfinishedwork.go`, move the fixed-round block into `func uwUnprovedFixRounds(env Env, name, anchor string, findingsJSON []byte) (rounds []int, cause string)`: the rounds, sorted, holding a fixed Critical/Important finding that no `flow record mutations -change <name> -C <anchor>` row of round N or N+1 proves; `cause` is non-empty when the findings or the mutation rows cannot be read, holding today's message after its prefix. `checkUnfinishedWork` calls it with its messages unchanged. In `stats/internal/guard/panelfindingsclosed.go`, after the uncounted-findings check, call it with the canonicalised worktree as `anchor`: a `cause` prints behind `cfcPrefix` and returns 2; rounds print `<cfcPrefix>fixed Critical/Important finding(s) in round(s) <n, …> and no fix-mutation row for that fix round — record each fix's fix-mutation: line, or one -mutated none -test <reason> row for a fix that changed nothing executable (skills/flow/review-panel-fix-round.md)` and set `violated` (decision `mutation-row-at-panel-close`). Add that violation to `scripts/check-panel-findings-closed.sh`'s header. Then apply the contract's diff:

````diff verified:applied with git apply in task order to this worktree's branch spectre/kan-1039-end-to-end-autonomy-rule before any fix commit; check-verbatim-moves, check-references, check-markdown-integrity green
diff --git a/skills/flow/review-panel-fix-round.md b/skills/flow/review-panel-fix-round.md
index 8ace34f6..4a60f94e 100644
--- a/skills/flow/review-panel-fix-round.md
+++ b/skills/flow/review-panel-fix-round.md
@@ -341,6 +341,7 @@ test exists to catch — and the test confirmed to fail under the flip is that a
 test, not merely any test that happens to trip: the flip is what shows the fix's own test tests
 what it claims. The flip is recorded with the round's other `flow record mutation` rows before the
 round closes.
+`check-panel-findings-closed.sh` refuses the panel's close while a round with a fixed Critical or Important finding has no `flow record mutation` row, the predicate integrate's unfinished-work gate applies, so the omission surfaces while the fix round is still open (kan-1039: it surfaced only at the landing gate).
 
 **Every fixed Critical or Important finding leaves a tracked suite test behind** — the
 `regression-test:` line of the REPORT FILE paragraph (`skills/flow/review-panel.md`), so a
````
  - [x] **Step 3: Verify.** `cd stats && gofmt -l internal/guard && go vet ./internal/guard/ && go test ./internal/guard/ -run '^Test(CheckPanelFindingsClosed|CheckUnfinishedWork)' -count=1`, then `scripts/check-verbatim-moves.sh && scripts/check-references.sh && scripts/check-markdown-integrity.sh` — green.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/panelfindingsclosed.go`, `stats/internal/guard/check_panel_findings_closed_test.go`, `scripts/check-panel-findings-closed.sh`, `skills/flow/review-panel-fix-round.md`
**Tests:** `TestCheckPanelFindingsClosedNeedsAFixMutationRow`
**Regression:** `TestCheckPanelFindingsClosedNeedsAFixMutationRow` — reverting the commit prints FINDINGS-CLOSED over the fixed Important round with no row.
**Baseline:** before=346 after=347
<!-- measured: cat stats/internal/guard/*_test.go | grep -c -E '^func Test' @ branch spectre/kan-1039-end-to-end-autonomy-rule (plus each earlier guard task's one test) -->
**Commit:** fix(guard): refuse the panel's close while a fixed round has no fix-mutation row
**After:** Task 5

- [x] 7. Floor a full-suite fix back-fill at the panel's reviewer ends
  - [x] **Step 1: Write the failing test.** Add `TestRecordDispatchBackfillFloorsAFullSuiteFixAtThePanelsReviewers` to `stats/cmd/flow/record_test.go`, on `TestRecordDispatchBackfillFloorsAPanelFixAtItsRoundsReviewers`' fixture (commits at `2026-10-23T09:19:34Z` and `2026-10-23T09:30:03Z`, the same four reviewer rows): an inline `-role panel-fix -key full-suite-fix-1` back-fill with `-commit` the second commit begins at `2026-10-23T09:28:18Z`, the newest reviewer end not after the commit (`panel-1-primary` ended after it); with a record holding no reviewer row it begins at `2026-10-23T09:19:34Z`; `-key full-suite-fix-x` keeps `2026-10-23T09:19:34Z`. Run `cd stats && make -s web-dist && go test ./cmd/flow/ -run '^TestRecordDispatchBackfillFloors' -count=1`: the new test fails.
  - [x] **Step 2: Fix.** In `stats/cmd/flow/record.go`, `panelFixFloor` takes the fix commit's time `endT` as its last argument, its one caller passing `endT`. For `role == "panel-fix"` and a key `full-suite-fix-<n>`, `<n>` a positive integer, the floor is the latest `EndedAt` among the dispatches `records.PanelDispatchRound` accepts whose `EndedAt` is not after `endT`; the `panel-fix-<r>` case is unchanged (decision `full-suite-fix-floor`). Extend its doc comment and the back-fill comment above its call (kan-1039 self-review: `full-suite-fix-1` claimed the whole panel wait).
  - [x] **Step 3: Repair row 8196.** Read it with `flow record dispatches -change kan-1039-end-to-end-autonomy-rule -C /Users/tweety53/Projects/agents-worktrees/kan-1039-end-to-end-autonomy-rule`: confirm `id` 8196 is `full-suite-fix-1` starting `2026-10-27T16:44:08+02:00`, and take its `commitSha` and `endedAt`. Then, from `stats/`, run `FLOWD_DSN='postgres://flow:flow@localhost:5433/flow?sslmode=disable' go run ./cmd/flowd retime-dispatch 8196 2026-10-27T15:14:22Z <its endedAt, in UTC> <its commitSha>` and `FLOWD_DSN='postgres://flow:flow@localhost:5433/flow?sslmode=disable' go run ./cmd/flowd restamp-dispatch 8196 /Users/tweety53/.claude/projects/-Users-tweety53-Projects-agents/f562af9f-7080-469d-b4aa-ede280788cf8.jsonl`; each exits 0. Re-read the row: it starts at `2026-10-27T17:14:22+02:00`; report its cache-read figure before and after. unverified:confirm restamp-dispatch re-attributes an already-inline row over its new window; if it refuses, report its output and leave the retimed row.
  - [x] **Step 4: Verify.** `cd stats && gofmt -l cmd/flow && make -s web-dist && go vet ./cmd/flow/ && go test ./cmd/flow/ -run '^TestRecordDispatch' -count=1` — green.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordDispatchBackfillFloorsAFullSuiteFixAtThePanelsReviewers`
**Regression:** `TestRecordDispatchBackfillFloorsAFullSuiteFixAtThePanelsReviewers` — reverting the commit begins the back-fill at the commit before the fix.
**Baseline:** before=339 after=340
<!-- measured: cat stats/cmd/flow/*_test.go | grep -c -E '^func Test' @ branch spectre/kan-1039-end-to-end-autonomy-rule -->
**Commit:** fix(flow): floor a full-suite fix back-fill at the panel's reviewer ends
**After:** none

- [x] 8. Record each declared suite's runtime from the verify list
  - [x] **Step 1: Write the failing test.** Add `TestRunVerifyListRecordsSuiteRuntimes` to `stats/internal/guard/run_verify_list_test.go`, on `TestRunVerifyList`'s fixture: a `## test` fence holding `# suite: demo-suite`, `echo one`, then `echo two`, and a fake `flow` on PATH that appends `$*` to a log file and, when its first two words are `suite record`, runs the words after `--`. The list exits 0 and prints both commands' report lines, and the log holds exactly one line, `suite record -C <worktree> -suite demo-suite -- bash -c echo one`; with no `flow` on PATH the list still exits 0. Run `cd stats && go test ./internal/guard/ -run '^TestRunVerifyListRecordsSuiteRuntimes$' -count=1`: it fails.
  - [x] **Step 2: Fix.** In `stats/internal/guard/runverifylist.go`, compile `rvlSuite` beside `rvlInputs` from the pattern `^[[:space:]]*# suite:[[:space:]]*([A-Za-z0-9._-]+)[[:space:]]*$`, and parse a `suites` slice beside `inputs` (pending, then cleared at the command it applies to). In the `test` branch, a command whose suite is set runs as `{flow, "suite", "record", "-C", abs(wt), "-suite", <suite>, "--", "bash", "-c", c}` when `lookPath(env, "flow")` finds one, else as today (decision `suite-line-in-test-list`). Add the `# suite:` line's meaning to the header comment's THE LIST paragraph and to `scripts/run-verify-list.sh`'s header. In `.flow/project.md` `## test`, put `# suite: guard-tests` above `scripts/run-guard-tests.sh`, `# suite: stats-go` above the Go test line and `# suite: stats-spa` above the SPA command's `# inputs:` line; then replace the sentence that opens `Record a run with` with: Each command a `# suite:` line names is recorded so on every `run-verify-list.sh <worktree> test`; record any other run with `flow suite record -suite <name> -- <command>` — a store that cannot be reached costs one warning line and never changes the suite's own exit code. Then apply the contract's diff:

````diff verified:applied with git apply in task order to this worktree's branch spectre/kan-1039-end-to-end-autonomy-rule before any fix commit; check-verbatim-moves, check-references, check-markdown-integrity green
diff --git a/skills/flow-contracts/project-configuration.md b/skills/flow-contracts/project-configuration.md
index 25037f84..7460bf10 100644
--- a/skills/flow-contracts/project-configuration.md
+++ b/skills/flow-contracts/project-configuration.md
@@ -24,7 +24,7 @@ exactly what that row specifies and nothing else. The row is authoritative.
 | `## run` | How to start the stack and individual services locally — the actual commands, including any parameter that must point at another app's root. |
 | `## stop` | Optional. The command that stops the project's local stack, run by bare `/flow` before its stack-stopped check so nothing holds a port or file handle open in a worktree about to be removed. Same shape as `## worktree setup`: one command per line inside a fence, read literally and run in order from the worktree; prose outside a fence is never run. **Absent means that check is skipped, not failed** — cleanup proceeds on the strength of the other two checks. A body with no fenced command declares no command and is skipped the same way. A project with no stack should say so here rather than omit the key, so its absence is a recorded fact rather than an oversight. |
 | `## credentials` | Local-development-only sign-in credentials (which app, user, password) and what seeds them. Never deployed-environment secrets. |
-| `## test` | The command(s) that run the project's tests. Every non-`#` line inside its bash fence runs literally, as `## lint`'s do: a template with a `<word>` placeholder, such as a verify bundle's `:<module>:test` selector, belongs in prose outside the fence, and `<agents repo>/scripts/run-verify-list.sh` refuses the list while one stands inside it. A `# inputs:` line naming pathspecs declares what the one command below it reads, so `run-verify-list.sh` skips that command when nothing it reads changed since the list last passed; a command declaring none reads everything. |
+| `## test` | The command(s) that run the project's tests. Every non-`#` line inside its bash fence runs literally, as `## lint`'s do: a template with a `<word>` placeholder, such as a verify bundle's `:<module>:test` selector, belongs in prose outside the fence, and `<agents repo>/scripts/run-verify-list.sh` refuses the list while one stands inside it. A `# inputs:` line naming pathspecs declares what the one command below it reads, so `run-verify-list.sh` skips that command when nothing it reads changed since the list last passed; a command declaring none reads everything. A `# suite: <name>` line names the one command below it as that suite: `run-verify-list.sh` runs it under `flow suite record -suite <name>`, so its runtime reaches the store `flow suite list` reads. |
 | `## worktree setup` | Optional. A fenced command block run once per worktree, from the worktree root, immediately after the run creates the worktree — **2. Isolate the workspace** (`skills/flow/implement.md`) on `/flow`, the kickoff section (`skills/flow-fast/SKILL.md`) on a `/flow-fast` creating run — and before anything else touches the tree — the place for a build that a gitignored, embedded artifact needs before the project's first `go test`/`go build` can succeed. Absent means nothing runs. Same shape as `## lint` and `## test`: one command per line inside the fence, read literally and run in order; resolved through `project-get.sh <worktree> "worktree setup"`, whose exit 1 is the absent case. |
 | `## lint` | The command(s) that verify lint, and the auto-fix command to run first. |
 | `## standards` | The project's own written standards: the files the principles reviewer receives, plus any opt-in shared rule the project has adopted. This same list is both the opt-in list and the reviewer's standards list. Absent, no standards resolve: the reviewer gets an empty list, never auto-detected files. |
````
  - [x] **Step 3: Verify.** `cd stats && gofmt -l internal/guard && go vet ./internal/guard/ && go test ./internal/guard/ -run '^TestRunVerifyList' -count=1`, then `scripts/check-verbatim-moves.sh && scripts/check-references.sh && scripts/check-markdown-integrity.sh` — green.

**Build:** green
**Files:** `stats/internal/guard/runverifylist.go`, `stats/internal/guard/run_verify_list_test.go`, `scripts/run-verify-list.sh`, `.flow/project.md`, `skills/flow-contracts/project-configuration.md`
**Tests:** `TestRunVerifyListRecordsSuiteRuntimes`
**Regression:** `TestRunVerifyListRecordsSuiteRuntimes` — reverting the commit runs the declared command bare and the log holds no `suite record` line.
**Baseline:** before=347 after=348
<!-- measured: cat stats/internal/guard/*_test.go | grep -c -E '^func Test' @ branch spectre/kan-1039-end-to-end-autonomy-rule (plus each earlier guard task's one test) -->
**Commit:** fix(guard): record each declared suite's runtime from the verify list
**After:** Task 6

- [x] 9. Require the no-cache flag on every mutant's test run
  - [x] **Step 1: Apply the sentence** (decision `mutation-no-cache`):

````diff verified:applied with git apply in task order to this worktree's branch spectre/kan-1039-end-to-end-autonomy-rule before any fix commit; check-verbatim-moves, check-references, check-markdown-integrity green
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index de7dacb9..fe55cced 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -780,6 +780,7 @@ the slot ran 55 minutes, and the parent narrowed the reproducer to one test's na
 `# reproducer-bound:` is that measured run's seconds, doubled. Its exits follow the
 mutation-reproducer convention, never the generic one: it exits 0 when every test passes with the
 mutation in place — the mutant survived, the defect present — and non-zero when a test kills it.
+Every test run of a mutant passes the test runner's no-cache flag — `-count=1` for `go test` — because a cached result never sees a mutated file a test reads through a subprocess (kan-1039: the slot discarded a whole cached sweep whose survivals were false).
 
 ### Principles
 
````
  - [x] **Step 2: Verify.** `scripts/check-verbatim-moves.sh && scripts/check-references.sh && scripts/check-markdown-integrity.sh` — green.

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentence.
**Regression:** none — no test is declared.
**Commit:** docs(flow): require the no-cache flag on every mutant's test run
**After:** Task 4
