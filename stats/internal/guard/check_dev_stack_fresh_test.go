package guard

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-check-dev-stack-fresh.sh at b1b0d601, one
// subtest per case label. The harness ran the guard against fixture trees and
// matched substrings of its combined output; here the guard runs in-process,
// and case 3's real HTTP server is an httptest server fetched by a real curl.

// dsfSection is section_with_fp: a minimal section whose commands table
// carries exactly one `fingerprint` row, cell already table-escaped.
func dsfSection(cell string) string {
	return "## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `src/**` |\n\n| Command | Runs |\n|---------|------|\n| `fingerprint` | `" + cell + "` |\n"
}

const dsfFenced = "## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `src/**` |\n\n```markdown\n| Command | Runs |\n|---------|------|\n| `fingerprint` | `exit 7` |\n```"

// dsfRoot is new_root plus write_cfg: a fixture worktree whose
// .flow/project.md holds cfg, and files written beside it.
func dsfRoot(t *testing.T, cfg string, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".flow"), 0o755); err != nil {
		t.Fatal(err)
	}
	files[".flow/project.md"] = cfg + "\n"
	for p, body := range files {
		if err := os.WriteFile(filepath.Join(root, p), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func dsfRun(args ...string) guardResult {
	return runGuard("check-dev-stack-fresh", args, Env{Getenv: os.Getenv, Dir: "/"})
}

func dsfAssert(t *testing.T, r guardResult, rc int, wants ...string) {
	t.Helper()
	if r.rc != rc {
		t.Fatalf("rc = %d, want %d; out:\n%s", r.rc, rc, r.out)
	}
	for _, w := range wants {
		if !strings.Contains(r.out, w) {
			t.Errorf("output lacks %q:\n%s", w, r.out)
		}
	}
}

func TestCheckDevStackFresh(t *testing.T) {
	t.Parallel()
	t.Run("case 1", func(t *testing.T) {
		root := dsfRoot(t, dsfSection("cmp -s served.html built.html"), map[string]string{"served.html": "served\n", "built.html": "served\n"})
		dsfAssert(t, dsfRun(root), 0, "stack-fresh")
	})
	t.Run("case 2", func(t *testing.T) {
		root := dsfRoot(t, dsfSection("cmp -s served.html built.html"), map[string]string{"served.html": "old\n", "built.html": "new\n"})
		dsfAssert(t, dsfRun(root), 1, "stack-stale")
	})
	t.Run("case 3", func(t *testing.T) {
		served := "stale-bundle\n"
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, served) }))
		defer srv.Close()
		root := dsfRoot(t, dsfSection("curl -sf "+srv.URL+"/index.html \\| cmp -s - dist.html"), map[string]string{"dist.html": "fresh-bundle\n"})
		dsfAssert(t, dsfRun(root), 1, "stack-stale")
		served = "fresh-bundle\n"
		dsfAssert(t, dsfRun(root), 0, "stack-fresh")
	})
	t.Run("case 4", func(t *testing.T) {
		root := dsfRoot(t, dsfSection("test -f .flow/project.md"), map[string]string{})
		dsfAssert(t, dsfRun(root), 0, "stack-fresh")
	})
	t.Run("case 5", func(t *testing.T) {
		root := dsfRoot(t, dsfSection(`printf a-b\n \| grep -q b`), map[string]string{})
		dsfAssert(t, dsfRun(root), 0, "printf a-b")
	})
	t.Run("case 6", func(t *testing.T) {
		root := dsfRoot(t, dsfFenced, map[string]string{})
		dsfAssert(t, dsfRun(root), 2, "no readable `fingerprint` row")
	})
	t.Run("case 7", func(t *testing.T) {
		root := dsfRoot(t, dsfFenced+"\n\n| Command | Runs |\n|---------|------|\n| `fingerprint` | `true` |", map[string]string{})
		dsfAssert(t, dsfRun(root), 0, "stack-fresh")
	})
	t.Run("case 8", func(t *testing.T) {
		dsfAssert(t, dsfRun(t.TempDir()), 2, "does not exist")
	})
	t.Run("case 9", func(t *testing.T) {
		root := dsfRoot(t, "## run\n\n- `make build`", map[string]string{})
		dsfAssert(t, dsfRun(root), 2, "no readable `fingerprint` row")
	})
	t.Run("case 10", func(t *testing.T) {
		root := dsfRoot(t, "## visual verification\n\n| Command | Runs |\n|---------|------|\n| `verify` | `npm run test:visual` |", map[string]string{})
		dsfAssert(t, dsfRun(root), 2, "no readable `fingerprint` row")
	})
	t.Run("case 11", func(t *testing.T) {
		root := t.TempDir()
		dsfAssert(t, dsfRun(), 2, "usage")
		dsfAssert(t, dsfRun(root, "extra"), 2)
		dsfAssert(t, dsfRun(root+"/no-such-dir"), 2)
	})
	t.Run("case 12", func(t *testing.T) {
		root := dsfRoot(t, "## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `fingerprint` | `true` |", map[string]string{})
		dsfAssert(t, dsfRun(root), 2)
	})
}
