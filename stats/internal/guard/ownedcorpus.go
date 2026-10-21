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

// THE CORPUS is the Markdown this repository owns, and this file is its one
// definition — check-ask-silence and check-normative-inventory both walk it.
// It was scripts/lib/owned-corpus.sh until KAN-920 ported its last bash
// consumer and deleted it; TestAskSilenceCorpusParity still pins this walk
// against that library as it stood at libTwinBase.
//
// Every .md and .mdc file under the scope roots skills/, rules/,
// spectre/specs/, commands-claude/ and .flow/, plus the ones sitting directly
// at the repository root. The root is deliberately NOT walked recursively:
// that would sweep in spectre/changes/, whose session records a /flow* run
// writes DURING a change, so an inventory captured before a change's first
// edit could never match one captured after its last.
//
// THE EXCLUSIONS are structural — a path component, or a path prefix, never a
// list of filenames: any component named node_modules or .superpowers, and
// the prefixes spectre/changes/archive/ and docs/superpowers/. The last three
// sit outside every scope root today; they are stated anyway so a later
// widening of the scope roots cannot silently acquire them.
//
// SYMLINKS ARE NOT OWNED CONTENT: a .md symlink contributes nothing — its
// target is inventoried in its own right if this repository owns it. A scope
// root that is a symlink is refused, not resolved and not skipped: a
// physical walk would contribute zero files from it with no error. A
// symlinked directory nested inside a scope root is refused when following
// it would reach any .md or .mdc (or when it cannot be looked through at
// all), and skipped otherwise — this repository's skills/flow*/scripts/lib
// links hold .sh files only, and refusing them would fail every real run.
// The predicate is deliberately the coarse one: a link into the same scope
// root loses nothing yet is refused, and that direction is safe — loud, and
// resolved by a human deciding the repository's shape.

// ownedCorpusScopes is the scope roots, relative to the repository root.
var ownedCorpusScopes = []string{"skills", "rules", "spectre/specs", "commands-claude", ".flow"}

// ownedCorpus enumerates the owned corpus under root: every .md/.mdc
// directly at the root and under the scope roots, exclusions applied, sorted
// for stable output. It returns nil with exit code 2 when it cannot answer
// at all, after one refusal line on stderr led by prefix.
func ownedCorpus(root, prefix string, stderr io.Writer) ([]string, int) {
	refuse := func(format string, a ...any) ([]string, int) {
		fmt.Fprintf(stderr, prefix+": "+format+"\n", a...)
		return nil, 2
	}
	if !gsReadableDir(root) {
		return refuse("not a readable directory: %s", root)
	}
	for _, dir := range ownedCorpusScopes {
		if lp, err := os.Lstat(filepath.Join(root, dir)); err == nil && lp.Mode()&fs.ModeSymlink != 0 {
			return refuse("scope root is a symlink: %s", filepath.Join(root, dir))
		}
		if !gsReadableDir(filepath.Join(root, dir)) {
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
		if ent.Type().IsRegular() && ocOwnedName(ent.Name()) {
			add(filepath.Join(root, ent.Name()))
		}
	}

	// Scope trees, physical mode: symlinks are never followed, so a nested
	// symlinked directory hides whatever sits behind it. As owned-corpus.sh
	// did, that is refused — loudly, naming the path — when the target
	// would reach any .md or .mdc, and skipped when it would not.
	for _, dir := range ownedCorpusScopes {
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
				if ocExcluded(ocRel(root, p)) {
					return nil
				}
				// A symlink is never descended into; one pointing at a
				// directory that would reach Markdown is refused, one
				// pointing at a file — or a dangling one — is not this
				// rule's subject and is skipped. A target that cannot be
				// read through is cannot-answer, exactly as the bash sweep's
				// failing `find -L` capture is.
				hides, err := ocHidesMarkdown(p)
				if err != nil {
					return fmt.Errorf("cannot look through the symlinked directory: %s — %v", p, err)
				}
				if hides {
					return fmt.Errorf("symlinked directory hides Markdown from the corpus: %s", p)
				}
				return nil
			case d.IsDir():
				if ocExcluded(ocRel(root, p)) {
					return filepath.SkipDir
				}
				return nil
			case d.Type().IsRegular() && ocOwnedName(d.Name()):
				if !ocExcluded(ocRel(root, p)) {
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

// ocHidesMarkdown reports whether following the symlink at p reaches a
// directory holding any .md or .mdc file — the coarse predicate
// owned-corpus.sh walked its link list with, deliberately: a refusal is loud
// and resolved by a human, a silent skip quietly shrinks the corpus. A link
// to a file, and a dangling one, hide nothing. The walk follows symlinks
// the way find -L does — a link→dir→link→.md chain is a hiding link — and
// skips a loop silently, which is find -L's measured behaviour. A directory
// that cannot be read is an error, not a hidden-nothing: find -L fails on
// it and the bash capture refuses, so the twin does too.
func ocHidesMarkdown(p string) (bool, error) {
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
			case fi.Mode().IsRegular() && ocOwnedName(ent.Name()):
				return true, nil
			}
		}
		return false, nil
	}
	return walk(target)
}

// asRel is filepath.Rel with the separator normalised, for exclusion tests.
func ocRel(root, p string) string {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return p
	}
	return filepath.ToSlash(rel)
}

// ocOwnedName is the corpus's extension test: .md and .mdc.
func ocOwnedName(name string) bool {
	return strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".mdc")
}

// ocExcluded is owned_corpus_excluded's rule: structural path-component and
// path-prefix tests, never a filename list.
func ocExcluded(rel string) bool {
	for _, part := range strings.Split(rel, "/") {
		if part == "node_modules" || part == ".superpowers" {
			return true
		}
	}
	return rel == "spectre/changes/archive" || strings.HasPrefix(rel, "spectre/changes/archive/") ||
		rel == "docs/superpowers" || strings.HasPrefix(rel, "docs/superpowers/")
}
