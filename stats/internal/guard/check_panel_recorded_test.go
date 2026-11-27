package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckPanelRecorded: a report that raises findings with no panel row in the store fails; a
// recorded panel row, or an all-zero report, passes.
func TestCheckPanelRecorded(t *testing.T) {
	t.Parallel()
	run := func(report, rows string) int {
		env, _, _ := awaitRowsEnv(t, rows, rows)
		writeFile(t, filepath.Join(env.Dir, "rows.json"), rows)
		wt := filepath.Join(t.TempDir(), "chg")
		writeFile(t, filepath.Join(wt, ".superpowers/sdd/panel-report-0-primary.md"), report)
		writeExec(t, filepath.Join(env.Dir, "bin", "flow"), "#!/usr/bin/env bash\ncat "+env.Dir+"/rows.json\n")
		var o, e bytes.Buffer
		return checkPanelRecorded([]string{wt, "0"}, env, &o, &e)
	}
	if rc := run("Critical: 0\nImportant: 1\nMinor: 0\n", `[]`); rc != 1 {
		t.Errorf("unrecorded: rc %d, want 1", rc)
	}
	if rc := run("Critical: 0\nImportant: 1\nMinor: 0\n", `[{"round":0,"slot":"primary"}]`); rc != 0 {
		t.Errorf("recorded: rc %d, want 0", rc)
	}
	if rc := run("Critical: 1\nImportant: 2\nMinor: 0\n", `[{"round":0,"slot":"primary"},{"round":1,"slot":"primary"}]`); rc != 1 {
		t.Errorf("partly recorded: rc %d, want 1", rc)
	}
	if rc := run("Critical: 0\nImportant: 0\nMinor: 0\n", `[]`); rc != 0 {
		t.Errorf("nothing raised: rc %d, want 0", rc)
	}
}

// TestCheckPanelRecordedReadsOnlyItsOwnChange: the guard reads the worktree's own change's rows
// alone, so a finished change's history -- the gated review's findings 0120 folded into the
// primary slot -- never counts toward an open change's round.
func TestCheckPanelRecordedReadsOnlyItsOwnChange(t *testing.T) {
	t.Parallel()
	env, _, _ := awaitRowsEnv(t, `[]`, `[]`)
	wt := filepath.Join(t.TempDir(), "open-change")
	writeFile(t, filepath.Join(wt, ".superpowers/sdd/panel-report-0-primary.md"), "Important: 1\n")
	args := filepath.Join(env.Dir, "args")
	writeExec(t, filepath.Join(env.Dir, "bin", "flow"), "#!/usr/bin/env bash\nprintf '%s\\n' \"$*\" >"+args+"\necho '[]'\n")
	var o, e bytes.Buffer
	if rc := checkPanelRecorded([]string{wt, "0"}, env, &o, &e); rc != 1 {
		t.Errorf("rc %d, want 1: the open change holds no row of its own", rc)
	}
	if got, _ := os.ReadFile(args); !strings.Contains(string(got), "-change open-change ") {
		t.Errorf("flow called with %q, want -change open-change", got)
	}
}
