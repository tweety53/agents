package guard

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

// The throwaway-worktree guard replaces three bash fences: the wave-group
// create fence in skills/flow/sdd-dispatch.md and the slot create and
// fold-back-and-remove fences in skills/flow/review-panel-optional-slots.md.
// Each test runs the fence as it stood at ae805186 (Decision:
// parity-ref-on-main) and the guard over identical fixtures, and asserts the
// trees they leave are the same.

// twFence is the n-th ```bash fence after marker in the file at ae805186,
// with each placeholder replaced by its value.
func twFence(t *testing.T, doc, marker string, n int, repl ...string) string {
	t.Helper()
	src, err := exec.Command(fixtureGit, "-C", "../../..", "show", "ae805186:"+doc).Output()
	if err != nil {
		t.Fatalf("git show ae805186:%s: %v", doc, err)
	}
	s := string(src)
	i := strings.Index(s, marker)
	if i < 0 {
		t.Fatalf("%s at ae805186 carries no %q", doc, marker)
	}
	fences := regexp.MustCompile("(?s)```bash\n(.*?)```").FindAllStringSubmatch(s[i:], n)
	if len(fences) < n {
		t.Fatalf("%s at ae805186 carries fewer than %d fences after %q", doc, n, marker)
	}
	return strings.NewReplacer(repl...).Replace(fences[n-1][1])
}

// twRepo is a committed repository whose working tree carries every kind of
// uncommitted change the create fence has to carry over: an unstaged edit, a
// staged addition, a staged rename, a deletion, untracked files at the root
// and in a new directory, and an ignored .superpowers/sdd.
func twRepo(t *testing.T) string {
	t.Helper()
	wt := filepath.Join(t.TempDir(), "wt")
	mkdir(t, wt)
	gitRun(t, wt, "init", "-q")
	writeFile(t, wt+"/.gitignore", ".superpowers/\n")
	writeFile(t, wt+"/a.txt", "a\n")
	writeFile(t, wt+"/b.txt", "b\n")
	writeFile(t, wt+"/old.txt", "moved\n")
	writeExec(t, wt+"/bin/run.sh", "#!/bin/sh\necho run\n")
	gitRun(t, wt, "add", "-A")
	gitRun(t, wt, "commit", "-q", "-m", "base")
	writeFile(t, wt+"/a.txt", "a edited\n")
	writeFile(t, wt+"/staged.txt", "staged\n")
	gitRun(t, wt, "add", "staged.txt")
	gitRun(t, wt, "mv", "old.txt", "new.txt")
	gitRun(t, wt, "rm", "-q", "b.txt")
	writeFile(t, wt+"/top.txt", "untracked\n")
	writeFile(t, wt+"/u/deep/x.txt", "untracked deep\n")
	writeFile(t, wt+"/.superpowers/sdd/dispatch.md", "bundle\n")
	writeFile(t, wt+"/.superpowers/sdd/reproducers/r.sh", "repro\n")
	return wt
}

// twTree is every file under root but .git: its path, mode and content.
func twTree(t *testing.T, root string) string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == ".git" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		out = append(out, rel+" "+info.Mode().String()+" "+string(b))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(out)
	return strings.Join(out, "")
}

func twBash(t *testing.T, script string) {
	t.Helper()
	cmd := exec.Command("bash", "-c", script)
	cmd.Env = append(os.Environ(), fixtureGitEnv...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fence: %v\n%s", err, out)
	}
}

func twGuard(t *testing.T, args ...string) {
	t.Helper()
	var out, errb bytes.Buffer
	run, ok := Registry["throwaway-worktree"]
	if !ok {
		t.Fatal("no throwaway-worktree guard registered")
	}
	if rc := run(args, Env{Getenv: os.Getenv}, &out, &errb); rc != 0 {
		t.Fatalf("throwaway-worktree %v: exit %d\nstdout:\n%s\nstderr:\n%s", args, rc, out.String(), errb.String())
	}
}

func twStatus(t *testing.T, dir string) string {
	t.Helper()
	out, err := exec.Command(fixtureGit, "-C", dir, "status", "--porcelain").Output()
	if err != nil {
		t.Fatalf("git status in %s: %v", dir, err)
	}
	return string(out)
}

func TestThrowawayWorktreeCreate(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, doc, marker, placeholder string
		sdd                            bool
	}{
		{"wave group", "skills/flow/sdd-dispatch.md", "**Waves — concurrent dispatch", "<worktree>-wave-group-<g>", false},
		{"slot --sdd", "skills/flow/review-panel-optional-slots.md", "## The throwaway worktree", "<worktree>-<slot>-<round>", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			bashWT := twRepo(t)
			twBash(t, twFence(t, tc.doc, tc.marker, 1, tc.placeholder, bashWT+"-copy", "<worktree>", bashWT))
			goWT := twRepo(t)
			args := []string{"create", goWT, goWT + "-copy"}
			if tc.sdd {
				args = append(args, "--sdd")
			}
			twGuard(t, args...)

			want, got := twTree(t, bashWT+"-copy"), twTree(t, goWT+"-copy")
			if got != want {
				t.Errorf("copy tree differs from the fence's\ngot:\n%s\nwant:\n%s", got, want)
			}
			if w, g := twStatus(t, bashWT+"-copy"), twStatus(t, goWT+"-copy"); g != w {
				t.Errorf("copy status differs from the fence's\ngot:\n%s\nwant:\n%s", g, w)
			}
			for _, f := range []string{"a.txt", "new.txt", "staged.txt", "top.txt", "u/deep/x.txt"} {
				if _, err := os.Stat(goWT + "-copy/" + f); err != nil {
					t.Errorf("copy lacks %s: %v", f, err)
				}
			}
			for _, f := range []string{"b.txt", "old.txt"} {
				if _, err := os.Stat(goWT + "-copy/" + f); err == nil {
					t.Errorf("copy still carries %s", f)
				}
			}
			_, err := os.Stat(goWT + "-copy/.superpowers/sdd/dispatch.md")
			if tc.sdd != (err == nil) {
				t.Errorf("--sdd=%v but .superpowers/sdd/dispatch.md present=%v", tc.sdd, err == nil)
			}
		})
	}
}

