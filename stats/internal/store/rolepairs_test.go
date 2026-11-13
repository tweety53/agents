package store_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestRolePairsSeedMatchesTable pins 0061's seed to the Allowed pairs table
// of skills/flow-contracts/model-policy.md as it stood when the pairs moved
// into the store: unmarked rows base, marked rows opt-in under their marker,
// every list in table order.
func TestRolePairsSeedMatchesTable(t *testing.T) {
	st := newTestStore(t)
	got, err := st.RolePairs(context.Background())
	if err != nil {
		t.Fatalf("RolePairs: %v", err)
	}
	want := map[string]store.RolePolicy{
		"explorer":                {Base: []string{"opus/Explore@low", "opus/Explore@medium"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"plan-designer":           {Base: []string{"opus/Plan@low", "opus/Plan@medium", "opus/Plan@high", "opus/Plan@xhigh", "fable/Plan@xhigh", "fable/Plan@high", "fable/Plan@medium"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"planner":                 {Base: []string{"opus/flow-planner-medium", "opus/flow-planner-low", "opus/flow-planner-high", "opus/flow-planner-xhigh", "fable/flow-planner-xhigh", "fable/flow-planner-high", "fable/flow-planner-medium"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"implementer":             {Base: []string{"opus/flow-low", "opus/flow-medium", "sonnet/flow-low", "sonnet/flow-medium"}, Optin: []string{}, Marked: []store.MarkedPair{{Pair: "opus/flow-high", Marker: "No-progress escalation"}}},
		"panel-fix":               {Base: []string{"opus/flow-low", "opus/flow-medium", "sonnet/flow-low", "sonnet/flow-medium"}, Optin: []string{}, Marked: []store.MarkedPair{{Pair: "opus/flow-high", Marker: "No-progress escalation"}}},
		"operator-asked":          {Base: []string{"opus/flow-medium", "opus/flow-high", "sonnet/flow-medium"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"design-reviewer":         {Base: []string{"opus/flow-low"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"gated-reviewer":          {Base: []string{"opus/flow-medium"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"gated-re-reviewer":       {Base: []string{"opus/flow-low"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"panel-reviewer":          {Base: []string{"opus/flow-low", "opus/flow-medium", "fable/flow-low", "fable/flow-medium"}, Optin: []string{}, Marked: []store.MarkedPair{{Pair: "opus/flow-high", Marker: "Operator-approved high"}}},
		"panel-sanity-efficiency": {Base: []string{"opus/flow-low", "sonnet/flow-low"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"panel-mutation-failure":  {Base: []string{"opus/flow-low", "sonnet/flow-low"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"panel-rerun":             {Base: []string{"opus/flow-low", "sonnet/flow-low"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"late-fix":                {Base: []string{"opus/flow-low"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"fix-run-judge":           {Base: []string{"opus/flow-low"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"panel-merger":            {Base: []string{"sonnet/flow-low"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"visual-verifier":         {Base: []string{"opus/flow-low"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"tooling-analyst":         {Base: []string{"opus/flow-medium"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"self-review-pass":        {Base: []string{"opus/flow-medium"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"self-review-plan":        {Base: []string{"opus/flow-planner-medium", "opus/flow-planner-low"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"self-review-fix":         {Base: []string{"opus/flow-medium", "sonnet/flow-medium"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"self-review-review-1":    {Base: []string{"opus/flow-medium", "opus/flow-low"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"self-review-review-r":    {Base: []string{"sonnet/flow-low"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"pipeline-fix-plan":       {Base: []string{"opus/flow-planner-medium", "opus/flow-planner-low"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"pipeline-fix-fix":        {Base: []string{"opus/flow-low"}, Optin: []string{}, Marked: []store.MarkedPair{{Pair: "opus/flow-high", Marker: "No-progress escalation"}}},
		"pipeline-fix-review-1":   {Base: []string{"opus/flow-medium", "opus/flow-low"}, Optin: []string{}, Marked: []store.MarkedPair{}},
		"pipeline-fix-review-r":   {Base: []string{"sonnet/flow-low"}, Optin: []string{}, Marked: []store.MarkedPair{}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("seed = %#v\nwant %#v", got, want)
	}
}

// TestReplaceRolePairsRoundTrip: a replace writes base and opt-in whole, in
// the order given, and leaves the role's marked pairs alone.
func TestReplaceRolePairsRoundTrip(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	if err := st.ReplaceRolePairs(ctx, "panel-fix", []string{"sonnet/flow-medium", "opus/flow-low"}, []string{"fable/flow-high"}); err != nil {
		t.Fatalf("ReplaceRolePairs: %v", err)
	}
	got, err := st.RolePairs(ctx)
	if err != nil {
		t.Fatalf("RolePairs: %v", err)
	}
	want := store.RolePolicy{
		Base:   []string{"sonnet/flow-medium", "opus/flow-low"},
		Optin:  []string{"fable/flow-high"},
		Marked: []store.MarkedPair{{Pair: "opus/flow-high", Marker: "No-progress escalation"}},
	}
	if !reflect.DeepEqual(got["panel-fix"], want) {
		t.Fatalf("panel-fix = %#v, want %#v", got["panel-fix"], want)
	}
	if err := st.ReplaceRolePairs(ctx, "new-role", []string{"opus/flow-low"}, nil); err != nil {
		t.Fatalf("ReplaceRolePairs new role: %v", err)
	}
	if got, _ := st.RolePairs(ctx); !reflect.DeepEqual(got["new-role"].Base, []string{"opus/flow-low"}) {
		t.Fatalf("new-role = %#v", got["new-role"])
	}
}

// TestReplaceRolePairsRefusesMalformedPair: a malformed or repeated pair
// refuses the whole write, leaving the role as it was.
func TestReplaceRolePairsRefusesMalformedPair(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	before, err := st.RolePairs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for name, c := range map[string][2][]string{
		"no slash":     {{"opus"}, nil},
		"bad effort":   {{"opus/Plan@"}, nil},
		"optin spaces": {{"opus/flow-low"}, {"opus flow-low"}},
		"twice":        {{"opus/flow-low", "opus/flow-low"}, nil},
	} {
		if err := st.ReplaceRolePairs(ctx, "planner", c[0], c[1]); !errors.Is(err, store.ErrInvalidPair) {
			t.Errorf("%s: err = %v, want ErrInvalidPair", name, err)
		}
	}
	if err := st.ReplaceRolePairs(ctx, "", []string{"opus/flow-low"}, nil); !errors.Is(err, store.ErrInvalidPair) {
		t.Errorf("empty role: err = %v, want ErrInvalidPair", err)
	}
	after, err := st.RolePairs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("a refused write changed the pairs")
	}
}

// TestGrantRevokeRoundTrip: grants are per change, idempotent, revocable,
// and lapse once the change is recorded FINISHED.
func TestGrantRevokeRoundTrip(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	if err := st.Grant(ctx, "kan-1", []string{"planner", "plan-designer"}); err != nil {
		t.Fatalf("Grant: %v", err)
	}
	if err := st.Grant(ctx, "kan-1", []string{"planner"}); err != nil {
		t.Fatalf("Grant again: %v", err)
	}
	if got, _ := st.Grants(ctx, "kan-1"); !reflect.DeepEqual(got, []string{"plan-designer", "planner"}) {
		t.Fatalf("Grants = %v", got)
	}
	if got, _ := st.Grants(ctx, "kan-2"); len(got) != 0 {
		t.Fatalf("kan-2 Grants = %v, want none", got)
	}
	if err := st.Revoke(ctx, "kan-1", []string{"planner"}); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	if err := st.Revoke(ctx, "kan-1", []string{"planner"}); !errors.Is(err, store.ErrGrantNotFound) {
		t.Fatalf("Revoke twice: err = %v, want ErrGrantNotFound", err)
	}
	if got, _ := st.Grants(ctx, "kan-1"); !reflect.DeepEqual(got, []string{"plan-designer"}) {
		t.Fatalf("Grants after revoke = %v", got)
	}
	if err := st.Grant(ctx, "", []string{"planner"}); !errors.Is(err, store.ErrInvalidPair) {
		t.Fatalf("Grant with no change: err = %v", err)
	}

	// A FINISHED change's grants lapse.
	projectKey := "proj-grants"
	seedChange(t, st, projectKey, "kan-1")
	ch := baseChange(projectKey, "kan-1")
	ch.State = store.StateFinished
	if err := st.PutChange(ctx, ch); err != nil {
		t.Fatalf("PutChange: %v", err)
	}
	if got, _ := st.Grants(ctx, "kan-1"); len(got) != 0 {
		t.Fatalf("Grants of a FINISHED change = %v, want none", got)
	}
}

// TestReplaceRolePairsRefusesOptinEqualToMarked: an opt-in pair equal to one
// of the role's marked pairs refuses the whole write rather than vanishing.
func TestReplaceRolePairsRefusesOptinEqualToMarked(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	before, err := st.RolePairs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = st.ReplaceRolePairs(ctx, "implementer", []string{"opus/flow-low"}, []string{"opus/flow-high"})
	if !errors.Is(err, store.ErrInvalidPair) {
		t.Fatalf("replace = %v, want ErrInvalidPair", err)
	}
	after, err := st.RolePairs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before["implementer"], after["implementer"]) {
		t.Fatalf("refused write changed implementer: %+v -> %+v", before["implementer"], after["implementer"])
	}
}
