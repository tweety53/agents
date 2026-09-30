package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// checkPanelFindingsClosed is scripts/check-panel-findings-closed.sh: that
// script's header comment is the contract -- exit 0 FINDINGS-CLOSED, 1 a
// violation (each named on stderr), 2 cannot answer. The findings read is
// Env.Findings when set, else `flow record findings`; the dispatches read is
// Env.Dispatches when set, else `flow record dispatches`.
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

	// THE DISPATCHES READ CARRIES THE FINDINGS READ'S POSTURE, unchanged:
	// read unconditionally -- a store that cannot answer the correlation
	// below cannot pronounce FINDINGS-CLOSED, the same blindness rule the
	// findings read's exit 2 exists for -- and a failed or unreadable read
	// is exit 2, never "no rows". Env.Dispatches is the hook
	// check-panel-fix-single-dispatch shares, nil meaning the CLI on PATH.
	var dispatchesJSON []byte
	if env.Dispatches != nil {
		dispatchesJSON, err = env.Dispatches(name)
		if err != nil {
			fmt.Fprintf(stderr, "%scannot read dispatches for '%s' from the store — cannot determine anything: %v\n", cfcPrefix, name, err)
			return 2
		}
	} else {
		var errOut []byte
		var rc int
		dispatchesJSON, errOut, rc = pcFlow(env, false, "record", "dispatches", "-change", name, "-C", worktree)
		if rc != 0 {
			fmt.Fprintf(stderr, "%scannot read dispatches for '%s' from the store — cannot determine anything: %s\n", cfcPrefix, name, strings.TrimRight(string(errOut), "\n"))
			return 2
		}
	}
	dispatches, ok := cfcParseDispatches(dispatchesJSON)
	if !ok {
		fmt.Fprintln(stderr, cfcJQFailed)
		return 2
	}

	// An open finding is any finding whose status is neither `fixed` nor a
	// `withdrawn <reason>` value -- the prefix covers the whole family,
	// reason text included.
	var open []string
	for _, f := range findings {
		if f.status != "fixed" && !strings.HasPrefix(f.status, "withdrawn") {
			open = append(open, f.ref)
		}
	}
	violated := false
	if len(open) > 0 {
		fmt.Fprintf(stderr, "%sfinding(s) still open: %s\n", cfcPrefix, strings.Join(open, " "))
		violated = true
	}

	// A FINDING IS RECORDED `fixed` ONLY AFTER THE RE-RUN THAT VERIFIES IT
	// (review-panel.md's **Recording findings**, KAN-770), and the store's
	// only witness of that re-run is the raising slot's own dispatch row at
	// a later round, ended `completed`. A `fixed` finding whose slot has
	// none stands verified before its verification existed, which is the
	// defect this class exists to catch: the ordering was luck, not
	// discipline. A slot's re-run qualifies when its row carries role
	// `reviewer`, a key of the shape panel-<round>-... (a handshake retry's
	// trailing -retry tolerated; a panel-fix or task key never matches), a
	// round strictly greater than the finding's, an outcome exactly
	// `completed` -- a timed-out or never-ended dispatch is not a clean
	// re-run -- and a slot whose `+`-components cover every component of the
	// finding's: a bundled re-run covers its members, a solo re-run of one
	// member does not cover a joined finding. Withdrawn findings claim no
	// verification and check nothing here.
	var unverified []string
	for _, f := range findings {
		if f.status != "fixed" {
			continue
		}
		// A fixed Minor closes on its fix alone (review-panel.md's **Panel re-runs**).
		if strings.EqualFold(f.severity, "minor") {
			continue
		}
		required := strings.Split(f.slot, "+")
		covered := false
		for _, d := range dispatches {
			if d.role != "reviewer" || d.outcome != "completed" || d.round <= f.round {
				continue
			}
			provides := strings.Split(d.slot, "+")
			all := true
			for _, r := range required {
				if !slices.Contains(provides, r) {
					all = false
					break
				}
			}
			if all {
				covered = true
				break
			}
		}
		if !covered {
			unverified = append(unverified, f.ref+" ("+f.slot+")")
		}
	}
	if len(unverified) > 0 {
		fmt.Fprintf(stderr, "%sfinding(s) recorded fixed with no clean re-run dispatch of their slot in any later round: %s\n", cfcPrefix, strings.Join(unverified, " "))
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
	slot                  string
}

// cfcDispatchRow is one row of `flow record dispatches`' array, reduced to
// what the fixed-without-clean-rerun correlation reads. round is parsed from
// the key -- the table carries no round column; the panel keys its slot
// dispatches panel-<round>-<slot> -- and -1 marks a key of another shape, so
// such a row can never satisfy a later-round requirement.
type cfcDispatchRow struct {
	key, role, slot, outcome string
	round                    int
}

// cfcPanelRoundRE is the dispatch-key shape the correlation reads. A
// handshake retry's key carries a trailing -retry after the slot, which this
// regexp's prefix match tolerates; panel-fix-<round> and task keys never
// match, panel- not being followed by digits in them.
var cfcPanelRoundRE = regexp.MustCompile(`^panel-([0-9]+)-`)

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
		if pcIsString(o["slot"]) {
			_ = json.Unmarshal(o["slot"], &f.slot)
		}
		out = append(out, f)
	}
	return out, true
}

// cfcParseDispatches decodes `flow record dispatches`' array with the same
// posture cfcParse carries: a non-array or a non-object row is jq failing
// (exit 2, cannot answer), and absent or non-string fields read as the empty
// string, the shape a row legitimately carries when its verb's flag was
// omitted -- an empty role, slot or outcome simply never qualifies below.
func cfcParseDispatches(b []byte) ([]cfcDispatchRow, bool) {
	var raw []map[string]json.RawMessage
	if json.Unmarshal(b, &raw) != nil || raw == nil {
		return nil, false
	}
	out := make([]cfcDispatchRow, 0, len(raw))
	for _, o := range raw {
		if o == nil {
			return nil, false
		}
		var d cfcDispatchRow
		for _, pair := range []struct {
			field string
			dest  *string
		}{{"key", &d.key}, {"role", &d.role}, {"slot", &d.slot}, {"outcome", &d.outcome}} {
			if pcIsString(o[pair.field]) {
				_ = json.Unmarshal(o[pair.field], pair.dest)
			}
		}
		d.round = -1
		if m := cfcPanelRoundRE.FindStringSubmatch(d.key); m != nil {
			if n, err := strconv.Atoi(m[1]); err == nil {
				d.round = n
			}
		}
		out = append(out, d)
	}
	return out, true
}
