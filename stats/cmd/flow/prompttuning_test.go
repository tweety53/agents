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
	"testing"

	"github.com/tweety53/agents/stats/internal/prompttuning"
)

func TestPromptTuningRenderWritesIndex(t *testing.T) {
	dir := t.TempDir()
	c := prompttuning.Case{ID: "x-1", Type: "implementer", Scope: "generic", Target: "skills/flow/implement.md"}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"case.json": string(data), "finding.md": "f", "bad.diff": "b", "good.diff": "g", "task.md": "t",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), []string{"prompt-tuning", "render", dir}, nil, &stdout, &stderr); code != 0 {
		t.Fatalf("code = %d, stderr = %s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want silence", stdout.String())
	}
	got, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	lc, lf, err := prompttuning.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	want, err := prompttuning.Render(lc, lf)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Error("index.html differs from Render's output")
	}

	stderr.Reset()
	if code := run(context.Background(), []string{"prompt-tuning", "render"}, nil, &stdout, &stderr); code != 2 {
		t.Errorf("missing argument: code = %d, want 2", code)
	}
	if code := run(context.Background(), []string{"prompt-tuning", "rendr", dir}, nil, &stdout, &stderr); code != 2 {
		t.Errorf("unknown subcommand: code = %d, want 2", code)
	}
	stderr.Reset()
	if code := run(context.Background(), []string{"prompt-tuning", "render", filepath.Join(dir, "absent")}, nil, &stdout, &stderr); code != 1 {
		t.Errorf("absent folder: code = %d, want 1", code)
	}
}

// promptChangeSink is a genuine-daemon httptest server collecting every
// POST /api/v1/prompt-changes body.
func promptChangeSink(t *testing.T) (*httptest.Server, *[]map[string]any) {
	t.Helper()
	var rows []map[string]any
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/prompt-changes" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var m map[string]any
		body, _ := readAll(r)
		_ = json.Unmarshal(body, &m)
		rows = append(rows, m)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	return srv, &rows
}

