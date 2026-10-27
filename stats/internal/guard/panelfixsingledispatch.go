package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

// checkPanelFixSingleDispatch is scripts/check-panel-fix-single-dispatch.sh:
// that script's header comment is the contract -- exit 0 every panel-fix row
// of the session token is shape- and count-clean, 1 violations (each on
// stderr), 2 cannot answer. The dispatch and findings reads are
// Env.Dispatches and Env.Findings when set, else `flow record dispatches` and
// `flow record findings`. The reasoning for each branch, moved here from the
// bash body it replaced (d71a2327), sits beside the code it explains.
//
// The bash needed jq and refused at exit 2 without it; the rows are decoded
// with encoding/json here, so there is no tool left to find missing.
func init() {
	Registry["check-panel-fix-single-dispatch"] = checkPanelFixSingleDispatch
}

const pfdPrefix = "check-panel-fix-single-dispatch: "

// pfdKeyRE is the bash's CANONICAL_KEY_RE; its round counts fix rounds
// from 1, the fix after panel-0 being panel-fix-1 (kan-1040). RE2 and bash's ERE agree on it:
// no backreference, no locale-dependent class ([0-9] is ASCII in both under
// the bash's LC_ALL=C), and ^/$ anchor the whole string in both --
// TestCheckPanelFixSingleDispatch's "canonical key regexp matches bash"
// pins the verdicts measured from bash.
var pfdKeyRE = regexp.MustCompile(`^panel-fix-[1-9][0-9]*(-[0-9]+)?(-retry)?$`)

// pfdShaRE is a recorded commit git may be handed: hex only, so a value
// opening with `-` never reaches git as an option.
var pfdShaRE = regexp.MustCompile(`^[0-9a-f]{4,64}$`)

var pfdFullSuiteRE = regexp.MustCompile(`^full-suite-fix-[0-9]+$`)

// pfdBase is one panel-fix base key (the key with a trailing -retry
// stripped) and its counts, kept in first-seen order as the bash's parallel
// arrays were. orig and retries count this session token's rows; foreign
// counts the originals other session tokens of the change recorded under the
// same key, which answer only whether the key's original exists (KAN-890).
type pfdBase struct {
	key                    string
	round, chunk           int64
	orig, retries, foreign int
}

