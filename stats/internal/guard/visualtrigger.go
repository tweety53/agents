package guard

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"syscall"
)

// The contract below moved here from the header of the deleted
// scripts/check-visual-trigger.sh shim (KAN-1058).
//
// flow-guard check-visual-trigger — decide whether a diff touched a project's
// declared `## visual verification` UI paths, in a guard instead of prose.
//
// Usage: flow-guard check-visual-trigger [--repo <name>] <project root>
//        (changed paths on stdin, one per line)
//
// --repo <name> matches each changed path as `<name>/<path>`: a cross-repo
// change's other worktree's diff, against the canonical project's `ui paths`,
// which name that repository's files by its main checkout's basename
// (`gymie-frontend/**`) — flow.visual-verify's step 2.
//
// Replaces flow.visual-verify's step 2 (skills/flow/verify-and-handoff.md):
// "run `git diff --name-only` and test every printed path as a glob against
// the comma-separated `ui paths`." Two runs reading the same globs
// differently was the defect this guard exists to make impossible — the
// glob semantics are the whole risk, per tasks.md task 9.
//
// THREE EXIT CODES, AND "NOT CONFIGURED" IS NEVER "NO MATCH":
//   0  at least one stdin path matched at least one `ui paths` glob
//   1  the section resolved cleanly and `ui paths` is non-empty, but no
//      stdin path (including an empty stdin — no diff at all) matched
//   2  cannot answer: the project root is not a directory, `.flow/project.md`
//      is missing, not a regular file, or unreadable, the heading scan
//      itself failed to look, the section is declared twice (ambiguous —
//      neither is read), or `ui paths` is absent or empty.
// "Not configured" (2) and "configured and unmatched" (1) are deliberately
// different answers — flow.visual-verify's step 1 and step 2 print two
// different messages and skip the stage for two different reasons.
// Every exit 2 ends its stderr with one cause token line, so a caller tells
// the two exit-2 answers apart without reading the prose above it:
//   VISUAL-TRIGGER-NOT-CONFIGURED: <root> — no visual verification section
//      `.flow/project.md` is missing, or declares no `## visual verification`
//   VISUAL-TRIGGER-CANNOT-ANSWER: <root> — <short cause>
//      every other exit 2 (a usage error prints `(no root)` as <root>)
//
// GLOB SEMANTICS, pinned because they already caused disagreement once:
//   - `**` matches any run of characters INCLUDING `/`, so it spans
//     multiple directory levels, not one segment.
//   - a bare `*` matches any run of characters EXCEPT `/` — one path
//     segment. (Every glob this repository declares today uses `**`; a
//     bare `*` is supported for completeness, not because a declaration
//     uses one.)
//   - `?` matches exactly one byte except `/`, and `*`/`**` count bytes too: the
//     port matches byte-wise whatever the caller's locale (c-locale-table-semantics).
//   - a leading `./` is stripped from BOTH the declared glob and the
//     candidate path before matching — a habit either side might write.
//   - a leading `/` on a DECLARED glob is stripped too: `.flow/project.md`
//     paths are inherently repository-root-relative, so `/stats/web/**` and
//     `stats/web/**` mean the same thing.
//   - an ABSOLUTE candidate path is never stripped and so never matches a
//     relative glob. `git diff --name-only` never emits one; a caller that
//     hands this guard one gets a defined refusal, not a silent grant.
//   - splitting `ui paths` on its separator comma never touches a SPACE
//     inside one glob — only leading/trailing whitespace around each
//     comma-separated element is trimmed.
//
// THE INPUT IS ATTACKER-INFLUENCED, exactly as flow-guard check-visual-verification's
// own header states: `.flow/project.md` is tracked and editable in any pull
// request. This guard never executes anything it reads — `ui paths` values
// are matched as glob patterns, never interpolated into a shell command —
// and every interpolated cell that reaches an exit-2 message or a MATCH line
// passes through sanitizeDisplay (visualsection.go) first — see its comment
// for the forged-verdict reason.
//
// THE HEADING SCAN AND TABLE PARSER ARE flow-guard check-visual-verification's OWN
// CONVENTIONS, reused rather than reinvented: the same `## visual
// verification` heading regex, the same "a heading at or above this
// section's own level ends it, a deeper one is prose" rule, and the same
// `|`-row split with `\|`-escape awareness. They are the Go twins in
// stats/internal/guard/visualsection.go (vvHeadingCount, vvSectionLines,
// vtSplitCells, vtTrimCell, vtFoldCell), whose comments carry the reasoning
// and TestVisualTableCellsParity its parity record.
//
// A LEADING UTF-8 BOM IS STRIPPED BEFORE EITHER HEADING READ OR THE TABLE
// PARSE, via the same file's stripBOM — see its comment for why a BOM would
// otherwise make this guard read a present `## visual verification` heading
// as absent and exit 2 ("not configured"), which is worse here than in the
// other two guards: this is the guard that decides whether flow.visual-verify
// runs at all, so that misread is a silent skip of the whole stage, not
// merely a misparse.
//
// `ui paths`' OWN PER-ELEMENT TRIM is trimGlobElement in the same file,
// shared with check-visual-verification — see its comment for the
// split-then-strip parse and the single-pass strip it performs.
//
// The guard is Go: stats/internal/guard/visualtrigger.go.

