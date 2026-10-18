package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of the plan-task-id-rule contract: a live /flow plan — its change
// has a spectre/<name> branch, local or on origin — stating task commits carry
// no `Task-Id:` trailer is refused, naming the line; a loop plan (no such
// branch), a fenced example, a plan requiring the trailer, an archived plan
// and a directory git cannot read pass.
func TestPlanTaskIDRule(t *testing.T) {
	const rule = "## Global constraints\n\n- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer.\n\n- [ ] 1. A\n\n**Files:** `a.md`\n"
	cases := []struct {
		name, rel, plan, ref string
		git                  bool
		code                 int
		want                 string
	}{
		{"/flow plan, local branch (self-review-backlog-flow-5-oct18)", "spectre/changes/c/tasks.md", rule, "refs/heads/spectre/c", true,
			1, "tasks.md:3: a /flow plan states task commits carry no `Task-Id:` trailer"},
		{"/flow plan, origin branch", "spectre/changes/c/tasks.md", rule, "refs/remotes/origin/spectre/c", true, 1, "tasks.md:3:"},
		{"loop plan, no spectre branch", "spectre/changes/c/tasks.md", rule, "refs/heads/spectre/other", true, 0, ""},
		{"fenced example", "spectre/changes/c/tasks.md", "```\nno `Task-Id:` trailer\n```\n", "refs/heads/spectre/c", true, 0, ""},
		{"trailer required", "spectre/changes/c/tasks.md", "- each commit carries a `Task-Id:` trailer\n", "refs/heads/spectre/c", true, 0, ""},
		{"archived plan", "spectre/changes/archive/c/tasks.md", rule, "refs/heads/spectre/c", true, 0, ""},
		{"not a git repository", "spectre/changes/c/tasks.md", rule, "", false, 0, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.git {
				gitRun(t, dir, "init", "-q")
				gitRun(t, dir, "commit", "-q", "--allow-empty", "-m", "init")
				gitRun(t, dir, "update-ref", tc.ref, "HEAD")
			}
			p := filepath.Join(dir, tc.rel)
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte(tc.plan), 0o644); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			code := Registry["plan-task-id-rule"]([]string{tc.rel}, Env{Dir: dir}, &stdout, &stderr)
			if code != tc.code {
				t.Fatalf("exit %d, want %d; stdout=%q stderr=%q", code, tc.code, stdout.String(), stderr.String())
			}
			if !strings.Contains(stdout.String(), tc.want) || (tc.code == 0 && stdout.Len() != 0) {
				t.Errorf("stdout %q, want %q", stdout.String(), tc.want)
			}
		})
	}
}
