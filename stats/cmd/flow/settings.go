package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
)

const settingsUsage = `usage: flow settings get [-addr url] [-timeout dur]
       flow settings set [-addr url] [-timeout dur] -model name -reviewers a,b,c

settings get prints the harness-wide settings record (default model and
reviewer slots) as one line of JSON.

settings set writes the whole record, replacing whatever was recorded
before. Unlike state/stage's never-block-on-store-failure pattern, a
rejected -model or -reviewers value is a caller mistake, not a store
failure: there is no fallback value to record for an invalid one, so this
prints the store's rejection reason and exits non-zero.
`

func runSettings(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, settingsUsage)
		return 2
	}

	switch args[0] {
	case "get":
		return runSettingsGet(ctx, args[1:], stdout, stderr)
	case "set":
		return runSettingsSet(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "flow: unknown settings command %q\n", args[0])
		fmt.Fprint(stderr, settingsUsage)
		return 2
	}
}

// settingsConnFlags is the flag.FlagSet common to `settings get` and
// `settings set`: just the store connection, since flow_settings holds
// one harness-wide row -- neither subcommand resolves a project key the
// way `state`/`stage` do.
type settingsConnFlags struct {
	addr    string
	timeout time.Duration
}

func registerSettingsConnFlags(fset *flag.FlagSet, f *settingsConnFlags) {
	fset.StringVar(&f.addr, "addr", resolveDefaultAddr(), "flowd base URL")
	fset.DurationVar(&f.timeout, "timeout", defaultTimeout, "store request timeout")
}

// runSettingsGet implements `flow settings get`. Unlike `state get`,
// a store failure here is reported and exits non-zero rather than falling
// back to a local file: there is no per-change on-disk fallback for a
// harness-wide record, and this task adds none.
func runSettingsGet(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow settings get", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f settingsConnFlags
	registerSettingsConnFlags(fset, &f)
	if err := fset.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, settingsUsage)
		return 2
	}
	noteAddrEnvUsage(fset, stderr)
	if fset.NArg() != 0 {
		fmt.Fprintln(stderr, "flow: settings get takes no positional arguments")
		fmt.Fprint(stderr, settingsUsage)
		return 2
	}

	s, err := getSettings(ctx, f.addr, f.timeout)
	if err != nil {
		fmt.Fprintf(stderr, "flow: settings get: %v\n", err)
		return 1
	}
	return writeSettingsOutput(stdout, stderr, s)
}

// runSettingsSet implements `flow settings set`. A successful write
// prints the record the store echoes back and exits 0. A rejection
// (client.ErrSettingsRejected -- the API's 400, naming the bad value) is a
// caller mistake: it is printed and this exits 1, never silently falling
// back, per this task's own instruction. Every other failure -- the store
// unreachable, a malformed response -- is likewise reported and exits
// non-zero: unlike `state set`, there is no fallback value to record for
// settings that were never successfully validated.
func runSettingsSet(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow settings set", flag.ContinueOnError)
	fset.SetOutput(stderr)
	var f settingsConnFlags
	registerSettingsConnFlags(fset, &f)
	model := fset.String("model", "", "the default model, e.g. sonnet (required)")
	reviewers := fset.String("reviewers", "", "comma-separated reviewer slots, e.g. primary,principles,code-review-low (required)")
	if err := fset.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		fmt.Fprint(stderr, settingsUsage)
		return 2
	}
	noteAddrEnvUsage(fset, stderr)
	if fset.NArg() != 0 {
		fmt.Fprintln(stderr, "flow: settings set takes no positional arguments")
		fmt.Fprint(stderr, settingsUsage)
		return 2
	}
	if *model == "" || *reviewers == "" {
		fmt.Fprintln(stderr, "flow: -model and -reviewers are both required")
		fmt.Fprint(stderr, settingsUsage)
		return 2
	}

	in := client.Settings{DefaultModel: *model, Reviewers: strings.Split(*reviewers, ",")}
	s, err := putSettings(ctx, f.addr, f.timeout, in)
	switch {
	case err == nil:
		return writeSettingsOutput(stdout, stderr, s)
	case errors.Is(err, client.ErrSettingsRejected):
		fmt.Fprintf(stderr, "flow: settings set refused: %v\n", err)
		return 1
	default:
		fmt.Fprintf(stderr, "flow: settings set: %v\n", err)
		return 1
	}
}

// writeSettingsOutput encodes s as one line of JSON to stdout, mirroring
// `state get`/`state list`'s own one-line-of-JSON output shape.
func writeSettingsOutput(stdout, stderr io.Writer, s client.Settings) int {
	encoded, err := json.Marshal(s)
	if err != nil {
		fmt.Fprintf(stderr, "flow: settings: encode output: %v\n", err)
		return 1
	}
	_, _ = stdout.Write(encoded)
	fmt.Fprintln(stdout)
	return 0
}

// getSettings calls the store's GET /api/v1/settings endpoint under
// addr/timeout, recovering from any panic in the client path and
// reporting it as client.ErrUnavailable -- the same guarantee
// getChange/putChange's own recover provides.
func getSettings(ctx context.Context, addr string, timeout time.Duration) (s client.Settings, err error) {
	defer func() {
		if r := recover(); r != nil {
			s, err = client.Settings{}, fmt.Errorf("%w: recovered panic: %v", client.ErrUnavailable, r)
		}
	}()

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cl := client.New(addr, &http.Client{Timeout: timeout})
	return cl.GetSettings(reqCtx)
}

// putSettings calls the store's PUT /api/v1/settings endpoint under
// addr/timeout, recovering from any panic in the client path exactly as
// getSettings does.
func putSettings(ctx context.Context, addr string, timeout time.Duration, in client.Settings) (s client.Settings, err error) {
	defer func() {
		if r := recover(); r != nil {
			s, err = client.Settings{}, fmt.Errorf("%w: recovered panic: %v", client.ErrUnavailable, r)
		}
	}()

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cl := client.New(addr, &http.Client{Timeout: timeout})
	return cl.PutSettings(reqCtx, in)
}
