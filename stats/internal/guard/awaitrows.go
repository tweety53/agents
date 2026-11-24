package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

// flow-guard await-rows <worktree> <key> [<key>…] [-timeout <s>] — the parent's one bounded wait on
// its children: for <worktree>'s change, every 5 s inside this one call, runs the
// end-stale-dispatch-rows sweep and reads the dispatch rows, until every named key's row is ended
// — one `ENDED <key> <outcome>` line per key, in argument order — or the timeout (default 540, max
// 570), which prints the ended keys' lines and `WAITING <keys still open>`, exit 0. A key the
// change does not record exits 2 (`unknown-key`); an unreadable store exits 1.
// skills/flow/implement.md's **Turn discipline** is the contract.
//
// Go, in flow-guard (stats/internal/guard/awaitrows.go).
func init() { Registry["await-rows"] = awaitRows }

const awaitRowsPoll = 5 * time.Second

func awaitRows(args []string, env Env, stdout, stderr io.Writer) int {
	timeout, rest, ok := awaitTimeout(args)
	if !ok || len(rest) < 2 {
		Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard await-rows <worktree> <key> [<key>…] [-timeout <seconds, max 570>]")
		return 2
	}
	worktree, keys := rest[0], rest[1:]
	change := filepath.Base(worktree)
	deadline := time.Now().Add(timeout)
	for {
		endStaleDispatchRows([]string{worktree}, env, io.Discard, stderr)
		out, _, rc := pcFlow(env, false, "record", "dispatches", "-change", change, "-C", worktree)
		var rows []struct {
			Key     string  `json:"key"`
			Outcome string  `json:"outcome"`
			EndedAt *string `json:"endedAt"`
		}
		if rc != 0 || json.Unmarshal(out, &rows) != nil {
			Reason(env, "store-unreadable")
			fmt.Fprintf(stderr, "await-rows: flow record dispatches exit %d\n", rc)
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
				Reason(env, "unknown-key")
				fmt.Fprintf(stderr, "await-rows: %s records no dispatch row %s\n", change, k)
				return 2
			}
			if !ended[k] {
				open = append(open, k)
			}
		}
		if len(open) == 0 || !time.Now().Before(deadline) {
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
		time.Sleep(min(awaitRowsPoll, time.Until(deadline)))
	}
}
