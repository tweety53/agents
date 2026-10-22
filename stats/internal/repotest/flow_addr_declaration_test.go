package repotest

// TestFlowAddrDeclaration pins the flow CLI's store-address resolution against
// the project's own declaration of it (.flow/project.md's `## workspace
// isolation` section). KAN-616: which commands resolve FLOW_RECORDS_ADDR was
// one fact stated four ways, none matching the code, until
// registerRecordConnFlags became the one seam — this test is what keeps the
// declared set and the wired set from drifting apart again.
//
// WHAT IT ASSERTS, against a root directory (the repository root, or a
// sandbox copy for the mutation cases):
//
//  1. The record-family sentence under `## workspace isolation` parses to a
//     non-empty set of backticked command names, and the isolation table
//     carries the `FLOW_ADDR` and `FLOW_RECORDS_ADDR` rows the sentence
//     speaks of.
//  2. Exactly one `-addr` registration under stats/cmd/flow resolves
//     FLOW_RECORDS_ADDR, and it is the one inside registerRecordConnFlags.
//  3. The files holding registerRecordConnFlags call sites are exactly the
//     files the declared commands map to — the CLI's own dispatch
//     convention: the command's subcommand word, hyphens dropped, is the
//     source file's stem (`flow record` -> record.go, `flow self-review
//     bundle` -> selfreview.go). Equality checks both directions at once:
//     no call site outside the declared family, no declared file without
//     one.
//  4. No registerConnFlags call site — the seam whose address follows
//     FLOW_ADDR alone — sits in a declared record-family file, and no
//     inline `-addr` registration through resolveDefaultAddr() sits in one
//     either, outside the registerConnFlags definition itself: a family
//     verb registering through its own helper (the registerJiraConnFlags
//     shape) resolves FLOW_ADDR exactly where the declaration says the
//     whole family resolves FLOW_RECORDS_ADDR.
//  5. Every `-addr` registration under stats/cmd/flow takes its default
//     from resolveDefaultAddr() or resolveRecordsAddr(), never a literal
//     and never a third resolver — so every store-touching command honors
//     FLOW_ADDR, and only the record family consults FLOW_RECORDS_ADDR.
//
// THE SEVEN CASES run the assertion against the real root and the unmutated
// copy (both must pass) and against one drift mutation each (every one must
// fail). The mutation cases make the test self-pinning: an assertion loosened
// into vacuity stops failing the mutations. Sources are read only; the real
// tree is never written.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// readLines returns a file's lines; a missing file yields none (grep's
// 2>/dev/null in the original).
func readLines(path string) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
}

// countIn counts lines across files containing sub.
func countIn(files []string, sub string) int {
	n := 0
	for _, f := range files {
		for _, l := range readLines(f) {
			if strings.Contains(l, sub) {
				n++
			}
		}
	}
	return n
}

// funcBody returns record.go's lines from `func <name>(` through the next
// line starting `func ` (sed's /start/,/end/p range).
func funcBody(path, name string) []string {
	var out []string
	in := false
	for _, l := range readLines(path) {
		if !in {
			if strings.HasPrefix(l, "func "+name+"(") {
				in = true
				out = append(out, l)
			}
			continue
		}
		out = append(out, l)
		if strings.HasPrefix(l, "func ") {
			break
		}
	}
	return out
}

func countLines(lines []string, sub string) int {
	n := 0
	for _, l := range lines {
		if strings.Contains(l, sub) {
			n++
		}
	}
	return n
}

var (
	recordSentence = regexp.MustCompile(`The record family \((.*)\) resolves`)
	backticked     = regexp.MustCompile("`[^`]*`")
	addrReg        = regexp.MustCompile(`\.String(Var)?\(.*"addr", `)
)

