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
