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
	return Read(func(args ...string) ([]byte, error) {
		return exec.Command("git", append([]string{"-C", repo}, args...)...).Output()
	}, name, file)
}

// Git runs one git command against a repository and returns its stdout, an
// error its non-zero exit -- the seam a caller with its own bounded git
// runner (the self-review bundle) reads through.
type Git func(args ...string) ([]byte, error)

// Read is File through git: the same resolution, every git call made by it.
func Read(git Git, name, file string) (b []byte, ok bool, err error) {
	sha, err := newest(git, "chore(spectre): drop "+name)
	if err != nil {
		return nil, false, err
	}
	rev := sha + "^:spectre/changes/" + name + "/" + file
	if sha == "" {
		if sha, err = newest(git, "chore(spectre): archive "+name); err != nil || sha == "" {
			return nil, false, err
		}
		rev = sha + ":spectre/changes/archive/" + name + "/" + file
	}
	if _, err := git("cat-file", "-e", rev); err != nil {
		return nil, false, nil
	}
	if b, err = git("show", rev); err != nil {
		return nil, false, fmt.Errorf("git show %s: %w", rev, err)
	}
	return b, true, nil
}

// newest is the sha of the newest commit, over every ref, whose subject is
// exactly subject; "" when none. The --grep is a prefilter only: it also
// matches a longer name (kan-10 for kan-1) and the message body.
func newest(git Git, subject string) (string, error) {
	out, err := git("log", "--all", "--format=%H %s", "--fixed-strings", "--grep", subject)
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
