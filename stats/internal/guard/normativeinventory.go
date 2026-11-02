package guard

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/check-normative-inventory.sh shim (KAN-1058).
//
// flow-guard check-normative-inventory — print every normative sentence in this
// repository's owned Markdown, one per line, sorted.
//
// It exists because nothing else in `## lint` would notice a deleted
// requirement. flow-guard check-references and flow-guard check-installed-citations catch a
// citation that stops resolving; flow-guard check-markdown-integrity catches structural
// damage; the harnesses under `## test` exercise script behaviour. A sentence
// carrying a SHALL can be deleted with every one of them green — which is the
// failure mode a corpus-wide prose trim risks, and the reason this guard is a
// precondition of one. A change that edits prose in bulk captures this output
// before its first edit and after its last, and requires the two to be
// byte-identical; a difference is resolved by restoring the sentence, never by
// accepting the new inventory.
//
// THIS GUARD REPORTS A SET. It has no violation exit code, because comparing two
// sets is the caller's act — `diff <(flow-guard check-normative-inventory)
// normative-baseline.txt` is the whole check. Exit codes: 0 the inventory
// printed, 2 it cannot answer at all (a scope root missing or unreadable, a file
// it cannot read, an override set but empty).
//
// A NORMATIVE SENTENCE is one containing `SHALL`, `SHALL NOT`, `MUST` or
// `MUST NOT` as a whole word. Nothing else is a keyword: `SHOULD` and `MAY`
// state no obligation whose loss changes behaviour, and including them would
// make the inventory churn on ordinary editing. The match is case-sensitive and
// word-bounded, so `MUST_NOT_TOUCH` and `shall` are not keywords.
//
// THE SENTENCE BOUNDARY RULE, in full, because the guard's whole value rests on
// it being stable:
//
//   1. BLOCKS. A block is a heading line, one table cell, one list item, one
//      line inside a fenced code block, or a run of consecutive non-blank prose
//      lines. Each source line is stripped of leading and trailing whitespace and
//      of any blockquote (`>`) markers before it is classified, and the lines of
//      a prose run are joined with a single space. Joining is what makes the
//      output REFLOW-INVARIANT: the same wording rewrapped across different line
//      boundaries, at any indentation, yields the same block and therefore the
//      same output. A blank line, a heading, a list marker, a table row, a
//      thematic break or a fence ends the block being accumulated.
//   2. SENTENCES. A block is split at `.`, `!` or `?` followed by whitespace,
//      with any closing quote, bracket, backtick or emphasis characters riding
//      with the terminator (`... X.**` and `... X.)` end there too). The block's
//      tail, terminated or not, is the last sentence.
//   3. OUTPUT. A sentence is printed when it carries a keyword. Its internal
//      whitespace is collapsed to single spaces and its ends stripped; NOTHING
//      ELSE is normalised, because case, punctuation and wording are the content
//      being protected. Lines are sorted with LC_ALL=C, so the output is
//      `diff`-comparable between runs and independent of file order. Duplicates
//      are KEPT: the same sentence restated in two files must be printed twice,
//      or deleting one of the two copies would leave the inventory unchanged.
//
// ABBREVIATIONS are exempt from rule 2. A period closing `e.g.`, `i.e.`, `cf.`,
// `vs.`, `etc.`, `al.`, `Dr.` or `No.` — matched case-insensitively on the token
// before it — is not a terminator, and the sentence carries on past it. Without
// that exemption the splitter truncates a requirement at its first abbreviation
// and inventories the stub alone, which is worse than missing the sentence
// outright: a stub is a STABLE line that keeps matching while the obligation
// behind it is replaced wholesale, and counting keyword occurrences cannot see
// the swap either, because the keyword never moves.
//
// WHAT REMAINS, stated as what it is rather than as a coverage gap:
//
//   * An abbreviation OUTSIDE the list above ("Fig.", "approx.", "Prof.") still
//     ends a sentence early, and a substantive change to the text after it can
//     hide behind the stub — the defect described above, unfixed for exactly
//     those spellings. The corpus carries none of them today; the fix for a new
//     one is a word added to the list, in this file, visible in its diff.
//   * A sentence that genuinely ENDS in a listed abbreviation is merged with the
//     sentence after it. That direction is safe: both texts remain in the line,
//     so no change to either can hide. The cost is a coarser diff, never a
//     missed edit.
//
// Fenced content is scanned line by line rather than joined, because code is not
// reflowable prose; a worked example carrying a SHALL is therefore inventoried,
// which is correct, since a worked example is one of the things a trim may never
// cut.
//
// The corpus is ownedCorpus (stats/internal/guard/ownedcorpus.go), the one
// definition check-ask-silence walks too; the mechanism is
// stats/internal/guard/normativeinventory.go.
//
// The repository root is this shim's own parent, exported as
// FLOW_GUARD_REPO_ROOT — that is what makes the guard argument-free and
// self-scoped from any cwd. CHECK_NORMATIVE_INVENTORY_ROOT is an explicit,
// opt-in override honored only when set, for a sandboxed fixture tree — never
// set it for a normal invocation; set-but-empty is refused.

