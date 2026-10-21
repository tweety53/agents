package guard

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

// checkMutationReproducerPin is scripts/check-mutation-reproducer-pin.sh's
// guard. The shim's header states the contract — the two canonical
// declaration lines, the presence and near-miss checks, and the exit codes;
// this file is the mechanism.
func init() {
	Registry["check-mutation-reproducer-pin"] = checkMutationReproducerPin
}

const mrpName = "check-mutation-reproducer-pin"

var (
	mrpMarkerDecl  = regexp.MustCompile(`(^|[[:space:]])rrMarker[[:space:]]*=`)
	mrpWindowDecl  = regexp.MustCompile(`range headLines\(path, `)
	mrpWindowTok   = regexp.MustCompile(`headLines\(path, ([0-9]+)\)`)
	mrpMarkerTok   = regexp.MustCompile(`= "([^"\\]*)"[[:space:]]*$`)
	mrpMarkerSpan  = regexp.MustCompile("`# mutation-[a-z0-9][a-z0-9-]*`")
	mrpWindowStmt  = regexp.MustCompile(`first [0-9]+ lines`)
	mrpMarkerAnchr = regexp.MustCompile(`# mutation-[a-z0-9][a-z0-9-]*`)
)

// mrpExit2 is die2: a "cannot answer at all" refusal in the same file:line
// shape as every finding, on stderr, exit 2.
type mrpExit2 struct{ path, msg string }

func checkMutationReproducerPin(_ []string, env Env, stdout, stderr io.Writer) int {
	// The root is this repository's own root by default -- the parent of
	// scripts/, derived as the bash guard derived it from its own location;
	// the binary lives in a cache, so the shim exports it as
	// FLOW_GUARD_REPO_ROOT. CHECK_MUTATION_REPRODUCER_PIN_ROOT is an explicit,
	// opt-in override honoured only when set, so the tests can point this
	// guard at a sandboxed fixture tree -- never set it for a normal
	// invocation. Set but empty is a refusal, not the default.
	root, set := env.LookupEnv("CHECK_MUTATION_REPRODUCER_PIN_ROOT")
	switch {
	case set && root == "":
		fmt.Fprintf(stderr, "%s: CHECK_MUTATION_REPRODUCER_PIN_ROOT is set but empty\n", mrpName)
		return 2
	case !set:
		if root = env.Getenv("FLOW_GUARD_REPO_ROOT"); root == "" {
			fmt.Fprintf(stderr, "%s: FLOW_GUARD_REPO_ROOT is unset — run scripts/check-mutation-reproducer-pin.sh, which sets it\n", mrpName)
			return 2
		}
	}
	violations, marker, window, refusal := mrpScan(env, root)
	if refusal != nil {
		fmt.Fprintf(stderr, "%s:0: %s\n", refusal.path, refusal.msg)
		return 2
	}
	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintln(stdout, v)
		}
		fmt.Fprintf(stdout, "MUTATION-REPRODUCER-PIN-INVALID: %s — %d violation(s)\n", root, len(violations))
		return 1
	}
	fmt.Fprintf(stdout, "MUTATION-REPRODUCER-PIN-OK: %s — marker '%s' window %s pinned across stats/internal/guard/runreproducer.go, scripts/run-reproducer.sh and skills/flow/review-panel.md\n", root, marker, window)
	return 0
}

