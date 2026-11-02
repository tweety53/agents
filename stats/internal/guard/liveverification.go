package guard

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/check-live-verification.sh shim (KAN-1058).
//
// flow-guard check-live-verification — refuse a live-verification record that carries
// no measured before figure (kan-998: the record read "Before figure: not
// measured on a merge-base stack", and **Live check** accepted it).
//
// Usage: flow-guard check-live-verification <changeRoot>/live-verification.md
//        flow-guard check-live-verification <changeRoot>
//
// Given the change directory instead — the gate flow.verify's stage end runs
// in-process (stats/cmd/flow/stage.go, gateLiveVerification;
// skills/flow/verify-and-handoff.md, **Close the stage**) — it
// reads <changeRoot>/design.md: a design naming no live check (no `## Live
// check` section, or one whose first line opens `None`, `No ` or `Not
// applicable`) prints `LIVE-CHECK-NONE: <changeRoot>`, exit 0; a design naming
// one judges <changeRoot>/live-verification.md as below, a missing record
// printing `<record>: missing — design.md names a live check`, exit 1; a
// directory with no design.md cannot answer, exit 2 (kan-1008's self-review:
// the stage end checked nothing, and the live check first ran at integrate).
//
// A line NAMES a before figure when it carries the word `before`, any case, as
// a whole word (`beforehand` is not one). A table header `| Before |` counts —
// lenient by design: the guard cannot tell a header from a figure. ANY line is
// UNMEASURED when it carries `not measured`, `unmeasured` or `not taken`, any
// case, as whole words — whether or not it names `before`, so a table whose
// `Before` header sits above a `not measured` cell fails too.
//
// Prints to stdout:
//   LIVE-VERIFICATION-OK: <record>                       exit 0
//   <record>: no before figure — no line names one       exit 1
//   <record>:<n>: before figure not measured — <line>    exit 1, one per line
//
// Exit 0 a before figure is named and none is unmeasured, 1 otherwise,
// 2 cannot answer (a usage error, a record that cannot be read) — stdout empty,
// the cause on stderr.
//
// Go, in flow-guard (stats/internal/guard/liveverification.go).

// checkLiveVerification is scripts/check-live-verification.sh: that script's
// header comment is the contract. A record passes when some line names a
// before figure and no line states a figure was not measured.
func init() { Registry["check-live-verification"] = checkLiveVerification }

var (
	lvBefore     = regexp.MustCompile(`(?i)\bbefore\b`)
	lvUnmeasured = regexp.MustCompile(`(?i)\b(not measured|unmeasured|not taken)\b`)
)

func checkLiveVerification(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-live-verification"
	if len(args) != 1 {
		fmt.Fprintf(stderr, "%s: usage: flow-guard check-live-verification <live-verification.md | changeRoot>\n", prog)
		return 2
	}
	record := args[0]
	// A change directory is the stage-end gate (kan-1008's self-review): no
	// check named passes, a named one needs its record, judged as below.
	if isDir(record) {
		design := record + "/design.md"
		if !isFile(design) {
			fmt.Fprintf(stderr, "%s: cannot read %s\n", prog, design)
			return 2
		}
		if !icLiveCheckNamed(design) {
			fmt.Fprintf(stdout, "LIVE-CHECK-NONE: %s\n", record)
			return 0
		}
		record += "/live-verification.md"
		if !isFile(record) {
			fmt.Fprintf(stdout, "%s: missing — design.md names a live check\n", record)
			return 1
		}
	}
	data, err := os.ReadFile(record)
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read %s: %v\n", prog, record, err)
		return 2
	}
	named := false
	var bad []string
	for i, line := range strings.Split(string(data), "\n") {
		if lvBefore.MatchString(line) {
			named = true
		}
		if lvUnmeasured.MatchString(line) {
			bad = append(bad, fmt.Sprintf("%s:%d: before figure not measured — %s", record, i+1, line))
		}
	}
	switch {
	case len(bad) > 0:
		fmt.Fprintln(stdout, strings.Join(bad, "\n"))
		return 1
	case !named:
		fmt.Fprintf(stdout, "%s: no before figure — no line names one\n", record)
		return 1
	}
	fmt.Fprintf(stdout, "LIVE-VERIFICATION-OK: %s\n", record)
	return 0
}
