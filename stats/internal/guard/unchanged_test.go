package guard

import (
	"os"
	"strings"
	"testing"
)

func TestUnchanged(t *testing.T) {
	t.Parallel()
	repo := t.TempDir() + "/r"
	var g fxGit
	g.git("", "init", "-q", "-b", "main", repo)
	g.git(repo, "commit", "-q", "--allow-empty", "-m", "base")
	g.git(repo, "update-ref", "refs/remotes/origin/main", "HEAD")
	g.err = os.MkdirAll(repo+"/docs", 0o755)
	g.write(repo+"/docs/a.md", "a")
	g.git(repo, "add", ".")
	g.git(repo, "commit", "-q", "-m", "docs")
	if g.err != nil {
		t.Fatal(g.err)
	}
	env := Env{Getenv: verifyFullGetenv(""), Dir: repo}
	specs := []string{".", ":!docs/"}

	if r := runGuard("unchanged", specs, env); r.rc != 0 || !strings.Contains(r.stdout, "suite skipped") {
		t.Errorf("only an excluded path changed: rc=%d out=%q, want 0 and the skip line", r.rc, r.out)
	}
	g.write(repo+"/x", "x")
	if r := runGuard("unchanged", specs, env); r.rc != 1 {
		t.Errorf("an untracked matching file: rc=%d, want 1", r.rc)
	}
	g.git(repo, "add", "x")
	g.git(repo, "commit", "-q", "-m", "x")
	if r := runGuard("unchanged", specs, env); g.err != nil || r.rc != 1 {
		t.Errorf("a committed matching file: rc=%d err=%v, want 1", r.rc, g.err)
	}
	if r := runGuard("unchanged", nil, env); r.rc != 2 || !strings.Contains(r.err, "usage:") {
		t.Errorf("no pathspec: rc=%d err=%q, want 2 and usage", r.rc, r.err)
	}
	if r := runGuard("unchanged", specs, Env{Getenv: verifyFullGetenv(""), Dir: t.TempDir()}); r.rc != 2 {
		t.Errorf("not a repository: rc=%d, want 2", r.rc)
	}
}

// verifyFullGetenv answers FLOW_VERIFY_FULL with full and everything else from os.Getenv.
func verifyFullGetenv(full string) func(string) string {
	return func(k string) string {
		if k == "FLOW_VERIFY_FULL" {
			return full
		}
		return os.Getenv(k)
	}
}

func TestUnchangedAnswersChangedUnderVerifyFull(t *testing.T) {
	t.Parallel()
	repo := t.TempDir() + "/r"
	var g fxGit
	g.git("", "init", "-q", "-b", "main", repo)
	g.git(repo, "commit", "-q", "--allow-empty", "-m", "base")
	g.git(repo, "update-ref", "refs/remotes/origin/main", "HEAD")
	g.err = os.MkdirAll(repo+"/docs", 0o755)
	g.write(repo+"/docs/a.md", "a")
	g.git(repo, "add", ".")
	g.git(repo, "commit", "-q", "-m", "docs")
	if g.err != nil {
		t.Fatal(g.err)
	}
	specs := []string{".", ":!docs/"}
	if r := runGuard("unchanged", specs, Env{Getenv: verifyFullGetenv("1"), Dir: repo}); r.rc != 1 || r.stdout != "" {
		t.Errorf("FLOW_VERIFY_FULL=1: rc=%d out=%q, want 1 and an empty stdout", r.rc, r.out)
	}
	if r := runGuard("unchanged", specs, Env{Getenv: verifyFullGetenv(""), Dir: repo}); r.rc != 0 || !strings.Contains(r.stdout, "suite skipped") {
		t.Errorf("FLOW_VERIFY_FULL empty: rc=%d out=%q, want 0 and the skip line", r.rc, r.out)
	}
	if r := runGuard("unchanged", specs, Env{Getenv: verifyFullGetenv("1"), Dir: t.TempDir()}); r.rc != 1 {
		t.Errorf("FLOW_VERIFY_FULL=1, not a repository: rc=%d out=%q, want 1", r.rc, r.out)
	}
	if r := runGuard("unchanged", nil, Env{Getenv: verifyFullGetenv("1"), Dir: repo}); r.rc != 2 || !strings.Contains(r.err, "usage:") {
		t.Errorf("FLOW_VERIFY_FULL=1, no pathspec: rc=%d err=%q, want 2 and usage", r.rc, r.err)
	}
}
