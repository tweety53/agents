package harvest_test

import (
	"context"
	"testing"

	"github.com/tweety53/agents/stats/internal/harvest"
)

const signalsFixture = "../../testdata/transcripts/signals.jsonl"

func countKinds(records []harvest.Record) map[harvest.SignalKind]int {
	out := map[harvest.SignalKind]int{}
	for _, r := range records {
		if r.Signal != nil {
			out[r.Signal.Kind]++
		}
	}
	return out
}

func TestParseSignalRecordsFromFixture(t *testing.T) {
	complete, _ := harvest.SplitCompleteLines(readFixture(t, signalsFixture))
	records := harvest.ParseSignalRecords(complete)

	got := countKinds(records)
	want := map[harvest.SignalKind]int{
		harvest.SignalToolUse: 3, harvest.SignalToolResult: 4, harvest.SignalTurn: 2,
		harvest.SignalAPIError: 1, harvest.SignalCompaction: 1,
	}
	for k, n := range want {
		if got[k] != n {
			t.Errorf("%s: got %d, want %d", k, got[k], n)
		}
	}

	var denied, errored int
	tools := map[string]int{}
	for _, r := range records {
		if r.Signal == nil {
			continue
		}
		if r.SessionID != "session-signals-1" {
			t.Errorf("signal record carries session %q", r.SessionID)
		}
		switch r.Signal.Kind {
		case harvest.SignalToolResult:
			if r.Signal.IsError {
				errored++
			}
			if r.Signal.Denied {
				denied++
			}
		case harvest.SignalToolUse:
			tools[r.Signal.ToolName]++
		case harvest.SignalCompaction:
			c := r.Signal.Compaction
			if c == nil || c.Trigger != "auto" || c.PreTokens != 180000 || c.PostTokens != 12000 || c.DurationMs != 9000 {
				t.Errorf("compaction = %+v", c)
			}
		case harvest.SignalTurn:
			if r.Signal.DurationMs == 0 || r.Signal.Messages == 0 {
				t.Errorf("turn signal missing figures: %+v", r.Signal)
			}
		}
	}
	if errored != 4 || denied != 3 {
		t.Errorf("tool results errored=%d denied=%d, want 4 and 3", errored, denied)
	}
	if tools["Bash"] != 1 || tools["Read"] != 2 {
		t.Errorf("tool_use counts = %v, want Bash:1 Read:2", tools)
	}
}

func TestParseSignalRecordsDeduplicatesRepeatedBlocks(t *testing.T) {
	line := []byte(`{"type":"assistant","isSidechain":false,"sessionId":"s","timestamp":"2026-09-01T10:00:00Z","message":{"id":"m","model":"claude-opus-5","content":[{"type":"tool_use","id":"toolu_x","name":"Bash","input":{}}],"usage":{"input_tokens":1}}}` + "\n")
	twice := append(append([]byte{}, line...), line...)
	got := countKinds(harvest.ParseSignalRecords(twice))
	if got[harvest.SignalToolUse] != 1 {
		t.Errorf("tool_use counted %d times for one repeated block id, want 1", got[harvest.SignalToolUse])
	}
}

