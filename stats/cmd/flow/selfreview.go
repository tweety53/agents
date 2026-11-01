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
	"path/filepath"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/records"
)

const selfReviewUsage = `usage: flow self-review bundle [-addr url] [-timeout dur] [-C dir]
                             -change name

Prints a change's whole self-review context bundle on stdout --
the ledger and panel record rendered from the store, the change directory's
tasks.md, design.md, narrative.md, visual-verification.md and
live-verification.md read from the worktree on disk, and the git log of its
first-parent merge-base..HEAD range, assembled by the daemon. The CLI
constructs none of it, the record-render rule: the bundle's shape is decided
where it is produced.

Run inside the change's worktree (-C, default cwd), it also writes the
bundle to <worktree>/.superpowers/sdd/<change>/self-review-context.md,
the file run 2's flow change save keeps. Run anywhere else -- a
deferred run once cleanup removed the worktree -- it prints the bundle saved
in flowd (flow change file <change> self-review-context.md)
instead, exit 1 when flowd holds none.

This is a read, with the findings read's contract: a store that cannot be
reached is reported to stderr and exits non-zero -- never a partial bundle
on stdout that a caller could mistake for the change's own. The only other
non-zero exits are caller mistakes: a missing -change, or a stray
positional argument.

usage: flow self-review finding [-addr url] [-timeout dur] [-C dir]
                              -change name -angle label
                              -disposition fixed|filed|declined|noted -note text
                              [-ref sha|KEY] [-blast-radius N]
                              [-severity critical|important|minor]

Records one self-review finding and its outcome (KAN-875): -ref is the
landed sha for fixed, the issue key for filed and omitted for declined
and noted;
-blast-radius is the finding's file count, omitted when none was counted;
-severity is the finding's grade, required for fixed and filed and omitted for declined and noted when none was given.

This is a write that never blocks: a store that cannot be reached, or
that fails the write for any reason but a refusal, is one warning line on
stderr and exit 0, with no journal. A store
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

// selfReviewBundleTimeout is the bundle call's default deadline: the daemon
// assembles the bundle from the worktree and git, and the call has no fallback, so the
// store-write fallback deadline (defaultTimeout) does not apply (kan-1072
// self-review, F3).
const selfReviewBundleTimeout = time.Minute

func runSelfReviewBundle(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow self-review bundle", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordConnFlags(fset, &f)
	f.timeout = selfReviewBundleTimeout
	fset.Lookup("timeout").DefValue = selfReviewBundleTimeout.String()
	fset.StringVar(&f.change, "change", "", "the change whose bundle is served (required)")

	projectKey, mainCheckout, code := parseSelfReviewFlags(fset, &f, args, stderr)
	if code >= 0 {
		return code
	}

	// Inside the change's own linked worktree -- on spectre/<change>, or
	// <change> for /flow-fast and /flow-interactive -- the
	// daemon assembles the bundle from that worktree and the bundle is kept
	// beside it for `flow change save`; anywhere else (a deferred run after
	// cleanup, or another change's worktree) the saved bundle is the input.
	wt := worktreeRoot(f.dir)
	branch, _ := exec.Command("git", "-C", f.dir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if b := strings.TrimSpace(string(branch)); wt == "" || wt == realPath(mainCheckout) || (b != "spectre/"+f.change && b != f.change) {
		saved, err := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (string, error) {
			return cl.GetChangeFile(ctx, projectKey, f.change, "self-review-context.md")
		})
		if err != nil {
			fmt.Fprintf(stderr, "flow: self-review bundle: no worktree for %s and no saved bundle: %v\n", f.change, err)
			return 1
		}
		_, _ = io.WriteString(stdout, saved)
		return 0
	}

	bundle, err := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) ([]byte, error) {
		return cl.GetSelfReviewBundle(ctx, projectKey, f.change, wt)
	})
	if err != nil {
		fmt.Fprintf(stderr, "flow: self-review bundle: %v\n", err)
		return 1
	}
	out := filepath.Join(wt, ".superpowers", "sdd", f.change, "self-review-context.md")
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err == nil {
		err = os.WriteFile(out, bundle, 0o644)
	}
	if err != nil {
		fmt.Fprintf(stderr, "flow: self-review bundle: write %s: %v\n", out, err)
		return 1
	}
	_, _ = stdout.Write(bundle)
	return 0
}

// worktreeRoot is dir's git worktree root, symlinks resolved; "" outside
// any worktree.
func worktreeRoot(dir string) string {
	top, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return ""
	}
	return realPath(strings.TrimSpace(string(top)))
}

// realPath is p with symlinks resolved, p itself when that fails.
func realPath(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return p
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
	fset.StringVar(&in.Disposition, "disposition", "", "fixed, filed, declined or noted (required)")
	fset.StringVar(&in.Note, "note", "", "the finding, one line (required)")
	fset.StringVar(&in.Ref, "ref", "", "the landed sha for fixed, the issue key for filed")
	fset.IntVar(&blastRadius, "blast-radius", -1, "the finding's file count (optional)")
	fset.StringVar(&in.Severity, "severity", "", "critical, important or minor (required for fixed and filed)")

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
	if (in.Disposition == "fixed" || in.Disposition == "filed") && in.Severity == "" {
		fmt.Fprintf(stderr, "flow: -severity is required for -disposition %s\n", in.Disposition)
		fmt.Fprint(stderr, selfReviewUsage)
		return 2
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
