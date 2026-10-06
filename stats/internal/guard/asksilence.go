package guard

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// checkAskSilence is scripts/check-ask-silence.sh's guard, and this comment
// is the contract's one home — the shim header defers here. It fails when a
// run-loaded ask site states none of its own silence outcome, a delegation
// of the ask, or a citation of the pipeline contract's
// **Unanswered mid-run asks** section (skills/flow-contracts/pipeline.md).
// KAN-880, deferred from KAN-772's self-review: the corpus-wide survey of
// AskUserQuestion sites was hand-run once, and nothing re-ran it, so a
// future ask site added without its own silence outcome drifted silently.
//
// THE UNIT is the section: from a Markdown heading line to the next heading
// of any level, plus the preamble before a file's first heading. A section
// that mentions AskUserQuestion outside a code fence is an ask site, and its
// whole section must pass.
//
// A section passes when its text outside code fences, case-insensitively,
// carries any of:
//
//   - the citation: "unanswered mid-run asks";
//   - the contract's own silence vocabulary — "silence", "silent",
//     "unanswered", "default", "recommended", "explicit", "without asking"
//     (substring match, so "explicitly" and "defaults" count);
//   - a delegation preposition before a bolded cross-reference — "per **",
//     "applies **", "under **", "exactly as **" — handing the ask's
//     governance to a section this guard cannot follow.
//
// LIMITS, deliberately. The detection is lexical, never semantic: a site
// whose outcome is stated in words outside the vocabulary fails, and the
// remedy is to state the outcome, delegate the ask, or cite the section —
// never to widen the
// vocabulary to hide one. A delegation's target is not checked, so a section
// can pass by pointing at prose that itself states nothing; that is the
// preposition list's price, paid for not parsing prose. Text inside a fenced
// code block is neither a heading, a mention nor an outcome — a worked
// example in a fence is not a live ask.
//
// THE CORPUS mirrors scripts/lib/owned-corpus.sh — the one definition of the
// run-loaded Markdown, which this Go port cannot source: the scope roots
// skills/, rules/, spectre/specs/, commands-claude/ and .flow/, plus the
// *.md and *.mdc files directly at the root; the exclusions node_modules,
// .superpowers, spectre/changes/archive and docs/superpowers; a symlinked
// scope root refused rather than skipped, and a nested symlinked directory
// refused when following it would reach any .md or .mdc or when it cannot be
// looked through at all, exactly as that library refuses them. Drift between
// the two definitions is a real defect: change them together or move the
// walk behind one implementation.
//
// Exit 0 when every ask section passes (and when the corpus holds none), 1
// on violation(s), 2 when the corpus cannot be answered at all —
// FLOW_GUARD_REPO_ROOT unset, a missing or unreadable scope root, a symlink
// hiding Markdown or one find -L cannot look through, or a corpus file that
// exists but cannot be read.
func init() {
	Registry["check-ask-silence"] = checkAskSilence
}

const askSilenceName = "check-ask-silence"

// askSilenceScopes is owned-corpus.sh's OWNED_CORPUS_SCOPE_DIRS.
var askSilenceScopes = []string{"skills", "rules", "spectre/specs", "commands-claude", ".flow"}

// asVocabulary is the contract's own words for a stated silence outcome; asDelegation
// is the prepositions a section delegates its ask's governance with.
var (
	asVocabulary  = []string{"silence", "silent", "unanswered", "default", "recommended", "explicit", "without asking"}
	asDelegation  = []string{"per **", "applies **", "under **", "exactly as **"}
	asCitation    = "unanswered mid-run asks"
	asTool        = "AskUserQuestion"
	asHeadingMark = "#"
)

type asSection struct {
	heading string // the heading line, trimmed; "(preamble)" before the first heading
	start   int    // 1-based line the section starts at, its heading included
	body    string // the section's text outside code fences, its heading line included
}

