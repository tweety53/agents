package guard

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"unicode"
	"unicode/utf8"
)

func init() { Registry["check-workspace-isolation"] = checkWorkspaceIsolation }

const (
	wiResourceHeader = "resource|variable|default|in a workspace"
	wiCommandHeader  = "command|runs"
)

var (
	wiVariable = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	wiInteger  = regexp.MustCompile(`^[0-9]+$`)
	wiBracket  = regexp.MustCompile(`<[^<>]*>`)
	wiValueRef = regexp.MustCompile(`^value:[A-Za-z_][A-Za-z0-9_]*$`)
	wiDelim    = regexp.MustCompile(`^:?-+:?$`)
)

// checkWorkspaceIsolation is scripts/check-workspace-isolation.sh: that
// script's header comment carries the contract.
//
// Divergences from the bash, each on input the bash read by accident of its
// tools rather than by its contract:
//   - Under a UTF-8 locale macOS awk aborts ("towc: multibyte conversion
//     failure") on a line that is not valid UTF-8, so the bash refused the
//     whole file as "the validator did not run to completion"; this port
//     reads such bytes as bytes.
//   - The heading rule is ccHeading over ASCII-lowercased text, the cleanup
//     guard's own — the rule the bash named as the one it shares with that
//     guard. The bash's grep and awk also took U+00A0 for the space in a
//     heading under a UTF-8 locale, so a `#<U+00A0>other` line also ended the
//     section there and does not here; a section the cleanup guard does not
//     read is not one this guard validates either.
//   - awk -v processed backslash escapes in the configuration's path before
//     quoting it in a violation line; the path is quoted here as given.
func checkWorkspaceIsolation(args []string, env Env, stdout, stderr io.Writer) int {
	refuse := func(msg string) int {
		fmt.Fprintln(stderr, "check-workspace-isolation: "+msg)
		return 2
	}
	// CHECK_WORKSPACE_ISOLATION_PRINT_ROWS — an internal, undocumented-to-
	// operators switch that prepare-workspace.sh sets when it invokes this
	// guard, so it can derive the workspace variables from the SAME parse this
	// guard already did rather than opening `.flow/project.md` a second time
	// and re-implementing the cell splitter. Unset (the default), it changes
	// nothing: no `#ROW` line is emitted, so stdout is byte-for-byte what it
	// always was. Set to `1`, each validated resource row is appended to this
	// project's report as a `#ROW\t<resource>\t<variable>\t<default>\t<in a
	// workspace>` line, after the verdict line, and only for a project whose
	// report reached its verdict — a project this guard refused on (exit 2)
	// never reaches the print that emits them. A caller that does not know
	// this switch exists reads exactly the output documented above; only
	// prepare-workspace.sh reads the `#ROW` lines.
	printRows := env.Getenv("CHECK_WORKSPACE_ISOLATION_PRINT_ROWS") == "1"

	// The self-resolution is deferred to exactly here, inside the branch that
	// is the only reader of its answer. Every OTHER caller passes an explicit
	// project root — `/flow`'s implement phase against each apply worktree,
	// and this repository's own `## lint` entry — and resolving
	// unconditionally meant a failure to resolve this script's own location
	// could abort a run that never needed the answer at all (F11).
	//
	// "One level above the script's directory" is NOT enough to derive the
	// repository root, because the script is reachable from more than one
	// directory — its real home at <repo>/scripts/, and a skills/<name>/scripts/
	// symlink a command skill carries it under. Going up one level from the
	// SECOND of those lands on the skill directory, which exists, so a
	// fixed-depth guard would proceed against it and return a confident wrong
	// answer (silently "declares nothing") rather than an error. The shim
	// therefore exports the path it was invoked by as FLOW_GUARD_SELF, and
	// resolveFile follows its symlinks first, so the root comes from the
	// script's real physical location whichever path invoked it.
	if len(args) == 0 {
		self := env.Getenv("FLOW_GUARD_SELF")
		if self == "" {
			return refuse("FLOW_GUARD_SELF is unset — run scripts/check-workspace-isolation.sh, which sets it")
		}
		real, ok := resolveFile(wiAbs(env, self))
		if !ok {
			return refuse("cannot resolve this script's own location")
		}
		args = []string{gdcDirname(gdcDirname(real))}
	}

	utf8Locale := smcUTF8(env)
	total := 0
	for _, root := range args {
		if !isDir(wiAbs(env, root)) {
			return refuse(root + " is not a directory — cannot tell whether it declares a workspace isolation section")
		}
		cfg := root + "/.flow/project.md"
		path := wiAbs(env, cfg)

		// THE ABSENT CASE IS AN LSTAT, not a test that follows symlinks. A
		// `.flow/project.md` pointing at a path that does not exist would
		// otherwise read as a project that declares nothing — a link anyone
		// able to edit the repository can create, resolving to the one
		// verdict this guard must never give away. The file is optional: a
		// project with no `.flow/project.md` is a supported, ordinary case.
		if _, err := os.Lstat(path); err != nil {
			fmt.Fprintf(stdout, "ISOLATION-OK: %s — no .flow/project.md, so nothing is declared\n", root)
			continue
		}
		// A REGULAR FILE, TESTED BEFORE READABILITY, because readability
		// ANSWERS ABOUT A DIRECTORY TOO: the bash's `grep` then failed with
		// "Is a directory" while its exit status was indistinguishable from
		// "no match", so `ln -s somedir .flow/project.md` produced
		// ISOLATION-OK and exit 0. isFile follows symlinks, so a configuration
		// that IS a symlink to a real file is still read; what it excludes is
		// a directory, a fifo (which a read would block on forever), a device
		// and a dangling link.
		if !isFile(path) {
			return refuse(cfg + " is not a regular file — cannot validate what it declares")
		}
		// A file that exists but cannot be read is NOT "declares no
		// isolation". Reading absence out of a path that was never readable
		// is the false pass this guard exists to prevent.
		if syscall.Access(path, 4) != nil {
			return refuse(cfg + " exists but is not readable — cannot validate what it declares")
		}
		// A COMMAND'S FAILURE IS NEVER READ AS ITS NEGATIVE ANSWER. The bash
		// read each grep's status as three outcomes — matched, did not match,
		// could not look; a read that fails here is the third, and refuses
		// with the message the bash gave it.
		body, err := os.ReadFile(path)
		if err != nil {
			return refuse("grep exited 2 while looking for the '## workspace isolation' heading in " + cfg + " — that is a failure to look rather than an absence, so nothing was validated")
		}
		// grep and awk alike ended a line at its first NUL.
		recs := lines(body)
		for i := range recs {
			recs[i], _, _ = strings.Cut(recs[i], "\x00")
		}

		// The heading rule, written once (ccHeading) and used by the presence
		// test, the duplicate count and the extraction alike. Two spellings of
		// it would be two answers to "is this project isolated?", and the
		// extraction's would win silently. It is check-cleanup-complete's
		// rule, so a section that guard reads is a section this one validates.
		count := 0
		for _, l := range recs {
			if ccHeading.MatchString(ccASCIILower(l)) {
				count++
			}
		}
		if count == 0 {
			fmt.Fprintf(stdout, "ISOLATION-OK: %s — declares no `## workspace isolation` section\n", cfg)
			continue
		}
		// TWO DECLARATIONS ARE NOT A DECLARATION. The file is hand-written, so
		// a heading duplicated by a bad merge or a copied block is a realistic
		// mistake, and the two sections can name two different commands
		// against two different services. check-cleanup-complete reports the
		// same shape as a SKIP, because its question is whether it may run one
		// of them; the question here is whether the declaration is well
		// formed, and an ambiguous one is not. The count is reported because
		// "you have two" is what the operator has to fix. Neither section is
		// validated: validating both would report each row twice, and
		// validating the first would be this guard picking the winner the
		// contract declines to pick.
		if count > 1 {
			fmt.Fprintf(stdout, "%s:0: the file declares %d `## workspace isolation` sections — a second declaration is an ambiguous declaration rather than a merge, so neither was validated and neither of their commands would be run\n", cfg, count)
			total++
			fmt.Fprintf(stdout, "ISOLATION-INVALID: %s — 1 violation(s) in the `## workspace isolation` section\n", cfg)
			continue
		}

		// The whole of the validation. wiValidate returns one violation line
		// per finding plus the row counts; the verdict is decided here, per
		// project, so a caller passing several project roots gets one verdict
		// per project. The bash ran its validator as an awk program and
		// refused when that printed no `#SUMMARY` sentinel (no awk on the
		// machine, a syntax error, a signal); this validator is in-process
		// code with no such way to stop short, and the shim is what refuses
		// when flow-guard itself cannot run.
		p := wiValidate(cfg, recs, wiSpace(utf8Locale), utf8Locale, printRows)
		if len(p.viol) > 0 {
			// Sanitized at the ONE point they reach stdout, rather than at each
			// of the twenty places a cell is interpolated into a message: a
			// chokepoint covers the next message somebody adds. The count is
			// taken from the unsanitized text because the escaping never adds
			// or removes a line.
			findings := strings.Join(p.viol, "\n")
			fmt.Fprint(stdout, ccSanitize(findings+"\n"))
			n := strings.Count(findings, "\n") + 1
			total += n
			fmt.Fprintf(stdout, "ISOLATION-INVALID: %s — %d violation(s) in the `## workspace isolation` section\n", cfg, n)
		} else {
			fmt.Fprintf(stdout, "ISOLATION-OK: %s — %d resource row(s) and %d command row(s) validated\n", cfg, p.nres, p.ncmd)
		}
		// `#ROW` lines, if any, always come last — after the verdict line, so
		// a caller not reading them sees only the verdict.
		for _, row := range p.rows {
			fmt.Fprintln(stdout, row)
		}
	}
	if total != 0 {
		return 1
	}
	return 0
}

