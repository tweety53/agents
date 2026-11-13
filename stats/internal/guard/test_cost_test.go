package guard

import (
	"strings"
	"testing"
)

const testCostFixture = `{"Action":"run","Package":"ex/a","Test":"TestSeq"}
{"Action":"run","Package":"ex/a","Test":"TestSeq/one"}
{"Action":"pass","Package":"ex/a","Test":"TestSeq/one","Elapsed":0.25}
{"Action":"pass","Package":"ex/a","Test":"TestSeq","Elapsed":0.5}
{"Action":"run","Package":"ex/a","Test":"TestPar"}
{"Action":"run","Package":"ex/a","Test":"TestPar/x"}
{"Action":"pause","Package":"ex/a","Test":"TestPar/x"}
{"Action":"run","Package":"ex/a","Test":"TestPar/y"}
{"Action":"pause","Package":"ex/a","Test":"TestPar/y"}
{"Action":"pass","Package":"ex/a","Test":"TestPar","Elapsed":0.25}
{"Action":"cont","Package":"ex/a","Test":"TestPar/x"}
{"Action":"cont","Package":"ex/a","Test":"TestPar/y"}
{"Action":"run","Package":"ex/a","Test":"TestPar/y/deep"}
{"Action":"pause","Package":"ex/a","Test":"TestPar/y/deep"}
{"Action":"cont","Package":"ex/a","Test":"TestPar/y/deep"}
{"Action":"pass","Package":"ex/a","Test":"TestPar/y/deep","Elapsed":0.25}
{"Action":"pass","Package":"ex/a","Test":"TestPar/x","Elapsed":1.25}
{"Action":"pass","Package":"ex/a","Test":"TestPar/y","Elapsed":0.5}
{"Action":"run","Package":"ex/a","Test":"TestParallel"}
{"Action":"pause","Package":"ex/a","Test":"TestParallel"}
{"Action":"cont","Package":"ex/a","Test":"TestParallel"}
{"Action":"pass","Package":"ex/a","Test":"TestParallel","Elapsed":1}
{"Action":"run","Package":"ex/b","Test":"TestSkipped"}
{"Action":"skip","Package":"ex/b","Test":"TestSkipped","Elapsed":0}
{"Action":"run","Package":"ex/b","Test":"TestFails"}
{"Action":"output","Package":"ex/b","Test":"TestFails","Output":"--- FAIL: TestFails (2.25s)\n"}
{"Action":"fail","Package":"ex/b","Test":"TestFails","Elapsed":2.25}
FAIL	ex/b [build failed]
{"Action":"fail","Package":"ex/b","Elapsed":2.5}
{"Action":"pass","Package":"ex/a","Elapsed":3}
`

func TestTestCostAddsPausedDescendants(t *testing.T) {
	t.Parallel()
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
		{"no argument", nil, testCostFixture, par + fails + sib + seq},
		{"-min 2.25", []string{"-min", "2.25"}, testCostFixture, par + fails},
		{"-min 2.26", []string{"-min", "2.26"}, testCostFixture, ""},
		{"empty stdin", nil, "", ""},
	} {
		r := runGuard("test-cost", c.args, Env{Stdin: strings.NewReader(c.input)})
		if r.rc != 0 || r.err != "" || r.stdout != c.want {
			t.Errorf("%s: rc=%d stderr=%q stdout=%q, want 0, empty stderr and %q", c.name, r.rc, r.err, r.stdout, c.want)
		}
	}
}

func TestTestCostUsage(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{{"-min"}, {"-min", "abc"}, {"x"}, {"-min", "1", "x"}} {
		r := runGuard("test-cost", args, Env{Stdin: strings.NewReader(testCostFixture)})
		if r.rc != 2 || r.err != "usage: flow-guard test-cost [-min <seconds>]\n" || r.stdout != "" {
			t.Errorf("%q: rc=%d stderr=%q stdout=%q, want 2 and the usage line", args, r.rc, r.err, r.stdout)
		}
	}
}
