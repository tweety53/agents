package harvest_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/harvest"
)

// Fixture line recorded from a real ZCode subagent rollout file
// (~/.zcode/cli/rollout/model-io-sess_subagent_agent_541a9582-890b-417e-a3c3-419047bb4ee6.jsonl,
// captured 2026-09-12 while the dispatch ran), trimmed of request-body bulk
// only. A subagent dispatch writes its own rollout file named
// model-io-sess_subagent_agent_<uuid>.jsonl and stamps every line with the
// per-dispatch session id sess_subagent_agent_<uuid> plus
// querySource "subagent"; the <uuid> is byte-identical to the agentId the
// dispatcher's Agent tool result returns, which `flow record dispatch begin
// -agent-id` already records on the dispatch row.
const rolloutSubagentLine = `{"completedAt":"2026-09-12T20:00:47.950Z","durationMs":28045,"requestId":"7fe37765","attempt":1,` +
	`"querySource":"subagent","traceId":"94fc46f4","type":"model_io",` +
	`"sessionId":"sess_subagent_agent_541a9582-890b-417e-a3c3-419047bb4ee6",` +
	`"model":{"modelId":"GLM-5.3-Flash","providerId":"builtin:zai-coding-plan","role":"main","source":"session","variant":"high"},` +
	`"response":{"usage":{"inputTokens":48023,"outputTokens":1419,"totalTokens":49442,"cacheReadTokens":44800,"cacheWriteTokens":0}}}`

// rolloutSubagentLineFor returns the measured subagent fixture line with
// every occurrence of the captured dispatch's uuid replaced, so a test can
// stand up two distinct dispatches' files from the one recorded shape.
func rolloutSubagentLineFor(uuid string) string {
	return strings.ReplaceAll(rolloutSubagentLine, "541a9582-890b-417e-a3c3-419047bb4ee6", uuid)
}

func TestParseRolloutRecordsSubagentSessionMarksSidechainAndAgentID(t *testing.T) {
	records := harvest.ParseRolloutRecords([]byte(rolloutSubagentLine + "\n"))
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}
	r := records[0]
	if r.SessionID != "sess_subagent_agent_541a9582-890b-417e-a3c3-419047bb4ee6" {
		t.Errorf("SessionID = %q", r.SessionID)
	}
	if !r.IsSidechain {
		t.Errorf("IsSidechain = false, want true: a subagent dispatch's usage is its dispatcher's sidechain spend, and the dispatch-grain pass only credits sidechain records")
	}
	if r.AgentID != "agent_541a9582-890b-417e-a3c3-419047bb4ee6" {
		t.Errorf("AgentID = %q, want the dispatch's own agent id as the dispatcher recorded it", r.AgentID)
	}
	if r.Model != "GLM-5.3-Flash" || r.Usage.OutputTokens != 1419 {
		t.Errorf("usage decimation changed: Model = %q, Usage = %+v", r.Model, r.Usage)
	}

	t.Run("main-session lines still carry neither field", func(t *testing.T) {
		records := harvest.ParseRolloutRecords([]byte(rolloutMainLineForSubagentTest + "\n"))
		if len(records) != 1 {
			t.Fatalf("got %d records, want 1", len(records))
		}
		if r := records[0]; r.IsSidechain || r.AgentID != "" {
			t.Errorf("main-session record grew subagent fields: IsSidechain = %v, AgentID = %q", r.IsSidechain, r.AgentID)
		}
	})

	t.Run("a main session id that merely contains subagent does not match", func(t *testing.T) {
		line := `{"type":"model_io","sessionId":"sess_subagent_agentX-1","completedAt":"2026-09-12T20:00:47.950Z","model":{"modelId":"m"},"response":{"usage":{"inputTokens":1,"outputTokens":1,"cacheReadTokens":0,"cacheWriteTokens":0}}}`
		records := harvest.ParseRolloutRecords([]byte(line + "\n"))
		if len(records) != 1 {
			t.Fatalf("got %d records, want 1", len(records))
		}
		if r := records[0]; r.IsSidechain || r.AgentID != "" {
			t.Errorf("prefix matched a non-subagent session id: IsSidechain = %v, AgentID = %q", r.IsSidechain, r.AgentID)
		}
	})
}

