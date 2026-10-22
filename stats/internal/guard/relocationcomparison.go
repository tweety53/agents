package guard

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

// generate-relocation-comparison: for a plan that declares itself a
// relocation, generate a before/after passage comparison across every file
// the plan's own tasks scope, so a review panel can audit a move the same way
// the contract-economy per-move ledger does (canonical spec: the
// review-panel-economics spec's "A relocation-declaring plan gets a
// mechanical passage comparison" requirement -- do not restate its rules
// here, only implement them).
//
// Usage: generate-relocation-comparison.sh <worktree> <changeRoot>
// <merge-base>. The plan read is <changeRoot>/tasks.md, spelled by plain
// concatenation; <changeRoot> is otherwise unused -- the comparison's output
// path is always <worktree>/.superpowers/sdd/relocation-comparison.md,
// independent of where the plan lives. A relative path resolves against the
// caller's working directory, as every git call here runs `git -C
// <worktree>` from it.
//
// Every diagnostic opens `generate-relocation-comparison:`, the shim's own
// name; the Python printed its deleted file's name, `…comparison.py:`.
//
// Exit codes:
//
//	0  written (a table naming every moved / repointed / added / removed
//	   passage across the plan's scoped files), OR the no-op case -- a
//	   missing, `no`, or malformed `**Relocation:**` header, or an explicit
//	   `yes` with an empty file scope. Both are success: this never blocks
//	   the panel it feeds.
//	2  cannot answer: wrong argument count, tasks.md cannot be read (a plan
//	   that is not UTF-8 included: the Python let that decode error escape
//	   as a traceback and exit 1; its header named it exit 2, and the header
//	   wins), <merge-base> does not resolve to a real commit, a scoped file
//	   is missing from the worktree on disk or is a symlink there, a scoped
//	   path is a tree (directory) or a symlink rather than a blob at
//	   <merge-base>, a `git cat-file -t` or `git show` failure once `git
//	   cat-file -e` has said the path exists there, a scoped file (worktree
//	   copy or merge-base blob) that is not valid UTF-8, or the comparison
//	   cannot be written.
//
// Algorithm
//
//  1. The first tasks.md line matching grcRelocationRE decides: `yes` goes
//     on; anything else -- no match at all, or an explicit `no` -- is the
//     no-op.
//  2. On `yes`, every task in the plan (pdbParseTasks, checked and unchecked
//     alike -- the header is plan-level) contributes its **Files:** paths as
//     check-task-commit-fields reads them (tcfPythonFiles: inline or
//     bulleted, KAN-860 F5 -- pdbParseTasks reads bullets alone), plan order
//     preserved, duplicates dropped. An empty union is the no-op.
//  3. For each scoped path, passages are extracted from the blob at
//     <merge-base>:<path> and from the worktree file. Absent at the
//     merge-base (`cat-file -e` fails) is empty text, the ordinary shape of
//     a file the plan creates fresh -- decided by `-e` alone, never by
//     parsing `git show` stderr (F14). Present but unreadable (`cat-file -t`
//     fails: a corrupt object `-e` still reports as existing, case l) is
//     cannot-answer, never folded into "absent". Present as anything but a
//     blob (a tree, F18) or as a symlink entry (mode 120000 from `git
//     ls-tree` -- `cat-file -t` says `blob` for a symlink too, F21) is
//     cannot-answer: `git show` would hand back a tree listing or a link
//     target, never content. A worktree symlink is likewise refused rather
//     than followed, so neither side resolves a link the other would not.
//     A passage is a run of non-blank lines between two blank lines or
//     `##`/`###` headings; OR a single `- `/`* ` bullet line; OR a single
//     Markdown table row. Each carries the nearest preceding heading text.
//  4. Passages sharing one text: a (file, heading) present on both sides is
//     unchanged, never a row; leftover before and after occurrences pair up
//     as moved. Remaining before-only and after-only texts pair by
//     grcRepointSignature as repointed; the rest are removed or added.
//  5. The table: Passage (first eight words, `…` past them) / Source /
//     Destination / Class, a literal `|` in a cell escaped as `\|`.
func init() { Registry["generate-relocation-comparison"] = generateRelocationComparison }

const grcLabel = "generate-relocation-comparison: "