// wiAbs is p as the bash's file tests resolved it: relative to the working
// directory. The empty path names nothing.
func wiAbs(env Env, p string) string {
	if p == "" || strings.HasPrefix(p, "/") {
		return p
	}
	return env.Dir + "/" + p
}

type wiRef struct {
	name, row, tok string
	line           int
}

// wiParse is the state of one validation: the port of the bash's awk
// program, whose only input was the file's text. THE INPUT IS
// ATTACKER-INFLUENCED — `.flow/project.md` is tracked in the repository and
// editable in any pull request — and NOTHING read here is executed: this
// guard never runs a project's `create`, `remove` or `survivors` command,
// never resolves a path out of the file, and never interpolates a cell into a
// shell.
type wiParse struct {
	cfg        string
	space      func(rune) bool // awk's [[:space:]] under the caller's locale
	utf8       bool
	printRows  bool
	viol, rows []string
	nres, ncmd int
	ntable     int
	resSeen    map[string]int
	varKind    map[string]string
	varLine    map[string]int
	cmdSeen    map[string]int
	refs       []wiRef
}

func wiValidate(cfg string, recs []string, space func(rune) bool, utf8Locale, printRows bool) *wiParse {
	p := &wiParse{cfg: cfg, space: space, utf8: utf8Locale, printRows: printRows,
		resSeen: map[string]int{}, varKind: map[string]string{}, varLine: map[string]int{}, cmdSeen: map[string]int{}}
	inSec, isoLine := false, 0
	var blk []string
	var blkNo []int
	flush := func() {
		if blk != nil {
			p.flushBlock(blk, blkNo)
			blk, blkNo = nil, nil
		}
	}
	for i, line := range recs {
		// A heading of any level ends the section, which is the same shape
		// the cleanup guard scans with: a project configuration is a
		// human-written Markdown file of sections, not a nested document.
		if ccAnyHeading.MatchString(line) {
			flush()
			inSec = ccHeading.MatchString(ccASCIILower(line))
			if inSec {
				isoLine = i + 1
			}
			continue
		}
		if !inSec {
			continue
		}
		// One contiguous run of table rows is one block.
		if strings.HasPrefix(strings.TrimLeftFunc(strings.ReplaceAll(line, "\r", ""), p.space), "|") {
			blk, blkNo = append(blk, line), append(blkNo, i+1)
			continue
		}
		flush()
	}
	flush()
	p.resolveRefs()
	// A HEADING WITH NO TABLE UNDER IT DECLARED NOTHING. That is how a
	// mistyped header actually presents — the tables are there and neither
	// was recognised — and it is also what a section holding only prose is.
	// Passing it would report a project isolated that declared nothing, which
	// is the one answer that must be reserved for a project carrying no
	// heading at all.
	if p.ntable == 0 && len(p.viol) == 0 {
		p.violation(isoLine, "the `## workspace isolation` section carries neither the resource table nor the command table — a heading with nothing under it declares nothing, and a project that is not isolated says so by carrying no heading")
	}
	return p
}

