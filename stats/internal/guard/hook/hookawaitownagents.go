package hook

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// flow-guard hook await-own-agents — the SubagentStop hook (flow-guard hook await-own-agents runs it): a
// subagent may not end its turn while an agent it launched still runs (kan-1068).
//
// A subagent's prompt cache lives 5 minutes. A planner that ended its turn to wait for its Plan
// agents woke on their hand-back 13 minutes later and rewrote its whole 206k-token context
// (kan-1061). This hook reads the `subagents/` folder beside the stopping agent's transcript,
// finds every `agent-<id>.meta.json` whose `parentAgentId` is the stopping agent and whose
// `finished` is not true, and WAITS for them itself, polling every few seconds for up to
// AWAIT_CHILDREN_SECONDS (240, inside the cache window; the default hook timeout is 10 minutes).
// When they all finish it answers {"decision":"block","reason":...} once, telling the agent to
// read their hand-backs; when some still run at the cap it blocks with "still running, stop
// again" -- each such re-stop is one cheap warm-cache turn, so the cache stays warm however long a
// child takes.
//
// A child whose transcript has not been written for 30 minutes is taken as dead and never blocks,
// so a crashed agent cannot trap its parent. stop_hook_active does not lift the block: the wait is
// the point, and every wait is bounded.
//
// A subagent may not stop while a Bash command it moved to the background still runs either: the
// tool_result of one of its own Bash tool_use calls that opens with "Command running in background
// with ID: <id>", or has a line opening with "Command [did not complete within its Ns timeout and
// ]was moved to the background (ID: <id>)" -- partial output may precede it --, launches <id>, and
// a `<task-id><id></task-id>` task-notification anywhere in its transcript ends it, as does a
// TaskStop result "Successfully stopped task: <id>" paired with one of the agent's own TaskStop
// calls. Pairing by tool_use_id and anchoring the sentence keep a launch the agent merely read --
// a Read of a transcript, a diff of a test, a grep of a JSONL line -- from trapping it. A child
// that ended with such a command out woke again after its parent had moved on -- kan-1085's fixer
// left a build running in the background and its parent waited 44 minutes on it.
//
// A stop that carries a final message (the payload's `last_assistant_message`) is a finished
// agent: its report is written, so a background command it still holds is no work anyone waits
// for. The hook kills it -- every process holding that command's `<id>.output` file open, found
// with `lsof`, so only that agent's own task is touched, never the parent's or a sibling's -- and
// lets the stop through instead of blocking on a notification that may never come (kan-1092 task
// 19: a hung build left the agent "waiting on 1 shell" until its parent stopped it by hand). An
// agent told to report through SubagentHandback that has not called it has written no report: its
// plain text is not final, so it blocks as below -- killing its build and closing its row let a
// parent remove the throwaway copies it was still working in (kan-1161). A
// stop with no final message blocks only on a background command that is still alive: one with a
// process holding its `<id>.output` file open (the same `lsof` lookup). A command whose output
// file is gone or has no holder has ended -- a shell that died, or finished with its notification
// lost -- and never blocks, so no notification is needed. When `lsof` is missing or fails, a
// command counts as alive only if its output file exists and was modified within the last 120
// seconds.
//
// A stop it lets through -- no child and no background command left -- closes the agent's flow
// dispatch row: end-dispatch-row reads the transcript's `Key:` line and final message
// and ends the row the begin hook opened, so the parent ends none itself;
// sweep-dispatch-rows then closes the rows whose own stop never did. Every stop it lets through
// or blocks first runs end-stale-dispatch-rows <cwd>, which ends the change's open
// rows whose agent finished (meta `finished`) and went quiet 10 minutes ago -- the rows of agents
// whose own stop never ended them. Each runs in process, bounded (30s, the sweeps 60s), under a
// flow-guard checkout (FLOW_GUARD_REPO_ROOT), output discarded.
//
// Fails open: any error -- bad input, unreadable folder -- prints nothing.
func init() { hooks["await-own-agents"] = awaitOwnAgents }

