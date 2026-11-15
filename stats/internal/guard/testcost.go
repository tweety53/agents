package guard

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/tweety53/agents/stats/internal/testcost"
)

// flow-guard test-cost — the per-test cost of a `go test -json` run.
//
// Usage: flow-guard test-cost [-min <seconds>]
//
// Reads `go test -json` output on stdin and prints, for every top-level test (no `/` in its
// name), what that test costs the run, by the rule package internal/testcost states: its own
// Elapsed plus the Elapsed of every descendant, at any depth, that paused. A top-level test whose
// terminal event is skip, or that has none, prints no row; an event with no Test, and a line that
// is not a test2json event, are ignored.
//
// Each row is `<cost to two decimals>` tab `<package>` tab `<test>`, heaviest first, equal costs
// ordered by package then test. -min prints only the rows whose cost is at least <seconds>. The
// guard reports costs and never judges one.
//
// Exit codes:
//
//	0  the table was printed (an empty one included)
//	2  usage
func init() { Registry["test-cost"] = testCost }

// testCost is `flow-guard test-cost`: the comment above is the contract.
func testCost(args []string, env Env, stdout, stderr io.Writer) int {
	min := 0.0
	switch {
	case len(args) == 0:
	case len(args) == 2 && args[0] == "-min":
		v, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			fmt.Fprintln(stderr, "usage: flow-guard test-cost [-min <seconds>]")
			Reason(env, "usage")
			return 2
		}
		min = v
	default:
		fmt.Fprintln(stderr, "usage: flow-guard test-cost [-min <seconds>]")
		Reason(env, "usage")
		return 2
	}

	var fold testcost.Fold
	if env.Stdin != nil {
		r := bufio.NewReader(env.Stdin)
		for {
			line, err := r.ReadBytes('\n')
			if ev, ok := testcost.ParseEvent(line); ok {
				fold.Add(ev)
			}
			if err != nil {
				if !errors.Is(err, io.EOF) {
					fmt.Fprintf(stderr, "test-cost: reading stdin failed: %v\n", err)
				}
				break
			}
		}
	}
	for _, r := range fold.Rows() {
		if r.Cost >= min {
			fmt.Fprintf(stdout, "%.2f\t%s\t%s\n", r.Cost, r.Package, r.Test)
		}
	}
	return 0
}
