package guard

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"slices"
	"strings"
)

// snapshotTreeState and checkTreeRestored are scripts/lib/post-mutation-check.sh's
// snapshot_tree_state and check_tree_restored for the Go guards. The bash
// library stays the source of truth while break-and-prove.sh still sources
// it; its header carries the reasoning (KAN-448 part 1, KAN-423's incident)
// and TestPostMutationCheckParity fails when the two differ in output, stderr
// or status.
//
// Only NEW entries are drift: a stash entry the snapshot recorded and that is
// still present afterwards is expected, and a deleted entry is not residue.

const postMutationCheckSentinel = "--- post-mutation-check: stash section ---"

// snapshotTreeState is snapshot_tree_state: every `git status
// --porcelain=v2` line, the sentinel, then every `git stash list` line, git's
// own stderr passed through to stderr. The status is the stash list's exit
// code, the bash function's own return.
func snapshotTreeState(worktree string, stderr io.Writer) (string, int) {
	status, _ := gitExec("", nil, stderr, "-C", worktree, "status", "--porcelain=v2", "--untracked-files=normal")
	stash, rc := gitExec("", nil, stderr, "-C", worktree, "stash", "list")
	return status + postMutationCheckSentinel + "\n" + stash, rc
}

// checkTreeRestored is check_tree_restored: one finding per new stash entry
// (`new stash entry: ...`), then per unexpected status line (`unexpected
// status line: ...`), against snapshot. The status is 0 when there is none,
// 1 on any drift, 2 when git itself failed on the recompute.
func checkTreeRestored(worktree, snapshot string, stderr io.Writer) ([]string, int) {
	var snapStatus, snapStash []string
	inStash := false
	for _, line := range strings.Split(snapshot, "\n") {
		switch {
		case line == "":
		case line == postMutationCheckSentinel:
			inStash = true
		case inStash:
			snapStash = append(snapStash, line)
		default:
			snapStatus = append(snapStatus, line)
		}
	}
	nowStatus, rc := gitExec("", nil, stderr, "-C", worktree, "status", "--porcelain=v2", "--untracked-files=normal")
	if rc != 0 {
		return nil, 2
	}
	nowStash, rc := gitExec("", nil, stderr, "-C", worktree, "stash", "list")
	if rc != 0 {
		return nil, 2
	}
	var findings []string
	for _, line := range strings.Split(nowStash, "\n") {
		if line != "" && !slices.Contains(snapStash, line) {
			findings = append(findings, "new stash entry: "+line)
		}
	}
	for _, line := range strings.Split(nowStatus, "\n") {
		if line != "" && !slices.Contains(snapStatus, line) {
			findings = append(findings, "unexpected status line: "+line)
		}
	}
	if len(findings) > 0 {
		return findings, 1
	}
	return nil, 0
}

// gitExec runs git args from dir: stdout returned (and copied to out when it
// is non-nil), stderr to errw, and the `$?` bash would see — git's exit
// status, 128+n when signal n killed it (rrExitCode, the package's one
// mapping), 127 when git could not run.
func gitExec(dir string, out, errw io.Writer, args ...string) (string, int) {
	cmd := exec.Command("git", args...)
	cmd.Dir, cmd.Stderr = dir, errw
	var buf bytes.Buffer
	cmd.Stdout = &buf
	if out != nil {
		cmd.Stdout = io.MultiWriter(&buf, out)
	}
	err := cmd.Run()
	var ee *exec.ExitError
	switch {
	case err == nil:
		return buf.String(), 0
	case errors.As(err, &ee):
		return buf.String(), rrExitCode(ee.ProcessState)
	default:
		return buf.String(), 127
	}
}
