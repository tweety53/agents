package guard

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// Every case_N of scripts/test-prove-reproducer.sh at d71a2327, one subtest
// per case named after the case's own comment, plus the Review Focus rows:
// the tree, HEAD and worktree list unchanged after a refused or spent leg,
// and a worktree path with a space. Both legs run the Go run-reproducer
// in-process against a real git repository, so the scratch worktree is a
// real `git worktree add`.

// prMaster is the harness's make_repo + defect_fixture + fix_commit, built
// once: init, a defect commit writing check.txt "old-behaviour", and a fix
// commit rewriting it "new-behaviour". It returns the repository and the
// defect commit's sha.
var prMaster = sync.OnceValues(func() ([2]string, error) {
	dir, err := os.MkdirTemp(execFixtures.dir, "prove-reproducer-master")
	if err != nil {
		return [2]string{}, err
	}
	git := func(args ...string) (string, error) {
		cmd := exec.Command(fixtureGit, append([]string{"-C", dir}, args...)...)
		cmd.Env = append(append(os.Environ(), fixtureGitEnv...),
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test")
		out, err := cmd.CombinedOutput()
		return strings.TrimSpace(string(out)), err
	}
	steps := [][]string{{"init", "-q"}, {"commit", "-q", "--allow-empty", "-m", "init"}}
	for _, s := range steps {
		if out, err := git(s...); err != nil {
			return [2]string{}, &prGitError{s, out}
		}
	}
	commit := func(body, msg string) (string, error) {
		if err := os.WriteFile(filepath.Join(dir, "check.txt"), []byte(body+"\n"), 0o644); err != nil {
			return "", err
		}
		for _, s := range [][]string{{"add", "check.txt"}, {"commit", "-q", "-m", msg}} {
			if out, err := git(s...); err != nil {
				return "", &prGitError{s, out}
			}
		}
		return git("rev-parse", "HEAD")
	}
	pre, err := commit("old-behaviour", "defect")
	if err != nil {
		return [2]string{}, err
	}
	if _, err := commit("new-behaviour", "fix"); err != nil {
		return [2]string{}, err
	}
	return [2]string{dir, pre}, nil
})

type prGitError struct {
	args []string
	out  string
}

func (e *prGitError) Error() string { return "git " + strings.Join(e.args, " ") + ": " + e.out }

// prRepo copies the master into its own directory under the test's temp dir
// (named sub, so a case can put a space in it) and returns it with the
// defect sha.
func prRepo(t *testing.T, sub string) (string, string) {
	t.Helper()
	m, err := prMaster()
	if err != nil {
		t.Fatal(err)
	}
	repo := filepath.Join(t.TempDir(), sub)
	if err := os.CopyFS(repo, os.DirFS(m[0])); err != nil {
		t.Fatal(err)
	}
	return repo, m[1]
}

// prReproducer is the harness's reproducer_fixture: an executable #!/bin/sh
// script at rel carrying body.
func prReproducer(t *testing.T, repo, rel, body string) {
	t.Helper()
	writeFile(t, filepath.Join(repo, rel), "#!/bin/sh\n"+body+"\n")
	if err := os.Chmod(filepath.Join(repo, rel), 0o755); err != nil {
		t.Fatal(err)
	}
}

type prResult struct {
	rc          int
	out, errs   string
	git0, git1  string // status, HEAD and worktree list before and after
	tmpLeftover []string
}

func prGitState(t *testing.T, repo string) string {
	t.Helper()
	var b strings.Builder
	for _, args := range [][]string{{"status", "--porcelain"}, {"rev-parse", "HEAD"}, {"worktree", "list", "--porcelain"}} {
		out, err := exec.Command(fixtureGit, append([]string{"-C", repo}, args...)...).Output()
		if err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
		b.Write(out)
	}
	return b.String()
}

func prRun(t *testing.T, repo string, env Env, args ...string) prResult {
	t.Helper()
	fn := Registry["prove-reproducer"]
	if fn == nil {
		t.Fatal("prove-reproducer is not registered")
	}
	tmp := t.TempDir()
	if env.Getenv == nil {
		env.Getenv = func(k string) string {
			if k == "TMPDIR" {
				return tmp
			}
			return os.Getenv(k)
		}
	}
	if env.Dir == "" {
		env.Dir = t.TempDir()
	}
	env.ReproducerBound, env.ReproducerGrace = rrTestBound, rrTestGrace
	var r prResult
	if repo != "" {
		r.git0 = prGitState(t, repo)
	}
	var out, errb bytes.Buffer
	r.rc = fn(args, env, &out, &errb)
	r.out, r.errs = out.String(), errb.String()
	if repo != "" {
		r.git1 = prGitState(t, repo)
	}
	ents, _ := os.ReadDir(tmp)
	for _, e := range ents {
		r.tmpLeftover = append(r.tmpLeftover, e.Name())
	}
	return r
}

// prUnchanged is the Review Focus row: status, HEAD and the worktree list
// are what they were before the run, and the scratch's temp base is gone.
func prUnchanged(t *testing.T, r prResult) {
	t.Helper()
	if r.git0 != r.git1 {
		t.Fatalf("git state changed:\nbefore:\n%s\nafter:\n%s", r.git0, r.git1)
	}
	if len(r.tmpLeftover) != 0 {
		t.Fatalf("temporary base survived: %v", r.tmpLeftover)
	}
}

func prWant(t *testing.T, r prResult, rc int, needle string) {
	t.Helper()
	if r.rc != rc || !strings.Contains(r.out+r.errs, needle) {
		t.Fatalf("want exit %d naming %q, got %d\nstdout:\n%s\nstderr:\n%s", rc, needle, r.rc, r.out, r.errs)
	}
}

const prRel = ".superpowers/sdd/reproducers/0-primary-1.sh"

func TestProveReproducer(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{"case_1 the proof holds and the scratch worktree is gone", func(t *testing.T) {
			repo, pre := prRepo(t, "repo")
			prReproducer(t, repo, prRel, "exec sh -c '! grep -q old-behaviour check.txt'")
			r := prRun(t, repo, Env{}, repo, pre, prRel)
			prWant(t, r, 0, "PROOF HELD — pre-fix: defect demonstrated (exit 0); post-fix: defect not demonstrated (exit 1)\n")
			if !strings.HasPrefix(r.out, "== pre-fix leg — scratch worktree at "+pre+"\n") ||
				!strings.Contains(r.out, "== pre-fix leg exit: 0\n== post-fix leg — "+repo+"\n") ||
				!strings.Contains(r.out, "== post-fix leg exit: 1\n") || r.errs != "" {
				t.Fatalf("output does not name both legs:\n%s\nstderr:\n%s", r.out, r.errs)
			}
			prUnchanged(t, r)
		}},
		{"case_2 the pre-fix leg reads not demonstrated", func(t *testing.T) {
			repo, pre := prRepo(t, "repo")
			prReproducer(t, repo, prRel, "exec grep -q old-behaviour check.txt")
			r := prRun(t, repo, Env{}, repo, pre, prRel)
			prWant(t, r, 1, "PROOF FAILED — pre-fix leg read defect not demonstrated: the reproducer does not demonstrate the defect at "+pre+"\n")
			prUnchanged(t, r)
		}},
		{"case_3 the post-fix leg still reads demonstrated", func(t *testing.T) {
			repo, pre := prRepo(t, "repo")
			prReproducer(t, repo, prRel, "exec sh -c 'exit 7'")
			r := prRun(t, repo, Env{}, repo, pre, prRel)
			prWant(t, r, 1, "PROOF FAILED — post-fix leg still reads defect demonstrated: the fix does not stop this reproducer\n")
			prUnchanged(t, r)
		}},
		{"case_4 an unresolvable pre-fix ref is refused before any scratch", func(t *testing.T) {
			repo, _ := prRepo(t, "repo")
			prReproducer(t, repo, prRel, "exec sh -c 'exit 1'")
			r := prRun(t, repo, Env{}, repo, "0123456789abcdef0123456789abcdef01234567", prRel)
			prWant(t, r, 2, "prove-reproducer: pre-fix ref does not resolve to a commit: 0123456789abcdef0123456789abcdef01234567\n")
			prUnchanged(t, r)
		}},
		{"case_5 an absolute reproducer path is refused at the door", func(t *testing.T) {
			repo, pre := prRepo(t, "repo")
			prReproducer(t, repo, prRel, "exec sh -c 'exit 1'")
			r := prRun(t, repo, Env{}, repo, pre, repo+"/"+prRel)
			prWant(t, r, 2, "prove-reproducer: reproducer path must be relative to the worktree, got an absolute path: "+repo+"/"+prRel+"\n")
			prUnchanged(t, r)
		}},
		{"case_6 the mutation-reproducer convention holds in both legs", func(t *testing.T) {
			repo, pre := prRepo(t, "repo")
			rel := ".superpowers/sdd/reproducers/0-mutation-1.sh"
			prReproducer(t, repo, rel, "# mutation-reproducer\nexec grep -q old-behaviour check.txt")
			r := prRun(t, repo, Env{}, repo, pre, rel)
			prWant(t, r, 0, "PROOF HELD")
			prUnchanged(t, r)
		}},
		{"case_7 a .. segment is refused at the door", func(t *testing.T) {
			repo, pre := prRepo(t, "repo")
			r := prRun(t, repo, Env{}, repo, pre, "../escape.sh")
			prWant(t, r, 2, "prove-reproducer: reproducer path may not contain a .. segment: ../escape.sh\n")
			prUnchanged(t, r)
		}},
		{"case_8 a leg the runner cannot verdict spends the proof", func(t *testing.T) {
			repo, pre := prRepo(t, "repo")
			writeFile(t, filepath.Join(repo, prRel), "#!/nonexistent-interpreter\n")
			if err := os.Chmod(filepath.Join(repo, prRel), 0o755); err != nil {
				t.Fatal(err)
			}
			r := prRun(t, repo, Env{}, repo, pre, prRel)
			prWant(t, r, 2, "prove-reproducer: cannot spend the proof — run-reproducer exited 4 on a leg (2 refused, 3 unverifiable, 4 cannot answer)\n")
			if !strings.Contains(r.out, "== pre-fix leg exit: 4\n") || !strings.Contains(r.out, "== post-fix leg exit: 4\n") {
				t.Fatalf("legs not printed:\n%s", r.out)
			}
			prUnchanged(t, r)
		}},
		{"a worktree path with a space proves and cleans up", func(t *testing.T) {
			repo, pre := prRepo(t, "my repo")
			prReproducer(t, repo, prRel, "exec sh -c '! grep -q old-behaviour check.txt'")
			r := prRun(t, repo, Env{}, repo, pre, prRel)
			prWant(t, r, 0, "== post-fix leg — "+repo+"\n")
			prUnchanged(t, r)
		}},
		{"a relative worktree resolves against the working directory", func(t *testing.T) {
			repo, pre := prRepo(t, "repo")
			prReproducer(t, repo, prRel, "exec sh -c '! grep -q old-behaviour check.txt'")
			r := prRun(t, repo, Env{Dir: filepath.Dir(repo)}, "repo", pre, prRel)
			prWant(t, r, 0, "== post-fix leg — repo\n")
			prUnchanged(t, r)
		}},
		{"usage and door refusals print the bash's lines", func(t *testing.T) {
			repo, pre := prRepo(t, "repo")
			writeFile(t, filepath.Join(repo, "noexec.sh"), "#!/bin/sh\n")
			usage := "usage: prove-reproducer.sh <worktree> <pre-fix-ref> <reproducer-path>\n" +
				"       <reproducer-path> is relative to <worktree>; the pre-fix leg runs\n" +
				"       it in a detached scratch worktree at <pre-fix-ref>, the post-fix\n" +
				"       leg against <worktree> itself\n"
			notGit := t.TempDir()
			for _, c := range []struct {
				args []string
				errs string
			}{
				{nil, usage},
				{[]string{repo, pre}, usage},
				{[]string{repo, pre, prRel, "x"}, usage},
				{[]string{"", pre, prRel}, "prove-reproducer: not a directory: \n"},
				{[]string{notGit, pre, prRel}, "prove-reproducer: not a git worktree: " + notGit + "\n"},
				{[]string{repo, "", prRel}, "prove-reproducer: pre-fix ref does not resolve to a commit: \n"},
				{[]string{repo, pre, ""}, "prove-reproducer: reproducer path is empty\n"},
				{[]string{repo, pre, "-x.sh"}, "prove-reproducer: reproducer path may not begin with a dash: -x.sh\n"},
				{[]string{repo, pre, ".."}, "prove-reproducer: reproducer path may not contain a .. segment: ..\n"},
				{[]string{repo, pre, "a/.."}, "prove-reproducer: reproducer path may not contain a .. segment: a/..\n"},
				{[]string{repo, pre, "a/../b"}, "prove-reproducer: reproducer path may not contain a .. segment: a/../b\n"},
				{[]string{repo, pre, "a..b/x.sh"}, "prove-reproducer: no reproducer file at a..b/x.sh in " + repo + "\n"},
				{[]string{repo, pre, "noexec.sh"}, "prove-reproducer: reproducer is not executable: " + repo + "/noexec.sh\n"},
			} {
				r := prRun(t, repo, Env{}, c.args...)
				if r.rc != 2 || r.out != "" || r.errs != c.errs {
					t.Errorf("%q: exit %d\nstdout: %q\nstderr: %q\nwant: %q", c.args, r.rc, r.out, r.errs, c.errs)
				}
				if c.args != nil && c.args[0] == repo {
					prUnchanged(t, r)
				}
			}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			c.run(t)
		})
	}
}

