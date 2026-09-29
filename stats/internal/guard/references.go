package guard

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// checkReferences is scripts/check-references.sh: that script's header
// comment is the contract -- for every backticked .md/.mdc path a line
// ASSOCIATES a **bold token** with, at least one associated token must match
// a heading of that file; exit 0 clean, 1 stale references or a coverage
// violation, 2 cannot answer. The reasoning for each step, moved here from
// the bash body it replaced (3e48ecac), sits beside the code it explains.
func init() {
	Registry["check-references"] = checkReferences
}

// crTargets is DEFAULT_TARGETS: the scan set lives here, in one place, so no
// call site can narrow it.
var crTargets = []string{"rules", "skills", "commands-claude", "README.md", "AGENTS.md", "CLAUDE.md"}

// EXPECTED-ZERO FILES — established by running this guard's own association
// and resolution logic against the real tree (2026-08-18, at df9d5dd), never
// guessed: each of these carries no `**Bold**`-adjacent `.md`/`.mdc` path in
// any of the shapes crAssociated recognizes, so this guard genuinely
// verifies nothing inside it. Declared here, once, rather than inferred from
// the tree — inferring it would restate the very assumption a silently
// uncovered file already encodes, which is exactly the case this exists to
// fail instead of pass.
//
// KAN-197 F2: batched by CATEGORY rather than one flat list sharing a single
// reason string. The panel's Primary slot found the original single reason
// attested only to HOW the zero was measured ("ran the association logic
// against this file"), never to WHY it is legitimate BY DESIGN — weaker than
// this requirement's own words, "legitimately checks nothing". Each category
// below states the actual shape that makes its members' bold/path pairs never
// associate, verified per category rather than guessed:
//
//	command-dispatch stub — every path it cites sits INSIDE the same bold
//	span as the verb citing it (e.g. "**load `skills/.../pipeline.md`
//	first**"); crLooksLikeSection rejects any bold span containing "/", so no
//	candidate section name is ever formed next to the path.
//
//	rule file — a path citation sits in a Markdown table cell (rules/agent-
//	baseline.md's rule table) separated from any bold text by far more than
//	the adjacency window, or the file cites no path in a bold-adjacent shape
//	at all.
//
//	contract/index doc — cites other files as a plain parenthetical backtick
//	path or a `[label](path)` Markdown link (skills/flow-contracts/SKILL.md's
//	own table), never as a bold token adjacent to the path.
//
//	reviewer-prompt file — deliberately self-contained; where it cites a
//	.md/.mdc path it never pairs the citation with an adjacent bold
//	section name.
//
//	rationale/exploration doc — prose-only; any path citation sits inside the
//	same bold span as its citing verb (the command-dispatch-stub shape), or
//	with no bold nearby at all.
var crExpectedZero = []struct {
	reason string
	files  []string
}{
	{"command-dispatch stub — every path it cites sits inside the SAME bold span as the verb citing it (e.g. \"**load `path` first**\"); looks_like_section rejects any bold span containing '/', so no candidate section name ever forms adjacent to the path",
		[]string{
			"commands-claude/flow.md",
			"commands-claude/flow-plan.md",
			"commands-claude/flow-self-review.md",
			"commands-claude/flow-settings.md",
			"commands-claude/flow-status.md",
		}},
	{"rule file — its own path citations (where present) sit in a Markdown table cell or plain prose, separated from any bold text by more than the adjacency window this guard's is_associated allows, or cite no path in a bold-adjacent shape at all",
		[]string{
			"rules/agent-baseline.md",
			"rules/be-brief.mdc",
			"rules/build-the-simplest-thing.mdc",
			"rules/commit-scope-is-the-module.mdc",
			"rules/context7.mdc",
			"rules/dependency-versions.mdc",
			"rules/design-mockups-are-specs.mdc",
			"rules/dispatch-carries-the-baseline.mdc",
			"rules/fix-determinism-at-the-source.mdc",
			"rules/lint-fix-priority.mdc",
			"rules/never-touch-production.mdc",
		}},
	{"contract/index doc — cites other files as a plain parenthetical backtick path or a [label](path) Markdown link, never as a bold token adjacent to the path",
		[]string{
			"CLAUDE.md",
			"skills/flow-contracts/plan-provenance.md",
			"skills/flow-contracts/build-green.md",
			"skills/flow-contracts/SKILL.md",
		}},
	{"reviewer-prompt file, deliberately self-contained — most cite no .md/.mdc path anywhere, and the rest never pair a citation with an adjacent bold section name",
		[]string{
			"skills/flow/engineering-principles.md",
			"skills/flow/reviewer-calibration.md",
		}},
	{"rationale/exploration doc, prose-only — any path citation sits inside the same bold span as its citing verb, or with no bold nearby at all",
		[]string{
			"rules/agent-baseline-rationale.md",
			"skills/flow-contracts/git-boundaries-rationale.md",
			"skills/flow-contracts/session-records-rationale.md",
			"skills/flow-contracts/worktree-resolution-rationale.md",
			"skills/flow-contracts/SKILL-rationale.md",
			"skills/flow-fast/SKILL-rationale.md",
			"skills/flow-plan/SKILL-rationale.md",
			"skills/flow-self-review/SKILL-rationale.md",
			"skills/flow-settings/SKILL-rationale.md",
			"skills/flow-status/SKILL-rationale.md",
		}},
}

