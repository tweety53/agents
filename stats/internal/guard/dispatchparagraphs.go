package guard

import (
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
)

// checkDispatchParagraphs is scripts/check-dispatch-paragraphs.sh: that
// script's header comment is the contract -- every required dispatch
// paragraph is present, with every shared phrase and every required variant,
// at every site the table below names; exit 0 clean, 1 a site missing a
// block or a phrase, 2 cannot answer. The reasoning for each step, moved here
// from the bash body it replaced (d71a2327), sits beside the code it explains.
func init() {
	Registry["check-dispatch-paragraphs"] = checkDispatchParagraphs
}

const dpName = "check-dispatch-paragraphs"

// dpEntry is one paragraph of the table: its label, its shared phrases
// (required of every block carrying the label, held as short literals rather
// than a copy of the whole blockquote), and its variants -- each variant's
// name mapped to its own load-bearing phrase; none means every block of that
// label is equivalent.
type dpEntry struct {
	label    string
	shared   []string
	variants map[string]string
}

// dpEntries is ENTRY_LABEL, ENTRY_SHARED_PHRASES, ENTRY_VARIANTS and
// VARIANT_PHRASE, keyed by the bash table's entry names.
var dpEntries = map[string]dpEntry{
	"reproduce": {"**REPRODUCE, DON'T READ:**",
		[]string{"crosses a boundary", "the store, the filesystem, a guard, a real transcript", "exercise the real thing"},
		map[string]string{"reviewer": "worth less than one you did", "implementer": "a fake or a hand-built value"}},
	"verbatim": {"**VERBATIM REPORT — THE FACT:**",
		[]string{"the reviewer's own report", "never a source of fact", "the report wins"}, nil},
	"foreground": {"**FOREGROUND BUILDS:**",
		[]string{"still executing in the background", "Run it in the foreground", "poll it to completion"}, nil},
	"targeted": {"**TARGETED TESTS:**",
		[]string{"the build tool's own selector", "once for RED, once for GREEN", "module or repository suite mid-task"}, nil},
	"mutation": {"**MUTATION PROOF:**",
		[]string{"mutation-proved before you end your turn", "confirm an existing test fails, and restore", "a surviving mutant",
			"confirm the edit landed", "a refusal, not a surviving mutant", "never buys a test"}, nil},
	"pixel": {"**PIXEL PROBE:**",
		[]string{"draw/geometry code", "actual rendered pixels or geometry", "rejected at the fix step"}, nil},
	"tools": {"**TOOLS:**",
		[]string{"in your first turn", "never a wildcard query", "re-prices your whole context"}, nil},
	"handshake": {"**MODEL HANDSHAKE:**",
		[]string{"the first line of your first reply", "and nothing else on that line", "before any tool call"}, nil},
	"independent": {"**INDEPENDENT PASSES:**",
		[]string{"starts from `final-review.diff`", "raise it again under this pass", "before beginning the next pass"}, nil},
	"delegation": {"**NO DELEGATION:**",
		[]string{"Never call the `Agent` tool", "never spawn a subagent", "the leaf of this run"}, nil},
	"prove": {"**PROVE THE GUARD BITES:**",
		[]string{"assert on configuration or file content", "break the property the assertion protects", "against the broken state"}, nil},
	"decide": {"**REPORT, DON'T DECIDE:**",
		[]string{"only the operator can settle", "reported, never decided in code", "carry the question in your REPORT FILE", "report BLOCKED"}, nil},
	"entry": {"**ENTRY CONTEXT:**",
		[]string{"your named entry context", "not from a whole-tree exploration", "the coverage trade"}, nil},
	"findings": {"**FINDINGS ARE INPUT:**",
		[]string{"input, not orders", "resolve the defect the finding names", "records the deviation and justifies it", "A silent deviation is an unfixed finding"}, nil},
	"contextbundle": {"**CONTEXT BUNDLE FAILURE:**",
		[]string{"exited non-zero, or the bundle file is still", "context bundle: build failed", "dispatch every slot without the bundle",
			"context bundle: dispatched without it — operator override", "outcome stopped"}, nil},
	"budget": {"**OUTPUT BUDGET:**",
		[]string{"re-reads your", "by line range", "never re-read a file already in your context", "Never print a generated file"}, nil},
	"readonly": {"**READ-ONLY REVIEW:**",
		[]string{"never mutate it", "no file edit outside your own report file", "a claim nobody can check"}, nil},
}

// dpSite is one row of SITE_ENTRY/SITE_PATHS/SITE_MIN_BLOCKS/SITE_VARIANTS:
// the entry (paragraph) it requires, its path relative to the root, its
// minimum block count, and the variants it requires (none: nothing beyond
// the shared phrases).
type dpSite struct {
	entry    string
	path     string
	min      int
	variants []string
}

