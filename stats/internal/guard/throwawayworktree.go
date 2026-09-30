package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// throwawayWorktree is scripts/throwaway-worktree.sh: that script's header
// comment is the contract. It ports, step for step, the create fence the
// wave-group copy and the slot copy shared and the slot copy's fold-back and
// remove fence (skills/flow/sdd-dispatch.md and
// skills/flow/review-panel-optional-slots.md at ae805186). One deliberate
// difference: the fences ran every step whatever the last one did, where the
// guard stops at the first failed step with exit 2, so a half-made copy is
// never handed to a dispatch as a whole one.
func init() {
	Registry["throwaway-worktree"] = throwawayWorktree
}

const twUsage = "usage: throwaway-worktree.sh create <worktree> <copy> [--sdd]\n" +
	"       throwaway-worktree.sh remove <worktree> <copy> [--fold-back]\n"

func throwawayWorktree(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) < 3 || len(args) > 4 {
		fmt.Fprint(stderr, twUsage)
		return 2
	}
	verb, wt, cp := args[0], args[1], args[2]
	flag := ""
	if len(args) == 4 {
		flag = args[3]
	}
	switch {
	case verb == "create" && (flag == "" || flag == "--sdd"):
		return twStep(stderr, twCreate(env, wt, cp, flag == "--sdd"))
	case verb == "remove" && (flag == "" || flag == "--fold-back"):
		return twStep(stderr, twRemove(env, wt, cp, flag == "--fold-back"))
	}
	fmt.Fprint(stderr, twUsage)
	return 2
}

func twStep(stderr io.Writer, err error) int {
	if err != nil {
		fmt.Fprintf(stderr, "throwaway-worktree: %v\n", err)
		return 2
	}
	return 0
}

// twRun runs name args in env.Dir, as the fence's shell ran it, with stdin
// piped in; a failure carries the command's stderr.
func twRun(env Env, stdin []byte, name string, args ...string) ([]byte, error) {
	if p, ok := lookPath(env, name); ok {
		name = p
	}
	cmd := exec.Command(name, args...)
	cmd.Dir = env.Dir
	cmd.Stdin = bytes.NewReader(stdin)
	var errb bytes.Buffer
	cmd.Stderr = &errb
	out, err := cmd.Output()
	if err != nil {
		return out, fmt.Errorf("%s %s: %v: %s", filepath.Base(name), strings.Join(args, " "), err, strings.TrimSpace(errb.String()))
	}
	return out, nil
}

// twDirname is dirname(1): a trailing slash is not a path component, so an
// untracked directory's `u/` entry is created under ".".
func twDirname(p string) string {
	return filepath.Dir(strings.TrimRight(p, "/"))
}

func twCreate(env Env, wt, cp string, sdd bool) error {
	if _, err := twRun(env, nil, "git", "-C", wt, "worktree", "add", "--detach", cp, "HEAD"); err != nil {
		return err
	}
	diff, err := twRun(env, nil, "git", "-C", wt, "diff", "--no-renames", "HEAD", "--binary")
	if err != nil {
		return err
	}
	if _, err := twRun(env, diff, "git", "-C", cp, "apply", "--whitespace=nowarn", "--allow-empty"); err != nil {
		return err
	}
	status, err := twRun(env, nil, "git", "-C", wt, "status", "--porcelain", "--untracked-files=normal", "-z")
	if err != nil {
		return err
	}
	// Every NUL-terminated token is read as an entry, a rename's origin token
	// included, exactly as the fence's `read -d ''` loop read them; only an
	// untracked (`??`) entry is copied.
	for _, entry := range strings.Split(strings.TrimSuffix(string(status), "\x00"), "\x00") {
		if len(entry) < 3 || entry[:2] != "??" {
			continue
		}
		f := entry[3:]
		if err := os.MkdirAll(smcAbs(env, cp+"/"+twDirname(f)), 0o755); err != nil {
			return err
		}
		if _, err := twRun(env, nil, "cp", "-a", wt+"/"+f, cp+"/"+f); err != nil {
			return err
		}
	}
	if !sdd {
		return nil
	}
	if err := os.MkdirAll(smcAbs(env, cp+"/.superpowers"), 0o755); err != nil {
		return err
	}
	if isDir(smcAbs(env, wt+"/.superpowers/sdd")) {
		if _, err := twRun(env, nil, "cp", "-a", wt+"/.superpowers/sdd", cp+"/.superpowers/sdd"); err != nil {
			return err
		}
	}
	return nil
}

// twNonEmpty is `[ -s <p> ]`.
func twNonEmpty(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.Size() > 0
}

// twRemove always removes the copy, as the fence did: a failed fold-back is
// still reported (exit 2), but never leaves the copy registered behind it.
func twRemove(env Env, wt, cp string, foldBack bool) error {
	var foldErr error
	if foldBack {
		foldErr = twFoldBack(env, wt, cp)
	}
	_, err := twRun(env, nil, "git", "-C", wt, "worktree", "remove", "--force", cp)
	return errors.Join(foldErr, err)
}

// twFoldBack copies a slot's report or reproducer back into the worktree
// unless the worktree already holds a non-empty file the copy's is not newer
// than.
func twFoldBack(env Env, wt, cp string) error {
	var found []string
	for _, pat := range []string{"/.superpowers/sdd/panel-report-*.md", "/.superpowers/sdd/reproducers/*.sh"} {
		m, _ := filepath.Glob(smcAbs(env, cp) + pat)
		found = append(found, m...)
	}
	prefix := smcAbs(env, cp) + "/.superpowers/sdd/"
	for _, f := range found {
		if !twNonEmpty(f) {
			continue
		}
		b := strings.TrimPrefix(f, prefix)
		if err := os.MkdirAll(smcAbs(env, wt+"/.superpowers/sdd/"+twDirname(b)), 0o755); err != nil {
			return err
		}
		c := smcAbs(env, wt+"/.superpowers/sdd/"+b)
		if twNonEmpty(c) {
			fs, _ := os.Stat(f)
			cs, _ := os.Stat(c)
			if !fs.ModTime().After(cs.ModTime()) {
				continue
			}
		}
		if _, err := twRun(env, nil, "cp", "-a", f, c); err != nil {
			return err
		}
	}
	return nil
}
