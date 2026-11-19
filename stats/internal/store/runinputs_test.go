package store_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// TestRunInputsReadsEachChangeInRefOrder pins RunInputs: every ref gets its
// own change's dispatches, findings, stage runs, verify runs and saved
// tasks.md/reports.* files, in ref order and never another change's, and an
// unknown ref fails the whole read with ErrChangeNotFound.
func TestRunInputsReadsEachChangeInRefOrder(t *testing.T) {
	t.Parallel()
	st, _ := newRecordStore(t)
	ctx := context.Background()
	p := fmt.Sprintf("proj-run-inputs-%d", time.Now().UnixNano())
	at := time.Date(2026, 11, 18, 9, 0, 0, 0, time.UTC)
	for i, name := range []string{"kan-1", "kan-2"} {
		seedChange(t, st, p, name)
		for j := 0; j <= i; j++ {
			if _, err := st.RecordDispatch(ctx, p, name, autoDispatch(fmt.Sprintf("ri-%s-%d", name, j), fmt.Sprintf("task-%d-implementer", j+1), at)); err != nil {
				t.Fatalf("RecordDispatch: %v", err)
			}
		}
	}
	if _, _, err := st.UpsertFinding(ctx, p, "kan-2", baseFinding("F1", 1)); err != nil {
		t.Fatalf("UpsertFinding: %v", err)
	}
	if _, err := st.RecordVerifyRun(ctx, p, "kan-1", records.VerifyRun{Kind: "run-verify-list", Label: "test", Seconds: 3, RecordedAt: at}); err != nil {
		t.Fatalf("RecordVerifyRun: %v", err)
	}
	for _, path := range []string{"tasks.md", "reports.panel.md", "proposal.md"} {
		if err := st.SaveChangeFile(ctx, p, "kan-2", path, path); err != nil {
			t.Fatalf("SaveChangeFile: %v", err)
		}
	}

	got, err := st.RunInputs(ctx, []store.ChangeRef{{ProjectKey: p, Name: "kan-2"}, {ProjectKey: p, Name: "kan-1"}})
	if err != nil {
		t.Fatalf("RunInputs: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	k2, k1 := got[0], got[1]
	if len(k2.Dispatches) != 2 || len(k1.Dispatches) != 1 {
		t.Errorf("dispatches = %d/%d, want 2/1", len(k2.Dispatches), len(k1.Dispatches))
	}
	if len(k2.Stages) == 0 || len(k1.Stages) == 0 {
		t.Errorf("stages = %d/%d, want each change's auto-opened runs", len(k2.Stages), len(k1.Stages))
	}
	if len(k2.Findings) != 1 || len(k1.Findings) != 0 {
		t.Errorf("findings = %d/%d, want 1/0", len(k2.Findings), len(k1.Findings))
	}
	if len(k2.VerifyRuns) != 0 || len(k1.VerifyRuns) != 1 {
		t.Errorf("verify runs = %d/%d, want 0/1", len(k2.VerifyRuns), len(k1.VerifyRuns))
	}
	if len(k2.Files) != 2 || k2.Files["tasks.md"] != "tasks.md" || k2.Files["reports.panel.md"] == "" || len(k1.Files) != 0 {
		t.Errorf("files = %v / %v, want kan-2's tasks.md and reports.panel.md only", k2.Files, k1.Files)
	}

	if _, err := st.RunInputs(ctx, []store.ChangeRef{{ProjectKey: p, Name: "kan-1"}, {ProjectKey: p, Name: "nope"}}); !errors.Is(err, store.ErrChangeNotFound) {
		t.Errorf("unknown ref: err = %v, want ErrChangeNotFound", err)
	}
}
