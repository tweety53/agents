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

// flow-guard fix-run-floor — the mechanical floor under a fix run's judged
// review-panel roster and verify mode, and the gate the judge's verdict
// passes through (skills/flow/review-panel-late-fix.md, The fix-run judge).
//
// Usage: flow-guard fix-run-floor <worktree> <merge-base>
//        flow-guard fix-run-floor -judge <verdict.json> -round <n> <worktree> <merge-base>
//
// Prints three lines on stdout:
//   class: prose|code|ui
//   floor_roster: primary|full
//   floor_verify: fast|full
//
// THE CLASS is read off this change's own touched paths -- the union of
// committed-since-merge-base, staged and unstaged, paneltouchedpaths.go's --
// by the classifiers the run already trusts, never a second copy of one:
//   ui     a touched path matches a `## visual verification` `ui paths` glob
//          (check-visual-trigger's match); a project declaring no such
//          section has no ui class
//   prose  not ui, check-panel-docs-only exits 0, and no `# inputs:` pathspec
//          of the project's `test` list (run-verify-list's own reading of
//          the list) is touched since <merge-base> -- documentation a suite
//          declares it reads is not prose
//   code   everything else, an empty touched-path set included
//
// THE FLOOR is the least a fix run of that class may run:
//   prose  floor_roster primary -- `primary` alone, no always-on extra
//          dispatch -- and floor_verify fast: the `lint` list, which carries
//          the project's citation check, and no `test` command
//   code   floor_roster full, floor_verify full
//   ui     floor_roster full, floor_verify full
//
// verify full is the whole `test` list, the default run -- never the full
// suite, which is run-verify-list -full and runs only when the operator asks.
//
// THE VERDICT. -judge reads the judge's JSON, {"roster": [<role id>...],
// "verify": "fast"|"full", "reasons": [<string>...]}, and admits it only at or
// above the floor: the class is prose, the roster names `primary`, verify is
// one of the two words and at least one reason is given. An admitted verdict
// is written, with the round, the merge base's sha and the class, to
// <worktree>/.superpowers/sdd/fix-run.json and printed as a fourth line,
//   verdict: roster <id>[,<id>...] · verify fast|full
// A class of code or ui has no room above its floor, so no verdict is
// admitted there. Every call first removes that file: a verdict governs the
// fix run that judged it and no later one, and a refused one governs nothing.
//
// WHO ENFORCES IT. flow-guard panel-round-begins refuses, in the verdict's own
// round, a slot carrying a role its roster omits; flow-guard run-verify-list
// skips every `test` command under verify fast. Both read the file through
// fixRunVerdict, which re-derives the class from the tree as it stands, so a
// verdict stops governing the moment the change is no longer prose; and
// run-verify-list drops fast once the store holds a Critical or Important
// finding of the verdict's round or a later one (frSevereSince).
//
// Exit codes:
//   0  the floor printed; with -judge, the verdict admitted and written
//   1  -judge only: the verdict is below the floor or malformed -- stderr
//      names why and nothing is written, so the run takes the full roster
//      and verify full
//   2  cannot answer: usage, <worktree> not a git worktree, <merge-base> not
//      resolving, a failed git invocation, an unreadable -judge file, or a
//      classifier that could not answer -- read as the full path

// fixRunFloor is `flow-guard fix-run-floor`: the comment above is the contract.
func init() { Registry["fix-run-floor"] = fixRunFloor }

const (
	frProg  = "fix-run-floor"
	frUsage = "usage: flow-guard fix-run-floor [-judge <verdict.json> -round <n>] <worktree> <merge-base>\n"
)

var frRole = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// frVerdict is fix-run.json: the judge's verdict as admitted, with the round
// it governs, the base its class was read against and that class.
type frVerdict struct {
	Round   int      `json:"round"`
	Base    string   `json:"base"`
	Class   string   `json:"class"`
	Roster  []string `json:"roster"`
	Verify  string   `json:"verify"`
	Reasons []string `json:"reasons"`
}

