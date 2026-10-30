package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/tweety53/agents/stats/internal/fallback"
)

// requestDaemon is a fake flowd for the request verb: it records every call
// it serves and answers a stage-run read with runs.
type requestDaemon struct {
	mu     sync.Mutex
	calls  []string
	bodies [][]byte
	runs   string
}

func (d *requestDaemon) handler(t *testing.T) http.HandlerFunc {
	return genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		body, err := readAll(r)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		d.mu.Lock()
		d.calls = append(d.calls, r.Method+" "+r.URL.Path)
		d.bodies = append(d.bodies, body)
		runs := d.runs
		d.mu.Unlock()
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"total":0,"stageRuns":` + runs + `}`))
			return
		}
		_, _ = w.Write([]byte(`{"n":1}`))
	})
}

func (d *requestDaemon) reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.calls, d.bodies = nil, nil
}

func (d *requestDaemon) count(prefix string) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	n := 0
	for _, c := range d.calls {
		if strings.HasPrefix(c, prefix) {
			n++
		}
	}
	return n
}

// requestFixture is a peer checkout on branch <name> carrying
// .superpowers/sdd/<name>/, with a bare origin whose origin/HEAD is main and
// a .flow/project.md declaring `ui paths` `web/**`.
type requestFixture struct {
	peer, name string
	git        func(args ...string) string
}

func newRequestFixture(t *testing.T) requestFixture {
	t.Helper()
	isolatedStateRoot(t)
	root := t.TempDir()
	peer, origin := filepath.Join(root, "peer"), filepath.Join(root, "origin.git")
	run := func(dir string, args ...string) string {
		t.Helper()
		out, err := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	const name = "interactive-20261030-1000"
	run(root, "init", "-q", "--bare", "-b", "main", origin)
	run(root, "init", "-q", "-b", "main", peer)
	fx := requestFixture{peer: peer, name: name, git: func(args ...string) string { return run(peer, args...) }}
	fx.writeProject(t, "## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `web/**` |\n")
	fx.git("add", ".flow")
	fx.git("commit", "-q", "-m", "base")
	fx.git("remote", "add", "origin", origin)
	fx.git("push", "-q", "origin", "main")
	fx.git("remote", "set-head", "origin", "main")
	fx.git("checkout", "-q", "-b", name)
	if err := os.MkdirAll(filepath.Join(peer, ".superpowers/sdd", name), 0o755); err != nil {
		t.Fatal(err)
	}
	return fx
}

func (fx requestFixture) writeProject(t *testing.T, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(fx.peer, ".flow"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fx.peer, ".flow/project.md"), []byte("# p\n\n"+body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// commit commits path on the session branch and, when push, lands it on
// origin/main; it returns the full sha.
func (fx requestFixture) commit(t *testing.T, path string, push bool) string {
	t.Helper()
	full := filepath.Join(fx.peer, path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(path+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fx.git("add", path)
	fx.git("commit", "-q", "-m", "feat: "+path)
	if push {
		fx.git("push", "-q", "origin", fx.name+":main")
		fx.git("fetch", "-q", "origin")
	}
	return fx.git("rev-parse", "HEAD")
}

func runRequest(addr, dir string, args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	full := append([]string{"record", "request", args[0], "-addr", addr, "-timeout", "2s", "-C", dir}, args[1:]...)
	code := run(context.Background(), full, strings.NewReader(""), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// TestRecordRequestBeginEnd pins `flow record request begin|end`'s wire
// and usage (KAN-1061): what each half posts, every usage error, and the
// journal a dead store leaves.
func TestRecordRequestBeginEnd(t *testing.T) {
	fx := newRequestFixture(t)
	d := &requestDaemon{runs: "[]"}
	srv := httptest.NewServer(d.handler(t))
	defer srv.Close()

	code, stdout, stderr := runRequest(srv.URL, fx.peer, "begin", "-change", fx.name, "-n", "1", "-title", "Fix the pill")
	if code != 0 || !strings.Contains(stdout, "recorded: request 1") {
		t.Fatalf("begin exit=%d stdout=%q stderr=%q, want 0 printing recorded: request 1", code, stdout, stderr)
	}
	if len(d.calls) != 1 || d.calls[0] != "POST /api/v1/records/"+mustProjectKey(t, fx.peer)+"/"+fx.name+"/requests" {
		t.Fatalf("calls = %v, want one POST to the requests route", d.calls)
	}
	var got map[string]any
	if err := json.Unmarshal(d.bodies[0], &got); err != nil {
		t.Fatalf("decode %s: %v", d.bodies[0], err)
	}
	if got["n"] != float64(1) || got["title"] != "Fix the pill" || got["receivedAt"] == nil {
		t.Errorf("begin body = %s, want n, title and receivedAt", d.bodies[0])
	}

	d.reset()
	if code, _, stderr := runRequest(srv.URL, fx.peer, "end", "-change", fx.name, "-n", "1", "-outcome", "answered"); code != 0 {
		t.Fatalf("end answered exit=%d stderr=%q, want 0", code, stderr)
	}
	if err := json.Unmarshal(d.bodies[0], &got); err != nil {
		t.Fatalf("decode %s: %v", d.bodies[0], err)
	}
	if got["outcome"] != "answered" || got["commitSha"] != nil || got["baseSha"] != nil || got["endedAt"] == nil {
		t.Errorf("end body = %s, want answered with no shas and an endedAt", d.bodies[0])
	}

	sha := strings.Repeat("a", 40)
	d.reset()
	for name, args := range map[string][]string{
		"begin without title or reason": {"begin", "-change", fx.name, "-n", "1"},
		"begin -n 0":                    {"begin", "-change", fx.name, "-n", "0", "-title", "t"},
		"begin without -change":         {"begin", "-n", "1", "-title", "t"},
		"answered with a commit":        {"end", "-change", fx.name, "-n", "1", "-outcome", "answered", "-commit", sha},
		"an unknown outcome":            {"end", "-change", fx.name, "-n", "1", "-outcome", "merged"},
		"landed without -base":          {"end", "-change", fx.name, "-n", "1", "-outcome", "landed", "-commit", sha},
	} {
		if code, _, stderr := runRequest(srv.URL, fx.peer, args...); code != 2 {
			t.Errorf("%s: exit=%d stderr=%q, want 2", name, code, stderr)
		}
	}
	if len(d.calls) != 0 {
		t.Errorf("a usage error reached the daemon: %v", d.calls)
	}

	dead := deadPortAddr(t)
	code, _, stderr = runRequest(dead, fx.peer, "begin", "-change", fx.name, "-n", "2", "-title", "Queued")
	if code != 0 || !strings.Contains(stderr, "wrote local journal") {
		t.Fatalf("begin on a dead store exit=%d stderr=%q, want 0 with the journal warning", code, stderr)
	}
	entries, err := fallback.ReadJournalEntries(recordJournalPath(mustProjectKey(t, fx.peer), fx.name))
	if err != nil || len(entries) != 1 || !strings.Contains(string(entries[0].Body), `"kind":"request-begin"`) {
		t.Fatalf("journal entries = %v (%v), want one request-begin", entries, err)
	}
}

// TestRecordRequestEndLandedGuard pins design.md's Request-end guard: a
// landed end is refused before any write unless its commit is on
// origin/<base> and, when its range touches `ui paths`, the request's latest
// flow.visual-verify run ended completed.
func TestRecordRequestEndLandedGuard(t *testing.T) {
	fx := newRequestFixture(t)
	base := fx.git("rev-parse", "HEAD")
	docs := fx.commit(t, "docs/a.md", true)
	ui := fx.commit(t, "web/x.ts", true)
	unpushed := fx.commit(t, "docs/b.md", false)

	d := &requestDaemon{runs: "[]"}
	srv := httptest.NewServer(d.handler(t))
	defer srv.Close()

	end := func(commit, from string) (int, string) {
		t.Helper()
		d.reset()
		code, _, stderr := runRequest(srv.URL, fx.peer, "end", "-change", fx.name, "-n", "3", "-outcome", "landed", "-commit", commit, "-base", from)
		return code, stderr
	}
	refused := func(t *testing.T, code int, stderr string, want string) {
		t.Helper()
		if code != 1 || !strings.Contains(stderr, want) {
			t.Errorf("exit=%d stderr=%q, want 1 naming %q", code, stderr, want)
		}
		if n := d.count("POST"); n != 0 {
			t.Errorf("a refused end reached the daemon with %d write(s)", n)
		}
	}
	const vv = `{"stageRunId":%d,"command":"/flow-interactive","stage":"flow.visual-verify","request":%d,"startedAt":%q,"outcome":%q}`
	runs := func(rows ...string) string { return "[" + strings.Join(rows, ",") + "]" }
	row := func(id, request int, started, outcome string) string {
		return fmt.Sprintf(vv, id, request, started, outcome)
	}

	t.Run("a commit not on the base", func(t *testing.T) {
		code, stderr := end(unpushed, docs)
		refused(t, code, stderr, "origin/main")
	})
	t.Run("a base not an ancestor of the commit", func(t *testing.T) {
		code, stderr := end(docs, ui)
		refused(t, code, stderr, "-base")
	})
	t.Run("a range touching no ui path", func(t *testing.T) {
		if code, stderr := end(docs, base); code != 0 {
			t.Fatalf("exit=%d stderr=%q, want 0", code, stderr)
		}
		if n := d.count("GET"); n != 0 {
			t.Errorf("a non-UI range read %d stage-run list(s), want none", n)
		}
	})
	t.Run("a ui range with no visual-verify run", func(t *testing.T) {
		d.runs = "[]"
		code, stderr := end(ui, docs)
		refused(t, code, stderr, "UI check")
	})
	t.Run("a ui range whose latest run failed", func(t *testing.T) {
		d.runs = runs(row(1, 3, "2026-10-30T10:00:00Z", "completed"), row(2, 3, "2026-10-30T10:05:00.5Z", "failed"))
		code, stderr := end(ui, docs)
		refused(t, code, stderr, "UI check")
	})
	t.Run("a ui range whose latest run completed", func(t *testing.T) {
		d.runs = runs(row(1, 3, "2026-10-30T10:00:00Z", "failed"), row(2, 3, "2026-10-30T10:05:00Z", "completed"))
		if code, stderr := end(ui, docs); code != 0 {
			t.Fatalf("exit=%d stderr=%q, want 0", code, stderr)
		}
		if n := d.count("POST"); n != 1 {
			t.Errorf("writes = %d, want 1", n)
		}
	})
	t.Run("a ui range checked only for another request", func(t *testing.T) {
		d.runs = runs(row(1, 2, "2026-10-30T10:00:00Z", "completed"))
		code, stderr := end(ui, docs)
		refused(t, code, stderr, "UI check")
	})
	t.Run("the store down", func(t *testing.T) {
		dead := deadPortAddr(t)
		code, _, stderr := runRequest(dead, fx.peer, "end", "-change", fx.name, "-n", "3", "-outcome", "landed", "-commit", unpushed, "-base", docs)
		if code != 1 || !strings.Contains(stderr, "origin/main") {
			t.Errorf("an unpushed commit on a dead store: exit=%d stderr=%q, want 1", code, stderr)
		}
		code, _, stderr = runRequest(dead, fx.peer, "end", "-change", fx.name, "-n", "3", "-outcome", "landed", "-commit", ui, "-base", docs)
		if code != 0 || !strings.Contains(stderr, "wrote local journal") {
			t.Fatalf("a UI end on a dead store: exit=%d stderr=%q, want 0 journalled", code, stderr)
		}
		warn := 0
		for _, line := range strings.Split(stderr, "\n") {
			if strings.HasPrefix(line, "⚠") && strings.Contains(line, "UI check") {
				warn++
			}
		}
		if warn != 1 {
			t.Errorf("stderr = %q, want one ⚠ line naming the skipped UI-check read", stderr)
		}
	})
	t.Run("no visual verification section", func(t *testing.T) {
		fx.writeProject(t, "## lint\n")
		defer fx.writeProject(t, "## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `web/**` |\n")
		d.runs = "[]"
		if code, stderr := end(ui, docs); code != 0 {
			t.Errorf("exit=%d stderr=%q, want 0", code, stderr)
		}
	})
	t.Run("an empty ui paths", func(t *testing.T) {
		fx.writeProject(t, "## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | |\n")
		defer fx.writeProject(t, "## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `web/**` |\n")
		code, stderr := end(ui, docs)
		refused(t, code, stderr, "`ui paths` is absent or empty")
	})
	t.Run("no checkout carrying the change", func(t *testing.T) {
		d.reset()
		code, _, stderr := runRequest(srv.URL, fx.peer, "end", "-change", "interactive-other", "-n", "1", "-outcome", "landed", "-commit", docs, "-base", base)
		refused(t, code, stderr, ".superpowers/sdd/interactive-other")
	})
}

func mustProjectKey(t *testing.T, dir string) string {
	t.Helper()
	key, _, err := fallback.ProjectKey(dir)
	if err != nil {
		t.Fatalf("project key: %v", err)
	}
	return key
}
