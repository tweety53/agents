package guard

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// checkVerbatimMoves is scripts/check-verbatim-moves.sh: that script's header
// comment is the contract. It proves a prompt trim moved or de-duplicated
// text and never reworded it, comparing the sentences of the run-loaded
// corpus at a base ref against the working tree; exit 0 clean (REVIEW lines
// included), 1 a violation, 2 cannot answer. It is the Go port of the
// prototype verbatim_check.py in docs/prompt-audit-2026-09-29/tools.md
// (KAN-852), and its splitting and verdict rules are that prototype's.
func init() {
	Registry["check-verbatim-moves"] = checkVerbatimMoves
}

// vmImperative is the prototype's IMPERATIVE: a sentence carrying one of
// these words states a rule, and a rule moved into a -rationale.md is out of
// every run's reach, so the move is reported for a human to confirm.
var vmImperative = regexp.MustCompile(`(?i)\b(never|always|must|only|do not|don't|exactly|every|each|required|refuse[sd]?|stop|before|after|unless|except)\b`)

// vmAllowedNew is the prototype's ALLOWED_NEW: the only new run-loaded
// sentences a trim may add are a load directive and a citation — what a
// moved block leaves behind in its place.
var vmAllowedNew = regexp.MustCompile("(\\*\\*Load `|[Ll]oad `[^`]+` only when|\\(`skills/[^`]+\\.md`\\))")

var (
	vmQuote      = regexp.MustCompile(`^(>\s?)+`)
	vmBreak      = regexp.MustCompile(`^(#+ |\||---|[-*+] |\d+[.)] )`)
	vmListMarker = regexp.MustCompile(`^([-*+]|\d+[.)])\s+`)
	vmSpace      = regexp.MustCompile(`\s+`)
)

// vmAckFile is the per-change acknowledgement list: one sentence per line,
// exactly as the guard prints it after "::", `#` lines ignored. A leading
// `\` is dropped and the rest taken literally, so a heading the guard prints
// (`## …`) is listed as `\## …` rather than read as a comment. A sentence
// in it is a rewording or deletion the change states it meant — the escape
// a non-trim change needs, recorded in the change's own artifacts so it
// archives with them and never outlives the change.
const vmAckFile = "verbatim-moves.txt"