func checkAskSilence(_ []string, env Env, stdout, stderr io.Writer) int {
	root := env.Getenv("FLOW_GUARD_REPO_ROOT")
	if root == "" {
		fmt.Fprintf(stderr, askSilenceName+": FLOW_GUARD_REPO_ROOT is unset — run scripts/check-ask-silence.sh, which sets it\n")
		return 2
	}
	files, code := asCorpus(root, stderr)
	if files == nil {
		return code
	}

	violations := 0
	sections := 0
	for _, path := range files {
		b, err := os.ReadFile(path)
		if err != nil {
			// An unreadable file is the corpus not answerable in full —
			// the same refusal owned_corpus_files makes for a missing
			// scope root, never a partial answer counted as violations.
			fmt.Fprintf(stderr, askSilenceName+": %s exists but cannot be read — %v\n", path, err)
			return 2
		}
		for _, sec := range asSections(string(b)) {
			if !strings.Contains(sec.body, asTool) {
				continue
			}
			sections++
			if asPasses(sec.body) {
				continue
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				rel = path
			}
			fmt.Fprintf(stderr, "%s:%d: ASK-SILENCE: section %q mentions %s and states no silence outcome, no delegation of the ask, and no citation of Unanswered mid-run asks (skills/flow-contracts/pipeline.md)\n",
				rel, sec.start, sec.heading, asTool)
			violations++
		}
	}

	if violations != 0 {
		fmt.Fprintf(stdout, "ASK-SILENCE-FAIL: %s — %d ask section(s) of %d state no silence outcome; state the outcome, delegate the ask, or cite Unanswered mid-run asks (skills/flow-contracts/pipeline.md)\n",
			root, violations, sections)
		return 1
	}
	if sections == 0 {
		fmt.Fprintf(stdout, "ASK-SILENCE-OK: %s — no AskUserQuestion site in the corpus\n", root)
		return 0
	}
	fmt.Fprintf(stdout, "ASK-SILENCE-OK: %s — %d ask section(s), each citing, delegating, or stating its silence outcome\n", root, sections)
	return 0
}

// asPasses reports whether one section's fence-free body states a silence
// outcome, cites the contract, or delegates the ask — the header's three
// passes, in that order.
func asPasses(body string) bool {
	low := strings.ToLower(body)
	if strings.Contains(low, asCitation) {
		return true
	}
	for _, w := range asVocabulary {
		if strings.Contains(low, w) {
			return true
		}
	}
	for _, p := range asDelegation {
		if strings.Contains(low, p) {
			return true
		}
	}
	return false
}

// asSections splits one file into sections at Markdown headings, outside
// code fences. The heading line opens the next section and is part of the
// body it opens — a section naming the tool only in its heading is still an
// ask site; the preamble before a file's first heading is its own section.
func asSections(text string) []asSection {
	var secs []asSection
	cur := asSection{heading: "(preamble)", start: 1}
	flush := func() {
		if cur.body != "" || cur.heading != "(preamble)" {
			secs = append(secs, cur)
		}
	}
	inFence := false
	line := 1
	for _, raw := range strings.SplitAfter(text, "\n") {
		trimmed := strings.TrimSpace(raw)
		switch {
		case strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~"):
			inFence = !inFence
		case !inFence && strings.HasPrefix(trimmed, asHeadingMark) && isHeadingLine(trimmed):
			flush()
			// The heading opens the section and is part of its body: a
			// section naming the tool only in its heading is still an ask
			// site, and the violation line anchors at the heading itself.
			cur = asSection{heading: trimmed, start: line, body: raw}
		default:
			if !inFence {
				cur.body += raw
			}
		}
		line++
	}
	flush()
	return secs
}

// isHeadingLine distinguishes a `# …` heading from a non-heading line that
// merely starts with the mark: one to six marks, then a space or the end of
// the line.
func isHeadingLine(trimmed string) bool {
	n := 0
	for n < len(trimmed) && trimmed[n] == '#' {
		n++
	}
	return n >= 1 && n <= 6 && (n == len(trimmed) || trimmed[n] == ' ')
}

