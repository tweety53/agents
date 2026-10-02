package guard

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

// checkPanelReproducerExitContract is
// scripts/check-panel-reproducer-exit-contract.sh: that script's header
// comment is the contract -- exit 0 every open finding's runnable reproducer
// demonstrated, 1 violations, 2 cannot answer (outranking 1). The runner is
// runReproducer called in-process, never a second implementation of it; the
// findings read is Env.Findings when set, else `flow record findings`. The
// reasoning for each branch, moved here from the bash body it replaced
// (0747740), sits beside the code it explains.
//
// The bash guard resolved its runner as a sibling in scripts/, never from
// PATH, so the guard under test in its own harness sandbox picked up the
// sandbox's stub, and checked its readability before the first call — a
// missing runner being "cannot answer at all", never a verdict. Calling
// runReproducer in-process has no file to resolve or find missing.
func init() {
	Registry["check-panel-reproducer-exit-contract"] = checkPanelReproducerExitContract
}

const (
	pcPrefix         = "check-panel-reproducer-exit-contract: "
	pcReadOK         = 0x4 // access(2)'s R_OK
	pcDeclare        = "# demonstrates: "
	pcPremiseDeclare = "# premise: "
	pcJQFailed       = pcPrefix + "jq failed — cannot determine anything"
	pcNotFound       = "flow: command not found"
	pcNotFoundC      = 127
)

// pcFinding is one `flow record findings` object as jq -r reads it.
type pcFinding struct {
	ref        *string // nil unless .ref is a JSON string: only then can `select(.ref == $ref)` match it
	refText    string  // `jq -r .ref`
	status     string
	statusOpen bool // .status == "open"
	reproducer string
	noRepro    bool // .reproducer == null or ""
}

