package guard

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// The begin hook records a dispatch row under its dispatcher's token -- the main thread's, else
// the launching agent's -- never the token of whatever stage run the session opened last, so the
// end-dispatch-row hook, reading the same dispatcher, ends the row it began.
func TestEnforceAgentBaselineBeginsUnderTheDispatchersToken(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"stageRuns":[{"changeName":"c1","sessionToken":"mf-t2"}]}`))
	}))
	t.Cleanup(srv.Close)
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	fixtureRepo(t, repo, repoShape{Branch: "main"})
	writeExec(t, filepath.Join(tmp, "bin", "flow"), "#!/usr/bin/env bash\necho '[]'\n")
	env := core.Env{Getenv: func(k string) string {
		switch k {
		case "FLOW_ADDR":
			return srv.URL
		case "FLOW_HOOK_ROLE_CHECK":
			return "off"
		case "PATH":
			return filepath.Join(tmp, "bin") + ":" + os.Getenv("PATH")
		}
		return ""
	}, Dir: tmp}
	yes := true
	for agent, want := range map[string]string{"": "cc-s1-main", "a9": "cc-s1-a9"} {
		payload := map[string]any{"cwd": repo, "session_id": "s1"}
		if agent != "" {
			payload["agent_id"] = agent
		}
		e := &eab{env: env, payload: payload, sessionID: "s1", cwd: repo, typ: "flow-low", session: &yes, run: &yes,
			in: map[string]any{"model": "opus", "effort": "low"}, prompt: "Role: implementer\nKey: task-1-implementer\n"}
		args, _, reason := e.beginRow()
		i := slices.Index(args, "-session-token")
		if reason != "" || i < 0 || args[i+1] != want {
			t.Errorf("agent %q: begin %q (deny %q), want -session-token %s", agent, args, reason, want)
		}
	}
}