var (
	grcRelocationRE = regexp.MustCompile(`^[` + tcfWS + `]*>[` + tcfWS + `]*\*\*Relocation:\*\*[` + tcfWS + `]+(yes|no)`)
	grcHeadingRE    = regexp.MustCompile(`^#{2,3}[` + tcfWS + `]+(.*)$`)
	grcBulletRE     = regexp.MustCompile(`^[` + tcfWS + `]*[-*][` + tcfWS + `]+.*[^` + tcfWS + `].*$`)
	grcTableRowRE   = regexp.MustCompile(`^[` + tcfWS + `]*\|.*\|[` + tcfWS + `]*$`)
	grcBacktickRE   = regexp.MustCompile("`[^`]*`")
	grcBoldRE       = regexp.MustCompile(`\*\*[^*]*\*\*`)
	// grcSeparatorRE is CITATION_SEPARATOR_RE, matched at a backtick span's
	// exact end: the separators the corpus's citations use between a quoted
	// path and the bold section token beside it -- whitespace, a possessive
	// `'s`, or a comma -- and nothing wider (F19.3).
	grcSeparatorRE = regexp.MustCompile(`^(?:'s|,)?[ \t]{0,3}`)
	// grcPositionWordRE is TRAILING_POSITION_WORD_RE less its leading \b,
	// which grcTrimPositionWord checks with Python's Unicode word classes: a
	// trailing above/below word, then whitespace and a punctuation run
	// (F20, F22).
	grcPositionWordRE = regexp.MustCompile(`(above|below)[` + tcfWS + `]*[.,;:!?]*[` + tcfWS + `]*$`)
	// grcTrailingPunctRE is stripped from both sides unconditionally (F16).
	grcTrailingPunctRE = regexp.MustCompile(`[.,;:!?]*[` + tcfWS + `]*$`)
)

func generateRelocationComparison(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 3 {
		fmt.Fprintln(stderr, "usage: generate-relocation-comparison.sh <worktree> <changeRoot> <merge-base>")
		return 2
	}
	worktree, base, tasksPath := args[0], args[2], args[1]+"/tasks.md"
	lines, err := tcfReadPlan(env, tasksPath)
	if err != nil {
		fmt.Fprintf(stderr, "%scannot read %s: %v\n", grcLabel, tasksPath, err)
		return 2
	}
	if !grcRelocates(lines) {
		return 0
	}
	paths := grcScopedFiles(lines)
	if len(paths) == 0 {
		return 0
	}
	if _, _, code, err := tcfExec(env.Dir, nil, "git", "-C", worktree, "cat-file", "-e", base+"^{commit}"); err != nil || code != 0 {
		fmt.Fprintf(stderr, "%smerge-base %s does not resolve to a commit in %s\n", grcLabel, tcfRepr(base), worktree)
		return 2
	}
	var before, after []grcPassage
	for _, p := range paths {
		wtPath := grcJoin(worktree, p)
		if fi, err := os.Lstat(psAbs(env, wtPath)); err == nil && fi.Mode()&os.ModeSymlink != 0 {
			fmt.Fprintf(stderr, "%sscoped file is a symlink in the worktree, cannot compare: %s\n", grcLabel, wtPath)
			return 2
		}
		afterText, err := grcReadWorktree(env, wtPath)
		if err != nil {
			fmt.Fprintf(stderr, "%s%v\n", grcLabel, err)
			return 2
		}
		beforeText, err := grcGitShow(env, worktree, base, p)
		if err != nil {
			fmt.Fprintf(stderr, "%s%v\n", grcLabel, err)
			return 2
		}
		before = append(before, grcPassages(beforeText, p)...)
		after = append(after, grcPassages(afterText, p)...)
	}
	var out strings.Builder
	out.WriteString("| Passage | Source | Destination | Class |\n| --- | --- | --- | --- |\n")
	for _, row := range grcClassify(before, after) {
		fmt.Fprintf(&out, "| %s | %s | %s | %s |\n", grcFirstEightWords(row.text), grcFormatLocation(row.source), grcFormatLocation(row.destination), row.class)
	}
	dir := psAbs(env, grcJoin(worktree, ".superpowers/sdd"))
	if err := os.MkdirAll(dir, 0o777); err != nil {
		fmt.Fprintf(stderr, "%scannot write the comparison: %v\n", grcLabel, err)
		return 2
	}
	if err := os.WriteFile(dir+"/relocation-comparison.md", []byte(out.String()), 0o666); err != nil {
		fmt.Fprintf(stderr, "%scannot write the comparison: %v\n", grcLabel, err)
		return 2
	}
	return 0
}

// grcRelocates is find_relocation: the first header line decides, and its
// value must end at a word boundary as Python's Unicode \b reads one.
func grcRelocates(lines []string) bool {
	for _, line := range lines {
		m := grcRelocationRE.FindStringSubmatchIndex(line)
		if m == nil {
			continue
		}
		if r, _ := utf8.DecodeRuneInString(line[m[1]:]); m[1] < len(line) && ppIsWord(r) {
			continue
		}
		return line[m[2]:m[3]] == "yes"
	}
	return false
}

// grcScopedFiles is scoped_files.
func grcScopedFiles(lines []string) []string {
	var paths []string
	for _, t := range pdbParseTasks(lines) {
		task, _ := tcfParseTask(lines, t.id)
		for _, p := range tcfPythonFiles(lines, task) {
			if !slices.Contains(paths, p) {
				paths = append(paths, p)
			}
		}
	}
	return paths
}

