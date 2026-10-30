package main

// request.go — `flow record request begin|end` (KAN-1061): one
// /flow-interactive request's record, and the guard a landed end runs
// before anything is written (design.md's Request-end guard).

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/guard"
	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/stages"
)

// requestOutcomes is the closed set `request end -outcome` accepts.
var requestOutcomes = []string{"landed", "answered", "dropped"}

// runRecordRequest routes `flow record request` to its two halves.
func runRecordRequest(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, recordUsage)
		return 2
	}
	switch args[0] {
	case "begin":
		return runRecordRequestBegin(ctx, args[1:], stdout, stderr)
	case "end":
		return runRecordRequestEnd(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "flow: unknown record request command %q\n", args[0])
		fmt.Fprint(stderr, recordUsage)
		return 2
	}
}

// requireRequestN refuses an -n that is not a positive integer.
func requireRequestN(n int, stderr io.Writer) bool {
	if n <= 0 {
		fmt.Fprintf(stderr, "flow: -n must be a positive integer, got %d\n", n)
		fmt.Fprint(stderr, recordUsage)
		return false
	}
	return true
}

// runRecordRequestBegin implements `flow record request begin`: a
// request's receipt, or an amendment of its title or brainstorm reason.
func runRecordRequestBegin(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow record request begin", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordIdentityFlags(fset, &f)
	n := fset.Int("n", 0, "the request's number in the session, a positive integer (required)")
	title := fset.String("title", "", "the request in one line")
	reason := fset.String("brainstorm-reason", "", "why the request needs a brainstorm")
	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
		return code
	}
	if !requireRequestN(*n, stderr) {
		return 2
	}
	if *title == "" && *reason == "" {
		fmt.Fprintln(stderr, "flow: -title or -brainstorm-reason is required")
		fmt.Fprint(stderr, recordUsage)
		return 2
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}
	now := time.Now().UTC()
	in := records.RequestBegin{N: *n, Title: *title, BrainstormReason: *reason, ReceivedAt: &now}
	_, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.Request, error) {
		return cl.BeginRequest(ctx, projectKey, f.change, in)
	})
	if callErr == nil {
		fmt.Fprintf(stdout, "recorded: request %d\n", *n)
	}
	return classifyRecordWrite(callErr, projectKey, f.change, "request-begin", in, stderr)
}

// runRecordRequestEnd implements `flow record request end`: a request's
// outcome. landed takes -commit and -base and runs gateLandedRequest
// before the write; answered and dropped take neither.
func runRecordRequestEnd(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow record request end", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f recordIdentityFlags
	registerRecordIdentityFlags(fset, &f)
	n := fset.Int("n", 0, "the request's number in the session, a positive integer (required)")
	outcome := fset.String("outcome", "", "landed, answered or dropped (required)")
	commit := fset.String("commit", "", "landed only: the request's last landed commit, a full sha")
	base := fset.String("base", "", "landed only: origin/<base> right after the request's first landing rebase, a full sha")
	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
		return code
	}
	if !requireRequestN(*n, stderr) {
		return 2
	}
	if !slices.Contains(requestOutcomes, *outcome) {
		fmt.Fprintf(stderr, "flow: -outcome %q must be one of %s\n", *outcome, strings.Join(requestOutcomes, ", "))
		fmt.Fprint(stderr, recordUsage)
		return 2
	}
	if *outcome == "landed" {
		if !requireRecordFlags(stderr, [2]string{"-commit", *commit}, [2]string{"-base", *base}) {
			return 2
		}
	} else if *commit != "" || *base != "" {
		fmt.Fprintf(stderr, "flow: -outcome %s takes neither -commit nor -base\n", *outcome)
		fmt.Fprint(stderr, recordUsage)
		return 2
	}
	for _, sha := range [][2]string{{"-commit", *commit}, {"-base", *base}} {
		if err := validateCommitSHA(sha[0], sha[1]); err != nil {
			fmt.Fprintf(stderr, "flow: %v\n", err)
			return 2
		}
	}

	projectKey, _, err := fallback.ProjectKey(f.dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}
	if *outcome == "landed" {
		if code := gateLandedRequest(ctx, f, projectKey, *n, *commit, *base, stderr); code != 0 {
			return code
		}
	}
	now := time.Now().UTC()
	in := records.RequestEnd{N: *n, Outcome: *outcome, CommitSHA: *commit, BaseSHA: *base, EndedAt: &now}
	_, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.Request, error) {
		return cl.EndRequest(ctx, projectKey, f.change, in)
	})
	if callErr == nil {
		fmt.Fprintf(stdout, "recorded: request %d %s\n", *n, *outcome)
	}
	return classifyRecordWrite(callErr, projectKey, f.change, "request-end", in, stderr)
}

