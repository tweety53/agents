package guard

import (
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

// flow-guard hook flow-session-activity — tell flowd what this session is doing now.
//
// The Live state board's Running now cards say whether a change's session is working or
// waiting on the operator. This hook reports each transition as one
// PUT /api/v1/sessions/<session_id>/activity {"state", "detail"}:
//
//	PreToolUse on AskUserQuestion  -> question, the first question's text
//	Notification naming permission -> permission, the notification's message
//	PreToolUse, PostToolUse        -> working, the tool and its short label
//	UserPromptSubmit               -> working
//	Stop                           -> idle
//	SessionEnd                     -> ended
//
// Any other event sends nothing. detail is cut to 500 bytes, the most flowd accepts.
//
// Fails open, like flow-active-change: malformed input, a payload without a session id
// or an unreachable store exits 0 with no output. A hook that reports state must never be
// the reason a tool call or a prompt is held up.
func init() { hooks["flow-session-activity"] = flowSessionActivity }

func flowSessionActivity(env Env, payload map[string]any, _, _ io.Writer) {
	sessionID := strOf(payload["session_id"])
	if sessionID == "" {
		return
	}
	state, detail, ok := sessionActivity(payload)
	if !ok {
		return
	}
	// Postgres text refuses U+0000, so a NUL would make the store reject the row.
	detail = strings.ReplaceAll(detail, "\x00", "")
	if len(detail) > 500 {
		detail = detail[:500]
		for !utf8.ValidString(detail) {
			detail = detail[:len(detail)-1]
		}
	}
	body := pyDumpASCII(pyObj{{"state", state}, {"detail", detail}})
	req, err := http.NewRequest(http.MethodPut, hookFlowAddr(env)+"/api/v1/sessions/"+pyQuoteURL(sessionID, "")+"/activity", strings.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if resp, err := (&http.Client{Timeout: 2 * time.Second, Transport: &http.Transport{}}).Do(req); err == nil {
		_ = resp.Body.Close()
	}
}

// sessionActivity is the (state, detail) an event reports; ok false sends nothing.
func sessionActivity(p map[string]any) (state, detail string, ok bool) {
	event := p["hook_event_name"]
	tool := event == "PreToolUse" || event == "PostToolUse"
	// A subagent's tool events fire under the parent's session_id and carry
	// agent_id (https://code.claude.com/docs/en/hooks); they must not overwrite
	// the main thread's state, e.g. turn an open question back to working
	// (kan-1048 self-review, 1-J). Its Notification still reaches the board.
	if tool && pyTruthy(p["agent_id"]) {
		return "", "", false
	}
	switch {
	case event == "PreToolUse" && p["tool_name"] == "AskUserQuestion":
		in := map[string]any{}
		if pyTruthy(p["tool_input"]) {
			if in, ok = p["tool_input"].(map[string]any); !ok {
				return "", "", false
			}
		}
		first := map[string]any{}
		if qs := in["questions"]; pyTruthy(qs) {
			l, isList := qs.([]any)
			if !isList {
				return "", "", false
			}
			if first, ok = l[0].(map[string]any); !ok {
				return "", "", false
			}
		}
		return "question", pyStrOr(first["question"], ""), true
	case event == "Notification":
		message := pyStrOr(p["message"], "")
		if strings.Contains(strings.ToLower(message), "permission") {
			return "permission", message, true
		}
		return "", "", false
	// Every tool call reports working, before it runs and after: a long call
	// (a poll, a build) must not read as a turn that ended while it runs.
	case tool:
		return "working", sessionAction(p), true
	case event == "UserPromptSubmit":
		return "working", "", true
	case event == "Stop":
		return "idle", "", true
	case event == "SessionEnd":
		return "ended", "", true
	}
	return "", "", false
}

// sessionAction is what the main thread just did, in a few words, for the board: the tool and
// its own short label -- a Bash call's description (else its first command line), an edited or
// read file's name, a dispatched agent's or skill's name.
func sessionAction(p map[string]any) string {
	tool := pyStrOr(p["tool_name"], "")
	in := map[string]any{}
	if pyTruthy(p["tool_input"]) {
		var ok bool
		if in, ok = p["tool_input"].(map[string]any); !ok {
			return tool
		}
	}
	label := ""
	switch tool {
	case "Bash":
		if label = pyStrOr(in["description"], ""); label == "" {
			label = pySplitFirst(pyStrOr(in["command"], ""))
		}
		if label != "" {
			return "Bash: " + label
		}
	case "Edit", "Write", "Read", "NotebookEdit":
		path := pyStrOr(in["file_path"], "")
		if path == "" {
			path = pyStrOr(in["notebook_path"], "")
		}
		if path != "" {
			return tool + " " + path[strings.LastIndex(path, "/")+1:]
		}
	case "Agent", "Task":
		if label = pyStrOr(in["description"], ""); label == "" {
			label = pyStrOr(in["subagent_type"], "")
		}
		if label != "" {
			return "Agent: " + label
		}
	case "Skill":
		if pyTruthy(in["skill"]) {
			return "Skill " + pyStr(in["skill"])
		}
	}
	return tool
}
