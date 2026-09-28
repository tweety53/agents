package guard

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"syscall"
)

// checkVisualTrigger is scripts/check-visual-trigger.sh: that script's
// header comment is the contract -- the changed paths on stdin, one MATCH
// line per (path, glob) pair that matched and one verdict line on stdout,
// exit 0 matched, 1 configured and unmatched, 2 cannot answer. visualTrigger
// is the same question with the changed paths already read, for an
// in-process caller.
func init() {
	Registry["check-visual-trigger"] = checkVisualTrigger
}

const vtPrefix = "check-visual-trigger: "

func checkVisualTrigger(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintf(stderr, "%susage: check-visual-trigger.sh <project root> (changed paths on stdin)\n", vtPrefix)
		return 2
	}
	// `while IFS= read -r changed || [ -n "$changed" ]`: a final line with
	// no newline is still read, a CR is kept, and bash's read drops a NUL
	// byte. A nil Env.Stdin is empty input. The bash reached its read loop
	// only after every refusal, so stdin is read only when visualTrigger
	// asks for the paths. A stdin that cannot be read is refused, never read
	// as "no changed paths".
	changed := func() ([]string, error) {
		var in []byte
		if env.Stdin != nil {
			var err error
			if in, err = io.ReadAll(env.Stdin); err != nil {
				return nil, err
			}
		}
		return awkRecords(strings.ReplaceAll(string(in), "\x00", "")), nil
	}
	return visualTrigger(env, args[0], changed, stdout, stderr)
}

// visualTrigger is the guard's whole answer for one project root and the
// diff's changed paths. root is printed as given; a relative one is resolved
// against env.Dir, as the bash resolved it against its working directory.
// changed is called once, after every refusal about the root and its
// config, for the diff's paths; its read error is the last refusal.
func visualTrigger(env Env, root string, changed func() ([]string, error), stdout, stderr io.Writer) int {
	fsRoot := root
	if root != "" && !strings.HasPrefix(root, "/") {
		fsRoot = env.Dir + "/" + root
	}
	if root == "" || !isDir(fsRoot) {
		fmt.Fprintf(stderr, "%s%s is not a directory — cannot tell whether it declares a visual verification section\n", vtPrefix, root)
		return 2
	}
	cfg, fsCfg := root+"/.flow/project.md", fsRoot+"/.flow/project.md"

	// "Not configured" (no `.flow/project.md` at all) is the SAME
	// cannot-answer case as "file exists but declares no section" below --
	// both are exit 2, never exit 1's "configured and unmatched". A dangling
	// symlink exists, and falls to "not a regular file".
	if _, err := os.Lstat(fsCfg); err != nil {
		fmt.Fprintf(stderr, "%s%s has no .flow/project.md — not configured, so nothing was matched\n", vtPrefix, root)
		return 2
	}
	if !isFile(fsCfg) {
		fmt.Fprintf(stderr, "%s%s is not a regular file — cannot resolve what it declares\n", vtPrefix, cfg)
		return 2
	}
	if syscall.Access(fsCfg, 4) != nil {
		fmt.Fprintf(stderr, "%s%s exists but is not readable — cannot resolve what it declares\n", vtPrefix, cfg)
		return 2
	}
	b, err := os.ReadFile(fsCfg)
	if err != nil {
		// The bash's grep exiting 2: a failure to look, not an absence.
		fmt.Fprintf(stderr, "%sgrep exited 2 while looking for the '## visual verification' heading in %s — that is a failure to look, not an absence\n", vtPrefix, cfg)
		return 2
	}
	text := string(b)
	switch n := vvHeadingCount(text); {
	case n == 0:
		fmt.Fprintf(stderr, "%s%s declares no '## visual verification' section — not configured, so nothing was matched\n", vtPrefix, cfg)
		return 2
	case n > 1:
		fmt.Fprintf(stderr, "%s%s declares %d '## visual verification' sections — a second declaration is ambiguous, so neither was read\n", vtPrefix, cfg, n)
		return 2
	}

	// The whole of the "ui paths" extraction: the first two-cell `ui paths`
	// row inside the section. This guard does not re-validate the rest of
	// the section's shape; check-visual-verification already does that, and
	// this guard's own job is narrower: resolve one field or say it cannot.
	value := ""
	for _, l := range vvSectionLines(text) {
		if cells := vtSplitCells(l.Text); len(cells) == 2 && vtFoldCell(cells[0]) == "ui paths" {
			value = vtTrimCell(cells[1])
			break
		}
	}
	if value == "" {
		fmt.Fprint(stderr, sanitizeDisplay(vtPrefix+"`ui paths` is absent or empty in "+cfg+" — cannot resolve what to match against\n"))
		return 2
	}

	// Split on comma first, then strip each element's own backticks and
	// surrounding whitespace with trimGlobElement -- NEVER an interior
	// space, so a glob containing one survives the split intact. The value
	// already went through the table parser's whole-cell vtTrimCell, which
	// strips only the outermost pair; each element's own backticks are
	// stripped only now.
	var globs []string
	for _, g := range strings.Split(value, ",") {
		if g = trimGlobElement(g); g != "" {
			globs = append(globs, g)
		}
	}
	if len(globs) == 0 {
		fmt.Fprint(stderr, sanitizeDisplay(vtPrefix+"`ui paths` in "+cfg+" resolved to no usable glob\n"))
		return 2
	}
	res := make([]*regexp.Regexp, len(globs))
	for i, g := range globs {
		res[i] = vtGlobRegexp(g)
	}

	paths, err := changed()
	if err != nil {
		fmt.Fprintf(stderr, "%scannot read stdin: %v\n", vtPrefix, err)
		return 2
	}
	matched := false
	for _, path := range paths {
		if path == "" {
			continue
		}
		p := vtBytesAsRunes(strings.TrimPrefix(path, "./"))
		for i, re := range res {
			if re.MatchString(p) {
				fmt.Fprint(stdout, sanitizeDisplay("MATCH: "+path+" — matched `"+globs[i]+"`\n"))
				matched = true
			}
		}
	}
	if matched {
		fmt.Fprintf(stdout, "VISUAL-TRIGGER-MATCH: %s — at least one changed path matched `ui paths`\n", root)
		return 0
	}
	fmt.Fprintf(stdout, "VISUAL-TRIGGER-NO-MATCH: %s — no changed path matched `ui paths`\n", root)
	return 1
}

