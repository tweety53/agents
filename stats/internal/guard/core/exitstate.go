package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/tweety53/agents/stats/internal/caller"
)

// Exit is a guard's exit code, returned up to its entry point (the
// integrate-change family's, visual-verify-pass's) once its lines are written
// and its reason recorded: the bash's `exit` from inside any helper.
type Exit int

func (e Exit) Error() string { return fmt.Sprintf("exit %d", int(e)) }

// ExitCodeOf is the exit code err carries: 0 for nil.
func ExitCodeOf(err error) int {
	if err == nil {
		return 0
	}
	return int(err.(Exit))
}

// State is the record `flow state get` prints, its worktrees in the
// record's own key order (jq's to_entries).
type State struct {
	Keys, Values []string
	PrURL, Jira  string
}

func ParseState(s string) State {
	var st State
	var top map[string]json.RawMessage
	if json.Unmarshal([]byte(s), &top) != nil {
		return st
	}
	var pr, jira any
	_ = json.Unmarshal(top["prUrl"], &pr)
	_ = json.Unmarshal(top["jiraIssue"], &jira)
	st.PrURL, st.Jira = JQText(pr), JQText(jira)
	dec := json.NewDecoder(bytes.NewReader(top["worktrees"]))
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return st
	}
	for dec.More() {
		k, err := dec.Token()
		if err != nil {
			return st
		}
		var v any
		if dec.Decode(&v) != nil {
			return st
		}
		val := JQText(v)
		if v == nil || v == false {
			val = "-" // `.value // "-"`: only null and false fall through
		}
		st.Keys, st.Values = append(st.Keys, k.(string)), append(st.Values, val)
	}
	return st
}

// JQText is `jq -r '.x // empty'` of one value.
func JQText(v any) string {
	switch x := v.(type) {
	case nil, bool:
		if x == true {
			return "true"
		}
		return ""
	case string:
		return x
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// ReadLines splits text into lines as `read` sees them.
func ReadLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

var Exportable = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

// EnvHarness is the harness a guard's own stage marks name when its caller
// passes no --harness: $FLOW_HARNESS, else claude-code under Claude Code's
// CLAUDECODE=1, else "" -- the `flow` CLI's resolveHarness, less its
// "unknown" floor, so a guard still refuses a mark it cannot attribute.
// EnvToken is $FLOW_SESSION_TOKEN, else the calling agent's token (internal/caller), else "".
func EnvToken(env Env) string {
	if t := env.Getenv("FLOW_SESSION_TOKEN"); t != "" {
		return t
	}
	return caller.Token(env.Getenv)
}

func EnvHarness(env Env) string {
	if h := env.Getenv("FLOW_HARNESS"); h != "" {
		return h
	}
	if env.Getenv("CLAUDECODE") == "1" {
		return "claude-code"
	}
	return ""
}