// dpSites is the site table, in the bash table's order -- the order its
// violations print in.
var dpSites = []dpSite{
	{"reproduce", "skills/flow/review-panel.md", 1, []string{"reviewer"}},
	{"reproduce", "skills/flow/implement.md", 3, []string{"reviewer", "implementer"}},
	{"verbatim", "skills/flow/review-panel.md", 1, nil},
	{"foreground", "skills/flow/implement.md", 3, nil},
	{"foreground", "skills/flow/review-panel.md", 2, nil},
	{"targeted", "skills/flow/implement.md", 1, nil},
	{"targeted", "skills/flow/review-panel.md", 1, nil},
	{"mutation", "skills/flow/review-panel.md", 1, nil},
	{"pixel", "skills/flow/review-panel.md", 1, nil},
	{"tools", "skills/flow/implement.md", 2, nil},
	{"tools", "skills/flow/review-panel.md", 2, nil},
	{"tools", "skills/flow/visual-verify.md", 2, nil},
	{"handshake", "skills/flow/implement.md", 2, nil},
	{"handshake", "skills/flow/review-panel.md", 2, nil},
	{"handshake", "skills/flow/visual-verify.md", 2, nil},
	{"independent", "skills/flow/review-panel.md", 1, nil},
	{"delegation", "skills/flow/implement.md", 2, nil},
	{"delegation", "skills/flow/review-panel.md", 2, nil},
	{"delegation", "skills/flow/visual-verify.md", 2, nil},
	{"prove", "skills/flow/implement.md", 1, nil},
	{"decide", "skills/flow/implement.md", 1, nil},
	{"entry", "skills/flow/review-panel.md", 1, nil},
	{"findings", "skills/flow/review-panel.md", 1, nil},
	{"contextbundle", "skills/flow/review-panel.md", 1, nil},
	{"budget", "skills/flow/implement.md", 1, nil},
	{"budget", "skills/flow/review-panel.md", 1, nil},
	{"readonly", "skills/flow/implement.md", 1, nil},
}

func checkDispatchParagraphs(_ []string, env Env, stdout, stderr io.Writer) int {
	// The root is this repository's own root by default -- the parent of
	// scripts/, derived as the bash guard derived it from its own location;
	// the binary lives in a cache, so the shim exports it as
	// FLOW_GUARD_REPO_ROOT. CHECK_DISPATCH_PARAGRAPHS_ROOT is an explicit,
	// opt-in override honoured only when set, so the tests can point this
	// guard at a sandboxed fixture tree -- never set it for a normal
	// invocation. Set but empty is a refusal, not the default.
	root, set := env.LookupEnv("CHECK_DISPATCH_PARAGRAPHS_ROOT")
	switch {
	case set && root == "":
		fmt.Fprintf(stderr, "%s: CHECK_DISPATCH_PARAGRAPHS_ROOT is set but empty\n", dpName)
		return 2
	case !set:
		if root = env.Getenv("FLOW_GUARD_REPO_ROOT"); root == "" {
			fmt.Fprintf(stderr, "%s: FLOW_GUARD_REPO_ROOT is unset — run scripts/check-dispatch-paragraphs.sh, which sets it\n", dpName)
			return 2
		}
	}
	// Paths print as the bash built them ("$ROOT/<site>"); only the file
	// system calls see a relative root joined to the working directory,
	// uncleaned so the kernel resolves any "..", as bash's did.
	fsPath := func(p string) string { return smcAbs(env, p) }
	if !gsReadableDir(fsPath(root)) {
		fmt.Fprintf(stderr, "%s: %s is not a readable directory — cannot scan\n", dpName, root)
		return 2
	}

	var violations []string
	for _, site := range dpSites {
		full := root + "/" + site.path
		lines, refusal := dpReadSite(fsPath(full), dpEntries[site.entry].label)
		if refusal != "" {
			// A hard "cannot answer at all" refusal, in the same file:line
			// shape as every other finding: an access failure at one required
			// site means the whole run cannot answer, not that this one site
			// failed -- so the violations gathered so far are never printed.
			fmt.Fprintf(stderr, "%s:0: %s\n", full, refusal)
			return 2
		}
		violations = append(violations, dpCheckSite(site, full, lines)...)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintln(stdout, v)
		}
		fmt.Fprintf(stdout, "DISPATCH-PARAGRAPHS-INVALID: %s — %d violation(s)\n", root, len(violations))
		return 1
	}
	fmt.Fprintf(stdout, "DISPATCH-PARAGRAPHS-OK: %s — %d site(s) validated\n", root, len(dpSites))
	return 0
}

