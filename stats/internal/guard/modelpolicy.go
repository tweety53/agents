package guard

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/tweety53/agents/stats/internal/modelpolicy"
)

// The allowed pairs as mechanics: policyTable reads them, for pair-judge and
// visual-verify-pass;
// policyAllows judges one pair against them.

// mpFile names the roles and their prose; the pairs are flowd's.
const mpFile = modelpolicy.File

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
// grant). No answer at all is an error naming flowd and the cache.
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