// checkNormativeInventory is scripts/check-normative-inventory.sh's guard. The
// shim's header states the contract — what a normative sentence is, the
// sentence boundary rule, the abbreviation exemption and the exit codes; this
// file is the mechanism.
func init() {
	Registry["check-normative-inventory"] = checkNormativeInventory
}

const niName = "check-normative-inventory"

var (
	niKeyword = regexp.MustCompile(`(^|[^A-Za-z0-9_])(SHALL|MUST)([^A-Za-z0-9_]|$)`)
	// niTerminator is `.`, `!` or `?` followed by whitespace, any closing
	// quote, bracket, backtick or emphasis characters riding with it. Longest
	// is awk's leftmost-longest match.
	niTerminator = func() *regexp.Regexp {
		re := regexp.MustCompile("[.!?][]\"'`*_)]* +")
		re.Longest()
		return re
	}()
	niBreak    = regexp.MustCompile(`^(---+|___+|\*\*\*+|===+)$`)
	niListItem = regexp.MustCompile(`^([-*+]|[0-9]+[.)])[ \t]+`)
	niSpaces   = regexp.MustCompile(`[ \t]+`)
)

// niAbbrevs are the abbreviations whose trailing period does NOT end a
// sentence (the header's ABBREVIATIONS paragraph).
var niAbbrevs = map[string]bool{"e.g.": true, "i.e.": true, "cf.": true, "vs.": true, "etc.": true, "al.": true, "dr.": true, "no.": true}

func checkNormativeInventory(_ []string, env Env, stdout, stderr io.Writer) int {
	// The root is this repository's own root by default -- the parent of
	// scripts/, derived as the bash guard derived it from its own location;
	// the binary lives in a cache, so the shim exports it as
	// FLOW_GUARD_REPO_ROOT. CHECK_NORMATIVE_INVENTORY_ROOT is an explicit,
	// opt-in override honoured only when set, so the tests can point this
	// guard at a sandboxed fixture tree -- never set it for a normal
	// invocation. Set but empty is refused rather than treated as unset:
	// falling back there would scan the real repository while believing it
	// scanned a fixture.
	root, set := env.LookupEnv("CHECK_NORMATIVE_INVENTORY_ROOT")
	switch {
	case set && root == "":
		fmt.Fprintf(stderr, "CHECK_NORMATIVE_INVENTORY_ROOT is set but empty\n")
		return 2
	case !set:
		if root = env.Getenv("FLOW_GUARD_REPO_ROOT"); root == "" {
			fmt.Fprintf(stderr, "%s: FLOW_GUARD_REPO_ROOT is unset — run flow-guard check-normative-inventory, which sets it\n", niName)
			return 2
		}
	}
	fsRoot := smcAbs(env, root)
	if !isDir(fsRoot) {
		fmt.Fprintf(stderr, "not a directory: %s\n", root)
		return 2
	}

	// The corpus walk gets the root as spelled, so its refusals and the
	// file paths below print as the bash printed them; a relative root
	// resolves against the process's working directory, the caller's.
	files, code := ownedCorpus(root, "owned_corpus_files", stderr)
	if files == nil {
		return code
	}
	// An empty corpus is refused, not reported as an empty inventory. The two
	// are indistinguishable on stdout, and one of them is "every normative
	// sentence in this repository has just been deleted" — the exact
	// comparison this guard feeds.
	if len(files) == 0 {
		fmt.Fprintf(stderr, "no .md or .mdc file found under %s — refusing to report an empty inventory\n", root)
		return 2
	}

	var sentences []string
	for _, path := range files {
		b, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(stderr, "cannot read %s\n", path)
			return 2
		}
		sentences = niScan(string(b), sentences)
	}

	// Byte order is LC_ALL=C sort's; duplicates are kept.
	sort.Strings(sentences)
	for _, s := range sentences {
		fmt.Fprintln(stdout, s)
	}
	// The counts go to STDERR, never stdout: stdout is the set the caller
	// diffs, and a verdict line in it would be a difference of its own every
	// time a sentence was legitimately added elsewhere. They are printed at
	// all because a silent run and a run that scanned nothing look identical
	// otherwise.
	fmt.Fprintf(stderr, "%s: %d normative sentence(s) from %d file(s) under %s\n", niName, len(sentences), len(files), root)
	return 0
}

