package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/tweety53/agents/stats/internal/prompttuning"
)

// check-prompt-tuning-cases.sh [dir]: is every prompt-tuning case folder
// complete and generic? Root is dir, else <FLOW_GUARD_REPO_ROOT>/docs/prompt-tuning;
// an absent root is clean. Per subdirectory (design.md §3 of
// kan-1032-prompt-tuning-angle): every prompttuning.RequiredFiles entry
// present; case.json parses with every field set; type, reviewer_kind, scope
// and the target that scope maps to valid; before verbatim in
// `git show <target_sha>:<target>` (or empty, for a scoped case whose rule
// file does not exist there yet); after naming no source change, Jira key,
// kan-<n> key or diff path/basename; index.html byte-equal to
// prompttuning.Render. <root>/LEDGER.md, when present, holds only
// well-shaped `- ` entries.
//
// Exit codes: 0 clean, silent; 1 one `<path>: <reason>` line per violation;
// 2 cannot answer (no git, root not inside a git checkout).
func init() {
	Registry["check-prompt-tuning-cases"] = checkPromptTuningCases
}

var (
	ptScope = regexp.MustCompile(`^(generic|language:([a-z0-9+#-]+)|layer:(frontend|backend)|language:([a-z0-9+#-]+)/layer:(frontend|backend))$`)
	ptKan   = regexp.MustCompile(`(?i)\bkan-\d+\b`)
	ptDiffH = regexp.MustCompile(`(?m)^(?:\+\+\+ b/|--- a/)(\S+)`)
	ptEntry = regexp.MustCompile("^- \\d{4}-\\d{2}-\\d{2} · `[^`]+` · (implementer|reviewer) · `[^`]+` · .+ — .+ — (applied [0-9a-f]{7,40}|rejected: .+)$")
)

// ptTarget is the target a valid scope maps to: a "skills/" prefix for
// generic, else the exact rules/ file.
func ptTarget(scope string) (exact string, prefix string) {
	m := ptScope.FindStringSubmatch(scope)
	switch {
	case m[1] == "generic":
		return "", "skills/"
	case m[2] != "":
		return "rules/lang-" + m[2] + ".mdc", ""
	case m[3] != "":
		return "rules/layer-" + m[3] + ".mdc", ""
	}
	return "rules/lang-" + m[4] + "-" + m[5] + ".mdc", ""
}

func checkPromptTuningCases(args []string, env Env, stdout, stderr io.Writer) int {
	const self = "check-prompt-tuning-cases"
	if len(args) > 1 {
		fmt.Fprint(stderr, "usage: check-prompt-tuning-cases.sh [dir]\n")
		return 2
	}
	root := filepath.Join(env.Getenv("FLOW_GUARD_REPO_ROOT"), "docs", "prompt-tuning")
	if len(args) == 1 {
		root = smcAbs(env, args[0])
	}
	if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
		return 0
	}
	git, ok := panelResolveGit(env, self, stderr)
	if !ok {
		return 2
	}
	rc, top := kwRun(env, "", git, "-C", root, "rev-parse", "--show-toplevel")
	if rc != 0 {
		fmt.Fprintf(stderr, "%s: %s is not inside a git checkout\n", self, root)
		return 2
	}
	repo := strings.TrimSpace(top)

	var bad []string
	entries, err := os.ReadDir(root)
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", self, err)
		return 2
	}
	for _, e := range entries {
		if e.IsDir() {
			bad = append(bad, ptCheckCase(env, git, repo, filepath.Join(root, e.Name()))...)
		}
	}
	bad = append(bad, ptCheckLedger(filepath.Join(root, "LEDGER.md"))...)
	for _, b := range bad {
		fmt.Fprintln(stdout, b)
	}
	if len(bad) > 0 {
		return 1
	}
	return 0
}

