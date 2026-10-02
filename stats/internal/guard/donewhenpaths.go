package guard

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// checkDoneWhenPaths is scripts/check-done-when-paths.sh: that script's
// header comment is the contract — one DONE-WHEN-PATH line per path a
// `## Done when` section names that the index does not track, then one
// DONE-WHEN-OK/DONE-WHEN-VIOLATION verdict line; exit 0 DONE-WHEN-OK (a tree
// with no `## Done when` section included), 1 DONE-WHEN-VIOLATION, 2 with
// nothing on stdout when it cannot answer.
//
// WHY THIS GUARD EXISTS (KAN-817, kan-743's self-review). A done criterion is
// only as good as the committed state it names: kan-743's own Done-when named
// three gitignored side-output PNGs as the re-baseline, cleanup later removed
// them as side output, and the change stayed Done while its definition of
// done no longer held anywhere. This guard makes that state a refusal at the
// archive commit instead: a Done-when's evidence must sit in the index —
// tracked, not merely present or ignored — or the archive does not happen.
func init() { Registry["check-done-when-paths"] = checkDoneWhenPaths }

// dwHeading opens a `## Done when` section (the \b keeps `Done whenever`
// out); dwAnyHeading closes it, the opening heading excepted.
var (
	dwHeading    = regexp.MustCompile(`(?i)^#{1,6}[ \t]+done[ \t-]when\b:?[ \t]*`)
	dwAnyHeading = regexp.MustCompile(`^#{1,6}[ \t]`)
	dwURL        = regexp.MustCompile(`(?i)^[a-z][a-z0-9+.-]*://`)
)

func checkDoneWhenPaths(args []string, env Env, stdout, stderr io.Writer) int {
	die := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "check-done-when-paths: "+format+"\n", a...)
		return 2
	}
	if len(args) != 1 {
		return die("usage: check-done-when-paths.sh <worktree>")
	}
	root, ok := cdPhysical(env, args[0])
	if !ok {
		return die("%s is not a directory", args[0])
	}
	git := envGit(env)
	if git("-C", root, "rev-parse", "--is-inside-work-tree").Run() != nil {
		return die("%s is not a git worktree", root)
	}
	// One listing is both the scan set (its .md entries) and the index the
	// named paths are checked against — the same tracked set, so a named
	// path fails exactly when it is missing from it.
	raw, ok := capture(git("-C", root, "ls-files", "-z"))
	if !ok {
		return die("cannot read the index of %s", root)
	}
	tracked := map[string]bool{}
	var markdown []string
	for _, p := range strings.Split(raw, "\x00") {
		if p == "" {
			continue
		}
		tracked[p] = true
		if strings.HasSuffix(p, ".md") {
			markdown = append(markdown, p)
		}
	}

	var offenders []string
	seen := map[string]bool{}
	for _, file := range markdown {
		b, err := os.ReadFile(root + "/" + file)
		if err != nil {
			return die("cannot read tracked file %s in %s: %v", file, root, err)
		}
		inSection := false
		fence := false
		for _, line := range strings.Split(string(b), "\n") {
			s := strings.TrimSpace(line)
			if strings.HasPrefix(s, "```") || strings.HasPrefix(s, "~~~") {
				fence = !fence
				continue
			}
			if fence {
				// A fenced line neither opens nor closes a section — a
				// quoted `## Done when` template is not a section — and
				// yields no paths.
				continue
			}
			if dwHeading.MatchString(line) {
				inSection = true
			} else if dwAnyHeading.MatchString(line) {
				inSection = false
			}
			if !inSection {
				continue
			}
			for _, p := range dwPaths(line) {
				if !tracked[p] && !seen[p+"\x00"+file] {
					seen[p+"\x00"+file] = true
					offenders = append(offenders, fmt.Sprintf("%s — %s", p, file))
				}
			}
		}
	}
	if len(offenders) > 0 {
		for _, o := range offenders {
			fmt.Fprintf(stdout, "DONE-WHEN-PATH: %s\n", o)
		}
		fmt.Fprintf(stdout, "DONE-WHEN-VIOLATION: %s — %d\n", root, len(offenders))
		return 1
	}
	fmt.Fprintf(stdout, "DONE-WHEN-OK: %s\n", root)
	return 0
}

// dwPaths extracts the path-like tokens of one line: a whitespace field
// stripped of its backticks and edge punctuation — markdown link and autolink
// syntax included, so `[e](shots/27.png)` and `<shots/27.png>` yield the path
// — kept when it carries a `/` whose last segment names a file: not empty and
// not purely numeric, so a ticket's `Snapshots 27/28/29` never reads as a
// path while a dotless `src/Makefile` is judged like any other name; dropped
// when it is a URL, a glob (`*?[` make it uncheckable against an index) or a
// directory. A leading `./` is folded away; membership is exact afterwards,
// so a peer-qualified name (`peer:shots/27.png`) fails here as the untracked
// name it is in this index.
func dwPaths(line string) []string {
	var out []string
	for _, f := range strings.Fields(line) {
		f = strings.Trim(f, "`")
		// A markdown link's target sits after the last `](`; the generic
		// edge trim alone would keep the `e](` of `[e](shots/27.png)`.
		if i := strings.LastIndex(f, "]("); i >= 0 {
			f = strings.TrimSuffix(f[i+2:], ")")
		}
		// Folded before the edge trim: that trim's cutset carries `.`, so a
		// leading `./` would otherwise lose only its dot and become `/…`.
		f = strings.TrimPrefix(f, "./")
		f = strings.Trim(f, "()[],.;:!?\"'<>")
		// Punctuation can wrap a backticked path — `(`shots/27.png`
		// re-baselined)` — so the backtick trim runs again once the edge
		// trim has exposed what it hid.
		f = strings.Trim(f, "`")
		if f == "" || !strings.Contains(f, "/") || dwURL.MatchString(f) {
			continue
		}
		if strings.ContainsAny(f, "*?[") || strings.HasSuffix(f, "/") {
			continue
		}
		// A lone leading `/` is a slash command (`/myflow-start`), never a
		// path: no index entry starts with `/`.
		if strings.LastIndex(f, "/") == 0 {
			continue
		}
		last := f[strings.LastIndex(f, "/")+1:]
		if last == "" || strings.Trim(last, "0123456789") == "" {
			continue
		}
		out = append(out, f)
	}
	return out
}
