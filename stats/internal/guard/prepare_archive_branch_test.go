package guard

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-prepare-archive-branch.sh at d71a2327, one
// subtest per ok: label, nested under the harness's own case. The bash
// harness built a bare origin, a seed pushed to it and a clone (the "main
// checkout") per case; here that trio is built once and each case copies it,
// its remote URLs repointed at the copy. The guard runs in-process, except
// case 15, whose PATH-shim git needs a real process: it runs the real shim.
// Each case also carries an "output pinned" subtest: the whole stdout, stderr
// and exit the bash script printed for that fixture at d71a2327, the case's
// directory replaced by <root>. The cases after the harness's 22 are the
// plan's Review Focus rows: tree and HEAD unchanged after every refusal exit
// (asserted inside each refusal case), and a landing path with a space.

const pabArchive = "chore/archive-fixture"

// pabBase is new_checkout, built once: <dir>/remote (bare, `main`),
// <dir>/seed (one commit, pushed) and <dir>/wt (a clone with origin/HEAD set).
func pabBase(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitRun(t, "", "init", "-q", "-b", "main", "--bare", "--template=", dir+"/remote")
	gitRun(t, "", "init", "-q", "-b", "main", "--template=", dir+"/seed")
	gitRun(t, dir+"/seed", "config", "user.email", "test@example.invalid")
	gitRun(t, dir+"/seed", "config", "user.name", "Test")
	writeFile(t, dir+"/seed/file.txt", "base\n")
	gitRun(t, dir+"/seed", "add", "file.txt")
	gitRun(t, dir+"/seed", "commit", "-qm", "base")
	gitRun(t, dir+"/seed", "remote", "add", "origin", dir+"/remote")
	gitRun(t, dir+"/seed", "push", "-q", "origin", "main")
	gitRun(t, "", "clone", "-q", "--template=", dir+"/remote", dir+"/wt")
	gitRun(t, dir+"/wt", "remote", "set-head", "origin", "-a")
	return dir
}

// pabCase is one case's copy of the base trio.
type pabCase struct {
	t                  *testing.T
	root, remote, seed string
	wt                 string
	before             string // pabState before the guard ran
	beforeDir          string
}

// newCheckout copies base under dir, every .git/config's paths repointed.
func pabNewCheckout(t *testing.T, base, dir string) *pabCase {
	t.Helper()
	err := filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := dir + strings.TrimPrefix(p, base)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if d.Name() == "config" {
			b = bytes.ReplaceAll(b, []byte(base), []byte(dir))
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, fi.Mode().Perm())
	})
	if err != nil {
		t.Fatal(err)
	}
	return &pabCase{t: t, root: dir, remote: dir + "/remote", seed: dir + "/seed", wt: dir + "/wt"}
}

// git runs git -C dir with the fixture identity, returning trimmed stdout.
func (c *pabCase) git(dir string, args ...string) string {
	c.t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), append(fixtureGitEnv,
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		c.t.Fatalf("git %v: %v\n%s", args, err, stderr.String())
	}
	return strings.TrimRight(string(out), "\n")
}

// ok reports whether git -C dir args exits 0.
func (c *pabCase) ok(dir string, args ...string) bool {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), fixtureGitEnv...)
	return cmd.Run() == nil
}

// advanceOrigin is advance_origin: origin/main one commit ahead of wt.
func (c *pabCase) advanceOrigin() string {
	writeFile(c.t, c.seed+"/file2.txt", "more\n")
	c.git(c.seed, "add", "file2.txt")
	c.git(c.seed, "commit", "-qm", "more")
	c.git(c.seed, "push", "-q", "origin", "main")
	return c.git(c.seed, "rev-parse", "main")
}

func (c *pabCase) branch(dir string) string {
	cmd := exec.Command("git", "-C", dir, "branch", "--show-current")
	out, _ := cmd.Output()
	return strings.TrimRight(string(out), "\n")
}

// state is the Review Focus row's view of dir: status, HEAD and its branch.
func (c *pabCase) state(dir string) string {
	return c.git(dir, "status", "--porcelain") + "\n" + c.git(dir, "rev-parse", "HEAD") + "\n" + c.branch(dir)
}

// snap records dir's state before a refusal; unchanged asserts it after.
func (c *pabCase) snap(dir string) { c.before, c.beforeDir = c.state(dir), dir }

func (c *pabCase) unchanged(name string) {
	c.t.Helper()
	got := c.state(c.beforeDir)
	gsCheck(c.t, name+": status and HEAD unchanged", got == c.before, "got %q, want %q", got, c.before)
}

// changeWorktree is new_change_worktree: <wt>/.worktrees/fixture on a branch
// `fixture` that commits file.
func (c *pabCase) changeWorktree(file string) {
	apply := c.wt + "/.worktrees/fixture"
	c.git(c.wt, "worktree", "add", "-q", "-b", "fixture", apply)
	writeFile(c.t, apply+"/"+file, "changed\n")
	c.git(apply, "add", file)
	c.git(apply, "commit", "-qm", "the change's edit")
}

// landing is `git worktree add --force --quiet <wt>/.worktrees/_landing-fixture main`.
func (c *pabCase) landing() string {
	l := c.wt + "/.worktrees/_landing-fixture"
	c.git(c.wt, "worktree", "add", "--force", "--quiet", l, "main")
	return l
}

