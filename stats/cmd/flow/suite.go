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

record turns the go test -json lines of the command's stdout back into
output text close to what go test prints without -v -- a test's output only when it
fails, or never ends -- every other line passing through unchanged, and
records each
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

// suiteWaitDelay is how long record waits, once the child has exited, for
// its stdout to close: a descendant still holding it is cut off there. A var
// only so a test can shorten it (TestSuiteRecordReturnsWhenTheChildExits).
var suiteWaitDelay = time.Second

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
	// The child writes to a pipe writer, never an *os.File, so exec copies
	// its stdout and WaitDelay bounds that copy: a descendant that inherits
	// stdout and outlives the child cannot hold the run open past it.
	childOut, childIn := io.Pipe()
	child.Stdout = childIn
	child.WaitDelay = suiteWaitDelay
	decoded := make(chan []records.SuiteRunTest, 1)
	go func() { decoded <- copyDecodingGoTestJSON(childOut, stdout) }()

	start := time.Now()
	if err := child.Start(); err != nil {
		// The child never started: nothing was measured, so nothing is
		// recorded -- a duration for a run that did not happen would be a
		// figure with no run behind it.
		_ = childIn.Close()
		<-decoded
		fmt.Fprintf(stderr, "flow: suite %s could not start %s: %v\n", *suite, command[0], err)
		return 127
	}
	runErr := child.Wait()
	measured := time.Since(start)
	_ = childIn.Close()
	costs := <-decoded

	exitCode := 0
	var exitErr *exec.ExitError
	switch {
	case errors.As(runErr, &exitErr):
		exitCode = exitErr.ExitCode()
	case errors.Is(runErr, exec.ErrWaitDelay):
		// The child exited 0; only a descendant still held its stdout.
	case runErr != nil:
		// Wait failed for a reason other than the child's exit status: the
		// run happened, so it is still recorded, with exit 1.
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
// no line-length limit (decision json-stream-line-detection), as go test
// prints without -v (a failing test's held log keeps test2json's === lines): a go test -json event with no Test writes its Output at
// once; a test's Output is held until its terminal event and written only
// when that is fail -- or at EOF, for a test that never ended. Every event is
// folded into the per-test costs; every other line -- a last one without a
// newline included -- is written through unchanged. It returns each
// non-skipped top-level test's cost, by testcost's rule (decision
// store-every-test).
func copyDecodingGoTestJSON(in io.Reader, out io.Writer) []records.SuiteRunTest {
	type testKey struct{ pkg, test string }
	r := bufio.NewReader(in)
	var fold testcost.Fold
	held := map[testKey]string{}
	var order []testKey // first-seen order, for the EOF flush
	for {
		line, err := r.ReadBytes('\n')
		if len(line) > 0 {
			if ev, ok := testcost.ParseEvent(line); ok {
				fold.Add(ev)
				k := testKey{ev.Package, ev.Test}
				switch {
				case ev.Test == "":
					_, _ = io.WriteString(out, ev.Output)
				case ev.Action == "pass" || ev.Action == "skip":
					delete(held, k)
				case ev.Action == "fail":
					_, _ = io.WriteString(out, held[k]+ev.Output)
					delete(held, k)
				default:
					if _, seen := held[k]; !seen {
						order = append(order, k)
					}
					held[k] += ev.Output
				}
			} else {
				_, _ = out.Write(line)
			}
		}
		if err != nil {
			break
		}
	}
	for _, k := range order {
		_, _ = io.WriteString(out, held[k])
		delete(held, k)
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
