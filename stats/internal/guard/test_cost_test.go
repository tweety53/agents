package guard

import (
	"os"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// testCostFixture reads the go test -json stream the shared rule's tests read too.
func testCostFixture(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("../testcost/testdata/go-test.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestTestCostAddsPausedDescendants(t *testing.T) {
	t.Parallel()
	fixture := testCostFixture(t)
	const (
		par   = "2.25\tex/a\tTestPar\n"
		fails = "2.25\tex/b\tTestFails\n"
		seq   = "0.50\tex/a\tTestSeq\n"
		// TestParallel's name extends TestPar's: a paused sibling, never TestPar's descendant.
		sib = "1.00\tex/a\tTestParallel\n"
	)
	for _, c := range []struct {
		name  string
		args  []string
		input string
		want  string
	}{
		{"no argument", nil, fixture, par + fails + sib + seq},
		{"-min 2.25", []string{"-min", "2.25"}, fixture, par + fails},
		{"-min 2.26", []string{"-min", "2.26"}, fixture, ""},
		{"empty stdin", nil, "", ""},
	} {
		r := runGuard("test-cost", c.args, core.Env{Stdin: strings.NewReader(c.input)})
		if r.Rc != 0 || r.Err != "" || r.Stdout != c.want {
			t.Errorf("%s: rc=%d stderr=%q stdout=%q, want 0, empty stderr and %q", c.name, r.Rc, r.Err, r.Stdout, c.want)
		}
	}
}

func TestTestCostUsage(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{{"-min"}, {"-min", "abc"}, {"x"}, {"-min", "1", "x"}} {
		r := runGuard("test-cost", args, core.Env{Stdin: strings.NewReader(testCostFixture(t))})
		if r.Rc != 2 || r.Err != "usage: flow-guard test-cost [-min <seconds>]\n" || r.Stdout != "" {
			t.Errorf("%q: rc=%d stderr=%q stdout=%q, want 2 and the usage line", args, r.Rc, r.Err, r.Stdout)
		}
	}
}
