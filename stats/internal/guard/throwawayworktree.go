package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// The contract below moved here from the header of the deleted
// scripts/throwaway-worktree.sh shim (KAN-1058).
//
// flow-guard throwaway-worktree — create or remove the throwaway copy of a
// worktree that a review-panel slot (skills/flow/review-panel-optional-slots.md) works in.
//
// Usage: flow-guard throwaway-worktree create <worktree> <copy> [--sdd]
//        flow-guard throwaway-worktree remove <worktree> <copy> [--fold-back]
//
// create: `git worktree add -B flow-copy/<basename of copy> <copy> HEAD` -- the
//   copy on its own branch, reset to HEAD when one is left from an earlier
//   copy -- and, when the worktree's branch records `branch.<branch>.flowBase`,
//   the same value set as `branch.flow-copy/<basename>.flowBase`, so a tool
//   reading `branch.<current>.flowBase` in the copy resolves the change's base
//   (kan-1090 self-review, F14); then the worktree's
//   uncommitted diff against HEAD (`diff HEAD --binary`) applied in the copy,
//   then every untracked (`??`) entry of `status --porcelain -z` copied with
//   `cp -a`, then the worktree's `<spec root>/changes/` -- gitignored, so
//   neither the diff nor the untracked entries carry it -- copied into the
//   same path of the copy, when present. --sdd also copies `<worktree>/.superpowers/sdd`, when present,
//   to the same path in the copy.
// remove: `git worktree remove --force <copy>`, then `branch -D
//   flow-copy/<basename of copy>` when that branch exists, which drops its
//   config section with it. --fold-back first copies each
//   non-empty `.superpowers/sdd/panel-report-*.md`,
//   `.superpowers/sdd/reproducers/*.sh`, `.superpowers/sdd/dispatch-context-group-*.md`
//   `.superpowers/sdd/implementer-report-*.md` and every file under each
//   `.superpowers/sdd/*-frames/` directory from the copy back into the worktree,
//   unless the worktree already holds a non-empty file the copy's is not
//   newer than. By the same rule it also folds each non-empty
//   `panel-report-*.md` at the copy's root (copied) and at the worktree's root
//   (moved) into `<worktree>/.superpowers/sdd/` — a report a slot wrote by a
//   relative path after a `cd` (kan-1119 self-review, F2).
//
// Prints nothing on success.
// Exit 0 done; exit 2 usage, or a step failed — one stderr line names it, and
// no later step ran.
//
// The logic lives in stats/internal/guard/throwawayworktree.go.

// throwawayWorktree is `flow-guard throwaway-worktree`: the comment above is the contract. It ports, step for step, the create fence the
// slot copy used and the slot copy's fold-back and
// remove fence (skills/flow/sdd-dispatch.md and
// skills/flow/review-panel-optional-slots.md at ae805186). One deliberate
// difference: the fences ran every step whatever the last one did, where the
// guard stops at the first failed step with exit 2, so a half-made copy is
// never handed to a dispatch as a whole one.
func init() {
	Registry["throwaway-worktree"] = throwawayWorktree
}

const twUsage = "usage: flow-guard throwaway-worktree create <worktree> <copy> [--sdd]\n" +
	"       flow-guard throwaway-worktree remove <worktree> <copy> [--fold-back]\n"

func throwawayWorktree(args []string, env core.Env, stdout, stderr io.Writer) int {
	if len(args) < 3 || len(args) > 4 {
		fmt.Fprint(stderr, twUsage)
		core.Reason(env, "usage")
		return 2
	}
	verb, wt, cp := args[0], args[1], args[2]
	flag := ""
	if len(args) == 4 {
		flag = args[3]
	}
	switch {
	case verb == "create" && (flag == "" || flag == "--sdd"):
		return twStep(env, stderr, twCreate(env, wt, cp, flag == "--sdd"))
	case verb == "remove" && (flag == "" || flag == "--fold-back"):
		return twStep(env, stderr, twRemove(env, wt, cp, flag == "--fold-back"))
	}
	fmt.Fprint(stderr, twUsage)
	core.Reason(env, "usage")
	return 2
}

func twStep(env core.Env, stderr io.Writer, err error) int {
	if err != nil {
		fmt.Fprintf(stderr, "throwaway-worktree: %v\n", err)
		core.Reason(env, "step-failed")
		return 2
	}
	return 0
}

