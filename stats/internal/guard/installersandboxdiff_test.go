package guard

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// isdSetup is the harness's stand-in installer: deterministic output derived
// from its own tree's data files, installed under $HOME the way
// `setup.sh global` installs.
const isdSetup = `#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
mkdir -p "$HOME/.cfg/bin"
ln -s "$DIR/data/rule-a.mdc" "$HOME/.cfg/rule-a.mdc"
{ printf 'begin\n'; cat "$DIR/data/block.txt"; printf 'end\n'; } > "$HOME/AGENTS.md"
printf '#!/bin/sh\necho skill\n' > "$HOME/.cfg/bin/skill.sh"
chmod 755 "$HOME/.cfg/bin/skill.sh"
`

// isdTree builds a fixture tree at dir. Variants: base, content (installed
// block text differs), mode (installed file's exec bit differs), nosetup (no
// installer at all), failing (the installer exits non-zero).
func isdTree(t *testing.T, dir, variant string) string {
	t.Helper()
	setup, block := isdSetup, "managed block v1\n"
	switch variant {
	case "content":
		block = "managed block v2\n"
	case "mode":
		setup += "chmod 700 \"$HOME/.cfg/bin/skill.sh\"\n"
	case "failing":
		setup += "echo install broke; exit 3\n"
	}
	for path, body := range map[string]string{"setup.sh": setup, "data/rule-a.mdc": "rule a\n", "data/block.txt": block} {
		if path == "setup.sh" && variant == "nosetup" {
			continue
		}
		p := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// TestInstallerSandboxDiff ports scripts/test-installer-sandbox-diff.sh at
// e111a883: one subtest per ok: label, the trees and the sandbox under
// t.TempDir(), TMPDIR injected through Env.
func TestInstallerSandboxDiff(t *testing.T) {
	t.Parallel()
	type isdCase struct {
		name string
		run  func(t *testing.T, root string, isd func(tmp string, args ...string) guardResult)
	}
	pair := func(t *testing.T, root, oldV, newV string) (string, string) {
		return isdTree(t, filepath.Join(root, "old"), oldV), isdTree(t, filepath.Join(root, "new"), newV)
	}
	expect := func(t *testing.T, r guardResult, rc int, needles ...string) {
		t.Helper()
		if r.rc != rc {
			t.Fatalf("expected exit %d, got %d: %s", rc, r.rc, r.out)
		}
		for _, n := range needles {
			if !strings.Contains(r.out, n) {
				t.Fatalf("output has no %q: %s", n, r.out)
			}
		}
	}
	for _, c := range []isdCase{
		{"case 1: identical trees exit 0 with the identical verdict", func(t *testing.T, root string, isd func(string, ...string) guardResult) {
			o, n := pair(t, root, "base", "base")
			tmp := t.TempDir()
			r := isd(tmp, o, n)
			expect(t, r, 0, "installer output identical: "+o+" vs "+n)
			if left, _ := os.ReadDir(tmp); len(left) != 0 || r.err != "" {
				t.Fatalf("sandbox not removed (%d left) or stderr %q", len(left), r.err)
			}
		}},
		{"case 2: content difference exits 1 and names the changed managed file", func(t *testing.T, root string, isd func(string, ...string) guardResult) {
			o, n := pair(t, root, "base", "content")
			r := isd(t.TempDir(), o, n)
			expect(t, r, 1, "AGENTS.md", "installer output differs: "+o+" vs "+n)
			if !strings.HasPrefix(r.stdout, "--- ") || !strings.Contains(r.stdout, "\n+f ") {
				t.Fatalf("stdout is not the unified manifest diff: %q", r.stdout)
			}
		}},
		{"case 3: mode difference exits 1 and names the file", func(t *testing.T, root string, isd func(string, ...string) guardResult) {
			o, n := pair(t, root, "base", "mode")
			expect(t, isd(t.TempDir(), o, n), 1, ".cfg/bin/skill.sh")
		}},
		{"case 4: setup.sh missing from the new tree exits 2 and names it", func(t *testing.T, root string, isd func(string, ...string) guardResult) {
			o, n := pair(t, root, "base", "nosetup")
			r := isd(t.TempDir(), o, n)
			expect(t, r, 2)
			if r.err != "ERROR: "+n+" has no setup.sh at its root — the installer cannot be run from it\n" {
				t.Fatalf("stderr %q", r.err)
			}
		}},
		{"case 5: setup.sh missing from the old tree exits 2", func(t *testing.T, root string, isd func(string, ...string) guardResult) {
			o, n := pair(t, root, "nosetup", "base")
			expect(t, isd(t.TempDir(), o, n), 2, "has no setup.sh at its root")
		}},
		{"case 6: identical trees under a glob-metacharacter path exit 0", func(t *testing.T, root string, isd func(string, ...string) guardResult) {
			o, n := pair(t, filepath.Join(root, "glob[1]*"), "base", "base")
			expect(t, isd(t.TempDir(), o, n), 0)
		}},
		{"case 7: unusable TMPDIR exits 2", func(t *testing.T, root string, isd func(string, ...string) guardResult) {
			o, n := pair(t, root, "base", "base")
			bad := filepath.Join(o, "no-such-tmpdir")
			r := isd(bad, o, n)
			expect(t, r, 2, "ERROR: cannot create the sandbox work directory under "+bad)
			if strings.Contains(r.out, "kept for inspection") {
				t.Fatalf("no sandbox exists to keep: %s", r.out)
			}
		}},
		{"case 8: differing run keeps the sandbox, names it, and it exists", func(t *testing.T, root string, isd func(string, ...string) guardResult) {
			o, n := pair(t, root, "base", "content")
			tmp := t.TempDir()
			r := isd(tmp, o, n)
			expect(t, r, 1, "kept for inspection")
			kept := strings.TrimSuffix(r.err[strings.LastIndex(r.err, "kept for inspection: ")+len("kept for inspection: "):], "\n")
			if filepath.Dir(kept) != tmp || !strings.HasPrefix(filepath.Base(kept), "installer-sandbox-diff.") {
				t.Fatalf("kept %q is not a sandbox under %s", kept, tmp)
			}
			for _, f := range []string{"old-home", "new-home", "old-install.log", "new-install.log", "old.manifest", "new.manifest", "manifest.diff"} {
				if _, err := os.Stat(filepath.Join(kept, f)); err != nil {
					t.Fatal(err)
				}
			}
		}},
		{"case 9: one-argument invocation exits 2", func(t *testing.T, root string, isd func(string, ...string) guardResult) {
			r := isd(t.TempDir(), isdTree(t, root, "base"))
			expect(t, r, 2)
			if r.err != "ERROR: Usage: installer-sandbox-diff.sh <old-tree> <new-tree>\n" {
				t.Fatalf("stderr %q", r.err)
			}
		}},
		{"an argument that is not a directory exits 2", func(t *testing.T, root string, isd func(string, ...string) guardResult) {
			missing := filepath.Join(root, "nope")
			r := isd(t.TempDir(), missing, isdTree(t, filepath.Join(root, "new"), "base"))
			if r.rc != 2 || r.err != "ERROR: "+missing+" is not a directory\n" {
				t.Fatalf("rc=%d stderr %q", r.rc, r.err)
			}
		}},
		{"a failing installer exits 2 with its log tail and keeps the sandbox", func(t *testing.T, root string, isd func(string, ...string) guardResult) {
			o, n := pair(t, root, "base", "failing")
			r := isd(t.TempDir(), o, n)
			expect(t, r, 2, "ERROR: setup.sh failed in "+n+" — last 15 lines of ", "new-install.log:\ninstall broke\n", "kept for inspection")
		}},
		{"manifest lines match the script's normalized shape", func(t *testing.T, root string, isd func(string, ...string) guardResult) {
			o, n := pair(t, root, "base", "content")
			r := isd(t.TempDir(), o, n)
			kept := strings.TrimSuffix(r.err[strings.LastIndex(r.err, "kept for inspection: ")+len("kept for inspection: "):], "\n")
			got, err := os.ReadFile(filepath.Join(kept, "old.manifest"))
			if err != nil {
				t.Fatal(err)
			}
			want := "d " + isdMode(t, filepath.Join(kept, "old-home/.cfg")) + " .cfg\n" +
				"d " + isdMode(t, filepath.Join(kept, "old-home/.cfg/bin")) + " .cfg/bin\n" +
				"f " + isdMode(t, filepath.Join(kept, "old-home/.cfg/bin/skill.sh")) + " 3786416187 21 .cfg/bin/skill.sh\n" +
				"l .cfg/rule-a.mdc -> <TREE>/data/rule-a.mdc\n" +
				"f " + isdMode(t, filepath.Join(kept, "old-home/AGENTS.md")) + " 1536719714 27 AGENTS.md\n"
			if string(got) != want {
				t.Fatalf("manifest\n%s\nwant\n%s", got, want)
			}
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			isd := func(tmp string, args ...string) guardResult {
				return runGuard("installer-sandbox-diff", args, Env{Getenv: func(k string) string {
					if k == "TMPDIR" {
						return tmp
					}
					return os.Getenv(k)
				}})
			}
			c.run(t, t.TempDir(), isd)
		})
	}
}

// isdMode is what the script's stat printed for path: the full st_mode in
// octal on Darwin (`stat -f '%p'`), the permission bits elsewhere
// (`stat -c '%a'`).
func isdMode(t *testing.T, path string) string {
	t.Helper()
	stat := exec.Command("stat", "-c", "%a", path)
	if runtime.GOOS == "darwin" {
		stat = exec.Command("stat", "-f", "%p", path)
	}
	out, err := stat.Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}
