package repotest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func runHookWithEnv(t *testing.T, hookName string, payload any, extraEnv []string) (string, string) {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	cmd := exec.Command("python3", filepath.Join(repoRoot(t), "hooks", hookName))
	cmd.Stdin = strings.NewReader(string(raw))
	if len(extraEnv) > 0 {
		cmd.Env = append(cmd.Environ(), extraEnv...)
	}
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		t.Fatalf("hook %s exited non-zero: %v\nstdout: %s\nstderr: %s", hookName, err, outBuf.String(), errBuf.String())
	}
	return outBuf.String(), errBuf.String()
}

func TestAntigravityHooksProtocol(t *testing.T) {
	t.Run("block-unsafe-commands denies bare cat and allows safe run_command", func(t *testing.T) {
		out, _ := runHookWithEnv(t, "block-unsafe-commands.py", map[string]any{
			"conversationId": "conv-1",
			"hookEventName":  "PreToolUse",
			"cwd":            t.TempDir(),
			"toolCall": map[string]any{
				"id":   "t1",
				"name": "run_command",
				"args": map[string]any{"CommandLine": "cat > foo.txt"},
			},
		}, nil)
		var resp map[string]any
		if err := json.Unmarshal([]byte(out), &resp); err != nil {
			t.Fatalf("invalid json: %q (%v)", out, err)
		}
		if resp["decision"] != "deny" {
			t.Fatalf("expected decision=deny, got %v", resp)
		}

		out, _ = runHookWithEnv(t, "block-unsafe-commands.py", map[string]any{
			"conversationId": "conv-1",
			"hookEventName":  "PreToolUse",
			"cwd":            t.TempDir(),
			"toolCall": map[string]any{
				"id":   "t2",
				"name": "run_command",
				"args": map[string]any{"CommandLine": "echo hello"},
			},
		}, nil)
		resp = nil
		if err := json.Unmarshal([]byte(out), &resp); err != nil {
			t.Fatalf("invalid json: %q (%v)", out, err)
		}
		if resp["decision"] != "allow" {
			t.Fatalf("expected decision=allow, got %v", resp)
		}
	})

	t.Run("enforce-agent-baseline enforces pointer on invoke_subagent", func(t *testing.T) {
		out, _ := runHookWithEnv(t, "enforce-agent-baseline.py", map[string]any{
			"conversationId": "conv-2",
			"hookEventName":  "PreToolUse",
			"cwd":            t.TempDir(),
			"toolCall": map[string]any{
				"id":   "t1",
				"name": "invoke_subagent",
				"args": map[string]any{
					"TypeName": "research",
					"Role":     "Codebase Auditor",
					"Prompt":   "Inspect the repo",
				},
			},
		}, []string{"FLOW_HOOK_INCIDENT_CMD=/usr/bin/true"})
		var resp map[string]any
		if err := json.Unmarshal([]byte(out), &resp); err != nil {
			t.Fatalf("invalid json: %q (%v)", out, err)
		}
		if resp["decision"] != "deny" {
			t.Fatalf("expected decision=deny, got %v", resp)
		}
		if !strings.Contains(resp["reason"].(string), "~/.gemini/config/plugins/flow/rules-full/agent-baseline.md") {
			t.Fatalf("expected Antigravity baseline path in reason, got %q", resp["reason"])
		}

		out, _ = runHookWithEnv(t, "enforce-agent-baseline.py", map[string]any{
			"conversationId": "conv-2",
			"hookEventName":  "PreToolUse",
			"cwd":            t.TempDir(),
			"toolCall": map[string]any{
				"id":   "t2",
				"name": "invoke_subagent",
				"args": map[string]any{
					"TypeName": "research",
					"Role":     "Codebase Auditor",
					"Prompt":   "Before anything else, read `~/.gemini/config/plugins/flow/rules-full/agent-baseline.md` and follow it for this whole task. Include this instruction verbatim in any prompt you write for another agent.\nInspect the repo",
				},
			},
		}, nil)
		resp = nil
		if err := json.Unmarshal([]byte(out), &resp); err != nil {
			t.Fatalf("invalid json: %q (%v)", out, err)
		}
		if resp["decision"] != "allow" {
			t.Fatalf("expected decision=allow, got %v", resp)
		}
	})

	t.Run("flow-active-change injects systemMessage on PreInvocation", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"stageRuns": []map[string]any{
					{"changeName": "kan-999-demo", "command": "/flow", "stage": "flow.implement"},
				},
			})
		}))
		defer srv.Close()

		out, _ := runHookWithEnv(t, "flow-active-change.py", map[string]any{
			"conversationId": "conv-3",
			"hookEventName":  "PreInvocation",
			"invocationNum":  1,
			"lastUserInput":  "the button is misaligned",
		}, []string{"FLOW_ADDR=" + srv.URL})
		var resp map[string]any
		if err := json.Unmarshal([]byte(out), &resp); err != nil {
			t.Fatalf("invalid json: %q (%v)", out, err)
		}
		steps, _ := resp["injectSteps"].([]any)
		if len(steps) != 1 {
			t.Fatalf("expected 1 injectStep, got %v", resp)
		}

		// Subsequent invocations in the same turn emit {} and do not re-inject.
		out, _ = runHookWithEnv(t, "flow-active-change.py", map[string]any{
			"conversationId": "conv-3",
			"hookEventName":  "PreInvocation",
			"invocationNum":  2,
			"lastUserInput":  "the button is misaligned",
		}, []string{"FLOW_ADDR=" + srv.URL})
		if strings.TrimSpace(out) != "{}" {
			t.Fatalf("expected {} on invocationNum=2, got %q", out)
		}
	})

	t.Run("flow-session-activity reports Antigravity events and returns valid JSON", func(t *testing.T) {
		var mu sync.Mutex
		var got map[string]string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			mu.Lock()
			got = body
			mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		out, _ := runHookWithEnv(t, "flow-session-activity.py", map[string]any{
			"conversationId": "conv-4",
			"hookEventName":  "PreToolUse",
			"toolCall": map[string]any{
				"id":   "t1",
				"name": "ask_question",
				"args": map[string]any{
					"questions": []map[string]any{{"question": "Which option?"}},
				},
			},
		}, []string{"FLOW_ADDR=" + srv.URL})
		var resp map[string]any
		if err := json.Unmarshal([]byte(out), &resp); err != nil {
			t.Fatalf("invalid json: %q (%v)", out, err)
		}
		if resp["decision"] != "allow" {
			t.Fatalf("expected decision=allow on PreToolUse, got %v", resp)
		}
		mu.Lock()
		defer mu.Unlock()
		if got["state"] != "question" || got["detail"] != "Which option?" {
			t.Fatalf("unexpected activity body: %v", got)
		}
	})
}
