package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
)

// repriceTimeout is reprice's own default request timeout: Price runs once
// per stage run in the window, serially, and the 56-day window holds about
// 15,000 runs -- far past defaultTimeout's 2s.
const repriceTimeout = 30 * time.Minute

const pricingUsage = `usage: flow pricing reprice -since YYYY-MM-DD [-addr url] [-timeout dur]

pricing reprice prices every stage run started on or after -since (UTC)
again, so dispatches harvested before their served model was priced get
their cost, and prints "priced N, skipped M". A run with no tokens or no
rate in effect counts as skipped.
`

func runPricing(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "reprice" {
		if len(args) > 0 {
			fmt.Fprintf(stderr, "flow: unknown pricing command %q\n", args[0])
		}
		fmt.Fprint(stderr, pricingUsage)
		return 2
	}

	fset := flag.NewFlagSet("flow pricing reprice", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var since string
	var f storeConnFlags
	registerStoreConnFlags(fset, &f)
	// Overrides the -timeout default registerStoreConnFlags set.
	fset.Lookup("timeout").DefValue = repriceTimeout.String()
	f.timeout = repriceTimeout
	fset.StringVar(&since, "since", "", "reprice stage runs started on or after this date (YYYY-MM-DD)")
	if err := fset.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, pricingUsage)
		return 2
	}
	noteAddrUsage(fset, stderr, f.addr)
	if _, err := time.Parse(time.DateOnly, since); err != nil || fset.NArg() != 0 {
		fmt.Fprintf(stderr, "flow: pricing reprice needs -since YYYY-MM-DD, got %q\n", since)
		fmt.Fprint(stderr, pricingUsage)
		return 2
	}

	reqCtx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()
	res, err := client.New(f.addr, &http.Client{Timeout: f.timeout}).Reprice(reqCtx, since)
	if err != nil {
		fmt.Fprintf(stderr, "flow: pricing reprice: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "priced %d, skipped %d\n", res.Priced, res.Skipped)
	return 0
}
