package harvest_test

import (
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
