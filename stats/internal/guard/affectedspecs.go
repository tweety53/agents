package guard

import (
	"cmp"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

// flow-guard affected-specs [-task <tasks.md>:<task-id>] <source-root> <base> <spec-root>
//
// Prints the capture specs a diff can move, one path per line, so a fixer
// recaptures those and no others (skills/flow/verify-fix-loop.md; kan-1099:
// every fix round recaptured every spec, 34-36 minutes a round).
//
// THE DIFF is `git -C <source-root> diff --name-only <base>`: the working
// tree against <base>, uncommitted edits and untracked files that are not
// ignored included, or a commit range when <base> is `<a>..<b>`. A path under a `*-snapshots/` directory and a
// `full-app-suite.zip` are a capture's output and are skipped.
//
// THE SPECS are every `*.spec.{ts,tsx,js,jsx,mjs,cjs}` under <spec-root> —
// the capture root, or any directory of it — `node_modules`, `.git`,
// `.worktrees`, `test-results` and `*-snapshots` pruned. Each is printed
// relative to the nearest directory at or above it holding a `package.json`
// — where the project's `capture` command runs Playwright — else relative to
// <spec-root>.
//
// THE MAP is a `<spec>.sources` sidecar beside the spec: one glob per line,
// relative to <source-root>, in `ui paths`' own glob syntax, naming the
// source files that spec's screen is made of; `#` lines and blank lines are
// skipped. A spec's roots are the spec file itself and every file its globs
// match. A glob ending `:<line>[,<line>…]` names the tests of the spec its
// files reach, each by the line its `test(` call opens on: a changed file
// that reaches the spec through such globs alone selects those tests, each
// printed `<spec>:<line>`, and the spec whole through any other.
//
// An `all.sources` beside the specs holds the same lines for what every spec
// renders and no import chain shows — a seed, a fixture, the runner's
// config: a changed file matching one selects every spec.
//
// A changed file selects a spec when it is that spec, its `.sources` or
// `.mockups` sidecar, one of its roots, or a file one of its roots imports
// through files no sidecar lists — so a per-screen file selects its own spec
// and a shared file (a header, a chip, tokens) selects every spec whose
// screen imports it. The climb from a changed file stops at the first root
// on each chain: a sidecar lists every file of its screen that another
// listed file imports, since a hub that imports every screen would otherwise
// join them all.
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
// when a chain of its importers ends, before any spec's root, at a UI file
// nothing imports: an entry point draws it in the shell around every
// screen. A UI change also selects every spec with no `.sources`
// sidecar, whose screen is unknown. A changed file outside `ui paths` that
// no spec reaches selects nothing and is named `unreached: <path>`. A changed
// file inside `<spec-root>` is never a UI change: its reach is its own
// selection and the import climb (kan-1099 self-review, F2).
//
// THE TASK SET: with `-task` the list is what an in-run fix task's own
// capture run takes (skills/flow/verify-fix-loop.md step 2; kan-1091: a
// fixture edit selected 130 specs inside the task, 36 minutes a run). A
// changed file selects a spec only when it is that spec, its `.sources` or
// `.mockups` sidecar, or a file one of its `.sources` globs matches -- no
// import climb, no `all.sources`, no safe side, nothing on stderr but `test:`
// lines -- and every spec whose printed path the task's **Tests:** or
// **Regression:** field names, behind any leading directories, is selected
// too -- the one test where the field follows the path with `:<line>`
// (kan-1091: five specs named whole ran 164 tests, 31-34 minutes).
//
// A TEST FILE outside every directory holding a spec (asTestRE) moves no capture: changed, it
// is named `test: <path>` and skipped, and it is never an importer.
//
// Exit codes:
//
//	0  the list is on stdout, sorted by spec then line, possibly empty.
//	2  usage, a root that is not a directory, a diff git cannot read, or a
//	   `-task` whose file or task cannot be read.
func init() { Registry["affected-specs"] = affectedSpecs }

// asExts are the extensions the import walk reads.
var asExts = map[string]bool{".kt": true, ".ts": true, ".tsx": true, ".js": true, ".jsx": true, ".mjs": true, ".cjs": true, ".css": true, ".scss": true, ".vue": true, ".svelte": true}

var (
	asSpecRE = regexp.MustCompile(`\.spec\.[cm]?[jt]sx?$`)
	// asTestRE is a test file outside the spec root: a test directory, a
	// `.test.`/`.spec.`/`_test.` name, or a `…Test(s).<ext>` stem.
	asTestRE = regexp.MustCompile(`(?i:(?:^|/)[^/]*tests?/|[._-](?:test|spec)\.)|[a-z]Tests?\.\w+$`)
	// asLinesRE is a sidecar glob's `:<line>[,<line>…]` suffix.
	asLinesRE  = regexp.MustCompile(`:(\d+(?:,\d+)*)$`)
	asWordRE   = regexp.MustCompile(`[A-Za-z_]\w{2,}`)
	asImportRE = regexp.MustCompile(`(?m)^.*(?:\bimport\b|\bfrom\b|\brequire\b|@import).*$`)
	// asKotlinDeclRE captures a top-level declared Kotlin name; asQuotedRE a
	// quoted module path.
	asKotlinDeclRE = regexp.MustCompile(`(?m)^(?:@\w+(?:\([^)\n]*\))?\s+)*(?:(?:public|internal|data|sealed|open|abstract|enum|annotation|value|inline|const|suspend|actual|expect)\s+)*(?:class|object|interface|fun|val|var|typealias)\s+(?:<[^>\n]+>\s*)?(?:[\w<>?, ]+\.)?([A-Za-z_]\w{2,})`)
	asQuotedRE     = regexp.MustCompile("[\"'`]([^\"'`\\s]+)[\"'`]")
)

// asGlob is one sidecar line: its glob, and the test lines a `:<line>,…`
// suffix names -- nil for the whole spec.
type asGlob struct {
	re    *regexp.Regexp
	lines []string
}

// asHit reports whether a glob of globs matches path, and the test lines
// the matching globs name -- nil when one of them names the whole spec.
func asHit(globs []asGlob, path string) (hit bool, lines []string) {
	p, whole := vtBytesAsRunes(path), false
	for _, g := range globs {
		if !g.re.MatchString(p) {
			continue
		}
		hit, whole = true, whole || g.lines == nil
		lines = append(lines, g.lines...)
	}
	if whole {
		lines = nil
	}
	return hit, lines
}

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
	named, task := "", len(args) == 5 && args[0] == "-task"
	if task {
		var ok bool
		if named, ok = asTaskNames(args[1]); !ok {
			fmt.Fprintf(stderr, "affected-specs: cannot read the task %s\n", args[1])
			Reason(env, "cannot-read-task")
			return 2
		}
		args = args[2:]
	}
	if len(args) != 3 {
		fmt.Fprintln(stderr, "usage: flow-guard affected-specs [-task <tasks.md>:<task-id>] <source-root> <base> <spec-root>")
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
	if !strings.Contains(args[1], "..") {
		diff += "\n" + asUntracked(env, src)
	}
	tracked, _ := capture(git("-C", src, "ls-files"))

	type spec struct {
		out, rel string // printed path; path relative to src, "" outside it
		globs    []asGlob
		mapped   bool
	}
	var specDirs []string // every directory holding a spec
	var specs []spec
	_ = filepath.WalkDir(specRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if n := d.Name(); n == "node_modules" || n == ".git" || n == ".worktrees" || n == "test-results" || strings.HasSuffix(n, "-snapshots") {
				return filepath.SkipDir
			}
			return nil
		}
		if !asSpecRE.MatchString(d.Name()) {
			return nil
		}
		s := spec{}
		s.out = asSpecOut(specRoot, p)
		if !slices.Contains(specDirs, filepath.Dir(p)) {
			specDirs = append(specDirs, filepath.Dir(p))
		}
		if rel, err := filepath.Rel(src, p); err == nil && !strings.HasPrefix(rel, "..") {
			s.rel = rel
		}
		if isFile(p + ".sources") {
			s.mapped, s.globs = true, asGlobs(p+".sources")
		}
		specs = append(specs, s)
		return nil
	})

	var everySpec []asGlob
	for _, d := range specDirs {
		everySpec = append(everySpec, asGlobs(filepath.Join(d, "all.sources"))...)
	}
	// `ui paths`, nil where the row is absent: every changed file is UI then.
	var ui []asGlob
	if text, err := os.ReadFile(filepath.Join(src, ".flow/project.md")); err == nil {
		for _, g := range strings.Split(vvCells(string(text))["ui paths"], ",") {
			if g = trimGlobElement(g); g != "" {
				ui = append(ui, asGlob{re: vtGlobRegexp(g)})
			}
		}
	}
	matches := func(globs []asGlob, path string) bool {
		hit, _ := asHit(globs, path)
		return hit
	}
	isUI := func(f string) bool { return ui == nil || matches(ui, f) }

	// inSpecRoot is a file in a directory holding a spec, or below one: the
	// suite's own helpers, never the application's unit tests.
	inSpecRoot := func(f string) bool {
		abs := filepath.Join(src, f)
		return slices.ContainsFunc(specDirs, func(d string) bool {
			return strings.HasPrefix(abs, d+string(filepath.Separator))
		})
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
	// reach is f and every file that imports it, climbing no further than a
	// spec's root: the screen a sidecar names is where a chain ends. shell
	// is true when a chain ends instead at a UI file nothing imports -- an
	// entry point: f is drawn outside every mapped screen, in the shell
	// around them all.
	reach := func(f string) (seen map[string]bool, shell bool) {
		seen = map[string]bool{f: true}
		for frontier := []string{f}; len(frontier) > 0; {
			var next []string
			for _, target := range frontier {
				if isRoot(target) {
					continue
				}
				in := importers(target)
				shell = shell || (len(in) == 0 && isUI(target) && !inSpecRoot(target))
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

	// selected is, per spec, the test lines picked; "" is the spec whole.
	selected := map[string]map[string]bool{}
	pick := func(out string, lines ...string) {
		if selected[out] == nil {
			selected[out] = map[string]bool{}
		}
		if len(lines) == 0 {
			lines = []string{""}
		}
		for _, l := range lines {
			selected[out][l] = true
		}
	}
	for _, s := range specs {
		if !task {
			break
		}
		for _, m := range regexp.MustCompile(`(?:^|[^\w.-])`+regexp.QuoteMeta(s.out)+`(?::(\d+))?(?:$|[^\w/-])`).FindAllStringSubmatch(named, -1) {
			pick(s.out, strings.Fields(m[1])...)
		}
	}
	for _, f := range strings.Split(diff, "\n") {
		if f == "" || filepath.Base(f) == "full-app-suite.zip" || strings.Contains(filepath.Dir(f)+"/", "-snapshots/") {
			continue
		}
		if !inSpecRoot(f) && asTestRE.MatchString(f) {
			fmt.Fprintf(stderr, "test: %s\n", f)
			continue
		}
		if task {
			for _, s := range specs {
				if s.rel != "" && (f == s.rel || f == s.rel+".sources" || f == s.rel+".mockups") {
					pick(s.out)
				} else if hit, lines := asHit(s.globs, f); hit {
					pick(s.out, lines...)
				}
			}
			continue
		}
		if matches(everySpec, f) {
			fmt.Fprintf(stderr, "all: %s — listed in all.sources\n", f)
			for _, s := range specs {
				pick(s.out)
			}
			continue
		}
		ui := isUI(f) && !inSpecRoot(f)
		seen, shell := reach(f)
		hit := false
		for _, s := range specs {
			// whole: the spec itself, or a root reached through a glob naming
			// no test; lines otherwise.
			whole := s.rel != "" && (seen[s.rel] || f == s.rel+".sources" || f == s.rel+".mockups")
			var lines []string
			for g := range seen {
				if ok, l := asHit(s.globs, g); ok {
					whole, lines = whole || l == nil, append(lines, l...)
				}
			}
			own := whole || len(lines) > 0
			switch {
			case whole || (ui && !s.mapped):
				pick(s.out)
			case own:
				pick(s.out, lines...)
			}
			hit = hit || own
		}
		switch {
		case ui && (shell || !hit):
			fmt.Fprintf(stderr, "all: %s — a UI path drawn outside every spec's sources\n", f)
			for _, s := range specs {
				pick(s.out)
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
		if selected[s][""] {
			fmt.Fprintln(stdout, s)
			continue
		}
		lines := slices.Collect(maps.Keys(selected[s]))
		slices.SortFunc(lines, func(a, b string) int { return cmp.Or(len(a)-len(b), strings.Compare(a, b)) })
		for _, l := range lines {
			fmt.Fprintln(stdout, s+":"+l)
		}
	}
	return 0
}

// asSpecOut is the path a spec is printed as: relative to the nearest
// directory at or above it holding a package.json, else to specRoot.
func asSpecOut(specRoot, spec string) string {
	pkg := specRoot
	for d := filepath.Dir(spec); ; d = filepath.Dir(d) {
		if isFile(filepath.Join(d, "package.json")) {
			pkg = d
			break
		}
		if d == filepath.Dir(d) {
			break
		}
	}
	out, _ := filepath.Rel(pkg, spec)
	return out
}

// asTaskNames is the text of the **Tests:** and **Regression:** fields of the
// task a `<tasks.md>:<task-id>` argument names; false when it cannot be read.
func asTaskNames(arg string) (string, bool) {
	i := strings.LastIndex(arg, ":")
	if i < 0 {
		return "", false
	}
	raw, err := os.ReadFile(arg[:i])
	if err != nil {
		return "", false
	}
	lines, err := tcfPlanLines(raw)
	if err != nil {
		return "", false
	}
	found := tcfSelectTask(lines, arg[i+1:])
	if found == nil {
		return "", false
	}
	f := tcfBodyFields(found.lines)
	return strings.Join(append(f["Tests"], f["Regression"]...), " "), true
}

// asUntracked is the worktree's untracked, unignored paths, newline-joined.
func asUntracked(env Env, worktree string) string {
	out, _ := capture(envGit(env)("-C", worktree, "ls-files", "--others", "--exclude-standard"))
	return out
}

// asGlobs reads a sidecar's globs: one per line, `#` and blank lines skipped,
// a `:<line>,…` suffix kept as the glob's test lines; nil when it cannot be
// read.
func asGlobs(path string) []asGlob {
	raw, _ := os.ReadFile(path)
	var out []asGlob
	for _, l := range strings.Split(string(raw), "\n") {
		if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "#") {
			g := asGlob{}
			if m := asLinesRE.FindStringSubmatch(l); m != nil {
				l, g.lines = l[:len(l)-len(m[0])], strings.Split(m[1], ",")
			}
			g.re = vtGlobRegexp(l)
			out = append(out, g)
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
