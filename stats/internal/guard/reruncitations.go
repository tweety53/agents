package guard

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// checkRerunCitations is scripts/check-rerun-citations.sh: that script's
// header comment is the contract. It judges whether a fix-round re-review
// read the fix it reports on from the report file itself: every named
// finding needs a `verdict: F<n> — fixed|not fixed — <citation>` line, and a
// `fixed` verdict's `<path>:<line>` must land inside a hunk of the diff the
// re-review was given.
func init() {
	Registry["check-rerun-citations"] = checkRerunCitations
}

var (
	rcRef      = regexp.MustCompile(`^F[0-9]+$`)
	rcCite     = regexp.MustCompile(`^(.+):([0-9]+)$`)
	rcHunkHead = regexp.MustCompile(`^@@ -([0-9]+)(?:,([0-9]+))? \+([0-9]+)(?:,([0-9]+))? @@`)
)

func checkRerunCitations(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-rerun-citations"
	if len(args) < 3 {
		fmt.Fprintf(stderr, "%s: usage: check-rerun-citations.sh <report> <diff> F<n> [F<n>…]\n", prog)
		return 2
	}
	var files [2]string
	for i, p := range args[:2] {
		b, err := os.ReadFile(rcPath(env, p))
		if err != nil {
			fmt.Fprintf(stderr, "%s: cannot read %s\n", prog, p)
			return 2
		}
		files[i] = string(b)
	}
	report, diff := files[0], files[1]
	refs := args[2:]
	for _, ref := range refs {
		if !rcRef.MatchString(ref) {
			fmt.Fprintf(stderr, "%s: not a finding ref shaped F<n>: '%s'\n", prog, ref)
			return 2
		}
	}

	var lines []string
	for _, l := range strings.Split(report, "\n") {
		l = strings.ReplaceAll(l, "`", "")
		l = strings.TrimPrefix(strings.TrimPrefix(l, "- "), "* ")
		lines = append(lines, strings.TrimRight(l, " \r"))
	}

	status := 0
	for _, ref := range refs {
		prefix := "verdict: " + ref + " — "
		line := ""
		for _, l := range lines {
			if strings.HasPrefix(l, prefix) {
				line = l
				break
			}
		}
		if line == "" {
			fmt.Fprintf(stdout, "UNCITED: %s — no `verdict: %s — …` line in the report\n", ref, ref)
			status = 1
			continue
		}
		rest := strings.TrimPrefix(line, prefix)
		switch {
		case strings.HasPrefix(rest, "not fixed — ") && len(rest) > len("not fixed — "):
			fmt.Fprintf(stdout, "CITED: %s — %s\n", ref, rest)
		case strings.HasPrefix(rest, "fixed — ") && len(rest) > len("fixed — "):
			cite := strings.TrimPrefix(rest, "fixed — ")
			m := rcCite.FindStringSubmatch(cite)
			if m == nil {
				fmt.Fprintf(stdout, "UNCITED: %s — fixed — '%s' is not <path>:<line>\n", ref, cite)
				status = 1
				continue
			}
			path := m[1][strings.LastIndex(m[1], ":")+1:] // a worktree prefix is stripped
			path = strings.TrimPrefix(path, "./")
			n, _ := strconv.Atoi(m[2])
			if rcInHunk(diff, path, n) {
				fmt.Fprintf(stdout, "CITED: %s — fixed — %s\n", ref, cite)
			} else {
				fmt.Fprintf(stdout, "UNCITED: %s — fixed — %s lands in no hunk of %s\n", ref, cite, args[1])
				status = 1
			}
		default:
			fmt.Fprintf(stdout, "UNCITED: %s — '%s' is neither `fixed — <path>:<line>` nor `not fixed — <citation>`\n", ref, line)
			status = 1
		}
	}
	return status
}

func rcPath(env Env, p string) string {
	if p != "" && !strings.HasPrefix(p, "/") && env.Dir != "" {
		return env.Dir + "/" + p
	}
	return p
}

// rcInHunk reports whether line n of path falls inside a hunk of diff, on
// the new side (`+++ b/<path>`) or the old side (`--- a/<path>`), so a
// deleted file's lines are citable too.
func rcInHunk(diff, path string, n int) bool {
	var oldPath, newPath string
	header := func(l, side string) string {
		p, _, _ := strings.Cut(l[4:], "\t")
		return strings.TrimPrefix(p, side)
	}
	count := func(s string) int {
		if s == "" {
			return 1
		}
		c, _ := strconv.Atoi(s)
		return c
	}
	in := func(start, cnt string) bool {
		s, _ := strconv.Atoi(start)
		c := count(cnt)
		e := s + c - 1
		if c == 0 {
			e = s
		}
		return n >= s && n <= e
	}
	// oldLeft/newLeft are the lines still owed to the open hunk: inside it a
	// removed `-- x` reads `--- x` and an added `++ x` reads `+++ x`, which are
	// body lines, not file headers.
	oldLeft, newLeft := 0, 0
	for _, l := range strings.Split(diff, "\n") {
		if oldLeft > 0 || newLeft > 0 {
			switch {
			case strings.HasPrefix(l, "-"):
				oldLeft--
			case strings.HasPrefix(l, "+"):
				newLeft--
			case strings.HasPrefix(l, "\\"):
			default:
				oldLeft--
				newLeft--
			}
			continue
		}
		switch {
		case strings.HasPrefix(l, "--- "):
			oldPath = header(l, "a/")
		case strings.HasPrefix(l, "+++ "):
			newPath = header(l, "b/")
		case strings.HasPrefix(l, "@@ "):
			m := rcHunkHead.FindStringSubmatch(l)
			if m == nil {
				continue
			}
			if (newPath == path && in(m[3], m[4])) || (oldPath == path && in(m[1], m[2])) {
				return true
			}
			oldLeft, newLeft = count(m[2]), count(m[4])
		}
	}
	return false
}
