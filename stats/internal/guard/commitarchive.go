package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// commitArchive is scripts/commit-archive.sh: that script's header is the
// contract. It makes run 1's archive commit on the change branch
// (skills/flow/integrate.md): asserts spectre/<name>, has
// check-done-when-paths (in-process)
// refuse a `## Done when` naming a path the index does not track, preserves
// the rendered ledger and panel record into the archived change, stages
// everything, has check-archive-scope (in-process) verify the staged diff
// stays under spectre/changes/, and commits with the fixed subject.
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
	// The name is spliced into the branch, the archive path and the ledger
	// paths: plainChangeName carries the rule and its reasoning.
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

	// Each record when present; an absent one copies nothing. One that is
	// present and cannot be copied is a stop: the archive would otherwise
	// land without it, silently.
	for _, p := range [][2]string{{"ledgers/" + name + ".md", "ledger.md"}, {"reviews/" + name + "-panel.md", "panel.md"}} {
		src := wt + "/.superpowers/sdd/" + p[0]
		b, err := os.ReadFile(src)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		dst := wt + "/spectre/changes/archive/" + name + "/" + p[1]
		if err == nil {
			err = os.WriteFile(dst, b, 0o644)
		}
		if err != nil {
			return refuse("cannot copy %s to %s: %v", src, dst, err)
		}
	}

	if git("-C", wt, "add", "-A").Run() != nil {
		return refuse("git add -A failed in %s", wt)
	}
	// The scope guard's SCOPE-OK is not relayed: this guard prints one
	// verdict. Its violation lines are the verdict when it refuses.
	var lines bytes.Buffer
	switch code := checkArchiveScope([]string{wt, "spectre/changes/"}, env, &lines, stderr); code {
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
	commit := git("-C", wt, "commit", "-q", "-m", "chore(spectre): archive "+name)
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
