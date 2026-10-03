package guard

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// shimLoad matches a flow_guard_exec shim's library line and the
// cannot-answer exit its refusal block names.
var shimLoad = regexp.MustCompile(`(?m)^.*\. ".*/lib/flow-guard\.sh" \|\| \{\n[^\n]*cannot load lib/flow-guard\.sh[^\n]*\n\s*exit ([0-9]+)\n\}`)

// shimExec matches a shim's final line: `flow_guard_exec <name> …` at column 0.
var shimExec = regexp.MustCompile(`(?m)^flow_guard_exec `)

// TestShimMissingLibCannotAnswer runs every shim alone — no lib/ beside it —
// under /bin/bash, which on macOS is 3.2: a bare `. <missing>` there exits 1
// before the `|| { …; exit <code>; }` block runs, so the shim must test the
// library's readability first. Every shim must exit its own cannot-answer
// code and name the missing library.
func TestShimMissingLibCannotAnswer(t *testing.T) {
	t.Parallel()
	if _, err := os.Stat("/bin/bash"); err != nil {
		t.Skip("no /bin/bash")
	}
	shims, err := filepath.Glob("../../../scripts/*.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range shims {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		// A shim is a non-harness script whose last command execs flow-guard;
		// every one must carry the refusal block, so none is skipped silently.
		if strings.HasPrefix(filepath.Base(path), "test-") || !shimExec.Match(src) {
			continue
		}
		m := shimLoad.FindSubmatch(src)
		if m == nil {
			t.Errorf("%s: flow_guard_exec shim without the template's library line and cannot-load block", path)
			continue
		}
		want, _ := strconv.Atoi(string(m[1]))
		abs, err := filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			link := filepath.Join(t.TempDir(), name)
			if err := os.Symlink(abs, link); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("/bin/bash", link)
			cmd.Env = append(os.Environ(), "FLOW_GUARD_TELEMETRY=off")
			out, _ := cmd.CombinedOutput()
			if rc := cmd.ProcessState.ExitCode(); rc != want || !strings.Contains(string(out), "cannot load lib/flow-guard.sh") {
				t.Errorf("rc=%d want %d\n%s", rc, want, out)
			}
		})
	}
}