// prBash runs scripts/prove-reproducer.sh as it stood at d71a2327, beside
// its lib/reproducer-path.sh; its run-reproducer.sh is an empty stand-in,
// readable as the bash's door check needs, since no case here reaches a leg.
func prBash(t *testing.T, env Env, args ...string) (int, string, string) {
	t.Helper()
	dir := t.TempDir()
	for _, rel := range []string{"prove-reproducer.sh", "lib/reproducer-path.sh"} {
		src, err := exec.Command(fixtureGit, "-C", "../../..", "show", "d71a2327:scripts/"+rel).Output()
		if err != nil {
			t.Fatalf("git show d71a2327:scripts/%s: %v", rel, err)
		}
		writeFile(t, filepath.Join(dir, rel), string(src))
	}
	writeFile(t, filepath.Join(dir, "run-reproducer.sh"), "")
	cmd := exec.Command("/bin/bash", append([]string{dir + "/prove-reproducer.sh"}, args...)...)
	cmd.Dir = env.Dir
	cmd.Env = append(os.Environ(), "TMPDIR="+env.Getenv("TMPDIR"))
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	_ = cmd.Run()
	return cmd.ProcessState.ExitCode(), out.String(), errb.String()
}

// prTempName is the scratch base's random suffix, mktemp's and MkdirTemp's
// alike, which is all that differs between two runs' messages.
var prTempName = regexp.MustCompile(`prove-reproducer\.[A-Za-z0-9]+`)

