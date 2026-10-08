package guard

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// renderSlotPrompt is scripts/render-slot-prompt.sh: that script's header
// comment is the contract -- one panel dispatch's brief rendered from
// review-panel.md's shared blocks and each slot's template, written to
// <canonical-wt>/.superpowers/sdd/slot-prompt-<round>-<slots>.md and its
// path printed; exit 0 written, 1 a placeholder left unfilled (named, and
// nothing written), 2 cannot answer.
func init() {
	Registry["render-slot-prompt"] = renderSlotPrompt
}

const rspUsage = "usage: render-slot-prompt.sh <skill-dir> <round> <canonical-wt> <plan-dir> <slot>[+<slot>...] \\\n" +
	"         -diff final|late-fix|delta|fix-round [-no-bundle] [-standard <path>...] [-fix-report <path>...] \\\n" +
	"         -- <worktree>...\n"

var (
	// rspUnfilled is what a render must no longer carry: a template
	// placeholder, or a block's own path/round/id/report placeholder.
	rspUnfilled = regexp.MustCompile(`\[[A-Z][A-Z0-9_]*\]|<abs-worktree[^>]*>|<fix report>|<round>|<id>`)
	// rspSlots are the slots with a template; mutation's brief stays typed.
	rspSlots = map[string]bool{"primary": true, "principles": true, "failure-modes": true}
)

