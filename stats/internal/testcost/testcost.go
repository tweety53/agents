// Package testcost folds a `go test -json` (test2json) stream into the cost of each top-level
// test — the one rule `flow-guard test-cost` prints and `flow suite record` stores (decision
// testcost-shared-package).
//
// A top-level test (no `/` in its name) costs its own Elapsed plus the Elapsed of every
// descendant, at any depth, that paused — a parallel subtest runs after its parent reported, so
// its time is not in the parent's Elapsed. A sequential subtest adds nothing: its time is already
// the parent's. A test's Elapsed is the one on its pass, fail or skip event. A top-level test
// whose terminal event is skip, or that has none, has no row; an event with no Test is ignored.
// Rows are heaviest first, equal costs ordered by package then test.
package testcost

import (
	"encoding/json"
	"sort"
	"strings"
)

// Event is one test2json event, reduced to the fields the fold and the log printer read.
type Event struct {
	Action, Package, Test, Output string
	Elapsed                       float64
}

// ParseEvent decodes one line of a test2json stream: a JSON object with a non-empty Action.
// Any other line — plain text, a JSON array, an object without Action — is not an event.
func ParseEvent(line []byte) (Event, bool) {
	var ev Event
	if json.Unmarshal(line, &ev) != nil || ev.Action == "" {
		return Event{}, false
	}
	return ev, true
}

// Row is one top-level test's cost.
type Row struct {
	Package, Test string
	Cost          float64 // seconds
}

type key struct{ pkg, test string }

type state struct {
	elapsed  float64
	paused   bool
	terminal string
}

// Fold accumulates events; its zero value is ready to use.
type Fold struct {
	tests map[key]*state
}

// Add folds one event in; an event with no Test is ignored.
func (f *Fold) Add(ev Event) {
	if ev.Test == "" {
		return
	}
	if f.tests == nil {
		f.tests = map[key]*state{}
	}
	k := key{ev.Package, ev.Test}
	s := f.tests[k]
	if s == nil {
		s = &state{}
		f.tests[k] = s
	}
	switch ev.Action {
	case "pause":
		s.paused = true
	case "pass", "fail", "skip":
		s.elapsed, s.terminal = ev.Elapsed, ev.Action
	}
}

// Rows returns the top-level tests' costs, heaviest first.
func (f *Fold) Rows() []Row {
	paused := map[key]float64{} // paused descendants' elapsed, by top-level test
	for k, s := range f.tests {
		if top, _, sub := strings.Cut(k.test, "/"); sub && s.paused {
			paused[key{k.pkg, top}] += s.elapsed
		}
	}
	var rows []Row
	for k, s := range f.tests {
		if strings.Contains(k.test, "/") || s.terminal == "" || s.terminal == "skip" {
			continue
		}
		rows = append(rows, Row{k.pkg, k.test, s.elapsed + paused[k]})
	}
	sort.Slice(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.Cost != b.Cost {
			return a.Cost > b.Cost
		}
		if a.Package != b.Package {
			return a.Package < b.Package
		}
		return a.Test < b.Test
	})
	return rows
}
