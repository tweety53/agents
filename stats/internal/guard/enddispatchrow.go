package guard

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/caller"
)

// flow-guard end-dispatch-row <agent-transcript> <session-id> <cwd> [<final-message>] — close a
// flow dispatch's row when its subagent stops. hooks/await-own-agents.py runs
// it on every SubagentStop it lets through, so the parent ends no row of its
// own (the begin is hooks/enforce-agent-baseline.py's).
//
// The row is the one the transcript's last prompt carrying a `Key:` line names
// — a child resumed by a SendMessage whose message carries its own `Key:` line
// ends that resume's row, never the first prompt's again — else the first
// prompt's edrMetaKey,
// under the session's change -- the session's latest open stage run,
// else its latest, of a change that is no loop's own (as the begin hook
// reads them) -- and its dispatcher's token, the one the begin hook recorded
// it under: caller.For the session and the sidecar's `parentAgentId`, the
// main thread's when it names none. A loop key (`pipeline-fix-…`, `self-review-…`) and a verifier
// key (`visual-verify…`, which visual-verify-pass close ends) are left alone.
//
// The outcome reads the agent's final message — the hook payload's
// `last_assistant_message` when passed, else the transcript's last assistant
// text, which a stopping agent's transcript may not hold yet:
//   - a final message: `completed`;
//   - no final message at all: the row is left open, never written `aborted` — an empty read is not
//     a verdict, and the parent's own end or sweep-dispatch-rows closes it from the real report.
//
// The commit is the last `commit: <task-id> <40-hex sha>` line of the final
// message, passed as -commit; on a change spanning repositories, whose line
// carries a `<worktree>=<sha>[,…]` map, the map's first sha.
//
// Exit 0 always: recording is best-effort and never blocks a stop. It prints
// END-DISPATCH-ROW-OK <key> <outcome>, or END-DISPATCH-ROW-NA <why>.
//
// Go, in flow-guard (stats/internal/guard/enddispatchrow.go).
func init() { Registry["end-dispatch-row"] = endDispatchRow }

var (
	edrKey    = regexp.MustCompile(`(?m)^Key:[ \t]*(\S+)[ \t]*$`)
	edrCommit = regexp.MustCompile(`(?m)^commit:[ \t]*\S+[ \t]+(?:[^=\s]+=)?([0-9a-f]{40})\b`)
)

func endDispatchRow(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) < 3 || len(args) > 4 {
		Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard end-dispatch-row <agent-transcript> <session-id> <cwd> [<final-message>]")
		return 2
	}
	na := func(why string) int { fmt.Fprintln(stdout, "END-DISPATCH-ROW-NA "+why); return 0 }
	prompt, final, err := edrTranscript(args[0])
	if err != nil {
		return na("transcript unreadable: " + err.Error())
	}
	if len(args) == 4 && strings.TrimSpace(args[3]) != "" {
		final = args[3]
	}
	var key string
	if m := edrKey.FindStringSubmatch(prompt); m != nil {
		key = m[1]
	} else if key = edrMetaKey(args[0]); key == "" {
		return na("no Key: line")
	}
	for _, p := range []string{"pipeline-fix-", "self-review-", "visual-verify"} {
		if strings.HasPrefix(key, p) {
			return na(key + " is ended by its own loop or guard")
		}
	}
	change := edrSessionChange(env, args[1])
	token := caller.For(args[1], edrParent(args[0]))
	if change == "" || token == "" {
		return na("no flow change for session " + args[1])
	}
	if strings.TrimSpace(final) == "" {
		return na(key + " left a no final message; its row stays open")
	}
	outcome := "completed"
	end := []string{"record", "dispatch", "end", "-C", args[2], "-change", change, "-key", key, "-session-token", token, "-outcome", outcome}
	if c := edrCommit.FindAllStringSubmatch(final, -1); len(c) > 0 {
		end = append(end, "-commit", c[len(c)-1][1])
	}
	if out, _, rc := pcFlow(env, true, end...); rc != 0 {
		return na(fmt.Sprintf("flow record dispatch end exit %d: %s", rc, strings.TrimSpace(string(out))))
	}
	fmt.Fprintf(stdout, "END-DISPATCH-ROW-OK %s %s\n", key, outcome)
	return 0
}

