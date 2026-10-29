package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/store"
)

const experimentUsage = `usage: flow experiment open      [-addr url] [-timeout dur] -name <name> -cell <role>=<pair>[,<pair>...] [-cell ...]
       flow experiment list      [-addr url] [-timeout dur] [-open]
       flow experiment overrides [-addr url] [-timeout dur]
       flow experiment stats <id> [-addr url] [-timeout dur]
       flow experiment main-check [-addr url] [-timeout dur] -model <model> -effort <effort>
       flow experiment close <id> [-addr url] [-timeout dur] -verdict positive|negative|inconclusive [-summary text]

An allowed-pairs experiment (KAN-1067) replaces whole unmarked rows of
skills/flow-contracts/model-policy.md's Allowed pairs table while it is open.
-cell repeats, one role each; role main takes one <model>@<effort>.

open prints "experiment <id> open". list prints one line per experiment,
newest first: <id> <status> <started yyyy-mm-dd> <name> <role>=<pairs>;...
overrides prints every open experiment's cells as one compact JSON object,
{} when none is open -- the allowed-pairs readers overlay it on the table.
stats prints the experiment's before/after figures per metric as the
daemon's JSON: runs per window, excluded main-cell mismatches, and each
metric's before and after means with its weekly series.
main-check prints one warning line when an open experiment's main cell names
another <model>@<effort> than this session's; it never blocks a run, so it
exits 0 on a match, with no main cell open and with the store unreachable.

Exit codes: 0 done; 1 the store refused (a role an open experiment holds,
a malformed experiment, an unknown or closed id) or could not be reached;
2 usage.
`

func runExperiment(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, experimentUsage)
		return 2
	}
	switch args[0] {
	case "open":
		return runExperimentOpen(ctx, args[1:], stdout, stderr)
	case "list":
		return runExperimentList(ctx, args[1:], stdout, stderr)
	case "overrides":
		return runExperimentOverrides(ctx, args[1:], stdout, stderr)
	case "close":
		return runExperimentClose(ctx, args[1:], stdout, stderr)
	case "stats":
		return runExperimentStats(ctx, args[1:], stdout, stderr)
	case "main-check":
		return runExperimentMainCheck(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "flow: unknown experiment command %q\n", args[0])
		fmt.Fprint(stderr, experimentUsage)
		return 2
	}
}

// cellFlags collects repeated -cell <role>=<pair>[,<pair>...] values.
type cellFlags map[string][]string

func (c cellFlags) String() string { return "" }

func (c cellFlags) Set(v string) error {
	role, pairs, ok := strings.Cut(v, "=")
	if !ok || role == "" || pairs == "" {
		return fmt.Errorf("want <role>=<pair>[,<pair>...], got %q", v)
	}
	if _, dup := c[role]; dup {
		return fmt.Errorf("role %q given twice", role)
	}
	c[role] = strings.Split(pairs, ",")
	return nil
}

// parseExperimentFlags parses args into fset, registering the store
// connection flags first; it returns (_, true) to go on, or (code, false)
// to stop with exit code code.
func parseExperimentFlags(fset *flag.FlagSet, f *storeConnFlags, args []string, stderr io.Writer) (int, bool) {
	fset.SetOutput(stderr)
	if err := fset.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0, false
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, experimentUsage)
		return 2, false
	}
	noteAddrUsage(fset, stderr, f.addr)
	if fset.NArg() != 0 {
		fmt.Fprintf(stderr, "flow: %s takes no further positional arguments\n", fset.Name())
		fmt.Fprint(stderr, experimentUsage)
		return 2, false
	}
	return 0, true
}

func newExperimentFlagSet(name string, f *storeConnFlags) *flag.FlagSet {
	fset := flag.NewFlagSet(name, flag.ContinueOnError)
	registerStoreConnFlags(fset, f)
	return fset
}

