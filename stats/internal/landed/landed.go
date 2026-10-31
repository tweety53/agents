// Package landed reads a landed change's files from git history. Integrate's
// run 1 deletes a change's directory from the tip with one
// `chore(spectre): drop <name>` commit, so a reader that runs after the drop
// -- the self-review bundle, the merge-order read, the board's plan of a
// finished change, the record checkout check -- finds the files in that
// commit's parent. Changes landed before the drop existed were moved to
// spectre/changes/archive/<name>/ by a `chore(spectre): archive <name>`
// commit instead; that commit is the fallback when no drop commit exists.
package landed

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// File returns <file> of change <name> as it stood when the change landed in
// the repository at repo: from the parent of the newest
// `chore(spectre): drop <name>` commit at spectre/changes/<name>/<file>, else
// from the newest `chore(spectre): archive <name>` commit at
// spectre/changes/archive/<name>/<file>. ok is false, with no error, when
// neither commit exists or the chosen commit lacks the file; err is any
// other git failure.
func File(repo, name, file string) (b []byte, ok bool, err error) {
	return Read(repoGit(repo), name, file)
}

// Recent is File over the newest recentDepth commits of HEAD alone -- a
// bounded read for a caller polling an in-flight change, whose run 1 drop
// sits at the tip of the branch its worktree has checked out.
func Recent(repo, name, file string) (b []byte, ok bool, err error) {
	return read(repoGit(repo), name, file, "-n", recentDepth, "HEAD")
}

// recentDepth bounds Recent's walk: run 1's drop is followed by at most a
// few fix-run commits before the change finishes.
const recentDepth = "50"

func repoGit(repo string) Git {
	return func(args ...string) ([]byte, error) {
		return exec.Command("git", append([]string{"-C", repo}, args...)...).Output()
	}
}

// Git runs one git command against a repository and returns its stdout, an
// error its non-zero exit -- the seam a caller with its own bounded git
// runner (the self-review bundle) reads through.
type Git func(args ...string) ([]byte, error)

// Read is File through git: the same resolution, every git call made by it.
func Read(git Git, name, file string) (b []byte, ok bool, err error) {
	return read(git, name, file, "--all")
}

// read is Read with the log's revision range given as span.
func read(git Git, name, file string, span ...string) (b []byte, ok bool, err error) {
	sha, err := newest(git, span, "chore(spectre): drop "+name)
	if err != nil {
		return nil, false, err
	}
	commit, path := sha+"^", "spectre/changes/"+name+"/"+file
	if sha == "" {
		if sha, err = newest(git, span, "chore(spectre): archive "+name); err != nil || sha == "" {
			return nil, false, err
		}
		commit, path = sha, "spectre/changes/archive/"+name+"/"+file
	}
	// ls-tree prints nothing for an absent path and fails only on a git
	// fault, which cat-file -e would conflate with absence.
	listed, err := git("ls-tree", "--name-only", commit, "--", path)
	if err != nil {
		return nil, false, fmt.Errorf("git ls-tree %s: %w", commit, err)
	}
	if len(bytes.TrimSpace(listed)) == 0 {
		return nil, false, nil
	}
	rev := commit + ":" + path
	if b, err = git("show", rev); err != nil {
		return nil, false, fmt.Errorf("git show %s: %w", rev, err)
	}
	return b, true, nil
}

// newest is the sha of the newest commit in span whose subject is exactly
// subject; "" when none. The --grep is a prefilter only: it also
// matches a longer name (kan-10 for kan-1) and the message body.
func newest(git Git, span []string, subject string) (string, error) {
	out, err := git(append(append([]string{"log"}, span...), "--format=%H %s", "--fixed-strings", "--grep", subject)...)
	if err != nil {
		return "", fmt.Errorf("git log: %w", err)
	}
	for _, line := range strings.Split(string(bytes.TrimSpace(out)), "\n") {
		if sha, s, found := strings.Cut(line, " "); found && s == subject {
			return sha, nil
		}
	}
	return "", nil
}
