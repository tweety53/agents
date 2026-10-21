package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

// checkVisualVerifyDispatched is scripts/check-visual-verify-dispatched.sh:
// that script's header comment is the contract -- one verdict line on
// stdout, VISUAL-VERIFY-OK (exit 0) or VISUAL-VERIFY-MISSING (exit 1), or
// exit 2 when it cannot answer. The trigger question is visualTrigger
// (visualtrigger.go), evaluated in-process; the dispatch read
// is Env.Dispatches when set, else `flow record dispatches`, and the
// decisions read Env.Decisions or `flow record decisions`; the verdict
// write and the prior-false-positive read are Env.Verdict and Env.Verdicts
// the same way. The reasoning for each branch, moved here from the bash body
// it replaced (2056def4, KAN-809's last commit to it), sits beside the code
// it explains.
//
// The bash needed jq and refused at exit 2 without it; the rows are decoded
// with encoding/json here, so there is no tool left to find missing.
func init() {
	Registry["check-visual-verify-dispatched"] = checkVisualVerifyDispatched
}

const vvdPrefix = "check-visual-verify-dispatched: "

func checkVisualVerifyDispatched(args []string, env Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	worktreeArg, name, base, canonicalArg := arg(0), arg(1), arg(2), arg(3)
	if worktreeArg == "" || !isDir(pcAbs(env, worktreeArg)) {
		shown := worktreeArg
		if shown == "" {
			shown = "<missing>"
		}
		fmt.Fprintf(stderr, "%snot a directory: %s\n", vvdPrefix, shown)
		return 2
	}
	if name == "" {
		fmt.Fprintf(stderr, "%susage: check-visual-verify-dispatched.sh <worktree> <change-name> <merge-base> [<canonical-worktree>]\n", vvdPrefix)
		return 2
	}
	if base == "" {
		fmt.Fprintf(stderr, "%sa merge-base is required — pass the state file's recorded value for this worktree\n", vvdPrefix)
		return 2
	}

	// CONTAINMENT, duplicated on purpose from check-panel-fix-single-
	// dispatch's copy: plainChangeName (changename.go), whose comment is
	// canonical for the reasoning.
	if !plainChangeName(name) {
		fmt.Fprintf(stderr, "%schange name '%s' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n", vvdPrefix, name)
		return 2
	}

	// The bash's `cd "$WORKTREE" && pwd -P`, refused as
	// panelfixsingledispatch.go refuses it, whose comment carries the
	// reasoning for each refusal and for the line naming no path.
	worktree, err := filepath.EvalSymlinks(pcAbs(env, worktreeArg))
	if err != nil || strings.HasPrefix(worktreeArg, "-") || syscall.Access(pcAbs(env, worktreeArg), 1) != nil {
		fmt.Fprintf(stderr, "%sworktree vanished: \n", vvdPrefix)
		return 2
	}

	// Every git invocation whose failure would otherwise be read as an
	// answer is captured and checked on its own, never piped straight into
	// the trigger question -- the same discipline check-base-moved's comment
	// documents for the identical hazard. The bash exported LC_ALL=C, so git
	// runs under it too; visualTrigger's matching is C-locale already.
	git := envGit(env, "LC_ALL=C")("-C", worktree, "diff", "--no-renames", "--name-only", base+"..HEAD")
	combined, gitErr := git.CombinedOutput()
	changed := strings.TrimRight(strings.ReplaceAll(string(combined), "\x00", ""), "\n")
	if gitErr != nil {
		fmt.Fprintf(stderr, "%sgit diff failed against merge-base '%s' in %s — cannot answer: %s\n", vvdPrefix, base, worktree, changed)
		return 2
	}

	// THE TRIGGER QUESTION IS ASKED IN-PROCESS, NOT RE-IMPLEMENTED: the
	// diff's paths go to visualTrigger (an empty diff is one empty line, as
	// the bash's `printf '%s\n'` piped it), its stdout and stderr discarded,
	// its three exit codes read as-is.
	//
	// A SATELLITE IS ALSO MATCHED AGAINST THE CANONICAL `ui paths` (kan-982
	// self-review): a cross-repo change's frontend worktree may declare no
	// section of its own while the canonical project names its files as
	// `<repo>/**`, so asking each worktree alone passed integrate with every
	// verifier blocked. Its paths are prefixed with its repository's name --
	// its main checkout's basename, check-visual-trigger.sh's --repo -- and
	// the better of the two answers stands: 0 over 1 over 2. The store is
	// read under the canonical worktree, the project the run's rows were
	// recorded in; a second repository's worktree resolves another project
	// key (unfinishedwork.go's anchor carries the reasoning).
	anchor, canonical := worktree, ""
	if canonicalArg != "" {
		c, err := filepath.EvalSymlinks(pcAbs(env, canonicalArg))
		if err != nil || !isDir(c) {
			fmt.Fprintf(stderr, "%scanonical worktree is not a directory: %s\n", vvdPrefix, canonicalArg)
			return 2
		}
		if c != worktree {
			anchor, canonical = c, c
		}
	}
	// ask is the trigger's answer for paths: 0, 1 or 2, or -1 when git
	// cannot name the satellite's repository (reported, exit 2).
	ask := func(paths []string) int {
		exit := visualTrigger(env, worktree, func() ([]string, error) { return paths, nil }, io.Discard, io.Discard)
		if canonical == "" || exit == 0 {
			return exit
		}
		common, ok := capture(envGit(env)("-C", worktree, "rev-parse", "--path-format=absolute", "--git-common-dir"))
		if !ok {
			fmt.Fprintf(stderr, "%sgit cannot name the repository of %s — cannot answer\n", vvdPrefix, worktree)
			return -1
		}
		repo := filepath.Base(filepath.Dir(strings.TrimSpace(common))) + "/"
		prefixed := make([]string, len(paths))
		for i, p := range paths {
			if p != "" {
				p = repo + p
			}
			prefixed[i] = p
		}
		switch visualTrigger(env, canonical, func() ([]string, error) { return prefixed, nil }, io.Discard, io.Discard) {
		case 0:
			return 0
		case 1:
			return 1
		}
		return exit
	}
	// A CHANGE'S OWN RECORDS ARE NOT UI WORK (kan-970 self-review): a
	// satellite's every path matches a canonical `<repo>/**`, so the archive
	// commit moving its `<spec root>/changes/<name>/link.md` read as a UI
	// change after the verified HEAD and forced a re-verify of nothing.
	// Paths under `<spec root>/changes/` are dropped before both the trigger
	// question and the freshness intersection, as unfinishedwork.go's range
	// walk drops them.
	records := specRootLeaf(worktree, io.Discard) + "/changes/"
	var paths []string
	for _, p := range strings.Split(changed, "\n") {
		if !strings.HasPrefix(p, records) {
			paths = append(paths, p)
		}
	}
	if len(paths) == 0 {
		paths = []string{""}
	}
	triggerExit := ask(paths)
	if triggerExit < 0 {
		return 2
	}
	switch triggerExit {
	case 2:
		// The project declares no `## visual verification` section at all.
		fmt.Fprintln(stdout, "VISUAL-VERIFY-OK: not configured")
		return 0
	case 1:
		// Declared, but this diff touched none of its `ui paths`.
		fmt.Fprintln(stdout, "VISUAL-VERIFY-OK: no UI paths touched")
		return 0
	}

	// A store the call could not reach is "cannot answer", never "no rows":
	// an outage that read as zero dispatches would pass every round it was
	// blind to.
	rows, err := pfdRead(env, env.Dispatches, "dispatches", name, anchor)
	if err != nil {
		fmt.Fprintf(stderr, "%sflow record dispatches failed for '%s' — cannot answer\n", vvdPrefix, name)
		return 2
	}
	heads, ok := vvdCompletedVerifiers(rows)
	if !ok {
		fmt.Fprintf(stderr, "%sdispatch rows were not readable JSON — cannot answer\n", vvdPrefix)
		return 2
	}
	// A VERIFIER COUNTS ONLY FOR THE HEAD IT VERIFIED (kan-939 self-review):
	// a completed row once answered OK however many UI commits landed after
	// it. Each row's `diffBase` is that HEAD (visual-verify.md's
	// **Recording**); the row counts when git resolves it in this worktree
	// and none of this change's own touched paths that changed since is a
	// UI path. Intersecting with the change's own paths keeps a base the
	// branch was rebased onto from reading as this change's UI work. A row
	// with no `diffBase`, or one this worktree cannot resolve -- another
	// worktree's verifier -- verified nothing here.
	own := map[string]bool{}
	for _, p := range paths {
		own[p] = true
	}
	found := false
	// The newest resolvable verified HEAD (heads arrive in row seq order,
	// oldest first) and its own paths changed since: what a MISSING re-run
	// is scoped to.
	var lastHead string
	var lastChanged []string
	// A canonical verifier that also covered a sibling worktree names each
	// HEAD it verified, comma-joined (kan-983 self-review): each sha is judged
	// on its own, and only the one this worktree resolves can count here.
	var shas []string
	for _, head := range heads {
		shas = append(shas, strings.Split(head, ",")...)
	}
	for _, head := range shas {
		if !vvdSHA.MatchString(head) {
			continue
		}
		if _, ok := capture(envGit(env)("-C", worktree, "rev-parse", "--verify", "--quiet", head+"^{commit}")); !ok {
			continue
		}
		since, ok := capture(envGit(env, "LC_ALL=C")("-C", worktree, "diff", "--no-renames", "--name-only", head, "HEAD"))
		if !ok {
			fmt.Fprintf(stderr, "%sgit diff failed against verified HEAD '%s' in %s — cannot answer\n", vvdPrefix, head, worktree)
			return 2
		}
		var changedSince []string
		for _, p := range strings.Split(since, "\n") {
			if own[p] && p != "" {
				changedSince = append(changedSince, p)
			}
		}
		lastHead, lastChanged = head, changedSince
		if len(changedSince) == 0 {
			changedSince = []string{""}
		}
		exit := ask(changedSince)
		if exit < 0 {
			return 2
		}
		if exit != 0 {
			found = true
			break
		}
	}

	// THE VERDICT IS RECORDED, AND PRIOR FALSE POSITIVES ARE ADVISORY -- the
	// habit unfinishedwork.go carries. The write's result is discarded: a
	// store outage can neither move the verdict nor change the exit code, and
	// a store that refuses the row leaves the gate exactly as it printed it.
	// The row is what lets an operator's `flow record verdict false-positive`
	// land at all (KAN-809's second break). The hint is read on MISSING only,
	// skipped on any failure or an empty array.
	const guardName = "check-visual-verify-dispatched"
	record := func(verdict string) {
		if env.Verdict != nil {
			_ = env.Verdict(name, guardName, worktree, verdict)
			return
		}
		pcFlow(env, false, "record", "verdict", "-change", name, "-guard", guardName, "-worktree", worktree, "-verdict", verdict, "-C", anchor)
	}
	// THE VERIFIER'S RECORD IS PART OF ITS EVIDENCE (kan-970 self-review): a
	// verifier that wrote visual-verification.md outside the change root,
	// citing session-scratch captures, passed on its dispatch row alone and
	// left the self-review bundle nothing to read. A found row counts only
	// with `<spec root>/changes/<name>/visual-verification.md`, or its
	// archived twin, in the worktree the store is read under -- the change
	// root the verifier writes (skills/flow/visual-verify-verifier.md step 11).
	missingRecord := ""
	if found {
		changes := anchor + "/" + specRootLeaf(anchor, io.Discard) + "/changes/"
		if !isFile(changes+name+"/visual-verification.md") && !isFile(changes+"archive/"+name+"/visual-verification.md") {
			found, missingRecord = false, changes+name+"/visual-verification.md"
		}
	}
	if found {
		verdict := "VISUAL-VERIFY-OK: UI paths touched and a completed (or ended-but-unclosed) verifier dispatch is recorded for '" + name + "'"
		record(verdict)
		fmt.Fprintln(stdout, verdict)
		return 0
	}

	// NO VERIFIER, BUT THE NEWEST DECISION SKIPPED IT AT DECIDE: the
	// operator saw that row and its reason at the plan gate, so the skip is
	// not silent (kan-30's failure) and is not this guard's to second-guess.
	// The read fails closed exactly as the dispatch read does.
	decisions, err := pfdRead(env, env.Decisions, "decisions", name, anchor)
	if err != nil {
		fmt.Fprintf(stderr, "%sflow record decisions failed for '%s' — cannot answer\n", vvdPrefix, name)
		return 2
	}
	reason, skipped, ok := vvdSkippedAtDecide(decisions)
	if !ok {
		fmt.Fprintf(stderr, "%sdecision rows were not readable JSON — cannot answer\n", vvdPrefix)
		return 2
	}
	if skipped && missingRecord == "" {
		verdict := "VISUAL-VERIFY-OK: skipped at Decide — " + reason
		record(verdict)
		fmt.Fprintln(stdout, verdict)
		return 0
	}
	verdict := "VISUAL-VERIFY-MISSING: this change's diff touched a declared UI path but no completed or ended 'verifier' dispatch (key starting 'visual-verify') is recorded for '" + name + "' — flow.visual-verify's stage marks were written with no verifier ever dispatched, its report was never read to completion, or the dispatch's closing outcome was lost to a session restart"
	if missingRecord != "" {
		verdict = "VISUAL-VERIFY-MISSING: a verifier verified this change's HEAD but its record " + missingRecord + " (or its archive/ twin) is absent — the verifier writes it into the change root, every file it cites copied under visual-verification/evidence/ (skills/flow/visual-verify-verifier.md step 11)"
		lastHead = ""
	} else if len(heads) > 0 {
		verdict = "VISUAL-VERIFY-MISSING: a verifier dispatch is recorded for '" + name + "' but none verified a HEAD after this change's last UI-path commit — a row counts only with the -diff-base it verified, and only when no UI path this change touches changed since"
	}
	// THE MISSING VERDICT NAMES THE HEAD IT JUDGED (kan-939 self-review):
	// integrate's hard stop passes only on an operator's false-positive flag
	// on this worktree's row for this very HEAD, so a flag never covers a
	// commit made after it.
	headSHA, ok := capture(envGit(env)("-C", worktree, "rev-parse", "HEAD"))
	if !ok {
		fmt.Fprintf(stderr, "%sgit cannot name HEAD in %s — cannot answer\n", vvdPrefix, worktree)
		return 2
	}
	if lastHead != "" {
		verdict += " — paths changed since verified HEAD " + lastHead + ": " + strings.Join(lastChanged, ", ")
	}
	verdict += " — HEAD " + headSHA
	record(verdict)
	var prior []byte
	var priorErr error
	if env.Verdicts != nil {
		prior, priorErr = env.Verdicts(guardName)
	} else if out, _, rc := pcFlow(env, false, "record", "verdicts", "-guard", guardName, "-false-positive", "-C", anchor); rc == 0 {
		prior = out
	} else {
		priorErr = fmt.Errorf("flow exit %d", rc)
	}
	if priorErr == nil {
		if n, last, ok := vvdPriorFalsePositives(prior); ok {
			fmt.Fprintf(stderr, "%sprior false positives for this guard on this project: %d — last: %s\n", vvdPrefix, n, last)
		}
	}
	fmt.Fprintln(stdout, verdict)
	return 1
}

