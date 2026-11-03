package repotest

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestImplementerBriefFixCommitTrailersParse pins the kan-1053 self-review
// fix, F2: FIX DISPATCH, followed literally, committed `Task-Id:` and
// `Flow-Key:` as two paragraphs, and git reads trailers from the last
// paragraph alone. The brief's fix-commit command is extracted, run in a
// scratch repository, and both trailers must parse from the commit it makes.
func TestImplementerBriefFixCommitTrailersParse(t *testing.T) {
	t.Parallel()
	src, err := os.ReadFile(filepath.Join(repoRoot(t), "skills", "flow", "implementer-brief.md"))
	if err != nil {
		t.Fatal(err)
	}
	var para []string
	for _, l := range strings.Split(string(src), "\n") {
		if len(para) == 0 && !strings.HasPrefix(l, "> **FIX DISPATCH:**") {
			continue
		}
		if strings.TrimSpace(l) == "" {
			break
		}
		para = append(para, strings.TrimPrefix(l, "> "))
	}
	text := strings.Join(para, " ")
	if text == "" {
		t.Fatal("no `> **FIX DISPATCH:**` paragraph in skills/flow/implementer-brief.md")
	}
	if strings.Contains(text, `-m "Flow-Key:`) {
		t.Error("FIX DISPATCH spells Flow-Key as its own -m paragraph")
	}
	m := regexp.MustCompile("`(git -C <abs-worktree> commit [^`]*)`").FindStringSubmatch(text)
	if m == nil {
		t.Fatal("FIX DISPATCH carries no `git -C <abs-worktree> commit …` command")
	}

	dir := t.TempDir()
	id := []string{"-c", "user.email=t@t", "-c", "user.name=t"}
	sh(t, dir, nil, "git", "init", "-q")
	sh(t, dir, nil, "git", "checkout", "-q", "-b", "t")
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("package a\n")
	sh(t, dir, nil, "git", "add", "a.go")
	sh(t, dir, nil, "git", append(id, "commit", "-qm", "init")...)
	write("package a\n\n// fixed\n")

	command := strings.NewReplacer(
		"<abs-worktree>", dir, "<module>", "x", "<subject>", "s", "<n>", "7",
		"<your key>", "task-7-fix", "<this task's files>", "a.go",
	).Replace(m[1])
	cmd := exec.Command("bash", "-c", command)
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s: %v\n%s", command, err, out)
	}
	for key, want := range map[string]string{"Task-Id": "7", "Flow-Key": "task-7-fix"} {
		if got := sh(t, dir, nil, "git", "log", "-1", "--format=%(trailers:key="+key+",valueonly)"); got != want {
			t.Errorf("trailer %s = %q, want %q (command: %s)", key, got, want, command)
		}
	}
}
