package guard

import (
	"bytes"
	"encoding/json"
	"io"
	"path"
	"regexp"
	"strings"

	"github.com/tweety53/agents/stats/internal/caller"
)

// The block-unsafe-commands hook's second job: name the subagent making a `flow` call.
//
// Subagents share their parent's CLAUDE_CODE_SESSION_ID, and only a hook payload names the agent
// (`agent_id`, absent on the main thread). When the hook lets a subagent's Bash call through, it
// answers with `updatedInput` and no permission decision — the harness then runs its own
// permission check on the rewritten command — inserting `-caller-agent <agent_id>` after every
// `flow` word and after the guard name of every `flow-guard` word at a command position: the
// command's start, after `;`, `&`, `|`, `(`, `$(` or a newline, past `NAME=value` words and the
// keywords `do`, `then`, `else`, `elif`, `!`, `{`, `time`. The command still starts with the
// program it named, so a `Bash(flow:*)` rule still matches it. Quoted text, heredoc bodies and
// any other word are left as they are. internal/caller turns the flag into the session token.
//
// A call this misses — `bash -c 'flow …'`, `"$X/flow"`, a flow inside `"$(…)"` — runs as the main
// thread's. An agent id that is not a plain word is never written into a command.

var callerAgentID = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

var callerKeywords = map[string]bool{"do": true, "then": true, "else": true, "elif": true, "!": true, "{": true, "time": true}

var callerAssignment = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

// callerAgentHook writes the updatedInput answer for a subagent's Bash call naming flow, else
// nothing.
func callerAgentHook(payload map[string]any, stdout io.Writer) {
	agent, _ := payload["agent_id"].(string)
	input, _ := payload["tool_input"].(map[string]any)
	command, _ := input["command"].(string)
	if payload["tool_name"] != "Bash" || !callerAgentID.MatchString(agent) || command == "" {
		return
	}
	tagged := callerAgentCommand(command, agent)
	if tagged == command {
		return
	}
	updated := map[string]any{}
	for k, v := range input {
		updated[k] = v
	}
	updated["command"] = tagged
	out, err := json.Marshal(map[string]any{"hookSpecificOutput": map[string]any{
		"hookEventName": "PreToolUse",
		"updatedInput":  updated,
	}})
	if err == nil {
		_, _ = stdout.Write(out)
	}
}

// callerAgentCommand is command with `-caller-agent agent` inserted into each flow call.
func callerAgentCommand(command, agent string) string {
	var b strings.Builder
	var heredocs []string
	start := true
	n := len(command)
	for i := 0; i < n; {
		c := command[i]
		switch {
		case c == '\'' || c == '"':
			j := i + 1
			for j < n && command[j] != c {
				if c == '"' && command[j] == '\\' {
					j++
				}
				j++
			}
			j = min(j+1, n)
			b.WriteString(command[i:j])
			i, start = j, false
		case c == '\\':
			j := min(i+2, n)
			b.WriteString(command[i:j])
			i, start = j, false
		case c == ' ' || c == '\t':
			b.WriteByte(c)
			i++
		case c == '\n':
			b.WriteByte(c)
			i, start = i+1, true
			for _, delim := range heredocs {
				for i < n {
					end := strings.IndexByte(command[i:], '\n')
					line := command[i:]
					if end >= 0 {
						line = command[i : i+end+1]
					}
					b.WriteString(line)
					i += len(line)
					if strings.TrimLeft(strings.TrimRight(line, "\n"), "\t") == delim {
						break
					}
				}
			}
			heredocs = nil
		case c == '$' && i+1 < n && command[i+1] == '(':
			b.WriteString("$(")
			i, start = i+2, true
		case strings.IndexByte(";&|(", c) >= 0:
			b.WriteByte(c)
			i, start = i+1, true
		case c == '<' && strings.HasPrefix(command[i:], "<<") && !strings.HasPrefix(command[i:], "<<<"):
			j := i + 2
			if j < n && command[j] == '-' {
				j++
			}
			for j < n && (command[j] == ' ' || command[j] == '\t') {
				j++
			}
			k := j
			for k < n && strings.IndexByte(" \t\n;&|()<>", command[k]) < 0 {
				k++
			}
			if delim := strings.Trim(command[j:k], `'"\`); delim != "" {
				heredocs = append(heredocs, delim)
			}
			b.WriteString(command[i:k])
			i, start = k, false
		default:
			j := i
			for j < n && strings.IndexByte(" \t\n;&|()<>'\"\\`", command[j]) < 0 {
				j++
			}
			if j == i {
				b.WriteByte(c)
				i, start = i+1, false
				continue
			}
			word := command[i:j]
			b.WriteString(word)
			i = j
			plain := j == n || strings.IndexByte(" \t\n;&|()<>", command[j]) >= 0
			switch {
			case !start || !plain:
				start = false
			case callerKeywords[word] || callerAssignment.MatchString(word):
			case path.Base(word) == "flow":
				b.WriteString(" " + caller.Flag + " " + agent)
				start = false
			case path.Base(word) == "flow-guard":
				k := i
				for k < n && (command[k] == ' ' || command[k] == '\t') {
					k++
				}
				m := k
				for m < n && strings.IndexByte(" \t\n;&|()<>'\"\\`", command[m]) < 0 {
					m++
				}
				if m > k {
					b.WriteString(command[i:m] + " " + caller.Flag + " " + agent)
					i = m
				}
				start = false
			default:
				start = false
			}
		}
	}
	return b.String()
}

// withCallerAgent runs hook and, when it wrote nothing, answers a subagent's flow call.
func withCallerAgent(hook func(Env, map[string]any, io.Writer, io.Writer)) func(Env, map[string]any, io.Writer, io.Writer) {
	return func(env Env, payload map[string]any, stdout, stderr io.Writer) {
		var out bytes.Buffer
		hook(env, payload, &out, stderr)
		if out.Len() > 0 {
			_, _ = stdout.Write(out.Bytes())
			return
		}
		if payload != nil {
			callerAgentHook(payload, stdout)
		}
	}
}