// crScan is one run's state: the root three ways, the coverage it records and
// the stale references it has printed.
type crScan struct {
	env                      Env
	stdout                   io.Writer
	root, rootNorm, rootPhys string
	cov                      coverage
	failures                 int
}

func checkReferences(_ []string, env Env, stdout, stderr io.Writer) int {
	// REPO_ROOT is the guard script's own checkout by default — what makes
	// the guard argument-free and self-scoped from any cwd: running it from
	// /tmp, or from a subdirectory of the repo, still scans the real repo
	// instead of silently scanning nothing and reporting a false "clean".
	// The bash guard derived it from BASH_SOURCE; this binary lives in a
	// cache, so its shim exports FLOW_GUARD_REPO_ROOT, derived the same way.
	//
	// CHECK_REFERENCES_ROOT is an explicit, opt-in override honored only when
	// set. It exists solely so the tests can point the guard at a sandboxed
	// fixture tree without touching this repo — never set it for a normal
	// invocation.
	root, set := env.LookupEnv("CHECK_REFERENCES_ROOT")
	switch {
	case set && root == "":
		fmt.Fprint(stderr, "CHECK_REFERENCES_ROOT is set but empty\n")
		return 2
	case !set:
		if root = env.Getenv("FLOW_GUARD_REPO_ROOT"); root == "" {
			fmt.Fprint(stderr, "check-references: FLOW_GUARD_REPO_ROOT is unset — run scripts/check-references.sh, which sets it\n")
			return 2
		}
	}
	// A root that is not a directory would make every target miss and the
	// guard report a clean run over zero files — a false "all clear", which
	// is the one outcome a guard must never produce.
	r := &crScan{env: env, stdout: stdout, root: root}
	r.rootNorm = r.lexicalNorm(root)
	if !isDir(r.rootNorm) {
		fmt.Fprintf(stderr, "not a directory: %s\n", root)
		return 2
	}
	phys, err := filepath.EvalSymlinks(r.rootNorm)
	if err != nil {
		fmt.Fprintf(stderr, "check-references: cannot resolve %s: %v\n", root, err)
		return 2
	}
	r.rootPhys = phys

	// declare_category: each file is declared ONLY when it exists under the
	// CURRENT root (KAN-197 F3 compatibility). These lists are this
	// repository's own real paths; a sandboxed fixture scans a different,
	// smaller tree where most of them do not exist at all, and declaring a
	// path that is not part of the current run's corpus would make it a
	// "declared but never recorded" violation for every such fixture — not a
	// real staleness, just a scope mismatch. Existence-gating keeps F3's
	// protection meaningful for the real, default run (every path genuinely
	// exists there today) without that false-positive noise.
	for _, cat := range crExpectedZero {
		for _, f := range cat.files {
			if !isFile(root + "/" + f) {
				continue
			}
			if err := r.cov.declare(f, cat.reason); err != nil {
				fmt.Fprintf(stderr, "%v\ncheck-references: coverage_declare failed for %s (see stderr above)\n", err, f)
				return 2
			}
		}
	}

	scanned := 0
	for _, target := range crTargets {
		full := root + "/" + target
		if _, err := os.Stat(full); err != nil {
			continue
		}
		files := []string{full}
		if isDir(full) {
			if files, err = crMarkdownUnder(env, full); err != nil {
				fmt.Fprintf(stderr, "check-references: cannot list %s: %v\n", full, err)
				return 2
			}
		}
		for _, f := range files {
			if err := r.checkFile(f); err != nil {
				if ref, ok := err.(crRefusal); ok {
					fmt.Fprintln(stderr, string(ref))
					return 2
				}
				fmt.Fprintf(stderr, "%v\ncheck-references: coverage_record failed for %s (see stderr above)\n", err, r.rel(f))
				return 2
			}
			scanned++
		}
	}
	if scanned == 0 {
		fmt.Fprintf(stderr, "no Markdown files found under %s — refusing to report a clean run\n", root)
		return 2
	}

	// COVERAGE — per-file count of what this guard actually verified. A file
	// whose coverage is zero and is not declared above is folded into the
	// ordinary stale-reference violations rather than a separate exit status,
	// exactly as check-guard-symlinks.sh does it: "<member>: <text>" becomes
	// "<member>:0: <text>".
	for _, line := range r.cov.verdict() {
		member, _, _ := strings.Cut(line, ":")
		_, text, _ := strings.Cut(line, ": ")
		fmt.Fprintf(stdout, "%s:0: %s\n", member, text)
		r.failures++
	}
	if r.failures != 0 {
		fmt.Fprintf(stderr, "\n%d stale reference(s) found.\n", r.failures)
		fmt.Fprint(stderr, "Fix the reference, or mark the line refs-guard:allow if the bold text is emphasis.\n")
		return 1
	}
	fmt.Fprint(stdout, "check-references: all referenced sections resolve\n")
	if frag := r.cov.report(); frag != "" {
		fmt.Fprintf(stdout, "  %s\n", frag)
	}
	return 0
}

