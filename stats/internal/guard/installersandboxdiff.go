package guard

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"syscall"
)

// installer-sandbox-diff runs `setup.sh global` from two trees under two
// fresh HOME sandboxes and diffs the installed result.
//
// Usage: installer-sandbox-diff.sh <old-tree> <new-tree>
//
//	scripts/installer-sandbox-diff.sh ../agents.main ../agents-worktrees/kan-536-…   # refactored installer vs main
//
// The installed result is compared as a normalized manifest of everything the
// installer created under each sandbox HOME, one line per entry, sorted by
// relative path in byte order:
//
//	l <rel-path> -> <readlink target, with the source tree replaced by <TREE>>
//	d <stat mode> <rel-path>
//	f <stat mode> <cksum crc size> <rel-path>
//
// Symlink targets are normalized against their own tree because every install
// link points back into the tree it was installed FROM — compared raw, an
// identical installer would diff on every single link. The tree path is
// replaced as a literal string, never a pattern, so a directory name holding
// glob metacharacters still normalizes. Regular files carry the `cksum`
// output, so the managed blocks (CLAUDE.md, AGENTS.md, .zshrc, the generated
// agent-baseline.md) are compared by content; modes catch an exec bit lost or
// gained. The mode is what the bash script's `stat` printed: the full st_mode
// in octal on Darwin (BSD `stat -f '%p'`), the permission bits in octal
// elsewhere (GNU `stat -c '%a'`). Both manifests are produced and compared on
// one machine, so the two formats never meet.
//
// The three per-attempt pitfalls this replaces (kan-536's hand-rolled diff
// needed three tries) are closed by construction:
//   - a fixture without setup.sh at its root — both trees are checked for
//     setup.sh before anything runs, and the error names the tree;
//   - a lost exec bit — the installer is invoked as `bash ./setup.sh`, which
//     needs no exec bit, while installed modes are still compared;
//   - a reused HOME that turned run two into an idempotent refresh — each
//     side gets its own fresh HOME inside a fresh work directory, never
//     either tree and never each other. Identical trees therefore diff clean
//     only if both runs installed into two separate, empty sandboxes.
//
// MANAGED_BLOCK_POSTPROCESS is removed from both runs' environment: it is the
// installer's own extension point, and a value exported by the caller would
// rewrite both sides' plain managed blocks identically — which could mask a
// genuine content difference between the trees. The verdict must be the
// installer's own output, not the caller's transform of it. (The installer's
// own zcode rewrite is prefix-scoped inside setup.sh and needs no help from
// here.)
//
// Exit codes:
//
//	0  the installed results are byte-identical; the sandbox is removed
//	1  they differ — the unified manifest diff (`diff -u`) goes to stdout and
//	   the whole sandbox (both HOMEs, both installer logs, both manifests) is
//	   KEPT and its path printed, so a difference can be walked in place
//	2  usage or environment error — bad argument count, an argument that is
//	   not a directory, setup.sh missing from a tree root (named), an unusable
//	   $TMPDIR, or an installer run that failed (log tail printed)
//
// Every non-zero exit after the work directory exists prints the kept-sandbox
// line on stderr.
func init() { Registry["installer-sandbox-diff"] = installerSandboxDiff }

func installerSandboxDiff(args []string, env Env, stdout, stderr io.Writer) int {
	die := func(msg string) int {
		fmt.Fprintln(stderr, "ERROR: "+msg)
		return 2
	}
	if len(args) != 2 {
		return die("Usage: installer-sandbox-diff.sh <old-tree> <new-tree>")
	}
	var trees [2]string
	for i, arg := range args {
		tree := arg
		if !filepath.IsAbs(tree) {
			tree = filepath.Join(env.Dir, tree)
			if !filepath.IsAbs(tree) {
				wd, _ := os.Getwd()
				tree = filepath.Join(wd, tree)
			}
		}
		tree = filepath.Clean(tree)
		if st, err := os.Stat(tree); err != nil || !st.IsDir() {
			return die(arg + " is not a directory")
		}
		if st, err := os.Stat(filepath.Join(tree, "setup.sh")); err != nil || !st.Mode().IsRegular() {
			return die(tree + " has no setup.sh at its root — the installer cannot be run from it")
		}
		trees[i] = tree
	}

	tmp := env.Getenv("TMPDIR")
	if tmp == "" {
		tmp = "/tmp"
	}
	// An environmental failure here is exit 2, never the 1 reserved for "differ".
	work, err := os.MkdirTemp(tmp, "installer-sandbox-diff.")
	if err != nil {
		fmt.Fprintln(stderr, "mktemp:", err)
		return die("cannot create the sandbox work directory under " + tmp)
	}
	rc := isdRun(trees, work, stdout, stderr)
	if rc == 0 {
		_ = os.RemoveAll(work)
	} else {
		fmt.Fprintln(stderr, "sandboxes, logs and manifests kept for inspection: "+work)
	}
	return rc
}

