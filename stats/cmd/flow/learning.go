package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
)

const learningUsage = `usage: flow learning add <topic> -note '<one line>' -change <name> [-C dir] [-addr url] [-timeout dur]
       flow learning list [-addr url] [-timeout dur]

learning add records one learning note shown by a run: <topic> is a
lowercase kebab slug, the project key resolves from -C (default: the
working directory). It exits 0 on write, 1 when the store refuses the
topic -- malformed, or already shown three times -- or cannot be reached,
and 2 on a usage error.

learning list prints every recorded topic with its count and newest note
as one JSON array line, exiting 1 when the store cannot be reached.

skills/flow-contracts/learning.md is canonical for when a run calls these.
`

func runLearning(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, learningUsage)
		return 2
	}

	switch args[0] {
	case "add":
		return runLearningAdd(ctx, args[1:], stderr)
	case "list":
		return runLearningList(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "flow: unknown learning command %q\n", args[0])
		fmt.Fprint(stderr, learningUsage)
		return 2
	}
}

// runLearningAdd implements `flow learning add`. The topic is the first
// positional argument, ahead of the flags.
func runLearningAdd(ctx context.Context, args []string, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "" || args[0][0] == '-' {
		fmt.Fprintln(stderr, "flow: learning add needs a <topic> first")
		fmt.Fprint(stderr, learningUsage)
		return 2
	}
	topic := args[0]

	fset := flag.NewFlagSet("flow learning add", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f storeConnFlags
	registerStoreConnFlags(fset, &f)
	note := fset.String("note", "", "the note's topic in one line (required)")
	change := fset.String("change", "", "the change the note was shown in (required)")
	dir := fset.String("C", "", "directory the project key resolves from (default: the working directory)")
	if err := fset.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, learningUsage)
		return 2
	}
	noteAddrUsage(fset, stderr, f.addr)
	if fset.NArg() != 0 || *note == "" || *change == "" {
		fmt.Fprintln(stderr, "flow: learning add takes one <topic>, -note and -change")
		fmt.Fprint(stderr, learningUsage)
		return 2
	}
	if *dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(stderr, "flow: resolve working directory: %v\n", err)
			return 1
		}
		*dir = wd
	}
	projectKey, _, err := fallback.ProjectKey(*dir)
	if err != nil {
		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
		return 1
	}

	reqCtx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()
	cl := client.New(f.addr, &http.Client{Timeout: f.timeout})
	in := client.LearningTopic{Topic: topic, Note: *note, Change: *change, ProjectKey: projectKey}
	if err := cl.RecordLearningTopic(reqCtx, in); err != nil {
		fmt.Fprintf(stderr, "flow: learning add: %v\n", err)
		return 1
	}
	return 0
}

// runLearningList implements `flow learning list`.
func runLearningList(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow learning list", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f storeConnFlags
	registerStoreConnFlags(fset, &f)
	if err := fset.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, learningUsage)
		return 2
	}
	noteAddrUsage(fset, stderr, f.addr)
	if fset.NArg() != 0 {
		fmt.Fprintln(stderr, "flow: learning list takes no positional arguments")
		fmt.Fprint(stderr, learningUsage)
		return 2
	}

	reqCtx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()
	cl := client.New(f.addr, &http.Client{Timeout: f.timeout})
	rows, err := cl.ListLearningTopics(reqCtx)
	if err != nil {
		fmt.Fprintf(stderr, "flow: learning list: %v\n", err)
		return 1
	}
	encoded, err := json.Marshal(rows)
	if err != nil {
		fmt.Fprintf(stderr, "flow: learning list: encode output: %v\n", err)
		return 1
	}
	_, _ = stdout.Write(encoded)
	fmt.Fprintln(stdout)
	return 0
}