// crMarkdownUnder is `find <dir> -type f \( -name '*.md' -o -name '*.mdc' \)
// | sort`: no symlink followed, each path spelled <dir>/<rest> as find
// prints it, ordered by the `sort` on PATH under env's locale — the bash
// guard's order, which the coverage fragment and the failure lines carry.
func crMarkdownUnder(env Env, dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if d != nil && d.IsDir() && p != dir {
				return fs.SkipDir
			}
			return nil
		}
		if d.Type().IsRegular() && (strings.HasSuffix(d.Name(), ".md") || strings.HasSuffix(d.Name(), ".mdc")) {
			rel, err := filepath.Rel(dir, p)
			if err != nil {
				return err
			}
			files = append(files, dir+"/"+rel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return crSort(env, files, false)
}

// crSort is `sort` (or `sort -u`) over lines, exec'd rather than
// reimplemented: its order is the locale's collation (en_US.UTF-8 on macOS
// sorts "SKILL-rationale.md" after "session-records-rationale.md"), which
// no byte comparison reproduces, and the guard's output order is its.
func crSort(env Env, lines []string, unique bool) ([]string, error) {
	if len(lines) < 2 {
		return lines, nil
	}
	bin, ok := lookPath(env, "sort")
	if !ok {
		return nil, fmt.Errorf("no sort on PATH")
	}
	var args []string
	if unique {
		args = []string{"-u"}
	}
	cmd := exec.Command(bin, args...)
	cmd.Env = []string{}
	for _, k := range []string{"LANG", "LC_ALL", "LC_COLLATE", "LC_CTYPE"} {
		if v, set := env.LookupEnv(k); set {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}
	cmd.Stdin = strings.NewReader(strings.Join(lines, "\n") + "\n")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("sort: %v", err)
	}
	return strings.Split(strings.TrimSuffix(string(out), "\n"), "\n"), nil
}

// rel is `${file#"$REPO_ROOT"/}`.
func (r *crScan) rel(file string) string { return strings.TrimPrefix(file, r.root+"/") }

// lexicalNorm is lexical_norm: the path with `.` and `..` resolved
// TEXTUALLY against the working directory, without touching the filesystem.
// Purely lexical on purpose: containment must be decided from the shape of
// the reference alone, so the guard's verdict cannot depend on whether the
// out-of-tree target happens to exist on this machine.
func (r *crScan) lexicalNorm(p string) string {
	if !strings.HasPrefix(p, "/") {
		p = r.env.Dir + "/" + p
	}
	return path.Clean("/" + p)
}

// contained is contained: the candidate normalized to an absolute path, but
// only when it lies inside the root. Repository Markdown is
// attacker-influenceable (any pull request can add a line), and the path in
// a backtick span is read from disk — without this, `../../outside/target.md`
// is opened, which is a heading-existence oracle over the whole filesystem
// and a stall on a large file.
//
// Two tests, in this order:
//
//  1. LEXICAL, always. `..` is resolved textually and the result must sit
//     under the root. This runs before any stat, so a crafted reference is
//     refused identically whether or not the target exists — otherwise the
//     same repository would warn on a laptop and pass in CI, and the
//     attacker-probe case (a reference to a path that may or may not be
//     there) would be the quiet one.
//  2. PHYSICAL, when the target's directory exists. Symlinks are resolved,
//     so a link inside the tree cannot point the guard out of it.
func (r *crScan) contained(candidate string) (string, bool) {
	norm := r.lexicalNorm(candidate)
	if !strings.HasPrefix(norm, r.rootNorm+"/") {
		return "", false
	}
	if dir := path.Dir(norm); isDir(dir) {
		phys, _ := filepath.EvalSymlinks(dir) // a failed `cd` left bash's `pwd -P` empty too
		if !strings.HasPrefix(phys+"/"+path.Base(norm), r.rootPhys+"/") {
			return "", false
		}
	}
	return norm, true
}

// crLines is `while IFS= read -r line || [ -n "$line" ]`: every line,
// the last one counted even without its newline.
func crLines(b []byte) []string {
	lines := strings.Split(string(b), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// crIsFence is is_fence_line, the single shared predicate for fence-toggle
// tracking: checkFile (which must preserve line numbers while scanning the
// REFERENCING file) and crHeadings (scanning the REFERENCED file) both call
// it, so "a line that starts or ends a fence" is defined in exactly one
// place and the two scans can never quietly diverge on it.
func crIsFence(line string) bool {
	return strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~")
}

// crNormalize is normalize_token, and the tail of headings_of: backticks and
// emphasis removed, trimmed, lowercased.
func crNormalize(s string) string {
	s = strings.NewReplacer("`", "", "*", "").Replace(s)
	return strings.ToLower(strings.TrimFunc(s, unicode.IsSpace))
}

// crHeadings is headings_of: every #, ##, ###, or #### heading OUTSIDE a
// fenced code block, normalized. Fence-aware so a "# comment" inside a
// ```bash example is never mistaken for a real section heading. ok is false
// when no heading normalizes to anything — the bash guard's empty `$heads`.
func crHeadings(file string) (heads map[string]bool, ok bool) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, false
	}
	heads = map[string]bool{}
	inFence := false
	for _, line := range crLines(b) {
		if crIsFence(line) {
			inFence = !inFence
			continue
		}
		n := len(line) - len(strings.TrimLeft(line, "#"))
		if inFence || n < 1 || n > 4 || len(line) == n || line[n] != ' ' {
			continue
		}
		if h := crNormalize(line[n+1:]); h != "" {
			heads[h], ok = true, true
		}
	}
	return heads, ok
}

// crMaskCodeSpans is mask_code_spans: every "**" INSIDE a backtick-delimited
// code span becomes "@@"; everything else — the backticks, other code-span
// text, any "**" outside a code span — is left byte-for-byte unchanged.
//
// Why: a line like "See **State file** and code `x**y` in
// `rules/contract.mdc`." carries a LITERAL "**" inside inline code that is
// not a bold delimiter at all; counting or pairing "**" without excluding
// code spans first lets that literal pair desync the real bold span next to
// it. Masking only the in-code occurrences (not the whole span) is
// deliberate: a bold span that itself WRAPS a code span, e.g.
// "**`prUrl: null`**", must still extract as "`prUrl: null`" untouched.
//
// Scoped to single-backtick spans (the shape every real reference in this
// repo uses); a run of 2+ backticks used to escape a literal backtick
// inside code is not specially handled.
func crMaskCodeSpans(line string) string {
	var out strings.Builder
	inCode := false
	for i := 0; i < len(line); i++ {
		switch {
		case line[i] == '`':
			inCode = !inCode
		case inCode && strings.HasPrefix(line[i:], "**"):
			out.WriteString("@@")
			i++
			continue
		}
		out.WriteByte(line[i])
	}
	return out.String()
}

// crNormalizedLine is normalized_line: in-code "**" masked, and a leading
// orphaned closing "**" stripped. A naive pairing mis-pairs markers when the
// line opens with the tail end of a bold span that started on the PREVIOUS
// physical line (prose soft-wrapped mid-span): on "sync** in **Jira
// integration** (...)" it would extract "** in **" and never see
// "**Jira integration**". Markers on a line with no cross-line continuation
// always come in pairs, so an odd count (on the code-masked line) means the
// first is such an orphan: strip exactly that one before pairing.
func crNormalizedLine(line string) string {
	masked := crMaskCodeSpans(line)
	if strings.Count(masked, "**")%2 == 1 {
		masked = strings.Replace(masked, "**", "", 1)
	}
	return masked
}

var (
	crQuote      = regexp.MustCompile(`^(>\s?)+`)
	crBlockStart = regexp.MustCompile(`^(#{1,6} |\||---)`)
	crNumbered   = regexp.MustCompile(`^\d+[.)] `)
	crMdSpan     = regexp.MustCompile("`[^`]*\\.mdc?`")
	crCodeSpan   = regexp.MustCompile("`[^`]*`")
)

// crAssociated is is_associated, the adjacency test applied to the text
// BETWEEN a bold span and a path span (in either order): the gap must be
// short (<= 60 bytes and, ignoring ordinary inline code, <= 3 words), must
// not contain another .md/.mdc reference or another bold marker, and must
// not contain a "." — a sentence boundary between the two means they belong
// to different statements, not to one reference. That is what separates
// "**Section** in `path`" from "**Never** commit. See `path`".
func crAssociated(gap string) bool {
	if len(gap) > 60 || crMdSpan.MatchString(gap) {
		return false
	}
	// Ordinary inline code between them (`jq`, `--target`) is connective
	// tissue, not a separator — drop it before judging the prose.
	clean := crCodeSpan.ReplaceAllString(gap, " ")
	if strings.Contains(clean, "**") || strings.Contains(clean, ".") {
		return false
	}
	words := strings.FieldsFunc(clean, func(c rune) bool {
		return !('A' <= c && c <= 'Z' || 'a' <= c && c <= 'z')
	})
	return len(words) <= 3
}

// crLooksLikeSection is looks_like_section: a bold span is a candidate
// SECTION NAME only if it reads like one. Bold is used for emphasis far more
// often than for section names in this repo ("**only** in projects whose
// `.flow/project.md`"), and an emphasis word next to a filename is not a
// cross-reference. Every heading in the referenced files starts with a
// capital, carries no "." and no trailing ":", so those cheap tests separate
// the two uses.
//
// Cost, stated plainly: a genuine reference written with a lowercase or
// punctuated bold token is not checked. Heading MATCHING stays
// case-insensitive, so this only ever decides whether a path is examined —
// it never turns a match into a miss.
func crLooksLikeSection(t string) bool {
	s := strings.Trim(strings.ReplaceAll(t, "`", ""), " \t")
	return s != "" && !strings.ContainsAny(s, "/.") && !strings.HasSuffix(s, ":") && 'A' <= s[0] && s[0] <= 'Z'
}

// crSpan is a bold or path span: start and end byte offsets (end is the
// last byte of its closing delimiter) and the text between the delimiters.
type crSpan struct {
	start, end int
	text       string
}

// crAssociations is associations_of over a normalized line: every (path,
// bold token) pair the adjacency rule associates, as "path\ttoken" lines
// split on tabs the way the bash guard's `cut -f1` and `awk -F'\t'` read
// them back.
func crAssociations(line string) [][2]string {
	return crAssociationsWhere(line, func(_, _ crSpan) bool { return true })
}

// crAssociationsAcross is crAssociations over two joined lines, keeping only
// the pairs whose bold and path spans do not both sit on one side of the
// join at byte offset brk.
func crAssociationsAcross(joined string, brk int) [][2]string {
	return crAssociationsWhere(joined, func(b, p crSpan) bool {
		return min(b.start, p.start) < brk && max(b.end, p.end) > brk
	})
}

// crAssociationsWhere is crAssociations, a pair kept only when keep accepts
// its bold and path spans. keep runs after the nearest-path choice, so a
// dropped pair never hands its token to a farther path.
func crAssociationsWhere(line string, keep func(bold, path crSpan) bool) [][2]string {
	var bolds, paths []crSpan
	for i := 0; ; {
		s := strings.Index(line[i:], "**")
		if s < 0 {
			break
		}
		s += i
		e := strings.Index(line[s+2:], "**")
		if e < 0 {
			break
		}
		e += s + 2
		if txt := line[s+2 : e]; txt != "" && !strings.Contains(txt, "*") && crLooksLikeSection(txt) {
			bolds = append(bolds, crSpan{s, e + 1, txt})
		}
		i = e + 2
	}
	for i := 0; ; {
		s := strings.IndexByte(line[i:], '`')
		if s < 0 {
			break
		}
		s += i
		e := strings.IndexByte(line[s+1:], '`')
		if e < 0 {
			break
		}
		e += s + 1
		if txt := line[s+1 : e]; strings.HasSuffix(txt, ".md") || strings.HasSuffix(txt, ".mdc") {
			paths = append(paths, crSpan{s, e, txt})
		}
		i = e + 1
	}
	// A bold token belongs to at most ONE path: the nearest it is associated
	// with. Without this, a line naming .flow/project.md and then saying
	// "(see **Project configuration** in skills/.../project-configuration.md)"
	// would demand a "Project configuration" heading in BOTH files, and fail
	// on the one the token was never about.
	var pairs [][2]string
	for _, b := range bolds {
		best, bestLen := -1, -1
		for p, ps := range paths {
			var gap string
			switch {
			case b.end < ps.start:
				gap = line[b.end+1 : ps.start]
			case ps.end < b.start:
				gap = line[ps.end+1 : b.start]
			default:
				continue
			}
			if crAssociated(gap) && (bestLen < 0 || len(gap) < bestLen) {
				best, bestLen = p, len(gap)
			}
		}
		if best >= 0 && keep(b, paths[best]) {
			f := strings.Split(paths[best].text+"\t"+b.text, "\t")
			pairs = append(pairs, [2]string{f[0], f[1]})
		}
	}
	return pairs
}

// crRefusal is a checkFile failure that is not coverage_record's: the
// complete refusal line. The bash guard's read loop died under `set -e` on an
// unreadable file, exiting 1 with only bash's own error; 2 is this guard's
// cannot-answer code, so the port refuses with 2 and names the file.
type crRefusal string

func (e crRefusal) Error() string { return string(e) }

// checkFile is check_file: every associated reference in file, checked, and
// the file's own checked-reference count recorded.
func (r *crScan) checkFile(file string) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return crRefusal(fmt.Sprintf("check-references: cannot read %s: %v", r.rel(file), err))
	}
	checked, inFence := 0, false
	lines := crLines(b)
	// carry is whether the previous prose line of this paragraph left a bold
	// span open — what crJoined needs to find the right "**" pairing on the
	// next line, where crNormalizedLine only guesses from an odd count.
	carry := false
	for n, line := range lines {
		if crIsFence(line) {
			inFence = !inFence
			carry = false
			continue
		}
		if inFence {
			continue
		}
		// A citation split by a soft wrap (KAN-852, audit-finish D23): the
		// bold heading on this line and its path on the next, or a bold span
		// wrapped across the break. Only a pair that straddles the break is
		// checked here, so a pair wholly on one line is never checked twice.
		if n+1 < len(lines) {
			if pairs := crSplitPairs(line, lines[n+1], carry); len(pairs) > 0 {
				c, err := r.checkPairs(file, n+1, pairs)
				if err != nil {
					return err
				}
				checked += c
			}
		}
		carry = crCarry(line, carry)
		if strings.Contains(line, "refs-guard:allow") ||
			!strings.Contains(line, "**") || !strings.Contains(line, "`") {
			continue
		}
		pairs := crAssociations(crNormalizedLine(line))
		if len(pairs) == 0 {
			continue
		}
		c, err := r.checkPairs(file, n+1, pairs)
		if err != nil {
			return err
		}
		checked += c
	}
	return r.cov.record(r.rel(file), checked)
}