// rolloutMainLineForSubagentTest is the pre-existing main-session fixture
// shape (request-body bulk trimmed), for the negative side of the
// session-id rule.
const rolloutMainLineForSubagentTest = `{"type":"model_io","sessionId":"sess_fa015a84-16a9-4a47-8d9c-113ed8da1a2a",` +
	`"completedAt":"2026-09-09T15:16:56.708Z",` +
	`"model":{"modelId":"GLM-5.3-Flash","providerId":"builtin:zai-coding-plan","role":"main","source":"session","variant":"high"},` +
	`"response":{"usage":{"inputTokens":51338,"outputTokens":187,"totalTokens":51525,"cacheReadTokens":10176,"cacheWriteTokens":42}}}`

func TestAgentIDFromRolloutPath(t *testing.T) {
	t.Run("subagent rollout file names its own dispatch", func(t *testing.T) {
		id, ok := harvest.AgentIDFromRolloutPath("/home/u/.zcode/cli/rollout/model-io-sess_subagent_agent_541a9582-890b-417e-a3c3-419047bb4ee6.jsonl")
		if !ok || id != "agent_541a9582-890b-417e-a3c3-419047bb4ee6" {
			t.Errorf("got (%q, %v), want the dispatch's own agent id", id, ok)
		}
	})
	t.Run("main rollout file is not one", func(t *testing.T) {
		if id, ok := harvest.AgentIDFromRolloutPath("/home/u/.zcode/cli/rollout/model-io-sess_fa015a84-16a9-4a47-8d9c-113ed8da1a2a.jsonl"); ok {
			t.Errorf("got (%q, %v), want not-ok: a main session names no dispatch", id, ok)
		}
	})
	t.Run("a claude agent transcript is not a rollout path", func(t *testing.T) {
		if id, ok := harvest.AgentIDFromRolloutPath("/proj/sess/subagents/agent-a1.jsonl"); ok {
			t.Errorf("got (%q, %v), want not-ok: the claude shape has its own resolver", id, ok)
		}
	})
}

// rolloutDispatchDeps stands in for the dispatch-grain pass's dependencies
// over a rollout source: per-agent windows for the agent-file pass, merges
// and ambiguity stamps recorded, and every inference-pass window lookup
// counted so a test can assert a batch never reached it.
type rolloutDispatchDeps struct {
	harvest.NoDeps
	windowsByAgent map[string][]harvest.DispatchWindow
	agentQueries   []string
	sessionQueries []string
	merges         map[int64][]json.RawMessage
	unattributed   [][]int64
}

func (d *rolloutDispatchDeps) DispatchWindowsForAgent(_ context.Context, agentID string) ([]harvest.DispatchWindow, error) {
	d.agentQueries = append(d.agentQueries, agentID)
	return d.windowsByAgent[agentID], nil
}

func (d *rolloutDispatchDeps) DispatchWindowsForSession(_ context.Context, sessionID string) ([]harvest.DispatchWindow, error) {
	d.sessionQueries = append(d.sessionQueries, sessionID)
	return nil, nil
}

func (d *rolloutDispatchDeps) MergeDispatchMetrics(_ context.Context, dispatchID int64, patch json.RawMessage) error {
	if d.merges == nil {
		d.merges = make(map[int64][]json.RawMessage)
	}
	d.merges[dispatchID] = append(d.merges[dispatchID], patch)
	return nil
}

func (d *rolloutDispatchDeps) MarkDispatchesUnattributedByID(_ context.Context, ids []int64, _ string, _ int) error {
	d.unattributed = append(d.unattributed, ids)
	return nil
}

