package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// checkVerifyGreen is scripts/check-verify-green.sh: that script's header
// comment is the contract. The dispatch read is Env.Dispatches when set,
// else `flow record dispatches`.
func init() { Registry["check-verify-green"] = checkVerifyGreen }

const vgPrefix = "check-verify-green: "

var vgFixSuffix = regexp.MustCompile(`^verify-fix-[0-9]+`)

func checkVerifyGreen(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 3 || args[1] == "" || args[2] == "" {
		fmt.Fprintf(stderr, "%susage: check-verify-green.sh <worktree> <change-name> <session-token>\n", vgPrefix)
		return 2
	}
	wt, name, token := args[0], args[1], args[2]
	if !isDir(pcAbs(env, wt)) {
		fmt.Fprintf(stderr, "%snot a directory: %s\n", vgPrefix, wt)
		return 2
	}
	if !plainChangeName(name) {
		fmt.Fprintf(stderr, "%schange name '%s' is not a plain change name\n", vgPrefix, name)
		return 2
	}
	raw, err := pfdRead(env, env.Dispatches, "dispatches", name, pcAbs(env, wt))
	var rows []struct{ Key, Role, SessionToken, Outcome string }
	if err == nil {
		err = json.Unmarshal(raw, &rows)
	}
	if err != nil {
		fmt.Fprintf(stderr, "%sflow record dispatches failed for '%s' -- cannot answer\n", vgPrefix, name)
		return 2
	}
	// Rows come back in seq order; only the latest row per worktree is
	// judged, a fix round's `verify-fix-<k>[-<basename>]` superseding the
	// `verify[-<basename>]` row it re-ran.
	latest := map[string]int{}
	var groups []string
	for i, r := range rows {
		if r.SessionToken != token || r.Role != "verifier" || (r.Key != "verify" && !strings.HasPrefix(r.Key, "verify-")) {
			continue
		}
		g := vgFixSuffix.ReplaceAllString(r.Key, "verify")
		if _, ok := latest[g]; !ok {
			groups = append(groups, g)
		}
		latest[g] = i
	}
	seen := len(groups)
	var open []string
	for _, g := range groups {
		r := rows[latest[g]]
		switch r.Outcome {
		case "completed":
		case "":
			open = append(open, r.Key+" (running)")
		default:
			open = append(open, r.Key+" ("+r.Outcome+")")
		}
	}
	switch {
	case seen == 0:
		fmt.Fprintf(stdout, "VERIFY-NOT-GREEN: %s — this run recorded no inline verify\n", name)
		return 1
	case len(open) > 0:
		fmt.Fprintf(stdout, "VERIFY-NOT-GREEN: %s — %s\n", name, strings.Join(open, ", "))
		return 1
	}
	fmt.Fprintf(stdout, "VERIFY-GREEN: %s — %d inline verify row(s) completed\n", name, seen)
	return 0
}