// checkPairs checks every path the pairs of one citation associate a token
// with, reported at lineno; it returns how many were CHECKED references.
func (r *crScan) checkPairs(file string, lineno int, pairs [][2]string) (int, error) {
	// The set of paths this line associates at least one bold token with.
	var paths []string
	for _, p := range pairs {
		paths = append(paths, p[0])
	}
	paths, err := crSort(r.env, paths, true)
	if err != nil {
		return 0, crRefusal(fmt.Sprintf("check-references: cannot sort the paths of %s:%d: %v", r.rel(file), lineno, err))
	}
	checked := 0
	for _, p := range paths {
		if p != "" && r.checkReference(file, lineno, p, pairs) {
			checked++
		}
	}
	return checked, nil
}

// crCarry is whether a bold span is still open after line, given whether one
// was open before it: a blank line ends the paragraph and every span in it.
func crCarry(line string, open bool) bool {
	if strings.TrimSpace(line) == "" {
		return false
	}
	if strings.Count(crMaskCodeSpans(line), "**")%2 == 1 {
		return !open
	}
	return open
}

// crSplitPairs is the (path, token) pairs a citation split across the break
// between line and next associates — only those whose bold and path spans
// straddle it. Neither line may be blank, a fence line or allow-marked. open
// is whether a bold span from an earlier line is still open when line
// starts: its closing "**" is dropped, so the pairing starts clean.
func crSplitPairs(line, next string, open bool) [][2]string {
	for _, l := range []string{line, next} {
		if strings.TrimSpace(l) == "" || crIsFence(l) || strings.Contains(l, "refs-guard:allow") {
			return nil
		}
	}
	// A soft wrap continues a paragraph; a heading or a table row never
	// wraps, and a list item on the next line starts a new block. A
	// blockquote's `>` markers are dropped from the continuation, so a quoted
	// paragraph joins as its reader sees it.
	cont := crQuote.ReplaceAllString(strings.TrimLeft(next, " \t"), "")
	if crBlockStart.MatchString(strings.TrimLeft(line, " \t>")) || crBlockStart.MatchString(cont) ||
		strings.HasPrefix(cont, "- ") || strings.HasPrefix(cont, "* ") || strings.HasPrefix(cont, "+ ") ||
		crNumbered.MatchString(cont) {
		return nil
	}
	a := crMaskCodeSpans(line)
	if open {
		a = strings.Replace(a, "**", "", 1)
	}
	joined := a + " " + crMaskCodeSpans(cont)
	if !strings.Contains(joined, "**") || !strings.Contains(joined, "`") {
		return nil
	}
	across := crAssociationsAcross(joined, len(a))
	if len(across) == 0 {
		return nil
	}
	// A path a straddling pair names is judged as the single-line check
	// judges one: any bold token the joined text associates with it may
	// resolve. So "per **The shape** (`x.md`): **Run this as a fix of
	// <wrap> `<name>`?**" passes on its live heading instead of failing on
	// the wrapped prompt text beside it.
	refs := map[string]bool{}
	for _, p := range across {
		refs[p[0]] = true
	}
	for _, p := range crAssociations(joined) {
		if refs[p[0]] {
			across = append(across, p)
		}
	}
	return across
}