func (p *wiParse) violation(line int, msg string) {
	p.viol = append(p.viol, fmt.Sprintf("%s:%d: %s", p.cfg, line, msg))
}

// splitCells is the cell splitter: a character walk, not a field split,
// because a realistic cell contains a `|` written `\|`, and a split on `|`
// would cut the row there and leave a shorter row that still looks well
// formed. A row may omit its trailing `|`, a file may arrive with CRLF line
// endings, and a row may be indented. The walk is over bytes, as awk's
// length and substr counted them.
func (p *wiParse) splitCells(line string) []string {
	line = strings.TrimFunc(strings.ReplaceAll(line, "\r", ""), p.space)
	if !strings.HasPrefix(line, "|") {
		return nil
	}
	line = line[1:]
	var cells []string
	var cur strings.Builder
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if ch == '\\' && i < len(line)-1 {
			if nxt := line[i+1]; nxt == '|' || nxt == '\\' {
				cur.WriteByte(nxt)
				i++
				continue
			}
			cur.WriteByte(ch)
			continue
		}
		if ch == '|' {
			cells = append(cells, cur.String())
			cur.Reset()
			continue
		}
		cur.WriteByte(ch)
	}
	cells = append(cells, cur.String())
	// A trailing `|` leaves one empty trailing field; a row that omits it
	// leaves a real one. Dropping only the empty one accepts both shapes.
	if len(cells) > 1 && cur.Len() == 0 {
		cells = cells[:len(cells)-1]
	}
	return cells
}

