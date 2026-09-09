package harvest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Fixture lines recorded from a real ZCode rollout file
// (~/.zcode/cli/rollout/model-io-sess_fa015a84-16a9-4a47-8d9c-113ed8da1a2a.jsonl,
// 2026-09-09), trimmed of request-body bulk only -- the key structure and the
// fields this package decodes are verbatim, and the fields it does not decode
// (requestId, durationMs, querySource, traceId, turnId, attempt) are present
// to prove they are tolerated.
const rolloutFullLine = `{"completedAt":"2026-09-09T15:16:56.708Z","durationMs":14314,"requestId":"fa7338a5","attempt":1,` +
	`"querySource":"main_turn","traceId":"4415981","turnId":"t1","type":"model_io",` +
	`"sessionId":"sess_fa015a84-16a9-4a47-8d9c-113ed8da1a2a",` +
	`"model":{"modelId":"GLM-5.3-Flash","providerId":"builtin:zai-coding-plan","role":"main","source":"session","variant":"high"},` +
	`"request":{"messages":[` +
	`{"role":"user","content":"run the marks"},` +
	`{"role":"assistant","toolCalls":[{"id":"call_1","name":"Read","input":{"file_path":"/tmp/x"}},{"id":"call_2","name":"Bash","input":{"command":"flow stage begin -command '/flow' -stage flow.brainstorm -harness zcode -session-token mf-kan479-z9k4 kan-479 --session sess_fa015a84"}}]},` +
	`{"role":"tool","toolCallId":"call_2","content":"ok"}]},` +
	`"response":{"usage":{"inputTokens":51338,"outputTokens":187,"totalTokens":51525,"cacheReadTokens":10176,"cacheWriteTokens":42},` +
	`"toolCalls":[{"id":"call_3","name":"Bash","input":{"command":"echo next"}}]}}`

const rolloutCacheWriteLine = `{"type":"model_io","sessionId":"sess_244b7722-d7d5-4fa8-b163-58faab1b6ea8",` +
	`"completedAt":"2026-09-09T12:00:00.000Z",` +
	`"model":{"modelId":"GLM-5.3-Flash"},` +
	`"response":{"usage":{"inputTokens":100,"outputTokens":5,"totalTokens":105,"cacheReadTokens":0,"cacheWriteTokens":3000}}}`

