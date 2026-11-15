package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/testcost"
)

// suiteUsage is `flow suite`'s usage block, in the shape hazardUsage states
// its verbs: flags first, then the one paragraph a caller needs about
// failure semantics.
var suiteUsage = `usage: flow suite record -suite name [-addr url] [-timeout dur] [-C dir] -- <command> [args...]
flow suite list   [-suite name | -tests name] [-limit n] [-json] [-addr url] [-timeout dur] [-C dir]

suite record runs <command> with its stdin and stderr passing through
untouched and its stdout decoded as below, wall-times it, and records one
suite_runs row: the suite's name, this machine's hostname, the measured
duration, the command's own exit code and its per-test costs.
The command's exit code is this command's exit code -- the caller's tooling
sees the suite's real verdict. A child that cannot start is exit 127 with
nothing recorded.

The store never gates the wrapped run: on any store failure one warning
line is printed and the child's exit code stands -- the runtime is lost,
never the verdict. The only non-zero exits are caller mistakes (a missing
-suite or no -- and command words, each refused before anything executes)
and the child's own. list is a read: a store that cannot be reached is
reported and exits 1.

list prints rows newest first, then one summary line per (suite, host):
the median duration of the last 10 passing runs -- a failed run's duration
is diagnostic, never a runtime figure, and one cold-cache run must not move
the number. -json emits the store's array alone.

record turns each go test -json line of the command's stdout back into its
output text -- every other line passes through unchanged -- and records each
top-level test's cost with the run: its own elapsed time plus that of every
descendant that paused for t.Parallel, a skipped test left out. list -tests
<suite> prints the newest run of that suite that has costs, as its row line
and then its tests heaviest first (-limit caps them), and "no test costs
recorded for <suite>" when none has; -tests and -suite are exclusive, and
-json prints that run alone, null for none.

Both verbs resolve the store address from FLOW_RECORDS_ADDR when it is set,
then FLOW_ADDR: the summary compares runs across changes, so a row recorded
in an apply worktree must outlive that worktree's daemon. An explicit -addr
still wins.
`

// suiteSummaryWindow is how many passing runs each per-(suite, host)
// summary line medians over: enough recent history to smooth one outlier,
// small enough that a real slowdown shows within a few runs.
const suiteSummaryWindow = 10

// runSuite routes `flow suite`'s two verbs. record wraps; list reads.
// Neither is optional and neither is a variant of the other.
func runSuite(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, suiteUsage)
		return 2
	}
	switch args[0] {
	case "record":
		return runSuiteRecord(ctx, args[1:], getenv, stdout, stderr)
	case "list":
		return runSuiteList(ctx, args[1:], getenv, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "flow: unknown suite subcommand %q\n", args[0])
		fmt.Fprint(stderr, suiteUsage)
		return 2
	}
}

// suiteJournalName is the journal a suite run the store could not take is
// written to; a suite run belongs to the project, not to one change.
const suiteJournalName = "_suite-runs"

