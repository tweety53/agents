package guard

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

// flow-guard affected-specs <source-root> <base> <spec-root>
//
// Prints the capture specs a diff can move, one path per line, so a fixer
// recaptures those and no others (skills/flow/verify-fix-loop.md; kan-1099:
// every fix round recaptured every spec, 34-36 minutes a round).
//
// THE DIFF is `git -C <source-root> diff --name-only <base>`: the working
// tree against <base>, uncommitted edits included, or a commit range when
// <base> is `<a>..<b>`. A path under a `*-snapshots/` directory and a
// `full-app-suite.zip` are a capture's output and are skipped.
//
// THE SPECS are every `*.spec.{ts,tsx,js,jsx,mjs,cjs}` under <spec-root>,
// `node_modules`, `.git` and `*-snapshots` pruned. Each is printed relative to
// the nearest directory at or above <spec-root> holding a `package.json` —
// where the project's `capture` command runs — else relative to <spec-root>.
//
// THE MAP is a `<spec>.sources` sidecar beside the spec: one glob per line,
// relative to <source-root>, in `ui paths`' own glob syntax, naming the
// source files that spec's screen is made of; `#` lines and blank lines are
// skipped. A spec's roots are the spec file itself and every file its globs
// match.
//
// `<spec-root>/all.sources` holds the same lines for what every spec
// renders and no import chain shows — a seed, a fixture, the runner's
// config: a changed file matching one selects every spec.
//
// A changed file selects a spec when it is that spec, its `.sources` or
// `.mockups` sidecar, one of its roots, or a file one of its roots imports,
// transitively — so a per-screen file selects its own spec and a shared file
// (a header, a chip, tokens) selects every spec whose screen imports it.
// File A imports file B when a line of A that imports (`import`, `from`,
// `require`, `@import`) quotes a module path whose last segment, extension
// dropped, is B's stem — the directory's name for an `index.*`. A Kotlin
// file imports a Kotlin file B when one of its `import` lines names B's stem
// or a name B declares at column 0 and not `private`, or ends in `*` and
// names B's directory, or when it sits in B's directory — one package — and
// its text holds such a name as a word. Only tracked files with a source
// extension (asExts) are read.
// ponytail: text matching, not a compiler's resolution — two files sharing a
// stem over-select; a Kotlin file split across packages in one directory, or
// reached only through a typealias in another package, is the ceiling, and
// its upgrade is the build's own module graph.
//
// THE SAFE SIDE: a changed file matching `<source-root>/.flow/project.md`'s
// `ui paths` — every changed file where that row is absent — selects EVERY
// spec, named on stderr as `all: <path> — …`, when no spec reaches it, or
// when a chain of its importers that passes through no spec's root ends at
// a UI file nothing imports: an entry point draws it in the shell around
// every screen. A UI change also selects every spec with no `.sources`
// sidecar, whose screen is unknown. A changed file outside `ui paths` that
// no spec reaches selects nothing and is named `unreached: <path>`.
//
// A TEST FILE outside <spec-root> (asTestRE) moves no capture: changed, it
// is named `test: <path>` and skipped, and it is never an importer.
//
// Exit codes:
//
//	0  the list is on stdout, sorted, possibly empty.
//	2  usage, a root that is not a directory, or a diff git cannot read.
func init() { Registry["affected-specs"] = affectedSpecs }

// asExts are the extensions the import walk reads.
var asExts = map[string]bool{".kt": true, ".ts": true, ".tsx": true, ".js": true, ".jsx": true, ".mjs": true, ".cjs": true, ".css": true, ".scss": true, ".vue": true, ".svelte": true}

var (
	asSpecRE = regexp.MustCompile(`\.spec\.[cm]?[jt]sx?$`)
	// asTestRE is a test file outside the spec root: a test directory, a
	// `.test.`/`.spec.`/`_test.` name, or a `…Test(s).<ext>` stem.
	asTestRE   = regexp.MustCompile(`(?i:(?:^|/)[^/]*tests?/|[._-](?:test|spec)\.)|[a-z]Tests?\.\w+$`)
	asWordRE   = regexp.MustCompile(`[A-Za-z_]\w{2,}`)
	asImportRE = regexp.MustCompile(`(?m)^.*(?:\bimport\b|\bfrom\b|\brequire\b|@import).*$`)
	// asKotlinDeclRE captures a top-level declared Kotlin name; asQuotedRE a
	// quoted module path.
	asKotlinDeclRE = regexp.MustCompile(`(?m)^(?:@\w+(?:\([^)\n]*\))?\s+)*(?:(?:public|internal|data|sealed|open|abstract|enum|annotation|value|inline|const|suspend|actual|expect)\s+)*(?:class|object|interface|fun|val|var|typealias)\s+(?:<[^>\n]+>\s*)?(?:[\w<>?, ]+\.)?([A-Za-z_]\w{2,})`)
	asQuotedRE     = regexp.MustCompile("[\"'`]([^\"'`\\s]+)[\"'`]")
)

