package guard

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/render-slot-prompt.sh shim (KAN-1058).
//
// flow-guard render-slot-prompt — render one review-panel dispatch's brief
// (skills/flow/review-panel.md, **Bundled dispatch**).
//
// Usage: flow-guard render-slot-prompt <skill-dir> <round> <canonical-wt> <plan-dir> <slot>[+<slot>...] \
//          -diff final|late-fix|delta|fix-round [-no-bundle] [-standard <path>...] \
//          [-fix-report <path>...] -- <worktree>...
//
// <skill-dir> is the directory review-panel.md was read from; <worktree>...
// is the change's resolved set, the canonical worktree first. Every path is
// absolute: the subagent resolves what it is given from its own directory.
//
// Writes <canonical-wt>/.superpowers/sdd/slot-prompt-<round>-<slot>[+<slot>...].md,
// atomically, and prints its path. In order:
//   1. the shared blocks, each extracted verbatim by its `> **<LABEL>:**`
//      blockquote from <skill-dir>/review-panel.md (the first one — the
//      slot's): TOOLS, FOREGROUND BUILDS, NO DELEGATION, REPRODUCE, DON'T
//      READ and WORKTREES (the resolved set filled in); CITATION CHECK once
//      per worktree whose .superpowers/sdd/citation-check.md exists; ENTRY
//      CONTEXT followed by the diff's .touched list (flow-guard write-panel-diff) —
//      each pass's own on a delta bundle; INDEPENDENT PASSES when more than
//      one slot; FIX-ROUND SCOPE from review-panel-fix-round.md, naming the
//      -fix-report path(s), only when -fix-report is given; on any -diff
//      but final, every `final-review.diff` in WORKTREES, ENTRY CONTEXT and
//      INDEPENDENT PASSES reads `the **Diff file:** each PASS section names`;
//   2. one `## PASS <id>` section per slot, in the order given: the fenced
//      body of <skill-dir>/<id>-reviewer-prompt.md — an exp-<name> slot's
//      <skill-dir>/experimental/<name>.md (review-panel-experimental-slot.md)
//      — dedented, every placeholder filled per review-panel.md's
//      placeholder table (**The roster**), then that slot's REPORT FILE block.
//
// The diff each pass reads, in <canonical-wt>/.superpowers/sdd/:
//   final      final-review.diff
//   late-fix   late-fix.diff
//   fix-round  fix-round-<round>.diff
//   delta      slot-delta-<round>-<id>.diff
// -no-bundle is the CONTEXT BUNDLE FAILURE continue path: [CONTEXT_BUNDLE_PATHS]
// becomes `none — the bundle was not built`, [GLOBAL_CONSTRAINTS] the plan's
// design.md (or `none`). -standard paths are [STANDARDS_PATHS]; empty when
// none are given.
//
// `mutation` has no template: its `## PASS mutation` section is the body of
// review-panel.md's `### The mutation-testing brief`, one line naming its
// throwaway copies — `<worktree>-mutation-<round>` per worktree given — and
// its REPORT FILE block. It reads no diff file, so a mutation-alone render
// carries no ENTRY CONTEXT and no touched list; review-panel.md carrying no
// brief is exit 2. MUTATION ENTRY CONTEXT and its tests, MODEL HANDSHAKE,
// CONTEXT BUNDLE, the relocation-comparison pointer and the reproducer rule
// stay typed in the Agent call.
//
// Exit codes:
//   0  written; its path on stdout
//   1  a placeholder is left unfilled — each named on stderr; nothing written
//   2  cannot answer — usage, a relative path, an unknown slot, a
//      missing template, block, touched list, standards, principles or
//      calibration file

// renderSlotPrompt is `flow-guard render-slot-prompt`: the comment above is the contract -- one panel dispatch's brief rendered from
// review-panel.md's shared blocks and each slot's template, written to
// <canonical-wt>/.superpowers/sdd/slot-prompt-<round>-<slots>.md and its
// path printed; exit 0 written, 1 a placeholder left unfilled (named, and
// nothing written), 2 cannot answer.
func init() {
	Registry["render-slot-prompt"] = renderSlotPrompt
}

const rspUsage = "usage: flow-guard render-slot-prompt <skill-dir> <round> <canonical-wt> <plan-dir> <slot>[+<slot>...] \\\n" +
	"         -diff final|late-fix|delta|fix-round [-no-bundle] [-standard <path>...] [-fix-report <path>...] \\\n" +
	"         -- <worktree>...\n"

