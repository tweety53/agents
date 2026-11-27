package guard

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// flow-guard build-slot-prompt <skill-dir> <abs-worktree> <round> <id[+id]>
// -diff <path> [-artifacts <p,…>] [-standards <p,…>] [-global-constraints <path|none>]
// [-worktrees <p,…>] [-repo <name>] [-fix-report <p,…>] [-panel <slot,…>] — assembles one panel
// dispatch's brief (skills/flow/review-panel.md, **Bundled dispatch**) so the
// parent never reads a reviewer template. It writes
// <abs-worktree>/.superpowers/sdd/slot-prompt-<round>-<ids>[-<repo>].md:
// the slot-brief read line, a PANEL SLOTS line naming every slot the round
// runs (-panel), a WORKTREES line (-worktrees), a SECTION line (-repo), the FIX-ROUND SCOPE blockquote of review-panel-fix-round.md with
// <fix report> filled (-fix-report), then per id a `## PASS <id>` section — the
// fenced body of <skill-dir>/<id>-reviewer-prompt.md, dedented, its
// placeholders filled ([CALIBRATION_PATH] and [PRINCIPLES_PATH] resolved
// beside the templates) — and its REPORT FILE line. Prints the written path.
// Exit 0 written; 2 usage, an id with no template (mutation has none), or a
// placeholder a template carries that no flag filled.
func init() { Registry["build-slot-prompt"] = buildSlotPrompt }

var bspPlaceholder = regexp.MustCompile(`\[(DIFF_PATH|ARTIFACT_PATHS|CALIBRATION_PATH|PRINCIPLES_PATH|STANDARDS_PATHS|GLOBAL_CONSTRAINTS)\]`)

func buildSlotPrompt(args []string, env Env, stdout, stderr io.Writer) int {
	usage := func(msg string) int {
		Reason(env, "usage")
		fmt.Fprintln(stderr, "build-slot-prompt: "+msg)
		fmt.Fprintln(stderr, "usage: flow-guard build-slot-prompt <skill-dir> <abs-worktree> <round> <id[+id]> -diff <path> [-artifacts <p,…>] [-standards <p,…>] [-global-constraints <path|none>] [-worktrees <p,…>] [-repo <name>] [-fix-report <p,…>] [-panel <slot,…>]")
		return 2
	}
	if len(args) < 4 {
		return usage("missing positional arguments")
	}
	skillDir, wt, ids := args[0], args[1], strings.Split(args[3], "+")
	round, err := strconv.Atoi(args[2])
	if err != nil {
		return usage("round is not a number")
	}
	fs := flag.NewFlagSet("build-slot-prompt", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fill := map[string]*string{
		"DIFF_PATH":          fs.String("diff", "", ""),
		"ARTIFACT_PATHS":     fs.String("artifacts", "", ""),
		"STANDARDS_PATHS":    fs.String("standards", "", ""),
		"GLOBAL_CONSTRAINTS": fs.String("global-constraints", "", ""),
	}
	worktrees := fs.String("worktrees", "", "")
	repo := fs.String("repo", "", "")
	fixReport := fs.String("fix-report", "", "")
	panel := fs.String("panel", "", "")
	if err := fs.Parse(args[4:]); err != nil || fs.NArg() != 0 {
		return usage("bad flags")
	}
	set := map[string]bool{"CALIBRATION_PATH": true, "PRINCIPLES_PATH": true}
	names := map[string]string{"diff": "DIFF_PATH", "artifacts": "ARTIFACT_PATHS", "standards": "STANDARDS_PATHS", "global-constraints": "GLOBAL_CONSTRAINTS"}
	fs.Visit(func(f *flag.Flag) { set[names[f.Name]] = true })
	cal := filepath.Join(skillDir, "reviewer-calibration.md")
	prin := filepath.Join(skillDir, "engineering-principles.md")
	fill["CALIBRATION_PATH"], fill["PRINCIPLES_PATH"] = &cal, &prin
	list := func(s string) string { return strings.ReplaceAll(s, ",", ", ") }

	suffix := ""
	if *repo != "" {
		suffix = "-" + *repo
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Read %s in full first — every paragraph binds you\n\n", filepath.Join(skillDir, "slot-brief.md"))
	if *panel != "" {
		fmt.Fprintf(&b, "PANEL SLOTS: %s\n\n", list(*panel))
	}
	if *worktrees != "" {
		fmt.Fprintf(&b, "WORKTREES: %s\n\n", list(*worktrees))
	}
	if *repo != "" {
		fmt.Fprintf(&b, "SECTION: %s\n\n", *repo)
	}
	if *fixReport != "" {
		src, err := os.ReadFile(filepath.Join(skillDir, "review-panel-fix-round.md"))
		quote := bspBlockquote(string(src), "> **FIX-ROUND SCOPE:**")
		if err != nil || quote == "" {
			return usage("no FIX-ROUND SCOPE blockquote in review-panel-fix-round.md")
		}
		b.WriteString(strings.ReplaceAll(quote, "<fix report>", list(*fixReport)) + "\n")
	}
	sdd := filepath.Join(wt, ".superpowers/sdd")
	for _, id := range ids {
		src, err := os.ReadFile(filepath.Join(skillDir, id+"-reviewer-prompt.md"))
		if err != nil {
			return usage(fmt.Sprintf("%s has no reviewer template (mutation's brief is The mutation-testing brief)", id))
		}
		body := bspFencedBody(string(src))
		var missing string
		body = bspPlaceholder.ReplaceAllStringFunc(body, func(m string) string {
			k := m[1 : len(m)-1]
			if !set[k] {
				missing = m
				return m
			}
			return list(*fill[k])
		})
		if missing != "" {
			return usage(fmt.Sprintf("%s's template carries %s and no flag fills it", id, missing))
		}
		fmt.Fprintf(&b, "## PASS %s\n\n%s\nREPORT FILE: %s\n\n", id, body,
			filepath.Join(sdd, fmt.Sprintf("panel-report-%d-%s%s.md", round, id, suffix)))
	}
	dst := filepath.Join(sdd, fmt.Sprintf("slot-prompt-%d-%s%s.md", round, args[3], suffix))
	if err := os.MkdirAll(sdd, 0o755); err != nil {
		return usage(err.Error())
	}
	if err := os.WriteFile(dst, []byte(b.String()), 0o644); err != nil {
		return usage(err.Error())
	}
	fmt.Fprintln(stdout, dst)
	return 0
}

// bspFencedBody is the first ``` fence's body, its common indentation removed.
func bspFencedBody(src string) string {
	var body []string
	in := false
	for _, l := range strings.Split(src, "\n") {
		if strings.HasPrefix(l, "```") {
			if in {
				break
			}
			in = true
			continue
		}
		if in {
			body = append(body, l)
		}
	}
	indent := -1
	for _, l := range body {
		if strings.TrimSpace(l) == "" {
			continue
		}
		if n := len(l) - len(strings.TrimLeft(l, " ")); indent < 0 || n < indent {
			indent = n
		}
	}
	var out strings.Builder
	for _, l := range body {
		if len(l) >= indent && indent > 0 {
			l = l[indent:]
		}
		out.WriteString(strings.TrimRight(l, " ") + "\n")
	}
	return out.String()
}

// bspBlockquote is the blockquote opening with head, through its last `>` line.
func bspBlockquote(src, head string) string {
	i := strings.Index(src, head)
	if i < 0 {
		return ""
	}
	var out strings.Builder
	for _, l := range strings.Split(src[i:], "\n") {
		if !strings.HasPrefix(l, ">") {
			break
		}
		out.WriteString(l + "\n")
	}
	return out.String()
}