func checkPanelReproducerExitContract(args []string, env Env, stdout, stderr io.Writer) int {
	// The bash guard's `export LC_ALL=C` reached every reproducer it ran.
	cEnv := env
	cEnv.Getenv = func(k string) string {
		if k == "LC_ALL" {
			return "C"
		}
		return env.Getenv(k)
	}
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
		fmt.Fprintf(stderr, "%snot a directory: %s\n", pcPrefix, shown)
		return 2
	}
	if name == "" {
		fmt.Fprintln(stderr, "usage: check-panel-reproducer-exit-contract.sh <worktree> <change-name>")
		return 2
	}
	// CONTAINMENT, the same allowlist check-panel-reproducers and
	// check-unfinished-work apply through plainChangeName: the change name reaches this guard from state a
	// pull request can edit and is passed to `flow record findings -change`,
	// so the same shapes are refused here, and
	// TestCheckPanelReproducerExitContract asserts the same rejected list so
	// the copies cannot drift apart silently. plainChangeName compares bytes,
	// as the bash's `export LC_ALL=C` made its enumeration do; the worktree is
	// canonicalised (EvalSymlinks, the bash's `cd ... && pwd -P`) before it is
	// ever handed to the runner, for the same dash-prefixed-relative-path and
	// symlinked-TMPDIR reasons both siblings record.
	if !plainChangeName(name) {
		fmt.Fprintf(stderr, "%schange name '%s' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n", pcPrefix, name)
		return 2
	}
	worktree, err := filepath.EvalSymlinks(pcAbs(env, worktreeArg))
	if err != nil {
		fmt.Fprintf(stderr, "%sworktree vanished before it could be resolved: %s\n", pcPrefix, worktreeArg)
		return 2
	}

	// THE STATE RECORD IS THE CHANGE'S WORKTREES BOOTSTRAP (KAN-658). The
	// store addresses records by project and change name together, and the
	// project key resolves from the worktree argument — so on a cross-repo
	// change only the canonical worktree resolves the record at all, and a
	// peer worktree's findings read below would answer `[]` at exit 0: a clean
	// verdict on a change this invocation never actually saw. The record is
	// read FIRST for exactly that reason, and this block is DUPLICATED, on
	// purpose, in panelreproducers.go, whose store read has the same blind
	// spot; TestCheckPanelReproducers and TestCheckPanelReproducerExitContract
	// assert the same refused shapes, which is what keeps the copies from
	// drifting. `flow state get` reached-and-absent exits 1 — a fact about this
	// project/change pair, and on a cross-repo change the signature of a guard
	// invoked on a peer tree — reported at this guard's own exit 2, never a
	// clean answer. Store-unreachable exits 0 with the fallback record only
	// when one exists, so an empty or non-JSON stdout at exit 0 is the same
	// cannot-answer class the findings read below already gives an unreachable
	// store. The `worktrees` map this record carries keys every worktree of
	// the change by absolute path; it is what the per-finding resolution below
	// resolves a reproducer's tree from.
	//
	// The CLI's own stderr rides along in the refusal, so a missing `flow`
	// binary or a dead daemon is reported with its evidence instead of being
	// flattened into the cross-repo prose (panel finding F3, kan-658 round 0)
	// — the same shape its sibling guard carries.
	stateOut, stateErr, stateRC := pcFlow(env, false, "state", "get", "-C", worktree, name)
	flowSaid := ""
	if s := strings.ReplaceAll(string(stateErr), "\n", " "); s != "" {
		flowSaid = " — flow said: " + s
	}
	if stateRC != 0 {
		if stateRC == 126 || stateRC == 127 {
			// The CLI never ran — reporting the store fact would be a guess
			// wearing evidence's clothes (panel finding F3, kan-658 round 0).
			fmt.Fprintf(stderr, "%scannot run the flow CLI (exit %d%s) — cannot determine anything\n", pcPrefix, stateRC, flowSaid)
		} else {
			fmt.Fprintf(stderr, "%sthe store has no record of change '%s' under this worktree's project (flow exit %d%s) — a cross-repo change's guards answer only through its canonical worktree\n", pcPrefix, name, stateRC, flowSaid)
		}
		return 2
	}
	var state map[string]json.RawMessage
	if json.Unmarshal(bytes.TrimRight(stateOut, "\n"), &state) != nil || state == nil {
		fmt.Fprintf(stderr, "%scannot read the state record for '%s' — cannot determine anything%s\n", pcPrefix, name, flowSaid)
		return 2
	}
	recorded, ok := pcWorktreeKeys(state["worktrees"])
	if !ok {
		fmt.Fprintln(stderr, pcJQFailed)
		return 2
	}
	// THE PREFIX INDEX (KAN-795). A cross-repo change's panel prefixes a
	// finding's citation with the basename of the worktree it was raised in —
	// `gymie-frontend:src/Foo.tsx:42`, the WORKTREES paragraph's form
	// (skills/flow/review-panel.md) — and the audit below resolves that prefix
	// to its tree through this index: the canonical worktree plus every
	// recorded worktrees-map entry, keyed by the basename each resolves to.
	// Map entries get the same treatment the per-finding token search gives
	// them — an empty or vanished path is skipped like an absent one, and
	// EvalSymlinks gives each the physical shape the containment comparison
	// answers in — and two entries naming one physical tree count once, so an
	// unambiguous prefix is never refused as ambiguous (panel finding F4,
	// kan-658 round 0). Built once, before the findings loop, so every
	// declaration of the run resolves through the same map.
	prefixTrees := map[string][]string{}
	addPrefixTree := func(resolved string) {
		if resolved == "" {
			return
		}
		b := filepath.Base(resolved)
		if !slices.Contains(prefixTrees[b], resolved) {
			prefixTrees[b] = append(prefixTrees[b], resolved)
		}
	}
	addPrefixTree(worktree)
	for _, rw := range recorded {
		if rw == "" || !isDir(pcAbs(env, rw)) {
			continue
		}
		if p, err := filepath.EvalSymlinks(pcAbs(env, rw)); err == nil {
			addPrefixTree(p)
		}
	}

	// THE STORE IS QUERIED ONCE, and a non-zero exit from `flow record findings`
	// is this guard's own exit 2 — the same reading its sibling guard gives the
	// same call, for the same reason: a read that could not be performed is
	// never a clean answer and never a violation.
	var findingsJSON []byte
	if env.Findings != nil {
		findingsJSON, err = env.Findings(name)
	} else {
		var rc int
		findingsJSON, _, rc = pcFlow(env, true, "record", "findings", "-change", name, "-C", worktree)
		if rc != 0 {
			err = fmt.Errorf("flow exit %d", rc)
		}
	}
	findingsText := strings.TrimRight(string(findingsJSON), "\n")
	if err != nil {
		fmt.Fprintf(stderr, "%scannot read findings for '%s' from the store — cannot determine anything: %s\n", pcPrefix, name, findingsText)
		return 2
	}
	findings, ok := pcParseFindings([]byte(findingsText))
	if !ok {
		fmt.Fprintln(stderr, pcJQFailed)
		return 2
	}

	var violations []string
	cannotAnswer := ""
	cannot := func(msg string) {
		if cannotAnswer == "" {
			cannotAnswer = msg
		}
	}
	runnable := 0
	// A finding carrying NO reproducer field at all is cannot-answer (exit 2),
	// not a skip: the sibling write-time validation makes the shape unreachable
	// in practice, and assuming the impossibility holds would quietly drop an
	// open finding's claim from this gate. Checked once, up front, scoped to
	// OPEN findings — findings at any other status are skipped wholesale below,
	// so whatever reproducer field they carry or lack claims nothing here.
	var empty []string
	for _, f := range findings {
		if f.statusOpen && f.noRepro {
			empty = append(empty, f.refText)
		}
	}
	if len(empty) > 0 {
		cannot(fmt.Sprintf("finding(s) %s carry no reproducer field at all", strings.Join(empty, " ")))
	}

	for _, f := range findings {
		ref := f.refText
		// `select(.ref == $ref) | .status` (and `.reproducer`): every
		// finding carrying this ref, one line each.
		var statuses, repros []string
		for _, g := range findings {
			if g.ref != nil && *g.ref == ref {
				statuses = append(statuses, g.status)
				repros = append(repros, g.reproducer)
			}
		}
		// ONLY AN OPEN FINDING CLAIMS THE CURRENT TREE. Every other status is
		// skipped before the reproducer is even classified: running a fixed
		// finding's reproducer and demanding "demonstrated" would require the
		// defect this same pipeline fixed to still be present.
		if strings.Join(statuses, "\n") != "open" {
			continue
		}
		// The two `none` forms are skipped, not run: the bare form is the
		// lexical guard's violation and that guard runs first in this
		// pipeline, and the exemption form claims nothing runnable by
		// construction. One word-boundary guard covers both, since the
		// exemption form also begins `none` + a space; unlike
		// check-panel-reproducers.sh's two-branch shape, no branch here
		// diverges. A command that merely STARTS with the letters `none`
		// (`nonexistent-script`) matches neither pattern and is a runnable
		// command line.
		reproducer := strings.Join(repros, "\n")
		if reproducer == "none" || strings.HasPrefix(reproducer, "none") && strings.ContainsAny(reproducer[4:5], " \t\n\v\f\r") {
			continue
		}

		// THE INSTRUMENT AUDIT (KAN-606), before the runner is ever invoked:
		// a verdict spent on an instrument whose citation does not resolve is
		// the green flip this guard exists to deny, so an audit failure is
		// recorded and the run is skipped for that finding (pcAudit). The
		// first-10-lines window and the exact-line mutation declaration are
		// the KAN-568 convention's own. The path token is derived the way
		// runReproducer's own tokenizer derives it — split on space and tab,
		// first element — never `${reproducer%% *}`: that bash idiom split on
		// a literal space only, so a tab-separated command line legal
		// everywhere else in the pipeline would reach this audit as one token
		// and be bounced as unreadable (panel finding F1, kan-606 round 0).
		firstLine, _, _ := strings.Cut(reproducer, "\n")
		token := ""
		if fields := strings.FieldsFunc(firstLine, func(r rune) bool { return r == ' ' || r == '\t' }); len(fields) > 0 {
			token = fields[0]
		}
		// THE FINDING'S TREE, RESOLVED PER FINDING (KAN-658). The canonical
		// tree first — every single-repo change resolves here, exactly as this
		// guard always has — then the change's recorded worktrees: on a
		// cross-repo change a peer finding's reproducer script lives in ITS
		// repository's worktree, the only tree its relative path token and
		// its demonstrates citations resolve in, and running both against the
		// canonical tree was the could-not-be-read false failure that made
		// every affected reproducer a hand-run substitution. Exactly one
		// recorded worktree carrying the path resolves the tree; several
		// cannot be told apart, which is no verdict for this finding rather
		// than a guessed one; none is the existing unreadable class below,
		// unchanged. A recorded path that is not a directory on disk is
		// skipped like an absent one — the map records git worktrees, and a
		// vanished entry resolves nothing.
		tree := worktree
		if !pcExists(worktree + "/" + token) {
			var matches []string
			for _, rw := range recorded {
				if rw == "" || !isDir(pcAbs(env, rw)) {
					continue
				}
				// EvalSymlinks (the bash's `cd ... && pwd -P`) gives the
				// recorded path the same physical shape the worktree itself
				// carries — the containment check in pcAudit compares resolved
				// paths, and a map entry reached through a symlinked prefix
				// would otherwise lose every comparison on its shape alone.
				// Two map entries can name one physical tree through a
				// symlinked prefix; counting the alias twice would refuse an
				// unambiguous reproducer as ambiguous (panel finding F4,
				// kan-658 round 0).
				p, err := filepath.EvalSymlinks(pcAbs(env, rw))
				if err != nil || slices.Contains(matches, p) || !pcExists(p+"/"+token) {
					continue
				}
				matches = append(matches, p)
			}
			if len(matches) == 1 {
				tree = matches[0]
			} else if len(matches) > 1 {
				cannot(fmt.Sprintf("%s's reproducer path token '%s' resolves in %d of the change's recorded worktrees (%s) — the finding's tree is ambiguous, so no verdict is possible", ref, token, len(matches), strings.Join(matches, " ")))
				continue
			}
		}
		reproPath := tree + "/" + token
		if !isFile(reproPath) || syscall.Access(reproPath, pcReadOK) != nil {
			violations = append(violations, fmt.Sprintf("%s's reproducer script '%s' could not be read — its demonstrates declaration cannot be audited, so its claim cannot be checked", ref, token))
			continue
		}
		if !declaresMutation(reproPath) {
			// KAN-839: the premise audit joins the demonstrates audit under
			// the same mutation skip — a mutation-declared reproducer's
			// instrument audit is the sha pin, never this tree's premises.
			audit := append(pcAudit(ref, tree, reproPath, prefixTrees, cannot), pcPremiseAudit(ref, tree, reproPath, prefixTrees, cannot)...)
			// The skip the audit promises: any violation this finding's
			// declarations added means the runner is never invoked for it.
			if len(audit) > 0 {
				violations = append(violations, audit...)
				continue
			}
		}

		fmt.Fprintf(stderr, "%s%s — running %s\n", pcPrefix, ref, reproducer)
		var out bytes.Buffer
		rc := runReproducer([]string{tree, reproducer}, cEnv, &out, &out)
		said := func() {
			fmt.Fprintf(stderr, "%s%s — runner said: %s\n", pcPrefix, ref, strings.TrimRight(out.String(), "\n"))
		}
		switch rc {
		case 0:
			runnable++
			fmt.Fprintf(stderr, "%s%s — defect demonstrated, claim holds\n", pcPrefix, ref)
		case 1:
			violations = append(violations, ref+"'s reproducer read 'defect not demonstrated' on the tree under review — an open finding's reproducer must read demonstrated here under whichever exit-code convention it declares: a generic one demonstrates with a non-zero exit, a declared mutation-reproducer with exit 0 (the build succeeds with the mutation landed) — this inverted reading is the exit-code class this guard exists for (KAN-554)")
			said()
		case 2:
			violations = append(violations, ref+"'s reproducer was refused by run-reproducer.sh as unusable — its claim cannot hold on any tree")
			said()
		case 3:
			cannot(ref + "'s reproducer could not be verdicted — the runner reported a timeout or a surviving process, which is no verdict at all")
			said()
		default:
			cannot(ref + "'s reproducer could not be verdicted — the runner cannot answer")
			said()
		}
	}

	// CANNOT-ANSWER OUTRANKS VIOLATIONS. A run whose reads could not be
	// completed has produced no verdict for at least one finding, and a
	// partial verdict printed at exit 1 would read as a completed check.
	if cannotAnswer != "" {
		fmt.Fprintf(stderr, "%scannot determine anything — %s\n", pcPrefix, cannotAnswer)
		return 2
	}
	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintf(stderr, "%s%s\n", pcPrefix, v)
		}
		return 1
	}
	// The count is what the loop above actually ran and credited — never a
	// second re-derivation of the runnable classification, which would be a
	// second encoding of the same rule and the copy that drifts.
	fmt.Fprintf(stdout, "REPRODUCER-EXIT-CONTRACT-OK (%d runnable open finding(s) demonstrated)\n", runnable)
	return 0
}

