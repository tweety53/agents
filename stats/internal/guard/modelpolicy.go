package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/tweety53/agents/stats/internal/modelpolicy"
)

// The **Allowed pairs** table of skills/flow-contracts/model-policy.md as
// mechanics: policyTable parses it, policyAllows judges one dispatch against
// it, and check-dispatch-pair.sh wraps both for hooks/enforce-agent-baseline.py.
// check-loop-dispatch, plan-agent-pair, plan-class and pipeline-fix read the
// same table, every one with open experiment overrides applied.
func init() {
	Registry["check-dispatch-pair"] = checkDispatchPair
}

// mpFile is the one canonical place the pairs are read from.
const mpFile = modelpolicy.File

var mpRole = regexp.MustCompile(`(?m)^Role: (\S+)$`)

// policyTable parses <repo>/mpFile (modelpolicy.Parse) and overlays every
// open allowed-pairs experiment (KAN-1067): `flow experiment overrides`'
// {role: [pairs]} replaces the pairs of each named role's first unmarked
// row. Marked rows, roles the table lacks and `main` -- the parent agent,
// never a dispatch -- are left alone, and a failed or unparsable read
// leaves the table alone (decision overrides-via-flow-cli).
func policyTable(env Env, repo string) (map[string][]modelpolicy.Row, error) {
	b, err := os.ReadFile(filepath.Join(repo, mpFile))
	if err != nil {
		return nil, err
	}
	rows, err := modelpolicy.Parse(b)
	if err != nil {
		return nil, err
	}
	out, _, rc := pcFlow(env, false, "experiment", "overrides")
	var overrides map[string][]string
	if rc != 0 || json.Unmarshal(out, &overrides) != nil {
		return rows, nil
	}
	for role, pairs := range overrides {
		if role == "main" || len(pairs) == 0 {
			continue
		}
		for i, r := range rows[role] {
			if r.Marker == "" {
				rows[role][i].Pairs = pairs
				break
			}
		}
	}
	return rows, nil
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
func policyAllows(rows map[string][]modelpolicy.Row, role, typ, model, effort, prompt string) error {
	rs, ok := rows[role]
	if !ok {
		return fmt.Errorf("unknown role %q — no row of the allowed-pairs table (%s)", role, mpFile)
	}
	pair := policyPair(typ, model, effort)
	lines := strings.Split(prompt, "\n")
	var allowed []string
	for _, r := range rs {
		has := slices.Contains(r.Pairs, pair)
		if has && (r.Marker == "" || slices.ContainsFunc(lines, func(l string) bool { return strings.TrimSpace(l) == r.Marker })) {
			return nil
		}
		for _, p := range r.Pairs {
			if r.Marker != "" {
				p += " (with the line `" + r.Marker + "`)"
			}
			allowed = append(allowed, p)
		}
	}
	return fmt.Errorf("%s runs on %s (the allowed-pairs table, %s); this dispatch carries %s", role, strings.Join(allowed, ", "), mpFile, pair)
}

func checkDispatchPair(args []string, env Env, stdout, stderr io.Writer) int {
	const self = "check-dispatch-pair"
	roleOnly := len(args) == 2 && args[1] == "--role-only"
	if !roleOnly && (len(args) < 3 || len(args) > 4) || args[0] == "" {
		fmt.Fprint(stderr, "usage: check-dispatch-pair.sh <agents-repo> (<subagent_type> <model> [<effort>] | --role-only) < prompt\n")
		return 2
	}
	effort := ""
	if len(args) == 4 {
		effort = args[3]
	}
	rows, err := policyTable(env, smcAbs(env, args[0]))
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
	if roleOnly {
		if _, ok := rows[role]; !ok {
			fmt.Fprintf(stdout, "DISPATCH-PAIR-REFUSED: unknown role %q — no row of the allowed-pairs table (%s)\n", role, mpFile)
			return 1
		}
		fmt.Fprintf(stdout, "DISPATCH-PAIR-OK %s\n", role)
		return 0
	}
	if err := policyAllows(rows, role, args[1], args[2], effort, string(prompt)); err != nil {
		fmt.Fprintf(stdout, "DISPATCH-PAIR-REFUSED: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "DISPATCH-PAIR-OK %s\n", role)
	return 0
}
