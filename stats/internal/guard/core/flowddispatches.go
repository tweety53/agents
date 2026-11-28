package core

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/records"
)

// FlowdDispatches is change's dispatch rows read straight from the records flowd with the flow
// CLI's own client, as `flow record dispatches -change <change> -C <dir>` read them: the project
// key from dir, the CLI's 2s timeout, a change flowd never heard of as no rows, and an
// unreachable flowd an error -- a read has nothing to journal.
func FlowdDispatches(env Env, dir, change string) ([]records.Dispatch, error) {
	project, _, err := fallback.ProjectKey(Abs(env, dir))
	if err != nil {
		return nil, err
	}
	const timeout = 2 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	run, err := client.New(RecordsAddr(env), &http.Client{Timeout: timeout}).GetRunRecord(ctx, project, change)
	if errors.Is(err, client.ErrNotFound) {
		return nil, nil
	}
	return run.Dispatches, err
}

// RecordsAddr is the flowd the run record lives in, as `flow record` resolves it:
// FLOW_RECORDS_ADDR, then FLOW_ADDR, then the dev flowd.
func RecordsAddr(env Env) string {
	if a := env.Getenv("FLOW_RECORDS_ADDR"); a != "" {
		return a
	}
	return FlowAddr(env)
}
