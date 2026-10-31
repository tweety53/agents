package landed

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// repo is a throwaway git repository the tests commit into.
type repo struct {
	t   *testing.T
	dir string
}

func newRepo(t *testing.T) repo {
	t.Helper()
	r := repo{t, t.TempDir()}
	r.git("init", "-q", "-b", "main")
	return r
}

func (r repo) git(args ...string) {
	r.t.Helper()
	cmd := exec.Command("git", append([]string{"-C", r.dir, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false"}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		r.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func (r repo) write(path, body string) {
	r.t.Helper()
	p := filepath.Join(r.dir, path)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		r.t.Fatal(err)
	}
	r.git("add", path)
}

func (r repo) commit(subject string) { r.git("commit", "-q", "--allow-empty", "-m", subject) }

func (r repo) drop(name string) {
	r.git("rm", "-r", "-q", "spectre/changes/"+name)
	r.commit("chore(spectre): drop " + name)
}

func mustFile(t *testing.T, dir, name, file, want string) {
	t.Helper()
	b, ok, err := File(dir, name, file)
	if err != nil || !ok || string(b) != want {
		t.Fatalf("File(%s, %s) = %q, %v, %v; want %q", name, file, b, ok, err, want)
	}
}

func TestFileReadsTheDropCommitsParent(t *testing.T) {
	r := newRepo(t)
	r.write("spectre/changes/kan-1/tasks.md", "one\n")
	r.commit("plan kan-1")
	r.drop("kan-1")
	mustFile(t, r.dir, "kan-1", "tasks.md", "one\n")

	r.write("spectre/changes/kan-1/tasks.md", "two\n")
	r.commit("restore kan-1")
	r.drop("kan-1")
	mustFile(t, r.dir, "kan-1", "tasks.md", "two\n")
}

func TestFileFallsBackToTheArchiveCommit(t *testing.T) {
	r := newRepo(t)
	r.write("spectre/changes/archive/kan-1/tasks.md", "old\n")
	r.commit("chore(spectre): archive kan-1")
	mustFile(t, r.dir, "kan-1", "tasks.md", "old\n")
}

func TestFileMatchesTheWholeName(t *testing.T) {
	r := newRepo(t)
	r.write("spectre/changes/kan-1/tasks.md", "one\n")
	r.commit("plan kan-1")
	r.drop("kan-1")
	r.write("spectre/changes/kan-10/tasks.md", "ten\n")
	r.commit("plan kan-10")
	r.drop("kan-10")
	mustFile(t, r.dir, "kan-1", "tasks.md", "one\n")
}

func TestFileAbsent(t *testing.T) {
	r := newRepo(t)
	r.write("README", "x\n")
	r.commit("init")
	if b, ok, err := File(r.dir, "kan-1", "tasks.md"); ok || err != nil {
		t.Fatalf("File = %q, %v, %v; want absent", b, ok, err)
	}
}
