# self-review-kan-991-panel-effort-roll-high-gated

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-17-self-review-kan-991-panel-effort-roll-high-gated-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; never write `spectre/changes/`. Never merge or push.
- Each task's change is the unified diff its step gives, applied with `git apply` from this worktree's root, in task order — every diff was applied and checked there (`design.md`, **Context**). A diff that does not apply is reported, never hand-merged by guesswork.
- Every prose line a diff adds or rewords is copied verbatim: this change's `verbatim-moves.txt` acknowledges each such sentence by its exact text, and any other wording fails `scripts/check-verbatim-moves.sh`.
- Run `cd stats && gofmt -w .` before committing any Go file; `gofmt -l .` and `go vet` stay clean.

## Review focus

- A plan step citing a script the plan itself creates or deletes must pass `plan-script-paths` — the `**Files:**` exemption (Task 1).
- `check-plan-shape.sh` with no argument must stay exit 0 on this branch, whose unlanded kan-991 archive still cites the retired guard (Task 1).
- A `verbatim-moves.txt` outside `<spec-root>/changes/` must still refuse the docs-only reduction (Task 2).
- No sentence in `skills/flow/` may still grant the gated reviewer bundle `high` (Task 3).
- A real pass note of two words must still record (Task 4).

- [ ] 1. Refuse a plan step citing a scripts path the tree lacks
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Copy `spectre/changes/archive/kan-991-panel-effort-roll-high-gated/tasks.md` into a scratch `spectre/changes/zz-probe/` directory, run `scripts/check-plan-shape.sh spectre/changes/zz-probe/tasks.md` and confirm it exits 0 though step 6 cites a script `52f9660b` deleted — the defect. Remove the scratch directory.
  - [ ] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at 56e6aaa9 in task order; with all five applied go test ./internal/guard/ ./cmd/flow/ passes; scripts/test-check-plan-shape.sh prints all cases passed
diff --git a/scripts/check-plan-shape.sh b/scripts/check-plan-shape.sh
index ca4e8f22..8e1975fa 100755
--- a/scripts/check-plan-shape.sh
+++ b/scripts/check-plan-shape.sh
@@ -64,8 +64,9 @@ if ! python3 -c 'import sys; sys.exit(0)'; then
   exit 2
 fi
 
-# flow_guard_exec runs `flow-guard plan-suite-steps` on every file below and
-# `flow-guard unlanded-archives` on the no-argument scan; named as
+# flow_guard_exec runs `flow-guard plan-suite-steps` and `flow-guard
+# plan-script-paths` on every file below and `flow-guard unlanded-archives`
+# on the no-argument scan; named as
 # $SCRIPT_DIR/<name> and checked before sourcing, like every sibling here.
 FLOW_GUARD_LIB="$SCRIPT_DIR/lib/flow-guard.sh"
 if [ ! -f "$FLOW_GUARD_LIB" ]; then
@@ -76,15 +77,18 @@ source "$FLOW_GUARD_LIB"
 
 # check_one <tasks.md>: check-plan-shape.py's F1-F11, then `flow-guard
 # plan-suite-steps` (stats/internal/guard/plansuitesteps.go) — a task step