// trim is the cell as the contract reads it: without its surrounding
// whitespace and without the backticks a project writes around a literal.
func (p *wiParse) trim(c string) string {
	return strings.TrimFunc(c, func(r rune) bool { return r == '`' || p.space(r) })
}

// fold is a header cell (or a Resource word, or a verb), folded for
// comparison: lowercased, with internal runs of whitespace collapsed, so
// `In  a workspace` and `In a workspace` are one heading and `In a
// workspace` and `Workspace` are two. awk's tolower folded every letter
// under a UTF-8 locale and ASCII alone otherwise.
func (p *wiParse) fold(c string) string {
	c = p.trim(c)
	var b strings.Builder
	inSpace := false
	for i := 0; i < len(c); {
		r, n := utf8.DecodeRuneInString(c[i:])
		switch {
		case r == utf8.RuneError && n <= 1:
			b.WriteByte(c[i])
			inSpace = false
		case p.space(r):
			if !inSpace {
				b.WriteByte(' ')
			}
			inSpace = true
		default:
			if p.utf8 || r < utf8.RuneSelf {
				r = unicode.ToLower(r)
			}
			b.WriteRune(r)
			inSpace = false
		}
		i += n
	}
	return b.String()
}

func (p *wiParse) header(cells []string) string {
	folded := make([]string, len(cells))
	for i, c := range cells {
		folded[i] = p.fold(c)
	}
	return strings.Join(folded, "|")
}

// isDelimiter is a markdown delimiter row (`|----|----|`), which carries no
// data.
func (p *wiParse) isDelimiter(cells []string) bool {
	for _, c := range cells {
		if !wiDelim.MatchString(p.trim(c)) {
			return false
		}
	}
	return len(cells) > 0
}