// assertFlowAddrRoot runs every declaration assertion against one root and
// returns the first violated assertion's description, or "" when all hold.
func assertFlowAddrRoot(root string) string {
	md := filepath.Join(root, ".flow", "project.md")
	cmd := filepath.Join(root, "stats", "cmd", "flow")
	if _, err := os.Stat(md); err != nil {
		return "no .flow/project.md under " + root
	}
	if fi, err := os.Stat(cmd); err != nil || !fi.IsDir() {
		return "no stats/cmd/flow under " + root
	}

	// 1. the declaration parses, and the rows it speaks of are in the table.
	// The anchor is the sentence's opening through its verb on one physical
	// line: a reflow that moves the paren group or the verb off that line
	// fails here — the safe direction.
	mdLines := readLines(md)
	var sentences []string
	for _, l := range mdLines {
		if m := recordSentence.FindStringSubmatch(l); m != nil {
			sentences = append(sentences, m[1])
		}
	}
	if len(sentences) != 1 {
		return fmt.Sprintf("expected exactly one record-family sentence in .flow/project.md, found %d", len(sentences))
	}
	var declared []string
	for _, b := range backticked.FindAllString(sentences[0], -1) {
		declared = append(declared, strings.Trim(b, "`"))
	}
	if len(declared) == 0 {
		return "the record-family sentence names no backticked commands"
	}
	if countLines(mdLines, "| `FLOW_ADDR` |") == 0 {
		return "the isolation table has no FLOW_ADDR row"
	}
	if countLines(mdLines, "| `FLOW_RECORDS_ADDR` |") == 0 {
		return "the isolation table has no FLOW_RECORDS_ADDR row"
	}

	// the file each declared command maps to, per the dispatch convention
	var declaredFiles []string
	for _, c := range declared {
		if f := strings.Fields(c); len(f) > 1 {
			declaredFiles = append(declaredFiles, strings.ReplaceAll(f[1], "-", "")+".go")
		}
	}
	slices.Sort(declaredFiles)
	declaredFiles = slices.Compact(declaredFiles)

	var srcs []string
	_ = filepath.WalkDir(cmd, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
			srcs = append(srcs, p)
		}
		return nil
	})
	recordGo := filepath.Join(cmd, "record.go")

	// 2. exactly one FLOW_RECORDS_ADDR registration, inside registerRecordConnFlags
	const recReg, defReg = `"addr", resolveRecordsAddr()`, `"addr", resolveDefaultAddr()`
	if n := countIn(srcs, recReg); n != 1 {
		return fmt.Sprintf("expected exactly one FLOW_RECORDS_ADDR -addr registration under stats/cmd/flow, found %d", n)
	}
	if countLines(funcBody(recordGo, "registerRecordConnFlags"), recReg) != 1 {
		return "the FLOW_RECORDS_ADDR -addr registration is not the one inside registerRecordConnFlags"
	}

	// 3. record-seam call sites live in exactly the declared commands' files
	var siteFiles []string
	for _, f := range srcs {
		for _, l := range readLines(f) {
			if strings.HasPrefix(l, "\tregisterRecordConnFlags(") {
				siteFiles = append(siteFiles, filepath.Base(f))
				break
			}
		}
	}
	slices.Sort(siteFiles)
	siteFiles = slices.Compact(siteFiles)
	if !slices.Equal(siteFiles, declaredFiles) {
		return fmt.Sprintf("registerRecordConnFlags call-site files (%s) are not exactly the declared record family's files (%s)",
			strings.Join(siteFiles, " "), strings.Join(declaredFiles, " "))
	}

	// 4. the FLOW_ADDR-only seam is never called from inside the record
	// family, and the family's files never carry an inline FLOW_ADDR-only
	// -addr registration either, outside registerConnFlags's own definition
	// (it lives in record.go and legitimately registers through the
	// resolver). Only the flag name plus resolver default identifies a
	// registration, whatever variable it registers into.
	var famFiles, leaked []string
	for _, f := range declaredFiles {
		p := filepath.Join(cmd, f)
		famFiles = append(famFiles, p)
		for _, l := range readLines(p) {
			if strings.HasPrefix(l, "\tregisterConnFlags(") {
				leaked = append(leaked, p)
				break
			}
		}
	}
	if len(leaked) > 0 {
		return "record-family files call the FLOW_ADDR-only seam: " + strings.Join(leaked, " ")
	}
	inlineTotal := countIn(famFiles, defReg)
	inSeam := countLines(funcBody(recordGo, "registerConnFlags"), defReg)
	if inlineTotal != inSeam {
		return fmt.Sprintf("declared record-family files carry %d inline -addr registration(s) through resolveDefaultAddr() outside the registerConnFlags definition itself", inlineTotal-inSeam)
	}

	// 5. every -addr registration takes a resolver default, never a literal
	regs := 0
	for _, f := range srcs {
		for _, l := range readLines(f) {
			if addrReg.MatchString(l) {
				regs++
			}
		}
	}
	dres, rres := countIn(srcs, defReg), countIn(srcs, recReg)
	if regs != dres+rres || regs != dres+1 {
		return fmt.Sprintf("%d -addr registrations but only %d take a resolver default — a literal or third resolver is wired somewhere", regs, dres+rres)
	}
	return ""
}