func pabAppend(t *testing.T, path, line string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(line); err != nil {
		t.Fatal(err)
	}
}

type pabRes struct {
	rc       int
	out, err string // as the harness's $(...) read them: trailing newlines stripped
}

// exec runs the guard in-process over args, from the case's directory.
func (c *pabCase) exec(args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	rc := prepareArchiveBranch(args, crEnv(c.root, nil), &out, &errb)
	return rc, out.String(), errb.String()
}

// run is run_guard.
func (c *pabCase) run(args ...string) pabRes {
	rc, out, errs := c.exec(args...)
	return pabRes{rc, strings.TrimRight(out, "\n"), strings.TrimRight(errs, "\n")}
}

// positioned is expect_positioned.
func (c *pabCase) positioned(r pabRes, name, from, to string) {
	t := c.t
	t.Helper()
	gsCheck(t, name+": exit 0", r.rc == 0, "rc=%d out=[%s] err=[%s]", r.rc, r.out, r.err)
	gsCheck(t, name+": stdout names '"+from+"' and '"+to+"'", has(r.out, from, to), "got %q", r.out)
	gsCheck(t, name+": stdout is exactly one line", !strings.Contains(r.out, "\n"), "got %q", r.out)
	gsCheck(t, name+": stderr empty", r.err == "", "got %q", r.err)
}

// refused is expect_refusal.
func (c *pabCase) refused(r pabRes, name string, rc int, sub string) {
	t := c.t
	t.Helper()
	gsCheck(t, name+": exit "+string(rune('0'+rc)), r.rc == rc, "rc=%d out=[%s] err=[%s]", r.rc, r.out, r.err)
	gsCheck(t, name+": stdout empty", r.out == "", "got %q", r.out)
	gsCheck(t, name+": stderr names the failure", strings.Contains(r.err, sub), "want %q in %q", sub, r.err)
}

// pinned asserts the whole result against pabPins (keyed by subtest name).
// The harness's $(...) had stripped trailing newlines; the pins keep them.
func (c *pabCase) pinned(rc int, out, errs string) {
	c.t.Helper()
	phys, err := filepath.EvalSymlinks(c.root)
	if err != nil {
		c.t.Fatal(err)
	}
	norm := strings.NewReplacer(phys, "<root>", c.root, "<root>")
	got := norm.Replace(out) + "--- stderr\n" + norm.Replace(errs) + "--- exit " + string(rune('0'+rc)) + "\n"
	want := pabPins[c.t.Name()]
	gsCheck(c.t, "output pinned", got == want, "got:\n%q\nwant:\n%q", got, want)
}

// runPinned is run plus pinned over the untrimmed streams.
func (c *pabCase) runPinned(args ...string) pabRes {
	c.t.Helper()
	rc, out, errs := c.exec(args...)
	c.pinned(rc, out, errs)
	return pabRes{rc, strings.TrimRight(out, "\n"), strings.TrimRight(errs, "\n")}
}