// vvdSkippedAtDecide reads `flow record decisions`' answer, newest row
// first: skipped only when that row's decision records `visual.verify`
// exactly `skipped` with a non-empty string `reason`. Every other readable
// shape -- no rows (a change decided before the field existed), no `visual`,
// `required`, `not configured`, a malformed value -- is not skipped, so the
// guard falls through to MISSING as it always did. !ok is cannot-answer:
// not one JSON array, or a newest row that is not an object.
func vvdSkippedAtDecide(raw []byte) (string, bool, bool) {
	var rows []json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil || rows == nil {
		return "", false, false
	}
	if len(rows) == 0 {
		return "", false, true
	}
	var row struct {
		Decision struct {
			Visual struct {
				Verify string `json:"verify"`
				Reason string `json:"reason"`
			} `json:"visual"`
		} `json:"decision"`
	}
	if !isJSONObject(rows[0]) {
		return "", false, false
	}
	if json.Unmarshal(rows[0], &row) != nil {
		// A row whose decision or visual is not the shape above is not a skip.
		return "", false, true
	}
	v := row.Decision.Visual
	// One verdict line: a reason spanning lines is folded onto it.
	reason := strings.Join(strings.Fields(v.Reason), " ")
	return reason, v.Verify == "skipped" && reason != "", true
}

