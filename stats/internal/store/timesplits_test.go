package store_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// TestTimeSplitsPendingLinkAndSet: a cc- stage run is pending until its
// split is set (open, only while unlinked), an mf- one never is; an ended
// dispatch with an agent id is pending until set; a set split is never
// overwritten, a set for a window the row no longer has is dropped, and a
// retime clears the split so the row is pending again.
func TestTimeSplitsPendingLinkAndSet(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()
	pk := fmt.Sprintf("proj-split-%d", time.Now().UnixNano())
	seedChange(t, st, pk, "c")
	pool := store.PoolOf(st)
	var cid int64
	if err := pool.QueryRow(ctx, `SELECT id FROM changes WHERE project_key=$1 AND name='c'`, pk).Scan(&cid); err != nil {
		t.Fatal(err)
	}
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Minute)
	ins := func(sql string, args ...any) int64 {
		t.Helper()
		var id int64
		if err := pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
		return id
	}
	stage := func(token string, n int, ended *time.Time) int64 {
		return ins(`INSERT INTO stage_runs (change_id, harness, command, stage, attempt, session_token, started_at, ended_at)
			VALUES ($1,'claude-code','/flow','flow.implement',$2,$3,$4,$5) RETURNING id`, cid, n, token, t0, ended)
	}
	ended, open, legacy := stage("cc-s-main", 1, &t1), stage("cc-s-a1", 2, nil), stage("mf-x", 3, &t1)
	disp := ins(`INSERT INTO dispatches (change_id, seq, role, model, agent_id, started_at, ended_at)
		VALUES ($1,1,'operator-asked','opus','a1',$2,$3) RETURNING id`, cid, t0, t1)
	type key struct {
		kind string
		id   int64
	}
	pending := func() map[key]store.PendingTimeSplit {
		t.Helper()
		ps, err := st.PendingTimeSplits(ctx, 100000)
		if err != nil {
			t.Fatal(err)
		}
		out := map[key]store.PendingTimeSplit{}
		for _, p := range ps {
			if p.Kind == store.SplitStage && (p.ID == ended || p.ID == open || p.ID == legacy) || p.Kind == store.SplitDispatch && p.ID == disp {
				out[key{p.Kind, p.ID}] = p
			}
		}
		return out
	}
	got := pending()
	sEnded, sOpen, dDisp := key{store.SplitStage, ended}, key{store.SplitStage, open}, key{store.SplitDispatch, disp}
	if len(got) != 3 || got[sEnded].Token != "cc-s-main" || got[sOpen].EndedAt != nil || got[dDisp].AgentID != "a1" {
		t.Fatalf("pending = %+v, want the two cc- stage runs and the dispatch", got)
	}
	if err := st.LinkStageRunTranscript(ctx, open, "/t/s/subagents/agent-a1.jsonl"); err != nil {
		t.Fatal(err)
	}
	if err := st.LinkStageRunTranscript(ctx, ended, "/t/s.jsonl"); err != nil {
		t.Fatal(err)
	}
	split := json.RawMessage(`{"llmMs":1,"toolMs":2,"humanMs":3}`)
	if err := st.SetTimeSplit(ctx, store.SplitDispatch, disp, t0, t0.Add(time.Second), split); err != nil {
		t.Fatal(err)
	}
	got = pending()
	if _, ok := got[sOpen]; ok || got[sEnded].TranscriptPath != "/t/s.jsonl" {
		t.Fatalf("after link: pending = %+v, want the open run gone and the ended one linked", got)
	}
	if _, ok := got[dDisp]; !ok {
		t.Fatal("a set for another window landed")
	}
	for _, s := range []struct {
		kind string
		id   int64
	}{{store.SplitStage, ended}, {store.SplitDispatch, disp}} {
		if err := st.SetTimeSplit(ctx, s.kind, s.id, t0, t1, split); err != nil {
			t.Fatal(err)
		}
		if err := st.SetTimeSplit(ctx, s.kind, s.id, t0, t1, json.RawMessage(`null`)); err != nil {
			t.Fatal(err)
		}
	}
	if got = pending(); len(got) != 0 {
		t.Fatalf("after set: pending = %+v, want none", got)
	}
	var kept string
	if err := pool.QueryRow(ctx, `SELECT time_split::text FROM dispatches WHERE id = $1`, disp).Scan(&kept); err != nil || kept != `{"llmMs": 1, "toolMs": 2, "humanMs": 3}` {
		t.Fatalf("kept split = %s (%v), want the first set", kept, err)
	}
	if _, err := st.RetimeDispatch(ctx, pk, "c", records.DispatchRetime{ID: disp, StartedAt: t0, EndedAt: t1.Add(time.Minute), Reason: "late end"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := pending()[dDisp]; !ok {
		t.Fatal("a retimed dispatch kept its split")
	}
}
