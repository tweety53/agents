// Command flow is the thin CLI every flow skill shells out to for
// state. `state get`/`state set` know only HTTP (internal/client) and the
// on-disk fallback format (internal/fallback) -- never SQL, never the
// daemon's internals -- so that fallback path is testable with the daemon
// absent, exactly as design.md's "Boundaries" section requires.
//
// `journal flush` (journal.go) is the one deliberate exception: an
// explicit, on-demand operator command to reconcile the journal right now,
// not a pipeline write on the hot path, so it talks to internal/store and
// internal/reconcile directly rather than through flowd's HTTP API.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
)

const usage = `usage: flow <command> [arguments]

commands:
  state get <name>    print the change's current state
  state set <name>    write the change's whole state, reading it from stdin
  state list          enumerate every change the store holds for this project
  stage begin <name>  record the start of one documented pipeline stage
  stage end <name>    record the end, outcome and metrics of a stage
  record dispatch     record one subagent dispatch of a change's run record
  record finding      record one review-panel finding, or replace it
  record status       set one recorded finding's status
  record findings     print a change's findings as a JSON array
  record render       render a change's run record from the store
  record journal-count  count a change's record writes still pending in the journal
  record cost-status  print how many of a change's dispatches carry no cost figure, and why
  journal flush        replay every pending journal entry into the store
  settings get         print the harness-wide settings record
  settings set         write the harness-wide settings record
  tasks tick <change> <task-id>  flip a task's checkbox and its steps' checkboxes
  workspace-id <name>  print a change's workspace id, derived from its name
`

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	os.Exit(run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run is the whole of main's logic, factored out so tests can drive it
// directly with in-memory streams and a cancellable context -- no
// subprocess needed, per go-cli's "testable main" pattern.
func run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}

	switch args[0] {
	case "state":
		return runState(ctx, args[1:], stdin, stdout, stderr)
	case "stage":
		return runStage(ctx, args[1:], stdout, stderr)
	case "record":
		return runRecord(ctx, args[1:], stdout, stderr)
	case "journal":
		return runJournal(ctx, args[1:], stdout, stderr)
	case "settings":
		return runSettings(ctx, args[1:], stdout, stderr)
	case "tasks":
		return runTasks(ctx, args[1:], stdout, stderr)
	case "workspace-id":
		return runWorkspaceID(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "flow: unknown command %q\n", args[0])
		fmt.Fprint(stderr, usage)
		return 2
	}
}
