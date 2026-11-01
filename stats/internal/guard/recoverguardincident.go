package guard

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

// recoverGuardIncident is scripts/recover-guard-incident.sh: that script's
// header comment is the contract — exit 0 a printed dry-run plan or a
// completed apply; 1 a failed precondition, cause on stderr, stdout empty; 2 a
// usage error. It aborts a stuck revert and prints the reflog diagnosis; it
// restores nothing, since the planning files it once restored from a
// `git stash -u` are gitignored (KAN-1073) and no stash carries them. Every
// git call is a child process with the arguments the bash passed, in the
// bash's order.
func init() {
	Registry["recover-guard-incident"] = recoverGuardIncident
}

func recoverGuardIncident(args []string, env Env, stdout, stderr io.Writer) int {
	die := func(rc int, msg string) int {
		fmt.Fprintf(stderr, "recover-guard-incident: %s\n", msg)
		return rc
	}
	if len(args) > 0 && args[0] == "--help" {
		fmt.Fprint(stdout, "usage: recover-guard-incident [--apply] [repo-dir]\n"+
			"  --apply    execute; default is a dry-run plan\n"+
			"  repo-dir   git repository, default cwd\n")
		return 0
	}
	apply := len(args) > 0 && args[0] == "--apply"
	if apply {
		args = args[1:]
	}
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			return die(2, "unknown option: "+a)
		}
	}

	if len(args) > 1 {
		return die(2, "unexpected argument: "+args[1])
	}

	// F1 contract: repo-dir (or the cwd) resolves to the toplevel however
	// deep inside the repo the tool is invoked.
	// `cd "$1" && pwd -P`: joined to the caller's cwd, then made physical.
	// cd refuses an empty argument (bash 5's "null directory"), a missing
	// component before `..` (checked on the uncleaned path, which the
	// kernel walks component by component) and a directory it cannot
	// search; a lexical clean alone would accept all three.
	target, given := env.Dir, len(args) > 0
	if given {
		raw := args[0]
		if !filepath.IsAbs(raw) {
			raw = env.Dir + "/" + raw
		}
		searchable := func(p string) bool {
			fi, err := os.Stat(p)
			return err == nil && fi.IsDir() && syscall.Access(p, 1) == nil
		}
		// Logical cd: the lexical canonical path -- every prefix standing
		// before a `..` must be a directory -- when it is searchable; else
		// the argument as given, resolved physically.
		target = rgiLexical(raw)
		if target == "" || !searchable(target) {
			target = raw
		}
		if args[0] == "" || !searchable(target) {
			return die(2, "not a directory: "+args[0])
		}
	}
	dir, err := filepath.EvalSymlinks(target)
	if given {
		if err != nil {
			return die(2, "not a directory: "+args[0])
		}
	} else if err != nil {
		return die(1, err.Error())
	}
	git := func(errw io.Writer, a ...string) (string, int) {
		return gitExec(env.Dir, nil, errw, a...)
	}
	top, rc := git(io.Discard, "-C", dir, "rev-parse", "--show-toplevel")
	if rc != 0 {
		return die(2, "not a git repository: "+dir)
	}
	repo := strings.TrimRight(top, "\n")

	if _, rc := git(io.Discard, "-C", repo, "rev-parse", "-q", "--verify", "REVERT_HEAD"); rc != 0 {
		return die(1, "no revert in progress in "+repo+" (REVERT_HEAD missing) — nothing to recover")
	}

	// The reflog block prints in every mode (header): git writes straight to
	// the caller's stdout, and its failure is ignored as `|| true` did.
	fmt.Fprint(stdout, "reflog diagnosis — the alternating reset pattern the incident showed:\n")
	rgiRun(env.Dir, stdout, stderr, "git", "-C", repo, "reflog", "-g", "HEAD", "-n", "15")

	fmt.Fprintf(stdout, "plan:\n  git -C %s revert --abort\n", repo)
	if !apply {
		return 0
	}
	fmt.Fprintf(stdout, "running: git -C %s revert --abort\n", repo)
	return rgiRun(env.Dir, stdout, stderr, "git", "-C", repo, "revert", "--abort")
}

// rgiRun runs argv with its stdout and stderr handed to the child as they
// are: an *os.File (the caller's terminal, a restore target) reaches the
// child directly, as the bash's inherited descriptors and redirects did, so
// git sees the terminal where the bash's git saw it. The exit status is the
// child's, as `set -e` exited with it.
func rgiRun(dir string, out, errw io.Writer, argv ...string) int {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir, cmd.Stdout, cmd.Stderr = dir, out, errw
	err := cmd.Run()
	var ee *exec.ExitError
	switch {
	case err == nil:
		return 0
	case errors.As(err, &ee):
		return rrExitCode(ee.ProcessState)
	default:
		return 127
	}
}

// rgiLexical canonicalises an absolute path the way bash's cd does with
// PATH_CHECKDOTDOT: `.` and empty components drop, `..` removes the previous
// component once that prefix is checked to be a directory. "" when a prefix
// before a `..` is not one -- cd then falls back to the physical path.
func rgiLexical(abs string) string {
	out := ""
	for _, c := range strings.Split(abs, "/") {
		switch c {
		case "", ".":
		case "..":
			if fi, err := os.Stat(out + "/"); err != nil || !fi.IsDir() {
				return ""
			}
			out = path.Dir("/" + out)
			if out == "/" {
				out = ""
			}
		default:
			out += "/" + c
		}
	}
	if out == "" {
		return "/"
	}
	return out
}