const (
	aoaStale        = 30 * time.Minute
	aoaLsofFallback = 120 * time.Second
	aoaPoll         = 3 * time.Second
)

var (
	aoaRunInBackground   = regexp.MustCompile(`^Command running in background with ID: ([A-Za-z0-9]+)`)
	aoaMovedToBackground = regexp.MustCompile(
		`(?m)^Command (?:did not complete within its \d+s timeout and )?was moved to the background \(ID: ([A-Za-z0-9]+)\)`)
	aoaTaskID  = regexp.MustCompile(`<task-id>([A-Za-z0-9]+)</task-id>`)
	aoaStopped = regexp.MustCompile(`Successfully stopped task: ([A-Za-z0-9]+)`)
)

func awaitOwnAgents(env core.Env, payload map[string]any, stdout, _ io.Writer) {
	if payload == nil {
		return
	}
	agentID, transcript := strOf(payload["agent_id"]), strOf(payload["agent_transcript_path"])
	if agentID == "" || transcript == "" {
		return
	}
	if transcript == "~" || strings.HasPrefix(transcript, "~/") {
		transcript = env.Getenv("HOME") + transcript[1:]
	}
	folder := filepath.Dir(transcript)
	children, err := aoaRunningChildren(agentID, folder)
	if err != nil {
		return
	}
	var background []string
	if _, err := os.Stat(transcript); err == nil {
		if background, err = aoaRunningBackground(transcript); err != nil {
			return
		}
	}
	if len(background) > 0 {
		b, err := os.ReadFile(transcript)
		if err != nil {
			return
		}
		text := strings.ToValidUTF8(string(b), "�")
		background = slices.DeleteFunc(background, func(task string) bool { return !aoaAlive(env, text, task) })
	}
	cwd := strOf(payload["cwd"])
	if cwd != "" {
		aoaInProcess(env, 60*time.Second, cwd, EndStaleDispatchRows, cwd)
	}
	if len(background) > 0 && strings.TrimSpace(strOf(payload["last_assistant_message"])) != "" && !aoaOwesHandback(transcript) {
		aoaKillBackground(env, transcript, background)
		background = nil
	}
	block := func(reason string) {
		_, _ = io.WriteString(stdout, hookJSON(map[string]any{"decision": "block", "reason": reason})+"\n")
	}
	if len(background) > 0 {
		block("Background commands you launched are still running (" + strings.Join(background, " ") + "). " +
			"FOREGROUND BUILDS: wait for each one's task-notification, or stop it with TaskStop, " +
			"before you end your turn -- a child that ends with work still out wakes again after " +
			"its parent has moved on.")
		return
	}
	if len(children) == 0 {
		if sessionID := strOf(payload["session_id"]); sessionID != "" && cwd != "" {
			aoaInProcess(env, 30*time.Second, cwd, endDispatchRow, transcript, sessionID, cwd, strOf(payload["last_assistant_message"]))
			// A stop is the moment to close rows whose own stop never did.
			aoaInProcess(env, 60*time.Second, cwd, SweepDispatchRows, folder, sessionID, cwd)
		}
		return
	}
	ids := strings.Join(children, " ")
	wait := 240
	if n, err := strconv.Atoi(strings.TrimSpace(env.Getenv("AWAIT_CHILDREN_SECONDS"))); err == nil {
		wait = n
	}
	deadline := time.Now().Add(time.Duration(wait) * time.Second)
	for len(children) > 0 && time.Now().Before(deadline) {
		time.Sleep(aoaPoll)
		if children, err = aoaRunningChildren(agentID, folder); err != nil {
			return
		}
	}
	if len(children) > 0 {
		block("Agents you launched are still running (" + strings.Join(children, " ") + "). Stop again to keep waiting " +
			"for them -- ending your turn on your own would lose your 5-minute prompt cache.")
		return
	}
	block("Agents you launched have finished (" + ids + "). Read each hand-back and continue; " +
		"say so when one returned nothing.")
}

