package guard

import (
	"io"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// flow-guard hook flow-active-change — UserPromptSubmit: name the session's active /flow or
// /flow-interactive change on every plain prompt.
//
// A problem report typed as a plain message, with no /flow prefix, needs to know which
// change it belongs to before the SKILL.md rule can turn it into a fix run. The change name
// is not carried in session state a hook can read directly, but every /flow run marks its
// own stages with `flow stage begin -command '/flow' ... <name>`,
// and the stats daemon harvests those marks out of the session's transcript into its store,
// binding each to the session that made it. This hook asks the store for this session's latest
// stage-begin mark of any command, skipping a loop's own runs -- a change named `pipeline-fix-...`
// or `self-review-...`, whose stage runs a loop dispatch's begins open in the parent's session --
// one GET against /api/v1/stage-runs, whose rows carry the owning change's name, the command and
// the stage -- and branches on its command: a /flow mark injects the fix-run line naming the
// change; a /flow-interactive mark, unless it is flow.cleanup, injects the line pointing a plain
// message at the session's request protocol (so it survives a compaction); any other command
// injects nothing.
//
// Fails open, like enforce-agent-baseline: any unexpected input, an unreachable store,
// or an answer that names no change exits 0 with no output. A hook that supplies context
// must never be the reason a prompt is refused. A mark younger than the daemon's harvest
// cycle is not in the store yet, and the hook answers nothing in that window rather than
// re-parsing the transcript itself -- a fallback would keep the per-prompt transcript read
// this hook exists to avoid.
func init() { hooks["flow-active-change"] = flowActiveChange }

func flowActiveChange(env core.Env, payload map[string]any, stdout, _ io.Writer) {
	if payload == nil {
		return
	}
	if p, ok := payload["prompt"].(string); ok && strings.HasPrefix(p, "/") {
		return
	}
	sessionID := facSessionID(env, payload)
	if sessionID == "" {
		return
	}
	q := func(s string) string { return strings.ReplaceAll(pyQuoteURL(s, " "), " ", "+") }
	v, err := hookGetJSON(env, core.FlowAddr(env)+"/api/v1/stage-runs?session_id="+q(sessionID)+"&sort=-started_at&limit=50", 2*time.Second)
	if err != nil {
		return
	}
	body, _ := v.(map[string]any)
	runs, _ := body["stageRuns"].([]any)
	var run map[string]any
	for _, r := range runs {
		// A pipeline-fix or self-review loop's own change: its stage runs carry the parent's session id.
		if m, ok := r.(map[string]any); ok && !hasAnyPrefix(pyStrOr(m["changeName"], ""), []string{"pipeline-fix-", "self-review-"}) {
			run = m
			break
		}
	}
	name := strOf(run["changeName"])
	if name == "" {
		return
	}
	var line string
	switch command := run["command"]; {
	case command == "/flow":
		line = "flow: this session's last /flow run marked change " + name + ". A plain message " +
			"reporting a problem or asking for a change to it is a fix run of " +
			name + " — **A plain message at IN_PROGRESS** (skills/flow/SKILL.md)."
	case command == "/flow-interactive" && run["stage"] != "flow.cleanup":
		line = "flow: this session runs /flow-interactive change `" + name + "`. A plain message " +
			"is its next request — or its stop (**One request**, **Explicit stop**, " +
			"`skills/flow-interactive/SKILL.md`)."
	default:
		return
	}
	_, _ = io.WriteString(stdout, hookJSON(map[string]any{"hookSpecificOutput": map[string]any{
		"hookEventName":     "UserPromptSubmit",
		"additionalContext": line,
	}}))
}

// facSessionID is the payload's session_id, else the stem of a `.jsonl` transcript_path (a
// claude-code transcript is named after the bare session id), else CLAUDE_SESSION_ID.
func facSessionID(env core.Env, payload map[string]any) string {
	if s := strOf(payload["session_id"]); s != "" {
		return s
	}
	if tp := strOf(payload["transcript_path"]); tp != "" {
		base := tp[strings.LastIndex(tp, "/")+1:]
		// Python's os.path.splitext: a basename's leading dots never start its extension.
		if i := strings.LastIndex(base, "."); i > 0 && strings.Trim(base[:i], ".") != "" && base[i:] == ".jsonl" {
			return base[:i]
		}
	}
	return env.Getenv("CLAUDE_SESSION_ID")
}