func TestPrepareArchiveBranch(t *testing.T) {
	t.Parallel()
	base := pabBase(t)
	shimCache := t.TempDir()
	newCheckout := func(t *testing.T) *pabCase { return pabNewCheckout(t, base, t.TempDir()) }
	cases := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{"1 on-base-clean", func(t *testing.T) {
			c := newCheckout(t)
			tip := c.advanceOrigin()
			r := c.runPinned(c.wt, "main", pabArchive)
			c.positioned(r, "on-base-clean", "main", pabArchive)
			gsCheck(t, "on-base-clean: HEAD is on "+pabArchive, c.branch(c.wt) == pabArchive, "on %q", c.branch(c.wt))
			gsCheck(t, "on-base-clean: archive branch carries the fast-forwarded base",
				c.ok(c.wt, "merge-base", "--is-ancestor", tip, "HEAD"), "not an ancestor")
		}},
		{"2 off-base-clean", func(t *testing.T) {
			c := newCheckout(t)
			tip := c.advanceOrigin()
			c.git(c.wt, "checkout", "-q", "-b", "other")
			r := c.runPinned(c.wt, "main", pabArchive)
			c.positioned(r, "off-base-clean", "other", pabArchive)
			gsCheck(t, "off-base-clean: HEAD is on "+pabArchive, c.branch(c.wt) == pabArchive, "on %q", c.branch(c.wt))
			gsCheck(t, "off-base-clean: archive branch carries the fast-forwarded base",
				c.ok(c.wt, "merge-base", "--is-ancestor", tip, "HEAD"), "not an ancestor")
		}},
		{"3 off-base-dirty", func(t *testing.T) {
			c := newCheckout(t)
			c.git(c.wt, "checkout", "-q", "-b", "other")
			pabAppend(t, c.wt+"/file.txt", "dirty\n")
			c.snap(c.wt)
			r := c.runPinned(c.wt, "main", pabArchive)
			c.refused(r, "off-base-dirty", 1, "other")
			gsCheck(t, "off-base-dirty: stderr names 'main'", strings.Contains(r.err, "main"), "got %q", r.err)
			gsCheck(t, "off-base-dirty: HEAD unchanged", c.branch(c.wt) == "other", "on %q", c.branch(c.wt))
			gsCheck(t, "off-base-dirty: nothing staged", c.git(c.wt, "diff", "--cached", "--name-only") == "", "staged")
			c.unchanged("off-base-dirty")
		}},
		{"4 on-base-dirty", func(t *testing.T) {
			c := newCheckout(t)
			pabAppend(t, c.wt+"/file.txt", "dirty\n")
			c.snap(c.wt)
			r := c.runPinned(c.wt, "main", pabArchive)
			c.refused(r, "on-base-dirty", 1, "dirty")
			gsCheck(t, "on-base-dirty: HEAD unchanged", c.branch(c.wt) == "main", "on %q", c.branch(c.wt))
			c.unchanged("on-base-dirty")
		}},
		{"5 detached-head", func(t *testing.T) {
			c := newCheckout(t)
			c.git(c.wt, "checkout", "-q", "--detach", "main")
			sha := c.git(c.wt, "rev-parse", "HEAD")
			c.snap(c.wt)
			r := c.runPinned(c.wt, "main", pabArchive)
			c.refused(r, "detached-head", 1, "detached")
			gsCheck(t, "detached-head: HEAD stays detached", c.branch(c.wt) == "", "on %q", c.branch(c.wt))
			gsCheck(t, "detached-head: HEAD sha unchanged", c.git(c.wt, "rev-parse", "HEAD") == sha, "moved")
			c.unchanged("detached-head")
		}},
		{"6 archive-branch-exists-descended", func(t *testing.T) {
			c := newCheckout(t)
			c.git(c.wt, "checkout", "-q", "-b", pabArchive, "main")
			writeFile(t, c.wt+"/extra.txt", "extra\n")
			c.git(c.wt, "add", "extra.txt")
			c.git(c.wt, "commit", "-qm", "extra")
			existing := c.git(c.wt, "rev-parse", pabArchive)
			c.git(c.wt, "checkout", "-q", "main")
			r := c.runPinned(c.wt, "main", pabArchive)
			c.positioned(r, "archive-branch-exists-descended", "main", pabArchive)
			gsCheck(t, "archive-branch-exists-descended: HEAD is on "+pabArchive, c.branch(c.wt) == pabArchive, "on %q", c.branch(c.wt))
			gsCheck(t, "archive-branch-exists-descended: existing commit survives, not recreated",
				c.git(c.wt, "rev-parse", "HEAD") == existing, "recreated")
		}},
		{"7 archive-branch-exists-unrelated", func(t *testing.T) {
			c := newCheckout(t)
			c.git(c.wt, "checkout", "-q", "--orphan", pabArchive)
			c.git(c.wt, "rm", "-rf", "-q", "--cached", ".")
			c.git(c.wt, "clean", "-fdq")
			writeFile(t, c.wt+"/unrelated.txt", "unrelated\n")
			c.git(c.wt, "add", "unrelated.txt")
			c.git(c.wt, "commit", "-qm", "unrelated")
			c.git(c.wt, "checkout", "-q", "main")
			c.snap(c.wt)
			r := c.runPinned(c.wt, "main", pabArchive)
			c.refused(r, "archive-branch-exists-unrelated", 1, "descended")
			gsCheck(t, "archive-branch-exists-unrelated: HEAD unchanged", c.branch(c.wt) == "main", "on %q", c.branch(c.wt))
			c.unchanged("archive-branch-exists-unrelated")
		}},
		{"8 not-a-worktree", func(t *testing.T) {
			c := &pabCase{t: t, root: t.TempDir()}
			r := c.runPinned(c.root, "main", pabArchive)
			c.refused(r, "not-a-worktree", 2, "not a git worktree")
		}},
		{"9 missing-checkout", func(t *testing.T) {
			c := &pabCase{t: t, root: t.TempDir()}
			r := c.runPinned(c.root+"/prepare-archive-branch-test-missing", "main", pabArchive)
			c.refused(r, "missing-checkout", 1, "not a .worktrees directory")
			gsCheck(t, "missing-checkout: nothing created", !isDir(c.root+"/prepare-archive-branch-test-missing"), "created")
		}},
		{"10 base-diverged", func(t *testing.T) {
			c := newCheckout(t)
			writeFile(t, c.wt+"/local.txt", "local\n")
			c.git(c.wt, "add", "local.txt")
			c.git(c.wt, "commit", "-qm", "local")
			local := c.git(c.wt, "rev-parse", "main")
			writeFile(t, c.seed+"/remote.txt", "remote\n")
			c.git(c.seed, "add", "remote.txt")
			c.git(c.seed, "commit", "-qm", "remote")
			c.git(c.seed, "push", "-q", "origin", "main")
			c.snap(c.wt)
			r := c.runPinned(c.wt, "main", pabArchive)
			c.refused(r, "base-diverged", 3, "fast-forward")
			gsCheck(t, "base-diverged: HEAD unchanged", c.branch(c.wt) == "main", "on %q", c.branch(c.wt))
			gsCheck(t, "base-diverged: local base commit unchanged", c.git(c.wt, "rev-parse", "main") == local, "moved")
			c.unchanged("base-diverged")
		}},
		{"11 no-origin", func(t *testing.T) {
			c := &pabCase{t: t, root: t.TempDir()}
			wt := c.root + "/noorigin"
			gitRun(t, "", "init", "-q", "-b", "main", "--template=", wt)
			writeFile(t, wt+"/file.txt", "base\n")
			c.git(wt, "add", "file.txt")
			c.git(wt, "commit", "-qm", "base")
			c.snap(wt)
			r := c.runPinned(wt, "main", pabArchive)
			c.refused(r, "no-origin", 3, "origin")
			c.unchanged("no-origin")
		}},
		{"12 creates-landing-worktree", func(t *testing.T) {
			c := newCheckout(t)
			tip := c.advanceOrigin()
			mainBefore := c.branch(c.wt)
			l := c.wt + "/.worktrees/_landing-fixture"
			r := c.runPinned(l, "main", pabArchive)
			c.positioned(r, "creates-landing-worktree", "main", pabArchive)
			gsCheck(t, "creates-landing-worktree: the landing worktree now exists", isDir(l), "absent")
			gsCheck(t, "creates-landing-worktree: landing worktree HEAD is on "+pabArchive, c.branch(l) == pabArchive, "on %q", c.branch(l))
			gsCheck(t, "creates-landing-worktree: archive branch carries the fast-forwarded base",
				c.ok(l, "merge-base", "--is-ancestor", tip, "HEAD"), "not an ancestor")
			gsCheck(t, "creates-landing-worktree: main checkout branch unchanged", c.branch(c.wt) == mainBefore, "on %q", c.branch(c.wt))
		}},
		{"13 dirty-landing-worktree", func(t *testing.T) {
			c := newCheckout(t)
			l := c.landing()
			pabAppend(t, l+"/file.txt", "dirty\n")
			mainBefore := c.branch(c.wt)
			c.snap(l)
			r := c.runPinned(l, "main", pabArchive)
			c.refused(r, "dirty-landing-worktree", 1, "dirty")
			gsCheck(t, "dirty-landing-worktree: main checkout branch unchanged", c.branch(c.wt) == mainBefore, "on %q", c.branch(c.wt))
			c.unchanged("dirty-landing-worktree")
		}},
		{"14 main-checkout-never-checked-out", func(t *testing.T) {
			c := newCheckout(t)
			before := c.branch(c.wt)
			r := c.runPinned(c.wt+"/.worktrees/_landing-fixture", "main", pabArchive)
			gsCheck(t, "main-checkout-never-checked-out: exit 0", r.rc == 0, "rc=%d out=[%s] err=[%s]", r.rc, r.out, r.err)
			gsCheck(t, "main-checkout-never-checked-out: branch --show-current in the main checkout is unchanged ("+before+")",
				c.branch(c.wt) == before, "changed from %q to %q", before, c.branch(c.wt))
		}},
		// 15. A PATH-shim git plants a real stash entry before forwarding the
		//     second `stash list` (the post-run recompute): the real shim, since
		//     only a real process reads PATH.
		{"15 stash-mid-run", func(t *testing.T) {
			c := newCheckout(t)
			l := c.landing()
			shim := t.TempDir()
			realGit, err := exec.LookPath("git")
			if err != nil {
				t.Fatal(err)
			}
			writeExec(t, shim+"/git", pabGitShim)
			cmd := exec.Command("/bin/bash", tcfScriptsDir(t)+"/prepare-archive-branch.sh", l, "main", pabArchive)
			cmd.Env = append(os.Environ(), "PAB_REAL="+realGit, "PAB_COUNT="+shim+"/count", "PAB_LANDING="+l,
				"PATH="+shim+":"+os.Getenv("PATH"), "FLOW_GUARD_CACHE_DIR="+shimCache)
			var out, errb bytes.Buffer
			cmd.Stdout, cmd.Stderr = &out, &errb
			_ = cmd.Run()
			rc := cmd.ProcessState.ExitCode()
			gsCheck(t, "stash appearing mid-run exits 2 naming the stash",
				rc == 2 && has(errb.String(), "new stash entry: ", "kan-448-injected"), "rc=%d err=%s", rc, errb.String())
			c.pinned(rc, out.String(), errb.String())
		}},
		{"16 dirty-files-named-and-classified-as-change-output", func(t *testing.T) {
			c := newCheckout(t)
			c.changeWorktree("file.txt")
			l := c.landing()
			pabAppend(t, l+"/file.txt", "dirty\n")
			c.snap(l)
			r := c.runPinned(l, "main", pabArchive)
			c.refused(r, "dirty-files-named-and-classified-as-change-output", 1, "dirty")
			gsCheck(t, "dirty-files-named-and-classified-as-change-output: file.txt named as change output",
				strings.Contains(r.err, "file.txt -- looks like this change's output"), "got %q", r.err)
			c.unchanged("dirty-files-named-and-classified-as-change-output")
		}},
		{"17 dirty-files-named-and-classified-as-not-change-output", func(t *testing.T) {
			c := newCheckout(t)
			c.changeWorktree("feature.txt")
			l := c.landing()
			pabAppend(t, l+"/file.txt", "dirty\n")
			c.snap(l)
			r := c.runPinned(l, "main", pabArchive)
			c.refused(r, "dirty-files-named-and-classified-as-not-change-output", 1, "dirty")
			gsCheck(t, "dirty-files-named-and-classified-as-not-change-output: file.txt named as not change output",
				strings.Contains(r.err, "file.txt -- does not look like this change's output"), "got %q", r.err)
			c.unchanged("dirty-files-named-and-classified-as-not-change-output")
		}},
		{"18 dirty-files-named-without-a-change-worktree", func(t *testing.T) {
			c := newCheckout(t)
			l := c.landing()
			pabAppend(t, l+"/file.txt", "dirty\n")
			c.snap(l)
			r := c.runPinned(l, "main", pabArchive)
			c.refused(r, "dirty-files-named-without-a-change-worktree", 1, "dirty")
			gsCheck(t, "dirty-files-named-without-a-change-worktree: file.txt still named", strings.Contains(r.err, "file.txt"), "got %q", r.err)
			gsCheck(t, "dirty-files-named-without-a-change-worktree: classification said unavailable",
				strings.Contains(r.err, "cannot classify"), "got %q", r.err)
			c.unchanged("dirty-files-named-without-a-change-worktree")
		}},
		{"19 merge-base-failure-classifies-nothing", func(t *testing.T) {
			c := newCheckout(t)
			apply := c.wt + "/.worktrees/fixture"
			c.git(c.wt, "worktree", "add", "-q", "-b", "tmp-root", apply)
			c.git(apply, "checkout", "-q", "--orphan", "fixture")
			writeFile(t, apply+"/feature.txt", "changed\n")
			c.git(apply, "add", "feature.txt")
			c.git(apply, "commit", "-qm", "orphan root")
			c.git(apply, "branch", "-D", "tmp-root")
			l := c.landing()
			pabAppend(t, l+"/file.txt", "dirty\n")
			c.snap(l)
			r := c.runPinned(l, "main", pabArchive)
			c.refused(r, "merge-base-failure-classifies-nothing", 1, "dirty")
			gsCheck(t, "merge-base-failure-classifies-nothing: file.txt still named", strings.Contains(r.err, "file.txt"), "got %q", r.err)
			gsCheck(t, "merge-base-failure-classifies-nothing: classification said unavailable",
				strings.Contains(r.err, "cannot classify"), "got %q", r.err)
			gsCheck(t, "merge-base-failure-classifies-nothing: no definitive negative claim",
				!strings.Contains(r.err, "does not look like this change's output"), "got %q", r.err)
			c.unchanged("merge-base-failure-classifies-nothing")
		}},
		{"20 glob-metacharacter-filenames-match-literally", func(t *testing.T) {
			c := newCheckout(t)
			c.changeWorktree("data1.txt")
			l := c.landing()
			writeFile(t, l+"/data*.txt", "stray\n")
			c.snap(l)
			r := c.runPinned(l, "main", pabArchive)
			c.refused(r, "glob-metacharacter-filenames-match-literally", 1, "dirty")
			gsCheck(t, "glob-metacharacter-filenames-match-literally: data*.txt named as not change output",
				strings.Contains(r.err, "data*.txt -- does not look like this change's output"), "got %q", r.err)
			c.unchanged("glob-metacharacter-filenames-match-literally")
		}},
		{"21 non-ascii-filenames-classify-unquoted", func(t *testing.T) {
			c := newCheckout(t)
			writeFile(t, c.wt+"/café.txt", "base")
			c.git(c.wt, "add", "café.txt")
			c.git(c.wt, "commit", "-qm", "seed non-ascii")
			c.changeWorktree("café.txt")
			l := c.landing()
			pabAppend(t, l+"/café.txt", "dirty\n")
			c.snap(l)
			r := c.runPinned(l, "main", pabArchive)
			c.refused(r, "non-ascii-filenames-classify-unquoted", 1, "dirty")
			gsCheck(t, "non-ascii-filenames-classify-unquoted: café.txt named as change output, unquoted",
				strings.Contains(r.err, "café.txt -- looks like this change's output"), "got %q", r.err)
			c.unchanged("non-ascii-filenames-classify-unquoted")
		}},
		{"22 off-base-refusal-classifies-with-a-sibling", func(t *testing.T) {
			c := newCheckout(t)
			c.changeWorktree("file.txt")
			l := c.landing()
			c.git(l, "checkout", "-q", "-b", "other")
			pabAppend(t, l+"/file.txt", "dirty\n")
			c.snap(l)
			r := c.runPinned(l, "main", pabArchive)
			c.refused(r, "off-base-refusal-classifies-with-a-sibling", 1, "other")
			gsCheck(t, "off-base-refusal-classifies-with-a-sibling: file.txt classified on the off-base refusal",
				strings.Contains(r.err, "file.txt -- looks like this change's output"), "got %q", r.err)
			c.unchanged("off-base-refusal-classifies-with-a-sibling")
		}},
		// Review Focus: a landing path with a space, created and positioned.
		{"landing path with a space", func(t *testing.T) {
			c := pabNewCheckout(t, base, t.TempDir()+"/a dir")
			tip := c.advanceOrigin()
			l := c.wt + "/.worktrees/_landing-fixture"
			r := c.runPinned(l, "main", pabArchive)
			c.positioned(r, "landing path with a space", "main", pabArchive)
			gsCheck(t, "landing path with a space: landing worktree HEAD is on "+pabArchive, c.branch(l) == pabArchive, "on %q", c.branch(l))
			gsCheck(t, "landing path with a space: archive branch carries the fast-forwarded base",
				c.ok(l, "merge-base", "--is-ancestor", tip, "HEAD"), "not an ancestor")
		}},
		{"usage", func(t *testing.T) {
			c := &pabCase{t: t, root: t.TempDir()}
			r := c.runPinned(c.root, "main")
			gsCheck(t, "usage: a missing argument exits 2", r.rc == 2, "rc=%d", r.rc)
		}},
		{"invalid branch names", func(t *testing.T) {
			c := newCheckout(t)
			c.snap(c.wt)
			r := c.runPinned(c.wt, "-main", pabArchive)
			c.refused(r, "invalid base", 1, "base branch '-main' is not a valid branch name")
			r = c.run(c.wt, "main", "chore/a b")
			c.refused(r, "invalid archive branch", 1, "archive branch 'chore/a b' is not a valid branch name")
			c.unchanged("invalid branch names")
		}},
		// Review Focus: each refusal that fires after the guard has moved
		// HEAD, and the classification mechanisms the bash got from its
		// expansions and `export LC_ALL=C`, each run beside the bash at
		// d71a2327 on its own copy: stdout, stderr, exit, and the status and
		// branch it left.
		{"port: origin/<base> missing, reached off base", func(t *testing.T) {
			pabParity(t, base, 3, func(c *pabCase) ([]string, string) {
				c.git(c.wt, "checkout", "-q", "-b", "other")
				c.git(c.wt, "config", "--unset-all", "remote.origin.fetch")
				c.git(c.wt, "update-ref", "-d", "refs/remotes/origin/main")
				return []string{c.wt, "main", pabArchive}, c.wt
			})
		}},
		{"port: base diverged, reached off base", func(t *testing.T) {
			pabParity(t, base, 3, func(c *pabCase) ([]string, string) {
				writeFile(t, c.wt+"/local.txt", "local\n")
				c.git(c.wt, "add", "local.txt")
				c.git(c.wt, "commit", "-qm", "local")
				writeFile(t, c.seed+"/remote.txt", "remote\n")
				c.git(c.seed, "add", "remote.txt")
				c.git(c.seed, "commit", "-qm", "remote")
				c.git(c.seed, "push", "-q", "origin", "main")
				c.git(c.wt, "checkout", "-q", "-b", "other")
				return []string{c.wt, "main", pabArchive}, c.wt
			})
		}},
		{"port: base cannot be checked out", func(t *testing.T) {
			pabParity(t, base, 2, func(c *pabCase) ([]string, string) {
				c.git(c.wt, "checkout", "-q", "-b", "other")
				return []string{c.wt, ".x", pabArchive}, c.wt
			})
		}},
		{"port: existing archive branch checked out elsewhere", func(t *testing.T) {
			pabParity(t, base, 2, func(c *pabCase) ([]string, string) {
				c.git(c.wt, "branch", pabArchive, "main")
				c.git(c.wt, "worktree", "add", "-q", c.root+"/elsewhere", pabArchive)
				c.git(c.wt, "checkout", "-q", "-b", "other")
				return []string{c.wt, "main", pabArchive}, c.wt
			})
		}},
		{"port: archive branch cannot be created", func(t *testing.T) {
			pabParity(t, base, 2, func(c *pabCase) ([]string, string) {
				c.git(c.wt, "checkout", "-q", "-b", "other")
				return []string{c.wt, "main", "a..b"}, c.wt
			})
		}},
		{"port: a rename is classified by its new path", func(t *testing.T) {
			pabParity(t, base, 1, func(c *pabCase) ([]string, string) {
				c.changeWorktree("renamed.txt")
				l := c.landing()
				c.git(l, "mv", "file.txt", "renamed.txt")
				return []string{l, "main", pabArchive}, l
			})
		}},
		{"port: an untracked directory entry matches the paths under it", func(t *testing.T) {
			pabParity(t, base, 1, func(c *pabCase) ([]string, string) {
				c.changeWorktree("newdir/a.txt")
				l := c.landing()
				writeFile(t, l+"/newdir/a.txt", "stray\n")
				return []string{l, "main", pabArchive}, l
			})
		}},
		// `lnk/..` is the symlink target's parent, as the kernel resolves
		// it, not the directory holding lnk: here the landing exists only
		// under the lexical reading.
		{"port: a landing path through a symlink's .. resolves physically", func(t *testing.T) {
			pabParity(t, base, 2, func(c *pabCase) ([]string, string) {
				c.landing()
				mkdir(t, c.root+"/deep/sub")
				if err := os.Symlink(c.root+"/deep/sub", c.root+"/lnk"); err != nil {
					t.Fatal(err)
				}
				return []string{"lnk/../wt/.worktrees/_landing-fixture", "main", pabArchive}, c.wt
			})
		}},
		{"port: a non-ASCII branch name is refused byte by byte", func(t *testing.T) {
			pabParity(t, base, 1, func(c *pabCase) ([]string, string) {
				return []string{c.wt, "é", pabArchive}, c.wt
			})
		}},
		{"basename(1) parity", func(t *testing.T) {
			for _, in := range []string{"/", "//", "/a", "/a/", "a", "a/b", "a//b/", "/a/b", "_landing-x/"} {
				want, err := exec.Command("basename", in).Output()
				if err != nil {
					t.Fatal(err)
				}
				gsCheck(t, "basename "+in, pabBasename(in) == strings.TrimSuffix(string(want), "\n"),
					"pabBasename(%q) = %q, basename(1) = %q", in, pabBasename(in), want)
			}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			c.fn(t)
		})
	}
}