func checkPanelFixSingleDispatch(args []string, env Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	worktreeArg, name, token := arg(0), arg(1), arg(2)
	if worktreeArg == "" || !isDir(pcAbs(env, worktreeArg)) {
		shown := worktreeArg
		if shown == "" {
			shown = "<missing>"
		}
		fmt.Fprintf(stderr, "%snot a directory: %s\n", pfdPrefix, shown)
		return 2
	}
	if name == "" {
		fmt.Fprintf(stderr, "%susage: check-panel-fix-single-dispatch.sh <worktree> <change-name> <session-token>\n", pfdPrefix)
		return 2
	}
	if token == "" {
		fmt.Fprintf(stderr, "%ssession token is required and must be this run's own literal token\n", pfdPrefix)
		return 2
	}

	// CONTAINMENT, identical to check-panel-findings-closed.sh's own copy:
	// the change name arrives from a pull-request-editable state file and is
	// passed to `flow record dispatches -change`, so `../../../planted` and a
	// glob metacharacter are hazards here exactly as they are there. The
	// check is plainChangeName (changename.go), whose comment is
	// canonical for the reasoning; it compares bytes, as the bash's
	// `export LC_ALL=C` made its enumeration do.
	if !plainChangeName(name) {
		fmt.Fprintf(stderr, "%schange name '%s' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n", pfdPrefix, name)
		return 2
	}

	// Canonicalise before the worktree is ever passed to `flow` as `-C`, and
	// refuse rather than proceed if it vanished between the directory check
	// above and here -- see check-panel-findings-closed.sh's own comment on
	// both hazards. EvalSymlinks is the bash's `cd ... && pwd -P`; an
	// argument starting with `-` is refused as the bash's `cd` refused it,
	// reading it as an option, and so is a directory the caller cannot
	// search (the bash's `cd` failed there; its own `cd: …: Permission
	// denied` line names the script's path and is not reproduced). The line
	// names no path: the bash printed $WORKTREE after the failed
	// substitution had already assigned it "".
	worktree, err := filepath.EvalSymlinks(pcAbs(env, worktreeArg))
	if err != nil || strings.HasPrefix(worktreeArg, "-") || syscall.Access(pcAbs(env, worktreeArg), 1) != nil {
		fmt.Fprintf(stderr, "%sworktree vanished: \n", pfdPrefix)
		return 2
	}

	// A store the call could not reach is "cannot answer", never "no rows":
	// an outage that read as zero rows would pass every round it was blind
	// to.
	rows, err := pfdRead(env, env.Dispatches, "dispatches", name, worktree)
	if err != nil {
		fmt.Fprintf(stderr, "%sflow record dispatches failed for '%s' -- cannot answer\n", pfdPrefix, name)
		return 2
	}

	// The findings read carries the same posture (kan-499): the chunk bound
	// is derived from it, so a read that fails is a question this guard
	// cannot answer, never a round it pretends was clean.
	findings, err := pfdRead(env, env.Findings, "findings", name, worktree)
	if err != nil {
		fmt.Fprintf(stderr, "%sflow record findings failed for '%s' -- cannot answer\n", pfdPrefix, name)
		return 2
	}

	// Each panel-fix row of THIS token, and apart from them every other
	// token's. A payload that is not the shape the verb documents is the same
	// cannot-answer posture as a failed call.
	fixRows, foreignRows, ok := pfdPanelFixRows(rows, token)
	if !ok {
		fmt.Fprintf(stderr, "%sdispatch rows were not readable JSON -- cannot answer\n", pfdPrefix)
		return 2
	}
	findingVals, ok := pfdStream(findings)
	if !ok {
		fmt.Fprintf(stderr, "%sfindings rows were not readable JSON -- cannot answer\n", pfdPrefix)
		return 2
	}

	var violations []string
	var bases []*pfdBase
	index := map[string]*pfdBase{}
	baseOf := func(base string) *pfdBase {
		b := index[base]
		if b == nil {
			// rest is "<round>" or "<round>-<chunk>"; parsed base 10 (the
			// bash's 10#) so a leading zero never reads as octal.
			rest := strings.TrimPrefix(base, "panel-fix-")
			roundS, chunkS, _ := strings.Cut(rest, "-")
			b = &pfdBase{key: base, round: bashDecimal(roundS), chunk: bashDecimal(chunkS)}
			index[base] = b
			bases = append(bases, b)
		}
		return b
	}
	for _, row := range fixRows {
		key := row.key
		if !pfdKeyRE.MatchString(key) {
			violations = append(violations, fmt.Sprintf("out-of-shape panel-fix key '%s' -- the canonical shape is panel-fix-<round>[-<chunk>][-retry]", key))
			continue
		}
		base, retry := strings.CutSuffix(key, "-retry")
		b := baseOf(base)
		if retry {
			b.retries++
		} else {
			b.orig++
		}
	}
	// A FIX ROUND RESUMED UNDER A FRESH SESSION TOKEN (KAN-890): kan-789's
	// resumed round recorded panel-fix-1-retry under its new token while
	// panel-fix-1 sat under the token that started the round, and the
	// retry-alone rule read a retry with no original. Another token's
	// original of a key therefore answers whether that key's original
	// exists -- for a retry, a chunked round's bare key and its chunk
	// contiguity, and so toward the chunk count the bound judges, which
	// counts the round's chunks under every token of the change. The per-key
	// dispatch counts stay this token's, so two runs that each dispatched a round once are never one round
	// dispatched twice, and a round no row of this token touched is never
	// judged.
	for _, row := range foreignRows {
		if pfdKeyRE.MatchString(row.key) && !strings.HasSuffix(row.key, "-retry") {
			baseOf(row.key).foreign++
		}
	}

	for _, b := range bases {
		total := b.orig + b.retries
		switch {
		case b.retries > 0 && b.orig == 0 && b.foreign == 0:
			violations = append(violations, fmt.Sprintf("round %s carries %d -retry dispatch(es) and no original -- a retry is never a round's only dispatch", b.key, b.retries))
		case b.orig > 1:
			violations = append(violations, fmt.Sprintf("round %s carries %d panel-fix dispatches -- the fix goes to ONE subagent as the combined list, never one per reviewer, slot or finding", b.key, b.orig))
		case total > 2 || (total == 2 && b.retries != 1):
			violations = append(violations, fmt.Sprintf("round %s carries %d panel-fix dispatches -- at most the original plus the handshake's one -retry", b.key, total))
		}
	}

	// ---- per-round chunk-shape checks (kan-499) ---------------------------
	// A round that chunked its fix list must have started numbering at its
	// own bare key, numbered contiguously from -2, and stayed under the
	// findings bound. Single-dispatch rounds (no chunked base at all) check
	// nothing here.
	var rounds []int64
	seen := map[int64]bool{}
	for _, b := range bases {
		if b.orig+b.retries > 0 && !seen[b.round] {
			seen[b.round] = true
			rounds = append(rounds, b.round)
		}
	}
	for _, r := range rounds {
		var bare *pfdBase
		var chunks []*pfdBase
		for _, b := range bases {
			switch {
			case b.round != r:
			case b.chunk == 0:
				bare = b
			default:
				chunks = append(chunks, b)
			}
		}
		if len(chunks) == 0 {
			continue
		}
		if bare == nil || bare.orig+bare.foreign == 0 {
			violations = append(violations, fmt.Sprintf("round panel-fix-%d carries chunked dispatches but no panel-fix-%d original -- chunk numbering starts at the round's own key", r, r))
		}

		// Contiguity runs over chunks that carry an original dispatch; a
		// retry-only chunk base is already flagged per base above and
		// proves nothing about numbering.
		var present []int64
		for _, b := range chunks {
			if b.orig+b.foreign >= 1 {
				present = append(present, b.chunk)
			}
		}
		sort.Slice(present, func(i, j int) bool { return present[i] < present[j] })
		expect := int64(2)
		for _, c := range present {
			if c != expect {
				violations = append(violations, fmt.Sprintf("round panel-fix-%d carries non-contiguous chunks: expected panel-fix-%d-%d, found panel-fix-%d-%d", r, r, expect, r, c))
				break
			}
			expect++
		}
		k := expect - 1

		prior, ok := pfdPriorFindings(findingVals, r)
		if !ok {
			fmt.Fprintf(stderr, "%sfindings rows were not readable JSON -- cannot answer\n", pfdPrefix)
			return 2
		}
		// An empty findings payload is jq printing nothing: the bash's prior
		// was "" -- shown empty, and 0 in the arithmetic.
		n, _ := strconv.ParseInt(prior, 10, 64)
		bound := (n + 9) / 10
		if k > bound {
			violations = append(violations, fmt.Sprintf("round panel-fix-%d carries %d chunk dispatch(es) for %s finding(s) raised in earlier rounds -- at most ceil(%s/10)=%d, never one per finding", r, k, prior, prior, bound))
		}
	}

	// A panel-fix row's -commit is the task commit its fixup folded into, or
	// on the pushed route the fix commit itself, carrying the round's
	// Flow-Key: trailer (review-panel.md). Any other commit is the wrong
	// record -- kan-952's panel-fix-0 recorded its plan commit while its fix
	// landed as 50af0ddf.
	commitViolations, ok := pfdCheckCommits(env, worktree, name, fixRows, stderr)
	if !ok {
		return 2
	}
	violations = append(violations, commitViolations...)

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintf(stderr, "%s%s\n", pfdPrefix, v)
		}
		return 1
	}
	fmt.Fprintln(stdout, "PANEL-FIX-SINGLE-DISPATCH-OK: every panel-fix dispatch of this run is one per chunked round (bounds and retries included)")
	return 0
}

