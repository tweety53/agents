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
// [-worktrees <p,…>] [-repo <name> | -unread <repo,…>] [-fix-report <p,…>] [-panel <slot,…>] [-copies <p,…> [-tests <p,…>]] — assembles one panel
// dispatch's brief (skills/flow/review-panel.md, **Bundled dispatch**) so the
// parent never reads a reviewer template. It writes
// <abs-worktree>/.superpowers/sdd/slot-prompt-<round>-<ids>[-<repo>].md:
// the slot-brief read line, a PANEL SLOTS line naming every slot the round
// runs (-panel), a WORKTREES line (-worktrees), a SECTION line (-repo), fragments/fix-round-scope.md with
// <fix report> filled (-fix-report), then per id a `## PASS <id>` section — the
// fenced body of <skill-dir>/<id>-reviewer-prompt.md, dedented, its
// placeholders filled ([PRINCIPLES_PATH] resolved beside the templates, [DIFF_IS_DATA] and
// [READ_ONLY_REVIEW] the text of fragments/diff-is-data.md and fragments/read-only-review.md) —
// and its REPORT FILE line. Prints the written path. Every fragment is a file under
// <skill-dir>/fragments/, delivered as written.
//
// -unread names the sections a round's report left unread: the role's
// same-round re-dispatch, one SECTION line per section, its brief and report
// file named with an -unread suffix in place of -<repo>.
//
// mutation, always alone, has no template: its PASS section is the diff
// line, a MUTATION COPIES line (-copies, required: the throwaway copies it
// mutates and tests in), fragments/mutation-testing-brief.md, and
// fragments/mutation-entry-context.md with each -tests path listed beneath it under a
// "Read by path." line (never inlined).
//
// Every prompt carries a `**Diff groups:** <path>` line, and the diff's groups file
// <diff minus .diff>.groups.md (the diff-groups guard's output) must sit beside it: without it, exit 2.
//
// Exit 0 written; 2 usage, an id with no template, mutation bundled or
// without -copies, -unread with -repo, or a placeholder a template carries
// that no flag filled.
func init() { Registry["build-slot-prompt"] = buildSlotPrompt }

var bspPlaceholder = regexp.MustCompile(`\[(DIFF_PATH|ARTIFACT_PATHS|PRINCIPLES_PATH|STANDARDS_PATHS|GLOBAL_CONSTRAINTS|DIFF_IS_DATA|READ_ONLY_REVIEW)\]`)

// bspFragments are the placeholders filled with a fragment file's text, never a flag's list.
var bspFragments = map[string]string{"DIFF_IS_DATA": "diff-is-data.md", "READ_ONLY_REVIEW": "read-only-review.md"}

