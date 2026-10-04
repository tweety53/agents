package guard

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

// unlandedArchivesCmd is `flow-guard unlanded-archives <root>`, the shared
// rule's entry point for the scanners not written in Go —
// check-task-build-green.sh, check-plan-shape.sh, plan-dispatch-bundles.sh
// and check-task-records.sh call it through scripts/lib/flow-guard.sh's
// flow_guard_exec and scan <spec-root>/changes/archive/<name>/ for each name
// it prints, one per line. The base is vmDefaultBase's — the merge base of
// HEAD with the change's base, resolved offline — so every scanner judges
// "landed" against the same commit check-verbatim-moves.sh does. A root that
// is not a git worktree, or one whose base does not resolve, has no base to
// compare against and prints nothing: every archived change stays skipped,
// the scanners' behaviour before archived changes were scanned at all.
// Exit 0 with the list (possibly empty), 2 on a usage error or an archive
// directory that cannot be read.
func init() {
	Registry["unlanded-archives"] = unlandedArchivesCmd
}

func unlandedArchivesCmd(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 1 || args[0] == "" {
		fmt.Fprintln(stderr, "usage: flow-guard unlanded-archives <root>")
		return 2
	}
	root := args[0]
	if !filepath.IsAbs(root) {
		root = filepath.Join(env.Dir, root)
	}
	git := envGit(env, "LC_ALL=C")
	names, err := unlandedArchivesAtDefaultBase(git, root)
	if err != nil {
		fmt.Fprintf(stderr, "unlanded-archives: %v\n", err)
		return 2
	}
	for _, n := range names {
		fmt.Fprintln(stdout, n)
	}
	return 0
}

// unlandedArchivesAtDefaultBase is unlandedArchives against vmDefaultBase's
// base, and nothing when root is not a git worktree or that base does not
// resolve.
func unlandedArchivesAtDefaultBase(git func(...string) *exec.Cmd, root string) ([]string, error) {
	if git("-C", root, "rev-parse", "--git-dir").Run() != nil {
		return nil, nil
	}
	base, _, ok := vmDefaultBase(git, root)
	if !ok {
		return nil, nil
	}
	return unlandedArchives(git, root, base)
}

// unlandedArchives is the one rule every repo-wide scanner of
// <spec-root>/changes/ applies to an archived change: it is in flight — and
// scanned like a live change — only while base does not carry its directory.
// Integrate's run 1 archives the change on its own branch, and a fix run
// after that writes into the archived directory, so a change archived on
// this branch is still this branch's work; one already archived at the base
// has landed and is never read again, which is what keeps every legacy
// archive skipped. It returns the names under
// <root>/<spec-root>/changes/archive/ whose directory base lacks, sorted.
// Only real directories count: a symlinked archive/ or entry is never
// listed, so no scanner is steered out of the tree through one. A missing
// archive/ is no names; one that cannot be read is an error.
func unlandedArchives(git func(...string) *exec.Cmd, root, base string) ([]string, error) {
	rel := filepath.Join(specRootLeaf(root, io.Discard), "changes", "archive")
	dir := filepath.Join(root, rel)
	if st, err := os.Lstat(dir); err != nil || !st.IsDir() {
		if err == nil || errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.ToSlash(filepath.Join(rel, e.Name()))
		if git("-C", root, "cat-file", "-e", base+":./"+path).Run() != nil {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}
