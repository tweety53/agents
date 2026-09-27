package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// checkPanelFindingsClosed is scripts/check-panel-findings-closed.sh: that
// script's header comment is the contract -- exit 0 FINDINGS-CLOSED, 1 a
// violation (each named on stderr), 2 cannot answer. The findings read is
// Env.Findings when set, else `flow record findings`.
func init() {
	Registry["check-panel-findings-closed"] = checkPanelFindingsClosed
}

const (
	cfcPrefix   = "check-panel-findings-closed: "
	cfcJQFailed = cfcPrefix + "jq failed — cannot determine anything"
)

func checkPanelFindingsClosed(args []string, env Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	worktreeArg, name := arg(0), arg(1)
	if worktreeArg == "" || !isDir(pcAbs(env, worktreeArg)) {
		shown := worktreeArg
		if shown == "" {
			shown = "<missing>"
		}
		fmt.Fprintf(stderr, "%snot a directory: %s\n", cfcPrefix, shown)
		return 2
	}
	if name == "" {
		fmt.Fprintln(stderr, "usage: check-panel-findings-closed.sh <worktree> <change-name>")
		return 2
	}
	// CONTAINMENT: plainChangeName (cleanupcomplete.go) is canonical for why
	// -- the name arrives from a pull-request-editable state file and is
	// passed to `flow record findings -change`.
	if !plainChangeName(name) {
		fmt.Fprintf(stderr, "%schange name '%s' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n", cfcPrefix, name)
		return 2
	}
	// Canonicalised before it is passed to `flow` as `-C`, and refused if it
	// vanished since the directory check -- panelreproducers.go's comment on
	// the same sequence is canonical for both hazards.
	worktree, err := filepath.EvalSymlinks(pcAbs(env, worktreeArg))
	if err != nil {
		fmt.Fprintf(stderr, "%sworktree vanished before it could be resolved: %s\n", cfcPrefix, worktreeArg)
		return 2
	}

	// STDOUT AND STDERR ARE CAPTURED SEPARATELY: `flow` prints diagnostics
	// such as `flow: using FLOW_ADDR=...` to stderr, and folding them into
	// stdout would put a non-JSON line at the head of what is parsed.
	var findingsJSON []byte
	if env.Findings != nil {
		findingsJSON, err = env.Findings(name)
		if err != nil {
			fmt.Fprintf(stderr, "%scannot read findings for '%s' from the store — cannot determine anything: %v\n", cfcPrefix, name, err)
			return 2
		}
	} else {
		var errOut []byte
		var rc int
		findingsJSON, errOut, rc = pcFlow(env, false, "record", "findings", "-change", name, "-C", worktree)
		if rc != 0 {
			fmt.Fprintf(stderr, "%scannot read findings for '%s' from the store — cannot determine anything: %s\n", cfcPrefix, name, strings.TrimRight(string(errOut), "\n"))
			return 2
		}
	}
	findings, ok := cfcParse(findingsJSON)
	if !ok {
		fmt.Fprintln(stderr, cfcJQFailed)
		return 2
	}

	// An open finding is any finding whose status is neither `fixed` nor a
	// `withdrawn <reason>` or `deferred <reason>` value -- each prefix covers
	// its whole family, reason text included.
	var open []string
	for _, f := range findings {
		if f.status != "fixed" && !strings.HasPrefix(f.status, "withdrawn") && !strings.HasPrefix(f.status, "deferred") {
			open = append(open, f.ref)
		}
	}
	violated := false
	if len(open) > 0 {
		fmt.Fprintf(stderr, "%sfinding(s) still open: %s\n", cfcPrefix, strings.Join(open, " "))
		violated = true
	}

	// A MINOR IS NEVER DEFERRED BESIDE A CRITICAL OR IMPORTANT OF ITS OWN
	// ROUND -- review-panel.md's **Panel re-runs** is canonical for the rule.
	// A withdrawn Critical or Important does not count: it takes no fix
	// round for a Minor to join, and the handback loop's own text defers a
	// Minor "with none" taking one.
	fixRound := map[int]bool{}
	for _, f := range findings {
		sev := strings.ToLower(f.severity)
		if (sev == "critical" || sev == "important") && !strings.HasPrefix(f.status, "withdrawn") {
			fixRound[f.round] = true
		}
	}
	misdeferred := map[int][]string{}
	for _, f := range findings {
		if strings.EqualFold(f.severity, "minor") && strings.HasPrefix(f.status, "deferred") && fixRound[f.round] {
			misdeferred[f.round] = append(misdeferred[f.round], f.ref)
		}
	}
	rounds := make([]int, 0, len(misdeferred))
	for r := range misdeferred {
		rounds = append(rounds, r)
	}
	slices.Sort(rounds)
	for _, r := range rounds {
		fmt.Fprintf(stderr, "%sround %d raised a Critical or Important, so its Minor finding(s) go to that same fix, never deferred: %s\n", cfcPrefix, r, strings.Join(misdeferred[r], " "))
		violated = true
	}
	if violated {
		return 1
	}
	fmt.Fprintln(stdout, "FINDINGS-CLOSED")
	return 0
}

type cfcFindingRow struct {
	ref, status, severity string
	round                 int
}

// cfcParse decodes `flow record findings`' array. It refuses, where the
// bash guard's jq failed: a non-array, a non-object row, a non-string
// status. It also refuses empty output, which jq read as zero findings -- a
// store read that printed nothing is not an answer, the same rule
// pcParseFindings applies. A null or absent ref joins as the empty string,
// a scalar ref as its JSON text (jq's join); a non-string severity reads as
// none, and an absent or non-integer round as 0, the store's zero value.
func cfcParse(b []byte) ([]cfcFindingRow, bool) {
	var raw []map[string]json.RawMessage
	if json.Unmarshal(b, &raw) != nil || raw == nil {
		return nil, false
	}
	out := make([]cfcFindingRow, 0, len(raw))
	for _, o := range raw {
		if o == nil || !pcIsString(o["status"]) {
			return nil, false
		}
		var f cfcFindingRow
		_ = json.Unmarshal(o["status"], &f.status)
		switch ref := strings.TrimSpace(string(o["ref"])); {
		case ref == "" || ref == "null":
		case pcIsString(o["ref"]):
			_ = json.Unmarshal(o["ref"], &f.ref)
		case strings.HasPrefix(ref, "{") || strings.HasPrefix(ref, "["):
			return nil, false
		default:
			f.ref = ref
		}
		if pcIsString(o["severity"]) {
			_ = json.Unmarshal(o["severity"], &f.severity)
		}
		if n, err := strconv.Atoi(strings.TrimSpace(string(o["round"]))); err == nil {
			f.round = n
		}
		out = append(out, f)
	}
	return out, true
}
