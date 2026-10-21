package setuptest

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/synctest"
)

// The harness's own safety logic, pinned: a refusal, a cleanup bound or a close-out verdict that
// quietly stopped working would leave every group green while the harness protected nothing.

func TestSetupRefusalKeepsTheInstallerInTheSandbox(t *testing.T) {
	in := filepath.Join(sandbox, "some-home")
	for _, c := range []struct {
		name, home, proj string
		refused          bool
	}{
		{"both inside", in, filepath.Join(sandbox, "proj"), false},
		{"HOME outside", realHome, filepath.Join(sandbox, "proj"), true},
		{"project outside", in, "/tmp/elsewhere", true},
		{"the sandbox itself as HOME", sandbox, filepath.Join(sandbox, "proj"), true},
	} {
		if err := setupRefusal(c.home, c.proj); (err != nil) != c.refused {
			t.Errorf("%s: setupRefusal(%s, %s) = %v, want refused=%v", c.name, c.home, c.proj, err, c.refused)
		}
	}
}

func TestSeedRefusalRefusesOutsidePathsAndSymlinks(t *testing.T) {
	g := newGroup(t)
	real := filepath.Join(g.dir, "real")
	mkdirAll(t, real)
	if err := os.Symlink(real, filepath.Join(g.dir, "linked-dir")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(real, "f"), filepath.Join(g.dir, "linked-file")); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name, path string
		refused    bool
	}{
		{"a plain sandbox path", filepath.Join(real, "f"), false},
		{"a path outside the sandbox", filepath.Join(realHome, ".claude/CLAUDE.md"), true},
		{"a symlink at the path", filepath.Join(g.dir, "linked-file"), true},
		{"a symlinked parent", filepath.Join(g.dir, "linked-dir", "f"), true},
	} {
		if err := seedRefusal(c.path); (err != nil) != c.refused {
			t.Errorf("%s: seedRefusal(%s) = %v, want refused=%v", c.name, c.path, err, c.refused)
		}
	}
}

func TestCleanupRemovesOnlyHarnessSandboxes(t *testing.T) {
	for p, want := range map[string]bool{
		"/tmp/flow-test-setup.123": true,
		"/tmp":                     false,
		"/":                        false,
		"/tmp/other":               false,
		realHome:                   false,
	} {
		if got := sandboxRemovable(p); got != want {
			t.Errorf("sandboxRemovable(%q) = %v, want %v", p, got, want)
		}
	}
}

func TestCloseOutFailsOnAnyMovedFingerprint(t *testing.T) {
	same := fpCheck{"unchanged", "a", func() string { return "a" }}
	moved := fpCheck{"moved", "a", func() string { return "b" }}
	var out bytes.Buffer
	if got := closeOut(&out, []fpCheck{same, same}); got != 0 {
		t.Errorf("closeOut over unchanged fingerprints = %d, want 0", got)
	}
	if got := closeOut(&out, []fpCheck{same, moved}); got != 1 {
		t.Errorf("closeOut with a moved fingerprint = %d, want 1", got)
	}
	if !strings.Contains(out.String(), "✗ moved") {
		t.Errorf("closeOut did not name the moved fingerprint:\n%s", out.String())
	}
	for _, c := range []struct{ test, close, want int }{{0, 0, 0}, {0, 1, 1}, {1, 0, 1}, {1, 1, 1}, {2, 1, 2}} {
		if got := exitCode(c.test, c.close); got != c.want {
			t.Errorf("exitCode(%d, %d) = %d, want %d", c.test, c.close, got, c.want)
		}
	}
}

func TestCountLinesMatchingCountsWholeLines(t *testing.T) {
	g := newGroup(t)
	f := filepath.Join(g.dir, "lines")
	g.seedFile(f, 0o644, "MARK\nMARK\nxMARKx\nMARK \n")
	if got := countLinesMatching(f, "MARK"); got != 2 {
		t.Errorf("countLinesMatching = %d, want 2 (whole lines only, every one counted)", got)
	}
	if got := countLinesMatching(filepath.Join(g.dir, "absent"), "MARK"); got != 0 {
		t.Errorf("countLinesMatching on a missing file = %d, want 0", got)
	}
}

func TestWritersGateHoldsCloseUntilEveryWriterLeaves(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var w gate
		if !w.enter() {
			t.Fatal("an open gate refused a writer")
		}
		closed := false
		go func() { w.close(); closed = true }()
		synctest.Wait()
		if closed {
			t.Fatal("close returned while an admitted writer was still in")
		}
		if w.enter() {
			t.Fatal("a closing gate admitted a new writer")
		}
		w.leave()
		synctest.Wait()
		if !closed {
			t.Fatal("close did not return after the last writer left")
		}
	})
}
