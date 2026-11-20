package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"net/http/httptest"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/store"
)

type fakeRoleDispatches []store.RoleDispatch

func (f fakeRoleDispatches) RoleDispatches(context.Context, store.Period, *string) ([]store.RoleDispatch, error) {
	return f, nil
}

func (f fakeRoleDispatches) ProjectKeysByDisplayName(_ context.Context, name string) ([]string, error) {
	return []string{name}, nil
}

func TestRoleOf(t *testing.T) {
	for _, c := range []struct{ stored, key, want string }{
		{"implementer", "task-3-implementer", "implementer"},
		{"implementer", "task-3-implementer-fix-1", "implementer"},
		{"implementer", "operator-2", "operator-asked"},
		{"reviewer", "task-3-reviewer", "gated-reviewer"},
		{"reviewer", "task-3-reviewer-fix-1", "gated-re-reviewer"},
		{"reviewer", "plan-review-1", "decider"},
		{"reviewer", "panel-0-primary+principles", "panel-reviewer"},
		{"reviewer", "panel-1-primary", "panel-rerun"},
		{"reviewer", "panel-2-primary", "panel-rerun"},
		{"merger", "panel-1-merger", "panel-merger"},
		{"panel-fix", "panel-fix-1", "panel-fix"},
		{"implementer", "full-suite-fix-1", "panel-fix"},
		{"verifier", "runner-1", "runner"},
		{"verifier", "verify", "runner"},
		{"verifier", "verify-gymie", "runner"},
		{"implementer", "minors-runner-fixer", "minor-fix"},
		{"verifier", "visual-verify-fix-2-gymie", "visual-verifier"},
		{"planner", "plan-1", "planner"},
		{"planner", "design-1", "plan-designer"},
		{"explorer", "explore-4", "explorer"},
		{"planner", "pipeline-fix-3-plan", "pipeline-fix-plan"},
		{"implementer", "pipeline-fix-3-fix-g2", "pipeline-fix-fix"},
		{"reviewer", "pipeline-fix-3-review-1", "pipeline-fix-review-1"},
		{"reviewer", "pipeline-fix-3-review-2", "pipeline-fix-review-r"},
		{"reviewer", "self-review-kan-1035-analyse-pass", "self-review-pass"},
		{"planner", "self-review-kan-1035-analyse-plan", "self-review-plan"},
		{"implementer", "self-review-kan-1035-analyse-fix-g1", "self-review-fix"},
		{"reviewer", "self-review-kan-1035-analyse-review-3", "self-review-review-r"},
		{"reviewer", "something-odd", "other"},
		{"conductor", "", "other"},
		{"headless", "pair-judge", "pair-judge"},
		{"headless", "decider", "decider"},
	} {
		if got := api.RoleOf(c.stored, c.key); got != c.want {
			t.Errorf("RoleOf(%q, %q) = %q, want %q", c.stored, c.key, got, c.want)
		}
	}
}