// pabParity runs the bash at d71a2327 and the port on separate copies of
// the base trio, each prepared by setup, and asserts the same stdout,
// stderr and exit (want, so the fixture reaches the refusal it names) and
// the same status and branch left in the directory setup names.
func pabParity(t *testing.T, base string, want int, setup func(c *pabCase) ([]string, string)) {
	t.Helper()
	type res struct{ rc, out, errs, state string }
	run := func(bash bool) res {
		c := pabNewCheckout(t, base, t.TempDir())
		args, dir := setup(c)
		var rc int
		var out, errs string
		if bash {
			cmd := exec.Command("bash", append([]string{bashAtBase(t, "prepare-archive-branch.sh", "lib/post-mutation-check.sh") + "/prepare-archive-branch.sh"}, args...)...)
			cmd.Dir = c.root
			var o, e bytes.Buffer
			cmd.Stdout, cmd.Stderr = &o, &e
			_ = cmd.Run()
			rc, out, errs = cmd.ProcessState.ExitCode(), o.String(), e.String()
		} else {
			rc, out, errs = c.exec(args...)
		}
		phys, err := filepath.EvalSymlinks(c.root)
		if err != nil {
			t.Fatal(err)
		}
		norm := strings.NewReplacer(phys, "<root>", c.root, "<root>")
		state := c.git(dir, "status", "--porcelain") + "\n" + c.branch(dir)
		return res{string(rune('0' + rc)), norm.Replace(out), norm.Replace(errs), state}
	}
	w, g := run(true), run(false)
	gsCheck(t, "the bash exits "+string(rune('0'+want)), w.rc == string(rune('0'+want)), "bash: %+v", w)
	gsCheck(t, "matches the bash at d71a2327", g == w, "got:\n%+v\nwant:\n%+v", g, w)
}

