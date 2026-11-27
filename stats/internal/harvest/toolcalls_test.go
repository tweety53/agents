package harvest_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/harvest"
)

// TestAgentToolCallsPairsUseWithResult: an agent's transcript is found under
// root as AgentTranscriptEnd finds it, each tool_use pairs with the
// tool_result naming its id -- two parallel calls included -- the span named
// by its tool and a short form of its input, a call never answered dropped,
// and a missing transcript reported as not found.
func TestAgentToolCallsPairsUseWithResult(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dir := filepath.Join(root, "proj", "sess", "subagents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	long := "go test ./internal/store/ -run 'TestTraceSpansOfOneChange|TestBottlenecksRanksEachDimension' -count=1\nsecond line"
	lines := `{"type":"assistant","timestamp":"2026-01-01T00:00:10Z","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":` + strconv.Quote(long) + `}},{"type":"tool_use","id":"t2","name":"Read","input":{"file_path":"/a/b.go"}}]}}
{"type":"user","timestamp":"2026-01-01T00:00:11Z","message":{"content":[{"type":"tool_result","tool_use_id":"t2"}]}}
not json
{"type":"user","timestamp":"2026-01-01T00:00:14Z","message":{"content":[{"type":"tool_result","tool_use_id":"t1","is_error":true}]}}
{"type":"assistant","timestamp":"2026-01-01T00:00:15Z","message":{"content":[{"type":"tool_use","id":"t3","name":"Grep","input":{"pattern":"x"}}]}}
`
	if err := os.WriteFile(filepath.Join(dir, "agent-a1.jsonl"), []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}
	got, found, err := harvest.AgentToolCalls(root, "a1")
	if err != nil || !found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	want := []harvest.ToolCall{
		{Name: "Bash: go test ./internal/store/ -run 'TestTraceSpansOfOneChange|Te…", Start: t0.Add(10 * time.Second), End: t0.Add(14 * time.Second), IsError: true},
		{Name: "Read: /a/b.go", Start: t0.Add(10 * time.Second), End: t0.Add(11 * time.Second)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("calls:\n got %+v\nwant %+v", got, want)
	}
	if _, found, err := harvest.AgentToolCalls(root, "absent"); found || err != nil {
		t.Errorf("absent: found=%v err=%v", found, err)
	}
}

// TestAgentTimeSplitClassesEachGap: every gap between two consecutive
// timestamped lines is classed by the line that ends it -- an assistant line
// is llm time, a user line answering a tool_use is tool time under that
// tool (Bash by its command word), any other user line is human time; a gap
// is clipped to the window and a line stamped before its predecessor adds
// nothing.
func TestAgentTimeSplitClassesEachGap(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dir := filepath.Join(root, "proj", "sess", "subagents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	lines := `{"type":"user","timestamp":"2026-01-01T00:00:00Z","message":{"content":"do it"}}
{"type":"assistant","timestamp":"2026-01-01T00:00:04Z","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"go test ./x"}},{"type":"tool_use","id":"t2","name":"Read","input":{"file_path":"/a"}}]}}
{"type":"user","timestamp":"2026-01-01T00:00:05Z","message":{"content":[{"type":"tool_result","tool_use_id":"t2"}]}}
{"type":"user","timestamp":"2026-01-01T00:00:15Z","message":{"content":[{"type":"tool_result","tool_use_id":"t1"}]}}
{"type":"assistant","timestamp":"2026-01-01T00:00:17Z","message":{"content":[{"type":"text","text":"done?"}]}}
{"type":"user","timestamp":"2026-01-01T00:01:17Z","message":{"content":"yes"}}
{"type":"assistant","timestamp":"2026-01-01T00:01:10Z","message":{"content":[{"type":"text","text":"backwards"}]}}
`
	if err := os.WriteFile(filepath.Join(dir, "agent-a1.jsonl"), []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	got, found, err := harvest.AgentTimeSplit(root, "a1", t0, t0.Add(time.Hour))
	if err != nil || !found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	want := harvest.TimeSplit{LLMMs: 6000, ToolMs: 11000, HumanMs: 60000, ByTool: map[string]int64{"Read": 1000, "Bash go": 10000}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("split:\n got %+v\nwant %+v", got, want)
	}
	clipped, _, _ := harvest.AgentTimeSplit(root, "a1", t0.Add(2*time.Second), t0.Add(10*time.Second))
	want = harvest.TimeSplit{LLMMs: 2000, ToolMs: 6000, ByTool: map[string]int64{"Read": 1000, "Bash go": 5000}}
	if !reflect.DeepEqual(clipped, want) {
		t.Errorf("clipped:\n got %+v\nwant %+v", clipped, want)
	}
}
