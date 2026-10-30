package repotest

// TestFlowActiveChangeHook ports the retired shell harness, the
// harness for hooks/flow-active-change.py. Each subtest starts its own stub
// store (an httptest server answering GET /api/v1/stage-runs with a canned body
// and status, logging every request's path), points the hook at it through
// FLOW_ADDR, and pipes fixture stdin JSON to the hook, asserting its output
// (stdout and stderr together) and exit code — never one alone. Subtest names
// are the harness's own case labels.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// stageStore is a stub flow store; requests holds every request's path and query.
type stageStore struct {
	addr     string
	mu       sync.Mutex
	requests []string
}

func newStageStore(t *testing.T, status int, body string) *stageStore {
	t.Helper()
	s := &stageStore{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.requests = append(s.requests, r.URL.RequestURI())
		s.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	s.addr = srv.URL
	return s
}

func (s *stageStore) last() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.requests) == 0 {
		return ""
	}
	return s.requests[len(s.requests)-1]
}

// stageBody is the store's stageRuns answer.
func stageBody(total int, rows ...string) string {
	return fmt.Sprintf(`{"total":%d,"stageRuns":[%s]}`, total, strings.Join(rows, ","))
}

// stageRow is one stage-run row with (or, for "", without) a changeName.
func stageRow(change string) string {
	if change == "" {
		return `{"stageRunId":7,"command":"/flow","stage":"plan.session","startedAt":"2026-09-17T09:00:00Z"}`
	}
	return `{"stageRunId":7,"command":"/flow","stage":"flow.kickoff","startedAt":"2026-09-17T09:00:00Z","changeName":"` + change + `"}`
}

func TestFlowActiveChangeHook(t *testing.T) {
	t.Parallel()
	const deadAddr = "http://127.0.0.1:1"

	// run pipes stdin to the hook against addr; it returns stdout+stderr and the exit code.
	run := func(t *testing.T, addr, stdin string) (string, int) {
		t.Helper()
		out, errOut, rc := runHook(t, "flow-active-change.py", []byte(stdin), "FLOW_ADDR="+addr)
		return strings.TrimRight(out+errOut, "\n"), rc
	}
	// rollout is a payload naming a rollout transcript under a temp dir.
	rollout := func(t *testing.T, extra string) string {
		return fmt.Sprintf(`{"prompt":%s"transcript_path":"%s/model-io-sess_abc.jsonl"}`, extra, t.TempDir())
	}
	case1 := func(t *testing.T) (string, int, *stageStore) {
		s := newStageStore(t, 200, stageBody(1, stageRow("fixture-one")))
		out, rc := run(t, s.addr, rollout(t, `"the button is misaligned",`))
		return out, rc, s
	}
	case2 := func(t *testing.T) (string, int) {
		s := newStageStore(t, 200, stageBody(2, stageRow("fixture-two"), stageRow("fixture-one")))
		return run(t, s.addr, rollout(t, `"still broken","session_id":"abc",`))
	}
	case4 := func(t *testing.T) (string, int, *stageStore) {
		s := newStageStore(t, 200, stageBody(1, stageRow("fixture-one")))
		out, rc := run(t, s.addr, `{"prompt":"/flow x","session_id":"abc"}`)
		return out, rc, s
	}
	case6 := func(t *testing.T) (string, int, *stageStore) {
		s := newStageStore(t, 200, stageBody(1, stageRow("fixture-six")))
		out, rc := run(t, s.addr, `{"prompt":"broken again","session_id":"sess-six"}`)
		return out, rc, s
	}
	exit0 := func(t *testing.T, out string, rc int) {
		t.Helper()
		if rc != 0 {
			t.Errorf("rc=%d out=%s", rc, out)
		}
	}
	silent := func(t *testing.T, out string, rc int) {
		t.Helper()
		if rc != 0 || out != "" {
			t.Errorf("rc=%d out=%s", rc, out)
		}
	}
	has := func(t *testing.T, what, s, needle string) {
		t.Helper()
		if !strings.Contains(s, needle) {
			t.Errorf("expected %q in %s=%s", needle, what, s)
		}
	}

	cases := []struct {
		name string
		fn   func(t *testing.T)
	}{
		// Case 1: the store holds the session's /flow mark — the context line names its
		// change, and the query filters by the session id derived from the rollout
		// transcript's filename, newest mark of any command first, one row.
		{"case 1: exit 0", func(t *testing.T) { out, rc, _ := case1(t); exit0(t, out, rc) }},
		{"case 1: context line names fixture-one", func(t *testing.T) {
			out, _, _ := case1(t)
			has(t, "out", out, "marked change fixture-one")
		}},
		{"case 1: session_id derived from the rollout filename", func(t *testing.T) {
			_, _, s := case1(t)
			has(t, "query", s.last(), "session_id=abc")
		}},
		{"case 1: no command filter in the query", func(t *testing.T) {
			_, _, s := case1(t)
			if q := s.last(); strings.Contains(q, "command=") {
				t.Errorf("expected no command filter in query=%s", q)
			}
		}},
		{"case 1: newest-mark-first, one row", func(t *testing.T) {
			_, _, s := case1(t)
			q := s.last()
			if i := strings.Index(q, "sort=-started_at"); i < 0 || !strings.Contains(q[i:], "limit=1") {
				t.Errorf("expected sort=-started_at&limit=1 in query=%s", q)
			}
		}},
		// Case 2: two /flow marks for different changes — the store answers newest
		// first and the hook reads stageRuns[0], so the last mark wins.
		{"case 2: exit 0", func(t *testing.T) { out, rc := case2(t); exit0(t, out, rc) }},
		{"case 2: last mark (fixture-two) wins", func(t *testing.T) {
			out, _ := case2(t)
			has(t, "out", out, "marked change fixture-two")
		}},
		// Case 3: no mark of the session in the store — no output, exit 0.
		{"case 3: no mark -> exit 0, empty output", func(t *testing.T) {
			s := newStageStore(t, 200, stageBody(0))
			out, rc := run(t, s.addr, `{"prompt":"any problem here","session_id":"abc"}`)
			silent(t, out, rc)
		}},
		// Case 4: a prompt starting with / — no output, exit 0, and no store call.
		{"case 4: slash prompt -> exit 0, empty output", func(t *testing.T) { out, rc, _ := case4(t); silent(t, out, rc) }},
		{"case 4: store not consulted", func(t *testing.T) {
			if _, _, s := case4(t); s.last() != "" {
				t.Errorf("expected no store call, log=%v", s.requests)
			}
		}},
		// Case 5: malformed stdin — no output, exit 0.
		{"case 5: malformed stdin -> exit 0, empty output", func(t *testing.T) {
			s := newStageStore(t, 200, stageBody(1, stageRow("fixture-one")))
			out, rc := run(t, s.addr, "not json")
			silent(t, out, rc)
		}},
		// Case 6: no transcript_path — the session_id from the hook payload is the
		// store query's session filter.
		{"case 6: exit 0", func(t *testing.T) { out, rc, _ := case6(t); exit0(t, out, rc) }},
		{"case 6: payload session_id queried", func(t *testing.T) {
			out, _, _ := case6(t)
			has(t, "out", out, "marked change fixture-six")
		}},
		{"case 6: query filtered by the payload session id", func(t *testing.T) {
			_, _, s := case6(t)
			has(t, "query", s.last(), "session_id=sess-six")
		}},
		// Case 7: the store is unreachable — fail open, no output, exit 0.
		{"case 7: unreachable store -> exit 0, empty output", func(t *testing.T) {
			out, rc := run(t, deadAddr, `{"prompt":"broken","session_id":"abc"}`)
			silent(t, out, rc)
		}},
		// Case 8: the store answers 500 — fail open, no output, exit 0.
		{"case 8: store 500 -> exit 0, empty output", func(t *testing.T) {
			s := newStageStore(t, 500, stageBody(1, stageRow("fixture-one")))
			out, rc := run(t, s.addr, `{"prompt":"broken","session_id":"abc"}`)
			silent(t, out, rc)
		}},
		// Case 9: the newest row is an unattached plan session (no changeName) — no
		// change to name, no output, exit 0.
		{"case 9: nil changeName -> exit 0, empty output", func(t *testing.T) {
			s := newStageStore(t, 200, stageBody(1, stageRow("")))
			out, rc := run(t, s.addr, `{"prompt":"broken","session_id":"abc"}`)
			silent(t, out, rc)
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			c.fn(t)
		})
	}
}

