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

// jiraTransitionTimeout is `flow jira transition`'s default request
// budget, and deliberately not the 2s defaultTimeout the other commands
// use: the retry budget that rides out a transient Atlassian outage lives
// daemon-side (internal/jira), so the daemon's honest answer to a
// transient failure can take tens of seconds, and a 2s client cutoff
// would abort the command precisely while flowd is still working. 30s
// covers the daemon's whole ladder with headroom under its own 30s
// writeTimeout cutoff.
const jiraTransitionTimeout = 30 * time.Second

// jiraUsage is `flow jira`'s own usage block, in the shape settingsUsage
// states its subcommands: flags first, then the one paragraph a caller
// needs about failure semantics and the vocabulary <target> accepts.
var jiraUsage = `flow jira transition  [-addr url] [-timeout dur] <KEY> <target>

flow jira transition moves a Jira issue to a pipeline position: the
four-position forward-only table (To Do, In Progress, In Review, Done),
resolved and fired by flowd -- with retries against a transient Atlassian
outage -- rather than in the calling session (KAN-571). <target> is a
position name in any of its usual spellings: "In Progress", "in-review"
and "Code Review" all resolve. An issue already at or past the target is
a success reported as "already <status> (no transition)"; nothing ever
moves backward.

A failure exits 1 with the reason on stderr: the pipeline prints it as
its one "Jira: skipped -- <reason>" line and carries on, per the
never-blocking Jira contract. Exit 2 is a caller mistake -- a missing
argument or a target that matches no pipeline position.
`

// runJira routes `flow jira`'s verbs. transition is the only one today;
// the switch keeps the shape every other multi-verb command uses, so the
// next verb is one case, not a rewrite.
func runJira(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, jiraUsage)
		return 2
	}
	switch args[0] {
	case "transition":
		return runJiraTransition(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "flow: unknown jira subcommand %q\n", args[0])
		fmt.Fprint(stderr, jiraUsage)
		return 2
	}
}

// jiraConnFlags is the flag.FlagSet `jira transition` takes: just the
// store connection, since the issue key and target are positional -- a
// pipeline call reads `flow jira transition KAN-571 "In Review"`, not a
// flag per part.
type jiraConnFlags struct {
	addr    string
	timeout time.Duration
}

func registerJiraConnFlags(fset *flag.FlagSet, f *jiraConnFlags) {
	fset.StringVar(&f.addr, "addr", resolveDefaultAddr(), "flowd base URL")
	fset.DurationVar(&f.timeout, "timeout", jiraTransitionTimeout, "request budget; the daemon retries transient outages within it")
}

// runJiraTransition implements `flow jira transition`. The retry budget
// lives daemon-side (internal/jira), so the CLI's own -timeout only
// bounds the one HTTP round trip to flowd, exactly as every other
// command's does -- a transient Atlassian outage is ridden out by the
// daemon, not by the caller retrying the command.
func runJiraTransition(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow jira transition", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f jiraConnFlags
	registerJiraConnFlags(fset, &f)
	if err := fset.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, jiraUsage)
		return 2
	}
	noteAddrUsage(fset, stderr, f.addr)
	if fset.NArg() != 2 {
		fmt.Fprintln(stderr, "flow: jira transition takes exactly two arguments, the issue key and the target position")
		fmt.Fprint(stderr, jiraUsage)
		return 2
	}
	key, target := fset.Arg(0), fset.Arg(1)

	reqCtx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()
	cl := client.New(f.addr, &http.Client{Timeout: f.timeout})
	result, err := cl.JiraTransition(reqCtx, key, target)
	if err != nil {
		fmt.Fprintf(stderr, "flow: jira transition %s: %v\n", key, err)
		return 1
	}
	if result.Moved {
		fmt.Fprintf(stdout, "Jira: %s → %s\n", result.Key, result.Status)
	} else {
		fmt.Fprintf(stdout, "Jira: %s already %s (no transition)\n", result.Key, result.Status)
	}
	return 0
}
