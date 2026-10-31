package repotest

// TestFlowSessionActivityHook drives hooks/flow-session-activity.py against a
// stub store (an httptest server logging every request's method, path and
// body), pointed at through FLOW_ADDR, for each hook event it maps, and
// asserts the PUT it sends -- or that it sends none and stays silent.

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// activityStore is a stub flow store; requests holds "METHOD path body".
type activityStore struct {
	addr     string
	mu       sync.Mutex
	requests []string
}

func newActivityStore(t *testing.T) *activityStore {
	t.Helper()
	s := &activityStore{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		s.mu.Lock()
		s.requests = append(s.requests, r.Method+" "+r.URL.EscapedPath()+" "+string(body))
		s.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	s.addr = srv.URL
	return s
}

func (s *activityStore) all() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.requests...)
}

func TestFlowSessionActivityHook(t *testing.T) {
	t.Parallel()

	run := func(t *testing.T, addr, stdin string) (string, int) {
		t.Helper()
		out, errOut, rc := runHook(t, "flow-session-activity.py", []byte(stdin), "FLOW_ADDR="+addr)
		return out + errOut, rc
	}

	const path = "PUT /api/v1/sessions/s-1/activity "
	long := strings.Repeat("é", 300) // 600 bytes
	cases := []struct {
		name, stdin, want string
	}{
		{"ask user question", `{"session_id":"s-1","hook_event_name":"PreToolUse","tool_name":"AskUserQuestion","tool_input":{"questions":[{"question":"Proceed?"},{"question":"Second?"}]}}`,
			path + `{"state": "question", "detail": "Proceed?"}`},
		{"permission notification", `{"session_id":"s-1","hook_event_name":"Notification","message":"Claude needs your permission to use Bash"}`,
			path + `{"state": "permission", "detail": "Claude needs your permission to use Bash"}`},
		{"other notification", `{"session_id":"s-1","hook_event_name":"Notification","message":"Claude is waiting for your input"}`, ""},
		{"prompt submit", `{"session_id":"s-1","hook_event_name":"UserPromptSubmit","prompt":"go"}`,
			path + `{"state": "working", "detail": ""}`},
		{"post tool use", `{"session_id":"s-1","hook_event_name":"PostToolUse","tool_name":"Bash"}`,
			path + `{"state": "working", "detail": "Bash"}`},
		{"post tool use names a bash call by its description", `{"session_id":"s-1","hook_event_name":"PostToolUse","tool_name":"Bash","tool_input":{"command":"make lint","description":"Run the linters"}}`,
			path + `{"state": "working", "detail": "Bash: Run the linters"}`},
		{"post tool use names a bash call by its first command line", `{"session_id":"s-1","hook_event_name":"PostToolUse","tool_name":"Bash","tool_input":{"command":"make test\nmore"}}`,
			path + `{"state": "working", "detail": "Bash: make test"}`},
		{"post tool use names an edited file", `{"session_id":"s-1","hook_event_name":"PostToolUse","tool_name":"Edit","tool_input":{"file_path":"/a/b/Board.tsx"}}`,
			path + `{"state": "working", "detail": "Edit Board.tsx"}`},
		{"post tool use names a dispatched agent", `{"session_id":"s-1","hook_event_name":"PostToolUse","tool_name":"Agent","tool_input":{"description":"Review the diff"}}`,
			path + `{"state": "working", "detail": "Agent: Review the diff"}`},
		{"a subagent's post tool use sends nothing", `{"session_id":"s-1","hook_event_name":"PostToolUse","tool_name":"Bash","agent_id":"a1"}`, ""},
		{"a subagent's ask user question sends nothing", `{"session_id":"s-1","hook_event_name":"PreToolUse","tool_name":"AskUserQuestion","agent_id":"a1","tool_input":{"questions":[{"question":"Proceed?"}]}}`, ""},
		{"a subagent's permission notification still sends", `{"session_id":"s-1","hook_event_name":"Notification","agent_id":"a1","message":"Claude needs your permission to use Bash"}`,
			path + `{"state": "permission", "detail": "Claude needs your permission to use Bash"}`},
		{"stop", `{"session_id":"s-1","hook_event_name":"Stop"}`,
			path + `{"state": "idle", "detail": ""}`},
		{"session end", `{"session_id":"s-1","hook_event_name":"SessionEnd","reason":"prompt_input_exit"}`,
			path + `{"state": "ended", "detail": ""}`},
		{"other pre tool use reports working before the call runs", `{"session_id":"s-1","hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"sleep 60","description":"Wait for the verifier"}}`,
			path + `{"state": "working", "detail": "Bash: Wait for the verifier"}`},
		{"no session id", `{"hook_event_name":"Stop"}`, ""},
		{"malformed json", `{not json`, ""},
		{"detail truncated to 500 bytes", `{"session_id":"s-1","hook_event_name":"PreToolUse","tool_name":"AskUserQuestion","tool_input":{"questions":[{"question":"` + long + `"}]}}`,
			path + `{"state": "question", "detail": "` + strings.Repeat(`\u00e9`, 250) + `"}`},
		{"ascii detail cut at exactly 500 bytes", `{"session_id":"s-1","hook_event_name":"PreToolUse","tool_name":"AskUserQuestion","tool_input":{"questions":[{"question":"` + strings.Repeat("x", 600) + `"}]}}`,
			path + `{"state": "question", "detail": "` + strings.Repeat("x", 500) + `"}`},
		{"NUL stripped from detail", `{"session_id":"s-1","hook_event_name":"PreToolUse","tool_name":"AskUserQuestion","tool_input":{"questions":[{"question":"a\u0000b"}]}}`,
			path + `{"state": "question", "detail": "ab"}`},
		{"session id escaped into one path segment", `{"session_id":"a/b c","hook_event_name":"Stop"}`,
			`PUT /api/v1/sessions/a%2Fb%20c/activity {"state": "idle", "detail": ""}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := newActivityStore(t)
			out, rc := run(t, s.addr, tc.stdin)
			if rc != 0 || out != "" {
				t.Fatalf("rc=%d out=%q, want 0 and no output", rc, out)
			}
			got := s.all()
			if tc.want == "" {
				if len(got) != 0 {
					t.Errorf("requests = %q, want none", got)
				}
				return
			}
			if len(got) != 1 || got[0] != tc.want {
				t.Errorf("requests = %q, want [%q]", got, tc.want)
			}
		})
	}

	t.Run("unreachable store", func(t *testing.T) {
		t.Parallel()
		out, rc := run(t, "http://127.0.0.1:1", `{"session_id":"s-1","hook_event_name":"Stop"}`)
		if rc != 0 || out != "" {
			t.Errorf("rc=%d out=%q, want 0 and no output", rc, out)
		}
	})
}