// isJSONObject reports whether raw holds a JSON object.
func isJSONObject(raw json.RawMessage) bool {
	t := bytes.TrimSpace(raw)
	return len(t) > 0 && t[0] == '{'
}

// vvdPriorFalsePositives is prior_false_positives_hint's two jq reads over
// the `flow record verdicts` answer, a stream of JSON values as jq reads it.
// The count is `if type == "array" then length else empty end`: ok only when
// the stream parses and exactly one value is an array of length >= 1 (no
// output, or one line per array, fails the bash's digits-only check). last
// is `(.[0].falsePositiveReason // "") + " (" + (.[0].change // "") + ", " +
// ((.[0].flaggedAt // "") | tostring | .[0:10]) + ")"` over every value,
// the lines joined as jq -r printed them. A value jq errors on prints no
// line; jq's exit status is the last value's alone, so an error there
// renders the whole of last empty, as `|| vv_last=""` did, and an error
// anywhere earlier only drops that value's line.
//
// tostring is pcRaw's rendering, as uwPriorFalsePositives uses: jq 1.7
// re-renders a number literal in decNumber form (1e2 as 1E+2) and
// re-escapes a string nested in an array or object, where pcRaw keeps the
// literal -- divergent only for a non-string flaggedAt, which the verb
// never serves (it prints an RFC 3339 string or null).
func vvdPriorFalsePositives(raw []byte) (int, string, bool) {
	dec := json.NewDecoder(bytes.NewReader(bytes.ReplaceAll(raw, []byte{0}, nil)))
	var vals []json.RawMessage
	for {
		var v json.RawMessage
		if err := dec.Decode(&v); err == io.EOF {
			break
		} else if err != nil {
			return 0, "", false
		}
		vals = append(vals, v)
	}
	n, arrays := 0, 0
	for _, v := range vals {
		var a []json.RawMessage
		if json.Unmarshal(v, &a) == nil && a != nil {
			n, arrays = len(a), arrays+1
		}
	}
	if arrays != 1 || n < 1 {
		return 0, "", false
	}
	lines := make([]string, 0, len(vals))
	for i, v := range vals {
		line, ok := vvdLastLine(v)
		if !ok && i == len(vals)-1 {
			return n, "", true
		}
		if ok {
			lines = append(lines, line)
		}
	}
	return n, strings.TrimRight(strings.Join(lines, "\n"), "\n"), true
}