// pfdRow is one panel-fix row of the run: its key, and the commitSha its
// end recorded ("" when none).
type pfdRow struct{ key, commit string }

// pfdPanelFixRows is `.[] | select((.sessionToken // "") == $tok and
// (.role // "") == "panel-fix")` then each row's `.key // ""` as jq -r
// printed it, beside the row's commitSha; the panel-fix rows of every other
// session token come back apart, as foreign. A null row selects nothing; any other non-object row, or a
// top-level value that is not an array, is jq failing. (jq iterated a
// top-level object's values in insertion order; that shape is refused here
// rather than guessed at -- the verb prints an array.)
func pfdPanelFixRows(raw []byte, token string) (keys, foreign []pfdRow, ok bool) {
	vals, ok := pfdStream(raw)
	if !ok {
		return nil, nil, false
	}
	for _, v := range vals {
		arr, ok := v.([]any)
		if !ok {
			return nil, nil, false
		}
		for _, e := range arr {
			if e == nil {
				continue
			}
			row, ok := e.(map[string]any)
			if !ok {
				return nil, nil, false
			}
			if row["role"] != "panel-fix" {
				continue
			}
			key := pfdRawText(row["key"])
			// A full-suite-fix-<n> row is the parent's own full-suite round
			// (implement.md), recorded under role panel-fix for cost with
			// -agent-id inline, not a panel fix round: it never counts here.
			// A subagent's row under that key still does (KAN-482).
			if row["agentId"] == "inline" && pfdFullSuiteRE.MatchString(key) {
				continue
			}
			if row["sessionToken"] != token {
				foreign = append(foreign, pfdRow{key: key})
				continue
			}
			commit, _ := row["commitSha"].(string)
			keys = append(keys, pfdRow{key: key, commit: commit})
		}
	}
	return keys, foreign, true
}