// grcJoin is str(Path(base) / p): an absolute p replaces base, and empty and
// `.` components drop out; `..` stays.
func grcJoin(base, p string) string {
	if !strings.HasPrefix(p, "/") {
		p = base + "/" + p
	}
	var parts []string
	for _, c := range strings.Split(p, "/") {
		if c != "" && c != "." {
			parts = append(parts, c)
		}
	}
	joined := strings.Join(parts, "/")
	if strings.HasPrefix(p, "/") {
		return "/" + joined
	}
	if joined == "" {
		return "."
	}
	return joined
}

// grcReadWorktree reads a scoped worktree file, its error already the
// Python's message.
func grcReadWorktree(env Env, path string) (string, error) {
	b, err := os.ReadFile(psAbs(env, path))
	if err != nil {
		if pe := (*os.PathError)(nil); errors.As(err, &pe) {
			pe.Path = path
		}
		return "", fmt.Errorf("scoped file missing from worktree: %s: %v", path, tcfOSError(path, err))
	}
	s, err := tcfText(b)
	if err != nil {
		return "", fmt.Errorf("%s is not valid UTF-8: %s", path, strings.TrimPrefix(err.Error(), "UnicodeDecodeError: "))
	}
	return s, nil
}

// grcGitShow is _git_show: the blob at <ref>:<path>, "" when nothing exists
// there, or the cannot-answer error.
func grcGitShow(env Env, worktree, ref, path string) (string, error) {
	spec := ref + ":" + path
	if _, _, code, err := tcfExec(env.Dir, nil, "git", "-C", worktree, "cat-file", "-e", spec); err != nil || code != 0 {
		return "", nil
	}
	out, errOut, code, err := tcfExec(env.Dir, nil, "git", "-C", worktree, "cat-file", "-t", spec)
	if err != nil {
		return "", err
	}
	if code != 0 {
		return "", fmt.Errorf("git cat-file -t %s failed: %s", spec, tcfStrip(string(errOut)))
	}
	if kind := tcfStrip(string(out)); kind != "blob" {
		return "", fmt.Errorf("%s is a %s, not a file (blob)", spec, kind)
	}
	if out, _, code, err := tcfExec(env.Dir, nil, "git", "-C", worktree, "ls-tree", ref, "--", path); err == nil && code == 0 {
		if f := strings.FieldsFunc(string(out), tcfIsSpace); len(f) > 0 && f[0] == "120000" {
			return "", fmt.Errorf("%s is a symlink, not a regular file", spec)
		}
	}
	out, errOut, code, err = tcfExec(env.Dir, nil, "git", "-C", worktree, "show", spec)
	if err != nil {
		return "", err
	}
	text, err := tcfText(out)
	if err != nil {
		return "", fmt.Errorf("%s is not valid UTF-8: %s", spec, strings.TrimPrefix(err.Error(), "UnicodeDecodeError: "))
	}
	if code != 0 {
		return "", fmt.Errorf("git show %s failed: %s", spec, tcfStrip(string(errOut)))
	}
	return text, nil
}

// grcLocation is a passage's (file, heading); heading is Python's None when
// !hasHeading, and an empty heading is a distinct, falsy value.
type grcLocation struct {
	file, heading string
	hasHeading    bool
}

type grcPassage struct {
	text string
	loc  grcLocation
}

// grcPassages is extract_passages.
func grcPassages(text, file string) []grcPassage {
	var passages []grcPassage
	loc := grcLocation{file: file}
	var buffer []string
	flush := func() {
		if s := tcfStrip(strings.Join(buffer, "\n")); s != "" {
			passages = append(passages, grcPassage{s, loc})
		}
		buffer = nil
	}
	for _, line := range tcfSplitLines(text) {
		if m := grcHeadingRE.FindStringSubmatch(line); m != nil {
			flush()
			loc = grcLocation{file, tcfStrip(m[1]), true}
			continue
		}
		if tcfStrip(line) == "" {
			flush()
			continue
		}
		if grcBulletRE.MatchString(line) || grcTableRowRE.MatchString(line) {
			flush()
			passages = append(passages, grcPassage{tcfStrip(line), loc})
			continue
		}
		buffer = append(buffer, line)
	}
	flush()
	return passages
}

