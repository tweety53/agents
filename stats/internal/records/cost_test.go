package records_test

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
)

func costDispatch(seq int, role, task string, bag string, cost *float64) records.Dispatch {
	var metrics json.RawMessage
	if bag != "" {
		metrics = json.RawMessage(bag)
	}
	return records.Dispatch{
		Seq: seq, Role: role, TaskID: task, Model: "opus",
		StartedAt: time.Date(2026, 1, 2, 3, 4, 0, 0, time.UTC),
		Metrics:   metrics, CostUSD: cost,
	}
}

func costPtr(v float64) *float64 { return &v }

// tokensBag builds the tokens shape the harvester's second attribution pass
// merges into a dispatch row (render.go's tokenTotals encoding).
func tokensBag(main, side map[string]int64) string {
	bucket := func(m map[string]int64) string {
		return `{"input":` + itoa(m["input"]) + `,"output":` + itoa(m["output"]) +
			`,"cache_read":` + itoa(m["cache_read"]) + `,"cache_creation":` + itoa(m["cache_creation"]) + `}`
	}
	return `{"tokens":{"main":` + bucket(main) + `,"sidechain":` + bucket(side) + `}}`
}

func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
}

// TestCostByRoleTaskGroupsAndSums pins the grouping and the arithmetic:
// rows sharing (role, task) sum their main+sidechain token fields and their
// present CostUSD figures; groups come out in the rows' seq order of first
// appearance.
func TestCostByRoleTaskGroupsAndSums(t *testing.T) {
	run := records.Run{
		Change: "demo",
		Dispatches: []records.Dispatch{
			costDispatch(1, "implementer", "1", tokensBag(
				map[string]int64{"input": 100, "output": 20, "cache_read": 5, "cache_creation": 3},
				map[string]int64{"input": 7, "output": 1}), costPtr(1.25)),
			costDispatch(2, "reviewer", "1", tokensBag(
				map[string]int64{"input": 10}, nil), nil),
			costDispatch(3, "implementer", "1", tokensBag(
				map[string]int64{"input": 1}, nil), costPtr(0.75)),
		},
	}

	got := records.CostByRoleTask(run)
	want := []records.RoleTaskCost{
		{Role: "implementer", Task: "1", Dispatches: 2, Input: 108, Output: 21, CacheRead: 5, CacheCreation: 3, CostUSD: 2.00},
		{Role: "reviewer", Task: "1", Dispatches: 1, Input: 10, CostUSD: 0},
	}
	if !reflect.DeepEqual(got.Groups, want) {
		t.Errorf("CostByRoleTask groups =\n%+v\nwant\n%+v", got.Groups, want)
	}
}

// TestCostByRoleTaskAbsenceIsNotZero pins the absence counts: a dispatch
// with no tokens key is unmeasured, one with nil CostUSD is unpriced, and
// neither contributes zero to a sum that would read as a measurement.
func TestCostByRoleTaskAbsenceIsNotZero(t *testing.T) {
	run := records.Run{
		Change: "demo",
		Dispatches: []records.Dispatch{
			costDispatch(1, "implementer", "1", "", nil),                                           // nothing measured, nothing priced
			costDispatch(2, "implementer", "1", tokensBag(map[string]int64{"input": 5}, nil), nil), // measured, unpriced
			costDispatch(3, "implementer", "1", "", costPtr(2.50)),                                 // priced, unmeasured
		},
	}

	got := records.CostByRoleTask(run)
	if got.Unmeasured != 2 {
		t.Errorf("Unmeasured = %d, want 2", got.Unmeasured)
	}
	if got.Unpriced != 2 {
		t.Errorf("Unpriced = %d, want 2", got.Unpriced)
	}
	if len(got.Groups) != 1 || got.Groups[0].Dispatches != 3 {
		t.Fatalf("groups = %+v, want one group of 3", got.Groups)
	}
	g := got.Groups[0]
	if g.Input != 5 {
		t.Errorf("Input = %d, want 5 (the only measured row's figure)", g.Input)
	}
	if g.CostUSD != 2.50 {
		t.Errorf("CostUSD = %v, want 2.50 (the only priced row's figure)", g.CostUSD)
	}
}
