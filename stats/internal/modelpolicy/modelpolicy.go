// Package modelpolicy parses the roles tables of
// skills/flow-contracts/model-policy.md -- each role's id, group and Does
// line. A role's allowed pairs are not here: they live in flowd's store
// (kan-1103, decision pairs-in-db), read with `flow policy get`.
package modelpolicy

import (
	"fmt"
	"regexp"
	"strings"
)

// File is the one canonical place the roles are read from, relative to the
// agents repository root.
const File = "skills/flow-contracts/model-policy.md"

// Entry is one table row: its role id, its group -- the nearest `###`
// heading above its table -- and its Does line.
type Entry struct {
	Role, Group, Does string
}

var tick = regexp.MustCompile("`([^`]+)`")

// tableHeader opens every roles table.
const tableHeader = "| Role | Does |"

// Rows reads the `| Role | Does |` tables out of b, one entry per row in
// table order -- the order flowd's roles table is drawn in. Every table is
// read until the first `## ` heading after a row. A file with no such
// table, a row of the wrong width, a row naming no single role or no Does
// line, or a role named twice is an error: a dispatch is never admitted off
// a table that cannot be read.
func Rows(b []byte) ([]Entry, error) {
	var out []Entry
	group, in := "", false
	seen := map[string]bool{}
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
		case t == tableHeader:
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
		if len(cells) != 2 {
			return nil, fmt.Errorf("%s: roles row %q has %d cells, want 2", File, t, len(cells))
		}
		role := tick.FindAllStringSubmatch(cells[0], -1)
		does := strings.TrimSpace(cells[1])
		if len(role) != 1 || does == "" || does == "—" {
			return nil, fmt.Errorf("%s: roles row %q names no single role or no Does line", File, t)
		}
		if seen[role[0][1]] {
			return nil, fmt.Errorf("%s: role %q has two rows", File, role[0][1])
		}
		seen[role[0][1]] = true
		out = append(out, Entry{Role: role[0][1], Group: group, Does: does})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s carries no `%s` table", File, tableHeader)
	}
	return out, nil
}
