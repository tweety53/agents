package guard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// The **Allowed pairs** table of skills/flow-contracts/model-policy.md as
// mechanics: policyTable parses it, policyAllows judges one dispatch against
// it, and check-dispatch-pair.sh wraps both for hooks/enforce-agent-baseline.py.
// check-loop-dispatch, plan-agent-pair and plan-class read the same table.
func init() {
	Registry["check-dispatch-pair"] = checkDispatchPair
}

// mpFile is the one canonical place the pairs are read from.
const mpFile = "skills/flow-contracts/model-policy.md"

// policyRow is one table row: the pairs it admits, and the whole prompt line
// it requires ("" for none).
type policyRow struct {
	pairs  []string
	marker string
}

var mpRole = regexp.MustCompile(`(?m)^Role: (\S+)$`)

// policyTable parses the `| Role | Allowed pairs | Requires marker |` table out
// of <repo>/mpFile, keyed by role id. A file with no such table, or a row
// missing a cell, is an error: a dispatch is never admitted off a table that
// cannot be read.
func policyTable(repo string) (map[string][]policyRow, error) {
	b, err := os.ReadFile(filepath.Join(repo, mpFile))
	if err != nil {
		return nil, err
	}
	rows := map[string][]policyRow{}
	in := false
	for _, line := range strings.Split(string(b), "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "| Role | Allowed pairs | Requires marker |") {
			in = true
			continue
		}
		if !in {
			continue
		}
		if !strings.HasPrefix(t, "|") {
			break
		}
		if strings.HasPrefix(t, "|---") {
			continue
		}
		cells := strings.Split(strings.Trim(t, "|"), "|")
		if len(cells) != 3 {
			return nil, fmt.Errorf("%s: allowed-pairs row %q has %d cells, want 3", mpFile, t, len(cells))
		}
		role, pairs, marker := ticks(cells[0]), ticks(cells[1]), ticks(cells[2])
		if len(role) != 1 || len(pairs) == 0 || len(marker) > 1 {
			return nil, fmt.Errorf("%s: allowed-pairs row %q names no single role, no pair, or more than one marker", mpFile, t)
		}
		r := policyRow{pairs: pairs}
		if len(marker) == 1 {
			r.marker = marker[0]
		}
		rows[role[0]] = append(rows[role[0]], r)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("%s carries no `| Role | Allowed pairs | Requires marker |` table", mpFile)
	}
	return rows, nil
}

// ticks is a cell's backticked tokens.
func ticks(c string) []string {
	var out []string
	for _, m := range lpTick.FindAllStringSubmatch(c, -1) {
		out = append(out, m[1])
	}
	return out
}

// policyPair spells a dispatch as the table does: `<model>/<type>`, or
// `<model>/<type>@<effort>` when the effort is passed apart (Explore, Plan).
func policyPair(typ, model, effort string) string {
	if effort != "" {
		return model + "/" + typ + "@" + effort
	}
	return model + "/" + typ
}

// policyAllows is nil when an unmarked row of role lists the pair, or a
// marked row lists it and its marker is a whole line of prompt; otherwise an
// error naming the role's allowed pairs.
func policyAllows(rows map[string][]policyRow, role, typ, model, effort, prompt string) error {
	rs, ok := rows[role]
	if !ok {
		return fmt.Errorf("unknown role %q — no row of the allowed-pairs table (%s)", role, mpFile)
	}
	pair := policyPair(typ, model, effort)
	lines := strings.Split(prompt, "\n")
	var allowed []string
	for _, r := range rs {
		has := slices.Contains(r.pairs, pair)
		if has && (r.marker == "" || slices.ContainsFunc(lines, func(l string) bool { return strings.TrimSpace(l) == r.marker })) {
			return nil
		}
		for _, p := range r.pairs {
			if r.marker != "" {
				p += " (with the line `" + r.marker + "`)"
			}
			allowed = append(allowed, p)
		}
	}
	return fmt.Errorf("%s runs on %s (the allowed-pairs table, %s); this dispatch carries %s", role, strings.Join(allowed, ", "), mpFile, pair)
}

func checkDispatchPair(args []string, env Env, stdout, stderr io.Writer) int {
	const self = "check-dispatch-pair"
	if len(args) < 3 || len(args) > 4 || args[0] == "" {
		fmt.Fprint(stderr, "usage: check-dispatch-pair.sh <agents-repo> <subagent_type> <model> [<effort>] < prompt\n")
		return 2
	}
	effort := ""
	if len(args) == 4 {
		effort = args[3]
	}
	rows, err := policyTable(smcAbs(env, args[0]))
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", self, err)
		return 2
	}
	var prompt []byte
	if env.Stdin != nil {
		if prompt, err = io.ReadAll(env.Stdin); err != nil {
			fmt.Fprintf(stderr, "%s: reading the prompt: %v\n", self, err)
			return 2
		}
	}
	m := mpRole.FindSubmatch(prompt)
	if m == nil {
		fmt.Fprintf(stdout, "DISPATCH-PAIR-REFUSED: no `Role:` line — a flow dispatch's prompt names its role as a line `Role: <role id>` of the allowed-pairs table (%s)\n", mpFile)
		return 1
	}
	role := string(m[1])
	if err := policyAllows(rows, role, args[1], args[2], effort, string(prompt)); err != nil {
		fmt.Fprintf(stdout, "DISPATCH-PAIR-REFUSED: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "DISPATCH-PAIR-OK %s\n", role)
	return 0
}
