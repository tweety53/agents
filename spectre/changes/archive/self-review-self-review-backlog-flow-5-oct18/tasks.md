# self-review-self-review-backlog-flow-5-oct18

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** `check-plan-shape.sh` refuses a live `/flow` plan stating task commits carry no `Task-Id:`
trailer; a loop plan stating it still passes.

**Architecture:** one new `flow-guard plan-task-id-rule` subcommand, run by `check-plan-shape.sh`'s
`check_one` beside `plan-script-paths`; a plan is a `/flow` plan when its change has a
`spectre/<name>` branch, local or on origin.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-18-self-review-self-review-backlog-flow-5-oct18-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-5-oct18`, branch `spectre/self-review-backlog-flow-5-oct18`; never touch a main checkout.
- The task's diff is the unified diff its Step 2 gives, applied with `git apply` from the worktree's root — it was applied and checked there at `045471f7` (`design.md`, **Context**). Save it by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`.

## Review Focus

- A loop plan (`self-review-<name>`, `pipeline-fix-<name>-<k>`) stating the no-trailer rule still passes `check-plan-shape.sh` (case "loop plan, no spectre branch").
- `scripts/test-check-plan-shape.sh`'s fixture trees, which are not git repositories, stay clean (case "not a git repository").
- A plan that requires the trailer, or quotes the rule inside a fence, is not refused.

- [x] 1. Refuse a `/flow` plan stating task commits carry no `Task-Id:` trailer
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In this worktree run
    `mkdir -p spectre/changes/self-review-backlog-flow-5-oct18 && cp spectre/changes/archive/self-review-backlog-flow-5-oct18/tasks.md spectre/changes/self-review-backlog-flow-5-oct18/ && scripts/check-plan-shape.sh spectre/changes/self-review-backlog-flow-5-oct18/tasks.md; echo $?`:
    it prints `0` — the offending plan passes. Keep the copy for Step 3. Save Step 2's diff, apply
    its test hunk alone (`git apply --include='stats/internal/guard/plan_task_id_rule_test.go' <the saved diff>`)
    and run `cd stats && make -s web-dist && go test ./internal/guard -run 'TestPlanTaskIDRule' -count=1`:
    it fails with `panic: runtime error: invalid memory address or nil pointer dereference` — no
    `plan-task-id-rule` is registered.
    <!-- measured: both commands at 045471f7 → exit 0; then --- FAIL: TestPlanTaskIDRule with the nil-pointer panic -->
  - [x] **Step 2: Apply the fix.** Apply the rest
    (`git apply --exclude='stats/internal/guard/plan_task_id_rule_test.go' <the saved diff>`):

````diff verified:applied with git apply to this worktree at 045471f7; TestPlanTaskIDRule passes all seven cases, and fails "loop plan, no spectre branch" alone with the branch test weakened to !ok && out != "x"; go vet and gofmt -l clean; scripts/test-check-plan-shape.sh and the no-argument scripts/check-plan-shape.sh exit 0
diff --git a/scripts/check-plan-shape.sh b/scripts/check-plan-shape.sh
index 8e1975fa..aba03ecb 100755
--- a/scripts/check-plan-shape.sh
+++ b/scripts/check-plan-shape.sh
@@ -64,8 +64,8 @@ if ! python3 -c 'import sys; sys.exit(0)'; then
   exit 2
 fi
 
-# flow_guard_exec runs `flow-guard plan-suite-steps` and `flow-guard
-# plan-script-paths` on every file below and `flow-guard unlanded-archives`
+# flow_guard_exec runs `flow-guard plan-suite-steps`, `flow-guard
+# plan-script-paths` and `flow-guard plan-task-id-rule` on every file below and `flow-guard unlanded-archives`
 # on the no-argument scan; named as
 # $SCRIPT_DIR/<name> and checked before sourcing, like every sibling here.
 FLOW_GUARD_LIB="$SCRIPT_DIR/lib/flow-guard.sh"
@@ -77,18 +77,21 @@ source "$FLOW_GUARD_LIB"
 
 # check_one <tasks.md>: check-plan-shape.py's F1-F11, then `flow-guard
 # plan-suite-steps` (stats/internal/guard/plansuitesteps.go) — a task step
-# running a module or repository test suite — and `flow-guard
+# running a module or repository test suite — `flow-guard
 # plan-script-paths` (stats/internal/guard/planscriptpaths.go) — a task step
