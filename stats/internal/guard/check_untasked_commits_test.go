package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckUntaskedCommits builds one repository per case: a base commit,
// then the case's commits above it, and runs the guard over base..HEAD.
func TestCheckUntaskedCommits(t *testing.T) {
	t.Parallel()
	type commit struct{ path, subject, trailer string }
	cases := []struct {
		name    string
		commits []commit
		code    int
		want    []string
		absent  []string
	}{
		{"task commit passes", []commit{{"src/a.kt", "feat(app): a", "Task-Id: 1"}}, 0, []string{"UNTASKED-OK: 1 commit(s) checked"}, nil},
		{"flow-key commit passes", []commit{{"src/a.kt", "fix(app): review Minors", "Flow-Key: panel-fix-1"}}, 0, []string{"UNTASKED-OK"}, nil},
		{"chore commit passes", []commit{{"src/a.kt", "chore(spectre): plan demo", ""}}, 0, []string{"UNTASKED-OK"}, nil},
		{"planning-only commit passes", []commit{{"spectre/changes/demo/tasks.md", "docs(spectre): widen task 3", ""}}, 0, []string{"UNTASKED-OK"}, nil},
		{"trailerless code commit is listed", []commit{
			{"src/a.kt", "feat(app): a", "Task-Id: 1"},
			{"src/b.kt", "test(update): drop checks", ""},
		}, 1, []string{"UNTASKED: 1 commit(s) carry no Task-Id: or Flow-Key: trailer", "test(update): drop checks"}, []string{"feat(app): a"}},
		{"planning plus code is listed", []commit{{"spectre/changes/demo/tasks.md\nsrc/c.kt", "docs(spectre): note", ""}}, 1, []string{"docs(spectre): note"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			g := &fxGit{}
			g.git(dir, "init", "-q", "-b", "main")
			writeFile(t, filepath.Join(dir, "README"), "base\n")
			g.git(dir, "add", "-A")
			g.git(dir, "commit", "-qm", "base")
			base := g.git(dir, "rev-parse", "HEAD")
			for i, c := range tc.commits {
				for _, p := range strings.Split(c.path, "\n") {
					writeFile(t, filepath.Join(dir, p), c.subject+string(rune('a'+i))+"\n")
				}
				g.git(dir, "add", "-A")
				msg := []string{"commit", "-qm", c.subject}
				if c.trailer != "" {
					msg = append(msg, "-m", c.trailer)
				}
				g.git(dir, msg...)
			}
			if g.err != nil {
				t.Fatal(g.err)
			}
			var out, errb bytes.Buffer
			code := checkUntaskedCommits([]string{dir, base}, Env{Getenv: os.Getenv, Dir: dir}, &out, &errb)
			if code != tc.code {
				t.Fatalf("exit %d, want %d\nstdout:\n%s\nstderr:\n%s", code, tc.code, out.String(), errb.String())
			}
			for _, w := range tc.want {
				if !strings.Contains(out.String(), w) {
					t.Errorf("stdout lacks %q:\n%s", w, out.String())
				}
			}
			for _, a := range tc.absent {
				if strings.Contains(out.String(), a) {
					t.Errorf("stdout carries %q:\n%s", a, out.String())
				}
			}
		})
	}
}

