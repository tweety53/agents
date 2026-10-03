package guard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"unicode/utf8"
)

func init() { Registry["check-self-review-report"] = checkSelfReviewReport }

// The report-line patterns. srrFindingShape is the LOOSE predicate: a line
// that is attempting to be a finding line, well-formed or not — a leading
// `-`, any amount (including zero) of horizontal whitespace, then `**[`.
// srrFindingLine is the STRICT predicate a shape must additionally satisfy
// to be accepted as a well-formed finding: exactly one space after the `-`,
// a `]**`, one space, then text. Both the orphan check (for a finding-shaped
// line before any heading) and the in-section check beside it (for a
// finding-shaped line that fails the strict pattern) test against this same
// srrFindingShape, so the two can never disagree about what counts as
// finding-shaped — they once used separately hardcoded patterns that did
// disagree, letting a malformed line slip through silently whenever it
// happened to fall inside a recognized section.
//
// srrIssueKey is an uppercase project key, a hyphen, and digits — e.g.
// `KAN-201`. Anything else naming itself `filed:` (empty, `yes`, `201`,
// `KAN-`) is a violation.
var (
	srrFindingShape = regexp.MustCompile(`^-[[:space:]]*\*\*\[`)
	srrFindingLine  = regexp.MustCompile(`^-[[:space:]]\*\*\[([^]]+)\]\*\*[[:space:]](.+)$`)
	srrFiled        = regexp.MustCompile(`^filed:[[:space:]]*(.*)$`)
	srrIssueKey     = regexp.MustCompile(`^[A-Z][A-Z0-9]*-[0-9]+$`)
	srrH2           = regexp.MustCompile(`^##[[:space:]]`)
	srrHeading      = regexp.MustCompile(`^#+[[:space:]]`)
	srrAngleRow     = regexp.MustCompile(`^[[:space:]]*[0-9]+[[:space:]]*$`)
)

const (
	srrNoneMarker = "_none — this angle produced no findings._"
	// srrOriginalAngles is the table's size when every report on the frozen
	// five-angle list was written (the script header's FIVE-ANGLE REPORTS
	// ARE A FROZEN LIST).
	srrOriginalAngles = 5
)

