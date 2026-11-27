package guard

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
)

// TestHookFlowSkillLoad pins flow-skill-load: a Read of a skills/, rules/ or docs/ file of an agents
// checkout — reached directly or through a symlink — and a Skill call whose installed SKILL.md
// resolves into one post one load, repo-relative; any other path, a plugin skill and another tool
// post none; and the hook prints nothing and exits 0, flowd up or down.
func TestHookFlowSkillLoad(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	for _, d := range []string{"stats/internal/guard", "skills/flow", "rules"} {
		if err := os.MkdirAll(filepath.Join(repo, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{"skills/flow/fix.md", "skills/flow/SKILL.md", "rules/agent-baseline.md", "README.md"} {
		if err := os.WriteFile(filepath.Join(repo, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	other := t.TempDir()
	if err := os.MkdirAll(filepath.Join(other, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(other, "skills", "a.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".claude", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(repo, "skills", "flow"), filepath.Join(home, ".claude", "skills", "flow")); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var got []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in struct{ Path, SessionID string }
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &in)
		mu.Lock()
		got = append(got, r.Method+" "+r.URL.Path+" "+in.Path+" "+in.SessionID)
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	payload := func(tool string, in map[string]any) string {
		b, _ := json.Marshal(map[string]any{"hook_event_name": "PostToolUse", "session_id": "s-1", "tool_name": tool, "tool_input": in})
		return string(b)
	}
	cases := []struct {
		name, stdin string
		want        []string
	}{
		{"skill file", payload("Read", map[string]any{"file_path": filepath.Join(repo, "skills/flow/fix.md")}),
			[]string{"POST /api/v1/skill-loads skills/flow/fix.md s-1"}},
		{"rules file", payload("Read", map[string]any{"file_path": filepath.Join(repo, "rules/agent-baseline.md")}),
			[]string{"POST /api/v1/skill-loads rules/agent-baseline.md s-1"}},
		{"through the installed symlink", payload("Read", map[string]any{"file_path": filepath.Join(home, ".claude/skills/flow/fix.md")}),
			[]string{"POST /api/v1/skill-loads skills/flow/fix.md s-1"}},
		{"Skill tool", payload("Skill", map[string]any{"skill": "flow"}),
			[]string{"POST /api/v1/skill-loads skills/flow/SKILL.md s-1"}},
		{"plugin skill", payload("Skill", map[string]any{"skill": "superpowers:brainstorming"}), nil},
		{"repo file outside skills, rules and docs", payload("Read", map[string]any{"file_path": filepath.Join(repo, "README.md")}), nil},
		{"skills dir of another repository", payload("Read", map[string]any{"file_path": filepath.Join(other, "skills/a.md")}), nil},
		{"missing file", payload("Read", map[string]any{"file_path": filepath.Join(repo, "skills/flow/none.md")}), nil},
		{"another tool", payload("Edit", map[string]any{"file_path": filepath.Join(repo, "skills/flow/fix.md")}), nil},
		{"no JSON", "{not json", nil},
	}
	for _, c := range cases {
		mu.Lock()
		got = nil
		mu.Unlock()
		out, rc := runHookInProcess(t, "flow-skill-load", c.stdin, "FLOW_ADDR="+srv.URL, "HOME="+home)
		mu.Lock()
		if rc != 0 || out != "" || !slices.Equal(got, c.want) {
			t.Errorf("%s: rc=%d out=%q posted %q, want %q", c.name, rc, out, got, c.want)
		}
		mu.Unlock()
	}

	// flowd down: still silent, still exit 0.
	if out, rc := runHookInProcess(t, "flow-skill-load", cases[0].stdin, "HOME="+home); rc != 0 || out != "" {
		t.Errorf("flowd down: rc=%d out=%q", rc, out)
	}
}
