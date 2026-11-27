package guard

import (
	"bytes"
	"path/filepath"
	"testing"
)

// TestCheckPanelRecorded: a report that raises findings with no panel row in the store fails; a
// recorded panel row, or an all-zero report, passes.
func TestCheckPanelRecorded(t *testing.T) {
	t.Parallel()
	run := func(report, rows string) int {
		env, _ := awaitRowsEnv(t, rows, rows)
		wt := filepath.Join(t.TempDir(), "chg")
		writeFile(t, filepath.Join(wt, ".superpowers/sdd/panel-report-0-primary.md"), report)
		writeExec(t, filepath.Join(env.Dir, "bin", "flow"), "#!/usr/bin/env bash\ncat "+env.Dir+"/rows.json\n")
		var o, e bytes.Buffer
		return checkPanelRecorded([]string{wt, "0"}, env, &o, &e)
	}
	gated := `[{"round":0,"slot":"gated-reviewer"}]`
	if rc := run("Critical: 0\nImportant: 1\nMinor: 0\n", gated); rc != 1 {
		t.Errorf("unrecorded: rc %d, want 1", rc)
	}
	if rc := run("Critical: 0\nImportant: 1\nMinor: 0\n", `[{"round":0,"slot":"primary"}]`); rc != 0 {
		t.Errorf("recorded: rc %d, want 0", rc)
	}
	if rc := run("Critical: 1\nImportant: 2\nMinor: 0\n", `[{"round":0,"slot":"primary"},{"round":1,"slot":"primary"}]`); rc != 1 {
		t.Errorf("partly recorded: rc %d, want 1", rc)
	}
	if rc := run("Critical: 0\nImportant: 0\nMinor: 0\n", gated); rc != 0 {
		t.Errorf("nothing raised: rc %d, want 0", rc)
	}
}
