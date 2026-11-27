package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

type splitStoreFake struct {
	rows   []store.PendingTimeSplit
	linked map[int64]string
	set    map[string]string
}

func (f *splitStoreFake) PendingTimeSplits(context.Context, int) ([]store.PendingTimeSplit, error) {
	return f.rows, nil
}

func (f *splitStoreFake) LinkStageRunTranscript(_ context.Context, id int64, path string) error {
	f.linked[id] = path
	return nil
}

func (f *splitStoreFake) SetTimeSplit(_ context.Context, kind string, id int64, _, _ time.Time, split json.RawMessage) error {
	f.set[kind+strconv.FormatInt(id, 10)] = string(split)
	return nil
}

// TestHarvestTimeSplits: a cc- stage run is linked to its token's
// transcript, open or ended; an ended one covered by its transcript gets its
// window's split, one not yet covered waits until timeSplitSettle has
// passed, one with no transcript gets JSON null; an ended dispatch gets its
// agent transcript's split at once, or JSON null when it has none.
func TestHarvestTimeSplits(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sess := "s-1"
	main := filepath.Join(root, "proj", sess+".jsonl")
	agent := filepath.Join(root, "proj", sess, "subagents", "agent-a1.jsonl")
	if err := os.MkdirAll(filepath.Dir(agent), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(main, []byte(`{"type":"user","timestamp":"2026-01-01T00:00:00Z","message":{"content":"go"}}
{"type":"assistant","timestamp":"2026-01-01T00:00:10Z","message":{"content":[{"type":"tool_use","id":"t1","name":"Agent","input":{}}]}}
{"type":"user","timestamp":"2026-01-01T00:01:00Z","message":{"content":[{"type":"tool_result","tool_use_id":"t1"}]}}
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agent, []byte(`{"type":"user","timestamp":"2026-01-01T00:00:10Z","message":{"content":"task"}}
{"type":"assistant","timestamp":"2026-01-01T00:00:40Z","message":{"content":[{"type":"text","text":"done"}]}}
`), 0o644); err != nil {
		t.Fatal(err)
	}
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	at := func(s int) *time.Time { v := t0.Add(time.Duration(s) * time.Second); return &v }
	f := &splitStoreFake{linked: map[int64]string{}, set: map[string]string{}, rows: []store.PendingTimeSplit{
		{Kind: store.SplitStage, ID: 1, Token: "cc-" + sess + "-main", StartedAt: t0, EndedAt: at(30)},
		{Kind: store.SplitStage, ID: 2, Token: "cc-" + sess + "-main", StartedAt: t0, EndedAt: at(90)},
		{Kind: store.SplitStage, ID: 3, Token: "cc-" + sess + "-a1", StartedAt: t0},
		{Kind: store.SplitStage, ID: 4, Token: "cc-gone-main", StartedAt: t0, EndedAt: at(30)},
		{Kind: store.SplitDispatch, ID: 5, AgentID: "a1", SessionID: "other", StartedAt: t0, EndedAt: at(50)},
		{Kind: store.SplitDispatch, ID: 6, AgentID: "codex", StartedAt: t0, EndedAt: at(50)},
	}}
	if err := harvestTimeSplits(context.Background(), root, f, *at(100)); err != nil {
		t.Fatal(err)
	}
	if f.linked[1] != main || f.linked[2] != main || f.linked[3] != agent || len(f.linked) != 3 {
		t.Errorf("linked = %v", f.linked)
	}
	want := map[string]string{
		"stage1":    `{"llmMs":10000,"toolMs":20000,"humanMs":0,"byTool":{"Agent":20000}}`,
		"stage4":    `null`,
		"dispatch5": `{"llmMs":30000,"toolMs":0,"humanMs":0}`,
		"dispatch6": `null`,
	}
	if len(f.set) != len(want) {
		t.Errorf("set = %v, want %v", f.set, want)
	}
	for k, v := range want {
		if f.set[k] != v {
			t.Errorf("%s = %s, want %s", k, f.set[k], v)
		}
	}
	f.set = map[string]string{}
	f.rows = f.rows[1:2]
	if err := harvestTimeSplits(context.Background(), root, f, at(90).Add(timeSplitSettle)); err != nil {
		t.Fatal(err)
	}
	if got := f.set["stage2"]; got != `{"llmMs":10000,"toolMs":50000,"humanMs":0,"byTool":{"Agent":50000}}` {
		t.Errorf("settled stage2 = %s", got)
	}
}