// vvdLastLine is the last-entry expression over one value; false where jq
// errors: `.[0]` of anything but an array or null, a field of anything but
// an object or null, or `+` on a reason or change that is not a string once
// `// ""` has replaced null and false.
func vvdLastLine(v json.RawMessage) (string, bool) {
	isNull := func(r json.RawMessage) bool { return len(r) == 0 || string(r) == "null" }
	first := v
	if !isNull(v) {
		var a []json.RawMessage
		if json.Unmarshal(v, &a) != nil || a == nil {
			return "", false
		}
		first = nil
		if len(a) > 0 {
			first = a[0]
		}
	}
	field := map[string]json.RawMessage{}
	if !isNull(first) && (json.Unmarshal(first, &field) != nil || field == nil) {
		return "", false
	}
	alt := func(r json.RawMessage) json.RawMessage { // jq's `// ""`
		if isNull(r) || string(r) == "false" {
			return json.RawMessage(`""`)
		}
		return r
	}
	var reason, change string
	if json.Unmarshal(alt(field["falsePositiveReason"]), &reason) != nil || json.Unmarshal(alt(field["change"]), &change) != nil {
		return "", false
	}
	flagged := []rune(pcRaw(alt(field["flaggedAt"])))
	flagged = flagged[:min(len(flagged), 10)]
	return reason + " (" + change + ", " + string(flagged) + ")", true
}