// checkSelfReviewReport is scripts/check-self-review-report.sh's body. Its
// contract — the report shape, the served angle labels, the frozen list,
// coverage and the exit codes — is the script's header; the reasoning
// behind each step is here.
//
// The bash's NO RECORD PROTOCOL note (KAN-211) carries over: each line is
// classified where it is read, with no intermediate encoding, so a raw 0x1F
// byte in a report is ordinary content.
func checkSelfReviewReport(args []string, env Env, stdout, stderr io.Writer) int {
	die := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "check-self-review-report: "+format+"\n", a...)
		return 2
	}
	// The bash derived the repository root as its own `$SCRIPT_DIR/..`; the
	// binary runs from a cache, so its shim exports it.
	root := env.Getenv("FLOW_GUARD_REPO_ROOT")
	if root == "" {
		return die("FLOW_GUARD_REPO_ROOT is unset — run scripts/check-self-review-report.sh, which sets it")
	}
	abs := func(p string) string {
		if filepath.IsAbs(p) {
			return p
		}
		return filepath.Join(env.Dir, p)
	}

	if len(args) > 1 {
		return die("usage: check-self-review-report.sh [dir]")
	}
	target := root + "/docs/self-review"
	if len(args) == 1 {
		target = args[0]
	}
	st, err := os.Stat(abs(target))
	switch {
	case err != nil:
		return die("no such directory: %s", target)
	case !st.IsDir():
		return die("not a directory: %s", target)
	case syscall.Access(abs(target), 4) != nil:
		return die("cannot read directory: %s", target)
	}

	// The angle labels, in the report shape's own order, parsed from the
	// canonical table (kan-585 — see the header). The override variable
	// exists for the tests alone: a fixture contract proves the guard
	// follows the source without writing the real tree. The table is the
	// only one in the contract whose rows are numbered in the first cell, so
	// "second `|` field numeric, last non-empty field is the backticked
	// label" selects exactly the angle rows and cannot leak from any other
	// table.
	contract := env.Getenv("CHECK_SELF_REVIEW_ANGLES_CONTRACT")
	if contract == "" {
		contract = root + "/skills/flow-self-review/SKILL.md"
	}
	if syscall.Access(abs(contract), 4) != nil {
		return die("the canonical angle table is unreadable: %s", contract)
	}
	labels := srrAngleLabels(abs(contract))
	if len(labels) == 0 {
		return die("the canonical angle table yielded no labels: %s", contract)
	}

	// The frozen list, newline-wrapped so a basename is matched whole. A
	// list that exists but cannot be read is "cannot answer", never an empty
	// exemption set; the read failing is reported as the bash's `cat` did.
	listPath := target + "/five-angle-reports.txt"
	fiveAngle := "\n"
	if _, err := os.Stat(abs(listPath)); err == nil {
		b, err := os.ReadFile(abs(listPath))
		if err != nil {
			// gsFindErr renders the line a tool prints for a failed path; cat's is
			// find's with its own name.
			fmt.Fprint(stderr, "cat: "+strings.TrimPrefix(gsFindErr(listPath, err), "find: "))
			return die("cannot read the five-angle report list: %s", listPath)
		}
		fiveAngle = "\n" + strings.TrimRight(string(b), "\n") + "\n"
	}

	files, rc := srrReports(env, target, abs(target), stderr)
	if rc != 0 {
		return rc
	}

	utf8Strict := srrUTF8Locale(env)
	violations, checked := 0, 0
	violation := func(format string, a ...any) {
		fmt.Fprintf(stdout, format+"\n", a...)
		violations++
	}
	var cov coverage
	for _, f := range files {
		rel := strings.TrimPrefix(f, root+"/")
		base := filepath.Base(f)

		// Whether the read happened at all is answered by the read's own
		// error — check the operation, never a precondition standing in for
		// it. The bash's `[[ -r ]]` precheck was removed for that reason
		// (KAN-211 round 2): it tests access()'s permission bits rather than
		// the open() that matters, and opens a window between the test and
		// the read. Without the check an unreadable report would count zero
		// and be reported as an undeclared-zero violation at exit 1 instead
		// of the "cannot answer" the header promises.
		body, err := os.ReadFile(abs(f))
		if err != nil {
			return die("cannot read report file: %s", f)
		}

		// One slot per angle label, walked by index; the slot count is the
		// parsed table's (kan-585), never a literal tuple — a literal went
		// stale the moment the canonical table grew.
		n := len(labels)
		headingFound := make([]bool, n)
		headingLine := make([]int, n)
		hasNone := make([]bool, n)
		findings := make([]int, n)
		// lastIdx is the highest angle index whose heading has been seen so
		// far, so a heading appearing before it in table order is named as
		// out of order (F4). -1 means no heading seen yet.
		lastIdx := -1
		// cur is the label of the section currently open ("" for none) and
		// idx its index into labels; idx is read only where cur is set.
		cur, idx := "", 0

		lines := strings.Split(string(body), "\n")
		if lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		for i, line := range lines {
			lineno := i + 1
			// A trailing `\r` is stripped before any pattern is tested, so a
			// CRLF-saved report reads the same as an LF one (F9).
			line = strings.TrimSuffix(line, "\r")

			// A `##` heading carrying one of the backtick-quoted labels opens
			// that angle's section; one carrying none of them resets the
			// current section to none, so lines under it are attributed to no
			// angle at all.
			if srrH2.MatchString(line) {
				cur = ""
				for j, l := range labels {
					if strings.Contains(line, "`"+l+"`") {
						cur, idx = l, j
						break
					}
				}
				if cur == "" {
					continue
				}
				if headingFound[idx] {
					violation("%s:%d: duplicate section for angle `%s` (first seen at line %d)", rel, lineno, cur, headingLine[idx])
				}
				if idx < lastIdx {
					violation("%s:%d: section for angle `%s` is out of order (expected after `%s`)", rel, lineno, cur, labels[lastIdx])
				}
				headingFound[idx], headingLine[idx] = true, lineno
				lastIdx = max(lastIdx, idx)
				continue
			}
			// A heading at ANY OTHER level (`#`, `###`, …) also ends the
			// current section, without becoming a new one — the report
			// shape's own headings are `##` (F3: a `###` heading once fell
			// through both checks and was folded into the previous section).
			if srrHeading.MatchString(line) {
				cur = ""
				continue
			}

			if cur == "" {
				// A finding-shaped line before any recognized section heading
				// (F6 — such a line was once dropped in silence).
				if srrFindingShape.MatchString(line) {
					violation("%s:%d: finding line appears before any recognized section heading", rel, lineno)
				}
				continue
			}

			if line == srrNoneMarker {
				hasNone[idx] = true
				continue
			}
			// The strict pattern is tested BEFORE the loose one (G2 — round
			// 2), so a well-formed line is never also reported as malformed.
			if m := srrFindingLine.FindStringSubmatch(line); m != nil && (!utf8Strict || utf8.ValidString(line)) {
				label, text := m[1], m[2]
				findings[idx]++
				if label != cur {
					violation("%s:%d: finding line label `%s` does not match its section `%s`", rel, lineno, label, cur)
				}
				// The disposition is the text after the LAST `— `.
				disposition := text
				if k := strings.LastIndex(text, "— "); k >= 0 {
					disposition = text[k+len("— "):]
				}
				if disposition == "declined" {
					continue
				}
				fm := srrFiled.FindStringSubmatch(disposition)
				switch {
				case fm == nil:
					violation("%s:%d: finding line disposition is neither `filed: <KEY>` nor `declined`", rel, lineno)
				case fm[1] == "":
					violation("%s:%d: finding line marked filed with no issue key", rel, lineno)
				case !srrIssueKey.MatchString(fm[1]):
					violation("%s:%d: finding line marked filed with a malformed issue key `%s` (want an uppercase project key, a hyphen, digits — e.g. KAN-201)", rel, lineno, fm[1])
				}
				continue
			}
			// Finding-shaped but not well-formed — e.g. more than one space
			// after the `-` (G2 — round 2). Without this branch such a line
			// was silently dropped: not counted, not flagged. A plain prose
			// line matches neither pattern and stays ignored, as ordinary
			// body prose should.
			if srrFindingShape.MatchString(line) {
				violation("%s:%d: finding-shaped line is malformed (want \"- **[%s]** <text> — filed: <KEY>\" or \"- **[%s]** <text> — declined\")", rel, lineno, cur, cur)
			}
		}

		found := 0
		for _, h := range headingFound {
			if h {
				found++
			}
		}
		required := n
		if strings.Contains(fiveAngle, "\n"+base+"\n") {
			required = srrOriginalAngles
		}

		// A report with no angle heading at all gets no per-section findings
		// — there was nothing in it to check against — and is flagged
		// through coverage's undeclared zero instead (KAN-73's regression
		// shape). A recognizable report gets a "missing section" finding for
		// each absent angle, a listed five-angle report for the first five
		// only.
		count := 0
		if found > 0 {
			for j, label := range labels {
				if !headingFound[j] {
					if j < required {
						violation("%s: missing section for angle `%s`", rel, label)
					}
					continue
				}
				count++
				if !hasNone[j] && findings[j] == 0 {
					violation("%s:%d: section for angle `%s` carries neither a finding line nor the none-marker", rel, headingLine[j], label)
				}
				if hasNone[j] && findings[j] > 0 {
					violation("%s:%d: section for angle `%s` carries both the none-marker and %d finding line(s)", rel, headingLine[j], label, findings[j])
				}
				count += findings[j]
			}
		}
		if err := cov.record(rel, count); err != nil {
			fmt.Fprintln(stderr, err)
			return die("coverage_record failed for '%s' (see stderr above)", rel)
		}
		checked++
	}

	// COVERAGE — a report checked for nothing is folded into the ordinary
	// violation count, never a separate exit status; an empty corpus is
	// reported the same way, by coverage's own empty-corpus check. Each
	// verdict line is split as the bash split it: the member before the
	// first `:`, the message after the first `: `.
	for _, l := range cov.verdict() {
		member, _, _ := strings.Cut(l, ":")
		_, msg, _ := strings.Cut(l, ": ")
		violation("%s:0: %s", member, msg)
	}

	if violations > 0 {
		fmt.Fprintf(stderr, "check-self-review-report: %d violation(s) across %d report(s) checked\n", violations, checked)
		return 1
	}
	fmt.Fprintf(stdout, "SELF-REVIEW-REPORT-OK: %s — %d report(s) checked\n", target, checked)
	if frag := cov.report(); frag != "" {
		fmt.Fprintf(stdout, "  %s\n", frag)
	}
	return 0
}

