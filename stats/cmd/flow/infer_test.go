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
)

// changeWorktree is a repository checked out on branch name and carrying
// .superpowers/sdd/<name>/ -- a change's worktree, as inferChange reads it.
func changeWorktree(t *testing.T, name string) string {
	t.Helper()
	repo := gitRepo(t)
	if out, err := exec.Command("git", "-C", repo, "checkout", "-q", "-b", name).CombinedOutput(); err != nil {
		t.Fatalf("checkout -b %s: %v: %s", name, err, out)
	}
	if err := os.MkdirAll(filepath.Join(repo, ".superpowers", "sdd", name), 0o755); err != nil {
		t.Fatal(err)
	}
	return repo
}

func TestInferChange(t *testing.T) {
	t.Parallel()
	wt := changeWorktree(t, "kan-7")
	sub := filepath.Join(wt, ".superpowers", "sdd", "kan-7")
	if got := inferChange(sub); got != "kan-7" {
		t.Errorf("inferChange(worktree subdir) = %q, want kan-7", got)
	}
	// A branch whose checkout carries no change directory names no change:
	// the main checkout on its base branch.
	if got := inferChange(gitRepo(t)); got != "" {
		t.Errorf("inferChange(main checkout) = %q, want empty", got)
	}
	// A /flow change's branch carries the spectre/ prefix.
	sp := gitRepo(t)
	if out, err := exec.Command("git", "-C", sp, "checkout", "-q", "-b", "spectre/kan-8").CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	if err := os.MkdirAll(filepath.Join(sp, "spectre", "changes", "kan-8"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := inferChange(sp); got != "kan-8" {
		t.Errorf("inferChange(spectre branch) = %q, want kan-8", got)
	}
}

// A parallel executor's worktree is on <name>-r<n> and carries the change's
// .superpowers/sdd/<name>/; a branch merely shaped like it, with no such
// directory, names no change.
func TestInferChangeExecutorBranch(t *testing.T) {
	t.Parallel()
	wt := changeWorktree(t, "kan-9")
	if out, err := exec.Command("git", "-C", wt, "checkout", "-q", "-b", "kan-9-r2").CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	if got := inferChange(wt); got != "kan-9" {
		t.Errorf("inferChange(kan-9-r2 worktree) = %q, want kan-9", got)
	}
	bare := gitRepo(t)
	if out, err := exec.Command("git", "-C", bare, "checkout", "-q", "-b", "foo-r2").CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	if got := inferChange(bare); got != "" {
		t.Errorf("inferChange(foo-r2, no change dir) = %q, want empty", got)
	}
}

func TestResolveHarnessUnderClaudeCode(t *testing.T) {
	t.Parallel()
	env := testEnv{"FLOW_HARNESS": "", "CLAUDECODE": "1"}
	if got := resolveHarness("", env.get); got != "claude-code" {
		t.Errorf("resolveHarness under CLAUDECODE=1 = %q, want claude-code", got)
	}
	if got := resolveHarness("codex", env.get); got != "codex" {
		t.Errorf("explicit -harness = %q, want codex", got)
	}
	env["CLAUDECODE"] = ""
	if got := resolveHarness("", env.get); got != defaultHarness {
		t.Errorf("resolveHarness outside a harness = %q, want %q", got, defaultHarness)
	}
}

// sessionStore records every begin/record body it is sent.
type sessionStore struct {
	mu     sync.Mutex
	bodies []map[string]any
}

func (s *sessionStore) server(t *testing.T) *httptest.Server {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		body["path"] = r.URL.Path
		s.bodies = append(s.bodies, body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stageRunId":1,"attempt":1,"id":7,"seq":1,"role":"implementer","model":"sonnet"}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// A begin naming no change or harness, run inside a change's worktree
// under Claude Code, carries the worktree's change and claude-code; every
// explicit flag still wins.
func TestStageBeginInfersChangeAndHarness(t *testing.T) {
	t.Parallel()
	wt := changeWorktree(t, "kan-9")
	env := newTestEnv(t)
	env["FLOW_HARNESS"], env["CLAUDECODE"] = "", "1"
	st := &sessionStore{}
	srv := st.server(t)

	begin := func(extra ...string) map[string]any {
		t.Helper()
		var stdout, stderr bytes.Buffer
		args := append([]string{"stage", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", wt,
			"-command", "/flow", "-stage", "flow.implement", "-session-token", "mf-k9"}, extra...)
		if code := run(context.Background(), args, env.get, strings.NewReader(""), &stdout, &stderr); code != 0 {
			t.Fatalf("exit %d; stderr:\n%s", code, stderr.String())
		}
		return st.bodies[len(st.bodies)-1]
	}

	if got := begin(); got["changeName"] != "kan-9" || got["harness"] != "claude-code" {
		t.Errorf("inferred begin = change %v harness %v, want kan-9, claude-code", got["changeName"], got["harness"])
	}
	if got := begin("-harness", "codex", "other"); got["changeName"] != "other" || got["harness"] != "codex" {
		t.Errorf("explicit begin = change %v harness %v, want other, codex", got["changeName"], got["harness"])
	}
}

// Outside a change's worktree, with no change named, begin is a usage error.
func TestStageBeginWithoutChangeOutsideWorktreeIsRefused(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"stage", "begin", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", gitRepo(t),
		"-command", "/flow", "-stage", "flow.implement"}, env.get, strings.NewReader(""), &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit %d, want 2; stderr:\n%s", code, stderr.String())
	}
}

// A dispatch end naming no -change, inside the change's worktree, closes the
// row of the worktree's change.
func TestRecordDispatchEndInfersChange(t *testing.T) {
	t.Parallel()
	wt := changeWorktree(t, "kan-10")
	env := newTestEnv(t)
	st := &sessionStore{}
	srv := st.server(t)
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"record", "dispatch", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", wt,
		"-key", "impl-1", "-session-token", "mf-kan10-tok", "-outcome", "completed"}, env.get, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d; stderr:\n%s", code, stderr.String())
	}
	got := st.bodies[len(st.bodies)-1]
	if !strings.Contains(got["path"].(string), "/kan-10/") || got["sessionToken"] != "mf-kan10-tok" {
		t.Errorf("dispatch end = path %v token %v, want the kan-10 route and mf-kan10-tok", got["path"], got["sessionToken"])
	}
}