// mrpScan runs every check; paths print as the bash built them
// ("$ROOT/<site>"), only file system calls see them resolved against the
// working directory.
func mrpScan(env Env, root string) (violations []string, marker, window string, refusal *mrpExit2) {
	if !isDir(smcAbs(env, root)) {
		return nil, "", "", &mrpExit2{root, "is not a directory — cannot scan"}
	}
	runner := root + "/scripts/run-reproducer.sh"
	panel := root + "/skills/flow/review-panel.md"
	source := root + "/stats/internal/guard/runreproducer.go"

	text := map[string]string{}
	for _, pinned := range []string{source, runner, panel} {
		p := smcAbs(env, pinned)
		if lp, err := os.Lstat(p); err == nil && lp.Mode()&fs.ModeSymlink != 0 {
			return nil, "", "", &mrpExit2{pinned, "is a symlink — a pinned site must be a real file, never a symlink"}
		}
		info, err := os.Stat(p)
		if err != nil {
			return nil, "", "", &mrpExit2{pinned, "does not exist — this is a pinned mutation-reproducer site"}
		}
		if !info.Mode().IsRegular() {
			return nil, "", "", &mrpExit2{pinned, "exists but is not a regular file — cannot scan"}
		}
		b, err := os.ReadFile(p)
		if err != nil || syscall.Access(p, 4) != nil {
			return nil, "", "", &mrpExit2{pinned, "is not readable — cannot scan"}
		}
		text[pinned] = string(b)
	}

	// The one line of source carrying half of the canonical declaration:
	// exactly one is answerable. Zero means the runner stopped declaring its
	// convention in this shape (moved, renamed or rewritten); more than one
	// means the reading is ambiguous. Either way this guard refuses to guess.
	declLine := func(re *regexp.Regexp, what string) (string, *mrpExit2) {
		var hits []string
		for _, l := range mrpLines(text[source]) {
			if re.MatchString(l) {
				hits = append(hits, l)
			}
		}
		switch len(hits) {
		case 0:
			return "", &mrpExit2{source, "the canonical declaration line (" + what + ") is absent — the pin cannot answer until it is re-pointed at the declaration's new shape"}
		case 1:
			return hits[0], nil
		}
		return "", &mrpExit2{source, fmt.Sprintf("carries %d lines matching the canonical declaration shape (%s) — the reading is ambiguous; keep exactly one", len(hits), what)}
	}
	markerDecl, refusal := declLine(mrpMarkerDecl, `rrMarker = "<marker>"`)
	if refusal != nil {
		return nil, "", "", refusal
	}
	windowDecl, refusal := declLine(mrpWindowDecl, "range headLines(path, <N>)")
	if refusal != nil {
		return nil, "", "", refusal
	}
	// No match is the pin's re-point signal: exit 2, never a guess. An
	// unextractable marker or window (the constant re-quoted as a raw
	// string, or the window made a named constant) must never be misreported
	// as drift with the whole line pasted in as the value (KAN-624 F2). The
	// bash took every match of grep -o; a declaration line carries one.
	windowTok := mrpWindowTok.FindStringSubmatch(windowDecl)
	markerTok := mrpMarkerTok.FindStringSubmatch(markerDecl)
	if windowTok == nil || markerTok == nil {
		return nil, "", "", &mrpExit2{source, "the canonical declaration lines matched but the marker or window could not be extracted from them — their shape moved; re-point the pin: " + markerDecl + " / " + windowDecl}
	}
	window, marker = windowTok[1], markerTok[1]

	// A pinned statement must be present verbatim; its absence is the drift
	// this guard exists to catch, reported at the first line carrying any
	// statement of the needle's own fact, so a drift row names an editable
	// line even when the needle itself has already drifted away (KAN-624
	// F4) — 0 when the file carries no such statement at all.
	contains := func(file, needle, msg string, anchor *regexp.Regexp) {
		if strings.Contains(text[file], needle) {
			return
		}
		line := 0
		for i, l := range mrpLines(text[file]) {
			if anchor.MatchString(l) {
				line = i + 1
				break
			}
		}
		violations = append(violations, fmt.Sprintf("%s:%d: %s", file, line, msg))
	}
	drift := "it drifted from stats/internal/guard/runreproducer.go's canonical declaration"
	markerMsg := fmt.Sprintf("no longer states the marker literal backticked — %s (marker: '%s'); move every site together", drift, marker)
	windowMsg := fmt.Sprintf(`no longer states the window as "first %s lines" — %s (window: %s); move every site together`, window, drift, window)
	contains(panel, "`"+marker+"`", markerMsg, mrpMarkerAnchr)
	contains(panel, "first "+window+" lines", windowMsg, mrpWindowStmt)
	contains(runner, "`"+marker+"`", "its header comment "+markerMsg, mrpMarkerAnchr)
	contains(runner, "first "+window+" lines", "its header comment "+windowMsg, mrpWindowStmt)

	for _, pinned := range []string{source, runner, panel} {
		violations = append(violations, mrpNearMiss(pinned, text[pinned], window, marker)...)
	}
	return violations, marker, window, nil
}

