package guard

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"
)

// checkInstalledCitations is scripts/check-installed-citations.sh: that
// script's header comment is the contract -- fail when a path citation in a
// file setup.sh installs names no root. Exit 0 clean (every citation names a
// root, every scanned member's coverage non-zero or declared), 1 violations
// (an unrooted citation, or an undeclared zero), 2 cannot answer (the root
// override set but empty or not a directory, the sandboxed installer run
// failed, or a file could not be read or decoded — reason on stderr, NOTHING
// on stdout, so no caller can read an absent report as a clean one).
//
// It was a bash wrapper around check-installed-citations.py (3e48ecac): the
// Python classified and handed back violation lines plus a root and
// per-member coverage behind sentinel prefixes, because only the wrapper
// could source scripts/lib/coverage.sh. Here the classifier and the coverage
// decision (coverage.go, coverage.sh's twin) are one process, so the
// sentinel hand-off is gone; the reasoning the two files carried sits beside
// the code it explains.
//
// A file setup.sh installs or copies is read by an agent standing in
// whatever tree the harness placed it — never necessarily this checkout. A
// bare citation like `README.md` in such a file is read against THAT tree,
// which is exactly how a bug ships: the citation resolves to nothing, or
// worse, to some unrelated file the target project happens to have. Exactly
// three FORMS of root are recognised (canonical definition: the kan-239
// citation-roots spec — do not restate the rule here):
//
//   - an installed root's own bare form (`skills/…`, `rules/…`,
//     `commands-claude/…`, `hooks/…`, wherever the harness placed it)
//   - a placeholder root — a CLOSED set, see cicPlaceholderRoots for the set
//     itself and why membership rather than bracket shape is what a
//     placeholder root requires
//   - (nothing else — anything not matching one of the above is a violation)
func init() {
	Registry["check-installed-citations"] = checkInstalledCitations
}

const cicName = "check-installed-citations"

// cicAllowMarker is CITATION_ALLOW_MARKER — the declared, per-line exemption
// for non-path tokens (KAN-619), the sibling of check-references's
// `refs-guard:allow`. A line carrying this substring anywhere in its raw
// text is skipped wholesale: no spans extracted, no words merged, no
// candidates, no coverage. It exists because two classes of corpus line are
// illustrations, not citations, and both used to cost a rewording commit
// every time the lint list caught them: a backticked `n/a`-style status
// marker (kan-548 — every word of a backtick span is a candidate, and a
// slashed abbreviation's first segment names no root), and a command shape
// quoting a placeholder list (kan-561 — the merged `<…>`-phrase token names
// no root). The classifier's own exclusions stay closed, anchored shapes
// that fail closed; the marker is the writeable escape hatch — write the
// shape as-is and declare the line. Line-scoped by construction: a marked
// line exempts only itself, and a member exempted down to zero checked
// citations still answers to the declared-zero discipline (cicExpectedZero).
const cicAllowMarker = "citations-guard:allow"

// cicExpectedZero — EXPECTED-ZERO MEMBERS, established by reading each
// file's own content (2026-08-20, at that change's own HEAD), never
// guessed: none of these carries a single backticked path (or, inside a
// bash/sh/zsh fence comment, a bare shell-word path) in any shape this
// guard's classifier recognises, so it genuinely verifies nothing inside
// them. Declared here, once, rather than inferred from the tree — inferring
// it would restate the very assumption a silently uncovered file already
// encodes.
//
// Each is declared ONLY when it is part of THIS run's corpus (which may be
// a sandboxed CHECK_INSTALLED_CITATIONS_ROOT fixture, not this repository),
// as guardsymlinks.go's declare_if_present port does: declaring a
// name outside the current corpus would make it a KAN-197 F3 "declared but
// never recorded" violation for every fixture that does not carry that file.
var cicExpectedZero = [][2]string{
	{"rules/be-brief.mdc", "always-on rule body — cites no .md/.mdc path at all, backticked or bare"},
	{"rules/build-the-simplest-thing.mdc", "always-on rule body — cites no .md/.mdc path at all, backticked or bare"},
	{"rules/context7.mdc", "always-on rule body — cites no .md/.mdc path at all, backticked or bare"},
	{"rules/dependency-versions.mdc", "always-on rule body — cites no .md/.mdc path at all, backticked or bare"},
	{"rules/design-mockups-are-specs.mdc", "always-on rule body — cites no .md/.mdc path at all, backticked or bare"},
	{"rules/fix-determinism-at-the-source.mdc", "always-on rule body — cites no .md/.mdc path at all, backticked or bare"},
	{"rules/never-touch-production.mdc", "always-on rule body — cites no .md/.mdc path at all, backticked or bare"},
	{"skills/flow/engineering-principles.md", "reviewer-prompt file, deliberately self-contained — cites principles-reviewer-prompt.md only via a Markdown link, a shape this guard's classifier does not scan"},
	{"commands-claude/flow.md", "command-dispatch stub — delegates to the flow skill by name, not by path; cites no .md/.mdc path at all"},
	{"commands-claude/flow-plan.md", "command-dispatch stub — delegates to the flow-plan skill by name, not by path; cites no .md/.mdc path at all"},
	{"commands-claude/flow-settings.md", "command-dispatch stub — delegates to the flow-settings skill by name, not by path; cites no .md/.mdc path at all"},
	{"commands-claude/flow-self-review.md", "command-dispatch stub — delegates to the flow-self-review skill by name, not by path; cites no .md/.mdc path at all"},
	{"skills/flow-contracts/plan-provenance.md", "the guard-facing sections moved to plan-provenance-guard.md — cites no .md/.mdc path at all"},
	{"agents/flow-low.md", "generic dispatch-target agent definition — cites no .md/.mdc path at all"},
	{"agents/flow-medium.md", "generic dispatch-target agent definition — cites no .md/.mdc path at all"},
	{"agents/flow-high.md", "generic dispatch-target agent definition — cites no .md/.mdc path at all"},
}