// dpReadSite reads one required site's lines, or names why it cannot: a
// symlink, a missing path, a non-regular file and an unreadable one are each
// refused -- never a silent skip, which would be a vacuous "✓ clean". A
// path that exists as neither a file nor a directory folds into "not a
// regular file".
func dpReadSite(p, label string) ([]string, string) {
	if fi, err := os.Lstat(p); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return nil, "is a symlink — a required dispatch-paragraph site must be a real file, never a symlink"
	}
	fi, err := os.Stat(p)
	if err != nil {
		return nil, "does not exist — this is a required dispatch-paragraph site"
	}
	if !fi.Mode().IsRegular() {
		return nil, "exists but is not a regular file (neither file nor a symlink to check) — cannot scan"
	}
	if syscall.Access(p, pcReadOK) != nil {
		return nil, "is not readable — cannot scan"
	}
	data, err := os.ReadFile(p)
	if err != nil {
		// The bash reported a grep exit >= 2 here; the port reads the file
		// itself, so the refusal names the read error instead.
		return nil, fmt.Sprintf("could not be read while scanning for the label \"%s\" (%v) — a failure to look, not an absence", label, err)
	}
	lines := strings.Split(string(data), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines, ""
}

// dpCheckSite returns one site's violations. A block counts toward a required
// variant only when it carries the label, every shared phrase, AND that
// variant's own phrase: two blocks of the same variant do NOT satisfy a
// two-variant site. Every failure is reported as `file:line` naming the
// missing element -- the label, the specific missing phrase, or the missing
// variant.
func dpCheckSite(site dpSite, full string, lines []string) []string {
	e := dpEntries[site.entry]
	var out []string
	blocks := 0
	hasVariant := map[string]bool{}
	for i, line := range lines {
		// The label is matched on the raw line, as the script's header states
		// and as GNU grep -aF and LC_ALL=C did; the text walked below is
		// mapfile's view of it, which ends at the first NUL byte. One
		// divergence from the bash at d71a2327: macOS BSD grep 2.6.0 in a
		// UTF-8 locale finds no non-ASCII label (the em-dash of **VERBATIM
		// REPORT — THE FACT:**) anywhere in a file holding a NUL byte, so the
		// bash reported such a label missing where this finds it.
		if !strings.Contains(line, e.label) {
			continue
		}
		blocks++
		text := dpBlockText(lines, i)
		missing := false
		for _, phrase := range e.shared {
			if !strings.Contains(text, phrase) {
				missing = true
				out = append(out, fmt.Sprintf("%s:%d: block carrying \"%s\" is missing the required phrase: \"%s\"", full, i+1, e.label, phrase))
			}
		}
		if !missing {
			for v, phrase := range e.variants {
				if strings.Contains(text, phrase) {
					hasVariant[v] = true
				}
			}
		}
	}
	if blocks < site.min {
		out = append(out, fmt.Sprintf("%s:0: requires at least %d block(s) carrying the label \"%s\", found %d", full, site.min, e.label, blocks))
	}
	for _, v := range site.variants {
		if !hasVariant[v] {
			out = append(out, fmt.Sprintf("%s:0: missing a block that satisfies the %s variant (the label \"%s\" plus every shared phrase plus \"%s\")", full, v, e.label, e.variants[v]))
		}
	}
	return out
}

// dpBlockText is a block's text: the label line plus every immediately
// following blockquote-continuation line, each stripped of its `>` and one
// space and joined with single spaces (matching how a wrapped markdown
// paragraph reads as prose), so a phrase split across two source lines by
// the wrap is still found as one substring. A label line that is not itself
// a blockquote line is a block of one line.
//
// The continuation walk stops at the first line that ITSELF carries any known
// label, even though that line still begins with '>'. Without this, two
// required blocks glued together by a missing blank line (a future edit's
// copy-paste slip) merge into one block, and a paraphrase in the first that
// quietly drops a required phrase can pass by absorbing the second block's
// own text -- the exact failure class this guard exists to catch, defeated
// by its own continuation rule.
func dpBlockText(lines []string, start int) string {
	quoted := strings.HasPrefix(dpMapfile(lines[start]), ">")
	var parts []string
	for j := start; ; {
		cur := dpMapfile(lines[j])
		if strings.HasPrefix(cur, ">") {
			cur = strings.TrimPrefix(cur[1:], " ")
		}
		parts = append(parts, cur)
		j++
		if !quoted || j >= len(lines) {
			break
		}
		next := dpMapfile(lines[j])
		if !strings.HasPrefix(next, ">") || dpCarriesLabel(next) {
			break
		}
	}
	return strings.Join(parts, " ")
}

func dpCarriesLabel(line string) bool {
	for _, e := range dpEntries {
		if strings.Contains(line, e.label) {
			return true
		}
	}
	return false
}

// dpMapfile is a line as bash's mapfile held it: a shell variable cannot
// carry a NUL, so the line ends at the first one.
func dpMapfile(line string) string {
	if i := strings.IndexByte(line, 0); i >= 0 {
		return line[:i]
	}
	return line
}