// vvdCompletedVerifiers is the `diffBase` of every row the bash's `[.[] |
// select((.role // "") == "verifier" and ((.key // "") |
// startswith("visual-verify")) and ((.outcome // "") == "completed" or
// ((.endedAt != null) and ((.outcome // "") == ""))))]` selects, "" for a
// row that carries none. A row qualifies when its role is `verifier`,
// its key starts with the stage's canonical prefix (the key's exact shape is
// a different guard's job), and it carries one of two evidence shapes: its
// outcome is exactly `completed`, or -- the session-continuation signature
// (KAN-809) -- it carries an end instant and no outcome, the verifier having
// ended while the closing `-outcome completed` call was lost to an exhausted
// session. An explicit `aborted`, `fallback` or `blocked` outcome is never
// evidence, end instant or not, and a still-open row (no endedAt) is not
// either: the guard cannot tell a running verifier from a dead one. jq's
// `and` short-circuits, so the key is only read on a verifier row; a null
// row selects nothing.
//
// !ok is the header's cannot-answer, for every payload the verb never
// prints: not one JSON array (the bash's `jq empty` passed an empty answer
// or several values, and its `[ -ge 1 ]` then failed on the count and fell
// through to MISSING at exit 1), a non-null row that is not an object, or a
// verifier row whose key is neither a string nor null/false (jq failing, the
// bash exiting 5 under set -e). A top-level object is refused as
// pfdPanelFixKeys's comment explains.
func vvdCompletedVerifiers(raw []byte) ([]string, bool) {
	vals, ok := pfdStream(raw)
	if !ok || len(vals) != 1 {
		return nil, false
	}
	arr, ok := vals[0].([]any)
	if !ok {
		return nil, false
	}
	var heads []string
	for _, e := range arr {
		if e == nil {
			continue
		}
		row, ok := e.(map[string]any)
		if !ok {
			return nil, false
		}
		if row["role"] != "verifier" {
			continue
		}
		var key string
		switch k := row["key"].(type) {
		case nil:
		case bool:
			if k {
				return nil, false
			}
		case string:
			key = k
		default:
			return nil, false
		}
		// `// ""` reads null, false and absent as the empty string.
		outcome := row["outcome"]
		if outcome == nil || outcome == false {
			outcome = ""
		}
		if strings.HasPrefix(key, "visual-verify") && (outcome == "completed" || (row["endedAt"] != nil && outcome == "")) {
			head, _ := row["diffBase"].(string)
			heads = append(heads, head)
		}
	}
	return heads, true
}

// vvdSHA is a commit id as `git rev-parse HEAD` prints it, abbreviated or
// whole -- never a ref name such as HEAD, which would verify whatever the
// worktree holds when the guard runs.
var vvdSHA = regexp.MustCompile(`^[0-9a-f]{7,40}$`)