func ptCheckCase(env Env, git, repo, dir string) []string {
	var bad []string
	for _, f := range prompttuning.RequiredFiles {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			bad = append(bad, fmt.Sprintf("%s/%s: missing", dir, f))
		}
	}
	if len(bad) > 0 {
		return bad
	}
	c, files, err := prompttuning.Load(dir)
	if err != nil {
		return []string{fmt.Sprintf("%s: %v", dir, err)}
	}
	cj := dir + "/case.json"
	for name, v := range map[string]string{
		"id": c.ID, "type": c.Type, "scope": c.Scope, "target": c.Target, "target_sha": c.TargetSHA,
		"defect_class": c.DefectClass, "why_picked": c.WhyPicked, "after": c.After,
		"source.project": c.Source.Project, "source.change": c.Source.Change, "source.jira": c.Source.Jira,
		"source.repo": c.Source.Repo, "source.fix_sha": c.Source.FixSHA, "source.finding_ref": c.Source.FindingRef,
		"source.slot": c.Source.Slot, "source.severity": c.Source.Severity,
	} {
		if v == "" {
			bad = append(bad, fmt.Sprintf("%s: %s is empty", cj, name))
		}
	}
	switch {
	case c.Source.Round == nil:
		bad = append(bad, fmt.Sprintf("%s: source.round is not set", cj))
	case *c.Source.Round < 0:
		bad = append(bad, fmt.Sprintf("%s: source.round %d is negative", cj, *c.Source.Round))
	}
	switch c.Type {
	case "implementer":
		if c.ReviewerKind != "" {
			bad = append(bad, fmt.Sprintf("%s: reviewer_kind is set on an implementer case", cj))
		}
	case "reviewer":
		if c.ReviewerKind != "miss" && c.ReviewerKind != "false-positive" {
			bad = append(bad, fmt.Sprintf("%s: reviewer_kind %q is not miss or false-positive", cj, c.ReviewerKind))
		}
	default:
		bad = append(bad, fmt.Sprintf("%s: type %q is not implementer or reviewer", cj, c.Type))
	}

	generic := c.Scope == "generic"
	if !ptScope.MatchString(c.Scope) {
		bad = append(bad, fmt.Sprintf("%s: scope %q is not generic, language:<lang>, layer:<frontend|backend> or language:<lang>/layer:<frontend|backend>", cj, c.Scope))
	} else if exact, prefix := ptTarget(c.Scope); (prefix != "" && !strings.HasPrefix(c.Target, prefix)) || (exact != "" && c.Target != exact) {
		want := exact
		if prefix != "" {
			want = prefix + "…"
		}
		bad = append(bad, fmt.Sprintf("%s: target %q does not match scope %q (want %s)", cj, c.Target, c.Scope, want))
	}

	if c.TargetSHA != "" {
		if rc, _ := kwRun(env, "", git, "-C", repo, "cat-file", "-e", c.TargetSHA+"^{commit}"); rc != 0 {
			bad = append(bad, fmt.Sprintf("%s: target_sha %s names no commit", cj, c.TargetSHA))
			c.TargetSHA = ""
		}
	}
	if c.Target != "" && c.TargetSHA != "" {
		rc, text := kwRun(env, "", git, "-C", repo, "show", c.TargetSHA+":"+c.Target)
		switch {
		case rc == 0 && c.Before == "":
			bad = append(bad, fmt.Sprintf("%s: before is empty but %s exists at %s", cj, c.Target, c.TargetSHA))
		case rc == 0 && !strings.Contains(text, c.Before):
			bad = append(bad, fmt.Sprintf("%s: before does not occur verbatim in %s at %s", cj, c.Target, c.TargetSHA))
		case rc != 0 && (generic || c.Before != ""):
			bad = append(bad, fmt.Sprintf("%s: before does not occur — %s is absent at %s", cj, c.Target, c.TargetSHA))
		}
	}

	forbidden := map[string]bool{}
	for _, s := range []string{c.Source.Change, c.Source.Jira} {
		if s != "" {
			forbidden[s] = true
		}
	}
	for _, m := range ptDiffH.FindAllStringSubmatch(files.BadDiff+"\n"+files.GoodDiff, -1) {
		forbidden[m[1]] = true
		forbidden[filepath.Base(m[1])] = true
	}
	terms := make([]string, 0, len(forbidden))
	for s := range forbidden {
		terms = append(terms, s)
	}
	sort.Strings(terms)
	low := strings.ToLower(c.After)
	for _, s := range terms {
		if strings.Contains(low, strings.ToLower(s)) {
			bad = append(bad, fmt.Sprintf("%s: after names %q", cj, s))
		}
	}
	for _, k := range ptKan.FindAllString(c.After, -1) {
		if !forbidden[k] {
			bad = append(bad, fmt.Sprintf("%s: after names %q", cj, k))
		}
	}

	page, err := prompttuning.Render(c, files)
	if err != nil {
		return append(bad, fmt.Sprintf("%s: render: %v", dir, err))
	}
	if have, _ := os.ReadFile(filepath.Join(dir, "index.html")); !bytes.Equal(have, page) {
		bad = append(bad, fmt.Sprintf("%s/index.html: differs from flow prompt-tuning render — re-render it", dir))
	}
	return bad
}

func ptCheckLedger(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var bad []string
	for i, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "- ") && !ptEntry.MatchString(line) {
			bad = append(bad, fmt.Sprintf("%s:%d: not a ledger entry — - <YYYY-MM-DD> · `<id>` · implementer|reviewer · `<target>` · <class> — <summary> — applied <sha> | rejected: <reason>", path, i+1))
		}
	}
	return bad
}