func TestParseRolloutRecords(t *testing.T) {
	t.Run("full line yields one record with usage and model", func(t *testing.T) {
		records := ParseRolloutRecords([]byte(rolloutFullLine + "\n"))
		if len(records) != 1 {
			t.Fatalf("got %d records, want 1", len(records))
		}
		r := records[0]
		if r.SessionID != "sess_fa015a84-16a9-4a47-8d9c-113ed8da1a2a" {
			t.Errorf("SessionID = %q", r.SessionID)
		}
		if r.Model != "GLM-5.3-Flash" {
			t.Errorf("Model = %q", r.Model)
		}
		want := time.Date(2026, 9, 9, 15, 16, 56, 708*int(time.Millisecond), time.UTC)
		if !r.Timestamp.Equal(want) {
			t.Errorf("Timestamp = %v, want %v", r.Timestamp, want)
		}
		if r.Usage.InputTokens != 51338 || r.Usage.OutputTokens != 187 ||
			r.Usage.CacheReadInputTokens != 10176 {
			t.Errorf("Usage = %+v", r.Usage)
		}
		if r.Usage.CacheCreationInputTokens != 42 {
			t.Errorf("CacheCreationInputTokens = %d, want 42", r.Usage.CacheCreationInputTokens)
		}
		if r.Usage.CacheSplitKnown {
			t.Errorf("CacheSplitKnown = true, want false: the rollout carries a single cacheWriteTokens with no 5m/1h split")
		}
		if r.IsSidechain || r.AgentID != "" || r.Effort != "" || r.Usage.Speed != "" {
			t.Errorf("rollout lines carry no sidechain/agent/effort/speed; got %+v", r)
		}
	})

	t.Run("non model_io type skipped", func(t *testing.T) {
		line := `{"type":"other","sessionId":"sess_x","completedAt":"2026-09-09T12:00:00.000Z"}`
		if got := ParseRolloutRecords([]byte(line + "\n")); len(got) != 0 {
			t.Fatalf("got %d records, want 0", len(got))
		}
	})

	t.Run("missing usage skipped", func(t *testing.T) {
		line := `{"type":"model_io","sessionId":"sess_x","completedAt":"2026-09-09T12:00:00.000Z","model":{"modelId":"GLM-5.3-Flash"}}`
		if got := ParseRolloutRecords([]byte(line + "\n")); len(got) != 0 {
			t.Fatalf("got %d records, want 0", len(got))
		}
	})

	t.Run("invalid json skipped", func(t *testing.T) {
		if got := ParseRolloutRecords([]byte("{not json\n")); len(got) != 0 {
			t.Fatalf("got %d records, want 0", len(got))
		}
	})

	t.Run("unparseable completedAt skipped", func(t *testing.T) {
		line := `{"type":"model_io","sessionId":"sess_x","completedAt":"not-a-time","model":{"modelId":"m"},"response":{"usage":{"inputTokens":1,"outputTokens":1,"cacheReadTokens":0,"cacheWriteTokens":0}}}`
		if got := ParseRolloutRecords([]byte(line + "\n")); len(got) != 0 {
			t.Fatalf("got %d records, want 0: a record with no timestamp cannot sit in any window", len(got))
		}
	})

	t.Run("cache write lands in the unknown split", func(t *testing.T) {
		records := ParseRolloutRecords([]byte(rolloutCacheWriteLine + "\n"))
		if len(records) != 1 {
			t.Fatalf("got %d records, want 1", len(records))
		}
		u := records[0].Usage
		if u.CacheCreationInputTokens != 3000 || u.CacheSplitKnown {
			t.Errorf("Usage = %+v; want 3000 cache-creation tokens, split unknown", u)
		}
	})
}

func TestParseRolloutCommandRecords(t *testing.T) {
	t.Run("request-side tool calls", func(t *testing.T) {
		got := ParseRolloutCommandRecords([]byte(rolloutFullLine + "\n"))
		var found []string
		for _, c := range got {
			if strings.Contains(c.Command, "stage begin") {
				found = append(found, c.Command)
			}
			if c.SessionID != "sess_fa015a84-16a9-4a47-8d9c-113ed8da1a2a" {
				t.Errorf("SessionID = %q", c.SessionID)
			}
		}
		if len(found) != 1 {
			t.Fatalf("got %d stage-begin commands from request.messages, want 1", len(found))
		}
	})

	t.Run("response-side tool calls", func(t *testing.T) {
		got := ParseRolloutCommandRecords([]byte(rolloutFullLine + "\n"))
		var found int
		for _, c := range got {
			if c.Command == "echo next" {
				found++
			}
		}
		if found != 1 {
			t.Fatalf("got %d response-side commands, want 1", found)
		}
	})

	t.Run("non-Bash tool calls ignored", func(t *testing.T) {
		got := ParseRolloutCommandRecords([]byte(rolloutFullLine + "\n"))
		for _, c := range got {
			if strings.Contains(c.Command, "/tmp/x") {
				t.Fatalf("non-Bash tool call leaked through: %q", c.Command)
			}
		}
	})
}

func TestDefaultZcodeRolloutRoot(t *testing.T) {
	t.Run("env override wins", func(t *testing.T) {
		t.Setenv(DefaultZcodeRolloutRootEnv, "/tmp/rollouts")
		root, err := DefaultZcodeRolloutRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root != "/tmp/rollouts" {
			t.Errorf("root = %q, want /tmp/rollouts", root)
		}
	})

	t.Run("default under the user's home", func(t *testing.T) {
		t.Setenv(DefaultZcodeRolloutRootEnv, "")
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skipf("no home directory: %v", err)
		}
		root, err := DefaultZcodeRolloutRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root != filepath.Join(home, ".zcode", "cli", "rollout") {
			t.Errorf("root = %q", root)
		}
	})
}
