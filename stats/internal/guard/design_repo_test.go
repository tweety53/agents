package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// The row is read from the worktree's own .flow/project.md: the main
// checkout, on another branch, may declare none.
func TestDesignRepoReadsTheWorktree(t *testing.T) {
	t.Parallel()
	main, wt := t.TempDir(), t.TempDir()
	for dir, row := range map[string]string{main: "", wt: "| `design repo` | `/d/design git@x:d.git` |\n"} {
		mkdir(t, filepath.Join(dir, ".flow"))
		writeFile(t, filepath.Join(dir, ".flow", "project.md"), "## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `mockups` | `m` |\n"+row)
	}
	run := func(args ...string) (string, int) {
		var out, errs bytes.Buffer
		rc := Registry["design-repo"](args, vvEnv(main), &out, &errs)
		return out.String(), rc
	}
	if out, rc := run(wt); rc != 0 || out != "/d/design git@x:d.git\n" {
		t.Errorf("worktree: rc=%d out=%q", rc, out)
	}
	if out, rc := run(main); rc != 1 || out != "" {
		t.Errorf("no row: rc=%d out=%q, want 1", rc, out)
	}
	// A wrong path or an unreadable file is never "no design repo".
	if _, rc := run(filepath.Join(wt, "absent")); rc != 2 {
		t.Errorf("absent worktree: rc=%d, want 2", rc)
	}
	if _, rc := run(); rc != 2 {
		t.Errorf("no argument: rc=%d, want 2", rc)
	}
	bare := t.TempDir()
	if _, rc := run(bare); rc != 1 {
		t.Errorf("no project.md: rc=%d, want 1", rc)
	}
	if err := os.MkdirAll(filepath.Join(bare, ".flow", "project.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, rc := run(bare); rc != 2 {
		t.Errorf("unreadable project.md: rc=%d, want 2", rc)
	}
}