// TestProveReproducerBashParity pins the refusals after the scratch base
// exists against the bash at d71a2327, stdout, stderr and exit: the
// worktree add, mkdir -p and cp -p failures with the tools' own
// diagnostics, rev-parse's un-silenced error, and the base under a
// TMPDIR with a trailing slash. Each leaves status, HEAD, the worktree
// list and TMPDIR as they were.
func TestProveReproducerBashParity(t *testing.T) {
	t.Parallel()
	gitIn := func(t *testing.T, repo string, args ...string) string {
		t.Helper()
		cmd := exec.Command(fixtureGit, append([]string{"-C", repo}, args...)...)
		cmd.Env = append(append(os.Environ(), fixtureGitEnv...),
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	cases := []struct {
		name  string
		setup func(t *testing.T, repo, pre string) (ref, rel string)
	}{
		{"worktree add fails", func(t *testing.T, repo, pre string) (string, string) {
			prReproducer(t, repo, prRel, "exit 0")
			wts := filepath.Join(repo, ".git", "worktrees")
			if err := os.MkdirAll(wts, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(wts, 0o555); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(wts, 0o755) })
			return pre, prRel
		}},
		{"mkdir fails on a tracked file", func(t *testing.T, repo, _ string) (string, string) {
			writeFile(t, filepath.Join(repo, "g"), "file\n")
			gitIn(t, repo, "add", "g")
			gitIn(t, repo, "commit", "-q", "-m", "g is a file")
			ref := gitIn(t, repo, "rev-parse", "HEAD")
			gitIn(t, repo, "rm", "-q", "g")
			gitIn(t, repo, "commit", "-q", "-m", "g is a directory")
			prReproducer(t, repo, "g/r.sh", "exit 0")
			return ref, "g/r.sh"
		}},
		{"mkdir of an uncleaned a/. fails on a tracked file", func(t *testing.T, repo, _ string) (string, string) {
			writeFile(t, filepath.Join(repo, "g"), "file\n")
			gitIn(t, repo, "add", "g")
			gitIn(t, repo, "commit", "-q", "-m", "g is a file")
			ref := gitIn(t, repo, "rev-parse", "HEAD")
			gitIn(t, repo, "rm", "-q", "g")
			gitIn(t, repo, "commit", "-q", "-m", "g is a directory")
			prReproducer(t, repo, "g/r.sh", "exit 0")
			return ref, "g/./r.sh"
		}},
		{"cp fails on an unreadable reproducer", func(t *testing.T, repo, pre string) (string, string) {
			prReproducer(t, repo, prRel, "exit 0")
			if err := os.Chmod(filepath.Join(repo, prRel), 0o111); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(filepath.Join(repo, prRel), 0o755) })
			return pre, prRel
		}},
		// rev-parse's own stderr is not silenced: --quiet hides a missing
		// ref, not a ref that names the wrong object type.
		{"a tree ref prints rev-parse's error", func(t *testing.T, repo, _ string) (string, string) {
			prReproducer(t, repo, prRel, "exit 0")
			return "HEAD^{tree}", prRel
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			repo, pre := prRepo(t, "repo")
			ref, rel := c.setup(t, repo, pre)
			tmp := t.TempDir() + "/"
			env := Env{Dir: repo, Getenv: func(k string) string {
				if k == "TMPDIR" {
					return tmp
				}
				return os.Getenv(k)
			}}
			wantRC, wantOut, wantErr := prBash(t, env, repo, ref, rel)
			r := prRun(t, repo, env, repo, ref, rel)
			gotErr, wantErr := prTempName.ReplaceAllString(r.errs, "prove-reproducer.X"), prTempName.ReplaceAllString(wantErr, "prove-reproducer.X")
			if r.rc != wantRC || r.out != wantOut || gotErr != wantErr {
				t.Fatalf("exit %d, want %d\nstdout: %q\nwant:   %q\nstderr: %q\nwant:   %q", r.rc, wantRC, r.out, wantOut, gotErr, wantErr)
			}
			if wantRC != 2 || !strings.Contains(wantErr, "prove-reproducer: ") {
				t.Fatalf("fixture reached no refusal: exit %d\n%s", wantRC, wantErr)
			}
			prUnchanged(t, r)
		})
	}
}

