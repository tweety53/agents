package store_test

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestExperimentOpenListClose walks one combined experiment through its
// life: open, listed as open, its cells served as overrides, closed with a
// verdict and summary, and gone from the overrides once closed.
func TestExperimentOpenListClose(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	cells := map[string][]string{
		"plan-designer": {"opus/Plan@xhigh"},
		"explorer":      {"sonnet/Explore@low"},
	}
	opened, err := st.OpenExperiment(ctx, store.Experiment{Name: "t", Cells: cells})
	if err != nil {
		t.Fatalf("OpenExperiment: %v", err)
	}
	if opened.ID == 0 || opened.StartedAt.IsZero() {
		t.Fatalf("OpenExperiment = %+v, want an id and a start time", opened)
	}

	list, err := st.ListExperiments(ctx, false)
	if err != nil {
		t.Fatalf("ListExperiments: %v", err)
	}
	if len(list) != 1 || list[0].ID != opened.ID || list[0].Status != "open" || list[0].Name != "t" || !reflect.DeepEqual(list[0].Cells, cells) {
		t.Fatalf("ListExperiments = %+v, want the one open experiment %d", list, opened.ID)
	}

	overrides, err := st.OpenOverrides(ctx)
	if err != nil {
		t.Fatalf("OpenOverrides: %v", err)
	}
	if !reflect.DeepEqual(overrides, cells) {
		t.Fatalf("OpenOverrides = %v, want %v", overrides, cells)
	}

	if err := st.CloseExperiment(ctx, opened.ID, "negative", "s"); err != nil {
		t.Fatalf("CloseExperiment: %v", err)
	}
	list, err = st.ListExperiments(ctx, false)
	if err != nil {
		t.Fatalf("ListExperiments after close: %v", err)
	}
	got := list[0]
	if got.ClosedAt == nil || got.Verdict != "negative" || got.Summary != "s" || got.Status != "closed" {
		t.Fatalf("closed experiment = %+v, want closedAt set, verdict negative, summary s, status closed", got)
	}
	if open, err := st.ListExperiments(ctx, true); err != nil || len(open) != 0 {
		t.Fatalf("ListExperiments(open) after close = %v, %v; want none", open, err)
	}
	overrides, err = st.OpenOverrides(ctx)
	if err != nil {
		t.Fatalf("OpenOverrides after close: %v", err)
	}
	if len(overrides) != 0 {
		t.Fatalf("OpenOverrides after close = %v, want empty", overrides)
	}

	if err := st.CloseExperiment(ctx, opened.ID, "negative", "again"); !errors.Is(err, store.ErrExperimentNotOpen) {
		t.Fatalf("second CloseExperiment = %v, want ErrExperimentNotOpen", err)
	}
	if err := st.CloseExperiment(ctx, opened.ID+1000, "negative", "s"); !errors.Is(err, store.ErrExperimentNotOpen) {
		t.Fatalf("CloseExperiment(unknown) = %v, want ErrExperimentNotOpen", err)
	}
}

// TestExperimentOpenRefusesOverlappingRole asserts one open experiment per
// role: a second open naming a role the first holds is refused, naming the
// role and the holder, writes nothing, and succeeds once the first closes.
func TestExperimentOpenRefusesOverlappingRole(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	first, err := st.OpenExperiment(ctx, store.Experiment{Name: "a", Cells: map[string][]string{
		"plan-designer": {"opus/Plan@xhigh"},
		"explorer":      {"sonnet/Explore@low"},
	}})
	if err != nil {
		t.Fatalf("first OpenExperiment: %v", err)
	}

	second := store.Experiment{Name: "b", Cells: map[string][]string{
		"implementer": {"opus/flow-medium"},
		"explorer":    {"opus/Explore@medium"},
	}}
	_, err = st.OpenExperiment(ctx, second)
	if !errors.Is(err, store.ErrExperimentConflict) {
		t.Fatalf("overlapping OpenExperiment = %v, want ErrExperimentConflict", err)
	}
	if msg := err.Error(); !strings.Contains(msg, "explorer") || !strings.Contains(msg, strconv.FormatInt(first.ID, 10)) {
		t.Fatalf("conflict message %q does not name role explorer and experiment %d", msg, first.ID)
	}
	if list, _ := st.ListExperiments(ctx, false); len(list) != 1 {
		t.Fatalf("refused open left %d experiments, want 1", len(list))
	}
	if overrides, _ := st.OpenOverrides(ctx); overrides["implementer"] != nil {
		t.Fatalf("refused open left an implementer override: %v", overrides)
	}

	if err := st.CloseExperiment(ctx, first.ID, "inconclusive", ""); err != nil {
		t.Fatalf("CloseExperiment: %v", err)
	}
	if _, err := st.OpenExperiment(ctx, second); err != nil {
		t.Fatalf("OpenExperiment after the holder closed: %v", err)
	}
}

// TestValidateExperiment asserts every malformed experiment and close
// verdict is refused before anything is written.
func TestValidateExperiment(t *testing.T) {
	cases := map[string]store.Experiment{
		"empty name":       {Name: "", Cells: map[string][]string{"explorer": {"opus/Explore@low"}}},
		"no cells":         {Name: "t"},
		"cell no pairs":    {Name: "t", Cells: map[string][]string{"explorer": {}}},
		"pair bad shape":   {Name: "t", Cells: map[string][]string{"explorer": {"opus-Explore"}}},
		"pair two efforts": {Name: "t", Cells: map[string][]string{"explorer": {"opus/Explore@low@high"}}},
		"main bad shape":   {Name: "t", Cells: map[string][]string{"main": {"opus/flow-low"}}},
		"main two pairs":   {Name: "t", Cells: map[string][]string{"main": {"opus@low", "opus@high"}}},
		"empty role":       {Name: "t", Cells: map[string][]string{"": {"opus/flow-low"}}},
	}
	for name, e := range cases {
		if err := store.ValidateExperiment(e); err == nil {
			t.Errorf("%s: ValidateExperiment(%+v) = nil, want an error", name, e)
		}
	}
	good := store.Experiment{Name: "t", Cells: map[string][]string{
		"main":     {"opus@xhigh"},
		"explorer": {"sonnet/Explore@low", "opus/Explore@medium"},
		"planner":  {"opus/flow-planner-xhigh"},
	}}
	if err := store.ValidateExperiment(good); err != nil {
		t.Errorf("ValidateExperiment(good) = %v, want nil", err)
	}

	st := newTestStore(t)
	ctx := context.Background()
	if _, err := st.OpenExperiment(ctx, cases["no cells"]); err == nil {
		t.Errorf("OpenExperiment(no cells) = nil, want an error")
	}
	opened, err := st.OpenExperiment(ctx, good)
	if err != nil {
		t.Fatalf("OpenExperiment(good): %v", err)
	}
	if err := st.CloseExperiment(ctx, opened.ID, "great", "s"); !errors.Is(err, store.ErrInvalidVerdict) {
		t.Errorf("CloseExperiment(verdict great) = %v, want ErrInvalidVerdict", err)
	}
}