// checkReference checks one path a line associates tokens with, reporting
// a stale or escaping reference; true when it was a CHECKED reference.
func (r *crScan) checkReference(file string, lineno int, ref string, pairs [][2]string) bool {
	// A citation may name this repository explicitly with a literal
	// `<agents repo>/` prefix, which names this repository's root. Strip it
	// before resolving, so a citation that carries that root is checked
	// exactly as it was before it carried one — without this, the citation
	// never resolves to a file and is silently skipped rather than checked.
	// `<project>/` is deliberately left untouched: it names a project this
	// repository cannot see, so it must fall through the ordinary
	// does-not-resolve path below rather than being stripped into something
	// that might resolve, or read as a traversal by contained.
	resolvePath := strings.TrimPrefix(ref, "<agents repo>/")
	// An absolute path is its own only candidate; a relative one is tried
	// against the repository root and against the referring file's directory.
	candidates := []string{resolvePath}
	if !strings.HasPrefix(resolvePath, "/") {
		candidates = []string{r.root + "/" + resolvePath, path.Dir(file) + "/" + resolvePath}
	}
	resolved, escaped := "", false
	for _, c := range candidates {
		safe, ok := r.contained(c)
		if !ok {
			escaped = true
			continue
		}
		if isFile(safe) {
			resolved = safe
			break
		}
	}
	// A reference pointing outside the repository is a defect in the
	// referring file, not a note in passing: this guard's contract is that a
	// clean exit means every reference was checked, and an unreadable one
	// was not.
	if resolved == "" && escaped {
		fmt.Fprintf(r.stdout, "%s:%d: reference %s resolves outside the repository root; not read\n", r.rel(file), lineno, ref)
		r.failures++
		return false
	}
	// A path that resolves to nothing is out of scope: templated paths like
	// spectre/changes/<name>/tasks.md are legitimate and must not fail the
	// guard.
	if resolved == "" {
		return false
	}
	heads, ok := crHeadings(resolved)
	if !ok {
		return false
	}
	// This is a CHECKED reference: a bold token associated with a path that
	// resolved to a real, headed file, and was actually compared against that
	// file's headings — see scripts/lib/coverage.sh's header for why this
	// count, not the mere fact that checkFile ran, is what "coverage" means
	// here. A file with none of these is a file this guard read but verified
	// nothing in, indistinguishable from the outside from a rule that
	// silently stopped checking it.
	for _, p := range pairs {
		if p[0] != ref || p[1] == "" {
			continue
		}
		// The path itself is often bolded; that is a file reference, not a
		// section name, so it never counts as a match.
		if token := crNormalize(p[1]); !strings.Contains(token, "/") && heads[token] {
			return true
		}
	}
	fmt.Fprintf(r.stdout, "%s:%d: no bold token resolves to a heading in %s\n", r.rel(file), lineno, ref)
	r.failures++
	return true
}