// pcAudit is the KAN-606 instrument audit of one reproducer's
// `# demonstrates: <path>:<line>:<content>` declarations against tree: the
// violations it found, none when every citation resolves. A reproducer
// carrying no demonstrates declaration at all is itself a violation — what
// the instrument reads and expects is unaudited.
func pcAudit(ref, tree, reproPath string, prefixTrees map[string][]string, cannot func(string)) []string {
	var v []string
	decls := 0
	for _, l := range headLines(reproPath, 10) {
		if strings.HasPrefix(l, pcDeclare) {
			decls++
			v = append(v, pcResolveCitation(ref, tree, l, pcDeclare, "demonstrates", prefixTrees, cannot)...)
		}
	}
	if decls == 0 {
		return []string{ref + "'s reproducer carries no '# demonstrates: <path>:<line>:<content>' declaration within its first 10 lines — what the instrument reads and expects is unaudited"}
	}
	return v
}

// pcPremiseAudit is KAN-839's premise audit of one reproducer's
// `# premise: <path>:<line>:<content>` declarations: the same resolution
// pcAudit applies to its demonstrates citations, but tolerant — a reproducer
// carrying no premise line violates nothing, so panel records predating the
// premise rule are never re-bounced by it. A declared-but-unresolvable
// premise is a violation like any demonstrates miss: the premise names what
// the reproducer's checks read, and checks running against nothing are the
// vacuous pass this audit exists to deny.
func pcPremiseAudit(ref, tree, reproPath string, prefixTrees map[string][]string, cannot func(string)) []string {
	var v []string
	for _, l := range headLines(reproPath, 10) {
		if strings.HasPrefix(l, pcPremiseDeclare) {
			v = append(v, pcResolveCitation(ref, tree, l, pcPremiseDeclare, "premise", prefixTrees, cannot)...)
		}
	}
	return v
}

