package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/records"
)

// hazardUsage is `flow hazard` and `flow hazards`' own usage block, in the
// shape recordUsage states its subcommands: flags first, then the one
// paragraph a caller needs about failure semantics. hazardAppliesSet is the
// closed applies vocabulary, mirrored by the store's CHECK constraint and
// validated here before the store is ever contacted.
var hazardUsage = `flow hazard add      [-addr url] [-timeout dur] [-C dir]
                      -name name -text text -applies all|cross-repo|single-repo
flow hazard remove   [-addr url] [-timeout dur] [-C dir] -name name
flow hazards         [-addr url] [-timeout dur] [-C dir]
                     [-shape all|cross-repo|single-repo] [-all]

hazard add records one proactive, per-project warning a dispatch bundle
carries before it costs time -- incidents' proactive sibling. Rows are
retired, never deleted: hazard remove sets active=false, keeping the record
of what warnings a project carried. hazards reads them back as a JSON
array, newest first; without -shape it is the operator's unfiltered
listing, with -shape it filters to that shape plus the always-on rows, and
-all includes retired rows.

A write never blocks: on any store failure the intent is journalled, one
warning line is printed, and the command exits 0. The only non-zero exits
are caller mistakes -- a missing required flag or an -applies/-shape
outside the closed vocabulary, each checked before the store is contacted
-- and a write the store was reached for and refused (a duplicate name,
an unknown name on remove).
`

// hazardAppliesSet is the closed applies vocabulary.
var hazardAppliesSet = map[string]bool{"all": true, "cross-repo": true, "single-repo": true}

// runHazard routes `flow hazard`'s two verbs. add records; remove retires.
// Neither is optional and neither is a variant of the other.
func runHazard(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, hazardUsage)
		return 2
	}
	switch args[0] {
	case "add":
		return runHazardAdd(ctx, args[1:], stdout, stderr)
	case "remove":
		return runHazardRemove(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "flow: unknown hazard subcommand %q\n", args[0])
		fmt.Fprint(stderr, hazardUsage)
		return 2
	}
}

// runHazardAdd implements `flow hazard add`: a per-project hazard row,
// recorded active. -applies is validated against the closed vocabulary
// before the store is contacted -- a caller mistake a retry could never
// fix, so it is refused outright rather than journalled.
func runHazardAdd(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow hazard add", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordConnFlags(fset, &f)
	name := fset.String("name", "", "a short stable identifier for the hazard (required)")
	text := fset.String("text", "", "the warning text a dispatch bundle carries (required)")
	applies := fset.String("applies", "", "all, cross-repo or single-repo (required)")

	if ok, code := parseRecordConnFlags(fset, &f, args, stderr); !ok {
		return code
	}
	if !requireRecordFlags(stderr,
		[2]string{"-name", *name},
		[2]string{"-text", *text},
		[2]string{"-applies", *applies},
	) {
		return 2
	}
	if !hazardAppliesSet[*applies] {
		fmt.Fprintf(stderr, "flow: -applies %q must be one of all, cross-repo, single-repo\n", *applies)
		fmt.Fprint(stderr, hazardUsage)
		return 2
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	in := records.Hazard{Name: *name, Body: *text, Applies: *applies}
	_, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.Hazard, error) {
		return cl.AddHazard(ctx, projectKey, in)
	})
	if callErr == nil {
		fmt.Fprintln(stdout, "recorded: hazard")
	}
	return classifyRecordWrite(callErr, projectKey, "", "hazard", in, stderr)
}

// runHazardRemove implements `flow hazard remove`: active=false on the
// named row, never a delete, so the record of what warnings a project
// carried survives. An unknown name is the store answering, reported and
// never journalled -- a replay of it could never succeed.
func runHazardRemove(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow hazard remove", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordConnFlags(fset, &f)
	name := fset.String("name", "", "the hazard's name (required)")

	if ok, code := parseRecordConnFlags(fset, &f, args, stderr); !ok {
		return code
	}
	if !requireRecordFlags(stderr, [2]string{"-name", *name}) {
		return 2
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	_, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.Hazard, error) {
		return cl.RetireHazard(ctx, projectKey, *name)
	})
	if callErr == nil {
		fmt.Fprintf(stdout, "retired: hazard %s\n", *name)
	}
	return classifyRecordWrite(callErr, projectKey, "", "hazard remove", *name, stderr)
}

// runHazards implements `flow hazards`: a project's hazards as a JSON
// array, newest first. Without -shape it is the operator's unfiltered
// listing; with -shape it filters to that shape plus the always-on rows;
// -all includes retired rows. A failed read never journals and exits
// non-zero, findings' own read contract.
func runHazards(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow hazards", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordConnFlags(fset, &f)
	shape := fset.String("shape", "", "filter to this shape plus the always-on rows (optional)")
	all := fset.Bool("all", false, "include retired rows")

	if ok, code := parseRecordConnFlags(fset, &f, args, stderr); !ok {
		return code
	}
	if *shape != "" && !hazardAppliesSet[*shape] {
		fmt.Fprintf(stderr, "flow: -shape %q must be one of all, cross-repo, single-repo\n", *shape)
		fmt.Fprint(stderr, hazardUsage)
		return 2
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	out, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) ([]records.Hazard, error) {
		return cl.ListHazards(ctx, projectKey, *shape, *all)
	})
	if callErr != nil {
		fmt.Fprintf(stderr, "flow: hazards: %v\n", callErr)
		return 1
	}
	if out == nil {
		out = []records.Hazard{}
	}
	body, err := json.Marshal(out)
	if err != nil {
		fmt.Fprintf(stderr, "flow: encode hazards: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, string(body))
	return 0
}
