package guard

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// fastRecordRepo builds a fixture repository in a temp dir: one commit on
// main with refs/remotes/origin/main pointing at it the way a fetch would,
// then the named commits replayed on the change branch `name`. Each commit
// is a subject and an optional body.
func fastRecordRepo(t *testing.T, name string, commits ...[2]string) string {
	t.Helper()
	dir := t.TempDir()
	gitRun(t, dir, "init", "-q", "-b", "main")
	gitRun(t, dir, "commit", "--allow-empty", "-q", "-m", "init: base")
	gitRun(t, dir, "update-ref", "refs/remotes/origin/main", "HEAD")
	gitRun(t, dir, "checkout", "-q", "-b", name)
	for _, c := range commits {
		args := []string{"commit", "--allow-empty", "-q", "-m", c[0]}
		if c[1] != "" {
			args = append(args, "-m", c[1])
		}
		gitRun(t, dir, args...)
	}
	return dir
}

func runFastRecord(t *testing.T, wt, base string, argv []string) (int, string, string) {
	t.Helper()
	if argv == nil {
		argv = []string{wt, base}
	}
	var stdout, stderr bytes.Buffer
	code := checkFastRouteRecord(argv, Env{Getenv: os.Getenv, Dir: wt}, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestFastRouteRecordClean(t *testing.T) {
	wt := fastRecordRepo(t, "kan-838-demo",
		[2]string{"feat(guard): read the series as the record", ""},
		[2]string{"docs(flow): point the fast route at the guard", ""},
	)
	code, out, errOut := runFastRecord(t, wt, "main", nil)
	if code != 0 || out != "" || errOut != "" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, out, errOut)
	}
}

func TestFastRouteRecordBreakingMarker(t *testing.T) {
	for _, subject := range []string{"feat!: reshape the record", "feat(guard)!: reshape the record"} {
		wt := fastRecordRepo(t, "kan-838-demo", [2]string{subject, ""})
		if code, out, _ := runFastRecord(t, wt, "main", nil); code != 0 || out != "" {
			t.Fatalf("%q: exit %d, stdout %q", subject, code, out)
		}
	}
}

func TestFastRouteRecordEmptySeries(t *testing.T) {
	wt := fastRecordRepo(t, "kan-838-demo")
	if code, out, errOut := runFastRecord(t, wt, "main", nil); code != 0 || out != "" || errOut != "" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, out, errOut)
	}
}

func TestFastRouteRecordFindings(t *testing.T) {
	tests := []struct {
		name    string
		branch  string
		subject string
		body    string
		want    string
	}{
		{"bad subject", "kan-838-demo", "updated the guard", "", "subject not in Conventional Commits form"},
		{"scope names the change", "kan-838-demo", "feat(kan-838-demo): x", "", `scope "kan-838-demo" names the change`},
		{"scope names a Jira key", "kan-838-demo", "fix(kan-838-flow): x", "", `scope "kan-838-flow" names a Jira key`},
		{"scope names a task id", "kan-838-demo", "fix(2): x", "", `scope "2" names a task id`},
		{"scope names a task- word id", "kan-838-demo", "fix(task-3): x", "", `scope "task-3" names a task id`},
		{"scope names a step id", "kan-838-demo", "fix(3/4): x", "", `scope "3/4" names a task id`},
		{"co-authored trailer", "kan-838-demo", "feat(guard): x", "Co-Authored-By: Claude <n@example.com>", "attribution trailer"},
		{"generated footer", "kan-838-demo", "feat(guard): x", "🤖 Generated with [Claude Code](https://claude.com/claude-code)", "attribution footer"},
		{"task-id trailer", "kan-838-demo", "feat(guard): x", "Task-Id: 3", "Task-Id trailer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wt := fastRecordRepo(t, tt.branch, [2]string{tt.subject, tt.body})
			code, out, _ := runFastRecord(t, wt, "main", nil)
			if code != 1 {
				t.Fatalf("exit %d, want 1", code)
			}
			if !strings.Contains(out, tt.want) {
				t.Fatalf("stdout %q lacks %q", out, tt.want)
			}
			// The finding names its commit: the line opens with the sha
			// rev-list printed.
			sha, err := exec.Command(fixtureGit, append([]string{"-C", wt}, "rev-parse", "HEAD")...).Output()
			if err != nil || !strings.HasPrefix(out, strings.TrimSpace(string(sha))) {
				t.Fatalf("finding line does not open with HEAD's sha: %q (%v)", out, err)
			}
		})
	}
}