// checkVisualTrigger is `flow-guard check-visual-trigger`: the comment above is the contract -- the changed paths on stdin, one MATCH
// line per (path, glob) pair that matched and one verdict line on stdout,
// exit 0 matched, 1 configured and unmatched, 2 cannot answer. visualTrigger
// is the same question with the changed paths already read, for an
// in-process caller.
func init() {
	Registry["check-visual-trigger"] = checkVisualTrigger
}

const vtPrefix = "check-visual-trigger: "

// The two exit-2 cause tokens (VH-25). NOT-CONFIGURED is the project
// declaring no section at all -- `Visual: not configured` to a caller;
// CANNOT-ANSWER is every other refusal. Each exit 2 ends its stderr with one
// token line after the refusal it already printed, so a caller branches on
// the token and never on the prose.
const (
	vtNotConfigured = "VISUAL-TRIGGER-NOT-CONFIGURED"
	vtCannotAnswer  = "VISUAL-TRIGGER-CANNOT-ANSWER"
	vtNoSection     = "no visual verification section"
)

// vtRefuse prints the cause token line and returns exit 2. root is printed
// as given, as the verdict lines print it.
func vtRefuse(stderr io.Writer, token, root, cause string) int {
	fmt.Fprintf(stderr, "%s: %s — %s\n", token, root, cause)
	return 2
}

func checkVisualTrigger(args []string, env Env, stdout, stderr io.Writer) int {
	// `--repo <name>` matches each changed path as `<name>/<path>`: a
	// cross-repo change's other worktree's diff, matched against the
	// canonical project's `ui paths`, which name that repository's files by
	// its name (`gymie-frontend/**`).
	repo := ""
	if len(args) == 3 && args[0] == "--repo" && args[1] != "" {
		repo, args = args[1]+"/", args[2:]
	}
	if len(args) != 1 {
		fmt.Fprintf(stderr, "%susage: flow-guard check-visual-trigger [--repo <name>] <project root> (changed paths on stdin)\n", vtPrefix)
		return vtRefuse(stderr, vtCannotAnswer, "(no root)", "usage")
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
		paths := awkRecords(strings.ReplaceAll(string(in), "\x00", ""))
		if repo != "" {
			for i, p := range paths {
				if p != "" && !strings.HasPrefix(p, "/") {
					paths[i] = repo + strings.TrimPrefix(p, "./")
				}
			}
		}
		return paths, nil
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
		return vtRefuse(stderr, vtCannotAnswer, root, "not a directory")
	}
	cfg, fsCfg := root+"/.flow/project.md", fsRoot+"/.flow/project.md"

	// "Not configured" (no `.flow/project.md` at all) is the SAME
	// cannot-answer case as "file exists but declares no section" below --
	// both are exit 2, never exit 1's "configured and unmatched". A dangling
	// symlink exists, and falls to "not a regular file".
	if _, err := os.Lstat(fsCfg); err != nil {
		fmt.Fprintf(stderr, "%s%s has no .flow/project.md — not configured, so nothing was matched\n", vtPrefix, root)
		return vtRefuse(stderr, vtNotConfigured, root, vtNoSection)
	}
	if !isFile(fsCfg) {
		fmt.Fprintf(stderr, "%s%s is not a regular file — cannot resolve what it declares\n", vtPrefix, cfg)
		return vtRefuse(stderr, vtCannotAnswer, root, ".flow/project.md is not a regular file")
	}
	if syscall.Access(fsCfg, 4) != nil {
		fmt.Fprintf(stderr, "%s%s exists but is not readable — cannot resolve what it declares\n", vtPrefix, cfg)
		return vtRefuse(stderr, vtCannotAnswer, root, ".flow/project.md is not readable")
	}
	b, err := os.ReadFile(fsCfg)
	if err != nil {
		// The bash's grep exiting 2: a failure to look, not an absence.
		fmt.Fprintf(stderr, "%sgrep exited 2 while looking for the '## visual verification' heading in %s — that is a failure to look, not an absence\n", vtPrefix, cfg)
		return vtRefuse(stderr, vtCannotAnswer, root, ".flow/project.md could not be read")
	}
	text := string(b)
	switch n := vvHeadingCount(text); {
	case n == 0:
		fmt.Fprintf(stderr, "%s%s declares no '## visual verification' section — not configured, so nothing was matched\n", vtPrefix, cfg)
		return vtRefuse(stderr, vtNotConfigured, root, vtNoSection)
	case n > 1:
		fmt.Fprintf(stderr, "%s%s declares %d '## visual verification' sections — a second declaration is ambiguous, so neither was read\n", vtPrefix, cfg, n)
		return vtRefuse(stderr, vtCannotAnswer, root, "the section is declared more than once")
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
		return vtRefuse(stderr, vtCannotAnswer, root, "`ui paths` is absent or empty")
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
		return vtRefuse(stderr, vtCannotAnswer, root, "`ui paths` resolved to no usable glob")
	}
	res := make([]*regexp.Regexp, len(globs))
	for i, g := range globs {
		res[i] = vtGlobRegexp(g)
	}

	paths, err := changed()
	if err != nil {
		fmt.Fprintf(stderr, "%scannot read stdin: %v\n", vtPrefix, err)
		return vtRefuse(stderr, vtCannotAnswer, root, "stdin could not be read")
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
