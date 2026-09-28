package setuptest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// statLine is `stat` without -L: name, size, mtime seconds, permission bits.
func statLine(p string) string {
	fi, err := os.Lstat(p)
	if err != nil {
		return "absent " + p + "\n"
	}
	return fmt.Sprintf("%s %d %d %o\n", p, fi.Size(), fi.ModTime().Unix(), fi.Mode().Perm())
}

func sha256hex(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }

// realHomeFingerprint samples only the paths setup.sh can write under home — the two harness
// directories, their skills/commands/rules subtrees one level down, the two managed
// instruction files, and any .flow.bak beside them. A full recursive listing of ~/.claude would
// be dominated by session state that changes on its own while this runs, which would make the
// check noisy and therefore ignored. For ~/.zcode the same shape holds a fortiori: the
// client's own state lives under ~/.zcode/cli and ~/.zcode/v2, neither of which is sampled.
// The zcode layer also writes the user's shell rc files (install_zcode_env), so the sample
// covers them too — a global run in a sandboxed HOME that leaked into the real ~/.zshrc must
// fail the comparison.
func realHomeFingerprint(home string) string {
	var b strings.Builder
	for _, d := range []string{".claude", ".zcode"} {
		for _, sub := range []string{"", "skills", "commands", "rules"} {
			p := filepath.Join(home, d, sub)
			if fi, err := os.Stat(p); err != nil || !fi.IsDir() {
				fmt.Fprintf(&b, "absent %s\n", p)
				continue
			}
			fmt.Fprintf(&b, "dir %s\n", p)
			entries, _ := os.ReadDir(p)
			for _, e := range entries {
				fmt.Fprintf(&b, "  %s\n", e.Name())
			}
		}
		for _, f := range []string{"CLAUDE.md", "AGENTS.md", "CLAUDE.md.flow.bak", "AGENTS.md.flow.bak"} {
			b.WriteString(statLine(filepath.Join(home, d, f)))
		}
	}
	for _, f := range []string{".zshrc", ".bashrc"} {
		b.WriteString(statLine(filepath.Join(home, f)))
	}
	return sha256hex(b.String())
}

// treeFingerprint stat-lines every regular file under paths, sorted, skipping any directory
// named __pycache__ (generated, gitignored bytecode — not source).
func treeFingerprint(paths ...string) string {
	var files []string
	for _, root := range paths {
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() && d.Name() == "__pycache__" {
				return filepath.SkipDir
			}
			if d.Type().IsRegular() {
				files = append(files, p)
			}
			return nil
		})
	}
	sort.Strings(files)
	var b strings.Builder
	for _, f := range files {
		b.WriteString(statLine(f))
	}
	return sha256hex(b.String())
}

// sourceTreeFingerprint covers the SOURCE tree, for a reason the bash harness demonstrated on
// itself: a global install fills the sandbox with symlinks pointing back into this repo, so
// any sandbox path a case writes may in fact be a repo path. seedGuard blocks the route that
// was found; this catches the routes that were not. scripts/ and README.md are included
// deliberately: without them the harness cannot detect damage to ITSELF, which is the one
// blind spot a write-through incident would exploit twice.
func sourceTreeFingerprint(root string) string {
	var ps []string
	for _, p := range []string{"skills", "rules", "commands-claude", "scripts", "setup.sh", "CLAUDE.md", "AGENTS.md", "README.md"} {
		ps = append(ps, filepath.Join(root, p))
	}
	return treeFingerprint(ps...)
}

func TestPycacheChurnLeavesFingerprintUnchanged(t *testing.T) {
	g := newGroup(t)
	fx := filepath.Join(g.dir, "pycache-fixture")
	mkdirAll(t, filepath.Join(fx, "scripts/__pycache__"))
	mkdirAll(t, filepath.Join(fx, "skills"))
	src := filepath.Join(fx, "scripts/real.py")
	pyc := filepath.Join(fx, "scripts/__pycache__/real.cpython-314.pyc")
	g.seedFile(src, 0o644, "x\n")
	g.seedFile(pyc, 0o644, "y\n")
	fp := func() string { return treeFingerprint(filepath.Join(fx, "scripts"), filepath.Join(fx, "skills")) }
	before := fp()
	// Simulate a sibling harness's Python invocation rewriting the bytecode cache mid-run — the
	// exact race KAN-369 reports. A fixed future mtime guarantees a changed timestamp even when
	// the whole test runs inside one filesystem-mtime-resolution tick.
	future := time.Date(2030, 1, 1, 0, 0, 0, 0, time.Local)
	if err := os.Chtimes(pyc, future, future); err != nil {
		t.Fatal(err)
	}
	g.assertEq("a .pyc rewrite under __pycache__ leaves the fingerprint unchanged", before, fp())
	// Sanity check: the prune must not blind the assertion to a real source change — a rewrite
	// of an ordinary file in the same tree still has to move the fingerprint.
	if err := os.Chtimes(src, future, future); err != nil {
		t.Fatal(err)
	}
	g.assertNe("a real source-file change still moves the fingerprint", before, fp())
}

// TestContainmentChecksDetectLeaks proves both containment fingerprints still trip on a leak
// into every location they sample, on fake roots under the sandbox — never by writing to the
// real HOME. Every write adds an entry, creates an absent file or grows a file, so no case
// depends on the filesystem's mtime resolution.
func TestContainmentChecksDetectLeaks(t *testing.T) {
	g := newGroup(t)
	home := filepath.Join(g.dir, "fake-home")
	var homeTargets []string
	for _, d := range []string{".claude", ".zcode"} {
		for _, sub := range []string{"", "skills", "commands", "rules"} {
			mkdirAll(t, filepath.Join(home, d, sub))
			homeTargets = append(homeTargets, filepath.Join(d, sub, "leaked-entry"))
		}
		for _, f := range []string{"CLAUDE.md", "AGENTS.md", "CLAUDE.md.flow.bak", "AGENTS.md.flow.bak"} {
			homeTargets = append(homeTargets, filepath.Join(d, f))
		}
	}
	homeTargets = append(homeTargets, ".zshrc", ".bashrc")
	g.seedFile(filepath.Join(home, ".claude/CLAUDE.md"), 0o644, "# mine\n")
	g.seedFile(filepath.Join(home, ".zshrc"), 0o644, "# mine\n")
	for _, rel := range homeTargets {
		before := realHomeFingerprint(home)
		leak(t, filepath.Join(home, rel))
		g.assertNe("leak detection: a write to ~/"+rel+" moves the real-home fingerprint", before, realHomeFingerprint(home))
	}

	src := filepath.Join(g.dir, "fake-repo")
	g.seedFile(filepath.Join(src, "skills/demo/SKILL.md"), 0o644, "# demo\n")
	g.seedFile(filepath.Join(src, "setup.sh"), 0o755, "#!/usr/bin/env bash\n")
	for _, rel := range []string{"skills/demo/leaked.md", "rules/leaked.mdc", "commands-claude/leaked.md", "scripts/leaked.sh",
		"setup.sh", "CLAUDE.md", "AGENTS.md", "README.md"} {
		before := sourceTreeFingerprint(src)
		leak(t, filepath.Join(src, rel))
		g.assertNe("leak detection: a write to "+rel+" moves the source-tree fingerprint", before, sourceTreeFingerprint(src))
	}
}

// leak writes into p — appending when it exists, creating it (and its parent) when not.
func leak(t *testing.T, p string) {
	t.Helper()
	mkdirAll(t, filepath.Dir(p))
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString("leak\n"); err != nil {
		t.Fatal(err)
	}
}