-# citing a scripts/ path the repository does not carry. Returns 2 if any
+# citing a scripts/ path the repository does not carry — and `flow-guard
+# plan-task-id-rule` (stats/internal/guard/plantaskidrule.go) — a /flow plan
+# stating task commits carry no Task-Id: trailer. Returns 2 if any
 # cannot answer, else 1 if any found a violation, else 0.
 check_one() {
-  local py=0 go=0 sp=0
+  local py=0 go=0 sp=0 ti=0
   python3 "$PYTHON_GUARD" "$1" || py=$?
   (flow_guard_exec plan-suite-steps 2 "check-plan-shape.sh:" "$1") || go=$?
   (flow_guard_exec plan-script-paths 2 "check-plan-shape.sh:" "$1") || sp=$?
-  if [ "$py" -eq 2 ] || [ "$go" -eq 2 ] || [ "$sp" -eq 2 ]; then
+  (flow_guard_exec plan-task-id-rule 2 "check-plan-shape.sh:" "$1") || ti=$?
+  if [ "$py" -eq 2 ] || [ "$go" -eq 2 ] || [ "$sp" -eq 2 ] || [ "$ti" -eq 2 ]; then
     return 2
-  elif [ "$py" -ne 0 ] || [ "$go" -ne 0 ] || [ "$sp" -ne 0 ]; then
+  elif [ "$py" -ne 0 ] || [ "$go" -ne 0 ] || [ "$sp" -ne 0 ] || [ "$ti" -ne 0 ]; then
     return 1
   fi
   return 0
diff --git a/stats/internal/guard/plan_task_id_rule_test.go b/stats/internal/guard/plan_task_id_rule_test.go
new file mode 100644
index 00000000..af016b83
--- /dev/null
+++ b/stats/internal/guard/plan_task_id_rule_test.go
@@ -0,0 +1,58 @@
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
+// Every case of the plan-task-id-rule contract: a live /flow plan — its change
+// has a spectre/<name> branch, local or on origin — stating task commits carry
+// no `Task-Id:` trailer is refused, naming the line; a loop plan (no such
+// branch), a fenced example, a plan requiring the trailer, an archived plan
+// and a directory git cannot read pass.
+func TestPlanTaskIDRule(t *testing.T) {
+	const rule = "## Global constraints\n\n- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer.\n\n- [ ] 1. A\n\n**Files:** `a.md`\n"
+	cases := []struct {
+		name, rel, plan, ref string
+		git                  bool
+		code                 int
+		want                 string
+	}{
+		{"/flow plan, local branch (self-review-backlog-flow-5-oct18)", "spectre/changes/c/tasks.md", rule, "refs/heads/spectre/c", true,
+			1, "tasks.md:3: a /flow plan states task commits carry no `Task-Id:` trailer"},
+		{"/flow plan, origin branch", "spectre/changes/c/tasks.md", rule, "refs/remotes/origin/spectre/c", true, 1, "tasks.md:3:"},
+		{"loop plan, no spectre branch", "spectre/changes/c/tasks.md", rule, "refs/heads/spectre/other", true, 0, ""},
+		{"fenced example", "spectre/changes/c/tasks.md", "```\nno `Task-Id:` trailer\n```\n", "refs/heads/spectre/c", true, 0, ""},
+		{"trailer required", "spectre/changes/c/tasks.md", "- each commit carries a `Task-Id:` trailer\n", "refs/heads/spectre/c", true, 0, ""},
+		{"archived plan", "spectre/changes/archive/c/tasks.md", rule, "refs/heads/spectre/c", true, 0, ""},
+		{"not a git repository", "spectre/changes/c/tasks.md", rule, "", false, 0, ""},
+	}
+	for _, tc := range cases {
+		t.Run(tc.name, func(t *testing.T) {
+			dir := t.TempDir()
+			if tc.git {
+				gitRun(t, dir, "init", "-q")
+				gitRun(t, dir, "commit", "-q", "--allow-empty", "-m", "init")
+				gitRun(t, dir, "update-ref", tc.ref, "HEAD")
+			}
+			p := filepath.Join(dir, tc.rel)
+			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
+				t.Fatal(err)
+			}
+			if err := os.WriteFile(p, []byte(tc.plan), 0o644); err != nil {
+				t.Fatal(err)
+			}
+			var stdout, stderr bytes.Buffer
+			code := Registry["plan-task-id-rule"]([]string{tc.rel}, Env{Dir: dir}, &stdout, &stderr)
+			if code != tc.code {
+				t.Fatalf("exit %d, want %d; stdout=%q stderr=%q", code, tc.code, stdout.String(), stderr.String())
+			}
+			if !strings.Contains(stdout.String(), tc.want) || (tc.code == 0 && stdout.Len() != 0) {
+				t.Errorf("stdout %q, want %q", stdout.String(), tc.want)
+			}
+		})
+	}
+}
diff --git a/stats/internal/guard/plantaskidrule.go b/stats/internal/guard/plantaskidrule.go
new file mode 100644
index 00000000..cbffead7
--- /dev/null
+++ b/stats/internal/guard/plantaskidrule.go
@@ -0,0 +1,72 @@
+package guard
+
+import (
+	"fmt"
+	"io"
+	"os"
+	"path/filepath"
+	"regexp"
+)
+
+// planTaskIDRuleCmd is `flow-guard plan-task-id-rule <tasks.md>`, run by
+// check-plan-shape.sh beside plan-script-paths on every tasks.md it scans. It
+// refuses a `/flow` plan stating, on any line outside a fence, that task
+// commits carry no `Task-Id:` trailer: close-task.sh refuses a `/flow` task
+// commit without one, so an implementer told otherwise meets a contradiction
+// (self-review-backlog-flow-5-oct18's Global constraints did). The rule is a
+// fix loop's own (skills/flow-self-review/SKILL.md, step 3 item 1), so a loop
+// plan carries it legitimately. A `/flow` plan is one whose change <name> has
+// a spectre/<name> branch, local or on origin, in its repository — the test
+// hooks/enforce-loop-dispatch.py applies, since a `/flow` change may itself be
+// named self-review-…, while a loop's branch is never spectre/<its own name>.
+// A repository git cannot read holds no such branch. An archived plan, or a
+// file outside a `changes/<name>/` directory, is not checked: its text is
+// history, judged when it was live. Exit 0 clean, 1 one `file:line:` line per
+// statement, 2 a usage error or a file that cannot be read.
+func init() { Registry["plan-task-id-rule"] = planTaskIDRuleCmd }
+
+var ptirRE = regexp.MustCompile("(?i)\\bno\\s+`Task-Id:`")
+
+func planTaskIDRuleCmd(args []string, env Env, stdout, stderr io.Writer) int {
+	if len(args) != 1 || args[0] == "" {
+		fmt.Fprintln(stderr, "usage: flow-guard plan-task-id-rule <tasks.md>")
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
+	name := filepath.Base(filepath.Dir(path))
+	root := filepath.Dir(filepath.Dir(changes))
+	if out, ok := gitOut(env, "-C", root, "for-each-ref", "--count=1", "--format=x",
+		"refs/heads/spectre/"+name, "refs/remotes/origin/spectre/"+name); !ok || out != "x" {
+		return 0
+	}
+	found := 0
+	tcfUnfenced(lines, func(i int, line string) bool {
+		if ptirRE.MatchString(line) {
+			fmt.Fprintf(stdout, "%s:%d: a /flow plan states task commits carry no `Task-Id:` trailer — close-task.sh refuses a /flow task commit without one; that rule is a fix loop's alone\n",
+				args[0], i+1)
+			found++
+		}
+		return false
+	})
+	if found > 0 {
+		return 1
+	}
+	return 0
+}
````

  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./internal/guard -run 'TestPlanTaskIDRule' -count=1`
    passes; `cd stats && go vet ./internal/guard && gofmt -l internal/guard` prints nothing;
    `scripts/check-plan-shape.sh spectre/changes/self-review-backlog-flow-5-oct18/tasks.md` exits 1
    naming line 16; then `rm -rf spectre/changes/self-review-backlog-flow-5-oct18` and the
    no-argument `scripts/check-plan-shape.sh` exits 0. `bash scripts/test-check-plan-shape.sh`,
    `scripts/check-references.sh` and `scripts/check-guard-symlinks.sh` exit 0.
  - [x] **Step 4: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/plantaskidrule.go`, `stats/internal/guard/plan_task_id_rule_test.go`, `scripts/check-plan-shape.sh`
**Tests:** `TestPlanTaskIDRule`
**Regression:** `TestPlanTaskIDRule` — reverting the commit leaves `plan-task-id-rule` unregistered: the test panics on a nil `Registry` entry, and `check-plan-shape.sh` again passes a `/flow` plan stating no `Task-Id:` trailer.
**Commit:** fix(guard): refuse a /flow plan stating task commits carry no Task-Id trailer
**After:** none

**Decision:** guard-over-prose
**Decision:** go-subcommand
**Decision:** spectre-branch-discriminator
**Decision:** no-repo-clean
**Decision:** match-no-task-id