// frPath is the verdict file of worktree wt, an absolute path.
func frPath(wt string) string { return filepath.Join(wt, ".superpowers", "sdd", "fix-run.json") }

// frBelowFloor names why v sits below a prose floor or is malformed, "" when
// it is admitted.
func frBelowFloor(v frVerdict) string {
	switch {
	case len(v.Roster) == 0:
		return "the roster is empty"
	case slices.ContainsFunc(v.Roster, func(r string) bool { return !frRole.MatchString(r) }):
		return "the roster names something that is no role id"
	case !slices.Contains(v.Roster, "primary"):
		return "the roster omits primary, below floor_roster"
	case v.Verify != "fast" && v.Verify != "full":
		return fmt.Sprintf("verify is %q, neither fast nor full", v.Verify)
	case len(v.Reasons) == 0 || slices.Contains(v.Reasons, ""):
		return "no reason is given"
	}
	return ""
}

// frClass is the class of worktree wt (as the caller named it) against base:
// rc 0 with the class, rc 2 when a classifier could not answer, its cause on
// stderr.
func frClass(env Env, wt, base string, stderr io.Writer) (string, int) {
	gitBin, ok := panelResolveGit(env, frProg, stderr)
	if !ok || !panelValidateWorktree(env, frProg, wt, base, gitBin, stderr) {
		return "", 2
	}
	paths, ok := panelTouchedPaths(env, frProg, wt, base, gitBin, stderr)
	if !ok {
		return "", 2
	}
	var vtErr bytes.Buffer
	switch visualTrigger(env, wt, func() ([]string, error) { return paths, nil }, io.Discard, &vtErr) {
	case 0:
		return "ui", 0
	case 1:
	default:
		if !strings.Contains(vtErr.String(), vtNotConfigured+":") {
			fmt.Fprint(stderr, vtErr.String())
			return "", 2
		}
	}
	switch checkPanelDocsOnly([]string{wt, base}, env, io.Discard, stderr) {
	case 0:
	case 1:
		return "code", 0
	default:
		return "", 2
	}
	// Documentation a suite declares it reads is not prose: the `test`
	// list's `# inputs:` pathspecs, as run-verify-list reads them.
	var list bytes.Buffer
	switch projectGet([]string{wt, "test"}, env, &list, stderr) {
	case 0:
	case 1:
		return "prose", 0
	default:
		return "", 2
	}
	_, inputs, _ := rvlParse(list.String())
	for _, in := range inputs {
		specs := strings.Fields(in)
		if len(specs) == 0 {
			continue
		}
		err := panelGit(env, gitBin, wt, append([]string{"diff", "--quiet", "--end-of-options", base, "--"}, specs...)...).Run()
		switch exitCode(err) {
		case 0:
		case 1:
			return "code", 0
		default:
			fmt.Fprintf(stderr, "%s: cannot diff the `# inputs:` %s against %s in %s\n", frProg, strings.TrimSpace(in), base, wt)
			return "", 2
		}
	}
	return "prose", 0
}