// TestCheckUntaskedCommitsSkipsAMergedBranch pins --first-parent: a
// trailerless commit a merged-in branch brought is not the change's own
// (kan-941's develop merge listed develop's commits as UNTASKED).
func TestCheckUntaskedCommitsSkipsAMergedBranch(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	g := &fxGit{}
	g.git(dir, "init", "-q", "-b", "main")
	g.git(dir, "commit", "-q", "--allow-empty", "-m", "base")
	base := g.git(dir, "rev-parse", "HEAD")
	g.git(dir, "checkout", "-q", "-b", "develop")
	writeFile(t, filepath.Join(dir, "y.kt"), "y\n")
	g.git(dir, "add", "-A")
	g.git(dir, "commit", "-qm", "feat(y): develop work")
	g.git(dir, "checkout", "-q", "main")
	writeFile(t, filepath.Join(dir, "a.kt"), "a\n")
	g.git(dir, "add", "-A")
	g.git(dir, "commit", "-qm", "feat(app): a", "-m", "Task-Id: 1")
	g.git(dir, "merge", "-q", "--no-ff", "-m", "chore: merge develop", "develop")
	if g.err != nil {
		t.Fatal(g.err)
	}
	var out, errb bytes.Buffer
	if code := checkUntaskedCommits([]string{dir, base}, Env{Getenv: os.Getenv, Dir: dir}, &out, &errb); code != 0 || strings.Contains(out.String(), "develop work") {
		t.Fatalf("exit %d, want 0 with the merged branch's commit unlisted\nstdout:\n%s\nstderr:\n%s", code, out.String(), errb.String())
	}
}

// TestCheckUntaskedCommitsCannotAnswer pins exit 2 on a usage error, a
// missing worktree and a base git cannot resolve.
func TestCheckUntaskedCommitsCannotAnswer(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	g := &fxGit{}
	g.git(dir, "init", "-q", "-b", "main")
	g.git(dir, "commit", "-q", "--allow-empty", "-m", "base")
	if g.err != nil {
		t.Fatal(g.err)
	}
	for _, args := range [][]string{{dir}, {filepath.Join(dir, "absent"), "HEAD"}, {dir, "no-such-ref"}} {
		var out, errb bytes.Buffer
		if code := checkUntaskedCommits(args, Env{Getenv: os.Getenv, Dir: dir}, &out, &errb); code != 2 {
			t.Errorf("args %q: exit %d, want 2 (stdout %q)", args, code, out.String())
		}
	}
}

// TestCheckUntaskedCommitsRefusesAPipelineFixKey pins kan-1033's finding
// 1.2: a pipeline-fix Flow-Key belongs on a pipeline-fix-<name>-<k> branch
// alone; on a change branch it is listed.
func TestCheckUntaskedCommitsRefusesAPipelineFixKey(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, branch, trailer string
		code                  int
		want                  []string
	}{
		{"pipeline-fix key on a change branch is listed", "spectre/demo", "Flow-Key: pipeline-fix-mid-run-question", 1,
			[]string{"UNTASKED: 1 commit(s)", "fix(app): a — Flow-Key: pipeline-fix-… on a change branch"}},
		{"pipeline-fix key on a pipeline-fix branch passes", "pipeline-fix-demo-3", "Flow-Key: pipeline-fix-mid-run-question", 0,
			[]string{"UNTASKED-OK"}},
		{"panel-fix key on a change branch passes", "spectre/demo", "Flow-Key: panel-fix-1", 0, []string{"UNTASKED-OK"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			g := &fxGit{}
			g.git(dir, "init", "-q", "-b", "main")
			writeFile(t, filepath.Join(dir, "README"), "base\n")
			g.git(dir, "add", "-A")
			g.git(dir, "commit", "-qm", "base")
			base := g.git(dir, "rev-parse", "HEAD")
			g.git(dir, "checkout", "-qb", tc.branch)
			writeFile(t, filepath.Join(dir, "src/a.kt"), "a\n")
			g.git(dir, "add", "-A")
			g.git(dir, "commit", "-qm", "fix(app): a", "-m", tc.trailer)
			if g.err != nil {
				t.Fatal(g.err)
			}
			var out, errb bytes.Buffer
			code := checkUntaskedCommits([]string{dir, base}, Env{Getenv: os.Getenv, Dir: dir}, &out, &errb)
			if code != tc.code {
				t.Fatalf("exit %d, want %d\nstdout:\n%s\nstderr:\n%s", code, tc.code, out.String(), errb.String())
			}
			for _, w := range tc.want {
				if !strings.Contains(out.String(), w) {
					t.Errorf("stdout lacks %q:\n%s", w, out.String())
				}
			}
		})
	}
}