// asFile is one source file's index for the import walk.
type asFile struct {
	words, imports map[string]bool // every word; the words on importing lines
	mods           map[string]bool // the stems of the module paths it quotes on importing lines
	wildcard       bool            // an import line ends in `*`
	names          []string        // its stem and the names it declares
}

// asImports reports whether file c, at path cp, imports target t at path tp.
func asImports(cp string, c *asFile, tp string, t *asFile) bool {
	if filepath.Ext(tp) != ".kt" {
		return c.mods[t.names[0]]
	}
	if filepath.Ext(cp) != ".kt" {
		return false
	}
	named := func(set map[string]bool) bool {
		return slices.ContainsFunc(t.names, func(n string) bool { return set[n] })
	}
	return named(c.imports) || (c.wildcard && c.imports[filepath.Base(filepath.Dir(tp))]) ||
		(filepath.Dir(cp) == filepath.Dir(tp) && named(c.words))
}

func affectedSpecs(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 3 {
		fmt.Fprintln(stderr, "usage: flow-guard affected-specs <source-root> <base> <spec-root>")
		Reason(env, "usage")
		return 2
	}
	src, specRoot := filepath.Clean(pcAbs(env, args[0])), filepath.Clean(pcAbs(env, args[2]))
	for _, d := range []string{src, specRoot} {
		if !isDir(d) {
			fmt.Fprintf(stderr, "affected-specs: %s is not a directory\n", d)
			Reason(env, "not-a-directory")
			return 2
		}
	}
	git := envGit(env)
	diff, ok := capture(git("-C", src, "diff", "--name-only", "--no-renames", args[1]))
	if !ok {
		fmt.Fprintf(stderr, "affected-specs: git cannot diff %s against %s\n", src, args[1])
		Reason(env, "cannot-read-diff")
		return 2
	}
	tracked, _ := capture(git("-C", src, "ls-files"))

	type spec struct {
		out, rel string // printed path; path relative to src, "" outside it
		globs    []*regexp.Regexp
		mapped   bool
	}
	pkg := specRoot
	for d := specRoot; ; d = filepath.Dir(d) {
		if isFile(filepath.Join(d, "package.json")) {
			pkg = d
			break
		}
		if d == filepath.Dir(d) {
			break
		}
	}
	var specs []spec
	_ = filepath.WalkDir(specRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if n := d.Name(); n == "node_modules" || n == ".git" || strings.HasSuffix(n, "-snapshots") {
				return filepath.SkipDir
			}
			return nil
		}
		if !asSpecRE.MatchString(d.Name()) {
			return nil
		}
		s := spec{}
		s.out, _ = filepath.Rel(pkg, p)
		if rel, err := filepath.Rel(src, p); err == nil && !strings.HasPrefix(rel, "..") {
			s.rel = rel
		}
		if isFile(p + ".sources") {
			s.mapped, s.globs = true, asGlobs(p+".sources")
		}
		specs = append(specs, s)
		return nil
	})

	everySpec := asGlobs(filepath.Join(specRoot, "all.sources"))
	// `ui paths`, nil where the row is absent: every changed file is UI then.
	var ui []*regexp.Regexp
	if text, err := os.ReadFile(filepath.Join(src, ".flow/project.md")); err == nil {
		for _, g := range strings.Split(vvCells(string(text))["ui paths"], ",") {
			if g = trimGlobElement(g); g != "" {
				ui = append(ui, vtGlobRegexp(g))
			}
		}
	}
	matches := func(res []*regexp.Regexp, path string) bool {
		p := vtBytesAsRunes(path)
		return slices.ContainsFunc(res, func(re *regexp.Regexp) bool { return re.MatchString(p) })
	}
	isUI := func(f string) bool { return ui == nil || matches(ui, f) }

	inSpecRoot := func(f string) bool {
		abs := filepath.Join(src, f)
		return abs == specRoot || strings.HasPrefix(abs, specRoot+string(filepath.Separator))
	}
	index := map[string]*asFile{}
	for _, f := range strings.Split(tracked, "\n") {
		if asExts[filepath.Ext(f)] && (inSpecRoot(f) || !asTestRE.MatchString(f)) {
			index[f] = asIndex(src, f)
		}
	}
	isRoot := func(f string) bool {
		return slices.ContainsFunc(specs, func(s spec) bool { return f == s.rel || matches(s.globs, f) })
	}
	importers := func(target string) []string {
		t := index[target]
		if t == nil {
			t = asIndex(src, target) // deleted or unlisted: its stem still names it
		}
		var out []string
		for c, file := range index {
			if c != target && asImports(c, file, target, t) {
				out = append(out, c)
			}
		}
		return out
	}
	// reach is f and every file that imports it, transitively. shell is true
	// when a chain of importers that passes through no spec's root ends at a
	// UI file nothing imports -- an entry point: f is drawn outside every
	// mapped screen, in the shell around them all.
	reach := func(f string) (seen map[string]bool, shell bool) {
		seen = map[string]bool{f: true}
		outside := map[string]bool{f: true} // reached through no root
		for frontier := []string{f}; len(frontier) > 0; {
			var next []string
			for _, target := range frontier {
				in := importers(target)
				if outside[target] && !isRoot(target) && !inSpecRoot(target) {
					shell = shell || (len(in) == 0 && isUI(target))
					for _, c := range in {
						outside[c] = true
					}
				}
				for _, c := range in {
					if !seen[c] {
						seen[c] = true
						next = append(next, c)
					}
				}
			}
			frontier = next
		}
		return seen, shell
	}

	selected := map[string]bool{}
	for _, f := range strings.Split(diff, "\n") {
		if f == "" || filepath.Base(f) == "full-app-suite.zip" || strings.Contains(filepath.Dir(f)+"/", "-snapshots/") {
			continue
		}
		if !inSpecRoot(f) && asTestRE.MatchString(f) {
			fmt.Fprintf(stderr, "test: %s\n", f)
			continue
		}
		if matches(everySpec, f) {
			fmt.Fprintf(stderr, "all: %s — listed in all.sources\n", f)
			for _, s := range specs {
				selected[s.out] = true
			}
			continue
		}
		ui := isUI(f)
		seen, shell := reach(f)
		hit := false
		for _, s := range specs {
			own := s.rel != "" && (seen[s.rel] || f == s.rel+".sources" || f == s.rel+".mockups")
			for g := range seen {
				own = own || matches(s.globs, g)
			}
			if own || (ui && !s.mapped) {
				selected[s.out] = true
			}
			hit = hit || own
		}
		switch {
		case ui && (shell || !hit):
			fmt.Fprintf(stderr, "all: %s — a UI path drawn outside every spec's sources\n", f)
			for _, s := range specs {
				selected[s.out] = true
			}
		case !hit:
			fmt.Fprintf(stderr, "unreached: %s\n", f)
		}
	}
	out := make([]string, 0, len(selected))
	for s := range selected {
		out = append(out, s)
	}
	sort.Strings(out)
	for _, s := range out {
		fmt.Fprintln(stdout, s)
	}
	return 0
}