var (
	edrPlanningKey = regexp.MustCompile(`(?:^|\s)((?:explore|design)-\d+)(?:\s|$)`)
	// edrBeginKey is a key the enforce-agent-baseline hook begins a row under, matched whole: a
	// Key shape the closed list names (**Dispatch sites -- the parent's closed list**,
	// skills/flow/implement.md) -- `panel-4-primary`, `visual-verify-fix-11-gymie`,
	// `full-suite-fix-2`, a fix's implementer `fix-2-implementer` or reviewer `fix-2-review`, a
	// planner's `plan-2`, its designer's `designer-1`, a /flow-interactive request's operator-asked
	// `operator-58` -- or a planner's Explore/Plan agent's `explore-<n>`/`design-<n>`.
	edrBeginKey = regexp.MustCompile(`^(?:task-\d+(?:\+\d+)*-(?:implementer|reviewer)(?:-[a-z0-9-]+)?` +
		`|panel-\d+-[a-z][a-z0-9+-]*|panel-fix-\d+(?:-\d+)?(?:-retry)?|fix-\d+-(?:implementer|review)` +
		`|full-suite-fix-\d+|visual-verify(?:-[a-z0-9][a-z0-9-]*)?|pipeline-fix-[a-z0-9][a-z0-9-]*` +
		`|plan-\d+|designer-\d+|operator-\d+(?:-[a-z0-9][a-z0-9-]*)?|explore-\d+|design-\d+)$`)
)

// edrMetaKey is the key the enforce-agent-baseline hook's rowKey gave an agent
// whose prompt has no Key: line, off its meta sidecar: the description's
// explore-/design- word, else — for an agent not of a planning type (Explore,
// Plan) — its first word when begin_key begins rows under it (`panel-0-mutation`),
// else agent-<toolUseId>; a planner's Explore agent described `plan-1 …` is never
// the planner's own plan-1 row. "" without a sidecar.
func edrMetaKey(transcript string) string {
	b, err := os.ReadFile(strings.TrimSuffix(transcript, ".jsonl") + ".meta.json")
	var meta struct {
		Description string `json:"description"`
		AgentType   string `json:"agentType"`
		ToolUseID   string `json:"toolUseId"`
	}
	if err != nil || json.Unmarshal(b, &meta) != nil {
		return ""
	}
	if m := edrPlanningKey.FindStringSubmatch(meta.Description); m != nil {
		return m[1]
	}
	planning := meta.AgentType == "Explore" || meta.AgentType == "Plan"
	if w := strings.Fields(meta.Description); !planning && len(w) > 0 && edrBeginKey.MatchString(w[0]) {
		return w[0]
	}
	if meta.ToolUseID == "" {
		return ""
	}
	return "agent-" + meta.ToolUseID
}

// edrParent is the agent that launched transcript's agent, off its sidecar's `parentAgentId`; ""
// for the main thread's child or without a sidecar.
func edrParent(transcript string) string {
	b, _ := os.ReadFile(strings.TrimSuffix(transcript, ".jsonl") + ".meta.json")
	var meta struct {
		Parent string `json:"parentAgentId"`
	}
	_ = json.Unmarshal(b, &meta)
	return meta.Parent
}

// edrTranscript is the transcript's last user prompt carrying a `Key:` line,
// else its first user prompt, and its last assistant text.
func edrTranscript(path string) (prompt, final string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 64<<20)
	for sc.Scan() {
		var rec struct {
			Type    string `json:"type"`
			Message struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(sc.Bytes(), &rec) != nil {
			continue
		}
		text := edrText(rec.Message.Content)
		switch {
		case rec.Type == "user" && (prompt == "" || edrKey.MatchString(text)):
			prompt = text
		case rec.Type == "assistant" && strings.TrimSpace(text) != "":
			final = text
		}
	}
	return prompt, final, sc.Err()
}

// edrText is a message content's text: the string itself, or its text parts
// joined, plus a SubagentHandback call's message.
func edrText(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var parts []struct {
		Type  string `json:"type"`
		Text  string `json:"text"`
		Name  string `json:"name"`
		Input struct {
			Message string `json:"message"`
		} `json:"input"`
	}
	if json.Unmarshal(raw, &parts) != nil {
		return ""
	}
	var b strings.Builder
	for _, p := range parts {
		switch {
		case p.Type == "text":
			b.WriteString(p.Text)
		case p.Type == "tool_use" && p.Name == "SubagentHandback":
			// A child that reports through the handback tool ends on a tool call, so its report is
			// the call's message, not a text part.
			b.WriteString(p.Input.Message)
		}
	}
	return b.String()
}