// TestProveReproducerExecBitCheck reaches the check on the scratch copy's
// executable bit -- in production a noexec TMPDIR -- through a `cp` first on
// PATH that copies the content without the mode, and compares the real
// shim with the bash at d71a2327 under that PATH.
func TestProveReproducerExecBitCheck(t *testing.T) {
	t.Parallel()
	repo, pre := prRepo(t, "repo")
	prReproducer(t, repo, prRel, "exit 0")
	stub := t.TempDir()
	writeExec(t, stub+"/cp", "#!/bin/sh\n# cp -p <src> <dst>, the mode dropped\ncat \"$2\" > \"$3\"\n")
	bashDir := bashAtBase(t, "prove-reproducer.sh", "lib/reproducer-path.sh")
	writeFile(t, bashDir+"/run-reproducer.sh", "")
	cache := guardCache(t)
	run := func(script string) (int, string, string) {
		tmp := t.TempDir()
		cmd := exec.Command("/bin/bash", script, repo, pre, prRel)
		cmd.Env = append(os.Environ(), "PATH="+stub+":"+os.Getenv("PATH"), "TMPDIR="+tmp, "FLOW_GUARD_CACHE_DIR="+cache)
		var out, errb bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errb
		_ = cmd.Run()
		norm := strings.NewReplacer(tmp, "<tmp>")
		return cmd.ProcessState.ExitCode(), out.String(), prTempName.ReplaceAllString(norm.Replace(errb.String()), "prove-reproducer.X")
	}
	wantRC, wantOut, wantErr := run(bashDir + "/prove-reproducer.sh")
	rc, out, errs := run(tcfScriptsDir(t) + "/prove-reproducer.sh")
	if !strings.Contains(wantErr, "lost its executable bit") {
		t.Fatalf("fixture did not reach the check: bash exit %d\n%s", wantRC, wantErr)
	}
	if rc != wantRC || out != wantOut || errs != wantErr {
		t.Fatalf("exit %d, want %d\nstdout: %q\nwant:   %q\nstderr: %q\nwant:   %q", rc, wantRC, out, wantOut, errs, wantErr)
	}
}

