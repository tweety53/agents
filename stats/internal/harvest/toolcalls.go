package harvest

import (
	"encoding/json"
	"strings"
	"time"
)

// ToolCall is one tool_use of a transcript paired with the tool_result
// naming its id: Start is the assistant line's instant, End the result
// line's. Name is the tool and a short form of its input.
type ToolCall struct {
	Name    string
	Start   time.Time
	End     time.Time
	IsError bool
}

// AgentToolCalls reads agentID's own transcript under root (the file
// AgentTranscriptEnd reads) and returns its tool calls in tool_use order.
// A call no result answers is left out; found is false where no
// transcript exists.
func AgentToolCalls(root, agentID string) (calls []ToolCall, found bool, err error) {
	raw, found, err := readAgentTranscript(root, agentID)
	if !found {
		return nil, false, err
	}
	complete, _ := SplitCompleteLines(raw)
	return parseToolCalls(complete), true, nil
}

func parseToolCalls(complete []byte) []ToolCall {
	var out []ToolCall
	open := map[string]int{} // tool_use id -> index in out
	for _, line := range transcriptLines(complete) {
		var raw rawLine
		if len(line) == 0 || json.Unmarshal(line, &raw) != nil || raw.Message == nil {
			continue
		}
		ts, err := time.Parse(time.RFC3339Nano, raw.Timestamp)
		if err != nil {
			continue
		}
		for _, b := range contentBlocks(raw.Message.Content) {
			switch {
			case raw.Type == recordTypeAssistant && b.Type == "tool_use" && b.ID != "":
				open[b.ID] = len(out)
				out = append(out, ToolCall{Name: toolCallName(b.Name, b.Input), Start: ts})
			case raw.Type == "user" && b.Type == "tool_result":
				if i, ok := open[b.ToolUseID]; ok {
					out[i].End, out[i].IsError = ts, b.IsError
					delete(open, b.ToolUseID)
				}
			}
		}
	}
	answered := out[:0]
	for _, c := range out {
		if !c.End.IsZero() {
			answered = append(answered, c)
		}
	}
	return answered
}

// toolCallName is tool, then the first line of the first of its input's
// command, file_path, pattern, url or description, cut to 60 runes.
func toolCallName(tool string, input json.RawMessage) string {
	var in map[string]any
	_ = json.Unmarshal(input, &in)
	for _, k := range []string{"command", "file_path", "pattern", "url", "description"} {
		s, _ := in[k].(string)
		if s == "" {
			continue
		}
		s, _, _ = strings.Cut(s, "\n")
		if r := []rune(s); len(r) > 60 {
			s = string(r[:60]) + "…"
		}
		return tool + ": " + s
	}
	return tool
}
