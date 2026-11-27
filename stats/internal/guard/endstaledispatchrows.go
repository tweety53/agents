package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// flow-guard end-stale-dispatch-rows <worktree> — end the open dispatch rows of <worktree>'s change
// whose agent already finished: a row with an agentId whose `subagents/agent-<id>.meta.json` says
// "finished": true and whose transcript was last written at least staleFinishedAge ago. A stop that
// was blocked, or an agent its parent killed, never reached end-dispatch-row. The row is ended
// `aborted`, dated at the transcript's last write. A row with no agentId, an unfinished or missing
// meta, or a recent transcript is left alone. An agent whose transcript holds a SubagentHandback
// report delivered it: its row ends with the outcome and commit end-dispatch-row derives from that
// report, never `aborted`. Prints `ENDED <key> <agentId>` per row. Exit 0
// always (fail open: a store failure on one row goes to stderr and the sweep continues).
// hooks/await-own-agents.py runs it after every SubagentStop.
//
// Go, in flow-guard (stats/internal/guard/endstaledispatchrows.go).
func init() { Registry["end-stale-dispatch-rows"] = endStaleDispatchRows }

const staleFinishedAge = 10 * time.Minute

func endStaleDispatchRows(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard end-stale-dispatch-rows <worktree>")
		return 2
	}
	worktree := args[0]
	change := filepath.Base(worktree)
	out, _, rc := pcFlow(env, false, "record", "dispatches", "-change", change, "-C", worktree)
	if rc != 0 {
		fmt.Fprintf(stderr, "end-stale-dispatch-rows: flow record dispatches exit %d\n", rc)
		return 0
	}
	var rows []struct {
		AgentID      string  `json:"agentId"`
		Key          string  `json:"key"`
		SessionToken string  `json:"sessionToken"`
		EndedAt      *string `json:"endedAt"`
	}
	if json.Unmarshal(out, &rows) != nil {
		fmt.Fprintln(stderr, "end-stale-dispatch-rows: unreadable dispatches")
		return 0
	}
	for _, r := range rows {
		if r.EndedAt != nil || r.AgentID == "" || r.AgentID == "none" || r.AgentID == "inline" {
			continue
		}
		transcript, last, ok := finishedAgentLastWrite(env.Getenv("HOME"), r.AgentID)
		if !ok || time.Since(last) < staleFinishedAge {
			continue
		}
		end := append([]string{"record", "dispatch", "end", "-C", worktree, "-change", change,
			"-key", r.Key, "-session-token", r.SessionToken, "-outcome", handbackOutcome(transcript),
			"-agent-id", r.AgentID, "-ended-at", last.UTC().Format(time.RFC3339Nano)}, handbackCommit(transcript)...)
		_, e, rc := pcFlow(env, false, end...)
		if rc != 0 {
			fmt.Fprintf(stderr, "end-stale-dispatch-rows: %s: flow record dispatch end exit %d: %s\n", r.Key, rc, strings.TrimSpace(string(e)))
			continue
		}
		fmt.Fprintf(stdout, "ENDED %s %s\n", r.Key, r.AgentID)
	}
	return 0
}

// finishedAgentLastWrite is the agent transcript's path and mtime when its meta says finished.
func finishedAgentLastWrite(home, id string) (string, time.Time, bool) {
	metas, _ := filepath.Glob(filepath.Join(home, ".claude", "projects", "*", "*", "subagents", "agent-"+id+".meta.json"))
	for _, meta := range metas {
		b, err := os.ReadFile(meta)
		if err != nil {
			continue
		}
		var m struct {
			Finished bool `json:"finished"`
		}
		if json.Unmarshal(b, &m) != nil || !m.Finished {
			continue
		}
		tr := strings.TrimSuffix(meta, ".meta.json") + ".jsonl"
		if fi, err := os.Stat(tr); err == nil {
			return tr, fi.ModTime(), true
		}
	}
	return "", time.Time{}, false
}

// handbackReport is the transcript's SubagentHandback report, "" when it delivered none.
func handbackReport(transcript string) string {
	b, err := os.ReadFile(transcript)
	if err != nil || !bytes.Contains(b, []byte(`"SubagentHandback"`)) {
		return ""
	}
	_, final, _ := edrTranscript(transcript)
	return final
}

// handbackOutcome is end-dispatch-row's outcome for a delivered report, `completed`, else `aborted`.
func handbackOutcome(transcript string) string {
	if final := handbackReport(transcript); strings.TrimSpace(final) != "" {
		return "completed"
	}
	return "aborted"
}

// handbackCommit is `-commit <sha>` for the report's last commit: line, else nothing.
func handbackCommit(transcript string) []string {
	if c := edrCommit.FindAllStringSubmatch(handbackReport(transcript), -1); len(c) > 0 {
		return []string{"-commit", c[len(c)-1][1]}
	}
	return nil
}
