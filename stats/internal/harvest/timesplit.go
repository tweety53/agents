package harvest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TimeSplit is an agent's transcript time inside a window, split by who
// held it. The rule: each gap between two consecutive timestamped lines
// belongs to the line that ends it -- an assistant line ends llm time (the
// model was generating it), a user line answering a tool_use ends tool time
// (the tool was running; filed under the first answered tool, Bash by its
// command word), any other user line ends human time (a user message after
// an assistant turn: someone was typing). Gaps never overlap, so the three
// sum to the window's covered span; a line stamped before its predecessor
// adds nothing.
//
// The parent idle split of the per-run route applies the same rule to the
// parent session's transcript and keeps only the part of each gap that lies
// inside an idle / parent segment, so its three sum to the share of idle
// the parent transcripts cover -- less than idleMs where no parent session
// transcript spans a segment.
type TimeSplit struct {
	LLMMs   int64            `json:"llmMs"`
	ToolMs  int64            `json:"toolMs"`
	HumanMs int64            `json:"humanMs"`
	ByTool  map[string]int64 `json:"byTool,omitempty"`
}

// Add adds o into s.
func (s *TimeSplit) Add(o TimeSplit) {
	s.LLMMs, s.ToolMs, s.HumanMs = s.LLMMs+o.LLMMs, s.ToolMs+o.ToolMs, s.HumanMs+o.HumanMs
	for k, v := range o.ByTool {
		if s.ByTool == nil {
			s.ByTool = map[string]int64{}
		}
		s.ByTool[k] += v
	}
}

// Gap classes, by TimeSplit's rule.
const (
	GapLLM   = "llm"
	GapTool  = "tool"
	GapHuman = "human"
)

// Gap is one classified gap of a transcript: [From, To) belongs to Class,
// and to Tool when Class is GapTool.
type Gap struct {
	From, To time.Time
	Class    string
	Tool     string
}

// ReadTranscriptGaps reads the transcript at path and returns its gaps.
func ReadTranscriptGaps(path string) ([]Gap, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return TranscriptGaps(raw), nil
}

// TranscriptGaps classes every gap of raw's complete lines by TimeSplit's
// rule, in order; a gap's To is the stamp of the line ending it, so the last
// gap's To is the transcript's latest stamp.
func TranscriptGaps(raw []byte) []Gap {
	complete, _ := SplitCompleteLines(raw)
	var out []Gap
	tools := map[string]string{} // tool_use id -> breakdown key
	var prev time.Time
	for _, line := range transcriptLines(complete) {
		var raw rawLine
		if len(line) == 0 || json.Unmarshal(line, &raw) != nil || raw.Message == nil {
			continue
		}
		ts, err := time.Parse(time.RFC3339Nano, raw.Timestamp)
		if err != nil {
			continue
		}
		tool := ""
		for _, b := range contentBlocks(raw.Message.Content) {
			switch {
			case raw.Type == recordTypeAssistant && b.Type == "tool_use" && b.ID != "":
				tools[b.ID] = toolKey(b.Name, b.Input)
			case raw.Type == "user" && b.Type == "tool_result" && tool == "":
				if tool = tools[b.ToolUseID]; tool == "" {
					tool = "unknown" // its tool_use was cut off, by a compaction or a torn line
				}
			}
		}
		if !prev.IsZero() && ts.After(prev) {
			g := Gap{From: prev, To: ts}
			switch {
			case raw.Type == recordTypeAssistant:
				g.Class = GapLLM
			case tool != "":
				g.Class, g.Tool = GapTool, tool
			case raw.Type == "user":
				g.Class = GapHuman
			}
			if g.Class != "" {
				out = append(out, g)
			}
		}
		if ts.After(prev) {
			prev = ts
		}
	}
	return out
}

// SplitGaps sums gaps into a TimeSplit, each clipped to [from, to).
func SplitGaps(gaps []Gap, from, to time.Time) TimeSplit {
	var out TimeSplit
	for _, g := range gaps {
		start, end := latest(g.From, from), g.To
		if to.Before(end) {
			end = to
		}
		if !end.After(start) {
			continue
		}
		ms := end.Sub(start).Milliseconds()
		switch g.Class {
		case GapLLM:
			out.LLMMs += ms
		case GapTool:
			out.ToolMs += ms
			if out.ByTool == nil {
				out.ByTool = map[string]int64{}
			}
			out.ByTool[g.Tool] += ms
		case GapHuman:
			out.HumanMs += ms
		}
	}
	return out
}

// AgentTranscriptPath is agentID's own transcript under root (the file
// AgentToolCalls reads), looked for under session's directory when session
// is known; found is false where none exists.
func AgentTranscriptPath(root, session, agentID string) (path string, found bool, err error) {
	dir := "*"
	if session != "" {
		dir = session
	}
	paths, err := filepath.Glob(filepath.Join(root, "*", dir, subagentsDirName, "agent-"+agentID+".jsonl"))
	if err != nil || len(paths) == 0 {
		return "", false, err
	}
	return paths[0], true, nil
}

// TokenTranscriptPath is the transcript whose agent made the marks carrying
// a Claude Code session token, cc-<session id>-<agent id, or main>
// (caller.For): the session's own <session>.jsonl for main, the agent's
// subagents/agent-<id>.jsonl otherwise. ok is false for any other token
// (mf-, fi-: no session id to derive it from); found is false where the
// file does not exist.
func TokenTranscriptPath(root, token string) (path string, ok, found bool, err error) {
	rest, isCC := strings.CutPrefix(token, "cc-")
	i := strings.LastIndex(rest, "-")
	if !isCC || i <= 0 || i == len(rest)-1 {
		return "", false, false, nil
	}
	session, agent := rest[:i], rest[i+1:]
	if agent != "main" {
		path, found, err = AgentTranscriptPath(root, session, agent)
		return path, true, found, err
	}
	path, found, err = SessionTranscriptPath(root, session)
	return path, true, found, err
}

// SessionTranscriptPath is session's own transcript, <session>.jsonl under
// root; found is false where none exists.
func SessionTranscriptPath(root, session string) (path string, found bool, err error) {
	paths, err := filepath.Glob(filepath.Join(root, "*", session+".jsonl"))
	if err != nil || len(paths) == 0 {
		return "", false, err
	}
	return paths[0], true, nil
}

// toolKey is tool, and for Bash its command's first word.
func toolKey(tool string, input json.RawMessage) string {
	if tool != "Bash" {
		return tool
	}
	var in struct {
		Command string `json:"command"`
	}
	_ = json.Unmarshal(input, &in)
	if f := strings.Fields(in.Command); len(f) > 0 {
		return tool + " " + f[0]
	}
	return tool
}

func latest(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