func checkInstalledCitations(_ []string, env Env, stdout, stderr io.Writer) int {
	// The root is this guard's own checkout by default — the physical
	// parent of scripts/, which has exactly one home, never symlinked into
	// a skill's own directory (the farm distributes skills/, not scripts/),
	// so it never answers a skill directory by accident. The binary lives
	// in a cache, so its shim exports it as FLOW_GUARD_REPO_ROOT.
	// CHECK_INSTALLED_CITATIONS_ROOT is an explicit, opt-in override honoured
	// only when set, solely so the tests can point this guard at a sandboxed
	// fixture tree without touching this repository — never set it for a
	// normal invocation.
	root, set := env.LookupEnv("CHECK_INSTALLED_CITATIONS_ROOT")
	switch {
	case set && root == "":
		fmt.Fprintf(stderr, "%s: CHECK_INSTALLED_CITATIONS_ROOT is set but empty\n", cicName)
		return 2
	case !set:
		if root = env.Getenv("FLOW_GUARD_REPO_ROOT"); root == "" {
			fmt.Fprintf(stderr, "%s: FLOW_GUARD_REPO_ROOT is unset — run scripts/check-installed-citations.sh, which sets it\n", cicName)
			return 2
		}
	}
	abs := root
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(env.Dir, abs)
	}
	if !isDir(abs) {
		fmt.Fprintf(stderr, "%s: %s is not a directory — cannot scan\n", cicName, root)
		return 2
	}
	repo, err := filepath.EvalSymlinks(abs)
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot resolve %s — %v\n", cicName, root, err)
		return 2
	}

	sources, code := cicInstalledSources(env, repo, stderr)
	if code != 0 {
		return code
	}
	roots := map[string]bool{}
	for s := range sources {
		if first, _, ok := strings.Cut(s, "/"); ok {
			roots[first] = true
		}
	}
	corpus := cicCorpus(repo, sources)
	rootFiles := map[string]bool{}
	for _, name := range cicListdir(repo) {
		if fi, err := os.Lstat(filepath.Join(repo, name)); err == nil && fi.Mode().IsRegular() {
			rootFiles[name] = true
		}
	}

	var violations []string
	var cov coverage
	for _, rel := range corpus {
		text, msg := cicRead(repo, rel)
		if msg != "" {
			fmt.Fprintf(stderr, "%s: %s\n", cicName, msg)
			return 2
		}
		checked := 0
		for _, c := range cicCandidates(text) {
			if !cicIsCitation(c.token, rootFiles) {
				continue
			}
			checked++
			if !cicNamesRoot(c.token, roots) {
				violations = append(violations, fmt.Sprintf("%s:%d: citation `%s` names no root — prefix with <agents repo>/ or <project>/", rel, c.line, c.token))
			}
		}
		if err := cov.record(rel, checked); err != nil {
			fmt.Fprintf(stderr, "%v\n%s: coverage_record failed for %s (see stderr above)\n", err, cicName, rel)
			return 2
		}
	}
	for _, d := range cicExpectedZero {
		if !slices.Contains(corpus, d[0]) {
			continue
		}
		if err := cov.declare(d[0], d[1]); err != nil {
			fmt.Fprintf(stderr, "%v\n%s: coverage_declare failed for '%s' (see stderr above)\n", err, cicName, d[0])
			return 2
		}
	}
	for _, line := range cov.verdict() {
		member, _, _ := strings.Cut(line, ":")
		_, text, _ := strings.Cut(line, ": ")
		violations = append(violations, member+":0: "+text)
	}

	frag := cov.report()
	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintln(stdout, v)
		}
		fmt.Fprint(stdout, "Fix the citation, or mark the line citations-guard:allow if the backticked content is a non-path marker or illustration.\n")
		fmt.Fprintf(stdout, "INSTALLED-CITATIONS-INVALID: %s — %d violation(s)\n", repo, len(violations))
		if frag != "" {
			fmt.Fprintf(stdout, "  %s\n", frag)
		}
		return 1
	}
	fmt.Fprintf(stdout, "INSTALLED-CITATIONS-OK: %s — %d file(s) scanned\n", repo, len(corpus))
	if frag != "" {
		fmt.Fprintf(stdout, "  %s\n", frag)
	}
	return 0
}

