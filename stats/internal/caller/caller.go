// Package caller is the identity of the agent making a `flow` or `flow-guard` call, and the
// session token derived from it.
//
// Subagents run under their parent's CLAUDE_CODE_SESSION_ID, so the session id alone cannot
// tell a coordinator from its executor; the store closes a token's still-open stage run as
// `superseded` on its next begin, so two agents sharing a token close each other's runs. The
// harness names the agent only in a hook payload (`agent_id`, absent on the main thread):
// the block-unsafe-commands PreToolUse hook rewrites a subagent's `flow …` into
// `flow -caller-agent <id> …` and `flow-guard <name> …` into `flow-guard <name> -caller-agent
// <id> …`, and each binary strips the flag (Strip) and exports it as AgentEnv, so every `flow`
// a guard runs inherits it.
//
// The token is derived, never stored: cc-<session id>-<agent id, or main>. Every call of one
// agent carries the same token, two agents of one session carry two, and an explicit
// -session-token always wins. With no harness session id there is no token to derive.
package caller

import (
	"cmp"
	"os"
)

const (
	// Flag is the leading argument the hook inserts.
	Flag = "-caller-agent"
	// AgentEnv carries the agent id to the process and its children.
	AgentEnv = "FLOW_CALLER_AGENT"
	// SessionEnv is the harness session id every agent of a session shares.
	SessionEnv = "CLAUDE_CODE_SESSION_ID"
)

// For is the token of agent (empty: the main thread) in session; empty when session is.
func For(session, agent string) string {
	if session == "" {
		return ""
	}
	if agent == "" {
		agent = "main"
	}
	return "cc-" + session + "-" + agent
}

// Token is the calling agent's token, from the environment.
func Token(getenv func(string) string) string {
	return For(getenv(SessionEnv), getenv(AgentEnv))
}

// Strip removes every leading `-caller-agent <id>` from args, returning the rest and the first
// id: the hook inserts its own pair before any the caller wrote, so a caller's pair never reaches
// the command as an unknown argument, and the hook's id, the one the harness vouches for, wins.
func Strip(args []string) ([]string, string) {
	agent := ""
	for len(args) >= 2 && args[0] == Flag {
		agent = cmp.Or(agent, args[1])
		args = args[2:]
	}
	return args, agent
}

// Export strips args and, when they named an agent, sets AgentEnv for this process and every
// process it starts.
func Export(args []string) []string {
	rest, agent := Strip(args)
	if agent != "" {
		_ = os.Setenv(AgentEnv, agent)
	}
	return rest
}
