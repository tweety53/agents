package guard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// flow-guard ignore-scratch — keep a session's `.superpowers/` scratch out of
// a worktree's commits (kan-1081 self-review, F2: a project that did not
// ignore the directory tracked its edit scripts, and nothing in worktree
// creation changed that).
//
// Usage: flow-guard ignore-scratch <worktree>
//
// When `git -C <worktree> check-ignore -q --no-index .superpowers/` already
// passes, nothing is written. Otherwise, in order:
//  1. a non-empty index (`git diff --cached --quiet` fails) is refused —
//     the commit below commits the index, so a pre-staged path would ride
//     along;
//  2. `.superpowers/` is appended on its own line to `<worktree>/.gitignore`
//     (created when absent; a newline first when the file does not end in
//     one);
//  3. `git add .gitignore`, then `git rm -r -q --cached --ignore-unmatch
//     .superpowers` — every tracked path under it untracked, the files kept
//     on disk;
//  4. `git commit -q -m "chore: ignore .superpowers/ scratch" -m "Flow-Key:
//     flow.kickoff"` — the index, never a pathspec, since a pathspec commit
//     re-reads the scratch from disk;
//  5. one line on stdout: `ignore-scratch: committed <sha12> — .superpowers/
//     ignored and untracked`.
//
// Exit codes: 0 already ignored, or committed; 1 a git step failed or the
// index was not clean (stderr names which; nothing committed); 2 usage (not
// exactly one argument, or <worktree> is not a directory).
func init() {
	Registry["ignore-scratch"] = ignoreScratch
}

func ignoreScratch(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: flow-guard ignore-scratch <worktree>")
		return 2
	}
	wt := args[0]
	if !filepath.IsAbs(wt) && env.Dir != "" {
		wt = filepath.Join(env.Dir, wt)
	}
	if fi, err := os.Stat(wt); err != nil || !fi.IsDir() {
		fmt.Fprintf(stderr, "ignore-scratch: %s is not a directory\n", args[0])
		return 2
	}
	git := envGit(env)
	run := func(a ...string) (string, bool) {
		out, err := git(append([]string{"-C", wt}, a...)...).CombinedOutput()
		return strings.TrimSpace(string(out)), err == nil
	}
	if _, ok := run("check-ignore", "-q", "--no-index", ".superpowers/"); ok {
		return 0
	}
	fail := func(step, out string) int {
		fmt.Fprintf(stderr, "ignore-scratch: %s failed in %s: %s\n", step, wt, out)
		return 1
	}
	if _, ok := run("diff", "--cached", "--quiet"); !ok {
		fmt.Fprintf(stderr, "ignore-scratch: the staged index in %s is not empty — commit or unstage it first; nothing committed\n", wt)
		return 1
	}
	p := filepath.Join(wt, ".gitignore")
	b, err := os.ReadFile(p)
	if err != nil && !os.IsNotExist(err) {
		return fail("read .gitignore", err.Error())
	}
	if len(b) > 0 && b[len(b)-1] != '\n' {
		b = append(b, '\n')
	}
	if err := os.WriteFile(p, append(b, ".superpowers/\n"...), 0o644); err != nil {
		return fail("write .gitignore", err.Error())
	}
	for _, step := range [][]string{
		{"add", ".gitignore"},
		{"rm", "-r", "-q", "--cached", "--ignore-unmatch", ".superpowers"},
		{"commit", "-q", "-m", "chore: ignore .superpowers/ scratch", "-m", "Flow-Key: flow.kickoff"},
	} {
		if out, ok := run(step...); !ok {
			return fail("git "+strings.Join(step[:1], " "), out)
		}
	}
	sha, ok := run("rev-parse", "--short=12", "HEAD")
	if !ok {
		return fail("git rev-parse", sha)
	}
	fmt.Fprintf(stdout, "ignore-scratch: committed %s — .superpowers/ ignored and untracked\n", sha)
	return 0
}
