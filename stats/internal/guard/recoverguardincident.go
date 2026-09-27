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
// usage error. Every abort precedes every restore, and each restore is a
// `git show` redirect that leaves the file unstaged. Every git call is a
// child process with the arguments the bash passed, in the bash's order.
func init() {
	Registry["recover-guard-incident"] = recoverGuardIncident
}

func recoverGuardIncident(args []string, env Env, stdout, stderr io.Writer) int {
	die := func(rc int, msg string) int {
		fmt.Fprintf(stderr, "recover-guard-incident: %s\n", msg)
		return rc
	}
	if len(args) > 0 && args[0] == "--help" {
		fmt.Fprint(stdout, "usage: recover-guard-incident [--apply] [repo-dir] [path...]\n"+
			"  --apply    execute; default is a dry-run plan\n"+
			"  repo-dir   git repository, default cwd\n"+
			"  path...    planning paths, repo-root-relative; default spectre/changes\n")
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

	// F1 contract: repo-dir (or the cwd) resolves to the toplevel, so path...
	// is repo-root-relative however deep inside the repo the tool is invoked.
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
		args = args[1:]
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
	if len(args) == 0 {
		args = []string{"spectre/changes"}
	}

	if _, rc := git(io.Discard, "-C", repo, "rev-parse", "-q", "--verify", "REVERT_HEAD"); rc != 0 {
		return die(1, "no revert in progress in "+repo+" (REVERT_HEAD missing) — nothing to recover")
	}
	if _, rc := git(io.Discard, "-C", repo, "rev-parse", "-q", "--verify", "stash@{0}"); rc != 0 {
		return die(1, "no stash entry in "+repo+" — recovery restores stash@{0}^3, which needs a stash")
	}
	if _, rc := git(io.Discard, "-C", repo, "rev-parse", "-q", "--verify", "stash@{0}^3"); rc != 0 {
		return die(1, "stash@{0} has no untracked third parent — it was not created with 'git stash -u'; re-stash with -u before any abort")
	}

	// The restore set: every file the stash's third parent holds under the
	// given paths, each checked against the working tree. Tracked anywhere
	// (index or HEAD) -> refuse the whole run; already present untracked ->
	// restore over it, named as an overwrite; absent -> a plain restore. The
	// ls-tree and ls-files exit statuses were never checked by the bash (a
	// process substitution and a `[ -n "$(...)" ]`); only their output is.
	var files []string
	overwrite := map[string]bool{}
	for _, p := range args {
		list, _ := git(stderr, "-C", repo, "ls-tree", "-r", "--name-only", "stash@{0}^3", "--", p)
		for _, f := range strings.Split(list, "\n") {
			if f == "" {
				continue
			}
			// The index first; HEAD only when the index names nothing, as the
			// bash's `||` short-circuited.
			tracked, _ := git(stderr, "-C", repo, "ls-files", "--", f)
			if tracked == "" {
				if _, rc := git(io.Discard, "-C", repo, "cat-file", "-e", "HEAD:"+f); rc == 0 {
					tracked = "HEAD"
				}
			}
			if tracked != "" {
				return die(1, "refusing: "+f+" is tracked in "+repo+" — recovery never clobbers tracked state")
			}
			files = append(files, f)
			if _, err := os.Stat(repo + "/" + f); err == nil {
				overwrite[f] = true
			}
		}
	}
	if len(files) == 0 {
		return die(1, "stash@{0}^3 holds no files under: "+strings.Join(args, " "))
	}

	// The reflog block prints in every mode (header): git writes straight to
	// the caller's stdout, and its failure is ignored as `|| true` did.
	fmt.Fprint(stdout, "reflog diagnosis — the alternating reset pattern the incident showed:\n")
	rgiRun(env.Dir, stdout, stderr, "git", "-C", repo, "reflog", "-g", "HEAD", "-n", "15")

	fmt.Fprintf(stdout, "plan:\n  git -C %s revert --abort\n", repo)
	for _, f := range files {
		if overwrite[f] {
			fmt.Fprintf(stdout, "  (overwrites an existing untracked file: %s)\n", f)
		}
		fmt.Fprintf(stdout, "  mkdir -p %s/%s\n", repo, path.Dir(f))
		fmt.Fprintf(stdout, "  git -C %s show \"stash@{0}^3:%s\" > %s/%s\n", repo, f, repo, f)
	}
	if !apply {
		return 0
	}

	// THE ORDER IS THE POINT (header): the one abort runs before any restore,
	// and a failed step ends the run with its exit status as `set -e` did.
	fmt.Fprintf(stdout, "running: git -C %s revert --abort\n", repo)
	if rc := rgiRun(env.Dir, stdout, stderr, "git", "-C", repo, "revert", "--abort"); rc != 0 {
		return rc
	}
	for _, f := range files {
		fmt.Fprintf(stdout, "running: git show \"stash@{0}^3:%s\" > %s/%s\n", f, repo, f)
		// mkdir is exec'd, not os.MkdirAll: its failure line is the platform
		// mkdir's own, byte for byte, as the bash printed it.
		if rc := rgiRun(env.Dir, stdout, stderr, "mkdir", "-p", repo+"/"+path.Dir(f)); rc != 0 {
			return rc
		}
		// The redirect: the target truncated (or created 0666 &^ umask)
		// before git runs, git's bytes written to it unstaged.
		out, err := os.Create(repo + "/" + f)
		if err != nil {
			return die(1, err.Error())
		}
		rc := rgiRun(env.Dir, out, stderr, "git", "-C", repo, "show", "stash@{0}^3:"+f)
		if cerr := out.Close(); rc == 0 && cerr != nil {
			return die(1, cerr.Error())
		}
		if rc != 0 {
			return rc
		}
	}
	return 0
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