// flushBlock reads one contiguous run of table rows. Its first row is the
// header, and the header is what says which of the two tables this is.
func (p *wiParse) flushBlock(blk []string, blkNo []int) {
	hdr := p.header(p.splitCells(blk[0]))
	kind, want := "", 0
	switch hdr {
	case wiResourceHeader:
		kind, want = "resource", 4
	case wiCommandHeader:
		kind, want = "command", 2
	}
	// A HEADER THIS GUARD DOES NOT RECOGNISE IS A VIOLATION, never a table
	// it skips. Skipping is the fail-open this guard exists to prevent: a
	// reordered header leaves every cell present and every one meaning
	// something else, and a section whose only table was skipped would
	// declare nothing and pass. The two column lists are stated in the
	// message because what the author has to see is the order they were
	// meant to be in.
	if kind == "" {
		p.violation(blkNo[0], "a table under `## workspace isolation` whose header is `"+hdr+"` — the section holds one resource table (`Resource | Variable | Default | In a workspace`, in that order) and one command table (`Command | Runs`), and nothing else in it is read")
		return
	}
	p.ntable++
	for i := 1; i < len(blk); i++ {
		cells := p.splitCells(blk[i])
		if p.isDelimiter(cells) {
			continue
		}
		// A SECOND HEADER INSIDE ONE BLOCK IS TWO TABLES THAT MERGED.
		// Markdown separates tables with a blank line, and two written back
		// to back are one contiguous run of rows to the scanner — so this
		// parser read the FIRST header and would go on to evaluate the second
		// table rows against the first table columns. It did exactly that
		// before this branch existed: a resource table with a command table
		// under it emitted four "cell count mismatch" violations, one per
		// command row, and named nothing that would lead the author to the
		// missing blank line.
		//
		// The block is ABANDONED here rather than split into two. Splitting
		// would be this guard repairing input the contract says to report,
		// and every row below this line belongs to a table whose header this
		// parser never started reading.
		if other := p.header(cells); other == wiResourceHeader || other == wiCommandHeader {
			p.violation(blkNo[i], "a second table header (`"+other+"`) begins here, inside the table that started at line "+strconv.Itoa(blkNo[0])+" — two tables written with no blank line or prose between them are one block to this parser, so it read only the first header and stopped here rather than reading the rows below against the wrong columns; separate the two tables with a blank line")
			return
		}
		// A ROW THAT DOES NOT LINE UP WITH ITS HEADER CANNOT BE READ AGAINST
		// IT. Reading a short row would report the wrong rule broken — a row
		// missing its `Default` presents as a row whose `In a workspace` cell
		// is empty, and the operator would be sent to fill in a cell that is
		// missing rather than blank.
		if len(cells) != want {
			p.violation(blkNo[i], fmt.Sprintf("a %s row with %d cell(s) where its header has %d (expected %d) — the row does not line up with its header, so it is dropped rather than read against the wrong columns", kind, len(cells), want, want))
			continue
		}
		if kind == "resource" {
			p.nres++
			p.resourceRow(blkNo[i], cells)
		} else {
			p.ncmd++
			p.commandRow(blkNo[i], cells)
		}
	}
}

// resolveRefs resolves every `<value:...>` reference once the whole table has
// been read. Resolving as each row was read would make a reference legal or
// illegal according to the order the author happened to write the rows in,
// and one level of reference resolves in a single pass precisely so there is
// no resolution order to declare rows in. These findings are therefore
// reported AFTER the row findings rather than in line order; each line names
// its own file and line number.
func (p *wiParse) resolveRefs() {
	for _, r := range p.refs {
		kind, ok := p.varKind[r.name]
		switch {
		case !ok:
			p.violation(r.line, "row `"+r.row+"`: "+r.tok+" names no row in this table — a reference names the row whose Variable column holds that name, so the row is dropped")
		case kind == "url":
			p.violation(r.line, "row `"+r.row+"`: "+r.tok+" names the `url` row at line "+strconv.Itoa(p.varLine[r.name])+", and a reference may name a `database`, `bucket` or `port` row and never another `url` row — the row is dropped")
		case kind == "cache index":
			p.violation(r.line, "row `"+r.row+"`: "+r.tok+" names the `cache index` row at line "+strconv.Itoa(p.varLine[r.name])+", whose value is claimed at run time rather than derived from the id and so cannot be referenced — the row is dropped")
		}
	}
}

// spacedTokens reports a near miss of a token. A TOKEN WITH A SPACE IN IT IS
// INVISIBLE TO ccUnknownToken, AND THAT IS THE LIKELIEST TYPO AT THIS
// KEYBOARD: `<value: API_PORT>` matches the token shape nowhere, so every rule
// keyed on tokens passes and the cell is read as literal text — a row reported
// valid whose value would reach a run with the brackets still in it.
//
// THE SHAPE CANNOT SIMPLY BE WIDENED TO ADMIT WHITESPACE: `cmd < in.txt >
// out.txt` is an ordinary redirection whose bracketed run holds only token
// characters and a space, and a guard that cried wolf on real commands would
// be worked around rather than fixed. SO THE TEST IS "IS THIS A NEAR MISS OF
// A TOKEN THIS CONTRACT DEFINES": the vocabulary is closed and tiny, so a
// bracketed run whose whitespace-squeezed content IS one of those names was
// meant to be that token. `< in.txt >` squeezes to `in.txt` and is left
// alone. WHAT IT DELIBERATELY DOES NOT CATCH is a bracketed run that is
// neither — `<my thing>` is read as literal text.
func (p *wiParse) spacedTokens(line int, who, cell, ctx string) {
	for _, run := range wiBracket.FindAllString(cell, -1) {
		body := run[1 : len(run)-1]
		if !strings.ContainsFunc(body, p.space) {
			continue
		}
		body = strings.Map(func(r rune) rune {
			if p.space(r) {
				return -1
			}
			return r
		}, body)
		if body == "id" || body == "id_underscored" || body == "offset" || wiValueRef.MatchString(body) {
			p.violation(line, ctx+" `"+who+"`: `"+run+"` carries whitespace inside its brackets, so it is not the token `<"+body+">` and nothing substitutes it — a bracketed run with a space in it is read as literal text, and the row is dropped")
		}
	}
}

