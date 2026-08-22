package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/reconcile"
	"github.com/tweety53/agents/stats/internal/store"
)

// defaultJournalDSN mirrors internal/config.DefaultDSN -- named again here
// rather than imported, for the same reason defaultAddr in state.go
// duplicates internal/config's address default (see that constant's own
// doc comment). `journal flush` is the one CLI command that departs from
// this package's usual "never SQL, never the daemon's internals" posture
// (this file's own package doc comment in main.go): it is an explicit,
// on-demand operator action rather than a pipeline write on the hot path,
// so unlike `state get`/`state set` it talks to the store directly through
// internal/store and internal/reconcile instead of going through myflowd's
// HTTP API and falling back to disk on any failure.
const defaultJournalDSN = "postgres://myflow:myflow@localhost:5433/myflow?sslmode=disable"

// defaultJournalConnectTimeout bounds how long `journal flush` waits to
// open a connection to the store before giving up. This is allowed to be
// longer than state get/set's never-block-the-pipeline timeout (2s,
// defaultTimeout in state.go) precisely because this command is not on
// that hot path -- it is an operator asking, right now, "apply whatever is
// pending", and it is allowed to take a little while rather than fail fast
// against a database that is merely slow to accept a new connection.
const defaultJournalConnectTimeout = 5 * time.Second

const journalUsage = `usage: myflow journal flush [-root dir] [-dsn url] [-timeout dur]

flush replays every pending journal entry found under root into the store,
connecting to it directly (not through myflowd's HTTP API).
`

func runJournal(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, journalUsage)
		return 2
	}

	switch args[0] {
	case "flush":
		return runJournalFlush(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "myflow: unknown journal command %q\n", args[0])
		fmt.Fprint(stderr, journalUsage)
		return 2
	}
}

// runJournalFlush implements `myflow journal flush`: it connects to the
// store, replays every journal file under -root (default
// fallback.StateRoot()), and reports the outcome on stdout. Unlike `state
// get`/`state set`, this command does not have a fallback of its own to
// take on failure -- there is nothing to fall back *to* for an explicit
// "reconcile now" request -- so a connection failure or a replay error is
// reported and this exits non-zero.
func runJournalFlush(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("myflow journal flush", flag.ContinueOnError)
	fset.SetOutput(stderr)
	root := fset.String("root", fallback.StateRoot(), "root of every project's state directory")
	dsn := fset.String("dsn", defaultJournalDSN, "store connection string")
	timeout := fset.Duration("timeout", defaultJournalConnectTimeout, "connection timeout")
	if err := fset.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "myflow: %v\n", err)
		fmt.Fprint(stderr, journalUsage)
		return 2
	}
	if fset.NArg() != 0 {
		fmt.Fprint(stderr, "myflow: journal flush takes no positional arguments\n")
		fmt.Fprint(stderr, journalUsage)
		return 2
	}

	connectCtx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()
	st, err := store.Open(connectCtx, *dsn)
	if err != nil {
		fmt.Fprintf(stderr, "myflow: journal flush: connect to store: %v\n", err)
		return 1
	}
	defer st.Close()

	reconciler := reconcile.New(st, st, st, *root, slog.New(slog.NewTextHandler(stderr, nil)))
	result, err := reconciler.Run(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "myflow: journal flush: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "journal flush: %d journal(s) visited, %d applied, %d refused\n",
		result.Journals, result.Applied, result.Refused)
	return 0
}