// pfdCheckCommits names each row whose recorded commit carries neither a
// task's Commit: subject nor a Flow-Key: trailer naming the row's key (its
// -retry stripped), looked up in the worktree and then in each peer its
// spec root's peers file names. Rows that recorded no commit are not judged;
// false is cannot answer -- a tasks.md that does not resolve or read.
func pfdCheckCommits(env Env, worktree, name string, fixRows []pfdRow, stderr io.Writer) ([]string, bool) {
	var judged []pfdRow
	for _, r := range fixRows {
		if r.commit != "" {
			judged = append(judged, r)
		}
	}
	if len(judged) == 0 {
		return nil, true
	}
	tasksMD, _ := tcfResolveTasksMD(env, io.Discard, "check-panel-fix-single-dispatch", worktree, name, worktree)
	raw, err := os.ReadFile(tasksMD)
	var plan []string
	if err == nil {
		plan, err = tcfPlanLines(raw)
	}
	if err != nil {
		fmt.Fprintf(stderr, "%scannot read tasks.md for '%s' -- cannot answer\n", pfdPrefix, name)
		return nil, false
	}
	subjects := map[string]bool{}
	for _, id := range tcfTaskIDs(plan) {
		if t, ok := tcfParseTask(plan, id); ok && t.commit != nil {
			subjects[*t.commit] = true
		}
	}
	repos := append([]string{worktree}, pfdPeerRoots(env, worktree)...)
	var violations []string
	for _, r := range judged {
		var meta []string
		for _, repo := range repos {
			if !pfdShaRE.MatchString(r.commit) {
				break
			}
			if out, err := tcfRunGit(env, repo, "log", "-1", "--format=%s%n%(trailers:key=Flow-Key,valueonly)", r.commit, "--"); err == nil {
				meta = strings.Split(strings.TrimSpace(out), "\n")
				break
			}
		}
		if meta == nil {
			violations = append(violations, fmt.Sprintf("%s records commit %s, which no repository of the change holds -- -commit is the commit the fix landed in", r.key, r.commit))
			continue
		}
		base, _ := strings.CutSuffix(r.key, "-retry")
		if !subjects[meta[0]] && !slices.Contains(meta[1:], base) && !slices.Contains(meta[1:], r.key) {
			violations = append(violations, fmt.Sprintf("%s records commit %s (%q), which is no task's Commit: subject and carries no Flow-Key: %s trailer -- -commit is the task commit the fixup was folded into, or the fix commit itself", r.key, r.commit, meta[0], base))
		}
	}
	return violations, true
}

