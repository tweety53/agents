package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"slices"
	"strings"

	"github.com/tweety53/agents/stats/internal/modelpolicy"
)

// The contract below moved here from the header of the deleted
// scripts/check-dispatch-pair.sh shim (KAN-1058).
//
// flow-guard check-dispatch-pair <agents-repo> <subagent_type> <model> [<effort>] — admit
// one flow dispatch, its prompt on stdin, against the role's allowed pairs in
// flowd's store (`flow policy get`, kan-1103), or refuse it
// (hooks/enforce-agent-baseline.py runs it on every dispatch in a flow run).
// The role is the prompt's first whole line `Role: <role id>`; <effort> is
// passed for Explore and Plan, whose subagent_type carries none.
//
// Exit 0, DISPATCH-PAIR-OK <role>: the role's base pairs list the pair; or a
// marked pair lists it and the prompt carries its marker as a whole line; or
// an opt-in pair lists it, the prompt carries the whole line
// `Operator pair: <role>`, and the dispatch's change -- the `spectre/<name>`
// branch of the prompt's `Worktree:` line, else of the working directory --
// holds a live grant for the role (`flow policy grants`).
// Exit 1, one DISPATCH-PAIR-REFUSED line: no `Role:` line, an unknown role, a
// pair off the role's pairs, a marker missing, or an opt-in pair without its
// line or its grant.
// flow-guard check-dispatch-pair <agents-repo> --role-only checks the Role line alone,
// for a harness whose pair is a fixed mapping (ZCode).
// Exit 2: cannot answer — usage, or the pairs unreadable: flowd unreachable
// and no cache of `flow policy get` either.
//
// Go, in flow-guard (stats/internal/guard/modelpolicy.go).

// The allowed pairs as mechanics: policyTable reads them, policyAllows judges
// one dispatch against them, and check-dispatch-pair wraps both for
// hooks/enforce-agent-baseline.py. check-loop-dispatch, plan-agent-pair,
// plan-class and pipeline-fix read the same pairs, every one with open
// experiment overrides applied.
func init() {
	Registry["check-dispatch-pair"] = checkDispatchPair
}

// mpFile names the roles and their prose; the pairs are flowd's.
const mpFile = modelpolicy.File

var (
	mpRole     = regexp.MustCompile(`(?m)^Role: (\S+)$`)
	mpWorktree = regexp.MustCompile(`(?m)^Worktree: (\S+)$`)
)

// mpOperatorMarker is the line an opt-in dispatch of role carries.
func mpOperatorMarker(role string) string { return "Operator pair: " + role }

// policyRow is one admission rule of a role: Pairs admitted when Marker is
// "" or a whole prompt line, and -- Grant set -- the dispatch's change holds
// an operator grant for the role.
type policyRow struct {
	Pairs  []string
	Marker string
	Grant  bool
}

// policyWire is one role of `flow policy get`.
type policyWire struct {
	Base   []string `json:"base"`
	Optin  []string `json:"optin"`
	Marked []struct {
		Pair   string `json:"pair"`
		Marker string `json:"marker"`
	} `json:"marked"`
}

// policyTable reads every role's pairs through `flow policy get` -- flowd's
// store, or the CLI's last-good cache when flowd is down (decision
// policy-cache) -- as rows: the base row first (unmarked), then one row per
// marker, then the opt-in row (marker `Operator pair: <role>`, needing a
// grant). Every open allowed-pairs experiment is overlaid (KAN-1067): `flow
// experiment overrides`' {role: [pairs]} replaces the base row's pairs;
// roles without pairs and `main` -- the parent agent, never a dispatch --
// are left alone, and a failed or unparsable overrides read leaves the
// pairs alone (decision overrides-via-flow-cli). A planning role's base
// pairs are replaced exactly as any other role's. No answer at all is an
// error naming flowd and the cache.
func policyTable(env Env) (map[string][]policyRow, error) {
	out, errb, rc := pcFlow(env, false, "policy", "get")
	var wire map[string]policyWire
	if rc != 0 || json.Unmarshal(out, &wire) != nil {
		return nil, fmt.Errorf("no allowed pairs: `flow policy get` exit %d — flowd unreachable and no policy cache: %s", rc, strings.TrimSpace(string(errb)))
	}
	rows := map[string][]policyRow{}
	for role, w := range wire {
		rs := []policyRow{{Pairs: w.Base}}
		for _, m := range w.Marked {
			if i := slices.IndexFunc(rs, func(r policyRow) bool { return r.Marker == m.Marker }); i > 0 {
				rs[i].Pairs = append(rs[i].Pairs, m.Pair)
			} else {
				rs = append(rs, policyRow{Pairs: []string{m.Pair}, Marker: m.Marker})
			}
		}
		if len(w.Optin) > 0 {
			rs = append(rs, policyRow{Pairs: w.Optin, Marker: mpOperatorMarker(role), Grant: true})
		}
		rows[role] = rs
	}
	out, _, rc = pcFlow(env, false, "experiment", "overrides")
	var overrides map[string][]string
	if rc != 0 || json.Unmarshal(out, &overrides) != nil {
		return rows, nil
	}
	for role, pairs := range overrides {
		if role == "main" || len(pairs) == 0 || len(rows[role]) == 0 {
			continue
		}
		rows[role][0].Pairs = pairs
	}
	return rows, nil
}

