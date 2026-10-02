package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// reshapeBranch is scripts/reshape-branch.sh: that script's header is the
// contract — rebuild every planning commit in <merge-base>..HEAD on top of
// <merge-base>, then `reset --soft` there, behind
// check-planning-commit-location.sh. Every git step's stderr passes through
// and its exit code is the guard's, as `set -euo pipefail` made it.
func init() { Registry["reshape-branch"] = reshapeBranch }

func reshapeBranch(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 3 {
		fmt.Fprintln(stderr, "usage: reshape-branch.sh <worktree> <name> <merge-base>")
		return 2
	}
	wt, name, base := args[0], args[1], args[2]
	scriptDir, ok := guardSelfDir(env, stderr, "reshape-branch: ", "reshape-branch")
	if !ok {
		return 2
	}
	if rc := planningLocation(env, scriptDir, wt, name, stdout, stderr, "reshape-branch"); rc != 0 {
		return rc
	}

	// THE WORKING TREE AND THE REAL INDEX ARE NEVER TOUCHED until the final
	// `reset --soft`: every tree is written through this scratch index, so a
	// failure anywhere before that leaves the branch exactly as it was.
	scratch, err := os.MkdirTemp("", "reshape-branch.")
	if err != nil {
		fmt.Fprintf(stderr, "reshape-branch.sh: cannot create a scratch directory: %v\n", err)
		return 1
	}
	defer func() { _ = os.RemoveAll(scratch) }()
	git, scratchGit := envGit(env), envGit(env, "GIT_INDEX_FILE="+scratch+"/index")
	// out is `VAR="$(git -C <wt> ...)"`: stdout raw, stderr passed through.
	out := func(g func(...string) *exec.Cmd, stdin []byte, args ...string) (string, int) {
		cmd := g(append([]string{"-C", wt}, args...)...)
		var b bytes.Buffer
		cmd.Stdout, cmd.Stderr = &b, stderr
		if stdin != nil {
			cmd.Stdin = bytes.NewReader(stdin)
		}
		rc := exitCode(cmd.Run())
		return b.String(), rc
	}

	baseSHA, rc := out(git, nil, "rev-parse", "--verify", "--quiet", base+"^{commit}")
	if rc != 0 {
		fmt.Fprintf(stderr, "reshape-branch.sh: base does not resolve to a commit: %s\n", base)
		return 2
	}
	baseSHA = strings.TrimRight(baseSHA, "\n")
	planDir := specRootLeaf(pcAbs(env, wt), stderr) + "/changes"

	plans, rc := out(git, nil, "rev-list", "--reverse", "--no-merges", baseSHA+"..HEAD", "--", planDir+"/")
	if rc != 0 {
		return rc
	}
	tip, kept := baseSHA, 0
	for _, sha := range strings.Fields(plans) {
		// The tip's tree with <leaf>/changes/ replaced by this planning
		// commit's own (or removed, when the commit has none).
		if _, rc := out(scratchGit, nil, "read-tree", tip); rc != 0 {
			return rc
		}
		if _, rc := out(scratchGit, nil, "rm", "-r", "-q", "-f", "--cached", "--ignore-unmatch", "--", planDir+"/"); rc != 0 {
			return rc
		}
		probe := scratchGit("-C", wt, "cat-file", "-e", sha+":"+planDir)
		if probe.Run() == nil {
			if _, rc := out(scratchGit, nil, "read-tree", "--prefix="+planDir+"/", sha+":"+planDir); rc != 0 {
				return rc
			}
		}
		tree, rc := out(scratchGit, nil, "write-tree")
		if rc != 0 {
			return rc
		}
		// Message and author kept, committer now. As in the bash, a failed
		// author read is not a stop: only commit-tree's own status is.
		field := func(f string) string {
			s, _ := out(scratchGit, nil, "log", "-1", "--format="+f, sha)
			return strings.TrimRight(s, "\n")
		}
		author := envGit(env, "GIT_INDEX_FILE="+scratch+"/index",
			"GIT_AUTHOR_NAME="+field("%an"), "GIT_AUTHOR_EMAIL="+field("%ae"), "GIT_AUTHOR_DATE="+field("%aI"))
		// `--format` terminates each commit's output with a newline after the
		// raw body, which already ends in its own: exactly that one is cut, so
		// the message is kept byte for byte. The bash fed %B's output whole,
		// and every rebuilt planning commit gained a trailing blank line.
		msg, _ := out(scratchGit, nil, "log", "-1", "--format=%B", sha)
		msg = strings.TrimSuffix(msg, "\n")
		next, rc := out(author, []byte(msg), "commit-tree", strings.TrimRight(tree, "\n"), "-p", tip, "-F", "-")
		if rc != 0 {
			return rc
		}
		tip = strings.TrimRight(next, "\n")
		kept++
	}

	if _, rc := out(git, nil, "reset", "-q", "--soft", tip); rc != 0 {
		return rc
	}
	fmt.Fprintf(stdout, "RESHAPED: %s — %d planning commit(s) kept on %s\n", wt, kept, baseSHA[:12])
	return 0
}