// pfdPeerRoots is every peer checkout <worktree>/<spec root>/peers names,
// resolved as changePlanPeerRoot resolves one; a peer that does not resolve
// is skipped.
func pfdPeerRoots(env Env, worktree string) []string {
	leaf := specRootLeaf(worktree, io.Discard)
	b, err := os.ReadFile(worktree + "/" + leaf + "/peers")
	if err != nil {
		return nil
	}
	var roots []string
	for _, line := range lines(b) {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		if root, ok := changePlanPeerRoot(env, worktree, leaf, f[0]); ok {
			roots = append(roots, root)
		}
	}
	return roots
}

// pfdRawText is `.key // ""` as jq -r printed it and `$(...)` captured it:
// a string raw with its NUL bytes dropped and then trailing newlines
// stripped, null/false as "", anything else as its JSON text. One
// divergence: for a key that is an array, object or exponent number, jq -r
// printed pretty, insertion-ordered JSON (`1E+2`); this prints Go's compact
// encoding. Only the shape violation's quoted key differs -- the exit is the
// same, and `flow` never writes a non-string key.
func pfdRawText(v any) string {
	switch k := v.(type) {
	case nil:
		return ""
	case bool:
		if !k {
			return ""
		}
		return "true"
	case string:
		return strings.TrimRight(strings.ReplaceAll(k, "\x00", ""), "\n")
	default:
		b, _ := json.Marshal(k)
		return string(b)
	}
}

// pfdPriorFindings is `[.[] | select((.round // 0) < $r)] | length` over
// the findings stream, printed as the bash captured it: "" for no input
// value. A null or non-object finding counts as round 0 or fails as jq did;
// a round compares in jq's order (null/false as 0, true below every
// number, strings/arrays/objects above). A stream of several values or a
// jq error is not a shape the verb prints: false, cannot answer -- the
// header's contract. The body diverged: a jq error exited 5 (or 1) under
// set -e, and a stream of several values printed an arithmetic error and
// then passed the round with the OK line.
func pfdPriorFindings(vals []any, r int64) (string, bool) {
	if len(vals) == 0 {
		return "", true
	}
	if len(vals) > 1 {
		return "", false
	}
	var elems []any
	switch v := vals[0].(type) {
	case []any:
		elems = v
	case map[string]any:
		for _, e := range v {
			elems = append(elems, e)
		}
	default:
		return "", false
	}
	n := 0
	for _, e := range elems {
		var round any
		switch f := e.(type) {
		case nil:
		case map[string]any:
			round = f["round"]
		default:
			return "", false
		}
		switch x := round.(type) {
		case nil, bool:
			if x == true || float64(0) < float64(r) {
				n++
			}
		case json.Number:
			if f, err := x.Float64(); err == nil && f < float64(r) {
				n++
			}
		}
	}
	return strconv.Itoa(n), true
}

// bashDecimal is bash's $((10#digits)): base 10, wrapping on overflow as
// bash's 64-bit arithmetic does; "" is 0.
func bashDecimal(s string) int64 {
	var n int64
	for _, c := range s {
		n = n*10 + int64(c-'0')
	}
	return n
}