func checkVerbatimMoves(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) > 1 {
		fmt.Fprintln(stderr, "usage: check-verbatim-moves.sh [<base-ref>]")
		return 2
	}
	root := env.Getenv("FLOW_GUARD_REPO_ROOT")
	if root == "" {
		fmt.Fprintln(stderr, "check-verbatim-moves: FLOW_GUARD_REPO_ROOT is unset — run scripts/check-verbatim-moves.sh, which sets it")
		return 2
	}
	gitEnv := []string{"LC_ALL=C"}
	if v, ok := env.LookupEnv("GIT_CONFIG_GLOBAL"); ok {
		gitEnv = append(gitEnv, "GIT_CONFIG_GLOBAL="+v)
	}
	git := envGit(env, gitEnv...)
	if git("-C", root, "rev-parse", "--git-dir").Run() != nil {
		fmt.Fprintf(stderr, "check-verbatim-moves: %s is not a git worktree\n", root)
		return 2
	}

	base, label := "", ""
	if len(args) == 1 {
		label = args[0]
		// `--end-of-options` keeps a ref that starts with `-` from being read
		// as an option.
		sha, ok := capture(git("-C", root, "rev-parse", "--verify", "--quiet", "--end-of-options", args[0]+"^{commit}"))
		if !ok || sha == "" {
			fmt.Fprintf(stderr, "check-verbatim-moves: base ref %q does not resolve to a commit\n", args[0])
			return 2
		}
		base = sha
	} else {
		var ok bool
		if base, label, ok = vmDefaultBase(git, root); !ok {
			fmt.Fprintln(stderr, "check-verbatim-moves: cannot resolve the merge base with the default branch — pass a base ref")
			return 2
		}
	}

	baseRun, baseRat, err := vmCorpusAt(git, root, base)
	if err != nil {
		fmt.Fprintf(stderr, "check-verbatim-moves: cannot read %s: %v\n", label, err)
		return 2
	}
	headRun, headRat, err := vmCorpusTree(root)
	if err != nil {
		fmt.Fprintf(stderr, "check-verbatim-moves: cannot read the working tree: %v\n", err)
		return 2
	}
	if len(baseRun)+len(headRun) == 0 {
		fmt.Fprintf(stderr, "check-verbatim-moves: no sentences under skills/, commands-claude/ or rules/ in %s — refusing to report a clean run\n", root)
		return 2
	}
	acks, err := vmAcks(root)
	if err != nil {
		fmt.Fprintf(stderr, "check-verbatim-moves: %v\n", err)
		return 2
	}

	bad, review := 0, 0
	for _, s := range vmKeys(baseRun) {
		n := baseRun[s] - headRun[s]
		switch {
		case n <= 0:
		case headRun[s] > 0:
			fmt.Fprintf(stdout, "ok   dedup   still stated in a run-loaded file :: %s\n", vmShort(s))
		case headRat[s]-baseRat[s] >= n:
			if vmImperative.MatchString(s) {
				review++
				fmt.Fprintf(stdout, "REVIEW moved to a -rationale.md but carries an imperative marker :: %s\n", s)
			} else {
				fmt.Fprintf(stdout, "ok   moved   to a -rationale.md :: %s\n", vmShort(s))
			}
		case acks[s]:
			fmt.Fprintf(stdout, "ok   acknowledged removal :: %s\n", vmShort(s))
		default:
			bad++
			fmt.Fprintf(stdout, "FAIL deleted or reworded :: %s\n", s)
		}
	}
	for _, s := range vmKeys(headRun) {
		if baseRun[s] != 0 || baseRat[s] != 0 || vmAllowedNew.MatchString(s) {
			continue
		}
		if acks[s] {
			fmt.Fprintf(stdout, "ok   acknowledged new text :: %s\n", vmShort(s))
			continue
		}
		bad++
		fmt.Fprintf(stdout, "FAIL new run-loaded text (paraphrase?) :: %s\n", s)
	}
	fmt.Fprintf(stdout, "check-verbatim-moves: %d violation(s), %d to review, against %s\n", bad, review, label)
	if bad != 0 {
		fmt.Fprintf(stderr, "Restore the sentence, move it verbatim, or — when the change means it — list it in <spec-root>/changes/<change>/%s.\n", vmAckFile)
		return 1
	}
	return 0
}

// vmDefaultBase is the merge base of HEAD with the change's base, resolved
// offline — the base recorded on the current branch
// (branch.<cur>.flowBase, resolve-base-branch.sh's first answer), else
// origin/HEAD, else origin/main, else main — so a lint run never fetches. On
// the default branch itself the base is HEAD, and the guard judges the
// uncommitted edits alone.
func vmDefaultBase(git func(...string) *exec.Cmd, root string) (sha, label string, ok bool) {
	var candidates []string
	if cur, ok := capture(git("-C", root, "branch", "--show-current")); ok && strings.TrimSpace(cur) != "" {
		if rec, ok := capture(git("-C", root, "config", "--get", "branch."+strings.TrimSpace(cur)+".flowBase")); ok && strings.TrimSpace(rec) != "" {
			candidates = append(candidates, "origin/"+strings.TrimSpace(rec))
		}
	}
	if sym, ok := capture(git("-C", root, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD")); ok && sym != "" {
		candidates = append(candidates, sym)
	}
	candidates = append(candidates, "origin/main", "main")
	for _, c := range candidates {
		if mb, ok := capture(git("-C", root, "merge-base", "HEAD", c)); ok && mb != "" {
			return mb, "merge-base(HEAD, " + c + ") " + mb[:min(12, len(mb))], true
		}
	}
	return "", "", false
}

// vmScoped is the prototype's files(): skills/**/*.md, commands-claude/*.md,
// rules/*.md and rules/*.mdc, relative to the root.
func vmScoped(rel string) bool {
	dir, name := filepath.Split(rel)
	dir = strings.TrimSuffix(dir, "/")
	switch {
	case strings.HasPrefix(rel, "skills/"):
		return strings.HasSuffix(name, ".md")
	case dir == "commands-claude":
		return strings.HasSuffix(name, ".md")
	case dir == "rules":
		return strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".mdc")
	}
	return false
}

