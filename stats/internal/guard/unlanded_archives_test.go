package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func uaRun(dir string, args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	env := Env{
		Dir:       dir,
		Getenv:    func(k string) string { v, _ := vmLookup(dir)(k); return v },
		LookupEnv: vmLookup(dir),
	}
	rc := unlandedArchivesCmd(args, env, &out, &errb)
	return rc, out.String(), errb.String()
}

// TestUnlandedArchives pins the shared rule every repo-wide scanner applies
// to an archived change: listed only while the base lacks its directory —
// archived on this branch, committed or not — never once the base carries
// it, and never through a symlink.
func TestUnlandedArchives(t *testing.T) {
	t.Parallel()
	root := vmRepo(t, map[string]string{"spectre/changes/archive/landed/tasks.md": "- [x] 1. old\n"})
	gitRun(t, root, "checkout", "-q", "-b", "spectre/demo")
	writeFile(t, filepath.Join(root, "spectre/changes/archive/committed/tasks.md"), "- [x] 1. done\n")
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "commit", "-q", "-m", "chore(spectre): archive committed")
	writeFile(t, filepath.Join(root, "spectre/changes/archive/untracked/tasks.md"), "- [x] 1. done\n")
	writeFile(t, filepath.Join(root, "elsewhere/tasks.md"), "- [x] 1. outside\n")
	if err := os.Symlink("../../../elsewhere", filepath.Join(root, "spectre/changes/archive/linked")); err != nil {
		t.Fatal(err)
	}

	rc, out, errb := uaRun(root, root)
	if rc != 0 || out != "committed\nuntracked\n" {
		t.Fatalf("rc = %d, stdout = %q, want 0 and the two unlanded archives\nstderr:\n%s", rc, out, errb)
	}
}

func TestUnlandedArchivesNoBaseListsNothing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "spectre/changes/archive/old/tasks.md"), "- [x] 1. old\n")
	if rc, out, errb := uaRun(dir, dir); rc != 0 || out != "" {
		t.Fatalf("not a git worktree: rc = %d, stdout = %q, want 0 and nothing\nstderr:\n%s", rc, out, errb)
	}
}

func TestUnlandedArchivesUsage(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{nil, {""}, {"a", "b"}} {
		if rc, _, _ := uaRun(t.TempDir(), args...); rc != 2 {
			t.Errorf("args %q: rc = %d, want 2", args, rc)
		}
	}
}

// TestUnlandedArchivesUnreadable pins the one refusal past usage: an
// archive directory that exists but cannot be listed is exit 2 — and, in
// check-plan-provenance, a scan-integrity failure — never "no archived
// changes".
func TestUnlandedArchivesUnreadable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root reads a mode-000 directory")
	}
	t.Parallel()
	root := vmRepo(t, map[string]string{"spectre/changes/live/tasks.md": ppClean})
	archive := filepath.Join(root, "spectre/changes/archive")
	writeFile(t, filepath.Join(archive, "new/tasks.md"), ppClean)
	ppChmod(t, archive, 0)
	t.Cleanup(func() { _ = os.Chmod(archive, 0o755) })

	if rc, out, errb := uaRun(root, root); rc != 2 {
		t.Errorf("unlanded-archives: rc = %d, want 2\nstdout:\n%s\nstderr:\n%s", rc, out, errb)
	}
	r := ppRun(root)
	ok(t, "check-plan-provenance: an unlistable archive directory exits 2", r.rc == 2 && has(r.out, "cannot list its archived changes"), r)
}