// twRun runs name args in env.Dir, as the fence's shell ran it, with stdin
// piped in; a failure carries the command's stderr.
func twRun(env core.Env, stdin []byte, name string, args ...string) ([]byte, error) {
	if p, ok := core.LookPath(env, name); ok {
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

func twCreate(env core.Env, wt, cp string, sdd bool) error {
	br := "flow-copy/" + filepath.Base(cp)
	if _, err := twRun(env, nil, "git", "-C", wt, "worktree", "add", "-B", br, cp, "HEAD"); err != nil {
		return err
	}
	if src, _ := twRun(env, nil, "git", "-C", wt, "symbolic-ref", "-q", "--short", "HEAD"); len(src) > 0 {
		if base, _ := twRun(env, nil, "git", "-C", wt, "config", "--get", "branch."+strings.TrimSpace(string(src))+".flowBase"); len(base) > 0 {
			if _, err := twRun(env, nil, "git", "-C", wt, "config", "branch."+br+".flowBase", strings.TrimSpace(string(base))); err != nil {
				return err
			}
		}
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
		if err := os.MkdirAll(core.Abs(env, cp+"/"+twDirname(f)), 0o755); err != nil {
			return err
		}
		if _, err := twRun(env, nil, "cp", "-a", wt+"/"+f, cp+"/"+f); err != nil {
			return err
		}
	}
	leaf := core.SpecRootLeaf(core.Abs(env, wt), io.Discard)
	if core.IsDir(core.Abs(env, wt+"/"+leaf+"/changes")) {
		if err := os.MkdirAll(core.Abs(env, cp+"/"+leaf+"/changes"), 0o755); err != nil {
			return err
		}
		if _, err := twRun(env, nil, "cp", "-a", wt+"/"+leaf+"/changes/.", cp+"/"+leaf+"/changes"); err != nil {
			return err
		}
	}
	if !sdd {
		return nil
	}
	if err := os.MkdirAll(core.Abs(env, cp+"/.superpowers"), 0o755); err != nil {
		return err
	}
	if core.IsDir(core.Abs(env, wt+"/.superpowers/sdd")) {
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
func twRemove(env core.Env, wt, cp string, foldBack bool) error {
	var foldErr error
	if foldBack {
		foldErr = twFoldBack(env, wt, cp)
	}
	_, err := twRun(env, nil, "git", "-C", wt, "worktree", "remove", "--force", cp)
	if err == nil {
		err = twDropCopyBranch(env, wt, cp)
	}
	return errors.Join(foldErr, err)
}

// twDropCopyBranch deletes the copy's flow-copy/<basename> branch when it
// exists; a copy made detached, before F14, has none.
func twDropCopyBranch(env core.Env, wt, cp string) error {
	ref := "refs/heads/flow-copy/" + filepath.Base(cp)
	if _, err := twRun(env, nil, "git", "-C", wt, "rev-parse", "--verify", "-q", ref); err != nil {
		return nil
	}
	_, err := twRun(env, nil, "git", "-C", wt, "branch", "-D", "flow-copy/"+filepath.Base(cp))
	return err
}

// twHeld reports whether c already holds a non-empty file that f is not newer than.
func twHeld(f, c string) bool {
	if !twNonEmpty(c) {
		return false
	}
	fs, _ := os.Stat(f)
	cs, _ := os.Stat(c)
	return !fs.ModTime().After(cs.ModTime())
}

// twFoldBack copies a slot's report or reproducer back into the worktree
// unless the worktree already holds a non-empty file the copy's is not newer
// than.
func twFoldBack(env core.Env, wt, cp string) error {
	var found []string
	for _, pat := range []string{"/.superpowers/sdd/panel-report-*.md", "/.superpowers/sdd/reproducers/*.sh",
		"/.superpowers/sdd/dispatch-context-group-*.md", "/.superpowers/sdd/implementer-report-*.md"} {
		m, _ := filepath.Glob(core.Abs(env, cp) + pat)
		found = append(found, m...)
	}
	// A plan's frames directory carries a render script and its frames at any
	// depth (kan-1132 self-review, F5).
	dirs, _ := filepath.Glob(core.Abs(env, cp) + "/.superpowers/sdd/*-frames")
	for _, d := range dirs {
		_ = filepath.WalkDir(d, func(p string, e fs.DirEntry, err error) error {
			if err == nil && e.Type().IsRegular() {
				found = append(found, p)
			}
			return nil
		})
	}
	prefix := core.Abs(env, cp) + "/.superpowers/sdd/"
	for _, f := range found {
		if !twNonEmpty(f) {
			continue
		}
		b := strings.TrimPrefix(f, prefix)
		if err := os.MkdirAll(core.Abs(env, wt+"/.superpowers/sdd/"+twDirname(b)), 0o755); err != nil {
			return err
		}
		c := core.Abs(env, wt+"/.superpowers/sdd/"+b)
		if twHeld(f, c) {
			continue
		}
		if _, err := twRun(env, nil, "cp", "-a", f, c); err != nil {
			return err
		}
	}
	for _, root := range []string{cp, wt} {
		m, _ := filepath.Glob(core.Abs(env, root) + "/panel-report-*.md")
		for _, f := range m {
			if !twNonEmpty(f) {
				continue
			}
			sdd := core.Abs(env, wt+"/.superpowers/sdd")
			if err := os.MkdirAll(sdd, 0o755); err != nil {
				return err
			}
			c := sdd + "/" + filepath.Base(f)
			if twHeld(f, c) {
				if root == wt {
					if err := os.Remove(f); err != nil {
						return err
					}
				}
				continue
			}
			var err error
			if root == wt {
				err = os.Rename(f, c)
			} else {
				_, err = twRun(env, nil, "cp", "-a", f, c)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
