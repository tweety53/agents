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
// contract. It makes run 2's archive commit (skills/flow/archive.md step 4):
// asserts chore/archive-<name>, has check-done-when-paths (in-process)
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
	if len(args) != 3 {
		return refuse("usage: commit-archive.sh <landing-worktree> <canonical-worktree> <name>")
	}
	landing, canonical, name := pcAbs(env, args[0]), pcAbs(env, args[1]), args[2]
	// The name is spliced into the branch, the archive path and the ledger
	// paths: plainChangeName carries the rule and its reasoning.
	if !plainChangeName(name) {
		return refuse("change name %q is not a plain change name", name)
	}
	git := envGit(env)
	if !isDir(landing) || git("-C", landing, "rev-parse", "--is-inside-work-tree").Run() != nil {
		return refuse("%s is not a git worktree", landing)
	}

	// The branch is asserted before anything is written, so a refusal leaves
	// the landing worktree exactly as it was found.
	branch, ok := capture(git("-C", landing, "branch", "--show-current"))
	if !ok {
		return refuse("cannot read the current branch of %s", landing)
	}
	if branch != "chore/archive-"+name {
		if branch == "" {
			branch = "(detached HEAD)"
		}
		fmt.Fprintf(stdout, "ARCHIVE-WRONG-BRANCH: %s\n", branch)
		return 1
	}

	// The Done-when cross-check runs before anything is written, so a
	// refusal leaves the landing worktree exactly as it was found. Its OK
	// verdict is not relayed: this guard prints one verdict. Its
	// DONE-WHEN-PATH lines are the verdict when it refuses.
	var dw bytes.Buffer
	switch code := checkDoneWhenPaths([]string{landing}, env, &dw, stderr); code {
	case 0:
	case 1:
		_, _ = stdout.Write(dw.Bytes())
		return 1
	default:
		return refuse("check-done-when-paths.sh could not answer for %s: exit status %d", landing, code)
	}

	// Each record when present; an absent one copies nothing. One that is
	// present and cannot be copied is a stop: the archive would otherwise
	// land without it, silently.
	for _, p := range [][2]string{{"ledgers/" + name + ".md", "ledger.md"}, {"reviews/" + name + "-panel.md", "panel.md"}} {
		src := canonical + "/.superpowers/sdd/" + p[0]
		b, err := os.ReadFile(src)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		dst := landing + "/spectre/changes/archive/" + name + "/" + p[1]
		if err == nil {
			err = os.WriteFile(dst, b, 0o644)
		}
		if err != nil {
			return refuse("cannot copy %s to %s: %v", src, dst, err)
		}
	}

	if git("-C", landing, "add", "-A").Run() != nil {
		return refuse("git add -A failed in %s", landing)
	}
	// The scope guard's SCOPE-OK is not relayed: this guard prints one
	// verdict. Its violation lines are the verdict when it refuses.
	var lines bytes.Buffer
	switch code := checkArchiveScope([]string{landing, "spectre/changes/"}, env, &lines, stderr); code {
	case 0:
	case 1:
		_, _ = stdout.Write(lines.Bytes())
		return 1
	default:
		return refuse("check-archive-scope.sh could not answer for %s: exit status %d", landing, code)
	}

	err := git("-C", landing, "diff", "--cached", "--quiet").Run()
	var ee *exec.ExitError
	switch {
	case err == nil:
		fmt.Fprintln(stdout, "ARCHIVE-NOTHING-STAGED")
		return 0
	case !errors.As(err, &ee) || ee.ExitCode() != 1:
		return refuse("cannot read the staged diff of %s", landing)
	}
	commit := git("-C", landing, "commit", "-q", "-m", "chore(spectre): archive "+name)
	commit.Stdout, commit.Stderr = stderr, stderr
	if commit.Run() != nil {
		return refuse("git commit failed in %s", landing)
	}
	sha, ok := capture(git("-C", landing, "rev-parse", "HEAD"))
	if !ok {
		return refuse("committed, but cannot read HEAD of %s", landing)
	}
	fmt.Fprintf(stdout, "ARCHIVE-COMMITTED: %s\n", sha)
	return 0
}
