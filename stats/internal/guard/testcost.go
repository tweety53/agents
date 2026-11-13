package guard

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// flow-guard test-cost — the per-test cost of a `go test -json` run.
//
// Usage: flow-guard test-cost [-min <seconds>]
//
// Reads `go test -json` output on stdin and prints, for every top-level test (no `/` in its
// name), what that test costs the run: its own Elapsed plus the Elapsed of every descendant, at
// any depth, that paused — a parallel subtest runs after its parent reported, so its time is not
// in the parent's Elapsed. A sequential subtest adds nothing: its time is already the parent's.
// A test's Elapsed is the one on its pass, fail or skip event. A top-level test whose terminal
// event is skip, or that has none, prints no row; an event with no Test, and a line that is not
// JSON, are ignored.
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
			return 2
		}
		min = v
	default:
		fmt.Fprintln(stderr, "usage: flow-guard test-cost [-min <seconds>]")
		return 2
	}

	type key struct{ pkg, test string }
	type state struct {
		elapsed  float64
		paused   bool
		terminal string
	}
	tests := map[key]*state{}
	get := func(k key) *state {
		s := tests[k]
		if s == nil {
			s = &state{}
			tests[k] = s
		}
		return s
	}
	if env.Stdin != nil {
		r := bufio.NewReader(env.Stdin)
		for {
			line, err := r.ReadString('\n')
			var ev struct {
				Action, Package, Test string
				Elapsed               float64
			}
			if json.Unmarshal([]byte(line), &ev) == nil && ev.Test != "" {
				s := get(key{ev.Package, ev.Test})
				switch ev.Action {
				case "pause":
					s.paused = true
				case "pass", "fail", "skip":
					s.elapsed, s.terminal = ev.Elapsed, ev.Action
				}
			}
			if err != nil {
				if !errors.Is(err, io.EOF) {
					fmt.Fprintf(stderr, "test-cost: reading stdin failed: %v\n", err)
				}
				break
			}
		}
	}

	type row struct {
		cost float64
		key
	}
	var rows []row
	for k, s := range tests {
		if strings.Contains(k.test, "/") || s.terminal == "" || s.terminal == "skip" {
			continue
		}
		cost := s.elapsed
		for d, ds := range tests {
			if ds.paused && d.pkg == k.pkg && strings.HasPrefix(d.test, k.test+"/") {
				cost += ds.elapsed
			}
		}
		if cost >= min {
			rows = append(rows, row{cost, k})
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.cost != b.cost {
			return a.cost > b.cost
		}
		if a.pkg != b.pkg {
			return a.pkg < b.pkg
		}
		return a.test < b.test
	})
	for _, r := range rows {
		fmt.Fprintf(stdout, "%.2f\t%s\t%s\n", r.cost, r.pkg, r.test)
	}
	return 0
}
