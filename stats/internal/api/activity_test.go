package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// SetSessionActivity on statsFake keeps the last report per session and
// mirrors it onto every canned board row, so the board read below sees
// what the PUT stored.
func (f *statsFake) SetSessionActivity(_ context.Context, sessionID, state, detail string) error {
	if f.activity == nil {
		f.activity = map[string][2]string{}
	}
	f.activity[sessionID] = [2]string{state, detail}
	for i := range f.liveStateBoard {
		f.liveStateBoard[i].Activity = &state
		f.liveStateBoard[i].ActivityDetail = &detail
	}
	return nil
}

func putActivity(t *testing.T, url, session, body string) int {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, url+"/api/v1/sessions/"+session+"/activity", bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT activity: %v", err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

// TestSessionActivityPut pins the activity route: a known state is stored
// (204), an unknown state or an over-long detail is refused (400) before
// the store is touched, and the board carries the stored value back.
func TestSessionActivityPut(t *testing.T) {
	stage := "flow.sdd-tdd"
	sts := &statsFake{liveStateBoard: []store.LiveStateRow{{
		ProjectKey: "agents", Name: "kan-1", State: store.StateInProgress,
		UpdatedAt: time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC), CurrentStage: &stage,
	}}}
	ts := newStatsTestServer(t, sts)

	if got := putActivity(t, ts.URL, "s-1", `{"state":"sleeping","detail":""}`); got != http.StatusBadRequest {
		t.Errorf("unknown state: status %d, want 400", got)
	}
	long := `{"state":"question","detail":"` + strings.Repeat("x", 501) + `"}`
	if got := putActivity(t, ts.URL, "s-1", long); got != http.StatusBadRequest {
		t.Errorf("501-byte detail: status %d, want 400", got)
	}
	if len(sts.activity) != 0 {
		t.Fatalf("refused PUTs reached the store: %v", sts.activity)
	}
	if got := putActivity(t, ts.URL, "s-1", `{"state":"question","detail":"Proceed?"}`); got != http.StatusNoContent {
		t.Fatalf("valid PUT: status %d, want 204", got)
	}
	if got := sts.activity["s-1"]; got != [2]string{"question", "Proceed?"} {
		t.Errorf("stored %v, want question/Proceed?", got)
	}

	status, env, body := getStats(t, ts, periodPath("state-board"))
	if status != http.StatusOK {
		t.Fatalf("board status %d, body %s", status, body)
	}
	var rows []struct {
		CurrentStage   *string `json:"currentStage"`
		Activity       *string `json:"activity"`
		ActivityDetail *string `json:"activityDetail"`
		DoneTasks      *int    `json:"doneTasks"`
	}
	if err := json.Unmarshal(env.Rows, &rows); err != nil {
		t.Fatalf("decode rows: %v", err)
	}
	if len(rows) != 1 || rows[0].CurrentStage == nil || *rows[0].CurrentStage != stage ||
		rows[0].Activity == nil || *rows[0].Activity != "question" ||
		rows[0].ActivityDetail == nil || *rows[0].ActivityDetail != "Proceed?" || rows[0].DoneTasks != nil {
		t.Errorf("board rows = %s, want currentStage, activity question, detail Proceed?, no doneTasks", env.Rows)
	}
}