var (
	// rspUnfilled is what a render must no longer carry: a template
	// placeholder, or a block's own path/round/id/report placeholder.
	rspUnfilled = regexp.MustCompile(`\[[A-Z][A-Z0-9_]*\]|<abs-worktree[^>]*>|<fix report>|<round>|<id>`)
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
	// mutation has no template: its PASS section is review-panel.md's
	// mutation-testing brief and the throwaway copies it mutates in
	// (kan-1029's self-review: an unrendered mutation pass never ran). roles
	// are the template slots, the ones reading a diff file.
	slots := strings.Split(slotArg, "+")
	var roles []string
	for _, s := range slots {
		switch {
		case s == "mutation":
		// Any other slot is valid when its template exists -- <id>-reviewer-prompt.md,
		// or experimental/<name>.md for exp-<name> -- so a slot added to the skill
		// directory renders without a rebuild of this guard.
		case !strings.Contains(s, "/") && (len(s) > len("exp-") && strings.HasPrefix(s, "exp-") || isFile(skill+"/"+s+"-reviewer-prompt.md")):
			roles = append(roles, s)
		default:
			return refuse("no reviewer template for slot %q", s)
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
	// repro is the REPRODUCER block as emitted: its `<round>`/`<id>` name
	// the file the slot itself fills in, so the unfilled check skips it.
	var repro string
	for _, label := range []string{"TOOLS", "FOREGROUND BUILDS", "NO DELEGATION", "REPRODUCE, DON'T READ", "REPRODUCER", "WORKTREES"} {
		b, ok := block(panel, "review-panel.md", label)
		if !ok {
			return 2
		}
		// A slot is told its round, not left to guess it: kan-1024's round-0
		// reproducers took round 1's names and were overwritten.
		if label == "REPRODUCER" {
			const phrase = "`<round>` this round's number"
			if !strings.Contains(b, phrase) {
				fmt.Fprintf(stderr, "render-slot-prompt: review-panel.md's REPRODUCER block no longer carries %q — cannot state the round\n", phrase)
				return 2
			}
			b = strings.Replace(b, phrase, phrase+" — `"+round+"` here", 1)
			repro = named(b)
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
	// The mutating pass runs last, whatever position the slot argument gives it
	// (review-panel-optional-slots.md: mutation is always a bundle's last pass).
	ordered := slices.DeleteFunc(slices.Clone(slots), func(s string) bool { return s == "mutation" })
	if len(ordered) < len(slots) {
		ordered = append(ordered, "mutation")
	}
	for _, id := range ordered {
		if id == "mutation" {
			brief := rspSection(panel, "### The mutation-testing brief")
			if brief == "" {
				return refuse("review-panel.md carries no mutation-testing brief")
			}
			copies := make([]string, len(wts))
			for i, wt := range wts {
				copies[i] = wt + "-mutation-" + round
			}
			emit("## PASS mutation")
			emit(brief)
			emit("Mutate and run the tests only in your throwaway copies, one per repository — `" +
				strings.Join(copies, "`, `") + "` — never in the shared worktrees above.")
			emit(strings.NewReplacer("<abs-worktree>", canon, "<round>", round, "<id>", id).Replace(report))
			continue
		}
		tmplPath := skill + "/" + id + "-reviewer-prompt.md"
		if name, ok := strings.CutPrefix(id, "exp-"); ok {
			tmplPath = skill + "/experimental/" + name + ".md"
		}
		tmpl, ok := read(tmplPath)
		if !ok {
			return 2
		}
		body, found := rspFenced(tmpl)
		if !found {
			return refuse("%s carries no fenced template body", tmplPath)
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
	if left := rspUnfilled.FindAllString(strings.Replace(rendered, repro, "", 1), -1); len(left) > 0 {
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

// rspSection is the body under heading in src: the lines after it up to the
// next Markdown heading, trimmed; "" when src has no such heading.
func rspSection(src, heading string) string {
	_, after, found := strings.Cut(src, "\n"+heading+"\n")
	if !found {
		return ""
	}
	var out []string
	for _, l := range strings.Split(after, "\n") {
		if strings.HasPrefix(l, "#") {
			break
		}
		out = append(out, l)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
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
