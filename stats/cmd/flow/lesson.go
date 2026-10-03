package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/lessons"
)

const lessonUsage = `usage: flow lesson resolve [-addr url] [-timeout dur] -topic <words>

Prints the process-lesson answer for one topic on stdout -- every
registered project's docs/briefs/ briefs and archived narratives matching
the words, briefs ranked first and rendered in full, assembled by the
daemon. The CLI constructs none of it, the record-render rule: the
answer's shape is decided where it is produced.

The address resolves from FLOW_RECORDS_ADDR when it is set, then FLOW_ADDR,
the record family's own resolution: the answer reads the persistent
workspace store, so it is the same from an apply worktree as from the main
checkout.

This is a read, with the findings read's contract: a store that cannot be
reached is reported to stderr and exits non-zero -- never a partial answer
on stdout that a caller could mistake for the workspace's own. The only
other non-zero exits are caller mistakes: a missing -topic, a -C (the
answer spans every project, so there is no project to resolve), or a stray
positional argument.
`

// runLesson implements `flow lesson`. One subcommand today: `resolve`,
// the read a run uses to resolve a practice a ticket names instead of
// searching repositories for it.
func runLesson(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, lessonUsage)
		return 2
	}
	switch args[0] {
	case "resolve":
		return runLessonResolve(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "flow: unknown lesson command %q\n", args[0])
		fmt.Fprint(stderr, lessonUsage)
		return 2
	}
}

func runLessonResolve(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow lesson resolve", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var (
		f     recordIdentityFlags
		topic string
	)
	// The record family's one seam, so this read resolves FLOW_RECORDS_ADDR
	// exactly where every other record verb does. The seam also registers
	// -C, which resolves a project key this read never uses -- the answer
	// spans every registered project -- so a -C given is refused below
	// rather than accepted as a flag that does nothing.
	registerRecordConnFlags(fset, &f)
	fset.Lookup("C").Usage = "refused: the answer spans every registered project"
	fset.StringVar(&topic, "topic", "", "the words a ticket uses to name the practice or brief (required)")

	if err := fset.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, lessonUsage)
		return 2
	}
	noteAddrUsage(fset, stderr, f.addr)
	cGiven := false
	fset.Visit(func(fl *flag.Flag) { cGiven = cGiven || fl.Name == "C" })
	if cGiven {
		fmt.Fprint(stderr, "flow: lesson resolve takes no -C\n")
		fmt.Fprint(stderr, lessonUsage)
		return 2
	}
	if fset.NArg() != 0 {
		fmt.Fprint(stderr, "flow: expected no positional arguments\n")
		fmt.Fprint(stderr, lessonUsage)
		return 2
	}
	// The daemon's own rule — empty after normalization — so "  " or "-"
	// is refused here as the caller mistake it is, not ridden to the store
	// to come back as a read failure.
	if lessons.Normalize(topic) == "" {
		fmt.Fprint(stderr, "flow: -topic is required\n")
		fmt.Fprint(stderr, lessonUsage)
		return 2
	}

	answer, err := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) ([]byte, error) {
		return cl.ResolveLesson(ctx, topic)
	})
	if err != nil {
		fmt.Fprintf(stderr, "flow: lesson resolve: %v\n", err)
		return 1
	}
	_, _ = stdout.Write(answer)
	return 0
}
