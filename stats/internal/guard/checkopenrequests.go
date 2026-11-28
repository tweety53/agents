package guard

// checkopenrequests.go — `flow-guard check-open-requests <worktree>`: every
// /flow-interactive request the change records that was begun and never ended
// (no outcome), read from the records flowd with the flow CLI's own client.
// The change is the worktree's base name, as await-rows derives it.
//
// Exit 0 when none is open; 1 listing each open one as `OPEN <n> <title>`, in
// number order; 2 on usage or an unreadable store.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/guard/core"
)

func init() { Registry["check-open-requests"] = checkOpenRequests }

func checkOpenRequests(args []string, env core.Env, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		core.Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard check-open-requests <worktree>")
		return 2
	}
	worktree := args[0]
	change := filepath.Base(worktree)
	project, _, err := fallback.ProjectKey(core.Abs(env, worktree))
	if err != nil {
		core.Reason(env, "store-unreadable")
		fmt.Fprintf(stderr, "check-open-requests: %v\n", err)
		return 2
	}
	const timeout = 2 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	reqs, err := client.New(recordsAddr(env), &http.Client{Timeout: timeout}).ListRequests(ctx, project, change)
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		core.Reason(env, "store-unreadable")
		fmt.Fprintf(stderr, "check-open-requests: read %s's requests: %v\n", change, err)
		return 2
	}
	open := 0
	for _, r := range reqs {
		if r.Outcome == "" {
			open++
			fmt.Fprintf(stdout, "OPEN %d %s\n", r.N, r.Title)
		}
	}
	if open > 0 {
		core.Reason(env, "requests-open")
		return 1
	}
	return 0
}