func fixRunFloor(args []string, env Env, stdout, stderr io.Writer) int {
	judge, round := "", -1
	for len(args) >= 2 && (args[0] == "-judge" || args[0] == "-round") {
		if args[0] == "-judge" {
			judge = args[1]
		} else if n, err := strconv.Atoi(args[1]); err == nil && n >= 0 {
			round = n
		} else {
			fmt.Fprint(stderr, frUsage)
			return 2
		}
		args = args[2:]
	}
	if len(args) != 2 || (judge == "") != (round < 0) {
		fmt.Fprint(stderr, frUsage)
		return 2
	}
	wt, base := args[0], args[1]
	class, rc := frClass(env, wt, base, stderr)
	if rc != 0 {
		return rc
	}
	roster, verify := "full", "full"
	if class == "prose" {
		roster, verify = "primary", "fast"
	}
	fmt.Fprintf(stdout, "class: %s\nfloor_roster: %s\nfloor_verify: %s\n", class, roster, verify)

	// An earlier fix run's verdict never governs this one; a refused verdict
	// leaves none either.
	file := frPath(smcAbs(env, wt))
	os.Remove(file)
	if judge == "" {
		return 0
	}
	raw, err := os.ReadFile(smcAbs(env, judge))
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read the verdict: %v\n", frProg, err)
		return 2
	}
	v := frVerdict{Round: round, Class: class}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var in struct {
		Roster  []string `json:"roster"`
		Verify  string   `json:"verify"`
		Reasons []string `json:"reasons"`
	}
	if err := dec.Decode(&in); err != nil {
		fmt.Fprintf(stderr, "%s: the verdict is not {roster, verify, reasons}: %v\n", frProg, err)
		return 1
	}
	v.Roster, v.Verify, v.Reasons = in.Roster, in.Verify, in.Reasons
	if class != "prose" {
		fmt.Fprintf(stderr, "%s: class %s has no room above its floor — no verdict is admitted\n", frProg, class)
		return 1
	}
	if why := frBelowFloor(v); why != "" {
		fmt.Fprintf(stderr, "%s: verdict refused — %s\n", frProg, why)
		return 1
	}
	gitBin, _ := panelResolveGit(env, frProg, io.Discard)
	sha, ok := capture(panelGit(env, gitBin, wt, "rev-parse", "--verify", "--end-of-options", base+"^{commit}"))
	if !ok {
		fmt.Fprintf(stderr, "%s: merge base '%s' does not resolve in %s\n", frProg, base, wt)
		return 2
	}
	v.Base = sha
	out, _ := json.MarshalIndent(v, "", "  ")
	if err := os.MkdirAll(filepath.Dir(file), 0o777); err == nil {
		err = os.WriteFile(file, append(out, '\n'), 0o666)
	}
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot write %s: %v\n", frProg, file, err)
		return 2
	}
	fmt.Fprintf(stdout, "verdict: roster %s · verify %s\n", strings.Join(v.Roster, ","), v.Verify)
	return 0
}

// fixRunVerdict is the verdict governing worktree wt (an absolute path), ok
// false when there is none in force: no fix-run.json, an unreadable or
// below-floor one, or a tree whose class, re-derived against the recorded
// base, is no longer prose -- a verdict never outlives the floor it was
// admitted over.
func fixRunVerdict(env Env, wt string) (frVerdict, bool) {
	var v frVerdict
	raw, err := os.ReadFile(frPath(wt))
	if err != nil || json.Unmarshal(raw, &v) != nil || v.Base == "" || frBelowFloor(v) != "" {
		return v, false
	}
	class, rc := frClass(env, wt, v.Base, io.Discard)
	return v, rc == 0 && class == "prose"
}

// frSevereSince names the first Critical or Important finding the store
// holds for change in round or a later one -- what voids a relaxed verify
// for the rest of the run -- and "" when there is none. A read that fails is
// named too: an unanswered question never keeps a relaxation.
func frSevereSince(env Env, wt, change string, round int) string {
	var raw []byte
	if env.Findings != nil {
		b, err := env.Findings(change)
		if err != nil {
			return "the findings could not be read"
		}
		raw = b
	} else {
		b, _, rc := pcFlow(env, false, "record", "findings", "-change", change, "-C", wt)
		if rc != 0 {
			return "the findings could not be read"
		}
		raw = b
	}
	findings, ok := cfcParse(raw)
	if !ok {
		return "the findings could not be read"
	}
	for _, f := range findings {
		if f.round >= round && (strings.EqualFold(f.severity, "critical") || strings.EqualFold(f.severity, "important")) {
			return fmt.Sprintf("%s finding %s of round %d", f.severity, f.ref, f.round)
		}
	}
	return ""
}
