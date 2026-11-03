package guard

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// checkNoBash is `flow-guard check-no-bash`, the guard of the no-bash rule
// (rules/no-bash-scripts.mdc): no script in this repository is written in
// bash. A bash script grows a second copy of logic Go already carries, is
// tested by a harness of its own instead of `go test`, and fails on the
// quoting, word-splitting and exit-status edges Go never has — so every
// script is Go in flow-guard, and this guard keeps the next one from being
// written in bash. The rule file alone would let it through; agents would
// learn the rule only from review.
//
// It reads `git ls-files` of FLOW_GUARD_REPO_ROOT (CHECK_NO_BASH_ROOT when
// set, a test fixture's), minus spectre/changes/archive/ — history, kept as
// the landed change carried it — and reports
//
//	<path>: bash script — write it in Go in flow-guard (rules/no-bash-scripts.mdc)
//
// for every file ending in `.sh`, or whose first line is a `#!` whose
// interpreter, or the command `env` runs past its options and assignments,
// is `bash` or `sh` — except
// setup.sh and scripts/flow-guard, the two bash files the repository keeps:
// setup.sh installs everything, and the bootstrap must build the Go binary
// before any Go can run. Only committed files count: a shell command line, a
// fenced command block, a `make` recipe and a panel reproducer under the
// gitignored .superpowers/sdd/reproducers/ are not scripts in the rule's
// sense, and an untracked file is not yet committed to be run.
//
// Then NO-BASH-OK or NO-BASH-INVALID. Exit 0 clean, 1 any finding, 2 cannot
// answer (no git worktree at the root, or a tracked file it cannot read).
func init() {
	Registry["check-no-bash"] = checkNoBash
}

// nbKept is the two bash files the repository keeps.
var nbKept = map[string]bool{"setup.sh": true, "scripts/flow-guard": true}

func checkNoBash(_ []string, env Env, stdout, stderr io.Writer) int {
	root := env.Getenv("CHECK_NO_BASH_ROOT")
	if root == "" {
		root = env.Getenv("FLOW_GUARD_REPO_ROOT")
	}
	if root == "" {
		fmt.Fprint(stderr, "check-no-bash: no repository root — run flow-guard check-no-bash, which sets FLOW_GUARD_REPO_ROOT\n")
		return 2
	}
	out, err := exec.Command("git", "-C", root, "ls-files", "-z").Output()
	if err != nil {
		fmt.Fprintf(stderr, "check-no-bash: %s is not a git worktree: %v\n", root, err)
		return 2
	}
	n := 0
	for _, p := range strings.Split(string(out), "\x00") {
		if p == "" || nbKept[p] || strings.HasPrefix(p, "spectre/changes/archive/") {
			continue
		}
		bash := strings.HasSuffix(p, ".sh")
		if !bash {
			var err error
			if bash, err = nbShebang(filepath.Join(root, p)); err != nil {
				fmt.Fprintf(stderr, "check-no-bash: cannot read %s: %v\n", p, err)
				return 2
			}
		}
		if bash {
			fmt.Fprintf(stdout, "%s: bash script — write it in Go in flow-guard (rules/no-bash-scripts.mdc)\n", p)
			n++
		}
	}
	if n > 0 {
		fmt.Fprintf(stdout, "NO-BASH-INVALID: %s — %d bash script(s)\n", root, n)
		return 1
	}
	fmt.Fprintf(stdout, "NO-BASH-OK: %s\n", root)
	return 0
}

// nbShebang reports whether path's first line is a `#!` running bash or sh,
// directly or as `env`'s first argument. A tracked symlink or a path deleted
// from the work tree is not a script file here.
func nbShebang(path string) (bool, error) {
	if st, err := os.Lstat(path); err != nil || !st.Mode().IsRegular() {
		return false, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	line, err := bufio.NewReader(f).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	rest, ok := strings.CutPrefix(line, "#!")
	if !ok {
		return false, nil
	}
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return false, nil
	}
	interp := filepath.Base(fields[0])
	if interp == "env" {
		// env's options (`-S`, `-i`) and NAME=value assignments come
		// before the command it runs; -u, -C and -P take the next field as
		// their argument, and -S's attached form (`-Sbash`) carries the
		// command itself.
		interp = ""
		for i := 1; i < len(fields); i++ {
			f := fields[i]
			switch {
			case f == "-u" || f == "-C" || f == "-P" || f == "--unset" || f == "--chdir":
				i++
			case strings.HasPrefix(f, "-S") && len(f) > 2:
				interp = f[2:]
			case !strings.HasPrefix(f, "-") && !strings.Contains(f, "="):
				interp = f
			default:
				continue
			}
			if interp != "" {
				break
			}
		}
	}
	return interp == "bash" || interp == "sh", nil
}