// vmCorpus is one side's sentence counts: run-loaded files and -rationale.md
// files kept apart.
type vmCorpus = map[string]int

func vmAdd(run, rat vmCorpus, rel, text string) {
	into := run
	if strings.HasSuffix(rel, "-rationale.md") {
		into = rat
	}
	for _, s := range vmSentences(text) {
		into[s]++
	}
}

// vmCorpusTree reads the working tree, symlinks skipped as the prototype's
// `not p.is_symlink()` skips them.
func vmCorpusTree(root string) (vmCorpus, vmCorpus, error) {
	run, rat := vmCorpus{}, vmCorpus{}
	for _, top := range []string{"skills", "commands-claude", "rules"} {
		dir := filepath.Join(root, top)
		if !isDir(dir) {
			continue
		}
		err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.Type().IsRegular() {
				return nil
			}
			rel, err := filepath.Rel(root, p)
			if err != nil || !vmScoped(filepath.ToSlash(rel)) {
				return err
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			vmAdd(run, rat, filepath.ToSlash(rel), string(b))
			return nil
		})
		if err != nil {
			return nil, nil, err
		}
	}
	return run, rat, nil
}

// vmCorpusAt reads the same scope at a commit: regular blobs only (mode
// 120000 is a symlink), read through one `git cat-file --batch`.
func vmCorpusAt(git func(...string) *exec.Cmd, root, commit string) (vmCorpus, vmCorpus, error) {
	ls := git("-C", root, "ls-tree", "-r", "-z", "--full-tree", commit, "--", "skills", "commands-claude", "rules")
	out, err := ls.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("git ls-tree: %v", err)
	}
	var paths, shas []string
	for _, ent := range strings.Split(string(out), "\x00") {
		meta, rel, ok := strings.Cut(ent, "\t")
		if !ok {
			continue
		}
		f := strings.Fields(meta)
		if len(f) != 3 || f[1] != "blob" || f[0] == "120000" || !vmScoped(rel) {
			continue
		}
		paths, shas = append(paths, rel), append(shas, f[2])
	}
	run, rat := vmCorpus{}, vmCorpus{}
	if len(shas) == 0 {
		return run, rat, nil
	}
	cat := git("-C", root, "cat-file", "--batch")
	cat.Stdin = strings.NewReader(strings.Join(shas, "\n") + "\n")
	raw, err := cat.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("git cat-file: %v", err)
	}
	r := bufio.NewReader(bytes.NewReader(raw))
	for i := range shas {
		hdr, err := r.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("git cat-file: short output at %s", paths[i])
		}
		f := strings.Fields(hdr)
		if len(f) != 3 {
			return nil, nil, fmt.Errorf("git cat-file: %q for %s", strings.TrimSpace(hdr), paths[i])
		}
		size, err := strconv.Atoi(f[2])
		if err != nil {
			return nil, nil, fmt.Errorf("git cat-file: bad size for %s", paths[i])
		}
		body := make([]byte, size+1) // the object, then its trailing LF
		if _, err := io.ReadFull(r, body); err != nil {
			return nil, nil, fmt.Errorf("git cat-file: short object %s", paths[i])
		}
		vmAdd(run, rat, paths[i], string(body[:size]))
	}
	return run, rat, nil
}

