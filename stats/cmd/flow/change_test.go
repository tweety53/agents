package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/store"
	"github.com/tweety53/agents/stats/internal/storetest"
)

// writeChangeFile writes rel under root, creating its directories.
func writeChangeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestChangeSaveUpsertsBothFiles drives `flow change save` against a real
// flowd and store: both files land, and `flow change file` reads each back
// byte for byte.
// smoke: core — the CLI through a real api.New to a real store
func TestChangeSaveUpsertsBothFiles(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	st := storetest.NewStore(t)
	projectKey, mainCheckout, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.PutChange(context.Background(), store.Change{
		ProjectKey: projectKey, MainCheckoutPath: mainCheckout, Name: "demo",
		State: store.StateInProgress, UpdatedAt: time.Now(), UpdatedBy: "test",
	}); err != nil {
		t.Fatalf("seed change: %v", err)
	}
	srv, err := api.New(config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}, st, st, st, st, st, nil, api.WithChangeFiles(st))
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	tasks := "# demo\n\n- [ ] 1. one task\n"
	bundle := "# Self-review context bundle for demo\n"
	writeChangeFile(t, repo, "spectre/changes/demo/tasks.md", tasks)
	writeChangeFile(t, repo, ".superpowers/sdd/demo/self-review-context.md", bundle)

	var stdout, stderr strings.Builder
	if code := run(context.Background(), []string{"change", "save", "-addr", ts.URL, "-C", repo, "demo"},
		strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("save exit = %d, stderr: %s", code, stderr.String())
	}
	for path, want := range map[string]string{"tasks.md": tasks, "self-review-context.md": bundle} {
		stdout.Reset()
		stderr.Reset()
		if code := run(context.Background(), []string{"change", "file", "-addr", ts.URL, "-C", repo, "demo", path},
			strings.NewReader(""), &stdout, &stderr); code != 0 {
			t.Fatalf("file %s exit = %d, stderr: %s", path, code, stderr.String())
		}
		if stdout.String() != want {
			t.Errorf("file %s = %q, want %q", path, stdout.String(), want)
		}
	}
}

// TestChangeSaveRefusesAMissingFileBeforeWriting pins the save's
// all-or-nothing refusal: with only tasks.md on disk it exits 1 naming the
// missing bundle, and flowd sees no PUT at all.
func TestChangeSaveRefusesAMissingFileBeforeWriting(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	puts := 0
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			puts++
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	writeChangeFile(t, repo, "spectre/changes/demo/tasks.md", "# demo\n")

	var stdout, stderr strings.Builder
	code := run(context.Background(), []string{"change", "save", "-addr", srv.URL, "-C", repo, "demo"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit = %d, want 1; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), ".superpowers/sdd/demo/self-review-context.md") {
		t.Errorf("stderr = %q, want the missing file named", stderr.String())
	}
	if puts != 0 {
		t.Errorf("flowd saw %d PUT(s), want none before every file is present", puts)
	}
}

// TestChangeSaveReadsEitherPlanOrNone pins the plan's two places: a
// /flow-fast plan under .superpowers/sdd/<name>/ is saved as tasks.md, and a
// change with no plan at all saves its bundle alone.
func TestChangeSaveReadsEitherPlanOrNone(t *testing.T) {
	for _, tc := range []struct {
		label, plan string
		want        []string
	}{
		{"flow-fast plan", ".superpowers/sdd/demo/tasks.md", []string{"tasks.md=fast plan", "self-review-context.md=bundle"}},
		{"no plan", "", []string{"self-review-context.md=bundle"}},
		{"both plans", "both", []string{"tasks.md=spectre plan", "self-review-context.md=bundle"}},
	} {
		t.Run(tc.label, func(t *testing.T) {
			repo := gitRepo(t)
			isolatedStateRoot(t)
			var puts []string
			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPut {
					b, _ := io.ReadAll(r.Body)
					var body struct{ Body string }
					_ = json.Unmarshal(b, &body)
					puts = append(puts, filepath.Base(r.URL.Path)+"="+body.Body)
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer srv.Close()
			switch tc.plan {
			case "":
			case "both":
				writeChangeFile(t, repo, "spectre/changes/demo/tasks.md", "spectre plan")
				writeChangeFile(t, repo, ".superpowers/sdd/demo/tasks.md", "fast plan")
			default:
				writeChangeFile(t, repo, tc.plan, "fast plan")
			}
			writeChangeFile(t, repo, ".superpowers/sdd/demo/self-review-context.md", "bundle")
			var stdout, stderr strings.Builder
			if code := run(context.Background(), []string{"change", "save", "-addr", srv.URL, "-C", repo, "demo"},
				strings.NewReader(""), &stdout, &stderr); code != 0 {
				t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
			}
			if strings.Join(puts, "|") != strings.Join(tc.want, "|") {
				t.Errorf("PUTs = %v, want %v", puts, tc.want)
			}
		})
	}
}

// TestChangeFilePrintsBody pins `flow change file`: the served body on
// stdout verbatim, exit 1 on a 404, and exit 2 when flowd is unreachable.
func TestChangeFilePrintsBody(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	var gotPath string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if strings.HasSuffix(r.URL.Path, "/missing.md") {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		_, _ = w.Write([]byte("# plan\n"))
	}))
	defer srv.Close()

	var stdout, stderr strings.Builder
	if code := run(context.Background(), []string{"change", "file", "-addr", srv.URL, "-C", repo, "demo", "tasks.md"},
		strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if stdout.String() != "# plan\n" {
		t.Errorf("stdout = %q", stdout.String())
	}
	if !strings.HasSuffix(gotPath, "/demo/files/tasks.md") {
		t.Errorf("request path = %s", gotPath)
	}

	stdout.Reset()
	if code := run(context.Background(), []string{"change", "file", "-addr", srv.URL, "-C", repo, "demo", "missing.md"},
		strings.NewReader(""), &stdout, &stderr); code != 1 || stdout.Len() != 0 {
		t.Errorf("missing file: exit = %d stdout = %q, want 1 and nothing", code, stdout.String())
	}
	if code := run(context.Background(), []string{"change", "file", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo, "demo", "tasks.md"},
		strings.NewReader(""), &stdout, &stderr); code != 2 {
		t.Errorf("unreachable store: exit = %d, want 2", code)
	}
}
