package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
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
	// CONTAINMENT: plainChangeName (changename.go) is canonical for why
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
	// `withdrawn <reason>` value -- and, since KAN-791, a bare `withdrawn`
	// (the word with no reason after it) is open too: a reasonless
	// withdrawal is the silent drop the finding-status contract forbids, so
	// it can no longer read as a closed state here, the same line the
	// store's withdrawnWithoutReason and check-unfinished-work's own copy
	// draw. A `deferred <reason>` row is open the same way: nothing is
	// deferred (KAN-862).
	var open []string
	for _, f := range findings {
		rest, withdrawn := strings.CutPrefix(f.status, "withdrawn")
		if f.status != "fixed" && !(withdrawn && strings.TrimSpace(rest) != "") {
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
	// a later round, ended `completed` or `clean`. A `fixed` finding whose slot has
	// none stands verified before its verification existed, which is the
	// defect this class exists to catch: the ordering was luck, not
	// discipline. A slot's re-run qualifies when its row carries role
	// `reviewer`, a key of the shape panel-<round>-... (a handshake retry's
	// trailing -retry tolerated; a panel-fix or task key never matches), a
	// round strictly greater than the finding's, an outcome exactly
	// `completed` or `clean` -- review-panel.md closes a slot `completed`,
	// while `flow record dispatch end` also takes a gated reviewer's
	// `clean`, and 50 panel rows in the store end that way (KAN-887's
	// self-review); a timed-out or never-ended dispatch is not a clean
	// re-run -- and a slot whose `+`-components cover every component of the
	// finding's: a bundled re-run covers its members, a solo re-run of one
	// member does not cover a joined finding. Withdrawn findings claim no
	// verification and check nothing here.
	//
	// A ROW FOLDED ONTO ANOTHER ROLE'S ROW is covered by that row's slot
	// instead: review-panel.md's **A defect raised by several roles is
	// recorded per role and fixed once** links it with -supersedes to a row
	// of the same round and re-runs only the first row's role. A
	// -supersedes naming a row of an earlier round is the same defect
	// re-raised by a later fix round, which still re-runs its own role, so
	// only a same-round target lends its slot.
	byRef := map[string]cfcFindingRow{}
	for _, f := range findings {
		byRef[f.ref] = f
	}
	// A RE-RUN COUNTS ONLY ONCE ITS REPORT FILE IS WRITTEN (KAN-915): a
	// dispatch provides a required component only when
	// panel-report-<round>-<component>.md is non-empty and not the parent's
	// `no verbatim report captured` placeholder. A re-run that would have
	// covered the finding but for a missing report names that file.
	covers := func(slot string, round int) (bool, []string) {
		required := strings.Split(slot, "+")
		var missing []string
		for _, d := range dispatches {
			if d.role != "reviewer" || (d.outcome != "completed" && d.outcome != "clean") || d.round <= round {
				continue
			}
			provides := strings.Split(d.slot, "+")
			all := true
			var lacking []string
			for _, r := range required {
				if !slices.Contains(provides, r) {
					all = false
					break
				}
				if !cfcReportWritten(worktree, d.round, r) {
					lacking = append(lacking, fmt.Sprintf("panel-report-%d-%s.md", d.round, r))
				}
			}
			if !all {
				continue
			}
			if len(lacking) == 0 {
				return true, nil
			}
			missing = append(missing, lacking...)
		}
		return false, missing
	}
	var unverified, unreported []string
	for _, f := range findings {
		if f.status != "fixed" {
			continue
		}
		// A fixed Minor closes on its fix alone (review-panel.md's **Panel re-runs**).
		if strings.EqualFold(f.severity, "minor") {
			continue
		}
		covered, missing := covers(f.slot, f.round)
		if t, ok := byRef[f.supersedes]; !covered && ok && f.supersedes != "" && t.round == f.round {
			var more []string
			covered, more = covers(t.slot, f.round)
			missing = append(missing, more...)
		}
		if !covered {
			unverified = append(unverified, f.ref+" ("+f.slot+")")
			for _, m := range missing {
				if !slices.Contains(unreported, m) {
					unreported = append(unreported, m)
				}
			}
		}
	}
	if len(unreported) > 0 {
		fmt.Fprintf(stderr, "%sre-run(s) answered without writing their report file: %s\n", cfcPrefix, strings.Join(unreported, " "))
	}
	if len(unverified) > 0 {
		fmt.Fprintf(stderr, "%sfinding(s) recorded fixed with no clean re-run dispatch of their slot in any later round: %s\n", cfcPrefix, strings.Join(unverified, " "))
		violated = true
	}

	// A SLOT'S REPORT COUNTS ITS OWN FINDINGS on one `Critical: <n>` and one
	// `Important: <n>` line (review-panel.md's REPORT FILE paragraph). When
	// the sum exceeds the store's Critical and Important rows of that round
	// whose slot's `+`-components include the report's id, any status, a
	// finding the slot raised never reached the store -- kan-1000's round 11.
	// A report missing either line predates the paragraph and is not compared.
	uncounted, err := cfcUncounted(worktree, findings)
	if err != nil {
		fmt.Fprintf(stderr, "%s%v\n", cfcPrefix, err)
		return 2
	}
	if len(uncounted) > 0 {
		fmt.Fprintf(stderr, "%sfinding(s) a slot's report counts that the store does not hold: %s\n", cfcPrefix, strings.Join(uncounted, " "))
		violated = true
	}
	if violated {
		return 1
	}
	fmt.Fprintln(stdout, "FINDINGS-CLOSED")
	return 0
}

// cfcUncounted answers the slot reports under <worktree>'s
// .superpowers/sdd whose Critical and Important sum exceeds the store's rows
// of that round and slot, each as `<file> (report <n>, recorded <m>)`.
func cfcUncounted(worktree string, findings []cfcFindingRow) ([]string, error) {
	reports, _ := filepath.Glob(filepath.Join(worktree, ".superpowers", "sdd", "panel-report-*.md"))
	var uncounted []string
	for _, p := range reports {
		m := cfcReportNameRE.FindStringSubmatch(filepath.Base(p))
		if m == nil {
			continue
		}
		round, _ := strconv.Atoi(m[1])
		body, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("cannot read panel report %s — cannot determine anything: %v", p, err)
		}
		c, i := cfcCriticalRE.FindSubmatch(body), cfcImportantRE.FindSubmatch(body)
		if c == nil || i == nil {
			continue
		}
		nc, _ := strconv.Atoi(string(c[1]))
		ni, _ := strconv.Atoi(string(i[1]))
		recorded := 0
		for _, f := range findings {
			if f.round == round && (strings.EqualFold(f.severity, "critical") || strings.EqualFold(f.severity, "important")) && slices.Contains(strings.Split(f.slot, "+"), m[2]) {
				recorded++
			}
		}
		if nc+ni > recorded {
			uncounted = append(uncounted, fmt.Sprintf("%s (report %d, recorded %d)", filepath.Base(p), nc+ni, recorded))
		}
	}
	return uncounted, nil
}

// cfcReportWritten reports whether a re-run wrote its report: the file is
// non-empty and does not begin with the parent's placeholder.
func cfcReportWritten(worktree string, round int, id string) bool {
	b, err := os.ReadFile(filepath.Join(worktree, ".superpowers", "sdd", fmt.Sprintf("panel-report-%d-%s.md", round, id)))
	return err == nil && len(b) > 0 && !bytes.HasPrefix(b, []byte("no verbatim report captured"))
}

type cfcFindingRow struct {
	ref, status, severity string
	round                 int
	slot, supersedes      string
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

// cfcReportNameRE matches a slot's panel-report-<round>-<slot>.md file name,
// and cfcCriticalRE / cfcImportantRE read the Critical: / Important: count
// lines of that report.
var (
	cfcReportNameRE = regexp.MustCompile(`^panel-report-([0-9]+)-(.+)\.md$`)
	cfcCriticalRE   = regexp.MustCompile(`(?m)^Critical: ([0-9]+)\s*$`)
	cfcImportantRE  = regexp.MustCompile(`(?m)^Important: ([0-9]+)\s*$`)
)

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
		if pcIsString(o["supersedes"]) {
			_ = json.Unmarshal(o["supersedes"], &f.supersedes)
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
