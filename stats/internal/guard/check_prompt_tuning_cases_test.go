package guard

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/prompttuning"
)

// ptRepo is a scratch agents repo carrying one shared prompt and one scoped
// rule file at its only commit, which every fixture case's target_sha names.
func ptRepo(t *testing.T) (repo, sha string) {
	t.Helper()
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repo = tmp + "/agents"
	writeFile(t, repo+"/skills/flow/implement.md", "# Implement\n\nWrite a test for every exit path.\n")
	writeFile(t, repo+"/rules/layer-backend.mdc", "# Backend\n\nClose every resource.\n")
	gitRun(t, "", "init", "-q", "-b", "main", repo)
	gitRun(t, repo, "add", "-A")
	gitRun(t, repo, "commit", "-q", "-m", "init")
	out, err := exec.Command(fixtureGit, "-C", repo, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return repo, strings.TrimSpace(string(out))
}

// ptCase is a complete, generic implementer case.
func ptCase(sha string) prompttuning.Case {
	return prompttuning.Case{
		ID: "demo-change-1", Type: "implementer", Scope: "generic",
		Source: prompttuning.Source{
			Project: "agents", Change: "demo-change", Jira: "KAN-77", Repo: "/repo",
			FixSHA: "abc1234", FindingRef: "F-1", Round: 2, Slot: "primary", Severity: "Important",
		},
		Target: "skills/flow/implement.md", TargetSHA: sha,
		DefectClass: "an exit path shipped without a test",
		WhyPicked:   "the four parts",
		Before:      "Write a test for every exit path.",
		After:       "Write a test for every exit path, the refusals included.",
	}
}

// ptWrite writes c and its repro files into root/<c.ID>, index.html rendered.
func ptWrite(t *testing.T, root string, c prompttuning.Case) string {
	t.Helper()
	dir := filepath.Join(root, c.ID)
	f := prompttuning.Files{
		Finding:  "the refusal path has no test",
		BadDiff:  "--- a/stats/thing/widget.go\n+++ b/stats/thing/widget.go\n+bad\n",
		GoodDiff: "--- a/stats/thing/widget.go\n+++ b/stats/thing/widget.go\n+good\n",
		Task:     "the task text",
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	page, err := prompttuning.Render(c, f)
	if err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"case.json": string(data), "finding.md": f.Finding, "bad.diff": f.BadDiff,
		"good.diff": f.GoodDiff, "task.md": f.Task, "index.html": string(page),
	} {
		writeFile(t, filepath.Join(dir, name), body)
	}
	return dir
}

func TestCheckPromptTuningCases(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		edit   func(c *prompttuning.Case)
		after  func(t *testing.T, dir, root string)
		rc     int
		out    string
		noRoot bool
	}{
		{name: "a complete generic case passes"},
		{name: "a missing good.diff is caught", after: func(t *testing.T, dir, _ string) {
			if err := os.Remove(dir + "/good.diff"); err != nil {
				t.Fatal(err)
			}
		}, rc: 1, out: "good.diff: missing"},
		{name: "a reviewer case without reviewer_kind is caught", edit: func(c *prompttuning.Case) { c.Type = "reviewer" },
			rc: 1, out: "reviewer_kind"},
		{name: "a before absent at target_sha is caught", edit: func(c *prompttuning.Case) { c.Before = "No such line." },
			rc: 1, out: "before does not occur"},
		{name: "a generic case targeting a rules file is caught", edit: func(c *prompttuning.Case) { c.Target = "rules/lang-kotlin.mdc" },
			rc: 1, out: "target"},
		{name: "a language case targeting skills is caught", edit: func(c *prompttuning.Case) { c.Scope = "language:kotlin" },
			rc: 1, out: "target"},
		{name: "a language case creating its absent rule file passes", edit: func(c *prompttuning.Case) {
			c.Scope, c.Target, c.Before = "language:kotlin", "rules/lang-kotlin.mdc", ""
		}},
		{name: "a layer case targeting its rule file passes", edit: func(c *prompttuning.Case) {
			c.Scope, c.Target, c.Before = "layer:backend", "rules/layer-backend.mdc", "Close every resource."
		}},
		{name: "a language/layer case targeting the language file is caught", edit: func(c *prompttuning.Case) {
			c.Scope, c.Target, c.Before = "language:kotlin/layer:frontend", "rules/lang-kotlin.mdc", ""
		}, rc: 1, out: "target"},
		{name: "a layer:web scope is caught", edit: func(c *prompttuning.Case) { c.Scope = "layer:web" },
			rc: 1, out: "scope"},
		{name: "an after naming the change is caught", edit: func(c *prompttuning.Case) { c.After += " As in Demo-Change." },
			rc: 1, out: `names "demo-change"`},
		{name: "an after naming the Jira key is caught", edit: func(c *prompttuning.Case) { c.After += " See KAN-77." },
			rc: 1, out: `names "KAN-77"`},
		{name: "an after naming any kan key is caught", edit: func(c *prompttuning.Case) { c.After += " Like kan-12." },
			rc: 1, out: `names "kan-12"`},
		{name: "an after naming a diff basename is caught", edit: func(c *prompttuning.Case) { c.After += " Mind widget.go." },
			rc: 1, out: `names "widget.go"`},
		{name: "a hand-edited index.html is caught", after: func(t *testing.T, dir, _ string) {
			writeFile(t, dir+"/index.html", "<html>hand</html>\n")
		}, rc: 1, out: "index.html: differs"},
		{name: "a malformed ledger line is caught", after: func(t *testing.T, _, root string) {
			writeFile(t, root+"/LEDGER.md", "# Prompt-tuning ledger\n\n"+
				"- 2026-10-25 · `x-1` · reviewer · `skills/flow/implement.md` · a class — an update — applied abc1234\n"+
				"- 2026-10-25 · x-2 · reviewer — broken\n")
		}, rc: 1, out: "LEDGER.md:4:"},
		{name: "a false-positive reviewer case passes", edit: func(c *prompttuning.Case) { c.Type, c.ReviewerKind = "reviewer", "false-positive" }},
		{name: "a target_sha naming no commit is caught", edit: func(c *prompttuning.Case) {
			c.Scope, c.Target, c.Before, c.TargetSHA = "language:kotlin", "rules/lang-kotlin.mdc", "", "0000000000000000000000000000000000000000"
		}, rc: 1, out: "names no commit"},
		{name: "an empty before for an existing target is caught", edit: func(c *prompttuning.Case) {
			c.Scope, c.Target, c.Before = "layer:backend", "rules/layer-backend.mdc", ""
		}, rc: 1, out: "before is empty"},
		{name: "a scoped before for an absent target is caught", edit: func(c *prompttuning.Case) {
			c.Scope, c.Target, c.Before = "language:kotlin", "rules/lang-kotlin.mdc", "Some rule."
		}, rc: 1, out: "is absent at"},
		{name: "an unset source round is caught", edit: func(c *prompttuning.Case) { c.Source.Round = 0 },
			rc: 1, out: "source.round"},
		{name: "an absent root exits 0", noRoot: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo, sha := ptRepo(t)
			root := repo + "/docs/prompt-tuning"
			if !tc.noRoot {
				c := ptCase(sha)
				if tc.edit != nil {
					tc.edit(&c)
				}
				dir := ptWrite(t, root, c)
				if tc.after != nil {
					tc.after(t, dir, root)
				}
			}
			env := Env{Getenv: func(k string) string {
				if k == "FLOW_GUARD_REPO_ROOT" {
					return repo
				}
				return os.Getenv(k)
			}, Dir: repo}
			var o, e bytes.Buffer
			rc := checkPromptTuningCases(nil, env, &o, &e)
			out := o.String() + e.String()
			if rc != tc.rc {
				t.Fatalf("rc %d, want %d: %s", rc, tc.rc, out)
			}
			if tc.out != "" && !strings.Contains(out, tc.out) {
				t.Fatalf("output %q carries no %q", out, tc.out)
			}
			if tc.rc == 0 && out != "" {
				t.Fatalf("a clean run printed %q", out)
			}
		})
	}
}