// mrpNearMiss is the near-miss absence scan — the presence checks cannot
// see a single site drifting while another copy of the same fact still
// reads correctly, which is the one-edit-at-a-time drift this guard exists
// to catch (KAN-624 F1: the real runner states its window more than once,
// the panel carries the marker three times, and the brief's window
// statement is wrapped across two lines where a single-line needle never
// looks). Every numeric window statement and every backticked
// mutation-marker span must carry exactly the canonical value — on a single
// line or wrapped across two adjacent lines — and each offending occurrence
// is reported at its own line (KAN-624 F4). Rows are sorted bytewise and
// de-duplicated, as the bash's `sort -u` over these ASCII rows did.
func mrpNearMiss(file, text, window, marker string) []string {
	var rows []string
	win, _ := strconv.Atoi(window)
	// b is the join boundary: 0 on a single-line scan, else the 1-based
	// index of the inserted space in "prev line current line". A joined scan
	// reports only a match that SPANS the boundary — one found wholly inside
	// either half belongs to that half's own single-line scan, and reporting
	// it here would duplicate it at the previous line's number, a phantom row
	// naming a line that carries no statement (KAN-624 F5). rstart is awk's
	// RSTART: 1-based within what is left of the line after the previous
	// match, exactly as the bash compared it with b.
	scan := func(s string, ln, b int) {
		cross := func(loc []int) bool {
			rstart, rlength := loc[0]+1, loc[1]-loc[0]
			return b == 0 || (rstart < b && rstart+rlength-1 > b)
		}
		for rest := s; ; {
			loc := mrpWindowStmt.FindStringIndex(rest)
			if loc == nil {
				break
			}
			if cross(loc) {
				m := rest[loc[0]:loc[1]]
				n, _ := strconv.Atoi(m[len("first ") : len(m)-len(" lines")])
				if n != win {
					rows = append(rows, fmt.Sprintf(`%s:%d: states the window as "first %d lines" — a near-miss of the canonical window %s; every window statement must read "first %s lines"`, file, ln, n, window, window))
				}
			}
			rest = rest[loc[1]:]
		}
		for rest := s; ; {
			loc := mrpMarkerSpan.FindStringIndex(rest)
			if loc == nil {
				break
			}
			if span := rest[loc[0]:loc[1]]; cross(loc) && span != "`"+marker+"`" {
				rows = append(rows, fmt.Sprintf("%s:%d: carries the backticked mutation-marker span %s, which is not the canonical marker `%s`", file, ln, span, marker))
			}
			rest = rest[loc[1]:]
		}
	}
	prev := ""
	for i, l := range mrpLines(text) {
		scan(l, i+1, 0)
		if i > 0 {
			scan(prev+" "+l, i, len(prev)+1)
		}
		prev = l
	}
	sort.Strings(rows)
	out := rows[:0]
	for i, r := range rows {
		if i == 0 || r != rows[i-1] {
			out = append(out, r)
		}
	}
	return out
}

// mrpLines is a file's lines as grep and awk read them: a final line without
// its newline still counts, and a trailing newline adds no empty line.
func mrpLines(text string) []string {
	lines := strings.Split(text, "\n")
	if strings.HasSuffix(text, "\n") {
		lines = lines[:len(lines)-1]
	}
	return lines
}