func TestRoleCostsRoute(t *testing.T) {
	at := time.Date(2026, 8, 12, 9, 0, 0, 0, time.UTC) // a Wednesday
	end := func(d time.Duration) *time.Time { e := at.Add(d); return &e }
	usd := func(v float64) *float64 { return &v }
	rows := fakeRoleDispatches{
		{Role: "implementer", Key: "task-1-implementer", Model: "opus", Effort: "low", StartedAt: at, EndedAt: end(10 * time.Minute), CostUSD: usd(2), Outcome: "completed"},
		{Role: "implementer", Key: "task-2-implementer", Model: "sonnet", Effort: "low", StartedAt: at, EndedAt: end(30 * time.Minute), CostUSD: usd(1), Outcome: "failed"},
		{Role: "implementer", Key: "task-3-implementer", Model: "opus", Effort: "low", StartedAt: at.Add(7 * 24 * time.Hour)},
		{Role: "explorer", Key: "explore-1", Model: "sonnet", Effort: "low", StartedAt: at, EndedAt: end(time.Minute), CostUSD: usd(0.1), Outcome: "completed"},
		{Role: "reviewer", Key: "panel-0-primary", Model: "opus", Effort: "medium", StartedAt: at, EndedAt: end(5 * time.Minute), CostUSD: usd(0.5), Outcome: "completed"},
	}
	table, err := os.ReadFile("../../../skills/flow-contracts/model-policy.md")
	if err != nil {
		t.Fatal(err)
	}
	fs := newFakeStore()
	srv, err := api.New(config.Config{Host: "127.0.0.1", DSN: "unused"}, fs, fs, fs, fs, fs, nil, api.WithPolicy(&fakePolicy{}, table), api.WithRoleCosts(rows))
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/api/v1/stats/role-costs?from=2026-08-01T00:00:00Z&to=2026-09-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	var got struct {
		Roles []struct {
			Role, Group                          string
			Dispatches, Priced, Failed           int
			CostUSD, AgentSeconds, MedianSeconds float64
			Pairs                                []struct {
				Pair       string
				CostUSD    float64
				Dispatches int
			}
		}
		Weeks []struct {
			Week  string
			Roles map[string]float64
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got.Roles) != 3 || got.Roles[0].Role != "explorer" || got.Roles[1].Role != "implementer" || got.Roles[2].Role != "panel-reviewer" {
		t.Fatalf("roles = %+v, want explorer (Planning), implementer, panel-reviewer: stage order before cost", got.Roles)
	}
	im := got.Roles[1]
	if im.Group != "Implementation" || im.Dispatches != 3 || im.Priced != 2 || im.Failed != 1 || im.CostUSD != 3 || im.AgentSeconds != 2400 || im.MedianSeconds != 1200 {
		t.Errorf("implementer = %+v", im)
	}
	if len(im.Pairs) != 2 || im.Pairs[0].Pair != "opus · low" || im.Pairs[0].CostUSD != 2 || im.Pairs[0].Dispatches != 2 {
		t.Errorf("implementer pairs = %+v, want opus · low first ($2, 2 dispatches)", im.Pairs)
	}
	if len(got.Weeks) != 1 || got.Weeks[0].Week != "2026-08-12" || got.Weeks[0].Roles["implementer"] != 3 || got.Weeks[0].Roles["panel-reviewer"] != 0.5 || got.Weeks[0].Roles["explorer"] != 0.1 {
		t.Errorf("weeks = %+v", got.Weeks)
	}
}

// TestRoleCostsCountsARowUnderItsPolicyRole: a row whose prompt named its
// role (-policy-role) is counted under it, not under the role its key
// derives -- a no-progress escalation keyed as the panel-fix round it
// repeats is no-progress-fixer's spend.
func TestRoleCostsCountsARowUnderItsPolicyRole(t *testing.T) {
	at := time.Date(2026, 8, 12, 9, 0, 0, 0, time.UTC)
	ended := at.Add(time.Minute)
	usd := 4.0
	rows := fakeRoleDispatches{
		{Role: "panel-fix", Key: "panel-fix-3", Model: "opus", Effort: "high", StartedAt: at, EndedAt: &ended, CostUSD: &usd, Outcome: "completed", PolicyRole: "no-progress-fixer"},
	}
	table, err := os.ReadFile("../../../skills/flow-contracts/model-policy.md")
	if err != nil {
		t.Fatal(err)
	}
	fs := newFakeStore()
	srv, err := api.New(config.Config{Host: "127.0.0.1", DSN: "unused"}, fs, fs, fs, fs, fs, nil, api.WithPolicy(&fakePolicy{}, table), api.WithRoleCosts(rows))
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/api/v1/stats/role-costs?from=2026-08-01T00:00:00Z&to=2026-09-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var got struct {
		Roles []struct {
			Role, Group string
			CostUSD     float64
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got.Roles) != 1 || got.Roles[0].Role != "no-progress-fixer" || got.Roles[0].Group != "Implementation" || got.Roles[0].CostUSD != 4 {
		t.Fatalf("roles = %+v, want the row under no-progress-fixer (Implementation)", got.Roles)
	}
}