// srrAngleLabels is the bash's awk over the canonical table: each line
// split on `|`, a row whose second field is a bare number yields its last
// non-empty field, trimmed and with its backticks removed. A contract that
// cannot be read yields nothing, which the caller refuses.
func srrAngleLabels(path string) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var labels []string
	for _, line := range strings.Split(string(b), "\n") {
		f := strings.Split(line, "|")
		if len(f) < 2 || !srrAngleRow.MatchString(f[1]) {
			continue
		}
		label := ""
		for i := len(f) - 1; i >= 0; i-- {
			if c := strings.Trim(f[i], " \t\n\v\f\r"); c != "" {
				label = c
				break
			}
		}
		// `read -r` then trims the blanks the removed backticks exposed.
		label = strings.Trim(strings.ReplaceAll(label, "`", ""), " \t")
		if label != "" {
			labels = append(labels, label)
		}
	}
	return labels
}

// srrReports is the bash's `find "$TARGET" -maxdepth 1 -type f -name '*.md'
// -not -name '*-context.md'` sorted by `sort -z`: each report spelled as
// find printed it, in `sort`'s order under the caller's locale.
//
// Self-review context bundles are not reports (KAN-512): Run 2 commits
// docs/self-review/<name>-context.md on every run and /flow-self-review
// deletes it once it runs the deferred pass, so a pending bundle is neither
// scanned for the angle shape nor flagged as an undeclared zero.
//
// A failing walk is "cannot answer" (exit 2), never folded into an empty
// corpus at exit 1 (F2). find failed on a directory it could read but not
// search, and on any entry it could not stat; both are reproduced here, with
// find's own message ahead of the refusal.
func srrReports(env Env, target, abs string, stderr io.Writer) ([]string, int) {
	fail := func(p string, err error) ([]string, int) {
		fmt.Fprint(stderr, gsFindErr(p, err))
		fmt.Fprintf(stderr, "check-self-review-report: find failed while scanning '%s' (permission denied, or another find error)\n", target)
		return nil, 2
	}
	// find's default -P never follows a symlinked starting point: given a
	// symlink to a directory without a trailing `/`, `-maxdepth 1 -type f`
	// tests the link itself, matches nothing, and the corpus is empty.
	if !strings.HasSuffix(target, "/") {
		if fi, err := os.Lstat(abs); err == nil && fi.Mode()&os.ModeSymlink != 0 {
			return nil, 0
		}
	}
	if err := syscall.Access(abs, 1); err != nil {
		return fail(target, err)
	}
	ents, err := os.ReadDir(abs)
	if err != nil {
		return fail(target, err)
	}
	sep := "/"
	if strings.HasSuffix(target, "/") {
		sep = ""
	}
	var files []string
	for _, e := range ents {
		fi, err := os.Lstat(filepath.Join(abs, e.Name()))
		if err != nil {
			return fail(target+sep+e.Name(), err)
		}
		name := e.Name()
		if fi.Mode().IsRegular() && strings.HasSuffix(name, ".md") && !strings.HasSuffix(name, "-context.md") {
			files = append(files, target+sep+name)
		}
	}
	// The bash's `sort -z` over find's -print0 list: NUL-joined, so a report
	// name carrying a newline stays one path.
	files, err = crSortSep(env, files, "\x00", "-z")
	if err != nil {
		fmt.Fprintf(stderr, "check-self-review-report: cannot sort the report list: %v\n", err)
		return nil, 2
	}
	return files, 0
}

// srrUTF8Locale reports whether the caller's character type is UTF-8, by
// POSIX precedence (LC_ALL, then LC_CTYPE, then LANG). Under it bash's `.`
// matches no invalid UTF-8 byte, so a line carrying one fails the strict
// finding pattern; RE2's `.` matches any byte, so the port asks explicitly.
func srrUTF8Locale(env Env) bool {
	for _, k := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		if v := strings.ToLower(env.Getenv(k)); v != "" {
			return strings.Contains(v, "utf-8") || strings.Contains(v, "utf8")
		}
	}
	return false
}
