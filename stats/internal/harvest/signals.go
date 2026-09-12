package harvest

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"
)

// SignalKind names which observability signal a Signal carries -- the
// same closed set design.md's "Harvesting" section lists for
// attribution (task 5): a compaction, a completed turn, a tool call
// starting, a tool call's result (including a denial), or a synthetic
// API-error assistant line.
type SignalKind string

const (
	SignalCompaction SignalKind = "compaction"
	SignalTurn       SignalKind = "turn"
	SignalToolUse    SignalKind = "tool_use"
	SignalToolResult SignalKind = "tool_result"
	SignalAPIError   SignalKind = "api_error"
)

// Compaction is a compact_boundary system line's own figures -- what
// triggered it and the token counts and wall-clock duration it cost.
type Compaction struct {
	Trigger    string
	PreTokens  int64
	PostTokens int64
	DurationMs int64
}

// Signal is one observability event ParseSignalRecords extracted from a
// transcript line, riding on the same Record every usage record uses --
// only the fields its Kind actually carries are populated, the rest
// stay zero.
type Signal struct {
	Kind       SignalKind
	ToolName   string // SignalToolUse
	IsError    bool   // SignalToolResult
	Denied     bool   // SignalToolResult
	DurationMs int64  // SignalTurn
	Messages   int64  // SignalTurn
	Compaction *Compaction
}

// DenialPrefixes are the literal openings of a tool_result a harness writes
// when a call was refused rather than run. Extend only with a transcript
// sample added to testdata/transcripts/signals.jsonl beside it.
var DenialPrefixes = []string{
	"Permission for this action was denied",
	"The user doesn't want to proceed",
	"PreToolUse hook",
}

func isDenial(text string) bool {
	for _, p := range DenialPrefixes {
		if strings.HasPrefix(text, p) {
			return true
		}
	}
	return false
}

// toolResultText flattens a tool_result's content, which is either a
// string or an array of {type:"text", text} blocks, into one string for
// the denial prefix check.
func toolResultText(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var blocks []struct {
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &blocks) != nil {
		return ""
	}
	var b strings.Builder
	for _, blk := range blocks {
		b.WriteString(blk.Text)
	}
	return b.String()
}

// ParseSignalRecords is ParseAssistantRecords' sibling over the same bytes:
// it yields one Record per observability signal -- a compaction, a turn,
// a tool call, a tool result, an API error -- with Usage zero and Signal
// set. tool_use blocks are deduplicated by block id and tool_result
// blocks by tool_use_id, since one API response is written as one line
// per content block and each block id appears once in a healthy
// transcript; a line repeated verbatim (a re-read overlap) is then counted
// once.
func ParseSignalRecords(complete []byte) []Record {
	var out []Record
	seenTool := map[string]bool{}
	start := 0
	for start < len(complete) {
		idx := bytes.IndexByte(complete[start:], '\n')
		if idx < 0 {
			break
		}
		end := start + idx + 1
		line := bytes.TrimSpace(complete[start:end])
		start = end
		if len(line) == 0 {
			continue
		}
		var raw rawLine
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}
		ts, err := time.Parse(time.RFC3339Nano, raw.Timestamp)
		if err != nil {
			continue
		}
		base := Record{Timestamp: ts, SessionID: raw.SessionID, IsSidechain: raw.IsSidechain, AgentID: raw.AgentID}
		emit := func(s Signal) {
			r := base
			r.Signal = &s
			out = append(out, r)
		}

		switch raw.Type {
		case "system":
			switch raw.Subtype {
			case "compact_boundary":
				c := &Compaction{}
				if raw.CompactMetadata != nil {
					c.Trigger = raw.CompactMetadata.Trigger
					c.PreTokens = raw.CompactMetadata.PreTokens
					c.PostTokens = raw.CompactMetadata.PostTokens
					c.DurationMs = raw.CompactMetadata.DurationMs
				}
				emit(Signal{Kind: SignalCompaction, Compaction: c})
			case "turn_duration":
				emit(Signal{Kind: SignalTurn, DurationMs: raw.DurationMs, Messages: raw.MessageCount})
			}
		case recordTypeAssistant:
			if raw.IsAPIErrorMessage {
				emit(Signal{Kind: SignalAPIError})
			}
			if raw.Message == nil {
				continue
			}
			for _, blk := range contentBlocks(raw.Message.Content) {
				if blk.Type != "tool_use" || blk.ID == "" || seenTool["use:"+blk.ID] {
					continue
				}
				seenTool["use:"+blk.ID] = true
				emit(Signal{Kind: SignalToolUse, ToolName: blk.Name})
			}
		case "user":
			if raw.Message == nil {
				continue
			}
			for _, blk := range contentBlocks(raw.Message.Content) {
				if blk.Type != "tool_result" || blk.ToolUseID == "" || seenTool["result:"+blk.ToolUseID] {
					continue
				}
				seenTool["result:"+blk.ToolUseID] = true
				text := toolResultText(blk.Content)
				emit(Signal{Kind: SignalToolResult, IsError: blk.IsError, Denied: blk.IsError && isDenial(text)})
			}
		}
	}
	return out
}
