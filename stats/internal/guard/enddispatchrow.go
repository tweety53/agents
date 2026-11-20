package guard

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// flow-guard end-dispatch-row <agent-transcript> <session-id> <cwd> [<final-message>] — close a
// flow dispatch's row when its subagent stops. hooks/await-own-agents.py runs
// it on every SubagentStop it lets through, so the parent ends no row of its
// own (the begin is hooks/enforce-agent-baseline.py's).
//
// The row is the one the transcript's first prompt names on its `Key:` line,
// under the session's change and token: the session's latest open stage run,
// else its latest, of a change that is no loop's own (as the begin hook
// reads them). A loop key (`pipeline-fix-…`, `self-review-…`) and a verifier
// key (`visual-verify…`, which visual-verify-pass close ends) are left alone.
//
// The outcome reads the agent's final message — the hook payload's
// `last_assistant_message` when passed, else the transcript's last assistant
// text, which a stopping agent's transcript may not hold yet:
//   - a gated reviewer bundle (`task-<ids>-reviewer…`): `fix` when any
//     `task <n>: fix` line, else `clean` when any `task <n>: clean` line;
//   - any other row: `completed`;
//   - no final message, or a gated bundle naming no verdict: `aborted`.
//
// The commit is the last `commit: <task-id> <40-hex sha>` line of the final
// message, passed as -commit.
//
// Exit 0 always: recording is best-effort and never blocks a stop. It prints
// END-DISPATCH-ROW-OK <key> <outcome>, or END-DISPATCH-ROW-NA <why>.
//
// Go, in flow-guard (stats/internal/guard/enddispatchrow.go).
func init() { Registry["end-dispatch-row"] = endDispatchRow }

var (
	edrKey        = regexp.MustCompile(`(?m)^Key:[ \t]*(\S+)[ \t]*$`)
	edrCommit     = regexp.MustCompile(`(?m)^commit:[ \t]*\S+[ \t]+([0-9a-f]{40})\b`)
	edrGatedKey   = regexp.MustCompile(`^task-[0-9]+(?:\+[0-9]+)*-reviewer`)
	edrGatedFix   = regexp.MustCompile(`(?m)^task [0-9]+: fix\b`)
	edrGatedClean = regexp.MustCompile(`(?m)^task [0-9]+: clean\b`)
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
	m := edrKey.FindStringSubmatch(prompt)
	if m == nil {
		return na("no Key: line")
	}
	key := m[1]
	for _, p := range []string{"pipeline-fix-", "self-review-", "visual-verify"} {
		if strings.HasPrefix(key, p) {
			return na(key + " is ended by its own loop or guard")
		}
	}
	change, token := edrSessionChange(env, args[1])
	if change == "" || token == "" {
		return na("no flow change for session " + args[1])
	}
	outcome := edrOutcome(key, final)
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

func edrOutcome(key, final string) string {
	switch {
	case strings.TrimSpace(final) == "":
		return "aborted"
	case !edrGatedKey.MatchString(key):
		return "completed"
	case edrGatedFix.MatchString(final):
		return "fix"
	case edrGatedClean.MatchString(final):
		return "clean"
	}
	return "aborted"
}

// edrTranscript is the transcript's first user prompt and its last assistant
// text.
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
		case rec.Type == "user" && prompt == "":
			prompt = text
		case rec.Type == "assistant" && strings.TrimSpace(text) != "":
			final = text
		}
	}
	return prompt, final, sc.Err()
}

// edrText is a message content's text: the string itself, or its text parts
// joined.
func edrText(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &parts) != nil {
		return ""
	}
	var b strings.Builder
	for _, p := range parts {
		if p.Type == "text" {
			b.WriteString(p.Text)
		}
	}
	return b.String()
}

// edrSessionChange is the session's change and token from flowd's stage
// runs, as hooks/enforce-agent-baseline.py's session_change reads them.
func edrSessionChange(env Env, sessionID string) (change, token string) {
	addr := env.Getenv("FLOW_ADDR")
	if addr == "" {
		addr = "http://127.0.0.1:4173"
	}
	q := url.Values{"session_id": {sessionID}, "sort": {"-started_at"}, "limit": {"50"}}
	resp, err := (&http.Client{Timeout: 3 * time.Second, Transport: &http.Transport{}}).Get(addr + "/api/v1/stage-runs?" + q.Encode())
	if err != nil {
		return "", ""
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
		return "", ""
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
		return "", ""
	}
	return pick.ChangeName, pick.SessionToken
}