// HOW "INSTALLED" IS DERIVED — never re-implemented, never a written list.
// cicInstalledSources creates its own throwaway sandbox, runs the real
// `setup.sh global`, `setup.sh claude-code <sandbox>/proj` and `setup.sh
// zcode <sandbox>/proj` into it (refusing first unless both the HOME and the
// project directory it is about to hand the installer lie inside that
// sandbox — cicRunSetup/cicWithinSandbox, adopting the refusal in
// stats/internal/setuptest/helpers_test.go's runSetup rather than restating
// it), and reads back what appeared (cicDeriveSources). Both project modes
// run because between them they copy CLAUDE.md and AGENTS.md into the
// project. A re-implementation of
// the installer's globs drifts the moment an install path changes;
// deriving from a real run cannot. A non-zero code is the guard's exit.
func cicInstalledSources(env Env, repo string, stderr io.Writer) (map[string]bool, int) {
	sandbox, err := os.MkdirTemp(env.Getenv("TMPDIR"), cicName+".")
	if err != nil {
		// A scratch directory that cannot be made is an environment
		// failure, never a verdict: exit 2 with the cause, not an empty
		// report that reads as clean.
		tmp := env.Getenv("TMPDIR")
		if tmp == "" {
			tmp = "/tmp"
		}
		fmt.Fprintf(stderr, "%s: mktemp -d failed under %s — cannot create a scratch directory\n", cicName, tmp)
		return nil, 2
	}
	// An interrupt removes the sandbox, then dies of that SIGINT (exit 130
	// to a shell), as the Python's try/finally did on KeyboardInterrupt: a
	// deferred RemoveAll alone never runs in a signal-killed process. The
	// context also kills a setup.sh still running, so nothing writes into
	// the sandbox after it is removed.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer func() {
		interrupted := ctx.Err() != nil
		stop()
		os.RemoveAll(sandbox)
		if interrupted {
			_ = syscall.Kill(os.Getpid(), syscall.SIGINT)
			time.Sleep(time.Second) // reached only when SIGINT was ignored at start
			os.Exit(130)
		}
	}()
	proj := filepath.Join(sandbox, "proj")
	if err := os.MkdirAll(proj, 0o777); err != nil {
		fmt.Fprintf(stderr, "%s: cannot create %s — %v\n", cicName, proj, err)
		return nil, 2
	}
	for _, run := range []struct{ mode, proj string }{{"global", ""}, {"claude-code", proj}, {"zcode", proj}} {
		res, err := cicRunSetup(ctx, repo, sandbox, sandbox, run.proj, run.mode)
		if ctx.Err() != nil {
			return nil, 2 // interrupted: the deferred cleanup ends the process
		}
		var refusal *cicSandboxRefusal
		switch {
		case errors.As(err, &refusal):
			fmt.Fprintf(stderr, "%s: refusing to run setup.sh — %v\n", cicName, err)
			return nil, 2
		case err != nil:
			fmt.Fprintf(stderr, "%s: cannot run the sandboxed `setup.sh %s` — %v\n", cicName, run.mode, err)
			return nil, 2
		case res.code != 0:
			fmt.Fprintf(stderr, "%s: sandboxed `setup.sh %s` exited %d — cannot derive the installed set\n", cicName, run.mode, res.code)
			if res.stdout != "" {
				fmt.Fprintln(stderr, res.stdout)
			}
			if res.stderr != "" {
				fmt.Fprintln(stderr, res.stderr)
			}
			return nil, 2
		}
	}
	return cicDeriveSources(repo, sandbox), 0
}

// cicSandboxRefusal is an installer invocation that would touch something
// outside the sandbox this guard created for itself.
type cicSandboxRefusal struct{ path, sandbox string }

func (e *cicSandboxRefusal) Error() string {
	return e.path + " is not inside the sandbox " + e.sandbox
}

// cicRealpath is Python's os.path.realpath: symlinks resolved where the
// path exists, the path made absolute and clean where it does not.
func cicRealpath(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	a, _ := filepath.Abs(p)
	return a
}

// cicWithinSandbox refuses unless the resolved, physical path lies inside
// sandbox. A standalone check with no side effect of its own, so a test can
// prove the refusal fires BEFORE any external command runs, never merely
// that the end result looked safe.
func cicWithinSandbox(path, sandbox string) error {
	s, p := cicRealpath(sandbox), cicRealpath(path)
	if p != s && !strings.HasPrefix(p, s+string(os.PathSeparator)) {
		return &cicSandboxRefusal{path, sandbox}
	}
	return nil
}

