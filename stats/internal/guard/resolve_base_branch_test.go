package guard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"testing"
)

// Every case of scripts/test-resolve-base-branch.sh at c5379c0a: one subtest
// per harness case, and inside it one subtest per ok: label -- exit, stdout,
// stderr. The harness asserted a stderr substring; each stderr here is the
// whole text the bash printed at c5379c0a. Fixture repositories are built
// once (rbbMasters) and copied per case. The harness's PATH-shimmed git is
// the same stub on Env's PATH here: the guard resolves git over env's PATH,
// as the bash did.

// rbbFixture is the master trees, built once: a bare remote, a clone of it
// on spectre/fixture with and without refs/remotes/origin/HEAD, and a
// repository with no remote at all -- the harness's new_worktree,
// new_worktree_no_head and new_worktree_no_origin.
type rbbFixture struct{ remote, wt, noHead, noOrigin string }

var rbbMasters = sync.OnceValues(func() (rbbFixture, error) {
	dir, err := os.MkdirTemp(execFixtures.dir, "resolve-base-branch-master")
	if err != nil {
		return rbbFixture{}, err
	}
	f := rbbFixture{dir + "/remote", dir + "/wt", dir + "/nohead", dir + "/noorigin"}
	seed := dir + "/seed"
	git := func(args ...string) error {
		cmd := exec.Command(fixtureGit, args...)
		cmd.Env = append(append(os.Environ(), fixtureGitEnv...),
			"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.invalid",
			"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.invalid")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git %v: %v\n%s", args, err, out)
		}
		return nil
	}
	steps := [][]string{
		{"init", "-q", "-b", "main", "--bare", f.remote},
		{"init", "-q", "-b", "main", seed},
		{"init", "-q", "-b", "spectre/fixture", f.noOrigin},
	}
	for _, s := range steps {
		if err := git(s...); err != nil {
			return f, err
		}
	}
	for _, d := range []string{seed, f.noOrigin} {
		if err := os.WriteFile(d+"/file.txt", []byte("base\n"), 0o644); err != nil {
			return f, err
		}
	}
	steps = [][]string{
		{"-C", seed, "add", "file.txt"},
		{"-C", seed, "commit", "-qm", "base"},
		{"-C", seed, "remote", "add", "origin", f.remote},
		{"-C", seed, "push", "-q", "origin", "main"},
		{"-C", f.noOrigin, "add", "file.txt"},
		{"-C", f.noOrigin, "commit", "-qm", "base"},
	}
	for _, wt := range []string{f.wt, f.noHead} {
		steps = append(steps, []string{"clone", "-q", f.remote, wt},
			[]string{"-C", wt, "remote", "set-head", "origin", "-a"},
			[]string{"-C", wt, "checkout", "-q", "-b", "spectre/fixture"})
	}
	steps = append(steps, []string{"-C", f.noHead, "remote", "set-head", "origin", "-d"})
	for _, s := range steps {
		if err := git(s...); err != nil {
			return f, err
		}
	}
	return f, nil
})

// rbbCopy is one case's own copy of a master tree.
func rbbCopy(t *testing.T, master string) string {
	t.Helper()
	dst := t.TempDir() + "/wt"
	if err := os.CopyFS(dst, os.DirFS(master)); err != nil {
		t.Fatal(err)
	}
	return dst
}

// rbbShim is the harness's shim_git: `fetch --quiet origin` is a no-op (so a
// deleted origin/HEAD survives the guard's own fetch), `remote show origin`
// answers from a `show` file beside the stub when there is one, and every
// other call reaches the real git. It returns the PATH entry holding it.
func rbbShim(t *testing.T, show string) string {
	t.Helper()
	bin := t.TempDir()
	writeExec(t, bin+"/git", `#!/usr/bin/env bash
d="$(dirname -- "$0")"
case "$*" in
  *"fetch --quiet origin"*) exit 0 ;;
  *"remote show origin"*) if [ -f "$d/show" ]; then cat "$d/show"; exit 0; fi ;;
esac
exec `+strconv.Quote(fixtureGit)+` "$@"
`)
	if show != "" {
		writeFile(t, bin+"/show", show)
	}
	return bin
}

func rbbGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command(fixtureGit, append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), fixtureGitEnv...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestResolveBaseBranch(t *testing.T) {
	t.Parallel()
	const invalid = "resolve-base-branch: the resolved base branch name is invalid\n"
	type res struct {
		label   string
		args    []string
		pathDir string // prepended to PATH; "" for none
		deadDir bool   // run from a working directory deleted before the guard runs
		code    int
		out     string // the harness's expected branch; "" for a refusal
		err     string
		skip    string
	}
	cases := []struct {
		name  string
		setup func(t *testing.T, fx rbbFixture) res
	}{
		{"1", func(t *testing.T, fx rbbFixture) res {
			wt := rbbCopy(t, fx.wt)
			return res{label: "origin/HEAD set", args: []string{wt}, out: "main"}
		}},
		{"2", func(t *testing.T, fx rbbFixture) res {
			wt := rbbCopy(t, fx.noHead)
			return res{label: "origin/HEAD deleted, remote show answers", args: []string{wt}, pathDir: rbbShim(t, ""), out: "main"}
		}},
		{"3", func(t *testing.T, fx rbbFixture) res {
			wt := rbbCopy(t, fx.wt)
			rbbGit(t, wt, "checkout", "-q", "--detach", "main")
			return res{label: "detached HEAD", args: []string{wt}, code: 1, err: "resolve-base-branch: HEAD is detached in " + wt + "\n"}
		}},
		{"4", func(t *testing.T, fx rbbFixture) res {
			wt := rbbCopy(t, fx.wt)
			rbbGit(t, wt, "checkout", "-q", "main")
			return res{label: "base equals current branch", args: []string{wt}, code: 1,
				err: "resolve-base-branch: base branch 'main' is the same as the current branch — refusing to compare a branch with itself\n"}
		}},
		{"5", func(t *testing.T, fx rbbFixture) res {
			wt := rbbCopy(t, fx.noHead)
			show := "* remote origin\n  Fetch URL: " + fx.remote + "\n  Remote branch:\n    main tracked\n"
			return res{label: "neither resolution path answers", args: []string{wt}, pathDir: rbbShim(t, show), code: 1,
				err: "resolve-base-branch: could not resolve a base branch in " + wt + "\n"}
		}},
		{"6", func(t *testing.T, fx rbbFixture) res {
			wt := rbbCopy(t, fx.noOrigin)
			return res{label: "no origin remote", args: []string{wt}, code: 3,
				err: "resolve-base-branch: no 'origin' remote configured in " + wt + "\n"}
		}},
		{"7", func(t *testing.T, fx rbbFixture) res {
			dir := t.TempDir()
			return res{label: "plain directory, not a worktree", args: []string{dir}, code: 2,
				err: "resolve-base-branch: " + dir + " is not a git worktree\n"}
		}},
		{"8", func(t *testing.T, fx rbbFixture) res {
			dir := t.TempDir() + "/missing"
			return res{label: "directory does not exist", args: []string{dir}, code: 2,
				err: "resolve-base-branch: " + dir + " is not a directory\n"}
		}},
		{"9", func(t *testing.T, fx rbbFixture) res {
			return res{label: "no argument", code: 2, err: "usage: resolve-base-branch.sh <dir>\n"}
		}},
		{"10", func(t *testing.T, fx rbbFixture) res {
			wt := rbbCopy(t, fx.noHead)
			return res{label: "hostile HEAD branch value: leading dash", args: []string{wt},
				pathDir: rbbShim(t, "* remote origin\n  HEAD branch: -x\n"), code: 1, err: invalid}
		}},
		{"11", func(t *testing.T, fx rbbFixture) res {
			wt := rbbCopy(t, fx.noHead)
			return res{label: "hostile HEAD branch value: control character", args: []string{wt},
				pathDir: rbbShim(t, "* remote origin\n  HEAD branch: bad\x01name\n"), code: 1, err: invalid}
		}},
		// The bash pinned LC_ALL=C because its `case` ranges collated; the
		// port compares bytes, so the caller's locale cannot reach the
		// validation at all, and git's children still get LC_ALL=C.
		{"12", func(t *testing.T, fx rbbFixture) res {
			wt := rbbCopy(t, fx.noHead)
			return res{label: "hostile HEAD branch value: non-ASCII byte under a UTF-8 locale", args: []string{wt},
				pathDir: rbbShim(t, "* remote origin\n  HEAD branch: main\xc3\xa9x\n"), code: 1, err: invalid}
		}},
		// Skipped, as the harness skipped it, where this process can read a
		// mode-000 file or git stores the ref packed.
		{"13", func(t *testing.T, fx rbbFixture) res {
			wt := rbbCopy(t, fx.wt)
			r := res{label: "current branch ref unreadable", args: []string{wt}, code: 2,
				err: "resolve-base-branch: could not read the current branch in " + wt + "\n"}
			ref := wt + "/.git/refs/heads/spectre/fixture"
			if _, err := os.Stat(ref); err != nil {
				r.skip = "this git stores the ref packed rather than loose"
				return r
			}
			if err := os.Chmod(ref, 0); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(ref, 0o644) })
			if f, err := os.Open(ref); err == nil {
				f.Close()
				r.skip = "this process can read a mode-000 file"
			}
			return r
		}},
		// Run from a deleted working directory, git's children inherit it as
		// the bash's did, instead of failing to enter it.
		{"14", func(t *testing.T, fx rbbFixture) res {
			wt := rbbCopy(t, fx.wt)
			return res{label: "run from a deleted working directory", args: []string{wt}, deadDir: true, out: "main"}
		}},
	}
	fn := Registry["resolve-base-branch"]
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if fn == nil {
				t.Fatal("resolve-base-branch is not registered")
			}
			fx, err := rbbMasters()
			if err != nil {
				t.Fatal(err)
			}
			r := c.setup(t, fx)
			if r.skip != "" {
				t.Skip(r.skip)
			}
			env := Env{Dir: t.TempDir(), Getenv: os.Getenv}
			if r.deadDir {
				env.Dir += "/gone"
			}
			if r.pathDir != "" {
				env.Getenv = pathEnv(r.pathDir)
			}
			var out, errb bytes.Buffer
			code := fn(r.args, env, &out, &errb)
			wantOut := ""
			if r.out != "" {
				wantOut = r.out + "\n"
			}
			exitLabel, outLabel, errLabel := r.label+": exit "+strconv.Itoa(r.code), r.label+": stdout empty", r.label+": stderr names the failure"
			if r.code == 0 {
				outLabel, errLabel = r.label+": stdout is '"+r.out+"'", r.label+": stderr empty"
			}
			t.Run(exitLabel, func(t *testing.T) {
				if code != r.code {
					t.Fatalf("exit %d, want %d; stdout %q stderr %q", code, r.code, out.String(), errb.String())
				}
			})
			t.Run(outLabel, func(t *testing.T) {
				if out.String() != wantOut {
					t.Fatalf("stdout %q, want %q", out.String(), wantOut)
				}
			})
			t.Run(errLabel, func(t *testing.T) {
				if errb.String() != r.err {
					t.Fatalf("stderr %q, want %q", errb.String(), r.err)
				}
			})
		})
	}
}