// runSuiteRecord implements `flow suite record`: the child runs with
// inherited stdin and stderr, its stdout decoded by copyDecodingGoTestJSON,
// is wall-timed, and its row is recorded. The store never
// gates the child -- a record failure prints one warning line and changes
// no exit code. The only non-zero exits are caller mistakes (usage, 2, a
// child that could not start, 127, nothing recorded) and the child's own.
func runSuiteRecord(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow suite record", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordConnFlags(fset, &f, getenv)
	suite := fset.String("suite", "", "the name this runtime figure belongs to (required)")

	// Parse stops at --; everything after it is the child command, handed
	// to the child verbatim.
	if err := fset.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprint(stderr, suiteUsage)
		return 2
	}
	command := fset.Args()
	if *suite == "" || len(command) == 0 {
		fmt.Fprint(stderr, suiteUsage)
		return 2
	}

	child := exec.Command(command[0], command[1:]...)
	child.Stdin = os.Stdin
	child.Stderr = stderr
	childOut, err := child.StdoutPipe()
	if err != nil {
		fmt.Fprintf(stderr, "flow: suite %s could not start %s: %v\n", *suite, command[0], err)
		return 127
	}

	start := time.Now()
	if err := child.Start(); err != nil {
		// The child never started: nothing was measured, so nothing is
		// recorded -- a duration for a run that did not happen would be a
		// figure with no run behind it.
		fmt.Fprintf(stderr, "flow: suite %s could not start %s: %v\n", *suite, command[0], err)
		return 127
	}
	costs := copyDecodingGoTestJSON(childOut, stdout)
	runErr := child.Wait()
	measured := time.Since(start)

	exitCode := 0
	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		exitCode = exitErr.ExitCode()
	} else if runErr != nil {
		// Wait failed for a reason other than the child's exit status (its
		// stdout copy broke): the run happened, so it is still recorded,
		// but its verdict is unknown -- never a pass.
		fmt.Fprintf(stderr, "flow: suite %s: wait for %s: %v\n", *suite, command[0], runErr)
		exitCode = 1
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	in := records.SuiteRun{Suite: *suite, Host: shortHostname(), DurationMs: measured.Milliseconds(), ExitCode: exitCode, Tests: costs}
	_, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.SuiteRun, error) {
		return cl.RecordSuiteRun(ctx, projectKey, in)
	})
	switch {
	case callErr == nil:
		fmt.Fprintf(stderr, "recorded: suite %s %s exit %d\n", *suite, measured.Round(time.Millisecond), exitCode)
	case errors.Is(callErr, client.ErrRecordRejected), errors.Is(callErr, client.ErrNotFound):
		fmt.Fprintf(stderr, "⚠ flow: suite run not recorded (exit %d stands): %v\n", exitCode, callErr)
	default:
		// The store could not take it: journal the run for replay, stamped
		// with its start so the replayed row keeps when it ran (kan-1040).
		in.RanAt = start
		journalRecordWrite(projectKey, suiteJournalName, "suite", in, stderr, getenv)
	}
	return exitCode
}

// copyDecodingGoTestJSON copies the child's stdout to out line by line, with
// no line-length limit (decision json-stream-line-detection): a line that is
// a go test -json event writes its Output, when it has one, and is folded
// into the per-test costs; every other line -- a last one without a newline
// included -- is written through unchanged. It returns each non-skipped
// top-level test's cost, by testcost's rule (decision store-every-test).
func copyDecodingGoTestJSON(in io.Reader, out io.Writer) []records.SuiteRunTest {
	r := bufio.NewReader(in)
	var fold testcost.Fold
	for {
		line, err := r.ReadBytes('\n')
		if len(line) > 0 {
			if ev, ok := testcost.ParseEvent(line); ok {
				_, _ = io.WriteString(out, ev.Output)
				fold.Add(ev)
			} else {
				_, _ = out.Write(line)
			}
		}
		if err != nil {
			break
		}
	}
	var tests []records.SuiteRunTest
	for _, row := range fold.Rows() {
		tests = append(tests, records.SuiteRunTest{Package: row.Package, Test: row.Test, CostMs: int64(math.Round(row.Cost * 1000))})
	}
	return tests
}

// runSuiteList implements `flow suite list`: rows newest first, then one
// summary line per (suite, host) -- the median of the last 10 passing runs.
// -json emits the store's array and nothing else. A failed read exits 1: it
// is a read, there is nothing to fall back to and nothing to gate.
func runSuiteList(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow suite list", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordConnFlags(fset, &f, getenv)
	suite := fset.String("suite", "", "list only this suite (optional)")
	limit := fset.Int("limit", 20, "how many rows to read")
	asJSON := fset.Bool("json", false, "print the store's array alone")
	tests := fset.String("tests", "", "print this suite's newest run that has per-test costs, heaviest tests first")

	if ok, code := parseRecordConnFlags(fset, &f, args, getenv, stderr); !ok {
		return code
	}
	if *tests != "" && *suite != "" {
		fmt.Fprint(stderr, suiteUsage)
		return 2
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}
	if *tests != "" {
		return printSuiteTests(ctx, f, projectKey, *tests, *limit, *asJSON, stdout, stderr)
	}

	out, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) ([]records.SuiteRun, error) {
		return cl.ListSuiteRuns(ctx, projectKey, *suite, *limit)
	})
	if callErr != nil {
		fmt.Fprintf(stderr, "flow: suite list: %v\n", callErr)
		return 1
	}

	if *asJSON {
		if out == nil {
			out = []records.SuiteRun{}
		}
		body, err := json.Marshal(out)
		if err != nil {
			fmt.Fprintf(stderr, "flow: encode suite runs: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, string(body))
		return 0
	}

	if len(out) == 0 {
		fmt.Fprintln(stdout, "no suite runs recorded")
		return 0
	}
	for _, run := range out {
		printSuiteRunRow(stdout, run)
	}
	for _, line := range suiteSummaryLines(out) {
		fmt.Fprintln(stdout, line)
	}
	return 0
}

// printSuiteRunRow prints one suite run as list's row line.
func printSuiteRunRow(w io.Writer, run records.SuiteRun) {
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\texit %d\n",
		run.Suite, run.RanAt.Format(time.RFC3339), time.Duration(run.DurationMs)*time.Millisecond, run.Host, run.ExitCode)
}