type cicSetupResult struct {
	code           int
	stdout, stderr string
}

// cicRunSetup runs `<repo>/setup.sh <mode> [proj]` with HOME=home and its
// cwd the sandbox, refusing first unless home and proj (when given) lie
// inside sandbox. Every real caller builds both AS subdirectories of a
// sandbox it just created, so the refusal never fires in normal operation;
// it exists to make that invariant a checked fact rather than an
// assumption. The installer inherits the process environment, HOME aside.
func cicRunSetup(ctx context.Context, repo, sandbox, home, proj, mode string) (cicSetupResult, error) {
	if err := cicWithinSandbox(home, sandbox); err != nil {
		return cicSetupResult{}, err
	}
	args := []string{mode}
	if proj != "" {
		if err := cicWithinSandbox(proj, sandbox); err != nil {
			return cicSetupResult{}, err
		}
		args = append(args, proj)
	}
	cmd := exec.CommandContext(ctx, filepath.Join(repo, "setup.sh"), args...)
	cmd.Dir = sandbox
	cmd.Env = append(os.Environ(), "HOME="+home)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	res := cicSetupResult{stdout: out.String(), stderr: errOut.String()}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		// A signal is Python's negative returncode.
		if ws, ok := exit.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
			res.code = -int(ws.Signal())
		} else {
			res.code = exit.ExitCode()
		}
		return res, nil
	}
	return res, err
}

// cicListdir is os.listdir: the directory's names in the order the
// filesystem returns them, which decides which of two identical top-level
// files a copied file is attributed to.
func cicListdir(dir string) []string {
	f, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer f.Close()
	names, _ := f.Readdirnames(-1)
	return names
}

// cicDeriveSources walks the sandbox (already populated by the two
// setup.sh runs) and returns the repository-relative source paths it
// reveals: for every symlink, its resolved physical target relative to the
// repository (whether a file or a whole directory — install_skills
// symlinks an entire skill directory as one unit, recorded and never
// descended into); for every regular file, the top-level repository file
// its bytes match, if any (a `cp` copy, e.g. CLAUDE.md/AGENTS.md). A
// regular file matching nothing is not part of the installed set — it is
// something the installer synthesised (a rendered managed block), not a
// citation-bearing file this repository tracks.
func cicDeriveSources(repo, sandbox string) map[string]bool {
	var top []string // top-level regular files, in listdir order
	for _, name := range cicListdir(repo) {
		if fi, err := os.Lstat(filepath.Join(repo, name)); err == nil && fi.Mode().IsRegular() {
			top = append(top, name)
		}
	}
	sources := map[string]bool{}
	_ = filepath.WalkDir(sandbox, func(p string, d fs.DirEntry, err error) error {
		switch {
		case err != nil || p == sandbox || d.IsDir():
			return nil
		case d.Type()&fs.ModeSymlink != 0:
			// ponytail: a dangling link is skipped where Python's
			// realpath would still name its target; the installer makes
			// none.
			if target, err := filepath.EvalSymlinks(p); err == nil &&
				(target == repo || strings.HasPrefix(target, repo+"/")) {
				rel, _ := filepath.Rel(repo, target)
				sources[rel] = true
			}
			return nil
		case !d.Type().IsRegular():
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		for _, name := range top {
			if tl, err := os.ReadFile(filepath.Join(repo, name)); err == nil && bytes.Equal(data, tl) {
				sources[name] = true
				break
			}
		}
		return nil
	})
	return sources
}

// cicCorpus is every `.md`/`.mdc` file this guard scans, sorted: for a
// source that is a directory (a whole installed skill), every such file
// anywhere beneath it in the REAL repository tree; for a source that is
// itself a `.md` file (a rule, a command, a copied CLAUDE.md/AGENTS.md),
// itself. A source that is a file of some other kind (a hook's `.py`)
// contributes its root but nothing to the corpus.
func cicCorpus(repo string, sources map[string]bool) []string {
	isMD := func(name string) bool { return strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".mdc") }
	set := map[string]bool{}
	for s := range sources {
		full := filepath.Join(repo, s)
		switch {
		case isDir(full):
			_ = filepath.WalkDir(full, func(p string, d fs.DirEntry, err error) error {
				// os.walk: a symlink to a directory is listed among the
				// directories and not followed; every other non-directory
				// entry is a file.
				if err != nil || d.IsDir() || !isMD(d.Name()) ||
					d.Type()&fs.ModeSymlink != 0 && isDir(p) {
					return nil
				}
				if rel, err := filepath.Rel(repo, p); err == nil {
					set[rel] = true
				}
				return nil
			})
		case isFile(full) && isMD(s):
			set[s] = true
		}
	}
	corpus := make([]string, 0, len(set))
	for rel := range set {
		corpus = append(corpus, rel)
	}
	sort.Strings(corpus)
	return corpus
}

