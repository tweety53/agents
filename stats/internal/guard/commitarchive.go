package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// commitArchive is scripts/commit-archive.sh: that script's header is the
// contract. It makes run 1's drop commit on the change branch
// (skills/flow/integrate.md): asserts spectre/<name>, has
// check-done-when-paths (in-process) refuse a `## Done when` naming a path
// the index does not track, removes spectre/changes/<name>/ and every
// spectre/changes/<name>-fix-<N>/ from the index and the disk, has
// check-archive-scope (in-process) verify the staged diff is deletions under
// those directories alone, and commits `chore(spectre): drop <name>`. The
// name and the ARCHIVE-* tokens stay (decision keep-guard-and-stage-names);
// nothing is archived.
func init() { Registry["commit-archive"] = commitArchive }

func commitArchive(args []string, env Env, stdout, stderr io.Writer) int {
	refuse := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "commit-archive: "+format+"\n", a...)
		return 2
	}
	if len(args) != 2 {
		return refuse("usage: commit-archive.sh <worktree> <name>")
	}
	wt, name := pcAbs(env, args[0]), args[1]
	// The name is spliced into the branch and the dropped paths:
	// plainChangeName carries the rule and its reasoning.
	if !plainChangeName(name) {
		return refuse("change name %q is not a plain change name", name)
	}
	git := envGit(env)
	if !isDir(wt) || git("-C", wt, "rev-parse", "--is-inside-work-tree").Run() != nil {
		return refuse("%s is not a git worktree", wt)
	}

	// The branch is asserted before anything is written, so a refusal leaves
	// the worktree exactly as it was found.
	branch, ok := capture(git("-C", wt, "branch", "--show-current"))
	if !ok {
		return refuse("cannot read the current branch of %s", wt)
	}
	if branch != "spectre/"+name {
		if branch == "" {
			branch = "(detached HEAD)"
		}
		fmt.Fprintf(stdout, "ARCHIVE-WRONG-BRANCH: %s\n", branch)
		return 1
	}

	// The Done-when cross-check runs before anything is written, so a
	// refusal leaves the worktree exactly as it was found. Its OK
	// verdict is not relayed: this guard prints one verdict. Its
	// DONE-WHEN-PATH lines are the verdict when it refuses.
	var dw bytes.Buffer
	switch code := checkDoneWhenPaths([]string{wt}, env, &dw, stderr); code {
	case 0:
	case 1:
		_, _ = stdout.Write(dw.Bytes())
		return 1
	default:
		return refuse("check-done-when-paths.sh could not answer for %s: exit status %d", wt, code)
	}

	// The live directory and every fix sub-change's, each from the index and
	// the disk: what git does not track there (self-review patches, scratch)
	// leaves with it.
	var dirs []string
	if ents, err := os.ReadDir(wt + "/spectre/changes"); err == nil {
		for _, e := range ents {
			n := e.Name()
			if e.IsDir() && (n == name || isFixOf(n, name)) {
				dirs = append(dirs, "spectre/changes/"+n)
			}
		}
	}
	for _, d := range dirs {
		if git("-C", wt, "rm", "-r", "-q", "--ignore-unmatch", "--", d).Run() != nil {
			return refuse("git rm -r %s failed in %s", d, wt)
		}
		if err := os.RemoveAll(wt + "/" + d); err != nil {
			return refuse("cannot remove %s/%s: %v", wt, d, err)
		}
	}
	// The scope guard's SCOPE-OK is not relayed: this guard prints one
	// verdict. Its violation lines are the verdict when it refuses. With no
	// directory left to drop (a re-run), the change's own path stays the
	// one prefix, so anything else staged is still refused.
	scope := append([]string{wt}, dirs...)
	if len(dirs) == 0 {
		scope = append(scope, "spectre/changes/"+name)
	}
	var lines bytes.Buffer
	switch code := checkArchiveScope(scope, env, &lines, stderr); code {
	case 0:
	case 1:
		_, _ = stdout.Write(lines.Bytes())
		return 1
	default:
		return refuse("check-archive-scope.sh could not answer for %s: exit status %d", wt, code)
	}

	err := git("-C", wt, "diff", "--cached", "--quiet").Run()
	var ee *exec.ExitError
	switch {
	case err == nil:
		fmt.Fprintln(stdout, "ARCHIVE-NOTHING-STAGED")
		return 0
	case !errors.As(err, &ee) || ee.ExitCode() != 1:
		return refuse("cannot read the staged diff of %s", wt)
	}
	commit := git("-C", wt, "commit", "-q", "-m", "chore(spectre): drop "+name)
	commit.Stdout, commit.Stderr = stderr, stderr
	if commit.Run() != nil {
		return refuse("git commit failed in %s", wt)
	}
	sha, ok := capture(git("-C", wt, "rev-parse", "HEAD"))
	if !ok {
		return refuse("committed, but cannot read HEAD of %s", wt)
	}
	fmt.Fprintf(stdout, "ARCHIVE-COMMITTED: %s\n", sha)
	return 0
}

// isFixOf is whether dir is a fix sub-change of name: <name>-fix-<digits>.
func isFixOf(dir, name string) bool {
	n, ok := strings.CutPrefix(dir, name+"-fix-")
	return ok && n != "" && strings.Trim(n, "0123456789") == ""
}