// TestRolloutSubagentBatchesCreditTheirOwnDispatchUnderConcurrency is
// KAN-506's regression: two dispatches run concurrently, their recorded
// windows overlap completely, and each dispatch's rollout file must credit
// that dispatch alone, in full, keyed on the per-dispatch session id --
// never offered to the window-inference pass that stamps such pairs
// "matched more than one dispatch".
func TestRolloutSubagentBatchesCreditTheirOwnDispatchUnderConcurrency(t *testing.T) {
	dir := t.TempDir()
	const uuid1 = "11111111-1111-4111-8111-111111111111"
	const uuid2 = "22222222-2222-4222-8222-222222222222"
	for _, uuid := range []string{uuid1, uuid2} {
		name := "model-io-sess_subagent_agent_" + uuid + ".jsonl"
		if err := os.WriteFile(filepath.Join(dir, name), []byte(rolloutSubagentLineFor(uuid)+"\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	// Overlapping windows: a timestamp-based choice could not tell these
	// two apart, which is precisely the case the session-id key exists to
	// make irrelevant.
	base := time.Date(2026, 9, 12, 20, 0, 0, 0, time.UTC)
	end1 := base.Add(30 * time.Minute)
	end2 := base.Add(35 * time.Minute)
	deps := &rolloutDispatchDeps{windowsByAgent: map[string][]harvest.DispatchWindow{
		"agent_" + uuid1: {{DispatchID: 101, AgentID: "agent_" + uuid1, StartedAt: base, EndedAt: &end1}},
		"agent_" + uuid2: {{DispatchID: 202, AgentID: "agent_" + uuid2, StartedAt: base.Add(5 * time.Minute), EndedAt: &end2}},
	}}

	w := harvest.NewWatcher([]harvest.Source{harvest.NewRolloutSource(dir)}, newFakeHarvestSink(), harvest.NewAttributor(&fakeWindowSource{}), deps, nil)
	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	if len(deps.sessionQueries) != 0 {
		t.Errorf("inference pass queried sessions %v, want none: a batch whose file names its own dispatch is never offered to window inference", deps.sessionQueries)
	}
	if len(deps.unattributed) != 0 {
		t.Errorf("ambiguity stamps = %v, want none: keyed attribution leaves nothing to be told apart", deps.unattributed)
	}
	if len(deps.agentQueries) != 2 {
		t.Errorf("agent window lookups = %v, want one per dispatch's file", deps.agentQueries)
	}
	for _, dispatchID := range []int64{101, 202} {
		patches := deps.merges[dispatchID]
		if len(patches) == 0 {
			t.Fatalf("dispatch %d received no merged figures", dispatchID)
		}
		var got int64
		for _, p := range patches {
			var mp harvest.MetricsPatch
			if err := json.Unmarshal(p, &mp); err != nil {
				t.Fatalf("dispatch %d patch: %v", dispatchID, err)
			}
			got += mp.Tokens.Sidechain.Output
		}
		if got != 1419 {
			t.Errorf("dispatch %d sidechain output = %d, want the file's full 1419", dispatchID, got)
		}
	}
}

// TestMainRolloutBatchNeverReachesTheAgentPass pins the routing boundary:
// only a subagent-named rollout file routes to the agent-file pass. A
// main-session rollout batch has no dispatch to name, and inventing one
// would be exactly the guess attribution refuses.
func TestMainRolloutBatchNeverReachesTheAgentPass(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "model-io-sess_fa015a84-16a9-4a47-8d9c-113ed8da1a2a.jsonl"), []byte(rolloutMainLineForSubagentTest+"\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	deps := &rolloutDispatchDeps{windowsByAgent: map[string][]harvest.DispatchWindow{}}
	w := harvest.NewWatcher([]harvest.Source{harvest.NewRolloutSource(dir)}, newFakeHarvestSink(), harvest.NewAttributor(&fakeWindowSource{}), deps, nil)
	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	if len(deps.agentQueries) != 0 {
		t.Errorf("agent window lookups = %v, want none: a main-session rollout batch names no dispatch", deps.agentQueries)
	}
	if len(deps.merges) != 0 {
		t.Errorf("merges = %v, want none", deps.merges)
	}
}