// cicRead is the file's text as Python's open(encoding="utf-8").read()
// gives it — strict UTF-8, universal newlines — or the refusal message.
func cicRead(repo, rel string) (string, string) {
	b, err := os.ReadFile(filepath.Join(repo, rel))
	if err != nil {
		var pe *fs.PathError
		var errno syscall.Errno
		if errors.As(err, &pe) && errors.As(err, &errno) {
			s := errno.Error()
			return "", fmt.Sprintf("cannot read %s — [Errno %d] %s%s: '%s'", rel, int(errno), strings.ToUpper(s[:1]), s[1:], pe.Path)
		}
		return "", fmt.Sprintf("cannot read %s — %v", rel, err)
	}
	if reason := cicDecodeError(b); reason != "" {
		return "", fmt.Sprintf("cannot decode %s as UTF-8 — 'utf-8' codec %s", rel, reason)
	}
	return strings.ReplaceAll(strings.ReplaceAll(string(b), "\r\n", "\n"), "\r", "\n"), ""
}

// cicDecodeError is Python's UnicodeDecodeError text for b's first
// malformed sequence, or "" when b is valid UTF-8.
func cicDecodeError(b []byte) string {
	for i := 0; i < len(b); {
		r, size := utf8.DecodeRune(b[i:])
		if r != utf8.RuneError || size != 1 {
			i += size
			continue
		}
		c, n := b[i], 0
		switch {
		case c >= 0xC2 && c <= 0xDF:
			n = 2
		case c >= 0xE0 && c <= 0xEF:
			n = 3
		case c >= 0xF0 && c <= 0xF4:
			n = 4
		}
		if n == 0 {
			return fmt.Sprintf("can't decode byte 0x%02x in position %d: invalid start byte", c, i)
		}
		// The second byte's valid range excludes overlongs, surrogates and
		// code points past U+10FFFF, as Python's decoder does.
		lo, hi := byte(0x80), byte(0xBF)
		switch c {
		case 0xE0:
			lo = 0xA0
		case 0xED:
			hi = 0x9F
		case 0xF0:
			lo = 0x90
		case 0xF4:
			hi = 0x8F
		}
		j := i + 1
		for ; j < i+n && j < len(b); j++ {
			if j == i+1 && (b[j] < lo || b[j] > hi) || j > i+1 && (b[j] < 0x80 || b[j] > 0xBF) {
				if j == i+1 {
					return fmt.Sprintf("can't decode byte 0x%02x in position %d: invalid continuation byte", c, i)
				}
				return fmt.Sprintf("can't decode bytes in position %d-%d: invalid continuation byte", i, j-1)
			}
		}
		if j-1 == i {
			return fmt.Sprintf("can't decode byte 0x%02x in position %d: unexpected end of data", c, i)
		}
		return fmt.Sprintf("can't decode bytes in position %d-%d: unexpected end of data", i, j-1)
	}
	return ""
}

// pyIsSpace is Python's str.isspace for one code point: Go's unicode.IsSpace
// plus the four information separators Python also counts.
func pyIsSpace(r rune) bool { return unicode.IsSpace(r) || r >= 0x1C && r <= 0x1F }

func pyStrip(s string) string { return strings.TrimFunc(s, pyIsSpace) }

// pySplit is Python's str.split() with no arguments.
func pySplit(s string) []string { return strings.FieldsFunc(s, pyIsSpace) }

// cicCandidate is one candidate citation: its 1-based line and its token.
type cicCandidate struct {
	line  int
	token string
}

