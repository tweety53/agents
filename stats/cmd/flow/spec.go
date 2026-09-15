package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/records"
)

// specUsage is `flow spec`'s usage block, in the shape suiteUsage states
// its verbs: flags first, then the one paragraph a caller needs about
// failure semantics.
var specUsage = `usage: flow spec record -spec file [-addr url] [-timeout dur] [-C dir]
flow spec list   [-dir dir] [-json] [-addr url] [-timeout dur] [-C dir]

spec record marks one spec file as run now: the store upserts the spec's
last-run timestamp, keeping the LATEST one it holds, and answers the
inventory row -- the stamp and the changes-since figure. Unlike suite
record there is no child command to run: the caller runs the spec itself
and records the run beside it. A store that cannot be reached is the one
failure: reported, exit 1 -- the record is this command's whole job.

list prints the per-spec inventory, one row per spec: the spec, its
last-run timestamp and the unrun-for-N stat -- how many of the project's
changes postdate that run. A spec with no recorded run renders as never
with the project's whole change count, which is how a suite that never ran
surfaces as a stat rather than by accident. -dir inventories the named
directory's *.spec.ts files: every file on disk answers a row, run or
never-run, recorded specs that no longer sit on the disk drop out, and a
directory holding no spec files at all answers that plainly rather than
the recorded-only listing. -json emits the store's array alone.
`

// runSpec routes `flow spec`'s two verbs. record writes; list reads.
func runSpec(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, specUsage)
		return 2
	}
	switch args[0] {
	case "record":
		return runSpecRecord(ctx, args[1:], stdout, stderr)
	case "list":
		return runSpecList(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "flow: unknown spec subcommand %q\n", args[0])
		fmt.Fprint(stderr, specUsage)
		return 2
	}
}

// runSpecRecord implements `flow spec record`: the spec's name is posted
// and the store's stamped row is confirmed on stderr. The store refusing
// (a 400, a 404) and the store being unreachable are both reported and
// exit 1 -- there is no child whose verdict could stand in their place.
func runSpecRecord(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow spec record", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordConnFlags(fset, &f)
	spec := fset.String("spec", "", "the spec file that ran, by basename (required)")

	if ok, code := parseRecordConnFlags(fset, &f, args, stderr); !ok {
		return code
	}
	if *spec == "" {
		fmt.Fprint(stderr, specUsage)
		return 2
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	out, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.SpecLastRun, error) {
		return cl.RecordSpecRun(ctx, projectKey, *spec)
	})
	if callErr != nil {
		fmt.Fprintf(stderr, "flow: spec record: %v\n", callErr)
		return 1
	}
	stamp := "never"
	if out.LastRanAt != nil {
		stamp = out.LastRanAt.Format(time.RFC3339)
	}
	fmt.Fprintf(stderr, "recorded: spec %s %s (%d changes since)\n", out.Spec, stamp, out.ChangesSince)
	return 0
}

// runSpecList implements `flow spec list`: rows ordered by spec, each
// carrying the unrun-for-N stat, a never-recorded spec rendering as never.
// -dir names the directory whose *.spec.ts files are the inventory -- the
// caller's answer to "which spec files exist", a question the store cannot
// ask on its own. A -dir that does not exist is reported and exits 1: a
// mistyped path silently degrading to the recorded-only listing would be
// the very accident this inventory exists to prevent.
func runSpecList(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow spec list", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordConnFlags(fset, &f)
	dir := fset.String("dir", "", "inventory this directory's *.spec.ts files (optional)")
	asJSON := fset.Bool("json", false, "print the store's array alone")

	if ok, code := parseRecordConnFlags(fset, &f, args, stderr); !ok {
		return code
	}

	var specs []string
	if *dir != "" {
		resolved := *dir
		if !filepath.IsAbs(resolved) {
			resolved = filepath.Join(f.dir, resolved)
		}
		if info, err := os.Stat(resolved); err != nil || !info.IsDir() {
			fmt.Fprintf(stderr, "flow: spec list: -dir %s is not a directory\n", *dir)
			return 1
		}
		matches, err := filepath.Glob(filepath.Join(resolved, "*.spec.ts"))
		if err != nil {
			fmt.Fprintf(stderr, "flow: spec list: scan %s: %v\n", *dir, err)
			return 1
		}
		for _, m := range matches {
			specs = append(specs, filepath.Base(m))
		}
		// A named universe that is empty is its own answer, never a fall
		// through to the recorded-only listing: the guard above refuses a
		// nonexistent dir for exactly this reason, and an existing dir with
		// no spec files deserves the same honesty.
		if len(specs) == 0 {
			fmt.Fprintf(stdout, "no *.spec.ts files in %s\n", *dir)
			return 0
		}
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	out, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) ([]records.SpecLastRun, error) {
		return cl.ListSpecLastRuns(ctx, projectKey, specs)
	})
	if callErr != nil {
		fmt.Fprintf(stderr, "flow: spec list: %v\n", callErr)
		return 1
	}

	if *asJSON {
		if out == nil {
			out = []records.SpecLastRun{}
		}
		body, err := json.Marshal(out)
		if err != nil {
			fmt.Fprintf(stderr, "flow: encode spec last runs: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, string(body))
		return 0
	}

	if len(out) == 0 {
		fmt.Fprintln(stdout, "no spec runs recorded")
		return 0
	}
	for _, row := range out {
		stamp := "never"
		if row.LastRanAt != nil {
			stamp = row.LastRanAt.Format(time.RFC3339)
		}
		fmt.Fprintf(stdout, "%s\t%s\t%d changes since\n", row.Spec, stamp, row.ChangesSince)
	}
	return 0
}