// isdRun installs both trees into fresh HOMEs under work, writes both
// manifests and diffs them.
func isdRun(trees [2]string, work string, stdout, stderr io.Writer) int {
	sides := [2]string{"old", "new"}
	for _, side := range sides {
		if err := os.Mkdir(filepath.Join(work, side+"-home"), 0o755); err != nil {
			fmt.Fprintln(stderr, "ERROR: cannot create the sandbox HOME directories")
			return 2
		}
	}
	for i, side := range sides {
		home, log := filepath.Join(work, side+"-home"), filepath.Join(work, side+"-install.log")
		if !isdInstall(trees[i], home, log) {
			fmt.Fprintf(stderr, "ERROR: setup.sh failed in %s — last 15 lines of %s:\n", trees[i], log)
			tail := exec.Command("tail", "-n", "15", log)
			tail.Stdout = stderr
			_ = tail.Run()
			return 2
		}
	}
	for i, side := range sides {
		m, err := isdManifest(filepath.Join(work, side+"-home"), trees[i])
		if err == nil {
			err = os.WriteFile(filepath.Join(work, side+".manifest"), m, 0o644)
		}
		if err != nil {
			fmt.Fprintln(stderr, "ERROR:", err)
			return 2
		}
	}
	var diff bytes.Buffer
	cmd := exec.Command("diff", "-u", filepath.Join(work, "old.manifest"), filepath.Join(work, "new.manifest"))
	cmd.Stdout, cmd.Stderr = &diff, stderr
	err := cmd.Run()
	_ = os.WriteFile(filepath.Join(work, "manifest.diff"), diff.Bytes(), 0o644)
	if err == nil {
		fmt.Fprintf(stdout, "installer output identical: %s vs %s\n", trees[0], trees[1])
		return 0
	}
	_, _ = stdout.Write(diff.Bytes())
	fmt.Fprintf(stderr, "installer output differs: %s vs %s\n", trees[0], trees[1])
	return 1
}

// isdInstall runs `bash ./setup.sh global` in tree with HOME=home, without
// MANAGED_BLOCK_POSTPROCESS, stdout and stderr both into log. PWD is the
// tree as given, as the bash script's `cd "$tree"` set it, so an installer
// that reads its own location through `pwd` sees the logical path, not the
// symlink-resolved one.
func isdInstall(tree, home, log string) bool {
	f, err := os.Create(log)
	if err != nil {
		return false
	}
	defer f.Close()
	cmd := exec.Command("bash", "./setup.sh", "global")
	cmd.Dir, cmd.Stdout, cmd.Stderr = tree, f, f
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "HOME=") && !strings.HasPrefix(kv, "PWD=") && !strings.HasPrefix(kv, "MANAGED_BLOCK_POSTPROCESS=") {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	cmd.Env = append(cmd.Env, "HOME="+home, "PWD="+tree)
	return cmd.Run() == nil
}

// isdManifest is the normalized manifest of one sandbox HOME.
func isdManifest(home, tree string) ([]byte, error) {
	infos := map[string]fs.FileInfo{}
	err := filepath.WalkDir(home, func(p string, d fs.DirEntry, err error) error {
		if err == nil && p != home {
			st, err := d.Info()
			if err != nil {
				return err
			}
			infos[strings.TrimPrefix(p, home+"/")] = st
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	rels := slices.Sorted(maps.Keys(infos))
	var b bytes.Buffer
	for _, rel := range rels {
		path := filepath.Join(home, rel)
		st := infos[rel]
		switch {
		case st.Mode()&fs.ModeSymlink != 0:
			target, err := os.Readlink(path)
			if err != nil {
				return nil, err
			}
			fmt.Fprintf(&b, "l %s -> %s\n", rel, strings.ReplaceAll(target, tree, "<TREE>"))
		case st.IsDir():
			fmt.Fprintf(&b, "d %s %s\n", isdEntryMode(st), rel)
		default:
			in, err := os.Open(path)
			if err != nil {
				return nil, err
			}
			cksum := exec.Command("cksum")
			cksum.Stdin = in
			out, err := cksum.Output()
			in.Close()
			if err != nil {
				return nil, err
			}
			fmt.Fprintf(&b, "f %s %s %s\n", isdEntryMode(st), strings.TrimRight(string(out), "\n"), rel)
		}
	}
	return b.Bytes(), nil
}

// isdEntryMode is what the bash script's stat printed for an entry.
func isdEntryMode(st fs.FileInfo) string {
	mode := uint32(st.Sys().(*syscall.Stat_t).Mode)
	if runtime.GOOS == "darwin" {
		return fmt.Sprintf("%o", mode)
	}
	return fmt.Sprintf("%o", mode&0o7777)
}
