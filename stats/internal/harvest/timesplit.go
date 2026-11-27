package harvest

import (
	"encoding/json"
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
type TimeSplit struct {
	LLMMs   int64            `json:"llmMs"`
	ToolMs  int64            `json:"toolMs"`
	HumanMs int64            `json:"humanMs"`
	ByTool  map[string]int64 `json:"byTool,omitempty"`
}

// AgentTimeSplit reads agentID's own transcript under root (the file
// AgentToolCalls reads) and splits its gaps that fall in [from, to), each
// clipped to it. found is false where no transcript exists.
func AgentTimeSplit(root, agentID string, from, to time.Time) (split TimeSplit, found bool, err error) {
	raw, found, err := readAgentTranscript(root, agentID)
	if !found {
		return TimeSplit{}, false, err
	}
	complete, _ := SplitCompleteLines(raw)
	return splitTime(complete, from, to), true, nil
}

func splitTime(complete []byte, from, to time.Time) TimeSplit {
	var out TimeSplit
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
		start, end := latest(prev, from), ts
		if to.Before(end) {
			end = to
		}
		if !prev.IsZero() && end.After(start) {
			ms := end.Sub(start).Milliseconds()
			switch {
			case raw.Type == recordTypeAssistant:
				out.LLMMs += ms
			case tool != "":
				out.ToolMs += ms
				if out.ByTool == nil {
					out.ByTool = map[string]int64{}
				}
				out.ByTool[tool] += ms
			case raw.Type == "user":
				out.HumanMs += ms
			}
		}
		if ts.After(prev) {
			prev = ts
		}
	}
	return out
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