func runExperimentOpen(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	var f storeConnFlags
	fset := newExperimentFlagSet("flow experiment open", &f)
	name := fset.String("name", "", "the experiment's name")
	cells := cellFlags{}
	fset.Var(cells, "cell", "<role>=<pair>[,<pair>...]; repeat for each role")
	if code, ok := parseExperimentFlags(fset, &f, args, stderr); !ok {
		return code
	}
	if *name == "" || len(cells) == 0 {
		fmt.Fprintln(stderr, "flow: experiment open needs -name and at least one -cell")
		fmt.Fprint(stderr, experimentUsage)
		return 2
	}
	e, err := withExperimentClient(ctx, f, func(ctx context.Context, cl *client.Client) (client.Experiment, error) {
		return cl.OpenExperiment(ctx, *name, cells)
	})
	if err != nil {
		fmt.Fprintf(stderr, "flow: experiment open: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "experiment %d open\n", e.ID)
	return 0
}

func runExperimentList(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	var f storeConnFlags
	fset := newExperimentFlagSet("flow experiment list", &f)
	openOnly := fset.Bool("open", false, "list only open experiments")
	if code, ok := parseExperimentFlags(fset, &f, args, stderr); !ok {
		return code
	}
	list, err := withExperimentClient(ctx, f, func(ctx context.Context, cl *client.Client) ([]client.Experiment, error) {
		return cl.ListExperiments(ctx, *openOnly)
	})
	if err != nil {
		fmt.Fprintf(stderr, "flow: experiment list: %v\n", err)
		return 1
	}
	for _, e := range list {
		roles := make([]string, 0, len(e.Cells))
		for r := range e.Cells {
			roles = append(roles, r)
		}
		sort.Strings(roles)
		cells := make([]string, len(roles))
		for i, r := range roles {
			cells[i] = r + "=" + strings.Join(e.Cells[r], ",")
		}
		fmt.Fprintf(stdout, "%d %s %s %s %s\n", e.ID, e.Status, e.StartedAt.UTC().Format("2006-01-02"), e.Name, strings.Join(cells, ";"))
	}
	return 0
}

func runExperimentOverrides(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	var f storeConnFlags
	fset := newExperimentFlagSet("flow experiment overrides", &f)
	if code, ok := parseExperimentFlags(fset, &f, args, stderr); !ok {
		return code
	}
	o, err := withExperimentClient(ctx, f, func(ctx context.Context, cl *client.Client) (map[string][]string, error) {
		return cl.ExperimentOverrides(ctx)
	})
	if err != nil {
		fmt.Fprintf(stderr, "flow: experiment overrides: %v\n", err)
		return 1
	}
	if o == nil {
		o = map[string][]string{}
	}
	encoded, err := json.Marshal(o)
	if err != nil {
		fmt.Fprintf(stderr, "flow: experiment overrides: encode output: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "%s\n", encoded)
	return 0
}

// experimentIDArg reads the leading <id> argument of `flow experiment
// <cmd>`; ok false means usage was printed and the exit is 2.
func experimentIDArg(cmd string, args []string, stderr io.Writer) (int64, bool) {
	if len(args) == 0 {
		fmt.Fprintf(stderr, "flow: experiment %s needs an experiment id\n", cmd)
		fmt.Fprint(stderr, experimentUsage)
		return 0, false
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || id <= 0 {
		fmt.Fprintf(stderr, "flow: experiment %s: id must be a positive integer, got %q\n", cmd, args[0])
		fmt.Fprint(stderr, experimentUsage)
		return 0, false
	}
	return id, true
}

func runExperimentStats(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	id, ok := experimentIDArg("stats", args, stderr)
	if !ok {
		return 2
	}
	var f storeConnFlags
	fset := newExperimentFlagSet("flow experiment stats", &f)
	if code, ok := parseExperimentFlags(fset, &f, args[1:], stderr); !ok {
		return code
	}
	body, err := withExperimentClient(ctx, f, func(ctx context.Context, cl *client.Client) (json.RawMessage, error) {
		return cl.ExperimentStats(ctx, id)
	})
	if err != nil {
		fmt.Fprintf(stderr, "flow: experiment stats: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "%s\n", body)
	return 0
}

// runExperimentMainCheck warns, once, when an open experiment's main cell
// names another pair than this session's -model/-effort (decision
// main-check-and-tag). A cell model names a family (opus), the session may
// name a full id (claude-opus-5-5): the cell's model inside the session's
// counts as a match.
func runExperimentMainCheck(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	var f storeConnFlags
	fset := newExperimentFlagSet("flow experiment main-check", &f)
	model := fset.String("model", "", "this session's model")
	effort := fset.String("effort", "", "this session's effort")
	if code, ok := parseExperimentFlags(fset, &f, args, stderr); !ok {
		return code
	}
	if *model == "" || *effort == "" {
		fmt.Fprintln(stderr, "flow: experiment main-check needs -model and -effort")
		fmt.Fprint(stderr, experimentUsage)
		return 2
	}
	open, err := withExperimentClient(ctx, f, func(ctx context.Context, cl *client.Client) ([]client.Experiment, error) {
		return cl.ListExperiments(ctx, true)
	})
	if err != nil {
		return 0 // unreachable store: the table's own pairs apply, nothing to warn about
	}
	for _, e := range open {
		cell := e.Cells["main"]
		if len(cell) != 1 {
			continue
		}
		m, eff, _ := strings.Cut(cell[0], "@")
		if eff != *effort || !strings.Contains(*model, m) {
			fmt.Fprintf(stdout, "⚠ experiment %d: main agent should run %s; this session runs %s@%s\n", e.ID, cell[0], *model, *effort)
		}
	}
	return 0
}

func runExperimentClose(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	id, ok := experimentIDArg("close", args, stderr)
	if !ok {
		return 2
	}
	var f storeConnFlags
	fset := newExperimentFlagSet("flow experiment close", &f)
	verdict := fset.String("verdict", "", "positive, negative or inconclusive")
	summary := fset.String("summary", "", "the closing summary")
	if code, ok := parseExperimentFlags(fset, &f, args[1:], stderr); !ok {
		return code
	}
	if !store.ValidVerdicts[*verdict] {
		fmt.Fprintf(stderr, "flow: experiment close: -verdict must be positive, negative or inconclusive, got %q\n", *verdict)
		fmt.Fprint(stderr, experimentUsage)
		return 2
	}
	_, err := withExperimentClient(ctx, f, func(ctx context.Context, cl *client.Client) (struct{}, error) {
		return struct{}{}, cl.CloseExperiment(ctx, id, *verdict, *summary)
	})
	if err != nil {
		fmt.Fprintf(stderr, "flow: experiment close: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "experiment %d closed %s\n", id, *verdict)
	return 0
}

// withExperimentClient runs call against a client under f's address and
// timeout, recovering from any panic in the client path and reporting it as
// client.ErrUnavailable -- getSettings' guarantee.
func withExperimentClient[T any](ctx context.Context, f storeConnFlags, call func(context.Context, *client.Client) (T, error)) (out T, err error) {
	defer func() {
		if r := recover(); r != nil {
			var zero T
			out, err = zero, fmt.Errorf("%w: recovered panic: %v", client.ErrUnavailable, r)
		}
	}()
	reqCtx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()
	return call(reqCtx, client.New(f.addr, &http.Client{Timeout: f.timeout}))
}
