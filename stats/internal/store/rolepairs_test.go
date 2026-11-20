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
// into the store, every list in table order -- as later migrations left it:
// 0085 dropped the marked escalation pairs for the no-progress-fixer role.
func TestRolePairsSeedMatchesTable(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	got, err := st.RolePairs(context.Background())
	if err != nil {
		t.Fatalf("RolePairs: %v", err)
	}
	want := map[string]store.RolePolicy{
		"explorer":              {Base: []string{"opus/Explore@low", "opus/Explore@medium"}, Optin: []string{}},
		"plan-designer":         {Base: []string{"opus/Plan@low", "opus/Plan@medium", "opus/Plan@high", "opus/Plan@xhigh", "fable/Plan@xhigh", "fable/Plan@high", "fable/Plan@medium"}, Optin: []string{}},
		"planner":               {Base: []string{"opus/flow-planner-medium", "opus/flow-planner-low", "opus/flow-planner-high", "opus/flow-planner-xhigh", "fable/flow-planner-xhigh", "fable/flow-planner-high", "fable/flow-planner-medium"}, Optin: []string{}},
		"implementer":           {Base: []string{"opus/flow-low", "opus/flow-medium", "sonnet/flow-low", "sonnet/flow-medium"}, Optin: []string{}},
		"panel-fix":             {Base: []string{"opus/flow-low", "opus/flow-medium", "sonnet/flow-low", "sonnet/flow-medium"}, Optin: []string{}},
		"minor-fix":             {Base: []string{"opus/flow-low", "opus/flow-medium", "sonnet/flow-low", "sonnet/flow-medium"}, Optin: []string{}},
		"operator-asked":        {Base: []string{"opus/flow-medium", "opus/flow-high", "sonnet/flow-medium"}, Optin: []string{}},
		"decider":               {Base: []string{"opus/flow-low"}, Optin: []string{}},
		"gated-reviewer":        {Base: []string{"opus/flow-medium"}, Optin: []string{}},
		"gated-re-reviewer":     {Base: []string{"opus/flow-low"}, Optin: []string{}},
		"panel-reviewer":        {Base: []string{"opus/flow-low", "opus/flow-medium", "fable/flow-low", "fable/flow-medium"}, Optin: []string{}},
		"panel-rerun":           {Base: []string{"opus/flow-low", "sonnet/flow-low"}, Optin: []string{}},
		"pair-judge":            {Base: []string{"sonnet/flow-low"}, Optin: []string{}},
		"panel-merger":          {Base: []string{"sonnet/flow-low"}, Optin: []string{}},
		"visual-verifier":       {Base: []string{"opus/flow-low"}, Optin: []string{}},
		"runner":                {Base: []string{"sonnet/flow-low"}, Optin: []string{}},
		"self-review-pass":      {Base: []string{"opus/flow-medium"}, Optin: []string{}},
		"self-review-plan":      {Base: []string{"opus/flow-planner-medium", "opus/flow-planner-low"}, Optin: []string{}},
		"plan-triage":           {Base: []string{"opus/flow-low"}, Optin: []string{}},
		"planner-short":         {Base: []string{"opus/flow-planner-medium", "opus/flow-planner-low", "opus/flow-planner-high", "opus/flow-planner-xhigh", "fable/flow-planner-xhigh", "fable/flow-planner-high", "fable/flow-planner-medium"}, Optin: []string{}},
		"self-review-fix":       {Base: []string{"opus/flow-medium", "sonnet/flow-medium"}, Optin: []string{}},
		"self-review-review-1":  {Base: []string{"opus/flow-medium", "opus/flow-low"}, Optin: []string{}},
		"self-review-review-r":  {Base: []string{"sonnet/flow-low"}, Optin: []string{}},
		"pipeline-fix-plan":     {Base: []string{"opus/flow-planner-medium", "opus/flow-planner-low"}, Optin: []string{}},
		"pipeline-fix-fix":      {Base: []string{"opus/flow-low"}, Optin: []string{}},
		"pipeline-fix-review-1": {Base: []string{"opus/flow-medium", "opus/flow-low"}, Optin: []string{}},
		"pipeline-fix-review-r": {Base: []string{"sonnet/flow-low"}, Optin: []string{}},
		"no-progress-fixer":     {Base: []string{"opus/flow-high"}, Optin: []string{}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("seed = %#v\nwant %#v", got, want)
	}
}

// TestReplaceRolePairsRoundTrip: a replace writes base and opt-in whole, in
// the order given.
func TestReplaceRolePairsRoundTrip(t *testing.T) {
	t.Parallel()
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
		Base:  []string{"sonnet/flow-medium", "opus/flow-low"},
		Optin: []string{"fable/flow-high"},
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
	t.Parallel()
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
	t.Parallel()
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