func renderSlotPrompt(args []string, env Env, stdout, stderr io.Writer) int {
	refuse := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "render-slot-prompt: "+format+"\n", a...)
		return 2
	}
	usage := func() int {
		fmt.Fprint(stderr, rspUsage)
		return 2
	}
	if len(args) < 5 {
		return usage()
	}
	skill, round, canon, plan, slotArg := args[0], args[1], args[2], args[3], args[4]
	var kind string
	var noBundle bool
	var standards, reports, wts []string
	rest := args[5:]
	for len(rest) > 0 {
		a := rest[0]
		rest = rest[1:]
		// values takes the flag's arguments up to the next flag.
		values := func() []string {
			n := 0
			for n < len(rest) && !strings.HasPrefix(rest[n], "-") {
				n++
			}
			v := rest[:n]
			rest = rest[n:]
			return v
		}
		switch a {
		case "-diff":
			if len(rest) == 0 {
				return usage()
			}
			kind, rest = rest[0], rest[1:]
		case "-no-bundle":
			noBundle = true
		case "-standard":
			standards = append(standards, values()...)
		case "-fix-report":
			reports = append(reports, values()...)
		case "--":
			wts, rest = rest, nil
		default:
			return usage()
		}
	}
	if !wpdNumber.MatchString(round) || len(wts) == 0 {
		return usage()
	}
	for _, p := range append(append(append([]string{skill, canon, plan}, wts...), standards...), reports...) {
		if !strings.HasPrefix(p, "/") {
			return refuse("%q is not an absolute path — a subagent resolves every path it is given from its own directory", p)
		}
	}
	// mutation's brief and throwaway copy stay typed by the parent, but a
	// bundle naming it still gets every other role's pass, the shared
	// INDEPENDENT PASSES paragraph and a file name carrying the whole bundle;
	// a mutation-alone dispatch renders the shared paragraphs and no PASS
	// section — the parent types the brief beside the rendered path.
	slots := strings.Split(slotArg, "+")
	var roles []string
	for _, s := range slots {
		switch {
		case s == "mutation":
		case !rspSlots[s]:
			return refuse("no reviewer template for slot %q", s)
		default:
			roles = append(roles, s)
		}
	}

	sdd := canon + "/.superpowers/sdd/"
	diffPath := func(id string) string {
		switch kind {
		case "final":
			return sdd + "final-review.diff"
		case "late-fix":
			return sdd + "late-fix.diff"
		case "fix-round":
			return sdd + "fix-round-" + round + ".diff"
		case "delta":
			return sdd + "slot-delta-" + round + "-" + id + ".diff"
		}
		return ""
	}
	switch kind {
	case "final", "late-fix", "fix-round", "delta":
	default:
		return usage()
	}
	read := func(p string) (string, bool) {
		b, err := os.ReadFile(p)
		if err != nil {
			fmt.Fprintf(stderr, "render-slot-prompt: cannot read %s — cannot determine anything\n", p)
		}
		return string(b), err == nil
	}
	panel, ok := read(skill + "/review-panel.md")
	if !ok {
		return 2
	}
	block := func(src, file, label string) (string, bool) {
		b := rspBlock(src, label)
		if b == "" {
			fmt.Fprintf(stderr, "render-slot-prompt: %s carries no %s block\n", file, label)
		}
		return b, b != ""
	}
	for _, p := range standards {
		if !isFile(p) {
			return refuse("standards file %s does not exist", p)
		}
	}

	var out strings.Builder
	emit := func(s string) { out.WriteString(s + "\n\n") }
	// The shared blocks are written for pass 1; a pass reading any other
	// diff is pointed at the one its PASS section names, never the branch.
	named := func(b string) string {
		if kind == "final" {
			return b
		}
		return strings.ReplaceAll(b, "`final-review.diff`", "the **Diff file:** each PASS section names")
	}
	for _, label := range []string{"TOOLS", "FOREGROUND BUILDS", "NO DELEGATION", "REPRODUCE, DON'T READ", "WORKTREES"} {
		b, ok := block(panel, "review-panel.md", label)
		if !ok {
			return 2
		}
		if label == "WORKTREES" {
			b = strings.Replace(b, "`<abs-worktree-1>`, `<abs-worktree-2>`, …", "`"+strings.Join(wts, "`, `")+"`", 1)
		}
		emit(named(b))
	}
	for _, wt := range wts {
		if isFile(wt + "/.superpowers/sdd/citation-check.md") {
			b, ok := block(panel, "review-panel.md", "CITATION CHECK")
			if !ok {
				return 2
			}
			emit(strings.Replace(b, "<abs-worktree>", wt, 1))
		}
	}
	// Entry context is a diff-reading concept: mutation reads no diff file,
	// so a mutation-alone render carries no ENTRY CONTEXT and no touched list.
	if len(roles) > 0 {
		entry, ok := block(panel, "review-panel.md", "ENTRY CONTEXT")
		if !ok {
			return 2
		}
		emit(named(entry))
	}
	// One touched list per distinct diff; a delta bundle names each pass's.
	seen := map[string]bool{}
	for _, id := range roles {
		d := diffPath(id)
		if seen[d] {
			continue
		}
		seen[d] = true
		touched, ok := read(d + ".touched")
		if !ok {
			return 2
		}
		if kind == "delta" && len(roles) > 1 {
			emit("PASS `" + id + "` touched files:")
		}
		emit(strings.TrimRight(touched, "\n"))
	}
	if len(slots) > 1 {
		b, ok := block(panel, "review-panel.md", "INDEPENDENT PASSES")
		if !ok {
			return 2
		}
		emit(named(b))
	}
	if len(reports) > 0 {
		fixRound, ok := read(skill + "/review-panel-fix-round.md")
		if !ok {
			return 2
		}
		b, ok := block(fixRound, "review-panel-fix-round.md", "FIX-ROUND SCOPE")
		if !ok {
			return 2
		}
		emit(strings.Replace(b, "<fix report>", strings.Join(reports, "`, `"), 1))
	}

	// The placeholder table's values (review-panel.md, **The roster**).
	bundles := make([]string, len(wts))
	for i, wt := range wts {
		bundles[i] = wt + "/.superpowers/sdd/dispatch-context.md"
	}
	fill := map[string]string{
		"[ARTIFACT_PATHS]":       plan + "/proposal.md, " + plan + "/design.md, " + plan + "/tasks.md",
		"[CALIBRATION_PATH]":     skill + "/reviewer-calibration.md",
		"[CONTEXT_BUNDLE_PATHS]": strings.Join(bundles, ", "),
		"[PRINCIPLES_PATH]":      skill + "/engineering-principles.md",
		"[STANDARDS_PATHS]":      strings.Join(standards, ", "),
		"[GLOBAL_CONSTRAINTS]":   "the design.md section of your context bundle",
	}
	if noBundle {
		fill["[CONTEXT_BUNDLE_PATHS]"] = "none — the bundle was not built"
		fill["[GLOBAL_CONSTRAINTS]"] = "none"
		if isFile(plan + "/design.md") {
			fill["[GLOBAL_CONSTRAINTS]"] = plan + "/design.md"
		}
	}
	report, ok := block(panel, "review-panel.md", "REPORT FILE")
	if !ok {
		return 2
	}
	for _, id := range roles {
		tmpl, ok := read(skill + "/" + id + "-reviewer-prompt.md")
		if !ok {
			return 2
		}
		body, found := rspFenced(tmpl)
		if !found {
			return refuse("%s-reviewer-prompt.md carries no fenced template body", id)
		}
		fill["[DIFF_PATH]"] = diffPath(id)
		for _, p := range rspUnfilled.FindAllString(body, -1) {
			v, known := fill[p]
			if !known {
				continue
			}
			// A path the subagent is told to read must exist.
			if (p == "[CALIBRATION_PATH]" || p == "[PRINCIPLES_PATH]") && !isFile(v) {
				return refuse("%s does not exist — never dispatch a blind reviewer", v)
			}
			body = strings.ReplaceAll(body, p, v)
		}
		emit("## PASS " + id)
		emit(strings.TrimRight(body, "\n"))
		r := strings.NewReplacer("<abs-worktree>", canon, "<round>", round, "<id>", id).Replace(report)
		emit(r)
	}

	rendered := strings.TrimRight(out.String(), "\n") + "\n"
	if left := rspUnfilled.FindAllString(rendered, -1); len(left) > 0 {
		fmt.Fprintf(stderr, "render-slot-prompt: unresolved placeholder(s): %s\n", strings.Join(left, ", "))
		return 1
	}
	path := sdd + "slot-prompt-" + round + "-" + slotArg + ".md"
	if err := os.MkdirAll(sdd, 0o755); err != nil {
		return refuse("%v", err)
	}
	if err := wpdWriteAtomic(path, []byte(rendered)); err != nil {
		return refuse("%v", err)
	}
	fmt.Fprintln(stdout, path)
	return 0
}

// rspBlock is the blockquote in src opening `> **<label>:**`: that line
// through the last consecutive `>` line; "" when src has none. The first
// one is the slot's: review-panel.md states each slot paragraph before the
// fix subagent's.
func rspBlock(src, label string) string {
	var out []string
	for _, l := range strings.Split(src, "\n") {
		if len(out) == 0 && !strings.HasPrefix(l, "> **"+label+":**") {
			continue
		}
		if !strings.HasPrefix(l, ">") {
			break
		}
		out = append(out, l)
	}
	return strings.Join(out, "\n")
}

// rspFenced is a template's first fenced body, each line's four-space
// indent removed.
func rspFenced(src string) (string, bool) {
	lines := strings.Split(src, "\n")
	start := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "```") {
			if start < 0 {
				start = i
				continue
			}
			body := lines[start+1 : i]
			for j, b := range body {
				body[j] = strings.TrimPrefix(b, "    ")
			}
			return strings.Join(body, "\n"), true
		}
	}
	return "", false
}
