package testcost

import (
	"bytes"
	"os"
	"reflect"
	"testing"
)

func TestFoldCostsTopLevelTests(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile("testdata/go-test.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	var f Fold
	for _, line := range bytes.Split(b, []byte("\n")) {
		if ev, ok := ParseEvent(line); ok {
			f.Add(ev)
		}
	}
	// TestSkipped has no row; TestPar adds its paused descendants at every depth; TestParallel
	// is a sibling whose name extends TestPar's, never its descendant; TestSeq's sequential
	// subtest adds nothing; equal costs order by package then test.
	want := []Row{
		{"ex/a", "TestPar", 2.25},
		{"ex/b", "TestFails", 2.25},
		{"ex/a", "TestParallel", 1},
		{"ex/a", "TestSeq", 0.5},
	}
	if got := f.Rows(); !reflect.DeepEqual(got, want) {
		t.Errorf("Rows() = %v, want %v", got, want)
	}
}

func TestParseEventDetectsGoTestJSON(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		line   string
		output string
	}{
		{`{"Action":"output","Package":"p","Test":"T","Output":"x\n"}`, "x\n"},
		{`{"Action":"build-output","ImportPath":"p","Output":"y\n"}`, "y\n"},
	} {
		ev, ok := ParseEvent([]byte(c.line))
		if !ok || ev.Output != c.output {
			t.Errorf("ParseEvent(%s) = %+v, %v; want an event with Output %q", c.line, ev, ok, c.output)
		}
	}
	for _, line := range []string{"FAIL\tex/b [build failed]", `{"Test":"T"}`, `{"Action":""}`, `[1]`, `{"x":1}`, ""} {
		if ev, ok := ParseEvent([]byte(line)); ok {
			t.Errorf("ParseEvent(%q) = %+v, true; want not an event", line, ev)
		}
	}
}
