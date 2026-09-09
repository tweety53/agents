package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/records"
)

// suiteUsage is `flow suite`'s usage block, in the shape hazardUsage states
// its verbs: flags first, then the one paragraph a caller needs about
// failure semantics.
var suiteUsage = `usage: flow suite record -suite name [-addr url] [-timeout dur] [-C dir] -- <command> [args...]
flow suite list   [-suite name] [-limit n] [-json] [-addr url] [-timeout dur] [-C dir]

suite record runs <command> with its stdio passing through untouched,
wall-times it, and records one suite_runs row: the suite's name, this
machine's hostname, the measured duration and the command's own exit code.
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
`

// suiteSummaryWindow is how many passing runs each per-(suite, host)
// summary line medians over: enough recent history to smooth one outlier,
// small enough that a real slowdown shows within a few runs.
const suiteSummaryWindow = 10

// runSuite routes `flow suite`'s two verbs. record wraps; list reads.
// Neither is optional and neither is a variant of the other.
func runSuite(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, suiteUsage)
		return 2
	}
	switch args[0] {
	case "record":
		return runSuiteRecord(ctx, args[1:], stdout, stderr)
	case "list":
		return runSuiteList(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "flow: unknown suite subcommand %q\n", args[0])
		fmt.Fprint(stderr, suiteUsage)
		return 2
	}
}

// runSuiteRecord implements `flow suite record`: the child runs with
// inherited stdio, is wall-timed, and its row is recorded. The store never
// gates the child -- a record failure prints one warning line and changes
// no exit code. The only non-zero exits are caller mistakes (usage, 2, a
// child that could not start, 127, nothing recorded) and the child's own.
func runSuiteRecord(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow suite record", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordConnFlags(fset, &f)
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
	child.Stdout = stdout
	child.Stderr = stderr

	start := time.Now()
	runErr := child.Run()
	measured := time.Since(start)

	exitCode := 0
	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		exitCode = exitErr.ExitCode()
	} else if runErr != nil {
		// The child never started: nothing was measured, so nothing is
		// recorded -- a duration for a run that did not happen would be a
		// figure with no run behind it.
		fmt.Fprintf(stderr, "flow: suite %s could not start %s: %v\n", *suite, command[0], runErr)
		return 127
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	in := records.SuiteRun{Suite: *suite, Host: shortHostname(), DurationMs: measured.Milliseconds(), ExitCode: exitCode}
	_, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.SuiteRun, error) {
		return cl.RecordSuiteRun(ctx, projectKey, in)
	})
	if callErr == nil {
		fmt.Fprintf(stderr, "recorded: suite %s %s exit %d\n", *suite, measured.Round(time.Millisecond), exitCode)
	} else {
		fmt.Fprintf(stderr, "⚠ flow: suite run not recorded (exit %d stands): %v\n", exitCode, callErr)
	}
	return exitCode
}

// runSuiteList implements `flow suite list`: rows newest first, then one
// summary line per (suite, host) -- the median of the last 10 passing runs.
// -json emits the store's array and nothing else. A failed read exits 1: it
// is a read, there is nothing to fall back to and nothing to gate.
func runSuiteList(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow suite list", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordConnFlags(fset, &f)
	suite := fset.String("suite", "", "list only this suite (optional)")
	limit := fset.Int("limit", 20, "how many rows to read")
	asJSON := fset.Bool("json", false, "print the store's array alone")

	if ok, code := parseRecordConnFlags(fset, &f, args, stderr); !ok {
		return code
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
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
		fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\texit %d\n",
			run.Suite, run.RanAt.Format(time.RFC3339), time.Duration(run.DurationMs)*time.Millisecond, run.Host, run.ExitCode)
	}
	for _, line := range suiteSummaryLines(out) {
		fmt.Fprintln(stdout, line)
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
