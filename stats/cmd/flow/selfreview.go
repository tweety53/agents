package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
)

const selfReviewUsage = `usage: flow self-review bundle [-addr url] [-timeout dur] [-C dir]
                             -change name

Prints a finished change's whole self-review context bundle on stdout --
the ledger and panel record rendered from the store, the archived change's
tasks.md, design.md and narrative.md, and the git log of the finish-run
commits, assembled by the daemon. The CLI constructs none of it, the
record-render rule: the bundle's shape is decided where it is produced.

This is a read, with the findings read's contract: a store that cannot be
reached is reported to stderr and exits non-zero -- never a partial bundle
on stdout that a caller could mistake for the change's own. The only other
non-zero exits are caller mistakes: a missing -change, or a stray
positional argument.
`

// runSelfReview implements `flow self-review`. One subcommand today:
// `bundle`, the read run 2 step 9 fetches its reasoning input through.
func runSelfReview(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, selfReviewUsage)
		return 2
	}
	switch args[0] {
	case "bundle":
		return runSelfReviewBundle(ctx, args[1:], stdout, stderr)
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
	fset.StringVar(&f.change, "change", "", "the finished change whose bundle is served (required)")

	if err := fset.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, selfReviewUsage)
		return 2
	}
	noteAddrEnvUsage(fset, stderr)
	if fset.NArg() != 0 {
		fmt.Fprint(stderr, "flow: expected no positional arguments\n")
		fmt.Fprint(stderr, selfReviewUsage)
		return 2
	}
	if f.change == "" {
		fmt.Fprint(stderr, "flow: -change is required\n")
		fmt.Fprint(stderr, selfReviewUsage)
		return 2
	}
	if f.dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(stderr, "flow: resolve working directory: %v\n", err)
			return 1
		}
		f.dir = wd
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	bundle, err := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) ([]byte, error) {
		return cl.GetSelfReviewBundle(ctx, projectKey, f.change)
	})
	if err != nil {
		fmt.Fprintf(stderr, "flow: self-review bundle: %v\n", err)
		return 1
	}
	_, _ = stdout.Write(bundle)
	return 0
}