// flowAddrRoot copies the two inputs the assertion reads from the real root
// into a sandbox, applying sed-style first-per-line replacement old->new to
// the one file rel names (rel empty: no mutation).
func flowAddrRoot(t *testing.T, rel, old, repl string) string {
	t.Helper()
	realRoot, dest := repoRoot(t), t.TempDir()
	copyFile := func(src, dst string) {
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	copyFile(filepath.Join(realRoot, ".flow", "project.md"), filepath.Join(dest, ".flow", "project.md"))
	srcs, err := filepath.Glob(filepath.Join(realRoot, "stats", "cmd", "flow", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range srcs {
		if !strings.HasSuffix(s, "_test.go") {
			copyFile(s, filepath.Join(dest, "stats", "cmd", "flow", filepath.Base(s)))
		}
	}
	if rel != "" {
		p := filepath.Join(dest, rel)
		lines := readLines(p)
		if lines == nil {
			t.Fatalf("mutation target %s missing", rel)
		}
		changed := false
		for i, l := range lines {
			if strings.Contains(l, old) {
				lines[i] = strings.Replace(l, old, repl, 1)
				changed = true
			}
		}
		if !changed {
			t.Fatalf("mutation %q matched nothing in %s", old, rel)
		}
		if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dest
}

func TestFlowAddrDeclaration(t *testing.T) {
	t.Parallel()

	t.Run("pass-unmutated", func(t *testing.T) {
		t.Parallel()
		for _, root := range []string{repoRoot(t), flowAddrRoot(t, "", "", "")} {
			if r := assertFlowAddrRoot(root); r != "" {
				t.Fatalf("unmutated tree asserts dirty — %s", r)
			}
		}
	})

	cases := []struct{ name, file, old, repl string }{
		// a verb outside the record family joins the record seam
		{"mutate-record-seam-outside-family", "stats/cmd/flow/hazard.go", "registerConnFlags(", "registerRecordConnFlags("},
		// registerRecordConnFlags stops resolving FLOW_RECORDS_ADDR
		{"mutate-record-seam-default", "stats/cmd/flow/record.go", `"addr", resolveRecordsAddr()`, `"addr", resolveDefaultAddr()`},
		// a record-family verb falls back to the FLOW_ADDR-only seam
		{"mutate-family-verb-switches-seam", "stats/cmd/flow/selfreview.go", "registerRecordConnFlags(fset, &f)", "registerConnFlags(fset, &f)"},
		// the declared set loses every `flow self-review` verb (they share
		// one source file); the closing paren stays, so the case fails
		// through assertion 3's set comparison, not the sentence parse
		{"mutate-declaration-drops-verb", ".flow/project.md", ", `flow self-review finding`, `flow self-review findings`, `flow self-review bundle`", ""},
		// `flow suite` (declared so its runtime rows outlive a worktree)
		// falls back to the FLOW_ADDR-only seam
		{"mutate-family-suite-switches-seam", "stats/cmd/flow/suite.go", "registerRecordConnFlags(fset, &f)", "registerConnFlags(fset, &f)"},
		// a registration takes a hardcoded default
		{"mutate-literal-addr-registration", "stats/cmd/flow/jira.go", `"addr", resolveDefaultAddr()`, `"addr", "http://127.0.0.1:9999"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if r := assertFlowAddrRoot(flowAddrRoot(t, c.file, c.old, c.repl)); r == "" {
				t.Fatal("expected fail, got pass")
			}
		})
	}
}