// policyGranted is whether the dispatch's change holds a live operator grant
// for role: the change is the `spectre/<name>` branch of the prompt's
// `Worktree:` line, else of env.Dir; no such branch, or no answer from
// `flow policy grants`, is no grant.
func policyGranted(env Env, role, prompt string) bool {
	dir := env.Dir
	if m := mpWorktree.FindStringSubmatch(prompt); m != nil {
		dir = m[1]
	}
	branch, err := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	name, ok := strings.CutPrefix(strings.TrimSpace(string(branch)), "spectre/")
	if err != nil || !ok || name == "" {
		return false
	}
	out, _, rc := pcFlow(env, false, "policy", "grants", "-change", name)
	var roles []string
	return rc == 0 && json.Unmarshal(out, &roles) == nil && slices.Contains(roles, role)
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
// marked row lists it and its marker is a whole line of prompt -- a grant
// row only when granted, the dispatch's change holding an operator grant
// for role; otherwise an error naming the role's allowed pairs.
func policyAllows(rows map[string][]policyRow, role, typ, model, effort, prompt string, granted bool) error {
	rs, ok := rows[role]
	if !ok {
		return fmt.Errorf("unknown role %q — no allowed pairs in flowd (`flow policy get`)", role)
	}
	pair := policyPair(typ, model, effort)
	lines := strings.Split(prompt, "\n")
	var allowed []string
	for _, r := range rs {
		has := slices.Contains(r.Pairs, pair)
		if has && (r.Marker == "" || slices.ContainsFunc(lines, func(l string) bool { return strings.TrimSpace(l) == r.Marker })) && (!r.Grant || granted) {
			return nil
		}
		for _, p := range r.Pairs {
			switch {
			case r.Grant:
				p += " (with the line `" + r.Marker + "` and an operator grant: flow policy grant -change <name> -roles " + role + ")"
			case r.Marker != "":
				p += " (with the line `" + r.Marker + "`)"
			}
			allowed = append(allowed, p)
		}
	}
	return fmt.Errorf("%s runs on %s (`flow policy get`); this dispatch carries %s", role, strings.Join(allowed, ", "), pair)
}

func checkDispatchPair(args []string, env Env, stdout, stderr io.Writer) int {
	const self = "check-dispatch-pair"
	roleOnly := len(args) == 2 && args[1] == "--role-only"
	if !roleOnly && (len(args) < 3 || len(args) > 4) || args[0] == "" {
		fmt.Fprint(stderr, "usage: flow-guard check-dispatch-pair <agents-repo> (<subagent_type> <model> [<effort>] | --role-only) < prompt\n")
		return 2
	}
	effort := ""
	if len(args) == 4 {
		effort = args[3]
	}
	rows, err := policyTable(env)
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
			fmt.Fprintf(stdout, "DISPATCH-PAIR-REFUSED: unknown role %q — no allowed pairs in flowd (`flow policy get`)\n", role)
			return 1
		}
		fmt.Fprintf(stdout, "DISPATCH-PAIR-OK %s\n", role)
		return 0
	}
	granted := false
	if slices.ContainsFunc(strings.Split(string(prompt), "\n"), func(l string) bool { return strings.TrimSpace(l) == mpOperatorMarker(role) }) {
		granted = policyGranted(env, role, string(prompt))
	}
	if err := policyAllows(rows, role, args[1], args[2], effort, string(prompt), granted); err != nil {
		fmt.Fprintf(stdout, "DISPATCH-PAIR-REFUSED: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "DISPATCH-PAIR-OK %s\n", role)
	return 0
}