// pabGitShim is case 15's PATH-shim git: it forwards every call to the real
// git and, before forwarding the second `stash list` (the post-run check's
// recompute; the first is the snapshot's), plants one real stash entry in the
// landing worktree.
const pabGitShim = `#!/usr/bin/env bash
is_stash_list=0
prev=""
for a in "$@"; do
  [ "$prev" = "stash" ] && [ "$a" = "list" ] && is_stash_list=1
  prev="$a"
done
if [ "$is_stash_list" -eq 1 ]; then
  n="$(cat "$PAB_COUNT" 2>/dev/null || echo 0)"
  n=$((n + 1))
  printf '%s\n' "$n" > "$PAB_COUNT"
  if [ "$n" = "2" ]; then
    printf stray > "$PAB_LANDING/stray.txt"
    "$PAB_REAL" -C "$PAB_LANDING" stash push -q --include-untracked -m kan-448-injected
  fi
fi
exec "$PAB_REAL" "$@"
`

// pabPins is each fixture's whole result from the bash script at d71a2327,
// captured by running scripts/prepare-archive-branch.sh over the same fixture.
var pabPins = map[string]string{
	"TestPrepareArchiveBranch/1_on-base-clean":                                          "main -> chore/archive-fixture\n--- stderr\n--- exit 0\n",
	"TestPrepareArchiveBranch/10_base-diverged":                                         "--- stderr\nprepare-archive-branch: 'main' cannot be fast-forwarded to origin/main — it has diverged\n--- exit 3\n",
	"TestPrepareArchiveBranch/11_no-origin":                                             "--- stderr\nprepare-archive-branch: no 'origin' remote configured in <root>/noorigin — cannot resolve origin/main\n--- exit 3\n",
	"TestPrepareArchiveBranch/12_creates-landing-worktree":                              "main -> chore/archive-fixture\n--- stderr\n--- exit 0\n",
	"TestPrepareArchiveBranch/13_dirty-landing-worktree":                                "--- stderr\nprepare-archive-branch: <root>/wt/.worktrees/_landing-fixture has a dirty working tree on 'main' — refusing\nprepare-archive-branch: dirty files:\nprepare-archive-branch:   (cannot classify -- no change worktree with a branch beside <root>/wt/.worktrees/_landing-fixture)\nprepare-archive-branch:    M file.txt\n--- exit 1\n",
	"TestPrepareArchiveBranch/14_main-checkout-never-checked-out":                       "main -> chore/archive-fixture\n--- stderr\n--- exit 0\n",
	"TestPrepareArchiveBranch/15_stash-mid-run":                                         "--- stderr\nprepare-archive-branch: post-run drift detected:\nnew stash entry: stash@{0}: On chore/archive-fixture: kan-448-injected\n--- exit 2\n",
	"TestPrepareArchiveBranch/16_dirty-files-named-and-classified-as-change-output":     "--- stderr\nprepare-archive-branch: <root>/wt/.worktrees/_landing-fixture has a dirty working tree on 'main' — refusing\nprepare-archive-branch: dirty files:\nprepare-archive-branch:    M file.txt -- looks like this change's output (changed on 'fixture')\n--- exit 1\n",
	"TestPrepareArchiveBranch/17_dirty-files-named-and-classified-as-not-change-output": "--- stderr\nprepare-archive-branch: <root>/wt/.worktrees/_landing-fixture has a dirty working tree on 'main' — refusing\nprepare-archive-branch: dirty files:\nprepare-archive-branch:    M file.txt -- does not look like this change's output (not changed on 'fixture')\n--- exit 1\n",
	"TestPrepareArchiveBranch/18_dirty-files-named-without-a-change-worktree":           "--- stderr\nprepare-archive-branch: <root>/wt/.worktrees/_landing-fixture has a dirty working tree on 'main' — refusing\nprepare-archive-branch: dirty files:\nprepare-archive-branch:   (cannot classify -- no change worktree with a branch beside <root>/wt/.worktrees/_landing-fixture)\nprepare-archive-branch:    M file.txt\n--- exit 1\n",
	"TestPrepareArchiveBranch/19_merge-base-failure-classifies-nothing":                 "--- stderr\nprepare-archive-branch: <root>/wt/.worktrees/_landing-fixture has a dirty working tree on 'main' — refusing\nprepare-archive-branch: dirty files:\nprepare-archive-branch:   (cannot classify -- no change worktree with a branch beside <root>/wt/.worktrees/_landing-fixture)\nprepare-archive-branch:    M file.txt\n--- exit 1\n",
	"TestPrepareArchiveBranch/2_off-base-clean":                                         "other -> chore/archive-fixture\n--- stderr\n--- exit 0\n",
	"TestPrepareArchiveBranch/20_glob-metacharacter-filenames-match-literally":          "--- stderr\nprepare-archive-branch: <root>/wt/.worktrees/_landing-fixture has a dirty working tree on 'main' — refusing\nprepare-archive-branch: dirty files:\nprepare-archive-branch:   ?? data*.txt -- does not look like this change's output (not changed on 'fixture')\n--- exit 1\n",
	"TestPrepareArchiveBranch/21_non-ascii-filenames-classify-unquoted":                 "--- stderr\nprepare-archive-branch: <root>/wt/.worktrees/_landing-fixture has a dirty working tree on 'main' — refusing\nprepare-archive-branch: dirty files:\nprepare-archive-branch:    M café.txt -- looks like this change's output (changed on 'fixture')\n--- exit 1\n",
	"TestPrepareArchiveBranch/22_off-base-refusal-classifies-with-a-sibling":            "--- stderr\nprepare-archive-branch: <root>/wt/.worktrees/_landing-fixture is on 'other' with uncommitted changes, not 'main' — refusing\nprepare-archive-branch: dirty files:\nprepare-archive-branch:    M file.txt -- looks like this change's output (changed on 'fixture')\n--- exit 1\n",
	"TestPrepareArchiveBranch/3_off-base-dirty":                                         "--- stderr\nprepare-archive-branch: <root>/wt is on 'other' with uncommitted changes, not 'main' — refusing\nprepare-archive-branch: dirty files:\nprepare-archive-branch:   (cannot classify -- no change worktree with a branch beside <root>/wt)\nprepare-archive-branch:    M file.txt\n--- exit 1\n",
	"TestPrepareArchiveBranch/4_on-base-dirty":                                          "--- stderr\nprepare-archive-branch: <root>/wt has a dirty working tree on 'main' — refusing\nprepare-archive-branch: dirty files:\nprepare-archive-branch:   (cannot classify -- no change worktree with a branch beside <root>/wt)\nprepare-archive-branch:    M file.txt\n--- exit 1\n",
	"TestPrepareArchiveBranch/5_detached-head":                                          "--- stderr\nprepare-archive-branch: HEAD is detached in <root>/wt\n--- exit 1\n",
	"TestPrepareArchiveBranch/6_archive-branch-exists-descended":                        "main -> chore/archive-fixture\n--- stderr\n--- exit 0\n",
	"TestPrepareArchiveBranch/7_archive-branch-exists-unrelated":                        "--- stderr\nprepare-archive-branch: 'chore/archive-fixture' already exists and is not descended from origin/main — refusing\n--- exit 1\n",
	"TestPrepareArchiveBranch/8_not-a-worktree":                                         "--- stderr\nprepare-archive-branch: <root> is not a git worktree\n--- exit 2\n",
	"TestPrepareArchiveBranch/9_missing-checkout":                                       "--- stderr\nprepare-archive-branch: <root>/prepare-archive-branch-test-missing's parent is not a .worktrees directory — refusing to create it\n--- exit 1\n",
	"TestPrepareArchiveBranch/invalid_branch_names":                                     "--- stderr\nprepare-archive-branch: base branch '-main' is not a valid branch name\n--- exit 1\n",
	"TestPrepareArchiveBranch/landing_path_with_a_space":                                "main -> chore/archive-fixture\n--- stderr\n--- exit 0\n",
	"TestPrepareArchiveBranch/usage":                                                    "--- stderr\nusage: prepare-archive-branch.sh <landing-worktree> <base> <archive-branch>\n--- exit 2\n",
}
