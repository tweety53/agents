package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/records"
)

const selfReviewUsage = `usage: flow self-review bundle [-addr url] [-timeout dur] [-C dir]
                             -change name

Prints a change's whole self-review context bundle on stdout --
the ledger and panel record rendered from the store, the archived change's
tasks.md, design.md and narrative.md, and the git log of the finish-run
commits, assembled by the daemon. The CLI constructs none of it, the
record-render rule: the bundle's shape is decided where it is produced.

This is a read, with the findings read's contract: a store that cannot be
reached is reported to stderr and exits non-zero -- never a partial bundle
on stdout that a caller could mistake for the change's own. The only other
non-zero exits are caller mistakes: a missing -change, or a stray
positional argument.

usage: flow self-review finding [-addr url] [-timeout dur] [-C dir]
                              -change name -angle label
                              -disposition fixed|filed|declined -note text
                              [-ref sha|KEY] [-blast-radius N]

Records one self-review finding and its outcome (KAN-875): -ref is the
landed sha for fixed, the issue key for filed and omitted for declined;
-blast-radius is the finding's file count, omitted when none was counted.

This is a write that never blocks: a store that cannot be reached, or
that fails the write for any reason but a refusal, is one warning line on
stderr and exit 0, with no journal -- the committed
self-review report already carries every finding's disposition. A store
refusal (a disposition outside the closed set, a ref that does not fit it)
is the caller's mistake and exits 2 with the store's message, as does a
missing required flag or a negative -blast-radius.

usage: flow self-review findings [-addr url] [-timeout dur] [-C dir]
                               -change name

Prints the change's recorded self-review findings as a JSON array on
stdout, in the order they were recorded. A read, with the bundle's
contract: a store that cannot be reached exits non-zero with nothing on
stdout.
`

// runSelfReview implements `flow self-review`: `bundle`, the read run 1's
// bundle step fetches its reasoning input through, and `finding`/`findings`, the
// write and read of each finding's outcome.
func runSelfReview(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, selfReviewUsage)
		return 2
	}
	switch args[0] {
	case "bundle":
		return runSelfReviewBundle(ctx, args[1:], stdout, stderr)
	case "finding":
		return runSelfReviewFinding(ctx, args[1:], stderr)
	case "findings":
		return runSelfReviewFindings(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "flow: unknown self-review command %q\n", args[0])
		fmt.Fprint(stderr, selfReviewUsage)
		return 2
	}
}

func runSelfReviewBundle(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow self-review bundle", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordConnFlags(fset, &f)
	fset.StringVar(&f.change, "change", "", "the change whose bundle is served (required)")

	projectKey, mainCheckout, code := parseSelfReviewFlags(fset, &f, args, stderr)
	if code >= 0 {
		return code
	}

	bundle, err := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) ([]byte, error) {
		return cl.GetSelfReviewBundle(ctx, projectKey, f.change, mainCheckout)
	})
	if err != nil {
		fmt.Fprintf(stderr, "flow: self-review bundle: %v\n", err)
		return 1
	}
	_, _ = stdout.Write(bundle)
	return 0
}

// parseSelfReviewFlags parses one self-review subcommand's flags, refuses a
// stray positional argument and a missing -change, and resolves the project
// key and main checkout. It returns a non-negative exit code when the
// caller should stop, and -1 when it should go on.
func parseSelfReviewFlags(fset *flag.FlagSet, f *recordIdentityFlags, args []string, stderr io.Writer) (projectKey, mainCheckout string, code int) {
	if err := fset.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return "", "", 0
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, selfReviewUsage)
		return "", "", 2
	}
	noteAddrUsage(fset, stderr, f.addr)
	if fset.NArg() != 0 {
		fmt.Fprint(stderr, "flow: expected no positional arguments\n")
		fmt.Fprint(stderr, selfReviewUsage)
		return "", "", 2
	}
	if f.change == "" {
		fmt.Fprint(stderr, "flow: -change is required\n")
		fmt.Fprint(stderr, selfReviewUsage)
		return "", "", 2
	}
	if f.dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(stderr, "flow: resolve working directory: %v\n", err)
			return "", "", 1
		}
		f.dir = wd
	}
	projectKey, mainCheckout, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return "", "", 1
	}
	return projectKey, mainCheckout, -1
}

func runSelfReviewFinding(ctx context.Context, args []string, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow self-review finding", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	var in records.SelfReviewFinding
	blastRadius := -1
	registerRecordConnFlags(fset, &f)
	fset.StringVar(&f.change, "change", "", "the change the finding belongs to (required)")
	fset.StringVar(&in.Angle, "angle", "", "the finding's angle label (required)")
	fset.StringVar(&in.Disposition, "disposition", "", "fixed, filed or declined (required)")
	fset.StringVar(&in.Note, "note", "", "the finding, one line (required)")
	fset.StringVar(&in.Ref, "ref", "", "the landed sha for fixed, the issue key for filed")
	fset.IntVar(&blastRadius, "blast-radius", -1, "the finding's file count (optional)")

	projectKey, _, code := parseSelfReviewFlags(fset, &f, args, stderr)
	if code >= 0 {
		return code
	}
	for _, req := range []struct{ name, val string }{{"-angle", in.Angle}, {"-disposition", in.Disposition}, {"-note", in.Note}} {
		if req.val == "" {
			fmt.Fprintf(stderr, "flow: %s is required\n", req.name)
			fmt.Fprint(stderr, selfReviewUsage)
			return 2
		}
	}
	setByCaller := false
	fset.Visit(func(fl *flag.Flag) { setByCaller = setByCaller || fl.Name == "blast-radius" })
	if setByCaller {
		if blastRadius < 0 {
			fmt.Fprintf(stderr, "flow: -blast-radius must be 0 or more, got %d\n", blastRadius)
			return 2
		}
		in.BlastRadius = &blastRadius
	}

	_, err := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.SelfReviewFinding, error) {
		return cl.RecordSelfReviewFinding(ctx, projectKey, f.change, in)
	})
	switch {
	case err == nil:
		return 0
	case errors.Is(err, client.ErrRecordRejected):
		fmt.Fprintf(stderr, "flow: self-review finding refused: %v\n", err)
		return 2
	default:
		fmt.Fprintf(stderr, "⚠ flow: self-review finding not recorded — the store could not take it: %v\n", err)
		return 0
	}
}

func runSelfReviewFindings(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow self-review findings", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordConnFlags(fset, &f)
	fset.StringVar(&f.change, "change", "", "the change whose findings are listed (required)")

	projectKey, _, code := parseSelfReviewFlags(fset, &f, args, stderr)
	if code >= 0 {
		return code
	}
	rows, err := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) ([]records.SelfReviewFinding, error) {
		return cl.ListSelfReviewFindings(ctx, projectKey, f.change)
	})
	if err != nil {
		fmt.Fprintf(stderr, "flow: self-review findings: %v\n", err)
		return 1
	}
	if rows == nil {
		rows = []records.SelfReviewFinding{}
	}
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rows); err != nil {
		fmt.Fprintf(stderr, "flow: self-review findings: %v\n", err)
		return 1
	}
	return 0
}