// aoaOwesHandback: the transcript's harness reminder says the report goes through
// SubagentHandback, and no assistant turn has called it. Only a string user content is read as the
// reminder, so a tool result quoting it never counts.
func aoaOwesHandback(transcript string) bool {
	b, err := os.ReadFile(transcript)
	if err != nil {
		return false
	}
	told := false
	for _, line := range strings.Split(string(b), "\n") {
		var e struct {
			Type    string `json:"type"`
			Message struct {
				Content any `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal([]byte(line), &e) != nil {
			continue
		}
		switch c := e.Message.Content.(type) {
		case string:
			told = told || e.Type == "user" && strings.HasPrefix(c, "<system-reminder>\nYour final report is delivered through SubagentHandback")
		case []any:
			for _, p := range c {
				if m, ok := p.(map[string]any); ok && e.Type == "assistant" && m["type"] == "tool_use" && m["name"] == "SubagentHandback" {
					return false
				}
			}
		}
	}
	return told
}

// aoaInProcess runs a dispatch-row guard in process in cwd, bounded, output discarded -- only
// under a flow-guard checkout (FLOW_GUARD_REPO_ROOT), as when each ran as its own
// scripts/flow-guard.
func aoaInProcess(env core.Env, bound time.Duration, cwd string, fn core.Func, args ...string) {
	if env.Getenv("FLOW_GUARD_REPO_ROOT") != "" {
		_, _, _, _ = hookInProcess(env, bound, cwd, fn, args...)
	}
}

// aoaRunningChildren is the ids of the agent's children still running: their meta not
// `finished`, their transcript written within aoaStale. An error on a meta that is no JSON object.
func aoaRunningChildren(agentID, folder string) ([]string, error) {
	metas, _ := filepath.Glob(filepath.Join(folder, "agent-*.meta.json"))
	var out []string
	for _, meta := range metas {
		b, err := os.ReadFile(meta)
		if err != nil {
			continue
		}
		var v any
		if json.Unmarshal(b, &v) != nil {
			continue
		}
		m, ok := v.(map[string]any)
		if !ok {
			return nil, os.ErrInvalid
		}
		if m["parentAgentId"] != agentID || m["finished"] == true {
			continue
		}
		child := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(meta), "agent-"), ".meta.json")
		if st, err := os.Stat(filepath.Join(folder, "agent-"+child+".jsonl")); err == nil && time.Since(st.ModTime()) > aoaStale {
			continue
		}
		out = append(out, child)
	}
	return out, nil
}

// aoaRunningBackground is the ids of the background commands the agent's own Bash calls launched
// that no task-notification and no own TaskStop ended, in launch order.
func aoaRunningBackground(transcript string) ([]string, error) {
	f, err := os.Open(transcript)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var launched []string
	done, bashCalls, stopCalls := map[string]bool{}, map[string]bool{}, map[string]bool{}
	addAll := func(re *regexp.Regexp, s string) {
		for _, m := range re.FindAllStringSubmatch(s, -1) {
			done[m[1]] = true
		}
	}
	r := bufio.NewReader(f)
	for {
		raw, rerr := r.ReadString('\n')
		if raw != "" {
			line := strings.ToValidUTF8(raw, "�")
			addAll(aoaTaskID, line)
			var entry any
			if json.Unmarshal([]byte(line), &entry) == nil {
				// A writer may escape `<` as <; the decoded entry carries the tag either way.
				aoaEachString(entry, func(s string) { addAll(aoaTaskID, s) })
				aoaScanEntry(entry, &launched, done, bashCalls, stopCalls)
			}
		}
		if rerr != nil {
			break
		}
	}
	return slices.DeleteFunc(launched, func(id string) bool { return done[id] }), nil
}

func aoaScanEntry(entry any, launched *[]string, done, bashCalls, stopCalls map[string]bool) {
	e, _ := entry.(map[string]any)
	kind := strOf(e["type"])
	if kind != "user" && kind != "assistant" {
		return
	}
	message, _ := e["message"].(map[string]any)
	content, ok := message["content"].([]any)
	if !ok {
		return
	}
	for _, b := range content {
		block, ok := b.(map[string]any)
		if !ok {
			continue
		}
		if kind == "assistant" && block["type"] == "tool_use" && block["name"] == "Bash" {
			bashCalls[aoaIDKey(block["id"])] = true
		}
		if kind == "assistant" && block["type"] == "tool_use" && block["name"] == "TaskStop" {
			stopCalls[aoaIDKey(block["id"])] = true
		}
		use := aoaIDKey(block["tool_use_id"])
		if kind != "user" || block["type"] != "tool_result" || !bashCalls[use] && !stopCalls[use] {
			continue
		}
		var text string
		switch c := block["content"].(type) {
		case string:
			text = c
		case []any:
			var parts []string
			for _, p := range c {
				if pm, ok := p.(map[string]any); ok {
					parts = append(parts, strOf(pm["text"]))
				}
			}
			text = strings.Join(parts, " ")
		default:
			continue
		}
		if stopCalls[use] {
			for _, m := range aoaStopped.FindAllStringSubmatch(text, -1) {
				done[m[1]] = true
			}
			continue
		}
		started := aoaRunInBackground.FindStringSubmatch(text)
		if started == nil {
			started = aoaMovedToBackground.FindStringSubmatch(text)
		}
		if started != nil && !slices.Contains(*launched, started[1]) {
			*launched = append(*launched, started[1])
		}
	}
}

// aoaIDKey keys a tool-use id as Python's set membership does: an absent id is None, never "".
func aoaIDKey(v any) string {
	if s, ok := v.(string); ok {
		return "s:" + s
	}
	return "o:" + jsonStr(v)
}

// aoaEachString calls fn on every string, key or value, in a decoded JSON value.
func aoaEachString(v any, fn func(string)) {
	switch x := v.(type) {
	case string:
		fn(x)
	case []any:
		for _, e := range x {
			aoaEachString(e, fn)
		}
	case map[string]any:
		for k, e := range x {
			fn(k)
			aoaEachString(e, fn)
		}
	}
}

// aoaOutputPaths is the `<dir>/<task>.output` paths the transcript names.
func aoaOutputPaths(text, task string) []string {
	var out []string
	for _, m := range regexp.MustCompile(`(/[^\s"']*/`+regexp.QuoteMeta(task)+`\.output)`).FindAllStringSubmatch(text, -1) {
		if !slices.Contains(out, m[1]) {
			out = append(out, m[1])
		}
	}
	return out
}

// aoaHolders is the pids holding path open, or ok false when lsof is missing or fails.
func aoaHolders(env core.Env, path string) (pids []int, ok bool) {
	out, rc, err := hookRun(env, 10*time.Second, "", "lsof", "-t", "--", path)
	if err != nil || rc > 1 {
		return nil, false
	}
	for _, f := range strings.Fields(string(out)) {
		if p, err := strconv.Atoi(f); err == nil && f[0] != '-' && f[0] != '+' && p != os.Getpid() {
			pids = append(pids, p)
		}
	}
	return pids, true
}

// aoaAlive: a background command lives while a process holds its output file; without lsof, while
// the file is fresh.
func aoaAlive(env core.Env, text, task string) bool {
	for _, path := range aoaOutputPaths(text, task) {
		st, err := os.Stat(path)
		if err != nil {
			continue
		}
		pids, ok := aoaHolders(env, path)
		if !ok && time.Since(st.ModTime()) <= aoaLsofFallback || ok && len(pids) > 0 {
			return true
		}
	}
	return false
}

// aoaKillBackground SIGTERMs every process holding an `<id>.output` file of the given background
// commands open.
func aoaKillBackground(env core.Env, transcript string, ids []string) {
	b, err := os.ReadFile(transcript)
	if err != nil {
		return
	}
	text := strings.ToValidUTF8(string(b), "�")
	for _, task := range ids {
		for _, path := range aoaOutputPaths(text, task) {
			if _, err := os.Stat(path); err != nil {
				continue
			}
			pids, _ := aoaHolders(env, path)
			for _, pid := range pids {
				_ = syscall.Kill(pid, syscall.SIGTERM)
			}
		}
	}
}