func buildSlotPrompt(args []string, env Env, stdout, stderr io.Writer) int {
	usage := func(msg string) int {
		Reason(env, "usage")
		fmt.Fprintln(stderr, "build-slot-prompt: "+msg)
		fmt.Fprintln(stderr, "usage: flow-guard build-slot-prompt <skill-dir> <abs-worktree> <round> <id[+id]> -diff <path> [-artifacts <p,…>] [-standards <p,…>] [-global-constraints <path|none>] [-worktrees <p,…>] [-repo <name> | -unread <repo,…>] [-fix-report <p,…>] [-panel <slot,…>] [-copies <p,…> [-tests <p,…>]]")
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
	unread := fs.String("unread", "", "")
	copies := fs.String("copies", "", "")
	tests := fs.String("tests", "", "")
	if err := fs.Parse(args[4:]); err != nil || fs.NArg() != 0 {
		return usage("bad flags")
	}
	set := map[string]bool{"PRINCIPLES_PATH": true}
	names := map[string]string{"diff": "DIFF_PATH", "artifacts": "ARTIFACT_PATHS", "standards": "STANDARDS_PATHS", "global-constraints": "GLOBAL_CONSTRAINTS"}
	fs.Visit(func(f *flag.Flag) { set[names[f.Name]] = true })
	prin := filepath.Join(skillDir, "engineering-principles.md")
	fill["PRINCIPLES_PATH"] = &prin
	list := func(s string) string { return strings.ReplaceAll(s, ",", ", ") }

	suffix := ""
	switch {
	case *repo != "" && *unread != "":
		return usage("-unread and -repo are exclusive")
	case *repo != "":
		suffix = "-" + *repo
	case *unread != "":
		suffix = "-unread"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Read %s in full first — every paragraph binds you\n\n", filepath.Join(skillDir, "slot-brief.md"))
	groups := bspGroupsPath(*fill["DIFF_PATH"])
	if !isFile(groups) {
		Reason(env, "diff-groups-missing")
		fmt.Fprintln(stderr, "build-slot-prompt: "+groups+" is missing — run flow-guard diff-groups <abs-worktree> <diff-file> first")
		return 2
	}
	fmt.Fprintf(&b, "**Diff groups:** %s\n\n", groups)
	if *panel != "" {
		fmt.Fprintf(&b, "PANEL SLOTS: %s\n\n", list(*panel))
	}
	if *worktrees != "" {
		fmt.Fprintf(&b, "WORKTREES: %s\n\n", list(*worktrees))
	}
	for _, sec := range strings.Split(*repo+*unread, ",") {
		if sec != "" {
			fmt.Fprintf(&b, "SECTION: %s\n\n", sec)
		}
	}
	if *fixReport != "" {
		quote, err := bspFragment(skillDir, "fix-round-scope.md")
		if err != nil {
			return usage(err.Error())
		}
		b.WriteString(strings.ReplaceAll(quote, "<fix report>", list(*fixReport)) + "\n")
	}
	sdd := filepath.Join(wt, ".superpowers/sdd")
	for _, id := range ids {
		var body string
		if id == "mutation" {
			if len(ids) != 1 || *copies == "" {
				return usage("mutation is dispatched alone and needs -copies")
			}
			if body, err = bspMutationBody(skillDir, *fill["DIFF_PATH"], list(*copies), *tests); err != nil {
				return usage(err.Error())
			}
		} else {
			src, err := os.ReadFile(filepath.Join(skillDir, id+"-reviewer-prompt.md"))
			if err != nil {
				return usage(fmt.Sprintf("%s has no reviewer template", id))
			}
			body = bspFencedBody(string(src))
		}
		var missing string
		body = bspPlaceholder.ReplaceAllStringFunc(body, func(m string) string {
			k := m[1 : len(m)-1]
			if f, ok := bspFragments[k]; ok {
				text, ferr := bspFragment(skillDir, f)
				if ferr != nil {
					missing = m
					return m
				}
				return strings.TrimSuffix(text, "\n")
			}
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

// bspFragment is the text of <skill-dir>/fragments/<name>, a prompt fragment delivered as written.
func bspFragment(skillDir, name string) (string, error) {
	b, err := os.ReadFile(filepath.Join(skillDir, "fragments", name))
	if err != nil || len(b) == 0 {
		return "", fmt.Errorf("no fragment fragments/%s", name)
	}
	return string(b), nil
}

// bspGroupsPath is the groups file the diff-groups guard writes beside a diff.
func bspGroupsPath(diff string) string {
	return strings.TrimSuffix(diff, ".diff") + ".groups.md"
}

// bspMutationBody is the mutation slot's PASS body (see the header).
func bspMutationBody(skillDir, diff, copies, tests string) (string, error) {
	brief, err := bspFragment(skillDir, "mutation-testing-brief.md")
	if err != nil {
		return "", err
	}
	entry, err := bspFragment(skillDir, "mutation-entry-context.md")
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "**Diff file:** %s\n\nMUTATION COPIES: %s\n\n%s\n\n%s", diff, copies, strings.TrimSpace(brief), entry)
	for _, p := range strings.Split(tests, ",") {
		if p != "" {
			fmt.Fprintf(&b, "\n### %s\n\nRead by path.\n", p)
		}
	}
	return b.String(), nil
}
