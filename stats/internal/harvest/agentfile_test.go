package harvest

import (
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
	if got[7].Sidechain.Input != 30 || got[7].Sidechain.Output != 5 {
		t.Errorf("row 7 = %+v, want the batch's whole sidechain usage", got[7])
	}
	if got[7].Main.Input != 0 {
		t.Errorf("row 7 main bucket = %+v, want zero — an agent file's spend is sidechain spend", got[7].Main)
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
	if got[10].Sidechain.Input != 10 || got[12].Sidechain.Input != 40 {
		t.Errorf("split = %+v, want each resume's own records only", got)
	}
}

func TestAgentFileRecordsBeforeTheFirstRowFloorToIt(t *testing.T) {
	row := DispatchWindow{DispatchID: 10, AgentID: "a1", StartedAt: time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC)}
	records := []Record{
		{AgentID: "a1", IsSidechain: true, Timestamp: row.StartedAt.Add(-time.Minute), Usage: Usage{InputTokens: 3}},
	}
	got := attributeAgentFileRecords([]DispatchWindow{row}, records)
	if got[10].Sidechain.Input != 3 {
		t.Errorf("row 10 = %+v, want the pre-start record floored to the first row", got[10])
	}
}

func TestAgentFileRecordsWithNoRowsCreditNothing(t *testing.T) {
	records := []Record{{AgentID: "a1", IsSidechain: true, Usage: Usage{InputTokens: 1}}}
	if got := attributeAgentFileRecords(nil, records); len(got) != 0 {
		t.Errorf("deltas = %+v, want none — an agent with no rows has nothing to credit", got)
	}
}