func TestAttributeCarriesSignalsSplitByMainAndSidechain(t *testing.T) {
	main, _ := harvest.SplitCompleteLines(readFixture(t, signalsFixture))
	agent, _ := harvest.SplitCompleteLines(readFixture(t, "../../testdata/transcripts/signals-agent.jsonl"))
	records := append(harvest.ParseAssistantRecords(main), harvest.ParseSignalRecords(main)...)
	records = append(records, harvest.ParseAssistantRecords(agent)...)
	records = append(records, harvest.ParseSignalRecords(agent)...)

	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
		"session-signals-1": {{StageRunID: 7, Attempt: 1, SessionID: "session-signals-1",
			StartedAt: mustParse(t, "2026-09-01T09:00:00Z")}},
	}}
	deltas, err := harvest.NewAttributor(windows).Attribute(context.Background(), records)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	d := deltas[7]
	m := d.Signals.Main
	if m.Compactions != 1 || m.Turns != 2 || m.TurnDurationMs != 8000 || m.TurnMessages != 8 {
		t.Errorf("main compactions/turns = %d/%d (%d ms, %d msgs)", m.Compactions, m.Turns, m.TurnDurationMs, m.TurnMessages)
	}
	if m.ToolCalls["Bash"] != 1 || m.ToolCalls["Read"] != 2 || m.ToolCallsTotal != 3 {
		t.Errorf("main tool calls = %v total %d", m.ToolCalls, m.ToolCallsTotal)
	}
	if m.ToolErrors != 4 || m.Denials != 3 || m.APIErrors != 1 {
		t.Errorf("main errors = %d/%d/%d", m.ToolErrors, m.Denials, m.APIErrors)
	}
	if m.ContextEnd != "11030" {
		t.Errorf("main context_end = %q, want 11030", m.ContextEnd)
	}
	if m.ServedModels["claude-opus-5"] != 2 || m.ServedModels["claude-sonnet-5"] != 1 || m.ServedModels["<synthetic>"] != 0 {
		t.Errorf("main served_models = %v", m.ServedModels)
	}
	if m.ServedEfforts["high"] != 2 || m.ServedEfforts["medium"] != 1 {
		t.Errorf("main served_efforts = %v", m.ServedEfforts)
	}
	ev, ok := m.CompactionEvents["2026-09-01T10:00:09Z"]
	if !ok || ev.PreTokens != 180000 {
		t.Errorf("compaction_events = %v", m.CompactionEvents)
	}

	s := d.Signals.Sidechain
	if s.Turns != 1 || s.ToolCalls["Grep"] != 1 || s.ContextEnd != "505" || s.ServedEfforts["low"] != 1 {
		t.Errorf("sidechain signals = %+v", s)
	}
	ds := d.DispatchSignals["agent-sig0001"]
	if ds.Turns != 1 || ds.ToolCallsTotal != 1 {
		t.Errorf("dispatch signals for agent-sig0001 = %+v", ds)
	}
}

func TestAttributeAgentFileCarriesDispatchSignals(t *testing.T) {
	agent, _ := harvest.SplitCompleteLines(readFixture(t, "../../testdata/transcripts/signals-agent.jsonl"))
	records := append(harvest.ParseAssistantRecords(agent), harvest.ParseSignalRecords(agent)...)
	windows := []harvest.DispatchWindow{{DispatchID: 3, AgentID: "agent-sig0001", StartedAt: mustParse(t, "2026-09-01T10:04:00Z")}}
	got := harvest.AttributeAgentFileRecordsForTest(windows, records)
	if got[3].Tokens.Sidechain.Input != 5 || got[3].Signals.Turns != 1 || got[3].Signals.ToolCalls["Grep"] != 1 {
		t.Errorf("agent-file delta = %+v", got[3])
	}
}

// TestAttributeSignalOnlyAgentRecordsCreateNoDispatchTokens is the guard
// task 4's review flagged: a batch made entirely of signal records
// carrying an AgentID must fold into DispatchSignals without ever
// touching Dispatches, the token map -- an agentId with signals but no
// usage has nothing to report as tokens, and a stray empty TokenDelta
// entry there would misrepresent that as "zero tokens measured" instead
// of "no tokens in this batch at all".
func TestAttributeSignalOnlyAgentRecordsCreateNoDispatchTokens(t *testing.T) {
	agent, _ := harvest.SplitCompleteLines(readFixture(t, "../../testdata/transcripts/signals-agent.jsonl"))
	records := harvest.ParseSignalRecords(agent)

	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
		"session-signals-1": {{StageRunID: 9, Attempt: 1, SessionID: "session-signals-1",
			StartedAt: mustParse(t, "2026-09-01T09:00:00Z")}},
	}}
	deltas, err := harvest.NewAttributor(windows).Attribute(context.Background(), records)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	d := deltas[9]
	if _, ok := d.Dispatches["agent-sig0001"]; ok {
		t.Errorf("Dispatches[agent-sig0001] = %+v, want no entry for a signal-only batch", d.Dispatches["agent-sig0001"])
	}
	if _, ok := d.Models["claude-sonnet-5"]; ok {
		t.Errorf("Models[claude-sonnet-5] = %+v, want no entry for a signal-only batch", d.Models["claude-sonnet-5"])
	}
	ds, ok := d.DispatchSignals["agent-sig0001"]
	if !ok || ds.Turns != 1 || ds.ToolCallsTotal != 1 {
		t.Errorf("DispatchSignals[agent-sig0001] = %+v", ds)
	}
}