// niScan appends one file's normative sentences to out — the bash guard's
// awk program, run once per file, so fenced mode and the open block never
// carry across files.
func niScan(text string, out []string) []string {
	block := ""
	inFence := false
	emit := func(s string) { out = niEmit(s, out) }
	flush := func() {
		if block != "" {
			emit(block)
			block = ""
		}
	}
	for _, line := range fileLines(text) {
		line = strings.TrimSuffix(line, "\r")
		line = strings.Trim(line, " \t")

		// A fence line ends the block it interrupts and toggles fenced mode.
		// Inside a fence every line is its own block: code is not reflowable
		// prose, so joining its lines would be a fiction, and dropping them
		// would lose a worked example.
		if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~") {
			flush()
			inFence = !inFence
			continue
		}
		if inFence {
			emit(line)
			continue
		}

		// Blockquote markers are stripped before classification, so a quoted
		// heading, bullet or paragraph is read exactly as an unquoted one.
		// rules/*.mdc carry normative text inside block quotes; without this
		// they would be one unreflowable line each.
		for strings.HasPrefix(line, ">") {
			line = strings.TrimPrefix(line, ">")
			if line != "" && (line[0] == ' ' || line[0] == '\t') {
				line = line[1:]
			}
			line = strings.TrimLeft(line, " \t")
		}
		line = strings.TrimRight(line, " \t")

		switch {
		case line == "" || niBreak.MatchString(line):
			flush()
		case niIsHeading(line):
			flush()
			emit(line)
		case strings.HasPrefix(line, "|"):
			// A table row is split on the cell separator: each cell is an
			// independent unit, so editing one cell cannot perturb the
			// inventory line another cell produces. agents-baseline.md's
			// rule table is one requirement per cell.
			flush()
			for _, cell := range strings.Split(line, "|") {
				emit(cell)
			}
		case niListItem.MatchString(line):
			flush()
			block = niListItem.ReplaceAllString(line, "")
		case block == "":
			block = line
		default:
			block += " " + line
		}
	}
	flush()
	return out
}

// niIsHeading is awk's /^#+ /.
func niIsHeading(line string) bool {
	rest := strings.TrimLeft(line, "#")
	return len(rest) < len(line) && strings.HasPrefix(rest, " ")
}

// niEmit normalises one block, splits it into sentences, and appends the
// normative ones.
func niEmit(s string, out []string) []string {
	s = strings.Trim(niSpaces.ReplaceAllString(s, " "), " ")
	if s == "" {
		return out
	}
	rest, pending := s, ""
	for {
		loc := niTerminator.FindStringIndex(rest)
		if loc == nil {
			break
		}
		candidate := pending + rest[:loc[1]]
		rest = rest[loc[1]:]
		// An abbreviation's period is not a boundary: carry the text forward
		// and keep looking for the real terminator.
		if niIsAbbrev(candidate) {
			pending = candidate
			continue
		}
		if sentence := strings.TrimRight(candidate, " "); niKeyword.MatchString(sentence) {
			out = append(out, sentence)
		}
		pending = ""
	}
	if rest = pending + rest; niKeyword.MatchString(rest) {
		out = append(out, rest)
	}
	return out
}

// niIsAbbrev reports whether text ends in an exempt abbreviation: its last
// whitespace-separated token, leading punctuation ("(e.g.") stripped and
// lower-cased, is in niAbbrevs.
//
// Without this the splitter truncates a requirement at its first
// abbreviation and inventories only the stub. That is worse than missing the
// sentence outright: the stub is a STABLE line that keeps matching while the
// obligation behind it is replaced wholesale — "The tool SHALL validate e.g.
// input for XSS and SQL injection before use." and "The tool SHALL validate
// e.g. nothing at all here whatsoever." both reduce to "The tool SHALL
// validate e.g." and the inventory does not move. A keyword-occurrence
// reconciliation cannot see it either, because the keyword never moves.
func niIsAbbrev(text string) bool {
	w := strings.TrimRight(text, " ")
	if !strings.HasSuffix(w, ".") {
		return false
	}
	last := strings.TrimLeftFunc(w[strings.LastIndex(w, " ")+1:], func(r rune) bool {
		return !('a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || '0' <= r && r <= '9')
	})
	return niAbbrevs[niASCIILower(last)]
}

// niASCIILower is awk's tolower: ASCII letters only.
func niASCIILower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if 'A' <= c && c <= 'Z' {
			b[i] = c + 'a' - 'A'
		}
	}
	return string(b)
}