// printSuiteTests implements `flow suite list -tests <suite>` (decision
// suite-tests-read): the newest run of that suite that has per-test costs,
// as its row line and then one line per test, heaviest first -- cost
// seconds, package, test, as flow-guard test-cost prints them. -json prints
// the response alone, null for none. No such run is a fact, exit 0; a
// failed read exits 1 as list does.
func printSuiteTests(ctx context.Context, f recordIdentityFlags, projectKey, suite string, limit int, asJSON bool, stdout, stderr io.Writer) int {
	run, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (*records.SuiteRun, error) {
		return cl.LatestSuiteRunTests(ctx, projectKey, suite, limit)
	})
	if callErr != nil {
		fmt.Fprintf(stderr, "flow: suite list: %v\n", callErr)
		return 1
	}
	if asJSON {
		body, err := json.Marshal(run)
		if err != nil {
			fmt.Fprintf(stderr, "flow: encode suite run: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, string(body))
		return 0
	}
	if run == nil {
		fmt.Fprintf(stdout, "no test costs recorded for %s\n", suite)
		return 0
	}
	printSuiteRunRow(stdout, *run)
	for _, test := range run.Tests {
		fmt.Fprintf(stdout, "%.2f\t%s\t%s\n", float64(test.CostMs)/1000, test.Package, test.Test)
	}
	return 0
}

// suiteSummaryLines renders one summary per (suite, host) in the rows'
// order of first appearance: the median duration of the last
// suiteSummaryWindow passing runs, rows being newest first. A failed run
// renders as a row above but never enters a median.
func suiteSummaryLines(runs []records.SuiteRun) []string {
	type key struct{ suite, host string }
	order := []key{}
	seen := map[string]bool{}
	passing := map[string][]time.Duration{}
	for _, run := range runs {
		k := key{run.Suite, run.Host}
		id := fmt.Sprintf("%s\x00%s", k.suite, k.host)
		if !seen[id] {
			seen[id] = true
			order = append(order, k)
		}
		if run.ExitCode == 0 {
			passing[id] = append(passing[id], time.Duration(run.DurationMs)*time.Millisecond)
		}
	}

	lines := []string{}
	for _, k := range order {
		id := fmt.Sprintf("%s\x00%s", k.suite, k.host)
		durs := passing[id]
		if len(durs) > suiteSummaryWindow {
			durs = durs[:suiteSummaryWindow]
		}
		if len(durs) == 0 {
			lines = append(lines, fmt.Sprintf("%s on %s: no passing runs recorded", k.suite, k.host))
			continue
		}
		sort.Slice(durs, func(i, j int) bool { return durs[i] < durs[j] })
		median := durs[(len(durs)-1)/2]
		lines = append(lines, fmt.Sprintf("%s on %s: median %s over the last %d passing runs",
			k.suite, k.host, median.Round(time.Millisecond), len(durs)))
	}
	return lines
}

// shortHostname is the recorded host: os.Hostname's first dot-separated
// label, so a fully qualified name records beside its short form rather
// than as a second machine.
func shortHostname() string {
	host, err := os.Hostname()
	if err != nil {
		return "unknown-host"
	}
	return strings.SplitN(host, ".", 2)[0]
}
