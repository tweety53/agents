package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"
)

// flow-guard resume-or-fresh — decide whether an immediate send-back resumes
// the agent that just handed back (KAN-1117).
//
// Usage: flow-guard resume-or-fresh <worktree> <change-name> <key>
//
// <key> is the key of the dispatch row the parent just closed. The guard reads
// the change's dispatch rows and takes the highest-seq row under that key.
//
// Exit 0, stdout `resume <agentId>`: the agent ended under 240 s ago and its
// final context is under 100000 tokens.
// Exit 1, stdout `fresh: <reason>`, the first failing check in this order:
// `rows unreadable`, `no row <key>`, `no agent id` (empty, none or inline),
// `row open`, `age <n>s >= 240s`, `context unknown`, `context <n> >= 100000`.
// Exit 2: a wrong argument count (usage on stderr).

const (
	// rofMaxAge leaves a 60 s margin under the subagent cache's 300 s TTL.
	rofMaxAge = 240 * time.Second
	// rofMaxContext is the ticket's bound on the context a resume carries.
	rofMaxContext = 100000
)

// rofRow is the part of a `flow record dispatches` row the guard reads.
type rofRow struct {
	Seq     int64      `json:"seq"`
	Key     string     `json:"key"`
	AgentID string     `json:"agentId"`
	EndedAt *time.Time `json:"endedAt"`
	Metrics struct {
		Signals struct {
			Sidechain struct {
				ContextEnd string `json:"context_end"`
			} `json:"sidechain"`
		} `json:"signals"`
	} `json:"metrics"`
}

func init() { Registry["resume-or-fresh"] = resumeOrFresh }

func resumeOrFresh(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 3 {
		Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard resume-or-fresh <worktree> <change-name> <key>")
		return 2
	}
	worktree, name, key := args[0], args[1], args[2]
	fresh := func(format string, a ...any) int {
		Reason(env, "agent-not-resumable")
		fmt.Fprintf(stdout, "fresh: "+format+"\n", a...)
		return 1
	}
	raw, err := pfdRead(env, env.Dispatches, "dispatches", name, worktree)
	var rows []rofRow
	if err == nil {
		err = json.Unmarshal(raw, &rows)
	}
	if err != nil {
		return fresh("rows unreadable")
	}
	var row *rofRow
	for i := range rows {
		if rows[i].Key == key && (row == nil || rows[i].Seq > row.Seq) {
			row = &rows[i]
		}
	}
	switch {
	case row == nil:
		return fresh("no row %s", key)
	case row.AgentID == "" || row.AgentID == "none" || row.AgentID == "inline":
		return fresh("no agent id")
	case row.EndedAt == nil:
		return fresh("row open")
	}
	if age := time.Since(*row.EndedAt); age >= rofMaxAge {
		return fresh("age %ds >= %ds", int(age.Seconds()), int(rofMaxAge.Seconds()))
	}
	ctx, err := strconv.ParseInt(row.Metrics.Signals.Sidechain.ContextEnd, 10, 64)
	if err != nil {
		return fresh("context unknown")
	}
	if ctx >= rofMaxContext {
		return fresh("context %d >= %d", ctx, rofMaxContext)
	}
	fmt.Fprintf(stdout, "resume %s\n", row.AgentID)
	return 0
}