func writeCase(t *testing.T, c prompttuning.Case) string {
	t.Helper()
	dir := t.TempDir()
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"case.json": string(data), "finding.md": "f", "bad.diff": "b", "good.diff": "g", "task.md": "t"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// TestPromptTuningRecord pins `flow prompt-tuning record`: one tuning row
// per applied case, its area from prompt-area-map or, off the map, from
// the case's type.
func TestPromptTuningRecord(t *testing.T) {
	srv, rows := promptChangeSink(t)
	reviewer := writeCase(t, prompttuning.Case{ID: "kan-1-x-1", Type: "reviewer", Scope: "generic",
		Target: "skills/flow/test-sanity-reviewer-prompt.md", DefectClass: "unasserted branch"})
	impl := writeCase(t, prompttuning.Case{ID: "kan-2-y-1", Type: "implementer", Scope: "generic",
		Target: "rules/x.mdc", DefectClass: "derived state stored"})

	var stdout, stderr bytes.Buffer
	for _, dir := range []string{reviewer, impl} {
		if code := run(context.Background(), []string{"prompt-tuning", "record", dir, "-sha", "abc123",
			"-at", "2026-10-27T10:00:00Z", "-addr", srv.URL}, nil, &stdout, &stderr); code != 0 {
			t.Fatalf("record %s: code %d, stderr %s", dir, code, stderr.String())
		}
	}
	if len(*rows) != 2 {
		t.Fatalf("posted %d rows, want 2", len(*rows))
	}
	r := (*rows)[0]
	want := map[string]any{"sha": "abc123:skills/flow/test-sanity-reviewer-prompt.md", "kind": "tuning", "area": "review",
		"slot": "test-sanity", "caseId": "kan-1-x-1", "summary": "unasserted branch", "at": "2026-10-27T10:00:00Z",
		"target": "skills/flow/test-sanity-reviewer-prompt.md"}
	for k, v := range want {
		if r[k] != v {
			t.Errorf("reviewer row %s = %v, want %v", k, r[k], v)
		}
	}
	if r := (*rows)[1]; r["area"] != "impl" || r["slot"] != nil || r["sha"] != "abc123:rules/x.mdc" {
		t.Errorf("implementer row = %v, want area impl, no slot", r)
	}

	if code := run(context.Background(), []string{"prompt-tuning", "record", reviewer, "-addr", srv.URL}, nil, &stdout, &stderr); code != 2 {
		t.Errorf("missing -sha: code %d, want 2", code)
	}

	// With no -at, the time is the applying commit's, read in the case
	// dir's repository.
	repo := gitRepo(t)
	caseDir := filepath.Join(repo, "docs", "prompt-tuning", "kan-1-x-1")
	if err := os.MkdirAll(filepath.Dir(caseDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(reviewer, caseDir); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "-C", repo, "-c", "user.name=t", "-c", "user.email=t@e", "commit", "-q", "--allow-empty", "-m", "apply")
	cmd.Env = append(os.Environ(), "GIT_COMMITTER_DATE=2026-10-20T08:30:00Z", "GIT_AUTHOR_DATE=2026-10-19T07:00:00Z")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("commit: %v\n%s", err, out)
	}
	head, err := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	if code := run(context.Background(), []string{"prompt-tuning", "record", caseDir, "-sha", strings.TrimSpace(string(head)), "-addr", srv.URL}, nil, &stdout, &stderr); code != 0 {
		t.Fatalf("record without -at: code %d, stderr %s", code, stderr.String())
	}
	if got := (*rows)[len(*rows)-1]["at"]; got != "2026-10-20T08:30:00Z" {
		t.Errorf("record without -at: at = %v, want the commit time 2026-10-20T08:30:00Z", got)
	}
}

// TestPromptTuningImport pins `flow prompt-tuning import`: one edit row per
// (commit, mapped file), keyed <sha>:<target>, with the commit time and
// subject; an unmapped file yields none and -since drops older commits.
func TestPromptTuningImport(t *testing.T) {
	repo := gitRepo(t)
	git := func(env []string, args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = append(append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@e", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@e"), env...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	commit := func(date, subject string, files ...string) string {
		for _, f := range files {
			p := filepath.Join(repo, f)
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte(subject), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		git(nil, append([]string{"add"}, files...)...)
		env := []string{"GIT_AUTHOR_DATE=" + date, "GIT_COMMITTER_DATE=" + date}
		git(env, "commit", "-q", "-m", subject)
		return git(nil, "rev-parse", "HEAD")
	}
	old := commit("2026-07-01T10:00:00Z", "old primary edit", "skills/flow/primary-reviewer-prompt.md")
	both := commit("2026-09-01T10:00:00Z", "primary and verifier", "skills/flow/primary-reviewer-prompt.md", "skills/flow/visual-verify-verifier.md")
	commit("2026-09-02T10:00:00Z", "unmapped", "README.md")
	added := commit("2026-09-03T10:00:00Z", "a slot added later", "skills/flow/new-slot-reviewer-prompt.md")
	commit("2026-09-04T10:00:00Z", "nested, not a slot", "skills/flow/sub/x-reviewer-prompt.md")

	srv, rows := promptChangeSink(t)
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), []string{"prompt-tuning", "import", "-C", repo, "-addr", srv.URL}, nil, &stdout, &stderr); code != 0 {
		t.Fatalf("import: code %d, stderr %s", code, stderr.String())
	}
	if got := strings.TrimSpace(stdout.String()); got != "imported: 4" {
		t.Errorf("stdout = %q, want imported: 4", got)
	}
	keys := map[string]map[string]any{}
	for _, r := range *rows {
		keys[r["sha"].(string)] = r
	}
	if r := keys[added+":skills/flow/new-slot-reviewer-prompt.md"]; r == nil || r["slot"] != "new-slot" || r["area"] != "review" {
		t.Errorf("a reviewer prompt off the fixed map = %v, want area review, slot new-slot", r)
	}
	for _, k := range []string{old + ":skills/flow/primary-reviewer-prompt.md", both + ":skills/flow/primary-reviewer-prompt.md", both + ":skills/flow/visual-verify-verifier.md"} {
		if _, ok := keys[k]; !ok {
			t.Errorf("no row %s in %v", k, keys)
		}
	}
	if r := keys[both+":skills/flow/visual-verify-verifier.md"]; r != nil && (r["kind"] != "edit" || r["area"] != "vv" || r["summary"] != "primary and verifier" || r["at"] != "2026-09-01T10:00:00Z") {
		t.Errorf("verifier row = %v", r)
	}

	*rows = nil
	stdout.Reset()
	if code := run(context.Background(), []string{"prompt-tuning", "import", "-C", repo, "-since", "2026-08-15", "-addr", srv.URL}, nil, &stdout, &stderr); code != 0 {
		t.Fatalf("import -since: code %d, stderr %s", code, stderr.String())
	}
	if len(*rows) != 3 || strings.TrimSpace(stdout.String()) != "imported: 3" {
		t.Errorf("-since 2026-08-15 posted %d rows, stdout %q; want 3", len(*rows), stdout.String())
	}
}