// twFoldFixture is twRepo plus a detached copy whose .superpowers/sdd holds
// a newer report, an older one, an empty one, a new reproducer and a file
// the fold-back never reads.
func twFoldFixture(t *testing.T) (wt, cp string) {
	t.Helper()
	wt = twRepo(t)
	cp = wt + "-mutation-1"
	gitRun(t, wt, "worktree", "add", "-q", "--detach", cp, "HEAD")
	old, now := time.Unix(1_700_000_000, 0), time.Unix(1_700_000_100, 0)
	for _, f := range []struct {
		path, body string
		at         time.Time
	}{
		{wt + "/.superpowers/sdd/panel-report-a.md", "wt a\n", old},
		{cp + "/.superpowers/sdd/panel-report-a.md", "copy a newer\n", now},
		{wt + "/.superpowers/sdd/panel-report-c.md", "wt c newer\n", now},
		{cp + "/.superpowers/sdd/panel-report-c.md", "copy c\n", old},
		{cp + "/.superpowers/sdd/panel-report-b.md", "copy b only\n", old},
		{cp + "/.superpowers/sdd/panel-report-e.md", "", now},
		{cp + "/.superpowers/sdd/reproducers/new.sh", "copy repro\n", now},
		{cp + "/.superpowers/sdd/other.md", "never folded\n", now},
	} {
		writeFile(t, f.path, f.body)
		if err := os.Chtimes(f.path, f.at, f.at); err != nil {
			t.Fatal(err)
		}
	}
	return wt, cp
}

func TestThrowawayWorktreeRemoveFoldBack(t *testing.T) {
	t.Parallel()
	for _, foldBack := range []bool{true, false} {
		name := "remove"
		if foldBack {
			name = "remove --fold-back"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			bashWT, bashCP := twFoldFixture(t)
			fence := "git -C <worktree> worktree remove --force <worktree>-<slot>-<round>\n"
			if foldBack {
				fence = twFence(t, "skills/flow/review-panel-optional-slots.md", "## The throwaway worktree", 2, "<worktree>-<slot>-<round>", bashCP, "<worktree>", bashWT)
			} else {
				fence = strings.NewReplacer("<worktree>-<slot>-<round>", bashCP, "<worktree>", bashWT).Replace(fence)
			}
			twBash(t, fence)
			goWT, goCP := twFoldFixture(t)
			args := []string{"remove", goWT, goCP}
			if foldBack {
				args = append(args, "--fold-back")
			}
			twGuard(t, args...)

			want, got := twTree(t, bashWT+"/.superpowers/sdd"), twTree(t, goWT+"/.superpowers/sdd")
			if got != want {
				t.Errorf("worktree's .superpowers/sdd differs from the fence's\ngot:\n%s\nwant:\n%s", got, want)
			}
			if _, err := os.Stat(goCP); err == nil {
				t.Errorf("copy %s still exists", goCP)
			}
			list, err := exec.Command(fixtureGit, "-C", goWT, "worktree", "list", "--porcelain").Output()
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(list), goCP) {
				t.Errorf("git still lists the copy:\n%s", list)
			}
			b, _ := os.ReadFile(goWT + "/.superpowers/sdd/panel-report-a.md")
			if want := map[bool]string{true: "copy a newer\n", false: "wt a\n"}[foldBack]; string(b) != want {
				t.Errorf("panel-report-a.md = %q, want %q", b, want)
			}
		})
	}
}

// A failed fold-back still exits 2 but never leaves the copy registered: the
// fence removed the copy whatever the fold-back did.
func TestThrowawayWorktreeRemoveFoldBackFailureStillRemoves(t *testing.T) {
	t.Parallel()
	wt, cp := twFoldFixture(t)
	// A file where the fold-back needs the reproducers directory makes its
	// MkdirAll fail.
	if err := os.RemoveAll(wt + "/.superpowers/sdd/reproducers"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(wt+"/.superpowers/sdd/reproducers", []byte("not a dir\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if rc := Registry["throwaway-worktree"]([]string{"remove", wt, cp, "--fold-back"}, Env{Getenv: os.Getenv}, &out, &errb); rc != 2 {
		t.Fatalf("exit %d, want 2\nstderr:\n%s", rc, errb.String())
	}
	if _, err := os.Stat(cp); err == nil {
		t.Errorf("copy %s still exists after a failed fold-back", cp)
	}
	list, err := exec.Command(fixtureGit, "-C", wt, "worktree", "list", "--porcelain").Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(list), cp) {
		t.Errorf("git still lists the copy:\n%s", list)
	}
}
