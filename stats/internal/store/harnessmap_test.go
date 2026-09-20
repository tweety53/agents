package store_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestRecordDispatchValidatesHarnessPair pins KAN-610: a dispatch row the
// harness mapping cannot vouch for is refused at write time, so the ledger
// never records a model the dispatch could not have run (KAN-553's round-0
// rows recorded the decision's sonnet/high where the zcode mapping calls
// for glm-5.3-flash/high, and the store accepted them without comment).
//
// The mapping speaks only where it can: a session token with no stage run,
// or none at all, carries no harness, so any pair on it passes -- the
// refusal must never fire on a value the mapping does not know.
func TestRecordDispatchValidatesHarnessPair(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-harnesspair-%d", time.Now().UnixNano())

	const (
		mapped = "glm-5.3-flash"
		effort = "high"
	)

	cases := []struct {
		name        string
		harness     string // stage run seeded for this case's token; "" seeds none
		withToken   bool   // whether the dispatch carries a session token at all
		model       string
		effort      string
		wantRefused bool
	}{
		{"mapped pair accepted", "zcode", true, mapped, effort, false},
		{"wrong model refused", "zcode", true, "sonnet", effort, true},
		{"default effort refused", "zcode", true, mapped, "default", true},
		{"empty effort refused", "zcode", true, mapped, "", true},
		{"unknown agent-defined refused", "zcode", true, "unknown (agent-defined)", effort, true},
		{"non-mapping harness accepts any pair", "claude-code", true, "sonnet", "high", false},
		{"no stage run accepts any pair", "", true, "sonnet", "high", false},
		{"no session token accepts any pair", "", false, "sonnet", "high", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			change := fmt.Sprintf("kan-%s", strings.ReplaceAll(tc.name, " ", "-"))
			token := "ff-harnesspair-" + change
			seedChange(t, st, projectKey, change)
			if tc.harness != "" {
				in := baseBeginInput(projectKey, change, "/flow", "SDD + TDD per task")
				in.Harness = tc.harness
				in.SessionID = nil
				in.SessionToken = ptr(token)
				if _, err := st.BeginStage(ctx, in); err != nil {
					t.Fatalf("seed stage run: %v", err)
				}
			}

			in := baseDispatch("implementer", tc.model)
			in.Effort = tc.effort
			in.Key = "the-dispatch"
			if tc.withToken {
				in.SessionToken = token
			}

			out, err := st.RecordDispatch(ctx, projectKey, change, in)
			if tc.wantRefused {
				if !errors.Is(err, store.ErrDispatchPairInvalid) {
					t.Fatalf("RecordDispatch(%q, %q) error = %v, want ErrDispatchPairInvalid", tc.model, tc.effort, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("RecordDispatch(%q, %q): %v", tc.model, tc.effort, err)
			}
			if out.Seq == 0 {
				t.Fatalf("accepted dispatch came back with seq 0")
			}
			if out.Model != tc.model {
				t.Errorf("stored model = %q, want %q", out.Model, tc.model)
			}
		})
	}
}