func TestFastRouteRecordWalkOrder(t *testing.T) {
	wt := fastRecordRepo(t, "kan-838-demo",
		[2]string{"older bad subject", ""},
		[2]string{"newer bad subject", ""},
	)
	code, out, _ := runFastRecord(t, wt, "main", nil)
	if code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("want two finding lines, got %q", out)
	}
	// Oldest first: rev-list without --reverse is newest first, so the
	// findings order is its reverse.
	all, err := exec.Command(fixtureGit, "-C", wt, "rev-list", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	revs := strings.Split(strings.TrimSpace(string(all)), "\n")
	if !strings.HasPrefix(lines[0], revs[len(revs)-2]) || !strings.Contains(lines[0], "older bad subject") ||
		!strings.HasPrefix(lines[1], revs[0]) || !strings.Contains(lines[1], "newer bad subject") {
		t.Fatalf("findings not oldest-first: %q (revs %v)", out, revs)
	}
}

func TestFastRouteRecordMergeCommit(t *testing.T) {
	wt := fastRecordRepo(t, "kan-838-demo", [2]string{"feat(guard): fine", ""})
	gitRun(t, wt, "checkout", "-q", "-b", "side", "main")
	gitRun(t, wt, "commit", "--allow-empty", "-q", "-m", "feat(side): elsewhere")
	gitRun(t, wt, "checkout", "-q", "kan-838-demo")
	gitRun(t, wt, "merge", "--no-ff", "-q", "-m", "Merge branch 'side'", "side")
	code, out, _ := runFastRecord(t, wt, "main", nil)
	if code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
	if !strings.Contains(out, "subject not in Conventional Commits form: Merge branch 'side'") {
		t.Fatalf("stdout %q lacks the merge-commit finding", out)
	}
}

func TestFastRouteRecordBaseFallback(t *testing.T) {
	dir := t.TempDir()
	gitRun(t, dir, "init", "-q", "-b", "main")
	gitRun(t, dir, "commit", "--allow-empty", "-q", "-m", "init: base")
	// No refs/remotes/origin/main: the base resolves through the bare-name
	// fallback leg, as a revision of its own.
	gitRun(t, dir, "checkout", "-q", "-b", "kan-838-demo")
	gitRun(t, dir, "commit", "--allow-empty", "-q", "-m", "feat(guard): fine")
	if code, out, errOut := runFastRecord(t, dir, "main", nil); code != 0 || out != "" || errOut != "" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, out, errOut)
	}
}

func TestFastRouteRecordProseBodyClean(t *testing.T) {
	// A prose body that merely opens a line with "Generated with" is not an
	// attribution banner — the banner shape is the copied-in footer's own,
	// a markdown link with an optional 🤖 prefix.
	wt := fastRecordRepo(t, "kan-838-demo",
		[2]string{"test(guard): cover the generated footer", "Generated with the new fixture helper, the report lists every finding."},
		[2]string{"feat(guard): real banner caught", "🤖 Generated with [Claude Code](https://claude.com/claude-code)"},
		[2]string{"feat(guard): bare banner caught", "Generated with [Some Tool](https://example.com/tool)"},
	)
	code, out, _ := runFastRecord(t, wt, "main", nil)
	if code != 1 {
		t.Fatalf("exit %d, want 1 (the two real banners)", code)
	}
	if strings.Contains(out, "attribution footer: Generated with the new fixture helper") {
		t.Fatalf("the prose body was flagged: %q", out)
	}
	if got := strings.Count(out, "attribution"); got != 2 {
		t.Fatalf("want exactly two attribution findings, got %d: %q", got, out)
	}
}

func TestFastRouteRecordCannotAnswer(t *testing.T) {
	wt := fastRecordRepo(t, "kan-838-demo")
	if code, out, errOut := runFastRecord(t, wt, "", []string{}); code != 2 || out != "" || errOut == "" {
		t.Fatalf("no args: exit %d, stdout %q, stderr %q", code, out, errOut)
	}
	if code, out, _ := runFastRecord(t, t.TempDir(), "main", nil); code != 2 || out != "" {
		t.Fatalf("not a repo: exit %d, stdout %q", code, out)
	}
	if code, out, _ := runFastRecord(t, wt, "nosuch", nil); code != 2 || out != "" {
		t.Fatalf("bad base: exit %d, stdout %q", code, out)
	}
}