// gateLandedRequest is design.md's Request-end guard, run in-process before
// the write and before the store is contacted for anything else: the
// change's checkout (the first linked worktree carrying .superpowers/sdd/<change>/,
// else the main checkout when it carries it), commit an ancestor of
// origin/<base> and base an ancestor of commit, the request's latest
// /flow-interactive flow.verify run ended completed, and -- when
// check-visual-trigger matches base..commit's files -- its latest
// flow.visual-verify run ended completed, at least one second after it
// began. A store that cannot be reached
// skips both stage-run reads, with one ⚠ line. 0 passes; 1
// refuses with one line naming the failed check.
func gateLandedRequest(ctx context.Context, f recordIdentityFlags, projectKey string, n int, commit, base string, stderr io.Writer) int {
	refuse := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "flow: record request end refused: "+format+"\n", a...)
		return 1
	}
	wt := ""
	if main, carrying, err := changeCheckouts(f.dir, f.change); err == nil {
		for _, c := range carrying {
			if fi, err := os.Stat(filepath.Join(c, ".superpowers/sdd", f.change)); err != nil || !fi.IsDir() {
				continue
			}
			if c != main {
				wt = c
				break
			}
			if wt == "" {
				wt = c
			}
		}
	}
	if wt == "" {
		return refuse("no checkout of %s carries .superpowers/sdd/%s/ — pass -C the session's worktree", f.dir, f.change)
	}

	env := guard.Env{Getenv: os.Getenv, Dir: wt}
	var out, errBuf bytes.Buffer
	if code := guard.Registry["resolve-base-branch"]([]string{wt}, env, &out, &errBuf); code != 0 {
		return refuse("resolve-base-branch exited %d on %s: %s", code, wt, strings.TrimSpace(errBuf.String()))
	}
	branch := strings.TrimSpace(out.String())
	if exec.Command("git", "-C", wt, "merge-base", "--is-ancestor", commit, "origin/"+branch).Run() != nil {
		return refuse("-commit %s is not on origin/%s — land it first", commit, branch)
	}
	if exec.Command("git", "-C", wt, "merge-base", "--is-ancestor", base, commit).Run() != nil {
		return refuse("-base %s is not an ancestor of -commit %s", base, commit)
	}
	files, err := exec.Command("git", "-C", wt, "diff", "--name-only", base+".."+commit).Output()
	if err != nil {
		return refuse("git diff --name-only %s..%s failed in %s: %v", base, commit, wt, err)
	}

	verify, err := latestRequestRun(ctx, f, projectKey, n, "flow.verify")
	if errors.Is(err, client.ErrUnavailable) && !errors.Is(err, client.ErrDaemonAnswered) {
		fmt.Fprintf(stderr, "⚠ flow: store unreachable — skipped request %d's verify and UI check reads; the git checks passed\n", n)
		return 0
	}
	if err != nil {
		return refuse("read request %d's verify runs: %v", n, err)
	}
	switch {
	case verify == nil:
		return refuse("request %d has no verify (flow.verify run carrying -request %d)", n, n)
	case verify.Outcome == nil || *verify.Outcome != stages.OutcomeCompleted:
		return refuse("request %d's latest verify (flow.verify) is %s, not completed", n, runOutcome(verify))
	}

	out.Reset()
	errBuf.Reset()
	env.Stdin = bytes.NewReader(files)
	switch code := guard.Registry["check-visual-trigger"]([]string{wt}, env, &out, &errBuf); {
	case code == 1:
		return 0
	case code == 2 && strings.Contains(errBuf.String(), "VISUAL-TRIGGER-NOT-CONFIGURED"):
		return 0
	case code != 0:
		return refuse("check-visual-trigger cannot answer: %s", strings.TrimSpace(errBuf.String()))
	}

	latest, err := latestRequestRun(ctx, f, projectKey, n, "flow.visual-verify")
	if errors.Is(err, client.ErrUnavailable) && !errors.Is(err, client.ErrDaemonAnswered) {
		fmt.Fprintf(stderr, "⚠ flow: store unreachable — skipped request %d's UI check read; the git checks and the verify read passed\n", n)
		return 0
	}
	if err != nil {
		return refuse("read request %d's UI check runs: %v", n, err)
	}
	switch {
	case latest == nil:
		return refuse("request %d touched `ui paths` and has no UI check (flow.visual-verify run carrying -request %d)", n, n)
	case latest.Outcome == nil || *latest.Outcome != stages.OutcomeCompleted:
		return refuse("request %d's latest UI check (flow.visual-verify) is %s, not completed — fix, re-check and end it completed", n, runOutcome(latest))
	}
	ended := ""
	if latest.EndedAt != nil {
		ended = *latest.EndedAt
	}
	endAt, err := time.Parse(time.RFC3339Nano, ended)
	if err != nil {
		return refuse("request %d's latest UI check (flow.visual-verify) carries an unreadable endedAt %q", n, ended)
	}
	startAt, _ := time.Parse(time.RFC3339Nano, latest.StartedAt) // parsed by latestRequestRun
	if took := endAt.Sub(startAt); took < time.Second {
		return refuse("request %d's latest UI check (flow.visual-verify) ended %s after its begin — a chained mark (step 7); end it after the screenshot is read", n, took)
	}
	return 0
}

// latestRequestRun reads the change's /flow-interactive runs of stage and
// returns request n's latest by start, nil when it has none.
func latestRequestRun(ctx context.Context, f recordIdentityFlags, projectKey string, n int, stage string) (*client.StageRunRow, error) {
	runs, err := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) ([]client.StageRunRow, error) {
		return cl.ListStageRuns(ctx, projectKey, f.change, string(stages.FlowInteractive), stage)
	})
	if err != nil {
		return nil, err
	}
	var latest *client.StageRunRow
	var latestAt time.Time
	for i, r := range runs {
		if r.Request == nil || *r.Request != n {
			continue
		}
		at, err := time.Parse(time.RFC3339Nano, r.StartedAt)
		if err != nil {
			return nil, fmt.Errorf("a %s run carries an unreadable startedAt %q", stage, r.StartedAt)
		}
		if latest == nil || at.After(latestAt) {
			latest, latestAt = &runs[i], at
		}
	}
	return latest, nil
}

// runOutcome is a run's outcome, "open" while it has none.
func runOutcome(r *client.StageRunRow) string {
	if r.Outcome == nil {
		return "open"
	}
	return *r.Outcome
}