// asGlobs reads a sidecar's globs: one per line, `#` and blank lines skipped;
// nil when it cannot be read.
func asGlobs(path string) []*regexp.Regexp {
	raw, _ := os.ReadFile(path)
	var out []*regexp.Regexp
	for _, l := range strings.Split(string(raw), "\n") {
		if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "#") {
			out = append(out, vtGlobRegexp(l))
		}
	}
	return out
}

// asIndex reads rel under src into its word sets and names; an unreadable
// file keeps its stem alone.
func asIndex(src, rel string) *asFile {
	stem := strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))
	if stem == "index" {
		stem = filepath.Base(filepath.Dir(rel))
	}
	f := &asFile{words: map[string]bool{}, imports: map[string]bool{}, mods: map[string]bool{}, names: []string{stem}}
	raw, err := os.ReadFile(filepath.Join(src, rel))
	if err != nil {
		return f
	}
	text := string(raw)
	for _, w := range asWordRE.FindAllString(text, -1) {
		f.words[w] = true
	}
	for _, line := range asImportRE.FindAllString(text, -1) {
		f.wildcard = f.wildcard || strings.HasSuffix(strings.TrimSpace(line), "*")
		for _, w := range asWordRE.FindAllString(line, -1) {
			f.imports[w] = true
		}
		for _, m := range asQuotedRE.FindAllStringSubmatch(line, -1) {
			mod := filepath.Base(m[1])
			f.mods[strings.TrimSuffix(mod, filepath.Ext(mod))] = true
		}
	}
	if filepath.Ext(rel) != ".kt" {
		return f
	}
	for _, m := range asKotlinDeclRE.FindAllStringSubmatch(text, -1) {
		f.names = append(f.names, m[1])
	}
	return f
}