// cicCandidates is scan_file_for_citations. THE CLASSIFIER — block
// structure before tokens, following the ordering check-plan-provenance's
// own docstring records as the reason it is not a regular expression,
// though far narrower in scope than that guard's fence/container parser:
// this one only needs to know whether a line sits inside a fenced
// bash/sh/zsh block, tracked the same way guardsymlinks.go's gsScanMD
// does (a fence's opening backtick run sets
// its length and language; closing requires a run at least as long, with
// nothing trailing). A line inside such a fence, if it is not a `#`
// comment, is a shell argument and is excluded WHOLESALE, matching that
// guard's own choice to scan content inside a non-bash/sh/zsh fence not at
// all, neither as shell nor as prose. Outside any fence, a citation is a
// backtick-delimited span; inside a bash/sh/zsh comment line, it is a bare
// whitespace-delimited word — code has no reason to backtick-quote its own
// comments' paths.
func cicCandidates(text string) []cicCandidate {
	var out []cicCandidate
	fenceLen, fenceLang := 0, ""
	for i, raw := range strings.Split(text, "\n") {
		stripped := pyStrip(raw)
		ticks := len(stripped) - len(strings.TrimLeft(stripped, "`"))
		if fenceLen == 0 {
			if ticks >= 3 {
				fenceLen, fenceLang = ticks, pyStrip(stripped[ticks:])
				continue
			}
			// The declared non-path marker: matched on the raw line so it
			// works bare or inside an HTML comment, and checked only where
			// candidates are taken — never around a fence delimiter, whose
			// state toggle must survive any marker text on the line.
			if strings.Contains(raw, cicAllowMarker) {
				continue
			}
			for _, tok := range cicBacktickTokens(raw) {
				out = append(out, cicCandidate{i + 1, tok})
			}
			continue
		}
		// Inside a fence: does this line close it? A run of backticks at
		// least as long as the opener, with nothing trailing.
		if ticks >= fenceLen && pyStrip(stripped[ticks:]) == "" {
			fenceLen, fenceLang = 0, ""
			continue
		}
		// A non-comment line inside a bash/sh/zsh fence is a shell argument
		// and is excluded wholesale; a fence tagged with any other language
		// (or none) is left alone entirely.
		if cicShellLang(fenceLang) && strings.HasPrefix(stripped, "#") && !strings.Contains(raw, cicAllowMarker) {
			for _, tok := range cicMergePlaceholders(pySplit(raw)) {
				out = append(out, cicCandidate{i + 1, tok})
			}
		}
	}
	return out
}

// cicShellLang is SHELL_FENCE_LANG_RE, `^(bash|sh|zsh)(\s|$)`.
func cicShellLang(lang string) bool {
	for _, l := range []string{"bash", "sh", "zsh"} {
		if rest, ok := strings.CutPrefix(lang, l); ok {
			r, _ := utf8.DecodeRuneInString(rest)
			if rest == "" || pyIsSpace(r) {
				return true
			}
		}
	}
	return false
}

// cicMergePlaceholders is merge_bracket_placeholder_words: a `<…>`-bracketed
// phrase spanning more than one word is merged back into one word, wherever
// it starts.
//
// A plain whitespace split — used both to walk a backtick span's own words
// and to tokenize a shell-fence comment line — breaks any placeholder whose
// own text contains a space (`<agents repo>`, `<project root>`) into several
// fragments before either ever gets a chance to classify it. Left unmerged,
// the FIRST fragment (`<agents`) is a candidate that never reaches judgment
// at all, and every OTHER fragment is judged on its own, which turns a
// single multi-word placeholder citation into a garbled, unreadable
// violation (`repo>/…`) instead of the real one. Generalised to ANY run of
// words starting with a `<` that has no `>` of its own, consumed until a
// later word supplies one — so the reported token is always the real
// placeholder text, whether or not that placeholder turns out to be a
// recognised root (cicNamesRoot's own, separately closed, business).
//
// A `<` that never finds a closing `>` before the words run out is left
// exactly as split — nothing here manufactures a placeholder that is not
// actually there. So is a `<` that DOES find one but with a `/` in some
// word strictly between them (panel round 2): a placeholder phrase never
// has one there, but a shell redirection does — `some-cmd <
// scripts/input.txt > output.log` would otherwise merge into one garbled
// false citation. Nested brackets closing at the first `>` (`<a <b> c>`
// merges only through `<b>`) and an unclosed `<` are both reviewed and
// left as-is — neither is a defect this corpus's own content exercises.
func cicMergePlaceholders(words []string) []string {
	var merged []string
	for i := 0; i < len(words); i++ {
		w := words[i]
		if strings.HasPrefix(w, "<") && !strings.Contains(w, ">") {
			parts := []string{w}
			for j := i + 1; j < len(words); j++ {
				parts = append(parts, words[j])
				if !strings.Contains(words[j], ">") {
					continue
				}
				if !slices.ContainsFunc(parts[1:len(parts)-1], func(p string) bool { return strings.Contains(p, "/") }) {
					merged = append(merged, strings.Join(parts, " "))
					i = j
					w = ""
				}
				break
			}
			if w == "" {
				continue
			}
		}
		merged = append(merged, w)
	}
	return merged
}