// TestProveReproducerSIGPIPE runs the real shim with stdout already
// closed, as `| true` leaves it: the verdict write raises SIGPIPE, and the
// guard exits 141, as both bash versions did, with the scratch removed.
func TestProveReproducerSIGPIPE(t *testing.T) {
	t.Parallel()
	repo, pre := prRepo(t, "repo")
	prReproducer(t, repo, prRel, "exec sh -c '! grep -q old-behaviour check.txt'")
	git0 := prGitState(t, repo)
	tmp := t.TempDir()
	cmd := exec.Command("/bin/bash", tcfScriptsDir(t)+"/prove-reproducer.sh", repo, pre, prRel)
	cmd.Env = append(os.Environ(), "TMPDIR="+tmp, "FLOW_GUARD_CACHE_DIR="+guardCache(t))
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	pr.Close()
	cmd.Stdout = pw
	_ = cmd.Run()
	pw.Close()
	if cmd.ProcessState.ExitCode() != 141 {
		t.Fatalf("status %v, want exit 141", cmd.ProcessState)
	}
	if git1 := prGitState(t, repo); git1 != git0 {
		t.Fatalf("git state changed:\nbefore:\n%s\nafter:\n%s", git0, git1)
	}
	if left, _ := filepath.Glob(tmp + "/prove-reproducer.*"); len(left) != 0 {
		t.Fatalf("scratch base survived: %v", left)
	}
}

