package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
)

const statsUsage = `flow stats bottlenecks [-addr url] [-timeout dur] [-last N] [-project key] [-json]

bottlenecks ranks where the last N changes (default 20; by their latest
stage run, every project unless -project names one) spent wall-clock time,
from rows flowd already records: stage runs by stage key, dispatches by
role and by role class, guard runs by guard -- their work, the run less
the time it spent waiting (a heavy slot, await-rows between reads), the
wait ranked apart under guard-wait and the scripts/flow-guard bootstrap
before it under guard-bootstrap -- and verify runs by label. Each
dimension prints its top 10 as name, runs, total, share of the changes'
summed wall clock, p50 and p95; one closing line names the top 5 time sinks
across dimensions (role class left out, as it re-counts the roles).
Dimensions overlap -- a dispatch runs inside a stage -- so shares across
them do not add up. -json prints flowd's response as is. Exits 1 when flowd
cannot be reached, 2 on a usage mistake.
`

type bottleneckRow struct {
	Dimension string  `json:"dimension"`
	Name      string  `json:"name"`
	Runs      int     `json:"runs"`
	TotalMs   int64   `json:"totalMs"`
	Share     float64 `json:"share"`
	P50Ms     int64   `json:"p50Ms"`
	P95Ms     int64   `json:"p95Ms"`
}

type bottlenecksReport struct {
	Last       int                        `json:"last"`
	WallMs     int64                      `json:"wallMs"`
	Dimensions []string                   `json:"dimensions"`
	Ranked     map[string][]bottleneckRow `json:"ranked"`
	Sinks      []bottleneckRow            `json:"sinks"`
}

func runStats(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "trace" {
		return runStatsTrace(ctx, args[1:], getenv, stdout, stderr)
	}
	if len(args) == 0 || args[0] != "bottlenecks" {
		fmt.Fprint(stderr, statsUsage+"\n"+statsTraceUsage)
		return 2
	}
	var f storeConnFlags
	fset := flag.NewFlagSet("flow stats bottlenecks", flag.ContinueOnError)
	fset.SetOutput(stderr)
	registerStoreConnFlags(fset, &f, getenv)
	last := fset.Int("last", 20, "rank the last N changes")
	project := fset.String("project", "", "one project's changes only (default: every project)")
	asJSON := fset.Bool("json", false, "print flowd's JSON response")
	if err := fset.Parse(args[1:]); err != nil || fset.NArg() != 0 || *last < 1 {
		fmt.Fprint(stderr, statsUsage)
		return 2
	}
	raw, err := withFlowdClient(ctx, f, func(ctx context.Context, cl *client.Client) (json.RawMessage, error) {
		return cl.Bottlenecks(ctx, *project, *last)
	})
	if err != nil {
		fmt.Fprintf(stderr, "flow: stats bottlenecks: %v\n", err)
		return 1
	}
	if *asJSON {
		fmt.Fprintln(stdout, string(raw))
		return 0
	}
	var rep bottlenecksReport
	if err := json.Unmarshal(raw, &rep); err != nil {
		fmt.Fprintf(stderr, "flow: stats bottlenecks: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "last %d changes, wall clock %s\n", rep.Last, dur(rep.WallMs))
	for _, d := range rep.Dimensions {
		fmt.Fprintf(stdout, "\n%s\n", d)
		tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "name\truns\ttotal\tshare\tp50\tp95")
		for _, r := range rep.Ranked[d] {
			fmt.Fprintf(tw, "%s\t%d\t%s\t%.1f%%\t%s\t%s\n", r.Name, r.Runs, dur(r.TotalMs), r.Share*100, dur(r.P50Ms), dur(r.P95Ms))
		}
		_ = tw.Flush()
	}
	sinks := make([]string, len(rep.Sinks))
	for i, r := range rep.Sinks {
		sinks[i] = fmt.Sprintf("%s %s %s (%.1f%%)", r.Dimension, r.Name, dur(r.TotalMs), r.Share*100)
	}
	fmt.Fprintf(stdout, "\ntime sinks: %s\n", strings.Join(sinks, "; "))
	return 0
}

// dur prints ms rounded to the second (to the millisecond under one).
func dur(ms int64) string {
	d := time.Duration(ms) * time.Millisecond
	if d >= time.Second {
		d = d.Round(time.Second)
	}
	return d.String()
}
