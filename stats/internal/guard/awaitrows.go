package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/hook"
)

// flow-guard await-rows <worktree> <key> [<key>…] [-any] [-timeout <s>] — the parent's one bounded
// wait on its children: for <worktree>'s change, inside this one call, runs the
// end-stale-dispatch-rows sweep and reads the dispatch rows, until every named key's row is ended
// — one `ENDED <key> <outcome>` line per key, in argument order — or the timeout (default 540, max
// 570), which prints the ended keys' lines and `WAITING <keys still open>`, exit 0. With -any it
// returns once any named key's row is ended, the same lines: the ended keys' ENDED lines and
// WAITING with the rest. A key the change does not record exits 2 (`unknown-key`); an unreadable
// store exits 1. The rows are read straight from the records flowd (FlowdDispatches). Between
// reads it long-polls that flowd's GET
// /api/v1/records/{project}/{change}/dispatches/wait, which answers the moment a row closes, for
// up to 20 s at a time; a flowd that is unreachable or does not answer that route with a JSON
// array leaves it polling every 5 s. The time blocked between reads is the run's wait (RecordWait),
// posted with its telemetry as wait_ms. skills/flow/implement.md's **Turn discipline** is the
// contract.
//
// Go, in flow-guard (stats/internal/guard/awaitrows.go).
func init() { Registry["await-rows"] = awaitRows }

const (
	awaitRowsPoll     = 5 * time.Second
	awaitRowsLongPoll = 20 * time.Second
)

func awaitRows(args []string, env core.Env, stdout, stderr io.Writer) int {
	anyMode := slices.Contains(args, "-any")
	timeout, rest, ok := awaitTimeout(slices.DeleteFunc(slices.Clone(args), func(a string) bool { return a == "-any" }))
	if !ok || len(rest) < 2 {
		core.Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard await-rows <worktree> <key> [<key>…] [-any] [-timeout <seconds, max 570>]")
		return 2
	}
	worktree, keys := rest[0], rest[1:]
	change := filepath.Base(worktree)
	waitURL := awaitRowsWaitURL(env, worktree, change, keys, anyMode)
	deadline := time.Now().Add(timeout)
	for {
		hook.EndStaleDispatchRows([]string{worktree}, env, io.Discard, stderr)
		rows, err := core.FlowdDispatches(env, worktree, change)
		if err != nil {
			core.Reason(env, "store-unreadable")
			fmt.Fprintf(stderr, "await-rows: dispatches unreadable: %v\n", err)
			return 1
		}
		outcome := map[string]string{}
		ended := map[string]bool{}
		for _, r := range rows {
			outcome[r.Key] = r.Outcome
			ended[r.Key] = r.EndedAt != nil
		}
		var open []string
		for _, k := range keys {
			if _, ok := outcome[k]; !ok {
				core.Reason(env, "unknown-key")
				fmt.Fprintf(stderr, "await-rows: %s records no dispatch row %s\n", change, k)
				return 2
			}
			if !ended[k] {
				open = append(open, k)
			}
		}
		if len(open) == 0 || (anyMode && len(open) < len(keys)) || !time.Now().Before(deadline) {
			for _, k := range keys {
				if ended[k] {
					fmt.Fprintf(stdout, "ENDED %s %s\n", k, outcome[k])
				}
			}
			if len(open) > 0 {
				fmt.Fprintf(stdout, "WAITING %s\n", strings.Join(open, " "))
			}
			return 0
		}
		awaitRowsBlock(env, waitURL, time.Until(deadline))
	}
}

// awaitRowsBlock blocks between two reads for at most left: a long-poll on waitURL, else a sleep
// of one poll interval, the time blocked recorded as the run's wait.
func awaitRowsBlock(env core.Env, waitURL string, left time.Duration) {
	blocked := time.Now()
	if !awaitRowsLongPollOnce(waitURL, min(awaitRowsLongPoll, left)) {
		time.Sleep(min(awaitRowsPoll, left))
	}
	core.RecordWait(env, time.Since(blocked))
}

// awaitRowsWaitURL is flowd's long-poll route for keys of change, without its timeout; "" when
// worktree yields no project key, which leaves await-rows polling.
func awaitRowsWaitURL(env core.Env, worktree, change string, keys []string, anyMode bool) string {
	project, _, err := fallback.ProjectKey(worktree)
	if err != nil {
		return ""
	}
	mode := "all"
	if anyMode {
		mode = "any"
	}
	q := url.Values{"keys": {strings.Join(keys, ",")}, "mode": {mode}}
	return core.RecordsAddr(env) + "/api/v1/records/" + url.PathEscape(project) + "/" + url.PathEscape(change) +
		"/dispatches/wait?" + q.Encode()
}

// awaitRowsLongPollOnce blocks on waitURL for up to d (whole seconds, at least one), true when
// flowd answered 200 with a JSON array — the caller re-reads the rows then; false sends it to a
// poll sleep.
func awaitRowsLongPollOnce(waitURL string, d time.Duration) bool {
	if waitURL == "" {
		return false
	}
	secs := max(1, int(math.Ceil(d.Seconds())))
	hc := &http.Client{Timeout: time.Duration(secs)*time.Second + 10*time.Second}
	resp, err := hc.Get(waitURL + "&timeout=" + strconv.Itoa(secs))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	var rows []json.RawMessage
	return resp.StatusCode == http.StatusOK && json.NewDecoder(resp.Body).Decode(&rows) == nil
}