// vtGlobRegexp is the bash's glob_to_ere plus glob_match's anchoring: a
// leading `./` and then a leading `/` stripped from the glob; `**` becomes
// `.*` (spans `/`); a bare `*` becomes `[^/]*` (one segment); `?` becomes
// `[^/]`; every listed metacharacter is escaped; anything else is literal.
//
// Byte-wise, as the bash matched under LC_ALL=C (c-locale-table-semantics):
// both the glob and the path go through vtBytesAsRunes, one rune per byte,
// so `?` is one byte and a non-UTF-8 byte is an ordinary character.
func vtGlobRegexp(glob string) *regexp.Regexp {
	glob = strings.TrimPrefix(strings.TrimPrefix(glob, "./"), "/")
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(glob); i++ {
		switch c := glob[i]; c {
		case '*':
			if i+1 < len(glob) && glob[i+1] == '*' {
				b.WriteString(".*")
				i++
				continue
			}
			b.WriteString("[^/]*")
		case '?':
			b.WriteString("[^/]")
		case '.', '+', '(', ')', '{', '}', '|', '^', '$', '\\', '[', ']':
			b.WriteByte('\\')
			b.WriteByte(c)
		default:
			b.WriteRune(rune(c))
		}
	}
	b.WriteString("$")
	return regexp.MustCompile(b.String())
}

// vtBytesAsRunes maps each byte of s to the rune of the same value, so a
// regexp over the result matches byte by byte.
func vtBytesAsRunes(s string) string {
	r := make([]rune, len(s))
	for i := 0; i < len(s); i++ {
		r[i] = rune(s[i])
	}
	return string(r)
}
