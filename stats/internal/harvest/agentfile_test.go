package harvest

import (
	"encoding/json"
	"testing"
	"time"
)

// The agent-file attribution pass's own tests. One resumed agent shares
// its agentId across several dispatch rows, and attributeAgentFileRecords
// splits that agent's transcript usage among its own rows by
// (started_at, id) order — never across agents, which is the error the
// refuse-to-guess rule exists to prevent (kan-357, design.md's
// resumed-split-by-started-at).

func TestAgentFileRecordsCreditOneRowWhole(t *testing.T) {
	row := DispatchWindow{DispatchID: 7, AgentID: "a1", StartedAt: time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC)}
	records := []Record{
		{SessionID: "sess-parent", AgentID: "a1", IsSidechain: true, Timestamp: row.StartedAt.Add(time.Minute), Usage: Usage{InputTokens: 10, OutputTokens: 5}},
		{SessionID: "sess-parent", AgentID: "a1", IsSidechain: true, Timestamp: row.StartedAt.Add(2 * time.Minute), Usage: Usage{InputTokens: 20}},
	}
	got := attributeAgentFileRecords([]DispatchWindow{row}, records)
	if len(got) != 1 {
		t.Fatalf("deltas cover %d rows, want 1", len(got))
	}
	if got[7].Tokens.Sidechain.Input != 30 || got[7].Tokens.Sidechain.Output != 5 {
		t.Errorf("row 7 = %+v, want the batch's whole sidechain usage", got[7])
	}
	if got[7].Tokens.Main.Input != 0 {
		t.Errorf("row 7 main bucket = %+v, want zero — an agent file's spend is sidechain spend", got[7].Tokens.Main)
	}
}

func TestAgentFileRecordsSplitAcrossResumedRows(t *testing.T) {
	first := DispatchWindow{DispatchID: 10, AgentID: "a1", StartedAt: time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC)}
	second := DispatchWindow{DispatchID: 12, AgentID: "a1", StartedAt: first.StartedAt.Add(10 * time.Minute)}
	records := []Record{
		{AgentID: "a1", IsSidechain: true, Timestamp: first.StartedAt.Add(time.Minute), Usage: Usage{InputTokens: 10}},
		{AgentID: "a1", IsSidechain: true, Timestamp: second.StartedAt.Add(time.Minute), Usage: Usage{InputTokens: 40}},
	}
	got := attributeAgentFileRecords([]DispatchWindow{first, second}, records)
	if got[10].Tokens.Sidechain.Input != 10 || got[12].Tokens.Sidechain.Input != 40 {
		t.Errorf("split = %+v, want each resume's own records only", got)
	}
}

func TestAgentFileRecordsBeforeTheFirstRowFloorToIt(t *testing.T) {
	row := DispatchWindow{DispatchID: 10, AgentID: "a1", StartedAt: time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC)}
	records := []Record{
		{AgentID: "a1", IsSidechain: true, Timestamp: row.StartedAt.Add(-time.Minute), Usage: Usage{InputTokens: 3}},
	}
	got := attributeAgentFileRecords([]DispatchWindow{row}, records)
	if got[10].Tokens.Sidechain.Input != 3 {
		t.Errorf("row 10 = %+v, want the pre-start record floored to the first row", got[10])
	}
}

func TestAgentFileRecordsWithNoRowsCreditNothing(t *testing.T) {
	records := []Record{{AgentID: "a1", IsSidechain: true, Usage: Usage{InputTokens: 1}}}
	if got := attributeAgentFileRecords(nil, records); len(got) != 0 {
		t.Errorf("deltas = %+v, want none — an agent with no rows has nothing to credit", got)
	}
}

// TestAgentFileDispatchPatchCreditsTheWholeFile pins the restamp repair's
// input: one agent transcript read from offset 0 yields that agent's whole
// sidechain usage and signals, in the shape MergeDispatchMetrics stores.
func TestAgentFileDispatchPatchCreditsTheWholeFile(t *testing.T) {
	patch, err := AgentFileDispatchPatch("../../testdata/transcripts/signals-agent.jsonl")
	if err != nil {
		t.Fatalf("AgentFileDispatchPatch: %v", err)
	}
	var got MetricsPatch
	if err := json.Unmarshal(patch, &got); err != nil {
		t.Fatalf("decode patch %s: %v", patch, err)
	}
	if got.Tokens.Sidechain.Input != 5 || got.Tokens.Sidechain.Output != 7 || got.Tokens.Main.Input != 0 {
		t.Errorf("tokens = %+v, want the file's sidechain usage (input 5, output 7) and nothing on main", got.Tokens)
	}
	if got.Signals == nil || got.Signals.Sidechain.ToolCallsTotal == 0 {
		t.Errorf("signals = %+v, want the file's tool calls counted", got.Signals)
	}
}
