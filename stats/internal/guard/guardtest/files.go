package guardtest

import (
	"os"
	"path/filepath"
	"testing"
)

// ReadString is a file's content, "" when it cannot be read.
func ReadString(path string) string {
	b, _ := os.ReadFile(path)
	return string(b)
}

// ProjectRepo is a fresh repository on spectre/<name>, with project.md committed
// when cfg is non-empty. Its path is named after the harness's own sandbox,
// never t.TempDir()'s: that one carries the running subtest's name, and the
// fixture runs under whichever label gets there first — case 3's label holds
// an `=`, which the relayed violation then carried onto stdout and failed
// that case's own "no KEY=value line" check.
func ProjectRepo(t *testing.T, name, cfg string) string {
	t.Helper()
	repo, err := os.MkdirTemp("", "prepare-workspace-test.")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(repo) })
	FixtureRepo(t, repo, RepoShape{})
	WriteFile(t, filepath.Join(repo, "README.md"), "seed\n")
	GitRun(t, repo, "add", "-A")
	GitRun(t, repo, "commit", "-q", "-m", "seed")
	GitRun(t, repo, "checkout", "-q", "-b", "spectre/"+name)
	if cfg != "" {
		WriteFile(t, filepath.Join(repo, ".flow", "project.md"), cfg+"\n")
		GitRun(t, repo, "add", "-A")
		GitRun(t, repo, "commit", "-q", "-m", "configure")
	}
	return repo
}

const ProjectCommands = `
| Command | Runs |
|---------|------|
| ` + "`create`" + ` | ` + "`./scripts/workspace create`" + ` |
| ` + "`remove`" + ` | ` + "`./scripts/workspace remove`" + ` |
`

func Exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
