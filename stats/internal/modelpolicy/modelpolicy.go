// Package modelpolicy parses the **Allowed pairs** table of
// skills/flow-contracts/model-policy.md -- the one parser the guards
// (internal/guard) and flowd share (KAN-1067's roles-table-embedded
// decision).
package modelpolicy

import (
	"fmt"
	"regexp"
	"strings"
)

// File is the one canonical place the pairs are read from, relative to the
// agents repository root.
const File = "skills/flow-contracts/model-policy.md"

// Row is one table row: the pairs it admits, and the whole prompt line it
// requires ("" for none).
type Row struct {
	Pairs  []string
	Marker string
}

var tick = regexp.MustCompile("`([^`]+)`")

// Entry is one table row with its role id, its group -- the nearest `###`
// heading above its table -- and its Does line ("" on a marked row), as Rows
// returns it.
type Entry struct {
	Role, Group, Does string
	Row
}

// tableHeader opens every allowed-pairs table.
const tableHeader = "| Role | Allowed pairs | Requires marker | Does |"

// Parse reads the `| Role | Allowed pairs | Requires marker | Does |` tables
// out of b, keyed by role id with each role's rows in table order. A file
// with no such table, a row of the wrong width, an unmarked row with no Does
// line, or a role under two groups is an error: a dispatch is never admitted
// off a table that cannot be read.
func Parse(b []byte) (map[string][]Row, error) {
	entries, err := Rows(b)
	if err != nil {
		return nil, err
	}
	rows := map[string][]Row{}
	for _, e := range entries {
		rows[e.Role] = append(rows[e.Role], e.Row)
	}
	return rows, nil
}

// Rows is Parse's tables as one entry per row in table order -- the order
// flowd's roles table is drawn in, which Parse's map does not keep. Every
// table is read until the first `## ` heading after a row.
func Rows(b []byte) ([]Entry, error) {
	var out []Entry
	group, in := "", false
	groupOf := map[string]string{}
	for _, line := range strings.Split(string(b), "\n") {
		t := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(t, "### "):
			group, in = strings.TrimSpace(t[4:]), false
			continue
		case strings.HasPrefix(t, "## "):
			if len(out) > 0 {
				return out, nil
			}
			continue
		case strings.HasPrefix(t, tableHeader):
			in = true
			continue
		case !in:
			continue
		case !strings.HasPrefix(t, "|"):
			in = false
			continue
		case strings.HasPrefix(t, "|---"):
			continue
		}
		cells := strings.Split(strings.Trim(t, "|"), "|")
		if len(cells) != 4 {
			return nil, fmt.Errorf("%s: allowed-pairs row %q has %d cells, want 4", File, t, len(cells))
		}
		role, pairs, marker := ticks(cells[0]), ticks(cells[1]), ticks(cells[2])
		if len(role) != 1 || len(pairs) == 0 || len(marker) > 1 {
			return nil, fmt.Errorf("%s: allowed-pairs row %q names no single role, no pair, or more than one marker", File, t)
		}
		e := Entry{Role: role[0], Group: group, Row: Row{Pairs: pairs}}
		if len(marker) == 1 {
			e.Marker = marker[0]
		}
		if d := strings.TrimSpace(cells[3]); d != "—" {
			e.Does = d
		}
		if e.Marker == "" && e.Does == "" {
			return nil, fmt.Errorf("%s: allowed-pairs row %q is unmarked and carries no Does line", File, t)
		}
		if g, ok := groupOf[e.Role]; ok && g != group {
			return nil, fmt.Errorf("%s: role %q sits under two groups, %q and %q", File, e.Role, g, group)
		}
		groupOf[e.Role] = group
		out = append(out, e)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s carries no `%s` table", File, tableHeader)
	}
	return out, nil
}

// ticks is a cell's backticked tokens.
func ticks(c string) []string {
	var out []string
	for _, m := range tick.FindAllStringSubmatch(c, -1) {
		out = append(out, m[1])
	}
	return out
}