// TestProveReproducerSignals runs the real shim with a pre-fix leg that
// blocks on a FIFO, and sends a signal once the leg is running. A SIGTERM
// to the guard, and a SIGINT to its process group as a Ctrl-C sends it,
// remove the scratch and the guard dies of the signal, as the bash's EXIT
// trap and death did; a SIGHUP ignored at start stays ignored and the run
// finishes, as a non-interactive bash could not trap it.
func TestProveReproducerSignals(t *testing.T) {
	t.Parallel()
	cache := guardCache(t)
	for _, c := range []struct {
		name   string
		sig    syscall.Signal
		ignore bool
	}{
		{"SIGTERM", syscall.SIGTERM, false},
		{"SIGINT to the process group", syscall.SIGINT, false},
		{"ignored SIGHUP", syscall.SIGHUP, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			repo, pre := prRepo(t, "repo")
			dir := t.TempDir()
			ready, release := dir+"/ready", dir+"/release"
			for _, p := range []string{ready, release} {
				if err := syscall.Mkfifo(p, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			// Only the pre-fix leg, run in the scratch, blocks.
			prReproducer(t, repo, prRel, "case $PWD in */scratch) echo ready > "+ready+"; read -r _ < "+release+";; esac; exit 0")
			git0 := prGitState(t, repo)
			tmp := t.TempDir()
			shim := tcfScriptsDir(t) + "/prove-reproducer.sh"
			cmd := exec.Command("/bin/bash", shim, repo, pre, prRel)
			if c.ignore {
				cmd = exec.Command("/bin/bash", "-c", `trap "" HUP; exec /bin/bash "$0" "$@"`, shim, repo, pre, prRel)
			}
			cmd.Env = append(os.Environ(), "TMPDIR="+tmp, "FLOW_GUARD_CACHE_DIR="+cache)
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			out, errf := mvCreate(t, dir+"/stdout"), mvCreate(t, dir+"/stderr")
			cmd.Stdout, cmd.Stderr = out, errf
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			if _, err := os.ReadFile(ready); err != nil { // blocks until the leg is running
				t.Fatal(err)
			}
			target := cmd.Process.Pid
			if c.sig == syscall.SIGINT {
				target = -target
			}
			if err := syscall.Kill(target, c.sig); err != nil {
				t.Fatal(err)
			}
			if c.ignore {
				// An ignored signal leaves no trace to wait on; half a second
				// is long past a handler's cleanup, which would remove the
				// scratch under the running leg.
				time.Sleep(500 * time.Millisecond)
				if left, _ := filepath.Glob(tmp + "/prove-reproducer.*/scratch"); len(left) != 1 {
					t.Fatalf("the ignored %v removed the scratch under the running leg", c.sig)
				}
			}
			done := make(chan struct{})
			go func() { _ = cmd.Wait(); close(done) }()
			if c.ignore {
				if f, err := os.OpenFile(release, os.O_WRONLY, 0); err == nil {
					f.Close()
				}
				<-done
			} else {
				<-done
				if f, err := os.OpenFile(release, os.O_WRONLY, 0); err == nil { // unblock the orphaned leg
					f.Close()
				}
			}
			ws, _ := cmd.ProcessState.Sys().(syscall.WaitStatus)
			o, _ := os.ReadFile(dir + "/stdout")
			e, _ := os.ReadFile(dir + "/stderr")
			if c.ignore {
				if ws.Signaled() || !strings.Contains(string(o), "== post-fix leg exit: 1\n") {
					t.Fatalf("status %v, want a finished run\nstdout:\n%s\nstderr:\n%s", cmd.ProcessState, o, e)
				}
			} else if !ws.Signaled() || ws.Signal() != c.sig {
				t.Fatalf("status %v, want death by %v\nstdout:\n%s\nstderr:\n%s", cmd.ProcessState, c.sig, o, e)
			}
			if git1 := prGitState(t, repo); git1 != git0 {
				t.Fatalf("git state changed:\nbefore:\n%s\nafter:\n%s", git0, git1)
			}
			if ents, _ := os.ReadDir(tmp); len(ents) != 0 {
				t.Fatalf("temporary files survived: %v", ents)
			}
		})
	}
}