-# running a module or repository test suite. Returns 2 if either cannot
-# answer, else 1 if either found a violation, else 0.
+# running a module or repository test suite — and `flow-guard
+# plan-script-paths` (stats/internal/guard/planscriptpaths.go) — a task step
+# citing a scripts/ path the repository does not carry. Returns 2 if any
+# cannot answer, else 1 if any found a violation, else 0.
 check_one() {
-  local py=0 go=0
+  local py=0 go=0 sp=0
   python3 "$PYTHON_GUARD" "$1" || py=$?
   (flow_guard_exec plan-suite-steps 2 "check-plan-shape.sh:" "$1") || go=$?
-  if [ "$py" -eq 2 ] || [ "$go" -eq 2 ]; then
+  (flow_guard_exec plan-script-paths 2 "check-plan-shape.sh:" "$1") || sp=$?
+  if [ "$py" -eq 2 ] || [ "$go" -eq 2 ] || [ "$sp" -eq 2 ]; then
     return 2
-  elif [ "$py" -ne 0 ] || [ "$go" -ne 0 ]; then
+  elif [ "$py" -ne 0 ] || [ "$go" -ne 0 ] || [ "$sp" -ne 0 ]; then
     return 1
   fi
   return 0
diff --git a/stats/internal/guard/plan_script_paths_test.go b/stats/internal/guard/plan_script_paths_test.go
new file mode 100644
index 00000000..75807ef1
--- /dev/null
+++ b/stats/internal/guard/plan_script_paths_test.go
@@ -0,0 +1,61 @@
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
+// Every case of the plan-script-paths contract: a live plan's unfenced inline
+// span citing a `scripts/…` path the repository lacks is refused, naming the
+// task, the line and the path; a present path, a path a task's **Files:**
+// names, a placeholder, a fenced example and an archived plan pass.
+func TestPlanScriptPaths(t *testing.T) {
+	cases := []struct {
+		name, rel, plan string
+		code            int
+		want            []string
+	}{
+		{"retired guard cited (kan-991 step 6)", "spectre/changes/c/tasks.md",
+			"- [ ] 1. A\n\n**Files:** `a.md`\n\n  - [ ] **Step 6: Verify: `scripts/present.sh && scripts/check-vocabulary.sh`.**\n",
+			1, []string{"tasks.md:5: task 1 cites `scripts/check-vocabulary.sh`"}},
+		{"present path with a line suffix", "spectre/changes/c/tasks.md",
+			"- [ ] 1. A\n\n- see `scripts/present.sh:12` and `bash scripts/present.sh`\n", 0, nil},
+		{"path the plan creates", "spectre/changes/c/tasks.md",
+			"- [ ] 1. A\n\n**Files:** `scripts/new.sh`\n\n- [ ] 2. B\n\n- run `scripts/new.sh`\n", 0, nil},
+		{"placeholder", "spectre/changes/c/tasks.md",
+			"- [ ] 1. A\n\n- every `scripts/check-*.sh` and `scripts/<name>.sh`\n", 0, nil},
+		{"fenced example", "spectre/changes/c/tasks.md",
+			"- [ ] 1. A\n\n```bash\n`scripts/gone.sh`\n```\n", 0, nil},
+		{"archived plan", "spectre/changes/archive/c/tasks.md",
+			"- [ ] 1. A\n\n- `scripts/gone.sh`\n", 0, nil},
+	}
+	for _, tc := range cases {
+		t.Run(tc.name, func(t *testing.T) {
+			dir := t.TempDir()
+			for p, body := range map[string]string{"scripts/present.sh": "", tc.rel: tc.plan} {
+				if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, p)), 0o755); err != nil {
+					t.Fatal(err)
+				}
+				if err := os.WriteFile(filepath.Join(dir, p), []byte(body), 0o644); err != nil {
+					t.Fatal(err)
+				}
+			}
+			var stdout, stderr bytes.Buffer
+			code := Registry["plan-script-paths"]([]string{tc.rel}, Env{Dir: dir}, &stdout, &stderr)
+			if code != tc.code {
+				t.Fatalf("exit %d, want %d; stdout=%q stderr=%q", code, tc.code, stdout.String(), stderr.String())
+			}
+			for _, w := range tc.want {
+				if !strings.Contains(stdout.String(), w) {
+					t.Errorf("stdout %q lacks %q", stdout.String(), w)
+				}
+			}
+			if tc.code == 0 && stdout.Len() != 0 {
+				t.Errorf("clean plan printed %q", stdout.String())
+			}
+		})
+	}
+}
diff --git a/stats/internal/guard/planscriptpaths.go b/stats/internal/guard/planscriptpaths.go
new file mode 100644
index 00000000..8ff9b8bf
--- /dev/null
+++ b/stats/internal/guard/planscriptpaths.go
@@ -0,0 +1,99 @@
+package guard
+
+import (
+	"fmt"
+	"io"
+	"os"
+	"path/filepath"
+	"regexp"
+	"strings"
+)
+
+// planScriptPathsCmd is `flow-guard plan-script-paths <tasks.md>`, run by
+// check-plan-shape.sh beside plan-suite-steps on every tasks.md it scans. It
+// refuses a task body citing, in an inline code span outside a fence, a
+// `scripts/…` path the repository does not carry — a step that names a
+// retired guard runs nothing (kan-991's step 6 cited a check-vocabulary.sh
+// 52f9660b had deleted). The repository is the one whose
+// `<spec-root>/changes/<name>/tasks.md` this is; a path any task's
+// `**Files:**` names is the plan's own to create or delete and passes; a
+// token carrying `<`, `*`, `{` or `$` is a placeholder, not a path. An
+// archived plan, or a file outside a `changes/<name>/` directory, is not
+// checked: its citations are history, judged when it was live. Exit 0 clean,
+// 1 one `file:line: task <id> cites … ` line per missing path, 2 a usage
+// error or a file that cannot be read.
+func init() { Registry["plan-script-paths"] = planScriptPathsCmd }
+
+var pspTokenRE = regexp.MustCompile(`(^|[\s(=])(scripts/[^\s` + "`" + `]+)`)
+
+func planScriptPathsCmd(args []string, env Env, stdout, stderr io.Writer) int {
+	if len(args) != 1 || args[0] == "" {
+		fmt.Fprintln(stderr, "usage: flow-guard plan-script-paths <tasks.md>")
+		return 2
+	}
+	path := args[0]
+	if !filepath.IsAbs(path) {
+		path = filepath.Join(env.Dir, path)
+	}
+	b, err := os.ReadFile(path)
+	if err != nil {
+		fmt.Fprintf(stderr, "%s: cannot read file: %v\n", args[0], err)
+		return 2
+	}
+	lines, err := tcfPlanLines(b)
+	if err != nil {
+		fmt.Fprintf(stderr, "%s: cannot read file: %v\n", args[0], err)
+		return 2
+	}
+	changes := filepath.Dir(filepath.Dir(path))
+	if filepath.Base(changes) != "changes" {
+		return 0
+	}
+	root := filepath.Dir(filepath.Dir(changes))
+	tasks := tcfIterTasks(lines)
+	declared := map[string]bool{}
+	for _, task := range tasks {
+		if t, ok := tcfParseTask(lines, task.id); ok {
+			for _, f := range t.files {
+				declared[f] = true
+			}
+		}
+	}
+	found := 0
+	for _, task := range tasks {
+		var text strings.Builder
+		var starts, offsets []int
+		tcfUnfenced(task.lines, func(offset int, line string) bool {
+			starts = append(starts, text.Len())
+			offsets = append(offsets, offset)
+			text.WriteString(line + "\n")
+			return false
+		})
+		joined := text.String()
+		for _, m := range pssSpanRE.FindAllStringSubmatchIndex(joined, -1) {
+			for _, tm := range pspTokenRE.FindAllStringSubmatch(joined[m[2]:m[3]], -1) {
+				p := strings.TrimRight(tm[2], ".,;:)'\"")
+				if i := strings.IndexByte(p, ':'); i >= 0 {
+					p = p[:i]
+				}
+				if strings.ContainsAny(p, "<*{$") || declared[p] {
+					continue
+				}
+				if _, err := os.Lstat(filepath.Join(root, p)); err == nil {
+					continue
+				}
+				i := 0
+				for i+1 < len(starts) && starts[i+1] <= m[0] {
+					i++
+				}
+				fmt.Fprintf(stdout, "%s:%d: task %s cites `%s`, which the repository does not carry — name a path that exists, or list it in a task's **Files:** when the plan creates it\n",
+					args[0], task.taskLine+1+offsets[i], task.id, p)
+				found++
+			}
+		}
+	}
+	if found > 0 {
+		return 1
+	}
+	return 0
+}
````

  - [ ] **Step 3: Verify.** `cd stats && go test -run 'TestPlanScriptPaths' ./internal/guard/` passes; re-run Step 1's probe — it now exits 1 naming task 1 and the retired script — and remove the scratch directory; `scripts/check-plan-shape.sh` (no argument) exits 0; `bash scripts/test-check-plan-shape.sh` prints `all cases passed`; `scripts/check-guard-symlinks.sh` exits 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/planscriptpaths.go`, `stats/internal/guard/plan_script_paths_test.go`, `scripts/check-plan-shape.sh`
