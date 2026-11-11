package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
)

const policyUsage = `usage: flow policy get    [-addr url] [-timeout dur]
       flow policy grants [-addr url] [-timeout dur] -change <name>
       flow policy grant  [-addr url] [-timeout dur] -change <name> -roles <role>[,<role>...]
       flow policy revoke [-addr url] [-timeout dur] -change <name> -roles <role>[,<role>...]

Every role's allowed pairs live in flowd's store (kan-1103); the Experiments
view edits them. get prints them as one JSON object,
{role: {base: [...], optin: [...], marked: [{pair, marker}]}}, and caches
every good answer at <state dir>/policy-cache.json; with flowd unreachable
it prints that cache, one warning on stderr. grants prints the JSON list of
roles <name> holds a live grant for. grant records the operator's say-so
that <name>'s dispatches of each role may run its opt-in pairs (each
carrying the line ` + "`Operator pair: <role>`" + `); revoke removes it, and a
grant lapses when the change reaches FINISHED.

Exit codes: 0 done; 1 the store refused, or could not be reached (get:
and no cache exists); 2 usage.
`

// policyCachePath is where `flow policy get` keeps its last good answer.
func policyCachePath() string { return filepath.Join(fallback.StateRoot(), "policy-cache.json") }

func runPolicy(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, policyUsage)
		return 2
	}
	var f storeConnFlags
	fset := flag.NewFlagSet("flow policy "+args[0], flag.ContinueOnError)
	registerStoreConnFlags(fset, &f)
	change := fset.String("change", "", "the change the grant is for")
	roles := fset.String("roles", "", "comma-separated role ids")
	switch args[0] {
	case "get", "grants", "grant", "revoke":
	default:
		fmt.Fprintf(stderr, "flow: unknown policy command %q\n", args[0])
		fmt.Fprint(stderr, policyUsage)
		return 2
	}
	fset.SetOutput(stderr)
	if err := fset.Parse(args[1:]); err != nil || fset.NArg() != 0 {
		fmt.Fprint(stderr, policyUsage)
		return 2
	}
	noteAddrUsage(fset, stderr, f.addr)
	needChange := args[0] != "get"
	needRoles := args[0] == "grant" || args[0] == "revoke"
	if needChange && *change == "" || needRoles && *roles == "" || !needChange && (*change != "" || *roles != "") {
		fmt.Fprintf(stderr, "flow: policy %s: wrong flags\n", args[0])
		fmt.Fprint(stderr, policyUsage)
		return 2
	}
	switch args[0] {
	case "get":
		return runPolicyGet(ctx, f, stdout, stderr)
	case "grants":
		list, err := withFlowdClient(ctx, f, func(ctx context.Context, cl *client.Client) ([]string, error) {
			return cl.Grants(ctx, *change)
		})
		if err != nil {
			fmt.Fprintf(stderr, "flow: policy grants: %v\n", err)
			return 1
		}
		if list == nil {
			list = []string{}
		}
		b, _ := json.Marshal(list)
		fmt.Fprintf(stdout, "%s\n", b)
		return 0
	default:
		revoke := args[0] == "revoke"
		_, err := withFlowdClient(ctx, f, func(ctx context.Context, cl *client.Client) (struct{}, error) {
			return struct{}{}, cl.Grant(ctx, *change, strings.Split(*roles, ","), revoke)
		})
		if err != nil {
			fmt.Fprintf(stderr, "flow: policy %s: %v\n", args[0], err)
			return 1
		}
		fmt.Fprintf(stdout, "policy %s %s %s\n", args[0], *change, *roles)
		return 0
	}
}

// runPolicyGet prints the store's pairs and caches them; with the store
// unreachable it prints the cache instead (decision policy-cache), and with
// neither it exits 1 naming both.
func runPolicyGet(ctx context.Context, f storeConnFlags, stdout, stderr io.Writer) int {
	cache := policyCachePath()
	p, err := withFlowdClient(ctx, f, func(ctx context.Context, cl *client.Client) (json.RawMessage, error) {
		return cl.Policy(ctx)
	})
	if err == nil {
		if werr := writePolicyCache(cache, p); werr != nil {
			fmt.Fprintf(stderr, "flow: policy get: could not write the cache: %v\n", werr)
		}
		fmt.Fprintf(stdout, "%s\n", p)
		return 0
	}
	b, rerr := os.ReadFile(cache)
	if rerr != nil || !json.Valid(b) {
		fmt.Fprintf(stderr, "flow: policy get: flowd unreachable (%v) and no usable cache at %s\n", err, cache)
		return 1
	}
	fmt.Fprintf(stderr, "flow: policy get: flowd unreachable (%v); answering from the cache at %s\n", err, cache)
	fmt.Fprintf(stdout, "%s\n", strings.TrimSpace(string(b)))
	return 0
}

// writePolicyCache replaces the cache atomically, so a reader never sees a
// half-written file.
func writePolicyCache(path string, p []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".policy-cache-*")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(p); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), path)
}