// TestFlowActiveChangeHookInteractive: the hook branches on the command of the
// session's latest stage run — /flow-interactive before flow.write-finished
// names the session's change and the request protocol, a finished one and a
// /flow-fast one inject nothing, and /flow keeps its fix-run line.
func TestFlowActiveChangeHookInteractive(t *testing.T) {
	t.Parallel()
	row := func(command, stage string) string {
		return `{"stageRunId":7,"command":"` + command + `","stage":"` + stage + `","startedAt":"2026-10-30T09:00:00Z","changeName":"kan-9-chat"}`
	}
	const interactiveLine = "flow: this session runs /flow-interactive change `kan-9-chat`. A plain message is its next request" +
		" — or its stop (**One request**, **Explicit stop**, `skills/flow-interactive/SKILL.md`)."
	for _, tc := range []struct {
		name, command, stage, want string
	}{
		{"interactive mid-session names the change", "/flow-interactive", "flow.sdd-tdd", interactiveLine},
		{"interactive finished injects nothing", "/flow-interactive", "flow.write-finished", ""},
		{"flow-fast injects nothing", "/flow-fast", "flow.sdd-tdd", ""},
		{"flow keeps its line", "/flow", "flow.kickoff", "marked change kan-9-chat"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := newStageStore(t, 200, stageBody(1, row(tc.command, tc.stage)))
			out, errOut, rc := runHook(t, "flow-active-change.py",
				[]byte(`{"prompt":"next thing","session_id":"abc"}`), "FLOW_ADDR="+s.addr)
			out = strings.TrimRight(out+errOut, "\n")
			if rc != 0 {
				t.Fatalf("rc=%d out=%s", rc, out)
			}
			if tc.want == "" {
				if out != "" {
					t.Errorf("expected no output, got %s", out)
				}
				return
			}
			var got struct {
				HookSpecificOutput struct {
					AdditionalContext string `json:"additionalContext"`
				} `json:"hookSpecificOutput"`
			}
			if err := json.Unmarshal([]byte(out), &got); err != nil {
				t.Fatalf("decode %s: %v", out, err)
			}
			if !strings.Contains(got.HookSpecificOutput.AdditionalContext, tc.want) {
				t.Errorf("expected %q in context=%s", tc.want, got.HookSpecificOutput.AdditionalContext)
			}
		})
	}
}