**Tests:** `TestPlanScriptPaths`
**Regression:** deleting `planscriptpaths.go` leaves `Registry["plan-script-paths"]` nil, so `TestPlanScriptPaths` fails and a step citing a retired script passes `check-plan-shape.sh` again.
**Baseline:** before=204 after=205
<!-- measured: cat stats/internal/guard/*_test.go | grep -c '^func Test' @ 56e6aaa9 (before; after with this task's diff applied) -->
**Commit:** fix(guard): refuse a plan step citing a scripts path the tree lacks
**After:** none

**Decision:** plan-script-paths-go-sibling

- [ ] 2. Count a change's verbatim-moves.txt as documentation for the docs-only reduction
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Apply the test hunks alone (`git apply --include='stats/internal/guard/check_panel_docs_only_test.go' <the diff in step 2>`) and run `cd stats && go test -run 'TestCheckPanelDocsOnly' ./internal/guard/`: the `committed .md plus a change's verbatim-moves.txt -> exit 0 (kan-991)` case fails — the defect.
  - [ ] **Step 2: Implement the fix.** Apply the rest of this diff (`git apply --exclude='stats/internal/guard/check_panel_docs_only_test.go' <the diff>`):

````diff verified:git apply at 56e6aaa9 in task order; with all five applied go test ./internal/guard/ ./cmd/flow/ passes; with paneldocsonly.go reverted the kan-991 case fails
diff --git a/scripts/check-panel-docs-only.sh b/scripts/check-panel-docs-only.sh
index eed5bc76..bdeb2864 100755
--- a/scripts/check-panel-docs-only.sh
+++ b/scripts/check-panel-docs-only.sh
@@ -5,8 +5,10 @@
 # Usage: check-panel-docs-only.sh <worktree> <merge-base>
 #
 # Exit codes:
-#   0  every touched path ends .md or .mdc
-#   1  at least one touched path does not — that first path is printed to
+#   0  every touched path ends .md or .mdc, or is a change's
+#      verbatim-moves.txt under <spec-root>/changes/ — the acknowledgement
+#      list check-verbatim-moves.sh reads, documentation though not Markdown
+#   1  at least one touched path is neither — that first path is printed to
 #      stdout — including an empty touched-path set
 #   2  cannot answer: a missing argument, <worktree> not a directory or not
 #      a git worktree, <merge-base> not resolving, or a git invocation
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index 4642b698..46e1898b 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -255,7 +255,7 @@ below): one `dispatches` row with `-slot test-sanity+call-efficiency` and the pa
 
 **Exit 0 from
 every worktree in the resolved set — every path this branch touched, committed since the merge
-base, staged or unstaged, ends `.md` or `.mdc` — reduces pass 1 to `primary` alone**, plus **The always-on extra dispatch** and every
+base, staged or unstaged, ends `.md` or `.mdc` or is a change's `<project>/<spec-root>/changes/<change>/verbatim-moves.txt` — reduces pass 1 to `primary` alone**, plus **The always-on extra dispatch** and every
 slot the operator's per-run instruction
 named at this stage's start. Every other resolved slot is recorded with
 `flow record pass -round 0 -note 'not dispatched — docs-only reduction: <slot>'`.
diff --git a/stats/internal/guard/check_panel_docs_only_test.go b/stats/internal/guard/check_panel_docs_only_test.go
index 39518656..eef38232 100644
--- a/stats/internal/guard/check_panel_docs_only_test.go
+++ b/stats/internal/guard/check_panel_docs_only_test.go
@@ -128,6 +128,19 @@ func TestCheckPanelDocsOnly(t *testing.T) {
 		{"committed .md plus staged main.go -> stdout is main.go", func(t *testing.T) { is(t, case4(t), 1, "main.go") }},
 		{"committed .md plus unstaged edit to base.go -> exit 1", func(t *testing.T) { rcIs(t, case5(t), 1) }},
 		{"committed .md plus unstaged edit to base.go -> stdout is base.go", func(t *testing.T) { is(t, case5(t), 1, "base.go") }},
+		{"committed .md plus a change's verbatim-moves.txt -> exit 0 (kan-991)", func(t *testing.T) {
+			fx := newFx(t)
+			if err := os.MkdirAll(fx.repo+"/spectre/changes/c", 0o755); err != nil {
+				t.Fatal(err)
+			}
+			commit(fx, "notes.md", "spectre/changes/c/verbatim-moves.txt")
+			is(t, run(t, fx, fx.repo, mb), 0, "")
+		}},
+		{"verbatim-moves.txt outside <spec-root>/changes/ -> exit 1", func(t *testing.T) {
+			fx := newFx(t)
+			commit(fx, "notes.md", "verbatim-moves.txt")
+			is(t, run(t, fx, fx.repo, mb), 1, "verbatim-moves.txt")
+		}},
 		{"staged-only .md change -> exit 0", func(t *testing.T) {
 			fx := newFx(t)
 			stage(fx, "staged.md")
diff --git a/stats/internal/guard/paneldocsonly.go b/stats/internal/guard/paneldocsonly.go
index 02a416b6..1c4c631b 100644
--- a/stats/internal/guard/paneldocsonly.go
+++ b/stats/internal/guard/paneldocsonly.go
@@ -3,11 +3,14 @@ package guard
 import (
 	"fmt"
 	"io"
+	"path"
+	"strings"
 )
 
 // checkPanelDocsOnly is scripts/check-panel-docs-only.sh: that script's
 // header comment is the contract -- exit 0 when every path this change
-// touched ends .md or .mdc, 1 with the first path that does not printed to
+// touched ends .md or .mdc or is a change's verbatim-moves.txt under
+// <spec-root>/changes/, 1 with the first path that is neither printed to
 // stdout (an empty touched-path set is 1 with nothing printed), 2 on a usage
 // error or a git invocation that failed. Resolving git, validating the
 // arguments and collecting the paths -- the union of committed-since-merge-
@@ -36,8 +39,12 @@ func checkPanelDocsOnly(args []string, env Env, stdout, stderr io.Writer) int {
 	if !ok {
 		return 2
 	}
+	changes := specRootLeaf(smcAbs(env, worktree), stderr) + "/changes/"
 	for _, p := range paths {
-		if !isDocPath(p) {
+		// A verbatim-moves.txt is check-verbatim-moves.sh's acknowledgement
+		// list of reworded prose: documentation, though not Markdown.
+		// isDocPath stays `.mdc?$`, shared with plan-class.
+		if !isDocPath(p) && !(strings.HasPrefix(p, changes) && path.Base(p) == "verbatim-moves.txt") {
 			fmt.Fprintln(stdout, p)
 			return 1
 		}
````

  - [ ] **Step 3: Verify.** `cd stats && go test -run 'TestCheckPanelDocsOnly' ./internal/guard/` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py` and `scripts/check-installed-citations.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/paneldocsonly.go`, `stats/internal/guard/check_panel_docs_only_test.go`, `scripts/check-panel-docs-only.sh`, `skills/flow/review-panel.md`
**Tests:** `committed .md plus a change's verbatim-moves.txt -> exit 0 (kan-991)`, `verbatim-moves.txt outside <spec-root>/changes/ -> exit 1`
**Regression:** reverting `paneldocsonly.go` makes the guard print `spectre/changes/c/verbatim-moves.txt` and exit 1, so the kan-991 case fails.
**Baseline:** before=25 after=27
<!-- measured: grep -c '^[[:space:]][[:space:]]{"' stats/internal/guard/check_panel_docs_only_test.go @ 56e6aaa9 (before; after with this task's diff applied) -->
**Commit:** fix(guard): count a change's verbatim-moves.txt as documentation for the docs-only reduction
**After:** none

**Decision:** docs-only-verbatim-moves

- [ ] 3. Keep the gated reviewer bundle off high
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -n 'high. only as' skills/flow/brainstorm-planner.md skills/flow/implement.md` and confirm the bundle bullet and the bundle sentence both grant `high` — the defect: no decision pair records the bundle.
  - [ ] **Step 2: Apply the edit.** Apply this diff:

````diff verified:git apply at 56e6aaa9; with it applied check-references, check-markdown-integrity, check-dispatch-paragraphs, check-installed-citations and check-ask-silence exit 0 and check-verbatim-moves reports 0 violations
diff --git a/skills/flow/SKILL-rationale.md b/skills/flow/SKILL-rationale.md
index 674cd7f3..d6a2a269 100644
--- a/skills/flow/SKILL-rationale.md
+++ b/skills/flow/SKILL-rationale.md
@@ -996,7 +996,9 @@ explicit **Yes** to a per-dispatch question added to the plan gate's AskUserQues
 fix run whose instructions report a lower-effort attempt already failed. **Rejected — a third
 option on the gate question:** the gate has exactly two options by contract. **Rejected —
 enforcing in `flow record decision`:** a Go change the ticket did not ask for; prose suffices for
-a planner-written record.
+a planner-written record. **gated-bundle-never-high** (kan-991's self-review): the gated per-task
+reviewer bundle runs at the panel roll's effort and never `high` — the decision JSON records no
+pair for it, so a raised effort had nowhere to be recorded and nothing that dispatches it reads one.
 
 ## brainstorm-planner.md — Decide, step 6 visual verification (visual-verify-decided-at-decide)
 
diff --git a/skills/flow/brainstorm-planner.md b/skills/flow/brainstorm-planner.md
index 5106b1eb..08ac25f5 100644
--- a/skills/flow/brainstorm-planner.md
+++ b/skills/flow/brainstorm-planner.md
@@ -566,7 +566,7 @@ assigns — the implementer, the fixer, each panel dispatch, the rerun pair, eac
   performance-sensitive code — `opus`, always;
 - **every other pass-1 panel dispatch** — `opus`, always;
 - **the extra dispatch** — its rolled pair, `sonnet` or `opus` at `low`: the one pass-1 panel dispatch that may run on `sonnet` (decision `extra-dispatch-sonnet-carve-out`, kan-959);
-- **every gated per-task reviewer bundle** — `opus`, always, at the panel roll's effort, `high` only as the paragraph below allows; it is a first-pass review too (`skills/flow/implement.md`);
+- **every gated per-task reviewer bundle** — `opus`, always, at the panel roll's effort, never `high`, since no decision pair records it; it is a first-pass review too (`skills/flow/implement.md`);
 - **the rerun pair** — `opus` or `sonnet`; nothing requires it to differ from a pass-1 dispatch.
 
 A micro decision records no pair, and a dispatch with no recorded pair — a micro panel,
@@ -583,7 +583,7 @@ record decision` refuses a decision whose implementer, fixer or group pair names
 `effort_roll < 90`, the implementer, the fixer and every implementer group take `low` — `medium` where
 the pair carries a hard seam.** **Every pass-1 panel dispatch and every gated per-task reviewer
 bundle takes `low` when `effort_roll < 60` and `medium` otherwise, whatever the band** — 60/40 per
-change. **`high` reaches one of them only two ways:** the operator's explicit **Yes** at the plan
+change. **`high` reaches a pass-1 panel dispatch — a `panel.dispatches` entry, never a gated bundle — only two ways:** the operator's explicit **Yes** at the plan
 gate (**Plan review gate** below), or a fix run's re-Decide when the fix instructions report that
 an earlier, lower-effort attempt failed to solve the same issue, the pair's `reason` naming that
 attempt.
@@ -657,7 +657,7 @@ Then the question — `/flow-plan`'s wording is **Push artifacts?**, `/flow`'s i
 implementation?** — with exactly two options: **Yes** *(recommended)* and **No (plan needs
 updates)**.
 
-**A dispatch the planner judges worth `high`** — a hard seam the roll's effort may not carry —
+**A `panel.dispatches` entry the planner judges worth `high`** — a hard seam the roll's effort may not carry —
 adds one question per such dispatch to the same AskUserQuestion call: **Raise `<dispatch>` to
 `high` — `<seam>`?**, **No** *(default, recommended)* / **Yes**. Only an explicit **Yes**
 re-records that pair at `high` through the `flow.decide` record sequence before implementation;
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index d6103d0d..f838747e 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -789,7 +789,7 @@ dispatches: every gate-fired task goes to the whole-branch panel as review focus
 (the GATE FOCUS paragraph of **Bundled dispatch**, `skills/flow/review-panel.md`).
 **The bundle is the implementer group**: at a boundary, every task of group N+1 whose gate fired
 goes out in one reviewer Agent call beside group N+2's implementer, on `opus` at the panel roll's
-effort, `high` only as **Model and effort** allows — a first-pass review, per **Model and effort**
+effort, never `high` — a first-pass review, per **Model and effort**
 (`skills/flow/brainstorm-planner.md`). On harnesses `zcode` and `muse`, **Harness mapping**
 (`skills/flow-contracts/model-policy.md`) replaces it, as it does every pair. Each task inside the bundle keeps its own pass: its own
 commit-range diff `git diff <task-sha>^..<task-sha>` — a real commit diff, never a snapshot of
````

  - [ ] **Step 3: Verify.** `grep -n 'high. only as' skills/flow/brainstorm-planner.md skills/flow/implement.md` prints nothing; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh` and `scripts/check-installed-citations.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`, `skills/flow/implement.md`, `skills/flow/SKILL-rationale.md`
**Tests:** **none** — prose; `check-verbatim-moves.sh` pins the reworded sentences.
**Regression:** reverting the commit lets a planner raise the gated bundle to `high` again with nowhere to record it.
**Commit:** docs(flow): keep the gated reviewer bundle off high
**After:** none

**Decision:** gated-bundle-never-high

- [ ] 4. Refuse a one-word pass note
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Apply the test hunks alone (`git apply --include='stats/cmd/flow/record_test.go' <the diff in step 2>`) and run `cd stats && go test -run 'TestRecordPassRefusesOneWordNote' ./cmd/flow/`: both subtests fail, the note recorded — the defect kan-991's parent hit with a `-note "x"` loop (`design.md`, **Context**).
  - [ ] **Step 2: Implement the fix.** Apply the rest of this diff (`git apply --exclude='stats/cmd/flow/record_test.go' <the diff>`):

````diff verified:git apply at 56e6aaa9 in task order; with all five applied go test ./internal/guard/ ./cmd/flow/ passes; with record.go reverted TestRecordPassRefusesOneWordNote fails
diff --git a/stats/cmd/flow/record.go b/stats/cmd/flow/record.go
index a1bcfa2e..e1f15d8d 100644
--- a/stats/cmd/flow/record.go
+++ b/stats/cmd/flow/record.go
@@ -2846,6 +2846,13 @@ func runRecordPass(ctx context.Context, args []string, stdout, stderr io.Writer)
 	if !requireRecordFlags(stderr, [2]string{"-note", *note}) {
 		return 2
 	}
+	// A one-word note is a placeholder, never a pass-log line: kan-991's
+	// parent looped `-note "x"` with its output silenced and landed two
+	// stray rows; no pass note in the store's history was one word and real.
+	if len(strings.Fields(*note)) < 2 {
+		fmt.Fprintf(stderr, "flow: record pass: -note %q is one word — a pass-log line says what happened, in two words at least\n", *note)
+		return 2
+	}
 	// A note opening with a marker label is a marker line in the wrong
 	// place: the renderer neutralises its colon, so the round would lose
 	// the structural row the line meant to be.
diff --git a/stats/cmd/flow/record_test.go b/stats/cmd/flow/record_test.go
index 32da1db3..ffbba3fe 100644
--- a/stats/cmd/flow/record_test.go
+++ b/stats/cmd/flow/record_test.go
@@ -4862,6 +4862,35 @@ func TestRecordPassRefusesMarkerNote(t *testing.T) {
 	}
 }
 
+// TestRecordPassRefusesOneWordNote pins the kan-991 self-review fix: a
+// one-word pass note is a placeholder -- kan-991's parent looped `-note "x"`
+// and landed two stray rows -- and `flow record pass` refuses it before the
+// store is contacted. The -addr is a closed port, so a regressed refusal
+// falls through to the journal instead of reaching a live daemon.
+func TestRecordPassRefusesOneWordNote(t *testing.T) {
+	for _, note := range []string{"x", "  probe  "} {
+		t.Run(note, func(t *testing.T) {
+			repo := recordRepo(t, "kan-258")
+			isolatedStateRoot(t)
+
+			var stdout, stderr bytes.Buffer
+			code := run(context.Background(),
+				[]string{"record", "pass", "-addr", "http://127.0.0.1:1", "-timeout", "500ms", "-C", repo,
+					"-change", "kan-258", "-round", "0", "-note", note},
+				strings.NewReader(""), &stdout, &stderr)
+			if code != 2 {
+				t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
+			}
+			if !strings.Contains(stderr.String(), "is one word") {
+				t.Errorf("stderr = %q, want it to say the note is one word", stderr.String())
+			}
+			if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
+				t.Errorf("a refused one-word note left a record journal behind")
+			}
+		})
+	}
+}
+
 // TestRecordWriteRefusesAChangeNoCheckoutCarries pins the project-key
 // sanity check: a run-record write for a change no checkout of the -C
 // repository carries is refused before the store is contacted, so it never
````

  - [ ] **Step 3: Verify.** `cd stats && go test -run 'TestRecordPass' ./cmd/flow/` passes; `cd stats && go vet ./cmd/flow && gofmt -l .` is clean. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordPassRefusesOneWordNote`
**Regression:** reverting `record.go` lets `-note x` reach the journal fallback, so `TestRecordPassRefusesOneWordNote` fails on exit code and journal.
**Baseline:** before=114 after=115
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ 56e6aaa9 (before; after with this task's diff applied) -->
**Commit:** fix(stats): refuse a one-word pass note
**After:** none

**Decision:** pass-note-two-words

- [ ] 5. List the rationale appendix a changed rule's reasoning lives in
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -n 'rationale' skills/flow/brainstorm-planner.md` and confirm section D carries no directive to list a `*-rationale.md` appendix in `**Files:**` — the defect kan-991's plan hit with `skills/flow/SKILL-rationale.md`.
  - [ ] **Step 2: Apply the edit.** On top of Task 3's commit, apply this diff:

````diff verified:git apply at 56e6aaa9 after Task 3's diff; check-references, check-markdown-integrity, check-dispatch-paragraphs, check-installed-citations and check-ask-silence exit 0 and check-verbatim-moves reports 0 violations
diff --git a/skills/flow/brainstorm-planner.md b/skills/flow/brainstorm-planner.md
index 08ac25f5..2d33672b 100644
--- a/skills/flow/brainstorm-planner.md
+++ b/skills/flow/brainstorm-planner.md
@@ -324,6 +324,10 @@ paragraph under **The build-green tag** (`skills/flow-contracts/build-green.md`)
 > code change, while an edit made before reproducing destroys the evidence the task set out to
 > gather.
 
+> **List the rationale appendix a changed rule's reasoning lives in.** A step that changes a rule
+> whose reasoning a `*-rationale.md` appendix records lists that appendix in its task's
+> `**Files:**` and updates it.
+
 **Load `skills/flow-contracts/plan-provenance.md`.** While enriching `tasks.md`, tag every fenced
 block, every numeric claim and every assumption the plan cannot verify at plan time per **Plan
 provenance** (`skills/flow-contracts/plan-provenance.md`).
````

  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py` and `scripts/check-ask-silence.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`
**Tests:** **none** — a planning directive; `check-verbatim-moves.sh` pins the added sentences.
**Regression:** reverting the commit removes the directive, so a plan changing a rule can leave its rationale appendix stale again.
**Commit:** docs(flow): list the rationale appendix a changed rule's reasoning lives in
**After:** Task 3

**Decision:** rationale-appendix-in-files