// vmAcks is every sentence listed in an in-flight change's acknowledgement
// file. Archived changes are skipped: their acknowledgements were for a diff
// that has already landed.
func vmAcks(root string) (map[string]bool, error) {
	acks := map[string]bool{}
	leaf := specRootLeaf(root, io.Discard)
	matches, err := filepath.Glob(filepath.Join(root, leaf, "changes", "*", vmAckFile))
	if err != nil {
		return nil, err
	}
	for _, m := range matches {
		b, err := os.ReadFile(m)
		if err != nil {
			return nil, fmt.Errorf("cannot read %s: %v", m, err)
		}
		for _, l := range strings.Split(string(b), "\n") {
			l = vmSpace.ReplaceAllString(strings.TrimSpace(l), " ")
			if l == "" || strings.HasPrefix(l, "#") {
				continue
			}
			acks[strings.TrimPrefix(l, `\`)] = true
		}
	}
	return acks, nil
}

// vmBlocks is the prototype's blocks(): paragraphs joined across soft wraps;
// a heading, table row, rule or list item stands alone; a fenced line stands
// alone; blockquote markers are dropped.
func vmBlocks(text string) []string {
	var out, cur []string
	flush := func() {
		if len(cur) > 0 {
			out = append(out, strings.Join(cur, " "))
			cur = nil
		}
	}
	fence := false
	for _, l := range strings.Split(text, "\n") {
		s := strings.TrimSpace(l)
		if strings.HasPrefix(s, "```") || strings.HasPrefix(s, "~~~") {
			flush()
			fence = !fence
			continue
		}
		if fence {
			out = append(out, s)
			continue
		}
		s = vmQuote.ReplaceAllString(s, "")
		if s == "" || vmBreak.MatchString(s) {
			flush()
			if s != "" {
				out = append(out, vmListMarker.ReplaceAllString(s, ""))
			}
			continue
		}
		cur = append(cur, s)
	}
	flush()
	return out
}

// vmSentences is the prototype's sentences(): each block split after
// `.`, `!` or `?` — the closing `)]"'` + "`*_" run after it dropped with the
// whitespace, as the prototype's lookbehind split drops it — whitespace
// collapsed, anything under 12 characters discarded.
func vmSentences(text string) []string {
	var out []string
	for _, b := range vmBlocks(text) {
		for _, s := range vmSplit(b) {
			s = strings.TrimSpace(vmSpace.ReplaceAllString(s, " "))
			if utf8.RuneCountInString(s) >= 12 {
				out = append(out, s)
			}
		}
	}
	return out
}

// vmNumberedHeading reports whether the "." at b[i] ends a bare number that
// opens a bold span -- "**4. Execute (SDD + TDD)**", a citation of a numbered
// heading -- where no sentence ends.
func vmNumberedHeading(b string, i int) bool {
	if b[i] != '.' {
		return false
	}
	d := i
	for d > 0 && '0' <= b[d-1] && b[d-1] <= '9' {
		d--
	}
	return d < i && strings.HasSuffix(b[:d], "**")
}

func vmSplit(b string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(b); i++ {
		if c := b[i]; c != '.' && c != '!' && c != '?' {
			continue
		}
		if vmNumberedHeading(b, i) {
			continue
		}
		j := i + 1
		for j < len(b) && strings.IndexByte(")]\"'`*_", b[j]) >= 0 {
			j++
		}
		k := j
		for k < len(b) {
			r, size := utf8.DecodeRuneInString(b[k:])
			if !vmIsSpace(r) {
				break
			}
			k += size
		}
		if k == j {
			continue
		}
		parts = append(parts, b[start:i+1])
		start, i = k, k-1
	}
	return append(parts, b[start:])
}

// vmIsSpace is Python's str `\s`: Unicode whitespace.
func vmIsSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\v', '\f', '\r', 0x1c, 0x1d, 0x1e, 0x1f, 0x85, 0xa0, 0x1680,
		0x2028, 0x2029, 0x202f, 0x205f, 0x3000:
		return true
	}
	return r >= 0x2000 && r <= 0x200a
}

func vmKeys(c vmCorpus) []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// vmShort is the prototype's s[:100] for the lines that pass: a verdict
// that needs no action needs no full sentence. FAIL and REVIEW lines carry
// the whole sentence, so it can be copied into the acknowledgement file.
func vmShort(s string) string {
	if utf8.RuneCountInString(s) <= 100 {
		return s
	}
	return string([]rune(s)[:100])
}
