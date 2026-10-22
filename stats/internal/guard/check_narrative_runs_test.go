package guard

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckNarrativeRuns carries the cases of the retired
// scripts/test-check-narrative-runs.sh, each run through the Go guard and
// through the bash at 8c54d6d3 over the same stub `flow` and narrative,
// asserting the harness's expectation and byte-for-byte parity.
func TestCheckNarrativeRuns(t *testing.T) {
	t.Parallel()
	script := filepath.Join(t.TempDir(), "check-narrative-runs.sh")
	writeExec(t, script, string(historicalFile(t, "../../..", "8c54d6d3", "scripts/check-narrative-runs.sh")))

	type fx struct {
		fix, integrate *string // nil: the store read fails
		narrative      *string // nil: no narrative file
		unreadable     bool
		args           []string // nil: <work> demo <work>/narrative.md
		rc             int
		prefix         string // on the combined output
	}
	s := func(v string) *string { return &v }
	cases := []struct {
		label string
		fx    fx
	}{
		{"a fix run with no section exits 1", fx{fix: s("1"), integrate: s("0"), narrative: s("## 2026-10-19 — implement run"), rc: 1, prefix: "NARRATIVE-RUNS-MISSING:"}},
		{"an integrate run with no section exits 1", fx{fix: s("1"), integrate: s("1"), narrative: s("## 2026-10-19 — fix run"), rc: 1, prefix: "NARRATIVE-RUNS-MISSING:"}},
		{"every run with its section exits 0", fx{fix: s("2"), integrate: s("1"), narrative: s("## 2026-10-18 — fix run\n## 2026-10-19 — fix run\n## 2026-10-19 — integrate run"), rc: 0, prefix: "NARRATIVE-RUNS-OK:"}},
		{"an absent narrative with no runs exits 0", fx{fix: s("0"), integrate: s("0"), rc: 0, prefix: "NARRATIVE-RUNS-OK:"}},
		{"a failed store read exits 2", fx{fix: s("0"), narrative: s("## x"), rc: 2, prefix: "check-narrative-runs: flow stage sessions -stage flow.preserve-sessions failed"}},
		{"a non-integer store read exits 2", fx{fix: s("two"), integrate: s("0"), rc: 2, prefix: "check-narrative-runs: flow stage sessions -stage flow.document-fix printed 'two', not one integer"}},
		{"an unreadable narrative exits 2", fx{fix: s("0"), integrate: s("0"), narrative: s("## x"), unreadable: true, rc: 2, prefix: "check-narrative-runs: cannot read "}},
		{"a usage error exits 2", fx{fix: s("0"), integrate: s("0"), args: []string{"only-one"}, rc: 2, prefix: "usage: check-narrative-runs.sh"}},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			work := t.TempDir()
			bin := filepath.Join(work, "bin")
			// flow stage sessions -C <wt> -stage <key> <change>: prints <work>/<key>, or fails when absent.
			writeExec(t, filepath.Join(bin, "flow"), "#!/usr/bin/env bash\n[ \"$1 $2\" = \"stage sessions\" ] || exit 9\ncat \""+work+"/$6\" 2>/dev/null || exit 1\n")
			if c.fx.fix != nil {
				writeFile(t, filepath.Join(work, "flow.document-fix"), *c.fx.fix+"\n")
			}
			if c.fx.integrate != nil {
				writeFile(t, filepath.Join(work, "flow.preserve-sessions"), *c.fx.integrate+"\n")
			}
			nf := filepath.Join(work, "narrative.md")
			if c.fx.narrative != nil {
				writeFile(t, nf, *c.fx.narrative+"\n")
			}
			if c.fx.unreadable {
				if err := os.Chmod(nf, 0); err != nil {
					t.Fatal(err)
				}
				if f, err := os.Open(nf); err == nil {
					f.Close()
					t.Skip("running as a user who reads a mode-0 file")
				}
			}
			args := c.fx.args
			if args == nil {
				args = []string{work, "demo", nf}
			}
			path := bin + ":" + os.Getenv("PATH")

			var gout, gerr bytes.Buffer
			grc := Registry["check-narrative-runs"](args, Env{Getenv: pathEnv(bin), Dir: work}, &gout, &gerr)

			cmd := exec.Command("bash", append([]string{script}, args...)...)
			cmd.Dir = work
			cmd.Env = append(os.Environ(), "PATH="+path)
			var bout, berr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &bout, &berr
			brc := 0
			if err := cmd.Run(); err != nil {
				var ee *exec.ExitError
				if !errors.As(err, &ee) {
					t.Fatal(err)
				}
				brc = ee.ExitCode()
			}

			if grc != c.fx.rc || !strings.HasPrefix(gout.String()+gerr.String(), c.fx.prefix) {
				t.Errorf("rc=%d stdout=%q stderr=%q, want rc=%d and output starting %q", grc, gout.String(), gerr.String(), c.fx.rc, c.fx.prefix)
			}
			if grc != brc || gout.String() != bout.String() || gerr.String() != berr.String() {
				t.Errorf("port rc=%d stdout=%q stderr=%q\nbash rc=%d stdout=%q stderr=%q", grc, gout.String(), gerr.String(), brc, bout.String(), berr.String())
			}
		})
	}
}