// asCorpus enumerates the owned corpus under root, mirroring
// owned_corpus_files: every .md/.mdc directly at the root and under the
// scope roots, exclusions applied, sorted for stable output. It returns nil
// with an exit code when it cannot answer at all.
func asCorpus(root string, stderr io.Writer) ([]string, int) {
	refuse := func(format string, a ...any) ([]string, int) {
		fmt.Fprintf(stderr, askSilenceName+": "+format+"\n", a...)
		return nil, 2
	}
	readable := func(path string) bool {
		info, err := os.Stat(path)
		return err == nil && info.IsDir()
	}
	for _, dir := range askSilenceScopes {
		if lp, err := os.Lstat(filepath.Join(root, dir)); err == nil && lp.Mode()&fs.ModeSymlink != 0 {
			return refuse("scope root is a symlink: %s", filepath.Join(root, dir))
		}
		if !readable(filepath.Join(root, dir)) {
			return refuse("scope root missing or unreadable: %s", filepath.Join(root, dir))
		}
	}

	// Non-nil from the start: nil is the cannot-answer signal below, and an
	// empty corpus is a success the caller must still print a verdict for.
	files := []string{}
	add := func(p string) { files = append(files, p) }

	// Root-level files, maxdepth 1.
	roots, err := os.ReadDir(root)
	if err != nil {
		return refuse("not a readable directory: %s", root)
	}
	for _, ent := range roots {
		if ent.Type().IsRegular() && asOwnedName(ent.Name()) {
			add(filepath.Join(root, ent.Name()))
		}
	}

	// Scope trees, physical mode: symlinks are never followed, so a nested
	// symlinked directory hides whatever sits behind it. As owned-corpus.sh
	// does, that is refused — loudly, naming the path — when the target
	// would reach any .md or .mdc, and skipped when it would not.
	for _, dir := range askSilenceScopes {
		base := filepath.Join(root, dir)
		err := filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			switch {
			case d.Type()&fs.ModeSymlink != 0:
				// Exclusions win for link paths too, exactly as the bash
				// sweep's owned_corpus_excluded skip does: a link named
				// node_modules hides nothing this corpus is missing.
				if asExcluded(asRel(root, p)) {
					return nil
				}
				// A symlink is never descended into; one pointing at a
				// directory that would reach Markdown is refused, one
				// pointing at a file — or a dangling one — is not this
				// rule's subject and is skipped. A target that cannot be
				// read through is cannot-answer, exactly as the bash sweep's
				// failing `find -L` capture is.
				hides, err := asHidesMarkdown(p)
				if err != nil {
					return fmt.Errorf("cannot look through the symlinked directory: %s — %v", p, err)
				}
				if hides {
					return fmt.Errorf("symlinked directory hides Markdown from the corpus: %s", p)
				}
				return nil
			case d.IsDir():
				if asExcluded(asRel(root, p)) {
					return filepath.SkipDir
				}
				return nil
			case d.Type().IsRegular() && asOwnedName(d.Name()):
				if !asExcluded(asRel(root, p)) {
					add(p)
				}
			}
			return nil
		})
		if err != nil {
			return refuse("%v", err)
		}
	}

	sort.Strings(files)
	return files, 0
}

// asHidesMarkdown reports whether following the symlink at p reaches a
// directory holding any .md or .mdc file — the coarse predicate
// owned-corpus.sh walks its link list with, deliberately: a refusal is loud
// and resolved by a human, a silent skip quietly shrinks the corpus. A link
// to a file, and a dangling one, hide nothing. The walk follows symlinks
// the way find -L does — a link→dir→link→.md chain is a hiding link — and
// skips a loop silently, which is find -L's measured behaviour. A directory
// that cannot be read is an error, not a hidden-nothing: find -L fails on
// it and the bash capture refuses, so the twin does too.
func asHidesMarkdown(p string) (bool, error) {
	target, err := filepath.EvalSymlinks(p)
	if err != nil {
		return false, nil
	}
	info, err := os.Stat(target)
	if err != nil || !info.IsDir() {
		return false, nil
	}
	visited := map[string]bool{target: true}
	var walk func(dir string) (bool, error)
	walk = func(dir string) (bool, error) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return false, err
		}
		for _, ent := range entries {
			full := filepath.Join(dir, ent.Name())
			resolved := full
			if ent.Type()&fs.ModeSymlink != 0 {
				resolved, err = filepath.EvalSymlinks(full)
				if err != nil {
					continue // a dangling link hides nothing
				}
			}
			fi, err := os.Stat(resolved) // Stat follows the whole chain
			if err != nil {
				continue
			}
			switch {
			case fi.IsDir():
				if visited[resolved] {
					continue // a loop: find -L skips it silently
				}
				visited[resolved] = true
				hides, err := walk(resolved)
				if err != nil || hides {
					return hides, err
				}
			case fi.Mode().IsRegular() && asOwnedName(ent.Name()):
				return true, nil
			}
		}
		return false, nil
	}
	return walk(target)
}

// asRel is filepath.Rel with the separator normalised, for exclusion tests.
func asRel(root, p string) string {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return p
	}
	return filepath.ToSlash(rel)
}

// asOwnedName is the corpus's extension test: .md and .mdc.
func asOwnedName(name string) bool {
	return strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".mdc")
}

// asExcluded is owned_corpus_excluded: structural path-component and
// path-prefix tests, never a filename list.
func asExcluded(rel string) bool {
	for _, part := range strings.Split(rel, "/") {
		if part == "node_modules" || part == ".superpowers" {
			return true
		}
	}
	return rel == "spectre/changes/archive" || strings.HasPrefix(rel, "spectre/changes/archive/") ||
		rel == "docs/superpowers" || strings.HasPrefix(rel, "docs/superpowers/")
}