// cicBacktickTokens is extract_backtick_tokens: every whitespace-delimited
// word of every span between a pair of single backticks, in order — EVERY
// word, not only the first. Greedy left-to-right pairing — this corpus
// writes citations in plain single-backtick spans, never nested or doubled.
//
// An earlier cut kept only a span's LEADING word, on the theory that a
// multi-word span is an inline shell example — `agents/setup.sh global` —
// and everything past the first word an argument. That dropped a genuine
// second citation — `see .flow/project.md` — which then went NEVER SEEN.
// Classifying every word closes that hole without reopening the original
// one: `global` carries no `/` and names no real file at the repository
// root, so cicIsCitation's bare-token rule excludes it independently, and a
// quoted fragment of shell output is excluded by the leading/trailing-quote
// rule.
//
// A SPAN whose own content begins with `<!--` (leading whitespace inside
// the backticks ignored, panel round 5) is skipped WHOLESALE — being a
// comment is a property of the whole span, not of any word inside it
// (panel round 4: an HTML comment illustrating a path inline, `<!--
// measured: ./gradlew test @ c515c42 -->`, carries a `/`-bearing word
// strictly between `<!--` and `-->`, so the redirection bound correctly
// refuses to merge it, and the unmerged `./gradlew` would be judged on its
// own). Deliberately NOT widened to "`<!--` appears anywhere in the span":
// this corpus discusses comment syntax in PROSE where a real citation can
// sit elsewhere in the same span; a comment that is not span-initial is
// accepted, documented residue instead — see the citation-roots spec.
func cicBacktickTokens(line string) []string {
	var tokens []string
	for i := 0; i < len(line); i++ {
		if line[i] != '`' {
			continue
		}
		j := strings.IndexByte(line[i+1:], '`')
		if j < 0 {
			break
		}
		span := line[i+1 : i+1+j]
		i += j + 1
		if strings.HasPrefix(strings.TrimLeftFunc(span, pyIsSpace), "<!--") {
			continue
		}
		tokens = append(tokens, cicMergePlaceholders(pySplit(span))...)
	}
	return tokens
}

var (
	cicURL        = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*://`)
	cicGitBranch  = regexp.MustCompile(`^[A-Za-z0-9_.-]+/(main|master|HEAD)$`)
	cicExtension  = regexp.MustCompile(`\.[A-Za-z0-9]+$`)
	cicFileLine   = regexp.MustCompile(`:[0-9]+$`)
	cicSpectreRef = regexp.MustCompile(`^spectre/<[^/<>]+>$`)
	cicChoreRef   = regexp.MustCompile(`^chore/(?:archive|self-review)-<[^/<>]+>$`)
)

// cicIsCitation is classify_token: true when token is a citation at all —
// before its root is ever judged. Per candidate token, excluded before a
// root is ever judged: an absolute path, a `~`-rooted path, a `../`-rooted
// path (relative to something the document never names, so no root could
// correctly prefix it), a URL, a token carrying a leading or trailing quote
// (a quoted fragment of program output — a real citation is never wrapped
// in one inside its own backtick span), a pipeline branch-name shape, a
// `file:line` location, a token with a NON-FINAL segment ending in `:`, a
// shell variable reference, a git ref shape, and a regular-expression or
// glob fragment. A token with no "/" is a citation only when it names a
// real file at the repository root — this is what catches `README.md` and
// `setup.sh` without flagging every generic mention of `SKILL.md` or
// `tasks.md`.
func cicIsCitation(token string, rootFiles map[string]bool) bool {
	tok := pyStrip(token)
	switch {
	case tok == "", strings.HasPrefix(tok, "/"), strings.HasPrefix(tok, "~"),
		strings.HasPrefix(tok, "../"), cicURL.MatchString(tok),
		strings.ContainsRune(`'"`, rune(tok[0])), strings.ContainsRune(`'"`, rune(tok[len(tok)-1])):
		return false
	// GIT_BRANCH_SPECTRE_RE — `spectre/<…>` names a branch, not a path: the
	// branch prefix followed by exactly one `<…>` placeholder segment and
	// NOTHING ELSE. The prefix became `spectre/` at the spectre cutover,
	// moved WITH the contract prose that writes the branch name: a shape
	// exclusion left pointing at a prefix the corpus no longer writes stops
	// matching silently, and every branch name it used to exempt is then
	// reported as an unrooted path. Anchored at both ends so it cannot widen
	// past that: `spectre/specs/x.md` (no brackets) and
	// `spectre/changes/<name>/` (a bracket segment followed by MORE path)
	// stay fully reportable — the deliberately narrow bound that stops the
	// shape becoming the fail-open hole task 9's placeholder generalisation
	// was reverted for. `[^/<>]+` inside the brackets, not `.+`, so the
	// bracket segment itself cannot smuggle in a second `/`.
	case cicSpectreRef.MatchString(tok):
		return false
	// GIT_BRANCH_CHORE_RE — kan-239's sibling for the pipeline's other
	// branch shapes: `chore/archive-<name>`, which `/flow`'s archive run
	// creates and names, and `chore/self-review-<name>`, which run 2 no
	// longer creates but which survives as real branches from before that
	// change. Same shape, same bounds, for the same reason: a branch name is
	// not a filesystem path; `chore/archive-<name>/spec.md` stays reportable.
	case cicChoreRef.MatchString(tok):
		return false
	// FILE_LINE_RE — a token ending `:<digits>` names a line inside a file
	// (a findings-table Location cell, taken verbatim from `git diff` output
	// and diff-relative by construction), not a path to cite; see
	// skills/flow/review-panel.md's example findings row. Matched against
	// the whole token, so a path that merely CONTAINS a colon elsewhere is
	// untouched. Narrowing it (strip `:<digits>`, classify the rest) was
	// tried and rejected, not overlooked: it forces the review-panel prose's
	// own deliberately fabricated example row — `src/Foo.kt:42` — to be
	// judged as an unrooted citation, and re-rooting a fabricated example
	// would misrepresent it. So this IS a documented hole (a real
	// `path:line` citation in this shape is unreportable) — see the
	// citation-roots spec.
	case cicFileLine.MatchString(tok):
		return false
	}
	// A NON-FINAL `/`-delimited segment ending in `:` is not a path segment
	// — this excludes a fragment of check-plan-provenance's own error
	// message (`measured:/predicted:`), quoted verbatim rather than
	// reworded. Narrowed to non-final (panel round 3) after the first cut
	// silently dropped a real citation ending its own span in a bare colon
	// (`.flow/project.md:`, from prose reading "see `.flow/project.md:` for
	// the list") — a final segment's trailing colon is just punctuation.
	segs := strings.Split(tok, "/")
	if slices.ContainsFunc(segs[:len(segs)-1], func(s string) bool { return strings.HasSuffix(s, ":") }) {
		return false
	}
	first := segs[0]
	if strings.HasPrefix(first, "$") {
		return false // shell variable reference
	}
	if strings.HasPrefix(tok, "refs/") || cicGitBranch.MatchString(tok) {
		return false // git ref shape
	}
	// ORIGIN — a token whose first segment is `origin` is a remote-tracking
	// ref only when what follows has no file extension (the per-task-review
	// bound, finding 1: the original rule excluded EVERY `origin/…` token,
	// so `origin/README.md` sailed through uncounted). A real path citation
	// in this corpus always ends in one (`.md`, `.sh`, `.py`, …), and
	// neither live ref site (`origin/$BASE`, `origin/spectre/<name>`) does.
	// A trailing "/" names a DIRECTORY, never a git ref, so `origin/rules/`
	// does not take this exclusion (panel round 3). `origin/README` — no
	// extension, no trailing slash — stays excluded: a branch genuinely
	// could be named README, and that residue is accepted rather than
	// closed.
	if first == "origin" {
		remainder := strings.TrimPrefix(tok, "origin/")
		if !strings.HasSuffix(remainder, "/") && !cicExtension.MatchString(remainder) {
			return false
		}
	}
	// "$" is deliberately excluded from this set: a leading "$" is the
	// shell-variable exclusion's own signal, and keeping it here too would
	// make that exclusion untestable in isolation — the untested-mutation
	// shape KAN-197 names.
	if strings.ContainsAny(tok, "[](){}*+?|^") {
		return false // regular-expression or glob fragment
	}
	if !strings.Contains(tok, "/") {
		return rootFiles[tok]
	}
	return true
}

