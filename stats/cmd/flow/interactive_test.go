package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// TestInteractiveLand pins `flow interactive land` (kan-1081 self-review,
// F4): one call rebases, checks the scratch and the lint stamp on the
// rebased HEAD, and pushes, marking flow.landing-routes around the push.
func TestInteractiveLand(t *testing.T) {
	smoke.Heavy(t)
	type post struct {
		path string
		body map[string]any
	}
	type fixture struct {
		wt, origin string
		git        func(dir string, args ...string) string
		posts      func() []post
		srv        *httptest.Server
	}
	newFixture := func(t *testing.T) fixture {
		t.Helper()
		isolatedStateRoot(t)
		root, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		origin, wt := filepath.Join(root, "origin.git"), filepath.Join(root, "wt")
		git := func(dir string, args ...string) string {
			t.Helper()
			out, err := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...).CombinedOutput()
			if err != nil {
				t.Fatalf("git %v: %v\n%s", args, err, out)
			}
			return strings.TrimSpace(string(out))
		}
		git(root, "init", "-q", "--bare", "-b", "main", origin)
		git(root, "init", "-q", "-b", "main", wt)
		writeTestFile(t, filepath.Join(wt, ".gitignore"), ".superpowers/\n")
		writeTestFile(t, filepath.Join(wt, ".flow/project.md"), "# p\n\n## lint\n\n```bash\ntrue\n```\n")
		git(wt, "add", "-A")
		git(wt, "commit", "-q", "-m", "base")
		git(wt, "remote", "add", "origin", origin)
		git(wt, "push", "-q", "origin", "main")
		git(wt, "remote", "set-head", "origin", "main")
		git(wt, "checkout", "-q", "-b", "kan-1-x")
		if err := os.MkdirAll(filepath.Join(wt, ".superpowers/sdd/kan-1-x"), 0o755); err != nil {
			t.Fatal(err)
		}

		var mu sync.Mutex
		var got []post
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			b, _ := readAll(r)
			var body map[string]any
			_ = json.Unmarshal(b, &body)
			mu.Lock()
			got = append(got, post{r.URL.Path, body})
			mu.Unlock()
			_, _ = w.Write([]byte(`{"stageRunId":1,"attempt":1}`))
		}))
		t.Cleanup(srv.Close)
		return fixture{wt: wt, origin: origin, git: git, srv: srv, posts: func() []post {
			mu.Lock()
			defer mu.Unlock()
			return append([]post(nil), got...)
		}}
	}
	commit := func(t *testing.T, fx fixture, path string) string {
		t.Helper()
		writeTestFile(t, filepath.Join(fx.wt, path), path+"\n")
		fx.git(fx.wt, "add", path)
		fx.git(fx.wt, "commit", "-q", "-m", "feat: "+path)
		return fx.git(fx.wt, "rev-parse", "HEAD")
	}
	stamp := func(t *testing.T, fx fixture, sha string) {
		t.Helper()
		writeTestFile(t, fx.git(fx.wt, "rev-parse", "--path-format=absolute", "--git-path", "flow-lint-green"), sha+"\n")
	}
	land := func(fx fixture, extra ...string) (int, string, string) {
		args := []string{"interactive", "land", "-addr", fx.srv.URL, "-timeout", "2s", "-C", fx.wt}
		if extra == nil {
			extra = []string{"-change", "kan-1-x", "-n", "3", "-harness", "claude-code", "-session-token", "fi-k1-20261109T100000Z"}
		}
		var o, e bytes.Buffer
		code := run(context.Background(), append(args, extra...), strings.NewReader(""), &o, &e)
		return code, o.String(), e.String()
	}
	originMain := func(fx fixture) string { return fx.git(fx.origin, "rev-parse", "main") }
	refused := func(t *testing.T, fx fixture, before string, code int, stderr string, want ...string) {
		t.Helper()
		if code != 1 {
			t.Errorf("exit=%d stderr=%q, want 1", code, stderr)
		}
		for _, w := range want {
			if !strings.Contains(stderr, w) {
				t.Errorf("stderr=%q, want it naming %q", stderr, w)
			}
		}
		if got := originMain(fx); got != before {
			t.Errorf("origin/main moved to %s", got)
		}
		if p := fx.posts(); len(p) != 0 {
			t.Errorf("a refused land posted %v", p)
		}
	}

	t.Run("landed", func(t *testing.T) {
		fx := newFixture(t)
		before := originMain(fx)
		head := commit(t, fx, "docs/a.md")
		stamp(t, fx, head)
		code, stdout, stderr := land(fx)
		if code != 0 || stdout != "commit: "+head+"\nbase: "+before+"\n" {
			t.Fatalf("exit=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
		if got := originMain(fx); got != head {
			t.Errorf("origin/main = %s, want %s", got, head)
		}
		p := fx.posts()
		if len(p) != 2 || p[0].path != "/api/v1/stages/begin" || p[1].path != "/api/v1/stages/end" {
			t.Fatalf("posts = %v, want one begin and one end", p)
		}
		if p[0].body["stage"] != "flow.landing-routes" || p[0].body["request"] != float64(3) {
			t.Errorf("begin = %v, want flow.landing-routes carrying request 3", p[0].body)
		}
		if p[1].body["outcome"] != "completed" {
			t.Errorf("end = %v, want completed", p[1].body)
		}
	})
	t.Run("dirty", func(t *testing.T) {
		fx := newFixture(t)
		before := originMain(fx)
		stamp(t, fx, commit(t, fx, "docs/a.md"))
		writeTestFile(t, filepath.Join(fx.wt, "stray.txt"), "x\n")
		code, _, stderr := land(fx)
		refused(t, fx, before, code, stderr, "not clean")
	})
	t.Run("not stamped", func(t *testing.T) {
		fx := newFixture(t)
		before := originMain(fx)
		commit(t, fx, "docs/a.md")
		code, _, stderr := land(fx)
		refused(t, fx, before, code, stderr, "flow-lint-green")
	})
	t.Run("rebased", func(t *testing.T) {
		fx := newFixture(t)
		stamp(t, fx, commit(t, fx, "docs/a.md"))
		other := filepath.Join(t.TempDir(), "other")
		fx.git("", "clone", "-q", fx.origin, other)
		writeTestFile(t, filepath.Join(other, "docs/b.md"), "b\n")
		fx.git(other, "add", "-A")
		fx.git(other, "commit", "-q", "-m", "feat: b")
		fx.git(other, "push", "-q", "origin", "HEAD:main")
		before := originMain(fx)
		code, _, stderr := land(fx)
		refused(t, fx, before, code, stderr, "flow-lint-green", "HEAD "+fx.git(fx.wt, "rev-parse", "HEAD"))
	})
	t.Run("tracked scratch", func(t *testing.T) {
		fx := newFixture(t)
		before := originMain(fx)
		path := ".superpowers/sdd/kan-1-x/land.sh"
		writeTestFile(t, filepath.Join(fx.wt, path), "x\n")
		fx.git(fx.wt, "add", "-f", path)
		fx.git(fx.wt, "commit", "-q", "-m", "feat: land.sh")
		stamp(t, fx, fx.git(fx.wt, "rev-parse", "HEAD"))
		code, _, stderr := land(fx)
		refused(t, fx, before, code, stderr, path)
	})
	t.Run("rejected", func(t *testing.T) {
		fx := newFixture(t)
		before := originMain(fx)
		stamp(t, fx, commit(t, fx, "docs/a.md"))
		hook := filepath.Join(fx.origin, "hooks/pre-receive")
		writeTestFile(t, hook, "#!/bin/sh\nexit 1\n")
		if err := os.Chmod(hook, 0o755); err != nil {
			t.Fatal(err)
		}
		code, _, stderr := land(fx)
		if code != 1 || originMain(fx) != before {
			t.Fatalf("exit=%d stderr=%q, origin/main moved=%v", code, stderr, originMain(fx) != before)
		}
		p := fx.posts()
		if len(p) != 2 || p[1].body["outcome"] != "failed" {
			t.Fatalf("posts = %v, want a begin and a failed end", p)
		}
		if m, _ := p[1].body["metrics"].(map[string]any); m == nil || m["reason"] == nil || m["reason"] == "" {
			t.Errorf("end = %v, want a reason", p[1].body)
		}
	})
	t.Run("usage", func(t *testing.T) {
		fx := newFixture(t)
		for name, args := range map[string][]string{
			"no -n":             {"-change", "kan-1-x", "-harness", "h", "-session-token", "fi-k1-20261109T100000Z"},
			"no -session-token": {"-change", "kan-1-x", "-n", "3", "-harness", "h"},
		} {
			if code, _, stderr := land(fx, args...); code != 2 {
				t.Errorf("%s: exit=%d stderr=%q, want 2", name, code, stderr)
			}
		}
		var o, e bytes.Buffer
		if code := run(context.Background(), []string{"interactive"}, strings.NewReader(""), &o, &e); code != 2 {
			t.Errorf("no verb: exit=%d, want 2", code)
		}
	})
}

func writeTestFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