// pcResolveCitation resolves one declaration line against tree — the checks
// pcAudit has applied since KAN-606: the declaration's shape, the cited
// path's lexical and resolved containment under tree, the file, the line and
// the line's content — returning the violation it found, none when the
// citation resolves. `noun` names the declaration kind in every message
// ("demonstrates", "premise"), so the two audits share one resolution and
// cannot drift.
//
// THE PREFIXED FORM (KAN-795). The declaration's first field may name one of
// the change's trees by basename — `# demonstrates:
// gymie-frontend:src/Foo.tsx:42:content`, the cross-repo form the panel
// records already use (skills/flow/review-panel.md's WORKTREES paragraph).
// A basename that answers to exactly one indexed tree redirects the whole
// resolution — the file, line, content and containment checks, the prefix's
// tree substituted for the finding's own; an unprefixed citation and a
// prefix nothing answers to read exactly as KAN-606 left them. Several trees
// answering to one basename is no verdict for the finding rather than a
// guessed one — the same ambiguity class the reproducer path token resolves
// to cannot-answer at — so the audit records cannot-answer through `cannot`
// and returns a violation the exit-2 precedence keeps unprinted; the finding
// is skipped either way, the runner never invoked on an instrument no tree
// can be chosen for.
func pcResolveCitation(ref, tree, decl, prefix, noun string, prefixTrees map[string][]string, cannot func(string)) []string {
	rest := strings.TrimPrefix(decl, prefix)
	if head, remainder, ok := strings.Cut(rest, ":"); ok {
		if trees, match := prefixTrees[head]; match {
			switch {
			case len(trees) == 1:
				tree = trees[0]
				rest = remainder
			case len(trees) > 1:
				cannot(fmt.Sprintf("%s's reproducer %s declaration's prefix '%s' names %d of the change's worktrees (%s) — the citation's tree is ambiguous, so no verdict is possible", ref, noun, head, len(trees), strings.Join(trees, " ")))
				return []string{fmt.Sprintf("%s's reproducer %s declaration's prefix '%s' is ambiguous — %d worktrees carry that basename", ref, noun, head, len(trees))}
			}
		}
	}
	dpath, rest2, ok1 := strings.Cut(rest, ":")
	dline, dcontent, ok2 := strings.Cut(rest2, ":")
	if dpath == "" || !ok1 || !ok2 || dline == "" || strings.Trim(dline, "0123456789") != "" || dcontent == "" {
		return []string{fmt.Sprintf("%s's reproducer carries a malformed %s declaration ('%s') — the form is '%s<path>:<line>:<content>'", ref, noun, decl, prefix)}
	}
	if strings.HasPrefix(dpath, "/") || hasDotDotSegment(dpath) {
		return []string{fmt.Sprintf("%s's reproducer %s declaration cites '%s' — a citation names a path relative to and inside the worktree under review", ref, noun, dpath)}
	}
	target := tree + "/" + dpath
	if !isFile(target) {
		return []string{fmt.Sprintf("%s's reproducer %s declaration cites '%s' — the tree under review carries no such file", ref, noun, dpath)}
	}
	// Resolved containment, the runner's own pattern: EvalSymlinks (the
	// bash's `realpath`) follows `..`, `.` and symlinks in one step, so a
	// citation whose declared path carries no lexical `..` segment but
	// escapes through a symlink inside the worktree is caught here rather
	// than read outside the tree (panel finding F4, kan-606 round 0). tree
	// is the guard's own physical worktree or a recorded map path resolved
	// the same way, the shape EvalSymlinks answers in.
	resolved, err := filepath.EvalSymlinks(target)
	if err != nil || !strings.HasPrefix(resolved, tree+"/") {
		shown := resolved
		if err != nil || shown == "" {
			shown = "an unresolvable path"
		}
		return []string{fmt.Sprintf("%s's reproducer %s declaration cites '%s' — it resolves to '%s', outside the worktree under review — a symlink escape", ref, noun, dpath, shown)}
	}
	b, _ := os.ReadFile(resolved)
	lines := strings.SplitAfter(string(b), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	n, err := strconv.Atoi(dline)
	if err != nil || n < 1 || n > len(lines) {
		return []string{fmt.Sprintf("%s's reproducer %s declaration cites '%s':%s — past the end of the file", ref, noun, dpath, dline)}
	}
	if !strings.Contains(strings.TrimSuffix(lines[n-1], "\n"), dcontent) {
		return []string{fmt.Sprintf("%s's reproducer %s declaration cites content absent from '%s':%s — the instrument's citation does not resolve on this tree", ref, noun, dpath, dline)}
	}
	return nil
}

// pcFlow runs the flow CLI from env's PATH in env.Dir: stdout (with stderr
// folded in when combined, as `2>&1` does), stderr, and the exit code --
// 127 with bash's "command not found" when there is no flow to run.
func pcFlow(env Env, combined bool, args ...string) ([]byte, []byte, int) {
	flow, ok := lookPath(env, "flow")
	if !ok {
		return []byte(pcNotFound), []byte(pcNotFound + "\n"), pcNotFoundC
	}
	cmd := exec.Command(flow, args...)
	cmd.Dir = env.Dir
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if combined {
		cmd.Stderr = &out
	}
	err := cmd.Run()
	var ee *exec.ExitError
	switch {
	case err == nil:
		return out.Bytes(), errb.Bytes(), 0
	case errors.As(err, &ee):
		return out.Bytes(), errb.Bytes(), rrExitCode(ee.ProcessState)
	default:
		return out.Bytes(), []byte(err.Error() + "\n"), 126
	}
}

// pcWorktreeKeys is `(.worktrees // {}) | keys[]`: an object's keys sorted,
// none for null, false or absent; any other shape is jq failing.
func pcWorktreeKeys(raw json.RawMessage) ([]string, bool) {
	if s := strings.TrimSpace(string(raw)); s == "" || s == "null" || s == "false" {
		return nil, true
	}
	var m map[string]json.RawMessage
	if json.Unmarshal(raw, &m) != nil || m == nil {
		return nil, false
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, true
}

// pcParseFindings reads the findings array the way the bash guard's jq
// filters read it. Anything that is not one array of objects -- a null
// element, an object, empty output, two values -- is refused as jq failing,
// though jq itself iterated some of these and passed them: the guards'
// headers rule out a clean verdict on input they cannot read (decided
// 2026-09-27, kan-778's task 3 Correction).
func pcParseFindings(b []byte) ([]pcFinding, bool) {
	var raw []map[string]json.RawMessage
	if json.Unmarshal(b, &raw) != nil || raw == nil {
		return nil, false
	}
	out := make([]pcFinding, 0, len(raw))
	for _, o := range raw {
		if o == nil {
			return nil, false
		}
		f := pcFinding{refText: pcRaw(o["ref"]), status: pcRaw(o["status"]), reproducer: pcRaw(o["reproducer"])}
		var s string
		if json.Unmarshal(o["ref"], &s) == nil && pcIsString(o["ref"]) {
			f.ref = &s
		}
		f.statusOpen = pcIsString(o["status"]) && f.status == "open"
		f.noRepro = f.reproducer == "null" && !pcIsString(o["reproducer"]) || pcIsString(o["reproducer"]) && f.reproducer == ""
		out = append(out, f)
	}
	return out, true
}

func pcIsString(raw json.RawMessage) bool {
	return strings.HasPrefix(strings.TrimSpace(string(raw)), `"`)
}

// pcRaw is `jq -r` of one value: a string's text, null (or absent) as
// "null", anything else as compact JSON.
func pcRaw(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "null"
	}
	var s string
	if pcIsString(raw) && json.Unmarshal(raw, &s) == nil {
		return s
	}
	var c bytes.Buffer
	if json.Compact(&c, raw) != nil {
		return string(raw)
	}
	return c.String()
}

// pcExists is bash's `[ -e ]`: it follows symlinks.
func pcExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