// cicPlaceholderRoots — the closed set the citation-roots spec enumerates.
// Task 9's first cut accepted ANY first segment shaped `<…>` — bracket-
// shaped alone — which fails open: `<foo>/spectre/specs/x.md` passed while
// naming an unrooted path, and so did a typo of a real placeholder
// (`<changeroot>/`, `<change-root>/`). A guard that reports clean while
// checking nothing is worse than no guard: a site leaves the report by
// acquiring a root, never by ceasing to look like a citation. Adding a
// placeholder root here is a deliberate act that extends the spec's own
// table; it is not something this guard does by pattern-matching brackets.
//
// `<skill-dir>` resolves at runtime to the INSTALLED skills root
// (`~/.claude/skills/<skill>/`, `~/.zcode/skills/<skill>/`, …) — neither
// this checkout nor the target project, the same criterion that earned
// `<abs-worktree>`, `<changeRoot>` and `<state-dir>` their places. It
// collapses three corpus wordings of the same concept ("the running
// command's own skill directory" and two variants); pipeline.md's "Guard
// resolution" section is canonical for what it means.
var cicPlaceholderRoots = map[string]bool{
	"<agents repo>": true, "<project>": true, "<abs-worktree>": true,
	"<changeRoot>": true, "<state-dir>": true, "<skill-dir>": true,
}

// cicNamesRoot is judges_ok: true when token — already classified as a
// citation — names a recognised root: an installed root's own first
// segment, or a MEMBER of cicPlaceholderRoots — never merely
// bracket-shaped.
//
// The installed roots are the first path segment of every derived source
// that HAS a second segment. A source with none — a copied top-level file
// such as CLAUDE.md, whose repository-relative path is its own bare
// basename — never becomes a root, so a bare citation of CLAUDE.md itself
// still needs a prefix: it is exactly this single-segment shape a bare,
// no-slash citation like `README.md` would otherwise collide with.
func cicNamesRoot(token string, roots map[string]bool) bool {
	first, _, _ := strings.Cut(token, "/")
	return cicPlaceholderRoots[first] || roots[first]
}