// resourceRow checks one row of the resource table.
func (p *wiParse) resourceRow(line int, cells []string) {
	// Emitted BEFORE any of the checks below run, and for every row already
	// known to line up with the resource header. A row this function goes on
	// to reject still gets its `#ROW` line: prepare-workspace.sh never reads
	// it, because a rejected row means this guard exits non-zero and
	// prepare-workspace.sh exits on that before it looks at a `#ROW` line.
	if p.printRows {
		p.rows = append(p.rows, "#ROW\t"+p.trim(cells[0])+"\t"+p.trim(cells[1])+"\t"+p.trim(cells[2])+"\t"+p.trim(cells[3]))
	}
	// The row named the way the operator has to find it: by the variable it
	// carries, as the contract asks.
	who := p.trim(cells[1])
	if who == "" {
		who = "(no Variable cell)"
	}

	// The `Resource` word, folded for case and internal whitespace and for
	// nothing else: `cache index` is two words whose space is part of the
	// word, so a fold that REMOVED whitespace would accept `cacheindex`.
	res := p.fold(cells[0])
	if res != "database" && res != "bucket" && res != "cache index" && res != "port" && res != "url" {
		p.violation(line, "row `"+who+"`: Resource `"+p.trim(cells[0])+"` is not one of `database`, `bucket`, `cache index`, `port` or `url` — the vocabulary is closed, so the row is dropped and nothing removes it")
		return
	}

	// The `Variable`, which a run EXPORTS. The ranges are written out
	// because a character class is exactly the construct that starts meaning
	// something else in another locale.
	v := p.trim(cells[1])
	if v == "" {
		p.violation(line, "resource row: the Variable cell is empty — a row with no variable carries nothing, so the row is dropped")
		return
	}
	if !wiVariable.MatchString(v) {
		p.violation(line, "row `"+who+"`: Variable `"+v+"` is not a legal environment-variable name (`[A-Za-z_][A-Za-z0-9_]*`) — a run exports these, so the row is dropped")
		return
	}

	// `port` and `url` repeat by design; the other three name one thing
	// each, and a second `database` row is two databases one workspace id
	// cannot name.
	if res == "database" || res == "bucket" || res == "cache index" {
		if first, ok := p.resSeen[res]; ok {
			p.violation(line, "row `"+who+"`: a second `"+res+"` row, the first being at line "+strconv.Itoa(first)+" — `database`, `bucket` and `cache index` each appear at most once, so the row is dropped")
		} else {
			p.resSeen[res] = line
		}
	}

	// The variable index a `<value:...>` reference resolves against. A
	// reference names THE row whose Variable column holds that name —
	// singular — so two rows holding one name resolve to no single row.
	if first, ok := p.varLine[v]; ok {
		p.violation(line, "row `"+who+"`: a second resource row holds the Variable `"+v+"`, the first being at line "+strconv.Itoa(first)+" — a `<value:"+v+">` reference names one row and cannot name two, and a run would export one value over the other")
	} else {
		p.varKind[v], p.varLine[v] = res, line
	}

	// The bare-integer `Default` binds `port` and `cache index` and no other
	// resource: a `database` default is a connection string and a `url`
	// default is a URL. A signed default is not a port number.
	if def := p.trim(cells[2]); (res == "port" || res == "cache index") && !wiInteger.MatchString(def) {
		p.violation(line, "row `"+who+"`: Default `"+def+"` is not a bare integer, which a `"+res+"` row requires — the workspace value is arithmetic on this number, so the row is dropped")
		return
	}

	// The `In a workspace` cell, whose form the ROW Resource selects. The
	// check is on the FORM and not only on the tokens, which is what catches
	// the cell that carries none: a `port` row whose cell reads `9090` names
	// no token at all.
	ws := p.trim(cells[3])
	switch res {
	case "port":
		if ws != "+<offset>" {
			p.violation(line, "row `"+who+"`: the In a workspace cell is `"+ws+"`, and a `port` row writes the literal `+<offset>` with nothing before or after it — the row is dropped")
		}
		return
	case "cache index":
		if ws != "probed" {
			p.violation(line, "row `"+who+"`: the In a workspace cell is `"+ws+"`, and a `cache index` row writes the literal `probed` alone, because the index is claimed at run time rather than derived — the row is dropped")
		}
		return
	}

	// `database`, `bucket` and `url` all write the value out in full, so
	// what separates them is which tokens they may carry: `<value:...>` is
	// legal in a `url` cell and nowhere else, because a `database` or
	// `bucket` cell is what the removal commands target.
	if ws == "" {
		p.violation(line, "row `"+who+"`: the In a workspace cell is empty, and a `"+res+"` row writes its value out in full — the row is dropped")
		return
	}
	p.spacedTokens(line, who, cells[3], "row")
	for _, tok := range ccUnknownToken.FindAllString(cells[3], -1) {
		if tok == "<id>" || tok == "<id_underscored>" {
			continue
		}
		if strings.HasPrefix(tok, "<value:") {
			if res != "url" {
				p.violation(line, "row `"+who+"`: the In a workspace cell names "+tok+", and `<value:...>` is legal in a `url` cell and nowhere else — the row is dropped")
				continue
			}
			p.refs = append(p.refs, wiRef{name: tok[len("<value:") : len(tok)-1], row: who, tok: tok, line: line})
			continue
		}
		urlToken := ""
		if res == "url" {
			urlToken = ", `<value:VARIABLE>`"
		}
		p.violation(line, "row `"+who+"`: the In a workspace cell names "+tok+", which is not a token this contract substitutes (`<id>`, `<id_underscored>`"+urlToken+") — the row is dropped")
	}
}