// grcRepointSignature is repoint_signature: every backtick span and every
// bold span found independently (F13 -- one combined regex could splice
// across two unrelated spans); a backtick span whose separator ends exactly
// where a bold span starts collapses with it into one placeholder, any other
// backtick span into its own, an unpaired bold span stays as written; then a
// trailing above/below word and a trailing punctuation run are dropped.
func grcRepointSignature(text string) string {
	bolds := map[int]int{}
	for _, m := range grcBoldRE.FindAllStringIndex(text, -1) {
		bolds[m[0]] = m[1]
	}
	var b strings.Builder
	pos := 0
	for _, m := range grcBacktickRE.FindAllStringIndex(text, -1) {
		start, end, placeholder := m[0], m[1], "\x00PATH\x00"
		if boldEnd, ok := bolds[end+len(grcSeparatorRE.FindString(text[end:]))]; ok {
			end, placeholder = boldEnd, "\x00PATH\x00\x00BOLD\x00"
		}
		if start < pos {
			continue // consumed by an earlier citation's bold span
		}
		b.WriteString(text[pos:start])
		b.WriteString(placeholder)
		pos = end
	}
	b.WriteString(text[pos:])
	s := strings.TrimRightFunc(grcTrimPositionWord(b.String()), tcfIsSpace)
	return strings.TrimRightFunc(s[:grcTrailingPunctRE.FindStringIndex(s)[0]], tcfIsSpace)
}

// grcTrimPositionWord drops the leftmost trailing above/below match that
// starts at a Python word boundary.
func grcTrimPositionWord(s string) string {
	for off := 0; off < len(s); {
		m := grcPositionWordRE.FindStringIndex(s[off:])
		if m == nil {
			break
		}
		start := off + m[0]
		if r, _ := utf8.DecodeLastRuneInString(s[:start]); start == 0 || !ppIsWord(r) {
			return s[:start]
		}
		off = start + 1
	}
	return s
}

type grcRow struct {
	text                string
	source, destination *grcLocation
	class               string
}

// grcClassify is classify.
func grcClassify(before, after []grcPassage) []grcRow {
	group := func(ps []grcPassage) ([]string, map[string][]grcLocation) {
		var order []string
		by := map[string][]grcLocation{}
		for _, p := range ps {
			if _, ok := by[p.text]; !ok {
				order = append(order, p.text)
			}
			by[p.text] = append(by[p.text], p.loc)
		}
		return order, by
	}
	beforeOrder, beforeBy := group(before)
	afterOrder, afterBy := group(after)
	var rows []grcRow
	var beforeOnly, afterOnly []grcPassage
	for _, text := range beforeOrder {
		afterLocs, ok := afterBy[text]
		if !ok {
			for _, loc := range beforeBy[text] {
				beforeOnly = append(beforeOnly, grcPassage{text, loc})
			}
			continue
		}
		remBefore := slices.Clone(beforeBy[text])
		remAfter := slices.Clone(afterLocs)
		for _, loc := range slices.Clone(remBefore) {
			if i := slices.Index(remAfter, loc); i >= 0 {
				remBefore = slices.Delete(remBefore, slices.Index(remBefore, loc), slices.Index(remBefore, loc)+1)
				remAfter = slices.Delete(remAfter, i, i+1)
			}
		}
		n := min(len(remBefore), len(remAfter))
		for i := range n {
			rows = append(rows, grcRow{text, &remBefore[i], &remAfter[i], "moved"})
		}
		for _, loc := range remBefore[n:] {
			beforeOnly = append(beforeOnly, grcPassage{text, loc})
		}
		for _, loc := range remAfter[n:] {
			afterOnly = append(afterOnly, grcPassage{text, loc})
		}
	}
	for _, text := range afterOrder {
		if _, ok := beforeBy[text]; !ok {
			for _, loc := range afterBy[text] {
				afterOnly = append(afterOnly, grcPassage{text, loc})
			}
		}
	}
	matched := map[int]bool{}
	var removed []grcPassage
	for _, p := range beforeOnly {
		sig, found := grcRepointSignature(p.text), false
		for i, a := range afterOnly {
			if !matched[i] && grcRepointSignature(a.text) == sig {
				matched[i], found = true, true
				rows = append(rows, grcRow{p.text, &p.loc, &a.loc, "repointed"})
				break
			}
		}
		if !found {
			removed = append(removed, p)
		}
	}
	for _, p := range removed {
		rows = append(rows, grcRow{p.text, &p.loc, nil, "removed"})
	}
	for i, a := range afterOnly {
		if !matched[i] {
			rows = append(rows, grcRow{a.text, nil, &a.loc, "added"})
		}
	}
	return rows
}

func grcEscape(s string) string { return strings.ReplaceAll(s, "|", `\|`) }

func grcFirstEightWords(text string) string {
	words := strings.FieldsFunc(text, tcfIsSpace)
	if len(words) > 8 {
		return grcEscape(strings.Join(words[:8], " ")) + "…"
	}
	return grcEscape(strings.Join(words, " "))
}

func grcFormatLocation(loc *grcLocation) string {
	if loc == nil {
		return "—"
	}
	if loc.heading != "" {
		return "`" + grcEscape(loc.file) + "` § " + grcEscape(loc.heading)
	}
	return "`" + grcEscape(loc.file) + "`"
}