// The session token is the calling agent's: with no -session-token, a
// dispatch end under Claude Code carries cc-<session>-<agent>, a subagent's
// call (-caller-agent, which the block-unsafe-commands hook inserts) its own,
// the main thread's "main" -- so an executor's marks never supersede its
// coordinator's. With no harness session it is still required.
func TestRecordDispatchEndInfersCallerToken(t *testing.T) {
	t.Parallel()
	wt := changeWorktree(t, "kan-11")
	st := &sessionStore{}
	srv := st.server(t)
	end := func(env testEnv, lead ...string) (int, string) {
		t.Helper()
		var stdout, stderr bytes.Buffer
		args := append(lead, "record", "dispatch", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", wt,
			"-key", "impl-1", "-outcome", "completed")
		code := run(context.Background(), args, env.get, strings.NewReader(""), &stdout, &stderr)
		return code, stderr.String()
	}
	env := newTestEnv(t)
	env["CLAUDE_CODE_SESSION_ID"] = "s-11"
	if code, errs := end(env); code != 0 || st.bodies[len(st.bodies)-1]["sessionToken"] != "cc-s-11-main" {
		t.Fatalf("main thread: exit %d token %v, want cc-s-11-main; stderr:\n%s", code, st.bodies[len(st.bodies)-1]["sessionToken"], errs)
	}
	if code, errs := end(env, "-caller-agent", "a7"); code != 0 || st.bodies[len(st.bodies)-1]["sessionToken"] != "cc-s-11-a7" {
		t.Fatalf("subagent: exit %d token %v, want cc-s-11-a7; stderr:\n%s", code, st.bodies[len(st.bodies)-1]["sessionToken"], errs)
	}
	if code, errs := end(newTestEnv(t)); code != 2 || !strings.Contains(errs, "-session-token") {
		t.Fatalf("no session: exit %d, want 2 naming -session-token; stderr:\n%s", code, errs)
	}
}

// A stage begin with no -session-token under Claude Code carries the calling
// agent's token, not a per-mark one, so its next begin supersedes its own run.
func TestStageBeginInfersCallerToken(t *testing.T) {
	t.Parallel()
	wt := changeWorktree(t, "kan-13")
	env := newTestEnv(t)
	env["CLAUDE_CODE_SESSION_ID"] = "s-13"
	st := &sessionStore{}
	srv := st.server(t)
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), []string{"-caller-agent", "b2", "stage", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", wt,
		"-command", "/flow", "-stage", "flow.implement"}, env.get, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d; stderr:\n%s", code, stderr.String())
	}
	if got := st.bodies[len(st.bodies)-1]["sessionToken"]; got != "cc-s-13-b2" {
		t.Errorf("token = %v, want cc-s-13-b2", got)
	}
}

// interactive land takes its change from the worktree: past the usage check
// it refuses the dirty tree (exit 1), never the missing -change (exit 2).
func TestInteractiveLandInfersChange(t *testing.T) {
	t.Parallel()
	wt := changeWorktree(t, "kan-12")
	if err := os.WriteFile(filepath.Join(wt, "dirty"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	env := newTestEnv(t)
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"interactive", "land", "-addr", deadPortAddr(t), "-C", wt, "-n", "1"},
		env.get, strings.NewReader(""), &stdout, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "not clean") {
		t.Fatalf("exit %d, want 1 refusing the dirty tree; stderr:\n%s", code, stderr.String())
	}
	stderr.Reset()
	code = run(context.Background(), []string{"interactive", "land", "-addr", deadPortAddr(t), "-C", gitRepo(t), "-n", "1"},
		env.get, strings.NewReader(""), &stdout, &stderr)
	if code != 2 {
		t.Fatalf("outside a change worktree: exit %d, want 2; stderr:\n%s", code, stderr.String())
	}
}

// A record naming no -change outside a change's worktree is a usage error,
// and never reaches the store.
func TestRecordWithoutChangeOutsideWorktreeIsRefused(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	st := &sessionStore{}
	srv := st.server(t)
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"record", "request", "begin", "-addr", srv.URL, "-timeout", "300ms", "-C", gitRepo(t),
		"-n", "1", "-title", "t"}, env.get, strings.NewReader(""), &stdout, &stderr)
	if code != 2 || len(st.bodies) != 0 {
		t.Fatalf("exit %d, %d writes, want 2 and none; stderr:\n%s", code, len(st.bodies), stderr.String())
	}
}