// commandRow checks one row of the command table. A command row is
// EXECUTED, at the same trust level as the `## run`, `## test` and `## lint`
// commands, which carry no containment rule either — so the one thing
// checked in its text is its tokens. No path is isolated inside a command
// and none is looked for: a parser that is wrong about one shell command
// reports the wrong thing about all of them.
func (p *wiParse) commandRow(line int, cells []string) {
	verb := p.fold(cells[0])
	if verb != "create" && verb != "remove" && verb != "survivors" {
		p.violation(line, "command row `"+p.trim(cells[0])+"`: the command table has three verbs — `create`, `remove` and `survivors` — and nothing calls anything else, so the row is dropped and never runs")
		return
	}
	if first, ok := p.cmdSeen[verb]; ok {
		p.violation(line, "command row `"+verb+"`: a second `"+verb+"` row, the first being at line "+strconv.Itoa(first)+" — two commands under one verb resolve to whichever is read first, so the row is dropped")
		return
	}
	p.cmdSeen[verb] = line
	if p.trim(cells[1]) == "" {
		p.violation(line, "command row `"+verb+"`: the Runs cell is empty — a verb declared with no command is a step that would run as the empty string, so the row is dropped")
		return
	}
	// `<id>` and `<id_underscored>` are substituted and nothing else is.
	// `<value:...>` is deliberately among the things that are not: a command
	// needing a derived value reads the exported variable.
	p.spacedTokens(line, verb, cells[1], "command row")
	for _, tok := range ccUnknownToken.FindAllString(cells[1], -1) {
		if tok != "<id>" && tok != "<id_underscored>" {
			p.violation(line, "command row `"+verb+"`: the command names "+tok+", and only `<id>` and `<id_underscored>` are substituted in a command — the pipeline would hand the shell a literal nobody intended, so the row is dropped")
		}
	}
}

// wiSpace is awk's and sed's [[:space:]]: the ASCII whitespace but newline,
// plus U+00A0 under a UTF-8 locale (measured on macOS; U+0085, U+2000–U+200A
// and U+3000 are not matched).
func wiSpace(utf8 bool) func(rune) bool {
	return func(r rune) bool { return strings.ContainsRune(gdcSpace, r) || (utf8 && r == ' ') }
}