// edrSessionChange is the session's change from flowd's stage
// runs, as hooks/enforce-agent-baseline.py's session_change reads them.
func edrSessionChange(env Env, sessionID string) string {
	addr := env.Getenv("FLOW_ADDR")
	if addr == "" {
		addr = "http://127.0.0.1:4173"
	}
	q := url.Values{"session_id": {sessionID}, "sort": {"-started_at"}, "limit": {"50"}}
	resp, err := (&http.Client{Timeout: 3 * time.Second, Transport: &http.Transport{}}).Get(addr + "/api/v1/stage-runs?" + q.Encode())
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var body struct {
		StageRuns []struct {
			ChangeName   string  `json:"changeName"`
			SessionToken string  `json:"sessionToken"`
			EndedAt      *string `json:"endedAt"`
		} `json:"stageRuns"`
	}
	if json.NewDecoder(resp.Body).Decode(&body) != nil {
		return ""
	}
	var pick *struct {
		ChangeName   string  `json:"changeName"`
		SessionToken string  `json:"sessionToken"`
		EndedAt      *string `json:"endedAt"`
	}
	for i := range body.StageRuns {
		r := &body.StageRuns[i]
		if strings.HasPrefix(r.ChangeName, "pipeline-fix-") || strings.HasPrefix(r.ChangeName, "self-review-") {
			continue
		}
		if pick == nil {
			pick = r
		}
		if r.EndedAt == nil {
			pick = r
			break
		}
	}
	if pick == nil {
		return ""
	}
	return pick.ChangeName
}

// flow-guard sweep-dispatch-rows <subagents-dir> <session-id> <cwd> — close the rows of agents whose
// stop never closed them. For each agent transcript in the dir that is 30 minutes to 24 hours
// old, holds a SubagentHandback report and carries no `.swept` sidecar, it runs end-dispatch-row
// on it (the guard reads the report) and writes the sidecar so it is tried once, unless the end
// failed for a transient reason (no flow change for the session, a refused end). It then aborts the
// change's visual-verify rows open past 30 minutes (sweepVisualVerifyRows). hooks/await-own-agents.py
// runs it after each stop. An agent with no report is never swept: a quiet long build looks the same.
// Exit 0 always. Go, in flow-guard (stats/internal/guard/enddispatchrow.go).
// Run in process by the await-own-agents hook only; no guard.Registry key.
func sweepDispatchRows(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 3 {
		Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard sweep-dispatch-rows <subagents-dir> <session-id> <cwd>")
		return 2
	}
	files, _ := filepath.Glob(filepath.Join(args[0], "agent-*.jsonl"))
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		age := time.Since(info.ModTime())
		if age < 30*time.Minute || age > 24*time.Hour {
			continue
		}
		if _, err := os.Stat(f + ".swept"); err == nil {
			continue
		}
		if b, err := os.ReadFile(f); err != nil || !bytes.Contains(b, []byte(`"SubagentHandback"`)) {
			continue
		}
		var out bytes.Buffer
		endDispatchRow([]string{f, args[1], args[2]}, env, &out, stderr)
		_, _ = stdout.Write(out.Bytes())
		// A transient miss (flowd down, no change for the session yet, a refused end) leaves no
		// sidecar, so the next stop retries; otherwise the parent's handoff would end the row aborted.
		if strings.Contains(out.String(), "no flow change") || strings.Contains(out.String(), "flow record dispatch end exit") {
			continue
		}
		_ = os.WriteFile(f+".swept", nil, 0o644)
	}
	sweepVisualVerifyRows(args, env)
	return 0
}

// sweepVisualVerifyRows ends the session's change's visual-verify rows still open 30 minutes after
// their begin, aborted, each under its own token. visual-verify-pass close ends these rows and
// end-dispatch-row leaves them to it, so a round that was interrupted, superseded or never
// launched an agent would otherwise leave its rows open for good.
func sweepVisualVerifyRows(args []string, env Env) {
	change := edrSessionChange(env, args[1])
	if change == "" {
		return
	}
	rows, err := flowdDispatches(env, args[2], change)
	if err != nil {
		return
	}
	for _, r := range rows {
		if strings.HasPrefix(r.Key, "visual-verify") && r.EndedAt == nil && r.SessionToken != "" && time.Since(r.StartedAt) > 30*time.Minute {
			pcFlow(env, true, "record", "dispatch", "end", "-C", args[2], "-change", change, "-key", r.Key, "-session-token", r.SessionToken, "-outcome", "aborted")
		}
	}
}
